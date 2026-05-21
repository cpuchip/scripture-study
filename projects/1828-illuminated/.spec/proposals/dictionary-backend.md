---
title: 1828-illuminated — dictionary backend
date: 2026-05-20
status: proposed
workstream: WS7
parent: backend-pivot.md
purpose: >
  Move Webster 1828, modern definitions, and tier-metadata from JSON
  bundles in the frontend into the 1828 backend's Postgres. Unlock
  class-E reach (every 1828 entry queryable, not just the curated 853),
  lazy modern-def expansion that grows from use, and a clean place to
  cross-read Thummim entries.
---

# Dictionary Backend

## I. What this surface owns

Three lookup paths, all dictionary-shaped:

1. **`/api/dict/1828/:word`** — the full Webster 1828 corpus (~98,000 headwords). Today only the ~853 tier words ship to the frontend; this expands to every word.
2. **`/api/dict/modern/:word`** — modern-English definitions from the Free Dictionary API. Today ~709 are pre-fetched at build time; this becomes lazy + write-back, growing the corpus from real use.
3. **`/api/dict/tier/:word`** — the tier metadata (`A++` / `A+` / `B` / `C` / `D` + linked studies). Today this is `tier-words.json` in the bundle; this moves to DB.

A fourth path, `/api/thummim/:word`, lives in the same backend code but pulls from a *separate* schema cache (`thummim_entries_cache`) synced from the substrate — covered briefly here, decision lives in `backend-pivot.md §D-BE-THM`.

## II. Why move the 1828 corpus to the backend

`scripts/webster-mcp/data/webster1828.json.gz` is ~5-7MB gzipped, ~25-30MB uncompressed. The frontend can't ship that in a Vue bundle — too heavy. Today's `build_data.py` extracts just the tier words (853 entries) into `definitions-1828.json` (~600KB), and that's what ships.

This gives us the **class-E reach gap.** A user searching for a word *not* in our tier list gets nothing — even if the 1828 has a perfectly good entry. Examples from the substrate's own studies:

- `gainsay` — appears 8× in scripture, has a rich 1828 entry, didn't make the tier list because no study lensed it explicitly
- `bowels` — 1828 sense ("the most internal seat of the affections; mercy") is sharply different from modern; not in tier
- `effectually` — Wesleyan / Reformed usage in 1828; not tier-listed
- `peradventure`, `peculiar`, `verily`, `wist`, `holpen`, `wot`

The fix is trivial once a backend exists: ship the *full* 98k headwords into Postgres at boot, and serve them on demand. ~98k rows of small JSONB is well under 100MB DB-side; lookups are PK by word, sub-millisecond.

## III. Schema (excerpted from backend-pivot.md §V)

```sql
CREATE TABLE webster_1828 (
  word            TEXT PRIMARY KEY,        -- lowercased headword
  entries         JSONB NOT NULL,          -- [{pos, definitions: [text...]}]
  source_offsets  JSONB                    -- provenance: which gz entries (audit)
);

CREATE TABLE modern_defs (
  word            TEXT PRIMARY KEY,
  entries         JSONB,                   -- nullable; null = looked up + 404
  fetched_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  source          TEXT NOT NULL DEFAULT 'free-dictionary-api',
  error           TEXT                     -- non-null when last fetch errored
);

CREATE TABLE tier_words (
  word            TEXT PRIMARY KEY,
  tier            TEXT NOT NULL CHECK (tier IN ('A++','A+','B','C','D')),
  study_tier      TEXT CHECK (study_tier IN ('A','B','C')),
  studies         JSONB NOT NULL,
  study_excerpts  JSONB NOT NULL,
  p4_score        INT,
  p4_reasons      JSONB NOT NULL
);

CREATE INDEX webster_1828_word_trgm ON webster_1828 USING GIN (word gin_trgm_ops);
-- Trigram on the 1828 headwords for the prefix-search box that today
-- only searches tier words. Class-E reach for the search UX too.
```

**The `source_offsets` JSONB on `webster_1828`** is small (`{"file":"webster1828.json.gz","line_or_index":N}`); used only when verifying that a definition we serve matches the source bundle. Cheap, serves the audit principle.

## IV. Ingest path

Three seed steps at first boot (or when `webster_1828` is empty):

```go
// backend/internal/seed/dictionary.go

//go:embed data/webster1828.json.gz
var webster1828Gz []byte

//go:embed data/tier-words.json
var tierWordsJSON []byte

//go:embed data/definitions-modern.seed.json
var modernDefsSeed []byte

func SeedDictionary(ctx context.Context, db *sql.DB) error {
    // 1. SeedWebster1828: gunzip + json-decode + INSERT … ON CONFLICT (word) DO UPDATE
    //    SET entries = EXCLUDED.entries
    //    ~98k rows; batched INSERT in chunks of 500. Bonus: COPY FROM STDIN would
    //    be faster, use it if the chunked-INSERT measurement is > 30s.
    // 2. SeedTierWords: parse tier-words.json (the existing build-time output);
    //    INSERT … ON CONFLICT (word) DO UPDATE.
    // 3. SeedModernDefs: prime the cache with what we already fetched
    //    (~709 entries). INSERT … ON CONFLICT (word) DO NOTHING so we don't
    //    overwrite anything users have lazily fetched fresher.
    return nil
}
```

