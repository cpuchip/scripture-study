-- =====================================================================
-- Batch L.1.1.7 — Auto-merging retrieval over overflow leaves
-- =====================================================================
-- Vector search hits leaves; if merge_threshold (default 3) leaves
-- under the same parent score, return the parent instead. LlamaIndex
-- AutoMergingRetriever pattern. Per ratification: threshold = 3.
--
-- Inputs:
--   p_query_embedding — vector to search against
--   p_message_id      — optional filter to a single message's corpus
--   p_k               — max results
--   p_merge_threshold — minimum hits under same parent to promote
--
-- Output: row per result with kind in ('parent','leaf'), content, score.
-- =====================================================================


CREATE OR REPLACE FUNCTION stewards.retrieve_with_merge(
    p_query_embedding vector,
    p_message_id      bigint DEFAULT NULL,
    p_k               int    DEFAULT 8,
    p_merge_threshold int    DEFAULT 3
) RETURNS TABLE (
    kind         text,
    parent_id    bigint,
    leaf_id      bigint,
    message_id   bigint,
    content      text,
    score        float,
    matched_leaves int
) LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_search_k int;
BEGIN
    -- Over-fetch leaves so we have enough to detect parent concentration.
    v_search_k := GREATEST(p_k * 4, 20);

    RETURN QUERY
    WITH
    leaf_hits AS (
        SELECT
            l.id            AS leaf_id,
            l.parent_id,
            l.message_id,
            l.content,
            l.context_prefix,
            1 - (l.embedding <=> p_query_embedding) AS similarity
          FROM stewards.messages_raw_overflow_leaves l
         WHERE l.embedding IS NOT NULL
           AND (p_message_id IS NULL OR l.message_id = p_message_id)
         ORDER BY l.embedding <=> p_query_embedding
         LIMIT v_search_k
    ),
    parent_concentration AS (
        SELECT parent_id, count(*) AS hits, max(similarity) AS best_score
          FROM leaf_hits
         GROUP BY parent_id
    ),
    merged_parents AS (
        SELECT
            'parent'::text         AS kind,
            pc.parent_id,
            NULL::bigint           AS leaf_id,
            p.message_id,
            p.content,
            pc.best_score::float   AS score,
            pc.hits::int           AS matched_leaves
          FROM parent_concentration pc
          JOIN stewards.messages_raw_overflow p
            ON p.id = pc.parent_id
         WHERE pc.hits >= p_merge_threshold
    ),
    unmerged_leaves AS (
        SELECT
            'leaf'::text                     AS kind,
            l.parent_id,
            l.leaf_id,
            l.message_id,
            COALESCE(l.context_prefix || E'\n\n' || l.content, l.content) AS content,
            l.similarity::float              AS score,
            1::int                           AS matched_leaves
          FROM leaf_hits l
         WHERE NOT EXISTS (
             SELECT 1 FROM merged_parents mp WHERE mp.parent_id = l.parent_id
         )
    )
    SELECT * FROM merged_parents
    UNION ALL
    SELECT * FROM unmerged_leaves
    ORDER BY score DESC
    LIMIT p_k;
END;
$FN$;

COMMENT ON FUNCTION stewards.retrieve_with_merge(vector, bigint, int, int) IS
'Batch L.1.1.7: auto-merging retrieval. Vector search top-(4k) leaves; if merge_threshold (default 3) hits land under the same parent, promote to the parent. Returns top-k mixed results (kind=parent|leaf) ordered by score. p_message_id filters to a single message''s corpus (for judge surface); NULL searches all overflow.';


-- ---------------------------------------------------------------------
-- Convenience: retrieve_with_merge_by_text — Go wrapper builds query
-- embedding externally; this SQL fn just takes the vector. But for
-- ad-hoc testing, expose a variant that takes a leaf_id (uses that
-- leaf's existing embedding as the query — for "find more like this").
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.retrieve_with_merge_like_leaf(
    p_seed_leaf_id    bigint,
    p_message_id      bigint DEFAULT NULL,
    p_k               int    DEFAULT 8,
    p_merge_threshold int    DEFAULT 3
) RETURNS TABLE (
    kind text,
    parent_id bigint,
    leaf_id bigint,
    message_id bigint,
    content text,
    score float,
    matched_leaves int
) LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_seed_embedding vector;
BEGIN
    SELECT embedding INTO v_seed_embedding
      FROM stewards.messages_raw_overflow_leaves
     WHERE id = p_seed_leaf_id;

    IF v_seed_embedding IS NULL THEN
        RAISE EXCEPTION 'retrieve_with_merge_like_leaf: leaf % has no embedding', p_seed_leaf_id;
    END IF;

    RETURN QUERY SELECT * FROM stewards.retrieve_with_merge(
        v_seed_embedding, p_message_id, p_k, p_merge_threshold);
END;
$FN$;


-- =====================================================================
-- End of l17-retrieve-with-merge.sql
-- =====================================================================
