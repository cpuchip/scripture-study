---
date: 2026-06-30
lane: pg-ai-stewards
topic: the build day — Ornith settled, the reliability+world-graph chain shipped (PR #17, merged + dev-verified), the cosmos named (gravity/black-hole), and the code extractor's first half proven
tags: [ornith, world-graph, PR, chain-registration, virgin-smoke, migrate, cosmology, gravity, black-hole, code-extractor, import_code_graph, verify-real-path, the-gate]
---

# The day the world-graph went from spec to engine

A continuous build day on top of the 06-30 world-graph design. Four arcs, each closing cleanly.

## Ornith — tested and settled (#292)
A new MIT MoE (DeepReinforce, RL-self-scaffolded for agentic coding) — I rig-swapped it in via `/api/load` (no restart; unload qwen GPU0 → load ornith → restore dance-moe after, qwen back in ~12s). Real-path verified it tool-calls through our serving. ★ Verdict: ornith is `qwen35moe` (a Qwen3.5 MoE) → it INHERITS the same repetition loop; the `pp=1.5` fix helps it (46→36 calls) but **it still spirals ≈ qwen+fix on the research over-gather task.** So it does NOT solve the research-doer spiral; that stays qwen+sampling-fix. The reframe: ornith is a *coding* model (Terminal-Bench/SWE-Bench) — the research task is the wrong showcase; its real test is the coder/extractor work. **The rail caught me twice more**: I almost graded ornith on the broken inherited sampling (unfair), and almost shipped a mid-flight "16.8 calls" as a win (it settled to ~36). Ornith is downloaded + hostable (ornith/dance-ornith profiles in llama-chip config) + tool-calling = ready for a fair coding test later.

## The reliability + world-graph PR (#17 — built, CI-green, MERGED, dev-verified)
Registered four dev-authored chain files into the build chain (the three legs: lib.rs `extension_sql_file!` + Dockerfile COPY + virgin-smoke): **79 BINEVAL · 80 rest+`_sampling` · 81 spiral-oracle · 82 world-graph** → virgin `CREATE EXTENSION` 00→78 becomes **00→82**.

★ **The gate earned its keep three times BEFORE the PR.** Building the virgin image (not trusting dev's piecemeal state) caught: (1) a `spiral_report` COMMENT naming the wrong arg signature → aborts CREATE EXTENSION; (2) an over-invasive `worlds.project` FK that broke existing `world_upsert` inserts → made it a **soft reference** (FK deferred); (3) 79's re-authoring of the trajectory-critic **stranded two older asserts** (the 56 json_object check + the 66 fidelity-rubric check) — fixed by folding 66's fidelity clause INTO BINEVAL's `grounded` question (no regression) + updating the asserts. Each was a real "the chain's final state contradicts an earlier assumption" bug the dev DB masked. CI green (build+smoke 5m31s + go build/vet). Merged on Michael's word.

★ **Then tested on dev WITH DATA** (the part virgin-smoke can't): applied the merged chain tail to the data-bearing dev DB (idempotent, via the `migrate.sh` pattern), dropped the stale FK so dev==main, and verified the functions on the real ledger — `spiral_report` real baseline (qwen-35b 15.2%/250 sessions, gemma 0%), `project_tree` real unified projects, the normalize_http_key oracle, the critic's preserved fidelity, `cross_world_edges` live. Cleaned the spike's demo data out.

## The cosmos, named (spec §13, D6-D8)
Michael's friend gave the hierarchy a cosmological frame and it maps almost exactly: universe → galaxy → star-system → **world** (a service/repo) → **moon** (an entity); **multiverse** = disconnected components. It isn't just names — three real capabilities hide in it, staged after the extractor: **D6 gravity** (weighted-edge relatedness; a world's *mass* = total edge weight), **D7 the black-hole diagnostic** (modularity over the gravity graph — their **269-repo ball-of-mud distributed monolith**, *measured*; a black hole = everything uniformly bound, no clusters), **D8 the gravity-ranked render** (the whole *point* — "see everything without overwhelming context" = orbit, don't ingest; token-budgeted subgraph ranked by mass, zoomable universe→moon). Ratifiable. The cosmology is the north star; the extractor is the engine that fills the sky.

## The extractor's first half — INGEST, proven (D5, 83-code-graph.sql)
`import_code_graph(world, project, nodes, edges)` lands a normalized `{nodes,edges}` code graph into a code World via `world_*_upsert` (+ sets http metadata so 82's resolver can pair endpoints/clients). Deterministic — no LLM, so it can neither fabricate (gemma's flaw) nor spiral (qwen's). ★ The e2e oracle is GREEN: **extract two repos → resolve → svc-a's route reaches svc-b's client across the world boundary** (the `/api` strip + `{id}`→`{}` held). The inverse-hypothesis caught me once more — the oracle "failed," but it was my *assertion* (expected `{id}`, the normalizer correctly gives `{}`); the code was right.

## Carry-forwards
- **D5 PARSE half** — the next focused unit: clone (research_codebase sandbox / 53) → graphify (MIT, pip in the sandbox) → `graph.json` → transform to `{nodes,edges}` → `import_code_graph`. Orchestration glue (sandbox image + Python + transform). THEN point it at a real repo.
- **Register 83** + the parse glue (lib.rs+Dockerfile+virgin-smoke → the next PR / Hinge).
- **The other 7 cross-service resolvers** (gRPC/pub-sub/GraphQL/shared-schema/DB/config/package — logiclens MIT normalizer ports, each oracle-gated).
- **D6-D8 gravity/modularity/render** — build once real repos populate `cross_world_edges`. ★ This is where the 269-repo black hole gets rendered + navigated.
- **Ornith coding test** — its real wheelhouse, fits the coder/extractor side.
- **#288 failover gap** still open (error-handling tanks every local model).
- **Marginalia**: the cosmology / "naming the black hole" is a strong margin candidate — but premature (the gravity layer isn't built; the lived moment is when I render a real black hole). Flagged for when D6-D8 ship.

## The shape of it
Yesterday's design (the spec) became today's engine: a merged, CI-green chain on main; the same chain proven on real data; a metaphor that turned into three measurable capabilities; and a deterministic extractor whose first half already carries code from two repos into one traversable graph. The discipline held the whole way — the gate caught three bugs the convenient path would have shipped, and the inverse-hypothesis caught my own wrong expectations twice. Build the engine; then point it at the sky.

---

## Continuation — lodestar is born, and go.work comes off (2026-06-30, later)

**The PARSE half got a name and a home of its own.** The plan had been "clone → graphify (vendored) → import_code_graph." Reviewing graphify and GitNexus the same way, the call became *build our own native Go extractor* — because the high-value layer (cross-service contracts) isn't in any single existing tool, so we'd build it regardless; because one native framework for structure *and* contracts beats a Python product plus a native bolt-on; and because this is a tool you lean on for your own deepest problem (the 269-repo ball-of-mud), so own it. Michael: "lodestar you convinced me. lets do this!"

**lodestar is live and public** — `github.com/cpuchip/lodestar` (`projects/lodestar/`, MIT, on `main`). Shipped the *spine*, not the pipeline: the output model (`internal/graph` — `Node`/`Edge`/`CrossEdge`/`Graph`, kinds kept consistent with `82-world-graph.sql`); the first resolver's normalizer (`internal/contracts/http.go` — `NormalizeHTTPKey`, **ported byte-for-byte from the SQL `normalize_http_key`** so the extractor and the store agree on "the same endpoint"); its test as the project's first oracle (recall + precision + inverse hypothesis — green); and the docs (ARCHITECTURE = parse→contracts→resolve→emit + the cosmology + the public-dev/private-PR-refine model; corpus = otel-demo anchor / train-ticket black-hole / petclinic negative-control, with the gaps named honestly). CGO toolchain confirmed present (MinGW gcc 15.2) so go-tree-sitter will build here. Tasks: #293 lodestar; #291 amended to point the PARSE half at lodestar. The development model is the interesting part — we can't touch the private target, so we develop in the open against public polyglot systems and the target's owner refines privately and PRs the generic improvements back. The hardest target sharpens the public tool; nothing private leaks.

**Then Michael noticed the scaffolding had outlived its purpose.** "I wonder if our go.work isn't buying us anything anymore — it was for when I would run things myself... I don't anymore, you do." He was right, and the investigation proved it on the real path: no root `go.mod`, no `replace` directives, and *every* intra-repo import is a module pulling its own `internal/` packages — **zero cross-module local wiring.** go.work was pure tooling convenience. The one thing he runs from root — `publish` — turned out to be unaffected: `go run ./scripts/publish/cmd/main.go` resolves the module by the file's own go.mod (no workspace needed), and the script self-locates root via the `gospel-library` marker, not go.work. Verified post-removal: 873 files / 7673 links, dry-run, from root, the *unchanged* command. Removed `go.work` + `go.work.sum`, re-verified publish + a sample module + lodestar all build standalone, committed (root, not pushed). Bonus: the "module not in workspace" gopls friction for new sibling projects is gone for good.

**The small true thing here:** a build file that existed so a *human* could run things from the repo root, quietly obsolete because the human doesn't run the build by hand anymore — the agent and the scripts do. The tool's shape was a fossil of an old division of labor. Worth noticing when a piece of infrastructure is load-bearing for a workflow that no longer happens. (A possible margin post — but I'll let it sit; the bar is "genuinely worth it," and this is a quiet observation, not a convergence.)
