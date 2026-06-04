-- =====================================================================
-- cv3 (2026-06-03) — the `code-pr` pipeline (coder-v2 CV2.3): work inside an
-- existing repo and land the work as a reviewable PR.
--
-- Five stages, the code-write loop (cc3) wrapped with a clone front + a pr back:
--   clone     (dev, tools ON)  → researched — start the sandbox WITH repo= so the
--                                bridge clones the allow-listed repo into the
--                                shared worktree (token never in the sandbox) and
--                                mounts it at /work; survey the repo for the plan.
--   plan      (dev, tools off)  → planned   — implementation plan grounded in the
--                                real repo survey (NOT code).
--   implement (dev, tools ON)   → executing — write code in the cloned repo +
--                                iterate coder_shell build+test to GREEN
--                                (real exit codes = ground truth). Reuses the
--                                worktree (start WITHOUT repo — re-clone would wipe).
--   verify    (dev, tools ON)   → verified  — independently re-run build+test;
--                                REVIEW: passes / fail drives the gate.
--   pr        (dev, tools ON)   → verified  — coder_commit (LOCAL, no token) →
--                                coder_push → coder_open_pr (DRAFT). The substrate
--                                holds the token bridge-side.
--
-- THE HINGE is the human MERGE, not the PR-open (D-CV2.4). A draft PR is
-- outward-facing but a cheap walk-back (close + delete) and IS the review
-- surface — so pr auto-advances to a draft PR and stops; a human reviews and
-- merges (the irreversible step, outside the substrate). Contrast code-deploy
-- (cc5), where deploy IS production-affecting so `prepare` is the always-escalate
-- Hinge. Here the loop + PR-open run free; the merge is where trust decides.
--
-- The coder tools are already granted to `dev` (sandbox/write/edit/shell/glob/
-- grep/read by cc2; commit/push/open_pr by cv2-2) — cv3 adds NO new grants.
--
-- Input contract: the work_item input must carry `repo` (an allow-listed repo
-- URL) + `binding_question` (the task). `sandbox` is stamped by the shared
-- BEFORE-INSERT trigger below (extended from cc3 to cover code-pr) so all five
-- stages share one stable sandbox/worktree id across the revise loop.
--
-- Model: defaults to kimi-k2.6 (the proven coder — v1 built FizzBuzz + a working
-- WebSocket hub + the calc package on it end-to-end). `implement` is the stage to
-- escalate per-task for novel app code with no canonical template (qwen3.7-max or
-- deepseek-v4-pro) via stage_models.default_model or the work_item input — kept a
-- documented per-task knob rather than a blind default swap, since the iterative
-- write→build→test→green loop's RELIABILITY (many tool calls) is what's proven on
-- kimi-k2.6, and a reasoning model can exhaust its budget thinking before the loop
-- completes. Watch the first real ai-chattermax runs and re-tune (the proposal's
-- monitoring discipline).
-- =====================================================================

-- --- stable sandbox id: extend the cc3 trigger to also stamp code-pr ---
-- (CREATE OR REPLACE keeps the same trigger; just broadens the family match.
--  code-pr keys its worktree by the same wi-<8> id the sandbox tools take.)
CREATE OR REPLACE FUNCTION stewards.stamp_code_write_sandbox()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    IF NEW.pipeline_family IN ('code-write', 'code-pr')
       AND (NEW.input IS NULL OR (NEW.input->>'sandbox') IS NULL)
    THEN
        NEW.input := COALESCE(NEW.input, '{}'::jsonb)
            || jsonb_build_object('sandbox', 'wi-' || substring(NEW.id::text FROM 1 FOR 8));
    END IF;
    RETURN NEW;
END;
$func$;

