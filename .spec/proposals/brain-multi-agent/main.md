# Brain Multi-Agent Routing (WS1 Phase 3)

**Binding problem:** brain.exe can run individual agent sessions (via `brain exec --agent study` or `POST /api/agent/ask`) but has no routing — no way to automatically determine which agent should handle a captured thought, no way to manage multiple concurrent agent sessions, and no governance on agent tool use. Every agent interaction requires explicit human selection. The gap between "classifier puts entries in buckets" and "agents act on entries" is the entire automation story.

**Created:** 2026-03-21
**Research:** [.spec/scratch/brain-multi-agent/main.md](../../scratch/brain-multi-agent/main.md)
**Depends on:** WS1 Phase 2.5 (workspace-aware sessions) — SHIPPED
**Affects:** WS1 Phase 3, WS2 (brain consolidation), WS3 (study quality)
**Status:** Draft — awaiting review

---

## 1. Problem Statement

The current flow is:

```
Capture (relay/discord/web) → Classify (qwen3.5-9b) → Store (SQLite) → STOP
```

The desired flow is:

```
Capture → Classify → Route → Agent Session → Output → Review → Store
```

What's missing is the middle: Route + Agent Session + Output + Review.

### What This Costs Today

- **Manual overhead.** Every agent interaction requires Michael to type `brain exec --agent study --prompt "..."` or use `POST /api/agent/ask`. The classifier already knows an entry is a "study" thought — it just doesn't do anything with that information.
- **Lost momentum.** A thought captured at 2am via brain-app gets classified and stored. By the time Michael sits down to route it to an agent, the moment has passed.
- **Single-session limitation.** The daemon's `Agent` struct holds one session. If a study is running and Michael asks a dev question, one has to wait.
- **No governance.** When agents run, they can write anywhere, use unlimited tokens, and execute any tool. The `PermissionHandler.ApproveAll` in `createSession()` approves everything without inspection.

### Success Criteria

1. **Routing works.** A "study" entry captured via relay produces a study document in `study/` without manual agent selection.
2. **Multi-session works.** The daemon can manage sessions for different agent types concurrently (study session + dev session can coexist).
3. **Governance hooks exist.** File-write paths are scoped per agent. Token usage is tracked per session.
4. **Entry-to-output link.** The source entry in SQLite is linked to the agent output (file path, status, tokens used).

---

## 2. Constraints & Boundaries

**In scope (Phase 3a — routing + pool):**
- Agent session pool: `map[string]*Agent` with lazy creation
- Routing table: classifier category → agent name mapping
- Route trigger: post-classification hook that decides whether to route an entry to an agent
- `POST /api/agent/ask` updated to accept optional `agent` parameter for named sessions
- Entry status tracking: pending → routed → complete → reviewed

**In scope (Phase 3b — governance):**
- `OnPreToolUse` hook: file-write path scoping per agent, destructive operation blocking
- `OnPostToolUse` hook: audit logging (tool name, agent, args, timestamp)
- Per-session token tracking (from `session.usage_info` events)
- Token budget per agent (configurable, with warning threshold)

**Out of scope (deferred to Phase 3c or later):**
- Reviewer lockout with model escalation (A4) — needs Phase 3b governance first
- Multi-agent handoffs (agent A passes work to agent B mid-task)
- SDK `CustomAgents` field testing — we use Go-level routing, not SDK routing
- Automated review (human reviews all agent output for now)
- UI for managing agent sessions (brain-app or web dashboard)

**Conventions:**
- Go. Same packages (`internal/ai`, `internal/config`, `internal/store`).
- No new dependencies.
- Agents share MCP server configs (all see all 7 servers). Per-agent tool filtering deferred.

---

## 3. Prior Art & Related Work

| Source | What we learned |
|--------|-----------------|
| [Squad learnings](../../proposals/squad-learnings.md) | A2 (routing table), A3 (hook governance), A4 (reviewer lockout). All three deferred to Phase 3. |
| [Workspace-aware sessions](../../proposals/brain-workspace-aware/main.md) | Agent parser, skill directories, infinite sessions — all SHIPPED. Foundation for routing. |
| [Overview plan](../../proposals/overview/main.md) WS1 Phase 3 | Spec: "Capture → classify → if spec-worthy → create proposal skeleton → assign to agent" |
| Current classifier | 6 categories (people, projects, ideas, actions, study, journal). Proven with 3 models. |
| Current Agent struct | Single session, reused across calls. `Ask()` and `AskStreaming()`. Lazy session creation on first call. |
| SDK hooks | `OnPreToolUse` can allow/deny/modify tool calls. `OnPostToolUse` can audit. `SessionHooks` struct with 6 hook types. |
| [Gated autonomy decision](../../memory/decisions.md) | "Agents wait for human-assigned specs." Routing must respect this — auto-route only when explicitly enabled, not by default. |

