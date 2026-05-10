---
title: Cost Tracking — token-multiplier model with provider-specific cache awareness
date: 2026-05-10
status: design sub-spec — ready for Phase A implementation
parent: full-agentic-substrate.md (D-A4 ratification)
purpose: >
  Replace the ratified flat dollar-cost cap idea with a token-multiplier
  cost model that handles input/cache-write/cache-read/output distinctions
  per model. Enables D-A4 (cost cap alongside token_budget) to be
  operational. Required before Phase A escalation logic ships.
---

# Cost Tracking — token-multiplier model

## I. Binding problem

The substrate today tracks `work_items.tokens_in` and `tokens_out` as flat counters. That's enough for volume awareness but useless for cost discipline because:

1. **Token volume ≠ cost.** A work_item that escalates from Qwen3.6 Plus to GLM-5.1 burns dramatically more dollars per token, but the existing counter treats all tokens equally.
2. **Cache-write tokens cost more than cache-reads.** Anthropic charges roughly 1.25× input rate for cache writes and roughly 0.1× input rate for cache reads. A naive "input tokens" counter misses both directions of this.
3. **D-A4 ratified a token-multiplier cost cap.** Without per-model rate metadata + a cost composer, the ratification is theoretical.

The agentic substrate's value depends on cost discipline being *real*. Phase A's escalation chain is a cost amplifier (failed attempt → escalate to bigger model → escalate again). Without accurate cost accounting, the substrate can spend $50 on a quarantined work_item before anyone notices.

## II. Success criteria

A Phase A work_item that completes (or quarantines) has:
1. **Per-attempt cost recorded** — accurate to within ~5% of the provider's billing
2. **Cost cap enforcement** — if cumulative cost crosses the cap, the steward refuses further dispatches and quarantines the work_item with reason `cost_cap_exceeded`
3. **Pricing table is data-only** — adding a new model = INSERT a row, no code change
4. **Cache distinction preserved** — for providers that expose cache_write vs cache_read in their token usage, cost accounting reflects the rate difference

## III. Constraints and boundaries

**In scope:**
- Schema additions for pricing + cost tracking
- SQL function for cost composition from token usage
- Integration with the Phase A steward tick (cost-cap check before retry)
- Stewards-UI surface showing cumulative cost per work_item

**Out of scope (explicitly):**
- Cost prediction (estimating cost before dispatch)
- Cost-aware model selection (escalation chain decides on cost grounds vs. capability grounds — capability wins; cost cap is a hard limit, not a soft optimizer)
- Per-user / per-pipeline budget rollups (this is per-work_item only for Phase A)
- Historical re-pricing if rates change (locked to the rate at the time of the dispatch)

## IV. Prior art

- **Substrate today:** `work_items.tokens_in` and `tokens_out` columns. No per-attempt breakdown. No cost.
- **Brain v3:** has `quarantineCostLimit` enforced in `steward.go` but uses a hardcoded per-model cost map (`internal/config/models.go: modelCosts`). One-line per-model, no cache distinction.
- **Anthropic API:** returns `usage: { input_tokens, output_tokens, cache_creation_input_tokens, cache_read_input_tokens }` on each call. The cache distinction is exposed by the API.
- **OpenCode Go providers** (Kimi K2.6, GLM-5.1, MiniMax M2.7, Qwen3.6 Plus): need to verify which expose cache distinction. Anthropic API spec is the canonical four-field shape; non-Anthropic providers usually return only `input_tokens` + `output_tokens`. The schema should accept both shapes.

## V. Proposed approach

### V.1 Schema additions

