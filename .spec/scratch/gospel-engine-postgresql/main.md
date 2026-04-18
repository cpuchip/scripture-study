# Gospel-Engine PostgreSQL Migration — Research Findings

*Created: 2026-04-17*
*Proposal: [.spec/proposals/gospel-engine-postgresql/main.md](../../proposals/gospel-engine-postgresql/main.md)*

---

## Binding Problem

Gospel-engine currently runs three separate storage systems (SQLite FTS5, chromem-go in-memory vectors, .vecf mmap files) that can't be queried together in a single transaction. Cross-cutting queries ("find verses semantically similar to X that are also cross-referenced by Y conference talks") require application-level joins across systems. There's no way to expose this as a hosted service for ibeco.me or study.ibeco.me without running the full MCP server. PostgreSQL with pgvector and Apache AGE could unify full-text search, vector similarity, and graph traversal into one queryable database.

---

## Current Architecture Inventory

### SQLite Schema (gospel-engine)

| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `scriptures` | Verse-level text | volume, book, chapter, verse, text, file_path, source_url |
| `chapters` | Chapter-level with enrichment | volume, book, chapter, title, full_content + enrichment_summary/keywords/key_verse/christ_types/connections |
| `talks` | Conference talks with TITSW | year, month, session, speaker, title, content + 6 TITSW scores + titsw_dominant/mode/pattern/summary/key_quote/keywords |
| `manuals` | CFM/TITSW/handbooks | content_type, collection_id, section, title, content |
| `books` | Additional texts (Lectures on Faith, etc.) | collection, section, title, content |
| `cross_references` | Verse→verse links | source/target volume/book/chapter/verse, reference_type (footnote/tg/bd/jst) |
| `edges` | Semantic/thematic connections | source_type/id, target_type/id, edge_type, weight, metadata (JSON) |
| `index_metadata` | Incremental indexing tracker | file_path, mtime, size |
| `vec_docs` | Vector metadata bridge | collection, vec_idx, doc_id, content, source, layer, book, chapter, reference, speaker, year, month |

### FTS5 Virtual Tables
- `scriptures_fts` (text)
- `chapters_fts` (title, enrichment_summary, enrichment_keywords, enrichment_christ_types)
- `talks_fts` (title, speaker, content, titsw_dominant, titsw_mode, titsw_keywords, titsw_summary)
- `manuals_fts` (title, content)
- `books_fts` (title, content)

### Vector Storage (chromem-go → .vecf)
- In-memory chromem-go DB persisted as gob.gz per source
- Collections: `{source}-{layer}` (e.g., "scriptures-verse", "conference-paragraph")
- Layers: verse, paragraph, summary, theme
- Convert pipeline: gob.gz → .vecf (mmap-friendly, 16-byte header + float32 arrays)
- Embedding model: `text-embedding-qwen3-embedding-4b` via LM Studio localhost:1234
- Embeddings pre-normalized → dot product = cosine similarity
- vec_docs table bridges vector indices to metadata

### Go Dependencies
```
github.com/mattn/go-sqlite3 v1.14.24
github.com/philippgille/chromem-go v0.7.0
golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90
```

### CLI Pipeline
1. `index` → SQLite + chromem-go gob.gz
2. `enrich` → TITSW enrichment via LLM
3. `enrich-scriptures` → Chapter-level 5-lens enrichment
4. `embed-enrichments` → Vectorize enrichment output
5. `convert` → gob.gz → .vecf
6. `serve` → MCP server (mmap .vecf or fallback to gob.gz)

---

## pgvector Research

**Source:** github.com/pgvector/pgvector (v0.8.2, 20.9K stars)

### Key Capabilities
- **Vector types:** `vector` (up to 16K dimensions, HNSW indexable up to 2K), `halfvec` (up to 16K, indexable up to 4K), `bit`, `sparsevec`
- **Distance functions:** L2 (`<->`), inner product (`<#>`), cosine (`<=>`), L1 (`<+>`), Hamming (`<~>`), Jaccard (`<%>`)
- **Index types:**
  - **HNSW** — multilayer graph, better query performance, slower build, more memory. No training step needed.
  - **IVFFlat** — divides vectors into lists, faster builds, less memory, slightly lower recall
- **Hybrid search:** Built-in support — combine `tsvector/tsquery` FTS with vector similarity via Reciprocal Rank Fusion
- **Filtering:** Filter by metadata columns + vector similarity in one query. Iterative scan (0.8.0+) handles filtered ANN well.
- **Storage:** 4 bytes × dimensions + 8 bytes per vector. ~2KB per 512-dim vector.
- **Table limit:** 32TB per non-partitioned table. Partitioning supported.
- **Replication:** Full WAL support, point-in-time recovery
- **Docker:** `pgvector/pgvector:pg18` (adds to official Postgres image)

