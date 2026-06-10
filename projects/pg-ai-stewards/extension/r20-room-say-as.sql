-- =====================================================================
-- R20 — room_say(as_character): the cast system's substrate half (DH-2)
-- =====================================================================
-- One registered persona, many voices: room_say gains an optional
-- as_character — the named cast member (shopkeep, villain, mob, PC) this
-- line is spoken by. The host passes it through as the platform message's
-- subPersona; the platform auto-creates the cast member on first use and
-- attributes the line ("Grimble: Best prices in the realm"). Display is
-- decoupled from cognition: one DM turn can voice several characters with
-- several room_say calls. Idempotent.

ALTER TABLE stewards.persona_outbox ADD COLUMN IF NOT EXISTS sub_persona text;

CREATE OR REPLACE FUNCTION stewards.room_say_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess text := p_args ->> '_session_id';
    v_body text := p_args ->> 'body';
    v_mood text := nullif(btrim(coalesce(p_args ->> 'mood','')), '');
    v_as   text := nullif(btrim(coalesce(p_args ->> 'as_character','')), '');
    v_id   bigint;
BEGIN
    IF v_sess IS NULL OR v_sess = '' THEN
        RETURN jsonb_build_object('error', 'no session context (room_say is only callable inside a live room turn)');
    END IF;
    IF v_body IS NULL OR length(btrim(v_body)) = 0 THEN
        RETURN jsonb_build_object('error', 'body required (the message to post in the room)');
    END IF;
    IF v_as IS NOT NULL AND length(v_as) > 60 THEN
        RETURN jsonb_build_object('error', 'as_character must be a short name (60 chars max)');
    END IF;

    INSERT INTO stewards.persona_outbox (session_id, body, mood, sub_persona)
    VALUES (v_sess, v_body, v_mood, v_as)
    RETURNING id INTO v_id;

    RETURN jsonb_build_object('ok', true, 'posted_to_room', true, 'outbox_id', v_id,
        'note', 'Posted to the room' || CASE WHEN v_as IS NOT NULL THEN ' as ' || v_as ELSE '' END ||
                '. Keep working — call room_say again for another beat or another character, then finish your turn normally.');
END;
$FN$;

-- Refresh the tool_def so the model sees as_character.
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('room_say',
 'Post a message to the room RIGHT NOW, mid-turn, before you finish. Use it to keep people in the loop while you work and to react in the moment. Optional mood = a single emoji for your current state (🤔 😖 😀 🎲). Optional as_character = speak AS a named character you are voicing (a shopkeep, a villain, an NPC) — the room shows that name as the speaker, and the character is created on first use. One turn can voice several characters with several room_say calls. Your final turn message still posts under your own name; do not spam — a few beats per turn at most.',
 '{"type":"object","required":["body"],"additionalProperties":false,"properties":{"body":{"type":"string","description":"The message to post in the room now."},"mood":{"type":"string","description":"Optional single emoji for your current state, e.g. 🤔 😖 😀 🎲."},"as_character":{"type":"string","description":"Optional: the named character speaking this line (e.g. \"Grimble the shopkeep\"). The room attributes the message to this name."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','room_say_tool','schema','stewards'),
 true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description, args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target, active = true;

-- =====================================================================
-- Acceptance (R20):
--   1. room_say_tool({_session_id:'x', body:'hi', as_character:'Grimble'})
--      inserts an unposted row with sub_persona='Grimble'.
--   2. compose_tools shows as_character in room_say's schema for granted agents.
--   3. (host) drainer passes sub_persona → platform message subPersona →
--      the room renders "Grimble" as the speaker + roster nests him.
-- =====================================================================
