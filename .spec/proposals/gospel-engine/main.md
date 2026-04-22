---
workstream: WS3
status: building
brain_project: 3
created: 2026-03-15
last_updated: 2026-04-21
phase_status: "v1 SHIPPED, v1.5 ergonomics next, v2 hosted SHIPPED, v3 graph deferred"
---

# gospel-engine — Combined Gospel Search & Indexing Tool

> **Phase Map (added 2026-04-21):**
>
> | Phase | Scope | Status |
> |-------|-------|--------|
> | **v1 — local** | Combined SQLite/FTS5 + chromem-go vector backend, MCP server over stdio. Replaces gospel-mcp + gospel-vec locally. | **SHIPPED** (`scripts/gospel-engine/`) |
> | **v1.5 — ergonomics** | Quality-of-life improvements on the v1 local engine | OPEN — see [phase1.5-ergonomics.md](phase1.5-ergonomics.md) |
> | **v2 — hosted** | Same capabilities deployed at `engine.ibeco.me` (Postgres + pgvector). Thin MCP client (`gospel-mcp.exe`) talks to the hosted API with bearer-token auth. | **SHIPPED** Apr 20 — see [v2-hosted.md](v2-hosted.md) |
> | **v3 — graph** | Add Apache AGE on top of the v2 backend for cross-reference / citation graph traversal. | DEFERRED — see [../gospel-graph/main.md](../gospel-graph/main.md) |
>
> The body below is the original v1 spec. v2 and v3 each have their own files.

---

**Binding problem:** gospel-mcp and gospel-vec are complementary tools split across two processes, two databases, and two MCP server entries. Full-text search lives in one, semantic search in the other, and neither knows about the other's data. The enriched indexer pipeline needs both — TITSW metadata generated during vector indexing, structured search over that metadata in SQLite. Keeping them separate means import pipelines, cache format coupling, and two tools that should be one.

**Created:** 2026-03-29
**Scratch:** [.spec/scratch/gospel-engine/main.md](../../scratch/gospel-engine/main.md)
**Replaces:** gospel-mcp (`scripts/gospel-mcp/`) + gospel-vec (`scripts/gospel-vec/`)
**Absorbs:** [enriched-indexer.md](../enriched-indexer.md) Phases 1-3, [enriched-search.md](../enriched-search.md) schema + tool enhancements
**Status:** v1 SHIPPED — see Phase Map above

---

## 1. Problem Statement

Today, studying scripture with AI search requires two separate MCP servers:

- **gospel-mcp** — SQLite/FTS5. 3 tools (`gospel_search`, `gospel_get`, `gospel_list`). Fast keyword search, content retrieval, cross-reference lookup. No LLM, no embeddings. 41,995 verses, 1,584 chapters, 496+ talks, 20,710+ manual sections, 1.5M+ cross-references.
- **gospel-vec** — chromem-go vector DB. 4 tools (`search_scriptures`, `search_talks`, `list_books`, `get_talk`). Semantic similarity search across 4 layers (verse, paragraph, summary, theme). LLM-generated summaries and themes. Embedding-based retrieval.

7 MCP tools across 2 servers. No way to combine a keyword query with semantic similarity. No TITSW teaching profiles. No structured filtering on teaching modes or dimension scores. The enriched pipeline (TITSW-aware summaries) needs to write to both SQLite and the vector DB — but the two tools don't share state.

**Who's affected:** Michael doing study and lesson prep. Every MCP consumer (brain app, study.ibeco.me, any future tool). The agent context window — 7 tools is a lot of tool surface to reason about.

**How would we know it's fixed:** One MCP server, 3 tools. `gospel_search` can do keyword search, semantic search, or both. Talks have TITSW teaching profiles. You can ask "find talks that enact love and mention the Atonement" and get a combined result. One `index` command builds everything. Cross-reference and thematic links computed at index time, not at query time.

---

## 2. Success Criteria

1. **3 MCP tools** replace the current 7 — `gospel_search`, `gospel_get`, `gospel_list`
2. **All existing capabilities preserved** — every query that worked on gospel-mcp or gospel-vec still works
3. **Combined search** — `gospel_search` supports `mode: "keyword"`, `mode: "semantic"`, or `mode: "combined"` (keyword + semantic reranking)
4. **TITSW teaching profiles** on all conference talks — dominant dimensions, mode, pattern, 6 dimension scores
5. **TITSW structured filters** — filter by mode, min scores, dominant dimensions
6. **One index command** writes to both SQLite and vector DB in a single pass
7. **Enriched summaries** — talks get calibrated prompt approach (titsw-calibrated.md); scripture gets lens approach (gospel-vocab + titsw-framework context)
8. **Parallelism** — `--concurrency` flag for batch indexing (1-4 workers)
9. **Data directory gitignored** — no gigabyte data in the repo
10. **Drop-in MCP replacement** — swap gospel + gospel-vec server configs for one gospel-engine config
11. **Graph layer at index time** — cross-reference links, thematic connections, and relational edges computed during indexing, stored in SQLite, available instantly at query time