### HNSW Index Details
- Parameters: `m` (max connections/layer, default 16), `ef_construction` (candidate list, default 64)
- Query: `hnsw.ef_search` (default 40, higher = better recall)
- Build: Set `maintenance_work_mem = '8GB'` for best build speed; parallel workers supported
- Iterative scans (0.8.0+): `SET hnsw.iterative_scan = strict_order` — auto-scans more when filters reduce results

### CPU vs GPU — Critical Finding
**pgvector is 100% CPU-based for both indexing and querying.** No GPU required at any stage. The HNSW and IVFFlat algorithms are pure CPU operations.

**Embedding generation** still needs a model (LM Studio on GPU, or an API). But once embeddings exist, pgvector handles all storage, indexing, and similarity search on CPU alone.

**For NOCIX server (Ryzen 8 3800x, 32GB, no GPU):**
- pgvector will work perfectly for vector storage and search
- Embedding generation options: (a) generate on desktop with dual 4090s, insert to PG; (b) use a remote API; (c) run a small quantized model on CPU (slow but possible)
- With ~31K verses + ~10K talks + chapters/manuals/books, total vectors at 512-dim ≈ 50K-200K depending on layers. At 2KB each, that's 100-400MB of vector data — trivially fits in 32GB RAM

### Go Driver (pgvector-go v0.3.0)
- Supports pgx, pg, Bun, Ent, GORM, sqlx
- `pgvector.NewVector([]float32{1,2,3})` → `$1` in queries
- `pgxvec.RegisterTypes(ctx, conn)` for pgx connection setup
- Hybrid search example available (RRF with OpenAI-compatible embedding endpoint)
- **pgx is the natural choice** — most common Go PG driver, best performance

### Size Estimate for Gospel Data

| Content | Est. Rows | × 4 layers | Vector Size (512-dim) |
|---------|-----------|------------|----------------------|
| Scriptures (verses) | ~31,100 | ~124K | ~248 MB |
| Conference talks | ~10,000 | ~40K | ~80 MB |
| Chapters | ~1,500 | ~6K | ~12 MB |
| Manuals | ~2,000 | ~8K | ~16 MB |
| Books | ~500 | ~2K | ~4 MB |
| **Total** | **~45K base** | **~180K vectors** | **~360 MB** |

Well within 32GB RAM. HNSW index adds ~1.5-2x overhead, so ~540-720MB total for vectors + index. PG shared_buffers at 8GB would be comfortable.

---

## Apache AGE Research

**Source:** github.com/apache/age (v1.7.0, 4.4K stars)

### What It Is
- PostgreSQL extension adding graph database functionality
- openCypher query language (same as Neo4j)
- Graph data stored alongside relational data in the same database
- Supports PG 11-18
- Docker image: `apache/age`

### Key Capabilities
- **Cypher queries** from SQL: `SELECT * FROM cypher('graph_name', $$ MATCH (v)-[r]->(v2) RETURN v,r,v2 $$) as (v agtype, r agtype, v2 agtype);`
- **Hybrid SQL+Cypher:** Mix relational queries with graph traversal
- **Property indexes** on vertices and edges
- **Variable-length path traversal** — find all connections within N hops
- **Built-in Go driver** (`drivers/golang` in AGE repo)
- Graph visualization tool: age-viewer (web UI)

### Relevance to Gospel Data
The `cross_references` and `edges` tables in the current schema ARE a graph. AGE would make them traversable with graph queries:

```cypher
-- "What talks cite verses connected to Alma 32?"
MATCH (v:Verse {book: 'alma', chapter: 32})-[:FOOTNOTE]->(v2:Verse)<-[:CITES]-(t:Talk)
RETURN t.speaker, t.title, v2.reference

-- "Show the connection path between two verses"
MATCH path = (v1:Verse {ref: 'john-3-16'})-[*1..4]-(v2:Verse {ref: 'mosiah-3-19'})
RETURN path

-- "Find talk clusters by shared cross-references"
MATCH (t1:Talk)-[:CITES]->(v:Verse)<-[:CITES]-(t2:Talk)
WHERE t1 <> t2
RETURN t1.title, t2.title, count(v) AS shared_refs
ORDER BY shared_refs DESC
```

