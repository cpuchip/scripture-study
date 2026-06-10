-- 006-seed-fingerprints.sql
-- Tracks the sha256 of each embedded seed corpus so the seeders can detect
-- when the bundled data changed and re-ingest instead of skipping. Born from
-- the 2026-06-09 Webster incident: the webster_1828 table would have kept
-- serving the mislabeled 1913 corpus forever, because SeedWebster1828's
-- "skip if rows exist" guard never compared the data itself.

CREATE TABLE IF NOT EXISTS seed_fingerprints (
  corpus      TEXT PRIMARY KEY,
  sha256      TEXT NOT NULL,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
