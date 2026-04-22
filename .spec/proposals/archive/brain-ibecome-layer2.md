# Layer 2: Bidirectional Brain ↔ ibecome Sync

> **SHIPPED via WebSocket implementation** (status not marked at the time — confirmed 2026-04-21 during cleanup). The bidirectional sync described here was built using the WebSocket relay path. Archived without further changes.

> Design plan — not code yet. Review and decide before building.

## Problem

Layer 1 (shipped `11b9029`) creates ibecome tasks when brain classifies a thought as `actions` or `projects`, and saves the `ibecome_task_id` back on the brain entry. But the link is one-way: brain → ibecome. If you complete a task in ibecome (phone, web), brain doesn't know. If brain updates an entry, ibecome doesn't know.

**Goal:** When either side changes, the other side learns about it — without polling.

---

## Architecture Options

### Option A: Relay WebSocket Message (Recommended)

The relay hub already sits between both sides. Add a new message type that flows **ibecome → hub → brain**.

**Flow: Task completed in ibecome UI**
```
Phone/Web → PUT /api/tasks/{id} (status: completed)
  → updateTask handler notices status changed
  → Hub sends WebSocket message to agent:
      { "type": "task_updated", "task_id": 42, "status": "completed", "brain_entry_id": "abc-123" }
  → brain.exe receives it, updates entry:
      entries.status = "done", entries.action_done = 1
```

**Flow: Brain entry updated (e.g., reclassified via `fix`)**
```
brain.exe reclassifies entry
  → PATCH /api/tasks/{id} with new title/description/status
  → or: send "entry_updated" message through relay → ibecome hub → update task
```

**Pros:**
- Real-time (already connected via WebSocket)
- No new infrastructure — uses existing relay
- Works when brain is behind NAT (no inbound connections needed)
- Natural extension of existing message types

**Cons:**
- Requires `brain_entry_id` column on ibecome's tasks table (one migration)
- Hub becomes slightly task-aware (routes one new message type)
- Sync only works when brain.exe is online (but relay queues, so it catches up)

### Option B: Brain Polls ibecome REST API

brain.exe periodically calls `GET /api/tasks?status=completed` and cross-references with its stored `ibecome_task_id` values.

**Pros:** Simple, no hub changes.
**Cons:** Not real-time, wasteful, requires brain to maintain poll loop, doesn't push brain→ibecome updates.

### Option C: Webhooks (ibecome → brain)

ibecome calls a webhook on brain.exe when tasks change.

**Pros:** Standard pattern.
**Cons:** brain.exe is behind NAT. Would need ngrok or similar. Fragile. Overkill.

---

## Recommended: Option A (Relay WebSocket)

### Changes Required

#### 1. ibecome: Add `brain_entry_id` to tasks table

**File:** `scripts/becoming/internal/db/schema.sql` (add to CREATE TABLE)
**Migration:** `scripts/becoming/internal/db/db.go` — ALTER TABLE, same pattern as other ad-hoc migrations

```sql
ALTER TABLE tasks ADD COLUMN brain_entry_id TEXT;
```

This lets ibecome look up which brain entry a task came from. Currently brain saves the ibecome task ID on its side (`entries.ibecome_task_id`), but ibecome doesn't know the brain entry ID.

**Also update:** `scripts/brain/internal/ibecome/client.go` — include brain entry ID in the `POST /api/tasks` request body so both sides have the link from creation.

#### 2. ibecome: Add `BrainEntryID` to Task struct and API

**File:** `scripts/becoming/internal/db/tasks.go`
- Add `BrainEntryID string` to `Task` struct
- Update `CreateTask`, `ListTasks`, `UpdateTask` queries

**File:** `scripts/becoming/internal/api/router.go`
- `updateTask` handler: after successful update, check if task has a `brain_entry_id` and if status changed → send `task_updated` message through hub

#### 3. New message type: `task_updated`

**File:** `scripts/becoming/internal/brain/messages.go`

```go
const TypeTaskUpdated = "task_updated"

type TaskUpdatedMessage struct {
    Type         string `json:"type"`          // "task_updated"
    TaskID       int64  `json:"task_id"`       // ibecome task ID
    BrainEntryID string `json:"brain_entry_id"` // brain entry UUID
    Status       string `json:"status"`        // new status
    Title        string `json:"title"`         // current title (in case it changed)
}
```

