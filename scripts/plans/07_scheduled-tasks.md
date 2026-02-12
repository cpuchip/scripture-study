# Scheduled Tasks / Recurring Routines — Design Plan

*Created: February 12, 2026*
*Context: Extending the Becoming app with frequency-based task scheduling*

---

## The Problem

The current practice system handles things you do **every day** (habits, trackers, memorize cards). But many real-life tasks follow more complex schedules:

- **Every other day**: Shaving, changing pants, watering some plants
- **Multiple times per day**: Pills (morning/lunch/dinner), prayer, water intake
- **Specific days per week**: Laundry on Monday, grocery shopping on Saturday
- **Monthly**: Pay bills, temple attendance, fasting
- **One-time with deadline**: Complete a project, read a specific book by X date

The current `habit` and `task` types don't model frequency. A habit just has `{"frequency": "daily"}` and tasks have `{"due_date": "..."}`. Neither supports "every 2 days starting from when I last did it" or "Mon/Wed/Fri" scheduling.

### Real-world examples from the user

| Task | Frequency | Notes |
|------|-----------|-------|
| Shave | Every 2 days | If done early, shift the next due date forward |
| Change pants | Every 2 days | Same shifting behavior |
| Water plants | Every 3 days | Some plants weekly |
| Take pills | 3x daily (morning, lunch, night) | Time-of-day slots |
| Laundry | Weekly (Saturday) | Fixed day |
| Pay rent | Monthly (1st) | Fixed day-of-month |

---

## Design Principles

1. **Low friction** — Checking things off should take 1-2 taps. The app should show "what's due now?" without the user calculating.
2. **Flexible shifting** — For "every N days" tasks, doing it early should shift the whole schedule forward. Doing it late should reset from today, not pile up missed instances.
3. **Routines/groups** — Tasks can be grouped into routines ("Morning Routine", "Getting Ready") that serve as quick checklists.
4. **Progressive disclosure** — Simple tasks (daily/weekly) should be simple to create. Advanced scheduling (multi-daily, shifting intervals) available but not mandatory.
5. **Offline-first** — All logic client-computable from the schedule config + log history.

---

## Data Model

### Option A: Extend `practices` table (Recommended)

Add a new practice type `scheduled` (or extend existing `habit`/`task` types) with richer config:

```json
// Practice type: "scheduled"
// Config examples:

// Every other day (interval-based, shifts on early completion)
{
  "schedule": {
    "type": "interval",
    "interval_days": 2,
    "anchor_date": "2026-02-12",
    "shift_on_early": true
  }
}

// Multiple times per day
{
  "schedule": {
    "type": "daily_slots",
    "slots": ["morning", "lunch", "night"]
  }
}

// Specific days of week
{
  "schedule": {
    "type": "weekly",
    "days": ["mon", "wed", "fri"]
  }
}

// Monthly on a specific day
{
  "schedule": {
    "type": "monthly",
    "day_of_month": 1
  }
}

// One-time with due date
{
  "schedule": {
    "type": "once",
    "due_date": "2026-03-15"
  }
}
```

### Why extend `practices` instead of a new table?

The `practices` + `practice_logs` system already handles:
- CRUD for trackable items
- Per-day logging with flexible value fields
- Active/inactive toggling
- Category grouping
- Daily summary queries

Adding `type: "scheduled"` with schedule config in the JSON `config` field means:
- No schema migration needed
- All existing UI patterns (DailyView groups, PracticesView CRUD, HistoryView charts) work
- Log entries work the same way (`practice_logs` with date, notes, etc.)

### Routine Grouping

Use the existing `category` field for routine grouping. Categories already render as sections in DailyView:
- category: "morning" → "MORNING" section header
- category: "getting ready" → "GETTING READY" section header

For explicit routine ordering within a category, use `sort_order`.

---

## Schedule Engine

### Core Algorithm

```
isDue(practice, date) → boolean | slot[]
```

Given a practice's schedule config and the log history, determine if it's due on a given date.

#### Interval-based ("every N days")

```
lastDone = most recent log date for this practice
if lastDone is null:
    due = anchor_date <= date
else if shift_on_early is true:
    due = date >= lastDone + interval_days
else:
    due = date >= anchor_date + (N * interval_days) for some N
```

