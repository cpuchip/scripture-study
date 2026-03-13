# Project Overview — Research & Inventory

*Working scratch file. Created 2026-03-12. Updated continuously during planning.*

---

## Binding Problem

Michael has 19 numbered plans, 9 formal proposals, multiple doc-level roadmaps, and scattered idea files — all at different levels of completion, with overlapping dependencies and no unified view of what's done, what's next, and what clashes. The project needs a single source of truth for all planned work, organized into executable workstreams that can run in parallel via Copilot SDK + multi-agent orchestration, following the 11-step creation cycle.

**Who's affected:** Michael — capacity is the bottleneck, not ideas.
**How would we know it's fixed:** One document maps all work → workstreams → dependencies. An agent can pick up any workstream and execute against it unattended.

---

## I. Complete Inventory

### A. Scripts/Plans (19 numbered plans)

| # | Title | Status | Category |
|---|-------|--------|----------|
| 01 | Gospel Library Downloader | Active (API done, TUI pending) | Infrastructure |
| 02 | Layout Decision | Done (use gospel-library only) | Infrastructure |
| 03 | Gospel MCP Server | **DONE** | Infrastructure |
| 04 | Tool Improvements | Backlog (roadmap) | Infrastructure |
| 05 | Tool TODO | Active (task breakdown) | Infrastructure |
| 06 | Becoming App Architecture | Phases 1-2.5 DONE, Phase 3 Sprint 1 DONE | App |
| 07 | Scheduled Tasks | Designed, not started | App |
| 08 | Becoming Phase 2.7 (Pillars/Notes/Reflections) | Designed, not started | App |
| 09 | Becoming Auth & Multi-User | Designed, not started | App |
| 10 | Brain Sub-tasks | **DONE** | Brain |
| 11 | Brain Rich Text | **DONE** | Brain |
| 12 | Brain Attachments | **BLOCKED** (S3/storage decision) | Brain |
| 13 | Brain Agentic Chat (Copilot SDK) | Designed, deprioritized | Brain/Agentic |
| 14 | Roadmap 2026 Q1 | Active planning doc | Meta |
| 15 | Brain-App Polish | Ready to code | Brain |
| 16 | Brain-App Today Screen | In development | Brain |
| 17 | Proactive Surfacing & Digest | Features 1-3 ready | Brain |
| 18 | Widget Overhaul | Phase 2 DONE, Phase 3-4 remaining | Brain |
| 19 | Brain-App Ideas | Idea backlog | Brain |

### B. Formal Proposals (.spec/proposals/)

| Proposal | Status | Category |
|----------|--------|----------|
| brain-ibecome-layer2 | Proposed (design complete) | Brain Sync |
| brain-memory | Proposed → Building | Brain Core |
| brain-relay | In Progress (Phases A-D done, E next) | Infrastructure |
| brain-unified-dashboard | In Progress (Phases 1-3 done, 4 next) | Brain UX |
| memory-architecture | **IMPLEMENTED** (files exist at .spec/memory/) | Meta/Memory |
| second-brain-architecture (Garvis) | Proposed (Phase 1 blocked on server) | Agentic |
| session-journal | **IMPLEMENTED** (scripts/session-journal/ exists) | Meta/Memory |
| tts-stt-reader | Draft (Phase 0 eval pending) | Multimedia |
| yt-emotion-analysis | Idea (Phase 0 POC pending) | Multimedia |

### C. Doc-Level Plans & Roadmaps

| Document | Status | Category |
|----------|--------|----------|
| [docs/becoming-ux-phases.md](../../../docs/becoming-ux-phases.md) | Phase 1 done, Phase 2 in design | App UX |
| [docs/mcp-improvements.md](../../../docs/mcp-improvements.md) | 7 action items, none started | Infrastructure |
| [docs/model-experiments.md](../../../docs/model-experiments.md) | Framework ready, baseline done | Infrastructure |
| [study/ai/multi-agent-ideas.md](../../../study/ai/multi-agent-ideas.md) | Ideas (6 next steps, not specs) | Agentic |
| [docs/work-with-ai/guide/](../../../docs/work-with-ai/guide/) | Parts 0-5 started, Part 6 planned | Education |

