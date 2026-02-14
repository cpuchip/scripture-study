---
description: 'Building and improving MCP servers, scripts, and tools'
tools:
  - search
  - editFiles
  - codebase
  - terminalLastCommand
  - runInTerminal
  - fetch
  - problems
  - testFailure
  - usages
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
