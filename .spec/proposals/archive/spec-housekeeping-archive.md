# Proposal: .spec Housekeeping — Archive Completed Work

*Created: 2026-04-12*
*Status: Proposed — awaiting Michael's review*
*Type: Quick plan (housekeeping)*

---

## Binding Problem

The `.spec/proposals/` folder has 27+ items. Many represent work that shipped weeks ago (per `active.md`) but whose proposal files were never moved to archive. This creates noise — when scanning proposals for "what's next," you have to mentally filter out completed work. The scratch folder (38 directories) has the same problem. Housekeeping restores signal.

## Current State

**Proposals:** 27 items in `.spec/proposals/` (files + directories)
**Archive:** 13 items already in `.spec/proposals/archive/`
**Deferred:** 2 items in `.spec/proposals/deferred/`
**Scratch:** 38 items in `.spec/scratch/`

---

## Tier 1: Archive — Fully Shipped (per active.md)

These proposals have ALL phases complete. Move to `.spec/proposals/archive/`.

| Proposal | Evidence | Shipped |
|----------|----------|---------|
| `brain-pipeline-fixes.md` | All 4 phases (incl 3.5) complete | Apr 9-10 |
| `commission-ui.md` | Phase 1-3 shipped, all phases complete | Apr 11 |
| `brain-ux-quality-of-life.md` | Phases 1-7b shipped (Phase 8 deferred to own proposal) | Apr 5-6 |
| `brain-pipeline-evolution.md` | Phases 1-9 shipped (governance through project scaffolding) | Apr 6-7 |
| `orchestrator-steward/` | All 6 phases shipped (retry → escalation → breaker → quarantine → nudge → commission) | Apr 10-11 |
| `project-aware-pipeline.md` | Phases 9a-9c shipped (was WS4 final phase) | Apr 7 |

**Note:** The proposal files themselves have stale status headers (the subagent found "Proposed" on orchestrator-steward, "Phase 1-3 shipped" on pipeline-evolution). The truth is in `active.md` — the code is built, tested, and running. The stale headers are part of the problem this housekeeping fixes.

## Tier 2: Superseded — Merge or Archive with Note

These proposals were early drafts later superseded by more complete proposals.

| Proposal | Superseded By | Action |
|----------|--------------|--------|
| `brain-phase4-pipeline.md` | `brain-pipeline-evolution.md` (which shipped all phases) | Archive with "superseded" note |
| `brain-pipeline-fixes-phase4.md` | `brain-pipeline-fixes.md` (shipped) + orchestrator-steward (handles timeouts/failures now) | Archive with "superseded" note — steward handles what this proposed |

## Tier 3: Consider Deferring

These proposals are valid but haven't been touched in weeks and aren't on the immediate roadmap. Consider moving to `.spec/proposals/deferred/`.

| Proposal | Last Activity | Rationale |
|----------|--------------|-----------|
| `brain-windows-service.md` | Draft, untouched | Nice-to-have, not blocking anything |
| `classify-bench.md` | Proposed | Benchmarking tool — useful but not urgent |
| `embedding-comparison.md` | Proposed | Comparison tool — qwen models benchmarked already in lm-studio experiments |
| `enriched-indexer.md` | Proposed | TITSW summaries for gospel-vec — good but not next |
| `gospel-graph/` | Proposed | Visualization — aspirational, no immediate plan |
| `claude-code-integration.md` | Proposed | Alternative backend — exploratory, not urgent |

## Tier 4: Keep Active

These are either actively being worked or are next-up on the roadmap.

| Proposal | Reason to Keep |
|----------|---------------|
| `overview/` | Master workstream plan — living document |
| `study-workstream.md` | WS-S — Priority 1 per active.md |
| `teaching-workstream.md` | WS-T — Priority 2 per active.md |
| `brain-inline-panel.md` | Next brain UX work (reply + nudge controls) |
| `brain-project-kanban.md` | Partially shipped, 4a-4b remaining |
| `brain-ibecome-layer2.md` | Brain ↔ ibecome sync — planned feature |
| `brain-workspace-aware/` | Agent session improvements — planned |
| `data-safety/` | Dev agent hardening — important safety work |
| `gospel-engine/` | Combined gospel search — planned migration |
| `classifier-qwen-fix.md` | Ready to build, affects pipeline quality |
| `memory-architecture.md` | Memory redesign — "Now" items ready |
| `sabbath-agent.md` | Ready to build, supports reflection workflow |
| `lm-studio-model-experiments/` | Active testing phase |

---

## Execution Plan

**Phase 1 — Archive shipped proposals (5 min)**
```
mv .spec/proposals/brain-pipeline-fixes.md → .spec/proposals/archive/
mv .spec/proposals/commission-ui.md → .spec/proposals/archive/
mv .spec/proposals/brain-ux-quality-of-life.md → .spec/proposals/archive/
mv .spec/proposals/brain-pipeline-evolution.md → .spec/proposals/archive/
mv .spec/proposals/orchestrator-steward/ → .spec/proposals/archive/
mv .spec/proposals/project-aware-pipeline.md → .spec/proposals/archive/
```

**Phase 2 — Archive superseded proposals (2 min)**
```
mv .spec/proposals/brain-phase4-pipeline.md → .spec/proposals/archive/
mv .spec/proposals/brain-pipeline-fixes-phase4.md → .spec/proposals/archive/
```

**Phase 3 — Defer low-priority proposals (optional, per Michael's judgment)**
Move selected items from Tier 3 to `.spec/proposals/deferred/`.

**Phase 4 — Update active.md**
Trim the "In Flight" section. Completed items get one-line references instead of full phase logs. Move detailed history to an archive snapshot.

---

## Scratch Files

Per project convention, scratch files are **permanent research provenance** — they don't get deleted or archived. But the scratch folder could benefit from a `README.md` noting which scratch dirs correspond to which shipped proposals, for future navigability.

---

## Recommendation

**Proceed with Tier 1 + Tier 2 now.** That's 8 proposals moved to archive, cutting the active folder from 27 to 19 items. Tier 3 (deferrals) is Michael's call — some of those might spark back to life. Tier 4 stays put.

The active.md trim (Phase 4) should happen at the next Sabbath or cycle-end, when there's time to properly archive the detailed history.
