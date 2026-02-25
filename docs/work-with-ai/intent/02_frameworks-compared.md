# Frameworks Compared

**Part of:** [Intent-Driven Development Research](00_index.md)
**Date:** February 2026
**Status:** Active

---

## The Question

We need something that works for **one person studying scriptures on their own** and for **a team building software across multiple repos, branches, and tickets.** That's a wide aperture. Let's see what exists and what's missing.

---

## Head-to-Head Comparison

| Capability | OpenSpec | TPG | Kiro (AWS) | BMAD | GitHub Spec Kit | What We Need |
|-----------|---------|-----|-----------|------|----------------|-------------|
| **Specs as files in repo** | ✅ Markdown | ❌ SQLite DB | ❌ IDE-internal | ✅ Process-based | ❓ Likely markdown | ✅ Required |
| **Single source of truth** | ✅ One spec doc | ❌ Tasks are disconnected | ❌ Per-feature specs | ❌ Methodology only | ❓ Unknown | ✅ One living doc per project/feature |
| **Delta-based changes** | ✅ ADDED/MODIFIED/REMOVED | ❌ Task status only | ❓ Unknown | ❌ No | ❓ Unknown | ✅ Track what changed and why |
| **Intent layer** | ❌ Specs = what, not why | ❌ No intent metadata | ❌ No | ❌ No | ❓ Unknown | ✅ Must encode purpose, values, constraints |
| **Task management** | Partial (tasks.md generated) | ✅ Full task/epic/dependency | ❌ No | ❌ Process only | ❓ Unknown | ✅ Tasks linked to specs and intent |
| **Multi-repo** | ❌ Single project | ❌ Single project | ❌ Single IDE | ❌ No | ❌ Likely per-repo | ✅ Hub+spoke across repos |
| **Cross-session memory** | ❌ No | ✅ Learnings system | ❌ Unknown | ❌ No | ❌ Unknown | ✅ Carry knowledge forward |
| **Branch-aware** | ❓ Lives in repo (implicit) | ❌ SQLite doesn't branch | ❌ IDE-locked | N/A | ❓ Git-native? | ✅ Specs travel with branches |
| **Team-readable** | ✅ Markdown | ❌ DB requires CLI | ❌ IDE-locked | ✅ Process docs | ❓ GitHub-native | ✅ Any team member can read |
| **Review against intent** | ❌ No | ❌ No | ❌ Tests only | ❌ No | ❌ No | ✅ Did we build what we intended? |
| **Brownfield support** | ✅ Can describe existing systems | ✅ Can track existing work | ❓ Unknown | ✅ Process applies anywhere | ❓ Unknown | ✅ Never starting from scratch |
| **AI agent integration** | ✅ Agents read spec | ✅ `tpg prime` context | ✅ Built for agents | ❌ Human-focused | ❓ GitHub Copilot | ✅ Agents consume and honor specs |
| **Custom workflows** | ✅ Schema system | ❌ Fixed workflow | ❌ Fixed workflow | ✅ Flexible process | ❓ Unknown | ✅ Adapt to project type |
| **Open source** | ✅ GitHub | ✅ GitHub | ❌ Proprietary | ✅ Methodology | ❓ Likely proprietary | ✅ Preferred |

---

## What Each Tool Gets Right

### OpenSpec Gets Right: File-Based, Single Source of Truth
OpenSpec's core insight is exactly ours: **one unified spec document as the authoritative reference.** Not scattered requirement files. Not database records. A living markdown document that evolves with the system. The delta spec workflow (Propose → Apply → Archive) creates clean audit trails without cluttering the main spec.

The custom schema system is smart — recognizing that a greenfield prototype and a mature enterprise system need different artifact pipelines.

### TPG Gets Right: Persistent State and Context Injection
TPG's `tpg prime` is the best implementation of "here's what the agent needs to know right now." Cross-session learnings mean the agent doesn't start ignorant every time. Dependency tracking between tasks prevents execution-order chaos.

### Kiro Gets Right: Making Specs Mandatory
The most radical idea in the landscape: **refuse to generate code until the spec is testable.** Every other tool makes specs optional-but-recommended. Kiro makes them a gate. This is the only tool that treats the spec bottleneck as a *design constraint* rather than a *best practice*.

