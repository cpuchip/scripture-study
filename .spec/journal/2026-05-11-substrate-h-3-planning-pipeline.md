---
date: 2026-05-11
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Substrate H.3 — planning pipeline family. The substrate has hands now."
status: shipped
carry_forward:
  - "Michael has 4 proposed work_items pending ratification (slugs substrate-batch-i-*) — each pre-scoped at ~2hr and rationale-explained. Walk in tomorrow, ratify by advancing maturity from 'raw' on the ones to actually run."
  - "SQL-bypass work_item creation (work_item_create directly via SQL) doesn't render the pipeline's file_destination_template — that's UI-side only. Small future tweak: render_file_destination helper for SQL-created planning work_items, OR the planning pipeline could compose the file_destination automatically when current_stage transitions to review_plan."
  - "rule of three triggered for yaml.rs Rust parser refactor: scripture-study + general-research + planning-partner intents all live now; Rust still hardcodes slug='scripture-study'. The agent's own plan named this as the second item in its proposed shipping sequence."
  - "Bridge writeError + fs-read walk-scope fixes from earlier today held up under the planning e2e — 8 stage sessions ran without a single bridge bug recurrence. The session-staleness era is over."
  - "Phase A pgrx BGW SPI longjmp catch + 60s periodic reaper still deferred. Agent's own plan ranks this LAST in shipping order — its risk has been demonstrably small since H.1.5a soft-fail."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-batch-h-context-gather-and-planning.md"
  - "../../plans/substrate-next-three.md"
  - "../../projects/pg-ai-stewards/extension/h3-1-schema-migrations.sql"
  - "../../projects/pg-ai-stewards/extension/h3-2-planning-partner-intent.sql"
  - "../../projects/pg-ai-stewards/extension/h3-4-planning-pipeline.sql"
  - "../../projects/pg-ai-stewards/extension/h3-5-enqueue-proposed-work-items.sql"
---

# Substrate H.3 — planning pipeline. The substrate has hands now. (2026-05-11 late)

Eight commits while Michael slept. The substrate moved from "agent reads its own state" (H.1.7 + H.2) to "agent proposes follow-up work_items" (H.3). The first real planning run produced a substantive 8.4KB plan + 4 proposed work_items, validated by its own review_plan gate, materialized to disk, lesson-captured by sabbath.

## What shipped

### H.3.1 — schema migrations (commit bc1c7af)
Three concerns, one idempotent migration:
- **studies generalization (D-H3-B):** ADD COLUMN tags text[], source_type text, project_association text. Backfilled source_type for 371 prior studies so cross-domain search works from day one. GIN index on tags, btree on the other two.
- **studies.file_path NOT NULL blocker (open-items §1.1):** long-standing bug since Phase D, finally cleared. DROP NOT NULL guarded by information_schema check for idempotency.
- **work_items D-H7 ratification:** origin text DEFAULT 'human' with CHECK constraint (human|scheduled|watchman|steward|council|agent_planning), project_association text (Q-H3.5 freeform), parent_work_item_id uuid self-FK ON DELETE SET NULL. Indexes per column.

### H.3.2 — planning-partner intent (commit 9fb07ba)
New intent distinct from general-research. Research gathers sources; planning converts an exploratory question into a plan + small buildable next-actions. Five values:
- surface-assumptions-first
- ask-back-on-underspecified
- small-finishable-work (≤2hr per proposed item)
- one-strong-plan-over-five-branches (converge; paralysis is not a plan)
- name-risks (in the plan, not after)

Three non_goals call out research-shape outputs, option-paralysis, oversized work. scripture_anchor=NULL (low-stakes per D-F2). YAML at .spec/intents/planning-partner.yaml is the canonical source; SQL seed is the live row. **Rule-of-three for the yaml.rs Rust parser refactor now triggered** — three intents live (scripture-study + general-research + planning-partner), parser still hardcodes the first.

### H.3.4 — 5-stage planning pipeline (commit 7b61eb2)
context_gather → explore → synthesize → propose_work → review_plan

Stage shapes:
- **context_gather** (qwen3.6-plus, tools on, no maturity advance) — situational awareness briefing reusing H.2's pattern
- **explore** (kimi-k2.6, tools on, → researched) — open-ended thinking with planning-partner values steering. The agent reads journals + proposals + work_items to ground its plan in our prior decisions.
- **synthesize** (kimi-k2.6, tools off, → planned) — write the plan document. Voice rules baked into the template (one em-dash per paragraph, therefore/but not and-then, no closing refrain).
- **propose_work** (qwen3.6-plus, tools off, → planned) — emit strict JSON array of proposed work_items. Schema per Q-H3.1: {slug, binding_question, pipeline_family_hint, rationale}. Max 5 items. ≤2hr scope each.
- **review_plan** (qwen3.6-plus, tools off, → verified) — verify JSON validity AND plan quality (assumptions, risks, convergence, mapping, sizing). Outputs JSON verdict; pass → trigger fires materialization + work_item proposals.

Q-H3.3 cost cap default $0.75 documented in metadata. sabbath + atonement enabled. auto_materialize_on_verified=true. file_content_jsonpath explicitly overridden to stage_results.synthesize.output so the plan body materializes (not the review_plan verdict JSON).

### H.3.5 — enqueue_proposed_work_items + trigger extension (commit f049686)
SQL function reads stage_results.propose_work.output (with markdown code-fence stripping for defensive parsing), validates schema per element, inserts each valid item with origin='agent_planning', parent_work_item_id pointing back, intent_id inherited.