**Shift behavior**: If shaving is every 2 days and anchor is Feb 12, normally due Feb 12, 14, 16...
If user shaves on Feb 13 (early), next due becomes Feb 15 (13 + 2), not Feb 14.

#### Daily slots ("N times per day")

```
logsToday = count logs for this practice today
slots = config.schedule.slots (e.g., ["morning", "lunch", "night"])
remainingSlots = slots - logsToday (tracking which slots are done via log.value field)
due = remainingSlots.length > 0
```

The `value` field on the log identifies which slot was completed: `log.value = "morning"`.

#### Weekly ("specific days")

```
dayOfWeek = date.getDay() // 0=Sun, 1=Mon, ...
due = config.schedule.days.includes(dayName)
```

#### Monthly ("specific day of month")

```
due = date.getDate() === config.schedule.day_of_month
```

#### One-time

```
due = date >= config.schedule.due_date && no completion log exists
```

### Schedule Resolution on Daily Summary

The DailySummary endpoint (`GET /api/daily/{date}`) currently returns all active practices. For scheduled practices, the server should **filter or annotate** based on whether they're due:

Option 1: **Filter server-side** — Only include scheduled practices that are due. Simple but hides "coming up" items.

Option 2: **Annotate** (Recommended) — Return all active practices but add an `is_due` boolean to the summary. The frontend decides whether to show non-due items (e.g., grayed out, collapsed, or hidden).

This means extending `DailySummary` with:
```go
type DailySummary struct {
    // ... existing fields ...
    IsDue    bool   `json:"is_due"`
    NextDue  string `json:"next_due,omitempty"`  // YYYY-MM-DD
    SlotsDue []string `json:"slots_due,omitempty"` // for daily_slots type
}
```

---

## Frontend UX

### DailyView Changes

**Current layout**:
```
MORNING
  ✓ Scripture study
  ✓ Prayer

PT
  [✓ 1] [  2]  Bridge  12 reps
```

**New layout with scheduled tasks**:
```
━━━ MORNING ROUTINE ━━━
  ☐ Take pills (morning)      every day · 3 slots
  ✓ Scripture study
  ☐ Prayer

━━━ GETTING READY ━━━
  ✓ Shower                     ← done today
  ☐ Shave                      due today (every 2 days)
  · Change pants               not due (tomorrow)

━━━ PT ━━━
  [✓ 1] [  2]  Bridge  12 reps
```

**Key UX decisions**:
- Items due today: normal appearance, checkable
- Items not due today: lighter/grayed, optionally hidden via a "show all" toggle
- Items done today that weren't due: still shown as checked (user overrode the schedule)
- Slot-based items: show which slots are done (pills: ✓morning  ☐lunch  ☐night)

### PracticesView Form Changes

When creating/editing a scheduled practice, the form shows:

```
Type: [Scheduled ▾]

Schedule:
  ○ Every day
  ○ Every ___ days  [2]  ☑ Shift if done early
  ○ Multiple times/day: [morning] [lunch] [night] [+]
  ○ Specific days: ☑Mon ☐Tue ☑Wed ☐Thu ☑Fri ☐Sat ☐Sun
  ○ Monthly on day: [1]
  ○ One-time, due: [____-__-__]
```

### HistoryView Additions

For scheduled practices, the history view could show:
- Compliance rate: "On schedule 85% of the time"
- Average interval: "Averaging every 2.1 days" (for interval-based)
- Completed slots chart (for multi-daily)

---

## Implementation Phases

### Phase A: Core Scheduling (MVP)

1. **Backend**: Add schedule engine functions in `internal/db/schedule.go`
   - `IsScheduledDue(practice, date, logs) → ScheduleStatus`
   - Update `GetDailySummary` to compute due status for scheduled practices
   - Handle `shift_on_early` interval logic

2. **Frontend**: Extend PracticesView form with schedule config UI
   - Radio group for schedule type
   - Dynamic fields per type (interval input, day checkboxes, slot list)

