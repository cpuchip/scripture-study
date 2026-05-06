-- Phase 2.6a — Workstream vertices + frontmatter :DECLARED edges
--
-- Live-DB migration. Run with:
--   docker exec -i pg-ai-stewards-dev psql -U stewards -d stewards \
--     -f /dev/stdin < projects/pg-ai-stewards/extension/2-6a-workstreams.sql
-- (or copy the file in via -v and run it).
--
-- This is intentionally a plain SQL script rather than an
-- extension_sql! block in lib.rs because we need to apply it to the
-- live DB without rebuilding the docker image and dropping the 359-doc
-- corpus. When the extension is next rebuilt, fold this into lib.rs
-- as a new extension_sql! block with `requires = ["create_studies"]`.
--
-- Idempotent: every CREATE uses IF NOT EXISTS / OR REPLACE.

-- BEGIN;  -- (folded into extension_sql_file! v0.2.0; CREATE EXTENSION already wraps in tx)
-- ============================================================
-- Table: stewards.workstreams
-- ============================================================
CREATE TABLE IF NOT EXISTS stewards.workstreams (
    id          text PRIMARY KEY,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    status      text NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'paused', 'retired')),
    -- Free-form bag for things that haven't earned columns yet.
    frontmatter jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS workstreams_status_idx
    ON stewards.workstreams (status);

-- ============================================================
-- Function: import_workstream(id, name, description, status)
--
-- Upserts the workstreams row AND merges the :Workstream vertex in
-- the AGE graph. Returns the id.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.import_workstream(
    p_id          text,
    p_name        text,
    p_description text DEFAULT '',
    p_status      text DEFAULT 'active'
) RETURNS text
LANGUAGE plpgsql AS $func$
BEGIN
    PERFORM stewards.ensure_studies_graph();

    INSERT INTO stewards.workstreams (id, name, description, status)
    VALUES (p_id, p_name, p_description, p_status)
    ON CONFLICT (id) DO UPDATE
        SET name        = EXCLUDED.name,
            description = EXCLUDED.description,
            status      = EXCLUDED.status,
            updated_at  = now();

    -- MERGE the :Workstream vertex.
    EXECUTE
        $cy$
        SELECT * FROM cypher('stewards_graph', $$
            MERGE (w:Workstream {id: $id})
            SET w.name = $name, w.status = $status
            RETURN w
        $$, $1) AS (v agtype)
        $cy$
    USING (jsonb_build_object(
        'id',     p_id,
        'name',   p_name,
        'status', p_status
    )::text)::ag_catalog.agtype;

    RETURN p_id;
END;
$func$;

-- ============================================================
-- Function: link_declared_edges(slug, frontmatter)
--
-- Reads workstream/feeds/supersedes/implements from the frontmatter
-- and creates :DECLARED-provenance edges:
--   :Workstream -[:HAS_PROPOSAL {provenance:'declared'}]-> :Study
--   :Study      -[:FEEDS        {provenance:'declared'}]-> :Study
--   :Study      -[:SUPERSEDES   {provenance:'declared'}]-> :Study
--   :Study      -[:IMPLEMENTS   {provenance:'declared'}]-> :Study
--
-- All edges carry: provenance='declared', confidence=1.0, source='frontmatter:<key>'.
--
-- Drops existing :DECLARED-provenance edges from this slug first so
-- re-imports stay in sync. Inferred/linked edges (later phases) are
-- not touched.
--
-- Frontmatter shapes accepted:
--   workstream: WS5                       -- string
--   feeds: [other-slug]                   -- array
--   feeds: other-slug                     -- string (single)
--   supersedes: [a, b]
--   implements: [a]
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.link_declared_edges(
    p_slug        text,
    p_frontmatter jsonb
) RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_count    int := 0;
    v_ws_id    text;
    v_target   text;
    v_targets  text[];
    v_relation text;