### AIDD Gets Right: The Organizational Layer
AIDD is the only framework thinking about how agent-assisted development changes *organizations*, not just individual workflows. Fluid teams, 12-factor agent principles, A2A+MCP protocols — this is the enterprise architecture that makes spec-driven development work at scale.

### The 12-Factor Agent Principles (from AIDD)
Particularly relevant to us:

| Principle | Meaning | Our Implication |
|-----------|---------|----------------|
| **Stateless processes** | Agents don't carry internal state; state is externalized | Spec files in repo, not SQLite DB |
| **Declarative config** | Agent behavior defined by configuration, not code | `.github/copilot-instructions.md`, agent definitions |
| **Externalized state** | State lives outside the agent, accessible to anyone | Markdown files, not tool-internal databases |
| **Disposability** | Any agent instance can be replaced without losing context | All knowledge in files, not in session memory |
| **Dev/prod parity** | Same specs in development and production | Specs travel with branches |

---

## The Gap Analysis

### What No Tool Does

**1. Intent as a first-class concept**
Every tool treats specs as "what to build." None treat intent as "why we're building it, what trade-offs we accept, what constraints are non-negotiable." Intent lives in people's heads or, at best, in a separate document nobody maintains.

**2. Multi-repo coordination**
Every tool assumes a single project. Real work spans multiple repos at different stages:
- Building a feature in repo A that depends on an API change in repo B
- Running three repos simultaneously at different maturity levels
- Shared intent across repos (e.g., "all our tools serve deep scripture study")

**3. Spec → Task → Review traceability**
OpenSpec generates tasks.md but doesn't track execution. TPG tracks tasks but doesn't connect them to specs. Nobody closes the loop from intent → spec → task → review → "did this serve the intent?"

**4. Branch-aware state**
Specs in the repo travel with branches (OpenSpec gets this implicitly). Task state in a database doesn't (TPG's problem). Nobody has integrated both: specs AND tasks AND intent all traveling with the branch.

**5. Review against intent (not just completion)**
Every tool tracks "is the task done?" Nobody tracks "did the completed task serve the stated purpose?" The Klarna failure pattern: technically complete, strategically wrong.

---

## Solo vs. Team Needs

| Need | Solo Dev | Team Dev | Same Pattern? |
|------|---------|---------|---------------|
| Intent statement | "Why am I building this?" — personal clarity | "Why are we building this?" — shared alignment | ✅ Same, different audience |
| Spec document | Living markdown in repo | Living markdown in repo, branch-aware | ✅ Same artifact, branch adds complexity |
| Task tracking | Lightweight, maybe just a checklist in the spec | Heavier — assignments, dependencies, status | ⚠️ Same concept, different weight |
| Review process | Self-review against intent | Peer review + intent audit | ⚠️ Same goal, more people |
| Multi-repo | Probably not needed | Essential | ❌ Different need |
| Context injection | `tpg prime` / copilot-instructions | Shared prime context across team | ⚠️ Same mechanism, shared access |
| Decision boundaries | "What do I let the agent decide?" — personal trust calibration | "What can the team's agents decide?" — organizational trust policy | ✅ Same question, more stakeholders |
| Learning persistence | Cross-session memory for one person | Shared knowledge base for the team | ⚠️ Same idea, shared ownership |

**Key insight: it's the same pattern.** Solo dev is the degenerate case of team dev where the team size is 1. The spec-driven workflow doesn't change — the coordination layer does. A tool that works for teams works for solo; a tool that only works for solo won't scale.

---

## What We Actually Need: The Composite

Taking the best of each, plus what's missing:

| Layer | Source | What It Gives Us |
|-------|--------|-----------------|
| **File-based specs** | OpenSpec | Single source of truth, delta changes, lives in repo |
| **Intent metadata** | Our invention | Purpose, values, constraints, decision boundaries ON the spec |
| **Task management** | TPG (rethought) | Tasks linked to specs, inheriting intent, but stored as files |
| **Context injection** | TPG's `tpg prime` | Agent gets intent + spec + task context at session start |
| **Cross-session learning** | TPG's learnings | Knowledge persists, but in markdown (not SQLite) |
| **Multi-repo awareness** | Our invention | Hub+spoke: per-repo specs + central index |
| **Mandatory spec gate** | Kiro's insight | Don't execute until the spec is reviewed |
| **Organizational model** | AIDD framework | Fluid teams, 12-factor agents, externalized state |
| **Review against intent** | Our invention | Not "is it done?" but "does it serve the purpose?" |
| **Gospel governance** | [D&C 121 model](../04_intent-engineering-gospel.md) | Persuasion over compulsion, stewardship over control |

