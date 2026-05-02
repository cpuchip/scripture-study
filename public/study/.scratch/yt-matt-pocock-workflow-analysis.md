# Scratch: Matt Pocock AI Workflow Analysis

## Videos Analyzed
1. **"Software Fundamentals Matter More Than Ever"** — Matt Pocock, AI Engineer conf keynote (2026-04-23, 18:26)
2. **"Full Walkthrough: Workflow for AI Coding"** — Matt Pocock, AI Engineer workshop (2026-04-24, 1:36:30)

The workshop is essentially the extended, hands-on version of the keynote thesis. Keynote sets up the argument; workshop demonstrates the full flow.

---

## Pocock's Core Framework

### Thesis
AI coding doesn't replace software engineering fundamentals — it amplifies them. Bad code bases produce bad agent output. Good code bases matter more than ever.

### The "Smart Zone" / "Dumb Zone"
- From Dex Hardy (Human Layer): LLMs have a smart zone (fresh context) and dumb zone (degraded attention at ~100K tokens regardless of total window size)
- Task sizing must stay within the smart zone
- Prefers **clearing context** (Memento-style reset) over compacting because compacting produces always-the-same summarized state
- This is his main reason for preferring small discrete tasks over long-running sessions

### Anti "Specs-to-Code"
- The specs-to-code movement says: write spec → AI turns it into code → if broken, fix spec not code
- Pocock tried it, got worse code each iteration
- Code is NOT cheap. Bad code is the most expensive it's ever been.
- You must keep a handle on the code. Code is your battleground.

### The Workflow (Grill → PRD → Issues → Implement → QA)

1. **Grill Me Skill** — "Interview me relentlessly about every aspect of this plan until we reach a shared understanding"
   - From Frederick P. Brooks' "design concept" — the ephemeral shared idea between collaborators
   - Asks 40-100 questions with recommendations
   - Goal: shared wavelength, not a plan asset
   - Better than Claude's plan mode because plan mode is "eager to create an asset"

2. **Ubiquitous Language Skill** — From DDD
   - Scans codebase, creates markdown file of shared terminology
   - Reduces verbosity, improves alignment
   - Reference during grilling and planning

3. **Write PRD Skill** — Destination document
   - Problem statement, solution, user stories, implementation decisions, testing decisions
   - Proposed modules to modify (keeps code in mind throughout)
   - He does NOT review the PRD — LLMs are good at summarization, and he already has shared design concept
   - Does NOT keep PRDs in repo (doc rot — code diverges, old PRDs mislead agents)

4. **PRD → Issues** — Journey document
   - Uses Kanban board with blocking relationships
   - **Vertical slices** (traceable bullets) not horizontal layers
   - AI naturally codes horizontally (schema → API → frontend); this delays feedback
   - Vertical slice: schema + service + minimal frontend in one go
   - Issues are independently grabbable → enables parallelization

5. **AFK Implementation (Ralph Loop / Sandcastle)**
   - Human leaves the loop
   - Sequential: pick next task, explore repo, TDD, feedback loops (types, tests), commit
   - Parallel (Sandcastle): planner picks batch → multiple sandboxed implementers → reviewer → merger
   - Uses weaker model (Sonnet) for implementation, stronger (Opus) for review

6. **QA / Code Review**
   - Human taste is imposed here
   - QA creates more issues for the Kanban board
   - Automated review step should clear context first (review in smart zone, not dumb zone)

### Software Engineering Foundations He Cites
- **John Ousterhout** — *A Philosophy of Software Design*: complexity definition, deep vs shallow modules
- **The Pragmatic Programmer** — software entropy, "no one knows exactly what they want", outrunning your headlights
- **Frederick P. Brooks** — *The Design of Design*: design concept, design tree
- **Kent Beck** — "Invest in the design of the system every day"
- **Martin Fowler** — refactoring, small steps

### TDD as AI Enabler
- Red-green-refactor forces small steps
- Without feedback loops, AI is "coding blind"
- Quality of feedback loops = ceiling of AI output quality
- AI tends to write bad tests (cheats by doing implementation first)
- TDD makes cheating harder because test exists first

### Deep Modules
- Simple interface hiding complex implementation
- AI is good at creating shallow modules (many tiny files)
- Shallow modules are hard for AI to navigate
- Deep modules are easier to test (test at the boundary)
- Design the interface, delegate the implementation
- This preserves developer sanity and codebase understanding

### Push vs Pull for Standards
- **Push**: standards in system prompt (Claude.md) — always sent
- **Pull**: skills that agent can invoke when needed
- Implementer should PULL standards (ask when needed)
- Reviewer should PUSH standards (compare code against explicit standards)

