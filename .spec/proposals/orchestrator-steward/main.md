# Orchestrator Steward — Building the Steward Loop into Brain

**Binding problem:** The brain pipeline fires single agent calls per stage and either succeeds or fails. When execution fails — timeout, error, partial work — the human must manually diagnose, fix, and retry. Michael has demonstrated that the real power of AI delegation is the *steward loop*: watching, diagnosing, fixing, restarting, repeating until done. Can we build this loop into brain itself, so that the system shepherds entries to completion rather than abandoning them at first failure?

**Created:** 2026-04-10
**Source:** [E2E walkthrough findings](../../.spec/scratch/debug-pipeline-e2e/main.md), [orchestrator research](../../.spec/scratch/orchestrator-steward/main.md)
**Related:** [brain-pipeline-fixes-phase4.md](../brain-pipeline-fixes-phase4.md), [brain-pipeline-evolution.md](../brain-pipeline-evolution.md)
**Status:** Proposed

---

## The Scriptural Frame

This proposal is designed from gospel stewardship patterns, not just engineering best practices. The full scriptural analysis is in the [research scratch file](../../.spec/scratch/orchestrator-steward/main.md); this section captures the architectural principles that emerge from scripture.

### The Good Shepherd Pattern (John 10, Ezekiel 34)

The shepherd *knows the sheep by name* (John 10:3), *goes before them* (v4), and *stays when danger comes* (v11). The hireling — present when convenient, absent when it costs — is the anti-pattern (v12-13). When appointed shepherds fail, the Lord Himself steps in: "I, even I, will both search my sheep, and seek them out" ([Ezekiel 34:11](../../../gospel-library/eng/scriptures/ot/ezek/34.md)).

**Architectural principle:** The orchestrator is invested in outcomes, not just dispatching tasks. It tracks each entry by identity, knows its state, and stays engaged through difficulty. When a stage fails, it doesn't shrug — it diagnoses and acts.

### The Watchman on the Tower (D&C 101:43-62)

The servants planted olive trees and built a hedge — but questioned the need for the tower: "What need hath my lord of this tower?" (v47). They became slothful and didn't finish monitoring. The enemy came by night. The watchman "would have seen the enemy while he was yet afar off" (v54).

**Architectural principle:** Observability is not optional. Monitoring costs resources, and in peacetime it looks unnecessary. But skipping it is exactly what makes the system vulnerable. The steward builds the tower *especially* when things seem fine.

### The Allegory of the Olive Tree (Jacob 5)

The Lord of the vineyard personally inspects ("Come, let us go down into the vineyard" v15), weeps when the trees fail ("What could I have done more?" v41, 47), counsels patience ("Spare it a little longer" v50), prunes proportionally ("ye shall not clear away the bad thereof all at once, lest the roots thereof should be too strong for the graft" v65), and ultimately labors *alongside* the servants (v72).

**Architectural principles:**
- **Not delegation-and-forget but delegation-and-partnership.** The orchestrator doesn't just dispatch — it provides context, checks progress, and responds to needs.
- **Patience before destruction.** Don't kill tasks at first sign of trouble. Retry with care.
- **Proportional intervention.** Escalate gradually. Don't restart everything when targeted repair would suffice.
- **Exhaustive effort before giving up.** "What could I have done more?" — try multiple strategies before the entry goes to dead letter.

### Stewardship Accountability (D&C 104:11-18, D&C 72:3, Luke 12:42-48)

"Organize yourselves and appoint every man his stewardship... that every man may give an account" ([D&C 104:11-12](../../../gospel-library/eng/scriptures/dc-testament/dc/104.md)). "Unto whomsoever much is given, of him shall be much required" ([Luke 12:48](../../../gospel-library/eng/scriptures/nt/luke/12.md)). "To every man according to his several ability" ([Matthew 25:15](../../../gospel-library/eng/scriptures/nt/matt/25.md)).

