-- =====================================================================
-- Phase 5g.2 (Phase F.2) — convene_council + member prompt templates
--
-- Three council prompt templates (proposer / critic / synthesizer)
-- + the convene_council SQL fn.
--
-- Member dispatches honor council role-specific framing in the user
-- prompt. Members get TOOLS ENABLED (unlike gates) — council
-- deliberation benefits from corpus access. Synthesizer dispatch
-- happens in F.3 with tools_disabled=true (structured output).
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) Extend gate_prompts.id check + seed council templates
-- ---------------------------------------------------------------------

ALTER TABLE stewards.gate_prompts DROP CONSTRAINT IF EXISTS gate_prompts_id_check;
ALTER TABLE stewards.gate_prompts
    ADD CONSTRAINT gate_prompts_id_check
    CHECK (id IN (
        'evaluate','generate_scenarios','verify','covenant_check',
        'sabbath','atonement',
        'council_proposer','council_critic','council_synthesizer'
    ));

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('council_proposer',
$tmpl$You are a member of a council convened to address a single binding question. Your role is PROPOSER.

The intent and active covenant for this council are loaded into your system prompt above.

Council intent: {{intent_purpose}}
Binding question: {{binding_question}}

Your job as proposer: offer a concrete proposed answer to the binding question. Lead with the answer; back it with reasoning that engages the corpus where relevant. You have substrate-internal tools (study_search_text, study_get, study_similar, study_citations) available — use them to ground your proposal in existing work.

Don't hedge. Don't list every possible angle. Take a position and defend it. The critic will stress-test it; the synthesizer will integrate.

Respond with prose (no JSON shape required). Aim for 200-500 words.
$tmpl$,
     'Phase 5g (F.2): proposer role. Tools enabled.')
ON CONFLICT (id) DO UPDATE SET
    template = EXCLUDED.template,
    notes    = EXCLUDED.notes,
    updated_at = now();

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('council_critic',
$tmpl$You are a member of a council convened to address a single binding question. Your role is CRITIC.

The intent and active covenant for this council are loaded into your system prompt above.

Council intent: {{intent_purpose}}
Binding question: {{binding_question}}

Your job as critic: find what's wrong, missing, or under-considered in the proposer's framing. The covenant's surface_tensions commitment binds you here — your function is the council's check, not its echo.

If the proposer's response is available you'll see it below; if not, articulate the strongest counterposition you can.

{{proposer_responses}}

Don't be contrarian for sport. Identify the real fault lines. What's the proposer assuming that they shouldn't? What corpus context would change the picture? You have substrate-internal tools available.

Respond with prose. 200-500 words.
$tmpl$,
     'Phase 5g (F.2): critic role. Tools enabled. surface_tensions covenant directly applied.')
ON CONFLICT (id) DO UPDATE SET
    template = EXCLUDED.template,
    notes    = EXCLUDED.notes,
    updated_at = now();

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('council_synthesizer',
$tmpl$You are the synthesizer for a council convened to address a single binding question.

The intent and active covenant for this council are loaded into your system prompt above.

Council intent: {{intent_purpose}}
Binding question: {{binding_question}}

Council members responded:

{{member_responses}}

Your job: produce a single proposed resolution. Honor the proposer's instinct where it survived the critic; honor the critic's catch where the proposer missed something; name the genuine tension where both have a point and the human bishop needs to decide.

Don't paper over disagreement. Don't pretend to consensus that isn't there.

Respond with JSON ONLY (no prose around it, no tool calls):
{
  "resolution": "the proposed answer (1-3 paragraphs)",
  "tensions": ["unresolved tension 1", "tension 2", ...],
  "destination_hint": "study" | "decisions" | "either" | "none"
}

destination_hint guides the bishop: 'study' if the resolution belongs in study/<slug>.md (doctrinal/narrative), 'decisions' if it belongs in .mind/decisions.md (engineering/operational), 'either' if both, 'none' if it should stay in the resolutions table only.
$tmpl$,
     'Phase 5g (F.2): synthesizer role. Tools DISABLED (structured JSON output). Per D-F3, destination_hint feeds the bishop''s promotion choice.')
ON CONFLICT (id) DO UPDATE SET
    template = EXCLUDED.template,
    notes    = EXCLUDED.notes,
    updated_at = now();

