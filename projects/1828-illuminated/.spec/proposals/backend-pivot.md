---
title: 1828-illuminated — backend pivot (umbrella)
date: 2026-05-20
status: proposed
workstream: WS7
parent:
  - ../../../../.spec/proposals/1828-illuminated-scriptures.md
  - ../../intent.yaml
children:
  - scripture-corpus.md
  - dictionary-backend.md
  - llm-proxy.md
  - deployment-shape.md
purpose: >
  Add a small Postgres-backed API service to the 1828-illuminated project,
  hosted as a sibling container alongside the existing nginx frontend.
  Solves: LLM CORS, key safety, lazy modern-def expansion, verse explorer
  rendering local public-domain scripture text instead of an iframe, and
  the "every 1828 entry queryable" reach for tier-D / class-E words.
---

# 1828-illuminated — Backend Pivot

## I. The decision Michael made (2026-05-20)

Pivot from "static SPA with bundled JSON" to **frontend + small backend + Postgres**, all three as siblings under one `docker-compose.yaml` at the project root. Frontend stays static (Vue → nginx). Backend is a small Go service. DB is its own Postgres container with a named volume.

The MVP shipped overnight (commit `1828_illuminated_scriptures` 2026-05-20) honors the old constraints: no backend, JSON bundled, scripture text iframed. Real-world contact with the tool surfaced six things only a backend can solve cleanly:

1. **LM Studio CORS.** Browser → `localhost:1234` works for Michael; doesn't work for any other reader the tool is intended to reach. Proxy server-side; readers configure the model URL, not their browser's CORS posture.
2. **API key safety.** If a reader plugs in an OpenAI / Anthropic / OpenRouter key, that key should never round-trip through `1828.ibeco.me`. Backend holds the key from a server-side env var; frontend never sees it.
3. **All scriptures in verse explorer.** Today the verse explorer iframes churchofjesuschrist.org or accepts pasted text. We want to *render* the verse ourselves with 1828-tier highlights inline. The public-domain bcbooks corpus (see §III) makes this possible without the copyright posture being violated.
4. **Lazy modern-def expansion.** Today we pre-fetch ~709 modern defs at build time, one word per second. Lazy fetch on demand + write back to DB means the corpus grows from use, and class-E words (every 1828 entry, not just the curated tier list) become queryable.
5. **Cross-reading Thummim.** The Thummim Restoration Dictionary entries live in `stewards.thummim_entries` on the pg-ai-stewards substrate. A backend gives us a place to expose those without the frontend reaching across DBs in the browser.
6. **Scripture search.** "Find every verse containing *intelligence*" is a one-line query against a DB and a multi-megabyte download in front of a browser. The DB is the right place.

Honor scope but exercise stewardship — when ingesting data, fix the data shape; when proxying calls, harden them; when laying out tables, schema them properly the first time. This is planning; the actual implementation is for sessions that follow.

## II. What does NOT change

- **Frontend stays Vue + Vite + Tailwind + vue-router.** No SSR. Built to static `dist/` and served by nginx.
- **Frontend Docker image stays small.** All Postgres + API weight lives in sibling containers.
- **Tier-list curation stays in source files.** The 853 tier-A/B/C words still drive what the verse explorer *highlights*. The backend just changes the source of the verse *text*.
- **Static deploy promise (cacheable, fast, indexable) holds for static surfaces.** Search, word lookups, LLM render are necessarily dynamic; everything else can be CDN-cached.
- **No coupling to pg-ai-stewards.** The 1828 backend is its OWN tiny service with its OWN Postgres. The substrate is a sibling product with a different lifecycle. The only crossing is reading Thummim entries (see §VI.D-BE-THM).
- **Cpuchip.net's iframe convention stays.** That's a different project. This one pivots; that one doesn't.

## III. The "is this really public-domain?" question — must answer before building

Michael's pivot rationale: "use the public-domain scripture source that `external_context/scriptures-mcp/` already wraps."

That MCP server's README says explicitly: *"2013 LDS edition text ensuring consistency with official sources."* The 2013 LDS edition is not public domain. The verse text underneath it (KJV for OT/NT, 1830 Book of Mormon, etc.) is. The apparatus on top — verse divisions in the modern form, chapter headings, italicized words, footnote markers, study aids, the 1981 LDS sectioning of D&C — is church-curated work, not all of which is unambiguously PD.

