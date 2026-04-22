# Active Context

*Updated: 2026-04-21 · Previous: [archive/active-2026-04-20.md](archive/active-2026-04-20.md)*
*Hardware: Dual 4090s desktop. NOCIX server live.*

> **Edit rule:** To rewrite this file, write the new content directly. Do NOT cat/append the existing content first — its archive snapshot lives in its own file under `.mind/archive/`. Appending to the live file (instead of replacing it) duplicates the document and silently doubles every memory load. Bug observed 2026-04-20 → 2026-04-21.

> **Workstream taxonomy:** [.mind/workstreams.md](workstreams.md) is the canonical list of WS1–WS9 + status enum + frontmatter convention. Every active proposal carries `workstream:` frontmatter. When tagging anything new, read that file first.

---

## Priorities

*Workstream tags per [.mind/workstreams.md](workstreams.md).*

1. ★ **[WS6] Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" → [.spec/proposals/study-workstream.md](../.spec/proposals/study-workstream.md)
2. ★ **[WS7] Teaching** — 11-episode arc (Option C). Agent + repo scaffolded. Content not started → [.spec/proposals/teaching-workstream.md](../.spec/proposals/teaching-workstream.md)
3. **[WS5] Spec hygiene** — cleanup-2026-04-part2 in progress. Phase A done, Phase B (16 archives + kanban Status fix + P8 spinoff) done today → [.spec/proposals/cleanup-2026-04-part2/main.md](../.spec/proposals/cleanup-2026-04-part2/main.md)
4. **[WS5] Token efficiency** — proposal from Apr 16 awaiting refresh review → [.spec/proposals/token-efficiency.md](../.spec/proposals/token-efficiency.md)

## Milestones

| Date | Event |
|------|-------|
| Mar 22 | Last Sabbath. "Infrastructure and Foundation" declared good |
| Apr 5 | Project sync: 101 entries (30 verified, 10 specced, 4 planned, 57 raw) |
| Apr 12 | 🏆 First fully automated commission (Space Center). 39 premium requests. AI converged on family's independent brainstorm |
| Apr 20 | engine.ibeco.me Phase 1-3 shipped. First study on user-minted engine token: "I Will Give Away All My Sins to Know Thee" |
| Apr 21 | Voice/bias harness updated (em-dash budget, three-beat pivot, refrains, stats audit). cleanup-2026-04 P1-3 + part2 Phases A-B executed (16 proposals archived) |

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

### [WS5] cleanup-2026-04 / part2
✅ cleanup-2026-04 P1-3 done Apr 21 (spec dedup, gospel-engine reorg, model audit + brain test fixes opus-4.6→4.7). Phase 4 deferred to [tokenomics-2026](../.spec/proposals/tokenomics-2026/main.md).
✅ cleanup-2026-04-part2 Phases A-B done Apr 21 (active.md duplication fix, 16 proposals archived, P8 split off, kanban Status corrected). Phase C in progress (workstream tagging). Phases D-E pending.
→ [cleanup-2026-04/main.md](../.spec/proposals/cleanup-2026-04/main.md) · [cleanup-2026-04-part2/main.md](../.spec/proposals/cleanup-2026-04-part2/main.md)

### [WS5] Brain ↔ VS Code Bridge (proposed)
📝 Architecture proposal written Apr 21. Bidirectional sync between proposals and brain entries. Phase 1-2 carry most of the value.
→ [.spec/proposals/brain-vscode-bridge/main.md](../.spec/proposals/brain-vscode-bridge/main.md)

### [WS2] Brain Inline Panel + Nudge Bot Controls
▶ P1: Reply textarea + close-with-reason slide-out. P2: Nudge bot in Scheduled Tasks
→ [.spec/proposals/brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md)

### [WS5] Token Efficiency & Memory Architecture v2
⏸ Proposal from Apr 16 needs refresh before execution. ~25K tokens at session start → target ≤10K
→ [.spec/proposals/token-efficiency.md](../.spec/proposals/token-efficiency.md)

### Other In-Flight

| WS | Item | Status | Ref |
|----|------|--------|-----|
| WS2 | Brain Project-Kanban Phase 4c | ▶ next | [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md) |
| WS5 | Claude Code Integration | researched | [claude-code-integration.md](../.spec/proposals/claude-code-integration.md) |
| WS2 | Brain Windows Service (systray) | proposed | [brain-windows-service.md](../.spec/proposals/brain-windows-service.md) |
| WS4 | ibeco.me Security Audit | not started | — |
| WS3 | Gospel Engine v1.5 Ergonomics | proposed | [gospel-engine/phase1.5-ergonomics.md](../.spec/proposals/gospel-engine/phase1.5-ergonomics.md) |
| WS3 | Gospel Graph Visualization | proposed (blocked on AGE/PG18) | [gospel-graph/main.md](../.spec/proposals/gospel-graph/main.md) |
| WS5 | tokenomics-2026 (research) | placeholder | [tokenomics-2026/main.md](../.spec/proposals/tokenomics-2026/main.md) |
| WS5 | Sabbath agent | ready to build | [sabbath-agent.md](../.spec/proposals/sabbath-agent.md) |
| WS1 | Classifier qwen fix | ready to build | [classifier-qwen-fix.md](../.spec/proposals/classifier-qwen-fix.md) |
| WS1 | Classification quality benchmark | proposed | [classify-bench.md](../.spec/proposals/classify-bench.md) |
| WS1 | Data safety: dev agent hardening + audit log | proposed | [data-safety/main.md](../.spec/proposals/data-safety/main.md) |
| WS5 | Debug agent: layer triage enhancement | proposed | [debug-layer-triage.md](../.spec/proposals/debug-layer-triage.md) |
| WS4 | study.ibeco.me UI | proposed | [study-ibeco-me/main.md](../.spec/proposals/study-ibeco-me/main.md) |

---

## Recently Shipped (rolling, last ~30 days)

Move to archive when older than ~60 days or when scope is fully closed.

| WS | Workstream / Item | Shipped | Notes |
|----|--------------------|---------|-------|
| WS3 | engine.ibeco.me Phase 1-3 | Apr 20 | Hosted gospel search at engine.ibeco.me. Token UI in ibeco.me Settings. First study used it |
| WS5 | Voice/bias harness v2 | Apr 21 | em-dash budget, three-beat pivot detection, stats cite-count rule extension |
| WS1 | Commission UX Fixes | Apr 15 | Path mangling, link normalization, and gaps surfaced from Space Center test |
| WS2 | Brain Project-Kanban | Apr 4-5 | Phases 1-3 + 4a-4b. Projects, kanban, auto-assignment, AI push-back. **Phase 4c still pending — see In Flight** |
| WS1 | Orchestrator Steward P1-6 | Apr 10-11 | Failure retry, model escalation, circuit breaker, quarantine, nudge bot, commission. 86 tests |
| WS2 | Commission UI P1-3 | Apr 11 | Types/API, dialog, triggers, status panel, badge, +New Entry dialog |
| WS2 | Brain UX QoL P1-7b | Apr 6 | Textarea, markdown render, file viewer, file browser, WebSocket push, cost tracking, reader UX, git status, inline diff, nested git repos |
| WS1 | Brain Pipeline Evolution P1-9 | Apr 6-7 | Governance, failure visibility, reflection pauses, notebook mode, nudge bot, 3-col board, schema injection, project scaffolding, agent-driven init, project-aware pipeline |

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
| Brain UX QoL Phase 8 (auto-commit) | After Phase 7 in daily use → [.spec/proposals/brain-ux-qol-phase8-autocommit.md](../.spec/proposals/brain-ux-qol-phase8-autocommit.md) |
