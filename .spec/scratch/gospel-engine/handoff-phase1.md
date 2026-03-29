# gospel-engine Phase 1 — Build Handoff

## What We're Building

**gospel-engine** is a new Go tool at `scripts/gospel-engine/` that replaces both `gospel-mcp` (SQLite/FTS5) and `gospel-vec` (chromem-go vector DB). One binary, two storage engines, 3 MCP tools replacing 7. The full proposal is at `.spec/proposals/gospel-engine/main.md` — read it completely before writing any code.

## Phase 1 Scope: Foundation + Fresh Index

Build the scaffold, index the full corpus from source markdown, serve all 3 MCP tools. No TITSW enrichment yet (that's Phase 2) — just prove the combined architecture works.

### What to port

**From gospel-mcp** (`scripts/gospel-mcp/`):
- `internal/db/` — SQLite schema, open/close/query patterns
- `internal/indexer/` — Scripture, talk, manual parsing + cross-reference extraction
- `internal/tools/` — gospel_search (FTS), gospel_get, gospel_list logic
- `internal/mcp/` — MCP server registration (stdio transport)
- `cmd/gospel-mcp/main.go` — Command dispatch pattern

**From gospel-vec** (`scripts/gospel-vec/`):
- `storage.go` — chromem-go initialization, collection management, persistence (.gob.gz files)
- `embed.go` — OpenAI-compatible embedding via LM Studio
- `index.go` — 4-layer indexing (verse, paragraph, summary, theme)
- `chunking.go` — Verse and paragraph chunking strategies
- `summary.go` — LLM summarization + theme detection
- `cache.go` — Per-chapter JSON summary cache with model/prompt version validation
- `lmstudio.go` — LM Studio lifecycle (auto-start, health check, model loading)
- `search.go` — Semantic similarity search across layers/collections
- `talk_parser.go` — Conference talk markdown parsing
- `manual_parser.go` — Manual section parsing
- `mcp.go` — MCP tool definitions and handlers
- `config.go` — Environment variable configuration

### New in Phase 1

- **`edges` table in SQLite** — graph edges for cross-references + semantic nearest-neighbors:
  ```sql
  CREATE TABLE IF NOT EXISTS edges (
      id INTEGER PRIMARY KEY,
      source_type TEXT NOT NULL,    -- 'scripture', 'talk', 'manual'
      source_id TEXT NOT NULL,      -- reference or path
      target_type TEXT NOT NULL,
      target_id TEXT NOT NULL,
      edge_type TEXT NOT NULL,      -- 'cross_reference', 'semantic', 'thematic', 'typological'
      weight REAL DEFAULT 1.0,
      metadata TEXT,                -- JSON
      created_at TEXT DEFAULT (datetime('now'))
  );
  ```
- **`internal/indexer/crossref.go`** — Footnote parsing → edges table (ported from gospel-mcp cross-reference logic, but writing to edges instead of cross_references)
- **`internal/indexer/graph.go`** — After embedding, compute top-N semantic nearest-neighbor edges per document
- **Combined `gospel_search`** — `mode: "keyword"` (FTS5), `mode: "semantic"` (chromem-go), not yet `mode: "combined"` (Phase 4)
- **`context/` directory** — Committed context files: `gospel-vocab.md`, `titsw-framework.md`, `talk-calibration.md` (the calibration context is already written at `scripts/gospel-engine/context/talk-calibration.md`)

### Commands

```
gospel-engine index [--source talks|scriptures|manuals|books|music|all] [--concurrency N] [--incremental] [--force]
gospel-engine serve                    # MCP on stdio
gospel-engine search "query"           # CLI test
gospel-engine stats                    # Database stats
gospel-engine version
```

### Environment variables

```
GOSPEL_ENGINE_DATA_DIR          # Default: ./data
GOSPEL_ENGINE_DB                # Default: ./data/gospel.db
GOSPEL_ENGINE_EMBEDDING_URL     # Default: http://localhost:1234/v1
GOSPEL_ENGINE_EMBEDDING_MODEL   # Default: text-embedding-qwen3-embedding-4b
GOSPEL_ENGINE_CHAT_URL          # Default: http://localhost:1234/v1
GOSPEL_ENGINE_CHAT_MODEL        # Default: (auto-detect nemotron-3-nano)
GOSPEL_ENGINE_ROOT              # Default: (auto-detect workspace root)
```

### Directory structure

```
scripts/gospel-engine/
├── cmd/gospel-engine/main.go
├── internal/
│   ├── db/           # SQLite (schema with edges table, open/close/query)
│   ├── vec/          # chromem-go (store, load/save/search, lmstudio lifecycle)
│   ├── indexer/      # Parse + index all content types
│   │   ├── scripture.go, talk.go, manual.go, book.go, music.go
│   │   ├── crossref.go    (footnote parsing → edges)
│   │   ├── graph.go       (semantic nearest-neighbor edges)
│   │   ├── chunking.go, summary.go, cache.go
│   │   └── enricher.go    (stub for Phase 2)
│   ├── search/       # keyword.go, semantic.go, types.go
│   ├── tools/        # search.go, get.go, list.go
│   └── mcp/          # server.go
├── context/          # gospel-vocab.md, titsw-framework.md, talk-calibration.md
├── data/             # .gitignored runtime data
├── go.mod
└── .gitignore
```

### MCP config (for `.vscode/mcp.json`, replaces both gospel + gospel-vec entries)

```json
{
  "servers": {
    "gospel-engine": {
      "type": "stdio",
      "command": "${workspaceFolder}/scripts/gospel-engine/gospel-engine.exe",
      "args": ["serve"]
    }
  }
}
```

## Key Reference Files

| File | What it contains |
|------|-----------------|
| `.spec/proposals/gospel-engine/main.md` | **Full proposal — READ THIS FIRST.** Architecture, tools, schema, phases, verification criteria. |
| `.spec/proposals/enriched-indexer.md` | TITSW pipeline design, Phase 0 experiment results, prompt evolution. Phase 2+ reference. |
| `.spec/scratch/gospel-engine/main.md` | Architecture decisions, inventory, name decision, calibration decision. |
| `scripts/gospel-mcp/` | SQLite/FTS5 source to port — `internal/db/`, `internal/indexer/`, `internal/tools/`. |
| `scripts/gospel-vec/` | chromem-go source to port — flat Go files at root level. |
| `scripts/gospel-engine/context/talk-calibration.md` | Refined calibration context (Bednar + Holland). For Phase 2 enricher. |
| `experiments/lm-studio/scripts/context/gospel-vocab.md` | Gospel vocabulary context (~1,960 tokens). Copy to `context/`. |
| `experiments/lm-studio/scripts/context/01-titsw-framework.md` | TITSW framework context (~1,990 tokens). Copy to `context/`. |
| `experiments/lm-studio/scripts/results/phase0-analysis.md` | Phase 0 experiment results (T4 best, MAE=1.83). Reference for Phase 2. |
| `experiments/lm-studio/scripts/prompts/titsw-enriched-talk.md` | Enriched talk prompt for Phase 2. |

## Verification Criteria (Phase 1)

1. `gospel-engine index` completes — full corpus indexed from source markdown
2. `gospel-engine stats` shows expected counts (scriptures ~42K, talks ~5.5K+, manuals ~20K+)
3. `gospel-engine search "faith in christ"` returns keyword results
4. `gospel-engine search --mode semantic "faith in christ"` returns semantic results
5. `edges` table populated with cross_reference + semantic edges
6. MCP serve mode: all 3 tools respond correctly — test with VS Code Copilot Chat

## What NOT to Do

- Do NOT add TITSW enrichment — that's Phase 2
- Do NOT build combined search mode — that's Phase 4
- Do NOT modify gospel-mcp or gospel-vec — they stay as fallback
- Do NOT migrate data — index fresh from source markdown files
- The `cross_references` table from gospel-mcp can still exist in the schema for backward compatibility, but new indexing should write cross-ref data to the `edges` table

## Module Setup

The workspace uses `go.work` at the root. Add `scripts/gospel-engine` to the workspace list. Module path: `github.com/stuffleberry/scripture-study/scripts/gospel-engine` (or match the existing pattern from gospel-mcp/gospel-vec).
