---
date: 2026-05-20
session_type: dev
project: 1828-illuminated
workstream: WS7
status: phases-1-4-shipped
commits:
  - 6b0e98a — phase 1 (compose skeleton)
  - 36419b1 — phase 2 (scripture corpus)
  - 70ede1f — phase 3 (dictionary backend)
  - 554db8c — phase 4 (LLM proxy + BYOK)
---

# Backend pivot build — phases 1–4

Started post-ratification. Picked up Michael's directive: build the four
phases of the backend pivot, stop before phase 5 (frontend cutover).
All four landed in one session. Each on `main` with its own commit
naming what shipped, what tests passed, and the carry-forward.

## What shipped, by phase

**Phase 1 — three-container skeleton.** Go 1.23 module under `backend/`,
`cmd/server` with healthcheck subcommand for the distroless image,
`internal/migrate` running lex-ordered embedded SQL migrations,
`internal/seed.RunAll` as a no-op stub. `Dockerfile.frontend` +
`Dockerfile.backend` (multi-stage; distroless static + nonroot for
backend) + `Dockerfile.legacy` preserved for the rollback escape hatch.
`docker-compose.yaml` with frontend on host port **8083** (not 80 —
preserves the legacy `1828` container on 8082 untouched per the hard
constraint). DB is internal-only Postgres 17-alpine. nginx adds the
`/api/*` reverse-proxy upstream block with `keepalive 16`. All 4
migrations (001 extensions, 002 scripture, 003 dictionary, 004 thummim
+ cache) declared up-front so the schema timeline stays monotonic
across phases. `docker compose up -d` brings all three containers up
healthy in ~30s.

**Phase 2 — scripture corpus + endpoints.** Embedded `scriptures.zip`
(2.3MB) from the workspace's `external_context/scriptures-mcp/`. Hand-
curated 87-book abbr map matching `gospel-library/eng/scriptures/`
directory naming. Ref parser handles BOTH the workspace abbr form
(`dc/84:38`, `1-ne/3:7-10`) AND the human form (`1 Nephi 3:7`,
`D&C 84:38`). Strip rules per D-BE-COPYRIGHT option D:
bracketed editorial inserts, HTML tag wrappers, HTML entities,
whitespace collapse, and a steward-mode space-before-punct fix surfaced
during smoke-test (`place [Kirtland].` had been ingested as `place .`).
`scripture_corpus_meta` records the bcbooks source SHA + applied strip
rules for audit. Four endpoints:
`/api/scripture/{ref}`, `/api/scripture/chapter/{ref}`,
`/api/scripture/search` (FTS with `<em>` snippet headlining + trigram
fallback for short queries), `/api/scripture/word-study/{word}`
(archaic-suffix expansion via tsquery OR). 41,995 verses ingest in
5.4s via CopyFrom.

**Phase 3 — dictionary backend.** Three seed corpora embedded:
`webster1828.json.gz` (8MB → 98,828 distinct headwords on ingest in
1.2s), `tier-words.json` (853 auto), `manual-additions.json` (7
hand-curated), `definitions-modern.json` (709 pre-fetched). Four
handlers: `/api/dict/1828/{word}` (literal → archaic-suffix stem
fallback returning `stem_matched`), `/api/dict/modern/{word}` (DB
lookup → lazy fetch + write-back from Free Dictionary API with 1
req/sec singleflight cap + per-day cap counter), `/api/dict/tier/{word}`,
`/api/dict/search` (returns BOTH `tier_results` AND `all_1828_results`
— the class-E reach made visible). Smoke proven: `gainsay` (NOT in
tier list, IS in 1828) returns its entry; `obtaineth` returns
`obtain`'s entries with `stem_matched: "obtain"`; `peradventure` first
call `source: fetched`, second call `source: cache`.