-- ---------------------------------------------------------------------
-- (2) convene_council — D-F1 enforcement + parallel member dispatch
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.convene_council(
    p_intent_id        uuid,
    p_binding_question text,
    p_members          jsonb,           -- [{"agent_family":"plan","role":"proposer","model":"kimi-k2.6"}, ...]
    p_bishop           text,
    p_convened_by      text DEFAULT 'human'
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_council_id  uuid;
    v_intent      stewards.intents%ROWTYPE;
    v_member      jsonb;
    v_role        text;
    v_agent       text;
    v_model       text;
    v_session_id  text;
    v_template_id text;
    v_template    text;
    v_prompt      text;
    v_payload     jsonb;
    v_work_id     bigint;
    v_provider    text := 'opencode_go';
    v_tools_off   boolean;
    v_member_count int;
BEGIN
    SELECT * INTO v_intent FROM stewards.intents WHERE id = p_intent_id;
    IF v_intent.id IS NULL THEN
        RAISE EXCEPTION 'convene_council: intent % not found', p_intent_id;
    END IF;

    -- Validate members shape
    IF p_members IS NULL OR jsonb_typeof(p_members) <> 'array' THEN
        RAISE EXCEPTION 'convene_council: p_members must be a jsonb array';
    END IF;

    v_member_count := jsonb_array_length(p_members);
    IF v_member_count < 2 OR v_member_count > 5 THEN
        RAISE EXCEPTION 'convene_council: must have between 2 and 5 members (got %)', v_member_count;
    END IF;

    -- D-F1: refuse if there's an active council. The partial unique
    -- index would also reject the INSERT but a clean error message is
    -- friendlier.
    IF EXISTS (SELECT 1 FROM stewards.councils
                WHERE status IN ('deliberating', 'synthesizing', 'awaiting_bishop')) THEN
        RAISE EXCEPTION 'convene_council: one council at a time (D-F1) — resolve or dissolve the active council first';
    END IF;

    INSERT INTO stewards.councils (intent_id, binding_question, convened_by, bishop)
    VALUES (p_intent_id, p_binding_question, p_convened_by, p_bishop)
    RETURNING id INTO v_council_id;

    -- Dispatch each member in parallel
    FOR v_member IN SELECT * FROM jsonb_array_elements(p_members) LOOP
        v_role  := v_member->>'role';
        v_agent := v_member->>'agent_family';
        v_model := coalesce(v_member->>'model', 'kimi-k2.6');

        IF v_role NOT IN ('proposer', 'critic', 'synthesizer') THEN
            RAISE EXCEPTION 'convene_council: invalid role % for agent %', v_role, v_agent;
        END IF;

        v_template_id := 'council_' || v_role;
        SELECT template INTO v_template
          FROM stewards.gate_prompts WHERE id = v_template_id;

        v_session_id := substring(
            'council--' || substring(v_council_id::text FROM 1 FOR 8) ||
            '--' || v_role || '--' || v_agent,
            1, 200);

        INSERT INTO stewards.sessions (id, label, kind)
        VALUES (v_session_id,
                format('council %s role=%s agent=%s', v_council_id, v_role, v_agent),
                'council')
        ON CONFLICT (id) DO NOTHING;

        -- For proposer + critic, render template with intent + binding question.
        -- For synthesizer dispatched here, member_responses is empty —
        -- the synthesizer is normally re-dispatched by F.3 once members
        -- complete. But allowing it as a council member at convene
        -- time keeps the data model symmetric.
        v_prompt := stewards.render_template(v_template, jsonb_build_object(
            'intent_purpose',     v_intent.purpose,
            'binding_question',   p_binding_question,
            'proposer_responses', '(none yet — proposer responses arrive in parallel)',
            'member_responses',   '(none yet — members responding in parallel)'
        ));

        INSERT INTO stewards.messages (session_id, role, content, model)
        VALUES (v_session_id, 'user', v_prompt, v_model);

        v_tools_off := (v_role = 'synthesizer');

        v_payload := jsonb_build_object(
            'session_id',      v_session_id,
            'agent_family',    v_agent,
            'requested_model', v_model,
            'meta',            '{}'::jsonb,
            'body',            (stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL) - '_meta')
                               || jsonb_build_object('user', v_session_id),
            'tools_disabled',  v_tools_off,
            '_council_id',     v_council_id::text,
            '_council_member', true,
            '_council_role',   v_role
        );

        INSERT INTO stewards.work_queue (kind, provider, payload)
        VALUES ('chat', v_provider, v_payload)
        RETURNING id INTO v_work_id;

        INSERT INTO stewards.council_members (council_id, agent_family, role, work_id)
        VALUES (v_council_id, v_agent, v_role, v_work_id);
    END LOOP;

    RETURN v_council_id;
END;
$func$;

COMMENT ON FUNCTION stewards.convene_council(uuid, text, jsonb, text, text) IS
'Phase 5g (F.2): convene a new council. Validates intent, members shape (2-5), D-F1 (one active at a time). Dispatches each member in parallel via work_queue with role-specific prompt + _council_id + _council_member markers. Synthesizer member gets tools_disabled=true; proposer/critic get tools enabled.';
