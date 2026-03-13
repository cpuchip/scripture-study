# Guidance — Questions Needing Michael's Judgment

*Created 2026-03-12 by plan-exp1. Review in the morning.*

---

## Priority 1 — Architectural Decisions

### Q1: Is brain.exe the same thing as Garvis?

Two proposals describe overlapping visions:
- [second-brain-architecture.md](../../proposals/second-brain-architecture.md) — "Garvis" as a new Go binary, new repo (`cpuchip/garvis-memory`), YAML+Git storage
- [brain-memory.md](../../proposals/brain-memory.md) — brain.exe evolving to SQLite + chromem-go

The scratch file [recommends merging them](../../scratch/overview/main.md): **brain.exe IS Garvis Phase 1.** Same binary, evolved. No new repo. SQLite + chromem-go + relay + MCP work IS the foundation for the always-on second brain.

**Your call:** Merge or keep separate? If merge, do we retire the "Garvis" name and just call it brain.exe?

---

### Q2: Copilot SDK — Which backend is active? *(Partially resolved)*

**Git audit confirmed:** Copilot SDK IS in brain.exe's `go.mod` at v0.1.29. It's implemented as a dual-backend system — `cfg.AIBackend` selects between `"copilot"` (GitHub Copilot SDK) and `"lmstudio"` (local inference). Both work. The integration lives in `internal/ai/client.go`.

**Remaining question:** Which backend are you actually using day-to-day? If LM Studio has fully replaced Copilot SDK in practice, we should understand why (latency? cost? offline preference?) before building multi-agent orchestration on top of the SDK. If both are in active use, that's a strength — local for speed, cloud for capability.

---

### Q3: Is Becoming intended to be multi-user?

[Plan 09 (Auth)](../../../scripts/plans/09_becoming-auth.md) designs full SaaS auth (email/password, Google OAuth, user_id on every table, PostgreSQL). But is that the actual goal? If the app is primarily for you, auth is unnecessary complexity. If it's for others too, it changes the priority of everything.

**Your call:**
- **Personal tool** → Skip auth. Keep SQLite. Simplify.
- **Shareable tool** → Auth stays on the roadmap but can remain deferred.
- **SaaS product** → Auth moves UP in priority (before adding features that assume single-user).

---

## Priority 2 — Scope & Sequencing

### Q4: How many parallel workstreams feel right?

The proposal identifies 5 potential workstreams. The 11-step cycle and Mosiah 4:27 both say "not faster than you have strength." With a full-time job and family:

- **Option A:** 2 workstreams max (Agentic Foundation + one other). Focused. Sustainable.
- **Option B:** 3 workstreams (Agentic + Brain + Becoming). Ambitious but doable if agents handle execution.
- **Option C:** Front-load agentic infrastructure (1-2 sessions), then fan out to 3-4 lanes once agents can execute.

Option C is the proposal's recommendation. But you know your schedule better than I do.

---

### Q5: What's the actual priority — tool infrastructure or study content?

The project's **intent** is scripture study. But almost all recent work has been infrastructure (brain-app, widget overhaul, relay, TTS). The infrastructure serves the study — but at some point, serving the study means actually *doing* study.

The MCP improvements (7 items) and gospel-vec experiments would directly improve study quality. The agentic infrastructure would eventually automate parts of the study workflow.

**Your call:** Should "Study & Content Quality" (Workstream 4) be higher priority than it currently is? Or is the infrastructure investment the right call for now?

---

### Q6: Widget overhaul — keep pushing or pause?

Plan 18 has Phase 2 done (practice widget). Phases 3-4 (memorize widget, background refresh) remain. These are useful but incremental. The active.md lists several SPEC-NEAR-TERM.md items not started.

**Your call:** Continue widget work, or redirect that energy to agentic foundation?

---

## Priority 3 — Technical Choices

### Q7: Storage decision for brain attachments

Plan 12 (attachments) is blocked on S3 vs. local storage. This doesn't need to be decided now unless attachments are high priority.

- **Local (filesystem alongside SQLite):** Simplest. Works offline. Backup = copy directory.
- **S3-compatible (Backblaze B2, Cloudflare R2):** Scalable. Accessible from server and phone. Costs pennies.
- **Defer:** Don't decide now. Attachments aren't on the critical path.

---

### Q8: Deploy brain.exe to server — when?

Multi-agent-ideas.md identifies "Get brain.exe on a server" as step 1 of 6. Garvis Phase 1 is blocked on this. ibeco.me already runs on Dokploy.

**Your call:** Is server deployment for brain.exe a near-term priority? Or does it wait until the agentic foundation is proven locally first?

---

## Priority 4 — Things I Think We Should Drop or Archive

These aren't urgent decisions, but naming them might help clear the cognitive load:

### Q9: Archive candidates?

| Item | Reasoning |
|------|-----------|
| Plan 01: Gospel Library Downloader TUI | The API pipeline works. The TUI is nice-to-have. Scripts work. |
| Plan 04: Tool Improvements (doc) | Superseded by Plan 05 (actionable tasks). Keep Plan 05, archive Plan 04. |
| Plan 14: Q1 Roadmap | It's Q1 2026. This document is being superseded by this overview. |
| Proposal: yt-emotion-analysis | Cool idea, but not on any critical path. Archive and revisit when yt-mcp gets more use. |

---

## Priority 5 — Emotional / Relational

### Q10: Is this overview itself the right thing?

I want to name this honestly: you asked for an overview and I've produced another large planning artifact. The risk is that planning about planning is its own form of avoidance. The 11-step cycle says "organize, prepare, establish" — but it also says "let us go down." At some point, the right action is to stop mapping and start building.

My recommendation: review this in the morning, make the judgment calls above, and then we start **building** in the next session. The overview exists so we don't lose context. It shouldn't become its own project.

---

*These questions are ordered by impact. Q1-Q3 affect architecture. Q4-Q6 affect sequencing. Q7-Q8 affect specific plans. Q9-Q10 affect mindset.*
