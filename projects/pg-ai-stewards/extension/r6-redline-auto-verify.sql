-- =====================================================================
-- Batch R.6 — redline one-shot auto-verify
-- =====================================================================
-- The redline + redline-condense pipelines are single-stage one-shots, same
-- shape as brainstorm lenses and aggregate-children. j6 auto-verifies those on
-- status=completed so the parent's aggregator-dispatch trigger can fire — but
-- its qualifying set didn't include redline. Result (caught in the R.6 live
-- smoke): redline children finish status=completed but stay maturity=raw, the
-- index aggregator waits forever, and the children never read as "done."
--
-- This extends on_one_shot_pipeline_completed (carrying j6 forward verbatim) to
-- also qualify pipeline_family LIKE 'redline%', and retroactively verifies any
-- already-completed redline rows (the R.6 smoke children) so their aggregator
-- fires. This is exactly the j6 pattern; the j6 file itself flagged "drive from
-- a metadata flag" as the future generalization — still deferred.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.on_one_shot_pipeline_completed()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_qualifies boolean;
BEGIN
    v_qualifies := NEW.pipeline_family = 'aggregate-children'
                OR NEW.pipeline_family LIKE 'brainstorm-%'
                OR NEW.pipeline_family LIKE 'redline%';   -- R.6: redline + redline-condense

    IF NOT v_qualifies THEN
        RETURN NEW;
    END IF;

    IF NEW.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    UPDATE stewards.work_items
       SET maturity = 'verified',
           updated_at = now()
     WHERE id = NEW.id;

    RAISE NOTICE 'on_one_shot_pipeline_completed: auto-verified % (pipeline=%)',
        NEW.id, NEW.pipeline_family;
    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION stewards.on_one_shot_pipeline_completed() IS
'Batch J.4 follow-up + R.6: auto-verify one-shot pipelines (aggregate-children + brainstorm-* + redline*) when their single stage completes. Cascades into on_maturity_verified for auto-materialize and aggregator-dispatch.';

DROP TRIGGER IF EXISTS work_items_on_one_shot_completed ON stewards.work_items;

CREATE TRIGGER work_items_on_one_shot_completed
AFTER UPDATE OF status ON stewards.work_items
FOR EACH ROW
WHEN (
    NEW.status = 'completed'
    AND (
        NEW.pipeline_family = 'aggregate-children'
        OR NEW.pipeline_family LIKE 'brainstorm-%'
        OR NEW.pipeline_family LIKE 'redline%'
    )
)
EXECUTE FUNCTION stewards.on_one_shot_pipeline_completed();

-- Retroactive: flip already-completed redline rows to verified so their
-- aggregator dispatches (mirrors j6's one-time catch-up for the J.5 lens).
UPDATE stewards.work_items
   SET maturity = 'verified',
       updated_at = now()
 WHERE pipeline_family LIKE 'redline%'
   AND status = 'completed'
   AND maturity <> 'verified';

-- =====================================================================
-- Acceptance (R.6): after apply, the redline-covenant-smoke children show
-- maturity=verified, and the index aggregator (aggregate-children sibling)
-- transitions out of 'pending' (dispatched, then completes + materializes the
-- index file). A fresh panel run auto-verifies each child on completion.
-- =====================================================================
