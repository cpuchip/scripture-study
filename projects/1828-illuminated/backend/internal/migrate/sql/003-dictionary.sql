-- 003-dictionary.sql
-- Webster 1828 (full 98k corpus, D-DICT-1) + modern definitions (lazy
-- write-back from Free Dictionary API) + tier-metadata (curated 853-word
-- highlight list with study cross-refs).

CREATE TABLE IF NOT EXISTS webster_1828 (
  word            TEXT PRIMARY KEY,
  entries         JSONB NOT NULL,
  source_offsets  JSONB
);

-- Trigram on the full 1828 corpus enables class-E reach in the search UX
-- (every 1828 entry queryable, not just tier words).
CREATE INDEX IF NOT EXISTS webster_1828_word_trgm
  ON webster_1828 USING GIN (word gin_trgm_ops);

-- Modern defs cache. NULL entries + NULL error = clean 404 (we asked the
-- Free Dictionary API and it had no entry; cache that signal forever
-- unless explicitly refetched).
CREATE TABLE IF NOT EXISTS modern_defs (
  word        TEXT PRIMARY KEY,
  entries     JSONB,
  fetched_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  source      TEXT NOT NULL DEFAULT 'free-dictionary-api',
  error       TEXT
);

-- Daily fetch counter for the rate cap. One row per UTC date; the lazy
-- fetcher uses ON CONFLICT to bump the counter without races.
CREATE TABLE IF NOT EXISTS modern_defs_fetch_log (
  fetch_date  DATE PRIMARY KEY,
  attempts    INT NOT NULL DEFAULT 0
);

-- Tier metadata. The `manual` source column records whether the entry
-- came from build_data.py's automated extraction or manual-additions.json
-- (D-DICT-7).
CREATE TABLE IF NOT EXISTS tier_words (
  word            TEXT PRIMARY KEY,
  tier            TEXT NOT NULL CHECK (tier IN ('A++','A+','B','C','D')),
  study_tier      TEXT CHECK (study_tier IN ('A','B','C')),
  studies         JSONB NOT NULL DEFAULT '[]'::jsonb,
  study_excerpts  JSONB NOT NULL DEFAULT '[]'::jsonb,
  p4_score        INT,
  p4_reasons      JSONB NOT NULL DEFAULT '[]'::jsonb,
  source          TEXT NOT NULL DEFAULT 'auto' CHECK (source IN ('auto','manual'))
);
