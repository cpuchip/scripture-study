# 3e MCP server build-out — findings

*Started 2026-05-08, soak paused for the duration.*

Working journal of the MCP sidecar implementation (Phase 3e). Updated incrementally as sub-phases land. Captures the surprises and patterns that future sub-phases (3e.2-3e.5) and post-shipped maintenance work should know about.

## Architecture (settled 2026-05-08)

**Sidecar binary at `projects/pg-ai-stewards/cmd/stewards-mcp/`** — Go process that runs alongside the substrate's pgrx Postgres extension. Connects to substrate via pgxpool, speaks MCP over stdio, exposes substrate tools to MCP clients (Claude Code, etc.).

**Why sidecar over in-pgrx HTTP listener.** Considered putting an HTTP listener inside the bgworker, but unconventional + would require Rust HTTP server code in pgrx (which is built for "extend Postgres," not "be a server"). Sidecar is the standard pattern (mirrors stewards-cli) and lets us use the official Anthropic+Google Go SDK.

**Why stdio over HTTP transport.** Claude Code spawns local MCP servers as child processes with pipes. stdio is the standard transport for that lifecycle. HTTP transport adds auth, port management, lifecycle complexity for zero gain.

## 3e.1 v1 — shipped 2026-05-08

### Scope

Two read-only tools over the studies corpus:

| Tool | Wraps | Notes |
|------|-------|-------|
| `study_search` | `stewards.study_search_text(query, kinds[], limit)` | FTS via websearch_to_tsquery; returns ranked hits |
| `study_get` | `stewards.study_get(slug, include_body, line_offset, line_count, max_chars)` | Line-paginated read; returns full jsonb body |

### Verified end-to-end

Manual smoke test via piped JSON-RPC stdin (without going through Claude Code):
- `initialize` (protocol 2025-11-25) → server responds with capabilities, serverInfo, protocolVersion ✓
- `notifications/initialized` → silent (correct — notifications get no response) ✓
- `tools/list` → both tools listed with full inputSchema and outputSchema ✓
- `tools/call` study_search "faith hope charity" → 3 expected hits in `result.structuredContent.results` ✓

### Files

```
projects/pg-ai-stewards/cmd/stewards-mcp/
├── go.mod          # standalone module, registered in workspace go.work
├── go.sum
├── main.go         # entry, pgxpool setup, server.Run on stdio
├── tools.go        # study_search + study_get handlers + toolError helper
└── README.md       # how to build, configure, troubleshoot
```

`projects/pg-ai-stewards/bin/stewards-mcp.exe` is the compiled binary.

### Surprises and gotchas (real, not theoretical)

1. **The SDK API differs from the research's stated examples in one place.** The research subagent (and many third-party docs) reference `mcp.NewToolResultError(...)` as a helper for tool-execution errors. **It doesn't exist in v1.6.0.** The actual pattern is to construct `&mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "..."}}}` directly. Discovered via `grep` in the SDK source after the first build failed. Captured a `toolError(format, args)` helper in `tools.go` to keep call sites terse.

2. **`go.work` registration is mandatory.** The repo uses a workspace go.work at the root listing all module paths. Forgot to add `./projects/pg-ai-stewards/cmd/stewards-mcp` and got `current directory is contained in a module that is not one of the workspace modules listed in go.work`. Easy fix once spotted; worth flagging in the skill so future module additions remember it.

3. **Smoke-test stdin needs a trailing sleep or pause** to keep the pipe open long enough for the server to respond before EOF triggers shutdown. First test just sent the JSON-RPC messages and saw an empty stdout because the server exited before flushing. `sleep 2` after the last message gave it enough time. Documented in the README.

4. **jsonschema struct-tag syntax for constraints isn't what I expected.** I wrote `jsonschema:"description=foo,minimum=1,maximum=100"` thinking it'd parse as separate constraints, but the SDK treats the whole tag value as the description string. Result: the JSON Schema's `minimum`/`maximum` constraints aren't emitted; the literal text "minimum=1,maximum=100" appears in the description. Cosmetic — the tools work fine — but worth fixing in a v1.1. Real syntax (per jsonschema-go docs) is multi-tag: `jsonschema:"description"` only, with constraints expressed via separate tag fields. **TODO for 3e.1.1.**

