-- =====================================================================
-- R12 — activate research_codebase for substrate-internal agents (P2)
-- =====================================================================
-- r10 registered the tool_defs row inactive, pending the Go handler
-- (P1.5, shipped 2026-06-09) + a bridge image carrying it. This flips
-- it active. Sequence matters (the grant≠catalog lesson):
--   1. bridge image rebuilt with the new stewards-mcp  ✓ (this window)
--   2. THIS file applied (active=true)
--   3. `stewards-mcp bridge refresh-tools` re-catalogs
-- After that, any agent whose perms allow 'research_codebase' can call
-- it via mcp_proxy. The A/B verdict (2026-06-09): deepseek-v4-flash is
-- the right default model for the inner researcher (correct + cited +
-- free tier; kimi-k2.6 is deeper at ~$0.84/run when it matters).
-- =====================================================================

UPDATE stewards.tool_defs
   SET active = true
 WHERE name = 'research_codebase';

-- =====================================================================
-- Acceptance (R12):
--   1. SELECT active FROM stewards.tool_defs WHERE name='research_codebase' → true.
--   2. After refresh-tools: the bridge catalog lists research_codebase.
--   3. compose_tools for an allowed family includes it; denied families
--      (persona, subagent-research-codebase itself — no recursion) do not.
-- =====================================================================
