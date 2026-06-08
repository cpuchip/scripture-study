-- =====================================================================
-- CT2.3 — Self-context-management: the levers as agent-callable tools
-- =====================================================================
-- spec: .spec/proposals/substrate-self-context-management.md §3.
--
-- CT2.1 added the levers as SQL functions (keyed by message_id). CT2.2
-- made the render emit [ctx:handle] addresses + honor the states. CT2.3
-- closes the loop: a DISPATCHED agent can now CALL the levers by handle.
--
-- Two halves:
--   (Rust, tools.rs) exec_sql_fn_tool now injects the dispatch _session_id
--     into every sql_fn tool's args (backward-compatible — existing sql_fn
--     tools ignore the extra key). Requires a pg rebuild + restart.
--   (SQL, this file) handle→message_id resolution scoped to the session,
--     five tool wrappers (p_args jsonb), tool_defs(sql_fn) registration,
--     and a compose_tools gate so the context tools appear ONLY for
--     families with context_tools_enabled (decision #6).
--
-- The wrappers catch the §4 lock RAISE and return it as a structured
-- {"error": …} reply, so a locked-message re-toggle informs the model
-- instead of erroring the dispatch.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Resolve a [ctx:handle] to a message_id WITHIN the agent's session.
-- ---------------------------------------------------------------------
-- Lenient input: accepts '7a3f', 'ctx:7a3f', or '[ctx:7a3f]'. Scoped to
-- the session so handles never collide across agents.
CREATE OR REPLACE FUNCTION stewards.context_resolve_handle(p_session_id text, p_handle text)
RETURNS bigint LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_h  text;
    v_id bigint;
BEGIN
    IF p_session_id IS NULL OR p_handle IS NULL THEN RETURN NULL; END IF;
    v_h := lower(substring(p_handle FROM '([0-9a-fA-F]{4})'));
    IF v_h IS NULL THEN RETURN NULL; END IF;
    SELECT m.id INTO v_id
      FROM stewards.messages m
     WHERE m.session_id = p_session_id
       AND stewards.context_handle(m.id) = v_h
     ORDER BY m.id DESC
     LIMIT 1;
    RETURN v_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.context_resolve_handle(text, text) IS
'CT2.3: resolve a [ctx:handle] to a message_id within one session (handles are session-scoped, so no cross-agent collision).';


-- ---------------------------------------------------------------------
-- 2. Shared body for the three lockable levers (compress/mute/expand).
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards._context_tool_lockable(p_args jsonb, p_lever text)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess   text := p_args ->> '_session_id';
    v_handle text := p_args ->> 'handle';
    v_cd     int  := COALESCE(NULLIF(p_args ->> 'cooldown','')::int, 3);
    v_id     bigint;
BEGIN
    IF v_sess IS NULL THEN
        RETURN jsonb_build_object('error', 'no session context (internal: _session_id missing)');
    END IF;
    IF v_handle IS NULL OR v_handle = '' THEN
        RETURN jsonb_build_object('error', 'handle required (e.g. the 4-char [ctx:XXXX] of the message to fold)');
    END IF;
    v_id := stewards.context_resolve_handle(v_sess, v_handle);
    IF v_id IS NULL THEN
        RETURN jsonb_build_object('error', 'no message with handle ' || v_handle || ' in this context (it may be locked — its handle is hidden until the cooldown passes)');
    END IF;
    BEGIN
        RETURN CASE p_lever
            WHEN 'compress' THEN stewards.context_compress(v_id, v_cd)
            WHEN 'mute'     THEN stewards.context_mute(v_id, v_cd)
            WHEN 'expand'   THEN stewards.context_expand(v_id, v_cd)
        END;
    EXCEPTION WHEN OTHERS THEN
        -- the §4 cooldown RAISE (or any error) → structured reply, not a crash
        RETURN jsonb_build_object('error', SQLERRM);
    END;
END;
$FN$;

-- ---------------------------------------------------------------------
-- 3. The five tool wrappers (p_args jsonb → jsonb).
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_compress_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql AS $$ SELECT stewards._context_tool_lockable(p_args, 'compress'); $$;

CREATE OR REPLACE FUNCTION stewards.context_mute_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql AS $$ SELECT stewards._context_tool_lockable(p_args, 'mute'); $$;

CREATE OR REPLACE FUNCTION stewards.context_expand_tool(p_args jsonb)
RETURNS jsonb LANGUAGE sql AS $$ SELECT stewards._context_tool_lockable(p_args, 'expand'); $$;

CREATE OR REPLACE FUNCTION stewards.context_pin_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE v_sess text := p_args->>'_session_id'; v_handle text := p_args->>'handle'; v_id bigint;
BEGIN
    IF v_sess IS NULL OR v_handle IS NULL OR v_handle='' THEN
        RETURN jsonb_build_object('error','handle required'); END IF;
    v_id := stewards.context_resolve_handle(v_sess, v_handle);
    IF v_id IS NULL THEN RETURN jsonb_build_object('error','no message with handle '||v_handle); END IF;
    RETURN stewards.context_pin(v_id);
END; $FN$;

CREATE OR REPLACE FUNCTION stewards.context_unpin_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE v_sess text := p_args->>'_session_id'; v_handle text := p_args->>'handle'; v_id bigint;
BEGIN
    IF v_sess IS NULL OR v_handle IS NULL OR v_handle='' THEN
        RETURN jsonb_build_object('error','handle required'); END IF;
    v_id := stewards.context_resolve_handle(v_sess, v_handle);
    IF v_id IS NULL THEN RETURN jsonb_build_object('error','no message with handle '||v_handle); END IF;
    RETURN stewards.context_unpin(v_id);
END; $FN$;


-- ---------------------------------------------------------------------
-- 4. compose_tools gate — context tools appear ONLY for enabled families.
-- ---------------------------------------------------------------------
-- Additive filter: context_* tools are excluded unless the family has
-- context_tools_enabled. Backward-compatible — these tools are new, so
-- every existing family's tool list is unchanged.
CREATE OR REPLACE FUNCTION stewards.compose_tools(p_agent_family text)
RETURNS jsonb LANGUAGE sql STABLE AS $function$
    SELECT coalesce(jsonb_agg(
        jsonb_build_object(
            'type', 'function',
            'function', jsonb_build_object(
                'name', t.name,
                'description', t.description,
                'parameters', t.args_schema
            )
        )
        ORDER BY t.name
    ), '[]'::jsonb)
    FROM stewards.tool_defs t
    WHERE t.active
      AND stewards.tool_permission(p_agent_family, t.name) <> 'deny'
      AND (t.name NOT LIKE 'context\_%' ESCAPE '\'
           OR stewards.context_tools_on(p_agent_family))
$function$;

COMMENT ON FUNCTION stewards.compose_tools(text) IS
'Active tool_defs not denied for the family. CT2.3: context_* tools are gated — included only when the family has context_tools_enabled (decision #6).';


-- ---------------------------------------------------------------------
-- 5. tool_defs registration (sql_fn). args_schema exposes handle [+ cooldown];
--    _session_id is injected by the dispatcher, never by the model.
-- ---------------------------------------------------------------------
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('context_compress',
 'Fold one of YOUR context messages to its compact engram, reclaiming tokens. Address it by the [ctx:XXXX] handle shown in the CONTEXT PRESSURE line. The message is recoverable with context_expand. A toggle locks that message for a few turns (you will not see its handle while locked).',
 '{"type":"object","required":["handle"],"additionalProperties":false,"properties":{"handle":{"type":"string","description":"The 4-char handle of the message, e.g. 7a3f or [ctx:7a3f]."},"cooldown":{"type":"integer","description":"Optional lock cooldown in turns (default 3)."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_compress_tool','schema','stewards'), true),

