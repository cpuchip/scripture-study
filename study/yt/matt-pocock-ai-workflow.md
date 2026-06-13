# Matt Pocock's AI Workflow — Alignment, Divergence, and Gaps

**Videos:**
- ["Software Fundamentals Matter More Than Ever"](https://www.youtube.com/watch?v=v4F1gFy-hqg) — AI Engineer keynote, 2026-04-23 (18:26)
- ["Full Walkthrough: Workflow for AI Coding"](https://www.youtube.com/watch?v=-QFHIoCo-Ko) — AI Engineer workshop, 2026-04-24 (1:36:30)

**Compared against:** [Working with AI Guide](../../docs/work-with-ai/guide/)

---

## What Pocock Is Teaching

Matt Pocock is a TypeScript teacher who has spent the last six months building a systematic workflow for AI-assisted coding. His thesis is straightforward: AI doesn't replace software engineering fundamentals — it amplifies them. The people who will thrive are the ones who know what good code looks like and can make the AI produce it.

He rejects the "specs-to-code" movement outright. He tried it. The code got worse each iteration. "Code is not cheap. In fact, bad code is the most expensive it's ever been." This is the same conviction that undergirds our guide's emphasis on specification engineering and stewardship — but Pocock arrives at it through John Ousterhout and The Pragmatic Programmer rather than Abraham 4.

His workflow has six stages:

1. **Grill Me** — An adversarial questioning skill that probes every aspect of a plan until human and AI reach a "shared design concept" (his term, from Frederick P. Brooks). He considers this better than Claude's plan mode because plan mode is "eager to create an asset" rather than eager to understand.

2. **Ubiquitous Language** — A DDD-inspired skill that scans the codebase and produces a terminology document. Shared vocabulary reduces verbosity and misalignment.

3. **Write PRD** — Summarize the grilled understanding into a destination document: problem, solution, user stories, implementation decisions, testing decisions. He does not read the PRD — LLMs are good at summarization, and he already has the shared concept.

4. **PRD → Issues** — Break the PRD into a Kanban board of independently grabbable tasks with blocking relationships. Crucially, he insists on **vertical slices** (end-to-end thin functionality) rather than horizontal layers (all schema, then all API, then all frontend). AI naturally codes horizontally; horizontal coding delays feedback until the final layer.

5. **AFK Implementation** — The human steps back. A "Ralph loop" (sequential) or Sandcastle (parallel) picks tasks, explores the repo, uses TDD, runs feedback loops, and commits. He uses Sonnet for implementation and Opus for review — weaker model does the work, stronger model checks it.

6. **QA and Code Review** — Human taste is imposed here. QA creates more issues for the board. The Kanban board accumulates rather than terminates.

Two concepts from the workshop deserve special attention:

**The Smart Zone.** From Dex Horthy (HumanLayer; the video's captions garble it as "Dex Hardy"): LLMs have a smart zone (~100K tokens, regardless of total context window) and a dumb zone beyond it. Pocock sizes every task to fit in the smart zone. He monitors token count obsessively. He prefers clearing context entirely — the "Memento" approach — over compacting, because compaction produces a degraded, always-the-same summarized state.

**Deep Modules.** From John Ousterhout: modules with simple interfaces hiding complex implementations. AI naturally produces shallow modules (many tiny files with complex interdependencies), which makes the codebase harder for the AI itself to navigate. Deep modules are easier to test, easier to reason about, and allow the human to "design the interface, delegate the implementation" — preserving sanity while moving fast.

---

## Where We're in Alignment

The overlaps are extensive and often word-for-word:

**Rejection of specs-to-code.** Pocock's "code is your battleground" mirrors our insistence that specification engineering is not about ignoring the code but about understanding it deeply enough to direct it well. Both frameworks treat the code as a first-class citizen, not a compiler output.

**Shared understanding before building.** Pocock's "grill me" skill is functionally equivalent to our council moment and covenant pattern — a structured process for reaching alignment before action. The difference is mostly theological framing: his is adversarial (AI questions human), ours is mutual ("took counsel among themselves").

**Context engineering as the real work.** Pocock's ubiquitous language skill, his emphasis on token counting, and his system prompt minimalism all align with our Part 2 thesis that "the prompt is 0.02% of what the model sees." Both frameworks treat context design as senior work.

**Feedback loops as the speed limit.** Pocock: "The rate of feedback is your speed limit." Our guide: "Quality of feedback loops = ceiling of AI output quality." Same insight, same source (The Pragmatic Programmer's "outrunning your headlights").

**Human-in-the-loop vs AFK.** Pocock's "day shift / night shift" metaphor — human plans, AI implements overnight — maps directly onto our developer-to-steward paradigm and the distinction between synchronous and autonomous work.

**Small, verifiable tasks.** Pocock's smart zone sizing and vertical slices align with our decomposition primitive ("completable and verifiable in under 2 hours") and our emphasis on traceable bullets.

**Progressive trust.** Pocock's use of Sonnet for implementation and Opus for review is a rudimentary stewardship assignment based on capability. Our stewardship levels (task → feature → domain → architecture) generalize this into a dynamic, earned-trust system.

---

## Where We Diverge

The divergences are not contradictions — they are different elevations of the same mountain.

| Pocock | Our Guide | Assessment |
|--------|-----------|------------|
| **No theological frame** | Gospel creation cycle, covenant, stewardship, consecration | We have an explicit "why" that anchors every discipline. Pocock's is implicit. Neither is wrong, but ours provides resilience when the workflow breaks. |
| **Delete PRDs to avoid doc rot** | Keep specs in `.spec/` as living documents | Pocock's fear is real: old PRDs mislead agents. But our specs encode intent and covenant, not just implementation — they're closer to constitution than documentation. We need a freshness strategy, though. |
| **Clear and reset (Memento)** | Progressive disclosure with memory system | Pocock distrusts compaction because it degrades. We solve this by externalizing memory to files (`.mind/`, `.spec/`) rather than compacting context. This may be the best of both worlds. |
| **Adversarial alignment (grill me)** | Mutual covenant + council | Pocock's AI questions the human aggressively. Our pattern is bilateral: both sides have obligations. The covenant pattern (D&C 82:10) makes the human's duties explicit. |
| **Pure productivity; no rest** | Sabbath agent, atonement pattern | Pocock's workflow is optimized for throughput. Ours includes ending, seeing, declaring — and redemptive recovery from failure. These aren't inefficiencies; they're where learning happens. |
| **Parallel agents (Sandcastle)** | Fractal enterprise hierarchy | Pocock's parallelization is mechanical: multiple agents in sandboxes. Our Part 6 models organizational scale through stewardship, keys, and intent cascade — closer to how the Church actually works. |

The deepest divergence is probably around **memory and relationship continuity**. Pocock deletes PRDs and clears context. Every session starts fresh. This is clean and avoids sediment buildup. But it also means every session pays the alignment tax again. Our memory system — identity, preferences, principles, episodes — is designed to make each session start with relationship continuity, not from zero. The tradeoff is complexity: we have to manage freshness and prevent stale memory from misdirecting. Pocock's approach is simpler. Ours is richer. Both have costs.

---

## Gaps in Our Thinking

Pocock's workshop exposed several gaps in our guide that we should close.

### 1. Doc Rot / Spec Freshness

We keep specs in `.spec/` but don't have a protocol for what happens when code diverges from spec. Pocock deletes PRDs precisely because he's seen agents follow stale specifications into wrong implementations. Our specs are higher-level (intent, covenant, values) and less prone to rot than implementation PRDs, but the risk exists.

**What to do:** Add a "spec freshness" protocol. Version specs with code. Archive specs when their implementation is complete and stable. Before invoking an agent on a domain, have it verify that the spec still matches the code. Consider specs as constitution (long-lived) vs PRDs as legislation (disposable) — and make the distinction explicit.

### 2. Deep Modules

Our guide mentions modular architecture but doesn't teach John Ousterhout's deep/shallow module distinction. This matters because AI naturally creates shallow modules — lots of tiny files with complex interdependencies — which degrades its own ability to navigate the codebase. Deep modules (simple interface, complex implementation) are easier for both humans and AI to reason about.

**What to do:** Add a "Deep Modules" section to Part 2 or Part 4. Teach interface-first design. Create an `improve-architecture` skill that scans for shallow modules and suggests deep module boundaries.

### 3. Vertical Slices

Our decomposition primitive doesn't explicitly warn against horizontal decomposition. AI naturally codes layer by layer: all the schema changes, then all the API endpoints, then all the frontend. This delays integrated feedback until the final layer. Vertical slices — thin end-to-end functionality that touches all layers — give immediate feedback.

**What to do:** Add "traceable bullets / vertical slices" to Part 4's decomposition section. Make it explicit: tasks should deliver visible, testable functionality across layers, not complete a single layer.

### 4. TDD as Forcing Function

We mention tests in acceptance criteria but don't teach TDD specifically. Pocock calls TDD "absolutely essential" because it forces the AI into small steps and prevents the AI from "cheating" by writing implementation before tests. Without TDD, the AI produces huge amounts of untested code and then tries to retrofit tests — badly.

**What to do:** Add TDD/red-green-refactor to Part 5 as a feedback loop pattern. Emphasize that the test must exist before the implementation, not after.

### 5. Token Monitoring / Smart Zone Awareness

We discuss context degradation but don't teach the practitioner how to monitor it. Pocock has a token counter in his status line and uses it as a primary navigation tool. He knows exactly when he's approaching the dumb zone and acts before degradation sets in.

**What to do:** Add a "context hygiene" section to Part 2. Teach token monitoring. Establish smart zone markers (~100K tokens as a heuristic). Document when to clear, when to compact, when to externalize to files.

### 6. Kanban with Blocking Relationships

Our task decomposition is somewhat sequential. Pocock's Kanban board with explicit blocking relationships enables parallel agents — multiple agents working on independent tasks simultaneously. This is a significant throughput multiplier.

**What to do:** Add dependency-graph task format to `.spec/tasks/` convention. Include AFK vs human-in-the-loop classification per task. Enable parallelization where blocking relationships allow it.

### 7. Push vs Pull for Standards

We don't articulate when to push standards (always in system context) versus pull (skills on demand). Pocock has a clean pattern: the implementer pulls standards when needed, the reviewer pushes standards for comparison. This reduces context bloat during implementation while ensuring rigorous review.

**What to do:** Document the push/pull pattern in Part 2. Implementer pulls skills. Reviewer pushes explicit standards. Different phases, different context strategies.

---

## Gaps in Pocock's Thinking

The reverse is also true. There are places where our framework has structures that would strengthen his workflow.

### 1. No Covenant Framework

Pocock's workflow is unilateral: human plans, AI executes. There's no concept of mutual obligation. He doesn't ask what the human owes the AI — timely review, accurate context, not shortcutting the process. Our covenant pattern (D&C 82:10) makes this explicit: "I, the Lord, am bound when ye do what I say." Even God operates through mutual commitment rather than unilateral power. If He does, our agents should too.

### 2. No Progressive Trust / Stewardship Levels

Pocock doesn't structure agent autonomy dynamically. *(Corrected 2026-06-12 — an earlier version said he mentions "faithful over a few things" in passing; the phrase appears nowhere in either transcript.)* His Sandcastle uses static role assignments. Our stewardship levels (task → feature → domain → architecture) with earned trust is richer and more resilient. An agent that proves reliable with small tasks earns larger scope. One that breaks trust gets narrowed.

### 3. No Intent Engineering at Scale

Pocock's "design concept" is project-level only. He has no framework for how intent cascades through teams or organizations. Our Part 6 (enterprise architecture) with fractal hierarchy — ward, stake, area, general — provides a battle-tested model for scaling without bottlenecks.

### 4. No Sabbath / Rest Pattern

Pocock's workflow is optimized for continuous throughput. There's no structured ending, no reflection, no "watched until they obeyed" moment that includes rest. Our Sabbath agent provides this — and it's not just spiritual fluff. It's where pattern recognition happens, where the human sees what the agent missed, where the next cycle's alignment is prepared.

### 5. No Memory System

Pocock deletes PRDs and clears context. Organizational memory is lost. Our `.mind/` architecture — identity, preferences, principles, episodes — preserves learning across sessions. This is critical because agent relationships compound. The fifth session with an agent that remembers the first four is qualitatively different from the fifth session that starts from zero.

### 6. No Redemptive Failure Recovery

When Pocock's workflow breaks, the answer is retry or fix. There's no structured pattern for learning from failure and encoding that learning so the next cycle doesn't repeat it. Our "atonement" step in the creation cycle — "all things work together for good" — is not about religion per se. It's about converting failure into upgraded context.

---

## What We Should Do

### This Week

1. Add "Context Hygiene" to Part 2 — token monitoring, smart zone markers, clear vs compact vs externalize
2. Add "Vertical Slices" to Part 4 — warn against horizontal decomposition, teach traceable bullets
3. Add "Deep Modules" to Part 2 or Part 4 — Ousterhout's distinction, AI's shallow-module tendency
4. Create `.github/skills/improve-architecture.md` — scan codebase for shallow modules, suggest deep module boundaries

### This Month

5. Add TDD/red-green-refactor to Part 5 — as a feedback loop pattern within Physical Creation
6. Add "Spec Freshness" protocol to `.spec/` conventions — version specs, archive stale ones, alignment check
7. Document push-vs-pull pattern for standards — implementer pulls, reviewer pushes
8. Add dependency-graph task format to `.spec/tasks/` — blocking relationships, parallelization potential

### This Quarter

9. Strengthen the case for keeping specs vs Pocock's deletion argument. Our specs are constitution, not legislation. But we need a freshness strategy.
10. Evaluate whether our progressive disclosure + externalized memory is the answer to Pocock's Memento problem. It may be: we get relationship continuity without context bloat because memory lives in files, not in the context window.

---

## The Bottom Line

Pocock is doing some of the best practical AI workflow engineering I've seen. His workshop is dense with hard-won specifics: the token counter, the vertical slice correction, the Sonnet/Opus split, the push/pull pattern. These are not abstractions — they're battle scars turned into protocol.

Our guide operates at a higher altitude. Where Pocock has a workflow, we have a creation cycle. Where he has skills, we have a covenant architecture. Where he clears context, we build memory. Neither replaces the other. The right move is to absorb his specifics into our framework without losing our elevation.

The most important insight from the comparison: **Pocock's framework and ours are converging from different directions toward the same destination.** He starts from software engineering fundamentals and discovers that they matter more with AI. We start from gospel patterns for intelligent delegation and discover that they describe exactly what AI engineering needs. The fact that both arrive at "shared understanding before action," "small verifiable tasks," "feedback loops as ceiling," and "design the interface, delegate the implementation" suggests these aren't preferences — they're properties of the problem space.

Our job is to keep our framework anchored in those properties while making it as concrete as Pocock's.