`bcbooks/scriptures-json` upstream does not include a license file in the data directory (only a repository LICENSE). Their position is implicit, not declared.

**This must be ratified before we ship anything that bundles their text.** Three options:

- **D-BE-COPYRIGHT (option A) — accept bcbooks as-is.** Treat the bcbooks corpus as community-resourced. Cite their repo as the source. Risk: if the LDS Church's intellectual property department disagrees, we have a takedown event.
- **D-BE-COPYRIGHT (option B) — strip to unambiguously-PD layer.** Start from a known-PD source for each volume: 1769 Cambridge KJV for OT/NT (PD), 1830 first edition Book of Mormon (PD), 1835 first edition D&C (PD but verse numbering differs from modern editions), 1851 Pearl of Great Price (PD with differences). Cross-reference modern verse numbering for the rendering layer. More work; copyright-clean.
- **D-BE-COPYRIGHT (option C) — keep iframe for canon, render only for user-pasted text.** Backend serves dictionary + LLM proxy + search-against-pasted-text. Scripture canon stays an iframe (or external link). Loses one of the six wins but keeps the copyright posture clean. **Default fallback if A and B both fail.**

This is the first decision Michael ratifies before any other proposal in this set is buildable. Defaulting silently to A would betray the `respect-the-canon` intent value.

## IV. Architecture — three containers under one compose file

```
1828.ibeco.me (Dokploy)
  └─ docker-compose.yaml  ← project root
       ├─ frontend       (nginx:alpine + Vue dist)
       │   ports 80 → host
       │   reverse-proxy /api/* → backend:8080
       ├─ backend        (Go binary)
       │   port 8080 (internal only)
       │   reads DB at db:5432
       │   reads SERVER-SIDE env vars: LLM_DEFAULT_URL, LLM_DEFAULT_KEY,
       │   LLM_DEFAULT_MODEL, THUMMIM_SNAPSHOT_URL (or DSN)
       └─ db             (postgres:17-alpine)
           volume: pg-data (named)
           init: /docker-entrypoint-initdb.d/ migrations
```

**Reverse-proxy posture:** nginx serves `/`, `/assets/*`, `/healthz` as today; new `location /api/ { proxy_pass http://backend:8080/; }` block routes the dynamic surfaces. Same-origin, no CORS dance.

**Why Go backend (not Node):** the workspace pattern is Go for backends (scripts/becoming, scripts/stewards-ui, external_context/scriptures-mcp). The bcbooks parser already exists in `external_context/scriptures-mcp/internal/scripture/service.go` — Go reuse, not reinvention. Single static binary; fast; small image.

**Why Postgres (not SQLite):** Dokploy deploys multiple replicas behind a load balancer; SQLite's file-lock model fights that. Postgres in a sibling container is the becoming-app pattern proven in this workspace. The DB is small (~150MB once loaded) and a single Postgres instance handles 1828's traffic for years.

## V. Postgres schema sketch

Five-namespace plan, each table prefixed by purpose:

```sql
-- Canon (loaded once at boot; immutable thereafter)
CREATE TABLE scripture_books (
  id              SERIAL PRIMARY KEY,
  volume          TEXT NOT NULL,        -- 'ot' | 'nt' | 'bofm' | 'dc' | 'pgp'
  abbr            TEXT NOT NULL UNIQUE, -- 'gen' 'matt' '1-ne' 'dc' 'moses' …
  name            TEXT NOT NULL,        -- 'Genesis' '1 Nephi' …
  display_order   INT NOT NULL
);
CREATE TABLE scripture_chapters (
  id              SERIAL PRIMARY KEY,
  book_id         INT NOT NULL REFERENCES scripture_books(id),
  chapter         INT NOT NULL,
  UNIQUE (book_id, chapter)
);
CREATE TABLE scripture_verses (
  id              SERIAL PRIMARY KEY,
  chapter_id      INT NOT NULL REFERENCES scripture_chapters(id),
  verse           INT NOT NULL,
  text            TEXT NOT NULL,
  text_tsv        tsvector GENERATED ALWAYS AS (to_tsvector('english', text)) STORED,
  UNIQUE (chapter_id, verse)
);
CREATE INDEX scripture_verses_text_tsv_idx ON scripture_verses USING GIN (text_tsv);
CREATE INDEX scripture_verses_text_trgm   ON scripture_verses USING GIN (text gin_trgm_ops);
-- ↑ trigram for LIKE %word% fallback; tsv for FTS. Both, cheap, GIN-indexed.

-- 1828 dictionary (loaded once from webster1828.json.gz)
CREATE TABLE webster_1828 (
  word            TEXT PRIMARY KEY,        -- lowercased headword
  entries         JSONB NOT NULL,          -- [{pos, definitions: [text...]}]
  source_offsets  JSONB                    -- provenance: which gz entries
);

-- Modern definitions (lazy + accumulating)
CREATE TABLE modern_defs (
  word            TEXT PRIMARY KEY,
  entries         JSONB,                   -- nullable; null = looked up + 404
  fetched_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  source          TEXT NOT NULL DEFAULT 'free-dictionary-api',
  error           TEXT                     -- non-null when last fetch errored
);

-- Tier metadata (loaded once from tier-words.json; surfaces in highlights)
CREATE TABLE tier_words (
  word            TEXT PRIMARY KEY,
  tier            TEXT NOT NULL CHECK (tier IN ('A++','A+','B','C','D')),
  study_tier      TEXT CHECK (study_tier IN ('A','B','C')),
  studies         JSONB NOT NULL,          -- ["intelligence.md", "art-of-presidency.md"]
  study_excerpts  JSONB NOT NULL,
  p4_score        INT,
  p4_reasons      JSONB NOT NULL
);

-- Thummim snapshot (read-only mirror of stewards.thummim_entries)
CREATE TABLE thummim_entries_cache (
  word            TEXT PRIMARY KEY,
  entries         JSONB NOT NULL,         -- {levels: {elementary, eighth, senior}}
  citations       JSONB NOT NULL,
  generated_at    TIMESTAMPTZ NOT NULL,
  imported_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Verse-highlight cache (memoized tokenize+tier-match per verse)
CREATE TABLE verse_highlights_cache (
  verse_id        INT PRIMARY KEY REFERENCES scripture_verses(id) ON DELETE CASCADE,
  segments        JSONB NOT NULL,         -- pre-computed [{text, word?, tier?}]
  tier_set        TEXT[] NOT NULL,        -- e.g. {'A++','A+','B'}
  computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  algo_version    INT NOT NULL DEFAULT 1  -- bumps invalidate cache
);
```

**Migrations live at `backend/migrations/Nxx-name.sql`**, applied at boot in lexicographic order. Idempotent statements only (`CREATE TABLE IF NOT EXISTS`, `INSERT … ON CONFLICT … DO UPDATE`). This mirrors the pg-ai-stewards naming convention, scaled down — no extension, no bgworker, no pgrx.

## VI. The endpoints

| Endpoint | Method | Returns | Notes |
|---|---|---|---|
| `/api/scripture/:ref` | GET | `{ ref, book, chapter, verse?, text, segments? }` | Parses 1 Nephi 3:7, John 3:16-17, etc. Supports verse + range. Includes pre-tokenized `segments` when `?highlight=1`. |
| `/api/scripture/chapter/:ref` | GET | `{ ref, book, chapter, verses: [{verse, text, segments?}] }` | Full chapter; same `?highlight=1` opt-in. |
| `/api/scripture/search?q=:phrase&limit=:n` | GET | `[{ ref, text, snippet }]` | FTS via `text_tsv`; falls back to trigram for short queries. Default limit 20, max 100. |
| `/api/dict/1828/:word` | GET | `{ word, entries: [{pos, definitions: [...]}] }` | Stem-fallback (`obtaineth → obtain`) applied server-side. |
| `/api/dict/modern/:word` | GET | `{ word, entries: [...]? , source: "cache" \| "fetched" \| "none" }` | If not in DB, lazy-fetch from Free Dictionary; write-back. Returns null entries + `source: none` on 404; caches 404s too. |
| `/api/dict/tier/:word` | GET | `{ word, tier, studies, excerpts, p4_* }` | The tier-words.json shape, served from DB. |
| `/api/thummim/:word` | GET | `{ word, levels: {elementary, eighth, senior}, citations }` | From `thummim_entries_cache`; 404 when not yet snapshotted. |
| `/api/llm/render` | POST | `{ modernized, promptUsed, durationMs }` | Body `{ verseText, tierWords?: [{word, sense}] }`. Provider-agnostic; backend picks provider from env. |
| `/api/healthz` | GET | `ok` | For Dokploy health checks. |

