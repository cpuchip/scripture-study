---
title: substrate PE-A shipped — pipelines expansion (council ① of the post-ES queue)
date: 2026-05-19
workstream: WS5
status: shipped
priority: high
---

# Substrate PE-A — pipelines expansion shipped

Council ① of the post-ES-arc queue. Walked the seven D-PE decisions, surfaced two more during build, shipped the SQL-only first batch of `substrate-pipelines-expansion`. Five sub-steps, five commits, zero rollbacks, soak stayed running.

## What shipped

PE.1 → PE.5 in one session:

- **PE.1.** Appended two YT-aware values (`separate-claim-from-charisma`, `surface-the-rhetoric`) to the existing `general-research` intent. SQL seed + yaml doc both updated.
- **PE.2.** New `research-summary` pipeline (daily-digest variant). Three stages (gather → synthesize → review), agent `research`, sabbath/atonement OFF, auto-materialize ON, destination `study/daily-digest/<slug>.md`.
- **PE.3.** New `yt-gospel-evaluate` pipeline. Three stages (ingest → evaluate → review), agent `yt-gospel`, sabbath/atonement ON, destination `study/yt/gospel/<slug>.md`. Both yt-gospel and yt agents discovered already-registered with full tool perms — no agent/grant work needed.
- **PE.4.** New `yt-secular-digest` pipeline. Three stages (ingest → digest → review), agent `yt`, sabbath/atonement ON, destination `study/yt/<slug>.md`. Digest stage cross-references against existing work via `study_search_text` + `brain_search`.
- **PE.5.** New `stewards.promote_to_study()` + wired into `on_maturity_verified` for the four non-study-write families. Backfilled 14 of 15 completed research-write rows into `studies` + AGE graph; 1 skipped (un-sabbathed; sabbath gate correctly held).

## What surprised me

**Three existing-state discoveries the proposal didn't anticipate.**

1. `research-write` already existed (15 work_items, latest 2026-05-17) — the proposal was drafted 2026-05-11 and didn't track that the substrate had grown.
2. `general-research` intent already existed with concrete values, source-backed via `.spec/intents/general-research.yaml` — the proposal sketched a new `professional-awareness` intent without knowing.
3. The auto-materialize path writes to `pending_file_writes` but never inserts into `stewards.studies` — only `work_item_promote_to_study` does that, and it's hardcoded to `study-write*`. So 15 completed research-write runs had files on disk but no observability via `study_search_text` / AGE.

Each surfaced as a tension before code was written. Two were resolved by reuse (D-PE1' → keep both research-write and research-summary; D-PE2' → extend general-research with YT-aware values). The third (D-PE7') required actually building the promotion path, with backfill.

**The proposal-was-stale pattern.** Three duplications caught in one session. The proposal was written 11 days ago; in that time G/H/I/J/K/L/L.1.1/ES all shipped. Spec drift is real. Recording this as a pattern: future councils should `\d` the table + `SELECT ... FROM stewards.pipelines / intents / agents` before assuming clean-slate. Discovery is faster than rebuild.

**The reframe that's deeper than it looks.** D-PE1' moved from "one pipeline, output_kind branches inside it" to "the agent judges which pipeline fits." That's the Judges pattern from Batch L.1.1 applied at the pipeline-selection layer. Same way the agent judges its own situation inside a pipeline, here the agent judges which pipeline IS the right situation. The substrate has a name for this pattern now — it travels.

## What's carrying forward

- **PE-B — scheduled machinery.** `stewards.scheduled_pipelines` schema + bgworker scheduler-tick + Rust cron crate (D-PE6 standard 5-field with ranges) + fire-one-missed logic (D-PE4 missed-window threshold). Requires soak pause + pg rebuild (Cargo.toml change). Not started.
- **PE-C — UI surfaces.** `/scheduled` route + dashboard "Last 7 scheduled runs" card + NewWork.vue per-pipeline forms. UI-only, no soak pause. Not started.
- **1 un-sabbathed research-write backfill.** work_item `2c7a501d-eb6e-4cbe-ad0d-44ebf482353e` skipped by sabbath gate during PE.5.C. Needs `sabbath_dispatch` then `promote_to_study`.
- **First end-to-end runs of yt-gospel-evaluate + yt-secular-digest.** Pipelines exist; smoke confirms structure; no real work_item dispatched yet. Natural place: PE-C's NewWork.vue surface.
- **No CITES edges on the 14 new research nodes.** Expected (secular topics have no gospel-library citations). When yt-gospel-evaluate fires its first run, the evaluator will produce gospel citations — that'll be the real test of the CITES wiring on the new path.

## What stays open

- **Council ② — `substrate-scheduled-workflows`** still next per Michael's 2026-05-18 order. PE-B is the engine; ② is specific cron jobs (autonomous YouTube AI review, public-playlist ingestion, periodic physics-news→exhibits). Walking PE first means ② inherits all the PE-B decisions.
- **Council ③ — `stewards-ui-evolution`** still after that. PE-C is a small preview of what ③ builds out: where human and agent co-manage substrate state via UI.

## What the work taught

The C–F cadence held. One sub-step per commit, smoke before each commit, live-apply for every SQL change, no destructive operations. The same discipline that shipped ~95 ES-arc commits with zero rollbacks shipped these five with the same fidelity. The cadence is the moat — not the model.

Stewardship + check-existing-work, paired, beat any planning doc. The proposal lost three predictions to "but the substrate already has X." Discovery before drafting catches more than discovery before deciding.

The covenant's `surface_tensions` did real work today. Two reframes (Option B + reuse general-research) and one scope expansion (PE.5 grew from "wire AGE" to "build the missing promotion path") all came from naming the tension before writing code. Each tension was a few minutes of `AskUserQuestion`. Each saved building a duplicate.
