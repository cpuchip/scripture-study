# Scratch: brain-status-aware-views — findings

**Binding problem:** Brain's `status` field is write-only — none of the views filter or hide entries by status. After the 2026-04-22 audit set 26 entries to `someday` and 23 to `archived`, the UI looks identical. The cleanup is invisible.

## Evidence (read 2026-04-22)

### Schema (private-brain/brain.db, entries table)
- `status` column: takes values `active`, `someday`, `roadmap`, `waiting`, `done`, `archived`, NULL (per brain-manual-stage-transitions.md)
- Post-audit distribution: done:29, active:29, someday:27, archived:23, NULL:5, roadmap:2

### Frontend audit — three views matter

**`scripts/brain/frontend/src/views/CaptureView.vue` (line 33):**
```ts
api.listEntries({ limit: 10 })
```
No status filter. No category filter. Shows the 10 most recent entries regardless of state. This is the homepage. This is where the noise lives.

**`scripts/brain/frontend/src/views/EntriesView.vue` (line 171):**
```vue
<span v-if="entry.status" class="text-xs px-2 py-0.5 rounded-full bg-gray-800 text-amber-400">
  {{ entry.status }}
</span>
```
Status is rendered as a badge but never filtered. Done entries get a strikethrough only when `category === 'projects' && status === 'done'`. `someday` and `archived` get no visual treatment beyond the badge.

**`scripts/brain/frontend/src/views/ProjectDetailView.vue` (line 120-141):**
```ts
// 3-column board: Inbox / Working / Done
for (const e of entries.value) {
  if (e.notebook || !e.maturity || e.maturity === 'raw') {
    inbox.push(e)
  } else if (e.maturity === 'verified' || e.maturity === 'complete') {
    done.push(e)
  } else {
    working.push(e)
  }
}
```
Board uses `maturity` and the `notebook` flag. `status` is never consulted. So a Star Trek UI entry set to `status=someday` STILL appears in the project Inbox column — exactly what Michael saw with `17749618`.

### API layer

`api.ts` line 239:
```ts
listEntries(params?: { category?: string; limit?: number; offset?: number; needs_review?: boolean; unassigned?: boolean })
```

No `status` param. No `include_archived` param. The HTTP endpoint may or may not support it server-side — needs verification in `scripts/brain/internal/server/handlers.go`.

### The "duplicate" diagnosis

Michael flagged: "I'm seriously looking at what it would take to do a science..." appears in both Space Center board AND Capture inbox. DB query confirms ONE row only (`22b8d8b2`, project_id=4, status=active, maturity probably=`verified` based on board placement). The entry shows in both because:
- Space Center board (`/projects/4`) lists all entries with project_id=4
- Capture "Recent" lists last 10 regardless of project assignment

**Not a duplicate. Two views overlapping.** The fix is the same as the status-filter fix — Capture's Recent should be scoped to "unprocessed" not "newest."

## Sister proposals

- `brain-manual-stage-transitions.md` — write side. Adds the UI to MOVE entries to `someday`/`archived`. Phase 1 = brain-app one-tap status change.
- `brain-non-pipeline-projects.md` — different axis. Adds `pipeline_enabled=false` flag for projects like Notebook. Doesn't address the visibility-of-archived problem.
- `brain-project-kanban.md` — parent vision (project-level org). Phase 4c pending. Doesn't touch the status-aware filtering.

This proposal slots between them: status-as-read filter. Read side of the same coin as brain-manual-stage-transitions.

## Critical analysis

- Is this the right thing to build? **Yes, urgently.** The audit work we just did is currently invisible. Without this, Michael will keep being overwhelmed by an inbox that doesn't honor his triage.
- Smallest useful version? **Phase 1 alone:** default-hide `someday` and `archived` from Capture's Recent and from EntriesView default list. Toggle to show. ~1 hour of work.
- What gets worse? Slight risk of "where did my entry go?" — mitigated by the toggle and by an "X hidden" indicator.
- Mosiah 4:27 check: This is the OPPOSITE of running faster than strength — it's making the cleanup we already did *legible*. Honors prior work.
- Does it duplicate brain-manual-stage-transitions? No. That's write side, this is read side. Both needed.

## Phase plan

- **Phase 1 (1 hr):** Hide `someday` + `archived` from CaptureView Recent and EntriesView default. Add toggle. Show "X hidden by status" indicator.
- **Phase 2 (2 hr):** Add `status` param to `listEntries` API + verify server-side. Status filter dropdown in EntriesView.
- **Phase 3 (1 hr):** ProjectDetailView board — collapse `someday`/`archived` entries into a small "X parked" badge below each lane, or hide entirely with a toggle.
- **Phase 4 (optional):** Capture "Recent" semantics — change from "last 10" to "last 10 unrouted/unreviewed."

## Open questions

- Should `done` also default-hide on the project board, or stay visible for satisfaction/momentum?
  - Lean: keep visible for momentum, but cap to last N done.
- Does the server already accept a status filter? Need a 2-min check of `handlers.go`.
- Should "archived" be searchable? (yes — Search view stays unfiltered.)
