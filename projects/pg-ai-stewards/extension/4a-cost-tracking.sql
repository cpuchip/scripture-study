-- =====================================================================
-- Phase 4a — Cost Tracking (D-A4 token-multiplier cost model)
--
-- Implements the cost-tracking schema from
-- projects/pg-ai-stewards/.spec/proposals/cost-tracking.md.
--
-- Builds on:
--   - 3e2-7-git-mcp-seed.sql (last block in lib.rs ordering)
--
-- This file adds:
--   1. stewards.model_pricing — per-model rate table (input/output/cache)
--   2. stewards.cost_events — append-only per-attempt cost ledger
--   3. stewards.cost_buckets — concentric tracking buckets (5h/daily/weekly/monthly)
--   4. work_items columns: cost_micro_dollars, cost_cap_micro, cost_capped_at
--   5. SQL functions: compute_cost, record_cost_event, cost_cap_exceeded
--   6. Bucket helpers: bucket_current, bucket_record, bucket_period_for
--   7. Trigger on cost_events to maintain work_items denormalized total
--   8. Seed data: real OpenCode Zen rates from opencode.ai/docs/zen
--   9. Seed data: bucket caps for OpenCode Go subscription ($12/day, $60/mo)
--
-- All money in micro-dollars (1 USD = 1_000_000) for integer arithmetic.
-- All rates per million tokens, so $0.95/MTok = 950000 micro-dollars/MTok.
-- =====================================================================

-- ---------------------------------------------------------------------
-- model_pricing: one row per (provider, model, effective_at)
-- Most-recent row whose effective_at <= now() wins.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.model_pricing (
    provider                    text  NOT NULL,
    model                       text  NOT NULL,
    input_micro_per_mtok        bigint NOT NULL CHECK (input_micro_per_mtok >= 0),
    output_micro_per_mtok       bigint NOT NULL CHECK (output_micro_per_mtok >= 0),
    cache_write_micro_per_mtok  bigint CHECK (cache_write_micro_per_mtok IS NULL OR cache_write_micro_per_mtok >= 0),
    cache_read_micro_per_mtok   bigint CHECK (cache_read_micro_per_mtok IS NULL OR cache_read_micro_per_mtok >= 0),
    effective_at                timestamptz NOT NULL DEFAULT now(),
    notes                       text,
    PRIMARY KEY (provider, model, effective_at)
);

COMMENT ON TABLE stewards.model_pricing IS
'Per-model pricing in micro-dollars per 1M tokens. NULL cache_*_micro_per_mtok means provider does not expose that distinction. Most-recent effective_at wins.';

-- ---------------------------------------------------------------------
-- cost_events: append-only per-dispatch cost audit ledger
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.cost_events (
    id                          bigserial PRIMARY KEY,
    work_item_id                uuid REFERENCES stewards.work_items(id) ON DELETE CASCADE,
    attempt_seq                 int NOT NULL,
    at                          timestamptz NOT NULL DEFAULT now(),
    provider                    text NOT NULL,
    model                       text NOT NULL,
    input_tokens                int NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
    output_tokens               int NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
    cache_write_tokens          int NOT NULL DEFAULT 0 CHECK (cache_write_tokens >= 0),
    cache_read_tokens           int NOT NULL DEFAULT 0 CHECK (cache_read_tokens >= 0),
    micro_dollars               bigint NOT NULL,
    pricing_effective_at        timestamptz NOT NULL,
    notes                       text
);
CREATE INDEX IF NOT EXISTS cost_events_work_item ON stewards.cost_events(work_item_id);
CREATE INDEX IF NOT EXISTS cost_events_at ON stewards.cost_events(at);
CREATE INDEX IF NOT EXISTS cost_events_provider_model ON stewards.cost_events(provider, model);

COMMENT ON TABLE stewards.cost_events IS
'Append-only audit of every LLM dispatch cost. micro_dollars is computed at insert from compute_cost(provider, model, tokens) and locked to pricing_effective_at.';

