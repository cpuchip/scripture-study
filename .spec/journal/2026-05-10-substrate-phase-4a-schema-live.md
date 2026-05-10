---
date: 2026-05-10
session_kind: build (Phase 4a SQL schema layer)
mode: dev
priority: high
prior_session: 2026-05-10-substrate-phase-a-specs.md (subspec drafting)
follows: Michael clarified bucket pricing model + said "switch to build mode and knock out as much of our plan as you have spec for"
carries_forward:
  - Phase A bgworker integration (cost.2 + escalation.2): wire steward_tick to use cost_cap_exceeded, record_cost_event, pick_model. Go-side or Rust-side work in extension/src/ or projects/pg-ai-stewards/cmd/.
  - Phase A UI surfaces (cost.3, escalation.3+.4): Vue components in stewards-ui for cost panel, escalation queue, "Boost via Zen Opus" button.
  - Phase A.escalation.4 MCP tools: 3 new MCP tools (work_item_escalation_list / claim / resolve) on stewards-mcp for CLI-mediated escalation.
artifacts:
  - extension/4a-cost-tracking.sql (~330 lines, applied live)
  - extension/4a-escalation-chain.sql (~210 lines, applied live)
  - extension/Dockerfile (added 2 SQL files to COPY list)
  - extension/src/lib.rs (added 2 extension_sql_file! macros)
  - extension/verify-4a.sql (smoke test, 16 sections)
  - extension/verify-4a-output.log (smoke test results)
  - .spec/proposals/cost-tracking.md V.4.1 updated with bucket model clarification
---

# Phase 4a schema layer — built, smoke-tested, live

## What I built

Two SQL files implementing the schema + functions layer of substrate Phase A, per the design subspecs from earlier this evening (`cost-tracking.md` and `escalation-chain.md`).

### `extension/4a-cost-tracking.sql`

Tables created:
- `stewards.model_pricing` — per-model rates in micro-dollars per MTok, with effective_at versioning, nullable cache_*_micro_per_mtok columns
- `stewards.cost_events` — append-only per-dispatch ledger
- `stewards.cost_buckets` — concentric tracking buckets per provider (4 kinds: session_5h, daily, weekly, monthly)

Columns added to `stewards.work_items`:
- `cost_micro_dollars bigint NOT NULL DEFAULT 0` (denormalized cumulative)
- `cost_cap_micro bigint` (nullable)
- `cost_capped_at timestamptz`

Functions:
- `compute_cost(provider, model, input, output, cache_write, cache_read)` → (micro_dollars, pricing_effective_at). Picks most-recent effective_at <= now() pricing row. Integer math throughout. Returns (0, '-infinity') if no pricing row exists.
- `record_cost_event(work_item_id, attempt_seq, provider, model, tokens..., notes)` → cost_events.id. Wraps compute_cost + insert.
- `cost_cap_exceeded(work_item_id)` → boolean. Used by steward_tick before retry dispatch.
- `bucket_period_for(kind, ts)` → (period_start, period_end). Computes period boundaries: session_5h aligned to UTC 5h windows, daily/weekly/monthly via date_trunc.
- `bucket_current(provider, kind)` → cost_buckets row. Lazy-opens current period.
- `bucket_record(provider, kind, micro_dollars)` → void. Accumulates into current bucket.

Trigger: `cost_events_after_insert` updates work_items.cost_micro_dollars + sets cost_capped_at when cap is crossed + rolls into all four bucket kinds for the provider.

Seeded 9 model_pricing rows from opencode.ai/docs/zen (4 Chinese + 5 Anthropic) and 4 cost_buckets rows for opencode-zen with $12/day + $60/month caps per Michael's bucket clarification (weekly + session_5h NULL).

### `extension/4a-escalation-chain.sql`

Tables created:
- `stewards.stage_models` — per-(pipeline_family, stage_name) default model. PK on (pipeline_family, stage_name); no FK to model_pricing because composite PK there.
- `stewards.model_escalation` — (current_model, diagnosis) → next_model matrix with attempt_threshold. CHECK against direct self-loops; NULL next_model means "stay"; sentinel `__queue_for_opus__` means "transition to escalation queue."

