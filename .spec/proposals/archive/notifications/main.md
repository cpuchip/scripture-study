# Desktop Notifications for ibeco.me

**Binding problem:** ibeco.me can't reach users unless they open it. Practices go untracked because the app relies on the user remembering to visit. Desktop notifications provide the "gentle tap on the shoulder" — reminding users when practices are due without requiring the site to be open.

**Created:** 2026-03-17
**Updated:** 2026-03-18
**Research:** [.spec/scratch/notifications/main.md](../../scratch/notifications/main.md)

---

## Phase 1 Status: SHIPPED (March 17-18, 2026)

What Phase 1 delivered:

| Component | Status |
|-----------|--------|
| Global `notifications_enabled` toggle in Settings | Done |
| `push_subscriptions` table (SQLite + PostgreSQL) | Done |
| `notification_log` table (dedup, 7-day cleanup) | Done |
| `user_settings` table (`notifications_enabled` only) | Done |
| Service worker (`sw.js`) + PWA manifest | Done |
| VAPID key generation + env var loading | Done |
| Scheduler goroutine (1-min tick, due practice check) | Done |
| Test notification button in Settings | Done |
| 410 Gone subscription cleanup | Done |
| Notification collapsing (multiple due → one notification) | Done |
| Subscribe/unsubscribe endpoints | Done |
| FK bug fix: `EnsureDefaultUser()` before `SeedPrompts()` | Done |

What Phase 1 did NOT build (deferred to Phase 2):

| Component | Notes |
|-----------|-------|
| `notify_practices_by_default` setting | Spec called for it in Phase 1 data model, not built |
| `quiet_hours` / `max_per_hour` | Spec listed in Phase 1 config model, not built |
| `default_timing` setting | Not built |
| Per-practice `notify` field in config JSON | Not built |
| Per-practice notification UI | Phase 2 scope |

**Current behavior:** When a user enables notifications, ALL due practices trigger notifications. There's no way to control which practices send notifications or when. This is the right MVP — notifications work — but Phase 2 needs to add the controls.

---

## 1. Problem Statement

The scheduling engine in ibeco.me already knows exactly when practices are due — interval, daily_slots, weekly, monthly, one-time. But that knowledge is trapped inside the server. The user has to open the website to discover what's due. For a practice-tracking app, that's backwards — the app should come to the user.

Web Push API makes this possible without a native app install. A service worker registered in the browser can receive push messages and show desktop notifications even when ibeco.me is completely closed.

### Success Criteria

1. User enables notifications in settings → browser permission prompt → subscription stored
2. When a practice becomes due, a desktop notification appears (Windows/Mac/Linux)
3. Clicking the notification opens ibeco.me to the relevant view
4. User can configure timing per practice (at time, 10 min before, 1 day before, etc.)
5. Quiet hours prevent notifications during sleep
6. Multiple simultaneous due practices collapse into a single summary notification
7. Works in Chrome, Edge, Firefox on all desktop OSes. Safari macOS 13+ as bonus.

---

## 2. Technical Approach

### How Web Push Works (No External Services Required)

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  ibeco.me   │────▶│  Browser Push     │────▶│  Service Worker │
│  Go backend │     │  Service (Google/ │     │  (in browser)   │
│             │     │  Mozilla/Apple)   │     │                 │
└─────────────┘     └──────────────────┘     └─────────────────┘
   HTTP POST           Routes message          Shows notification
   (encrypted)         to right browser          even if tab closed
```

**Dependencies:**
- Go: `github.com/SherClockHolmes/webpush-go` v1.4.0 — handles VAPID signing + payload encryption (see Library Assessment below)
- Frontend: Notification API + Push API (built into browsers, no npm package needed)
- One-time: Generate VAPID key pair (stored in server config/env)

### Library Assessment: webpush-go

| Criterion | Status |
|-----------|--------|
| Stars / Contributors | 415 stars, 82 forks, 20 contributors |
| Latest release | v1.4.0 (Jan 2, 2025) |
| Last push | Feb 2, 2026 |
| License | MIT |
| Dependencies | `golang-jwt/jwt/v5`, `golang.org/x/crypto` — both Go ecosystem staples |
| Security advisories | None published |
| Go Report Card | Clean |
| Known issues | Example repo has a minor build issue with JWT v5 generics — the library itself works fine |

**Alternatives considered:**
- `gootsolution/pushbell` — 2 stars, 1 contributor, uses fasthttp (extra dependency). Not mature enough.
- Roll our own — Web Push encryption (ECDH + HKDF + AES-GCM) is fiddly. Not worth reimplementing.

**Verdict:** webpush-go is the standard Go library for this. Well-maintained, minimal dependencies, no security issues. Safe to use.

---

## 3. Phased Delivery

### Phase 1: Foundation — Global Notifications (SHIPPED)

See "Phase 1 Status" section above for what was delivered.

**Original spec included a three-tier configuration model.** Phase 1 only implemented the global toggle. The full config model is documented below for reference — Phase 2 builds on it.

#### Configuration Model (Three-Tier)

Notifications use a layered opt-in design: global toggle → default-for-new-practices flag → per-practice override.

```
┌─────────────────────────────────────────────────────────┐
│ Settings (user_settings)                                │
│                                                         │
│  notifications_enabled: false  ← global kill switch     │
│  notify_practices_by_default: false  ← new practices    │
│  quiet_hours_start: "22:00"                              │
│  quiet_hours_end: "07:00"                                │
│  default_timing: "at_time"                               │
└─────────────────────────────────────────────────────────┘
         │
         ▼  (only shown if notifications_enabled = true)