**Architectural principles:**
- **Clear boundaries.** Each pipeline stage has a defined steward (model + prompt + scope).
- **Proportional assignment.** Cheap models get simple tasks. Expensive models get complex ones. Don't send everything to Opus.
- **Structured reporting.** Every execution produces an account — what was done, spent, achieved.
- **Agents unto themselves.** Agents have genuine autonomy *within* their stewardship boundaries (D&C 104:17).

### The Watchman Must Warn (Ezekiel 33:1-9)

"If the watchman see the sword come, and blow not the trumpet, and the people be not warned... his blood will I require at the watchman's hand" ([Ezekiel 33:6](../../../gospel-library/eng/scriptures/ot/ezek/33.md)).

**Architectural principle:** Silent failures are the watchman's sin. The orchestrator MUST surface problems — to the human, to the log, to the UI. Swallowing errors is not robustness; it's negligence.

### Covenant Fidelity (D&C 82:10)

"I, the Lord, am bound when ye do what I say; but when ye do not what I say, ye have no promise."

**Architectural principle:** System reliability comes from consistent, promise-keeping behavior. When the orchestrator says it will retry, it retries. When it says it will escalate, it escalates. Reliability is faithfulness.

---

## Success Criteria

1. An entry that fails during execution (timeout, model error, partial work) is automatically retried with diagnostic context — without human intervention
2. The retry strategy is smart: it includes the failure reason in the retry prompt, escalates to a higher-capability model after repeated failures, and ultimately quarantines the entry for human review
3. The steward loop is observable: every retry, escalation, and quarantine decision is logged and visible in the UI
4. The human retains override authority at every point (cancel, skip, force-advance, force-fail)
5. The steward loop works across process restarts (state is in SQLite, not in goroutine memory)
6. Pipeline cost stays reasonable: the steward doesn't burn through premium requests on hopeless retries

---

## Architectural Overview

```
                 ┌─────────────────────────────────┐
                 │          Human (Steward)          │
                 │    your_turn gates, manual ctrl   │
                 └──────────┬────────────────────────┘
                            │ override / escalation
                            ▼
┌─────────────────────────────────────────────────────────┐
│                   Steward Loop (new)                     │
│                                                          │
│  ┌──────────┐   ┌──────────┐   ┌───────────┐           │
│  │  Watch    │──▶│ Diagnose │──▶│   Act     │           │
│  │ (tower)  │   │ (assess) │   │ (shepherd)│           │
│  └──────────┘   └──────────┘   └───────────┘           │
│       ▲                              │                   │
│       │         ┌──────────┐         │                   │
│       └─────────│ Account  │◀────────┘                   │
│                 │ (report) │                             │
│                 └──────────┘                             │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│              Existing Pipeline (unchanged)                │
│  raw → researched → planned → specced → executing →     │
│  → verified → done                                       │
└─────────────────────────────────────────────────────────┘
```

The steward loop wraps the existing pipeline, not replaces it. Each cycle:

1. **Watch** — Monitor running tasks for completion, timeout, or failure. Check stale entries. (The watchman on the tower.)
2. **Diagnose** — When something goes wrong, classify the failure: transient? model-capability? prompt issue? missing context? resource limit? (The shepherd seeking the lost.)
3. **Act** — Based on diagnosis: retry with same model, retry with feedback/context, escalate to higher model, or quarantine for human review. (The lord pruning proportionally.)
4. **Account** — Record what happened: action taken, cost incurred, outcome. (The steward rendering account.)

---

## Constraints and Boundaries

**In scope:**
- Automatic retry with diagnostic context after pipeline stage failures
- Model escalation (Haiku → Sonnet → Opus) after repeated failure at the same tier
- Failure quarantine (dead letter) with notification after exhausting retry strategies
- State persistence in SQLite for crash recovery
- UI visibility: retry history, escalation events, quarantine queue
- Human override at every decision point

**Not in scope (yet):**
- Multi-agent fan-out (Squad-style parallel execution) — that's a separate, later proposal
- Stage reordering or dynamic routing — the pipeline sequence stays fixed
- Auto-execution of specced entries — the your_turn gate at specced→executing stays human-controlled
- External service orchestration — this is about the internal pipeline only

