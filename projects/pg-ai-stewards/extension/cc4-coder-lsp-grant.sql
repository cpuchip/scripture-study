-- =====================================================================
-- cc4 (2026-06-03) — grant coder_lsp to the dev agent.
-- substrate-coding-capability CC.4 (LSP diagnostics).
--
-- The coder_lsp tool runs the language's checker (gopls / tsc / pyright,
-- already in the coder-runtime image) for fast type/compile diagnostics.
-- Idempotent.
-- =====================================================================

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('dev', 'coder_lsp', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
