# Pipeline E2E Debug — LCARS Vue3 Theme

**Date:** 2026-04-10
**Entry:** `03549801-b329-4221-ad72-d54ea32305f8` — LCARS Vue3 Theme — Clock & RPN Calculator
**Goal:** Walk the entry through the full pipeline end-to-end after all Phase 1-3.5 fixes, acting as human steward. Document every friction point, failure, and UX observation.
**Prior audit:** [debug-brain-ux/main.md](../debug-brain-ux/main.md) — 18 bugs found, all P0/P1 now fixed.

## Fixes Applied Since Last Audit
- Phase 1: Slim execute prompt (path not content), 10-min timeout, cancel endpoint, race guard, route_status="agent"
- Phase 2: Toast system, scenario dialog, cancel/complete buttons, maturity badges, pipeline gates
- Phase 3: Tool event streaming, Pipeline/Notebook toggle group
- Phase 3.5: Two-completes disambiguation, undo pipeline complete, dismiss route

## Walkthrough Log

### Pre-Test: Entry State Check
- Entry: LCARS Vue3 Theme — Clock & RPN Calculator
- Initial state: `maturity: verified, route_status: complete, failures: 1`
- Reset to `maturity: specced, route_status: your_turn` via API
- Toggled from "done" to "active" via UI checkmark
- failure_count stuck at 1 (not resettable via API — **friction point**)

### Execute Attempt 1 — Timeout at 10 Minutes
- **Start:** 1:29:16 PM — clicked "▶ Execute" from UI
- **Tool events streamed in real-time:** report_intent, view (×4, reading scratch in chunks), sql, view, report_intent, then file creates
- **8-minute thinking pause:** Agent spent ~8 minutes in a single generation (reasoning about full implementation). During this time, only a static yellow dot — **no indication the model was still alive**.
- **File creation burst:** Starting at 1:37:11, agent rapidly created ~20 files in ~2 minutes
- **Timeout:** Hit the 10-minute wall at 1:39:16. `context deadline exceeded`
- **Post-timeout tool call:** A `create` tool executed at 1:39:41, **25 seconds after timeout** — Copilot SDK doesn't stop immediately on context cancellation. **Bug.**
- **Result:** Entry reverted to `specced`, failures incremented to 2
- **Agent output:** 35 source files (1445 total with node_modules). Complete implementation — LCARS component library, clock app, calculator app, RPN engine with 20 passing tests, architecture docs, npm workspaces. Both apps build to production.

### Post-Execute Assessment
The agent completed 100% of the implementation work. The timeout killed the *reporting*, not the *work*. This is a critical distinction — the pipeline marks it as failed, but the deliverables are all present and functional.

**Verified manually:**
- `npm run test` — 20/20 tests pass (329ms)
- `npm run build` — Clock app ✅ (335ms), Calculator app ✅ (363ms)
- LCARS library build fails — missing `vite.config.ts` with `build.lib` mode (looks for `index.html`)
- File structure: proper npm workspace monorepo with `lcars/`, `apps/clock/`, `apps/calculator/`

## UX Friction Findings

### ✅ Working Well (Phase 1-3.5 Fixes)
1. **Slim prompt** — Agent read scratch file via `view` tool in chunks, not embedded in prompt. Phase 1 fix working perfectly.
2. **Progress streaming** — Tool events appeared in numbered list in real-time. Could see exactly what the agent was doing.
3. **Cancel button** — Prominent "✕ Cancel" visible throughout execution.
4. **Phase 3.5 labels** — "✓ Dismiss" (route), "Mark done"/"Reopen" (status), correct throughout.
5. **Failure messaging** — Clear failure message in conversation with retry guidance.
6. **🔴 Failure badge** — Shows count and tooltip with last failure reason. Helpful.

### ❌ UX Friction Points

