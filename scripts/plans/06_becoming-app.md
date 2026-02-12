# Becoming App — Architecture Plan

*Created: February 11, 2026*
*Updated: February 11, 2026 — confirmed SQLite, local-first, Copilot SDK roadmap*
*Context: Tools to help apply the "Become" commitments from our truth studies*

---

## The Problem

We've built a rich library of study documents — [truth.md](../../study/truth.md), [truth-atonement.md](../../study/truth-atonement.md), [truth-modern-prophets.md](../../study/truth-modern-prophets.md), and many more. Each one ends with a "Become" section containing specific commitments. But:

1. **No tracking.** There's no way to see all commitments in one place or track progress on them.
2. **Context switching pain.** Reading a study doc in markdown preview and clicking a scripture link navigates *away* from the study, losing your place. There's no side-by-side reading experience.
3. **No spaced repetition.** Memorization tasks (scriptures, quotes) need repeated exposure on a schedule, not a one-time read.
4. **No daily practice integration.** The "Become" items include things like physical exercise, prayer habits, daily study — these need a lightweight daily check-in, not a doc re-read.

President Oaks at BYU (February 10, 2026) reinforced this:

> "There are two methods of gaining needed knowledge. One, the evolving disclosures of man discovered by the scientific method, and two, the truths disclosed by the spiritual method, which begins with faith in God and relies on scriptures, inspired teaching, and personal revelation. There is no ultimate conflict between knowledge gained by these different methods because God, our omnipotent, eternal Father, knows all truth and beckons us to learn by both methods."

The tools we build are the "scientific method" side — organizing, surfacing, scheduling, tracking — so that the spiritual method has room to work.

---

## Two Apps, One Backend

### App 1: **Become** (Daily Practice Tracker)
A lightweight daily-use app for tracking commitments, memorization, habits, and reminders.

### App 2: **Study** (Scripture Study Reader)
A markdown reading app with side-panel scripture/talk loading, designed for deep study without context-switching.

### Backend: **Go API + SQLite**
A single Go backend serving both apps, with an MCP server interface so the AI can help curate content.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Vue 3 Frontend                     │
│                                                       │
│  ┌──────────────────┐    ┌──────────────────────┐    │
│  │   Become App     │    │    Study App          │    │
│  │                  │    │                        │    │
│  │ • Daily checkin  │    │ • Markdown viewer      │    │
│  │ • Habits/todos   │    │ • Side panel for refs  │    │
│  │ • Memorization   │    │ • Tab navigation       │    │
│  │ • Scripture reps │    │ • Linked highlights    │    │
│  │ • Progress view  │    │ • Footnote following   │    │
│  └──────┬───────────┘    └──────────┬─────────────┘  │
│         │                           │                 │
│         └─────────┬─────────────────┘                 │
│                   │ HTTP/JSON                         │
└───────────────────┼───────────────────────────────────┘
                    │
