-- =====================================================================
-- cv6 (2026-06-04) — code-pr plan-conformance CRITIC stage + revise loop.
--
-- The gap the kimi night-build exposed: `verify` only checks GROUND TRUTH
-- (does it compile, do tests pass). It never asks "does this fulfil the PLAN" —
-- room-scoped? handler actually tested? every acceptance criterion met? A
-- fresh, strong critic (a DIFFERENT model than the implementer) catches what
-- the implementer's self-report misses, and bounces it back to dev — exactly
-- the manual review I did, made into a pipeline stage.
--
-- Shape: clone -> plan -> implement -> verify -> REVIEW -> pr.
--   review (critic, qwen3.7-max, tools ON): inspects the REAL diff against the
--     binding question + an explicit acceptance-criteria checklist (carried in
--     input.acceptance_criteria). Emits "REVIEW: passes" or "REVIEW: revise"+fixes.
--   On revise: work_item_advance loops back to `implement` with the feedback
--     injected (input.review_feedback), capped at input.revise_cap (default 2);
--     past the cap it surfaces awaiting_review (the Hinge — never auto-PRs a
--     thrice-deficient change).
--
-- The loop-back lives in work_item_advance (SQL — the auto-advance trigger
-- 3c2 calls it). The new branch is tightly gated to pipeline_family='code-pr'
-- AND completing stage='review', so no other pipeline changes. The function
-- body below is the LIVE definition (h1-6-1) verbatim + that one branch.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. work_item_advance — live body + the code-pr review loop-back branch.
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
    -- cv6: critic review loop-back state
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

    -- Record this stage's output keyed by stage name.
    v_results := v_wi.stage_results
              || jsonb_build_object(v_completing,
                     p_stage_output
                     || jsonb_build_object('completed_at', now()));

    -- ----- cv6: code-pr critic review loop-back -----
    -- When the `review` (critic) stage completes, branch on its verdict.
    -- REVIEW: passes  -> fall through to the normal advance (next = pr).
    -- REVIEW: revise  -> loop back to `implement` with feedback (capped),
    --                    else surface for a human (awaiting_review).
    IF v_wi.pipeline_family = 'code-pr' AND v_completing = 'review' THEN
        v_verdict_text := COALESCE(p_stage_output->>'output', '');
        v_revise_count := COALESCE((v_wi.input->>'revise_count')::int, 0);
        v_revise_cap   := COALESCE((v_wi.input->>'revise_cap')::int, 2);

        IF v_verdict_text !~* '^\s*REVIEW:\s*passes' THEN
            IF v_revise_count < v_revise_cap THEN
                -- Loop back to implement with the critic's feedback injected.
                UPDATE stewards.work_items
                   SET stage_results = v_results,
                       current_stage = 'implement',
                       input         = input
                                     || jsonb_build_object(
                                          'review_feedback', v_verdict_text,
                                          'revise_count', v_revise_count + 1),
                       status        = 'pending',
                       updated_at    = now()
                 WHERE id = p_work_item_id;
                RETURN 'implement';
            ELSE
                -- Cap exhausted: surface to a human (the Hinge). Do not PR.
                UPDATE stewards.work_items
                   SET stage_results     = v_results,
                       status            = 'awaiting_review',
                       quarantine_reason = COALESCE(
                           quarantine_reason,
                           format('critic: still deficient after %s revise cycle(s)', v_revise_cap)),
                       error             = COALESCE(
                           error,
                           'critic review deficient after revise cap; needs a human'),
                       updated_at        = now()
                 WHERE id = p_work_item_id;
                RETURN NULL;
            END IF;
        END IF;
        -- passes: fall through to the normal advance below (next = pr).
    END IF;
    -- ----- end cv6 review loop-back -----

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

        IF v_current_idx IS NOT NULL
           AND v_new_idx IS NOT NULL
           AND v_new_idx > v_current_idx
        THEN
            NULL;
        ELSE
            v_new_maturity := NULL;
        END IF;
    END IF;
    -- ----- end maturity advance hook -----

    IF v_next_name IS NULL OR v_next_name = '' THEN
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

-- ---------------------------------------------------------------------
-- 2. Pipeline surgery: insert `review` between verify and pr (idempotent).
-- ---------------------------------------------------------------------

-- 2a. verify -> review (was verify -> pr). Idempotent.
UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{3,next}', '"review"')
 WHERE family = 'code-pr'
   AND (stages->3->>'name') = 'verify';

