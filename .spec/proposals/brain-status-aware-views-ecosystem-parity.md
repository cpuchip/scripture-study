---
workstream: WS2
status: proposed
brain_project: 6
created: 2026-04-23
binding_problem: |
  brain.exe (desktop) now hides someday/archived entries by default after 04-23 shipping.
  But the brain ecosystem has three surfaces: brain.exe, ibeco.me (web cache), and brain-app (Flutter mobile).
  The other two still show parked entries with no filter. After today's audit set 49 entries to parked,
  the website TasksView and the phone history both still display them — making the desktop cleanup invisible
  on every other device. The triage decision should propagate to every surface that consumes the data.
related_brain_entries: []
sister_proposals:
  - brain-status-aware-views.md
  - brain-status-field-on-list-queries.md
predecessor:
  - brain-status-aware-views.md (Phases 1-3 + Dashboard, shipped 04-23 — desktop only)
---

# Brain — Status-Aware Views: Ecosystem Parity

## Implementation Status

| Phase | Surface | Codebase | Effort | Status |
|-------|---------|----------|--------|--------|
| 1 | ibeco.me web TasksView | `scripts/becoming/` | ~30 min | proposed |
| 2 | brain-app history screen | `scripts/brain-app/` | ~20 min | proposed |
| 3 | (verify) Migration / data integrity | both | ~10 min | proposed |

## Background

The 04-23 session shipped status-aware filtering on brain.exe across three layers (data, server-side `/api/entries` filter, three Dashboard agent surfaces, project-board toggle). Audit afterward found two adjacent surfaces consume the same data without parity.

| Surface | Endpoint consumed | Filter today | Parked visibility |
|---------|------------------|--------------|-------------------|
| brain.exe Capture/Entries | `/api/entries` (local) | ✅ defaults exclude parked, opt-in via `?include_parked=1` | hidden by default |
| brain.exe Dashboard agent surfaces | `/agent/routable`, `/agent/review`, `/entries/your-turn` | ✅ server-side, no toggle | always hidden (work surface) |
| brain.exe Project board | `/api/projects/{id}/entries` | ✅ client filter + visible header checkbox | toggle-able |
| **ibeco.me TasksView** | `/api/brain/entries` (cached on web) | ❌ none | **always visible** |
| **brain-app history** | local entries cache via WebSocket `entries_request` | partial (`_showArchived` only, not `someday`) | someday always visible |

## Phase 1 — ibeco.me parity

### Goal
The web brain view honors the same triage as desktop: parked hidden by default, opt-in to see them.

### Files

- [scripts/becoming/internal/db/brain_entries.go](../../scripts/becoming/internal/db/brain_entries.go) — `ListBrainEntries(userID, category)`
- [scripts/becoming/internal/brain/hub.go](../../scripts/becoming/internal/brain/hub.go) — `HandleBrainEntries` (~L683)
- [scripts/becoming/frontend/src/api.ts](../../scripts/becoming/frontend/src/api.ts) — `listBrainEntries`
- [scripts/becoming/frontend/src/views/TasksView.vue](../../scripts/becoming/frontend/src/views/TasksView.vue)

### Changes

1. **Extend `ListBrainEntries` signature** to accept an `includeParked bool`:
   ```go
   func (db *DB) ListBrainEntries(userID int64, category string, includeParked bool) ([]*BrainEntry, error)
   ```
   When `!includeParked`, append `AND (status IS NULL OR status NOT IN ('someday','archived'))` to the WHERE clause.
   Audit other callers (`handleEntriesRequest` at hub.go:658 — pass `true` since the agent wants the full picture for app sync; the app filters locally).

2. **`HandleBrainEntries`**: read `include_parked` query param (matching brain.exe convention `?include_parked=1` or `=true`), pass to `ListBrainEntries`. Default `false`.