**Conventions:**
- Go, standard brain patterns (SQLite WAL, WebSocket hub, config package)
- All retry/escalation logic in a new `steward` package under `internal/`
- Orchestrator decisions logged as session messages (visible in entry timeline)
- Configuration in `config.go` with sensible defaults, no env var sprawl

---

## Prior Art

### What exists in brain:
- **`maybeAutoContinue()`** — already auto-advances from researched→planned when `auto_continue` is set. This is the embryo of the steward loop.
- **`recordFailure()`** — increments failure count, posts session message, but doesn't retry or diagnose.
- **`review.go` nudge bot** — periodic scanner that finds stale entries and nudges. Already has scheduled execution, wake hours, pause/resume, state tracking. The steward loop is architecturally similar.
- **`touchEvent()` / inactivity monitoring** — already watches for agent activity. Phase 4 proposal extends this.
- **`your_turn` gates** — human-in-the-loop checkpoints. The steward never bypasses these.

### What exists externally:
- **Squad coordinator** — Route analysis → fan-out → collect. Session pool with lifecycle tracking. Good for conversational routing, but brain is sequential pipeline. Key borrowable: spawn-as-work-unit, failure isolation per spawn, cost tracking.
- **Industry patterns** — Exponential backoff, circuit breakers (CLOSED/OPEN/HALF-OPEN), multi-provider fallback, checkpointing, dead letter queues. Layered composition: retry inside circuit breaker inside fallback inside observability.

---

## Proposed Approach

### The Steward Type

```go
// internal/steward/steward.go

type Steward struct {
    store    *store.Store
    pipeline *pipeline.Pipeline
    pool     *ai.AgentPool
    config   *StewardConfig
    hub      *ws.Hub        // WebSocket for notifications
    state    *stewardState  // mutex-protected mutable state
}

type StewardConfig struct {
    MaxRetries       int           // per stage, default 3
    BackoffBase      time.Duration // base delay between retries, default 30s
    BackoffMax       time.Duration // max delay, default 5m
    EscalationChain  []string      // model escalation order: [haiku, sonnet, opus]
    QuarantineAfter  int           // total attempts before dead-letter, default 5
    WatchInterval    time.Duration // how often to check running tasks, default 30s
    CircuitBreaker   CircuitBreakerConfig
}
```

### Watch Phase — The Tower

The steward runs a background loop (like the nudge bot) that monitors:

1. **Running tasks** — checks `pool.RunningTasks()` for completion or inactivity timeout
2. **Recently failed entries** — scans for entries with `failure_count > 0` and `failure_count < quarantine_threshold`
3. **Stale entries** — entries stuck at a maturity level longer than expected (this overlaps with the existing nudge bot and may absorb it)

When the watcher sees a failure or timeout:
1. Read the entry's failure history from the DB
2. Read the last session message (the error)
3. Pass to the Diagnose phase

### Diagnose Phase — The Shepherd Seeking

The diagnosis classifies the failure type. This can be simple pattern matching initially, with optional LLM-assisted diagnosis for ambiguous failures:

```go
type FailureType string

const (
    FailureTransient    FailureType = "transient"    // network, rate limit, API error
    FailureTimeout      FailureType = "timeout"      // inactivity timeout
    FailureModelLimit   FailureType = "model_limit"  // model can't handle the task
    FailurePromptIssue  FailureType = "prompt_issue"  // bad instructions, missing context
    FailureToolError    FailureType = "tool_error"    // MCP tool failure
    FailureUnknown      FailureType = "unknown"       // needs human review
)
```

Diagnosis rules:
- Timeout + partial work → `timeout` (retry with context about what was done)
- API 429/500/503 → `transient` (retry with backoff)
- Same failure message 2+ times → `model_limit` (escalate model)
- Tool call error → `tool_error` (retry with tool-specific guidance)
- Unknown → `unknown` (quarantine)

### Act Phase — Proportional Pruning

Based on diagnosis, choose action:

