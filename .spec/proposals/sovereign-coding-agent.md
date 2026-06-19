# Garrison — A Lean, Sovereign, Local-First Coding Agent

**Date:** 2026-06-13
**Status:** **P0 RATIFIED — council CLOSED 2026-06-18** (all six open questions resolved; P1 buildable, the post-cut gate satisfied when the substrate cut executed 06-15). Refined 2026-06-14: Go-only · **embedded-SQLite default, Postgres optional** · isolated harness · MCP/JSON-RPC/HTTP/WebSocket + WASM extensions (**no gRPC**) · self-extension tiers · LM Studio + Ollama built in. 06-18 closures: two-model split starting on the pg-ai **FlexLLama** rig (qwen3.6-27b) · store = **FTS5 + chromem-go (RRF)** in P1 · greenfield-Go-util dogfood · `garrison watch` ships with spawn · core MCP-exposable from P1.
**Origin:** general-workspace session, out of the Euclid digestion and the "can pg-ai-stewards become our own opinionated CLI coding tool?" question.
**Name:** **Garrison** / `garrison-cli` — ratified 2026-06-13. From the preside study (Webster 1828 *praesidium*, the fortified position held when the field is threatened). Michael's gloss: *"the person who drives it presides."* The name scales down the chain of watches — whoever drives a Garrison, human or steward, presides over the field it holds.

---

## Binding question

What is the **leanest stack that lets Michael keep coding productively on his own hardware, with a weak local model (qwen3.6-27B class), owning the whole thing — if frontier-model and Claude Code access were gone tomorrow?** And the prior question underneath it: how do pg-ai-stewards' principles make a model that weak *trustworthy enough to ship code*?

## The heart (why this exists)

In Michael's words: *"if I lose access to claude code and frontier models, and all I am left with is something like qwen3.6-27B then I'd want a lean stack that enables me to code with my local hardware without the fuss of dealing with something I don't have full control over."*

This is not a market play and not a feature race. It is a **go-bag** — a fortified fallback position. The values are resilience, sovereignty, and control: a coding agent Michael fully owns, runs on hardware he owns, against models he runs, that survives the loss of everything rented. He doesn't love how opencode or "pi agents" are put together and hasn't tried hermes, so there is room for an opinionated alternative built on principles already proven here.

Refined in council (2026-06-13): the *pure* go-bag — nothing but a binary and a local model — is the north star, not the v1 target. Michael already runs Docker and LM Studio to develop, so v1 may lean on them, and on Postgres, for power. The invariant that actually matters is **sovereignty, not minimalism**: every prerequisite is something he owns and runs, never something rented. v1 trades minimalism for the presiding-orchestration power of pg-as-the-machine; the minimal survival mode is a later hardening.

## What already exists (we are not starting from zero)

- **`stewards-cli`** — ~1,160 LOC, twelve subcommands including **`materialize-writes`** (DB → working tree). The separation Garrison needs — *think somewhere, write to local files* — already exists in the substrate.
- **`coder-mcp` + the `code-pr` pipeline** — plan → implement → verify (`go test`) → commit → push → PR, proven end-to-end on OSS (M2). Today it runs in a sandbox clone and opens a PR. Garrison brings that capability **home**: the working tree, interactive, no clone.
- **This workspace** — `covenant.yaml`, session lanes, grounding hooks, skills, `verify-quotes`, the study-linter, the reground counter. That *is* an opinionated, principled coding harness; it just happens to be layered on Claude Code rather than shipped as a binary. It is the client-side prototype of Garrison.
- **Spin** — Michael's local-model voice front (qwen on LM Studio). The local-model gotchas are already mapped (thinking-budget behavior, non-thinking instruct models for tool loops).
- **`principles.md` → "Harness > Intelligence."** The whole bet, already written.

### Prior art: the gospel-engine lineage (local SQLite + FTS + vectors) — borrow it

Before search moved into Postgres, three predecessors in `scripts/` already
solved Garrison-standalone's exact problem — keyword + semantic retrieval from a
single owned, local store. Mine them directly rather than reinventing:

