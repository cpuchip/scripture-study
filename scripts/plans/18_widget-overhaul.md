# Plan 18: Widget Overhaul (absorbs Plan 16 Phase 4)

*Created: March 8, 2026*
*Priority: Near-term*
*Absorbs: Plan 16 Phase 4 (Widget/Android) — the original 3-liner is now this full plan*
*Depends on: Plan 16 (Today Screen — Phases 1-3 done), practice log API (just fixed)*
*Connects to: Plan 17 (Proactive Surfacing)*

---

## Problem

Current widget system has three sizes — 4x2 standard (task list + checkboxes), 2x2 compact (just add/mic buttons), 2x1 mini (brain icon + add/mic). Issues:

1. **Standard widget checkboxes are cosmetic** — they fire `MARK_DONE` intents and the data updates, but the widget never re-renders with a checked visual. The checkbox icon stays `checkbox_off_background` forever. Looks fake.
2. **2x2 compact is wasted space** — just a title and two buttons. No glanceable info.
3. **No practice or memorize visibility** — you have to open the app to see what's due. The whole point of ibeco.me practices is daily consistency, and you can't see them without context-switching.

## Inspiration: Microsoft TODO Widget

The Microsoft TODO widget (see screenshots) shows how to do this well:
- **Scrollable list** inside the widget (`ListView` / `StackView` RemoteViews adapter)
- **Category dropdown** at the top — tapping it spawns a **separate activity** that shows the full category list, then closes when you pick one. The widget updates to show that category's items.
- **Tap-to-complete** checkboxes that actually toggle state

We already have the transparent-activity pattern from `QuickAddActivity`. A `WidgetFilterActivity` would follow the same model — Flutter activity with its own Dart entrypoint, launched from a widget button, closes on selection, triggers widget update.

---

## New Widgets

### Widget 1: Practices (2x2 or 4x2 resizable)

**What it shows:**
- Header: Category name (tappable → opens filter activity) + progress (e.g. "3/5")
- List of practices in that category, each with set buttons (like the new TodayScreen tiles)
- Tap a set button → logs via `POST /api/logs` (same as in-app)
- Completed sets show checkmark, incomplete show number

**Category filter flow:**
1. User taps the category dropdown in the widget header
2. Launches `WidgetFilterActivity` (transparent Flutter activity, like QuickAddActivity)
3. Shows list of categories pulled from practice data: "All", "PT", "Spiritual", etc.
4. User taps a category → saves to SharedPreferences → triggers `HomeWidget.updateWidget()` → activity closes
5. Widget re-renders showing only practices in that category

**Data source:** `GET /api/daily/{date}` → filter by category → push to HomeWidget shared prefs

**Sizes:**
- 2x2: 2-3 practices visible, compact set buttons  
- 4x2+: 4-6 practices, full set button row

### Widget 2: Memorize (2x2)

**What it shows:**
- One scripture card at a time — reference + first line or prompt text
- Left/right arrow buttons to flip through due cards
- Current position indicator ("2 of 5")
- Tap the card area → opens app to Today tab for full review

**The key insight:** This is for **passive exposure**, not full review. Seeing "D&C 93:36 — The glory of God is intelligence" on your home screen 20 times a day builds familiarity without any conscious effort. The review (with quality rating) happens in-app.

**Filter flow:** Same pattern — tapping the header opens `WidgetFilterActivity` in memorize mode, lets you filter by category or show all due cards.

**Data source:** `GET /api/memorize/due/{date}` → push card list to HomeWidget shared prefs. Arrow buttons cycle `current_card_index` in prefs and update widget.

**Sizes:**
- 2x2: Fixed. Card text, arrows, position indicator.
- Could stretch to 4x2 for longer scripture text.

### Widget 3: Fix Existing Standard Widget (4x2)

**Checkbox fix:**
- Track done state in shared prefs: `entry_{i}_done` (boolean)
- In `buildStandardView()`, swap checkbox icon based on state: `checkbox_on_background` vs `checkbox_off_background`
- When `MARK_DONE` fires, the app marks done + writes updated prefs + calls `HomeWidget.updateWidget()` to re-render

**Optional enhancement:** Mix brain actions with top practices/memorize for a unified "today" widget.

---

## Architecture

### WidgetFilterActivity (New)

Same pattern as `QuickAddActivity`:

```
WidgetFilterActivity.kt
├── getBackgroundMode() → transparent
├── getDartEntrypointFunctionName() → "widgetFilterMain"
└── configureFlutterEngine() → passes filter_type (practices|memorize) + widget_id via MethodChannel
```

