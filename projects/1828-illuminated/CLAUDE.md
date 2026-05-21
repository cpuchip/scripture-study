# 1828-illuminated — Claude Code project context

Public-facing reading-frame tool. Renders the standard works with a 1828
Webster meaning lens. Three-container deploy at 1828.ibeco.me (frontend
nginx, Go backend, Postgres). Backend pivot ratified 2026-05-20 — see
`.spec/proposals/backend-pivot.md` and siblings for the canonical plan.

## Where things live

| Need | Path |
|------|------|
| Intent + values | [`intent.yaml`](intent.yaml) |
| Backend pivot proposals (canonical) | [`.spec/proposals/`](.spec/proposals/) |
| Parent proposal (workspace-level) | [`../../.spec/proposals/1828-illuminated-scriptures.md`](../../.spec/proposals/1828-illuminated-scriptures.md) |
| Word-list research | [`../../research/gospel/1828/`](../../research/gospel/1828/) |
| Frontend source | `frontend/src/` |
| Backend source (Go) | `backend/` (created at phase 1) |
| Embedded seed data | `backend/internal/seed/data/` (scripture zip, 1828 gz, tier-words JSON, modern-defs JSON) |
| Local data (legacy bundle) | `frontend/src/data/` — kept as canonical seed source for DB rebuilds |
| Fetch / build scripts | `scripts/` |
| Per-session journals | `.spec/journal/` |

## Build + deploy

Three containers under one compose file:

- **frontend** — `nginx:1.27-alpine` serving Vue dist (built by `node:22-alpine`)
- **backend** — Go binary on `gcr.io/distroless/static:nonroot`, embeds scripture corpus + 1828 dict + seed data via `//go:embed`
- **db** — `postgres:17-alpine`, named volume `pg-data`

Local:

```
cp .env.example .env       # fill POSTGRES_PASSWORD, LLM_*, etc.
docker compose up -d
```

Production: Dokploy's Compose project type, env vars set in Dokploy UI. The
old single-Dockerfile path stays as `Dockerfile.legacy` for one deploy cycle
as a rollback escape hatch.

⚠️ **Never run `docker compose down -v` in production.** That wipes the
named volume. We learned this in pg-ai-stewards. Use `docker compose down`
without the `-v` flag.

## Conventions

- **Scripture text bundled (verse-only).** The bcbooks/scriptures-json
  2013-edition corpus is embedded in the backend image. On ingest, footnote
  markers, chapter/section headings, and study-aid apparatus are stripped —
  only verse text + verse numbering survives. Posture: home / personal-study
  fair use of public-domain verse text with bcbooks cited as source.
  (D-BE-COPYRIGHT option D, ratified 2026-05-20.)
- **Always provide tabbed-iframe breakout.** Every rendered scripture surface
  shows a small `↗` affordance that opens the same passage at
  churchofjesuschrist.org for the full canonical apparatus. Pattern source:
  `projects/cpuchip.net/src/components/ScripturePanel.vue`.
- **Backend exists.** It owns scripture corpus + dictionary serving + LLM
  proxy + Thummim cross-read. Frontend is still a static Vue SPA; it just
  fetches from `/api/*` instead of importing JSON at build time.
- **Stack:** Vue 3 + Vite 8 + Tailwind + vue-router (frontend); Go 1.23 +
  pgx + chi (backend, conventions match `scripts/becoming/` and
  `external_context/scriptures-mcp/`).
- **LM Studio is for embeddings only.** When pgvector arrives in a future
  phase, the 1828 backend will reach LM Studio on the Dokploy host via the
  host-network tunnel. LM Studio is NOT used for verse rendering — that's
  OpenAI / OpenRouter / opencode-go / opencode-zen via the BYOK proxy.
  gospel-engine-v2's LM Studio pipeline is independent and unaffected.
- **BYOK + session-key for LLM render.** Reader pastes their provider key
  in Settings; backend mints a 24h sliding-TTL in-memory session; subsequent
  /api/llm/render calls carry the session_id cookie. Keys never touch disk
  or DB; sessions die on backend restart. (D-LP-2, ratified 2026-05-20.)
- **Rate-limit errors attributed to us.** When 10/min or 1000/day caps
  trigger, the response body explicitly identifies "rate_limited_by_1828"
  so readers don't blame their provider. Upstream provider errors get
  passed through unchanged.

## Data flow (post-pivot)

```
Build-time seeds (live in repo, embedded into backend image):
  research/gospel/1828/00-FINAL-highlight-candidates.md (tier list)
    → scripts/build-data.py extracts tier-A/B/C words
    → scripts/fetch_modern_defs.py adds modern definitions (1/sec)
    → frontend/src/data/{tier-words,definitions-1828,definitions-modern}.json
    → backend/internal/seed/data/* (copied at build time)

Boot-time ingest (backend reads seeds → Postgres):
  bcbooks/scriptures-json (2013 ed., zip) → scripture_books/chapters/verses
    (footnotes + headings stripped on ingest)
  webster1828.json.gz → webster_1828 (~98k entries)
  tier-words.json → tier_words
  definitions-modern.json → modern_defs (seed for the lazy cache)

Runtime:
  Frontend → GET /api/scripture/:ref?highlight=1 → backend → DB → JSON
  Frontend → GET /api/dict/1828/:word         → backend → DB → JSON
  Frontend → GET /api/dict/modern/:word       → backend → DB (or lazy-fetch + write-back)
  Frontend → POST /api/llm/session            → backend mints session
  Frontend → POST /api/llm/render             → backend → upstream LLM (BYOK)
  Frontend → GET /api/scripture/search?q=…    → backend → tsvector + trgm

Nightly (separate job):
  pg-ai-stewards thummim_entries → /api/thummim/* snapshot → DB
```

## Stewardship

Per workspace covenant (`agent_commits_to`), agent has stewardship over the
code within Michael's intent. Per Michael's 2026-05-20 directive on this
project specifically: **if you see neighboring bugs, fix them — don't just
name them**, even when the bug isn't on the path of the feature you're
shipping. The boundary test stays the same (would Michael, asked in advance,
say "yes obviously do that"?), but the default leans toward acting.

End-of-session protocol:

1. Journal entry to `.spec/journal/YYYY-MM-DD-short-title.md`
2. Update `intent.yaml` if stretch goals advance or constraints change
3. Update `.mind/active.md` (workspace-level) when project state shifts materially
4. Commit + describe what shipped, name what remains

Honest cautions held throughout (from the parent proposal §IV):

- 1828 isn't always deeper; curation matters more than data
- Decoder-ring posture is the failure mode to avoid
- Good-faith reads still differ
- The hybrid copyright posture (D-BE-COPYRIGHT option D) is a measured fair-use
  position, not a permission slip. If the LDS Church's IP department ever
  reaches out, we strip to PD-only sources cooperatively and quickly.
