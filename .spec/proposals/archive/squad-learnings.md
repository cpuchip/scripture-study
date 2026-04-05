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

### A1. Decisions File ~~(Adopt immediately)~~ ✅ DONE (Mar 19)

**Source:** Squad's `decisions.md` + `decisions/inbox/` pattern.

**Completed:** `.spec/memory/decisions.md` created with 15 structured decisions extracted from active.md and guidance.md. Added to session-start sequence (step 4). intent.yaml added as step 1. active.md trimmed to reference decisions.md instead of duplicating.

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

### A4. Reviewer Lockout with Model Escalation (Adopt in WS1 Phase 3)

**Source:** Squad's `reviewer-protocol` skill — rejected author locked out, different agent must revise.

**Refinement (Michael, Mar 19):** Instead of just routing to a different agent, escalate the model tier. If haiku produced rejected work → retry with sonnet. If sonnet → try opus. If opus → try gpt-5.4 (noting that GPT models feel different for study work — better suited to dev/infra tasks). If all tiers exhausted → escalate to human.

**Action:** When brain.exe routes work that gets rejected at review:
1. **First:** Bump to the next model tier for the same agent
2. **If still rejected:** Route to a different agent (same domain or secondary from routing table)
3. **If no alternate exists or all fail:** Escalate to human
4. Scope is per-artifact (rejecting one file doesn't lock out the whole agent)

This combines Squad's lockout principle with a cost-efficient escalation: try the cheaper fix (better model) before the expensive fix (different agent re-reading all context).

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

### A7. Iterative Retrieval — Spawn Contracts (Adopt in WS1 Phase 3c)

**Source:** Squad's `iterative-retrieval` skill — 3-cycle max with structured spawn prompts.

**Problem:** Our current agent dispatch is "give it a prompt and see what comes back." No success criteria, no escalation path, no cycle cap.

**Action:** When brain.exe spawns sub-agents, every dispatch must include four sections:

```
## Task
{What needs done — concrete and bounded}

## WHY this matters
{Motivation + context. What system or user goal does this serve? What breaks if skipped?}

## Success criteria
{How to know the output is correct. Checkboxes, not vibes.}
- [ ] File X exists and contains Y
- [ ] No regressions in existing tests

## Escalation path
{What to do if stuck. "Stop and ask" is valid.}
- If requirements ambiguous → stop, surface to coordinator
- If blocked by dependency → note the block, explain
- If 3 cycles exhausted → write summary, escalate to human
```

**3-Cycle Protocol:**

| Cycle | Description | Exit |
|-------|-------------|------|
| 1 | Initial attempt | Done → validate. Incomplete → surface delta. |
| 2 | Targeted retry with specific corrections from cycle 1 | Done → validate. Incomplete → one more. |
| 3 | Final attempt with all context from 1-2 | Done or escalate — no cycle 4. |

**Key rules:**
- Coordinator validates output against success criteria between cycles (not just at end)
- Each subsequent cycle includes what was tried and what's still missing — not just the original prompt
- Cycle 3 exhausted → escalate to human with full context of all attempts

**Why this matters for us:** brain.exe Phase 3c auto-routes work to agents. Without spawn contracts, agents run unbounded, produce unvalidated output, and there's no escalation when they fail. The 3-cycle cap prevents infinite loops. The WHY context prevents agents from making scope trade-offs they don't have authority to make.

### A8. Reflect — In-Session Learning Capture (Adopted Mar 31)

**Source:** Squad's `reflect` skill — learning capture triggered by corrections, praise, edge cases.

**Problem:** Our `.spec/learnings/` captures post-incident analysis of big failures. Most of Michael's feedback is smaller: micro-corrections, tool preferences, formatting adjustments. These are lost by session end.

**Action:** Created `.github/skills/reflect/SKILL.md`. During sessions, corrections are logged to `.spec/scratch/reflect.md`. At session end, entries graduate to learnings/, preferences.yaml, decisions.md, or agent instructions as appropriate.

**Status:** Skill created. Needs to be listed in copilot-instructions.md and wired into all agents.

### A9. Task Coordination — Persistent Backlog (Evaluate for WS1 Phase 3)

**Source:** tpg (cpuchip/tpg) — SQLite-based task tracker designed specifically for AI agent session boundaries. Also: squad's 0% markdown / 85% Issues finding from retro-enforcement.

**Problem:** Our carry-forward items live in active.md (markdown) and session journal carry_forward fields (YAML). Squad measured 0% completion on markdown checklists across 6 retros, vs 85%+ when the same items were GitHub Issues. tpg addresses this directly: tasks with IDs, dependencies, progress logs, and stale detection in a local SQLite DB.

**Options under consideration:**

| Approach | Pros | Cons |
|----------|------|------|
| **tpg** (local SQLite) | Purpose-built for agent sessions, dependency-aware `ready` queue, context engine for learnings, Go binary, local-only | Another CLI tool to maintain, no notifications, no mobile access |
| **GitHub Issues** | Notifications, assignees, mobile access, proven 85% completion rate, existing ecosystem | Requires internet, public/private repo complexity, heavier than needed for personal scale |
| **brain.exe tasks table** | Already have the DB, already have the app, already have mobile via brain-app | Would need to add dependency tracking, progress logs, and `ready` queue semantics |
| **ibeco.me tasks** (existing) | Already built, already deployed, already has practices/journal | Missing dependency tracking, not agent-oriented, no `tpg prime` equivalent |

**Recommendation (Mar 31):** Evaluate tpg's architecture for integration into brain.exe rather than running it as a separate tool. The key patterns worth adopting:
1. **Dependency-aware ready queue** — `ready` only shows unblocked work
2. **Progress logs per task** — timestamped, not just status changes
3. **Context engine** — learnings tagged to concepts, two-phase retrieval
4. **Agent onboarding** — `prime` injects backlog into session start
5. **Stale detection** — in-progress tasks older than threshold get flagged

This needs a proper proposal when WS1 Phase 3c is scoped.

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
4. **Spawn contracts** (A7) — every agent dispatch gets Task + WHY + Success Criteria + Escalation Path, 3-cycle max
5. **Task coordination** (A9) — persistent backlog with dependency-aware ready queue (evaluate tpg patterns for brain.exe integration)
6. **Multi-agent format** — assembled result at top, raw agent outputs in appendix (Squad pattern)
7. **Decision propagation** — agents read decisions.md, write to inbox, brain.exe merges

### New: A1 (Decisions File) — Immediate, no code needed ✅ DONE
Create `.spec/memory/decisions.md` now. Start the habit before the infrastructure.

### New: A8 (Reflect Skill) — Immediate, no code needed ✅ DONE (Mar 31)
Created `.github/skills/reflect/SKILL.md`. In-session learning capture for micro-corrections.

---

## 6. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | To build multi-agent orchestration that works, learning from production experience rather than inventing from scratch |
| Covenant | Rules of engagement? | Adopt patterns that are proven. Don't adopt patterns that serve a different scale or purpose. |
| Stewardship | Who owns what? | brain.exe owns orchestration (Go). Agents own their domains. Decisions.md owned by the human. |
| Spiritual Creation | Is the spec precise enough? | A2-A6 need implementation specs when we reach WS1 Phase 3. A1 is ready now. |
| Line Upon Line | Phasing? | A1 now ✅ → A8 now ✅ → A5-A6 in Phase 2 → A2-A4, A7, A9 in Phase 3 |
| Physical Creation | Who executes? | dev agent for implementation. plan-exp1 for Phase 3 spec expansion. |
| Review | How do we know it's right? | Each adoption item has a verification criterion. See Section 2. |
| Atonement | If it goes wrong? | These are additive patterns. If a hook is too restrictive, relax it. If cost tracking is noisy, reduce granularity. All reversible. |
| Sabbath | When do we stop? | After A1 (decisions file). Pause. After Phase 2 additions. Pause. After Phase 3 expansion. Full sabbath review. |
| Consecration | Who benefits? | Michael directly. Eventually the Work-with-AI guide readers. |
| Zion | Serves the whole? | Yes — these patterns make every agent more reliable, every session more efficient, every decision more durable. |

---

## 7. Recommendation

**Proceed — but practice before building.**

The critical self-assessment (see scratch file Section X) found we practice ~28% of our own 11-step cycle. Before adopting 6 new patterns from Squad, we should operationalize what we already wrote.

### Phase 0: Practice What We Preach (Before any new code)

1. **Add intent.yaml to session-start sequence** in copilot-instructions.md — 5-minute edit
2. **Create decisions.md** (A1) — 15 minutes, immediate value
3. **Practice Sabbath** — after this session, stop. Let the work breathe.
4. **Promote exp1 agents** — they've proven themselves, make them the standard

### Phase 1: WS1 Phase 2 additions (When building)
- Response tiers (A5)
- Cost tracking (A6)

### Phase 2: WS1 Phase 3 expansion (After Phase 1 proves out)
- Routing table (A2)
- Hook governance (A3)
- Model-escalation lockout (A4)

### YouTube Review Gate
Michael has two YouTube videos covering similar ideas to Squad but without implementation. Review those BEFORE starting Phase 2. The videos may change the approach.

The combination of Squad's implementation patterns (steps 3-7) with our creation cycle's wisdom patterns (steps 1-2, 8-11) produces something neither has alone — **but only if we actually practice the wisdom patterns, not just describe them.**