### Compatibility Concern
**AGE + pgvector on the same database:** AGE's GitHub issues page has a proposal (#1121) for "Vector handling with extension(pgvector)" — suggesting this is desired but may not be seamless yet. The extensions should technically work together since they operate on different aspects (AGE adds agtype + cypher parser, pgvector adds vector type + index methods), but this needs testing.

### Docker Considerations
- No official "pgvector + AGE" combined image exists
- Would need a custom Dockerfile: start from `pgvector/pgvector:pg18`, add AGE build
- Or start from `apache/age`, add pgvector build
- Either way, ~15 minutes of Dockerfile work

---

## PostgreSQL Built-in Full-Text Search (FTS)

### tsvector/tsquery vs SQLite FTS5

| Feature | SQLite FTS5 | PostgreSQL tsvector |
|---------|-------------|-------------------|
| Tokenization | Simple/unicode61 | Configurable parsers, dictionaries, stop words |
| Stemming | English by default | Snowball, Ispell, custom dictionaries |
| Ranking | BM25 (built-in) | ts_rank, ts_rank_cd (cover density) |
| Highlighting | snippet(), highlight() | ts_headline() |
| Phrase search | NEAR, phrase | `<->` (adjacent), `<N>` (within N) |
| Index type | Automatic (FTS5 table) | GIN or GiST on tsvector column |
| Language support | Limited | 30+ languages, customizable |
| Synonyms | None | Synonym dictionaries, thesaurus |
| Performance | Very fast for single-file | Very fast, scales better with concurrent queries |

**PostgreSQL FTS is more capable than FTS5** — better stemming, ranking, phrase proximity, and synonym support. It also integrates natively with pgvector's hybrid search pattern.

### Hybrid Search in One Query
```sql
-- FTS + vector similarity in one query with RRF
WITH keyword_results AS (
  SELECT id, ts_rank_cd(tsv, query) AS keyword_score
  FROM scriptures, plainto_tsquery('english', 'faith hope charity') query
  WHERE tsv @@ query
  ORDER BY keyword_score DESC LIMIT 20
),
semantic_results AS (
  SELECT id, 1 - (embedding <=> $1) AS semantic_score
  FROM scriptures
  ORDER BY embedding <=> $1 LIMIT 20
)
SELECT COALESCE(k.id, s.id) AS id,
  1.0 / (60 + COALESCE(k_rank, 999)) + 1.0 / (60 + COALESCE(s_rank, 999)) AS rrf_score
FROM (SELECT id, keyword_score, ROW_NUMBER() OVER (ORDER BY keyword_score DESC) AS k_rank FROM keyword_results) k
FULL OUTER JOIN (SELECT id, semantic_score, ROW_NUMBER() OVER (ORDER BY semantic_score DESC) AS s_rank FROM semantic_results) s ON k.id = s.id
ORDER BY rrf_score DESC LIMIT 10;
```

This is exactly what gospel-engine's `combined` search mode does now, but implemented across two separate systems. PostgreSQL unifies it into a single query.

---

## Other Relevant PG Extensions

### pg_trgm (Trigram Similarity)
- Built-in extension, no install needed
- Fuzzy text matching: `similarity('word1', 'word2')` → float
- GIN/GiST indexable for `LIKE`, `ILIKE`, `%` (similarity), `<->` (word distance)
- **Use case:** Fuzzy scripture reference parsing ("1 Nephi" vs "1ne" vs "first nephi")

### pg_stat_statements (Performance Monitoring)
- Track query performance, find slow queries
- Built-in, just needs `shared_preload_libraries`

### pg_cron (Scheduled Jobs)
- Schedule re-indexing, re-embedding jobs inside PostgreSQL
- Could replace external cron for enrichment pipeline

### JSONB (Built-in)
- Replaces the `metadata JSON` column in edges table with indexable, queryable JSONB
- GIN indexable for containment queries: `WHERE metadata @> '{"type": "christological"}'`

---

## Hardware Assessment

### NOCIX Server (Production Target)
- **CPU:** AMD Ryzen 7 3800X (8C/16T, 3.9GHz base, 4.5GHz boost)
- **RAM:** 32GB DDR4
- **GPU:** None
- **Storage:** Likely SSD (confirm with Michael)

**Assessment:**
- PostgreSQL + pgvector: Excellent fit. HNSW index builds are CPU-bound and 8 cores will handle 180K vectors well. Query latency should be <10ms for ANN search.
- Apache AGE: CPU-only, no GPU needed. Graph traversal is memory/CPU bound.
- FTS: GIN indexes are CPU-bound for builds, fast for queries.
- Total estimated DB size: ~2-4GB (text + vectors + indexes). Fits trivially in 32GB.
- `shared_buffers = 8GB`, `maintenance_work_mem = 4GB` would be comfortable.