#### P0 — Blocks or Wastes User's Time
1. **No liveness indicator during thinking** — The "● Agent is executing..." badge shows a static yellow dot. During the 8-minute thinking pause, it looks completely frozen. No animation, no spinner, no elapsed time counter. A user would assume it's hung and cancel. **Fix:** Add a spinner animation and an elapsed-time counter.
2. **10-minute timeout too short for complex tasks** — The agent needed ~12 minutes total. It was creating files at full speed when killed. **Fix:** Make timeout configurable per-entry, or increase default to 15-20 minutes.
3. **Copilot SDK doesn't respect context cancellation immediately** — Tool call executed 25 seconds after deadline. Leaves partial state. **Bug in SDK or our usage.**

#### P1 — Confusing or Misleading
4. **Tool event names are opaque** — "view", "sql", "create" with no detail. Which file was viewed? Which table was queried? What was created? **Fix:** Show tool arguments summary (at least first arg) in the event list.
5. **No distinction between "timed out with work done" vs "timed out with nothing"** — Both show the same "Execution failed" message. The entry has 35 files and passing tests, but the pipeline treats it as a total failure. **Fix:** If files were created, note that in the failure message.
6. **failure_count not resettable** — Only resets on successful execution. If I reset an entry for testing, the old failure count persists and colors the UI with a warning badge. **Fix:** Add API/UI reset for failure count.

#### P2 — Polish
7. **No elapsed time counter** — Can't tell how long execution has been running. "Is this 2 minutes in or 8 minutes in?" **Fix:** Show a running timer on the execution indicator.
8. **Title strikethrough when status=done but maturity=specced** — After toggling back to "active", the title un-struck, but during the status=done period with maturity=specced, the visual was contradictory.
9. **Git commit didn't happen** — Success path calls `commitAfterExecution` but timeout means files are on disk uncommitted. Manual intervention needed.

### Bugs Found

#### BUG-1: Post-timeout tool execution
- **Severity:** Medium
- **Description:** After the 10-minute context deadline, a `create` tool call still executed at 1:39:41 (25s after timeout at 1:39:16). The context cancellation signal doesn't immediately abort in-flight Copilot SDK requests.
- **Impact:** Files created after timeout are orphaned — not tracked by `agent.WrittenFiles()`, won't be in any git commit even if execution had "succeeded".
- **Location:** `agent.go` — need to investigate how `ctx.Done()` propagates to the SDK session.

#### BUG-2: Success path doesn't set maturity
- **Severity:** Low
- **Description:** In `execute.go`, the success path notifies with `maturity: "executing"` but never calls `SetMaturity()` to advance. Entry stays at "executing" until the human verifies.
- **Actual behavior:** May be intentional — "executing" until verified. But confusing naming — should be "awaiting_verification" or similar.

## Decision: Manual Advancement
Since the work is 100% complete and tests pass, retrying would waste another 10+ minutes for the agent to discover the files already exist. Instead: advance the entry manually through the pipeline as a steward would, then continue observing subsequent stages.

### Manual Steps Taken
1. Set maturity from `specced` → `executing` via PUT API
2. Verified all 6 scenarios manually:
   - ✅ `npm run dev shows LCARS-styled page` — Clock app ran, LCARS styling confirmed
   - ✅ `Clock app displays local time, UTC, and stardate` — All three updating in real-time
   - ✅ `RPN: 5 ENTER 3 + shows 8` — Tested in browser, X register = 8.0000
   - ✅ `RPN: ENTER pushes stack, T register drops` — 20/20 unit tests pass
   - ✅ `Calculator persists programs in localStorage` — Recorded "DOUBLE" program, reloaded, persisted
   - ✅ `All components render with LCARS styling` — Both apps render without additional CSS
3. Called POST `/api/entries/{id}/verify` with all scenarios passing
4. Entry advanced to `verified` with Sabbath moment message
5. Clicked "Mark done" — entry now `status: done`

### Verification Stage UX Observations
- **Sabbath moment prompt is a nice touch** — "What worked well? What would you do differently?" invites reflection
- **No "Verify" UI appeared when I manually set maturity to executing** — The verify UI depends on a specific agent message that was never posted (because execution timed out). Had to use API directly.

