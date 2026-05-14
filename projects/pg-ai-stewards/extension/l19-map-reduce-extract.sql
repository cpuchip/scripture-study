-- =====================================================================
-- Batch L.1.1.9 — map_reduce_extract_engrams
-- =====================================================================
-- For unattended cases (no live judge present), run engram extraction
-- in parallel over each PARENT chunk of an indexed corpus, then merge
-- the engrams into messages.engrams.items[] keyed by parent_ordinal
-- so they remain distinguishable.
--
-- Uses the existing engram-extractor agent (K.1). Each parent gets one
-- chat. Completion handler apply_map_reduce_parent_engrams writes the
-- engrams back to the message.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. map_reduce_extract_engrams — enqueues one chat per parent.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.map_reduce_extract_engrams(
    p_message_id bigint,
    p_binding    text DEFAULT NULL
) RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_parent     stewards.messages_raw_overflow%ROWTYPE;
    v_agent      stewards.agents%ROWTYPE;
    v_user_msg   text;
    v_body       jsonb;
    v_wq_id      bigint;
    v_count      int := 0;
    v_binding    text;
    v_msg_prefix text;
BEGIN
    -- Default binding from the parent's stored binding_question if not given.
    SELECT binding_question INTO v_binding
      FROM stewards.messages_raw_overflow
     WHERE message_id = p_message_id LIMIT 1;
    v_binding := COALESCE(p_binding, v_binding, 'Extract key facts and findings.');

    SELECT * INTO v_agent FROM stewards.agents
     WHERE family = 'engram-extractor' AND active LIMIT 1;
    IF v_agent.family IS NULL THEN
        RAISE EXCEPTION 'map_reduce_extract_engrams: engram-extractor agent missing';
    END IF;

    -- Ensure session row exists.
    INSERT INTO stewards.sessions (id, kind, label)
    VALUES ('mr-extract-' || p_message_id::text, 'tool',
            'map-reduce engram extraction for message ' || p_message_id::text)
    ON CONFLICT (id) DO NOTHING;

    v_msg_prefix := substring(p_message_id::text FROM 1 FOR 8);

    FOR v_parent IN
        SELECT * FROM stewards.messages_raw_overflow
         WHERE message_id = p_message_id
         ORDER BY parent_ordinal
    LOOP
        v_user_msg :=
            E'BINDING QUESTION:\n' || v_binding ||
            E'\n\nENGRAM ID PREFIX (use this in engram ids): ' || v_msg_prefix || '-p' || v_parent.parent_ordinal::text ||
            E'\n\nNOTE: this is one parent chunk of a larger document. Extract engrams from THIS CHUNK ONLY.' ||
            E'\n\nDOCUMENT CHUNK (' || length(v_parent.content)::text || E' chars):\n---\n' ||
            v_parent.content ||
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

        INSERT INTO stewards.work_queue (kind, provider, payload, status)
        VALUES (
            'chat',
            'opencode_go',
            jsonb_build_object(
                'session_id', 'mr-extract-' || p_message_id::text,
                'agent_family', 'engram-extractor',
                'requested_model', 'deepseek-v4-flash',
                'body', v_body,
                'tools_disabled', true,
                '_map_reduce_extract_target_msg_id', p_message_id,
                '_map_reduce_extract_parent_id',    v_parent.id,
                '_map_reduce_extract_parent_ord',   v_parent.parent_ordinal
            ),
            'pending'
        )
        RETURNING id INTO v_wq_id;
        v_count := v_count + 1;
    END LOOP;

    RETURN jsonb_build_object(
        'message_id', p_message_id,
        'parents_dispatched', v_count,
        'binding', v_binding
    );
END;
$FN$;

