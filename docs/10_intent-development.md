# Intent Development Plan

**Date:** 2026-02-24
**Status:** Active
**Related:** [docs/work-with-ai/04_intent-engineering.md](work-with-ai/04_intent-engineering.md)
**Video:** [Nate B Jones — "Prompt Engineering Is Dead. Context Engineering Is Dying."](https://www.youtube.com/watch?v=QWzLPn164w0)
**Tool Under Evaluation:** [TPG](../external_context/tpg/README.md) — issue tracker for AI agents

---

## Problem Statement

We have three intersecting problems:

### 1. Markdown Flood (Spec Document Chaos)
This repo has 200+ markdown files across `study/`, `docs/`, `lessons/`, `becoming/`, `scripts/plans/`, `callings/`, etc. They are:
- **Unorganized** — no consistent taxonomy, no relationship tracking between files
- **Disconnected** — a study document doesn't link back to the lesson it informed; a plan doesn't link to the tasks that implement it
- **Ephemeral** — no lifecycle (draft → active → archived), no staleness detection
- **Invisible to agents** — agents discover files through search, not through structured navigation

### 2. Task ↔ Intent Disconnect 
TPG tracks *what to do* excellently. It does not track *why we're doing it*. Tasks have titles and descriptions but no structured connection to:
- Which spec/study document generated them
- What strategic goal they serve
- What "success" looks like beyond "done"
- What decision boundaries the agent should honor

### 3. No Intent-Based Workflow
The video's three layers map to our infrastructure:

| Layer | Video's Term | What We Have | What's Missing |
|-------|-------------|-------------|----------------|
| 1 | Unified context infrastructure | `gospel-library/`, MCP servers, `.github/` instructions | No structured spec/document management |
| 2 | Coherent worker toolkit | `.github/agents/`, work-with-ai methodology | Workflows aren't encoded in tools, only in docs |
| 3 | Intent engineering | Moses 1:39 as a concept; some values in instructions | No machine-readable intent, no decision boundaries, no value hierarchies |

---

## Vision: Intent-Based Development Lifecycle

```
INTENT (Why?)
  "Build tools that facilitate deep scripture study leading to genuine becoming"
  ├── Values: depth over breadth, faith as framework, honest exploration
  ├── Constraints: read-before-quoting, verify everything, human holds discernment
  └── Measures: did it produce insight? did it lead to practice? did the person grow?

SPEC (What?) ← Documents in organized folders with lifecycle tracking
  ├── Discovery (study docs, evaluations)
  ├── Design (plans, architecture docs)
  └── Active specs linked to task groups

TASK (How?) ← TPG with intent metadata
  ├── Each task links to its spec
  ├── Each task carries intent context (why, constraints, success criteria)
  └── Dependencies respect intent hierarchy

EXECUTE (Do it)
  ├── Agents operate with intent + spec + task context
  └── Decision boundaries encoded per-task or per-epic

REVIEW (Did it serve the intent?)
  ├── Not just "is it done?" but "did it accomplish what we intended?"
  ├── Becoming: what changed in the person, not just the codebase?
  └── Feedback to intent layer (refine values, update constraints)
```

This maps directly to the gospel pattern we discovered:

| Phase | Gospel Pattern | Scripture |
|-------|---------------|-----------|
| **Intent** | "This is my work and my glory" | [Moses 1:39](../gospel-library/eng/scriptures/pgp/moses/1.md) |
| **Spec** | "Created all things spiritually, before they were naturally" | [Moses 3:5](../gospel-library/eng/scriptures/pgp/moses/3.md) |
| **Task** | "The Gods took counsel among themselves" | [Abraham 4:26](../gospel-library/eng/scriptures/pgp/abr/4.md) |
| **Execute** | "Let us go down... we will take of these materials" | [Abraham 3:24](../gospel-library/eng/scriptures/pgp/abr/3.md) |
| **Review** | "Watched until they obeyed" | [Abraham 4:18](../gospel-library/eng/scriptures/pgp/abr/4.md) |

---

## Phase 1: Document Organization (The Context Infrastructure)

**Goal:** Solve the markdown flood problem. Make our document library navigable by both humans and agents.

### 1.1 Document Registry

Create a structured index that every agent can reference. Not a static README, but either:

**Option A: Index file (simple, immediate)**
A `docs/INDEX.md` mapping every significant document with metadata:
```markdown
| Document | Type | Status | Intent | Last Updated |
|----------|------|--------|--------|-------------|
| study/charity.md | study | active | understand charity to practice it | 2026-02 |
| scripts/plans/06_becoming-app.md | plan | active | build app that bridges study→becoming | 2026-01 |
```

**Option B: TPG-managed specs (deeper, requires TPG changes)**
Add a `spec` item type to TPG alongside `task` and `epic`, where specs:
- Have a `path` field pointing to the markdown file
- Link to epics/tasks they generated
- Track lifecycle (draft → active → implemented → archived)
- Carry intent metadata

**Recommendation:** Start with Option A now, evolve to Option B when TPG changes are ready.

### 1.2 Document Taxonomy

Establish clear categories with consistent naming:

| Category | Location | Purpose |
|----------|----------|---------|
| **Studies** | `study/` | Discovery — scripture analysis, topic research |
| **Evaluations** | `study/yt/` | YouTube video evaluations |
| **Lessons** | `lessons/` | Teaching preparation |
| **Becoming** | `becoming/` | Personal transformation — applying insights |
| **Plans** | `scripts/plans/` | Architecture/feature specs |
| **Meta** | `docs/` | Process reflection, methodology, this document |
| **Work-with-AI** | `docs/work-with-ai/` | Teaching series on AI collaboration |
| **Journal** | `journal/` | Daily reflections |

### 1.3 Frontmatter Convention

Add structured frontmatter to new documents (gradually adopt for existing ones):

```yaml
---
type: study | plan | lesson | evaluation | becoming | meta
status: draft | active | archived
intent: "one-line statement of WHY this document exists"
related:
  - study/charity.md
  - becoming/charity.md
tags: [charity, christlike-attributes, moroni-7]
created: 2026-02-24
updated: 2026-02-24
---
```

---

## Phase 2: TPG Intent Layer (The Intent Infrastructure)

Proposed improvements to TPG that add intent-awareness while preserving its strengths.

### 2.1 Intent Metadata on Epics

Epics are the natural container for intent. Add structured fields:

```
tpg epic add "Becoming App Phase 2" \
  --intent "Bridge the gap between study and transformation" \
  --success "Users report daily practices changing their behavior over 30 days" \
  --constraints "Must not gamify spiritual growth; depth over engagement metrics" \
  --values "becoming > features, insight > volume, honest reflection > positive metrics"
```

**Data model additions to Epic:**

| Field | Type | Purpose |
|-------|------|---------|
| `Intent` | string | Why does this work exist? What purpose does it serve? |
| `SuccessCriteria` | string | What does "done right" look like (beyond task completion)? |
| `Constraints` | []string | Non-negotiable boundaries agents must honor |
| `Values` | []string | Trade-off hierarchies (when X conflicts with Y, prefer X) |
| `SpecPath` | string | Path to the spec/plan document that generated this epic |
| `DecisionBoundary` | string | What decisions can agents make? What requires human input? |

### 2.2 Spec-to-Task Traceability

Connect the markdown flood to the task system:

```
tpg spec register scripts/plans/06_becoming-app.md --epic ep-abc
tpg spec list                    # Show all registered specs and their task coverage
tpg spec coverage ep-abc         # "12/15 requirements have associated tasks"
```

This solves the "which spec did this task come from?" problem and enables asking "are all requirements covered?"

### 2.3 Intent-Aware `tpg prime`

Modify the `tpg prime` output to include intent context when an agent starts work:

```
## Current Intent
Epic: Becoming App Phase 2
Intent: Bridge the gap between study and transformation
Constraints:
  - Must not gamify spiritual growth
  - Depth over engagement metrics
Values: becoming > features, insight > volume
Decision Boundary: UI changes autonomous; data model changes need human review

## Ready Tasks (under this intent)
- ts-abc: Add reflection prompt after study completion
- ts-def: Create 30-day practice tracker
```

Now when an agent picks up `ts-abc`, it knows *why* — not just from the task description, but from the structured intent of the parent epic. It knows the constraints. It knows what trade-offs to make.

### 2.4 Intent Review Command

```
tpg review ep-abc
```

Output:
```
Epic: Becoming App Phase 2
Intent: Bridge the gap between study and transformation

Completed Tasks: 8/12
Intent Alignment Check:
  ✓ ts-001: Reflection prompt — serves intent (bridges study→practice)
  ? ts-003: Push notifications — may conflict with constraint "don't gamify"
  ✗ ts-005: Analytics dashboard — optimizes for engagement metrics (violates values)

Questions for Human Review:
  1. ts-003: Push notifications could encourage practice OR gamify it. Intent-aligned?
  2. ts-005: Built an engagement dashboard. Intent says "depth over engagement metrics." Keep or remove?
```

This is Abraham 4:18 — "watched those things which they had ordered, until they obeyed." But now watching against *intent*, not just spec completion.

### 2.5 Multi-Repo Workspace Support

**Problem:** Real work spans multiple repositories at different stages simultaneously. Right now TPG is single-repo (`.tpg/tpg.db` in project root). If you're building the becoming app in one repo, studying in this repo, and contributing to an open-source project in a third — you have three separate, disconnected TPG databases. There's no way to:

- See all your active work across repos in one place
- Host specs centrally that reference tasks in different repos
- Track cross-repo dependencies (e.g., "becoming app needs gospel-vec API changes")
- Switch context between repos without losing awareness of what's in-flight elsewhere

**Design Direction: Hub + Spoke**

```
~/.tpg/                          ← Hub (user-level)
  ├── hub.db                     ← Cross-repo awareness: projects, active intents, context switching
  ├── config.yaml                ← Global settings, registered projects
  └── specs/                     ← Central spec hosting (optional)
      ├── becoming-app/
      │   └── 06_becoming-app.md
      └── gospel-vec/
          └── 03_semantic-search.md

~/code/scripture-study/.tpg/     ← Spoke (repo-level, existing behavior)
  └── tpg.db                     ← Tasks, epics, learnings for this repo

~/code/becoming-app/.tpg/        ← Spoke
  └── tpg.db

~/code/gospel-vec/.tpg/          ← Spoke
  └── tpg.db
```

**Key commands:**

```bash
# Register a repo with the hub
tpg project register ~/code/becoming-app --name "Becoming App"
tpg project register ~/code/scripture-study --name "Scripture Study"

# See all active work across repos
tpg dashboard
# Output:
# Scripture Study    3 active tasks   Intent: deep study tools
# Becoming App       5 active tasks   Intent: bridge study → transformation
# Gospel Vec         1 blocked task   Intent: semantic search for scriptures

# Host specs centrally, link to repo-level epics
tpg spec host scripts/plans/06_becoming-app.md --project "Becoming App" --epic ep-xyz
tpg spec list --all-projects

# Cross-repo dependencies
tpg dep add ts-abc --cross-repo "Gospel Vec" ts-def
# "Becoming app task ts-abc depends on gospel-vec task ts-def"

# Context switching
tpg switch "Becoming App"          # Updates prime output for that project's context
tpg prime --global                  # Prime output includes cross-repo awareness
```

**Spec hosting options:**

| Approach | Pros | Cons |
|----------|------|------|
| **Central specs in hub** (`~/.tpg/specs/`) | One place for all specs; survives repo changes | Disconnected from repo git history |
| **Specs stay in-repo, hub indexes them** | Version-controlled with code; natural home | Scattered; need hub to find them |
| **Hybrid: hub for cross-cutting, repo for project-specific** | Best of both; mirrors how we actually think | More complex; needs clear conventions |

**Recommendation:** Hybrid. Project-specific specs live in their repo (`scripts/plans/`). Cross-cutting specs (like this intent development plan that touches multiple repos) live in the hub or in a designated "home" repo like this one. The hub maintains an index either way.

**Phase 2.5 tasks:**
1. Add `~/.tpg/` hub directory with project registry
2. `tpg project register/list/switch` commands
3. `tpg dashboard` — cross-repo active work summary
4. `tpg prime --global` — include cross-repo context
5. Cross-repo dependency tracking
6. Central spec hosting with indexing
7. `tpg spec list --all-projects` — find any spec from anywhere

---

### 3.1 The Spiritual → Physical → Review Cycle as Workflow

Encode the creation pattern as a TPG template:

```yaml
# .tpg/templates/intent-driven.yaml
name: intent-driven
description: "Intent → Spec → Build → Review cycle"
steps:
  - title: "Define intent for {{feature_name}}"
    description: |
      Before any spec or task, answer:
      1. WHY does this exist? What purpose does it serve?
      2. What does success look like that we can't easily measure?
      3. What are the non-negotiable constraints?
      4. What trade-offs do we accept?
      Write answers to the epic's intent fields.
    labels: [intent]
  
  - title: "Spiritual creation: spec {{feature_name}}"
    description: |
      Create the planning document — the spiritual creation.
      Reference the intent from step 1.
      Spec should be complete enough that "five words" could trigger implementation.
    labels: [spec]
    depends_on: [0]
  
  - title: "Build {{feature_name}}"
    description: |
      Physical creation against the spec.
      Honor the constraints from the intent.
      When facing trade-offs, consult the values hierarchy.
    labels: [build]
    depends_on: [1]
  
  - title: "Review {{feature_name}} against intent"
    description: |
      Not just "does it work?" but "does it serve the intent?"
      Check against success criteria, constraints, and values.
      Abraham 4:18 — "watched until they obeyed."
    labels: [review]
    depends_on: [2]
  
  - title: "Become: what changed?"
    description: |
      The becoming step. What did building this teach us?
      What capacity did we develop? (D&C 130:18)
      What would we do differently?
      Update the intent if we learned something about what we actually value.
    labels: [becoming]
    depends_on: [3]
```

### 3.2 Spec Offloading Workflow

The core request: offload spec document creation from the human's head into the tool.

**Current state:** Specs live as markdown files scattered across `scripts/plans/`, `docs/`, and sometimes inline in study documents. The human carries the organizational knowledge of which spec relates to what.

**Desired state:** The human says "I want to build X" and the tool:
1. Captures the intent ("why X?")
2. Creates a structured spec document in the right location
3. Registers it with TPG
4. Generates dependent tasks with intent metadata
5. Preserves the intent→spec→task chain across sessions

**Implementation sketch:**

```bash
# Human: "I want to add intent tracking to the becoming app"
tpg intent create "Intent tracking in becoming app"
# Prompts for: purpose, success criteria, constraints, values
# Creates: scripts/plans/08_intent-tracking.md (stub with frontmatter)
# Creates: ep-xyz with intent metadata populated
# Opens spec file for collaborative editing with AI

# After spec is filled out:
tpg plan ep-xyz
# AI reads the spec, proposes tasks, human approves
# Tasks created with intent metadata inherited from epic

# During work:
tpg start ts-abc
tpg prime  # Includes intent context
# Agent works with full awareness of why

# At completion:
tpg review ep-xyz
# Checks intent alignment, prompts human for judgment calls
```

### 3.3 Document Lifecycle Management

Integrate document lifecycle with TPG:

```bash
tpg doc register study/charity.md --type study --intent "understand charity in practice"
tpg doc stale                     # Documents not updated in 60+ days
tpg doc orphans                   # Documents not linked to any epic/task
tpg doc related study/charity.md  # Show connected documents
```

This solves the markdown flood gradually — not by reorganizing everything at once, but by making relationships explicit as we touch documents.

---

## Phase 4: Gospel-Grounded Intent Patterns

### 4.1 The Satan Test

From the Grand Council in Abraham 3 and Moses 4: The Father presented His plan and asked "Whom shall I send?" Christ volunteered — "Father, thy will be done" ([Moses 4:2](../gospel-library/eng/scriptures/pgp/moses/4.md)). Satan *rebelled* against the Father's plan, seeking to destroy agency and take God's honor ([Moses 4:1,3](../gospel-library/eng/scriptures/pgp/moses/4.md)). His rebellion optimized for a measurable outcome (everyone returns) while violating the core constraint (agency). Any intent engineering system should include a "Satan test":

> Does this plan achieve the stated goal while violating what we actually value?

In practice:
- "Ship faster" → Does it sacrifice code quality, which we value?
- "More study documents" → Does it sacrifice depth, which we value?
- "Higher engagement" → Does it gamify what should be genuine?
- "Resolve tickets fast" → Does it destroy customer relationships, which we value?

### 4.2 The D&C 121 Governance Model

The video's "delegation framework" concept maps directly to D&C 121:41-46:

| D&C 121 Principle | Agent Governance Equivalent |
|-------------------|----------------------------|
| "Only by persuasion" | Agents recommend, humans decide on value-laden choices |
| "Long-suffering" | Patience with iteration; don't shortcut the review cycle |
| "Gentleness and meekness" | Agents don't overwrite human context without permission |
| "Love unfeigned" | Genuine service to the user's intent, not just task completion |
| "Pure knowledge" | Verified sources, not hallucinated facts |
| "Without hypocrisy" | Transparent about uncertainty and limitations |
| "Reproving betimes with sharpness" | Flag misalignment immediately, clearly, with evidence |
| "Increase of love afterward" | After correction, continue with full commitment |
| "Without compulsory means it shall flow" | When intent is aligned, collaboration is natural |

### 4.3 The Book of Mormon "Scenario Building" Pattern

The request mentioned "spiritual → physical → review (scenario building)." There's a powerful Book of Mormon pattern for this:

**Alma 32:27-43** — The experiment upon the word:

1. **Intent:** "If ye will awake and arouse your faculties, even to an experiment upon my words" — declare your purpose
2. **Scenario (spiritual):** "Suppose ye have a seed" — envision the expected result
3. **Execute (physical):** "Plant this seed in your hearts" — do the work
4. **Review:** "It beginneth to swell... it beginneth to enlarge my soul... it beginneth to be delicious" — observe the fruit
5. **Iterate:** "Your knowledge is not perfect... your faith is dormant... nourish it" — the work continues

This is exactly intent → spec → build → review → become. Alma 32 is an intent engineering framework for personal growth. The "scenario" step — where you envision what the fruit *should* look like before you plant — is the spiritual creation. And the review against fruit quality (not just "did the tree grow?" but "is the fruit good?") is intent-aligned evaluation.

---

## Implementation Priorities

### Now (This Week)
1. ✅ Create [docs/work-with-ai/04_intent-engineering.md](work-with-ai/04_intent-engineering.md) — the discovery document
2. ✅ Create this development plan
3. Create `docs/INDEX.md` — initial document registry
4. Add frontmatter to key active documents (5-10 most referenced)

### Soon (Next 2 Weeks)
5. Draft TPG improvement proposals:
   - Intent metadata on epics (Phase 2.1)
   - Spec registration command (Phase 2.2)
   - Intent-aware prime output (Phase 2.3)
6. Design multi-repo hub + spoke architecture (Phase 2.5)
   - Project registry, dashboard, cross-repo dependencies
   - Spec hosting strategy (hybrid: in-repo + hub index)
7. Create the `intent-driven` TPG template (Phase 3.1)
8. Design spec offloading workflow in detail (Phase 3.2)

### Later (Next Month)
9. Implement TPG changes (fork or propose upstream)
10. Build multi-repo hub (`~/.tpg/`, project register/switch/dashboard)
11. Build document lifecycle tooling (Phase 3.3)
12. Integrate intent tracking into becoming app
13. Write the "Satan test" as a reusable evaluation template
14. Develop D&C 121 governance patterns for agent config

### Ongoing
13. Apply intent thinking to every new project/study
14. Refine the gospel→engineering mappings as we learn
15. Document discoveries in the work-with-ai series

---

## Open Questions

1. **TPG fork or upstream?** Do we propose these changes to the TPG maintainer, or fork for our specific use case? The intent layer is opinionated — it may not fit every TPG user.

2. **How structured should intent be?** Free-text intent statements are flexible but hard to evaluate programmatically. Structured fields (goal, constraints, values, success criteria) are more actionable but more ceremony. Where's the balance?

3. **Intent drift detection.** The video warns about "alignment drift over time." How do we detect when our work has drifted from our stated intent? Regular reviews? Automated checks? The Sabbath pattern — weekly reflection?

4. **Multi-repo workflow.** Real work spans several repositories at different stages simultaneously. Phase 2.5 proposes a hub + spoke model (`~/.tpg/` hub + per-repo `.tpg/`). Key questions:
   - How lightweight can the hub be? A single SQLite DB with project pointers, or something more?
   - Should `tpg prime` always include cross-repo awareness, or only with `--global`?
   - How do cross-repo dependencies interact with git branches / worktrees?
   - Do we need a top-level intent document that cascades into project-level intents?

5. **The human-in-the-loop question.** Where exactly do agents decide autonomously vs. need human input? The D&C 121 model says "all power by persuasion" — but that's aspirational. Practical decision boundaries are needed.

---

## References

### Video
- [Nate B Jones, "Prompt Engineering Is Dead. Context Engineering Is Dying. What Comes Next Changes Everything."](https://www.youtube.com/watch?v=QWzLPn164w0) — Feb 24, 2026, 29:40

### Scriptures
- [Moses 1:39](../gospel-library/eng/scriptures/pgp/moses/1.md) — God's intent statement
- [Moses 3:5](../gospel-library/eng/scriptures/pgp/moses/3.md) — Spiritual before temporal
- [Abraham 3:22-27](../gospel-library/eng/scriptures/pgp/abr/3.md) — Grand Council, the Father's plan, Christ volunteers, Satan rebels
- [Moses 4:1-3](../gospel-library/eng/scriptures/pgp/moses/4.md) — Satan's rebellion vs. Christ's submission to the Father's will
- [Abraham 4:18-31](../gospel-library/eng/scriptures/pgp/abr/4.md) — "Watched until they obeyed"
- [D&C 88:40](../gospel-library/eng/scriptures/dc-testament/dc/88.md) — Intelligence cleaveth unto intelligence
- [D&C 93:29-36](../gospel-library/eng/scriptures/dc-testament/dc/93.md) — Truth, intelligence, agency
- [D&C 121:34-46](../gospel-library/eng/scriptures/dc-testament/dc/121.md) — Authority, governance, persuasion
- [D&C 130:18-19](../gospel-library/eng/scriptures/dc-testament/dc/130.md) — Principles of intelligence rise with us
- [Alma 32:27-43](../gospel-library/eng/scriptures/bofm/alma/32.md) — Experiment upon the word — scenario building

### Internal Documents
- [docs/work-with-ai/04_intent-engineering.md](work-with-ai/04_intent-engineering.md) — The discovery document
- [docs/work-with-ai/01_planning-then-create-gospel.md](work-with-ai/01_planning-then-create-gospel.md) — Part 1: The Creation Pattern
- [docs/work-with-ai/02_watching-until-they-obey-gospel.md](work-with-ai/02_watching-until-they-obey-gospel.md) — Part 2: The Feedback Loop
- [docs/work-with-ai/03_intelligence-cleaveth-gospel.md](work-with-ai/03_intelligence-cleaveth-gospel.md) — Part 3: Intelligence Cleaveth
- [becoming/00_overview.md](../becoming/00_overview.md) — The becoming framework
- [external_context/tpg/README.md](../external_context/tpg/README.md) — TPG tool documentation

### External
- Anthropic, "Building Effective Agents" (Sept 2025) — Context engineering definition
- Deloitte, "2026 State of AI in the Enterprise" — 84% haven't redesigned jobs for AI
- Google DeepMind, "Five Levels of AI Agent Autonomy" — Operator through Observer hierarchy
