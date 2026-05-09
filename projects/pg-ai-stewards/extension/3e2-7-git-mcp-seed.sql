-- =====================================================================
-- Phase 3d v1 (2026-05-09) — register git-mcp + grant to study agent
--
-- New scripts/git-mcp/ Go MCP server that wraps git/gh with a tight
-- allow-list (branch namespace agent/<pipeline>/<work-item>-<slug>,
-- protected branches refused, --force not exposed, env-only token).
--
-- enabled=true at seed time, but the server itself will fail any
-- network operation until GITHUB_TOKEN lives in the bridge daemon's
-- env (.env file at projects/pg-ai-stewards/extension/.env). Until
-- then, local git ops (status, branch, commit) work fine.
--
-- Initial grants: study agent only (study-write pipeline is the
-- first real consumer). Other agents stay deny-by-default; operator
-- (Michael) opens grants per-agent as new pipelines need them.
-- =====================================================================

INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'git',
  'Sandboxed git/gh wrapper for substrate-driven repo ops. Tools: '
    || 'git_clone, git_status, git_branch_create, git_add, git_commit, '
    || 'git_push, gh_pr_create, gh_issue_create. Branch namespace locked '
    || 'to agent/<pipeline>/<work-item-id>-<slug>; protected branches '
    || '(main, master, release/*) refused at the tool layer; --force '
    || 'and other destructive ops not exposed.',
  'stdio',
  '/usr/local/bin/git-mcp',
  ARRAY[]::text[],
  NULL,
  -- Token resolution: bridge sets GITHUB_TOKEN in its process env
  -- from the .env file. git-mcp inherits via os.Environ() when it
  -- spawns git/gh subprocesses. The $$env: placeholder here makes
  -- intent visible in the registry but is not strictly needed since
  -- git-mcp reads from its own process env, not from this jsonb.
  jsonb_build_object(
    'GITHUB_TOKEN', '$$env:GITHUB_TOKEN'
  ),
  true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       command     = EXCLUDED.command,
       args        = EXCLUDED.args,
       env         = EXCLUDED.env,
       enabled     = EXCLUDED.enabled,
       updated_at  = now();

-- Grant the 8 git-mcp tools to the study agent. Once bridge refresh-tools
-- runs, the trigger from 3e.2.d auto-promotes mcp_tool_cache rows into
-- stewards.tool_defs, so these grants light up the moment the cache
-- has the matching tool_names.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('study', 'git_clone',         'allow', 'manual'),
  ('study', 'git_status',        'allow', 'manual'),
  ('study', 'git_branch_create', 'allow', 'manual'),
  ('study', 'git_add',           'allow', 'manual'),
  ('study', 'git_commit',        'allow', 'manual'),
  ('study', 'git_push',          'allow', 'manual'),
  ('study', 'gh_pr_create',      'allow', 'manual')
  -- gh_issue_create deliberately not granted to study; opens for
  -- watchman-consolidator or a future research pipeline that triages
  -- findings into issues.
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
