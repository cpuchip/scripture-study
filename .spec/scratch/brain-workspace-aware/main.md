# Scratch: Brain Workspace-Aware Agent Sessions

**Binding problem:** brain.exe's agent sessions operate as a generic "development agent" with no awareness of the workspace's agent definitions, skills, writing process, or project conventions. The result: studies without scratch files, no critical analysis phase, no Webster word study, no cross-study connections — despite the SDK supporting all the fields needed to replicate VS Code Copilot Chat's workspace awareness.

**Created:** 2026-03-20

---

## Research Findings

### SDK SessionConfig — Unused Fields (Go SDK v0.1.32)

From `go doc github.com/github/copilot-sdk/go.SessionConfig`:

| Field | Type | What it does | Currently used? |
|-------|------|-------------|-----------------|
| `CustomAgents` | `[]CustomAgentConfig` | Register named agents with prompts, tools, MCP servers | NO |
| `Agent` | `string` | (Not in Go SDK docs — may be TypeScript only) | NO |
| `SkillDirectories` | `[]string` | Point to directories containing skills | NO |
| `SystemMessage` | `*SystemMessageConfig` | Content + Mode ("append" or "replace") | YES but static hardcoded text |
| `InfiniteSessions` | `*InfiniteSessionConfig` | Context compaction for long sessions | NO |
| `ConfigDir` | `string` | Override config directory | NO |
| `DisabledSkills` | `[]string` | Disable specific skills | NO |

**Note:** Go SDK `SessionConfig` does NOT have an explicit `Agent` field. Need to verify whether CustomAgents routing happens differently.

### CustomAgentConfig Structure

```go
type CustomAgentConfig struct {
    Name        string            `json:"name"`
    DisplayName string            `json:"displayName,omitempty"`
    Description string            `json:"description,omitempty"`
    Tools       []string          `json:"tools,omitempty"`
    Prompt      string            `json:"prompt"`
    MCPServers  map[string]MCPServerConfig `json:"mcpServers,omitempty"`
    Infer       *bool             `json:"infer,omitempty"`
}
```

### How Squad Does It — Key Discovery

Squad does NOT use `CustomAgents` or `SkillDirectories`. It reads `.squad/` files from disk and embeds them into the SystemMessage. Each agent gets a separate Copilot SDK session with the agent's charter.md content as the system message.

Pattern:
1. Coordinator session: system message = team.md + routing.md
2. Agent sessions: system message = agent's charter.md
3. Skills: read from disk, embedded in prompts (not SDK fields)
4. Routing: coordinator LLM parses message → returns `ROUTE: agentName`
5. Agent sessions are lazily created and cached in a map

This means we have TWO valid approaches:
- **Approach A: SDK-native** — Use CustomAgents, SkillDirectories, and let the SDK handle agent/skill loading
- **Approach B: Squad-style** — Read files ourselves, embed in SystemMessage, manage sessions ourselves

### What VS Code Copilot Chat Does

VS Code reads `.github/copilot-instructions.md`, `.github/agents/*.agent.md`, and `.github/skills/*/SKILL.md` and maps them to the SDK's session config fields. This is the "full workspace experience."

### Current Agent System Prompt

brain.exe hardcodes a generic system prompt in both `run()` and `runExec()`:

```
You are a development agent for the scripture-study project. You have access to:
1. SCRIPTURE TOOLS (MCP): gospel_search, gospel_get, gospel_list, search_scriptures, search_talks, webster_define
2. BUILT-IN FILE TOOLS: You can read, search, and edit files in the workspace.
```

This says NOTHING about:
- Phased study workflow
- Scratch files
- Critical analysis
- Writing voice
- Session memory
- Source verification
- Scripture linking conventions

### MCP Server Auto-Discovery

Currently discovers 3 sibling servers:
- gospel-mcp (serve)
- gospel-vec (mcp)
- webster-mcp (serve)

Additional MCP servers that exist but aren't discovered:
- becoming (serves becoming/practice tools)
- byu-citations (BYU citation tools)
- search-mcp (Exa search)
- yt-mcp (YouTube tools)

### .github/ Directory Structure

```
.github/
├── copilot-instructions.md           — Root system prompt
├── agents/                           — 11 agents
│   ├── study.agent.md               — 7-phase study workflow
│   ├── lesson.agent.md              — Lesson planning
│   ├── ... (9 more)
├── skills/                          — 13 skills
│   ├── becoming/SKILL.md
│   ├── critical-analysis/SKILL.md
│   ├── deep-reading/SKILL.md
│   ├── dokploy/SKILL.md
│   ├── playwright-cli/SKILL.md
│   ├── publish-and-commit/SKILL.md
│   ├── quote-log/SKILL.md
│   ├── scripture-linking/SKILL.md
│   ├── source-verification/SKILL.md
│   ├── webster-analysis/SKILL.md
│   ├── wide-search/SKILL.md
│   ├── ben-test/SKILL.md
│   ├── byu-citations/SKILL.md
├── prompts/                         — Reusable prompt templates
```

