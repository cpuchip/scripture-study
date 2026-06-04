-- =====================================================================
-- cv11 (2026-06-04) — code-pr PLAN-CRITIC loop ("council before you build").
--
-- The implement-critic (cv6) catches problems AFTER the code is written. This
-- adds the same idea one stage EARLIER and cheaper: a `plan_review` critic
-- (qwen3.7-max) reviews the kimi-written plan against the binding question +
-- acceptance criteria BEFORE any code. They iterate to agreement:
--   PLAN: approved -> proceed to implement.
--   PLAN: revise   -> loop back to `plan` with the feedback injected (capped at
--                     input.plan_revise_cap, default 2); past the cap, proceed
--                     to implement anyway with the best plan reached (do NOT
--                     deadlock the build on plan-perfectionism).
--
-- Pipeline becomes: clone -> plan -> plan_review -> implement -> verify -> review -> pr.
-- The loop-back lives in work_item_advance (live cv6 body verbatim + one new
-- branch gated to code-pr plan_review, alongside the cv6 review branch). EXPERIMENTAL:
-- to be A/B'd vs a no-plan-review run; revert (like cv8/cv9) if it does not earn
-- its cost.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. work_item_advance — cv6 body + the plan_review loop-back branch.
-- ---------------------------------------------------------------------
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
    v_verdict_text    text;
    v_revise_count    int;
    v_revise_cap      int;
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

    v_results := v_wi.stage_results
              || jsonb_build_object(v_completing,
                     p_stage_output
                     || jsonb_build_object('completed_at', now()));

    -- ----- cv6: code-pr implement-critic review loop-back -----
    IF v_wi.pipeline_family = 'code-pr' AND v_completing = 'review' THEN
        v_verdict_text := COALESCE(p_stage_output->>'output', '');
        v_revise_count := COALESCE((v_wi.input->>'revise_count')::int, 0);
        v_revise_cap   := COALESCE((v_wi.input->>'revise_cap')::int, 2);
        IF v_verdict_text !~* '^\s*REVIEW:\s*passes' THEN
            IF v_revise_count < v_revise_cap THEN
                UPDATE stewards.work_items
                   SET stage_results = v_results,
                       current_stage = 'implement',
                       input         = input || jsonb_build_object(
                                          'review_feedback', v_verdict_text,
                                          'revise_count', v_revise_count + 1),
                       status        = 'pending',
                       updated_at    = now()
                 WHERE id = p_work_item_id;
                RETURN 'implement';
            ELSE
                UPDATE stewards.work_items
                   SET stage_results     = v_results,
                       status            = 'awaiting_review',
                       quarantine_reason = COALESCE(quarantine_reason,
                           format('critic: still deficient after %s revise cycle(s)', v_revise_cap)),
                       error             = COALESCE(error,
                           'critic review deficient after revise cap; needs a human'),
                       updated_at        = now()
                 WHERE id = p_work_item_id;
                RETURN NULL;
            END IF;
        END IF;
    END IF;
    -- ----- end cv6 review loop-back -----

    -- ----- cv11: code-pr plan-critic loop-back -----
    IF v_wi.pipeline_family = 'code-pr' AND v_completing = 'plan_review' THEN
        v_verdict_text := COALESCE(p_stage_output->>'output', '');
        v_revise_count := COALESCE((v_wi.input->>'plan_revise_count')::int, 0);
        v_revise_cap   := COALESCE((v_wi.input->>'plan_revise_cap')::int, 2);
        IF v_verdict_text !~* '^\s*PLAN:\s*approved' THEN
            IF v_revise_count < v_revise_cap THEN
                -- Loop back to plan with the plan-critic's feedback injected.
                UPDATE stewards.work_items
                   SET stage_results = v_results,
                       current_stage = 'plan',
                       input         = input || jsonb_build_object(
                                          'plan_feedback', v_verdict_text,
                                          'plan_revise_count', v_revise_count + 1),
                       status        = 'pending',
                       updated_at    = now()
                 WHERE id = p_work_item_id;
                RETURN 'plan';
            ELSE
                -- Cap reached: proceed to implement with the best plan (don't
                -- deadlock the build on plan disagreement).
                UPDATE stewards.work_items
                   SET stage_results = v_results,
                       current_stage = 'implement',
                       status        = 'pending',
                       updated_at    = now()
                 WHERE id = p_work_item_id;
                RETURN 'implement';
            END IF;
        END IF;
        -- approved: fall through to the normal advance (next = implement).
    END IF;
    -- ----- end cv11 plan-critic loop-back -----

    -- ----- H.1.6.1: maturity advance hook (forward-only) -----
    SELECT produces_maturity INTO v_new_maturity
      FROM stewards.pipeline_stage_maturity
     WHERE pipeline_family = v_wi.pipeline_family
       AND stage_name      = v_completing;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;

    IF v_new_maturity IS NOT NULL AND v_pipeline.maturity_ladder IS NOT NULL THEN
        SELECT pos - 1 INTO v_current_idx
          FROM jsonb_array_elements_text(v_pipeline.maturity_ladder)
          WITH ORDINALITY AS t(rung, pos)
         WHERE rung = COALESCE(v_wi.maturity, 'raw');

        SELECT pos - 1 INTO v_new_idx
          FROM jsonb_array_elements_text(v_pipeline.maturity_ladder)
          WITH ORDINALITY AS t(rung, pos)
         WHERE rung = v_new_maturity;

        IF v_current_idx IS NOT NULL AND v_new_idx IS NOT NULL AND v_new_idx > v_current_idx THEN
            NULL;
        ELSE
            v_new_maturity := NULL;
        END IF;
    END IF;

    IF v_next_name IS NULL OR v_next_name = '' THEN
        UPDATE stewards.work_items
           SET stage_results = v_results, status = 'completed', completed_at = now(),
               maturity = COALESCE(v_new_maturity, maturity), updated_at = now()
         WHERE id = p_work_item_id;
        RETURN NULL;
    END IF;

    IF stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_next_name) IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage %s `next` references missing stage %',
            p_work_item_id, v_completing, v_next_name;
    END IF;

    UPDATE stewards.work_items
       SET stage_results = v_results,
           current_stage = v_next_name,
           status        = CASE WHEN v_auto_advance THEN 'pending' ELSE 'awaiting_review' END,
           maturity      = COALESCE(v_new_maturity, maturity),
           updated_at    = now()
     WHERE id = p_work_item_id;

    RETURN v_next_name;
