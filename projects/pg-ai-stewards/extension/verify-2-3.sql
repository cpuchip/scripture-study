-- Phase 2.3 verification — pgvector cosine -> AGE :SIMILAR_TO bridge.
--
-- Run after a fresh boot + import-studies.ps1, with enough time for
-- the embed bgworker to embed all 69 studies (~90s on the dev box).
--
-- Run with:
--   Get-Content verify-2-3.sql | docker exec -i pg-ai-stewards-dev psql -U stewards -d stewards
--
-- Tests:
--   1. refresh_all_study_similarity writes top_k * embedded_count edges.
--   2. study_similar returns top-K with descending scores.
--   3. Direction labeling matches outgoing/incoming/mutual semantics.
--   4. Re-running refresh is idempotent (same edge count, no duplicates).
--   5. Tightening min_score reduces edge count monotonically.
--   6. Inverse hypothesis: lowering an embedding's similarity drops
--      edges; restoring it brings them back.

\set ON_ERROR_STOP 1
SET search_path TO ag_catalog, "$user", public;

\echo === Test 1: corpus refresh writes K=5 edges per embedded study ===
SELECT count(embedding) AS embedded FROM stewards.studies;
SELECT stewards.refresh_all_study_similarity() AS edges_after_refresh;

-- Direct AGE count for cross-check.
SELECT (ag_catalog.agtype_to_int8(c)) AS edge_count_via_cypher
  FROM cypher('stewards_graph', $$
        MATCH ()-[r:SIMILAR_TO {method: 'pgvector_cosine'}]->()
        RETURN count(r)
       $$) AS (c agtype);

\echo === Test 2: top-K reads, scores descending ===
SELECT slug, round(score::numeric, 3) AS score, direction
  FROM stewards.study_similar('art-of-delegation', 5);

\echo === Test 3: direction labels include mutual + outgoing + incoming ===
-- Across the corpus we expect a mix of all three. If we ONLY ever
-- saw 'mutual' something is wrong with the asymmetric K cutoff
-- handling.
WITH all_dirs AS (
    SELECT s.slug AS src,
           (sim).direction
      FROM stewards.studies s,
           LATERAL stewards.study_similar(s.slug, 10) sim
     WHERE s.embedding IS NOT NULL
)
SELECT direction, count(*) AS n
  FROM all_dirs
 GROUP BY direction
 ORDER BY direction;

\echo === Test 4: refresh is idempotent (no edge multiplication) ===
SELECT stewards.refresh_all_study_similarity() AS edges_after_second_refresh;
SELECT (ag_catalog.agtype_to_int8(c)) AS edge_count_unchanged
  FROM cypher('stewards_graph', $$
        MATCH ()-[r:SIMILAR_TO {method: 'pgvector_cosine'}]->()
        RETURN count(r)
       $$) AS (c agtype);

\echo === Test 5: tightening min_score reduces edges ===
SELECT 'min=0.50' AS threshold,
       stewards.refresh_all_study_similarity(5, 0.50) AS edges;
SELECT 'min=0.80' AS threshold,
       stewards.refresh_all_study_similarity(5, 0.80) AS edges;
SELECT 'min=0.85' AS threshold,
       stewards.refresh_all_study_similarity(5, 0.85) AS edges;
-- Restore default for downstream tests.
SELECT 'restore' AS threshold,
       stewards.refresh_all_study_similarity() AS edges;

\echo === Test 6: inverse hypothesis ===
-- Two-step inverse: (1) refreshing only A after nulling A's
-- embedding drops A's OUTGOING edges to 0, but A's neighbors still
-- have outgoing edges TO A (which read as 'incoming' from A's view).
-- (2) Refreshing the WHOLE corpus then drops those too. Restoring
-- the embedding + corpus refresh brings the original neighbors back.
\echo --- baseline neighbors of art-of-delegation ---
SELECT slug, round(score::numeric, 3) AS score, direction
  FROM stewards.study_similar('art-of-delegation', 5);

-- Save and zero-out the embedding.
CREATE TEMP TABLE _bak AS
  SELECT embedding FROM stewards.studies WHERE slug = 'art-of-delegation';
UPDATE stewards.studies SET embedding = NULL WHERE slug = 'art-of-delegation';

-- Step 1: refresh A only \u2014 A's outgoing edges go to 0, neighbors
-- haven't refreshed yet so 'incoming' edges remain.
SELECT stewards.refresh_study_similarity('art-of-delegation') AS step1_a_outgoing_after_null;
\echo --- after refresh A only (incoming edges remain until neighbors refresh) ---
SELECT slug, round(score::numeric, 3) AS score, direction
  FROM stewards.study_similar('art-of-delegation', 5);

-- Step 2: corpus refresh \u2014 every other study re-picks its top-K
-- without A as a candidate, so all incoming-to-A edges disappear.
SELECT stewards.refresh_all_study_similarity() AS step2_corpus_after_null;
\echo --- after full corpus refresh (should be empty) ---
SELECT slug, round(score::numeric, 3) AS score, direction
  FROM stewards.study_similar('art-of-delegation', 5);

-- Restore embedding + corpus refresh.
UPDATE stewards.studies
   SET embedding = (SELECT embedding FROM _bak)
 WHERE slug = 'art-of-delegation';
SELECT stewards.refresh_all_study_similarity() AS restored_corpus;

\echo --- restored neighbors (should match baseline) ---
SELECT slug, round(score::numeric, 3) AS score, direction
  FROM stewards.study_similar('art-of-delegation', 5);

DROP TABLE _bak;

\echo === Done. ===
