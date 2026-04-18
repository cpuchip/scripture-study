# study.ibeco.me — Hosted Gospel Study Service

**Binding problem:** Gospel search is trapped inside a local MCP server. It can't serve web clients, can't be shared, and dies when the desktop goes offline. There's no way for ibeco.me, study.ibeco.me, remote agents, or other users to search scriptures without running the full gospel-engine binary locally. Moving to a hosted service at study.ibeco.me would make gospel search a permanent, always-available API — accessible from MCP clients, web browsers, and any HTTP client — while keeping the database self-hosted on NOCIX.

**Created:** 2026-04-18
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
2. **PostgreSQL with pgvector + FTS** stores all content and embeddings in one database
3. **Embedding generation** runs on NOCIX via LM Studio headless (nomic-embed-text, CPU). LM Link to desktop GPU is optional acceleration.
4. **Gospel-library downloads** are managed server-side — the backend pulls content from the Church API, indexes, and embeds it
5. **Token-based auth** protects the API from abuse. Tokens provisioned via ibeco.me (service token delegation)
6. **A new MCP client** (`gospel-mcp`) translates MCP JSON-RPC to HTTP calls against study.ibeco.me
7. **Deployed on Dokploy** alongside ibeco.me on the NOCIX server
8. **Current MCP tools are preserved** — agents using `gospel_search`, `gospel_get`, `gospel_list` see identical behavior through the new MCP client

---

## 3. Scope

### In scope (Phase 1 — Core API)
- Go backend HTTP server with chi router
- PostgreSQL schema (scriptures, talks, embeddings, cross-references) — from PG migration proposal
- Custom Docker image (pgvector + pg_trgm on PG17)
- REST endpoints: `/api/search`, `/api/get/{ref}`, `/api/list`, `/api/health`
- Token validation middleware (bearer tokens, `stdy_` prefix)
- Content indexing pipeline (parse gospel-library markdown → PG)
- Embedding pipeline (content → llmster → pgvector)
- docker-compose.yml for Dokploy deployment
- MCP client binary that wraps the HTTP API

### In scope (Phase 2 — Content Management)
- Gospel-library download module (port from `scripts/gospel-library/`)
- Automatic indexing of new content
- Admin endpoints for triggering re-index / re-embed
- Apache AGE graph extension (if PG17 build is stable)

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
- Apache AGE graph queries (move to Phase 2 if PG17 build is problematic)

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
- Docker deployment via Dokploy
- Bearer token auth: `stdy_` prefix + 64 hex chars, bcrypt hashed
- Environment variables for all config (same pattern as ibeco.me)

---

## 4. Prior Art

| Source | Relevance |
|--------|-----------|
| ibeco.me (scripts/becoming/) | Proven Dokploy deployment pattern. Chi router. PG backend. Token auth with bcrypt + prefix lookup. Google OAuth. Docker-compose for Dokploy. |
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
│   │   └── admin.go           # /api/admin/* handlers
│   ├── auth/                  # Token validation + rate limiting
│   │   ├── middleware.go
│   │   └── tokens.go
│   ├── db/                    # PostgreSQL access layer
│   │   ├── db.go              # Connection, migrations
│   │   ├── scriptures.go      # Scripture queries
│   │   ├── talks.go           # Conference talk queries
│   │   ├── search.go          # Hybrid search queries
│   │   └── tokens.go          # Token storage
│   ├── embed/                 # Embedding client (port embedder.go)
│   │   └── embedder.go
│   ├── index/                 # Content indexing pipeline
│   │   └── indexer.go
│   └── search/                # Search orchestration
│       └── hybrid.go          # FTS + vector + RRF merging
├── docker/
│   └── gospel-db/
│       ├── Dockerfile         # PG17 + pgvector + pg_trgm
│       └── init.sql           # Extension creation
├── migrations/
│   ├── 001_schema.sql
│   └── 002_indexes.sql
├── docker-compose.yml         # Dokploy deployment
├── Dockerfile                 # App server image
├── go.mod
├── go.sum
└── README.md
```

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
    content     TEXT NOT NULL,        -- the text that was embedded
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

1. Initialize Go module in gospel-engine-v2
2. Build custom PG Docker image (pgvector + pg_trgm on PG17)
3. Write docker-compose.yml (PG + app)
4. Create migration files (schema above, minus AGE)
5. Implement DB layer: connection pool, migrations, queries
6. Port indexing pipeline from gospel-engine (parse markdown → PG, generate tsvector)
7. Port embedding pipeline (content → llmster `/v1/embeddings` → pgvector)
8. Implement search endpoints: `/api/search`, `/api/get`, `/api/list`
9. Add token auth middleware (validate `stdy_` tokens, rate limit)
10. Seed initial token manually (for Michael's MCP use)
11. Write MCP client (`gospel-mcp`) that translates JSON-RPC → HTTP
12. Deploy to Dokploy, configure study.ibeco.me domain
13. Index existing gospel-library content on NOCIX
14. Verify: MCP client produces identical results to current gospel-engine MCP

**Deliverable:** `study.ibeco.me/api/search?q=faith` works. MCP client in `.vscode/mcp.json` replaces all three current MCP servers.
**Estimate:** 4-6 sessions.

### Phase 2: Content Management + Graph
**Scope:** Self-managing content pipeline. Server downloads and indexes gospel-library.

1. Port gospel-library downloader into backend
2. Admin endpoints: trigger download, re-index, re-embed, check status
3. Scheduled indexing (new conference talks, manual updates)
4. Add Apache AGE if PG17 build is stable
5. Graph-aware search (cross-reference traversal)

**Deliverable:** Server manages its own content. No manual file copying.
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
- **Docker image:** Custom PG image needs building and hosting (GHCR, free for public repos).
- **Maintenance:** Running web service requires monitoring, backups, updates.
- **Complexity:** Three components (PG, backend, MCP client) vs current single binary.

### Risks
- **AGE on PG17:** Apache AGE PG17 support may be unstable. Mitigation: defer to Phase 2, start without graph.
- **llmster in Docker vs host:** LM Link needs host networking, but gospel-engine runs in Docker. Mitigation: `host.docker.internal` or install llmster on host, access from Docker container.
- **Church API rate limiting:** Downloading gospel-library from NOCIX (server IP) may trigger different rate limits than desktop. Mitigation: same 20 req/sec limit, respectful User-Agent.
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
**First action:** Initialize Go module, build custom PG Docker image, write docker-compose.yml, create schema migrations

---

## 12. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-18 | New repo (gospel-engine-v2) instead of refactoring in-place | Clean start. Old MCP servers stay working until v2 is verified. |
| 2026-04-18 | Two binaries: gospel-engine (server) + gospel-mcp (client) | Server is heavy (PG, embedding). MCP client is thin (HTTP calls). Different deployment targets. |
| 2026-04-18 | Service token delegation (ibeco.me → study.ibeco.me) | Clean separation. study.ibeco.me is independent but ibeco.me manages user-facing token provisioning. |
| 2026-04-18 | `stdy_` token prefix | Distinct from `bec_` (ibeco.me). Same bcrypt + prefix lookup pattern. |
| 2026-04-18 | Defer AGE to Phase 2 | PG17 AGE build may be unstable. Core value is FTS + vector, not graph. |
| 2026-04-18 | Defer user features to Phase 4 | Stateless API is the foundation. User features are additive. |
| 2026-04-18 | LM Studio headless on host, not in Docker | LM Link requires host networking. Container accesses via host.docker.internal. |
