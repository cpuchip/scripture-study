-- Smoke H.1.2/H.1.3: research-write pipeline shape + tools_disabled wiring
-- IMPORTANT: this is a SCHEMA-LEVEL smoke. It creates a work_item, advances
-- it manually through the stages, dispatches each, and inspects the
-- work_queue payload. It DOES NOT let the bgworker actually fire the
-- chats — we cancel before the worker can pick them up (and the worker
-- should be paused via `compose stop bridge` for full safety).
-- =====================================================================

DO $$
DECLARE
    v_wi uuid;
    v_pipe stewards.pipelines%ROWTYPE;
    v_stage_gather    jsonb;
    v_stage_synth     jsonb;
    v_stage_review    jsonb;
    v_intent_id       uuid;
    v_work_id_gather  bigint;
    v_work_id_synth   bigint;
    v_work_id_review  bigint;
    v_payload         jsonb;
    v_tools_off       boolean;
BEGIN
    -- 1. Pipeline shape
    SELECT * INTO v_pipe FROM stewards.pipelines WHERE family = 'research-write';
    IF v_pipe.family IS NULL THEN
        RAISE EXCEPTION 'research-write pipeline missing';
    END IF;
    RAISE NOTICE 'pipeline: family=% stages=% sabbath=% atone=% template=%',
        v_pipe.family,
        jsonb_array_length(v_pipe.stages),
        v_pipe.sabbath_enabled,
        v_pipe.atonement_enabled,
        v_pipe.file_destination_template;

    -- 2. Stage tools_disabled values
    SELECT stewards.pipeline_stage_lookup('research-write', 'gather') INTO v_stage_gather;
    SELECT stewards.pipeline_stage_lookup('research-write', 'synthesize') INTO v_stage_synth;
    SELECT stewards.pipeline_stage_lookup('research-write', 'review') INTO v_stage_review;
    RAISE NOTICE 'gather tools_disabled=%   synthesize tools_disabled=%   review tools_disabled=%',
        v_stage_gather->>'tools_disabled',
        v_stage_synth->>'tools_disabled',
        v_stage_review->>'tools_disabled';
    IF (v_stage_review->>'tools_disabled')::boolean IS NOT TRUE THEN
        RAISE EXCEPTION 'review stage should have tools_disabled=true';
    END IF;

    -- 3. Create a work_item attached to general-research intent
    SELECT id INTO v_intent_id FROM stewards.intents WHERE slug = 'general-research';
    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"Smoke H.1.2: validate pipeline shape only — bgworker should be paused before this runs"}'::jsonb,
        'h12-smoke',
        'human',
        NULL,
        v_intent_id
    ) INTO v_wi;
    RAISE NOTICE 'work_item created: %', v_wi;

    -- 4. Dispatch gather; inspect payload
    v_work_id_gather := stewards.work_item_dispatch_stage(v_wi);
    SELECT payload INTO v_payload FROM stewards.work_queue WHERE id = v_work_id_gather;
    v_tools_off := COALESCE((v_payload->>'tools_disabled')::boolean, false);
    RAISE NOTICE 'gather dispatch work_id=% payload.tools_disabled=% (expect false)', v_work_id_gather, v_tools_off;
    IF v_tools_off THEN RAISE EXCEPTION 'gather should NOT have tools_disabled'; END IF;
    -- Cancel before bgworker picks it up
    UPDATE stewards.work_queue SET status='error', error='h12 smoke (gather)' WHERE id = v_work_id_gather AND status='pending';

    -- 5. Force-advance to synthesize, dispatch, inspect
    UPDATE stewards.work_items SET status='pending', current_stage='synthesize',
       stage_results = jsonb_build_object('gather', jsonb_build_object('output','synthetic gather output'))
     WHERE id = v_wi;
    v_work_id_synth := stewards.work_item_dispatch_stage(v_wi);
    SELECT payload INTO v_payload FROM stewards.work_queue WHERE id = v_work_id_synth;
    v_tools_off := COALESCE((v_payload->>'tools_disabled')::boolean, false);
    RAISE NOTICE 'synthesize dispatch work_id=% payload.tools_disabled=% (expect false)', v_work_id_synth, v_tools_off;
    IF v_tools_off THEN RAISE EXCEPTION 'synthesize should NOT have tools_disabled'; END IF;
    UPDATE stewards.work_queue SET status='error', error='h12 smoke (synth)' WHERE id = v_work_id_synth AND status='pending';

    -- 6. Force-advance to review, dispatch, inspect — THE KEY CHECK
    UPDATE stewards.work_items SET status='pending', current_stage='review',
       stage_results = stage_results || jsonb_build_object('synthesize', jsonb_build_object('output','synthetic synth output'))
     WHERE id = v_wi;
    v_work_id_review := stewards.work_item_dispatch_stage(v_wi);
    SELECT payload INTO v_payload FROM stewards.work_queue WHERE id = v_work_id_review;
    v_tools_off := COALESCE((v_payload->>'tools_disabled')::boolean, false);
    RAISE NOTICE 'review dispatch work_id=% payload.tools_disabled=% (expect TRUE)', v_work_id_review, v_tools_off;
    IF NOT v_tools_off THEN RAISE EXCEPTION 'review SHOULD have tools_disabled=true'; END IF;
    UPDATE stewards.work_queue SET status='error', error='h12 smoke (review)' WHERE id = v_work_id_review AND status='pending';

    RAISE NOTICE 'H.1.2/H.1.3 smoke PASSED — pipeline shape + tools_disabled propagation verified';

    -- Cleanup
    DELETE FROM stewards.work_queue WHERE payload->>'_work_item_id' = v_wi::text;
    DELETE FROM stewards.work_items WHERE id = v_wi;
    DELETE FROM stewards.sessions WHERE id LIKE 'wi--' || substring(v_wi::text FROM 1 FOR 8) || '%';
END
$$;
