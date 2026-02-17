# Becoming App — Architecture Plan

*Created: February 11, 2026*
*Updated: February 16, 2026 — Phases 1, 2, 6 complete. Enhancement Sprints 1-6 complete. Phase 3 replanned with git integration.*
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
│  │              REST API (chi v5)               │     │
│  │                                              │     │
│  │  /api/practices — CRUD for trackable items   │     │
│  │  /api/logs      — log practice completions   │     │
│  │  /api/daily/:d  — daily summary view         │     │
│  │  /api/tasks     — goals and commitments      │     │
│  │  /api/docs      — list/read study docs (P3)  │     │
│  │  /api/content   — serve scripture/talk (P3)  │     │
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
│  │  practices     — generalized trackable items │     │
│  │  practice_logs — per-practice daily logs     │     │
│  │  tasks         — goals, todos, commitments   │     │
│  │  reading_log   — what docs/chapters read (P3)│     │
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

### Project Structure (Actual — Phase 1)

```
scripts/becoming/
├── cmd/
│   ├── server/
│   │   ├── main.go           # HTTP server + embedded frontend (go:embed)
│   │   └── dist/             # Built frontend (gitignored, copied from frontend/dist)
│   └── mcp/
│       └── main.go           # MCP server (Phase 4)
├── internal/
│   ├── db/
│   │   ├── db.go             # DB init (Open, initSchema with embedded SQL)
│   │   ├── schema.sql        # SQLite schema (embedded via go:embed)
│   │   ├── practices.go      # Practice CRUD (Create, Get, List, Update, Delete)
│   │   ├── logs.go           # PracticeLog CRUD + DailySummary join query
│   │   └── tasks.go          # Task CRUD
│   └── api/
│       └── router.go         # All REST routes (chi router)
├── frontend/
│   ├── src/
│   │   ├── App.vue           # Nav bar + router-view shell
│   │   ├── router.ts         # Routes: /, /practices, /practices/:id/history, /tasks
│   │   ├── api.ts            # Typed API client (all endpoints)
│   │   └── views/
│   │       ├── DailyView.vue      # Today screen: grouped practices, quick-log, date nav
│   │       ├── PracticesView.vue  # Create/edit practices with type-specific config
│   │       ├── HistoryView.vue    # 30-day bar chart, streak, stats, recent activity
│   │       └── TasksView.vue      # Task CRUD with status toggle
│   ├── index.html
│   ├── vite.config.ts        # Tailwind plugin + API proxy to localhost:8080
│   ├── tsconfig.json
│   └── package.json
├── go.mod                    # chi v5.2.1, cors v1.2.1, go-sqlite3 v1.14.24
├── go.sum
├── .gitignore                # becoming.db, server.exe, cmd/server/dist/
└── README.md                 # (Phase 1 — to be written)
```

### Database Schema (Actual — Generalized Model)

Instead of separate tables for habits, memorize, and exercises, we unified everything into a **generalized practice model**. Each practice has a `type` field and a JSON `config` blob for type-specific settings.

```sql
-- Practices: anything you do repeatedly and want to track
-- Types: memorize, exercise, habit, task
CREATE TABLE practices (
    id           INTEGER PRIMARY KEY,
    name         TEXT NOT NULL,           -- "D&C 93:29" or "Clamshell" or "Morning prayer"
    description  TEXT,                    -- Full verse text, exercise instructions, etc.
    type         TEXT NOT NULL,           -- memorize | exercise | habit | task
    category     TEXT,                    -- "scripture", "pt", "spiritual", "fitness"
    source_doc   TEXT,                    -- link to study doc that generated this
    source_path  TEXT,                    -- path to source file (scripture, talk, etc.)
    config       TEXT DEFAULT '{}',       -- JSON: type-specific settings (see below)
    sort_order   INTEGER DEFAULT 0,
    active       BOOLEAN DEFAULT 1,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Config JSON by type:
-- memorize: {"ease_factor": 2.5, "interval": 1, "repetitions": 0}
-- exercise: {"target_sets": 2, "target_reps": 15, "unit": "reps"}
-- habit:    {"frequency": "daily"}
-- task:     {"due_date": "2026-03-01"}

-- Practice logs: each time you do a practice
CREATE TABLE practice_logs (
    id          INTEGER PRIMARY KEY,
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    logged_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    date        DATE NOT NULL,
    quality     INTEGER,        -- SM-2 quality rating (0-5) for memorize
    value       TEXT,           -- freeform: "25 min", "3 miles"
    sets        INTEGER,        -- number of sets for exercise
    reps        INTEGER,        -- reps per set
    duration_s  INTEGER,        -- duration in seconds
    notes       TEXT,
    next_review DATE            -- spaced repetition: next review date
);

-- Tasks/commitments (separate from practices — one-time or ongoing goals)
CREATE TABLE tasks (
    id             INTEGER PRIMARY KEY,
    title          TEXT NOT NULL,
    description    TEXT,
    source_doc     TEXT,
    source_section TEXT,
    scripture      TEXT,
    type           TEXT NOT NULL DEFAULT 'ongoing',
    status         TEXT NOT NULL DEFAULT 'active',
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at   DATETIME
);
```

