# gospel-graphs — Combined Gospel Search & Indexing Tool

**Binding problem:** gospel-mcp and gospel-vec are complementary tools split across two processes, two databases, and two MCP server entries. Full-text search lives in one, semantic search in the other, and neither knows about the other's data. The enriched indexer pipeline needs both — TITSW metadata generated during vector indexing, structured search over that metadata in SQLite. Keeping them separate means import pipelines, cache format coupling, and two tools that should be one.

**Created:** 2026-03-29
**Scratch:** [.spec/scratch/gospel-graphs/main.md](../../scratch/gospel-graphs/main.md)
**Replaces:** gospel-mcp (`scripts/gospel-mcp/`) + gospel-vec (`scripts/gospel-vec/`)
**Absorbs:** [enriched-indexer.md](../enriched-indexer.md) Phases 1-3, [enriched-search.md](../enriched-search.md) schema + tool enhancements
**Status:** Proposed

---

## 1. Problem Statement

Today, studying scripture with AI search requires two separate MCP servers:

- **gospel-mcp** — SQLite/FTS5. 3 tools (`gospel_search`, `gospel_get`, `gospel_list`). Fast keyword search, content retrieval, cross-reference lookup. No LLM, no embeddings. 41,995 verses, 1,584 chapters, 496+ talks, 20,710+ manual sections, 1.5M+ cross-references.
- **gospel-vec** — chromem-go vector DB. 4 tools (`search_scriptures`, `search_talks`, `list_books`, `get_talk`). Semantic similarity search across 4 layers (verse, paragraph, summary, theme). LLM-generated summaries and themes. Embedding-based retrieval.

7 MCP tools across 2 servers. No way to combine a keyword query with semantic similarity. No TITSW teaching profiles. No structured filtering on teaching modes or dimension scores. The enriched pipeline (TITSW-aware summaries) needs to write to both SQLite and the vector DB — but the two tools don't share state.

**Who's affected:** Michael doing study and lesson prep. Every MCP consumer (brain app, study.ibeco.me, any future tool). The agent context window — 7 tools is a lot of tool surface to reason about.

**How would we know it's fixed:** One MCP server, 3 tools. `gospel_search` can do keyword search, semantic search, or both. Talks have TITSW teaching profiles. You can ask "find talks that enact love and mention the Atonement" and get a combined result. One `index` command builds everything.

---

## 2. Success Criteria

1. **3 MCP tools** replace the current 7 — `gospel_search`, `gospel_get`, `gospel_list`
2. **All existing capabilities preserved** — every query that worked on gospel-mcp or gospel-vec still works
3. **Combined search** — `gospel_search` supports `mode: "keyword"`, `mode: "semantic"`, or `mode: "combined"` (keyword + semantic reranking)
4. **TITSW teaching profiles** on all conference talks — dominant dimensions, mode, pattern, 6 dimension scores
5. **TITSW structured filters** — filter by mode, min scores, dominant dimensions
6. **One index command** writes to both SQLite and vector DB in a single pass
7. **Enriched summaries** — talks get TITSW vocabulary approach; scripture gets lens approach (gospel-vocab + titsw-framework context)
8. **Parallelism** — `--concurrency` flag for batch indexing (1-4 workers)
9. **Data directory gitignored** — no gigabyte data in the repo
10. **Drop-in MCP replacement** — swap gospel + gospel-vec server configs for one gospel-graphs config

---

## 3. Constraints & Boundaries

### In scope
- New Go module at `scripts/gospel-graphs/`
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
| [enriched-indexer.md](../enriched-indexer.md) | TITSW vocabulary approach for talks, lens approach for scripture, calibration context, Phase 0 experiment results (T4 best, MAE=1.83). |
| [enriched-search.md](../enriched-search.md) | Schema design for TITSW columns, FTS enhancement, search filter params, get response format. Now superseded architecturally but schema designs are valid. |
| [Phase 0 analysis](../../../experiments/lm-studio/scripts/results/phase0-analysis.md) | 18 experiments confirming calibration context works, gospel-vocab causes inflation on talks, love/spirit inflate inherently. |

---

## 5. Architecture

### One binary, two storage engines

```
gospel-graphs
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
gospel-graphs index [flags]           # Index all content (scriptures, talks, manuals, books, music)
gospel-graphs index --source talks    # Index only conference talks
gospel-graphs index --source scriptures --volumes bofm,nt
gospel-graphs index --concurrency 2   # Parallel LLM requests
gospel-graphs index --incremental     # Only reindex changed files
gospel-graphs index --force           # Full reindex, ignore cache

gospel-graphs serve                   # MCP server on stdio
gospel-graphs search "atonement"      # CLI search for testing
gospel-graphs stats                   # Database statistics
gospel-graphs version                 # Version info

gospel-graphs migrate-mcp            # Import existing gospel-mcp SQLite
gospel-graphs migrate-vec            # Import existing gospel-vec data files
```

