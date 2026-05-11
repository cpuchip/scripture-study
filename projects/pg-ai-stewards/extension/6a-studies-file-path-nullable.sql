-- =====================================================================
-- Phase 6 / Batch G.1 — studies.file_path nullable
--
-- The studies table was originally designed for workspace-imported
-- markdown files where file_path is the canonical source. Substrate-
-- promoted studies don't have a file_path until the optional
-- materialization step (G.4) writes them to disk. NULL means "exists
-- in DB; no on-disk file yet."
--
-- Discovered during Phase D.5 smoke: the sabbath gate refused
-- promotion until sabbath ran (correct); after sabbath ran, the
-- second promote attempt hit this NOT NULL because work_item_promote_to_study
-- only passes (slug, kind, title, body, frontmatter) — no file_path.
-- =====================================================================

ALTER TABLE stewards.studies ALTER COLUMN file_path DROP NOT NULL;

COMMENT ON COLUMN stewards.studies.file_path IS
'Batch G.1 (2026-05-11): NULL = exists in DB only; no on-disk file. Substrate-promoted studies default to NULL. Once G.4 materialization fires (via pending_file_writes + stewards-cli materialize-writes), this column is updated to the actual on-disk path. Workspace-imported studies always have file_path populated from import time.';
