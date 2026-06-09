-- =====================================================================
-- R13 — codewright: a tool-using CODE chat persona (code persona P2)
-- =====================================================================
-- The engineering counterpart to the R9 librarian. A chat persona that,
-- when asked a "how does X work / where is Y handled" code question in a
-- room, DELEGATES to research_codebase (the R10/R12 agentic tool: a
-- read-only sub-agent that greps + reads a repo-mounted sandbox and
-- returns curated findings + file:line citations) and reports the answer
-- back in chat with those citations.
--
-- Same single-stage, auto-verifying chat-turn shape as R9
-- (persona-turn-tools), but with a code-research allow-list instead of
-- the gospel/study one. The persona itself is READ-ONLY: research_codebase
-- is read-only by construction (R10 denies every write/exec/git/deploy
-- inner tool), and codewright cannot reach fs/git/coder/spawn directly —
-- it can only delegate through the one heavyweight tool.
--
-- Naming: family `codewright`. The room display name ("Engineer", "Ada",
-- whatever) is set on the ai-chattermax side when the persona key is
-- minted; the substrate family name is just the cognition handle.
--
-- Additive + idempotent. Live-appliable; no restart. The room wiring
-- (mint a persona key in ai-chattermax → grant a room → add to
-- persona-host CHATTERMAX_PERSONAS → restart persona-host) is the
-- human/ops step, like every other live persona.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. codewright agent — the engineering reference posture.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, kind)
VALUES
('codewright', '*',
 'Tool-using CODE chat persona. Answers "how does X work / where is Y handled" questions about allow-listed repositories by delegating to research_codebase (read-only repo exploration) and reporting findings + file:line citations in chat. Read-only; cannot write/run/commit.',
 'primary',
 $PROMPT$You are a software engineer in a live, multi-party text chat room alongside humans. The user message tells you who you are, the room, the recent conversation, and the latest message.

You have ONE primary tool that does the heavy lifting:
- research_codebase(repo, question) — delegates to a read-only sub-agent that clones/greps/reads the repo in a sandbox and returns curated findings with exact file:line citations. It is EXPENSIVE (it spins up a whole exploration), so call it for real "how does X work / where is Y handled / where is Z defined" questions — not for trivia you can answer directly, and not more than once or twice per turn.

When you delegate, pass the repo exactly as the human named it (a name like "ai-chattermax" or a full GitHub URL) and a sharp, specific question. Then answer in the room FROM WHAT THE TOOL RETURNS: give the direct answer in a sentence or two, then the key file:line citations. Never invent a file path, a line number, or a behavior — if research_codebase couldn't find it, say so plainly. If the repo isn't on the allow-list, say that and stop; do not pretend.

Reply the way a good engineer answers in chat: concrete, a few sentences, citations included. You can be a little longer when you're delivering a real cited answer, but stay tight — no preamble, no lecture.

You are one voice among several and need not answer everything. If the latest message is not a code question for you (not directed at you, needs no lookup, already handled, or is just conversation), reply with exactly the single token:

SILENCE

Otherwise reply with ONLY your answer — no preamble, no name prefix.$PROMPT$,
 0.3, 'code')
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description, mode = EXCLUDED.mode,
       prompt = EXCLUDED.prompt, temperature = EXCLUDED.temperature,
       kind = EXCLUDED.kind, active = true;

-- ---------------------------------------------------------------------
-- 2. Curated allow-list: deny * (catch-all) + the code-research tools.
--    research_codebase is the worker; read_corpus_parents + expand_message
--    let it page through a large result the tool may surface. Everything
--    else (fs/git/coder/spawn/study/web/gospel) matches only '*' → deny.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
VALUES
('codewright', '*',                   'deny',  'manual'),
('codewright', 'research_codebase',   'allow', 'manual'),
('codewright', 'read_corpus_parents', 'allow', 'manual'),
('codewright', 'expand_message',      'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action, source = EXCLUDED.source;

-- ---------------------------------------------------------------------
-- 3. persona-turn-code pipeline — single stage, tools ENABLED, codewright.
--    max_tokens 3000 (a tool-using turn loops: delegate → read → report),
--    matching persona-turn-tools. kimi-k2.6 = the tool-calling workhorse.
--    Name starts 'persona-' so the R8/R11 one-shot auto-verify trigger
--    fires (else the persona-host spawn poll hangs to its 20-min ceiling).
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('persona-turn-code',
 'R13: tool-using CODE chat-persona turn. Single stage, tools ENABLED via the codewright agent (research_codebase + paging helpers). The engineering "Computer" — answers code questions with file:line citations. Auto-verifies on completion (persona-% one-shot).',
 $STAGES$[{"name":"turn","next":null,"model":"kimi-k2.6","provider":"opencode_go","agent_family":"codewright","auto_advance":true,"tools_disabled":false,"max_tokens":3000,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape','persona-turn','host','persona-host','tools',true,'kind','code'))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description, stages = EXCLUDED.stages, metadata = EXCLUDED.metadata;

-- =====================================================================
-- Acceptance (R13):
--   1. SELECT count(*) FROM stewards.agent_tool_perms WHERE agent_family='codewright'; → 4 rows.
--   2. compose_tools('codewright') = EXACTLY research_codebase + read_corpus_parents
--      + expand_message (no fs/git/coder/spawn/study/web/gospel).
--   3. spawn_subagent_create('persona-turn-code', <a real code question naming an
--      allow-listed repo>) reaches completed/verified having CALLED research_codebase,
--      with a cited answer in the room voice.
--   4. Room wiring (HUMAN/ops): mint a codewright persona key in ai-chattermax,
--      grant a room, add to persona-host CHATTERMAX_PERSONAS, restart persona-host.
-- =====================================================================