5. **Stdout buffering is a real risk that the SDK handles for us.** `os.Stdout` is fully buffered when not connected to a TTY (Claude Code spawns the binary as a pipe). The SDK flushes after every protocol message; the discipline cost lives entirely in "don't let any other code in the process write to stdout." `log.SetOutput(os.Stderr)` at the top of main is the necessary precaution.

6. **`.mcp.json` is gitignored on this project.** The existing entries have API tokens (gospel-engine-v2, becoming) so the file was added to `.gitignore` long ago. Adding `pg-ai-stewards` (which has no secrets, just a local DSN) means the entry can't be committed. Workaround: documented the entry shape in the cmd/stewards-mcp/README.md and in this findings doc. Future cleanup: maybe split tokens out into a separate file and commit a sanitized `.mcp.json`, but not for tonight.

7. **First-run approval dialog.** Project-scoped MCP servers prompt the user for approval the first time Claude Code spawns them. Documented for teammates so they don't think the server is broken on first session restart.

### What 3e.1 v1 does NOT include

- **Outbound MCP-client capability** (consuming gospel-engine-v2 from substrate-internal agents). That's 3e.2-3e.3. Requires extending the bgworker with `tool_http` execute_target, OR building an in-Go MCP client that the bgworker can call. Open design question.
- **stewards_brain / stewards_work_item / gospel_passthrough.** That's 3e.4-3e.5. Mechanical to add once we have v1's pattern proved.
- **JSON Schema constraints in input/output.** Items 4 above. Cosmetic v1.1.
- **Authentication.** Local stdio doesn't need it. If we ever expose stewards-mcp over HTTP for remote IDE clients, that becomes a real concern.
- **Resources** (the MCP "read-only data sources" surface). Not used by the substrate yet; not blocking anything.
- **Prompts** (the MCP "templated prompts" surface). Same — not used.

## 3e.1.1 — shipped 2026-05-08

### What changed

- **jsonschema struct-tag syntax fixed.** Per jsonschema-go's `For` documentation, tag values are description-only — there is **no** `description=foo,minimum=1,maximum=100` syntax. The library explicitly reserves `WORD=` prefixes for future syntax and forbids descriptions starting with that pattern. My v1 tags violated this in two ways: prepending `description=` and embedding `,minimum=1,maximum=100` constraints. Rewrote all tags to plain prose. Constraints would require manual `*Schema` construction; the substrate's own SQL functions enforce reasonable bounds, so we don't bother at the MCP layer.
- **`study_similar(slug, limit)` added.** Wraps `stewards.study_similar`. Live-tested: returns 0 results because substrate's similarity-edge graph isn't populated yet (separate Watchman 2.x work). Tool wrapper is correct; it'll surface results as edges materialize.
- **`study_citations(slug)` added.** Wraps `stewards.study_citations`. Live-tested against gadianton-robbers: 35 citations including Ether 8-9 (15 refs), Helaman 6:38-39 (13), Moses 5:51 (9), Mosiah 18:5 (1) — exactly the cross-references threaded through the secret-combinations argument.

### Surprises

- **The jsonschema-go library doesn't support constraints at all via struct tags.** The MCP SDK's research output suggested `jsonschema:"minimum=1,maximum=100"` as syntax; that was wrong. The actual library reads only the description string. Constraints (min, max, enum, format) require constructing `*Schema` manually. This is a real limitation worth noting — captured in the mcp-server-go skill.

## 3e.4 v1 — shipped 2026-05-08

### What changed

Four read-only inbound tools in a new `inspection.go` module:

- **`work_item_list(pipeline?, status?, limit?)`** — list recent work_items with optional filters. Returns `id, slug, pipeline, current_stage, status, tokens_in, tokens_out, token_budget, actor, created_at, updated_at, completed_at`.
- **`work_item_show(id_or_slug)`** — full detail for one work_item including `stage_results` JSONB and original `input`. Looks up by UUID OR slug.
- **`watchman_passes_list(limit?)`** — recent soak passes with `pass_id, status, trigger, started_at, finished_at, provider, model, agent_family, doc_count_planned/done, tokens_in/out, token_budget, budget_stopped, verdict_counts`.
- **`watchman_pass_show(pass_id)`** — one pass header plus per-doc verdicts (`study_id, verdict, reasoning, model, tokens, actor, created_at`).

### Live-tested

- `work_item_list(pipeline=study-write, limit=3)` returned the 3 expected voice-experiment runs in created_at DESC order.
- `watchman_passes_list(limit=2)` returned the 2 most recent soak passes with `verdict_counts` JSONB decoded properly into the response (e.g. `{clean: 3, drift: 1, skipped: 1}`).

