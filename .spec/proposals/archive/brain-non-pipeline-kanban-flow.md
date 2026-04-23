---
title: Brain — Non-Pipeline Kanban Flow (manual columns, manual actions, drag-drop)
status: shipped (2026-04-23) — Phases 1-3 + classify-gate piece of P4. brain-app portion of P4 explicitly skipped per user direction (no Project surface exists there yet; revisit if/when one is added). Done-dialog theming fix and handleCreateEntry project_id support landed in same session as part of post-ship adjacent surface cleanup.
workstream: WS2 Brain UX
created: 2026-04-23
brain_project: 6
binding_problem: >
  Now that non-pipeline projects exist (Notebook, future personal projects), the kanban board
  is broken for them in two ways. (1) Column placement is driven by `entry.maturity` (raw →
  Inbox, planned/specced → Working, verified/complete → Done) — but `maturity` is a pipeline
  field that will be permanently `raw` for non-pipeline entries. Every Notebook task sits in
  Inbox forever, no matter what `status` the user sets. (2) The card action buttons
  (Commission, Advance, Revise, Defer, Execute, Verify, Complete) are all pipeline operations.
  For "Get birthday present for mom" they're meaningless noise. The manual entry has nowhere
  to go and nothing to press. Result — the new pipeline_enabled=false flag is technically
  correct but practically useless until the kanban surface itself learns about manual flow.
sister_proposals:
  - brain-non-pipeline-projects.md      # ships pipeline_enabled flag (DONE)
  - brain-manual-stage-transitions.md   # ships status dropdown (DONE web/mobile)
---

## Evidence (audit, file:line)

