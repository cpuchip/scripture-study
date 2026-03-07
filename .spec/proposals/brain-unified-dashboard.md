# Brain Unified Dashboard

**Status:** In Progress (Phase 4 next)
**Created:** 2026-03-06
**Phase 1:** ✅ Committed `dbd1a86` (brain)
**Phase 2:** ✅ Committed `c6b21ce` (brain) + `b677f35` (brain-app)
**Phase 3:** ✅ Committed `8ed08a8` (brain) + `5b0d911` (scripture-study)
**Problem:** Three surfaces (brain-app, brain web UI, ibeco.me tasks) show different data with different capabilities. No single place to manage all your thoughts.

## Problem Analysis

### Current State: Three Surfaces, Three Stories

| Surface | Data Source | Shows | Can Create | Can Edit | Can Complete | Can Delete |
|---------|-----------|-------|------------|----------|-------------|------------|
| **Brain-app** (Flutter) | Relay message queue (`brain_messages`) | Old relay traffic — zombies with "Pending classification" | Yes (via WebSocket) | No | No | No |
| **Brain Web UI** (localhost:8445) | SQLite (source of truth) | All classified entries | Yes (capture form) | Category, body, tags — **not title** | No | Yes |
| **ibeco.me Tasks** | PostgreSQL (ibecome) | Only actions + projects synced from brain | Yes (manual) | No (just toggle status) | Yes (checkbox) | Yes |

### Root Cause

Three separate data stores with partial one-way syncs:
- **Relay queue** (ibeco.me `brain_messages` table) — log of WebSocket traffic, not a real data store
- **Brain SQLite** (local) — the actual source of truth for thoughts
- **ibecome PostgreSQL** (cloud) — only gets actions/projects, one-way push

The brain-app reads from the relay queue (wrong source), ibecome only sees 2 of 6 categories, and the brain web UI is the only place with real data but lacks key management features.

## Solution: Phased Unification

### Phase 1: Make Brain Web UI the Real Dashboard ← CURRENT
The brain web UI already has the source of truth (SQLite). Add what's missing:

- [ ] **Title editing** — inline edit on the entry detail page
- [ ] **Complete/Archive buttons** — use existing `status` and `action_done` fields
- [ ] **Visual feedback** on category change — the save already works, add transition/toast
- [ ] **Status indicators** — show done/active/archived state on entry cards
- [ ] Delete already works ✓

**Files to modify:**
- `scripts/brain/frontend/` (Vue SPA)
- `scripts/brain/internal/web/` (API endpoints, if status update endpoint missing)

### Phase 2: Make Brain-App Show Real Data
Replace the relay history API with direct brain communication:

- [ ] Brain-app History → calls brain.exe's REST API (`GET /api/entries`) instead of relay queue
- [ ] Brain-app can mark done, delete, edit — sends to brain.exe which syncs to ibecome
- [ ] Eliminate "Pending classification" zombie entries
- [ ] Option A: Direct HTTP to brain.exe (requires local network / VPN)
- [ ] Option B: Proxy through relay (relay forwards CRUD to brain agent)

**Key decision:** How does the phone reach brain.exe? Direct LAN or relay proxy?

### Phase 3: ibeco.me Tasks → Brain Hybrid View
When `brain_enabled` is true for a user:

- [ ] Tasks page shows **all brain entries** grouped by category (not just actions/projects)
- [ ] Native ibecome tasks still displayed (mixed view)
- [ ] All CRUD operations go through ibecome API → relay notifies brain
- [ ] Full management from any browser, even without brain.exe running
- [ ] Consider renaming "Tasks" to "Brain" when brain is enabled

**Requires:**
- New API endpoints on ibecome for brain entry CRUD
- Brain syncs all categories (not just actions/projects)
- ibecome stores full brain entry data (not just task subset)

### Phase 4: Bidirectional Full Sync
- [ ] Brain syncs **all** categories to ibecome (people, ideas, study, journal, inbox too)
- [ ] ibecome becomes the cloud backup/mirror of your brain
- [ ] Any surface can edit anything, all stay in sync via relay
- [ ] Conflict resolution: last-write-wins with timestamps (or prompt user)
- [ ] ibecome brain entries stored in a dedicated `brain_entries` table (not overloading `tasks`)

## Category Clarity

Current categories and their task-like nature:

| Category | Is it a "task"? | Should sync to ibecome? |
|----------|----------------|------------------------|
| **actions** | Yes — has due date, completable | Yes (today: `once` task) |
| **projects** | Yes — has status, next action | Yes (today: `ongoing` task) |
| **ideas** | No — it's a seed, not a commitment | Phase 3+ |
| **people** | No — it's a relationship note | Phase 3+ |
| **study** | No — it's a learning note | Phase 3+ |
| **journal** | No — it's a reflection | Phase 3+ |
| **inbox** | No — it's unclassified | Phase 3+ |

## Architecture After Phase 4

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Brain-App   │     │  Brain Web   │     │  ibeco.me    │
│  (Flutter)   │     │  (Vue SPA)   │     │  (Vue SPA)   │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │ WebSocket          │ HTTP                │ HTTP
       ▼                    ▼                     ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  ibeco.me    │◄───►│  brain.exe   │◄───►│  ibeco.me    │
│  relay hub   │     │  (local)     │     │  REST API    │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────┴───────┐
                     │   SQLite +   │
                     │  chromem-go  │
                     │ (source of   │
                     │   truth)     │
                     └──────────────┘
```

## Open Questions

1. **Phone → brain.exe connectivity**: Direct LAN, relay proxy, or both?
2. **Should ibecome get its own `brain_entries` table** or extend `tasks` with more fields?
3. **Offline support**: If brain.exe is down, should brain-app queue locally?
4. **MCP capture**: Should `brain.exe mcp` also support a `brain_capture` write tool?