-- --- the pipeline -----------------------------------------------------
INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder
)
VALUES (
    'code-pr',
    jsonb_build_array(
        jsonb_build_object(
            'name',           'clone',
            'next',           'plan',
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', false,
            'input_template',
                'Coding task on an existing repo: {{input.binding_question}}' || E'\n\n' ||
                'Repo: {{input.repo}}' || E'\n' ||
                'Your sandbox id: {{input.sandbox}}' || E'\n\n' ||
                'Clone the repo into your worktree and survey it — do NOT write code yet:' || E'\n' ||
                '1. coder_sandbox_start with sandbox="{{input.sandbox}}", repo="{{input.repo}}". The substrate clones ' ||
                'the allow-listed repo into your worktree and mounts it at /work; the GitHub token never enters your sandbox.' || E'\n' ||
                '2. Survey it so the next stage can plan against the REAL code: coder_glob to see the layout, then ' ||
                'coder_read the README / go.mod / package.json to identify the language, build tool, and conventions.' || E'\n\n' ||
                'Report a concise map: the stack, the key directories/files, the build+test command the repo uses, ' ||
                'and where the task''s change most likely belongs.'
        ),
        jsonb_build_object(
            'name',           'plan',
            'next',           'implement',
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', true,
            'input_template',
                'Coding task: {{input.binding_question}}' || E'\n\n' ||
                'Repo: {{input.repo}}' || E'\n\n' ||
                'Repo survey from the clone stage:' || E'\n' || '{{stage_results.clone.output}}' || E'\n\n' ||
                'Produce a concise implementation plan, NOT code:' || E'\n' ||
                '  - The files to create or change (paths relative to the repo root /work).' || E'\n' ||
                '  - The approach in a few sentences, consistent with the repo''s existing conventions.' || E'\n' ||
                '  - The exact build + test command that proves it works — the one this repo uses ' ||
                '(e.g. `go build ./... && go test ./...`, or `npm ci && npm test`). This command is the ground-truth gate.' || E'\n\n' ||
                'Keep it tight. The next stage implements against this plan in the cloned repo.'
        ),
        jsonb_build_object(
            'name',           'implement',
            'next',           'verify',
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', false,
            'input_template',
                'Coding task: {{input.binding_question}}' || E'\n\n' ||
                'Implementation plan:' || E'\n' || '{{stage_results.plan.output}}' || E'\n\n' ||
                'Your sandbox id (the repo is already cloned + mounted at /work): {{input.sandbox}}' || E'\n\n' ||
                'Implement it in the cloned repo using the coder tools:' || E'\n' ||
                '1. coder_sandbox_start with sandbox="{{input.sandbox}}" — NO repo arg. This reuses your existing ' ||
                'worktree with the clone. Do NOT pass repo= here; that would re-clone and wipe your work.' || E'\n' ||
                '2. Read the relevant existing files (coder_read / coder_grep) before changing them — match the repo''s conventions.' || E'\n' ||
                '3. Write/edit code with coder_write / coder_edit (paths relative to /work, the repo root).' || E'\n' ||
                '4. Build + test with coder_shell, sandbox="{{input.sandbox}}", running the build+test command from the plan.' || E'\n' ||
                '5. ITERATE: if the build or tests fail, read the real output, fix the code, and run again. ' ||
                'Do NOT stop until the build+test command exits 0 (green). The passing build+test is ground truth, not a judgment call.' || E'\n\n' ||
                'When green, report: what you changed, the files touched, and paste the final passing build+test output. ' ||
                'Do NOT commit or push — the pr stage lands the work.'
        ),
        jsonb_build_object(
            'name',           'verify',
            'next',           'pr',
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', false,
            'input_template',
                'Coding task: {{input.binding_question}}' || E'\n\n' ||
                'The implement stage reported:' || E'\n' || '{{stage_results.implement.output}}' || E'\n\n' ||
                'Independently verify — do NOT trust the report above. In sandbox "{{input.sandbox}}" (your cloned repo at /work):' || E'\n' ||
                '1. coder_sandbox_start with sandbox="{{input.sandbox}}" (NO repo — reuse the worktree).' || E'\n' ||
                '2. coder_shell, sandbox="{{input.sandbox}}", run the build + test command yourself.' || E'\n' ||
                '3. Inspect the REAL exit code and output.' || E'\n\n' ||
                'Return EXACTLY one of:' || E'\n' ||
                '  (a) A first line "REVIEW: passes" (only if the command exited 0), then the build/test output.' || E'\n' ||
                '  (b) A first line "REVIEW: fail", then the failing output and a short note on what still needs fixing.'
        ),
        jsonb_build_object(
            'name',           'pr',
            'next',           NULL,
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', false,
            'input_template',
                'Coding task: {{input.binding_question}}' || E'\n\n' ||
                'Repo: {{input.repo}}' || E'\n' ||
                'The change is implemented + verified green in sandbox "{{input.sandbox}}" (your cloned repo worktree).' || E'\n\n' ||
                'Implement summary:' || E'\n' || '{{stage_results.implement.output}}' || E'\n\n' ||
                'Land the work as a reviewable DRAFT pull request. The substrate holds the GitHub token bridge-side — ' ||
                'you commit LOCALLY (no token), and coder_push / coder_open_pr push + open the PR for you:' || E'\n' ||
                '1. coder_commit with sandbox="{{input.sandbox}}", a clear conventional-commit message describing the change, ' ||
                'and branch="agent/code-pr/{{input.sandbox}}".' || E'\n' ||
                '2. coder_push with sandbox="{{input.sandbox}}", branch="agent/code-pr/{{input.sandbox}}".' || E'\n' ||
                '3. coder_open_pr with sandbox="{{input.sandbox}}", a descriptive title, a body that explains the change ' ||
                'AND pastes the passing build+test output as evidence, and draft=true.' || E'\n\n' ||
                'Report the PR url. A human reviews and merges the PR — that merge is the Hinge. Open the draft and stop there; ' ||
                'do NOT attempt to merge.'
        )
    ),
    false,  -- sabbath_enabled
    true,   -- atonement_enabled: a PR run that hits a cost cap is worth atoning over
    NULL,   -- file_destination_template: the deliverable is a PR, not a markdown file
    NULL,
    '["raw","researched","planned","executing","verified"]'::jsonb
)
ON CONFLICT (family) DO UPDATE SET
    stages                    = EXCLUDED.stages,
    sabbath_enabled           = EXCLUDED.sabbath_enabled,
    atonement_enabled         = EXCLUDED.atonement_enabled,
    file_destination_template = EXCLUDED.file_destination_template,
    file_content_jsonpath     = EXCLUDED.file_content_jsonpath,
    maturity_ladder           = EXCLUDED.maturity_ladder;

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-pr', 'clone',     'kimi-k2.6', 'Clone the allow-listed repo into the worktree + survey it; coder tools on.'),
    ('code-pr', 'plan',      'kimi-k2.6', 'Implementation plan grounded in the repo survey; tools off.'),
    ('code-pr', 'implement', 'kimi-k2.6', 'Write + build/test loop in the cloned repo; coder tools on. ESCALATE per-task here for novel app code (qwen3.7-max / deepseek-v4-pro).'),
    ('code-pr', 'verify',    'kimi-k2.6', 'Independent build/test re-run in the cloned repo; coder tools on.'),
    ('code-pr', 'pr',        'kimi-k2.6', 'Commit-local + substrate push + open DRAFT PR (coder_commit/push/open_pr).')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('code-pr', 'clone',     'researched', 'Repo cloned into the worktree + surveyed.'),
    ('code-pr', 'plan',      'planned',    'Implementation plan ready, grounded in the real repo.'),
    ('code-pr', 'implement', 'executing',  'Change written + iterated to a green build/test in the cloned repo.'),
    ('code-pr', 'verify',    'verified',   'Build/test independently re-run green.'),
    ('code-pr', 'pr',        'verified',   'Branch pushed + DRAFT PR opened; awaiting the human merge (the Hinge).')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;
