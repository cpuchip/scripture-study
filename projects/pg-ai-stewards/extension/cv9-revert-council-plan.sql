-- =====================================================================
-- cv9 (2026-06-04) — revert cv8: code-pr plan stage back to kimi-k2.6.
--
-- The council A/B (m3-plan vs kimi-plan, same chatcore task, only plan model
-- differs) showed m3-on-plan cost ~2x time + ~1.75x tokens + 33% more code +
-- MORE implement iterations (20 vs 15), with NO gate improvement (both passed
-- the critic first-try). "Careful architect -> easier build" did not hold.
-- kimi-solo is the efficient default; m3-on-plan is a premium per-task opt-in
-- (set work_item input or model_override when a task warrants the extra design),
-- not the standing roster. So: plan -> kimi-k2.6 by default again.
-- (review stays qwen3.7-max — the critic is the one differentiated stage that
--  earns its keep.)
-- =====================================================================

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{1,model}', '"kimi-k2.6"')
 WHERE family = 'code-pr'
   AND (stages->1->>'name') = 'plan';

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-pr', 'plan', 'kimi-k2.6', 'Default kimi-k2.6 (cv9 revert of cv8). The A/B showed m3-on-plan was ~2x cost for no gate gain; m3 is a per-task opt-in, not the default.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;
