# pg-ai-stewards Phase 2.7b.3 — per-pass token budget enforcement

*2026-05-06 (Claude Code, Opus 4.7)*

## What this session was

Third sub-phase landed in the same day. After 2.7b.2's scheduler
proved itself by firing a real pass on a 350+ dirty corpus, the
remaining cost guardrail was per-pass token budgeting. Pre-2.7b.3,
`schedule_enabled=false` was the only kill switch — turning the
scheduler back on for the soak (2.7b.4) without a budget would have
been a runaway risk.

## What shipped

**Phase 2.7b.3 — per-pass token budget enforcement.** `token_budget`
on `watchman_passes` was informational; now it's load-bearing.

### Files

- `extension/2-7b3-watchman-budget.sql` — adds `budget_stopped`
  column + `estimate_chat_tokens(slug)` function + replaces
  `watchman_pass_start()` with a budget-aware version + updates the
  `watchman_pass_summary` view.
- `extension/src/lib.rs` — eighth `extension_sql_file!` reference.
- `extension/Dockerfile` — added `2-7b3-watchman-budget.sql` to the
  COPY list. Image rebuild deferred to next batched change.
- `cmd/stewards-cli/internal/show/show.go`:
  - `WatchmanPasses` adds a BUDGET column showing `ok` or `STOPPED`.
  - `printWatchmanPassDetail` shows a ⚠ warning on the tokens line
    when `budget_stopped=true`.
- `extension/verify-2-7b3-budget.sql` — 4-trial SQL verification
  with a `pg_temp.abort_test_pass(pass_id)` helper that errors
  pending work_queue rows before the bgworker can dispatch them.
  Zero model tokens spent during verification.

### Estimation formula

```
estimate(slug) = chars(watchman_input(slug)) / 4   -- input tokens
              + 1500                                -- system + persona overhead
              + avg(verdicts.tokens_out, last 30d)  -- output (3500 fallback if cold)
```

Per-doc estimates ranged from 7700 (small studies) to 14500 (large
proposal docs) across the live corpus. The 1500 overhead constant is
empirical from compose_system_prompt; the 4-chars/token ratio is the
standard rough heuristic.

### Enforcement

In `watchman_pass_start`, before each enqueue:

```sql
v_estimate := stewards.estimate_chat_tokens(v_slug);
IF v_planned_tokens + v_estimate > v_budget THEN
    v_budget_stopped := true;
    EXIT;
END IF;
-- ... enqueue, increment v_planned_tokens by v_estimate
```

Stricter than "always allow at least one doc": if the FIRST doc's
estimate exceeds budget, refuse to enqueue (empty pass,
`budget_stopped=true`). Honest signal that the budget is unworkable.

The estimate is also written into the work_queue payload as
`_watchman_estimate`. This isn't read by anything yet but enables
future estimate-vs-actual analysis to refine the formula.

### Verification (4 SQL trials, zero tokens)

| # | Budget | Result | Why |
|---|--------|--------|-----|
| 1 | 1000 | 0 planned, `budget_stopped=true`, `status=completed` | First doc estimate ~9748 alone exceeds 1000 |
| 2 | 10000 | 1 planned, `budget_stopped=true` | First doc 9748 fits; +second 8794 → 18542 stops |
| 3 | 25000 | 2 planned, `budget_stopped=true` | 9748 + 8794 = 18542 fits; +14400 → 32942 > 25000 stops |
| 4 | 999999 | 5 planned (limit), `budget_stopped=false` | All 5 fit, hit `p_limit` not budget |

Each trial called `watchman_pass_start`, observed
`(doc_count_planned, budget_stopped, status)`, and aborted the test
pass via `pg_temp.abort_test_pass(pass_id)` which marks pending
work_queue rows errored. The bgworker never sees them because they
flip from `pending` → `error` directly without `in_progress` ever
appearing.

## What was surprising

**`CREATE OR REPLACE VIEW` requires identical column order.** Tried
to insert `budget_stopped` between `token_budget` and `actor` —
psql rejected with `cannot change name of view column "actor" to
"budget_stopped"`. PostgreSQL only lets you APPEND columns when
replacing a view. Moved `budget_stopped` to the end. Minor, but
worth knowing.

**`watchman_pass_start` was already taking a `p_token_budget`
parameter from 2.7b.1; nothing was using it.** I noticed it during
the rewrite and the earlier code had been "informational" through
the entire 2.7b.1 session. The signature just worked when I changed
the body — no callers needed updating.

**Estimates are surprisingly accurate.** The 5-doc scheduler-fired
pass (2.7b.2 verification) had actual avg ~8600 tokens/doc. The
estimator computes 7961 for the same docs. 7% under, well within
the noise of provider-side variance. Good enough for v1.

## Why the master kill switch is no longer load-bearing

Pre-2.7b.3, on a corpus of 350+ dirty docs, leaving
`schedule_enabled=true` overnight could have triggered:
- pressure fire every 1h (cooldown), each consuming 5×~8600 tokens
- 24h × 5 docs/h × 8600 tok/doc = ~1M tokens/day uncapped

Post-2.7b.3, that same setup with default `token_budget=50000` per
pass caps at:
- pressure fire every 1h, each capped at 50000 tokens
- 24h × 50000 tok/pass = 1.2M tokens/day worst case
- BUT in practice each pass plans ~5-6 docs (estimates ~8000 each)
  and most verdicts come back `clean` → docs leave the dirty queue
- Within a few hours, dirty count drops below `dirty_threshold=50`
  and pressure stops firing entirely; only weekly cron remains

The scheduler is now safe to leave on. The kill switch is a kill
switch, not the only kill switch.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Phase 2.7b.4** — `regenerate_active_md()` SQL fn that produces a markdown rendering of the workstream/project state from graph data, plus the 7-day soak with `schedule_enabled=true`. Soak proves the trend line for tokens-per-day declines as the corpus stabilizes. |
| 2 | **Image rebuild** — fold lib.rs has 2-7b3 wired but no rebuild yet. Next batched change picks up the cumulative foldback. |
| 3 | **Estimate calibration** — write a small query that compares `_watchman_estimate` (in payload) vs. actual tokens recorded by the trigger. Use it to tune the formula constants if drift exceeds 20%. |
| 4 | ws6 AGE upstream PRs (#2, #6, #7 are bug-candidate). |

## What's still solid

- Three independent cost guards now stack: per-pass `p_limit`, the
  60s scheduler tick interval (only one fire per minute max),
  per-pass `token_budget`, and `schedule_enabled`. Even if one
  fails, the others contain the blast radius.
- Pure-SQL verification with `abort_test_pass` is the right pattern
  for testing functions that *enqueue* work — we don't have to
  spend tokens to verify enqueue-time logic. Reuse for 2.8 LLM-
  inferred edges when that ships.
- The Restoration discernment frame holds at the substrate level:
  the scheduler proposes work, the budget caps cost, the human
  flips `schedule_enabled` to grant or revoke autonomy. Nothing
  runs without consent.
