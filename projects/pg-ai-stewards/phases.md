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

   **⏸ Deferred 2026-05-03.** Phase 1 was originally framed as
   "replace brain.exe storage." That framing is now stale: the
   substrate (composition + agent loop, steps 1.5/1.6) turned out
   to be the load-bearing deliverable, and we proved it without
   migrating brain. brain.exe on SQLite continues to work; this
   becomes a "do it when SQLite hurts" item, not a Phase 2 blocker.
   Tracked as a future Phase 1.7 if we re-prioritize.
5. **Brain CLI driver** — new backend in `scripts/brain/` that talks
   to Postgres via the existing brain API surface. Old SQLite driver
   stays as read-only fallback.

   **⏸ Deferred 2026-05-03.** Same reason as #4 — paired work;
   together they form Phase 1.7 if/when we revisit.
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
7. **Second real provider call: chat via OpenCode Go.** ✅ done 2026-05-03.
   Built on top of the Phase 1.5 harness:
   - `stewards.chat_enqueue(agent_family, model, session_id, user_input,
     provider)` composes the body via `dry_run_chat`, persists the
     user turn, and enqueues `kind='chat'` with the body in payload.
   - Bgworker `dispatch()` `chat` arm POSTs to
     `<base>/chat/completions`, parses standard OpenAI shape
     (`choices[0].message`, `usage`, `model`), and phase 3 inserts
     the assistant message into `stewards.messages` (with `tool_calls`
     verbatim if present, `finish_reason`, `tokens_in/out`).
   - Verified: 4.4s round-trip kimi-k2.6 via OpenCode Go gateway. Kimi
     answered "what is your job here?" by accurately restating the
     persona we composed in `agents.prompt` — proving the Phase 1.5
     harness shape arrives intact at the model.
   - Provider echo persisted: we asked for `kimi-k2.6`, OpenCode Go's
     gateway returned `moonshotai/kimi-k2.6-20260420`. We store what
     the provider actually used, not what we asked for.
   - Inverse hypothesis (Agans Rule 9): bad provider →
     `work_queue.status='error'` with `unknown provider:
     does_not_exist`, no row leaks, no broken state.
   - Stewardship action: an early draft included a `chat_round_trip()`
     SQL fn that enqueued + polled in one tx. Caught immediately on
     first run when the open tx hid its own enqueued row from the
     bgworker (MVCC) AND blocked every other writer on the session
     row lock. Removed; left an inline comment so future-me doesn't
     reach for it. Real polling needs `LISTEN stewards_done` from
     outside Postgres, or a separate statement.

   What's still NOT here (Phase 1.6 / step 8):
   - Tool execution. Assistant's `tool_calls` jsonb is persisted but
     nothing reads it yet. (Confirmed kimi DOES invoke tools when
     the question warrants it — "name two virtues from Moroni 7"
     produced a valid `brain_search_text` call with sensible args.)
   - The agent loop. One turn only — no `while assistant.tool_calls
     and steps < agent.steps`.
   - Tool result messages (`role='tool'`, `tool_call_id`). Schema
     supports them; nothing writes them yet.

   Cost-correctness adds (same session, post-first-roundtrip):
   - `messages.reasoning_tokens int` column, populated from
     `usage.completion_tokens_details.reasoning_tokens`. Kimi-class
     models bill reasoning tokens SEPARATELY from `completion_tokens`;
     under-counting them halves the apparent cost. Real test showed
     `tokens_out=111, reasoning_tokens=93, billable_out=204`.
   - `chat_enqueue` injects `user = <session_id>` into the outgoing
     body (OpenAI-spec field). Providers that surface per-session
     billing (OpenCode Go's usage dashboard) tag the request with
     our session id, giving free cost-per-session attribution.
   - `work_queue.result.billable_output` = `tokens_out +
     reasoning_tokens`, ready for a future cost helper to multiply
     by per-model rates.

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

## Phase 1.6 — Agent loop ✅ done 2026-05-03

**Goal:** close the loop. Phase 1.5 settled composition; step 7
proved one round-trip. Phase 1.6 makes the model's `tool_calls`
actually execute, feed back as `role='tool'` messages, and
continue until `finish_reason ∈ {stop, length, content_filter}`
or the step budget runs out.

### Architectural choice: work-item-per-iteration

Two reasonable architectures: (A) keep one bgworker tick busy until
the loop terminates, (B) every iteration is its own work_queue row.
We chose B. Reasons:
- Every step is durable (work_queue row) and observable (NOTIFY on
  each transition).
- Cancellation is one UPDATE on a pending continuation.
- A 30-second tool call can't starve other sessions — bgworker
  picks up siblings between iterations.
- Steps remain auditable after the fact; you can replay or branch
  from any point.

Cost: more SQL writes per loop. Acceptable.

### Deliverables

1. **Schema additions to `stewards.messages`**
   - `parent_work_id bigint` — back-pointer to the work_queue row
     that produced this message. Used by tool_dispatch to find the
     assistant message it's responding to.
   - `reasoning_content text` — captured from gateway response. Some
     gateways (Moonshot direct) require this back on the next request
     for thinking-enabled models or return 400.
   - `reasoning_details jsonb` — captured from gateway response.
     OpenCode Go uses this field name; we emit both because different
     gateways read different names.

2. **`kind='tool_dispatch'` work item + dispatcher**
   - Reads `tool_calls` from the parent assistant message.
   - For each call, resolves the tool against `stewards.tool_defs`,
     checks `agent_tool_perms`, executes via `sql_fn` or `http`.
   - Inserts `role='tool'` messages with proper `tool_call_id`
     echoes (one per tool call).
   - Enqueues a continuation `kind='chat'` work item.

3. **Phase-3 of `chat` arm**
   - When response has `tool_calls` AND iteration < `agent.steps`,
     enqueues `tool_dispatch` instead of stopping.
   - `parent_work_id` on the assistant message links the chain.

4. **`compose_messages` upgrades**
   - Emits full message shape: `tool_calls` on assistant, `tool_call_id`
     on tool, `reasoning_content`/`reasoning_details` echoed back.
   - Builds `[system, ...history, ?user]` with monotonically growing
     prefix — enables prompt caching on identical `system + tools`.

5. **Two seeded sql_fn tools**
   - `brain_search_text_tool(jsonb) -> jsonb` — wrapper around the
     existing FTS function.
   - `load_skill_tool(jsonb) -> jsonb` — pulls a skill body from
     `stewards.skills` for the `skill` builtin.

6. **Bgworker resilience**
   - **Stale-claim reaper at startup** — any `in_progress` row at
     bgworker startup is by definition orphaned (we run one worker).
     Marked errored with a clear message so callers can decide what
     to do. Window is zero.
   - **`pg_proc` pre-flight on sql_fn dispatch** — checks the target
     function exists before constructing the SELECT. Returns a normal
     Rust `Err` if not, so the missing-function ereport is never
     triggered. Workaround for a pgrx 0.18 quirk: PgTryBuilder does
     NOT empirically catch ereports through `BackgroundWorker::
     transaction` + `Spi::connect`, so the cheapest defense is to
     not trigger them.

### Verification

- **Success path** (`verify-loop.sql` `loop-3` session): "In one
  sentence, name two virtues from Moroni 7." → kimi calls
  `brain_search_text` (empty result) AND `skill` (loads
  source-verification body) → reads both replies → answers
  "I found no brain entries on this topic, but Moroni 7:45 names
  virtues such as patience and kindness." `finish_reason: stop`.
  18s end-to-end, ~$0.0005.
- **Inverse hypothesis** (Agans Rule 9, `loop-err2` session):
  pre-registered `always_fails` tool pointing at a non-existent
  function. Request: "Please call the always_fails tool and report
  what happens." → tool dispatch returns
  `{"error":"sql_fn target stewards.nonexistent_function(jsonb) does not exist"}`
  as a `role='tool'` message → kimi reads it and replies "The tool
  failed with a SQL error: ...". `finish_reason: stop`. **Bgworker
  did not crash.** Verified in postmaster logs.
- **Reasoning replay** verified: both turns of `loop-3` carry
  reasoning_content (266 chars turn 1, 2982 chars turn 2). Without
  echoing them, Moonshot returns 400.

### What this still doesn't build (deliberately)

- **`tool_dispatch`-itself error recovery.** If the dispatcher row
  ERRORS (vs. a tool returning an error string), no `role='tool'`
  reply gets written and the model never sees what happened — the
  parent chat's continuation expectation is unfulfilled. Acceptable
  now (only happens on truly broken tool config that a developer
  fixes), but tracked for Phase 1.6.1 below.
- **`steps` budget enforcement.** Column exists on `agents`;
  iteration counter exists on messages via `parent_work_id` chain.
  Wire the actual cutoff next time we need it (current default
  agent has `steps=10`, we haven't hit it).
- **Per-call billing aggregation.** Each work_queue row records
  cost; nothing rolls them up per-session yet. One SELECT away when
  needed.

## Phase 1.6.1 — tool_dispatch error recovery ✅ done 2026-05-03

**Why this sub-phase:** the spec gap surfaced during Phase 1.6
verification matters once anything other than a developer drives
the loop. Tools fail for many reasons — network blips, rate limits,
sidecar restarts, provider quota exhaustion, schema drift. The loop
must degrade gracefully, not stall.

### Failure modes and resolution

1. **Tool returns an error string** — handled in 1.6 (per-tool
   error path wraps as `{"error":...}` `role='tool'` content;
   model recovers).
2. **Tool function ereports** — handled in 1.6 via `pg_proc`
   pre-flight + PgTryBuilder belt-and-suspenders.
3. **Tool dispatcher itself errors** (the previously open gap) —
   prep tx fails, parent assistant message missing, payload
   malformed. Now: dispatcher's `Err(msg)` arm calls
   `synthesize_tool_failure()` which writes `role='tool'` replies
   for every tool_call_id in the parent assistant message AND
   enqueues continuation chat. Loop continues.
4. **Bgworker crash mid-dispatch** — startup reaper now (a) marks
   stale `in_progress` rows errored as before, (b) for every
   reaped `tool_dispatch` row, calls `synthesize_tool_failure()`
   to write replies + enqueue continuation. The reaper is the
   "always-on safety net" — even a hard kernel-level kill of the
   bgworker recovers on the next start.
5. **HTTP tool timeout** — already handled correctly. `exec_http_tool`
   returns `Err` on reqwest timeout, which routes through the
   per-tool error wrap (mode 1). Tool reply content is
   `{"error":"POST <url>: ... timed out"}`.
6. **Tool returns malformed JSON** — already handled. Args decode
   failure produces `{"_decode_error":...,"_raw":...}` sentinel
   in the tool args; the per-tool error path handles bad return
   values from the tool itself.

### Deliverables

1. **`synthesize_tool_failure(parent_work_id, agent_family, model,
   session_id, provider, error)` SQL function.** Looks up the
   parent assistant message's `tool_calls`, inserts a synthetic
   `role='tool'` message for each `tool_call_id` that doesn't
   already have a reply (idempotent), and calls
   `chat_post_internal()` to enqueue the continuation chat.
   Synthetic content is JSON: `{"error":"<msg>","_synthetic":true,
   "_reason":"dispatcher failure; no tool execution occurred"}`.
   Returns the continuation work_id.
2. **Dispatcher Err arm wired** — when the bgworker's
   `process_one_pending` matches `Err(msg)` and `kind == "tool_dispatch"`,
   it calls `synthesize_tool_failure` via SPI before stamping
   the work_queue row errored. The error result also records
   `continuation_after_failure` so the audit trail shows what
   the recovery path enqueued.
3. **Reaper enhanced** — startup reaper now reads `(id, kind,
   provider, payload)` for every stale `in_progress` row. For
   `tool_dispatch` kind, it calls `synthesize_tool_failure`
   before marking errored. Logs `reaper synthesized tool failure
   for tool_dispatch id=N (parent=M)` so operators can see the
   recovery happen.
4. **`stewards.session_status` view.** One row per session with
   `last_finish_reason`, `last_loop_stop_reason`, `pending_work`,
   `errored_work`, `total_tokens_in`, `total_billable_out`,
   `last_assistant_at`. Single SELECT answers "did this loop
   finish or stall?".
5. **Step budget enforcement** — already implemented in 1.6
   (chat handler's phase 3 checks `iteration_count` vs
   `agent.steps`, sets `loop_stop_reason: "steps_exhausted"`
   in `work_queue.result`). Verified during 1.6.1 review.

### Verification

`verify-1-6-1.sql` — pure-SQL unit tests of the helper:
- Synthetic replies inserted with correct `tool_call_id` echo, JSON
  content, and `_synthetic: true` marker.
- Continuation chat enqueued (kind=chat, status=pending).
- Idempotent: second call writes zero new replies, dedup query
  catches it via `WHERE tool_call_id = v_tc_id`.
- `session_status` returns useful state for prior live sessions
  (`loop-3`: 7300 tokens in, 7414 billable out across 5 iterations;
  `loop-err2`: 1076/156 because the inverse failed fast).

`verify-1-6-1-reaper.ps1` + `verify-1-6-1-reaper-setup.sql` +
`verify-1-6-1-reaper-check.sql` — end-to-end integration test
of mode 4:
- Insert assistant message with `tool_calls` + reasoning_content.
- Insert orphaned `tool_dispatch` row directly in `in_progress`
  (simulates a worker that claimed-then-crashed).
- `docker compose restart pg`. Reaper runs at startup.
- Verify: orphan row marked errored, synthetic tool reply
  written, continuation chat enqueued, kimi reads the failure,
  retries with a real `brain_search_text` call (gets real result
  back), and finishes with `finish_reason='stop'`.

**End-to-end recovered loop verified.** Bgworker crash → reaper →
synthetic reply → model retry-with-different-args → success →
clean stop. Zero stalled rows.

### What this deliberately doesn't build

- **Per-tool retry policy** (`tool_defs.retry jsonb`). YAGNI for
  now. The agent loop already handles retries naturally — the
  model sees the error reply and decides whether to retry with
  different args, retry as-is, or give up. Adding a tool-level
  retry layer would be speculative without evidence that the
  model loops on broken tools. If we observe that pattern, add
  an attempt counter on the parent chat then.
- **Synthetic stop message at step budget exhaustion.** When
  `iteration_count == agent.steps`, the loop stops with the
  assistant's last `finish_reason` (which is `tool_calls`) and
  `loop_stop_reason='steps_exhausted'` in the work_queue result.
  Writing an extra synthetic assistant message saying "budget
  exhausted" would fabricate words the model never said. The
  truth is in `session_status.last_loop_stop_reason`; that's
  enough.
- **24-hour synthetic load test.** Will run before any production
  rollout, not before Phase 2.

## Phase 2 — Studies + AGE: citations as edges

**Goal:** make studies first-class rows that link to canonical
sources via AGE edges, with cross-DB resolution to gospel-engine-v2.

Broken into sub-phases the same way Phase 1 was. Phase 2.1 ships the
substrate (rows + graph + importer); Phase 2.2 adds the resolver;
Phase 2.3 adds similarity bridging; Phase 2.4 adds the CLI.

## Phase 2.1 — studies table + AGE citations ✅ done 2026-05-04

**What shipped:**

1. **`stewards.studies` table** — id, slug (UNIQUE), title, file_path,
   body, frontmatter (jsonb), embedding (vector(768)),
   embedded_at/model/error, body_tsv (FTS), created_at/updated_at.
   Indexes on slug, created_at DESC, body_tsv (gin), embedding (hnsw),
   frontmatter (gin).
2. **`stewards.study_versions`** — snapshot history identical in shape
   to brain_versions. Trigger snapshots on title/body/frontmatter
   change, NOT on embedding-only writes.
3. **Embed-enqueue trigger** — same pattern as brain_entries; reuses
   the existing `embed` work_kind (which UPDATEs
   `stewards.<target_table>` by id, so adding a new embeddable table
   is just "set up the same column shape").
4. **`stewards.ensure_studies_graph()`** — idempotent AGE bootstrap.
   `LOAD 'age'`, sets search_path, creates the `stewards_graph` if
   missing. Called from `00-extensions.sql` at first boot AND
   defensively from `import_study()` so a fresh session never sees
   "graph does not exist".
5. **`stewards.parse_gospel_links(body)`** — extracts every
   `[text](.../gospel-library/eng/...)` link from a markdown body.
   Returns (uri, anchor_text, kind ∈ {scripture, talk, manual,
   other}). Strips `../` prefixes; preserves `#verse` anchors;
   handles workspace-relative and workspace-absolute paths.
6. **`stewards.import_study(slug, file_path, title, body, frontmatter)`** —
   upserts the row (ON CONFLICT on slug, keeps id stable across
   re-imports), MERGEs the Study vertex, deletes existing CITES
   edges, then re-creates them from the current body. Sync
   semantics — re-importing always reflects the present markdown.
7. **`stewards.study_citations(slug)`** — read-side helper that
   round-trips the graph back to relational rows. Returns
   (study_slug, cited_uri, cited_kind, anchor_text, citation_count)
   ordered by citation_count DESC.
8. **PowerShell importer** (`import-studies.ps1`) — bulk-loads all
   markdown under `study/`. Per-file SQL written to a temp dir and
   `psql -f`'d via `docker cp` to avoid heredoc/pipe encoding issues
   with large bodies. Reads with `[System.IO.File]::ReadAllText(...,
   UTF8)` to keep em-dashes intact through PS5's default Windows-1252
   codepage.
9. **Verification** — `verify-2-1.sql` runs seven inverse-hypothesis
   tests: corpus loaded, parser shapes, apostrophe/em-dash/paren
   survival, re-import idempotency, edge-removal-on-link-removal,
   cross-study cypher query, cleanup. End-to-end: 69 studies, 432
   unique scripture vertices, 1256 CITES edges, all 69 embeddings
   populated by the bgworker.

**The actual bug we found and closed (worth recording):** AGE Cypher
does NOT recognize PG's `''` as an escape for a single quote inside
string literals. First implementation used `replace(p_title, '''',
'''''')` and `format()`-built Cypher; this silently produced syntax
errors on every study with an apostrophe or em-dash in the title or a
link's anchor text (13 of 69). Inverse hypothesis confirmed: raw
`SELECT * FROM cypher('g', $$ MERGE (x:T {l: 'don''t'}) ... $$)`
errors with "syntax error at or near `'t`". Fix: use `cypher()`'s
3-argument form to bind values via `$param` placeholders, with the
agtype built from `jsonb_build_object(...)::text::ag_catalog.agtype`.
This is the *only* safe way to inject user data into Cypher under
pg_age — record everywhere AGE writes happen.

**URI scheme.** Workspace-relative paths under `gospel-library/`
serve as canonical IDs:
- `eng/scriptures/bofm/mosiah/18.md` (chapter)
- `eng/scriptures/bofm/moro/7.md#47` (verse)
- `eng/general-conference/2024/04/<slug>.md` (talk)

This avoids inventing an `lds://` scheme before knowing what the
resolver needs; gospel-engine-v2's `/api/get?ref=...` already accepts
these paths.

### Deferred to later sub-phases (deliberately, not gaps)

- **Phase 2.2 — resolver.** Currently CITES edges only carry the URI;
  no scripture text is materialized inside the stewards DB. The
  resolver will hit gospel-engine-v2's `/api/get?ref=...` over HTTP
  (via the existing http tool dispatch path) and cache results in
  `stewards.resolved_refs` with TTL. Done when `study_citations()`
  can optionally return resolved verse text alongside the URI.
- **Phase 2.3 — similarity bridge.** All 69 studies are embedded.
  The bridge pattern (pgvector cosine + AGE `:SIMILAR_TO` edge) is
  proven by the probe; just port it into a `stewards.refresh_study_similarity()`
  function. Done when "similar studies" appears in the study_show
  output.
- **Phase 2.4 — `stewards study show <slug>` CLI.** Pulls together
  row + citations (resolved) + similar studies into a single view.
  Done when running it on `give-away-all-my-sins` returns the
  study, scripture citations *with verse text*, and three similar
  studies.

## Phase 2 — original deliverable list (preserved for reference)


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

## Phase 3 — Pipelines + MCP + External arms: agents that work without an IDE

**Goal:** long-running agent work runs without an open VS Code
window. Tool sidecars execute work; the bgworker dispatches; results
flow back to a thin web UI for review. **Multi-model dispatch** —
the substrate routes to Anthropic (Opus/Sonnet via API), Google
(Gemini Pro/Flash, Veo, TTS), Kimi k2.6 (via opencode go/zen),
and local models (LM Studio). Token cost per task becomes a queryable
metric.

> This is the long-form spec for what the proposal
> ([phase-2-5-generic-substrate.md](../../.spec/proposals/pg-ai-stewards-phase-2-5-generic-substrate.md))
> sketched as "Phase 3 — External arms." The two are the same phase,
> different altitudes. The proposal sketch named the *what*
> (multi-model dispatch, sandboxed git, MCP wiring, tokenomics as
> telemetry); this section names the *how* (pipeline tables, sidecar
> protocol, web UI, becoming integration).

### Inaugural pipeline (POC) — automated scripture studies

First pipeline to land: **the system writes its own scripture studies
using gospel-engine-v2 + the existing MCPs as tools.** Drop a
`study_request` row (binding question + scope), bgworker dispatches,
Kimi k2.6 (or whichever model is configured) does discovery via
`gospel_search`, reads sources via `gospel_get`, builds intermediate
**scratch documents** as the work proceeds (collections of quotes,
notes, outlines — each row, each with the original source URI
preserved as a graph edge), drafts the study, self-reviews against
`source-verification` skill, and **inserts the finished study as a
new row in `stewards.studies` with `kind='study'`**. From there the
embed trigger fires, similarity edges grow, and the human reviews via
the web UI (becoming/ibeco.me, eventually). **Output is never a file
on disk.** This is the proof of concept that the substrate can do the
agent work itself.

> **Architectural principle (named explicitly here because the file-vs-DB
> distinction shapes everything downstream):** pg-ai-stewards is
> **DB-centric**. Outputs that are *documents* — studies, scratch notes,
> outlines, quote collections, journal entries the agent writes — live
> as DB rows, not files. This means every output is **immediately a
> graph citizen**: it gets an embedding, similarity edges, frontmatter-
> declared edges, and shows up in `context_for()` walks the next
> instant. No filesystem round-trip, no re-import step, no slug
> collisions. The graph self-grows and self-links from the *process*,
> not just the artifact.
>
> File writing is reserved for **code and binary deliverables** that
> don't fit a DB — generated source files, images, audio (TTS for
> Marsfield/Empty Epsilon), video (Veo), `.go` files for new sidecars.
> Everything else lives in Postgres.
>
> **Reading surface.** The web UI lives in `becoming/` (ibeco.me) —
> the same cloud hub the brain ecosystem already uses. Studies render
> as DB-backed pages with their citations, similar studies, and graph
> neighborhood inline. study.ibeco.me (WS4) is already pointed at the
> reading surface; pg-ai-stewards becomes its second backend.
>
> **Working name for the agent surface: `a.ibeco.me`** ("a becoming" /
> "ai become"). Sibling to study.ibeco.me. Not a code editor — a
> worklist, a review queue for agent-produced studies, and an
> in-flight model-call inspector. Don't build until Phase 3's
> bgworker side is real; the meaningful version of this UI is the
> one that lets you watch an agent work, intervene, accept/reject —
> which requires the producer side to exist first.

Why scripture studies as the POC and not something more impressive:

- **Stakes are right-sized.** A bad study is a learning opportunity,
  not a production outage.
- **Tooling already exists.** gospel-engine-v2, webster, byu-citations
  are running today. No new MCP work blocks it.
- **Verification is concrete.** The `source-verification` skill +
  cite-count rule + read-before-quoting are checkable. The agent's
  work can be graded against the same standards we hold ourselves to.
- **Output has independent value.** Even imperfect drafts save
  setup time on real studies.

If this works for studies, the next pipelines (Marsfield exhibit
briefs, D&D campaign elements, Empty Epsilon mission scripts) become
variations on the same shape.

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
6. **Multi-model dispatch.** Each model is a tool sidecar with its
   own credentials and pricing. Per-call token cost lands as a
   `stewards.model_calls` row alongside the work_item. `tokenomics`
   (Michael's coining) becomes first-class queryable telemetry, not
   an estimate. First sidecars: Anthropic (Opus/Sonnet API),
   Google (Gemini Pro/Flash, Veo, TTS — for space-center work),
   Kimi k2.6 (via opencode go/zen), LM Studio (local models).

### Done when

- A long-running task ("review all studies and identify the three
  that contradict each other") can be queued from the web UI,
  runs without an IDE window, posts a result back, and Michael
  reviews it from any device.
- The web UI can interrupt an in-flight model call.
- **The scripture-study POC has produced at least 3 studies that
  pass source-verification on first review** (no fabricated quotes,
  links resolve, citations match what's actually in the cited text).

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
- **Multi-tenant RLS** if ibeco.me ever hosts other people. **Open
  design questions worth recording now (not solving):**
  - **Workstream as the unit of sharing.** Workstreams are already
    vertices with stable ids. Visibility scopes (private / group /
    public) attach naturally there; everything reachable through
    `:HAS_PROPOSAL` / `:HAS_TODO` inherits unless explicitly
    overridden.
  - **Row-level: `owner_user_id` + `visibility` on every table.**
    Add to `stewards.studies`, `stewards.todos`, `stewards.workstreams`,
    `stewards.model_calls`. RLS policies based on session-set
    `app.current_user_id`.
  - **The AGE caveat.** pgvector RLS works naturally (it's a
    standard `WHERE`). **AGE Cypher does NOT automatically enforce
    RLS on traversed labels** — a `MATCH (a)-[:CITES]->(b)` can
    return `b` rows the current user shouldn't see. Mitigation
    options to evaluate: (a) every Cypher query post-filtered by
    joining `cypher()` output back to the SQL table with RLS active,
    (b) per-user materialized graph subsets, (c) write a Cypher
    wrapper that injects visibility predicates. Path (a) is
    cheapest; (b) is fastest at read time; (c) is most correct.
    Decide when first non-Michael user is real.
  - **API key vault.** New `stewards.user_secrets` table
    (`user_id, provider, key_encrypted, created_at`), encrypted at
    rest with `pgcrypto`. bgworker fetches per-user keys at
    dispatch time; per-user tokenomics roll up against the same
    keys. Never logged, never returned through the MCP surface.
  - **Edge ownership.** Whichever endpoint has lower visibility wins
    — a private todo on a public workstream is private; a public
    annotation on a private study is private. Default-deny.
  - **Not before:** authentication (let becoming/ibeco.me's existing
    OAuth do the work — `app.current_user_id` is set from the
    authenticated session), audit logging, sharing-link UX. Those
    are surface-level concerns once the substrate enforces correctly.
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
