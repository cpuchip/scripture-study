# Brain — Expose `status` Field on List Queries

**Date:** 2026-04-22 (next session: 04-23)
**Status:** Proposal → Active fix
**Severity:** Critical — silent data loss in UX. Two phases of UI work shipped against a non-existent field.

## Symptom

User set entries to `status='someday'` via the new entry-detail dropdown.
Entry detail page shows the badge `someday ▾` correctly.
Project board, Project list view, Capture view, Entries view, Dashboard — entry remains visible. Filter never hides it.

## Root Cause (verified via Rule 3 — Quit Thinking and Look)

The `Entry` Go struct has a `Status` field. `GetEntry` (used by `/api/entries/{id}`) populates it correctly via a SELECT that includes `status`.

**But the LIST queries do NOT select the `status` column:**

- `ListAll(limit, offset)` at `scripts/brain/internal/store/db.go:789` — used by `/api/entries`, drives Capture/Entries/Dashboard
- `ListEntriesByProject(projectID)` at `scripts/brain/internal/store/db.go:1520` — used by `/api/projects/{id}/entries`, drives ProjectDetailView

So every entry returned by these endpoints has `status: ""` (Go zero value) → `null` in JSON. Client-side filter `e.status === 'someday'` can never match.

**Verification:**
```bash
$ curl -s "http://localhost:8445/api/projects/4/entries" | python -c "..."
22b8d8b2 | None | verified | I'm seriously looking at what it would take...
17749618 | None | raw     | Star Trek UI with Pretext      ← actual status: someday
```

## Why This Bit Us Twice

I assumed the wire format included every Entry struct field. Should have run `curl` against the API in Phase 1 of the original work. **Rule 1 — Understand the System** failure: I read the frontend code and the proposal, never the SELECT statements.

## Fix

Three SELECT queries need `status` added (also `action_done`, `due_date` for completeness — they're on `Entry` and surface in the same UI but also affected):

### 1. `ListAll` (db.go ~line 789)
Add `status, action_done, due_date` to the SELECT and scan targets.

### 2. `ListEntriesByProject` (db.go ~line 1520)
Same.

### 3. `ListCategory` (db.go ~line 749) — same bug, lower urgency
Same. Used by `/api/entries?category=…`.

## Out of Scope (separate proposals)

- **Server-side filter** (`?include_parked=false` query param) — still deferred. Once status is on the wire, the client filter works. Server filter is a perf optimization for later.
- **Hiding parked from Review queue** — Dashboard's review queue probably should respect status too, but that's a separate UX decision.
- **Other list functions** (`ListByRouteStatus`, `ListPipeline`, `ListUnassigned`) — audit later. For now, fix the three that drive the views in question.

## Verification Plan

1. Rebuild brain.exe (requires `go mod tidy` first per go.sum gap)
2. Restart server
3. `curl /api/projects/4/entries` → expect Star Trek UI entry to show `status: "someday"`
4. Refresh `/projects/4` in browser → Star Trek UI entry should disappear from Inbox lane
5. Refresh `/entries` → should show "X hidden by status" affordance
6. Click "show all" toggle → entry returns
7. Playwright-cli sequence: open `/projects/4`, snapshot, verify Star Trek UI not in main board, click toggle, verify it reappears

## Lessons (for memory)

- **Don't assume struct fields auto-flow to JSON.** Verify every API endpoint returns the field you're filtering on. `curl | jq` is faster than reading code.
- **The covenant rule "read before quoting" applies to API contracts too.** I quoted the API in my filter logic without reading what it actually returns.
- **Rule 7 — Check the Plug.** Before building elaborate filter UI, verify the data exists at the wire.
