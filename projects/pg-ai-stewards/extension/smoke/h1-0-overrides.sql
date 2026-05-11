-- Smoke H.1.0: work_item override semantics for sabbath + atonement
DO $$
DECLARE
    v_wi_a uuid;  -- inherits pipeline default (sabbath_enabled=true, atonement_enabled=false)
    v_wi_b uuid;  -- explicit override: sabbath=false, atonement=true
    v_wi_c uuid;  -- explicit override: sabbath=true, atonement=false (same as pipeline)
    v_pipe_sabbath  boolean;
    v_pipe_atone    boolean;
    v_work_id_a     bigint;
    v_work_id_b_sab bigint;
    v_work_id_b_atn bigint;
    v_err           text;
BEGIN
    SELECT sabbath_enabled, atonement_enabled INTO v_pipe_sabbath, v_pipe_atone
      FROM stewards.pipelines WHERE family = 'study-write';
    RAISE NOTICE 'study-write pipeline defaults: sabbath=% atonement=%', v_pipe_sabbath, v_pipe_atone;

    SELECT stewards.work_item_create('study-write',
        '{"binding_question":"H.1.0 smoke A — inherits"}'::jsonb,
        'h10-smoke-a', 'human', NULL, NULL) INTO v_wi_a;

    SELECT stewards.work_item_create('study-write',
        '{"binding_question":"H.1.0 smoke B — overrides"}'::jsonb,
        'h10-smoke-b', 'human', NULL, NULL) INTO v_wi_b;
    UPDATE stewards.work_items
       SET sabbath_enabled = false, atonement_enabled = true
     WHERE id = v_wi_b;

    -- ---- A: pipeline default (sabbath ON) → sabbath_dispatch should succeed
    BEGIN
        v_work_id_a := stewards.sabbath_dispatch(v_wi_a);
        RAISE NOTICE 'A sabbath_dispatch ok work_id=%', v_work_id_a;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'A UNEXPECTED FAILURE: %', SQLERRM;
    END;

    -- ---- B: override sabbath=false → sabbath_dispatch should RAISE
    BEGIN
        v_work_id_b_sab := stewards.sabbath_dispatch(v_wi_b);
        RAISE NOTICE 'B sabbath_dispatch UNEXPECTED SUCCESS work_id=%', v_work_id_b_sab;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'B sabbath correctly blocked: %', SQLERRM;
    END;

    -- ---- B: override atonement=true (pipeline default is false) → maybe_enqueue should fire
    BEGIN
        v_work_id_b_atn := stewards.maybe_enqueue_atonement(v_wi_b);
        IF v_work_id_b_atn IS NULL THEN
            RAISE NOTICE 'B atonement override=true returned NULL (failure inside dispatch — see notice above; this is OK for smoke if dispatch raised)';
        ELSE
            RAISE NOTICE 'B atonement override fired ok work_id=%', v_work_id_b_atn;
        END IF;
    END;

    -- ---- A inherits atonement=false → maybe_enqueue should return NULL silently
    DECLARE
        v_id_a_atn bigint;
    BEGIN
        v_id_a_atn := stewards.maybe_enqueue_atonement(v_wi_a);
        IF v_id_a_atn IS NULL THEN
            RAISE NOTICE 'A inherits atonement=false → NULL (correct)';
        ELSE
            RAISE NOTICE 'A UNEXPECTED SUCCESS atonement returned %', v_id_a_atn;
        END IF;
    END;

    -- ---- D-H2: confirm pipelines.maturity_ladder column populated
    DECLARE v_ladder jsonb;
    BEGIN
        SELECT maturity_ladder INTO v_ladder FROM stewards.pipelines WHERE family = 'study-write';
        RAISE NOTICE 'D-H2 study-write ladder: %', v_ladder::text;
        IF jsonb_array_length(v_ladder) <> 6 THEN
            RAISE EXCEPTION 'D-H2 ladder wrong length: %', v_ladder;
        END IF;
    END;
END
$$;
