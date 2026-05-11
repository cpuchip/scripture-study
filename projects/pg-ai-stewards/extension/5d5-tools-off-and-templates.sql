-- =====================================================================
-- Phase 5d.5 (Phase C.6) — tools_disabled propagation + revised gate
-- prompts + new covenant_check template
--
-- Three pieces:
--
-- (1) gate_prompts.id CHECK extended to include 'covenant_check'
-- (2) gate_prompts.evaluate revised to reference intent (composer
--     prepends intent in compose_system_prompt now, but the user
--     prompt also asks the model to honor it)
-- (3) New gate_prompts.covenant_check template (free-form per D-C4
--     but tools-off — the bgworker honors payload.tools_disabled)
-- (4) evaluate_gate / generate_scenarios / verify_work_item / a new
--     covenant_check_dispatch all set tools_disabled=true on payload
--
-- Phase B's lesson (2026-05-11): gate-eval through plan-agent with
-- tools enabled = 5x cost from research loop. tools_disabled=true
-- is now the standard for any JSON-output gate-style dispatch.
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) Extend gate_prompts.id check to include covenant_check
-- ---------------------------------------------------------------------

ALTER TABLE stewards.gate_prompts
    DROP CONSTRAINT IF EXISTS gate_prompts_id_check;

ALTER TABLE stewards.gate_prompts
    ADD CONSTRAINT gate_prompts_id_check
    CHECK (id IN ('evaluate','generate_scenarios','verify','covenant_check'));

-- ---------------------------------------------------------------------
-- (2) Revise evaluate template to reference intent + covenant
-- ---------------------------------------------------------------------

UPDATE stewards.gate_prompts SET template = $tmpl$You are a gate evaluator for a structured second-brain pipeline. Your job is to decide whether a piece of work has matured enough to advance, needs revision, or needs human steering.

The intent and covenant for this work are loaded into your system prompt above — keep them in mind. The covenant's surface_tensions and check_existing_work commitments apply to your evaluation.

Pipeline: {{pipeline_family}}
Current stage just completed: {{current_stage}}
Current maturity: {{maturity}}
Maturity this stage produces: {{produces_maturity}}
Revision count for this maturity: {{revision_count}}

Binding question / input:
{{input_summary}}

Latest stage output:
{{stage_output}}

Decide ONE of:
- "advance" — the work has clearly satisfied the criteria for this maturity AND advances the stated intent. Move to the next stage / next maturity.
- "revise" — the work is on the right track but needs another pass. Provide specific, actionable feedback for what to improve.
- "surface" — the work needs human steering. Either it drifts from the stated intent, hit a constraint you can't resolve, or the binding question shifted. Provide a brief explanation of what the human needs to decide.

Respond with JSON ONLY (no prose around it, no tool calls):
{
  "action": "advance" | "revise" | "surface",
  "reasoning": "1-3 sentences explaining the decision, referencing intent/covenant where relevant",
  "feedback": "if revise: what to do differently next pass; if surface: what the human needs to decide; if advance: omit or empty string"
}
$tmpl$,
notes = 'Phase 5d (C.6 revision): references intent + covenant from system prompt; reminds model no tool calls. Default gate evaluation prompt; bgworker dispatches with tools_disabled=true.',
updated_at = now()
WHERE id = 'evaluate';

-- ---------------------------------------------------------------------
-- (3) New covenant_check template (Phase C.4 / D-C4 free-form check)
-- ---------------------------------------------------------------------

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('covenant_check',
$tmpl$You are evaluating whether a piece of work honors the active covenant.

The covenant commitments are loaded into your system prompt above. Pay particular attention to the agent commitments — those are what THIS work was supposed to honor.

Pipeline: {{pipeline_family}}
Stage: {{current_stage}}
Target maturity (the rung this work is about to advance to): {{target_maturity}}

The work produced this output:
{{stage_output}}

Question: does this output honor the agent's covenant commitments? Specifically check:
- read_before_quoting: are direct quotes verifiable, or does the output paraphrase what isn't checked?
- check_existing_work: does the output engage with prior work in the corpus, or build in isolation?
- surface_tensions: does the output acknowledge counterarguments / blind spots, or only build toward a thesis?
- honor_scope: did the output stay within the requested scope, or expand into adjacent territory?
- exercise_stewardship: where the output found adjacent issues, did it act on them or only flag them?

Respond with JSON ONLY (no prose, no tool calls):
{
  "honors_covenant": true | false,
  "concerns": ["concern 1", "concern 2", ...],   // empty array if no concerns
  "recommendation": "pass" | "flag"               // flag = surface to human even if technically passes
}
$tmpl$,
     'Phase 5d (C.6, D-C4): free-form covenant check. Bgworker dispatches with tools_disabled=true.')
ON CONFLICT (id) DO UPDATE SET
    template = EXCLUDED.template,
    notes    = EXCLUDED.notes,
    updated_at = now();

-- ---------------------------------------------------------------------
-- (4) Stamp tools_disabled=true on the three existing gate dispatches.
--     Each one has the same shape:
--       jsonb_build_object(...everything..., '_marker', true, ...)
--     The simplest non-invasive fix: drop and re-create each function
--     with the same body but adding 'tools_disabled' true to the
--     payload jsonb_build_object.
--
--     Rather than rewriting all three function bodies here, ship a
--     small helper that wraps the payload:
--         stewards.gate_payload_disable_tools(payload jsonb)
--     and have the existing functions opt-in via a one-line edit.
--     But since the goal is "ship Phase C", do the surgical edits
--     directly here for evaluate_gate, generate_scenarios,
--     verify_work_item.
-- ---------------------------------------------------------------------

