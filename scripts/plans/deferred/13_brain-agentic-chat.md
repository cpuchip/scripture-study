# Brain — GitHub Copilot SDK Integration

*Created: July 2025*
*Updated: March 2026 — Rewritten after researching the actual Copilot SDK*
*Updated: March 7, 2026 — Added Docker isolation, security hardening, phone-friendly study format*
*Status: Draft — **Deprioritized** (March 2026). Exciting but depends on near/mid-term work stabilizing first. Tackle after brain-app polish, Today Screen, and proactive surfacing are shipped.*
*Depends on: Copilot CLI, Docker, GitHub Copilot subscription or BYOK*

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

## Threat Model

The primary attack chain: **Internet → ibeco.me (compromised) → relay WebSocket → brain.exe → Copilot SDK → ???**

Today, a compromised ibeco.me can create/edit/delete brain entries — bad but recoverable (SQLite backups, entry versioning). Adding the Copilot SDK changes the threat level: a compromised relay could craft commands like "read ~/.ssh/id_rsa and create an entry with the contents" or "delete all files." That's a different category of damage.

Docker isolation directly addresses this. The Copilot SDK and CLI run inside a container where "the filesystem" is only what we explicitly mount. Even if an attacker crafts the perfect malicious prompt, the agent's world is a sandbox.

---

## Architecture (Docker-Isolated)

```
┌─────────┐     ┌─────────────────────────────────────────────────────────┐
│  Phone   │────►│                brain.exe (HOST)                         │
│ (Flutter)│◄────│                                                         │
└─────────┘     │  ┌─────────────┐                                        │
                │  │  Classifier  │  (ministral, local, no network)        │
                │  │  brain.db    │  (SQLite, entries CRUD)                │
                │  │  relay client│  (ibeco.me WebSocket)                  │
                │  └──────┬──────┘                                        │
                │         │ HTTP (localhost only)                          │
                │         ▼                                                │
                │  ┌──────────────────── Docker Container ───────────────┐ │
                │  │                                                     │ │
                │  │  copilot-agent (Go binary, Copilot SDK)             │ │
                │  │  copilot CLI  (managed by SDK)                      │ │
                │  │  Models: claude-sonnet / opus / haiku / gpt-5       │ │
                │  │                                                     │ │
                │  │  MCP Servers (inside container):                     │ │
                │  │  • gospel-vec  (read-only vector DB)                │ │
                │  │  • gospel-mcp  (read-only FTS index)                │ │
                │  │  • webster-mcp (read-only dictionary)               │ │
                │  │                                                     │ │
                │  │  /workspace/                                        │ │
                │  │  ├── .github/copilot-instructions.md                │ │
                │  │  ├── .github/agents/study.agent.md                  │ │
                │  │  ├── .github/skills/                                │ │
                │  │  ├── /scriptures/  ← bind mount READ-ONLY           │ │
                │  │  ├── /study/       ← bind mount READ-ONLY           │ │
                │  │  └── /output/      ← bind mount READ-WRITE          │ │
                │  │                                                     │ │
                │  │  ✕ NO access to: brain.db, host filesystem,         │ │
                │  │    Docker socket, SSH keys, git creds,              │ │
                │  │    host network (except mapped ports)               │ │
                │  └─────────────────────────────────────────────────────┘ │
                └─────────────────────────────────────────────────────────┘
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

### Docker Container Details

The container is **ephemeral** — brain.exe starts it on demand and can destroy/recreate it at any time. No persistent state inside the container.

**Bind mounts:**

| Mount | Source (host) | Target (container) | Mode |
|-------|--------------|-------------------|------|
| Scriptures | `gospel-library/eng/scriptures/` | `/workspace/scriptures/` | `ro` |
| Study docs | `study/` (selected subset) | `/workspace/study/` | `ro` |
| MCP data | `data/gospel.db`, `data/vectors/`, `data/webster.db` | `/data/` | `ro` |
| Output | `brain-data/copilot-output/` | `/workspace/output/` | `rw` |
| Workspace config | `brain-data/copilot-workspace/.github/` | `/workspace/.github/` | `ro` |

**Not mounted:** `brain.db`, `~/.ssh/`, `~/.gitconfig`, Docker socket, any host directory not listed above.

**Network:** `--network=none` for the POC. The container needs internet for Copilot API calls, so in practice we'd use a restricted network that only allows outbound HTTPS to `api.github.com` and Anthropic/OpenAI endpoints. Or: brain.exe proxies the Copilot API calls itself, and the container uses `CLIUrl` to connect to a CLI process running on the host.

### Isolated Workspace Layout

```
brain-data/copilot-workspace/
├── .github/
│   ├── copilot-instructions.md   ← curated for phone study
│   ├── agents/
│   │   └── study.agent.md        ← phone study agent (concise format)
│   └── skills/
│       ├── scripture-linking/
│       ├── webster-analysis/
│       └── source-verification/
└── (bind mounts appear here at runtime)
    ├── scriptures/   → gospel-library/eng/scriptures/ (ro)
    ├── study/        → study/ (ro)
    └── output/       → brain-data/copilot-output/ (rw)
