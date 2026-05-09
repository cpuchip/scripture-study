-- =====================================================================
-- Phase 3e.2.a — MCP bridge: server registry + tool cache
--
-- The substrate-internal AI agents need access to external MCP servers
-- (gospel-engine-v2, webster, becoming, yt, exa-search, byu-citations,
-- search-mcp). The architecture chosen 2026-05-08 (see docs/3e-mcp-findings.md):
--
--   - bgworker stays Rust + reqwest only
--   - sidecar `stewards-mcp` gains a `bridge` mode (long-running daemon)
--   - bridge holds the MCP client sessions and brokers tool calls between
--     the substrate and the external MCP world
--   - substrate ↔ bridge IPC will be Postgres LISTEN/NOTIFY on a work_queue
--     row of kind 'mcp_proxy' (3e.2.c — not yet wired)
--
-- This file ships only the registry + tool cache schemas plus seed rows
-- for our seven known MCP servers. Bridge connects to the substrate,
-- reads `mcp_servers WHERE enabled`, calls `tools/list` against each,
-- and upserts results into `mcp_tool_cache`.
--
-- 3e.2.c (later) adds the LISTEN/NOTIFY wire and the
-- `execute_target='mcp_proxy'` dispatch arm in the bgworker.
-- 3e.2.d (later) auto-promotes mcp_tool_cache rows into stewards.tool_defs
-- with deny-by-default agent_tool_perms.
--
-- Note on hosting (see findings doc): tonight's seed rows use absolute
-- Windows paths matching the existing .mcp.json. Future migration to a
-- Linux-binaries-in-Docker setup (3e.2.x?) will rewrite these paths to
-- in-container locations. The schema accommodates both.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. mcp_servers — registry of external MCP servers we may consume
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.mcp_servers (
    name        text PRIMARY KEY
                CHECK (name ~ '^[a-z0-9](?:[a-z0-9_-]{0,62}[a-z0-9])?$'),
    description text NOT NULL DEFAULT '',
    transport   text NOT NULL CHECK (transport IN ('stdio', 'http')),
    -- transport='stdio': command + args + env. Bridge spawns this binary
    -- and pipes JSON-RPC over stdin/stdout. command is absolute path on
    -- the bridge's host (currently Michael's Windows dev box; eventually
    -- in-container Linux paths).
    command     text,
    args        text[] NOT NULL DEFAULT ARRAY[]::text[],
    -- transport='http': remote URL (e.g. https://mcp.exa.ai/mcp?...).
    -- Bridge speaks Streamable HTTP.
    url         text,
    -- Common: env vars passed through to spawned process (stdio) or as
    -- request headers (http). SECRETS LIVE HERE — bearer tokens, API
    -- keys. Keep stewards role permissions tight.
    env         jsonb NOT NULL DEFAULT '{}'::jsonb,
    enabled     boolean NOT NULL DEFAULT false,
    -- Operational telemetry — bridge updates these on each refresh / call.
    last_health_check_at  timestamptz,
    last_tools_refresh_at timestamptz,
    last_error            text,
    notes       text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    -- Transport-specific field validation:
    --   stdio MUST have command, MAY have args + env
    --   http MUST have url, MAY have env (headers)
    CONSTRAINT mcp_servers_transport_fields CHECK (
        (transport = 'stdio' AND command IS NOT NULL AND command <> '')
        OR
        (transport = 'http'  AND url IS NOT NULL AND url <> '')
    )
);

CREATE INDEX IF NOT EXISTS mcp_servers_enabled_idx
    ON stewards.mcp_servers (enabled) WHERE enabled;

COMMENT ON TABLE stewards.mcp_servers IS
  'Registry of external MCP servers the bridge daemon connects to. '
  'Single source of truth for both the substrate (knows which tools '
  'are routable) and the bridge (knows what to spawn/connect). Secrets '
  '(bearer tokens, API keys) live in the env jsonb and are read only '
  'by the bridge process.';

