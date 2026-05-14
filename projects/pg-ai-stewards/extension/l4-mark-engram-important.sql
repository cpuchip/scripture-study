-- =====================================================================
-- Batch L.4 — mark_engram_important write-side (MCP tool registration)
-- =====================================================================
-- The read-side (compose_messages anchoring important engrams at HOT
-- through pressure) ships in L.1. L.4 ships the write-side: a SQL
-- function and MCP tool the agent can call to flag a specific engram.
--
-- mark_engram_important(message_id, engram_id, important=true)
-- updates the items[] array element where id matches by rebuilding the
-- array with the target item's is_important set.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. SQL fn: mark_engram_important.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.mark_engram_important(
    p_message_id bigint,
    p_engram_id  text,
    p_important  boolean DEFAULT true
) RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_engrams jsonb;
    v_items   jsonb;
    v_new_items jsonb := '[]'::jsonb;
    v_item    jsonb;
    v_found   boolean := false;
BEGIN
    SELECT engrams INTO v_engrams FROM stewards.messages WHERE id = p_message_id;

    IF v_engrams IS NULL THEN
        RAISE EXCEPTION 'mark_engram_important: message % has no engrams', p_message_id;
    END IF;

    v_items := COALESCE(v_engrams -> 'items', '[]'::jsonb);

    FOR v_item IN SELECT * FROM jsonb_array_elements(v_items) LOOP
        IF (v_item ->> 'id') = p_engram_id THEN
            v_new_items := v_new_items || jsonb_build_array(
                v_item || jsonb_build_object('is_important', p_important)
            );
            v_found := true;
        ELSE
            v_new_items := v_new_items || jsonb_build_array(v_item);
        END IF;
    END LOOP;

    IF NOT v_found THEN
        RAISE EXCEPTION 'mark_engram_important: no engram with id=% on message %',
            p_engram_id, p_message_id;
    END IF;

    v_engrams := jsonb_set(v_engrams, '{items}', v_new_items);

    UPDATE stewards.messages SET engrams = v_engrams WHERE id = p_message_id;

    RETURN jsonb_build_object(
        'message_id', p_message_id,
        'engram_id', p_engram_id,
        'is_important', p_important,
        'total_engrams', jsonb_array_length(v_new_items)
    );
END;
$FN$;

COMMENT ON FUNCTION stewards.mark_engram_important(bigint, text, boolean) IS
'Batch L.4 write-side: flag a specific engram (by message_id + engram_id) as is_important. Read-side in compose_messages (L.1) anchors important engrams at HOT through pressure — only crisis can drop them, and even then they emit first. Pass p_important=false to clear the flag.';


-- ---------------------------------------------------------------------
-- 2. Tool definition for mark_engram_important.
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    'mark_engram_important',
    'Flag a specific engram (by message_id + engram_id) as is_important. ' ||
    'Important engrams are anchored at HOT through context pressure — they survive all pressure thresholds except crisis, and even then they emit first. ' ||
    'Use this when an engram contains a quote, URL, date, or claim you''ll cite later and can''t afford to lose under compaction. ' ||
    'Pass important=false to clear the flag.',
    $JSON$
    {
      "type": "object",
      "required": ["message_id", "engram_id"],
      "additionalProperties": false,
      "properties": {
        "message_id": {
          "type": "integer",
          "description": "The message id from the engram block header in active context."
        },
        "engram_id": {
          "type": "string",
          "description": "The engram's id (e.g. 'msg-2381-e3') from the engram you want to mark."
        },
        "important": {
          "type": "boolean",
          "default": true,
          "description": "true to mark important (default); false to clear the flag."
        }
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 'mark_engram_important'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of l4-mark-engram-important.sql
-- =====================================================================
