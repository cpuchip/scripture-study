# Proposal: give every loop the Vivint corpus treatment (compounding pools)

**Status:** draft for council / ratification · **Date:** 2026-06-16 · for Michael
**Pairs with:** the reflect-steward (`.spec/proposals/reflect-steward.md`), the
skills proposal, the `claude -p` harness provider.

## The goal

Vivint compounds; the other loops don't. The reflect-steward's findings publish to
a **searchable, project-tagged knowledge pool** (`stewards.docs`, FTS + vector),
get **deduped** (`intent_source_ledger`), surveyed before re-proposing
(`intent_work_survey` — the council moment), and read back scoped to a project
neighborhood (`pool_search`). That is why a fresh Vivint cycle *builds on* prior
findings instead of repeating them — and why Vera can answer from the pool.

The **book-study**, **video-study**, and **general-research** (ai-news, science)
loops do **not** compound. They write **file digests** (`study/books/`,
`study/yt/`) and stop. No pool, no dedup-memory, no survey, no scoped read, no
persona that can answer from them. This proposal gives them the same treatment so
each becomes a compounding corpus with a face.

## Current state (verified 2026-06-16)

| Loop | Intent | Pipeline | Output today | Pool docs |
|---|---|---|---|---|
| book-digest (hourly) | book-study | book-digest | file in `study/books/` (9 so far) | 0 |
| playlist-digest (~6h) | video-study | playlist-digest | file in `study/yt/` (6 so far) | 0 |
| ai-news-7am (daily) | general-research | research-summary | file/digest | 0 |
| science-news-weekly | general-research | research-write | file/digest | 0 |
| **vivint-reflect (3h)** | **vivint** | **planning** | **pool docs + proposals** | **8 ✓** |

The machinery Vivint uses is already **generic** (built during the reflect-steward
loop-closing): `on_maturity_verified` pool-publish via `import_doc` +
`project_association`; `intent_source_ledger` + `intent_sources_recent/record`;
`intent_work_survey`; `pool_search` + `project_neighborhood`. The other loops just
aren't wired to it.

## The treatment (what each loop gains)

For each loop's intent, register a **project** and turn on the four pieces:

1. **Pool-publish.** The digest pipeline's terminal stage sets
   `auto_materialize_on_verified` + `project_association` so each finding
   (a book digest, a video evaluation, an AI-news brief) lands in `stewards.docs`
   tagged to its project — searchable by FTS + vector, not just a file. (Files can
   stay too; the pool is additive.)
2. **Dedup ledger.** Record each gathered source (book title, video id, news
   source+date) in `intent_source_ledger` so the loop doesn't re-digest what it
   already has, and a stale source is fair to re-gather after the freshness window.
3. **Council survey.** The digest prompt calls `intent_work_survey` first — see
   what's already pooled/in-flight before adding more (kills duplicate digests).
4. **Scoped read + neighborhoods.** `pool_search` resolves the caller's project;
   `project_neighborhood` lets chosen pools cross-pollinate. Proposed default
   neighborhood: **books ↔ ai ↔ video cross-pollinate** (research themes recur
   across them); **vivint stays walled** (work, no bleed) — exactly the line you
   drew earlier.

The payoff beyond compounding: **a persona per intent** (like Vera for Vivint) — a
"books librarian," an "AI-research analyst," each reading its pool. That's where
this meets the skills + persona work.

## Per-loop specifics

- **book-study → project `books`.** Each book digest → a pool doc (kind `book`),
  ledger key = the book slug. A persona ("the librarian of the books we've read")
  can then answer "what does Stoicism say about X across the books we've digested?"
- **video-study → project `video` (or fold into `ai`).** Each video eval → a pool
  doc (kind `video`), ledger key = video id. **Open question:** is video its own
  project, or part of `ai`? (Most videos are AI-research; leaning fold into `ai`.)
- **general-research → project `ai`.** ai-news briefs + science digests → pool docs
  (kind `news`/`research`), ledger key = source+date. A persona = the AI-research
  analyst (the original telos: "AI ideas and research applicable to us").

## Decisions for council (D1–D6)

- **D1** — project taxonomy: `books`, `ai`, `video` separate, or consolidate video
  into `ai`? (Lean: `books` + `ai`, video folds into `ai`.)
- **D2** — neighborhoods: books ↔ ai (↔ video) cross-pollinate, vivint walled?
- **D3** — backfill: import the existing file digests (9 books, 6 videos) into the
  pool, or pool only from now forward? (Lean: backfill — the corpus already exists.)
- **D4** — a persona per intent now (books-librarian, ai-analyst), or pool first +
  faces later? (Lean: pool first, faces as a fast follow — reuse Vera's `analyst`
  family.)
- **D5** — cost: pooling adds embeddings per digest. Acceptable? (Local nomic
  embeddings via LM Studio = ~free; the watchman guard caps autonomous spend.)
- **D6** — should these loops also gain `request_research` (queue a gap), like Vera?
  (Lean: yes — same gated-proposal pattern.)

## RATIFIED 2026-06-16 (Michael, via ask-tool) + build findings

