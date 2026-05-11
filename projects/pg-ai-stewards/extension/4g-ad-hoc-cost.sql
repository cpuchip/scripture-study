-- =====================================================================
-- Phase 4g — Ad-hoc chat cost tracking + cache token plumbing
--
-- Builds on:
--   - 4a-cost-tracking.sql (model_pricing, cost_events, cost_buckets,
--     compute_cost, record_cost_event)
--
-- Two changes:
--
-- 1. **Relax cost_events.work_item_id to nullable** so chats not tied
--    to a work_item (watchman passes, ad-hoc explorations) can still
--    be cost-tracked. Add session_id column to identify the owner of
--    standalone chats. The existing trigger already no-ops on the
--    work_items UPDATE when work_item_id is NULL (NULL != NULL in
--    WHERE), so trigger logic doesn't change.
--
-- 2. **Update record_cost_event signature** to accept p_session_id
--    (paired with the optional work_item_id) and to use cache token
--    parameters that are now plumbed through from bgworker.rs (Phase
--    4h). Existing callers that pass the old 9-arg form will fail
--    with the signature change — bgworker.rs is updated in lockstep.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: Schema relaxations
-- ---------------------------------------------------------------------

ALTER TABLE stewards.cost_events
    ALTER COLUMN work_item_id DROP NOT NULL;

ALTER TABLE stewards.cost_events
    ADD COLUMN IF NOT EXISTS session_id text;

CREATE INDEX IF NOT EXISTS cost_events_session ON stewards.cost_events(session_id);

COMMENT ON COLUMN stewards.cost_events.work_item_id IS
'Phase 4g: nullable. NULL means an ad-hoc chat not tied to a work_item (e.g., watchman pass). Use session_id for owner tracking.';
COMMENT ON COLUMN stewards.cost_events.session_id IS
'Phase 4g: stewards.sessions.id of the chat that produced this event. Always set; for work-item-tied chats it derives from the work_item dispatch (wi--<uuid>--<stage>).';

-- ---------------------------------------------------------------------
-- Section 2: Replace record_cost_event with new 10-arg signature
--
-- New param: p_session_id at position 9 (between cache_read_tokens
-- and notes). Default NULL so callers can omit if they don't have it.
-- ---------------------------------------------------------------------

DROP FUNCTION IF EXISTS stewards.record_cost_event(uuid, int, text, text, int, int, int, int, text);

CREATE OR REPLACE FUNCTION stewards.record_cost_event(
    p_work_item_id       uuid,
    p_attempt_seq        int,
    p_provider           text,
    p_model              text,
    p_input_tokens       int,
    p_output_tokens      int,
    p_cache_write_tokens int  DEFAULT 0,
    p_cache_read_tokens  int  DEFAULT 0,
    p_session_id         text DEFAULT NULL,
    p_notes              text DEFAULT NULL
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_micro      bigint;
    v_pricing_at timestamptz;
    v_id         bigint;
    v_notes      text;
BEGIN
    SELECT micro_dollars, pricing_effective_at
      INTO v_micro, v_pricing_at
      FROM stewards.compute_cost(p_provider, p_model,
                                  p_input_tokens, p_output_tokens,
                                  p_cache_write_tokens, p_cache_read_tokens);

    -- If no pricing row exists, flag in notes so the gap is visible.
    v_notes := p_notes;
    IF v_pricing_at = '-infinity'::timestamptz THEN
        v_notes := coalesce(v_notes || ' | ', '')
                 || 'no_pricing_row(' || p_provider || '/' || p_model || ')';
    END IF;

    INSERT INTO stewards.cost_events
        (work_item_id, session_id, attempt_seq, provider, model,
         input_tokens, output_tokens, cache_write_tokens, cache_read_tokens,
         micro_dollars, pricing_effective_at, notes)
    VALUES
        (p_work_item_id, p_session_id, p_attempt_seq, p_provider, p_model,
         p_input_tokens, p_output_tokens, p_cache_write_tokens, p_cache_read_tokens,
         v_micro, v_pricing_at, v_notes)
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$func$;

COMMENT ON FUNCTION stewards.record_cost_event(uuid, int, text, text, int, int, int, int, text, text) IS
'Phase 4g: insert a cost_events row with computed micro_dollars. Trigger updates work_items (when work_item_id non-NULL) + buckets. Pass p_session_id for chats not tied to a work_item.';

-- =====================================================================
-- Done. Phase 4g ad-hoc cost tracking + cache plumbing is operational.
--
-- Acceptance:
--   1. SELECT stewards.record_cost_event(NULL, 1, 'opencode_go', 'kimi-k2.6', 1000, 100, 0, 0, 'test-session', 'ad-hoc test');
--      → bigint id; cost_events row exists with work_item_id=NULL, session_id='test-session'
--   2. work_items.cost_micro_dollars not affected (no work_item_id to match)
--   3. cost_buckets.opencode_go.daily incremented by computed micro_dollars
-- =====================================================================
