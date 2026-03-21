# Brain Workspace-Aware Agent Sessions

**Binding problem:** brain.exe's agent sessions operate as a generic "development agent" — no study workflow, no scratch files, no critical analysis, no Webster word study, no source verification, no writing voice. The [only-begotten study test](../../scratch/brain-workspace-aware/main.md) proved the agent works mechanically but misses everything that makes our studies *ours*. This is entirely because brain.exe doesn't load the workspace's agent definitions (`.github/agents/`), skills (`.github/skills/`), or instructions (`.github/copilot-instructions.md`) into the Copilot SDK session.

**Created:** 2026-03-20
**Research:** [.spec/scratch/brain-workspace-aware/main.md](../../scratch/brain-workspace-aware/main.md)
**Affects:** WS1 Phase 2→3 (agentic foundation), WS3 (study quality)
**Status:** Draft

---

## 1. Problem Statement

The Copilot SDK Go v0.1.32 `SessionConfig` already supports:

| Field | Type | Purpose | brain.exe uses? |
|-------|------|---------|:---:|
| `CustomAgents` | `[]CustomAgentConfig` | Named agents with prompts, tools, MCP servers | NO |
| `SkillDirectories` | `[]string` | Load skills from directories | NO |
| `SystemMessage` | `*SystemMessageConfig` | System prompt with Mode (append/replace) | Partial — content only, no mode, hardcoded text |
| `InfiniteSessions` | `*InfiniteSessionConfig` | Context compaction for long sessions | NO |

VS Code Copilot Chat reads `.github/` files and maps them into these fields. brain.exe doesn't. The result: when I ask the brain to do a study, I get a competent but soulless response because none of our workflow, voice, conventions, or tools-by-mode reach the session.

### What This Costs

- **Study quality gap.** The only-begotten test produced 205 lines — good content, but no scratch file, no phased workflow, no Webster, no critical analysis, no cross-study connections.
- **Duplicated prompting.** Every `brain exec` call requires the caller to re-explain conventions already defined in `.github/`.
- **Agent modes unusable.** 11 agent definitions exist. brain.exe can't activate any of them.
- **Skills unreachable.** 13 skills (source-verification, scripture-linking, webster-analysis, etc.) defined but never loaded.
- **Context wasted.** Long study sessions hit the context wall because InfiniteSessions isn't enabled.

### Success Criteria

1. `brain exec --agent study --prompt "Study D&C 93:36"` produces a study that follows the 7-phase workflow, creates a scratch file, uses Webster, and applies the writing voice.
2. `brain exec --prompt "Add markdown_link to GetResponse"` (no `--agent` flag) loads `copilot-instructions.md` as the base system message, same as VS Code default mode.
3. Skills from `.github/skills/` are available to all sessions via `SkillDirectories`.
4. InfiniteSessions enabled — studies that run 5+ minutes don't hit context walls.
5. All 6 MCP servers discovered (not just 3).

---

## 2. Constraints & Boundaries

**In scope:**
- Parse `.github/agents/*.agent.md` → `CustomAgentConfig`
- Point `SkillDirectories` at `.github/skills/`
- Load `copilot-instructions.md` as base system message
- Add `--agent <name>` flag to `brain exec`
- Enable `InfiniteSessions`
- Expand MCP auto-discovery (add becoming, byu-citations, search-mcp, yt-mcp)

