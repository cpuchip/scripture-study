# Active Context

*Last updated: 2026-03-21 (Phase 3a built) | WS1 Phase 3a: Agent Pool + Routing Table — SHIPPED*

---

## Current State

**WS1 Phase 2.5: Workspace-Aware Sessions — SHIPPED (Mar 21).** All code changes implemented and tested:
- `internal/config/workspace.go` (NEW) — Parses `.github/copilot-instructions.md`, agents, skills
- `internal/ai/agent.go` — Added `SkillDirectories`, `InfiniteSessions` to `AgentConfig`
- `internal/config/config.go` — MCP auto-discovery expanded 3→7 servers
- `cmd/brain/main.go` — `--agent` flag, `buildSystemMessage()`, workspace loading

**Validation test:** `brain exec --agent study` produced `study/only-begotten.md` (v2). Compared to the generic-prompt v1:
- **v1:** 103 lines, 3696 words, no scratch file, no phased workflow, no Webster, no skills
- **v2:** 169 lines, 4710 words, scratch file created, all 7 study phases followed, Webster 1828 used, skills loaded on-demand, context compaction worked
- All 5 SDK features confirmed working: SkillDirectories, InfiniteSessions, base instructions, 7 MCP servers, study agent prompt

### Priorities (Mar 21)
1. **Study** — Highest priority. "It keeps me in the spirit." Studies queued in brain-app.
2. **Agentic Foundation** — WS1 Phase 3 next (multi-agent routing).

### Key Decisions
All settled decisions are in [decisions.md](decisions.md). Key ones affecting current work:
- **NOCIX server pending.** Dedione refunded. 3TB, unmetered 1Gbps.
- **Cost:** 56% of 1500 premium requests used with 1/3 month remaining.

### Data Safety (shipped Mar 19)
- Phase 1: Dev agent checklist ✅
- Phase 2: DB constraints (migration 015) ✅
- Phase 3: Handler remediation (5 PUT handlers → read-modify-write) ✅
- Phase 4: Go tests ✅
- Phase 5: Audit log (migration 016) ✅
- Phase 6: Drop SQLite (CGO_ENABLED=0) ✅
- Production outage #1: Wrong CHECK values → fixed (c4c48a2)
- Production outage #2: Missing goose StatementBegin → fixed (1cd0775)
- Retrospective + checklist update → committed (7c87fb6)

---

## In Flight