3. **`api.ts` + `TasksView.vue`**: optional toggle in the brain tab header (mirror desktop's project-board checkbox pattern). Pass `?include_parked=1` when checked. Show parked count in the label ("Show {{ parkedCount }} parked").

4. **Schema check**: confirm `brain_entries` table already has `status` column (it does — verified in `BrainEntry` struct + INSERT). No migration needed.

### Verification (Phase 1)
```powershell
# After deploy:
curl -H "Cookie: ..." https://ibeco.me/api/brain/entries | jq '.entries | length'        # default (filtered)
curl -H "Cookie: ..." "https://ibeco.me/api/brain/entries?include_parked=1" | jq '.entries | length'  # all
# Difference should equal count of someday+archived entries cached for that user.
```

Inverse hypothesis: temporarily revert filter → verify count increases by ~49 → reapply → verify drops back. (Only do locally if running becoming server.)

## Phase 2 — brain-app history screen parity

### Goal
Mobile history honors the same triage. Single "Show parked" toggle replaces the existing archive-only toggle, covering both `someday` and `archived`.

### Files

- [scripts/brain-app/lib/screens/history_screen.dart](../../scripts/brain-app/lib/screens/history_screen.dart) — `_showArchived` state + filter logic ~L33, L99-L101

### Changes

1. Rename state: `_showArchived` → `_showParked` (or add new `_showParked` and remove `_showArchived` entirely — they conceptually merge).
2. Update filter:
   ```dart
   // Before:
   if (!_showArchived && e.status == 'archived') return false;
   if (_showArchived && e.status != 'archived') return false;
   // After:
   final isParked = e.status == 'archived' || e.status == 'someday';
   if (!_showParked && isParked) return false;
   if (_showParked && !isParked) return false;
   ```
3. Update toggle label: "Show archived" → "Show parked" (covers both verbs).
4. Update parked count badge to count `someday + archived` entries.

### Verification (Phase 2)
- Build: `flutter build` succeeds
- Manual on phone or emulator: with toggle off, neither someday nor archived appear; with toggle on, only parked appear; count badge accurate

### Out of scope for Phase 2
- Source of truth for the history list is the app's local cache via WebSocket `entries_request` (not `/api/entries`), so the API default change does not affect this view. No API contract change needed for mobile.
- Did not investigate whether the app has any "agent work surface" view (analog to Dashboard routable/review). If found later, treat as a separate phase.

## Phase 3 — Verification & data integrity

1. **Cross-surface smoke test**: pick one specific entry (e.g. one we set to `someday` in the 04-22 audit). Verify it is hidden by default on:
   - brain.exe Capture, Entries, Dashboard ✅ (already verified 04-23)
   - ibeco.me TasksView (Phase 1)
   - brain-app history (Phase 2)
   And visible in all three when respective parked toggles are on.

2. **Sync direction check**: after a desktop status change to `someday`, does the change reach ibeco.me's cache promptly? (It should via existing `entries_sync` WebSocket flow — `BulkUpsertBrainEntries` already updates status on conflict.) Manually verify one round-trip.

3. **No-toggle assertion for agent surfaces**: confirm Phase 1 does NOT introduce parked into the WebSocket `entries_request` flow that brain-app relies on for its full local cache. Mobile filters locally, so the cache should remain complete.

## Costs & risks

| Risk | Mitigation |
|------|-----------|
| ibeco.me caller signature change cascades | Only one prod caller besides the new handler (`handleEntriesRequest`) — passes `true` to preserve current behavior. |
| Mobile users on stale app version | Filter is client-side, so old apps continue to work as today (no regression). They just won't get the parked filter until they update. |
| Web frontend dist files are committed | After Phase 1 frontend changes, must run `npm run build` and commit dist (existing convention — see prior frontend commits). |

## Phasing notes

- **Phase 1 first** because it has the highest visibility (the website is what others might see) and the bug class is identical to what we just fixed — pattern is fresh.
- **Phase 2 can lag** — affects only Michael's phone, lower urgency, can wait until next time brain-app is touched.
- **Phase 3 is a 10-minute manual check** that closes the loop — not optional, but cheap.

## Success criteria

A 50-entry triage on brain.exe (status=someday) reduces the visible count by 50 on:
- ✅ brain.exe (already shipping)
- ⏳ ibeco.me TasksView brain tab
- ⏳ brain-app history screen

with an opt-in toggle on each surface to show what was hidden.

## Decision log

- **Why merge `someday` and `archived` into a single "parked" toggle on mobile?** They share the same UX semantic ("not currently in scope, don't show me"). The current `_showArchived`-only toggle was a partial implementation from before `someday` existed as a status. One toggle is simpler and matches desktop terminology ("Show N parked").
- **Why no migration?** All three layers already have the `status` column. This is purely a query-layer + UI change.
- **Why opt-in (default off) instead of opt-out?** Matches brain.exe convention shipped 04-23 and Mosiah 4:27 principle: parked entries are noise by definition; the default should respect the triage.
