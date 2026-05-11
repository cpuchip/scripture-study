---
date: 2026-05-11
session_kind: build
workstream: WS5
substrate_phase: D
commits: [2922551, 3410796, 50f0379, 2e642cd, 7f054e0, 143ad76, 56f859b]
cost_usd: 0.005
---

# Substrate Phase D — shipped end-to-end on dev in one session

## What happened

Michael said "Lets do phase D and git commit at clean checkpoints." Same cadence as Phase C — seven commits in build order, each with a smoke test, all in one session.

## What shipped

**D.1 (2922551)** — schema. `pipelines.sabbath_enabled` + `atonement_enabled` (study-write + study-write-qwen seeded sabbath ON per D-D1; atonement stays opt-in everywhere per D-D2). `stewards.lessons` audit ledger mirroring `gate_decisions` shape per the 2026-05-11 ratification — kind in (principle | decision | lesson | sabbath_reflection), with ratified_at + ratified_by + promoted_to fields for D-D3 human curation. `lessons_recent_ratified` view keyed on (pipeline_family, current_stage) ready for Phase E's retry composer. `sessions_kind_check` extended to add 'sabbath' + 'atonement' (mirrors 5c's pattern adding 'gate'). `work_items.sabbath_completed_at` column for D.5's promotion gate.

**D.2 (3410796)** — Sabbath. Template asks for {reflection, carry_forward, surprise} as JSON-only output. `sabbath_dispatch(uuid)` validates pipeline.sabbath_enabled, composes a session with kind='sabbath', enqueues a chat with `_sabbath=true` + `tools_disabled=true` markers. `apply_sabbath_result` writes the lesson row (kind='sabbath_reflection') with composed content + raw_response, then timestamps `work_items.sabbath_completed_at`.

**D.3 (50f0379)** — Atonement. Template asks for {principles_to_record, decisions, lessons} arrays with explicit guidance to be sparse ("three lessons that survive scrutiny beat thirty that get pruned"). `atonement_dispatch` includes the last 20 `steward_actions` rows in the prompt context for the work_item. `apply_atonement_result` writes one `stewards.lessons` row PER ITEM across the three arrays, all `ratified_at = NULL`. Returns the total count inserted. Smoke verified: synthetic apply with `{P1,P2 / D1 / L1,L2,L3}` produced 6 unratified rows across the 3 kinds.

**D.4 (2e642cd)** — bgworker auto-fire. Two new payload markers (`_sabbath`, `_atonement`) added to the existing 3-marker switch from Phase 5b. Same shape — error logged, never propagated. Bgworker now handles 5 marker variants total: `_gate_eval` / `_scenarios_gen` / `_verify` / `_sabbath` / `_atonement`. Worth refactoring to a `payload._kind` enum once the 6th lands (council in Phase F) — flagged as carry-forward.

**D.5 (7f054e0)** — promotion gate + triggers. `work_item_promote_to_study` now raises `check_violation` if pipeline.sabbath_enabled AND `sabbath_completed_at IS NULL`. The error message points the human at `stewards.sabbath_dispatch(uuid)` to fix. `apply_gate_decision` (Phase 5a) extended: when action='advance' lands maturity at 'verified' on a sabbath-enabled pipeline AND sabbath hasn't already run, fires `sabbath_dispatch` automatically (errors swallowed via NOTICE so the gate decision still applies cleanly — sabbath is a side effect, not a blocker). `apply_verify_result` (Phase 5b) extended with the same pattern. New helper `maybe_enqueue_atonement(uuid)` for the steward quarantine path — no-op if pipeline.atonement_enabled is false.

