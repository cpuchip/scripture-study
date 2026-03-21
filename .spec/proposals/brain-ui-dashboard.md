# Brain UI Dashboard — Approval Queue, Status, and Kill Switch

**Binding problem:** Michael is hesitant to start brain.exe because he has no visibility into what agents are doing, no control surface for approving routed work, and no way to safely stop the system from the same interface where he sees its activity. The backend gates exist (Phase 3a shipped suggest-mode routing) but the UI to *use* them doesn't.

**Created:** 2026-03-21
**Research:** [.spec/scratch/brain-ui-dashboard/main.md](../../scratch/brain-ui-dashboard/main.md)
**Depends on:** WS1 Phase 3a (agent pool + routing table) — SHIPPED
**Affects:** Brain startup confidence, daily brain usage
**Status:** Draft — awaiting review

---

## 1. Problem Statement

brain.exe's web frontend has 4 views (Capture, Entries, EntryDetail, Search) — all focused on entry CRUD. Phase 3a added 5 agent API endpoints for routing, but no frontend consumes them. The approval queue workflow exists in code but has no surface:

- `GET /api/agent/routable` returns entries ready for routing — **nothing renders them**
- `POST /api/agent/route` triggers agent work — **no button calls it**
- `GET /api/agent/sessions` shows active agents — **nothing displays them**
- `GET /api/brain/status` returns health info — **only the Flutter app uses it**

Additionally:
- **No shutdown mechanism** from the UI — only SIGINT/SIGTERM from the terminal
- **No way to see** if an agent is currently running or what it's working on
- **Classifier doesn't annotate routing** — classifying an entry as "study" doesn't mark it as routable

Michael's daily experience should be: open brain UI → see what's queued → approve what looks right → watch progress → hit kill switch if something goes wrong.

### Success Criteria

1. **Dashboard shows brain health** — agent sessions, running tasks, system status at a glance
2. **Approval queue works** — Michael sees routable entries and can approve or dismiss each one
3. **Kill switch works** — one click gracefully shuts down the brain, cancelling running agent tasks
4. **Classification auto-annotates** — classifying an entry as "study" automatically marks it as routable
5. **No new dependencies** — uses existing Vue 3 + Tailwind stack, existing API patterns

---

## 2. Constraints & Boundaries

**In scope:**
- One new frontend view: `DashboardView` (`/dashboard`)
- Three new backend endpoints: shutdown, running tasks, dismiss route
- One handler update: auto-annotate routing in classify
- One goroutine fix: tracked contexts for cancellable agent tasks
- api.ts additions: agent methods + updated Entry type
- Navigation update: add Dashboard link

