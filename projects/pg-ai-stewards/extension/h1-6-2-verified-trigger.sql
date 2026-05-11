-- =====================================================================
-- Batch H.1.6.2 + H.1.6.3 + H.1.6.4 — AFTER UPDATE OF maturity trigger
--
-- Three ratified decisions land as one migration:
--   D-H6.2: sabbath_dispatch moves to a single trigger; removed from
--           apply_gate_decision + apply_verify_result.
--   D-H6.3: auto-materialize is opt-in per pipeline + per-work_item
--           override (mirrors D-H5 sabbath/atonement pattern).
--   D-H6.4: both sabbath and materialize fire on the same trigger
--           (AFTER UPDATE OF maturity → verified).
-- =====================================================================

-- ---------------------------------------------------------------------
-- D-H6.3: schema additions
-- ---------------------------------------------------------------------

ALTER TABLE stewards.pipelines
    ADD COLUMN IF NOT EXISTS auto_materialize_on_verified boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN stewards.pipelines.auto_materialize_on_verified IS
'D-H6.3 (Batch H): when true, enqueue_work_item_file fires automatically on maturity→verified for work_items with file_destination set. Default false preserves Batch G''s "explicit gesture" design. Flip per pipeline once trustworthy.';

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS auto_materialize_enabled boolean NULL;

COMMENT ON COLUMN stewards.work_items.auto_materialize_enabled IS
'D-H6.3 (Batch H): per-work_item override for pipeline.auto_materialize_on_verified. NULL = inherit; true = force on; false = skip auto-mat for this work_item.';

-- ---------------------------------------------------------------------
-- D-H6.2 + D-H6.3 + D-H6.4: trigger function on maturity→verified
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.on_maturity_verified()
RETURNS trigger
LANGUAGE plpgsql AS $func$
DECLARE
    v_pipeline      stewards.pipelines%ROWTYPE;
    v_sabbath       boolean;
    v_auto_mat      boolean;
    v_pwid          bigint;
    v_dispatch_id   bigint;
BEGIN
    -- Only act on transition TO 'verified'. NULL/non-verified previous
    -- values both fire if the new value is 'verified' and not already.
    IF NEW.maturity <> 'verified' OR OLD.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = NEW.pipeline_family;
    IF v_pipeline.family IS NULL THEN
        RAISE NOTICE 'on_maturity_verified: pipeline % not found', NEW.pipeline_family;
        RETURN NEW;
    END IF;

    -- D-H6.2: sabbath_dispatch on transition to verified
    v_sabbath := COALESCE(NEW.sabbath_enabled, v_pipeline.sabbath_enabled);
    IF v_sabbath AND NEW.sabbath_completed_at IS NULL THEN
        BEGIN
            v_dispatch_id := stewards.sabbath_dispatch(NEW.id);
            RAISE NOTICE 'on_maturity_verified: sabbath_dispatch work_id=% for work_item=%',
                v_dispatch_id, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: sabbath_dispatch failed: %', SQLERRM;
        END;
    END IF;

    -- D-H6.3 + D-H6.4: auto-materialize on transition to verified
    -- Requires file_destination set + auto_materialize resolved true
    -- + not already materialized.
    v_auto_mat := COALESCE(NEW.auto_materialize_enabled, v_pipeline.auto_materialize_on_verified);
    IF v_auto_mat
       AND NEW.file_destination IS NOT NULL
       AND NEW.materialized_at IS NULL
    THEN
        BEGIN
            v_pwid := stewards.enqueue_work_item_file(NEW.id, 'auto_materialize_on_verified');
            RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file pwid=% for work_item=%',
                v_pwid, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file failed: %', SQLERRM;
        END;
    END IF;

    RETURN NEW;
END;
$func$;

COMMENT ON FUNCTION stewards.on_maturity_verified() IS
'H.1.6.2/3/4 trigger function: when work_item.maturity transitions TO verified, fire sabbath_dispatch (if sabbath_enabled resolved) AND enqueue_work_item_file (if file_destination set + auto_materialize resolved). Both wrapped in BEGIN/EXCEPTION → NOTICE so the UPDATE always succeeds. Errors surface but don''t block the parent transaction.';

DROP TRIGGER IF EXISTS work_items_on_maturity_verified ON stewards.work_items;
CREATE TRIGGER work_items_on_maturity_verified
    AFTER UPDATE OF maturity ON stewards.work_items
    FOR EACH ROW
    EXECUTE FUNCTION stewards.on_maturity_verified();

