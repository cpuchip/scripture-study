# Decisions Log

*Canonical record of settled questions. All agents read this as session context.*
*Active.md tracks current state. This file tracks what we decided and why.*

---

## Architecture

### Decision: brain.exe IS the second brain
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** brain.exe and the "Garvis" concept are the same thing. No new repo, no new binary. brain.exe evolves into the always-on second brain with SQLite + chromem-go + relay + MCP.
- **Rationale:** Same idea in Michael's head already. "Brain" is copyright-safe. The deferred Garvis proposal is merged into brain.exe evolution.
- **Supersedes:** `.spec/proposals/deferred/second-brain-architecture.md` (Garvis as separate repo)

### Decision: Dual AI backend — role separation
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** LM Studio (qwen3.5-9b on fermion's 4090) handles classification. Copilot SDK (Opus 4.6 / Sonnet 4.6) handles agent work — spec execution, reasoning, complex tasks. "Lepton" is a second 4090 machine available for inference load.
- **Rationale:** Progressive stewardship model. LM Studio is trusted, tested, free/hardware, no API costs — right fit for classification. Copilot SDK brings capability for agent-level reasoning.

### Decision: ibeco.me is multi-user
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Designed for Michael but general enough for kids, eventually families/groups via webeco.me. Google OAuth + email/password auth already implemented and deployed. brain.exe and brain-app remain single-user with token auth.
- **Rationale:** Auth is further along than Plan 09 describes. Plan 09 is stale vs actual code.
- **Supersedes:** Plan 09 (Becoming Auth) — stale

### Decision: Storage — local for brain, S3 for ibeco.me
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** brain.exe stores data locally (filesystem alongside SQLite and vector DB). ibeco.me storage goes to S3, self-hosted on the new NOCIX server (3TB, unmetered 1Gbps).
- **Rationale:** Brain is personal and local-first. ibeco.me needs server-accessible storage for multi-user.

### Decision: brain.exe deployment — local first
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Local → dockerize → deploy to NOCIX server alongside ibeco.me. Sequential, not rushed.
- **Rationale:** Prove it works locally before deploying. The new NOCIX server will host both.

---

## Priorities & Sequencing

### Decision: Study is the highest priority
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Study IS the top priority — "it keeps me in the spirit." Fully agentic and study are the two priorities, running in parallel. Study isn't deferred FOR infrastructure; they feed each other.
- **Rationale:** The project's intent is scripture study. Infrastructure serves the study, but serving the study means actually *doing* study.

### Decision: Option C — front-load agentic, then fan out
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Front-load agentic infrastructure (1-2 sessions), then fan out to 3-4 workstreams once agents can execute. Interested in VS Code hooks (v1.111) for chaining specs with the ~5 premium request budget.
- **Rationale:** Sequential focus yields better results than parallel sprawl. Still cautious about unsupervised agent work.

### Decision: Widget paused, not deferred
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Plan 18 stays in the main roadmap. Phases 3-4 (memorize widget, background refresh) paused until agent infrastructure is rolling.
- **Rationale:** Valuable but not urgent. Revisit once agentic work proves out.

### Decision: No more research gates before Phase 0
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Stop researching and start building. Phase 0 (practice what we preach — decisions.md, intent.yaml in session-start, Sabbath) before any new infrastructure.
- **Rationale:** "Time to go down and build." The overview was valuable but shouldn't become its own project.

---

## Agentic Architecture

### Decision: Gated autonomy, not unlimited
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Agents wait for human-assigned specs. Human assigns work. Level 2 autonomy requires more harness first. "Scared of letting you go without direct oversight."
- **Rationale:** Progressive trust model. The plan agent creates a well-drafted backlog for agents to work through — agents don't run autonomously without sufficient harness.

### Decision: Cost unit is premium requests
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Track cost in GitHub Copilot premium requests (1500/month), not raw tokens. Currently 56% utilization with 1/3 month remaining — best utilization month yet.
- **Rationale:** Premium requests are the actual constraint. Token counting is implementation detail.

### Decision: Ben Test canonized
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** The "Ben Test" is a formal skill at `.github/skills/ben-test/SKILL.md`. Practice calibrated self-assessment before claiming strengths. Evidence levels: practiced, occasional, aspirational, mythical.
- **Rationale:** Ben's feedback ("Your AI is very complimentary. Perhaps too complimentary?") was the right kind of honest. Canonized in his honor.

---

## Archived / Dropped

### Decision: Drop TUI (Plan 01)
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Gospel Library Downloader TUI archived. The API/CLI pipeline works better.
- **Rationale:** Already archived. Scripts work. TUI was nice-to-have.

### Decision: Archive yt-emotion-analysis
- **Date:** 2026-03-19
- **Decided by:** Michael
- **Decision:** Archive the yt-emotion-analysis proposal. Revisit when yt-mcp gets more use.
- **Rationale:** Cool idea, not on any critical path.

---

## Technical Implementation (Mar 8-12)

### Decision: Secondary Flutter entrypoints in main.dart
- **Date:** 2026-03-08
- **Decided by:** Michael + dev agent
- **Decision:** `@pragma('vm:entry-point')` functions for widget overlays must be in `main.dart` for reliable AOT compilation.
- **Rationale:** Dart AOT tree-shaking removes entrypoints not in main.dart. Learned through debugging QuickAddPractice widget.

### Decision: Daily slots use value field for logging
- **Date:** 2026-03-11
- **Decided by:** Michael + dev agent
- **Decision:** `POST /api/logs` with `value: "morning"` (same endpoint, value field holds slot name). Backend `dailySlotsDue()` computes remaining slots. Config is nested JSON: `{"schedule": {"type": "daily_slots", "slots": ["morning", "bedtime"]}}`.
- **Rationale:** Reuses existing log endpoint. No new API surface.

### Decision: Per-instance widget filtering
- **Date:** 2026-03-12
- **Decided by:** Michael + dev agent
- **Decision:** Each widget instance stores `practice_filter_{widgetId}` in shared prefs. Cycle header tap or flyover activity to change.
- **Rationale:** Users want different widgets showing different categories.
