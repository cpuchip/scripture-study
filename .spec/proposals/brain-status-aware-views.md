---
workstream: WS2
status: proposed
brain_project: 6
created: 2026-04-22
binding_problem: Brain stores status verbs (someday, archived, done, waiting) but no view filters by them. After today's audit set 26 entries to someday and 23 to archived, the UI looks identical — the cleanup is invisible. Michael is still overwhelmed by an inbox that doesn't honor his own triage.
related_brain_entries: ["22b8d8b2", "17749618"]
sister_proposals: ["brain-manual-stage-transitions.md", "brain-project-kanban.md"]
---

# Brain — Status-Aware Views

**Binding problem:** Brain's `status` field is write-only. The 2026-04-22 audit moved 26 entries to `someday` and 23 to `archived`, but the Capture homepage, the Entries list, and the project boards all look exactly the same as before. Status is a hidden DB attribute, not a visible signal. The cleanup work has zero UX payoff until the views honor what the data already says.

**Research:** [.spec/scratch/brain-status-aware-views/findings.md](../scratch/brain-status-aware-views/findings.md)

This proposal is the **read side** of [brain-manual-stage-transitions.md](brain-manual-stage-transitions.md). That proposal lets Michael MOVE entries between statuses; this one makes the move actually mean something visually.

---

## Success Criteria

1. Setting an entry to `someday` or `archived` immediately removes it from the Capture homepage's Recent list and from the EntriesView default list.
2. A clear, persistent toggle ("Show parked / Show all") makes hidden entries findable in one click.
3. Project board does not surface `someday`/`archived` entries in the Inbox/Working columns. They live in a collapsible "X parked" footer or behind a toggle.
4. The Search view remains unfiltered — searching always finds everything regardless of status.
5. A small indicator on filtered views shows "X hidden by status" so Michael never wonders where an entry went.

## In Scope

- `CaptureView.vue`, `EntriesView.vue`, `ProjectDetailView.vue` filtering changes.
- `api.ts` — add optional `status` filter param to `listEntries`.
- `scripts/brain/internal/server/handlers.go` — verify or add server-side support for status filter.
- New tiny component or composable: status-filter toggle + "hidden count" badge.

## Explicitly Out of Scope

- Changing the status vocabulary itself (already settled — see brain-manual-stage-transitions).
- Adding any new status values.
- Any change to maturity-driven board lanes (different axis, deliberate).
- Adding bulk-status-change UI (different proposal — could come later).
- Notebook/non-pipeline project special-casing (handled by brain-non-pipeline-projects).

## Prior Art

- [brain-manual-stage-transitions.md](brain-manual-stage-transitions.md) — write side. Builds the UI to set status. This proposal honors what that proposal sets.
- [brain-project-kanban.md](brain-project-kanban.md) — parent vision (project-level organization). Phases 1-4b shipped. This proposal is a small addition to the same surface area.
- [brain-non-pipeline-projects.md](brain-non-pipeline-projects.md) — adds the `pipeline_enabled` flag for projects like Notebook. Adjacent but independent — Notebook entries also benefit from this proposal because they tend to accumulate `someday` items fast.

## Proposed Approach

The fix is small and concentrated in three files. The expensive part is taste decisions about what to hide by default and what to show.

### Default visibility rules

| Status | CaptureView Recent | EntriesView default | ProjectBoard lanes | Search |
|--------|---------------------|----------------------|---------------------|--------|
| `active` (or NULL) | shown | shown | shown | shown |
| `done` | hidden after N days | shown with strikethrough | shown in Done lane | shown |
| `waiting` | shown with badge | shown | shown in Working lane | shown |
| `someday` | **hidden** | hidden by default | collapsed footer | shown |
| `archived` | **hidden** | hidden by default | collapsed footer | shown |
| `roadmap` | shown with badge | shown | shown in Working lane | shown |

### Toggle UX

Single small control near the top of EntriesView and ProjectDetailView:

```
[ Active ]  [ + Parked ]  [ All ]    23 hidden
```

Persists to `localStorage` per view. CaptureView gets just an "X parked entries hidden" link that expands the section.

## Phased Delivery

### Phase 1 — Hide & toggle in client (highest leverage, ~1 hr)