┌───────────────────┼───────────────────────────────────┐
│                   │  Go Backend                       │
│                   ▼                                   │
│  ┌─────────────────────────────────────────────┐     │
│  │              REST API (chi/echo)             │     │
│  │                                              │     │
│  │  /api/tasks      — CRUD for become items     │     │
│  │  /api/memorize   — spaced repetition queue   │     │
│  │  /api/habits     — daily habit tracking      │     │
│  │  /api/docs       — list/read study docs      │     │
│  │  /api/content    — serve scripture/talk md    │     │
│  │  /api/progress   — reading/study progress    │     │
│  └──────────────────┬──────────────────────────┘     │
│                     │                                 │
│  ┌──────────────────┼──────────────────────────┐     │
│  │           MCP Server (stdio)                 │     │
│  │                                              │     │
│  │  become_add_task    — AI adds a task/goal    │     │
│  │  become_list_tasks  — AI reads current tasks │     │
│  │  become_add_memorize — queue scripture to mem │     │
│  │  become_log_progress — record a completion   │     │
│  │  become_suggest_review — what's due today?   │     │
│  └──────────────────┬──────────────────────────┘     │
│                     │                                 │
│  ┌──────────────────▼──────────────────────────┐     │
│  │              SQLite Database                  │     │
│  │                                              │     │
│  │  tasks         — goals, todos, commitments   │     │
│  │  habits        — daily recurring items       │     │
│  │  habit_logs    — daily check-ins             │     │
│  │  memorize      — scriptures to memorize      │     │
│  │  memorize_reps — each review + quality        │     │
│  │  reading_log   — what docs/chapters read     │     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
└───────────────────────────────────────────────────────┘
```

---

## App 1: Become (Daily Practice Tracker)

### Core Features

#### 1.1 Tasks / Commitments
Extracted from "Become" sections of study documents. Each task has:
- **Title** — short description ("Partake of sacrament with broken heart")
- **Source** — link to the study doc and section it came from
- **Type** — `once` | `daily` | `weekly` | `ongoing`
- **Scripture** — optional linked scripture reference
- **Status** — `active` | `completed` | `paused` | `archived`
- **Notes** — personal reflections on progress

#### 1.2 Habits (Daily Recurring)
Lightweight daily check-in items:
- Morning prayer
- Scripture study (minutes tracked)
- Exercise (type + duration)
- Temple attendance (weekly)
- Custom habits

Each day renders a simple grid: check / skip / not-yet. Historical view shows streaks and patterns.

#### 1.3 Memorization (Spaced Repetition)
Scriptures and quotes to memorize, using a simple SM-2-style algorithm:
- **Card front:** Reference (e.g., "D&C 93:29")
- **Card back:** Full verse text (pulled from gospel-library markdown)
- **Review:** Rate recall 1-5 after each attempt
- **Schedule:** Next review date calculated from quality rating
- **Progress:** Track mastery level per card

The AI (via MCP) can suggest scriptures to add based on current study topics.

#### 1.4 Daily View
A single "today" screen showing:
- Habits to check off
- Memorization cards due for review
- Active tasks/commitments with quick-add notes
- A motivating scripture (random from memorization deck or curated)

### UI Sketch

```
┌─────────────────────────────────────────────┐
│  Become                        Feb 11, 2026 │
├─────────────────────────────────────────────┤
│                                             │
│  TODAY'S HABITS                    5/7 ✓    │
│  ┌─────────────────────────────────────┐    │
│  │ ✅ Morning prayer                   │    │
│  │ ✅ Scripture study (25 min)         │    │
│  │ ✅ Exercise — pushups 3x15         │    │
│  │ ☐  Journal entry                   │    │
│  │ ☐  Temple (this week)              │    │
│  │ ✅ Sacrament prep                   │    │
│  │ ✅ Family prayer                    │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  MEMORIZE (3 due today)                     │
│  ┌─────────────────────────────────────┐    │
│  │ D&C 93:29  ▸ Review                │    │
│  │ Mosiah 3:19 ▸ Review               │    │
│  │ D&C 88:67  ▸ Review                │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  ACTIVE COMMITMENTS                         │
│  ┌─────────────────────────────────────┐    │
│  │ See Christ in the mechanics         │    │
│  │   from: truth-atonement.md          │    │
│  │   ✎ journaled about this today     │    │
│  │                                     │    │
│  │ Trust the grace-for-grace process   │    │
│  │   from: truth-atonement.md          │    │
│  │   + read D&C 93:12-13 today        │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  "The glory of God is intelligence"         │
│   — D&C 93:36                               │
└─────────────────────────────────────────────┘
```

---

## App 2: Study (Scripture Study Reader)

### Core Features

#### 2.1 Document Browser
Left sidebar listing all study documents from `./study/` and `./journal/`:
- Grouped by category (topic studies, talks, CFM, journal)
- Shows title, date, word count
- Quick search/filter

#### 2.2 Markdown Viewer (Main Panel)
Renders study documents with full markdown support:
- Proper blockquote styling for scripture quotes
- Rendered links (scripture refs, talk refs)
- Table of contents sidebar for long documents
- "Become" section highlighted/pinned

#### 2.3 Reference Side Panel (Key Feature)
When you click a scripture or talk link in the main document:
- Instead of navigating away, it opens in a **side panel**
- The side panel loads the referenced markdown file
- Multiple references can be opened as **tabs** in the side panel
- Your place in the main document is preserved

```
┌──────────────────────────┬──────────────────────────┐
│   MAIN DOCUMENT          │   REFERENCE PANEL        │
│                          │                          │
│   truth-atonement.md     │  [D&C 88] [Mosiah 3] ← tabs
│                          │                          │
│   > "He descended below  │  D&C 88:6-13            │
│   > all things, in that  │                          │
│   > he comprehended all  │  6 He that ascended up   │
│   > things"              │  on high, as also he     │
│   > — D&C 88:6 ←[click] │  descended below all     │
│                          │  things, in that he      │
│   Christ didn't merely   │  comprehended all things │
│   create a program...    │  ...                     │
│                          │  7 Which truth shineth.  │
│                          │  This is the light of    │
│                          │  Christ.                 │
│                          │                          │
│                          │  [footnotes visible]     │
│                          │  [cross-refs clickable]  │
└──────────────────────────┴──────────────────────────┘
```

#### 2.4 Reading Progress
Track which documents you've read, which scriptures you've visited, and which "Become" items you've engaged with. Feed this back to the Become app.

#### 2.5 Scripture Quick-Add
From the reference panel, one-click to:
- Add a scripture to the memorization deck
- Add a reading note
- Mark as "studied in depth"

---

## Go Backend

### Technology Choices

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Consistent with existing MCP servers (gospel-mcp, webster-mcp, gospel-vec, yt-mcp) |
| HTTP Router | chi or echo | Lightweight, idiomatic Go |
| Database | SQLite | Single file, no server, already used by gospel-mcp |
| Frontend Build | Vite + Vue 3 | Modern, fast, component-based |
| CSS | Tailwind CSS | Utility-first, rapid UI development |
| Markdown | markdown-it (JS) | Client-side rendering with plugin support |
| MCP | stdio JSON-RPC | Same pattern as all existing MCP servers |

### Project Structure

```
scripts/becoming/
├── cmd/
│   ├── server/
│   │   └── main.go           # HTTP server + static file serving
│   └── mcp/
│       └── main.go           # MCP server (stdio mode)
├── internal/
│   ├── db/
│   │   ├── sqlite.go         # DB init, migrations
│   │   ├── tasks.go          # Task CRUD
│   │   ├── habits.go         # Habit CRUD + logging
│   │   ├── memorize.go       # Memorization cards + SM-2
│   │   └── reading.go        # Reading progress
│   ├── api/
│   │   ├── router.go         # Route setup
│   │   ├── tasks.go          # /api/tasks handlers
│   │   ├── habits.go         # /api/habits handlers
│   │   ├── memorize.go       # /api/memorize handlers
│   │   ├── docs.go           # /api/docs handlers (file browser)
│   │   └── content.go        # /api/content handlers (serve markdown)
│   └── mcp/
│       └── server.go         # MCP tool definitions
├── frontend/
│   ├── src/
│   │   ├── App.vue
│   │   ├── router.ts         # Vue Router — /become and /study routes
│   │   ├── views/
│   │   │   ├── BecomeView.vue    # Daily tracker
│   │   │   ├── StudyView.vue     # Document reader
│   │   │   ├── MemorizeView.vue  # Flashcard review
│   │   │   └── ProgressView.vue  # Stats/history
│   │   ├── components/
│   │   │   ├── HabitGrid.vue
│   │   │   ├── TaskCard.vue
│   │   │   ├── FlashCard.vue
│   │   │   ├── DocBrowser.vue
│   │   │   ├── MarkdownViewer.vue
│   │   │   └── ReferencePanel.vue
│   │   └── composables/
│   │       ├── useApi.ts
│   │       ├── useMemorize.ts
│   │       └── useMarkdown.ts
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── package.json
├── go.mod
├── go.sum
├── becoming.db               # SQLite database (gitignored)
└── README.md
```

### Database Schema

```sql
-- Goals and commitments extracted from study docs
CREATE TABLE tasks (
    id          INTEGER PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    source_doc  TEXT,          -- e.g., "study/truth-atonement.md"
    source_section TEXT,       -- e.g., "Become"
    scripture   TEXT,          -- e.g., "D&C 93:29"
    type        TEXT NOT NULL, -- once, daily, weekly, ongoing
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Daily recurring habits
CREATE TABLE habits (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    frequency   TEXT NOT NULL DEFAULT 'daily', -- daily, weekly
    icon        TEXT,
    sort_order  INTEGER DEFAULT 0,
    active      BOOLEAN DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Daily habit completion log
CREATE TABLE habit_logs (
    id          INTEGER PRIMARY KEY,
    habit_id    INTEGER NOT NULL REFERENCES habits(id),
    date        DATE NOT NULL,
    completed   BOOLEAN NOT NULL DEFAULT 0,
    value       TEXT,            -- optional: "25 min", "3x15 pushups"
    notes       TEXT,
    UNIQUE(habit_id, date)
);

-- Scriptures/quotes to memorize
CREATE TABLE memorize (
    id          INTEGER PRIMARY KEY,
    reference   TEXT NOT NULL,   -- "D&C 93:29" or "Mosiah 3:19"
    text        TEXT NOT NULL,   -- full verse/quote text
    source_path TEXT,            -- path to source file
    tags        TEXT,            -- comma-separated
    -- SM-2 fields
    ease_factor REAL DEFAULT 2.5,
    interval    INTEGER DEFAULT 1,  -- days
    repetitions INTEGER DEFAULT 0,
    next_review DATE,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Individual review attempts
CREATE TABLE memorize_reps (
    id          INTEGER PRIMARY KEY,
    card_id     INTEGER NOT NULL REFERENCES memorize(id),
    reviewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    quality     INTEGER NOT NULL,  -- 0-5 (SM-2 scale)
    time_ms     INTEGER            -- how long the review took
);

-- Reading/study progress
CREATE TABLE reading_log (
    id          INTEGER PRIMARY KEY,
    doc_path    TEXT NOT NULL,     -- relative path to document
    doc_type    TEXT,              -- study, scripture, talk, manual
    read_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    notes       TEXT
);
```

### MCP Tools

The MCP server enables the AI to help manage the Become app during study sessions:

| Tool | Description | Use Case |
|------|-------------|----------|
| `become_add_task` | Create a new commitment/goal | After a study session produces "Become" items, AI adds them directly |
| `become_list_tasks` | List active tasks, optionally filtered | AI checks what commitments exist before suggesting new ones |
| `become_add_memorize` | Queue a scripture for memorization | During study, AI suggests key verses and adds them to the deck |
| `become_log_progress` | Record a completion or note on a task | AI helps journal progress |
| `become_suggest_review` | Get today's due memorization cards | AI can incorporate review into study session |
| `become_get_habits` | List habits and today's completion status | AI can ask "have you done your study today?" |

Example interaction during a study session:
```
User: "Let's study D&C 93 today"
AI: [reads D&C 93, surfaces insights]
AI: "D&C 93:29 is a keystone verse for the matter spectrum. Want me to add it to your memorization deck?"
User: "Yes"
AI: [calls become_add_memorize with reference and text]
AI: "Added. You have 4 cards due for review today — want to do those first?"
```

---

## Build Phases

### Phase 1: Foundation (Backend + Become MVP)
**Goal:** Daily habit tracking and task management working end-to-end.

1. Go project setup (`scripts/becoming/`, go.mod, go.work entry)
2. SQLite database init + migrations
3. REST API: tasks CRUD, habits CRUD, habit_logs
4. Vue 3 + Vite + Tailwind scaffold
5. BecomeView: habit grid, task list, daily view
6. Basic MCP server with `become_add_task` and `become_list_tasks`

### Phase 2: Memorization
**Goal:** Spaced repetition system for scriptures.

1. Memorize table + SM-2 algorithm implementation
2. REST API: memorize CRUD, review endpoint, schedule calculation
3. MemorizeView: flashcard UI with flip animation, quality rating
4. MCP: `become_add_memorize`, `become_suggest_review`
5. Seed initial cards from "Become" sections of existing studies

### Phase 3: Study Reader
**Goal:** Side-by-side markdown reader with reference panel.

1. REST API: `/api/docs` (list study files), `/api/content` (serve markdown by path)
2. DocBrowser component (sidebar file tree)
3. MarkdownViewer component (main panel, markdown-it rendering)
4. ReferencePanel component (side panel, tabbed, opens on link click)
5. Link interception: internal scripture/talk links open in side panel instead of navigating
6. Reading progress tracking

### Phase 4: Integration
**Goal:** Both apps talk to each other and the AI.

1. Study reader "Add to memorize" button in reference panel
2. Study reader surfaces "Become" sections with task-creation buttons
3. Progress dashboard showing study patterns, memorization streaks, habit consistency
4. MCP tools for reading log and progress queries
5. Polish, mobile responsiveness, dark mode

### Phase 5: In-App AI Assistant (GitHub Copilot SDK)
**Goal:** Chat with AI directly inside the Study reader.

The [GitHub Copilot SDK](https://github.com/github/copilot-sdk) (Go SDK: `go get github.com/github/copilot-sdk/go`) embeds Copilot's agentic runtime — the same engine behind Copilot CLI — into our Go backend. This gives us:

- An **in-app chat panel** in the Study reader where you can ask questions while reading
- The AI agent has access to all existing MCP servers (gospel-mcp, gospel-vec, webster-mcp) plus the Become MCP tools
- Multi-turn conversations with planning, tool invocation, and streaming responses
- Model selection (GPT-5, Claude, etc.) — same models available through Copilot

**Implementation:**
1. Add `github.com/github/copilot-sdk/go` to go.mod
2. Create `/api/chat` endpoint that proxies to the Copilot SDK agent runtime
3. Register existing MCP servers as tools the agent can invoke
4. Build ChatPanel Vue component (collapsible side panel in Study reader)
5. Stream agent responses to the frontend via SSE or WebSocket

**Requirements:**
- GitHub Copilot subscription (or BYOK with own API keys)
- Copilot CLI installed on the host
- Each prompt counts against premium request quota

**Why wait until Phase 5:**
- Currently in Technical Preview — API may change
- Phases 1-4 don't need AI in the runtime (CRUD + spaced repetition + markdown rendering)
- MCP server (Phase 4) already covers AI integration during VS Code study sessions
- The Copilot SDK adds value specifically in the Study reader, where you're reading and want to *ask* about what you're reading

**Example interaction:**
```
[Reading truth-atonement.md in the Study reader]
[Chat panel open]

You: "What does 'comprehended' mean in the 1828 dictionary? D&C 88:6 says He comprehended all things."
AI: [calls webster_define("comprehend")] "In Webster 1828: 'To include; to contain...
     also: to understand; to conceive.' So 'comprehended all things' carries both
     meanings — He contained all things AND understood all things."

You: "Add D&C 88:6 to my memorization deck."
AI: [calls become_add_memorize] "Added. You have 6 cards due for review tomorrow."
```

### Phase 6: Deployment + Multi-User (Dokploy on VPS)
**Goal:** Deploy to a VPS so others can benefit from the app.

1. Dockerize the app (single container: Go binary + embedded Vue build + SQLite volume)
2. Deploy via Dokploy on VPS with SSL and domain routing
3. Add user authentication (JWT or session-based)
4. Multi-tenant SQLite strategy:
   - Option A: One SQLite file per user (simpler isolation, easy backup/restore per user)
   - Option B: Single DB with user_id foreign keys (simpler queries, standard approach)
   - Leaning: Option A for clean isolation — each user's data is one file
5. User registration + onboarding flow
6. Shared content (scriptures, talks) served from a common gospel-library mount
7. Per-user study docs (or read-only shared + personal notes layer)
8. For Copilot SDK: each user needs their own Copilot auth, or use BYOK with a shared API key and rate limiting

---

## Content Serving Strategy

The Go backend serves markdown content from the existing filesystem:

- **Study docs:** `../../study/**/*.md` (relative to the backend)
- **Scriptures:** `../../gospel-library/eng/scriptures/**/*.md`
- **Conference talks:** `../../gospel-library/eng/general-conference/**/*.md`
- **Manuals:** `../../gospel-library/eng/manual/**/*.md`

The API does NOT copy or duplicate files. It reads directly from the workspace. This means:
- New study docs appear immediately
- Scripture edits (if any) are reflected live
- The gospel-library download can continue expanding without app changes

Path resolution and security: the API validates that requested paths stay within the allowed content roots. No directory traversal outside the workspace.

---

## Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|----------|
| 1 | Database | **SQLite3** | Single user, single writer, already used by gospel-mcp. Zero ops. One file. |
| 2 | Hosting | **Local first** (localhost:8080) | Build and use it before worrying about deployment. |
| 3 | Deployment (later) | **Dokploy on VPS** (Phase 6) | Already familiar with Dokploy. Skip K8s — it solves problems we don't have. |
| 4 | Multi-user (later) | **Phase 6** | Get the app working for one user first. Add auth + multi-tenant in deployment phase. |
| 5 | AI integration | **MCP** (Phase 4) + **Copilot SDK** (Phase 5) | MCP for VS Code sessions. Copilot SDK for in-app AI chat in the Study reader. |
| 6 | Scripture text in cards | **Snapshot at creation** | Store verse text in DB so cards work even if files move. Keep `source_path` for linking back. |
| 7 | Mobile | **PWA** (Phase 4) | Vue 3 + Vite has good PWA support. Add after core features work. |
| 8 | Notifications | **Later** | Start with web UI showing "X cards due today." Push notifications are a Phase 6 concern. |

## Open Questions

1. **Copilot SDK pricing at scale:** If we go multi-user, each prompt counts against premium request quota. BYOK with a shared API key + rate limiting? Or require each user to have their own Copilot subscription?

2. **Shared vs. personal study docs in multi-user:** Read-only shared content (our published studies) + personal notes layer? Or full personal study doc editing?

3. **Offline support:** SQLite + PWA could work offline. Worth investing in service worker + IndexedDB sync?

---

## Why This Matters

President Oaks (February 10, 2026):
> "Strong faith requires more than strong desire. It means daily trying, one step at a time, with prayer and scripture study."

The truth studies mapped the mechanics. The Atonement study showed the mechanism of transformation. The modern prophets study confirmed it works. But *knowing* the framework is not the same as *living inside it.*

The Become app is the tool for living inside it. It turns "I should memorize D&C 93:29" into a card that comes back tomorrow, and next week, and next month, until it's part of you. It turns "See Christ in the mechanics" into a tracked commitment you revisit daily. It turns study documents from things you *read once* into things you *inhabit*.

The Study reader removes the friction between understanding and absorption. When you can read a study doc and simultaneously explore every scripture it references — with footnotes, cross-references, and the full chapter context — the study gets deeper, faster, without the mental overhead of navigation.

Together, these tools serve the doctrine: "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection" ([D&C 130:18](../../gospel-library/eng/scriptures/dc-testament/dc/130.md)). The goal is not to build an app. The goal is to *become*.
