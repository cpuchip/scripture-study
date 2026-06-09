-- =====================================================================
-- R11 — auto-verify subagent-* one-shot pipelines (P1.5 adjacent fix)
-- =====================================================================
-- Found during the research_codebase Go-handler e2e (2026-06-09): the
-- one-shot auto-verify trigger covers aggregate-children / brainstorm-% /
-- redline% / persona-% but NOT subagent-%. A subagent-research-codebase
-- child therefore finishes status=completed but stalls at maturity=raw,
-- and spawn_subagent's verified-wait hangs to its 20-min ceiling — the
-- exact j6/R.6 failure mode r7 warned about. The L.6 wrapper pipelines
-- (subagent-url-summary, subagent-file-audit, …) share the same family
-- prefix and the same exposure.
--
-- Based on the LIVE function definition (l13 lesson: never rebuild a
-- multiply-evolved function from an old migration file), adding ONE arm.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.on_one_shot_pipeline_completed()
RETURNS trigger LANGUAGE plpgsql AS $function$
DECLARE
    v_qualifies boolean;
BEGIN
    v_qualifies := NEW.pipeline_family = 'aggregate-children'
                OR NEW.pipeline_family LIKE 'brainstorm-%'
                OR NEW.pipeline_family LIKE 'redline%'
                OR NEW.pipeline_family LIKE 'persona-%'    -- R.8: any persona-* pipeline
                OR NEW.pipeline_family LIKE 'subagent-%';  -- R11: L.6 wrappers + research_codebase
    IF NOT v_qualifies THEN
        RETURN NEW;
    END IF;
    IF NEW.maturity = 'verified' THEN
        RETURN NEW;
    END IF;
    UPDATE stewards.work_items SET maturity = 'verified', updated_at = now() WHERE id = NEW.id;
    RAISE NOTICE 'on_one_shot_pipeline_completed: auto-verified % (pipeline=%)', NEW.id, NEW.pipeline_family;
    RETURN NEW;
END;
$function$;

COMMENT ON FUNCTION stewards.on_one_shot_pipeline_completed() IS
'J.4 + R.6 + R.7/R.8 + R11: auto-verify one-shot pipelines (aggregate-children + brainstorm-* + redline* + persona-* + subagent-*) when their single stage completes. Cascades into on_maturity_verified.';

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
        OR NEW.pipeline_family LIKE 'persona-%'
        OR NEW.pipeline_family LIKE 'subagent-%'
    )
)
EXECUTE FUNCTION stewards.on_one_shot_pipeline_completed();

-- =====================================================================
-- Acceptance (R11):
--   1. pg_get_triggerdef shows the subagent-% arm in the WHEN clause.
--   2. A spawned subagent-research-codebase child reaches
--      status=completed AND maturity=verified without manual help, so
--      spawn_subagent's poll terminates inside seconds, not 20 min.
-- =====================================================================