**Stewardship fixes during ingest:**

- `definitions-modern.json` today stores some entries as JSON `null` (the "lazy-looked-up-and-not-in-dictionary" signal). On ingest, that becomes `entries IS NULL AND error IS NULL` in the DB row — same signal, explicit, queryable.
- Some 1828 entries have malformed POS strings (`"v. t. & i."`, `"adj. & n."`). We preserve them verbatim; the frontend renders them as-is. Worth normalizing later, not in this phase.
- `webster1828.json.gz` has multiple entries per headword for words with multiple senses (e.g., `lay` as verb + noun). The existing frontend already handles this via `entries: [{pos, definitions}]`. The DB schema preserves it.

## V. Endpoints

### `GET /api/dict/1828/:word`

```json
{
  "word": "intelligence",
  "entries": [
    {
      "pos": "n.",
      "definitions": [
        "The act or state of knowing; the perception of facts and truth…",
        "Knowledge imparted or acquired by communication; …",
        …
      ]
    }
  ],
  "found": true,
  "stem_matched": null
}
```

**Stem fallback** is server-side now (it was client-side in `useWordData.ts`). The handler:
1. Looks up the literal word.
2. If miss, tries archaic-suffix stripping (`eth`, `edst`, `est`, `ing`, `ed`, `s`) the same way `useWordData.ts` does.
3. If a stem hits, returns the stem's entries with `stem_matched: "obtain"` so the frontend can render `Showing definition of "obtain" for "obtaineth"`.
4. If all miss, returns `{found: false}` with 200 (not 404 — the client wants to render "no 1828 entry" gracefully, not handle an error code).

### `GET /api/dict/modern/:word`

```json
{
  "word": "intelligence",
  "entries": [...] | null,
  "source": "cache" | "fetched" | "none",
  "found": true
}
```

Logic:
1. SELECT from `modern_defs` by word.
2. If row exists AND `entries IS NOT NULL`: return `source: "cache"`, found: true.
3. If row exists AND `entries IS NULL` AND `error IS NULL`: it's a cached 404, return `source: "none"`, found: false.
4. If row exists AND `error IS NOT NULL` AND `fetched_at < 24h ago`: return cached error.
5. If no row OR error is stale: trigger fetch (rate-limited 1/sec via a server-side leaky-bucket), write the result back, return `source: "fetched"`.

**Server-side rate limiter is a singleton.** Concurrent requests for the same word coalesce (singleflight pattern); concurrent requests for different uncached words queue at 1/sec. This is required to be a good citizen of the Free Dictionary API. Document in the README that the API is community-supported and we throttle accordingly.

### `GET /api/dict/tier/:word`

```json
{
  "word": "charity",
  "tier": "A++",
  "study_tier": "A",
  "studies": ["give-away-all-my-sins.md", "art-of-presidency.md"],
  "study_excerpts": ["…"],
  "p4_score": 8,
  "p4_reasons": ["archaic marker", "high frequency", "study-attested differ"]
}
```

Just a row-shape; trivial.

### `GET /api/dict/search?q=:prefix&limit=:n`

The "prefix-search box" UX. Backed by `webster_1828_word_trgm`. Returns both:

```json
{
  "query": "obt",
  "tier_results": [...],     // from tier_words where word LIKE 'obt%' or has trigram match
  "all_1828_results": [...]  // from webster_1828 (full corpus) where word LIKE 'obt%'
}
```

Two lists, surfaced in the UI as a primary section ("Words we've curated") and a secondary section ("Other 1828 words matching"). This is the class-E reach made visible in the search UX.

### `GET /api/thummim/:word`

Reads from `thummim_entries_cache`. If word not present, returns `{found: false}` cleanly (frontend renders "no Thummim entry yet"). Sync of the cache is a separate concern — see `backend-pivot.md §D-BE-THM`.

## VI. Frontend cutover

`frontend/src/composables/useWordData.ts` today:
- Statically imports `tier-words.json`, `definitions-1828.json`, `definitions-modern.json`, `manual-additions.json`.
- Exposes synchronous lookups.

After cutover:
- Static imports remain only for `tier-words.json` (small, drives highlight tier display) AND we keep the `manual-additions.json` shape — actually no, those merge into the DB seed at ingest. Static imports go away.
- Lookups become async (`Promise<Def1828Entry[]>` etc.).
- A small in-memory LRU cache in the composable avoids repeated round-trips for the same word in a session.
- The `tokenize()` function stays client-side; it doesn't need the backend, and the verse-highlight render path already runs over local segments.

