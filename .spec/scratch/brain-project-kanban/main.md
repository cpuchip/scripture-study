# Research: Brain as Project-Based Kanban Platform

**Created:** 2026-04-04
**Status:** Active research — council session in progress
**Binding problem:** brain.exe captures and classifies thoughts but has no project-level organization, no iterative agent sessions ("turns"), and no way to break free of VS Code as the primary work surface. Michael wants brain to be the hub for ALL projects (scripture study, teaching, space center, Sunday School presidency) with goal-based orientation, not just a thought capture tool.

---

## Source: Simon Scrapes "Command Center" Video

**Video:** "Stop Using Claude Code in Terminal (It's Holding You Back)" — Simon Scrapes, 2026-04-02, 16:32
**URL:** https://youtu.be/uhMCy25NBfw

### Problem He Identifies
- Managing multiple Claude Code terminals is chaos — no history, no handoff, no context between sessions
- Existing tools (T-Mox, Anthropic desktop, Vibe Kanban, Paperclip) solve parts but not the whole

### His Solution: "Agentic OS Command Center"
A web UI dashboard on top of Claude Code CLI with:

1. **"Your Turn" / "Claude's Turn" kanban** — iterative ping-pong, not linear pipeline. Cards move between "needs your input" and "agent working" columns.
2. **Goal-level task cards** (not session-level) — a single goal persists across multiple agent interactions. Conversation history attached to each card.
3. **Multi-client project tabs** — filter by client/project (his: "Client One", "Test Client")
4. **Scheduled tasks** — cron-like recurring agent work:
   - "Claude Code Trending" — daily research pass
   - "Monthly Learnings Health Check" — audit learnings.md for bloat/contradictions  
   - "Skill Update Check" — weekdays, check for new/updated skills
   - "Weekly Activity Digest" — Friday summary
   - "Nano Banana Trending Research" — daily, every 10min (domain-specific research)
5. **Skills management UI** — browse, search, view SKILL.md files organized by category (META, MKT, OPS, STR)
   - Each skill has: name, description, references directory, frameworks/examples
   - Skills have dependencies on each other
   - Visible output path: "Persuasive copy saved to `projects/mkt-copywriting/{campaign-name}/`"
6. **Docs browser** — view CLAUDE.md, SOUL.md, USER.md, learnings.md, memory files, project files
   - SOUL.md defines agent personality/values ("Core Truths": genuinely helpful, have opinions, be resourceful, anticipate needs, own mistakes, work across domains)
   - Projects are directories in the filesystem with their own context
7. **Feed view** — unified activity stream with history sidebar. "Start Here" and "Wrap Up" action buttons at top.
8. **ACHIEVED section** — completed goals at the bottom (25 shown), collapsed history by day ("Yesterday", "10 older")

