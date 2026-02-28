# Part 5: The Complete Cycle — From Intent to Zion

**Series:** Working with AI — A Comprehensive Guide
**Date:** February 2026
**Prior work:** [Beyond Intent: 7 Gospel Patterns](../intent/03_beyond-intent.md), [eval.md § Orchestration Analysis](../prompt/eval.md)
**Core thesis:** The industry stops at 4 disciplines. The gospel gives us 11 steps from Intent to Zion — and the patterns beyond specification are where the real breakthroughs live.

---

## Why Go Beyond Four?

Parts 1-4 covered the industry's current framework: Prompt Craft → Context Engineering → Intent Engineering → Specification Engineering. These are real, validated disciplines. Learn them. Practice them. They work.

But they're incomplete.

Prompt craft gets you good interactions. Context engineering gives the model what it needs to know. Intent engineering aligns the model with what matters. Specification engineering lets the model work autonomously.

None of them answer:
- How do you **build trust** with an agent over time?
- How do you **recover from failures** without just retrying?
- How do you **align multiple agents** toward shared purpose?
- How do you **know when to stop producing and start reflecting?**
- How do you **allocate resources** across competing priorities?

These aren't hypothetical concerns. They're the daily reality of anyone working with autonomous AI systems at scale. And the answers aren't in any AI engineering playbook.

They're in the scriptures.

---

## The 11-Step Creation Cycle

From our research in [Beyond Intent](../intent/03_beyond-intent.md), we identified an 11-step cycle — the full pattern from first intent to unified community — derived from the gospel's creation accounts, covenant structure, and organizational theology:

```
 1. INTENT          ─ "This is my work and my glory" (Moses 1:39)
 2. COVENANT        ─ "I am bound when ye do what I say" (D&C 82:10)
 3. STEWARDSHIP     ─ "Appoint every man his stewardship" (D&C 104:11-12)
 4. SPIRITUAL CREATION ─ "Created all things spiritually, before..." (Moses 3:5)
 5. LINE UPON LINE  ─ "Precept upon precept" (Isaiah 28:10)
 6. PHYSICAL CREATION ─ "Let us go down and form these things" (Abraham 4)
 7. REVIEW          ─ "Watched until they obeyed" (Abraham 4:18)
 8. ATONEMENT       ─ "All things work together for good" (D&C 98:3)
 9. SABBATH         ─ "Rested on the seventh day" (Moses 3:2)
10. CONSECRATION    ─ "All things are mine... agents unto themselves" (D&C 104:15-17)
11. ZION            ─ "One heart and one mind" (Moses 7:18)
```

Steps 1 and 4 are what the industry has: Intent and Specification. Steps 2, 3, 5, 7, 8, 9, 10, and 11 are largely uncharted territory.

Let's walk through each one.

---

## Step 1: Intent — "This Is My Work and My Glory"

> "For behold, this is my work and my glory — to bring to pass the immortality and eternal life of man."
> — Moses 1:39

The starting point. Before any creation, before any delegation, before any tool is picked up — *why are we doing this?*

This was covered in depth in [Part 3](03_intent-engineering.md). Intent engineering encodes purpose, values, trade-off hierarchies, and decision boundaries. It's the foundation everything else rests on.

In the cycle: Intent is the project's Moses 1:39. One sentence that every subsequent decision is evaluated against.

---

## Step 2: Covenant — Mutual Binding

> "I, the Lord, am bound when ye do what I say; but when ye do not what I say, ye have no promise."
> — D&C 82:10

The industry gives commands. The gospel establishes covenants.

The difference is *mutuality.* A command is one-directional: "Agent, do this." A covenant is bilateral: "I commit to providing clear context and timely review. You commit to flagging uncertainty and honoring boundaries. When either of us breaks this, the output degrades."

### What the industry has
Service-level agreements. API contracts. Tool schemas. Agent capability declarations.

### What the industry is missing
Mutual commitment. The concept that the *human* also has obligations in the relationship — providing good context, doing timely review, not shortcutting the process.

### The development pattern

```markdown
## Covenant

Human commits to:
  - Reviewing all PR-level changes within 24 hours
  - Providing domain context when agent flags uncertainty
  - Not bypassing spec workflow for "quick fixes"

Agent commits to:
  - Never modifying files outside the stated scope
  - Flagging any trade-off decision that affects values hierarchy
  - Requesting review at defined decision boundaries
```