-- ---------------------------------------------------------------------
-- D-H6.2: remove sabbath_dispatch from apply_gate_decision
-- The maturity UPDATE on this function transitions to 'verified', so
-- the new trigger fires sabbath. Single source of truth.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_gate_decision(
    p_work_item_id uuid,
    p_decision jsonb,
    p_work_id bigint DEFAULT NULL::bigint
)
RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi             stewards.work_items%ROWTYPE;
    v_action         text;
    v_reasoning      text;
    v_feedback       text;
    v_new_maturity   text;
    v_produces_mat   text;
    v_maturity_order text[] := ARRAY['raw','researched','planned','specced','executing','verified'];
    v_idx            int;
    v_new_revision   int;
    v_actor          jsonb;
    v_trust_level    text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    v_action    := p_decision->>'action';
    v_reasoning := p_decision->>'reasoning';
    v_feedback  := p_decision->>'feedback';

    IF v_action NOT IN ('advance', 'revise', 'surface') THEN
        RAISE EXCEPTION 'apply_gate_decision: invalid action %', v_action;
    END IF;

    INSERT INTO stewards.gate_decisions
        (work_item_id, from_maturity, action, reasoning, feedback,
         work_id, revision_count, raw_response)
    VALUES
        (p_work_item_id, v_wi.maturity, v_action, v_reasoning, v_feedback,
         p_work_id, v_wi.revision_count, p_decision);

    v_new_maturity := v_wi.maturity;

    IF v_action = 'advance' THEN
        v_actor := stewards.work_item_stage_actor(p_work_item_id);
        IF v_actor IS NOT NULL THEN
            SELECT trust_level INTO v_trust_level
              FROM stewards.trust_scores
             WHERE agent_family    = v_actor->>'agent_family'
               AND pipeline_family = v_actor->>'pipeline_family'
               AND model           = v_actor->>'model';

            IF v_trust_level IS NULL OR v_trust_level = 'trainee' THEN
                UPDATE stewards.work_items
                   SET status = 'awaiting_review',
                       updated_at = now()
                 WHERE id = p_work_item_id;
                RETURN v_wi.maturity;
            END IF;
        END IF;

        SELECT produces_maturity INTO v_produces_mat
          FROM stewards.pipeline_stage_maturity
         WHERE pipeline_family = v_wi.pipeline_family
           AND stage_name = v_wi.current_stage;

        IF v_produces_mat IS NOT NULL THEN
            v_new_maturity := v_produces_mat;
        ELSE
            v_idx := array_position(v_maturity_order, v_wi.maturity);
            IF v_idx IS NOT NULL AND v_idx < array_length(v_maturity_order, 1) THEN
                v_new_maturity := v_maturity_order[v_idx + 1];
            END IF;
        END IF;

        UPDATE stewards.work_items
           SET maturity       = v_new_maturity,
               revision_count = 0,
               updated_at     = now()
         WHERE id = p_work_item_id;

        IF v_new_maturity = 'verified' AND v_actor IS NOT NULL THEN
            BEGIN
                PERFORM stewards.trust_record_success(
                    v_actor->>'agent_family',
                    v_actor->>'pipeline_family',
                    v_actor->>'model'
                );
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'trust_record_success raised: %', SQLERRM;
            END;
        END IF;

        -- H.1.6.2: sabbath_dispatch removed — now fires via the
        -- work_items_on_maturity_verified trigger when maturity
        -- transitions to verified (regardless of which path drove
        -- the transition).

    ELSIF v_action = 'revise' THEN
        v_new_revision := v_wi.revision_count + 1;

        IF v_new_revision > 2 THEN
            UPDATE stewards.work_items
               SET status = 'awaiting_review',
                   revision_count = v_new_revision,
                   updated_at = now()
             WHERE id = p_work_item_id;
        ELSE
            UPDATE stewards.work_items
               SET status                 = 'failed',
                   revision_count         = v_new_revision,
                   last_failure_reason    = 'gate revise: ' || coalesce(v_feedback, '(no feedback)'),
                   last_failure_diagnosis = 'gate_revise',
                   updated_at             = now()
             WHERE id = p_work_item_id;
        END IF;

    ELSIF v_action = 'surface' THEN
        UPDATE stewards.work_items
           SET status     = 'awaiting_review',
               updated_at = now()
         WHERE id = p_work_item_id;
    END IF;

    RETURN v_new_maturity;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_gate_decision(uuid, jsonb, bigint) IS
'H.1.6.2 refactor (Batch H): sabbath_dispatch removed; now fires via the work_items_on_maturity_verified trigger on transition to verified.';

-- ---------------------------------------------------------------------
-- D-H6.2: remove sabbath_dispatch from apply_verify_result too.
-- This function fires sabbath when all_passed=true but does NOT advance
-- maturity itself. After this refactor, verify success no longer fires
-- sabbath directly — sabbath waits for apply_gate_decision(advance) to
-- transition maturity to verified, then the trigger fires.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_verify_result(
    p_work_item_id uuid,
    p_result jsonb,
    p_work_id bigint DEFAULT NULL::bigint
)
RETURNS boolean
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_all_passed  boolean;
    v_results     jsonb;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'apply_verify_result: work_item % not found', p_work_item_id;
    END IF;

    v_all_passed := coalesce((p_result->>'all_passed')::boolean, false);
    v_results    := coalesce(p_result->'results', '[]'::jsonb);

    INSERT INTO stewards.verify_results
        (work_item_id, all_passed, results, work_id, raw_response)
    VALUES
        (p_work_item_id, v_all_passed, v_results, p_work_id, p_result);

    IF NOT v_all_passed THEN
        UPDATE stewards.work_items
           SET maturity               = 'planned',
               status                 = 'failed',
               last_failure_reason    = 'verify failed: see verify_results',
               last_failure_diagnosis = 'verify_failed',
               updated_at             = now()
         WHERE id = p_work_item_id;
    END IF;
    -- H.1.6.2: success path no longer fires sabbath_dispatch directly.
    -- Sabbath fires from the work_items_on_maturity_verified trigger
    -- when apply_gate_decision (or any other path) transitions maturity
    -- to verified. apply_verify_result no longer needs the pipeline
    -- lookup; behavior change is intentional and documented in the
    -- substrate-batch-h-pipeline-expansion proposal.

    RETURN v_all_passed;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_verify_result(uuid, jsonb, bigint) IS
'H.1.6.2 refactor (Batch H): sabbath_dispatch removed; now fires via the work_items_on_maturity_verified trigger when a subsequent apply_gate_decision (or any path) transitions maturity to verified.';
