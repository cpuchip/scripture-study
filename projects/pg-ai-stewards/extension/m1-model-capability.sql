-- =====================================================================
-- Batch M.1 — Model capability registry
-- =====================================================================
-- The substrate had no way to know which catalogued models it can
-- actually dispatch. The 2026-05-29 brainstorm run picked qwen3.7-max
-- (gateway rejects it: "not supported for format oa-compat") and glm-5
-- (a reasoning model whose content never arrives over the streaming path
-- the substrate uses) — both came back empty, diagnosed by hand via a
-- gateway probe (extension/smoke/test-glm-qwen-models.sh).
--
-- This batch makes that knowledge first-class:
--   - model_capability — per (provider, model) usable / streaming signal
--   - model_usable()   — the gate predicate (unknown defaults to usable)
--   - model_catalog    — a view joining pricing + capability for tooling
--   - seed             — tonight's hand-verified results
--
-- M.2 wires model_usable() into the dispatch chokepoint (substitute-and-
-- log). M.4 adds an auto-probe so the signal stays current without a human
-- curling the gateway. Applied live (docker cp + psql -f) like every
-- post-am1 batch; persists in the data volume.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. model_capability — one row per (provider, model) we have a verdict on.
-- ---------------------------------------------------------------------
-- usable=false is the ONLY thing that gates dispatch. A model with no row
-- is treated as usable (innocent until proven guilty) — mirrors the J.11
-- cap gate, where a provider with no cap row is never gated. This keeps
-- every working, un-probed model dispatching exactly as it does today.
CREATE TABLE IF NOT EXISTS stewards.model_capability (
    provider           text NOT NULL,
    model              text NOT NULL,
    usable             boolean NOT NULL DEFAULT true,
    supports_streaming boolean,            -- NULL = not yet determined
    last_probed_at     timestamptz,
    probe_detail       text,               -- the error, or a short 'ok' note
    probed_via         text NOT NULL DEFAULT 'seed',  -- seed | manual | auto-probe
    updated_at         timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, model)
);

COMMENT ON TABLE stewards.model_capability IS
'Batch M.1: per-model dispatchability signal. usable=false gates the model in work_item_dispatch_stage (M.2, substitute-and-log). A model with no row defaults to usable (model_usable()). supports_streaming isolates the GLM failure axis: content arrives non-streaming but not over the substrate''s streaming path. Kept current by the M.4 auto-probe.';

COMMENT ON COLUMN stewards.model_capability.supports_streaming IS
'M.1: whether content arrives over the streaming path the substrate dispatches with (stream:true, required since ES.6). GLM-5/5.1 are false here despite working non-streaming.';

COMMENT ON COLUMN stewards.model_capability.probed_via IS
'M.1: seed (hand-verified), manual (probe tool), or auto-probe (bgworker, M.4).';


-- ---------------------------------------------------------------------
-- 2. model_usable(provider, model) -> boolean — the gate predicate.
-- ---------------------------------------------------------------------
-- False ONLY when an explicit row says usable=false. Unknown -> true.
CREATE OR REPLACE FUNCTION stewards.model_usable(p_provider text, p_model text)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT COALESCE(
        (SELECT usable
           FROM stewards.model_capability
          WHERE provider = p_provider AND model = p_model),
        true
    );
$$;

COMMENT ON FUNCTION stewards.model_usable(text, text) IS
'Batch M.1: true unless model_capability explicitly marks (provider, model) usable=false. Unknown models default to usable so existing dispatch is never broken. The substitution gate in work_item_dispatch_stage (M.2) consults this.';


