---
date: 2026-05-11
session_kind: build
workstream: WS5
substrate_phase: E
commits: [25f8599, 06efcb5, d0a0b92, c7e6404, b5ecd25, 422b1dd, 2747fb3]
cost_usd: 0
---

# Substrate Phase E — shipped end-to-end on dev in one session

## What happened

Michael said "Lets continue with Phase E and continue with git commit at stable points." Same cadence as Phases C + D — seven commits in build order, each with a smoke test, all in one session. No LLM cost this time — Phase E is structural state machinery that gets exercised by future work_items, not by direct LLM dispatches.

## What shipped

**E.1 (25f8599)** — schema. Four tables:
- `stewards.trust_scores` keyed `(agent_family, pipeline_family, model)` per the 2026-05-11 keying refinement (model dimension recognizes that "kimi-k2.6 doing study-write outline" ≠ "qwen3.6-plus doing the same"). Columns for the three counters + trust_level + last_evaluated_at + last_completion_at.
- `stewards.trust_transitions` append-only audit. transition_kind = auto | manual; justification required for manual (D-E2); metrics jsonb snapshot at transition time.
- `stewards.gate_overrides` with FK to gate_decisions. justification required.
- `stewards.trust_thresholds` config seeded with `trainee_to_journeyman = 5/5/demote=true`, `journeyman_to_master = 15/15/demote=true`.

**E.2 (06efcb5)** — five SQL functions. `trust_record_success / _failure / _override` are the counter-bump helpers (each calls `evaluate_trust` after the bump). `evaluate_trust` applies promotion / demotion rules. `trust_adjust` is the manual-change path with required ≥10-char justification.

Smoke verified: 5x success → trainee→journeyman; 1x override → demoted back to trainee; manual master with proper justification worked; short ("short") justification correctly rejected.

**E.3 (d0a0b92)** — gate trust check. New helper `work_item_stage_actor(uuid)` returns `{agent_family, pipeline_family, model}` for a work_item's current stage, honoring model_override. `apply_gate_decision` extended: on advance, look up the actor; if no trust row OR trainee, surface for review instead of auto-advancing. On a real advance that lands at verified, fire `trust_record_success` alongside the existing `sabbath_dispatch`.

Smoke verified: gate-test-e2e-1 reset to outline/raw, applied {action:advance} → result was 'raw' (no maturity change), status='awaiting_review'. The (plan, study-write, kimi-k2.6) cell had no trust row → defaulted to trainee → surface.

**E.4 (c7e6404)** — retry composer. `retry_guidance_with_lessons(diagnosis, attempt, pipeline, stage)` wraps Phase A's `retry_guidance` and appends a "Recent lessons from this pipeline + stage:" section pulling last 3 from `lessons_recent_ratified` view. Stage-specific keying so outline retries don't get polluted with draft lessons.

Smoke: empty pool returns base text only; after inserting + ratifying a lesson, it appears bulleted in retry context.

The steward retry path will switch from `retry_guidance` to `retry_guidance_with_lessons` in a future commit (small surgery to 4c-steward-dispatch.sql); not in this commit because the steward path is the consumer, the composer is the provider.

**E.5 (b5ecd25)** — `apply_gate_override(gate_decision_id, overridden_by, new_action, justification)` does the full override flow atomically: INSERT into gate_overrides → trust_record_override (auto-demotes per D-E3) → re-apply `apply_gate_decision` with the new action. Synthetic decision jsonb prepends "[human override by X]" to reasoning so the audit trail stays complete. Validates: new_action ∈ {advance, revise, surface}; justification ≥10 chars; non-self-override (new_action ≠ original).

Smoke verified end-to-end: override decision #8 (advance) → revise. gate_overrides row #1 written; trust_scores row created for (plan, study-write, kimi-k2.6) with human_overrides=1 trust_level=trainee (already at floor); apply_gate_decision re-fired through revise path. Short-justification + same-action correctly rejected.

**E.6 (422b1dd)** — backend api/trust.go. Four endpoints: GET /api/trust/scores (full matrix), GET /api/trust/transitions (audit log; optional cell filter), POST /api/trust/adjust, POST /api/gate-overrides/apply. All four registered in api.go.

**E.7 (2747fb3)** — Vue surfaces. Trust.vue at /trust shows matrix grouped by pipeline_family (rows = agent×model, tier-tinted level badges, color-coded success/fail/override counters), recent transitions ledger with from→to arrow + auto/manual badge + justification, per-cell adjust modal with D-E2-enforced ≥10-char justification. WorkItemDetail.vue gets a per-row "Override gate decision…" button on each gate_decisions entry; modal collects new_action + overridden_by + justification, banner reminds that override counts as failure (D-E3 auto-demote).

