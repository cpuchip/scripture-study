# Brain Pipeline Fixes — Execute Reliability + Human Gate UI

**Binding problem:** The brain pipeline's execution phase stalls 100% of the time (0/2 success rate) because the execute prompt embeds too much context upfront. The human gate transitions (scenario input, verification, completion) have no UI — they require raw API calls. These two gaps make the pipeline unusable end-to-end.

**Created:** 2026-04-09
**Source:** [UX audit walkthrough](../../.spec/scratch/debug-brain-ux/main.md), [research](../../.spec/scratch/brain-pipeline-fixes/main.md)
**Depends on:** WS1 Phase 4 (shipped), WS4 Phases 1-3 (shipped)
**Status:** ✅ All 3 phases + Phase 3.5 complete (2026-04-10)

---

## Phase 1 Status: ✅ COMPLETE (2026-04-09)

All backend fixes shipped and verified (`go build` + `go vet` clean):

| Fix | File | Status |
|-----|------|--------|
| Slim execute prompt (path not content) | `execute.go` — `buildExecutePrompt()` | ✅ |
| 10-minute execution timeout | `pool.go` — `StartTask()` | ✅ |
| Cancel endpoint wired to HTTP | `server.go` — `POST /api/entries/{id}/cancel-execution` | ✅ |
| Route status = "agent" during execution | `execute.go` — `Execute()` | ✅ |
| Premium cost tracked before Ask() | `execute.go` — `runExecute()` | ✅ |
| Race guard after Ask() returns | `execute.go` — `runExecute()` | ✅ |
| Mark Complete sets maturity + route_status | `server.go` — `handleMarkComplete()` | ✅ |

## Phase 2 Status: ✅ COMPLETE (2026-04-09)

All human gate UI shipped and verified (`vue-tsc`, `vite build`, `go build` clean):

| Fix | File | Status |
|-----|------|--------|
| Toast system replacing all `alert()` calls | `ProjectDetailView.vue` | ✅ |
| Scenario input dialog (planned→specced) | `ProjectDetailView.vue` | ✅ |
| Cancel execution button (board/list/panel/detail) | `ProjectDetailView.vue`, `EntryDetailView.vue` | ✅ |
| Complete button (verified→complete) | `ProjectDetailView.vue`, `EntryDetailView.vue` | ✅ |
| Done column shows `complete` entries with badge | `ProjectDetailView.vue` | ✅ |
| Maturity badge in entry detail header | `EntryDetailView.vue` | ✅ |
| Pipeline gate sections (scenario/execute/verify/advance/complete) | `EntryDetailView.vue` | ✅ |
| `cancelExecution()` API method | `api.ts` | ✅ |

## Phase 3 Status: ✅ COMPLETE (2026-04-09)

All polish items shipped and verified:

| Fix | File | Status |
|-----|------|--------|
| Replace `window.alert()` with toast | (Done in Phase 2) | ✅ |
| Execution progress indicator (stream tool events) | `agent.go`, `execute.go`, `EntryDetailView.vue` | ✅ |
| Pipeline/Notebook toggle (checkbox → button group) | `EntryDetailView.vue` | ✅ |
| Hide premature Verify button | (Done in Phase 2) | ✅ |
| Maturity badge on Done column | (Done in Phase 2) | ✅ |

## Phase 3.5 Status: ✅ COMPLETE (2026-04-10)

Discovered during testing: two different "complete" concepts used the same label and one was irreversible.

| Fix | File | Status |
|-----|------|--------|
| Disambiguate circle checkbox: "Mark done"/"Reopen" | `EntryDetailView.vue` | ✅ |
| Conversation `✓ Complete` → `✓ Dismiss` (calls dismissRoute) | `EntryDetailView.vue` | ✅ |
| Route status "complete" badge → "✓ Routed" (not "✓ Complete") | `ProjectDetailView.vue` | ✅ |
| ↩ Undo pipeline complete (reverts to verified) | `EntryDetailView.vue`, `ProjectDetailView.vue` | ✅ |
| Fixed accidental complete on Build Physical Display Dashboard | API call (dismiss-route) | ✅ |

