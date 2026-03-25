# ibeco.me — Becoming

Personal discipleship tracking app. Go backend with embedded Vue 3 frontend, PostgreSQL database, JWT auth, and an MCP server for AI assistant integration.

Deployed at [ibeco.me](https://ibeco.me) via Dokploy.

## Architecture

- **Backend:** Go + Chi router + PostgreSQL (pgx) + Goose migrations
- **Frontend:** Vue 3 + TypeScript + Vite + Tailwind (embedded in Go binary)
- **MCP Server:** Separate binary that proxies the Becoming API to AI assistants

## Build

```bash
# Web server
cd cmd/server
go build -o becoming.exe .

# MCP server
cd cmd/mcp
go build -o becoming-mcp.exe .
```

## Entry Points

| Binary | Purpose |
|--------|---------|
| `cmd/server/` | Web server — REST API + embedded SPA |
| `cmd/mcp/` | MCP stdio server for AI assistants |

## MCP Tools

The MCP server exposes tools for tracking personal transformation:

| Tool | Description |
|------|-------------|
| `create_task` | Create a new task |
| `list_tasks` | List tasks |
| `update_task` | Update a task |
| `create_practice` | Create a practice to track |
| `log_practice` | Log a practice entry |
| `list_practices` | List practices |
| `get_today` | Get today's practices and status |
| `get_report` | Get a practice report |
| `get_reflection` | Get a reflection |
| `upsert_reflection` | Create or update a reflection |

## Deployment

See `Dockerfile` and `docker-compose.yaml` for container deployment via Dokploy.

## Ecosystem

Part of the brain/becoming ecosystem:

| Component | Purpose |
|-----------|---------|
| **ibeco.me** (this) | Cloud hub — API, web UI, practices, journaling |
| **brain.exe** (`../brain/`) | Local brain — capture, classify, store, search |
| **brain-app** (`../brain-app/`) | Flutter mobile/desktop app |