**Rate-limiting:** deferred to a later phase. Document per-IP token bucket as a known carry-forward in `llm-proxy.md`.

## VII. Migration path — old JSON → DB

Two-step, no flag day:

1. **Ingest at first boot.** Backend on boot inspects DB; if `scripture_verses` has zero rows, runs `seed_canon.go` against the bundled bcbooks zip (or the chosen PD source per D-BE-COPYRIGHT). Same for `webster_1828` (seeded from `scripts/webster-mcp/data/webster1828.json.gz`, embedded into the backend binary the way scriptures-mcp embeds its data). Same for `tier_words` (seeded from `frontend/src/data/tier-words.json` — copied into `backend/seed/` at build time). Same for `modern_defs` (seeded from `frontend/src/data/definitions-modern.json`).
2. **Frontend cuts over.** `useWordData.ts` and `useLLMRender.ts` swap from static imports to fetch calls. The static JSON files **remain** as the build-time seed (single source of truth for the seed; backup if DB ever needs reseeding). They no longer ship in the frontend bundle.

**Stewardship: while we ingest, fix data quirks.** The current `definitions-modern.json` stores some entries as `null` for "looked up + 404." The DB schema preserves that signal explicitly (`entries IS NULL AND error IS NULL` = clean 404). The current `build_data.py` has a hard-coded Windows path; the Go ingest path doesn't.

After cutover, the JSON bundles in `frontend/src/data/` are no longer load-bearing for the running site, but they stay in the repo as the canonical seed (so the DB can always be rebuilt from source). `build_data.py` and `fetch_modern_defs.py` keep working as research-time tools that regenerate the seeds; ingestion picks the seeds up at backend rebuild.

## VIII. CLAUDE.md / intent.yaml lines that change

After this pivot ratifies and ships, these specific lines need rewriting (not now — flagged so the implementing session does it):

**`projects/1828-illuminated/CLAUDE.md`:**
- *"No backend in MVP. Static site."* → "Backend now exists for dynamic surfaces (search, LLM proxy, lazy modern-defs, Thummim cross-read). Frontend remains a static SPA served by nginx; the backend is a sibling container."
- *"No LM Studio load."* → "The 1828 backend proxies LLM calls. The default provider in the deployed env is the user's choice; the local dev `.env` may point at LM Studio. The gospel-engine-v2 LM Studio pipeline is unaffected (separate container, separate concern)."
- *"Scripture text is NOT bundled."* — **conditional.** Stays as written if D-BE-COPYRIGHT lands on option C. Rewrites to "Scripture text is bundled from a public-domain source (see scripture-corpus.md)" if A or B.

**`projects/1828-illuminated/intent.yaml`:**
- `constraints.static-deploy-target` — soften "no backend in MVP" to "static frontend with sibling API; no SSR."
- `constraints.no-mcp-load-on-lm-studio` — keep wording; clarify that the *new* backend's LLM proxy is the user's choice, not gospel-engine-v2.
- `constraints.no-scripture-text-bundled` — same conditional as CLAUDE.md.
- `stretch_goals.llm-rendering` — promote from `priority: low` to `priority: medium`; the LLM proxy makes this no longer gated.

## IX. Decisions for Michael to ratify

