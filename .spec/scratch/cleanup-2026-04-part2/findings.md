# Cleanup 2026-04 Part 2 — Cross-Reference Findings

**Date:** 2026-04-21
**Scope:** Inventory + reconcile: `.spec/proposals/` ↔ `.mind/active.md` ↔ brain DB projects/entries
**Inputs:**
- `.spec/proposals/` (37 status-bearing files, 35 active + 12 archive + 2 deferred)
- `.mind/active.md` (149 lines)
- `.mind/archive/active-{2026-03-22, 2026-04-04, 2026-04-16, 2026-04-20}.md`
- `private-brain/brain.db` snapshot (Apr 20, 22:29 — yesterday)

---

## 0. Critical Discovery — `.mind/active.md` is broken

```
Line 1:   # Active Context  (current, 2026-04-21)
Line 100: # Active Context  (old Apr 20 copy appended)
```

The Apr 20 archive snapshot was created correctly, but the rewrite of the live file appended to it instead of replacing it. Memory loaders read both. This needs a surgical fix before any other archiving — the archive copy already exists at `.mind/archive/active-2026-04-20.md`. Fix = trim live file to lines 1-99 (or similar — verify ending).

---

## 1. Brain Projects (10) — High-level umbrellas

| ID | Name | Workspace | Active? |
|----|------|-----------|---------|
| 1 | study | integrated | yes — primary |
| 2 | Sunday School | integrated | yes |
| 3 | Workspace improvements | integrated | yes — biggest umbrella |
| 4 | Space Center | external (`projects/space-center`) | yes |
| 5 | ibeco.me | integrated | yes |
| 6 | 2nd Brain | integrated | yes — biggest by entry count |
| 7 | YouTube / Content | integrated | yes — = teaching workstream |
| 8 | Budget App | integrated | yes |
| 9 | Notebook | integrated | low — random capture bucket |
| 10 | cpuchip.net | external (`projects/cpuchip.net`) | yes — needs site rebuild |

## 2. Brain Entries — 115 total, 107 still "open"

```
By category:    actions=15, ideas=45, inbox=8, journal=2, people=1, projects=39, study=5
By status:      None=96, active=6, archived=3, done=5, roadmap=2, someday=2, waiting=1
By route:       your_turn=72, pending=14, complete=3, suggested=1, dismissed=1, None=24
```

Reading: **96 entries have no `status` set at all.** That's the principal source of brain mess — entries advance through `route_status` (your_turn / pending / complete) but never get an explicit lifecycle status. So everything looks "open" forever.

72 entries sitting in `your_turn` = 72 things waiting on Michael. Not all are real asks; many are post-pipeline outputs that were never closed.

## 3. Proposal Status Inventory

### 3a. Active proposals — by truth state

#### ✅ Fully shipped — archive candidates (strict criteria)

| Proposal | Stated status | Reality (active.md says) | Action |
|----------|---------------|--------------------------|--------|
| `commission-ux-fixes.md` | "Shipped (2026-04-15)" | Listed as in-flight ▶ | **Archive.** active.md is stale. |
| `brain-project-ux.md` | "Phase 1-4 Shipped (2026-04-14)" | Not listed | **Archive.** Done ≥ 1 week. |
| `commission-ui.md` | "Phase 1-3 Shipped" (Apr 11) | Recently Shipped table | **Archive.** Confirmed done. |
| `brain-project-kanban.md` | "Phases 1-3 shipped, P4a-4b shipped, P4c next" | "ALL PHASES COMPLETE Apr 4-5" | **Conflict.** Read & verify P4c, then archive or split. |
| `orchestrator-steward/main.md` | "Proposed" | "P1-6 ALL COMPLETE Apr 10-11, 86 tests" | **Update status to SHIPPED, then archive.** |
| `brain-ux-quality-of-life.md` | "P1-7b ✅, P8 deferred" | "P1-7b ALL COMPLETE Apr 6" | **Strict = stays open** (P8). Move P8 to its own one-line entry, archive parent. |

#### ⚙ Partially shipped — stays in proposals/

| Proposal | What's open |
|----------|-------------|
| `gospel-engine/main.md` + `v2-hosted.md` | v1 SHIPPED, v1.5 ergonomics OPEN, v2 SHIPPED, v3 (graph) deferred |
| `brain-pipeline-evolution.md` | "P1-3 shipped, P4-7 specced" — but active.md says WS4 P1-9 ALL COMPLETE. **Conflict — needs reconciliation read.** |
| `brain-pipeline-fixes-phase4.md` | "Planned" — verify what this covers |
| `brain-pipeline-fixes.md` | (no top-level Status line — read needed) |
| `study-ibeco-me/main.md` | Refocused yesterday as UI-only Phase 1+ |
| `gospel-graph/main.md` | Specced, depends on engine v2 + AGE on PG18 |
| `brain-inline-panel.md` | "planned" — actively in flight |
| `brain-windows-service.md` | specced, not started |
| `brain-workspace-aware/` | "Draft" — read needed |
| `claude-code-integration.md` | "researched" |

#### 🟡 Open / specced / draft — keep but audit

