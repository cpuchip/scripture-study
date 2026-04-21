# study.ibeco.me — Hosted Gospel Study Service

> **Scope Refocus (2026-04-21):**
>
> This proposal originally bundled two things: the engine-server backend (PG + pgvector + MCP client + auth delegation) and the user-facing study site (search UI, histories, notes, annotations). The backend has **SHIPPED** as `engine.ibeco.me` and now lives in its own proposal: [`gospel-engine/v2-hosted.md`](../gospel-engine/v2-hosted.md).
>
> Going forward, **`study-ibeco-me/` is the UI-only roadmap.** What was "Phase 4 deferred" (user-facing web features, search histories, notes, studies, annotations) is the next thing to spec under this name. The text below is the original combined proposal, kept for historical context until the UI-only rewrite lands.
>
> **Backend status:** SHIPPED Apr 20. First study used a user-minted token same day.
> **UI status:** not started.

---

**Binding problem:** Gospel search is trapped inside a local MCP server. It can't serve web clients, can't be shared, and dies when the desktop goes offline. There's no way for ibeco.me, study.ibeco.me, remote agents, or other users to search scriptures without running the full gospel-engine binary locally. Moving to a hosted service at study.ibeco.me would make gospel search a permanent, always-available API — accessible from MCP clients, web browsers, and any HTTP client — while keeping the database self-hosted on NOCIX.

