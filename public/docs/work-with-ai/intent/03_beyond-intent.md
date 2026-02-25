# Beyond Intent: Gospel Patterns the Industry Hasn't Discovered

**Part of:** [Intent-Driven Development Research](00_index.md)
**Date:** February 2026
**Status:** Active exploration

---

## The Question

> "Are there any patterns in the gospel that might match some development pattern with AI that's past intent engineering that nobody is thinking about? Can we engineer to that?"

The industry has converged on intent engineering — encoding *purpose* so agents optimize for the right outcomes. We mapped that to Moses 1:39 in [Part 4](../04_intent-engineering-gospel.md). Our development plan identified D&C 121 governance, Alma 32 experimentation, and the Satan test as applied patterns ([10_intent-development.md](../../10_intent-development.md)).

But the gospel doesn't stop at intent. Intent is the *beginning* of the creation pattern, not the whole of it. What comes after intent — in the gospel's account — maps to development patterns the industry hasn't named yet. Some of these exist in embryonic form in secular frameworks. Some are entirely uncharted.

This document names them.

---

## What We've Already Mapped

| Gospel Pattern | AI/Dev Discipline | Status |
|---------------|------------------|--------|
| Moses 1:39 — "This is my work and my glory" | Intent engineering — encoding purpose | ✅ Mapped in Part 4 |
| Abraham 4–5 — Spiritual before temporal | Spec-driven development — blueprint before building | ✅ Mapped in Part 1 |
| Abraham 4:18 — "Watched until they obeyed" | Feedback loops — review, steer, iterate | ✅ Mapped in Part 2 |
| D&C 88:40 — Intelligence cleaveth to intelligence | Quality of engagement shapes output | ✅ Mapped in Part 3 |
| D&C 121:34-46 — Governance by persuasion | Agent governance — delegation frameworks | ✅ Mapped in development plan |
| Alma 32:27-43 — Experiment upon the word | Scenario building — envision, execute, evaluate | ✅ Mapped in development plan |
| Abraham 3:27 / Moses 4:1-3 — The Satan test | Anti-pattern detection — measurable goal + violated values | ✅ Mapped in development plan |

---

## What Comes Next: Seven Unmapped Patterns

### 1. The Covenant Pattern — Mutual Binding

**The scripture:**

