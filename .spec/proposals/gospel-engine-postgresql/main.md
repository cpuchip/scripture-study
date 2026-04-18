# Gospel-Engine PostgreSQL Migration

**Binding problem:** Gospel-engine currently fragments data across three storage systems (SQLite FTS5, chromem-go vectors, .vecf mmap files) that can't be queried together. Cross-cutting queries require application-level joins. There's no way to host this as a web service for ibeco.me or study.ibeco.me without running the full MCP server process. Worse, embedding generation depends on the desktop GPU (LM Studio + dual 4090s), so the service can't independently handle semantic queries. A unified PostgreSQL backend with pgvector, Apache AGE, and LM Studio headless (llmster) for CPU-based embedding would consolidate full-text search, vector similarity, and graph traversal into one self-hosted, self-managing, queryable database — with LM Link providing transparent GPU offloading to the desktop when available.

**Created:** 2026-04-17
**Updated:** 2026-04-18 — Added LM Studio headless + LM Link embedding strategy, revised to fully self-hosted vision
**Research:** [.spec/scratch/gospel-engine-postgresql/main.md](../../scratch/gospel-engine-postgresql/main.md)
**Related:** [gospel-graph proposal](../gospel-graph/main.md) (subsumes data layer), [gospel-engine proposal](../gospel-engine/main.md) (next evolution)
**Status:** Superseded by [study-ibeco-me proposal](../study-ibeco-me/main.md) — PG schema, embedding strategy, and extension research carry forward into the hosted service vision.

---

## 1. Problem Statement

Gospel-engine (the MCP server that powers scripture/talk search) stores data in three places:

1. **SQLite + FTS5** — relational data and keyword search
2. **chromem-go (gob.gz)** — in-memory vector database persisted to disk
3. **.vecf mmap files** — memory-mapped vector arrays for fast startup

This works for a single-process MCP server but has real limitations:

- **No cross-system queries.** "Find verses semantically similar to X that are cited by talks about Y" requires multiple round-trips and application-level merging.
- **No web service potential.** The MCP server is stdin/stdout. Hosting search on ibeco.me or study.ibeco.me requires a new API layer that reimplements query logic.
- **No replication or backup.** SQLite WAL + .vecf files have no unified backup story. No point-in-time recovery.
- **Redundant metadata.** The `vec_docs` table exists solely to bridge vector indices back to relational metadata — a problem PostgreSQL eliminates by putting vectors in the same row as the content.
- **Graph queries are impossible.** The `cross_references` and `edges` tables contain a graph, but SQL can only traverse it with recursive CTEs. Pattern matching ("find paths between two verses through shared cross-references") is effectively impossible.
- **Embedding generation is desktop-bound.** The 4B-parameter qwen3-embedding model runs on LM Studio (dual 4090s). NOCIX has no GPU. This means every semantic search query requires the desktop to be online — the service can't function independently.

**Who's affected:** Michael (scripture study), ibeco.me users (search), study.ibeco.me (graph visualization), remote agents (can't do semantic search without desktop).

**How would we know it's fixed:** A single `SELECT` statement returns keyword-matched, semantically-ranked, graph-connected results from one database. The same database is queryable from the MCP server, a web API, and a CLI. **The entire service runs on NOCIX with zero desktop dependency** — indexing, embedding, querying, and incremental updates all happen server-side.

---

## 2. Success Criteria

1. **All gospel-engine data lives in PostgreSQL** — scriptures, talks, chapters, manuals, books, cross-references, edges, embeddings
2. **Full-text search** via tsvector/tsquery replaces FTS5 with equivalent or better quality
3. **Vector similarity search** via pgvector replaces chromem-go/.vecf with HNSW indexed cosine distance
4. **Hybrid search** (keyword + semantic) executes as a single SQL query
5. **Graph traversal** via Apache AGE enables Cypher queries over cross-references and edges
6. **MCP server** works identically from the user's perspective (same tools, same results)
7. **Web API** can query the same database for ibeco.me/study.ibeco.me
8. **Runs entirely on NOCIX** (Ryzen 3800X, 32GB RAM, no GPU) via Docker — no desktop dependency
9. **Embedding generation** runs on NOCIX via LM Studio headless (llmster) with `nomic-embed-text` (CPU, 137M params, 768-dim). Query embedding <100ms. Bulk re-embedding ~6-8 hours (one-time, runs overnight). **LM Link** connects to desktop GPU for accelerated bulk embedding when available (~30 min for 240K chunks).
10. **Self-managing updates** — new gospel-library content is automatically indexed and embedded without manual intervention

