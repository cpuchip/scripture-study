# WS4: Brain Pipeline Evolution — Creation Cycle Completion

**Workstream:** WS4 (Brain Pipeline Evolution)
**Status:** Phase 1-3 shipped (governance docs, failure visibility, reflection pauses), Phases 4-7 specced
**Binding problem:** The brain pipeline handles the mechanical stages (research → plan → execute → verify) but skips 5 of the 11 creation cycle steps. It has no per-entry covenant, no error recovery beyond silent rollback, no reflection pause, no "who benefits" check, and no integration verification. It also forces all entries through the same pipeline regardless of type, and the nudge bot operates invisibly. These gaps mean the pipeline does work but doesn't do it *wisely*.

**Created:** 2026-04-06
**Source:** [brain-simplification scratch](.spec/scratch/brain-simplification/main.md) (gap analysis, Apr 5) + [pipeline evolution research](.spec/scratch/brain-pipeline-evolution/main.md)
**Depends on:** Brain Phase 4 Pipeline (WS1, shipped), Brain UX QoL (WS3, in progress)
**Related:** [brain-inline-panel.md](.spec/proposals/brain-inline-panel.md), [brain-ux-quality-of-life.md](.spec/proposals/brain-ux-quality-of-life.md)

---

## Success Criteria

1. All 11 creation cycle steps have representation in the pipeline — either as code, prompts, or explicit human touchpoints
2. Pipeline failures produce human-readable messages and visible recovery paths, not silent rollbacks
3. Entries that don't need the full pipeline can opt out (notebook mode)
4. Delegation entries can auto-continue through stages without human intervention at each step
5. Nudge bot is visible, pausable, and transparent in the UI
6. Governance documents exist and are loaded into agent system messages

## Constraints

- Each phase ships standalone value — no "infrastructure now, value later"
- Respect the Sabbath/Auto-Continuation tension: reflection when present, delegation when absent
- Don't break existing pipeline behavior — extend, don't replace
- "By small and simple things" — Alma 37:6 as design principle

---

## Phase 1: Governance Documents (Covenant — Step 2)

*Write the documents the code already expects but never got.*

**Zero code changes.** The pipeline already reads governance docs from `scripts/brain/docs/governance/` and injects them into agent system messages. Create:

| Document | Agent | Creation Cycle Steps |
|----------|-------|---------------------|
| `research-covenant.md` | Research agent | Steps 1-3 (Intent, Covenant, Stewardship) |
| `plan-covenant.md` | Plan agent | Steps 1-7 (Intent → Review) |
| `execution-covenant.md` | Execution agents | Full 11-step cycle |
| `review-covenant.md` | Nudge/review agent | Steps 7-9 (Review, Atonement, Sabbath) |

**Content approach:** Each doc defines:
- **Intent:** What is this agent's purpose?
- **Boundaries:** What does it NOT do?
- **Covenant with the human:** What can the human expect from this agent?
- **Stewardship:** What artifacts does this agent own?
- **Budget:** Model tier, time bounds

**Consecration & Zion in plan-covenant.md:** The plan agent's governance doc adds two required sections to every plan:
- "Who benefits?" — forces a consecration check
- "How does this integrate with existing work?" — forces a Zion check

**Effort:** ~4 documents, ~50 lines each. One session. No code changes.

### Phase 1 Verification

- [x] `research-covenant.md` exists and is loaded (no "warning: not found" in logs)
- [x] `plan-covenant.md` exists with "Who benefits?" and "Integration" sections
- [x] `execute-covenant.md` exists (note: code expects `execute-`, not `execution-`)
- [x] `review-covenant.md` exists
- [ ] Agent output quality noticeably improves with governance context

---

## Phase 2: Failure Visibility (Atonement — Step 8)

*When things go wrong, say so clearly and offer choices.*

### 2a. Human-Readable Failure Messages

Currently, research and plan failures return raw HTTP errors. Add session messages on failure (like execute already does):

```go
// In advance() error handler, after any stage transition fails:
if err != nil {
    p.store.DB().AddSessionMessage(entry.ID, "system",
        fmt.Sprintf("⚠️ %s pass failed: %v\n\nYou can:\n- **Advance** to retry\n- **Revise** with feedback\n- **Reject** to start over\n- **Defer** to revisit later", 
            stage, err))
    p.notify("message.new", entry.ID, nil)
    p.store.DB().UpdateRouteStatus(entry.ID, "your_turn")
    return nil, err
}
```