END;
$func$;

-- ---------------------------------------------------------------------
-- 2. Pipeline surgery: insert plan_review between plan and implement.
-- ---------------------------------------------------------------------

-- 2a. plan -> plan_review (was plan -> implement). Idempotent.
UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{1,next}', '"plan_review"')
 WHERE family = 'code-pr' AND (stages->1->>'name') = 'plan';

-- 2b. Insert the plan_review stage at position 3 (after plan), only if absent.
UPDATE stewards.pipelines p
   SET stages = (
        SELECT jsonb_agg(elem ORDER BY ord)
        FROM (
            SELECT elem, ord FROM jsonb_array_elements(p.stages) WITH ORDINALITY AS a(elem, ord)
             WHERE ord <= 2                       -- clone(1), plan(2)
            UNION ALL
            SELECT jsonb_build_object(
                     'name',           'plan_review',
                     'next',           'implement',
                     'model',          'qwen3.7-max',
                     'provider',       'opencode_go',
                     'agent_family',   'dev',
                     'auto_advance',   true,
                     'tools_disabled', true,
                     'input_template',
                       'You are the PLAN REVIEWER (critic) — a fresh, strict architect reviewing an implementation plan BEFORE any code is written. A different model wrote the plan; judge it against the task, not the planner.' || E'\n\n' ||
                       'Task (binding question): {{input.binding_question}}' || E'\n\n' ||
                       'ACCEPTANCE CRITERIA the final code must satisfy:' || E'\n' || '{{input.acceptance_criteria}}' || E'\n\n' ||
                       'The plan to review:' || E'\n' || '{{stage_results.plan.output}}' || E'\n\n' ||
                       'Judge whether this plan, IF implemented faithfully, would satisfy every acceptance criterion and is sound, idiomatic, and right-sized (not over- or under-engineered). Look specifically for: scope the task implies but the plan omits (e.g. room-scoping), criteria with no corresponding plan element, a missing or vague test strategy, and unnecessary complexity.' || E'\n\n' ||
                       'Return EXACTLY one of:' || E'\n' ||
                       '  (a) First line "PLAN: approved" — only if the plan would meet every criterion and is sound — then one short line per criterion noting how the plan covers it.' || E'\n' ||
                       '  (b) First line "PLAN: revise" — if anything is missing, unsound, or wrong-sized — then a NUMBERED list of the specific changes the planner must make. The planner gets this verbatim and must address each point.'
                   ) AS elem, 2.5::numeric AS ord
            UNION ALL
            SELECT elem, ord FROM jsonb_array_elements(p.stages) WITH ORDINALITY AS a(elem, ord)
             WHERE ord >= 3                       -- implement(3) onward
        ) combined(elem, ord)
   )
 WHERE p.family = 'code-pr'
   AND NOT (p.stages @> '[{"name":"plan_review"}]');

-- 2c. plan stage: add a plan-feedback section (idempotent string-append). On a
--     plan-critic loop-back, work_item_advance injects the critic's feedback here.
UPDATE stewards.pipelines
   SET stages = jsonb_set(
        stages, '{1,input_template}',
        to_jsonb(
            (stages->1->>'input_template')
            || E'\n\n## PLAN REVIEW FEEDBACK (address fully if present)\n'
            || 'A plan reviewer checked a prior version of this plan. If the section below is non-empty, revise the plan to address EVERY point.\n'
            || '{{input.plan_feedback}}'
        ))
 WHERE family = 'code-pr'
   AND (stages->1->>'name') = 'plan'
   AND (stages->1->>'input_template') NOT LIKE '%PLAN REVIEW FEEDBACK%';

-- ---------------------------------------------------------------------
-- 3. stage_models — plan_review critic defaults to qwen3.7-max.
-- ---------------------------------------------------------------------
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-pr', 'plan_review', 'qwen3.7-max', 'Plan critic (cv11): reviews the plan vs acceptance criteria before build; PLAN: approved -> implement, PLAN: revise -> loop back to plan (capped). EXPERIMENTAL — A/B vs no-plan-review.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model, notes = EXCLUDED.notes;
