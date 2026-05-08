-- =====================================================================
-- Phase 3c.3.3 followup — agent_tool_perms.source provenance column
--
-- The 2026-05-08 overnight session surfaced a subtle bug: the importer
-- in cmd/stewards-cli/internal/importer/agents.go does
--     DELETE FROM stewards.agent_tool_perms WHERE agent_family = $1;
--     INSERT ... -- one row per declared tool from frontmatter
-- on every default-variant import. That delete is too broad — it wipes
-- substrate-internal *broadcast* perms (e.g. 3c.2.5's blanket
-- `study_*: allow` for all non-watchman families) that aren't declared
-- in any agent's frontmatter.
--
-- Surfaced when run #2 of the FtC/WtL voice experiment ran with no
-- corpus tools because my prior reimport had nuked the broadcast.
-- The kimi-tuned agent honestly refused to fabricate (the prompt's
-- discipline rule worked exactly as designed) and the bug surfaced.
--
-- Two fixes:
--   1. Workaround (already applied 2026-05-08): declared `study_*`
--      in study agent frontmatter. Survives reimport via the normal
--      path, but doesn't help OTHER agents (lesson, talk, etc.) that
--      need study_* via the broadcast.
--   2. Proper fix (this file): add a `source` column to track row
--      origin. Importer's DELETE filters to source='frontmatter';
--      broadcasts are tagged source='broadcast' and survive.
--
-- This is idempotent: ADD COLUMN IF NOT EXISTS, retroactive UPDATE
-- only touches rows currently set to the default. Foldback into
-- lib.rs alongside other 3c.3.x SQL files.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Schema migration
-- ---------------------------------------------------------------------
ALTER TABLE stewards.agent_tool_perms
  ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'frontmatter'
  CHECK (source IN ('frontmatter', 'broadcast', 'manual'));

COMMENT ON COLUMN stewards.agent_tool_perms.source IS
  'Provenance of this perm row: frontmatter (declared in agent .agent.md), '
  'broadcast (substrate-internal SQL grant, e.g. 3c.2.5 study_*), '
  'manual (one-off psql grant). Importer only deletes/rebuilds rows '
  'with source=frontmatter on agent re-import; broadcast/manual rows '
  'are preserved.';

-- ---------------------------------------------------------------------
-- 2. Retroactively mark known broadcasts.
--
-- The 3c.2.5 broadcast: study_* on all non-watchman families.
-- After 2026-05-08 the `study` family also has study_* declared in
-- its frontmatter; for that family the row's authoritative source is
-- frontmatter. For OTHER families (lesson, talk, journal, etc.),
-- study_* is a broadcast.
--
-- A row's source is broadcast iff:
--   - tool_pattern = 'study_*'
--   - AND the agent's frontmatter doesn't declare study_*
--
-- We don't have a clean way to query "is this in the agent's frontmatter"
-- from SQL alone, but we know that as of 2026-05-08:
--   - study agent has study_* in frontmatter (added 4da7b77)
--   - kimi-* and qwen-* study variants also have it
--   - No other agent .agent.md declares study_* explicitly
-- So: mark study_* on all families EXCEPT 'study' as broadcast.
-- ---------------------------------------------------------------------
UPDATE stewards.agent_tool_perms
   SET source = 'broadcast'
 WHERE tool_pattern = 'study_*'
   AND agent_family <> 'study'
   AND source = 'frontmatter';  -- only touch defaults; manual/broadcast unchanged

-- ---------------------------------------------------------------------
-- 3. Verification view (for soak / debugging)
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW stewards.agent_tool_perms_by_source AS
SELECT source, count(*) AS row_count, count(DISTINCT agent_family) AS family_count
  FROM stewards.agent_tool_perms
 GROUP BY source;
