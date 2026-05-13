-- =====================================================================
-- i2 — work_items.project_association → projects.slug FK constraint
--
-- Hardens the soft reference established in i1. Safe to apply now
-- because:
--   * i1 backfilled stewards.projects from DISTINCT non-NULL values
--     of work_items.project_association.
--   * Pre-flight verified zero orphans: 27 NULL, 20 'space-center',
--     5 'pg-ai-stewards' — both non-NULL values exist in projects.
--
-- Semantics ratified 2026-05-12:
--   * ON UPDATE CASCADE  — slug renames propagate to work_items rows
--   * ON DELETE RESTRICT — can't delete a project that owns work_items
--     (archive via projects UI instead; hard delete is not exposed).
-- NULL is allowed (work_items without a project).
-- =====================================================================

-- Belt-and-suspenders: surface any pre-existing orphan before the
-- ALTER would otherwise fail with a less-readable error.
DO $$
DECLARE v_orphans int;
BEGIN
    SELECT count(*) INTO v_orphans
      FROM stewards.work_items wi
     WHERE wi.project_association IS NOT NULL
       AND NOT EXISTS (
           SELECT 1 FROM stewards.projects p WHERE p.slug = wi.project_association
       );
    IF v_orphans > 0 THEN
        RAISE EXCEPTION 'i2: % work_items have project_association values not in projects.slug; backfill before applying FK', v_orphans;
    END IF;
END
$$;

ALTER TABLE stewards.work_items
    ADD CONSTRAINT work_items_project_association_fkey
    FOREIGN KEY (project_association)
    REFERENCES stewards.projects(slug)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

COMMENT ON CONSTRAINT work_items_project_association_fkey ON stewards.work_items IS
'i2 (2026-05-12): hardens the soft reference to stewards.projects. ON UPDATE CASCADE propagates slug renames; ON DELETE RESTRICT prevents deleting a project with work_items (archive instead). NULL allowed (work_items without a project).';