---

## 4. Proposed Approach

### 4.1 Routing Table

A simple Go map, not a file. The routing table maps classifier categories to agent names with a routing mode:

```go
type RouteRule struct {
    AgentName string    // which agent handles this category
    Mode      string    // "auto" | "suggest" | "none"
    Prompt    string    // template for the agent prompt
}

var defaultRoutes = map[string]RouteRule{
    "study":    {AgentName: "study",   Mode: "suggest", Prompt: "Study this insight: {{.Body}}"},
    "journal":  {AgentName: "journal", Mode: "suggest", Prompt: "Reflect on this: {{.Body}}"},
    "ideas":    {AgentName: "plan",    Mode: "suggest", Prompt: "Evaluate this idea: {{.Body}}"},
    "projects": {AgentName: "dev",     Mode: "none",    Prompt: ""},
    "actions":  {AgentName: "",        Mode: "none",    Prompt: ""},
    "people":   {AgentName: "",        Mode: "none",    Prompt: ""},
}
```

**Modes:**
- `"auto"` — route immediately after classification (future, when governance is in place)
- `"suggest"` — mark entry as "agent-eligible" and surface in brain-app/web UI for one-click routing
- `"none"` — just store, no agent involvement

**Default mode is `"suggest"`, not `"auto"`.** This respects the gated autonomy decision. Michael sees "this study entry could go to the study agent" and clicks to approve. Full auto-routing is a config change once governance (Phase 3b) ships.

### 4.2 Agent Session Pool

Replace the single `Agent` in `main.go` with a pool:

```go
// internal/ai/pool.go

type AgentPool struct {
    client     *copilot.Client
    baseConfig AgentConfig          // shared config (MCP servers, working dir, skills)
    agents     map[string]*Agent    // lazy-created, keyed by agent name
    mu         sync.RWMutex
}

func NewAgentPool(client *copilot.Client, baseCfg AgentConfig) *AgentPool

// GetOrCreate returns the agent for the given name, creating it if needed.
// The agent's system message is composed from workspace config + agent prompt.
func (p *AgentPool) GetOrCreate(agentName string, wc config.WorkspaceConfig) *Agent

// ActiveSessions returns the names of agents with live sessions.
func (p *AgentPool) ActiveSessions() []string

// Reset destroys a specific agent's session.
func (p *AgentPool) Reset(agentName string)

// ResetAll destroys all sessions.
func (p *AgentPool) ResetAll()
```

The daemon creates an `AgentPool` instead of a single `Agent`. The web API routes to the pool:

- `POST /api/agent/ask` — unchanged behavior (uses default agent from pool)
- `POST /api/agent/ask` with `{"agent": "study", "prompt": "..."}` — routes to named agent
- `POST /api/agent/reset` with optional `{"agent": "study"}` — resets specific or all
- `GET /api/agent/sessions` — (new) lists active sessions

### 4.3 Route Trigger

After classification, the store checks the routing table and annotates the entry:

```go
// In store.Save() — after classification

route := routing.Lookup(entry.Category)
if route.AgentName != "" && route.Mode != "none" {
    entry.AgentRoute = route.AgentName
    entry.RouteStatus = "suggested"  // or "pending" if auto
}
```

New fields on Entry:
- `AgentRoute string` — which agent should handle this (empty = no agent)
- `RouteStatus string` — "" | "suggested" | "pending" | "running" | "complete" | "failed"
- `AgentOutput string` — path to the output file (e.g., `study/captured-insight.md`)
- `TokensUsed int64` — tokens consumed by the agent session

### 4.4 Routing Execution

A new `Router` component connects classification → pool:

