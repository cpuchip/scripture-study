-- =====================================================================
-- cv5 (2026-06-04) — code-pr base-branch support (chained / fan-out builds).
--
-- For a multi-PR app build, each item must build on the prior items' work and
-- land WITHOUT touching main (main's merge = the human Hinge + a deploy). This
-- adds an `input.base_branch` knob to the code-pr clone + pr stages:
--   clone: coder_sandbox_start ... branch="{{input.base_branch}}"  → clones that
--          base branch (the tool already supports branch; CloneRepo --branch),
--          so the worktree carries prior chained items' merged work.
--   pr:    coder_open_pr ... base="{{input.base_branch}}"           → opens the
--          PR against the base branch, not main.
-- The orchestrator creates an integration branch (e.g. agent/night-build) off
-- main, sets input.base_branch on each child, reviews + merges each PR into the
-- integration branch, and leaves the integration-branch -> main merge (the
-- deploy Hinge) to the human.
--
-- Targeted string-replace via jsonb_set on just the two stage templates (low
-- blast radius; idempotent — the replace no-ops once applied). Callers MUST set
-- input.base_branch (use "main" for an ordinary single-PR code-pr).
-- =====================================================================

-- clone stage (stages[0]) — clone the base branch.
UPDATE stewards.pipelines
SET stages = jsonb_set(
    stages, '{0,input_template}',
    to_jsonb(replace(
        stages->0->>'input_template',
        'repo="{{input.repo}}". The substrate clones the allow-listed repo into your worktree',
        'repo="{{input.repo}}", branch="{{input.base_branch}}" (clone this base branch so your work builds on the prior chained items). The substrate clones the allow-listed repo at that base branch into your worktree'
    ))
)
WHERE family = 'code-pr';

-- pr stage (stages[4]) — open the PR against the base branch, not main.
UPDATE stewards.pipelines
SET stages = jsonb_set(
    stages, '{4,input_template}',
    to_jsonb(replace(
        stages->4->>'input_template',
        'as evidence, and draft=true.',
        'as evidence, base="{{input.base_branch}}" (open the PR against the base branch, NOT main), and draft=true.'
    ))
)
WHERE family = 'code-pr';
