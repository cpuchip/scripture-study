-- =====================================================================
-- Phase: bridge-in-docker (2026-05-09) — migrate mcp_servers.command +
-- args from Windows .exe paths to Linux /usr/local/bin paths matching
-- the bridge image (projects/pg-ai-stewards/extension/bridge.Dockerfile).
--
-- The 3e2-1 seed and 3e2-4 seed (fetch-md) shipped with absolute
-- Windows paths because the bridge initially ran on the host. Now
-- that the bridge is a docker-compose service alongside pg, the
-- spawn-target binaries live inside the bridge container at
-- /usr/local/bin/*. This migration is idempotent: re-running it on
-- the live DB is a no-op once paths are aligned.
-- =====================================================================

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/gospel-mcp',
    args    = ARRAY[]::text[],
    updated_at = now()
 WHERE name = 'gospel-engine-v2';

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/webster-mcp',
    args    = ARRAY['-dict', '/opt/webster/data/webster1828.json.gz'],
    updated_at = now()
 WHERE name = 'webster';

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/yt-mcp',
    args    = ARRAY['serve'],
    env     = jsonb_build_object(
        'YT_DIR',         '/opt/yt/yt',
        'YT_STUDY_DIR',   '/opt/yt/study',
        'YT_COOKIE_FILE', '/opt/yt/cookies.txt'
    ),
    updated_at = now()
 WHERE name = 'yt';

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/byu-citations',
    args    = ARRAY[]::text[],
    updated_at = now()
 WHERE name = 'byu-citations';

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/becoming-mcp',
    args    = ARRAY[]::text[],
    updated_at = now()
 WHERE name = 'becoming';

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/search-mcp',
    args    = ARRAY[]::text[],
    updated_at = now()
 WHERE name = 'search';

UPDATE stewards.mcp_servers SET
    command = '/usr/local/bin/fetch-md-mcp',
    args    = ARRAY[]::text[],
    updated_at = now()
 WHERE name = 'fetch-md';

-- exa-search is HTTP transport, no command — left unchanged.
