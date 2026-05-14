-- =====================================================================
-- Batch L.1.1.8 — Judge surface intercept (AFTER INSERT trigger)
-- =====================================================================
-- When a role='tool' message lands with content > intercept_threshold,
-- this trigger:
--   1. Captures the raw via chunk_and_index (populates overflow tables)
--   2. Synthesizes a top-level overview from the first parent chunk
--   3. Renders the judge prompt template with the corpus metadata
--   4. UPDATEs messages.content with the rendered surface
--
-- The raw is preserved in messages_raw_overflow parents (FK to message)
-- and the leaves are queued for contextualization + embedding via the
-- chain we built in L.1.1.5/L.1.1.6.
--
-- expand_message tier='raw' will stitch parents back together when
-- the agent needs the original. (Implemented separately.)
--
-- Trigger name prefixed 'aa_' so it fires alphabetically BEFORE the
-- K.1 engram extraction trigger — by the time K.1 evaluates its WHERE,
-- our trigger has replaced content with the small judge surface and
-- K.1's threshold check returns false naturally.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Trigger function: intercept_oversized_tool_after.
-- ---------------------------------------------------------------------
-- AFTER INSERT so the FK from messages_raw_overflow to messages.id
-- can satisfy. We UPDATE messages.content with the rendered surface
-- after indexing.

CREATE OR REPLACE FUNCTION stewards.intercept_oversized_tool_after()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_threshold     int;
    v_binding       text;
    v_first_parent  text;
    v_top_overview  text;
    v_surface       text;
    v_index_result  jsonb;
BEGIN
    v_threshold := stewards.intercept_threshold_chars(NEW.session_id);

    IF NEW.content LIKE '%[CORPUS-INDEXED]%' THEN RETURN NEW; END IF;
    IF NEW.role <> 'tool' THEN RETURN NEW; END IF;
    IF length(NEW.content) <= v_threshold THEN RETURN NEW; END IF;

    SELECT input ->> 'binding_question' INTO v_binding
      FROM stewards.work_items
     WHERE NEW.session_id = ANY(session_ids)
     LIMIT 1;

    BEGIN
        v_index_result := stewards.chunk_and_index(NEW.id, v_binding);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'intercept_oversized_tool_after: chunk_and_index failed for msg=%: %; leaving raw',
            NEW.id, SQLERRM;
        RETURN NEW;
    END;

    SELECT substring(content FROM 1 FOR 500) INTO v_first_parent
      FROM stewards.messages_raw_overflow
     WHERE message_id = NEW.id
     ORDER BY parent_ordinal ASC
     LIMIT 1;

    v_top_overview := COALESCE(v_first_parent, '(no parent content)') ||
        CASE WHEN length(v_first_parent) >= 500 THEN '…' ELSE '' END;

    v_surface := E'[CORPUS-INDEXED]\n\n' || stewards.render_judge_surface(NEW.id, v_top_overview);

    UPDATE stewards.messages SET content = v_surface WHERE id = NEW.id;

    RAISE NOTICE 'intercept_oversized_tool_after: msg=% indexed (% parents, % leaves) surface replaces % chars',
        NEW.id,
        v_index_result ->> 'parent_count',
        v_index_result ->> 'leaf_count',
        v_index_result ->> 'source_bytes';

    RETURN NEW;
END;
$FN$;

COMMENT ON FUNCTION stewards.intercept_oversized_tool_after() IS
'Batch L.1.1.8: AFTER INSERT trigger. Same logic as intercept_oversized_tool but uses AFTER INSERT + UPDATE so the FK from messages_raw_overflow to messages.id satisfies.';

-- Drop both possible prior trigger names; only register the AFTER variant.
DROP TRIGGER IF EXISTS messages_aa_intercept_oversized_tool ON stewards.messages;
DROP TRIGGER IF EXISTS messages_aa_intercept_oversized ON stewards.messages;

CREATE TRIGGER messages_aa_intercept_oversized
AFTER INSERT ON stewards.messages
FOR EACH ROW
WHEN (NEW.role = 'tool' AND length(NEW.content) > 50000)
EXECUTE FUNCTION stewards.intercept_oversized_tool_after();


-- ---------------------------------------------------------------------
-- 3. Update expand_message to stitch overflow parents for tier='raw'.
-- ---------------------------------------------------------------------
-- Carry-forward: this is a partial — expand_message lives in K.3. For
-- now, agents can read raw via a direct query on messages_raw_overflow
-- ORDER BY parent_ordinal. Full integration with expand_message tier=
-- 'raw' deferred to a follow-up.

CREATE OR REPLACE FUNCTION stewards.read_overflow_raw(
    p_message_id bigint,
    p_max_chars int DEFAULT 50000
) RETURNS text LANGUAGE sql STABLE AS $$
    SELECT string_agg(content, E'\n\n--- chunk boundary ---\n\n' ORDER BY parent_ordinal)
      FROM (
        SELECT content, parent_ordinal,
               sum(length(content) + 30) OVER (ORDER BY parent_ordinal) AS running_size
          FROM stewards.messages_raw_overflow
         WHERE message_id = p_message_id
      ) sub
     WHERE running_size <= p_max_chars
$$;

COMMENT ON FUNCTION stewards.read_overflow_raw(bigint, int) IS
'Batch L.1.1.8: stitch overflow parents back into a single text stream, capped at p_max_chars. Used by agents (or the K.3 expand_message tier=raw path) when they need original content after the judge surface replaced messages.content.';


-- ---------------------------------------------------------------------
-- 4. Tell K.1 engram extraction to skip [CORPUS-INDEXED] surfaces.
-- ---------------------------------------------------------------------
-- Without this, K.1 fires on the AFTER INSERT (seeing NEW.content as
-- the original raw) and queues an extraction whose dispatch reads
-- messages.content as the already-replaced surface — wasted DeepSeek
-- call. Idempotency at the trigger level.

CREATE OR REPLACE FUNCTION stewards.trigger_extract_engrams_on_large_tool()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_threshold int;
BEGIN
    -- L.1.1.8: skip if already indexed into a judge corpus.
    IF NEW.content LIKE '%[CORPUS-INDEXED]%' THEN
        RETURN NEW;
    END IF;
    v_threshold := stewards.effective_extraction_threshold(NEW.session_id);
    IF length(NEW.content) <= v_threshold THEN
        RETURN NEW;
    END IF;
    BEGIN
        PERFORM stewards.extract_engrams(NEW.id);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'trigger_extract_engrams_on_large_tool: enqueue failed for msg=%: %',
            NEW.id, SQLERRM;
    END;
    RETURN NEW;
END;
$FN$;


-- =====================================================================
-- End of l23-judge-surface-intercept.sql
-- =====================================================================
