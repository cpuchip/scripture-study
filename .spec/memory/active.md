# Active Context

*Last updated: 2026-04-10 (Pipeline E2E test complete — LCARS entry walked raw→verified→done)*
*Previous cycle archived: [archive/active-2026-04-04.md](archive/active-2026-04-04.md)*
*Hardware: Dual 4090s desktop (Mar 27). NOCIX server live.*

---

## Current State

**Last Sabbath:** March 22, 2026. Cycle "Infrastructure and Foundation" (Mar 18–22) declared good.
**Project Sync:** Apr 5 — brain entries synced with spec/plan files. 101 total entries. 30 verified, 10 specced, 4 planned, 57 raw.
**KISS Reflection:** Apr 5 — honest assessment: pipeline is sophisticated but overserves 90% of entries. Both simplified AND automated paths needed. Scratch notes at [.spec/scratch/brain-simplification/main.md](../scratch/brain-simplification/main.md).

### Priorities
1. **Study** — "It keeps me in the spirit." Next: "Zion in a Presidency" (third sabbath seed). Workstream: [.spec/proposals/study-workstream.md](../proposals/study-workstream.md).
2. **Teaching** — 11-episode experiential arc (Option C). Teaching agent + repo scaffolded. Content not yet started. Proposal: [.spec/proposals/teaching-workstream.md](../proposals/teaching-workstream.md).
3. **Brain pipeline** — Phase 4 complete. Pipeline fixes (execute reliability + human gate UI) ALL SHIPPED (Apr 9-10). Execution stall fixed, 10-min timeout, cancel endpoint, toast system, scenario dialogs, undo-complete, progress streaming. Ready for end-to-end testing. Next: simplification + inline panel + nudge bot controls + UX quality-of-life.
4. **Space Center pipeline test** — Practice automated pipeline on Space Center project as low-stakes test bed. First entry (display dashboard) already revealed the UX gaps. Plan: [.spec/scratch/space-center-pipeline-test/main.md](../scratch/space-center-pipeline-test/main.md).

### Key Decisions (Apr 5 Session)
- **Both paths wanted.** Simplified workflow (notebook, 3 columns) AND fully automated pipeline (auto-continuation). Not one or the other.
- **KISS for captures, power for delegation.** 90% of entries are simple captures. 10% are delegation. Design for both.
- **Nudge bot needs controls.** Must be visible in Scheduled Tasks, pausable, transparent. Currently invisible hardcoded goroutine.
- **Board simplification: 3 columns.** Inbox / Working / Done. Sub-stage badges instead of separate columns.
- **Space Center as test bed.** 5 seed entries, observe full automated cycle end-to-end.
- **"By small and simple things."** Stop building infrastructure ahead of use case. Build, use, then iterate.

### Key Facts
- **GitHub Copilot billing:** 1500 premium requests/month ($40/mo Pro+). Multipliers: Haiku=0.33, Sonnet=1.0, Opus=3.0, GPT-4.1/4o/5-mini=0 (free). Tool calls within agentic sessions are free. 19% utilization on Apr 4.
- **Claude Code:** Pro $20/mo (includes Claude Code + Sonnet + Opus). Usage-based with 5-hour session window + weekly limits. 200K context. Project caching. Different billing model from Copilot.
- **Pipeline costs:** Research pass=0.33 + Plan pass=1.0 = 1.33 premium requests per entry. Nudge bot: 0.33 per entry, up to 10 entries 4x/day = 13.2/day worst case.
- **Space Center:** Dream business — planetarium, space/science center, starship bridge simulator. Related repos on GitHub (cpuchip). Haiku-class models need this in prompt context (not in training data).
- **Review nudge bot:** Fires at [7,11,15,19] hours, uses Haiku, scans up to 10 stale entries, not visible in UI, not pausable. Creates VS Code sidebar sessions per nudge.

---

## In Flight

### Brain Inline Panel + Nudge Bot Controls
- **Phase 1:** Reply textarea + close-with-reason in slide-out panel. Self-contained, one session build.
- **Phase 2:** Surface nudge bot in Scheduled Tasks tab with pause/resume and run history.
- Proposal: [.spec/proposals/brain-inline-panel.md](../proposals/brain-inline-panel.md)

