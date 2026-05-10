-- =====================================================================
-- Phase 4c — Steward dispatch wiring (Push B of
-- steward-bgworker-integration.md)
--
-- Builds on:
--   - 4a-steward.sql (defines steward_actions, diagnose_failure,
--     retry_guidance, pipeline_breakers, the original steward_tick)
--   - 4b-dispatch-override.sql (adds work_item_dispatch_stage's
--     p_allow_failed_status param + provider_override column)
--
-- This file ONLY replaces stewards.steward_tick(). The previous
-- version (in 4a-steward.sql) wrote `retry_with_escalation` action
-- rows but did not actually dispatch. The new version:
--
--   1. Sets work_items.model_override = picked_model
--   2. Sets work_items.provider_override = 'opencode_go' (the OpenCode
--      Go chain is single-provider; cross-provider escalation is
--      explicitly out of scope per D-EC3)
--   3. Increments failure_count (the steward owns this — bridge just
--      sets status='failed' on failure; steward counts retries)
--   4. Calls stewards.work_item_dispatch_stage(work_item.id,
--      retry_guidance, true) — true unlocks status='failed' re-dispatch
--   5. Records the dispatched work_queue id in steward_actions.details
--
-- Per-item EXCEPTION WHEN OTHERS so a dispatch failure on one item
-- doesn't poison the rest of the tick batch.
--
-- Note on failure_count semantics: this design has the steward own
-- failure_count (incremented on retry-trigger, NOT on dispatch-failure).
-- A successful dispatch leaves failure_count at the new value; the
-- next stage advance OR a successful response should reset it to 0.
-- Reset-on-success lives in a separate concern (work_item_advance
-- extension or bridge response handler) and is TODO for next push.
-- For now: failure_count grows monotonically until quarantine threshold.
-- =====================================================================

DROP FUNCTION IF EXISTS stewards.steward_tick();

CREATE OR REPLACE FUNCTION stewards.steward_tick()
RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_count               int := 0;
    v_item                record;
    v_diagnosis           text;
    v_next_model          text;
    v_breaker_ok          boolean;
    v_attempt             int;
    v_retry_text          text;
    v_dispatched_work_id  bigint;