---

## 3. Constraints & Boundaries

### In scope
- New Go module at `scripts/gospel-engine/`
- SQLite + FTS5 for structured data (from gospel-mcp)
- chromem-go for vector embeddings (from gospel-vec)
- LM Studio integration for summaries, themes, and TITSW enrichment (from gospel-vec + enriched indexer)
- All existing content types: scriptures, talks, manuals, books, music, cross-references
- All existing indexing layers: verse, paragraph, summary, theme
- TITSW enrichment pipeline (Phases 1-3 from enriched-indexer proposal)
- Incremental indexing support
- Summary/theme caching with prompt version validation

### NOT in scope
- Modifying gospel-mcp or gospel-vec — they stay as-is for fallback
- study.ibeco.me integration (separate proposal, downstream consumer)
- brain app integration (downstream consumer)
- New content sources beyond what the two tools already index
- Auth, multi-user, or deployment — this is a local CLI/MCP tool

### Conventions
- Same Go patterns as gospel-mcp and gospel-vec (chi-less — no HTTP server needed, MCP is stdio)
- `github.com/mattn/go-sqlite3` for SQLite
- `github.com/philippgille/chromem-go` for vector
- LM Studio at localhost:1234 (OpenAI-compatible API)
- Environment variable configuration with sensible defaults

---

## 4. Prior Art

| Source | What it contributes |
|--------|--------------------|
| [gospel-mcp](../../../scripts/gospel-mcp/) | SQLite schema, FTS5 search, cross-reference parsing, content retrieval, list browsing. 3 MCP tools. |
| [gospel-vec](../../../scripts/gospel-vec/) | chromem-go storage, 4-layer indexing, LLM summarization, semantic search, LM Studio lifecycle. 4 MCP tools. |
| [enriched-indexer.md](../enriched-indexer.md) | TITSW vocabulary approach for talks, lens approach for scripture, calibration context. Phase 0 experiments on nemotron (T4 calibration context, MAE=1.83). |
| [enriched-search.md](../enriched-search.md) | Schema design for TITSW columns, FTS enhancement, search filter params, get response format. Now superseded architecturally but schema designs are valid. |
| [Phase 0 analysis](../../../experiments/lm-studio/scripts/results/phase0-analysis.md) | 18 experiments confirming calibration context works, gospel-vocab causes inflation on talks, love/spirit inflate inherently. |
| [talk-calibration.md](../../../scripts/gospel-engine/context/talk-calibration.md) | Refined calibration context: 2 examples (Bednar doctrine-dominant + Holland spirit-dominant) with anti-inflation guidance. Replaces original single-Kearon example that caused same-speaker anchoring. |
| [titsw-experiment-spec.md](../../../experiments/lm-studio/scripts/references/titsw-experiment-spec.md) | Comprehensive experiment reference. **Production config: ministral-3-14b-reasoning + titsw-calibrated.md + T=0.2, MAE=1.32.** All models tested, prompt evolution, known limitations, batch timing estimates. |
| [titsw-calibrated.md](../../../experiments/lm-studio/scripts/prompts/titsw-calibrated.md) | Production TITSW scoring prompt — per-dimension 4-level anchor tables (3/5/7/9), spirit ABOUT/BY distinction, scoring distribution warning. Used for Phase 2 talk enrichment. |

---

## 5. Architecture

### One binary, two storage engines

```
gospel-engine
├── SQLite (gospel.db)          ← structured data + FTS5
│   ├── scriptures              (41,995 verses)
│   ├── chapters                (1,584 chapters)
│   ├── talks                   (5,500+ talks) + TITSW columns
│   ├── manuals                 (20,710+ sections)
│   ├── books                   (additional texts)
│   ├── cross_references        (1.5M+ links)
│   └── FTS5 indexes            (keyword + TITSW vocabulary search)
│
├── chromem-go (*.gob.gz)       ← vector embeddings
│   ├── scriptures_{layer}      (verse, paragraph, summary, theme)
│   ├── conference_{layer}      (paragraph, summary)
│   ├── manual_{layer}          (paragraph, summary)
│   └── music_{layer}           (paragraph, summary)
│
└── summaries/ (JSON cache)     ← LLM output cache
    ├── talk-{year}-{month}-{filename}.json
    └── {book}-{chapter}.json
```

