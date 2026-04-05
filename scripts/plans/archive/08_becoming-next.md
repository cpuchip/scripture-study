# Becoming App — Phase 2.7: Pillars, Notes, Reflections & Trends

*Created: February 12, 2026*
*Updated: February 12, 2026 — Renamed "Becomes" → "Pillars", added sub-pillars, onboarding, clarified notes vs reflections*
*Context: Features to deepen the "becoming" loop before Phase 3 (Study Reader)*
*Domains: ibeco.me (personal/app) + webeco.me (future social/group)*

---

## Overview

Phases 1-2.5 built the **doing** engine — track practices, memorize scriptures, schedule recurring tasks, view reports. But doing without reflection is just motion. These features add the **meaning** layer:

1. **Pillars** — The "why" behind your practices (structured vision / growth areas)
2. **Notes** — Quick outward-facing information capture, attached to things
3. **Reflections** — Fast inward-facing daily journal (1-2 minutes)
4. **Trend Lines** — Visual progress over time in reports

### Pillars vs. Categories

These are distinct and complementary:

- **Pillars** are *structured* — dimensions of Christlike growth. They answer "why am I doing this?" and provide a framework for balanced living. Default: Spiritual, Social, Intellectual, Physical (from Luke 2:52 and the Children & Youth Personal Development program).
- **Categories** are *unstructured* — freeform labels for grouping practices by context. They answer "what do I do together?" Examples: "morning routine", "pt session", "scripture study". Categories stay as-is (comma-separated on the practice).

A practice has **pillar(s)** for meaning and **category** for organization. Plank might be under pillar "Physical > PT Recovery" and category "pt".

---

## Feature 1: Pillars (Vision Layer)

### The Problem

The app tracks *what* you do but has no place for *who you're becoming*. Practices exist in isolation — "Plank" and "D&C 93:29" and "Morning Prayer" are all just rows. There's no way to say: "These practices are all part of growing spiritually" or "These exercises strengthen me physically for my family."

### The Concept

A **Pillar** is a dimension of growth — a structural support for who you're becoming. The scriptural foundation is Luke 2:52:

> "And Jesus increased in wisdom and stature, and in favour with God and man."

- **Wisdom** → Intellectual
- **Stature** → Physical
- **Favour with God** → Spiritual
- **Favour with man** → Social

This is the same framework used by the Church's [Children and Youth Personal Development](../gospel-library/eng/manual/personal-development-youth-guidebook/get-started.md) program: *"Consider creating goals in each of the four areas to keep your life balanced."* The program's cycle of **Discover → Plan → Act → Reflect** maps beautifully to our app's flow.

Pillars support **one level of sub-pillars** for finer granularity. For example:
- **Physical** → PT Recovery, Cardio, Nutrition
- **Spiritual** → Scripture Study, Prayer, Temple
- **Intellectual** → Career Skills, Music, Languages
- **Social** → Family, Friendships, Service

Practices link to either a top-level pillar or a sub-pillar. Linking to a sub-pillar implicitly links to its parent — reports and grouping honor the hierarchy.

### Data Model

```sql
CREATE TABLE IF NOT EXISTS pillars (
    id          INTEGER PRIMARY KEY,
    parent_id   INTEGER REFERENCES pillars(id) ON DELETE CASCADE,  -- NULL = top-level pillar
    name        TEXT NOT NULL,           -- "Spiritual" or "PT Recovery"
    description TEXT,                    -- Longer narrative / vision statement
    scripture   TEXT,                    -- Anchoring scripture reference
    icon        TEXT,                    -- Emoji or short icon label
    sort_order  INTEGER DEFAULT 0,
    active      BOOLEAN DEFAULT 1,
    is_default  BOOLEAN DEFAULT 0,      -- TRUE for the 4 seed pillars
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Junction table: many-to-many between practices and pillars
CREATE TABLE IF NOT EXISTS practice_pillars (
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    pillar_id   INTEGER NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    PRIMARY KEY (practice_id, pillar_id)
);

-- Junction table: many-to-many between tasks and pillars
CREATE TABLE IF NOT EXISTS task_pillars (
    task_id     INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    pillar_id   INTEGER NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, pillar_id)
);
```

### Seed Data (Onboarding)

On first launch, the app shows a setup screen: *"The Savior grew in four areas (Luke 2:52). Here are four suggested pillars to organize your growth. You can rename, remove, or add more."*

