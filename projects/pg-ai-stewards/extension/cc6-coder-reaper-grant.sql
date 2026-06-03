-- =====================================================================
-- cc6 (2026-06-03) — grant the sandbox visibility + reaper tools to dev.
-- substrate-coding-capability CC.6 (hardening).
--
-- coder_sandbox_list (visibility) + coder_sandbox_reap (remove sandboxes
-- older than max_age_minutes; coder-mcp also reaps >2h on startup). Resource
-- caps (mem/cpu/pids) already ship in CC.1's Provision. The deploy-secret
-- broker is deferred to v2/Dokploy (CC.7) — a local sidecar needs no creds.
-- Idempotent.
-- =====================================================================

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('dev', 'coder_sandbox_list', 'allow', 'manual'),
  ('dev', 'coder_sandbox_reap', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
