---
title: Brain — Non-Pipeline Projects
status: proposed
workstream: WS2 Brain UX
created: 2026-04-22
binding_problem: Brain currently routes every entry through the AI classification + planning pipeline. There is no way to create a project where entries skip the pipeline and become directly user-actionable (errands, family items, personal notebook stuff). Consequence — personal todos like "Custom Desk Project" or "Grocery list" land in the same processing flow as workspace ideas, which both wastes pipeline cycles and clutters AI context with stuff the AI shouldn't be touching.
---

# Brain — Non-Pipeline Projects

## Binding Problem

Brain treats every entry the same: capture → classify → plan → action. That works for workspace projects (study/, ibeco.me, gospel-engine, etc.) where AI involvement is the point. It breaks down for personal life — desk builds, grocery lists, birthday presents, temple visits. These need a place to live in brain (so capture from the watch app still works, so they're searchable, so they're tracked) but they should NOT be classified, planned, or auto-actioned by AI.

Today the only workaround is to leave the entry uncategorized (status=NULL) forever, which makes it invisible to inbox cleanup and shows up as "needs triage" forever.

## Success Criteria

- A project can be flagged `pipeline_enabled=false` (call it "manual project" or "notebook project").
- Entries created in or moved to such a project skip the auto-classification step.
- Brain UI shows manual projects with a different visual treatment (no "needs planning" / "needs review" badges).
- The Notebook project (id=9, post-merge) becomes the first manual project.
- Existing pipeline behavior is unchanged for all other projects.

## Constraints

- DB change must be backward-compatible (default `pipeline_enabled=true` for all existing projects).
- Brain app (Flutter) and ibeco.me web UI both need to honor the flag.
- Pipeline router (the part that decides "should this entry be classified?") needs one early-exit check.

## Proposed Approach

1. Add `projects.pipeline_enabled BOOLEAN DEFAULT 1` column.
2. Set `pipeline_enabled=0` on Notebook (id=9).
3. Update brain.exe pipeline entrypoint: if entry's project has `pipeline_enabled=0`, mark it as `manually-managed` and skip downstream stages.
4. Update brain-app: hide "ready for planning" UI affordances on entries belonging to manual projects.
5. (Optional Phase 2) Add UI to create a new manual project / toggle the flag on existing ones.

## Phased Delivery

- **Phase 1:** schema + flag-honoring pipeline router. Notebook becomes manual. All existing pipeline-eligible entries unaffected.
- **Phase 2:** UI to create/toggle manual projects.

## Verification

- After Phase 1, capture an entry to Notebook → it appears in Notebook with no "needs classification" badge and is not picked up by the AI agent.
- Capture an entry to any other project → behavior unchanged.

## Related

- Pairs with `brain-manual-stage-transitions` (this proposal makes manual projects exist; the other gives them a usable lifecycle).
- Closes the gap surfaced by the 2026-04-22 brain audit (7 personal entries needed a home that didn't trigger AI processing).

## Costs / Risks

- Small surface area, but touches 3 codebases (brain.exe, brain-app, ibeco.me). Phase 1 alone is the high-leverage part.
- Risk: forgetting to gate a pipeline stage somewhere → manual entries get partially processed. Mitigation: single early-exit check in the router.