---

## Compatibility Analysis: Remote Access vs. Autonomous Agents

### Goal 1: Remote Access (same as VS Code Copilot Chat)

brain.exe → Copilot SDK → same experience as VS Code Chat locally. Michael sends a prompt from brain-app/Discord/relay, agent works as if he's sitting at VS Code.

**What this needs:**
- Workspace-aware sessions (agents, skills, system instructions)
- Same MCP tools available
- Same file access
- Same model
- One session per user interaction

### Goal 2: Autonomous Agents (Squad pattern)

brain.exe orchestrates multiple agents working on different tasks. Coordinator routes work. Agents execute specs autonomously. Review gates between stages.

**What this needs:**
- Multiple concurrent sessions
- Routing logic (coordinator or classifier)
- Hook-based governance (code, not prompts)
- Per-agent session management
- Cost tracking

### Are They Compatible?

YES — they're the SAME infrastructure with different activation patterns:

| Feature | Remote Access | Autonomous |
|---------|--------------|------------|
| Session creation | On user request | On task assignment |
| Agent selection | User specifies (--agent study) | Coordinator decides |
| Session lifecycle | One per interaction | Pool, reuse, reset |
| Governance | User-supervised | Hook-enforced |
| System message | Agent prompt from .github/ | Same agent prompt |
| Skills | Loaded via SkillDirectories | Same |
| MCP servers | Same set | Same set |

The workspace-aware session is the FOUNDATION for both. Remote access is "human-triggered, single agent." Autonomous is "machine-triggered, multiple agents."

---

## Conflict Analysis with Existing Plans

### WS1 Phase 3: Multi-Agent Routing

Squad learnings proposal (A2-A4) specifies:
- Agent routing table
- Hook governance
- Reviewer lockout with model escalation

This proposal ENABLES Phase 3 by building the session infrastructure those features need. No conflict — this is the prerequisite.

### Cost Tracking (A6)

The squad proposal says add token counting. This is orthogonal — workspace awareness doesn't affect cost tracking. Can add later.

### Model Tiering (A5)

Response tier selection is per-agent and per-task. This proposal's AgentConfig already supports per-agent model selection via CustomAgentConfig. Compatible.

### Decision: Gated Autonomy

"Agents wait for human-assigned specs. Level 2 autonomy requires more harness."

This proposal builds the harness. Remote access (Goal 1) is human-triggered — fully compatible with gated autonomy. Autonomous mode (Goal 2) is deferred until harness proves out.

### Decision: Front-load agentic, then fan out

This IS front-loading the agentic foundation. The workspace-aware session is the specific thing WS1 needs.

---

## Critical Analysis

### Is this the RIGHT thing to build?

YES. The only-begotten study test proved brain exec works but the quality gap vs VS Code study agent is obvious. The gap is entirely caused by missing workspace context — not missing tools or model capability.

### What's the simplest version?

Approach B (Squad-style: read files, embed in SystemMessage) is simpler and more predictable than Approach A (SDK fields). We know exactly what goes into the prompt.

BUT: Approach A lets the SDK handle skill loading, which VS Code does natively. If SkillDirectories works, skills are loaded on-demand (same as VS Code) rather than ALL upfront in the system message. That's better for context budget.

**Recommended:** Hybrid. Use SkillDirectories for skills (let SDK manage). Read agent definitions ourselves and set SystemMessage (like Squad does). This avoids depending on SDK's CustomAgents routing behavior which we haven't tested.

### What gets WORSE?

- System message gets longer (agent prompts are 100-300 lines)
- Token usage increases (more context per session)
- More `.github/` files to parse at startup

### Mosiah 4:27 check

This is a SMALL implementation. Agent file parsing + SystemMessage composition + MCP auto-discovery expansion. Estimated 100-200 lines of Go. Not a new project — it's extending createSession() in the existing agent.go.

### Unknown: Does SkillDirectories actually work in the Go SDK?

The field exists. It may or may not be fully wired. Squad doesn't use it. Need to test.

### Unknown: Does the CLI's existing agent support handle agents within the session?

The Go SDK SessionConfig doesn't have an explicit `Agent` field like the TypeScript SDK. Need to verify if `CustomAgents` combined with something else triggers agent selection.

**Fallback if SDK native fields don't work:** Do exactly what Squad does — read everything, embed in system message. This is guaranteed to work because it's just prompt engineering.