### WS3: Brain UX Quality-of-Life (from real usage)
- **Phase 1 (DONE):** Auto-expanding textarea, markdown rendering in messages, clickable file paths, inline file viewer sidebar panel, content shift when panel open. External links open in new tabs. Backslash path normalization for Windows.
- **Phase 5 (DONE):** Smarter auto-advance messages — extractQuestionSummary reads scratch file, counts questions, lists categories.
- **Phase 2 (DONE):** File browser in Library tab — recursive tree endpoint, TreeNode component, search filter, wide layout. Verified Apr 6.
- **Phase 3 (DONE):** WebSocket push updates — Hub broadcasts entry.updated/message.new/entry.created. Dashboard, EntryDetailView, ProjectDetailView all receive live updates. Auto-reconnect with exponential backoff. Verified Apr 6.
- **Phase 4 (DONE):** Cost tracking per entry — `premium_requests_used` column, IncrementPremiumRequests after each pipeline agent call (research 0.33, plan 1.0, execute 1.0, nudge 0.33), badge in EntryDetailView, aggregate in ProjectDetailView. Shipped Apr 6.
- **Phase 6 (DONE):** Reader UX — backtick/code link fix (added `>` and backtick to FILE_PATH_RE lookbehind), internal link following in Library + FileViewer, route deep linking (`/library?file=path`), "Open in Reader" button in FileViewer, back/forward navigation history with full history stack. Shipped Apr 6.
- **Phase 7 (DONE):** Git status in file browser — `GET /api/git/status` endpoint (runs `git status --porcelain`, parses output), TreeNode status dots (green=new, yellow=modified, red=deleted), directories inherit most severe child status, clickable summary bar above tree with counts + filter-to-changed toggle. Refreshes on files tab activation. Shipped Apr 6.
- **Phase 7a (DONE):** Inline diff viewer — `GET /api/git/diff?path=` endpoint (path-safe, serves unified diff for tracked/untracked files), `diff2html` npm library with `ColorSchemeType.DARK`, "Δ Diff" / "✕ Diff" toggle in header bar (visible only for changed files), line-by-line / side-by-side mode toggle, lazy diff loading, resets on file navigation. Shipped Apr 6.
- **Phase 7b (DONE):** Nested git repo awareness — `discoverGitRepos()` walks workspace for `.git` dirs (13 found), `handleGitStatus` aggregates across all repos with `repo` field, `handleGitDiff` routes to correct repo via `findRepoForPath()`, `is_git_repo` flag on file tree nodes. Frontend shows ⎇ badge on repo directories. Shipped Apr 6.
- Proposal: [.spec/proposals/brain-ux-quality-of-life.md](../proposals/brain-ux-quality-of-life.md)
- Research: [.spec/scratch/brain-ux-quality-of-life/main.md](../scratch/brain-ux-quality-of-life/main.md)