BEGIN
    PERFORM stewards.ensure_studies_graph();

    -- 1. Drop existing :DECLARED-provenance edges FROM this slug.
    --    We match on provenance so we don't clobber linked/inferred
    --    edges that other passes may have written.
    EXECUTE
        $cy$
        SELECT * FROM cypher('stewards_graph', $$
            MATCH (s:Study {slug: $slug})-[r]->()
            WHERE r.provenance = 'declared' AND type(r) <> 'CITES'
            DELETE r
        $$, $1) AS (v agtype)
        $cy$
    USING (jsonb_build_object('slug', p_slug)::text)::ag_catalog.agtype;

    -- 2. Drop incoming :HAS_PROPOSAL edges (workstream membership).
    EXECUTE
        $cy$
        SELECT * FROM cypher('stewards_graph', $$
            MATCH ()-[r:HAS_PROPOSAL]->(s:Study {slug: $slug})
            WHERE r.provenance = 'declared'
            DELETE r
        $$, $1) AS (v agtype)
        $cy$
    USING (jsonb_build_object('slug', p_slug)::text)::ag_catalog.agtype;

    -- 3. Workstream membership: :Workstream -[:HAS_PROPOSAL]-> :Study
    v_ws_id := p_frontmatter->>'workstream';
    IF v_ws_id IS NOT NULL AND v_ws_id <> '' THEN
        -- MERGE the :Workstream vertex if it doesn't exist (e.g. doc
        -- references a workstream not yet seeded). AGE Cypher does
        -- NOT support ON CREATE SET — properties beyond the MERGE
        -- key get filled in lazily by import_workstream().
        EXECUTE
            $cy$
            SELECT * FROM cypher('stewards_graph', $$
                MERGE (w:Workstream {id: $ws_id})
                WITH w
                MATCH (s:Study {slug: $slug})
                MERGE (w)-[r:HAS_PROPOSAL]->(s)
                SET r.provenance = 'declared',
                    r.confidence = 1.0,
                    r.source = 'frontmatter:workstream'
                RETURN r
            $$, $1) AS (v agtype)
            $cy$
        USING (jsonb_build_object(
            'ws_id', v_ws_id,
            'slug',  p_slug
        )::text)::ag_catalog.agtype;
        v_count := v_count + 1;
    END IF;

    -- 4. Typed semantic edges: feeds / supersedes / implements
    FOREACH v_relation IN ARRAY ARRAY['feeds', 'supersedes', 'implements']
    LOOP
        v_targets := NULL;

        -- Accept array shape
        IF jsonb_typeof(p_frontmatter->v_relation) = 'array' THEN
            SELECT array_agg(value::text) INTO v_targets
            FROM jsonb_array_elements_text(p_frontmatter->v_relation) AS value;
        -- Accept string shape (single target)
        ELSIF jsonb_typeof(p_frontmatter->v_relation) = 'string' THEN
            v_targets := ARRAY[p_frontmatter->>v_relation];
        END IF;

        IF v_targets IS NULL THEN CONTINUE; END IF;

        FOREACH v_target IN ARRAY v_targets
        LOOP
            IF v_target IS NULL OR v_target = '' THEN CONTINUE; END IF;

            -- Build the per-relation Cypher (edge type varies).
            EXECUTE format(
                $cy$
                SELECT * FROM cypher('stewards_graph', $$
                    MATCH (s:Study {slug: $slug})
                    MERGE (t:Study {slug: $target})
                    MERGE (s)-[r:%s]->(t)
                    SET r.provenance = 'declared',
                        r.confidence = 1.0,
                        r.source = 'frontmatter:%s'
                    RETURN r
                $$, $1) AS (v agtype)
                $cy$,
                upper(v_relation),
                v_relation
            )
            USING (jsonb_build_object(
                'slug',   p_slug,
                'target', v_target
            )::text)::ag_catalog.agtype;
            v_count := v_count + 1;
        END LOOP;
    END LOOP;

    RETURN v_count;
END;
$func$;