One indexer writes to both databases in a single pass. One search engine queries both. No import pipeline.

### Commands

```
gospel-engine index [flags]           # Index all content (scriptures, talks, manuals, books, music)
gospel-engine index --source talks    # Index only conference talks
gospel-engine index --source scriptures --volumes bofm,nt
gospel-engine index --concurrency 2   # Parallel LLM requests
gospel-engine index --incremental     # Only reindex changed files
gospel-engine index --force           # Full reindex, ignore cache

gospel-engine serve                   # MCP server on stdio
gospel-engine search "atonement"      # CLI search for testing
gospel-engine stats                   # Database statistics
gospel-engine version                 # Version info
```

### Unified indexer flow

For each content file (scripture chapter, conference talk, manual section):

```
1. Parse markdown → structured data (title, speaker, content, metadata)
2. Write to SQLite (verses/talks/manuals table + FTS)
3. Parse cross-references → write to cross_references table
4. Chunk content by requested layers (verse, paragraph)
5. Generate LLM summary + themes (with cache)
6. For talks: generate TITSW teaching profile (calibrated prompt + talk-calibration.md context)
7. For scripture: generate enriched summary (lens approach with gospel-vocab + titsw-framework)
8. Write TITSW columns to SQLite talks table
9. Build graph edges — cross-reference links, thematic connections, related content
10. Write all chunks + metadata to chromem-go
11. Update index_metadata for incremental tracking
```

Steps 2-3 (SQLite structured) and 4-10 (LLM + vector + graph) happen in the same pass. Graph edges are computed at index time so queries don't pay that cost.

### MCP tools (3 tools, replacing 7)

#### `gospel_search`

Combined keyword + semantic search. Replaces: `gospel_search`, `search_scriptures`, `search_talks`.

```json
{
  "name": "gospel_search",
  "description": "Search across all gospel content — keyword, semantic, or combined. Supports scripture, conference talks, manuals, books.",
  "parameters": {
    "query": { "type": "string", "required": true, "description": "Search query — keywords, phrases, or natural language" },
    "mode": { "type": "string", "enum": ["keyword", "semantic", "combined"], "default": "keyword", "description": "keyword = FTS5 boolean/phrase. semantic = embedding similarity. combined = keyword results reranked by semantic similarity." },
    "source": { "type": "string", "enum": ["scriptures", "conference", "manual", "magazine", "music", "books", "all"], "default": "all" },
    "path": { "type": "string", "description": "Narrow to path pattern (bofm, 2024/10, come-follow-me-*)" },
    "layers": { "type": "array", "items": { "type": "string", "enum": ["verse", "paragraph", "summary", "theme"] }, "description": "Semantic search layers (semantic/combined mode only)" },
    "limit": { "type": "integer", "default": 20, "max": 100 },
    "context": { "type": "integer", "default": 3, "description": "Context lines around keyword matches" },
    "include_content": { "type": "boolean", "default": false, "description": "Return full content vs excerpts" },
    "speaker": { "type": "string", "description": "Filter talks by speaker last name" },
    "year_from": { "type": "integer", "description": "Start year (inclusive)" },
    "year_to": { "type": "integer", "description": "End year (inclusive)" },
    "titsw_mode": { "type": "string", "enum": ["enacted", "declared", "doctrinal", "experiential"], "description": "Filter by TITSW teaching mode" },
    "titsw_dominant": { "type": "string", "description": "Filter by dominant dimension (teach_about_christ, love, etc.)" },
    "titsw_min_teach": { "type": "integer", "min": 0, "max": 9 },
    "titsw_min_help": { "type": "integer", "min": 0, "max": 9 },
    "titsw_min_love": { "type": "integer", "min": 0, "max": 9 },
    "titsw_min_spirit": { "type": "integer", "min": 0, "max": 9 },
    "titsw_min_doctrine": { "type": "integer", "min": 0, "max": 9 },
    "titsw_min_invite": { "type": "integer", "min": 0, "max": 9 }
  }
}
```

**Combined mode** mechanics: run FTS query to get candidate set → compute embedding similarity for each result → rerank by weighted combination of FTS relevance + semantic similarity. This is the query that neither tool can do alone.

#### `gospel_get`

Content retrieval with TITSW metadata. Replaces: `gospel_get`, `get_talk`.

