---
date: 2026-05-08
agent: dev
session_kind: substantive
tags: [pg-ai-stewards, 3e, mcp, bgworker, sidecar, async-fan-out]
priority: medium
carry_forward:
  - 3e.2.d — auto-promote `mcp_tool_cache` rows into `stewards.tool_defs` with deny-by-default `agent_tool_perms` (today's 3 tool_defs hand-curated)
  - per-agent grants for the new mcp_proxy tools (currently deny-by-default — soak surface untouched)
  - bridge session crash recovery — if a stdio MCP server dies mid-conversation, future calls fail and the bridge needs a restart to reinitialize. v2 should detect connection errors and rebuild on next access
  - real chat exercising mcp_proxy end-to-end — synthetic test passed but the assistant-loop integration (chat → tool_calls → mcp_proxy → completion → continuation chat) hasn't run organically yet
  - search-mcp protocol regression from 3e.2.a still outstanding — DuckDuckGo MCP returns "invalid request" on tools/list
---

# 3e.2.b/c v1 — async fan-out mcp_proxy + bridge daemon

Built the second half of the outbound MCP-client path. Substrate-internal
agents can now route `tool_calls` through external MCP servers via
`execute_target='mcp_proxy'`. The wiring uses **async fan-out** rather
than block-poll — Michael's deliberate architectural choice when I
surfaced the trade-off — so a chat with N parallel mcp_proxy tool_calls
resolves through the bridge concurrently instead of monopolizing one
bgworker per call.

**The shape, briefly.** When `tool_dispatch` sees an mcp_proxy tool, it
enqueues a child `kind='mcp_proxy'` row and returns
`WorkOutcome::WaitingForTools` instead of `ToolsDispatched`. The
bgworker writes the parent row to `status='waiting_for_tools'` with a
`{resolved: [...sync], pending: [{tc_id, name, child_work_id}, ...]}`
split persisted in `result`. A new SQL completion pass
(`tool_dispatch_complete_waiting()`) — called from every bgworker tick,
not just the leader — joins pending children by id and, when all are
done/errored, runs the original Phase 3 work (insert tool messages +
enqueue continuation chat). FOR UPDATE SKIP LOCKED keeps the four
workers from racing on the same parent row.

**The bridge daemon.** New `stewards-mcp bridge run` subcommand. LISTEN
on `stewards_mcp_proxy` + 1s safety tick. Claim oldest pending
mcp_proxy row, hand to a worker goroutine, look up or lazy-init a
session, `CallTool`, write result, NOTIFY `stewards_done`. Sessions
cached by server name in a `sync.Map` — first call pays subprocess
spawn (~10s warmup for gospel-engine-v2), subsequent calls under 200ms.

**What worked first try.** The Rust refactor was clean: `ToolReply` enum,
new `WorkOutcome` variant, completion pass call in the tick loop. Build
succeeded on first attempt for both Rust and Go. Async fan-out's
abstraction held — sync sql_fn/http path is unchanged, mcp_proxy is
purely additive.

**What surprised.** Two real issues:

1. **The 3e.2.a seed used `$$env:` (over-escaped dollar).** I'd written
   `'$$env:GOSPEL_ENGINE_TOKEN'` in the SQL because I wasn't sure if a
   single dollar would trip Postgres's dollar-quote parser inside a
   string literal. It doesn't — string literals are immune to
   `$tag$ ... $tag$` quoting — and the doubled dollar got stored
   verbatim. My `resolveSecret()` only stripped `$env:` (single), so the
   subprocess saw the literal placeholder as the token and gospel-engine
   returned 401. Fixed in `bridge.go` to handle both prefixes; left the
   stored data alone since either form works now.

2. **psql `:variable` substitution doesn't enter `DO $$ ... $$` blocks.**
   First verify-3e2-2.sql polled status inside a DO block referring to
   `:enqueued_id`. Got "syntax error at or near `:`" because psql does
   NOT substitute inside dollar-quoted strings. Fix: drop the polling
   loop and use `SELECT pg_sleep(10)` — bridge is fast enough that a
   fixed wait is fine for synthetic tests.

**Stewardship moment.** When I caught the `$$env:` bug, the boundary-
test asked: would Michael, asked in advance, want the resolver to
handle both forms, or want me to re-migrate the data? Both forms is
clearly the answer — it's a one-line resolver change vs. a migration
that complicates idempotent reseeds for no gain. Fixed and reported.

**End-to-end verified.** Synthetic test:
- `mcp_proxy_enqueue('gospel-engine-v2', 'gospel_search', '{"query":"faith hope charity","limit":3}', NULL)` → enqueued id=821
- Bridge claimed within 2.4ms (NOTIFY-driven, no poll latency)
- Total round-trip 258ms (after one-time session warmup)
- 3 real corpus results returned: "Titles, Tunes, and Meters" manual,
  Uchtdorf's "The Infinite Power of Hope" talk, "1 Corinthians 8–13"
- `isError=false`, `status=done`

The completion pass path (`tool_dispatch_complete_waiting`) wasn't
exercised in the synthetic test because there's no parent
tool_dispatch row pointing at id=821. That integration runs the first
time a real chat calls a granted mcp_proxy tool — which currently
nothing does (deny-by-default). The SQL function compiled and the
JSON shapes match what the bgworker writes, so confidence is high but
not yet verified-by-execution.

**Time.** Roughly 1.5 hours. SQL migration ~20 min, Rust refactor ~30
min, Go bridge daemon ~30 min, build + smoke + verify + the two
debugging hops ~20 min. Most of the time was carefully refactoring
exec_one_tool's signature (sync→Result<ToolReply, _>) and the
WorkOutcome write-arm split — invasive enough to deserve attention,
mechanical enough that the tests held it.

**Why this matters.** Substrate-internal agents now have a path to
the outside world. The next time we run a real chat with a granted
mcp_proxy tool, the agent can `gospel_search` against engine.ibeco.me
through the bridge instead of being stuck with sql_fn substrate-
internal tools. That unlocks gospel_passthrough (3e.5) and changes
what kinds of pipelines we can run inside the substrate.

The architecture also opens future work: 3e.2.d auto-promotion (every
cached MCP tool becomes a tool_def automatically), per-agent grants
(operator workflow for "let kimi-tuned-study use webster_define"),
and Linux-in-Docker bridge (currently runs on the Windows host
because that's where the .exe binaries live).

Commit: `bf4cb7c`. Findings: `projects/pg-ai-stewards/docs/3e-mcp-findings.md`.