### Decisions

- **Dropped `stewards_brain` from the original 3e spec.** The substrate's `brain_entries` table has 1 row vs `studies` 370. The v3→v4 migration consolidated brain corpus into studies; a separate `stewards_brain` tool wraps a dead table.
- **Write-mutating tools deferred to 3e.4 v2.** `work_item_create`, `work_item_dispatch`, `work_item_advance`, `watchman_pass_now`. Cost risk: a confused tool call could fire real model work. Mitigation: substrate's `token_budget` per work_item bounds blast radius, and Claude Code prompts for approval per tool. Still: let v1 read-only prove out before letting Claude Code drive the substrate.
- **Module split (inspection.go separate from tools.go).** Clean concern boundary — studies-corpus tools in one file, runtime-state inspection in another. Future write-tool ops will go in their own file.

## 3e.2.a v1 — shipped 2026-05-08

### Scope

First half of the outbound MCP-client path — schema + bridge daemon
skeleton + multi-bgworker concurrency. Substrate now knows about the
external MCP servers and can refresh their tool catalogs on demand.
What's still missing (deferred to 3e.2.b/c) is the LISTEN/NOTIFY wire
between bgworker `execute_target='mcp_proxy'` rows and the bridge.

### What changed

- **`stewards.mcp_servers`** registry table — name PK, transport
  (stdio|http), command/args/url/env, enabled flag, telemetry columns
  (last_health_check_at, last_tools_refresh_at, last_error). Transport
  CHECK enforces stdio→command, http→url. Seven seed rows for
  gospel-engine-v2, webster, yt, byu-citations, becoming, search,
  exa-search — all `enabled=false` initially.
- **`stewards.mcp_tool_cache`** table — per-server tool catalog
  populated by `bridge refresh-tools`. Keys on (server_name, tool_name);
  active=false soft-hides without losing schema. `input_schema` and
  `output_schema` stored as jsonb so 3e.2.d can synthesize
  `stewards.tool_defs` rows.
- **`stewards.mcp_bridge_state`** view — at-a-glance: which servers
  are responding, when they were last checked, how many tools cached.
- **Multi-bgworker registration.** `_PG_init` now registers N (default
  4, max 16, override `STEWARDS_DISPATCHER_WORKERS`) dispatcher
  workers. Each tick-loops on the same `process_one_pending` claim
  but `FOR UPDATE SKIP LOCKED` keeps them from racing. **Worker 0 is
  the "leader"** — owns the once-per-postmaster stale-claim reaper
  and the periodic Watchman scheduler tick. Other workers skip both
  to avoid duplicating work.
- **`stewards-mcp bridge refresh-tools`** subcommand. Reads
  `mcp_servers WHERE enabled` (or `--all`), connects via
  `CommandTransport` (stdio) or `StreamableClientTransport` (http),
  calls `tools/list`, upserts results, stamps timestamps. Failures
  are recorded in `last_error`; one server's failure doesn't abort
  the rest. Resolves `$env:VARNAME` placeholders in the row's `env`
  jsonb against the bridge process's environment.

### Live-tested 2026-05-08

```
$ stewards-mcp bridge refresh-tools --all --timeout 60
Refreshing 7 MCP server(s)
  [ OK ] becoming             24 tool(s)
  [ OK ] byu-citations         3 tool(s)
  [ OK ] exa-search            1 tool  (Streamable HTTP, remote)
  [ OK ] gospel-engine-v2      3 tool(s)
  [FAIL] search                tools/list: invalid request
  [ OK ] webster               5 tool(s)
  [ OK ] yt                    4 tool(s)

Refresh complete: 6/7 successful
```

40 tools cached. Both transport types (stdio + Streamable HTTP)
verified. Multi-bgworker registration verified on both ephemeral
smoke and live container restart — 4 workers spawn, exactly one
becomes leader.

### Surprises

1. **search-mcp protocol mismatch.** Our DuckDuckGo MCP server
   returns "invalid request" to `tools/list`. Likely an older MCP
   protocol version that doesn't recognize the method, or a
   deserialization bug in our server code. Out of scope for 3e.2.a;
   noted in `last_error` so the operator sees it. Lowest-priority
   server of the seven (DuckDuckGo can be hit via plain HTTP if
   needed — that's the kind of fallback Path A `execute_target='http'`
   exists for).