### WS4: Brain Pipeline Evolution (from creation cycle gap analysis)
- **Graduated** from scratch research to full proposal on Apr 6.
- 11-step creation cycle gap analysis: Steps 2, 8, 9, 10, 11 now have specced phases.
- 7 phases: Governance Docs, Failure Visibility, Reflection Pauses + Auto-Continue, Notebook Mode, Nudge Bot Controls, 3-Column Board, Project Scaffolding.
- **Phase 1 (DONE):** Governance Documents — updated `research-covenant.md` (added Boundaries + Budget), updated `plan-covenant.md` (added Boundaries, Budget, "Who Benefits?" consecration check, "How Does This Integrate?" Zion check), created `execute-covenant.md` (full 11-step cycle mapping), created `review-covenant.md` (Steps 7-9: Review, Atonement, Sabbath). No code changes. Brain restarts with zero governance warnings. Shipped Apr 6.
- **Phase 2 (DONE):** Failure Visibility — `failure_count` + `last_failure_reason` columns via SQLite migration, `IncrementFailureCount`/`ResetFailureCount` DB methods, `recordFailure` pipeline helper posts session messages with recovery options on any advance/revise/execute failure, escalation message at 3+ consecutive failures, failure count reset on success, red 🔴 badge in EntryDetailView with tooltip showing last failure reason. Shipped Apr 6.
- **Phase 3 (DONE):** Reflection Pauses + Auto-Continue — `auto_continue BOOLEAN DEFAULT FALSE` column via SQLite migration, `SetAutoContinue` DB method, `PUT /api/entries/{id}/auto-continue` endpoint. Sabbath path (default): `route_status = "your_turn"` set after research and plan complete, with session messages prompting review. Delegation path (`auto_continue = true`): `maybeAutoContinue` goroutine fires 2s after advance, auto-advances through researched→planned stages, always stops before verification. Sabbath prompt: verification success message includes "What worked well? What would you do differently?" reflection. Frontend: checkbox toggle in entry detail header (⚡ Auto / 🕊️ Sabbath). Shipped Apr 6.
- **Phase 4 (DONE):** Notebook Mode — `notebook BOOLEAN DEFAULT FALSE` column via SQLite migration, `SetNotebook`/`BulkSetNotebook` DB methods, `PUT /api/entries/{id}/notebook` + `POST /api/entries/bulk-notebook` endpoints. Notebook entries: searchable, taggable, appear in project views, but excluded from `ListStaleEntries` (no nudging), `ListByRouteStatus` (no review queue), `GetProjectStats` pipeline counts. `Advance()` and `Execute()` reject notebook entries. Frontend: 📓/🔄 toggle in EntryDetailView (hides auto-continue toggle + pipeline controls when active), notebook badge in EntriesView with multi-select + bulk "Move to Notebook"/"Back to Pipeline" floating action bar, "Save as notebook" checkbox in CaptureView. Shipped Apr 6.
- **Phase 5 (DONE):** Nudge Bot Controls — `reviewState` struct with mutex-protected fields (paused, lastRunAt, nextRunAt, nudge counts, cost, lastActivityAt) added to Pipeline. `GetReviewStatus()`/`SetReviewPaused()`/`TouchActivity()` public methods. `GET /api/nudge-bot/status` + `PUT /api/nudge-bot/pause` endpoints. Presence-aware: CORS middleware calls `TouchActivity()` on every request; `runReviewScan` skips if no activity in 2 hours. `nudge_count INTEGER DEFAULT 0` column via migration, `IncrementNudgeCount` called per nudge. Frontend: nudge bot status card in ScheduledView (active/paused badge, presence indicator, wake hours, last/next run, nudge counts, cost), pause/resume toggle. 🔔 nudge count badge in EntryDetailView metadata. Shipped Apr 6.
- **Phase 6 (DONE):** 3-Column Board — Replaced 6-column maturity board in ProjectDetailView with 3 columns: Inbox (raw/unset/notebook), Working (researched/planned/specced/executing), Done (verified). `boardColumns` computed groups entries, `grid-cols-3` layout replaces horizontal scroll. Working entries show maturity sub-stage badge, notebook entries show 📓 badge in Inbox. Removed unused `maturityBorderColor` and `boardStages`. Pure frontend change, ~0 backend. Shipped Apr 6.
- **Phase 7a-7b-7f (DONE):** Schema + Governance Injection + Context Cap — (7a) Added `workspace_type`, `workspace_path`, `github_repo`, `repo_visibility` columns to projects table via `migrateProjectWorkspace()`. Updated CreateProject, GetProject, ListProjects, UpdateProject. Frontend: Project interface extended, edit form shows workspace type dropdown, conditional workspace path + GitHub fields for subfolder/external types. (7b) Added `loadBaseInstructions()` + `trimBaseInstructions()` that loads and trims `.github/copilot-instructions.md` (strips MCP tables, agent modes, session memory — keeps voice, covenant, principles). Injected as Layer 0 in research, plan, execute, and review agent system messages. Wired `review-covenant.md` into nudge bot via `buildNudgeSystemMessage()`. (7f) Raised context file cap from 3000→8000 chars. Shipped Apr 6.
- Phase 7c-7d-7e (DONE) — External project scaffolding (`ScaffoldProject` creates dirs, README, thin copilot-instructions.md, git init, initial commit, optional `gh repo create`), `resolveWorkDir` routes external projects to their own directory (falls back to workspace root if dir missing), `POST /api/projects/{id}/scaffold` endpoint, frontend "Initialize" button with result feedback. All 4 pipeline agents (research, plan, execute, review) now use `resolveWorkDir(entry)` for `WorkingDir`. Phase 7 complete. Shipped Apr 7.
- **Phase 8 (DONE):** Agent-Driven Project Initialization — `InitializeProject` uses Copilot SDK agent to scaffold project dirs, README, copilot-instructions based on entry descriptions. Mechanical fallback. v-else build fix applied. Shipped Apr 7.
- **Phase 9 (DONE):** Project-Aware Pipeline — (9a) `FormatProjectContext` includes workspace_type+workspace_path so agents know where the project dir is, (9b) scratch/proposal paths use `projectRelPath()` to scope files under project dir for external projects, (9c) selective post-execution git commit — `Agent.writtenFiles` map tracks files via PostToolUse hook + `isWriteTool()` + `extractPathCandidates()`, `commitAfterExecution()` groups files by git repo via `findGitRoot()`, `gitCommitSelective()` stages only tracked files (never `git add -A`), Haiku generates commit messages with `brain({slug}):` prefix (fallback to mechanical on failure). All tests pass. Shipped Apr 7.
- Proposal: [.spec/proposals/project-aware-pipeline.md](../proposals/project-aware-pipeline.md)
- Research: [.spec/scratch/project-aware-pipeline/main.md](../scratch/project-aware-pipeline/main.md)
- Proposal: [.spec/proposals/brain-pipeline-evolution.md](../proposals/brain-pipeline-evolution.md)
- Research: [.spec/scratch/brain-pipeline-evolution/main.md](../scratch/brain-pipeline-evolution/main.md)
- Prior research: [.spec/scratch/brain-simplification/main.md](../scratch/brain-simplification/main.md)

