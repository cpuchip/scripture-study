---
title: Brain — Manual Stage Transitions
status: shipped Phase 1 mobile (2026-04-23); Phase 2 web pre-existing
workstream: WS2 Brain UX
created: 2026-04-22
refined: 2026-04-23
brain_project: 6
binding_problem: Once an entry is in a non-pipeline project (or any entry the user wants to drive manually), there is no UI to move it through stages. Brain assumes the AI pipeline does state transitions; nothing exposes "I'm working on this" → "I'm done" for human-driven items. Result — manual entries either stay forever as `status=NULL/active` or the user has to edit raw DB records to close them out.
sister_proposals:
  - brain-non-pipeline-projects.md
  - brain-status-aware-views.md
---

# Brain — Manual Stage Transitions

## Codebase audit (2026-04-23)

Good news: the **API already supports manual status transitions**. The audit found that `handleUpdateEntry` does read-modify-write with `json.RawMessage` field detection — sending `{"status": "done"}` works today on every entry. The gap is purely UI.

| Layer | Capability today | Gap |
|-------|------------------|-----|
| HTTP API | [scripts/brain/internal/web/server.go:369](../../scripts/brain/internal/web/server.go) `handleUpdateEntry` | ✅ partial update with `status` — nothing to add |
| Status vocabulary | `active, someday, roadmap, waiting, done, archived` (verified by 04-22 audit) | ✅ stable, no new statuses needed |
| brain.exe project board | drag between status columns? | ❓ partial — need to verify (board exists, drag-drop unclear) |
| brain.exe entry detail | status dropdown | ❌ missing |
| brain-app entry detail | `archiveEntry()` only sets to 'archived' | ❌ no general status picker |
| ibeco.me web TasksView | edit dialog has status field (verified line 312 in TasksView.vue) | ✅ already exposes it; needs prominence |

**Insight:** the original proposal called for kanban + drag UI as Phase 2. After the audit, the highest-leverage move is mobile (one-tap close-out) because that's where the friction is sharpest — you triage on the phone but have to switch to desktop to close. ibeco.me's TasksView edit dialog already surfaces status; brain-app does not.

## Success Criteria

- Brain app and brain.exe both expose a status picker on every entry detail screen.
- Changing status writes to DB immediately via existing PUT `/api/entries/{id}` endpoint and is reflected across clients via the existing WebSocket push mechanism.
- For non-pipeline projects (see `brain-non-pipeline-projects.md`), manual transitions are the ONLY way state changes — the routing pipeline doesn't touch them.
- (Optional, Phase 3) Stage history is preserved (so we can see "moved to done on 2026-04-22").

## Constraints

- Don't break existing pipeline-driven transitions for AI-managed entries.
- Status vocabulary stays small and stable — the 6 statuses already in use, plus the orthogonal `action_done` boolean.
- Mobile-first: the brain app needs one-tap close-out from the entry detail screen.

## Phased Delivery

### Phase 1 — brain-app status picker (~45 min)

**Highest-leverage phase.** Solves the "can't close out personal todos from my phone" pain.

1. In [scripts/brain-app/lib/screens/edit_entry_screen.dart](../../scripts/brain-app/lib/screens/edit_entry_screen.dart) (or wherever entry detail is rendered): add a `DropdownButton<String>` for status with the 6 values. Default to current value.
2. On change, call existing `BrainApi.updateEntry(id, {'status': newValue})`. The API and partial-update logic already exist — nothing backend-side.
3. Show the current status as a colored chip on the history list tile (visual reinforcement). Status colors: active=blue, someday=amber, roadmap=purple, waiting=gray, done=green, archived=stone.
4. Add a quick-action: long-press an entry on history screen → popup → "Mark done", "Park (someday)", "Archive". Single tap close-out from list.

### Phase 2 — brain.exe entry detail status picker (~30 min)

- Project board (Vue) likely already supports column drag for kanban — verify and document.
- Entry detail dialog: same dropdown pattern as mobile. Pulls from PUT `/api/entries/{id}`.
- Confirmation note in UI: when user manually sets status on a pipeline-managed entry, show a small note: "Override applied; pipeline will not change this status." (See risk mitigation below.)

### Phase 3 (deferred) — Stage history audit trail

- Add `entries_status_history (entry_id, old_status, new_status, source, changed_at)` table.
- Trigger or app-level write on every UpdateEntry that changes status.
- Surface in entry detail as a timeline.
- Defer until Phases 1+2 prove out the manual-transition workflow is actually used.

## Verification

**Phase 1:**
- Open a Notebook entry on phone → select "done" from dropdown → entry status updates immediately, list refreshes, entry disappears from default view (because parked filter from 04-23 also hides done).
- Open a routed entry (e.g. workspace idea) → manually set to "waiting" → status changes → confirm pipeline doesn't subsequently overwrite.

**Inverse hypothesis (Agans Rule 9):** revert the dropdown widget → verify there's no other affordance to change status from the app → confirm the only way is re-add the dropdown.

## Costs / Risks

- **Risk: introducing manual transitions could let the user put entries into states the pipeline doesn't expect.** Mitigation: pipeline reads entry status only when re-running stages. Manual override = state of the world; pipeline respects it. The non-pipeline-projects proposal handles the "never touch this" case more robustly.
- **Risk: the user marks something done that the pipeline will then "un-done" when its next stage runs.** Mitigation (defensive): if entry has `route_status=complete`, don't let the pipeline reset `status` away from a manually-set done. Honestly, low likelihood until we see it in practice.
- **Risk: brain-app two-codebase change for one feature.** Acceptable — mobile is the highest-leverage surface for this specific pain.

## Decision log

- **Why mobile first?** Triage happens on the phone. The 04-22 audit closed 60+ entries via raw SQL because the desktop dropdown (if present) wasn't where Michael was when he wanted to close them out.
- **Why no kanban Phase 1?** Web kanban is a bigger lift, and the project board likely already has *something*. Verify before building.
- **Why defer history audit trail?** YAGNI until Phases 1+2 prove the workflow. Adding it later is non-breaking (status changes are forward-only timestamped, can backfill from `updated_at`).

## Related

- Pairs with `brain-non-pipeline-projects.md` — this proposal gives manual projects their lifecycle UI.
- Pairs with `brain-status-aware-views.md` (already shipped on desktop, ecosystem-parity in flight) — once entries can be parked or completed via UI, the existing hide-parked filter immediately rewards the user with visible cleanup.
- Without this, the non-pipeline project is read-only-ish (entries enter but never gracefully leave).