```

**Why Docker instead of just hooks:**
- **Hooks are software controls.** A sufficiently clever prompt injection could try to convince the model to bypass hook logic. Docker is a kernel-level boundary — no prompt can escape a container.
- **Trust building** — "line upon line, precept upon precept" (2 Nephi 28:30). We don't hand over the keys to everything on day one.
- **Progressive context** — start with read-only scripture tools, earn write access over time.
- **Stewardship** — the agent operates within a defined stewardship. It can search scriptures, answer questions, and save study notes. It cannot edit code, push commits, or touch brain.db.
- **Defense in depth** — hooks are the inner layer (what the agent *should* do), Docker is the outer layer (what the agent *can* do). Both must agree for a write to succeed.

### Relay Hardening

Even with Docker, we should harden the relay→copilot path:

1. **Separate auth token** — copilot commands require a different token than the relay. Compromising ibeco.me gets the relay token, not the copilot token.
2. **Rate limiting** — max 10 copilot commands per hour from relay. Legitimate phone use stays well under this; attack scripts would hit the wall.
3. **Command audit log** — every copilot command logged with timestamp, source IP, hash of content. Reviewable via brain.exe web UI.
4. **Kill switch** — brain.exe config flag to disable copilot entirely. One setting, agent goes dark.
5. **No relay for POC** — Phase 1 is direct-mode only (phone on same LAN as brain.exe). Relay support added later after trust is established.

### Full Workspace (Phase 4+ — After Trust is Established)

Once the isolated Docker workspace proves reliable:
- Add more bind mounts (lessons, journal, etc.)
- Graduate to a less restricted network
- Eventually point `Cwd` at the actual scripture-study repo (still containerized)
- Enable additional agents (dev, review, lesson)
- Write operations with both hook approval AND Docker mount control

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

## Phone-Friendly Study Format

Deep study documents don't work on a phone — 3000 words of cross-referencing is a desktop experience. The phone study agent should produce concise, structured responses optimized for small screens.

### Phone Study Agent Instructions

```markdown
# Phone Study Agent

You are a scripture study companion. The user is reading on a phone.
Keep responses concise and scannable.

## Format Rules
- Max 300 words unless the user asks for more
- Use headers and bullet points, not dense paragraphs
- One "sit with this" insight per response — the single most
  interesting thing you found
- Cross-references as a short list, not woven into prose
- Always cite the source, but don't quote extensively —
  the user can look it up
- If there's a Webster 1828 insight, include it as a one-liner

## Read Before Quoting
Same rule as the full study agent — read_file the source
before putting anything in quotes. On phone, paraphrase
is usually better than long quotes anyway.

## Example Response Shape

### [Topic] — [Primary Reference]

**Key insight:** [1-2 sentences]

**Cross-refs:**
- [ref 1] — [one-line summary]
- [ref 2] — [one-line summary]

**Webster 1828:** "[word]" — [short definition]

**Sit with this:** [One provocative observation or question]
```

### Example Output

```markdown
### Natural Man — 1 Cor 2:14

**Key insight:** Paul's "natural man" (Greek ψυχικός / psychikos)
means "soulish" — governed by appetites, not evil. Mosiah 3:19
expands this: the natural man is an *enemy to God* until he
yields to the Holy Spirit.

**Cross-refs:**
- Mosiah 3:19 — puts off natural man through the atonement
- Alma 26:21 — we were natural men before conversion
- Romans 8:7 — carnal mind is enmity against God

