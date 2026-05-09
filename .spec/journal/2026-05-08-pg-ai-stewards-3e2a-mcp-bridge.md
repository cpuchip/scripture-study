---
date: 2026-05-08
agent: dev
session_kind: substantive
tags: [pg-ai-stewards, 3e, mcp, bgworker, sidecar]
priority: medium
carry_forward:
  - 3e.2.b/c — wire bgworker `execute_target='mcp_proxy'` + bridge daemon (LISTEN/NOTIFY) so substrate-internal agents can actually call cached MCP tools
  - 3e.2.d — auto-promote `mcp_tool_cache` rows into `stewards.tool_defs` with deny-by-default `agent_tool_perms`
  - search-mcp protocol regression — DuckDuckGo MCP server returned "invalid request" on tools/list; needs investigation (possibly older protocol version or method name mismatch)
  - StreamableClientTransport lacks Headers field — first server that needs bearer-token auth will require a custom http.Client RoundTripper
  - bridge currently runs on Windows host pointing at .exe paths; long-run plan is Linux-in-Docker (Michael's preference)
---

# 3e.2.a v1 — MCP bridge skeleton + multi-bgworker

Built outbound MCP-client capability for the substrate. Two-step shape:
SQL schemas land + bridge daemon skeleton runs as `stewards-mcp bridge
refresh-tools`. Substrate now knows about the seven external MCP servers
(gospel-engine-v2, webster, yt, byu-citations, becoming, search,
exa-search) and can populate per-server tool catalogs on demand. The
LISTEN/NOTIFY wire that lets substrate-internal agents actually *call*
cached tools is deferred to 3e.2.b/c.

**What worked.** First-try clean build for both the Rust extension and
the Go sidecar bridge. Both transport types verified end-to-end —
CommandTransport for the six stdio servers, StreamableClientTransport
for exa-search remote HTTP. Bridge cached 40 tools across 6 servers in
one pass. Multi-bgworker registration (4 workers, leader-gated reaper +
scheduler) worked first time on smoke and live.

**What surprised.** search-mcp — our own DuckDuckGo MCP server —
returned "invalid request" on `tools/list`. The bridge handled the
failure gracefully (recorded `last_error`, kept refreshing the rest)
but it's a real protocol regression worth chasing. Lowest-priority of
the seven and out of scope for tonight, so just logged.

The pgrx-rust skill paid off again. `extension_sql_file!` foldback was
purely mechanical: one line in lib.rs, one entry in the Dockerfile
COPY. No path-relative gotchas because we kept the macro in lib.rs.

**Concurrency model verified.** Container restart logs show 4
bgworkers spawn, exactly one becomes leader (worker 0), all 4 inherit
the provider registry from the postmaster. The
`STEWARDS_DISPATCHER_WORKERS` env var is the knob if we want to tune
later. With FOR UPDATE SKIP LOCKED in the claim query, the queue
draining is now genuinely concurrent — the bottleneck moves from
"single tick loop" to "single chat HTTP call per worker."

**Stewardship.** Decided to ship 3e.2.a as defined and *not* fix
search-mcp tonight even though I touched the bridge that surfaced it.
Boundary-test failed: search-mcp behavior is a separate concern,
fixing it would touch a different binary in a different language, and
Michael had already scoped tonight as "schema + bridge skeleton." Right
call to surface, not act.

**Time.** Hour-ish for the SQL schema + Dockerfile + lib.rs fold +
multi-bgworker rewrite + bridge.go (~350 lines) + smoke + live verify.
Cheap because the Phase 3e.1 pattern handed me the SDK conventions and
the pgrx-rust skill cleared the foldback path.