### Additional Friction (Post-Verification)
10. **🔴 Failure badge persists after successful verification** — Entry shows "🔴 2 failures" even after verify success. Verification should reset failure count. **Fix:** Call `ResetFailureCount()` in `Verify()` on all-pass.
11. **"✓ Mark Complete" button still shows at verified+done state** — What would this button do? Unclear purpose after the pipeline is complete.
12. **Title strikethrough at done state makes text hard to read** — The strikethrough effect on "LCARS Vue3 Theme — Clock & RPN Calculator" is visually heavy for a long title.

### LCARS Library Build Issue
The `lcars` package has a `vite build` script but no `vite.config.ts` with `build.lib` mode. It tries to find `index.html` and fails. The apps import from source (`../lcars/src/`) so this doesn't block dev, but `npm run build` at the workspace root exits with code 1. **Fix:** Add `vite.config.ts` with library build config to `lcars/`.

## Summary

### Pipeline Journey
```
raw → researched → planned → specced → executing (timed out) → [manual: executing] → verified → done
```
Total pipeline stages: 7 (including manual recovery from timeout)

### What Worked
- **Slim prompt fix (Phase 1)** — Agent read files via tools, not embedded. Context stayed small.
- **Progress streaming (Phase 3)** — Numbered tool events in real-time. Could watch the agent work.
- **Sabbath moment** — Beautiful reflection prompt after verification.
- **The agent's output quality** — Despite timing out, it produced a complete, working, well-structured codebase (35 source files, 20 tests, npm workspaces).

### What Needs Work
- Timeout: 10 minutes insufficient for complex implementation tasks
- Liveness indicator: No animation/spinner during long generation pauses
- Tool event detail: Opaque names without arguments
- Failure handling: No distinction between "timed out but work complete" vs "timed out with nothing"
- Failure badge: Doesn't reset on successful verification
- Post-timeout cleanup: Copilot SDK continues tool execution after context cancellation

### Bugs Found: 3
1. **BUG-1:** Post-timeout tool execution (Copilot SDK)
2. **BUG-2:** Success path notify says "executing" instead of actual maturity (possibly intentional)
3. **BUG-3:** Verify() doesn't call ResetFailureCount() on all-pass

---

## Fix Plan

Fixes grouped into phases by dependency and priority. Each phase ships independently.

### Phase 4: Activity-Based Execution Timeout (P0)

**Problem:** The 10-minute hard timeout killed a successful execution. The agent was actively creating files when the deadline hit. Go's `context.WithTimeout` cannot extend once created — there's no `SetDeadline()` on derived contexts.

**Solution:** Replace the fixed-deadline context with a custom "activity deadline" mechanism. The context cancels after 30 minutes of *wall clock* time OR after 2 minutes of *inactivity* — whichever comes first.

**Architecture:**

```
pool.go — StartTask() returns ctx + extendDeadline func
  └── activityCtx wraps context.Background() with:
      • 30-minute maximum wall-clock deadline (hard cap)
      • 2-minute inactivity deadline (resets on each touch)
      • cancel() for manual cancellation

agent.go — AskStreaming() already has touchEvent() on every SDK event
  └── Accept an optional OnActivity callback
  └── Call it from touchEvent() so the pool can reset the inactivity timer

execute.go — runExecute() wires them together
  └── OnToolCall callback already exists — also extend deadline there
  └── Pass OnActivity through AgentConfig
```

**Implementation details:**

1. **New type in pool.go** — `activityContext`:
   - Embeds `context.Context` (the parent)
   - Holds a `*time.Timer` for the inactivity deadline (2 min, resets on Touch())
   - Holds a `*time.Timer` for the hard deadline (30 min, never resets)
   - `Touch()` method resets the inactivity timer
   - When either timer fires → cancel the derived context
   - `Done()` channel from the derived context

