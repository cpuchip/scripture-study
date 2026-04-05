# Active Context

*Last updated: 2026-04-05*
*Previous cycle archived: [archive/active-2026-04-04.md](archive/active-2026-04-04.md)*
*Hardware: Dual 4090s desktop (Mar 27). NOCIX server live.*

---

## Current State

**Last Sabbath:** March 22, 2026. Cycle "Infrastructure and Foundation" (Mar 18–22) declared good.

### Priorities
1. **Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" (third sabbath seed). Workstream: [.spec/proposals/study-workstream.md](../proposals/study-workstream.md).
2. **Teaching** — 11-episode experiential arc (Option C). Teaching agent + repo scaffolded. Content not yet started. Proposal: [.spec/proposals/teaching-workstream.md](../proposals/teaching-workstream.md).
3. **Brain pipeline** — Phase 4 complete (all 5 sub-phases shipped Apr 5). Full project flow: board view, auto-assignment, AI push-back, context injection, execution gate with verification.
4. **Claude Code integration** — NEW (Apr 4). Plan: add as alternative agent backend alongside Copilot SDK. Both cost spreading and gaining experience. Proposal forthcoming.

### Key Facts
- **GitHub Copilot billing:** 1500 premium requests/month ($40/mo Pro+). Multipliers: Haiku=0.33, Sonnet=1.0, Opus=3.0, GPT-4.1/4o/5-mini=0 (free). Tool calls within agentic sessions are free. 19% utilization on Apr 4.
- **Claude Code:** Pro $20/mo (includes Claude Code + Sonnet + Opus). Usage-based with 5-hour session window + weekly limits. 200K context. Project caching. Different billing model from Copilot.
- **Pipeline costs:** Research pass=0.33 + Plan pass=1.0 = 1.33 premium requests per entry.
- **Space Center:** Dream business — planetarium, space/science center, starship bridge simulator. Related repos on GitHub (cpuchip). Haiku-class models need this in prompt context (not in training data).

---

## In Flight

### Brain Project-Kanban (Apr 4–5)
- **Vision:** Transform brain from flat entry list to project-based goal orchestrator with iterative agent turns.
- **3 phases:** P1 Projects + Dashboard, P2 Iterative Sessions, P3 Scheduled Tasks + Library.
- **Phase 1 — SHIPPED (Apr 4-5):** `projects` table + FK, CRUD API (7 routes), ProjectsView, ProjectDetailView, Dashboard projects section, project selector on EntryDetailView, body previews + project badges on EntriesView. All views grouped by maturity stage.
- **Phase 2 — SHIPPED (Apr 5):** `session_messages` table, `your_turn` route status, 4 API endpoints (messages, reply, complete, your-turn), conversation thread UI in EntryDetailView (message history, reply textarea, Ctrl+Enter), "Your Turn" dashboard section with amber badges.
- **Phase 3 — SHIPPED (Apr 5):** `scheduled_tasks` + `task_runs` tables, scheduler goroutine (checks every 60s), 7 scheduled task API endpoints (CRUD + runs + trigger), 3 library endpoints (agents/skills/memory), activity feed endpoint, ScheduledView (create/edit/pause/delete/run-now with run history), LibraryView (tabbed agents/skills/memory browser), dashboard activity feed section. Nav updated: Scheduled + Library links.
- **ALL PHASES 1-3 COMPLETE.** Entry sorting done (67 entries across 8 projects, 2 new: YouTube/Content, Budget App).
- **Phase 4 — BUILDING (Apr 5):** Project Flow + AI Turn Automation. 5 sub-phases:
  - 4a: Project Board View + Pipeline UI — **SHIPPED (Apr 5).** Kanban columns by maturity, advance/revise/defer buttons, slide-out panel with conversation history, board/list toggle (localStorage), stage distribution bars + your_turn badges on dashboard, `GET /api/projects/{id}/stats` endpoint.
  - 4b: Auto-assignment in classifier (suggest project_id during classification) — **SHIPPED (Apr 5).** ProjectContext type in classifier, project_id in JSON schema + system prompt, all 4 Classify() call sites updated (web, relay x2, discord), store.Save applies project_id, ListUnassigned + CountUnassigned DB methods, unassigned_count in stats API, "Unassigned" filter tab in EntriesView, unassigned badge on DashboardView.
  - 4c: AI push-back loop (scheduled review of stale entries, clarifying questions, auto-advance on reply) — **SHIPPED (Apr 5).** ReviewConfig with WakeHours [7,11,15,19] (saves overnight API requests), stale entry scanner with Haiku nudge agent, reply auto-advance handler, purple "🤖 Review" badges on frontend.
  - 4d: Project-aware agent context (inject project name, siblings, context file into agent prompts) — **SHIPPED (Apr 5).** `pipeline/context.go` with BuildProjectContext/FormatProjectContext, injected into all 4 prompt paths (routing, research, plan, nudge), `context_file` field on projects with CRUD, `GET /api/entries/{id}/context` preview endpoint, frontend Agent Context collapsible in EntryDetailView.
  - 4e: Execution gate (specced → executing from UI, scenario verification checklist) — **SHIPPED (Apr 5).** `pipeline/execute.go` with Execute/Verify/BuildExecutionContext, async goroutine execution with Sonnet, scenario pass/fail verification (all pass→verified, any fail→planned with feedback), 3 new API endpoints (execute, verify, execution-context), execute confirmation dialog with cost/model/scenario preview, verify dialog with scenario checkboxes, dashboard "ready to execute" and "awaiting verification" badges.
