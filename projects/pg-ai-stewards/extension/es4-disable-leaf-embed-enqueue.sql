-- =====================================================================
-- ES.2/CF-2 (Option B) — Disable leaf embed enqueue
-- =====================================================================
-- CF-2: messages_raw_overflow_leaves.id is bigserial; the bgworker
-- embed handler's `WHERE id = $1` binds target_id as text and crashes
-- (`operator does not exist: bigint = text`).
--
-- Ratified Option B: rather than a text-id cascade through
-- chunk_and_index / contextualize_leaf / apply_contextualize_leaf /
-- tool_defs (ES.3's judge-brief rearchitecture will likely delete leaf
-- embedding entirely), simply STOP enqueueing embed jobs for leaves.
--
-- apply_contextualize_leaf still writes context_prefix to each leaf —
-- read_corpus_parents (paginated, no vectors) keeps working. Only the
-- embed INSERT is removed. retrieve_with_merge (vector search) goes
-- dormant until ES.3 decides the leaf-embedding architecture.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.apply_contextualize_leaf(
    p_work_queue_id bigint
) RETURNS void LANGUAGE plpgsql AS $FN$
DECLARE
    v_wq         stewards.work_queue%ROWTYPE;
    v_leaf_id    bigint;
    v_content    text;
    v_reasoning  text;
BEGIN
    SELECT * INTO v_wq FROM stewards.work_queue WHERE id = p_work_queue_id;
    IF v_wq.id IS NULL THEN
        RAISE EXCEPTION 'apply_contextualize_leaf: wq % not found', p_work_queue_id;
    END IF;

    v_leaf_id := (v_wq.payload ->> '_contextualize_leaf_id')::bigint;
    IF v_leaf_id IS NULL THEN
        RAISE EXCEPTION 'apply_contextualize_leaf: missing _contextualize_leaf_id on wq %', p_work_queue_id;
    END IF;

    SELECT m.content, m.reasoning_content INTO v_content, v_reasoning
      FROM stewards.messages m
     WHERE m.parent_work_id = p_work_queue_id
       AND m.role = 'assistant'
     ORDER BY m.id DESC LIMIT 1;

    -- L.1.1.12: fall back to reasoning_content when content is empty.
    IF v_content IS NULL OR length(v_content) = 0 THEN
        IF v_reasoning IS NOT NULL AND length(v_reasoning) > 0 THEN
            v_content := v_reasoning;
            RAISE NOTICE 'apply_contextualize_leaf: leaf=% empty content; using reasoning_content (% chars)',
                v_leaf_id, length(v_reasoning);
        ELSE
            RAISE NOTICE 'apply_contextualize_leaf: no content for wq=%; leaving leaf=% uncontextualized',
                p_work_queue_id, v_leaf_id;
            RETURN;
        END IF;
    END IF;

    IF length(v_content) > 500 THEN
        v_content := substring(v_content FROM 1 FOR 500);
    END IF;

    UPDATE stewards.messages_raw_overflow_leaves
       SET context_prefix = v_content
     WHERE id = v_leaf_id;

    -- ES.2/CF-2 Option B: leaf embed enqueue REMOVED. The bgworker
    -- embed handler's `WHERE id = $1` crashes on this table's bigserial
    -- id (`bigint = text`). Leaves keep context_prefix; vector search
    -- over leaves is dormant until ES.3 settles the architecture.
    -- (Prior code here: INSERT INTO work_queue (kind='embed', ...).)

    RAISE NOTICE 'apply_contextualize_leaf: leaf=% prefix written (% chars); embed skipped (ES.2/CF-2 Option B)',
        v_leaf_id, length(v_content);
END;
$FN$;

COMMENT ON FUNCTION stewards.apply_contextualize_leaf(bigint) IS
'ES.2/CF-2 Option B: writes context_prefix to a leaf from the contextualizer chat result (reasoning_content fallback per L.1.1.12). Leaf embed enqueue REMOVED — messages_raw_overflow_leaves.id is bigserial and the embed handler crashes on bigint=text. ES.3 decides whether leaf embedding returns at all.';

-- =====================================================================
-- End of es4-disable-leaf-embed-enqueue.sql
-- =====================================================================