-- 2b. Append the review (critic) stage just before pr, only if absent.
--     pr is currently the last element (index 4); rebuild as
--     [..verify] + [review] + [pr].
UPDATE stewards.pipelines
   SET stages = (stages - 4)
              || jsonb_build_array(
                   jsonb_build_object(
                     'name',           'review',
                     'next',           'pr',
                     'model',          'qwen3.7-max',
                     'provider',       'opencode_go',
                     'agent_family',   'dev',
                     'auto_advance',   true,
                     'tools_disabled', false,
                     'input_template',
                       'You are the REVIEWER (critic) for a code change — a fresh, strict set of eyes, a DIFFERENT model than the implementer. Judge the change against the plan, not the implementer''s self-report.' || E'\n\n' ||
                       'Task (binding question): {{input.binding_question}}' || E'\n\n' ||
                       'ACCEPTANCE CRITERIA — the change must satisfy EVERY one:' || E'\n' ||
                       '{{input.acceptance_criteria}}' || E'\n\n' ||
                       'The change is implemented in sandbox "{{input.sandbox}}" (the cloned repo at /work) and built+tested green by verify. The implementer reported:' || E'\n' ||
                       '{{stage_results.implement.output}}' || E'\n\n' ||
                       'Inspect the ACTUAL change — do NOT trust the report:' || E'\n' ||
                       '1. coder_sandbox_start with sandbox="{{input.sandbox}}" (reuse the worktree; no repo arg).' || E'\n' ||
                       '2. coder_shell, sandbox="{{input.sandbox}}": run `git -c safe.directory=* diff {{input.base_branch}}...HEAD` and `git -c safe.directory=* log --oneline {{input.base_branch}}..HEAD` to see the real diff.' || E'\n' ||
                       '3. coder_read / coder_grep the changed files as needed; re-run the build+test command if a criterion needs it.' || E'\n\n' ||
                       'Judge against EACH acceptance criterion AND the binding question. A criterion is met only if the actual code shows it — not because the report claims it. Watch specifically for: scope the plan implies but the code skipped (e.g. room-scoping), the actual handler/entrypoint being untested (a test that re-implements the logic inline does NOT count), and any criterion silently dropped.' || E'\n\n' ||
                       'Return EXACTLY one of:' || E'\n' ||
                       '  (a) First line "REVIEW: passes" — ONLY if every acceptance criterion is met — then one short line per criterion confirming how.' || E'\n' ||
                       '  (b) First line "REVIEW: revise" — if ANY criterion is unmet or the change diverges from the plan — then a NUMBERED list: each unmet criterion, what is wrong (cite the file/line), and the SPECIFIC fix the implementer must make. The implementer receives this verbatim and must fix exactly these points.'
                   )
                 )
              || jsonb_build_array(stages->4)
 WHERE family = 'code-pr'
   AND (stages->4->>'name') = 'pr'
   AND NOT (stages @> '[{"name":"review"}]');

-- 2c. implement stage: add a revision-feedback section (idempotent string-append).
--     On the first pass input.review_feedback is empty; on a critic loop-back
--     work_item_advance injects the critic's feedback here.
UPDATE stewards.pipelines
   SET stages = jsonb_set(
        stages, '{2,input_template}',
        to_jsonb(
            (stages->2->>'input_template')
            || E'\n\n## REVISION REQUESTED (address fully if present)\n'
            || 'A reviewer checked a prior attempt against the plan and asked for these changes. If the section below is non-empty, you are on a revise cycle: address EVERY point before reporting green.\n'
            || '{{input.review_feedback}}'
        ))
 WHERE family = 'code-pr'
   AND (stages->2->>'name') = 'implement'
   AND (stages->2->>'input_template') NOT LIKE '%REVISION REQUESTED%';

-- ---------------------------------------------------------------------
-- 3. stage_models — the critic defaults to qwen3.7-max (a DIFFERENT strong
--    model than the dev/implement model; the bake-off holds it constant).
--    No pipeline_stage_maturity row: review is a gate, it must not change the
--    high-water maturity (verify already set 'verified').
-- ---------------------------------------------------------------------
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-pr', 'review', 'qwen3.7-max', 'Plan-conformance critic. A DIFFERENT strong model than the implementer (default qwen3.7-max). Inspects the real diff vs acceptance_criteria; REVIEW: passes -> pr, REVIEW: revise -> loop back to implement (capped, then awaiting_review).')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;
