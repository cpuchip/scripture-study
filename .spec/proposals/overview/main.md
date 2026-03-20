# Project Overview — Unified Workstream Plan

**Binding problem:** Too many plans (28+), no unified execution strategy, one person's capacity as bottleneck. Need parallel workstreams that agents can execute against, organized by the 11-step creation cycle.

**Created:** 2026-03-12
**Research:** [.spec/scratch/overview/main.md](../../scratch/overview/main.md)
**Guidance questions:** [guidance.md](guidance.md) — ALL ANSWERED (Mar 19)
**Status:** Decisions recorded. Ready to execute.

---

## 1. Problem Statement

Michael has built a remarkable ecosystem: 6 MCP servers, a brain capture system, a becoming/practice-tracking app, a Flutter mobile app, a TTS pipeline, a publishing system, and a deep scripture study practice — all in ~50 days.

But the plans have outpaced the execution capacity. There are:
- **19 numbered plans** in `scripts/plans/` (6 done, 4 in progress, 9 designed-but-waiting)
- **9 formal proposals** in `.spec/proposals/` (2 implemented, 3 in progress, 4 waiting)
- **5+ doc-level roadmaps** scattered across `docs/`
- **1 multi-agent vision** in `study/ai/multi-agent-ideas.md`
- **Multiple unresolved architectural questions** (Garvis vs. brain.exe, auth scope, storage decisions)

The daily experience is: start a session → remember there are 5 things to do → pick one → get excited about a 6th → end session with 6 things to do. Mosiah 4:27 isn't being practiced at the project management level.

### Success Criteria

1. **Single source of truth** — One map showing all work, its status, and its dependencies
2. **Executable workstreams** — 2-3 lanes of work that can run simultaneously
3. **Agent-ready specs** — Each workstream's Phase 1 has a spec precise enough for an agent to execute
4. **Reduced in-flight count** — Defer or archive work that isn't on the critical path
5. **Copilot SDK operational** — At least one proof-of-concept showing agent-driven execution

---

## 2. Prior Art & What Exists

| Asset | Location | Status |
|-------|----------|--------|
| Q1 Roadmap | [scripts/plans/archive/14_roadmap-2026-q1.md](../../../scripts/plans/archive/14_roadmap-2026-q1.md) | Archived — superseded by this document |
| Multi-agent ideas | [study/ai/multi-agent-ideas.md](../../../study/ai/multi-agent-ideas.md) | Vision doc, not spec |
| 11-step creation cycle | [docs/work-with-ai/guide/05_complete-cycle.md](../../../docs/work-with-ai/guide/05_complete-cycle.md) | Framework — our foundation |
| Spec engineering guide | [docs/work-with-ai/guide/04_spec-engineering.md](../../../docs/work-with-ai/guide/04_spec-engineering.md) | 5 primitives — our spec language |
| Intent | [intent.yaml](../../../intent.yaml) | Root values — all work must trace here |
| Active state | [.spec/memory/active.md](../../memory/active.md) | Current session context |

---

## 3. Proposed Workstreams

### Workstream 1: Agentic Foundation

**Intent:** Enable agents to execute work autonomously. This is the multiplier that makes everything else faster.

**Decision (Mar 19):** Option C confirmed — front-load agentic infrastructure, then fan out. Progressive trust model: start supervised, expand as confidence grows. VS Code hooks (v1.111) for chaining specs with limited premium requests is the near-term execution model.