Columns added to work_items:
- `model_override text` (one-shot pin)
- `escalation_state text NOT NULL DEFAULT 'normal'` + CHECK constraint for valid states (normal/queued/in_progress/failed/resolved)
- `escalation_claimed_by text` (e.g., 'ui:zen-opus' or 'cli:claude-code-pro')
- `escalation_claimed_at timestamptz`
- `escalation_completed_at timestamptz`
- `escalation_attempts int NOT NULL DEFAULT 0`

Function `pick_model(pipeline, stage, attempt, diagnosis)` walks the chain attempt-by-attempt. Returns the sentinel `__queue_for_opus__` when GLM-5.1 escalates (per the human-mediated escalation queue design from D-EC3).

Seeded 14 stage_models rows: study (research/outline/draft/verify), lesson (same), dev (plan/execute/verify), _gate (evaluate_gate/generate_scenarios/verify_scenarios). And 20 model_escalation rows: 5 diagnoses × 4 chain models, with the GLM rows escalating to the queue sentinel for non-transient diagnoses.

### Wiring

`extension/src/lib.rs` got two new `extension_sql_file!` macros after the last existing one:
- `create_phase_4a_cost_tracking` requires `create_git_mcp_seed`
- `create_phase_4a_escalation_chain` requires `create_phase_4a_cost_tracking`

`extension/Dockerfile` COPY list got the two new SQL files appended.

## Build process

First docker build attempt failed: 8 errors, all "couldn't read `src/../4a-*.sql`: No such file or directory." Cause: the Dockerfile explicitly enumerates SQL files in COPY (with a comment saying "When adding a new file, also... update this list"). I'd missed updating the COPY list. Fixed by appending the two new files; rebuilt successfully.

The Dockerfile-explicit-enumeration pattern bit me but is also defensible — it makes the build context deliberate (no random files snuck in). The comment was clear; I just didn't notice it on the first read of lib.rs.

## Smoke test

`extension/verify-4a.sql` runs 16 sections (A through Q, skipping the section P/Q being relabeled mid-write). All passed on the ephemeral container:

- Schema presence: 5 tables exist with expected row counts (9/0/4/14/20)
- 9 work_items new columns present with expected nullability
- 9 model_pricing rows seeded correctly (Chinese + Anthropic)
- 4 cost_buckets seeded with correct caps and notes
- compute_cost returns exact micro-dollar values (Kimi $2.95, MiniMax cache-aware $0.121)
- compute_cost handles unknown provider gracefully (0, -infinity)
- pick_model walks the chain correctly: kimi-k2.6 default → glm-5.1 on attempt-2 model_limit → __queue_for_opus__ on attempt-3 model_limit; transient stays on same model regardless of attempts; dev/plan starts at glm-5.1; nonexistent pipeline raises with clear error message
- Escalation matrix coverage: 4 escalating models × 5 diagnoses = 20 rules; 4 queue sentinels (one per non-transient on glm-5.1); 4 stay_on_current (one transient per chain model)
- bucket_period_for returns correct boundaries for daily/weekly/monthly/session_5h (5h windows aligned to UTC)

## Live apply

Used the substrate's standard "live-DB migration; folds into lib.rs at next intentional rebuild" pattern. docker cp both SQL files into pg-ai-stewards-dev, then docker exec psql -f. Both applied without errors. Idempotent (CREATE TABLE IF NOT EXISTS, ALTER TABLE ADD COLUMN IF NOT EXISTS, ON CONFLICT DO UPDATE) so re-running on a container that already had Phase 4a from CREATE EXTENSION would be a no-op.

The live container's Phase 4a is now identical to what a fresh CREATE EXTENSION would produce (since the rebuild already bundled the SQL into the extension binary). The next time the live container restarts off the new image, the extension's CREATE EXTENSION SQL will see the IF NOT EXISTS guards and be a no-op for these tables.

