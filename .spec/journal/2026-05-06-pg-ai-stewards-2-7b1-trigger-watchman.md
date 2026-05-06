# pg-ai-stewards Phase 2.7b.1 — Trigger-driven Watchman automation

*2026-05-06 (Claude Code, Opus 4.7)*

## What this session was

The 2.7b unblocker. The journal entry from earlier the same day flagged
2.7b (Watchman bgworker automation) as the #1 carry-forward. After the
corpus re-import discussion (DB had drifted to 143 docs vs the
documented 359), Michael split 2.7b into b.1–b.4 and chose Option B —
pure SQL + `AFTER UPDATE` trigger over a Rust-side result harvester —
so the bgworker stays generic and harvest happens transactionally.

## What shipped

**Phase 2.7b.1 — trigger-driven Watchman automation.**

### Files

- `extension/2-7b1-watchman-automation.sql` — live-DB migration
  (applied via `psql -f`).
- `extension/src/lib.rs` — sixth `extension_sql_file!` reference,
  `requires = ["create_watchman_pass"]`.
- `extension/verify-2-7b1-inverse.sql` — pure-SQL inverse hypothesis
  test (no model tokens needed).
- `extension/verify-2-7b1.log` — captured CLI output of the 5-doc
  real-model verification.
- `cmd/stewards-cli/main.go` + `internal/show/show.go` — four new
  subcommands: `pass-now`, `passes`, `pass-detail`, `config show|set`.
- `projects/pg-ai-stewards/phases.md` — Phase 2.7 section added with
  2.7a marked shipped, 2.7b sub-phase table, 2.7b.1 done-when +
  verification record. Phase 3a's reference to "2.7b unblocker"
  preserved.

### SQL surface

- `stewards.watchman_passes` — one row per pass; counters
  (`doc_count_done`, `tokens_in/out`, `verdict_counts`) advanced by
  the trigger.
- `stewards.watchman_config` — singleton (id=1) with cron schedule,
  default provider/model/agent, token_budget, dirty_threshold,
  idle_threshold_hours, last_pass_at. 2.7b.2 reads it; 2.7b.1 just
  creates the row.
- `stewards.watchman_pass_start(...)` — pulls top-N dirty docs,
  composes user input via `watchman_input(slug)`, builds payload with
  `_watchman_pass_id` / `_watchman_slug` / `_watchman_actor` markers,
  enqueues kind='chat' rows. Returns the new pass_id.
- `stewards.advance_watchman_pass_counters(...)` — helper used by
  the trigger to roll up per-pass stats and auto-mark `completed`.
- `stewards.handle_watchman_chat_completion()` + trigger
  `watchman_harvest_completion` — `AFTER UPDATE OF status` on
  `work_queue` with WHEN-clause prefilter for cost. Reads assistant
  message, strips `\`\`\`json` fences, casts to jsonb, validates
  verdict against the 5-element enum, calls `record_verdict` (and
  `record_finding` if non-clean). Every harvest call wrapped in
  `BEGIN...EXCEPTION` so a bug in the harvester never breaks the
  bgworker's status flip.
- `stewards.watchman_pass_summary` view — convenience for CLI listing.

### Verification (3 layers)

**Layer 1 — smoke test (1 doc, real model).** Pass
`watchman-20260506T200536Z-9b2de6` on `ai-responsible-use-reflections`.
Elapsed 3m22s (opencode_go was unusually slow). Trigger harvested:
verdict=`skipped`, drift finding with severity=`medium`. Pass
auto-marked `completed`. Tokens: 3897 in / 6961 out.

**Layer 2 — inverse hypothesis (synthetic, no tokens).** 4 trials in
`verify-2-7b1-inverse.sql`:

| Trial | Setup | Got |
|-------|-------|-----|
| 1 | trigger present, drift+finding JSON | 1 verdict + 1 finding, completed |
| 2 | trigger DROPPED | 0 verdicts + 0 findings, in_progress |
| 3 | trigger restored, clean JSON | 1 verdict, completed |
| 4 | trigger present, malformed JSON | verdict=`skipped` with parse-error reasoning, no raise |

Trial 2 is the load-bearing test — proves the trigger is what's doing
the work, not some other side-effect.

**Layer 3 — 5-doc real-model verification.** Pass
`watchman-20260506T201415Z-b9056d` (`actor=verify-2-7b1`):

- 5/5 docs harvested in 7m45s. Tokens: 18902 in / 18677 out (under
  50k budget).
- Verdicts: 1 clean (`art-of-delegation`) + 4 skipped (kimi keeps
  surfacing the "I can't see external context" pattern from 3a).
- 3 findings recorded (2 drift, 1 synthesis).

## What was surprising

**The opencode HTTP 502 was the best test we got.** Doc 3
(`art-of-presidency`) hit `chat HTTP 502 Bad Gateway` mid-pass. The
trigger's error path recorded `verdict='skipped'` with the HTTP error
as reasoning, advanced the pass counters, and the rest of the pass
kept going. Trial 4's defensive path proven in the wild on the very
first 5-doc run. We didn't need to manufacture a failure to validate
the error path — opencode_go provided one.

**The architectural choice paid off immediately.** Going pure-SQL +
trigger meant zero Rust changes. The bgworker is still entirely
generic — knows nothing about Watchman. All semantics live in
`stewards.handle_watchman_chat_completion()`. When 2.7b.2 lands the
scheduler tick, the trigger keeps doing its job unchanged.

**Same bug, same fix didn't apply.** No same-bug-same-fix opportunities
surfaced this session — this was net-new code, not extending a
pattern. The closest analog was the 3a CLI orchestrator's polling
loop, which we deliberately preserved as a fallback rather than
replace.

## Known limitations / open

- **2.7b.2 — bgworker scheduler tick** is the next step. Should read
  `watchman_config` and call `watchman_pass_start()` on cron, pressure
  (dirty_queue > threshold), or idle (>48h no human session) triggers.
  Concrete next session.
- **2.7b.3 — per-pass token budget enforcement.** Currently the
  budget column is informational; nothing stops `watchman_pass_start`
  from enqueueing past it. Worth doing before a 7-day soak.
- **2.7b.4 — `regenerate_active_md()`** + 7-day soak.
- **AGE-QUIRKS PRs** still on the open list; not blocked, just
  awaiting bandwidth.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Phase 2.7b.2** — bgworker scheduler tick. Add a sibling 60s tick to the existing 500ms work-drain loop; SPI-call `watchman_pass_start` when due. |
| 2 | **Phase 2.7b.3** — per-pass token budget enforcement inside `watchman_pass_start` (stop enqueueing when projected tokens cross threshold). |
| 3 | **Phase 2.7b.4** — `regenerate_active_md()` + 7-day soak. |
| 4 | ws6 AGE upstream PRs (#2, #6, #7 are bug-candidate). |

## What's still solid

- Trigger-driven harvest is now the canonical Watchman path. Every
  watchman chat that completes lands a verdict in the same tx as the
  status flip. No race window, no Go polling, no result-harvest
  daemon to babysit.
- Sixth `extension_sql_file!` foldback complete in lib.rs. The
  auto-generated `pg_ai_stewards--0.2.0.sql` will pick up the new
  block on next `cargo pgrx schema` regen.
- The 3a Go-orchestrator path (`watchman pass`) is preserved for
  `--slug` single-doc repro and Go-side log visibility. Same SQL
  fixtures, different control loop. Not deprecated, just no longer
  the canonical path.