```sql
-- Per-model pricing. One row per (provider, model). Rates in micro-dollars
-- per 1M tokens (so 3.00 USD/MTok = 3000000 micro-dollars/MTok). Integer
-- math throughout to avoid float drift on aggregations.
CREATE TABLE stewards.model_pricing (
  provider                    text  not null,
  model                       text  not null,
  input_micro_per_mtok        bigint not null,    -- standard input
  output_micro_per_mtok       bigint not null,    -- standard output
  cache_write_micro_per_mtok  bigint,             -- nullable: provider may not expose
  cache_read_micro_per_mtok   bigint,             -- nullable: provider may not expose
  effective_at                timestamptz not null default now(),
  notes                       text,
  primary key (provider, model, effective_at)
);

-- Per-attempt cost ledger. One row per dispatched LLM call. Append-only.
CREATE TABLE stewards.cost_events (
  id                          bigserial primary key,
  work_item_id                uuid references stewards.work_items(id),
  attempt_seq                 int  not null,           -- 1, 2, 3 within work_item
  at                          timestamptz default now(),
  provider                    text not null,
  model                       text not null,
  input_tokens                int  not null default 0,
  output_tokens               int  not null default 0,
  cache_write_tokens          int  not null default 0,
  cache_read_tokens           int  not null default 0,
  micro_dollars               bigint not null,         -- computed at insert
  pricing_effective_at        timestamptz not null,    -- which pricing row was used
  notes                       text                     -- free-form (e.g., "retry after timeout")
);
CREATE INDEX cost_events_work_item ON stewards.cost_events(work_item_id);

-- Add to work_items:
ALTER TABLE stewards.work_items
  ADD COLUMN cost_micro_dollars   bigint  default 0,    -- denormalized cumulative
  ADD COLUMN cost_cap_micro       bigint,                -- nullable: NULL means no cap
  ADD COLUMN cost_capped_at       timestamptz;           -- set when cap was hit
```

**Why integer micro-dollars:** float arithmetic on cost aggregations across thousands of rows accumulates error. Integer micro-dollars give 6-decimal precision and exact summation. Display layer divides by 1_000_000 for human-readable USD.

