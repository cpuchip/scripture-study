-- =====================================================================
-- cv12 (2026-06-04) — seed code-pr feedback defaults so the FIRST dispatch works.
--
-- Bug: cv11 appended a "## PLAN REVIEW FEEDBACK" block referencing
-- {{input.plan_feedback}} to the plan stage template, and cv6 references
-- {{input.review_feedback}} in the implement template. resolve_template_path
-- THROWS when a referenced path is NULL (not "" ) — so a freshly-created
-- code-pr work_item whose input omits these fields fails to auto-dispatch the
-- plan stage and lands in status='awaiting_review' with
-- "resolve_template_path: path input.plan_feedback resolved to NULL".
--
-- The bake-off/plancritic runs only worked because their seed input happened to
-- carry review_feedback/plan_feedback as "". This makes that implicit
-- convention explicit + automatic: the BEFORE-INSERT stamp trigger now defaults
-- plan_feedback and review_feedback to "" for code-pr work_items when missing.
-- The critic loops overwrite them with real feedback on a bounce, so the empty
-- default only affects the first pass. SQL-only; no rebuild.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.stamp_code_write_sandbox()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    -- stable sandbox id (cc3 + cv3): one worktree id across the revise loop.
    IF NEW.pipeline_family IN ('code-write', 'code-pr')
       AND (NEW.input IS NULL OR (NEW.input->>'sandbox') IS NULL)
    THEN
        NEW.input := COALESCE(NEW.input, '{}'::jsonb)
            || jsonb_build_object('sandbox', 'wi-' || substring(NEW.id::text FROM 1 FOR 8));
    END IF;

    -- cv12: seed the critic-loop feedback fields the code-pr stage templates
    -- reference, so the first dispatch (no bounce yet) doesn't hit a NULL path.
    IF NEW.pipeline_family = 'code-pr' THEN
        IF (NEW.input->>'plan_feedback') IS NULL THEN
            NEW.input := COALESCE(NEW.input, '{}'::jsonb) || jsonb_build_object('plan_feedback', '');
        END IF;
        IF (NEW.input->>'review_feedback') IS NULL THEN
            NEW.input := COALESCE(NEW.input, '{}'::jsonb) || jsonb_build_object('review_feedback', '');
        END IF;
    END IF;

    RETURN NEW;
END;
$func$;
