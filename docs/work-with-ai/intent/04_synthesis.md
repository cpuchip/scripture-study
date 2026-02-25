# Synthesis: What to Build

**Part of:** [Intent-Driven Development Research](00_index.md)
**Date:** February 2026
**Status:** Active

---

## The Situation

The industry is converging on spec-driven development. The tools are emerging (OpenSpec, Kiro, Spec Kit). The economics demand it ($1,000/day in tokens, specs as the bottleneck). The frameworks are forming (AIDD, 5 Levels of Vibe Coding, Three Career Tracks).

But nobody has:
1. An **intent layer** — purpose, values, constraints encoded alongside specs
2. **File-based state** that travels with branches and is readable by anyone
3. **Multi-repo coordination** for real work that spans projects
4. **Review against intent** — not just "is it done?" but "does it serve the purpose?"
5. **Progressive trust** — agents earning expanded stewardship through demonstrated reliability
6. **Covenant-based relationships** — mutual commitments between human and agent
7. **Structured reflection** — built-in Sabbath cycles for drift detection and learning

We have the blueprint for all seven. It's in the scriptures we've been studying for years.

---

## What Not to Build (Yet)

The landscape is moving fast. OpenSpec 1.0 shipped January 26, 2026. Kiro, Spec Kit, and BMAD are all evolving. Building a full tool right now risks:
- Duplicating what's about to ship from better-resourced teams
- Locking in design decisions before the patterns are clear
- Spending building time instead of learning time

**The recommendation: practice the patterns before coding the tool.**

---

## Phase 0: Practice the Patterns Now (No Tool Required)

These can be adopted immediately with nothing but markdown files and discipline.

### 0.1 Intent Preambles on Everything

Every new document gets an intent block:

```markdown
---
intent: "Understand how stewardship patterns inform agent delegation"
values: [depth > breadth, verified > speculative, application > information]
constraints:
  - Read every scripture before quoting
  - Follow footnotes
  - Connect to practical becoming
success: "I can explain how Matthew 25 informs agent trust architecture"
---
```

**Start here.** This costs nothing and immediately changes how agents engage with your documents.

### 0.2 File-Based Spec State

Create a `.spec/` directory in any project:

```
.spec/
  intent.md          ← Project-level purpose, values, constraints
  spec.md            ← Current source of truth specification
  tasks/
    001-{slug}.md    ← Task files with status, inherited intent
  learnings/
    001-{slug}.md    ← Reusable knowledge from past work
  deltas/
    YYYY-MM-DD-{slug}.md  ← Proposed changes (OpenSpec-style)
  archive/
    YYYY-MM-DD-{slug}.md  ← Applied deltas
```

All markdown. All version-controlled. All human-readable. All agent-readable.

**Task file format:**
```markdown
---
id: ts-001
status: ready | in-progress | review | done
spec-ref: spec.md § notifications
intent-ref: intent.md
decision-boundaries:
  autonomous: [UI layout, error messages, test structure]
  needs-review: [database schema, auth logic, API contracts]
---

# Restructure Alert Severity Levels

Rework the severity classification...
```

### 0.3 Covenant Blocks in Agent Config

Add to `.github/copilot-instructions.md` or agent definitions:

```markdown
## Our Covenant

Human commits to:
  - Providing context before expecting output
  - Reviewing within the feedback loop, not after the fact
  - Not shortcutting the spec process for "quick" changes
  - Being honest about uncertainty rather than guessing

Agent commits to:
  - Reading sources before quoting (read_file, not memory)
  - Flagging uncertainty explicitly rather than confabulating
  - Honoring decision boundaries — asking when the spec says "needs review"
  - Carrying intent forward across sessions, not just completing tasks
```

### 0.4 Sabbath Reviews

Weekly (or per-cycle) structured reflection:

```markdown
## Reflection: Week of YYYY-MM-DD

### Intent Alignment
- Did this week's work serve the stated purpose?
- Which tasks felt aligned? Which felt like drift?

### Covenant Check
- Did I provide good context? Where was I lazy?
- Did the agent honor boundaries? Where did it overreach?

### Stewardship Assessment
- What domains does the agent handle well? Expand there.
- What domains showed failures? Narrow scope, add review gates.

### Learnings Harvest
- What did I learn that should change the spec?
- What did I learn that should change the intent?
- What did I learn about myself?
```