---

## The File-Based State Question

This is the architectural decision that matters most. Two philosophies:

### Database State (TPG's Current Approach)
```
.tpg/tpg.db  ← SQLite database
  Tasks, epics, dependencies, learnings all in tables
  CLI required to read/write
  Doesn't branch with code
```

**Pros:** Fast queries, relational integrity, rich filtering
**Cons:** Not version-controlled, not human-readable, not portable, doesn't travel with branches

### File-Based State (What We're Gravitating Toward)
```
.spec/
  intent.md           ← Project-level intent (purpose, values, constraints)
  spec.md             ← Source of truth specification (OpenSpec-style)
  deltas/
    2026-02-25-notifications.md  ← Delta spec (proposed change)
  tasks/
    ts-001-reflection-prompt.md  ← Task file (status, intent, spec reference)
    ts-002-practice-tracker.md
  learnings/
    ls-001-caching-strategy.md   ← Reusable knowledge
  archive/
    delta-2026-02-20-auth.md     ← Applied and archived deltas
```

**Pros:**
- Version-controlled with code
- Travels with branches
- Human-readable (any team member can read status without CLI)
- AI-readable (agents can `read_file` directly)
- Diffable (changes show up in PR reviews)
- Portable (works with any tool, any IDE, any agent)
- Grep-able (find all tasks, all specs, all learnings with file search)

**Cons:**
- No relational integrity (must maintain links manually or via convention)
- Slower for complex queries (no SQL)
- File proliferation risk (many small files)

**Our take:** The cons are manageable (conventions + tooling can enforce consistency). The pros are fundamental. **File-based state aligns with the 12-factor agent principle of externalized state** and with our core requirement that state must travel with the code.

A thin CLI tool (like OpenSpec) can provide convenience commands for common operations (`new-delta`, `apply-delta`, `list-tasks`, `prime`) while the underlying data remains plain markdown files.

---

## The Ideal Tool Architecture

```
intent-spec/                     ← CLI tool (thin layer over markdown files)
  ├── init                       ← Create .spec/ directory structure
  ├── intent set                 ← Write/update intent.md
  ├── delta propose              ← Create delta spec from description
  ├── delta apply                ← Merge delta into spec.md, archive delta
  ├── task create                ← Create task file linked to spec/delta
  ├── task list                  ← List tasks with status, inherited intent
  ├── prime                      ← Output context for agent: intent + spec + active tasks
  ├── review                     ← Check completed tasks against intent
  ├── learn                      ← Capture learning as markdown file
  └── hub                        ← Multi-repo commands
      ├── register               ← Add this repo to central index
      ├── dashboard              ← Cross-repo status
      └── switch                 ← Set active project context
```

**All state in `.spec/` as markdown files.** The CLI is a convenience layer, not a requirement. Any agent can read `.spec/intent.md` directly. Any team member can open `.spec/tasks/` in their editor. Any PR reviewer can see spec changes in the diff.

This is what OpenSpec almost is — but with intent as a first-class concept, tasks as files (not just a generated document), and multi-repo awareness.

---

## Recommendation

Don't build yet. The landscape is moving fast (OpenSpec 1.0 shipped January 26, 2026). Instead:

1. **Adopt OpenSpec's patterns** — Start using single-source-of-truth specs and delta workflows *now*, manually, in this repo. No tool required.
2. **Add intent preambles** — Put intent blocks on every new spec. This costs nothing and teaches the muscle memory.
3. **Prototype file-based tasks** — Try using a `.spec/tasks/` directory alongside OpenSpec for task tracking. See if it works before building a tool.
4. **Watch the landscape** — OpenSpec custom schemas, GitHub Spec Kit, and Kiro are all evolving rapidly. The right tool may emerge.
5. **Design the intent layer** — This is our real contribution. Nobody else is building it. The gospel patterns in [03_beyond-intent.md](03_beyond-intent.md) inform the design.

---

## Next

→ [03_beyond-intent.md](03_beyond-intent.md) — Gospel patterns the industry hasn't discovered
