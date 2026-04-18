# study.ibeco.me — Research & Findings

**Binding problem:** Gospel search needs to be a hosted web service at study.ibeco.me — not just a local MCP server. The current gospel-engine is a single-process stdio MCP that can't serve web clients, can't share state between users, and requires the desktop to be online for semantic search. The vision is a three-part architecture: (1) PostgreSQL with extensions on Dokploy, (2) Go backend service managing downloads and API, (3) new MCP server that connects to the backend over the network.

**Created:** 2026-04-18
**Related:** [gospel-engine-postgresql proposal](../../proposals/gospel-engine-postgresql/main.md) (PG schema and embedding strategy carry forward), [gospel-graph proposal](../../proposals/gospel-graph/main.md) (graph viz frontend)

---

## Relationship to Existing Proposal

The [gospel-engine-postgresql proposal](../../proposals/gospel-engine-postgresql/main.md) planned a local migration — same process, different storage backend. This new vision **subsumes and extends** it:

### What carries forward (reuse directly)
- PostgreSQL schema design (scriptures, talks, embeddings, cross_references, edges)
- pgvector + Apache AGE + FTS extension stack
- LM Studio headless (llmster) + nomic-embed-text for CPU embedding
- LM Link for desktop GPU acceleration
- HNSW index strategy
- Hybrid search query patterns (FTS + vector + RRF)
- Docker image for PG with extensions

### What changes
- **Separate repo** (`gospel-engine-v2`) instead of refactoring in-place
- **HTTP API is primary**, not stdin/stdout MCP
- **MCP becomes a thin client** that calls the HTTP API — not a monolith
- **Auth/token system** needed for public API
- **Gospel-library downloads move server-side** — the backend manages its own content
- **User-facing features** (search histories, notes, studies) need user management
- **Two-tier API**: stateless scripture search (token for abuse protection) + stateful user features (full auth)

### What the old proposal missed
- No auth story at all (assumed single-user)
- No separation of MCP client from service
- No user-facing web features
- No content management (assumed gospel-library exists locally)

---

## Architecture Research

### Three Components

**1. PostgreSQL + Extensions (Dokploy database service)**

Custom Docker image needed. Dokploy supports database services but the standard `postgres:17-alpine` doesn't include pgvector or AGE. Options:

- **Build custom image** with pgvector + AGE + pg_trgm. Push to GitHub Container Registry (ghcr.io/cpuchip/gospel-db). Dokploy can pull from there.
- **Use `pgvector/pgvector:pg17` as base**, add AGE extension on top. pgvector image already includes pg_trgm (part of contrib).
- AGE 1.7.0 needs to be compiled or installed from their PG17 release.

Dokploy database service pattern:
- Create as "Database" type in project (not Application)
- Or include in docker-compose alongside the app (like ibeco.me does)
- Connection string via env var: `GOSPEL_DB=postgres://gospel:$PW@db:5432/gospel?sslmode=disable`

**2. Go Backend Service (the API server)**

Responsibilities:
- Serve REST/HTTP API for search, get, list (same capabilities as current MCP tools)
- Manage gospel-library downloads (using the existing gospel-library downloader logic, or calling the Church API directly)
- Index and embed new content into PostgreSQL
- Expose endpoints for user features (search history, notes, studies — Phase 2+)
- Token validation middleware

Architecture pattern from ibeco.me:
- Chi router (`go-chi/chi/v5`)
- Middleware chain: logging → CORS → auth (optional/required per route)
- Health check endpoint
- Graceful shutdown
- Docker deployment via Dokploy

**3. MCP Client (thin wrapper over HTTP API)**

This is the key architectural shift. The MCP server becomes a **client** of the backend:

```
VS Code agent → MCP server (local process, stdio) → HTTP API (study.ibeco.me) → PostgreSQL
```

The MCP server is a thin translation layer:
- Receives JSON-RPC calls over stdin/stdout
- Translates them to HTTP requests to study.ibeco.me
- Returns results in MCP format
- Carries a bearer token for auth

This could live in the same repo or be a separate small package. Pros of same repo: shared types. Cons: MCP binary ships to users' machines, shouldn't need the full backend.

**Recommendation:** Same repo, separate `cmd/` entry point. Build produces two binaries: `gospel-engine` (server) and `gospel-mcp` (MCP client).

---

## Auth & Token Architecture

### The Two Sides

Michael identified two distinct auth needs:

**Side 1 — Stateless scripture API (search, get, list)**
- Read-only access to public scripture data
- Token needed only for abuse protection (rate limiting, DOS prevention)
- No user identity needed — just "is this a valid client?"
- Could use simple API keys or bearer tokens
- Similar to how APIs use API keys for rate limiting without user accounts