**Phase 4 — LLM proxy + BYOK session-key flow.** This was the most
specification-sensitive surface. Full session lifecycle:
- POST `/api/llm/session` with `{provider, base_url?, api_key, model}`
  → probe upstream `/v1/models` (skip for mock) → mint 32-byte
  URL-safe-base64 session_id → store in in-memory map keyed by id
  with held provider+key+model+expires_at → return JSON + set
  cookie `i1828_session` (Path /api/llm, SameSite=Lax, Secure when
  TLS).
- DELETE `/api/llm/session` drops the held key, expires the cookie.
- GET `/api/llm/session` inspection returns `{active, provider, model,
  expires_at}`.
- POST `/api/llm/render` resolves credentials by cookie OR
  `Authorization: Bearer session_id` (curl-friendly) OR server-default
  when `LLM_PROVIDER` is set OR `LLM_PROVIDER=mock` (anonymous canned
  response path). Prompt built from `useLLMRender.ts §V` verbatim plus
  the steward-mode 800-token cap line. Sliding TTL extends `expires_at`
  by `LLM_SESSION_TTL_HOURS` on every successful render.

Rate limits applied: per-IP per-minute (default 10), per-IP per-day
(default 1000), global daily token cap on the server-default key only.
Rate-limit error body shape matches D-BE-AUTH **exactly**:
`{error: "rate_limited_by_1828", limit_type, retry_after_seconds,
message: "1828.ibeco.me throttled this … this is our cap, not theirs."}`
plus `Retry-After` header. Janitor goroutine evicts expired sessions
every 60s.

## What fought me

- The Go workspace's `go.work` excluded the new module — needed to add
  `./projects/1828-illuminated/backend` to the `use` block before
  `go mod tidy` could resolve.
- Docker healthcheck on distroless requires `["CMD", "/i1828",
  "healthcheck"]` (not just `["/i1828", "healthcheck"]`). Compose's
  validator rejected the latter.
- Frontend's `wget http://localhost/healthz` in the compose healthcheck
  resolved to IPv6 `::1` but nginx alpine binds IPv4 only. Container was
  unhealthy in `docker compose ps` despite serving outside traffic fine.
  Fixed to explicit `127.0.0.1` in both compose file and Dockerfile.
- `strings.Replacer.Replace` does not take a count argument (I had
  written `.Replace(text, -1)` reflex from `strings.Replace`).
- `pgx.PgError` is actually `pgconn.PgError` in v5.
- bcbooks D&C uses a different JSON shape (`sections` not
  `chapters/books`). Built a small `flatten()` to normalize.

## Decisions I made that weren't in the proposals

- **Frontend host port = 8083** (proposals showed `80:80`). Real-world
  constraint: the legacy `1828` static container is on host port 8082
  and must keep serving during the soak. Changed to 8083 with a comment
  pointing at the rationale. Dokploy production uses Traefik routing,
  so the host port is local-dev only.
- **Migration files declared all at phase 1.** Spec showed phase-2 and
  phase-3 inserting their own migration files. I batched all 4
  migrations into phase 1's commit so the schema timeline stays
  monotonic and the per-phase commits are pure code. This is mildly
  out of spec but cleaner; the spec text says lex-order + idempotent,
  which is satisfied either way.
- **Mock provider explicitly anonymous-allowed and skips probe.**
  Otherwise the `LLM_PROVIDER=mock` "useful for the frontend
  integration test without spending tokens" property fails. Code path
  is two early-returns in `resolveCallerCredentials` and `probeKey`.
- **`source_offsets` on `webster_1828`** stores `{source: "...",
  entry_count: N}` rather than the byte-offset shape the spec hinted
  at. The gz format doesn't have stable byte offsets after gunzip; the
  entry-count provenance answers the same audit question without
  pretending to a precision we don't have.

## Steward-mode fixes folded in (per Michael's 2026-05-20 directive)

- `.dockerignore` explicitly excludes `.env` / `.env.*` (defense in
  depth against accidentally COPYing the opencode-go key).
- `.gitignore` redundantly excludes `.env` at the project level (the
  workspace-level rule already covers it, but the redundancy is
  cheap).
- Space-before-punctuation collapse after bracket-stripping in scripture
  normalize. Caught during smoke-test, not in any proposal.
