---
name: substrate-batch-l-1-1-context-engine-v2-1
title: Batch L.1.1 — Context Engine v2.1 (Budget-Aware Compaction + Judge Pattern + Super-Large Data Mode)
status: RATIFIED 2026-05-14 — build in progress
created: 2026-05-14
supersedes: nothing (extends L.1)
blocks: bacteriopolis-retry, future-research-pipelines, any-pipeline-touching-large-fetches
ratifies_from: 2026-05-14 council moment + research-agent run + 3 AskUserQuestion ratification batches
---

## Ratified decisions (2026-05-14)

12 decisions across 3 batches, all by user vote:

| # | Decision | Outcome |
|---|---|---|
| 1 | Budget location | Pipeline-stage > agent > provider cascade |
| 2 | Cost cap policy | Hard cap per oversized input — $0.50 default, configurable per pipeline |
| 3 | Intercept threshold | × 0.25 of remaining budget |
| 4 | Judge cadence | Always when intercepted (one judge round per oversized result) |
| 5 | Overflow storage | Two tables — `messages_raw_overflow` (parents) + `messages_raw_overflow_leaves` (embedded chunks) |
| 6 | Contextual retrieval scope | L.1.1 new content only; studies-corpus backfill deferred |
| 7 | Judge template location | Single canonical template in stewards table; per-pipeline override available |
| 8 | Auto-merge threshold | 3 leaves under same parent → return parent |
| 9 | Index timing during intercept | Synchronous block — full index complete before tool result returns |
| 10 | Build order | Infra first (L.1.1.1-6) → then judge surface (L.1.1.7-11) |
| 11 | Verification target | Bacteriopolis retry to verified |
| 12 | Soak protocol | Pause `schedule_enabled` at session start, resume at end |

Plus prior research-confirmed (not voted but accepted):
- Chunk sizes: 4K parents / 512 leaves / 64 overlap
- Embedding model: reuse studies' 768-dim
- Splitter: recursive character (not semantic)
- Extraction style: map-reduce (not refine)
- Contextual prepend: 50-100 tok per leaf, prompt-cached source

---

# Batch L.1.1 — Context Engine v2.1

> Council artifact. Not yet ratified. Research pending.

## The reframe that makes this one batch

Treating a jumbo fetch as a mini-corpus — index it with pgvector, retrieve relevant chunks against the binding question — is the same operation we already perform over gospel-library. The substrate already knows how to do this; what it doesn't yet do is *apply* it to in-flight tool results and oversized study inputs.

Once we frame it that way:

- The four gaps (A: agent-aware threshold, B: predictive compaction, C: per-stage strategy, super-large mode) collapse into one architectural shift — "every oversized input becomes a queryable substrate, not a thing to truncate."
- "Sliding window over 30K chunks" is one chunking strategy *into* that substrate; it's not the substrate itself.
- "Engrams" become the retrieved summary tier of the per-document index, not a separate compression scheme.

## The four gaps (named)

| Gap | Where in the lifecycle | Today | What budget-aware looks like |
|---|---|---|---|
| **A — Agent-aware threshold** | Tool-result-time / Insert-time | Global 60K trigger | Threshold ratio against consuming agent's budget |
| **B — Predictive compaction** | Insert-time | React at compose-time only | Pre-insert overflow-as-corpus; raw preserved for expand |
| **C — Per-stage strategy** | Compose-time | One pressure ladder | Stage declares breadth / depth / structure |
| **Super-large mode** | Tool-result-time AND extraction-time | Single-shot extract; fails > ~200K | Slide a window, chunked extract + embed, treat as mini-corpus |

## The reframe in one diagram

```
Tool returns 500K body
         │
         ▼
   bridge interceptor (size > consuming-stage-budget × 0.25?)
         │
         ├── small → pass through (today's path)
         │
         └── large → slide window (e.g. 30K with 5K overlap)
                          │
                          ▼
                    chunk_extract(chunk_i, binding) → engrams_i
                          │
                          ▼
                    embed(chunk_i) → vector → engram_embeddings
                          │
                          ▼
                    write to messages.engrams[] (consolidated, deduped)
                          │
                          ▼
                    raw chunks → messages_raw_overflow (FK to message)
                          │
                          ▼
                    consuming agent sees engram-only render in active context
                          │
                          ▼
                    agent can: search_engrams (substrate-wide, L.3)
                             OR: expand_message tier='raw' (K.3 + L.1.1)
                             OR: query the message's own engram-mini-corpus
```