**Out of scope (deferred to WS1 Phase 3):**
- Multi-agent routing (coordinator pattern from Squad)
- Hook-based governance (A3)
- Reviewer lockout with model escalation (A4)
- Cost tracking per session (A6)
- Handoff processing (agent→agent transfers)
- YAML frontmatter `tools:` field filtering (VS Code uses this for tool whitelisting — we'll pass nil/all for now)

**Conventions:**
- Go. Same packages (`internal/ai`, `internal/config`).
- No new dependencies beyond what's in go.mod (YAML frontmatter parsing can use stdlib strings or a light YAML parser if already present).
- Agent file format: mixed — some use `---` YAML fences, others use ` ```chatagent ` code fences. Parser must handle both.

---

## 3. Prior Art & Related Work

### What Exists

| Source | Relevance |
|--------|-----------|
| [Squad](../../proposals/squad-learnings.md) | Multi-agent runtime. Reads `.squad/` files and embeds them in SystemMessage. Does NOT use SDK's `CustomAgents` or `SkillDirectories`. Uses LLM-based routing. |
| [Overview plan](../../proposals/overview/main.md) WS1 Phase 3 | Multi-agent routing spec — depends on workspace-aware sessions existing first. |
| VS Code Copilot Chat | The reference implementation. Reads `.github/` files → SDK fields. |
| Current `createSession()` | [agent.go line 258](../../../scripts/brain/internal/ai/agent.go) — sets 6 fields. Needs 3 more. |
| Current `discoverMCPServers()` | [config.go line 402](../../../scripts/brain/internal/config/config.go) — hardcoded list of 3 servers. Needs 4 more. |
| Current system prompt | [main.go line 477](../../../scripts/brain/cmd/brain/main.go) — ~10 lines of generic instructions. Needs to be replaced with copilot-instructions.md content. |

### Design Choice: SDK Fields vs. System Message Embedding

Two approaches observed in the wild:

**Approach A: SDK-native fields** — Set `CustomAgents`, `SkillDirectories` on `SessionConfig`. Let the SDK handle loading, routing, and context management. This is what VS Code does.

**Approach B: System message embedding** — Read files ourselves, compose them into `SystemMessage.Content`. This is what Squad does.

**Recommendation: Hybrid.**
- `SkillDirectories` → SDK-native. Skills are designed for on-demand loading (the SDK knows when to inject a skill based on context). Stuffing all 13 skills into the system message wastes tokens.
- `CustomAgents` → SDK-native. Register all 11 agents. The agent prompt lives in the `Prompt` field.
- `SystemMessage` → Read `copilot-instructions.md` ourselves and set as base message with `Mode: "append"`. This ensures our project conventions are always present, supplementing whatever the SDK adds.
- Agent selection → When `--agent study` is specified, compose the agent's prompt content into the SystemMessage instead of relying on an `Agent` field (which doesn't exist in the Go SDK). Alternatively, register via `CustomAgents` and test whether the SDK routes to the named agent when referenced in conversation.

**Why not pure embedding?** SkillDirectories likely enables the SDK to load skills on-demand based on conversation context, rather than stuffing all skills into every session's system message. That's a meaningful context budget optimization for long studies.

**Fallback:** If `SkillDirectories` or `CustomAgents` don't work as expected in the Go SDK (they may be wired for VS Code's agent host, not standalone SDK usage), fall back to embedding everything in SystemMessage. The parsing code is the same either way — only the destination changes.

---

## 4. Proposed Approach

### 4.1 Agent File Parser

Parse `.github/agents/*.agent.md` into `CustomAgentConfig` structs.

**Input formats (both observed in our agents):**

Format 1 — standard YAML frontmatter:
```
---
description: '...'
tools: [...]
handoffs:
  - label: ...
---

# Agent Title
<body>
```

Format 2 — chatagent code fence:
```
```chatagent
---
description: '...'
tools: [...]
handoffs:
  - label: ...
---

# Agent Title
<body>
``` ← closing fence
```

**Output per agent:**
```go
copilot.CustomAgentConfig{
    Name:        "study",                    // from filename: study.agent.md → "study"
    DisplayName: "Scripture Study Agent",    // from first # heading or description
    Description: "Scripture study agent...", // from YAML description field
    Prompt:      "<full body content>",      // everything after frontmatter
}
```

**Note on MCP servers per agent:** Agent YAML frontmatter has a `tools:` list (e.g. `'gospel/*'`, `'webster/*'`). These map to MCP server names. For Phase 1, we register ALL discovered MCP servers on every agent (same as current behavior). Per-agent MCP server filtering is a Phase 3 optimization.

### 4.2 System Message Composition

```go
// Load copilot-instructions.md
baseInstructions, _ := os.ReadFile(filepath.Join(workspaceDir, ".github", "copilot-instructions.md"))

cfg.SystemMessage = &copilot.SystemMessageConfig{
    Mode:    "append",
    Content: string(baseInstructions),
}
```

When `--agent <name>` is specified and the SDK's `CustomAgents` routing doesn't activate the agent automatically, compose the agent's full prompt into the SystemMessage:

```go
if agentName != "" {
    agentPrompt := loadAgentPrompt(workspaceDir, agentName)
    cfg.SystemMessage = &copilot.SystemMessageConfig{
        Mode:    "replace",
        Content: string(baseInstructions) + "\n\n" + agentPrompt,
    }
}
```

### 4.3 Skill Directory Registration

```go
skillDir := filepath.Join(workspaceDir, ".github", "skills")
if info, err := os.Stat(skillDir); err == nil && info.IsDir() {
    cfg.SkillDirectories = []string{skillDir}
}
```

### 4.4 InfiniteSessions

```go
cfg.InfiniteSessions = &copilot.InfiniteSessionConfig{
    Enabled: boolPtr(true),
}
```

Default thresholds (80% background compaction, 95% buffer exhaustion) are reasonable for studies.

### 4.5 Expanded MCP Discovery

Add 4 more servers to the `specs` list in `discoverMCPServers()`:

```go
specs := []serverSpec{
    {"gospel-mcp",   "gospel-mcp",   []string{"serve"}},
    {"gospel-vec",   "gospel-vec",   []string{"mcp"}},
    {"webster-mcp",  "webster-mcp",  []string{"serve"}},
    {"becoming",     "becoming",     []string{"mcp"}},       // NEW
    {"byu-citations","byu-citations", []string{"serve"}},    // NEW
    {"search-mcp",   "search-mcp",   []string{"serve"}},    // NEW
    {"yt-mcp",       "yt-mcp",       []string{"serve"}},    // NEW
}
```

(Verify each server's MCP subcommand name before implementation.)

### 4.6 CLI Changes

Extend `brain exec` argument parsing:

```
brain exec --agent study --prompt "Study D&C 93:36"
brain exec --agent dev test-spec.md
brain exec --prompt "do something"          # no agent = base instructions only
```

Parse `--agent <name>` before the existing `--prompt` / file logic.

---

## 5. Phased Delivery

### Phase 1: Load What Exists (~100-150 lines of Go)

1. **Agent file parser** — new function in `internal/config/` that reads `.github/agents/*.agent.md`, extracts YAML frontmatter + body, returns `[]copilot.CustomAgentConfig`
2. **System message from file** — replace hardcoded system prompt with `copilot-instructions.md` content
3. **`--agent` flag** — parse in `runExec()`, compose agent prompt into SystemMessage
4. **SkillDirectories** — point at `.github/skills/`
5. **InfiniteSessions** — enable with defaults
6. **Expanded MCP discovery** — add 4 servers to the list

**Verify:** `brain exec --agent study --prompt "Study D&C 93:36"` produces a study with scratch file, phased workflow, Webster usage.

### Phase 2: Validate SDK Fields (1 session)

Test whether `CustomAgents` and `SkillDirectories` actually work in standalone Go SDK sessions (not just VS Code):
- Do skills get loaded on-demand?
- Does the SDK reference CustomAgents when the user mentions `@study`?
- If not: fall back to SystemMessage embedding for agents, keep SkillDirectories if it works.

**Verify:** Compare token usage and behavior between SDK-native vs. embedded approaches.

### Phase 3: Per-Agent Tool Filtering (deferred — WS1 Phase 3)

Parse the `tools:` YAML field and map tool patterns to MCP server names. Only register matching MCP servers per agent session. This is an optimization, not a correctness requirement.

---

## 6. Verification Criteria

| # | Criterion | How to test |
|---|-----------|------------|
| 1 | Study agent follows 7-phase workflow | `brain exec --agent study --prompt "Study Mosiah 4:27"` → check output for phases, scratch file creation, Webster usage |
| 2 | Base instructions loaded without `--agent` | `brain exec --prompt "What files are in study/"` → agent should follow source-verification conventions |
| 3 | Skills available | Study output should reference source-verification behavior (read before quoting) |
| 4 | InfiniteSessions prevents context overflow | Run a long study (5+ min) without hitting "context too large" errors |
| 5 | All 7 MCP servers discovered | Log output shows 7 registered servers on startup |
| 6 | Plan/dev/journal agents work | `brain exec --agent plan --prompt "..."` uses plan agent's prompt |

---

## 7. Compatibility with Existing Plans

### WS1 Phase 3 (Multi-Agent Routing)

This proposal **enables** Phase 3. Phase 3 needs workspace-aware sessions to exist before it can route between them. The `CustomAgents` registration we do here is the data structure Phase 3's routing logic will consume.

No conflicts. This is a prerequisite.

### Squad Adoption Items

| Item | Impact |
|------|--------|
| A1: Decisions file | No impact (already done) |
| A2: Routing table | Enabled — agents are registered, routing can reference them |
| A3: Hook governance | No impact — hooks not changed. Phase 3 adds them |
| A4: Reviewer lockout | No impact — Phase 3 concern |
| A5: Model tiering | Compatible — CustomAgentConfig supports per-agent model if needed later |
| A6: Cost tracking | No impact — orthogonal |

### Gated Autonomy Decision

This proposal respects gated autonomy:
- **Goal 1 (Remote Access):** User triggers `brain exec --agent study`. Human-supervised. Fully gated.
- **Goal 2 (Autonomous Agents):** Deferred to Phase 3. This proposal builds the session infrastructure but doesn't add autonomous triggering.

### Two-Goal Compatibility

The two goals — "brain as remote workspace access" and "autonomous agents like Squad" — are **the same infrastructure at different activation levels:**

| Dimension | Remote Access (Goal 1) | Autonomous (Goal 2) |
|-----------|----------------------|---------------------|
| Session creation | On `brain exec --agent` | On task assignment by coordinator |
| Agent selection | User specifies `--agent` | Coordinator routes via LLM |
| Session lifecycle | One per invocation | Pool + reuse |
| Governance | User reviews output | Hook-enforced review gates |
| Agents | Same `.github/agents/` | Same `.github/agents/` |
| Skills | Same `.github/skills/` | Same `.github/skills/` |
| MCP servers | Same auto-discovered set | Same auto-discovered set |

This proposal delivers Goal 1. Goal 2 adds routing + governance on top of the same foundation. No contradiction.

---

## 8. Costs & Risks

| Cost/Risk | Assessment |
|-----------|-----------|
| **Token usage increase** | System message grows from ~10 lines to 200-400 lines (copilot-instructions.md + agent prompt). Meaningful but acceptable — these are instructions that directly improve output quality. |
| **Agent parsing complexity** | Two frontmatter formats (YAML fences vs chatagent code fences). Manageable with ~30 lines of parsing code. |
| **SkillDirectories may not work** | The Go SDK field exists but may not be fully wired for standalone usage. Fallback: embed skill text in system message. |
| **CustomAgents may not route** | No `Agent` field in Go SDK to select an agent by name. We compose the agent prompt directly into SystemMessage as the primary approach. |
| **Additional MCP servers = more processes** | 7 servers vs 3. Each is a lightweight Go binary started on-demand by the SDK. Negligible resource impact. |
| **Implementation scope** | ~150 lines of Go across 3 files. Small. No new packages, no new architecture. |

---

## 9. Creation Cycle Review

| Step | This Proposal |
|------|---------------|
| **Intent** | brain.exe should work like sitting at VS Code — same agents, same skills, same conventions. |
| **Covenant** | Same `.github/` files are the source of truth for both VS Code and brain.exe. No drift. |
| **Stewardship** | dev agent executes. `internal/ai/` and `internal/config/` are the affected packages. |
| **Spiritual Creation** | This proposal. The scratch file research. The compatibility analysis. |
| **Line upon Line** | Phase 1 is self-contained (load files, compose session). Phase 2 validates. Phase 3 extends. |
| **Physical Creation** | ~150 lines of Go. 1 session to implement and test. |
| **Review** | 6 verification criteria. The study quality test is the real measure. |
| **Atonement** | If SDK fields don't work → fallback to system message embedding (same parsing, different destination). 30-minute pivot. |
| **Sabbath** | This is a small, focused change. Study afterwards. |
| **Consecration** | Every future brain session benefits. Studies, plans, lessons — all get the full workspace context. |
| **Zion** | One workspace, one set of conventions, multiple entry points (VS Code, brain.exe, brain-app). Same experience everywhere. |

---

## 10. Recommendation

**Build.** This is a small, high-impact change that directly addresses the study quality gap identified in the only-begotten test. It's a prerequisite for everything in WS1 Phase 3. No conflicts with existing plans. ~150 lines of Go, 1 session to implement.

**Phase 1 is the deliverable.** Phases 2-3 are validation and optimization.

**Executing agent:** dev agent.

**Next action:** `brain exec --agent dev` with this proposal as the spec. Or hand to the dev agent in VS Code.
