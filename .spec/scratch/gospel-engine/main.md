# Scratch: gospel-engine — Combined Gospel Tool

*Created: 2026-03-29*
*Proposal: [.spec/proposals/gospel-engine/main.md](../../proposals/gospel-engine/main.md)*

---

## Binding Problem

gospel-mcp and gospel-vec are complementary tools split across two processes, two databases, and two MCP server registrations. One does full-text search (SQLite/FTS5), the other does semantic search (chromem-go/embeddings). Neither knows about the other's data. The enriched indexer pipeline needs both — TITSW metadata generated during vector indexing, structured search over that metadata in SQLite. Keeping them separate means import pipelines, cache coupling, and two tools that should be one.

Michael's direction: build a new combined tool that replaces both. Keep the originals available during the transition/reindexing, but the new tool is the replacement — it does everything both do, plus the enriched pipeline.

---

## Inventory: What the Combined Tool Must Do

### From gospel-mcp (structured search)

| Capability | Detail |
|-----------|--------|
| SQLite schema | scriptures (41,995 verses), chapters (1,584), talks (496+), manuals (20,710+), books, cross_references (1.5M+), index_metadata |
| FTS5 search | Boolean, phrase, prefix, field filters across all content types |
| Content retrieval | `gospel_get` — verse ranges, chapter, talk by reference or file path |
| Content browsing | `gospel_list` — hierarchical browsing of all content |
| Incremental indexing | Track file mtime/size, only reindex changed files |
| Cross-references | Footnote/TG/BD parsing, bidirectional lookup |
| No external dependencies | Pure SQLite, no LLM, no network at query time |

### From gospel-vec (semantic search)

| Capability | Detail |
|-----------|--------|
| Vector embeddings | chromem-go with OpenAI-compatible embedding API (LM Studio) |
| 4 index layers | verse, paragraph (4-verse chunks), summary (LLM), theme (LLM) |
| Semantic search | Similarity search across all layers and sources |
| Talk retrieval | `get_talk` — full talk text by speaker/year/month |
| Talk search | `search_talks` — filtered semantic search by speaker, year range |
| LLM summaries | 50-75 word summaries per chapter, 10-15 keywords, key verse/quote |
| LLM themes | 2-5 narrative sections detected per chapter |
| Summary caching | Per-chapter JSON files with model + prompt version validation |
| Per-source persistence | Separate .gob.gz files (scriptures, conference, manual, music) |
| LM Studio lifecycle | Auto-start server, auto-load model, health checks |
| Concurrency | Index lock prevents parallel indexing |

### New (from enriched pipeline)

| Capability | Detail |
|-----------|--------|
| TITSW teaching profiles | Dominant dimensions, teaching mode, pattern, 6 dimension scores per talk |
| Vocabulary approach (talks) | TITSW terms in system prompt, no context docs |
| Lens approach (scripture) | gospel-vocab + titsw-framework injected as context |
| Calibration context | Few-shot score anchoring for talk pipeline |
| TITSW structured search | Filter by mode, min scores, dominant dimensions |
| TITSW in FTS | mode, keywords, summary searchable via full-text |
| Combined search | "atonement talks that enact love" = semantic + structured |
| Parallelism | --concurrency flag for batch indexing (2-4 workers) |

---

## Architecture Decisions

### One binary, two databases

**SQLite** for structured data + FTS. **chromem-go** for vector embeddings. These are fundamentally different storage engines for different query types. Combining them into one application doesn't mean combining them into one database — it means one process that owns both.

No import pipeline. No cache coupling. One indexer writes to both databases in a single pass. One MCP server exposes all tools.

### Consolidated MCP tools

gospel-mcp has 3 tools, gospel-vec has 4. Some overlap. The combined tool consolidates:

| Current | Combined | Notes |
|---------|----------|-------|
| `gospel_search` (FTS) | `gospel_search` | Extended with TITSW filters + optional semantic mode |
| `search_scriptures` (semantic) | `gospel_search` | Same tool, `mode: "semantic"` parameter |
| `search_talks` (semantic, filtered) | `gospel_search` | Same tool, `source: "conference"` + semantic mode |
| `gospel_get` | `gospel_get` | Extended with TITSW metadata in response |
| `get_talk` | `gospel_get` | Folded in — `gospel_get` handles talks now |
| `gospel_list` | `gospel_list` | Unchanged |
| `list_books` | `gospel_list` | Folded in — `gospel_list` handles book listing |

