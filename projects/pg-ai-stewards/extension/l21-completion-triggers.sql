-- =====================================================================
-- Batch L.1.1 — Completion triggers for new markers
-- =====================================================================
-- Wires apply_contextualize_leaf (L.1.1.5) and apply_map_reduce_parent_
-- engrams (L.1.1.9) as AFTER UPDATE triggers on work_queue, matching
-- the K.1 apply_engram_extraction pattern. No bgworker.rs change needed.
-- =====================================================================


-- L.1.1.5 — contextualize_leaf completion.
DROP TRIGGER IF EXISTS work_queue_apply_contextualize_leaf ON stewards.work_queue;

CREATE OR REPLACE FUNCTION stewards.trigger_apply_contextualize_leaf()
RETURNS trigger LANGUAGE plpgsql AS $FN$
BEGIN
    BEGIN
        PERFORM stewards.apply_contextualize_leaf(NEW.id);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'trigger_apply_contextualize_leaf: wq=% failed: %', NEW.id, SQLERRM;
    END;
    RETURN NEW;
END;
$FN$;

CREATE TRIGGER work_queue_apply_contextualize_leaf
AFTER UPDATE OF status ON stewards.work_queue
FOR EACH ROW
WHEN (
    NEW.kind = 'chat'
    AND NEW.status IN ('done', 'error')
    AND OLD.status IS DISTINCT FROM NEW.status
    AND NEW.payload ? '_contextualize_leaf_id'
)
EXECUTE FUNCTION stewards.trigger_apply_contextualize_leaf();


-- L.1.1.9 — map_reduce_extract_engrams completion.
DROP TRIGGER IF EXISTS work_queue_apply_map_reduce_engrams ON stewards.work_queue;

CREATE OR REPLACE FUNCTION stewards.trigger_apply_map_reduce_engrams()
RETURNS trigger LANGUAGE plpgsql AS $FN$
BEGIN
    BEGIN
        PERFORM stewards.apply_map_reduce_parent_engrams(NEW.id);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'trigger_apply_map_reduce_engrams: wq=% failed: %', NEW.id, SQLERRM;
    END;
    RETURN NEW;
END;
$FN$;

CREATE TRIGGER work_queue_apply_map_reduce_engrams
AFTER UPDATE OF status ON stewards.work_queue
FOR EACH ROW
WHEN (
    NEW.kind = 'chat'
    AND NEW.status IN ('done', 'error')
    AND OLD.status IS DISTINCT FROM NEW.status
    AND NEW.payload ? '_map_reduce_extract_parent_id'
)
EXECUTE FUNCTION stewards.trigger_apply_map_reduce_engrams();


-- =====================================================================
-- End of l21-completion-triggers.sql
-- =====================================================================
