---
title: becoming — scripture lookup via engine.ibeco.me (drop the bundled download)
date: 2026-05-23
status: ratified — building
workstream: WS2
purpose: >
  Stop bundling ~20MB of scripture markdown files into becoming's
  Docker image. Have the running app fetch scripture verses from
  gospel-engine-v2 (engine.ibeco.me) at runtime. Soft-dependency
  semantics — when engine is down, the UI still works, scripture
  lookup returns "unavailable" instead of crashing.
---

# Scripture via engine.ibeco.me

## I. Why

The current `scripts/becoming/Dockerfile` carries a dedicated build
stage (`scriptures`) that runs `gospel-downloader -standard` against
the bcbooks/scriptures-json corpus, producing a ~20MB `/scriptures`
tree the runtime stage copies forward. On every Dokploy build —
including frontend-only changes — this stage either re-runs (when
layer cache is cold) or sits on the critical path even cached.
**Builds have been taking too long** and the scripture corpus is
already living at `engine.ibeco.me`, which becoming's deploy is
already coupled to (shared ibeco.me key management).

The right answer Michael named in the request: drop the bundled copy,
fetch from engine. This proposal does that.

## II. Surface change

### II.1 What the app's external HTTP API does NOT change

Becoming's three scripture endpoints keep their shapes:

- `GET /api/scriptures/lookup?ref=...`
- `GET /api/scriptures/books`
- `GET /api/scriptures/search?q=...`

Frontend (`ReferencePanel.vue`, `ReaderView.vue`, `PublicReaderView.vue`)
gets the same JSON it gets today. **No frontend code changes.**

### II.2 What the BACKEND does

Two of the three endpoints don't actually need engine — they operate
on data that's already metadata-only in `internal/scripture/scripture.go`:

| Endpoint | Backing today | Backing after this proposal |
|---|---|---|
| `/api/scriptures/lookup?ref=...` | `scripture.Lookup(root, ref)` reads markdown from `/scriptures` | **HTTP call to `engine.ibeco.me/api/get?reference=<ref>`** |
| `/api/scriptures/books` | `scripture.ListBooks()` — hardcoded Go list of OT/NT/BoM/D&C/PGP books | **Unchanged.** Already metadata, no IO needed. |
| `/api/scriptures/search?q=...` | `scripture.SearchBooks(q)` — substring match against hardcoded book names | **Unchanged for now.** Future: opt-in upgrade to engine's full-content `/api/search` as a second tier of results. |

Only the lookup path crosses the engine boundary. That's where the
20MB bundle paid off; that's where the engine call substitutes.

### II.3 What the Dockerfile drops

- The entire `scriptures` Go stage (lines 10-19 in
  `scripts/becoming/Dockerfile`) — gone
- The runtime stage's `COPY --from=scriptures /gospel-library/eng/scriptures /scriptures` — gone
- The runtime stage's `ENV BECOMING_SCRIPTURES=/scriptures` — gone (replaced by engine env)

Image shrinks by ~20MB. Two fewer build stages. Builds get faster.

## III. Engine contract (verified live 2026-05-23)

```
GET https://engine.ibeco.me/api/get?reference=John+3:16
Authorization: Bearer <ENGINE_SERVICE_TOKEN>

200 OK
{
  "reference_query": "John 3:16",
  "source_type": "scripture",
  "verses": [
    {
      "id":        12345,
      "volume":    "nt",
      "book":      "John",
      "chapter":   3,
      "verse":     16,
      "reference": "John 3:16",
      "text":      "For God so loved the world...",
      "file_path": "..."
    }
  ]
}
```

Verse-range and full-chapter lookups return arrays of verses in the
same shape. Engine handles range parsing (`John 3:16-17`), chapter
references (`John 3`), and book aliasing (`D&C 93` ↔ `Doctrine and
Covenants 93`) — that's now their concern, not ours.

## IV. Soft-dependency semantics (the load-bearing part)

When engine is unreachable, slow, or returns 5xx:

| Endpoint | Behavior |
|---|---|
| `/api/scriptures/lookup` | **HTTP 503 Service Unavailable** with body `{"available": false, "error": "scripture lookup temporarily unavailable"}`. Frontend already gracefully degrades when this endpoint errors (verified via existing error-handling in `ReferencePanel.vue`); no frontend changes needed. |
| `/api/scriptures/books` | Unchanged — always works (hardcoded data). |
| `/api/scriptures/search` | Unchanged — always works (hardcoded data). |

UI continues to function. Memorization cards display. Reading view
still renders. Only the live verse-pull from engine is the part that
gracefully degrades — and even THAT survives if it's been cached this
session (see §V).

### IV.1 Definition of "down"