| Proposal | State |
|----------|-------|
| `cleanup-2026-04/main.md` | Phases 1-3 done yesterday, Phase 4 deferred to tokenomics-2026 |
| `tokenomics-2026/main.md` | Research placeholder (created yesterday) |
| `token-efficiency.md` | "NEW Apr 16" — needs refresh per active.md |
| `memory-architecture.md` | Read needed — likely partially superseded by `.mind/` adoption |
| `sabbath-agent.md` | "Ready to build" |
| `classifier-qwen-fix.md` | "Ready to build" |
| `project-aware-pipeline.md` | "Draft" — but active.md says WS4 P9 ALL COMPLETE includes "project-aware pipeline". **Likely shipped — verify and archive.** |
| `enriched-indexer.md` | Status unclear |
| `embedding-comparison.md` | Status unclear |
| `debug-layer-triage.md` | Status unclear |
| `classify-bench.md` | Status unclear |
| `claude-code-integration.md` | researched, no execution |
| `study-workstream.md` | Has DONE markers per individual study, parent is rolling |
| `teaching-workstream.md` | 11-episode arc, content not started per active.md |
| `data-safety/` | Read needed |
| `lm-studio-model-experiments/` | Likely superseded — `archive/context-engineering-dev.md` says "COMPLETE" with results |
| `spec-housekeeping-archive.md` | "Proposed" — itself a cleanup proposal, ironic |
| `brain-ibecome-layer2.md` | Read needed |
| `brain-phase4-pipeline.md` | "Draft awaiting review" — but active.md says WS1 P4d "next" |
| `overview/main.md` | "Decisions recorded, ready to execute" — outdated; should rewrite or retire |

### 3b. Already archived (12) — leave alone

`brain-memory.md`, `brain-multi-agent/`, `brain-phase3c-sdk-agents.md`, `brain-relay.md`, `brain-ui-dashboard.md`, `brain-unified-dashboard.md`, `context-engineering-dev.md`, `context-engineering.md`, `enriched-search.md`, `notifications/`, `session-journal.md`, `squad-learnings.md`, `yt-emotion-analysis.md`

### 3c. Deferred (2) — leave alone

`second-brain-architecture.md` (MERGED into brain.exe), `tts-stt-reader.md` (Draft, iteration expected)

---

## 4. active.md ↔ Reality Drift Map

| active.md item | Truth | Action |
|----------------|-------|--------|
| Recently Shipped: Brain Project-Kanban "All phases" | brain-project-kanban.md says P4c next | Either reconcile (mark P4c shipped) or move back to In Flight |
| Recently Shipped: Orchestrator Steward "P1-6 ALL" | proposal still says "Proposed" | Update proposal status → SHIPPED |
| Recently Shipped: Commission UI "P1-3" | matches | OK — archive proposal |
| Recently Shipped: WS3 Brain UX QoL P1-7b | matches but P8 deferred | OK — keep P8 as one-liner |
| Recently Shipped: WS4 Brain Pipeline Evolution P1-9 | brain-pipeline-evolution.md says only P1-3 shipped | **Conflict — does WS4 = brain-pipeline-evolution? Or are these different docs?** |
| Recently Shipped: engine.ibeco.me P1-3 | matches gospel-engine/v2-hosted.md | OK |
| In Flight: Commission UX Fixes | proposal says SHIPPED Apr 15 | **MOVE to Recently Shipped** |
| In Flight: Other table — WS1 P4d | proposal says "Draft awaiting review" | Read needed |

## 5. Brain Entries ↔ Proposals — overlap analysis

Spot-check (not exhaustive — would require reading 107 entry titles):
- 39 entries in category=projects → many are likely tracking the same workstreams as proposals, but with no foreign key linking entry → proposal
- 5 study entries — should map to `study-workstream.md` rows
- 45 ideas — likely the biggest unsorted backlog; many candidates for promotion or close
- 15 actions — should be the most actionable, but most have status=None

**No `proposal_id` or `proposal_path` column on entries.** No `entry_ids` array in proposal frontmatter. The two systems can't actually link.

---

## 6. Workstream Reconstruction

The user mentioned: "We used to have these workstreams that we used to keep track of everything." Looking at `overview/main.md` and historical archive snapshots, the workstreams were:

- **WS1:** Brain pipeline core
- **WS2:** Brain UI/UX
- **WS3:** Brain UX QoL (per active.md)
- **WS4:** Brain Pipeline Evolution
- **WS5:** ibeco.me / engine
- (study + teaching are parallel "content" workstreams not numbered)

Recommendation: fold these back into the proposal/active.md vocabulary explicitly. Each in-flight item gets `[WS#]` prefix. brain projects table maps WS# → project_id.

---

## 7. Open Questions

1. Does brain-pipeline-evolution.md (says P1-3 shipped) = WS4 (says P1-9 all complete)? Likely two different phase numberings. Need to read the proposal.
2. Is project-aware-pipeline.md a separate proposal or was it folded into WS4 P9 ("project-aware pipeline (selective git commit)")?
3. brain-pipeline-fixes.md vs brain-pipeline-fixes-phase4.md — separate or a phase tree? 
4. study-workstream.md has individual study DONE markers and an open list — how does this map to brain entries in the `study` project (5 entries)?

These need read-passes during execution; not blocking the plan.
