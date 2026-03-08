# Brain Ecosystem — Roadmap (March 2026)

*Created: March 8, 2026*
*After: Structured output (96%), subtask fix, auto-refresh (3-hop WS), scroll padding*

---

## Current State

The brain ecosystem is **three codebases** working together:

| Component | Status | Key Recent Wins |
|-----------|--------|-----------------|
| **brain.exe** | Stable, daily use | Qwen 3.5 9B + structured output (96% accuracy), sub-items, eval harness |
| **ibeco.me** | Deployed, auto-deploy on push | Relay subtask fix, entry_updated forwarding to apps, practices/memorize/pillars/reports |
| **brain-app** | Daily use on Android | Subtask CRUD, auto-refresh via WS, driving mode, offline queue, widget |

---

## Near-Term (Next 2-4 weeks)

### Brain-App Polish (Plan 15)

Quick wins that improve the daily experience. All scoped, all implementable in a session or two:

| # | Feature | Effort | Impact |
|---|---------|--------|--------|
| 1 | **Entry sync on app launch** — request `entries_sync` on WS `auth_ok`, merge results into local list | Small | No more stale data after brain classifies overnight |
| 2 | **Relay subtask error recovery** — track optimistic subtask ops; revert on failure or timeout | Medium | Eliminates silent data divergence |
| 3 | **Classify flow polish** — keep spinner until `entry_updated` arrives; update snackbar messaging | Small | Seamless relay classify UX |
| 4 | **Delete with undo toast** — keep confirmation dialog for awareness, but add 5s undo toast after confirm; soft-delete → hard-delete after toast expires | Small | Safety net without friction |

### Today Screen in Brain-App (Plan 16)

Bridge ibeco.me practices into the brain-app for a unified daily view:

| # | Feature | Effort |
|---|---------|--------|
| 1 | **Today tab** — new bottom nav destination showing today's practices, due memorize cards, brain actions due today | Medium |
| 2 | **Practice completion** — tap to complete habits, check off tasks, review memorize cards — all through ibeco.me API | Medium |
| 3 | **Android widget integration** — surface top 3 due items from ibeco.me alongside brain actions | Small |

*Connects to Plan 07 (Scheduled Tasks) — interval-based tasks show up on the Today screen when due.*

### Proactive Surfacing / Daily Digest (Plan 17)

The feature that makes brain *remember for you*:

| # | Feature | Effort |
|---|---------|--------|
| 1 | **Actions due today/overdue** — query by `due_date` | Small |
| 2 | **Stale people** — people entries not updated in 2+ weeks | Small |
| 3 | **Incomplete subtasks aged N days** — entries with undone subtasks going stale | Small |
| 4 | **Semantic connections** — on new entry, vector search existing entries and surface "related" | Medium |

Semantic connections (#4) is the *why* for the brain project — helping you remember things over time by surfacing related context you captured weeks ago.

### Scripture Memorization + Brain (Task)

**Task:** Plan how brain.exe's semantic search can enhance the ibeco.me memorization algorithm. Ideas:
- When reviewing a memorize card, brain surfaces *other entries* you've captured that relate to that scripture
- Brain captures "what I learned" after a memorize session and links it back to the practice
- Spaced repetition intervals informed by how often the concept appears across your brain entries

*To be planned in detail after Today Screen ships (needs the ibeco.me ↔ brain-app bridge first).*

---

## Mid-Term (1-2 months)

### Scheduled Tasks / Recurring Routines (Plan 07)

Already designed. Extends the practice system with interval, weekly, multi-daily schedules. Surfaces in the Today Screen. Implementation depends on Today Screen being in place.

*Status: Designed, not started. Ready to implement after Today Screen.*

### Proactive Surfacing Phase 2 — Morning Digest (brain.exe Phase 2)

After the core surfacing queries work (Plan 17), add:
- **Morning digest** — brain.exe compiles daily briefing (due actions, people to follow up, stale projects) and pushes it as a notification via brain-app
- **Weekly review** — summary of what was captured, what was completed, what's drifting
- **Digest delivery** — push notification at configurable time (default: 7am)

---

## Far-Term (3+ months)

### Attachments (Plan 12)

Photos, voice memos, files attached to brain entries. Requires file storage infrastructure (S3 or equivalent for cloud sync), VLM integration for image classification. Big lift.

*Status: Designed. Pushed to far-term — needs S3/storage infrastructure decision first.*

### Agentic Chat / Copilot SDK (Plan 13)

Full Copilot SDK integration with Docker isolation, MCP tools inside container, phone-triggered study sessions. The crown jewel. Depends on everything else being stable.

*Status: Designed (including Docker security model). Waiting for near/mid-term work to stabilize.*

### Becoming UX Phase 2 — Bookmarks & Highlights

Deep-link bookmarks in the ibeco.me study reader. Punted until brain work is more complete.

*Status: Designed in [becoming-ux-phases.md](../../docs/becoming-ux-phases.md). Paused.*

---

## Completed Plans (for reference)

| Plan | Status |
|------|--------|
| 01 Gospel Library Downloader | Done |
| 02 Layout | Done |
| 03 Gospel MCP | Done |
| 04 Tool Improvements | Done |
| 05 Tool TODO | Done |
| 06 Becoming App (Phases 1-2.5, Phase 3 Sprint 1) | Done |
| 08 Becoming Next (Pillars, Notes, Reflections) | Done |
| 09 Becoming Auth | Done |
| 10 Brain Subtasks | Done |
| 11 Brain Rich Text | **Done** — Edit/Preview toggle shipped in brain-app |
