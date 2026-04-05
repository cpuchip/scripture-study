# Brain Project-Kanban: From Capture Tool to Goal Orchestrator

**Binding problem:** brain.exe captures and classifies thoughts but operates as a flat entry list with no project-level organization, no iterative agent sessions, and no way for Michael to manage work outside VS Code. The approval queue surfaces individual entries but not goals. Agents fire-and-forget instead of iterating. There's no way to say "show me everything for Sunday School" or "what's the agent working on for the space center?" The UI is a monitoring afterthought when it should be the primary work surface.

**Created:** 2026-04-04
**Research:** [.spec/scratch/brain-project-kanban/main.md](../../scratch/brain-project-kanban/main.md)
**Depends on:** WS1 Phase 3c (auto-routing + review queue) — SHIPPED, brain-ui-dashboard (in progress)
**Affects:** All brain usage, daily workflow, ibeco.me, brain-app
**Status:** Building — Phases 1-3 shipped (Apr 4-5). Phase 4a-4b shipped (Apr 5). Phase 4c next.
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
- tpg-style task/epic/dependency model (may reimplement concepts if the model works out — tpg is Richard's project, not ours)
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

---

## Phase 4: Project Flow + AI Turn Automation

**Binding problem:** Phases 1-3 gave us projects, turns, scheduled tasks, and a library — the building blocks. But nothing ties them together into a managed workflow. There's no board view to see entries flowing through maturity stages, no AI push-back when entries are under-specified, no auto-assignment of incoming entries to projects, and no execution gate. Michael still has to manually drive every entry through every stage. The infrastructure is there. The workflow isn't.

**Research:** [.spec/scratch/brain-project-kanban-phase4/main.md](../../scratch/brain-project-kanban-phase4/main.md)

### What Exists (from Phases 1-3 + Pipeline 4a-c)

| Capability | Status | Gap |
|------------|--------|-----|
| Projects table + FK on entries | Shipped | — |
| Maturity stages (raw→verified) | Shipped (pipeline 4a-c) | No UI to drive transitions |
| Session messages (iterative turns) | Shipped | No AI-initiated push-back |
| Scheduled tasks | Shipped | No "project review" task type |
| Pipeline advance/revise/reject/defer | Shipped (MCP only) | No REST endpoint or UI button |
| Research pass (raw→researched) | Shipped | Only via MCP `brain_advance` |
| Plan pass (researched→planned) | Shipped | Only via MCP `brain_advance` |
| Classifier | Shipped | Doesn't suggest projects |
| Agent routing | Shipped | No project-aware context |
| ProjectDetailView groups by maturity | Shipped | Vertical list, not board; no actions |

### Sub-Phases

Phase 4 is broken into five independent sub-phases. Each delivers value alone. Order matters — 4a is foundation, the rest build on it.

---

### Phase 4a: Project Board View + Pipeline UI

**Delivers:** Kanban-style project board, pipeline actions in UI, per-project status rollup on dashboard.

**Backend changes:**
- `POST /api/pipeline/{id}/advance` — REST wrapper around existing `pipeline.Advance()`. Accepts: `action` (advance/revise/reject/defer), `feedback`, `scenarios`. Returns updated entry.
- `GET /api/projects/{id}/stats` — Returns per-maturity-stage entry counts, your_turn count, in-flight count.
- `POST /api/entries/{id}/advance-quick` — Shorthand: advance to next stage with no feedback (for drag-and-drop).

**Frontend changes:**

**ProjectDetailView → Board mode:**
- Horizontal columns for maturity stages (raw | researched | planned | specced | executing | verified)
- Entry cards show: title, category badge, route_status indicator (your_turn amber, running blue, complete green)
- Click entry card → slide-out panel with entry details, conversation history, scratch file preview
- Action buttons per entry: Advance ▶, Revise ↻, Defer ⏸ (calls pipeline advance endpoint)
- "Your Turn" entries get amber left border + bell icon
- Column headers show count
- Toggle between board view and current list view

**DashboardView enhancements:**
- Project cards show: stage distribution bar (colored segments for raw/researched/planned/specced/executing/verified)
- "Blocked on you" count badge (your_turn entries across all projects)
- "In flight" count (running entries)

**Scenarios (testable):**
1. ProjectDetailView shows horizontal kanban columns grouped by maturity
2. Clicking "Advance" on a raw entry triggers research pass; entry moves to researched column
3. Dashboard project cards show colored stage distribution bars
4. "Blocked on you" badge shows accurate count of your_turn entries
5. Board/list toggle persists across navigation

**Stands alone:** Yes. This is the highest-value sub-phase — it turns the existing data into a visible, actionable workflow.

---

### Phase 4b: Project Auto-Assignment in Classifier

**Delivers:** New entries get a suggested project_id during classification.

**Backend changes:**

In `classifier.go`:
- After classification, load all active projects (name + description) from DB
- Add project context to classifier prompt: "Given these projects: [{name}: {description}, ...], suggest which project_id this entry best fits. Return null if none fit well."
- Add `ProjectID *int` field to `classifier.Result`
- Store suggested project_id on the entry during `InsertEntry`
- Low-confidence suggestions (classifier isn't sure) → leave null, surface in UI as "unassigned"

**Frontend changes:**
- EntriesView: "Unassigned" filter to show entries without a project
- DashboardView: "Unassigned entries" count that links to filtered view
- EntryDetailView: one-click project assignment from suggestions (if classifier suggested but confidence was low)

**Scenarios:**
1. A new entry about "bridge simulator design" gets auto-assigned to Space Center (project 4)
2. A new entry about "grocery list" gets no project assignment (null)
3. Unassigned entries are visible on the dashboard with a count badge
4. Manual override still works via EntryDetailView dropdown

**Stands alone:** Yes. Even without the board view, auto-assignment reduces manual sorting (like the 67 entries we just sorted by hand).

---

### Phase 4c: AI Push-Back Loop

**Delivers:** AI proactively reviews stale entries, asks clarifying questions, drives entries toward specced.

**Backend changes:**

New scheduled task behavior: `project_review`
- System-provided (not user-created) scheduled task that runs every 4 hours
- Scans all active projects for entries where:
  - `maturity = 'raw'` and `updated_at` is 24h+ ago
  - `maturity = 'researched'` and `updated_at` is 48h+ ago
  - `route_status = 'complete'` and `updated_at` is 24h+ ago (agent finished but human hasn't reviewed)
- For each stale entry:
  - AI reads entry body + scratch file (if exists) + project context
  - Generates 2-3 specific clarifying questions or next-step suggestions
  - Posts as session message (role: "agent")
  - Sets route_status to "your_turn"
- Uses cheap model (Haiku) to keep costs low
- Configurable: enable/disable per project, adjustable staleness thresholds

**Reply → Auto-advance handler:**
- When Michael replies to a your_turn entry that was pushed back by the review agent:
  - If entry is raw + reply has enough substance → auto-trigger research pass
  - If entry is researched + reply refines direction → auto-trigger plan pass
  - If entry is planned + reply includes scenarios → advance to specced
  - Otherwise: just store the reply, keep as your_turn for continued conversation

**Frontend changes:**
- Entries pushed back by AI show "🤖 Review" badge to distinguish from agent-route turns
- Notification dot on project card when there are your_turn entries from push-back

**Scenarios:**
1. A raw entry untouched for 24h gets AI-generated questions posted as session messages
2. Entry moves to your_turn after push-back
3. Michael replies with clarification → research pass auto-triggers
4. Push-back can be disabled per project via project settings
5. Push-back uses cheap model (≤0.33 premium requests per entry)

**Stands alone:** Yes. This is the "keep tasks moving" behavior Michael asked for.

---

### Phase 4d: Agent Context Injection (Project-Aware Agents)

**Delivers:** When agents work on entries, they get project context: what the project is about, what other entries are in it, and what stage they're at.

**Backend changes:**

In `routeEntry()` / agent prompt construction:
- Load entry's project (if assigned): name, description
- Load sibling entries (same project, limit 20): title, maturity, route_status
- Load project scratch/context files if they exist
- Inject into prompt template:

```
Project: {project.Name}
Description: {project.Description}

Related entries in this project:
- [specced] Build bridge simulator movement system
- [researched] Star Trek UI with Pretext
- [raw] Build Physical Display Dashboard

Entry context:
- Maturity: {entry.Maturity}
- Previous research: {scratch file summary or "none"}
- Scenarios: {entry.Scenarios or "not yet defined"}
```

- New project field: `context_file TEXT` — optional path to a project-level context document that agents always receive

**Frontend changes:**
- Project edit form gets "Context file" field (path to a markdown file in the workspace)
- Entry detail shows "Agent context" expandable section showing what the agent will receive

**Scenarios:**
1. Agent working on a Space Center entry sees other Space Center entries in its context
2. Agent working on a 2nd Brain entry gets brain project description in its prompt
3. Project context file (if set) is included in every agent prompt for that project
4. Agent output quality improves with project context (qualitative, human-assessed)

---

### Phase 4e: Execution Gate

**Delivers:** Specced entries with scenarios can be kicked to execution from the UI, with cost visibility and result verification against scenarios.

**Backend changes:**
- `POST /api/entries/{id}/execute` — Validates entry is specced + has scenarios. Creates pipeline advance to "executing". Routes to appropriate agent with full context (project + spec + scenarios + scratch file). Returns immediately (runs async).
- `POST /api/entries/{id}/verify` — After execution completes, present scenarios as a checklist. Michael marks each pass/fail. All pass → verified. Any fail → revise with feedback.
- `GET /api/entries/{id}/execution-context` — Returns the full prompt that would be sent to the agent, so Michael can preview before approving.

**Frontend changes:**

In ProjectDetailView board:
- Specced entries with scenarios show "Execute ▶" button
- Click shows confirmation dialog: agent name, model, scenario count, estimated cost (model multiplier × rough token estimate)
- Approve → entry moves to executing column, spinner shows
- When complete → entry shows "Verify" button
- Verify view: list of scenarios as checkboxes. Check each one that passes. Submit → verified or back to planned with feedback.

In DashboardView:
- "Ready to Execute" count badge (specced entries with scenarios)
- "Awaiting Verification" count badge (executing-complete entries)

**Scenarios:**
1. Specced entry with 3 scenarios shows "Execute" button in board view
2. Execution preview shows the full agent prompt before approval
3. After agent completes, verify view shows scenario checklist
4. All scenarios pass → entry moves to verified
5. Failed scenario → entry returns to planned with feedback attached

---

### Phase Summary

| Sub-Phase | Core Deliverable | Dependencies | Estimated Size |
|-----------|-----------------|--------------|---------------|
| 4a | Board view + pipeline UI | Existing pipeline API | Medium (1-2 sessions) |
| 4b | Auto-assign projects in classifier | Projects exist | Small (1 session) |
| 4c | AI push-back loop | Session messages + scheduled tasks | Medium (1-2 sessions) |
| 4d | Project-aware agent context | Agent routing exists | Small (1 session) |
| 4e | Execution gate + verification | Board view (4a) | Medium (1-2 sessions) |

### Recommended Build Order

**4a first** — highest value, most visible. Everything else builds on seeing the board.
**4b second** — quick win, prevents manual sorting as new entries come in.
**4d third** — improves agent quality for everything that follows.
**4c fourth** — the automation layer, depends on having good agent context (4d).
**4e last** — the full-circle close, needs everything else working.

### Creation Cycle Review (Phase 4)

| Step | Answer |
|------|--------|
| Intent | Close the loop: entries flow through stages visibly, AI keeps them moving, agents work with project context |
| Covenant | Gated autonomy. AI pushes entries forward but never executes without approval. Cheap models for push-back. |
| Stewardship | brain.exe (Go + Vue). Dev agent builds. Michael reviews each sub-phase. |
| Spiritual Creation | This spec. Precise enough for 4a, directional for 4b-4e. |
| Line upon Line | 5 sub-phases, each stands alone. 4a is the priority. |
| Physical Creation | Dev agent executes against this spec. |
| Review | Scenarios per sub-phase above. |
| Atonement | Each sub-phase is reversible. Board view is additive (list view preserved as toggle). Push-back can be disabled per project. |
| Sabbath | After 4a ships, use it for a few days before building 4b. Natural pause after each sub-phase. |
| Consecration | Michael directly. Eventually: the Agentic OS pattern becomes shareable. |
| Zion | Brain becomes the coordination fabric across all projects. Agents work with project context instead of in isolation. |

### Costs & Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Board view adds JS complexity | Medium | Keep it simple — CSS grid columns, not a drag-and-drop library. Click to advance, not drag. |
| Push-back could annoy if too aggressive | Medium | Configurable per project. Start with 24h staleness. Off by default. |
| Auto-assignment may misclassify | Low | Confidence threshold — only assign when clear. Null is fine. |
| Project context bloats agent prompts | Low | Limit to 20 sibling entries (titles only). Project description ≤500 chars. |
| Scope creep into "full PM tool" | High | This is NOT Jira. No sprints, no story points, no burndown charts. Maturity stages + gated autonomy is the whole model. |
