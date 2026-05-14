-- =====================================================================
-- Batch L.1.1.8 (SQL side) — Judge surface helpers
-- =====================================================================
-- The bridge-side Go intercept calls these to build the rendered
-- judge-surface message it returns to the consuming agent.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. should_intercept_tool_result — single function the bridge calls
--    to ask "is this oversized for this session's consuming agent?"
-- ---------------------------------------------------------------------
-- Returns the intercept-budget-chars threshold; bridge compares with
-- the tool result body length.

CREATE OR REPLACE FUNCTION stewards.intercept_threshold_chars(
    p_session_id text
) RETURNS int LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_budget_tokens int;
    v_chars_per_token constant numeric := 3.5;
    v_intercept_ratio constant numeric := 0.25;
BEGIN
    v_budget_tokens := stewards.effective_budget(p_session_id, NULL);
    IF v_budget_tokens IS NULL OR v_budget_tokens <= 0 THEN
        RETURN 60000;  -- conservative floor
    END IF;
    RETURN (v_budget_tokens::numeric * v_chars_per_token * v_intercept_ratio)::int;
END;
$FN$;

COMMENT ON FUNCTION stewards.intercept_threshold_chars(text) IS
'Batch L.1.1.8: returns the bridge intercept threshold in chars. = effective_budget(session) tokens × 3.5 chars/tok × 0.25 (ratified intercept ratio). Bridge compares tool result length to this; intercepts if exceeded.';


-- ---------------------------------------------------------------------
-- 2. render_judge_surface — given an indexed message_id, produce the
--    rendered surface text (substitutes variables into the template).
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.render_judge_surface(
    p_message_id bigint,
    p_top_overview text DEFAULT NULL
) RETURNS text LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_meta jsonb;
    v_template text;
    v_pipeline text;
    v_msg stewards.messages%ROWTYPE;
    v_tool_name text;
    v_source_bytes int;
    v_parent_count int;
    v_leaf_count int;
    v_binding text;
    v_rendered text;
BEGIN
    -- Look up message + binding + tool.
    SELECT * INTO v_msg FROM stewards.messages WHERE id = p_message_id;
    IF v_msg.id IS NULL THEN
        RAISE EXCEPTION 'render_judge_surface: message % not found', p_message_id;
    END IF;

    -- Aggregate corpus metadata.
    SELECT
        COALESCE(p.tool_name, 'unknown'),
        max(p.byte_size * (1 + p.parent_ordinal))  -- best-effort source bytes; better: SELECT length(content)
    INTO v_tool_name, v_source_bytes
      FROM stewards.messages_raw_overflow p
     WHERE p.message_id = p_message_id
     GROUP BY p.tool_name
     LIMIT 1;
    -- Use actual message content length for source_bytes (more accurate).
    v_source_bytes := length(v_msg.content);

    SELECT count(*), MAX(binding_question)
      INTO v_parent_count, v_binding
      FROM stewards.messages_raw_overflow
     WHERE message_id = p_message_id;

    SELECT count(*) INTO v_leaf_count
      FROM stewards.messages_raw_overflow_leaves
     WHERE message_id = p_message_id;

    -- Resolve pipeline for template override.
    SELECT pipeline_family INTO v_pipeline
      FROM stewards.work_items
     WHERE v_msg.session_id = ANY(session_ids)
     LIMIT 1;

    v_template := stewards.judge_template_for_pipeline(v_pipeline);

    -- Variable substitution. {{tool_name}}, {{source_bytes}}, etc.
    v_rendered := v_template;
    v_rendered := replace(v_rendered, '{{tool_name}}', COALESCE(v_tool_name, 'unknown'));
    v_rendered := replace(v_rendered, '{{source_bytes}}', v_source_bytes::text);
    v_rendered := replace(v_rendered, '{{parent_count}}', v_parent_count::text);
    v_rendered := replace(v_rendered, '{{leaf_count}}', v_leaf_count::text);
    v_rendered := replace(v_rendered, '{{binding_question}}', COALESCE(v_binding, '(no binding recorded)'));
    v_rendered := replace(v_rendered, '{{top_overview}}', COALESCE(p_top_overview, '(no overview generated)'));
    v_rendered := replace(v_rendered, '{{message_id}}', p_message_id::text);

    RETURN v_rendered;
END;
$FN$;

COMMENT ON FUNCTION stewards.render_judge_surface(bigint, text) IS
'Batch L.1.1.8: renders the judge prompt template against an indexed message''s corpus metadata. Variables substituted: tool_name, source_bytes, parent_count, leaf_count, binding_question, top_overview, message_id. Returns the final markdown body the bridge sends as the tool result.';


-- =====================================================================
-- End of l22-judge-surface-helpers.sql
-- =====================================================================