2. **`StreamableClientTransport` doesn't expose Headers field.** For
   bearer-token-authenticated remote MCP servers, auth has to flow
   through a custom `http.Client.Transport` wrapping
   `http.DefaultTransport` with a `RoundTripper` that injects
   `Authorization`. Tonight's seed only has exa-search using token
   in URL (`?token=...`), so we don't hit this. Document for the
   first server that needs it.

3. **`extension_sql_file!` foldback was mechanical.** Just added one
   `extension_sql_file!("../3e2-1-mcp-bridge-schemas.sql", name=...,
   requires=["create_work_items_to_studies_promotion"])` block in
   `lib.rs` and one entry to the Dockerfile COPY. The pgrx-rust skill
   was right: file location of `extension_sql!`/`extension_sql_file!`
   doesn't matter; only the dependency-graph names do. Built clean
   first try.

4. **`pg_sys::Datum::from(u64)` + `arg.value() as usize`** is the
   correct round-trip for passing a worker index through
   `BackgroundWorkerBuilder::set_argument()`. Verified working;
   compiled clean first try.

5. **Soft-deactivation of stale tools.** When `tools/list` returns
   N tools, we `UPDATE ... SET active=false WHERE tool_name <> ALL($2)`
   — but Postgres doesn't accept `<> ALL(empty array)`. Padded with
   a sentinel empty string when the server returns zero tools. Edge
   case but worth getting right.

### What 3e.2.a does NOT include

- **bgworker `execute_target='mcp_proxy'` dispatch arm.** Substrate
  agents can't yet route a tool call to the bridge. That's 3e.2.c.
- **LISTEN/NOTIFY wire** between substrate and bridge. Same.
- **Long-running `bridge run` daemon.** Today's `refresh-tools` is a
  one-shot. Daemon mode is needed before `mcp_proxy` can work.
- **Auto-promotion of cached tools to `stewards.tool_defs`.** Each
  cached tool would synthesize a tool_def with deny-by-default
  agent_tool_perms. That's 3e.2.d.
