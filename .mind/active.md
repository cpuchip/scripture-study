# Active Context

*Updated: 2026-05-08 (pg-ai-stewards overnight session — **kimi-tuned AND qwen-tuned study prompts both validated end-to-end**. Five-run multi-model voice experiment on FtC/WtL binding question. Run #4 (kimi-tuned, corpus) caught+removed two fabricated quotes from its own draft. Run #5 (qwen-tuned, corpus) cleared all 12 targeted signatures, 54% shorter / 16% fewer tokens / 61% faster than the qwen-base baseline. **Local qwen3.6-27b on 4090 GPU now produces studies comparable in quality to kimi-k2.6 at zero variable cost.** Five phases shipped: 3c.3.3 (importer model_match), 3c.3.3.1 (agent_tool_perms `source` provenance column — broadcasts now survive frontmatter reimports), 3c.3.4 (4-way experiment), 3c.3.4.1 (qwen-3.6 variant + run #5). Both `.stewards/kimi-k2.6/` and `.stewards/qwen-3.6/` variants now stable v1. 3c.4 (gospel-engine HTTP tools) still daytime work — no pg_net in base image. Total experiment spend ~$0.35. Soak ran 7 passes through the night untouched. Comparison memo at `study/.scratch/two-triplets-comparison-2026-05-08/`. Future model variants follow a repeatable workflow: baseline run → identify signatures → author variant → re-run → validate.) · Previous: [archive/active-2026-04-23.md](archive/active-2026-04-23.md)*
*Hardware: Dual 4090s desktop. NOCIX server live.*

> **Edit rule:** Rewrite this file directly. Do NOT cat/append the old content first — its archive snapshot lives under `.mind/archive/`. Appending duplicates the document and doubles every memory load. (Bug: 2026-04-20.)

> **Workstream taxonomy:** [.mind/workstreams.md](workstreams.md) is canonical: WS1–WS9, status enum, frontmatter convention. Every active proposal carries `workstream:` frontmatter. Read that file before tagging anything new.

> **Model:** GitHub Copilot now runs Claude Opus 4.7 (was 4.6 through April). 4.7 is more literal — see Foresight & Adjacent Surfaces section in [.github/copilot-instructions.md](../.github/copilot-instructions.md). Adjacent Surface Audit is now standard before declaring dev/debug work complete.

---

## Priorities

1. ★ **[WS6] Study** — "It keeps me in the spirit." Latest: [broken-heart-and-contrite-spirit.md](../study/broken-heart-and-contrite-spirit.md) (May 1, fourth in arc: art-of-presidency + art-of-delegation + zion-in-a-presidency + broken-heart). Next: TBD → [study-workstream.md](../.spec/proposals/study-workstream.md)
2. ★ **[WS7] Teaching** — 11-episode arc (Option C). Agent + repo scaffolded. Content not started → [teaching-workstream.md](../.spec/proposals/teaching-workstream.md)
3. **[WS5] Token efficiency** — Apr 16 proposal awaiting refresh review → [token-efficiency.md](../.spec/proposals/token-efficiency.md)
4. **[WS2] Brain Inline Panel + Kanban 4c** — next active build → [brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md) · [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md)
5. **[WS5] pg-ai-stewards** — Postgres-as-AI-substrate. **Phases 2.5 + 2.6a/b/c + 2.7a + 3a + 3b + 2.7b.1 + 2.7b.2 + 2.7b.3 + 2.7b.4 shipped May 4–6.** **v0.2.0:** nine SQL files via `extension_sql_file!`. **2.7 stack complete:** trigger-driven harvest + scheduler tick (60s) + per-pass token budget + soak prep (frontmatter watchman-skip exemption + `regenerate_active_md()` markdown status report). docs/architecture.md (466-line reading map) shipped. Container has all SQL live-applied; image rebuild deferred. **Open:** soak start (bulk-tag journals + flip `schedule_enabled=true` for 7 days), agent corpus import (option A — Copilot/Claude agent defs into substrate), ws6 AGE upstream PRs, 2.8 (LLM-inferred edges), 2.9 (self-modification, gated on soak), 3c–3g. → [proposal](../.spec/proposals/pg-ai-stewards-phase-2-5-generic-substrate.md) · [phases.md](../projects/pg-ai-stewards/phases.md) · [docs/architecture.md](../projects/pg-ai-stewards/docs/architecture.md) · [AGE-QUIRKS.md](../projects/pg-ai-stewards/docs/AGE-QUIRKS.md)

## Key Facts

- Copilot: 1500 premium/mo ($40 Pro+). Haiku 4.5=0.33x, Sonnet 4.6=1.0x, Opus 4.7=7.5x, GPT-5/5-mini/4.1/4o=0
- Brain default model: gpt-5-mini (0x). Pipeline big = claude-opus-4.7 (7.5x)
- **Brain stage defaults (Apr 23 cost discipline):** research=sonnet, plan=opus, spec=sonnet, execute=sonnet, verify=haiku (hard-pinned), revise=sonnet. Commission `Model` field = steward judgment only (gate eval). Revise loop capped at 2 → surface.
- Claude Code: Pro $20/mo. 200K context. Project caching
- Pipeline cost: research=0.33 + plan=1.0 = 1.33/entry  *(stale — see stage defaults above)*
- Active MCP servers: gospel-engine-v2 (engine.ibeco.me), webster, yt, byu-citations, becoming, exa-search
- gospel-engine v2 hosted is the single canonical search backend; gospel-mcp + gospel-vec retired as fallback

---

## In Flight

| WS | Item | Status | Ref |
|----|------|--------|-----|
| WS2 | Brain status-aware-views ecosystem parity (ibeco.me + brain-app) | ✅ Phases 1+2 shipped Apr 23; Phase 3 verify post-deploy | [brain-status-aware-views-ecosystem-parity.md](../.spec/proposals/brain-status-aware-views-ecosystem-parity.md) |
| WS2 | Brain non-pipeline kanban flow | ✅ archived Apr 23 (all phases shipped or scoped out) | [archive/brain-non-pipeline-kanban-flow.md](../.spec/proposals/archive/brain-non-pipeline-kanban-flow.md) |
| WS2 | Brain Inline Panel + Nudge Bot Controls | ▶ P1 next | [brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md) |
| WS2 | Brain Project-Kanban Phase 4c | ▶ next | [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md) |
| WS5 | Token Efficiency & Memory v2 | ⏸ awaiting refresh | [token-efficiency.md](../.spec/proposals/token-efficiency.md) |
| WS5 | Brain ↔ VS Code Bridge | � building (Phase 0 shipped Apr 22) | [brain-vscode-bridge/main.md](../.spec/proposals/brain-vscode-bridge/main.md) |
| WS2 | Brain non-pipeline projects | ✅ archived Apr 23 | [archive/brain-non-pipeline-projects.md](../.spec/proposals/archive/brain-non-pipeline-projects.md) |
| WS2 | Brain manual stage transitions | ✅ archived Apr 23 | [archive/brain-manual-stage-transitions.md](../.spec/proposals/archive/brain-manual-stage-transitions.md) |
| WS2 | Johari window agent mode | 📝 proposed Apr 22 | [johari-window-agent-mode.md](../.spec/proposals/johari-window-agent-mode.md) |
| WS5 | pg-ai-stewards (Postgres substrate for agent state, memory, work, model calls) | 🔨 **3c.3.2 ran end-to-end May 7 — substrate produced a real meta-study** (`study/two-triplets-one-ascent.md`). **Watchman soak started May 7 ~16:30 CDT** — `schedule_enabled=true`, dirty_queue 269, will fire pressure passes hourly until corpus stabilizes. | [proposal](../.spec/proposals/pg-ai-stewards-phase-2-5-generic-substrate.md) |
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
| WS2 | brain-steward-cost-discipline | Apr 23 | Three-defect fix: (1) commission `Model` no longer overrides every stage — `modelForStage` helper routes through `config.StageDefaults`; `c.Model` reserved for `EvaluateGate` only. (2) Verify hard-pinned to haiku regardless of catalog. (3) Revise loop capped at 2 → surface `loop_limit_exceeded`. `RevisionCount` field on Commission with DB migration; "Revised X/2" badge on EntryDetailView. Same fix applied to `commissionWaitForExecution` (the loop that actually burned 105 credits). Best-case opus commission ~25 credits (was ~52); worst case ~28-35 (was unbounded, hit 105). Followup Apr 23: research bumped haiku→sonnet per Michael ("stronger model researching is good") — chain still escalates from sonnet, just one fewer step before quarantine. |
| WS2 | brain-model-catalog-sot | Apr 23 | Single source of truth at `internal/config/models.go`: `Catalog` slice + `StageDefaults` map. Two pre-existing drifted maps (`modelCosts`, `AvailableModels`) and `steward.EscalationChain` now derive from it. New `GET /api/models` endpoint; frontend composable + dynamic dropdowns in CommissionDialog and ProjectDetailView inline commission. Default is now Claude Opus 4.7 (7.5×) instead of stale Opus 4.6 (3.0×). Stewardship sweep: same-bug-same-fix on `feedbackDialog` and `executeDialog` modals (UA-stylesheet dark-theme bug from doneDialog fix earlier same day). Note: `/api/models` path collided with legacy LM Studio profiles handler — moved to `/api/models/profiles` (no consumers found). |
| WS2 | brain non-pipeline kanban flow (Phases 1-3 + classify gate) | Apr 23 | Status vocab gained `working` (in-progress lane). `boardColumns` branches on `pipeline_enabled`: manual path uses literal status keys (active/working/done) instead of route-status pipeline. 5-button manual row (▶ Start / ✓ Done / ↩ Reopen / ⏸ Someday / 🗄 Archive) on board + list views. Optional reason dialog on ✓ Done appends a `_Closed YYYY-MM-DD: reason_` line to body. Native HTML5 drag-and-drop between columns (no library dep). Auto-classify gated for non-pipeline projects in relay client. Done-dialog theming fix (centered, larger, dark-theme textarea). `handleCreateEntry` now accepts `project_id` — single-call POST round-trips correctly. brain-app P4 mirror skipped per user direction (no Project surface). |
| WS2 | brain non-pipeline projects (Phases 1+2) | Apr 23 | `pipeline_enabled` column on `projects` (default true). Notebook auto-flipped to false. Single primary gate in `routeEntry`; defense-in-depth in `BuildProjectContext`. UI checkbox + 📓 badge on web. Inverse-hypothesis verified via PUT-set project_id + explicit `/api/agent/route`. |
| WS2 | brain manual stage transitions (Phase 1 mobile) | Apr 23 | brain-app: status dropdown in edit screen, long-press quick-actions sheet (done/park/waiting/archive/reactivate) + undo, semantic chip colors. Phase 2 web: pre-existing — `EntryDetailView.vue` already had full STATUS_OPTIONS dropdown. |
| WS2 | brain status-aware-views ecosystem parity (Phases 1+2) | Apr 23 | ibeco.me `ListBrainEntries` honors `?include_parked=1` (default off). brain-app history `_showParked` covers both someday+archived. Phase 3 cross-surface check pending Dokploy deploy. |
| WS5 | Opus 4.7 harness tuning | Apr 23 | Foresight & Adjacent Surfaces section in copilot-instructions.md, dev.agent.md update, Council Moment extended to dev/debug/ux. Diagnosis + fix for literalism failures. |
| WS2 | brain-status-aware-views (Phases 1-3 + Dashboard) | Apr 23 | All planned phases shipped. Phase 4 (Capture semantic) deferred. Server-side `/api/entries` filter (`?include_parked=1` opts in). Project board toggle relocated to visible header checkbox. Dashboard agent surfaces (routable/review/your-turn) filter parked server-side. |
| WS2 | brain-status-field-on-list-queries (full SELECT audit) | Apr 23 | All 6 list queries now expose `status`: ListAll, ListCategory, ListEntriesByProject, ListByRouteStatus, ListUnassigned, ListPipeline. Future filter UIs no longer guess data layer contracts. |
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
