-- =====================================================================
-- Phase 5b — Scenarios + Verify functions (Phase B push 3)
--
-- Builds on 5a (maturity ladder, gate_decisions, gate_prompts,
-- verify_results, render_template, evaluate_gate, apply_gate_decision).
--
-- Mirrors the gate_eval pattern for two more LLM-mediated gates:
--   - generate_scenarios — produces JSON array of acceptance criteria
--     when maturity advances to specced. Stored in work_items.scenarios.
--   - verify_work_item — checks executed output against scenarios.
--     Stored in verify_results. all_passed=false drops maturity back
--     to planned with feedback for re-execute.
--
-- Pattern for each:
--   1. enqueue helper (generate_scenarios / verify_work_item) sets
--      payload._scenarios_gen=true OR _verify=true so bgworker can
--      auto-detect after the chat completes
--   2. parse helper (parse_scenarios_response / parse_verify_response)
--      extracts JSON from the assistant message
--   3. apply helper (apply_scenarios_result / apply_verify_result)
--      writes the result to the appropriate substrate state
--
-- Bgworker auto-fire of all three lands in the same Rust push as
-- gate_eval auto-fire (B.2), so this file only delivers SQL.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: generate_scenarios(work_item_id) — enqueue scenarios chat
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.generate_scenarios(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_template        text;
    v_input_summary   text;
    v_spec_or_output  text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    -- _gate/generate_scenarios default per stage_models seed
    v_model           text := 'kimi-k2.6';     -- needs creativity
    v_provider        text := 'opencode_go';
    v_agent           text := 'plan';
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
    -- Prefer explicit spec column; fall back to most-recent stage output.
    v_spec_or_output := coalesce(
        v_wi.spec,
        substring(coalesce(v_wi.stage_results->v_wi.current_stage->>'output', ''), 1, 8000),
        '');

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',       v_wi.pipeline_family,
        'input_summary',         v_input_summary,
        'spec_or_stage_output',  v_spec_or_output
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--scenarios-' ||
        to_char(extract(epoch FROM now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('scenarios gen work_item=%s', v_wi.id),
            'gate')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_model);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_agent,
        'requested_model',    v_model,
        'meta',               '{}'::jsonb,
        'body',               (stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL) - '_meta')
                              || jsonb_build_object('user', v_session_id),
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family,
        '_scenarios_gen',     true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.generate_scenarios(uuid) IS
'Phase 5b: enqueue a chat that generates 3-7 acceptance criteria for a work_item. Output written to work_items.scenarios via apply_scenarios_result (auto-fired by bgworker on _scenarios_gen marker).';

-- ---------------------------------------------------------------------
-- Section 2: apply_scenarios_result(work_item_id, scenarios_array)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_scenarios_result(
    p_work_item_id uuid,
    p_scenarios    jsonb,
    p_work_id      bigint DEFAULT NULL
) RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_count int;
BEGIN
    -- Accept either {"scenarios": [...]} or [...] directly
    IF jsonb_typeof(p_scenarios) = 'object' AND p_scenarios ? 'scenarios' THEN
        p_scenarios := p_scenarios->'scenarios';
    END IF;

    IF jsonb_typeof(p_scenarios) != 'array' THEN
        RAISE EXCEPTION 'apply_scenarios_result: expected JSON array, got %',
            jsonb_typeof(p_scenarios);
    END IF;

    v_count := jsonb_array_length(p_scenarios);

    UPDATE stewards.work_items
       SET scenarios  = p_scenarios,
           updated_at = now()
     WHERE id = p_work_item_id;

    -- Audit in steward_actions for traceability
    INSERT INTO stewards.steward_actions
        (work_item_id, observation, diagnosis, action, details)
    VALUES
        (p_work_item_id,
         format('scenarios generated: %s criteria', v_count),
         'gate',
         'scenarios_generated',
         jsonb_build_object('count', v_count, 'work_id', p_work_id));

    RETURN v_count;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_scenarios_result(uuid, jsonb, bigint) IS
'Phase 5b: write generated scenarios to work_items.scenarios. Accepts {"scenarios":[...]} or bare array. Returns count.';

-- ---------------------------------------------------------------------
-- Section 3: verify_work_item(work_item_id) — enqueue verify chat
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.verify_work_item(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_template        text;
    v_input_summary   text;
    v_stage_output    text;
    v_scenarios_text  text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_model           text := 'qwen3.6-plus';   -- _gate/verify_scenarios default
    v_provider        text := 'opencode_go';
    v_agent           text := 'plan';
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    IF jsonb_array_length(coalesce(v_wi.scenarios, '[]'::jsonb)) = 0 THEN
        RAISE EXCEPTION 'work_item % has no scenarios; run generate_scenarios first',
            p_work_item_id;
    END IF;

    SELECT template INTO v_template
      FROM stewards.gate_prompts WHERE id = 'verify';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.verify template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_output  := substring(
        coalesce(v_wi.stage_results->v_wi.current_stage->>'output', ''),
        1, 8000);

    -- Render scenarios as a numbered list for the prompt
    SELECT string_agg(
               format('%s. %s', ord, scenario),
               E'\n'
               ORDER BY ord)
      INTO v_scenarios_text
      FROM (
          SELECT row_number() OVER () AS ord, value::text AS scenario
            FROM jsonb_array_elements_text(v_wi.scenarios)
      ) s;

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',  v_wi.pipeline_family,
        'input_summary',    v_input_summary,
        'scenarios',        coalesce(v_scenarios_text, '(none)'),
        'stage_output',     v_stage_output
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--verify-' ||
        to_char(extract(epoch FROM now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('verify work_item=%s', v_wi.id),
            'gate')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_model);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_agent,
        'requested_model',    v_model,
        'meta',               '{}'::jsonb,
        'body',               (stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL) - '_meta')
                              || jsonb_build_object('user', v_session_id),
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family,
        '_verify',            true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.verify_work_item(uuid) IS
