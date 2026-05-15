-- =====================================================================
-- ES.1.s2 — chunk_and_index circuit breaker (CF-5)
-- =====================================================================
-- A 900K-char fetch produced ~501 leaves → 501 synchronous
-- contextualizer chats. No ceiling. This breaker caps the work:
--
--   If a source would produce more than 40 leaves, chunk_and_index
--   REFUSES — it returns {skipped:true, ...} without creating any
--   parents/leaves or enqueueing any contextualizer chats.
--
-- intercept_oversized_tool_after handles the skip by leaving the
-- message raw (the pre-L.1.1.8 K.1/L.1 paths handle oversized raw
-- messages) and logging a flag. ES.3's judge-read rearchitecture is
-- the real fix for genuinely-huge inputs.
--
-- Stopgap by design — chunk_and_index is likely retired wholesale in
-- ES.3. Minimal guard, no elaborate machinery.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. chunk_and_index — add the 40-leaf ceiling guard.
-- ---------------------------------------------------------------------
-- Only the guard is new; the body is the L.1.1.13 sha256 version with
-- a pre-flight projected-leaf check prepended.

CREATE OR REPLACE FUNCTION stewards.chunk_and_index(
    p_message_id      bigint,
    p_binding_question text DEFAULT NULL,
    p_parent_chars    int  DEFAULT 14000,
    p_leaf_chars      int  DEFAULT 1800,
    p_parent_overlap  int  DEFAULT 400,
    p_leaf_overlap    int  DEFAULT 64
) RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_message     stewards.messages%ROWTYPE;
    v_remaining   text;
    v_parent_text text;
    v_parent_consumed int;
    v_parent_id   bigint;
    v_parent_ord  int := 0;
    v_leaf_remaining text;
    v_leaf_text   text;
    v_leaf_consumed int;
    v_leaf_id     bigint;
    v_leaf_ord    int;
    v_parent_count int := 0;
    v_leaf_count   int := 0;
    v_tool_name    text;
    v_content_sha  text;
    v_projected_leaves int;
    -- ES.1.s2 ratified ceiling.
    v_leaf_ceiling constant int := 40;
BEGIN
    SELECT * INTO v_message FROM stewards.messages WHERE id = p_message_id;
    IF v_message.id IS NULL THEN
        RAISE EXCEPTION 'chunk_and_index: message % not found', p_message_id;
    END IF;

    IF v_message.content IS NULL OR length(v_message.content) = 0 THEN
        RAISE EXCEPTION 'chunk_and_index: message % has no content', p_message_id;
    END IF;

    -- ES.1.s2 CIRCUIT BREAKER: project leaf count; refuse if over ceiling.
    v_projected_leaves := ceil(
        length(v_message.content)::numeric
        / GREATEST(p_leaf_chars - p_leaf_overlap, 1)
    )::int;

    IF v_projected_leaves > v_leaf_ceiling THEN
        RAISE NOTICE 'chunk_and_index: CIRCUIT BREAKER — msg=% would produce ~% leaves (ceiling %); REFUSING to chunk. Flagged for ES.3 judge.',
            p_message_id, v_projected_leaves, v_leaf_ceiling;
        RETURN jsonb_build_object(
            'skipped', true,
            'reason', 'exceeds-leaf-ceiling',
            'message_id', p_message_id,
            'projected_leaves', v_projected_leaves,
            'leaf_ceiling', v_leaf_ceiling,
            'source_bytes', length(v_message.content)
        );
    END IF;

    v_tool_name   := stewards.tool_name_for_tool_call_id(v_message.session_id, v_message.tool_call_id);
    v_content_sha := encode(digest(v_message.content, 'sha256'), 'hex');

    v_remaining := v_message.content;

    WHILE length(v_remaining) > 0 LOOP
        SELECT chunk, consumed INTO v_parent_text, v_parent_consumed
          FROM stewards.split_one_chunk(v_remaining, p_parent_chars) LIMIT 1;
        IF v_parent_consumed = 0 THEN EXIT; END IF;

        INSERT INTO stewards.messages_raw_overflow
            (message_id, parent_ordinal, content, byte_size, tool_name, binding_question, content_sha256)
        VALUES
            (p_message_id, v_parent_ord, v_parent_text, length(v_parent_text), v_tool_name, p_binding_question, v_content_sha)
        RETURNING id INTO v_parent_id;
        v_parent_count := v_parent_count + 1;

        v_leaf_remaining := v_parent_text;
        v_leaf_ord := 0;
        WHILE length(v_leaf_remaining) > 0 LOOP
            SELECT chunk, consumed INTO v_leaf_text, v_leaf_consumed
              FROM stewards.split_one_chunk(v_leaf_remaining, p_leaf_chars) LIMIT 1;
            IF v_leaf_consumed = 0 THEN EXIT; END IF;

            INSERT INTO stewards.messages_raw_overflow_leaves
                (message_id, parent_id, leaf_ordinal, content, byte_size)
            VALUES
                (p_message_id, v_parent_id, v_leaf_ord, v_leaf_text, length(v_leaf_text))
            RETURNING id INTO v_leaf_id;
            v_leaf_count := v_leaf_count + 1;
            v_leaf_ord := v_leaf_ord + 1;

            PERFORM stewards.contextualize_leaf(v_leaf_id);

            IF v_leaf_consumed > p_leaf_overlap AND length(v_leaf_remaining) > v_leaf_consumed THEN
                v_leaf_remaining := substring(v_leaf_remaining FROM (v_leaf_consumed - p_leaf_overlap + 1));
            ELSE
                EXIT;
            END IF;
        END LOOP;

        v_parent_ord := v_parent_ord + 1;
        IF v_parent_consumed > p_parent_overlap AND length(v_remaining) > v_parent_consumed THEN
            v_remaining := substring(v_remaining FROM (v_parent_consumed - p_parent_overlap + 1));
        ELSE
            EXIT;
        END IF;
    END LOOP;

    RETURN jsonb_build_object(
        'message_id', p_message_id,
        'parent_count', v_parent_count,
        'leaf_count', v_leaf_count,
        'source_bytes', length(v_message.content),
        'tool_name', v_tool_name,
        'content_sha256', v_content_sha
    );
