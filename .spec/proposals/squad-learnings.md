# Squad Learnings — What to Adopt

**Binding problem:** Our agentic infrastructure (WS1) is theoretical. Squad is a working production multi-agent runtime. We need to incorporate proven patterns into our WS1 spec before building.

**Created:** 2026-03-19
**Research:** [.spec/scratch/squad-analysis/main.md](../../scratch/squad-analysis/main.md)
**Source:** [bradygaster/squad](https://github.com/bradygaster/squad) (cloned to `external_context/squad/`)
**Status:** Recommended for adoption. Updates WS1 Phase 3 and adds new infrastructure.

---

## 1. Problem Statement

WS1 Phase 3 (Multi-Agent Routing) is specified as:

> brain.exe routes a captured idea to the appropriate agent session. Pattern: Capture → classify → if "spec-worthy" → create proposal skeleton → assign to agent.

This is thin. Squad shows us that multi-agent coordination requires at minimum:
- Routing rules (not just classification)
- Shared decision state (not just per-session memory)
- Governance enforcement (not just prompt instructions)
- Review gates between agents (not just human review)
- Model tiering (not just "pick a model")
- Cost tracking (not running faster than you have strength)

Our 11-step cycle provides the WHY. Squad provides the HOW for steps 3-7.

---

## 2. What We're Adopting

### A1. Decisions File (Adopt immediately)

**Source:** Squad's `decisions.md` + `decisions/inbox/` pattern.

**Action:** Create `.spec/memory/decisions.md` as a canonical, structured decisions log. All agents read this as session context. Format:

```markdown
## Decision: {Title}
- **Date:** YYYY-MM-DD
- **Decided by:** {who}
- **Decision:** {what}
- **Rationale:** {why}
- **Supersedes:** {previous decision, if any}
```

**Why now:** Decisions currently scatter across active.md, guidance.md, agent conversations. When agents can't find prior decisions, they re-ask or assume. This is a 15-minute task that improves every future session.

**Migration:** Extract key decisions from active.md into decisions.md. active.md keeps only *current state* (what's in flight, what's blocked). decisions.md keeps *settled questions* (what we decided and why).

### A2. Agent Routing Table (Adopt when building WS1 Phase 3)

**Source:** Squad's `routing.md` — Work Type → Agent table with primary/secondary ownership.

**Action:** Create `.github/routing.md` mapping task types to agents:

| Work Type | Primary Agent | Secondary |
|-----------|--------------|-----------|
| Scripture study | study | study-exp1 |
| Lesson prep | lesson | lesson-exp1 |
| Talk prep | talk | — |
| Conference talk review | review | — |
| YouTube evaluation | eval | yt-exp1 |
| Journal/reflection | journal | — |
| Code implementation | dev | — |
| Planning/spec | plan-exp1 | — |
| Study→podcast format | podcast | — |
| Narrative writing | story | — |
| UI/UX design | ux | — |

**Why:** When brain.exe does automated routing, it needs a lookup table, not just LM classification. Classification determines the work type; the routing table determines the agent.

### A3. Hook-Based Governance (Adopt in WS1 Phase 3)

**Source:** Squad's HookPipeline — programmatic pre/post tool-use interception.

**Action:** When brain.exe orchestrates agent sessions via Copilot SDK, implement Go middleware for tool-call governance:

```go
type Hook struct {
    OnPreToolUse  func(ctx context.Context, call ToolCall) (ToolCall, error) // modify or block
    OnPostToolUse func(ctx context.Context, call ToolCall, result any) error  // audit/log
}
```

Governance policies (implemented as hooks, not prompts):
- **File-write guards:** Agents can only write to paths in their scope
- **Destructive operation blocking:** No `rm -rf`, no `git push --force`, no `DROP TABLE`
- **Token budget:** Per-agent, per-session token ceiling
- **Audit logging:** All tool calls logged with agent identity

**Key Squad insight:** "Prompts can be ignored. Hooks are code."

### A4. Reviewer Lockout (Adopt in WS1 Phase 3)

**Source:** Squad's `reviewer-protocol` skill — rejected author locked out, different agent must revise.

**Action:** When brain.exe routes work that gets rejected at review:
1. The original agent is locked out of the revision
2. A different agent (same domain or secondary from routing table) handles the fix
3. If no alternate exists → escalate to human
4. Scope is per-artifact (rejecting one file doesn't lock out the whole agent)

**Why:** Prevents defensive loops. The fresh perspective often finds what the original missed.

### A5. Response Tier / Model Selection (Adopt in WS1 Phase 2-3)

**Source:** Squad's 4-tier response system + cost-first heuristic.

**Action:** Implement model selection as a function of task type in brain.exe:

| Task Type | Model | Backend | Rationale |
|-----------|-------|---------|-----------|
| Classification | qwen3.5-9b | LM Studio | Free, local, fast, proven |
| Quick lookup | Haiku 3.5 | Copilot SDK | Cheap, fast |
| Study/content | Sonnet 4.6 | Copilot SDK | Quality writing |
| Architecture/spec | Opus 4.6 | Copilot SDK | Deep reasoning |
| Code generation | Sonnet 4.6 | Copilot SDK | Quality + speed |

**Selection hierarchy:** User override → task type match → default (Sonnet 4.6).
**Fallback chain:** Primary → secondary → omit model param (let Copilot pick).

### A6. Cost Tracking (Adopt in WS1 Phase 2)

**Source:** Squad's CostTracker — per-agent token accumulation.

**Action:** Add token counting to brain.exe's Copilot SDK wrapper:

```go
type SessionMetrics struct {
    AgentID      string
    InputTokens  int64
    OutputTokens int64
    Model        string
    StartedAt    time.Time
    Duration     time.Duration
}
```

Store in SQLite alongside brain entries. Surface in brain-app dashboard.

**Consecration principle:** Know exactly where tokens go. If study sessions cost 10x dev sessions, that's a conscious choice — not an invisible drain.

---

## 3. What We're NOT Adopting

| Squad Feature | Why Not |
|---|---|
| Casting (movie universe names) | Our agents have meaningful purpose-names. Personality through voice, not branding. |
| Squad CLI / interactive shell | We have VS Code + brain.exe. No need for a third interface. |
| 20+ micro-specialized agents | Our 14 agents cover modes of engagement, not micro-domains. Right model for one person. |
| TypeScript SDK | We build in Go. Copilot SDK available in Go. |
| OpenTelemetry infrastructure | Overkill for personal scale. SQLite metrics sufficient. |
| Scribe as a separate agent | Our session-journal binary + memory update ritual covers this. Improve the ritual, don't add an agent. |

---

## 4. What Squad Could Learn From Us (Confidence Check)

Not our job to teach them, but this validates our framework:

| Our Pattern | Squad Gap |
|---|---|
| **Intent hierarchy** (intent.yaml) | Squad has project context but no root values document |
| **Mutual covenant** (human obligations) | Squad governs agents only, not the human |
| **Progressive stewardship** (earned autonomy) | Squad has static roles — agents never grow |
| **Atonement** (failure → system growth) | Squad has retrospectives but no forward-recovery pattern |
| **Sabbath** (intentional rest) | Squad has no concept of stopping |
| **Consecration** (token purpose alignment) | Squad tracks cost but not purpose |
| **Relational memory** (what it meant, not just what happened) | Squad's Scribe logs facts, not meaning |

**Verdict:** Squad is strong on steps 3-7 (stewardship through review). We're strong on steps 1-2 and 8-11 (intent, covenant, and the redemptive/reflective patterns). The combination is powerful.

---

## 5. Impact on WS1

### Phase 1 (Copilot SDK + MCP Integration) — No change
Already scoped correctly. Extend brain.exe's existing SDK integration.

### Phase 2 (Agent as Spec Executor) — Add A5, A6
- Add response tier selection when the agent spawns sessions
- Add token counting/cost tracking to the SDK wrapper
- These are small additions that pay forward to Phase 3

### Phase 3 (Multi-Agent Routing) — Major expansion
The current spec says "brain.exe routes captured idea to appropriate agent." The expanded spec:

1. **Routing table** (A2) — classification determines work type, routing table determines agent
2. **Hook governance** (A3) — Go middleware on tool calls, not prompt instructions
3. **Reviewer lockout** (A4) — rejected work routed to different agent
4. **Multi-agent format** — assembled result at top, raw agent outputs in appendix (Squad pattern)
5. **Decision propagation** — agents read decisions.md, write to inbox, brain.exe merges

### New: A1 (Decisions File) — Immediate, no code needed
Create `.spec/memory/decisions.md` now. Start the habit before the infrastructure.

---

## 6. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | To build multi-agent orchestration that works, learning from production experience rather than inventing from scratch |
| Covenant | Rules of engagement? | Adopt patterns that are proven. Don't adopt patterns that serve a different scale or purpose. |
| Stewardship | Who owns what? | brain.exe owns orchestration (Go). Agents own their domains. Decisions.md owned by the human. |
| Spiritual Creation | Is the spec precise enough? | A2-A6 need implementation specs when we reach WS1 Phase 3. A1 is ready now. |
| Line Upon Line | Phasing? | A1 now → A5-A6 in Phase 2 → A2-A4 in Phase 3 |
| Physical Creation | Who executes? | dev agent for implementation. plan-exp1 for Phase 3 spec expansion. |
| Review | How do we know it's right? | Each adoption item has a verification criterion. See Section 2. |
| Atonement | If it goes wrong? | These are additive patterns. If a hook is too restrictive, relax it. If cost tracking is noisy, reduce granularity. All reversible. |
| Sabbath | When do we stop? | After A1 (decisions file). Pause. After Phase 2 additions. Pause. After Phase 3 expansion. Full sabbath review. |
| Consecration | Who benefits? | Michael directly. Eventually the Work-with-AI guide readers. |
| Zion | Serves the whole? | Yes — these patterns make every agent more reliable, every session more efficient, every decision more durable. |

---

## 7. Recommendation

**Proceed.** The Squad investigation validates our direction and fills concrete gaps:

1. **Now:** Create decisions.md (A1). 15 minutes. Immediate value.
2. **WS1 Phase 2:** Add response tiers (A5) and cost tracking (A6). Small additions to existing plan.
3. **WS1 Phase 3:** Expand spec with routing table (A2), hook governance (A3), reviewer lockout (A4). This is the big change — Phase 3 is no longer "route to agent" but "orchestrate agents with governance."

The combination of Squad's implementation patterns (steps 3-7) with our creation cycle's wisdom patterns (steps 1-2, 8-11) produces something neither has alone: **orchestration with purpose**.
