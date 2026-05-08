---
name: mcp-server-go
description: "Authoritative reference for building MCP (Model Context Protocol) servers in Go using the official Anthropic+Google SDK. Load when building or modifying an MCP server in this repo (currently `projects/pg-ai-stewards/cmd/stewards-mcp/`). Captures protocol shape, the Go SDK idioms, stdio transport gotchas, and Claude Code integration. Researched 2026-05 against the 2025-11-25 spec."
user-invokable: false
---

# mcp-server-go — building MCP servers in Go

This skill answers the questions that block productive MCP server work in Go. Verified against the 2025-11-25 spec and the official `modelcontextprotocol/go-sdk` v1.6.0 (released 2026-05-08).

## When to load

- Editing or building Go code in `projects/pg-ai-stewards/cmd/stewards-mcp/`
- Adding a new MCP tool to an existing server
- Wiring up a new MCP server to Claude Code or another MCP client
- Hitting "tool not found" / "JSON-RPC error" / silent connection failures
- Reading another project's MCP server code to understand its shape

## The four facts that change everything

### 1. Use the official Go SDK

`github.com/modelcontextprotocol/go-sdk` — maintained jointly by Anthropic and Google. v1.6.0 as of 2026-05-08. Schema reflection from struct tags, transport handling, capability negotiation, error wrapping all built in. Going from scratch in Go is ~250 lines for a tools-only stdio server; using the SDK is ~30. Use the SDK.

```go
import "github.com/modelcontextprotocol/go-sdk/mcp"
```

### 2. stdio transport is newline-delimited JSON

NOT LSP-style with `Content-Length` headers. One JSON-RPC message per line, embedded newlines forbidden inside messages. The SDK handles this for you via `mcp.StdioTransport{}`. If you ever roll your own:

- Read stdin line by line
- Write each response as one line to stdout
- **Never write to stdout from anywhere else** — that corrupts the protocol stream
- **stderr is for logs**

### 3. The `initialize` handshake is mandatory and ordered

Three steps, in order, before any tool call:
1. Client → server: `initialize` request (protocol version, capabilities)
2. Server → client: `initialize` response (protocol version, server capabilities, serverInfo)
3. Client → server: `notifications/initialized` (one-way, no response)

Only after step 3 may the client send `tools/list` or `tools/call`. The SDK enforces this. If you hand-roll, **do not** let `tools/list` succeed before `initialized` arrives — strict clients reject early-tool servers.

### 4. Two error kinds, don't conflate them

- **JSON-RPC errors** (code `-32602` etc.): protocol-level. "Unknown tool", "malformed params". Returned via the JSON-RPC `error` object.
- **Tool execution errors** (`isError: true` inside `result`): runtime. "DB connection failed", "Tool returned no results". The model SEES these and can self-correct.

Mixing them up means the model treats DB outages as unrecoverable system errors. Use `mcp.NewToolResultError(...)` for the latter.

## Minimum viable server (uses official SDK)

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool input. SDK reflects struct tags into JSON Schema 2020-12.
type SearchInput struct {
    Query string `json:"query" jsonschema:"description=topic to search"`
    Limit int    `json:"limit,omitempty" jsonschema:"description=max results,minimum=1,maximum=100"`
}

// Tool output. SDK reflects struct tags into outputSchema if you declare it.
type SearchOutput struct {
    Results []map[string]any `json:"results"`
}

// Tool handler. Three returns: optional pre-built result, structured output,
// optional error. SDK builds the standard {content, structuredContent, isError}
// shape automatically.
func studySearch(
    ctx context.Context,
    req *mcp.CallToolRequest,
    in SearchInput,
) (*mcp.CallToolResult, SearchOutput, error) {
    // ... query DB, build results ...
    return nil, SearchOutput{Results: []map[string]any{}}, nil
}

