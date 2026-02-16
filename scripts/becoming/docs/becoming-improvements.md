# Becoming App — Improvement Plans

*Created: 2026-02-16*
*Updated: 2026-02-16*
*Status: Planning*

---

## Problem Statement

The app isn't being used much because:
1. **Memorization games/cards need work** — no adaptive difficulty, no study mode that keeps users engaged
2. **Practices have no lifecycle** — can't archive, complete, or set an end date for ongoing practices
3. **These are the core features** — if they don't work well, people won't come back

There are real users signed up on ibeco.me, so any schema changes that aren't purely additive need proper migrations (both local SQLite and production PostgreSQL via goose).

---

## Sprint 1: Practice Lifecycle — Schema & Backend

### Current State
- `practices.active` (bool) — used as pause/resume toggle
- `practices.completed_at` (nullable timestamp) — exists in schema but **never set by any handler** (dead code)
- No `archived_at`, no `end_date`, no status field
- `tasks` table already has a proper `status` field (`active | completed | paused | archived`) — practices should follow this model

### Schema Changes (all additive — safe for existing users)

```sql
ALTER TABLE practices ADD COLUMN archived_at TIMESTAMPTZ;
ALTER TABLE practices ADD COLUMN end_date DATE;
ALTER TABLE practices ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

-- Backfill status from existing active/completed_at columns
UPDATE practices SET status = CASE
    WHEN completed_at IS NOT NULL THEN 'completed'
    WHEN active = FALSE THEN 'paused'
    ELSE 'active'
END;
```

**Status values:** `active | paused | completed | archived`

### Backend Changes
- `ListPractices` gains `?status=active|paused|completed|archived` filter (replaces `?active=`)
- New endpoints:
  - `POST /api/practices/{id}/complete` — sets `completed_at = now()`, `status = 'completed'`
  - `POST /api/practices/{id}/archive` — sets `archived_at = now()`, `status = 'archived'`
  - `POST /api/practices/{id}/restore` — clears timestamps, `status = 'active'`
- Completed/archived practices excluded from daily view and memorize due cards
- Keep backward compat: `?active=true` still works (maps to `status IN ('active')`)

### Migration Strategy
1. **PostgreSQL (production):** `002_practice_lifecycle.sql` via goose
2. **SQLite (local dev):** New entry in `runSQLiteMigrations()` in `db.go`
3. All changes are additive (new columns with defaults) — no data loss risk

---

## Sprint 2: Practice Lifecycle — Frontend

### PracticesView Tabs
- Tab bar: **Active** | **Paused** | **Completed** | **Archived**
- Default view: Active
- Each practice card gets action icons:
  - ✓ Complete (marks done, moves to Completed tab)
  - 📦 Archive (moves to Archived tab, out of sight)
  - ↩ Restore (from Completed/Archived back to Active)

### End Date Support
- Date picker on practice create/edit form ("Target end date")
- End date badge on DailyView cards (e.g., "3 days left" or "overdue")
- Optional: auto-prompt to complete when end date passes

### Completed/Archived Views
- Read-only history (all logs preserved)
- "Restore" button to bring back
- Stats summary (total logs, duration, date range)

---

## Sprint 3: Memorization Study Mode — Adaptive Difficulty

### Vision

A **Study Mode** that automatically selects memorization cards and adjusts difficulty based on user performance. The goal is a Goldilocks experience — not so easy they're bored, not so hard they give up.

The existing single-card modes (**Review**, **Practice**, **Quiz**) remain for focused work on one card at a time. **Study Mode** is a new mode within the Memorize page that works across all cards (with optional pillar/category sub-filtering).

### Difficulty Ladder — Forward & Reverse Tracks

Each difficulty level has two directions: **forward** (recall the text) and **reverse** (identify the reference). The algorithm picks both level *and* direction for each exercise.

**Forward track** (recall — "what does this scripture say?"):

| Level | Mode | Description |
|-------|------|-------------|
| 1 | **Reveal Whole** | Show the full text, user reads it. Lowest barrier. |
| 2 | **Reveal Words** | ~35% of words blanked, tap to reveal one at a time |
| 3 | **Type Words / Arrange Words** | Randomly chosen. ~35% blanked + type, or all words shuffled + arrange. |
| 4 | **Type Full Text** | Blank canvas, user types entire verse from memory |

**Reverse track** (recognition — "where is this from?"):