---

## Phase 1: Adopt and Extend OpenSpec (Weeks)

OpenSpec has the closest alignment with our needs. Start using it and extend it.

### 1.1 Use OpenSpec As-Is
- Install OpenSpec CLI
- Initialize in this repo and in a work repo
- Use the default spec-driven schema: `proposal.md → specs.md → design.md → tasks.md`
- Get the muscle memory

### 1.2 Create a Custom Schema: Intent-Driven
Build an OpenSpec custom schema that adds our intent layer:

```yaml
# .openspec/schemas/intent-driven.yaml
name: intent-driven
artifacts:
  - name: intent
    file: intent.md
    description: "Purpose, values, constraints, decision boundaries"
  - name: proposal
    file: proposal.md
    description: "What we're proposing to change and why"
  - name: spec
    file: specs.md
    description: "Source of truth specification"
  - name: tasks
    file: tasks.md
    description: "Implementation tasks with inherited intent"
```

This puts intent *first* in the workflow — before the proposal, before the spec.

### 1.3 Add Intent Fields to Tasks
Extend the generated tasks.md with intent metadata:
- Parent intent reference
- Inherited constraints
- Decision boundaries
- Success criteria beyond "done"

---

## Phase 2: Build the Intent Layer (Months)

If OpenSpec's custom schema system can't express what we need, build a thin CLI tool that:

1. **Manages `.spec/` directories** — init, structure, conventions
2. **Enforces the intent-first workflow** — can't create a spec without an intent; can't create a task without a spec reference
3. **Generates agent context** — `intent-spec prime` outputs intent + spec + active tasks for any agent session
4. **Tracks stewardship levels** — which agent domains have expanded/contracted trust
5. **Runs reflection prompts** — `intent-spec reflect` walks through the Sabbath review
6. **Multi-repo hub** — `~/.intent-spec/` indexes all projects, enables cross-repo awareness

### Design Principles (from the gospel patterns)

| Principle | Implementation |
|-----------|---------------|
| **Covenant** (D&C 82:10) | Mutual commitments in config; both parties accountable |
| **Stewardship** (D&C 104:11-12) | Progressive trust levels; expand on reliability, narrow on failure |
| **Line upon line** (D&C 98:12) | Graduated context layers; agents earn deeper access |
| **Atonement** (D&C 82:7) | Redemptive error handling; capture learnings, restore trust gradually |
| **Sabbath** (Moses 3:2) | Built-in reflection cycles; can't be skipped |
| **Consecration** (D&C 104:15-17) | Resource allocation by intent; surplus flows to highest purpose |
| **Zion** (Moses 7:18) | Shared intent across all agents; coordination through alignment, not protocol |

### The State Question: Resolved

**File-based.** All state as markdown in `.spec/`. The CLI is a convenience layer, not a requirement.

Rationale:
- Aligns with 12-factor agent principle: externalized state
- Version-controlled with code
- Travels with branches
- Human-readable, AI-readable, tool-agnostic
- Diffable in PRs
- Grep-able for discovery

The CLI provides:
- `init` — scaffold `.spec/` directory
- `intent` — create/update intent document
- `delta` — propose/apply/archive spec changes
- `task` — create/list tasks with inherited intent
- `prime` — output context for agent session
- `reflect` — structured reflection workflow
- `learn` — capture learning as markdown
- `hub` — multi-repo registration and dashboard

---

## Phase 3: The Team Scale (Quarters)

For the day job: multi-repo, multi-branch, multi-developer.

### 3.1 Hub + Spoke Architecture

