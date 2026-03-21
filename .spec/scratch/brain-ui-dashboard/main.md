# Scratch: Brain UI Dashboard

**Binding problem:** brain.exe has a web frontend (Vue 3 + Tailwind, embedded via go:embed) with 4 views (Capture, Entries, EntryDetail, Search) but no visibility into agent activity, no approval queue for routing, and no way to stop the brain from the UI. Michael is hesitant to start the brain because he can't see or control what it's doing once running.

---

## Inventory: Current Frontend

**Stack:** Vue 3 + TypeScript + vue-router + TailwindCSS 4 + Vite
**Source:** `scripts/brain/frontend/src/`
**Build output:** `scripts/brain/cmd/brain/dist/` (embedded via `//go:embed all:dist`)
**No auth** — trusts localhost / local network

### Existing Views
1. **CaptureView** (`/`) — textarea + stats bar + recent 10 entries
2. **EntriesView** (`/entries`) — category filter tabs + entry list cards
3. **EntryDetailView** (`/entries/:id`) — full entry view/edit + subtasks + classify button
4. **SearchView** (`/search`) — full-text + semantic search toggle

### Navigation
- Sticky top nav: 🧠 Brain | Capture | Entries | Search | "{N} thoughts"
- App.vue is the shell: nav + RouterView
- max-w-4xl centered layout, dark theme (gray-950 bg, gray-200 text)

### API Client (`api.ts`)
- `request<T>(path, init)` — generic fetch wrapper with error handling
- All calls go to `/api/*` (relative, same origin)
- Types: `Entry`, `SubTask`, `Stats`, `SearchResult`
- **Agent endpoints NOT in api.ts yet** — no client methods for agent/routable, agent/route, agent/sessions

### Entry Type (TypeScript side)
Missing from api.ts: `agent_route`, `route_status`, `agent_output`, `tokens_used` — these were added to the Go struct but not exposed to the frontend types yet.

---

## Inventory: Existing Backend Endpoints (Agent-Related)

| Endpoint | Method | Input | Output | Notes |
|----------|--------|-------|--------|-------|
| `/api/agent/sessions` | GET | — | `{sessions: string[]}` | Active agent names |
| `/api/agent/routable` | GET | — | `{entries: [{id, title, category, agent_name}]}` | Entries eligible for routing (suggest mode) |
| `/api/agent/route` | POST | `{entry_id}` | `{status: "routed", agent, entry_id}` | Triggers background agent work |
| `/api/agent/ask` | POST | `{prompt, agent?}` | `{response}` | Direct agent interaction |
| `/api/agent/reset` | POST | `{agent?}` | `{status, reset}` | Reset agent session(s) |
| `/api/brain/status` | GET | — | `{agent_online, model, total_entries, categories}` | Health check |

### What's Missing from Backend
1. **No shutdown endpoint** — signal (SIGINT/SIGTERM) only. Server has `Shutdown(ctx)` method but nothing triggers it from HTTP.
2. **No cancel/kill for running agent tasks** — goroutine uses `context.Background()`, no cancellation path
3. **No entry-level route status in list responses** — `handleAgentRoutable` returns a flat list, but doesn't include route_status for already-routed entries
4. **Classifier doesn't auto-annotate routing** — classify handler doesn't call `SetAgentRoute` on classified entries
5. **No "running tasks" list** — `/api/agent/sessions` shows which agents have sessions, but not what entries they're currently working on

---

## Gap Analysis

### What Michael Wants

1. **Dashboard** — see at a glance what the brain is doing
   - Active agent sessions (which agents are live)
   - Currently running tasks (entry X being processed by study agent)
   - Recent completions/failures
   - System health (DB status, vector store, relay, models)

