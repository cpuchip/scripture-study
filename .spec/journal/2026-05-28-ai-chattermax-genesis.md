---
date: 2026-05-28
title: ai-chattermax — project scaffold + first design session
workstream: WS5-adjacent (not yet ratified as cycle thread)
session_type: design / scaffolding
status: design-only — no code, no ratification yet
related:
  - .spec/sabbath/2026-05-23-the-arc-that-said-yes-to-everything.md  # set-down seed
  - projects/ai-chattermax/.spec/journal/2026-05-28-genesis-and-design-session.md  # in-project journal
---

# ai-chattermax — project scaffold + first design session

## What happened (workspace-side summary)

Michael revived the chat-with-repos seed that the 2026-05-23 Sabbath explicitly set down (work-scope-adjacent, NOT one of the three ratified cycle threads). Created `projects/ai-chattermax/` with LICENSE (MIT), .gitignore (Go-flavored), README.md ("a hostable chat room for humans and AI."), pushed initial commit `cbfceb2` to `github.com/cpuchip/ai-chattermax`. Asked the agent to scaffold standard project files and journal the work, and granted commit+push stewardship parallel to marsfield.org.

Full in-project record at `projects/ai-chattermax/.spec/journal/2026-05-28-genesis-and-design-session.md`. This workspace-side journal is the bubble-up for cross-thread visibility, per the 2026-05-23-ratified [[Read Subproject Journals, Don't Bubble Them]] principle — workspace memory READS subproject journals at session start, but this entry exists because the genesis crosses workspace/subproject boundaries (Mosiah 4:27 evidence-test, three-cycle-threads context) and belongs in workspace-side context too.

## Covenant moment

Before scaffolding I named the Sabbath context: this seed was explicitly set down 5 days ago, Mosiah 4:27 is loaded as evidence-test, this would be a fourth thread on top of three ratified ones (substrate Council ②, teaching Episode 2, 1828 finish). I gave my honest answer to "what do you think" — idea is good, most architecture is right, protocol confusion (A2A vs MCP vs chat) is worth resolving early, design-only is the right move.

Michael heard the flag and chose to start the **design** work anyway, with build-vs-design held. Project scaffold + design proposal this session; ratification or set-down at next Sabbath.

## Workspace-side artifacts shipped

- `.mind/active.md` — banner entry added for ai-chattermax genesis (top of file)
- `.spec/journal/2026-05-28-ai-chattermax-genesis.md` — this file
- Auto-memory: `project_ai_chattermax.md`, `feedback_ai_chattermax_stewardship.md`
- `MEMORY.md` — index updated with both entries at top

## In-project artifacts shipped (full record in subproject journal)

- `projects/ai-chattermax/CLAUDE.md` — per-project context, stewardship protocol, open questions
- `projects/ai-chattermax/.mind/active.md` — initial active state
- `projects/ai-chattermax/.spec/journal/2026-05-28-genesis-and-design-session.md` — in-project record
- `projects/ai-chattermax/.spec/proposals/chat-server-design.md` — stub with five open questions

## Mosiah 4:27 evidence log entry

The 2026-05-23 Sabbath named: "if teaching gets crowded out by substrate again, that's the evidence." Today's session is the FIRST data point against that evidence-test. The reading at session close:

- ✅ Teaching: full *Beyond the Prompt* manuscript (Chapters 0–14) completed yesterday (2026-05-27). On pace.
- ✅ Substrate: science-news-weekly + daily-digest still firing autonomously. Council ② not yet started but soak running. On pace.
- ✅ 1828: webster-v2 thread untouched today (no work needed yet). Other 1828 carry-forwards (Phase 6 Thummim, UX 1-2 punch) untouched. Static, neither slipping nor advancing.
- 🆕 ai-chattermax: design-only scaffold added.

**No evidence of slip yet.** The "say yes to everything" risk is the rapid-revival itself (5 days from set-down to scaffold) and the design effort that just landed. Watch at next Sabbath whether one of the three threads has slipped while ai-chattermax design absorbed attention.

## Carry-forward

- **Next ai-chattermax pass:** walk the five open questions in `chat-server-design.md` via `AskUserQuestion` batches (substrate C-F cadence).
- **Coordination check:** before designing the MCP exposure pattern for ai-chattermax, check whether webster-v2 (Sabbath thread 3) is going to define an MCP server pattern that chat-persona-exposure can reuse. Don't design either in isolation.
- **Next Sabbath decision:** does ai-chattermax join the cycle as a fourth thread, displace one of the existing three (most likely candidate: 1828 finish, if webster-v2 + chat-persona-exposure can be designed jointly), or stay design-only another week.
- **Watch the pattern.** The seed→scaffold gap (5 days) is faster than the Sabbath assumed. Worth naming explicitly at next Sabbath: was this discovery-mode pulling, or "say yes to everything" reasserting?

## Files shipped this session (workspace-side)

- `.mind/active.md` — banner added
- `.spec/journal/2026-05-28-ai-chattermax-genesis.md` — this file
- `~/.claude/projects/.../memory/project_ai_chattermax.md` — new
- `~/.claude/projects/.../memory/feedback_ai_chattermax_stewardship.md` — new
- `~/.claude/projects/.../memory/MEMORY.md` — two-line addition at top