### Embedding Generation Strategy
Embeddings must be generated elsewhere since NOCIX has no GPU:
1. **Desktop (dual 4090s):** Generate embeddings locally via LM Studio, insert to PG over network or via dump/restore
2. **Remote API:** Use OpenAI/Voyage/Cohere embedding API (costs money but simple)
3. **CPU inference on NOCIX:** Possible with ONNX runtime + quantized model, but slow (~10-50x slower than GPU)
4. **Hybrid:** Generate on desktop during indexing, serve from NOCIX. This mirrors the current workflow (LM Studio on desktop → .vecf files → deploy).

**Recommendation:** ~~Option 4 (hybrid) is the natural evolution.~~ **Revised: Option 5 (LM Studio headless on NOCIX + LM Link to desktop)** — see CPU Embedding Research below.

---

## CPU Embedding Research (2026-04-18)

### The Architectural Gap

The original proposal assumed embedding generation on desktop (GPU) with pgvector on NOCIX for storage/search. **Michael identified the critical flaw: query-time embedding also needs a model.** Every vector similarity search requires embedding the search phrase first. If that requires GPU/desktop, the service isn't self-contained — it depends on the desktop being online for every query.

**Revised vision:** A fully self-hosted, self-managing scripture service on NOCIX that handles all operations (indexing, embedding, querying) independently. No desktop dependency for runtime operation.

### Current Embedding Setup

| Setting | Value |
|---------|-------|
| Model | `text-embedding-qwen3-embedding-4b` (4B params) |
| Dimensions | 1024 |
| Endpoint | `http://localhost:1234/v1` (LM Studio OpenAI-compatible) |
| Total chunks | ~240K vectors (verse, paragraph, summary, theme layers) |
| Backend | chromem-go (gob.gz) → .vecf (mmap) |

The `embedder.go` in gospel-engine already uses the generic OpenAI-compatible `/v1/embeddings` endpoint. **This means switching to LM Studio headless (llmster) is a config change — same API format, different URL and model name.**

### CPU Embedding Model Benchmarks (April 2026)

Source: Morph benchmark, BentoML guide, PE Collective, SearchLayer, InsiderLLM, AIMUltiple.

| Model | Params | Dims | Context | MTEB Score | Disk | RAM (loaded) | CPU Speed |
|-------|--------|------|---------|------------|------|-------------|-----------|
| all-MiniLM-L6-v2 | 23M | 384 | 256 | ~56 | 46MB | ~90MB | Very fast (~15ms/query) |
| e5-small-v2 | 118M | 768 | 512 | ~100% Top-5* | ~250MB | ~300MB | Fast (~16ms/query) |
| nomic-embed-text v1.5 | 137M | 768 | 8192 | 62.39 | 274MB | ~300MB | Fast (~30-80ms/query) |
| nomic-embed-text-v2-moe | ~305M | 768 | 8192 | ~63 | ~610MB | ~700MB | Medium |
| mxbai-embed-large | 335M | 1024 | 512 | 64.68 | 670MB | ~700MB | Medium (~100-300ms/query) |
| Qwen3-Embedding 0.6B | 600M | 32-4096* | 8192 | ~60 | ~400MB | ~1.2GB | Slower on CPU |
| bge-m3 | 568M | 1024 | 8192 | ~63 | 1.2GB | ~1.2GB | Medium |
| EmbeddingGemma-300M | 308M | 768 | 2048 | Good | 622MB | ~600MB | Medium |

*e5-small achieved 100% Top-5 accuracy in AIMUltiple's Amazon product retrieval benchmark at 16ms latency — but this is one specific benchmark, not MTEB overall.
*Qwen3-Embedding supports configurable dimensions from 32 to 4096.

### Hosting Options for CPU Embedding

