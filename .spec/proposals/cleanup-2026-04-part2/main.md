---
workstream: WS5
status: building
brain_project: 3
created: 2026-04-21
last_updated: 2026-04-21
phase_status: "Phase A done, Phase B done (16 archives), Phase C in progress"
---

# Cleanup 2026-04 Part 2 — Reality Reconciliation

**Binding problem:** Three separate views of "what we're doing" — `.spec/proposals/`, `.mind/active.md`, and brain DB projects+entries — have drifted out of sync with each other and with the actual state of the work. Michael burns tokens every session re-establishing what's actually in flight. Proposals say "Proposed" for work that shipped two weeks ago. `active.md` lists shipped work as in-flight. Brain has 107 open entries with no link back to the proposals that own them.

This proposal does the reconciliation pass. It does NOT include the structural fix to keep them from drifting again — that's [brain-vscode-bridge](../brain-vscode-bridge/main.md), a sibling proposal.

**Created:** 2026-04-21
**Type:** Cleanup / housekeeping
**Status:** Proposed — plan only, awaiting Michael's review before execution
**Predecessor:** [cleanup-2026-04/main.md](../cleanup-2026-04/main.md) (Phases 1-3 done 2026-04-21)
**Findings:** [.spec/scratch/cleanup-2026-04-part2/findings.md](../../scratch/cleanup-2026-04-part2/findings.md)
**Scratch:** [.spec/scratch/cleanup-2026-04-part2/](../../scratch/cleanup-2026-04-part2/)

---

## 1. The Three Views

| View | Lives at | Purpose | Currently |
|------|----------|---------|-----------|
| Proposals | `.spec/proposals/*.md` | Long-form rationale, phased plan, rolling status | 35 active + 12 archived + 2 deferred. Status fields drifted vs reality. |
| Active context | `.mind/active.md` | Snapshot of what's in flight right now | **Broken** — has two `# Active Context` headers (lines 1 and 100). Recent rewrite appended to its archive copy instead of replacing. |
| Brain | `private-brain/brain.db` | Actionable work items + projects | 10 projects, 115 entries (107 still open, 96 with `status=None`). 72 entries in `route_status=your_turn`. No FK to proposals. |

Each view is internally semi-coherent. Cross-view they're inconsistent enough that the cheapest path to "what am I doing?" is to ask Claude.

## 2. Findings — what's actually broken

Detail in [findings.md](../../scratch/cleanup-2026-04-part2/findings.md). Summary:

1. **active.md is duplicated**. Two `# Active Context` blocks. Memory loaders read both. **Surgical 2-minute fix.**
2. **At least one proposal shipped without status update**: `commission-ux-fixes.md` says "Shipped (2026-04-15)" but active.md still has it as in-flight.
3. **At least one proposal is shipped per active.md but still says "Proposed"**: `orchestrator-steward/main.md`.
4. **Phase numbering conflict**: `brain-pipeline-evolution.md` says P1-3 shipped, but active.md says "WS4 Brain Pipeline Evolution P1-9 ALL COMPLETE". These are likely the same proposal with re-numbered phases — needs a read.
5. **Brain has 96 entries with no `status` set** — the lifecycle never closes them. Not all 107 "open" entries are real work; many are post-pipeline outputs that were never marked done.
6. **No link between proposals and brain entries**. Updating one never propagates. (Structural — addressed in the bridge proposal.)
7. **Workstream vocabulary (WS1-WS5) was lost**. active.md uses it inconsistently. Proposals don't use it at all.

## 3. Out of Scope (deferred to sibling proposals)

- **Bidirectional sync between proposals and brain entries** → [brain-vscode-bridge](../brain-vscode-bridge/main.md) (this session)
- **Token efficiency execution** → existing `token-efficiency.md` + `tokenomics-2026/main.md` (deferred per Michael)
- **Closing the 96 statusless brain entries** → needs Michael's eyes on each; mass-archive would lose real work
- **Rewriting `overview/main.md`** → too entangled with workstream re-vocabulary; do separately

## 4. Phased Plan

### Phase A — Surgical fixes (15 minutes, low risk)

Goal: stop the bleeding. These are pure dedup/edit operations.

A1. **Fix `.mind/active.md` duplication.** Trim lines 100+ (the appended Apr 20 copy already lives at `.mind/archive/active-2026-04-20.md`). Verify the surviving file ends cleanly.

A2. **Move `commission-ux-fixes.md` from active.md "In Flight" to "Recently Shipped"** with the Apr 15 date the proposal already records.

A3. **Update proposal Status lines that contradict active.md:**
   - `orchestrator-steward/main.md`: "Proposed" → "SHIPPED (Apr 10-11) — P1-6 + 86 tests"
   - `commission-ui.md`: confirm "Phase 1-3 Shipped" is fully closed (no follow-on)
   - `brain-project-ux.md`: confirm "Phase 1-4 Shipped (2026-04-14)" is closed

A4. **Stop the active.md drift loop.** Add a one-line rule at top of active.md: *"To rewrite this file: write the new content directly. Do NOT cat the existing content first — its archive snapshot is its own file."* Belt-and-suspenders against the bug that just bit us.

### Phase B — Read & Reconcile (30-60 minutes, requires Michael verification on a few)

Goal: turn the "Conflict — needs read" rows in [findings.md §3a/§4](../../scratch/cleanup-2026-04-part2/findings.md) into clear archive/keep decisions.

For each of these proposals, read the body and decide *with Michael* (one-line flag):

