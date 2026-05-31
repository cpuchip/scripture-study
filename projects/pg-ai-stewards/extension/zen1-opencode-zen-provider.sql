-- =====================================================================
-- Batch ZEN.1 — opencode_zen provider: real Claude (pay-per-use) + $18 cap
-- =====================================================================
-- opencode has TWO tiers on the SAME key:
--   go subscription  -> https://opencode.ai/zen/go/v1  (provider opencode_go)
--   zen pay-per-use  -> https://opencode.ai/zen/v1      (provider opencode_zen, NEW)
-- The real Anthropic Claude models live on the zen tier (/zen/v1/messages,
-- Anthropic format — which the AT-batch already dispatches). Confirmed live
-- 2026-05-31: /zen/v1/models lists claude-opus-4-8/4-7/4-6, claude-sonnet-4-6,
-- claude-haiku-4-5 (+ gpt-5.x, gemini-3.x, and -free variants). A free model
-- (deepseek-v4-flash-free) streams cost:"0".
--
-- The opencode_zen PROVIDER itself is registered from env
-- (STEWARDS_PROVIDER_OPENCODE_ZEN_* — added to .env; pg recreated to load it;
-- the registry parses STEWARDS_PROVIDER_<NAME>_* dynamically). This file adds
-- the catalog rows, the api_format verdicts, and the ENFORCED $18 spend cap.
--
-- Cost note: pay-per-use real money. The $18 enforced cap (Michael's ~$20 zen
-- balance, gemini-pattern buffer) refuses dispatch once spend-since-refill hits
-- it. opus-4.8 = $5/$25 per Mtok — the cap is the runaway guard.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Pricing (zen pay-per-use rates, confirmed from opencode docs 2026-05-31).
-- ---------------------------------------------------------------------
INSERT INTO stewards.model_pricing
    (provider, model, input_micro_per_mtok, output_micro_per_mtok,
     cache_write_micro_per_mtok, cache_read_micro_per_mtok, effective_at, notes)
VALUES
    ('opencode_zen', 'claude-opus-4-8',   5000000, 25000000, 6250000, 500000, '2026-05-31 00:00:00+00', 'Anthropic Claude Opus 4.8 (zen pay-per-use, /messages). api_format=anthropic.'),
    ('opencode_zen', 'claude-opus-4-7',   5000000, 25000000, 6250000, 500000, '2026-05-31 00:00:00+00', 'Anthropic Claude Opus 4.7 (zen pay-per-use).'),
    ('opencode_zen', 'claude-opus-4-6',   5000000, 25000000, 6250000, 500000, '2026-05-31 00:00:00+00', 'Anthropic Claude Opus 4.6 (zen pay-per-use).'),
    ('opencode_zen', 'claude-sonnet-4-6', 3000000, 15000000, 3750000, 300000, '2026-05-31 00:00:00+00', 'Anthropic Claude Sonnet 4.6 (zen pay-per-use).'),
    ('opencode_zen', 'claude-haiku-4-5',  1000000,  5000000, 1250000, 100000, '2026-05-31 00:00:00+00', 'Anthropic Claude Haiku 4.5 (zen pay-per-use). Cheapest claude — good default.'),
    ('opencode_zen', 'deepseek-v4-flash-free', 0, 0, NULL, 0, '2026-05-31 00:00:00+00', 'FREE on zen (streams cost:"0", verified 2026-05-31). OpenAI-format. The zen tier also serves mimo-v2.5-free, minimax-m2.5-free, qwen3.6-plus-free, nemotron-3-super-free, big-pickle — add after per-model format probing.')
ON CONFLICT (provider, model, effective_at) DO UPDATE
SET input_micro_per_mtok       = EXCLUDED.input_micro_per_mtok,
    output_micro_per_mtok      = EXCLUDED.output_micro_per_mtok,
    cache_write_micro_per_mtok = EXCLUDED.cache_write_micro_per_mtok,
    cache_read_micro_per_mtok  = EXCLUDED.cache_read_micro_per_mtok,
    notes                      = EXCLUDED.notes;

-- ---------------------------------------------------------------------
-- 2. Capability + api_format. Claude = anthropic format (AT dispatch path);
--    deepseek-v4-flash-free = openai (verified). usable seeded true; the
--    M.4 auto-probe re-verifies on the watchman cadence.
-- ---------------------------------------------------------------------
INSERT INTO stewards.model_capability
    (provider, model, usable, supports_streaming, api_format, last_probed_at, probe_detail, probed_via)
VALUES
    ('opencode_zen', 'claude-opus-4-8',   true, true, 'anthropic', now(), 'Claude Opus 4.8 via zen /messages.', 'seed'),
    ('opencode_zen', 'claude-opus-4-7',   true, true, 'anthropic', now(), 'Claude Opus 4.7 via zen /messages.', 'seed'),
    ('opencode_zen', 'claude-opus-4-6',   true, true, 'anthropic', now(), 'Claude Opus 4.6 via zen /messages.', 'seed'),
    ('opencode_zen', 'claude-sonnet-4-6', true, true, 'anthropic', now(), 'Claude Sonnet 4.6 via zen /messages.', 'seed'),
    ('opencode_zen', 'claude-haiku-4-5',  true, true, 'anthropic', now(), 'Claude Haiku 4.5 via zen /messages.', 'seed'),
    ('opencode_zen', 'deepseek-v4-flash-free', true, true, 'openai', now(), 'FREE zen model; streams cost:"0" (verified).', 'seed')
ON CONFLICT (provider, model) DO UPDATE
SET usable = EXCLUDED.usable, supports_streaming = EXCLUDED.supports_streaming,
    api_format = EXCLUDED.api_format, probe_detail = EXCLUDED.probe_detail, probed_via = EXCLUDED.probed_via;

-- ---------------------------------------------------------------------
-- 3. Enforced prepaid spend cap — $18 (Michael's ~$20 zen balance, gemini pattern).
-- ---------------------------------------------------------------------
INSERT INTO stewards.provider_spend_caps (provider, cap_micro, since, enforced, notes)
VALUES ('opencode_zen', 18000000, now(), true,
    'Prepaid pay-per-use cap — $18 of ~$20 zen balance (buffer). opus-4.8=$5/$25 per Mtok; this refuses dispatch once spend-since-refill hits $18. Top up + reset: SELECT stewards.provider_cap_refill(''opencode_zen''[, <new_cap_micro>]);')
ON CONFLICT (provider) DO UPDATE
SET cap_micro = EXCLUDED.cap_micro, enforced = EXCLUDED.enforced, notes = EXCLUDED.notes, updated_at = now();

-- =====================================================================
-- Acceptance (ZEN.1, after env + recreate):
--   1. provider opencode_zen registered (providers_loaded shows it post-recreate).
--   2. probe claude-haiku-4-5 (opencode_zen) -> usable, real content via /messages.
--   3. probe deepseek-v4-flash-free (opencode_zen) -> usable, cost $0.
--   4. provider_cap_exceeded('opencode_zen') = false initially ($0 < $18).
-- =====================================================================
