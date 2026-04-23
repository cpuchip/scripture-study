---
title: Brain — Non-Pipeline Projects
status: shipped Phase 1+2 (2026-04-23)
workstream: WS2 Brain UX
created: 2026-04-22
refined: 2026-04-23
brain_project: 6
binding_problem: Brain currently routes every entry through the AI classification + planning pipeline. There is no way to create a project where entries skip the pipeline and become directly user-actionable (errands, family items, personal notebook stuff). Consequence — personal todos like "Custom Desk Project" or "Grocery list" land in the same processing flow as workspace ideas, which both wastes pipeline cycles and clutters AI context with stuff the AI shouldn't be touching.
sister_proposals:
  - brain-manual-stage-transitions.md
  - brain-status-aware-views.md
---

# Brain — Non-Pipeline Projects

## Codebase audit (2026-04-23)

Before writing implementation: traced the pipeline entry points to find where the gate must live.

| Concern | Location | Notes |
|---------|----------|-------|
| Project schema | [scripts/brain/internal/store/db.go:1249](../../scripts/brain/internal/store/db.go) `migrateProjects` | Uses `columnNames()` helper for ALTER TABLE ADD COLUMN IF NOT EXISTS pattern — follow exactly. |
| Project struct | [scripts/brain/internal/store/types.go:8](../../scripts/brain/internal/store/types.go) `type Project struct` | Add `PipelineEnabled bool` with default true. |
| Routing fork (gate point #1) | [scripts/brain/internal/web/server.go:1243](../../scripts/brain/internal/web/server.go) `routeEntry()` | Called from create-entry handlers and `CreateAndRouteEntry`. Single early-exit check here covers both manual creation and scheduler. |
| Brain research/plan pipeline | [scripts/brain/internal/pipeline/context.go:87-213](../../scripts/brain/internal/pipeline/context.go) | Already takes `entry.ProjectID` to fetch project context. Add `if !project.PipelineEnabled { return nil }` early-exit to context builder OR to whatever calls it. |
| Project CRUD UI | (likely brain frontend `ProjectsView.vue` or similar) | Phase 2 — needs investigation. |

**Existing partial-update pattern:** `handleUpdateProject` (if it exists, mirroring `handleUpdateEntry` at server.go:369) does read-modify-write with `json.RawMessage` field detection. Follow that pattern for the new flag so unsetting other fields doesn't blank out the project.

## Success Criteria

- A project can be flagged `pipeline_enabled=false` (call it "manual project" or "notebook project").
- Entries created in or moved to such a project skip the auto-routing step (`routeEntry`) entirely.
- Brain UI shows manual projects with a different visual treatment (no "needs planning" / "needs review" badges).
- The Notebook project (id=9, post-merge) becomes the first manual project.
- Existing pipeline behavior is unchanged for all other projects.

## Constraints

- DB change must be backward-compatible (default `pipeline_enabled=1` for all existing projects).
- Brain app (Flutter) and ibeco.me web UI both need to honor the flag for badges; backend gate is the source of truth.
- Pipeline router needs ONE early-exit check, not multiple scattered ones (single source of truth).

## Phased Delivery

### Phase 1 — Schema + backend gate (~45 min)

1. **Migration** in `migrateProjects()`:
   ```sql
   ALTER TABLE projects ADD COLUMN pipeline_enabled INTEGER NOT NULL DEFAULT 1
   ```
   Use the existing `columnNames("projects")` check pattern. **Manually set Notebook id=9 to 0** in the same migration step (idempotent: `UPDATE projects SET pipeline_enabled=0 WHERE name='Notebook' AND pipeline_enabled=1`).

2. **Project struct**: add `PipelineEnabled bool \`json:"pipeline_enabled"\`` to `types.go`. Update `INSERT/UPDATE` SQL in db.go to include the column. **Default in CreateProject must be true.**

3. **Gate in `routeEntry`** (web/server.go:1243): before `s.pool.StartTask`, look up project. If `entry.ProjectID != nil && project.PipelineEnabled == false`, skip routing entirely (no UpdateRouteStatus, no goroutine, no nothing). Set entry.Source to indicate it's a manual entry that didn't route.

4. **Gate in pipeline/context.go**: in whichever stage entry-point reads `entry.ProjectID` (lines 87, 187, 210), early-exit if project is non-pipeline. Defense in depth.

### Phase 2 — Project CRUD UI (~30 min)

- Add `pipeline_enabled` checkbox to project create/edit dialog in brain.exe frontend.
- Visual treatment for manual projects: muted color or `📓` notebook icon next to project name on dashboard.
- Hide "needs planning" / "needs review" badges on entries belonging to manual projects.

### Phase 3 — brain-app + ibeco.me parity (~20 min each)

- Flutter app: same visual treatment, same badge suppression.
- ibeco.me TasksView: add `pipeline_enabled` to BrainEntry/Project cache schema (goose migration on PostgreSQL), surface in UI.

## Verification

**Phase 1 inverse hypothesis (Agans Rule 9):**
```powershell
# 1. Reproduce baseline: create entry to Notebook (id=9) → confirm route is attempted (check logs for "Agent route ...")
# 2. Apply migration + gate.
# 3. Restart brain.exe.
# 4. Create entry to Notebook → confirm NO route logs and entry.RouteStatus stays empty.
# 5. Create entry to non-Notebook project → confirm routing still works (regression check).
# 6. Temporarily revert the gate → step 4 should now route again. Restore. Confirm.
```

## Costs / Risks

- **Risk: forgetting a gate location.** Mitigation: routes flow through `routeEntry()` and pipeline/context.go is the only deeper consumer. After Phase 1, grep for any other reads of `entry.ProjectID` in goroutine/queue paths.
- **Risk: existing Notebook entries already have a route status from previous routing attempts.** Acceptable — they remain visible; new entries skip routing. No backfill needed.
- **Risk: migration runs on production but data is empty for Notebook id=9.** Mitigation: the `WHERE name='Notebook'` clause is no-op safe.

## Decision log

- **Why a single boolean and not a richer "pipeline mode" enum?** The audit found exactly one binary need: skip routing or don't. Future modes (e.g. "classify only, no plan") can extend later without breaking this flag (`pipeline_enabled=false` would still mean "no AI processing", richer modes would override).
- **Why gate in routing AND pipeline/context.go?** Defense in depth. Routing is the primary entry point but if any future code path calls into `BuildProjectContext` directly for a manual project, the second gate keeps the AI from reading non-pipeline data. Cheap insurance.
- **Why not a `notebook_only` boolean instead?** Naming. "Notebook" is a project name; the flag is about pipeline behavior, not category.

## Related

- Pairs with `brain-manual-stage-transitions.md` — this proposal makes manual projects exist; the other gives them a usable lifecycle UI.
- Closes the gap surfaced by the 2026-04-22 brain audit (7 personal entries needed a home that didn't trigger AI processing).