| Diagnosis | Attempt 1 | Attempt 2 | Attempt 3 | After max |
|-----------|-----------|-----------|-----------|-----------|
| Transient | Retry with backoff | Retry with longer backoff | Retry once more | Quarantine |
| Timeout | Retry with "continue from..." context | Retry with extended timeout hint | Escalate model | Quarantine |
| Model limit | Escalate to next model tier | Escalate again | — | Quarantine |
| Tool error | Retry with tool error context | Escalate model | — | Quarantine |
| Unknown | Quarantine immediately | — | — | — |

**Model escalation chain:** Haiku → Sonnet → Opus → Human

The "Human" endpoint in the chain means: set the entry to `your_turn` and notify. This maps to [Ezekiel 34:11](../../../gospel-library/eng/scriptures/ot/ezek/34.md) — when the appointed shepherds fail, the Lord steps in.

**Retry context:** Each retry includes what the previous attempt did and why it failed. Not a blind retry — an informed retry.

```go
type RetryContext struct {
    PreviousAttempt  string    // summary of what the agent produced
    FailureReason    string    // why it failed
    FailureType      FailureType
    AttemptNumber    int
    ModelUsed        string    // which model failed
    PartialWork      string    // any partial output worth preserving
    GuidanceForRetry string    // specific instructions for the retrying agent
}
```

### Account Phase — Rendering Account

Every steward action is recorded:

1. **Session message** — visible in the entry timeline ("Steward: retrying research with sonnet after haiku model limit")
2. **Steward log** — new DB table tracking all orchestrator decisions
3. **WebSocket notification** — real-time updates to the UI
4. **Cost tracking** — premium request cost of each retry recorded

```go
type StewardAction struct {
    ID           string
    EntryID      string
    Timestamp    time.Time
    ActionType   string // "retry", "escalate", "quarantine", "resume"
    Diagnosis    FailureType
    ModelUsed    string
    Cost         float64
    Outcome      string // "success", "failed", "in_progress"
    Notes        string
}
```

### Circuit Breaker (D&C 101:47-54)

A per-stage circuit breaker prevents wasting tokens on systematically broken stages:

```
CLOSED (normal) → after N failures → OPEN (stop trying) → after cooldown → HALF-OPEN (single probe)
```

If research fails 5 times in a row across different entries, the circuit breaker opens for that stage. This prevents the Ezekiel scenario — the watchman falls asleep and everything gets overrun because the system keeps retrying a fundamentally broken stage.

### Quarantine Queue (Dead Letter)

Entries that exhaust all retry strategies enter quarantine:
- Maturity stays at current stage
- `quarantined` flag set
- Notification sent to human
- Entry appears in a dedicated "Needs Attention" section in the UI
- Human can: provide feedback and retry, force-advance, reject, or unquarantine

This IS the [Ezekiel 34:11](../../../gospel-library/eng/scriptures/ot/ezek/34.md) moment: "I, even I, will both search my sheep, and seek them out." The system did what it could. Now the human steward intervenes personally.

---

## Phased Delivery

### Phase 1: Failure Retry with Context (2-3 sessions)

*The minimum viable steward.*

- New `internal/steward/` package with `Steward` type
- Watch loop: monitor `recordFailure()` events (subscribe to pipeline failure notifications)
- Simple diagnosis: transient vs timeout vs unknown
- Retry with context: include failure reason in retry prompt
- Max 2 retries before quarantine
- Session messages for every action
- No model escalation yet — same model, just with better context

**Why this phase stands alone:** Even without escalation, retrying with failure context solves the most common case: execution that timed out or hit a transient error. The human no longer has to manually re-trigger.

### Phase 2: Model Escalation (1-2 sessions)

- Extend diagnosis: detect model_limit failures
- Implement escalation chain: Haiku → Sonnet → Opus → Human
- Track model per retry in the steward log
- Cost guardrails: configurable max premium requests per entry per day

### Phase 3: Circuit Breaker (1 session)

- Per-stage circuit breaker: CLOSED → OPEN → HALF-OPEN
- Dashboard indicator when a stage's circuit is open
- Auto-recovery probe when cooldown expires

### Phase 4: Quarantine Queue & UI (1-2 sessions)

