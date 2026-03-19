# Scratch — Desktop Notifications for ibeco.me + brain-app

**Binding problem:** Your daughter wants ibeco.me to tap her on the shoulder — desktop notifications for due practices, scheduled tasks, and reminders. Currently there's no notification infrastructure at all. The app can't reach you unless you open it.

**Created:** 2026-03-17

---

## Phase 2: Research & Inventory

### Current State

**ibeco.me (web):**
- Vue 3 + Vite SPA embedded in Go binary
- No service worker
- No PWA manifest
- No notification code
- No push subscription storage
- No user settings/preferences table
- Auth: JWT sessions + API tokens (cookie-based)
- Practices already have scheduling engine (`internal/db/schedule.go`) — knows what's due and when
- TODO.md lists "PWA support" as "Not started"
- Original docs (`06_becoming-app.md`) punted notifications to "Phase 6"

**brain-app (Flutter/Android):**
- Separate repo (cpuchip/brain-app)
- Connects to ibeco.me via WebSocket relay
- No notification plugins currently
- Could use `flutter_local_notifications` for scheduled local notifications
- Could use FCM for server-triggered push

**brain.exe (Go agent):**
- Runs locally on Michael's machine
- Has proactive surfacing plan (Plan 17) — due actions, stale people, stalled subtasks
- Could be a notification *source* but not a delivery mechanism for other users

### Web Push API — How It Works

1. **Service Worker** — browser background script. Outlives the tab. Required for push.
2. **VAPID keys** — public/private key pair generated once. Identifies the app to push services.
3. **PushSubscription** — per-browser object containing:
   - `endpoint` — URL at browser vendor's push service (Google, Mozilla, Apple)
   - `keys.p256dh` — public encryption key
   - `keys.auth` — authentication secret
4. **Backend stores subscriptions** — `push_subscriptions` table, linked to user_id
5. **Backend sends push** — HTTP POST to the endpoint URL, encrypted with the subscription keys
6. **Service worker receives** — `push` event fires, calls `self.registration.showNotification()`
7. **User clicks notification** — `notificationclick` event fires, can open/focus the app

**Go library:** `github.com/SherClockHolmes/webpush-go` v1.4.0 — the standard Go Web Push library.

