-- =====================================================================
-- ES.1.s1 — work_item_cancel cascade (hard stop)
-- =====================================================================
-- CF-1: the existing work_item_cancel only flips work_items.status. The
-- chat→tool_dispatch→chat loop runs on session_id and never checks the
-- owning work_item's status, so a cancelled work_item keeps spending.
--
-- Fix: work_item_cancel now ALSO marks every non-terminal work_queue
-- row for the work_item's session_ids as error. Hard stop (ratified):
-- pending, in_progress, AND waiting_for_tools rows are all killed so
-- tool_dispatch_complete_waiting cannot resurrect the loop.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.work_item_cancel(
    p_work_item_id uuid,
    p_reason text DEFAULT NULL
) RETURNS void LANGUAGE plpgsql AS $FN$
DECLARE
    v_sessions text[];
    v_killed   int;
BEGIN
    UPDATE stewards.work_items
       SET status       = 'cancelled',
           error        = coalesce(p_reason, error),
           updated_at   = now(),
           completed_at = now()
     WHERE id = p_work_item_id
       AND status NOT IN ('completed', 'cancelled')
    RETURNING session_ids INTO v_sessions;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'work_item_cancel: % not found or already in terminal status',
            p_work_item_id;
    END IF;

    -- ES.1.s1 cascade: hard-stop every non-terminal work_queue row tied
    -- to this work_item's sessions. waiting_for_tools is included so
    -- tool_dispatch_complete_waiting won't fire a continuation chat.
    IF v_sessions IS NOT NULL AND array_length(v_sessions, 1) > 0 THEN
        WITH killed AS (
            UPDATE stewards.work_queue
               SET status = 'error'
             WHERE status IN ('pending', 'in_progress', 'waiting_for_tools')
               AND payload->>'session_id' = ANY(v_sessions)
            RETURNING 1
        )
        SELECT count(*) INTO v_killed FROM killed;

        RAISE NOTICE 'work_item_cancel: % cancelled; cascade killed % non-terminal work_queue row(s) across % session(s)',
            p_work_item_id, v_killed, array_length(v_sessions, 1);
    END IF;
END;
$FN$;

COMMENT ON FUNCTION stewards.work_item_cancel(uuid, text) IS
'ES.1.s1: cancel a work_item AND hard-stop its session chat loops. Marks every pending/in_progress/waiting_for_tools work_queue row for the work_item''s session_ids as error so the chat→tool_dispatch→chat loop cannot keep spending after cancellation.';

-- =====================================================================
-- End of es1-work-item-cancel-cascade.sql
-- =====================================================================