Pure client-side filter. No API change.

1. In `CaptureView.vue`, filter `recentEntries` to exclude `status === 'someday' || status === 'archived'`.
2. Add a "X parked" expandable footer.
3. In `EntriesView.vue`, default-filter the same way. Add the three-state toggle.
4. Show "X hidden by status" badge in both.

Verification: After the 2026-04-22 audit, Capture Recent shows ~3 actionable items instead of 10 mixed-state entries. EntriesView default list drops by ~50 entries.

### Phase 2 — Server-side filter (~2 hr)

Move the filter to the API so list endpoints don't return parked entries by default.

1. Verify `handlers.go` `GET /entries` parameter handling. Add `status` (multi-value) and `include_parked` (bool) if missing.
2. Update `api.ts` `listEntries` signature.
3. Default to `status NOT IN ('someday', 'archived')` server-side. `?include_parked=1` opts in.

Verification: API call without params returns only active. With `include_parked=1` returns all.

### Phase 3 — Project board lane treatment (~1 hr)

Update `ProjectDetailView.vue` `boardColumns` computation:

1. Skip `someday`/`archived` entries when building Inbox/Working/Done lanes.
2. Add a slim "X parked" expandable footer below the board (or per-lane).
3. Toggle to surface them inline if needed.

Verification: Open `/projects/4` (Space Center). The Star Trek UI entry (`17749618`, `status=someday`) no longer appears in the Inbox lane.

### Phase 4 (optional) — Capture Recent semantic fix

Currently CaptureView shows "last 10 entries." Change to "last 10 unrouted entries" (`project_id IS NULL`) OR "last 10 needing review." This eliminates the "duplicate appearance" problem (e.g. the science center entry showing in both Capture and Space Center board) for free.

This is more opinionated and could land later as a separate refinement.

## Verification Per Phase

| Phase | Test |
|-------|------|
| 1 | Set test entry to `someday` via API → reload Capture → entry not in Recent. Click "Show parked" → entry visible. |
| 2 | `curl /api/entries` returns no `someday`/`archived`. `curl /api/entries?include_parked=1` returns all. |
| 3 | Space Center board: `17749618` not visible in Inbox lane. Toggle "Show parked" → reappears. |
| 4 | Capture Recent does NOT show `22b8d8b2` (already in Space Center board). |

## Costs & Risks

- **Time:** Phases 1-3 = ~4 hours total. Each phase ships independently and provides value alone.
- **Risk: "where did my entry go?"** — Mitigated by the persistent "X hidden" badge and one-click toggle. Search always finds everything.
- **Risk: hiding `done` from project board hurts momentum** — Specifically NOT hiding `done`. Only `someday` and `archived`. Done stays visible for satisfaction.
- **Risk: server-side filter breaks existing callers** — Mitigated in Phase 2 by making the default change opt-out (param `include_parked=1`). Audit `api.ts` and any external callers (MCP tools, brain-app) before flipping the server default.
- **Maintenance burden:** Tiny — three files touched, one new toggle component.

## Creation Cycle Quick-Map

| Step | This proposal |
|------|---------------|
| Intent | Make the audit work visible. Honor cleanup with UX. |
| Covenant | Don't lose data; always recoverable via toggle/search. |
| Stewardship | brain.exe frontend (WS2). |
| Spiritual creation | This spec. |
| Line upon line | Phase 1 alone is useful; Phases 2-4 layer cleanly. |
| Physical creation | dev agent. |
| Review | Manual smoke test against the science center + Star Trek entries. |
| Atonement | Toggle + "X hidden" indicator = always recoverable. |
| Sabbath | Natural pause after Phase 1 (could ship and stop). |
| Consecration | Cleaner brain serves daily decision-making. |
| Zion | Pairs with brain-manual-stage-transitions to make the whole status verb system finally coherent. |

## Open Questions

1. Should `done` entries auto-hide from CaptureView Recent after N days? (Lean: yes, after 7 days.)
2. Should the project board collapse `someday`/`archived` per-lane or as one footer? (Lean: one footer beneath the board, simpler.)
3. Does the API server already accept a status filter, or is this a new param? (Verify first thing in Phase 2.)
