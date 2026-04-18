# Active Context

*Updated: 2026-04-16 · Previous: [archive/active-2026-04-04.md](archive/active-2026-04-04.md)*
*Hardware: Dual 4090s desktop. NOCIX server live.*

---

## Priorities

1. ★ **Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" → [.spec89/proposals/study-workstream.md]
2. ★ **Teaching** — 11-episode arc (Option C). Agent + repo scaffolded. Content not started → [.spec/proposals/teaching-workstream.md]
3. **Brain pipeline** — Phase 4 complete. Pipeline fixes ALL SHIPPED (Apr 9-10). Next: simplification + inline panel + UX QoL
4. **Token efficiency** — NEW (Apr 16). Compress memory, tiered loading → [.spec/proposals/token-efficiency.md]

## Milestones

| Date | Event |
|------|-------|
| Mar 22 | Last Sabbath. "Infrastructure and Foundation" declared good |
| Apr 5 | Project sync: 101 entries (30 verified, 10 specced, 4 planned, 57 raw) |
| Apr 5 | KISS reflection: pipeline overserves 90% of entries. Both paths needed |
| Apr 12 | 🏆 First fully automated commission (Space Center). 39 premium requests. 18+ research docs. AI converged on family's independent brainstorm |

## Key Decisions (Recent)

- Both paths: simplified (notebook, 3-col) AND automated pipeline
- KISS for captures, power for delegation. 90% simple, 10% delegation
- Space Center as pipeline test bed
- "By small and simple things" — build, use, iterate

## Key Facts

- Copilot: 1500 premium/mo ($40 Pro+). Haiku=0.33, Sonnet=1.0, Opus=3.0, GPT-4.1/4o/5-mini=0
- Claude Code: Pro $20/mo. 200K context. Project caching
- Pipeline cost: research=0.33 + plan=1.0 = 1.33/entry. Nudge: 0.33/entry, ≤10 entries 4x/day
- Space Center: dream business (planetarium, science center, bridge sim). Haiku needs this in prompt

---

## In Flight

### Brain Inline Panel + Nudge Bot Controls
▶ P1: Reply textarea + close-with-reason slide-out. P2: Nudge bot in Scheduled Tasks
→ [.spec/proposals/brain-inline-panel.md]

### Orchestrator Steward
✓ P1-6 ALL COMPLETE (Apr 10-11): failure retry, model escalation (Haiku→Sonnet→Opus→Human), circuit breaker, quarantine queue, nudge bot integration, commission model. 86 tests
→ [.spec/proposals/orchestrator-steward/main.md]
▶ Next: E2E commission testing. Consider P7+ (multi-entry, project-scope)

### Commission UI
✓ P1-3 ALL COMPLETE: types/API, dialog, triggers, status panel, badge, guards, "+New Entry" dialog
→ [.spec/proposals/commission-ui.md]

### Commission UX Fixes
▶ From Space Center test. Real usage gaps discovered
→ [.spec/proposals/commission-ux-fixes.md]

### WS3: Brain UX Quality-of-Life
✓ P1-7b ALL COMPLETE (Apr 6): textarea, markdown render, file viewer, file browser, WebSocket push, cost tracking, reader UX, git status, inline diff, nested git repos
→ [.spec/proposals/brain-ux-quality-of-life.md]

### WS4: Brain Pipeline Evolution
✓ P1-9 ALL COMPLETE (Apr 6-7): governance docs, failure visibility, reflection pauses, notebook mode, nudge bot controls, 3-col board, schema+governance injection+context cap, project scaffolding, agent-driven init, project-aware pipeline (selective git commit)
→ [.spec/proposals/brain-pipeline-evolution.md]
→ [.spec/proposals/project-aware-pipeline.md]

### Brain Project-Kanban
✓ ALL PHASES COMPLETE (Apr 4-5). Projects, iterative sessions, scheduled tasks, library, kanban, auto-assignment, AI push-back, context injection, execution gate
→ [.spec/proposals/brain-project-kanban.md]

### Token Efficiency & Memory Architecture v2
▶ NEW (Apr 16). ~25K tokens at session start → target ≤10K
P1: Compress active.md. P2: Tiered loading. P3: ctx CLI. P4: Symbol standard. P5: Inherent audit. P6: PG hybrid (conditional)
→ [.spec/proposals/token-efficiency.md]

### Other In-Flight

| Item | Status | Ref |
|------|--------|-----|
| WS1 P4d: Pipeline REST + Execution | ▶ next | [.spec/proposals/brain-phase4-pipeline.md] |
| Claude Code Integration | researched | [.spec/proposals/claude-code-integration.md] |
| Brain Windows Service (systray) | specced | [.spec/proposals/brain-windows-service.md] |
| ibeco.me Security Audit | not started | — |
| Gospel Engine P1.5 Ergonomics | specced | [.spec/proposals/gospel-engine/phase1.5-ergonomics.md] |
| Gospel Graph Visualization | specced, after gospel-engine | [.spec/proposals/gospel-graph/main.md] |

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