-- ---------------------------------------------------------------------
-- 3. first_usable_model(provider) -> text — substitution target helper.
-- ---------------------------------------------------------------------
-- Used by M.2 when the catalog default itself is unusable: returns any
-- model for the provider that is priced AND not marked unusable, cheapest
-- output rate first (so a forced substitution doesn't surprise the bill).
CREATE OR REPLACE FUNCTION stewards.first_usable_model(p_provider text)
RETURNS text LANGUAGE sql STABLE AS $$
    SELECT mp.model
      FROM (
          SELECT DISTINCT ON (provider, model) provider, model, output_micro_per_mtok
            FROM stewards.model_pricing
           ORDER BY provider, model, effective_at DESC
      ) mp
     WHERE mp.provider = p_provider
       AND stewards.model_usable(mp.provider, mp.model)
     ORDER BY mp.output_micro_per_mtok ASC NULLS LAST
     LIMIT 1;
$$;

COMMENT ON FUNCTION stewards.first_usable_model(text) IS
'Batch M.1: cheapest priced + usable model for a provider, or NULL if none. M.2 substitution fallback when the catalog default is itself unusable.';


-- ---------------------------------------------------------------------
-- 4. model_catalog view — pricing + capability, one row per model.
-- ---------------------------------------------------------------------
-- Latest pricing row per (provider, model) left-joined to its capability
-- verdict. Backs the list_models MCP tool (M.3) and human catalog reads.
CREATE OR REPLACE VIEW stewards.model_catalog AS
SELECT
    mp.provider,
    mp.model,
    mp.input_micro_per_mtok,
    mp.output_micro_per_mtok,
    mp.notes                       AS pricing_notes,
    COALESCE(mc.usable, true)      AS usable,
    mc.supports_streaming,
    mc.last_probed_at,
    mc.probe_detail,
    COALESCE(mc.probed_via, 'unprobed') AS probed_via
FROM (
    SELECT DISTINCT ON (provider, model)
           provider, model, input_micro_per_mtok, output_micro_per_mtok, notes
      FROM stewards.model_pricing
     ORDER BY provider, model, effective_at DESC
) mp
LEFT JOIN stewards.model_capability mc
       ON mc.provider = mp.provider AND mc.model = mp.model;

COMMENT ON VIEW stewards.model_catalog IS
'Batch M.1: latest pricing per (provider, model) joined to capability verdict. usable defaults true for un-probed models. Backs the list_models MCP tool.';


-- ---------------------------------------------------------------------
-- 5. Seed — tonight's hand-verified results (2026-05-29 gateway probe).
-- ---------------------------------------------------------------------
-- Corrected 2026-05-29 by the M.4 auto-probe (see commit history): my initial
-- hand-diagnosis marked glm-5/glm-5.1 unusable on the strength of a shell-grep
-- streaming test that returned 0 content chars. The substrate's real SSE parser
-- (parse_chat_sse) extracts GLM's content fine — an auto-probe with a
-- substantive prompt returned 385 chars, finish=stop. So GLM streams fine; the
-- brainstorm-run emptiness was a per-lens budget/transient issue, NOT a
-- streaming incompatibility. The only verified-unusable opencode model is
-- qwen3.7-max (the substrate's streaming dispatch gets HTTP 401; a direct curl
-- also reports "not supported for format oa-compat"). Everything unrowed
-- defaults usable until the M.4 auto-probe verifies it.
INSERT INTO stewards.model_capability
    (provider, model, usable, supports_streaming, last_probed_at, probe_detail, probed_via)
VALUES
    -- --- unusable (verified via the real dispatch path) ---
    ('opencode_go', 'qwen3.7-max', false, false, now(),
     'Unusable via substrate: dispatch returns HTTP 401 whose body is "Model qwen3.7-max is not supported for format oa-compat" — the gateway rejects this model on the OpenAI-compat endpoint and expresses it as a 401. Consistent across probes.', 'seed'),
    -- --- usable (verified streaming via the substrate path) ---
    ('opencode_go', 'glm-5', true, true, now(),
     'Backend frank/GLM-5.1 (reasoning model). Streams content fine via the substrate (auto-probe: 385 chars, finish=stop). Give adequate per-call max_tokens for substantive prompts so reasoning does not exhaust the budget.', 'seed'),
    ('opencode_go', 'glm-5.1', true, true, now(),
     'Backend frank/GLM-5.1 (reasoning model). Streams content fine via the substrate (auto-probe verified). Give adequate per-call max_tokens for substantive prompts.', 'seed'),
    -- --- usable (direct evidence on the 2026-05-29 v2 brainstorm run) ---
    ('opencode_go', 'kimi-k2.6',         true, true, now(), 'Substrate main chain; streams reliably.', 'seed'),
    ('opencode_go', 'kimi-k2.5',         true, true, now(), 'Re-fired empties successfully on the v2 run.', 'seed'),
    ('opencode_go', 'deepseek-v4-flash', true, true, now(), 'FREE; streamed reliably on the v2 run. Good fan-out default.', 'seed'),
    ('opencode_go', 'mimo-v2.5',         true, true, now(), 'FREE; streamed reliably on the v2 run. Good fan-out default.', 'seed'),
    ('opencode_go', 'qwen3.6-plus',      true, true, now(), 'Ran SCAMPER/Crazy8s lenses on the v2 run.', 'seed')
ON CONFLICT (provider, model) DO UPDATE
SET usable             = EXCLUDED.usable,
    supports_streaming = EXCLUDED.supports_streaming,
    last_probed_at     = EXCLUDED.last_probed_at,
    probe_detail       = EXCLUDED.probe_detail,
    probed_via         = EXCLUDED.probed_via,
    updated_at         = now();


-- =====================================================================
-- Acceptance (verify before commit):
--   1. model_usable('opencode_go','glm-5')        = true   (streams fine)
--   2. model_usable('opencode_go','qwen3.7-max')  = false
--   3. model_usable('opencode_go','kimi-k2.6')    = true
--   4. model_usable('opencode_go','never-heard')  = true   (unknown defaults usable)
--   5. first_usable_model('opencode_go') returns a cheap usable model
--      (a free one: deepseek-v4-flash or mimo-v2.5), never qwen3.7-max.
--   6. SELECT count(*) FROM stewards.model_catalog WHERE NOT usable; = 1
--      (qwen3.7-max only)
-- =====================================================================
