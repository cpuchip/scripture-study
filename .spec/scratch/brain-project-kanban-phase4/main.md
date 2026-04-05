# Scratch: Brain Project-Kanban Phase 4 — Project Flow + AI Turn Automation

**Created:** 2026-04-05
**Binding problem:** brain.exe has projects, entries, turns, and a pipeline — but nothing ties them together into a managed workflow. There's no project board view, no AI push-back loop driving entries through maturity stages, no auto-assignment of incoming entries to projects, and no way for Michael to see at a glance "what's blocked on me, what's the AI working on, and what's done" across all projects.

---

## Inventory: What Exists (Phase 1-3 shipped)

### Data Layer
- `projects` table: id, name, description, status (active/paused/archived), emoji, created_at, updated_at
- `entries.project_id` FK (nullable)
- `session_messages` table: iterative conversation turns per entry
- `scheduled_tasks` + `task_runs`: recurring automated work
- `entry.maturity`: raw → researched → planned → specced → executing → verified
- `entry.route_status`: "" → suggested → pending → running → complete → accepted/rejected/dismissed/your_turn
- `entry.scenarios`: JSON array of testable conditions (set at specced stage)
- `entry.scratch_path`: workspace path to research/plan scratch file

### Pipeline Layer (Phase 4a-c shipped)
- Research pass: raw → researched (Haiku, writes scratch file)
- Plan pass: researched → planned (Sonnet, structures plan)
- Spec finalization: planned → specced (human sets scenarios, generates proposal)
- `AdvanceRequest`: action (advance/revise/reject/defer), feedback, scenarios
- `pipeline/research.go`: Advance() orchestrates transitions
- Governance docs: classifier-stewardship, research-covenant, plan-covenant

### UI Layer
- DashboardView: project cards, your-turn section, review queue, activity feed
- ProjectsView: project list
- ProjectDetailView: entries for a project (no maturity grouping currently)
- EntriesView: all entries (flat list with badges)
- EntryDetailView: edit, conversation thread, subtasks
- ScheduledView: recurring task CRUD
- LibraryView: agents/skills/memory browser

### Agent Layer
- AgentPool: lazy-creates named agents (study, journal, plan, default, etc.)
- Copilot SDK sessions with MCP tools
- routeEntry(): approve → pending → running → complete
- Session messages for iterative turns

---

## Gap Analysis: What's Missing for the Workflow Michael Described

### Gap 1: Project Board View (Kanban)
Currently ProjectDetailView lists entries flat. No grouping by maturity stage. Michael wants to see entries flowing through stages within a project — like a kanban board where columns are: raw | researched | planned | specced | executing | verified.

### Gap 2: AI Push-Back on Under-Specified Entries
When an entry is "raw" or "researched" and assigned to a project, nothing happens until Michael manually triggers pipeline advancement. The AI should proactively:
- Analyze raw entries and ask clarifying questions via session messages
- Mark entry as "your_turn" with specific questions
- When Michael replies, AI advances the entry

### Gap 3: Auto-Assignment of New Entries to Projects
Classifier assigns category but not project. Michael had to manually sort 67 entries. The classifier should suggest a project based on:
- Entry content similarity to existing project entries
- Project name/description matching
- Fallback: leave unassigned for manual sorting

### Gap 4: Project-Level Status Rollup
Dashboard shows individual entries but no per-project progress. Missing:
- How many entries at each maturity stage per project
- What's blocked on human input
- What's in-flight with agents
- Overall project health/momentum

### Gap 5: Gated Autonomy Flow
The "auto-execute specced items" flow doesn't exist yet. The pipeline can research and plan, but there's no trigger to say "this entry is fully specced with scenarios, execute it." Currently that all happens manually in VS Code.

### Gap 6: Agent Context About Files
Entries get routed to agents but agents don't know what files they've edited, what other entries in the same project look like, or what the project's context file says. Missing project-aware context injection.

---

## What Michael Said (this session)

1. "I'm not seeing any way in the UI to have any sort of project management and AI turn flow to keep tasks moving"
2. "a UI workflow that will really unlock this multi project workflow for me with multi agent working"
3. "enough context on what the agents are doing, the files they're editing"
4. "getting/giving my input at each gate"
5. "we might want to look at adding to the classifier a project it could be assigned to"

---

## Design Decisions

### Decision: Kanban columns = maturity stages (not custom columns)
The maturity ladder (raw → researched → planned → specced → executing → verified) is already the right progression. No need for custom kanban columns — maturity IS the column. Plus "your_turn" and "agent_turn" overlays on any column.

### Decision: Push-back via scheduled task, not real-time
Rather than monitoring entries continuously, a scheduled "Project Review" task runs (e.g. every 4 hours) that:
- Scans projects for raw/researched entries that haven't been touched in 24h+
- Generates session messages with questions
- Marks them your_turn
This is cheaper and more predictable than real-time monitoring.

### Decision: Project auto-assignment in classifier is a separate concern
Add project suggestion to classification result. Simple approach: embed project names/descriptions in the classifier prompt and ask it to suggest a project_id. If confidence is low, leave null. This can ship independently.

### Decision: Phase 4 has sub-phases
Given Mosiah 4:27, break Phase 4 into deliverable chunks:
- 4a: Project Board View (UI only — uses existing data)
- 4b: Project auto-assignment in classifier
- 4c: AI push-back loop (scheduled task + session messages)
- 4d: Agent context injection (project-aware prompts)
- 4e: Execution gate (specced → executing flow in UI)

---

## Architecture Notes

### Project Board View (4a)
Rewrite ProjectDetailView to show a kanban-style layout:
- Columns: raw | researched | planned | specced | executing | verified
- Cards: entry title, agent badge, your_turn/agent_turn indicator
- Drag to advance (calls pipeline advance endpoint) — or click to advance
- "Your Turn" entries highlighted with amber border
- Click card → expand inline or navigate to EntryDetailView
- Stats bar: total entries, blocked-on-you count, in-flight count

### Project Auto-Assignment (4b)
In classifier.go, after classification:
- Load project names + descriptions
- Add to prompt: "Given these projects: [...], suggest which project this entry belongs to. Return project_id or null."
- Store in classified Result, saved to entry.project_id
- Allow override in UI

### AI Push-Back Loop (4c)
New scheduled task type: "project_review"
- Iterates entries with maturity="raw" or "researched" that are 24h+ stale
- For "raw": AI reads entry body, generates 2-3 clarifying questions, posts as session messages
- For "researched": AI reads scratch file, identifies gaps, asks specific questions
- Marks entry your_turn
- When Michael replies, a handler detects the reply and triggers pipeline advance

### Agent Context Injection (4d)
When routing an entry to an agent:
- Load project context: name, description, other entries in same project (titles + maturity)
- Load entry context: conversation history, scratch file path, scenarios
- Inject into agent prompt template:
  "You are working on the project '{project.Name}': {project.Description}
   Other entries in this project: [list]
   This entry's pipeline context: [research findings, plan]"

### Execution Gate (4e)
In ProjectDetailView, entries at "specced" stage with scenarios show an "Execute" button.
- Click → confirms agent, shows cost estimate (model × estimated tokens)
- Approve → creates pipeline advance to "executing", routes to agent
- Agent runs with full context (project + spec + scenarios)
- On completion → entry moves to "verified pending review"
- Review: human checks output against scenarios → accept/reject