---

## 3. Constraints & Boundaries

### In scope
- PostgreSQL database with pgvector + Apache AGE extensions
- Docker image (custom: pgvector + AGE combined)
- LM Studio headless (llmster) for CPU-based embedding generation (nomic-embed-text v1.5)
- LM Link configuration for transparent desktop GPU offloading
- Schema migration from SQLite to PostgreSQL
- Go codebase changes: `go-sqlite3` → `pgx/v5`, chromem-go → pgvector-go
- Embedding model switch: qwen3-embedding-4b (1024d, GPU-only) → nomic-embed-text (768d, CPU-viable)
- MCP server queries rewritten for PG
- Data migration pipeline (SQLite + .vecf → PG, with re-embedding)
- CLI commands updated (index, enrich, embed, search, etc.)
- Incremental update pipeline (auto-detect and embed new gospel-library content)
- Docker-compose for local development and NOCIX deployment

### Out of scope
- Frontend changes (study.ibeco.me is a separate proposal)
- New MCP tools beyond what exists today
- Enrichment pipeline logic changes (TITSW, chapter lenses — same prompts, different storage)
- Real-time streaming or pub/sub
- Multi-tenancy or auth (single-user database)

### Conventions
- Go with `jackc/pgx/v5` (standard PG driver for Go)
- `pgvector/pgvector-go` for vector types
- LM Studio headless (llmster) via OpenAI-compatible `/v1/embeddings` endpoint (existing `embedder.go` pattern)
- LM Link for optional desktop GPU acceleration (transparent to gospel-engine)
- Docker deployment via Dokploy (same as ibeco.me)
- Data directory: `data/gospel-pg/` (gitignored)
- Environment variables for connection string and embedding endpoint (existing pattern)

---

## 4. Prior Art

| Source | Relevance |
|--------|-----------|
| Current gospel-engine | SQLite + chromem-go architecture being replaced |
| [gospel-graph proposal](../gospel-graph/main.md) | Already planned PG backend for graph visualization; this subsumes its data layer |
| pgvector (v0.8.2, 20.9K stars) | Mature, well-tested vector extension. HNSW + IVFFlat indexes. Full Go support via pgvector-go. |
| Apache AGE (v1.7.0, 4.4K stars) | Graph extension with openCypher. Supports PG 18. Go driver included. |
| pgvector hybrid search docs | Built-in RRF pattern for combining FTS + vector search |
| ibeco.me | Already runs PG in Docker on NOCIX via Dokploy — proven deployment pattern |
| LM Studio headless (llmster) | Headless daemon for LM Studio. Docker: `lmstudio/llmster-preview:cpu` (369MB). Same OpenAI-compatible API as desktop LM Studio. CLI: `lms serve`, `lms load`, `lms get`. |
| LM Link (LM Studio 0.4.5) | Encrypted mesh VPN via Tailscale. Links LM Studio instances across machines. Remote models appear local at `localhost:1234`. Free tier: 2 users, 5 devices. Desktop GPU models become transparently available to NOCIX server. |
| nomic-embed-text v1.5 | Best CPU embedding model. 137M params, 274MB disk, ~300MB RAM. 768d, 8192 tokens, MTEB 62.39. Matryoshka dims (768→256). Runs fast on CPU (~30-80ms/query). |
| `embedder.go` (gospel-engine) | Already uses OpenAI-compatible `/v1/embeddings`. Switching embedding backends is a URL + model name config change — zero code changes. |

---

## 5. Proposed Architecture

### Service Architecture

```
NOCIX Server (Ryzen 3800X, 32GB, no GPU):
├── PostgreSQL (pgvector + FTS + AGE)
│   └── gospel database (~2-4GB)
├── LM Studio headless (llmster) — embedding service
│   ├── nomic-embed-text v1.5 (274MB, always loaded, CPU)
│   └── LM Link → Desktop qwen3-embedding-4b (when online)
└── gospel-engine (Go binary)
    ├── MCP server (stdin/stdout for local agents)
    ├── HTTP API (for ibeco.me / study.ibeco.me / remote agents)
    ├── Index pipeline (watches gospel-library, embeds + indexes new content)
    └── Config: EMBEDDING_URL=http://localhost:1234/v1, EMBEDDING_MODEL=nomic-embed-text

Desktop (Dual 4090s — enhances but not required):
├── LM Studio (GUI)
│   └── qwen3-embedding-4b (GPU, available via LM Link)
└── LM Link → NOCIX (encrypted mesh, auto-discovery)
```