**Side 2 — Stateful user features (search histories, notes, studies)**
- Needs user identity
- User accounts with login
- Personal data requires authorization
- Full auth: who are you + what can you access?

### Token Flow Options

**Option A: ibeco.me as OAuth provider for study.ibeco.me** ⭐ LIKELY BEST
- ibeco.me already has: user accounts, Google OAuth, session management, API tokens (`bec_` prefix)
- ibeco.me could issue tokens scoped to study.ibeco.me
- Flow: User logs into ibeco.me → clicks "Get study.ibeco.me access" → ibeco.me generates a `study_` token → user configures MCP/CLI with that token → study.ibeco.me validates by calling ibeco.me's token verification endpoint
- Pros: Single user account across both services. Existing auth infrastructure. User management stays in one place.
- Cons: study.ibeco.me depends on ibeco.me for token validation (or needs to cache/replicate tokens)

**Option B: Shared database**
- Both ibeco.me and study.ibeco.me read from the same PostgreSQL user/token tables
- study.ibeco.me validates tokens directly against the DB
- Pros: No inter-service calls for validation. Fast.
- Cons: Schema coupling. Both services need access to same DB.

**Option C: study.ibeco.me has its own auth**
- Separate user accounts, separate tokens
- Pros: Fully independent
- Cons: Users manage two accounts. Duplicates work.

