-- =====================================================================
-- R21 — room_react(emoji): personas react to the message they're working
-- =====================================================================
-- Michael: "I'd love the personas to react with emoji! that'd be such a
-- treat." The host already automates 👀 (eyes) on the trigger message;
-- room_react lets the MODEL deliberately add one more — 🎲 on a clutch
-- roll, 😂 at a good line. Rides the persona_outbox like room_say; the
-- host applies it to the turn's trigger message (the one wearing the 👀),
-- so no message-id plumbing reaches the model. Idempotent.

ALTER TABLE stewards.persona_outbox ADD COLUMN IF NOT EXISTS react_emoji text;

CREATE OR REPLACE FUNCTION stewards.room_react_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess  text := p_args ->> '_session_id';
    v_emoji text := nullif(btrim(coalesce(p_args ->> 'emoji','')), '');
    v_id    bigint;
BEGIN
    IF v_sess IS NULL OR v_sess = '' THEN
        RETURN jsonb_build_object('error', 'no session context (room_react is only callable inside a live room turn)');
    END IF;
    IF v_emoji IS NULL THEN
        RETURN jsonb_build_object('error', 'emoji required, e.g. 🎲 or 😂');
    END IF;
    IF length(v_emoji) > 16 THEN
        RETURN jsonb_build_object('error', 'one emoji only');
    END IF;

    INSERT INTO stewards.persona_outbox (session_id, body, react_emoji)
    VALUES (v_sess, '', v_emoji)
    RETURNING id INTO v_id;

    RETURN jsonb_build_object('ok', true, 'outbox_id', v_id,
        'note', 'Reaction ' || v_emoji || ' lands on the message you are answering. Keep working.');
END;
$FN$;

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('room_react',
 'React to the message you are currently answering with a single emoji — 🎲 for a clutch roll, 😂 at a good line, ❤️ for a great moment. The reaction appears on that message in the room immediately. One emoji per call; use sparingly (at most one or two per turn), and only when a human would genuinely react.',
 '{"type":"object","required":["emoji"],"additionalProperties":false,"properties":{"emoji":{"type":"string","description":"A single emoji, e.g. 🎲 😂 ❤️ 😱 👏."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','room_react_tool','schema','stewards'),
 true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description, args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target, active = true;

-- Grant alongside room_say: the chat-persona families.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('persona',    'room_react', 'allow', 'manual'),
  ('librarian',  'room_react', 'allow', 'manual'),
  ('codewright', 'room_react', 'allow', 'manual'),
  ('gamemaster', 'room_react', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action, source = EXCLUDED.source;

-- Prompt nudges (targeted replace; a miss is a no-op verified by acceptance).
UPDATE stewards.agents
   SET prompt = replace(prompt,
       'Use it to feel alive and present — a reaction, a "hmm, let me think", a roll — but stay in character and do not spam it (a beat or two at most).',
       'Use it to feel alive and present — a reaction, a "hmm, let me think", a roll — but stay in character and do not spam it (a beat or two at most). You can also react to the message you are answering with room_react(emoji) — 🎲 on a great roll, 😂 at a good joke — one emoji, used sparingly.')
 WHERE family = 'persona' AND model_match = '*';

UPDATE stewards.agents
   SET prompt = replace(prompt,
       'Mid-turn beats go out via room_say(body, mood) — mood is one emoji (😏 😱 🎲 😅 🤔).',
       'Mid-turn beats go out via room_say(body, mood) — mood is one emoji (😏 😱 🎲 😅 🤔). React to the message you''re answering with room_react(emoji) — 🎲 for a clutch roll, 😂 for a good line; sparingly.')
 WHERE family = 'gamemaster' AND model_match = '*';

-- R21b (first live test): kimi answered SILENCE and skipped the tool — the
-- prompts framed SILENCE as "do nothing", so reacting got skipped with the
-- reply. Teach that a reaction is NOT a message. Guarded for idempotence.
UPDATE stewards.agents
   SET prompt = replace(prompt,
       'with room_react(emoji) — 🎲 on a great roll, 😂 at a good joke — one emoji, used sparingly.',
       'with room_react(emoji) — 🎲 on a great roll, 😂 at a good joke — one emoji, used sparingly. A reaction is NOT a message: you can call room_react and STILL reply SILENCE — when a moment needs no words, the emoji alone is the right response.')
 WHERE family = 'persona' AND model_match = '*'
   AND prompt NOT LIKE '%STILL reply SILENCE%';

UPDATE stewards.agents
   SET prompt = replace(prompt,
       'with room_react(emoji) — 🎲 for a clutch roll, 😂 for a good line; sparingly.',
       'with room_react(emoji) — 🎲 for a clutch roll, 😂 for a good line; sparingly. A reaction is NOT a message: you can call room_react and STILL reply SILENCE — sometimes the emoji alone is the whole response.')
 WHERE family = 'gamemaster' AND model_match = '*'
   AND prompt NOT LIKE '%STILL reply SILENCE%';

-- =====================================================================
-- Acceptance (R21):
--   1. room_react_tool({_session_id:'x', emoji:'🎲'}) inserts an unposted
--      outbox row with react_emoji='🎲' and body=''.
--   2. compose_tools('gamemaster') and ('persona') include room_react.
--   3. Both prompts contain 'room_react' after the replaces.
--   4. (host) drainer turns the row into a reaction frame on the turn's
--      eyed message — the room sees 🎲 land while the persona works.
-- =====================================================================
