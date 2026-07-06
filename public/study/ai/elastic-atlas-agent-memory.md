# Elastic "Atlas" Agent Memory — read against our own substrate

**Binding question:** Elastic open-sourced a persistent agent-memory system called Atlas. We have already built a memory system (engrams, graduated context-shedding, hybrid RRF retrieval, a self-tending memory graph). Where did two independent teams arrive at the same answer, where do we differ, and what should we take?

**Sources.** Primary: Noam Schwartz, ["Agent memory with Elasticsearch"](https://www.elastic.co/search-labs/blog/agent-memory-elasticsearch), Elastic Search Labs (2026); demo repo [`noamschwartz/atlas-memory-demo`](https://github.com/noamschwartz/atlas-memory-demo) (MIT). Pointer: [InfoQ, June 2026](https://www.infoq.com/news/2026/06/elastic-atlas-agent-memory/). Read alongside our own [anatomy-of-a-turn.md](../../projects/pg-ai-stewards-oss/docs/anatomy-of-a-turn.md), [41-memory-tend.sql](../../projects/pg-ai-stewards-oss/extension/41-memory-tend.sql), [71-hybrid-rrf.sql](../../projects/pg-ai-stewards-oss/extension/71-hybrid-rrf.sql), and the [lab-and-wiki proposal](../../projects/pg-ai-stewards-oss/.spec/proposals/lab-and-wiki.md). All Atlas numbers below are from the Search Labs blog as fetched.

---

## What Atlas actually is

A memory service that sits beside an agent and speaks MCP. It rejects the premise that a bigger context window is memory: *"A 1M-token context window is a scratchpad. It is not a memory system."* Loading full history into the prompt fails on cost, latency, and lost-in-the-middle, so Atlas keeps a persistent store the agent queries by content, time, and user.

The store is three Elasticsearch indices, one per memory type borrowed from cognitive science:

| Type | Question it answers | Contents | Lifecycle |
|---|---|---|---|
| **Episodic** | "what happened" | one document per user turn, verbatim, timestamped, written on the hot path (sub-100ms, `refresh=True`) | mostly decays; a few entries become evidence for durable facts |
| **Semantic** | "what's true" | distilled assertions ("Sarah owns a Lumio Hub v2"); fields `text`, `user_id`, `confidence`, `superseded_by`, `superseded_at`, `last_used_at` | survives across sessions; what the agent grounds in |
| **Procedural** | "what works" | multi-step playbooks; fields `success_count`, `failure_count`, `refined_steps`, `last_used_at` | accumulates outcome feedback |

**Write path.** Every user message lands raw in the episodic index before the model responds. A **consolidation** LLM (per-turn in the demo; a background job every 24h or every N events in production) reads recent episodes plus existing facts and playbooks, then emits three things: new semantic facts with `supporting_episode_ids` for provenance; new procedural playbooks when a resolution matches no existing trigger; and procedural updates that bump `success_count`/`failure_count` from whether the user confirmed the fix. Dedup narrows candidates with the same hybrid retriever, then an LLM makes the meaning judgment — a fact whose top similarity clears ≥ 0.90 is treated as a duplicate.

**Retrieval.** One agent tool, `recall_memory`, fans across all three indices at once; the agent never picks a type, because ranking and per-index decay route on its behalf. Each write is indexed twice from one document (`copy_to` sends the text into a `semantic_text` field that auto-generates Jina v5 vectors alongside the BM25 inverted index). A query fetches 80 candidates per leg, fuses BM25 and dense with RRF at `rank_constant=30` (tighter than Elastic's default 60, so top-ranked items dominate more), then a Jina v2 cross-encoder reranks the merged pool. Every turn also opens with an automatic pre-recall on the *verbatim* user message, injected as if the agent had made the call — it captures literal tokens like version numbers and error codes before paraphrasing strips them.

**Decay.** A Gauss-shaped multiplier in Painless over each index's date field: a 180-day flat zone (multiplier 1.0), then a scale of 1825 days (~5 years) to reach 0.5. Semantic memories bump `last_used_at` on recall, so "old" quietly becomes "not needed recently" — *"relevance decay, not truth decay. Truth decay is handled by supersession."* A use-count boost (`1 + log10(1 + use_count) * weight`) separates recalled-once from recalled-often. Procedural is deliberately exempt from time-decay, because bumping `last_used_at` on every recall would reward "recently tried" over "recently effective."

**Isolation.** Per-user via Elasticsearch document-level security: each user's API key carries a DLS query admitting only their documents, enforced server-side, with a redundant `user_id` filter in code as a paranoia pass. Evaluation is a 168-question QA-style passage-retrieval harness (an LLM writes two questions per doc; Recall@K measures whether the source doc appears in top-K): R@10 ≈ 0.89, R@5 ≈ 0.75, zero cross-tenant leaks, gated in CI. A follow-up promises the full LoCoMo benchmark.

---

## Where we converge

Two teams, no contact, same answers on the load-bearing questions. That convergence is the most useful signal in the whole piece, because it means these are not fashion.

**The diagnosis is identical.** Atlas's "1M-token window is a scratchpad" is our context-rot chapter almost verbatim — the same Liu *Lost in the Middle* citation anchors both. Neither of us believes long context is memory; both build a store and keep dispatches lean.

**Hybrid RRF is the retrieval floor for both.** We run real Reciprocal Rank Fusion over a lexical leg and a semantic leg in [71-hybrid-rrf.sql](../../projects/pg-ai-stewards-oss/extension/71-hybrid-rrf.sql); Atlas runs BM25 + Jina dense fused by RRF. Atlas's own justification, *"either leg alone misses cases that the other handles,"* is the exact rationale in our file's header comment. We independently landed on fuse-by-rank-position so cross-leg agreement wins.

**Raw ledger, distilled views.** Atlas keeps immutable timestamped episodes and derives semantic/procedural memory from them by LLM. That is precisely our engram model (tool results distilled at ingest, rendered as engrams instead of full text) and precisely the wiki proposal's spine: dumps are the ledger, pages are the working memory, pages are regenerable. Atlas's `supporting_episode_ids` is our `source_refs`; its consolidation LLM is our curator digester.

**Protect the literal.** Atlas pre-recalls on the verbatim message before paraphrase eats the version numbers; our context engine renders the newest messages and anything from the user *raw* for the same reason. Both teams learned that distillation must not touch the literal tokens the model will need to match.

**Server-side tenant isolation, not app-layer trust.** Atlas uses Elasticsearch DLS keyed on `user_id`, plus a redundant in-code filter against config drift. We use Postgres RLS for multi-tenancy. Same instinct, same belt-and-suspenders: the datastore refuses to return another tenant's rows rather than trusting the query to be well-formed.

---

## Where we diverge, and why ours fits a Postgres substrate

**Memory as a service vs. memory as composition.** This is the deepest split. Atlas is a separate store the agent calls over MCP — a bolt-on that any agent can plug into with no rewrite, which is exactly right for a product meant to serve Claude Desktop, Cursor, and arbitrary clients. Our memory is not called; it is *composed*. `compose_messages` renders history, engrams, and self-notes straight into the prompt by SQL, in the same transaction as the turn. There is no network hop and no second source of truth, and the whole turn stays replayable because everything is a row. Atlas's design gives portability; ours gives transactional consistency and observability. For a Postgres substrate whose thesis is "everything is a row," memory belongs inside `dry_run_chat`, not behind an endpoint.

**Two different axes of "decay."** Atlas decays at *retrieval ranking* — a Gauss multiplier down-weights old memories so they surface less, but nothing is deleted. We decay at *compose time*: graduated context-pressure sheds engram tiers at 50/70/85/95% of budget so the prompt fits. These are complementary, not competing, and the honest finding is that we have only one of them. Our `graph_recall` decays by graph *distance* (per-hop 0.5), not by *recency* or *use-count*; I found no retrieval-time recency or frequency term on `doc_search` or memory recall. Atlas has the axis we are missing.

**Autonomous consolidation vs. the Hinge.** Atlas promotes facts and supersedes them on a confidence floor alone, no human in the loop — appropriate for a personalization store where a wrong fact costs one bad support answer. Our memory-graph growth is Hinge-gated: `memory_link_propose` queues a typed edge and the edge is created only on approval ([41-memory-tend.sql](../../projects/pg-ai-stewards-oss/extension/41-memory-tend.sql)). We gate because our memory shapes covenant-governed behavior, so a confidence floor is not a strong enough guardian for the moves that matter. The lab-and-wiki proposal already names the reconciliation: a lightning/mountain split where cheap consolidations auto-apply and structural merges get review. We keep the gate on the mountain, not on every pebble.

**Procedural memory is our real gap.** Atlas has a first-class "what works" store — playbooks that carry `success_count`/`failure_count` harvested from the conversation itself ("thanks, that worked" → `success_count++`, no thumbs-up widget). Our nearest analog is the trajectory-critic and self-improvement loop, which turns recurring failures into prompt clauses: outcome feedback, but pointed at prompt tuning rather than at a rankable playbook. We have episodic (messages) and semantic (engrams, self-notes, world facts) covered well; we are thin on explicit procedural memory with outcome counters.

**Constants tuned for different horizons.** Atlas fuses at `rank_constant=30`; we use the canonical 60. Atlas's decay assumes a five-year consumer-support horizon. These are their tunings for their workload, not laws to import.

---

## Worth stealing

1. **Retrieval-time relevance decay + use-count boost, as an explicit scoring term.** This is the cleanest steal. Atlas's `1 + log10(1 + use_count) * weight`, plus a recency multiplier that bumps `last_used_at` on recall, is a few lines of SQL on top of our RRF fusion. The philosophical framing is the real gift: bump-on-recall converts "old" into "not-recently-needed," and supersession, not decay, handles truth. We have supersession-by-regeneration on the doc side already; wiring a recency/frequency term into `doc_search_hybrid` and memory recall closes the one decay axis we lack.

2. **Success/failure counters harvested by the consolidation LLM.** The conversation is the feedback signal; the classifier reads confirmation or rejection out of the next user turn and increments a counter on the playbook. This is the missing bridge between our trajectory-critic (which judges a run) and *rankable* procedural memory (which surfaces the playbook that has worked before). It needs no new UI.

3. **The verbatim pre-recall.** Auto-firing a memory/doc recall on the raw user message every turn, before the model paraphrases, is a cheap `compose_messages` or auto-tool addition with a concrete payoff on literal tokens (error codes, versions, exact names). We recompose from scratch each round already; adding a pre-recall on the raw message is a small extension of that.

4. **The three-way taxonomy as an audit lens.** Even without adopting separate indices, "episodic / semantic / procedural" is a sharp lens over what we already store: messages are episodic, engrams and self-notes and world facts are semantic, skills and tuning clauses are our thin procedural layer. Naming it that way makes the procedural gap visible and gives the wiki a vocabulary for what kind of page a consolidation produced.

---

## Reject, with reasons

1. **A separate Elasticsearch store behind MCP.** Rejected for our substrate. It re-introduces a network hop, a second source of truth, and loss of transactional consistency with the turn — the exact costs our "everything is a row" design exists to avoid. Right for a portable product, wrong for an integrated substrate. We keep memory in composition.

2. **Ungated autonomous promotion and supersession.** Rejected as the default. A confidence floor is fine for a support bot; it is not the Hinge, and our memory changes behavior under covenant. We take the mechanism (LLM consolidation with provenance) but keep it under the lightning/mountain gate.

3. **Per-turn consolidation.** Atlas itself rejects this for production (doubles LLM calls per message) and moves to a background job. We already do async engram extraction, so this is agreement, not a change.

4. **The specific decay constants.** 180-day flat zone and a five-year scale are consumer-support tunings. Take the Gauss shape and the bump-on-recall idea; set our own horizons per surface.

---

## The lab-and-wiki connection

Atlas is external validation for the [wiki half of the proposal](../../projects/pg-ai-stewards-oss/.spec/proposals/lab-and-wiki.md), and it hands us three concrete parts.

The wiki's architecture (dump everything cheaply, let a curator digester consolidate into regenerable topic pages over an immutable dump-plus-provenance layer) *is* Atlas's episodic-ledger → consolidated-views architecture, arrived at independently. Atlas's episodic index is our inbox pool; its consolidation LLM is our curator; its semantic/procedural entries are our wiki pages; its `supporting_episode_ids` is our provenance. The Karpathy-regenerable property the proposal argues for is the same property Atlas leans on when it says a better model can later re-derive better facts without losing the episodes.

Where the wiki is *ahead*: page identity. Atlas dedups facts by similarity ≥ 0.90 plus a confidence floor and auto-supersedes; it has no notion of proposing a merge or split for review. The wiki's "page identity is the hard part, propose merges via the Hinge" is the more careful design on the organization axis. Where Atlas is *ahead*: the decay and counter machinery is built and measured, and the wiki proposal has none yet. So the trade is legible — borrow Atlas's `superseded_by`/`superseded_at` fields, its ≥ 0.90 similarity + confidence-floor dedup as the auto-apply *lightning* tier, and its use-count boost so hot pages surface first; keep the Hinge on the *mountain* merges.

The fable thread lands on the lab half. The proposal's first registered experiment is the Fable-hinge A/B: Fable vs. Opus vs. `claude-p` sonnet as the top rung for hinge review, scored against Michael's own past verdicts. Atlas makes the *consolidation* LLM a load-bearing judge in its own right: it classifies "is this a duplicate fact?" and "did the user confirm the fix?" on every promotion. If Atlas-style consolidation lands in our substrate, that consolidation classifier is a rung-choosable model, and the Fable-hinge experiment shape applies unchanged — a memory-consolidation A/B on the same instrument. (I read the proposal, not the fable video Michael referenced; the connection above is drawn from the proposal's experiment, so the video may sharpen or redirect it.)

---

## Open questions

- **Do we already have a retrieval-time recency term anywhere I did not read?** I checked `graph_recall`, `doc_search_hybrid`, and the RRF files; I did not audit every recall path. Worth a grep before building the steal, so we do not duplicate.
- **Is procedural memory worth a first-class store, or does the self-improvement loop cover it?** Atlas argues playbooks-with-counters; we argue prompt-clauses-from-failures. These may be the same need met two ways, or two different needs. This is a design question for a council, not a foregone steal.
- **Consolidation cadence for the wiki curator.** Atlas settled on 24h-or-N-events after rejecting per-turn. Our digesters run async at ingest already; the wiki curator needs its own cadence decision, and Atlas's reasoning (per-turn doubles cost) is the relevant prior.
- **Should the verbatim pre-recall be a compose-time behavior or an agent tool?** Atlas injects it as if the agent called it. We could do either; the compose-time version is more reliable, the tool version is more inspectable.
