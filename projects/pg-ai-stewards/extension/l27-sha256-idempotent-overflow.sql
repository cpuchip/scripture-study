-- =====================================================================
-- L.1.1.13 — sha256 idempotent overflow indexing
-- =====================================================================
-- Requires pgcrypto for digest(text, sha256). Enable if absent.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================================
-- Bacteriopolis re-fetched the same Exploratorium URL in synthesize
-- (after the missing retrieve_from_corpus tool blocked it from reading
-- the gather-stage corpus). Both fetches produced identical 247K-char
-- content; both triggered L.1.1.8; both indexed into 18 parents/160
-- leaves; both fired 160 contextualizer chats.
--
-- Fix: hash the inbound content before chunk_and_index. If the same
-- session already indexed the identical content, skip — point the new
-- message at the existing parents/leaves instead of re-indexing.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Add content_sha256 column to messages_raw_overflow.
-- ---------------------------------------------------------------------

ALTER TABLE stewards.messages_raw_overflow
    ADD COLUMN IF NOT EXISTS content_sha256 text;

-- Index for fast lookup. Per-session uniqueness via partial index since
-- the same content can legitimately appear in different sessions.
CREATE INDEX IF NOT EXISTS messages_raw_overflow_sha_session
    ON stewards.messages_raw_overflow (content_sha256, message_id);

COMMENT ON COLUMN stewards.messages_raw_overflow.content_sha256 IS
'Batch L.1.1.13: sha256 of the SOURCE content (the original tool message body, NOT the parent chunk). Same value across all parents of the same source. Used by intercept_oversized_tool_after to detect duplicate fetches.';


-- ---------------------------------------------------------------------
-- 2. Helper: source_sha256_already_indexed_in_session.
-- ---------------------------------------------------------------------
-- Returns the prior message_id whose corpus we can reuse, or NULL.

CREATE OR REPLACE FUNCTION stewards.source_sha256_already_indexed_in_session(
    p_session_id text,
    p_sha256     text
) RETURNS bigint LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_prior_msg_id bigint;
BEGIN
    SELECT p.message_id INTO v_prior_msg_id
      FROM stewards.messages_raw_overflow p
      JOIN stewards.messages m ON m.id = p.message_id
     WHERE p.content_sha256 = p_sha256
       AND m.session_id = p_session_id
       AND p.parent_ordinal = 0  -- one row check sufficient (all parents share sha256)
     LIMIT 1;
    RETURN v_prior_msg_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.source_sha256_already_indexed_in_session(text, text) IS
'Batch L.1.1.13: return the prior message_id in this session whose corpus has matching content_sha256, or NULL.';


-- ---------------------------------------------------------------------
-- 3. Update chunk_and_index to record content_sha256 on each parent.
-- ---------------------------------------------------------------------
-- Recompute content sha once at the top, write it to every parent row.

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
BEGIN
    SELECT * INTO v_message FROM stewards.messages WHERE id = p_message_id;
    IF v_message.id IS NULL THEN
        RAISE EXCEPTION 'chunk_and_index: message % not found', p_message_id;
    END IF;

    IF v_message.content IS NULL OR length(v_message.content) = 0 THEN
        RAISE EXCEPTION 'chunk_and_index: message % has no content', p_message_id;
    END IF;

    v_tool_name   := stewards.tool_name_for_tool_call_id(v_message.session_id, v_message.tool_call_id);
    v_content_sha := encode(digest(v_message.content, 'sha256'), 'hex');

    v_remaining := v_message.content;

    WHILE length(v_remaining) > 0 LOOP
        SELECT chunk, consumed INTO v_parent_text, v_parent_consumed
          FROM stewards.split_one_chunk(v_remaining, p_parent_chars) LIMIT 1;

        IF v_parent_consumed = 0 THEN
            EXIT;
        END IF;

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

            IF v_leaf_consumed = 0 THEN
                EXIT;
            END IF;

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
-- 4. Update intercept_oversized_tool_after — sha-check before indexing.
-- ---------------------------------------------------------------------

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

    -- L.1.1.13: sha-check before incurring index cost.
    v_content_sha := encode(digest(NEW.content, 'sha256'), 'hex');
    v_prior_msg_id := stewards.source_sha256_already_indexed_in_session(NEW.session_id, v_content_sha);

    IF v_prior_msg_id IS NOT NULL THEN
        -- Reuse the prior corpus. Build a surface that points at it.
        RAISE NOTICE 'intercept_oversized_tool_after: msg=% sha matches prior msg=% in session=%; reusing corpus',
            NEW.id, v_prior_msg_id, NEW.session_id;

        v_surface := E'[CORPUS-INDEXED]\n\n' ||
            E'**Duplicate content detected** — this tool result is byte-identical to a prior tool result (message id ' || v_prior_msg_id::text ||
            E') already indexed in this session. Use `read_corpus_parents(message_id=' || v_prior_msg_id::text || E')` to read the existing corpus instead of re-indexing.\n\n' ||
            E'If you intended the re-fetch to surface NEW content, the source returned the same bytes — your re-fetch was wasted. Adjust your retrieval strategy.';

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

    SELECT substring(content FROM 1 FOR 500) INTO v_first_parent
      FROM stewards.messages_raw_overflow
     WHERE message_id = NEW.id
     ORDER BY parent_ordinal ASC
     LIMIT 1;

    v_top_overview := COALESCE(v_first_parent, '(no parent content)') ||
        CASE WHEN length(v_first_parent) >= 500 THEN '…' ELSE '' END;

    v_surface := E'[CORPUS-INDEXED]\n\n' || stewards.render_judge_surface(NEW.id, v_top_overview);

    UPDATE stewards.messages SET content = v_surface WHERE id = NEW.id;

    RAISE NOTICE 'intercept_oversized_tool_after: msg=% sha=% indexed (% parents, % leaves) surface replaces % chars',
        NEW.id,
        substring(v_content_sha FROM 1 FOR 8),
        v_index_result ->> 'parent_count',
        v_index_result ->> 'leaf_count',
        v_index_result ->> 'source_bytes';

    RETURN NEW;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 5. Backfill existing messages_raw_overflow rows with content_sha256.
-- ---------------------------------------------------------------------
-- Use the parent's own content as a placeholder identifier (we don't
-- have the original message content for old rows since intercept may
-- have replaced messages.content already). For new inserts the chunk_
-- and_index path computes from messages.content properly.

UPDATE stewards.messages_raw_overflow p
   SET content_sha256 = encode(digest(
       (SELECT string_agg(content, '' ORDER BY parent_ordinal)
          FROM stewards.messages_raw_overflow
         WHERE message_id = p.message_id),
       'sha256'), 'hex')
 WHERE content_sha256 IS NULL;


-- =====================================================================
-- End of l27-sha256-idempotent-overflow.sql
-- =====================================================================