- **Column derivation:** [scripts/brain/frontend/src/views/ProjectDetailView.vue:147-165](scripts/brain/frontend/src/views/ProjectDetailView.vue#L147-L165). The `boardColumns` computed routes by `e.maturity` exclusively. Non-pipeline entries have empty/`raw` maturity → all land in Inbox.
- **Pipeline action button block:** [scripts/brain/frontend/src/views/ProjectDetailView.vue:1055-1108](scripts/brain/frontend/src/views/ProjectDetailView.vue#L1055-L1108). The whole row of Commission/Advance/Revise/Defer/Execute/Cancel/Verify/Complete/Undo. Visibility gated by `canAdvance / canRevise / canExecute / canVerify / canCancel / canComplete / canCommission`. List view duplicates the same buttons around line 1158.
- **Project context already loaded on this view:** `project.value.pipeline_enabled` is available — same prop the badge uses.
- **Status vocabulary today:** `'', active, waiting, roadmap, someday, done, archived`. No explicit "in progress / working" value. The closest is `active` (default after capture).

## Root cause

Two coupled problems hiding behind one symptom:

1. **Column placement assumes pipeline.** Maturity is the only signal. Adding a non-pipeline path means we need a parallel routing rule that uses status.
2. **Button row assumes pipeline.** Every action mutates pipeline state. Even the "advance" verb is a pipeline term.

Drag-and-drop is the mechanism the user wants for manual flow — it's the natural gesture when the buttons stop being meaningful.

## Proposed approach

### Status vocabulary addition

Add **`working`** to the status enum. Position in the canonical order: `'', active, working, waiting, roadmap, someday, done, archived`.

- `active` keeps its current meaning ("captured, not yet started") — the default after capture.
- `working` = "in progress" (the column the user is actively touching).
- This avoids overloading `waiting` (which means *blocked*, not *being-worked*).
- Pipeline entries can also use `working` later if useful; not required.

**Migration:** None needed — empty/missing status remains valid; old entries just don't get the new value until the user moves them.

### Column placement rule

```
if (project.pipeline_enabled === false) {
  // Status-driven columns
  Inbox   = status in {'', 'active'}
  Working = status === 'working'
  Done    = status === 'done'
  // Parked (someday/archived) goes to footer as today.
} else {
  // Existing maturity-driven columns (unchanged)
}
```

### Action buttons

Replace the entire pipeline button row with a manual button row when `project.pipeline_enabled === false`:

| Button | Action | Confirm/Dialog |
|--------|--------|----------------|
| **▶ Start** (Inbox only) | `status = 'working'` | None |
| **✓ Done** (Inbox + Working) | `status = 'done'` | Optional reason dialog (small, dismissible) |
| **↩ Reopen** (Done only) | `status = 'active'` | None |
| **⏸ Someday** (any) | `status = 'someday'` | None |
| **🗄 Archive** (any) | `status = 'archived'` | Confirm — "archive forever?" — same as today |

The optional-reason dialog is a single textarea + "Save" / "Skip". The reason gets appended to the entry body as `\n\n---\n_Closed: {reason}_` (or stored as a comment if a comments table exists — to be decided in implementation). Skipping is the default fast path — one click closes it out.

### Drag and drop

Use a small library (`vuedraggable` is the Vue 3 standard, sortablejs underneath; ~15 KB gzipped). On drop:
- Inbox ↔ Working ↔ Done sets `status` to the column's anchor value.
- Drag to footer (parked) sets `status = 'someday'`.
- Optimistic UI update; PUT `/api/entries/{id}` with `{status: '...'}`; reload on error.

### Capture default for non-pipeline projects

Today the create-entry-in-project form defaults `category: 'ideas'` and triggers the AI pipeline. For non-pipeline projects that capture step is wasted work. Confirm via test — does setting `pipeline_enabled=false` on the project already short-circuit classification? (We added the gate in `routeEntry` and `BuildProjectContext`, but classify runs separately.) If classification still fires, gate it the same way.

## Phased delivery

| Phase | Scope | Standalone value |
|-------|-------|------------------|
| **1** | Status vocab adds `working`. Column rule branches on `pipeline_enabled`. Replace button row with manual buttons (no reason dialog yet, no DnD). | Notebook works end-to-end with click-to-move. The user's pain is gone. |
| **2** | Optional-reason dialog on the Done button. | Small QoL — closing out feels deliberate, not just a click. |
| **3** | Drag-and-drop via vuedraggable. | The "cool" part the user mentioned. Pure UX upgrade. |
| **4** | Audit: does the classify step fire for non-pipeline entries on creation? Gate it if so. Mirror the manual button row in brain-app's project board view. | Closes the loop on token waste + cross-surface parity. |

Phase 1 is the floor. Phase 3 is genuinely optional — keyboard/click flow already covers the use case.

## Adjacent surfaces (foresight audit)

1. **Scope:** brain-app project-board view has the same Commission/Advance/etc. buttons. Phase 4 mirrors there.
2. **Discoverability:** the new buttons should look visually distinct from pipeline buttons — consider grayer/calmer palette since these are manual not AI. Notebook's 📓 badge already signals the project type.
3. **Contracts:** `handleUpdateEntry` already accepts `{status: ...}` (verified yesterday). DnD just calls it. No new API.
4. **Spec gap to surface:** the user said "drag and drop between columns" — assume they also want drag *out* of the active board (to parked/someday). Confirmed by their "kick it to the backlog" phrasing. Do not assume they want cross-project DnD — that's a different feature.

## Costs and risks

- **New dependency** (vuedraggable): small but is a new dep. Acceptable.
- **Status enum drift:** adding `working` means the brain-app dropdown, ibeco.me filters, and any future SQL filters need updating. Grep before shipping. Mitigation: search `status === 'active'` and `'waiting', 'someday', 'done', 'archived'` first.
- **Confusion risk:** pipeline projects keep their existing buttons. Non-pipeline projects get different buttons. The 📓 badge + `pipeline_enabled` flag should make the divergence legible. If users get confused, that's a discoverability issue to fix, not a reason to merge the surfaces.

## Verification (Agans Rule 9)

- Create a Notebook entry → set status `working` via button → entry should appear in Working column. Revert button → entry stays in Inbox. (Restore.)
- Create a pipeline-project entry → confirm Commission/Advance/etc. buttons still appear and column placement still uses maturity (no regression).
- DnD: drag from Inbox to Done → status becomes `'done'` → reload → entry persists in Done column.

## Open questions for the implementer

1. Where to store the optional close-out reason? Append to body, or use a new comments table? (Default: append to body — simplest, no schema change.)
2. Should the "Done" button preserve original status as `previous_status` so Reopen restores it? (Default: no — Reopen always goes to `active`. Simpler.)
3. brain-app: drag-and-drop on mobile is touchy. Phase 4 should probably stick with long-press + tap (already shipped) and skip DnD on mobile.
