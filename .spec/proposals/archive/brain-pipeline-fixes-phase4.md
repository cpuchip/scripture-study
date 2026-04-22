# Brain Pipeline Fixes — Phase 4: Execution Reliability & UX Polish

**Binding problem:** The pipeline's execution phase times out on complex tasks (10-minute hard cap killed a complete, passing implementation), gives no indication of agent liveness during long thinking pauses, and several UX artifacts from the E2E test confuse or mislead the steward.

**Created:** 2026-04-10
**Source:** [E2E walkthrough findings](../../.spec/scratch/debug-pipeline-e2e/main.md)
**Extends:** [brain-pipeline-fixes.md](brain-pipeline-fixes.md) (Phases 1–3.5, all shipped)
**Status:** Planned

---

## Success Criteria

1. Execution survives a 12+ minute task that is actively working (no timeout while agent is producing output)
2. Steward can see the agent is alive during long generation pauses (animated indicator + elapsed time)
3. Failure badge resets on successful verify
4. Timeout failure messages distinguish "work complete but timed out" from "nothing produced"
5. Tool events show enough detail to understand what's happening (file paths, query snippets)

---

## Phase 4: Activity-Based Execution Timeout (P0)

**Problem:** Go's `context.WithTimeout` is a hard wall — cannot extend. The 10-minute limit killed a successful execution that was actively creating files.

**Solution:** Replace the fixed-deadline context with an **inactivity-based** timeout. The context cancels only when the agent goes silent — no hard wall clock. The human (or a future orchestrator) is the circuit breaker, not an arbitrary timer.

**Design rationale:** Michael's experience includes 8-hour iterative sessions and 14-hour monitoring runs where the agent worked continuously. A wall clock cap would kill legitimate long-running work. The real signal for "something is wrong" is *silence* — no SDK events flowing. If events are flowing, the agent is alive and productive.

**Architecture:**

```
pool.go — StartTask() returns ctx + touch func
  └── activityContext wraps context.Background() with:
      • 5-minute inactivity deadline (resets on each touch)
      • No wall clock cap (manual cancel is the circuit breaker)
      • cancel() for manual cancellation

agent.go — AskStreaming() already has touchEvent() on every SDK event
  └── Accept OnActivity callback in AgentConfig
  └── Call it from touchEvent() so the pool can reset the inactivity timer

execute.go — runExecute() wires them together
  └── Pass OnActivity = touch through AgentConfig
```

**Implementation details:**

1. **New type `activityContext` in pool.go:**
   - Embeds `context.Context` (the parent, via `context.WithCancel`)
   - `*time.Timer` for inactivity deadline (5 min, resets on `Touch()`)
   - `Touch()` method resets the inactivity timer
   - When timer fires → cancel the derived context
   - `Done()` channel from the derived context

2. **pool.go changes:**
   - `ExecutionTimeout` → `InactivityTimeout = 5 * time.Minute`
   - Remove hard wall clock cap
   - `StartTask()` returns `(ctx context.Context, touch func())` instead of just `ctx`
   - `runningTask` gains a `touch func()` field

3. **agent.go changes:**
   - `AgentConfig` gains `OnActivity func()` callback
   - In `AskStreaming()`, inside `touchEvent()`, call `a.config.OnActivity()` if non-nil

4. **execute.go changes:**
   - `ctx, touch := p.pool.StartTask(entry.ID, "execute")`
   - Set `agentCfg.OnActivity = touch`

**Why 5 minutes for inactivity:** The 8-minute "thinking pause" observed in E2E wasn't inactivity — the SDK sends `AssistantReasoningDelta` events every few seconds. `touchEvent()` fires continuously. A 5-minute gap with zero events means the connection is genuinely stalled. Generous enough to avoid false positives, short enough to catch real stalls.

**Files:** `pool.go`, `agent.go`, `execute.go`
**Tests:** Unit test `activityContext` — verify inactivity fires, verify touch resets, verify cancel works.

---

## Phase 4a: BUG-3 — Reset Failure Count on Verify (P0)

**Problem:** `Verify()` on all-pass doesn't call `ResetFailureCount()`. Red "🔴 2 failures" badge persists forever after verify.

**Fix:** One line in `execute.go`, in the `allPassed` block of `Verify()`:
```go
p.store.DB().ResetFailureCount(entry.ID)
```
After `SetMaturity(entry.ID, "verified", ...)`.

**Files:** `execute.go`

---

## Phase 4b: Liveness Indicator + Elapsed Timer (P0)

**Problem:** Static yellow dot during 8-minute thinking pause looks frozen.

