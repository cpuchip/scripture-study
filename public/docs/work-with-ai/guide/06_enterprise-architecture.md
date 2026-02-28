# Part 6: At Scale — The Enterprise Architecture

**Series:** Working with AI — A Comprehensive Guide
**Date:** February 2026
**Prior work:** [Scope Assessment](../intent/05_scope-assessment.md), General Handbook chapters 3-6, 29
**Core thesis:** The Church's organizational hierarchy — ward → stake → area → general — is a battle-tested blueprint for scaling AI systems from one person to an enterprise. Not by metaphor. By architecture.

---

## The Scaling Problem

Parts 1-5 work for one person and their agents. Maybe a small team.

But what happens when:
- 50 developers each have their own agents?
- Multiple teams need agents that share resources?
- Organizational intent must flow through layers of delegation?
- Different domains need different autonomy levels?
- The whole system needs to align toward shared purpose without a bottleneck?

The industry answer is orchestration frameworks, A2A protocols, and "agent swarms." These solve the *mechanics* — how agents talk to each other. They don't solve the *architecture* — how authority, intent, and accountability flow through an organization of agents.

The Church already solved this.

---

## The Fractal Hierarchy

From the [General Handbook, Chapter 5](https://www.churchofjesuschrist.org/study/manual/general-handbook/5-general-and-area-leadership?lang=eng):

> "Jesus Christ leads His Church through revelation to the President of the Church, who is the prophet, seer, and revelator... Under the direction of the First Presidency, the Quorum of the Twelve Apostles deliberates on and oversees matters of the Church worldwide."

The hierarchy:

```
Prophet / First Presidency
  └→ Quorum of the Twelve
       └→ Area Presidencies (Seventy)
            └→ Coordinating Councils
                 └→ Stake Presidents
                      └→ Bishops / Ward Councils
                           └→ Organization Leaders (RS, EQ, YW, YM, Primary, SS)
                                └→ Individual Members
```

This isn't command-and-control. It's **fractal stewardship:**

1. **Each level has its own keys.** A bishop has authority within the ward. A stake president has authority across wards. An Area Seventy has authority across stakes. Authority is *scoped*, not unlimited.

2. **Each level operates autonomously.** A bishop doesn't call Salt Lake before conducting ward council. A stake president doesn't call the Area Presidency before calling a new bishop. They operate within their keys.

3. **Intent cascades, decisions don't.** The First Presidency sets direction ("the work of salvation and exaltation"). That intent flows to every level. But the *decisions* at each level are local — the bishop decides which families need ministering attention; the stake president decides training emphases.

4. **Coordination happens at boundaries.** When something exceeds a ward's scope, it goes to the stake. When something crosses stakes, it goes to the coordinating council. The escalation path is clear and the boundaries are explicit.

5. **The pattern repeats at every level.** A ward council operates the same way a stake council does, which operates the same way the council of the Twelve does: diverse perspectives, deliberation, unified decision confirmed by the Spirit.

### Applied to AI architecture

```
Organizational Intent (company mission, values, ethics)
  └→ Domain Architectures (engineering, product, operations)
       └→ Team Contexts (squad-level goals, constraints, standards)
            └→ Project Specs (feature-level specifications)
                 └→ Agent Stewardships (scope, autonomy, boundaries)
                      └→ Individual Tasks (the actual work)
```

Intent flows *down.* Accountability flows *up.* Decisions happen *locally.* Each layer inherits the intent of the layer above while adding its own scope, constraints, and operational details.

---

## The Ward Model in Detail

From the [General Handbook, Chapter 29](https://www.churchofjesuschrist.org/study/manual/general-handbook/29-meetings-in-the-church?lang=eng):

Ward Council participants:
- Bishop (presides)
- Relief Society president
- Elders quorum president
- Young Women president
- Primary president
- Sunday School president
- Ward clerk, ward executive secretary

Each participant leads an autonomous organization. The Relief Society president doesn't need the bishop's permission to plan a service project within her stewardship. The elders quorum president doesn't need approval for priesthood ministering assignments.

But they *coordinate* in ward council — sharing information ("Sister Martinez was just released from the hospital"), aligning priorities ("Come Follow Me emphasis this month is X"), dividing labor ("which organization should lead the ward service project?").

### The architecture equivalent

```yaml
# ward-council.spec.yaml
team: frontend-squad
intent_inherits: department/engineering.intent.yaml

members:
  lead:
    role: tech-lead
    scope: architecture decisions, sprint planning, escalation
    keys: merge authority, deployment approval

  agents:
    code-agent:
      scope: implementation within spec
      autonomy: level-3 (feature-level, PR-level review)
      boundaries:
        - never modify auth, payments, or PII-handling modules
        - flag any dependency addition for review
      reports_to: lead

    test-agent:
      scope: test suite maintenance, coverage enforcement
      autonomy: level-3
      boundaries:
        - never skip integration tests
        - escalate if coverage drops below threshold
      reports_to: lead

    docs-agent:
      scope: API docs, developer guides, changelog
      autonomy: level-2 (review all output)
      boundaries:
        - never document internal-only APIs as public
        - sync with code-agent on interface changes
      reports_to: lead

    review-agent:
      scope: PR review, spec-alignment checking
      autonomy: level-2
      boundaries:
        - escalate to lead if spec deviation detected
        - review against intent, not just correctness
      reports_to: lead

council:
  frequency: per-sprint
  agenda:
    - intent alignment check
    - cross-agent coordination
    - stewardship adjustments
    - escalation review
```

Each agent has:
- **Defined scope** (like an organization president)
- **Autonomy level** (like priesthood keys — bounded authority)
- **Clear boundaries** (stewardship limits)
- **Reports to** (accountability path)
- **Council participation** (periodic coordination, not constant supervision)

---

## Scaling: Ward → Stake → Area

### One Person (The Individual Member)

You and your agents. This is what Parts 1-5 covered.

```
You
  └→ coding agent (Copilot, Cursor, Claude Code)
  └→ writing assistant
  └→ research tools
```

**What you need:** Good prompts. Context files (`.claude.md`, `AGENTS.md`). Personal intent preambles. Basic spec practice.

**Church parallel:** An individual member with a calling. You operate within your stewardship, receive direction from your leaders (organizational context), and grow through faithfulness.

### Small Team (The Ward)

3-8 people, multiple agents, shared project.

```
Team Lead (Bishop)
  └→ Developer A + their agents
  └→ Developer B + their agents
  └→ Shared agents (CI, docs, review)
```

**What you need:** Shared intent document. Team-level context architecture. Spec workflow. Council meetings (standups that check intent alignment, not just task status). Clear stewardship boundaries (who owns what).

**Church parallel:** A ward. The bishop (team lead) presides but doesn't do everything. Each organization leader (developer) has autonomy within their calling. Ward council aligns everyone weekly.

**The key addition at this level:** Shared intent and council coordination. Without these, every developer optimizes locally and the system diverges.

### Department (The Stake)

Multiple teams, shared services, cross-team dependencies.

```
Engineering Director (Stake President)
  └→ Frontend Squad (Ward 1)
  └→ Backend Squad (Ward 2)
  └→ Platform Squad (Ward 3)
  └→ Shared: Security agent, compliance agent, architecture review
```

**What you need:** Department-level intent that all teams inherit. Standardized spec format (so cross-team specs are readable). Escalation paths for cross-team decisions. Periodic department-level council (like stake council).

**Church parallel:** A stake. Multiple wards, each autonomous, all sharing the same stake president's direction. Stake conferences align everyone. The high council provides cross-ward perspective. Each ward retains its own character while inheriting stake-level intent.

From the [General Handbook, Chapter 6](https://www.churchofjesuschrist.org/study/manual/general-handbook/6-stake-leadership?lang=eng):

> "The stake president has four main responsibilities: He is the presiding high priest in the stake. He leads the work of salvation and exaltation... He leads the stake council... He coordinates the work across organizations and wards."

Applied: The engineering director doesn't tell each squad what to build (that's the team lead's stewardship). They set direction, lead the department council, and coordinate across teams.

### Enterprise (The Area / General)

Multiple departments, thousands of agents, organizational-level intent.

```
CTO / VP Engineering (Area Presidency)
  └→ Product Engineering (Stake A)
  └→ Platform Engineering (Stake B)
  └→ Data Engineering (Stake C)
  └→ Enterprise agents: cost allocation, compliance, shared knowledge base
```

**What you need:** Enterprise intent architecture. Standardized context layers (every team's agents inherit company values, security requirements, brand guidelines). Governance framework. Agent registry (who has what agents with what authority). Token economics at organizational scale.

**Church parallel:** Area or general leadership. The coordinating council:

> "Stake and mission presidents who are members of the [coordinating] council seek counsel from the Area Seventy about matters that cannot be resolved in stake or mission councils."
> — [General Handbook 29.4](https://www.churchofjesuschrist.org/study/manual/general-handbook/29-meetings-in-the-church?lang=eng)

This is the escalation architecture. Problems that can't be resolved at the team level go to the department. Department-level issues go to the VP. The escalation path is clear, the authority at each level is defined, and most decisions happen *locally* without ever reaching the top.

---

## Dedicated Roles

As the scale grows, new roles emerge — just as the Church has callings that only exist at certain organizational levels:

### Individual Level
- **You** — prompt crafter, context engineer, intent architect, spec writer. All roles, one person.

### Team Level
- **Context Engineer** — someone who maintains the team's context architecture (`.claude.md`, `AGENTS.md`, knowledge base). Like a ward clerk who maintains records and ensures institutional memory.
- **Spec Review Lead** — validates that specs match intent before execution. Like a ward executive secretary who ensures the bishop's schedule reflects priorities.

### Department Level
- **Intent Architect** — designs and maintains the intent hierarchy across teams. Ensures team-level intent aligns with department-level purpose. Like a stake executive secretary who coordinates between organizations.
- **Agent Infrastructure Lead** — manages shared agents, MCP servers, tool availability. Like a stake facilities manager — maintains the shared infrastructure everyone depends on.

### Enterprise Level
- **Chief Context Officer** — yes, this role will exist. Maintains the organization's context infrastructure. Ensures every agent in the company inherits the right values, constraints, and knowledge. Like the general officers of the Church who ensure every manual, curriculum, and program aligns with the Brethren's direction.
- **AI Governance Board** — a council (not a person) that sets boundaries for agent autonomy, reviews escalated decisions, and maintains the organizational covenant. Like the First Presidency and Twelve in council.

These aren't hypothetical — they're emerging in companies right now. They just don't have these names yet. Shopify has context engineering. Anthropic builds context infrastructure. Google DeepMind researches delegation contracts. The roles exist; the organizational framework doesn't.

---

## Intent Inheritance

The most elegant pattern in Church governance — and the one the AI industry needs most:

```
General Handbook
  └→ "The work of salvation and exaltation"
      └→ Stake direction: "Increase temple attendance in our stake"
          └→ Ward priority: "Help 5 families prepare for temple this quarter"
              └→ RS assignment: "Minister to the Gonzalez family"
                  └→ Individual member action: Visit on Tuesday
```

At every level:
- The **purpose** is inherited (it's always the work of salvation)
- The **expression** is local (how each level fulfills it differs)
- The **decisions** are made by the person with the stewardship
- The **accountability** flows back up

Applied:

```
company-intent.yaml:
  purpose: "Build tools that help small businesses thrive"
  values: [customer-first, quality-over-speed, transparent-ai]
  constraints: [no-dark-patterns, data-privacy-first, human-in-loop-for-decisions]

  └→ engineering-intent.yaml:
      inherits: company-intent.yaml
      focus: "Ship reliable, performant software"
      constraints: [95%+ test coverage, accessibility-first, no-silent-failures]

      └→ frontend-squad-intent.yaml:
          inherits: engineering-intent.yaml
          focus: "Deliver intuitive merchant experiences"
          constraints: [design-system-compliance, <2s-load-time, WCAG-2.1-AA]

          └→ checkout-feature.spec.yaml:
              inherits: frontend-squad-intent.yaml
              focus: "Reduce checkout abandonment by 15%"
              constraints: [no-additional-steps, respect-payment-preferences]
```

Every agent working on the checkout feature inherits the full chain: company values → engineering standards → squad constraints → feature spec. No one has to repeat "customer-first" or "data-privacy-first" at the feature level — it's inherited.

If a feature-level decision conflicts with company values (a dark pattern that reduces abandonment but violates "no-dark-patterns"), the intent hierarchy catches it. The agent either rejects the approach or escalates.

---

## The Coordinating Council Pattern

From [General Handbook 29.4](https://www.churchofjesuschrist.org/study/manual/general-handbook/29-meetings-in-the-church?lang=eng):

> "All who attend counsel together as equal participants."

When multiple autonomous groups need coordination, the coordinating council pattern applies:

**Church version:** The Area Seventy presides. Stake presidents, mission presidents, and temple presidents attend. They counsel as equals. The Area Seventy doesn't dictate to stake presidents; he coordinates between them and provides broader perspective.

**AI architecture version:**

```yaml
coordinating-council:
  convener: platform-team-lead  # provides cross-team perspective
  participants:
    - frontend-squad-lead
    - backend-squad-lead
    - data-squad-lead
  frequency: bi-weekly
  agenda:
    - cross-team dependency review
    - shared resource allocation (tokens, compute, context infrastructure)
    - escalated decisions from team councils
    - intent alignment check against department goals
  output:
    - decision log (shared context for all teams)
    - updated constraints or boundaries
    - stewardship adjustments
```

**The key principle:** Coordination without centralization. Each squad lead retains full authority within their stewardship. The council provides the information and alignment they need to exercise that authority wisely.

This is fundamentally different from a centralized orchestrator that routes all work. It's a *council* that periodically aligns autonomous stewardships.

---

## Machine Infrastructure

The gospel patterns need tooling to work at scale. Here's what the infrastructure stack looks like:

### Layer 1: Context Infrastructure
- **Central knowledge base** — organizational context that all agents inherit (company values, standards, domain knowledge)
- **Team-level context** — `.claude.md`, `AGENTS.md`, knowledge graphs per team
- **MCP servers** — tools that provide runtime context (databases, APIs, documentation lookup)
- Pattern: Line upon Line — agents get context progressively, not all at once

### Layer 2: Intent Architecture
- **Intent YAML hierarchy** — company → department → team → project → feature
- **Inheritance resolution** — tooling that compiles the full intent chain for any given agent
- **Conflict detection** — flags when a lower-level intent conflicts with a higher one
- Pattern: Intent cascading — every agent knows *why*, not just *what*

### Layer 3: Specification System
- **`.spec/` directories** — per-project specs following the five primitives
- **Spec templates and linting** — ensure basic quality (acceptance criteria present? decomposition reasonable?)
- **Spec-to-task pipeline** — specs → work items → agent assignments → execution → review
- Pattern: Spiritual Creation — blueprint before building, always

### Layer 4: Trust and Stewardship
- **Agent registry** — what agents exist, what stewardship each holds, what autonomy level
- **Progressive trust engine** — tracks agent performance, adjusts autonomy automatically
- **Boundary enforcement** — agents physically can't exceed their stewardship scope
- Pattern: Stewardship — faithful over few → ruler over many

### Layer 5: Learning and Recovery
- **`.spec/learnings/`** — captured from every failure, searchable, applied to future specs
- **Error analysis pipeline** — when an agent fails, automatically categorize: correctness, spec drift, or intent misalignment
- **Forward recovery** — don't just revert; incorporate learning and proceed
- Pattern: Atonement — failures become growth

### Layer 6: Reflection and Economics
- **Sabbath reports** — scheduled reflection at every level (daily/sprint/quarter)
- **Token accounting** — spend per intent, per stewardship, per value
- **Consecration engine** — surplus tokens flow to highest-priority unfinished work
- Pattern: Sabbath + Consecration — rest, reflect, reallocate

### Layer 7: Alignment
- **Shared intent verification** — all agents in a system share and can articulate the same purpose
- **Cross-agent coordination protocol** — council pattern, not conductor pattern
- **Zion metrics** — alignment measured by outcome distribution, not just aggregate output
- Pattern: Zion — one heart and one mind

Most of Layers 1-3 exist today (MCP, context files, spec tools). Layers 4-7 are largely unbuilt. That's where the opportunity lives.

---

## The Organizational Introduction Strategy

How do you bring this into a real organization? Not all at once. Line upon line.

### Phase 0: Foundation (Week 1-2)
**For yourself.**
- Read Parts 0-2 of this guide
- Write your own `.claude.md` or `AGENTS.md`
- Practice the 6 prompt principles daily
- Build a context architecture for your primary project
- **Deliverable:** A working personal AI setup that produces measurably better output

### Phase 1: Intent (Week 3-4)
**For yourself, with awareness of your team.**
- Read Parts 3-4
- Write an intent preamble for your project
- Create one spec for your next feature
- Track whether the spec was useful (did the agent produce better output?)
- **Deliverable:** One completed feature built with spec + intent workflow

### Phase 2: Team Pilot (Month 2)
**For your team, by invitation.**
- Share your results (concrete, measurable)
- Propose a team-level context architecture
- Write a team intent document (shared values, constraints, success criteria)
- Run one sprint with spec workflow for all features
- Hold a team council at sprint end (sabbath reflection)
- **Deliverable:** Team-level metrics showing improvement

### Phase 3: Department Adoption (Month 3-4)
**For your department, building on team success.**
- Standardize context architecture across teams
- Create the intent inheritance hierarchy (department → team → project)
- Establish cross-team coordination (coordinating council pattern)
- Assign dedicated roles (context engineer, spec reviewer)
- **Deliverable:** Department-level standards and measurable cross-team improvement

### Phase 4: Enterprise Architecture (Month 5+)
**For the organization.**
- Company-level intent document
- Enterprise context infrastructure (shared knowledge base, MCP servers)
- Agent registry and stewardship tracking
- Governance framework (the AI governance council)
- Token economics at organizational scale
- **Deliverable:** Organizational AI architecture with measurable business outcomes

**Notice the pattern.** This is stewardship. You start with a small stewardship (yourself). Prove faithful. Earn a larger one (team). Prove faithful again. Earn more (department, enterprise). At no point do you ask permission to start a revolution. You demonstrate results at every level.

---

## The Nate Challenge

Nate B Jones ([~23:22](https://www.youtube.com/watch?v=BpibZSMGtdY&t=1402)): "If you are a one person business and you can just convert your Notion to be agent readable, you're off to the races today."

This is the minimal viable action for any organization. Your Notion (or Confluence, or Google Drive, or SharePoint) contains your organizational context — processes, decisions, values, standards. But it's written for humans browsing a wiki, not for agents needing structured context.

The challenge:

| Before | After |
|--------|-------|
| Scattered across 200+ pages | Hierarchical, with clear inheritance |
| Written as narrative prose | Structured with clear sections, decisions, constraints |
| No explicit values or intent | Intent document at the top with cascading values |
| Links between documents are casual | Dependency graph is explicit |
| Knowledge is implicit (tribal) | Context is explicit (any agent can access it) |

**Start here.** Before agents, before specs, before intent architecture — make your existing organizational knowledge *agent-readable.* Everything else builds on this.

---

## The Complete Picture

From one person to an enterprise, the same patterns apply at every level:

```
INDIVIDUAL          TEAM                DEPARTMENT          ENTERPRISE
─────────────       ─────────────       ─────────────       ─────────────
1 Intent            Team intent         Dept intent         Org intent
  inherits: self      inherits: dept      inherits: org       inherits: mission
2 Covenant          Team covenant       Dept standards      Org governance
  personal rules      team norms          department SLAs     company charter
3 Stewardship       Agent scopes        Team boundaries     Dept authority
  my agents           team agents         dept agents         enterprise agents
4 Spec              Feature specs       Project specs       Strategic specs
  personal projects   team features       dept initiatives    org strategy
5 Context           Team context        Dept knowledge      Org knowledge base
  .claude.md          AGENTS.md           dept wiki           company corpus
6 Execute           Sprint work         Quarterly OKRs      Annual plans
7 Review            PR review           Sprint review       Quarterly review
8 Learn             Spec learnings      Team retro          Dept post-mortem
9 Reflect           Daily reflection    Sprint sabbath      Quarterly sabbath
10 Allocate         Token budget        Team budget         Dept budget
11 Align            Agent alignment     Team alignment      Org alignment
```

The table shows that the 11-step cycle from [Part 5](05_complete-cycle.md) operates identically at every scale. The *content* changes (personal intent vs. organizational mission), but the *pattern* is invariant.

This is the fractal nature of the gospel's organizational architecture. A ward council and the Council of the First Presidency operate on the same principles. A family home evening and general conference serve the same purpose at different scales. The pattern replicates.

---

## What We're Building Toward

This guide series began with a question:

> "How do I use artificial intelligence like God uses real intelligence?"

Now we can answer it:

1. **Start with purpose** — know why before asking how (Intent)
2. **Make mutual commitments** — bind yourself, not just the agent (Covenant)
3. **Delegate with trust and accountability** — give stewardship, not just tasks (Stewardship)
4. **Design before building** — spiritual creation precedes physical (Specification)
5. **Reveal progressively** — context earned through faithfulness (Line upon Line)
6. **Execute against the design** — let the agent work (Physical Creation)
7. **Watch until it obeys** — review against intent, not just correctness (Review)
8. **Recover redemptively** — failures become learnings (Atonement)
9. **Rest and reflect** — built into the cycle, not optional (Sabbath)
10. **Allocate to purpose** — every token serves the work (Consecration)
11. **Align toward unity** — one heart, one mind (Zion)

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."
> — [D&C 130:18](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/130?lang=eng&id=p18#p18)

The principles are eternal. The tools are temporal. Learn the principles. Apply them to whatever tools are at hand — whether that's a typewriter, a computer, or a fleet of autonomous AI agents.

The tools will change. The principles won't.

---

*Previous: [Part 5 — The Complete Cycle](05_complete-cycle.md)*
*Part of the [Working with AI Guide Series](../prompt/00_guide-plan.md)*