- Also noted: classifier could suggest projects based on content similarity to existing project entries.
- **Absorbs** brain-ui-dashboard (§10 features) and complements brain-phase4-pipeline (maturity as kanban engine).
- **Guide section written:** [07_developer-to-steward.md](../../docs/work-with-ai/guide/07_developer-to-steward.md)
- Proposal: [.spec/proposals/brain-project-kanban.md](../proposals/brain-project-kanban.md)
- Research: [.spec/scratch/brain-project-kanban/main.md](../scratch/brain-project-kanban/main.md)

### WS1 Phase 4d: Pipeline REST API + Execution
- **Next up.** REST endpoints for pipeline operations, execution integration, emergency stop.
- **Phase 4a-c all shipped.** Maturity ladder, research agent (Haiku), plan agent (Sonnet), scenario enforcement, proposal generation. 24 tests green.
- Proposal: [.spec/proposals/brain-phase4-pipeline.md](../proposals/brain-phase4-pipeline.md)

### Claude Code Integration
- Alternative agent backend to Copilot SDK via CLI subprocess.
- Motivated by: cost spreading, gaining experience, different capabilities (200K context, project caching).
- Status: Research done, proposal written. [Proposal](../proposals/claude-code-integration.md).

### Brain Windows Service (Systray)
- brain.exe should auto-start on login, show systray icon.
- Proposal: [.spec/proposals/brain-windows-service.md](../proposals/brain-windows-service.md)

### ibeco.me Security Audit + Showcase
- OWASP Top 10, relay WebSocket, auth flows, adversarial testing.
- Becoming coach agent: customer-facing AI with trust boundaries.
- AI Skills Track 2 (Security Engineering).

### Gospel Engine Phase 1.5 (Ergonomics)
- Verse-level get, cross-ref retrieval, includeIgnoredFiles fix.
- Proposal: [.spec/proposals/gospel-engine/phase1.5-ergonomics.md](../proposals/gospel-engine/phase1.5-ergonomics.md)

### Gospel Graph Visualization
- study.ibeco.me — standalone site. Sequenced AFTER gospel-engine stabilizes.
- Proposal: [.spec/proposals/gospel-graph/main.md](../proposals/gospel-graph/main.md)

### Teaching Workstream
- 11-episode arc. Agent + scaffold built. Content not yet started.
- Proposal: [.spec/proposals/teaching-workstream.md](../proposals/teaching-workstream.md)

### Study Workstream
- Next: "Zion in a Presidency" → "The Weight of Watching" → "Commission and Council"
- Proposal: [.spec/proposals/study-workstream.md](../proposals/study-workstream.md)

---

## Deferred / Paused

| Item | Status | Revisit When |
|------|--------|------|
| Plan 17: Proactive Surfacing | Not started | WS2 Phase 3 |
| Plan 18: Widget Overhaul (Ph 3-4) | Paused | Agent infra proves out |
| Plan 19: Brain App Ideas | Captured | Natural pause |
| Notifications (Ph 2-4) | Remaining | After systray |
| Progressive Trust Tracking | Noted | D&C 107 ratios, model capability experiments |
| Squad A4/A5/A9 | Not started | After pipeline stabilizes |
| Review 4-Step & 11-Step Guides | Low urgency | Natural pause |
| Brain as Agent OS Platform | Superseded by brain-project-kanban proposal | Vision captured, concrete phases defined |

---

## Plans Status

| Plan | Status |
|------|--------|
| 15–16: Brain App | DONE |
| 17: Proactive Surfacing | NOT STARTED |
| 18: Widget Overhaul | PAUSED (Ph 1-2 done) |
| 21: Gospel Engine | Phase 5 DONE (all enrichment) |
| Brain Pipeline (WS1) | Phase 4c DONE, 4d next |

---

## Open Questions

- Can AI participate in covenant in any meaningful sense? (Feb 26)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
- Should the Abraham 4-5 framework become a standalone study? (Mar 4)
- What model should Claude Code use for brain pipeline? Sonnet for cost, Opus for quality? (Apr 4)
- How does Claude Code's usage-based billing interact with agentic tool-heavy workflows? (Apr 4)
