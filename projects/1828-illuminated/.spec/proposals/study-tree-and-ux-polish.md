---
title: 1828-illuminated — study tree + canon-browse UX polish
date: 2026-05-21
status: proposed
workstream: WS7
parent:
  - backend-pivot.md
purpose: >
  Add a branching study-tree that tracks both word-card navigation
  ("word train") and scripture-passage navigation as a unified graph
  the reader can traverse, branch off, and toggle between paths.
  Plus the Phase 5.5 polish bundle — pg18 bump, route-back-button
  for canon browse, per-verse rendering, single-verse + verse-range
  selectors.
---

# Study Tree + Phase 5.5 UX Polish

## I. Vision — "five-dimensional time-travel chess"

Michael's naming: *you study a chain, but you can click a bubble earlier and click a new word — that branches — and you can toggle fast between the various branches.*

A real study tool. The cross-reference structure of scripture made visible. Not just a breadcrumb. A graph the reader navigates with intent.

**The critical move:** unify word-card navigation AND scripture-passage navigation into ONE tree, not two. So a node might be the word *school*, the next a verse *Helaman 12:3* (the reader clicked through), the next another word *endure* clicked inside that verse, the next a *render* of that verse via LLM, the next a different word branched off two nodes back. The path *across the boundary* between dictionary and canon is where the real insight lives, and a unified tree captures it.

## II. What a node is

```ts
type StudyNode = {
  id: string                  // ulid; unique across the whole tree
  kind: 'word' | 'verse' | 'chapter' | 'render' | 'note'
  parentId: string | null     // null only for the root(s); a tree may have multiple roots if reader opens unrelated chains
  createdAt: number           // ms epoch; node never moves once created
  label: string               // short human-readable ("school", "D&C 88:40", "Render: …")
  payload:                    // kind-discriminated data
    | { kind: 'word'; word: string; stemMatched?: string }
    | { kind: 'verse'; abbrRef: string; humanRef: string; verse: number; text: string }
    | { kind: 'chapter'; abbrRef: string; humanRef: string; verseCount: number }
    | { kind: 'render'; sourceText: string; modernized: string; model: string }
    | { kind: 'note'; body: string }
}
```

**Why this shape:**
- `kind` is the discriminant — the renderer picks a card style per kind.
- `parentId` is the ONLY edge; no separate `children` array (compute children from `nodes.filter(n => n.parentId === id)` — keeps mutations atomic).
- `createdAt` lets us order siblings chronologically (the order in which the reader branched).
- `payload` is denormalized — when a node renders, it doesn't have to refetch the verse text or the 1828 entry. The tree is the source of truth for what was being studied at that moment.

**Multiple roots are allowed.** If the reader opens the tool, walks a word chain, then comes back later and starts at a different scripture, that's a second root in the same tree. Don't force a single root.

## III. Where the tree lives

| Surface | Choice | Reason |
|---|---|---|
| **Per-tab** | Reactive module state in a new `useStudyTree.ts` composable | The active tree the reader is building right now |
| **Browser persistence** | `localStorage` with key `study-tree-v1` | Survives refresh, doesn't survive cross-device |
| **Named saving** | NOT in v1 — but design accommodates | Future: "Save tree as 'D&C 84 priesthood study'" → server-side via `/api/study-tree/:slug` |
| **Multiple trees** | NOT in v1 — single active tree | Future feature; would need a tree-list UI |

Three reads of `localStorage` consideration:
- The tree can grow large (deep studies → hundreds of nodes with verse text payloads). `localStorage` caps at ~5MB; at ~500 bytes/node average we'd hit it at ~10,000 nodes. Plenty for v1.
- Server-side persistence is the right future move, but it crosses into BYOK or anonymous-tree territory — defer.
- If the tree gets unwieldy, surface a "Trim older roots" button rather than aggressive auto-eviction.

## IV. The UX — three surfaces

### IV.1. The tree panel (new component, side panel)

A collapsible right-side panel (`StudyTreePanel.vue`) that's available on every page where study happens — WordDetail, WordSearch, VerseExplorer, Dictionary, Present.

**Layout:**
- **Top:** "Study tree (N nodes)" + a "Start fresh" button (confirms before wiping)
- **Middle:** the tree itself, rendered top-down or as nested cards
- **Bottom:** "Export as markdown" (one-click, copy-to-clipboard)

