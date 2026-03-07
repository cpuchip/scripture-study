# Brain — GitHub Copilot SDK Integration

*Created: July 2025*
*Updated: March 2026 — Rewritten after researching the actual Copilot SDK*
*Status: Draft*
*Depends on: Copilot CLI installed on host, GitHub Copilot subscription*

---

## The Real Shape of This Feature

The GitHub Copilot SDK (`github.com/github/copilot-sdk/go`) is a Go package that wraps the Copilot CLI into a programmable agent runtime. It's not a VS Code extension API — it's a standalone library that brain.exe can import directly. The SDK:

- Manages the Copilot CLI process lifecycle (start, stop, auto-restart)
- Communicates via JSON-RPC (stdio or TCP)
- Supports **model selection** per session (gpt-5, claude-sonnet-4.5, claude-opus-4, haiku, etc.)
- Supports **MCP servers** — same MCP config as VS Code / Copilot CLI
- Supports **custom tools** with type-safe Go handlers
- Has **session hooks** for permission control (allow/deny/modify tool calls)
- Has **workspace paths** where sessions can read `.github/copilot-instructions.md`, agents, and skills
- Supports **BYOK** — can point to LM Studio, Ollama, or any OpenAI-compatible endpoint
- Supports **streaming**, **infinite sessions** (auto context compaction), **image attachments**
- Can **embed the CLI binary** in the brain.exe distribution

This means brain.exe can offer the **same agentic experience as a VS Code Copilot session** — with the same instructions, agents, skills, and MCP tools — triggered from your phone.

---

## Architecture

```
┌─────────┐     ┌───────────────────────────────────────────────────┐
│  Phone   │────►│                  brain.exe                        │
│ (Flutter)│◄────│                                                   │
└─────────┘     │  ┌─────────────┐     ┌──────────────────────┐    │
                │  │  Classifier  │     │  Copilot SDK Client  │    │
                │  │ (ministral)  │     │  (Go SDK)            │    │
                │  │              │     │                      │    │
                │  │ raw text ──► │     │  ┌─ claude-sonnet ─┐ │    │
                │  │ if command ──┼────►│  │  claude-opus    │ │    │
                │  │ else normal  │     │  │  claude-haiku   │ │    │
                │  └─────────────┘     │  │  gpt-5          │ │    │
                │                      │  └──────────────────┘ │    │
                │                      │         │             │    │
                │                      │    Copilot CLI        │    │
                │                      │    (managed process)  │    │
                │                      │         │             │    │
                │                      │    MCP Servers        │    │
                │                      │    • gospel-vec       │    │
                │                      │    • gospel-mcp       │    │
                │                      │    • webster-mcp      │    │
                │                      │    • becoming-mcp     │    │
                │                      │    • github-mcp       │    │
                │                      │                      │    │
                │                      │    Workspace          │    │
                │                      │    • instructions.md  │    │
                │                      │    • agents/          │    │
                │                      │    • skills/          │    │
                │                      └──────────────────────┘    │
                └───────────────────────────────────────────────────┘
```

### Two-tier classification flow

1. **Ministral** (local, fast, free) classifies incoming text as it does today
2. If the input is detected as a **command** (or user explicitly selects "copilot mode"), ministral routes it to the Copilot SDK instead of normal entry creation
3. The Copilot SDK session uses **cloud models** (haiku for quick answers, sonnet for complex tasks, opus for deep work) with full tool access

This keeps the fast/free path for normal captures and reserves the premium path for agentic work.

---

## Inheriting the Workspace

The key insight: the Copilot CLI reads `.github/copilot-instructions.md`, `.github/agents/`, and `.github/skills/` from its **working directory**. brain.exe already lives inside the scripture-study repo. The SDK's `ClientOptions.Cwd` sets the working directory:

```go
client := copilot.NewClient(&copilot.ClientOptions{
    Cwd:      "/path/to/workspace",  // Where to find .github/ config
    LogLevel: "error",
})
```

### Isolated Workspace (Phase 1 — Trust Building)

Start with a **dedicated, sandboxed workspace** rather than pointing at the full scripture-study repo:

```
brain-data/
├── brain.db
├── attachments/
└── copilot-workspace/          ← isolated workspace
    ├── .github/
    │   ├── copilot-instructions.md   ← curated subset of instructions
    │   ├── agents/
    │   │   └── study.agent.md        ← study agent only (read-only operations)
    │   └── skills/
    │       ├── scripture-linking/
    │       └── webster-analysis/
    └── context/                      ← symlinked or copied study documents
        └── (read-only access to gospel-library, study/, etc.)
```

