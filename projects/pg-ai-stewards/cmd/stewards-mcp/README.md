# stewards-mcp

MCP (Model Context Protocol) sidecar for the pg-ai-stewards substrate. Connects to Postgres via pgxpool and exposes substrate tools to MCP clients (Claude Code, etc.) over stdio.

## Phase 3e.1 — initial tool surface

Two read-only tools over the substrate's studies corpus:

| Tool | Wraps | Purpose |
|------|-------|---------|
| `study_search` | `stewards.study_search_text(query, kinds[], limit)` | FTS over slugs+titles+bodies, returns `{slug, kind, title, snippet, rank}` per hit |
| `study_get` | `stewards.study_get(slug, include_body, line_offset, line_count, max_chars)` | Read a study by slug with line-range pagination |

Future phases (3e.2-3e.5) add stewards_brain, stewards_work_item, gospel_passthrough, and outbound MCP-client capability for consuming external MCP servers like gospel-engine-v2.

## Build

The module is registered in the workspace `go.work` at the repo root. Build with:

```bash
cd projects/pg-ai-stewards/cmd/stewards-mcp
go build -o ../../bin/stewards-mcp.exe .
```

## Configure Claude Code

Add to `.mcp.json` at the repo root (gitignored — local config only):

```json
{
  "mcpServers": {
    "pg-ai-stewards": {
      "type": "stdio",
      "command": "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/pg-ai-stewards/bin/stewards-mcp.exe",
      "env": {
        "STEWARDS_DSN": "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"
      }
    }
  }
}
```

Restart the Claude Code session — `.mcp.json` is read at session startup. After restart, `mcp__pg-ai-stewards__study_search` and `mcp__pg-ai-stewards__study_get` should appear in the deferred-tools list.

## Manual smoke test

Without restarting Claude Code, you can test the protocol by piping JSON-RPC messages through stdin:

```bash
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"smoke","version":"0.1"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"study_search","arguments":{"query":"faith hope charity","limit":3}}}'
  sleep 2
} | ./bin/stewards-mcp.exe 2>/dev/null
```

Expected: two newline-delimited JSON-RPC responses on stdout, the second containing the FTS hits in `result.structuredContent.results`.

## Troubleshooting

- **Server starts but no tools appear in Claude Code:** restart the session — `.mcp.json` is only read at startup.
- **"connection refused" in stderr logs:** verify the docker container is up (`docker ps`) and Postgres is reachable on port 55433.
- **JSON-RPC errors / corrupted output:** make sure no Go code in the project writes to stdout. The server pins logging to stderr; transitive deps doing `fmt.Println` would corrupt the protocol stream.
- **First-run approval dialog:** project-scoped MCP servers prompt for approval. Accept once to whitelist; clear with `claude mcp reset-project-choices` to test approval flow again.

## Implementation notes

- **Go module:** standalone `go.mod` (matching the stewards-cli pattern). Registered in workspace `go.work`.
- **SDK:** [`github.com/modelcontextprotocol/go-sdk` v1.6.0](https://github.com/modelcontextprotocol/go-sdk) — official Anthropic+Google SDK. Released 2026-05-08.
- **Transport:** stdio with newline-delimited JSON-RPC. SDK handles framing.
- **Logging:** stderr only. The `log.SetOutput(os.Stderr)` in main.go is critical.
- **Connection:** single pgxpool, opened at startup, closed on graceful shutdown.

See `.github/skills/mcp-server-go/SKILL.md` for protocol patterns and gotchas. See `projects/pg-ai-stewards/docs/3e-mcp-findings.md` for the build-out journal.
