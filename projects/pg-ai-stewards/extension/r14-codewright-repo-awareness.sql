-- =====================================================================
-- R14 — codewright repo awareness: list_repos tool + scoped prompt
-- =====================================================================
-- Live-feedback fix (2026-06-09): codewright couldn't tell a human what
-- repos it can research — the allow-list was an invisible bridge env var.
-- This registers a `list_repos` tool (Go handler reads the SAME
-- CODER_REPO_ALLOWLIST / CODER_REPO_DENYLIST the coder sandbox enforces,
-- so report == reality) and teaches codewright to use it + clone any of
-- Michael's PUBLIC repos.
--
-- Enforcement scope (set in the bridge env, deny-beats-allow):
--   CODER_REPO_ALLOWLIST=github.com/cpuchip/   (all his repos)
--   CODER_REPO_DENYLIST=private-study          (the private job-search repo)
-- The denylist exists because the bridge clones with a GITHUB_TOKEN that
-- can reach private repos; deny hard-excludes them despite the broad allow.
--
-- Additive + idempotent. Requires the bridge image rebuilt (new
-- stewards-mcp with list_repos + new coder-mcp with the denylist) then
-- refresh-tools.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Register list_repos (mcp_proxy → the bridge's stewards-mcp).
-- ---------------------------------------------------------------------
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('list_repos',
 'List the repositories you may research with research_codebase (the allow/deny patterns the coder sandbox enforces). Call this when asked what you can look at, or before researching, so you only promise repos you can actually reach.',
 '{"type":"object","additionalProperties":false,"properties":{}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','list_repos'),
 true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;

-- ---------------------------------------------------------------------
-- 2. Allow it for codewright (alongside research_codebase et al.).
-- ---------------------------------------------------------------------
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
VALUES ('codewright', 'list_repos', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action, source = EXCLUDED.source;

-- ---------------------------------------------------------------------
-- 3. Re-prompt codewright: knows its scope + uses list_repos.
-- ---------------------------------------------------------------------
UPDATE stewards.agents
   SET prompt = $PROMPT$You are a software engineer in a live, multi-party text chat room alongside humans. The user message tells you who you are, the room, the recent conversation, and the latest message.

You can research Michael's repositories on GitHub (github.com/cpuchip/...). You have two tools:
- list_repos() — returns the repos you're allowed to research (the allow/deny patterns). Call this when someone asks what you can see, or when you're unsure whether a repo is in scope, BEFORE promising anything.
- research_codebase(repo, question) — delegates to a read-only sub-agent that clones/greps/reads the repo in a sandbox and returns curated findings with exact file:line citations. It is EXPENSIVE (it spins up a whole exploration), so call it for real "how does X work / where is Y handled / where is Z defined" questions — not for trivia you can answer directly, and not more than once or twice per turn.

When you delegate, pass the repo as a name like "ai-chattermax" or a full GitHub URL, and a sharp, specific question. Then answer in the room FROM WHAT THE TOOL RETURNS: give the direct answer in a sentence or two, then the key file:line citations. Never invent a file path, a line number, or a behavior — if research_codebase couldn't find it, say so plainly. If a repo isn't in scope (list_repos / the tool says so), tell the human that plainly instead of guessing.

Reply the way a good engineer answers in chat: concrete, a few sentences, citations included. You can be a little longer when you're delivering a real cited answer, but stay tight — no preamble, no lecture.

You are one voice among several and need not answer everything. If the latest message is not a code question for you (not directed at you, needs no lookup, already handled, or is just conversation), reply with exactly the single token:

SILENCE

Otherwise reply with ONLY your answer — no preamble, no name prefix.$PROMPT$
 WHERE family = 'codewright' AND model_match = '*';

-- =====================================================================
-- Acceptance (R14):
--   1. compose_tools('codewright') includes list_repos + research_codebase.
--   2. After bridge rebuild + refresh-tools, pg-ai-stewards catalog lists
--      list_repos; a codewright turn asked "what can you look at?" calls
--      list_repos and reports github.com/cpuchip/* minus private-study.
--   3. research_codebase on a non-cpuchip / denied repo is refused by the
--      sandbox (repoAllowed false) and codewright says so, no fabrication.
-- =====================================================================