- nginx upstream block uses `keepalive 16` (the legacy nginx.conf used
  one fresh TCP connection per proxy request).
- Frontend healthcheck IPv6/IPv4 mismatch (above).
- Session ID logged as a truncated 8-char prefix so even debug-level
  logs never reveal the full bearer token.
- Free Dictionary fetch sets `User-Agent: 1828-illuminated.ibeco.me/0.1`
  so the API maintainers can reach us if they need to.
- `htmlEntities.Replace(text)` (not `Replace(text, -1)` which would
  have been `strings.Replace` semantics, not `strings.Replacer`).

## What didn't get touched (deliberately)

- The existing static `1828` container on port 8082 — never restarted,
  never reconfigured.
- `pg-ai-stewards-dev/bridge/ui` and `gospel-engine-v2-app/db` — also
  untouched. The 1828 stack uses its own internal-only Postgres on its
  own named volume `i1828-pg-data`.
- The frontend Vue/Vite code — phase 5 work, deliberately deferred.
- The Thummim entries backfill on the substrate — out of the 1828
  build's scope; would have been a pg-ai-stewards write, not a 1828
  write. The 1828 backend's `thummim_entries_cache` table exists
  (migration 004) but is empty until phase 6's sync job lands.

## Final state of the stack

```
NAME             STATUS                    PORTS
i1828-backend    Up (healthy)              8080/tcp (internal)
i1828-db         Up (healthy)              5432/tcp (internal)
i1828-frontend   Up (healthy)              0.0.0.0:8083->80/tcp

legacy:
1828             Up 4h (unhealthy*)        0.0.0.0:8082->80/tcp
                 *the legacy container's healthcheck pre-dates this
                  session — not caused by my changes
```

11 tables in the DB. 41,995 verses + 98,828 1828 headwords + 859 tier
words + 710 modern defs + 1 modern-fetch counter row. Total DB size
under 200MB.

## Phase 5 carry-forward (next session)

Detailed in the phase 4 commit body, but the headline:

- `useWordData.ts`: drop static imports of `definitions-1828.json` and
  `definitions-modern.json`; keep `tier-words.json` static for
  highlight rendering; swap synchronous lookups to async fetches against
  `/api/dict/1828/{word}` and `/api/dict/modern/{word}`. Remove
  client-side stem fallback (server is now the source of truth).
- `useLLMRender.ts`: drop direct browser→provider fetch; call
  `POST /api/llm/render` with `credentials: 'include'`.
- `useLLMSettings.ts`: reshape localStorage to hold
  `{session_id, expires_at, provider, model, temperature, maxTokens}`
  — NOT the api_key.
- New `useLLMSession.ts`: `startSession()` / `endSession()` /
  `isSessionActive()`.
- Vite: add `VITE_API_BASE_URL`. Defaults to `/api` in production
  (nginx proxies), `http://localhost:8083/api` in dev against the
  compose stack.
- `Settings.vue`: add the `type="password"` API-key input + Start
  session / Sign out buttons. One-time banner for users with v1
  localStorage shape.

## Spec ambiguities I chose through

- The proposals didn't specify a host port; I picked 8083 (above).
- The proposals didn't specify a session_id encoding; I chose
  URL-safe-base64 without padding (32 random bytes → 43-char id).
- The proposals didn't specify whether GET `/api/llm/session` was a
  thing; I added it as inspection so the frontend can show "Session
  active until …" without a render attempt.
- The 1828 corpus seeder uses `CopyFrom` with grouped JSONB rows; the
  proposal said "INSERT in chunks of 500 OR `COPY FROM STDIN`." Picked
  CopyFrom because we're well under the 30s threshold the spec offered
  as the switchover criterion.
- The proposal mentioned `definitions-modern.seed.json` (renamed); I
  kept the file as `definitions-modern.json` to avoid an extra rename
  step. Filename doesn't matter; the embed path resolves either way.

These are flagged here, not as defects, so the spec can be tightened
in a future revision if Michael disagrees with any of them.