-- evaluate_gate: append tools_disabled=true to the payload composition.
CREATE OR REPLACE FUNCTION stewards.evaluate_gate(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_produces_maturity text;
    v_template        text;
    v_input_summary   text;
    v_stage_output    text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_gate_model      text := 'qwen3.6-plus';
    v_gate_provider   text := 'opencode_go';
    v_gate_agent      text := 'plan';
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    SELECT produces_maturity INTO v_produces_maturity
      FROM stewards.pipeline_stage_maturity
     WHERE pipeline_family = v_wi.pipeline_family
       AND stage_name = v_wi.current_stage;

    SELECT template INTO v_template
      FROM stewards.gate_prompts WHERE id = 'evaluate';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.evaluate template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_output  := substring(
        coalesce(v_wi.stage_results->v_wi.current_stage->>'output', ''),
        1, 8000);

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',   v_wi.pipeline_family,
        'current_stage',     v_wi.current_stage,
        'maturity',          v_wi.maturity,
        'produces_maturity', coalesce(v_produces_maturity, '(none)'),
        'revision_count',    v_wi.revision_count::text,
        'input_summary',     v_input_summary,
        'stage_output',      v_stage_output
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--gate-' ||
        v_wi.maturity || '--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('gate eval work_item=%s maturity=%s', v_wi.id, v_wi.maturity),
            'gate')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_gate_agent,
        'requested_model',    v_gate_model,
        'meta',               '{}'::jsonb,
        'body',               (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                              || jsonb_build_object('user', v_session_id),
        'tools_disabled',     true,           -- Phase C.6: structured JSON output
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family,
        '_gate_eval',         true,
        '_gate_from_maturity', v_wi.maturity
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.evaluate_gate(uuid) IS
'Phase 5a + 5d (C.6): enqueues a gate-eval chat for a work_item. Now sets tools_disabled=true so the model returns JSON without a tool research loop (Phase B 2026-05-11 lesson).';

-- generate_scenarios (Phase 5b) — re-create with tools_disabled
CREATE OR REPLACE FUNCTION stewards.generate_scenarios(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_template        text;
    v_input_summary   text;
    v_stage_output    text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_gate_model      text := 'kimi-k2.6';
    v_gate_provider   text := 'opencode_go';
    v_gate_agent      text := 'plan';
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    SELECT template INTO v_template
      FROM stewards.gate_prompts WHERE id = 'generate_scenarios';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.generate_scenarios template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_output := substring(
        coalesce(v_wi.spec, v_wi.stage_results->v_wi.current_stage->>'output', ''),
        1, 8000);

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',     v_wi.pipeline_family,
        'input_summary',       v_input_summary,
        'spec_or_stage_output', v_stage_output
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--scenarios--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('scenarios gen work_item=%s', v_wi.id),
            'gate')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_gate_agent,
        'requested_model',    v_gate_model,
        'meta',               '{}'::jsonb,
        'body',               (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                              || jsonb_build_object('user', v_session_id),
        'tools_disabled',     true,           -- Phase C.6
        '_work_item_id',      p_work_item_id::text,
        '_scenarios_gen',     true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.generate_scenarios(uuid) IS
'Phase 5b + 5d (C.6): scenarios dispatch with tools_disabled=true.';

-- verify_work_item (Phase 5b) — re-create with tools_disabled
CREATE OR REPLACE FUNCTION stewards.verify_work_item(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_template        text;
    v_input_summary   text;
    v_stage_output    text;
    v_scenarios_str   text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_gate_model      text := 'qwen3.6-plus';
    v_gate_provider   text := 'opencode_go';
    v_gate_agent      text := 'plan';
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    IF v_wi.scenarios IS NULL OR jsonb_array_length(v_wi.scenarios) = 0 THEN
        RAISE EXCEPTION 'verify_work_item: work_item % has no scenarios — call generate_scenarios first', p_work_item_id;
    END IF;

    SELECT template INTO v_template
      FROM stewards.gate_prompts WHERE id = 'verify';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.verify template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_output := substring(
        coalesce(v_wi.stage_results->v_wi.current_stage->>'output', ''),
        1, 8000);

    SELECT string_agg('  - ' || s, E'\n')
      INTO v_scenarios_str
      FROM jsonb_array_elements_text(v_wi.scenarios) s;

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family', v_wi.pipeline_family,
        'input_summary',   v_input_summary,
        'scenarios',       coalesce(v_scenarios_str, '(none)'),
        'stage_output',    v_stage_output
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--verify--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('verify work_item=%s', v_wi.id),
            'gate')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_gate_agent,
        'requested_model',    v_gate_model,
        'meta',               '{}'::jsonb,
        'body',               (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                              || jsonb_build_object('user', v_session_id),
        'tools_disabled',     true,           -- Phase C.6
        '_work_item_id',      p_work_item_id::text,
        '_verify',            true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.verify_work_item(uuid) IS
'Phase 5b + 5d (C.6): verify dispatch with tools_disabled=true.';
