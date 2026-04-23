---
title: Brain — Manual Stage Transitions
status: proposed
workstream: WS2 Brain UX
created: 2026-04-22
binding_problem: Once an entry is in a non-pipeline project (or any entry the user wants to drive manually), there is no UI to move it through stages. Brain assumes the AI pipeline does state transitions; nothing exposes "I'm working on this" → "I'm done" for human-driven items. Result — manual entries either stay forever as `status=NULL/active` or the user has to edit raw DB records to close them out.
---

# Brain — Manual Stage Transitions

## Binding Problem

Brain's status field (`active`, `someday`, `roadmap`, `waiting`, `done`, `archived`) is well-defined and the audit proved it's the right vocabulary. But the only way to transition between states today is via the pipeline (auto-classification, planning agent, etc.) or by editing the database directly. There's no kanban-style drag-between-columns or "mark this done" button on a non-pipeline entry.

The 2026-04-22 audit closed 60+ entries via SQL because there was no UI to do it. That's a tooling gap, not an audit habit.

## Success Criteria

- Brain app and ibeco.me web UI both expose a status selector / drag target on every entry.
- Changing status writes to DB immediately and is reflected across clients.
- For non-pipeline projects (see `brain-non-pipeline-projects`), the manual transitions are the ONLY way state changes — pipeline doesn't touch them.
- Stage history is preserved (so we can see "moved to done on 2026-04-22").

## Constraints

- Don't break the existing pipeline-driven transitions for AI-managed entries.
- Action vocabulary stays small and stable — the 7 statuses already in use, plus `action_done` for the boolean-completion flag.
- Mobile-first: the brain app capture screen needs a one-tap status change.

## Proposed Approach

1. Add an entry detail action: status dropdown / kanban column drag.
2. Wire it through the existing brain.exe HTTP API (the API likely already supports status updates — verify).
3. Surface in brain-app as a long-press / swipe action on the entry list.
4. Surface in ibeco.me as a kanban-style board view (columns = statuses, drag between).
5. (Optional) Add a status_changed_at audit trail column.

## Phased Delivery

- **Phase 1:** Brain app one-tap status change on entry detail. (Highest leverage — solves the immediate "can't close out personal todos" pain.)
- **Phase 2:** ibeco.me kanban view.
- **Phase 3:** Status history audit trail.

## Verification

- After Phase 1: open Notebook entry "Get birthday present for mom" on phone → tap "Done" → entry moves to done status, disappears from active list.
- Pipeline-driven entries: their status changes still flow from the AI agent; manual override still possible but logged differently.

## Related

- Pairs with `brain-non-pipeline-projects` — this proposal gives manual projects their lifecycle UI.
- Without this, the non-pipeline project is read-only-ish (entries enter but never gracefully leave).

## Costs / Risks

- Touches 2 client codebases (brain-app, ibeco.me). Phase 1 alone is the unblocker.
- Risk: introducing manual transitions could let the user put entries into states the pipeline doesn't expect. Mitigation: pipeline-managed entries get a separate code path; manual transitions only apply where `pipeline_enabled=false` OR via an explicit "override pipeline" action.