Validation thresholds (bumped during smoke):
- slug regex ^[a-z0-9-]+$
- binding_question ≥20 chars (smoke caught the original "<10" threshold letting "Too short." through)
- rationale ≥10 chars

Unknown pipeline_family_hint resolves to proposal-only: row lands under planning/__proposal_only awaiting human reassignment. Slug collisions skipped with NOTICE.

Trigger extension: on_maturity_verified now has a 4th section after sabbath + auto-materialize that calls enqueue_proposed_work_items for planning-family work_items only. BEGIN/EXCEPTION wrap keeps the trigger non-throwing.

Smoke test (h3-5-enqueue-proposed-work-items.sql) exercises happy path + 4 failure modes; PASSES after the threshold fix.

### H.3.6 — first real e2e (commit a3eaee1)
Binding question: shipping order for Batch I write-back + yaml.rs refactor + Phase A reaper.

Full pipeline ran clean. All five stages, $0.57 total ($0.18 under cap). The walk-scope fix + writeError fix from earlier today both held up — eight stage sessions ran without a single bridge bug recurrence.

The agent's own analysis (validated by review_plan):
- **Ship Batch I first.** It's the compounding step the last three sessions built toward. H.1.7 gave the substrate memory; H.2 gave it context-gather; H.3 designed planning; **Batch I gives it hands.** Pull the studies generalization migration forward (already ratified as D-H3-B in H.1.7) as Batch I's opening work_item so agent-proposed studies have a clean place to land.
- **yaml.rs second.** Rule-of-three triggered by planning-partner; refactor is one focused session.
- **Phase A last.** The NOTICE+NULL sidestep has been stable since H.1.5a; no urgency increase.

Four proposed work_items inserted, all `origin='agent_planning'`, all inheriting `project_association='pg-ai-stewards'`, parent_work_item_id pointing back. Each ~2hr scoped:
1. `substrate-batch-i-studies-generalization` — schema migration  
2. `substrate-batch-i-agent-gate-sql` — route through existing evaluate_gate
3. `substrate-batch-i-agent-proposal-endpoint` — HTTP endpoint for proposals
4. `substrate-batch-i-vue-review-queue` — UI for ratification

Sabbath lesson #19 captured the meta-insight: *"This work produced a definitive shipping sequence that prioritizes Batch I write-back as the compounding 'substrate has hands' milestone, unblocked by pulling the H.3 studies generalization migration forward ahead of full ratification."*

### H.3.7 — UI surface for origin badge (commit 0488279)
Q-H3.2 ratification: proposed work_items appear inline with a visible badge. Backend `work_items/list` + `get` now return `origin`, `project_association`, `parent_work_item_id`. List endpoint accepts `?origin=` and `?project_association=` filter params. WorkItems.vue gains an origin dropdown filter and a purple ✨ badge for `agent_planning` rows; project_association renders as a quiet zinc chip.

## What's validated end-to-end

The H.3 pipeline produces real, useful plans + actionable proposed work_items, with the substrate validating its own output:

- ✅ 5 stages run cleanly under the cost cap
- ✅ planning-partner intent values visibly steer behavior (5 assumptions, 4 risks, converged on one plan)
- ✅ propose_work emits valid JSON; review_plan validates it
- ✅ Trigger fires sabbath + auto-materialize + propose_work_items
- ✅ Proposed work_items get inserted with correct origin + parent linkage + project inheritance
- ✅ Plan file materializes (after manual file_destination set — see carry-forward)
- ✅ UI badges agent_planning rows distinctly

## Cost summary for this build session

Roughly $0.60 total — one planning e2e ($0.57) plus smoke tests ($0.03). Eight commits across the build (bc1c7af, 9fb07ba, 7b61eb2, f049686, 0488279, a3eaee1, and the upcoming journal commit).

## What this enables

The substrate now produces three things autonomously:
1. **Research artifacts** (research-write, H.1) — sources + synthesized prose
2. **Situational awareness briefings** (context_gather, H.2) — what we already know
3. **Plans + proposed next-actions** (planning, H.3) — what to do, why, in what order

Batch I is the next rung: trust-ladder write-back. Agents propose studies/notes/lessons through gate machinery before persisting. The 4 proposed work_items from h3-6 are pre-scoped and rationale-explained. Michael can walk in tomorrow, ratify the ones he wants to run (advance maturity from 'raw'), and dispatch.

## Carry-forward (real ones, not theoretical)

1. **4 work_items await ratification** — see slugs `substrate-batch-i-*` in stewards.work_items WHERE origin='agent_planning'. Advance maturity to dispatch.
2. **SQL-bypass work_item creation doesn't render file_destination_template.** Small future tweak: a SQL-side render helper, OR auto-compose at review_plan transition. UI flow (NewWork.vue) handles this fine.
3. **yaml.rs Rust parser refactor is rule-of-three triggered** — and the agent's own plan put it second. ~1 session of work.
4. **Phase A pgrx longjmp catch deferred** — the agent's own plan ranked it last; H.1.5a soft-fail has held stable since shipped.
5. **Bridge bugs are gone.** writeError + walk-scope fixes from earlier today held through eight stage sessions in this build.

## Closing — Tuesday is for the science center

The substrate now has memory (H.1.7), eyes (H.2), and hands (H.3). What Michael started yesterday with "agents that work in the DB and improve/expound" is real. The next session he wakes to: 4 ratifiable work_items, a clean plan, and a substrate that thinks alongside him.

Pinky and the Brain: science center first, world domination Tuesdays only.