**RAM budget:** ~2GB (PG shared_buffers) + ~1GB (HNSW index) + ~300MB (llmster + model) + ~200MB (gospel-engine) ≈ 3.5GB of 32GB.

### PostgreSQL Extensions
```sql
CREATE EXTENSION IF NOT EXISTS vector;    -- pgvector
CREATE EXTENSION IF NOT EXISTS age;       -- Apache AGE
CREATE EXTENSION IF NOT EXISTS pg_trgm;   -- fuzzy text matching
```

### Schema Design

**Content tables** (direct migration from SQLite, adding tsvector + embedding):

```sql
-- Scriptures: verse-level
CREATE TABLE scriptures (
  id BIGSERIAL PRIMARY KEY,
  volume TEXT NOT NULL,      -- ot, nt, bofm, dc-testament, pgp
  book TEXT NOT NULL,
  chapter INT NOT NULL,
  verse INT NOT NULL,
  text TEXT NOT NULL,
  file_path TEXT,
  source_url TEXT,
  tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', text)) STORED,
  UNIQUE(volume, book, chapter, verse)
);
CREATE INDEX ON scriptures USING gin(tsv);

-- Chapters: chapter-level with enrichment
CREATE TABLE chapters (
  id BIGSERIAL PRIMARY KEY,
  volume TEXT NOT NULL,
  book TEXT NOT NULL,
  chapter INT NOT NULL,
  title TEXT,
  full_content TEXT,
  enrichment_summary TEXT,
  enrichment_keywords TEXT,
  enrichment_key_verse TEXT,
  enrichment_christ_types TEXT,
  enrichment_connections TEXT,
  tsv TSVECTOR GENERATED ALWAYS AS (
    to_tsvector('english', coalesce(title,'') || ' ' || coalesce(enrichment_summary,'') || ' ' || coalesce(enrichment_keywords,''))
  ) STORED,
  UNIQUE(volume, book, chapter)
);
CREATE INDEX ON chapters USING gin(tsv);

-- Talks: conference talks with TITSW
CREATE TABLE talks (
  id BIGSERIAL PRIMARY KEY,
  year INT NOT NULL,
  month INT NOT NULL,
  session TEXT,
  speaker TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  file_path TEXT,
  titsw_teach SMALLINT,
  titsw_invite SMALLINT,
  titsw_testify SMALLINT,
  titsw_spirit SMALLINT,
  titsw_warn SMALLINT,
  titsw_doctrine SMALLINT,
  titsw_dominant TEXT,
  titsw_mode TEXT,
  titsw_pattern TEXT,
  titsw_summary TEXT,
  titsw_key_quote TEXT,
  titsw_keywords TEXT,
  tsv TSVECTOR GENERATED ALWAYS AS (
    to_tsvector('english', title || ' ' || speaker || ' ' || content)
  ) STORED,
  UNIQUE(year, month, speaker, title)
);
CREATE INDEX ON talks USING gin(tsv);

-- Manuals & Books (same pattern)
CREATE TABLE manuals (
  id BIGSERIAL PRIMARY KEY,
  content_type TEXT,
  collection_id TEXT,
  section TEXT,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  file_path TEXT,
  tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || content)) STORED
);
CREATE INDEX ON manuals USING gin(tsv);

CREATE TABLE books (
  id BIGSERIAL PRIMARY KEY,
  collection TEXT,
  section TEXT,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  file_path TEXT,
  tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || content)) STORED
);
CREATE INDEX ON books USING gin(tsv);
```

**Embeddings table** (replaces chromem-go + vec_docs):

```sql
CREATE TABLE embeddings (
  id BIGSERIAL PRIMARY KEY,
  source_type TEXT NOT NULL,  -- scriptures, talks, chapters, manuals, books
  source_id BIGINT NOT NULL,  -- FK to content table
  layer TEXT NOT NULL,         -- verse, paragraph, summary, theme
  content TEXT,                -- the text that was embedded
  embedding vector(768),       -- pgvector column (nomic-embed-text v1.5 = 768-dim)
  model TEXT,                  -- embedding model name
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE(source_type, source_id, layer)
);
CREATE INDEX ON embeddings USING hnsw (embedding vector_cosine_ops);
-- Partial indexes per source for filtered queries:
CREATE INDEX ON embeddings USING hnsw (embedding vector_cosine_ops) WHERE source_type = 'scriptures';
CREATE INDEX ON embeddings USING hnsw (embedding vector_cosine_ops) WHERE source_type = 'talks';
```