- **`scripts/gospel-mcp`** — FTS5-only keyword search (CGo `mattn/go-sqlite3`).
- **`scripts/gospel-vec`** — vector-only, built on **`philippgille/chromem-go`**
  (v0.7.0): a **pure-Go, CGo-free** embeddable vector DB — "SQLite for vector
  search." In-process, persists to per-source `.gob.gz` files, pluggable
  embedding func that it points at **LM Studio**. Brute-force cosine, fine at
  this scale.
- **`scripts/gospel-engine`** (v1) — the *combined* one. FTS5 virtual
  tables + triggers (`talks_fts`/`chapters_fts`, external-content + `'rebuild'`)
  for keyword, a pluggable semantic retriever, fused with **Reciprocal Rank
  Fusion** (`internal/search/combined.go`): rank-based, k=60, parallel
  retrievers, fall back to whichever survives if one fails. RRF sidesteps the
  unsolvable problem of normalizing FTS5-rank against cosine-similarity — it
  compares *positions*, not scores. It also keeps a relational `edges` table
  (the same relational-graph choice the substrate later validated when it
  dropped AGE) and an mtime+size `index_metadata` table for incremental reindex.

**Design implication — resolves a latent contradiction in standalone mode.**
The standalone bullet below says "optional vectors via `sqlite-vec`," but
`sqlite-vec` is a **C extension**, and pure-Go `modernc.org/sqlite` *cannot load
C extensions* — so the two cannot coexist in a "one binary, no CGo" go-bag. The
lineage already shows the resolution: keep keyword in SQLite **FTS5** (modernc
supports FTS5 natively, pure-Go), and run vectors with **chromem-go** (pure-Go,
file-persisted, LM-Studio-backed) alongside it, the two fused with **RRF**. That
preserves the sovereignty invariant exactly, and the embedding side is already
wired to a local model. (Flagged for council — the architecture decision is
theirs; this only names the option the prior art proved.)

## The thesis: Harness > Intelligence is the enabling bet

The opinionated harness is what makes a weak local model produce code worth shipping. Source-verification cut confabulation more than any model upgrade did; phased workflows beat single-pass prompts regardless of model. On a 27B local model the governance is not decoration — it is **load-bearing**. That is simultaneously why Garrison can work on local hardware *and* why it is differentiated from every other CLI agent. State it plainly: Garrison's value is **highest exactly in the survival scenario it is built for**, because the weaker the model, the more the harness matters.

## Architecture: B + C — an isolated harness, optionally substrate-backed

Garrison is **its own isolated harness**: its own model client, its own agent loop, its own store. It never depends on pg-ai-stewards being up, because that independence *is* the sovereignty requirement. The substrate is an optional backend, never a foundation.

