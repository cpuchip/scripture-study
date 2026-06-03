-- =====================================================================
-- cc3 (2026-06-03) — the `code-write` pipeline (substrate-coding-capability
-- CC.3, path A: the agent-driven ground-truth loop).
--
-- Three stages, modeled on research-write (h1-2):
--   plan      (dev, tools off)  → produces maturity `planned`
--   implement (dev, tools ON)   → produces `executing` — the coder agent
--                                  starts the sandbox, writes code, and
--                                  iterates `coder_shell` build+test to GREEN
--                                  (real exit codes = ground truth)
--   verify    (dev, tools ON)   → produces `verified` — independently re-runs
--                                  build+test in the same sandbox; REVIEW:
--                                  passes / fail drives the gate's advance/revise
--
-- D-CC7: the loop runs free; the trust ladder gates only at the PR rung
-- (which arrives with git push, a later phase). CC.3.1 (path B) will add a
-- deterministic substrate gate (tool_dispatch build+test, no LLM) as a
-- co-usable path.
--
-- Sandbox id: a BEFORE-INSERT trigger stamps input.sandbox = wi-<8> so the
-- implement + verify stages share one stable sandbox across the revise loop.
-- The coder tools are already granted to `dev` (cc2).
-- =====================================================================

-- --- stable sandbox id per code-write work_item -----------------------
CREATE OR REPLACE FUNCTION stewards.stamp_code_write_sandbox()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    IF NEW.pipeline_family = 'code-write'
       AND (NEW.input IS NULL OR (NEW.input->>'sandbox') IS NULL)
    THEN
        NEW.input := COALESCE(NEW.input, '{}'::jsonb)
            || jsonb_build_object('sandbox', 'wi-' || substring(NEW.id::text FROM 1 FOR 8));
    END IF;
    RETURN NEW;
END;
$func$;

DROP TRIGGER IF EXISTS trg_stamp_code_write_sandbox ON stewards.work_items;
CREATE TRIGGER trg_stamp_code_write_sandbox
    BEFORE INSERT ON stewards.work_items
    FOR EACH ROW EXECUTE FUNCTION stewards.stamp_code_write_sandbox();

-- --- the pipeline -----------------------------------------------------
INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder
)
VALUES (
    'code-write',
    jsonb_build_array(
        jsonb_build_object(
            'name',           'plan',
            'next',           'implement',
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', true,
            'input_template',
                'Coding task (binding question): {{input.binding_question}}' || E'\n\n' ||
                'Produce a concise implementation plan, NOT code:' || E'\n' ||
                '  - The files to create or change (paths relative to the project root).' || E'\n' ||
                '  - The approach in a few sentences.' || E'\n' ||
                '  - The exact build + test command that will prove it works ' ||
                '(e.g. `go build ./... && go test ./...`, or `npm ci && npm test`). ' ||
                'This command is the ground-truth gate; choose it deliberately.' || E'\n\n' ||
                'Keep it tight. The next stage implements against this plan in a sandbox.'
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
                'Your sandbox id is: {{input.sandbox}}' || E'\n\n' ||
                'Implement it in the sandbox using the coder tools:' || E'\n' ||
                '1. coder_sandbox_start with sandbox="{{input.sandbox}}" (reuses the sandbox if it already exists).' || E'\n' ||
                '2. Write the code with coder_write / coder_edit (paths are relative to /work, the project root).' || E'\n' ||
                '3. Build and test with coder_shell, sandbox="{{input.sandbox}}", running the build+test command from the plan.' || E'\n' ||
                '4. ITERATE: if the build or tests fail, read the real output, fix the code, and run again. ' ||
                'Do NOT stop until the build+test command exits 0 (green). The passing build+test is your done ' ||
                'condition — it is ground truth, not a judgment call.' || E'\n\n' ||
                'When green, report: what you built, the files written, and paste the final passing build+test output.'
        ),
        jsonb_build_object(
            'name',           'verify',
            'next',           NULL,
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', false,
            'input_template',
                'Coding task: {{input.binding_question}}' || E'\n\n' ||
                'The implement stage reported:' || E'\n' || '{{stage_results.implement.output}}' || E'\n\n' ||
                'Independently verify — do NOT trust the report above. In sandbox "{{input.sandbox}}":' || E'\n' ||
                '1. coder_shell, sandbox="{{input.sandbox}}", run the build + test command yourself.' || E'\n' ||
                '2. Inspect the REAL exit code and output.' || E'\n\n' ||
                'Return EXACTLY one of:' || E'\n' ||
                '  (a) A first line "REVIEW: passes" (only if the command exited 0), then the build/test output.' || E'\n' ||
                '  (b) A first line "REVIEW: fail", then the failing output and a short note on what still needs fixing.'
        )
    ),
    false,  -- sabbath_enabled: code is mechanical; skip sabbath for v1
    true,   -- atonement_enabled: a code run that hits a cost cap is worth atoning over
    NULL,   -- file_destination_template: deliverable is code in the sandbox (+ later a PR), not one markdown file
    NULL,
    '["raw","planned","executing","verified"]'::jsonb
)
ON CONFLICT (family) DO UPDATE SET
    stages                    = EXCLUDED.stages,
    sabbath_enabled           = EXCLUDED.sabbath_enabled,
    atonement_enabled         = EXCLUDED.atonement_enabled,
    file_destination_template = EXCLUDED.file_destination_template,
    file_content_jsonpath     = EXCLUDED.file_content_jsonpath,
    maturity_ladder           = EXCLUDED.maturity_ladder;

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-write', 'plan',      'kimi-k2.6', 'Implementation plan; tools off.'),
    ('code-write', 'implement', 'kimi-k2.6', 'Write + build/test loop in the sandbox; coder tools on. Model tunable.'),
    ('code-write', 'verify',    'kimi-k2.6', 'Independent build/test re-run; coder tools on.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('code-write', 'plan',      'planned',   'Implementation plan ready.'),
    ('code-write', 'implement', 'executing', 'Code written + iterated to a green build/test in the sandbox.'),
    ('code-write', 'verify',    'verified',  'Build/test independently re-run green.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;