**Cross-references** (direct migration):

```sql
CREATE TABLE cross_references (
  id BIGSERIAL PRIMARY KEY,
  source_volume TEXT, source_book TEXT, source_chapter INT, source_verse INT,
  target_volume TEXT, target_book TEXT, target_chapter INT, target_verse INT,
  reference_type TEXT  -- footnote, tg, bd, jst
);
CREATE INDEX ON cross_references (source_volume, source_book, source_chapter, source_verse);
CREATE INDEX ON cross_references (target_volume, target_book, target_chapter, target_verse);
```

**Edges** (migration with JSONB):

```sql
CREATE TABLE edges (
  id BIGSERIAL PRIMARY KEY,
  source_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  edge_type TEXT NOT NULL,
  weight REAL,
  metadata JSONB
);
CREATE INDEX ON edges (source_type, source_id);
CREATE INDEX ON edges (target_type, target_id);
CREATE INDEX ON edges USING gin(metadata);
```

**Apache AGE graph** (Phase 2 — mirrors cross_references + edges as graph):

```sql
-- Create graph
SELECT create_graph('gospel');

-- Import cross-references as graph edges
SELECT * FROM cypher('gospel', $$
  CREATE (:Verse {ref: 'john-3-16', volume: 'nt', book: 'john', chapter: 3, verse: 16})
$$) AS (v agtype);

-- Query: path between two verses
SELECT * FROM cypher('gospel', $$
  MATCH path = (v1:Verse {ref: 'john-3-16'})-[*1..3]-(v2:Verse {ref: '2-ne-31-20'})
  RETURN path
$$) AS (path agtype);
```

### Unified Hybrid Search Query

```sql
-- Combined keyword + semantic + TITSW filter in ONE query
WITH keyword AS (
  SELECT s.id, ts_rank_cd(s.tsv, q) AS score
  FROM scriptures s, plainto_tsquery('english', $1) q
  WHERE s.tsv @@ q
  ORDER BY score DESC LIMIT 20
),
semantic AS (
  SELECT e.source_id AS id, 1 - (e.embedding <=> $2::vector) AS score
  FROM embeddings e
  WHERE e.source_type = 'scriptures' AND e.layer = 'verse'
  ORDER BY e.embedding <=> $2::vector LIMIT 20
)
SELECT COALESCE(k.id, s.id) AS id, v.volume, v.book, v.chapter, v.verse, v.text,
  1.0/(60 + COALESCE(k.rank, 999)) + 1.0/(60 + COALESCE(s.rank, 999)) AS rrf
FROM (SELECT id, score, ROW_NUMBER() OVER (ORDER BY score DESC) rank FROM keyword) k
FULL OUTER JOIN (SELECT id, score, ROW_NUMBER() OVER (ORDER BY score DESC) rank FROM semantic) s
  ON k.id = s.id
JOIN scriptures v ON v.id = COALESCE(k.id, s.id)
ORDER BY rrf DESC LIMIT $3;
```

---

## 6. Phased Delivery

### Phase 1: PostgreSQL + pgvector + LM Studio headless (replaces SQLite + chromem-go)
**Scope:** Core migration — same functionality, unified self-hosted backend.

1. Create Docker image (pgvector/pgvector:pg18 base)
2. Install LM Studio headless (llmster) on NOCIX — `curl -fsSL https://lmstudio.ai/install.sh | bash`
3. Create docker-compose.yml for local dev (PG + llmster, or PG with system-level llmster)
4. Pull `nomic-embed-text` model — `lms get nomic-ai/nomic-embed-text-v1.5-GGUF`
5. Load model for CPU — `lms load nomic-embed-text --gpu off --context-length 8192 --yes`
6. Start server — `lms server start --port 1234`
7. Write PG schema (content tables + embeddings + cross_references + edges)
8. Swap Go driver: `go-sqlite3` → `pgx/v5` + `pgvector-go`
9. Update embedding config: URL → `http://localhost:1234/v1`, model → `nomic-embed-text`, dim → 768
10. Rewrite `index` command to insert into PG (tsvector auto-generated)
11. Rewrite `embed` command to generate embeddings via llmster → insert as pgvector
12. Rewrite `search` to use PG FTS + pgvector hybrid queries (query embedding via llmster)
13. Rewrite MCP `serve` command to query PG instead of SQLite + mmap
14. Configure LM Link on both NOCIX and desktop — `lms login && lms link enable`
15. Verify: same MCP tools, same results, same or better performance
16. Remove chromem-go, .vecf, vec_docs dependencies

