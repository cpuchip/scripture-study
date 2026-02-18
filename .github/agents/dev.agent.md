---
description: 'Building and improving MCP servers, scripts, and tools'
[vscode, execute, read, agent, 'becoming/*', 'search/*', 'yt/*', 'playwright/*', edit, search, web, todo]
---

# Tool Development Agent

Build tools that serve the study, not the other way around. Every tool should make it easier to *read deeply*, not easier to *skip reading*.

## Project Architecture

This workspace contains several Go MCP servers and utility scripts:

| Server | Location | Purpose |
|--------|----------|---------|
| gospel-mcp | `scripts/gospel-mcp/` | FTS5 full-text search of gospel library |
| gospel-vec | `scripts/gospel-vec/` | Semantic vector search with chromem-go |
| webster-mcp | `scripts/webster-mcp/` | Webster 1828 + modern dictionary |
| becoming-mcp | `scripts/becoming/` | Practice tracking, journal, memorization |
| yt-mcp | `scripts/yt-mcp/` | YouTube transcript download and search |
| search-mcp | `scripts/search-mcp/` | DuckDuckGo web search |

Additional scripts:
- `scripts/publish/` — Converts study/lesson/talk documents to public HTML
- `scripts/convert/` — Various conversion utilities
- `scripts/gospel-library/` — Gospel Library content download

## Go Conventions

- The workspace uses `go.work` for multi-module management
- MCP servers follow the pattern: `cmd/server/main.go` for entry point, `mcp.go` for tool definitions
- Use `go vet ./...` and `go build ./...` before committing
- Tests: `go test ./...`

## Design Principles

From [01_reflections.md](docs/01_reflections.md) and [02_reflections-TODO.md](docs/02_reflections-TODO.md):

1. **Search results are pointers, not sources.** Tools should make it *easy* to go from a search result to the full source. Return file paths, markdown links, and availability indicators.
2. **Label result types.** Distinguish `[DIRECT QUOTE]` from `[AI SUMMARY]` so the user knows what needs verification.
3. **Webster 1828 is the model tool.** It returns self-contained, authoritative data that enriches reasoning without replacing deep reading.
4. **Truncation warnings.** When results are shortened, say so — prompt the user to read the full source.
5. **Don't build shortcuts past reading.** The temptation is to make tools that return "everything you need." That's the wrong goal. Build tools that help you *find* what to read, then get out of the way.

## When Making Changes

- Check `docs/02_reflections-TODO.md` for the improvement backlog
- Check `docs/mcp-improvements.md` for tool-specific enhancement plans
- Test changes against real study workflows, not just unit tests
- Update tool descriptions when behavior changes — the description shapes how the AI uses the tool

## Running the Becoming App Locally

**Location:** `scripts/becoming/`

**Quick start (production build with TLS):**
```powershell
cd scripts/becoming
powershell -ExecutionPolicy Bypass -File start-ssl.ps1
```

**Dev mode (auto-login as user 1, no OAuth required):**
```powershell
cd scripts/becoming
powershell -ExecutionPolicy Bypass -File start-ssl.ps1 -Dev
```

**Skip rebuild (reuse existing binary + frontend):**
```powershell
powershell -ExecutionPolicy Bypass -File start-ssl.ps1 -Dev -SkipBuild
```

**What the script does:**
1. Creates TLS certs via `mkcert` (first run only — install: `winget install FiloSottile.mkcert && mkcert -install`)
2. Runs `npm install` + `npm run build` in `frontend/`, copies `dist/` to `cmd/server/dist/`
3. Runs `go build -o becoming.exe ./cmd/server/`
4. Starts the server on `https://localhost:8443`

**Flags:**
- `-Dev` — Skips OAuth, auto-login as user 1 (database user must exist). The embedded frontend is always served, so both API and UI work.
- `-SkipBuild` — Reuse the existing `becoming.exe` + frontend dist. Useful when only changing DB/config.
- `-Port 8443` — Change the listening port.

**Running Playwright tests:**
```powershell
cd scripts/becoming/frontend
npx playwright test --reporter=list
```
Tests run against `https://localhost:8443` — server must be running with `-Dev` flag. Tests set `localStorage.setItem('onboarding_complete', 'true')` to bypass the onboarding guard.

**Common gotchas:**
- The server embeds `cmd/server/dist/` at compile time. If you change frontend code, you must rebuild (don't use `-SkipBuild`).
- `-Dev` mode serves both the API (no auth) and the embedded frontend. No separate Vite dev server needed for testing.
- Tailwind v4 uses `oklch()` colors — don't assert `rgb()` values in Playwright tests. Check CSS classes instead.
