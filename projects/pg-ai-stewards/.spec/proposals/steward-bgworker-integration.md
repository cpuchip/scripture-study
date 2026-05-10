---
title: Steward bgworker integration — wire steward_tick into the dispatcher loop + dispatch override
date: 2026-05-10
status: design sub-spec — ready to execute, requires container restart
parent: full-agentic-substrate.md, escalation-chain.md, cost-tracking.md
prereqs: 4a-cost-tracking.sql, 4a-escalation-chain.sql, 4a-steward.sql (all live on pg-ai-stewards-dev)
purpose: >
  Make the Watch→Diagnose→**Act**→Account loop actually act. The Account
  layer (steward_actions audit) works today; the Act layer (re-dispatch
  with escalated model + retry guidance) needs three pieces: (1) a
  bgworker tick that calls steward_tick periodically, (2) modification
  of work_item_dispatch_stage to honor work_items.model_override +
  provider_override, (3) a small steward_tick extension that resets
  status='pending' + sets the override + calls dispatch_stage on the
  retry_with_escalation action path.
---

# Steward bgworker integration

## I. Binding problem

`stewards.steward_tick()` (live on dev as of 2026-05-10) walks failed work_items, applies cost-cap + breaker + diagnosis logic, picks a model via `pick_model`, and writes a `steward_actions` row with `action='retry_with_escalation'` (or `'queue_for_opus'` / `'quarantine'` / `'defer_breaker_open'`). What it doesn't do: actually re-dispatch the failed stage on the new model.

Two reasons the Act layer didn't ship in the schema push:

1. **The bgworker doesn't call steward_tick yet.** `stewards_dispatcher_main` polls `work_queue` every 500ms and has hooks for the watchman scheduler + tool_dispatch reaper, but no steward tick. Adding one is a Rust change, which means rebuilding the extension binary and restarting the live container (with soak pause).

2. **`work_item_dispatch_stage` doesn't honor `model_override`.** The dispatcher reads model + provider from `pipeline_stage_lookup`, which returns the stage's static defaults. My `pick_model` and `work_items.model_override` are inert until the dispatcher consults them. This is a SQL function modification that can ship without container restart, but it's surgery on existing live code.

This sub-spec defines both pieces and the order to ship them.

## II. Success criteria

After this sub-spec ships:

1. **A failed work_item retries automatically.** Failure → steward_tick fires (~30s tick) → diagnose → escalate model via pick_model → set override + status='pending' → work_item_dispatch_stage uses override → new work_queue row with the escalated model → bridge dispatches → result.
2. **The escalation_queue handoff works.** GLM exhaustion → `__queue_for_opus__` sentinel → escalation_state='queued' → no auto-dispatch (waits for human/CLI claim).
3. **Cost-cap quarantine is honored.** A work_item that crosses cost_cap_micro is quarantined by steward_tick before any further dispatch.
4. **Circuit breaker prevents thundering-herd.** A pipeline/stage that fails 5× in a row trips the breaker; subsequent steward_tick passes defer until cooldown.
5. **No regression in normal dispatch path.** Successful dispatches (model_override IS NULL, no failures) behave identically to today.

## III. Three sub-pushes

Ship in this order; each is independently verifiable and rollback-safe.

### Push A — Modify work_item_dispatch_stage to honor overrides (SQL only, no restart)

Add `provider_override text` column to work_items (sibling of `model_override`).

Replace `stewards.work_item_dispatch_stage` with a new version. Three changes vs current:

1. Allow re-dispatch from `status='failed'` (currently only allows pending/awaiting_review). Required for the steward to trigger a retry.
2. Read `model = COALESCE(work_items.model_override, stage->>'model')`.
3. Read `provider = COALESCE(work_items.provider_override, stage->>'provider')`.

Idempotent migration; no data loss; no restart. Existing call sites (NewWork form, watchman) pass model_override=NULL so behavior is unchanged.

Risk: status='failed' re-dispatch could cause a NewWork-side bug if anything assumed only pending could dispatch. Mitigation: add an explicit `p_allow_failed_status boolean DEFAULT false` param so existing call sites stay safe; steward_tick passes true.

File: `4b-dispatch-override.sql`. Acceptance: existing watchman dispatch still works; SELECT calling work_item_dispatch_stage on a failed work_item with model_override set produces a work_queue row with the override model.

### Push B — Extend steward_tick to actually dispatch (SQL only, no restart)

Modify `steward_tick` to, on the `retry_with_escalation` path:

1. Reset `work_items.status = 'pending'`
2. Set `work_items.model_override = v_next_model`
3. Set `work_items.provider_override = 'opencode-zen'` (single-provider chain for now)
4. Increment `work_items.failure_count + 1` is already implicit via the action attempt
5. Call `stewards.work_item_dispatch_stage(work_item.id, retry_guidance, p_allow_failed_status := true)`
6. Capture the returned work_id into `steward_actions.details->>'dispatched_work_id'`

Wrap the override-set + dispatch in a sub-transaction so a dispatch failure rolls back the override (won't re-dispatch with stale override on next tick).

File: `4b-steward-dispatch.sql`. Acceptance: synthetically mark a work_item as failed with last_failure_reason='context deadline exceeded' and failure_count=1; SELECT steward_tick(); confirm a new work_queue row exists, model_override is set, status is 'pending'.

### Push C — Bgworker tick that calls steward_tick (Rust change, container restart)

Add a new tick handler in `extension/src/bgworker.rs` modeled on the existing `stewards_watchman_scheduler_check` pattern. New function `stewards_steward_tick_check()` that:

1. Runs every `STEWARD_TICK_INTERVAL_SECONDS` (env var, default 30)
2. Calls `SELECT stewards.steward_tick()`
3. Logs the count of actions taken
4. Errors are logged but don't crash the worker

Wire into the main poll loop alongside watchman_scheduler check.

Build artifacts:
1. `extension/src/bgworker.rs` — add ~40 lines for the tick handler
2. No SQL changes
3. No Dockerfile changes (the SQL files were added in 4a)

Container restart procedure (per pgrx-rust skill):
```sql
-- Pause soak first
UPDATE stewards.watchman_config SET schedule_enabled = false WHERE id = 1;
-- (wait for any in-flight passes to complete; check stewards.watchman_passes)
```
Then:
```bash
cd projects/pg-ai-stewards/extension
docker compose down && docker compose up -d
# wait for ready
```
Then re-enable soak:
```sql
UPDATE stewards.watchman_config SET schedule_enabled = true WHERE id = 1;
```

Acceptance: bgworker logs include `stewards: steward_tick processed N actions` lines. Synthetic failed work_item gets retried within 30s without manual intervention.

## IV. Total scope

- Push A: ~50 lines SQL, 1 file. ~30 min.
- Push B: ~80 lines SQL (replacing steward_tick body), 1 file. ~30 min including a smoke test.
- Push C: ~40 lines Rust, 1 file change, container rebuild + restart. ~30 min including soak pause + verify.

Total: ~90 min for full Act-layer activation. Each push verifiable independently before the next ships.

## V. What's NOT in this sub-spec

- Stewards-UI surfaces showing steward_actions, escalation queue, cost panels
- 3 new MCP tools for CLI-mediated escalation queue (work_item_escalation_list/claim/resolve)
- Bgworker handler for the escalation_queued state (it just sits in queued; UI/CLI claims it)

Those are independent next pushes after Push C ships.

## VI. Open question

Push C requires a container restart, which interrupts the watchman soak briefly. Worth doing on Sunday evening when nothing is queued, vs. mid-week when work might be in flight. Michael picks the timing.

(Push A + B are SQL-only, can ship anytime without restart.)