Result: **3 tools** (search, get, list) that do everything 7 tools did, with cleaner semantics.

### Name question

**DECIDED: gospel-engine.** Michael chose it. The name captures both indexing and search — and it IS the engine that study.ibeco.me, brain app, and all downstream tools build on. Not pretentious when it's actually the foundation.

Options that were considered:
- `gospel-graphs` — implies the graph visualization site, not the search tool
- `gospel` — too generic, conflicts with existing MCP server name
- `gospel-search` — emphasizes search but it does indexing too
- `gospel-index` — emphasizes indexing but it does search too
- `gospel-db` — boring but accurate
- `gospel-engine` — ✅ captures both indexing and search, expandable
- `gospel-core` — the foundational tool everything else builds on

### Data directory structure

```
scripts/gospel-engine/
├── data/                          # .gitignored
│   ├── gospel.db                  # SQLite (FTS + structured)
│   ├── gospel.db-shm              # SQLite WAL
│   ├── gospel.db-wal              # SQLite WAL
│   ├── scriptures.gob.gz          # chromem-go vectors
│   ├── conference.gob.gz          # chromem-go vectors
│   ├── manual.gob.gz              # chromem-go vectors
│   ├── music.gob.gz               # chromem-go vectors
│   └── summaries/                 # LLM summary cache (JSON files)
│       ├── talk-2024-04-kearon-receive-his-gift.json
│       ├── 1-nephi-3.json
│       └── ...
├── cmd/
│   └── gospel-engine/
│       └── main.go                # Entry point
├── internal/
│   ├── db/                        # SQLite layer
│   │   ├── db.go
│   │   ├── schema.sql
│   │   └── metadata.go
│   ├── vec/                       # Vector layer
│   │   ├── store.go
│   │   ├── embedder.go
│   │   └── lmstudio.go
│   ├── indexer/                   # Unified indexer
│   │   ├── indexer.go
│   │   ├── scripture.go
│   │   ├── talk.go
│   │   ├── manual.go
│   │   ├── book.go
│   │   ├── music.go
│   │   ├── enricher.go            # TITSW enrichment
│   │   ├── chunking.go
│   │   ├── summary.go
│   │   ├── cache.go
│   │   └── crossref.go
│   ├── search/                    # Unified search
│   │   ├── structured.go          # FTS queries
│   │   ├── semantic.go            # Vector similarity
│   │   ├── combined.go            # Hybrid search
│   │   └── types.go
│   ├── tools/                     # MCP tool implementations
│   │   ├── search.go
│   │   ├── get.go
│   │   └── list.go
│   └── mcp/                       # MCP server registration
│       └── server.go
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

---

## Critical Analysis

### Is this the right thing to build?

**Yes.** The enriched pipeline requires both structured and semantic capabilities. Running two separate tools with an import pipeline between them is the complexity we're trying to eliminate. The combined tool is the natural convergence.

### What gets worse?

- **Larger binary.** Depends on both SQLite and chromem-go.
- **More complex indexer.** Writes to two databases. More modes to test.
- **Transition period.** Old gospel-mcp and gospel-vec stay active until the new tool proves itself.
- **Reindexing time.** Full reindex needs both FTS and embeddings — longer than either alone.

### What gets better?

- **One MCP config instead of two.** 3 tools instead of 7.
- **No import pipeline.** TITSW metadata written at index time, not synced between tools.
- **Combined search.** "atonement talks that enact love" in one query.
- **One build, one deploy, one update.**
- **Single source of truth.** No cache coupling, no format contracts between tools.

### Mosiah 4:27 check

Michael has enriched indexer Phase 1, teaching workstream, and gospel graph all in flight. This tool IS the enriched pipeline — it's not a new project, it's where the enriched indexer work lands. The question is sequencing: build the combined tool first, THEN run the enriched batch through it? Or build enriched indexer into gospel-vec first, then combine?

Recommendation: Build the combined tool as the target for enriched indexer Phase 1. Don't build into gospel-vec and then port — build once, build right. The combined tool's indexer IS the enriched indexer.

---

## Dependency Map

- `github.com/mattn/go-sqlite3` — SQLite driver (from gospel-mcp)
- `github.com/philippgille/chromem-go` — In-memory vector DB (from gospel-vec)
- LM Studio at localhost:1234 — Embedding + chat API (from gospel-vec)
- `text-embedding-qwen3-embedding-4b` — Embedding model (from gospel-vec)
- `nemotron-3-nano` — Chat model for summaries/enrichment
