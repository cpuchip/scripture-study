# Brain Pipeline Fixes — Research & Scratch

**Date:** 2026-04-09
**Source:** UX audit walkthrough (.spec/scratch/debug-brain-ux/main.md)
**Binding problem:** The brain pipeline's execution phase is unreliable (0/2 success rate) and the human gate transitions (scenario input, verification, completion) lack UI support entirely.

## Inventory: What Exists

### Brain's Architecture (Go + Copilot SDK)
- **Agent system:** `internal/ai/agent.go` — wraps `@github/copilot-sdk/go`, sessions via `copilot.Client`
- **Pool management:** `internal/ai/pool.go` — `AgentPool` tracks named agents + running tasks per entry
- **Pool already has cancel:** `StartTask()` returns cancellable context, `CancelTask(entryID)` exists, `FinishTask()` cleans up
- **Execute flow:** `internal/pipeline/execute.go` — validates specced+scenarios, sets maturity="executing", fires `runExecute()` goroutine
- **runExecute():** Creates agent with Sonnet 4.6, PremiumRequestCost=1.0, sends single prompt. On success: git commit written files, store output, post verify message. On failure: reset to specced, increment failure count.
- **Scratch content:** Already truncated to 10K chars in `loadScratchContent()` — but the system message + prompt + project context pushes total well beyond comfortable limits

### Squad's Architecture (TypeScript + Copilot SDK)
- **Same SDK:** Uses `@github/copilot-sdk` (JS equivalent)
- **Session lifecycle:** Dedicated `CopilotSessionAdapter` wrapping SDK sessions, with `sendAndWait()`, `abort()`, event normalization
- **Health monitoring:** Dedicated `HealthMonitor` with ping-based checks, timeout detection
- **Streaming pipeline:** Dedicated `StreamingPipeline` — tracks time-to-first-token, usage, delta handlers, per-session state
- **Fan-out:** `spawnParallel()` spawns multiple agents via `Promise.allSettled` — error isolation per agent
- **Session pool:** Tracks active/idle/error sessions with capacity limits
- **Event bus:** Cross-session observability — `session.created`, `session.error`, `session.destroyed`
- **Cost tracking:** Real-time `estimateCost()` per session, wired to event bus
- **Scheduler:** Provider-based schedule system with retry config and execution state tracking

### Key Differences

| Aspect | Brain | Squad |
|--------|-------|-------|
| SDK language | Go | TypeScript |
| Session lifecycle | Create → Ask → wait for idle | Create → sendMessage → events |
| Health checks | Watchdog (30s log warning) | HealthMonitor (ping + timeout) |
| Cancel | `pool.CancelTask()` exists but NOT wired to HTTP | `session.abort()` |
| Timeout | None | Config-based per provider |
| Cost tracking | After completion only | Real-time via event bus |
| Error isolation | Goroutine crashes → entry stuck | `Promise.allSettled` per agent |
| Progress | Server logs only | Event bus → streaming to UI |
| Streaming | `io.Discard` for pipeline agents | Full delta/reasoning pipeline |

### What Brain Has That Squad Doesn't
- **Governance documents** — covenant-based agent constraints
- **Scratch files** — persistent research/plan artifacts in workspace
- **Selective git commits** — agent file writes tracked and committed
- **Human gate model** — explicit maturity stages with human approval
- **Premium budget tracking** — per-entry cost accounting

## The Stall Problem

### Reproduction
- 2/2 execute attempts stalled at the same point
- Pattern: Agent reads full scratch file (~350 lines) → SDK stops sending events
- Watchdog fires but only logs — no recovery action
- First attempt: 94s initial response, then stall after 4 min of reading
- Second attempt: 6s initial response (fresh session), stall after reading scratch file

### Root Cause Hypothesis
The execute prompt is too large. It includes:
1. System message: base instructions + governance doc + execute rules (~3-4K)
2. Prompt: entry metadata + project context + scenarios + scratch content (truncated at 10K) + instructions
3. copilot-instructions.md loaded by SDK automatically (~12K)
4. The agent then reads MORE files via tool calls, pushing context further

Total context likely exceeds what Sonnet can comfortably process in the Copilot SDK's streaming pipeline. The SDK's internal handling of large contexts with tool-use may have edge cases that cause the event stream to stall.

### Evidence
- Research agent (Haiku) hit 112K tokens — WARNING at 100K threshold
- Plan agent (Opus) hit 157K tokens — WARNING at 150K threshold
- Execute agent uses Sonnet with 200K threshold — likely exceeded before first tool call given pre-loaded context

## What Actually Needs Fixing (Priority Order)

### Tier 1: Make Execute Work
1. **Execution timeout** — `context.WithTimeout` on the goroutine
2. **Cancel endpoint** — wire `pool.CancelTask()` to HTTP API
3. **Route status during execution** — set to "agent" not "your_turn"
4. **Track premium cost on failure** — move IncrementPremiumRequests before Ask()
5. **Reduce prompt size** — don't embed full scratch content; tell agent where to read it

### Tier 2: Human Gate UI
6. **Scenario input** — textarea on entry detail when maturity=planned
7. **Scenario verification UI** — checkboxes per scenario when maturity=executing
8. **Mark Complete behavior** — set maturity to "complete" not just route_status

### Tier 3: Feedback & Polish
9. **Replace window.alert with toasts**
10. **Progress indicators** — surface agent activity to UI
11. **Fix Pipeline/Notebook toggle semantics**
12. **Hide premature Verify button**

## Key Insight: Prompt Size Is The Fix

The #1 thing to fix isn't timeout/cancel (those are safety nets). It's **why the agent stalls in the first place**.

Brain embeds the full scratch file (up to 10K chars) INTO the prompt, plus the system message includes base instructions. The agent then tries to read the same file via tools. Double-loading.

Squad doesn't pre-load content — it gives agents tasks and lets them read what they need. The coordinator sends a focused InitialPrompt: priority + task + context. Short and targeted.

**The fix:** Don't embed scratch content in the execute prompt. Give the agent the *path* to the scratch file and let it read what it needs. This alone could drop the prompt from ~15K+ chars to ~2K chars and likely fix the stall.
