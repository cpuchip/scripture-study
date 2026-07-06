---
title: "Google Cloud's agentic playlist → what pg-ai-stewards should steal and skip"
source: "Google Cloud Tech — 'agentic era' conference playlist (47 videos, 39 unique digested)"
playlist: https://youtube.com/playlist?list=PLIivdWyY5sqLZY4WH03ns4Pt2B8VzE1fq
method: "39 unique talks (deduped from 47), one Sonnet digester each + an Opus synthesis (fan-out workflow, 2026-06-27)"
companion: study/yt/ai-native-databases-google-cloud.md
date_digested: 2026-06-27
tags: [pg-ai-stewards, agentic, mcp, postgres, substrate, convergence, steal-list, eval, context-engineering, governance]
relevance: high — the industry roadmap, read as a steal/skip list for an in-DB agent runtime
---

> **How to read this.** This is the playlist-level companion to the single-talk study
> (`ai-native-databases-google-cloud.md`). It was produced by fanning out an agent per talk to
> extract the take/leave for our substrate, then synthesizing. It is an **outside-in** read — the
> synthesizer saw the talks, not our codebase — so a few "steals" are things we have already shipped
> (the trajectory critic is file 56; capability-substitution is file 19; failure-clustering
> primitives are in 18). Where it names those, read it as *external validation of existing work*, not
> a gap. The value is the ranked, deduped technique list and the sharp topology rejection.

---

# pg-ai-stewards Strategy Report: The Google Cloud "Agentic Era" Playlist

*39 talks digested. The honest headline: ~28 are vendor marketing for a managed, GCP-locked agent platform (Gemini Enterprise Agent Platform / ADK / Agent Engine / AlloyDB / Cloud Run). The transferable signal concentrates in ~11 talks, and almost none of it is the topology — it's individual techniques. Below, techniques only; topology rejected.*

---

## 1. The industry view — and where it validates our thesis

Google's public architecture is consistent across every talk: **the agent runtime lives outside the database, in managed cloud compute (Cloud Run / Agent Engine), and reaches *into* data stores (AlloyDB, Spanner, BigQuery, Bigtable) through MCP adapters.** The database is the *grounding floor*, not the runtime. This is stated almost defiantly in *"Build AI agents on Cloud Run"* — AlloyDB is framed as "a reliable grounding layer beneath the LLM," precisely the passive-store role our thesis rejects.

But the *convergence underneath the marketing* validates the core bet in three concrete ways:

- **Search is collapsing into the DB.** *"Boost AI context with hybrid search in Spanner"* and *"What's new in AlloyDB"* both ship FTS + vector + graph in one engine with RRF fusion — they found Graph RAG "significantly outperforms RAG in terms of recall and precision." We already practice "search lives in the DB"; their roadmap is catching up to it.
- **State must outlive the process.** Comcast's hard-won lesson in *"Scale AI Agents in Production"* — in-process session storage kills sessions on every agent upgrade, so they bolted on a *separate* session agent — is a problem we never have. State lives in Postgres by construction.
- **Governance is moving to the data call path.** The Agent Gateway (*"Cross-cloud infrastructure for the agentic enterprise"*) intercepts every tool call to enforce scopes and DLP-redact before data reaches the model. They build this as a managed network proxy bolted *in front of* the runtime. We can build it as a `tool_dispatch` middleware step *inside* the runtime — strictly fewer moving parts.

The pattern: **every capability Google announces as a new managed *service*, we can express as a *table, function, or trigger* in the same Postgres the agents already run in.** Their architecture spends enormous effort re-externalizing state, identity, observability, and governance that an in-DB runtime gets for free. That is the validation — not their slides, their pain.

**Ben-Test caveat:** A large share of what these talks call "agentic data infrastructure" is **NL-to-SQL plus SQL-callable AI functions** — `AI.GENERATE`, `ai.if`, Conversational Analytics, BigQuery Assistant, Looker dashboard agents. *"What's new with data agents,"* *"What's new in Google Cloud databases for the agentic era,"* *"What's new in AlloyDB,"* and *"What's new in Looker"* are this in four costumes. Calling a Gemini endpoint from a SQL function is not an in-DB agent runtime; there is no loop, no queue, no autonomy. Don't let their "agentic database" branding blur our distinction — *we run the loop in the DB; they run a function that phones a model.*

---

## 2. Steal list (ranked by leverage to our substrate)