- **Bearer-token headers for HTTP transport** (item 2 above).
- **Linux-in-Docker bridge** (Michael's preferred long-run shape).
  Tonight runs the bridge on the Windows host because the binaries
  live there. Future migration when 3e.2.c lands.

### Files

```
projects/pg-ai-stewards/extension/
├── 3e2-1-mcp-bridge-schemas.sql   # registry + cache + 7 seed rows + view
├── src/lib.rs                     # +1 extension_sql_file! block
├── src/bgworker.rs                # multi-worker registration + leader gating
└── Dockerfile                     # +1 COPY entry

projects/pg-ai-stewards/cmd/stewards-mcp/
├── main.go                        # +bridge subcommand dispatch
└── bridge.go                      # NEW — refreshOneServer, transports, upserts
```

## 3e.2.b/c v1 — shipped 2026-05-08

### Scope

Second half of the outbound MCP-client path. Substrate-internal
agents can now (in principle — see "deny-by-default" below) route
tool calls through `execute_target='mcp_proxy'`. The bgworker emits
a child work_queue row, the bridge daemon picks it up, and the SQL
completion pass releases the parent tool_dispatch row when all
children resolve.

### Architecture chosen

**Async fan-out + continuation** over block-poll. When `tool_dispatch`
sees an mcp_proxy tool, it enqueues a child mcp_proxy row and returns
`WorkOutcome::WaitingForTools` instead of `ToolsDispatched`. The
bgworker writes the parent row to `status='waiting_for_tools'` with
`{resolved: [...sync], pending: [{tc_id, name, child_work_id}, ...]}`
in `result`. A new SQL function `tool_dispatch_complete_waiting()` —
called from each bgworker tick — joins pending children by id, and
when they're all `done`/`error`, runs the original Phase 3 work
(insert tool messages + enqueue continuation chat + promote to done).

This was a deliberate ~2× code investment over a simpler block-poll
pattern. The win is concurrency: a chat with three mcp_proxy tools
resolves them in parallel through the bridge instead of
sequentially monopolizing one bgworker. With four bgworkers and
parallel children, throughput stays alive even when individual
calls are slow.

### What changed

- **`stewards.work_queue.status` CHECK** gained `'waiting_for_tools'`.
- **`stewards.mcp_proxy_enqueue(server, tool, args, parent_id)`** —
  inserts a child row with `kind='mcp_proxy'`, `provider=<server>`,
  payload `{server, tool, args, parent_tool_dispatch_id}`. NOTIFY's
  `stewards_mcp_proxy` so the bridge wakes immediately. Refuses if
  the server isn't registered or `enabled=false`.
- **`stewards.tool_dispatch_complete_waiting()`** — completion pass.
  Concurrency-safe via FOR UPDATE SKIP LOCKED. Returns the count of
  rows promoted (for log-on-nonzero discipline).
- **`WorkOutcome::WaitingForTools`** + bgworker write arm — pauses
  the parent in `waiting_for_tools` with the (resolved, pending)
  split persisted in result jsonb.
- **`ToolReply` enum in tools.rs** — `Sync(String)` for sql_fn/http,
  `Async { child_work_id }` for mcp_proxy. `tool_dispatch` now
  branches on whether any async children were emitted.
- **bgworker claim filter** — `WHERE kind <> 'mcp_proxy'`. Bridge
  uses the inverse filter; the two sides partition by kind without
  coordinating beyond row-locks.
- **Reaper filter** — bgworker startup reaper skips mcp_proxy rows;
  bridge has its own startup reaper for those.
- **3 example tool_defs** routed to mcp_proxy: `gospel_search`,
  `gospel_get`, `webster_define`. **Deny-by-default** —
  agent_tool_perms not granted to any agent. Operators must
  explicitly allow these per-agent before substrate chats can use
  them. Preserves the existing soak surface.
- **`stewards-mcp bridge run`** subcommand — long-running daemon.
  LISTEN+claim+dispatch with N worker goroutines. Lazy
  per-server session cache (sync.Map under a single mutex).
  Graceful shutdown on SIGINT/SIGTERM. Reaps stale in_progress
  mcp_proxy rows on startup.

### Live-tested 2026-05-08

```
$ stewards-mcp bridge run --workers 2 --tick-ms 500
bridge run: connected to substrate
bridge run: spawned 2 worker(s); call-timeout=60s
bridge run: LISTENing on stewards_mcp_proxy

$ psql ...
SELECT stewards.mcp_proxy_enqueue(
   'gospel-engine-v2', 'gospel_search',
   '{"query":"faith hope charity","limit":3}', NULL);
                            -- enqueued id=821
                            -- bridge claimed within 2.4ms
                            -- result returned in 258ms total
                            -- 3 real search results
                            -- isError=false, status=done
```

End-to-end: tool_def lookup → mcp_proxy_enqueue → bridge claim →
session.CallTool → result write → row done. The path works.

### Surprises

1. **`$$env:` vs `$env:` secrets prefix.** The 3e.2.a seed SQL wrote
   `'$$env:GOSPEL_ENGINE_TOKEN'` (double dollar — I'd been unsure if
   `$env:` would be parsed as a dollar-quote tag, so I over-escaped).
   It's not — Postgres dollar-quotes need a closing tag, single `$`
   inside string literals is fine. But the stored value ended up
   with two dollars, and my `resolveSecret` only stripped `$env:`
   (single). Result: subprocess got the literal placeholder as the
   token, gospel-engine returned 401. Fix: `resolveSecret` now
   strips both `$$env:` and `$env:` prefixes. The 3e.2.a SQL stays
   as-is to avoid a needless re-migration.

2. **psql `:variable` substitution doesn't enter `DO $$ ... $$`
   blocks.** First verify-3e2-2.sql polled status inside a
   `DO $$ DECLARE ... LOOP ... END $$` block referring to
   `:enqueued_id`. Got "syntax error at or near `:`" because psql
   does NOT substitute inside dollar-quoted strings. Fix: drop the
   polling loop and use `SELECT pg_sleep(10)` — bridge is fast
   enough that a fixed wait is fine.

3. **Lazy session cache makes the first call expensive.** First
   gospel-engine-v2 call in a fresh bridge run paid ~12s of
   subprocess spawn + initialize handshake. Subsequent calls were
   under 200ms. For production agents this is fine (sessions
   persist across calls); for one-shot smoke tests it's worth
   knowing the warmup cost.

4. **The completion pass deliberately doesn't sit in the bgworker
   leader.** Earlier I wrote the watchman scheduler tick as
   leader-only to avoid duplicate firings; reflexively I almost did
   the same here. But unlike scheduler firings (which need
   exactly-once semantics), `tool_dispatch_complete_waiting()` is
   safe to run concurrently — `FOR UPDATE SKIP LOCKED` partitions
   work cleanly, and the function only commits side-effects after
   confirming all children are resolved. All 4 workers run it on
   each tick, so completion latency is ~500ms regardless of which
   worker first sees a row become ready.

5. **The bridge writes errors AS results.** When the MCP server
   returns `IsError=true` (tool-level failure — bad args, server
   error), the bridge still writes `status='done'` with the error
   payload in `content`. The model needs to see the error so it
   can recover. Real bridge-side failures (server unreachable,
   payload decode fail) write `status='error'` and the completion
   pass synthesizes a `{"error":"..."}` reply. This split was
   conscious — it mirrors the existing sql_fn/http arms.

### What 3e.2.b/c v1 does NOT include

- **Auto-promotion** of mcp_tool_cache rows into tool_defs (3e.2.d).
  Today's three example tool_defs are hand-curated; an operator
  must `INSERT` more by hand or write the auto-promotion SQL.
- **Per-agent grants** for the new mcp_proxy tools. Deny-by-default
  is the safe choice; the soak's existing tool surface is
  preserved. Operators flip `agent_tool_perms` rows when they want
  a specific agent to use the bridge.
- **Session crash recovery.** If a stdio MCP server crashes
  mid-conversation, future calls fail and the bridge needs a
  restart to re-init the session. v2 should detect connection
  errors and rebuild on next access.
- **Multi-bridge coordination.** Today only one bridge can run
  per substrate (NOTIFY/LISTEN works fine for any number, but
  there's no bridge ID claimed_by attribution). For HA we'd want
  bridge identity in the work_queue rows.
- **HTTP transport with bearer auth.** Same as 3e.2.a — first
  server that needs it adds the RoundTripper plumbing.

### Files

```
projects/pg-ai-stewards/extension/
├── 3e2-2-mcp-proxy-dispatch.sql  # status check + 2 SQL fns + 3 tool_defs
├── verify-3e2-2.sql              # synthetic e2e test
├── src/lib.rs                    # +1 extension_sql_file! block
├── src/types.rs                  # +WorkOutcome::WaitingForTools
├── src/tools.rs                  # ToolReply enum + mcp_proxy arm
├── src/bgworker.rs               # claim filter, write arm, completion pass
└── Dockerfile                    # +1 COPY entry

projects/pg-ai-stewards/cmd/stewards-mcp/
├── bridge.go                     # +run dispatch, $$env: handling
└── bridge_run.go                 # NEW — daemon, session cache, workers
```

## Future sub-phases (planned)

### 3e.2 — outbound HTTP path (the former 3c.4)
The hard part. Substrate-internal agents (running inside pipeline work_items) need to verify scripture quotes by calling gospel-engine-v2. Two options:

a) **Extend the bgworker** with `execute_target='http_proxy'` + a Rust HTTP client. Touches the chat-dispatch code path the soak depends on.

b) **Add MCP-client capability to the sidecar** + a new `execute_target='mcp_proxy'` that pgrx tool dispatch routes to the sidecar via... something. NOTIFY/LISTEN? A new work_kind that the bgworker dispatches and the sidecar consumes? The dispatch architecture isn't obvious.