**Webster 1828:** "Natural" — belonging to nature; not acquired;
not spiritual or supernatural.

**Sit with this:** The natural man isn't fallen — he's
unfinished. Not broken, but not yet yielded.
```

### Write Access for Study Saves

The `/output/` bind mount (read-write) is where the agent saves study session notes. When the user says "save this" or the session produces something worth keeping, the agent writes a markdown file:

```
/workspace/output/2026-03-07-natural-man.md
```

brain.exe can then:
1. Watch the output directory for new files
2. Create a brain entry from the file (category: study)
3. Optionally copy to the main `study/` directory after human review

This gives write access that's contained (one directory, only markdown files) and reviewable (human sees what was produced before it enters the main corpus).

---

## Use Cases (Ordered by Phase)

### Phase 1: Study Mode (Docker-isolated, read-only, direct LAN only)

```
User (phone, same LAN): "What does Paul mean by 'the natural man'?"

brain.exe → Docker container → copilot-agent → claude-sonnet
  Agent uses:
  • gospel-vec (semantic search for "natural man") — inside container
  • gospel-mcp (read the actual verses) — inside container
  • webster-mcp (look up "natural" in Webster 1828) — inside container
  • Workspace skills: scripture-linking, source-verification
  • Phone study agent format (concise, 300 words max)

→ Returns phone-friendly study response with citations
→ User says "save this" → agent writes to /output/ → brain.exe picks it up
```

This works because:
- All MCP servers run inside the container with read-only data
- The container's workspace has the phone study agent and curated skills
- The agent can only write to `/output/` — everything else is read-only
- No relay involved — direct LAN connection only
- Docker boundary prevents any filesystem escape

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

### Phase 4: Relay Support (with hardened auth)

```
User (away from home): same study questions, via ibeco.me relay
```

Relay path enabled with:
- Separate copilot auth token (not the relay token)
- Rate limiting (10/hour)
- Audit logging
- Kill switch in brain.exe config

### Phase 5: GitHub + Code Integration (full workspace, still containerized)

Expand the container's mounts and agents:
- Point `Cwd` at scripture-study repo (containerized, not host-direct)
- Add GitHub MCP server
- Enable dev, review, lesson agents
- Code operations with hook approval + Docker mount boundaries

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

### Phase 1: Docker POC + Read-Only Study (direct LAN only)

1. **Dockerfile** — Alpine + Copilot CLI + copilot-agent Go binary + MCP server binaries
2. **docker-compose.yml** — bind mounts, network restrictions, resource limits
3. Create `brain-data/copilot-workspace/.github/` — phone study agent, curated skills
4. `internal/copilot/agent.go` — client init via `CLIUrl` (TCP to container)
5. `OnPreToolUse` hook: deny all write operations except `/output/`
6. `POST /api/copilot/chat` endpoint with SSE streaming
7. Output watcher: brain.exe picks up saved studies from `/output/`
8. Test: phone → brain.exe → container → study question → streamed response
9. Test: "save this" → file appears in `brain-data/copilot-output/`

### Phase 2: Flutter Chat UI

1. Chat screen with Brain/Copilot mode toggle
2. SSE streaming text display with markdown rendering
3. Tool call status indicators
4. Voice input integration
5. "Save as entry" action on responses
6. Model selection in settings (haiku/sonnet/opus)

### Phase 3: Brain Intelligence (read-only brain tools in container)

1. Custom tools: `brain_search`, `brain_get`, `brain_recent` — brain.exe exposes a read-only API that the container calls (HTTP on host network)
2. Trust levels in config (read-only → create-only → full)
3. `OnPreToolUse` enforces trust level
4. "Chat about this entry" button on entry detail
5. Conversation history via `ResumeSession`

### Phase 4: Relay Support (hardened)

1. Separate copilot auth token
2. Rate limiting + audit logging in brain.exe
3. SSE passthrough or WebSocket-based streaming via ibeco.me
4. Kill switch in brain.exe config

### Phase 5: Expanded Workspace (still containerized)

1. Wider bind mounts (lessons, journal, full study/)
2. Additional agents (dev, review, lesson)
3. GitHub MCP server integration
4. Write operations with Docker mount boundaries + hook approval

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

## Docker POC: Concrete Steps

### Dockerfile

```dockerfile
FROM alpine:3.20