Engine is "down" when ANY of:
- DNS fails / network refused
- TCP timeout (configurable — default 5s)
- HTTP 5xx response
- HTTP 401/403 (token misconfigured — operationally "down" from app's POV)
- JSON parse fails

For all of these, the client returns a sentinel error that the
handler maps to the 503 + body above. Per-error counters logged via
the existing slog so an operator can correlate.

### IV.2 What we do NOT do (intentional)

- **No retry policy.** A single 5s-budget call. If it fails, surface
  the error. The frontend is responsible for any retry-on-user-action.
- **No circuit breaker.** Engine being down isn't a thundering-herd
  risk; calls are user-initiated reference lookups, not bots.
- **No fallback to bundled scriptures.** That defeats the whole point
  of dropping the 20MB. The in-memory cache (§V) is the only
  intra-process resilience layer.

## V. Caching

A tiny LRU on the client side:
- Capacity: 200 entries
- Key: the canonicalized reference string (lowercased + whitespace-collapsed)
- Value: the `LookupResult` (full JSON-deserialized struct)
- TTL: forever (process lifetime) — scripture text is immutable; engine
  rebuilds happen rarely + go through a redeploy that wipes the cache
- Survival on engine outage: if a verse has been looked up THIS
  session, subsequent calls return from cache regardless of engine
  status. This makes memorization-card review during an engine blip
  invisibly survive.

200 entries × ~200 bytes typical verse = ~40KB of RAM. Negligible.

## VI. Auth — using the existing ibeco.me ↔ engine link

Michael's hint: "engine's key management is ibeco.me." Translation:
becoming-app already has a token for engine because ibeco.me's
deploy already issues + holds them.

Env shape (new):

```
ENGINE_URL=https://engine.ibeco.me        # default; can be overridden
ENGINE_TOKEN=<bearer>                     # required for /api/get + /api/search
ENGINE_TIMEOUT_SECONDS=5                  # default
```

`BECOMING_SCRIPTURES=/scriptures` is **removed** from `.env.example`
and `docker-compose.yml`. If still set in production, the runtime
ignores it harmlessly.

If `ENGINE_TOKEN` is unset at startup, becoming logs a clear startup
warning ("scripture lookup will return 503 — set ENGINE_TOKEN in
Dokploy env") and continues to boot. Books + search keep working.
This makes the env wiring forgiving — a deploy without the token still
brings the app up; only the lookup endpoint degrades.

## VII. Decisions

| # | Decision | Choice |
|---|---|---|
| **D-EE-1** | Where the engine client lives | **`internal/scripture/engine_client.go`** — sibling of `scripture.go`, shares the existing `LookupResult` + `Verse` types so handler code is a one-line swap |
| **D-EE-2** | Books endpoint sourcing | **Hardcoded Go list, unchanged.** Engine has no exact equivalent; the list is stable across decades; no runtime dep needed for this surface |
| **D-EE-3** | Search endpoint sourcing | **Hardcoded book-name search, unchanged for v1.** Future: add engine's `/api/search` as opt-in second tier (off in v1 to keep scope tight) |
| **D-EE-4** | Cache scope | **Process-local LRU(200), no TTL.** No Redis, no shared cache — single process, restart wipes |
| **D-EE-5** | Auth failure semantics | **401/403 from engine → 503 to client** (same as network failure). Operationally indistinguishable from the app's POV |
| **D-EE-6** | Behavior when `ENGINE_TOKEN` is unset | **Boot succeeds, log warning, lookup returns 503.** Forgiving to operators who forget env vars |
| **D-EE-7** | Drop the bundled `/scriptures` directory in production | **Yes.** The whole point. Existing volume mounts to that path become no-ops; no breakage |

All defaults align with Michael's stated direction. No questions to ratify.

## VIII. Phases (one session — small)

1. **§I client** — write `internal/scripture/engine_client.go`. New
   `EngineClient` struct with `Lookup(ctx, ref) (*LookupResult, error)`.
   Standard `http.Client` with timeout. LRU cache. Sentinel
   `ErrEngineUnavailable` for the soft-dep cases.
2. **§II handler swap** — in `internal/api/router.go`, change
   `lookupScripture(scripturesRoot)` to accept an `*EngineClient`
   instead of a string. Map `ErrEngineUnavailable` → HTTP 503 +
   the `{available: false, error: ...}` body.
3. **§III wiring** — `cmd/server/main.go` reads `ENGINE_URL` +
   `ENGINE_TOKEN`, constructs the client, passes it to `Router`. Drop
   the `--scriptures` / `BECOMING_SCRIPTURES` flag and env.
4. **§IV Dockerfile** — drop the `scriptures` stage. Drop the
   runtime `COPY --from=scriptures`. Drop the `ENV BECOMING_SCRIPTURES`.
   Image shrinks by ~20MB.
5. **§V compose + env** — `scripts/becoming/docker-compose.yml`
   replaces `BECOMING_SCRIPTURES` with `ENGINE_URL` + `ENGINE_TOKEN`.
   `.env.example` updated to match.
6. **§VI smoke** — local Docker build green; local container call to
   `/api/scriptures/lookup?ref=John 3:16` returns the verse (engine
   reachable); kill engine credential, hit same endpoint, get the 503
   body; confirm `/api/scriptures/books` still returns.
7. **§VII commit + push** — single commit. Redeploy on Dokploy.

## IX. Rollback

The legacy `scripture.Lookup(root, ref)` filesystem code stays in
`scripts/becoming/internal/scripture/scripture.go` after this
proposal. If engine integration has issues post-deploy:

- Restore the `scriptures` Docker stage + runtime COPY (one git revert)
- Restore `BECOMING_SCRIPTURES=/scriptures` env
- Handler reverts to passing the path instead of the engine client

Rollback is a clean revert — no schema changes, no data migration.

## X. Out of scope (deferred deliberately)

- Engine-backed full-text search in becoming (D-EE-3 names it; v1 keeps
  the existing book-name search)
- Sharing the engine client across the workspace (1828 backend has its
  own scripture corpus in Postgres, so this isn't its concern; the
  shared abstraction would only have one consumer for now)
- Per-user / per-account rate-limiting on the engine token (becoming's
  traffic is low; engine handles its own quotas)
- Failover to a second engine instance — single engine for now
