-- Phase 2.1 verification — studies + AGE citations.
-- Run with: cat verify-2-1.sql | docker exec -i pg-ai-stewards-dev psql -U stewards -d stewards
--
-- Tests:
--   1. Row count + edge count baseline
--   2. parse_gospel_links handles the link shapes we expect
--   3. import_study survives apostrophes, em-dashes, parentheses
--      (the inverse hypothesis — these were the actual failure mode)
--   4. Re-importing a study syncs edges (no duplicates, removed
--      links disappear)
--   5. study_citations() round-trips the graph back to relational
--   6. Cross-study graph queries: which studies cite Mosiah 18?

\set ON_ERROR_STOP 1
LOAD 'age';
SET search_path = ag_catalog, public;

\echo === Test 1: corpus loaded ===
SELECT count(*) AS studies FROM stewards.studies;
SELECT * FROM cypher('stewards_graph', $$
    MATCH (s:Study) RETURN count(s) AS study_vertices
$$) AS (study_vertices agtype);
SELECT * FROM cypher('stewards_graph', $$
    MATCH (s:Scripture) RETURN count(s) AS scripture_vertices
$$) AS (scripture_vertices agtype);
SELECT * FROM cypher('stewards_graph', $$
    MATCH ()-[r:CITES]->() RETURN count(r) AS cites_edges
$$) AS (cites_edges agtype);

\echo === Test 2: parse_gospel_links handles the common shapes ===
WITH sample AS (
    SELECT $$Some text [Mosiah 18:8-9](../gospel-library/eng/scriptures/bofm/mosiah/18.md)
and another [Moroni 7:47](../gospel-library/eng/scriptures/bofm/moro/7.md#47)
and a talk [Maxwell 1991](../gospel-library/eng/general-conference/1991/04/maxwell.md)
and a deeply-nested [D&C 88](../../../gospel-library/eng/scriptures/dc-testament/dc/88.md)
and a non-link reference to gospel-library/foo without brackets$$ AS body
)
SELECT * FROM stewards.parse_gospel_links((SELECT body FROM sample));

\echo === Test 3: apostrophes, em-dashes, parentheses survive ===
SELECT stewards.import_study(
    'verify-quote-test',
    'verify/quote-test.md',
    $$The Serpent's Craft — Title (with em-dash & parens)$$,
    $$Body with [Maxwell's "Notwithstanding My Weakness"](../gospel-library/eng/general-conference/1981/10/maxwell.md)
and [Mosiah 18:8-9](../gospel-library/eng/scriptures/bofm/mosiah/18.md)
and a stray apostrophe: don't break.$$,
    '{}'::jsonb
) AS new_id;
SELECT * FROM stewards.study_citations('verify-quote-test');

\echo === Test 4: re-import is idempotent (edge count unchanged) ===
SELECT count(*) AS edges_before FROM (
    SELECT * FROM cypher('stewards_graph', $$
        MATCH (s:Study {slug: 'verify-quote-test'})-[r:CITES]->() RETURN r
    $$) AS (r agtype)
) sub;
SELECT stewards.import_study(
    'verify-quote-test',
    'verify/quote-test.md',
    $$The Serpent's Craft — Title (with em-dash & parens)$$,
    $$Body with [Maxwell's "Notwithstanding My Weakness"](../gospel-library/eng/general-conference/1981/10/maxwell.md)
and [Mosiah 18:8-9](../gospel-library/eng/scriptures/bofm/mosiah/18.md)
and a stray apostrophe: don't break.$$,
    '{}'::jsonb
) AS reimport_id;
SELECT count(*) AS edges_after FROM (
    SELECT * FROM cypher('stewards_graph', $$
        MATCH (s:Study {slug: 'verify-quote-test'})-[r:CITES]->() RETURN r
    $$) AS (r agtype)
) sub;

\echo === Test 5: removing a link from the body removes the edge ===
SELECT stewards.import_study(
    'verify-quote-test',
    'verify/quote-test.md',
    $$The Serpent's Craft — Title$$,
    $$Body with only one link now: [Mosiah 18:8-9](../gospel-library/eng/scriptures/bofm/mosiah/18.md)$$,
    '{}'::jsonb
) AS shrunk_id;
SELECT * FROM stewards.study_citations('verify-quote-test');

\echo === Test 6: cross-study graph query — who cites Mosiah 18? ===
SELECT * FROM cypher('stewards_graph', $$
    MATCH (s:Study)-[:CITES]->(t:Scripture {uri: 'eng/scriptures/bofm/mosiah/18.md'})
    RETURN s.slug, s.title
    ORDER BY s.slug
$$) AS (slug agtype, title agtype);

\echo === Test 7: cleanup verify-quote-test ===
DELETE FROM stewards.studies WHERE slug = 'verify-quote-test';
SELECT * FROM cypher('stewards_graph', $$
    MATCH (s:Study {slug: 'verify-quote-test'}) DETACH DELETE s
$$) AS (v agtype);

\echo === Done. ===
