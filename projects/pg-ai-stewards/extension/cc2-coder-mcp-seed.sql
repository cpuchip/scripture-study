-- =====================================================================
-- cc2 (2026-06-03) — register coder-mcp + grant its tools to the dev agent.
-- substrate-coding-capability CC.2.
--
-- coder-mcp is a stdio MCP server (cross-compiled into the bridge image at
-- /usr/local/bin/coder-mcp) that exposes the coding tool surface. Each tool
-- operates on a named sandbox (the work_item id); the sandbox-manager spawns
-- hardened coder-runtime containers against the host docker daemon.
--
-- Applied via the migration ledger (stewards-cli migrate). Idempotent.
-- After refresh-tools caches the tools, the 3e.2.d auto-promote trigger turns
-- the grants below into live tool_defs.
--
-- Grants: `dev` agent family (the code-write pipeline, CC.3, will add its own
-- agent grants). Read-only-substrate discipline does NOT apply here — these
-- tools mutate a disposable sandbox, never the live workspace.
-- =====================================================================

INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'coder',
  'Substrate coding capability — write, build, test, and run code in an isolated, '
    || 'hardened, ephemeral sandbox (Go + Node/TS + Python + LSP). Tools: '
    || 'coder_sandbox_start / coder_sandbox_stop (lifecycle), coder_write / coder_read / '
    || 'coder_edit / coder_apply_patch (files), coder_shell (build/test/run — the '
    || 'ground-truth gate), coder_glob / coder_grep (search). Each tool takes a `sandbox` '
    || 'id (the work_item id). The coder never touches the live workspace.',
  'stdio',
  '/usr/local/bin/coder-mcp',
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

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('dev', 'coder_sandbox_start', 'allow', 'manual'),
  ('dev', 'coder_sandbox_stop',  'allow', 'manual'),
  ('dev', 'coder_write',         'allow', 'manual'),
  ('dev', 'coder_read',          'allow', 'manual'),
  ('dev', 'coder_edit',          'allow', 'manual'),
  ('dev', 'coder_apply_patch',   'allow', 'manual'),
  ('dev', 'coder_shell',         'allow', 'manual'),
  ('dev', 'coder_glob',          'allow', 'manual'),
  ('dev', 'coder_grep',          'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