| Level | Mode | Description |
|-------|------|-------------|
| R1 | **Full Text → Reference** | All text shown, user identifies the reference. Easy — it's just matching. |
| R2 | **Partial Text → Reference** | Text with ~35% of words missing, user identifies the reference. |
| R3 | **Fragment → Reference** | Only 3-5 key words shown, user identifies the reference. Hard — requires deep familiarity. |

The reverse levels parallel the forward levels in difficulty:
- **R1** pairs with **Level 1-2** (easy tier)
- **R2** pairs with **Level 3** (medium tier)
- **R3** pairs with **Level 4** (hard tier)

This gives every exercise two dimensions: *how much help* (level) and *which direction* (forward vs. reverse). The algorithm treats forward and reverse aptitudes independently — a user might ace Type Full Text but struggle with Fragment → Reference, or vice versa.

### The Double-Spectrum Aptitude Model

This is the core insight: there are **two spectrums** operating simultaneously.

**Spectrum 1: Per-mode aptitude.** Each mode (forward *and* reverse) has its own rolling aptitude score per card. A user might be great at Reveal Words but terrible at Type Full Text, or nail Full Text → Reference but struggle with Fragment → Reference. These are all independent skills.

**Spectrum 2: Overall card aptitude.** Across all modes in both tracks, how well does the user know this card? This is the aggregate — a weighted combination of per-mode scores that represents total mastery.

```
Per-mode aptitude  = rolling average of last 3-5 scores at that mode for that card
                     (7 modes total: 4 forward + 3 reverse)
Overall aptitude   = weighted average across all per-mode aptitudes
                     (higher-level modes weighted more heavily)
```

### Adaptive Algorithm — Goldilocks Selection

**The core goal is emotional, not statistical.** The user should feel challenged but capable — never frustrated, never bored. The algorithm manages *session momentum* to keep them in that zone. If they start missing, we don't double down on hard cards — we give them ones they know well so they feel "I've actually learned this." If they're cruising, we gradually push harder until they start missing, then ease back. The session should feel like it's *with* them, not *testing* them.

#### Card + Level Selection

When Study Mode needs to pick the next exercise:

1. **Fresh card (little data):** Present all difficulty levels to gather data. Start with Level 1, but cycle through levels quickly to discover where the user actually is.

2. **Card with history:** Look at per-mode aptitudes:
   - **High aptitude at current level** (rolling avg ≥ 80%) → favor harder levels
   - **Low aptitude at current level** (rolling avg < 60%) → favor easier levels
   - **Mid-range** → stay at current level, occasionally test adjacent levels

3. **Level selection is probabilistic, not deterministic:**
   - The system *favors* the appropriate level but doesn't lock into it
   - This prevents staleness and provides ongoing calibration data

4. **Level 3 randomization:** When the algorithm picks Level 3 forward, it randomly chooses between Type Words and Arrange Words. Both are tracked as the same aptitude level.

5. **Direction selection:** After picking a difficulty tier, the algorithm chooses forward or reverse:
   - Both tracks have independent aptitudes, so the algorithm can favor whichever track needs more practice

#### Session Momentum — The Feel

The session-level algorithm sits *above* per-card aptitude. It watches the running trajectory and adjusts what gets served next:

| Session state | What the user sees | Why |
|---|---|---|
| **Struggling** (2+ poor scores in a row) | Drop back immediately. Serve high-aptitude cards at easy levels — things they *know*. | Rebuild confidence. Remind them they've actually learned things. Prevent frustration spiral. |
| **Steady** (mixed results) | Stay in the Goldilocks zone. Match cards to their current aptitude levels. | This is the target state. Challenged but comfortable. Learning happens here. |
| **Cruising** (3+ good scores in a row) | Gradually push harder. Bump up one level, or serve a lower-aptitude card. | Keep engagement. Boredom is as dangerous as frustration. |

**Intentional outliers** — even in steady state, the algorithm deliberately injects:
- **Stretch cards** (~15% of exercises): A difficulty level above where aptitude suggests. This probes whether the user has grown beyond their current level. They might miss it — that's fine, it's data. The next card should be something comfortable.
- **Confidence cards** (~15% of exercises): A high-aptitude card at an easy level. These are freebies. They exist purely so the user feels "I actually know this stuff." This is motivationally critical — especially after a stretch card or a miss.

The remaining ~70% are **Goldilocks cards**: matched to where the user's aptitude actually sits.

