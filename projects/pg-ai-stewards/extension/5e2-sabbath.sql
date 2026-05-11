-- =====================================================================
-- Phase 5e.2 (Phase D.2) — Sabbath dispatch + apply + template
--
-- Sabbath fires when a work_item reaches verified maturity on a
-- sabbath_enabled pipeline. The dispatch is a tools-off chat that
-- produces a structured reflection (D-C6 lesson: tools-off cuts cost ~7x).
--
-- The reflection is journaled to stewards.lessons (kind=sabbath_reflection)
-- and work_items.sabbath_completed_at is timestamped. promote_to_study
-- (D.5) gates on sabbath_completed_at being non-NULL for sabbath_enabled
-- pipelines.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: Sabbath gate prompt template
-- ---------------------------------------------------------------------

ALTER TABLE stewards.gate_prompts DROP CONSTRAINT IF EXISTS gate_prompts_id_check;

ALTER TABLE stewards.gate_prompts
    ADD CONSTRAINT gate_prompts_id_check
    CHECK (id IN ('evaluate','generate_scenarios','verify','covenant_check','sabbath','atonement'));

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('sabbath',
$tmpl$A work_item just reached verified maturity. Mark its ending with a structured reflection. This is not more work — it is the recording of an ending.

The intent and covenant for this work are loaded into your system prompt above.

Pipeline: {{pipeline_family}}
Binding question: {{input_summary}}
Final output (truncated):
{{stage_results_summary}}

Reflect on:
- What did this work produce that you did not expect at the start?
- What got harder than predicted? What got easier?
- What pattern would you carry forward to the next work in this pipeline?
- What is the one sentence the human should remember from this work?

Respond with JSON ONLY (no prose around it, no tool calls):
{
  "reflection": "2-4 sentences naming what this work produced and what it cost",
  "carry_forward": "one sentence: what pattern to bring to the next work in this pipeline",
  "surprise": "one sentence: what didn't go as predicted (positive or negative)"
}
$tmpl$,
     'Phase 5e (D.2): Sabbath reflection. Bgworker dispatches with tools_disabled=true (D-C6 cost lesson).')
ON CONFLICT (id) DO UPDATE SET
    template = EXCLUDED.template,
    notes    = EXCLUDED.notes,
    updated_at = now();

-- ---------------------------------------------------------------------
-- Section 2: sabbath_dispatch — enqueue a tools-off chat
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.sabbath_dispatch(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline        stewards.pipelines%ROWTYPE;
    v_template        text;
    v_input_summary   text;
    v_stage_summary   text;
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
        RAISE EXCEPTION 'sabbath_dispatch: work_item % not found', p_work_item_id;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;
    IF NOT v_pipeline.sabbath_enabled THEN
        RAISE EXCEPTION 'sabbath_dispatch: pipeline % is not sabbath_enabled', v_wi.pipeline_family;
    END IF;

    SELECT template INTO v_template FROM stewards.gate_prompts WHERE id = 'sabbath';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.sabbath template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_summary := substring(coalesce(v_wi.stage_results::text, ''), 1, 8000);

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',       v_wi.pipeline_family,
        'input_summary',         v_input_summary,
        'stage_results_summary', v_stage_summary
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--sabbath--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('sabbath work_item=%s', v_wi.id),
            'sabbath')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    v_payload := jsonb_build_object(
        'session_id',      v_session_id,
        'agent_family',    v_gate_agent,
        'requested_model', v_gate_model,
        'meta',            '{}'::jsonb,
        'body',            (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                           || jsonb_build_object('user', v_session_id),
        'tools_disabled',  true,
        '_work_item_id',   p_work_item_id::text,
        '_sabbath',        true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.sabbath_dispatch(uuid) IS
'Phase 5e (D.2): enqueue a Sabbath reflection dispatch. tools_disabled=true. bgworker auto-fires apply_sabbath_result on completion.';

-- ---------------------------------------------------------------------
-- Section 3: apply_sabbath_result — write lesson row + timestamp work_item
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_sabbath_result(
    p_work_item_id uuid,
    p_result       jsonb,
    p_work_id      bigint DEFAULT NULL
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_lesson_id    bigint;
    v_reflection   text;
    v_carry        text;
    v_surprise     text;
    v_content      text;
BEGIN
    v_reflection := coalesce(p_result->>'reflection', '');
    v_carry      := coalesce(p_result->>'carry_forward', '');
    v_surprise   := coalesce(p_result->>'surprise', '');

    -- Compose a single content block; raw_response carries the structured trio
    v_content := v_reflection;
    IF length(v_carry) > 0 THEN
        v_content := v_content || E'\n\nCarry forward: ' || v_carry;
    END IF;
    IF length(v_surprise) > 0 THEN
        v_content := v_content || E'\nSurprise: ' || v_surprise;
    END IF;

    INSERT INTO stewards.lessons
        (work_item_id, kind, content, raw_response, work_id)
    VALUES
        (p_work_item_id, 'sabbath_reflection', v_content, p_result, p_work_id)
    RETURNING id INTO v_lesson_id;

    UPDATE stewards.work_items
       SET sabbath_completed_at = now(),
           updated_at = now()
     WHERE id = p_work_item_id;

    RETURN v_lesson_id;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_sabbath_result(uuid, jsonb, bigint) IS
'Phase 5e (D.2): write Sabbath reflection to stewards.lessons + timestamp work_item.sabbath_completed_at. Returns lesson id.';
