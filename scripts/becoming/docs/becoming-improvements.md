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

### Difficulty Ladder

Six levels, each tracked independently per card per user:

| Level | Mode | Description |
|-------|------|-------------|
| 1 | **Reveal Whole** | Show the full text, user reads it. Lowest barrier. |
| 2 | **Reveal Words** | ~35% of words blanked, tap to reveal one at a time |
| 3 | **Type Words / Arrange Words** | Randomly chosen. ~35% blanked + type, or all words shuffled + arrange. |
| 4 | **Type Full Text** | Blank canvas, user types entire verse from memory |
| 5 | **Reverse (Reference)** | Given the text, user must identify the scripture reference/title |

Level 5 (Reverse) is its own rung — easier than remembering the full text, harder than just reading it. It tests a different axis: "I know where this comes from."

### The Double-Spectrum Aptitude Model

This is the core insight: there are **two spectrums** operating simultaneously.

**Spectrum 1: Per-mode aptitude.** Each difficulty level has its own rolling aptitude score per card. A user might be great at Reveal Words but terrible at Type Full Text *for the same card*. These are independent skills.

**Spectrum 2: Overall card aptitude.** Across all modes, how well does the user know this card? This is the aggregate — a weighted combination of per-mode scores that represents total mastery.

```
Per-mode aptitude  = rolling average of last 3-5 scores at that mode for that card
Overall aptitude   = weighted average across all per-mode aptitudes
                     (higher-level modes weighted more heavily)
```

### Adaptive Algorithm — Goldilocks Selection

When Study Mode needs to pick a difficulty level for a card:

1. **Fresh card (little data):** Present all difficulty levels to gather data. Start with Level 1, but cycle through levels quickly to discover where the user actually is.

2. **Card with history:** Look at per-mode aptitudes:
   - **High aptitude at current level** (rolling avg ≥ 80%) → favor harder levels
   - **Low aptitude at current level** (rolling avg < 60%) → favor easier levels
   - **Mid-range** → stay at current level, occasionally test adjacent levels

3. **Level selection is probabilistic, not deterministic:**
   - The system *favors* the appropriate level but doesn't lock into it
   - Even a struggling user occasionally gets an easy win (Level 1)
   - Even a strong user occasionally gets challenged (Level 4-5)
   - This prevents staleness and provides ongoing calibration data

4. **Session momentum:**
   - Track running accuracy within a session
   - If user is crushing it (3+ good scores in a row) → bump difficulty more aggressively
   - If user is struggling (2+ poor scores in a row) → drop back immediately, don't let frustration build
   - The session should *feel* like it's responding to them in real time

5. **Level 3 randomization:** When the algorithm picks Level 3, it randomly chooses between Type Words and Arrange Words. Both are tracked as the same aptitude level.

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
    mode        TEXT NOT NULL,       -- 'reveal_whole', 'reveal_words', 'type_words', 'arrange', 'type_full', 'reverse'
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

## Decisions Log

Answers to open questions, recorded for future reference:

| Question | Decision | Rationale |
|----------|----------|-----------|
| Reverse mode placement | Its own level (Level 5) on the ladder | Easier than full recall but tests a different skill (reference knowledge). Own rung keeps it clean. |
| Study session length | Default: all due card reps. Then "Keep Studying" for open-ended. | Matches SM-2 contract (do your reps), but doesn't trap the user. Extra study is opt-in. |
| Promotion/demotion thresholds | Rolling average of last 3-5 scores per mode per card | Develops into an aptitude score. Small window handles cold-start; grows more stable over time. |
| Level 3 selection | System randomly picks between Type Words and Arrange Words | Same difficulty tier, variety keeps it fresh. Both map to same aptitude level. |
| Seed from existing data | Yes — use SM-2 quality history and interval to set initial levels | Users with strong review history shouldn't be bored at Level 1. One-time seed, then rolling aptitude takes over. |
| Study mode entry point | New mode within Memorize page, with pillar/category sub-filtering | Existing modes are for single-card focus. Study mode is for multi-card adaptive sessions. Sub-filtering lets users scope to "just scriptures" or a pillar. |

---

## Priority Order

1. **Sprint 1** — Practice Lifecycle schema/backend (unblocks everything else)
2. **Sprint 2** — Practice Lifecycle frontend (makes the app usable for non-memorize practices)
3. **Sprint 3** — Study Mode adaptive difficulty (the big UX win for memorization)
4. **Sprint 4** — Memorize card lifecycle (completes the picture)

Sprints 1+2 can probably be done together. Sprint 3 is the meatiest — it needs schema, backend, algorithm, and significant frontend work.