**Why this model?** PT exercises, scripture memorization, daily habits, and tracked tasks all share the same core pattern: something you do, tracked over time. The `type` + `config` approach means one set of API endpoints, one daily summary query, and one chart component handles all of them. No code duplication across habit/memorize/exercise modules.

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

### Phase 1: Foundation (Backend + Become MVP) ✅ COMPLETE
**Goal:** Generalized daily practice tracking and task management working end-to-end.

**What was built:**
1. ✅ Go project setup (`scripts/becoming/`, go.mod, go.work entry with 10 modules)
2. ✅ SQLite database with embedded schema.sql (WAL mode, foreign keys)
3. ✅ REST API (chi v5): practices CRUD, logs CRUD, daily summary, tasks CRUD
4. ✅ Vue 3 + Vite + Tailwind scaffold with vue-router
5. ✅ DailyView: grouped practices by category, quick-log, exercise set tracking, date navigation
6. ✅ PracticesView: create/edit with type selector, category presets, exercise config (sets/reps/unit)
7. ✅ HistoryView: 30-day bar chart, streak counter, completion stats, recent activity
8. ✅ TasksView: task CRUD with status toggle, source doc/scripture fields
9. ✅ Go server embeds Vue frontend (go:embed), single binary deployment
10. ✅ Dev mode: `--dev` flag enables CORS for Vite dev server proxy

**Build & run:**
```bash
cd scripts/becoming/frontend && npm run build
cd .. && cp -r frontend/dist cmd/server/dist
go build -o server ./cmd/server/
./server -db becoming.db     # production (embedded frontend)
./server -db becoming.db -dev  # dev mode (CORS for Vite)
```

### Phase 2: Memorization ✅ COMPLETE
**Goal:** Spaced repetition system for scriptures.

**What was built:**
1. ✅ SM-2 algorithm (Wozniak 1987) in `internal/db/memorize.go` — quality 0-5, ease factor, interval, repetitions
2. ✅ REST API: `GET /memorize/due/{date}` (due cards), `POST /memorize/review` (review + auto-schedule)
3. ✅ Auto-populate SM-2 defaults when creating memorize-type practices
4. ✅ MemorizeView: tap-to-flip flashcard, 6-button quality rating (0=Blackout to 5=Easy), card stats
5. ✅ DailyView integration: shows "N cards due — Review →" banner when cards are due
6. ✅ "Memorize" nav link in App.vue
7. ✅ PracticesView: memorize-specific form hints (reference as name, verse text as description)

**SM-2 behavior:**
- New card: due immediately (next_review = today, interval = 0)
- Quality ≥ 3: advance (rep 1 → 1d, rep 2 → 6d, then interval × ease_factor)
- Quality < 3: reset to rep 0, interval 1d
- Ease factor adjusts each review (min 1.3)
- Due query: `json_extract(config, '$.next_review') <= date OR repetitions = 0`

### Phase 3: Study Reader (with Git-Based Document Sources)
**Goal:** Side-by-side markdown reader with reference panel, powered by git repos as document sources.

**The problem with local-only documents:** The current scripture-study repo bundles study documents alongside the Become app, MCP servers, gospel-library content, and tooling. This couples personal study content to infrastructure. For multi-user, each person needs their own study documents — but the app shouldn't care *where* they live.

