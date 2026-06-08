-- =====================================================================
-- CT2.1 — Self-context-management: the SQL state model
-- =====================================================================
-- The agent-governed context window (spec:
--   .spec/proposals/substrate-self-context-management.md, §§1–6).
--
-- This is the SAFE, inert-by-itself layer: per-message context state +
-- addressable handles + the anti-thrash circuit breaker + a pressure
-- helper + the agent's levers as SQL functions. NOTHING reads the new
-- state yet — compose_messages (the render, CT2.2) is unchanged — so
-- live-applying this cannot alter behavior. It is purely additive:
--   ALTER TABLE ... ADD COLUMN IF NOT EXISTS  (defaults preserve today's behavior)
--   CREATE OR REPLACE FUNCTION                (idempotent)
--
-- Decisions adopted (2026-06-05, spec §"Decisions adopted"):
--   1. N (lock cooldown) = 3 turns.
--   2. Handle = substr(md5(message_id::text),1,4) → stable across turns.
--   3. Mute = tombstone (recoverable), not deletion.
--   8. Lock applies to compress/expand/mute; pin/unpin are lock-exempt.
--
-- "Turn" is defined as the monotonic message count in a session — the
-- number that compose_messages would see when it builds the next turn.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Per-message context state + the lock column.
-- ---------------------------------------------------------------------
ALTER TABLE stewards.messages
    ADD COLUMN IF NOT EXISTS context_state text NOT NULL DEFAULT 'verbatim';

ALTER TABLE stewards.messages
    ADD COLUMN IF NOT EXISTS locked_until_turn int;

-- Constrain the state vocabulary. Dropped-then-added so a re-run picks up
-- any future vocabulary change idempotently.
ALTER TABLE stewards.messages
    DROP CONSTRAINT IF EXISTS messages_context_state_check;
ALTER TABLE stewards.messages
    ADD CONSTRAINT messages_context_state_check
    CHECK (context_state IN ('verbatim', 'compressed', 'muted', 'pinned'));

COMMENT ON COLUMN stewards.messages.context_state IS
'CT2.1: agent-governed render state. verbatim=full (default); compressed=render its engram; muted=recoverable tombstone; pinned=full + exempt from automatic compaction. Honored by compose_messages in CT2.2.';
COMMENT ON COLUMN stewards.messages.locked_until_turn IS
'CT2.1 circuit breaker: when set, this message is under cooldown until session_turn() reaches it. While locked, CT2.2 strips its handle from the render so the agent cannot re-toggle it. NULL = not locked.';


-- ---------------------------------------------------------------------
-- 2. Addressable handle — short, stable hash of the message id.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_handle(p_message_id bigint)
RETURNS text LANGUAGE sql IMMUTABLE AS $$
    SELECT substr(md5(p_message_id::text), 1, 4);
$$;

COMMENT ON FUNCTION stewards.context_handle(bigint) IS
'CT2.1: the [ctx:7a3f] handle the agent uses to address a message. Stable across turns (pure function of the id); same derivation in SQL and the CT2.2 Rust render.';


-- ---------------------------------------------------------------------
-- 3. "Turn" counter — monotonic message count per session.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.session_turn(p_session_id text)
RETURNS int LANGUAGE sql STABLE AS $$
    SELECT COALESCE(count(*), 0)::int
      FROM stewards.messages
     WHERE session_id = p_session_id;
$$;

COMMENT ON FUNCTION stewards.session_turn(text) IS
'CT2.1: the current turn = monotonic count of messages in the session. The lock cooldown is expressed in these units (locked_until_turn = session_turn + N).';


-- ---------------------------------------------------------------------
-- 4. Is a message currently under the cooldown lock?
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_is_locked(p_message_id bigint)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS (
        SELECT 1
          FROM stewards.messages m
         WHERE m.id = p_message_id
           AND m.locked_until_turn IS NOT NULL
           AND stewards.session_turn(m.session_id) < m.locked_until_turn
    );
$$;

COMMENT ON FUNCTION stewards.context_is_locked(bigint) IS
'CT2.1: true while a message is under cooldown (session_turn < locked_until_turn). CT2.2 enforces by absence (strips the handle); this function is the SQL-layer guard the lever functions also honor.';


-- ---------------------------------------------------------------------
-- 5. The core applicator + the agent's five levers.
-- ---------------------------------------------------------------------
-- _context_apply is internal. The five public levers wrap it:
--   context_compress / context_mute / context_expand  → lockable toggles
--   context_pin / context_unpin                        → lock-exempt
--
-- A lockable toggle on a currently-locked message is REFUSED here (belt),
-- in addition to CT2.2's handle-stripping (suspenders). pin/unpin neither
-- set nor are blocked by the lock (decision #8: pin is voluntary, not
-- thrash-prone).
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards._context_apply(
    p_message_id bigint,
    p_state      text,
    p_lockable   boolean,
    p_cooldown   int DEFAULT 3
) RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_session text;
    v_turn    int;
    v_lock    int;
BEGIN
    SELECT session_id INTO v_session FROM stewards.messages WHERE id = p_message_id;
    IF v_session IS NULL THEN
        RAISE EXCEPTION 'context lever: no message % (handle %)',
            p_message_id, stewards.context_handle(p_message_id);
    END IF;

    IF p_lockable AND stewards.context_is_locked(p_message_id) THEN
        SELECT locked_until_turn INTO v_lock FROM stewards.messages WHERE id = p_message_id;
        RAISE EXCEPTION
            'context lock: message % (handle %) is under cooldown until turn % (now %); cannot re-toggle yet',
            p_message_id, stewards.context_handle(p_message_id), v_lock,
            stewards.session_turn(v_session);
    END IF;

    v_turn := stewards.session_turn(v_session);

    UPDATE stewards.messages
       SET context_state     = p_state,
           locked_until_turn = CASE WHEN p_lockable
                                    THEN v_turn + GREATEST(p_cooldown, 0)
                                    ELSE locked_until_turn END
     WHERE id = p_message_id;

    RETURN jsonb_build_object(
        'message_id',        p_message_id,
        'handle',            stewards.context_handle(p_message_id),
        'state',             p_state,
        'current_turn',      v_turn,
        'locked_until_turn', CASE WHEN p_lockable THEN v_turn + GREATEST(p_cooldown, 0) ELSE NULL END
    );
END;
$FN$;

-- Lockable toggles ---------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_compress(p_message_id bigint, p_cooldown int DEFAULT 3)
RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_apply(p_message_id, 'compressed', true, p_cooldown);
$$;

CREATE OR REPLACE FUNCTION stewards.context_mute(p_message_id bigint, p_cooldown int DEFAULT 3)
RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_apply(p_message_id, 'muted', true, p_cooldown);
$$;

CREATE OR REPLACE FUNCTION stewards.context_expand(p_message_id bigint, p_cooldown int DEFAULT 3)
RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_apply(p_message_id, 'verbatim', true, p_cooldown);
$$;

-- Lock-exempt protections --------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_pin(p_message_id bigint)
RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_apply(p_message_id, 'pinned', false, 0);
$$;

CREATE OR REPLACE FUNCTION stewards.context_unpin(p_message_id bigint)
RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_apply(p_message_id, 'verbatim', false, 0);
$$;

COMMENT ON FUNCTION stewards.context_compress(bigint, int) IS 'CT2.1 lever: fold a message to its engram (lockable toggle).';
COMMENT ON FUNCTION stewards.context_mute(bigint, int)     IS 'CT2.1 lever: tombstone a resolved sub-thread, recoverable (lockable toggle).';
COMMENT ON FUNCTION stewards.context_expand(bigint, int)   IS 'CT2.1 lever: pull a folded/muted message back to verbatim (lockable toggle).';
COMMENT ON FUNCTION stewards.context_pin(bigint)           IS 'CT2.1 lever: protect a message from automatic compaction (lock-exempt, voluntary).';
COMMENT ON FUNCTION stewards.context_unpin(bigint)         IS 'CT2.1 lever: release a pin (lock-exempt).';


-- ---------------------------------------------------------------------
-- 6. The pressure helper — what CT2.2 renders as the §5 pressure line.
-- ---------------------------------------------------------------------
-- Returns an estimate of window pressure + the foldable candidates with
-- their handles + approximate token weights. Token estimate is a cheap
-- chars/4 proxy (no tokenizer in SQL); the agent uses it as a relative
-- signal, not an exact count. "Foldable" = torso (older than the last 8),
-- currently verbatim, not locked, with non-trivial content.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_pressure(p_session_id text)
RETURNS jsonb LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_tail_size int := 8;   -- mirror compose_messages' tail
    v_total     int;
    v_est       bigint;
    v_foldable  jsonb;
BEGIN
    WITH ordered AS (
        SELECT m.id, m.role, m.content, m.context_state, m.locked_until_turn,
               ROW_NUMBER() OVER (ORDER BY m.created_at DESC, m.id DESC) AS rn_from_end,
               CEIL(length(m.content) / 4.0)::bigint AS est_tokens
          FROM stewards.messages m
         WHERE m.session_id = p_session_id
    )
    SELECT count(*)::int,
           COALESCE(sum(est_tokens), 0),
           COALESCE(jsonb_agg(
               jsonb_build_object(
                   'handle',     stewards.context_handle(id),
                   'message_id', id,
                   'role',       role,
                   'est_tokens', est_tokens
               ) ORDER BY est_tokens DESC
           ) FILTER (
               WHERE rn_from_end > v_tail_size
                 AND context_state = 'verbatim'
                 AND role IN ('tool', 'assistant')
                 AND length(content) > 200
                 AND (locked_until_turn IS NULL
                      OR stewards.session_turn(p_session_id) >= locked_until_turn)
           ), '[]'::jsonb)
      INTO v_total, v_est, v_foldable
      FROM ordered;

    RETURN jsonb_build_object(
        'session_id',    p_session_id,
        'current_turn',  stewards.session_turn(p_session_id),
        'message_count', v_total,
        'est_tokens',    v_est,
        'foldable',      v_foldable
    );
END;
$FN$;

COMMENT ON FUNCTION stewards.context_pressure(text) IS
'CT2.1: window-pressure estimate + foldable candidates (handle/id/role/est_tokens). chars/4 token proxy. CT2.2 formats the §5 CONTEXT PRESSURE line from this; the agent uses it to decide when/what to fold.';


-- =====================================================================
-- End of ct2-1-context-state-model.sql
-- =====================================================================
