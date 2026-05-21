---
title: 1828-illuminated — scripture corpus backend
date: 2026-05-20
status: proposed
workstream: WS7
parent: backend-pivot.md
purpose: >
  Ingest the public-domain scripture corpus into the 1828 backend's
  Postgres so the verse explorer renders local text with inline 1828-tier
  highlights, and the verse search returns "every occurrence of X" in a
  single round-trip.
---

# Scripture Corpus Backend

## I. The problem this surface solves

Today's verse explorer has two modes:

1. **Iframe to churchofjesuschrist.org** — works, but breaks our render layer (we can't highlight tier words inside someone else's iframe), can't be styled, can't be linked-into, costs the user a network hop to a third party.
2. **Paste your own text** — works for the curious, doesn't scale to "open D&C 84 and read it with 1828 lens applied."

Michael's pivot solves both by **rendering scripture text we serve ourselves**, with tier-word highlights inline, plus cross-links to the relevant study documents. The blocker has been copyright on the canon text. The bcbooks corpus is the proposed unblock — pending the D-BE-COPYRIGHT decision in [backend-pivot.md §III](backend-pivot.md).

This proposal assumes a public-domain (or community-resourced) corpus is selected. If D-BE-COPYRIGHT lands on option C (iframe-only-for-canon), this proposal's scope shrinks: keep the search endpoint, drop the render endpoint, frontend retains iframe.

## II. Source corpus options (the D-BE-COPYRIGHT branches)

**Option A — bcbooks/scriptures-json (2013 LDS edition):**
- 5 files, ~13.6MB raw JSON, ~2.3MB zipped.
- Already parsed by `external_context/scriptures-mcp/internal/scripture/service.go` — exact reuse possible.
- Refs match the modern verse numbering readers expect.
- *Risk:* upstream is community, not church-licensed; takedown possible.

