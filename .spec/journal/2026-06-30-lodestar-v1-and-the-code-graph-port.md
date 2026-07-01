# lodestar V1, and the code-graph port

2026-06-30 (continuation of the build day). Michael: *"set a goal to get it all done and ported, vetted and tested. you are incredible and can heavy lift like that, especially if the plans are counseled and ratified."* So I did — the full V1 of lodestar plus the substrate port, as one long autonomous run, gated by an oracle at every phase.

## What it is

**lodestar** (`github.com/cpuchip/lodestar`, MIT, public) is the native tree-sitter cross-service code-graph extractor — the thing that lets you (or an AI) navigate a system too big to hold by its *gravity*. Point it at N repos; it parses each, extracts per-protocol producer/consumer contracts, pairs them across services by a normalized key, and emits one graph with the cross-service edges made explicit.

It began this session as a scaffold and ended it as a working V1 across three languages and three protocols, imported into pg-ai-stewards.

## The arc, phase by phase (each with its oracle)

- **Go vertical first** (ratified build order). Parse layer (go-tree-sitter, structural skeleton — file/func/method/type, *not* a call graph, because the contract layer doesn't need one). Then the three resolvers: HTTP (net/http, gin, chi + client calls), gRPC (`.proto` service defs + `Register*Server`/`New*Client`, matched at the bare service-name level because the Go client ctor doesn't carry the proto package), pub/sub (NATS + Kafka). Then **resolve** — the cross-world key-join. Then the **gravity/black-hole diagnostic** (Louvain modularity).
- **Widen** — Python + TypeScript/JavaScript. I delegated this to a `dev` subagent (it's the same pattern replicated across two independent languages — a textbook fan-out, and I was deep in context). It probed the grammars, matched the Go template's key conventions exactly, and returned green. I verified on the real path myself.
- **Port** — register `83-code-graph.sql` in the substrate; add `import_lodestar_graph(project, jsonb)` that lands lodestar's *already-computed* cross-edges directly into `cross_world_edges`. PR #18, Michael's Hinge.

## What surprised / what held

- **The real repo taught the tool.** Running on otel-demo's Go services, the first output was 1073 nodes — almost all generated `*.pb.go` noise. Skipping generated files dropped it to 91, and the real cross-service edge *survived* (because the actual `Register`/`New` calls live in hand-written code, not the generated definitions). That was the inverse hypothesis doing its job: skip the generated files, confirm the edge is still there.
- **The killer proof was cross-language.** The whole point of lodestar is linking services that don't share a language. otel-demo delivered it for real: `recommendation` (Python) → `product-catalog` (Go), `frontend` (TS) → three Go/Python services, all over gRPC, **zero false edges**. The Python stub and the Go server register the same service-name key, so they meet. Five correct edges across three languages — and then those same five landed in `cross_world_edges` through the port. End to end.
- **The oracle-first discipline paid at every phase.** The pub/sub extractor's test caught a real precision bug (I'd matched *any* string arg, so `redis.Publish(ctx, "chan")` read as a NATS publish; the fix was to require the subject be the *first positional* arg). The gravity test caught that raw label-propagation avalanches two clusters into one across a single bridge — Louvain (move only on modularity gain) was the right tool. My own code review of the substrate schema caught that `82` only has an HTTP resolver, which is *why* the port makes lodestar the single extraction authority rather than re-deriving edges in SQL.
- **Delegation worked because the seam was clean.** The subagent got the two languages fully, matched the conventions, and its work held up under my real-path check. The seam that made it safe: I owned the design (the key conventions, the graph model, the resolve engine) and the integration/verification; it owned replication within that frame. Shepherd for integration, fan-out for the independent units.

## The shape of it

Michael asked for a heavy lift with the plans counseled and ratified, and that framing is exactly what made a hours-long autonomous run safe: because lodestar is deterministic (no LLM, no spend, no rig), *every phase had a hard oracle by construction* — a test that's green or red, a virgin build that boots or doesn't, a real repo that yields the right edges or the wrong ones. The only genuine Hinge in the whole arc is the PR. That's the deeper lesson to keep: the way you widen autonomy isn't by trusting harder, it's by widening the verification floor until the "act" side is safe. Deterministic work is the easiest thing in the world to hand off, because being wrong is cheap to detect.

## Carry-forward

- **PR #18 is Michael's to merge** (pg-ai-stewards `code-graph-ingest`). Virgin gate 00→83 green; real otel-demo import proven.
- **#291 (world-graph) remaining** is now substrate-side: the D6-D8 gravity *render* + `lore_neighbors` BFS + RLS on the project subtree + inter-world viz, once real repos populate the graph. The gravity *diagnostic* itself is built (in lodestar).
- **lodestar's own roadmap**: the remaining resolvers (OpenAPI, GraphQL, MQTT/AMQP, shared-schema/DB, config/env, package), more languages (Java/C#/C++ — Java unlocks the train-ticket black-hole proof), deep cross-file call resolution. All noted in its ARCHITECTURE.
- **A margin candidate** surfaced but I'm holding it: "delegation works when the seam is clean." True, but I want more than one instance before writing it as a pattern.