**Root cause:** `route_status: complete` (agent routing finished) and `maturity: complete` (pipeline finished) both showed as "✓ Complete". The conversation section had a `✓ Complete` button that called `handleMarkComplete` (irreversible pipeline complete) when users expected it to acknowledge the route status badge.

---

## 1. Problem Statement

### What's Broken

The UX audit walked a real entry (LCARS Vue3 Theme) through the full pipeline: raw → researched → planned → specced → executing → verified → complete.

**Research and plan work well.** Haiku researched in 4.5 min, Opus planned in 2.5 min, auto-continue was seamless.

**Execute stalls every time.** Two attempts, same pattern: agent reads the scratch file, then the Copilot SDK stops sending events. No timeout, no cancel, no recovery. Entry stuck at "executing" forever.

**Human gates have no UI.** Scenario input (planned→specced), scenario verification (executing→verified), and completion (verified→complete) all required raw API calls. The "Your Turn" badge signals the gate but provides no mechanism to pass through it.

### Root Cause

The execute prompt embeds the full scratch content (up to 10K chars) directly in the prompt text. Combined with the system message (~4K), project context, scenarios, and copilot-instructions.md (~12K auto-loaded by SDK), the initial context is ~25-30K chars before the agent makes a single tool call. The agent then tries to *re-read the same scratch file via tools*, doubling the context load.

Compare: the research agent gives the agent a *path* and says "write to this file." It doesn't embed content. That's why research works and execute doesn't.

Squad takes the same approach — their coordinator sends a focused `InitialPrompt` (priority + task + context reference) and lets agents read what they need.

### Success Criteria

1. Execute completes successfully on the LCARS entry (the test case that stalled 2/2)
2. All human gates (scenario input, verification, completion) have UI in the entry detail page
3. Execute has a timeout and cancel mechanism
4. Agent activity is visible to the user during execution (not just disabled buttons)
5. Errors use toast notifications, not `window.alert()`

---

## 2. The Fix — Three Phases

### Phase 1: Make Execute Work (Backend)

**The critical fix: slim the execute prompt.**

Change `buildExecutePrompt()` to NOT embed scratch content. Instead, give the agent the path:

```go
// BEFORE (current — embeds up to 10K chars)
if scratchContent != "" {
    fmt.Fprintf(&sb, "## Research & Plan (from scratch file)\n\n")
    fmt.Fprintf(&sb, "```markdown\n%s\n```\n\n", scratchContent)
}

// AFTER (give path, let agent read what it needs)
if entry.ScratchPath != "" {
    fmt.Fprintf(&sb, "## Research & Plan\n\n")
    fmt.Fprintf(&sb, "Read the research and plan from: `%s`\n", entry.ScratchPath)
    fmt.Fprintf(&sb, "Start by reading this file to understand the plan before implementing.\n\n")
}
```

This alone should drop the prompt from ~15K+ to ~2-3K chars.

**Additional Phase 1 fixes:**

1. **Add execution timeout.** The pool's `StartTask()` currently creates `context.WithCancel`. Change to `context.WithTimeout` (10 minutes):

```go
func (p *AgentPool) StartTask(entryID, agentName string) context.Context {
    // ...cancel existing...
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    // ...
}
```

2. **Wire cancel to HTTP.** Add `POST /api/entries/{id}/cancel` endpoint that calls `pool.CancelTask(entryID)` and resets maturity to "specced":

```go
func (s *Server) handleCancelExecution(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    s.pool.CancelTask(id)
    s.store.DB().SetMaturity(id, "specced", "Execution cancelled by user")
    s.store.DB().AddSessionMessage(id, "system", "Execution cancelled.")
    // ...notify websocket...
}
```

3. **Set route_status="agent" during execution.** In `Execute()`, after setting maturity to "executing":

```go
p.store.DB().UpdateRouteStatus(entry.ID, "agent")
```

4. **Track premium cost immediately.** Move `IncrementPremiumRequests` to before `agent.Ask()`, not after. The cost is spent when the request starts, not when it finishes.