**Recovery pattern:** After a miss or a stretch card that went poorly, the *next* card should almost always be a confidence card. Never serve two hard misses in a row. The sequence should feel like: challenge → easy win → challenge → steady → steady → easy win → harder challenge. The user should always feel like the floor is close, even when the ceiling is being tested.
   - Fresh cards get a mix of both directions early on to gather data
   - Within a session, alternate between forward and reverse to keep variety high
   - The reverse track's difficulty tier maps to the forward tier: easy (R1 ↔ L1-2), medium (R2 ↔ L3), hard (R3 ↔ L4)

### Seeding from Existing Data

Users who already have SM-2 review history shouldn't start at Level 1 for every card. On first Study Mode launch:

- Cards with many quality 4-5 reviews → seed at Level 3-4
- Cards with mixed reviews → seed at Level 2
- New cards or cards with poor history → start at Level 1
- Cards with high SM-2 intervals (> 30 days) → seed higher

This is a one-time calculation per card. After that, the rolling aptitude takes over.

### Study Mode Session Flow

**Default mode: "Review Due Cards"**
1. User taps "Study" on the Memorize page
2. Optional: filter by pillar or category (e.g., "just scripture" or "just PT exercises")
3. System loads all due cards (SM-2 `next_review <= today`)
4. For each card, algorithm picks a difficulty level based on aptitude
5. User completes the exercise → score recorded → aptitude updated → SM-2 review logged
6. Next card selected, difficulty adjusted based on session momentum
7. Session ends when all due card reps are completed
8. Summary screen: cards reviewed, accuracy by level, aptitude changes

**Extra mode: "Keep Studying"**
- After all due reps are done, user can tap "Keep Studying" for open-ended practice
- System picks cards that would benefit from extra review (weakest aptitudes, longest since last review)
- No SM-2 impact (or reduced impact) — this is bonus practice, not scheduled review
- User stops whenever they want

### Score Tracking Schema

```sql
-- Per-exercise score history (the raw data for aptitude calculation)
CREATE TABLE memorize_scores (
    id          SERIAL PRIMARY KEY,
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    mode        TEXT NOT NULL,       -- Forward: 'reveal_whole', 'reveal_words', 'type_words', 'arrange', 'type_full'
                                     -- Reverse: 'reverse_full', 'reverse_partial', 'reverse_fragment'
    score       REAL NOT NULL,       -- 0.0 to 1.0 (accuracy)
    quality     INTEGER,             -- SM-2 quality 0-5 (user's self-rating)
    duration_s  INTEGER,             -- how long the exercise took
    date        DATE NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Per-card aptitude cache (denormalized for fast lookups, recalculated on each score)
CREATE TABLE memorize_aptitude (
    id            SERIAL PRIMARY KEY,
    practice_id   INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id       INTEGER NOT NULL REFERENCES users(id),
    mode          TEXT NOT NULL,     -- same mode values as memorize_scores
    aptitude      REAL NOT NULL,     -- rolling average, 0.0 to 1.0
    sample_count  INTEGER DEFAULT 0, -- how many scores in the rolling window
    last_score_at TIMESTAMPTZ,
    UNIQUE(practice_id, user_id, mode)
);

-- Overall card difficulty level (what level the algorithm currently targets)
ALTER TABLE practices ADD COLUMN memorize_level INTEGER DEFAULT 1;
```

### What Stays vs. What Changes

| Feature | Status |
|---------|--------|
| **Review mode** (tap-to-flip flashcard) | Stays — for focused single-card review |
| **Practice mode** (reveal/type/order) | Stays — for focused single-card practice |
| **Quiz mode** (type full text) | Stays — for focused single-card testing |
| **Study mode** (NEW) | Added — adaptive multi-card sessions |
| **SM-2 scheduling** | Stays — determines *when* cards are due |
| **Adaptive difficulty** | New — determines *how* cards are practiced in Study Mode |

---

## Sprint 4: Memorize Card Lifecycle

- **Mastered state:** After N consecutive quality-5 reviews with interval > 30 days, auto-suggest "Mark as memorized"
- **Complete a card:** One-tap to move to Completed, preserving all review history and aptitude data
- **Archive a card:** Remove from rotation without deleting history
- **End date / target:** "Memorize by [date]" with countdown on card
- **Card progression indicator:** Visual progress (streak, current level, per-mode aptitudes, "Rep 3/5 today")
- **Aptitude dashboard:** Per-card breakdown of mode aptitudes — user can see where they're strong/weak

