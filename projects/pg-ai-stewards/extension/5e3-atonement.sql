-- =====================================================================
-- Phase 5e.3 (Phase D.3) — Atonement dispatch + apply + template
--
-- Atonement fires when a work_item is quarantined on an
-- atonement_enabled pipeline. The dispatch is a tools-off chat that
-- produces structured lessons:
--   {principles_to_record: [], decisions: [], lessons: []}
-- Each item lands as a row in stewards.lessons (unratified). Humans
-- curate via Stewards-UI before promoting to .mind/principles.md
-- (D-D3).
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: Atonement gate prompt template
-- ---------------------------------------------------------------------

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('atonement',
$tmpl$A work_item was quarantined after {{failure_count}} failures. Walk back through what was tried, what failed, what was eventually completed (or not), and propose lessons that should outlive this work_item.

The intent and covenant for this work are loaded into your system prompt above.

Pipeline: {{pipeline_family}}
Binding question: {{input_summary}}
Failure count: {{failure_count}}
Quarantine reason: {{quarantine_reason}}

Failure history (steward actions, most recent first):
{{steward_actions_summary}}

Final stage results:
{{stage_results_summary}}

Distinguish three kinds of takeaways:
- principles: enduring insights about HOW the work should be done (candidate for .mind/principles.md)
- decisions: specific choices made about THIS pipeline/stage that should be recorded (candidate for .mind/decisions.md)
- lessons: ephemeral observations relevant only for similar future work (substrate-only)

Be sparse. Three lessons that survive scrutiny beat thirty that get pruned.

Respond with JSON ONLY (no prose around it, no tool calls):
{
  "principles_to_record": ["principle 1", "principle 2", ...],
  "decisions": ["decision 1", ...],
  "lessons": ["lesson 1", "lesson 2", ...]
}
$tmpl$,
     'Phase 5e (D.3): Atonement extraction. Bgworker dispatches with tools_disabled=true.')
ON CONFLICT (id) DO UPDATE SET
    template = EXCLUDED.template,
    notes    = EXCLUDED.notes,
    updated_at = now();

-- ---------------------------------------------------------------------
-- Section 2: atonement_dispatch
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.atonement_dispatch(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline        stewards.pipelines%ROWTYPE;
    v_template        text;
    v_input_summary   text;
    v_stage_summary   text;
    v_actions_summary text;
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
        RAISE EXCEPTION 'atonement_dispatch: work_item % not found', p_work_item_id;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;
    IF NOT v_pipeline.atonement_enabled THEN
        RAISE EXCEPTION 'atonement_dispatch: pipeline % is not atonement_enabled', v_wi.pipeline_family;
    END IF;

    SELECT template INTO v_template FROM stewards.gate_prompts WHERE id = 'atonement';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.atonement template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_summary := substring(coalesce(v_wi.stage_results::text, ''), 1, 6000);

    -- Last 20 steward actions for this work_item, formatted compactly
    SELECT string_agg(
             '  - [' || to_char(at, 'YYYY-MM-DD HH24:MI') || '] ' || action ||
             coalesce(' (' || diagnosis || ')', '') ||
             ': ' || observation,
             E'\n' ORDER BY at DESC)
      INTO v_actions_summary
      FROM (
        SELECT at, action, diagnosis, observation
          FROM stewards.steward_actions
         WHERE work_item_id = p_work_item_id
         ORDER BY at DESC
         LIMIT 20
      ) t;

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',         v_wi.pipeline_family,
        'input_summary',           v_input_summary,
        'failure_count',           v_wi.failure_count::text,
        'quarantine_reason',       coalesce(v_wi.quarantine_reason, '(none)'),
        'steward_actions_summary', coalesce(v_actions_summary, '  (no steward actions recorded)'),
        'stage_results_summary',   v_stage_summary
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--atonement--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('atonement work_item=%s', v_wi.id),
            'atonement')
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
        '_atonement',      true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.atonement_dispatch(uuid) IS
'Phase 5e (D.3): enqueue an Atonement extraction dispatch. tools_disabled=true. bgworker auto-fires apply_atonement_result on completion.';

-- ---------------------------------------------------------------------
-- Section 3: apply_atonement_result — write one lesson row per item
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_atonement_result(
    p_work_item_id uuid,
    p_result       jsonb,
    p_work_id      bigint DEFAULT NULL
) RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_principles jsonb;
    v_decisions  jsonb;
    v_lessons    jsonb;
    v_item       text;
    v_count      int := 0;
BEGIN
    v_principles := coalesce(p_result->'principles_to_record', '[]'::jsonb);
    v_decisions  := coalesce(p_result->'decisions',            '[]'::jsonb);
    v_lessons    := coalesce(p_result->'lessons',              '[]'::jsonb);

    FOR v_item IN SELECT jsonb_array_elements_text(v_principles) LOOP
        INSERT INTO stewards.lessons
            (work_item_id, kind, content, raw_response, work_id)
        VALUES
            (p_work_item_id, 'principle', v_item, p_result, p_work_id);
        v_count := v_count + 1;
    END LOOP;

    FOR v_item IN SELECT jsonb_array_elements_text(v_decisions) LOOP
        INSERT INTO stewards.lessons
            (work_item_id, kind, content, raw_response, work_id)
        VALUES
            (p_work_item_id, 'decision', v_item, p_result, p_work_id);
        v_count := v_count + 1;
    END LOOP;

    FOR v_item IN SELECT jsonb_array_elements_text(v_lessons) LOOP
        INSERT INTO stewards.lessons
            (work_item_id, kind, content, raw_response, work_id)
        VALUES
            (p_work_item_id, 'lesson', v_item, p_result, p_work_id);
        v_count := v_count + 1;
    END LOOP;

    RETURN v_count;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_atonement_result(uuid, jsonb, bigint) IS
'Phase 5e (D.3): write one stewards.lessons row per item across {principles, decisions, lessons}. All rows land unratified (D-D3 human curation). Returns total count inserted.';
