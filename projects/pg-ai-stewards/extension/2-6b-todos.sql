-- Phase 2.6b — Todos as persistent connector vertices
--
-- Live-DB migration. Same pattern as 2-6a-workstreams.sql.
--
-- Todos live in their own table (not stewards.studies) because
-- their lifecycle is different: rapid mutation (status changes)
-- vs. write-once+versioned (studies/proposals). The :HAS_TODO
-- edge connects parent (Workstream | Study | Phase | other Todo)
-- to the :Todo vertex. Parent fields are denormalized on the row
-- for fast roll-up audits without a graph walk.
--
-- Single-write rule: stewards.create_todo() writes BOTH the row
-- AND the graph edge in one transaction. Never INSERT directly.

-- BEGIN;  -- (folded into extension_sql_file! v0.2.0; CREATE EXTENSION already wraps in tx)
-- ============================================================
-- Table: stewards.todos
-- ============================================================
CREATE TABLE IF NOT EXISTS stewards.todos (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Optional human-friendly slug. NULL allowed because most todos
    -- are session-scoped and don't earn a slug; long-lived ones do.
    slug        text UNIQUE,
    title       text NOT NULL,
    body        text NOT NULL DEFAULT '',
    status      text NOT NULL DEFAULT 'open'
                CHECK (status IN ('open', 'in_progress', 'done', 'dropped')),

    -- Parent denormalization. Parent kind is the AGE label string
    -- ('Workstream', 'Study', 'Phase', 'Todo'). Parent slug is the
    -- vertex's slug/id (workstream id for :Workstream, slug for
    -- everything else). Both nullable for free-floating todos but
    -- the create function rejects that path.
    parent_kind text,
    parent_slug text,

    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    completed_at    timestamptz,

    created_by_session   text,
    completed_by_session text,

    frontmatter jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS todos_status_idx ON stewards.todos (status);
CREATE INDEX IF NOT EXISTS todos_parent_idx
    ON stewards.todos (parent_kind, parent_slug);
CREATE INDEX IF NOT EXISTS todos_created_idx
    ON stewards.todos (created_at DESC);

-- touch trigger
CREATE OR REPLACE FUNCTION stewards.touch_todo() RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        NEW.updated_at := now();
        -- Auto-stamp completed_at on transition into a terminal state.
        IF NEW.status IN ('done', 'dropped')
           AND OLD.status NOT IN ('done', 'dropped')
           AND NEW.completed_at IS NULL
        THEN
            NEW.completed_at := now();
        END IF;
    END IF;
    RETURN NEW;
END;
$func$;

DROP TRIGGER IF EXISTS todos_touch ON stewards.todos;
CREATE TRIGGER todos_touch
    BEFORE UPDATE ON stewards.todos
    FOR EACH ROW EXECUTE FUNCTION stewards.touch_todo();

