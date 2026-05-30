---
title: UX 1-2 punch — always-visible unified search + three-mode flows
date: 2026-05-29
status: search bar SHIPPED + browser-verified; connective actions + full-text deferred
workstream: WS7
---

# UX 1-2 punch

The 1828 finish thread's UX item (Sabbath carry-forward, 2026-05-23). Design
pass + first implementation done 2026-05-29.

## Diagnosis (current UX)

- **Search is a destination, not a tool.** It lived only inside `/word`
  (terms) and `/verse` (scripture) views. To look anything up from Home, a
  word page, or mid-reading you had to navigate to a search view first AND know
  which one. Friction on the most common action.
- **The three modes are invisible.** study (StudyTreePanel + breadcrumbs +
  WordStudy), explore (WordSearch + VerseExplorer + Dictionary), present
  (Present.vue) are emergent from surfaces, never named or linked.

## Ratified decisions (2026-05-29)

- **D-UX1: unified smart search input** (not a Cmd-K palette, not a scope
  toggle). One header input that detects intent and routes.
- **D-UX2: modes stay implicit + connective actions** (no explicit mode
  switcher). Smooth the transitions instead.
- **D-UX3: this session = the search bar.** Connective actions + full-text
  follow.

## Shipped (this session) — `SearchBar.vue` + header wiring

- Always-visible search in the header, `hidden sm:block`, hidden on `/present`
  (distraction-free, same posture as the study tree per D-ST-10).
- Intent detection + routing:
  - **Reference** ("1 Ne 3:7", "John 3:16", "Alma 32") → parsed against `CANON`
    (book name OR abbr, normalized; chapter bounds-checked) → Verse Explorer
    with `{mode:canon, v, b, c, r}` query params.
  - **Word** ("charity") → `word-detail` (which does the full class-E + stem
    lookup). Instant tier-prefix suggestions from `useWordData.searchPrefix`.
- Keyboard: `/` and `Cmd/Ctrl-K` focus from anywhere; ArrowUp/Down move the
  highlight; Enter activates; Esc closes.
- **Neighboring fix (1828 "fix-don't-name" directive):** VerseExplorer's
  `watch(route.query)` now **re-fetches** the passage on URL change (guarded by
  a `loadedKey` against fetchCanonChapter's own `syncRouteFromState` replace).
  Previously it only synced selector state — so a same-page ref search (and,
  latently, browser back/forward between chapters) updated the dropdowns but
  never reloaded the verses. Now both work.
- **Browser-verified** (per the project's "real browser, not the build exit
  code" rule): rebuilt the frontend container against the live backend; `1 ne
  3:7` → dropdown → Enter → correct URL → verse loaded; `charity` → "Look up"
  suggestion. Build clean (vue-tsc + vite).

## Deferred (next passes)

1. **Full-text scripture search.** `/api/scripture/search?q=` exists on the
   backend but there is **no results surface** in the frontend. The unified
   search currently handles reference + word (both have destinations); free
   text that isn't a reference falls through to the word path. A
   ScriptureSearchResults view + wiring the dropdown's "Search scriptures for
   '…'" option is the next build.
2. **Connective actions (D-UX2).** From any rendered verse/word: "study this"
   (add to the tree) and "present this" (jump to fullscreen) buttons, so the
   three modes link instead of siloing. Present should be reachable from a
   passage.
3. **Fuzzy / partial book matching** in the reference parser (today it requires
   an exact name-or-abbr match after normalization; "nephi" alone is ambiguous
   and not matched). Low priority.
4. **Mobile search** (currently `hidden sm:block`; small screens use the Word
   Search view). A mobile affordance (e.g. a search icon opening a sheet) later.

## Files

- `frontend/src/components/SearchBar.vue` (new)
- `frontend/src/App.vue` (header wiring + `showSearch`)
- `frontend/src/views/VerseExplorer.vue` (watch re-fetch + `loadedKey` guard)

## Remaining to reach production

Local verification done. **Deploy is Michael's action** — 1828 is part of the
workspace repo (no auto-deploy on commit); production at 1828.ibeco.me updates
via the Dokploy Compose project rebuild.
