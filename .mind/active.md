# Active Context

*Updated: 2026-04-21 · Previous: [archive/active-2026-04-20.md](archive/active-2026-04-20.md)*
*Hardware: Dual 4090s desktop. NOCIX server live.*

> **Edit rule:** To rewrite this file, write the new content directly. Do NOT cat/append the existing content first — its archive snapshot lives in its own file under `.mind/archive/`. Appending to the live file (instead of replacing it) duplicates the document and silently doubles every memory load. Bug observed 2026-04-20 → 2026-04-21.

---

## Priorities

1. ★ **Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" → [.spec/proposals/study-workstream.md](../.spec/proposals/study-workstream.md)
2. ★ **Teaching** — 11-episode arc (Option C). Agent + repo scaffolded. Content not started → [.spec/proposals/teaching-workstream.md](../.spec/proposals/teaching-workstream.md)
3. **Spec hygiene** — cleanup-2026-04 in progress. Phases 1-3 today → [.spec/proposals/cleanup-2026-04/main.md](../.spec/proposals/cleanup-2026-04/main.md)
4. **Token efficiency** — proposal from Apr 16 awaiting refresh review → [.spec/proposals/token-efficiency.md](../.spec/proposals/token-efficiency.md)

## Milestones

| Date | Event |
|------|-------|
| Mar 22 | Last Sabbath. "Infrastructure and Foundation" declared good |
| Apr 5 | Project sync: 101 entries (30 verified, 10 specced, 4 planned, 57 raw) |
| Apr 12 | 🏆 First fully automated commission (Space Center). 39 premium requests. AI converged on family's independent brainstorm |
| Apr 20 | engine.ibeco.me Phase 1-3 shipped. First study on user-minted engine token: "I Will Give Away All My Sins to Know Thee" |
| Apr 21 | Voice/bias harness updated (em-dash budget, three-beat pivot, refrains, stats audit). cleanup-2026-04 proposal written + Phase 1-3 executed |

## Key Decisions (Recent)

- Both paths: simplified (notebook, 3-col) AND automated pipeline
- KISS for captures, power for delegation. 90% simple, 10% delegation
- gospel-engine v2 (hosted at engine.ibeco.me) is the single canonical search backend; gospel-mcp + gospel-vec retired as fallback only
- `.mind/` is the canonical memory location; `.spec/memory/` deleted (git preserves it)

## Key Facts

- Copilot: 1500 premium/mo ($40 Pro+). Haiku 4.5=0.33x, Sonnet 4.6=1.0x, Opus 4.7=7.5x, GPT-5/5-mini/4.1/4o=0
- Brain default model: gpt-5-mini (0x). Pipeline big = claude-opus-4.7 (7.5x)
- Claude Code: Pro $20/mo. 200K context. Project caching
- Pipeline cost: research=0.33 + plan=1.0 = 1.33/entry
- Active MCP servers: gospel-engine-v2 (engine.ibeco.me), webster, yt, byu-citations, becoming, exa-search

---

## In Flight

### cleanup-2026-04
✅ Phase 1-3 done today (spec dedup, gospel-engine reorg as v2-hosted.md, model audit + brain test fixes opus-4.6→4.7). Phase 4 (token-efficiency execution) deferred pending Michael's review of [tokenomics-2026](../.spec/proposals/tokenomics-2026/main.md)
→ [.spec/proposals/cleanup-2026-04/main.md](../.spec/proposals/cleanup-2026-04/main.md)

### Brain Inline Panel + Nudge Bot Controls
▶ P1: Reply textarea + close-with-reason slide-out. P2: Nudge bot in Scheduled Tasks
→ [.spec/proposals/brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md)

### Token Efficiency & Memory Architecture v2
⏸ Proposal from Apr 16 needs refresh before execution. ~25K tokens at session start → target ≤10K
→ [.spec/proposals/token-efficiency.md](../.spec/proposals/token-efficiency.md)

### Other In-Flight

| Item | Status | Ref |
|------|--------|-----|
| WS1 P4d: Pipeline REST + Execution | ▶ next | [.spec/proposals/brain-phase4-pipeline.md](../.spec/proposals/brain-phase4-pipeline.md) |
| Claude Code Integration | researched | [.spec/proposals/claude-code-integration.md](../.spec/proposals/claude-code-integration.md) |
| Brain Windows Service (systray) | specced | [.spec/proposals/brain-windows-service.md](../.spec/proposals/brain-windows-service.md) |
| ibeco.me Security Audit | not started | — |
| Gospel Engine v1.5 Ergonomics | specced | [.spec/proposals/gospel-engine/phase1.5-ergonomics.md](../.spec/proposals/gospel-engine/phase1.5-ergonomics.md) |
| Gospel Graph Visualization | specced, depends on engine v2 | [.spec/proposals/gospel-graph/main.md](../.spec/proposals/gospel-graph/main.md) |
| tokenomics-2026 (research) | placeholder | [.spec/proposals/tokenomics-2026/main.md](../.spec/proposals/tokenomics-2026/main.md) |

---

## Recently Shipped (rolling, last ~30 days)

Move to archive when older than ~60 days or when scope is fully closed.

| Workstream | Shipped | Notes |
|------------|---------|-------|
| engine.ibeco.me Phase 1-3 | Apr 20 | Hosted gospel search at engine.ibeco.me. Token UI in ibeco.me Settings. First study used it |
| Voice/bias harness v2 | Apr 21 | em-dash budget, three-beat pivot detection, stats cite-count rule extension |
| Commission UX Fixes | Apr 15 | Path mangling, link normalization, and gaps surfaced from Space Center test |
| Brain Project-Kanban | Apr 4-5 | All phases. Projects, kanban, auto-assignment, AI push-back |
| Orchestrator Steward P1-6 | Apr 10-11 | Failure retry, model escalation, circuit breaker, quarantine, nudge bot, commission. 86 tests |
| Commission UI P1-3 | Apr 11 | Types/API, dialog, triggers, status panel, badge, +New Entry dialog |
| WS3 Brain UX QoL P1-7b | Apr 6 | Textarea, markdown render, file viewer, file browser, WebSocket push, cost tracking, reader UX, git status, inline diff, nested git repos |
| WS4 Brain Pipeline Evolution P1-9 | Apr 6-7 | Governance, failure visibility, reflection pauses, notebook mode, nudge bot, 3-col board, schema injection, project scaffolding, agent-driven init, project-aware pipeline |

---

## Deferred / Paused

| Item | Revisit When |
|------|------|
| Plan 17: Proactive Surfacing | WS2 Phase 3 |
| Plan 18: Widget Overhaul (Ph 3-4) | Agent infra proves out |
| Plan 19: Brain App Ideas | Natural pause |
| Notifications (Ph 2-4) | After systray |
| Progressive Trust Tracking | D&C 107, model experiments |
| Squad A4/A5/A9 | After pipeline stabilizes |
