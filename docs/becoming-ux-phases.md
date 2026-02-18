# Becoming App — UX Phases

A living document for planning UX improvements to the Becoming app. Features are organized into phases by priority and complexity.

---

## Phase 1: Reader Polish (Current)

These are in-progress or just shipped:

- [x] **Dark mode whole-window** — Toggle dark mode on `<html>` so nav bar, page background, and reader all go dark together
- [x] **Shared reader sidebar expand/scroll** — Deep-link URLs (`?f=path/to/file.md`) now expand the sidebar tree and scroll to the active file in the public reader
- [x] **Header anchor links** — GitHub-style link icons on markdown headings. Hover to reveal, click to get a shareable `#hash` URL
- [x] **Emoji picker for pillars** — Categorized dropdown (spiritual, intellectual, social, physical, creative, symbols) replaces free-text input
- [x] **Tri-state practice filters** — Click once for positive (include), twice for negative (exclude with red strikethrough), third to clear
- [x] **Pillar emoji icons in practice rows** — Purple pills showing each practice's pillar emoji

---

## Phase 2: Bookmarks & Highlights

Deep-link bookmarks let users save specific passages with context.

### Bookmark Model
- **What gets saved:** file path + heading anchor (or text excerpt hash) + optional user note
- **Storage:** New `bookmarks` table: `id`, `user_id`, `source_id`, `file_path`, `anchor` (heading slug or selection hash), `excerpt` (text snippet for display), `note` (optional), `created_at`
- **Deep link format:** `/reader/{sourceId}?f=path/to/file.md#heading-slug`

### UX Design
- **Creating a bookmark:** Click the heading anchor link icon → option to "Bookmark this section." Or select text → existing floating toolbar gets a "Bookmark" button alongside "Memorize" and "Create Practice"
- **Bookmark management:** New "Bookmarks" page accessible from nav bar. Shows bookmarks grouped by source/file with excerpts and notes. Click to jump directly to the bookmarked location
- **Visual indicator:** Small bookmark icon in the sidebar tree next to files that have bookmarks. Heading anchors that are bookmarked could show a filled bookmark icon instead of the link icon

### API Endpoints
```
POST   /api/bookmarks          { source_id, file_path, anchor, excerpt, note }
GET    /api/bookmarks          ?source_id=...  (list, filterable)
DELETE /api/bookmarks/:id
PATCH  /api/bookmarks/:id      { note }
```

### Open Questions
- Should bookmarks be shareable (public short-links to specific headings)?
- Should highlights (selected text) be persisted separately from heading bookmarks?
- Tags/folders for organizing bookmarks?

---

## Phase 3: Reading Progress & History

- **Recently read:** Track which files were opened and when. Show a "Recent" section on Today page
- **Reading progress:** Optional per-source progress indicator (which files have been read)
- **Continue where you left off:** Re-open the last file when returning to a source

---

## Phase 4: Collaborative Features

- **Shared annotations:** When sharing a study, include bookmark annotations for others to see
- **Study groups:** Multiple users can annotate the same source, see each other's highlights
- **Discussion threads:** Comment on specific headings or passages

---

## Phase 5: Search & Discovery

- **Full-text search within reader:** Search across all files in the current source
- **Cross-source search:** Find a passage across all your sources
- **Related content suggestions:** When reading a section, surface related bookmarks, practices, or journal entries

---

## Design Principles

1. **Reading first.** Features should enhance reading, not interrupt it. Anchor links appear on hover. Bookmarks are one click away but never in the way.
2. **Deep links are the currency.** Every piece of content should have a shareable URL. Headings, selections, bookmarks — all linkable.
3. **Progressive disclosure.** Simple by default, powerful on demand. The reader looks clean until you interact with it.
4. **Offline-friendly data model.** Bookmarks and highlights are small, cacheable, and syncable.
