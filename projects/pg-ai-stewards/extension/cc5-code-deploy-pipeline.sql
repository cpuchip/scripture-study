-- =====================================================================
-- cc5 (2026-06-03) — the `code-deploy` pipeline + the always-escalate Hinge.
-- substrate-coding-capability CC.5 (v1: deploy to a local sidecar).
--
-- THE HINGE: the `prepare` stage has auto_advance=FALSE. After the agent
-- prepares the deploy (builds the artifact, proposes run_command/port/
-- health_path), the work_item goes to `awaiting_review` and STOPS. A human
-- must ratify before the `deploy` stage runs. Deploy is outward-facing and
-- not a cheap walk-back, so it always escalates regardless of trust — this
-- is the always-escalate rung the delegation-pattern audit proposed
-- (docs/delegation-pattern-skills-and-gates.md), now built. Exodus 18:22
-- made literal: small matters the gate judges, great matters come to the human.
--
-- Input requires input.sandbox (the sandbox id of a verified code-write).
-- coder_deploy runs the artifact as a background service in that sandbox
-- (the sandbox IS its docker sidecar) + healthchecks it. v2 (CC.7) reaches
-- Dokploy with scoped access.
-- =====================================================================

INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder
)
VALUES (
    'code-deploy',
    jsonb_build_array(
        jsonb_build_object(
            'name',           'prepare',
            'next',           'deploy',
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   false,   -- <<< THE HINGE: stop for human ratification before deploy
            'tools_disabled', false,
            'input_template',
                'Deploy task: {{input.binding_question}}' || E'\n\n' ||
                'Sandbox (build+test already passed): {{input.sandbox}}' || E'\n\n' ||
                'Prepare this code for deployment — do NOT deploy yet:' || E'\n' ||
                '1. Inspect the sandbox (coder_glob / coder_read) to see what was built.' || E'\n' ||
                '2. If a build step is needed (e.g. `go build -o app .`), run it via coder_shell, sandbox="{{input.sandbox}}".' || E'\n' ||
                '3. Determine how to run it as a service: the run_command (from /work), the TCP port it listens on, and an HTTP health_path.' || E'\n\n' ||
                'Report the DEPLOY PLAN clearly and explicitly: the exact run_command, the port, and the health_path. ' ||
                'A human reviews and ratifies this plan before the deploy runs — this is the Hinge; the deploy never fires on its own.'
        ),
        jsonb_build_object(
            'name',           'deploy',
            'next',           NULL,
            'model',          'kimi-k2.6',
            'provider',       'opencode_go',
            'agent_family',   'dev',
            'auto_advance',   true,
            'tools_disabled', false,
            'input_template',
                'Deploy task: {{input.binding_question}}' || E'\n\n' ||
                'Sandbox: {{input.sandbox}}' || E'\n\n' ||
                'The deploy plan (ratified by a human):' || E'\n' || '{{stage_results.prepare.output}}' || E'\n\n' ||
                'Execute the deploy now: call coder_deploy with sandbox="{{input.sandbox}}" and the run_command, ' ||
                'port, and health_path from the ratified plan. Report whether the service came up healthy, the ' ||
                'healthcheck result, and the service log tail.'
        )
    ),
    false,  -- sabbath_enabled
    true,   -- atonement_enabled
    NULL,
    NULL,
    '["raw","planned","verified"]'::jsonb   -- prepare->planned, deploy->verified (reuse valid rungs)
)
ON CONFLICT (family) DO UPDATE SET
    stages                    = EXCLUDED.stages,
    sabbath_enabled           = EXCLUDED.sabbath_enabled,
    atonement_enabled         = EXCLUDED.atonement_enabled,
    file_destination_template = EXCLUDED.file_destination_template,
    file_content_jsonpath     = EXCLUDED.file_content_jsonpath,
    maturity_ladder           = EXCLUDED.maturity_ladder;

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-deploy', 'prepare', 'kimi-k2.6', 'Build artifact + propose run_command/port/health_path. THE HINGE: auto_advance=false.'),
    ('code-deploy', 'deploy',  'kimi-k2.6', 'Run the artifact in its sandbox sidecar + healthcheck (coder_deploy).')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('code-deploy', 'prepare', 'planned',  'Deploy plan ready; awaiting human ratification (the Hinge).'),
    ('code-deploy', 'deploy',  'verified', 'Deployed to the sandbox sidecar + healthchecked.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;

-- coder_deploy grant
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source) VALUES
  ('dev', 'coder_deploy', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO NOTHING;
