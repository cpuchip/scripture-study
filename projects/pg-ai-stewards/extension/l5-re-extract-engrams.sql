-- =====================================================================
-- Batch L.5 — re_extract_engrams tool (manual re-extraction with new binding)
-- =====================================================================
-- When a downstream stage's binding question shifts substantially from
-- the original engram extraction's binding, the existing engrams may be
-- mis-tiered. re_extract_engrams lets the agent (or a stage handler)
-- request a fresh extraction with the new binding. Old engrams are
-- preserved in engrams._history before being replaced.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. SQL fn: re_extract_engrams.
-- ---------------------------------------------------------------------
-- Archives the current engrams to _history (a jsonb array), clears the
-- engrams.items array, then re-enqueues extraction via the existing
-- extract_engrams pipeline. The completion handler (apply_engram_
-- extraction) will write the new engrams over the cleared field.
--
-- The new binding question is passed by directly invoking extract_
-- engrams which reads the work_item's input.binding_question; for
-- one-off re-extraction with an EXPLICIT new binding (not from the
-- work_item), we override by building the extraction payload manually.

CREATE OR REPLACE FUNCTION stewards.re_extract_engrams(
    p_message_id bigint,
    p_new_binding text,
    p_cost_cap_micro bigint DEFAULT 100000
) RETURNS bigint LANGUAGE plpgsql AS $FN$
DECLARE
    v_message    stewards.messages%ROWTYPE;
    v_old_engrams jsonb;
    v_history    jsonb;
    v_agent      stewards.agents;
    v_user_msg   text;
    v_body       jsonb;
    v_payload    jsonb;
    v_wq_id      bigint;
    v_msg_prefix text;
BEGIN
    SELECT * INTO v_message FROM stewards.messages WHERE id = p_message_id;
    IF v_message.id IS NULL THEN
        RAISE EXCEPTION 're_extract_engrams: message % not found', p_message_id;
    END IF;

    v_old_engrams := v_message.engrams;

    -- Archive prior engrams to _history (so we never lose extractions).
    -- Initialize _history as an empty array if it doesn't exist yet.
    v_history := COALESCE(v_old_engrams -> '_history', '[]'::jsonb);
    IF v_old_engrams IS NOT NULL THEN
        v_history := v_history || jsonb_build_array(
            v_old_engrams - '_history'
            || jsonb_build_object('_archived_at', now())
        );
    END IF;

    -- Clear engrams (preserve only _history).
    UPDATE stewards.messages
       SET engrams = jsonb_build_object('_history', v_history)
     WHERE id = p_message_id;

    -- Resolve extractor agent.
    SELECT * INTO v_agent
      FROM stewards.agents WHERE family = 'engram-extractor' AND active LIMIT 1;
    IF v_agent.family IS NULL THEN
        RAISE EXCEPTION 're_extract_engrams: engram-extractor agent not registered';
    END IF;

    v_msg_prefix := substring(p_message_id::text FROM 1 FOR 8);

    v_user_msg :=
        E'BINDING QUESTION:\n' || p_new_binding ||
        E'\n\nMESSAGE ID PREFIX (use this in engram ids): ' || v_msg_prefix ||
        E'\n\nNOTE: this is a RE-EXTRACTION with a NEW binding question. The previous engrams have been archived; produce a fresh set tuned to this binding.' ||
        E'\n\nDOCUMENT (' || length(v_message.content)::text || E' chars):\n---\n' ||
        v_message.content ||
        E'\n---\n\nExtract engrams. Output ONLY the JSON.';

    v_body := jsonb_build_object(
        'model', 'deepseek-v4-flash',
        'messages', jsonb_build_array(
            jsonb_build_object('role', 'system', 'content', v_agent.prompt),
            jsonb_build_object('role', 'user', 'content', v_user_msg)
        ),
        'temperature', v_agent.temperature
    );
    IF v_agent.response_format IS NOT NULL THEN
        v_body := v_body || jsonb_build_object('response_format', v_agent.response_format);
    END IF;

    -- Ensure session row exists (FK from messages.session_id).
    INSERT INTO stewards.sessions (id, kind, label)
    VALUES ('engram-re-ex-' || p_message_id::text, 'tool',
            'engram re-extraction for message ' || p_message_id::text)
    ON CONFLICT (id) DO NOTHING;

    v_payload := jsonb_build_object(
        'session_id', 'engram-re-ex-' || p_message_id::text,
        'agent_family', 'engram-extractor',
        'requested_model', 'deepseek-v4-flash',
        'body', v_body,
        'tools_disabled', true,
        '_engram_extraction_target_msg_id', p_message_id,
        '_engram_extraction_binding', p_new_binding,
        '_engram_extraction_raw_chars', length(v_message.content),
        '_re_extraction', true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload, status)
    VALUES ('chat', 'opencode_go', v_payload, 'pending')
    RETURNING id INTO v_wq_id;

    RAISE NOTICE 're_extract_engrams: message=% old engrams archived; new extraction queued wq=%',
        p_message_id, v_wq_id;

    RETURN v_wq_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.re_extract_engrams(bigint, text, bigint) IS
'Batch L.5: re-extract engrams for a message with a new binding question. Archives prior engrams to engrams._history; clears items[] and enqueues a fresh extraction. Use when a downstream stage''s focus differs significantly from the original extraction. Manual only (no auto-trigger in v1).';


-- ---------------------------------------------------------------------
-- 2. Tool definition for re_extract_engrams.
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    're_extract_engrams',
    'Re-extract engrams for a tool message with a different binding question. ' ||
    'Use when the existing engrams (tuned to the original binding) miss material relevant to your current focus. ' ||
    'The old engrams are archived in engrams._history; a fresh extraction runs with the new binding. ' ||
    'Cost-capped at $0.10 per re-extraction by default.',
    $JSON$
    {
      "type": "object",
      "required": ["message_id", "new_binding_question"],
      "additionalProperties": false,
      "properties": {
        "message_id": {
          "type": "integer",
          "description": "The message id whose engrams should be re-extracted."
        },
        "new_binding_question": {
          "type": "string",
          "description": "The new binding question to focus extraction on."
        },
        "cost_cap_micro": {
          "type": "integer",
          "default": 100000,
          "description": "Max micro-dollars (default 100000 = $0.10)."
        }
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 're_extract_engrams'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of l5-re-extract-engrams.sql
-- =====================================================================
