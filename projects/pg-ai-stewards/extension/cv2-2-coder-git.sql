-- =====================================================================
-- cv2-2 (2026-06-04) — coder-v2 CV2.2: give coder-mcp the GitHub token +
-- grant the git tools (commit/push/open_pr) to the dev agent.
--
-- The bridge resolves `$$env:GITHUB_TOKEN` against its own env (Michael's
-- fine-grained PAT in extension/.env) and passes it to coder-mcp at spawn —
-- so coder-mcp's bridge-side git/gh ops authenticate. The token NEVER enters
-- the sandbox (D-CV2.3): the agent commits locally in the worktree, coder-mcp
-- pushes + opens the PR from the bridge with the token via a one-shot
-- credential helper (not persisted in .git). Repo allow-list still constrains
-- which repos at the tool layer.
-- =====================================================================

UPDATE stewards.mcp_servers
   SET env = jsonb_build_object('GITHUB_TOKEN', '$$env:GITHUB_TOKEN'),
       updated_at = now()
 WHERE name = 'coder';

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('dev', 'coder_commit',  'allow', 'manual'),
  ('dev', 'coder_push',    'allow', 'manual'),
  ('dev', 'coder_open_pr', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