**Created:** 2026-04-18
**Updated:** 2026-04-18 — Dokploy raw Dockerfile deployment, PG18, self-hosted MCP binary for auto-update, pre-loaded gospel-library
**Research:** [.spec/scratch/study-ibeco-me/main.md](../../scratch/study-ibeco-me/main.md)
**Supersedes:** [gospel-engine-postgresql proposal](../gospel-engine-postgresql/main.md) (PG schema, embedding strategy, and extension stack carry forward; deployment model and architecture change)
**Related:** [gospel-graph proposal](../gospel-graph/main.md) (graph viz frontend becomes a consumer of this API)
**Repo:** `scripts/gospel-engine-v2/` ([github.com/cpuchip/gospel-engine](https://github.com/cpuchip/gospel-engine))
**Status:** Proposed

---

## 1. Problem Statement

The current gospel search ecosystem is fragmented across three MCP servers (gospel-mcp, gospel-vec, gospel-engine), each running as a local stdio process. This creates hard limits:

- **No web access.** ibeco.me can't search scriptures. study.ibeco.me can't exist.
- **No sharing.** Each user needs the full binary + gospel-library + SQLite databases + vector files.
- **Desktop-dependent.** Semantic search requires LM Studio on the desktop with dual 4090s.
- **No user features.** Search histories, notes, personal studies — impossible without a server.
- **Fragmented data.** Three separate MCP servers with three separate storage systems (SQLite FTS5, chromem-go, mmap vectors).

**Who's affected:** Michael (scripture study from any device), ibeco.me users (future search), remote agents (can't do semantic search), anyone who might use a public gospel search API.

**How would we know it's fixed:** `curl https://study.ibeco.me/api/search?q=faith` returns hybrid search results. An MCP client configured with a bearer token does the same thing from VS Code. The service runs 24/7 on NOCIX with zero desktop dependency.

---

## 2. Success Criteria

1. **study.ibeco.me serves a REST API** for gospel search (keyword, semantic, hybrid), content retrieval, and listing — the same capabilities as the current three MCP servers combined
2. **PostgreSQL 18 with pgvector + pg_trgm** stores all content and embeddings in one database (via `pgvector/pgvector:pg18` image, no custom build needed)
3. **Embedding generation** runs on NOCIX via LM Studio headless (nomic-embed-text, CPU). LM Link to desktop GPU is optional acceleration.
4. **Gospel-library downloads** are managed server-side — the backend pulls content from the Church API, indexes, and embeds it
5. **Token-based auth** protects the API from abuse. Tokens provisioned via ibeco.me (service token delegation)
6. **A new MCP client** (`gospel-mcp`) translates MCP JSON-RPC to HTTP calls against study.ibeco.me
7. **Deployed on Dokploy** alongside ibeco.me on the NOCIX server using **two services**: a Database service (`pgvector/pgvector:pg18`) and an Application service (raw Dockerfile, built on the Dokploy server)
8. **Current MCP tools are preserved** — agents using `gospel_search`, `gospel_get`, `gospel_list` see identical behavior through the new MCP client
9. **MCP client auto-updates** by checking the server's `/api/version` endpoint on startup and downloading new builds from `/download/gospel-mcp-{os}-{arch}`
10. **Gospel-library content is pre-loaded** to `/opt/gospel/gospel-library/` and `/opt/gospel/books/` on NOCIX (already streaming via rsync from Michael's machine) and mounted read-only into the container — no Church API bulk download from server IP
11. **Embeddings are pre-computed on desktop using nomic-embed-text v1.5** (the SAME model that runs on NOCIX — embeddings cannot be mixed across models) accelerated by Michael's dual 4090s via LM Studio, then uploaded to NOCIX. Server bulk-loads on first run; only query-time embedding (~50ms) runs on NOCIX CPU using the identical nomic model.
12. **TITSW and enrichment data** from the current gospel-engine is migrated to PG (TITSW columns on conference_talks, chapter_lenses table, etc.)
13. **Backups** automated via Dokploy's built-in PG backup (local + optional S3)

---

## 3. Scope

### In scope (Phase 1 — Core API)
- Go backend HTTP server with chi router (single Dockerfile, built by Dokploy)
- PostgreSQL 18 via Dokploy **Database** service using `pgvector/pgvector:pg18` directly (no custom image needed for Phase 1)
- Schema migrations run by the Go server on startup (pgvector + pg_trgm extensions are already included in the pgvector image)
- REST endpoints: `/api/search`, `/api/get/{ref}`, `/api/list`, `/api/health`
- Token validation middleware (bearer tokens, `stdy_` prefix)
- Content indexing pipeline (parse gospel-library markdown → PG)
- Embedding pipeline (content → llmster → pgvector)
- Gospel-library + books pre-loaded to `/opt/gospel/gospel-library/` and `/opt/gospel/books/` on NOCIX (mounted into container as read-only volume) — avoids initial Church API download
- **Embeddings pre-computed on desktop using nomic-embed-text v1.5** (identical model to server) and rsynced to `/opt/gospel/embeddings/` on NOCIX (mounted read-only). Server bulk-loads via PG `COPY` on first run.
- **TITSW + enrichment migration script:** one-time export from existing gospel-engine SQLite → PG insert
- Static file server for MCP client binaries (`/download/gospel-mcp-{os}-{arch}`) — enables one-binary distribution and auto-update
- MCP client binary that wraps the HTTP API and supports self-update (with rollback safety: `<self>.prev` backup, SHA256 verify, opt-out env, "first successful run" gate)
- Single Dockerfile deployed to Dokploy (no docker-compose)
- Dokploy-scheduled PG backups configured at deploy time

### In scope (Phase 2 — Content Management)
- Gospel-library download module (port from `scripts/gospel-library/`) — for incremental updates after initial pre-load
- Automatic indexing of new content (scheduled checks for new conference talks, manual updates)
- Admin endpoints for triggering re-index / re-embed
- Apache AGE graph extension (requires custom PG image at that point — see Phase 2 architecture notes)

### In scope (Phase 3 — Auth Delegation)
- Service token for ibeco.me → study.ibeco.me delegation
- ibeco.me endpoint: "Get study.ibeco.me access token"
- Rate limiting per token / per IP
- Anonymous tier with strict rate limits

### Deferred
- User-facing web features (search histories, notes, studies)
- study.ibeco.me frontend / UI
- GitHub Copilot SDK integration
- Multi-user study sharing
- Apache AGE graph queries (Phase 2; may require custom PG18 image if AGE doesn't support PG18 by then)

### Out of scope
- Changes to ibeco.me's existing auth system (only adding one new endpoint)
- Mobile apps
- Real-time collaboration
- Content from non-Church sources

### Conventions
- Go with `jackc/pgx/v5` (standard PG driver)
- `pgvector/pgvector-go` for vector types
- `go-chi/chi/v5` for HTTP routing (same as ibeco.me)
- LM Studio headless via OpenAI-compatible `/v1/embeddings`
- Deployment: **single Dockerfile** per service, built on the Dokploy server itself. No docker-compose, no intermediate registry. PG is a separate Dokploy Database service.
- PostgreSQL 18 (latest stable). pgvector image: `pgvector/pgvector:pg18`.
- Gospel-library + books: pre-loaded to `/opt/gospel/gospel-library/` and `/opt/gospel/books/` on NOCIX, mounted read-only into the app container at `/data/gospel-library` and `/data/books`.
- **Embedding model is fixed: nomic-embed-text v1.5 (768-dim).** Used identically on desktop (GPU pre-compute) and NOCIX (CPU query-time). Embeddings are not portable across models, so changing requires full re-embed.
- Bearer token auth: `stdy_` prefix + 64 hex chars, bcrypt hashed
- Environment variables for all config (same pattern as ibeco.me)

---

## 4. Prior Art

| Source | Relevance |
|--------|-----------|
| ibeco.me (scripts/becoming/) | Proven Dokploy deployment pattern. Chi router. PG backend. Token auth with bcrypt + prefix lookup. Google OAuth. |
| gospel-engine (scripts/gospel-engine/) | MCP server with 3 tools (search, get, list). Hybrid search (FTS5 + vector). Embedder uses OpenAI-compatible API. Schema and query patterns carry forward. |
| gospel-mcp (scripts/gospel-mcp/) | FTS-only MCP server. Being replaced. |
| gospel-vec (scripts/gospel-vec/) | Vector-only MCP server with chromem-go. Being replaced. |
| gospel-library (scripts/gospel-library/) | Downloads from Church API. Rate-limited Go client. HTML→Markdown pipeline. Will be ported into backend. |
| [PG migration proposal](../gospel-engine-postgresql/main.md) | Schema design, pgvector research, embedding strategy (llmster + nomic + LM Link), hybrid search queries. Core technical foundation. |
| LM Studio headless (llmster) | CPU embedding on NOCIX. Docker: `lmstudio/llmster-preview:cpu`. LM Link for optional GPU offload. Researched in PG migration scratch file. |

---

## 5. Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│ NOCIX Server (Ryzen 3800X, 32GB RAM, no GPU)                │
│                                                              │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │ PostgreSQL           │    │ LM Studio headless (llmster)│ │
│  │ + pgvector           │    │ nomic-embed-text v1.5 (CPU) │ │
│  │ + pg_trgm            │    │ LM Link → desktop (optional)│ │
│  │ + AGE (Phase 2)      │    └──────────┬──────────────────┘ │
│  └──────────┬───────────┘               │                    │
│             │                           │                    │
│  ┌──────────┴───────────────────────────┴──────────────────┐ │
│  │ gospel-engine (Go HTTP server)                          │ │
│  │                                                          │ │
│  │  /api/search  — hybrid FTS + vector search              │ │
│  │  /api/get     — retrieve scripture/talk by reference     │ │
│  │  /api/list    — browse content, stats                    │ │
│  │  /api/health  — health check                             │ │
│  │  /api/admin/* — re-index, token management (Phase 2+)   │ │
│  │                                                          │ │
│  │  Auth middleware: Bearer token (stdy_ prefix)            │ │
│  │  Rate limiting: per-token + per-IP                       │ │
│  └──────────────────────────────────────────────────────────┘ │
│                              │                                │
│                   Dokploy (Traefik reverse proxy)             │
│                              │                                │
└──────────────────────────────┼────────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
   ┌──────┴───────┐   ┌───────┴──────┐   ┌────────┴────────┐
   │ gospel-mcp   │   │ ibeco.me     │   │ Web browsers    │
   │ (MCP client) │   │ (search UI)  │   │ (future)        │
   │ stdio ↔ HTTP │   │              │   │                 │
   └──────────────┘   └──────────────┘   └─────────────────┘
```

### Token Flow (Service Token Delegation)

```
1. User logs into ibeco.me (Google OAuth or email/password)
2. User clicks "Get study.ibeco.me access token"
3. ibeco.me calls:
   POST https://study.ibeco.me/api/admin/tokens
   Authorization: Bearer <service_token>
   Body: { "user_id": "...", "name": "Michael's MCP" }
4. study.ibeco.me creates token, returns:
   { "token": "stdy_a1b2c3...", "prefix": "stdy_a1b2c3d4" }
5. ibeco.me displays token to user (shown once)
6. User configures MCP client:
   { "env": { "GOSPEL_TOKEN": "stdy_a1b2c3..." } }
7. MCP client sends:
   GET https://study.ibeco.me/api/search?q=faith
   Authorization: Bearer stdy_a1b2c3...
```

### Two Binaries from One Repo

```
gospel-engine-v2/
├── cmd/
│   ├── gospel-engine/    # HTTP server — deployed on NOCIX
│   └── gospel-mcp/       # MCP client — runs on user's machine
```

**gospel-engine** (server): Full backend. Connects to PG, llmster, serves HTTP API.
**gospel-mcp** (client): Thin MCP server that translates JSON-RPC to HTTP calls. Carries a bearer token. Ships as a single binary — no databases, no models, no dependencies.

### Repo Structure

```
gospel-engine-v2/
├── cmd/
│   ├── gospel-engine/         # HTTP server binary
│   │   └── main.go
│   └── gospel-mcp/            # MCP client binary
│       └── main.go
├── internal/
│   ├── api/                   # HTTP handlers + router
│   │   ├── router.go          # Chi router setup
│   │   ├── search.go          # /api/search handler
│   │   ├── get.go             # /api/get handler
│   │   ├── list.go            # /api/list handler
│   │   ├── download.go        # /download/gospel-mcp-{os}-{arch}
│   │   └── admin.go           # /api/admin/* handlers
│   ├── auth/                  # Token validation + rate limiting
│   │   ├── middleware.go
│   │   └── tokens.go
│   ├── db/                    # PostgreSQL access layer
│   │   ├── db.go              # Connection, migrations
│   │   ├── migrate.go         # Runs SQL migrations on startup
│   │   ├── scriptures.go      # Scripture queries
│   │   ├── talks.go           # Conference talk queries
│   │   ├── search.go          # Hybrid search queries
│   │   └── tokens.go          # Token storage
│   ├── embed/                 # Embedding client (port embedder.go)
│   │   └── embedder.go
│   ├── index/                 # Content indexing pipeline
│   │   └── indexer.go
│   ├── selfupdate/            # MCP client self-update logic
│   │   └── updater.go
│   └── search/                # Search orchestration
│       └── hybrid.go          # FTS + vector + RRF merging
├── migrations/                # SQL migration files (embedded via go:embed)
│   ├── 001_schema.sql
│   └── 002_indexes.sql
├── Dockerfile                 # Single Dockerfile, built by Dokploy
├── .dockerignore
├── go.mod
├── go.sum
└── README.md
```

**Note:** No docker-compose, no `docker/gospel-db/` directory. PG is provisioned as a separate Dokploy Database service using the `pgvector/pgvector:pg18` image directly. The app's Dockerfile builds both binaries (`gospel-engine` server and `gospel-mcp` client) for distribution.

### Dockerfile Pattern (single-stage build, multiple targets)

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build server
RUN go build -o /build/gospel-engine ./cmd/gospel-engine

# Build MCP client for multiple platforms (distributed via HTTP)
RUN mkdir -p /build/binaries && \
    GOOS=windows GOARCH=amd64 go build -o /build/binaries/gospel-mcp-windows-amd64.exe ./cmd/gospel-mcp && \
    GOOS=linux   GOARCH=amd64 go build -o /build/binaries/gospel-mcp-linux-amd64 ./cmd/gospel-mcp && \
    GOOS=darwin  GOARCH=amd64 go build -o /build/binaries/gospel-mcp-darwin-amd64 ./cmd/gospel-mcp && \
    GOOS=darwin  GOARCH=arm64 go build -o /build/binaries/gospel-mcp-darwin-arm64 ./cmd/gospel-mcp

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/gospel-engine /usr/local/bin/
COPY --from=builder /build/binaries /opt/mcp-binaries
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gospel-engine"]
```

Dokploy builds this Dockerfile locally on the NOCIX server on each deploy. No intermediate registry needed.

### Self-Updating MCP Client

The gospel-engine server hosts the MCP client binaries it built at `/download/gospel-mcp-{os}-{arch}`. The MCP client checks for updates on startup:

1. On startup, `gospel-mcp` calls `GET https://study.ibeco.me/api/version` — returns `{ "version": "2026.04.18-abc123", "sha256": "..." }`
2. If the running version differs, download the new binary to `<self>.new`
3. Verify SHA256 match
4. Platform-specific replacement:
   - **Unix:** `os.Rename("<self>.new", "<self>")`, re-exec
   - **Windows:** can't overwrite running exe. Write `<self>.new`, spawn a small updater script that waits for the parent to exit, then renames and re-launches.
5. Continue normal operation

This means users install the MCP client once; subsequent updates are automatic whenever the server is redeployed. Auto-update can be disabled via `GOSPEL_AUTO_UPDATE=false` env var.

### Gospel-Library Pre-Loading

To avoid hammering the Church API from the NOCIX IP on first deploy:

1. **Upload to NOCIX:** `rsync -avz gospel-library/ nocix:/opt/gospel/gospel-library/` and `rsync -avz books/ nocix:/opt/gospel/books/` (one-time, from Michael's machine — already in progress)
2. **Mount read-only into container** via Dokploy volume config:
   - `/opt/gospel/gospel-library:/data/gospel-library:ro`
   - `/opt/gospel/books:/data/books:ro`
   - `/opt/gospel/embeddings:/data/embeddings:ro`
3. **Server reads from `/data/gospel-library` and `/data/books`** on startup, indexes everything into PG
4. **Phase 2** adds incremental downloads for new conference talks (April + October) — these are minimal Church API hits (~30 talks twice a year)

---

## 6. Schema

Carries forward from the [PG migration proposal](../gospel-engine-postgresql/main.md) with additions for tokens:

```sql
-- Core content tables (from PG migration proposal)
CREATE TABLE scriptures (
    id          BIGSERIAL PRIMARY KEY,
    volume      TEXT NOT NULL,       -- ot, nt, bofm, dc-testament, pgp
    book        TEXT NOT NULL,
    chapter     INT NOT NULL,
    verse       INT,
    reference   TEXT NOT NULL,       -- "1 Nephi 3:7"
    content     TEXT NOT NULL,
    file_path   TEXT NOT NULL,
    tsv         tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED
);
CREATE INDEX idx_scriptures_tsv ON scriptures USING GIN(tsv);
CREATE INDEX idx_scriptures_ref ON scriptures(reference);

CREATE TABLE conference_talks (
    id          BIGSERIAL PRIMARY KEY,
    year        INT NOT NULL,
    month       TEXT NOT NULL,
    speaker     TEXT NOT NULL,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    file_path   TEXT NOT NULL,
    position    TEXT,
    session     TEXT,
    tsv         tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED
);
CREATE INDEX idx_talks_tsv ON conference_talks USING GIN(tsv);
CREATE INDEX idx_talks_speaker ON conference_talks(speaker);
CREATE INDEX idx_talks_year ON conference_talks(year, month);

-- Unified embeddings table
CREATE TABLE embeddings (
    id          BIGSERIAL PRIMARY KEY,
    source_type TEXT NOT NULL,        -- scriptures, conference, manual, book
    source_id   BIGINT NOT NULL,
    layer       TEXT NOT NULL,        -- verse, paragraph, summary, theme
    embedding   vector(768) NOT NULL, -- nomic-embed-text dimensions
    model       TEXT NOT NULL DEFAULT 'nomic-embed-text',
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_embeddings_hnsw ON embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 128);
CREATE INDEX idx_embeddings_source ON embeddings(source_type, source_id);

-- Cross-references (existing data)
CREATE TABLE cross_references (
    id          BIGSERIAL PRIMARY KEY,
    from_ref    TEXT NOT NULL,
    to_ref      TEXT NOT NULL,
    note        TEXT,
    source      TEXT                  -- footnote, tg, bd, etc.
);
CREATE INDEX idx_xref_from ON cross_references(from_ref);
CREATE INDEX idx_xref_to ON cross_references(to_ref);

-- Auth tokens
CREATE TABLE api_tokens (
    id          BIGSERIAL PRIMARY KEY,
    external_user TEXT,               -- ibeco.me user identifier (optional)
    name        TEXT NOT NULL,
    token_hash  TEXT NOT NULL,        -- bcrypt hash
    prefix      TEXT NOT NULL,        -- first 12 chars for fast lookup
    is_service  BOOLEAN DEFAULT FALSE,-- service tokens can create user tokens
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    last_used   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    rate_limit  INT DEFAULT 60        -- requests per minute
);
CREATE INDEX idx_tokens_prefix ON api_tokens(prefix);

-- Index metadata (for incremental updates)
CREATE TABLE index_metadata (
    file_path   TEXT PRIMARY KEY,
    mtime       TIMESTAMPTZ NOT NULL,
    indexed_at  TIMESTAMPTZ DEFAULT NOW(),
    checksum    TEXT
);
```

---

## 7. Phased Delivery

### Phase 1: Core API + Database + MCP Client
**Scope:** Stateless scripture search service. Identical functionality to current MCP servers, but hosted.

1. Initialize Go module in gospel-engine-v2 (single Dockerfile, no compose)
2. Write Dockerfile: Go builder stage → compiles both `gospel-engine` server and `gospel-mcp` for 4 platforms → alpine runtime
3. Create migration files with embedded SQL (`go:embed`), schema above, extensions: `CREATE EXTENSION vector; CREATE EXTENSION pg_trgm;`
4. Implement DB layer: connection pool, migration runner (runs on startup), queries
5. Pre-load gospel-library + books: `rsync` from Michael's machine to NOCIX `/opt/gospel/gospel-library/` and `/opt/gospel/books/` (in progress)
6. **Pre-compute embeddings on desktop using nomic-embed-text v1.5** (loaded into LM Studio with GPU acceleration on the 4090s — same model as server, just GPU-accelerated): export as JSONL with `(source_type, source_id, layer, embedding)` rows, rsync to NOCIX `/opt/gospel/embeddings/`
7. Implement indexing pipeline (reads mounted `/gospel-library`, parses markdown → PG, generates tsvector)
8. Implement embedding pipeline:
   - Bulk load: COPY pre-computed embeddings from `/data/embeddings/` into pgvector
   - Query-time: content → llmster `/v1/embeddings` → pgvector (only for new content + queries)
9. **Migrate TITSW + enrichment data** from existing gospel-engine SQLite into PG (script reads SQLite, writes to PG)
10. Implement search endpoints: `/api/search`, `/api/get`, `/api/list`, `/api/health` (with TITSW filters preserved)
11. Implement static download endpoints: `/download/gospel-mcp-{os}-{arch}`, `/api/version`
12. Add token auth middleware (validate `stdy_` tokens, simple rate limit)
13. Implement MCP client (`gospel-mcp`) that translates JSON-RPC → HTTP + self-update logic with safeguards (prior binary backup, opt-out env, SHA256 verify, "first successful run" gate before enabling auto-update)
14. Provision Dokploy services on NOCIX:
    - **Database service:** name `gospel-db`, image `pgvector/pgvector:pg18`, mount volume for `/var/lib/postgresql/data`, **enable scheduled backups (Dokploy built-in)**
    - **Application service:** name `gospel-engine`, connect to GitHub repo, build from Dockerfile, mount `/opt/gospel/gospel-library:/data/gospel-library:ro`, `/opt/gospel/books:/data/books:ro`, `/opt/gospel/embeddings:/data/embeddings:ro`, set env vars (`GOSPEL_DB`, `EMBEDDING_URL=http://host.docker.internal:1234/v1`, `EMBEDDING_MODEL=nomic-embed-text-v1.5`, etc.)
    - **Domain:** configure `study.ibeco.me` pointing at the app service (Dokploy handles TLS via Traefik)
15. Seed initial service token manually (direct DB insert or first-run bootstrap)
16. Verify: MCP client produces identical results to current gospel-engine MCP, including TITSW filters

**Deliverable:** `study.ibeco.me/api/search?q=faith` works. MCP client in `.vscode/mcp.json` replaces all three current MCP servers and auto-updates on redeploy.
**Estimate:** 4-6 sessions.

### Phase 2: Content Management + Graph
**Scope:** Self-managing content pipeline. Server handles incremental downloads for new content.

1. Port gospel-library downloader into backend (the bulk content is already pre-loaded; downloader handles new conference talks and manual updates)
2. Admin endpoints: trigger download, re-index, re-embed, check status
3. Scheduled indexing (April/October conference checks, quarterly manual check)
4. **If AGE is needed:** switch to custom PG image at this point (`pgvector/pgvector:pg18` base + AGE compiled in). Published to GHCR, Dokploy Database service updated to pull from there. Until then, pg18 image used directly.
5. Graph-aware search (cross-reference traversal via AGE, or recursive CTEs as fallback)

**Deliverable:** Server manages its own content. No manual rsync after initial pre-load.
**Estimate:** 2-3 sessions.

### Phase 3: Auth Delegation (ibeco.me integration)
**Scope:** ibeco.me provisions tokens for study.ibeco.me.

1. Add service token concept to study.ibeco.me (`is_service` flag)
2. Admin API: `POST /api/admin/tokens` (service token required)
3. Add endpoint to ibeco.me: "Get study access token" button
4. Rate limiting tiers: anonymous (strict), authenticated (generous)

**Deliverable:** ibeco.me users can get study.ibeco.me tokens with one click.
**Estimate:** 1-2 sessions.

### Phase 4: User Features (deferred)
**Scope:** Personal study features on study.ibeco.me.

- Search histories
- Notes / annotations
- Saved studies
- GitHub Copilot SDK integration (stretch)
- Frontend UI

**Estimate:** TBD — depends on scope decisions.

---

## 8. Verification Criteria

### Phase 1
- [ ] `curl https://study.ibeco.me/api/health` returns 200
- [ ] `curl -H "Authorization: Bearer stdy_..." https://study.ibeco.me/api/search?q=faith` returns hybrid search results
- [ ] MCP client configured in `.vscode/mcp.json` responds to `gospel_search`, `gospel_get`, `gospel_list`
- [ ] Search results match current gospel-engine output (same references, similar ranking)
- [ ] Unauthenticated requests are rejected (401)
- [ ] Query embedding latency < 200ms (including network round-trip)
- [ ] Database contains all 31K+ scripture verses and conference talks

### Phase 2
- [ ] Backend downloads new conference talks without manual intervention
- [ ] Admin can trigger re-index from API
- [ ] Content count matches gospel-library file count

### Phase 3
- [ ] ibeco.me user can generate a study.ibeco.me token
- [ ] Token works in MCP client
- [ ] Rate limiting enforced per-token

---

## 9. Costs & Risks

### Costs
- **Development time:** Phase 1 is substantial (4-6 sessions). New repo, new service, new deployment.
- **NOCIX resources:** PG + app + llmster ≈ 4GB RAM. Well within 32GB.
- **Docker image:** Phase 1 uses `pgvector/pgvector:pg18` directly (no custom image). Phase 2 may add a custom image only if AGE is needed.
- **Maintenance:** Running web service requires monitoring, backups, updates.
- **Complexity:** Three components (PG, backend, MCP client) vs current single binary.

### Risks
- **AGE on PG18:** Apache AGE's PG18 support may lag behind (PG18 is the latest stable). Mitigation: defer graph to Phase 2, build custom image only if AGE is needed. Phase 1 uses vanilla `pgvector/pgvector:pg18`.
- **llmster in Docker vs host:** LM Link needs host networking, but gospel-engine runs in Docker. Mitigation: llmster installed on host (NOCIX), app accesses via `host.docker.internal:1234` (works on Dokploy's Docker setup).
- **Church API rate limiting on incremental downloads:** Downloading new talks from NOCIX IP may trigger different rate limits. Mitigation: initial content is pre-loaded via rsync (Phase 1); only incremental downloads hit the API (~30 talks twice a year at 20 req/sec — well within limits).
- **Self-update on Windows:** Can't overwrite a running exe. Mitigation: spawn updater helper that waits for parent exit (well-known pattern; used by VS Code, Chrome, etc.).
- **Scope creep:** User features (Phase 4) could expand indefinitely. Mitigation: strict phasing. Phase 1 is stateless API only.
- **Token security:** Tokens over HTTPS are standard. bcrypt + prefix lookup is proven (same as ibeco.me). No novel security risks.

### What gets worse
- **Setup complexity:** Local dev now needs Docker + PG running (vs current SQLite file).
- **Network dependency:** MCP client needs internet to reach study.ibeco.me. Current MCP works offline.
- **Ops burden:** A web service needs monitoring, backups, SSL cert renewals (handled by Dokploy/Traefik).

### What gets better
- **Always available.** Gospel search works from any device, any network, any agent.
- **Single source of truth.** One database, one API, one deployment.
- **Shareable.** Other people could use the API (with tokens).
- **Web-ready.** ibeco.me, study.ibeco.me, future apps can all query the same service.
- **Replaces three MCP servers** with one thin client + one hosted backend.

---

## 10. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Make gospel search always-available and web-accessible, not trapped in local processes |
| Covenant | Rules? | Go + pgx + chi conventions. Docker/Dokploy deployment. Bearer token auth. Same MCP tool contract. |
| Stewardship | Who owns? | dev agent builds. Michael owns NOCIX ops and Dokploy deployment. |
| Spiritual Creation | Spec precise enough? | Schema defined. API endpoints defined. Auth flow defined. Repo structure defined. |
| Line upon Line | Phasing? | Phase 1 (stateless API) stands alone and replaces current MCP servers. Each phase adds value independently. |
| Physical Creation | Who executes? | dev agent, with Michael reviewing deployment and auth decisions. |
| Review | How to verify? | curl tests, MCP parity tests, search quality comparison. |
| Atonement | If wrong? | Current MCP servers still work. gospel-engine-v2 is a new repo — no risk to existing tools. Rollback is just re-enabling old MCP configs. |
| Sabbath | When stop? | After Phase 1 verification. Phases 2-4 are separate decisions. |
| Consecration | Who benefits? | Michael (study from anywhere), ibeco.me users, potentially anyone who wants gospel search. |
| Zion | Whole system? | Unifies gospel-mcp + gospel-vec + gospel-engine into one service. Becomes the data layer for study.ibeco.me, ibeco.me search, graph visualization, and any future gospel-related tool. |

---

## 11. Recommendation

**Build.** Phase 1 first.

This is the natural next step for the gospel tools. The PG migration proposal was already heading here — this just names the destination explicitly: study.ibeco.me as a hosted service, not a local process. The technical foundation (PG schema, pgvector, embedding strategy) is already researched and designed. What's new is the API layer, auth, and MCP client — all well-understood patterns already proven in ibeco.me.

Phase 1 is concrete and self-contained: a Go HTTP server, a PostgreSQL database, and a thin MCP client. When it's done, three MCP servers collapse into one hosted backend, and gospel search becomes permanently available from any device.

The deferred pieces (content management, auth delegation, user features) are genuinely separable. Phase 1 delivers full value without them.

**Execute with:** `@dev` agent
**Phase 1 scope:** 4-6 sessions
**First action:** Initialize Go module, write single Dockerfile (builds server + 4 MCP client binaries), create `go:embed`-ed migration files with `CREATE EXTENSION vector; CREATE EXTENSION pg_trgm;` targeting PG18

---

## 12. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-18 | New repo (gospel-engine-v2) instead of refactoring in-place | Clean start. Old MCP servers stay working until v2 is verified. |
| 2026-04-18 | Two binaries: gospel-engine (server) + gospel-mcp (client) | Server is heavy (PG, embedding). MCP client is thin (HTTP calls). Different deployment targets. |
| 2026-04-18 | Service token delegation (ibeco.me → study.ibeco.me) | Clean separation. study.ibeco.me is independent but ibeco.me manages user-facing token provisioning. |
| 2026-04-18 | `stdy_` token prefix | Distinct from `bec_` (ibeco.me). Same bcrypt + prefix lookup pattern. |
| 2026-04-18 | Defer AGE to Phase 2 (or later) | Latest PG (18) may not have AGE support yet. Core value is FTS + vector. |
| 2026-04-18 | Defer user features to Phase 4 | Stateless API is the foundation. User features are additive. |
| 2026-04-18 | LM Studio headless on host, not in Docker | LM Link requires host networking. Container accesses via `host.docker.internal`. |
| 2026-04-18 | Single Dockerfile, Dokploy builds on server | Matches Michael's existing Dokploy workflow. No intermediate registry. Dokploy Database service handles PG separately. |
| 2026-04-18 | PostgreSQL 18 (latest stable) via `pgvector/pgvector:pg18` | Latest stable. pgvector already builds an image for it. No custom image needed for Phase 1. |
| 2026-04-18 | MCP binaries hosted by the server (`/download/gospel-mcp-{os}-{arch}`) with self-update | Distribution + update channel in one. Users install once; subsequent updates are automatic on server redeploy. |
| 2026-04-18 | Gospel-library + books pre-loaded via rsync to `/opt/gospel/gospel-library/` and `/opt/gospel/books/` | Avoids initial Church API bulk download from NOCIX IP. Mounted read-only into container at `/data/`. Incremental updates in Phase 2. |
| 2026-04-18 | Embeddings pre-computed on desktop with **same** nomic-embed-text v1.5 model (just GPU-accelerated) | Embeddings cannot be mixed across models. Pre-compute uses identical model that NOCIX runs at query time, only difference is GPU vs CPU. NOCIX cannot rely on GPU access. |
| 2026-04-18 | TITSW + enrichment migrated from existing SQLite to PG | Years of enrichment work shouldn't be lost. One-time migration script in Phase 1. |
| 2026-04-18 | MCP self-update safeguards: prior binary backup, SHA256 verify, opt-out env, "first successful run" gate | Auto-update is risky. These prevent silent breakage from a bad release. |
| 2026-04-18 | Backups via Dokploy built-in PG backup | Free, integrated. Embeddings are expensive to regenerate (6-8h CPU). Backup is essential. |
