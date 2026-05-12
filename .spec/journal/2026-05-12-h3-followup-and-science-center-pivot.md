---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards (substrate) + space-center (USE pivot)
title: "H.3 followup — small items + science-center pivot. Substrate USE pivots to Marsfield."
status: shipped
carry_forward:
  - "5 Marsfield AI-literacy MVP work_items await Michael's ratification — slugs space-center-* with origin=agent_planning. Each gated on a real assumption from the plan. New UI Ratify button on WorkItemDetail.vue makes ratification clickable."
  - "yaml.rs Rust parser refactor — rule-of-three triggered (3 intents live: scripture-study + general-research + planning-partner; parser still hardcodes the first). ~1 session. Claude-only per kimi-trust ratification."
  - "Phase A pgrx BGW SPI longjmp catch + 60s periodic reaper tick — agent's own plan ranked last; H.1.5a soft-fail stable since shipped. ~1 session."
  - "Pre-existing naming drift: work_items.materialized_at is set at enqueue time, not actual-write time. Confusing but not breaking. Future tweak — either rename or move the timestamp to materialize-writes CLI completion."
  - "Bulk ratify + dedicated Proposals tab deferred per Q2 ratification (per-item button was sufficient for now). If the queue gets noisy, revisit."
  - "Full per-pipeline-scoped fs-read deferred — global-union allow-list is unblocking today. Add pipelines.fs_read_paths jsonb[] when a second pipeline needs different paths than planning."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-h3-followup-small-items-and-sc-pivot.md"
  - "../../plans/h3-followup-sc-ai-literacy-mvp.md"
  - "../../projects/pg-ai-stewards/extension/h3-followup-1-fs-read-sc-scope.sql"
  - "../../projects/pg-ai-stewards/extension/h3-followup-2-render-file-destination.sql"
---

# H.3 followup — small items + science-center pivot (2026-05-12)

Seven commits this session. Three concerns from Michael after H.3 shipped autonomously overnight, all handled. Plus first substrate-USE pivot delivered: the planning pipeline runs on Marsfield instead of substrate-on-substrate.

## What shipped this session

