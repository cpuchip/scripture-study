-- =====================================================================
-- Batch J.4 follow-up — brainstorm-lens auto-verify on completion
-- =====================================================================
-- Same shape as aggregate-children: a single-stage pipeline that needs
-- to reach maturity=verified on stage completion so the parent's
-- aggregator-dispatch trigger can fire.
--
-- Generalizes the existing on_aggregate_completed trigger to handle all
-- "one-shot" pipelines whose pipeline metadata.shape ends up in our
-- single-stage-verify list. For simplicity in J.4 we hard-code the set:
--   aggregate-children, brainstorm-scamper, brainstorm-six-hats,
--   brainstorm-crazy8s, brainstorm-reverse
--
-- Future enhancement: drive this from pipelines.metadata->>'shape' or
-- a dedicated auto_verify_on_complete flag on the pipelines table.
-- Deferred — explicit list is fine for the current 5 pipelines.
--
-- Also runs a one-time UPDATE to flip any already-completed lens rows
-- from maturity=raw to verified (catches the J.5 reverse-lens that
-- finished before this migration applied).
-- =====================================================================

-- 1. Drop the narrow trigger; replace with a broader one.

DROP TRIGGER IF EXISTS work_items_on_aggregate_completed ON stewards.work_items;

CREATE OR REPLACE FUNCTION stewards.on_one_shot_pipeline_completed()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_qualifies boolean;
BEGIN
    v_qualifies := NEW.pipeline_family = 'aggregate-children'
                OR NEW.pipeline_family LIKE 'brainstorm-%';

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
'Batch J.4 follow-up: auto-verify one-shot pipelines (aggregate-children + brainstorm-*) when their single stage completes. Cascades into on_maturity_verified for auto-materialize and aggregator-dispatch.';

DROP TRIGGER IF EXISTS work_items_on_one_shot_completed ON stewards.work_items;

CREATE TRIGGER work_items_on_one_shot_completed
AFTER UPDATE OF status ON stewards.work_items
FOR EACH ROW
WHEN (
    NEW.status = 'completed'
    AND (
        NEW.pipeline_family = 'aggregate-children'
        OR NEW.pipeline_family LIKE 'brainstorm-%'
    )
)
EXECUTE FUNCTION stewards.on_one_shot_pipeline_completed();

-- 2. Drop the obsolete narrower function (the trigger is gone; nothing
--    references it).

DROP FUNCTION IF EXISTS stewards.on_aggregate_completed();

-- 3. Retroactive: flip any already-completed brainstorm-lens rows to
--    maturity=verified so the in-flight J.5 chain proceeds.

UPDATE stewards.work_items
   SET maturity = 'verified',
       updated_at = now()
 WHERE pipeline_family LIKE 'brainstorm-%'
   AND status = 'completed'
   AND maturity <> 'verified';

-- =====================================================================
-- End of j6-brainstorm-lens-auto-verify.sql
-- =====================================================================
