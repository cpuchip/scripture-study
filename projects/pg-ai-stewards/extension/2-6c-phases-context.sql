-- Phase 2.6c — Phase splitter support + context_for() graph walk.
--
-- Live-DB migration. Same pattern as 2-6a / 2-6b.
--
-- Design note on context_for: AGE's variable-length path syntax has
-- enough rough edges (length(path), direction extraction, UNWIND of
-- relationship lists) that the cleanest correct implementation is
-- iterative in PL/pgSQL: do 1-hop Cypher per depth level, accumulate
-- results, dedupe in SQL. AGE quirks list now at 5: variable-length
-- paths are awkward enough we just don't use them.

-- BEGIN;  -- (folded into extension_sql_file! v0.2.0; CREATE EXTENSION already wraps in tx)
-- ============================================================
-- Function: link_phase_to_doc(phase_slug, parent_doc_slug)
--
-- MERGE :HAS_PHASE edge from :Study (the phase-doc) to :Study
-- (the phase row). Both are :Study label because import_study
-- writes everything as :Study; the discriminator is the row's
-- `kind` column, not the AGE label.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.link_phase_to_doc(
    p_phase_slug      text,
    p_parent_doc_slug text
) RETURNS void
LANGUAGE plpgsql AS $func$
BEGIN
    PERFORM stewards.ensure_studies_graph();

    EXECUTE
        $cy$
        SELECT * FROM cypher('stewards_graph', $$
            MATCH (parent:Study {slug: $parent}), (phase:Study {slug: $phase})
            MERGE (parent)-[r:HAS_PHASE]->(phase)
            SET r.provenance = 'declared',
                r.confidence = 1.0,
                r.source = 'phase_split'
            RETURN r
        $$, $1) AS (v agtype)
        $cy$
    USING (jsonb_build_object(
        'parent', p_parent_doc_slug,
        'phase',  p_phase_slug
    )::text)::ag_catalog.agtype;
END;
$func$;

-- ============================================================
-- Function: context_for_hop(slug)  — internal helper
--
-- Returns the 1-hop neighborhood of a vertex (matched by either
-- `slug` or `id` so it works for :Study/:Todo/:Phase as well as
-- :Workstream which uses `id`).
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.context_for_hop(
    p_seed_slug text
) RETURNS TABLE (
    direction     text,
    edge_type     text,
    neighbor      text,
    neighbor_kind text,
    provenance    text,
    confidence    float
)
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_params ag_catalog.agtype := (jsonb_build_object('slug', p_seed_slug)::text)::ag_catalog.agtype;
BEGIN
    PERFORM stewards.ensure_studies_graph();

    RETURN QUERY
    SELECT 'out'::text,
            etype::text,
               coalesce(n_slug::text, n_id::text, n_uri::text),
            coalesce(n_kind::text, 'Study'),
            coalesce(prov::text, 'unknown'),
            coalesce(conf::text::float, 0.0)
      FROM cypher('stewards_graph', $$
            MATCH (s)-[r]->(n)
            WHERE s.slug = $slug OR s.id = $slug
              RETURN type(r), n.kind, n.slug, n.id, n.uri,
                   r.provenance, r.confidence
                $$, v_params)
        AS h(etype agtype, n_kind agtype,
               n_slug agtype, n_id agtype, n_uri agtype,
             prov agtype, conf agtype)
    UNION ALL
    SELECT 'in'::text,
            etype::text,
               coalesce(n_slug::text, n_id::text, n_uri::text),
            coalesce(n_kind::text, 'Study'),
            coalesce(prov::text, 'unknown'),
            coalesce(conf::text::float, 0.0)
      FROM cypher('stewards_graph', $$
            MATCH (n)-[r]->(s)
            WHERE s.slug = $slug OR s.id = $slug
              RETURN type(r), n.kind, n.slug, n.id, n.uri,
                   r.provenance, r.confidence
                $$, v_params)
        AS h(etype agtype, n_kind agtype,
               n_slug agtype, n_id agtype, n_uri agtype,
             prov agtype, conf agtype);
END;
$func$;

-- ============================================================
-- Function: context_for(slug, depth)
--
-- Walks the graph iteratively up to `depth` hops (clamped 1..4).
-- Returns one row per (hop, direction, edge_type, neighbor) found,
-- deduplicated by neighbor across hops (closest hop wins).
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.context_for(
    p_slug  text,
    p_depth int DEFAULT 2
) RETURNS TABLE (
    hop           int,
    direction     text,
    edge_type     text,
    neighbor      text,
    neighbor_kind text,
    provenance    text,
    confidence    float
)
LANGUAGE plpgsql AS $func$
DECLARE
    v_depth int := greatest(1, least(p_depth, 4));
    v_hop   int := 1;
    v_added int;
BEGIN
    CREATE TEMP TABLE IF NOT EXISTS _ctx_results (
        hop           int,
        direction     text,
        edge_type     text,
        neighbor      text,
        neighbor_kind text,
        provenance    text,
        confidence    float
    ) ON COMMIT DROP;
    DELETE FROM _ctx_results;

    CREATE TEMP TABLE IF NOT EXISTS _ctx_frontier (
        slug text PRIMARY KEY,
        hop  int
    ) ON COMMIT DROP;
    DELETE FROM _ctx_frontier;
    INSERT INTO _ctx_frontier(slug, hop) VALUES (p_slug, 0);

    CREATE TEMP TABLE IF NOT EXISTS _ctx_seen (
        slug text PRIMARY KEY
    ) ON COMMIT DROP;
    DELETE FROM _ctx_seen;
    INSERT INTO _ctx_seen(slug) VALUES (p_slug);

    WHILE v_hop <= v_depth LOOP
        WITH frontier_slugs AS (
            SELECT f.slug FROM _ctx_frontier f WHERE f.hop = v_hop - 1
        ),
        expanded AS (
            SELECT v_hop AS hop, h.direction, h.edge_type,
                   h.neighbor, h.neighbor_kind, h.provenance, h.confidence
              FROM frontier_slugs fs
              CROSS JOIN LATERAL stewards.context_for_hop(fs.slug) h
             WHERE h.neighbor IS NOT NULL
               AND NOT EXISTS (SELECT 1 FROM _ctx_seen s WHERE s.slug = h.neighbor)
        ),
        ins_results AS (
            INSERT INTO _ctx_results(hop, direction, edge_type, neighbor,
                                     neighbor_kind, provenance, confidence)
            SELECT e.hop, e.direction, e.edge_type, e.neighbor,
                   e.neighbor_kind, e.provenance, e.confidence
              FROM expanded e
            RETURNING _ctx_results.neighbor AS new_neighbor
        ),
        ins_seen AS (
            INSERT INTO _ctx_seen(slug)
            SELECT DISTINCT i.new_neighbor FROM ins_results i
            ON CONFLICT DO NOTHING
            RETURNING slug
        )
        INSERT INTO _ctx_frontier(slug, hop)
        SELECT s.slug, v_hop FROM ins_seen s
        ON CONFLICT (slug) DO NOTHING;

        GET DIAGNOSTICS v_added = ROW_COUNT;
        EXIT WHEN v_added = 0;
        v_hop := v_hop + 1;
    END LOOP;

    RETURN QUERY
    SELECT r.hop, r.direction, r.edge_type, r.neighbor,
           r.neighbor_kind, r.provenance, r.confidence
      FROM _ctx_results r
     ORDER BY r.hop, r.direction DESC, r.edge_type, r.neighbor;
END;
$func$;

-- COMMIT; -- (folded into extension_sql_file! v0.2.0; CREATE EXTENSION already wraps in tx)