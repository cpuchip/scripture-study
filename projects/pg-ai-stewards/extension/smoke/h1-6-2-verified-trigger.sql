-- Smoke H.1.6.2/3/4: maturity → verified trigger fires sabbath +
-- auto-materialize. Bridge should be paused before running.

DO $$
DECLARE
    v_wi          uuid;
    v_intent_id   uuid;
    v_count_before bigint;
    v_count_after  bigint;
    v_pwid_count_before bigint;
    v_pwid_count_after  bigint;
BEGIN
    SELECT id INTO v_intent_id FROM stewards.intents WHERE slug = 'general-research';

    -- Case 1: auto_materialize ON (work_item override = true on pipeline default false)
    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"H.1.6.2 trigger smoke — auto_mat ON"}'::jsonb,
        'h162-smoke-mat-on', 'human', NULL, v_intent_id
    ) INTO v_wi;
    UPDATE stewards.work_items
       SET file_destination = 'research/h162-smoke.md',
           auto_materialize_enabled = true,
           sabbath_enabled = false  -- skip sabbath for this synthetic test
     WHERE id = v_wi;

    -- Count pending_file_writes before
    SELECT count(*) INTO v_pwid_count_before
      FROM stewards.pending_file_writes WHERE source_id = v_wi::text;

    -- Synthetically transition to verified (simulates what work_item_advance does on review completion)
    -- Need to populate stage_results.review.output so enqueue_work_item_file has content
    UPDATE stewards.work_items
       SET stage_results = jsonb_build_object(
               'gather',     jsonb_build_object('output', 'synth gather output'),
               'synthesize', jsonb_build_object('output', 'synth synthesize output for materialization'),
               'review',     jsonb_build_object('output', 'REVIEW: passes' || E'\n\n' || 'synth review output')
           )
     WHERE id = v_wi;

    -- This is the trigger-firing UPDATE
    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi;

    SELECT count(*) INTO v_pwid_count_after
      FROM stewards.pending_file_writes WHERE source_id = v_wi::text;

    RAISE NOTICE '[case 1: auto_mat ON, sabbath OFF] pending_file_writes count: % → %', v_pwid_count_before, v_pwid_count_after;
    IF v_pwid_count_after = v_pwid_count_before THEN
        RAISE EXCEPTION 'expected new pending_file_write row';
    END IF;

    -- Case 2: auto_materialize OFF (pipeline default), no override
    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"H.1.6.2 trigger smoke — auto_mat OFF"}'::jsonb,
        'h162-smoke-mat-off', 'human', NULL, v_intent_id
    ) INTO v_wi;
    UPDATE stewards.work_items
       SET file_destination = 'research/h162-smoke-off.md',
           -- auto_materialize_enabled stays NULL → inherit pipeline default (false)
           sabbath_enabled = false,
           stage_results = jsonb_build_object('review', jsonb_build_object('output','REVIEW: passes' || E'\n\n' || 'x'))
     WHERE id = v_wi;

    SELECT count(*) INTO v_pwid_count_before
      FROM stewards.pending_file_writes WHERE source_id = v_wi::text;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi;

    SELECT count(*) INTO v_pwid_count_after
      FROM stewards.pending_file_writes WHERE source_id = v_wi::text;

    RAISE NOTICE '[case 2: auto_mat OFF (default)] pending_file_writes count: % → % (expect no change)', v_pwid_count_before, v_pwid_count_after;
    IF v_pwid_count_after <> v_pwid_count_before THEN
        RAISE EXCEPTION 'expected NO new pending_file_write row when auto_mat is off';
    END IF;

    -- Case 3: sabbath ON via pipeline default (research-write defaults sabbath=true), no work_item override
    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"H.1.6.2 trigger smoke — sabbath ON"}'::jsonb,
        'h162-smoke-sab', 'human', NULL, v_intent_id
    ) INTO v_wi;
    -- sabbath_enabled stays NULL → inherit pipeline default (true)
    -- auto_materialize_enabled = false → skip auto-mat for this test

    SELECT count(*) INTO v_count_before
      FROM stewards.work_queue
     WHERE payload->>'_sabbath' = 'true' AND payload->>'_work_item_id' = v_wi::text;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi;

    SELECT count(*) INTO v_count_after
      FROM stewards.work_queue
     WHERE payload->>'_sabbath' = 'true' AND payload->>'_work_item_id' = v_wi::text;

    RAISE NOTICE '[case 3: sabbath ON (default), auto_mat OFF] sabbath work_queue count: % → %', v_count_before, v_count_after;
    IF v_count_after = v_count_before THEN
        RAISE EXCEPTION 'expected new sabbath work_queue row';
    END IF;

    -- Case 4: forward-only — UPDATE maturity from verified back to verified should not re-fire
    SELECT count(*) INTO v_count_before
      FROM stewards.work_queue
     WHERE payload->>'_sabbath' = 'true' AND payload->>'_work_item_id' = v_wi::text;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi;  -- no transition

    SELECT count(*) INTO v_count_after
      FROM stewards.work_queue
     WHERE payload->>'_sabbath' = 'true' AND payload->>'_work_item_id' = v_wi::text;

    RAISE NOTICE '[case 4: no-op UPDATE (already verified)] sabbath count: % → % (expect no change)', v_count_before, v_count_after;
    IF v_count_after <> v_count_before THEN
        RAISE EXCEPTION 'no-op UPDATE should not re-fire sabbath';
    END IF;

    -- Cleanup: error out any queued sabbath chats, delete synthetic rows
    UPDATE stewards.work_queue SET status='error', error='h1-6-2 smoke cleanup'
     WHERE payload->>'_work_item_id' IN (
        SELECT id::text FROM stewards.work_items WHERE slug LIKE 'h162-smoke-%'
     ) AND status='pending';

    DELETE FROM stewards.pending_file_writes
     WHERE source_id IN (SELECT id::text FROM stewards.work_items WHERE slug LIKE 'h162-smoke-%');
    DELETE FROM stewards.work_queue
     WHERE payload->>'_work_item_id' IN (
        SELECT id::text FROM stewards.work_items WHERE slug LIKE 'h162-smoke-%'
     );
    DELETE FROM stewards.work_items WHERE slug LIKE 'h162-smoke-%';

    RAISE NOTICE 'H.1.6.2/3/4 smoke PASSED';
END
$$;
