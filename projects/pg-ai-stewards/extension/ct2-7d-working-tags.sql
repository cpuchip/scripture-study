-- =====================================================================
-- CT2.7d — Working tags (§7.4): batch a whole task's context by one tag
-- =====================================================================
-- The unit of work is a TASK: an agent grinds across many turns, then moves
-- on and wants to sweep ALL of it out of context at once. A working tag makes
-- a task's footprint a single addressable thing.
--
-- context_set_tag(tag) → from then on every new message in the session is
-- auto-stamped with `tag` (no per-message tagging) until context_clear_tag().
-- Batch levers context_fold_tag/mute_tag/expand_tag/pin_tag operate on every
-- message bearing the tag in ONE call = ONE circuit-breaker event (the whole
-- set locks together, so a deliberate sweep isn't penalized as thrash).
--
-- Distinct from §7.6 self-note `tags` (durable-note labels). These tag
-- CURRENT-WINDOW messages for batch fold/mute. Pure SQL; context_* tools are
-- already gated by compose_tools (the context_% prefix).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Columns: per-message tags + the session's active working tag.
-- ---------------------------------------------------------------------
ALTER TABLE stewards.messages  ADD COLUMN IF NOT EXISTS context_tags text[] NOT NULL DEFAULT '{}';
ALTER TABLE stewards.sessions  ADD COLUMN IF NOT EXISTS working_tag  text;
CREATE INDEX IF NOT EXISTS messages_context_tags_idx ON stewards.messages USING gin (context_tags);

COMMENT ON COLUMN stewards.messages.context_tags IS 'CT2 §7.4: working tags stamped on this message (for batch context_*_tag ops).';
COMMENT ON COLUMN stewards.sessions.working_tag IS 'CT2 §7.4: the session''s active working tag — new messages are auto-stamped with it until cleared.';


-- ---------------------------------------------------------------------
-- 2. Auto-stamp trigger: stamp new messages with the session's working tag.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.stamp_working_tag() RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE v_tag text;
BEGIN
    SELECT working_tag INTO v_tag FROM stewards.sessions WHERE id = NEW.session_id;
    IF v_tag IS NOT NULL AND v_tag <> '' AND NOT (NEW.context_tags @> ARRAY[v_tag]) THEN
        NEW.context_tags := array_append(COALESCE(NEW.context_tags, '{}'), v_tag);
    END IF;
    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS messages_stamp_working_tag ON stewards.messages;
CREATE TRIGGER messages_stamp_working_tag
    BEFORE INSERT ON stewards.messages
    FOR EACH ROW EXECUTE FUNCTION stewards.stamp_working_tag();


-- ---------------------------------------------------------------------
-- 3. Batch applicator + the set/clear + four batch levers.
-- ---------------------------------------------------------------------
-- One circuit-breaker event: the whole tagged set gets the new state and a
-- single shared cooldown (not per-message), so a deliberate task-sweep is not
-- penalized as thrash (§7.4).
CREATE OR REPLACE FUNCTION stewards._context_tag_apply(
    p_session text, p_tag text, p_state text, p_lockable boolean, p_cooldown int DEFAULT 3
) RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE v_turn int; v_n int; v_lock int;
BEGIN
    IF p_tag IS NULL OR btrim(p_tag) = '' THEN
        RETURN jsonb_build_object('error', 'tag required');
    END IF;
    v_turn := stewards.session_turn(p_session);
    v_lock := CASE WHEN p_lockable THEN v_turn + GREATEST(p_cooldown, 0) ELSE NULL END;
    UPDATE stewards.messages
       SET context_state     = p_state,
           locked_until_turn = CASE WHEN p_lockable THEN v_turn + GREATEST(p_cooldown,0) ELSE locked_until_turn END
     WHERE session_id = p_session
       AND context_tags @> ARRAY[p_tag];
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n = 0 THEN
        RETURN jsonb_build_object('ok', true, 'tag', p_tag, 'state', p_state, 'messages', 0,
            'note', 'no messages bear that tag yet');
    END IF;
    RETURN jsonb_build_object('ok', true, 'tag', p_tag, 'state', p_state, 'messages', v_n, 'locked_until_turn', v_lock);
END;
$FN$;

CREATE OR REPLACE FUNCTION stewards.context_set_tag_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE v_sess text := p_args->>'_session_id'; v_tag text := btrim(COALESCE(p_args->>'tag',''));
BEGIN
    IF v_tag = '' THEN RETURN jsonb_build_object('error','tag required'); END IF;
    UPDATE stewards.sessions SET working_tag = v_tag WHERE id = v_sess;
    IF NOT FOUND THEN RETURN jsonb_build_object('error','unknown session'); END IF;
    RETURN jsonb_build_object('ok', true, 'working_tag', v_tag,
        'note', 'new messages will be tagged "'||v_tag||'" until you set another tag or clear it');
END;
$FN$;

CREATE OR REPLACE FUNCTION stewards.context_clear_tag_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE v_sess text := p_args->>'_session_id';
BEGIN
    UPDATE stewards.sessions SET working_tag = NULL WHERE id = v_sess;
    RETURN jsonb_build_object('ok', true, 'working_tag', null);
END;
$FN$;

CREATE OR REPLACE FUNCTION stewards.context_fold_tag_tool(p_args jsonb)   RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_tag_apply(p_args->>'_session_id', p_args->>'tag', 'compressed', true, COALESCE(NULLIF(p_args->>'cooldown','')::int,3)); $$;
CREATE OR REPLACE FUNCTION stewards.context_mute_tag_tool(p_args jsonb)   RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_tag_apply(p_args->>'_session_id', p_args->>'tag', 'muted', true, COALESCE(NULLIF(p_args->>'cooldown','')::int,3)); $$;
CREATE OR REPLACE FUNCTION stewards.context_expand_tag_tool(p_args jsonb) RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_tag_apply(p_args->>'_session_id', p_args->>'tag', 'verbatim', true, COALESCE(NULLIF(p_args->>'cooldown','')::int,3)); $$;
CREATE OR REPLACE FUNCTION stewards.context_pin_tag_tool(p_args jsonb)    RETURNS jsonb LANGUAGE sql AS $$
    SELECT stewards._context_tag_apply(p_args->>'_session_id', p_args->>'tag', 'pinned', false, 0); $$;


-- ---------------------------------------------------------------------
-- 4. tool_defs (sql_fn). context_* prefix → already gated by compose_tools.
-- ---------------------------------------------------------------------
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('context_set_tag',
 'Start tagging your work: from now on every new message in this turn-thread is stamped with this tag, so you can later fold/mute the WHOLE task at once with context_*_tag. Set it once at the start of a sub-task; call context_clear_tag or set a new tag when you move on.',
 '{"type":"object","required":["tag"],"additionalProperties":false,"properties":{"tag":{"type":"string","description":"A short task label, e.g. todo-3 or auth-refactor."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_set_tag_tool','schema','stewards'), true),