Both have real risk. Need a focused design session before building.

### 3e.4 — stewards_brain / stewards_work_item
Inbound tools — IDE clients invoke them, sidecar runs SQL. Mechanical extension of the v1 pattern. Probably ~50 lines per tool.

### 3e.5 — gospel_passthrough
Inbound tool that wraps an outbound HTTP call to gospel-engine-v2. Easy IF 3e.2's outbound HTTP is built. Otherwise the sidecar needs its own HTTP client just for this.

## Skill

`.github/skills/mcp-server-go/SKILL.md` (symlinked to `.claude/skills/mcp-server-go`) captures the patterns and gotchas. Future sessions on MCP work should load it first. Updated based on findings here, especially:
- Item 1 (toolError helper, not NewToolResultError)
- Item 4 (jsonschema struct-tag syntax)

## Time and effort

- Research (subagent): ~3 min
- Skill authoring: ~10 min
- Plan 3e.1 v1 scope: ~5 min
- Build (main.go + tools.go + go.mod + go.work registration): ~20 min
- Build error iteration (NewToolResultError fix): ~5 min
- Smoke test + verification: ~10 min
- README + findings doc: ~10 min

**Total: ~1 hour 10 minutes** for v1. Same pattern as the lib.rs refactor: research front-loaded the unknowns, leaving the build mechanical.