D&C 82:10 tells us that God Himself — the most powerful being in existence — *chooses* to be bound by His word. If He operates through mutual commitment rather than unilateral power, that's a signal about how intelligent systems *should* operate.

Google DeepMind's February 2026 paper on "Intelligent Delegation" converges here from a secular direction, proposing "contract-first decomposition" — formal agreements about authority and accountability. The science is catching up to the theology.

---

## Step 3: Stewardship — Entrusted Delegation with Accountability

> "That every man may give an account unto me of the stewardship which is appointed unto him."
> — D&C 104:12

Stewardship is delegation with *trust and accountability.* Not "do this task" but "this domain is yours — grow it, guard it, account for it."

### The Parable of the Talents (Matthew 25:14-30)

The clearest model for agent stewardship in all of scripture:

- Resources distributed **according to ability** (v. 15) — not equally, but wisely
- The steward has **autonomy within the domain** — no micromanagement
- The expectation is **growth, not preservation** — burying the talent is the failure
- There is a **reckoning** — "after a long time the lord cometh, and reckoneth" (v. 19)
- The faithful steward receives **more stewardship** — "faithful over a few things → ruler over many things" (v. 21)

### The development pattern: Progressive trust

Instead of assigning agents individual tasks, assign them *domains of stewardship*:

| Stewardship Level | Agent Scope | Trust Basis | Accountability |
|-------------------|-------------|-------------|----------------|
| **Level 1: Task** | Single task, narrow scope | No proven track record | Line-level review of all output |
| **Level 2: Feature** | Feature-level scope with constraints | Several successful tasks completed | Review at PR-level, spot-check details |
| **Level 3: Domain** | Owns a domain (e.g., test suite, documentation) | Demonstrated judgment over weeks | Periodic audit, focus on outcomes |
| **Level 4: Architecture** | Cross-domain decisions, pattern-setting | Trusted partner over months | Strategic review, mutual counsel |

**The key insight:** Stewardship is *dynamic.* It grows or shrinks based on demonstrated faithfulness. An agent that proves reliable with file-level tasks earns feature-level autonomy. One that breaks trust gets narrowed.