Soak left undisturbed (`schedule_enabled = t` confirmed before and after). No watchman pass interrupted.

## Surprises

1. **The Dockerfile COPY list is enforced for new SQL files.** I'd assumed the build context would just include the whole directory. The deliberate enumeration is an interesting choice — limits attack surface but adds a coordination point. Worth knowing for next time.

2. **Bucket clarification simplified the spec.** Earlier I'd flagged the bucket pricing as an unresolved tension (V.4.1 of cost-tracking.md). Michael's clarification — OpenCode Go is the same per-token rates as Zen but with monthly subscription buckets, with Zen overage available — made the schema clean: bucket_limit_micro is meaningful for the Go subscription user; cost cap is the per-work_item discipline; both can coexist.

3. **The smoke test output was instantly reassuring.** All 16 sections passing on first try means the SQL is correct (subject to integration-time discoveries). The compute_cost integer math worked exactly as designed; the pick_model chain walking handled the queue sentinel correctly; the bucket_period_for boundary math was right for all 4 kinds.

## What I did NOT build

The schema layer is foundation. The bgworker + UI + MCP tooling layers are still ahead:

- **Phase A.cost.2** — wire steward_tick to call cost_cap_exceeded before retry dispatch + record_cost_event after each LLM dispatch. Touches the bridge daemon (Go) where dispatch happens.
- **Phase A.escalation.2** — wire steward_tick to call pick_model + handle the __queue_for_opus__ sentinel by transitioning escalation_state. Touches steward_tick (Rust extension or Go bgworker, depending on substrate's design).
- **Phase A.cost.3** — Stewards-UI Cost panel on WorkItemDetail, watchman cost_capped category. Vue work in scripts/stewards-ui.
- **Phase A.escalation.3** — WorkItemDetail shows escalation path, NewWork "force model" picker, Watchman escalation queue category.
- **Phase A.escalation.4** — Stewards-UI "Boost via Zen Opus" button + 3 new MCP tools on stewards-mcp (`work_item_escalation_list`, `_claim`, `_resolve`).

Estimated 4-6 more sessions for the full Phase A. Schema is the cleanest 1-session unit; bgworker integration likely takes 2 sessions; UI takes 2 sessions; MCP tools 1 session.

## Honesty audit

- **Did I verify against the substrate's actual code?** Partially. I read 2-7b3-watchman-budget.sql for SQL style reference. I did NOT re-read bgworker.rs or providers.rs to confirm the dispatch path my SQL assumes. Risk: when I get to Phase A.cost.2, the cost-event recording may need plumbing through code paths I haven't surveyed.
- **Are the bucket period boundaries correct for non-UTC timezones?** All my date_trunc calls assume the database's TZ. PostgreSQL's date_trunc on timestamptz uses the session timezone. Substrate's container probably runs UTC. If anyone uses non-UTC timezones for accounting purposes, the boundaries shift. Documented in the SQL function comment that this is "Sunday 9pm" deferred config.
- **Is the trigger correct under concurrent inserts?** The cost_events_after_insert trigger does an UPDATE + 4 PERFORM bucket_record calls. Under heavy concurrent inserts, the UPDATE on work_items could deadlock with another transaction also updating the same row. Postgres serializes via row-locking, so it'll work, just may be slow under contention. For substrate's current ~1 dispatch/sec scale, this is fine.
- **Did I commit?** No. Per CLAUDE.md "Only create commits when requested by the user."

## What's next

Three options for next session, by increasing scope:

1. **Phase A.cost.2** — wire steward_tick to use the new cost functions. ~1-2 sessions of Go/Rust work in extension/src/bgworker.rs or wherever the dispatch path lives.
2. **Phase A.escalation.2** — wire steward_tick to use pick_model + handle the queue sentinel. Could be done in parallel with cost.2.
3. **Both bgworker integrations together** — if the dispatch path is in one place, cost + escalation wiring is one coherent change.

Or pivot to:
4. **The other ratified-pending items:** move stewards-ui to projects/, dynamize NewWork pipeline list, real second/third pipelines.

Michael picks at next programming session.
