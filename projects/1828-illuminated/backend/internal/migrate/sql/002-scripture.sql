-- 002-scripture.sql
-- Scripture canon schema. Verse text only; footnotes + headings + study
-- apparatus stripped on ingest (per D-BE-COPYRIGHT option D). The
-- abbr-on-book convention mirrors gospel-library/eng/scriptures/ paths
-- so ref strings cross-link cleanly with workspace study docs.

CREATE TABLE IF NOT EXISTS scripture_books (
  id              SERIAL PRIMARY KEY,
  volume          TEXT NOT NULL CHECK (volume IN ('ot','nt','bofm','dc','pgp')),
  abbr            TEXT NOT NULL UNIQUE,
  name            TEXT NOT NULL,
  display_order   INT NOT NULL
);

CREATE TABLE IF NOT EXISTS scripture_chapters (
  id          SERIAL PRIMARY KEY,
  book_id     INT NOT NULL REFERENCES scripture_books(id) ON DELETE CASCADE,
  chapter     INT NOT NULL,
  UNIQUE (book_id, chapter)
);

CREATE TABLE IF NOT EXISTS scripture_verses (
  id          SERIAL PRIMARY KEY,
  chapter_id  INT NOT NULL REFERENCES scripture_chapters(id) ON DELETE CASCADE,
  verse       INT NOT NULL,
  text        TEXT NOT NULL,
  text_tsv    tsvector GENERATED ALWAYS AS (to_tsvector('english', text)) STORED,
  UNIQUE (chapter_id, verse)
);

CREATE INDEX IF NOT EXISTS scripture_verses_text_tsv_idx
  ON scripture_verses USING GIN (text_tsv);

CREATE INDEX IF NOT EXISTS scripture_verses_text_trgm
  ON scripture_verses USING GIN (text gin_trgm_ops);

-- Ingest provenance. One row per re-seed; lets a future audit answer
-- "which bcbooks snapshot did we load and what strip rules applied?"
CREATE TABLE IF NOT EXISTS scripture_corpus_meta (
  id             SERIAL PRIMARY KEY,
  source         TEXT NOT NULL,
  source_sha     TEXT,
  strip_rules    JSONB NOT NULL DEFAULT '[]'::jsonb,
  loaded_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  verse_count    INT NOT NULL DEFAULT 0
);
