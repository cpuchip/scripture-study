-- =====================================================================
-- Batch H.1.6.1 — maturity advance hook in work_item_advance
--
-- Gap surfaced 2026-05-11 during H.1.5d: every work_item in the DB
-- has maturity='raw'. The maturity column was advanced ONLY via the
-- gate-decision machinery (apply_gate_decision action='advance'),
-- never via the auto_advance stage progression. study-write items
-- never had maturity set; research-write items reached 'completed'
-- status with maturity still 'raw'.
--
-- Fix: patch work_item_advance to look up pipeline_stage_maturity
-- for the completing stage. If produces_maturity is set, UPDATE
-- work_items.maturity — but FORWARD-ONLY (per D-H6.1 ratification).
-- Re-running an earlier stage does NOT downgrade the high-water mark.
--
-- The forward-only check uses pipelines.maturity_ladder (D-H2 column).
-- We compute ladder array indices for current + new and only UPDATE if
-- new_idx > current_idx.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.work_item_advance(
    p_work_item_id uuid,
    p_stage_output jsonb DEFAULT '{}'::jsonb
)
RETURNS text
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline        stewards.pipelines%ROWTYPE;
    v_stage           jsonb;
    v_next_name       text;
    v_auto_advance    boolean;
    v_results         jsonb;
    v_completing      text;
    v_new_maturity    text;
    v_current_idx     int;
    v_new_idx         int;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;
    IF v_wi.status NOT IN ('in_progress', 'awaiting_review', 'pending') THEN
        RAISE EXCEPTION 'work_item %: cannot advance from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_next_name    := v_stage->>'next';
    v_auto_advance := COALESCE((v_stage->>'auto_advance')::bool, true);
    v_completing   := v_wi.current_stage;

    -- Record this stage's output keyed by stage name.
    v_results := v_wi.stage_results
              || jsonb_build_object(v_completing,
                     p_stage_output
                     || jsonb_build_object('completed_at', now()));

    -- ----- H.1.6.1: maturity advance hook (forward-only) -----
    -- Look up what maturity rung this completing stage produces.
    SELECT produces_maturity INTO v_new_maturity
      FROM stewards.pipeline_stage_maturity
     WHERE pipeline_family = v_wi.pipeline_family
       AND stage_name      = v_completing;

    -- Resolve the work_item's pipeline once for the ladder lookup.
    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;

    -- Compute ladder indices. The maturity_ladder jsonb is an ordered
    -- array; index() returns the position (0-based). If a rung isn't
    -- in the ladder, idx is NULL.
    IF v_new_maturity IS NOT NULL AND v_pipeline.maturity_ladder IS NOT NULL THEN
        SELECT pos - 1 INTO v_current_idx
          FROM jsonb_array_elements_text(v_pipeline.maturity_ladder)
          WITH ORDINALITY AS t(rung, pos)
         WHERE rung = COALESCE(v_wi.maturity, 'raw');

        SELECT pos - 1 INTO v_new_idx
          FROM jsonb_array_elements_text(v_pipeline.maturity_ladder)
          WITH ORDINALITY AS t(rung, pos)
         WHERE rung = v_new_maturity;

        -- Forward-only: only set if both rungs are in the ladder AND
        -- new index > current index. If either is NULL (rung missing
        -- from this pipeline's ladder), skip — keep current.
        IF v_current_idx IS NOT NULL
           AND v_new_idx IS NOT NULL
           AND v_new_idx > v_current_idx
        THEN
            -- The UPDATE happens as part of the next status UPDATE below;
            -- carry the value through.
            NULL;
        ELSE
            v_new_maturity := NULL;  -- signal: do not change maturity
        END IF;
    END IF;
    -- ----- end maturity advance hook -----

    IF v_next_name IS NULL OR v_next_name = '' THEN
        -- Terminal: no next stage.
        UPDATE stewards.work_items
           SET stage_results = v_results,
               status        = 'completed',
               completed_at  = now(),
               maturity      = COALESCE(v_new_maturity, maturity),
               updated_at    = now()
         WHERE id = p_work_item_id;
        RETURN NULL;
    END IF;

    IF stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_next_name) IS NULL THEN
        RAISE EXCEPTION
            'work_item %: stage %s `next` references missing stage %',
            p_work_item_id, v_completing, v_next_name;
    END IF;

    UPDATE stewards.work_items
       SET stage_results = v_results,
           current_stage = v_next_name,
           status        = CASE WHEN v_auto_advance THEN 'pending'
                                ELSE 'awaiting_review' END,
           maturity      = COALESCE(v_new_maturity, maturity),
           updated_at    = now()
     WHERE id = p_work_item_id;

    RETURN v_next_name;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_advance(uuid, jsonb) IS
'H.1.6.1 (Batch H): on each stage completion, look up pipeline_stage_maturity for the completing stage. If produces_maturity is set AND the new rung is forward of current in the pipeline''s maturity_ladder, UPDATE work_items.maturity. Forward-only per D-H6.1 (re-running earlier stages does not downgrade the high-water mark).';