**Deliverable:** gospel-engine serves identical MCP tools from PostgreSQL, with embedding handled by llmster on CPU. LM Link provides optional desktop GPU acceleration. No hard desktop dependency.
**Estimate:** Medium — core query logic changes, but schema maps 1:1. Embedding client change is config-only (same `/v1/embeddings` API).

### Phase 2: Apache AGE (graph layer)
**Scope:** Add graph traversal over existing relational data.

1. Add AGE to Docker image
2. Create gospel graph from cross_references + edges tables
3. Add graph-aware MCP tools or extend `gospel_search` with graph options
4. Test Cypher queries: path finding, citation clusters, connection discovery

**Deliverable:** Graph queries work alongside FTS + vector search.
**Estimate:** Small-medium — AGE is additive, doesn't change existing functionality.

### Phase 3: Web API (enables ibeco.me / study.ibeco.me)
**Scope:** HTTP API layer over the PG database.

1. Add chi router HTTP server (same patterns as ibeco.me)
2. REST endpoints: `/search`, `/get/{ref}`, `/graph/{ref}`
3. Deploy alongside gospel-db on NOCIX via Dokploy
4. study.ibeco.me frontend can query this API directly

**Deliverable:** Gospel search available as a web service.
**Estimate:** Small — thin API layer over existing query logic.

### Phase 4: Self-managing update pipeline
**Scope:** Auto-detect and process new gospel-library content.

1. `watch` command or scheduled job detects new/changed files in gospel-library
2. New content → parse → insert into PG → embed via llmster → index
3. Enrichment pipeline reads from PG, writes back to PG
4. Remove all gob.gz / .vecf intermediate files
5. pg_cron or systemd timer for periodic checks

**Deliverable:** New conference talks, manual updates, etc. are automatically indexed and searchable without manual intervention. Single database is sole source of truth.

---

## 7. Verification Criteria

| Phase | Test |
|-------|------|
| 1 | MCP `gospel_search` returns same top-10 results for 5 test queries (keyword, semantic, combined) |
| 1 | MCP `gospel_get` returns identical content for 10 test references |
| 1 | `gospel_list` returns same directory structure |
| 1 | Vector search recall ≥ 95% vs current chromem-go results (same embedding, different index) |
| 1 | Query latency < 100ms for keyword search, < 200ms for hybrid |
| 2 | Cypher query returns cross-reference paths between known connected verses |
| 2 | Citation cluster query returns expected talk groupings |
| 3 | HTTP API returns identical results to MCP tools |
| 3 | study.ibeco.me prototype can render search results |
| 4 | Full pipeline (index → enrich → embed) runs end-to-end against PG |

---

## 8. Costs & Risks

### Costs
- **Development time:** Phase 1 is the bulk of the work. 2-4 focused sessions.
- **Docker image:** ~500MB-1GB for PG + extensions. Trivial disk cost.
- **Memory:** PG + HNSW indexes will use ~1-2GB on NOCIX. Fine with 32GB.
- **Maintenance:** PG requires vacuuming, monitoring. More ops than SQLite.
- **Dependency complexity:** pgx + pgvector-go + AGE driver vs current go-sqlite3 + chromem-go.