-- ============================================================
-- Read function: workstream_proposals(ws_id) — list proposals
-- declared as belonging to a workstream.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.workstream_proposals(p_ws_id text)
RETURNS TABLE (slug text, kind text, title text, file_path text)
LANGUAGE plpgsql STABLE AS $func$
BEGIN
    PERFORM stewards.ensure_studies_graph();
    RETURN QUERY EXECUTE
        $cy$
        SELECT
            ag_catalog.agtype_to_text(s_slug)::text,
            ag_catalog.agtype_to_text(s_kind)::text,
            ag_catalog.agtype_to_text(s_title)::text,
            ag_catalog.agtype_to_text(s_file)::text
        FROM cypher('stewards_graph', $$
            MATCH (w:Workstream {id: $ws_id})-[r:HAS_PROPOSAL]->(s:Study)
            RETURN s.slug, s.kind, s.title, s.file_path
            ORDER BY s.slug
        $$, $1) AS (s_slug agtype, s_kind agtype, s_title agtype, s_file agtype)
        $cy$
    USING (jsonb_build_object('ws_id', p_ws_id)::text)::ag_catalog.agtype;
END;
$func$;

-- ============================================================
-- Read function: declared_edges(slug) — list outbound declared edges.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.declared_edges(p_slug text)
RETURNS TABLE (
    from_slug    text,
    edge_type    text,
    to_slug      text,
    provenance   text,
    confidence   float,
    source       text
)
LANGUAGE plpgsql STABLE AS $func$
BEGIN
    PERFORM stewards.ensure_studies_graph();
    RETURN QUERY EXECUTE
        $cy$
        SELECT
            ag_catalog.agtype_to_text(s_slug)::text,
            ag_catalog.agtype_to_text(r_type)::text,
            ag_catalog.agtype_to_text(t_slug)::text,
            ag_catalog.agtype_to_text(r_prov)::text,
            ag_catalog.agtype_to_float8(r_conf)::float,
            ag_catalog.agtype_to_text(r_src)::text
        FROM cypher('stewards_graph', $$
            MATCH (s:Study {slug: $slug})-[r]->(t)
            WHERE r.provenance IS NOT NULL AND type(r) <> 'CITES'
            RETURN s.slug, type(r), t.slug, r.provenance, r.confidence, r.source
            ORDER BY type(r), t.slug
        $$, $1) AS (s_slug agtype, r_type agtype, t_slug agtype,
                    r_prov agtype, r_conf agtype, r_src agtype)
        $cy$
    USING (jsonb_build_object('slug', p_slug)::text)::ag_catalog.agtype;
END;
$func$;

-- COMMIT; -- (folded into extension_sql_file! v0.2.0; CREATE EXTENSION already wraps in tx)
-- ============================================================
-- Seed: WS1–WS9 from .mind/workstreams.md (canonical taxonomy).
-- Read from that file; do not invent. If WS1-9 changes there, this
-- block must be updated to match.
-- ============================================================
SELECT stewards.import_workstream('WS1', 'Brain Core',
    'Pipeline, steward, commissions, classifier, retry/escalation, model selection, data safety',
    'active');
SELECT stewards.import_workstream('WS2', 'Brain UX',
    'UI panels, dialogs, kanban, file viewer, inline panel, Windows service/systray',
    'active');
SELECT stewards.import_workstream('WS3', 'Gospel Engine',
    'engine.ibeco.me, gospel-engine MCP, search/index, graph, hosted backend',
    'active');
SELECT stewards.import_workstream('WS4', 'study.ibeco.me',
    'Web UI for studies, notes, reader, public study pages',
    'active');
SELECT stewards.import_workstream('WS5', 'Memory & Process',
    '.mind/, agents, skills, voice/bias, cleanup passes, tokenomics, brain<->VS Code bridge, debug agent, Claude Code integration, Sabbath agent, pg-ai-stewards',
    'active');
SELECT stewards.import_workstream('WS6', 'Studies',
    'Scripture study output (study/, becoming/)',
    'active');
SELECT stewards.import_workstream('WS7', 'Teaching',
    'YouTube content arc, talks, public-facing teaching',
    'active');
SELECT stewards.import_workstream('WS8', 'Sunday School',
    'Calling — lesson prep, ward council',
    'active');
SELECT stewards.import_workstream('WS9', 'Other Apps',
    'Budget app, cpuchip.net rebuild, Space Center',
    'active');