-- ---------------------------------------------------------------------
-- cost_buckets: concentric tracking periods per provider
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.cost_buckets (
    id                  bigserial PRIMARY KEY,
    provider            text NOT NULL,
    bucket_kind         text NOT NULL CHECK (bucket_kind IN ('session_5h','daily','weekly','monthly')),
    period_start        timestamptz NOT NULL,
    period_end          timestamptz NOT NULL,
    micro_dollars       bigint NOT NULL DEFAULT 0,
    bucket_limit_micro  bigint,  -- NULL = informational only, no enforcement
    notes               text,
    UNIQUE (provider, bucket_kind, period_start)
);
CREATE INDEX IF NOT EXISTS cost_buckets_period ON stewards.cost_buckets(provider, bucket_kind, period_end);

COMMENT ON TABLE stewards.cost_buckets IS
'Rolling consumption buckets per provider/kind. Closes at period_end; bucket_current() opens the next period lazily. bucket_limit_micro NULL means informational only (no enforcement).';

-- ---------------------------------------------------------------------
-- work_items: cost columns
-- ---------------------------------------------------------------------
ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS cost_micro_dollars  bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_cap_micro      bigint,
    ADD COLUMN IF NOT EXISTS cost_capped_at      timestamptz;

COMMENT ON COLUMN stewards.work_items.cost_micro_dollars IS
'Phase 4a: denormalized cumulative cost in micro-dollars. Maintained by trigger on cost_events insert.';
COMMENT ON COLUMN stewards.work_items.cost_cap_micro IS
'Phase 4a: per-work_item cost cap in micro-dollars. NULL = no cap.';
COMMENT ON COLUMN stewards.work_items.cost_capped_at IS
'Phase 4a: timestamp when cost_micro_dollars first crossed cost_cap_micro.';

-- =====================================================================
-- Functions
-- =====================================================================

-- ---------------------------------------------------------------------
-- compute_cost(provider, model, tokens...) -> (micro_dollars, pricing_effective_at)
-- Picks the most-recent pricing row whose effective_at <= now().
-- Returns (0, '-infinity'::timestamptz) if no pricing row exists.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.compute_cost(
    p_provider           text,
    p_model              text,
    p_input_tokens       int,
    p_output_tokens      int,
    p_cache_write_tokens int DEFAULT 0,
    p_cache_read_tokens  int DEFAULT 0
) RETURNS TABLE (micro_dollars bigint, pricing_effective_at timestamptz)
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_pricing record;
    v_micro bigint;
BEGIN
    SELECT * INTO v_pricing
      FROM stewards.model_pricing
     WHERE provider = p_provider
       AND model = p_model
       AND effective_at <= now()
     ORDER BY effective_at DESC
     LIMIT 1;

    IF v_pricing IS NULL THEN
        -- No pricing row; return zero cost and a sentinel timestamp.
        RETURN QUERY SELECT 0::bigint, '-infinity'::timestamptz;
        RETURN;
    END IF;

    -- Integer math throughout. tokens * micro_per_mtok / 1_000_000
    -- = micro_dollars contribution from that token category.
    v_micro := (p_input_tokens::bigint  * v_pricing.input_micro_per_mtok  / 1000000)
             + (p_output_tokens::bigint * v_pricing.output_micro_per_mtok / 1000000);

    IF v_pricing.cache_write_micro_per_mtok IS NOT NULL AND p_cache_write_tokens > 0 THEN
        v_micro := v_micro + (p_cache_write_tokens::bigint
                              * v_pricing.cache_write_micro_per_mtok / 1000000);
    END IF;

    IF v_pricing.cache_read_micro_per_mtok IS NOT NULL AND p_cache_read_tokens > 0 THEN
        v_micro := v_micro + (p_cache_read_tokens::bigint
                              * v_pricing.cache_read_micro_per_mtok / 1000000);
    END IF;

    RETURN QUERY SELECT v_micro, v_pricing.effective_at;
END;
$func$;

COMMENT ON FUNCTION stewards.compute_cost(text, text, int, int, int, int) IS
'Phase 4a: compute cost in micro-dollars from token usage. Picks most-recent pricing whose effective_at <= now().';