END;
$FN$;


-- ---------------------------------------------------------------------
-- 2. intercept_oversized_tool_after — handle the skipped result.
-- ---------------------------------------------------------------------
-- When chunk_and_index returns {skipped:true}, do NOT build a corpus
-- surface. Leave the message raw (K.1 engram extraction + L.1
-- graduated rendering handle oversized raw messages the prior way),
-- and log the flag. ES.3 will query for these.

CREATE OR REPLACE FUNCTION stewards.intercept_oversized_tool_after()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_threshold     int;
    v_binding       text;
    v_first_parent  text;
    v_top_overview  text;
    v_surface       text;
    v_index_result  jsonb;
    v_content_sha   text;
    v_prior_msg_id  bigint;
BEGIN
    v_threshold := stewards.intercept_threshold_chars(NEW.session_id);

    IF NEW.content LIKE '%[CORPUS-INDEXED]%' THEN RETURN NEW; END IF;
    IF NEW.role <> 'tool' THEN RETURN NEW; END IF;
    IF length(NEW.content) <= v_threshold THEN RETURN NEW; END IF;

    v_content_sha := encode(digest(NEW.content, 'sha256'), 'hex');
    v_prior_msg_id := stewards.source_sha256_already_indexed_in_session(NEW.session_id, v_content_sha);

    IF v_prior_msg_id IS NOT NULL THEN
        v_surface := E'[CORPUS-INDEXED]\n\n' ||
            E'**Duplicate content detected** — this tool result is byte-identical to a prior tool result (message id ' || v_prior_msg_id::text ||
            E') already indexed in this session. Use `read_corpus_parents(message_id=' || v_prior_msg_id::text || E')` to read the existing corpus instead of re-indexing.';
        UPDATE stewards.messages SET content = v_surface WHERE id = NEW.id;
        RETURN NEW;
    END IF;

    SELECT input ->> 'binding_question' INTO v_binding
      FROM stewards.work_items
     WHERE NEW.session_id = ANY(session_ids)
     LIMIT 1;

    BEGIN
        v_index_result := stewards.chunk_and_index(NEW.id, v_binding);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'intercept_oversized_tool_after: chunk_and_index failed for msg=%: %; leaving raw',
            NEW.id, SQLERRM;
        RETURN NEW;
    END;

    -- ES.1.s2: circuit breaker tripped — leave the message raw, flag it.
    IF COALESCE((v_index_result ->> 'skipped')::boolean, false) THEN
        RAISE NOTICE 'intercept_oversized_tool_after: msg=% NOT indexed (% — projected % leaves). Left raw; flagged for ES.3 judge.',
            NEW.id, v_index_result ->> 'reason', v_index_result ->> 'projected_leaves';
        RETURN NEW;  -- raw message stays; K.1/L.1 handle it the prior way
    END IF;

    SELECT substring(content FROM 1 FOR 500) INTO v_first_parent
      FROM stewards.messages_raw_overflow
     WHERE message_id = NEW.id
     ORDER BY parent_ordinal ASC
     LIMIT 1;

    v_top_overview := COALESCE(v_first_parent, '(no parent content)') ||
        CASE WHEN length(v_first_parent) >= 500 THEN '…' ELSE '' END;

    v_surface := E'[CORPUS-INDEXED]\n\n' || stewards.render_judge_surface(NEW.id, v_top_overview);
    UPDATE stewards.messages SET content = v_surface WHERE id = NEW.id;

    RAISE NOTICE 'intercept_oversized_tool_after: msg=% indexed (% parents, % leaves)',
        NEW.id, v_index_result ->> 'parent_count', v_index_result ->> 'leaf_count';

    RETURN NEW;
END;
$FN$;


-- =====================================================================
-- End of es3-chunk-index-circuit-breaker.sql
-- =====================================================================