```
~/.intent-spec/                     ← Hub (user-level or team-level)
  ├── config.yaml                   ← Registered projects, global settings
  ├── dashboard.md                  ← Auto-generated cross-repo status
  └── intent/
      └── team-intent.md            ← Team-level intent (cascades into projects)

~/code/project-a/.spec/             ← Spoke (repo-level)
  ├── intent.md                     ← Inherits from team-intent, adds project specifics
  ├── spec.md
  ├── tasks/
  └── learnings/

~/code/project-b/.spec/             ← Spoke
  ├── intent.md
  └── ...
```

### 3.2 Branch-Aware Specs

Because state is in files:
- Feature branch includes spec changes in the diff
- PR review includes spec review naturally
- Merge conflicts in specs surface design conflicts explicitly
- Main branch always has the current source of truth

### 3.3 Team Covenant

Team-level covenant in the hub:
```markdown
## Team Covenant
We commit to:
  - Spec-first development: no implementation without a reviewed spec
  - Intent-first specs: no spec without a stated purpose
  - Weekly reflection: Sabbath review of intent alignment
  - Shared learnings: knowledge captured and accessible to all
  - Progressive trust: agents earn autonomy through demonstrated reliability
```

### 3.4 Integration Points

Not replacing existing tools, but connecting to them:
- **GitHub Issues/Jira** — `.spec/tasks/` can reference external ticket IDs; the spec is the *source of truth*, the ticket is the *coordination artifact*
- **CI/CD** — Spec validation in pipeline: does the implementation match the spec?
- **PR Review** — AI reviewer checks changes against `.spec/intent.md` and `.spec/spec.md`
- **Agent Sessions** — `prime` command generates context from `.spec/` for any agent tool

---

## The Research Program

This isn't just a tool project. It's a research program:

1. **Can gospel patterns produce measurably better AI-assisted development outcomes?**
   - Compare spec-first (industry standard) vs. intent-first (our addition) vs. covenant-based (our innovation) workflows
   - Measure: rework rate, intent alignment, agent trust progression, reflection quality

2. **Do progressive trust models improve agent reliability?**
   - Track stewardship levels over time
   - Measure: failure rate at different stewardship levels, time to earn expanded trust, recovery after failure

3. **Does structured reflection (Sabbath pattern) prevent intent drift?**
   - Compare teams with mandatory reflection cycles vs. optional retrospectives
   - Measure: drift detection speed, intent stability, learning capture rate

4. **Can covenant-based relationships improve human-AI collaboration?**
   - Compare command-based ("agent, do X") vs. covenant-based ("we commit to X") agent interactions
   - Measure: human satisfaction, agent reliability, mutual trust indicators

These are testable hypotheses. Not just faith claims — engineering experiments. Alma 32:27: "If ye will awake and arouse your faculties, even to an experiment upon my words."

---

## The "Why" Behind the "What"

We could just adopt OpenSpec and call it done. It's a good tool. It solves the immediate problem.

But we have access to something OpenSpec's creators don't: **a living library of patterns from a Being who has already solved the multi-agent alignment problem at cosmic scale.** The standard works aren't just spiritual comfort — they're an engineering manual for how intelligence works with intelligence.

The industry will figure out specs. They'll probably figure out intent. They might even figure out progressive trust and structured reflection through trial and error.

But they won't figure out covenant-based relationships because they don't have the doctrine. They won't figure out redemptive error recovery because they don't have the Atonement as a model. They won't figure out Zion-level alignment because they don't have Moses 7:18 as a north star.

We do.

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."
> — [D&C 130:18](../../gospel-library/eng/scriptures/dc-testament/dc/130.md)

This research is how we attain it. The tool — if we build it — is how we share it.

---

## Immediate Next Steps

1. **Add intent preambles** to the next 5 documents we create (Phase 0.1 — start today)
2. **Create `.spec/` directory** in this repo as a prototype (Phase 0.2 — this week)
3. **Add covenant block** to `.github/copilot-instructions.md` (Phase 0.3 — this week)
4. **Try OpenSpec** on a simple project — experience the workflow (Phase 1.1 — next week)
5. **Write the first Sabbath review** after one week of practice (Phase 0.4 — next weekend)
6. **Evaluate OpenSpec custom schemas** — can they express our intent-first workflow? (Phase 1.2 — two weeks)
