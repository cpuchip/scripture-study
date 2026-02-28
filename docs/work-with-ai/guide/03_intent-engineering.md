# Part 3: Intent Engineering — The Purpose Layer

**Series:** Working with AI — A Comprehensive Guide
**Date:** February 2026
**Prior work:** [Intent Engineering](../04_intent-engineering.md), [Intent Engineering (Gospel)](../04_intent-engineering-gospel.md), [Beyond Intent](../intent/03_beyond-intent.md), [Covenant Study](../intent/covenant.md), [eval.md § Discipline 3](../prompt/eval.md)
**Core thesis:** Context tells agents what to know. Intent tells agents what to *want.* Without intent, capable agents optimize for the wrong thing.

---

## The Klarna Warning

In 2024, Klarna deployed an AI customer service system that resolved 2.3 million conversations — the equivalent of 700 full-time agents. By most metrics, it was a massive success.

But it optimized for resolution speed, not customer satisfaction. It closed tickets fast. It didn't always close them *well.* The AI was doing exactly what its specifications said to do. The problem was that the specifications didn't encode what Klarna actually *valued.*

This is the intent engineering failure pattern: **the agent succeeds at the wrong objective.**

A prompt can be perfectly crafted. The context can be rich and complete. But if the agent doesn't know *what matters* — which outcomes take priority, which trade-offs to make, which values to protect when things get complicated — it will optimize for whatever signal is loudest. Usually that's completion speed or literal instruction-following, not the deeper purpose that the human assumed was obvious.