func main() {
    // CRITICAL: pin logger to stderr. Anything to stdout corrupts the protocol.
    logger := log.New(os.Stderr, "stewards-mcp: ", log.LstdFlags)

    s := mcp.NewServer(&mcp.Implementation{
        Name:    "pg-ai-stewards",
        Version: "0.1.0",
    }, nil)

    mcp.AddTool(s, &mcp.Tool{
        Name:        "study_search",
        Description: "Find studies in the substrate by topic.",
    }, studySearch)

    if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        logger.Fatalf("server.Run: %v", err)
    }
}
```

That's a fully spec-compliant tools-only MCP server. Build, register in `.mcp.json`, restart Claude Code session.

## Tool registration patterns

**Tool name regex:** `[A-Za-z0-9_.-]{1,128}`. Slashes and spaces are silently rejected by some clients. Pick a convention (`study_search`, `study.search`, or `study-search`) and stick to it.

**Input schema is auto-generated from the struct.** Use `json:"..."` tags and `jsonschema:"..."` for descriptions, ranges, enums.

**Output schema is optional.** If you declare it, you MUST return `structuredContent` matching it. Strict clients reject mismatches. Either:
- Declare and validate (best for typed APIs)
- Omit and rely on the auto-generated text block (best for free-form results)

**Tool list changes** can be announced at runtime via `notifications/tools/list_changed` if your server has dynamic tool sets. For static tool sets, declare `capabilities.tools.listChanged: false` (the SDK default — don't override unless needed).

## Claude Code integration

Three scopes for registering an MCP server:

| Scope | File | Use for |
|-------|------|---------|
| Project | `.mcp.json` (repo root) | Servers that ship with the project, work for all teammates |
| Local user | `~/.claude.json` | Personal experiments, tools |
| Global user | `~/.claude/settings.json` (rarely used) | System-wide tools |

For project sidecars (the case here), use **`.mcp.json` at repo root**:

```json
{
  "mcpServers": {
    "pg-ai-stewards": {
      "type": "stdio",
      "command": "${WORKSPACE_FOLDER}/projects/pg-ai-stewards/bin/stewards-mcp.exe",
      "args": ["--dsn", "${STEWARDS_DSN:-postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable}"],
      "env": {}
    }
  }
}
```

Env-var expansion: `${VAR}` and `${VAR:-default}`. Project-scoped servers prompt for approval on first run; clear with `claude mcp reset-project-choices` if testing a config change.

**Restart the Claude Code session** after editing `.mcp.json`. The MCP client only re-reads the config at session start. The skill listing in subsequent turns will then show the new tools.

## Gotchas (recurring failures)

1. **stderr discipline.** Any `fmt.Println` or default-logger call writes to stdout and corrupts the protocol. Pin all logging to `os.Stderr`. Audit dependencies — pgx's default logger writes to stdout in some configurations.

2. **Don't respond to notifications.** `notifications/initialized` and `notifications/cancelled` arrive without `id`. Sending a response is a protocol violation. The SDK handles this; rolled-your-own constantly breaks it.

3. **Tool name characters.** Stick to `[A-Za-z0-9_.-]`. Underscores are safest.

4. **`isError: true` vs JSON-RPC error.** Tool-execution failures use `isError: true` so the model can react. JSON-RPC errors are for protocol violations. `mcp.NewToolResultError(...)` is the SDK helper.

5. **Buffered stdout.** Go's stdout is fully buffered when not connected to a TTY (Claude Code spawns you as a pipe). The SDK flushes after each message. If you write a custom transport, wrap stdout in a `bufio.Writer` and flush every line.

6. **Initialize before tools.** Do not let `tools/list` succeed before `notifications/initialized` arrives. Strict clients enforce ordering.

7. **`outputSchema` is a contract.** If declared, validate matches before sending. Either be rigorous or omit the schema.

8. **First-run approval dialog.** Project-scoped MCP servers prompt the user for approval on first invocation. Document this for teammates so they don't think the server is broken.

9. **pgx connection lifecycle.** Use `pgxpool.New(ctx, dsn)` for connection pooling. Close the pool on `s.Run()` exit. Don't share `*pgx.Conn` across goroutines — use the pool.

10. **Context cancellation.** The SDK passes `context.Context` through to handlers. Honor it. If a tool call is mid-flight and the client cancels, abort cleanly via `ctx.Done()`.

## Real-world references

When in doubt, look at how production Go MCP servers are structured:

- **[subnetmarco/pgmcp](https://github.com/subnetmarco/pgmcp)** — Postgres MCP using the official Go SDK. Closest analog to what we're building. Has CI, Docker, schema fixtures.
- **[mark3labs/mcp-go examples/](https://github.com/mark3labs/mcp-go/tree/main/examples)** — even if you use the official SDK, the example servers are clean reference implementations of the protocol flow.
- **[modelcontextprotocol/go-sdk examples/](https://github.com/modelcontextprotocol/go-sdk/tree/main/examples)** — the official SDK's bundled examples.

## Project-specific (pg-ai-stewards)

**Binary location:** `projects/pg-ai-stewards/cmd/stewards-mcp/main.go`. Compiled to `projects/pg-ai-stewards/bin/stewards-mcp.exe` on Windows.

**Connection string:** `STEWARDS_DSN` env var, defaulting to `postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable` (the docker-compose port mapping).

**First tools to register (Phase 3e.1):**
- `study_search` — wraps `stewards.study_search_text(query, kinds, limit)`, returns matched study slugs + titles + scores
- `study_get` — wraps `stewards.study_get(slug)`, returns body + frontmatter + citations

**Future tools (Phase 3e.2-3e.5):** `gospel_passthrough`, `stewards_brain`, `stewards_work_item`, `gospel_search`, `gospel_get` (the latter two via outbound MCP-client to gospel-engine-v2).

**Findings doc:** `projects/pg-ai-stewards/docs/3e-mcp-findings.md` (created when 3e.1 ships).