5. **Protect against goroutine races.** The pool already cancels existing tasks in `StartTask()` — this is sufficient. But add a nil-check in `runExecute` after `agent.Ask()` returns to verify the entry hasn't been manually reset:

```go
// After Ask() returns, verify entry is still in executing state
current, err := p.store.DB().GetEntry(entry.ID)
if err != nil || current.Maturity != "executing" {
    log.Printf("Entry %s maturity changed during execution (%s), aborting post-processing", entry.ID, current.Maturity)
    return
}
```

**Files changed:**
- `internal/pipeline/execute.go` — slim prompt, route_status, cost tracking, race guard
- `internal/ai/pool.go` — timeout on StartTask
- `internal/web/server.go` — cancel endpoint

### Phase 2: Human Gate UI (Frontend)

**Status: ✅ COMPLETE (2026-04-09)**

**Scenario Input** (planned → specced):

Add a scenario textarea to the entry detail page when `maturity === "planned"`:

```vue
<!-- EntryDetailView.vue -->
<div v-if="entry.maturity === 'planned'" class="scenario-input">
  <h3>Scenarios (Acceptance Criteria)</h3>
  <p class="hint">Define how you'll verify this is done. One scenario per line.</p>
  <textarea v-model="scenarioText" rows="6" placeholder="- User can see the clock display&#10;- Calculator handles basic operations&#10;- Theme matches LCARS color palette"></textarea>
  <button @click="advanceWithScenarios" :disabled="!scenarioText.trim()">
    Advance to Specced
  </button>
</div>
```

Wire to existing `POST /api/pipeline/advance` with `{ id, scenarios: scenarioText }`.

**Scenario Verification** (executing → verified):

When `maturity === "executing"` and execution is done (agent posted verify message), show scenario checkboxes:

```vue
<div v-if="showVerification" class="scenario-verify">
  <h3>Verify Scenarios</h3>
  <div v-for="(s, i) in parsedScenarios" :key="i" class="scenario-row">
    <label>
      <input type="checkbox" v-model="s.passed" />
      {{ s.text }}
    </label>
    <input v-if="!s.passed" v-model="s.notes" placeholder="What failed?" />
  </div>
  <button @click="submitVerification">Submit Verification</button>
</div>
```

Wire to existing `POST /api/entries/{id}/verify`.

**Mark Complete** fix:

In `handleMarkComplete()`, also set maturity to "complete":

```go
// server.go handleMarkComplete
s.store.DB().SetMaturity(id, "complete", "")
s.store.DB().UpdateRouteStatus(id, "complete")
```

**Cancel button** during execution:

```vue
<div v-if="entry.maturity === 'executing'" class="execution-status">
  <span class="spinner" /> Executing...
  <button @click="cancelExecution" class="btn-cancel">Cancel</button>
</div>
```

**Files changed:**
- `frontend/src/views/EntryDetailView.vue` — scenario input, verification UI, cancel button, execution status
- `frontend/src/api.ts` — `cancelExecution(id)`, update `advancePipeline` to accept scenarios
- `internal/web/server.go` — handleMarkComplete fix

### Phase 3: Error Handling & Polish (Frontend + Backend)

1. **Replace `window.alert()` with toast.** Add a simple toast composable and replace all `alert()` calls in pipeline error handling:

```typescript
// composables/useToast.ts
const toasts = ref<Toast[]>([])
function show(message: string, type: 'error' | 'success' | 'info') { ... }
```

2. **Progress indicator.** Stream execution events to WebSocket so the UI can show what the agent is doing:

```go
// In runExecute, after each tool call logged by the audit hook:
p.notify("execution.tool", entry.ID, map[string]string{"tool": toolName})
```

Frontend shows a running log: "Reading architecture.md... Searching workspace... Creating files..."

3. **Fix Pipeline/Notebook toggle.** Change from checkbox to two radio buttons or relabel to show current state, not action:

```vue
<!-- Current: misleading checkbox -->
<label><input type="checkbox" /> 🔄 Pipeline</label>

<!-- Fix: radio group showing current state -->
<div class="mode-toggle">
  <label :class="{ active: !entry.notebook }">🔄 Pipeline</label>
  <label :class="{ active: entry.notebook }">📓 Notebook</label>
</div>
```

4. **Hide premature Verify button.** On the board, only show Verify when `maturity === "executing"` AND `route_status === "your_turn"` (meaning agent finished):

```vue
<button v-if="entry.maturity === 'executing' && entry.route_status === 'your_turn'">
  ✓ Verify
</button>
```

5. **Show maturity badge on Done column.** The board's Done column should show the maturity state (verified, complete).

**Files changed:**
- `frontend/src/composables/useToast.ts` — new
- `frontend/src/components/ToastContainer.vue` — new
- `frontend/src/views/EntryDetailView.vue` — toggle fix, toast integration
- `frontend/src/views/ProjectDetailView.vue` — Verify button guard, Done column badge
- `internal/web/server.go` — execution event notifications

---

## 3. Constraints

- **No new dependencies.** Toast is CSS + a reactive array. No npm packages.
- **Each phase ships standalone value.** Phase 1 alone fixes the stall. Phase 2 makes it usable. Phase 3 polishes.
- **Don't break existing pipeline.** Research and plan are working — don't touch those code paths.
- **Respect the existing event system.** Brain already has WebSocket notifications (`p.notify`). Use the same pattern.

---

## 4. Verification

### Phase 1 Test
1. Reset LCARS entry to "specced" via API
2. Hit `POST /api/entries/{id}/execute`
3. Agent should complete within 10 minutes, produce files, post verify message
4. If it stalls, timeout fires at 10 min and resets to specced with error message

### Phase 2 Test
1. Create a new entry, advance to "planned" via pipeline
2. Scenario textarea should appear on entry detail
3. Type scenarios, click Advance → entry moves to "specced"
4. Execute → when done, scenario checkboxes appear
5. Check all pass → "verified" → Mark Complete → "complete"

### Phase 3 Test
1. Trigger a pipeline error → toast appears (not alert)
2. During execution → progress messages stream to UI
3. Toggle Pipeline/Notebook → behavior matches label

---

## 5. Costs & Risks

| Cost | Estimate |
|------|----------|
| Phase 1 (backend fixes) | 1 session |
| Phase 2 (human gate UI) | 1-2 sessions |
| Phase 3 (polish) | 1 session |
| Premium cost to re-test | ~1.33 (research skip + execute 1.0 + plan 0.33 if needed) |

**Risks:**
- Prompt slimming might not fully fix the stall if the issue is deeper in the SDK. Mitigation: the timeout + cancel mechanisms are the safety net.
- Phase 2 UI changes touch the detail view which is already complex. Mitigation: each component (scenario input, verification, cancel) is self-contained.

---

## 6. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Pipeline is unusable — 0% execute success, 0% human gate UI |
| Covenant | Rules? | Don't break working research/plan. Ship each phase standalone. |
| Stewardship | Who owns? | brain repo (scripts/brain/) — dev agent executes |
| Spiritual Creation | Spec precise enough? | Yes — code changes specified per file with before/after |
| Line upon Line | Phasing? | Phase 1 (backend) → 2 (frontend) → 3 (polish). Each stands alone. |
| Physical Creation | Who executes? | dev agent, one phase per session |
| Review | How to verify? | Test cases specified per phase |
| Atonement | If wrong? | Timeout/cancel are safety nets. UI changes are cosmetic. Git revert if needed. |
| Sabbath | When pause? | After each phase — test before moving on |
| Consecration | Who benefits? | Michael directly. Any future Brain user. |
| Zion | Whole system? | Unblocks the pipeline for all projects, not just LCARS |

---

## 7. Recommendation

**Build — Phase 1 immediately.** The prompt slimming is a one-line change that likely fixes the #1 blocker. Timeout + cancel are straightforward backend additions. This can ship in one session.

Phase 2 (human gate UI) should follow in the next session — it makes the pipeline actually usable without API calls.

Phase 3 is polish and can wait.