**Why isolated first:**
- **Trust building** — "line upon line, precept upon precept" (2 Nephi 28:30). We don't hand over the keys to everything on day one.
- **Progressive context** — start with read-only scripture tools, earn write access over time
- **Stewardship** — the agent operates within a defined stewardship. It can search scriptures and answer questions. It cannot edit code, push commits, or modify the becoming app database — not yet.
- **Safety** — if the agent hallucinates a file edit, it operates in a sandbox where damage is contained

### Full Workspace (Phase 3+ — After Trust is Established)

Once the isolated workspace proves reliable:
- Point `Cwd` at the actual scripture-study repo
- Grant access to additional agents (dev, review, lesson, etc.)
- Enable write operations (create entries, update files) with hook-based approval

---

## Copilot SDK Integration in brain.exe

### Dependencies

```
go get github.com/github/copilot-sdk/go
```

### Client Initialization

```go
// internal/copilot/agent.go

type Agent struct {
    client  *copilot.Client
    session *copilot.Session
}

func NewAgent(workspacePath string) (*Agent, error) {
    client := copilot.NewClient(&copilot.ClientOptions{
        Cwd:      workspacePath,
        LogLevel: "error",
    })

    if err := client.Start(context.Background()); err != nil {
        return nil, fmt.Errorf("starting copilot CLI: %w", err)
    }

    return &Agent{client: client}, nil
}
```

### Session Creation with Model Selection

```go
func (a *Agent) CreateSession(model string) error {
    session, err := a.client.CreateSession(context.Background(), &copilot.SessionConfig{
        Model:     model,  // "claude-sonnet-4.5", "claude-haiku-3.5", "claude-opus-4"
        Streaming: true,
        Tools:     a.brainTools(),  // Custom brain entry tools

        // Permission control — the stewardship layer
        Hooks: &copilot.SessionHooks{
            OnPreToolUse: func(input copilot.PreToolUseHookInput, inv copilot.HookInvocation) (*copilot.PreToolUseHookOutput, error) {
                // Phase 1: only allow read-only tools
                if isWriteOperation(input.ToolName) {
                    return &copilot.PreToolUseHookOutput{
                        PermissionDecision: "deny",
                    }, nil
                }
                return &copilot.PreToolUseHookOutput{
                    PermissionDecision: "allow",
                }, nil
            },
        },
    })
    if err != nil {
        return err
    }
    a.session = session
    return nil
}
```

### Custom Brain Tools

Expose brain.exe's data as tools the agent can use:

```go
func (a *Agent) brainTools() []copilot.Tool {
    searchEntries := copilot.DefineTool("brain_search", "Search brain entries by text query",
        func(params struct {
            Query string `json:"query" jsonschema:"Search query text"`
            Limit int    `json:"limit,omitempty" jsonschema:"Max results (default 10)"`
        }, inv copilot.ToolInvocation) (any, error) {
            results, err := a.store.SearchText(params.Query, params.Limit)
            if err != nil {
                return nil, err
            }
            return results, nil
        })

    getEntry := copilot.DefineTool("brain_get", "Get a specific brain entry by ID",
        func(params struct {
            ID string `json:"id" jsonschema:"Entry ID"`
        }, inv copilot.ToolInvocation) (any, error) {
            return a.store.GetEntry(params.ID)
        })

    listRecent := copilot.DefineTool("brain_recent", "List recent brain entries",
        func(params struct {
            Limit    int    `json:"limit,omitempty" jsonschema:"Max results (default 20)"`
            Category string `json:"category,omitempty" jsonschema:"Filter by category"`
        }, inv copilot.ToolInvocation) (any, error) {
            if params.Category != "" {
                return a.store.ListCategory(params.Category)
            }
            limit := params.Limit
            if limit == 0 {
                limit = 20
            }
            return a.store.ListAll(limit, 0)
        })

    return []copilot.Tool{searchEntries, getEntry, listRecent}
}
```

### Sending Messages

```go
func (a *Agent) Chat(ctx context.Context, message string, onEvent func(event copilot.SessionEvent)) error {
    done := make(chan struct{})

    a.session.On(func(event copilot.SessionEvent) {
        onEvent(event)  // Forward events to caller (for streaming to phone)
        if event.Type == "session.idle" {
            close(done)
        }
    })

    _, err := a.session.Send(ctx, copilot.MessageOptions{
        Prompt: message,
    })
    if err != nil {
        return err
    }

    <-done
    return nil
}
```

