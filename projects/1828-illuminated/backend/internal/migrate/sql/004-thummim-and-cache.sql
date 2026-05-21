-- 004-thummim-and-cache.sql
-- Forward-declared tables for phases 5+ (Thummim snapshot cache, verse
-- highlight memoization). Creating them now keeps the migration list
-- monotonic; phase-4 backend doesn't write to them but the schema is in
-- place when phase-6's sync job lands.

CREATE TABLE IF NOT EXISTS thummim_entries_cache (
  word          TEXT PRIMARY KEY,
  entries       JSONB NOT NULL,
  citations     JSONB NOT NULL DEFAULT '[]'::jsonb,
  generated_at  TIMESTAMPTZ NOT NULL,
  imported_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS verse_highlights_cache (
  verse_id      INT PRIMARY KEY REFERENCES scripture_verses(id) ON DELETE CASCADE,
  segments      JSONB NOT NULL,
  tier_set      TEXT[] NOT NULL DEFAULT '{}',
  computed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  algo_version  INT NOT NULL DEFAULT 1
);