**One stewardship fix to land at cutover:** today's `useWordData.ts` `stemMatch` returns a synchronous result. The backend `dict/1828/:word` does its own server-side stem-fallback. The frontend should *stop* doing client-side stem fallback and just pass the raw word — server is the single source of stem truth. Removes ~20 lines of duplicated logic.

## VII. Lazy fetch — friendliness contract with Free Dictionary API

This is the surface most likely to upset an external dependency. Hard requirements:

1. **1 req/sec global rate cap** (not per-user). Implemented as a `time.Ticker` channel that the fetch goroutine drains; all callers wait their turn.
2. **Cache 404s permanently** unless the user explicitly re-requests with `?refetch=1`. A 404 today is a 404 in 3 months.
3. **Cache errors for 24h** (network blips happen).
4. **`User-Agent: 1828-illuminated.ibeco.me/0.1`** — identifiable, contactable.
5. **Daily fetch ceiling** (`MODERN_FETCH_DAILY_CAP`, env-configurable, default 5000): once hit, return cached-only for the rest of the day. Protects against runaway crawl scenarios. Logs a warning when hit.
6. **Optional pre-seed jobs.** A `make warm-modern-cache` target reads the full tier word list + every 1828 headword referenced by recent search queries and warms the cache during off-hours. Out of scope for v1; mentioned so we don't paint into a corner.

## VIII. Decisions

| # | Decision | Default | Stakes |
|---|---|---|---|
| **D-DICT-1** | Ingest full 98k 1828 entries vs only curated | Full | Class-E reach is one of the six wins; partial ingest defeats the point |
| **D-DICT-2** | Server-side stem fallback only (remove client-side) | Yes | One source of truth; reduces duplication |
| **D-DICT-3** | Lazy modern-def fetch enabled in production from day one | Yes | The win is real-user-driven corpus growth |
| **D-DICT-4** | Daily fetch ceiling default | 5000 | Conservative; cheap to raise once observed |
| **D-DICT-5** | Surface "this word is not in 1828" vs hide | Surface explicitly | Honest signal; not every word has a Webster entry |
| **D-DICT-6** | Trigram-search the full 1828 corpus from the search box | Yes | The UX gain is what makes class-E reach visible |
| **D-DICT-7** | Manual additions (today's `manual-additions.json`) migrate to a `tier_words_manual` table or merge into `tier_words` with a `source` column | Merge with `source` column | Single table; no per-source-of-truth fork |

## IX. Verification

After phase ships:
- `curl /api/dict/1828/gainsay` returns the 1828 entry that today's frontend can't find.
- `curl /api/dict/1828/obtaineth` returns the entry for `obtain` with `stem_matched: "obtain"`.
- `curl /api/dict/modern/peradventure` triggers a lazy fetch, writes to DB, returns `source: "fetched"`. A second call returns `source: "cache"`.
- `curl /api/dict/search?q=int` returns tier hits (`intelligence`) AND all-1828 hits (`intemperance`, `interpret`, …).
- Frontend wordcard renders identically to today for tier words; renders 1828 defs for non-tier words that previously showed "no definition".
- Verse explorer's hover-card still works; backend round-trip is < 50ms for cached lookups.

## X. Risks

- **Free Dictionary API outage / takedown.** Cached corpus continues to serve. New words pile up as `error IS NOT NULL` rows; auto-retry after 24h. Document the risk; recommend `MODERN_FETCH_DAILY_CAP=0` as the kill switch.
- **Webster 1828 false matches.** The 1828 has entries for proper nouns (place names), arcane terms, and some that look like modern words but mean utterly different things (`prevent` = "to come before / go before"; not "to stop"). Surface tier matters here — `Showing 1828 entry` is honest; "this is the original meaning" would not be. Mitigation: the existing `intent.yaml` value `illuminate-not-encode` already binds this; the new UI strings should reflect it. Spell it out in user-facing copy.
- **Schema mismatch with future Thummim corpus.** The Thummim entries have multi-grade-level renderings; 1828 doesn't. We keep them in separate tables on purpose — different shape, different lifecycle, different copyright posture.
- **Migration of `manual-additions.json` losing intent.** The today file has comments + structure that imply "these are the words P1 missed." Merging into the DB drops that file-shape narrative. Mitigation: the migration commits the manual-additions JSON content as a labeled migration (`Nxx-seed-manual-additions.sql`), so the provenance is in git history.

## XI. Out of scope

- Multi-edition Webster (1913, 1828 only for now).
- Multi-lingual definitions.
- Audio pronunciations.
- User-contributed corrections / annotations (a real feature later, separate proposal).
- Etymology graphs (the 1828 has etymology lines we currently ignore; we keep ignoring them).