---

## Classifier Integration: Detecting Commands

### Option A: New "command" category

Add `command` to the classifier's category list. Update the system prompt:

```
- command: A request directed at an AI agent — asking for information, analysis,
  code changes, scripture research, or any task that requires tools and reasoning.
  NOT a thought to capture — an instruction to execute.
```

When ministral classifies input as `command`, brain.exe routes to the Copilot SDK instead of creating an entry.

### Option B: Explicit copilot mode (recommended for Phase 1)

Don't change classification at all. Instead:
- Add a **"Copilot" mode toggle** in the Flutter app (alongside the existing mic/text input)
- When copilot mode is active, input goes directly to the Copilot SDK
- Normal mode continues to classify → create entry as today

This is safer for Phase 1 — no risk of mistakenly routing a thought capture to an agent.

### Model Selection Heuristic

```
Quick questions, single-tool lookups  → claude-haiku (fast, cheap)
Multi-step research, cross-referencing → claude-sonnet (balanced)
Deep analysis, code generation         → claude-opus (thorough)
```

For Phase 1: default to sonnet, let the user override in settings.

---

## Use Cases (Ordered by Phase)

### Phase 1: Study Mode (read-only, isolated workspace)

```
User (phone): "What does Paul mean by 'the natural man' in 1 Corinthians 2:14?"

Copilot SDK → claude-sonnet → uses:
  • gospel-vec (semantic search for "natural man")
  • gospel-mcp (read the actual verses)
  • webster-mcp (look up "natural" in Webster 1828)
  • Workspace skills: scripture-linking, webster-analysis, source-verification

→ Returns a study-quality answer with citations and cross-references
```

This works because:
- All MCP servers already exist and run on the same machine
- The SDK inherits the workspace's copilot-instructions.md (read before quoting, link everything, etc.)
- The study agent's instructions guide the response format

### Phase 2: Brain Intelligence (read entries, still no writes)

```
User: "What ideas have I captured about faith this month?"
User: "Summarize my overdue actions"
User: "Find entries related to my conversation with Josh"
```

Uses the custom `brain_search`, `brain_recent`, `brain_get` tools plus the cloud model's reasoning.

### Phase 3: Write Operations (with approval hooks)

```
User: "Create an action item: call Josh about the project next Tuesday"
User: "Move all my inbox entries about scripture to the study category"
```

The `OnPreToolUse` hook can:
- Allow automatically for low-risk writes (create entry)
- Prompt user confirmation for bulk operations
- Deny destructive operations (delete, reclassify all)

### Phase 4: Code/GitHub Integration

```
User: "What PRs are open on scripture-study?"
User: "Create an issue for the Samsung widget bug"
User: "Summarize what changed since my last commit"
```

Uses GitHub MCP server. Still within the isolated workspace stewardship.

### Phase 5: Full Workspace Access

Point `Cwd` at the real scripture-study repo. The agent can now:
- Use all agents (dev, review, lesson, study, etc.)
- Read study documents, lessons, plans
- Propose code changes (with approval)
- Run the dev agent for bug fixes

---

## Session Hooks: The Stewardship Layer

The SDK's hook system is the trust mechanism. Every tool call passes through `OnPreToolUse`:

```go
OnPreToolUse: func(input copilot.PreToolUseHookInput, inv copilot.HookInvocation) (*copilot.PreToolUseHookOutput, error) {
    switch trustLevel {
    case TrustReadOnly:
        // Phase 1: deny all writes
        if isWriteOperation(input.ToolName) {
            return deny("Read-only stewardship — write operations not permitted")
        }

    case TrustCreateOnly:
        // Phase 2: allow creating entries, deny edits/deletes
        if isDeleteOperation(input.ToolName) {
            return deny("Delete operations require full trust")
        }

    case TrustFull:
        // Phase 3+: allow everything, log for review
        log.Printf("AGENT: %s(%v)", input.ToolName, input.ToolArgs)
    }

    return allow()
}
```

Trust levels are configured in brain.exe settings, not controlled by the agent. The human decides when to elevate.

---

## Flutter UI

### Copilot Mode Toggle

```
┌──────────────────────────────────┐
│         [Brain] [Copilot]        │  ← mode toggle
│──────────────────────────────────│
│                                  │
│ 🧑 What does Paul mean by       │
│    "the natural man"?            │
│                                  │
│ 🤖 Paul's use of "the natural   │
│    man" in 1 Cor 2:14...         │
│                                  │
│    🔧 gospel-vec: 3 results      │  ← tool call indicator
│    🔧 webster: "natural"         │
│                                  │
│    📖 Mosiah 3:19 expands...     │
│                                  │
│──────────────────────────────────│
│ [Ask anything...           ] 🎤  │
└──────────────────────────────────┘
```