**Tree visualization options:**
- **A. Indented list** — each node a card with title + meta, children indented. Simple, works in narrow panel. Default.
- **B. Side-scrolling graph** — nodes as bubbles connected by lines. Prettier but expensive to build with d3/svelte-flow. v2.
- **C. ASCII-style tree** — pure CSS, tab-indented with `├─` and `│` glyphs. Charming, very compact. Tie-break candidate with A.

**Recommend: A for v1.** Indented list with collapse/expand per node, click to navigate to that node's content, double-click to fork a new branch from it.

### IV.2. The "you are here" marker

Whatever node the reader is currently viewing gets a highlighted state. Navigating away updates the marker. The tree panel always tells you where you are.

### IV.3. Branching mechanics

The hard part. When does a click create a NEW node vs navigate to an EXISTING node?

**Proposed rule:**
- **From an active node, clicking forward** (any tier-word link, any verse link, any "Open chapter" button) **creates a new child node** under the active one, and becomes the new active node.
- **Clicking a node in the tree panel** (or hitting back/forward) navigates to that existing node — does NOT create a duplicate.
- **From an existing node, clicking forward** creates a new child under THAT node — i.e. branches.
- **Idempotency:** if a node with the same `(parentId, kind, payload-identity)` already exists, REUSE it instead of creating a duplicate. So clicking `school` twice in a row from the same parent gives one node, not two.

This is the time-travel-chess move: a reader can rewind to any earlier node, click a different forward link, and a new branch forms. Toggling between branches is just selecting different leaves of the tree.

### IV.4. Cross-domain navigation (the real prize)

A node's kind doesn't constrain what its child can be:
- **word → verse:** clicking a verse-occurrence link inside a word card creates a verse child
- **verse → word:** clicking a tier-word inside a rendered verse creates a word child
- **chapter → render:** clicking "Render in modern English" creates a render child
- **render → word:** clicking a word inside the rendered text creates a word child

This is what unlocks "five-dimensional" study. The reader can walk: *school* → its 1828 entry → "this appears in Helaman 12:3" → that verse → click *endure* in the verse → endure's entry → click *priesthood* in endure's etymology → branch back two levels → click a different occurrence link → ...

## V. Decisions for ratification

| # | Decision | Default | Stakes |
|---|---|---|---|
| **D-ST-1** | Unified tree (word + verse + chapter + render in one) vs separate word-train + scripture-history | **Unified** | Cross-domain insight depends on it. Single tree, single export. |
| **D-ST-2** | Tree visualization style (A indented / B graph / C ASCII) | **A indented list** | Simple, narrow-panel-friendly. Graph deferred to v2. |
| **D-ST-3** | Persistence layer at v1 | **localStorage only** | Server-side after BYOK trees stabilize |
| **D-ST-4** | Multiple concurrent trees | **No (v1 has one active tree)** | Saved-trees feature deferred |
| **D-ST-5** | Idempotency rule for repeat-clicks under same parent | **Reuse existing node, don't duplicate** | Keeps tree clean; reader can't accidentally fork into N duplicates |
| **D-ST-6** | Auto-create root vs require explicit "start a chain" | **Auto** | Lower friction; the first click anywhere becomes a root |
| **D-ST-7** | "Start fresh" confirmation copy | **"This will clear N nodes from your study tree. Continue?"** | Tree wipes are destructive; require confirm |
| **D-ST-8** | Export format | **Markdown with indented bullets + verse blockquotes + word-card excerpts** | Honest carry-over to journaling |
| **D-ST-9** | Tree panel default state on first visit | **Collapsed, with a small "Study tree (0)" pill in the corner** | Don't shove a feature in their face; let them discover |
| **D-ST-10** | Should the tree appear on Present.vue (the fullscreen tablet view)? | **No** | Present is for distraction-free reading; tree would clutter |

## VI. The Phase 5.5 polish bundle (independent of the tree)

These ride along but don't depend on the tree work. Build first; the tree builds on top.

### VI.1. pg18 bump for i1828-db

