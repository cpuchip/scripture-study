-- =====================================================================
-- Batch L.1.1.6 — chunk_and_index orchestrator
-- =====================================================================
-- Takes an oversized message and:
--   1. Splits content into 4K-token (~14K chars) parents with 400-char overlap
--   2. Splits each parent into 512-token (~1800 chars) leaves with 64-char overlap
--   3. INSERTs parents into messages_raw_overflow
--   4. INSERTs leaves into messages_raw_overflow_leaves
--   5. Enqueues contextualize_leaf for each leaf (which itself queues embed)
--   6. Returns counts + cost estimate
--
-- The character splitter is a simple paragraph-aware splitter: prefers
-- paragraph breaks ("\n\n") within the last 30% of the window, then
-- newlines, then word boundaries, then hard cut. Carry-forward:
-- upgrade to a fuller recursive splitter if signal warrants (research
-- says fixed-size with overlap is already 80%+ of the win).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: find_last_break_pos.
-- ---------------------------------------------------------------------
-- Returns the position of the LAST occurrence of breaker within the
-- first p_max chars of p_text, but only if it occurs in the last
-- p_lookback fraction (default 30%) of the window. Returns 0 if no
-- suitable break found.

CREATE OR REPLACE FUNCTION stewards.find_last_break_pos(
    p_text text,
    p_max int,
    p_breaker text,
    p_lookback_frac numeric DEFAULT 0.3
) RETURNS int LANGUAGE plpgsql IMMUTABLE AS $FN$
DECLARE
    v_window text;
    v_rev_window text;
    v_rev_breaker text;
    v_pos_from_end int;
    v_pos_from_start int;
    v_min_pos int;
BEGIN
    IF p_text IS NULL OR length(p_text) <= p_max THEN
        RETURN COALESCE(length(p_text), 0);
    END IF;
    IF p_breaker = '' THEN
        RETURN 0;
    END IF;

    v_window := substring(p_text FROM 1 FOR p_max);
    v_rev_window := reverse(v_window);
    v_rev_breaker := reverse(p_breaker);
    v_pos_from_end := position(v_rev_breaker IN v_rev_window);

    IF v_pos_from_end = 0 THEN
        RETURN 0;
    END IF;

    -- Convert reversed position to forward position (start of breaker).
    v_pos_from_start := length(v_window) - v_pos_from_end - length(p_breaker) + 2;

    -- Honor lookback constraint: only accept if the break is in the
    -- last p_lookback_frac of the window (avoids tiny early chunks).
    v_min_pos := (p_max * (1.0 - p_lookback_frac))::int;
    IF v_pos_from_start < v_min_pos THEN
        RETURN 0;
    END IF;

    -- Position of end of breaker (where the cut happens).
    RETURN v_pos_from_start + length(p_breaker) - 1;
END;
$FN$;

COMMENT ON FUNCTION stewards.find_last_break_pos(text, int, text, numeric) IS
'Batch L.1.1.6: paragraph-aware split helper. Returns end-of-breaker position of the LAST occurrence of p_breaker within first p_max chars of p_text, but only if in the last p_lookback_frac of the window. 0 if not found.';


-- ---------------------------------------------------------------------
-- 2. Helper: split_one_chunk — returns next chunk + new remaining.
-- ---------------------------------------------------------------------
-- Tries breakers in order: \n\n, then \n, then space, then hard cut.
-- Returns the chunk text and how many chars were consumed (the caller
-- subtracts overlap when advancing).

CREATE OR REPLACE FUNCTION stewards.split_one_chunk(
    p_text text,
    p_max int
) RETURNS TABLE(chunk text, consumed int) LANGUAGE plpgsql IMMUTABLE AS $FN$
DECLARE
    v_pos int;
BEGIN
    IF p_text IS NULL OR length(p_text) = 0 THEN
        chunk := '';
        consumed := 0;
        RETURN NEXT;
        RETURN;
    END IF;

    IF length(p_text) <= p_max THEN
        chunk := p_text;
        consumed := length(p_text);
        RETURN NEXT;
        RETURN;
    END IF;

    -- Try paragraph break.
    v_pos := stewards.find_last_break_pos(p_text, p_max, E'\n\n', 0.4);
    IF v_pos = 0 THEN
        v_pos := stewards.find_last_break_pos(p_text, p_max, E'\n', 0.3);
    END IF;
    IF v_pos = 0 THEN
        v_pos := stewards.find_last_break_pos(p_text, p_max, ' ', 0.2);
    END IF;
    IF v_pos = 0 THEN
        v_pos := p_max;  -- hard cut
    END IF;

    chunk := substring(p_text FROM 1 FOR v_pos);
    consumed := v_pos;
    RETURN NEXT;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 3. chunk_and_index — the orchestrator.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.chunk_and_index(
    p_message_id      bigint,
    p_binding_question text DEFAULT NULL,
    p_parent_chars    int  DEFAULT 14000,     -- ~4K tokens
    p_leaf_chars      int  DEFAULT 1800,      -- ~512 tokens
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
BEGIN
    SELECT * INTO v_message FROM stewards.messages WHERE id = p_message_id;
    IF v_message.id IS NULL THEN
        RAISE EXCEPTION 'chunk_and_index: message % not found', p_message_id;
    END IF;

    IF v_message.content IS NULL OR length(v_message.content) = 0 THEN
        RAISE EXCEPTION 'chunk_and_index: message % has no content', p_message_id;
    END IF;

    -- Resolve tool name via L.8 helper for denormalization.
    v_tool_name := stewards.tool_name_for_tool_call_id(v_message.session_id, v_message.tool_call_id);

    v_remaining := v_message.content;

    WHILE length(v_remaining) > 0 LOOP
        SELECT chunk, consumed INTO v_parent_text, v_parent_consumed
          FROM stewards.split_one_chunk(v_remaining, p_parent_chars) LIMIT 1;

        IF v_parent_consumed = 0 THEN
            EXIT;
        END IF;

        INSERT INTO stewards.messages_raw_overflow
            (message_id, parent_ordinal, content, byte_size, tool_name, binding_question)
        VALUES
            (p_message_id, v_parent_ord, v_parent_text, length(v_parent_text), v_tool_name, p_binding_question)
        RETURNING id INTO v_parent_id;
        v_parent_count := v_parent_count + 1;

        -- Now split this parent into leaves.
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

            -- Enqueue contextualization for this leaf.
            PERFORM stewards.contextualize_leaf(v_leaf_id);

            -- Advance with overlap.
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
        'tool_name', v_tool_name
    );
END;
$FN$;

COMMENT ON FUNCTION stewards.chunk_and_index(bigint, text, int, int, int, int) IS
'Batch L.1.1.6: orchestrator. Splits message content into 4K-token parents and 512-token leaves with overlap; inserts both into messages_raw_overflow / _leaves; enqueues contextualize_leaf for each leaf (which then enqueues embed). Returns counts + source byte size. Defaults: parent_chars=14000, leaf_chars=1800, parent_overlap=400, leaf_overlap=64.';


-- =====================================================================
-- End of l16-chunk-and-index.sql
-- =====================================================================