3. **Frontend**: Update DailyView to show due/not-due state
   - Gray out non-due scheduled items
   - Show due badge/indicator
   - Support slot-based items (pills)

### Phase B: Routines & Quick Mode

1. **Routine view**: Tap a routine name to expand/collapse its items
2. **Quick check mode**: Swipe or tap through items in sequence — optimized for "I'm getting ready, let me check things off as I go"
3. **Sort within groups**: Drag to reorder items within a category/routine

### Phase C: Notifications (Future)

1. **Browser notifications** (Service Worker)
   - "Time for your morning pills"
   - "Shaving is due today"

2. **Mobile push** (requires PWA or native wrapper)
   - Notification scheduling based on slot configuration
   - Snooze/complete from notification

3. **Desktop toast** (Electron or system tray app — very future)

### Phase D: Calendar Integration (Far Future)

- Export scheduled tasks as iCal feed
- Google Calendar / Apple Calendar sync
- Show practices on a calendar month view
- Block time for study sessions

---

## Schema Changes Needed

### No DDL migration required

The `practices` table already has:
- `type TEXT NOT NULL` — add `"scheduled"` as a valid type
- `config TEXT DEFAULT '{}'` — store schedule config here
- `category TEXT` — use for routine grouping

The `practice_logs` table already has:
- `value TEXT` — use for slot identification ("morning", "lunch", "night")
- `date DATE NOT NULL` — for schedule calculations

### Backend Code Changes

1. **New file**: `internal/db/schedule.go`
   - Schedule config types (Go structs matching the JSON)
   - `IsScheduledDue()` engine
   - Helper: `NextDueDate()` for display

2. **Modify**: `internal/db/logs.go`
   - Update `GetDailySummary()` to compute `is_due` / `next_due` / `slots_due`
   - Need to query last log date per scheduled practice for interval calculation

3. **Modify**: `internal/api/router.go`
   - Update `createPractice` to validate schedule configs
   - No new routes needed (uses existing practice/log CRUD)

### Frontend Code Changes

1. **Modify**: `PracticesView.vue`
   - Schedule type radio group
   - Dynamic config form per schedule type
   - Day picker, slot editor, interval input

2. **Modify**: `DailyView.vue`
   - New template section for `scheduled` type
   - Due/not-due styling
   - Slot completion UI (for multi-daily)
   - "Show upcoming" toggle

3. **Modify**: `HistoryView.vue`
   - Compliance rate calculation
   - Schedule-aware chart annotations

---

## Open Questions

> **All decisions recorded February 12, 2026.**

1. **Should `scheduled` be a new type or extend `habit`?**
   - **Decision: New type `scheduled`.** Keep `habit` as simple "did I do this today?" items. No migration needed.

2. **Overdue stacking?**
   - **Decision: Single due, with overdue badge.** Show as "due (2 days overdue)" — one completion catches up. No stacking.

3. **Non-due items in daily view?**
   - **Decision: Show grayed out.** Always visible but dimmed, with "tomorrow" / "in 2 days" label. No toggle needed.

4. **MVP scope — which schedule types?**
   - **Decision: All five.** Interval, daily slots, weekly, monthly, one-time. Full coverage from the start.

5. **Shift-on-early for intervals?**
   - **Decision: Per-task toggle (`shift_on_early`).** Default true for new items, user can disable per task.

6. **Daily slots UI?**
   - **Decision: Inline slot buttons.** Row of pill-shaped buttons: `[✓ morning] [☐ lunch] [☐ night]`

7. **Navigation / tab?**
   - **Decision: No new tab.** Scheduled items live in the existing Today view, grouped by category like everything else.

8. **Build order?**
   - **Decision: Backend first.** Schedule engine → create practice validation → daily summary integration → then frontend forms → daily view.

---

## Summary

This feature transforms the "Become" practice types from flat daily items to a schedule-aware system. The key insight is that the existing `practices` + `practice_logs` architecture handles all this — we just need:

1. Richer config JSON for schedule definitions
2. A schedule engine to compute "is due?" from config + logs
3. Frontend forms to configure schedules
4. DailyView updates to show due/not-due states

No schema migration. No new tables. Just smarter logic on top of the existing model.