-- ---------------------------------------------------------------------
-- record_cost_event(work_item, attempt, provider, model, tokens..., notes)
-- Inserts a cost_events row using compute_cost. The trigger then updates
-- work_items.cost_micro_dollars and cost_buckets.
-- Returns the new cost_events.id.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.record_cost_event(
    p_work_item_id       uuid,
    p_attempt_seq        int,
    p_provider           text,
    p_model              text,
    p_input_tokens       int,
    p_output_tokens      int,
    p_cache_write_tokens int DEFAULT 0,
    p_cache_read_tokens  int DEFAULT 0,
    p_notes              text DEFAULT NULL
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_micro bigint;
    v_pricing_at timestamptz;
    v_id bigint;
    v_notes text;
BEGIN
    SELECT micro_dollars, pricing_effective_at
      INTO v_micro, v_pricing_at
      FROM stewards.compute_cost(p_provider, p_model,
                                  p_input_tokens, p_output_tokens,
                                  p_cache_write_tokens, p_cache_read_tokens);

    -- If no pricing row exists, flag in notes so the gap is visible.
    v_notes := p_notes;
    IF v_pricing_at = '-infinity'::timestamptz THEN
        v_notes := coalesce(v_notes || ' | ', '')
                 || 'no_pricing_row(' || p_provider || '/' || p_model || ')';
    END IF;

    INSERT INTO stewards.cost_events
        (work_item_id, attempt_seq, provider, model,
         input_tokens, output_tokens, cache_write_tokens, cache_read_tokens,
         micro_dollars, pricing_effective_at, notes)
    VALUES
        (p_work_item_id, p_attempt_seq, p_provider, p_model,
         p_input_tokens, p_output_tokens, p_cache_write_tokens, p_cache_read_tokens,
         v_micro, v_pricing_at, v_notes)
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$func$;

COMMENT ON FUNCTION stewards.record_cost_event(uuid, int, text, text, int, int, int, int, text) IS
'Phase 4a: insert a cost_events row with computed micro_dollars. Trigger updates work_items + buckets.';

-- ---------------------------------------------------------------------
-- cost_cap_exceeded(work_item) -> boolean
-- True if the work_item has cost_cap_micro set and cost_micro_dollars >= cap.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.cost_cap_exceeded(p_work_item_id uuid)
RETURNS boolean
LANGUAGE sql STABLE AS $func$
    SELECT cost_cap_micro IS NOT NULL
           AND cost_micro_dollars >= cost_cap_micro
      FROM stewards.work_items
     WHERE id = p_work_item_id;
$func$;

COMMENT ON FUNCTION stewards.cost_cap_exceeded(uuid) IS
'Phase 4a: true if work_item has hit its cost cap. Used by steward_tick before retry dispatch.';

-- =====================================================================
-- Bucket helpers
-- =====================================================================

-- ---------------------------------------------------------------------
-- bucket_period_for(kind, ts) -> (period_start, period_end)
-- Returns the period boundaries containing ts for the given bucket_kind.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.bucket_period_for(
    p_kind text,
    p_ts   timestamptz DEFAULT now()
) RETURNS TABLE (period_start timestamptz, period_end timestamptz)
LANGUAGE plpgsql IMMUTABLE AS $func$
BEGIN
    -- session_5h: 5-hour windows aligned to UTC midnight
    -- (so 00:00, 05:00, 10:00, 15:00, 20:00 UTC)
    IF p_kind = 'session_5h' THEN
        period_start := date_trunc('hour', p_ts)
                      - (extract(hour FROM p_ts)::int % 5) * interval '1 hour';
        period_end   := period_start + interval '5 hours';
    ELSIF p_kind = 'daily' THEN
        period_start := date_trunc('day', p_ts);
        period_end   := period_start + interval '1 day';
    ELSIF p_kind = 'weekly' THEN
        -- ISO week (Monday start). Sunday-9pm cutover (Michael's rough memory)
        -- can be added via a config column later.
        period_start := date_trunc('week', p_ts);
        period_end   := period_start + interval '1 week';
    ELSIF p_kind = 'monthly' THEN
        period_start := date_trunc('month', p_ts);
        period_end   := period_start + interval '1 month';
    ELSE
        RAISE EXCEPTION 'unknown bucket_kind: %', p_kind;
    END IF;
    RETURN NEXT;
END;
$func$;

-- ---------------------------------------------------------------------
-- bucket_current(provider, kind) -> cost_buckets row
-- Returns the active bucket row for the current period, creating it
-- lazily if it doesn't exist. UPSERT on conflict.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.bucket_current(
    p_provider text,
    p_kind     text
) RETURNS stewards.cost_buckets
LANGUAGE plpgsql AS $func$
DECLARE
    v_period record;
    v_bucket stewards.cost_buckets;
    v_default_limit bigint;
BEGIN
    SELECT * INTO v_period
      FROM stewards.bucket_period_for(p_kind, now());

    -- Try to find existing
    SELECT * INTO v_bucket
      FROM stewards.cost_buckets
     WHERE provider = p_provider
       AND bucket_kind = p_kind
       AND period_start = v_period.period_start;

    IF v_bucket IS NOT NULL THEN
        RETURN v_bucket;
    END IF;

    -- Fetch default limit from any historical row for this (provider, kind)
    -- so new periods inherit the configured cap.
    SELECT bucket_limit_micro INTO v_default_limit
      FROM stewards.cost_buckets
     WHERE provider = p_provider
       AND bucket_kind = p_kind
       AND bucket_limit_micro IS NOT NULL
     ORDER BY period_start DESC
     LIMIT 1;

    INSERT INTO stewards.cost_buckets
        (provider, bucket_kind, period_start, period_end,
         micro_dollars, bucket_limit_micro)
    VALUES
        (p_provider, p_kind, v_period.period_start, v_period.period_end,
         0, v_default_limit)
    ON CONFLICT (provider, bucket_kind, period_start) DO NOTHING
    RETURNING * INTO v_bucket;

    -- If ON CONFLICT skipped (race), refetch
    IF v_bucket IS NULL THEN
        SELECT * INTO v_bucket
          FROM stewards.cost_buckets
         WHERE provider = p_provider
           AND bucket_kind = p_kind
           AND period_start = v_period.period_start;
    END IF;

    RETURN v_bucket;
END;
$func$;

-- ---------------------------------------------------------------------
-- bucket_record(provider, kind, micro_dollars)
-- Adds micro_dollars to the current bucket of (provider, kind).
-- Lazily opens the bucket if needed.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.bucket_record(
    p_provider     text,
    p_kind         text,
    p_micro_dollars bigint
) RETURNS void
LANGUAGE plpgsql AS $func$
DECLARE
    v_bucket stewards.cost_buckets;
BEGIN
    v_bucket := stewards.bucket_current(p_provider, p_kind);
    UPDATE stewards.cost_buckets
       SET micro_dollars = micro_dollars + p_micro_dollars
     WHERE id = v_bucket.id;
END;
$func$;

-- =====================================================================
-- Trigger: maintain work_items.cost_micro_dollars + buckets on cost_event insert
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.cost_events_after_insert()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    -- Update work_item cumulative + cap-tripped timestamp
    UPDATE stewards.work_items
       SET cost_micro_dollars = cost_micro_dollars + NEW.micro_dollars,
           cost_capped_at = CASE
               WHEN cost_capped_at IS NOT NULL THEN cost_capped_at
               WHEN cost_cap_micro IS NOT NULL
                    AND (cost_micro_dollars + NEW.micro_dollars) >= cost_cap_micro
                    THEN now()
               ELSE NULL
           END
     WHERE id = NEW.work_item_id;

    -- Roll into all four bucket kinds for this provider
    PERFORM stewards.bucket_record(NEW.provider, 'session_5h', NEW.micro_dollars);
    PERFORM stewards.bucket_record(NEW.provider, 'daily',      NEW.micro_dollars);
    PERFORM stewards.bucket_record(NEW.provider, 'weekly',     NEW.micro_dollars);
    PERFORM stewards.bucket_record(NEW.provider, 'monthly',    NEW.micro_dollars);

    RETURN NEW;
END;
$func$;

DROP TRIGGER IF EXISTS cost_events_after_insert ON stewards.cost_events;
CREATE TRIGGER cost_events_after_insert
AFTER INSERT ON stewards.cost_events
FOR EACH ROW EXECUTE FUNCTION stewards.cost_events_after_insert();

-- =====================================================================
-- Seed: model_pricing (from opencode.ai/docs/zen, fetched 2026-05-10)
-- All rates per-1M-tokens, in micro-dollars. NULL cache_write means
-- provider does not expose that distinction.
-- ON CONFLICT DO UPDATE so re-running the migration refreshes rates.
-- =====================================================================

INSERT INTO stewards.model_pricing
    (provider, model, input_micro_per_mtok, output_micro_per_mtok,
     cache_write_micro_per_mtok, cache_read_micro_per_mtok, notes)
VALUES
    -- Chinese models (substrate's main chain)
    ('opencode_go', 'kimi-k2.6',          950000,  4000000,    NULL,  160000,
     'Cache write rate not exposed by provider'),
    ('opencode_go', 'glm-5.1',           1400000,  4400000,    NULL,  260000,
     'Reasoning model (backend frank/GLM-5.1). Streams content fine via the substrate (auto-probe verified 2026-05-29). Give adequate per-call max_tokens for substantive prompts so reasoning does not exhaust the budget before content.'),
    ('opencode_go', 'minimax-m2.7',       300000,  1200000,  375000,   60000,
     'Anthropic-FORMAT model (api_format=anthropic). Usable via the substrate AN.2 /messages dispatch path (2026-05-30).'),
    ('opencode_go', 'qwen3.6-plus',       500000,  3000000,  625000,   50000, '')
    -- claude-* (opus-4.5/4.6/4.7, sonnet-4.6, haiku-4.5) PRUNED 2026-05-31:
    -- opencode removed them from the zen/go gateway ("Model ... is not supported"
    -- on both /chat/completions and /messages). They were seeded as
    -- human-mediated-escalation targets; nothing dispatches them by name. Re-add
    -- if opencode restores them.
ON CONFLICT (provider, model, effective_at) DO UPDATE
SET input_micro_per_mtok       = EXCLUDED.input_micro_per_mtok,
    output_micro_per_mtok      = EXCLUDED.output_micro_per_mtok,
    cache_write_micro_per_mtok = EXCLUDED.cache_write_micro_per_mtok,
    cache_read_micro_per_mtok  = EXCLUDED.cache_read_micro_per_mtok,
    notes                      = EXCLUDED.notes;

-- =====================================================================
-- Seed: cost_buckets initial limits (OpenCode Go subscription caps)
-- Per Michael's clarification: ~$12/day, ~$60/month soft caps. Weekly
-- TBD; left NULL until confirmed. session_5h informational only.
--
-- These rows are created at the *current* period boundaries so
-- bucket_current() finds them on first call.
-- =====================================================================

DO $seed$
DECLARE
    v_5h record;
    v_d  record;
    v_w  record;
    v_m  record;
BEGIN
    SELECT * INTO v_5h FROM stewards.bucket_period_for('session_5h', now());
    SELECT * INTO v_d  FROM stewards.bucket_period_for('daily',      now());
    SELECT * INTO v_w  FROM stewards.bucket_period_for('weekly',     now());
    SELECT * INTO v_m  FROM stewards.bucket_period_for('monthly',    now());

    INSERT INTO stewards.cost_buckets
        (provider, bucket_kind, period_start, period_end,
         micro_dollars, bucket_limit_micro, notes)
    VALUES
        ('opencode_go', 'session_5h', v_5h.period_start, v_5h.period_end,
         0, NULL, 'informational; OpenCode does not expose 5h window'),
        ('opencode_go', 'daily',      v_d.period_start,  v_d.period_end,
         0, 12000000, 'OpenCode Go daily soft cap ($12); overage via Zen pay-per-token'),
        ('opencode_go', 'weekly',     v_w.period_start,  v_w.period_end,
         0, NULL, 'OpenCode Go weekly cap TBD; bucket informational until confirmed'),
        ('opencode_go', 'monthly',    v_m.period_start,  v_m.period_end,
         0, 60000000, 'OpenCode Go monthly cap ($60); overage via Zen pay-per-token')
    ON CONFLICT (provider, bucket_kind, period_start) DO UPDATE
    SET bucket_limit_micro = EXCLUDED.bucket_limit_micro,
        notes              = EXCLUDED.notes;
END;
$seed$;

-- =====================================================================
-- Done. Phase 4a-cost is operational.
-- Acceptance: SELECT * FROM stewards.compute_cost('opencode_go','kimi-k2.6',1000000,500000);
-- Expected:   micro_dollars = 950000 + 2000000 = 2950000  ($2.95)
-- =====================================================================