**Decisions:** D1 = `books` + `ai` (video folds into `ai`). D2 = books↔ai
cross-pollinate, vivint walled. D3 = **backfill all** (9 books + 25 yt files). D4
= pool first, personas a fast follow (reuse Vera's `analyst` family). D5 = cost
fine (local nomic + the watchman guard). D6 = loops get `request_research` too.

**Build findings (grounded in the running stack + 08-gates):**
- Projects `books`, `ai`, `vivint` already registered. ✓
- The digest loops DO reach `maturity=verified` (book-digest 50, playlist-digest
  9, research-write 5). ✓
- **The catch — pool-publish is coupled to file-materialize.** In
  `on_maturity_verified`, `import_doc`-to-pool fires only inside
  `IF auto_materialize AND NEW.file_destination IS NOT NULL`. But the digester
  *agents write their own files via fs tools* (book-digest has no
  `file_destination_template`, `auto_materialize=false`). So you cannot just flip
  `auto_materialize` — that path renders a destination + `enqueue_work_item_file`,
  a SECOND file write. **The clean fix: decouple pool-publish from
  file-materialize** — pool-publish any verified digest's content
  (`extract_work_item_file_content`) regardless of `file_destination`. That's a
  careful **core change to `on_maturity_verified`** (the clobber-prone fn; needs
  rebuild + virgin-smoke + the overlay-clobber-check) — not an overlay flip.
- **Project tagging:** intents are `book-study`/`video-study`/`general-research`;
  projects are `books`/`ai`. `work_item_create` defaults project = intent slug
  only if a matching project exists (so these tag NULL today). Add an
  intent→project map + an **additive** BEFORE-INSERT trigger on `work_items` that
  fills `project_association` when NULL (book-study→books, video-study→ai,
  general-research→ai). Additive = no core re-author, clobber-check-safe.
- **Backfill:** import the 9 `study/books/*.md` (→books) + 25 `study/yt/*.md`
  (→ai) via `import_doc` + tag. A one-time script.
- **Neighborhoods:** `project_neighborhood` (books,ai)+(ai,books); vivint no rows.

**Build order (a focused pass, not the marathon tail):** (1) the core decouple in
`on_maturity_verified` + rebuild + smoke + clobber-check; (2) the intent→project
map + additive trigger; (3) flip the two digest pipelines to pool-publish; (4)
neighborhoods; (5) backfill script; (6) verify a fresh book/video digest pools to
the right project + a survey de-dups. Then P1 personas + request_research.

## Phasing

- **P0** — register the projects + neighborhoods; wire pool-publish + ledger +
  survey into the three digest pipelines; backfill the existing digests (D3). Prove
  a fresh book/video/ai run compounds (reads prior, doesn't duplicate) like Vivint.
- **P1** — a persona per intent (reuse the `analyst` family) + `request_research`.
- **P2** — cross-pollination tuning (what the neighborhoods actually surface).

This is the generic capability already in core; P0 is mostly operator overlay
(project rows, neighborhood rows, the three pipelines' flags) + a backfill script.

## BUILT 2026-06-16 (P0 shipped, clean-image-proven)

The decouple turned out cleaner than feared — **no new pipeline flag**. The single
signal is `project_association`:

- **CORE `08-gates.sql`** — `on_maturity_verified`: the `import_doc`-to-pool block
  moved OUT of the `IF file_destination IS NOT NULL` nest into its own block gated
  `(v_auto_mat AND file_destination IS NOT NULL) OR project_association IS NOT NULL`.
  So any project-tagged verified work pools (digesters write their own files; this
  no longer requires one), the prior reflect/planning pooling is preserved, and a
  content guard skips empty extractions. `extract_work_item_file_content` already
  falls back to the final stage output → works for digesters; `import_doc` takes a
  NULL file_path and dedups by slug.
- **CORE `25-corpus.sql`** (new) — `intent_project_map` (empty in core) + an
  ADDITIVE BEFORE-INSERT trigger `fill_project_association` that fills
  `project_association` when NULL from the map, ONLY if the mapped project exists
  (work_items.project_association FKs projects(slug) — guarded like 09's tagging).
  No core function re-author → clobber-safe.
- **OVERLAY `corpus-pools.sql`** — ensures projects books/ai exist; seeds the map
  (book-study→books, video-study→ai, general-research→ai); neighborhoods
  (books↔ai); vivint walled. Idempotent/non-clobbering (no-op on live).
- **`cut3-restore-core-finals.sql`** — its verbatim `on_maturity_verified` copy was
  updated in lockstep (or the clobber-check would flag it as a stale revert).
- **`parity/backfill-corpus.py`** — one-time import of study/books (→books) +
  study/yt (→ai) into the pool. Run at live deploy.

**Proven:** virgin-smoke OK 10 (00→25; the decouple — a project-tagged verified
work_item with NO file_destination pools; the trigger fill; FK-safe when the
project is missing); overlay-clobber-check PASS (3 overrides / 0 clobbers —
`on_maturity_verified` not flagged, so cut3 matches core); a real book-study
work_item routes to `project=books`; `project_neighbors` = books↔ai, vivint walled.
**Live-DB deploy + the backfill batched with skills #174** (one image rebake +
ordered apply). P1 (per-loop personas + request_research) is the fast-follow.