## Proposed sub-phases (draft, pre-research)

- **L.1.1.1 — Budget schema.** `working_budget` int on agents + pipelines.stages[]. Helper `effective_budget(session_id, stage_name)` walks pipeline.stage → agent → provider.context_window for the canonical answer.
- **L.1.1.2 — Agent-aware extraction threshold.** Replace constant `60000` in K.1 INSERT trigger with `effective_extraction_threshold(session_id) = working_budget / N`. (Gap A)
- **L.1.1.3 — Per-stage context strategy.** `stages[].context_strategy` enum: `breadth` (many small engrams, COLD-leaning), `depth` (fewer larger HOT engrams), `structure` (preserve specific fields). compose_messages reads it. (Gap C)
- **L.1.1.4 — Overflow corpus storage.** `messages_raw_overflow` table with FK to messages.id, stores chunked raw + per-chunk embedding. expand_message tier='raw' reads from there. (Gap B at insert; sliding-window chunks land here too)
- **L.1.1.5 — Sliding-window extraction SQL fn.** `slide_window_extract(content, binding, window_size, overlap, max_chunks)` chunks content, runs per-chunk engram extraction, dedupes engrams across windows, returns consolidated set. Cost-capped. (Super-large mode core)
- **L.1.1.6 — Bridge-side tool-result budget guard.** Bridge intercepts tool results above the consuming stage's budget × 0.25 (TBD by research), runs sliding-window extraction, replaces the tool result body with an engram-only render. Agent never sees raw 500K dumps. (Gap B at wire + super-large in flight)
- **L.1.1.7 — Subagent self-window-management.** Subagent agents can call `summarize_my_context()` (new tool) that triggers re-extraction of their own session's heavy messages with a fresher binding. Plus `expand_message tier='raw'` already reaches into overflow.

## Tensions (the ones surfaced in council, awaiting ratification)

1. **Sync vs async sliding-window extraction.** Sync (bridge blocks until extraction completes ~3-10s + cost) vs async (substrate returns placeholder, agent retries via existing work_queue). Same total cost; different agent-experience contract.

2. **Where the canonical budget lives.** Pipeline-stage > agent > provider, or agent > provider with stage as override? Currently leaning pipeline-stage-first because the same agent can run in pipelines with different budget expectations.

3. **Overflow storage shape.** Separate `messages_raw_overflow` (clean FK), or `messages.raw_overflow_id` pointer (more cohesive), or external (S3-style) for future-proof really-big bodies.

4. **Bridge intercept threshold.** Always intercept (predictable, every tool result is budget-checked), or threshold-gated (only when result > consuming-budget × 0.25)? Currently leaning threshold-gated.

5. **Sliding-window chunk size — the big tension Michael raised.** Cost/latency vs consuming-model-window:
   - K2.6 has ~260K context → could swallow 100K chunks → fewer extractions ($, latency) but lower per-chunk extraction quality?
   - Smaller-context model in the extractor → forced into 20-30K chunks → more extractions but each chunk's extraction has more room to think
   - Generalize via `chunk_size = f(extractor_model.context_window, consuming_model.context_window)` or pick 30K/5K-overlap and ship it?
   - **Michael's instinct:** lean toward picking 30K/5K and going, with the budget-aware threshold the only generalization. Don't over-generalize chunk size before we have signal.

6. **Cost discipline.** 1MB body at 30K chunks = 33 extractions ≈ $1.00 at deepseek-v4-flash. Cap per-message extraction cost (e.g., $0.50 max → if exceeded, truncate body + warn), or trust intent (anything > 1MB came from a research-mode action worth the spend)?

7. **Embedding model.** Reuse studies' 768-dim embedding (already in L.3 engram_embeddings table), or different model for in-flight chunks? Reuse keeps the semantic space unified — engram_embeddings already does this.

## Michael's reframe (the key insight)

> "we have a whole corpus of data, gospel-library, and we cannot hold all of it in memory. on a large fetch we may want to treat it the same, 'index it' with pgvector and treat it like a mini collection to mine data from."

This is the load-bearing idea. The engine doesn't need a new "super-large data" code path — it needs the same pattern we already use for the corpus, applied to per-message overflow:

