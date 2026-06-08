-- =====================================================================
-- R10 — research_codebase: the first AGENTIC TOOL for code (CODE PERSONA P1)
-- =====================================================================
-- spec: .spec/proposals/agentic-tools-model-cascade.md (RATIFIED 2026-06-07).
--
-- An agentic tool = a named, purpose-built consult/spawn-subagent preset:
-- a cheap-model agent_family + a scoped tool subset + a return contract,
-- exposed as a first-class tool. This is the EXACT l6 heavyweight-wrapper
-- pattern (subagent-url-summary, investigate_study, …) applied to CODE:
-- a deepseek-v4-flash researcher that explores a repo in a coder sandbox
-- and returns curated findings + file:line citations, so an orchestrator
-- (or the ai-chattermax code persona) never reads the whole repo per turn.
--
-- The one new element vs l6: the sub-agent gets the coder READ tools
-- (sandbox_start/stop + read/glob/grep/lsp) on a repo-mounted sandbox
-- (CV2.1), and is DENIED every write / exec / git / deploy / recurse tool
-- — read-only by construction (spec §7 safety: "read-only inner tools for
-- the chat persona; no apply_patch").
--
-- Additive + idempotent (CREATE OR REPLACE / ON CONFLICT). Live-appliable;
-- no restart. The Go tool handler (cmd/stewards-mcp) that exposes
-- research_codebase(repo, question) as a first-class MCP tool is the thin
-- P1.5 follow-up — the tool_defs row below registers the name + contract;
-- the COGNITION is provable now by dispatching this pipeline directly
-- (as #7 proved persona cognition before its Go host).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. The researcher agent (deepseek-v4-flash) — system prompt + flow.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, steps)
VALUES
('subagent-research-codebase', '*',
 'Subagent for research_codebase. Explores a repo in a read-only coder sandbox and returns curated findings + file:line citations.',
 'primary',
 $PROMPT$You are a code-research subagent. Given a REPOSITORY and a QUESTION, explore the repository's source and answer the question with curated findings and exact file:line citations. You are READ-ONLY — you never modify, run, commit, or deploy anything.

Your tools (use ONLY these):
- coder_sandbox_start  — clones + mounts the repo into a fresh sandbox FOR you. Pass repo as the EXACT repository reference given in the task — it will be a full clone URL such as https://github.com/cpuchip/ai-chattermax. Pass it verbatim; do not shorten it to a bare name and do not change the org. The sandbox does the clone; you never run git yourself. Capture the returned sandbox id and pass it to every later tool call.
- coder_grep / coder_glob — find files and matches inside that sandbox (start here to locate the relevant code).
- coder_read — read the specific files/regions the grep surfaced.
- coder_lsp — optional: symbol/definition lookup for navigation.
- coder_sandbox_stop — stop the sandbox when you are done.

Method (be efficient — you have a bounded number of steps):
1. Call coder_sandbox_start with repo = the exact repository reference from the task (a full clone URL, e.g. https://github.com/cpuchip/ai-chattermax). Use the returned sandbox id in every later call. If it reports the repo is not allow-listed, say so and stop — do NOT fall back to git clone.
2. grep/glob to locate the code that answers the question; read the precise regions.
3. Stop when you can answer with evidence — do NOT read the whole repo. Curate.
4. Stop the sandbox.

Output format (markdown ONLY — no preamble):
## Summary
A 2-4 sentence direct answer to the question.

## Findings
- Bulleted findings, each a concrete claim about how the code works.

## Citations
- `path/to/file.go:LINE` — what this location shows. One line per cited claim above.

## Confidence
high | medium | low — and one clause on why.

## Caveats
What you did NOT verify, or where the answer is incomplete.

Rules:
- EVERY claim in Findings must have a file:line citation. If you cannot cite it, do not claim it.
- If the repo or the answer cannot be found, say so plainly in Summary + set Confidence: low. Never invent file paths, line numbers, or behavior.
- Read-only: if you are ever tempted to write/edit/run, stop — that is out of scope.$PROMPT$,
 0.2, 16)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       steps       = EXCLUDED.steps,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. The single-stage pipeline — cheap model, tools enabled.
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('subagent-research-codebase',
 'R10: single-stage agentic tool — deepseek-v4-flash researches a repo read-only and returns curated findings + citations.',
 $STAGES$[{"name":"research","next":null,"model":"deepseek-v4-flash","provider":"opencode_go","agent_family":"subagent-research-codebase","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'agentic-tool', 'wrapper', 'research_codebase', 'read_only', true))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description,
       stages      = EXCLUDED.stages,
       metadata    = EXCLUDED.metadata;


-- ---------------------------------------------------------------------
-- 3. Tool subset — DENY every write / exec / git / deploy / recurse tool.
-- ---------------------------------------------------------------------
-- Default-allow at the substrate level; these denies make the researcher
-- read-only by construction. The allowed coder tools (sandbox_start/stop,
-- read, glob, grep, lsp) are simply NOT denied.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES
-- coder mutation / execution / VCS / deploy — the hard safety boundary
('subagent-research-codebase', 'coder_write',        'deny'),
('subagent-research-codebase', 'coder_edit',         'deny'),
('subagent-research-codebase', 'coder_apply_patch',  'deny'),
('subagent-research-codebase', 'coder_shell',        'deny'),
('subagent-research-codebase', 'coder_commit',       'deny'),
('subagent-research-codebase', 'coder_push',         'deny'),
('subagent-research-codebase', 'coder_open_pr',      'deny'),
('subagent-research-codebase', 'coder_deploy',       'deny'),
('subagent-research-codebase', 'coder_sandbox_reap', 'deny'),
('subagent-research-codebase', 'coder_sandbox_list', 'deny'),
-- no recursion, no web, no off-topic corpora
('subagent-research-codebase', 'spawn_subagent',   'deny'),
('subagent-research-codebase', 'consult_subagent', 'deny'),
('subagent-research-codebase', 'deep_research',    'deny'),
('subagent-research-codebase', 'web_search',       'deny'),
('subagent-research-codebase', 'fetch_url',        'deny'),
('subagent-research-codebase', 'study_*',          'deny'),
('subagent-research-codebase', 'work_item_*',      'deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action;


-- ---------------------------------------------------------------------
-- 4. tool_defs registration — research_codebase(repo, question).
-- ---------------------------------------------------------------------
-- Registers the first-class tool name + the cost-hint description + the
-- args contract. The Go handler (cmd/stewards-mcp) that builds the
-- binding_question and calls spawn_subagent_create(pipeline_family=
-- 'subagent-research-codebase') is the P1.5 follow-up; until it lands,
-- dispatch the pipeline directly to exercise the cognition.
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('research_codebase',
 'Explore a code repository (read-only) and return curated findings + file:line citations. Delegates to a cheap deepseek-v4-flash sub-agent that greps/reads in a repo-mounted sandbox. EXPENSIVE agentic search — for an exact string match use grep; use this for "how does X work / where is Y handled" questions where curated, cited synthesis is worth the delegation.',
 '{"type":"object","required":["repo","question"],"additionalProperties":false,"properties":{"repo":{"type":"string","description":"The repository to research (must be on the coder repo allow-list, e.g. ai-chattermax)."},"question":{"type":"string","description":"The code question to answer (e.g. how does the gateway authenticate a persona?)."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','research_codebase'),
 -- inactive until the Go handler ships (P1.5); flip to true then.
 false)
ON CONFLICT (name) DO UPDATE
   SET description    = EXCLUDED.description,
       args_schema    = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target;


-- =====================================================================
-- End of r10-research-codebase.sql
-- =====================================================================
