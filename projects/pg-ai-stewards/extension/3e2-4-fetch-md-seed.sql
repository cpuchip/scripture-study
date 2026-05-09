-- =====================================================================
-- Phase 3e.2 follow-up — register fetch-md MCP server + research grants
--
-- Adds an 8th external MCP server (a Go binary at scripts/fetch-md-mcp/
-- that wraps Mozilla Readability + html-to-markdown for clean, agent-
-- friendly web fetches). Tools: fetch_url, fetch_urls, extract_links,
-- fetch_url_raw. Plain HTTP only — no JS rendering — sufficient for
-- docs sites, blog posts, READMEs, Wikipedia.
--
-- Grants the four tools to the `research` agent family alongside the
-- existing `web_search_exa` grant — together those are the v1
-- "research outside the canon" surface.
-- =====================================================================

INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'fetch-md',
  'Web page fetcher — clean markdown via Mozilla Readability + JohannesKaufmann/html-to-markdown. Tools: fetch_url, fetch_urls, extract_links, fetch_url_raw. Plain HTTP only (no JS rendering); good for docs sites, blog posts, READMEs, Wikipedia.',
  'stdio',
  'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/fetch-md-mcp/fetch-md-mcp.exe',
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

-- Note: fresh installs won't have any rows in stewards.mcp_tool_cache
-- for fetch-md until the bridge daemon does its first refresh-tools
-- pass. The tool_defs auto-promote (3e.2.d trigger) fires from cache
-- inserts; until that happens the grants below match nothing and the
-- research agent's compose_tools result simply won't include them.
-- After `stewards-mcp bridge refresh-tools` runs once, the tool_defs
-- materialize and the grants light up.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('research', 'fetch_url',     'allow', 'manual'),
  ('research', 'fetch_urls',    'allow', 'manual'),
  ('research', 'extract_links', 'allow', 'manual'),
  ('research', 'fetch_url_raw', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
