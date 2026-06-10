-- =====================================================================
-- R16 — room_say foundation: the persona outbox + tool (INERT)
-- =====================================================================
-- Step 1 of expressive-live-personas (spec). Lets a persona post to its
-- room MID-TURN ("🤔 hang on, searching…" → research_codebase → "found it"),
-- the way Claude Code emits text between tool calls.
--
-- ★ Design (verified 2026-06-09): the persona-host already owns the
-- session→channel map (GatewayConn.channels[ch].sessionID). So room_say
-- only needs the SESSION — it writes a persona_outbox row keyed by
-- _session_id; the host's drainer (NEXT step, Go) matches the row to the
-- channel holding that session and posts it, then stamps posted_at. No
-- session_facets / channel-id in SQL required.
--
-- INERT by construction: this registers the table + tool but grants it to
-- NO agent. compose_tools won't emit room_say for anyone (codewright/persona
-- are deny-* families) until the drainer ships and we grant + re-prompt
-- together — so a persona can never call room_say into a void where nothing
-- posts it. Pure SQL, additive, no live-host change.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. The outbox.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.persona_outbox (
    id          bigserial PRIMARY KEY,
    session_id  text NOT NULL,                 -- the dispatch session (host maps → channel)
    body        text NOT NULL,                 -- what to post in the room
    mood        text,                          -- optional emoji/state (🤔 😖 😀 …)
    created_at  timestamptz NOT NULL DEFAULT now(),
    posted_at   timestamptz                    -- set by the host once posted
);
-- The drainer scans for unposted rows; partial index keeps that cheap.
CREATE INDEX IF NOT EXISTS persona_outbox_unposted_idx
    ON stewards.persona_outbox (created_at) WHERE posted_at IS NULL;

COMMENT ON TABLE stewards.persona_outbox IS
'expressive-live-personas: mid-turn room messages a persona emits via room_say. The persona-host drains unposted rows (matching session_id → its channel) and posts them, stamping posted_at. INERT until the host drainer ships + room_say is granted.';

-- ---------------------------------------------------------------------
-- 2. room_say(body, mood?) — the agent tool. _session_id is injected by
--    the CT2.3 dispatcher; the model never supplies it.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.room_say_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess text := p_args ->> '_session_id';
    v_body text := p_args ->> 'body';
    v_mood text := nullif(btrim(coalesce(p_args ->> 'mood','')), '');
    v_id   bigint;
BEGIN
    IF v_sess IS NULL OR v_sess = '' THEN
        RETURN jsonb_build_object('error', 'no session context (room_say is only callable inside a live room turn)');
    END IF;
    IF v_body IS NULL OR length(btrim(v_body)) = 0 THEN
        RETURN jsonb_build_object('error', 'body required (the message to post in the room)');
    END IF;

    INSERT INTO stewards.persona_outbox (session_id, body, mood)
    VALUES (v_sess, v_body, v_mood)
    RETURNING id INTO v_id;

    RETURN jsonb_build_object('ok', true, 'posted_to_room', true, 'outbox_id', v_id,
        'note', 'Posted to the room. Keep working — call room_say again to give another update, then finish your turn normally.');
END;
$FN$;

-- ---------------------------------------------------------------------
-- 3. Register the tool_def (active) — but grant it to NO agent here, so it
--    stays inert until the host drainer + grants land together.
-- ---------------------------------------------------------------------
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('room_say',
 'Post a message to the room RIGHT NOW, mid-turn, before you finish — like saying "hang on, let me look that up" and then later "found it". Use it to keep people in the loop while you work (e.g. before a slow research_codebase call) and to react in the moment. Optional mood = a single emoji for your current state (🤔 thinking, 😖 frustrated, 😀, 🎲). Your final turn message still posts normally; room_say is for the in-between beats. Do not spam it — a couple of beats per turn at most.',
 '{"type":"object","required":["body"],"additionalProperties":false,"properties":{"body":{"type":"string","description":"The message to post in the room now."},"mood":{"type":"string","description":"Optional single emoji for your current state, e.g. 🤔 😖 😀 🎲."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','room_say_tool','schema','stewards'),
 true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description, args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target, active = true;

-- =====================================================================
-- Acceptance (R16):
--   1. room_say_tool({_session_id:'x', body:'hi', mood:'🤔'}) inserts an
--      unposted persona_outbox row; returns ok.
--   2. No agent sees room_say in compose_tools yet (granted to none) — INERT.
--   3. (next step) persona-host drainer posts unposted rows for its sessions
--      + stamps posted_at; then grant room_say to codewright/personas + re-prompt.
-- =====================================================================