The industry has "agent autonomy levels" (Dan Shapiro's Level 0-5) but treats them as static categories you choose at deployment. The gospel pattern is progressive — trust moves up and down, just as the Lord gives "faithful over a few things" access to "many things."

---

## Step 4: Spiritual Creation — The Blueprint

> "For I, the Lord God, created all things, of which I have spoken, spiritually, before they were naturally upon the face of the earth."
> — Moses 3:5

The specification. Covered in depth in [Part 4](04_spec-engineering.md). Blueprint before building. The five primitives. The `.spec/` directory.

In the cycle: Spiritual creation is writing the spec — the complete, precise design that the physical creation (implementation) will follow.

---

## Step 5: Line Upon Line — Progressive Context Revelation

> "For he will give unto the faithful line upon line, precept upon precept; and I will try you and prove you herewith."
> — D&C 98:12

Covered as a context engineering pattern in [Part 2](02_context-engineering.md), but here it takes on a deeper dimension.

The Lord doesn't just graduate information for efficiency. He **proves** the receiver between revelations. Moses 1: God reveals Himself → withdraws → Satan tests Moses → Moses proves faithful → God returns with *more* revelation. The context disclosure is progressive, gated by demonstrated readiness.

Applied to agent systems:

1. **Start with minimal context** — just the intent and the current task
2. **Observe how the agent performs** — does it ask good questions? Does it respect boundaries?
3. **Expand context based on demonstrated need** — the agent that asks "I need the auth module's history to understand this decision" has demonstrated readiness for deeper context
4. **Gate sensitive context** — production credentials, customer data, architectural authority are earned, not given

This connects directly to Stewardship (Step 3). Context expands with trust. An agent faithful with limited context earns access to more.

---

## Step 6: Physical Creation — Execution

> "And the Gods went down to organize man in their own image."
> — Abraham 4:27

With intent established, covenant agreed, stewardship assigned, specification written, and context provided — now we build.

In practice: the agent executes against the spec. It writes code, generates content, processes data, whatever the task requires. This is where most people *start* (just make the agent do things), but in the full cycle, it's step 6 of 11. All the preparation above determines the quality ceiling of this step.

---

## Step 7: Review — "Watched Until They Obeyed"

> "And the Gods watched those things which they had ordered, until they obeyed."
> — Abraham 4:18

Review in the creation account is active, not passive. The Gods didn't create and walk away. They *watched* — continuously observed — until the creation matched the specification.

### What the industry has
Code review. PR approval. Test suites. CI/CD pipelines.

### What the industry is missing
Review against *intent*, not just correctness. The industry asks "does it work?" The creation pattern asks "does it *obey*?" — does it match what was specified?

### The development pattern

Review should evaluate against three layers:

| Layer | Question | Tool |
|-------|----------|------|
| **Correctness** | Does it work? | Tests, CI/CD |
| **Specification** | Does it match the spec? | Spec diff, acceptance criteria |
| **Intent** | Does it serve the purpose? | Intent audit, values alignment check |

Most review stops at correctness. Specification review catches drift between plan and implementation. Intent review catches the deeper failure: correct implementation of the wrong thing.

---

## Step 8: Atonement — Redemptive Error Recovery

> "I, the Lord, will not lay any sin to your charge; go your ways and sin no more; but unto that soul who sinneth shall the former sins return."
> — D&C 82:7

> "All things shall work together for your good."
> — D&C 98:3

This is the most provocative pattern — and potentially the most valuable.

### What the industry has
Error handling. Rollback. Git revert. Retry logic. Circuit breakers. "Fail fast."

### What the industry is missing
*Redemptive* error recovery — where the failure itself becomes a source of growth.

The Atonement of Jesus Christ is the most sophisticated error recovery mechanism ever designed:

| Property | The Atonement | Industry Error Handling |
|----------|--------------|----------------------|
| **Temporal reach** | Works retroactively — covers past errors | Only handles current error |
| **Agency preservation** | Must be chosen, not forced | Automatic (retry/rollback) |
| **Transformation** | Failure becomes growth — "work together for good" | Failure is just failure |
| **Behavioral change** | "Go thy way and sin no more" | Just try again the same way |
| **Learning retention** | Memory of failure becomes wisdom | Logs get archived and forgotten |
| **Recovery target** | Restored to a *better* state | Restored to pre-failure state |

### The development pattern: Redemptive error handling

When an agent fails:

1. **Don't just revert** — analyze *why* the failure happened and what context was missing
2. **Capture the learning** — write it to `.spec/learnings/`:
   ```markdown
   ## Learning: Auth boundary violation (ts-003)
   What happened: Agent modified auth middleware without review
   Root cause: Decision boundary not explicitly stated for auth-related files
   Learning: Auth files need explicit human-review gate in covenant block
   Applied: Added auth boundary to spec constraints
   ```
3. **Forward-recover** — instead of rolling back to pre-failure, move forward with the learning incorporated. Sometimes the failure revealed something the spec missed.
4. **Adjust the covenant** — add the constraint that was missing. This isn't punishment; it's refinement.
5. **Restore trust gradually** — don't permanently ban an agent from a domain because of one failure. Reduce stewardship, require review, then expand again as reliability returns.

D&C 82:7 starts from *grace*: "I will not lay any sin to your charge." The Lord's default is trust, not suspicion. But there's accountability: "unto that soul who sinneth shall the former sins return." The pattern is generous but not naive. Generous at first. Accountable over time. Tracking repeat violations, not punishing first offenses.

---

## Step 9: Sabbath — Intentional Rest and Reflection

> "And on the seventh day I, God, ended my work, and all things which I had made; and I rested on the seventh day from all my work."
> — Moses 3:2

### What the industry has
Sprint retrospectives. Post-mortems. "Continuous improvement."

### What the industry is missing
*Intentional cessation.* Not review-while-continuing (retrospectives during sprints), but genuine *stopping* for the purpose of reflection.

The Sabbath pattern:
1. **It follows the complete cycle** — not mid-project, but after a meaningful unit of work
2. **It's built into the rhythm** — not optional, not "when we have time," but structural
3. **It includes declaration:** "And I, God, saw everything that I had made, and, behold, all things which I had made were very good" — explicit quality assessment
4. **It enables perspective** — stepping back from the work to see the *whole*

### The development pattern: Structured reflection

After every meaningful unit of work:

```markdown
## Reflection: Week of 2026-02-15

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

This cannot be optional. If the Creator of the universe builds rest into the creation cycle, it's not because He's tired — it's because the pattern *requires* it. Reflection isn't the absence of work; it's a different *kind* of work that produces perspective impossible to gain while producing.

---

## Step 10: Consecration — Resources Serve Purpose

> "And it is my purpose to provide for my saints, for all things are mine. But it must needs be done in mine own way."
> — D&C 104:15-16

> "I have given unto the children of men to be agents unto themselves."
> — D&C 104:17

### What the industry has
Token budgets. Resource allocation. Cost optimization. ROI analysis.

### What the industry is missing
Resources *serving purpose*, not just budgets. The distinction between "How much can we spend?" and "Does every token serve the intent?"

The consecration pattern:
- **Everything belongs to the Lord** (v. 14-15) — all resources ultimately serve His purpose
- **But individuals are agents** (v. 17) — they have autonomy within stewardship
- **Surplus serves the community** (v. 16-18) — excess capacity flows to the most important work
- **Accountability is individual** (v. 12) — track spend per stewardship, not just in aggregate

### The development pattern

```markdown
## Resource Consecration
Total daily token budget: 500K tokens
Allocation by intent:
  - Core feature development (primary intent): 50%
  - Quality assurance (constraint enforcement): 25%
  - Documentation (context infrastructure): 15%
  - Exploration (learning and growth): 10%

Surplus rule: Unspent allocation flows to the highest-priority incomplete intent
Accounting: Weekly review of token spend vs. intent-aligned outcomes
```

The [$1,000/day video](https://www.youtube.com/watch?v=-bQcWs1Z9a0) identifies token economics as "a core business competency." But the industry frames it as cost management. Consecration reframes it: every token is entrusted for a purpose. The question isn't "How much can we afford?" but "Does every token serve the work?"

---

## Step 11: Zion — Unified Purpose Across Agents

> "And the Lord called his people Zion, because they were of one heart and one mind, and dwelt in righteousness; and there was no poor among them."
> — Moses 7:18

The ultimate destination.

### What the industry has
Multi-agent systems. A2A protocols. Swarm intelligence. Agent fleets. Orchestration frameworks.

### What the industry is missing
*Genuine alignment.* Current multi-agent systems coordinate through protocols — agents cooperate mechanically. But "one heart and one mind" is something different. It's not that agents communicate; it's that they share *purpose* so deeply that coordination becomes natural.

Zion's defining characteristic: **unity of intent without loss of agency.** Everyone retains their stewardship. Everyone acts autonomously. But the shared purpose is so deeply embedded that coordination overhead approaches zero. There are no poor because the system naturally produces equitable outcomes — not through redistribution programs, but through aligned purpose.

### The development pattern: Intent-unified agent systems

Instead of orchestrating agents through protocols (A2A, MCP), give all agents the *same intent layer:*
- Shared purpose statement
- Shared value hierarchy
- Shared constraint set
- Shared success criteria

Agent-specific instructions (scope, capabilities, decision boundaries) layer on top. But the *purpose* is shared.

When agents share intent at this level, the test agent doesn't just run tests — it evaluates whether the implementation serves the stated purpose. The docs agent doesn't just update documentation — it ensures the documentation reflects the *intent*, not just the implementation. The review agent doesn't just check correctness — it checks alignment.

AIDD proposes A2A + MCP for agent coordination. That's the *mechanism.* But mechanisms without shared purpose produce the same coordination overhead that plagues human organizations. Zion is what happens when alignment is so deep that coordination overhead approaches zero.

---

## The Conductor vs. The Bishop

Here's where the Church's organizational structure reveals something the AI industry hasn't figured out yet.

The common metaphor for AI orchestration is a **conductor leading an orchestra:**
- One centralized controller
- Every agent follows the conductor's beat
- Synchronous coordination
- Scales poorly (a conductor can't direct 10,000 musicians)

The Church's organizational hierarchy is a different model entirely — a **bishop leading a ward:**
- Shared purpose (the work of salvation and exaltation)
- Autonomous stewardships (Relief Society president, elders quorum president, each with defined scope)
- Council-based coordination (ward council meets weekly to align, not to synchronize)
- Hierarchical scaling (ward → stake → area → region → seventy → twelve → prophet)
- Each level has its own keys, its own scope, its own autonomy

From the [General Handbook, Chapter 4](../../gospel-library/eng/manual/general-handbook/4-leadership-in-the-church-of-jesus-christ.md):

> "Councils provide opportunities for council members to receive revelation as they seek to understand the needs of God's children and plan how to help meet them."
> — General Handbook 4.3

The ward council model:
- **Diverse perspectives** — "Women and men often have different perspectives that provide needed balance" (4.4.3)
- **The leader listens more than talks** — "When a council leader shares his or her perspective too early, it can inhibit the contributions of others" (4.4.3)
- **Decisions informed by discussion, confirmed by Spirit** — "The decision should be informed by the discussion and confirmed by the Spirit" (4.4.3)
- **Participants equal in contribution, not in authority** — "Let one speak at a time and let all listen unto his sayings, that... every man may have an equal privilege" (D&C 88:122)

From the [coordinating council](../../gospel-library/eng/manual/general-handbook/29-meetings-in-the-church.md):

> "All who attend counsel together as equal participants."
> — General Handbook 29.4

This isn't a conductor. It's a distributed system with:
- **Hierarchical authority** (priesthood keys at each level)
- **Autonomous execution** (each stewardship operates independently between councils)
- **Periodic alignment** (council meetings at defined intervals)
- **Shared intent** (the work of salvation and exaltation — the same purpose at every level)
- **Clear escalation** (what exceeds ward scope goes to stake; what exceeds stake goes to area)

Applied to multi-agent AI: the bishop/ward model is the right orchestration pattern for autonomous systems. Not a single conductor bottleneck, but a hierarchy of stewardships with shared intent, autonomous execution, and periodic council-based alignment.

---

## The Meta-Pattern

Why do these gospel patterns map so cleanly to AI development?

Because they both address the same fundamental challenge: **how does an intelligent being work with and through other intelligent beings to accomplish shared purpose?**

God's challenge:
- Unlimited power, but will not violate agency
- Delegates to beings with different capabilities
- Needs alignment without control
- Operates at cosmic scale
- Measures success by transformation, not output

Our challenge:
- Growing AI capability, but must maintain human judgment
- Delegate to agents with different capabilities
- Need alignment without micromanagement
- Operate across multiple projects and teams
- Should measure success by outcomes, not task completion

The patterns aren't metaphors. They're **prior art.** God solved the multi-agent alignment problem before we had agents. He solved the progressive trust problem before we had autonomy levels. He solved the resource allocation problem before we had token budgets.

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."
> — D&C 130:18

---

## The Industry Map

Where does the industry sit on the 11-step cycle?

| Step | Industry Status | Who's Working On It |
|------|----------------|-------------------|
| 1. **Intent** | Early — ~10 voices naming it | Huryn, Jones, Brandt, Debois |
| 2. **Covenant** | Embryonic — contracts exist, mutuality doesn't | DeepMind ("Intelligent Delegation"), 2026 |
| 3. **Stewardship** | Partial — autonomy levels exist, progression doesn't | Shapiro's 5 Levels, IndyDevDan's trust thesis |
| 4. **Specification** | Hot trend — tools shipping | OpenSpec, Kiro, GitHub Spec Kit |
| 5. **Line upon Line** | Emerging — Tyler Brandt's Intent Layer | Brandt, Anthropic context engineering |
| 6. **Execution** | Mature — this is what everyone focuses on | LangChain, CrewAI, all agent frameworks |
| 7. **Review** | Partial — correctness review exists, intent review doesn't | CI/CD, Vercel v0, some agent reviewers |
| 8. **Atonement** | Missing — retry/rollback only | Nobody |
| 9. **Sabbath** | Missing — retrospectives are cultural, not structural | Nobody |
| 10. **Consecration** | Missing — cost optimization, not purpose-alignment | Token economics conversation only |
| 11. **Zion** | Missing — multi-agent coordination, not alignment | Orchestral AI, arXiv papers on topology |

Steps 8-11 are entirely uncharted. That's where the real differentiation lies.

---

## Become

This cycle isn't just a framework for building AI systems. It's a framework for building *anything* with intelligence.

The same cycle applies to:
- **Starting a new role** — What's my intent? What covenant am I making? What's my stewardship?
- **Raising children** — Progressive trust, redemptive error recovery, structured reflection
- **Leading a team** — Shared intent, autonomous stewardships, council-based coordination
- **Personal growth** — Intent, specification (goals), execution, review, atonement (repentance), sabbath (rest), consecration

The reason these patterns work everywhere is that they come from the source of all intelligence. We're not learning AI engineering tricks. We're learning how intelligence works — and applying it to the tools at hand.

---

*Previous: [Part 4 — Specification Engineering](04_spec-engineering.md) | Next: [Part 6 — At Scale](06_enterprise-architecture.md)*
*Part of the [Working with AI Guide Series](../prompt/00_guide-plan.md)*