### Unified indexer flow

For each content file (scripture chapter, conference talk, manual section):

```
1. Parse markdown → structured data (title, speaker, content, metadata)
2. Write to SQLite (verses/talks/manuals table + FTS)
3. Parse cross-references → write to cross_references table
4. Chunk content by requested layers (verse, paragraph)
5. Generate LLM summary + themes (with cache)
6. For talks: generate TITSW teaching profile (vocabulary approach + calibration context)
7. For scripture: generate enriched summary (lens approach with gospel-vocab + titsw-framework)
8. Write TITSW columns to SQLite talks table
9. Write all chunks + metadata to chromem-go
10. Update index_metadata for incremental tracking
```

Steps 2-3 (SQLite) and 4-9 (vector + LLM) happen in the same pass. No second import step.

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
scripts/gospel-graphs/
├── cmd/gospel-graphs/main.go       # Entry point, command dispatch
├── internal/
│   ├── db/                          # SQLite layer
│   │   ├── db.go                    # Open, Close, Query, Exec
│   │   ├── schema.sql               # Full schema (embedded)
│   │   └── metadata.go              # Incremental indexing metadata
│   ├── vec/                          # Vector layer
│   │   ├── store.go                  # chromem-go wrapper, per-source persistence
│   │   ├── embedder.go               # Embedding function
│   │   └── lmstudio.go              # LM Studio lifecycle management
│   ├── indexer/                      # Unified indexer (writes to both)
│   │   ├── indexer.go                # Orchestrator: index all sources
│   │   ├── scripture.go              # Parse + index scripture markdown
│   │   ├── talk.go                   # Parse + index talk markdown
│   │   ├── manual.go                 # Parse + index manuals
│   │   ├── book.go                   # Parse + index books
│   │   ├── music.go                  # Parse + index music
│   │   ├── crossref.go              # Cross-reference extraction
│   │   ├── chunking.go              # Verse, paragraph, summary, theme chunking
│   │   ├── enricher.go              # TITSW enrichment (vocabulary + lens + calibration)
│   │   ├── summary.go               # LLM summarization + theme detection
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
│   ├── gospel.db
│   ├── scriptures.gob.gz
│   ├── conference.gob.gz
│   ├── manual.gob.gz
│   ├── music.gob.gz
│   └── summaries/
├── context/                          # Enrichment context documents (committed)
│   ├── gospel-vocab.md
│   ├── titsw-framework.md
│   └── talk-calibration.md
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

### Configuration

```
GOSPEL_GRAPHS_DATA_DIR          # Default: ./data
GOSPEL_GRAPHS_DB                # Default: ./data/gospel.db
GOSPEL_GRAPHS_EMBEDDING_URL     # Default: http://localhost:1234/v1
GOSPEL_GRAPHS_EMBEDDING_MODEL   # Default: text-embedding-qwen3-embedding-4b
GOSPEL_GRAPHS_CHAT_URL          # Default: http://localhost:1234/v1
GOSPEL_GRAPHS_CHAT_MODEL        # Default: (auto-detect nemotron-3-nano)
GOSPEL_GRAPHS_ROOT              # Default: (auto-detect workspace root)
```

### MCP server config (replaces both gospel + gospel-vec)

```json
{
  "gospel-graphs": {
    "command": "C:/.../scripts/gospel-graphs/gospel-graphs.exe",
    "args": ["serve", "--data", "C:/.../scripts/gospel-graphs/data"],
    "type": "stdio"
  }
}
```

---

## 6. Migration Strategy

The originals stay put. gospel-graphs builds its own databases from the same source markdown files. No data migration needed for the core content — just re-run `gospel-graphs index`.

For the vector embeddings (which take hours to generate), migration commands import existing data:

```
gospel-graphs migrate-mcp --db ../gospel-mcp/gospel.db     # Import SQLite data
gospel-graphs migrate-vec --data ../gospel-vec/data/        # Import vector data + summary cache
```

**Transition plan:**
1. Build gospel-graphs with all capabilities
2. Run `migrate-mcp` + `migrate-vec` to bootstrap data (fast, no LLM calls)
3. Verify: every query that worked on the originals still works
4. Run enriched reindex on talks (this is the first new value — TITSW profiles)
5. Swap MCP config: remove gospel + gospel-vec, add gospel-graphs
6. Keep originals in `scripts/` as fallback (not deleted)

---

## 7. Phased Delivery

### Phase 1: Foundation — Scaffold + Migration (1-2 sessions)

**Scope:** New Go module that compiles, runs, and passes through to both databases. No enrichment yet — just proving the combined architecture works.

