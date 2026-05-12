-- =====================================================================
-- H.3 followup #1 — expand fs-read MCP allow-list for science-center pivot
--
-- The H.1.7 seed (h1-7a) registered fs-read with the original scope:
--   .spec/journal/*, .spec/proposals/*, .mind/*, docs/**
--
-- For the science-center planning pipeline runs to consult prior notes,
-- expand to also include:
--   projects/space-center/*.md       (space-center-prompt.md, README.md)
--   projects/space-center/docs/**    (15+ research/planning notes)
--   projects/space-center/.spec/**   (scratch + proposals)
--
-- This is the simplified "union of all needed paths" approach. A future
-- substrate item adds per-pipeline-scoped fs-read via pipelines.fs_read_paths
-- jsonb[]; the union approach unblocks today.
--
-- node_modules/ + firmware/build/ etc. are explicitly NOT in scope —
-- the allowed-paths globs above don't match them. Even if they did,
-- the walk-allowed-filtered fix from H.3-day means we walk only by
-- the allow-list prefixes, not by user glob.
--
-- Idempotent: UPDATEs the row by name.
-- =====================================================================

UPDATE stewards.mcp_servers
   SET args = ARRAY[
           '-repo-root', '/workspace',
           '-allowed-paths',
               '.spec/journal/*'
            || ',.spec/proposals/*'
            || ',.mind/*'
            || ',docs/**'
            || ',projects/space-center/*.md'
            || ',projects/space-center/docs/**'
            || ',projects/space-center/.spec/**'
       ],
       updated_at = now()
 WHERE name = 'fs-read';

-- Sanity.
SELECT name, args FROM stewards.mcp_servers WHERE name='fs-read';