- **The executor is a lean Go loop Michael owns.** Model backends are **built in, not plugins**: one OpenAI-compatible client (`/v1/chat/completions`) makes **LM Studio and Ollama** (and vLLM) first-class out of the box, with no per-runtime sprawl. These two ship first. (Answers open Q #3: target the OpenAI-compatible endpoint, not a single runtime.)
- **The governance is built in; the heavy harness is optional.** The principles (oracle-first verify, council/critic, gated autonomy, the watch) live in the binary. When more muscle is wanted, the substrate plugs in over MCP as a **model-runner harness** (its model catalog, capability-substitution, cost caps, dispatch/council) and/or a **shared ledger**. This is the (b) relationship — substrate presides-by-proxy, the lean loop labors — but as a power-up, never a prerequisite.

**Two run modes (refined 2026-06-14 — supersedes the earlier "v1 requires Postgres"):**

- **Garrison standalone — the default, and the real go-bag.** Go binary + a local model + an **embedded SQLite store** (pure-Go `modernc.org/sqlite`, no CGo, cross-compiles to one file). SQLite is not a degraded mode: it carries the full presiding ledger — work-item hierarchy (recursive CTEs), context/engrams (JSON + **FTS5 keyword + chromem-go vectors, RRF-fused** — both pure-Go; ratified for P1 06-18, *not* `sqlite-vec`, a C extension `modernc` can't load), dispatch/cost — for **one Garrison and the sub-agents it spawns**. Needs nothing external but a chat endpoint and a local embedding endpoint (LM Studio `text-embedding-nomic-embed-text-v1.5`, already running). This is what makes Garrison a true sovereignty tool: no Docker, no Postgres, one binary.
- **Garrison substrate-backed — the power-up.** Point Garrison at a running pg-ai-stewards over MCP for a **shared, multi-session ledger** and the substrate's full dispatch/council/catalog/cost engine. The reason to reach for Postgres is concurrency and sharing (many Garrisons, or Garrison coordinated alongside other substrate agents), not single-session work, which SQLite already covers.

**Frontier-as-luxury, never as dependency.** When Claude Code or a frontier API *is* available, Garrison may dispatch heavy steps to it as an optional stronger pair of hands. It must never need it.

The discipline across all of it: every prerequisite is something Michael **owns and runs himself** — his binary, his SQLite file, his LM Studio, and (only for the power-up) his Postgres. Sovereignty, not minimalism. Nothing rented, nothing revocable.

## What Garrison is deliberately NOT

- **Not a frontier-feature competitor.** It will not out-edit Claude Code or aider, and trying would be the losing game. The niche is governance, not edit quality.
- **Not opencode-complexity.** Lean is a hard requirement, not a preference — Michael named the dislike directly.
- **Not a standalone-agent maximalist rewrite.** Reinventing a full frontier-grade agent loop betrays the substrate's identity (presider, not executor) and drowns in tool-protocol churn. Rejected.
- **Not frontier-dependent** and **not substrate-dependent.** Garrison runs standalone on embedded SQLite; pg-ai-stewards is an optional power-up, never required. Also **not a replacement for `stewards-cli`** (a separate thing that may share libraries).
- **No gRPC, and no native `plugin`/.so.** Extensions speak MCP / JSON-RPC over stdio, HTTP, or WebSocket (Michael's call: gRPC is a hard no). Sandboxed in-process code runs as WASM. Go's native `plugin` is Windows-hostile and version-locked, so never.

## The lean core loop

`read working tree → plan (council-lite) → edit (local model) → verify (the oracle) → watch / repeat`, in small steps, with the oracle as the safety net under a weak model. Each substrate principle has a concrete job in that loop:

- **Build-the-oracle-first** → the verify gate: build + tests must pass, plus code detectors in the study-linter spirit ("cite the warrant" for code = every change carries a passing test or a named reason). The deterministic floor is what lets a 27B model be *trusted* rather than *believed*.
- **Judges, not executors** → surface decisions to Michael instead of burying them in an opaque path; the weaker the model, the more it should ask.
- **Council / D&C 88:122** → one local doer plus a critic pass (even the same model, a second adversarial look) catches what the tired doer missed. The workspace already learned that the critic loop beats per-stage gift-matching.
- **Inverse hypothesis** → after a fix: reproduce the failure, apply, confirm gone, remove, confirm it returns. "Tests pass" is not verification.
- **Gated autonomy** → human-in-the-loop by default; tighten the gate as the model weakens.
- **Presiding / watch** → Garrison watches its own sub-steps to *intent*, not just to completion.

## The presiding chain (what pg-as-the-machine really buys)

This is the capability that makes Garrison more than a lean local agent, and it is the presiding covenant made operational. The chain of watches becomes a running system:

- **Michael presides over the main agent** — he gives it a stewardship and a binding question. His attention is the top watch (the base covenant's `read_fully` / `review_same_session`).
- **The main steward presides over the sub-agents it spawns** — when it divides the work it becomes a presider in turn (`preside_under_121`, `watch_what_you_order`). It does not lose sight of the work it ordered; it watches that work to intent.
- **The ledger is the presiding instrument** — SQLite by default (one Garrison and its sub-agents), Postgres when shared. Either store holds the work-item hierarchy, the context/engram records, and dispatch/cost, so a presiding steward can actually *see the whole field*: every sub-agent's work and context, tracked and durable. (The substrate's Batch J work-items and K/L engram engine are the proven design Garrison borrows.)

Two things fall out that an in-memory-only loop cannot do:

1. **Full sub-agent tracking.** Because the work and context live in pg, the presider watches every sub-agent to intent rather than firing and forgetting. This is `watch_what_you_order` given *eyes* — the clause was always an obligation; pg is the infrastructure that makes it keepable.
2. **Fast context switching between agent modes and sessions.** Because context is durable (engrams + work-items) instead of held in a process, Garrison can suspend, resume, and switch an agent's context without losing it. The substrate's context engine, pointed at local development.

The obligations that ride with the power, named so they are not lost:

- **Tracking must surface, not just record.** Watching that no one watches is not watching. Garrison has to *show* the presider what the sub-agents are doing — a CLI/board surface for the live chain — or the ledger is a tree falling in an empty forest.
- **Spawning is gated, hardest on weak models.** A 27B model presiding over 27B sub-agents can compound errors. The fan-out discipline applies (shepherd for integration, fan-out for independent verification), and the spawn gate tightens as the model weakens. Orchestration without the oracle and critic gates is a force multiplier for mistakes.
- **The chain is accountable upward.** When a sub-agent's work goes wrong, the naming goes up the chain to the presider (`when_presiding_is_broken`). pg makes that chain auditable rather than anecdotal.

## Self-extension (how Garrison grows new capability)

pi's lesson is that an agent should extend itself; pi does it with TypeScript and npm, which is exactly the surface Michael left. Garrison does it the Go way — out-of-process and capability-gated — leaner *and* safer. Four mechanisms, cleanest first:

1. **Skills = data, not code.** A skill is a prompt/markdown file (plus an optional script). Garrison reads skill files at runtime and injects the relevant ones; the model writes a new skill and it is live next turn. No dynamic linking. This is most of "self-extension," and it is the turn-for-turn instruction change Michael already approved.
2. **Tools = subprocess over MCP.** The persistent-capability path. An extension is a separate executable speaking **MCP / JSON-RPC over stdio, HTTP, or WebSocket** (never gRPC). Crash-isolated, language-agnostic, and **hot-addable without restarting Garrison** — a new process, not a relink. The substrate attaches the same way (it already speaks MCP).
3. **Sandboxed code = WASM (wazero, pure-Go, no CGo).** When the model should write *code* that runs in-process but caged, compile it to WASM and grant it only the imports chosen. Capability-based security is the covenant's walls in executable form — the path for self-built code that runs fast and locally without being trusted wholesale.
4. **Self-recompile = the coder loop.** The model writes Go, Garrison runs `go build` + `go test`, and spawns the result as a subprocess tool (#2). The machinery already exists as the substrate's coder (code-write/build/test).

Go's lack of good native dynamic loading (`plugin`/.so is Windows-hostile and version-locked) is not a limitation here. It is what forces the process-isolated, capability-granted model — the safer one, and the one the presiding covenant already wants.

### Building a door for itself, while it works

Because subprocess/MCP tools and WASM modules hot-add without a restart, Garrison can build new capability mid-task: the model hits a wall, writes a tool, Garrison builds and registers it, and it is available next turn. The building is friction-free. The judgment is one line: **build the door in the moment; hang it with consent.** Tiers, each gated to its blast radius:

- **Tier 0 — self-instruction** (a new skill file, an updated prompt): ephemeral, reversible, data-only. Runs free.
- **Tier 1 — compose existing tools, or write a throwaway script and run it**: gated by the normal exec approval. Mostly free.
- **Tier 2 — build a NEW persistent tool that joins its own capability surface**: a new standing dominion. The door must pass the oracle (build/test), and the presider grants it standing — *"I hit a wall; here is the door I built and tested; may I install it?"*
- **Tier 3 — rewrite its own core loop**: off the table for autonomous; a human-only operation if ever.

This is `dominion_in_council` applied to the agent itself. Self-extension is the sovereignty payoff: when you cannot `npm install` a missing capability, the agent building its own tool is how it adapts, so Garrison *wants* the power and merely declines to *seize* it. It organizes new capability from existing materials (Abraham 4: organize, not create from nothing), watches it until the oracle passes it, and receives dominion in council. The weaker the model, the tighter the gate, because a 27B model's self-built door is likelier to be subtly wrong. The garrison may build its own fortifications; it does not expand its commission without orders from up the chain.

## Why governance is load-bearing here, not luxury

A 27B model hallucinates more, plans worse over long horizons, and drifts faster. The harness compensates, mechanically: decompose into steps small enough that a weak model rarely goes wrong; gate every step behind a hard oracle (build/test/lint) it cannot talk its way past; add a critic pass to catch the doer's misses; keep the autonomy gate tight so a human confirms the consequential moves. Strip those four away and a local-model coding agent is a liability. Keep them and it becomes usable. They are the product.

## Local-model design constraints (from Spin + memory)

- qwen3.6-27B on LM Studio always reasons; a small `max_tokens` yields empty `content` with `finish_reason=length`. Give it ≥2000 tokens; the answer is in `content`, the reasoning in `reasoning_content`. Tool-calling on local models is weaker and inconsistent.
- Design for that reality: structured output with a forgiving parser, retries, and possibly a split — non-thinking instruct models for the tool loop, reasoners reserved for planning. Distrust a negative result from a parser written in haste (the verify-via-real-path lesson applies double here).

### Starting model configuration (borrowed from the pg-ai FlexLLama rig, 2026-06-18)

Garrison borrows the model setup the pg-ai-stewards work already runs, so the daily driver and the go-bag share one rig from day one. **FlexLLama** (`external_context/flexllama`) is a local dual-4090 llama.cpp manager exposing **one OpenAI-compatible endpoint** that routes by model alias to per-GPU runners — exactly Garrison's "one client, model-by-alias" design. The active `stewards-3way.json` (all Q4, `n_ctx` 32768, `jinja: true` so tool-call templates work):

| Alias | Model | GPU | Role in Garrison |
|---|---|---|---|
| `qwen3.6-27b` | Qwen3.6-27B Q4_K_M | 0 | planner / reasoner + initial doer |
| `gemma-12b` | gemma-4-12B-it QAT Q4_0 | 1 | mid instruct (alt doer / critic) |
| `nemotron-4b` | NVIDIA-Nemotron-3-Nano-4B Q4_K_M | 1 | fast small loop model |

- **Chat endpoint:** `http://localhost:8090/v1` for a native Garrison binary (the containerized substrate uses `http://host.docker.internal:8090/v1`). API key is the placeholder `flexllama`; health at `/health`; `request_timeout` 1800s.
- **Embedding endpoint (for chromem-go vectors):** LM Studio `text-embedding-nomic-embed-text-v1.5` at `http://localhost:1234/v1` (placeholder key `lmstudio`) — 768-dim, matches gospel-engine-v2's nomic embeddings, so retrieval is apples-to-apples. (Ollama isn't installed on the rig — LM Studio serves both chat and embeddings; corrected 06-18.)
- **Start model:** `qwen3.6-27b` (Michael's stated floor); the doer/planner split maps onto the three hot aliases. Devstral Small 2 becomes a 4th FlexLLama alias when the tool-tuned-doer upgrade is wanted.

Recorded here (not in code) because P1 isn't scaffolded yet; P1's first act is to seed `projects/garrison/.env.example` from this table.

## Tensions and risks (honest)

- **Capability floor.** It will not match Claude Code, full stop. The bet is "good enough, fully owned, always available," not "best." Name it so no one is surprised.
- **The yet-another-agent trap.** Mitigated by staying lean, owning the governance niche, and the dogfood test: in survival mode Michael uses it by necessity; in luxury mode he would only reach for it *for the governance*. If the honest answer in luxury mode is "I wouldn't," that argues for keeping Garrison small and the substrate-as-MCP path primary.
- **Maintenance.** Owning the whole stack is a real ongoing cost; lean and library-reuse are the only defenses.
- **Effort vs. the parity roadmap.** Garrison is post-cut. Spec now, build later. Do not fork the parity push.

## Phasing (post-parity / post-cut)

- **P0** — this spec + council ratification (`dominion_in_council`).
- **P1** — the standalone MVP. **[✅ COMPLETE 2026-06-18 — `cpuchip/garrison` (private), `projects/garrison/`. Floor (one OpenAI-compatible client, ping/chat/embed live) → G1 SQLite presiding ledger (work-item hierarchy/recursive-CTE Tree, FK'd messages, cost rollup) → G2 FTS5 + chromem-go (RRF) retrieval (live fusion test green) → G3 read→plan→edit→verify loop (Dispatcher, forgiving ===FILE=== parser, path-safe apply, the Verify oracle, `garrison run`, every step to the ledger) → G4 self-extension Tiers 0–1 (skills-as-data + gated exec tool + MCP-client stub) → DOGFOOD reached + independently verified TWICE (qwen3.6-27b wrote strutil and mathx through the loop, oracle-passed in 1 attempt each; `docs/dogfood-01.md`). Every package build+vet+test green; ~9 tested commits. Remaining: wire retrieval into the loop's read-step for non-empty trees (greenfield needs none).]** Lean Go loop + OpenAI-compatible client (the **FlexLLama :8090 rig**, `qwen3.6-27b` to start — see *Starting model configuration*) + embedded SQLite ledger **with FTS5 + chromem-go (RRF) retrieval**. Read / plan / edit / verify on the working tree; the presiding ledger comes online here (work-item + context tracking for any sub-agents). **No Docker** (FlexLLama + Ollama already run on the host). Self-extension Tiers 0–1 (skills-as-data + a general exec tool + the MCP client); the core is designed MCP-exposable. Dogfood = a **greenfield Go utility + tests** (hard `go build && go test` oracle); drive pi + qwen3.6-27b on that same task as the baseline first.
- **P2** — the code oracle suite: a build/test wrapper plus code detectors reusing the `verify-quotes` / study-linter patterns (precision-tuned, oracle-first).
- **P3** — the council/critic pass (the D&C 88:122 lever).
- **P4** — the substrate-backed power-up: point Garrison at pg-ai-stewards over MCP for the shared ledger + dispatch/council/catalog. Optional, never required.
- **P5** — self-extension Tier 2 (self-built persistent tools, WASM-sandboxed, behind the build-the-door / hang-with-consent gate). Its own council item, since it is a new standing capability.
- **P6** — package and share; ties to `plugin-someday` and ai-jumpstart / *Beyond the Prompt*. Garrison is the tool that practices what the book preaches.

## Relationship to existing assets

`stewards-cli` (sibling; shares libraries; `materialize-writes` is the seed of the local-write path) · `coder-mcp` (the capability brought home from the sandbox) · the Claude Code workspace layer (the client-side prototype) · `plugin-someday` (the P5 packaging) · ai-jumpstart / *Beyond the Prompt* (Garrison as the embodied companion) · Spin (the local-model sibling that already pathfound the runtime gotchas).

## Decided in council — P0 CLOSED (2026-06-18)

> Companion research (evidence for the model + mechanism calls): [Garrison — Landscape & Design Inputs](./sovereign-coding-agent-landscape.md) (2026-06-14). Headline: **pi** proves the four-tool lean core ships; **goose** is the MCP-framework cousin minus the governance; **Devstral Small 2** is the tool-tuned local model that answers the weak-model tool-calling risk; the governance niche is empirically empty.

**Decided 2026-06-13/14:** name (Garrison / `garrison-cli`) · Go-only · isolated harness · **embedded-SQLite default**, Postgres as the optional shared-ledger / power-up backend · model backends built in via the OpenAI-compatible endpoint (**LM Studio + Ollama first**) · extensions over **MCP / JSON-RPC / HTTP / WebSocket + WASM** (no gRPC, no native `plugin`) · self-extension Tiers 0–3 with the build-the-door / hang-with-consent gate.

**Closed 2026-06-18 — the six open questions resolved (post-cut gate satisfied 06-15):**

1. **Weak-model tool-calling / default pairing.** A **two-model split** — a tool-tuned/fast model runs the edit-verify *loop*, a reasoner runs the *planning* step — both local, one OpenAI-compatible client, routed by model alias. **Start on the pg-ai FlexLLama rig** (`qwen3.6-27b` + `gemma-12b` + `nemotron-4b` already hot on one `:8090` endpoint — see *Starting model configuration*); Devstral Small 2 is the documented tool-tuned-doer upgrade (a 4th alias later). The minimal go-bag still runs on **one** model. Robustness: native tool-calls first, a forgiving parser + bounded retries as fallback (defer the pi-style text protocol to P2), **the oracle gate as the real safety net** — never trust a raw tool-call.
2. **pg-backend boundary.** Two Go interfaces — `Ledger` (work-items, context/engrams, dispatch/cost) and `Dispatcher` (the model loop). Standalone = SQLite `Ledger` + local-loop `Dispatcher`. Substrate-backed (P4) supplies a pg `Ledger` adapter (shared multi-session store) and/or an MCP `Dispatcher` adapter (offload to the substrate's dispatch/council/catalog/cost), *independently selectable*, additive, never required. Deferred to P4 planning: shared-ledger-first vs dispatch-offload-first (lead: shared-ledger).
3. **Presiding-chain surface.** Terminal-native, **no web**. Inline compact chain status in the main loop + a read-only **`garrison watch` TUI** (pure-Go bubbletea; tails the ledger; renders the live presider→sub-agent tree with state / intent / last-action / cost). Ships in the **same phase as sub-agent spawning**, which is gated behind the oracle (P2) and critic (P3) existing first. P1 = structured log lines only.
4. **Tier-2 self-extension.** **P5, its own council item** (a self-built persistent tool is new standing dominion → `dominion_in_council` applied recursively). "Hang with consent" reuses the reflect-steward's **approve→queue→capacity-gated-drain**: build + test the door immediately and free, but **gate installation** — interactive = a hard inline pause; semi-autonomous = **queued, never auto-installed** (degrade-and-continue, or block-and-surface), with a watchman-guard pause on too many pending doors. Tighter the weaker the model.
5. **Plugin relationship.** **Yes** — `plugin-someday` is Garrison's luxury-mode client of the same MCP surface. Garrison runs as a standalone loop **and** exposes its ledger / watch / tools as an MCP server (one surface, two directions: client to the substrate, server to Claude Code, plus multi-Garrison coordination). **Locked implication: design the core MCP-exposable from P1**; ship the plugin client at P6.
6. **P1 dogfood target.** A **greenfield Go utility + tests** Garrison writes from scratch with a hard `go build && go test` oracle (the substrate-coder calc/FizzBuzz pattern) — cleanest proof of the loop, lowest risk. The same task is the pi + `qwen3.6-27b` baseline drive *before* Garrison writes a line. The "practices what it preaches" stretch dogfood (Garrison builds a small piece of its own tooling — a code detector in the study-linter spirit) follows once the basic loop is proven.

**Store decision (the latent contradiction, resolved).** P1 ships **FTS5 keyword (modernc, pure-Go) + chromem-go vectors (pure-Go, file-persisted) fused with RRF** — the gospel-engine lineage borrowed whole; embeddings via the local **LM Studio `text-embedding-nomic-embed-text-v1.5`** endpoint (`:1234`) already running. *Not* `sqlite-vec` (a C extension `modernc` cannot load — the two can't coexist in a one-binary, CGo-free go-bag).

## Recommendation

Pursue it, as specced: a Go, isolated, **standalone-on-SQLite** harness with the governance built in and the substrate as an optional MCP-attached power-up; extensions out-of-process (MCP / JSON-RPC / HTTP / WebSocket) and WASM, no gRPC; self-extension gated by build-the-door / hang-with-consent; built after the cut. Hold the one discipline that defines it: **every prerequisite is something Michael owns and runs — the binary, the SQLite file, the local model — never something rented that can be revoked.** SQLite-by-default is what makes the daily driver and the last position held the *same* tool, not two.