| Deliverable | Detail |
|-------------|--------|
| `scripts/gospel-graphs/` | New Go module with both dependencies |
| `.gitignore` | Data directory excluded |
| `cmd/gospel-graphs/main.go` | Command dispatch: index, serve, stats, version, migrate-mcp, migrate-vec |
| `internal/db/` | SQLite layer — schema from gospel-mcp, open/close/query |
| `internal/vec/` | chromem-go layer — store from gospel-vec, load/save/search |
| `internal/indexer/` | Scripture + talk + manual + book + music indexing, writing to both SQLite and chromem-go |
| `internal/indexer/summary.go` | LLM summarization + theme detection (ported from gospel-vec) |
| `internal/indexer/cache.go` | Summary cache (ported from gospel-vec) |
| `internal/vec/lmstudio.go` | LM Studio lifecycle (ported from gospel-vec) |
| `internal/search/` | Keyword search (FTS5) + semantic search (chromem-go) — separate, not yet combined |
| `internal/tools/` + `internal/mcp/` | 3 MCP tools: gospel_search (keyword + semantic via mode), gospel_get, gospel_list |
| `migrate-mcp` | Import gospel-mcp SQLite into gospel-graphs SQLite |
| `migrate-vec` | Import gospel-vec .gob.gz + summaries into gospel-graphs data |
| `context/` | Embed gospel-vocab.md, titsw-framework.md, talk-calibration.md |

**Verification:**
1. `gospel-graphs migrate-mcp && gospel-graphs migrate-vec` completes
2. `gospel-graphs stats` shows same counts as gospel-mcp + gospel-vec separately
3. `gospel-graphs search "faith in christ"` returns keyword results
4. `gospel-graphs search --mode semantic "faith in christ"` returns semantic results
5. MCP serve mode: all 3 tools respond correctly

**Stands alone:** Yes. This is a working replacement for both tools, without enrichment.

### Phase 2: TITSW Talk Enrichment (1-2 sessions)

**Scope:** Add TITSW enrichment to the talk indexing pipeline. This IS enriched-indexer Phase 1, built directly into gospel-graphs.

| Deliverable | Detail |
|-------------|--------|
| `internal/indexer/enricher.go` | TITSW teaching profile extraction — vocabulary approach + calibration context |
| Enriched talk prompt | TEACHING_PROFILE output format in system prompt |
| TITSW columns in SQLite | titsw_dominant, titsw_mode, etc. populated during index |
| TITSW metadata in chromem-go | Flat `map[string]string` fields in DocMetadata |
| TITSW search filters | `gospel_search` accepts titsw_mode, titsw_dominant, titsw_min_* params |
| TITSW in `gospel_get` | Talk responses include teaching profile |
| `--concurrency` flag | Worker pool for parallel LLM requests (default 1, max 4) |
|  FTS5 extended | talks_fts includes titsw_dominant, titsw_mode, titsw_keywords, titsw_summary |

**Verification:**
1. Index 5 ground-truth talks — scores within ±2 of targets
2. `gospel_search --source conference --titsw_mode enacted` returns only enacted talks
3. `gospel_search --source conference --titsw_min_teach 7` returns high-teach talks
4. `gospel_get` for a talk includes TITSW profile
5. FTS: `gospel_search "enacted love"` matches via FTS on titsw columns

### Phase 3: Scripture Enrichment (1 session)

**Scope:** Enriched-indexer Phase 2 — lens approach for scripture summaries.

| Deliverable | Detail |
|-------------|--------|
| Lens injection | gospel-vocab.md + titsw-framework.md injected into scripture summary prompt |
| Deeper keywords | Typological connections, Christ-type identification in keywords |
| Enriched scripture summaries | Richer summaries with theological depth |

**Verification:** Compare Alma 32, Zechariah 3, 1 Nephi 11 summaries before/after — typological terms present that weren't before.

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
| MCP config swap | Replace gospel + gospel-vec with gospel-graphs |
| Verification suite | Automated checks against ground truth |

**Verification:**
1. `gospel-graphs stats` shows complete counts
2. 5 ground-truth talks within ±2 of targets
3. All existing study agent/lesson agent workflows still work
4. Old tools still compile and are available as fallback

---

## 8. Verification Strategy

| Phase | Verification | Criteria |
|-------|-------------|---------|
| 1 | Migration correctness | Row counts match originals. 10 random queries produce same results. |
| 1 | MCP tool parity | All 7 original tool capabilities accessible through 3 new tools |
| 2 | TITSW scores | 5 ground-truth talks within ±2 |
| 2 | TITSW filters | Structured queries return correct subsets |
| 3 | Enriched scripture | Typological keywords present in Alma 32, Zechariah 3 |
| 4 | Combined search | Hybrid results outperform keyword-only for ambiguous queries |
| 5 | Full reindex | No errors across 5,500+ talks. Stats match expectations. |
| 5 | Cutover | All agents work with new MCP config |

