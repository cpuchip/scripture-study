-- =====================================================================
-- Batch J.10 — Gemini + expanded opencode_go model pricing
-- =====================================================================
-- Adds model_pricing rows so the substrate can cost-track + surface
-- (via /api/models, the /models view, and the Brainstorm datalist) the
-- models Michael asked for on 2026-05-29:
--   * Google Gemini chat models (provider already loaded; base_url fixed
--     in .env this session — appended /openai for the OpenAI-compat path)
--   * Additional opencode_go chat models, incl. DeepSeek
--
-- Prices are per 1M tokens in micro-dollars (1 USD = 1_000_000), pulled
-- 2026-05-29 from:
--   - opencode_go: https://opencode.ai/docs/zen/#pricing
--   - gemini:      https://cloud.google.com/.../generative-ai/pricing#standard
--                  (cross-checked against ai.google.dev/gemini-api/docs/pricing)
--
-- Fixed effective_at so re-running the migration updates in place
-- (model_pricing PK is provider, model, effective_at; most-recent wins).
--
-- Tiered Gemini models (2.5-pro, 3.x-pro): the schema carries a single
-- input/output rate, so we use the <=200k-token tier and record the
-- >200k tier in notes. Substrate calls rarely exceed 200k; when they do
-- (context-engine corpora), input cost is mildly under-counted — flagged.
--
-- Gemini cache columns left NULL: the OpenAI-compat surface does not
-- expose Gemini's context-cache token categories the way the native
-- API does. compute_cost handles NULL cache fine.
--
-- NOT priced here (no published rate on opencode's pricing table as of
-- 2026-05-29): deepseek-v4-pro, mimo-v2-pro, mimo-v2.5-pro, mimo-v2-omni,
-- hy3-preview. These still DISPATCH if named explicitly; compute_cost
-- flags 'no_pricing_row(opencode_go/<model>)' honestly until a rate is
-- published. Do not invent a 0 rate for them — 0 would silently
-- under-track real spend.
-- =====================================================================

INSERT INTO stewards.model_pricing
    (provider, model, input_micro_per_mtok, output_micro_per_mtok,
     cache_write_micro_per_mtok, cache_read_micro_per_mtok, effective_at, notes)
VALUES
    -- -----------------------------------------------------------------
    -- Google Gemini (OpenAI-compat; bare model ids; cache NULL)
    -- -----------------------------------------------------------------
    ('google_gemini', 'gemini-3.5-flash',        1500000,  9000000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier'),
    ('google_gemini', 'gemini-3-pro-preview',    2000000, 12000000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier; input >200k = $4.00/M'),
    ('google_gemini', 'gemini-3.1-pro-preview',  2000000, 12000000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier; input >200k = $4.00/M'),
    ('google_gemini', 'gemini-3-flash-preview',   500000,  3000000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier'),
    ('google_gemini', 'gemini-3.1-flash-lite',    250000,  1500000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier'),
    ('google_gemini', 'gemini-2.5-pro',          1250000, 10000000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier; >200k = $2.50/$15.00 per M'),
    ('google_gemini', 'gemini-2.5-flash',         300000,  2500000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier'),
    ('google_gemini', 'gemini-2.5-flash-lite',    100000,   400000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'standard tier'),
    ('google_gemini', 'gemini-2.0-flash',         100000,   400000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'token-based standard tier'),
    ('google_gemini', 'gemini-2.0-flash-lite',     75000,   300000, NULL, NULL,
     '2026-05-29 00:00:00+00', 'token-based standard tier'),

    -- -----------------------------------------------------------------
    -- opencode_go — additional chat models (cache_read exposed; cache_write NULL)
    -- -----------------------------------------------------------------
    ('opencode_go', 'deepseek-v4-flash',          0,        0,        NULL,      0,
     '2026-05-29 00:00:00+00', 'FREE (limited-time per opencode zen; confirm when GA)'),
    ('opencode_go', 'qwen3.7-max',          2500000,  7500000,        NULL, 500000,
     '2026-05-29 00:00:00+00', 'Anthropic-FORMAT model (api_format=anthropic in model_capability). USABLE via the substrate''s AN.2 /messages dispatch path (x-api-key + anthropic-version), added 2026-05-30. Reasoning model — thinking captured to reasoning_content. Price confirmed 2026-05-30.'),
    ('opencode_go', 'qwen3.5-plus',          200000,  1200000,        NULL,  20000,
     '2026-05-29 00:00:00+00', ''),
    ('opencode_go', 'glm-5',                1000000,  3200000,        NULL, 200000,
     '2026-05-29 00:00:00+00', 'Reasoning model (backend frank/GLM-5.1). Streams content fine via the substrate (auto-probe verified 2026-05-29). Give adequate per-call max_tokens for substantive prompts so reasoning does not exhaust the budget before content.'),
    ('opencode_go', 'kimi-k2.5',             600000,  3000000,        NULL, 100000,
     '2026-05-29 00:00:00+00', ''),
    ('opencode_go', 'minimax-m2.5',          300000,  1200000,        NULL,  60000,
     '2026-05-29 00:00:00+00', ''),
    ('opencode_go', 'mimo-v2.5',                   0,        0,        NULL,      0,
     '2026-05-29 00:00:00+00', 'FREE per opencode zen')
ON CONFLICT (provider, model, effective_at) DO UPDATE
SET input_micro_per_mtok       = EXCLUDED.input_micro_per_mtok,
    output_micro_per_mtok      = EXCLUDED.output_micro_per_mtok,
    cache_write_micro_per_mtok = EXCLUDED.cache_write_micro_per_mtok,
    cache_read_micro_per_mtok  = EXCLUDED.cache_read_micro_per_mtok,
    notes                      = EXCLUDED.notes;

-- =====================================================================
-- Acceptance:
--   SELECT provider, count(*) FROM stewards.model_pricing
--    WHERE effective_at = '2026-05-29 00:00:00+00' GROUP BY provider;
--   Expected: google_gemini=10, opencode_go=7
--
--   SELECT * FROM stewards.compute_cost('google_gemini','gemini-2.5-flash',1000000,500000);
--   Expected: 300000 + 1250000 = 1550000  ($1.55)
-- =====================================================================
