# 2026-06-11 — Session lanes: the workspace learns it has five minds

**Why:** Michael runs ~5 parallel Claude Code sessions (1–3 active as he has brain space), and today's duplicate-persona-host saga showed the memory system's multi-session gaps: pull-without-push (nobody knows when a sibling writes the board), no ownership registry (the exe hunt was process-parent forensics), invisible session identity (he typed into the wrong terminal — the root cause of the afternoon), and an 87K-token active.md no session could fully read.

**Counseled with Michael; his sharpenings:** per-session FILES not one registry file (write contention → zero by construction — his idea, strictly better); signals between sessions; keep sessions topic-based ("pipelines of work"). Ratified: **pull + mail badge** delivery (he stays the scheduler — gated autonomy; push/auto-wake deferred as per-lane opt-in someday) and **board surgery now** (other sessions idle).

**Shipped (root `<this commit>`):**
- `.mind/sessions/` — lane files (frontmatter: lane/session_id/status/started/last_active; body: Working on / Claims / Handoffs) + `inbox/<lane>.md` signals + protocol README. Rules: write own lane only, read everyone's, clear inbox after acting, check lanes before touching processes you didn't start.
- `.claude/hooks/` — `lane_start.py` (SessionStart: claim lane from title, grounding text now lives here), `lane_prompt.py` (heartbeat + 📬 nudge + board-changed nudge + bootstrap guidance for already-open sessions), `lane_bg.py` (background launches auto-logged as claims), `lane_end.py` (mark ended), `lanes_common.py`. Wired in settings.json; 50-tool reground text extended.
- **Statusline:** `⟨lane⟩ [Model] ▓▓▓░ 34% ctx · 📬 N · 5h %` — the wrong-terminal fix and the visible mail.
- **Board surgery:** active.md 87K → ~1.2K tokens; full banner ledger archived verbatim at `.mind/archive/active-ledger-thru-2026-06-11.md`; board header carries the discipline (journal it, then delete the line; re-add missing threads as one-liners).
- CLAUDE.md addendum: Session lanes section (the reground path teaches existing sessions).

**Verified by simulation:** claim → mail nudge → board-changed nudge → bootstrap message → background claim (foreground ignored) → statusline render → ended → cleanup. All green before wiring went live.

**Bootstrap state:** this session's lane (`general-workspase.md`) created with a placeholder session_id (stamps on Michael's next message); the other 4 idle sessions will be prompted to claim lanes on their next engagement.

**Carry-forward:** per-lane opt-in push wake (background watcher; the harness re-invokes on exit) when a real pipeline needs it; the plugin-someday now has three chapters (covenant memory, context statusline, lanes).