---

## 9. Costs & Risks

### Costs
- **Development time:** 5-8 sessions across 5 phases
- **Reindex time:** Full talk reindex ~14-28 hours (depending on concurrency). Full scripture reindex ~4-6 hours. Run overnight.
- **Both dependencies:** Binary is larger (SQLite + chromem-go). Both are well-tested Go libraries.

### Risks

| Risk | Mitigation |
|------|-----------|
| Scope: replacing two working tools is high-stakes | Migration commands + verification before cutover. Originals stay as fallback. |
| Combined search complexity | Phase 4 — build after simpler modes are proven. Combined mode is additive, not required. |
| TITSW enrichment quality | Phase 0 experiments already complete (MAE=1.83). Calibration context proven. |
| Reindex time | --concurrency flag. Run overnight. Incremental mode after initial build. |
| Two databases in one process | Memory: chromem-go is in-memory during search. SQLite is file-backed. Tested separately, should compose fine. |

### What gets worse
- Larger single binary
- More complex indexer (writes to two databases per content item)
- If gospel-graphs has a bug, both search types are down (vs. one of two being down)

### What gets better
- One MCP config instead of two
- 3 tools instead of 7 — less agent confusion, less context window consumption
- No import pipeline — TITSW metadata generated and stored in one pass
- Combined search — the capability that neither tool offers alone
- One build, one update, one mental model

---

## 10. Creation Cycle Review

| Step | Question | This proposal |
|------|----------|--------------|
| **Intent** | Why? | Unified search is the foundation everything else builds on — study.ibeco.me, brain app, lesson prep. |
| **Covenant** | Rules? | Existing Go conventions. gospel-mcp + gospel-vec patterns. No data loss during migration. |
| **Stewardship** | Who? | dev agent builds. Michael reviews. gospel-mcp/gospel-vec maintained as fallback by existing code. |
| **Spiritual Creation** | Spec precise? | Schema, tool definitions, data flow, migration strategy all specified. Phase 1 is buildable. |
| **Line upon Line** | Phasing? | 5 phases. Phase 1 stands alone as a working replacement. Enrichment is additive. |
| **Physical Creation** | Who executes? | dev agent, one phase at a time. |
| **Review** | How to verify? | Migration row counts, ground-truth scores, MCP tool parity tests. |
| **Atonement** | If wrong? | Originals stay in scripts/. Revert MCP config. No data destroyed. |
| **Sabbath** | When to pause? | After Phase 1 — does the combined tool work as a replacement? After Phase 2 — are TITSW profiles good? After Phase 5 — full cutover, is study improved? |
| **Consecration** | Who benefits? | Michael directly. study.ibeco.me. brain app. Any downstream consumer. |
| **Zion** | Integration? | This IS the integration — two tools becoming one. Everything downstream gets simpler. |

---

## 11. Recommendation

**Build. This is the natural convergence of gospel-mcp and gospel-vec, and the home for the enriched pipeline.**

Phase 1 is the critical path — prove the combined architecture works, migrate existing data, verify parity. This is mostly porting existing code into a new structure. Phase 2 is where the new value starts — TITSW profiles on talks. Phase 5 is the commitment — swapping MCP config.

The enriched indexer proposal (Phases 1-3) folds directly into this tool's Phases 2-4. Instead of building enrichment into gospel-vec and then porting, build once into gospel-graphs.

**Sequencing in the overall plan:**
1. gospel-graphs Phase 1 (foundation + migration) — first
2. gospel-graphs Phase 2 (TITSW talk enrichment) — this IS enriched-indexer Phase 1
3. gospel-graphs Phases 3-4 (scripture enrichment, combined search)
4. gospel-graphs Phase 5 (full reindex, cutover)
5. study.ibeco.me Phase 1 — reads from gospel-graphs instead of two separate tools

**Hand off to:** dev agent, with this proposal + [enriched-indexer.md](../enriched-indexer.md) as spec references.

---

## Appendix: Name Discussion

Michael said "gospel-graphs" but isn't sure. The name implies graph visualization (which is study.ibeco.me's job). Options for reconsideration:

| Name | Pros | Cons |
|------|------|------|
| `gospel-graphs` | Michael's first instinct. "Graphs" can mean data graphs, not just visual. | Could confuse with study.ibeco.me graph visualization |
| `gospel-core` | Captures "foundation everything builds on" | Slightly generic |
| `gospel-search` | Clear purpose | Doesn't capture indexing role |
| `gospel-engine` | Captures both indexing + search | Slightly pretentious |
| `gospel-db` | Accurate | Boring |

Michael decides. Proposal uses `gospel-graphs` throughout as the working name.
