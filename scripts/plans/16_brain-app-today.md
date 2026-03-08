# Brain-App Today Screen — ibeco.me Bridge

*Created: March 8, 2026*
*Priority: Near-term*
*Depends on: Plan 15 Gap 1 (entry sync on launch) for brain entries; ibeco.me API already has all needed endpoints*
*Connects to: Plan 07 (Scheduled Tasks), Plan 17 (Proactive Surfacing)*

---

## Vision

A **unified daily view** in brain-app that surfaces everything due today — practices from ibeco.me, scripture memorize cards, brain actions, and scheduled tasks. One place to start the day.

Right now, brain-app is a **capture tool** and ibeco.me is a **practice/growth tool** and they're separate worlds. The Today Screen bridges them: your morning starts with brain-app showing "here's your day" from both systems.

---

## ibeco.me Endpoints Available

These already exist and are production-deployed:

| Endpoint | Returns | Use |
|----------|---------|-----|
| `GET /api/daily/{date}` | Practices due today + completion status | Main practice list |
| `GET /api/memorize/due/{date}` | Memorize cards due for review | SM-2 spaced repetition |
| `GET /api/memorize/study/next` | Next card in study queue | Study session flow |
| `POST /api/memorize/review` | Submit card review (quality 0-5) | Record review result |
| `GET /api/practices` | All practices (active/paused/archived) | Practice management |
| `POST /api/practices/{id}/complete` | Mark practice complete for today | One-tap completion |
| `GET /api/tasks` | Task list (with due dates) | Brain tasks surfaced here too |
| `GET /api/reports/activity` | Activity stats | Optional: streak/stats widget |

---

## Design

### Bottom Nav Addition

Current brain-app nav: **Home** (thought capture) | **History** (entries)

Proposed: **Home** | **Today** | **History**

Today tab icon: calendar or sun (☀️). Badge count shows number of incomplete items.

### Today Screen Layout

```
┌─────────────────────────────┐
│  Today — March 8            │
│  ─────────────────────────  │
│                             │
│  📖 Scripture Memorize (3)  │  ← Expandable section
│  ┌─────────────────────┐   │
│  │ D&C 93:36  [Review] │   │  ← Tap to start review
│  │ Alma 32:21 [Review] │   │
│  │ Moses 6:57 [Review] │   │
│  └─────────────────────┘   │
│                             │
│  ✅ Practices (5/8)        │  ← Section with progress
│  ┌─────────────────────┐   │
│  │ ☑ Morning prayer     │   │  ← Completed (tapped earlier)
│  │ ☑ Scripture study    │   │
│  │ ☐ Exercise           │   │  ← Tap to complete
│  │ ☐ Journal            │   │
│  │ ☐ Evening prayer     │   │
│  │ ...                  │   │  ← Expandable if >5
│  └─────────────────────┘   │
│                             │
│  🧠 Brain Actions (2)      │  ← Due/overdue brain entries
│  ┌─────────────────────┐   │
│  │ Follow up with Josh  │   │  ← Tap to open in edit screen
│  │ Review project plan  │   │
│  └─────────────────────┘   │
│                             │
│  📊 Streak: 14 days        │  ← Optional: motivational stat
│                             │
└─────────────────────────────┘
```

### Interaction Patterns

**Practice completion**: Tap checkbox → `POST /api/practices/{id}/complete` → checkbox animates to filled. No confirmation needed (reversible via ibeco.me web UI).

**Memorize review**: Tap "Review" → opens inline card review flow:
1. Show scripture reference + prompt
2. User recalls (or taps "Show")
3. Rate recall quality (0-5 buttons, or simplified: Again / Hard / Good / Easy)
4. `POST /api/memorize/review` with quality rating
5. Card slides away, next card appears (or section shows "All done! 🎉")

**Brain actions**: Tap entry → navigates to EditEntryScreen (existing). Due date already tracked in brain entry metadata.

### Data Flow

```
App launch / Today tab focus
  ├── GET /api/daily/{today}        → practices + completion status
  ├── GET /api/memorize/due/{today} → memorize cards
  └── Filter local brain entries    → entries with due_date <= today
      (or: GET /api/brain/entries?due_before={today} if ibeco.me caches them)

Each section updates independently:
  - Practice tap → POST complete → update local state
  - Card review → POST review → remove card from due list
  - Brain action tap → navigate to edit screen (existing flow)
```

### Offline Behavior

Practices and memorize cards are fetched on tab focus. If offline:
- Show cached data from last fetch (timestamp shown: "Updated 2h ago")
- Practice completions and card reviews queue in OfflineQueue (already exists in brain-app)
- Brain actions always available (local data)

---

## Implementation Phases

### Phase 1: Practices (MVP)

1. Add `TodayScreen` widget with bottom nav integration
2. Fetch `/api/daily/{date}` on tab focus
3. Render practice list with tap-to-complete
4. Cache response for offline viewing
5. Queue completions in OfflineQueue when offline

### Phase 2: Memorize Cards

1. Fetch `/api/memorize/due/{date}` 
2. Inline card review flow (show/rate/next)
3. POST reviews to ibeco.me

### Phase 3: Brain Actions

1. Filter local entries by `due_date <= today`
2. Show in "Brain Actions" section
3. Tap navigates to existing EditEntryScreen

### Phase 4: Widget (Android)

1. Surface top 3 due items in existing Android widget
2. Mix practices + memorize + brain actions by priority
3. Tap widget item → deep-link to Today tab

---

## Auth Considerations

brain-app already authenticates to ibeco.me via WebSocket (token flow). For REST calls to ibeco.me practice/memorize endpoints, we need to either:

1. **Reuse the WS token** — if the token is a JWT or session token that also works for REST
2. **Add REST auth headers** — `Authorization: Bearer {token}` on practice/memorize API calls
3. **Proxy through WS** — send practice/memorize requests through the relay (new message types)

Option 1 is simplest if the auth token works for both. Option 3 is most consistent with the relay architecture. Decision needed during implementation.

---

## Files to Create/Modify

- `scripts/brain-app/lib/screens/today_screen.dart` — **New** main Today screen
- `scripts/brain-app/lib/screens/home_screen.dart` — Add Today to bottom nav
- `scripts/brain-app/lib/services/becoming_api.dart` — **New** REST client for ibeco.me practice/memorize endpoints
- `scripts/brain-app/lib/widgets/practice_tile.dart` — **New** practice completion tile
- `scripts/brain-app/lib/widgets/memorize_card.dart` — **New** inline card review widget

---

## Effort: Medium-Large (3-4 sessions across phases)

Phase 1 alone (practices) is a solid 1-session deliverable.
