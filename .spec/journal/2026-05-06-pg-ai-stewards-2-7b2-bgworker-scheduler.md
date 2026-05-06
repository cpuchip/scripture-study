# pg-ai-stewards Phase 2.7b.2 — bgworker scheduler tick

*2026-05-06 (Claude Code, Opus 4.7)*

## What this session was

Continuation after 2.7b.1 shipped earlier the same day. The user said
"Lets work on the next task 2.7b.2". I designed the scheduler with
decision logic in SQL (so all schedule semantics stay in one place
and Rust just polls), and Michael confirmed the Option-B "pure SQL +
trigger" architectural choice was right for this work too.

One open design call surfaced before coding: should `schedule_enabled`
default to `true` (the experiment is the point) or `false` (gated
autonomy)? Michael: "default to true — it's the point of the
experiment, but I appreciate you asking. it's the kind of setting
that costs money and I'll like the say on that." Defaulted true,
documented the master kill switch as "the load-bearing kill-switch
until 2.7b.3 lands."

## What shipped

**Phase 2.7b.2 — bgworker scheduler tick.** SQL decides; Rust polls.

### Files

- `extension/2-7b2-watchman-scheduler.sql` — adds 7 schedule columns
  to `watchman_config`, three new SQL functions:
  - `watchman_scheduler_inputs()` — observability helper, returns the
    live values feeding the decision (dirty count, hours since last
    pass, hours since last human session, in-progress pass id+age).
  - `watchman_should_fire()` — pure decision function. Returns
    `'cron'|'pressure'|'idle'|NULL`. All schedule semantics here.
  - `watchman_scheduler_fire()` — convenience wrapper used by the
    bgworker: calls `should_fire()`, if non-NULL calls
    `watchman_pass_start()` with `actor='scheduler'`. Returns the
    new pass_id (or NULL).
- `extension/src/lib.rs` — bgworker main loop now runs a 60s
  scheduler tick alongside the existing 500ms work-drain. New
  `check_watchman_schedule()` function calls
  `stewards.watchman_scheduler_fire()` via SPI and logs on fire.
  `last_sched: Option<Instant> = None` on startup forces an
  immediate first check. Added `Instant` to the `std::time` import.
- `extension/Dockerfile` — added `2-7b1-watchman-automation.sql` and
  `2-7b2-watchman-scheduler.sql` to the `COPY` directive in stage 1.
  Also added a comment about the requirement to update this list
  when adding new `extension_sql_file!` references.
- `cmd/stewards-cli/main.go` + `internal/show/show.go`:
  - `watchman config show` now displays scheduler fields (enabled,
    cron label, pass limit, min interval, preferred DOW/hour with
    name, both cooldowns, dirty/idle thresholds).
  - `watchman config set` accepts 7 new flags: `--enabled`,
    `--min-interval-hours`, `--preferred-dow` (-1=any, 0=Sun..6=Sat),
    `--preferred-hour` (-1=any, 0..23 UTC), `--pass-limit`,
    `--pressure-cooldown-hours`, `--idle-cooldown-hours`. Refactored
    `WatchmanConfigSet` to take a `[]WatchmanConfigSetField` slice
    so the parameter list stays sane.
  - New `watchman scheduler-status` CLI command — prints
    `should_fire()` decision and every input feeding it, with
    annotations like "(threshold 50 → pressure when ≥)" and
    "(now: 21)" so the user can see WHY the answer is what it is.
- `extension/verify-2-7b2-decision.sql` — pure-SQL verification of
  `watchman_should_fire()`, walks 9 trials by mutating config in
  place, then restores. No model tokens needed.

### Decision matrix

Order matters: pressure > cron > idle.

| Trigger | Fires when |
|---------|------------|
| (none) | `schedule_enabled = false` OR a pass started <1h ago is still in_progress |
| `pressure` | `count(dirty_queue) >= dirty_threshold` AND last_pass older than `schedule_pressure_cooldown_hours` |
| `cron` | last_pass older than `schedule_min_interval_hours` AND we're inside the preferred DOW + hour window (NULL = any) |
| `idle` | `idle_threshold_hours > 0` AND last_pass older than `schedule_idle_cooldown_hours` AND no `kind='chat'` session in N hours |

### Verification (3 layers)

**Layer 1 — pure-SQL decision verification (9 trials, no model tokens).**