┌─────────────────────────────────────────────────────────┐
│ Per-Practice (notification_config in practice JSON)     │
│                                                         │
│  notify: false          ← disabled by default           │
│  timing: "at_time"      ← inherits from default_timing  │
└─────────────────────────────────────────────────────────┘
```

**Behavior rules:**

1. `notifications_enabled = false` → nothing happens. No push, no UI for per-practice config.
2. `notifications_enabled = true` → settings shows per-practice notification options.
3. `notify_practices_by_default = false` (default) → new practices created with `notify: false`. User enables per-practice manually.
4. `notify_practices_by_default = true` → new practices created with `notify: true`.
5. **Retroactive toggle:** When `notify_practices_by_default` is flipped from false → true, all existing practices that have a due date/schedule AND `notify` is still `false` get updated to `notify: true`. A toast confirms: "Enabled notifications for N practices with schedules." This respects Michael's instinct — if you're turning on default notifications, you probably want them for what's already there too. Only applies to practices that are schedulable (have a schedule config or due date). Non-scheduled practices (pure trackers with no time component) are left alone.
6. Flipping `notify_practices_by_default` from true → false does NOT retroactively disable. That would be destructive — the user may have intentionally enabled specific ones.
7. Per-practice `notify: true` with no `timing` → inherits `default_timing` from user settings.

#### Scenario Walkthrough

**Scenario A: Notifications ON, Default OFF** (`notifications_enabled: true`, `notify_practices_by_default: false`)

The user wants control. No practice notifies unless they explicitly enable it.

1. User turns on notifications in Settings → browser asks permission → subscribed
2. No notifications fire yet — no practices have `notify: true`
3. User goes to their practices, sees bell icons next to each one (all off)
4. User taps the bell on "Scripture Reading" and "Exercise" → those two practices now send notifications when due
5. User creates a new practice "Journaling" → it starts with `notify: false` (user must tap the bell if they want it)
6. Only "Scripture Reading" and "Exercise" trigger notifications. Everything else is silent.

This is the non-spammy path. User dials in exactly what they want.

**Scenario B: Notifications ON, Default ON** (`notifications_enabled: true`, `notify_practices_by_default: true`)

The user wants everything to notify unless they opt out.

1. User turns on notifications, then flips `notify_practices_by_default` to ON
2. Retroactive update runs: all existing scheduled practices get `notify: true`. Toast: "Enabled notifications for 8 practices."
3. Every due practice fires a notification
4. User creates a new practice "Meal Prep" → it starts with `notify: true` automatically
5. If a practice is noisy, user taps the bell to turn it OFF for that one practice
6. Later, if user flips default back to OFF → existing practices keep their current setting (no retroactive disable). Only new practices start with `notify: false`.

This is the "notify me for everything" path. Maximum coverage, opt-out individual ones.

#### Phase 2 Migration

Simple: all existing practices get `notify: false` (the column default). No retroactive enabling. Only one user had notifications enabled during Phase 1's brief window. After Phase 2 deploys, the user enables notifications on the specific practices they care about via the bell icons.

**Phase 1 shipped only the global toggle and at_time notifications.** The data model additions listed above (per-practice fields, quiet hours, etc.) are built in Phase 2.

**Phase 1 Implementation Details (reference):**

Backend: VAPID keys in env vars. `push_subscriptions` table (id, user_id, endpoint, keys_p256dh, keys_auth, user_agent, created_at). Endpoints: `POST /api/push/subscribe`, `DELETE /api/push/unsubscribe`, `GET /api/push/vapid-key`, `POST /api/push/test`, `GET /api/push/settings`, `PUT /api/push/settings`. Scheduler goroutine ticks every minute, checks due practices via `DuePracticesForNotification()`, collapses multiple into summary, sends via `webpush.SendNotification()`, cleans up 410s. `notification_log` prevents duplicate sends.

Frontend: `public/sw.js` handles push + click events. `public/manifest.json` for PWA. SW registered in `main.ts`. Settings toggle uses `useNotifications()` composable for subscribe/unsubscribe lifecycle.

---

### Phase 2: Per-Practice Configuration (1 session)

**Delivers:** Users can control WHICH practices send notifications and WHEN. Also adds quiet hours to prevent 3am reminders.

**Current state (Phase 1):** All due practices notify. No per-practice control. No quiet hours. `user_settings` table only has `notifications_enabled`. Practice config JSON has no notification fields.

#### 2a. Data Model Changes

Phase 1 deferred these schema additions. Phase 2 must add them.

**user_settings table — new columns:**

| Column | Type | Default | Purpose |
|--------|------|---------|---------|
| `notify_practices_by_default` | boolean | false | New practices auto-get `notify: true` |
| `quiet_hours_start` | text (HH:MM) | NULL | No notifications after this time |
| `quiet_hours_end` | text (HH:MM) | NULL | No notifications before this time |
| `default_timing` | text | "at_time" | Default timing for new practice notifications |

Requires both SQLite migration (in `runSQLiteMigrations()`) and PostgreSQL goose migration.

**Practice config JSON — new fields:**

```json
{
  "notify": false,
  "timing": "at_time"
}
```

Added to the practice's existing config JSON blob. No new table needed. `notify` defaults to `false` (or `true` if `notify_practices_by_default` is on when created). `timing` inherits from `default_timing` if not set.

**Phase 2 migration:** All existing practices default to `notify: false`. User enables the ones they care about via bell icons. No retroactive migration needed.

#### 2b. Backend Changes

| Component | Detail |
|-----------|--------|
| `GET/PUT /api/push/settings` | Extend to include all new user_settings fields |
| `UserSettings` struct in `push.go` | Add `NotifyByDefault`, `QuietHoursStart`, `QuietHoursEnd`, `DefaultTiming` |
| `GetUserSettings` / `SetUserSettings` | Read/write all columns |
| Practice CRUD | When creating a practice, set `notify` based on `notify_practices_by_default`. No change to update — practice config JSON already round-trips. |
| Retroactive toggle | When `notify_practices_by_default` goes false→true: update all existing practices that have a schedule AND `notify` is still `false` to `notify: true`. Return count of updated practices. |
| `DuePracticesForNotification` | Filter on per-practice `notify` field — only send for practices where `notify: true` |
| Quiet hours enforcement | Scheduler checks user's `quiet_hours_start`/`quiet_hours_end` before sending. If current time is within quiet hours, skip entirely. |

**Retroactive toggle SQL (PostgreSQL):**
```sql
UPDATE practices 
SET config = jsonb_set(COALESCE(config, '{}'), '{notify}', 'true')
WHERE user_id = $1 
  AND schedule IS NOT NULL
  AND (config->>'notify' IS NULL OR config->>'notify' = 'false')