BEGIN
    FOR v_item IN
        SELECT id, pipeline_family, current_stage, failure_count,
               last_failure_reason, escalation_state
          FROM stewards.work_items
         WHERE status = 'failed'
           AND failure_count < 3
           AND quarantined_at IS NULL
           AND escalation_state = 'normal'
         ORDER BY updated_at ASC  -- oldest failures first
         LIMIT 10
         FOR UPDATE SKIP LOCKED
    LOOP
        v_attempt := v_item.failure_count + 1;

        -- 1. Cost cap check
        IF stewards.cost_cap_exceeded(v_item.id) THEN
            UPDATE stewards.work_items
               SET quarantined_at = now(),
                   quarantine_reason = 'cost_cap_exceeded'
             WHERE id = v_item.id;

            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action, details)
            VALUES
                (v_item.id,
                 'cumulative cost exceeded cap; quarantining',
                 'cost_limit',
                 'quarantine',
                 jsonb_build_object('quarantine_reason','cost_cap_exceeded'));
            v_count := v_count + 1;
            CONTINUE;
        END IF;

        -- 2. Diagnose
        v_diagnosis := stewards.diagnose_failure(
            v_item.last_failure_reason, v_item.failure_count);
        UPDATE stewards.work_items
           SET last_failure_diagnosis = v_diagnosis
         WHERE id = v_item.id;

        -- 3. Breaker check
        v_breaker_ok := stewards.breaker_check(
            v_item.pipeline_family, v_item.current_stage);
        IF NOT v_breaker_ok THEN
            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action)
            VALUES
                (v_item.id,
                 format('breaker open for %s/%s; deferring',
                        v_item.pipeline_family, v_item.current_stage),
                 v_diagnosis,
                 'defer_breaker_open');
            v_count := v_count + 1;
            CONTINUE;
        END IF;

        -- 4. Pick model
        v_next_model := stewards.pick_model(
            v_item.pipeline_family, v_item.current_stage,
            v_attempt, v_diagnosis);

        -- 5. Queue sentinel handling — transition to escalation_queued
        IF v_next_model = '__queue_for_opus__' THEN
            UPDATE stewards.work_items
               SET escalation_state = 'queued',
                   escalation_attempts = escalation_attempts + 1
             WHERE id = v_item.id;

            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action, model_used,
                 details)
            VALUES
                (v_item.id,
                 'OpenCode chain exhausted; queued for human-mediated Opus boost',
                 v_diagnosis,
                 'queue_for_opus',
                 '__queue_for_opus__',
                 jsonb_build_object(
                     'attempt', v_attempt,
                     'escalation_attempts',
                         (SELECT escalation_attempts FROM stewards.work_items
                           WHERE id = v_item.id)));
            v_count := v_count + 1;
            CONTINUE;
        END IF;

        -- 6. Retry path: set overrides + dispatch + log
        v_retry_text := stewards.retry_guidance(v_diagnosis, v_attempt);

        -- Set overrides FIRST so dispatch picks them up.
        -- failure_count incremented here; dispatch will fire even though
        -- count grew, because work_item_dispatch_stage doesn't read it.
        UPDATE stewards.work_items
           SET model_override     = v_next_model,
               provider_override  = 'opencode_go',
               failure_count      = failure_count + 1
         WHERE id = v_item.id;

        -- Dispatch wrapped so one bad item doesn't poison the batch.
        BEGIN
            v_dispatched_work_id := stewards.work_item_dispatch_stage(
                v_item.id, v_retry_text, true);

            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action, model_used,
                 details)
            VALUES
                (v_item.id,
                 format('attempt #%s after %s; dispatched as work_id %s',
                        v_attempt, v_diagnosis, v_dispatched_work_id),
                 v_diagnosis,
                 'retry_dispatched',
                 v_next_model,
                 jsonb_build_object(
                     'attempt', v_attempt,
                     'retry_guidance', v_retry_text,
                     'dispatched_work_id', v_dispatched_work_id,
                     'provider_override', 'opencode_go'));
        EXCEPTION WHEN OTHERS THEN
            -- Dispatch failed. Log the error; the per-item update to
            -- failure_count + overrides was already committed via the
            -- enclosing function-transaction. Next tick will see this
            -- item again with the bumped failure_count (so quarantine
            -- threshold still applies).
            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action, model_used,
                 details)
            VALUES
                (v_item.id,
                 'dispatch failed during retry: ' || SQLERRM,
                 v_diagnosis,
                 'dispatch_error',
                 v_next_model,
                 jsonb_build_object(
                     'attempt', v_attempt,
                     'sqlerrm', SQLERRM,
                     'sqlstate', SQLSTATE));
        END;

        v_count := v_count + 1;
    END LOOP;

    RETURN v_count;
END;
$func$;

COMMENT ON FUNCTION stewards.steward_tick() IS
'Phase 4c: Watch→Diagnose→Act→Account orchestration with actual dispatch. Walks failed work_items, applies cost-cap + breaker + diagnosis + escalation logic, sets model+provider overrides, calls work_item_dispatch_stage(allow_failed=true), logs to steward_actions. Returns count of actions taken. Per-item EXCEPTION isolation. Called by bgworker on tick (Push C).';

-- =====================================================================
-- Done. Phase 4c steward dispatch wiring is operational.
--
-- Acceptance:
--   1. Synthetic failed work_item with last_failure_reason='timeout'
--      and failure_count=0 → SELECT steward_tick() → returns 1, work_item
--      now has failure_count=1, model_override set to stage default
--      (Kimi for study), status=in_progress, new work_queue row exists.
--   2. After 3 retries fail (failure_count reaches 3), next steward_tick
--      doesn't pick the work_item up (filter is failure_count < 3).
--   3. Synthetic 5 failures on (study, research) trips breaker; next
--      steward_tick on a study/research item logs 'defer_breaker_open'
--      action without dispatching.
--   4. Synthetic failure on a work_item with cost_micro_dollars >=
--      cost_cap_micro → quarantined, no dispatch.
--   5. Force pick_model to return __queue_for_opus__ (e.g., set
--      failure_count=2 + diagnosis=model_limit on a GLM stage) →
--      escalation_state transitions to 'queued', no dispatch.
-- =====================================================================