**Option B — strip-to-PD (per-volume PD sources):**
- KJV (OT + NT): 1769 Cambridge Standard Text, available from Project Gutenberg + others. Verse numbering matches LDS edition for the most part; italicized words may differ.
- Book of Mormon: 1830 first edition (PD). Chapter divisions differ from 1981 LDS edition (Orson Pratt's 1879 versification is what we use today). Need a chapter-mapping table or accept 1830 chapters.
- D&C: section text mostly PD (revealed pre-1844). Modern section ordering + study apparatus is not. Risk: D&C 138 (1918) and the 1981 OD-1/OD-2 + section 137/138 are within copyright. Workable: include sections 1-136 as PD; flag 137-138 + OD as church-published, link out.
- PGP: Joseph Smith Translation excerpts + Moses + Abraham + JS-History + Articles of Faith. Mixed PD status; most PD, but the modern compiled volume is church-curated.

**Option C — keep iframe for canon:** no corpus ingest; this proposal collapses to the search endpoint operating over the bcbooks corpus indexed *server-side only*, with the rendered output being a reference + snippet (a few words of context), not a full verse. Acceptable substring-snippets are fair use; full verses are not. Frontend behavior: when a search result is clicked, open the iframe / external link.

**Recommendation:** option A if Michael's risk tolerance accepts it (cleanest implementation, fastest delivery). Option C if not (preserves the existing copyright posture cleanly; loses the inline-rendered verse explorer). Option B is the most work for ambiguous gain — the 1830 BoM chapter divisions in particular will confuse readers who paste a modern reference.

**RATIFIED 2026-05-20 — option D (hybrid).** Use bcbooks as the source for the verse text + modern numbering, BUT on ingest the function MUST strip:
- Footnote markers (e.g. `[a]`, `*`, superscript letters)
- Chapter / section headings (the publisher-curated paragraph above each chapter)
- Bracketed publisher additions and study-aid apparatus
- Italic markers (KJV's added-for-clarity italics) — accepted as flat text per §IV

Verse text + verse number + chapter number + book identity is what survives. Every rendered surface in the frontend MUST also surface a tabbed-iframe breakout to churchofjesuschrist.org (the `projects/cpuchip.net/src/components/ScripturePanel.vue` pattern) so any reader who wants the full apparatus has one-click access to the canonical source.

This decision settles the §II branch: the ingest path follows option D's stripping rules; the corpus ships from bcbooks; the frontend retains the breakout-to-iframe affordance per scripture render.

## III. Schema (excerpted from backend-pivot.md §V)

```sql
CREATE TABLE scripture_books (
  id              SERIAL PRIMARY KEY,
  volume          TEXT NOT NULL,     -- 'ot' | 'nt' | 'bofm' | 'dc' | 'pgp'
  abbr            TEXT NOT NULL UNIQUE,
  name            TEXT NOT NULL,
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
-- Requires: CREATE EXTENSION IF NOT EXISTS pg_trgm; (in migration 00)
```

**Why `abbr`?** So that workspace-style refs (`dc/84:38`, `bofm/1-ne/3:7`) work end-to-end without parsing the human name on every request. The `abbr` values mirror the directory names under `gospel-library/eng/scriptures/`:

| Volume | Abbr examples |
|---|---|
| ot | `gen`, `ex`, `lev`, `num`, `deut`, `josh`, `judg`, `ruth`, `1-sam`, … |
| nt | `matt`, `mark`, `luke`, `john`, `acts`, `rom`, `1-cor`, `2-cor`, … |
| bofm | `1-ne`, `2-ne`, `jacob`, `enos`, `jarom`, `omni`, `w-of-m`, `mosiah`, `alma`, `hel`, `3-ne`, `4-ne`, `morm`, `ether`, `moro` |
| dc | `dc` (single book of sections), `od` (Official Declarations) |
| pgp | `moses`, `abr`, `js-m`, `js-h`, `a-of-f` |

This is the **cross-link payoff:** when the dictionary surface returns "this word is studied in `study/intelligence.md`," and that study links to `gospel-library/eng/scriptures/dc-testament/dc/93.md`, the 1828 frontend can route directly to `/scripture/dc/93:36` and render the verse with highlights without anything translating between the two ref shapes.

## IV. Ingest path

Single Go function, runs at boot if `scripture_verses` is empty:

```go
// backend/internal/seed/scriptures.go
//
//go:embed data/scriptures.zip   ← copied at build time from external_context/scriptures-mcp/internal/scripture/data/scriptures.zip
//                                   (or option B's hand-curated PD bundle)
var scripturesZip []byte

func SeedCanon(ctx context.Context, db *sql.DB) error {
    // 1. Read zip → parse to [{book, chapters: [{chapter, verses: [{verse, text}]}]}]
    // 2. Build the books table with our `abbr` mapping (a hand-maintained
    //    map[upstreamBookName]volumeAndAbbr in the same file — short, stable,
    //    auditable)
    // 3. INSERT … ON CONFLICT (abbr) DO NOTHING for books
    // 4. INSERT … ON CONFLICT (book_id, chapter) DO NOTHING for chapters
    // 5. INSERT … ON CONFLICT (chapter_id, verse) DO UPDATE SET text = EXCLUDED.text
    //    ← so a re-seed picks up text corrections without throwing
    // 6. Run ANALYZE on the three tables
    return nil
}
```

**Stewardship fix during ingest (per Michael's directive):**
- bcbooks/scriptures-json verse references use modern numbering. We trust them, but we record the source SHA in a `scripture_corpus_meta` row so a future audit can compare.
- Some bcbooks verses contain HTML entities (`&mdash;`, `&rsquo;`). Normalize on ingest.
- Italic words in KJV (added-for-clarity) are not marked in bcbooks. We don't reconstruct them; we accept the flatter text.

**The abbr map is the only piece of "data engineering" required.** It looks like:

```go
var volumeAbbr = map[string]struct{ Volume, Abbr string }{
    "Genesis":            {"ot", "gen"},
    "Exodus":             {"ot", "ex"},
    "1 Nephi":            {"bofm", "1-ne"},
    "Doctrine and Covenants": {"dc", "dc"},
    "Moses":              {"pgp", "moses"},
    // … ~80 entries total
}
```

Hand-curated, ~80 lines, stable across decades, audit-friendly. Adding new books (or correcting names) is a one-line PR.

## V. Endpoints

### `GET /api/scripture/:ref`

Parse the ref string (`1 Nephi 3:7`, `1-ne/3:7`, `dc/84:38`, `John 3:16-17`). Both human and abbr forms accepted. Returns:

```json
{
  "ref": "1 Nephi 3:7",
  "abbr_ref": "1-ne/3:7",
  "book": "1 Nephi",
  "chapter": 3,
  "verse_start": 7,
  "verse_end": 7,
  "verses": [
    { "verse": 7, "text": "And it came to pass that I, Nephi, said unto my father: …", "segments": [...] }
  ]
}
```

With `?highlight=1`, `segments` is populated from `verse_highlights_cache` (computed on miss). Without, `segments` is omitted and the client can tokenize itself (slower, but functions if the cache is cold).

### `GET /api/scripture/chapter/:ref`

`:ref` is `book chapter` (`John 3` or `john/3`). Returns the full chapter the same way. `?highlight=1` works.

### `GET /api/scripture/search`

Query params: `q` (required), `limit` (default 20, max 100), `volume` (optional filter).

```json
{
  "query": "intelligence",
  "results": [
    {
      "ref": "Abraham 3:19",
      "abbr_ref": "abr/3:19",
      "text": "And the Lord said unto me: These two facts do exist…",
      "snippet": "…there are two spirits, and one being more <em>intelligent</em> than the other…",
      "rank": 0.0381
    }
  ]
}
```

Implementation: `text_tsv @@ websearch_to_tsquery('english', $1)` for the primary path, ranked with `ts_rank_cd`. Falls back to trigram `text % $1` for short queries (≤3 chars) where FTS is poor. The `<em>` markers in snippet come from `ts_headline`.

### `GET /api/scripture/word-study/:word`

The "every occurrence" view. Tier-word context. Returns:

```json
{
  "word": "intelligence",
  "tier": "A++",
  "occurrences": [
    { "ref": "D&C 93:36", "abbr_ref": "dc/93:36", "text": "The glory of God is intelligence…" },
    …
  ],
  "study_cross_refs": [
    { "study": "intelligence.md", "excerpt": "…" },
    …
  ]
}
```

Combines `scripture_verses` (LIKE-and-stem-match for occurrences) + `tier_words` (for the cross-refs). Behind a cache key `(word, algo_version)` because this is expensive on cold call.

## VI. Verse explorer integration

Today's frontend (`frontend/src/views/VerseExplorer.vue`, not read in this planning session but inferred from `useWordData.ts`):
- Demo verses are paraphrased text in a static JSON.
- Pasted text is tokenized client-side.
- Iframe to churchofjesuschrist.org when a real ref is wanted.

After this proposal ships:
- Demo verses become *real* references (`dc/84:38`, `1-cor/13:1-13`, `dc/93:36`).
- The frontend calls `/api/scripture/:ref?highlight=1` and renders verses inline.
- Hover/click on a highlighted word still uses the existing `selectedWord` reactive; `useWordData.ts` switches to `/api/dict/1828/:word` and `/api/dict/modern/:word` (see `dictionary-backend.md`).
- Iframe code path becomes optional / removable, depending on D-BE-COPYRIGHT.

The frontend's tokenize logic stays as a fallback for pasted text — that path needs no backend.

## VII. Decisions

| # | Decision | Default | Stakes |
|---|---|---|---|
| **D-SC-1** | Default for `?highlight` (always on / opt-in) | Opt-in | Avoids paying tokenize cost on API calls that don't need it (e.g. word-study reverse lookups) |
| **D-SC-2** | Include `text_tsv` `english` config or `simple` for KJV-archaic stemming | **RATIFIED 2026-05-20:** `english` stemmer + custom archaic-suffix expansion layer at search time. The server-side handler mirrors `useWordData.ts`'s `ARCHAIC_SUFFIXES` (-eth, -edst, -est, -ing, -ed, -s) so that `suffereth` queries match `suffer` rows. Implementation: do the suffix-strip before constructing the tsquery, emit an OR-tsquery (`suffer | suffereth`) for fuller recall. Document in the dictionary-backend too — the stem layer is shared with the dictionary's `dict/1828/:word` server-side fallback. | Settled |
| **D-SC-3** | Embedded zip in backend image vs mounted volume | Embedded | Self-contained image; one-deploy reseeds. Volume mount is needed only if we want hot-swap corpora without rebuild — not a v1 need. |
| **D-SC-4** | `scripture_corpus_meta` table for source SHA, ingest timestamp, license string | Yes | Cheap; serves audit + the `respect-the-canon` value commitment. |
| **D-SC-5** | Apocrypha / non-LDS-canon books from KJV | No | bcbooks doesn't ship them; out of scope. |

## VIII. Verification

After phase ships, this is true:
- `curl /api/scripture/dc/84:38?highlight=1` returns the verse with `segments` populated where tier words appear.
- `curl /api/scripture/search?q=intelligence` returns ≥3 hits, ranked, with snippet highlighting.
- `curl /api/scripture/word-study/charity` returns occurrences across at least BoM + NT, plus cross-refs to `give-away-all-my-sins.md` (if `charity` is in `tier_words.studies`).
- VerseExplorer.vue, with no client-side tokenization, renders the same highlights as before, sourced from the server cache.
- Re-deploy clears no data (named volume survives), and a Postgres restart re-loads in seconds (no re-ingest required).

## IX. Risks

- **D-BE-COPYRIGHT must be answered first.** If it lands on C, this proposal degrades to the search-only shape.
- **The `english` text-search config doesn't stem -eth.** A user searching for "sufferest" won't find "suffer". Mitigation: maintain a small stem-expansion layer in the search handler that mirrors `useWordData.ts`'s `ARCHAIC_SUFFIXES`.
- **Encoding edge cases.** Some bcbooks verses use Unicode em-dashes; our `text_tsv` and trigram indexes handle them, but the cited snippet rendering needs to escape correctly for HTML. Test pass needed.
- **`pg_trgm` adds index size.** ~100MB for the full corpus is the order of magnitude. Acceptable on a $5 VPS, but measured.

## X. Out of scope (for this proposal)

- pgvector / embeddings for semantic search (carry-forward in `backend-pivot.md §XII`)
- Footnote / cross-reference apparatus (church-published, copyright-encumbered, separate proposal)
- Topical Guide / Bible Dictionary integration (separate corpus, separate proposal)
- KJV italics restoration (would need a different upstream source)
- Multi-translation support (NIV, NRSV, etc. — out of scope; the Restoration project uses KJV)
