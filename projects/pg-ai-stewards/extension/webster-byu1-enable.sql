-- =====================================================================
-- webster-byu1 (2026-06-02) — operator flips webster + byu-citations ON.
--
-- 3e2-1-mcp-bridge-schemas.sql seeds these two (and becoming) with
-- enabled=false, per its comment "Operator flips them to true once they
-- verify the bridge can reach them." Their Linux binaries are already in
-- the bridge image (bridge.Dockerfile) and 3e2-6 already points their
-- command/args at /usr/local/bin/*, so enabling is the only step; after
-- `bridge refresh-tools` caches their tools they auto-promote to tool_defs
-- and the EXISTING study/lesson/talk/journal/review grants (3e2-5) light up.
--
-- becoming stays disabled deliberately — its brain_*/practice/note tools
-- mutate personal data and remain Claude-Code / operator-controlled (the
-- read-only-substrate discipline named in 3e2-5).
--
-- Filename sorts after 3e2-1 / 3e2-6 so a fresh-rebuild migrate sets the
-- final state to enabled=true. Idempotent.
-- =====================================================================

UPDATE stewards.mcp_servers
   SET enabled = true,
       updated_at = now()
 WHERE name IN ('webster', 'byu-citations')
   AND enabled = false;
