# Title of Liberty — Becoming App Integration

How the Title of Liberty program maps to the existing Becoming app infrastructure (ibeco.me) and what new features are needed for the community layer (webeco.me).

---

## Current App Capabilities (ibeco.me)

The Becoming app already has the building blocks:

| Feature | Status | Title of Liberty Use |
|---------|--------|---------------------|
| **Practices** (habit, tracker, memorize, task types) | ✅ Built | Daily scripture study, prayer, exercise, chores — all tracked as practices |
| **Practice Logs** (per-day logging with value, sets, reps, duration) | ✅ Built | Track daily completion of program activities |
| **Tasks** (goals/commitments with status lifecycle) | ✅ Built | Merit badge requirements as tasks |
| **Memorization** (spaced repetition with SM-2 scoring) | ✅ Built | Scripture mastery requirements for every rank |
| **Pillars** (growth areas with hierarchy) | ✅ Built | Map directly to the 4 pillars: Faith & Covenant, Liberty & Service, Knowledge & Skill, Strength & Discipline |
| **Notes** (attached to practices/tasks/pillars) | ✅ Built | Badge work notes, family council minutes |
| **Reflections** (daily journal with prompts) | ✅ Built | Rank advancement reflections, daily journaling |
| **Reports** (activity heatmaps, practice reports) | ✅ Built | Track consistency, streaks, progress over time |
| **Scheduled practices** (interval, daily_slots, weekly, monthly) | ✅ Built | Weekly badge work sessions, monthly challenge nights |

### What Exists That We Just Need to Configure

1. **Pillars** — Create the 4 Title of Liberty pillars as custom pillars in the app:
   - ⚔️ Faith & Covenant (replaces default "Spiritual")
   - 🛡️ Liberty & Service (replaces default "Social")
   - 🏗️ Knowledge & Skill (replaces default "Intellectual")
   - 💪 Strength & Discipline (replaces default "Physical")

2. **Practices** — Create standing daily practices:
   - Scripture study (habit, daily, linked to Faith & Covenant pillar)
   - Prayer (habit, daily, linked to Faith & Covenant)
   - Physical activity (tracker, daily, linked to Strength & Discipline)
   - Service/kindness (habit, daily, linked to Liberty & Service)

3. **Memorization cards** — Add scripture mastery cards per rank:
   - Rank 1: 1 Nephi 3:7, selected verses from 1 Nephi 1–7
   - Rank 2: Additional verses from 1 Nephi
   - Rank 3: Alma 53:20–21 + additional
   - Rank 4: Alma 48:11–13 + additional
   - Rank 5: Comprehensive scripture mastery set

4. **Tasks** — Each merit badge becomes a task group with checklist items

---

## New Features Needed

### Phase 1: Individual Tracking Enhancements (ibeco.me)

These are small additions to the existing app that make Title of Liberty work for a single user:

#### 1.1 Badge / Achievement System

A lightweight achievement layer on top of existing tasks.

```sql
CREATE TABLE IF NOT EXISTS badges (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    name        TEXT NOT NULL,           -- "Scripture Mastery" or "Cooking"
    pillar_id   INTEGER REFERENCES pillars(id),
    badge_type  TEXT NOT NULL,           -- required | elective | honors
    rank        INTEGER NOT NULL,        -- 1-5 which rank this badge belongs to
    degree      INTEGER NOT NULL,        -- 1-4 which degree version
    status      TEXT NOT NULL DEFAULT 'not_started', -- not_started | in_progress | completed
    completed_at DATETIME,
    notes       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS badge_requirements (
    id          INTEGER PRIMARY KEY,
    badge_id    INTEGER NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    completed   BOOLEAN DEFAULT 0,
    completed_at DATETIME,
    evidence    TEXT,                    -- description of evidence / link to log
    sort_order  INTEGER DEFAULT 0
);
```

**API additions:**
- `GET /api/badges` — list badges (filter by rank, pillar, status)
- `POST /api/badges` — create badge
- `PUT /api/badges/:id` — update badge status
- `GET /api/badges/:id/requirements` — list requirements
- `PUT /api/badges/:id/requirements/:rid` — mark requirement complete

**UI:** A "Badges" tab in the app showing:
- Current rank and degree
- Required badges with progress bars
- Elective badges with completion status
- A badge grid (earned badges shown full color, unearned grayed out)

#### 1.2 Rank Tracking

```sql
CREATE TABLE IF NOT EXISTS rank_progress (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    rank        INTEGER NOT NULL,        -- 1-5
    degree      INTEGER NOT NULL,        -- 1-4
    status      TEXT NOT NULL DEFAULT 'in_progress', -- in_progress | completed
    started_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    reflection  TEXT,                    -- advancement reflection
    UNIQUE(user_id, rank, degree)
);
```

**UI:** A rank progression view showing the user's journey through the 5 ranks, with current position highlighted.

#### 1.3 Program Template

A way to load the full Title of Liberty program (all badges, requirements, suggested practices) from a template. So when a new participant joins, they don't have to manually create 59 badges — they load the "Title of Liberty" template and it populates their account.

---

### Phase 2: Family / Group Layer (webeco.me)

This is the community/group feature that makes the program shareable.

#### 2.1 Troops (Family Groups)

