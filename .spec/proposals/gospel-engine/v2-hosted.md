---
workstream: WS3
status: shipped
brain_project: 3
created: 2026-04-21
last_updated: 2026-04-21
---

# gospel-engine v2 — Hosted Backend (engine.ibeco.me)

**Status:** **SHIPPED** 2026-04-20 (Phases 1–3)
**Created (this file):** 2026-04-21
**Lineage:** Started life as Phases 1–3 of [`study-ibeco-me/main.md`](../study-ibeco-me/main.md) and PG migration of [`gospel-engine-postgresql/main.md`](../gospel-engine-postgresql/main.md). When the backend shipped and the UI roadmap was carved off into its own scope, this file was created as the canonical home for the hosted-engine work.

---

## What it is

A hosted gospel-search backend at **`engine.ibeco.me`**. Same capabilities as the local `gospel-engine` (v1) — keyword (FTS), semantic (vector), combined, content retrieval, listing — but running on NOCIX, PostgreSQL + pgvector, accessible to any HTTP client with a bearer token.

A thin MCP client (`gospel-mcp.exe`) translates MCP JSON-RPC ↔ HTTPS calls. It self-updates on startup by checking `engine.ibeco.me/api/version` and downloading new binaries from `engine.ibeco.me/download/gospel-mcp-{os}-{arch}` (with SHA256 verify + rollback safety).

## What shipped (Phases 1–3)

| Phase | Scope | Where |
|-------|-------|-------|
| **Phase 1 — Core API** | Go HTTP server (chi), PostgreSQL 18 + pgvector + pg_trgm, REST endpoints (`/api/search`, `/api/get/{ref}`, `/api/list`, `/api/health`), bearer-token middleware (`stdy_` prefix, bcrypt + prefix lookup), content indexing pipeline, MCP client binary with self-update, single-Dockerfile deployment to Dokploy on NOCIX. | `scripts/gospel-engine-v2/` |
| **Phase 2 — Content Management** | Gospel-library + books pre-loaded to `/opt/gospel/` on NOCIX (mounted read-only). Embeddings pre-computed on desktop using `nomic-embed-text v1.5` (768-dim) and rsynced to NOCIX; bulk-loaded via PG `COPY` on first run. Server-side query-time embedding via LM Studio headless / OpenAI-compatible endpoint on CPU. | `scripts/gospel-engine-v2/` + `/opt/gospel/` on NOCIX |
| **Phase 3 — Auth Delegation** | ibeco.me `/api/engine-tokens/{status,list,create,revoke}` endpoints. Internal client at `scripts/becoming/internal/engine/handlers.go`. Vue Settings UI (`SettingsView.vue` line 304+) lets a logged-in ibeco.me user mint a study token. First user-minted token used Apr 20 for the "Give Away All My Sins" study. | `scripts/becoming/` |

## What's deferred

- **Apache AGE / graph traversal** — depends on AGE supporting PG18. Tracked in [`../gospel-graph/main.md`](../gospel-graph/main.md).
- **Anonymous tier with strict rate limits** — only authenticated tokens for now.
- **Admin endpoints for re-index / re-embed via API** — currently a manual rsync + restart.
- **Auto-discovery of new conference talks** — planned for an incremental indexer; for now, new content is added by hand on the desktop and rsynced.

## MCP integration (currently active)

`.vscode/mcp.json` registers exactly one gospel server:

```jsonc
"gospel-engine-v2": {
  "command": ".../scripts/gospel-engine/gospel-mcp.exe",
  "env": {
    "GOSPEL_ENGINE_URL": "https://engine.ibeco.me",
    "GOSPEL_ENGINE_TOKEN": "stdy_…",
    "GOSPEL_AUTO_UPDATE": "true"
  },
  "type": "stdio"
}
```

Tool names visible to agents: `mcp_gospel-engine_gospel_search`, `mcp_gospel-engine_gospel_get`, `mcp_gospel-engine_gospel_list` (note: VS Code strips the `-v2` suffix from the server name when generating function names). The local v1 MCP server (`scripts/gospel-engine/`) and the legacy `gospel-mcp` / `gospel-vec` servers are no longer registered; they remain on disk as fallback only.