| Proposal | Question to answer |
|----------|--------------------|
| `brain-project-kanban.md` | Is Phase 4c shipped or still open? |
| `brain-pipeline-evolution.md` | Does P1-9 in active.md = P1-3 + P4-7 in this proposal, all shipped? |
| `brain-pipeline-fixes.md` + `brain-pipeline-fixes-phase4.md` | Are these still relevant or absorbed into evolution? |
| `project-aware-pipeline.md` | Folded into WS4 P9 already? |
| `brain-phase4-pipeline.md` | Is "WS1 P4d REST + Execution" still the next thing? |
| `brain-workspace-aware/` | Active or stale? |
| `brain-ibecome-layer2.md` | Active, shipped, or stale? |
| `enriched-indexer.md` | Superseded by gospel-engine v2? |
| `embedding-comparison.md` | Superseded? |
| `lm-studio-model-experiments/` | Superseded by `archive/context-engineering-dev.md` results? |
| `memory-architecture.md` | Superseded by `.mind/` adoption? |

For each: archive (move under `.spec/proposals/archive/`) or keep with updated Status line.

### Phase C — Workstream Re-vocabulary (single edit pass)

Goal: restore WS1-WS5 as the shared vocabulary. Tag everything once.

Define the active workstreams (proposed):

| WS | Name | Owns | Brain project |
|----|------|------|---------------|
| WS1 | Brain Core | Pipeline, steward, commissions, retry/escalation | 6 (2nd Brain) |
| WS2 | Brain UX | UI panels, dialogs, kanban, file viewer | 6 (2nd Brain) |
| WS3 | Gospel Engine | engine.ibeco.me, gospel-engine MCP, search/index, graph | 3 (Workspace improvements) |
| WS4 | study.ibeco.me | Web UI for studies, notes, reader | 5 (ibeco.me) |
| WS5 | Memory & Process | `.mind/`, agents, skills, voice, cleanup | 3 (Workspace improvements) |
| WS6 | Studies | Scripture study output | 1 (study) |
| WS7 | Teaching | YouTube content arc | 7 (YouTube/Content) |
| WS8 | Sunday School | Calling | 2 (Sunday School) |
| WS9 | Other apps | Budget app, cpuchip.net, space-center | 4, 8, 10 |

Each in-flight proposal gets a `**Workstream:** WSn` line. active.md "In Flight" gets a `WS` column.

### Phase D — active.md rewrite (replaces current after Phase A)

Once Phases A-C are done, rewrite active.md from the cleaned reality. Target shape:

```markdown
# Active Context — 2026-04-21

## Priorities (top 3-4)
## In Flight (per workstream, ≤5 items)
## Recently Shipped (rolling, last ~30 days)
## Deferred / Paused
## Key Facts (slim)
```

Drop "Recently Shipped" rows older than 30 days into a one-line summary in the relevant archive snapshot.

### Phase E — Archive snapshot

After all the above, archive the current `active.md` to `.mind/archive/active-2026-04-21.md` and start the next cycle.

## 5. Verification Criteria

| Phase | How we know it's done |
|-------|----------------------|
| A | `grep -c "^# Active Context" .mind/active.md` returns 1. `commission-ux-fixes.md` is in "Recently Shipped" not "In Flight". |
| B | Every proposal in §3a-c of findings has a clear Status line OR has been moved to `archive/`. No "Conflict" rows remain. |
| C | Every "In Flight" row in active.md has a WS tag. Every active proposal has a Workstream line. |
| D | active.md ≤ 80 lines. Every priority maps to a proposal. Every "In Flight" item has a `→` link. |
| E | `.mind/archive/active-2026-04-21.md` exists. New active.md `Updated:` date is 2026-04-22 or later. |

## 6. Costs and Risks

- **Cost:** Phase A is ~15 min. Phase B is the real work — 30-60 min, mostly reading + a few quick decisions from Michael. Phases C-E are mechanical once B is done. Total: roughly one session.
- **Risk:** Mass-archive of an actually-in-flight proposal. Mitigation: Phase B requires Michael to OK each "Conflict" row. Strict criteria: only fully shipped + verified gets archived in this pass.
- **Risk:** Workstream vocabulary doesn't fit cleanly. Mitigation: Phase C is a proposal in this proposal — Michael can revise the WS map before tagging.
- **Won't help:** This pass doesn't reduce session-start token load (that's `token-efficiency.md`). It doesn't link brain↔proposals (that's the bridge proposal). It just brings the existing systems into agreement.

## 7. Recommendation

**Build it, in order.** Phase A is small enough to do in any session opportunistically. Phases B-E should land together as one cleanup pass, after Michael has 30 min to read and OK the Phase B verdicts.

Pair this with the brain-vscode-bridge proposal so the next drift cycle is shorter.

## 8. Creation Cycle Check

| Step | Notes |
|------|-------|
| Intent | Reduce token waste re-reconciling drift every session |
| Covenant | Strict archive criteria. No mass moves without Michael's OK on conflicts. |
| Stewardship | Cleanup agent / dev agent executes; Michael owns Phase B verdicts. |
| Spiritual Creation | This spec — phases are scoped, verification criteria observable. |
| Line upon Line | Phase A is independently shippable today. |
| Physical Creation | dev agent (file edits, no code). |
| Review | grep checks per Phase. |
| Atonement | All edits reversible via git. |
| Sabbath | Natural pause after Phase A. |
| Consecration | Reduces every future session's overhead. |
| Zion | Pairs with bridge proposal for the structural fix. |
