-- =====================================================================
-- i1 — stewards.projects table (Projects A, first new-style migration)
--
-- Formalizes work_items.project_association into a real entity. Soft
-- reference (no FK yet) so existing rows with project_association
-- values that aren't in projects.slug don't break.
--
-- Backfill: SELECT DISTINCT project_association FROM work_items WHERE
-- project_association IS NOT NULL — insert each as a project with
-- name = slug (operator renames via UI). Existing values today are
-- 'space-center' and 'pg-ai-stewards'.
--
-- This is also the first migration applied through the new ledger
-- on a fresh bridge restart — proves the runner works end-to-end.
-- =====================================================================

CREATE TABLE IF NOT EXISTS stewards.projects (
    slug             text PRIMARY KEY,
    name             text NOT NULL,
    description      text,
    root_directory   text,                       -- nullable; future Projects B uses it
    archived         boolean NOT NULL DEFAULT false,
    created_at       timestamp with time zone NOT NULL DEFAULT now(),
    updated_at       timestamp with time zone NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.projects IS
'i1 (Batch I): formalizes work_items.project_association into a real entity. Soft reference initially (no FK) so existing rows don''t break; future migration can add the FK once the projects table is stable. Projects B (full workspace + sub-git) deferred — root_directory column is the hook.';

-- Slug regex enforced at the application layer (same shape as work_items.slug):
-- ^[a-z0-9-]+$. Not enforced via CHECK so future migrations can relax.

CREATE INDEX IF NOT EXISTS projects_archived_idx
    ON stewards.projects(archived) WHERE NOT archived;

-- Backfill from existing work_items.project_association distinct values.
-- ON CONFLICT DO NOTHING keeps it idempotent.
INSERT INTO stewards.projects (slug, name, description)
SELECT DISTINCT
       project_association AS slug,
       project_association AS name,  -- operator renames via UI
       'Backfilled from existing work_items.project_association on ' || now()::date::text
                                  AS description
  FROM stewards.work_items
 WHERE project_association IS NOT NULL
   AND project_association !~ '^\s*$'
   AND project_association ~ '^[a-z0-9-]+$'  -- skip non-conforming values
ON CONFLICT (slug) DO NOTHING;

-- Sanity check.
SELECT 'projects backfill:' AS check_name,
       count(*) AS total_projects,
       (SELECT count(DISTINCT project_association)
          FROM stewards.work_items
         WHERE project_association IS NOT NULL) AS distinct_in_work_items
  FROM stewards.projects;