**The insight:** Git repos are the natural unit for study collections. They version, collaborate, sync, and organize markdown documents — exactly what we already do. Instead of hardcoding paths, the Study reader treats git repos as pluggable document libraries.

#### Architecture: Document Sources

Each user configures one or more **document sources** — git repos that contain study materials:

```
┌─────────────────────────────────────────────────────────────┐
│                     Study Reader                             │
│                                                              │
│  ┌──────────────┐  ┌─────────────────────────────────────┐  │
│  │ Source Panel  │  │ Document Viewer                     │  │
│  │              │  │                                     │  │
│  │ 📚 My Studies │  │  [Main Panel]    [Reference Panel]  │  │
│  │  ├─ creation │  │                                     │  │
│  │  ├─ truth    │  │  study doc ←→ scripture side panel  │  │
│  │  └─ cfm/     │  │                                     │  │
│  │              │  │                                     │  │
│  │ 📖 Gospel Lib │  │                                     │  │
│  │  ├─ bofm/    │  │                                     │  │
│  │  ├─ dc/      │  │                                     │  │
│  │  └─ nt/      │  │                                     │  │
│  │              │  │                                     │  │
│  │ 📝 Shared Repo│  │                                     │  │
│  │  └─ lessons/ │  │                                     │  │
│  └──────────────┘  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

#### Document Source Types

| Source Type | Description | Multi-user? |
|-------------|-------------|-------------|
| **Local filesystem** | Direct path to a folder of markdown files (e.g., `../../study/`). Works for single-user/dev. | No — server-only |
| **Git repo (clone)** | Shallow clone a GitHub/GitLab repo into user's content directory. Periodic pull to sync. | Yes — each user links their own repos |
| **Git repo (API)** | Browse via GitHub API without cloning. Read markdown on-demand from raw content URLs. | Yes — minimal storage |
| **Gospel Library (built-in)** | The shared `gospel-library/` content — scriptures, talks, manuals. Read-only, available to all users. | Yes — common mount |

#### Git Integration Design

**Per-user repo configuration:**
```sql
CREATE TABLE document_sources (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    name        TEXT NOT NULL,           -- "My Studies", "Ward Lessons"
    source_type TEXT NOT NULL,           -- 'local' | 'git_clone' | 'git_api' | 'gospel_library'
    url         TEXT,                    -- git URL or local path
    branch      TEXT DEFAULT 'main',
    sub_path    TEXT DEFAULT '',         -- subfolder within repo (e.g., 'study/')
    auth_token  TEXT,                    -- encrypted GitHub PAT for private repos
    last_synced TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**Sync strategy for `git_clone` sources:**
- On first add: `git clone --depth 1 --filter=blob:none` (minimal footprint)
- On app load or manual refresh: `git pull --ff-only`
- Storage: `data/repos/{user_id}/{source_id}/` on server
- Only `.md` files are indexed/served — ignore everything else

**GitHub API approach (`git_api` sources):**
- Use GitHub Contents API to list directories and read files on-demand
- Cache responses with ETags for efficient re-fetching
- No storage needed, but slower for browsing deep trees
- Rate limit: 5,000 req/hour authenticated, 60/hour unauthenticated

#### Step 7: Repository Separation

The current `scripture-study` repo should be split:
- **`scripture-study`** — The app: Become backend, frontend, MCP servers, gospel-library content
- **`studies`** (new repo) — Personal study documents, lessons, journal entries, CFM notes

The `studies` repo becomes the first git document source for the Study reader. This cleanly separates infrastructure from content, and lets other users create their own study repos that plug into the same app.

#### Core Study Reader Features

1. **REST API:**
   - `GET /api/sources` — list user's document sources
   - `POST /api/sources` — add a new source (git URL, local path, etc.)
   - `POST /api/sources/{id}/sync` — trigger git pull
   - `GET /api/docs?source={id}` — list documents in a source
   - `GET /api/content?source={id}&path={path}` — serve markdown file content
2. **DocBrowser component** — sidebar file tree, grouped by source
3. **MarkdownViewer component** — main panel, markdown-it rendering with proper blockquote styling
4. **ReferencePanel component** — side panel with tabs, opens scripture/talk links without navigating away
5. **Link interception** — internal `gospel-library/` links open in reference panel
6. **Reading progress tracking** — which docs/chapters have been read
7. **"Add to memorize" button** — one-click from reference panel to memorization deck

### Phase 4: Integration & Polish
**Goal:** Connect all the pieces, polish the experience.

**Already done (via Enhancement Sprints 1-6):**
1. ✅ Practice lifecycle (pause, complete, archive, restore)
2. ✅ Adaptive study mode with 8 exercise types and SM-2 quality gating
3. ✅ Memorize card lifecycle (mastery detection, end dates, aptitude dashboard)
4. ✅ Activity calendar heatmap on Reports page
5. ✅ Start date & future planning with filtering
6. ✅ MCP tools (becoming-mcp server with practice tracking, journal, memorization)

**Remaining:**
1. Study reader ↔ Become integration ("Add to memorize" from reader)
2. Study reader surfaces "Become" sections with task-creation buttons
3. Mobile responsiveness & UX polish (see below)
4. Dark mode
5. PWA support (service worker, installable)

#### Mobile & UX Polish

The app works but wasn't designed mobile-first. Key improvements needed:

- **Responsive nav:** Hamburger menu / slide-out drawer on small screens (top bar currently overflows)
- **Collapsible filters:** Filter icon (funnel) that expands filter rows on Practices page — the 3 filter rows (Type, Cat, Time) are unwieldy on mobile
- **Touch-friendly targets:** Larger tap areas on card pills, buttons, and practice rows
- **Study mode on narrow screens:** Exercise layouts (especially Arrange Words) need mobile-optimized rendering
- **Heatmap responsiveness:** Scrollable or condensed heatmap on Reports page for small screens
- **Bottom nav option:** Consider bottom tab bar on mobile instead of top nav

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

### Phase 6: Deployment + Multi-User (Dokploy on VPS) ✅ COMPLETE
**Goal:** Deploy to a VPS so others can benefit from the app.

**What was built:**
1. ✅ Dockerized app deployed via Dokploy on VPS at ibeco.me
2. ✅ SSL and domain routing configured
3. ✅ JWT-based authentication with login/register/logout
4. ✅ PostgreSQL for production (not SQLite — chose Option B: single DB with `user_id` foreign keys)
5. ✅ SQLite retained for local dev with automatic schema migration compatibility
6. ✅ DB portability layer: `DateCast()`, `DateText()`, `rebind()`, `InsertReturningID()`, `JSONExtract()`
7. ✅ Goose migrations for PostgreSQL (`internal/db/migrations/postgres/001-004`)
8. ✅ User registration with privacy/terms pages, public landing page
9. ✅ Dynamic branding and logo
10. ✅ Gospel library content served from common mount

**Decision change:** Went with Option B (single PostgreSQL DB with `user_id` FK) instead of Option A (per-user SQLite). Standard approach, simpler queries, better for hosted deployment. SQLite kept for local dev.

---

## Completed Enhancement Sprints

These sprints enhanced Phases 1-2 after deployment. Full specs in [becoming-improvements.md](becoming-improvements.md).

| Sprint | What | Status |
|--------|------|--------|
| 1 | Practice Lifecycle — Schema & Backend (status, archived_at, end_date, migrations) | ✅ |
| 2 | Practice Lifecycle — Frontend (tabs, action icons, end date badges) | ✅ |
| 3 | Memorization Study Mode — Adaptive Difficulty (8 exercise modes, aptitude model, SM-2 quality gating, session momentum) | ✅ |
| 4 | Memorize Card Lifecycle (mastery detection, complete/archive cards, aptitude dashboard, level progression) | ✅ |
| 5 | Activity Calendar Heatmap (GitHub-style heatmap on Reports page) | ✅ |
| 6 | Start Date & Future Planning (start_date column, future card filtering, time-based practice filters) | ✅ |

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
| 9 | Data model | **Generalized practice model** | Single `practices` table with `type` + JSON `config` instead of separate habit/memorize/exercise tables. One API, one daily summary, one chart component. |
| 10 | HTTP router | **chi v5** | Lightweight, idiomatic Go, popular, well-maintained. |
| 11 | Frontend embed | **go:embed** | Single binary deployment. No separate file server needed. |

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
