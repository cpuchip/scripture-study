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

## Future sub-phases (planned)

### 3e.1.1 — polish (small)
- Fix the jsonschema struct-tag syntax to emit proper minimum/maximum constraints
- Add `study_similar(slug, limit)` — wraps `stewards.study_similar(slug, limit)`. Mechanical.
- Add `study_citations(slug)` — wraps `stewards.study_citations(slug)`. Mechanical.

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
