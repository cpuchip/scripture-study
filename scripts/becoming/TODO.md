# Becoming App — TODO

*Consolidated from `docs/06_becoming-app.md`, `docs/becoming-improvements.md`, and `docs/becoming-ux-phases.md`.*
*Last updated: 2026-02-20*

---

## What's Done

### Build Phases

| Phase | Status | Description |
|-------|--------|-------------|
| 1 — Foundation | **Done** | Go backend, SQLite, REST API, Vue 3 + Vite + Tailwind, DailyView, PracticesView, HistoryView, TasksView |
| 2 — Memorization | **Done** | SM-2 spaced repetition, flashcard UI, quality rating, daily due cards |
| 3 — Study Reader | **Done** | GitHub-backed document sources, markdown viewer, sidebar file tree, reference panel with tabs, reading progress |
| 6 — Deployment | **Done** | Dokploy on VPS at ibeco.me, JWT auth, PostgreSQL for prod, SQLite for dev, user registration |

### Enhancement Sprints

| Sprint | Status | Description |
|--------|--------|-------------|
| 1 — Practice Lifecycle Backend | **Done** | Status field, archived_at, end_date, migrations |
| 2 — Practice Lifecycle Frontend | **Done** | Tabs (active/paused/completed/archived), action icons, end date badges |
| 3 — Study Mode | **Done** | Adaptive difficulty, 8 exercise modes, aptitude model, session momentum |
| 4 — Memorize Card Lifecycle | **Done** | Mastery detection, complete/archive cards, aptitude dashboard |
| 5 — Activity Heatmap | **Done** | GitHub-style contribution heatmap on Reports page |
| 6 — Start Date & Future Planning | **Done** | Editable start_date, future practice filtering |

### UX Phase 1: Reader Polish

| Feature | Status |
|---------|--------|
| Dark mode whole-window (reader views) | **Done** |
| Shared reader sidebar expand/scroll | **Done** |
| Header anchor links (GitHub-style) | **Done** |
| Anchor click copies URL to clipboard | **Done** |
| Emoji picker for pillars | **Done** |
| Tri-state practice filters | **Done** |
| Pillar emoji icons in practice rows | **Done** |

### Phase 4 Polish (Recent)

| Feature | Status |
|---------|--------|
| Global dark mode (all views, useTheme composable) | **Done** |
| Mobile hamburger nav | **Done** |
| ActivityHeatmap dark-mode-aware colorScale | **Done** |

---

## What's Next

### UX Phase 2: Bookmarks & Highlights <-- CURRENT

Save specific passages in the reader with context and notes.

**Bookmark model:**
- `bookmarks` table: `id`, `user_id`, `source_id`, `file_path`, `anchor` (heading slug), `excerpt` (text snippet), `note` (optional), `created_at`
- Deep link format: `/reader/{sourceId}?f=path/to/file.md#heading-slug`

**Done:**
- [x] DB layer (`internal/db/bookmarks.go`) — Bookmark struct + CRUD
- [x] SQLite migration (`migrateBookmarks` in db.go)
- [x] PostgreSQL migration (`007_bookmarks.sql`)
- [x] API endpoints (GET/POST/PATCH/DELETE `/api/bookmarks`)
- [x] Frontend API client (Bookmark interface + api methods in api.ts)
- [x] BookmarksView page (grouped by source, search, inline note editing)
- [x] Reader integration (bookmark button on heading anchors, toggle on/off, visual state)
- [x] Navigation (desktop + mobile nav links, route + page title)

**Remaining:**
- [ ] Visual indicator in sidebar tree for files with bookmarks
- [ ] Shareable bookmarks via public short-links?
- [ ] Text highlights (selected text) as separate bookmark type?
- [ ] Tags/folders for organizing bookmarks?

---

### UX Phase 3: Reading Progress & History

- **Recently read:** Track file opens + timestamps, show "Recent" section on Today page
- **Reading progress:** Per-source progress indicator (which files read)
- **Continue where you left off:** Reopen last file on returning to a source

---

### Phase 4 Remaining

| Item | Status | Notes |
|------|--------|-------|
| PWA support | Not started | Service worker, installable, offline-capable |
| Memorize-from-reader quick-add | Not started | One-click "Add to memorize" from reader reference panel |
| Collapsible mobile practice filters | Not started | Funnel icon to expand/collapse filter rows on mobile |

---

### UX Phase 4: Collaborative Features

- **Shared annotations:** Include bookmark annotations when sharing a study
- **Study groups:** Multiple users annotate same source, see each other's highlights
- **Discussion threads:** Comment on specific headings or passages

---

### UX Phase 5: Search & Discovery

- **Full-text search within reader:** Search across all files in current source
- **Cross-source search:** Find passages across all sources
- **Related content suggestions:** Surface related bookmarks, practices, or journal entries while reading

---

### Phase 5: In-App AI Assistant (Copilot SDK)

Chat with AI directly inside the Study reader. Uses GitHub Copilot SDK to embed agentic runtime in the Go backend. The AI has access to all MCP tools (gospel-mcp, gospel-vec, webster-mcp, becoming-mcp).

**Waiting on:** Copilot SDK leaving technical preview.

**Implementation:**
- `/api/chat` endpoint proxying to Copilot SDK agent runtime
- Register existing MCP servers as tools
- ChatPanel Vue component (collapsible side panel in reader)
- Streaming responses via SSE or WebSocket

---

## Design Principles

1. **Reading first.** Features enhance reading, never interrupt it.
2. **Deep links are the currency.** Every piece of content gets a shareable URL.
3. **Progressive disclosure.** Simple by default, powerful on demand.
4. **Search results are pointers, not sources.** Tools help you *find* what to read, then get out of the way.
5. **Don't build shortcuts past reading.** The goal is deeper study, not faster skimming.

---

## Database Tables (Current)

| Table | Purpose |
|-------|---------|
| `practices` | Generalized trackable items (memorize, exercise, habit, task) |
| `practice_logs` | Per-practice daily completion logs |
| `tasks` | Goals, commitments, one-time items |
| `notes` | User notes |
| `prompts` | Daily reflection prompts |
| `reflections` | Daily reflections |
| `pillars` | Growth pillars (spiritual, intellectual, etc.) |
| `practice_pillars` | Many-to-many: practices ↔ pillars |
| `task_pillars` | Many-to-many: tasks ↔ pillars |
| `memorize_scores` | Per-exercise score history for adaptive difficulty |
| `memorize_aptitude` | Per-card per-mode aptitude cache |
| `document_sources` | GitHub repo configs for the reader |
| `reading_progress` | Track which files opened, scroll position |

PostgreSQL migrations: `internal/db/migrations/postgres/001-006`
SQLite migrations: `runSQLiteMigrations()` in `internal/db/db.go`
