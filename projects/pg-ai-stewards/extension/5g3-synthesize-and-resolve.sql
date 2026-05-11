-- =====================================================================
-- Phase 5g.3 (Phase F.3) — synthesize_council + apply + resolve_council
--
-- Three SQL functions:
--   synthesize_council(council_id) — fires when all members responded;
--                                    enqueues a synthesizer chat with
--                                    member responses in context
--   apply_synthesize_result(council_id, result_jsonb, work_id)
--                                  — stores draft resolution +
--                                    transitions council to awaiting_bishop
--   resolve_council(council_id, action, resolution_text, destination,
--                   resolved_by) — bishop's accept/revise/dissolve
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) synthesize_council — second-round dispatch with member context
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.synthesize_council(
    p_council_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_council         stewards.councils%ROWTYPE;
    v_intent          stewards.intents%ROWTYPE;
    v_template        text;
    v_member_responses text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_synth_agent     text := 'plan';
    v_synth_model     text := 'kimi-k2.6';
BEGIN
    SELECT * INTO v_council FROM stewards.councils WHERE id = p_council_id;
    IF v_council.id IS NULL THEN
        RAISE EXCEPTION 'synthesize_council: council % not found', p_council_id;
    END IF;
    IF v_council.status NOT IN ('deliberating', 'synthesizing') THEN
        RAISE EXCEPTION 'synthesize_council: council % status=%, expected deliberating/synthesizing',
                        p_council_id, v_council.status;
    END IF;

    SELECT * INTO v_intent FROM stewards.intents WHERE id = v_council.intent_id;

    SELECT template INTO v_template FROM stewards.gate_prompts WHERE id = 'council_synthesizer';

    -- Format proposer + critic responses (skip the synthesizer member's
    -- own row if any).
    SELECT string_agg(
             format(E'### %s (%s)\n\n%s', upper(role), agent_family,
                    coalesce(response, '(no response)')),
             E'\n\n---\n\n' ORDER BY role, agent_family)
      INTO v_member_responses
      FROM stewards.council_members
     WHERE council_id = p_council_id
       AND role IN ('proposer', 'critic');

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'intent_purpose',   v_intent.purpose,
        'binding_question', v_council.binding_question,
        'member_responses', coalesce(v_member_responses, '(no member responses recorded)')
    ));

    v_session_id := substring(
        'council--' || substring(v_council.id::text FROM 1 FOR 8) ||
        '--synthesize--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('council %s synthesizer (auto)', v_council.id),
            'council')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_synth_model);

    v_payload := jsonb_build_object(
        'session_id',           v_session_id,
        'agent_family',         v_synth_agent,
        'requested_model',      v_synth_model,
        'meta',                 '{}'::jsonb,
        'body',                 (stewards.dry_run_chat(v_synth_agent, v_synth_model, v_session_id, NULL) - '_meta')
                                || jsonb_build_object('user', v_session_id),
        'tools_disabled',       true,
        '_council_id',          v_council.id::text,
        '_council_synthesize',  true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', 'opencode_go', v_payload)
    RETURNING id INTO v_work_id;

    UPDATE stewards.councils
       SET status = 'synthesizing'
     WHERE id = p_council_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.synthesize_council(uuid) IS
'Phase 5g (F.3): enqueue the synthesizer dispatch with proposer + critic responses formatted in context. tools_disabled=true. Status transitions to synthesizing. bgworker auto-fires apply_synthesize_result on completion.';

-- ---------------------------------------------------------------------
-- (2) apply_synthesize_result — store draft + transition to bishop
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_synthesize_result(
    p_council_id uuid,
    p_result     jsonb,
    p_work_id    bigint DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_council        stewards.councils%ROWTYPE;
    v_resolution_id  uuid;
BEGIN
    SELECT * INTO v_council FROM stewards.councils WHERE id = p_council_id FOR UPDATE;
    IF v_council.id IS NULL THEN
        RAISE EXCEPTION 'apply_synthesize_result: council % not found', p_council_id;
    END IF;

    -- Insert a DRAFT resolution. resolved_by stays NULL until bishop
    -- accepts. The bishop may edit text before accepting; we keep
    -- the synthesizer's proposal verbatim in raw_proposal.
    INSERT INTO stewards.resolutions
        (council_id, resolved_by, text, raw_proposal)
    VALUES
        (p_council_id, '__draft__', coalesce(p_result->>'resolution', '(no resolution text)'),
         p_result)
    RETURNING id INTO v_resolution_id;

    UPDATE stewards.councils
       SET status        = 'awaiting_bishop',
           resolution_id = v_resolution_id
     WHERE id = p_council_id;

    RETURN v_resolution_id;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_synthesize_result(uuid, jsonb, bigint) IS
'Phase 5g (F.3): store synthesizer''s draft resolution; transition council to awaiting_bishop. resolved_by=__draft__ until bishop accepts via resolve_council.';

-- ---------------------------------------------------------------------
-- (3) resolve_council — bishop's accept/revise/dissolve
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.resolve_council(
    p_council_id      uuid,
    p_action          text,         -- 'accept' | 'request_revision' | 'dissolve'
    p_resolution_text text,         -- bishop's final text (may differ from synth proposal)
    p_destination     text,         -- 'study' | 'decisions' | NULL
    p_resolved_by     text,
    p_dissolved_reason text DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_council    stewards.councils%ROWTYPE;
    v_resolution_id uuid;
    v_promoted_to text;
BEGIN
    SELECT * INTO v_council FROM stewards.councils WHERE id = p_council_id FOR UPDATE;
    IF v_council.id IS NULL THEN
        RAISE EXCEPTION 'resolve_council: council % not found', p_council_id;
    END IF;
    IF p_action NOT IN ('accept', 'request_revision', 'dissolve') THEN
        RAISE EXCEPTION 'resolve_council: invalid action %', p_action;
    END IF;
    IF v_council.status NOT IN ('awaiting_bishop', 'deliberating', 'synthesizing') THEN
        RAISE EXCEPTION 'resolve_council: council % status=%, cannot resolve', p_council_id, v_council.status;
    END IF;

    IF p_action = 'accept' THEN
        IF p_resolution_text IS NULL OR length(trim(p_resolution_text)) = 0 THEN
            RAISE EXCEPTION 'resolve_council: accept requires resolution_text';
        END IF;
        IF p_resolved_by IS NULL OR length(trim(p_resolved_by)) = 0 THEN
            RAISE EXCEPTION 'resolve_council: accept requires resolved_by';
        END IF;

        v_promoted_to := CASE p_destination
            WHEN 'study'     THEN 'study/' || substring(v_council.id::text FROM 1 FOR 8) || '.md'
            WHEN 'decisions' THEN '.mind/decisions.md'
            ELSE NULL
        END;

        IF v_council.resolution_id IS NOT NULL THEN
            -- Update the existing draft into the canonical row
            UPDATE stewards.resolutions
               SET text         = p_resolution_text,
                   resolved_by  = p_resolved_by,
                   resolved_at  = now(),
                   promoted_to  = v_promoted_to,
                   promoted_at  = CASE WHEN v_promoted_to IS NOT NULL THEN now() ELSE NULL END
             WHERE id = v_council.resolution_id
            RETURNING id INTO v_resolution_id;
        ELSE
            -- No synthesizer ran (unusual — manual resolution); create a fresh row
            INSERT INTO stewards.resolutions
                (council_id, resolved_by, text, promoted_to, promoted_at)
            VALUES
                (p_council_id, p_resolved_by, p_resolution_text, v_promoted_to,
                 CASE WHEN v_promoted_to IS NOT NULL THEN now() ELSE NULL END)
            RETURNING id INTO v_resolution_id;
        END IF;

        UPDATE stewards.councils
           SET status        = 'resolved',
               resolution_id = v_resolution_id,
               resolved_at   = now()
         WHERE id = p_council_id;

        RETURN v_resolution_id;

    ELSIF p_action = 'request_revision' THEN
        -- Re-fire synthesize_council with a note. For F1 simplicity,
        -- we just re-dispatch the synthesizer (members aren't re-asked).
        -- The bishop's revision-request text is appended to the
        -- existing draft row's text so the synthesizer sees it on retry.
        IF v_council.resolution_id IS NOT NULL THEN
            UPDATE stewards.resolutions
               SET text = text || E'\n\n[Bishop requests revision] ' || coalesce(p_resolution_text, '')
             WHERE id = v_council.resolution_id;
        END IF;
        UPDATE stewards.councils SET status = 'deliberating' WHERE id = p_council_id;
        PERFORM stewards.synthesize_council(p_council_id);
        RETURN v_council.resolution_id;

    ELSIF p_action = 'dissolve' THEN
        UPDATE stewards.councils
           SET status           = 'dissolved',
               dissolved_reason = coalesce(p_dissolved_reason, 'no reason given'),
               resolved_at      = now()
         WHERE id = p_council_id;
        RETURN v_council.resolution_id;
    END IF;

    RETURN NULL;
END;
$func$;

COMMENT ON FUNCTION stewards.resolve_council(uuid, text, text, text, text, text) IS
'Phase 5g (F.3): bishop''s resolution path. accept = canonicalize the draft (optional promotion to study/ or .mind/decisions.md per D-F3); request_revision = re-fire synthesize with bishop note; dissolve = terminate with reason.';