- gospel-library is too big to hold → indexed once, retrieved per-query
- jumbo fetch is too big to hold → indexed at insert, retrieved per-binding-question

The substrate already has pgvector, the extraction pipeline, and the engram_embeddings retrieval surface (L.3). L.1.1 wires those existing pieces into the in-flight path.

## Research findings (2026-05-14 — research agent run)

> Citations paraphrased from the research-agent summary; specific numbers worth spot-verifying before they go load-bearing into the build doc. Directional consensus is well-established.

### Tensions resolved

- **Tension 1 (chunk strategy):** Recursive character splitting at ~512 tokens / 64 overlap is the field consensus. NAACL 2025 + Vecta 2026 + multiple benchmarks: it beats semantic chunking on realistic heterogeneous content. Skip semantic.
- **Tension 4 (map-reduce vs refine):** Map-reduce wins. Parallel, no quality penalty. Refine's sequential dependency dilutes early insights.
- **Tension 5 (chunk size scaling to consuming model — Michael's instinct):** Confirmed. The field does NOT scale retrieval chunks to consuming-model context. Adaptive sizing is real but scales with *content density* (code/tables small, prose larger), not model size. Pick fixed and go.
- **Tension 7 (embedding model):** Reuse studies' 768-dim. Confirmed — 512-token chunks with BGE-family embeddings are the empirical sweet spot, which is what the studies corpus already uses.

### Architectural sharpening — two-level (parent-child) chunking

The biggest shape change from research: **separate sizes for retrieval-side vs extraction-side**. Instead of one sliding-window doing both jobs:

- **Leaf chunks (retrieval-side):** ~512 tokens / 64 overlap, recursive splitter, embedded into pgvector.
- **Parent chunks (extraction-side):** ~4K tokens / 200-400 overlap, map-reduce engram extraction via DeepSeek V4 Flash. No embedding (parents stored, not retrieved directly).
- **At retrieval time:** vector search hits leaves; when multiple leaves under the same parent score, return the parent. LlamaIndex AutoMergingRetriever pattern.

This is exactly what we already do for gospel-library — generalized to oversized in-flight content. Michael's reframe was right.

### Highest-ROI addition — Anthropic Contextual Retrieval

Research surfaced this as a separate big-win pattern: **prepend a 50-100 token LLM-generated context blurb to each leaf chunk before embedding** (Anthropic's contextual retrieval). Reported 35-49% reduction in retrieval failures. Composes with everything else. Prompt-cache the source document so each chunk's contextualization call reuses the same prefix.

Worth considering not just for in-flight content — also for the existing studies corpus as a foundational upgrade.

### What to skip (per research)

- Semantic chunking (overrated; doesn't pay back on heterogeneous web content)
- Proposition-level extraction (43-token chunks underperform on generation)
- Scaling retrieval chunks to consuming-model context (solving the wrong problem)
- Refine-mode summarization (use map-reduce instead)
- Stuffing K2.6's full 260K on every request (Databricks: most LLMs degrade past 32K-96K context)

### Concrete shape per research

```
oversized input (e.g. 500K-1MB fetch_url result)
       │
       ▼
recursive split at 4K (parents)  ←──── stored in messages_raw_overflow (no embedding)
       │
       ▼
recursive split at 512 (leaves)
       │
       ▼
Anthropic contextual prepend (50-100 tok blurb per leaf, prompt-cached source)
       │
       ▼
embed leaves → pgvector (engram_embeddings, 768-dim)
       │
       ▼
map-reduce engram extraction over parents (DeepSeek, parallel chunks)
       │
       ▼
consuming agent: vector search hits leaves → merge to parents → 
                 either send retrieved parents to K2.6, OR for holistic queries
                 bypass retrieval entirely with map-reduce summarization
```

### Open tensions (still need ratification)

- **Tension 2 (canonical budget location):** pipeline-stage-first vs agent-first. Research didn't speak to this directly.
- **Tension 3 (overflow storage shape):** separate table vs pointer vs external. Research implies separate-table-with-FK is fine for the scale we're operating at.
- **Tension 6 (cost discipline):** cap per-message extraction cost vs trust intent. Research shows the cost is real but predictable — likely cap is right.
- **Tension 8 (sync vs async sliding-window extraction):** still open.

## Sub-phase shape (revised after research)

- **L.1.1.1 — Budget schema.** `working_budget` int on agents + pipelines.stages[]. Helper `effective_budget(session_id, stage_name)`.
- **L.1.1.2 — Agent-aware extraction threshold.** Replace constant 60000 with effective_extraction_threshold().
- **L.1.1.3 — Per-stage context strategy.** stages[].context_strategy enum: breadth | depth | structure.
- **L.1.1.4 — Overflow corpus storage (parent-child).** `messages_raw_overflow` table storing 4K parent chunks. `messages_raw_overflow_leaves` storing 512-token embedded leaves with parent FK.
- **L.1.1.5 — Contextual prepend before embedding.** New `contextualize_leaf(leaf_text, full_doc_summary)` SQL fn → LLM call → prepend before embed enqueue. Prompt-cache the doc summary.
- **L.1.1.6 — Two-level chunking SQL fn.** `chunk_and_index(content, binding, parent_size, leaf_size, overlap)` does the full recursive-split + contextualize + embed pipeline.
- **L.1.1.7 — Auto-merging retrieval fn.** `retrieve_with_merge(query, k, merge_threshold)` — vector search hits leaves; if N+ leaves under same parent score, return parent.
- **L.1.1.8 — Bridge-side tool-result budget guard.** Bridge intercepts oversized tool results, calls chunk_and_index, returns engram-only summary as the tool message.
- **L.1.1.9 — Map-reduce engram extraction.** `slide_window_extract` from earlier draft, now using parent chunks (4K) not the unified 30K. Renamed `map_reduce_extract_engrams`.
- **L.1.1.10 — Subagent self-window-management.** `summarize_my_context()` tool + expand_message tier='raw' reads from overflow.

10 sub-phases. Heavier than L was, but the architecture is more right.

## Eternal-truth council moment (2026-05-14)

Stepping back from the nuts-and-bolts. Looking for principles we may have imported by accident but not yet named, or principles we're missing.

### Already imported

| What we built | What it mirrors |
|---|---|
| Maturity ladder (raw → researched → planned → specced → executing → verified) | Line upon line, precept upon precept |
| Covenant.yaml (bilateral commitments, when_broken, recovery) | Covenant theology — D&C 82:10 |
| Watching loop (bgworker tick, Phase A) | Abraham 4:18 — "watched until they obeyed" |
| Council (Phase F proposer/critic/synthesizer) | Abraham 4:26 — "took counsel among themselves" |
| Atonement dispatch (failure → learning → forward-recovery) | Repentance, restitution |
| Sabbath dispatch (structured reflection after verified) | Sabbath rest |
| Memory architecture (identity/principles/episodes/active) | Personal becoming + journals |
| Trust scores (per-cell, line-upon-line promotion) | Stewardship grown by faithfulness |
| Stewards (the project name itself) | D&C 104 — "every man accountable to me, a steward" |

### The pattern we're missing — Judges (Exodus 18:21-22)

The captains of tens / fifties / hundreds weren't executors running Moses's script. They were **judges with authority within their stewardship**. Moses taught principles, then trusted judgment, then escalated only by *difficulty* — "every great matter they shall bring unto thee, but every small matter they shall judge."

The substrate today is mostly **executors**. Agents follow stage prompts. They make narrow decisions within prompts (e.g., engram extractor picks HOT/MEDIUM/COLD tier) but they don't *judge whether the work is theirs* or *whether to escalate this case to a higher authority*. Every work_item gets the full pipeline; cost cap is the only escalation gate.

The captains weren't spawned per dispute. They *were there*, with authority, knowing their domain.

#### Architectural shift this creates for L.1.1

**Old framing (executor pattern):** the engine measures pressure, applies rules, produces output. Agent has no say.

**New framing (judge pattern, Michael's call 2026-05-14):** the engine indexes the oversized content into a mini-corpus (using the parent-child + contextual-retrieval infrastructure from the research findings), then surfaces the situation to the consuming agent:

> "Fetched 500K from `<url>`. Indexed into per-message mini-corpus (N parent chunks, M leaf chunks, embedded). Top-level overview: `<100 tokens>`. You may: retrieve from the mini-corpus by query, expand specific sections, mark engrams as precious, or discard."

The agent then judges, applying their stewardship and binding question:

- **Is the fruit good?** (quality assessment — should this content be preserved at all? Is it on-topic, credible, well-formed?)
- **What is most precious to save?** (selection — which engrams matter for the binding question; mark them via `mark_engram_important`)
- **What should be discarded?** (an active "this is noise" judgment, not just passive timeout)

This is **expensive in LLM calls** (one or two extra rounds per oversized input — Michael ratified this cost) but **right in pattern**. It honors agency. It surfaces non-obvious quality decisions to the agent who has the binding-question context. It produces better engrams because the agent doing the work also judges what mattered.

It also retroactively re-frames existing tools:

- `mark_engram_important` (L.4) — not just "anti-loss insurance under pressure," but **active judgment of what is precious**. Reuse without modification.
- `re_extract_engrams` (L.5) — when the judge says "this isn't tuned to my actual question," they can re-extract. Already shipped; the judge pattern explains why it matters.
- `expand_message` (K.3) — the judge pulls back to raw when needed. Already shipped.

#### Sub-phase shape — revised for judges

- **L.1.1.1 — Budget schema.** (unchanged)
- **L.1.1.2 — Agent-aware extraction threshold.** (unchanged — still wanted)
- **L.1.1.3 — Per-stage context strategy.** (unchanged — still wanted for routine cases)
- **L.1.1.4 — Overflow corpus storage (parent-child).** (unchanged — this is the judge's tooling)
- **L.1.1.5 — Contextual prepend before embedding.** (unchanged)
- **L.1.1.6 — Two-level chunking SQL fn (`chunk_and_index`).** (unchanged)
- **L.1.1.7 — Auto-merging retrieval fn (`retrieve_with_merge`).** (unchanged)
- **L.1.1.8 — Judge surface for oversized tool results.** NEW. Replaces the "bridge auto-compacts" framing. When a tool result exceeds the consuming stage's budget × 0.25:
  - Index the content via L.1.1.6
  - Generate a `<100-token` top-level overview via DeepSeek (cheap, prompt-cached)
  - Surface to the consuming agent as a structured tool result with: overview + corpus metadata (parent/leaf counts, byte counts, top-level theme) + handle to a query tool (retrieve_from_corpus(corpus_id, query, k))
  - Agent judges — within the same turn it can call retrieve_from_corpus, expand specific parents, mark engrams important, or discard the corpus entirely
- **L.1.1.9 — Map-reduce engram extraction (`map_reduce_extract_engrams`).** Still needed for unattended cases (sabbath_dispatch reflection over an archive, etc.) where there's no live judge.
- **L.1.1.10 — Subagent self-window-management.** (unchanged — judge pattern at the subagent level)
- **L.1.1.11 — NEW: per-corpus judge prompt template.** Standardized template for how the judge surface presents the situation to the consuming agent — "Is the fruit good? What is most precious to save? What should be discarded?" Surfaces the three judgments explicitly, lets the agent reason about them.

### Two more eternal patterns named (lower priority, but noted)

**Multiplicity of witnesses (D&C 6:28, 2 Cor 13:1):** "in the mouth of two or three witnesses every word may be established." Our extractor is one model. Our reviewer is one model. Council pattern (Phase F) has three but is rarely used. For load-bearing extraction — especially super-large data mode where one bad chunk extraction corrupts retrieval — do we want two independent witnesses on critical engrams? Cost goes up; trustworthiness goes up. **Deferred to consideration as a Phase F extension; not in L.1.1 scope.**

**Bear one another's burdens (Mosiah 18:8):** when a subagent fails today, the parent retries or surfaces. Errors fall on whoever was assigned. No peer-agent-comes-alongside pattern before parent escalation. Subagents are isolated, not in community. **Deferred to consideration as a separate batch (likely lives near the Trust + Council layers, not Context Engine).**

## Carry-forward into ratification

Open tensions consolidate to 3-4 AskUserQuestion batches:
1. Budget location (pipeline-stage vs agent-first) + cost-cap policy
2. Sync vs async sliding-window extraction (likely resolved by the judge pattern — judge surface is sync by nature, since the agent is right there in the turn)
3. Whether to also apply contextual retrieval to existing studies corpus (scope question)
4. Sub-phase ordering and gating
5. Per-corpus judge prompt template — exact wording for "fruit good / what is precious / what to discard"