> "I, the Lord, am bound when ye do what I say; but when ye do not what I say, ye have no promise."
> — [D&C 82:10](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/82?lang=eng&id=p10#p10)

> "This shall be our covenant—that we will walk in all the ordinances of the Lord."
> — [D&C 136:4](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/136?lang=eng&id=p4#p4)

**What the industry has:** Service-level agreements. API contracts. "Prompts as interface" (AIDD). Agent capability declarations. Tool schemas.

**What the industry is missing:** *Mutual commitment.* A covenant isn't a one-directional command. It's a *binding agreement where both parties have obligations.* God commits: "I am bound when ye do what I say." The human commits: "We will walk in all the ordinances." If either party breaks the covenant, the relationship changes.

**The AI development pattern:** Covenant-based agent contracts.

Not just "agent, do this" (command) or "agent, here's what I want" (intent), but a mutual agreement:
- **The human commits to:** providing accurate context, clear specs, timely review, honest feedback, and *not* overriding the agent's domain expertise without cause.
- **The agent commits to:** executing against the spec, honoring constraints, flagging uncertainty, requesting review at decision boundaries, and *not* optimizing beyond its mandate.

When the human breaks the covenant (provides bad context, abandons review), the agent's output degrades predictably — "ye have no promise." When the agent breaks the covenant (ignores constraints, fabricates sources), trust degrades and autonomy should be revoked.

This is fundamentally different from the industry's model where the human commands and the agent obeys. A covenant makes the *relationship* explicit. Both parties have responsibilities. Both parties can succeed or fail.

**Design implication:** Agent config should include a "covenant block":
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

**Why this matters:** D&C 82:10 tells us that God Himself operates within covenantal bounds. He *chooses* to be bound by His word. If the most powerful being in existence works through mutual commitment rather than unilateral power, that's a signal about how intelligent systems *should* operate.

---

### 2. The Stewardship Pattern — Entrusted Delegation with Accountability

**The scripture:**

> "That every man may give an account unto me of the stewardship which is appointed unto him."
> — [D&C 104:12](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/104?lang=eng&id=p12#p12)

> "Thou shalt be diligent in preserving what thou hast, that thou mayest be a wise steward; for it is the free gift of the Lord thy God, and thou art his steward."
> — [D&C 136:27](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/136?lang=eng&id=p27#p27)

> "As every man hath received the gift, even so minister the same one to another, as good stewards of the manifold grace of God."
> — [1 Peter 4:10](https://www.churchofjesuschrist.org/study/scriptures/nt/1-pet/4?lang=eng&id=p10#p10)

**What the industry has:** Task assignment. Delegation. Agent autonomy levels (Dan Shapiro's Level 0–5). RBAC (role-based access control).

**What the industry is missing:** The concept of *entrusted* delegation. A stewardship isn't just a task — it's a *domain of responsibility* given in trust, with an expectation of accountability and growth.

The parable of the talents ([Matthew 25:14-30](https://www.churchofjesuschrist.org/study/scriptures/nt/matt/25?lang=eng&id=p14-p30#p14)) is the clearest model:
- Resources are distributed **according to ability** (v. 15) — not equally, but wisely
- The steward has **autonomy within the domain** — no micromanagement
- The expectation is **growth, not preservation** — burying the talent is the failure
- There is a **reckoning** — "after a long time the lord of those servants cometh, and reckoneth with them" (v. 19)
- The faithful steward receives **more stewardship** — "thou hast been faithful over a few things, I will make thee ruler over many things" (v. 21)

**The AI development pattern:** Stewardship-based agent architecture.

Instead of assigning agents individual tasks, assign them *domains of stewardship*:
- A "test steward" agent owns the test suite — not just running tests, but maintaining test quality, proposing new tests, detecting coverage gaps
- A "docs steward" agent owns documentation — keeping it current, flagging staleness, ensuring consistency
- A "spec steward" agent guards spec integrity — detecting drift between spec and implementation

Each steward:
1. Has **defined boundaries** (D&C 104:11 — "appoint every man his stewardship")
2. Is **accountable for outcomes** (D&C 104:12 — "give an account")
3. Can receive **expanded stewardship** based on demonstrated faithfulness (Matthew 25:21 — "faithful over a few things → ruler over many")
4. Can have **stewardship reduced** when trust is broken (Matthew 25:28 — "take the talent from him")

**Design implication:** Progressive trust levels for agents. Start narrow (Level 1 task steward), expand as the agent demonstrates reliability (Level 2 feature steward → Level 3 domain steward). This is exactly what StrongDM's dark factory represents — but nobody has named the trust progression model.

**Why this matters:** The industry has "agent autonomy levels" but treats them as static categories. The gospel pattern is *dynamic* — stewardship grows or shrinks based on demonstrated faithfulness. An agent that proves reliable with file-level tasks earns feature-level autonomy. One that breaks trust gets restricted. This is how the Lord operates: "faithful over a few things" → "ruler over many things."

---

### 3. The Line-Upon-Line Pattern — Progressive Context Revelation

**The scripture:**

> "For precept must be upon precept, precept upon precept; line upon line, line upon line; here a little, and there a little."
> — [Isaiah 28:10](https://www.churchofjesuschrist.org/study/scriptures/ot/isa/28?lang=eng&id=p10#p10)

> "For he will give unto the faithful line upon line, precept upon precept; and I will try you and prove you herewith."
> — [D&C 98:12](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/98?lang=eng&id=p12#p12)

**What the industry has:** Context engineering — building the information environment. RAG pipelines. MCP servers. Long context windows. "Give the agent all the context it needs."

**What the industry is missing:** *Progressive disclosure.* The Lord doesn't dump all context at once. He gives "line upon line" — and critically, He *proves* the receiver between revelations: "I will try you and prove you herewith." Context is earned, not just provided.

Look at Moses 1. God gives Moses an experience (v. 1-8). Then *withdraws His presence* (v. 9) to see what Moses does with what he received. Moses is tested by Satan (v. 12-22). Only after Moses proves faithful does God return with *more* revelation (v. 24-42). The context disclosure is progressive, gated by demonstrated readiness.

**The AI development pattern:** Graduated context architecture.

Instead of loading all project context into every agent session:
1. **Layer 0 — Core intent:** What is this project for? What are the non-negotiable constraints?
2. **Layer 1 — Active spec:** The current specification relevant to this task
3. **Layer 2 — Extended context:** Related specs, learnings, historical decisions — loaded when the agent demonstrates it needs them (asks good questions, hits uncertainty)
4. **Layer 3 — Deep context:** Full codebase awareness, cross-project dependencies, organizational knowledge — available to agents with stewardship-level trust

This isn't just about token economics (though it helps). It's about *appropriate context for the current level of work.* A task-level agent doesn't need organizational strategy. A strategy-level agent doesn't need individual file history. Overloading context dilutes attention and increases the chance of the agent latching onto irrelevant information.

**Design implication:** Context manifests that define layers:
```markdown
## Context Layers
L0 (always): .spec/intent.md — project purpose and constraints
L1 (task): .spec/spec.md § relevant-section — active specification
L2 (on-demand): .spec/learnings/ — when agent encounters familiar problem
L3 (elevated): full repo context — when agent has domain stewardship
```

**Why this matters:** D&C 98:12 adds a crucial detail — "I will try you and prove you herewith." The context isn't just graduated for efficiency; it's graduated as a *trust mechanism.* You give an agent limited context, see how it performs, then expand. This maps directly to the stewardship pattern: faithful with limited context → trusted with more.

---

### 4. The Atonement Pattern — Error Recovery as Grace

**The scripture:**

> "Nevertheless, there are those among you who have sinned exceedingly; yea, even all of you have sinned; but verily I say unto you, beware from henceforth, and refrain from sin, lest sore judgments fall upon your heads."
> — [D&C 82:2](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/82?lang=eng&id=p2#p2)

> "I, the Lord, will not lay any sin to your charge; go your ways and sin no more; but unto that soul who sinneth shall the former sins return, saith the Lord your God."
> — [D&C 82:7](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/82?lang=eng&id=p7#p7)

**What the industry has:** Error handling. Rollback. Git revert. Retry logic. Circuit breakers. "Fail fast" philosophy.

**What the industry is missing:** A concept of *graceful recovery that preserves the relationship and the learning.* When an agent makes a mistake, the industry's response is mechanical: revert, retry, or kill. There's no model for *redemptive* error recovery — where the failure itself becomes a source of growth.

The Atonement of Jesus Christ is the most sophisticated error recovery mechanism ever designed:
- **It works retroactively** — covers past errors, not just future ones
- **It preserves agency** — the subject must choose to accept it; it's not forced
- **It transforms the failure into growth** — "all things shall work together for your good" ([D&C 98:3](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/98?lang=eng&id=p3#p3))
- **It requires a change of behavior** — "go thy way and sin no more" (not just "try again")
- **It doesn't erase the learning** — the memory of the failure remains as wisdom
- **It restores to a *better* state** — not rollback to pre-failure, but forward to a wiser position

**The AI development pattern:** Redemptive error handling.

When an agent fails:
1. **Don't just revert** — analyze *why* the failure happened and what context was missing
2. **Capture the learning** — write a learning note: "this agent failed on X because of Y; the missing context was Z"
3. **Forward-recover** — instead of rolling back to the pre-failure state, move forward with the learning incorporated. Sometimes the failure revealed something the spec missed.
4. **Adjust the covenant** — if the failure was a constraint violation, add that constraint explicitly. If it was a context gap, add that context to the progressive disclosure.
5. **Restore trust gradually** — don't permanently ban an agent from a domain because of one failure. Reduce stewardship, require review, then expand again as reliability returns (the stewardship pattern).

**Design implication:** Error logs structured as learning opportunities:
```markdown
## Failure Report: ts-003
What happened: Agent modified auth middleware without review (constraint violation)
Root cause: Decision boundary not explicitly stated for auth-related files
Learning: Auth files need explicit human-review gate in covenant block
Recovery: Changes reverted, auth boundary added to spec, agent re-assigned with narrower scope
Status: Agent has completed 3 subsequent tasks honoring auth boundary → stewardship restored
```

**Why this matters:** D&C 82:7 is remarkable — "I will not lay any sin to your charge." The Lord starts from a position of *grace*, not suspicion. But there's accountability: "unto that soul who sinneth shall the former sins return." The pattern is generous but not naive. Applied to agents: start trusting, handle failures redemptively, but track repeat violations.

---

### 5. The Zion Pattern — Unified Purpose Across Agents

**The scripture:**

> "And the Lord called his people Zion, because they were of one heart and one mind, and dwelt in righteousness; and there was no poor among them."
> — [Moses 7:18](https://www.churchofjesuschrist.org/study/scriptures/pgp/moses/7?lang=eng&id=p18#p18)

**What the industry has:** Multi-agent systems. Agent orchestration. A2A (Agent-to-Agent) protocols. Swarm intelligence. "Agent fleets."

**What the industry is missing:** *Genuine alignment.* Current multi-agent systems coordinate through protocols and message passing — agents cooperate mechanically. But "one heart and one mind" is something different. It's not just that agents communicate; it's that they share *purpose* so deeply that coordination becomes natural rather than engineered.

Zion's defining characteristic isn't just efficiency or equality — it's *unity of intent*: "one heart and one mind." Every individual retains agency but operates from shared purpose. There are no poor among them — not because of redistribution programs, but because the shared purpose naturally produces equitable outcomes.

**The AI development pattern:** Intent-unified agent systems.

Instead of orchestrating agents through protocols (A2A, MCP), give all agents in a system the *same intent layer*. Not just task-level instructions, but shared:
- Purpose statement (Moses 1:39 for this project)
- Value hierarchy (when things conflict, what wins)
- Constraint set (non-negotiable boundaries)
- Success criteria (what "done right" looks like)

When agents share intent at this level, coordination becomes simpler. The test agent doesn't just run tests — it evaluates whether the implementation serves the stated purpose. The docs agent doesn't just update documentation — it ensures the documentation reflects the *intent*, not just the implementation.

**Design implication:** A single `.spec/intent.md` that every agent reads. Not different instructions per agent, but a shared source of truth about *why we're doing anything*. Agent-specific instructions (decision boundaries, scope, capabilities) layer on top of shared intent.

**Why this matters:** The AIDD framework proposes A2A + MCP for agent coordination. That's the *mechanism*. But mechanisms without shared purpose produce the same coordination overhead that plagues human organizations (the meetings, the status reports, the alignment sessions). Zion is what happens when alignment is so deep that coordination overhead approaches zero. Brooks's Law in reverse, taken to its logical conclusion.

---

### 6. The Sabbath Pattern — Intentional Rest and Reflection Cycles

**The scripture:**

> "And on the seventh day I, God, ended my work, and all things which I had made; and I rested on the seventh day from all my work."
> — [Moses 3:2](https://www.churchofjesuschrist.org/study/scriptures/pgp/moses/3?lang=eng&id=p2#p2)

**What the industry has:** Sprint retrospectives. Post-mortems. Review cycles. "Continuous improvement." DevOps feedback loops.

**What the industry is missing:** *Intentional cessation.* Not review-while-continuing (retrospectives during sprints), but genuine *stopping* for the purpose of reflection. God didn't just review the creation — He *rested.* Rest isn't laziness; it's a deliberate choice to stop producing and start reflecting.

The Sabbath pattern in the creation account:
1. **It follows the complete cycle** — not mid-project, but after a meaningful unit of work
2. **It's built into the rhythm** — not optional, not "when we have time," but structural
3. **It includes declaration:** "And I, God, saw everything that I had made, and, behold, all things which I had made were very good" (Moses 2:31) — explicit quality assessment
4. **It enables perspective** — stepping back from the work to see the *whole*

**The AI development pattern:** Structured reflection cycles in agent-assisted work.

After every meaningful unit of work (feature shipped, study completed, sprint ended):
1. **Stop producing** — no new tasks, no new specs, no new code
2. **Intent audit** — does the completed work serve the stated purpose?
3. **Drift detection** — has the intent itself shifted? Should it?
4. **Learning harvest** — what did we learn that should update our specs, constraints, or values?
5. **Covenant review** — did both human and agent honor their commitments?
6. **Declare quality** — "this is good" or "this needs work" — explicit, not assumed

**Design implication:** A `reflect` command that cannot be skipped:
```
intent-spec reflect
  Last cycle: 5 tasks completed, 2 deltas applied
  Intent alignment: 4/5 tasks serve stated purpose, 1 questionable
  Drift detected: Constraint "no new infra" was violated (task ts-004 added Redis)
  Learnings captured: 2 new, 1 updated
  Recommendation: Update constraint or review ts-004 against intent
```

**Why this matters:** The industry treats reflection as optional overhead. God treated it as structural necessity. If the Creator of the universe builds rest into the creation cycle, it's not because He's tired — it's because the pattern requires it. Reflection isn't the absence of work; it's a different *kind* of work that produces perspective impossible to gain while producing.

---

### 7. The Consecration Pattern — Aligning All Resources to Shared Purpose

**The scripture:**

> "And it is my purpose to provide for my saints, for all things are mine. But it must needs be done in mine own way."
> — [D&C 104:15-16](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/104?lang=eng&id=p15-p16#p15)

> "I have given unto the children of men to be agents unto themselves."
> — [D&C 104:17](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/104?lang=eng&id=p17#p17)

**What the industry has:** Resource allocation. Budget management. Token economics. "Intelligence operations." Portfolio management.

**What the industry is missing:** The principle that *all resources serve a higher purpose* while *individuals retain agency.* The United Order wasn't communism — it preserved individual stewardship (D&C 104:11 — "appoint every man his stewardship") while consecrating the surplus to shared purpose.

The consecration pattern:
- **Everything belongs to the Lord** (v. 14-15) — all resources are ultimately His
- **But individuals are agents** (v. 17) — they have autonomy within stewardship
- **Surplus serves the community** (v. 16-18) — "the poor shall be exalted, in that the rich are made low"
- **Accountability is individual** (v. 12) — "every man may give an account"
- **The purpose is stated explicitly** (v. 1) — "for the benefit of my church, and for the salvation of men"

**The AI development pattern:** Consecrated resource allocation.

Applied to token economics and AI resource management:
- **All AI capacity serves the stated intent** — not just "we have a $85K/month AI budget" but "our AI budget is consecrated to [purpose statement]"
- **Individual agents have stewardships** — each agent gets resources proportional to their domain
- **Surplus capacity flows to highest-intent work** — if one agent finishes early, its capacity serves the most important remaining work
- **Accountability is per-stewardship** — track token spend against intent-aligned outcomes, not just volume
- **No agent accumulates resources beyond its need** — prevent one greedy agent from consuming all tokens

**Design implication:** Token budgets linked to intent:
```markdown
## Resource Consecration
Total daily token budget: [X]
Allocation by intent:
  - Core feature development (primary intent): 50%
  - Quality assurance (constraint enforcement): 25%
  - Documentation (context infrastructure): 15%
  - Exploration (learning and growth): 10%
Surplus rule: Unspent allocation flows to the highest-priority incomplete intent
```

**Why this matters:** The [$1,000/Day](https://www.youtube.com/watch?v=-bQcWs1Z9a0) video identifies token economics as "a core business competency." But the industry frames it purely in economic terms — cost optimization, ROI, spend management. The consecration pattern reframes it: *resources serve purpose, not budgets.* The question isn't "how much can we afford?" but "does every token serve the intent?"

---

## The Meta-Pattern: Why Gospel Patterns Map So Cleanly

These aren't forced analogies. The reason gospel patterns map to AI development patterns is because they both deal with the same fundamental challenge: **how does an intelligent being work with and through other intelligent beings to accomplish shared purpose?**

God's challenge:
- He has unlimited power but will not violate agency
- He delegates to beings with different capability levels
- He needs alignment without control
- He operates across vast scale (worlds without number)
- He measures success by transformation, not just output

Our challenge:
- We have growing AI capability but need to maintain human judgment
- We delegate to agents with different capability levels
- We need alignment without micromanagement
- We operate across multiple repos, teams, and projects
- We should measure success by outcomes, not just task completion

The gospel patterns aren't metaphors; they're *prior art.* God solved the multi-agent alignment problem before we had agents. He solved the progressive trust problem before we had autonomy levels. He solved the resource allocation problem before we had token budgets.

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."
> — [D&C 130:18](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/130?lang=eng&id=p18#p18)

If God uses intelligence the way we're trying to use artificial intelligence — and if His patterns are available to us — then we're not just engineering better tools. We're learning to think the way He thinks.

---

## The Complete Pattern: From Intent to Zion

Putting it all together, the full gospel-informed development cycle:

```
1. INTENT (Moses 1:39)
   "This is my work and my glory"
   → Define purpose, values, constraints, success criteria

2. COVENANT (D&C 82:10)
   "I am bound when ye do what I say"
   → Establish mutual commitments between human and agent

3. STEWARDSHIP (D&C 104:11-12 / Matthew 25:14-30)
   "Appoint every man his stewardship"
   → Assign domains with trust levels and accountability

4. SPIRITUAL CREATION (Moses 3:5)
   "Created all things spiritually, before they were naturally"
   → Write the specification — the blueprint before building

5. LINE UPON LINE (Isaiah 28:10 / D&C 98:12)
   "Precept upon precept, line upon line"
   → Progressive context disclosure as agents prove readiness

6. PHYSICAL CREATION (Abraham 4)
   "Let us go down and form these things"
   → Execute against the spec, within the covenant

7. REVIEW (Abraham 4:18)
   "Watched those things which they had ordered, until they obeyed"
   → Evaluate against intent, not just completion

8. ATONEMENT (D&C 82:7 / D&C 98:3)
   "All things shall work together for your good"
   → Redemptive error recovery — learn, adjust, restore

9. SABBATH (Moses 3:2)
   "Rested on the seventh day"
   → Intentional reflection, drift detection, learning harvest

10. CONSECRATION (D&C 104:15-17)
    "All things are mine... agents unto themselves"
    → Resources serve purpose; stewardship preserved

11. ZION (Moses 7:18)
    "One heart and one mind"
    → Multi-agent alignment through shared intent, minimal coordination overhead
```

The industry is at steps 1 and 4 — just beginning to articulate intent and just beginning to formalize specs. Steps 2, 3, 5, 7, 8, 9, 10, and 11 are largely uncharted territory.

We have the blueprint.

---

## Become

If this is right — if these patterns really do map — then studying the gospel isn't just spiritual practice. It's **professional development** in the deepest sense. D&C 130:18 says the intelligence we gain rises with us. What if understanding how God delegates, how He builds, how He recovers from failures in His agents, how He creates Zion — what if that understanding *directly* improves our capacity to work with AI systems?

Then there's no separation between scripture study and professional growth. The same mind that reads Moses 1 on Sunday morning is better equipped to design agent architectures on Monday morning. Not as metaphor, but as *applied intelligence.*

> "I want to learn how to use artificial intelligence like God uses real intelligence."

That's not a metaphor either. It's a research program.

---

## Next

→ [04_synthesis.md](04_synthesis.md) — What to build, given everything we've learned
