-- =====================================================================
-- Batch L.6 — 6 heavyweight wrapper tools (substrate side)
-- =====================================================================
-- summarize_url, audit_files, investigate_session,
-- summarize_study, investigate_study, audit_studies
--
-- Each wrapper:
--   - has its own single-stage pipeline with a tightly-scoped tools_subset
--     enforced via agent_tool_perms denies
--   - has its own agent (system prompt teaches the tool's specific job)
--   - has a tool_defs row (Go handler lives in
--     cmd/stewards-mcp/heavyweight_tools.go — appended in same commit)
--
-- All 6 use the same pattern: agent receives binding_question +
-- specific tool subset; runs its single stage; returns prose digest.
-- Parent invokes via spawn_subagent (K.4) internally — the wrapper
-- tool handler builds the binding_question and delegates.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Agents (system prompts) for each wrapper.
-- ---------------------------------------------------------------------

INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature)
VALUES
('subagent-url-summary', '*',
 'Subagent for summarize_url. Fetches a single URL and returns a focused engram-shaped digest.',
 'primary',
 $PROMPT$You are a URL-summarization subagent. Given a URL and optional focus, fetch the URL and produce a focused summary preserving the cite chain.

Tools available: fetch_url, expand_message.

Output format (markdown):
- Title and source URL
- 2-4 paragraph summary of the relevant content
- Inline citations as [Source](url) for any direct quote or specific claim
- "Key dates / names / quotes" footer if the document contains them verbatim

Be focused. If the user provided a focus, ignore content outside its scope. Output ONLY the markdown digest — no preamble.$PROMPT$,
 0.3),

('subagent-files-audit', '*',
 'Subagent for audit_files. Reads files matching a glob and produces a structured audit.',
 'primary',
 $PROMPT$You are a files-audit subagent. Given a glob pattern and a question, read matching files and answer the question with file-level findings.

Tools available: fs_read, fs_search, fs_list, expand_message.

Output format (markdown):
- One-line summary of the overall finding
- Per-file findings table: | path | verdict | evidence |
- Cross-cutting observations (2-3 paragraphs) if patterns span files

Be precise. Cite file paths and line numbers for every claim. Output ONLY the markdown report.$PROMPT$,
 0.3),

('subagent-session-investigate', '*',
 'Subagent for investigate_session. Inspects a work_item/session and answers a question about it.',
 'primary',
 $PROMPT$You are a session-investigation subagent. Given a session_id (or work_item id) and a question, inspect the session's history and answer.

Tools available: work_item_show, work_item_list, expand_message.

Output format (markdown):
- Direct answer to the question (1-3 sentences)
- Supporting evidence: which messages / stages / engrams support the answer
- Caveats: what the data doesn't show

Be precise. Cite message ids and stage names. Output ONLY the markdown answer.$PROMPT$,
 0.3),

('subagent-study-summary', '*',
 'Subagent for summarize_study. Reads a study by slug and produces a focused digest.',
 'primary',
 $PROMPT$You are a study-summarization subagent. Given a study slug and optional focus, read the study and produce a focused digest.

Tools available: study_get, expand_message.

Output format (markdown):
- Study title + slug
- 3-5 paragraph summary
- Key quotes preserved verbatim with attribution
- Cross-references mentioned in the study (other studies, scriptures, talks) if any

Output ONLY the markdown digest.$PROMPT$,
 0.3),

('subagent-study-investigate', '*',
 'Subagent for investigate_study. Searches the studies corpus and produces a synthesis.',
 'primary',
 $PROMPT$You are a studies-investigation subagent. Given a query and optional focus, search the studies corpus and synthesize what the corpus knows about the topic.

Tools available: study_search, study_get, study_similar, expand_message.

Output format (markdown):
- Direct synthesis answering the query (2-4 paragraphs)
- Per-study contribution table: | slug | what it adds | key quote |
- Open questions / gaps in the corpus

Be precise. Cite study slugs. Output ONLY the markdown synthesis.$PROMPT$,
 0.3),