('context_mute',
 'Set one of YOUR context messages aside as a recoverable tombstone (for a resolved sub-thread you are done with). Address it by its [ctx:XXXX] handle. Recoverable with context_expand. Locks the message for a few turns.',
 '{"type":"object","required":["handle"],"additionalProperties":false,"properties":{"handle":{"type":"string","description":"The 4-char handle, e.g. 7a3f."},"cooldown":{"type":"integer","description":"Optional lock cooldown in turns (default 3)."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_mute_tool','schema','stewards'), true),

('context_expand',
 'Pull one of YOUR previously folded/muted context messages back to full verbatim. Address it by its [ctx:XXXX] handle. Locks the message for a few turns.',
 '{"type":"object","required":["handle"],"additionalProperties":false,"properties":{"handle":{"type":"string","description":"The 4-char handle, e.g. 7a3f."},"cooldown":{"type":"integer","description":"Optional lock cooldown in turns (default 3)."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_expand_tool','schema','stewards'), true),

('context_pin',
 'Protect one of YOUR context messages from automatic compaction (e.g. a spec or acceptance criteria you need every turn). Address it by its [ctx:XXXX] handle. Lock-exempt; release with context_unpin.',
 '{"type":"object","required":["handle"],"additionalProperties":false,"properties":{"handle":{"type":"string","description":"The 4-char handle, e.g. 7a3f."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_pin_tool','schema','stewards'), true),

('context_unpin',
 'Release a context_pin on one of YOUR messages. Address it by its [ctx:XXXX] handle.',
 '{"type":"object","required":["handle"],"additionalProperties":false,"properties":{"handle":{"type":"string","description":"The 4-char handle, e.g. 7a3f."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_unpin_tool','schema','stewards'), true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of ct2-3-context-tools.sql
-- =====================================================================