### D. Chip-Voice (separate repo)

| Item | Status |
|------|--------|
| Phase 0 eval | **DONE** (Qwen3-TTS 1.7B + Kokoro selected) |
| gen_audio.py pipeline | **WORKING** |
| Multi-voice, MP3, batch | Not started |
| Dual-mode engine | Not started |

---

## II. Status by Category

### DONE (Shipped & Working)
- Plan 02: Layout decision
- Plan 03: Gospel MCP
- Plan 06: Becoming Phases 1-2.5
- Plan 10: Brain sub-tasks
- Plan 11: Brain rich text
- Plan 18: Widget Phase 2 (practice widget)
- Proposal: memory-architecture (implemented)
- Proposal: session-journal (implemented)
- Chip-voice Phase 0 + gen_audio.py
- Brain relay Phases A-D
- Brain unified dashboard Phases 1-3

### IN PROGRESS (Active Work)
- Plan 16: Today screen (in development)
- Plan 18: Widget Phases 3-4 (memorize widget, background refresh)
- Brain relay Phase E (MCP next)
- Brain unified dashboard Phase 4 (bidirectional sync)

### DESIGNED & READY (Could start tomorrow)
- Plan 07: Scheduled tasks
- Plan 08: Pillars/Notes/Reflections
- Plan 15: Brain-app polish (4 quick wins)
- Plan 17: Proactive surfacing features 1-3
- Proposal: brain-ibecome-layer2

### DESIGNED BUT BLOCKED/DEFERRED
- Plan 09: Becoming auth (large scope, deferred)
- Plan 12: Brain attachments (blocked on S3 decision)
- Plan 13: Brain agentic chat (deprioritized)
- Proposal: Garvis/second-brain (blocked on server)

### IDEAS / DRAFT
- Plan 19: Brain-app ideas backlog
- Proposal: tts-stt-reader Phase 0 eval
- Proposal: yt-emotion-analysis POC
- MCP improvements (7 items)
- Model experiment runs
- Multi-agent orchestration infrastructure

---

## III. Dependency Graph

```
INFRASTRUCTURE LAYER
  gospel-library (01, 02) → gospel-mcp (03) ✅ → mcp-improvements (docs)
                                              → gospel-vec experiments (docs)
  webster-mcp ✅
  search-mcp ✅

BRAIN ECOSYSTEM
  brain-memory (proposal) → brain-relay (proposal, phases A-D ✅, E next)
                          → brain-unified-dashboard (proposal, phases 1-3 ✅, 4 next)
                          → brain-ibecome-layer2 (proposal)
                          → proactive surfacing (plan 17)
  brain-app: polish (15) → today screen (16) → widget P3-4 (18)

BECOMING APP
  architecture (06) ✅ → scheduled tasks (07) → today screen integration (16)
                       → pillars/notes/reflections (08)
                       → UX Phase 2 bookmarks (docs)
                       → auth & multi-user (09) [large, deferred]

AGENTIC / ORCHESTRATION
  Copilot SDK exploration → brain agentic chat (13) → Garvis (proposal)
  multi-agent-ideas.md → THIS OVERVIEW → execution workstreams

MULTIMEDIA
  chip-voice ✅ → tts-stt-reader (proposal) → audio publishing
  yt-mcp → yt-emotion-analysis (proposal)

EDUCATION / META
  work-with-ai guide (Parts 0-5)
  intent.yaml → all agents/skills
  session-journal ✅ → memory-architecture ✅
```

---

## IV. Clash & Ambiguity Analysis

### 1. Garvis vs. Brain.exe — Identity Crisis
**Problem:** Two overlapping visions for the same concept.
- `second-brain-architecture.md` proposes "Garvis" — Go binary, YAML+Git, always-on VPS, self-improving agent
- `brain-memory.md` proposes SQLite + chromem-go for brain.exe (same binary)
- `brain-relay.md` already connects brain.exe to ibeco.me
- `multi-agent-ideas.md` envisions brain.exe as the orchestration hub