-- ---------------------------------------------------------------------
-- 2. mcp_tool_cache — per-server tool catalog from tools/list
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.mcp_tool_cache (
    server_name      text NOT NULL
                     REFERENCES stewards.mcp_servers(name) ON DELETE CASCADE,
    tool_name        text NOT NULL,
    description      text NOT NULL DEFAULT '',
    title            text,
    -- The MCP server's own JSON Schema for inputs. Used by 3e.2.d to
    -- generate stewards.tool_defs.args_schema. Stored as jsonb so we
    -- can query subschemas if needed (we don't yet).
    input_schema     jsonb NOT NULL,
    -- Optional outputSchema if the MCP server declares one.
    output_schema    jsonb,
    last_refreshed_at timestamptz NOT NULL DEFAULT now(),
    -- active=false hides the tool from agents without losing its schema
    -- (e.g., during incident response when a server's tool is misbehaving).
    active           boolean NOT NULL DEFAULT true,
    PRIMARY KEY (server_name, tool_name)
);

CREATE INDEX IF NOT EXISTS mcp_tool_cache_active_idx
    ON stewards.mcp_tool_cache (active) WHERE active;

COMMENT ON TABLE stewards.mcp_tool_cache IS
  'Discovered tools from each MCP server, populated by the bridge daemon '
  'via tools/list at startup and on tools/list_changed notifications. '
  '3e.2.d will auto-create stewards.tool_defs rows from this cache, but '
  'agent_tool_perms defaults to deny — explicit grant required before '
  'agents can call any cached tool.';

-- ---------------------------------------------------------------------
-- 3. Seed rows — our seven known MCP servers
--
-- Match the shape of .mcp.json's mcpServers entries. Tokens redacted
-- via env_var-style indirection where the bridge resolves them at
-- runtime; for tonight's dev setup we inline them since the substrate
-- already holds equivalent secrets in STEWARDS_PROVIDER_*. When we
-- formalize secret rotation, switch these to env-pointer indirection
-- (e.g. {"GOSPEL_ENGINE_TOKEN": "$$env:GOSPEL_ENGINE_TOKEN"}) so
-- secrets aren't duplicated.
--
-- All `enabled = false` initially. Operator flips them to true once
-- they verify the bridge can reach them.
-- ---------------------------------------------------------------------
INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled) VALUES
  ('gospel-engine-v2',
   'Hosted scripture/talk/manual lookup. Tools: gospel_search, gospel_get, gospel_list. Currently the highest-leverage external MCP for substrate-internal study agents.',
   'stdio',
   'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/gospel-engine/gospel-mcp.exe',
   ARRAY[]::text[],
   NULL,
   jsonb_build_object(
     'GOSPEL_ENGINE_URL', 'https://engine.ibeco.me',
     'GOSPEL_ENGINE_TOKEN', '$$env:GOSPEL_ENGINE_TOKEN',
     'GOSPEL_AUTO_UPDATE', 'true'
   ),
   false),

  ('webster',
   'Webster 1828 + modern dictionary lookup. Tools: define, webster_define, webster_search, webster_search_definitions, modern_define.',
   'stdio',
   'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/webster-mcp/webster-mcp.exe',
   ARRAY['-dict', 'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/webster-mcp/data/webster1828.json.gz'],
   NULL,
   '{}'::jsonb,
   false),

  ('yt',
   'YouTube transcript download + lookup. Tools: yt_download, yt_get, yt_list, yt_search.',
   'stdio',
   'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/yt-mcp/yt-mcp.exe',
   ARRAY['serve'],
   NULL,
   jsonb_build_object(
     'YT_DIR', 'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/yt',
     'YT_STUDY_DIR', 'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/study/yt',
     'YT_COOKIE_FILE', 'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/yt/cookies.txt'
   ),
   false),

  ('byu-citations',
   'BYU Scripture Citation Index. Tools: byu_citations, byu_citations_books, byu_citations_bulk.',
   'stdio',
   'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/byu-citations/byu-citations.exe',
   ARRAY[]::text[],
   NULL,
   '{}'::jsonb,
   false),

  ('becoming',
   'Personal brain + practices via ibeco.me. Tools: brain_*, get_today, list_practices, etc.',
   'stdio',
   'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/becoming/mcp.exe',
   ARRAY[]::text[],
   NULL,
   jsonb_build_object(
     'BECOMING_URL', 'https://ibeco.me',
     'BECOMING_TOKEN', '$$env:BECOMING_TOKEN'
   ),
   false),

  ('search',
   'DuckDuckGo web search. Tool: web_search (fast, no API key).',
   'stdio',
   'C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/search-mcp/search-mcp.exe',
   ARRAY[]::text[],
   NULL,
   '{}'::jsonb,
   false),

  ('exa-search',
   'Exa AI neural web search. Remote MCP server (Streamable HTTP transport). Tool: web_search_exa.',
   'http',
   NULL,
   ARRAY[]::text[],
   'https://mcp.exa.ai/mcp?tools=web_search_exa',
   '{}'::jsonb,
   false)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       transport   = EXCLUDED.transport,
       command     = EXCLUDED.command,
       args        = EXCLUDED.args,
       url         = EXCLUDED.url,
       env         = EXCLUDED.env,
       updated_at  = now();

-- ---------------------------------------------------------------------
-- 4. View for at-a-glance bridge state
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW stewards.mcp_bridge_state AS
SELECT s.name AS server,
       s.transport,
       s.enabled,
       s.last_health_check_at,
       s.last_tools_refresh_at,
       coalesce((SELECT count(*) FROM stewards.mcp_tool_cache c
                  WHERE c.server_name = s.name AND c.active), 0) AS active_tools,
       s.last_error
  FROM stewards.mcp_servers s
 ORDER BY s.name;

COMMENT ON VIEW stewards.mcp_bridge_state IS
  'At-a-glance view of bridge connectivity. After bridge refresh-tools '
  'runs, active_tools should be > 0 for every enabled server. last_error '
  'NULL means most recent health check passed.';