---

## Comparison with Our Guide

### Alignments (Strong)

| Pocock | Our Guide |
|--------|-----------|
| Rejects specs-to-code / vibe coding | Rejects "just prompting" — four disciplines required |
| Code is not cheap; bad code is expensive | "Bad code bases make bad agents" — same thesis |
| Grill Me = reach shared understanding before building | Intent engineering, covenant, council moment — alignment before action |
| Ubiquitous language = shared terminology | Context engineering, ubiquitous language in DDD section |
| Feedback loops are the speed limit | "Rate of feedback is your speed limit" — identical |
| Small tasks in smart zone | Decomposition primitive: "completable and verifiable in under 2 hours" |
| Human-in-the-loop for planning, AFK for implementation | Day shift (human) / night shift (AI) — matches our steward paradigm |
| TDD forces small deliberate steps | Spec engineering with evaluation design |
| Deep modules = testable boundaries | Interface-first design, modular architecture |
| Design the interface, delegate implementation | Stewardship: "this domain is yours — grow it, guard it" |
| Clear context over compacted context | "Files are durable, context is not" — we prefer externalized memory |
| Progressive trust (Sandcastle: Sonnet implements, Opus reviews) | Stewardship levels: faithfulness earns more authority |

### Divergences (Meaningful)

| Dimension | Pocock | Our Guide |
|-----------|--------|-----------|
| **Theological frame** | None — purely secular engineering | Gospel creation cycle, covenant, stewardship, consecration |
| **PRD lifecycle** | Delete after use (doc rot) | Keep in `.spec/` directory as living document |
| **Context strategy** | Clear and reset (Memento) | Progressive disclosure (line upon line) with memory system |
| **Planning output** | PRD + Kanban issues | Intent → Covenant → Stewardship → Spec → Tasks |
| **Agent relationship** | Tool/contractor | Covenant partner with mutual obligations |
| **Failure recovery** | Retry, fix, continue | Atonement pattern — redemptive recovery with learning capture |
| **Rest pattern** | None — continuous optimization | Sabbath — structured ending, seeing, declaring |
| **Organizational scale** | Parallel agents (Sandcastle) | Fractal hierarchy (ward → stake → area) with keys and stewardship |
| **Memory** | GitHub issues (closed = stale) | `.mind/` memory architecture — identity, preferences, principles, episodes |
| **Session model** | Discrete sessions, clear context | Iterative turns with persistent goal context (steward paradigm) |

### Gaps in Our Thinking (What Pocock Strengthens)

1. **Doc rot / spec freshness**
   - We keep PRDs in `.spec/` but don't have a freshness strategy
   - Pocock's observation: old PRDs actively mislead agents when code has diverged
   - Suggestion: Add "spec freshness" protocol — version specs with code, archive stale specs, or use LLM to verify spec-code alignment before invoking

2. **Deep modules as AI enabler**
   - Our guide mentions modular architecture but doesn't emphasize John Ousterhout's deep/shallow module distinction
   - This is a huge deal: AI creates shallow modules by default, which degrades its own future performance
   - Suggestion: Add "Deep Modules" section to context engineering or spec engineering

3. **TDD as forcing function**
   - We mention tests in acceptance criteria but don't teach TDD specifically
   - Pocock: TDD is "absolutely essential" because it forces the AI into small steps and prevents test-cheating
   - Suggestion: Add TDD/red-green-refactor to the complete cycle as a step 6b or feedback loop pattern

4. **Vertical slices vs horizontal layers**
   - Our decomposition primitive doesn't explicitly warn against horizontal decomposition
   - AI naturally codes horizontally (all schema, then all API, then all frontend)
   - Vertical slices (end-to-end thin functionality) give earlier feedback
   - Suggestion: Add "traceable bullets / vertical slices" to spec engineering decomposition section

5. **Token monitoring / smart zone awareness**
   - We discuss context degradation but don't teach the practitioner how to monitor it
   - Pocock has a token counter in his status line and uses it actively
   - Suggestion: Add "context hygiene" section — token monitoring, when to clear vs compact, smart zone markers

6. **Kanban with blocking relationships for parallelization**
   - Our task decomposition is somewhat sequential
   - Pocock's Kanban with explicit blocking enables parallel agents
   - Suggestion: Add dependency-graph task format to `.spec/tasks/` convention

7. **Push vs pull for standards**
   - We don't articulate when to push standards (always in context) vs pull (skills on demand)
   - Pocock: implementer pulls, reviewer pushes
   - Suggestion: Document this pattern in context engineering or agent setup

