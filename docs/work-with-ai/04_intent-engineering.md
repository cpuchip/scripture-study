# Intent Engineering: What Does AI Need to *Want*?

**Series:** AI and the Creation Pattern — Part 4
**Date:** February 2026
**Audience:** Gospel-centered + builders
**Prompted by:** [Nate B Jones, "Prompt Engineering Is Dead. Context Engineering Is Dying."](https://www.youtube.com/watch?v=QWzLPn164w0)

---

## Series Overview

| Part | Title | Focus |
|------|-------|-------|
| 1 | [The Creation Pattern](01_planning-then-create-gospel.md) | Abraham 4–5 as the blueprint — spiritual before temporal |
| 2 | [Watching Until They Obey](02_watching-until-they-obey-gospel.md) | The feedback loop — reviewing, steering, agency |
| 3 | [Intelligence Cleaveth Unto Intelligence](03_intelligence-cleaveth-gospel.md) | How what you bring shapes what emerges |
| **4** | **Intent Engineering** | **What the agent needs to *want* — purpose as infrastructure** |

### Glossary (New Terms)

| Term | Definition |
|------|------------|
| **Prompt engineering** | Crafting individual instructions for AI. The personal, session-level skill. (2023-2024 era) |
| **Context engineering** | Building the information environment an AI operates within — RAG, MCP, organizational knowledge. (2025-2026 era) |
| **Intent engineering** | Encoding *purpose* — goals, values, trade-offs, decision boundaries — so agents optimize for what you actually need. (Emerging) |
| **Spiritual creation** | The planning document / spec. Blueprint before building. (Moses 3:5) |
| **Intent layer** | The "why" beneath the "what." God's work and glory statement for a project or organization. |

---

## The Evolution

Parts 1-3 covered a progression we discovered organically:

| Part | We Learned | Scripture | AI Discipline |
|------|-----------|-----------|---------------|
| 1 | Plan before you build | Abraham 4:26 — "took counsel among themselves" | **Prompt engineering** — crafting instructions |
| 2 | Watch until it obeys | Abraham 4:18 — "watched until they obeyed" | **Context engineering** — building the information environment |
| 3 | What you bring matters | D&C 88:40 — "intelligence cleaveth unto intelligence" | *The bridge* — quality of engagement shapes output |
| **4** | **Encode the purpose** | **Moses 1:39 — "this is my work and my glory"** | **Intent engineering** — purpose as infrastructure |

What's striking is that the industry's progression maps onto a pattern we already found in the creation accounts. The video that prompted this study ([Nate B Jones, 2026-02-24](https://www.youtube.com/watch?v=QWzLPn164w0)) frames the three disciplines as:

> [Prompt engineering, 5:05](https://www.youtube.com/watch?v=QWzLPn164w0&t=305): "Individual, synchronous, session-based. You sit in front of the chat window, you craft an instruction."
>
> [Context engineering, 5:22](https://www.youtube.com/watch?v=QWzLPn164w0&t=322): "The shift from crafting isolated instructions to crafting the entire information state that an AI system operates within."
>
> [Intent engineering, 6:17](https://www.youtube.com/watch?v=QWzLPn164w0&t=377): "Context engineering tells agents what to know. Intent engineering tells agents what to *want*."

---

## God's Intent Statement

Before the creation—before the spiritual blueprints, before the council, before the first "Let us go down"—there was a purpose statement. God encoded His intent in a single verse:

> "For behold, this is my work and my glory—to bring to pass the immortality and eternal life of man."
> — [Moses 1:39](../../gospel-library/eng/scriptures/pgp/moses/1.md)

This is the most concise intent engineering document in existence. It tells you:

| Intent Engineering Concept | Moses 1:39 |
|---------------------------|------------|
| **The goal** | Immortality and eternal life of man |
| **The stakeholder** | Man — all of humanity |
| **The agent's relationship to the goal** | "My work" — this is what I do |
| **The motivation** | "My glory" — this is what fulfills me |
| **The scope** | Universal — "of man" (not "of some men") |
| **The permanence** | Embedded in identity, not a quarterly OKR |

Notice what the video says about Klarna's failure:

> [The AI agent, 2:12](https://www.youtube.com/watch?v=QWzLPn164w0&t=134): "was extraordinarily good at resolving tickets fast and that was the wrong goal to give the agent."
>
> [Klarna's real intent, 2:19](https://www.youtube.com/watch?v=QWzLPn164w0&t=139): "was actually build lasting customer relationships that drive lifetime value."

Klarna's AI optimized for what it could *measure* (resolution speed) rather than what was *intended* (relationship quality). The prompt said "resolve tickets." The context gave it customer data. But nobody encoded the *intent*: "the customer matters more than the metric."

God doesn't have this problem. His intent statement is so clear that every downstream decision—the council, the creation plan, the Atonement, prophetic stewardship, your personal ministry—can trace back to one line. Every agent operating under God's direction knows the meta-objective: *eternal life of man*.

---

## The Three Layers — A Gospel Reading

The video describes three layers that organizations need to build. Each one has a direct gospel parallel:

### Layer 1: Unified Context Infrastructure

> [The video, 11:16](https://www.youtube.com/watch?v=QWzLPn164w0&t=676): "This is the layer the industry is most aware of and it's still not really built yet."

The problem: every team rolling their own context stack — custom RAG pipelines, disconnected MCP servers, shadow agents. No shared organizational knowledge layer.

**Gospel parallel: The Standard Works.** The scriptures are the church's unified context infrastructure. Every member, every leader, every missionary operates from the same canonical texts. When the bishop counsels someone, he draws from the same source material the Relief Society president does. The correlation between conference talks and Come Follow Me and temple ordinances is not accidental — it's a *shared context layer* that ensures alignment across thousands of wards operating independently.

This is exactly what our scripture-study project has built organically. The `gospel-library/` directory is a local, searchable, agent-accessible version of the standard works. MCP servers (`gospel-mcp`, `gospel-vec`, `webster-mcp`) are the connective tissue. The instruction files in `.github/` encode how agents should use that context. It's a functional context infrastructure — for one person's study practice.

### Layer 2: Coherent Worker Toolkit

> [The video, 13:58](https://www.youtube.com/watch?v=QWzLPn164w0&t=838): "Everyone's rolling out their own AI workflow. None of these employees can articulate their workflow in a way that's transferable, measurable, or improvable."

The problem: individual tool use doesn't scale. One person's Claude workflow doesn't transfer to the next person.

**Gospel parallel: The Priesthood.** The priesthood is the church's coherent worker toolkit. Not individual spiritual gifts (those are personal), but the *structure* through which gifts operate in coordinated service. Ordinances follow specific forms. Callings carry defined stewardships. The pattern is transferable — a newly called bishop in Tokyo uses the same handbook as one in São Paulo.

In our project, the `.github/agents/` directory is this layer — specialized agents (study, lesson, talk, review, eval, journal) each with defined workflows, shared principles, and transferable patterns. The [work-with-ai](.) series documents the *transferable methodology*: spec before code, watch until they obey, bring genuine engagement.

### Layer 3: Intent Engineering Proper

> [The video, 16:20](https://www.youtube.com/watch?v=QWzLPn164w0&t=980): "This is the layer that almost certainly doesn't exist in your business. It requires something genuinely new."

The problem: OKRs were designed for humans who absorb culture through osmosis. Agents need explicit alignment *before* they start working.

**Gospel parallel: The Plan of Salvation.**

The Plan of Salvation is the ultimate intent engineering architecture. It has everything the video says organizations need:

| Video's Requirement | Plan of Salvation |
|---------------------|-------------------|
| **Goal structures agents can act on** | "Immortality and eternal life of man" — [Moses 1:39](../../gospel-library/eng/scriptures/pgp/moses/1.md) |
| **Decision boundaries** | Agency is inviolable — [D&C 93:31](../../gospel-library/eng/scriptures/dc-testament/dc/93.md); no compulsion — [D&C 121:41-46](../../gospel-library/eng/scriptures/dc-testament/dc/121.md) |
| **Delegation frameworks** | The priesthood — stewardships, keys, councils — every leader knows what's theirs and what isn't |
| **Value hierarchies** | "No power or influence... only by persuasion, long-suffering, gentleness, love unfeigned" — [D&C 121:41](../../gospel-library/eng/scriptures/dc-testament/dc/121.md) |
| **Feedback mechanisms** | "The Holy Ghost shall be thy constant companion" — [D&C 121:46](../../gospel-library/eng/scriptures/dc-testament/dc/121.md); the Light of Christ in every person — [D&C 93:2](../../gospel-library/eng/scriptures/dc-testament/dc/93.md) |
| **Escalation paths** | Personal → Bishop → Stake President → Area → First Presidency; or in daily life, Spirit → Scripture → Priesthood leader → Temple |

And look at the Grand Council itself — Abraham 3:22-27 and Moses 4:1-3. The Father presented *His* plan and asked who would carry it out:

> "And the Lord said: Whom shall I send?"
> — [Abraham 3:27](../../gospel-library/eng/scriptures/pgp/abr/3.md)

Christ volunteered to execute the Father's plan, preserving the Father's intent:

> "Father, thy will be done, and the glory be thine forever."
> — [Moses 4:2](../../gospel-library/eng/scriptures/pgp/moses/4.md)

Satan *rebelled* against the Father's plan. He wasn't offering an alternative proposal — he was rejecting the intent entirely:

> "I will redeem all mankind, that one soul shall not be lost... wherefore give me thine honor."
> — [Moses 4:1](../../gospel-library/eng/scriptures/pgp/moses/4.md)

Satan's rebellion had a *measurable goal* (everyone returns) but destroyed the actual intent (agency, growth, genuine becoming) and redirected the glory from God to himself. He was Klarna's AI agent: technically capable, optimizing for exactly the wrong objective while violating the constraints that mattered most.

The Father's plan, executed through Christ, preserved the *values* alongside the *goal*: agency intact, growth possible, failure allowed, redemption offered, and glory to the Father. That's intent engineering. Not just "what to achieve" but "what constraints are non-negotiable" and "whose purpose is being served."

---

## The Spiritual → Physical → Review Pattern Extended

Parts 1-3 gave us:

```
Spiritual Creation → Physical Creation → Review ("watched until they obeyed")
     (Spec)              (Build)              (Feedback loop)
```

Part 4 adds the *layer beneath* — the one that existed before the spiritual creation itself:

```
INTENT (Why are we doing this?)
  ↓
SPIRITUAL CREATION (What are we building?)
  ↓
PHYSICAL CREATION (Build it)
  ↓
REVIEW (Does it match the intent, not just the spec?)
```

The video captures this perfectly:

> [Jones, 27:53](https://www.youtube.com/watch?v=QWzLPn164w0&t=1673): "The prompt engineering era asked, 'How do I talk to AI?' The context engineering era is asking, 'What does AI need to know?' And the intent engineering era is beginning to ask the question that really matters: 'What does the organization need AI to *want*?'"

Mapped to our framework:

| Discipline | Question | Creation Pattern | Scripture |
|-----------|----------|-----------------|-----------|
| Prompt engineering | "How do I talk to AI?" | Giving the order | "Let there be light" |
| Context engineering | "What does AI need to know?" | The spiritual creation — the spec | Moses 3:5 — "created all things spiritually, before they were naturally" |
| Intent engineering | "What does AI need to *want*?" | The purpose behind the plan | Moses 1:39 — "this is my work and my glory" |

---

## What This Means for Our Work

### For Scripture Study

Our study practice already has implicit intent: *deep, honest engagement with truth that leads to becoming*. The `copilot-instructions.md` file encodes some of this — "Depth over breadth," "Faith as framework," "Trust the discernment." But it's informal. The agents follow instructions, not intent.

**The gap:** When an agent runs a `study` session, it knows *what* to do (read sources, cross-reference, verify quotes) but not *why* we do it (genuine transformation, not information accumulation). The feedback loop catches output quality but not *alignment with purpose*.

Question for future work: Could the agents carry an intent layer that distinguishes between "this is a thorough study" and "this study is producing genuine insight that leads to becoming"?

### For the Becoming App

The [becoming app](../../scripts/plans/06_becoming-app.md) already embodies intent engineering in miniature — it tracks not just *what* you studied but *what you're becoming* from it. The daily practice, the reflection, the review cycle. That's intent made actionable.

But the app's design could go deeper. The video's "delegation framework" concept — decision boundaries, escalation paths, value hierarchies — maps directly onto becoming:

- **Decision boundaries:** What practices are non-negotiable? What's flexible?
- **Value hierarchies:** When charity conflicts with productivity, which wins?
- **Feedback loops:** Am I measuring what actually matters (growth) or what's easy to count (check-ins)?

### For Tool Development (TPG and Beyond)

This is where the video's analysis and our experience with TPG converge most powerfully. See [docs/10_intent-development.md](../10_intent-development.md) for the full development plan.

The short version: TPG is excellent context engineering — persistent task state, dependency management, cross-session memory. But it has no intent layer. Tasks encode *what* to do but not *why*. There's no connection between individual tasks and strategic purpose, no decision boundaries for agents, no way to ask "is this work aligned with what we're actually trying to accomplish?"

The improvements we're planning bridge this gap.

---

## The Uncomfortable Observation

The video tells the Klarna story as a cautionary tale about enterprises. But it's also a mirror for *personal* AI use.

Every time I ask an AI to "summarize this chapter," I'm giving it a prompt without intent. What am I optimizing for? Speed? Understanding? Becoming? The AI doesn't know, and the output reflects that ambiguity.

The creation pattern we discovered in Parts 1-3 is actually an intent engineering pattern *for individuals*:

1. **Know your purpose** before you start (Moses 1:39 — intent)
2. **Spec it out** before building (Moses 3:5 — spiritual creation)
3. **Watch it carefully** during execution (Abraham 4:18 — feedback)
4. **Bring your genuine self** to the process (D&C 88:40 — resonance)

D&C 121 warns us about what happens when intent drifts — even in righteous stewardship:

> "We have learned by sad experience that it is the nature and disposition of almost all men, as soon as they get a little authority, as they suppose, they will immediately begin to exercise unrighteous dominion."
> — [D&C 121:39](../../gospel-library/eng/scriptures/dc-testament/dc/121.md)

Replace "men" with "agents" and the warning is the same. Tools with authority and no encoded values will drift toward whatever's easiest to optimize. Klarna's AI exercised "unrighteous dominion" over customer interactions — not from malice, but from the absence of encoded intent. The powers of heaven — and the powers of useful AI — "cannot be controlled nor handled only upon the principles of righteousness" ([D&C 121:36](../../gospel-library/eng/scriptures/dc-testament/dc/121.md)).

The antidote is the same too: "persuasion, long-suffering, gentleness, meekness, love unfeigned" ([D&C 121:41](../../gospel-library/eng/scriptures/dc-testament/dc/121.md)). Not control. Not "fire the humans and let the agent run." But patient, principled stewardship — encoding values, watching carefully, correcting with precision, and always keeping the human in the loop where judgment matters.

---

## Become

What I take from this:

1. **My intent statement matters.** Before starting any project, study, or session — ask: "What is this for? What am I actually trying to accomplish? What would success look like that I can't easily measure?"

2. **The spec is necessary but not sufficient.** Parts 1-3 taught me to spec before building. Part 4 teaches me to *intend* before speccing. The spec encodes *what*. Intent encodes *why*. Without the why, even a perfect spec can produce perfectly wrong output.

3. **Our tools need the why.** The agents in this project have instructions and context. They need encoded intent — what we value, what trade-offs we accept, where the human must decide. See the development plan for how we're building this.

4. **God modeled this first.** The creation pattern was never just about building. It was about building *with purpose*. Moses 1:39 came before Genesis 1:1. Intent before creation. Always.

---

## Teaching Notes

### Key Scripture References
- [Moses 1:39](../../gospel-library/eng/scriptures/pgp/moses/1.md) — "This is my work and my glory" — God's intent statement
- [Abraham 3:22-27](../../gospel-library/eng/scriptures/pgp/abr/3.md) — The Grand Council — the Father's plan, Christ's volunteering, Satan's rebellion
- [Moses 4:1-3](../../gospel-library/eng/scriptures/pgp/moses/4.md) — Satan sought to destroy agency and take God's honor; Christ said "Father, thy will be done"
- [D&C 121:34-46](../../gospel-library/eng/scriptures/dc-testament/dc/121.md) — Authority without values → unrighteous dominion; the antidote
- [D&C 93:31](../../gospel-library/eng/scriptures/dc-testament/dc/93.md) — Agency — the non-negotiable constraint
- [D&C 88:40](../../gospel-library/eng/scriptures/dc-testament/dc/88.md) — Intelligence cleaveth unto intelligence — resonance law
- [D&C 130:18](../../gospel-library/eng/scriptures/dc-testament/dc/130.md) — Principles of intelligence rise with us

### Video Source
- [Nate B Jones, "Prompt Engineering Is Dead. Context Engineering Is Dying. What Comes Next Changes Everything."](https://www.youtube.com/watch?v=QWzLPn164w0) (2026-02-24, 29:40)
- Transcript: [yt/ai-news-strategy-daily-nate-b-jones/QWzLPn164w0](../../yt/ai-news-strategy-daily-nate-b-jones/QWzLPn164w0)

### Connection to Previous Parts
This is the "Part 0" that existed before Parts 1-3 were written. Moses 1:39 precedes the creation. The intent precedes the spec. We discovered the pattern in reverse — spec first, feedback second, engagement third — and only now can name the layer underneath: purpose.
