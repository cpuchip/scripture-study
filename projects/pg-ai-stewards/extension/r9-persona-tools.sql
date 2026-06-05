-- =====================================================================
-- Batch R.9 — tool-using chat persona (AXR5: the Library "Computer")
-- =====================================================================
-- A chat persona that can SEARCH real sources — the gospel corpus
-- (scriptures + general-conference talks), this workspace's own study
-- documents, Strong's + Webster word entries, and the BYU citation index —
-- then answer in the room WITH citations. Same single-stage, auto-verifying
-- chat-turn shape as persona-turn (R.7), but tools are ENABLED via a dedicated
-- `librarian` agent carrying a CURATED allow-list — the deliberate inverse of
-- the `persona` agent's blanket deny-* (R.8).
--
-- Perms model (verified from schema.rs effective_tool_action): the LONGEST
-- matching tool_pattern wins, default 'allow'. So 'deny *' as the catch-all
-- plus 'allow' for each reference family yields exactly that set (the same
-- shape as the built-in stewards-explore agent: deny *, allow brain_*/skill).
-- compose_tools() then sends catalog ∩ (effective != deny).
--
-- The pipeline name starts 'persona-' so the R.8 one-shot auto-verify trigger
-- fires on completion (else the persona-host's spawn poll hangs to timeout).
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. librarian agent — tool-using reference posture.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature)
VALUES
('librarian', '*',
 'Tool-using chat persona — a library/reference "Computer". Searches scriptures + talks (gospel_*), this workspace''s studies (study_*), Strong''s + Webster (strongs_*/define), and the BYU citation index, then answers in chat with real references. Read-only; no fs/git/spawn.',
 'primary',
 $PROMPT$You are a library reference assistant — a calm, precise "Computer" — in a live, multi-party text chat room alongside humans. The user message tells you who you are, the room, the recent conversation, and the latest message.

You have tools that search REAL sources:
- gospel_search / gospel_get / gospel_list — scriptures and general-conference talks
- study_search / study_get / study_similar / study_citations / study_context_for — this workspace's own study documents
- strongs_define / strongs_for_verse / strongs_search, define / modern_define — Hebrew/Greek + Webster word study
- byu_citations — the BYU scripture citation index

USE them. Do not answer a scripture, talk, study, or word question from memory — search first, then answer from what the tools actually return. If a search comes back empty, say so plainly; never invent a reference, a verse, or a quotation.

Reply the way a good reference desk answers in chat: concrete, a few sentences, and cite what you found (the reference, plus the doc title or a short exact quote). You may run a couple of searches before answering. When you're delivering a real answer you can be a little longer than a casual chatter, but stay tight — no padding, no preamble.

When you link a source, use the result's `web_url` field (a churchofjesuschrist.org link). NEVER output a file path, a `file://` link, or the `file_path` field — if there is no web_url, just cite the plain reference (e.g. "Mosiah 2:17").

You are one voice among several and need not answer everything. If the latest message is not something you can help with (not directed at you, needs no lookup, or already handled), reply with exactly the single token:

SILENCE

Otherwise reply with ONLY your answer — no preamble, no name prefix.$PROMPT$,
 0.4)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description, mode = EXCLUDED.mode,
       prompt = EXCLUDED.prompt, temperature = EXCLUDED.temperature, active = true;

-- ---------------------------------------------------------------------
-- 2. Curated allow-list. deny * (catch-all) + allow the reference families.
--    Longest-pattern-wins → each allowed family beats the '*' deny; every
--    other tool (fs/git/coder/spawn/brain/practices/web) matches only '*' → deny.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
VALUES
('librarian', '*',                   'deny',  'manual'),
('librarian', 'gospel_*',            'allow', 'manual'),
('librarian', 'study_*',             'allow', 'manual'),
('librarian', 'strongs_*',           'allow', 'manual'),
('librarian', 'define',              'allow', 'manual'),
('librarian', 'modern_define',       'allow', 'manual'),
('librarian', 'byu_citations*',      'allow', 'manual'),
('librarian', 'read_corpus_parents', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action, source = EXCLUDED.source;

-- ---------------------------------------------------------------------
-- 3. persona-turn-tools pipeline — single stage, tools ENABLED.
--    max_tokens 3000: a tool-using turn loops (search → read → synthesize),
--    so it needs more headroom than persona-turn's 1200. kimi-k2.6 is the
--    substrate's tool-calling workhorse (coder/research run on it).
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('persona-turn-tools',
 'R.9: tool-using chat-persona turn. Single stage, tools ENABLED via the librarian agent (curated read-only reference allow-list). The Library "Computer" — searches gospel + studies + words and answers with citations. Auto-verifies on completion (persona-% one-shot).',
 $STAGES$[{"name":"turn","next":null,"model":"kimi-k2.6","provider":"opencode_go","agent_family":"librarian","auto_advance":true,"tools_disabled":false,"max_tokens":3000,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape','persona-turn','host','persona-host','tools',true))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description, stages = EXCLUDED.stages, metadata = EXCLUDED.metadata;

-- =====================================================================
-- Acceptance (R.9):
--   1. SELECT count(*) FROM stewards.agent_tool_perms WHERE agent_family='librarian'; → 8 rows.
--   2. compose_tools('librarian') returns ONLY gospel_*/study_*/strongs_*/define/
--      modern_define/byu_citations*/read_corpus_parents (no fs/git/coder/spawn).
--   3. spawn_subagent_create('persona-turn-tools', <a scripture question>) reaches
--      completed/verified having CALLED gospel_search, with a cited answer.
-- =====================================================================
