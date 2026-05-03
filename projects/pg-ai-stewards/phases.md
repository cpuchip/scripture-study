# pg-ai-stewards — phases

A phased delivery plan. Each phase has a concrete deliverable that
*works end-to-end*, not "scaffolding for future work." Each phase
either lands real value or kills the project early.

The proposal is at [proposal.md](proposal.md). The probe that backs
Phase 1's feasibility is at [probe/](probe/).

## Phase 0 — Done (this session, 2026-05-02)

- Research scratch with verified sources: [scratch.md](scratch.md)
- Probe stack with passing bridge tests: [probe/](probe/), [probe/RESULTS.md](probe/RESULTS.md)
- Direction confirmed; spec written: [proposal.md](proposal.md)

## Phase 1 — Foundation: extension scaffold + bgworker + brain port

**Goal:** prove the architecture by replacing `scripts/brain/`'s
SQLite store with a Postgres+pgvector+pg_ai_stewards equivalent. By
the end of Phase 1, the existing brain CLI works against Postgres
and at least one LLM-using path goes through the bgworker.

### Deliverables

1. **Extension scaffold** — `scripts/pg-ai-stewards/` (or sibling)
   - `cargo pgrx new pg_ai_stewards` skeleton against PG18.
   - Builds and loads in a docker compose alongside pgvector + AGE.
   - One trivial SQL function (`stewards.version()`) to prove the
     extension actually loaded.

   **✅ Done 2026-05-02.** Lives at [extension/](extension/). Image
   `pg-ai-stewards-dev:pg18` runs on host port 55433 with all three
   extensions installed. `stewards.version()` returns `0.1.0`. See
   [extension/README.md](extension/README.md).