**Backend:**
- Include `execution_started_at` timestamp in `execution.started` WebSocket event
- Emit `execution.heartbeat` events every 30s from watchdog loop via new `OnHeartbeat` callback

**Frontend:**
- Replace static dot with CSS-animated pulsing dot
- Add elapsed-time counter: `"● Agent is executing... (3m 42s)"`
- Compute from `execution_started_at` using `setInterval`

**Files:** `agent.go`, `execute.go`, frontend executing badge component

---

## Phase 4c: Tool Event Detail (P1)

**Problem:** Tool events show "view", "sql", "create" with no arguments.

**Fix:** Extract summary from `OnToolCall` args (already available):
- `create` / `view` → file path
- `sql` → first 80 chars of query
- Others → first arg as string, truncated to 80 chars

Include `"detail": summary` in `execution.tool` WebSocket event. Frontend shows detail next to tool name.

**Files:** `execute.go`, frontend event list component

---

## Phase 4d: Smart Failure Messages (P1)

**Problem:** Same "Execution failed" message whether agent created 35 files or zero.

**Fix:** In `runExecute()` failure path, check `agent.WrittenFiles()`:
- Files exist → "Execution timed out, but the agent created N files. The work may be complete — check the workspace before retrying."
- No files → "Execution failed: {error}. No files were created."

Also: attempt `commitAfterExecution` even on timeout failures (preserve work in git).

**Files:** `execute.go` (failure path)

---

## Phase 4e: Failure Count Reset (P1)

**Problem:** `failure_count` only resets on successful execution. Can't reset for testing or manual recovery.

**Fix:** Add `POST /api/entries/{id}/reset-failures` endpoint. Wire to frontend as "reset" link next to failure badge.

Explicit > implicit. Option B (auto-reset on maturity change) would surprise users.

**Files:** API handler, frontend failure badge component

---

## Phase 4f: "Mark Complete" Button Visibility (P2)

**Problem:** "✓ Mark Complete" shows at `verified + done` state. Redundant.

**Fix:** Hide when `status == "done"`.

**Files:** Frontend entry detail component

---

## Phase 4g: Title Strikethrough at Done (P2)

**Problem:** Strikethrough on long titles is visually heavy.

**Fix:** Replace with opacity reduction (0.6) + "Done" badge. Communicates completion without harming readability.

**Files:** Frontend CSS/component

---

## Phase 4h: BUG-1 — Post-Timeout Tool Execution (P1, Investigate)

**Problem:** `create` tool executed 25s after context deadline. SDK doesn't abort in-flight requests immediately.

**Risk:** Low after Phase 4 (hard timeouts become rare). But files created post-timeout are invisible to `WrittenFiles()` and git commit.

**Mitigation:** In `commitAfterExecution`, scan for recently-modified files beyond `WrittenFiles()` reports. Log warning when tools fire after deadline.

**Investigation:** Does the SDK check `ctx.Done()` between tool calls? May need to file SDK issue.

**Files:** TBD after investigation

---

## Phase 4i: BUG-2 — Success Notify Label (P2)

**Problem:** Success path sends `{"maturity": "executing"}` in notify — technically correct but confusing.

**Fix:** Change to `{"maturity": "executing", "route_status": "your_turn"}` to make clear what changed.

**Files:** `execute.go`

---

## Phase 4j: Verify UI Without Agent Message (P1)

**Problem:** Verify UI triggers on agent "Execution complete" message, not on maturity state. Manual `maturity: executing` has no Verify button.

**Fix:** Show Verify button when `maturity == "executing" && route_status == "your_turn"`, regardless of message presence.

**Files:** Frontend entry detail component

---

## Execution Order

| Phase | Priority | Effort | Dependencies |
|-------|----------|--------|-------------|
| **4a** BUG-3 fix | P0 | 5 min | None |
| **4** Activity timeout | P0 | 1-2 hours | None |
| **4b** Liveness indicator | P0 | 30 min | Helpful after Phase 4 |
| **4d** Smart failure msgs | P1 | 20 min | None |
| **4c** Tool event detail | P1 | 20 min | None |
| **4e** Failure count reset | P1 | 20 min | None |
| **4j** Verify UI state-based | P1 | 20 min | None |
| **4h** Post-timeout tools | P1 | Investigate | None |
| **4f** Button visibility | P2 | 5 min | None |
| **4g** Strikethrough style | P2 | 10 min | None |
| **4i** Success notify label | P2 | 5 min | None |

Start with 4a (trivial) and 4 (highest impact), then sweep through P1 and P2.