**Option 1: LM Studio headless (llmster) on NOCIX + LM Link to desktop** ⭐ RECOMMENDED
- Pros: Already using LM Studio. Same OpenAI-compatible `/v1/embeddings` API. Docker image available (`lmstudio/llmster-preview:cpu`). LM Link gives transparent GPU offloading to desktop when available. Can run any model already in your LM Studio ecosystem.
- Cons: LM Link is in Preview (free tier: 2 users, 5 devices). Auth for headless server is CLI-only (no `--auth` flag yet, open issue #489).
- Integration: Existing `embedder.go` works **completely unchanged**. Same URL format, same API, same model names. Literally just the model name and maybe the URL changes in config.
- Docker: `lmstudio/llmster-preview:cpu` (369MB), or install llmster directly via `curl -fsSL https://lmstudio.ai/install.sh | bash`
- LM Link setup: `lms login` + `lms link enable` on both machines → automatic device discovery → remote models appear locally

**Option 2: Ollama on NOCIX**
- Pros: Go-friendly REST API (OpenAI-compatible `/v1/embeddings`), runs many models, easy Docker sidecar, proven ecosystem, active development
- Cons: External service dependency (but lightweight — 300MB for nomic model). Separate ecosystem from LM Studio. No remote GPU linking.
- Integration: Existing `embedder.go` works unchanged. Just change `GOSPEL_ENGINE_EMBEDDING_URL` and `GOSPEL_ENGINE_EMBEDDING_MODEL`
- Docker: `ollama/ollama` image, mount model volume

**Option 3: llama.cpp server**
- Pros: Lightweight, GGUF model loading, OpenAI-compatible API mode, fastest raw performance (~10-50% faster than LM Studio/Ollama wrappers)
- Cons: Less ecosystem, manual model management, no remote linking
- Integration: Same OpenAI-compatible API

**Option 4: ONNX Runtime in Go**
- Pros: No external server, fastest CPU inference, embedded in binary
- Cons: CGo complexity (`yalue/onnxruntime_go`), model conversion needed, less mature Go bindings

### Why LM Studio > Ollama for This Project

1. **Already in the ecosystem.** Michael already uses LM Studio on the desktop with qwen3-embedding-4b. No context switching.
2. **LM Link is the killer feature.** With LM Link, the desktop's dual 4090s become transparently available to the NOCIX server. Bulk embedding runs at GPU speed when the desktop is online. Query-time embeddings fall back to CPU locally when it's not.
3. **Same model everywhere.** Can run `nomic-embed-text` on CPU for queries AND have `qwen3-embedding-4b` available via LM Link for bulk work or quality comparisons. No need to pick one model forever.
4. **Same API surface.** LM Studio's server exposes `/v1/embeddings` — identical to what `embedder.go` already calls. Zero code changes.
5. **Docker-native.** `lmstudio/llmster-preview:cpu` is purpose-built for headless CPU deployment.

### LM Studio Headless Research (2026-04-18)

**llmster (the headless daemon):** Introduced in LM Studio v0.4.0 (Jan 2026). Decoupled from the GUI — runs as a standalone background process. Exposes the same REST API as the desktop app. Available as:
- Docker image: `lmstudio/llmster-preview:cpu` (369MB, x86 CPU-only, 10K+ pulls)
- Direct install: `curl -fsSL https://lmstudio.ai/install.sh | bash`
- Systemd service via `lms daemon up` + `lms server start`

**CLI workflow on NOCIX:**
```bash
# Install
curl -fsSL https://lmstudio.ai/install.sh | bash

# Start daemon
lms daemon up

# Download embedding model
lms get nomic-ai/nomic-embed-text-v1.5-GGUF

# Load model (CPU mode, no GPU)
lms load nomic-embed-text --gpu off --context-length 8192 --yes

# Start server (bind to network for Docker inter-container access)
lms server start --port 1234 --bind 0.0.0.0

# Enable LM Link (for desktop GPU access)
lms login
lms link enable
```

**Network binding:** By default, lms binds to `127.0.0.1`. For Docker inter-container or LAN access:
- CLI flag: `lms server start --bind 0.0.0.0`
- Env var: `LMS_SERVER_HOST=0.0.0.0`
- Config file: `~/.lmstudio/.internal/http-server-config.json` → `"networkInterface": "0.0.0.0"`
- Systemd: `Environment="LMS_SERVER_HOST=0.0.0.0"` in service file

**Auth concern:** Open issue (#489, Feb 2026) — `lms server start` has no `--auth` flag. When binding to 0.0.0.0, the API is unauthenticated. Mitigation: keep behind reverse proxy (Nginx/Caddy on Dokploy) or use firewall rules. For Docker inter-container communication, 127.0.0.1 binding is fine since gospel-engine connects locally.

### LM Link Research (2026-04-18)

**What it is:** Feature in LM Studio 0.4.5 (Feb 25, 2026). Built in partnership with Tailscale. Creates end-to-end encrypted mesh VPN connections between LM Studio instances. Remote models appear as local in the model loader and API.

**How it works:**
- Uses Tailscale's `tsnet` library (userspace WireGuard, no kernel-level permissions)
- Devices discover each other automatically through encrypted tunnels
- Works across different networks (CGNAT, firewalls, different subnets)
- Coexists independently with existing Tailscale VPN usage
- Any tool pointing to `localhost:1234` uses remote models transparently

**Setup:**
1. Enable LM Link on both desktop and NOCIX server
2. Sign in with same LM Studio account on both
3. Devices auto-discover via Tailscale mesh
4. `lms link set-preferred-device` to control routing
5. Load remote model: desktop's GPU models appear in NOCIX's model list

**Free tier:** 2 users, 5 devices each (10 total). Plenty for Michael's setup.

**Critical capability for this project:** When desktop is online, NOCIX can use `qwen3-embedding-4b` running on the 4090s — accessed at `localhost:1234` just like a local model. When desktop is offline, NOCIX falls back to locally-loaded `nomic-embed-text` on CPU. **Gospel-engine doesn't need to know or care which is happening.** The API is identical either way.

**Limitation:** LM Link routes requests to a single remote machine. No distributed inference (can't split a model across nodes). Fine for embedding models.

### Model Strategy with LM Link

**Two-tier approach:**

| Scenario | Model | Where | Speed |
|----------|-------|-------|-------|
| Query embedding (real-time) | nomic-embed-text v1.5 | NOCIX CPU (always available) | ~30-80ms |
| Bulk embedding (initial load) | nomic-embed-text v1.5 | NOCIX CPU | ~6-8 hours |
| Bulk embedding (desktop online) | qwen3-embedding-4b | Desktop GPU via LM Link | ~30 min |
| On-demand quality boost | qwen3-embedding-4b | Desktop GPU via LM Link | ~5-15ms |

**Important constraint:** All embeddings in the same pgvector index MUST use the same model and dimensions. You can't mix nomic (768d) and qwen3 (1024d) embeddings in the same column. Choose one for the production index.

**Recommendation:** Use `nomic-embed-text` (768d) as the production model. It's the one that must work independently on NOCIX. If quality testing shows nomic is insufficient, could switch to `qwen3-embedding-0.6b` (configurable dims, could set to 768) or accept the dependency on desktop for the larger model.

**Alternative approach:** Use LM Link as the *primary* strategy — always embed via the desktop's qwen3-embedding-4b when online, queue requests when offline, process queued items when desktop comes back online. This preserves the current model quality but adds complexity and doesn't achieve full independence.

### Model Recommendation: `nomic-embed-text` v1.5

**Why nomic over others:**
1. **Best quality/size ratio for CPU.** 137M params, 274MB. Runs in ~300MB RAM. MTEB 62.39 is solid for RAG.
2. **8192 token context.** Handles full paragraphs, chapter summaries, talk excerpts without truncation. (all-MiniLM only supports 256 tokens — too short for scripture passages.)
3. **Most popular lightweight embedding model.** Best-tested, most documented, available in every runtime (LM Studio, Ollama, llama.cpp).
4. **Matryoshka dimensions.** Can reduce from 768 to 256/384 for faster search if needed, without re-embedding.
5. **Fast on CPU.** ~30-80ms per query embedding on a Ryzen CPU. Sub-100ms is imperceptible for interactive search.

**Why not keep 1024-dim (mxbai/bge-m3)?**
- mxbai-embed-large only supports 512 token context — too short for many use cases
- bge-m3 at 1.2GB is heavy for a CPU-only server
- 768-dim is sufficient. More dimensions means more storage, slower search, minimal quality gain.

**Why not Qwen3-Embedding 0.6B?**
- At 600M params, noticeably slower on CPU
- Less battle-tested in LM Studio's GGUF ecosystem than nomic
- Its configurable dimensions could match 1024, but the CPU speed cost isn't worth it

### Bulk Embedding Estimates (NOCIX Ryzen 3800X, CPU only)

Using nomic-embed-text via LM Studio headless (llmster):

| Content | Chunks | Time/chunk (est.) | Total Time |
|---------|--------|-------------------|------------|
| Scriptures (verse) | 31K | ~50ms | ~26 min |
| Scriptures (paragraph) | ~31K | ~50ms | ~26 min |
| Scriptures (summary) | ~31K | ~60ms | ~31 min |
| Conference talks (paragraph) | ~40K | ~60ms | ~40 min |
| Chapters/manuals/books | ~10K | ~50ms | ~8 min |
| **Total ~240K chunks** | | | **~6-8 hours** |

Compared to desktop GPU via LM Link: ~30 min for 240K chunks. CPU is ~12x slower, but this is a one-time operation that can run overnight (or use LM Link when the desktop is online). Incremental updates (new conference talks twice a year, manual updates) would be minutes, not hours.

### Ollama as Fallback Option

If LM Studio headless proves problematic (LM Link preview instability, auth gaps), Ollama remains a solid backup:

Three SDK options (Oct 2025 comparison by Rost Glukhov):

1. **Official Ollama Go SDK** (`github.com/ollama/ollama/api`): Full-featured, typed client. `client.Embed(ctx, &EmbedRequest{Model: "nomic-embed-text", Input: "text"})` → `EmbedResponse{Embeddings: [][]float32{...}}`
2. **OpenAI Go SDK** (`github.com/openai/openai-go`): Works via Ollama's OpenAI-compatible endpoint. **Already compatible with existing `embedder.go` code.**
3. **Community SDK** (`github.com/rozoomcool/go-ollama-sdk`): Simpler, Ollama-specific.

Both LM Studio and Ollama expose the same OpenAI-compatible `/v1/embeddings` endpoint, so switching between them is purely a config change (URL + model name).

### Revised Architecture: Self-Hosted Scripture Service

```
NOCIX Server (Ryzen 3800X, 32GB, no GPU):
├── PostgreSQL (pgvector + FTS + AGE)
│   └── gospel database (~2-4GB)
├── LM Studio headless (llmster) — embedding service
│   ├── nomic-embed-text v1.5 (274MB, always loaded, CPU)
│   └── LM Link → Desktop's qwen3-embedding-4b (when online)
└── gospel-engine (Go binary)
    ├── MCP server (stdin/stdout for local agents)
    ├── HTTP API (for ibeco.me / study.ibeco.me / remote agents)
    ├── Index pipeline (watches gospel-library, embeds + indexes new content)
    └── Config: EMBEDDING_URL=http://localhost:1234/v1, EMBEDDING_MODEL=nomic-embed-text

Desktop (Dual 4090s, optional — enhances but not required):
├── LM Studio (GUI or daemon)
│   └── qwen3-embedding-4b (GPU, available via LM Link)
└── LM Link → NOCIX (encrypted mesh, auto-discovery)
```

**Total NOCIX RAM estimate:** ~2GB (PG shared_buffers) + ~1GB (HNSW index in memory) + ~300MB (llmster + nomic model) + ~200MB (gospel-engine) = ~3.5GB. Well within 32GB.

**Key insight: zero code changes to the embedding client.** The `embedder.go` already calls `/v1/embeddings` with JSON `{model, input}`. LM Studio headless responds with the same `{data: [{embedding: [...]}]}` format. It's a config change.

### Dimension Change Impact

Switching from 1024-dim (qwen3-embedding-4b) to 768-dim (nomic-embed-text) means **all existing embeddings must be regenerated.** Different models produce vectors in different semantic spaces — you can't mix them. This is expected as part of the PG migration anyway.

**Storage savings:** 768-dim × 4 bytes × 240K = ~700MB vs 1024-dim × 4 = ~940MB. HNSW index similarly smaller. Net save ~400MB including index overhead.

### Gospel-Library Update Strategy

For the "self-managing as new updates come out" vision:

1. **Watch for changes:** gospel-engine periodically checks gospel-library directory for new/modified files (already has `index_metadata` table with mtime tracking)
2. **Incremental index:** New files get parsed → inserted into PG content tables (auto-generates tsvector)
3. **Incremental embed:** New content gets embedded via local llmster → inserted into embeddings table
4. **Cross-reference update:** Re-scan footnotes for any new cross-reference links
5. **No desktop involvement required.** The entire pipeline runs on NOCIX. Desktop GPU is available via LM Link for acceleration when online.

Conference talks arrive twice a year (~20-30 new talks). Embedding 30 talks at ~5 chunks each × 4 layers = ~600 embeddings × ~50ms = ~30 seconds. Trivial.

---

## Existing Related Proposals

### gospel-graph (.spec/proposals/gospel-graph/main.md)
- Proposed study.ibeco.me as a graph visualization site
- Already planned PostgreSQL backend with "graph-optimized schema"
- Used recursive CTEs, not AGE
- **This new proposal subsumes gospel-graph's data layer.** If gospel-engine moves to PG, the graph visualization frontend can query it directly instead of needing its own import pipeline.

### gospel-engine (.spec/proposals/gospel-engine/main.md)
- Existing proposal for gospel-engine unification (already partially complete)
- Unified SQLite + FTS5 + chromem-go into single MCP server
- **This new proposal is the next evolution** — same architecture goals, different storage backend.

---

## Docker Strategy

### Custom Image: pgvector + AGE + FTS
```dockerfile
FROM pgvector/pgvector:pg18

# Install AGE build dependencies
RUN apt-get update && apt-get install -y \
    build-essential libreadline-dev zlib1g-dev flex bison \
    postgresql-server-dev-18 git

# Build and install Apache AGE
RUN cd /tmp && \
    git clone --branch PG18/v1.7.0-rc0 https://github.com/apache/age.git && \
    cd age && \
    make && make install && \
    cd / && rm -rf /tmp/age

# pg_trgm is already included in contrib
```

### docker-compose.yml
```yaml
services:
  gospel-db:
    build: ./docker/gospel-db
    volumes:
      - ./data/gospel-pg:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: gospel
      POSTGRES_USER: gospel
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
    shm_size: '4gb'  # For parallel HNSW builds
    secrets:
      - db_password

  # LM Studio headless for embedding generation
  # Alternative: install llmster directly on the host (preferred for LM Link)
  llmster:
    image: lmstudio/llmster-preview:cpu
    volumes:
      - ./data/lmstudio:/root/.lmstudio
    ports:
      - "1234:1234"
    environment:
      LMS_SERVER_HOST: "0.0.0.0"
    # Pre-download model: lms get nomic-ai/nomic-embed-text-v1.5-GGUF
    # Load model: lms load nomic-embed-text --gpu off --context-length 8192 --yes

secrets:
  db_password:
    file: ./secrets/db_password.txt
```

**Note:** For LM Link support, llmster is better installed directly on the host (not in Docker) since LM Link uses Tailscale's userspace networking. Docker's network namespace may interfere with device discovery. The docker-compose above is for local dev without LM Link.

Data directory would live at `data/gospel-pg/` (gitignored, like private-brain).

---

## Migration Complexity Assessment

### What's Easy
- Schema migration: SQLite tables → PG tables is nearly 1:1
- FTS migration: FTS5 virtual tables → tsvector columns + GIN indexes
- Go driver swap: `mattn/go-sqlite3` → `jackc/pgx/v5`
- Vector storage: chromem-go → pgvector columns on existing tables (NO separate vec_docs needed)
- CLI pipeline: Same commands, different backend

### What's Medium
- Embedding re-insertion: Need to read .vecf or gob.gz → insert as pgvector columns
- Query rewriting: SQLite FTS5 MATCH syntax → PG tsquery syntax
- MCP server adaptation: New query patterns for combined FTS+vector+graph
- Docker deployment: Custom image build, data volume management

### What's Hard
- Apache AGE Cypher queries from Go: AGE's Go driver is in-repo but less mature than pgx
- Graph schema design: Mapping cross_references + edges into AGE's vertex/edge model
- Testing: Need a PG instance running for tests (Docker-based test setup)

### Migration Risk: LOW
The current architecture was designed with clean separations. The `Searcher` interface for vectors already allows transparent backend swapping. SQLite → PG is well-trodden territory.

---

## Key Decision Points

1. **pgvector vs keep chromem-go:** pgvector wins. Unified storage, SQL-queryable, no separate .vecf pipeline, WAL replication, standard tooling.

2. **Apache AGE vs recursive CTEs:** AGE is more expressive for complex graph traversal (variable-length paths, pattern matching). CTEs work for simple 1-2 hop queries. The gospel-graph proposal already identified graph queries as the goal. **Recommendation: Include AGE but as Phase 2** — get pgvector + FTS working first.

3. **Inline vectors vs separate vector table:** Inline wins. Add `embedding vector(512)` directly to scriptures, chapters, talks tables. Eliminates the vec_docs bridge table entirely. Multiple layers can be separate columns (`embedding_verse`, `embedding_summary`, etc.) or a separate embeddings table keyed by source.

4. **Single embeddings table vs inline columns:** For flexibility, a dedicated embeddings table keyed by (source_type, source_id, layer) is cleaner than adding 4 vector columns to every content table. Allows adding new layers without schema changes.

5. **Embedding generation workflow:** Index on desktop (GPU) → insert to PG. Then PG is the single source of truth that both MCP server and web apps query.

6. **Docker vs native PG:** Docker. Matches existing deployment patterns (Dokploy), portable, easy to version-lock extensions.