---

## Sprint 6: Start Date & Future Planning

### Vision
Change `created_at` to a separate `start_date` field that defaults to `created_at` but is editable. This lets users plan future practices — e.g., schedule Come Follow Me lessons for upcoming weeks, or set a memorization card to start when they'll need it.

### Schema
```sql
ALTER TABLE practices ADD COLUMN start_date DATE;
UPDATE practices SET start_date = date(created_at);
```

### Behavior
- `start_date` replaces `created_at` in the temporal daily summary query
- Practices with `start_date > today` don't show in daily view until that date
- Editable in practice create/edit form (date picker, defaults to today)
- DailyView shows "starts in X days" badge for future-scheduled practices on the Practices page

---

## Decisions Log

Answers to open questions, recorded for future reference:

| Question | Decision | Rationale |
|----------|----------|-----------|
| Reverse mode placement | Parallel track with 3 tiers (R1 full text, R2 partial, R3 fragment) running alongside forward levels | Tests a complementary skill (recognition vs. recall). Parallel tracks double exercise variety at every difficulty tier without adding algorithmic complexity. R1=easy, R2=medium, R3=hard mirrors the forward ladder. |
| Study session length | Default: all due card reps. Then "Keep Studying" for open-ended. | Matches SM-2 contract (do your reps), but doesn't trap the user. Extra study is opt-in. |
| Promotion/demotion thresholds | Rolling average of last 3-5 scores per mode per card | Develops into an aptitude score. Small window handles cold-start; grows more stable over time. |
| Level 3 selection | System randomly picks between Type Words and Arrange Words | Same difficulty tier, variety keeps it fresh. Both map to same aptitude level. |
| Seed from existing data | Yes — use SM-2 quality history and interval to set initial levels | Users with strong review history shouldn't be bored at Level 1. One-time seed, then rolling aptitude takes over. |
| Study mode entry point | New mode within Memorize page, with pillar/category sub-filtering | Existing modes are for single-card focus. Study mode is for multi-card adaptive sessions. Sub-filtering lets users scope to "just scriptures" or a pillar. |

---

## Priority Order

1. **Sprint 1** — Practice Lifecycle schema/backend (unblocks everything else) ✅
2. **Sprint 2** — Practice Lifecycle frontend (makes the app usable for non-memorize practices) ✅
3. **Sprint 3** — Study Mode adaptive difficulty (the big UX win for memorization) ✅
4. **Sprint 4** — Memorize card lifecycle (completes the picture)
5. **Sprint 5** — Activity Calendar Heatmap

Sprints 1-3 are done. Sprint 4 is next — memorize card lifecycle (pause, complete, archive cards).

---

## Sprint 5: Activity Calendar Heatmap

### Vision

A GitHub-style contribution heatmap on a Calendar view, giving a 30,000 ft view of how active you are at *becoming*. Each day is a small square whose color intensity reflects how much activity was logged that day.

### Design

- **Grid:** 7 rows (days of week) × ~13 columns (weeks), showing ~90 days by default (expandable to full year)
- **Colors:** White box with black border = no activity; light orange → vibrant orange for increasing activity
- **Activity metric:** Total practice logs for the day. Could also weight by practice type (e.g., memorize cards count more than a simple habit check)
- **Interaction:**
  - Hover: tooltip showing date + log count + top practices
  - Click: navigates to that day's DailyView

### Backend

- New endpoint: `GET /api/reports/activity?start=YYYY-MM-DD&end=YYYY-MM-DD`
- Returns: `[{ date: "2026-02-16", log_count: 5, practice_count: 3 }, ...]`
- Query: `SELECT date, COUNT(*) as log_count, COUNT(DISTINCT practice_id) as practice_count FROM practice_logs WHERE practice_id IN (SELECT id FROM practices WHERE user_id = ?) AND date BETWEEN ? AND ? GROUP BY date`

### Frontend

- New `CalendarView.vue` (or section within ReportsView)
- SVG or CSS Grid rendering of the heatmap
- Color scale: 0 logs = `#ebedf0` (gray), 1-2 = `#ffedd5`, 3-5 = `#fdba74`, 6-9 = `#f97316`, 10+ = `#ea580c`
- Month labels along top, day-of-week labels along left
- Responsive: full grid on desktop, scrollable on mobile
