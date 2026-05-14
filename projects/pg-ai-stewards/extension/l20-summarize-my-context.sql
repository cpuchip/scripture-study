-- =====================================================================
-- Batch L.1.1.10 — Subagent self-window-management
-- =====================================================================
-- summarize_my_context() lets a subagent trigger re-extraction of its
-- OWN session's heavy tool messages with a fresher binding question.
-- Useful when the agent realizes mid-session that their context has
-- shifted and the original engrams (extracted under a different
-- binding) are no longer optimal.
--
-- Pure SQL fn + tool_def. The agent invokes this with the binding
-- they want; we loop over their session's tool messages that exceed
-- the current effective_extraction_threshold and enqueue re-extraction
-- (uses L.5's re_extract_engrams).
-- =====================================================================


CREATE OR REPLACE FUNCTION stewards.summarize_my_context(
    p_session_id  text,
    p_new_binding text,
    p_max_messages int DEFAULT 10
) RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_msg         stewards.messages%ROWTYPE;
    v_threshold   int;
    v_wq_id       bigint;
    v_dispatched  int := 0;
    v_total_chars bigint := 0;
BEGIN
    v_threshold := stewards.effective_extraction_threshold(p_session_id);

    FOR v_msg IN
        SELECT *
          FROM stewards.messages
         WHERE session_id = p_session_id
           AND role = 'tool'
           AND length(content) > v_threshold
         ORDER BY id DESC
         LIMIT p_max_messages
    LOOP
        v_total_chars := v_total_chars + length(v_msg.content);
        v_wq_id := stewards.re_extract_engrams(v_msg.id, p_new_binding);
        v_dispatched := v_dispatched + 1;
    END LOOP;

    RETURN jsonb_build_object(
        'session_id', p_session_id,
        'new_binding', p_new_binding,
        'threshold_chars', v_threshold,
        'messages_dispatched', v_dispatched,
        'total_chars_processed', v_total_chars
    );
END;
$FN$;

COMMENT ON FUNCTION stewards.summarize_my_context(text, text, int) IS
'Batch L.1.1.10: subagent self-window-management. Loops over the session''s tool messages that exceed the current effective_extraction_threshold and re-extracts engrams with the new binding (via L.5''s re_extract_engrams). Limits to p_max_messages most-recent to bound cost.';


INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    'summarize_my_context',
    'Re-extract engrams across your own session''s heavy tool messages with a new binding question. ' ||
    'Use when you realize mid-session that your current focus has shifted and the prior engrams (tuned ' ||
    'to an earlier binding) are no longer optimal. Each message above the agent-aware extraction ' ||
    'threshold gets re-extracted; old engrams archived to engrams._history.',
    $JSON$
    {
      "type": "object",
      "required": ["new_binding"],
      "additionalProperties": false,
      "properties": {
        "new_binding": {
          "type": "string",
          "description": "The new binding question to focus re-extraction on."
        },
        "max_messages": {
          "type": "integer",
          "default": 10,
          "description": "Max number of heavy messages to re-extract this call (bounds cost)."
        }
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 'summarize_my_context'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of l20-summarize-my-context.sql
-- =====================================================================