```go
// internal/ai/router.go

type Router struct {
    pool   *AgentPool
    store  *store.Store
    routes map[string]RouteRule
    wc     config.WorkspaceConfig
}

// RouteEntry sends an entry to the appropriate agent and stores the result.
func (r *Router) RouteEntry(ctx context.Context, entry *store.Entry) error {
    route := r.routes[entry.Category]
    if route.AgentName == "" || route.Mode == "none" {
        return nil
    }

    agent := r.pool.GetOrCreate(route.AgentName, r.wc)
    prompt := route.RenderPrompt(entry)

    entry.RouteStatus = "running"
    r.store.UpdateRouteStatus(entry.ID, "running")

    response, err := agent.AskStreaming(ctx, prompt, io.Discard)
    if err != nil {
        entry.RouteStatus = "failed"
        r.store.UpdateRouteStatus(entry.ID, "failed")
        return err
    }

    entry.RouteStatus = "complete"
    entry.AgentOutput = response // or extract file path from response
    r.store.UpdateRouteStatus(entry.ID, "complete")
    return nil
}
```

### 4.5 Governance Hooks (Phase 3b)

When ready for auto-routing, add governance via SDK hooks:

```go
// internal/ai/governance.go

type GovernanceConfig struct {
    AllowedPaths  map[string][]string // agent name → allowed write path prefixes
    BlockedTools  []string            // globally blocked tool patterns
    TokenBudget   int64               // max tokens per session before warning
    TokenHardCap  int64               // max tokens per session before termination
}

func NewGovernanceHooks(cfg GovernanceConfig) *copilot.SessionHooks {
    return &copilot.SessionHooks{
        OnPreToolUse: func(input copilot.PreToolUseHookInput, inv copilot.HookInvocation) (*copilot.PreToolUseHookOutput, error) {
            // Check file-write paths
            if isWriteTool(input.ToolName) {
                path := extractPath(input.ToolArgs)
                if !isAllowedPath(path, cfg.AllowedPaths[agentName]) {
                    return &copilot.PreToolUseHookOutput{
                        PermissionDecision:       "deny",
                        PermissionDecisionReason: fmt.Sprintf("Agent %s cannot write to %s", agentName, path),
                    }, nil
                }
            }

            // Block destructive operations
            if isDestructive(input.ToolName, input.ToolArgs) {
                return &copilot.PreToolUseHookOutput{
                    PermissionDecision:       "deny",
                    PermissionDecisionReason: "Destructive operations blocked",
                }, nil
            }

            return &copilot.PreToolUseHookOutput{PermissionDecision: "allow"}, nil
        },

        OnPostToolUse: func(input copilot.PostToolUseHookInput, inv copilot.HookInvocation) (*copilot.PostToolUseHookOutput, error) {
            // Audit log
            log.Printf("AUDIT: agent=%s tool=%s", agentName, input.ToolName)
            return nil, nil
        },
    }
}
```

**Agent path scoping (default):**

| Agent | Allowed write paths |
|-------|-------------------|
| study | `study/`, `study/.scratch/` |
| lesson | `lessons/` |
| journal | `journal/` |
| plan | `.spec/proposals/`, `.spec/scratch/` |
| dev | `scripts/`, `internal/` |
| eval | `study/yt/` |
| review | `study/talks/` |
| talk | `callings/` |

### 4.6 Token Tracking

Wire into the `session.usage_info` events already flowing through `AskStreaming`:

```go
case "session.usage_info":
    if event.Data.InputTokens != nil {
        session.inputTokens += *event.Data.InputTokens
    }
    if event.Data.OutputTokens != nil {
        session.outputTokens += *event.Data.OutputTokens
    }
```

Store per-session and per-agent cumulative token counts. Surface in `GET /api/agent/sessions` response.

---

## 5. Phased Delivery

### Phase 3a: Agent Pool + Routing Table (1 session)

**Deliverables:**
1. `internal/ai/pool.go` — AgentPool with `GetOrCreate`, `Reset`, `ResetAll`, `ActiveSessions`
2. `internal/ai/router.go` — Router with `RouteRule`, routing table, `RouteEntry`
3. Update `internal/store/types.go` — add `AgentRoute`, `RouteStatus`, `AgentOutput`, `TokensUsed` fields
4. Update `internal/store/db.go` — migration for new columns, `UpdateRouteStatus` method
5. Update `cmd/brain/main.go` — replace single Agent with AgentPool + Router
6. Update `internal/web/server.go` — `POST /api/agent/ask` accepts `agent` field, `GET /api/agent/sessions`
7. Route table with all 6 categories mapped (study, journal, ideas → suggest; others → none)

**Verification:**
- `POST /api/agent/ask {"agent": "study", "prompt": "Study D&C 93:36"}` creates a study session and returns results
- `GET /api/agent/sessions` shows one active session
- Entry classified as "study" gets `AgentRoute: "study"` and `RouteStatus: "suggested"`
- `POST /api/agent/route {"entry_id": "..."}` triggers routing for a suggested entry

