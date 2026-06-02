-- =====================================================================
-- strongs1 (2026-06-02) — register strongs-concordance-mcp + grant the
-- three tools to the scripture agents (study / lesson / talk).
--
-- The Hebrew/Greek companion to webster-mcp: Strong's Concordance word-
-- study keyed to the KJV. A standalone Go MCP server
-- (github.com/cpuchip/strongs-concordance-mcp) cross-compiled into the
-- bridge image at /usr/local/bin/strongs-mcp by bridge.Dockerfile. The
-- lexicon + KJV tagging data is embedded in the binary, so no data-file
-- mount/args are needed (unlike webster's -dict).
--
-- Applied via the migration ledger (stewards-cli migrate, lexical order,
-- tracked in stewards.schema_migrations). Idempotent — ON CONFLICT keeps
-- re-runs a no-op. After this lands, `bridge refresh-tools` caches the
-- three strongs tools and the 3e.2.d auto-promote trigger lights up the
-- grants below.
--
-- Grants: study / lesson / talk — the scripture-study triple that
-- already holds the webster_define grants. Mirrors 3e2-5/3e2-7.
-- =====================================================================

INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'strongs',
  'Strong''s Concordance — Hebrew/Greek word-study keyed to the King James Bible. '
    || 'Tools: strongs_define (H#/G# -> original-language lemma + Strong''s 1890 '
    || 'definition + STEPBible BDB/Abbott-Smith modern gloss, side by side), '
    || 'strongs_search (KJV English word -> the Strong''s number(s) behind it), '
    || 'strongs_for_verse (verse reference -> the word-by-word KJV->Strong''s '
    || 'tagging). The original-language companion to webster_define. Bundled '
    || 'offline; KJV only. Glosses are a starting point for study, not doctrine.',
  'stdio',
  '/usr/local/bin/strongs-mcp',
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

-- Grant the three strongs tools to the scripture agents. Once
-- `bridge refresh-tools` caches them, the auto-promote trigger turns these
-- into live tool_defs grants.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('study',  'strongs_define',    'allow', 'manual'),
  ('study',  'strongs_search',    'allow', 'manual'),
  ('study',  'strongs_for_verse', 'allow', 'manual'),
  ('lesson', 'strongs_define',    'allow', 'manual'),
  ('lesson', 'strongs_search',    'allow', 'manual'),
  ('lesson', 'strongs_for_verse', 'allow', 'manual'),
  ('talk',   'strongs_define',    'allow', 'manual'),
  ('talk',   'strongs_search',    'allow', 'manual'),
  ('talk',   'strongs_for_verse', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