2. **Approval Queue** — entries classified as routable, awaiting Michael's click to route
   - List of entries with suggested agent
   - "Route" button per entry
   - Ability to skip/dismiss (don't route this one)
   - See the entry body before approving

3. **Kill Switch** — safely stop the brain from the UI
   - Graceful shutdown (finish current writes, close connections)
   - Cancel running agent tasks
   - Visual confirmation that shutdown is happening

### What Exists vs What's Needed

| Feature | Backend | Frontend | Gap |
|---------|---------|----------|-----|
| Approval queue | `GET /api/agent/routable` ✅ | Nothing | Frontend view needed |
| Route an entry | `POST /api/agent/route` ✅ | Nothing | Button in approval queue |
| Active sessions | `GET /api/agent/sessions` ✅ | Nothing | Dashboard widget |
| Running tasks status | Partial — route_status on entry | Nothing | Need "running" list endpoint |
| System health | `GET /api/brain/status` ✅ | Nothing | Dashboard widget |
| Shutdown | Server.Shutdown() exists | Nothing | New endpoint + UI button |
| Cancel agent task | Not implemented | Nothing | New endpoint + context.WithCancel |
| Dismiss from queue | Not implemented | Nothing | New endpoint or status update |

---

## Design Decisions

### New Views
- **DashboardView** (`/dashboard`) — system health + active work + recent activity
- Modify **EntriesView** or EntryDetailView to show routing status inline
- **Approval queue** — could be a tab in EntriesView OR a section in DashboardView

### Backend Additions Needed

1. **`POST /api/shutdown`** — triggers `Server.Shutdown()` + signal to main goroutine
2. **`GET /api/agent/running`** — list entries with route_status = "running" or "pending"
3. **`POST /api/agent/cancel`** — cancel a running agent task (requires context.WithCancel tracking)
4. **`POST /api/entries/{id}/dismiss-route`** — set route_status to "dismissed" to skip routing
5. **Update classify handler** — auto-set `agent_route` + `route_status = "suggested"` when classifier matches a routable category

### Kill Switch Design
- Shutdown needs to: cancel running agent contexts → close HTTP server → push git archive → clean exit
- Frontend: button that calls `/api/shutdown`, shows "shutting down..." state, then detects connection loss
- Consider: should kill switch cancel running agents immediately or wait for completion?
  - **Recommendation:** Cancel immediately. Agent work can be re-run. Data safety > completion.
  - The goroutine in handleAgentRoute uses `context.Background()` — needs to switch to a tracked, cancellable context

### Approval Queue Design
- Shows entries from `GET /api/agent/routable`
- Each card: title, category, body preview, suggested agent, "Route ✓" button, "Skip ✗" button
- After routing: card moves to "Pending/Running" section or disappears
- Polling interval: every 10-15 seconds (not WebSocket for now — KISS)

### Dashboard Design
- **System section:** brain status (online, model info), entry counts, vector store status
- **Active work section:** running agent tasks with entry title + agent name + elapsed time
- **Approval section:** routable entries count + link to full queue
- **Recent activity:** last 5 completed routes with status + output preview

---

## Scope Assessment

### Phase 1 (this proposal): Dashboard + Approval Queue + Kill Switch
Backend:
- 1 new endpoint: `POST /api/shutdown`
- 1 new endpoint: `GET /api/agent/running`
- 1 new endpoint: `POST /api/entries/{id}/dismiss-route`
- 1 handler update: classify → auto-annotate routing
- 1 goroutine fix: tracked contexts for agent tasks

Frontend:
- 1 new view: DashboardView (`/dashboard`)
- 1 new nav link
- api.ts additions: agent endpoint methods + new Entry fields
- Polling for live updates (setInterval)

### Phase 2 (deferred): Agent cancel + richer status
- Cancel running agent tasks
- WebSocket for real-time updates
- Agent output preview in dashboard
- Token usage tracking display

---

## Prior Art Check
- No existing proposals for brain web UI beyond the embedded SPA
- `scripts/brain-app/` (Flutter) handles mobile but no agent routing UI
- ibeco.me (`scripts/becoming/frontend/`) has a Vue 3 frontend but different domain (practices, not brain entries)
- Squad learnings: approval gates are Pattern A2 (already in routing table design)