### Tier 1 — deterministic guardrails (highest leverage, smallest code)

1. **Identity-at-transport, not identity-from-payload ("parameter security").** Inject the authenticated session's user identity into the tool at call time, *bypassing the agent's payload entirely*, so prompt injection can't escalate to another user's data. Formalize as a substrate invariant: any tool touching user-scoped rows receives identity from the session, never from the model's arguments. — *"Building an AI app: A low-code guide"* (hOQ_pIFSvh0), reinforced by the Agent Gateway DLP-redaction-before-model pattern in *"Cross-cloud infrastructure"* (gY95kEL-JGI).

2. **Before-dispatch input gate (`classify_input()` trigger).** The "intercept + classify at a gate before the agent acts" pattern (Google's Model Armor) maps to a before-dispatch SQL trigger or classify function — an in-DB equivalent of the gateway, no appliance. — *"Build AI agents at scale"* (ZRs1PHngOIA), *"Cross-cloud infrastructure"* (gY95kEL-JGI).

3. **Risk-tiered action classification (1–5 enum) on the dispatch path.** Deterministic heuristics + model judgment assign a risk tier to every action; low/medium runs in background, high/very-high requires explicit approval. Gate the Hinge from *inside* `work_item_dispatch_stage`, not in the calling client. Pair it with a **named risk-assessor pipeline stage** distinct from the doer and the shepherd. — *"From prompts to production"* (7FjVoGD3K-Y, Vellum), *"Navigate the agentic shift"* (Z9Zz75pmOeg), *"Scale AI Agents in Production"* (LHcjN11nNPU, PAN's complete→clarify→handoff gates).

4. **Bounded iteration cap as substrate config, not model instruction.** PAN's RCA agent loops its DAG "up to three times" before forced synthesis. A hard substrate-level ceiling on investigation depth is a real guard against runaway loops; model-level "please don't loop" is not. — *"Scale AI Agents in Production"* (LHcjN11nNPU).

### Tier 2 — context engineering (cuts token cost at scale)

5. **Progressive disclosure / two-level tool catalog.** Ship only level-1 metadata (name + when-to-invoke) into gather context; load the full schema only on invocation. Today every granted tool ships its full schema — this is the direct fix for the **159-tool gather grant** problem already logged in the substrate. Frame the grant catalog as name+summary-in-context, body-fetched-on-dispatch; make the catalog itself query-able at dispatch time. — *"Agent context engineering for production"* (YKLkHvzjFDk), *"From prompts to production"* (7FjVoGD3K-Y, lazy catalog >200 skills), *"Build AI agents at scale"* (ZRs1PHngOIA, runtime tool discovery).

6. **Decouple ingestion from extraction.** Write session turns to a buffer on the hot path (no LLM cost); trigger engram extraction in a bgworker on inactivity/flush. The turn log is already in Postgres; the extraction call should never block the hot path. — *"Agent context engineering"* (YKLkHvzjFDk).

7. **Memory consolidation, not append.** Compare candidate engrams against the existing per-scope corpus and merge near-duplicates (cheap embedding-similarity gate) before committing. The engram subsystem currently appends — this directly reduces future context noise. — *"Agent context engineering"* (YKLkHvzjFDk).

8. **Name always-on vs just-in-time context as two modes.** Precomputed profile in system instructions (zero retrieval latency) vs memory retrieved mid-turn via tool call (relevant, +1 round-trip). Let callers pick per use-case instead of conflating them. — *"Agent context engineering"* (YKLkHvzjFDk).

### Tier 3 — eval / quality flywheel

9. **Adaptive rubrics (generator → validator).** Generate per-conversation scoring criteria *from* the trace first, *then* score the output against them — avoids one monolithic if-else judge prompt. Maps to a two-stage `judge_template`: `rubric_generator` emits criteria JSON, `rubric_validator` scores; both in-DB. — *"The agent-quality flywheel"* (eLQAJqydXqY).

10. **Rubric-seeded LLM judge, tuned to a human rater, stored as a `judge_template` row, fired post-completion.** Shopify/Sidekick: PMs write a rubric, hand-rate a sample, tune the judge until statistically indistinguishable from the human, then run at scale. — *"Build AI agents at scale"* (ZRs1PHngOIA).

11. **Failure-clustering pipeline shape.** Embed traces → cluster by similarity (pgvector already present) → triage the highest-priority *cluster*, fix the bucket not the log. Geotab's evidence: the only thing that scales past ~50 agents. Ship as a named scheduled-pipeline archetype (a "sentinel" cron pass mining recent work_item traces, pre-diagnosing clusters before users report). All primitives already shipped in 18. — *"The agent-quality flywheel"* (eLQAJqydXqY).

12. **Evaluator/optimizer decoupling as a hard rule.** Any coder/optimizer pipeline must treat the `judge_templates` it's scored by as read-only — *"a smart optimizer without an independent evaluation could be a really smarter way of gaming yourself."* (This is our eval-gaming guard, stated by Google.) — *"The agent-quality flywheel"* (eLQAJqydXqY).

13. **Trajectory evaluation, three tiers.** Exact match (tools+params+order), in-order match (extras allowed, required tools in sequence), any-order match (right tools, order irrelevant). Record a satisfying conversation as the golden set, replay it, score the tool trajectory mechanically. — *"From prototype to production: 45 minutes"* (fkCTifAqVGg).

14. **Eval-set creeping difficulty as an operational signal.** A good eval makes the *same* model score *worse* over time (Roblox: 98%→20% as the bar rose). If virgin-smoke scores are flat, the gate is stale. — *"From pilot to production"* (dBRPi41nrJw).

15. **Citation/grounding-verifier subagent.** A nested verifier inside the dispatch pipeline checks every generated claim against the source record, returns a `grounding_score`, and prunes ungrounded claims before completion. *(This is literally our own read-before-quoting covenant, mechanized into the dispatch loop — high cultural fit.)* — *"Build AI agents at scale"* (ZRs1PHngOIA, HCA Healthcare).

### Tier 4 — retrieval & dispatch determinism

16. **Hybrid retrieval as one fused function + post-vector graph expand.** Expose RRF as a single SQL function (text + vector → fused ranking) for the engram/memory surface; then walk explicit edges (work_item→corpus→chunk, engram relationships) to expand context after the initial vector hit. Evaluate **roaring bitmap** (open-source) for FTS index acceleration — no GCP dependency. — *"What's new in AlloyDB"* (vw1AzTNUiE4), *"Boost AI context with hybrid search in Spanner"* (fAf4Zh-CC08).

17. **Tiered cascade dispatch (validation of capability-substitution).** Cheap classifier / small model for routing, escalate only uncertain cases to the full model — Roblox holds P75 at 300ms; VMO2 confirms in production. Direct real-world validation of the substitution already in `work_item_dispatch_stage` (19); the steal is a deterministic SQL pre-filter *before* the model call. — *"From pilot to production"* (dBRPi41nrJw), *"Build AI agents on Cloud Run"* (zthWHEU3Y7M).

18. **Golden-queries / dispatch-hint routing table.** A catalog of (intent pattern → expected tool-call sequence) few-shot pairs that steer dispatch on high-frequency work shapes — improves determinism without prompting cold. — *"What's new with data agents"* (Z-AfOcWO_kk), *"What's new in Google Cloud databases"* (MNr7scIro9Y, query blueprints).

19. **Parameterized tools over free-form SQL.** Tools accept named params wrapping a fixed template: *"the agent doesn't have to write SQL... that gives you a lot more determinism."* Audit whether our tool defs actually enforce parameterization vs. allowing free-form compose. — *"Agent development and AgentOps with BigQuery"* (AKGV5wPQdd8).

20. **Tool-surface pruning / one-tool-per-function.** Google's finding: MCP proliferation ("everyone exposed everything") actively *hurt* agents — grant scope is a quality lever, not just a security one. Give one tool per function, test each offline in isolation. Reinforces the 159-tool gather grant as a real scaling pathology. — *"Navigate the agentic shift"* (Z9Zz75pmOeg), *"The Gemini 3 playbook"* (lbUkqPj63eQ).

### Tier 5 — observability & operational shape

21. **Dedicated `dispatch_trace` relation + LLM-judge SQL directly on it.** Log per-invocation `tool_name, latency_ms, input_tokens, output_tokens, raw_input, raw_output`; run LLM-as-judge SQL over the relation for batch eval against golden examples — *"agent logs are as important as the agent code."* — *"Agent development and AgentOps"* (AKGV5wPQdd8).

22. **Two-altitude trajectory observability.** Individual (chain-of-thought per turn, step events) + aggregate (path clustering, tool-call reliability heatmaps, recurring dead-end detection). Formalize `trajectory_event` / `work_item_step` as first-class queryable SQL surfaces, not just logs. — *"Navigate the agentic shift"* (Z9Zz75pmOeg).

23. **Conversation-exit classification (cheap diagnostic).** Two-step: get final message per session, then batch-LLM-classify the ending (resolved / new-question / wanted-to-run-query). A scheduled pipeline over ended work_items surfaces failure-mode distributions without per-trace review. — *"Agent development and AgentOps"* (AKGV5wPQdd8).

24. **Per-agent behavioral baseline / anomaly detector.** Store tool-call frequencies, avg spend, error rate per agent in the engram layer; flag statistical drift to catch the "perfectly logical path to a wrong conclusion" failure. — *"Build AI agents at scale"* (ZRs1PHngOIA).

25. **Collaborative-planning state + three-phase research shape.** `PLANNING → AWAITING_PLAN_APPROVAL → EXECUTING` gives a cheap human correction point before long token spend; the (meta-plan → research loop → verified synthesis) shape makes progress resumable at phase boundaries. Document client-reconnect-by-work-item-ID + SSE thought summaries as the resilience contract. — *"Implementing DeepMind Innovation: Deep Research API"* (05043f3GseE).

### Tier 6 — design notes worth naming (lower urgency)

26. **Intelligent batching for bulk AI calls** — re-embed engrams / score a work-item queue / classify-at-rest should batch, not invoke per-row. — *"What's new in Google Cloud databases"* (MNr7scIro9Y).
27. **Schema-semantic metadata table** keyed on `schema.table.column` storing NL descriptions + canonical aliases, fed into dispatch context — closes a text-to-SQL grounding gap without prompt bloat. — *"What's new in Cloud SQL"* (zKXbKmpqWB0).
28. **Constraint-as-catalyst** — Klarna's hard determinism constraint forced AlphaEvolve into deeper rewrites instead of cheap shortcuts. Tight bounds aren't only safety; they drive non-obvious solutions. — *"Co-Scientist and AlphaEvolve"* (nhoLefVqv6I).
29. **Nested sandbox within a warm worker** (<400ms vs cold container) — if we ever expose safe LLM code-exec, sandbox-within-bgworker beats a fresh Docker per call; microVM-per-exec is the harder-isolation next step. — *"Build AI agents on Cloud Run"* (zthWHEU3Y7M), *"What's new in Cloud Run"* (AoisAy_LGpI).
30. **OTel GenAI semantic conventions** are now standardized (model/temperature/token-count as span attributes) — track the spec for a future `pg_ai_stewards` OTel emitter that Grafana/Jaeger read without custom parsing. — *"From prototype to production: 45 minutes"* (fkCTifAqVGg).

---

## 3. Leave / wary list — adopt techniques, not topology

- **The whole managed-runtime topology.** Gemini Enterprise Agent Platform / Agent Engine / ADK-to-Cloud-Run is "hand us your runtime, we scale it." Every governance, eval, registry, and identity capability assumes the agent lives *outside* the DB in GCP. Adopting the framing concedes the entire thesis.
- **Temporal as durability layer.** Its whole value prop — durable workflows, logged LLM calls, replay, rollback — is what we get for free *by living in Postgres*. Adding it duplicates the substrate and creates a SaaS dependency.
- **Cloud-IAM / SPIFFE per-agent identity.** The shared-service-account problem is real; their fix (cloud IAM minted on deploy) maps poorly to in-DB agents that are *rows and functions, not network services*. We already solve this with per-agent SQL grants + RLS.
- **Centralized agent registry as discovery.** Google owns the registry. We own our tool catalog inside the DB. Steal "query-able catalog at dispatch time"; leave the central authority.
- **Managed MCP hosting.** "Let Google host your tool layer" (50+ GCP MCP servers) is the literal inverse of owning the runtime in Postgres.
- **Centralized-governance authority assumption.** The tiered-governance "universal floor" presumes one company with one policy team decreeing it. A self-hosted/multi-operator substrate has no such center — governance must be operator-configurable. *Steal the three-layer structure (floor / team / individual) as grant-scope layering; reject the single owner.*
- **Proprietary indexes & SQL AI functions.** ScaNN, Spanner SCAN, AI.GENERATE/ai.if/AI.RANK — all call managed Gemini or are Google-only binaries. Not portable to pgrx.
- **Unverifiable superlatives — do not cite as evidence.** "6 trillion tokens/month through ADK," "65% of Fortune 100," "Flash = Opus 4.6 parity at 10x cheaper," the context-engineering "degradation law" (a chart, no methodology). Vendor-run evals throughout — useful *intuition*, not data.
- **The "agentic database" rebrand of NL-to-SQL.** Treat "data agent" tiers and SQL-callable AI functions as what they are — not a competing runtime.
- **K8s/serving-layer plumbing.** gVisor warm pools, pod snapshots, DRA topology, disaggregated prefill/decode, KV-cache routing — problems we *sidestep* by living in Postgres. Awareness only.

---

## 4. Where pg-ai-stewards is AHEAD

The gaps in Google's story that our in-DB runtime already fills:

1. **The autonomy loop lives in the DB.** Their agents are stateless cloud compute that *call into* data. We run the loop, queue, and dispatch as bgworkers pulling a Postgres work queue — `work_item_dispatch_stage` with 4-layer routing, spend caps, and capability substitution. *"Build AI agents on Cloud Run"* even names AlloyDB "the grounding layer beneath the LLM," explicitly the passive role we invert.
2. **State outlives the process by construction.** Comcast had to *invent* a separate session agent because in-process state died on every upgrade. Our sessions, work items, and engrams are rows — upgrades can't kill them. Their hardest production lesson is our default.
3. **Memory is tables, not a managed custodian.** Their Agent Memory Bank is cloud-hosted engrams plus a third-party trusted custodian of user memory. Ours is engrams + pgvector + session tracking in the same DB, owned by the operator.
4. **Agent-to-agent coordination is in the DB.** We have a2a (register/claim/submit/answer/needs_input/inbox/receipt) as relational primitives. Google's A2A story is cloud gateways and registries between externally-hosted services. Our handoffs are transactions.
5. **Governance is SQL grants + RLS + bounds, not a network appliance.** Their Agent Gateway / Model Armor / per-agent IAM are managed proxies *in front of* the runtime. Every one maps to a trigger, a grant table, or a middleware function *inside* the runtime — fewer components, same guarantees, no vendor in the trust boundary.
6. **Observability is in the same store as execution.** They route traces to BigQuery/Looker/Cloud Trace — a second managed warehouse. Our traces are queryable SQL in the DB that ran them; LLM-as-judge runs as a query *next to* the data.
7. **Owned posture is the whole point.** Self-hosted, model-agnostic (llama-chip federation), no data leaving the operator's infrastructure. Every Google talk's "govern and optimize" means *Google watches your agents on their infra in real time*. The threat model we escape is the one they sell.

The honest framing: **Google is rebuilding, as a constellation of managed services, the properties an in-DB runtime has intrinsically.** Their roadmap is a map of what we get for free.

---

## 5. Watch-in-full shortlist

1. **`eLQAJqydXqY` — "The agent-quality flywheel."** The densest extractable talk. Adaptive rubrics (generator→validator), failure-clustering, sentinel pipeline, evaluator/optimizer decoupling.
2. **`YKLkHvzjFDk` — "Agent context engineering for production."** Progressive disclosure, ingestion/extraction decoupling, memory consolidation, always-on vs JIT modes. Directly attacks the 159-tool gather-grant problem and the append-only engram subsystem.
3. **`Z9Zz75pmOeg` — "Navigate the agentic shift in software development."** Google's *internal* 100k-dev lessons (not a product pitch): two-altitude observability, tiered governance, risk-assessor + shepherd role decomposition, and the empirical finding that tool-surface proliferation hurts agents.
4. **`ZRs1PHngOIA` — "Build AI agents at scale."** Four concrete patterns despite being the keynote: rubric-seeded LLM judge (Shopify), citation-verifier subagent (HCA — our covenant mechanized), runtime tool discovery, behavioral anomaly detection.
5. **`05043f3GseE` — "Deep Research API."** The cleanest articulation of collaborative-planning-as-a-work-item-state and the three-phase resumable research pipeline.
6. **`fkCTifAqVGg` — "From prototype to production: 45 minutes."** Trajectory evaluation in three tiers (exact / in-order / any-order) + the OTel GenAI conventions pointer.
7. **`fAf4Zh-CC08` — "Boost AI context with hybrid search in Spanner."** Short; the one real retrieval result worth internalizing — post-vector graph-edge expansion "significantly outperforms RAG in recall and precision."

*Everything else is marketing.*