2. **bgworker with reqwest + tokio**
   - Listens on `LISTEN stewards_dispatch`.
   - On notify, reads a row from `stewards.work_queue`, calls a stub
     "echo" provider, writes the result back, `NOTIFY stewards_done`.
   - Handles SIGTERM cleanly. Restarts cleanly.
   - **Reads provider registry from env vars on startup**
     (`STEWARDS_PROVIDER_*` — see [proposal § Provider abstraction and
     secrets](proposal.md#provider-abstraction-and-secrets)). Even
     the echo stub goes through this so the registry is real from
     day one.

   **✅ Done 2026-05-02 (revised approach).** Lives in
   [extension/src/lib.rs](extension/src/lib.rs). The honest scope
   change: we **poll** every 500ms with `wait_latch` rather than
   `LISTEN`-driven wake-up. Reason: `LISTEN` from a bgworker requires
   going under pgrx's covers (SPI doesn't expose libpq's NOTIFY
   channel). Polling-with-completion-NOTIFY matches `pg_vectorize`'s
   pattern. End-to-end latency observed: avg **138 ms** for a small
   batch (well under tick), max bound by the 500 ms tick.
   We still `NOTIFY stewards_done '<id>'` on completion so external
   listeners can react in real time. tokio + reqwest deferred to
   step 6/7 when there's an actual HTTP call to make.

   Verified via inverse hypothesis (Agans Rule 9): with
   `shared_preload_libraries=pg_ai_stewards` removed, an enqueued
   row stays `pending` forever; restoring it drains the same row in
   under a tick. SIGTERM exits cleanly
   (`stewards: bgworker received SIGTERM, exiting` in the log) and
   the postmaster respawns the worker on container restart.
3. **Schema for brain replacement**
   - `stewards.brain_entries` (six categories, JSONB props,
     embedding column, HNSW index).
   - `stewards.messages` (basic conversation log so we have something
     to embed and query end-to-end).

   **✅ Done 2026-05-02.** Lives alongside step 2 in
   [extension/src/lib.rs](extension/src/lib.rs) as a second
   `extension_sql!` block (`requires = ["create_work_queue"]` for
   ordering). Implementation notes:
   - **Seven** categories in the CHECK constraint, not six —
     `inbox` is the unclassified default that brain's classifier and
     `web/server.go` both write. Read from
     `scripts/brain/internal/classifier/classifier.go` per the
     data-safety checklist (categories never get listed from memory).
   - Single `brain_entries` table + JSONB `props` instead of one
     table per category. Matches chromem-go's storage shape and
     keeps the migrator (step 4) simple. Brain's category-specific
     columns (`name`, `follow_ups`, `status`, `due_date`, `mood`,
     `gratitude`, ...) all fold into `props`.
   - Aux tables landed too: `brain_entry_tags`, `brain_subtasks`,
     `brain_versions`, `sessions`, `messages`. Step 4's migrator
     reads from SQLite tables of the same shape, so doing them now
     keeps step 4 to pure read/write/verify.
   - Embedding column is `vector(768)` to match gospel-engine-v2.
     HNSW with `vector_cosine_ops`. NULL embeddings are skipped
     by the index naturally; queries also filter `IS NOT NULL`.
   - `body_tsv tsvector GENERATED ALWAYS AS (...) STORED` plus a
     GIN index gives free FTS — no triggers, no inconsistency
     window. Wrapped by `stewards.brain_search_text()`.
   - Two triggers on `brain_entries`: `BEFORE UPDATE` snapshots
     OLD into `brain_versions` and bumps `updated_at`;
     `AFTER INSERT OR UPDATE OF title, body` enqueues
     `(kind='embed', provider='ollama')` in `stewards.work_queue`.
     The bgworker's echo stub (still in place from step 2) marks
     them `done` without writing the embedding — step 6 swaps the
     stub for a real Ollama HTTP call.
   - Helpers: `brain_upsert(category, title, body, props, tags,
     id?, source?)`, `brain_search_text(query, category?, limit)`,
     `brain_search_vec(embedding, category?, limit)`.
   - `requires = 'vector'` added to `.control` so `CREATE EXTENSION
     pg_ai_stewards` pulls in pgvector transitively if missing.
   - **Hybrid FTS+vector search deferred.** Phases.md mentions a
     combined `brain_search`. Until step 6 fills the embedding
     column there's nothing to combine, and rank-fusion strategy
     benefits from real query traffic to tune. Surfaced as an
     explicit deferral instead of shipping a half-done version.

   Verified end-to-end via init SQL + manual probes: brain entry
   inserted, embed-enqueue trigger fired on both INSERT and UPDATE,
   FTS finds revised text, version snapshot captured the OLD title,
   sessions/messages cascade works.
4. **Migrator** — one-shot Go binary that reads existing SQLite +
   chromem, writes to Postgres.
5. **Brain CLI driver** — new backend in `scripts/brain/` that talks
   to Postgres via the existing brain API surface. Old SQLite driver
   stays as read-only fallback.
6. **At least one real provider call through the bgworker** — the
   "embedding generation" path. Insert a brain entry → bgworker
   computes embedding via Ollama → writes pgvector column → search
   works.

   **✅ Done 2026-05-02 (with LM Studio, not Ollama).** Michael
   doesn't run Ollama locally; LM Studio serves the same
   nomic-embed-text-v1.5 at 768 dims via the same OpenAI-compatible
   `/v1/embeddings` endpoint. Trigger updated to enqueue with
   `provider='lm_studio'`. Implementation notes in
   [extension/src/lib.rs](extension/src/lib.rs):
   - **`reqwest = { default-features = false, features = ["blocking",
     "json", "rustls-tls"] }`** — blocking client (worker is already
     a sync per-tick loop, no tokio runtime needed) with rustls so
     we don't need libssl-dev in the runtime image.
   - **Three-phase dispatch** in `process_one_pending`: Tx A claims
     the row and commits, Tx B holds nothing while HTTP runs (LM
     Studio's first cold load takes 2–3s and we don't want to hold
     a row lock through that), Tx C writes the result and NOTIFYs.
   - **`dispatch(kind, provider, payload)`** matches on kind. Echo
     keeps working unchanged. New `embed` arm calls
     `<base_url>/embeddings` with `{model, input}`, expects the
     standard `{data: [{embedding: [f64...]}]}` shape, validates
     `len == dimensions`, formats the floats as pgvector's text
     literal (`[v1,v2,...]`), and returns `WorkOutcome::Embedded`.
   - **120s HTTP timeout** for cold-load tolerance.
   - **Cast in the UPDATE**: `SET embedding = $2::vector(768)`.
     Dimension mismatch raises a Postgres error rather than silently
     storing wrong shape.
   - **Failure path** stamps `embedding_error` on the brain row
     too, so app queries see why a row never embedded — not just
     a NULL vector.
   - **Trigger fix bundled in:** `touch_brain_entry` now only
     snapshots into `brain_versions` when title/category/body/props
     actually change, so embedding writes don't generate junk
     version rows.

   Verified end-to-end: 5 brain entries embedded via LM Studio
   (avg **610ms** warm, ~3s first cold call), `vector_dims = 768`,
   `brain_search_vec` ranks correctly ("Charity is the pure love
   of Christ" → 0.195 distance from "pure love of Christ moroni",
   "Faith hope and charity" → 0.363, self → 0.0). Inverse hypothesis
   confirmed (Agans Rule 9): rewriting the trigger to point at a
   non-existent provider produces `work_queue.status='error'` with
   message `unknown provider: no_such_provider` and stamps
   `embedding_error` on the brain row. Restoring the trigger and
   re-UPDATEing succeeds and clears the error.
7. **Second real provider call: chat via OpenCode Go.** Send a
   `stewards.work_items` row with `kind = 'chat'` and
   `provider = 'opencode_go'`, model `kimi-k2.6`. Bgworker hits
   `https://opencode.ai/zen/go/v1/chat/completions` with the bearer
   key from env, writes the response back. Same code path as Ollama
   chat — just a different provider entry. This is the proof that
   the OpenAI-compat lingua franca decision actually pays off.

### Done when

- `brain search "charity"` returns the same results from the new
  Postgres backend that it does from the old SQLite backend.
- A brain entry inserted now has its embedding generated by the
  bgworker (verified by checking the row's `embedding` is non-null
  ~1 second after insert).
- The bgworker survives a `docker compose restart` without losing
  in-flight work (it should re-read pending rows from `work_queue`
  on startup).

### Kill criteria

- pgrx bgworker + tokio runtime turns out to be fundamentally broken
  on Windows-hosted Docker (i.e. PG worker can't keep a tokio
  runtime alive). Probability: very low; pg_vectorize ships this on
  Linux containers and we'll run the same.
- Migration from SQLite + chromem loses data we can't reconstruct.
  Probability: low if we keep SQLite as read-only fallback.

## Phase 1.5 — Harness sketch (detour) ✅ done 2026-05-03

**Why a detour:** after step 6 landed real LM Studio embeddings,
Michael flagged the obvious gap: copilot-sdk had been carrying the
agentic plumbing (prompt assembly, tool registry, skill dispatch,
MCP server lifecycle) silently. Step 7 ("OpenCode Go chat through
the bgworker") would have built another single-shot provider call
without answering: when the agent loop arrives, where do the
`messages[]` come from? what's in the `tools[]`? how do skills
show up? Better to sketch the harness first, look at the JSON we'd
send, and let the schema critique itself — *before* committing to a
chat-shaped data path that might want different bones.

**What it builds:** a minimum read-only harness in pure SQL, no HTTP.
Deliverable is `stewards.dry_run_chat(agent_family, model, session,
input)` returning the exact JSON body that would be POSTed to
`/v1/chat/completions`. We *look* at the body and judge the shape
before step 7 makes it real.

**Inputs that shaped the design** (after reading [opencode source](https://github.com/anomalyco/opencode/) and docs):
- Skills are NOT injected into the system prompt by default. They're
  advertised via an `<available_skills>` XML block inside the `skill`
  tool's description; the agent calls `skill({name})` to load a body.
  Token-efficient. We adopt this.
- Agent IS its config. `(name, mode, prompt, model_pin?, temperature,
  top_p, steps, permissions)`. Subagent invocation is just another
  tool call. Built-ins: `build`/`plan` (primary), `general`/`explore`
  (subagent), three hidden housekeeping (`compaction`/`title`/
  `summary`).
- Tool name = `<prefix>_<name>` is universal. MCP server prefix or
  filename prefix. Permissions glob on the prefix (`brain_*: allow`).
- Permissions are 3-state (`allow`/`ask`/`deny`), glob-matched, last
  matching rule wins. Per-agent overrides global.

**Variant-by-glob (Michael's contribution):** Different models reason
about the same instructions differently. Kimi over-explains; GPT-5
ignores temperature; Qwen wants its own defaults. We add a
`model_match` column to `agents`, `skills`, and `instructions` —
glob like `kimi-*`, with `'*'` as the catch-all default. Resolver
picks the longest matching pattern. Tools deliberately *don't* get
variants (a tool's description is structural, not stylistic).

### Schema (in [extension/src/lib.rs](extension/src/lib.rs))

- `stewards.agents` — PK `(family, model_match)`, persona prompt,
  temperature/top_p/steps. NULL eliminated by using `'*'` sentinel
  so the PK works and `ON CONFLICT` is honest.
- `stewards.skills` — same shape. Family must match
  `^[a-z0-9]+(-[a-z0-9]+)*$` (opencode rule). Description
  1-1024 chars (opencode rule).
- `stewards.instructions` — `(family, model_match, scope)` UNIQUE,
  `scope` is `'global' | 'agent:<family>' | 'session:<id>'`,
  `ord` for sort order.
- `stewards.tool_defs` — `name` PK with `^[a-z][a-z0-9_]*$` check,
  `args_schema` jsonb (JSON Schema), `execute_target` jsonb
  describing dispatch (`{kind:'sql_fn'|'http'|'subagent', ...}`).
  No model variants in v1.
- `stewards.agent_tool_perms` / `stewards.agent_skill_perms` —
  glob patterns + 3-state action.
- `stewards.tool_calls` — empty in v1, exists so step 7+ can write
  without a migration.

### Functions

- `glob_match(pattern, value)` — escape `\`, `%`, `_` then turn `*`
  into `%`, run as `LIKE`. Doesn't support `?` (single-char) — model
  names don't need it.
- `resolve_agent(family, model)` / `resolve_skill(family, model)` —
  longest matching `model_match` wins; `'*'` is length 1 so any
  specific glob beats it.
- `tool_permission(agent, tool)` / `skill_permission(agent, skill)` —
  longest matching pattern wins; default `'allow'` if no rule.
- `compose_system_prompt(family, model, session)` — agent persona +
  matching instructions (deduped per family by best variant) +
  `<available_skills>` XML if `skill` tool isn't denied.
- `compose_messages(family, model, session, user_input?)` —
  `[system, ...history, ?user]` as jsonb.
- `compose_tools(family)` — OpenAI-shape `tools[]` filtered by
  permissions (only `deny` excluded; `ask` included for the loop
  to handle).
- `dry_run_chat(...)` — the verification target. Returns full POST
  body plus `_meta` showing which variant resolved.

### Seed data

- One agent family `stewards-explore` with two variants: default
  (`'*'`) and `'kimi-*'` (with extra "be terse" clause).
- Two instructions families: `honesty` (global) and `search-budget`
  (agent-scoped). Both `'*'` (model-agnostic for v1).
- Two skills modeled on real `.github/skills/` entries:
  `source-verification` and `scripture-linking`.
- Two tool defs: `brain_search_text` (real, dispatches to existing
  `stewards.brain_search_text` SQL fn) and `skill` (special loader).
- Permissions for `stewards-explore`: `*: deny`, `brain_*: allow`,
  `skill: allow`. Explicitly proves the deny-by-default-then-whitelist
  pattern.

### Verification (the actual point)

```
dry_run_chat('stewards-explore', 'kimi-k2.6', 'dry-run-1', 'and what about hope?')
dry_run_chat('stewards-explore', 'gpt-5.1',   'dry-run-1', 'and what about hope?')
```

- Kimi: `_meta.agent_variant_match = 'kimi-*'`, system prompt 1049 chars.
- GPT-5: `_meta.agent_variant_match = '*'`,    system prompt 963 chars.
- 86-char delta = the "be terse" paragraph, present only on Kimi.
- Same instructions block, same `<available_skills>`, same tools[],
  same temperature. Persona is the only delta.
- `tools[]` has 2 entries with canonical OpenAI shape
  (`{type, function: {name, description, parameters}}`); JSON Schema
  intact (enum, min/max).
- `messages[]` = `[system, user, assistant, user]` (system + 2-turn
  history + new user input).
- Inverse hypothesis (Agans Rule 9): `dry_run_chat('does-not-exist',
  ...)` raises `no agent variant resolved: family=does-not-exist
  model=gpt-5.1` cleanly.

### What this unlocks for step 7

Step 7 now becomes: "call `dry_run_chat()` to get the body, POST it
to `<provider>/chat/completions`, parse the response, append the
assistant message to `stewards.messages`, write `tool_calls` rows
for any `tool_calls[]` in the response." The agent loop is
then one wrapper around step 7 plus the dispatcher for
`tool_defs.execute_target`. The composition concerns are settled.

### What it deliberately doesn't build

- No agent loop. Single-turn dry run only.
- No real tool execution. `execute_target` is data; nothing reads it yet.
- No real MCP transport. "MCP equivalent" in v1 is the
  `execute_target: {kind:'sql_fn', name:...}` shape. Real MCP client
  comes later when we want to consume gospel-engine's MCP from
  inside stewards.
- No `steps` enforcement. Column exists; the loop that respects it doesn't.
- No session-scoped instructions. Schema supports it; nothing writes them yet.

## Phase 2 — Studies + AGE: citations as edges

**Goal:** make studies first-class rows that link to canonical
sources via AGE edges, with cross-DB resolution to gospel-engine-v2.

### Deliverables

1. **`stewards.studies` table** + a one-shot importer for existing
   markdown studies in `study/`. Each gets an embedding.
2. **AGE graph wiring** — for each study, parse the markdown's
   scripture/talk links, create `:Study`-`[:CITES]`->`:Scripture`
   and `:Study`-`[:CITES_AS_CORE]`->`:Scripture` edges with the
   `lds://...` URI as a property.
3. **Resolver** — a small Go service (or extension function calling
   out via the bgworker) that takes an `lds://...` URI and returns
   the resolved scripture/talk text + metadata via gospel-engine-v2's
   HTTP API. Caches in `stewards.resolved_refs` with TTL.
4. **Bridge in production** — when a new study is written, run the
   probe's bridge pattern: pgvector finds the N nearest existing
   studies/scriptures; high-similarity pairs become
   `:SIMILAR_TO {score, method}` edges.
5. **CLI: `stewards study show <slug>`** — prints the study, its
   citations, its similar studies, and pulls in the cited verse
   text via the resolver.

### Done when

- Importing all of `study/` produces studies + edges + resolver
  cache populated.
- `stewards study show give-away-all-my-sins` returns the study,
  its scripture citations *with verse text resolved from gospel*,
  and three similar studies.
- A new study saved into `study/` triggers a sync that updates the
  graph (edges added/removed to match the markdown).

### Kill criteria

- Resolver round-trip cost is so high it makes interactive use
  painful (>500ms p50). Mitigation: aggressive caching;
  resolver-as-bgworker.
- AGE on PG18 has more rough edges than expected at the volume of
  edges we'd create. Probability: low based on probe; revisit if
  >100k edges starts misbehaving.

## Phase 3 — Pipelines + MCP: agents that work without an IDE

**Goal:** long-running agent work runs without an open VS Code
window. Tool sidecars execute work; the bgworker dispatches; results
flow back to a thin web UI for review.

### Deliverables

1. **Pipeline tables**
   - `stewards.pipelines` — definition of a flow.
   - `stewards.work_items` — instances flowing through statuses.
   - Status transitions audited; SQL functions enforce valid
     transitions; bgworker reacts to status changes via `NOTIFY`.
2. **Tool sidecar protocol** — small JSON-RPC over a unix socket or
   HTTP. Sidecars register their capabilities; bgworker dispatches.
   First sidecars: filesystem (Docker container with mounted repo),
   git/gh, shell.
3. **MCP server** — exposes a small set of tools modeled on the
   Azure-Samples repo:
   - `stewards_search` (combined keyword + vector + graph)
   - `stewards_brain` (read/write brain entries)
   - `stewards_work_item` (queue / promote / cancel)
   - `gospel_passthrough` (proxies to gospel-engine-v2)
4. **Web UI surface** — minimal. Lives in `becoming/` (existing
   cloud hub). Shows worklist, lets Michael promote/demote/cancel
   items, see in-flight model calls. Not a code editor.
5. **Becoming integration** — the Discord relay reads/writes brain
   entries and work items via the same DB. ibeco.me web reads from
   the same DB.

### Done when

- A long-running task ("review all studies and identify the three
  that contradict each other") can be queued from the web UI,
  runs without an IDE window, posts a result back, and Michael
  reviews it from any device.
- The web UI can interrupt an in-flight model call.

### Kill criteria

- The MCP routing layer adds enough latency that interactive use
  feels worse than current VS Code Copilot. Mitigation: keep
  direct DB clients (CLI, brain.exe) as the fast path; MCP for
  cross-process orchestration only.

## Phase 4 — Optional: GraphRAG over the canon

**Goal:** higher-quality holistic queries over scripture and
conference talks via Microsoft GraphRAG community summaries.

This is *optional*. We do it only if Phases 1–3 surface the need.

### Deliverables (if pursued)

1. Run [`microsoft/graphrag`](https://github.com/microsoft/graphrag)
   indexing pass over the scripture + talk corpus.
2. Write the resulting community summaries into AGE in the `gospel`
   database (with gospel-engine-v2's blessing — this would be a
   schema change in their territory).
3. Add a `gospel_global_search` MCP tool that calls GraphRAG global
   search.

### Why this is Phase 4

- It's expensive to run.
- It only matters for "themes across the whole corpus" questions
  that current vector search handles poorly.
- We won't know if we need it until we've used Phases 1–3 for a few
  months.

## Phase 5+ — Maybe-someday

Listed but not committed:

- **`postgres_fdw` from stewards into gospel** for SQL-level joins
  (currently we go via HTTP API).
- **Multi-tenant RLS** if ibeco.me ever hosts other people.
- **`pg_diskann`** if HNSW becomes the bottleneck.
- **Replace VS Code chat client with a Postgres-native one.** Big
  scope; only if the existing surface stops being good enough.
- **Public release of `pg_ai_stewards`** if it turns out to be useful
  beyond this project.

## Sequencing

Phase 1 is a hard prerequisite for everything else.
Phase 2 needs Phase 1 (it uses the bgworker for resolver / embedding).
Phase 3 needs Phases 1 + 2 (pipelines act on studies and brain entries).
Phase 4 is independent of Phase 3 and could happen in parallel — but
shouldn't, until 3 is real.

## Cadence

No time estimates. Each phase ends when its "done when" criteria are
met. Move on after a brief sabbath reflection (`.spec/journal/`,
`.mind/active.md` updates) to capture what we learned.