### Space Center Pipeline Test
- 5 seed entries, observe fully automated pipeline end-to-end
- Low-stakes test bed for the delegation workflow
- **E2E test completed (Apr 10):** LCARS Vue3 Theme entry walked through full pipeline raw→verified→done.
  - Execution timed out at 10 min but agent created 35 source files, 20 passing tests, both apps build.
  - Manually advanced past timeout, verified all 6 scenarios (clock, calculator, LCARS styling, RPN stack, localStorage persistence).
  - 12 UX friction points documented in [debug-pipeline-e2e/main.md](../scratch/debug-pipeline-e2e/main.md).
  - Key issues: 10-min timeout too short, no liveness indicator during long generation, failure badge doesn't reset on verify.
- Plan: [.spec/scratch/space-center-pipeline-test/main.md](../scratch/space-center-pipeline-test/main.md)

### Brain Pipeline Fixes — ALL PHASES COMPLETE (Apr 9-10)
- **Phase 1 (DONE):** Backend — slim execute prompt (path not content), 10-min timeout, cancel endpoint, route_status="agent" during execution, premium cost before Ask(), race guard, handleMarkComplete fix.
- **Phase 2 (DONE):** Frontend — toast system replacing all alert() calls, scenario input dialog (planned→specced), cancel/complete/verify buttons everywhere (board/list/panel/detail), maturity badges, pipeline gate sections.
- **Phase 3 (DONE):** Polish — OnToolCall callback streams execution tool events to WebSocket, Pipeline/Notebook toggle button group replaces confusing checkbox.
- **Phase 3.5 (DONE):** Two-completes disambiguation — circle checkbox renamed "Mark done"/"Reopen", conversation Complete→Dismiss (calls dismissRoute instead of pipeline complete), route_status "complete" badge→"Routed", ↩ Undo pipeline complete (reverts to verified).
- **Ready for end-to-end testing.** ✅ TESTED (Apr 10). LCARS entry walked through full pipeline. Timeout hit at 10 min but work was complete. 12 friction points documented. See [debug-pipeline-e2e/main.md](../scratch/debug-pipeline-e2e/main.md).
- **Next:** Fix BUG-3 (verify doesn't reset failure count), increase timeout, add liveness indicator.
- Proposal: [.spec/proposals/brain-pipeline-fixes.md](../proposals/brain-pipeline-fixes.md)

### Session-First Flow (Exploring)
- Every entry becomes a session instead of going through classification stages
- Research notes: [.spec/scratch/session-first-flow/main.md](../scratch/session-first-flow/main.md)
- Idea #5 in Plan 19

### Brain Project-Kanban — ALL PHASES COMPLETE (Apr 4–5)
- **Vision:** Transform brain from flat entry list to project-based goal orchestrator with iterative agent turns.
- **Phases 1-4 ALL SHIPPED (Apr 5).** Projects, iterative sessions, scheduled tasks, library, kanban board, auto-assignment, AI push-back, context injection, execution gate with verification.
- **Project Sync Audit (Apr 5):** Cross-referenced all spec/plan files against brain entries. Created 22 new entries, marked 22 existing entries as verified, deleted 8 duplicates. Fixed ListAll/ListCategory/ListUnassigned SQL queries to include maturity column. Enhanced handleUpdateEntry to support maturity + route_status updates.
- **Final state:** 101 entries (30 verified, 10 specced, 4 planned, 57 raw). 8 plans + 12 proposals archived.
- **Absorbs** brain-ui-dashboard and complements brain-phase4-pipeline.
- **Guide section written:** [07_developer-to-steward.md](../../docs/work-with-ai/guide/07_developer-to-steward.md)
- Proposal: [.spec/proposals/brain-project-kanban.md](../proposals/brain-project-kanban.md)
- Cross-ref audit: [.spec/scratch/project-sync/main.md](../scratch/project-sync/main.md)

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
| 02: Layout | Active |
| 05: Tool Todo | Active |
| 07: Scheduled Tasks | Active |
| 16: Brain App Today | Active |
| 17: Proactive Surfacing | NOT STARTED |
| 18: Widget Overhaul | PAUSED (Ph 1-2 done) |
| 20: NOCIX Dokploy | Active |
| 21: Gospel Engine | ARCHIVED (Phase 5 done) |
| Brain Pipeline (WS1) | Phase 4c DONE, 4d next |
| Brain Pipeline Fixes | ALL PHASES COMPLETE (Apr 9-10) |
| Brain UX (WS3) | Phase 1✅ 5✅ 2✅ 3✅, Phase 4 next |
| Brain Pipeline Evolution (WS4) | Phases 1-3 shipped, Phases 4-7 specced |
| Brain Project-Kanban | ALL PHASES COMPLETE |

*Archived (Apr 5):* Plans 03, 06, 08, 09, 10, 11, 15, 21. Proposals: brain-memory, brain-phase3c-sdk-agents, brain-relay, brain-ui-dashboard, context-engineering, context-engineering-dev, enriched-search, session-journal, squad-learnings, brain-multi-agent, notifications, brain-unified-dashboard.

---

## Open Questions

- Can AI participate in covenant in any meaningful sense? (Feb 26)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
- Should the Abraham 4-5 framework become a standalone study? (Mar 4)
- What model should Claude Code use for brain pipeline? Sonnet for cost, Opus for quality? (Apr 4)
- How does Claude Code's usage-based billing interact with agentic tool-heavy workflows? (Apr 4)
- Does the review nudge bot actually move things along, or is it just noise? (Apr 5) — test via Space Center
- What's the right threshold for auto-continuation? Run until stuck? Run N stages max? (Apr 5)
- Should notebook entries be a category or a flag on any entry? (Apr 5)
- Is the inline file viewer (Phase 1) sufficient, or do we need a full file browser (Phase 2)? (Apr 5) — use it first, then decide
- Should WebSocket events replace polling entirely, or coexist as fallback? (Apr 5)
