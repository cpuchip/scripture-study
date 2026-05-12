-- =====================================================================
-- Batch H.1.7a — register fs-read + pg-ai-stewards MCP servers
--
-- Two MCP servers come online here:
--   1. fs-read — new path-scoped filesystem read server, located at
--      /usr/local/bin/fs-read-mcp in the bridge image. Args carry
--      --repo-root /workspace (the bridge's read-only mount of the
--      repo root) and --allowed-paths scoped to research-agent
--      context surfaces only: .spec/journal/*, .spec/proposals/*,
--      .mind/*, docs/*.
--
--   2. pg-ai-stewards — re-uses the existing /usr/local/bin/stewards-mcp
--      binary (the same one Claude Code talks to in stdio mode).
--      When invoked without subcommand args, it serves the inbound MCP
--      tool surface: study_search, study_get, study_similar,
--      study_citations, work_item_list, work_item_show,
--      watchman_pass_show, watchman_passes_list, work_item_escalation_*.
--      We register it here as a *bridge-spawnable* MCP so the substrate's
--      internal research agent can call those tools through the bridge
--      proxy — same machinery as gospel-engine-v2 etc.
--
-- After this seed lands and the bridge restarts, `stewards-mcp bridge
-- refresh-tools` discovers the new tools and populates mcp_tool_cache.
-- A trigger from 3e.2.d then auto-promotes the cache rows into
-- stewards.tool_defs, making the tools available for agent_tool_perms
-- grants (h1-7b).
-- =====================================================================

INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'fs-read',
  'Path-scoped filesystem read for substrate-internal agents. Tools: '
    || 'fs_list, fs_read, fs_search. Scope is enforced at the MCP tool '
    || 'layer via the --allowed-paths flag — even if the bridge container '
    || 'mounts more of the repo, the agent only sees what is in scope. '
    || 'H.1.7 scope: journals (.spec/journal/*), proposals (.spec/proposals/*), '
    || 'mind files (.mind/*), and docs (docs/**). Per-pipeline scope '
    || 'extensions land with H.2/H.3 (planning pipeline gets /projects/* '
    || 'reads added).',
  'stdio',
  '/usr/local/bin/fs-read-mcp',
  ARRAY[
    '-repo-root', '/workspace',
    '-allowed-paths', '.spec/journal/*,.spec/proposals/*,.mind/*,docs/**'
  ],
  NULL,
  '{}'::jsonb,
  true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       command     = EXCLUDED.command,
       args        = EXCLUDED.args,
       env         = EXCLUDED.env,
       enabled     = EXCLUDED.enabled,
       updated_at  = now();

-- pg-ai-stewards MCP — substrate's own tool surface exposed to internal
-- agents through the bridge. The stewards-mcp binary defaults to inbound
-- stdio mode when called with no subcommand args; that's exactly what
-- the bridge needs. STEWARDS_DSN propagates from the bridge container's
-- env to the spawned process, so the substrate connects to itself via
-- the same pg:5432 service name.
--
-- Tool surface (registered by tools.go in stewards-mcp):
--   - study_search          full-text search
--   - study_get             read by slug
--   - study_similar         embedding edges
--   - study_citations       cited canonical sources
--   - work_item_list        list work_items by filter
--   - work_item_show        read a work_item
--   - watchman_pass_show    read a watchman pass
--   - watchman_passes_list  list passes
--   - work_item_escalation_list/_claim/_resolve (write — NOT granted
--                                                 to research agent)
INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'pg-ai-stewards',
  'Substrate self-surface — exposes the substrate''s own studies/work_items/'
    || 'watchman read tools to internal agents through the bridge proxy. The '
    || 'agent calls study_search/work_item_show to consult prior work before '
    || 'doing external research. Escalation write tools (work_item_escalation_*) '
    || 'exist on the same MCP but are excluded from research-agent grants — '
    || 'they belong to the operator review surface.',
  'stdio',
  '/usr/local/bin/stewards-mcp',
  ARRAY[]::text[],
  NULL,
  '{}'::jsonb,
  true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       command     = EXCLUDED.command,
       args        = EXCLUDED.args,
       env         = EXCLUDED.env,
       enabled     = EXCLUDED.enabled,
       updated_at  = now();