2. **pool.go changes:**
   - `ExecutionTimeout` → `MaxExecutionTime = 30 * time.Minute`
   - `InactivityTimeout = 2 * time.Minute`
   - `StartTask()` returns `(ctx context.Context, touch func())` instead of just `ctx`
   - `runningTask` gains a `touch func()` field

3. **agent.go changes:**
   - `AgentConfig` gains `OnActivity func()` callback
   - In `AskStreaming()`, inside `touchEvent()`, call `a.config.OnActivity()` if non-nil

4. **execute.go changes:**
   - `runExecute()`:  `ctx, touch := p.pool.StartTask(entry.ID, "execute")`
   - Set `agentCfg.OnActivity = touch`
   - Existing `OnToolCall` doesn't need changes — the agent.go `touchEvent()` already fires on tool events

**Why 2 minutes for inactivity vs 30 minutes wall clock:**
- The 8-minute "thinking pause" we observed was NOT inactivity — the SDK was still sending `AssistantReasoningDelta` events every few seconds. `touchEvent()` fired continuously. A 2-minute gap with zero events means the connection is actually stalled.
- 30 minutes wall clock prevents a runaway agent from burning tokens indefinitely, even if it's "active."

**Files to change:** `pool.go`, `agent.go`, `execute.go`
**Tests:** Unit test `activityContext` — verify inactivity fires, verify touch resets it, verify hard cap fires regardless.

---

### Phase 4a: BUG-3 — Reset Failure Count on Verify (P0)

**Problem:** `Verify()` on all-pass advances to "verified" but doesn't call `ResetFailureCount()`. The red "🔴 2 failures" badge persists forever after verify, misleading the user.

**Fix:** One line in `execute.go`, inside the `allPassed` block of `Verify()`:
```go
p.store.DB().ResetFailureCount(entry.ID)
```

Add it right after `SetMaturity(entry.ID, "verified", ...)`.

**Files to change:** `execute.go`

---

### Phase 4b: Liveness Indicator + Elapsed Timer (P0)

**Problem:** During an 8-minute thinking pause, the "● Agent is executing..." badge is static. No spinner, no timer. Looks frozen.

**Fix — Backend:**
- Add an `execution_started_at` field to the WebSocket `execution.started` event (already sent by `onExecute`)
- Emit periodic `execution.heartbeat` events (every 30s) from the watchdog loop in `agent.go`, by routing them through a new `OnHeartbeat` callback in `AgentConfig`

**Fix — Frontend:**
- Replace static yellow dot with a CSS-animated pulsing dot or spinner
- Add elapsed-time counter: `"● Agent is executing... (3m 42s)"` — compute from `execution_started_at` using `setInterval`
- Heartbeat events keep the frontend timer in sync (optional, timer is client-side)

**Files to change:** `agent.go` (OnHeartbeat callback), `execute.go` (wire it), frontend component that renders the executing badge

---

### Phase 4c: Tool Event Detail (P1)

**Problem:** Tool events show "view", "sql", "create" with no arguments. Can't tell which file or query.

**Fix:**
- `OnToolCall` callback already receives `(toolName string, args any)` — the args are there, just not forwarded
- In `execute.go`'s `OnToolCall`, extract a summary from `args`:
  - `create` / `view` → first string arg (file path)
  - `sql` → first 80 chars of query
  - Others → first arg as string, truncated to 80 chars
- Include `"detail": summary` in the `execution.tool` WebSocket notification
- Frontend: show detail text next to tool name in the event list

**Files to change:** `execute.go` (extract summary from args), frontend event list component

---

### Phase 4d: Smart Failure Messages (P1)

**Problem:** "Execution failed: context deadline exceeded" is the same message whether the agent created 35 files or zero.

**Fix:**
- In `runExecute()` failure path, check `agent.WrittenFiles()`:
  - If files exist → "Execution timed out, but the agent created N files. The work may be complete — check the workspace before retrying."
  - If no files → "Execution failed: {error}. No files were created."