**Clash:** Are we building *one* brain (brain.exe evolving into Garvis) or *two* systems (brain.exe local + Garvis server)?
**Recommendation:** Merge. brain.exe IS Garvis Phase 1. The SQLite + chromem-go + relay + MCP server work IS the foundation. Don't create a new repo — evolve the existing one.

### 2. Copilot SDK Status — CONFIRMED WORKING
**Correction (git audit):** Copilot SDK IS in brain.exe's `go.mod` at v0.1.29. It's an alternative AI backend (`"copilot"` vs `"lmstudio"`) configured via `cfg.AIBackend`. The initial build used Copilot SDK exclusively (commit `89309fe`), then LM Studio was added as an alternative (`6759347`). Both backends exist and work.
**Impact:** This is good news — Copilot SDK proof-of-concept is ALREADY DONE in brain.exe. Multi-agent orchestration can build on this existing integration rather than starting from scratch.
**Revised action:** The Workstream 1 POC effort is smaller than estimated. We can skip the standalone POC and build directly on brain.exe's existing `internal/ai/client.go`.

### 3. Plan 13 (Agentic Chat) vs. Multi-Agent Ideas — Scope Confusion
**Problem:** Plan 13 is narrow (Docker-isolated Copilot SDK for phone study). Multi-agent-ideas.md envisions a full pipeline (capture → proposal → execute → verify → ship).
**Recommendation:** Plan 13 is a subset. Multi-agent orchestration is the umbrella. Sequence: Copilot SDK proof-of-concept → Plan 13 as first use case → expand to full pipeline.

### 4. Three Surfaces, One API — Data Consistency
**Problem:** brain-app, brain web UI, and ibeco.me tasks page all access brain data differently. Unified dashboard (proposal) addresses this but Phase 4 (bidirectional sync) isn't started.
**Risk:** Building more features on top of inconsistent data is building on sand.
**Recommendation:** Phase 4 bidirectional sync before new brain features.

### 5. Becoming Auth (Plan 09) — Premature?
**Problem:** Auth & multi-user transforms the app from single-user local to SaaS. Every table gets user_id. It's the biggest migration and it's designed but deferred.
**Question:** Is multi-user the goal? Or is the app primarily for Michael? If primarily for Michael, auth can stay deferred indefinitely. If multi-user is the goal, it should come before features that assume single-user.

### 6. MCP Improvements — Aging Backlog
**Problem:** `docs/mcp-improvements.md` has 7 action items from Feb 3. None started. The improvements (markdown_link returns, full doc retrieval, preserved markdown) would significantly improve every study.
**Recommendation:** These are high-leverage, low-effort improvements. Should be in the first workstream.

### 7. Model Experiments — Blocking Full Reindex
**Problem:** gospel-vec experiments framework is ready but no experiments have been run. Full reindex takes 6 hours and is blocked until the right model is chosen.
**Impact:** gospel-vec semantic search stays at baseline quality until this is resolved.
**Recommendation:** Dedicate one focused session to run the experiments. Not a workstream — a one-time task.

### 8. Work-with-AI Guide — Unfinished
**Problem:** Parts 0-5 started but the confabulation audit (45 errors across 7 docs) happened in Feb. Were all corrections applied? What's the current quality state?
**Impact:** If we're going to use these docs as the framework for multi-agent orchestration, they need to be accurate.

### 9. Overwhelming Volume — Mosiah 4:27 Check
**Problem:** 19 plans + 9 proposals + 5 doc-level roadmaps + scattered ideas = too much in flight. Michael has named this pattern multiple times. Starting another "overview" project is itself an act of adding more.
**Mitigation:** This overview must REDUCE total in-flight work, not add to it. Output = fewer active threads, clearer priorities.

---

## V. Copilot SDK Research Findings

### What Is It?
GitHub Copilot SDK (v0.1.32, technical preview) — embeds the Copilot CLI agentic engine into any app. Go SDK available at `github.com/github/copilot-sdk/go`.

### Go Quick Start
```go
client := copilot.NewClient(&copilot.ClientOptions{LogLevel: "error"})
client.Start(ctx)
defer client.Stop()

session, _ := client.CreateSession(ctx, &copilot.SessionConfig{
    Model: "gpt-5",
})
defer session.Destroy()

// Synchronous
response, _ := session.SendAndWait(ctx, &copilot.Prompt{
    Prompt: "What is 2 + 2?",
})
fmt.Println(response.Data.Content)
```