```sql
CREATE TABLE IF NOT EXISTS troops (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,           -- "Pucheta Family" or "Ward 3 Youth Group"
    leader_id   INTEGER NOT NULL REFERENCES users(id),
    program     TEXT NOT NULL DEFAULT 'title-of-liberty', -- which program template
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS troop_members (
    troop_id    INTEGER NOT NULL REFERENCES troops(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    role        TEXT NOT NULL DEFAULT 'member', -- leader | member
    joined_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (troop_id, user_id)
);
```

**Leader Dashboard:** Troop leaders see:
- All members' current rank/degree
- Badge progress per member
- Who's working on what this week
- Who's behind or stalled (gentle nudge system)
- Family council agenda generator (based on who's close to completing a badge or rank)

#### 2.2 Program Templates

```sql
CREATE TABLE IF NOT EXISTS program_templates (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,           -- "Title of Liberty"
    description TEXT,
    version     TEXT,                    -- semver
    author_id   INTEGER REFERENCES users(id),
    badge_data  TEXT NOT NULL,           -- JSON: full badge catalog with requirements
    rank_data   TEXT NOT NULL,           -- JSON: rank structure and advancement criteria
    public      BOOLEAN DEFAULT 0,      -- can other troops use this template?
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

This allows:
- Our family creates the Title of Liberty template
- Other families can discover and adopt it
- The template defines the badge catalog, rank structure, and degree scaling
- Each troop can customize (add custom badges, adjust requirements)

#### 2.3 Ceremonies & Milestones

Rank advancement triggers a "ceremony" workflow:
1. System detects all required badges complete
2. Leader gets a notification: "[Name] is ready for Rank [X] advancement"
3. Leader schedules a family council ceremony
4. Post-ceremony, leader confirms advancement
5. Participant writes their reflection
6. The milestone is recorded with date, reflection, and optionally a photo

---

## Integration Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Title of Liberty                        │
│                                                           │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │  ibeco.me   │    │  webeco.me   │    │ Flutter App  │ │
│  │  (personal) │    │ (community)  │    │  (mobile)    │ │
│  │             │    │              │    │              │ │
│  │ • Practices │    │ • Troops     │    │ • Daily view │ │
│  │ • Memorize  │    │ • Templates  │    │ • Quick log  │ │
│  │ • Badges    │    │ • Leader     │    │ • Memorize   │ │
│  │ • Ranks     │    │   dashboard  │    │ • Streaks    │ │
│  │ • Journal   │    │ • Ceremonies │    │              │ │
│  └──────┬──────┘    └──────┬───────┘    └──────┬───────┘ │
│         │                  │                    │         │
│         └──────────┬───────┘────────────────────┘         │
│                    │                                      │
│         ┌──────────▼──────────┐                           │
│         │   Go Backend API    │                           │
│         │   (chi v5 + SQLite) │                           │
│         │                     │                           │
│         │ • Existing API      │                           │
│         │ • + /api/badges     │                           │
│         │ • + /api/ranks      │                           │
│         │ • + /api/troops     │                           │
│         │ • + /api/templates  │                           │
│         └──────────┬──────────┘                           │
│                    │                                      │
│         ┌──────────▼──────────┐                           │
│         │   Becoming MCP      │                           │
│         │   (22 tools)        │                           │
│         │                     │                           │
│         │ • + badge_create    │                           │
│         │ • + badge_complete  │                           │
│         │ • + rank_advance    │                           │
│         │ • + troop_progress  │                           │
│         └─────────────────────┘                           │
└──────────────────────────────────────────────────────────┘
```

---

## MCP Tool Additions

New tools for the becoming-mcp server to support AI-assisted program management:

| Tool | Description |
|------|-------------|
| `badge_create` | Create a badge assignment for a user (from template or custom) |
| `badge_complete` | Mark a badge requirement as complete with evidence |
| `badge_list` | List badges by rank/pillar/status for a user |
| `rank_status` | Get a user's current rank, degree, and progress toward advancement |
| `rank_advance` | Record a rank advancement with reflection |
| `troop_progress` | Get summary progress for all members of a troop |
| `program_load` | Load a program template into a user's account |

---

## Implementation Priority

| Priority | Feature | Effort | Dependency |
|----------|---------|--------|------------|
| **1** | Configure 4 pillars in existing app | Minimal (config) | None |
| **2** | Add daily practices for scripture/prayer/exercise | Minimal (config) | None |
| **3** | Add scripture mastery cards for Rank 1 | Minimal (config) | None |
| **4** | Badge system (db + API + basic UI) | Medium (1-2 sessions) | None |
| **5** | Rank tracking (db + API + UI) | Small (1 session) | Badge system |
| **6** | Program template loading | Medium (1 session) | Badge system |
| **7** | Troop/family system (webeco.me) | Large (3-4 sessions) | Auth system (Plan 09) |
| **8** | Leader dashboard | Large (2-3 sessions) | Troop system |
| **9** | Ceremony workflow | Medium (1-2 sessions) | Rank tracking + Troops |

Priorities 1–3 can be done **today** with no code changes — just creating the right practices, pillars, and memorization cards in the existing app. Priorities 4–6 are new features but build on existing patterns. Priorities 7–9 require the multi-user infrastructure.