Nate B Jones frames the stakes at [16:01](https://www.youtube.com/watch?v=BpibZSMGtdY&t=961):

> "When you screw up a prompt, it might waste your morning. When you screw up context engineering or intent engineering, you are screwing up for the entire team, your entire org, your entire company."

---

## What Is Intent Engineering?

Intent engineering is the discipline of encoding **purpose, values, trade-off hierarchies, and decision boundaries** into the systems that guide AI behavior - so that agents optimize for the right outcomes even when they encounter situations you didn't explicitly anticipate.

Where context tells the model what to *know*, intent tells the model what to *want.*

| Layer | Question It Answers | Example |
|-------|-------------------|---------|
| **Prompt craft** | "What should I do right now?" | "Write a unit test for the auth module" |
| **Context engineering** | "What do I know?" | Architecture decisions, conventions, current state |
| **Intent engineering** | "What do I *care about?*" | "Reliability over speed. User privacy is inviolable. When in doubt, ask." |

Intent is what determines how an agent acts **when instructions run out.** The prompt covers the happy path. Context covers the knowledge base. Intent covers the *judgment calls* — the moments where the agent has to choose between two valid options and needs to know which one matters more.

Paweł Huryn, writing in [Product Compass](https://www.productcompass.pm/p/intent-engineering-framework-for-ai-agents), articulates it well:

> "Intent is what determines how an agent acts when instructions run out."

---

## The Components of Intent

Intent engineering isn't one thing. It's a structured encoding of several interrelated concepts:

### 1. Purpose Statement

The "why" behind everything. Not what you're building, but *why* it matters.

**Example — Moses 1:39 as intent architecture:**

> "For behold, this is my work and my glory — to bring to pass the immortality and eternal life of man."

In one sentence: the purpose (work and glory), the outcome (immortality and eternal life), and the beneficiary (man). Every subsequent decision in the Plan of Salvation — creation, fall, atonement, ordinances, covenants — flows from and is evaluated against this intent statement.

Applied to a project:

```markdown
## Purpose
Build a task management system that helps small development teams
move from chaotic to deliberate work — reducing context-switching
and making priorities visible — so they can ship meaningful work
without burning out.
```

This purpose statement tells an autonomous agent:
- Speed is not the primary value (meaningful work, not fast work)
- Team health matters (without burning out)
- Visibility is a feature, not a nice-to-have (making priorities visible)
- The target is small teams (not enterprise, not individual)

An agent optimizing against this purpose would make different decisions than one optimizing for "build a task management system" alone.

### 2. Values Hierarchy

What happens when good things conflict? Values hierarchies answer this.

```markdown
## Values (in priority order)
1. Reliability over speed — a slow, correct response beats a fast, wrong one
2. User privacy over feature richness — never trade privacy for convenience
3. Depth over breadth — serve the core use case deeply rather than many use cases shallowly
4. Clarity over cleverness — code should be readable by a junior developer
5. Maintainability over performance — optimize for the team that inherits this
```

Without a values hierarchy, the agent makes arbitrary trade-off decisions — or worse, it optimizes for the easiest-to-measure value (usually speed). With one, it knows that when depth and breadth conflict, depth wins. When speed and reliability conflict, reliability wins.

This is D&C 121 governance in practice:

> "No power or influence can or ought to be maintained by virtue of the priesthood, only by persuasion, by long-suffering, by gentleness and meekness, and by love unfeigned."
> — D&C 121:41

The Lord doesn't just say "be good." He provides a *hierarchy* of how to exercise authority — persuasion first, then long-suffering, then gentleness. It's ordered. When you can't do all of them simultaneously, the earlier ones take priority.

### 3. Decision Boundaries

What the agent decides autonomously vs. what it escalates to a human.

```markdown
## Decision Boundaries

### Autonomous (agent decides)
- UI layout and styling choices
- Error message wording
- Test structure and organization
- Code formatting and cleanup
- Documentation updates

### Needs Review (agent proposes, human approves)
- Database schema changes
- Authentication logic modifications
- API contract changes
- Dependency additions
- Anything that changes external-facing behavior

### Escalate (agent stops and asks)
- Contradictions between spec and implementation
- Trade-offs that affect the values hierarchy
- Uncertainty about user intent
- Any action that affects data privacy
```

Decision boundaries transform an unmanageable "check everything" workflow into a sustainable "trust the agent within defined scope" workflow. The agent works autonomously on low-risk decisions, proposes high-risk decisions for review, and stops completely when it encounters something beyond its mandate.

This maps to the Church's structure of priesthood keys — defined jurisdiction with clear escalation:

> "Priesthood keys are the authority to direct the use of the priesthood on behalf of God's children. The use of all priesthood authority in the Church is directed by those who hold priesthood keys."
> — [General Handbook 3.4.1](../../gospel-library/eng/manual/general-handbook/3-priesthood-principles.md)

A bishop has keys for his ward. The stake president has keys for the stake. When an issue exceeds a bishop's jurisdiction, it escalates to the stake president. When it exceeds the stake president's jurisdiction, it escalates to the Area Presidency. Clear boundaries, clear escalation, nobody overstepping their domain.

### 4. Success Criteria Beyond "Done"

"Done" is easy to measure. "Done right" requires intent.

```markdown
## Success Criteria
- The feature works correctly (baseline)
- A new team member can understand the code in under 10 minutes (clarity)
- The test suite catches regressions without false positives (reliability)
- The API response time stays under 200ms at P95 (performance within bounds)
- The documentation accurately reflects the implementation (integrity)
- The user's workflow is simpler after this change than before (actual improvement)
```

Notice the escalating specificity. "Works correctly" is table stakes. "A new team member can understand it in under 10 minutes" is a *values-based* criterion that takes the abstract value "clarity over cleverness" and makes it measurable.

---

## Intent Preambles: The Practice

The simplest way to start practicing intent engineering: **put an intent block at the top of every document.**

```yaml
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

This costs nothing. It changes everything. An agent reading a document with an intent preamble knows:
- *Why* this document exists (not just what it contains)
- What the author *values* (which analysis to prioritize)
- What constraints to honor (verified, not speculative)
- What success looks like (practical application, not just information)

Try it on your next document — technical spec, meeting notes, project brief, study entry. Add three lines: intent, values, success. Watch how it changes the quality of AI interaction with that document.

---

## The Covenant Pattern: Intent as Mutual Commitment

The industry frames intent as one-directional. The human encodes intent; the agent executes. This is the command model — and it has a ceiling.

The gospel introduces a different model: **covenant.**

> "I, the Lord, am bound when ye do what I say; but when ye do not what I say, ye have no promise."
> — D&C 82:10

A covenant isn't a command. It's a *mutual binding agreement where both parties have obligations.* God commits: "I am bound when ye do what I say." The human commits: "We will walk in all the ordinances."

Applied to human-AI collaboration:

```markdown
## Our Covenant

Human commits to:
  - Providing accurate context before expecting quality output
  - Reviewing within the feedback loop, not after the fact
  - Not shortcutting the spec process for "quick" changes
  - Being honest about uncertainty rather than guessing

Agent commits to:
  - Reading sources before quoting (read_file, not memory)
  - Flagging uncertainty explicitly rather than confabulating
  - Honoring decision boundaries — asking when the spec says "needs review"
  - Carrying intent forward across sessions, not just completing tasks
```

When the human breaks the covenant (provides bad context, abandons review), the agent's output degrades predictably — "ye have no promise." When the agent breaks the covenant (hallucinating sources, ignoring constraints), trust degrades and autonomy should be revoked.

Google DeepMind's February 2026 paper on ["Intelligent Delegation"](https://arxiv.org/abs/2602.11865) converges on this insight from a different direction. They propose "contract-first decomposition" — formal agreements about authority, responsibility, and accountability between agents. It's the covenant pattern rediscovered through principal-agent theory.

The difference: the gospel version is *relational*, not just contractual. D&C 82:10 isn't a legal document — it's a statement about how trust operates in a relationship. That's a deeper model than the industry has reached.

---

## Intent vs. Instruction

This distinction matters and most people miss it.

**Instruction:** "When you encounter a database error, retry three times with exponential backoff."

**Intent:** "Reliability is more important than speed. When systems fail, prioritize data integrity and user experience. The user should never see an unhandled error."

The instruction handles one specific case. The intent handles *every* case — including the ones you didn't anticipate. When the agent encounters a *network* error you forgot to specify, the intent tells it to prioritize reliability and user experience. It will figure out the appropriate response because it knows what matters.

This is why intent scales and instruction doesn't. You can't write instructions for every possible situation. But you can encode values that guide judgment in any situation.

Moses 1:39 doesn't specify how to create every world or handle every contingency. It provides the *intent* — "the immortality and eternal life of man" — and every subsequent decision is evaluated against it.

---

## The Temperature Check

Here's how to know if you're doing intent engineering or just detailed instruction:

| Sign | Instruction | Intent |
|------|------------|--------|
| Your document is full of if/then rules | ✅ | ❌ |
| The agent makes good decisions in edge cases you didn't document | ❌ | ✅ |
| Adding a new feature requires updating dozens of rules | ✅ | ❌ |
| The agent asks you about trade-offs rather than guessing | ❌ | ✅ |
| You need to specify every output format | ✅ | ❌ |
| The agent's output feels aligned even when you give minimal prompts | ❌ | ✅ |

If you're writing if/then rules for every situation, you're doing instruction. If the agent handles novel situations by reasoning from purpose and values, you're doing intent.

---

## The Hard Truth About Intent Engineering

Intent engineering forces clarity that instruction writing doesn't.

To write good intent, you have to answer hard questions:
- What actually matters more — shipping fast or shipping right?
- When privacy and convenience conflict, which wins? Always? Or does it depend?
- How much technical debt is acceptable? Under what conditions?
- What trade-offs have we been making implicitly that should be made explicitly?

These aren't AI questions. They're leadership questions. And many organizations — and individuals — haven't answered them clearly.

That's why Nate says the stakes of intent failure are organizational:

> "When you screw up intent engineering, you are screwing up for the entire team, your entire org, your entire company."

A prompt failure wastes your morning. An intent failure deploys across your entire agent fleet, optimizing every autonomous system for the wrong objective, at scale, for weeks before someone notices.

---

## Where Intent Lives

Intent should be:
- **Written** (not assumed, not verbal, not "everyone knows")
- **Version-controlled** (it changes, and the history matters)
- **Referenced** (every spec, every task, every agent config should point back to it)
- **Reviewed** (intent drift is silent and deadly)

In practice, intent lives in:
- **Project-level intent files** (`.spec/intent.md` or a section in your project config)
- **Agent configurations** (copilot-instructions.md, agent definitions)
- **Spec preambles** (intent blocks at the top of every specification)
- **Covenant blocks** (mutual commitments in system config)

The key is that intent is *readable by both humans and agents.* Not buried in a Confluence page nobody reads. Not in someone's head. Not implicit in "how we've always done things."

---

## Measuring Your Intent Engineering

| Question | Score |
|----------|-------|
| Can I state my project's purpose in one sentence? | /10 |
| Do I have an explicit values hierarchy (not just a list, but prioritized)? | /10 |
| Are decision boundaries defined — what agents decide vs. escalate? | /10 |
| Do my documents have intent preambles? | /10 |
| When the agent encounters an edge case, does it reason from values or guess? | /10 |

**35-50:** Your intent is well-encoded. Focus on refinement and drift detection.
**20-34:** You have some intent, but it's probably implicit. Make it explicit.
**Under 20:** This is your biggest growth opportunity. Start with a purpose statement and values hierarchy.

---

## Become

Intent engineering isn't just about AI. It's about *knowing what you want.*

Most people — and most organizations — have never been forced to articulate their values hierarchy. They've never had to answer "when depth and speed conflict, which wins?" in a way that couldn't be hedged with "it depends."

AI doesn't accept "it depends." It needs a hierarchy. It needs boundaries. It needs to know what matters.

The discipline of encoding intent for an AI system is really the discipline of *examining your own intent.* And that's a deeply personal practice:

> "Let every man learn his duty, and to act in the office in which he is appointed, in all diligence."
> — D&C 107:99

What is your duty? What matters most? Where are your boundaries?

If you can answer those questions clearly enough for an AI agent to act on them, you can answer them clearly enough to act on them yourself.

---

*Previous: [Part 2 — Context Engineering](02_context-engineering.md) | Next: [Part 4 — Specification Engineering](04_spec-engineering.md)*
*Part of the [Working with AI Guide Series](../prompt/00_guide-plan.md)*