```json
{
  "name": "gospel_get",
  "description": "Retrieve specific content by scripture reference, file path, or talk metadata. Returns full content with cross-references and TITSW teaching profile (for talks).",
  "parameters": {
    "reference": { "type": "string", "description": "Scripture reference (1 Nephi 3:7, D&C 93:36) or topic guide/dictionary entry" },
    "path": { "type": "string", "description": "Direct file path" },
    "speaker": { "type": "string", "description": "Talk speaker last name" },
    "year": { "type": "integer", "description": "Conference year" },
    "month": { "type": "string", "description": "Conference month (04 or 10)" },
    "context": { "type": "integer", "default": 0, "description": "Additional verses/paragraphs of context" },
    "include_chapter": { "type": "boolean", "default": false }
  }
}
```

Response includes TITSW fields for talks:
```json
{
  "reference": "Patrick Kearon, April 2024",
  "title": "Welcome to the Church of Joy",
  "content": "...",
  "titsw": {
    "dominant": ["help_come_to_christ", "love"],
    "mode": "enacted",
    "pattern": "invitation→doctrine→testimony",
    "scores": { "teach": 5, "help": 7, "love": 7, "spirit": 5, "doctrine": 4, "invite": 7 },
    "summary": "Elder Kearon invites members to rediscover joy...",
    "key_quote": "Welcome to the church of joy!",
    "keywords": ["joy", "welcome", "belonging", "sacrament", "worship", "testimony"]
  },
  "cross_references": [...],
  "file_path": "...",
  "markdown_link": "..."
}
```

#### `gospel_list`

Content browsing. Replaces: `gospel_list`, `list_books`.

```json
{
  "name": "gospel_list",
  "description": "Browse available gospel content — scripture volumes, conference years, manual collections, book titles.",
  "parameters": {
    "source": { "type": "string", "enum": ["scriptures", "conference", "manual", "magazine", "music", "books", "all"], "default": "all" },
    "path": { "type": "string", "description": "Path to browse (bofm, 2025/04, come-follow-me-*)" },
    "depth": { "type": "integer", "default": 1 },
    "volume": { "type": "string", "description": "Filter books by volume (bofm, ot, nt, dc, pgp)" }
  }
}
```

### SQLite schema (extended)

The existing gospel-mcp schema, plus TITSW columns on the talks table:

```sql
-- All existing tables preserved: scriptures, chapters, talks, manuals, books,
-- cross_references, index_metadata, plus FTS5 virtual tables.

-- TITSW extension on talks table:
ALTER TABLE talks ADD COLUMN titsw_dominant TEXT;       -- "teach_about_christ,invite"
ALTER TABLE talks ADD COLUMN titsw_mode TEXT;           -- "enacted"
ALTER TABLE talks ADD COLUMN titsw_pattern TEXT;        -- "story→doctrine→invitation"
ALTER TABLE talks ADD COLUMN titsw_teach INTEGER;       -- 0-9
ALTER TABLE talks ADD COLUMN titsw_help INTEGER;        -- 0-9
ALTER TABLE talks ADD COLUMN titsw_love INTEGER;        -- 0-9
ALTER TABLE talks ADD COLUMN titsw_spirit INTEGER;      -- 0-9
ALTER TABLE talks ADD COLUMN titsw_doctrine INTEGER;    -- 0-9
ALTER TABLE talks ADD COLUMN titsw_invite INTEGER;      -- 0-9
ALTER TABLE talks ADD COLUMN titsw_summary TEXT;        -- enriched LLM summary
ALTER TABLE talks ADD COLUMN titsw_key_quote TEXT;      -- key quote
ALTER TABLE talks ADD COLUMN titsw_keywords TEXT;       -- enriched keywords

-- Extended FTS for TITSW vocabulary search
CREATE VIRTUAL TABLE IF NOT EXISTS talks_fts USING fts5(
    title, speaker, content,
    titsw_dominant, titsw_mode, titsw_keywords, titsw_summary,
    content='talks', content_rowid='id'
);
```

### Code structure

