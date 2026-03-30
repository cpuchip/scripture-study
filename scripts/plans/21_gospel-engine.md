# Plan 21: gospel-engine — Combined Gospel MCP Server

**Status:** READY TO BUILD
**Proposal:** [.spec/proposals/gospel-engine/main.md](../../.spec/proposals/gospel-engine/main.md) (updated Mar 29, 2026)
**Agent:** dev
**Target:** `scripts/gospel-engine/`

---

## What This Is

Merge gospel-mcp (SQLite/FTS5, 3 tools) and gospel-vec (chromem-go/embeddings, 4 tools) into one MCP server with 3 tools, TITSW enrichment, and a pre-computed graph layer.

## Production TITSW Config

| Parameter | Value |
|-----------|-------|
| Model | `ministral-3-14b-reasoning` (14B, 9.12 GB Q4_K_M) |
| Prompt | `experiments/lm-studio/scripts/prompts/titsw-calibrated.md` |
| Temperature | 0.2 |
| Context length | 65,536 tokens |
| MAE | **1.32** |
| Calibration context | `scripts/gospel-engine/context/talk-calibration.md` |

Full experiment reference: `experiments/lm-studio/scripts/references/titsw-experiment-spec.md`

## Key Architecture Decisions

- **Fresh build from source** — no migration from old databases
- **SQLite + chromem-go** — structured + vector in one process
- **3 MCP tools** — `gospel_search` (keyword/semantic/combined), `gospel_get`, `gospel_list`
- **Graph edges at index time** — cross_reference, semantic, thematic, typological
- **Talk enrichment uses calibrated prompt** — NOT gospel-vocab/titsw-framework (those are for scripture lens in Phase 3)
- **Originals stay as fallback** — gospel-mcp and gospel-vec not modified

## Phases

### Phase 1: Foundation — Scaffold + Index (1-2 sessions)
New Go module, index full corpus from source markdown, serve 3 MCP tools. No TITSW enrichment yet.

**Port from:**
- gospel-mcp (`scripts/gospel-mcp/`) → SQLite schema, FTS5, content retrieval, cross-references
- gospel-vec (`scripts/gospel-vec/`) → chromem-go storage, LLM summarization, embedding, caching

**Key files to read first:**
| File | What it provides |
|------|-----------------|
| `scripts/gospel-mcp/internal/db/schema.sql` | Full SQLite schema (scriptures, chapters, talks, manuals, FTS5) |
| `scripts/gospel-mcp/internal/mcp/server.go` | MCP protocol + tool registration pattern |
| `scripts/gospel-mcp/internal/mcp/protocol.go` | JSON-RPC types |
| `scripts/gospel-mcp/internal/tools/tools.go` | Tool response types |
| `scripts/gospel-vec/mcp.go` | MCP server + 4 tool definitions |
| `scripts/gospel-vec/search.go` | Semantic search implementation |
| `scripts/gospel-vec/index.go` | Indexing pipeline |
| `scripts/gospel-vec/summary.go` | LLM summarization |
| `scripts/gospel-vec/storage.go` | chromem-go wrapper |
| `scripts/gospel-vec/lmstudio.go` | LM Studio API client |
| `scripts/gospel-vec/cache.go` | Summary caching |
| `scripts/gospel-vec/chunking.go` | Content chunking (verse, paragraph, summary, theme) |
| `.spec/proposals/gospel-engine/main.md` | Full architecture spec (schema, tools, code structure) |

**Verification (Phase 1):**
1. `gospel-engine index` completes — full corpus indexed
2. `gospel-engine stats` shows expected counts
3. Keyword search works
4. Semantic search works
5. Edges table has cross_reference + semantic edges
6. MCP serve mode works

### Phase 2: TITSW Talk Enrichment (1-2 sessions)
Add TITSW scoring to talk indexing. Uses `titsw-calibrated.md` prompt + `talk-calibration.md` context.

### Phase 3: Scripture Enrichment (1 session)
Lens approach for scripture: gospel-vocab.md + titsw-framework.md injected into summary prompt.

### Phase 4: Combined Search (1 session)
Hybrid keyword+semantic reranking. Manual enrichment.

### Phase 5: Full Batch Reindex + Cutover (1-2 sessions)
Run across all 5,500+ talks. Swap MCP config.

## Dependencies

```
github.com/mattn/go-sqlite3       # SQLite (from gospel-mcp, Go 1.23)
github.com/philippgille/chromem-go # Vector DB (from gospel-vec, Go 1.25.6)
```

## Concurrency Notes

- 2× RTX 4090 (48GB total) — can run 2 model instances
- `--concurrency` flag: 1 (default) to 4 workers
- Sequential: ~18-20s per talk, ~28h for 5,500 talks
- 2× concurrent: ~15h
- 4× (with remote): ~8h
