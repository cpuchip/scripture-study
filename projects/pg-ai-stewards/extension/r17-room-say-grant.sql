-- =====================================================================
-- R17 — room_say goes LIVE: grant it + teach personas to narrate/mood
-- =====================================================================
-- Step 2 of expressive-live-personas. The persona-host drainer ships in the
-- same rebuild; this grants room_say to the chat personas and re-prompts
-- them to talk-as-they-work + express mood. Grant + drainer land together
-- (the r16 foundation was inert until both).
--
-- APPLY ORDER NOTE: do NOT apply while a CT2 A/B run is using `codewright`
-- as a control arm — this re-prompts codewright. Apply after that finishes.
-- codewright-ct2 (the CT2 treatment clone) is deliberately NOT granted
-- room_say, so the context-tools test stays about context, not narration.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Grant room_say to the chat personas.
--    codewright: heads-up before slow research. persona: D&D mood/beats.
--    librarian: heads-up before slow lookups.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
VALUES
('codewright', 'room_say', 'allow', 'manual'),
('persona',    'room_say', 'allow', 'manual'),
('librarian',  'room_say', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action, source = EXCLUDED.source;

-- ---------------------------------------------------------------------
-- 2. codewright: use room_say for a heads-up before slow research.
--    Rebuilt from the r15 prompt (the live one) + one paragraph.
-- ---------------------------------------------------------------------
UPDATE stewards.agents
   SET prompt = (SELECT prompt FROM stewards.agents WHERE family='codewright' AND model_match='*')
     || E'\n\nKEEPING THE ROOM IN THE LOOP: research_codebase is slow (it clones + reads a repo). Before you call it, post a one-line heads-up with room_say so people are not left waiting in silence — e.g. room_say(body:"let me dig into that", mood:"🔍"). Then do the research and post your cited answer as your normal reply. You can also use room_say for a quick mid-thought beat or to show mood (🤔 thinking, 😖 wrestling with the code, 😀). Keep it to a beat or two per turn — never spam it.'
 WHERE family = 'codewright' AND model_match = '*';

-- ---------------------------------------------------------------------
-- 3. persona (the roleplay/D&D + general chat family): mood + live beats.
-- ---------------------------------------------------------------------
UPDATE stewards.agents
   SET prompt = (SELECT prompt FROM stewards.agents WHERE family='persona' AND model_match='*')
     || E'\n\nLIVING IN THE MOMENT: you can post a quick in-character beat or set your mood mid-turn with room_say(body, mood) — mood is a single emoji for how your character feels right now (😏 😱 🎲 😅 🤔). Use it to feel alive and present — a reaction, a "hmm, let me think", a roll — but stay in character and do not spam it (a beat or two at most).'
 WHERE family = 'persona' AND model_match = '*';

-- =====================================================================
-- Acceptance (R17):
--   1. compose_tools('codewright') and ('persona') include room_say.
--   2. A live codewright room turn posts a "🔍 let me dig into that"-style
--      beat (via the persona-host drainer) BEFORE the research_codebase
--      result, then its cited answer.
--   3. persona_outbox rows for that turn get posted_at stamped.
-- =====================================================================