```
scripts/gospel-engine/
├── cmd/gospel-engine/main.go       # Entry point, command dispatch
├── internal/
│   ├── db/                          # SQLite layer
│   │   ├── db.go                    # Open, Close, Query, Exec
│   │   ├── schema.sql               # Full schema including edges table (embedded)
│   │   └── metadata.go              # Incremental indexing metadata
│   ├── vec/                          # Vector layer
│   │   ├── store.go                  # chromem-go wrapper, per-source persistence
│   │   ├── embedder.go               # Embedding function
│   │   └── lmstudio.go              # LM Studio lifecycle management
│   ├── indexer/                      # Unified indexer (writes to both + graph)
│   │   ├── indexer.go                # Orchestrator: index all sources
│   │   ├── scripture.go              # Parse + index scripture markdown
│   │   ├── talk.go                   # Parse + index talk markdown
│   │   ├── manual.go                 # Parse + index manuals
│   │   ├── book.go                   # Parse + index books
│   │   ├── music.go                  # Parse + index music
│   │   ├── crossref.go              # Cross-reference extraction → edges
│   │   ├── graph.go                  # Semantic nearest-neighbor edges (batch)
│   │   ├── chunking.go              # Verse, paragraph, summary, theme chunking
│   │   ├── enricher.go              # TITSW enrichment (calibrated prompt for talks, lens approach for scripture)
│   │   ├── summary.go               # LLM summarization + theme detection + thematic edges
│   │   └── cache.go                  # Summary/TITSW cache (JSON files)
│   ├── search/                       # Unified search
│   │   ├── keyword.go                # FTS5 queries
│   │   ├── semantic.go               # Vector similarity queries
│   │   ├── combined.go               # Hybrid: keyword candidates + semantic rerank
│   │   └── types.go                  # SearchParams, SearchResult, etc.
│   ├── tools/                        # MCP tool implementations
│   │   ├── search.go                 # gospel_search handler
│   │   ├── get.go                    # gospel_get handler
│   │   └── list.go                   # gospel_list handler
│   └── mcp/                          # MCP server registration
│       └── server.go
├── data/                             # .gitignored — all runtime data
│   ├── gospel.db                     # SQLite (structured + FTS + edges)
│   ├── scriptures.gob.gz
│   ├── conference.gob.gz
│   ├── manual.gob.gz
│   ├── music.gob.gz
│   └── summaries/
├── context/                          # Enrichment context documents (committed)
│   ├── talk-calibration.md           # Phase 1-2: calibration examples for talk TITSW scoring
│   ├── gospel-vocab.md               # Phase 3: scripture lens approach (NOT used for talks)
│   └── titsw-framework.md            # Phase 3: scripture lens approach (NOT used for talks)
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

### Configuration

```
GOSPEL_ENGINE_DATA_DIR          # Default: ./data
GOSPEL_ENGINE_DB                # Default: ./data/gospel.db
GOSPEL_ENGINE_EMBEDDING_URL     # Default: http://localhost:1234/v1
GOSPEL_ENGINE_EMBEDDING_MODEL   # Default: text-embedding-qwen3-embedding-4b
GOSPEL_ENGINE_CHAT_URL          # Default: http://localhost:1234/v1
GOSPEL_ENGINE_CHAT_MODEL        # Default: ministral-3-14b-reasoning (production model, MAE=1.32)
GOSPEL_ENGINE_ROOT              # Default: (auto-detect workspace root)
```

### MCP server config (replaces both gospel + gospel-vec)

```json
{
  "gospel-engine": {
    "command": "C:/.../scripts/gospel-engine/gospel-engine.exe",
    "args": ["serve", "--data", "C:/.../scripts/gospel-engine/data"],
    "type": "stdio"
  }
}
```

---

## 6. Data Strategy — Fresh Build

No migration. gospel-engine builds its own databases from the source markdown files in `/gospel-library/`. The enriched pipeline generates new embeddings and summaries that wouldn't exist in the old data anyway — migrating old data just to overwrite it adds complexity for no value.

**Build-up plan:**
1. Build gospel-engine with full indexing capabilities
2. Run `gospel-engine index` — builds SQLite, embeddings, summaries, TITSW profiles, and graph edges from scratch
3. Verify: capabilities match or exceed the originals
4. Swap MCP config: remove gospel + gospel-vec, add gospel-engine
5. Keep originals in `scripts/` as fallback (not deleted)

The full index (structured + embeddings + LLM summaries + enrichment) will take time — run overnight. Incremental mode handles subsequent updates.

### Graph Layer — Building Links at Index Time

**The question:** Should relational links (cross-references, thematic connections, related talks) be computed at index time or at query time?

**Answer: Index time.** Here's why:

1. **Cross-references are already in the source.** Every scripture chapter has footnotes linking to other passages, topic guide entries, and Bible Dictionary entries. gospel-mcp already parses these into `cross_references` (1.5M+ rows). This is pure extraction — no LLM needed, no ambiguity. gospel-engine inherits this.

2. **Thematic links from embeddings come free.** During indexing, every chunk gets embedded. After embedding, computing nearest-neighbor connections between chunks is a vector similarity operation — fast, batch-friendly, and already happening implicitly when we store vectors. The question is only whether to *persist* those connections. Yes — store the top-N most similar documents for each indexed item as edges in SQLite. This turns "find related content" from a runtime vector search into a SQLite lookup.

3. **LLM-detected thematic connections belong at index time.** When the LLM generates summaries and themes for a chapter, it can also identify which other books/themes this connects to (types, prophecy-fulfillment, doctrinal parallels). This is the same LLM call — just an additional output field. No extra cost.

4. **Study time should be fast.** The whole point of pre-computing is that when you're in a study session and ask "what connects to this passage?" the answer comes back instantly from SQLite, not from a real-time embedding search + LLM synthesis.

**What we store:**

```sql
-- Graph edges table (new)
CREATE TABLE IF NOT EXISTS edges (
    id INTEGER PRIMARY KEY,
    source_type TEXT NOT NULL,       -- 'scripture', 'talk', 'manual'
    source_id TEXT NOT NULL,         -- reference or path
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    edge_type TEXT NOT NULL,         -- 'cross_reference', 'thematic', 'semantic', 'typological'
    weight REAL DEFAULT 1.0,         -- similarity score or confidence
    metadata TEXT,                   -- JSON: { "reason": "both discuss faith unto repentance" }
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX idx_edges_source ON edges(source_type, source_id);
CREATE INDEX idx_edges_target ON edges(target_type, target_id);
CREATE INDEX idx_edges_type ON edges(edge_type);
```

**Edge types built at index time:**

| Edge Type | Source | Cost | When |
|-----------|--------|------|------|
| `cross_reference` | Footnote parsing (gospel-mcp logic) | Zero LLM cost | Phase 1 |
| `semantic` | Top-N nearest neighbors from embeddings | Vector math, no LLM | Phase 1 |
| `thematic` | LLM summary output: "this chapter connects to..." | Same LLM call as summary | Phase 2-3 |
| `typological` | Enriched scripture lens: Christ-type identification | Same LLM call as enrichment | Phase 3 |

**What this means for study.ibeco.me:** The graph visualization site reads pre-computed edges from gospel-engine's SQLite. No real-time computation at query time — just SELECT and render. This is the reason to build the graph at index time: study.ibeco.me becomes a thin reader, not a compute engine.

**Do we need to worry about this now?** The cross-reference edges (Phase 1) and semantic edges (Phase 1) are essentially free — we're already doing the work. Thematic and typological edges ride on LLM calls we're already making. The `edges` table just needs to exist in the schema. The indexer writes to it as a side effect of work it's already doing. This isn't a separate feature — it's capturing connections the engine is already computing.

---

## 7. Phased Delivery

### Phase 1: Foundation — Scaffold + Index (1-2 sessions)

**Scope:** New Go module that compiles, indexes the full corpus from source, and serves all 3 MCP tools. No enrichment yet — just proving the combined architecture works and building up data from scratch.

| Deliverable | Detail |
|-------------|--------|
| `scripts/gospel-engine/` | New Go module with both dependencies |
| `.gitignore` | Data directory excluded |
| `cmd/gospel-engine/main.go` | Command dispatch: index, serve, stats, version |
| `internal/db/` | SQLite layer — schema with edges table, open/close/query |
| `internal/vec/` | chromem-go layer — store, load/save/search |
| `internal/indexer/` | Scripture + talk + manual + book + music indexing, writing to both SQLite and chromem-go |
| `internal/indexer/summary.go` | LLM summarization + theme detection (ported from gospel-vec) |
| `internal/indexer/cache.go` | Summary cache (ported from gospel-vec) |
| `internal/indexer/crossref.go` | Cross-reference extraction → edges table |
| `internal/indexer/graph.go` | Semantic nearest-neighbor edges (post-embedding batch) |
| `internal/vec/lmstudio.go` | LM Studio lifecycle (ported from gospel-vec) |
| `internal/search/` | Keyword search (FTS5) + semantic search (chromem-go) — separate, not yet combined |
| `internal/tools/` + `internal/mcp/` | 3 MCP tools: gospel_search (keyword + semantic via mode), gospel_get, gospel_list |
| `context/` | talk-calibration.md committed (Phase 1). gospel-vocab.md and titsw-framework.md are Phase 3 (scripture lens — NOT used for talks). |

**Verification:**
1. `gospel-engine index` completes — full corpus indexed from source markdown
2. `gospel-engine stats` shows expected counts (scriptures ~42K, talks ~5.5K, etc.)
3. `gospel-engine search "faith in christ"` returns keyword results
4. `gospel-engine search --mode semantic "faith in christ"` returns semantic results
5. `edges` table populated with cross_reference + semantic edges
6. MCP serve mode: all 3 tools respond correctly

**Stands alone:** Yes. This is a working replacement for both tools, with graph edges as a bonus, without enrichment.

### Phase 2: TITSW Talk Enrichment (1-2 sessions)

**Scope:** Add TITSW enrichment to the talk indexing pipeline. This IS enriched-indexer Phase 1, built directly into gospel-engine.

| Deliverable | Detail |
|-------------|--------|
| `internal/indexer/enricher.go` | TITSW teaching profile extraction — calibrated prompt approach ([titsw-calibrated.md](../../../experiments/lm-studio/scripts/prompts/titsw-calibrated.md)) + [talk-calibration.md](../../../scripts/gospel-engine/context/talk-calibration.md) context |
| Enriched talk prompt | TEACHING_PROFILE output format in system prompt |
| TITSW columns in SQLite | titsw_dominant, titsw_mode, etc. populated during index |
| TITSW metadata in chromem-go | Flat `map[string]string` fields in DocMetadata |
| TITSW search filters | `gospel_search` accepts titsw_mode, titsw_dominant, titsw_min_* params |
| TITSW in `gospel_get` | Talk responses include teaching profile |
| `--concurrency` flag | Worker pool for parallel LLM requests (default 1, max 4) |
| FTS5 extended | talks_fts includes titsw_dominant, titsw_mode, titsw_keywords, titsw_summary |
| Thematic edges | LLM summary call includes "related talks/themes" output → edges table |

**Verification:**
1. Index 5 ground-truth talks — scores within ±2 of targets
2. `gospel_search --source conference --titsw_mode enacted` returns only enacted talks
3. `gospel_search --source conference --titsw_min_teach 7` returns high-teach talks
4. `gospel_get` for a talk includes TITSW profile
5. FTS: `gospel_search "enacted love"` matches via FTS on titsw columns
6. Thematic edge rows exist linking related talks

### Phase 3: Scripture Enrichment (1 session)

**Scope:** Enriched-indexer Phase 2 — lens approach for scripture summaries.

| Deliverable | Detail |
|-------------|--------|
| Lens injection | gospel-vocab.md + titsw-framework.md injected into scripture summary prompt |
| Deeper keywords | Typological connections, Christ-type identification in keywords |
| Enriched scripture summaries | Richer summaries with theological depth |
| Typological edges | LLM identifies Christ-types, prophecy-fulfillment pairs → edges table |

**Verification:** Compare Alma 32, Zechariah 3, 1 Nephi 11 summaries before/after — typological terms present that weren't before. Typological edges link OT passages to their NT/BofM fulfillments.

### Phase 4: Combined Search + Manual Enrichment (1 session)

**Scope:** The hybrid search that neither tool can do alone, plus enriched-indexer Phase 3.

| Deliverable | Detail |
|-------------|--------|
| `internal/search/combined.go` | Keyword candidate set → semantic reranking |
| `mode: "combined"` | Works in `gospel_search` |
| Manual enrichment | Content manuals get talk-style TITSW (meta-teaching manuals skipped) |
| Talk theme detection | Rhetorical section identification |

**Verification:** `gospel_search --mode combined "atonement" --titsw_mode enacted` returns talks ranked by both keyword relevance and semantic similarity to "atonement."

### Phase 5: Full Batch Reindex + Cutover (1-2 sessions)

**Scope:** Run the enriched pipeline across all 5,500+ talks. Deploy as replacement.

| Deliverable | Detail |
|-------------|--------|
| Full talk reindex | All conference talks with TITSW enrichment |
| Full scripture reindex | All standard works with lens-enriched summaries |
| MCP config swap | Replace gospel + gospel-vec with gospel-engine |
| Verification suite | Automated checks against ground truth |
| Complete graph | All edge types populated across full corpus |

**Verification:**
1. `gospel-engine stats` shows complete counts
2. 5 ground-truth talks within ±2 of targets
3. All existing study agent/lesson agent workflows still work
4. Old tools still compile and are available as fallback
5. Edges table has cross-reference, semantic, thematic, and typological edges

---

## 8. Verification Strategy

| Phase | Verification | Criteria |
|-------|-------------|---------|
| 1 | Index correctness | Row counts match expectations. 10 random queries produce sensible results. |
| 1 | MCP tool parity | All 7 original tool capabilities accessible through 3 new tools |
| 1 | Graph edges | cross_reference + semantic edges populated in edges table |
| 2 | TITSW scores | 5 ground-truth talks within ±2 |
| 2 | TITSW filters | Structured queries return correct subsets |
| 2 | Thematic edges | Talk-to-talk thematic connections stored |
| 3 | Enriched scripture | Typological keywords present in Alma 32, Zechariah 3 |
| 3 | Typological edges | OT→NT/BofM fulfillment links stored |
| 4 | Combined search | Hybrid results outperform keyword-only for ambiguous queries |
| 5 | Full reindex | No errors across 5,500+ talks. Stats match expectations. |
| 5 | Cutover | All agents work with new MCP config |

---

## 9. Costs & Risks

### Costs
- **Development time:** 5-8 sessions across 5 phases
- **Reindex time:** Full talk reindex ~18-20s per talk sequential on ministral-3-14b-reasoning (50-63 tok/s). ~28 hours for 5,500 talks at 1× concurrency, ~15 hours at 2× (dual 4090), ~8 hours at 4× (with remote). Scripture reindex ~4-6 hours. Run overnight.
- **Both dependencies:** Binary is larger (SQLite + chromem-go). Both are well-tested Go libraries.

### Risks

| Risk | Mitigation |
|------|-----------|
| Scope: replacing two working tools is high-stakes | Fresh index from source + verification before cutover. Originals stay as fallback. |
| Combined search complexity | Phase 4 — build after simpler modes are proven. Combined mode is additive, not required. |
| TITSW enrichment quality | Full experiment suite complete (MAE=1.32 production on ministral-3-14b-reasoning with calibrated prompt). See [titsw-experiment-spec.md](../../../experiments/lm-studio/scripts/references/titsw-experiment-spec.md). |
| Full index time | --concurrency flag. Run overnight. Incremental mode after initial build. |
| Two databases in one process | Memory: chromem-go is in-memory during search. SQLite is file-backed. Tested separately, should compose fine. |
| Graph edge explosion | Top-N cap on semantic edges (e.g., 20 per document). Thematic edges are sparse by nature. |

### What gets worse
- Larger single binary
- More complex indexer (writes to two databases + edges per content item)
- If gospel-engine has a bug, both search types are down (vs. one of two being down)

### What gets better
- One MCP config instead of two
- 3 tools instead of 7 — less agent confusion, less context window consumption
- No import pipeline — TITSW metadata generated and stored in one pass
- Combined search — the capability that neither tool offers alone
- Graph edges computed at index time — study.ibeco.me reads pre-computed connections instead of computing them
- One build, one update, one mental model

---

## 10. Creation Cycle Review

| Step | Question | This proposal |
|------|----------|--------------|
| **Intent** | Why? | Unified search + pre-computed graph is the foundation everything else builds on — study.ibeco.me, brain app, lesson prep. |
| **Covenant** | Rules? | Existing Go conventions. gospel-mcp + gospel-vec patterns. No data loss during migration. |
| **Stewardship** | Who? | dev agent builds. Michael reviews. gospel-mcp/gospel-vec maintained as fallback by existing code. |
| **Spiritual Creation** | Spec precise? | Schema, tool definitions, data flow, migration strategy all specified. Phase 1 is buildable. |
| **Line upon Line** | Phasing? | 5 phases. Phase 1 stands alone as a working replacement. Enrichment is additive. |
| **Physical Creation** | Who executes? | dev agent, one phase at a time. |
| **Review** | How to verify? | Migration row counts, ground-truth scores, MCP tool parity tests. |
| **Atonement** | If wrong? | Originals stay in scripts/. Revert MCP config. Data built from source — re-index to recover. |
| **Sabbath** | When to pause? | After Phase 1 — does the combined tool work as a replacement? After Phase 2 — are TITSW profiles good? After Phase 5 — full cutover, is the graph useful for study? |
| **Consecration** | Who benefits? | Michael directly. study.ibeco.me. brain app. Any downstream consumer. |
| **Zion** | Integration? | This IS the integration — two tools becoming one. Everything downstream gets simpler. |

---

## 11. Recommendation

**Build. This is the natural convergence of gospel-mcp and gospel-vec, the home for the enriched pipeline, and the engine that study.ibeco.me reads from.**

Phase 1 is the critical path — prove the combined architecture works, index from source, verify parity. This is mostly porting existing code into a new structure with the edges table as an addition. Phase 2 is where the new value starts — TITSW profiles on talks. Phase 5 is the commitment — swapping MCP config.

The enriched indexer proposal (Phases 1-3) folds directly into this tool's Phases 2-4. Instead of building enrichment into gospel-vec and then porting, build once into gospel-engine.

**Sequencing in the overall plan:**
1. gospel-engine Phase 1 (foundation + fresh index) — first
2. gospel-engine Phase 2 (TITSW talk enrichment) — this IS enriched-indexer Phase 1
3. gospel-engine Phases 3-4 (scripture enrichment, combined search)
4. gospel-engine Phase 5 (full reindex, cutover)
5. study.ibeco.me Phase 1 — reads pre-computed edges + search from gospel-engine

**Hand off to:** dev agent, with this proposal + [enriched-indexer.md](../enriched-indexer.md) as spec references.