# Copilot CLI
COPY --from=builder /copilot /usr/local/bin/copilot

# Our binaries
COPY copilot-agent /usr/local/bin/copilot-agent
COPY gospel-mcp /usr/local/bin/gospel-mcp
COPY gospel-vec /usr/local/bin/gospel-vec
COPY webster-mcp /usr/local/bin/webster-mcp

# Workspace (instructions baked in, data mounted at runtime)
COPY workspace/ /workspace/

EXPOSE 8090
ENTRYPOINT ["copilot-agent", "--port", "8090", "--workspace", "/workspace"]
```

### docker-compose.yml

```yaml
services:
  copilot-sandbox:
    build: ./docker/copilot-sandbox
    ports:
      - "127.0.0.1:8090:8090"  # Only accessible from localhost
    volumes:
      - ../gospel-library/eng/scriptures:/workspace/scriptures:ro
      - ../study:/workspace/study:ro
      - ../data:/data:ro
      - ./copilot-output:/workspace/output:rw
    # Security restrictions
    read_only: true  # Root filesystem read-only
    tmpfs:
      - /tmp:size=100M
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    mem_limit: 512m
    cpus: 1.0
    # Network: allow outbound HTTPS only (for Copilot API)
    # For true POC, use network_mode: none and proxy API calls through brain.exe
```

### copilot-agent (the Go binary inside the container)

```go
// Minimal HTTP server that wraps the Copilot SDK
// brain.exe calls this via HTTP on localhost:8090