| # | Decision | Default | Stakes |
|---|---|---|---|
| **D-BE-COPYRIGHT** | bcbooks as-is / strip-to-PD / iframe-only-for-canon | **must answer first; no default** | If wrong, takedown risk or scope creep |
| **D-BE-1** | Go backend vs Node | Go | Go; matches workspace pattern (becoming, stewards-ui, scriptures-mcp) |
| **D-BE-2** | Postgres 17-alpine vs 16-alpine vs 18 | 17-alpine | Matches becoming-app; supported through 2029 |
| **D-BE-3** | Same docker-compose project as Dokploy / separate | Same | Single deploy unit, single rollback |
| **D-BE-4** | Backend embeds seed data via `//go:embed` (mirrors scriptures-mcp) or reads from a mounted volume | `//go:embed` | Embed = self-contained binary; volume = swappable corpus without rebuild |
| **D-BE-5** | Migrations directory (`backend/migrations/Nxx-name.sql`, lex-order) vs a tool like goose/migrate | Plain SQL + lex-order | Lower dep surface for a small service |
| **D-BE-6** | nginx reverse-proxies `/api/*` to backend (same-origin) vs CORS-allowed cross-origin | Reverse-proxy | No CORS complexity |
| **D-BE-7** | DB backup strategy | `pg_dump` nightly to a workspace-local path + Dokploy volume snapshot | Modern-def cache is the only non-reproducible data; weekly is probably fine |
| **D-BE-8** | Frontend keeps the seed JSON files in `frontend/src/data/` after cutover | Yes (as canonical seed) | Source of truth for rebuilds; not bundled into runtime |
| **D-BE-THM** | Thummim sync: postgres_fdw vs nightly snapshot vs HTTP endpoint exposed by stewards-ui | Nightly snapshot | Decouples the deploys; 14 entries today don't change minute-to-minute |
| **D-BE-AUTH** | `/api/llm/render` open to anonymous traffic? Token-gated? Per-IP rate-limit? | Open with per-IP rate-limit (deferred to phase 2) | Could become a free LLM proxy if not careful |
| **D-BE-CORS-FOR-PASTE** | If a power user wants to call `/api/scripture/search` from another site, allow it? | No (same-origin only at v1) | Permissive CORS can come later if requested |

Eleven decisions plus the prerequisite copyright question. Ratify D-BE-COPYRIGHT first; the rest can be batched.

## X. Phases

| Phase | What ships | Depends on |
|---|---|---|
| **0** | D-BE-COPYRIGHT ratified; CLAUDE.md/intent.yaml line-change list written into the commit message | none |
| **1** | docker-compose.yaml + Postgres container + empty backend Go skeleton + nginx proxy block + healthchecks | 0 |
| **2** | `scripture-corpus.md` execution — ingest scripture, expose `/api/scripture/*` + search | 1 |
| **3** | `dictionary-backend.md` execution — ingest 1828 + tier_words; expose `/api/dict/1828/*` + `/api/dict/tier/*`; modern-defs lazy fetch | 1 |
| **4** | `llm-proxy.md` execution — `/api/llm/render` with provider abstraction | 1 |
| **5** | Frontend cutover — `useWordData.ts` + `useLLMRender.ts` swap to API | 2, 3, 4 |
| **6** | Thummim snapshot job + `/api/thummim/*` endpoint + frontend Dictionary view wires through | 1 + Thummim entries exist |
| **7** | Backup + observability polish (pg_dump cadence, basic metrics endpoint) | 1 |

Phase 5 is the user-visible cutover. Phases 2-4 can be built in any order or in parallel by separate sessions; they don't cross.

## XI. Risks

- **Copyright on bcbooks corpus.** Already named (§III). Must be answered before phase 2.
- **Image bloat.** Postgres 17-alpine + a 150-200MB seed embed + Go binary is still well under 500MB total; not a real risk, but worth measuring.
- **Lazy modern-def fetch hitting Free Dictionary API rate limits.** They're community-supported; we should cap to 1 req/sec server-side and cache 404s permanently. Document in `dictionary-backend.md`.
- **DB-as-source-of-truth makes the deploy a state-bearing system.** First time this project has had durable state. The `frontend/src/data/*.json` seeds remain so the DB can be rebuilt; pg_dump cadence (D-BE-7) is the second line of defense.
- **Coupling drift between 1828 backend and pg-ai-stewards.** Mitigated by D-BE-THM picking snapshot, not FDW. If we ever want sub-minute Thummim freshness, revisit.
- **Steward-mode neighboring fixes scope creep.** Real risk. Bounded by "fixes in the data and code we touch this phase; not fixes elsewhere in the workspace." If we find an issue in `webster-mcp/data/`, we fix it in the ingest path of the 1828 backend rather than upstream.

## XII. Carry-forward (out of scope here, named so we don't lose it)

- pgvector for semantic verse search (currently FTS-only)
- AGE graph (Michael's Thummim-graph idea, §VI.5 of thummim-restoration-dictionary.md) — would live on the substrate, not the 1828 backend
- Presentation mode (intent.yaml stretch goal) — UI work, unaffected by this pivot
- Streaming LLM render (today's render is single-shot; SSE would be the upgrade)
- Multi-language: 1828 is English-only; the bcbooks corpus is English-only; multi-lang is a separate project altogether

---

See the child proposals — `scripture-corpus.md`, `dictionary-backend.md`, `llm-proxy.md`, `deployment-shape.md` — for the per-surface details.
