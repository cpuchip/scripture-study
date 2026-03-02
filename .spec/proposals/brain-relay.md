# Spec: Brain Relay via ibeco.me

*Created: March 2, 2026*
*Status: Draft — pending Michael's review*
*Origin: [Second Brain Architecture proposal](.spec/proposals/second-brain-architecture.md) → Phase 3 pulled forward as Phase 1*
*Related: [Becoming App plan](scripts/plans/06_becoming-app.md), [Intent Engineering](docs/work-with-ai/04_intent-engineering.md)*

---

## Intent

> Replace Discord as the brain's capture interface with a private, always-available channel through ibeco.me — using the existing auth, infrastructure, and deployment pipeline. The phone app (Dart/Flutter) becomes the universal capture point. The brain runs wherever Michael is working.

### Why

Discord requires: a server, an invite, a bot token, privileged intents, and Discord installed everywhere. It's a public platform bolted into a private workflow. The becoming app (ibeco.me) is already:
- Deployed on Dokploy (SLC region, auto-deploy on main push)
- Authenticated (bearer tokens, Google OAuth, session cookies)
- On Michael's phone (via PWA or installable as an app)
- A Go backend Michael knows intimately
- The natural home for "capture → classify → become" flow

### What Changes and What Doesn't

| Stays the Same | Changes |
|---|---|
| `classifier.go` — same AI classification pipeline | New `internal/relay/` transport replaces Discord as primary |
| `store.go` + `git.go` — same file + git storage | ibeco.me gains WebSocket relay hub |
| `config.go` — same model presets | New Dart/Flutter mobile app |
| `ai/client.go` — same Copilot SDK | Brain connects out to ibeco.me instead of in to Discord |
| Discord code stays in repo (future public study bot) | Config gains `RELAY_URL` + `RELAY_TOKEN` |

### Success Criteria

1. Michael can type a thought on his phone and find it classified + committed in `private-brain/` within seconds
2. Works when brain.exe runs on any machine (home PC, laptop, etc.) — no port forwarding, no firewall config
3. Messages queue on ibeco.me if brain is offline; deliver on reconnect
4. Zero prompt injection surface — only Michael's authenticated token can send
5. Total new code < 1000 lines across all three components
6. Deploys within one session (~3 hours)

### Constraints