**Option D: Service token + user token (Michael's suggestion)**
- ibeco.me holds a service token for study.ibeco.me
- ibeco.me can provision user tokens on study.ibeco.me on behalf of the user
- Flow: User logs into ibeco.me → ibeco.me calls study.ibeco.me's admin API (using service token) → study.ibeco.me creates user token → returns to ibeco.me → ibeco.me shows token to user
- Pros: Clean separation. study.ibeco.me is independent but delegated. Similar to how cloud APIs work.
- Cons: More moving parts. ibeco.me needs a new endpoint. study.ibeco.me needs admin API.

### Recommendation: Option D (Service Token Delegation)

This matches Michael's vision most closely and is the cleanest separation:

1. **study.ibeco.me** has its own token table. Tokens prefixed `stdy_`.
2. **ibeco.me** holds a privileged service token for study.ibeco.me (created manually once).
3. When a user on ibeco.me wants a study token:
   - ibeco.me calls `POST study.ibeco.me/api/admin/tokens` with service token
   - study.ibeco.me creates a user token, returns it
   - ibeco.me shows the token to the user (shown once, just like current `bec_` tokens)
4. User configures their MCP client with the `stdy_` token.
5. study.ibeco.me validates tokens locally — no calls back to ibeco.me at runtime.

**For the stateless API:** Same tokens, but rate-limited. Or a separate "anonymous" tier with stricter rate limits for unauthenticated requests.

---

## ibeco.me Existing Auth Patterns (for reuse)

From `scripts/becoming/internal/auth/` and `scripts/becoming/internal/db/tokens.go`:

- **Token format:** `bec_` prefix + 64 hex chars (32 random bytes)
- **Storage:** bcrypt hash in DB, prefix (first 12 chars) for fast lookup
- **Validation:** Prefix-based narrowing → bcrypt compare → expiry check
- **Middleware chain:** Session cookie → Bearer token → dev mode fallback
- **Token metadata:** name, created_at, last_used, expires_at, per-token feature flags
- **Admin check:** `ADMIN_EMAILS` env var (comma-separated)

This pattern can be directly ported to study.ibeco.me with minimal changes:
- Change prefix from `bec_` to `stdy_`
- Add service token concept (a token that can create other tokens)
- Add rate limiting middleware (per-token or per-IP)

---

## Dokploy Deployment Structure

Based on ibeco.me pattern:

```yaml
# Dokploy project: study.ibeco.me
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      GOSPEL_DB: postgres://gospel:${POSTGRES_PASSWORD}@db:5432/gospel?sslmode=disable
      GOSPEL_PORT: ":8080"
      EMBEDDING_URL: http://localhost:1234/v1  # or llmster address
      EMBEDDING_MODEL: nomic-embed-text
      SERVICE_TOKEN_HASH: ${SERVICE_TOKEN_HASH}  # for ibeco.me delegation
    depends_on:
      db:
        condition: service_healthy

  db:
    build:
      context: ./docker/gospel-db
      # Custom image: postgres + pgvector + AGE + pg_trgm
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: gospel
      POSTGRES_USER: gospel
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gospel"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
    driver: local
```

**Custom PG Docker image** would live at `docker/gospel-db/Dockerfile` in the gospel-engine-v2 repo. Published to GHCR for Dokploy to pull.

---

## Repo Structure (gospel-engine-v2)

```
gospel-engine-v2/
├── cmd/
│   ├── gospel-engine/     # HTTP server binary
│   │   └── main.go
│   └── gospel-mcp/        # MCP client binary (talks to HTTP API)
│       └── main.go
├── internal/
│   ├── api/               # HTTP handlers (search, get, list, admin)
│   ├── auth/              # Token validation, middleware, rate limiting
│   ├── db/                # PostgreSQL queries, migrations
│   ├── embed/             # Embedding client (existing embedder.go pattern)
│   ├── index/             # Content indexing pipeline
│   ├── download/          # Gospel-library downloader (port from gospel-library)
│   └── search/            # Hybrid search logic (FTS + vector + graph)
├── docker/
│   └── gospel-db/         # Custom PG image with extensions
│       └── Dockerfile
├── migrations/            # SQL migration files
├── docker-compose.yml     # Local dev + Dokploy deployment
├── Dockerfile             # App server image
├── go.mod
└── README.md
```

---

## Embedding Strategy on NOCIX

This carries forward from the PG migration proposal:

- **LM Studio headless (llmster)** installed directly on NOCIX (not in Docker — needed for LM Link)
- **nomic-embed-text v1.5** for CPU embedding (always available)
- **LM Link** to desktop for GPU acceleration (optional)
- **API:** `http://localhost:1234/v1/embeddings` — same OpenAI-compatible format

The Go backend calls this just like the current embedder.go. Config: `EMBEDDING_URL` + `EMBEDDING_MODEL` env vars.

**Important:** llmster runs on the host, not in Docker. Gospel-engine (in Docker) accesses it via `host.docker.internal:1234` or Docker's host network mode.

---

## What Replaces What

| Current | Replaced By | Notes |
|---------|-------------|-------|
| gospel-mcp (scripts/gospel-mcp/) | study.ibeco.me backend | FTS search moves to PG |
| gospel-vec (scripts/gospel-vec/) | study.ibeco.me backend | Vector search moves to pgvector |
| gospel-engine (scripts/gospel-engine/) | study.ibeco.me backend | Unified service |
| gospel-library (scripts/gospel-library/) | gospel-engine-v2 download module | Ported into backend |
| Local MCP (stdin/stdout) | gospel-mcp (HTTP client) | Thin wrapper over API |

---

## Custom PG Docker Image

Need: PostgreSQL 17+ with pgvector + Apache AGE + pg_trgm.

```dockerfile
FROM pgvector/pgvector:pg17

# AGE requires building from source for PG17
RUN apt-get update && apt-get install -y \
    build-essential \
    libreadline-dev \
    zlib1g-dev \
    flex \
    bison \
    git \
    && rm -rf /var/lib/apt/lists/*

# Build Apache AGE
RUN git clone --branch release/PG17/1.7.0 https://github.com/apache/age.git /tmp/age \
    && cd /tmp/age \
    && make install \
    && rm -rf /tmp/age

# pg_trgm is already in contrib (included in pgvector base image)

# Init script to create extensions
COPY init.sql /docker-entrypoint-initdb.d/
```

`init.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS age;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
LOAD 'age';
SET search_path = ag_catalog, "$user", public;
```

Push to: `ghcr.io/cpuchip/gospel-db:pg17`

---

## Deferred: Study Website Features

Michael mentioned these as stretch goals:
- **Studies** — store and share study documents
- **Search histories** — track what users search for
- **Notes** — personal annotations on scriptures/talks
- **GitHub Copilot SDK integration** — AI-powered study from the web

These all need user management (Side 2). Defer to after the core API is working.

---

## Open Questions

1. **Gospel-library download on server:** Does the backend download gospel-library at startup? On a schedule? On-demand? The current downloader hits the Church's API with rate limiting (20 req/sec). Running this server-side means the Church's API gets hit from NOCIX, not the desktop.

2. **AGE on PG17:** The Apache AGE PG17 release exists but may need testing. If it's unstable, defer graph to Phase 2 and use recursive CTEs for now.

3. **Rate limiting specifics:** How many requests/min for anonymous? For authenticated? Per-token? Need concrete numbers.

4. **MCP transport:** Standard MCP is stdin/stdout. The new MCP client translates to HTTP. Could also support SSE (Server-Sent Events) for streaming results — MCP 2025 spec supports SSE transport.

5. **Docker host networking for llmster:** If gospel-engine runs in Docker but llmster runs on the host, need `host.docker.internal` or `--network host`. Test on NOCIX.