Dart side (`widget_filter_main.dart`):
- Fetches categories from BecomingApi (or reads from cached shared prefs)
- Shows a clean list UI: category names with current selection highlighted
- On tap: save to `HomeWidget.saveWidgetData('practice_filter', category)` → `HomeWidget.updateWidget()` → `SystemNavigator.pop()`

### PracticeWidgetProvider (New Kotlin)

New receiver in AndroidManifest. Reads practice data from SharedPreferences (populated by Flutter side), renders set buttons as individual `ImageButton`s with click intents.

**Set logging from widget:** Tap set button → PendingIntent launches a lightweight `WidgetActionReceiver` (BroadcastReceiver, not Activity) that:
1. Reads token from shared prefs
2. Fires `POST /api/logs` directly via `HttpURLConnection` in a coroutine (or uses WorkManager for reliability)
3. Updates shared prefs with new completion count
4. Triggers widget re-render

Alternative: Launch invisible Flutter activity to do the API call (simpler, reuses existing auth, but slower to spin up).

### MemorizeWidgetProvider (New Kotlin)

Reads card list from SharedPreferences. Arrow buttons update `current_card_index` in prefs and call `updateWidget()` — no API call needed, just cycling through cached data.

### Data Refresh Strategy

Widget data needs to stay fresh. Options:
- **On app background** — when brain-app goes to background, push latest practice/memorize data to widget prefs (already partially done in `WidgetService.updateWidget()`)
- **WorkManager periodic** — schedule a 30-min background job that fetches fresh data and updates widgets. Needs stored auth token.
- **On widget update** — `updatePeriodMillis` triggers `onUpdate()`, but minimum is 30 min and it's unreliable

Start with on-app-background (simplest), add WorkManager later if freshness matters.

---

## Implementation Phases

### Phase 1: Fix Existing Widget Checkboxes
1. Add `entry_{i}_done` to WidgetService shared prefs
2. Update `buildStandardView()` to swap checkbox drawable based on done state
3. After `MARK_DONE` intent fires, update prefs + re-render widget
**Effort: Small (1 hour)**

### Phase 2: Practices Widget
1. Create `PracticeWidgetProvider.kt` + XML layout + widget_info
2. Create `WidgetFilterActivity.kt` + `widgetFilterMain` Dart entrypoint
3. Extend `WidgetService` to push practice data to shared prefs
4. Wire set-logging from widget (API call on tap)
5. Register in AndroidManifest
**Effort: Medium (1-2 sessions)**

### Phase 3: Memorize Widget
1. Create `MemorizeWidgetProvider.kt` + XML layout + widget_info
2. Push due card data from WidgetService
3. Arrow button cycling (shared prefs index update + re-render)
4. Card tap → deep link to Today tab
5. Filter activity support (reuse from Phase 2)
**Effort: Medium (1 session)**

### Phase 4: Background Refresh (Optional)
1. WorkManager periodic task to fetch practice/memorize data
2. Store auth token securely for background use
3. Update all widget providers on fresh data
**Effort: Small-Medium**

---

## Files to Create

| File | Purpose |
|------|---------|
| `android/.../PracticeWidgetProvider.kt` | Practice widget Kotlin provider |
| `android/.../MemorizeWidgetProvider.kt` | Memorize widget Kotlin provider |
| `android/.../WidgetFilterActivity.kt` | Transparent category picker activity |
| `android/.../WidgetActionReceiver.kt` | BroadcastReceiver for set-logging from widget |
| `android/res/layout/practice_widget.xml` | Practice widget layout |
| `android/res/layout/memorize_widget.xml` | Memorize widget layout |
| `android/res/xml/practice_widget_info.xml` | Practice widget metadata |
| `android/res/xml/memorize_widget_info.xml` | Memorize widget metadata |
| `lib/widget_filter_main.dart` | Dart entrypoint for filter activity |
| `lib/services/widget_service.dart` | Extended with practice/memorize data push |

## Files to Modify

| File | Change |
|------|--------|
| `AndroidManifest.xml` | New receivers + WidgetFilterActivity |
| `BrainWidgetProvider.kt` | Checkbox state fix in buildStandardView |
| `widget_service.dart` | Push practice/memorize data + done state tracking |
| `home_screen.dart` | Push widget data on app background (didChangeAppLifecycleState) |

---

## Effort: Medium-Large (3-4 sessions across phases)

Phase 1 alone (checkbox fix) is a quick win. Phase 2 (practices) is the biggest lift. Phase 3 (memorize) reuses a lot of Phase 2 infrastructure.