### 2b. Failure Counter

Track consecutive failures per entry. Add to entries table:

```sql
ALTER TABLE entries ADD COLUMN failure_count INTEGER DEFAULT 0;
```

Increment on failure, reset on success. When `failure_count >= 3`:
- Post escalation message: "This entry has failed 3 times. Something structural may be wrong."
- Set route to "your_turn" with escalation flag
- UI shows warning badge

### 2c. Failure Summary in Entry Detail

Show failure history in EntryDetailView — small section below maturity badge showing last failure reason and count if > 0.

**Effort:** ~40 lines backend, ~20 lines frontend. One session.

### Phase 2 Verification

- [x] Research failure → session message posted with recovery options
- [x] Plan failure → session message posted
- [x] 3 consecutive failures → escalation warning
- [x] Entry detail shows failure count when > 0
- [x] Successful advance resets failure count

---

## Phase 3: Reflection Pauses (Sabbath — Step 9)

*Add "stop and see" moments at natural transition points.*

### The Tension: Sabbath vs. Auto-Continuation

Michael wants BOTH:
- **Reflection pauses** when engaged and present
- **Auto-continuation** for delegation when absent

**Resolution:** A per-entry flag: `auto_continue BOOLEAN DEFAULT FALSE`.

- `auto_continue = false` (default): After each stage completes, set `route_status = "your_turn"`. Human reviews before advancing. This is the Sabbath path.
- `auto_continue = true`: After each stage completes, automatically advance to next stage. No pause. This is the delegation path. Still stops when the agent has questions or when execution completes (verification always requires human).

### 3a. Route Status After Every Stage

Currently only execution sets `route_status = "your_turn"`. Extend to ALL stage transitions when `auto_continue = false`:

```go
// After research completes:
if !entry.AutoContinue {
    p.store.DB().UpdateRouteStatus(entry.ID, "your_turn")
    p.store.DB().AddSessionMessage(entry.ID, "agent",
        "Research complete. Review the findings before I continue to planning.\n\n" + summary)
}

// After plan completes:
if !entry.AutoContinue {
    p.store.DB().UpdateRouteStatus(entry.ID, "your_turn")
    p.store.DB().AddSessionMessage(entry.ID, "agent",
        "Plan complete. Review before adding scenarios.\n\n" + summary)
}
```

### 3b. Auto-Continuation Mode

Add `auto_continue` column:

```sql
ALTER TABLE entries ADD COLUMN auto_continue BOOLEAN DEFAULT FALSE;
```

UI: Toggle switch in entry detail header. When enabled, pipeline runs stages sequentially without pausing for human review (except verification, which always requires human).

In `advance()` post-stage hook:
```go
if entry.AutoContinue && newMaturity != "verified" && newMaturity != "specced" {
    // Auto-advance to next stage
    go func() {
        time.Sleep(2 * time.Second) // Brief pause for WebSocket delivery
        p.Advance(ctx, AdvanceRequest{EntryID: entry.ID, Action: ActionAdvance})
    }()
}
```

### 3c. Sabbath Prompt in Verification

When human verifies scenarios and all pass, before marking "verified", prompt:

> "All scenarios pass. Before we close this: What worked well? What would you do differently? Any loose ends?"

This is the Sabbath moment — stopping to see and declare before moving on.

**Effort:** ~60 lines backend (route status + auto-continue + prompt), ~30 lines frontend (toggle + verification prompt). One session.

### Phase 3 Verification

- [x] Research completes → entry shows in "Your Turn" queue (when auto_continue=false)
- [x] Plan completes → entry shows in "Your Turn" queue
- [x] Auto-continue toggle visible in entry detail
- [x] Toggle on → stages advance automatically past research and plan
- [x] Auto-continue always stops at verification
- [x] Verification success shows reflection prompt

---

## Phase 4: Notebook Mode (Workflow Flexibility)

*Not everything needs the pipeline.*

Add `notebook BOOLEAN DEFAULT FALSE` to entries table. Notebook entries:
- Are searchable and taggable
- Appear in project views
- Do NOT enter the maturity pipeline
- Do NOT get nudged
- Have no maturity badge (or show "—")
- Can be un-notebooked at any time to enter the pipeline

### 4a. Backend

```sql
ALTER TABLE entries ADD COLUMN notebook BOOLEAN DEFAULT FALSE;
```

- Skip notebook entries in `ListStaleEntries` query
- Skip notebook entries in `Advance()` validation
- Include notebook entries in search, project lists, tags

