# WS1 Phase 3c Revision: SDK Custom Agents + Auto-Routing

**Extends:** [brain-multi-agent/main.md](brain-multi-agent/main.md) Phase 3c
**Created:** 2026-04-02
**Status:** Draft — awaiting review

---

## Problem Statement

Phase 3c as currently spec'd delivers auto-routing and a review queue — which closes the "capture → agent → output" loop. But it leaves brain.exe at **Go-level routing only**: the AgentPool creates individual sessions per agent name, each with its own system message. The SDK doesn't know agents exist as distinct entities.

The Copilot SDK v0.1.32 already has `CustomAgentConfig` with:
- `Tools []string` — scope which tools each agent can see
- `Prompt string` — per-agent system message
- `MCPServers map[string]MCPServerConfig` — per-agent MCP servers
- `Infer *bool` — whether the SDK runtime can auto-delegate to this agent

This means the SDK can handle agent selection and tool scoping natively, which gives us:
1. **Intent-based delegation** for interactive sessions (user asks brain a question → SDK routes to the right agent)
2. **Tool restriction at the SDK level** — complementing Go-level governance
3. **Lifecycle observability** — the SDK emits events when sub-agents are selected, started, completed, failed
4. **Level 3 autonomy** — the SDK decides which agent handles a request, human reviews output

The architecture needs BOTH routing paths:
- **Entry-triggered routing** (current Phase 3c plan) — captured thoughts auto-route to agents based on classification
- **Intent-triggered delegation** (SDK custom agents) — interactive conversations automatically route to the right agent

These are complementary, not competing.

---

## What Changes from Original Phase 3c

### Original Phase 3c (1 session)
1. Route mode "auto" — entries routed immediately after classification
2. Review queue: `GET /api/agent/review`
3. Accept/reject workflow: `POST /api/agent/review/{id}`
4. On reject: entry status → "failed"

### Revised Phase 3c (2 sessions)

**Session 1: Auto-Routing + Review Queue** (original Phase 3c, unchanged)

Ship the original plan. Entry-triggered routing with review queue. This closes the core loop.

**Session 2: SDK Custom Agent Integration**

Wire the existing agent definitions into SDK `CustomAgentConfig` so the SDK runtime handles delegation for interactive sessions.

---

## Session 2 Design: SDK Custom Agents

### Agent Definitions

Map existing workspace agents to `CustomAgentConfig` entries. Each agent gets:
- A scoped tool list (derived from governance `defaultAllowedWritePaths` + relevant MCP servers)
- The system message already built by `BuildSystemMessage(wc, agentName)`
- `Infer: true` for agents with clear intent patterns, `false` for specialized agents

```go
// internal/ai/custom_agents.go

func BuildCustomAgents(wc config.WorkspaceConfig) []copilot.CustomAgentConfig {
    infer := true
    noInfer := false

    return []copilot.CustomAgentConfig{
        {
            Name:        "study",
            DisplayName: "Study Agent",
            Description: "Deep scripture study — phased writing with externalized memory and critical analysis",
            Prompt:       BuildSystemMessage(wc, "study"),
            Tools:        studyTools(),   // read_file, grep_search, gospel tools, webster, create_file, etc.
            Infer:        &infer,
        },
        {
            Name:        "journal",
            DisplayName: "Journal Agent",
            Description: "Personal reflection, journaling, and becoming",
            Prompt:       BuildSystemMessage(wc, "journal"),
            Tools:        journalTools(),
            Infer:        &infer,
        },
        {
            Name:        "plan",
            DisplayName: "Plan Agent",
            Description: "Planning — from idea to spec with critical analysis",
            Prompt:       BuildSystemMessage(wc, "plan"),
            Tools:        planTools(),
            Infer:        &infer,
        },
        {
            Name:        "dev",
            DisplayName: "Dev Agent",
            Description: "MCP server and tool development",
            Prompt:       BuildSystemMessage(wc, "dev"),
            Tools:        devTools(),
            Infer:        &noInfer, // dev tasks need explicit delegation
        },
        {
            Name:        "eval",
            DisplayName: "Eval Agent",
            Description: "YouTube video evaluation with charitable critical analysis",
            Prompt:       BuildSystemMessage(wc, "eval"),
            Tools:        evalTools(),
            Infer:        &noInfer,
        },
    }
}
```

### Tool Scoping

Each agent gets a subset of available tools. This complements the Go-level governance (which blocks write paths) with SDK-level restriction (which hides tools entirely):

| Agent | Visible Tools | Hidden Tools |
|-------|--------------|-------------|
| study | read_file, grep_search, file_search, create_file, replace_string_in_file, gospel_search, gospel_get, webster_define, semantic_search | run_in_terminal, git operations |
| journal | read_file, create_file, replace_string_in_file, becoming tools | run_in_terminal, gospel tools, git |
| plan | read_file, grep_search, file_search, create_file, replace_string_in_file, list_dir | run_in_terminal, git |
| dev | ALL tools (full access needed for development) | — |