**AI Backend Strategy (Mar 19):** Dual-backend, role-separated. LM Studio (qwen3.5-9b on fermion/lepton's 4090s) for classification — trusted, tested, free. Copilot SDK (Opus 4.6 or Sonnet 4.6) for agent abilities — spec execution, reasoning, complex tasks.

#### Phase 1: Copilot SDK + MCP Integration — DONE (Mar 20)

**Note (git audit):** Copilot SDK is ALREADY integrated in brain.exe at v0.1.29. The `internal/ai/client.go` wraps the SDK as a configurable backend (`"copilot"` vs `"lmstudio"`). We're not starting from zero — we're extending what exists.

| Item | Detail |
|------|--------|
| **Task** | Extend brain.exe's existing Copilot SDK integration to connect gospel-mcp as an MCP tool |
| **Starting point** | `internal/ai/client.go` already has `copilot.NewClient()` → session management |
| **Add** | MCP tool registration so the agent can call gospel-mcp tools (gospel_search, etc.) |
| **Input** | "What does D&C 93:36 teach about intelligence?" |
| **Expected output** | Agent uses gospel_search, retrieves verse, provides contextual answer |
| **Verify** | Agent correctly cites the verse text. No confabulation. |

**What was built (Mar 20):**
- `internal/ai/agent.go` — `Agent` struct with `Ask`/`Reset`/`createSession`. Lazy session creation, conversational reuse, error-triggered session reset. MCP servers registered as stdio tools in Copilot SDK `SessionConfig`.
- `internal/config/config.go` — `MCPServerDef` type, `AgentModel`/`MCPServers` fields, auto-discovery of gospel-mcp/gospel-vec/webster-mcp binaries from sibling directories.
- `internal/web/server.go` — `POST /api/agent/ask` and `POST /api/agent/reset` endpoints. Nil-agent guard (503 when copilot backend not active).
- `internal/ai/client.go` — `CopilotClient()` getter to expose raw SDK client for agent sessions.
- `cmd/brain/main.go` — Agent creation wired: converts config MCPServerDefs → ai.MCPDefs, creates Agent when copilot backend + MCP servers both available.

**Remaining from original constraints:**
- ~~Build on existing `internal/ai/` package~~ ✅
- ~~Must connect to gospel-mcp as an MCP tool~~ ✅ (gospel-mcp, gospel-vec, webster-mcp all auto-discovered)
- ~~Must run locally~~ ✅
- Streaming output not yet implemented (batch response only) — deferred to Phase 2

#### Phase 2: Agent as Spec Executor (1-2 sessions)

| Item | Detail |
|------|--------|
| **Task** | Give the agent a spec file and have it execute against it |
| **Test spec** | One of the MCP improvement items from [docs/mcp-improvements.md](../../../docs/mcp-improvements.md) |
| **Verify** | Agent produces a PR-worthy diff. Human reviews. |

#### Phase 3: Multi-Agent Routing (2-3 sessions, after Phase 2 works)

| Item | Detail |
|------|--------|
| **Task** | brain.exe routes a captured idea to the appropriate agent session |
| **Pattern** | Capture → classify → if "spec-worthy" → create proposal skeleton → assign to agent |
| **Verify** | End-to-end: brain capture → proposal draft appears in `.spec/proposals/` |

### Workstream 2: Brain Consolidation

**Intent:** Make the brain ecosystem reliable and integrated before building on top of it.

**Decision (Mar 19 — Q1):** Garvis IS brain.exe. Merged conceptually. "Garvis" name retired. brain.exe is the second brain, evolved with SQLite + chromem-go + relay + MCP. No new repo.

**Deployment (Mar 19 — Q8):** Local first → dockerize → deploy to NOCIX server alongside ibeco.me. Sequential, not rushed.

#### Phase 1: Quick Wins (1 session)
- [Plan 15](../../../scripts/plans/15_brain-app-polish.md): Entry sync on launch, relay error recovery, classify flow polish, delete with undo
- These are 4 small fixes that improve daily experience immediately

#### Phase 2: Bidirectional Sync (1-2 sessions)
- [Brain unified dashboard Phase 4](../../proposals/brain-unified-dashboard.md): Last-write-wins conflict resolution across brain-app, brain web UI, ibeco.me
- This is the "data consistency" foundation everything else needs

#### Phase 3: Proactive Surfacing (1-2 sessions)
- [Plan 17](../../../scripts/plans/17_brain-proactive-surfacing.md) features 1-3: Due actions, stale people, stalled subtasks
- This is what makes brain *useful* beyond storage — it's the "why brain exists" feature

#### Phase 4: Server Deployment (after local proving)
- ~~Merge Garvis proposal into brain.exe evolution~~ DONE (decision: they're the same thing)
- Dockerize brain.exe
- Deploy to NOCIX server alongside ibeco.me

### Workstream 3: Becoming App + Study Quality

**Intent:** Improve the tools that serve the core mission (scripture study + personal becoming).

**Decision (Mar 19 — Q5):** Study is the HIGHEST priority — "it keeps me in the spirit." Agentic and study are the two priorities, running in parallel. Infrastructure serves study, not the other way around.

**Multi-user (Mar 19 — Q3):** ibeco.me IS multi-user. Google OAuth + email/password auth ALREADY DEPLOYED. Plan 09 is stale — auth is further along than it describes. webeco.me planned for families/groups.

**Widget (Mar 19 — Q6):** Plan 18 stays in main roadmap. Phases 3-4 paused (not deferred) until agent work is rolling.

**Storage (Mar 19 — Q7):** Brain uses local filesystem. ibeco.me uses S3 on the NOCIX server (3TB, unmetered 1Gbps).

#### Phase 1: Scheduled Tasks (1-2 sessions)
- [Plan 07](../../../scripts/plans/07_scheduled-tasks.md): Already fully designed. Extends practices to interval, weekly, daily_slots, monthly, one-time.
- Backend engine first → frontend forms → DailyView updates.

#### Phase 2: MCP Improvements (1-2 sessions)
- [docs/mcp-improvements.md](../../../docs/mcp-improvements.md): Priority 1-3 (markdown_link returns, full doc retrieval, preserved markdown)
- These directly improve every future study session

#### Phase 3: gospel-vec Experiments (1 focused session)
- [docs/model-experiments.md](../../../docs/model-experiments.md): Run the benchmark against 4-5 embedding models
- Choose best model → schedule full reindex
- This is a one-time task, not an ongoing workstream

#### Phase 4: Pillars/Notes/Reflections (2-3 sessions)
- [Plan 08](../../../scripts/plans/08_becoming-next.md): Add the meaning layer to the becoming app
- Depends on scheduled tasks being done first

---

## 4. Git Audit Corrections (2026-03-12)

Cross-referenced all plans against actual git history and code. Key corrections:

1. **Copilot SDK IS in brain.exe** — v0.1.29 in `go.mod`, dual backend system. Phase 1 of Workstream 1 is smaller than estimated.
2. **brain.exe has 10 internal packages + 5 cmd binaries** — includes `bench` and `eval` tools (model testing infrastructure) not in any plan.
3. **brain-app ROADMAP.md is stale** — rich text and sub-tasks shown as unchecked but are done (Plans 10-11). Far-term section (Play Store, BYOK, standalone) exists but isn't plans.
4. **SPEC-NEAR-TERM.md v2** has 4 items not in plans: done filter bug, history bottom inset, home screen widget redesign, widget mic recording.
5. **becoming-mcp has 22 tools** — brain tools already consolidated in. More mature than inventoried.
6. **byu-citations MCP** built (commit `870702c`, Mar 2) — not in any plan. Add to tool inventory.
7. **chip-voice has 6 internal proposals** — separate scope, acknowledged in deferred list.
8. **private-brain repo** — 2 commits, scaffolding only. Confirms Garvis → brain.exe merge recommendation.
9. **Uncaptured scripts:** chromem-exp (chunking experiments), convert (slides converter), lectures-on-faith (downloader) — utility scripts, no plans needed.

---

## 5. Deferred / Archived Work

| Item | Decision | Location |
|------|----------|----------|
| Plan 01: TUI for downloader | **Archived** | `scripts/plans/archive/` |
| Plan 04: Tool Improvements doc | **Archived** | `scripts/plans/archive/` |
| Plan 14: Q1 Roadmap | **Archived** | `scripts/plans/archive/` (superseded by this document) |
| Proposal: yt-emotion-analysis | **Archived** | `.spec/proposals/archive/` |
| Plan 09: Auth & Multi-user | **Update needed** | `scripts/plans/deferred/` — Auth is ALREADY DEPLOYED (Google OAuth + email/password). Plan is stale. Needs rewrite to reflect reality and remaining gaps (password recovery, email service). |
| Plan 12: Attachments | **Deferred** | `scripts/plans/deferred/` — Storage decided: brain=local, ibeco.me=S3 on NOCIX. Unblocked when ready. |
| Plan 13: Agentic Chat | **Deferred** | `scripts/plans/deferred/` — subsumed into Workstream 1 |
| Proposal: Garvis | **Merged** | `.spec/proposals/deferred/` — Garvis IS brain.exe. Name retired. Decision made Mar 19. |
| Proposal: tts-stt-reader | **Deferred** | `.spec/proposals/deferred/` — revisit after chip-voice batch/multi-voice ships |
| Plan 19: Ideas backlog | **Keep** | `scripts/plans/` — it's a backlog, not a plan |
| Widget Phases 3-4 | **Paused** | *(noted in Plan 18)* — Keep in roadmap, revisit after agent work is rolling |
| Becoming UX Phase 2 (Bookmarks) | **Deferred** | *(no standalone file — noted in docs/becoming-ux-phases.md)* |
| chip-voice 6 proposals | **Keep in scope** | Managed within chip-voice's own `.spec/proposals/` |
| byu-citations MCP | **Built, no plan needed** | Already working. Add to tool inventory. |
| brain-app SPEC-NEAR-TERM v2 | **Triage** | 4 items should be incorporated into WS2 Phase 1 or explicitly deferred |
| brain-app Far Term (Play Store, BYOK) | **Parked** | Long-term aspiration, not actionable yet |

| Title of Liberty (Boot Camp) | **New — Active** | `study/yt/title-of-liberty/` — Family discipleship program. Cross-cuts Becoming app (WS3) |

**Result:** 28+ active items → 3 workstreams with ~12 sequenced tasks + 1 new cross-cutting project. The rest is parked with clear revisit conditions.

---

## 5b. New Project: Title of Liberty

**Added:** 2026-03-14
**Location:** [study/yt/title-of-liberty/](../../../study/yt/title-of-liberty/README.md)

A multi-year family discipleship program grounded in the Book of Mormon. 5 ranks, 4 degrees per rank, 59 merit badges (7 required + 26 elective + 26 honors). Modeled after BSA rank progression with Book of Mormon character development.

**Cross-cuts:** Workstream 3 (Becoming App). The program integrates with ibeco.me for personal tracking (practices, memorization, badges) and webeco.me for the family/community layer (troops, leader dashboard, program templates). Auth (Plan 09, currently deferred) becomes relevant if this moves to multi-family.

**Priority decision needed:** Is this Workstream 3 Phase 5 (after Pillars/Notes/Reflections), or does it become its own workstream? The badge/rank tracking features (integration.md Priorities 4-6) are medium-effort and build on existing patterns. The community layer (Priorities 7-9) requires multi-user infrastructure.

**Recommended sequencing:**
1. Configure pillars + daily practices + scripture cards (Priority 1-3) — **now, no code needed**
2. Badge + rank tracking features — **after WS3 Phase 1 (Scheduled Tasks)**
3. Community layer — **after Auth decision (guidance.md Q3)**

---

## 6. Execution Plan

### Week 1: Foundation Sprint

| Day | Workstream 1 (Agentic) | Workstream 2 (Brain) | Workstream 3 (Becoming) |
|-----|------------------------|---------------------|------------------------|
| 1 | Copilot SDK POC setup | Plan 15: Quick wins | — |
| 2 | Copilot SDK + gospel-mcp | — | Plan 07: Scheduled tasks (backend) |
| 3 | — | Bidirectional sync | Plan 07: Scheduled tasks (frontend) |

### Week 2: Expansion

| Day | Workstream 1 | Workstream 2 | Workstream 3 |
|-----|-------------|-------------|-------------|
| 4 | Agent as spec executor | Proactive surfacing | MCP improvements P1-P3 |
| 5 | Test: agent executes MCP improvement | — | gospel-vec experiments |
| 6 | — | Garvis merge / server decision | Plan 08 start |

### Week 3+: Sustained

Agents handle more of the routine execution. Michael focuses on spec review, architectural decisions, and study. Workstreams continue but at sustainable pace.

---

## 7. Creation Cycle Review

| Step | This Proposal |
|------|---------------|
| **Intent** | Reduce overwhelm. Enable parallel execution. Ship more of what matters. |
| **Covenant** | Michael: review agent output within 24hrs, make judgment calls promptly. Agent: stay in scope, produce PR-quality work, flag uncertainty. |
| **Stewardship** | WS1 → dev agent. WS2 → dev agent (brain scope). WS3 → dev agent (becoming scope). Studies → study agent. |
| **Spiritual Creation** | This proposal. The scratch file. The guidance questions. |
| **Line Upon Line** | Phase 1 of each workstream is small (1 session each). Expand only after success. |
| **Physical Creation** | Execute workstreams. Agents build against specs. |
| **Review** | Each phase has verification criteria. "Watch until they obey." |
| **Atonement** | When agents produce bad output: capture in .spec/learnings/, adjust spec, retry. Don't just retry blindly. |
| **Sabbath** | No more than 3 sessions/week on infrastructure. Study at least 1x/week. Sabbath is real. |
| **Consecration** | Work-with-AI guide shared publicly. Studies published. Tools open-source. |
| **Zion** | Multiple agents, one purpose: "facilitate deep, honest scripture study." One intent.yaml to rule them all. |

---

## 8. Recommendation

**Build.** All guidance questions answered (Mar 19). Key decisions locked:

1. **Garvis = brain.exe.** Name retired. No new repo.
2. **Dual AI backend:** LM Studio for classification, Copilot SDK for agents.
3. **ibeco.me is multi-user.** Auth already deployed. Plan 09 needs rewrite.
4. **Option C confirmed.** Front-load agentic, fan out. Progressive trust.
5. **Study is highest priority.** Agentic and study run in parallel.
6. **Widget paused, not deferred.** Plan 18 stays in roadmap.
7. **Storage resolved.** Brain=local, ibeco.me=S3 on NOCIX.
8. **brain.exe deployment:** local → docker → NOCIX server.
9. **TUI and yt-emotion archived.** Already done.
10. **"Time to go down and build."**

Start with Workstream 1 Phase 1 (Copilot SDK + MCP integration) and study.

---

## 8. Costs & Risks

| Risk | Mitigation |
|------|-----------|
| Copilot SDK is technical preview — may have breaking changes | Pin version. Isolate SDK code. Don't build critical path on unstable API. |
| Agent-generated code may confabulate or be low quality | Always human-review. Start with low-stakes tasks (MCP improvements, not auth). |
| Planning about planning becomes its own avoidance | This overview exists to reduce planning, not add to it. After guidance answers → build. |
| Three workstreams may still be too much | If overwhelm returns, drop to 2. Workstream 3 is independent and can pause. |
| gospel-vec reindex takes 6 hours | Schedule during sleep or work time. Not a blocker for other work. |

---

*Research provenance: [.spec/scratch/overview/main.md](../../scratch/overview/main.md)*
*Judgment needed: [guidance.md](guidance.md)*
