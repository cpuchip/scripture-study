---
date: 2026-05-24
session_window: afternoon
workstream: WS8
shipped:
  - Reactive count for E-Tier words (useWordData.ts reactivity fix)
  - Direct routing toggles on WordCard titles and WordStudy page headers
  - In-place definition aside panel layout for WordStudy view
  - Shift+Click / modifier-key shortcut toggles in HighlightedText.vue
  - Project spec/proposal: ux-and-presentation-enhancements.md
status: spec proposed & quick-fixes shipped
---

# 2026-05-24 — UX Improvements & Spec Proposal

We shifted gears from cpuchip.net to explore 1828 and address navigation friction points in scripture/occurrences mode, search E-tier counts, and draft a roadmap for study and presentation modes.

## What landed

### E-Tier Count Reactivity Fix
- **Problem**: The word search E-tier count was stuck at `0` even after the headwords loaded async.
- **Cause**: `tierCounts` was a plain, non-reactive JavaScript object. Updating `tierCounts.E` did not trigger a Vue re-render.
- **Fix**: Wrapped `tierCounts` inside Vue's `reactive()` in [useWordData.ts](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/1828-illuminated/frontend/src/composables/useWordData.ts). Now, the E-tier filter updates to `97,969` (the full Webster headwords count minus curated lists) immediately when the list finishes loading.

### Direct Toggles on Word/Scripture Views
- **Problem**: Already viewing a defined word in `WordDetail.vue` (definition view), the title link pointed to the same page, doing nothing. In `WordStudy.vue` (scripture occurrences mode), there was no clear way to click the word to see its definition.
- **Fix**:
  - [WordCard.vue](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/1828-illuminated/frontend/src/components/WordCard.vue) title now checks the active route. If on `/word/:word`, it points to `/word-study/:word` (scripture mode) with an informative tooltip; otherwise, it points to `/word/:word` (definition mode).
  - [WordStudy.vue](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/1828-illuminated/frontend/src/views/WordStudy.vue) header title converted to a clickable router-link heading back to its 1828 definition `/word/:word`.

### Scripture Mode Definition Sidebar & Shift+Click Shortcut
- **Problem**: In scripture occurrences mode, clicking any word in the verses would route the reader to its occurrences page. There was no way to simply preview the word's 1828 definition in-place without toggling the global click mode or navigating away.
- **Fix**:
  - Widened `WordStudy.vue` to a `max-w-6xl` grid layout with a right-hand sticky sidebar rendering the `WordCard` component for `selectedWord`, identical to `VerseExplorer.vue`.
  - Added a modifier key listener (`event.shiftKey || event.ctrlKey || event.altKey || event.metaKey`) to the click handler in [HighlightedText.vue](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/1828-illuminated/frontend/src/components/HighlightedText.vue).
    - If in **Scripture Mode**: standard click finds occurrences; `Shift+Click` opens the definition preview in the right sidebar.
    - If in **Definition Mode**: standard click opens the definition preview; `Shift+Click` routes to scripture occurrences.
    - Hover tooltips updated to display these shortcut options dynamically.

### Spec Proposal Written
- Created [ux-and-presentation-enhancements.md](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/1828-illuminated/.spec/proposals/ux-and-presentation-enhancements.md) outlining:
  - Arbitrary canon books and custom pasted texts in presentation mode.
  - Presenting scripture chains directly from the Study Tree deck.
  - Interactive breadcrumbs inline at the top of the main viewport to support time-travel branch navigation and sibling forks.
  - Account integration (Google OAuth) for cross-device sync and named cloud-saved trees.
  - gospel-engine-v2 local MCP server integration for citation mapping and semantic deep search.

## Verification
- Clean build: `vue-tsc -b && vite build` completed successfully.
- Code committed to `main` and pushed to origin.

## Carry-forward
- Implementation of Phase 1 (dynamic query-driven Present mode) and Phase 2 (Inline Breadcrumbs component).
- Stewardship Council review of the WS8 proposal decisions.
