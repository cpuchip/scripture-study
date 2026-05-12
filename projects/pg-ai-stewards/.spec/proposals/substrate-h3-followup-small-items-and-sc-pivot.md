---
title: H.3 followup — small items + science-center pivot
date: 2026-05-12
status: ratified 2026-05-12 (this session, 4 questions)
parent: substrate-batch-h-context-gather-and-planning.md
purpose: >
  H.3 shipped autonomously overnight. This proposal captures the
  followup decisions made on review: cancel the 4 substrate-internal
  Batch I work_items the planning pipeline proposed, pivot substrate
  USE to science-center planning, address three small substrate-side
  items, and add the UI surface needed to actually ratify proposed
  work_items.
---

# H.3 followup — small items + science-center pivot

## Why this proposal

Three observations from Michael after H.3 shipped:

1. **kimi-k2.6 isn't trusted yet on pgrx Rust + bgworker code.** Substrate-internal Rust work stays Claude's (external; can rebuild + restart + debug from outside the container). The substrate's USE pivots to work it CAN do safely — domain planning, research, etc.

2. **Chicken-and-egg confirmed.** The substrate can't self-modify safely because it can't restart itself if the change breaks. Claude (external) does substrate-on-substrate work; the substrate works on Michael's other projects.

3. **UI ratification gap.** H.3 ships proposed work_items with an `origin=agent_planning` badge, but there's no button to advance maturity or dispatch them. The bridge between "substrate proposes" and "human dispatches" is incomplete.

So the substrate-on-substrate Batch I work (`substrate-batch-i-*` work_items the H.3 planning pipeline proposed) is cancelled, and substrate USE pivots to the science center project.

## Ratifications

### Q1 — Order for Claude-only substrate-internal tasks
**Ratified:** `file_template → yaml.rs → Phase A` (Recommended path).

- **file_destination render helper** (this session): SQL function to render `<slug>` / `<project>` / `<id>` substitutions for SQL-bypass paths. Fixes the manual UPDATE I had to do after the H.3.6 e2e. ~30 min.
- **yaml.rs Rust parser refactor** (next session): rule-of-three triggered (scripture-study + general-research + planning-partner intents all live). Refactor `parse_yaml_intent` to read slug from yaml + accept the new `values_hierarchy:` array shape. ~1 session.
- **Phase A pgrx longjmp catch + 60s periodic reaper tick** (subsequent session): the H.1.5a `NOTICE+NULL` soft-fail in `mcp_proxy_enqueue` is the workaround; this is the durable fix. ~1 session. Ranked last because H.1.5a has held stable for the entire substrate's recent history.

### Q2 — UI ratification surface
**Ratified:** Per-item `Ratify` button on `WorkItemDetail.vue` (Recommended).

Three buttons appear on detail pages where `origin=agent_planning` AND `maturity='raw'`:
- **Ratify** — advance maturity `raw → researched`. The work_item stays at status='pending'; the user can then dispatch.
- **Dispatch** — call `work_item_dispatch_stage` to fire the current stage. Visible after ratification.
- **Cancel proposal** — set `status='cancelled'` with a `quarantine_reason` note.

API: three new POST endpoints under `/api/work-items/`: `ratify`, `dispatch`, `cancel-proposal`.

### Q3 — Disposition of the 4 Batch I work_items
**Ratified:** Cancel them.

The 4 work_items (`substrate-batch-i-studies-generalization`, `substrate-batch-i-agent-gate-sql`, `substrate-batch-i-agent-proposal-endpoint`, `substrate-batch-i-vue-review-queue`) get `status='cancelled'` with a `quarantine_reason` explaining the pivot. They stay in the DB as historical record; the cancel reason names this proposal so future Claude knows why.

### Q4 — First science-center planning question
**Ratified:** AI-literacy MVP exhibit scope (Recommended option, with Michael's correction).

Original recommended binding question used $3K. Michael's actual constraints:
- **Budget: $500** (not $3K)
- **Existing hardware:** 5 laptops (repurposed from Bridge Simulator) + one 10" ESP32 panel (waveshare-esp32-s3-specs.md in docs)
- **Existing planning:** business plans + research notes in `projects/space-center/docs/` from prior sessions

Rewritten binding question:
> *"What's the minimum-viable AI-literacy exhibit we could build for the Marsfield science center in 8 weeks with one staffer, ~$500 in new materials, and the existing hardware I have (5 laptops repurposed from the Bridge Simulator project + one 10" ESP32 panel)? Consult `/projects/space-center/docs/` for prior planning notes, ESP32 panel specs, the diy-science-exhibits research, and the existing business plans. Plan 3-5 follow-up work_items to get from MVP scope to opening day."*

Cost cap: $0.75 per Q-H3.3.

## fs-read scope expansion (required for Q4)

The planning pipeline's `context_gather` stage uses `fs-read` MCP with allow-list:
`.spec/journal/*, .spec/proposals/*, .mind/*, docs/**`

For the science-center binding question to find prior notes, expand to also include:
- `projects/space-center/*.md` (top-level — space-center-prompt.md, README.md)
- `projects/space-center/docs/**` (the 15+ research/planning notes)
- `projects/space-center/.spec/**` (scratch + proposals)

This is the **per-pipeline-scoped fs-read** Q-H3-C mentioned, simplified: one global allow-list with the union of all needed paths, scope enforcement still at the MCP tool layer. A future cleanup is real per-pipeline scoping with grants in `pipelines.fs_read_paths jsonb[]`, but the union approach unblocks today.

## What this session does

1. ✅ Write this proposal
2. ⏳ Cancel the 4 Batch I work_items (status='cancelled' + reason)
3. ⏳ Expand fs-read allow-list + restart bridge
4. ⏳ Build file_destination render helper (smallest Claude-only item)
5. ⏳ Build UI Ratify button (makes proposals actually usable)
6. ⏳ Dispatch the AI-literacy MVP planning question
7. ⏳ Journal + summary

## What this session DOES NOT do (deferred to next session)

- **yaml.rs Rust parser refactor.** Real work; deserves a focused session. pgrx build cycle is ~5min per attempt; if it breaks something, debugging takes time. Better fresh.
- **Phase A pgrx longjmp catch + reaper.** Same reasoning. H.1.5a soft-fail is stable; no urgency.
- **Full per-pipeline-scoped fs-read** (`pipelines.fs_read_paths jsonb[]`). The global-union approach unblocks Q4 today; refactor when a second pipeline needs different paths.
- **Bulk ratify / Proposals tab.** Q2 ratified the per-item button surface; bulk + dedicated tab can come if the queue gets noisy.

## Carry-forward after this session

- Michael ratifies the SC AI-literacy plan's proposed work_items (using the new buttons) — those become real next-action work in the science-center project.
- yaml.rs Rust parser refactor (~1 session, Claude)
- Phase A pgrx longjmp catch + periodic reaper (~1 session, Claude)
- Substrate-USE pipeline runs continue toward Marsfield opening: vendor evaluation, first-rotation theme plan, opening timeline, etc. The substrate plans; Michael ratifies; the substrate executes.
