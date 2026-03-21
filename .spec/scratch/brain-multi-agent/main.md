# Scratch: Brain Multi-Agent Routing (WS1 Phase 3)

**Binding problem:** brain.exe can run individual agent sessions (via `--agent study` or `POST /api/agent/ask`) but cannot route — cannot take a captured thought and determine which agent should handle it, cannot manage multiple concurrent agent sessions, and cannot enforce governance on agent tool use. Everything requires explicit human selection.

**Proposal:** `.spec/proposals/brain-multi-agent/main.md`
**Status:** Research complete. Writing proposal.

---

## Architecture Inventory

### Current State (as of Mar 21)

**Entry flow:** Relay/Discord/Web → classifier (qwen3.5-9b) → store (SQLite) → STOP. No agent routing.

**Classifier categories:** people, projects, ideas, actions, study, journal (6 total)

**Agent infrastructure:**
- Single `Agent` struct in `internal/ai/agent.go`
- One session per Agent instance
- `POST /api/agent/ask` uses the single agent
- `brain exec --agent <name>` creates a one-off session
- 11 agents parsed from `.github/agents/` (study, dev, lesson, eval, journal, plan, review, talk, podcast, story, ux)
- 7 MCP servers auto-discovered

**SDK hook types available (v0.1.32):**
- `OnPreToolUse(PreToolUseHookInput, HookInvocation) → PreToolUseHookOutput`
  - Input: Timestamp, Cwd, ToolName, ToolArgs
  - Output: PermissionDecision (allow/deny/ask), PermissionDecisionReason, ModifiedArgs, AdditionalContext, SuppressOutput
- `OnPostToolUse(PostToolUseHookInput, HookInvocation) → PostToolUseHookOutput`
  - Input: Timestamp, Cwd, ToolName (confirmed from agent.go usage)
- `OnUserPromptSubmitted`
- `OnSessionStart` / `OnSessionEnd`
- `OnErrorOccurred`

### Key Gap

Categories ≠ Agents. The classifier has 6 categories; the workspace has 11 agents. The mapping is not 1:1:

| Category | Agent(s) | Notes |
|----------|----------|-------|
| study | study | Direct match |
| journal | journal | Direct match |
| ideas | plan | Ideas → spec/planning? Or just store? |
| actions | dev? | Actions could be dev tasks, but also "buy milk" |
| projects | dev | Active code projects |
| people | — | No agent for people notes |

Only 2 of 6 categories map cleanly to agents. The others either don't need agents or need classification refinement.

### What the Overview Plan Actually Says

Phase 3 spec from overview/main.md:
> brain.exe routes a captured idea to the appropriate agent session
> Pattern: Capture → classify → if "spec-worthy" → create proposal skeleton → assign to agent
> Verify: End-to-end: brain capture → proposal draft appears in `.spec/proposals/`

The Squad learnings expanded this to:
- A2: Agent routing table
- A3: Hook-based governance (Go middleware on tool calls)
- A4: Reviewer lockout with model escalation

### Critical Analysis

**Is this the RIGHT thing to build?** Yes, but the scope matters. The spec says "routes captured idea to agent" — that's clear. But the Squad additions (A3, A4) are governance and review patterns that matter when agents run unsupervised. Michael explicitly said agents should wait for human-assigned specs (gated autonomy decision). So A3/A4 are forward-looking, not immediately needed.

**What's the simplest version that would be useful?**
1. Route study entries to the study agent automatically
2. Route idea entries to... what? A plan agent that creates a proposal skeleton? That's exactly what the spec says.
3. Route action entries to... a todo list? Not an agent.

The minimum useful routing: study entries → study agent → produce study document. That's the proven path — we validated it in Phase 2.5.

**What gets worse?**
- Token cost: automatic routing means automatic agent sessions → automatic token consumption
- Complexity: multi-session management, session lifecycle, cleanup
- Failure modes: agent produces bad output with no human review

**Mosiah 4:27 check:** Michael's at 56% of premium requests with 1/3 month remaining. Automatic agent sessions would burn through the rest fast. Governance (A3) would help control this, but governance is the complex part.

### Design Decision: Pool vs Pool-Per-Agent

**Option A: Agent pool** — One pool of Agent instances, each with its own session. Route by creating the right Agent for the task.

**Option B: Single agent, dynamic system message** — One Agent struct, but swap the system message based on routing. Cheaper (one session), but loses the benefit of per-agent conversation history.

**Option C: Keep `brain exec`, add routing** — Don't change the daemon; add classification-based routing that spawns `brain exec --agent <name> --prompt <text>` as a subprocess. Simple, uses what works, but expensive (new session per task).

**Recommendation: A with lazy instantiation.** Create Agent instances on first use for a given agent name. Sessions are expensive, so don't pre-create all 11. The pool is a map[string]*Agent.

### Phasing Recommendation

**Phase 3a: Agent Router + Session Pool (1 session)** — The minimum. Classification adds "route to agent" decision. Agent pool manages per-agent sessions. Study entries get routed. No governance hooks, no reviewer lockout.

**Phase 3b: Hook-Based Governance (1 session)** — File-write guards, token budgets, audit logging via OnPreToolUse/OnPostToolUse. This is what makes unsupervised execution safe.

**Phase 3c: Reviewer Lockout + Escalation (1 session)** — Rejected work re-routed. Model escalation before agent escalation. This is what makes quality consistent.

Phase 3a is useful alone. Phase 3b is needed before any autonomy. Phase 3c is needed before Michael stops reviewing every output.

---

## SDK CustomAgents Research

The SDK has `CustomAgents []CustomAgentConfig` but we're not using it. In Phase 2.5 we embedded the agent prompt in SystemMessage instead. Two questions:

1. Does `CustomAgents` enable in-session routing? (i.e., can the LLM hand off to a named agent without the caller specifying?)
2. If so, would a coordinator agent + CustomAgents handle multi-agent routing at the SDK level?

We chose SystemMessage embedding because it was simpler and proven. CustomAgents might be the right approach for Phase 3 if it enables automatic routing. But testing is needed — this is SDK-level routing, not our code.

**Decision for now:** Build routing in Go code (agent pool, classification-based selection). Don't depend on CustomAgents until we've tested it. We can migrate later if CustomAgents proves useful.