('context_clear_tag',
 'Stop auto-tagging new messages (untagged work resumes). Does not change already-tagged messages.',
 '{"type":"object","additionalProperties":false,"properties":{}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_clear_tag_tool','schema','stewards'), true),
('context_fold_tag',
 'Compress EVERY message bearing a tag to its engram, in one move — reclaim a finished task''s tokens. One circuit-breaker event (the whole set locks together). Recover with context_expand_tag.',
 '{"type":"object","required":["tag"],"additionalProperties":false,"properties":{"tag":{"type":"string"},"cooldown":{"type":"integer"}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_fold_tag_tool','schema','stewards'), true),
('context_mute_tag',
 'Tombstone EVERY message bearing a tag (a resolved task you are done with), recoverable. One circuit-breaker event.',
 '{"type":"object","required":["tag"],"additionalProperties":false,"properties":{"tag":{"type":"string"},"cooldown":{"type":"integer"}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_mute_tag_tool','schema','stewards'), true),
('context_expand_tag',
 'Bring EVERY message bearing a tag back to full verbatim (a task reopened). One circuit-breaker event.',
 '{"type":"object","required":["tag"],"additionalProperties":false,"properties":{"tag":{"type":"string"},"cooldown":{"type":"integer"}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_expand_tag_tool','schema','stewards'), true),
('context_pin_tag',
 'Protect EVERY message bearing a tag from automatic compaction (e.g. the spec + acceptance criteria for the task in flight). Lock-exempt.',
 '{"type":"object","required":["tag"],"additionalProperties":false,"properties":{"tag":{"type":"string"}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','context_pin_tag_tool','schema','stewards'), true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description, args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target, active = true;


-- ---------------------------------------------------------------------
-- 5. Pressure line echoes the active working tag (§7.4).
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_pressure_line(p_session_id text)
RETURNS text LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v jsonb; v_est bigint; v_fold jsonb; v_n int; v_list text; v_line text; v_tag text;
BEGIN
    v      := stewards.context_pressure(p_session_id);
    v_est  := COALESCE((v ->> 'est_tokens')::bigint, 0);
    v_fold := COALESCE(v -> 'foldable', '[]'::jsonb);
    v_n    := jsonb_array_length(v_fold);

    v_line := 'CONTEXT PRESSURE: ~' || to_char(v_est, 'FM999,999,999,999') || ' tokens in this window.';
    SELECT working_tag INTO v_tag FROM stewards.sessions WHERE id = p_session_id;
    IF v_tag IS NOT NULL AND v_tag <> '' THEN
        v_line := v_line || E'\nWorking tag: ' || v_tag || ' (new messages are tagged; context_fold_tag/mute_tag to sweep it).';
    END IF;
    IF v_n > 0 THEN
        SELECT string_agg('[ctx:' || (f ->> 'handle') || '] ' || to_char((f ->> 'est_tokens')::bigint, 'FM999,999,999,999') || 't', '  ·  ')
          INTO v_list
          FROM (SELECT f FROM jsonb_array_elements(v_fold) f LIMIT 6) x;
        v_line := v_line || E'\nFoldable now: ' || v_list;
        v_line := v_line ||
            E'\n(Fold the least-relevant with context_compress/context_mute; context_pin protects a message; context_expand restores it. A toggle locks that message for a few turns.)';
    END IF;
    RETURN v_line;
END;
$FN$;


-- =====================================================================
-- End of ct2-7d-working-tags.sql
-- =====================================================================
