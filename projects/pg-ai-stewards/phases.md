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

## Phase 2.7 — Watchman (consolidation, dirty-bit, anti-loop)

Full design lives in the proposal: [phase-2-5-generic-substrate.md
§ Phase 2.7](../../.spec/proposals/pg-ai-stewards-phase-2-5-generic-substrate.md#phase-27--watchman-consolidation-freshness-anti-loop-discipline).
phases.md tracks delivery status only.

### Phase 2.7a — Watchman substrate ✅ done 2026-05-04

`last_consolidated_at` column, `stewards.verdicts` + `stewards.findings`
tables, `dirty_queue` view (terminal verdicts + open-finding suppression
encoded directly), `record_verdict` / `record_finding` /
`acknowledge_finding`, `study_history`. `stewards-cli watchman queue |
verdict | finding | ack | history` for human-driven passes.

### Phase 2.7b — Watchman automation (bgworker-driven)

> **Honest scope split (2026-05-06).** 2.7b as originally specced bundled
> bgworker, scheduling, token budget, and a 7-day soak. Splitting it
> the same way Phase 3 was split (3a/3b/...) so each piece ships
> independently. The trigger-based architecture (chosen over a Rust
> result-harvester) means the bgworker stays generic — Watchman
> semantics live entirely in SQL.
>
> | Sub-phase | What | Status |
> |-----------|------|--------|
> | **2.7b.1** | SQL substrate: `watchman_passes` + `watchman_config` tables, `watchman_pass_start()` enqueuer, `AFTER UPDATE` trigger on `work_queue` that records verdict/finding from completed watchman chats. CLI: `watchman pass-now`, `watchman passes`, `watchman pass-detail`, `watchman config show/set`. **No bgworker changes.** | **shipped 2026-05-06** |
> | **2.7b.2** | Bgworker scheduler tick (~60s). Reads `watchman_config`, calls `watchman_pass_start` when cron / pressure / idle trigger fires. SQL `watchman_should_fire()` owns the decision; Rust just polls. | **shipped 2026-05-06** |
> | **2.7b.3** | Per-pass token budget enforcement: `watchman_pass_start` stops enqueueing when projected tokens cross threshold; partial passes are valid. `estimate_chat_tokens(slug)` produces per-doc estimate from input length + system overhead + recent avg output. `budget_stopped` flag distinguishes "budget hit" from "queue empty / limit reached". | **shipped 2026-05-06** |
> | **2.7b.4** | `dirty_queue` excludes `frontmatter->>'watchman' = 'skip'` (option 3 from the design conversation); `regenerate_active_md()` returns markdown status report (In Flight / Findings / Todos / Recent Watchman / Corpus Stats). CLI: `watchman active-md`. The 7-day soak is the runtime observation that follows. | **shipped 2026-05-06** (soak: pending start) |

#### Phase 2.7b.1 — SQL substrate + completion trigger

**Goal:** replace the 3a Go orchestrator's polling loop with a
trigger-driven, transactional version. Manual triggering only; the
scheduler that wakes the pass automatically is 2.7b.2.

**Architecture (Option B — pure SQL + trigger):**

- `stewards.watchman_passes` — one row per pass. Carries pass-level
  config (provider, model, budget, trigger), planned/done counts,
  rolled-up tokens, and per-verdict counters advanced by the trigger.
- `stewards.watchman_config` — singleton (id=1) with default provider /
  model / agent / token_budget / cron_schedule / dirty_threshold /
  idle_threshold_hours / last_pass_at. 2.7b.2 reads it; 2.7b.1 just
  creates it with sane defaults.
- `stewards.watchman_pass_start(p_limit, p_provider, p_model,
  p_agent_family, p_actor, p_trigger, p_token_budget) → text` —
  inserts the `watchman_passes` row, pulls top-N dirty docs, for each:
  composes input via `watchman_input(slug)`, creates a session,
  inserts the user message, composes the body via `dry_run_chat`,
  enqueues a `kind='chat'` work_queue row whose payload includes
  `_watchman_pass_id`, `_watchman_slug`, `_watchman_actor`. Returns
  the new pass_id.
- **`AFTER UPDATE OF status` trigger on `stewards.work_queue`** with
  `WHEN ((NEW.kind = 'chat') AND (NEW.payload ? '_watchman_pass_id'))`.
  When a watchman chat row transitions to `done`/`error`:
    1. Reads the latest assistant message for the session.
    2. Strips optional ```json fences.
    3. Casts content to `jsonb`. Bad JSON → records `verdict='skipped'`
       with reasoning describing the parse error.
    4. Validates verdict against the 5-element enum. Invalid → `skipped`.
    5. Calls `record_verdict(...)`; if `verdict != 'clean'` and a
       `finding` object is present, calls `record_finding(...)`.
    6. Advances `watchman_passes` counters via
       `advance_watchman_pass_counters(pass_id, verdict, tokens_in,
       tokens_out)`. When `doc_count_done >= doc_count_planned`, marks
       the pass `completed`.
  All side-effects happen in the same transaction as the work_queue
  status flip — no race window where a row is `done` but its verdict
  isn't recorded.

**CLI additions** (`stewards-cli`):

- `watchman pass-now [--limit N] [--provider P] [--model M] [--budget T] [--actor A]`
  — calls `watchman_pass_start`, polls `watchman_passes` row until
  `status='completed'`, prints summary.
- `watchman passes [--limit N]` — list past passes.
- `watchman pass-detail <pass-id>` — verdict + finding rows for one pass.
- `watchman config show` / `watchman config set --schedule X --budget T
  --model Y --provider Z` — view/edit the singleton (used by 2.7b.2;
  schema-only role in 2.7b.1).

The existing `watchman pass` (3a Go orchestrator) stays for now as a
fallback — useful for repro and `--slug` single-doc runs without
creating a `watchman_passes` row.

**Done when:**

1. `pass-now --limit 5` enqueues 5 chats; bgworker drains them; trigger
   writes 5 verdicts; `dirty_queue` shrinks by 5; `watchman_passes` row
   shows `status='completed'` with verdict_counts populated.
2. `pass-now` against a slug whose model returns malformed JSON records
   `verdict='skipped'` cleanly (no trigger error, no work_queue stall).
3. Inverse hypothesis (Agans Rule 9): drop the trigger → run a pass →
   confirm verdicts NOT recorded; restore the trigger → re-fire the
   completion → confirm verdicts now recorded.

#### 2.7b.1 verification (2026-05-06)

All three "done when" gates met. Files:

- `extension/2-7b1-watchman-automation.sql` — live-DB migration applied,
  also referenced via `extension_sql_file!` in
  `extension/src/lib.rs` (sixth folded file: 2-6a/b/c, 2-7a, 3a, 2-7b1).
- `extension/verify-2-7b1-inverse.sql` — pure-SQL inverse hypothesis
  test (4 trials).
- `extension/verify-2-7b1.log` — captured CLI output of the 5-doc
  real-model verification pass.

**Smoke test (1 doc):** pass `watchman-20260506T200536Z-9b2de6` →
elapsed 3m22s (opencode_go was being unusually slow today), trigger
harvested verdict=`skipped` + drift finding cleanly, pass auto-marked
`completed` with `verdict_counts={"skipped":1}`.

**Inverse hypothesis (synthetic, no model tokens):**

| Trial | Setup | Expected | Got |
|-------|-------|----------|-----|
| 1 | trigger present, drift+finding JSON | 1 verdict, 1 finding, pass completed | ✓ |
| 2 | trigger DROPPED | 0 verdicts, 0 findings, pass stays in_progress | ✓ proves trigger is load-bearing |
| 3 | trigger restored, clean JSON | 1 verdict, 0 findings, pass completed | ✓ |
| 4 | trigger present, malformed JSON | verdict=`skipped` with parse-error reasoning, no raise | ✓ defensive path works |

**5-doc real-model verification (`actor=verify-2-7b1`):**

- 5/5 docs harvested in 7m45s. Tokens: 18902 in / 18677 out (well under
  50k budget). Verdicts: 1 clean + 4 skipped (kimi keeps surfacing the
  "I can't see external context" pattern from 3a, which is honest).
- 3 findings recorded (2 drift, 1 synthesis).
- **Real-world error-path validation:** `art-of-presidency` hit
  `opencode_go` HTTP 502 mid-pass. The trigger's error path recorded
  `verdict='skipped'` with `reasoning="watchman chat errored: chat
  HTTP 502 Bad Gateway: error code: 502"`. The rest of the pass kept
  going; doc_count_done advanced; pass auto-completed. Trial 4's
  defensive path proven in the wild.

**Architectural notes:**

- Trigger fires `AFTER UPDATE OF status` with WHEN-clause filtering
  on `kind = 'chat' AND payload ? '_watchman_pass_id' AND status IN
  ('done','error') AND OLD.status IS DISTINCT FROM NEW.status`. Cheap
  pre-filter on every work_queue UPDATE; only watchman rows allocate
  a function call.
- Every `record_verdict` / `record_finding` / `advance_counters` call
  is wrapped in `BEGIN...EXCEPTION WHEN OTHERS THEN ... END` so a bug
  in the harvester never breaks the bgworker's status flip. The
  trigger logs `RAISE WARNING` for any non-fatal failure.
- The 3a Go-orchestrator path (`watchman pass`) is preserved as a
  fallback. Same SQL fixtures, different control loop. Useful for
  `--slug` single-doc repro and Go-side log visibility.

#### Phase 2.7b.2 — bgworker scheduler tick (shipped 2026-05-06)

The bgworker now wakes the scheduler on its own. Three triggers
(pressure, cron, idle), all decided in SQL.

**Files**

- `extension/2-7b2-watchman-scheduler.sql` — adds 7 schedule columns
  to `watchman_config`, `watchman_scheduler_inputs()` (observability),
  `watchman_should_fire()` (decision: returns `'cron'|'pressure'|
  'idle'|NULL`), and `watchman_scheduler_fire()` (decide → if non-NULL,
  call `watchman_pass_start` with `actor='scheduler'`).
- `extension/src/lib.rs` — bgworker main loop now runs a 60s
  scheduler tick alongside the 500ms work-drain. `last_sched=None` on
  startup so the first tick happens immediately. Calls
  `stewards.watchman_scheduler_fire()` via `Spi::connect_mut`. Logs
  on fire only — silent on no-op (don't flood the postmaster log).
- `extension/Dockerfile` — added `2-7b1-watchman-automation.sql`
  and `2-7b2-watchman-scheduler.sql` to the `COPY` directive for the
  build context.
- `cmd/stewards-cli/main.go` + `internal/show/show.go` —
  `watchman config show` now displays scheduler fields;
  `watchman config set` accepts `--enabled`, `--min-interval-hours`,
  `--preferred-dow`, `--preferred-hour`, `--pass-limit`,
  `--pressure-cooldown-hours`, `--idle-cooldown-hours`. New
  `watchman scheduler-status` command prints the live decision and
  every input feeding it.
- `extension/verify-2-7b2-decision.sql` — pure-SQL decision-function
  verification (9 trials).

**Decision matrix (priority order)**

| Trigger | Fires when |
|---------|------------|
| (none) | `schedule_enabled = false` OR a pass started <1h ago is still in_progress |
| `pressure` | `count(dirty_queue) >= dirty_threshold` AND last_pass older than `schedule_pressure_cooldown_hours` |
| `cron` | last_pass older than `schedule_min_interval_hours` AND we're inside the preferred DOW + hour window (NULL = any) |
| `idle` | `idle_threshold_hours > 0` AND last_pass older than `schedule_idle_cooldown_hours` AND no `kind='chat'` session in N hours |

**Decision verification (9 SQL trials, no model tokens)**

All 9 trials pass:

| # | Setup | Got |
|---|-------|-----|
| 1 | `schedule_enabled = false` | NULL ✓ |
| 2 | dirty heavy, past cooldown | `pressure` ✓ |
| 3 | dirty_threshold = 9999 (suppresses pressure), DOW/hour mismatch | NULL ✓ |
| 4 | preferred DOW/hour = NULL (any) + min_interval=0 | `cron` ✓ |
| 5 | min_interval=168, last_pass 12h ago, dirty under threshold | NULL ✓ |
| 6 | idle_threshold_hours=1, idle_cooldown=1, no human sessions | `idle` ✓ |
| 7 | idle_threshold_hours=0 | NULL ✓ |
| 8 | inflight pass <1h old | NULL ✓ (don't pile up) |
| 9 | inflight pass >1h old (90 min) | `pressure` ✓ (allowed) |

**End-to-end live verification (real bgworker tick)**

After rebuild + restart with `schedule_enabled=true` and the corpus
358 docs over the 50 threshold:

- 21:36:21 UTC: bgworker started fresh
- 21:36:22 UTC: first scheduler tick (`last_sched=None`) → returns NULL because schedule was still disabled at that exact instant (we hadn't flipped it yet)
- 21:36:30ish UTC: human flipped `schedule_enabled=true`
- 21:37:22 UTC: 60s after `last_sched`, second tick fires
- 21:37:22 UTC: bgworker logs `stewards: scheduler fired Watchman pass: watchman-20260506T213722Z-0705d4`
- pass started with `trigger='pressure'`, `actor='scheduler'`, `doc_count_planned=5` (used `schedule_pass_limit=5` from config)

Within 60s of being enabled, the bgworker scheduler decided pressure
was hot, called `watchman_scheduler_fire`, which dispatched a real
pass through the same trigger-driven harvest path 2.7b.1 ships.

**Discovery during build:** `Spi::connect` is read-only and would
silently block `INSERT`/`UPDATE` operations performed by the SQL
function it invoked (PG SPI propagates the read-only flag down).
Switched `check_watchman_schedule()` to `Spi::connect_mut`. Caught
proactively (not by failure) by reviewing the existing
`process_one_pending` and reaper code, both of which use
`connect_mut`.

**Discovery during build:** the `Dockerfile` `COPY` directive lists
SQL files explicitly. New `extension_sql_file!` references in lib.rs
require updating that list — otherwise the rust compile fails with
`couldn't read 'src/../<file>.sql'`. 2-7b1 had been added to lib.rs
without updating the Dockerfile in 2.7b.1, but only failed now
because that section never got rebuilt — the live-DB migration
pattern lets the SQL file land without a docker rebuild. Added a
TODO-style comment in the Dockerfile as a reminder.

**Cost discipline.** The 2.7b.2 scheduler is structurally bounded
(per-pass `schedule_pass_limit`, per-trigger cooldowns, in-progress
guard) but no token budget is enforced inside `watchman_pass_start`
yet — that's 2.7b.3. The `schedule_enabled` master switch is the
load-bearing kill-switch until 2.7b.3 lands.

#### Phase 2.7b.3 — per-pass token budget enforcement (shipped 2026-05-06)

The `token_budget` column on `watchman_passes` was informational in
2.7b.1; 2.7b.3 makes it load-bearing.

**Files**

- `extension/2-7b3-watchman-budget.sql` — adds `budget_stopped boolean`
  column to `watchman_passes`, `stewards.estimate_chat_tokens(slug)`
  function, replaces `watchman_pass_start()` with a budget-aware
  version. Updates `watchman_pass_summary` view to expose
  `budget_stopped`.
- `extension/src/lib.rs` — eighth `extension_sql_file!` reference.
- `extension/Dockerfile` — added `2-7b3-watchman-budget.sql` to COPY.
- `cmd/stewards-cli/internal/show/show.go` —
  `WatchmanPasses` adds a `BUDGET` column ("ok" / "STOPPED").
  `printWatchmanPassDetail` shows a ⚠ marker on the tokens line
  when `budget_stopped=true`.
- `extension/verify-2-7b3-budget.sql` — pure-SQL verification (4
  trials at budgets 1000 / 10000 / 25000 / 999999); aborts each test
  pass before dispatch so no model tokens are spent.

**Estimation formula**

```
estimate(slug) = chars(watchman_input(slug)) / 4   -- input tokens
              + 1500                                -- system + persona overhead
              + avg(verdicts.tokens_out, last 30d)  -- output (3500 fallback)
```

Per-doc estimates ranged 7700–14500 across the live corpus.

**Enforcement**

In `watchman_pass_start`, before enqueueing each candidate doc:
- Compute `v_estimate = estimate_chat_tokens(slug)`.
- If `v_planned_tokens + v_estimate > v_budget`, exit the loop and
  set `v_budget_stopped = true`.
- Otherwise enqueue + add `v_estimate` to running total.

Stricter than the obvious "always allow at least one doc" rule: if
the FIRST doc's estimate exceeds the budget, refuse to enqueue
(empty pass, `budget_stopped=true`). Honest signal that the budget
is unworkable.

The estimate is also written into the work_queue payload as
`_watchman_estimate` for post-hoc analysis (e.g., comparing
estimate-vs-actual to refine the formula).

**Verification (4 SQL trials, no tokens spent)**

| # | Budget | Result | Why |
|---|--------|--------|-----|
| 1 | 1000 | 0 planned, `budget_stopped=true`, status=`completed` | First doc estimate ~9748 alone exceeds 1000 |
| 2 | 10000 | 1 planned, `budget_stopped=true` | First doc 9748 fits; +second 8794 → 18542 stops |
| 3 | 25000 | 2 planned, `budget_stopped=true` | 9748 + 8794 = 18542 fits; +14400 → 32942 > 25000 stops |
| 4 | 999999 | 5 planned (limit), `budget_stopped=false` | All 5 fit, hit `p_limit` not budget |

Each trial calls `watchman_pass_start`, observes the result, then
aborts the test pass via a `pg_temp.abort_test_pass(pass_id)` helper
that marks pending work_queue rows errored before the bgworker can
dispatch them. Zero model tokens spent.

**What this deliberately doesn't build**

- **Mid-pass abort.** If a chat already enqueued runs much longer
  than its estimate predicted, the actual spend will exceed budget
  by some margin. Acceptable for v1. If it bites, 2.7b.3.1 could
  add an actual-spend watcher that aborts pending chats once
  realized cost crosses budget.
- **Estimate calibration.** The formula uses raw averages from the
  last 30 days. A more sophisticated approach would weight by doc
  size or model. Defer until we observe estimate-vs-actual drift.

**Why the master kill switch is no longer load-bearing**

Pre-2.7b.3, `schedule_enabled=false` was the only way to prevent a
runaway scheduler loop on 350+ dirty docs. Post-2.7b.3, the budget
caps spend per pass: even if pressure fires every hour for the next
day, total spend is bounded by `(passes_per_day × token_budget)`,
which is a knowable number. The master switch remains a kill switch;
it's no longer the *only* kill switch.

#### Phase 2.7b.4 — soak prep (shipped 2026-05-06)

Closed the `watchman-frontmatter-exempt` todo (1c503ff6) AND shipped
`regenerate_active_md()` in one cut. The soak itself (the third
deliverable per the original 2.7b.4 plan) is runtime observation,
not code; starts when `schedule_enabled=true` is flipped on for a
sustained period.

**Files**

- `extension/2-7b4-watchman-soak-prep.sql` — modifies `dirty_queue`
  view to exclude docs where `frontmatter->>'watchman'` is `'skip'`
  or `'exempt'`; adds `stewards.regenerate_active_md()` returning
  markdown text.
- `extension/src/lib.rs` — ninth `extension_sql_file!` reference
  (`requires = ["create_watchman_budget"]`).
- `extension/Dockerfile` — added `2-7b4-watchman-soak-prep.sql` to
  the `COPY` directive in stage 1.
- `cmd/stewards-cli/internal/show/show.go` — new `WatchmanActiveMD`
  function. `cmd/stewards-cli/main.go` — new `watchman active-md`
  subcommand.

**Frontmatter exemption — option 3 implementation**

```sql
-- New gate in dirty_queue:
AND coalesce(lower(s.frontmatter->>'watchman'), '')
    NOT IN ('skip', 'exempt')
```

`lower()` for case insensitivity. `NOT IN ('skip','exempt')` because
both are reasonable spellings; we accept either. The frontmatter
`jsonb` column + GIN index already exist (Phase 2.1) — zero schema
change. Users add `watchman: skip` to YAML and re-import.

**`regenerate_active_md()` sections**

- **In Flight** — workstreams (status='active') with their declared
  proposals (joined via `frontmatter->>'workstream'`). Workstreams
  without proposals show `_(no declared proposals)_`.
- **Open Findings** — drift + synthesis findings without
  `acknowledged_at`, ordered by severity (high → low). Shows
  message + suggested action, indented.
- **Open Todos** — `status IN ('open','in_progress')`, grouped by
  parent. `▶` marker for in-progress.
- **Recent Watchman Activity** — last 5 passes with verdict_counts.
- **Corpus Stats** — markdown table with kind / total / embedded /
  in-dirty-queue. The dirty-queue column shows the gates working:
  e.g., a corpus where journals are tagged `watchman: skip` would
  show `journal | 70 | 70 | 0` instead of the current `70 | 70 | 70`.

**What it deliberately doesn't do**

- **Doesn't write to disk.** Returns text. Caller decides where it
  goes. CLI prints to stdout; future Watchman automation may pipe to
  `.mind/active.md`.
- **Doesn't include human-curated sections** (Priorities, Key Facts).
  Those live in the hand-written `.mind/active.md` and are not
  derivable from substrate state.
- **Doesn't auto-tag journals.** The dirty_queue gate is in place;
  the YAML files still need `watchman: skip` added before the soak
  starts. Easy bulk edit; deferred to be human-driven.

**Soak gating**

Before flipping `schedule_enabled=true` for the 7-day soak:

1. Bulk-tag all `kind='journal'` YAML files with `watchman: skip`
   (in `.spec/journal/*.yaml`). 70 files; trivial sed/python loop.
2. Re-run `stewards-cli import --source journal:.spec/journal` so
   the substrate picks up the tags.
3. Verify `SELECT count(*) FROM stewards.dirty_queue WHERE kind='journal'` returns 0.
4. Set `schedule_enabled = true`.
5. Watch trend: `tokens_in + tokens_out` per day across
   `watchman_passes` should *decline* as the corpus stabilizes. If
   it rises or stays flat after a few days, the discernment loop is
   leaking somewhere.

## Phase 3 — Pipelines + MCP + External arms: agents that work without an IDE

> **Honest scope split (2026-05-05).** Phase 3 as originally written
> bundled six deliverables. Same trap as Phase 2.7. Shipped as 3a; the
> rest are real but not blocking and ship when needed:
>
> | Sub-phase | What | Status |
> |-----------|------|--------|
> | **3a** | Model dispatch + Watchman pass (CLI orchestrator on top of bgworker) | **shipped 2026-05-05** |
> | **3a.1** | Agent + skill corpus import — `.github/agents/*.agent.md` (19) and `.github/skills/<name>/SKILL.md` (20) imported into `stewards.agents` / `stewards.skills` / `stewards.agent_tool_perms` via `stewards-cli import --source agent:... --source skill:...`. Tolerant YAML parser handles both Copilot list-style and Claude comma-string `tools` formats. Tool perms rebuilt deny-by-default + per-tool allow + skill loader allow. Idempotent reimport. Six `.github/agents/*.agent.md` files had malformed frontmatter (missing `tools:` key on the bare-list line); fixed in same commit. | **shipped 2026-05-06** |
> | **3b** | Input shaping for big docs (trim or bump bgworker reqwest timeout) + `response_format: json_object` injection | **shipped 2026-05-06** |
> | 3c | `stewards.pipelines` + `stewards.work_items` tables (deliverable 1 below) | **in flight** (3c.1 shipped 2026-05-07; 3c.2 auto-advance trigger + 3c.3 first real pipeline pending) |
> | **3c.1** | Pipelines + work_items schema, transition functions (`work_item_create/dispatch_stage/advance/fail/cancel`), seed `echo-test` pipeline, CLI surface (`pipeline list/show`, `work-item create/list/show/dispatch/advance/cancel`). Manual transitions only — same architecture as 2.7b.1 (orchestration in SQL, dispatch via existing `work_queue`, payload markers `_work_item_id` / `_stage_name` / `_pipeline_family`). | **shipped 2026-05-07** |
> | **3c.2** | `AFTER UPDATE` trigger on `work_queue` that auto-advances work_items when their dispatched chat completes. Mirrors the Watchman harvest trigger from 2.7b.1. Includes intermediate-vs-final detection (tool_calls iterations don't advance), token rollup (always), and a token_budget gate that stops auto-dispatch with `awaiting_review` instead of failing. | **shipped 2026-05-07** |
> | **3c.2.5** | Substrate-internal `sql_fn` tools (`study_search_text`, `study_get`, `study_similar`, `study_citations`, `study_context_for`) so the imported study agent has a real tool surface. Plus token-budget hook columns on `tool_defs` (left empty until real costs observed) and a blanket `study_*: allow` perm grant to all non-watchman agents (the imported Copilot tool patterns don't bridge to substrate tools automatically). Path B from the 2026-05-07 conversation. → [proposal](../../.spec/proposals/pg-ai-stewards-3c-2-5-study-tools.md) | **shipped 2026-05-07** |
> | **3c.3 v1** | Stage input templating (`render_stage_input`, `resolve_template_path`) + 3-stage `study-write` pipeline (outline / draft / review). **First end-to-end run produced a real outcome — and surfaced 3 substrate bugs (auto-advance trigger).** Agent successfully used substrate tools (`study_search_text`, `study_get`, `study_citations`) but never reached synthesis before step-exhausted on all 3 stages. Bugs documented in journal; v1 ships the templating + pipeline; **3c.3.1 is the bug-fix follow-up.** | **shipped 2026-05-07** (with caveats) |
> | **3c.3.1** | Fix the 3 bugs from 3c.3 v1: (a) `v_is_final` NULL coalesce in trigger; (b) `chat_post_internal` propagates `_*` payload markers from the session's most recent chat — generic across watchman + work_items + future systems; (c) `agents.steps` bumped from 8 to 50 across all non-watchman agents (watchman stays at 1 for its single-shot-no-tools contract). | **shipped 2026-05-07** |
> | **3c.3.2** | Multi-stage pipeline run after 3c.3.1 — re-ran study-write on the same FtC/WtL binding question. **Produced a real 6-section meta-study with self-review revision notes.** 17m14s elapsed; 626K in / 64K out (~$0.30 with caching, well under 2M budget). The substrate now does meaningful agent work end-to-end. | **shipped 2026-05-07** |
> | **3c.3.3** | Importer `model_match` extension — `cmd/stewards-cli/internal/importer/agents.go` reads `model_match` from frontmatter (falls back to `'*'`), so per-model agent variants (e.g., kimi-tuned study prompt at `.stewards/kimi-k2.6/study.agent.md`) can be imported without overwriting the base. Tool perms only rebuild on default-variant imports (variants share family-level perms). README.md and other frontmatter-less files in agent dirs are skipped silently. **Followup found: `agent_tool_perms` needs source provenance** — substrate-internal broadcasts (e.g. 3c.2.5 `study_*: allow`) get wiped on reimport. Workaround applied: declared `study_*` in study agent frontmatter on both `.github/agents/` and `.stewards/kimi-k2.6/`. | **shipped 2026-05-08** |
> | **3c.3.4** | Multi-model voice experiment — re-ran FtC/WtL binding question 3 ways on top of the 3c.3.2 original: (#2) kimi-k2.6 + kimi-tuned prompt, (#3) qwen3.6-27b + base prompt, (#4) kimi-k2.6 + kimi-tuned + corpus access. Run #4 is the strongest output: scene-opener, claim-style headers, anti-symmetry framing, and **actively caught + removed two fabricated quotes from its own draft** via `study_search_text` corpus checks. Validated the kimi-tuned prompt against five of the six 2026-05-07 voice signatures. Six qwen-specific signatures captured for a future `.stewards/qwen-3.6/study.agent.md` variant. Comparison memo at `study/.scratch/two-triplets-comparison-2026-05-08/`. | **shipped 2026-05-08** |
> | **3c.3.5** | Auto-promote completed `work_items` into `stewards.studies` via AFTER UPDATE trigger + standard `import_study()` path. Slug namespace `'substrate--{work_item_slug}'` to avoid collision with workspace studies. Frontmatter records source/pipeline/tokens/actor for provenance. Backfilled 6 prior runs from the 2026-05-08 voice experiment. Future Watchman passes can now graph-walk pipeline-produced studies. | **shipped 2026-05-08** |
> | **3c.3.6 v1** | Split monolithic `extension/src/lib.rs` — first module extracted: `providers.rs` (~120 lines, Provider/ProviderRegistry/GospelEngineConfig + statics). Cleanest leaf in the file (pure data types, no pgrx macros, no `extension_sql!` blocks). Build clean, smoke-tested CREATE EXTENSION clean, live container restarted onto new image. lib.rs now 4138 lines (down from 4246). Future moves (tools.rs, bgworker.rs, schema.rs) documented in detail at `docs/lib-rs-refactor-findings.md` with risks identified per move. Stopped after move 1 to commit a clean ship rather than rush 2-3 moves under build pressure. | **shipped 2026-05-08 (v1)** |
> | **3c.3.6 v2-v4** | Remaining lib.rs splits: tools.rs (~390 lines, requires moving `WorkOutcome` to types.rs first), bgworker.rs (~1075 lines, `#[pg_guard]` placement risk for `_PG_init`), schema.rs (~2400 lines, uncertain if pgrx accepts `extension_sql!` outside crate root). Each documented in findings doc. | not started |
> | ~~3c.4~~ | gospel-engine-v2 HTTP tool registration. **Absorbed into 3e.2 (2026-05-08).** The HTTP-out infrastructure is the same shape whether the call site is substrate-internal agents (was 3c.4's purpose) or MCP passthrough (3e's `gospel_passthrough` capability). Building it twice would duplicate effort. Decision recorded after Michael flagged the redundancy: *"3e supersedes 3c.4 — register gospel-engine-v2 as a consumed tool via the same MCP infrastructure."* | absorbed |
> | **3c.3.3.1 (agent_tool_perms provenance)** | (followup from 3c.3.3) Added `source text` column to `agent_tool_perms` (`'frontmatter'` / `'broadcast'` / `'manual'`). Importer's `DELETE WHERE agent_family=$1` now filters `AND source='frontmatter'`. Substrate-internal broadcasts (3c.2.5 `study_*: allow`) survive reimports without being declared in agent frontmatter. Migration in `3c3-3-agent-tool-perms-provenance.sql`, foldback into lib.rs. Verified end-to-end: pre-state 19 broadcast + 275 frontmatter, post-reimport unchanged. | **shipped 2026-05-08 (overnight)** |
> | **3c.3.4.1 (qwen-3.6 variant)** | `.stewards/qwen-3.6/study.agent.md` authored targeting the six qwen-specific signatures from run #3 (tool-name confusion, broken `(#)` links, heavy tables, bold-clause density, mid-paragraph triadics, verbosity) plus six kimi-shared rules. Imported via `model_match='qwen*'`. Run #5 (qwen + qwen-tuned + corpus) dispatched against FtC/WtL for comparison vs run #3 baseline. | **shipped 2026-05-08 (overnight)**, run #5 in flight |
> | 3d | Tool sidecars: sandboxed git, shell (deliverable 2 below) | not started |
> | 3e | MCP server + client. Sidecar binary (probably `cmd/stewards-mcp/`) that connects to substrate via pgxpool, serves substrate tools to IDE clients (`stewards_search`, `stewards_brain`, `stewards_work_item`), AND consumes external MCP servers (gospel-engine-v2) for substrate-internal use. Sub-staging: 3e.1 transport, 3e.2 outbound HTTP path (former 3c.4), 3e.3 outbound tool registration for substrate-internal use, 3e.4 inbound substrate tools, 3e.5 `gospel_passthrough` inbound. **Most leverage for "usable shape" daily-use from Claude Code.** | not started |
> | 3f | Web UI surface = `a.ibeco.me` (deliverable 4 + becoming integration deliverable 5) | not started |
> | 3g | Multi-provider expansion: Anthropic, Gemini, Veo, TTS (deliverable 6 expansion beyond opencode_go + lm_studio) | not started |
> | **3h** | Per-model prompt tuning generalized — cross-topic validation of `.stewards/<model>/*.agent.md` variants, `study-bench` CLI mirroring `classify-bench`, extension to non-study agent families. **Prototype shipped 2026-05-08 (kimi-k2.6 + qwen-3.6 study variants validated on FtC/WtL only); full effort deferred until 3e/3c.4/3f land**, because voice tuning is downstream of substrate-as-external-tool surface and real-quote verification. References prior brain-classifier rubric work. → [proposal](../../.spec/proposals/pg-ai-stewards-per-model-prompt-tuning.md) | **deferred** (prototype complete) |
>
> 3a is the unblocker for 2.7b (Watchman bgworker), 2.8 (LLM-inferred
> edges), and the `a.ibeco.me` reading surface. The rest can ship in
> any order driven by actual need.

### Phase 3a — shipped 2026-05-05

**What landed:**

- `watchman-consolidator` agent family in `stewards.agents` (default
  + kimi-* variant, same prompt, temp=0, steps=1, no tools)
- `agent_tool_perms ('watchman-consolidator', '*', 'deny')` —
  structural enforcement (compose_tools is allow-by-default; new
  no-tool agents need explicit deny)
- `stewards.watchman_input(slug)` SQL function — composes the user
  message: doc body + 1-hop graph neighborhood from `context_for(slug, 1)`
- `stewards-cli watchman pass [--slug X] [--limit N] [--provider opencode_go] [--model kimi-k2.6] [--timeout 180] [--dry-run]`
  — Go orchestrator in `cmd/stewards-cli/internal/show/show.go::WatchmanPass`
- System prompt enforces strict JSON output (`{verdict, reasoning, finding?}`)
  with five verdicts: `clean | drift | done | superseded | skipped`

**Architectural choice:** the bgworker stays generic (just `chat`,
`embed`, `tool_dispatch`, `resolve_ref` work kinds — no
watchman-specific semantics in Rust). All watchman orchestration
lives in the CLI Go for 3a. When 2.7b lands, this same logic
transcribes into a Postgres bgworker scheduled pass.

**Verified end-to-end:** 2 model verdicts in `stewards.verdicts`
with actor=watchman, model=kimi-k2.6, tokens logged (734 in / 3861 out
on phase-pg-ai-stewards-0; 882 in / 1277 out on .scratch-README).
The first verdict was `skipped` with the reasoning *"I cannot verify
external artifacts not in the 1-hop neighborhood"* — kimi
self-surfaced a 3b agenda item. Discipline holds.

**Known limitations resolved in post-3a work (May 5–6):**

- **Bgworker timeout:** bumped from 120s to 600s; CLI `--max-input-chars`
  flag added with 60/40 head/tail split + elision marker. Big docs
  (proposal, scratch files) no longer time out with a 30K char trim.
  Verified: 180s timeout → ERROR; 660s + `--max-input-chars 30000` →
  `skipped` with synthesis finding. ✅
- **Non-determinism:** `response_format: {"type":"json_object"}` is now
  a first-class column on `stewards.agents`. `dry_run_chat` injects it
  when non-NULL. `watchman-consolidator` agent seeded with it at
  `temp=0`. Verified via `payload->'body'->'response_format'` in
  `stewards.work_queue`. ✅

**Foldback debt resolved (v0.2.0, May 5–6):** All five SQL files
(2-6a/b/c + 2-7a + 3a) are now folded into `extension/src/lib.rs`
via `extension_sql_file!`. Extension bumped to 0.2.0. `init/01-seed-workstreams.sql`
extracted as post-install script (avoids search_path corruption during
CREATE EXTENSION — documented as AGE-QUIRKS #9). ✅

**AGE-QUIRKS.md now at 9 entries** (added #9: `set_config` search_path
leak during CREATE EXTENSION corrupts PGRX pg_extern declarations).

### Phase 3b–3g — original full-scope spec (preserved)

Everything below is the as-written Phase 3 spec from 2026-05. It
remains accurate as a target; we just don't ship all of it at once.


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

## Phase 6 — AGE upstream contributions

**Status:** sketched 2026-05-04. Triggered by 8 AGE quirks accumulated
during Phase 2.6 work. Not committed — do this when pg-ai-stewards
reaches a steady state and we have bandwidth to invest in the
ecosystem rather than ship features.

### Goal

Give back to Apache AGE in the form of (a) PR fixes for the quirks
that are genuinely bugs and (b) upstream documentation for the
quirks that are spec-divergences-by-necessity. Reduces our future
technical debt; helps everyone else who hits the same walls.

### Catalog

Full catalog with reproductions, workarounds, and category
([bug-candidate / spec-divergence / by-design / our-mistake]) lives
at [docs/AGE-QUIRKS.md](docs/AGE-QUIRKS.md). Categories matter:

- **PR-worthy (bug-candidate):** quirks #2, #6, #7
  - Apostrophe-in-interpolated-Cypher error message + auto-escape.
  - `cypher()` 3rd-arg should accept any `ag_catalog.agtype`
    expression, not just placeholders.
  - `#>>` (and likely `->>`, `->`) should handle agtype scalars as
    pass-through.
- **Document upstream as caveats (spec-divergence):** quirks #1, #3, #5
  - `ON CREATE SET` / `ON MATCH SET` after MERGE.
  - Implicit `WITH *` between MERGE and a subsequent MATCH.
  - Variable-length path syntax operations.
- **By design (don't fight):** quirk #4 (labels as schema).
- **Our problem (don't bother upstream):** quirk #8.

### Why Phase 6 and not earlier

PRs need a working test environment, a familiar codebase, and
reviewer cycles. None of those are cheap. The honest move is to
file the catalog now (we have it), keep adding to it as we hit
new quirks, and contribute back when our own substrate is stable
enough that AGE work isn't pulling cycles from delivery.

### Done when (if pursued)

- At least one PR landed against Apache AGE for one of the
  bug-candidate quirks.
- AGE-QUIRKS.md updated with PR links + status (merged / open /
  rejected with reasoning).
- For rejected PRs, the quirk's category is updated to reflect the
  upstream verdict (e.g., "by-design per maintainer feedback").

### Risks

- AGE is Apache-governed; PR review can be slow. Don't block our own
  work waiting on upstream.
- A PR for #6 or #7 may require deep changes to the agtype type
  system. Scope each PR before committing time.

## Phase 7+ — Maybe-someday

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
**Phase 2.7a** (Watchman substrate) needs Phase 2.6 only — no model
dispatch required, just dirty-bit + verdict/finding tables.
**Phase 2.7b** (Watchman automation) needs Phase 3 (model dispatch).
This split means the anti-loop discipline is human-drivable now and
automatable when Phase 3 lands.
**Phase 6** (AGE upstream contributions) is fully independent of
delivery phases — pursue when steady-state allows.

## Cadence

No time estimates. Each phase ends when its "done when" criteria are
met. Move on after a brief sabbath reflection (`.spec/journal/`,
`.mind/active.md` updates) to capture what we learned.