Tool names must match the exact tool names the MCP servers register. Need to enumerate these at startup from the connected MCP servers.

### Integration with AgentPool

Two approaches:

**Option A: Replace AgentPool sessions with SDK custom agents.**
- Pro: Single source of truth for agent definition.
- Con: Breaks entry-triggered routing (SDK custom agents are for interactive use).

**Option B (Recommended): Parallel paths — AgentPool for entry routing, SDK custom agents for interactive.**
- Entry-triggered routing continues using `AgentPool.GetOrCreate()` → individual sessions with governance hooks
- Interactive `POST /api/agent/ask` sessions get `CustomAgents` wired in, so the SDK auto-delegates
- Both paths share the same system messages, tool scoping rules, and governance policy

```go
// In createSession() — add CustomAgents when pool has workspace config
func (a *Agent) createSession() (*copilot.Session, error) {
    cfg := copilot.SessionConfig{
        SystemMessage: a.config.SystemMessage,
        // ... existing config ...
    }

    // Wire SDK custom agents for interactive sessions
    if a.config.CustomAgents != nil {
        cfg.CustomAgents = a.config.CustomAgents
    }

    return a.client.NewSession(cfg)
}
```

The Default agent (unnamed, used for `POST /api/agent/ask` without specifying an agent name) gets all custom agents wired in. When a user sends "I want to study Alma 32" to the default agent, the SDK routes to the study custom agent automatically.

Named agents (from entry routing) don't get custom agents — they ARE the agent the router selected.

### Verification

1. `POST /api/agent/ask {"prompt": "Study the concept of faith in Alma 32"}` → SDK delegates to study custom agent → response uses gospel tools only
2. `POST /api/agent/ask {"prompt": "What's my practice streak?"}` → SDK delegates to journal agent → response uses becoming tools
3. `POST /api/agent/ask {"agent": "study", "prompt": "..."}` → Explicit agent selection bypasses SDK delegation (existing behavior)
4. Entry auto-routed as study → uses AgentPool study session → output in review queue (existing Phase 3c behavior)

---

## What This Achieves for Level 3 Autonomy

| Autonomy Level | What It Means | Where brain.exe Will Be |
|----------------|--------------|------------------------|
| Level 1 | Human does everything, AI assists | (where we started) |
| Level 2 | Human assigns specs, AI executes | (where we are — gated autonomy) |
| **Level 3** | **AI routes and executes, human reviews output** | **Phase 3c delivers this** |
| Level 4 | AI operates autonomously with exception-based human involvement | (future — requires extended track record at Level 3) |

Phase 3c Session 1 (auto-routing) gives Level 3 for **captured entries**: brain classifies, routes to the right agent, agent produces output, human reviews.

Phase 3c Session 2 (SDK custom agents) gives Level 3 for **interactive sessions**: brain receives a request, SDK routes to the right agent, agent executes with scoped tools, human reviews.

Together, they cover both paths into brain.exe — passive (capture) and active (conversation).

---

## Dependencies & Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| SDK custom agents may not work as documented at v0.1.32 | Medium | Types exist in source. Build a minimal test first before wiring into pool. |
| Tool name mismatch between MCP registration and CustomAgentConfig.Tools | Medium | Enumerate tools at startup from MCP server handshake responses |
| Two routing paths = complexity | Medium | Clean separation: pool.go handles entry routing, custom_agents.go handles SDK delegation |
| SDK v0.1.32 → v0.2.0 may change CustomAgentConfig | Low | Pin version, update when upgrading |

---

## Sequence

1. **Ship original Phase 3c** (auto-routing + review queue) — this is the core value
2. **Spike: test SDK custom agents** — create 1 custom agent (study), wire into default session, verify intent-based delegation works
3. **Build full custom agent set** — all 5 agents with tool scoping
4. **Integrate tool enumeration** — read available tools from MCP servers at startup
5. **Update brain-app / web UI** — show which agent is handling the current conversation

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Close the gap between Level 2 and Level 3 autonomy. Make brain.exe a multi-agent system, not a single-agent toolkit. |
| Covenant | Rules? | Suggest-first for entries (existing). SDK delegation for interactive. Human reviews all output. |
| Stewardship | Who owns what? | SDK owns delegation logic. brain.exe owns governance. Michael owns review. |
| Line upon Line | Phasing? | Phase 3c Session 1 first (core loop), Session 2 second (SDK enhancement). Either stands alone. |
| Sabbath | When stop? | After Session 1 ships and the full auto-route loop works end-to-end. |
