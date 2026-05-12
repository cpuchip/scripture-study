-- =====================================================================
-- Batch H.3.1 — schema migrations for planning pipeline family
--
-- Three concerns, one migration (idempotent throughout):
--
--   1. studies generalization (D-H3-B ratified) — let non-scripture
--      content live in studies. Adds tags text[], source_type text,
--      project_association text. Existing rows backfill source_type
--      from kind where possible.
--
--   2. studies.file_path NOT NULL blocker (open-items §1.1) — this
--      has been failing promote_to_study since Phase D. Make it
--      nullable so work_items that reach verified maturity without
--      a file_destination don't error at promote time.
--
--   3. work_items columns for H.3 planning + D-H7 origin
--      (RATIFIED in parent batch-h-pipeline-expansion proposal):
--      - origin text DEFAULT 'human'
--          values: human|scheduled|watchman|steward|council|agent_planning
--          (the H.3 planning pipeline inserts proposed work_items
--           with origin='agent_planning' so the UI can badge them)
--      - project_association text
--          freeform per Q-H3.5 stewardship decision; identifies
--          which project the work belongs to (e.g., 'space-center',
--          'pg-ai-stewards', 'scripture-study'). Future UI surface:
--          a "known projects" view aggregates distinct values.
--      - parent_work_item_id uuid REFERENCES work_items(id)
--          for proposed work_items: points back at the planning
--          run that proposed them. ON DELETE SET NULL so deleting
--          the planning run doesn't cascade.
-- =====================================================================

-- studies generalization
ALTER TABLE stewards.studies
    ADD COLUMN IF NOT EXISTS tags text[]               NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS source_type text,
    ADD COLUMN IF NOT EXISTS project_association text;

-- Backfill source_type from kind so existing rows are cross-domain
-- searchable from day one. kind values in the seed: study, doc,
-- proposal, phase-doc, journal — map them to source_type buckets.
UPDATE stewards.studies
   SET source_type = CASE
                       WHEN kind IN ('study')             THEN 'scripture-study'
                       WHEN kind IN ('proposal')          THEN 'proposal'
                       WHEN kind IN ('journal')           THEN 'journal'
                       WHEN kind IN ('doc', 'phase-doc')  THEN 'doc'
                       ELSE kind
                     END
 WHERE source_type IS NULL;

-- studies.file_path NOT NULL drop. The check below avoids erroring
-- on a column constraint that's already been dropped on a previous
-- run — pg17/18 lacks ALTER COLUMN DROP NOT NULL IF EXISTS.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema='stewards' AND table_name='studies'
           AND column_name='file_path' AND is_nullable='NO'
    ) THEN
        ALTER TABLE stewards.studies ALTER COLUMN file_path DROP NOT NULL;
    END IF;
END $$;

-- Indexes for cross-domain studies search
CREATE INDEX IF NOT EXISTS studies_tags_gin
    ON stewards.studies USING gin(tags);
CREATE INDEX IF NOT EXISTS studies_source_type_idx
    ON stewards.studies(source_type);
CREATE INDEX IF NOT EXISTS studies_project_association_idx
    ON stewards.studies(project_association);

-- ---------------------------------------------------------------------
-- work_items: origin + project_association + parent_work_item_id
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS origin text NOT NULL DEFAULT 'human',
    ADD COLUMN IF NOT EXISTS project_association text,
    ADD COLUMN IF NOT EXISTS parent_work_item_id uuid;

-- Add the self-FK separately so the ADD COLUMN IF NOT EXISTS above is
-- idempotent on re-run. The FK can be created only once; we guard with
-- a constraint-name check.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'work_items_parent_work_item_fk'
    ) THEN
        ALTER TABLE stewards.work_items
            ADD CONSTRAINT work_items_parent_work_item_fk
            FOREIGN KEY (parent_work_item_id)
            REFERENCES stewards.work_items(id)
            ON DELETE SET NULL;
    END IF;
END $$;

-- origin CHECK constraint: known values only. agent_planning is the
-- new value contributed by this batch.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'work_items_origin_check'
    ) THEN
        ALTER TABLE stewards.work_items
            ADD CONSTRAINT work_items_origin_check
            CHECK (origin = ANY (ARRAY[
                'human', 'scheduled', 'watchman', 'steward',
                'council', 'agent_planning'
            ]));
    END IF;
END $$;

-- Indexes for the new columns. parent_work_item_id needs an index for
-- the future "show me everything this planning run produced" query.
CREATE INDEX IF NOT EXISTS work_items_origin_idx
    ON stewards.work_items(origin);
CREATE INDEX IF NOT EXISTS work_items_project_association_idx
    ON stewards.work_items(project_association)
    WHERE project_association IS NOT NULL;
CREATE INDEX IF NOT EXISTS work_items_parent_work_item_idx
    ON stewards.work_items(parent_work_item_id)
    WHERE parent_work_item_id IS NOT NULL;

-- Sanity check.
SELECT 'studies new columns:' AS check_name,
       count(*) FILTER (WHERE source_type IS NOT NULL) AS with_source_type,
       count(*) FILTER (WHERE tags IS NOT NULL) AS with_tags,
       count(*) AS total
  FROM stewards.studies;

SELECT 'work_items new columns:' AS check_name,
       count(*) FILTER (WHERE origin = 'human') AS as_human,
       count(*) FILTER (WHERE origin != 'human') AS as_other,
       count(*) AS total
  FROM stewards.work_items;
