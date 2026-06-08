-- =====================================================================
-- CT2.7c — persona + room facets (the chat-persona headline case)
-- =====================================================================
-- spec §7. Adds the `persona` and `room` audience facets so durable notes can
-- be scoped to one persona (Starlet, across her rooms) or one location
-- (everyone in 10-Forward). The faceted mechanism (7a) already supports any
-- selector key; this just POPULATES persona/room from a session→facets map
-- that persona-host writes (the Go wiring is gate 7c-host).
--
-- Also upgrades the remember/forget DEFAULT: a chat-persona dispatch now
-- defaults its note audience to {persona:self} (was {agent_family:self}, which
-- for chat personas was too coarse — all roleplay personas share family
-- 'persona'). Owner (the cap + forget scope) likewise becomes the persona.
-- Pure SQL; the persona-host call lands in 7c-host.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. session → (persona, room) map. persona-host writes it via
--    set_session_facets after it establishes a (persona, room) session.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.session_facets (
    session_id text PRIMARY KEY,
    persona    text,
    room       text,
    updated_at timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE stewards.session_facets IS
'CT2 §7c: per-session persona/room facets (written by persona-host). dispatch_facets reads these so durable notes can be scoped {persona:…} / {room:…}.';

CREATE OR REPLACE FUNCTION stewards.set_session_facets(p_session_id text, p_persona text, p_room text)
RETURNS void LANGUAGE sql AS $$
    INSERT INTO stewards.session_facets (session_id, persona, room)
    VALUES (p_session_id, nullif(btrim(p_persona),''), nullif(btrim(p_room),''))
    ON CONFLICT (session_id) DO UPDATE
        SET persona = EXCLUDED.persona, room = EXCLUDED.room, updated_at = now();
$$;
COMMENT ON FUNCTION stewards.set_session_facets(text,text,text) IS
'CT2 §7c: persona-host calls this once per (persona,room) session so dispatch_facets can expose persona/room.';


-- ---------------------------------------------------------------------
-- 2. dispatch_facets — now merges persona + room from the map.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.dispatch_facets(p_agent_family text, p_session_id text)
RETURNS jsonb LANGUAGE sql STABLE AS $$
    SELECT jsonb_strip_nulls(jsonb_build_object(
        'global',       true,
        'session',      p_session_id,
        'agent_family', p_agent_family,
        'kind',         (SELECT a.kind FROM stewards.agents a
                          WHERE a.family = p_agent_family AND a.kind IS NOT NULL LIMIT 1),
        'pipeline',     (SELECT w.pipeline_family FROM stewards.work_items w
                          WHERE p_session_id = ANY(w.session_ids) ORDER BY w.id DESC LIMIT 1),
        'persona',      (SELECT sf.persona FROM stewards.session_facets sf WHERE sf.session_id = p_session_id),
        'room',         (SELECT sf.room    FROM stewards.session_facets sf WHERE sf.session_id = p_session_id)
    ));
$$;
COMMENT ON FUNCTION stewards.dispatch_facets(text, text) IS
'CT2 §7: the facets of the current dispatch (global/session/agent_family/kind/pipeline + persona/room from session_facets). A self-note renders iff dispatch_facets @> note.audience.';


-- ---------------------------------------------------------------------
-- 3. remember — persona-aware owner + default audience.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.remember_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess    text  := p_args ->> '_session_id';
    v_note    text  := p_args ->> 'note';
    v_aud     jsonb := p_args -> 'audience';
    v_tags    text[];
    v_facets  jsonb := stewards.dispatch_facets(COALESCE(stewards.session_agent_family(v_sess), '~none~'), v_sess);
    v_persona text  := v_facets ->> 'persona';
    v_fam     text  := NULLIF(v_facets ->> 'agent_family', '~none~');
    v_owner   text;
    v_count   int;
    v_id      bigint;
    v_cap     int := 40;
BEGIN
    IF v_note IS NULL OR length(btrim(v_note)) = 0 THEN
        RETURN jsonb_build_object('error', 'note text required');
    END IF;

    -- owner (cap + forget scope) and default audience: persona > family > session.
    v_owner := COALESCE(v_persona, v_fam, v_sess);
    IF v_aud IS NULL OR jsonb_typeof(v_aud) <> 'object' OR v_aud = '{}'::jsonb THEN
        v_aud := CASE
            WHEN v_persona IS NOT NULL THEN jsonb_build_object('persona', v_persona)
            WHEN v_fam     IS NOT NULL THEN jsonb_build_object('agent_family', v_fam)
            ELSE jsonb_build_object('session', v_sess) END;
    END IF;

    IF p_args ? 'tags' AND jsonb_typeof(p_args -> 'tags') = 'array' THEN
        SELECT array_agg(t) INTO v_tags FROM jsonb_array_elements_text(p_args -> 'tags') t;
    END IF;

    SELECT count(*) INTO v_count FROM stewards.agent_self_notes WHERE created_by = v_owner;
    IF v_count >= v_cap THEN
        RETURN jsonb_build_object('error',
            format('note budget full (%s/%s for %s) — forget() an integrated one first', v_count, v_cap, v_owner));
    END IF;

    INSERT INTO stewards.agent_self_notes (note, audience, tags, created_by, created_session)
    VALUES (v_note, v_aud, COALESCE(v_tags, '{}'), v_owner, v_sess)
    RETURNING id INTO v_id;

    RETURN jsonb_build_object('ok', true,
        'handle', stewards.context_note_handle(v_id), 'audience', v_aud, 'note_id', v_id);
END;
$FN$;


-- ---------------------------------------------------------------------
-- 4. forget — persona-aware owner.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.forget_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess    text  := p_args ->> '_session_id';
    v_handle  text  := lower(substring(COALESCE(p_args ->> 'handle', '') FROM '([0-9a-fA-F]{4})'));
    v_facets  jsonb := stewards.dispatch_facets(COALESCE(stewards.session_agent_family(v_sess), '~none~'), v_sess);
    v_owner   text  := COALESCE(v_facets ->> 'persona', NULLIF(v_facets ->> 'agent_family','~none~'), v_sess);
    v_deleted int;
BEGIN
    IF v_handle IS NULL THEN
        RETURN jsonb_build_object('error', 'handle required (the [note:xxxx] of the note to drop)');
    END IF;
    WITH del AS (
        DELETE FROM stewards.agent_self_notes n
         WHERE stewards.context_note_handle(n.id) = v_handle
           AND (v_facets @> n.audience OR n.created_by = v_owner)
        RETURNING n.id
    )
    SELECT count(*) INTO v_deleted FROM del;
    IF v_deleted = 0 THEN
        RETURN jsonb_build_object('error', 'no note [note:' || v_handle || '] you can forget in this context');
    END IF;
    RETURN jsonb_build_object('ok', true, 'forgotten', v_handle, 'count', v_deleted);
END;
$FN$;


-- =====================================================================
-- End of ct2-7c-persona-room-facets.sql
-- =====================================================================
