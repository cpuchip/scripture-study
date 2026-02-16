# Becoming App — Improvement Plans

*Created: 2026-02-16*
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

### Difficulty Ladder

Each memorization card has its own difficulty level per user, tracked independently:

| Level | Mode | Description |
|-------|------|-------------|
| 1 | **Reveal Whole** | Show the full text, user reads it. Lowest barrier. |
| 2 | **Reveal Words** | ~35% of words blanked, tap to reveal one at a time |
| 3 | **Type Words** | ~35% of words blanked, user types them in |
| 3 | **Arrange Words** | All words shuffled, user arranges in order (same difficulty as Type Words) |
| 4 | **Type Full Text** | Blank canvas, user types entire verse from memory |

**Reverse mode** (at any level): Given the text, user must identify the reference/title. This tests a different axis — recognition vs. location knowledge.

### Adaptive Algorithm

The system tracks **per-card, per-mode scores** over time to determine user skill level:

```
Card Skill = f(recent_scores, mode_history, trend)
```

**Rules:**
1. **New cards** start at Level 1 (Reveal Whole)
2. **Promotion:** After N consecutive good scores (e.g., ≥80% accuracy, quality ≥ 4) at a level → promote to next level
3. **Demotion:** After M poor scores (e.g., < 60% accuracy, quality ≤ 2) at a level → demote to previous level
4. **Full text typing** (Level 4) is only unlocked for cards where the user consistently rates knowledge high (quality 4-5 for several reviews)
5. **If they miss a lot, drop back to Reveal** — don't let frustration build
6. **Mix it up:** Within a study session, vary the modes to keep engagement (don't do 10 Reveal Wholes in a row)

### Study Mode Session Flow

1. User taps "Study" → enters Study Mode
2. System selects a card based on:
   - SM-2 due date (prioritize overdue cards)
   - Card skill level (determines which mode to present)
   - Variety (mix different cards and modes)
3. User completes the exercise
4. Score is recorded → skill level adjusts → SM-2 review logged
5. Next card is selected, adjusting difficulty based on session performance
6. **Session momentum:** If user is doing well mid-session, occasionally bump difficulty. If they're struggling, ease off.

### Score Tracking Schema

Need to track per-card, per-mode performance history for the adaptive algorithm:

```sql
-- New table or extend practice_logs
CREATE TABLE memorize_scores (
    id          SERIAL PRIMARY KEY,
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    mode        TEXT NOT NULL,     -- 'reveal_whole', 'reveal_words', 'type_words', 'arrange', 'type_full', 'reverse'
    score       REAL NOT NULL,     -- 0.0 to 1.0
    quality     INTEGER,           -- SM-2 quality 0-5
    duration_s  INTEGER,           -- how long the exercise took
    date        DATE NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Per-card skill level (denormalized for fast lookups)
ALTER TABLE practices ADD COLUMN memorize_level INTEGER DEFAULT 1;
```

### What Stays vs. What Changes

The existing modes (**Review**, **Practice**, **Quiz**) remain as manual options — the user can always choose a specific mode. **Study Mode** is a new fourth option that handles selection automatically.

The SM-2 scheduling system stays as-is — it determines *when* cards are due. The adaptive difficulty determines *how* the user practices them.

---

## Sprint 4: Memorize Card Lifecycle

- **Mastered state:** After N consecutive quality-5 reviews with interval > 30 days, auto-suggest "Mark as memorized"
- **Complete a card:** One-tap to move to Completed, preserving all review history
- **Archive a card:** Remove from rotation without deleting history
- **End date / target:** "Memorize by [date]" with countdown on card
- **Card progression indicator:** Visual progress (streak, current level, "Rep 3/5 today")

---

## Open Questions

1. **Reverse mode placement:** Given the text → guess the reference. Does this belong at every difficulty level as a toggle, or is it its own level in the ladder? (My instinct: it's orthogonal — can appear at any level as a bonus challenge.)

2. **Study session length:** Should Study Mode have a target? Options:
   - Time-based: "Study for 10 minutes"
   - Card-based: "Do 10 cards"
   - Open-ended: Keep going until the user stops
   - SM-2 driven: "Review all due cards" (current behavior)

3. **Promotion/demotion thresholds:** What feels right?
   - Promote after 3 consecutive ≥80% scores at a level?
   - Demote after 2 consecutive <60% scores?
   - Or should it be a rolling average over the last N attempts?

4. **Level 3 selection (Type Words vs. Arrange Words):** When the user is at Level 3, does the system randomly pick between Type Words and Arrange Words, or does the user choose?

5. **Existing review data:** Users have SM-2 review history. Should we use that to seed initial difficulty levels? (e.g., cards with many quality-5 reviews start at a higher level)

6. **Study mode entry point:** Separate nav item? Button on the Memorize page? Replace the current mode selector?

---

## Priority Order

1. **Sprint 1** — Practice Lifecycle schema/backend (unblocks everything else)
2. **Sprint 2** — Practice Lifecycle frontend (makes the app usable for non-memorize practices)
3. **Sprint 3** — Study Mode adaptive difficulty (the big UX win for memorization)
4. **Sprint 4** — Memorize card lifecycle (completes the picture)

Sprints 1+2 can probably be done together. Sprint 3 is the meatiest — it needs schema, backend, algorithm, and significant frontend work.
