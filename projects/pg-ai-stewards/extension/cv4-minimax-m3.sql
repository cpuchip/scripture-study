-- =====================================================================
-- cv4 (2026-06-03) — register minimax-m3 on opencode_go.
--
-- opencode_go's /models serves `minimax-m3` (the go subscription tier). A
-- direct /chat/completions probe returned content, so it is OPENAI-format
-- (unlike its sibling minimax-m2.7 which is anthropic-format; m2.5 is openai —
-- the minimax family is mixed, so we pin the format explicitly). It is a
-- REASONING model (emits <think>) with a 1M-token context window — the draw
-- for large-repo code-pr work. Reasoning models need a generous per-call
-- max_tokens so thinking does not exhaust the budget before content.
--
-- usable=true is set optimistically; the substrate auto-probe (M.5) / a manual
-- enqueue_model_probe verifies via the REAL streamed dispatch path and will
-- flip usable=false if it does not stream. Pricing mirrors the minimax-m2.x
-- family ($0.30/$1.20 per Mtok) as a cost-tracking estimate (opencode go-tier
-- per-token rate is unpublished; it is a subscription, so this only feeds the
-- cost buckets + cap math, not a real bill).
-- =====================================================================

INSERT INTO stewards.model_capability
    (provider, model, usable, supports_streaming, api_format,
     last_probed_at, probe_detail, probed_via, updated_at)
VALUES
    ('opencode_go', 'minimax-m3', true, true, 'openai',
     now(),
     'Manual register 2026-06-03: opencode_go /models lists minimax-m3; direct /chat/completions probe returned content (openai-format). Reasoning model (emits <think>), 1M-token context. Give generous per-call max_tokens so reasoning does not exhaust the budget before content. Auto-probe will refresh.',
     'manual', now())
ON CONFLICT (provider, model) DO UPDATE SET
    usable             = EXCLUDED.usable,
    supports_streaming = EXCLUDED.supports_streaming,
    api_format         = EXCLUDED.api_format,
    probe_detail       = EXCLUDED.probe_detail,
    probed_via         = EXCLUDED.probed_via,
    updated_at         = now();

INSERT INTO stewards.model_pricing
    (provider, model, input_micro_per_mtok, output_micro_per_mtok, effective_at, notes)
VALUES
    ('opencode_go', 'minimax-m3', 300000, 1200000, '2026-06-03 00:00:00+00',
     'opencode_go subscription (go tier); per-token rate unpublished — mirrors the minimax-m2.x family ($0.30/$1.20) as a cost-tracking estimate. 1M context, reasoning model.')
ON CONFLICT (provider, model, effective_at) DO UPDATE SET
    input_micro_per_mtok  = EXCLUDED.input_micro_per_mtok,
    output_micro_per_mtok = EXCLUDED.output_micro_per_mtok,
    notes                 = EXCLUDED.notes;