| # | Setup | Result |
|---|-------|--------|
| 1 | `schedule_enabled = false` | `NULL` ✓ |
| 2 | dirty heavy, past cooldown | `pressure` ✓ |
| 3 | dirty_threshold = 9999 (suppresses pressure), DOW/hour mismatch | `NULL` ✓ |
| 4 | preferred DOW/hour = NULL + min_interval=0 | `cron` ✓ |
| 5 | min_interval=168, last_pass 12h ago, dirty under threshold | `NULL` ✓ |
| 6 | idle_threshold_hours=1, idle_cooldown=1, no human sessions | `idle` ✓ |
| 7 | idle_threshold_hours=0 | `NULL` ✓ |
| 8 | inflight pass <1h old | `NULL` ✓ (don't pile up) |
| 9 | inflight pass >1h old (90 min) | `pressure` ✓ (allowed) |

**Layer 2 — silent operation with `schedule_enabled=false`.** Rebuilt
the container, started fresh with disabled scheduler. After 80s of
the bgworker running, zero scheduler-fired passes. Logs quiet —
proves the tick is silent on no-op decisions.

**Layer 3 — live end-to-end fire.**

```
21:36:21 UTC — bgworker started fresh (last_sched=None)
21:36:22 UTC — first scheduler tick → should_fire returns NULL
               (schedule still disabled at that exact instant)
21:36:30 UTC — human flipped schedule_enabled=true
21:37:22 UTC — second scheduler tick (60s after first) →
               should_fire returns 'pressure', fires a pass
21:37:22 UTC — bgworker logs: "stewards: scheduler fired Watchman
               pass: watchman-20260506T213722Z-0705d4"
21:37:22 UTC — pass row created with trigger='pressure',
               actor='scheduler', doc_count_planned=5 (from
               schedule_pass_limit=5)
```

The 5 chats then dispatched normally; the 2.7b.1 trigger harvested
each verdict. Nothing about 2.7b.1 needed to know about 2.7b.2.

## What was surprising

**Two build-time discoveries, both proactively caught.**

1. **`Spi::connect` is read-only.** I initially used `Spi::connect`
   in `check_watchman_schedule()`, reasoning "the SPI client only
   does a SELECT." But the SQL function being SELECTed
   (`watchman_scheduler_fire`) does INSERTs internally, and PG's SPI
   propagates the read-only flag down through nested calls. Switched
   to `Spi::connect_mut` proactively after re-reading
   `process_one_pending` and the reaper, both of which use
   `connect_mut`. Pattern-matching against the existing code
   surfaced this before it failed.
2. **Dockerfile `COPY` lists SQL files explicitly.** My first build
   failed with `couldn't read 'src/../2-7b2-watchman-scheduler.sql'`.
   Looking at the Dockerfile, the COPY directive at line 47-48 lists
   each SQL file by name. Adding a new `extension_sql_file!` in
   lib.rs requires updating that list. Worse: 2-7b1 was *also* not
   in the list — it had been folded into lib.rs in the 2.7b.1
   session but never docker-rebuilt because the live-DB migration
   pattern lets the SQL file land without a docker rebuild. Both
   added in one Dockerfile edit. Added a TODO-style comment so
   future-me sees it.

The lesson generalizes: when there are TWO mechanisms for getting
SQL into the DB (live-applied vs. baked-into-image), they can drift.
Today's docker rebuild was the first time we'd built since 2.7b.1
shipped, so the latent debt surfaced now. A pre-commit check that
the Dockerfile `COPY` lines mention every `extension_sql_file!`
reference in lib.rs would be cheap insurance.

## Cost discipline

Default `schedule_enabled = true` per Michael's call ("the
experiment IS the point"), but with structural cost guards:

- `schedule_pass_limit = 5` per pass
- `schedule_pressure_cooldown_hours = 1` between pressure passes
- `schedule_min_interval_hours = 168` between cron passes (weekly)
- in_progress guard: no new fire while a pass <1h old is still running
- master kill switch via `schedule_enabled`

No token-budget enforcement inside `watchman_pass_start` yet — that's
2.7b.3. After today's verification I disabled the scheduler so it
doesn't loop with the dirty corpus until 2.7b.3 + 2.7b.4 land.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Phase 2.7b.3** — per-pass token budget enforcement inside `watchman_pass_start` (stop enqueueing when projected tokens cross threshold). Becomes the second load-bearing cost guard alongside the master switch. |
| 2 | **Phase 2.7b.4** — `regenerate_active_md()` + 7-day soak with `schedule_enabled=true`. Trend line for tokens-per-day should decline as the corpus stabilizes. |
| 3 | **Pre-commit hygiene** — script that diffs `extension_sql_file!` references in lib.rs vs. the Dockerfile `COPY` list. Five-minute write; saves a build cycle every time. |
| 4 | **Auto-regenerate `pg_ai_stewards--0.2.0.sql`?** That file is auto-generated by `cargo pgrx package` inside the docker builder, but the repo copy may drift. Not blocking; documented for whoever needs it. |
| 5 | ws6 AGE upstream PRs (#2, #6, #7 are bug-candidate). |

## What's still solid

- The trigger-driven harvest path (2.7b.1) just keeps working when
  the scheduler (2.7b.2) fires. No coordination needed because
  payload markers (`_watchman_pass_id`, `_watchman_slug`,
  `_watchman_actor`) are the contract; both sides read/write the
  same table columns.
- "All schedule semantics in SQL" was the right call. When I was
  testing trial 8 (in-flight pass blocks new fires), I just adjusted
  the SQL `should_fire` function and re-ran without recompiling Rust.
  The Rust scheduler tick is 30 lines and has nothing to maintain.
- The `watchman scheduler-status` CLI is the right diagnostic
  surface — when something doesn't fire, you can see immediately
  whether it's enabled, what the inputs are, and which gate stopped
  it. Will earn its keep during the 7-day soak.