### Key Capabilities
- **Tool calling:** Define tools the agent can invoke (MCP server integration built-in)
- **Model selection:** Choose model per session (gpt-5, claude, etc.)
- **Streaming:** Real-time event stream for custom UIs
- **Session management:** Create/destroy sessions, carry state
- **MCP integration:** Connect existing MCP servers (gospel-mcp, webster-mcp, etc.)

### Multi-Agent Patterns (from research)
1. **Mission Control** (GitHub Blog): Assign tasks across repos, watch real-time logs, steer mid-run. Mental model shift from sequential to parallel.
2. **YAML State Machine** (Amaresh's agno-mission-control): Define missions as YAML state machines with stages and transitions. Not just code — any workflow.
3. **GitHub Actions Integration**: Run agents as scheduled GitHub Actions (weekly profile updates, etc.)
4. **Session → Workspace isolation**: Each agent session gets its own scoped context.

### What This Means for Us
- brain.exe can create Copilot SDK sessions programmatically
- Each MCP server (gospel-mcp, webster-mcp, gospel-vec, becoming-mcp) becomes a tool the agent can call
- A "study agent" session gets gospel tools + readonly scripture access
- A "dev agent" session gets filesystem access + build tools
- Orchestrator pattern: brain.exe routes captured ideas to the appropriate agent session

### Prerequisites
- GitHub Copilot CLI installed and authenticated
- Active Copilot subscription (Michael has this)
- Go 1.21+ (brain.exe already uses Go)

---

## VI. The 11-Step Creation Cycle Application

From [docs/work-with-ai/guide/05_complete-cycle.md](../../../docs/work-with-ai/guide/05_complete-cycle.md):

| Step | Applied to This Overview |
|------|--------------------------|
| 1. **Intent** | Reduce overwhelm. Enable parallel execution. Ship more, plan less. |
| 2. **Covenant** | Michael: review within 24hrs. Agent: stay in scope, flag trade-offs. |
| 3. **Stewardship** | Each workstream has an owner (agent mode). Progressive trust. |
| 4. **Spiritual Creation** | This proposal IS the spiritual creation. Blueprint before building. |
| 5. **Line Upon Line** | Phase 1 is small (Copilot SDK POC + 3 quick wins). Expand from there. |
| 6. **Physical Creation** | Agents execute workstreams against specs. |
| 7. **Review** | Each phase has verification criteria. "Watch until they obey." |
| 8. **Atonement** | When things break, capture the learning. Forward-recover. |
| 9. **Sabbath** | Built-in rest between phases. Not everything at once. |
| 10. **Consecration** | Work-with-AI guide shared publicly. Studies published. |
| 11. **Zion** | Multiple agents, one purpose, one intent.yaml. |

---

## VII. Emerging Workstream Structure

Based on the dependency graph and clash analysis, work naturally groups into these lanes:

### Workstream 1: Agentic Foundation (PRIORITY — unblocks everything)
- [ ] Copilot SDK proof-of-concept in Go (standalone, not brain.exe yet)
- [ ] Wire one MCP server (gospel-mcp) as a tool
- [ ] Test: agent answers a scripture question using gospel-mcp
- [ ] If successful: integrate into brain.exe
- **Why first:** This is the infrastructure that lets agents execute the other workstreams.

### Workstream 2: Brain Consolidation (high leverage)
- [ ] Plan 15: 4 quick polish wins
- [ ] Brain unified dashboard Phase 4 (bidirectional sync)
- [ ] Plan 17: Proactive surfacing features 1-3
- [ ] Clarify: Garvis = evolved brain.exe (merge proposals)
- **Why second:** Brain is the central hub. It needs to be solid before agents build on it.

### Workstream 3: Becoming App (independent lane)
- [ ] Plan 07: Scheduled tasks
- [ ] Plan 08: Pillars/Notes/Reflections
- [ ] Becoming UX Phase 2: Bookmarks
- [ ] MCP improvements (7 items) — feeds study quality
- **Why parallel:** Becoming app work is independent of brain work. Different codebase, different agent.

### Workstream 4: Study & Content Quality
- [ ] gospel-vec model experiments (one focused session)
- [ ] MCP improvements for gospel-mcp
- [ ] Complete any remaining Work-with-AI guide corrections
- **Why now:** Quality infrastructure that makes everything else better.

### Workstream 5: Multimedia (defer unless energy)
- [ ] chip-voice: multi-voice, MP3, batch
- [ ] tts-stt-reader Phase 0
- [ ] yt-emotion-analysis POC
- **Why deferred:** Nice-to-have, not on the critical path.

---

## VIII. Open Questions for Michael

See `.spec/proposals/overview/guidance.md` for the full list of questions needing judgment.

---

## IX. Git Audit Findings (2026-03-12)

Cross-referenced all plans against actual git history and code state.

### Corrections to Inventory

1. **Copilot SDK is WORKING in brain.exe** — `go.mod` has `github.com/github/copilot-sdk/go v0.1.29`. It's an alternative AI backend alongside LM Studio. Guidance Q2 is resolved: not aspirational, it's shipped code.

2. **brain.exe architecture is larger than inventoried** — 10 internal packages (`ai`, `classifier`, `config`, `discord`, `ibecome`, `lmstudio`, `mcp`, `relay`, `store`, `web`) + 5 cmd binaries (`brain`, `brain-cli`, `brain-mcp`, `bench`, `eval`). The `bench` and `eval` tools for model testing are noteworthy — they're infrastructure for the model experiments workstream.

3. **brain-app ROADMAP.md is stale** — Lists rich text and sub-tasks as "Medium Term" unchecked, but both are DONE (Plans 10-11, committed). The ROADMAP needs updating or archiving.

4. **brain-app SPEC-NEAR-TERM.md v2** has items NOT in any plan:
   - Done filter bug (shows all done items incorrectly)
   - History screen bottom inset bleeding behind nav bar
   - Home screen widget redesign (Microsoft To Do-inspired with starred field, checkboxes, mic recording)
   - Widget mic recording overlay
   These should be incorporated into Workstream 2 or deferred explicitly.

5. **becoming-mcp has 22 tools** — 8 read becoming tools, 6 write becoming tools, 5 read brain tools, 3 write brain tools. Brain tools already consolidated from standalone brain-mcp into becoming-mcp. This is more mature than the inventory suggested.

6. **byu-citations MCP server** — Built March 2 (commit `870702c`), not in any plan. A standalone tool for BYU scripture citation index queries. Should be added to the tool inventory.

7. **chip-voice has 6 internal proposals** in `.spec/proposals/`:
   - 1080ti-testing.md
   - batch-generation.md
   - dual-mode-engine.md
   - mp3-output.md
   - multi-voice.md
   - phase0-results.md
   These are separate from the main scripture-study proposals and need to be acknowledged.

8. **private-brain repo** — Scaffolding only (2 commits). This is the Garvis-era structure (YAML categories, config, guardrails, principles). Confirms the Garvis → brain.exe merge recommendation — private-brain is dead code.

9. **brain-app ROADMAP.md Far Term** — Contains Play Store / public release / BYOK / standalone mode vision not captured anywhere else. This is a significant scope expansion if pursued.

10. **chromem-exp directory** — Experimental chunking code for vector embeddings. Related to gospel-vec model experiments.

### Missing from Any Plan
| Item | Source | Action |
|------|--------|--------|
| byu-citations MCP | git commit `870702c` | Add to tool inventory |
| SPEC-NEAR-TERM v2 items | brain-app/SPEC-NEAR-TERM.md | Triage into Workstream 2 or defer |
| brain-app Far Term vision | brain-app/ROADMAP.md | Acknowledge as long-term aspiration, don't plan yet |
| chip-voice 6 proposals | chip-voice/.spec/proposals/ | Keep in chip-voice scope, add to deferred list |
| brain bench/eval tools | brain/cmd/bench, brain/cmd/eval | Note as existing infrastructure |

---

*Proposal at `.spec/proposals/overview/main.md`. Guidance questions at `.spec/proposals/overview/guidance.md`.*