-- ============================================================
-- Function: create_todo(parent_kind, parent_slug, title, body, slug, session)
--
-- Single-write rule: row + :HAS_TODO edge in one transaction.
-- Returns the new uuid as text.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.create_todo(
    p_parent_kind text,
    p_parent_slug text,
    p_title       text,
    p_body        text DEFAULT '',
    p_slug        text DEFAULT NULL,
    p_session     text DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_id uuid;
BEGIN
    PERFORM stewards.ensure_studies_graph();

    IF p_parent_kind IS NULL OR p_parent_slug IS NULL THEN
        RAISE EXCEPTION 'create_todo: parent_kind and parent_slug required (free-floating todos not allowed)';
    END IF;
    IF p_parent_kind NOT IN ('Workstream', 'Study', 'Phase', 'Todo') THEN
        RAISE EXCEPTION 'create_todo: parent_kind must be one of Workstream|Study|Phase|Todo, got %', p_parent_kind;
    END IF;

    INSERT INTO stewards.todos (slug, title, body, parent_kind, parent_slug, created_by_session)
    VALUES (p_slug, p_title, p_body, p_parent_kind, p_parent_slug, p_session)
    RETURNING id INTO v_id;

    -- MERGE the :Todo vertex AND the :HAS_TODO edge from parent.
    -- AGE matches parent by the appropriate key per kind: Workstream
    -- by id, everything else by slug.
    EXECUTE
        $cy$
        SELECT * FROM cypher('stewards_graph', $$
            MERGE (t:Todo {id: $id})
            SET t.title = $title, t.status = 'open', t.slug = $slug
            RETURN t
        $$, $1) AS (v agtype)
        $cy$
    USING (jsonb_build_object(
        'id',    v_id::text,
        'title', p_title,
        'slug',  coalesce(p_slug, '')
    )::text)::ag_catalog.agtype;

    -- Edge from parent. Workstream uses {id}, others use {slug}.
    IF p_parent_kind = 'Workstream' THEN
        EXECUTE
            $cy$
            SELECT * FROM cypher('stewards_graph', $$
                MATCH (p:Workstream {id: $parent_slug}), (t:Todo {id: $id})
                MERGE (p)-[r:HAS_TODO]->(t)
                SET r.provenance = 'declared',
                    r.confidence = 1.0,
                    r.source = 'create_todo'
                RETURN r
            $$, $1) AS (v agtype)
            $cy$
        USING (jsonb_build_object(
            'parent_slug', p_parent_slug,
            'id',          v_id::text
        )::text)::ag_catalog.agtype;
    ELSE
        -- Study / Phase / Todo all matched by slug. parent_kind goes
        -- into the Cypher label via format() since AGE doesn't bind
        -- labels through params.
        EXECUTE format(
            $cy$
            SELECT * FROM cypher('stewards_graph', $$
                MATCH (p:%s {slug: $parent_slug}), (t:Todo {id: $id})
                MERGE (p)-[r:HAS_TODO]->(t)
                SET r.provenance = 'declared',
                    r.confidence = 1.0,
                    r.source = 'create_todo'
                RETURN r
            $$, $1) AS (v agtype)
            $cy$,
            p_parent_kind
        )
        USING (jsonb_build_object(
            'parent_slug', p_parent_slug,
            'id',          v_id::text
        )::text)::ag_catalog.agtype;
    END IF;

    RETURN v_id;
END;
$func$;

-- ============================================================
-- Function: complete_todo(id_or_slug, session)
--
-- Marks a todo done, syncs the :Todo vertex's status property.
-- Accepts either uuid or slug for ergonomics from the CLI.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.complete_todo(
    p_ref     text,           -- uuid string or slug
    p_session text DEFAULT NULL,
    p_status  text DEFAULT 'done'
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_id uuid;
BEGIN
    PERFORM stewards.ensure_studies_graph();

    IF p_status NOT IN ('done', 'dropped', 'in_progress', 'open') THEN
        RAISE EXCEPTION 'complete_todo: invalid status %', p_status;
    END IF;

    -- Resolve ref to id. Try uuid cast first; fall back to slug lookup.
    BEGIN
        v_id := p_ref::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        SELECT id INTO v_id FROM stewards.todos WHERE slug = p_ref;
        IF v_id IS NULL THEN
            RAISE EXCEPTION 'complete_todo: no todo with id-or-slug %', p_ref;
        END IF;
    END;

    UPDATE stewards.todos
       SET status = p_status,
           completed_by_session = CASE WHEN p_status IN ('done','dropped')
                                       THEN p_session ELSE completed_by_session END
     WHERE id = v_id;

    -- Sync status onto the :Todo vertex.
    EXECUTE
        $cy$
        SELECT * FROM cypher('stewards_graph', $$
            MATCH (t:Todo {id: $id})
            SET t.status = $status
            RETURN t
        $$, $1) AS (v agtype)
        $cy$
    USING (jsonb_build_object(
        'id',     v_id::text,
        'status', p_status
    )::text)::ag_catalog.agtype;

    RETURN v_id;
END;
$func$;

-- ============================================================
-- Function: todo_rollup_audit()
--
-- Returns rows where the parent/child status invariants are broken:
--   - parent is done but has open/in_progress children
--   - all children are done but parent is still open/in_progress
--
-- This is the function Watchman (2.7) will call periodically to find
-- dangling state. Called manually for now via CLI.
--
-- Currently checks one level (parent → todo); recursive todo trees
-- (todo with sub-todos) audited too because parent_kind='Todo' is
-- a valid case.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.todo_rollup_audit()
RETURNS TABLE (
    finding      text,    -- 'parent_done_open_children' | 'all_done_parent_open'
    parent_kind  text,
    parent_slug  text,
    parent_title text,    -- best-effort label for human reading
    todo_count   int,
    open_count   int,
    done_count   int
)
LANGUAGE plpgsql STABLE AS $func$
BEGIN
    RETURN QUERY
    WITH child_counts AS (
        SELECT t.parent_kind,
               t.parent_slug,
               COUNT(*)::int                                          AS todo_count,
               COUNT(*) FILTER (WHERE t.status IN ('open','in_progress'))::int AS open_count,
               COUNT(*) FILTER (WHERE t.status = 'done')::int         AS done_count
          FROM stewards.todos t
         GROUP BY t.parent_kind, t.parent_slug
    ),
    -- Self-check on todo-as-parent. We treat a Todo parent as "done"
    -- if its row.status is done.
    parents AS (
        SELECT cc.*,
               CASE
                   WHEN cc.parent_kind = 'Todo'
                   THEN (SELECT pt.status FROM stewards.todos pt
                          WHERE pt.id::text = cc.parent_slug
                             OR pt.slug    = cc.parent_slug LIMIT 1)
                   -- For non-Todo parents we can't generically know
                   -- "done" without per-kind status. Treat as 'open'
                   -- so we only flag the all-done-parent-open finding
                   -- via Watchman's later kind-specific query.
                   ELSE 'open'
               END AS parent_status,
               CASE
                   WHEN cc.parent_kind = 'Workstream'
                   THEN (SELECT name FROM stewards.workstreams WHERE id = cc.parent_slug)
                   WHEN cc.parent_kind = 'Study'
                   THEN (SELECT title FROM stewards.studies WHERE slug = cc.parent_slug)
                   WHEN cc.parent_kind = 'Todo'
                   THEN (SELECT title FROM stewards.todos
                          WHERE id::text = cc.parent_slug
                             OR slug    = cc.parent_slug LIMIT 1)
                   ELSE NULL
               END AS parent_title
          FROM child_counts cc
    )
    -- Finding 1: parent done, open children
    SELECT 'parent_done_open_children'::text,
           p.parent_kind, p.parent_slug, p.parent_title,
           p.todo_count, p.open_count, p.done_count
      FROM parents p
     WHERE p.parent_status IN ('done', 'dropped')
       AND p.open_count > 0
    UNION ALL
    -- Finding 2: all children done, parent still open (only meaningful
    -- for parent_kind='Todo' until per-kind status is added).
    SELECT 'all_done_parent_open'::text,
           p.parent_kind, p.parent_slug, p.parent_title,
           p.todo_count, p.open_count, p.done_count
      FROM parents p
     WHERE p.parent_kind = 'Todo'
       AND p.parent_status IN ('open', 'in_progress')
       AND p.open_count = 0
       AND p.done_count > 0
    ORDER BY 1, 2, 3;
END;
$func$;

-- ============================================================
-- Read function: list_todos(parent_kind, parent_slug, status)
--
-- All three filters optional. NULL = no filter on that dimension.
-- Returns the row plus a parent_title best-effort label.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.list_todos(
    p_parent_kind text DEFAULT NULL,
    p_parent_slug text DEFAULT NULL,
    p_status      text DEFAULT NULL
) RETURNS TABLE (
    id           uuid,
    slug         text,
    title        text,
    status       text,
    parent_kind  text,
    parent_slug  text,
    created_at   timestamptz,
    completed_at timestamptz
)
LANGUAGE sql STABLE AS $func$
    SELECT t.id, t.slug, t.title, t.status,
           t.parent_kind, t.parent_slug,
           t.created_at, t.completed_at
      FROM stewards.todos t
     WHERE (p_parent_kind IS NULL OR t.parent_kind = p_parent_kind)
       AND (p_parent_slug IS NULL OR t.parent_slug = p_parent_slug)
       AND (p_status      IS NULL OR t.status      = p_status)
     ORDER BY t.parent_kind, t.parent_slug, t.created_at DESC;
$func$;

-- COMMIT; -- (folded into extension_sql_file! v0.2.0; CREATE EXTENSION already wraps in tx)