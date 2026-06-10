-- =====================================================================
-- CT2.5 — sub-agent id as a natural auto-tag (the RUN 2 addressing fix)
-- =====================================================================
-- RUN 2 finding (2026-06-10): a scaffolded persona DID reach for the lever
-- but addressed it by the SUB-AGENT ID it saw in the tool-result header —
-- context_mute(handle:"subagent-20260610-023003-067") — its natural
-- reference. context_resolve_handle only knew [ctx:xxxx] 4-hex handles
-- (it grabbed "2026" from that string, matched nothing → errored). Michael:
-- "accept an id for a sub-agent call as a natural auto tag."
--
-- Two small pieces, additive + backward-compatible:
--   1. Auto-tag a sub-agent tool result with its slug (from the
--      "[spawn_subagent <slug> complete …]" header that finalize() writes)
--      into the §7.4 context_tags[] column.
--   2. context_resolve_handle falls back to a context_tags match — so EVERY
--      lever (mute/pin/expand/compress/unpin all share this one resolver)
--      resolves the sub-agent id → its bulky digest message. No new vocab,
--      no [ctx:]-on-live-messages rendering: the lever now speaks the
--      model's existing language.
--
-- Pure SQL, live-appliable. The auto-tag trigger is on the messages insert
-- hot path but cheap (regex only for role='tool' with content) and purely
-- additive (appends a tag; never raises, never blocks an insert).
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Auto-tag trigger: stamp a sub-agent tool result with its slug.
--    BEFORE INSERT so we can set NEW.context_tags. Coexists with the §7.4
--    stamp_working_tag trigger — both only APPEND, so firing order is moot.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.tag_subagent_result() RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_slug text;
BEGIN
    IF NEW.role = 'tool' AND NEW.content IS NOT NULL THEN
        -- The header finalize() writes: "[spawn_subagent <slug> complete in …]".
        -- All sub-agent wrappers (research_codebase, deep_research, the L.6
        -- set) route through spawn_subagent, so this catches every one.
        v_slug := substring(NEW.content FROM '\[spawn_subagent (\S+) complete');
        IF v_slug IS NOT NULL AND v_slug <> '' AND NOT (NEW.context_tags @> ARRAY[v_slug]) THEN
            NEW.context_tags := NEW.context_tags || v_slug;
        END IF;
    END IF;
    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS messages_tag_subagent_result ON stewards.messages;
CREATE TRIGGER messages_tag_subagent_result
BEFORE INSERT ON stewards.messages
FOR EACH ROW EXECUTE FUNCTION stewards.tag_subagent_result();

COMMENT ON FUNCTION stewards.tag_subagent_result() IS
'CT2.5: auto-tag a sub-agent tool result with its slug (from the spawn_subagent digest header) into context_tags[], so a persona can mute/pin it by the id it naturally saw.';

-- ---------------------------------------------------------------------
-- 2. context_resolve_handle: try the [ctx:] 4-hex handle first (unchanged),
--    then fall back to a context_tags match (the sub-agent id). Most
--    recent message wins on ties.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_resolve_handle(p_session_id text, p_handle text)
RETURNS bigint LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_h  text;
    v_id bigint;
BEGIN
    IF p_session_id IS NULL OR p_handle IS NULL THEN RETURN NULL; END IF;

    -- (a) the original [ctx:xxxx] 4-hex handle scheme.
    v_h := lower(substring(p_handle FROM '([0-9a-fA-F]{4})'));
    IF v_h IS NOT NULL THEN
        SELECT m.id INTO v_id
          FROM stewards.messages m
         WHERE m.session_id = p_session_id
           AND stewards.context_handle(m.id) = v_h
         ORDER BY m.id DESC
         LIMIT 1;
        IF v_id IS NOT NULL THEN RETURN v_id; END IF;
    END IF;

    -- (b) CT2.5 fallback: a context_tags match — e.g. a sub-agent id the
    --     model saw in a tool-result header (its natural reference).
    SELECT m.id INTO v_id
      FROM stewards.messages m
     WHERE m.session_id = p_session_id
       AND m.context_tags @> ARRAY[btrim(p_handle)]
     ORDER BY m.id DESC
     LIMIT 1;
    RETURN v_id;  -- NULL if neither scheme matched
END;
$FN$;

COMMENT ON FUNCTION stewards.context_resolve_handle(text, text) IS
'CT2.3 + CT2.5: resolve a context reference to a message_id within one session — a [ctx:xxxx] 4-hex handle first, then a context_tags match (e.g. a sub-agent id the model saw in a result header).';

-- =====================================================================
-- Acceptance (CT2.5):
--   1. Insert a role='tool' message containing "[spawn_subagent foo-123
--      complete …]" → its context_tags includes 'foo-123'.
--   2. context_resolve_handle(session, 'foo-123') → that message's id.
--   3. context_resolve_handle still resolves a real [ctx:xxxx] handle.
--   4. RUN 3: codewright-ct2 mutes by the sub-agent id → the mute LANDS
--      (no error) and the bulky digest leaves the rendered context.
-- =====================================================================