**Out of scope:**
- WebSocket real-time updates (polling is fine for V1)
- Agent output preview / review queue (Phase 3c)
- "Needs input" status for agent roadblocks (deferred per Michael's direction)
- Token usage display (Phase 3b governance)
- brain-app (Flutter) changes
- Mobile-responsive layout optimization

**Conventions:**
- Vue 3 Composition API (`<script setup lang="ts">`)
- TailwindCSS 4 utility classes
- Dark theme (gray-950 bg, consistent with existing views)
- Same `request<T>()` pattern in api.ts
- Go 1.22 `mux.HandleFunc("METHOD /path", ...)` routing

---

## 3. Prior Art & Related Work

| Source | Relevance |
|--------|-----------|
| Phase 3a (shipped) | Backend routing endpoints — approval queue API exists |
| Phase 3 proposal | Designed suggest mode, entry status lifecycle |
| Squad learnings (A2) | Approval gates are the pattern; UI is the missing piece |
| Existing frontend | 4 views establish patterns for layout, API calls, state management |
| `handleBrainStatus` | Existing health endpoint — already returns model, counts, agent_online |
| `Server.Shutdown(ctx)` | Method exists but nothing calls it via HTTP |

---

## 4. Proposed Approach

### 4.1 Backend Changes

#### New Endpoint: `POST /api/shutdown`

Triggers graceful shutdown of the brain daemon.

```go
func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
    // Cancel all running agent contexts
    s.pool.CancelAll()
    
    jsonResponse(w, map[string]string{"status": "shutting_down"})
    
    // Signal the main goroutine to shut down
    // (send to the same signal channel used for SIGINT)
    go func() {
        time.Sleep(500 * time.Millisecond) // let response flush
        s.shutdownCh <- struct{}{}
    }()
}
```

**Requires:** A shutdown channel passed from main.go to Server, replacing the direct SIGINT wait. The Server gets `shutdownCh chan<- struct{}` and the main function selects on both the OS signal and this channel.

#### New Endpoint: `GET /api/agent/running`

Returns entries currently being processed by agents.

```go
// Response: {entries: [{id, title, category, agent_name, route_status, started_at}]}
```

Implementation: query entries where `route_status IN ('pending', 'running')`. Simple DB query on existing columns. Could include elapsed time if we add a `route_started_at` column (nice-to-have).

#### New Endpoint: `POST /api/entries/{id}/dismiss-route`

Marks an entry as "dismissed" so it won't appear in the approval queue.

```go
func (s *Server) handleDismissRoute(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    err := s.store.UpdateRouteStatus(id, "dismissed")
    // ...
}
```

Add `RouteStatusDismissed = "dismissed"` to router.go constants. The `handleAgentRoutable` already filters by status — dismissed entries won't appear because the filter skips anything with a non-empty status.

#### Handler Update: Auto-Annotate in Classify

In `handleClassify`, after successful classification, check if the entry's category maps to a route:

```go
// After classification succeeds and entry is saved:
route := ai.LookupRoute(entry.Category)
if route.AgentName != "" && route.Mode != ai.RouteModeNone {
    _ = s.store.SetAgentRoute(entry.ID, route.AgentName, ai.RouteStatusSuggested)
}
```

This is the one-liner that connects classification to routing. Without it, entries are classified but never appear in the approval queue.

#### Goroutine Fix: Tracked Contexts

Currently `handleAgentRoute` uses `context.Background()` for agent work — uncancellable. Change to:

```go
// In AgentPool: track running contexts
type runningTask struct {
    entryID string
    cancel  context.CancelFunc
}

// Store running tasks in pool, keyed by entry ID
// CancelAll() iterates and cancels each
```

The pool gains `StartTask(entryID) context.Context` and `CancelTask(entryID)` and `CancelAll()` methods. `handleAgentRoute`'s goroutine uses the tracked context instead of `context.Background()`.

### 4.2 Frontend Changes

#### Updated `api.ts`

Add the agent routing fields to the Entry type and new API methods:

```typescript
// Add to Entry interface:
agent_route?: string
route_status?: string
agent_output?: string
tokens_used?: number

// New methods:
agentSessions(): Promise<{sessions: string[]}>
agentRoutable(): Promise<{entries: RoutableEntry[]}>
agentRoute(entryId: string): Promise<{status: string, agent: string, entry_id: string}>
agentRunning(): Promise<{entries: RunningEntry[]}>
dismissRoute(entryId: string): Promise<void>
shutdown(): Promise<{status: string}>
brainStatus(): Promise<BrainStatus>
```

#### New View: `DashboardView.vue` (`/dashboard`)

Three sections in a single view:

**Section 1: System Status (top)**
- 🧠 Brain status badge (online/offline)
- Model info (from `/api/brain/status`)
- Entry counts by category (from `/api/stats`)
- Active agent sessions (from `/api/agent/sessions`)
- 🛑 Kill Switch button (red, top-right corner, calls `POST /api/shutdown`)

**Section 2: Active Work (middle)**
- Cards for entries with route_status = "running" or "pending"
- Each card shows: entry title, agent name, status badge
- Auto-refreshes via polling (10-second interval)
- Empty state: "No active agent work" with muted text

**Section 3: Approval Queue (bottom, primary focus)**
- Cards for entries from `GET /api/agent/routable`
- Each card shows:
  - Entry title (clickable → EntryDetailView)
  - Category badge
  - Body preview (first ~100 chars)
  - Suggested agent name
  - ✓ "Route" button (green) — calls `POST /api/agent/route`
  - ✗ "Skip" button (muted) — calls `POST /api/entries/{id}/dismiss-route`
- After routing: card disappears (moves to Active Work section on next poll)
- Empty state: "No entries waiting for approval" — the good state

**Polling strategy:**
- `onMounted`: load all three sections
- `setInterval(15000)`: refresh all three sections
- After user action (route/dismiss): immediate refresh

**Kill switch UX:**
- Confirmation dialog: "Shut down the brain? Running agent tasks will be cancelled."
- On confirm: call `POST /api/shutdown`
- Show "Shutting down..." overlay
- Detect connection loss → show "Brain stopped" with no retry

#### Navigation Update

Add "Dashboard" to the nav bar between "Capture" and "Entries":

```html
<RouterLink to="/dashboard" ...>Dashboard</RouterLink>
```

Consider making Dashboard the default landing page (replace `/` → `/dashboard`, move Capture to `/capture`). Decision: **keep Capture as `/`** — it's the fastest path for the primary action. Dashboard is a monitoring view, not the default action.

### 4.3 File Inventory

**Backend (Go):**
| File | Change |
|------|--------|
| `internal/web/server.go` | 3 new handlers + routes + shutdownCh field on Server |
| `internal/web/server.go` | Update handleClassify to auto-annotate routing |
| `internal/ai/pool.go` | Add task tracking: StartTask, CancelTask, CancelAll |
| `internal/ai/router.go` | Add RouteStatusDismissed constant |
| `cmd/brain/main.go` | Pass shutdown channel to Server, select on it alongside OS signals |

**Frontend (Vue/TypeScript):**
| File | Change |
|------|--------|
| `frontend/src/api.ts` | Add agent methods + Entry fields + new types |
| `frontend/src/views/DashboardView.vue` | New file — the whole dashboard |
| `frontend/src/main.ts` | Add `/dashboard` route |
| `frontend/src/App.vue` | Add Dashboard nav link |

---

## 5. Phased Delivery

### Phase 1: Wiring (backend only, 1 session)
1. Add `RouteStatusDismissed` to router.go
2. Add `POST /api/shutdown` endpoint + shutdown channel plumbing
3. Add `GET /api/agent/running` endpoint (query entries by route_status)
4. Add `POST /api/entries/{id}/dismiss-route` endpoint
5. Update `handleClassify` to auto-annotate routing
6. Add task tracking to AgentPool (StartTask/CancelTask/CancelAll)
7. Update `handleAgentRoute` to use tracked context
8. Tests for new endpoints

### Phase 2: Frontend (1 session)
1. Update api.ts with new types and methods
2. Create DashboardView.vue with all three sections
3. Add route to main.ts
4. Add nav link to App.vue
5. Build frontend (`npm run build`)
6. Copy dist to cmd/brain/dist
7. Test end-to-end: start brain → classify entry → see in approval queue → route → see active → complete

---

## 6. Verification Criteria

| Criterion | How to verify |
|-----------|--------------|
| Dashboard shows brain status | Start brain → navigate to /dashboard → see "online" status |
| Approval queue populates | Create entry → classify → refresh dashboard → entry appears in queue |
| Route button works | Click "Route" on queued entry → entry moves to Active Work |
| Skip button works | Click "Skip" → entry disappears from queue, doesn't return |
| Running tasks visible | Route an entry → Active Work shows it with "running" badge |
| Kill switch works | Click kill switch → confirm → brain stops → "Brain stopped" message |
| Auto-annotate works | POST entry → POST classify → entry.route_status = "suggested" |
| Polling refreshes | Wait 15 seconds → dashboard updates without manual refresh |

---

## 7. Costs & Risks

**Costs:**
- ~2 sessions of development work (1 backend, 1 frontend)
- Frontend rebuild needed after changes (existing friction, not new)
- 3 new API endpoints to maintain
- Shutdown endpoint is a power operation — no auth currently (trusted localhost only)

**Risks:**
- **Shutdown endpoint has no auth.** On localhost this is fine. If brain.exe ever runs on a server, this needs protection. Mitigation: document as localhost-only; add auth later as part of deployment hardening.
- **Context cancellation may leave agent sessions in weird state.** The Copilot SDK may not handle context cancellation cleanly. Mitigation: call `pool.Reset(agentName)` after cancellation to clean up.
- **Polling creates steady API traffic.** Every 15 seconds × 3 endpoints = mild load. On localhost, negligible. Mitigation: only poll when dashboard tab is active (use document visibility API).

---

## 8. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | Michael can't start the brain without visibility and control |
| Covenant | Rules of engagement? | Same frontend patterns, dark theme, no new deps, suggest mode |
| Stewardship | Who owns what? | dev agent executes; Michael reviews output and frontend UX |
| Spiritual Creation | Is the spec precise enough? | Yes — file list, endpoint shapes, UI sections all specified |
| Line upon Line | What's the phasing? | Phase 1 (backend wiring) stands alone — endpoints work via curl |
| Physical Creation | Who executes? | dev agent for backend; dev or manual for frontend Vue work |
| Review | How do we know it's right? | Verification criteria table — 8 observable tests |
| Atonement | What if it goes wrong? | Shutdown is graceful; cancelled agents can be re-run; git archive pushes on exit |
| Sabbath | When do we stop and reflect? | After Phase 1 (test endpoints), after Phase 2 (test full flow) |
| Consecration | Who benefits? | Michael directly; pattern reusable for any agent-augmented tool |
| Zion | How does this serve the whole? | Unblocks brain startup → enables daily study routing → feeds the intent |

---

## 9. Recommendation

**Build it.** Two sessions. Phase 1 (backend) can ship independently — endpoints work via curl/Postman even without the frontend. Phase 2 (frontend) makes it human-friendly.

This is the bridge between "Phase 3a shipped the routing infrastructure" and "Michael actually uses it." Without this UI, the routing table is infrastructure that sits idle.