| Pillar | Icon | Scripture | Description |
|--------|------|-----------|-------------|
| Spiritual | 🕊️ | Luke 2:52 — "in favour with God" | Prayer, scripture study, temple, sacrament |
| Social | 🤝 | Luke 2:52 — "in favour with man" | Service, relationships, community |
| Intellectual | 📖 | D&C 130:18 — "principle of intelligence" | Learning, skills, career, creativity |
| Physical | 💪 | 1 Cor 6:19 — "your body is the temple" | Exercise, nutrition, health, rest |

The onboarding screen lets the user confirm, rename, or skip each pillar. They proceed with whatever they accept. Pillars can always be edited later.

### API

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/pillars` | List all pillars with sub-pillars nested, practice/task counts |
| POST | `/api/pillars` | Create a pillar (include `parent_id` for sub-pillar) |
| PUT | `/api/pillars/:id` | Update a pillar |
| DELETE | `/api/pillars/:id` | Delete a pillar (CASCADE deletes sub-pillars; unlinks practices) |
| POST | `/api/pillars/:id/practices/:pid` | Link a practice to a pillar |
| DELETE | `/api/pillars/:id/practices/:pid` | Unlink a practice |
| POST | `/api/pillars/:id/tasks/:tid` | Link a task to a pillar |
| DELETE | `/api/pillars/:id/tasks/:tid` | Unlink a task |

### UI: PillarsView

A dedicated view showing your growth pillars, each expandable to show sub-pillars and linked practices/tasks with aggregate progress.

```
┌─────────────────────────────────────────────────┐
│  Pillars of Growth              [+ New Pillar]  │
│  "Jesus increased in wisdom and stature,        │
│   and in favour with God and man" — Luke 2:52   │
├─────────────────────────────────────────────────┤
│                                                 │
│  🕊️  Spiritual                          82% ▲  │
│  ┌─────────────────────────────────────────┐    │
│  │  Scripture Study                        │    │
│  │   ✦ D&C 93:29           memorize ████  │    │
│  │   ✦ Mosiah 3:19         memorize ███   │    │
│  │  Prayer                                 │    │
│  │   ✦ Morning Prayer      habit    85%   │    │
│  │   ✦ Family Prayer       habit    90%   │    │
│  │  (no sub-pillar)                        │    │
│  │   ✦ Sacrament prep      scheduled      │    │
│  │  + Link a practice...                   │    │
│  └─────────────────────────────────────────┘    │
│                                                 │
│  💪 Physical                             79% ▲  │
│  ┌─────────────────────────────────────────┐    │
│  │  PT Recovery                            │    │
│  │   ✦ Clamshell           tracker  12/14 │    │
│  │   ✦ Plank               tracker  10/14 │    │
│  │   ✦ Bridge              tracker  11/14 │    │
│  │  + Link a practice... | + Sub-pillar   │    │
│  └─────────────────────────────────────────┘    │
│                                                 │
│  📖 Intellectual                         65% → │
│  🤝 Social                               40% ▼ │
│                                                 │
│  ── BALANCE ──────────────────────────────────  │
│  🕊️ ████████░░  82%                            │
│  💪 ███████░░░  79%                             │
│  📖 ██████░░░░  65%                             │
│  🤝 ████░░░░░░  40%  ← needs attention         │
└─────────────────────────────────────────────────┘
```

The **balance bar** at the bottom is key — it shows at a glance whether your growth is lopsided. The Savior grew in *all four areas*. If you're crushing your PT exercises but neglecting social connection, the app gently surfaces that.

### Integration Points

- **DailyView**: Optionally group by Pillar instead of category. A toggle: "Group by: Category | Pillar"
- **Reports**: Filter by Pillar in addition to type/category. Pillar balance chart.
- **PracticesView**: When creating/editing a practice, select which Pillar(s) it serves

### Design Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | Name for the concept | **Pillar** | Gospel-rooted (pillars of the temple, pillars of the community). Structural metaphor for growth areas. Avoids redundancy with "Becoming." |
| 2 | Default pillars | **4 seed pillars via onboarding** | Spiritual, Social, Intellectual, Physical — from Luke 2:52 and Children & Youth. User confirms/modifies during first launch. |
| 3 | Sub-pillars | **One level deep** (parent_id, NULL = top-level) | Simple tree: Pillar → Sub-pillar. Deeper nesting adds complexity without clear value. |
| 4 | Practice linking | **To either level** | A practice can link to "Physical" directly or to "Physical > PT Recovery". Sub-pillar implies parent. |
| 5 | Practices per Pillar | **Many-to-many** | Morning prayer serves "Spiritual" AND "Social" (family). Junction tables. |
| 6 | Pillars per practice | **Multiple allowed** | Same reason — a practice can serve multiple growth areas. |
| 7 | Required? | **No** | Practices can exist without a Pillar. Pillars are opt-in structure. |
| 8 | Aggregate stats | **Derived, not stored** | Calculate from linked practices' report data. No separate tracking. |
| 9 | Categories still used? | **Yes, unchanged** | Categories are freeform organizational labels. Pillars are structured growth areas. Different purposes. |

---

## Feature 2: Notes (Information Capture)

### The Problem

There's no lightweight place to jot things down. Practice logs have a `notes` field, but there's no standalone notes feature. You might want to capture a thought during the day, note something from a conversation, or write a quick reference — things that don't fit neatly into a practice log or journal reflection.

### The Concept

Notes are **outward-facing** — they capture *information about the world*. They're the facts, references, instructions, and observations that attach to your practices and tasks.

**Notes answer: "What do I need to remember?"**

| | Notes | Reflections |
|---|-------|-------------|
| **Direction** | Outward — about the world | Inward — about me |
| **Trigger** | "I need to remember this" | "What did that mean to me?" |
| **Tone** | Factual, brief, reference | Personal, introspective, growth |
| **Attached to** | A specific thing (practice, task, pillar) | The day as a whole |
| **Frequency** | Many per day, terse | One per day, thoughtful |
| **Examples** | "PT trainer said focus on form, not speed" | "Today I realized my PT exercises are teaching me patience" |
| | "D&C 93:29 cross-ref: Abraham 3:18-19" | "Studying intelligence today — I felt the Spirit confirm that learning IS worship" |
| | "Need to ask bishop about recommend renewal" | "I'm grateful for clear direction in prayer this morning" |

Think of it this way: **Notes are the bricks. Reflections are the capstone of the day.**

### Data Model

```sql
CREATE TABLE IF NOT EXISTS notes (
    id          INTEGER PRIMARY KEY,
    content     TEXT NOT NULL,            -- The note text (plain text or markdown)
    
    -- Optional foreign keys (at most one set)
    practice_id INTEGER REFERENCES practices(id) ON DELETE SET NULL,
    task_id     INTEGER REFERENCES tasks(id) ON DELETE SET NULL,
    pillar_id   INTEGER REFERENCES pillars(id) ON DELETE SET NULL,
    
    pinned      BOOLEAN DEFAULT 0,       -- Pin to top
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### API

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/notes` | List notes (optional filters: `?practice_id=`, `?task_id=`, `?pillar_id=`, `?pinned=`) |
| POST | `/api/notes` | Create a note |
| PUT | `/api/notes/:id` | Update a note |
| DELETE | `/api/notes/:id` | Delete a note |

### UI

Two modes:
1. **Standalone NotesView** — A searchable list of all notes, with filter pills for linked/unlinked. Quick-create at the top.
2. **Inline notes** — On practice detail, task detail, and pillar detail views, show attached notes with quick-add.

```
┌──────────────────────────────────────────────┐
│  Notes                            [+ New]    │
├──────────────────────────────────────────────┤
│  🔍 Search notes...                          │
│  [All] [Pinned] [Practices] [Tasks] [Free]  │
│                                              │
│  📌 PT trainer said focus on form, not speed │
│     — linked to: Clamshell, Bridge           │
│     Feb 12                                   │
│                                              │
│  D&C 93:29 cross-ref: Abraham 3:18-19.      │
│  Also see GS "Intelligence" entry.           │
│     — linked to: D&C 93:29 (memorize)        │
│     Feb 11                                   │
│                                              │
│  Need to ask bishop about temple recommend   │
│  renewal process.                            │
│     Feb 10                                   │
└──────────────────────────────────────────────┘
```

### Design Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | Markdown support? | **Yes, basic** | Bold, italic, links. Rendered on display, plain textarea on edit. |
| 2 | Attach to multiple? | **No, keep simple** | One optional link (practice OR task OR pillar OR none). If you need it in two places, write two notes. |
| 3 | Tags? | **Not yet** | The linked entity IS the tag. Free notes are untagged. Revisit if needed. |

---

## Feature 3: Reflections (Daily Meaning-Making)

### The Problem

The journal folder exists in the scripture-study project but isn't in the app. More importantly, daily reflection — "What did I learn? What am I grateful for? How did I grow today?" — is the bridge between doing and becoming. But in a busy life, a blank journal page is intimidating. It needs to be fast.

### The Concept

Reflections are **inward-facing** — they process *experience into meaning*. While notes capture facts about the world, reflections ask "What did that mean to me?" One per day, prompted or free-form, always attached to the date.

**Reflections answer: "What is this experience making of me?"**

(See the Notes vs Reflections comparison table in Feature 2.)

### Data Model

```sql
CREATE TABLE IF NOT EXISTS prompts (
    id          INTEGER PRIMARY KEY,
    text        TEXT NOT NULL,            -- The prompt question
    active      BOOLEAN DEFAULT 1,       -- Can be deactivated without deleting
    sort_order  INTEGER DEFAULT 0,       -- Controls rotation order
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reflections (
    id          INTEGER PRIMARY KEY,
    date        DATE NOT NULL,            -- One reflection per day (can be edited)
    prompt_id   INTEGER REFERENCES prompts(id) ON DELETE SET NULL,  -- Links to the prompt used
    prompt_text TEXT,                     -- Snapshot of the prompt (survives prompt deletion)
    content     TEXT NOT NULL,            -- The response
    mood        INTEGER,                 -- Optional 1-5 mood rating (simple emoji scale)
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(date)                         -- One per day, editable
);
```

**Note:** One reflection per day with `UNIQUE(date)`. If you open the reflection and it already exists for today, you edit it. No endless journal entries — just one thoughtful moment per day.

### Prompts System

Prompts are stored in the database from the start. Seed data provides defaults; the user can add, edit, reorder, and deactivate prompts. The app cycles through active prompts in `sort_order`, one per day.

**Seed prompts (inserted on first run):**
1. "What did I learn today?"
2. "What am I grateful for?"
3. "How did I see God's hand today?"
4. "What's one thing I did well?"
5. "What do I want to do better tomorrow?"
6. "What scripture spoke to me today?"
7. "How did I serve someone today?"

The prompt text is **snapshotted** into `reflections.prompt_text` when the reflection is created. This way if a prompt is later edited or deleted, the historical record shows what was actually asked on that day.

### API

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/reflections` | List reflections (paginated, optional `?from=&to=` date range) |
| GET | `/api/reflections/:date` | Get today's (or any date's) reflection |
| POST | `/api/reflections` | Create or update today's reflection (upsert on date) |
| DELETE | `/api/reflections/:id` | Delete a reflection |
| GET | `/api/prompts` | List all prompts (active and inactive) |
| POST | `/api/prompts` | Create a custom prompt |
| PUT | `/api/prompts/:id` | Update a prompt (text, active, sort_order) |
| DELETE | `/api/prompts/:id` | Delete a prompt |
| GET | `/api/prompts/today` | Get today's prompt (based on day-of-year % active prompt count) |

### UI: Two Entry Points

**1. DailyView integration (primary)**
A collapsible section at the bottom of the daily view. If today's reflection doesn't exist yet, it shows the prompt with a text area. Minimal friction — you're already on the daily page.

```
┌─────────────────────────────────────────────┐
│  TODAY'S HABITS                    7/7 ✓    │
│  ...                                        │
│                                             │
│  DAILY REFLECTION                           │
│  ┌─────────────────────────────────────┐    │
│  │ What did I learn today?             │    │
│  │                                     │    │
│  │ [                                 ] │    │
│  │ [                                 ] │    │
│  │                                     │    │
│  │  😟 😐 🙂 😊 😄  (optional mood)    │    │
│  │                          [Save]     │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  "The glory of God is intelligence"         │
└─────────────────────────────────────────────┘
```

**2. ReflectionsView (history/browse)**
A dedicated view for reading past reflections. Searchable, filterable by mood, shows prompt + response in a timeline.

```
┌──────────────────────────────────────────────┐
│  Reflections                                 │
├──────────────────────────────────────────────┤
│  🔍 Search reflections...                    │
│  [All] [😊 Happy] [📖 Scripture] [Week]     │
│                                              │
│  Feb 12, 2026  😊                            │
│  "What did I learn today?"                   │
│  Learned about the Hebrew word for           │
│  intelligence in D&C 93. It connects to...   │
│                                              │
│  Feb 11, 2026  🙂                            │
│  "How did I see God's hand today?"           │
│  My PT exercises went well. Felt guided to   │
│  try a new stretch that really helped.       │
│                                              │
│  Feb 10, 2026  😄                            │
│  "What am I grateful for?"                   │
│  The boys. This project. Clear mind today.   │
└──────────────────────────────────────────────┘
```

### Design Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | How many per day? | **One, editable** | UNIQUE(date). Keeps it simple. You can add to it throughout the day. |
| 2 | Prompts: fixed or custom? | **DB-stored from the start** | `prompts` table with seed data. Users can add, edit, reorder, deactivate. No hardcoded prompts. |
| 3 | Mood tracking? | **Optional 1-5** | Super low friction (one tap). Enables mood trends over time. |
| 4 | Length guidance? | **No minimum, no maximum** | A single sentence is fine. The goal is consistency, not volume. |
| 5 | Link to Pillars? | **Not directly** | Reflections are about the whole day, not a specific pillar. They naturally reference practices and growth areas in prose. |
| 6 | Prompt snapshot? | **Yes, store text at creation** | `prompt_text` column preserves what was asked even if the prompt is later edited or deleted. History stays accurate. |

---

## Feature 4: Trend Lines (Reports Enhancement)

### The Problem

The current Reports page shows summary cards and per-practice bar charts for a date range. But there's no visual story of *change over time*. "Am I becoming more consistent?" requires comparing this week to last month, and the current view doesn't show that.

### The Concept

Add line charts to the Reports page showing trends:
- **Overall completion rate** over time (weekly rolling average)  
- **Per-practice** completion trend (sparkline or expandable chart)
- **Mood trend** from reflections (once Feature 3 is built)
- **Streak history** — visual timeline of streak start/break/restart

### Implementation

No new backend needed — the existing `GET /api/reports` already returns `daily_data` (per-practice per-day). The trend lines are a **frontend-only enhancement**.

#### Chart Components

Use a lightweight chart approach (CSS/SVG — no heavy chart library dependency):

1. **Summary trend line** — A single SVG line chart at the top of Reports showing daily completion % across all practices. Rolling 7-day average smooths the noise.

2. **Per-practice sparklines** — In each practice card on Reports, replace or supplement the bar chart with a small line trend. Shows at-a-glance whether you're improving, stable, or declining.

3. **Mood overlay** (Phase 2.7b, after reflections) — Small colored dots or emoji on the timeline showing mood correlation with practice completion.

#### UI Sketch

```
┌─────────────────────────────────────────────────┐
│  Reports                          Last 30 days  │
├─────────────────────────────────────────────────┤
│                                                 │
│  COMPLETION TREND                               │
│  ┌─────────────────────────────────────────┐    │
│  │      .  .                               │    │
│  │  .  . .. . ..  .                        │    │
│  │ . ..       . .. ...  ...                │    │
│  │                  .  .. .....  .... .     │    │
│  │  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─   │    │
│  │  Jan 12              Jan 27    Feb 12   │    │
│  └─────────────────────────────────────────┘    │
│  Avg: 72% → 85%  ↑ 13% improvement             │
│                                                 │
│  PER-PRACTICE                                   │
│  ┌─────────────────────────────────────────┐    │
│  │ Clamshell  tracker  ▂▃▄▅▆▇▇▇  85% ↑   │    │
│  │ Plank      tracker  ▃▃▄▄▅▅▆▇  79% ↑   │    │
│  │ D&C 93:29  memorize ▅▆▆▇▇▇▇▇  95% →   │    │
│  │ Prayer     habit    ▇▇▆▇▇▇▇▇  96% →   │    │
│  └─────────────────────────────────────────┘    │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Design Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | Chart library? | **Pure SVG/CSS** | No new dependency. The data is simple (daily points over time). Canvas or D3 is overkill. |
| 2 | Rolling average? | **7-day** | Daily data is noisy (weekends, sick days). 7-day rolling smooths without hiding trends. |
| 3 | Trend direction | **Compare first half to second half of range** | Simple: avg(first 50% of days) vs avg(last 50%). Shows ↑ ↓ → |
| 4 | Mood on timeline? | **Phase 2.7b** | Build reflections first, then overlay mood data on trend charts. |

---

## Future Notes (Not Building Now)

### PWA / Mobile (Phase 4+)

The app needs to be in your hand at the gym, at church, during the day. Current priorities:

- **Service worker** for offline caching of the app shell
- **Web app manifest** for "Add to Home Screen" on iOS/Android
- **Responsive design** — already mostly there with Tailwind, but needs testing on actual devices
- **Touch optimization** — larger tap targets, swipe gestures for quick-log

The Vue 3 + Vite stack has excellent PWA plugin support (`vite-plugin-pwa`). When we're ready, it's a relatively straightforward add.

### Push Notifications (Phase 4+)

Scheduled practices lose effectiveness if you forget to open the app. Notifications would be the nudge:

- **Web Push API** for browser notifications (works on Android, desktop; limited on iOS)
- **Daily summary push** — "You have 3 cards due and 2 scheduled tasks"
- **Overdue alerts** — "Shave is 1 day overdue"
- **Reflection reminder** — Evening nudge: "Take a moment to reflect on today"

Requires a push subscription management system and a background job/cron in the Go backend.

### Social / Group Becoming (Phase 6+ — webeco.me)

The body of Christ working together. A late-phase vision:

- **Becoming Groups** — Create a group (family, ward, study group) with shared Pillars
- **Accountability partners** — See each other's streaks (opt-in, privacy-first)
- **Shared practice templates** — "PT Recovery Pack" or "Come Follow Me Weekly" that can be imported
- **Group reflections** — Share a reflection with your group (like a testimony in miniature)
- **Discussion threads** — Comment on shared reflections or study notes
- **webeco.me** domain for the social/community-facing side
- **ibeco.me** domain for the personal app

This is the "ward choir" version of the app — individual voices becoming something greater together. But it requires auth, multi-user, privacy controls, and real deployment infrastructure. Build the solo instrument first, then the orchestra.

---

## Build Order

These features are relatively independent and can be built in any order. Recommended sequence based on impact and dependency:

### Sprint 1: Trend Lines
**Why first:** Easiest win. Frontend-only changes to ReportsView. No new tables, no new API. Uses existing `daily_data` from reports endpoint. Immediately makes the Reports page more useful for showing your PT trainer progress.

**Scope:**
- SVG line chart component for overall completion trend
- Per-practice sparklines in report cards
- Trend direction indicator (↑ ↓ →)
- ~1-2 hours

### Sprint 2: Notes
**Why second:** Simple CRUD feature. Adds immediate utility — a place to jot things down that doesn't exist yet. Foundation for attaching context to practices.

**Scope:**
- `notes` table + migration
- Notes CRUD API (4 endpoints)
- NotesView (list, search, filter, create/edit)
- Inline notes on practice detail
- ~2-3 hours

### Sprint 3: Reflections
**Why third:** Builds the journal/reflection habit. Benefits from notes being done (similar patterns). The DailyView integration is the key — it should feel effortless.

**Scope:**
- `prompts` table + seed data (7 default prompts)
- `reflections` table + migration
- Prompts CRUD API (4 endpoints + `/api/prompts/today`)
- Reflections API (4 endpoints)
- DailyView reflection section (prompt + textarea + mood)
- ReflectionsView (history/browse)
- ~3-4 hours

### Sprint 4: Pillars
**Why last:** Most complex feature. Needs junction tables, sub-pillar hierarchy, onboarding flow, and careful UI. Benefits from everything else being stable. The vision layer is the capstone.

**Scope:**
- `pillars`, `practice_pillars`, `task_pillars` tables + migration
- Pillars CRUD API + link/unlink endpoints
- Onboarding screen: suggest 4 default pillars (Spiritual, Social, Intellectual, Physical), user can accept/skip/customize
- PillarsView (pillar cards with sub-pillar grouping, linked practices, balance bar)
- PracticesView integration (select Pillar when creating/editing)
- DailyView grouping toggle (Category | Pillar)
- Reports filter by Pillar
- ~5-6 hours

### Total estimated: ~11-15 hours of focused work

---

## How This Changes the App

Before Phase 2.7:
```
Practices → Logs → Reports
(what you DO)  (that you DID it)  (how much you DID)
```

After Phase 2.7:
```
Pillars → Practices → Logs → Reports + Trends
(WHY)      (WHAT)      (DID)   (HOW MUCH + DIRECTION)
                ↕
             Notes  ←→  Reflections
           (CONTEXT)    (MEANING)
```

The loop closes. **Why** you're doing it (Pillars) shapes **what** you practice. **Notes** capture context and insight along the way. **Reflections** synthesize the day's experience. **Trends** show the arc of change. And all of it serves the real goal: not to track practices, but to *become*.

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection." — D&C 130:18