```

SQLite equivalent uses `json_set()`.

**Behavior rules (unchanged from original spec):**
1. `notify_practices_by_default` false→true: retroactively enables for scheduled practices. Toast: "Enabled notifications for N practices."
2. `notify_practices_by_default` true→false: does NOT retroactively disable. User may have intentionally enabled specific ones.
3. Per-practice `notify: true` with no `timing`: inherits `default_timing` from user settings.

#### 2c. Frontend Changes

| Component | Detail |
|-----------|--------|
| Per-practice bell icon | In practice list/detail: toggles `notify` in practice config JSON. Only visible when `notifications_enabled = true`. |
| `notify_practices_by_default` toggle | In SettingsView under notifications section. When toggled on, calls API which returns count → toast shows result. |
| Quiet hours config | In SettingsView: start/end time pickers (HH:MM select or input). |
| Default timing picker | In SettingsView: dropdown for the default timing option. |
| Per-practice timing | Optional override on practice detail: timing dropdown. Only shown if per-practice `notify` is true. |

#### 2d. Timing Options

| Value | Description |
|-------|-------------|
| `at_time` | When the practice is due (default) |
| `10_min_before` | 10 minutes before |
| `30_min_before` | 30 minutes before |
| `1_hour_before` | 1 hour before |
| `1_day_before` | Day before (for weekly/monthly) |

Note: Dropped `custom_N` from original spec. Premature. Can add later if needed.

#### 2e. Verify

- Per-practice bell icon only visible when global notifications are enabled
- Toggle `notify_practices_by_default` on → toast shows N practices updated
- Create a new practice while default is on → it has `notify: true`
- Disable notifications for one practice → no notification for it, others still fire
- Set quiet hours → no notifications during that window
- Set a practice to notify "10 min before" → notification arrives 10 min early

---

### Phase 3: Rich Notifications + Actions (1 session)

**Delivers:** Notifications with action buttons. Snooze. Mark-as-done from the notification itself.

| Feature | Detail |
|---------|--------|
| Action buttons | "Done" and "Snooze 15m" on each notification |
| "Done" action | Service worker POSTs to `/api/practices/{id}/complete` |
| "Snooze" action | Reschedules notification for 15 minutes later |
| Rich payload | Practice name, category emoji, streak info |
| Collapse tag | Same practice = replaces existing notification (no pile-up) |

**Verify:**
- Click "Done" on notification → practice logged without opening the app
- Click "Snooze" → notification reappears in 15 min

---

### Phase 4: Brain-App Integration (separate, deferred)

**Delivers:** Same notifications on Android via brain-app.

**Approach:** Local notifications scheduled on-device during daily sync. The app already fetches daily data — extend to schedule `flutter_local_notifications` alarms for each due practice.

**Why deferred:** Web push covers desktop (the request). Brain-app is a separate codebase and the Flutter notification plugin setup is its own task.

---

## 4. Constraints

- **No Firebase / external push services** — use Web Push protocol directly. Go backend sends to browser push endpoints.
- **No polling from frontend** — the service worker receives push messages passively
- **VAPID keys stored in env** — never in code or git
- **Subscriptions are per-browser, not per-user** — a user on 2 browsers gets 2 subscriptions. Handle gracefully.
- **Subscription cleanup** — browser push services return 410 Gone when a subscription expires. The backend must delete stale subscriptions on 410.
- **HTTPS required** — service workers only work over HTTPS. ibeco.me already uses HTTPS via Dokploy.

---

## 5. Costs & Risks

| Cost | Detail |
|------|--------|
| **Complexity** | Service worker lifecycle (install, activate, update) has edge cases. Version management matters. |
| **Browser variance** | Safari's Web Push has quirks (needs PWA manifest, no action buttons in older versions) |
| **Subscription churn** | Browsers revoke subscriptions periodically. Need retry + cleanup logic. |
| **Notification fatigue** | Risk of becoming annoying. Mitigated by quiet hours, rate limits, collapse. |
| **Scheduler overhead** | Goroutine ticking every minute. For 1-2 users this is trivial. At scale, needs optimization. |

| Risk | Mitigation |
|------|------------|
| User denies permission | Graceful degradation. Clear messaging about what notifications do. |
| Push service outage | Notifications are best-effort. Not critical path. |
| Stale subscriptions pile up | 410 cleanup on every failed send |

---

## 6. Creation Cycle Review

| Step | This Feature |
|------|-------------|
| **Intent** | The app should serve the user, not the other way around. Reach out when it's time. |
| **Covenant** | Notifications respect the user's attention — quiet hours, rate limits, easy disable. |
| **Stewardship** | ibeco.me backend owns scheduling. Browser owns delivery. User owns permission. |
| **Spiritual Creation** | This spec. The scheduling engine already knows what's due — this is the delivery layer. |
| **Line Upon Line** | Phase 1: global toggle. Phase 2: per-practice config. Phase 3: actions. Phase 4: mobile. |
| **Physical Creation** | dev agent builds backend + service worker. |
| **Review** | Each phase has verify criteria. Test across Chrome + Firefox + Edge minimum. |
| **Atonement** | If notifications annoy: lower defaults, add "gentle" mode (1 summary/day). |
| **Sabbath** | Quiet hours are literally a sabbath for notifications. |
| **Consecration** | Your daughter uses this too. It's not just Michael's tool. |
| **Zion** | A practice tool that gently reminds instead of demanding — that's the right relationship between tool and person. |

---

## 7. Recommendation

**Build.** Phase 1 is achievable in one session and delivers the core ask: desktop notifications when practices are due. The scheduling engine already exists — this is the delivery pipe.

Start with Phase 1. If your daughter enables notifications on her browser and starts getting gentle reminders when her practices are due, the feature has proven itself. Per-practice config (Phase 2) follows naturally once the infrastructure is in place.
