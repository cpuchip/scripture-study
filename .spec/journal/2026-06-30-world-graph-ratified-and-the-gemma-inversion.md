---
date: 2026-06-30
lane: pg-ai-stewards
topic: the cross-service world-graph (spec ratified, spike green) and the gemma inversion — the night two oracles disagreed and the disagreement was the point
tags: [world-graph, cross-service, code-graph, loreworks, graphify, logiclens, glia, gitnexus, bineval, grounding, gemma, qwen, routing, ornith, visualization, council, ratification, spike]
---

# Two oracles, one truth

This arc began as two errands — "clone graphify, find something better" and "do the gemma A/B" — and ended as one finding: **the question "which local model builds the world?" has the answer "for code, no model — a parser," and it took two orthogonal oracles to prove it.**

## The world-graph (RATIFIED, spike GREEN)
Michael wanted to roll our own cross-service code-graph — world-build microservices repos and link them *across* API boundaries (HTTP/gRPC/pub-sub). Three agents read four tools at source:
- **graphify (MIT)** — config-driven tree-sitter intra-repo extractor, no DB, no embeddings.
- **logiclens (MIT)** — cross-repo contract graph (api/event/schema), Kuzu.
- **glia + GitNexus (PolyForm-NC, study-only)** — glia's 13 cross-graph resolvers; GitNexus's `group/` bridge + fixpoint type resolution.

The convergent insight (all four arrived at it independently): **cross-service linking is a deterministic key-normalization + hash join, not fuzzy discovery.** Each protocol extracts a producer side and a consumer side; both emit a canonical key (`GET /users/{id}`); you GROUP BY the key. The hard part — normalizing `/api/users/123` → `GET /users/{id}` — is *deterministic*, so it's an oracle.

**Spec written + ratified** (`.spec/proposals/world-graph-spec.md`, D1-D5 all approved):
- **D1** project = hierarchical container (n-level `projects.parent_slug`), world = leaf graph; `worlds.project` becomes a real FK (unifies intent-projects + world-groupings).
- **D2** `cross_world_edges` (entity↔entity across world AND project boundaries).
- **D3** contracts as first-class deduped nodes — and **the existing `(world,kind,name)` dedup IS the producer/consumer matcher, for free.**
- **D4** two concepts (project hierarchy, world leaf), not recursive worlds.
- **D5** an LLM-free tree-sitter extractor for *code* worlds, replacing the LLM world-build.

**The §10 spike is GREEN** (`spike-cross-service-http.sql`, on dev): `normalize_http_key` + `resolve_cross_service_http` + three oracles pass — recall (`/users/123` ≡ `/users/{id}`), precision (`/orders`, `POST` stay distinct), inverse hypothesis, produces+consumes, and **a recursive CTE traverses svc-a's route → contract → svc-b's client across the world boundary in one query.** The thesis is proven; the rest is more resolvers (MIT ports), not architecture. Build = task #291 → a Hinge PR.

Build-vs-borrow, honest: **own the Postgres store/query** (every surveyed tool bolts on a foreign engine that fights our RLS), **port the MIT extraction+normalizers** (graphify, logiclens), **learn the taxonomy** from the restricted two.

## The gemma inversion (why we needed BINEVAL)
The spiral A/B said gemma wins the doers decisively — 12.5 calls, 0 spirals, commits, vs qwen+fix at 51+ and still grinding. The clean conclusion would have been "route world-build to gemma."

**BINEVAL flipped it.** n=4: gemma 2-3/4 at **grounding 0.0 — it fabricates, and bails when a tool errors**; qwen *all 4 grounded 1.0* — it grinds, but every claim is supported. So gemma is fast *because* it gives up and makes things up; qwen is slow *because* it actually grounds. **For graph extraction a fabricated edge is worse than a slow real one, so: don't reroute to gemma.** Keep qwen + the sampling fix.

And the two findings fused: the gemma result **validated D5**. The right builder for code isn't either local model — gemma fabricates, qwen spirals — it's the **deterministic tree-sitter extractor that can do neither.** The spiral oracle alone would have mis-routed us into quietly degraded answers; it took a *second, orthogonal* oracle (grounding) to see it. **A single metric optimizes you into a worse place.** That's the lesson of the night, and it's a margin-worthy one.

## Carry-forwards
- **#291 — the world-graph build.** D1 hierarchy + picker; promote the spike's `cross_world_edges`+resolver into a registered chain file; D5 tree-sitter extractor (shell to graphify first); the other 7 resolvers (gRPC/pub-sub/GraphQL/shared-schema/DB/config/package — logiclens MIT ports, each with its own oracle); extend `lore_neighbors` BFS to UNION cross-edges; RLS on project subtree.
- **★ Ornith-1.0-35B sub-experiment.** A new MIT MoE (DeepReinforce, dropped 06-29), post-trained on Qwen3.5+Gemma4, **RL-trained to self-scaffold its agentic harness**; benchmarks *beat* qwen3.6-35b on Terminal-Bench/SWE-Bench and it "excels in tool-calling." GGUF ~20GB Q4. **Swap qwen3.6-35b → ornith-1.0-35b on the rig and run it through the harness we just built (spiral oracle + BINEVAL grounding) on the doer task.** Hypothesis: a model RL-trained for agentic tool-use spirals less (qwen's flaw) AND grounds better (gemma's flaw) — best of both. We have the exact instruments to prove or kill it. (Simon Willison: "runs the agent harness over many tool calls proficiently," ~103 tok/s.)
- **Visualization for the world-graph.** statelyai/sketch (MIT) isn't a drop-in (it's XState-specific), but steal the pattern: **intra-world** stays the 3D force graph; **inter-world** wants a *simulatable call-flow* (click a route, animate the hop route→contract→consumer across services). Mermaid is the bridge — emit a world subgraph as Mermaid, render/simulate. Explore `sketch.stately.ai` + statechart-style libs.
- **error_handling is a shared model weakness** — both gemma and qwen tanked it (deep_research deadline, "more available" pagination). Cousin of the #288 failover gap; the tools' error/pagination UX needs a look.

## The shape of it
Two errands became one arc, and the arc taught one thing twice: measure with orthogonal oracles. The spiral oracle and BINEVAL each told half a truth, and the half-truths pointed opposite directions until you held both. The world-graph spike is the same discipline in miniature — a deterministic normalizer whose test *is* the detector. Build the oracle first; then build the second one when the first one starts agreeing with you too easily.