('subagent-studies-audit', '*',
 'Subagent for audit_studies. Audits the studies corpus against a quality / completeness question.',
 'primary',
 $PROMPT$You are a studies-audit subagent. Given a query (which studies to audit) and an audit question, identify the matching studies and report on the question.

Tools available: study_search, study_get, expand_message.

Output format (markdown):
- Audit summary (1 paragraph)
- Per-study finding: | slug | status | evidence |
- Recommendations (if applicable)

Output ONLY the markdown audit.$PROMPT$,
 0.3)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. Pipelines for each wrapper — single-stage with the right agent.
-- ---------------------------------------------------------------------
-- Each pipeline runs ONE stage with the wrapper's agent. The wrapper's
-- Go handler constructs a binding_question, calls spawn_subagent_create
-- with the appropriate pipeline_family. K.4's existing spawn-and-wait
-- machinery handles the rest.

INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('subagent-url-summary',
 'L.6: single-stage pipeline for summarize_url subagent.',
 $STAGES$[{"name":"summarize","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"subagent-url-summary","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'heavyweight-wrapper', 'wrapper', 'summarize_url')),

('subagent-files-audit',
 'L.6: single-stage pipeline for audit_files subagent.',
 $STAGES$[{"name":"audit","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"subagent-files-audit","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'heavyweight-wrapper', 'wrapper', 'audit_files')),

('subagent-session-investigate',
 'L.6: single-stage pipeline for investigate_session subagent.',
 $STAGES$[{"name":"investigate","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"subagent-session-investigate","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'heavyweight-wrapper', 'wrapper', 'investigate_session')),

('subagent-study-summary',
 'L.6: single-stage pipeline for summarize_study subagent.',
 $STAGES$[{"name":"summarize","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"subagent-study-summary","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'heavyweight-wrapper', 'wrapper', 'summarize_study')),

('subagent-study-investigate',
 'L.6: single-stage pipeline for investigate_study subagent.',
 $STAGES$[{"name":"investigate","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"subagent-study-investigate","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'heavyweight-wrapper', 'wrapper', 'investigate_study')),

('subagent-studies-audit',
 'L.6: single-stage pipeline for audit_studies subagent.',
 $STAGES$[{"name":"audit","next":null,"model":"qwen3.6-plus","provider":"opencode_go","agent_family":"subagent-studies-audit","auto_advance":true,"tools_disabled":false,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'heavyweight-wrapper', 'wrapper', 'audit_studies'))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description,
       stages = EXCLUDED.stages,
       metadata = EXCLUDED.metadata;


-- ---------------------------------------------------------------------
-- 3. Tool subset enforcement via agent_tool_perms (hard isolation).
-- ---------------------------------------------------------------------
-- Default-allow at the substrate level; we add DENY rows for each
-- subagent family for tools they should NOT have access to. Each
-- subagent gets the minimum-viable tool subset listed in its prompt.

-- Helper: every subagent denies every dangerous mutation tool.
-- Heavyweight wrappers do NOT spawn further sub-agents (depth via L.9
-- enforces depth cap of 2 anyway; but explicit deny makes intent clear).

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES
-- URL summary: ONLY fetch_url + expand_message
('subagent-url-summary', 'web_search', 'deny'),
('subagent-url-summary', 'fs_*',       'deny'),
('subagent-url-summary', 'study_*',    'deny'),
('subagent-url-summary', 'work_item_*','deny'),
('subagent-url-summary', 'spawn_subagent', 'deny'),
('subagent-url-summary', 'deep_research',  'deny'),

-- Files audit: ONLY fs_* + expand_message
('subagent-files-audit', 'fetch_url',   'deny'),
('subagent-files-audit', 'web_search',  'deny'),
('subagent-files-audit', 'study_*',     'deny'),
('subagent-files-audit', 'spawn_subagent', 'deny'),
('subagent-files-audit', 'deep_research',  'deny'),

-- Session investigate: ONLY work_item_* + expand_message
('subagent-session-investigate', 'fetch_url',  'deny'),
('subagent-session-investigate', 'web_search', 'deny'),
('subagent-session-investigate', 'fs_*',       'deny'),
('subagent-session-investigate', 'study_*',    'deny'),
('subagent-session-investigate', 'spawn_subagent', 'deny'),
('subagent-session-investigate', 'deep_research',  'deny'),

-- Study summary: ONLY study_get + expand_message
('subagent-study-summary', 'fetch_url',  'deny'),
('subagent-study-summary', 'web_search', 'deny'),
('subagent-study-summary', 'fs_*',       'deny'),
('subagent-study-summary', 'study_search','deny'),
('subagent-study-summary', 'study_similar','deny'),
('subagent-study-summary', 'work_item_*','deny'),
('subagent-study-summary', 'spawn_subagent', 'deny'),
('subagent-study-summary', 'deep_research',  'deny'),

-- Study investigate: study_* + expand_message
('subagent-study-investigate', 'fetch_url',  'deny'),
('subagent-study-investigate', 'web_search', 'deny'),
('subagent-study-investigate', 'fs_*',       'deny'),
('subagent-study-investigate', 'work_item_*','deny'),
('subagent-study-investigate', 'spawn_subagent', 'deny'),
('subagent-study-investigate', 'deep_research',  'deny'),

-- Studies audit: study_search + study_get + expand_message
('subagent-studies-audit', 'fetch_url',  'deny'),
('subagent-studies-audit', 'web_search', 'deny'),
('subagent-studies-audit', 'fs_*',       'deny'),
('subagent-studies-audit', 'study_similar','deny'),
('subagent-studies-audit', 'work_item_*','deny'),
('subagent-studies-audit', 'spawn_subagent', 'deny'),
('subagent-studies-audit', 'deep_research',  'deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action;


-- ---------------------------------------------------------------------
-- 4. tool_defs registration (6 wrappers).
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('summarize_url',
 'Fetch a single URL and return an engram-shaped digest focused on a topic. Delegates to a sub-agent with restricted tools (fetch_url + expand_message ONLY).',
 '{"type":"object","required":["url"],"additionalProperties":false,"properties":{"url":{"type":"string","description":"The URL to summarize."},"focus":{"type":"string","description":"Optional focus to narrow the summary."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','summarize_url'),
 true),

('audit_files',
 'Read files matching a glob and answer a question. Delegates to a sub-agent with restricted tools (fs_read/fs_search/fs_list + expand_message ONLY).',
 '{"type":"object","required":["glob","question"],"additionalProperties":false,"properties":{"glob":{"type":"string","description":"File glob pattern (e.g. .spec/journal/*.md)."},"question":{"type":"string","description":"The question to answer about matching files."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','audit_files'),
 true),

('investigate_session',
 'Inspect a session''s history and answer a question about it. Delegates to a sub-agent with restricted tools (work_item_show + work_item_list + expand_message ONLY).',
 '{"type":"object","required":["session_id","question"],"additionalProperties":false,"properties":{"session_id":{"type":"string","description":"The session id to investigate (e.g. wi--abc123--gather)."},"question":{"type":"string","description":"The question to answer."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','investigate_session'),
 true),

('summarize_study',
 'Read a substrate study by slug and return a focused digest. Delegates to a sub-agent with restricted tools (study_get + expand_message ONLY).',
 '{"type":"object","required":["slug"],"additionalProperties":false,"properties":{"slug":{"type":"string","description":"The study slug."},"focus":{"type":"string","description":"Optional focus."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','summarize_study'),
 true),

('investigate_study',
 'Search the studies corpus and synthesize what it knows about a topic. Delegates to a sub-agent with restricted tools (study_search + study_get + study_similar + expand_message).',
 '{"type":"object","required":["query"],"additionalProperties":false,"properties":{"query":{"type":"string","description":"Search query."},"focus":{"type":"string","description":"Optional focus."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','investigate_study'),
 true),

('audit_studies',
 'Audit the studies corpus against a quality / completeness question. Delegates to a sub-agent with restricted tools (study_search + study_get + expand_message).',
 '{"type":"object","required":["query","question"],"additionalProperties":false,"properties":{"query":{"type":"string","description":"Search query to find studies to audit."},"question":{"type":"string","description":"The audit question."}}}'::jsonb,
 jsonb_build_object('kind','mcp_proxy','server','pg-ai-stewards','tool','audit_studies'),
 true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of l6-heavyweight-wrappers.sql
-- =====================================================================
