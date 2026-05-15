-- =====================================================================
-- L.1.1.16-followup — tool_choice:"none" reinforces tools_disabled cap
-- =====================================================================
-- Bacteriopolis fix-bundle retry showed: tools_disabled=true was being
-- set on payload at round 6+ of context_gather, but the model
-- (qwen3.6-plus) kept emitting tool_calls for 5 more rounds. Either
-- bgworker isn't stripping tools from body, or qwen hallucinates
-- tool_calls when no tools are defined.
--
-- Stronger signal: when tools_disabled=true, ALSO inject tool_choice:
-- "none" directly into the body. OpenAI/Anthropic-compatible API
-- semantics: tool_choice="none" explicitly forbids the model from
-- calling tools. Defense in depth.
-- =====================================================================


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

    v_pipeline_family := v_inherited_markers ->> '_pipeline_family';
    v_stage_name      := v_inherited_markers ->> '_stage_name';

    IF v_pipeline_family IS NOT NULL AND v_stage_name IS NOT NULL THEN
        v_max_rounds := COALESCE(
            stewards.stage_max_tool_rounds(v_pipeline_family, v_stage_name),
            5
        );

        SELECT count(*) INTO v_rounds_so_far
          FROM stewards.messages
         WHERE session_id = p_session_id
           AND role = 'assistant';

        IF v_rounds_so_far >= v_max_rounds THEN
            v_force_tools_disabled := true;
            RAISE NOTICE 'chat_post_internal: session=% rounds=%/% — forcing tools_disabled+tool_choice=none',
                p_session_id, v_rounds_so_far, v_max_rounds;
        END IF;
    END IF;

    -- Strip _meta and (when capped) inject tool_choice="none" directly
    -- into the body. Defense in depth alongside the bgworker's tools-
    -- stripping path.
    v_body := v_body - '_meta';
    IF v_force_tools_disabled THEN
        v_body := v_body || jsonb_build_object('tool_choice', 'none');
    END IF;

    v_payload := jsonb_build_object(
        'session_id',      p_session_id,
        'agent_family',    p_agent_family,
        'requested_model', p_model,
        'body',            v_body
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


-- =====================================================================
-- End of l31-tool-choice-none-on-cap.sql
-- =====================================================================
