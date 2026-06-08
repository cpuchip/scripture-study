-- =====================================================================
-- CT2.7a — Durable self-notes: the store + facet engine + renderer
-- =====================================================================
-- spec §7 (RATIFIED 2026-06-08, §7.6). The faceted-audience durable memory.
-- This gate is the READ path + building blocks (no agent tools yet, that's
-- 7b). Pure SQL, live-applyable, inert until notes exist.
--
-- A note carries `audience` selectors (dimension→value). Every dispatch has
-- `facets`. A note renders iff facets @> audience (the dispatch's facets
-- contain all the note's selector pairs) — ONE match rule, no scope tiers.
-- Facets available from the substrate alone: global / session / agent_family
-- / kind / pipeline. persona + room are threaded from persona-host in 7c.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. The store.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.agent_self_notes (
    id              bigserial PRIMARY KEY,
    note            text NOT NULL,
    audience        jsonb NOT NULL DEFAULT '{}'::jsonb,   -- selectors; {} matches nothing
    tags            text[] NOT NULL DEFAULT '{}',         -- free-form labels (search only)
    created_by      text,                                 -- agent_family / persona that wrote it
    created_session text,                                 -- the session that wrote it
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS agent_self_notes_audience_idx   ON stewards.agent_self_notes USING gin (audience);
CREATE INDEX IF NOT EXISTS agent_self_notes_created_by_idx ON stewards.agent_self_notes (created_by);
CREATE INDEX IF NOT EXISTS agent_self_notes_tags_idx       ON stewards.agent_self_notes USING gin (tags);

COMMENT ON TABLE stewards.agent_self_notes IS
'CT2 §7: durable self-notes (the Hermes loop). audience = faceted selectors matched against dispatch_facets via @>. tags = free-form labels (search only, do not gate delivery). Human-prunable; the model add/removes via remember/forget (7b).';


-- ---------------------------------------------------------------------
-- 2. kind — a coarse persona/agent class, drives the `kind` facet.
-- ---------------------------------------------------------------------
ALTER TABLE stewards.agents ADD COLUMN IF NOT EXISTS kind text;
COMMENT ON COLUMN stewards.agents.kind IS
'CT2 §7: coarse agent class (roleplay/code/librarian/general/…) for the `kind` audience facet. A {kind:code} note reaches every code-kind agent (the shared per-kind pool). NULL = no kind facet.';

-- Sensible initial kinds (NULL-guarded so a human re-class sticks).
UPDATE stewards.agents SET kind = 'roleplay'  WHERE family = 'persona'                     AND kind IS NULL;
UPDATE stewards.agents SET kind = 'librarian' WHERE family = 'librarian'                   AND kind IS NULL;
UPDATE stewards.agents SET kind = 'code'      WHERE family IN ('dev','debug','subagent-research-codebase') AND kind IS NULL;


-- ---------------------------------------------------------------------
-- 3. Note handle ([note:xxxx]) — distinct namespace from message handles.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.context_note_handle(p_note_id bigint)
RETURNS text LANGUAGE sql IMMUTABLE AS $$
    SELECT substr(md5('note:' || p_note_id::text), 1, 4);
$$;


-- ---------------------------------------------------------------------
-- 4. dispatch_facets — what THIS dispatch is, for audience matching.
-- ---------------------------------------------------------------------
-- Substrate-recoverable facets only (7c adds persona + room from
-- persona-host). jsonb_strip_nulls drops kind/pipeline when absent so a
-- note that selects on them simply won't match.
CREATE OR REPLACE FUNCTION stewards.dispatch_facets(p_agent_family text, p_session_id text)
RETURNS jsonb LANGUAGE sql STABLE AS $$
    SELECT jsonb_strip_nulls(jsonb_build_object(
        'global',       true,
        'session',      p_session_id,
        'agent_family', p_agent_family,
        'kind',         (SELECT a.kind FROM stewards.agents a
                          WHERE a.family = p_agent_family AND a.kind IS NOT NULL LIMIT 1),
        'pipeline',     (SELECT w.pipeline_family FROM stewards.work_items w
                          WHERE p_session_id = ANY(w.session_ids) ORDER BY w.id DESC LIMIT 1)
    ));
$$;

COMMENT ON FUNCTION stewards.dispatch_facets(text, text) IS
'CT2 §7: the facets of the current dispatch (global/session/agent_family/kind/pipeline; persona+room added in 7c). A self-note renders iff dispatch_facets @> note.audience.';


-- ---------------------------------------------------------------------
-- 5. render_self_notes — the "YOUR DURABLE NOTES" block (or '').
-- ---------------------------------------------------------------------
-- Returns '' when no notes match this dispatch (so the system prompt is
-- byte-identical — the §6 safety property). Caps: ~40 notes / ~4,000
-- tokens (≈16,000 chars), most-recent first.
CREATE OR REPLACE FUNCTION stewards.render_self_notes(p_agent_family text, p_session_id text)
RETURNS text LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_facets jsonb := stewards.dispatch_facets(p_agent_family, p_session_id);
    v_block  text  := '';
    v_count  int   := 0;
    v_chars  int   := 0;
    r        record;
BEGIN
    FOR r IN
        SELECT n.id, n.note
          FROM stewards.agent_self_notes n
         WHERE n.audience <> '{}'::jsonb       -- empty audience matches nothing
           AND v_facets @> n.audience          -- the one match rule
         ORDER BY n.created_at DESC, n.id DESC
    LOOP
        EXIT WHEN v_count >= 40 OR v_chars >= 16000;   -- ~40 notes / ~4k tokens
        v_block := v_block || '- [note:' || stewards.context_note_handle(r.id) || '] ' || r.note || E'\n';
        v_count := v_count + 1;
        v_chars := v_chars + length(r.note);
    END LOOP;

    IF v_count = 0 THEN
        RETURN '';
    END IF;
    RETURN E'\n\n## YOUR DURABLE NOTES\n'
        || E'(things you chose to remember; forget(handle) to drop one once integrated)\n'
        || v_block;
END;
$FN$;

COMMENT ON FUNCTION stewards.render_self_notes(text, text) IS
'CT2 §7: renders the durable-notes block for a dispatch (audience-matched, capped ~40/~4k tok). Empty string when nothing matches so the system prompt stays backward-compatible. Wired into compose_messages in gate 7a2.';


-- =====================================================================
-- End of ct2-7a-self-notes-store.sql
-- =====================================================================