Smoke: /trust serves 200; matrix renders the row left by E.5's override test.

## Surprises

**Counter-driven demotion is cleaner than the original gate_overrides query.** Initial design had evaluate_trust query the gate_overrides table to find recent overrides. But trust_record_override only bumps the trust_scores counter — it doesn't insert into gate_overrides (that path is owned by apply_gate_override). So the demotion never fired in standalone counter testing. Fixed by switching to a counter-driven heuristic: compare current human_overrides against the snapshot stored in metrics jsonb of the most recent promotion transition. Cleaner because the counter is the single source of truth, and now the helper works whether overrides come via gate_overrides INSERT or direct trust_record_override calls.

**No LLM cost this session.** Phase E is structural state machinery — promotion/demotion rules, gate behavior changes, override flows. Everything was exercised with synthetic counter increments and direct SQL calls. Phase E will get exercised by real work_items as they pass through gates in subsequent sessions.

**The Lesson #3 hook saved another rebuild.** No new pg_extern functions in Phase E (all SQL functions go through extension_sql_file!), so the auto-refresh hook didn't have anything to refresh. But it ran cleanly after the (purely SQL-side) image rebuilds without any manual intervention required — confirming the dev workflow stays smooth even when the hook is a no-op.

## Process / covenant

Seven commits, each with a smoke test. Same cadence as Phases C + D. Stewardship caught the demotion-counter bug in E.2 mid-build before commit — that's the right pattern: catch issues at smoke time, not after.

Cost: $0 (all synthetic; no LLM dispatches).

Soak paused at session start, re-enabled at session end. Bridge restarted at session end. UI rebuilt + restarted twice (E.6 backend, E.7 surfaces). pg never rebuilt this session — all changes were SQL-only.

## Open / carry-forward

- **Steward retry switch from `retry_guidance` to `retry_guidance_with_lessons`** — small surgery to 4c-steward-dispatch.sql; defer until next session or fold into the first Phase F commit.
- **Trust state needs to be exercised by real work_items** — synthetic smoke proves the machinery; the real validation comes when a work_item passes through verify maturity and trust_record_success fires through the actual gate path.
- **Stewards-UI nav is getting busy** — 11 routes now (Dashboard, Studies, Work items, Sessions, Watchman, Bridge, Graph, New work, Intents, Covenant, Sabbath, Lessons, Trust). Sidebar grouping (Substrate / Surfaces / Records) is the open-question note from the phase-e sub-spec — defer to Phase F or after.
- **bgworker payload._kind enum refactor** still pending from Phase D; will be more attractive when Phase F adds the council marker (6+ variants).
- **studies.file_path NOT NULL** still pending from Phase D.5 — pre-existing constraint that prevents promote_to_study from succeeding past the sabbath gate. Worth a separate small fix.
- **Phase F (Council) is next** when Michael's ready. Sub-spec at `phase-f-design.md`.

## Files touched

Repo:
- `projects/pg-ai-stewards/extension/5f-trust.sql` (new)
- `projects/pg-ai-stewards/extension/5f2-evaluate-trust.sql` (new)
- `projects/pg-ai-stewards/extension/5f3-gate-trust-check.sql` (new)
- `projects/pg-ai-stewards/extension/5f4-retry-with-lessons.sql` (new)
- `projects/pg-ai-stewards/extension/5f5-apply-gate-override.sql` (new)
- `projects/pg-ai-stewards/extension/src/lib.rs` (5 new extension_sql_file!)
- `projects/pg-ai-stewards/extension/Dockerfile` (5 new SQL files in COPY)
- `scripts/stewards-ui/api/trust.go` (new)
- `scripts/stewards-ui/api/api.go` (registerTrust)
- `scripts/stewards-ui/frontend/src/api.ts` (trust + override types + wrappers)
- `scripts/stewards-ui/frontend/src/views/Trust.vue` (new)
- `scripts/stewards-ui/frontend/src/views/WorkItemDetail.vue` (override modal)
- `scripts/stewards-ui/frontend/src/router.ts` (1 new route)
- `scripts/stewards-ui/frontend/src/App.vue` (nav)

Live containers:
- `pg-ai-stewards-dev`: 5 new SQL files live-applied via docker cp + psql.
- `pg-ai-stewards-ui`: rebuilt + restarted twice (E.6 backend, E.7 surfaces).
- `pg-ai-stewards-bridge`: restarted at session end.
- Soak: schedule_enabled=true at session end.
