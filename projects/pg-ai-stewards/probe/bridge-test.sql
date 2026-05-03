-- bridge-test.sql
--
-- Run this with: docker exec -i pg-ai-stewards-probe psql -U stewards -d stewards -f /dev/stdin < bridge-test.sql
-- Or paste blocks one at a time inside `docker exec -it pg-ai-stewards-probe psql -U stewards -d stewards`.
--
-- Goal: prove pgvector + Apache AGE can coexist in one DB on PG18, and
-- that the "bridge" pattern (compute similarity with pgvector, store
-- as AGE edge, mix in Cypher) works in a single transaction.

-- AGE setup must be re-applied per session.
LOAD 'age';
SET search_path TO ag_catalog, "$user", public;

-- ============================================================
-- Block 1 — sanity: both extensions present
-- ============================================================
SELECT extname, extversion
  FROM pg_extension
 WHERE extname IN ('vector', 'age')
 ORDER BY extname;

-- ============================================================
-- Block 2 — relational + vector table (study/scripture stand-ins)
-- ============================================================
CREATE TABLE IF NOT EXISTS docs (
    id           BIGSERIAL PRIMARY KEY,
    kind         TEXT NOT NULL,            -- 'study' | 'scripture' | 'talk'
    label        TEXT NOT NULL,
    embedding    vector(4)                 -- tiny vec for the probe
);

TRUNCATE docs RESTART IDENTITY;

INSERT INTO docs (kind, label, embedding) VALUES
    ('scripture', '1 Ne 3:7',     '[1.0, 0.0, 0.0, 0.1]'),
    ('scripture', '2 Ne 32:3',    '[0.9, 0.1, 0.0, 0.0]'),
    ('scripture', 'Moroni 7:45',  '[0.1, 0.9, 0.0, 0.0]'),
    ('talk',      'oaks-charity', '[0.0, 1.0, 0.0, 0.0]'),
    ('study',     'on-charity',   '[0.05, 0.95, 0.0, 0.0]');

CREATE INDEX IF NOT EXISTS idx_docs_emb
    ON docs USING hnsw (embedding vector_cosine_ops);

-- ============================================================
-- Block 3 — pgvector: nearest neighbors
-- "Find the docs most semantically near the 'on-charity' study."
-- ============================================================
WITH q AS (SELECT embedding FROM docs WHERE label = 'on-charity')
SELECT d.id, d.kind, d.label,
       round((1 - (d.embedding <=> q.embedding))::numeric, 4) AS sim
  FROM docs d, q
 WHERE d.label <> 'on-charity'
 ORDER BY d.embedding <=> q.embedding
 LIMIT 3;

-- Expected top hit: 'oaks-charity' and 'Moroni 7:45' (vectors are close).

-- ============================================================
-- Block 4 — AGE: create graph nodes for each doc
-- We use the doc id as a stable property (the URI in real life).
-- ============================================================
SELECT * FROM cypher('stewards_graph', $$
    UNWIND [
        {id: 1, kind: 'scripture', ref: 'lds://scripture/bofm/1-ne/3.7'},
        {id: 2, kind: 'scripture', ref: 'lds://scripture/bofm/2-ne/32.3'},
        {id: 3, kind: 'scripture', ref: 'lds://scripture/bofm/moroni/7.45'},
        {id: 4, kind: 'talk',      ref: 'lds://talk/oaks-charity'},
        {id: 5, kind: 'study',     ref: 'lds://study/on-charity'}
    ] AS d
    MERGE (n:Doc {id: d.id})
    SET n.kind = d.kind, n.ref = d.ref
    RETURN count(n)
$$) AS (n agtype);

-- ============================================================
-- Block 5 — bridge: pgvector similarity → AGE edge
-- For each candidate near 'on-charity', create a SIMILAR_TO edge.
-- Using a DO block + format() to inject scalar values into Cypher,
-- which is the documented AGE pattern (Cypher can't bind PG params).
-- ============================================================
DO $bridge$
DECLARE
    src_id   bigint := (SELECT id FROM docs WHERE label = 'on-charity');
    rec      record;
BEGIN
    FOR rec IN
        WITH q AS (SELECT embedding FROM docs WHERE label = 'on-charity')
        SELECT d.id AS dst_id,
               round((1 - (d.embedding <=> q.embedding))::numeric, 4) AS sim
          FROM docs d, q
         WHERE d.label <> 'on-charity'
         ORDER BY d.embedding <=> q.embedding
         LIMIT 3
    LOOP
        EXECUTE format($cy$
            SELECT * FROM cypher('stewards_graph', $$
                MATCH (a:Doc {id: %s}), (b:Doc {id: %s})
                MERGE (a)-[r:SIMILAR_TO {method: 'pgvector_cosine'}]->(b)
                SET r.score = %s
                RETURN r
            $$) AS (r agtype)
        $cy$, src_id, rec.dst_id, rec.sim);
    END LOOP;
END
$bridge$;

-- ============================================================
-- Block 6 — Cypher reads the bridge edges back
-- ============================================================
SELECT * FROM cypher('stewards_graph', $$
    MATCH (s:Doc {kind: 'study'})-[r:SIMILAR_TO]->(t:Doc)
    RETURN s.ref AS study, t.kind AS target_kind, t.ref AS target, r.score AS score
    ORDER BY r.score DESC
$$) AS (study agtype, target_kind agtype, target agtype, score agtype);

-- Expected: three rows, the on-charity study pointing at the two
-- charity-shaped docs and one weaker neighbor, ordered by score desc.

-- ============================================================
-- Block 7 — combined: pgvector + cypher in one statement
-- "Find scriptures semantically near the study, but only ones that
--  the graph already marks as 'core' for some other study."
-- (We seed one such marker first.)
-- ============================================================
SELECT * FROM cypher('stewards_graph', $$
    MERGE (other:Doc {id: 999})
    SET other.kind = 'study', other.ref = 'lds://study/older-study'
    WITH other
    MATCH (s:Doc {id: 3})  // Moroni 7:45
    MERGE (other)-[:CITES_AS_CORE]->(s)
    RETURN s
$$) AS (s agtype);

WITH near AS (
    SELECT d.id, d.label,
           round((1 - (d.embedding <=> (SELECT embedding FROM docs WHERE label='on-charity')))::numeric, 4) AS sim
      FROM docs d
     WHERE d.kind = 'scripture'
     ORDER BY d.embedding <=> (SELECT embedding FROM docs WHERE label='on-charity')
     LIMIT 3
),
core AS (
    -- Project a scalar property out of cypher() so we can cast cleanly.
    -- Casting a vertex agtype to jsonb fails because of the trailing
    -- "::vertex" suffix; project n.id instead and cast text->bigint.
    SELECT (nid::text)::bigint AS id
      FROM cypher('stewards_graph', $$
          MATCH (s:Doc)-[:CITES_AS_CORE]->(n:Doc) RETURN n.id
      $$) AS (nid agtype)
)
SELECT n.id, n.label, n.sim,
       CASE WHEN c.id IS NOT NULL THEN 'yes' ELSE 'no' END AS already_core_for_some_study
  FROM near n LEFT JOIN core c ON c.id = n.id
 ORDER BY n.sim DESC;

-- This is the "stop paying the multi-database tax" claim: a single
-- statement that uses pgvector AND AGE in the same transaction,
-- planned by one planner, returning one resultset.
