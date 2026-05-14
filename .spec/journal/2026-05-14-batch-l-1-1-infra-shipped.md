---
date: 2026-05-14
mode: build (continuation — same calendar day as Batch L ship)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "Batch L.1.1 infra layer shipped — Context Engine v2.1 (Judges pattern named, 10 SQL sub-phases done, judge surface + bridge work carries forward)"
status: partial — infra (L.1.1.1-7) + judge template (L.1.1.11) + map-reduce (L.1.1.9) + subagent self-mgmt (L.1.1.10) shipped; L.1.1.8 judge surface and Go MCP/bridge work carry forward
carry_forward:
  - "L.1.1.8 judge surface — bridge-side Go work that intercepts oversized tool results, calls chunk_and_index synchronously, surfaces the judge template + corpus handle to the consuming agent. SQL helpers ready (judge_template_for_pipeline, list_overflow_parents); needs Go bridge integration."
  - "Go MCP handlers for new tools — retrieve_from_corpus, summarize_my_context (handler wraps the SQL fn we shipped today). Plus bgworker.rs additions for the 3 new completion markers: _contextualize_leaf_id (L.1.1.5), _map_reduce_extract_parent_id (L.1.1.9). Batch into one bridge rebuild."
  - "Bacteriopolis retry to verified — final verification target. Requires Go bridge work above to be live."
  - "L.3 search_engrams Go wrapper still open (carry-forward from Batch L)."
  - "chunk_and_index splitter has a slight off-by-one in find_last_break_pos — chunks are bounded and overlap covers it, but worth refining if retrieval quality signals it."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-batch-l-1-1-context-engine-v2-1.md"
  - "../../projects/pg-ai-stewards/extension/l11-budget-schema.sql"
  - "../../projects/pg-ai-stewards/extension/l12-agent-aware-extraction-threshold.sql"
  - "../../projects/pg-ai-stewards/extension/l13-per-stage-context-strategy.sql"
  - "../../projects/pg-ai-stewards/extension/l14-overflow-corpus-storage.sql"
  - "../../projects/pg-ai-stewards/extension/l15-contextual-prepend.sql"
  - "../../projects/pg-ai-stewards/extension/l16-chunk-and-index.sql"
  - "../../projects/pg-ai-stewards/extension/l17-retrieve-with-merge.sql"
  - "../../projects/pg-ai-stewards/extension/l18-judge-template.sql"
  - "../../projects/pg-ai-stewards/extension/l19-map-reduce-extract.sql"
  - "../../projects/pg-ai-stewards/extension/l20-summarize-my-context.sql"
---

# 2026-05-14 — Batch L.1.1 infra shipped (Judges pattern named)

Same calendar day as Batch L ship. Michael returned with the question "should we try kicking off the failed work item" — I noted bacteriopolis would still fail with L alone (the 170K of medium non-engram messages dominate even with crisis-mode rendering), and that led into a deeper council on what the engine was missing. We surfaced four gaps (agent-aware threshold, predictive compaction, per-stage strategy, super-large data mode), then Michael's reframe collapsed them: treat any oversized input as a mini-corpus to index, same pattern we already use for the gospel-library.