- **No new infrastructure.** Use ibeco.me (Dokploy) — don't add a VPS, don't add a database, don't add Redis.
- **Same SQLite.** Becoming already uses SQLite. The message queue lives in that same DB.
- **Bearer token auth only.** The brain connects with the same `BECOMING_TOKEN` mechanism the MCP uses. No new auth systems.
- **Go + Dart only.** No TypeScript/Node in the pipeline.
- **Outbound-only from brain.** Brain connects to ibeco.me (wss://). Brain never listens on a port. This is critical — Michael's PC is never exposed.

### Decision Boundaries (Agent)

- **Autonomous:** File structure, function signatures, error handling patterns, test structure
- **Clarify first:** Database schema changes to becoming.db, API route naming, Dart UI layout choices
- **Human decides:** Auth token scoping, what data is visible in the mobile app, deployment timing

---

## Architecture

```
┌─────────────────┐     wss://ibeco.me      ┌──────────────┐    wss://ibeco.me     ┌────────────────┐
│  Dart/Flutter    │ ◄════════════════════► │   ibeco.me   │ ◄════════════════════► │   brain.exe    │
│  Mobile App      │   role: "app"          │  (relay hub) │   role: "agent"        │  (Michael's PC) │
│                  │                         │              │                         │                │
│ • Text input     │   ──── thought ────►  │  • Auth      │   ──── thought ────►   │ • Classify     │
│ • Voice-to-text  │                         │  • Route     │                         │ • Store + git  │
│ • History view   │   ◄── result ──────   │  • Queue     │   ◄── result ──────    │ • Audit log    │
│ • Offline queue  │                         │  • Presence  │                         │                │
└─────────────────┘                         └──────────────┘                         └────────────────┘
```

### Message Protocol (WebSocket JSON)

All messages are JSON objects with a `type` field:

```jsonc
// Client → Server: authenticate after WS connect
{ "type": "auth", "token": "bec_...", "role": "app" | "agent" }

// Server → Client: auth result
{ "type": "auth_ok", "user_id": 1 }
{ "type": "auth_error", "error": "invalid token" }

// App → Server → Agent: new thought to classify
{ "type": "thought", "id": "uuid", "text": "...", "timestamp": "2026-03-02T10:30:00Z" }

// Agent → Server → App: classification result
{ "type": "result", "thought_id": "uuid", "category": "projects", "title": "...", "confidence": 0.87, "tags": [...], "needs_review": false, "file_path": "projects/2026-03-02-brain-relay.md" }

// App → Server → Agent: reclassify
{ "type": "fix", "thought_id": "uuid", "new_category": "ideas" }

// Agent → Server → App: fix confirmed
{ "type": "fix_ok", "thought_id": "uuid", "new_path": "ideas/2026-03-02-brain-relay.md" }

// Server → App: agent status
{ "type": "presence", "agent_online": true }

// Server → Client: queued messages on reconnect
{ "type": "queued", "messages": [...] }

// Agent → Server: status update
{ "type": "status", "model": "gpt-5-mini", "categories": {"people": 3, "projects": 12, ...} }
```

### Queue Behavior

| Scenario | Behavior |
|---|---|
| App sends thought, agent online | Relay immediately; app sees result in ~2-5 seconds |
| App sends thought, agent offline | Queue in SQLite; app sees `{"type": "presence", "agent_online": false}`; app shows "Queued — brain offline" |
| Agent comes online | Server sends `{"type": "queued", "messages": [...]}` with all pending thoughts |
| Agent sends result for queued thought | Server relays to app if connected; otherwise stores result for next app connect |
| Both offline | Queue on server; deliver when each reconnects |

---

## Component Specs

### 1. ibeco.me — WebSocket Relay (Go)

**New files:**

| File | Purpose |
|---|---|
| `internal/brain/hub.go` | WebSocket hub — manages connections, routing, presence |
| `internal/brain/messages.go` | Message type definitions + JSON serialization |
| `internal/brain/queue.go` | SQLite queue — store/retrieve pending messages |

**New database table:**

```sql
CREATE TABLE IF NOT EXISTS brain_messages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id  TEXT NOT NULL UNIQUE,        -- UUID from the thought
    user_id     INTEGER NOT NULL,            -- owner
    direction   TEXT NOT NULL,               -- 'to_agent' or 'to_app'
    payload     TEXT NOT NULL,               -- full JSON message
    status      TEXT NOT NULL DEFAULT 'pending',  -- 'pending' | 'delivered'
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_brain_messages_pending ON brain_messages(user_id, status, direction);
```

**Route registration** (in `cmd/server/main.go`):

```go
// Brain relay (WebSocket, auth via token in first message)
brainHub := brain.NewHub(database)
r.Get("/ws/brain", brainHub.HandleWebSocket)

// Brain REST endpoints (for history, status)
r.Group(func(r chi.Router) {
    r.Use(auth.Required(database, *dev))
    r.Get("/api/brain/history", brainHub.HandleHistory)
    r.Get("/api/brain/status", brainHub.HandleStatus)
})
```

**Hub logic:**
- Accepts WebSocket upgrade
- First message must be `{"type": "auth", ...}` — validates bearer token via `database.ValidateAPIToken()`
- Registers connection as `app` or `agent` role
- Routes `thought` messages from app → agent (or queue if offline)
- Routes `result` / `fix_ok` messages from agent → app (or queue if offline)
- On agent connect: deliver all `pending` + `direction=to_agent` messages
- On app connect: deliver all `pending` + `direction=to_app` messages
- Ping/pong keepalive every 30 seconds
- Clean disconnect on close

**Estimated lines:** ~250

### 2. brain.exe — Relay Transport (Go)

**New files:**

| File | Purpose |
|---|---|
| `internal/relay/client.go` | WebSocket client — connect to ibeco.me, auth, receive/send |

**New config vars:**

| Variable | Default | Description |
|---|---|---|
| `RELAY_URL` | `wss://ibeco.me/ws/brain` | Relay WebSocket endpoint |
| `RELAY_TOKEN` | _(required if relay enabled)_ | Bearer token (same `BECOMING_TOKEN` format) |
| `RELAY_ENABLED` | `true` | Enable relay transport |
| `DISCORD_ENABLED` | `false` | Enable Discord transport (keep for future) |

**Relay client logic:**
- Connects to `RELAY_URL` via gorilla/websocket
- Sends `{"type": "auth", "token": "bec_...", "role": "agent"}`
- On `thought` message: calls existing `classifier.Classify()` → `store.Save()` → `store.SaveAudit()` → sends `result` back
- On `fix` message: calls existing `store.Reclassify()` → sends `fix_ok` back
- Auto-reconnect with exponential backoff (1s → 2s → 4s → ... → 30s max)
- On shutdown: graceful close, `git.Push()`

**`cmd/brain/main.go` changes:**

```go
// Current: always starts Discord
// New: start relay and/or Discord based on config

if cfg.RelayEnabled {
    relayClient := relay.NewClient(cfg.RelayURL, cfg.RelayToken, classify, st)
    go relayClient.Run(ctx)  // blocks until ctx cancelled
}

if cfg.DiscordEnabled && cfg.DiscordToken != "" {
    bot, err := discord.NewBot(cfg.DiscordToken, classify, st, cfg.RateLimits.MaxNotificationsPerDay)
    // ... existing Discord setup
}
```

**Estimated lines:** ~200

### 3. Dart/Flutter Mobile App

**New repo or directory:** TBD — either `scripts/brain-app/` or a separate repo

**Tech stack:**
- Flutter (Dart)
- `web_socket_channel` package for WebSocket
- Material 3 / Material You theming
- Shared preferences for token storage

**Screens:**

| Screen | Purpose |
|---|---|
| **Login** | Enter ibeco.me URL + API token (or scan QR from web UI). Store locally. |
| **Capture** | Text input (main screen). Send button. Optional voice-to-text via platform API. Shows last few classified thoughts as a scrollable list. |
| **History** | Full list of classified thoughts. Filter by category. Tap to see details. |
| **Settings** | Server URL, token, theme, about. |

**Capture screen layout (main):**

```
┌─────────────────────────────────┐
│  🧠 Brain          [●] Online  │  ← status bar + presence indicator
├─────────────────────────────────┤
│                                 │
│  ┌─────────────────────────┐   │
│  │ 📁 projects  87%        │   │  ← most recent result
│  │ "Brain relay spec"      │   │
│  │ #brain #architecture    │   │
│  └─────────────────────────┘   │
│                                 │
│  ┌─────────────────────────┐   │
│  │ 📁 actions  92%         │   │  ← previous result
│  │ "Call dentist Monday"   │   │
│  │ #health #admin          │   │
│  └─────────────────────────┘   │
│                                 │
│  ... more results ...          │
│                                 │
├─────────────────────────────────┤
│  [Type a thought...        ] 📤│  ← input bar (always visible)
└─────────────────────────────────┘
```

**Offline behavior:**
- If WebSocket disconnects, show "Brain offline — message will be queued"
- App-side queue in shared preferences (simple list) — send on reconnect
- Results from queued thoughts arrive asynchronously when brain processes them

**Estimated lines:** ~500 (Dart)

---

## Implementation Sequence

### Phase A — ibeco.me relay hub (~45 min)

| Step | Task | Test |
|---|---|---|
| A1 | Create `brain_messages` table migration in `internal/db/` | `go test` — table exists |
| A2 | Create `internal/brain/messages.go` — type definitions | Compiles |
| A3 | Create `internal/brain/queue.go` — SQLite queue operations | Unit tests for enqueue/dequeue/markDelivered |
| A4 | Create `internal/brain/hub.go` — WebSocket hub + routing | Integration test with two WS clients |
| A5 | Register `/ws/brain` route in `cmd/server/main.go` | `wscat` can connect + auth |
| A6 | Add `/api/brain/history` and `/api/brain/status` REST endpoints | curl returns data |
| A7 | Deploy via Dokploy | `wss://ibeco.me/ws/brain` accepts connections |

### Phase B — brain.exe relay transport (~30 min)

| Step | Task | Test |
|---|---|---|
| B1 | Add `RELAY_*` config vars to `config.go` | `config.Load()` reads them |
| B2 | Create `internal/relay/client.go` — WS client + reconnect | Compiles |
| B3 | Wire `thought` → classify → store → `result` pipeline | End-to-end: send thought via wscat → see file in private-brain |
| B4 | Wire `fix` → reclassify → `fix_ok` pipeline | Reclassify works |
| B5 | Update `main.go` — relay+Discord as parallel transports | Brain starts with relay, Discord optional |
| B6 | Test full loop: phone-simulated → ibeco.me → brain → file | File appears in git |

### Phase C — Dart mobile app (~45 min)

| Step | Task | Test |
|---|---|---|
| C1 | Create Flutter project with Material 3 scaffold | `flutter run` shows empty screen |
| C2 | Build login screen (URL + token input, persistent storage) | Token persists across app restart |
| C3 | Build WebSocket service (connect, auth, send/receive JSON) | Can send `thought` and receive `result` from relay |
| C4 | Build capture screen (input bar + result list) | Type thought → see classified result appear |
| C5 | Add presence indicator (agent online/offline) | Shows green/red dot |
| C6 | Add offline queue (local buffer, deliver on reconnect) | Queue survives app restart |
| C7 | Add history screen (paginated, filterable) | Browse past thoughts |
| C8 | Polish — loading states, error handling, haptic feedback | Feels solid |

### Phase D — Integration + polish (~15 min)

| Step | Task | Test |
|---|---|---|
| D1 | End-to-end test: phone → ibeco.me → brain → git → response on phone | Full loop works |
| D2 | Offline test: send while brain offline → brain comes online → receive results | Queue delivers |
| D3 | Commit all three components, push | Clean repos |
| D4 | Update `.spec/memory/active.md` | Current state captured |

---

## Security Model

| Risk | Mitigation |
|---|---|
| **Prompt injection via thought text** | Only Michael's authenticated token can send. No public input surface. |
| **Token exposure in mobile app** | Stored in platform secure storage (SharedPreferences on Android, Keychain on iOS). Not logged. |
| **Brain.exe as attack surface** | Never listens on a port. Outbound WebSocket only. No code execution from messages — just text classification. |
| **ibeco.me relay as open pipe** | Auth required on WebSocket connect. Only validated tokens accepted. Rate limiting on message frequency. |
| **Man-in-the-middle** | WSS (TLS) enforced by Dokploy/Traefik. Certificate pinning optional in Dart app. |
| **Future public Discord bot** | Completely isolated context. Separate config. No access to private-brain repo. Free/local models only. No git write. Explicit scope boundary. |

---

## Resolved Questions

1. **Dart app repo** — Separate repo: `cpuchip/brain-app` (cloned to `scripts/brain-app/`). Different language, different deployment target. Added to `.gitignore`.

2. **Voice** — YES. Both directions:
   - **Voice-to-text** (capture): Flutter `speech_to_text` package. Lowers friction dramatically for on-the-road capture.
   - **Text-to-voice** (playback): Flutter `flutter_tts` package. When Michael is driving and wants to hear what the brain classified, or chat with the second brain about past entries, the app reads responses aloud. This turns the brain into a **conversational companion** for road time.

3. **Token auth** — QR code scan or paste from web UI. Matches existing MCP token flow. Google OAuth login can be added later.

4. **History** — Recent by default (server-side from `brain_messages` table). Full history available via commands. But the bigger vision: **brain as cross-platform memory hub.** Each entry tagged with `workspace`/`repo` context so no matter where Michael works, there's continuity. See "Cross-Platform Memory" section below.

## Cross-Platform Memory (Vision — from Michael)

> *"This could be a good help for you in cross platform memory. We could have it save entries from what workspace/repo we have these notes/memories in so that no matter where I work I've got you (well some piece of you) to help along for the ride."*

This is bigger than capture-and-classify. This is about making the brain a **context bridge** between workspaces:

| Field | Example | Purpose |
|---|---|---|
| `workspace` | `scripture-study` | Which repo/project the thought relates to |
| `source` | `phone`, `vscode-copilot`, `brain-app` | Where it was captured |
| `session_id` | `2026-03-02-study` | Link to a specific working session |

When the brain runs as an MCP server (which it will — Copilot SDK gives us that), any workspace can query: "What did Michael capture about this project from his phone last week?" or "What decisions were made in the scripture-study workspace this month?"

This connects directly to Nate's "Open Brain" concept (see video analysis below) — one brain, every tool, persistent memory that never starts from zero.

---

## Relationship to Existing Specs

- **[Second Brain Architecture](second-brain-architecture.md)** — This spec implements Phase 1 + Phase 3 of that proposal simultaneously, skipping VPS Phase 2 (not needed since ibeco.me is already always-on).
- **[Becoming App](scripts/plans/06_becoming-app.md)** — ibeco.me gains brain relay capability alongside its existing practice/task/memorization features. Same backend, same deploy.
- **[Intent Engineering](docs/work-with-ai/04_intent-engineering.md)** — The brain's classifier system is a concrete implementation of intent engineering: the system prompt encodes *what kind of thinking to do*, not just *what to do*.

---

## Video Analysis: Nate B Jones — "The $0.10 System That Replaced My AI Workflow"

*[You Don't Need SaaS. The $0.10 System That Replaced My AI Workflow](https://www.youtube.com/watch?v=2JiMmye2ezg) — March 2, 2026*

### Nate's Core Thesis

Nate evolves his second brain concept from Part 1/2 into what he calls **"Open Brain"** — a database-backed, MCP-accessible knowledge system that any AI tool can plug into. The key shift: the original second brain (Slack → Notion) was built for the **human web**. Open Brain is infrastructure for the **agent web**.

### Architecture He Proposes

```
[Capture] → [Supabase Edge Function] → [Postgres + pgvector]
                                              ↑
                                        [MCP Server]
                                              ↑
                              [Claude / ChatGPT / Cursor / any MCP client]
```

- Postgres + pgvector for storage and semantic search
- Every thought gets: raw text + vector embedding + extracted metadata (people, topics, type, actions)
- MCP server exposes 3 tools: `semantic_search`, `list_recent`, `stats`
- Any MCP-compatible client becomes both a capture point and search tool
- Cost: ~$0.10-0.30/month on Supabase free tier

### What Nate Gets Right

1. **[Memory architecture > model selection, 4:48](https://www.youtube.com/watch?v=2JiMmye2ezg&t=288)** — "Memory architecture determines agent capabilities much more than model selection does." This is exactly what we've discovered with `.spec/memory/` — the memory structure IS the intelligence multiplier.

2. **[The walled garden problem, 5:34](https://www.youtube.com/watch?v=2JiMmye2ezg&t=334)** — "Claude's memory doesn't know what you told ChatGPT. ChatGPT's memory doesn't follow you into Cursor." This is the exact problem Michael articulated — cross-platform memory. Our brain relay solves this by making ibeco.me the hub that *any* client connects to.

3. **[One brain, every AI, 14:52](https://www.youtube.com/watch?v=2JiMmye2ezg&t=892)** — "One brain, every AI, persistent memory that never starts from zero." This is our vision. The difference: Nate's Open Brain is a passive database. Ours is an active agent that classifies, stores, AND responds.

4. **[MCP as write + read, 20:18](https://www.youtube.com/watch?v=2JiMmye2ezg&t=1218)** — "MCP means you can write directly into the brain from anywhere." This validates our architecture — the brain isn't just a search endpoint, it's a bidirectional capture + retrieval system.

5. **[Memory migration, 22:17](https://www.youtube.com/watch?v=2JiMmye2ezg&t=1337)** — Extract existing memory from Claude/ChatGPT into your own system. We should do this — our `.spec/memory/` system has months of accumulated context that should flow into the brain.

6. **[Compounding advantage, 18:41](https://www.youtube.com/watch?v=2JiMmye2ezg&t=1121)** — "Every thought captured makes the next search smarter." This is the gospel principle of line upon line. The system grows.

### What Nate Doesn't Have (That We Do)

| Gap | Our Answer |
|---|---|
| **No active agent** — his system is a passive DB + MCP search | Our brain.exe actively classifies, routes, and responds in real-time |
| **No git versioning** — Postgres is the only store | Every entry is a markdown file in a git repo — versioned, diffable, portable |
| **No conversational capture** — type text, get confirmation | Our app will have voice-to-text AND text-to-voice for road conversations |
| **No spiritual integration** — general productivity only | Our classifier has a `study` category; system connects to gospel library |
| **No self-improvement** — static architecture | Our Phase 4 (future) has the agent proposing improvements to itself |
| **No relay for agents** — brain is the MCP server, period | Our relay makes ibeco.me the hub — brain connects outbound, never exposed |
| **Supabase dependency** — "no SaaS" but uses Supabase | Our system is truly self-owned: SQLite/PostgreSQL + Dokploy + git |

### What We Should Adopt from This Video

1. **Vector embeddings on capture.** Nate's right that semantic search is the killer feature. Our brain stores markdown files — we should ALSO generate embeddings on capture (gospel-vec already has the infrastructure). Add a `brain_embeddings` table to ibeco.me or have brain.exe call the local embeddings endpoint. This gives us "search by meaning" across all captured thoughts.

2. **MCP server for the brain.** Brain.exe should expose an MCP interface — not just receive via relay, but also be queryable by Copilot/Claude/any MCP client in any workspace. "What did I capture about covenants this week?" from inside a study session. This is the cross-platform memory Michael is asking for.

3. **Memory migration prompt.** We should create a prompt/tool that exports `.spec/memory/`, `.spec/journal/`, and key study insights into the brain's storage. Bootstrap the brain with 4+ months of accumulated context.

4. **Weekly review pattern.** Nate's weekly synthesis prompt is good: cluster by topic, scan for unresolved actions, detect patterns, find connections. Our session-journal already does some of this but we should formalize it as a brain capability.

### Impact on Our Spec

Nate's video **validates** our architecture and **extends** it in one important direction: the brain should be both a **consumer** (classify via relay) and a **provider** (MCP server for any workspace). This means:

**Phase E (post-launch):**
- Brain.exe exposes an MCP server alongside the relay client
- Any VS Code workspace can add the brain as an MCP server
- Tools: `brain_search` (semantic), `brain_recent`, `brain_capture`, `brain_stats`
- Embeddings generated on capture (reuse gospel-vec infrastructure)
- Workspace/repo tagging on every entry for cross-platform context

This doesn't change Phases A-D. Build the relay first, make it work, then add MCP read access as a Phase E enhancement.

---

## Implementation Progress

| Phase | Status | Notes |
|-------|--------|-------|
| **A — ibeco.me relay hub** | ✅ Done | `internal/brain/` package (hub.go, messages.go, queue.go). WebSocket at `/ws/brain`, REST at `/api/brain/status` and `/api/brain/history`. All routes wired in main.go. Compiles + vet clean. |
| **B — brain.exe relay client** | ✅ Done | `internal/relay/client.go`. Config now reads `IBECOME_TOKEN`/`IBECOME_URL` (with `RELAY_*` fallback). Auto-converts HTTP→WSS URLs. Auto-reconnect with exponential backoff. main.go supports relay + Discord as parallel transports. Compiles + vet clean. |
| **C — Dart mobile app** | Not started | `cpuchip/brain-app` repo cloned to `scripts/brain-app/`. |
| **D — Integration test** | ✅ Verified | Local test: app→hub→agent→classify→result→app. All 6 steps pass. Presence notification works both directions. Queue table created automatically. |
| **D.5 — Production test** | ✅ Verified | CLI→ibeco.me (production): auth works, status/history REST endpoints work, thought capture queues correctly when agent offline, timeout handling clean. Fixed PostgreSQL migration (008_brain_messages.sql) — EnsureTable DDL was SQLite-specific. |
| **E — CLI tool** | ✅ Done | `cmd/brain-cli/main.go`. Subcommands: `capture`, `status`, `recent`, `fix` (shortcuts: c/s/r/f). Bare text auto-captured. WS for capture/fix, REST for status/history. Proper timeout handling. |

---

## Expanded Roadmap (New Features — March 2 Update)

Based on Michael's feedback, the brain is growing beyond a simple capture-classify tool. Here's the expanded vision:

### Embeddings Strategy

**No external API needed.** Claude Pro ($20/mo) is web UI only — no embeddings API. Instead:
- **chromem-go** (already used by gospel-vec) provides local sentence-transformer embeddings — free, fast, zero external dependency
- Brain.exe generates embeddings on every capture alongside classification
- Stored in a local `brain.db` SQLite file with vector columns (same pattern as gospel-vec)
- Enables semantic search: "What did I capture about covenants?" without exact keyword matches

### Phase E — MCP Server + CLI (Both needed)

Michael can't use local MCP servers at work. Two access patterns:

**MCP Server (home/personal workspaces):**
- Brain.exe exposes `stdio` MCP server alongside relay client
- Add to any workspace's MCP config
- Tools: `brain_search`, `brain_recent`, `brain_capture`, `brain_stats`

**CLI Tool (work + anywhere):**
- `brain capture "thought text"` — sends through relay
- `brain search "query"` — semantic search against embeddings
- `brain recent` — last N classified entries
- `brain status` — agent online, queue counts, model info
- `brain fix <id> <category>` — reclassify
- Works anywhere Go runs. Talks to ibeco.me REST API (no WebSocket needed for CLI).
- **Copilot Chat Skill** — a `@brain` participant that wraps the CLI, so brain is accessible from any VS Code Chat panel even without MCP

### Phase F — Brain App Enhanced (Flutter)

**Home Screen Widgets:**
- **Quick Capture (2×3 or 1×3):** Text field + send button, like Microsoft TODO widget. Tap → type → capture → done. No app launch needed.
- **Brain Digest (2×3):** Shows last 3-5 captured thoughts with categories and confidence. Tap to expand.
- **Action Items (1×3):** Shows pending actions/todos extracted from recent thoughts.
- Uses Android `home_widget` Flutter package + iOS WidgetKit via `home_widget`

**Push Notifications:**
- Brain.exe → ibeco.me relay → FCM/APNs → phone notification
- Types: result ready (for queued thoughts), daily digest, action reminders, brain alerts
- Notification channels: urgent (alerts), normal (results), low (digests)

**Voice I/O (Phase C includes this):**
- `speech_to_text` for capture — hold mic button, speak, release to send
- `flutter_tts` for playback — brain reads results aloud while driving

### Phase G — Periodic Wake-Up + Intelligence

The brain doesn't just wait for input — it proactively works:

**Scheduled Tasks:**
- **Morning digest:** Summarize yesterday's captures, surface pending actions
- **Weekly review:** Cluster by topic, find patterns, detect stale/unresolved items
- **News monitoring:** Watch configured sources (YouTube channels, RSS, etc.)
- **Auto-evaluate:** When Nate B Jones (or other configured creators) posts a video, download transcript, evaluate relevance, generate digest
- **Action follow-up:** "You captured 'Call dentist Monday' 3 days ago — did you do it?"

**Implementation:**
- Brain.exe runs a scheduler (cron-style) alongside the relay client
- Uses existing Copilot SDK for analysis
- Results pushed via relay → notifications on phone
- Config in `.brain/config.yaml`: which sources to watch, digest times, alert thresholds

### Phase H — Memory Hub (Cross-Platform)

Every entry tagged with:
- `workspace`: which repo/project context
- `source`: phone, cli, vscode, discord
- `session_id`: link to working session
- `embedding`: vector for semantic search

Any workspace can query the brain: "What decisions were made about the becoming app last week?" — works from VS Code Chat, CLI, or MCP.

---

## Review Checkpoint

Phases A, B, D, D.5 (production), E complete and verified. Production ibeco.me is live with brain relay. CLI tested end-to-end through production.

**Next steps:** Start brain.exe agent locally to process queued thoughts, then Phase C (Dart app) or Phase E.5 (MCP server + embeddings).

*"Created all things spiritually, before they were naturally upon the face of the earth."* — [Moses 3:5](gospel-library/eng/scriptures/pgp/moses/3.md)