**Direction:** Server → Agent (routed via `routeToAgent` with queue support)

#### 4. Hub: Route `task_updated` to agent

**File:** `scripts/becoming/internal/brain/hub.go`

The hub already has `routeMessage` with a switch on type. Add:

```go
case TypeTaskUpdated:
    // Server-initiated (from API handler), route to agent
    // This doesn't come from a WebSocket client — it's injected by the API
    // Need a Hub method: SendToAgent(userID, data)
```

**Key design question:** The `updateTask` HTTP handler needs access to the Hub to inject messages. Options:
- Pass `*Hub` to the router (simplest — hub is already a singleton)
- Use a channel/callback that the hub listens on
- Event bus (overkill)

**Recommendation:** Pass `*Hub` to router. Add `Hub.NotifyAgent(userID int64, data []byte)` that calls `routeToAgent`.

#### 5. brain.exe: Handle `task_updated` message

**File:** `scripts/brain/internal/relay/client.go`

In the message handler loop, add a case for `"task_updated"`:

```go
case "task_updated":
    var msg TaskUpdatedMessage
    json.Unmarshal(data, &msg)
    // Map ibecome status → brain status
    // "completed" → set action_done=1, status="done"
    // "paused"    → status="waiting"
    // "active"    → status="" (clear), action_done=0
    c.store.UpdateEntryStatus(msg.BrainEntryID, mappedStatus)
```

**File:** `scripts/brain/internal/store/db.go`
- Add `UpdateEntryStatus(entryID string, status string, actionDone bool)` method

#### 6. brain.exe → ibecome updates (optional, Phase 2b)

When brain reclassifies an entry via `fix`:
- If entry has `ibecome_task_id`, call `PUT /api/tasks/{id}` with updated title/category
- Add to `ibecome.Client`: `UpdateTask(ctx, taskID, updates)`

This is lower priority — brain-initiated changes are less common than completing tasks in the app.

---

## Status Mapping

| ibecome status | brain entry fields |
|---|---|
| `active` | `status = ""`, `action_done = 0` |
| `completed` | `status = "done"`, `action_done = 1` |
| `paused` | `status = "waiting"` |
| `archived` | `status = "archived"` |

| brain status | ibecome status |
|---|---|
| `"done"` or `action_done = 1` | `completed` |
| `"waiting"` / `"blocked"` | `paused` |
| (empty / active) | `active` |

---

## Execution Order

1. **ibecome: `brain_entry_id` column + migration** — Schema change, low risk
2. **brain: Send `brain_entry_id` in task creation** — Small change to `client.go`
3. **ibecome: `TaskUpdatedMessage` type** — New message struct
4. **ibecome: `Hub.NotifyAgent` method** — Hub API for server-side message injection
5. **ibecome: Wire `updateTask` handler → Hub** — The trigger
6. **brain: Handle `task_updated`** — Receive and apply
7. **brain: `UpdateEntryStatus` in store** — DB write
8. **(Optional) brain: Push entry updates to ibecome** — Reclassification sync

Steps 1-2 can ship independently (backward compatible). Steps 3-7 ship together as the sync feature.

---

## Open Questions

- **Conflict resolution:** If both sides change at the same time, who wins? Proposal: last-write-wins with timestamp comparison. Simple and good enough for a single-user system.
- **Delete sync:** If a task is deleted in ibecome, should the brain entry be deleted too? Probably not — the entry still has value as a classified thought. Just clear `ibecome_task_id`.
- **Bulk sync on reconnect:** When brain.exe comes online, should it fetch all task statuses? The relay queue handles messages sent while offline, but doesn't handle changes that happened *outside* the relay (direct API calls from the web UI). A one-time reconciliation on connect might be useful.
- **MCP tool for task links:** Should `mcp_becoming_list_tasks` show the brain entry ID? Could be useful for Copilot to cross-reference.

---

## Not In Scope

- Real-time UI updates (ibecome frontend refresh when brain creates a task) — that's a separate WebSocket-to-frontend feature
- Multi-user task sharing — single user for now
- Task comments or conversation threads — future feature
