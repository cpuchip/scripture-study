# Active Context

*Updated: 2026-04-22 · Previous: [archive/active-2026-04-21.md](archive/active-2026-04-21.md)*
*Hardware: Dual 4090s desktop. NOCIX server live.*

> **Edit rule:** Rewrite this file directly. Do NOT cat/append the old content first — its archive snapshot lives under `.mind/archive/`. Appending duplicates the document and doubles every memory load. (Bug: 2026-04-20.)

> **Workstream taxonomy:** [.mind/workstreams.md](workstreams.md) is canonical: WS1–WS9, status enum, frontmatter convention. Every active proposal carries `workstream:` frontmatter. Read that file before tagging anything new.

---

## Priorities

1. ★ **[WS6] Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" → [study-workstream.md](../.spec/proposals/study-workstream.md)
2. ★ **[WS7] Teaching** — 11-episode arc (Option C). Agent + repo scaffolded. Content not started → [teaching-workstream.md](../.spec/proposals/teaching-workstream.md)
3. **[WS5] Token efficiency** — Apr 16 proposal awaiting refresh review → [token-efficiency.md](../.spec/proposals/token-efficiency.md)
4. **[WS2] Brain Inline Panel + Kanban 4c** — next active build → [brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md) · [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md)

## Key Facts

- Copilot: 1500 premium/mo ($40 Pro+). Haiku 4.5=0.33x, Sonnet 4.6=1.0x, Opus 4.7=7.5x, GPT-5/5-mini/4.1/4o=0
- Brain default model: gpt-5-mini (0x). Pipeline big = claude-opus-4.7 (7.5x)
- Claude Code: Pro $20/mo. 200K context. Project caching
- Pipeline cost: research=0.33 + plan=1.0 = 1.33/entry
- Active MCP servers: gospel-engine-v2 (engine.ibeco.me), webster, yt, byu-citations, becoming, exa-search
- gospel-engine v2 hosted is the single canonical search backend; gospel-mcp + gospel-vec retired as fallback

---

## In Flight

| WS | Item | Status | Ref |
|----|------|--------|-----|
| WS2 | Brain Inline Panel + Nudge Bot Controls | ▶ P1 next | [brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md) |
| WS2 | Brain Project-Kanban Phase 4c | ▶ next | [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md) |
| WS5 | Token Efficiency & Memory v2 | ⏸ awaiting refresh | [token-efficiency.md](../.spec/proposals/token-efficiency.md) |
| WS5 | Brain ↔ VS Code Bridge | � building (Phase 0 shipped Apr 22) | [brain-vscode-bridge/main.md](../.spec/proposals/brain-vscode-bridge/main.md) |
| WS2 | Brain non-pipeline projects | 📝 proposed Apr 22 | [brain-non-pipeline-projects.md](../.spec/proposals/brain-non-pipeline-projects.md) |
| WS2 | Brain manual stage transitions | 📝 proposed Apr 22 | [brain-manual-stage-transitions.md](../.spec/proposals/brain-manual-stage-transitions.md) |
| WS2 | Johari window agent mode | 📝 proposed Apr 22 | [johari-window-agent-mode.md](../.spec/proposals/johari-window-agent-mode.md) |
| WS2 | Motivation coach agent mode | 📝 proposed Apr 22 | [motivation-coach-agent-mode.md](../.spec/proposals/motivation-coach-agent-mode.md) |
| WS3 | LightRAG investigation | 📝 proposed Apr 22 | [lightrag-investigation.md](../.spec/proposals/lightrag-investigation.md) |
| WS3 | Gospel engine v3 proxy-pointer | 📝 proposed Apr 22 | [gospel-engine-v3-proxy-pointer.md](../.spec/proposals/gospel-engine-v3-proxy-pointer.md) |
| WS5 | VS Code agent hooks integration | 📝 proposed Apr 22 | [vscode-agent-hooks-integration.md](../.spec/proposals/vscode-agent-hooks-integration.md) |
| WS5 | Memory & context research bundle | 📝 proposed Apr 22 | [memory-research-bundle.md](../.spec/proposals/memory-research-bundle.md) |
| WS6 | Study: Nate B Jones on delegation | 📝 proposed Apr 22 | [study-nate-jones-delegation.md](../.spec/proposals/study-nate-jones-delegation.md) |
| WS7 | AI presentation site tool | 📝 proposed Apr 22 | [ai-presentation-site-tool.md](../.spec/proposals/ai-presentation-site-tool.md) |
| WS7 | Launch YouTube channel | 📝 proposed Apr 22 | [launch-youtube-channel.md](../.spec/proposals/launch-youtube-channel.md) |
| WS5 | Sabbath agent | ready to build | [sabbath-agent.md](../.spec/proposals/sabbath-agent.md) |
| WS5 | Debug agent: layer triage | proposed | [debug-layer-triage.md](../.spec/proposals/debug-layer-triage.md) |
| WS5 | Claude Code Integration | researched | [claude-code-integration.md](../.spec/proposals/claude-code-integration.md) |
| WS5 | tokenomics-2026 (research) | placeholder | [tokenomics-2026/main.md](../.spec/proposals/tokenomics-2026/main.md) |
| WS1 | Classifier qwen fix | ready to build | [classifier-qwen-fix.md](../.spec/proposals/classifier-qwen-fix.md) |
| WS1 | Classification quality benchmark | proposed | [classify-bench.md](../.spec/proposals/classify-bench.md) |
| WS1 | Data safety: dev hardening + audit log | proposed | [data-safety/main.md](../.spec/proposals/data-safety/main.md) |
| WS2 | Brain Windows Service (systray) | proposed | [brain-windows-service.md](../.spec/proposals/brain-windows-service.md) |
| WS3 | Gospel Engine v1.5 Ergonomics | proposed | [gospel-engine/phase1.5-ergonomics.md](../.spec/proposals/gospel-engine/phase1.5-ergonomics.md) |
| WS3 | Gospel Graph Visualization | proposed (blocked on AGE/PG18) | [gospel-graph/main.md](../.spec/proposals/gospel-graph/main.md) |
| WS4 | study.ibeco.me UI | proposed | [study-ibeco-me/main.md](../.spec/proposals/study-ibeco-me/main.md) |
| WS4 | ibeco.me Security Audit | not started | — |