func main() {
    client := copilot.NewClient(&copilot.ClientOptions{
        Cwd:      "/workspace",
        LogLevel: "error",
    })
    client.Start(context.Background())
    defer client.Stop()

    http.HandleFunc("/chat", handleChat(client))
    http.HandleFunc("/health", handleHealth)
    http.ListenAndServe(":8090", nil)
}
```

### Alternative: CLI on Host, SDK in Container via TCP

The Copilot SDK supports connecting to an external CLI server via `CLIUrl`. This means:
- Copilot CLI runs on the **host** (where it has GitHub auth)
- copilot-agent runs in the **container** with `CLIUrl: "host.docker.internal:8080"`
- Container has zero network access except to the host CLI
- The CLI still respects the container's workspace for file operations

This might be simpler for auth and avoids putting the CLI inside the container. Trade-off: the CLI process runs on the host, so its built-in tools (file read/write, terminal) operate on the host filesystem. Need to test whether the SDK's `Cwd` constrains the CLI's file operations to that directory.

**Safer option for POC:** Everything inside the container. CLI + SDK + MCP servers. The container gets internet access for API calls (via a restricted network), but nothing else.

---

## Open Questions

- **MCP server transport:** Copilot CLI reads MCP config from `.github/copilot-instructions.md` or a config file. Inside the container, do MCP servers run as stdio subprocesses (launched by the CLI), or as HTTP servers? stdio is simpler — the CLI spawns them. Need to test.
- **CLI workspace sandboxing:** Does the Copilot CLI respect `Cwd` as a boundary, or can it navigate up? If the CLI inside the container tries `read_file /etc/passwd`, does it succeed? In Docker, yes — but `/etc/passwd` is the container's, which is fine. The bind mounts are the real boundary.
- **Copilot auth inside Docker:** The CLI needs GitHub auth. Options: (a) mount a read-only auth token file, (b) pass via environment variable, (c) BYOK with API key (no GitHub auth needed). For POC, BYOK with an Anthropic key is simplest.
- **Command detection accuracy:** If we go with Option A (classifier detects commands), how accurate will ministral be at distinguishing "capture this thought" from "do this for me"? For POC, use explicit copilot mode toggle (Option B) — no detection needed.
- **Container startup time:** Alpine + Go binaries should be fast (<2s). But does Copilot CLI startup add latency? Warm-start by keeping the container running.
- **Session lifetime:** Keep the container running continuously (like brain.exe itself) or start/stop per request? Keep running — startup overhead matters on phone.
- **LM Studio fallback:** When the machine is running but internet is down (no Copilot access), should we fall back to LM Studio inside the container with BYOK? This would require mounting LM Studio's API access. Defer for now — clean error is fine.
- **Cost monitoring:** Need a way to track premium request usage from the phone. Display in settings: "X copilot requests today / this month."
- **dokploy network isolation:** Separate concern from this feature. But worth noting: a VLAN between proxmox and your workstation would limit blast radius of a VPS compromise. Brain.exe connections to ibeco.me are outbound-only (WebSocket initiated by brain.exe), which helps.

---

## Coder Evaluation (March 8, 2026)

Evaluated [github.com/coder/coder](https://github.com/coder/coder) as a potential replacement for our custom Docker isolation layer.

### What Coder Is

A self-hosted **cloud development environment platform**. Teams define workspaces in Terraform, Coder provisions them (Docker containers, K8s pods, cloud VMs), runs an agent inside each one, and connects developers via Wireguard tunnels. It's enterprise infrastructure for at-scale dev environments.

### What Overlaps with Plan 13

| Feature | Coder | Our Plan 13 |
|---------|-------|-------------|
| Docker container isolation | Yes (Terraform-provisioned) | Yes (docker-compose, bind mounts) |
| Agent process inside container | Yes (Coder agent: SSH, file transfer, port forwarding, process monitoring) | Yes (copilot-agent: Copilot SDK + MCP servers) |
| MCP integration | Yes (Coder Tasks + AI Bridge) | Yes (gospel-vec, webster-mcp, etc.) |
| Workspace lifecycle (start/stop/delete) | Yes (full state machine, auto-shutdown) | Minimal (keep-alive, restart on failure) |
| AI agent execution | Yes (Claude Code, Aider, Goose via Coder Tasks) | Yes (Copilot SDK with custom tools) |

### What Coder Adds That We Don't Need

- **PostgreSQL requirement** — Coder needs a full Postgres instance for state. We'd go from "brain.exe is a single binary" to needing a database server for the isolation layer.
- **External provisioner daemon** — Terraform-based provisioning is a heavyweight dependency for "start one Docker container."
- **Wireguard tunnels** — Secure networking for multi-user, multi-region access. We're on localhost.
- **Team/RBAC/quota management** — Enterprise features for organizations. We're one person.
- **IDE integration** (SSH, VS Code Remote) — Designed for developers working *inside* the workspace. Our agent works autonomously; the user interacts via phone.
- **Template system** — Powerful but complex. We need one container config, not a Terraform template ecosystem.
- **AI Bridge** (enterprise) — Centralized AI gateway with audit trails and token spend monitoring. Interesting but behind a paywall and overkill for single-user.

### What Coder Does Well That's Interesting

- **Coder Tasks** — dedicated UI for running autonomous coding agents (Claude Code, Aider) with MCP status reporting. This is the closest overlap: running an AI agent in an isolated workspace and reporting progress back. But it's designed for coding tasks, not scripture study.
- **Agent architecture** — their workspace agent is mature: SSH, file transfer, reconnecting PTY, process monitoring, health reporting. If we ever wanted a richer agent runtime (e.g., live terminal output streamed to phone), Coder's agent code is a reference implementation.
- **Security model** — external provisioners, scoped keys, network isolation, read-only filesystems. Good patterns to borrow even if we don't use Coder itself.

### Verdict: Overkill, but a Good Reference

**Don't adopt Coder.** The complexity-to-value ratio is wrong for our use case:

- We need: "run Copilot SDK + MCP servers in a locked-down Docker container, accessible only from brain.exe on localhost."
- Coder provides: "provision multi-cloud dev environments for teams with Terraform, Wireguard, PostgreSQL, and RBAC."
- The overhead (PostgreSQL, provisioner daemon, Wireguard, Terraform) exceeds the value for a single-user, single-container scenario on a Windows workstation.

**What to borrow from Coder:**
- Their Docker template security patterns (read-only root, `no-new-privileges`, `cap_drop: ALL`, memory/CPU limits) — we already have these in our docker-compose.yml above
- Coder Tasks' MCP reporting model — when our agent completes a study session, it could report status via MCP to brain.exe (like Coder Tasks report to coderd)
- Their agent reconnection logic — if we ever need the copilot-agent to survive container restarts gracefully
- The `AI Bridge` concept — if we eventually want centralized API key management and audit logging for LLM calls, this is a clean pattern (but build our own lightweight version, not adopt Coder's enterprise feature)

**Bottom line:** Keep our custom Docker isolation. It's simpler, fits the single-user model, and we control the entire stack. Coder is a good project to study but the wrong tool for this job.
