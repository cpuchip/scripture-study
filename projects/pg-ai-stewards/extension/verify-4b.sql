-- =====================================================================
-- Phase 4b smoke test — verify dispatch override + non-regression
-- =====================================================================

\echo '=== A. Provider rename: model_pricing rows now use opencode_go ==='
SELECT provider, count(*) FROM stewards.model_pricing GROUP BY provider ORDER BY provider;

\echo ''
\echo '=== B. Provider rename: cost_buckets rows now use opencode_go ==='
SELECT provider, count(*) FROM stewards.cost_buckets GROUP BY provider ORDER BY provider;

\echo ''
\echo '=== C. compute_cost now finds Kimi pricing under opencode_go ==='
\echo '    Expected: $2.95 = 2950000 micro_dollars (was 0 before this push)'
SELECT * FROM stewards.compute_cost('opencode_go','kimi-k2.6', 1000000, 500000);

\echo ''
\echo '=== D. work_items.provider_override column exists ==='
SELECT column_name, data_type, is_nullable
  FROM information_schema.columns
 WHERE table_schema='stewards' AND table_name='work_items'
   AND column_name='provider_override';

\echo ''
\echo '=== E. work_item_dispatch_stage signature now has 3 params ==='
SELECT proname, pronargs, pg_get_function_arguments(oid)
  FROM pg_proc
 WHERE pronamespace=(SELECT oid FROM pg_namespace WHERE nspname='stewards')
   AND proname='work_item_dispatch_stage';

\echo ''
\echo '=== F. NON-REGRESSION: existing pipelines still defined and dispatchable ==='
SELECT family, jsonb_array_length(stages) AS stage_count, stages->0->>'agent_family' AS first_agent
  FROM stewards.pipelines
 ORDER BY family;

\echo ''
\echo '=== G. NON-REGRESSION: synthetic dispatch test on a pending work_item ==='
\echo '    Setup: create work_item, dispatch via existing 2-arg signature (omit p_allow_failed_status)'
\echo '    Expected: returns work_id, work_queue gets a chat row with provider opencode_go'
DO $reg$
DECLARE
    v_wi_id uuid;
    v_work_id bigint;
    v_provider text;
    v_status text;
BEGIN
    -- Pick the study-write pipeline (3-stage, exists in seeds)
    v_wi_id := stewards.work_item_create(
        p_pipeline_family := 'study-write',
        p_input := jsonb_build_object('binding_question', 'smoke test 4b'),
        p_slug := 'smoke-4b-' || (extract(epoch from now())::int)::text,
        p_actor := 'verify-4b'
    );

    RAISE NOTICE 'created work_item %', v_wi_id;

    -- Dispatch using the legacy 2-arg form. p_allow_failed_status defaults
    -- to false (the safe gate); status='pending' so it dispatches normally.
    v_work_id := stewards.work_item_dispatch_stage(v_wi_id);

    RAISE NOTICE 'dispatched work_id %', v_work_id;

    -- Confirm provider on the resulting work_queue row
    SELECT provider INTO v_provider FROM stewards.work_queue WHERE id = v_work_id;
    RAISE NOTICE 'work_queue.provider = %', v_provider;
    IF v_provider != 'opencode_go' THEN
        RAISE EXCEPTION 'REGRESSION: expected opencode_go, got %', v_provider;
    END IF;

    -- Confirm work_item moved to in_progress
    SELECT status INTO v_status FROM stewards.work_items WHERE id = v_wi_id;
    RAISE NOTICE 'work_item.status after dispatch = %', v_status;
    IF v_status != 'in_progress' THEN
        RAISE EXCEPTION 'REGRESSION: expected in_progress, got %', v_status;
    END IF;

    -- Cleanup: cancel the work_item so it doesn't sit in flight
    PERFORM stewards.work_item_cancel(v_wi_id, 'smoke test cleanup');
    -- Don't try to delete the work_queue row (FK cascades from cost_events
    -- -- not present here); just leave it. work_item_cancel marks the item.
END;
$reg$;

\echo ''
\echo '=== H. Override path: synthetic dispatch with model_override + provider_override ==='
\echo '    Expected: work_queue row uses the override provider/model, not pipeline default'
DO $ovr$
DECLARE
    v_wi_id uuid;
    v_work_id bigint;
    v_provider text;
    v_payload jsonb;
BEGIN
    v_wi_id := stewards.work_item_create(
        p_pipeline_family := 'study-write',
        p_input := jsonb_build_object('binding_question', 'override test 4b'),
        p_slug := 'smoke-4b-ovr-' || (extract(epoch from now())::int)::text,
        p_actor := 'verify-4b'
    );

    -- Set overrides BEFORE dispatch
    UPDATE stewards.work_items
       SET model_override = 'qwen3.6-plus',
           provider_override = 'opencode_go'  -- same provider but explicit
     WHERE id = v_wi_id;

    v_work_id := stewards.work_item_dispatch_stage(v_wi_id);

    SELECT provider, payload INTO v_provider, v_payload
      FROM stewards.work_queue WHERE id = v_work_id;

    RAISE NOTICE 'override dispatch: provider=% requested_model=%',
        v_provider, v_payload->>'requested_model';

    IF v_payload->>'requested_model' != 'qwen3.6-plus' THEN
        RAISE EXCEPTION 'override path failed: expected qwen3.6-plus, got %',
            v_payload->>'requested_model';
    END IF;

    PERFORM stewards.work_item_cancel(v_wi_id, 'override test cleanup');
END;
$ovr$;

\echo ''
\echo '=== I. Status gate: dispatch on failed work_item without flag raises ==='
\echo '    Expected: ERROR'
DO $gate$
DECLARE
    v_wi_id uuid;
    v_work_id bigint;
BEGIN
    v_wi_id := stewards.work_item_create(
        p_pipeline_family := 'study-write',
        p_input := jsonb_build_object('binding_question', 'gate test 4b'),
        p_slug := 'smoke-4b-gate-' || (extract(epoch from now())::int)::text,
        p_actor := 'verify-4b'
    );

    -- Force into failed state
    UPDATE stewards.work_items SET status = 'failed' WHERE id = v_wi_id;

    BEGIN
        v_work_id := stewards.work_item_dispatch_stage(v_wi_id);
        RAISE EXCEPTION 'GATE BROKEN: dispatch from failed status returned %', v_work_id;
    EXCEPTION
        WHEN OTHERS THEN
            -- Only check that it raised; reraise if it's not the expected message
            IF SQLERRM NOT LIKE '%cannot dispatch from status failed%' THEN
                RAISE;
            END IF;
            RAISE NOTICE 'gate raised correctly: %', SQLERRM;
    END;

    -- Now retry with p_allow_failed_status=true; should succeed
    v_work_id := stewards.work_item_dispatch_stage(v_wi_id, NULL, true);
    RAISE NOTICE 'allow_failed retry succeeded: work_id=%', v_work_id;

    PERFORM stewards.work_item_cancel(v_wi_id, 'gate test cleanup');
END;
$gate$;

\echo ''
\echo '=== Phase 4b smoke test complete ==='
