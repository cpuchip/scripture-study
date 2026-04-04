# Brain Project-Kanban: From Capture Tool to Goal Orchestrator

**Binding problem:** brain.exe captures and classifies thoughts but operates as a flat entry list with no project-level organization, no iterative agent sessions, and no way for Michael to manage work outside VS Code. The approval queue surfaces individual entries but not goals. Agents fire-and-forget instead of iterating. There's no way to say "show me everything for Sunday School" or "what's the agent working on for the space center?" The UI is a monitoring afterthought when it should be the primary work surface.

**Created:** 2026-04-04
**Research:** [.spec/scratch/brain-project-kanban/main.md](../../scratch/brain-project-kanban/main.md)
**Depends on:** WS1 Phase 3c (auto-routing + review queue) — SHIPPED, brain-ui-dashboard (in progress)
**Affects:** All brain usage, daily workflow, ibeco.me, brain-app
**Status:** Draft — council complete, awaiting review
**Inspiration:** Simon Scrapes "Agentic OS Command Center" (https://youtu.be/uhMCy25NBfw)

---

## 1. Problem Statement

brain.exe's current shape:

```
Capture → Classify (category) → Route (agent) → Agent runs → Review → Done
```

This pipeline is linear, entry-level, and fire-and-forget. Three problems:

### 1.1 No Project Organization

Entries exist in a flat list. "Fix gospel-engine search" and "Prepare Sunday School lesson on Alma 32" sit in the same undifferentiated queue. There's no way to:
- Filter by project (scripture study, teaching, space center, Sunday School)
- See progress toward a goal across multiple entries
- Give agents project-specific context

### 1.2 No Iterative Sessions

Agent routing is one-shot: approve → route → agent runs → done. But real work is iterative — the agent researches, comes back with questions, you refine, the agent continues. Simon Scrapes calls this "Your Turn / Claude's Turn." Our system has no concept of turns.

### 1.3 Brain Is Not the Hub

Michael lives in VS Code for development. All planning, progress tracking, and agent orchestration happens in VS Code chat. Brain captures thoughts but doesn't manage work. The goal is for brain to be the primary surface for project management and agent orchestration, with VS Code remaining the coding tool.

### Success Criteria

1. **Entries belong to projects.** Every routable entry can be assigned to a project. Unassigned entries still work.
2. **Dashboard shows projects, not just entries.** Top-level view is project cards with progress indicators.
3. **Agent sessions support turns.** An agent can pause, show output, and wait for feedback — not just fire-and-forget.
4. **Scheduled research runs on cadence.** Brain gathers AI news, articles, YouTube videos on a configurable schedule.
5. **Agent outputs are files, not just DB rows.** Research, plans, and deliverables land in project directories where they can be read, diffed, and versioned.
6. **Library view exposes agents, skills, and docs.** No need to open VS Code to browse what's available.
7. **Single-user, personal projects.** Not multi-client. Sunday School is a personal calling, not a client engagement.

---

## 2. Constraints & Boundaries

**In scope:**
- Schema additions: `projects` table, `entries.project_id` FK
- New views: Dashboard (project-level), Projects, Scheduled, Library
- Iterative session support via Copilot SDK session persistence
- Filesystem-based project directories for agent outputs
- Scheduled task engine (research passes on cadence)
- Skills and agents browser (reads `.github/skills/` and `.github/agents/`)

**Out of scope:**
- Multi-client / multi-user (this is personal)
- VS Code extension (noted as future possibility, not this proposal)
- tpg integration (may revisit in Phase 2 — tpg's task/epic/dependency model could complement projects)
- Mobile brain-app changes (Flutter app stays as-is for now)
- Full Paperclip-style org charts, role budgeting, or hierarchy

**Anchors:**
- **Mosiah 4:27** — things in order and wisdom, not faster than we have strength
- **Gated autonomy** — agents prepare, Michael decides
- **Files are durable** — agent outputs belong in the filesystem, not just the database

---

## 3. Prior Art & Related Work

| Source | Status | Relationship |
|--------|--------|-------------|
| brain-ui-dashboard.md | In-progress proposal | Absorbed — body preview (§10.1), model selector (§10.2), edit dialog (§10.3) fold into Phase 1 |
| brain-phase4-pipeline.md | Draft proposal | Complementary — maturity model IS the kanban engine underneath |
| brain-as-agent-os-platform scratch | Active research | Foundation — workspace-aware sessions, 11 agents, MCP tools |
| brain-multi-agent proposal | Phase 3c shipped | Prerequisite — routing, sessions, review queue all exist |
| tpg (Michael's issue tracker) | External tool | Adjacent — tasks/epics/dependencies/logs for agent context preservation |
| Simon Scrapes "Command Center" | External inspiration | Cherry-picked: turns model, scheduled tasks, skills UI, project tabs. Not: multi-client, paid academy, CLI-first |

---

## 4. Architecture

### 4.1 Projects Entity

```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'active',  -- active, paused, archived
    dir_path TEXT,                           -- workspace path for project files
    context_file TEXT,                       -- path to project context.md
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- FK on entries
ALTER TABLE entries ADD COLUMN project_id INTEGER REFERENCES projects(id);
```

**Initial projects** (based on Michael's stated scope):
- Scripture Study — the core workspace
- Teaching — YouTube series, podcast, shareable content
- Sunday School — calling management, presidency tools, lesson prep
- Space Center — (scope TBD)
- Brain Development — meta-project for brain.exe itself

### 4.2 Project Filesystem

Each project gets a directory in `private-brain/projects/`:

```
private-brain/projects/{project-slug}/
  context.md          # project intent, constraints, key decisions
  scratch/            # agent research outputs (durable provenance)
  outputs/            # agent deliverables
  notes.md            # ongoing notes, links, ideas
```

Agent sessions write to these directories. Outputs are files, not just database records. This means you can `read_file` an agent's research, diff versions, and carry context across sessions.

### 4.3 Iterative Sessions ("Turns")

Current flow:
```
Approve → Route → Agent runs (one shot) → Output → Done
```

Proposed flow:
```
Approve → Route → Agent runs → Agent pauses with output → YOUR TURN
  → You review, reply with feedback → AGENT'S TURN
  → Agent continues (same session, full context) → pauses → YOUR TURN
  → ... repeat until satisfied → Mark complete
```

Implementation: The Copilot SDK session model already supports this. `agent.AskStreaming()` returns a response, but the session stays alive. Calling `AskStreaming()` again with follow-up continues in the same context.

New entry statuses:
- `agent_turn` — agent is working or has pending output
- `your_turn` — agent paused, waiting for human input
- `completed` — goal achieved, output accepted

### 4.4 Scheduled Tasks

A lightweight scheduler in brain.exe that triggers research passes on defined intervals.

```go
type ScheduledTask struct {
    ID          int
    Name        string
    Description string
    Schedule    string    // cron expression or simple interval
    ProjectID   *int      // optional project scope
    AgentName   string    // which agent handles this
    Prompt      string    // what to do
    LastRun     time.Time
    NextRun     time.Time
    Status      string    // active, paused
}
```

**Initial scheduled tasks:**
- "AI Landscape Research" — weekly, search for latest AI articles/videos/tools → creates brain entries tagged to Brain Development project
- "Sunday School Prep" — weekly (configurable), pull next week's Come Follow Me block → create study entry tagged to Sunday School project
- "Weekly Review Digest" — Friday, summarize the week's brain activity, completed goals, open items

Not everything needs to be scheduled. Start with 2-3. See what's actually useful before adding more.

### 4.5 Navigation Structure

Inspired by Agentic OS but adapted for our workflow:

| View | Purpose | Maps to |
|------|---------|---------|
| **Dashboard** | Project-level kanban. Your Turn / Agent's Turn / Completed. Activity feed. | His "Feed" — but project-grouped |
| **Capture** | Quick thought capture (already exists) | Stays the same |
| **Projects** | Project list with detail views. Entries by maturity stage. Goals and progress. | His client tabs — but projects not clients |
| **Scheduled** | Recurring task management. Run history, enable/disable, create new. | Direct parallel to his Scheduled view |
| **Library** | Browse agents, skills, docs, memory files. Read-only reference. | His Skills + Docs combined |
| **Search** | Full-text and semantic search (already exists) | Stays the same |
| **Settings** | Models, governance rules, kill switch, system health | His Settings + our kill switch |

---

## 5. Phased Delivery

### Phase 1: Projects + Dashboard Rewrite

**Delivers:** Project CRUD, entries assigned to projects, dashboard grouped by project, entry body preview, model selector, edit dialog.

Schema changes:
- `projects` table
- `entries.project_id` FK
- New API endpoints: project CRUD, entries-by-project

Frontend changes:
- DashboardView rewrite: project cards → click into project detail
- Project cards show: name, status, entry count by maturity, recent activity
- Entry cards show: title, body preview (~200 chars), maturity badge, agent routing
- Model selector dropdown per entry
- Edit dialog (title, body, category, agent route, project assignment)

Backend changes:
- Project CRUD handlers
- Update entry handlers to support project_id
- Dashboard aggregation endpoint: projects with entry summaries

**Stands alone:** Yes. Even without turns or scheduled tasks, project grouping transforms the daily experience.

### Phase 2: Iterative Sessions

**Delivers:** Turn-based agent interaction, reply-in-context from dashboard, output preview, conversation history.

Schema changes:
- Entry status additions: `agent_turn`, `your_turn`
- `session_messages` table or markdown files for turn history

Frontend changes:
- Entry detail shows conversation history (turns)
- Reply input in entry detail view
- Status badges: "Your Turn" (red), "Agent's Turn" (blue), "Completed" (green)

Backend changes:
- Session persistence (keep SDK session alive between turns)
- Reply endpoint: send human feedback into existing session
- Output writing to project filesystem (`private-brain/projects/{slug}/scratch/`)

**Stands alone:** Yes. Even without scheduled tasks, iterative sessions transform agent interaction.

### Phase 3: Scheduled Tasks + Library

**Delivers:** Recurring research passes, skills/agents/docs browser, activity feed.

Schema changes:
- `scheduled_tasks` table
- `task_runs` table (history)

Frontend changes:
- ScheduledView: task list, schedule info, last run, run history
- LibraryView: tree browser for agents, skills, docs, memory files
- Dashboard activity feed (recent events across projects)

Backend changes:
- Scheduler goroutine (timer-based, runs tasks at defined intervals)
- Task execution: creates entries and/or runs agent sessions
- Library endpoints: list agents, skills, docs from filesystem

**Stands alone:** Yes. Scheduled tasks create value independently (overnight research, weekly digests).

---

## 6. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | Break free of VS Code as the work hub. Manage all projects from brain. |
| Covenant | Rules of engagement? | Gated autonomy. Agents prepare, Michael decides. Files are durable. |
| Stewardship | Who owns what? | brain.exe (Go backend + Vue frontend). Dev agent executes. Michael reviews. |
| Spiritual Creation | Is the spec precise enough? | Yes for Phase 1. Phases 2-3 need refinement after Phase 1 ships. |
| Line upon Line | What's the phasing? | 3 phases, each stands alone. Phase 1 is achievable in 2-3 sessions. |
| Physical Creation | Who executes? | Dev agent, with Michael reviewing each phase. |
| Review | How do we know it's right? | Dashboard shows projects. Entries group correctly. Body preview works. |
| Atonement | What if it goes wrong? | Rollback: project_id is nullable, existing entries unaffected. |
| Sabbath | When do we stop and reflect? | After each phase ships. Explicitly: use it for a week before Phase 2. |
| Consecration | Who benefits? | Michael directly. Eventually: the Work-with-AI guide becomes a testable reference. |
| Zion | How does this serve the whole? | Brain becomes the living proof-of-concept for steward-based AI orchestration. |

---

## 7. Costs and Risks

| Cost/Risk | Severity | Mitigation |
|-----------|----------|------------|
| Frontend complexity increases (7 views → 10+) | Medium | Phase delivery. Each view is independent. |
| Schema migration on live brain.exe | Low | SQLite `ALTER TABLE` is safe. project_id is nullable. |
| Scheduled tasks could burn premium requests | Medium | Cheap models only (Haiku). Configurable schedules. Budget cap per task. |
| Scope creep toward "full platform" | High | This proposal is the scope boundary. Not multi-client. Not tpg integration. Not VS Code extension. |
| Over-building before validating | Medium | Phase 1 must ship and be used for a week before Phase 2 starts. |

---

## 8. Recommendation

**Build.** Phase 1 first. The schema change is small (one table, one FK), the dashboard rewrite absorbs work already designed in brain-ui-dashboard §10, and the payoff is immediate — brain becomes navigable by project instead of a flat entry dump.

This proposal supersedes and absorbs:
- brain-ui-dashboard.md (body preview, model selector, edit dialog → Phase 1)
- brain-phase4-pipeline.md (maturity stages are the kanban columns → Phase 1 displays them)

It complements but does not replace:
- brain-as-agent-os-platform scratch (broader vision, this is the next concrete step)
- brain-multi-agent (routing infrastructure stays, this adds the project layer above)