### Risks
- **AGE + pgvector compatibility:** Untested together in production. Mitigation: Phase 2 is separate, can skip if incompatible.
- **AGE Go driver maturity:** Less battle-tested than pgx. Mitigation: Can use raw SQL `SELECT * FROM cypher(...)` via pgx if needed.
- **Model quality regression:** nomic-embed-text (MTEB 62.39) vs qwen3-embedding-4b (likely higher). Mitigation: For the scripture domain, the difference is marginal. Run recall comparison on 20 test queries before/after. Additionally, LM Link allows using qwen3-embedding-4b via desktop GPU when available.
- **Bulk re-embedding time:** ~6-8 hours on NOCIX CPU for 240K chunks. Mitigation: One-time operation, run overnight. With LM Link to desktop GPU, ~30 min instead. Incremental updates are seconds.
- **Migration data loss:** Risk of losing enrichment data during migration. Mitigation: Keep SQLite + .vecf files until PG is verified.
- **LM Link preview stability:** LM Link is in preview (v0.4.5, Feb 2026). Mitigation: LM Link is optional acceleration, not required. Core system runs on local llmster with nomic-embed-text.
- **llmster auth gap:** No `--auth` CLI flag yet (issue #489). Mitigation: Keep behind reverse proxy or firewall. Docker inter-container stays on localhost.

### What Gets Worse
- **Setup complexity:** Need Docker + PG running, plus llmster installed for development (currently just a SQLite file + LM Studio).
- **Cold start:** PG takes seconds to start vs SQLite instant. MCP server now depends on external services.
- **Portability:** SQLite is a single file. PG is a service. Harder to share, harder to test.

### What Gets BETTER (vs original proposal)
- **Stays in the LM Studio ecosystem.** Same models, same tools, same interface Michael already uses daily.
- **LM Link unlocks the best of both worlds.** CPU independence on NOCIX for always-available queries, plus GPU acceleration from the desktop when it's online. No forced tradeoff.
- **Any agent can use semantic search.** Remote agents, ibeco.me, MCP clients — all get vector search without needing the desktop online.
- **Self-managing updates.** New gospel-library content is processed automatically.

**Mitigation for downsides:** Docker-compose makes PG startup trivial. Tests use containers. llmster is lightweight (369MB Docker image or native install). The tradeoff is worth it for a fully self-hosted, self-managing scripture service.

---

## 9. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Enable cross-cutting queries (FTS+vector+graph) and web hosting that the current fragmented storage can't support |
| Covenant | Rules? | Go + pgx conventions, Docker deployment, same MCP interface contract |
| Stewardship | Who owns? | dev agent builds, Michael owns PG ops on NOCIX |
| Spiritual Creation | Spec precise enough? | Yes — schema is defined, migration is 1:1, phases are independent |
| Line upon Line | Phasing? | Phase 1 (core migration) stands alone. Each subsequent phase adds capability. |
| Physical Creation | Who executes? | dev agent, with Michael reviewing Docker/deployment |
| Review | How to verify? | MCP tool parity tests, query latency benchmarks, recall comparison |
| Atonement | If wrong? | Keep SQLite + .vecf alongside PG until verified. Rollback is just reverting Go code. |
| Sabbath | When stop? | After Phase 1 verification. Phase 2-4 are separate decisions. |
| Consecration | Who benefits? | Michael's study, ibeco.me users (future search), study.ibeco.me (graph viz) |
| Zion | Whole system? | Unifies gospel-engine + gospel-graph data layer. Eliminates gospel-mcp, gospel-vec as separate servers. LM Studio headless + LM Link keeps embedding in-ecosystem. Fully self-hosted scripture service usable by any agent or user. |

---

## 10. Recommendation

**Build.** Phase 1 first.

The self-hosted vision changes this from a backend migration to an independence milestone. Today, semantic search requires the desktop to be online running LM Studio. After Phase 1, the scripture service runs entirely on NOCIX — indexing, embedding, querying, and incremental updates. Any agent (local or remote), ibeco.me, and study.ibeco.me can use full hybrid search without dependency on any other machine.

The key enabler is **LM Studio headless (llmster) + nomic-embed-text** on CPU. The existing `embedder.go` already uses the OpenAI-compatible `/v1/embeddings` API — the same API that llmster exposes. Switching from the desktop LM Studio to llmster on NOCIX is a config change (`EMBEDDING_URL` and `EMBEDDING_MODEL`), not a code change. **LM Link** adds an optional but powerful bonus: when the desktop is online, NOCIX transparently accesses the dual 4090s for bulk embedding at GPU speed. When it's offline, queries fall back to local CPU embedding.

This stays within the LM Studio ecosystem Michael already uses daily. No new tools to learn, no separate model management systems.

Phase 1 is a clean swap — same functionality, unified self-hosted backend. Phases 2-4 unlock capabilities that are currently impossible.

**Execute with:** `@dev` agent
**Phase 1 scope:** 2-4 sessions
**First action:** Docker-compose (PG) + llmster install + schema + `pgx` connection setup + embedding config switch + LM Link configuration