**Why per-attempt rows + denormalized cumulative on work_items:** the events table is the audit trail (every dispatch's exact tokens + price). The denormalized total on work_items is for the cap check (one row read instead of an aggregate). A trigger keeps them consistent.

### V.2 SQL functions

```sql
-- Compute cost in micro-dollars from token usage + provider/model.
-- Picks the most-recent pricing row whose effective_at <= now().
-- Returns 0 if no pricing row exists (logs a warning row in cost_events.notes).
CREATE FUNCTION stewards.compute_cost(
  p_provider           text,
  p_model              text,
  p_input_tokens       int,
  p_output_tokens      int,
  p_cache_write_tokens int default 0,
  p_cache_read_tokens  int default 0
) RETURNS TABLE (micro_dollars bigint, pricing_effective_at timestamptz);

-- Insert a cost event for a work_item. Computes cost via compute_cost,
-- inserts into cost_events, updates the work_items denormalized total
-- via trigger.
CREATE FUNCTION stewards.record_cost_event(
  p_work_item_id       uuid,
  p_attempt_seq        int,
  p_provider           text,
  p_model              text,
  p_input_tokens       int,
  p_output_tokens      int,
  p_cache_write_tokens int default 0,
  p_cache_read_tokens  int default 0,
  p_notes              text default null
) RETURNS bigint;  -- the cost_events.id

-- Check whether a work_item has hit its cost cap.
-- Returns true if cost_micro_dollars >= cost_cap_micro and cap is non-null.
-- Used by the steward tick before dispatching a retry.
CREATE FUNCTION stewards.cost_cap_exceeded(p_work_item_id uuid) RETURNS boolean;
```

**Trigger behavior:** `AFTER INSERT ON stewards.cost_events`, update `work_items.cost_micro_dollars += NEW.micro_dollars`. If the new total crosses `cost_cap_micro`, set `work_items.cost_capped_at = now()`. Steward tick polls this on its retry pass.

### V.3 Integration with Phase A steward loop

The Phase A `steward_tick` bgworker (per the parent proposal) gets an additional check before dispatching a retry:

```sql
-- In steward_tick pseudocode:
FOR each work_item WHERE status='failed' AND failure_count < 3 AND NOT quarantined:
  IF stewards.cost_cap_exceeded(work_item.id) THEN
    UPDATE work_items SET quarantined_at = now(),
      quarantine_reason = 'cost_cap_exceeded'
    WHERE id = work_item.id;
    INSERT INTO steward_actions (work_item_id, observation, diagnosis, action, ...)
      VALUES (work_item.id, 'cumulative cost ' || cost_micro_dollars || ' exceeds cap',
              'cost_limit', 'quarantine', ...);
    CONTINUE;
  END IF;
  -- ... existing diagnosis + retry logic
```

The bgworker that dispatches a stage call records cost via `stewards.record_cost_event(...)` after the provider response is parsed. The token-counts come from the provider response's `usage` field.

### V.4 Pricing seed data (from opencode.ai/docs/zen, fetched 2026-05-10)

```sql
-- All rates from OpenCode Zen pricing page. Per-token, USD per 1M tokens,
-- expressed in micro-dollars per MTok (so $0.95/MTok = 950000 micro-dollars/MTok).
-- Cache fields: NULL = provider doesn't expose this distinction.
INSERT INTO stewards.model_pricing
  (provider, model, input_micro_per_mtok, output_micro_per_mtok,
   cache_write_micro_per_mtok, cache_read_micro_per_mtok, notes)
VALUES
  -- Chinese models (substrate's main escalation chain — cheap tier)
  ('opencode-zen', 'kimi-k2.6',      950000,  4000000,    NULL,  160000,
    'Cache write rate not exposed by provider'),
  ('opencode-zen', 'glm-5.1',       1400000,  4400000,    NULL,  260000,
    'Cache write rate not exposed by provider'),
  ('opencode-zen', 'minimax-m2.7',   300000,  1200000,  375000,   60000, ''),
  ('opencode-zen', 'qwen3.6-plus',   500000,  3000000,  625000,   50000, ''),

  -- Anthropic models via OpenCode Zen (used only via human-mediated escalation
  -- queue, NOT by automatic chain dispatch — see escalation-chain.md)
  ('opencode-zen', 'claude-opus-4-7',     5000000, 25000000, 6250000, 500000, ''),
  ('opencode-zen', 'claude-opus-4-6',     5000000, 25000000, 6250000, 500000, ''),
  ('opencode-zen', 'claude-opus-4-5',     5000000, 25000000, 6250000, 500000, ''),
  ('opencode-zen', 'claude-sonnet-4-6',   3000000, 15000000, 3750000, 300000, ''),
  ('opencode-zen', 'claude-haiku-4-5',    1000000,  5000000, 1250000, 100000, '');
```

**Provider naming:** all rows use `provider='opencode-zen'` because per WebFetch, OpenCode Zen carries both the cheap Chinese models AND the Anthropic models — single provider, different tiers within. (See bucket-vs-per-token note below.)

### V.4.1 Bucket modeling — UNRESOLVED, needs Michael clarification

Michael described OpenCode pricing as having three concentric session buckets:
- **5-hour session bucket**
- **Weekly bucket** (resets Sunday 9pm — configurable?)
- **Monthly bucket** (default 1st of month — configurable)

The OpenCode Zen documentation page (fetched 2026-05-10) shows **only per-token pricing** — no session-bucket or weekly/monthly reset tiers, only "optional monthly spending limits." This is a tension worth resolving before Phase A.cost ships.

**Possible interpretations:**
1. The bucket pricing exists on a different OpenCode tier/page (e.g., OpenCode Go vs OpenCode Zen)
2. The bucket pricing is a customer-account feature not on public docs
3. Michael's mental model conflated OpenCode with Claude Code's own session-bucket model (Claude Pro's 5-hour usage windows) — possible since Claude Code itself uses 5h sessions

**Spec decision:** ratified per Michael's D-A4 answer ("Lets track both, but no limit on bucket headroom"). Schema accommodates both:

```sql
-- Bucket tracking schema, additive to model_pricing.
CREATE TABLE stewards.cost_buckets (
  id                 bigserial primary key,
  provider           text  not null,         -- 'opencode-zen', 'claude-code-cli', etc.
  bucket_kind        text  not null check (bucket_kind IN ('session_5h','weekly','monthly')),
  period_start       timestamptz not null,
  period_end         timestamptz not null,
  micro_dollars      bigint default 0,       -- accumulated this period
  -- bucket_limit_micro is NULL — per Michael's D-A4: "no limit on bucket headroom"
  -- Set non-NULL later if/when actual bucket caps are confirmed
  bucket_limit_micro bigint,
  notes              text
);
CREATE INDEX cost_buckets_period ON stewards.cost_buckets
  (provider, bucket_kind, period_end);

-- Functions for rolling buckets:
-- stewards.bucket_current(provider, bucket_kind) returns current period's row
-- stewards.bucket_record(provider, bucket_kind, micro_dollars) accumulates
-- stewards.bucket_roll() called by bgworker to close expired periods + open new ones
```

**What ships in Phase A.cost:**
- Schema for `cost_buckets`
- 3 default rows per provider (5h/weekly/monthly) initialized empty
- Bucket roll logic (5h on the hour-mark; weekly Sunday 9pm in user's TZ; monthly 1st day)
- Accumulation alongside `cost_events` (every cost_event also increments the relevant buckets)
- UI surface: Stewards-UI shows current bucket consumption — informational only, no enforcement

**What does NOT ship:**
- Bucket limit enforcement (per Michael's D-A4)
- Bucket-aware model selection (escalation chain doesn't consult buckets)

If Michael later confirms bucket pricing is real and wants enforcement, it's a non-blocking addition: set `bucket_limit_micro` to a value, add a check in steward_tick. Schema is ready.

### V.5 Stewards-UI surface

WorkItemDetail page gets a Cost panel:

```
Cost panel:
- Cumulative: $X.XXXX (Y events)
- Cap: $Z.ZZZZ ([NOT SET | hit at TIMESTAMP])
- Per-attempt breakdown:
    Attempt 1: kimi-k2.6, 1234 in / 567 out → $0.0042
    Attempt 2: glm-5.1, 2345 in / 890 out → $0.0231
    ...
```

Cost-cap status visible at the work_item card level (not just in detail page). Watchman page surfaces cost-capped work_items as a category alongside quarantined.

## VI. Phased delivery

Three sub-phases, each independently shippable:

### Phase A.cost.1 — Schema + composer (1 session)
- Migration: 3 tables, 3 columns on work_items, the trigger
- SQL functions: `compute_cost`, `record_cost_event`, `cost_cap_exceeded`
- Seed `model_pricing` with the Anthropic reference rows + placeholders for OpenCode Go
- pgTAP tests covering: integer-math correctness, trigger denormalization, cap-detection edge cases (exactly-at-cap, cap-changes-mid-flight)

### Phase A.cost.2 — Bgworker integration (1 session)
- Wire the bridge daemon's tool-dispatch handler to call `record_cost_event` after parsing the provider response
- Wire the steward tick's pre-retry check to `cost_cap_exceeded`
- Quarantine logic with `cost_cap_exceeded` as the new reason

### Phase A.cost.3 — UI surface (1 session, can run parallel with A.cost.2)
- WorkItemDetail Cost panel
- WorkItem card cost summary
- Watchman cost-capped category

## VII. Verification criteria

For each sub-phase:

**A.cost.1 verification:**
- Insert a synthetic cost event with known token counts and known pricing → micro_dollars matches hand-calculated value to the dollar
- Inserting an event that crosses the cap sets `cost_capped_at`
- Three sequential events on one work_item produce a denormalized total = sum of individuals
- A pricing row with `effective_at = future` is NOT picked by `compute_cost`
- A NULL `cache_write_micro_per_mtok` causes cache_write_tokens to contribute zero (not crash)

**A.cost.2 verification:**
- A real dispatch through the bridge writes a cost_events row with provider/model/tokens matching the response
- A work_item with cost_cap_micro = 1000 that's already at 1100 → next steward tick quarantines it without dispatching
- The quarantine writes a steward_actions row with diagnosis=`cost_limit`

**A.cost.3 verification:**
- WorkItemDetail Cost panel shows accurate per-attempt breakdown
- Cost-capped work_items appear in Watchman with the `cost_capped` badge
- Manually inflating a cost_event in SQL shows up in the UI within one auto-refresh cycle

**Inverse hypothesis verification (Agans Rule 9):**
- Reproduce a cost-cap quarantine → confirm quarantined → manually clear `cost_capped_at` → confirm re-dispatched → restore the cap state.

## VIII. Costs and risks

**Build cost:** ~3 sessions. Schema migration, 3 SQL functions, bgworker wiring, UI panel. Mostly straightforward.

**Runtime cost:** Each dispatched LLM call writes one cost_events row (~200 bytes). At 100 dispatches/day, that's ~7 MB/year. Cost_events table can be partitioned by month if growth concerns emerge.

**Risks:**
1. **Pricing drift.** Provider rates change. Mitigation: `effective_at` column means new rates can be inserted without rewriting history. Locking dispatched events to `pricing_effective_at` preserves audit accuracy.
2. **Provider returns no usage data.** Some providers may not include `usage` in responses. Mitigation: `record_cost_event` inserts a row with zeros and `notes='no_usage_reported'` so the gap is visible. Cost cap won't trigger from these events (zero dollars).
3. **OpenCode Go bucket model mismatch.** Theoretical per-token rates don't match the actual flat-rate bill. Risk: cost cap triggers when the bucket actually has headroom. Mitigation: cost cap is configurable per work_item; default cap can be high or NULL during initial Phase A.
4. **Cache-distinction provider variance.** If a provider returns a four-field usage shape we don't expect, the parser silently drops the cache fields. Mitigation: `notes` column captures parse anomalies for review.

## IX. Open questions for Michael — RESOLVED 2026-05-10

1. **Pricing values** — RESOLVED: pulled from opencode.ai/docs/zen, seeded in V.4. All four Chinese models + 5 Anthropic models loaded with real per-token rates.
2. **Default cost cap** — RESOLVED: NULL by default. Cost cap is opt-in per work_item. Schema column is nullable; no default value in migration.
3. **Theoretical-rate vs bucket-aware** — RESOLVED: track BOTH. Per-token cost via cost_events (already specced). Bucket consumption via cost_buckets (added in V.4.1). No enforcement on buckets — informational only.
4. **Cost cap on NewWork form** — RESOLVED: inherit from pipeline default, override per-item. NewWork form pre-fills with the pipeline's default cap (NULL if pipeline has no default); user can override before submit.

### IX.1 NEW unresolved — bucket pricing tension

The OpenCode Zen docs page shows only per-token pricing (no session/weekly/monthly buckets), but Michael described 3 concentric buckets in D-A4. Possible explanations in V.4.1. **Action needed:** Michael confirms whether (a) buckets exist on a different OpenCode tier, (b) account-level feature, or (c) we drop bucket schema entirely and use per-token only. Phase A.cost.1 can ship without resolution (bucket schema is additive, harmless if buckets aren't real).

## X. Acceptance scenarios (decision-ready handoff)

After Phase A.cost ships, a developer or Phase A coding agent can verify it works by:

1. **Happy path:** Create a work_item with cost_cap_micro=10_000 ($0.01). Dispatch a stage that consumes ~$0.005. Confirm cost_events row written with correct micro-dollar value, work_items.cost_micro_dollars updated, no cap trigger.
2. **Cap trip:** Same work_item, dispatch another stage that pushes total over $0.01. Confirm cost_capped_at set. Run steward_tick; confirm quarantine with `cost_cap_exceeded` reason.
3. **Pricing change:** Insert a new model_pricing row with a higher rate, effective tomorrow. Dispatch today; confirm old rate used. Advance system clock; confirm new rate used.
4. **Provider without cache:** Mock a response with only input/output (no cache fields). Confirm cost_event has cache_*_tokens = 0 and micro_dollars computed from input + output only.
5. **UI verification:** Open WorkItemDetail for a cost-capped item. Cost panel shows breakdown, total, cap, and trip timestamp. Watchman page shows the item in the cost_capped category.

If all five pass, this sub-spec is operationally complete and Phase A's escalation chain (next sub-spec) can build on it.
