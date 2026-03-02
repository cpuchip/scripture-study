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
- Deployed on Fly.io (SLC region, always on)
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

- **No new infrastructure.** Use ibeco.me (Fly.io) — don't add a VPS, don't add a database, don't add Redis.
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
| A7 | Deploy to Fly.io | `wss://ibeco.me/ws/brain` accepts connections |

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
| **Man-in-the-middle** | WSS (TLS) enforced by Fly.io. Certificate pinning optional in Dart app. |
| **Future public Discord bot** | Completely isolated context. Separate config. No access to private-brain repo. Free/local models only. No git write. Explicit scope boundary. |

---

## Open Questions

1. **Dart app as separate repo?** Or `scripts/brain-app/` in scripture-study? Separate repo keeps it clean but adds overhead. Leaning toward **separate repo** (`cpuchip/brain-app`) since it's a different language and deployment target.

2. **Voice-to-text?** Flutter's `speech_to_text` package works natively on Android/iOS. Worth adding in Phase C or defer to a follow-up? Leaning toward **include it** — it's ~30 lines of Dart and dramatically lowers capture friction.

3. **API token generation UI?** Right now tokens are created via the becoming web UI. The Dart app needs a token. Two options:
   - (a) Login with Google OAuth in the app → auto-generate token
   - (b) Generate token in web UI, paste/scan QR into app
   Option (b) is simpler and matches how the MCP already works. Leaning **(b)**.

4. **History in the app vs. just recent?** The app could show the last 20 thoughts (from server memory) or pull from the git repo. Leaning toward **server-side recent list** (the brain_messages table already has them) — no need to parse git on every refresh.

---

## Relationship to Existing Specs

- **[Second Brain Architecture](second-brain-architecture.md)** — This spec implements Phase 1 + Phase 3 of that proposal simultaneously, skipping VPS Phase 2 (not needed since ibeco.me is already always-on).
- **[Becoming App](scripts/plans/06_becoming-app.md)** — ibeco.me gains brain relay capability alongside its existing practice/task/memorization features. Same backend, same deploy.
- **[Intent Engineering](docs/work-with-ai/04_intent-engineering.md)** — The brain's classifier system is a concrete implementation of intent engineering: the system prompt encodes *what kind of thinking to do*, not just *what to do*.

---

## Review Checkpoint

Before implementation: Michael reviews this spec, confirms architecture, answers open questions. Then we go.

*"Created all things spiritually, before they were naturally upon the face of the earth."* — [Moses 3:5](gospel-library/eng/scriptures/pgp/moses/3.md)