**D.6 (143ad76)** — backend `api/lessons.go`. Three endpoints: `/api/lessons/list?kind=&ratified=` (filters), `/api/lessons/ratify` (POST sets ratified_at + by + optional promoted_to), `/api/sabbath/list?pipeline=` (pulls reflection / carry_forward / surprise out of raw_response jsonb so the UI doesn't need to parse it). Both endpoints registered in api.go.

**D.7 (56f859b)** — Vue surfaces. `/sabbath` route shows reflection cards with kind-colored sidebars (carry_forward in emerald, surprise in amber). `/lessons` route groups by kind with tinted badges and Approve / Approve & promote → .mind/principles.md / Approve & promote → .mind/decisions.md buttons per row. Filter dropdowns for kind + ratified state. Ratify-as text input lets the user set the ratifier name (defaults to "michael"). Router + App.vue nav extended.

## End-to-end e2e (D.4 commit)

Triggered `sabbath_dispatch` on the existing `gate-test-e2e-1` work_item. Real LLM dispatch:
- 1 chat round (qwen3.6-plus), ~22 seconds
- tools_disabled correctly honored — no research loop
- bgworker auto-fired `apply_sabbath_result` on completion
- `stewards.lessons` row #7 written with the full {reflection, carry_forward, surprise} JSON in raw_response
- `work_items.sabbath_completed_at` timestamped
- Cost ~$0.005

The model's reflection was genuinely useful. Because gate-test-e2e-1 was in a degraded state (the C.6 test had the model "surface" because stage_results was empty/error-shaped), the sabbath model correctly diagnosed it: *"This work produced a pipeline failure report rather than a scriptural outline, revealing a critical context drop during the agent handoff... The friction came not from the theological complexity of D&C 130:18-19, but from a broken spec-to-agent bridge that left the model operating blind."* Carry forward: *"Always verify that the complete spec, binding question, and context are successfully injected into downstream agents before triggering pipeline stages."* That's a real lesson from a real failure mode — exactly the substrate's value proposition.

## Surprises

**The Lesson #3 fix paid for itself immediately.** The pg rebuild for D.4's bgworker change normally would have meant manually `CREATE FUNCTION`-ing the YAML helpers + sabbath_dispatch + atonement_dispatch + apply_sabbath_result + apply_atonement_result. Instead the PostToolUse hook fired automatically after `docker compose build pg`, refreshed the 7 pg_extern functions in 1 second. Zero friction. The build flow is now: edit code → `docker compose build pg` → `docker compose down && up -d pg ui` → done. No manual intervention.

**Sabbath produced its own form of usefulness on a degraded test.** The model didn't get derailed by the bad input — it correctly identified the pipeline failure and produced an actionable carry-forward. That's the model + substrate working together: substrate gives the model intent + covenant + stage_results context; model produces structured judgment. Tools-off means no research loop pollutes the focus.

**The promotion gate fires the way the proposal said it would.** Synthetic test: study-write/review/completed work_item with NULL sabbath_completed_at → promote refuses with the exact error message the proposal predicted. Set sabbath_completed_at → promote moves past the gate. The discipline is endings recorded. (The studies.file_path NOT NULL caught the second test before insert — that's a pre-existing schema constraint unrelated to D.5; flagged as future cleanup.)

## Process / covenant

Seven commits, each with a smoke test. Same cadence as Phase C. No stewardship-shaped surprises this session — D.5's apply_gate_decision rewrite caught the right adjacent integration (firing sabbath_dispatch from inside the existing gate apply rather than as a separate trigger), but that was the expected design from the sub-spec, not an unexpected discovery.

Cost: $0.005 for the D.4 e2e Sabbath test. All other work was schema + code + Vue.

Soak paused at session start, re-enabled at session end. Bridge restarted at session end. UI rebuilt + restarted twice (D.6 and D.7 each got their own image cycle).

## Open / carry-forward

- **bgworker payload._kind enum refactor** — 5 marker variants now (gate_eval, scenarios_gen, verify, sabbath, atonement); when council lands in Phase F it'll be 6+. Worth collapsing to a single `_kind` enum field. Two-line cleanup.
- **studies.file_path NOT NULL** is unrelated to Phase D but caught during D.5 smoke. Pre-existing constraint that means promote_to_study has been failing silently for a long time. Worth a separate fix (probably make file_path nullable or populate it from the slug).
- **Atonement end-to-end** wasn't tested with a real LLM dispatch — just synthetic apply. The atonement path is symmetric with sabbath so the auto-fire wiring is verified by D.4's sabbath test, but a real Atonement-on-quarantine test is worth running once we have a quarantined work_item to sacrifice.
- **Steward integration of maybe_enqueue_atonement** — the helper exists but the steward's quarantine path doesn't yet call it. Two-line addition to `steward_dispatch.sql` (phase 4c). Could fold into Phase D or wait for the next quarantine event.
- **File-write mechanism for promoted lessons** — UI promote buttons set promoted_to in the lessons row but nothing actually writes to .mind/principles.md or .mind/decisions.md. Per Phase D sub-spec V.6, this is a pending-write pattern (substrate emits a "pending file write" record, sidecar/next git commit materializes). Not blocking for Phase D shipping; flagged.
- **Phase E is next** when Michael wants to keep going. Sub-spec already drafted (`phase-e-design.md`) — trust ladder + lessons-in-retry-context.

## Files touched

Repo:
- `projects/pg-ai-stewards/extension/5e-lessons-and-pipeline-flags.sql` (new)
- `projects/pg-ai-stewards/extension/5e2-sabbath.sql` (new)
- `projects/pg-ai-stewards/extension/5e3-atonement.sql` (new)
- `projects/pg-ai-stewards/extension/5e4-promotion-gate-and-triggers.sql` (new)
- `projects/pg-ai-stewards/extension/src/bgworker.rs` (2 marker variants added)
- `projects/pg-ai-stewards/extension/src/lib.rs` (4 new extension_sql_file!)
- `projects/pg-ai-stewards/extension/Dockerfile` (4 new SQL files in COPY)
- `scripts/stewards-ui/api/lessons.go` (new)
- `scripts/stewards-ui/api/api.go` (registerLessons)
- `scripts/stewards-ui/frontend/src/api.ts` (lessons + sabbath types + wrappers)
- `scripts/stewards-ui/frontend/src/views/Sabbath.vue` (new)
- `scripts/stewards-ui/frontend/src/views/Lessons.vue` (new)
- `scripts/stewards-ui/frontend/src/router.ts` (2 new routes)
- `scripts/stewards-ui/frontend/src/App.vue` (nav)

Live containers:
- `pg-ai-stewards-dev`: 4 new SQL files live-applied; bgworker rebuilt + auto-refresh fired by Lesson #3 hook.
- `pg-ai-stewards-ui`: rebuilt + restarted twice (D.6 backend, D.7 surfaces).
- `pg-ai-stewards-bridge`: restarted at session end.
- Soak: schedule_enabled=true at session end.