### His Navigation Structure
- **Feed** — main dashboard / kanban board (Your Turn / Claude's Turn / Achieved)
- **Scheduled** — recurring task management
- **Skills** — skill catalog browsing and management
- **Docs** — documentation/context files browser
- **Settings** — configuration

### Key Design Observations from Screenshots
- Left sidebar: HISTORY with chronological task list
- Main area: Cards in "YOUR TURN" and "SCHEDULED" sections
- Cards show: title, status ("Needs Input"), turn count (e.g., "26/28" for long conversations)
- Expanded card: shows conversation history with "2 earlier messages" collapsed, reply input at bottom
- Clean, warm design (not dark mode — light with muted coral/salmon accents)
- Skills organized by prefix convention: `meta-*`, `mkt-*`, `ops-*`, `str-*`

---

## Source: Michael's Vision (Apr 4 session)

### Pain Points
1. "I have not had any experience using our new agentic OS in brain but I still don't think it's quite right"
2. Well-planned tasks in .spec but jumbled mess in brain — some entries already done but never checked off
3. Lives in VS Code when developing but wants to "break free and use brain and the various UIs we've built"
4. Brain needs to work across ALL projects (scripture study, teaching, space center, Sunday School)
5. Wants goals/outcomes orientation vs specific tasks

### Key Phrases
- "agents act like people where 'turns' are back-and-forths on 'tickets' on a board"
- "context into what happened, what changed, when to put data back in, kick off to agent, agent kicks back"
- "organized approach" — the Agentic OS navigation structure resonates
- "I don't want to go full paperclip" — don't over-engineer roles/hierarchy
- "a harmony between the two" — project-based organization that fits our workflow
- "gated autonomy" — this vision requires more trust infrastructure

### Anchors
- **Single-user, personal projects** — not multi-client. Sunday School is a calling (personal project), not a client engagement.
- **Mosiah 4:27** — things in order and wisdom, not faster than we have strength
- **VS Code extensions** could complement (not replace) the brain UI

---

## Council Synthesis: What We Have vs What We Need

### Comparison Table

| His System | Our System | Gap |
|---|---|---|
| Goal-level kanban | Entry-level approval queue | **Big:** No "goal" entity above individual entries |
| Your Turn / Claude's Turn | Approve → Route → done | **Medium:** No iterative back-and-forth in UI |
| Multi-project tabs | Single flat list | **Big:** No project scoping on entries |
| Scheduled tasks | Pipeline auto-advance (raw→researched) | **Small:** Engine exists, no schedule UI |
| Skills management UI | `.github/skills/` files | **Medium:** Files exist, no dashboard |
| Docs browser | `.spec/memory/`, intent.yaml, covenant | **Small:** Better context architecture, file-based only |
| Output preview | Nothing in UI | **Medium:** Dashboard §10.1 covers basic preview |
| Feed with history | Entries list | **Medium:** No activity stream or history sidebar |
| ACHIEVED section | No completion tracking in UI | **Medium:** Pipeline has "verified" stage but no celebration |

### What's Actually Good About What We Already Have
1. **Maturity pipeline IS a kanban** waiting for a UI (raw→researched→planned→specced→executing→verified)
2. **Workspace-aware sessions** already load agents/skills/tools per session
3. **Memory architecture** is more sophisticated than SOUL.md + USER.md
4. **Governance model** (gated autonomy) is more principled than "send to Claude"
5. **MCP tools** give agents real capabilities beyond just "run Claude CLI commands"

### Recommended Architecture: 3-Phase Evolution

**Phase 1 — Projects + Basic Dashboard (schema + UI)**
- `projects` table: id, name, description, status (active/paused/archived), created_at, updated_at
- `entries.project_id` FK (nullable — existing entries stay unassigned)
- Dashboard grouped by project instead of flat queue
- Project detail view: entries by maturity stage, goals, outputs
- Entry body preview in cards (§10.1 already designed)
- Model selector per entry (§10.2 already designed)  
- Edit dialog for entries (§10.3 already designed)

**Phase 2 — Iterative Sessions ("Turns")**
- Agent sessions persist across turns (not fire-and-forget)
- "Your Turn / Agent's Turn" status on entries
- Reply-in-context from dashboard (send feedback without opening VS Code)
- Agent output preview inline
- Conversation history per goal/entry

**Phase 3 — Scheduled Tasks + Management UIs**
- Scheduled research passes (latest AI topics, articles, YouTube videos on cadence)
- Skills/agents browser in brain UI (read from `.github/skills/` and `.github/agents/`)
- Docs browser for `.spec/memory/`, proposals, active work
- Activity feed with history sidebar
- VS Code extension that surfaces brain notifications / quick capture

### Critical Analysis Notes
1. **Is this the right thing to build?** Yes — Michael has said repeatedly he wants to break free of VS Code. The brain UI IS the right surface.
2. **Mosiah 4:27 check:** Phase 1 is achievable in 2-3 sessions. The phasing prevents building too much at once.
3. **What gets worse?** Frontend complexity increases. More views to maintain. But this is the intentional investment.
4. **Does this duplicate?** No — the brain-ui-dashboard proposal and brain-phase4-pipeline proposal feed INTO this. This is the unifying vision.
5. **tpg integration?** Michael's own issue tracker (tpg) solves context preservation for agents. Could integrate as a backend component for project/task management. Worth considering in Phase 2.

### Filesystem Approach for Projects
Simon's approach: each project is a directory with its own context files. Skills output to project-specific paths (`projects/mkt-copywriting/{campaign-name}/`).

**Our equivalent:** Each brain project could have a workspace directory:
```
private-brain/projects/{project-name}/
  context.md          # project intent, constraints, key decisions
  scratch/            # agent research outputs (durable, not DB-only)
  outputs/            # agent deliverables  
  history.md          # conversation log / turn history
```
This means agent outputs are files you can read, diff, and version — not just database rows. The scratch file approach we already use for studies would extend to all agent work.

### Navigation Structure (Inspired by Agentic OS, adapted for us)

His: Feed | Scheduled | Skills | Docs | Settings

Ours (proposed):
- **Dashboard** — project-level kanban, Your Turn / Agent's Turn, activity feed
- **Capture** — quick thought capture (already exists)
- **Projects** — project list with detail views, entries by maturity
- **Scheduled** — recurring research/review tasks
- **Library** — agents, skills, docs, memory browser (our equivalent of Skills + Docs combined)
- **Settings** — models, governance rules, kill switch

---

## Michael's Framing: Developer-Focused vs Steward Goal-Based

Michael identified a paradigm distinction worth capturing in the Work-with-AI guide:

**Developer-focused** (where most AI tools live):
- Terminal/IDE as primary surface
- Session-oriented (start conversation, get output, end)
- Code-centric (PRs, commits, tests)
- Tools: Copilot, Cursor, Claude Code CLI

**Steward goal-based** (where we're heading):
- Dashboard/brain as primary surface
- Goal-oriented (goals persist, agent turns iterate on them)
- Outcome-centric (did the goal advance? not "did the session produce code?")
- Project-scoped context (each project carries its own intent, constraints, history)
- Iterative not sequential — "Your Turn / Agent's Turn" ping-pong
- Tools: brain.exe, ibeco.me, brain-app, future VS Code extensions

This isn't a replacement — it's a maturation. You start developer-focused (learning the tools, building skills, establishing patterns). You mature into steward goal-based (orchestrating outcomes, managing agent relationships, progressive trust).

The guide could frame this as Part 7: "From Developer to Steward" — the shift from "I use AI tools" to "I orchestrate AI toward outcomes."

---

## Open Questions

1. **Project scope:** How granular? Is "Sunday School" one project, or is each quarter a sub-project?
2. **History format:** Should turn history be markdown files (durable, readable) or DB records (queryable, fast)?
3. **tpg integration:** Bring tpg's task/epic/dependency model into brain, or keep them separate?
4. **VS Code extension:** What would it do? Quick capture? Brain notifications? Dashboard embed?
5. **Migration:** How do the 45+ existing brain entries get assigned to projects?
6. **Scheduled task engine:** New feature in brain.exe? Or leverage existing OS scheduling (Windows Task Scheduler) with brain CLI commands?