## Architecture (as deployed)

```
┌─────────────── NOCIX (Ryzen 3800X, 32GB, no GPU) ──────────────┐
│                                                                 │
│  Dokploy:                                                        │
│    ├─ Database service:  pgvector/pgvector:pg18 (PG + pgvector + pg_trgm)
│    └─ Application service: gospel-engine-v2 (chi HTTP server)    │
│                                                                  │
│  Volumes (read-only mount into app container):                   │
│    /opt/gospel/gospel-library/  → /data/gospel-library           │
│    /opt/gospel/books/           → /data/books                    │
│    /opt/gospel/embeddings/      → /data/embeddings               │
│                                                                  │
│  Embedding (server, query-time):                                 │
│    LM Studio headless (CPU) · nomic-embed-text v1.5 (768-dim)    │
└─────────────────────────────────────────────────────────────────┘
              │ HTTPS (Bearer stdy_…)
              ▼
   ┌──────────────────────────┐    ┌─────────────────────────┐
   │ gospel-mcp.exe           │    │ ibeco.me                │
   │ (MCP client, stdio↔HTTP, │    │ (mints stdy_ tokens     │
   │  self-update)            │    │  via /api/engine-tokens)│
   └──────────────────────────┘    └─────────────────────────┘
```

## Schema highlights (carried forward from PG migration spec)

- `scriptures`, `chapters`, `talks`, `manuals`, `books` — content tables with `content_tsvector` for FTS and `embedding vector(768)` for pgvector
- `cross_references` — bidirectional graph edges (footnote-derived) with HNSW index on the embedding column
- `tokens` — bcrypt-hashed, `stdy_` prefix lookup column, `revoked_at`, `last_used_at`
- TITSW columns on conference talks: `titsw_mode`, `titsw_pattern`, `titsw_dominant_dimensions`, six dimension scores (0–9)
- `chapter_lenses` — enriched scripture summaries from the lens-approach pipeline

Hybrid search uses `pgvector` cosine distance + `tsquery` ranked together with RRF.

## Lessons that landed in the harness

- **Embeddings are not portable across models.** Switching from `nomic-embed-text v1.5` requires full re-embed. The model is fixed on both desktop (GPU pre-compute) and NOCIX (CPU query-time).
- **Pre-loading content + embeddings via rsync** beats hammering the Church API from a server IP. The initial corpus moved over a single SSH connection.
- **Self-update with rollback safety** — `<self>.prev` backup + SHA256 verify + "first successful run" gate. The MCP client survives a bad release.
- **Token UX matters.** The Apr 20 study only happened because the Vue Settings page lets the user mint a token in two clicks. Without that UI, every new device would have been a manual SSH session.

## Pointers

- **Code:** `scripts/gospel-engine-v2/` (server) and `scripts/gospel-engine/gospel-mcp.exe` (client binary, built from same repo)
- **Repo:** [github.com/cpuchip/gospel-engine](https://github.com/cpuchip/gospel-engine)
- **Token UI:** `scripts/becoming/frontend/src/views/SettingsView.vue` (line 304+)
- **Token API (ibeco.me):** `scripts/becoming/cmd/server/main.go` + `scripts/becoming/internal/engine/handlers.go`
- **First study using it:** [`study/give-away-all-my-sins.md`](../../../study/give-away-all-my-sins.md) (or the dated variant) — Apr 20
- **Original combined spec (historical):** [`../study-ibeco-me/main.md`](../study-ibeco-me/main.md) (pre-refocus body)
- **Original PG migration research:** [`../gospel-engine-postgresql/main.md`](../gospel-engine-postgresql/main.md)

## What this file deliberately does NOT cover

- The user-facing study site (UI, histories, notes, annotations) — that's [`../study-ibeco-me/main.md`](../study-ibeco-me/main.md).
- Graph traversal / Cypher queries — that's [`../gospel-graph/main.md`](../gospel-graph/main.md).
- Local-only v1 ergonomics — that's [`phase1.5-ergonomics.md`](phase1.5-ergonomics.md).
- Future engine work (LoRA fine-tuning, proxy-pointer RAG, etc.) — file new proposals under `gospel-engine/` for each.