- **Brain mode**: existing capture → classify → entry flow
- **Copilot mode**: input goes to Copilot SDK, streaming response displayed
- Streaming text display with markdown rendering
- Tool call indicators (which tools the agent used)
- Voice input via existing STT
- "Save as entry" button on responses worth keeping
- Model selector in settings (haiku / sonnet / opus)

### Conversation Continuity

The SDK supports `ResumeSession` — the phone can continue a multi-turn conversation:

```go
session, err := client.ResumeSession(ctx, "session-abc", &copilot.ResumeSessionConfig{
    Tools: agent.brainTools(),
})
```

The SDK also has `InfiniteSessions` with automatic context compaction — long conversations don't overflow.

---

## Implementation Phases

### Phase 1: Isolated Workspace + Read-Only Study

1. `go get github.com/github/copilot-sdk/go`
2. Create `internal/copilot/agent.go` — client init, session management
3. Create isolated workspace directory with curated `.github/` config
4. Add brain tools (search, get, recent) — read-only
5. `OnPreToolUse` hook: deny all write operations
6. `POST /api/copilot/chat` endpoint with SSE streaming
7. Test with study questions against gospel-vec/gospel-mcp/webster-mcp
8. Verify that `.github/copilot-instructions.md` and skills are picked up

### Phase 2: Flutter Chat UI

1. Chat screen with Brain/Copilot mode toggle
2. SSE streaming text display with markdown rendering
3. Tool call status indicators
4. Voice input integration
5. "Save as entry" action on responses
6. Model selection in settings (haiku/sonnet/opus)

### Phase 3: Write Operations + Brain Intelligence

1. Add write tools (create entry, update entry, reclassify)
2. Introduce trust levels in config
3. `OnPreToolUse` enforces trust level
4. "Chat about this entry" button on entry detail
5. Conversation history via `ResumeSession`

### Phase 4: Relay Support

1. Copilot chat relay through ibeco.me
2. SSE passthrough or WebSocket-based streaming
3. Session persistence across connections

### Phase 5: Full Workspace Access

1. Point `Cwd` at scripture-study repo
2. Enable additional agents (dev, review, lesson)
3. GitHub MCP server integration
4. Code operations with approval workflow

---

## Requirements & Costs

### Prerequisites

- **Copilot CLI** installed on the machine running brain.exe
- **GitHub Copilot subscription** (Free tier has limited premium requests; Pro/Pro+ for heavier use)
- Or **BYOK** — the SDK supports bringing your own API keys (OpenAI, Anthropic, Azure) without GitHub auth

### Billing

Each prompt counts toward the Copilot premium request quota. The pricing model: quick haiku calls are cheap, opus calls are premium. Classification still uses local ministral (free). Only agentic commands hit the Copilot quota.

### Embedding the CLI

The SDK supports bundling the Copilot CLI binary:

```bash
go get -tool github.com/github/copilot-sdk/go/cmd/bundler
go tool bundler  # Run before go build
```

This embeds the CLI in brain.exe — no separate install needed on the target machine.

---

## Open Questions

- **MCP server transport:** Copilot CLI can use MCP servers. Do our Go MCP servers need to run as separate processes (stdio) or can they be wired in differently? The Copilot CLI typically reads MCP config from `.github/copilot-instructions.md` or VS Code settings. Need to test how the SDK discovers and connects to MCP servers.
- **Command detection accuracy:** If we go with Option A (classifier detects commands), how accurate will ministral be at distinguishing "capture this thought" from "do this for me"? Testing needed — "remind me to call Josh" is a capture, "call Josh and schedule a meeting" is a command. Subtle.
- **Session lifetime:** How long do Copilot SDK sessions live? Can we keep a session open for the phone's entire connection, or create per-request sessions? The `InfiniteSessions` feature suggests long-lived is possible.
- **Offline behavior:** When the machine is running but internet is down (no Copilot access), should we fall back to LM Studio for a degraded agent experience? Or just error cleanly?
- **BYOK vs. Copilot auth:** For initial development, BYOK with an Anthropic key might be simpler than setting up Copilot CLI auth. Get the plumbing working first, switch to Copilot auth later.
- **Cost monitoring:** Need a way to track premium request usage from the phone. Display in settings: "X copilot requests today / this month."