**Value delivered:** Multi-session support and routing infrastructure. Human still approves routing (suggest mode). Token tracking visible.

### Phase 3b: Governance Hooks (1 session)

**Deliverables:**
1. `internal/ai/governance.go` — GovernanceConfig, path scoping, destructive operation blocking
2. Update `internal/ai/agent.go` — accept GovernanceConfig, wire into SessionHooks.OnPreToolUse/OnPostToolUse
3. Token tracking from usage_info events
4. Token budget config with warning/hard-cap thresholds
5. Audit log (agent, tool, args, timestamp) — log.Printf initially, SQLite later

**Verification:**
- Study agent can write to `study/` but `OnPreToolUse` denies writes to `scripts/`
- Token usage appears in `GET /api/agent/sessions`
- Destructive commands (rm, git push --force) are blocked

**Value delivered:** Safe enough for "suggest + auto-approve" mode on study entries. The path to auto-routing.

### Phase 3c: Auto-Routing + Review (1 session)

**Deliverables:**
1. Route mode `"auto"` — entries routed immediately after classification
2. Review queue: `GET /api/agent/review` — pending agent outputs for human review
3. Accept/reject workflow: `POST /api/agent/review/{id}` with `{"action": "accept|reject"}`
4. On reject: entry status → "failed", output optionally deleted or archived

**Verification:**
- Capture a study thought via relay → classified as "study" → auto-routed to study agent → output appears in review queue → Michael approves → file stays in `study/`
- End-to-end without manual `brain exec`

**Value delivered:** The full loop. Capture → agent → review → output.

---

## 6. Costs & Risks

| Cost | Severity | Mitigation |
|------|----------|------------|
| Token consumption | High | Suggest mode first; governance + token budgets before auto |
| Multiple sessions = multiple MCP server processes | Medium | Lazy creation; consider sharing MCP servers across sessions |
| Complexity in daemon | Medium | Clean separation: pool.go, router.go, governance.go |
| Agent produces bad output unnoticed | High | Human review queue (Phase 3c); no auto-routing until Phase 3b governance ships |
| Session memory grows unbounded | Low | InfiniteSessions already handles compaction |
| Entry schema changes need DB migration | Low | Single migration with new nullable columns |

---

## 7. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | To close the gap between classification and action — make the brain actually *do* things with captured thoughts |
| Covenant | Rules of engagement? | Suggest-first, not auto-first. Governance before autonomy. Human reviews all output. Gated autonomy decision respected. |
| Stewardship | Who owns what? | brain.exe owns routing and pool. Each agent owns its domain output. Michael owns review. |
| Spiritual Creation | Is the spec precise enough? | Phase 3a is precise enough for the dev agent. Phases 3b-3c are directional. |
| Line upon Line | What's the phasing? | 3a (pool + routing) → 3b (governance) → 3c (auto + review). Each phase stands alone. |
| Physical Creation | Who executes? | dev agent, with human reviewing the resulting code |
| Review | How do we know it's right? | Verification criteria per phase. End-to-end: capture → agent → output file |
| Atonement | What if it goes wrong? | Suggest mode = no automatic damage. Failed routes don't lose data (entry still in SQLite). Session reset available. |
| Sabbath | When do we stop and reflect? | After Phase 3a ships. Does routing feel right? Is the token cost acceptable? |
| Consecration | Who benefits? | Michael immediately. Eventually brain-app users. Study quality scales. |
| Zion | How does this serve the whole? | Routing makes the entire agent ecosystem useful. Without it, agents are CLI tools. With it, they're part of the living system. |

---

## 8. Recommendation

**Build Phase 3a.** The routing table + agent pool is ~150 lines of Go, uses proven infrastructure (the Agent struct works, the classifier works), and delivers visible value: multi-session support and entry-to-agent linking.

**Defer 3b and 3c.** Governance and auto-routing are important but not urgent. In suggest mode, Michael is the governance layer. Auto-routing should wait until:
1. Token budget from premium requests is better understood (56% used with 1/3 month remaining)
2. Study agent output quality is validated over multiple runs (only-begotten was one test)
3. The review workflow is designed (how does Michael see and approve agent output in brain-app?)

**Phase 3a is a 1-session build.** If it works well, Phase 3b follows naturally. If the routing concept doesn't feel right in practice, we've invested one session, not three.