---

## Recently Shipped (last ~30 days)

| WS | Item | Shipped | Notes |
|----|------|---------|-------|
| WS2 | Brain audit 2026-04-22 | Apr 22 | 69 entries triaged. Personal merged into Notebook. inbox 39→10, status=NULL 96→5 |
| WS5 | Brain↔VS Code bridge Phase 0 | Apr 22 | Schema migration (workstream + proposal_path columns). Read-only inspector at scripts/harness/harness_inspect.py |
| WS5 | cleanup-2026-04 + part2 (all phases) | Apr 21-22 | Spec dedup, 19 proposals archived, workstream taxonomy + frontmatter convention, active.md rewritten |
| WS5 | Voice/bias harness v2 | Apr 21 | em-dash budget, three-beat pivot, refrains, stats cite-count |
| WS3 | engine.ibeco.me Phase 1-3 | Apr 20 | Hosted gospel search. Token UI. First study used it |
| WS1 | Commission UX Fixes | Apr 15 | Path mangling, link normalization (Space Center) |
| WS1 | Orchestrator Steward P1-6 | Apr 10-11 | Retry, escalation, circuit breaker, quarantine, nudge, commission. 86 tests |
| WS2 | Commission UI P1-3 | Apr 11 | Types/API, dialog, triggers, status panel, badge |
| WS2 | Brain UX QoL P1-7b | Apr 6 | Textarea, markdown, file viewer, WebSocket push, cost tracking, git status, nested repos |
| WS1 | Brain Pipeline Evolution P1-9 | Apr 6-7 | Governance, failure visibility, notebook mode, 3-col board, project-aware pipeline |
| WS2 | Brain Project-Kanban P1-4b | Apr 4-5 | Projects, kanban, auto-assignment, AI push-back. **4c still in flight** |

---

## Deferred / Paused

| Item | Revisit When |
|------|--------------|
| Brain UX QoL Phase 8 (auto-commit) | After human-in-loop signal stabilizes |
| Plan 17: Proactive Surfacing | WS2 Phase 3 |
| Plan 18: Widget Overhaul (Ph 3-4) | Agent infra proves out |
| Plan 19: Brain App Ideas | Natural pause |