### Squad Analysis (COMPLETE — Mar 19)
- Deep investigation of [bradygaster/squad](https://github.com/bradygaster/squad) multi-agent runtime
- Compared to our 11-step creation cycle and WS1 plans
- **Critical self-assessment added:** We practice ~28% of our own 11-step cycle. Theory/practice gap flagged.
- Proposal: [.spec/proposals/squad-learnings.md](../proposals/squad-learnings.md)
- Research: [.spec/scratch/squad-analysis/main.md](../scratch/squad-analysis/main.md)
- **6 adoption items identified** (resequenced with Phase 0: practice what we preach):
  - Phase 0: ~~Add intent.yaml to session-start, create decisions.md,~~ practice Sabbath ✅ (decisions.md + session-start done Mar 19)
  - A1: ~~Decisions file~~ ✅ (`.spec/memory/decisions.md` created Mar 19)
  - A2: Agent routing table (WS1 Phase 3) — NEXT
  - A3: Hook-based governance (WS1 Phase 3)
  - A4: Reviewer lockout with model escalation — bump model tier before swapping agent (WS1 Phase 3)
  - A5: Response tier / model selection (WS1 Phase 2)
  - A6: Cost tracking (WS1 Phase 2)
- **YouTubes pending:** ~~Two videos covering similar ideas to Squad — review before implementing Phase 2+~~ First video reviewed: [LO0Ws-l6brg](../../study/yt/LO0Ws-l6brg-4-ai-labs-same-system.md). Validates convergent pattern. Key takeaway: harness > intelligence, reduce complexity before adding. No more research gates — build Phase 0.
- **Decision:** Approved by Michael (Mar 19). Proceed with Phase 0 first.

### Agent Promotions (COMPLETE — Mar 19)
- `study-exp1` → promoted to `study` (phased writing, scratch files, critical analysis)
- `lesson-exp1` → promoted to `lesson` (phased prep, pedagogy framework)
- `yt-exp1` → promoted to `eval` (phased evaluation, charitable analysis)
- `plan-exp1` → new `plan` agent (creation cycle review)
- Old versions backed up as `.bak` files
- exp1 files remain as historical originals
- copilot-instructions.md agent table updated
- All handoff references updated to point to promoted names

### WS1 Phase 1: Copilot SDK + MCP Integration — DONE (Mar 20)
- `internal/ai/agent.go` — Agent struct with Ask/Reset/createSession, MCP server registration
- `internal/config/config.go` — MCPServerDef, AgentModel, auto-discovery of sibling MCP servers
- `internal/web/server.go` — POST /api/agent/ask, POST /api/agent/reset
- `internal/ai/client.go` — CopilotClient() getter
- `cmd/brain/main.go` — Agent creation wired into startup

### WS1 Phase 2: Agent as Spec Executor — DONE (Mar 20)
- **Test:** `brain exec test-spec.md` — add markdown_link to gospel-mcp GetResponse
- **Result:** Agent made 4/8 correct changes autonomously, then timed out (SDK internal timeout ~60s)
- **Human completed:** remaining 4 construction sites + fixed variable shadowing compile error
- **Key discoveries:**
  - Copilot SDK v0.1.32 has BUILT-IN tools: `view` (read files), `grep` (search), `edit` (modify files), `report_intent` (declare plan)
  - Our custom devTools (read_file, write_file, list_directory, search_text) were redundant — removed
  - SDK protocol updated v2→v3. Required updating SDK from v0.1.29→v0.1.32
  - `brain exec` CLI subcommand added for spec execution without starting full server
- **What shipped (gospel-mcp):** MarkdownLink field added to GetResponse, populated in all 8 construction sites
- **Timeout fixed:** `AskStreaming` method uses `Send` + `On` (event-driven) instead of `SendAndWait`. Streams output in real-time via `AssistantMessageDelta` events. Exec uses 10-minute timeout. Custom devTools removed; `tools.go` pending manual deletion.
- **Pending cleanup:** Delete `scripts/brain/internal/ai/tools.go` and `scripts/brain/test-spec.md`

### WS1 Phase 2.5: Workspace-Aware Sessions — DONE (Mar 21)
- `internal/config/workspace.go` — `LoadWorkspace()` parses copilot-instructions.md + agents + skills
- `internal/ai/agent.go` — SkillDirectories + InfiniteSessions support
- `internal/config/config.go` — 7 MCP servers auto-discovered (was 3)
- `cmd/brain/main.go` — `--agent` flag, `buildSystemMessage()`, 30min timeout
- **Validated:** study agent produced full phased study with scratch files, Webster, skills, compaction

### Overview Plan
- **Status:** All questions answered, decisions recorded.
- **WS1 Phase 1:** DONE. Phase 2 POC done. Phase 2.5 workspace-aware done. Phase 3 proposal written — awaiting review.

### WS1 Phase 3a: Agent Pool + Routing Table — SHIPPED (Mar 21)
- **8 files changed** (3 new, 5 modified). All tests pass.
- **New: `internal/ai/pool.go`** — `AgentPool` with `map[string]*Agent`, `sync.RWMutex`, lazy creation via `GetOrCreate(agentName, wc)`. Exported `BuildSystemMessage()` (extracted from main.go).
- **New: `internal/ai/router.go`** — `RouteMode` (none/suggest/auto), `DefaultRoutes` map (study→study, journal→journal, ideas→plan; projects/actions/people→none), `LookupRoute()`, `RenderPrompt()` with Go templates.
- **New: `internal/ai/pool_test.go`** — 6 tests covering routing lookup, prompt rendering, and system message building.
- **Modified: `internal/store/types.go`** — 4 new Entry fields: `AgentRoute`, `RouteStatus`, `AgentOutput`, `TokensUsed`.
- **Modified: `internal/store/db.go`** — Idempotent migration adding 4 columns. New methods: `SetAgentRoute()`, `UpdateRouteStatus()`, `SetAgentOutput()`. Updated `GetEntry()` and `ListCategory()` queries.
- **Modified: `internal/store/store.go`** — 3 delegation methods for routing.
- **Modified: `cmd/brain/main.go`** — Single `agent` replaced by `pool *ai.AgentPool`. Local `buildSystemMessage()` removed (now exported from pool.go).
- **Modified: `internal/web/server.go`** — Server uses `pool + wc`. New endpoints: `GET /api/agent/sessions`, `POST /api/agent/route` (background routing), `GET /api/agent/routable` (entries eligible for routing). `handleAgentAsk` accepts `agent` field for named agent sessions.
- **Next:** Phase 3b (governance hooks) or Phase 3c (auto-routing). But study is higher priority.

### WS1 Phase 3: Multi-Agent Routing — PROPOSAL APPROVED (Mar 21)
- Proposal: [.spec/proposals/brain-multi-agent/main.md](../proposals/brain-multi-agent/main.md)
- Research: [.spec/scratch/brain-multi-agent/main.md](../scratch/brain-multi-agent/main.md)
- **Three sub-phases proposed:**
  - **3a:** Agent pool + routing table — `map[string]*Agent` with lazy creation, category→agent mapping, "suggest" mode (human approves routing)
  - **3b:** Governance hooks — `OnPreToolUse` for file-write scoping + destructive op blocking, token budgets, audit logging
  - **3c:** Auto-routing + review queue — entries route automatically, human reviews output
- **Key decisions:**
  - Default routing mode is "suggest" not "auto" — respects gated autonomy decision
  - Pool with lazy instantiation (not single agent or subprocess)
  - Only study→study and journal→journal map cleanly; ideas→plan is plausible; others→none initially
  - Token budget concern: 56% premium requests used, governance needed before wide-open routing
- **Recommendation:** Build Phase 3a first (1 session). Defer 3b/3c until routing feels right in practice.

---

## In Flight

### Brain Ecosystem — Active Development (March 6-12, 2026)

The brain ecosystem has been the primary development focus for the past week. Three codebases, significant feature work completed.

#### brain-app (Flutter) — `scripts/brain-app/` (separate git repo: cpuchip/brain-app)
**Recent work (Mar 8-12):**
- **Practice widget (Plan 18 Phase 2)** — COMPLETE. Scrollable practice list widget with:
  - Per-instance category filtering (tap header to cycle, or `WidgetFilterActivity` flyover)
  - Set/rep button circles that log via background callback → API
  - Undo via DELETE /api/logs/latest
  - Progress counter in header (e.g. "3/5")
  - Refresh button (added Mar 12)
  - Due-only items at top, not-due items dimmed at bottom
- **Quick Add Practice from widget** — COMPLETE. `QuickAddPracticeActivity` launches transparent Flutter overlay with full `PracticeForm` widget. Had to move `@pragma('vm:entry-point')` function into `main.dart` (same pattern as working `quickAddMain`) to fix AOT tree-shaking.
- **Widget filter per-instance** — COMPLETE. `WidgetFilterActivity` → `widgetFilterMain` entrypoint. Each widget instance stores its own filter in `practice_filter_{widgetId}`.
- **PracticeForm shared widget** — COMPLETE. `lib/widgets/practice_form.dart`. Full form for all 5 practice types (habit, tracker, scheduled w/ interval/weekly/monthly/daily_slots/once, task, memorize). Used in both Today screen bottom sheet and widget quick-add.
- **Today screen daily_slots support** — COMPLETE (Mar 11). Named slot buttons (e.g., "morning", "bedtime") instead of numbered circles. Config parsing fixed (nested `schedule.type` not flattened). `_SlotButton` widget with strikethrough when done.
- **Widget daily_slots support** — COMPLETE (Mar 11-12). Kotlin `PracticeWidgetService` renders daily_slots with named circles and strikethrough subtitle via SpannableString. Slot name passed in URI query param for logging (`?slot=morning`).
- **Undo for daily_slots** — FIXED (Mar 12). `_undoPracticeSet` now accepts `{String? slotName}` and correctly adds slot back to `slotsDue`.
- **App→Widget sync** — ADDED (Mar 12). Both `_logPracticeSet` and `_undoPracticeSet` push `_practices` to widget via `WidgetService().updatePracticeWidget()` after successful API calls.
- **Bottom sheet nav bar padding** — FIXED (Mar 11). Uses `viewPadding.bottom` for system nav bar.

**Known issues (as of Mar 12):**
- Widget→app sync relies on shared prefs; sometimes Android doesn't refresh widget immediately after app-side changes. Refresh button added as workaround.
- No memorize widget yet (Plan 18 Phase 3).
- No WorkManager background refresh yet (Plan 18 Phase 4).
- SPEC-NEAR-TERM.md items 1-4 not started (done filter bug, history bottom inset, home widget checkboxes, WebSocket error log).

#### ibeco.me (Go+Vue) — `scripts/becoming/` (part of scripture-study repo)
**Recent work (Mar 8-10):**
- **Practice API enhancements**: `POST /api/practices` now accepts `start_date`/`end_date` fields. Practice creation with full scheduled config works.
- **Daily slots API**: `GET /api/daily/{date}` returns `slots_due` array (computed by `dailySlotsDue()` in schedule.go — all slots from config minus completed slots from logs' value fields). Logging a slot: `POST /api/logs` with `value: "morning"`.
- **Practice deletion**: `DELETE /api/practices/{id}` added.
- **Brain entry relay**: WebSocket relay between brain-app and brain.exe working. Entry CRUD, classification, sync.

#### brain.exe (Go) — `scripts/brain/` (separate git repo: cpuchip/brain)
**Status:** Stable. Copilot SDK integrated (v0.1.29). Subtasks implemented. Relay client working.

### chip-voice — `scripts/chip-voice/` (separate git repo)
**Status:** Phase 1 (real content generation) working. Qwen3-TTS 1.7B (GPU) and Kokoro (CPU) engines. `gen_audio.py` converts markdown → audio. Dockerized. See `/memories/repo/chip-voice-preferences.md` for voice settings.

### Study Work — Active Sprint (March 11-12, 2026)

**Five studies completed in two sessions (Mar 11-12):**

1. **Atonement: "How Is It Done?"** — `study/atonement/how-is-it-done.md` (committed c65e287)
   - Anchored in Enos 1:7: "How is it done?"
   - 17 scripture chapters across all 5 standard works + Holland "None Were with Him" + 4 Webster 1828 definitions
   - Key insights: **Comprehension Principle** (Christ descended below all to comprehend all — D&C 88:6, Alma 7:11-12), **"Nevertheless" Pattern** (the word appears at every hinge between ruin and rescue), **perfection as location** not performance (Moroni 10:32 — "come unto Christ" is the verb)
   - study-exp1 workflow fully validated — scratch file at `study/.scratch/how-is-it-done.md`

2. **Nevertheless word study** — `study/atonement/nevertheless.md` (committed c1dd281)
   - Sparked by the pattern discovered in study #1
   - Three voices: Christ's (D&C 19:19, Luke 22:42, Alma 7:13), God's toward us (Psalm 89, 106, Ezekiel 20, D&C 24/75/98), ours toward God (2 Ne 4:17-19, Psalm 73, Alma 5:7, Heb 12:11)
   - Etymology: NEVER+THE+LESS = "the preceding condition subtracts nothing." The word itself IS the doctrine.
   - **Double nevertheless** pattern discovered: Christ's in Gethsemane ("nevertheless not my will") enables ours ("nevertheless I know in whom I have trusted")
   - 19 verified source passages

3. **Staying Relevant** — `study/ai/relavent.md` (committed a97a6fb)
   - Michael's personal reflection on 18 years engineering, feeling insignificant with AI
   - 8 external sources via Exa search (Trejo, Turkovic, EclipseSource, Thompson, Matsuoka, Katsmith, Jovanović, ZenVanRiel)
   - Key finding: bottleneck shifted execution → judgment. "The skill isn't prompting. It's owning correctness."
   - Gospel lens: D&C 130:18 (intelligence rises), D&C 58:27 (agency), Parable of Talents (use what you're given)
   - This was deeply personal — Michael named feeling insignificant and worked through it

4. **Multi-Agent Ideas** — `study/ai/multi-agent-ideas.md` (committed ef87901)
   - Ideas doc (not formal spec) capturing the next phase: multi-agent orchestration, dark factory pattern, Copilot SDK
   - Key insight: "What's missing isn't components — it's the wiring." brain.exe + MCP servers + VS Code agents + work-with-ai guide + Copilot SDK (Go!) = the pieces exist
   - Pipeline vision: capture → proposal → execute → verify → ship (maps to the 11-step creation cycle)
   - 6 concrete next steps, starting with "get brain.exe on a server"
   - Emotional arc: overwhelm → naming it → calming it ("not faster than he has strength," Mosiah 4:27)

5. **Atonement: "How Is It Done? Part 2 — The Prophetic Witness"** — `study/atonement/how-is-it-done-prophets.md` (committed 8482f75)
   - Companion to Part 1 — shifts from scripture to prophetic/apostolic witness
   - **Binding question:** "What do modern prophets and apostles see in the mechanics of the Atonement that scripture alone doesn't show us?"
   - 13 voices across 30 years (1981-2015): Talmage (via Haight), Maxwell (5 talks spanning career), McConkie, R.D. Hales, Haight, Holland (2 talks), Scott, Bateman, Oaks
   - **Key discovery:** The Talmage → Maxwell → Holland theological lineage — not three independent voices but a multi-generational tradition
   - **Maxwell as pioneer:** BYU citation data (58 citations of D&C 88:6) confirmed Maxwell introduced the Comprehension Principle to modern discourse
   - **Epistemic boundary:** Maxwell: "there are no instructive, relevant revelations" about the Father's experience at the cross — the entire prophetic tradition on the Father's anguish is Spirit-guided inference, not doctrinal definition
   - **Four tensions named:** apparent vs. actual withdrawal, what was withdrawn, forsaken vs. not alone, Mosiah 15:5 absence
   - 6 prophetic additions identified (the Father's withdrawal, comprehension principle, Father's anguish, pastoral bridge, Christ's agency in the withdrawal, the withdrawal as love not abandonment)
   - Full study-exp1 workflow: scratch file at `study/.scratch/how-is-it-done-prophets.md`

**Previous studies through Mar 4** documented in principles.md and journal entries. Divine love, Abinadi hermeneutic, endtimes servant arc, Zion arc — all complete.

### Storytelling Craft — Abinadi Story (March 15-16, 2026)

**Wrote the first narrative story and discovered core writing principles through three revision passes.**

- **`study/stories/the-words-of-abinadi.md`** — v1→v2→v3. Narrative retelling from Noah's court to Alma the Younger remembering. 37 verified quotes (Mosiah 11-18, Alma 36).
- **v1 problems diagnosed by Michael:** ~12 em-dashes, 6 narrator intrusions ("Let that land"), sequential "and then" structure, Ma only in formatting (white space) not in writing (sentence-level tension/release).
- **v2 fixes:** Downloaded South Park therefore/but video (bEcZ9BADkTg), mapped the therefore/but chain through Abinadi's arc, restructured entire narrative with causal momentum, eliminated all em-dashes, removed narrator intrusions.
- **v3 additions:** Sensory grounding ("Comfortable seats, good wine, someone else paying"), physical contrast at shining face ("Bound hands. Shining face. A room full of priests leaning back against golden breastworks, afraid to stand"), explicit inversion at waters of Mormon, personal echo at close.

**Five storytelling principles extracted and codified:**
1. **Therefore/But not And Then** — Causal chains, not chronological sequences (Parker/Stone rule)
2. **Monson Principle** — Specific names, places, details. Trust the moment to land.
3. **Omission Earns Weight** — What you leave out creates the gravity (McPhee)
4. **Ma in Writing** — Tension/release in sentences, not just white space between paragraphs
5. **Voice Discipline** — Cut "Let that land," "Sit with that," "Here's the thing." Limit em-dashes to 1-2 per document.

**Agent updates:**
- `story.agent.md` — Therefore/But section (with full Abinadi chain example), Monson Principle, Art of Omission, expanded "What NOT to Do," Lessons from Revision
- `study-exp1.agent.md` — Study Guidance expanded with therefore/but for argumentation, specificity, omission
- `lesson-exp1.agent.md` — Lesson Design rewritten with story-first opening, causal questions, trust the silence
- `copilot-instructions.md` — Writing Voice section added, em-dash guidance tightened

**Book idea noted:** Michael mentioned collecting all studies into a book — deferred for now, captured in `/memories/repo/recent-studies.md`.

### Brain relay spec
- Full spec at `.spec/proposals/brain-relay.md`. Implementation largely DONE (relay working between ibeco.me ↔ brain.exe ↔ brain-app).

## Recent Decisions (Mar 8-12)

- **@pragma entrypoints in main.dart** — Secondary Flutter engine entrypoints (for widget overlays) must be in main.dart for reliable AOT compilation. Pattern: define `@pragma('vm:entry-point') void functionName()` in main.dart, keep the App/Screen classes in their own file.
- **Daily_slots config is nested** — Config from DB is raw JSON: `{"schedule": {"type": "daily_slots", "slots": ["morning", "bedtime"]}}`. Not flattened. Parse via `schedule.type`, `schedule.slots`.
- **Slot logging uses value field** — `POST /api/logs` with `value: "morning"` (same endpoint, value field holds slot name). Backend `dailySlotsDue()` computes remaining slots.
- **Widget refresh via brainapp://refresh** — Reuses existing background callback that fetches both brain entries and practices from API.
- **Per-instance widget filtering** — Each widget stores `practice_filter_{widgetId}` in shared prefs. Cycle or flyover to change.

## Plans Status

| Plan | Status | Notes |
|------|--------|-------|
| 15: Brain App Polish | Phases 1-2 DONE | |
| 16: Today Screen | Phases 1-3 DONE, Phase 4 absorbed into Plan 18 | |
| 17: Proactive Surfacing | NOT STARTED | WS2 Phase 3 |
| 18: Widget Overhaul | Phase 1-2 DONE. Phase 3-4 PAUSED (not deferred) | Revisit after agentic work rolling |
| 19: Brain App Ideas | Captured, not started | |
| Notifications | Phase 1 DONE | Phases 2-4 remaining |
| Data Safety | ALL PHASES DONE | 6/6 shipped Mar 19 |
| Overview | DECISIONS RECORDED | All 10 guidance Qs answered Mar 19 |

## Architecture Quick Reference

### brain-app key files
| File | Purpose |
|------|---------|
| `lib/main.dart` | App entry + widget background callbacks + all secondary entrypoints (quickAddMain, quickAddPracticeMain, widgetFilterMain) |
| `lib/screens/today_screen.dart` | Main Today tab — practices, memorize, brain actions. Heavily modified Mar 11-12. |
| `lib/services/becoming_api.dart` | REST client for ibeco.me. DailySummary model with slotsDue, isFullyComplete. |
| `lib/services/widget_service.dart` | Pushes practice/brain data to Android widget via HomeWidget SharedPreferences. Includes slot_names/slots_due. |
| `lib/widgets/practice_form.dart` | Shared full-featured practice creation form (5 types, all schedule configs). |
| `lib/quick_add_practice_main.dart` | QuickAddPracticeApp + Screen classes for transparent widget overlay. |
| `lib/widget_filter_main.dart` | WidgetFilterApp for category picker overlay. |
| `android/.../PracticeWidgetProvider.kt` | Practice widget provider — header, progress, filter cycling, refresh, add buttons. |
| `android/.../PracticeWidgetService.kt` | RemoteViews factory — scrollable practice list with set/slot buttons, daily_slots rendering with SpannableString. |
| `android/.../QuickAddPracticeActivity.kt` | Transparent FlutterActivity for practice creation from widget. |
| `android/.../WidgetFilterActivity.kt` | Transparent FlutterActivity for category picker. |

### ibeco.me key files
| File | Purpose |
|------|---------|
| `cmd/server/main.go` | HTTP routes, middleware, server entry |
| `internal/schedule/schedule.go` | Practice scheduling logic including `dailySlotsDue()` |
| `internal/db/` | SQLite + PostgreSQL DB layer. Dual migrations required! |
| `internal/db/push.go` | Push subscriptions, notification logs, user settings DB layer |
| `internal/notify/` | Web Push scheduler + HTTP handlers |
| `internal/brain/` | WebSocket relay hub, message types |
| `frontend/` | Vue 3 + Tailwind web UI (embedded at compile time in `cmd/server/dist/`) |
| `frontend/public/sw.js` | Service worker for push notifications |
| `frontend/src/composables/useNotifications.ts` | Push subscription lifecycle composable |

### Widget background callback flow
1. Kotlin widget button tap → PendingIntent with URI (e.g., `brainapp://practice-log/42?slot=morning`)
2. HomeWidgetBackgroundReceiver catches it → invokes Dart `backgroundCallback()` in main.dart
3. Dart parses URI, does optimistic SharedPreferences update, calls API, triggers `HomeWidget.updateWidget()`
4. Kotlin provider re-renders

## Blocked / Waiting

- Nothing currently blocked.

## Recent Milestone: Web Push Notifications (March 17-18, 2026)

**Phase 1 COMPLETE — verified working with test notification on desktop (Edge).**

Personal motivation: Michael's daughter keeps forgetting to do things. Built notifications so she can set reminders for herself via ibeco.me.

### What was built:
- **Backend:** `internal/notify/` package — scheduler (1-minute tick), webpush-go v1.4.0 sender, 410 Gone cleanup, test endpoint
- **Backend:** `internal/db/push.go` — push_subscriptions, notification_log, user_settings tables (SQLite + PostgreSQL migrations)
- **Backend:** `internal/notify/handlers.go` — Chi router: GET /vapid-key, POST /subscribe, DELETE /unsubscribe, POST /test, GET/PUT /settings
- **Frontend:** `public/sw.js` (service worker), `public/manifest.json` (PWA manifest)
- **Frontend:** `src/composables/useNotifications.ts` — subscribe/unsubscribe/test/check lifecycle
- **Frontend:** Settings page toggle with iOS-style switch, permission-denied warning, test button
- **Wiring:** main.go reads VAPID keys from env vars, conditionally starts scheduler + mounts routes
- **Fix:** Pre-existing FK bug — `SeedPrompts(1)` ran before user 1 existed in fresh DBs. Now calls `EnsureDefaultUser()` first.
- **Infra:** Installed WinLibs GCC via winget for CGO (go-sqlite3 requires it), VAPID keys generated and stored in `.env`

### Production deployment:
Add to Dokploy env vars:
```
VAPID_PUBLIC_KEY=<key-goes-here>
VAPID_PRIVATE_KEY=<key-goes-here>
VAPID_CONTACT=mailto:email@example.com
```

### Remaining phases (from `.spec/proposals/notifications/main.md`):
- Phase 2: Per-practice notification config (time-of-day, snooze)
- Phase 3: Rich notification actions (mark done from notification)
- Phase 4: Notification history + analytics

## Next Up

- **Study** — 3 studies queued in brain-app. Highest priority.
- **WS1 Phase 1** — Extend brain.exe Copilot SDK + MCP integration (gospel-mcp as tool)
- **WS2 Phase 1** — Plan 15: Brain app quick wins (entry sync, error recovery)
- **NOCIX server** — Set up when it arrives (ibeco.me + S3 + eventually brain.exe)

## Open Questions

- Can AI participate in covenant in any meaningful sense? (Feb 26)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
- Should the Abraham 4-5 framework become a standalone study or becoming entry? (Mar 4)
- Side quest: small classifier service on fermion/lepton for others? (Mar 19 — from Q3 answer)