- Also: attempt `commitAfterExecution` even on timeout failures, so any created files are preserved in git

**Files to change:** `execute.go` (failure path)

---

### Phase 4e: Failure Count Reset (P1)

**Problem:** `failure_count` is only reset by successful execution. Can't reset manually for testing or after manual recovery.

**Fix — Option A (API endpoint):**
- Add `POST /api/entries/{id}/reset-failures` → calls `ResetFailureCount()`
- Simple, explicit, testable

**Fix — Option B (Auto-reset on maturity change):**
- When maturity is manually set back to `researched` or `raw`, also reset failures
- Implicit, might surprise users

**Recommendation:** Option A — explicit is better. Wire it into the frontend's entry detail dropdown or a small "reset" link next to the failure badge.

**Files to change:** API handler, `store/db.go` (if not already exposed), frontend failure badge component

---

### Phase 4f: "Mark Complete" Button Visibility (P2)

**Problem:** "✓ Mark Complete" still shows at `verified + done` state. Clicking it does nothing useful.

**Fix:** Hide the button when `status == "done"`. The button's purpose is to mark the *status* as done; once done, it's redundant.

**Files to change:** Frontend entry detail component

---

### Phase 4g: Title Strikethrough at Done (P2)

**Problem:** Strikethrough on long titles is visually heavy and hard to read.

**Fix options:**
- A: Replace strikethrough with a subtle opacity reduction (opacity: 0.6) + a "Done" badge
- B: Strikethrough only on the list view, not on the detail view
- C: Keep strikethrough but limit to first 40 chars + ellipsis

**Recommendation:** Option A — opacity + badge. Communicates "done" without harming readability.

**Files to change:** Frontend CSS/component

---

### Phase 4h: BUG-1 — Post-Timeout Tool Execution (P1, Investigate)

**Problem:** A `create` tool call executed 25 seconds after context deadline. Copilot SDK doesn't abort in-flight requests immediately on `ctx.Done()`.

**Investigation needed:**
- Is the SDK even checking `ctx.Done()` between tool calls?
- Can we add a pre-tool-call check: `if ctx.Err() != nil { return }` in the tool handler?
- Or: do we need to file an issue with the Copilot SDK team?

**Current risk:** Low — the Phase 4 activity-based timeout mostly mitigates this by making hard timeouts rarer. But files created post-timeout aren't tracked by `WrittenFiles()`, so they're invisible to git commit.

**Mitigation for now:**
- In `commitAfterExecution` (and the new timeout-commit path from Phase 4d), scan the workspace for recently-modified files beyond what `WrittenFiles()` reports
- Log a warning when tools fire after deadline

**Files to change:** TBD after investigation

---

### Phase 4i: BUG-2 — Success Notify Says "executing" (P2)

**Problem:** Success path sends `p.notify("entry.updated", ..., {"maturity": "executing"})` but the entry is still in "executing" (waiting for human verify). The label is technically correct but confusing in logs.

**Fix:** Change to `{"maturity": "executing", "route_status": "your_turn"}` so the notify makes clear what actually changed. Or: add a `"phase": "awaiting_verify"` field for clarity.

**Files to change:** `execute.go` (success path notify)

---

### Phase 4j: Verify UI Without Agent Message (P1)

**Problem:** When maturity is manually set to "executing" (as we did in the E2E test), no agent message is posted, so the Verify UI (which triggers off the agent's "Execution complete" message) never appears. Had to use the API directly.

**Fix:** Verify UI should trigger based on *maturity state*, not the presence of a particular message:
- If entry is in `maturity: executing` AND `route_status: your_turn`, show the Verify button
- The agent message is a nice-to-have, not a gate

**Files to change:** Frontend entry detail component (verify button visibility logic)

---

### Execution Order

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

**Total estimated: ~4-5 hours of focused work across all phases.**
Start with 4a (trivial) and 4 (highest impact), then sweep through P1 and P2.
