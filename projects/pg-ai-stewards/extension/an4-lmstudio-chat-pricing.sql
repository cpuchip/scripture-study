-- =====================================================================
-- Batch AN.4 — LM Studio local chat models in the catalog ($0)
-- =====================================================================
-- LM Studio chat already works via the EXISTING OpenAI-compat path: the
-- provider `lm_studio` (kind=openai, base http://host.docker.internal:1234/v1,
-- default qwen/qwen3.6-27b) is registered + reachable, and an auto-probe of
-- qwen/qwen3.6-27b returned usable/finish=stop (2026-05-30). No code change.
--
-- This just adds model_pricing rows so LM Studio chat models appear in the
-- catalog (list_models / model_catalog) and cost-track at $0 — they run on
-- Michael's own hardware (dual 4090s), so there is no API charge.
--
-- NOTE: the served set depends on what's loaded in LM Studio; these are the
-- chat models present 2026-05-30. Unloading one leaves a harmless stale row;
-- the auto-probe (M.5) will mark it unusable on its next pass if it 404s.
-- =====================================================================

INSERT INTO stewards.model_pricing
    (provider, model, input_micro_per_mtok, output_micro_per_mtok,
     cache_write_micro_per_mtok, cache_read_micro_per_mtok, effective_at, notes)
VALUES
    ('lm_studio', 'qwen/qwen3.6-27b',            0, 0, NULL, 0, '2026-05-30 00:00:00+00',
     'Local chat model on host LM Studio (dual 4090s). $0 — own hardware. Provider default; auto-probe usable 2026-05-30.'),
    ('lm_studio', 'unsloth/qwen3.6-27b',         0, 0, NULL, 0, '2026-05-30 00:00:00+00',
     'Local chat model on host LM Studio. $0 — own hardware. Available when loaded.'),
    ('lm_studio', 'nvidia/nemotron-3-nano-omni', 0, 0, NULL, 0, '2026-05-30 00:00:00+00',
     'Local chat model on host LM Studio. $0 — own hardware. Available when loaded.')
ON CONFLICT (provider, model, effective_at) DO UPDATE
SET input_micro_per_mtok       = EXCLUDED.input_micro_per_mtok,
    output_micro_per_mtok      = EXCLUDED.output_micro_per_mtok,
    cache_write_micro_per_mtok = EXCLUDED.cache_write_micro_per_mtok,
    cache_read_micro_per_mtok  = EXCLUDED.cache_read_micro_per_mtok,
    notes                      = EXCLUDED.notes;

-- =====================================================================
-- Acceptance (AN.4): lm_studio appears in list_connectors with model_count>=3,
-- and qwen/qwen3.6-27b shows usable in model_catalog.
-- =====================================================================
