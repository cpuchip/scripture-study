-- =====================================================================
-- cv8 (2026-06-04) — code-pr "ward council" model roster.
--
-- The bake-off measured each model's gift; assign the three stages where the
-- gift matters (the mechanical stages — clone/verify/pr — gain nothing from a
-- special model, so they stay on the fast/lean default):
--   plan      -> minimax-m3   (the architect/documenter: best docs, 1M context,
--                              careful structure; one tools-off call, so its
--                              slowness is acceptable)
--   implement -> kimi-k2.6    (the builder: measured fastest + leanest + cleanest;
--                              already the default — unchanged)
--   review    -> qwen3.7-max  (the critic: constant fresh judge; already set by
--                              cv6 — unchanged)
--   clone/verify/pr -> kimi-k2.6 (mechanical tool-calls; unchanged)
--
-- Dispatch resolves model from the stages-jsonb `model` field (work_item_advance
-- / dispatch COALESCE: model_override -> stage.model -> ...), so the authoritative
-- change is stages[1].model. stage_models is updated too for consistency/display.
-- One stage changes (plan). Idempotent.
-- =====================================================================

-- plan stage (index 1) -> minimax-m3.
UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{1,model}', '"minimax-m3"')
 WHERE family = 'code-pr'
   AND (stages->1->>'name') = 'plan';

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('code-pr', 'plan', 'minimax-m3', 'Council roster: the architect/documenter — m3 for its care + 1M context. One tools-off call per run.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;