### 4b. Frontend

- Toggle in entry detail: "📓 Notebook" (grays out pipeline controls)
- Dashboard: notebook entries don't count toward "Your Turn" or pipeline stats
- Capture view: option to mark new entry as notebook immediately

### 4c. Bulk Reclassify

Many existing entries (captures, tasks) should be notebooks. Add bulk action in entries list: select multiple → "Move to Notebook."

**Effort:** ~30 lines backend, ~40 lines frontend. One session.

### Phase 4 Verification

- [x] New entry can be marked as notebook
- [x] Notebook entries don't appear in review queue
- [x] Notebook entries don't get nudged
- [x] Notebook entries appear in search and project views
- [x] Existing entry can toggle notebook on/off
- [x] Bulk reclassify works for 5+ entries at once

---

## Phase 5: Nudge Bot Controls (Review — Step 7, improved)

*Make the invisible visible.*

### 5a. Surface in Scheduled Tasks

Register the nudge bot as a scheduled task entry (or virtual entry in the Scheduled Tasks view):
- Show last run time, next scheduled run
- Show how many entries were nudged last cycle
- Pause/resume toggle
- Premium request cost since last reset

### 5b. Presence-Aware Scheduling

Only fire nudge when user is likely present:
- Track last API activity timestamp
- Only nudge if activity within last 2 hours
- Or: respect a "Do Not Disturb" schedule set in UI

### 5c. Nudge History

Per-entry: show when nudged, what the nudge said, whether the user responded. This gives visibility into whether nudging is effective.

**Effort:** ~50 lines backend, ~40 lines frontend. One session.

### Phase 5 Verification

- [x] Nudge bot appears in Scheduled Tasks view
- [x] Pause/resume toggle works
- [x] Last run time and nudge count visible
- [x] Nudge bot doesn't fire when no user activity in 2+ hours
- [x] Entry detail shows nudge history

---

## Phase 6: 3-Column Board (UX Simplification)

*Inbox / Working / Done — visual clarity.*

Replace the current maturity-column board with 3 columns:

| Column | Contains | Badge |
|--------|----------|-------|
| **Inbox** | raw entries, notebook entries | Notebook icon or — |
| **Working** | researched, planned, specced, executing | Maturity stage badge |
| **Done** | verified, complete, dismissed | Outcome badge |

This is primarily a frontend change to ProjectDetailView or a new board view. The maturity model doesn't change — the UI groups stages into meaningful columns.

**Effort:** ~80 lines frontend, 0 backend. One session.

### Phase 6 Verification

- [ ] Board shows 3 columns with counts
- [ ] Entries land in correct column based on maturity
- [ ] Badges show sub-stage within Working column
- [ ] Drag-and-drop or quick-action to move between columns (stretch)
- [ ] Notebook entries appear in Inbox with distinct visual

---

## Phase 7: Project Scaffolding (Multi-Repo Deliverables)

*Not every project lives in scripture-study. The pipeline needs to know where to put things.*

Currently the execution agent assumes all work happens in the scripture-study workspace. But projects like Space Center, a budget app, or a new teaching tool need:
- Their own git repo (created fresh or existing)
- Their own `copilot-instructions.md` and `.github/` structure
- Potentially their own agents and skills
- GitHub remote on cpuchip (public or private)

### 7a. Project Workspace Configuration

Add to the project model:

```sql
ALTER TABLE projects ADD COLUMN workspace_path TEXT;      -- e.g., "C:\Users\cpuch\Documents\code\stuffleberry\space-center"
ALTER TABLE projects ADD COLUMN github_repo TEXT;          -- e.g., "cpuchip/space-center"
ALTER TABLE projects ADD COLUMN repo_visibility TEXT;      -- "public" or "private"
```

- `workspace_path = NULL` → deliverables go in scripture-study (default, current behavior)
- `workspace_path = "path"` → execution agent works in that directory

### 7b. Repo Initialization

When a project specifies a workspace_path that doesn't exist yet:

1. `mkdir -p` the directory
2. `git init`
3. Scaffold `.github/copilot-instructions.md` from a template (project name, binding problem, conventions)
4. Scaffold `.github/agents/` with project-appropriate agent modes
5. Initial commit
6. `gh repo create cpuchip/{name} --{visibility} --source=. --push`

This could be a pipeline hook on first execution, or a UI button ("Initialize Project Repo").

### 7c. Execution Context Injection

