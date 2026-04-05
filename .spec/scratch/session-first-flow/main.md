# Session-First Flow — Research & Development Notes

*Created: Apr 5, 2026*
*Status: Exploring — not yet specced*

## The Idea (Michael's Framing)

Currently: capture text → classify (strict JSON) → assign to project → nudge agent reviews → user advances through stages manually → each stage is a discrete API call.

Proposed: What if every entry started as a **session** instead? Raptor (brain's AI) would open a session, make MCP/CLI calls to research, edit, and move the entry along. The session shows up in VS Code like Copilot sessions already do. All interactions with that entry happen consistently within that thread. When done, archive the session so it drops off the sidebar.

Key insight: "Maybe we can drop some of the formality like having the strict JSON output step and just have it call an MCP or CLI to move the work along."

## What We Discovered — Current Architecture

### How Sessions Are Created Now

Brain uses the **Copilot SDK** (Go) to create sessions via `client.CreateSession()`. Each agent gets a session on its first `Ask()` call. Sessions are:
- **Agent-scoped, not entry-scoped** — one "study" session handles all study entries
- **Ephemeral** — lost on brain restart (no serialization)
- **Invisible to VS Code** — they DO show up in the sidebar (confirmed from screenshots)
- **Tool-enabled** — 7 MCP servers auto-discovered and registered

Wait — correction from the screenshots: Sessions created by brain's Copilot SDK **DO** show up in VS Code's sessions sidebar. The "Entry: Widget size and mic functionality in..." session visible in the screenshot was created by brain's nudge review agent. This is because the Copilot SDK shares session state with VS Code's session manager.

### How Nudge Sessions Work

In `pipeline/review.go`, the nudge agent:
1. Creates a new `ai.NewAgent(client, agentCfg)` — NOT from the pool
2. Calls `agent.Ask(ctx, prompt)` — which calls `createSession()` internally
3. Each nudge creates a **separate session** → separate sidebar entry
4. The response is stored in `session_messages` table
5. Entry gets `agent_route = "review"`, `route_status = "your_turn"`

So each nudged entry = one VS Code session. The session lives in VS Code's sidebar.

### The Two Session Systems

| System | Storage | Visible in VS Code | Persistent | Multi-turn |
|--------|---------|-------------------|------------|------------|
| Copilot SDK session | SDK internal + VS Code | YES (sidebar) | NO (lost on restart) | YES (within one brain run) |
| session_messages table | SQLite | NO (brain web UI only) | YES | YES |

The gap: brain stores conversation in SQLite, but the *actual* SDK session is ephemeral. The VS Code session shows the initial prompt/response, but NOT the subsequent replies stored in SQLite.

### What Copilot CLI Adds

From the VS Code screenshot, Michael is using Copilot CLI (visible at bottom: "Copilot CLI" mode). When Copilot CLI runs tasks:
- Sessions appear in VS Code sidebar with full interactivity
- User can type in the session directly from VS Code
- Sessions persist across CLI invocations (VS Code manages them)
- CLI sessions have MCP tool access

## The Architecture Shift

### Current: Fire-and-Forget Agent Calls

```
Brain classifies → routes to agent → agent runs (background) → result stored
→ User sees result in web UI panel → replies go to SQLite
→ Future agent calls don't see prior conversation context (separate sessions)
```

### Proposed: Session = Entry Lifecycle

```
Entry created → Session created → Session IS the entry's lifecycle
→ Raptor (in session) classifies via MCP tool call
→ Raptor researches via MCP tool calls
→ User interacts via VS Code session
→ Raptor advances maturity via MCP tool calls
→ Entry done → Session archived → drops off sidebar
```

### What Changes

1. **No separate classify step with strict JSON.** Raptor calls `brain.classify(entry_id, category, tags)` as an MCP tool from within the session. The classification happens naturally in conversation, not as a forced JSON response.

2. **No separate pipeline stages.** The session IS the progression. Raptor reads `brain.get_entry(id)`, decides what to do, calls `brain.advance(id)` or `brain.set_maturity(id, "researched")`.

3. **Brain MCP server becomes the control plane.** Brain exposes tools like:
   - `brain_get_entry(id)` — read entry state
   - `brain_update_entry(id, fields)` — edit
   - `brain_advance(id)` — trigger next pipeline stage
   - `brain_close(id, reason)` — close with note
   - `brain_list_project(project_id)` — see siblings
   - `brain_add_message(id, content)` — add conversation turn

4. **VS Code becomes the primary interaction surface.** The web dashboard becomes a *monitoring* view, not the primary work surface. Sessions in VS Code are where the work happens.

5. **Session pinning = active work.** Unpin = out of focus. Archive = done.

### What Doesn't Change

- Brain still stores entries in SQLite (source of truth)
- Web dashboard still exists for overview/board view
- MCP servers (gospel, webster, etc.) still work
- Maturity pipeline concept stays

## Open Questions

1. **Can brain register itself as an MCP server within its own SDK sessions?** Circular dependency: brain → SDK → MCP → brain. Might need a separate process or use the HTTP-based MCP transport instead of stdio.

2. **Entry-scoped vs. agent-scoped sessions?** Current model: one session per agent (shared). Proposed: one session per entry. Cost implications — each session has startup cost. Possible middle ground: one session per *project*, handling entries within that project.

3. **SDK session persistence.** Can Copilot SDK sessions be serialized/resumed? If brain restarts, are sessions lost? If yes, we need a recovery path.

4. **Copilot CLI interaction model.** Michael can interact with CLI sessions from VS Code. Does the SDK expose the same capability? Can we create a session in the SDK and have the user interact with it directly in VS Code's chat panel?

5. **How does archiving work?** VS Code sessions can be deleted/hidden. Is there an API for this? Or does the user manually manage their sidebar?

6. **Cost model.** One session per entry means N sessions. Each session might use premium requests for tool calls. Need to understand the billing implications.

7. **Can we just use Copilot CLI instead of the SDK?** If CLI sessions already show up nicely in VS Code, maybe brain should invoke `copilot -p "..."` as a subprocess instead of using the SDK directly. CLI might handle session management better.

## Potential Simplifications

Michael's key insight: "drop some of the formality like having the strict JSON output step." If the agent can call MCP tools, then:
- Classification = `brain.update_entry(id, {category: "ideas", tags: ["brain-app"]})` — no JSON schema needed
- Project assignment = `brain.set_project(id, project_id)` — direct tool call
- Maturity advance = `brain.advance(id)` — tool call, not button click
- Research = agent reads files, calls search tools, writes to scratch file — all within the session
- Close/defer = `brain.close(id, "going a different direction")` — tool call

The formality of the classify step exists because we needed structured output. If the agent can take *actions* via tools, we don't need structured output — we need structured *effects*.

## Connection to Other Workstreams

- **brain-workspace-aware proposal**: Already identifies that skills, agents, and custom agent configs need to be loaded into SDK sessions. Session-first flow builds on this.
- **claude-code-integration**: If Claude Code CLI also creates VS Code sessions, dual-backend becomes more interesting — some entries handled by Copilot, some by Claude.
- **brain-inline-panel**: Stop-gap. If session-first flow succeeds, the web panel becomes monitoring-only and inline reply is less critical there.

## Recommendation

This is worth pursuing but needs more investigation on:
1. Whether SDK sessions can surface interactively in VS Code (not just as sidebar entries)
2. Whether brain can register as its own MCP server within sessions
3. Cost model for per-entry sessions

Start with a **proof of concept**: Create one entry, open one session via SDK, register brain's MCP server, have the agent classify and advance the entry via tool calls. See if the session shows up interactively in VS Code.

If POC succeeds, this becomes a major workstream that could simplify the entire pipeline architecture.