8. **Codebase architecture skill**
   - Pocock has an "improve codebase architecture" skill that scans and suggests deep module opportunities
   - We don't have an equivalent skill for restructuring code for AI friendliness
   - Suggestion: Create an `.github/skills/improve-architecture.md` skill

### Gaps in Pocock's Thinking (What We Could Strengthen)

1. **No covenant / mutual commitment framework**
   - Pocock's workflow is unilateral: human plans, AI executes
   - No concept that the human has obligations (timely review, accurate context, not shortcutting)
   - Our covenant pattern (D&C 82:10) makes this explicit and mutual

2. **No progressive trust / stewardship levels**
   - Pocock mentions "faithful over a few things" in passing but doesn't structure agent autonomy dynamically
   - His Sandcastle uses static role assignments (Sonnet implements, Opus reviews)
   - Our stewardship levels (task → feature → domain → architecture) with earned trust is richer

3. **No intent engineering at organizational level**
   - His "design concept" is project-level only
   - No framework for cascading intent through teams or organizations
   - Our Part 6 (enterprise architecture) with fractal hierarchy fills this

4. **No Sabbath / rest pattern**
   - Pure productivity optimization with no structured reflection
   - Our Sabbath agent (ending, seeing, declaring) provides this
   - Also: no "atonement" pattern for redemptive failure recovery

5. **No memory system beyond issues**
   - Deletes PRDs, closes issues — organizational memory is lost
   - Our `.mind/` architecture preserves identity, preferences, principles, episodes across sessions
   - This is critical for multi-session relationship building with agents

6. **No session journal / learning capture**
   - Pocock mentions "tuning" the prompt but doesn't systematize learning capture
   - Our `session-journal` tool and `.spec/scratch/reflect.md` pattern ensures learning persists

7. **Over-reliance on clearing context**
   - "Memento" approach means every session starts from zero
   - This loses relationship continuity and forces re-alignment each time
   - Our progressive disclosure with memory system preserves earned context

8. **No council / multi-perspective deliberation**
   - Pocock's "grill me" is adversarial (AI questions human)
   - Our council moment is mutual — "took counsel among themselves" (Abraham 4:26)
   - Multiple agents with different perspectives could strengthen his planning phase

---

## Synthesis: What We Should Improve

### Immediate (This Week)

1. **Add "Context Hygiene" section to Part 2 (Context Engineering)**
   - Token monitoring, smart zone markers, when to clear vs compact vs externalize
   - Cite Pocock's ~100K heuristic

2. **Add "Vertical Slices" to Part 4 (Spec Engineering) decomposition primitive**
   - Warn against horizontal layer-by-layer decomposition
   - Teach traceable bullets: end-to-end thin functionality that crosses all layers

3. **Add "Deep Modules" section to Part 2 or Part 4**
   - John Ousterhout's deep vs shallow modules
   - How AI creates shallow modules by default
   - Design interface, delegate implementation

4. **Create `.github/skills/improve-architecture.md`**
   - Scan codebase for shallow modules
   - Suggest clusters that could become deep modules
   - Include test boundary recommendations

### Short Term (This Month)

5. **Add TDD/red-green-refactor to Part 5 (Complete Cycle)**
   - As a feedback loop pattern within Physical Creation
   - Emphasize it forces small steps and prevents test-cheating

6. **Add "Spec Freshness" protocol to `.spec/` conventions**
   - How to version specs with code
   - When to archive vs update
   - Automated spec-code alignment check

7. **Document push-vs-pull pattern for standards**
   - Implementer pulls (skills on demand)
   - Reviewer pushes (standards in context)
   - When to use which

8. **Add dependency-graph task format to `.spec/tasks/` convention**
   - Blocking relationships, parallelization potential
   - AFK vs human-in-the-loop classification per task

### Deeper (This Quarter)

9. **Strengthen the case for keeping specs vs Pocock's deletion argument**
   - Our `.spec/` directory serves a different purpose than his PRDs
   - Our specs include intent, covenant, values — not just implementation details
   - But we need a freshness strategy to counter doc rot
   - Maybe: specs are living documents, updated as code changes, not one-time throwaways

10. **Evaluate whether our progressive disclosure contradicts Pocock's clear-and-reset**
    - He says compacting produces "always the same" degraded state
    - Our memory system externalizes context to files, avoiding compaction
    - This may be the best of both worlds: persistent relationship without context bloat
    - Document this explicitly as our answer to the Memento problem