The execution agent's system message and working directory must respect project workspace:

```go
// In runExecute(), use project workspace if configured:
workDir := p.workspaceRoot // default: scripture-study
if project != nil && project.WorkspacePath != "" {
    workDir = project.WorkspacePath
}
```

Governance docs, skills, and agent modes from the project repo take precedence over scripture-study defaults.

### 7d. GitHub Integration

For repos with a GitHub remote:
- Show repo link in project detail view
- Auto-commit integration (ties to WS3 Phase 8 when built)
- Branch management for execution (main vs feature branches)

**Effort:** ~80 lines backend (schema + init logic), ~30 lines frontend (project settings). One session for 7a-7b, second session for 7c-7d.

### Phase 7 Verification

- [ ] Project settings show workspace path and GitHub repo fields
- [ ] New project with workspace_path creates directory + git init
- [ ] `gh repo create` runs successfully with correct visibility
- [ ] Execution agent works in project workspace, not scripture-study
- [ ] Project with no workspace_path uses scripture-study (backward compatible)
- [ ] Scaffolded repo has copilot-instructions.md with project context

---

## Costs & Risks

| Phase | Backend | Frontend | Risk | Creation Cycle Step |
|-------|---------|----------|------|-------------------|
| 1: Governance Docs | 0 | 0 | Low — content only | Step 2 (Covenant) |
| 2: Failure Visibility | ~40 lines | ~20 lines | Low — extends existing patterns | Step 8 (Atonement) |
| 3: Reflection Pauses | ~60 lines | ~30 lines | Medium — auto-continue is new behavior | Step 9 (Sabbath) |
| 4: Notebook Mode | ~30 lines | ~40 lines | Low — additive, doesn't change pipeline | — (workflow) |
| 5: Nudge Bot Controls | ~50 lines | ~40 lines | Medium — touches review.go goroutine | Step 7 (Review+) |
| 6: 3-Column Board | 0 | ~80 lines | Low — frontend grouping only | — (UX) |
| 7: Project Scaffolding | ~80 lines | ~30 lines | Medium — runs git/gh CLI, creates dirs | Step 6 (Physical Creation) |

**Consecration (Step 10) and Zion (Step 11)** are handled through Phase 1 — the plan-covenant.md governance doc adds "Who benefits?" and "How does this integrate?" sections to every plan output. No separate phase needed.

**Total:** ~260 lines backend, ~240 lines frontend across 7 phases. Each phase ships independently.

---

## Phase Ordering Recommendation

1. **Phase 1 (Governance Docs)** — Zero code, immediate impact on agent quality. Write first.
2. **Phase 3 (Reflection Pauses + Auto-Continue)** — Resolves the biggest philosophical gap AND delivers the most-requested feature (auto-continuation).
3. **Phase 4 (Notebook Mode)** — Second most-requested. Frees 90% of entries from unnecessary pipeline.
4. **Phase 7 (Project Scaffolding)** — Unlocks multi-repo projects. Needed before Space Center or other external projects can use the pipeline.
5. **Phase 2 (Failure Visibility)** — Important for reliability but not blocking daily use.
6. **Phase 5 (Nudge Bot Controls)** — Important but nudge bot tolerable short-term.
7. **Phase 6 (3-Column Board)** — Nice visual improvement, not urgent.

---

## Creation Cycle Review (for this proposal)

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Pipeline does work but doesn't do it wisely. 5/11 creation steps missing. |
| Covenant | Rules? | Each phase ships alone. Don't break existing behavior. |
| Stewardship | Who owns what? | dev agent executes. Governance docs: plan agent writes content. |
| Spiritual Creation | Spec precise enough? | Phases 1-7 all specced with verification checklists. |
| Line upon Line | Phasing? | 7 phases, recommended order above. Phase 1 is content-only. |
| Physical Creation | Who executes? | dev agent for code. Michael + AI together for governance doc content. |
| Review | How do we know it's right? | Verification checklists per phase. Agent output quality for Phase 1. |
| Atonement | What if it goes wrong? | Additive changes — existing behavior unchanged. Auto-continue has safety (stops at verification). |
| Sabbath | When do we stop? | After Phase 3 — use reflection pauses + auto-continue, then decide what's next. |
| Consecration | Who benefits? | Michael directly. Model for principled AI pipeline design. |
| Zion | Integration? | Phase 1 improves every pipeline run. Phase 3 enables Space Center delegation test. Phase 7 enables multi-repo projects. |