**Library assessment (Exa search, 2026-03-17):**
- 415 stars, 82 forks, 20 contributors, MIT license
- Latest release: v1.4.0 (Jan 2, 2025). Last push: Feb 2, 2026. Actively maintained.
- Dependencies: `golang-jwt/jwt/v5` + `golang.org/x/crypto` — both well-maintained Go ecosystem staples
- No published security advisories
- Minor issue: example repo has a build issue with JWT v5 generics syntax (#72), but the library itself compiles and works fine. PR #84 addresses it.
- Go Report Card: Clean

**Alternatives evaluated:**
- `gootsolution/pushbell` — 2 stars, 1 contributor, uses fasthttp (adds an extra dependency). Too immature.
- `crow-misia/go-push-receiver` — This is for *receiving* FCM push, not sending. Wrong direction.
- Roll your own — Web Push encryption involves ECDH key agreement, HKDF key derivation, AES-128-GCM encryption, and VAPID JWT signing. Not worth reimplementing.
- `draphy/pushforge` — TypeScript/Node.js only. Not applicable for Go backend.

**Verdict:** webpush-go is the right choice. It's the only serious Go implementation, well-maintained, minimal attack surface.

**No external service required.** No Firebase, no OneSignal, no Pusher. Just your Go backend + the browser's built-in push service.

### Gorush Evaluation (2026-03-17)

Michael asked about running [Gorush](https://github.com/appleboy/gorush) as a Dokploy sidecar.

**What Gorush is:** A push notification micro server (8,700 stars, 882 forks, 60 contributors, MIT, v1.21.0 released Mar 8, 2026). Actively maintained by appleboy. Supports APNS (iOS), FCM/GCM (Android), and Huawei push.

**What Gorush is NOT:** A Web Push server. It does not support the Web Push protocol (RFC 8030), VAPID (RFC 8292), or browser push subscriptions. Its topics are literally "android, apns, gcm, ios, ios-notification."

**Why it doesn't fit:** Desktop notifications require Web Push — the protocol where the server encrypts a payload and POSTs to the browser vendor's push service (Google FCM endpoint for Chrome, Mozilla Autopush for Firefox, WNS for Edge). Gorush speaks a different protocol entirely (direct APNS/FCM device token push for native mobile apps).

**When Gorush would matter:** If brain-app (Flutter/Android) needs server-triggered push via FCM. But that's Phase 4, and even then a direct FCM HTTP v1 API call from the Go backend would suffice for 1-2 users. A separate notification server is designed for high-throughput mobile push at scale — not this use case.

**Also found:** `mpizenberg/go-notify-server` (Mar 2026, 0 stars, 1 contributor, LLM-generated) — a minimal self-hosted Web Push server. Its README confirms the architectural insight: "The 'application server' role in Web Push is intentionally thin and application-specific — the ecosystem provides libraries for the hard cryptographic parts (webpush-go), not turnkey servers."

**Decision:** Stay with webpush-go embedded in ibeco.me. No sidecar needed.

### Browser Support

| Browser | Desktop | Mobile | Notes |
|---------|---------|--------|-------|
| Chrome | Full | Full | Uses FCM under the hood |
| Edge | Full | Full | Uses WNS under the hood |
| Firefox | Full | Full | Uses Mozilla push service |
| Safari | macOS 13+ | iOS 16.4+ PWA only | Standard Web Push since Safari 16.1 |

### Configuration Model — What Michael Described

> "a full configuration from just the time of, to 10 minutes before, to 1 day before to repeated notifications, we shouldn't spam the user, but give them broad configuration capabilities"

This maps to a three-tier opt-in model:

**Tier 1 — Global settings (`user_settings` table or JSON column):**
```json
{
  "notifications_enabled": false,
  "notify_practices_by_default": false,
  "default_timing": ["at_time"],
  "quiet_hours": { "start": "22:00", "end": "07:00" },
  "max_per_hour": 5
}
```

**Tier 2 — Per-practice config (in practice JSON config field):**
```json
{
  "notify": false,
  "timing": ["10_min_before", "at_time"]
}
```

**Behavior rules:**
1. `notifications_enabled = false` → no push sent, no per-practice UI shown
2. `notifications_enabled = true` → per-practice notification bell icons appear
3. Individual practices default to `notify: false` — user enables manually
4. `notify_practices_by_default = true` → NEW practices get `notify: true`
5. **Retroactive:** Flipping `notify_practices_by_default` false→true also enables notifications for all EXISTING practices that have schedules/due dates. Only scheduled practices — pure trackers with no time component are skipped.
6. Flipping true→false is NOT retroactive (would be destructive)
7. Per-practice `timing` is optional — inherits `default_timing` from user settings if empty
```

Timing options:
- `at_time` — when the practice is due
- `10_min_before` — 10 minutes before
- `30_min_before` — 30 minutes before
- `1_hour_before` — 1 hour before
- `1_day_before` — day before (good for weekly/monthly practices)
- `custom` — user-specified minutes before

Anti-spam:
- **Quiet hours** — global and per-practice
- **Max per hour** — configurable rate limit
- **Collapse** — if 5 practices are due at the same time, send ONE notification listing them
- **Snooze** — "Remind me in 15 min" action on the notification itself

### What This Requires (New Infrastructure)

**Database:**
1. `user_settings` table or `settings` JSON column on `users` — store notification preferences
2. `push_subscriptions` table — `id`, `user_id`, `endpoint`, `keys_p256dh`, `keys_auth`, `user_agent`, `created_at`
3. `notification_log` table — prevent duplicate sends, track delivery

**Backend:**
1. VAPID key pair generation (one-time setup)
2. `/api/push/subscribe` endpoint — stores subscription
3. `/api/push/unsubscribe` endpoint — removes subscription
4. `/api/settings` endpoint — CRUD for notification preferences
5. **Notification scheduler** — goroutine that checks for due practices and sends push
6. Web Push sending via `webpush-go`

**Frontend:**
1. Service worker (`sw.js`) — handles push events, shows notifications
2. PWA manifest (`manifest.json`) — required for service worker registration
3. Permission flow in settings — "Enable Notifications" button → browser permission prompt
4. Per-practice notification config UI in settings
5. Notification click handler — opens the relevant practice/view

**Scope estimate:** This is a medium-sized feature. The scheduling engine already exists. The hard parts are:
- Service worker lifecycle (install, activate, update)
- Push subscription management (browsers revoke subscriptions periodically)
- The scheduler goroutine (needs to be efficient, not poll every second)
- Testing across browsers

### Prior Art Within Project

- `internal/db/schedule.go` — already computes `IsDue`, `NextDue`, `DaysOverdue`, `SlotsDue`
- `internal/brain/hub.go` — WebSocket relay, message queuing pattern for offline delivery
- `07-ux-improvements.md` § 5.4 — planned toast notification system (in-app, not push)
- `TODO.md` — PWA support listed as "not started"
- `06_becoming-app.md` line 726 — "Notifications: Later"
- `06_becoming-app.md` line 608 — "PWA support (service worker, installable)"

### Brain-App Notification Path

For brain-app (Flutter), the options are:

1. **Local notifications** — `flutter_local_notifications` plugin, scheduled on-device. Would need the app to sync due practices and set alarms locally. Works offline. Most reliable.

2. **Server push via FCM** — ibeco.me backend sends to both web push endpoints AND FCM tokens. More complex (Firebase project setup), but uses the same scheduler.

3. **WebSocket-based** — Brain-app already has a WebSocket connection to ibeco.me. Could receive a `notification` message type and show a local notification. Only works when the app is running (or has a foreground service).

**Recommendation:** Start with option 1 (local notifications) for brain-app. The app already syncs daily data — it can schedule local notifications for due practices during that sync. No Firebase dependency.

---

## Phase 3a: Critical Analysis

1. **Is this the RIGHT thing to build?** Yes. Notifications are the "gentle tap on the shoulder" that makes a practice tool useful. Without them, the user has to remember to open the app. That's the opposite of the tool serving the user.

2. **Does this solve the binding problem?** Directly. The daughter wants desktop notifications. This delivers desktop notifications.

3. **Simplest useful version?** Global on/off + notifications for all due practices at the scheduled time. No per-practice config in Phase 1. That alone would be valuable.

4. **What gets WORSE?** Notification fatigue is real. If there are 20 daily practices, 20 notifications is spam. The collapse/max-per-hour design addresses this, but it needs to be right from the start.

5. **Does this duplicate something?** No. Nothing in the project does notifications.

6. **Is this the right time?** It's a natural extension of the scheduled tasks feature that already exists. The scheduling engine is the hard part and it's done.

7. **Mosiah 4:27 check:** This is a bounded feature. Phase 1 is achievable in 1-2 sessions. It doesn't create a new project — it extends the existing one.

8. **Creation Cycle:** This is "Physical Creation" — the spiritual creation (schedule engine) is done; this is making it reach the user.

**Verdict: Proceed.** This is a clearly useful feature with a defined scope and existing infrastructure to build on.

---

*Proposal at `.spec/proposals/notifications/main.md`.*
