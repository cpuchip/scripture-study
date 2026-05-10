-- =====================================================================
-- Phase 4d — Steward realignment + per-item exception isolation
--
-- Two fixes caught during 4c smoke testing:
--
-- 1. **stage_models name mismatch.** 4a-escalation-chain seeded
--    pipeline_family='study'/'lesson'/'dev' but the substrate's actual
--    pipelines are 'study-write', 'study-write-qwen', 'echo-test'.
--    pick_model raised "no stage_models row for study-write/outline"
--    when steward_tick tried to retry a real failed work_item. Add
--    rows for the actual pipelines.
--
--    NOTE: 'study-write-qwen' uses provider='lm_studio' and
--    model='qwen/qwen3.6-27b' (a local LM Studio test variant, not
--    on the OpenCode chain). Adding it here would tell pick_model to
--    keep retrying on that local model with no escalation possible
--    (no model_escalation rules exist for qwen/qwen3.6-27b). For now
--    we DO add a row for study-write-qwen so steward_tick processes
--    it without raising — pick_model returns the same model on every
--    attempt; if it fails 3 times, normal quarantine-threshold logic
--    applies. The steward respects provider boundaries via the
--    provider_override it sets.
--
-- 2. **steward_tick per-item exception isolation insufficient.** The
--    4c version wrapped only the dispatch call in EXCEPTION WHEN
--    OTHERS. But pick_model can raise BEFORE dispatch (e.g., the
--    pipeline-name mismatch above). When that happens mid-batch,
--    the entire function transaction rolls back, losing all prior
--    actions. Wrap the whole per-item body in BEGIN/EXCEPTION so
--    one bad item logs to steward_actions and the loop continues.
--
-- Also: provider_override is now derived from model_pricing per
-- model (instead of hardcoded 'opencode_go') so cross-provider
-- chains work correctly. The model's pricing row carries its
-- provider; that's the canonical mapping.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: Add stage_models for actual pipelines
-- ---------------------------------------------------------------------

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    -- study-write (the production OpenCode chain)
    ('study-write',      'outline',  'kimi-k2.6',      'matches existing pipeline default'),
    ('study-write',      'draft',    'kimi-k2.6',      'matches existing pipeline default'),
    ('study-write',      'review',   'kimi-k2.6',      'matches existing pipeline default'),

    -- echo-test (1-stage smoke test)
    ('echo-test',        'echo',     'kimi-k2.6',      'matches existing pipeline default'),

    -- study-write-qwen (local LM Studio test variant; no escalation chain
    -- defined for qwen/qwen3.6-27b so pick_model returns same model on
    -- every attempt; quarantine after 3 failures still applies)
    ('study-write-qwen', 'outline',  'qwen/qwen3.6-27b', 'LM Studio variant; no OpenCode escalation'),
    ('study-write-qwen', 'draft',    'qwen/qwen3.6-27b', 'LM Studio variant; no OpenCode escalation'),
    ('study-write-qwen', 'review',   'qwen/qwen3.6-27b', 'LM Studio variant; no OpenCode escalation')

ON CONFLICT (pipeline_family, stage_name) DO UPDATE
SET default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

-- ---------------------------------------------------------------------
-- Section 2: Replace steward_tick with full per-item exception isolation
-- ---------------------------------------------------------------------

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
    v_provider            text;
BEGIN
    FOR v_item IN
        SELECT id, pipeline_family, current_stage, failure_count,
               last_failure_reason, escalation_state
          FROM stewards.work_items
         WHERE status = 'failed'
           AND failure_count < 3
           AND quarantined_at IS NULL
           AND escalation_state = 'normal'
         ORDER BY updated_at ASC
         LIMIT 10
         FOR UPDATE SKIP LOCKED
    LOOP
        -- Per-item exception isolation. Any error inside this block
        -- logs to steward_actions and the outer loop continues. This
        -- prevents one bad item (e.g., missing stage_models, broken
        -- dispatch) from poisoning the entire tick batch.
        BEGIN
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

            -- 4. Pick model (may raise if no stage_models row exists;
            -- caught by outer per-item EXCEPTION below)
            v_next_model := stewards.pick_model(
                v_item.pipeline_family, v_item.current_stage,
                v_attempt, v_diagnosis);

            -- 5. Queue sentinel handling
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

            -- 6. Resolve provider from model_pricing (NOT hardcoded
            -- 'opencode_go'). Each model knows its provider; that's
            -- the canonical mapping. Falls back to 'opencode_go' if
            -- somehow not in pricing (shouldn't happen given FK
            -- discipline but defensive).
            SELECT provider INTO v_provider
              FROM stewards.model_pricing
             WHERE model = v_next_model
             ORDER BY effective_at DESC
             LIMIT 1;
            v_provider := COALESCE(v_provider, 'opencode_go');

            -- 7. Retry path: set overrides + dispatch + log
            v_retry_text := stewards.retry_guidance(v_diagnosis, v_attempt);

            UPDATE stewards.work_items
               SET model_override     = v_next_model,
                   provider_override  = v_provider,
                   failure_count      = failure_count + 1
             WHERE id = v_item.id;

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
                     'provider_override', v_provider));

            v_count := v_count + 1;
        EXCEPTION WHEN OTHERS THEN
            -- Per-item failure isolation. Log and move on.
            -- Important: the BEGIN block opened a sub-transaction
            -- (PL/pgSQL semantics for BEGIN/EXCEPTION); when this
            -- handler fires, that sub-transaction rolls back, undoing
            -- any partial work for this item (e.g., the work_items
            -- UPDATE that set overrides + bumped failure_count).
            -- The action log row below is in a fresh sub-transaction
            -- so it commits.
            BEGIN
                INSERT INTO stewards.steward_actions
                    (work_item_id, observation, diagnosis, action, details)
                VALUES
                    (v_item.id,
                     'tick error: ' || SQLERRM,
                     COALESCE(v_diagnosis, 'unknown'),
                     'tick_error',
                     jsonb_build_object(
                         'sqlerrm', SQLERRM,
                         'sqlstate', SQLSTATE,
                         'pipeline_family', v_item.pipeline_family,
                         'current_stage', v_item.current_stage));
            EXCEPTION WHEN OTHERS THEN
                -- If even logging fails, swallow so the loop continues.
                NULL;
            END;
            v_count := v_count + 1;
        END;
    END LOOP;

    RETURN v_count;
END;
$func$;

COMMENT ON FUNCTION stewards.steward_tick() IS
'Phase 4d: Watch→Diagnose→Act→Account orchestration with full per-item exception isolation. Walks failed work_items, applies cost-cap + breaker + diagnosis + escalation logic, sets model+provider overrides (provider derived from model_pricing), calls work_item_dispatch_stage(allow_failed=true), logs to steward_actions. One item failure logs ''tick_error'' action and continues; never poisons the batch.';
