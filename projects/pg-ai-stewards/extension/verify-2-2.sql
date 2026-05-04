-- Phase 2.2 verification — gospel-engine resolver.
--
-- Run after a fresh boot + import-studies.ps1, with enough time for
-- the bgworker to drain the resolve_ref queue (a corpus-wide refresh
-- enqueues ~1300 items; LM Studio + gospel-engine round-trips finish
-- in ~4 minutes on the dev box).
--
-- Run with:
--   Get-Content verify-2-2.sql | docker exec -i pg-ai-stewards-dev psql -U stewards -d stewards
--
-- Tests:
--   1. parse_reference handles canonical anchor shapes.
--   2. normalize_book maps LDS abbreviations to gospel-engine forms.
--   3. enqueue_resolve is idempotent and errors are sticky.
--   4. refresh_study_refs is idempotent on re-call.
--   5. invalidate_ref + refresh round-trips a single ref.
--   6. study_citations_resolved joins citations to verse text.
--   7. 404s from gospel-engine are cached as error rows (no retry).

\set ON_ERROR_STOP 1

\echo === Test 1: parse_reference shapes ===
SELECT i, array_agg(r ORDER BY r) AS refs
  FROM (VALUES
        ('Mosiah 18:8'),
        ('Mosiah 18:8-9'),
        ('Mosiah 18:8' || E'\u2013' || '9'),
        ('D&C 88:67-68'),
        ('Mosiah 12:28, 32'),
        ('1 Nephi 3:7'),
        ('JS-H 1:17'),
        ('D&C 76'),               -- chapter-only -> empty
        ('Maxwell 1991')          -- not a ref -> empty
  ) AS t(i)
  CROSS JOIN LATERAL stewards.parse_reference(i) AS r
 GROUP BY i ORDER BY i;

\echo === Test 2: normalize_book maps LDS abbrevs ===
SELECT i, array_agg(r ORDER BY r) AS refs
  FROM (VALUES
        ('Psalm 23:1'),       -- singular -> Psalms
        ('3 Ne. 11:1'),       -- 3 Ne -> 3 Nephi
        ('Rom. 8:28'),        -- Rom -> Romans
        ('Heb. 11:1'),        -- Heb -> Hebrews
        ('Jas. 1:5'),         -- Jas -> James
        ('1 Cor. 13:1'),      -- 1 Cor -> 1 Corinthians
        ('Gal. 6:2')          -- Gal -> Galatians
  ) AS t(i)
  CROSS JOIN LATERAL stewards.parse_reference(i) AS r
 GROUP BY i ORDER BY i;

\echo === Test 3: enqueue_resolve idempotency + sticky errors ===
SELECT 'first call enqueues'  AS label,
       stewards.enqueue_resolve('TestBook 99:99') IS NOT NULL AS result;
SELECT 'second call skips'    AS label,
       stewards.enqueue_resolve('TestBook 99:99') IS NULL AS result;
DELETE FROM stewards.work_queue WHERE payload->>'ref' = 'TestBook 99:99';
DELETE FROM stewards.resolved_refs WHERE ref = 'TestBook 99:99';

\echo === Test 4: refresh_study_refs is idempotent on second call ===
SELECT stewards.refresh_study_refs('art-of-delegation') AS rerun_enqueued;

\echo === Test 5: invalidate + refresh round-trips a single ref ===
SELECT stewards.invalidate_ref('Mosiah 18:8') AS invalidated;
SELECT stewards.refresh_study_refs('art-of-delegation') AS new_after_invalidate;
SELECT pg_sleep(3);
SELECT ref, length(content->>'text') AS chars, attempt_count
  FROM stewards.resolved_refs WHERE ref = 'Mosiah 18:8';

\echo === Test 6: study_citations_resolved end-to-end ===
SELECT cited_uri,
       anchor_text,
       jsonb_array_length(resolved_verses) AS verse_count,
       (resolved_verses->0->>'ref') AS first_ref,
       left(resolved_verses->0->'content'->>'text', 60) AS first_text_preview
  FROM stewards.study_citations_resolved('art-of-delegation')
 WHERE cited_kind = 'scripture'
 ORDER BY citation_count DESC
 LIMIT 5;

\echo === Test 7: corpus state (404s cached, no retries) ===
SELECT count(*) AS total_cached,
       count(*) FILTER (WHERE error IS NULL) AS success,
       count(*) FILTER (WHERE error IS NOT NULL) AS missing,
       round(100.0 * count(*) FILTER (WHERE error IS NULL) / count(*), 1) AS success_pct
  FROM stewards.resolved_refs;

SELECT max(attempt_count) AS max_attempts_on_any_ref,
       count(*) FILTER (WHERE attempt_count > 1) AS refs_retried
  FROM stewards.resolved_refs;

\echo === Done. ===