'Phase 5b: enqueue a verify chat that checks execution output against the work_item.scenarios. Result written via apply_verify_result (auto-fired by bgworker on _verify marker).';

-- ---------------------------------------------------------------------
-- Section 4: apply_verify_result(work_item_id, result_jsonb)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_verify_result(
    p_work_item_id uuid,
    p_result       jsonb,
    p_work_id      bigint DEFAULT NULL
) RETURNS boolean
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_all_passed  boolean;
    v_reasoning   text;
    v_results     jsonb;
    v_failed_text text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    v_all_passed := coalesce((p_result->>'all_passed')::boolean, false);
    v_reasoning  := p_result->>'reasoning';
    v_results    := coalesce(p_result->'results', '[]'::jsonb);

    INSERT INTO stewards.verify_results
        (work_item_id, all_passed, reasoning, results, work_id)
    VALUES
        (p_work_item_id, v_all_passed, v_reasoning, v_results, p_work_id);

    INSERT INTO stewards.steward_actions
        (work_item_id, observation, diagnosis, action, details)
    VALUES
        (p_work_item_id,
         format('verify %s: %s',
                CASE WHEN v_all_passed THEN 'PASSED' ELSE 'FAILED' END,
                coalesce(v_reasoning, '(no reasoning)')),
         'gate',
         CASE WHEN v_all_passed THEN 'verify_passed' ELSE 'verify_failed' END,
         jsonb_build_object(
             'all_passed', v_all_passed,
             'work_id', p_work_id,
             'failed_count', (
                 SELECT count(*) FROM jsonb_array_elements(v_results) r
                  WHERE coalesce((r->>'passed')::boolean, false) = false
             )));

    IF v_all_passed THEN
        -- Advance maturity to verified (or whatever's next)
        UPDATE stewards.work_items
           SET maturity   = 'verified',
               updated_at = now()
         WHERE id = p_work_item_id;
    ELSE
        -- Drop maturity back to planned, surface failed scenarios as
        -- feedback so re-execute can target them. Status='failed' so
        -- the steward retry path picks it up.
        SELECT string_agg(
                   format('- %s%s',
                          coalesce(r->>'scenario', '(unknown)'),
                          CASE WHEN (r->>'notes') IS NOT NULL
                               THEN ': ' || (r->>'notes')
                               ELSE '' END),
                   E'\n')
          INTO v_failed_text
          FROM jsonb_array_elements(v_results) r
         WHERE coalesce((r->>'passed')::boolean, false) = false;

        UPDATE stewards.work_items
           SET maturity               = 'planned',
               status                 = 'failed',
               last_failure_reason    = 'verify failed: ' ||
                                        coalesce(v_reasoning, '') ||
                                        E'\n\nFailed criteria:\n' ||
                                        coalesce(v_failed_text, '(none)'),
               last_failure_diagnosis = 'verify_failed',
               updated_at             = now()
         WHERE id = p_work_item_id;
    END IF;

    RETURN v_all_passed;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_verify_result(uuid, jsonb, bigint) IS
'Phase 5b: write verify result to verify_results + steward_actions. all_passed=true → maturity=verified; all_passed=false → maturity=planned + status=failed (steward retry will re-execute) with failed criteria as feedback.';

-- =====================================================================
-- Done. Phase 5b scenarios + verify functions are operational.
--
-- Manual flow (auto-fire lands in the bgworker push):
--   1. SELECT stewards.generate_scenarios('<wi_id>') → work_id
--   2. wait for chat
--   3. SELECT stewards.parse_gate_response(<work_id>) → jsonb
--      (the parse fn doubles for any JSON-returning gate chat)
--   4. SELECT stewards.apply_scenarios_result('<wi_id>', <jsonb>, <work_id>)
--      → int (count of scenarios written)
--   5. (later, after execute) SELECT stewards.verify_work_item('<wi_id>')
--   6. SELECT stewards.parse_gate_response(<work_id>) → jsonb
--   7. SELECT stewards.apply_verify_result('<wi_id>', <jsonb>, <work_id>)
--      → boolean (true=passed, false=failed and dropped maturity)
-- =====================================================================