- D-BE-2 ratified pg17-alpine; substrate (pg-ai-stewards-dev) is on pg18.3. No reason 1828 lags.
- One-line image change in `docker-compose.yaml`: `image: postgres:17-alpine` → `image: postgres:18-alpine`.
- Recreate the container; seed data re-ingests automatically at first boot (Go backend's `boot-seed` path).
- Confirm with `docker exec i1828-db psql -U i1828 -d i1828 -c "SELECT version();"`
- ~30s downtime, no data loss (seed is the source of truth, lazy-fetched modern-defs accumulated count is small enough to rebuild).

### VI.2. Route-back-button for canon browse

Current bug: `canonVolumeId`, `canonBookAbbr`, `canonChapter`, `singleVerse` are all in-memory refs, never pushed to `route.query`. Browser back/forward skip past all selections.

- On every selection change (or on Open-chapter click), `router.replace({ name: 'verse-explorer', query: { mode: 'canon', v, b, c, s? } })`.
- On component mount, sync refs FROM `route.query` (initial state), then watch the query to handle back/forward.
- Same pattern we used for WordDetail — bug-fix shape inherited.

### VI.3. Per-verse rendering in canon browse

Current: `verses.map(v => `${v.verse} ${v.text}`).join(' ')` collapses the whole chapter into one paragraph. Reads poorly.

- New component `VerseList.vue` that takes a `verses: VerseRow[]` array and renders each as a paragraph with verse-number as a small superscript or inline pill
- Each verse-paragraph passes through the existing `HighlightedText` tokenizer for tier-word highlights
- Optional: hover-handle "🔗" affordance per verse to copy the verse ref (`dc/93:36`) to clipboard

### VI.4. Single-verse + verse-range selector

Backend already accepts both — confirmed live:
- `/api/scripture/dc/93:36?highlight=1` → single verse
- `/api/scripture/dc/93:36-37?highlight=1` → range

Add a fourth selector in canon mode:
- A "Range" input (text or two number inputs) — empty means "whole chapter"
- Parse: empty → fetch chapter endpoint; "36" → fetch /dc/93:36; "36-40" → fetch /dc/93:36-40
- URL state: `?mode=canon&v=dc&b=dc&c=93&r=36-40`

## VII. Phases (build order)

| Phase | Scope | Depends on |
|---|---|---|
| **5.5a** | VI.1 pg18 bump + smoke that backend still serves | — |
| **5.5b** | VI.2 route-back + VI.3 per-verse render + VI.4 verse-range selector — all in VerseExplorer.vue + new VerseList.vue | 5.5a |
| **5.5c** | useStudyTree composable + StudyTreePanel.vue (indented-list view) + integration on WordDetail + VerseExplorer + WordCard | 5.5b |
| **5.5d** | Cross-domain wiring — word→verse + verse→word + render→word + render→verse navigation creates tree nodes idempotently | 5.5c |
| **5.5e** | Tree export as markdown + "start fresh" button + persistence to localStorage | 5.5c |

Each phase commits cleanly and the surface stays usable in between.

## VIII. Risks

- **localStorage scope blowup.** Verse payloads include text; a deep canon-walk could accumulate hundreds of KB. Mitigate: cap per-node text at ~2KB, store `verseRef` for chapter nodes and lazy-fetch on render.
- **Idempotency edge cases.** A reader clicking the same word twice in fast succession from the same parent: must not double-create. Use the (parentId, kind, identity) lookup before adding.
- **Tree panel real estate.** On narrow viewports (phone) the side panel overlays content. Use a slide-out drawer pattern, not a permanent column.
- **Render nodes are expensive.** Each render is an LLM call that costs the reader's BYOK budget. The tree must NOT silently retrigger renders when navigating back to a render node — the existing modernized text in the payload IS the truth. Future "re-render" is a button on the node, not implicit.
- **D-BE-COPYRIGHT.** Verse text payloads embedded in the tree could be exported as part of the markdown. We're already inside the fair-use posture (bcbooks 2013 corpus, stripped of footnotes); the export inherits that, and each export carries the same "↗ canonical apparatus at churchofjesuschrist.org" line per verse so the export is the same posture as the in-app view.

## IX. Out of scope (for this proposal)

- Saved/named study trees (server-side)
- Sharing trees with another reader
- Multiple concurrent trees in the UI
- Graph visualization (svelte-flow / d3) — Phase B in a later proposal if A's indented list doesn't satisfy
- Tree merge / fork-from-shared
- Search within the tree
- Annotations on edges (`why did the reader make this jump?`)

These are good ideas, deferred deliberately so v1 ships in one session block.

## X. Carry-forward / future inspirations

- **Tree as a substrate work-item.** A completed study tree could be ingested by pg-ai-stewards as a `study-write` family work_item, with the LLM authoring a synthesis study from the traversal. Bridge between exploration and articulation.
- **AGE graph traversal of trees.** The tree's parent edges + cross-domain edges are a natural fit for the substrate's AGE graph — eventually merge with the Thummim graph traversal idea from `thummim-restoration-dictionary.md §VI.5`.
- **Becoming integration.** A tree could be the input to a `becoming/` reflection — "what does this study want to change in me?" generated from the chain.

These point at how this feature stays alive over time.