COMMENT ON FUNCTION stewards.map_reduce_extract_engrams(bigint, text) IS
'Batch L.1.1.9: enqueue one engram-extractor chat per parent chunk of an indexed corpus. Markers: _map_reduce_extract_target_msg_id, _map_reduce_extract_parent_id, _map_reduce_extract_parent_ord. The completion handler apply_map_reduce_parent_engrams (bgworker-wired) merges results into messages.engrams.items[] with engram ids prefixed by parent_ordinal so duplicates from different parents remain distinguishable.';


-- ---------------------------------------------------------------------
-- 2. apply_map_reduce_parent_engrams — completion handler.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_map_reduce_parent_engrams(
    p_work_queue_id bigint
) RETURNS void LANGUAGE plpgsql AS $FN$
DECLARE
    v_wq            stewards.work_queue%ROWTYPE;
    v_target_msg_id bigint;
    v_parent_id     bigint;
    v_content       text;
    v_extracted     jsonb;
    v_new_items     jsonb;
    v_existing      jsonb;
    v_merged        jsonb;
BEGIN
    SELECT * INTO v_wq FROM stewards.work_queue WHERE id = p_work_queue_id;
    IF v_wq.id IS NULL THEN
        RAISE EXCEPTION 'apply_map_reduce_parent_engrams: wq % not found', p_work_queue_id;
    END IF;

    v_target_msg_id := (v_wq.payload ->> '_map_reduce_extract_target_msg_id')::bigint;
    v_parent_id     := (v_wq.payload ->> '_map_reduce_extract_parent_id')::bigint;

    SELECT m.content INTO v_content
      FROM stewards.messages m
     WHERE m.parent_work_id = p_work_queue_id
       AND m.role = 'assistant'
     ORDER BY m.id DESC LIMIT 1;

    IF v_content IS NULL OR length(v_content) = 0 THEN
        RAISE NOTICE 'apply_map_reduce_parent_engrams: no content for wq=%; skipping parent=%',
            p_work_queue_id, v_parent_id;
        RETURN;
    END IF;

    -- Parse the JSON output. Tolerate either {items:[...]} or [...].
    BEGIN
        v_extracted := v_content::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'apply_map_reduce_parent_engrams: invalid JSON for wq=% parent=%; content head=%',
            p_work_queue_id, v_parent_id, substring(v_content FROM 1 FOR 80);
        RETURN;
    END;

    IF jsonb_typeof(v_extracted) = 'array' THEN
        v_new_items := v_extracted;
    ELSIF jsonb_typeof(v_extracted -> 'items') = 'array' THEN
        v_new_items := v_extracted -> 'items';
    ELSIF jsonb_typeof(v_extracted -> 'engrams') = 'array' THEN
        v_new_items := v_extracted -> 'engrams';
    ELSE
        RAISE NOTICE 'apply_map_reduce_parent_engrams: unexpected shape for wq=% parent=%',
            p_work_queue_id, v_parent_id;
        RETURN;
    END IF;

    -- Merge with existing engrams.items.
    SELECT COALESCE(engrams, '{}'::jsonb) INTO v_existing
      FROM stewards.messages WHERE id = v_target_msg_id;

    v_merged := COALESCE(v_existing -> 'items', '[]'::jsonb) || v_new_items;

    UPDATE stewards.messages
       SET engrams = jsonb_set(COALESCE(engrams, '{}'::jsonb), '{items}', v_merged)
     WHERE id = v_target_msg_id;

    RAISE NOTICE 'apply_map_reduce_parent_engrams: merged % engrams from parent=% into msg=%',
        jsonb_array_length(v_new_items), v_parent_id, v_target_msg_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.apply_map_reduce_parent_engrams(bigint) IS
'Batch L.1.1.9: completion handler for map_reduce_extract_engrams. Parses the engram-extractor JSON output (accepts items/engrams/array shapes), appends to messages.engrams.items[] on the target message. Bgworker fires this when a chat row has _map_reduce_extract_parent_id on its payload.';


-- =====================================================================
-- End of l19-map-reduce-extract.sql
-- =====================================================================
