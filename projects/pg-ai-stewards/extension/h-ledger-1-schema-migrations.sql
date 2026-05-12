-- =====================================================================
-- H-Ledger #1 — stewards.schema_migrations table
--
-- The first "real" migration in the new ledger system. Ironic but
-- correct: this file's job is to create the table that records which
-- files have run, including itself (the migrator backfills its own
-- record after successful CREATE).
--
-- Naming convention going forward:
--   h-ledger-N-NAME.sql — substrate infrastructure (ledger itself)
--   iN-NAME.sql         — Batch I (next feature batch; projects A
--                          is i1; future Batch I work is i2, i3, …)
--
-- Existing 94 .sql files keep their original names (3c2-*, h1-*,
-- h3-followup-*, etc.). The migrator processes lexically; old names
-- sort before new ones.
--
-- Migration record columns:
--   name        — base filename minus .sql (e.g. 'h3-1-schema-migrations')
--   sha256      — hex of file contents at apply time (catches drift)
--   applied_at  — when the migrator recorded it
--   notes       — 'backfilled', 'auto', 'manual', or freeform
-- =====================================================================

CREATE TABLE IF NOT EXISTS stewards.schema_migrations (
    name        text PRIMARY KEY,
    sha256      text NOT NULL,
    applied_at  timestamp with time zone NOT NULL DEFAULT now(),
    notes       text
);

COMMENT ON TABLE stewards.schema_migrations IS
'H-Ledger #1: tracks which extension/*.sql files have run. The migrator (stewards-cli migrate) reads files in lexical order, checks this table, applies unrecorded ones in a transaction, records on success. sha256 tracks file content at apply time; if a recorded file changes later, migrator warns + skips (catches the regression class where one file silently overwrites another''s changes). Backfilled with existing 94 files marked notes=''backfilled''; new migrations marked notes=''auto'' when entrypoint runs them.';

-- Small helper: is_applied(name) → bool
CREATE OR REPLACE FUNCTION stewards.migration_is_applied(p_name text)
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    SELECT EXISTS (SELECT 1 FROM stewards.schema_migrations WHERE name = p_name);
$$;

-- Small helper: mark_applied(name, sha, notes) idempotent
CREATE OR REPLACE FUNCTION stewards.migration_mark_applied(
    p_name text, p_sha256 text, p_notes text DEFAULT 'auto'
) RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_existed boolean;
BEGIN
    SELECT EXISTS (SELECT 1 FROM stewards.schema_migrations WHERE name = p_name) INTO v_existed;
    INSERT INTO stewards.schema_migrations (name, sha256, notes)
    VALUES (p_name, p_sha256, p_notes)
    ON CONFLICT (name) DO NOTHING;
    RETURN NOT v_existed;
END;
$$;

-- Sanity check.
SELECT 'schema_migrations:' AS check_name,
       count(*) AS rows,
       (SELECT count(*) FROM stewards.schema_migrations) AS row_count
  FROM stewards.schema_migrations;
