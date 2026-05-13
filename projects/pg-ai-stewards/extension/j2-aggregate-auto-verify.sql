-- =====================================================================
-- Batch J.2 follow-up — aggregate-children auto-verify on completion
-- =====================================================================
-- The aggregate-children pipeline is a single-stage "one shot" — after
-- the aggregate stage finishes, the work_item should be considered
-- verified so on_maturity_verified fires and auto-materializes the
-- index file (auto_materialize_on_verified=true on the pipeline).
--
-- Without this trigger, aggregator work_items stall at status=completed
-- maturity=raw, and the index file is stranded in stage_results.
--
-- Surfaced by smoke-j2-fanout end-to-end on 2026-05-13.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.on_aggregate_completed()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    -- Only fires for aggregate-children pipeline; trigger WHEN clause
    -- already filters, but defense-in-depth check.
    IF NEW.pipeline_family <> 'aggregate-children' THEN
        RETURN NEW;
    END IF;

    IF NEW.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    UPDATE stewards.work_items
       SET maturity = 'verified',
           updated_at = now()
     WHERE id = NEW.id;

    RAISE NOTICE 'on_aggregate_completed: auto-verified aggregator %', NEW.id;

    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION stewards.on_aggregate_completed() IS
'Batch J.2 follow-up: auto-verify aggregate-children work_items when their single stage completes. Cascades into on_maturity_verified which auto-materializes the index file.';

DROP TRIGGER IF EXISTS work_items_on_aggregate_completed ON stewards.work_items;

CREATE TRIGGER work_items_on_aggregate_completed
AFTER UPDATE OF status ON stewards.work_items
FOR EACH ROW
WHEN (NEW.status = 'completed' AND NEW.pipeline_family = 'aggregate-children')
EXECUTE FUNCTION stewards.on_aggregate_completed();

-- =====================================================================
-- End of j2-aggregate-auto-verify.sql
-- =====================================================================
