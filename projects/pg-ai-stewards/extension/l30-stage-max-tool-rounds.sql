-- =====================================================================
-- L.1.1.16 — Per-stage max_tool_rounds enforcement
-- =====================================================================
-- Bacteriopolis synthesize stage ran 53 rounds when it should run 1.
-- Stage prompts can ask for limits but the substrate didn't enforce.
--
-- Add stages[].max_tool_rounds field (jsonb, no schema change). When
-- chat_post_internal enqueues a continuation chat, count prior
-- assistant rounds in the session; if at-or-above max_tool_rounds,
-- force tools_disabled=true so the agent must produce a final answer
-- in the next round (no more tool calls allowed).
--
-- Default cap = 5. Per Michael: 'cap it to something like 3 or 5
-- and then lower I want to see if end up having ones that needs
-- more rounds after the fixes.' Tunable per-stage going forward.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: stage_max_tool_rounds(pipeline, stage) → int.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.stage_max_tool_rounds(
    p_pipeline_family text,
    p_stage_name      text
) RETURNS int LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_stage    jsonb;
    v_rounds   int;
BEGIN
    IF p_pipeline_family IS NULL OR p_stage_name IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT s INTO v_stage
      FROM stewards.pipelines p,
           LATERAL jsonb_array_elements(p.stages) s
     WHERE p.family = p_pipeline_family
       AND (s ->> 'name') = p_stage_name
     LIMIT 1;

    IF v_stage IS NULL THEN
        RETURN NULL;
    END IF;

    BEGIN
        v_rounds := (v_stage ->> 'max_tool_rounds')::int;
    EXCEPTION WHEN invalid_text_representation THEN
        v_rounds := NULL;
    END;
    RETURN v_rounds;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 2. Rewrite chat_post_internal — enforce max_tool_rounds on continue.
-- ---------------------------------------------------------------------
-- Count prior assistant rounds in this session. If >= cap, set
-- tools_disabled=true on the payload so the agent must answer.

CREATE OR REPLACE FUNCTION stewards.chat_post_internal(
    p_agent_family text,
    p_model        text,
    p_session_id   text,
    p_provider     text
) RETURNS bigint LANGUAGE plpgsql AS $FN$
DECLARE
    v_body                  jsonb;
    v_payload               jsonb;
    v_work_id               bigint;
    v_inherited_markers     jsonb;
    v_stage_name            text;
    v_pipeline_family       text;
    v_max_rounds            int;
    v_rounds_so_far         int;
    v_force_tools_disabled  boolean := false;
BEGIN
    v_body := stewards.dry_run_chat(p_agent_family, p_model, p_session_id, NULL, p_provider);

    SELECT jsonb_object_agg(je.key, je.value)
      INTO v_inherited_markers
      FROM stewards.work_queue wq
      CROSS JOIN LATERAL jsonb_each(wq.payload) je
     WHERE wq.payload->>'session_id' = p_session_id
       AND wq.kind = 'chat'
       AND wq.id = (
           SELECT max(id) FROM stewards.work_queue
            WHERE payload->>'session_id' = p_session_id
              AND kind = 'chat'
       )
       AND je.key LIKE '\_%' ESCAPE '\';

    -- L.1.1.16: enforce max_tool_rounds.
    v_pipeline_family := v_inherited_markers ->> '_pipeline_family';
    v_stage_name      := v_inherited_markers ->> '_stage_name';

    IF v_pipeline_family IS NOT NULL AND v_stage_name IS NOT NULL THEN
        v_max_rounds := COALESCE(
            stewards.stage_max_tool_rounds(v_pipeline_family, v_stage_name),
            5  -- default cap per L.1.1.16 ratification
        );

        SELECT count(*) INTO v_rounds_so_far
          FROM stewards.messages
         WHERE session_id = p_session_id
           AND role = 'assistant';

        IF v_rounds_so_far >= v_max_rounds THEN
            v_force_tools_disabled := true;
            RAISE NOTICE 'chat_post_internal: session=% rounds=%/% — forcing tools_disabled=true',
                p_session_id, v_rounds_so_far, v_max_rounds;
        END IF;
    END IF;

    v_payload := jsonb_build_object(
        'session_id',      p_session_id,
        'agent_family',    p_agent_family,
        'requested_model', p_model,
        'body',            v_body - '_meta'
    );

    IF v_force_tools_disabled THEN
        v_payload := v_payload || jsonb_build_object('tools_disabled', true);
    END IF;

    IF v_inherited_markers IS NOT NULL THEN
        v_payload := v_payload || v_inherited_markers;
    END IF;

    INSERT INTO stewards.work_queue (kind, provider, payload, status)
    VALUES ('chat', p_provider, v_payload, 'pending')
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 3. Set initial caps on research-write's iterative-prone stages.
-- ---------------------------------------------------------------------
-- Per Michael: start at 5 (or 3 for one-shots). Adjust based on
-- observed behavior after the L.1.1.12 fixes land properly.

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{0,max_tool_rounds}', '5'::jsonb)  -- context_gather
 WHERE family = 'research-write';

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{1,max_tool_rounds}', '5'::jsonb)  -- gather
 WHERE family = 'research-write';

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{2,max_tool_rounds}', '3'::jsonb)  -- synthesize (should be near-1-shot)
 WHERE family = 'research-write';

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{3,max_tool_rounds}', '1'::jsonb)  -- review (no tools anyway, but explicit)
 WHERE family = 'research-write';


-- =====================================================================
-- End of l30-stage-max-tool-rounds.sql
-- =====================================================================