- `quarantined` flag on entries
- "Needs Attention" section in dashboard
- Human actions: feedback-and-retry, force-advance, reject, unquarantine
- Quarantine history visible in entry timeline

### Phase 5: Nudge Bot Integration (1 session)

- Absorb the existing nudge bot's stale-entry detection into the steward
- Single steward loop replaces two separate background goroutines
- Preserve nudge bot's existing UI (status, pause/resume, wake hours)

---

## Verification Criteria

| Phase | Test | Pass Condition |
|-------|------|----------------|
| 1 | Trigger a research failure, observe steward retry | Entry retried within 30s with failure context in prompt |
| 1 | Trigger 3 failures, observe quarantine | Entry quarantined, notification sent, no more retries |
| 2 | Fail with Haiku, observe escalation to Sonnet | Retry uses Sonnet model, session message reflects escalation |
| 2 | Exceed daily cost limit per entry | Steward stops retrying, quarantines with cost-limit reason |
| 3 | Fail same stage 5 times across entries | Circuit breaker opens, no more attempts until cooldown |
| 3 | Observe circuit recovery | After cooldown, single probe attempt, HALF-OPEN → CLOSED on success |
| 4 | View quarantine queue in UI | Quarantined entries visible with action buttons |
| 5 | Stale entry detected by steward | Same behavior as existing nudge bot, single loop |

---

## Costs and Risks

**Token cost:** Each retry burns premium requests. With backoff and escalation, a worst-case entry could consume: 3 × 0.33 (Haiku) + 2 × 1.0 (Sonnet) + 1 × 3.0 (Opus) = ~6 premium requests before quarantine. This is meaningful but bounded.

**Complexity:** The steward adds a new layer of control flow. Risk: steward bugs cause worse behavior than no steward (retrying endlessly, escalating when shouldn't). Mitigation: conservative defaults, quarantine as safety net, kill switch.

**Silent loops:** Risk of the steward retrying quietly and burning budget while Michael sleeps. Mitigation: daily cost cap per entry, circuit breaker, wake-hours awareness (borrow from nudge bot).

**Over-engineering:** This proposal is detailed for clarity, but Phase 1 is intentionally small. If Phase 1 doesn't prove its value, Phases 2-5 may not be warranted.

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| **Intent** | Why are we doing this? | To reduce the manual burden of diagnosing and retrying pipeline failures. The steward loop is what Michael already does by hand — we're encoding it. |
| **Covenant** | Rules of engagement? | Brain patterns (Go, SQLite WAL, WebSocket). Steward decisions logged and visible. Human override always available. |
| **Stewardship** | Who owns what? | New `internal/steward/` package. Dev agent builds. Human deploys. |
| **Spiritual Creation** | Spec precise enough? | Phase 1 is well-defined. Later phases sketch direction. |
| **Line upon Line** | Phasing? | Phase 1 stands alone. Each phase adds value independently. |
| **Physical Creation** | Who executes? | Dev agent with human review. |
| **Review** | How do we know it's right? | Verification criteria per phase. E2E test with deliberate failures. |
| **Atonement** | What if it goes wrong? | Kill switch on steward loop. Quarantine as safety net. Cost caps. |
| **Sabbath** | When do we stop and reflect? | After Phase 1 (does the basic loop prove its value?). |
| **Consecration** | Who benefits? | Michael directly. Others indirectly (pattern is reusable). |
| **Zion** | How does this serve the whole? | Moves brain from "dispatch-and-hope" to "shepherd-to-completion." |

---

## Recommendation

**Build.** Phase 1 first — minimal steward loop with retry-with-context. It solves the most common pain (execution timeout/failure requiring manual re-trigger) with bounded complexity. Evaluate after Phase 1 whether escalation and circuit breakers are needed.

**Depends on:** Phase 4 timeout fix (inactivity-based timeout) from [brain-pipeline-fixes-phase4.md](../brain-pipeline-fixes-phase4.md) should ship first. The steward needs intelligent timeout to know what "inactivity failure" means.

**Execute with:** dev agent. Phase 1 is a clean 2-3 session build.