The research agent run added: two-level (parent-child) chunking, recursive character splitting, Anthropic-style contextual prepend, map-reduce > refine, fixed sizes (don't scale to consuming model).

Then the eternal-truths reflection — and that's where the session got most interesting.

## The Judges pattern (Exodus 18:21-22) as architectural principle

Michael named what was missing: the substrate is mostly executors running stage prompts. The captains of tens / fifties / hundreds in Exodus weren't executors — they were **judges with authority within their stewardship**. Moses taught principles, then trusted judgment, then escalated only by difficulty. The system worked because the judges decided.

The architectural shift this creates for L.1.1: instead of bridge-side automatic compaction (rule-driven), index the oversized content into a per-message mini-corpus, then surface the situation to the consuming agent who judges three things:
- **Is the fruit good?** (quality / on-topic / credible?)
- **What is most precious to save?** (selection → `mark_engram_important`)
- **What should be discarded?** (active noise judgment)

Michael's framing: "I think we'll get more out of the fetch if we made it inclusive and judge here with if the fruit is good, and what is most precious to save." The pg-ai-stewards project name already implied this — stewards ARE judges who act within delegated authority. Naming it now means L.1.1 is designed toward it, not retrofitted.

Two other patterns named but deferred from this batch:
- **Multiplicity of witnesses** (D&C 6:28) — for high-stakes extraction. Likely Phase F extension.
- **Bear one another's burdens** (Mosiah 18:8) — peer help before parent escalation. Likely separate batch near Trust + Council.

## Ratification — 12 decisions across 3 batches

All 12 chose recommended:
1. Budget: pipeline-stage > agent > provider cascade
2. Cost cap: hard cap $0.50/oversized, configurable per pipeline
3. Intercept threshold: × 0.25 of remaining budget
4. Judge cadence: always when intercepted
5. Overflow storage: two tables (parents + leaves)
6. Contextual retrieval scope: new only, defer studies backfill
7. Judge template: single canonical + per-pipeline override available
8. Merge threshold: 3 leaves under same parent
9. Index timing: synchronous block
10. Build order: infra first (1-7) → judge surface (8-11)
11. Verification target: bacteriopolis retry to verified
12. Soak protocol: pause/resume same as L

## What shipped this session (across both batches)

**Batch L (earlier in session)** — 10 commits including Go handlers + bin rebuild. Already journaled separately.

**Batch L.1.1 infra layer (this addition)** — 10 commits in sequence:

| Commit | Sub-phase | Contents |
|---|---|---|
| `fa15b9a` | L.1.1.1 | working_budget col on agents; effective_budget(session, stage) cascade |
| `489cbaa` | L.1.1.2 | effective_extraction_threshold replaces K.1 constant 60000 |
| `ff6c5d1` | L.1.1.3 | stages[].context_strategy + strategy_pressure_multiplier; compose_messages rewritten to use effective_budget + strategy |
| `d511c6f` | L.1.1.4 | messages_raw_overflow + messages_raw_overflow_leaves tables w/ HNSW |
| `61eb24f` | L.1.1.5 | leaf-contextualizer agent (Anthropic pattern) + contextualize_leaf + apply_contextualize_leaf |
| `dde529f` | L.1.1.6 | chunk_and_index orchestrator — paragraph-aware splitter + 4K parents / 1800 leaves with overlap; smoke 369K message → 28 parents 247 leaves |
| `b98644f` | L.1.1.7 | retrieve_with_merge (LlamaIndex AutoMergingRetriever pattern, threshold=3) + retrieve_with_merge_like_leaf |
| `a6fcb8e` | L.1.1.11 | judge_templates table + canonical 'fruit good / precious / discard' template |
| `c94aff7` | L.1.1.9 | map_reduce_extract_engrams + apply_map_reduce_parent_engrams (parallel per-parent extraction for unattended cases) |
| `1f95094` | L.1.1.10 | summarize_my_context tool — subagent self-window-management via L.5 |

## What's NOT shipped (clear carry-forward)

- **L.1.1.8 — judge surface.** Bridge-side Go work. SQL helpers ready; needs Go integration to actually intercept and surface.
- **Go MCP handlers** for retrieve_from_corpus + summarize_my_context. Plus bgworker wiring for new completion markers (_contextualize_leaf_id, _map_reduce_extract_parent_id).
- **Bacteriopolis retry to verified** — the ratified verification target. Requires the Go bridge work to be live.

## The discipline that held

Twelve commits across two batches in one extended session. The C-F pattern of smoke-and-commit per sub-phase still works at this scale, with one caveat: some sub-phases (L.1.1.7, L.1.1.9, L.1.1.10, L.1.1.11) had only SQL-application smoke, not end-to-end smoke, because their dependencies are multi-step async flows. The bacteriopolis retry is the integration test that catches what unit smokes don't — explicit carry-forward, not silent gap.

Also held: pausing soak before build, will resume at end of next session (when bacteriopolis verification is complete and the full batch is closed).

## What I'd want a future-me to remember

The judges pattern isn't decoration. It's the move from "engine compacts, agent consumes" to "engine indexes, agent judges." Every future addition to the substrate should ask: am I building an executor pattern (rule-driven, opaque to the agent) or a judge pattern (state-aware, surfaces choices)? The substrate's name (stewards) already pointed here. We finally named it.

Also: research-agent output is good directionally; verify the specific numbers before they go load-bearing into the build doc. The 35-49% Anthropic contextual retrieval reduction and the chunking benchmarks are probably right but paraphrased — spot-check if they become arguments rather than guidance.