### Ratification process (commit ac71f83)
Four-question AskUserQuestion captured Michael's decisions on the small items + UI gap + pivot:
- Q1 — Order for Claude-only substrate-internal tasks: `file_template → yaml.rs → Phase A`
- Q2 — UI ratification surface: per-item button on WorkItemDetail
- Q3 — Disposition of 4 Batch I work_items: cancel
- Q4 — First science-center planning question: AI-literacy MVP (with Michael's $500 budget + existing-hardware corrections)

Proposal: `substrate-h3-followup-small-items-and-sc-pivot.md` captures the reasoning + scope.

### Housekeeping (no new commit — direct UPDATEs)
- Paused soak for build session
- Cancelled the 4 `substrate-batch-i-*` work_items with `quarantine_reason` naming the pivot
- Recorded the cancellation rationale so future Claude knows why (kimi-trust + chicken-and-egg)

### Followup #1 — fs-read scope expansion (commit a6f0878 part 1)
The planning pipeline's `context_gather` stage needs to read Michael's space-center docs for the SC binding question to be useful. Updated `stewards.mcp_servers` row for `fs-read` to include:
- `projects/space-center/*.md` (space-center-prompt.md, README.md)
- `projects/space-center/docs/**` (15+ research/planning notes)
- `projects/space-center/.spec/**` (scratch + proposals)

Bridge restarted; sandbox log confirmed the expanded scope. `node_modules/` and `firmware/build/` deliberately NOT included — the walk-allowed-filtered fix from yesterday means the agent walks only the allow-list prefixes, never the whole repo.

This is the simplified "union allow-list" approach from the proposal. A future cleanup is real per-pipeline-scoped fs-read via `pipelines.fs_read_paths jsonb[]`, but the union unblocks today.

### Followup #2 — render_file_destination helper (commit a6f0878 part 2)
The H.3.6 e2e surfaced this gap: SQL-bypass `work_item_create` leaves `file_destination` NULL, and the `on_maturity_verified` trigger then skipped `enqueue_work_item_file` because the file path was missing. Michael had to manually `UPDATE file_destination` + call `enqueue_work_item_file` to land the plan.

Fix:
- New `stewards.render_file_destination(uuid)` SQL function reads `pipeline.file_destination_template` and substitutes `<slug>` → work_item.slug, `<project>` → work_item.project_association (fallback 'misc'), `<id>` → first 8 chars of work_item.id
- `on_maturity_verified` trigger extended: when `auto_materialize` is enabled AND `file_destination` is NULL AND the pipeline has a template, auto-render first, then enqueue
- UI flow unaffected (NewWork.vue still pre-renders client-side)

The SC plan e2e (later this session) confirmed the helper works: file landed at `plans/h3-followup-sc-ai-literacy-mvp.md` without any manual intervention.

### Followup-B — UI Ratify/Dispatch/Cancel buttons (commit 7841dca)
Q2 ratification, the missing bridge between "substrate proposes" and "human dispatches."

Backend — three new POST endpoints under `/api/work-items/`, each with `origin=agent_planning` validation so the buttons can't be repurposed as general maturity-advance levers:
- `/ratify` — UPDATE maturity raw → researched
- `/dispatch` — call work_item_dispatch_stage  
- `/cancel-proposal` — UPDATE status=cancelled with quarantine_reason

Frontend — a "Proposed work" panel on WorkItemDetail.vue renders when `origin=agent_planning` AND `status != cancelled`. Shows the parent planning run as a back-link, the rationale_from_planning field, and three state-gated buttons. Reloads work_item after each action so the UI reflects new state.

### Followup #3 — first SC planning e2e (commits a1ad6c5 + ed5f498)
Binding question carried Michael's real constraints:
- $500 budget (not the $3K placeholder)
- 5 laptops from Bridge Simulator project
- 10" ESP32-P4 panel (with specs in waveshare-esp32-s3-specs.md)
- "Consult prior business plans + research the last brain developed"

Pipeline ran clean: $0.31 total (well under $0.75 cap). All 5 stages converged. Auto-render fired correctly — file landed at `plans/h3-followup-sc-ai-literacy-mvp.md` after the materialize-writes CLI ran.

**context_gather read EXACTLY the right files** (verified by inspecting tool_calls):
- `projects/space-center/.spec/scratch/im-seriously-looking-at-what-it-would-take-to-do-a-science/main.md` (the foundational doc)
- `projects/space-center/docs/diy-science-exhibits-research.md`
- `projects/space-center/docs/marshfield-research.md`
- `projects/space-center/docs/opening-timeline.md`
- `projects/space-center/docs/financial-model.md`
- `projects/space-center/docs/waveshare-esp32-s3-specs.md` (Michael called these out specifically)
- `projects/space-center/docs/exhibits/README.md`
- `projects/space-center/docs/bridge-simulator-research.md` (because 5 laptops repurposed from there)

**The plan that emerged** ("Teach the Machine" MVP):
- Software-first exhibit using browser-based ML (Teachable Machine + TensorFlow.js offline)
- 4 stations on 5 laptops: image classification with 3D-printed space props; pose classification with toddler-friendly rocket launch; "Fix It" bias-demo HTML; "Spot It" scavenger hunt
- ESP32 panel as wayfinding kiosk with LVGL+SquareLine (no inference)
- Hard $250-$350 target with $500 ceiling
- 8-week timeline split into 5 concrete phases
- Single-staffer at 10-15 hrs/week
- Specific risks named: Windows 11 webcam permissions, ESP-IDF v5.4 bleeding-edge toolchain, toddler attention loss, single-staffer recovery buffer
- The agent caught that Michael has a Bambu X1C 3D printer from docs and incorporated 3D-printed props + brackets + kiosk enclosure into the plan

**5 proposed work_items inserted, each gated on a real assumption:**
1. `space-center-laptop-webcam-ml-validation` — validates the core hardware assumption before build time
2. `space-center-esp32-lvgl-kiosk-flash` — proves the ESP32 toolchain works
3. `space-center-fix-it-bias-page-prototype` — the exhibit's core pedagogical hook
4. `space-center-3d-prop-tray-cad-test` — physical-prop / camera-frame interaction
5. `space-center-exhibit-material-budget-lock` — pins down exact shopping list within $250-$350

Each has `origin=agent_planning`, `project_association=space-center`, `parent_work_item_id` pointing back, and a `rationale_from_planning` field Michael will see at ratification.

## What this enables

Michael can walk in tomorrow, open WorkItems with the `agent_planning` filter chip, click into each work_item, and use the new Ratify/Dispatch/Cancel buttons. The 5 SC proposals are pre-scoped (~2hr each), rationale-explained, and ready to advance.

The substrate has now done what Michael originally hoped for: it read his prior planning (8 docs), thought alongside him (planning-partner intent values), proposed a concrete plan ($500/8-week/5-laptops shape), and emitted 5 specific next-action work_items.

## Carry-forward (real, not theoretical)

1. **5 SC AI-literacy work_items await ratification** — purple ✨ badges in the UI, Ratify button on each detail page. Recommended start: `space-center-laptop-webcam-ml-validation` (the hardware gate).
2. **yaml.rs Rust parser refactor** — rule-of-three triggered, agent's own plan ranked it second. ~1 session, Claude-only.
3. **Phase A pgrx BGW SPI longjmp catch + 60s periodic reaper** — agent ranked last; H.1.5a soft-fail stable. ~1 session, Claude-only.
4. **work_items.materialized_at naming drift** — set at enqueue time, not actual-write time. Confusing but not breaking. Either rename column or set in materialize-writes CLI completion.
5. **Bulk ratify + dedicated Proposals tab** — deferred per Q2 (per-item sufficient). Revisit if queue gets noisy.
6. **Full per-pipeline-scoped fs-read** — global-union unblocks today; add `pipelines.fs_read_paths jsonb[]` when a second pipeline needs different paths.

## Substrate state after this session

- **6 pipeline families**: study-write, study-write-qwen, echo-test, research-write, planning, and the cancelled stuff that doesn't count
- **6 real artifacts on disk**: 4 research pieces + 2 plans (substrate-next-three + sc-ai-literacy-mvp)
- **20 sabbath lessons** total
- **9 active intents**, three live (scripture-study, general-research, planning-partner)
- **Bridge bugs from yesterday still held** through this session's 5-stage planning run

## Closing

Yesterday: substrate has hands. Today: substrate has its first real assignment. The Marsfield AI-literacy MVP is no longer "we should plan that someday" — it's a specific 8-week build with 5 ratifiable next steps and an articulated risk surface.

Tuesday is for the science center, indeed.
