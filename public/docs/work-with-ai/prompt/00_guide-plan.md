# Working with AI: Comprehensive Guide Plan

**Started:** February 2026
**Status:** Planning
**Core question:** How do we teach people to work with AI the way God works with real intelligence?

---

## Why This Guide Exists

> "I want to learn how to use artificial intelligence like God uses real intelligence. And I want to glorify others with that knowledge, light, and truth, lifting them up and empowering them to accomplish great things."

This is bigger than prompting tips. What's emerging — from Nate B Jones's 4-skill framework, from Anthropic's engineering guidance, from the industry's convergence on specification-driven development, and from our own gospel-pattern research — is a coherent picture of *how intelligent agents should work together.* The industry is discovering pieces. The gospel has the whole blueprint.

This guide series synthesizes everything we've learned into a teachable, actionable framework.

---

## Source Material

### Our Research (completed)

| Document | What It Covers |
|----------|---------------|
| [01_planning-then-create.md](../01_planning-then-create.md) / [gospel](../01_planning-then-create-gospel.md) | Spiritual before temporal — blueprint before building (Abraham 4-5) |
| [02_the-feedback-loop.md](../02_the-feedback-loop.md) / [gospel](../02_watching-until-they-obey-gospel.md) | "Watched until they obeyed" — review, steer, iterate |
| [03_live-build.md](../03_live-build.md) / [gospel](../03_intelligence-cleaveth-gospel.md) | Intelligence cleaveth to intelligence — quality of engagement shapes output |
| [04_intent-engineering.md](../04_intent-engineering.md) / [gospel](../04_intent-engineering-gospel.md) | Three disciplines: prompt → context → intent. Moses 1:39 as intent architecture |
| [intent/01_landscape.md](../intent/01_landscape.md) | Current landscape: tools, frameworks, key voices, trends |
| [intent/02_frameworks-compared.md](../intent/02_frameworks-compared.md) | OpenSpec, TPG, Kiro, BMAD, Spec Kit — comparison |
| [intent/03_beyond-intent.md](../intent/03_beyond-intent.md) | **Seven gospel patterns the industry hasn't discovered** — Covenant, Stewardship, Line-upon-Line, Atonement, Zion, Sabbath, Consecration |
| [intent/04_synthesis.md](../intent/04_synthesis.md) | What to build: .spec/ directory, Phase 0 practices, file-based state |
| [intent/05_scope-assessment.md](../intent/05_scope-assessment.md) | Concentric rings: Agentic Engineering → SDD → Context → Intent → Beyond Intent |
| [intent/covenant.md](../intent/covenant.md) | Deep scripture study on the covenant mechanism — cutting, binding, tokens, progressive trust |

### External Sources (analyzed)

| Source | Key Contribution |
|--------|-----------------|
| [Nate B Jones — "4 Skills" video](https://www.youtube.com/watch?v=BpibZSMGtdY) | The 4-discipline framework: Prompt Craft → Context → Intent → Specification |
| [Anthropic — Claude Prompting Best Practices](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices) | Official guide for Claude Opus 4.6, saved as [claude-guide.md](claude-guide.md) |
| [Nate B Jones — Prior 5 videos](../intent/00_index.md) | 5 Levels of AI Coding, $1000/Day, Job Market Split, Career Opportunity, Prompt Engineering is Dead |
| [Patrick Debois — Intent-Driven Development](https://www.youtube.com/watch?v=kMRHuc36AK4) | DevOps pioneer's 4 AI-native patterns |
| [Anthropic — Context Engineering for Agents](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) | Anthropic's own context engineering guide |
| [Tyler Brandt — The Intent Layer](https://intent-systems.com/learn/intent-layer) | Hierarchical AGENTS.md files as progressive context disclosure |
| [Paweł Huryn — Intent Engineering Framework](https://www.productcompass.pm/p/intent-engineering-framework-for-ai-agents) | Objective + Desired Outcomes + Health Metrics + Strategic Context + Constraints |
| [GitHub Blog — Spec-Driven Development](https://github.blog/ai-and-ml/generative-ai/spec-driven-development/) | GitHub's official entry into SDD |

### Self-Assessment (completed)

| Document | Purpose |
|----------|---------|
| [eval.md](eval.md) | Personal evaluation against Nate's 4-skill framework + orchestration analysis |
| [claude-guide.md](claude-guide.md) | Claude Opus 4.6 best practices reference |

---

## The Guide Series: Proposed Structure

### Part 0: The Foundation — Why This Matters

**File:** `../guide/00_foundation.md`
**Audience:** Anyone who uses AI for work — developer, manager, individual contributor
**Core thesis:** AI is enforcing a communication discipline that the best leaders have always practiced intuitively. Now everyone needs it.

Covers:
- The 10x gap between 2025 and 2026 prompting skills (Nate's Tuesday example)
- The shift from synchronous to autonomous — why conversational prompting has a ceiling
- What "prompting" actually means now: four disciplines, not one
- Why this isn't just about AI — it's about clear thinking (Toby Lütke's insight)
- The gospel connection: if God works through intelligent agents, His patterns are our curriculum

**Status:** Not started

---

### Part 1: Prompt Craft — The Foundation Layer

**File:** `../guide/01_prompt-craft.md`
**Prior work:** Original copilot-instructions.md, [eval.md](eval.md) § Discipline 1
**Core thesis:** Prompt craft is table stakes. If you can't write a clear, well-structured prompt, nothing else matters. But most people overestimate their skill here.

Covers:
- Claude's best practices: clarity, examples (3-5), XML tags, roles, context at top
- The golden rule: if a colleague would be confused, the model will be too
- Building a prompt library — saved, tested, baseline prompts for recurring tasks
- Tell the model what TO do, not what NOT to do
- Self-contained problem statements (Nate's Primitive #1)
- Format control and output steering
- Individual prayer as gospel parallel (Matthew 7:7 — "ask, and it shall be given")

**Status:** Not started

---

### Part 2: Context Engineering — The Information Architecture

**File:** `../guide/02_context-engineering.md`
**Prior work:** [03_intelligence-cleaveth-gospel.md](../03_intelligence-cleaveth-gospel.md), [04_intent-engineering.md](../04_intent-engineering.md) § context sections, [eval.md](eval.md) § Discipline 2
**Core thesis:** The prompt is 0.02% of what the model sees. Context engineering is the other 99.98%.

Covers:
- What context engineering actually is: system prompts, tool definitions, retrieved docs, memory, MCP
- Building your personal context layer (.claude.md, agent configs, skills)
- MCP servers as context infrastructure
- Progressive disclosure: line upon line (our gospel pattern #3)
- Token optimization: relevant tokens vs. all tokens
- Context degradation as context grows — Anthropic's core insight
- Toby Lütke: "A lot of what people call politics is actually bad context engineering for humans"
- Architecture examples from this project: 9 agents, 8 skills, 6 MCP servers
- Claude's best practices: long context, structured documents, ground responses in quotes

**Status:** Not started

---

### Part 3: Intent Engineering — The Purpose Layer

**File:** `../guide/03_intent-engineering.md`
**Prior work:** [04_intent-engineering.md](../04_intent-engineering.md), [04_intent-engineering-gospel.md](../04_intent-engineering-gospel.md), [intent/03_beyond-intent.md](../intent/03_beyond-intent.md), [intent/covenant.md](../intent/covenant.md), [eval.md](eval.md) § Discipline 3
**Core thesis:** Context tells agents what to know. Intent tells agents what to want. Without intent, capable agents optimize for the wrong thing.

Covers:
- The Klarna story: 2.3M conversations resolved, wrong metric optimized
- Intent = purpose + values + trade-off hierarchies + decision boundaries
- Moses 1:39 as the intent architecture model
- D&C 121 governance: persuasion, not compulsion
- Values hierarchies: when depth conflicts with speed, which wins?
- Decision boundaries: what agents decide vs. what they escalate
- The covenant pattern: mutual commitment, not unilateral command
- Intent preambles on every document (Phase 0 practice)
- The hard truth: intent engineering failure doesn't waste a morning — it screws up the company

**Status:** Not started

---

### Part 4: Specification Engineering — The Blueprint Layer

**File:** `../guide/04_spec-engineering.md`
**Prior work:** [01_planning-then-create-gospel.md](../01_planning-then-create-gospel.md), [intent/04_synthesis.md](../intent/04_synthesis.md), Nate's 5 Primitives, [eval.md](eval.md) § Discipline 4
**Core thesis:** The practical skill going forward is not writing code or crafting prompts — it's the ability to describe an outcome with enough precision that an autonomous system can execute against it for days.

Covers:
- Abraham 4-5: spiritual creation before temporal — blueprint before building
- The 5 Specification Primitives:
  1. Self-contained problem statements
  2. Acceptance criteria (what "done" looks like)
  3. Constraint architecture (musts, must-nots, preferences, escalation triggers)
  4. Decomposition (<2-hour independently verifiable tasks)
  5. Evaluation design (3-5 test cases with known good outputs)
- The .spec/ directory in practice (file-based, version-controlled, agent-readable)
- Your entire organizational corpus as specification
- Anthropic's Opus 4.5 example: spec engineering as the fix for over-scoped agents
- Planner-worker architecture: the spec determines the quality ceiling
- Real-time prompting rewards verbal fluency; spec engineering rewards completeness of thinking

**Status:** Not started

---

### Part 5: The Complete Cycle — From Intent to Zion

**File:** `../guide/05_complete-cycle.md`
**Prior work:** [intent/03_beyond-intent.md](../intent/03_beyond-intent.md), [eval.md](eval.md) § orchestration analysis
**Core thesis:** The industry stops at 4 disciplines. The gospel gives us 11 steps from Intent to Zion — and the patterns beyond specification are where the real breakthroughs live.

Covers:
- The 11-step creation cycle: Intent → Covenant → Stewardship → Spiritual Creation → Line upon Line → Physical Creation → Review → Atonement → Sabbath → Consecration → Zion
- The 7 gospel patterns the industry hasn't discovered
- Conductor vs. Bishop: why the ward hierarchy is the right orchestration model for autonomous AI
- Multi-agent alignment through shared intent (Zion pattern)
- Progressive trust through demonstrated faithfulness (Stewardship/Matthew 25)
- Redemptive error recovery (Atonement pattern vs. "revert and retry")
- Structured reflection cycles (Sabbath pattern — built into the rhythm, not optional)
- Consecrated resource allocation (token budgets serve purpose, not just economics)
- The meta-pattern: God solved multi-agent alignment before we had agents

**Status:** Not started

---

### Part 6: At Scale — The Enterprise Architecture

**File:** `../guide/06_enterprise-architecture.md`
**Prior work:** [intent/05_scope-assessment.md](../intent/05_scope-assessment.md), [intent/machine-proposal.md](../intent/machine-proposal.md), [eval.md](eval.md) § orchestration
**Core thesis:** Everything that works for one person works for 100 people — the patterns are fractal. The gospel's hierarchical structure (ward→stake→area→region→seventy→twelve→prophet) is the scaling architecture.

Covers:
- The ward model applied: organizational intent inheritance at each level
- One-person business → small team → department → enterprise — fractal scaling
- Dedicated roles: context engineers, spec engineers, intent architects
- Machine infrastructure (local + cloud hybrid)
- How to introduce these practices in a company that hasn't started
- Agent governance at organizational scale
- The Nate challenge: "If you are a one-person business, just convert your Notion to be agent-readable"

**Status:** Not started

---

## Dependencies and Sequencing

```
Part 0 (Foundation)
  └→ Part 1 (Prompt Craft) — builds on foundation
       └→ Part 2 (Context) — builds on prompt craft
            └→ Part 3 (Intent) — builds on context
                 └→ Part 4 (Spec) — builds on intent
                      └→ Part 5 (Complete Cycle) — synthesizes 1-4 with gospel patterns
                           └→ Part 6 (Enterprise) — scales the complete cycle
```

Each part is independently readable but cumulative. Nate's key point: you cannot skip lower layers. You cannot write good specs if you can't write good prompts. You can't align agents without understanding context. They all go together.

---

## Where This Sits in the Larger Project

```
docs/work-with-ai/
├── 01-04_*.md              ← Original 4-part series (secular + gospel)
├── expound-prompt.md       ← Prompt engineering documentation
├── intent/                 ← Deep research (5 docs + covenant + machine proposal)
├── prompt/                 ← THIS DIRECTORY
│   ├── 00_guide-plan.md   ← This file — master plan
│   ├── claude-guide.md    ← Claude Opus 4.6 prompting best practices
│   └── eval.md            ← Personal evaluation against 4-skill framework
└── guide/                  ← THE GUIDE SERIES (to be created)
    ├── 00_foundation.md
    ├── 01_prompt-craft.md
    ├── 02_context-engineering.md
    ├── 03_intent-engineering.md
    ├── 04_spec-engineering.md
    ├── 05_complete-cycle.md
    └── 06_enterprise-architecture.md
```

## Key Insight

Nate B Jones at [40:19](https://www.youtube.com/watch?v=BpibZSMGtdY&t=2419):

> "The prompt by itself is dead. The specification, the context, the organizational intent — that is where the value in prompting is moving toward."

Our research at [intent/03_beyond-intent.md](../intent/03_beyond-intent.md):

> "God solved the multi-agent alignment problem before we had agents."

The guide series brings these two threads together: the industry's best practices and the gospel's eternal patterns. Not as forced analogy, but as practical architecture. The same mind that reads Moses 1 on Sunday morning is better equipped to design agent architectures on Monday morning.

---

## Become

This plan is itself a specification. By writing it, I'm practicing Discipline 4. Each part has:
- A self-contained purpose
- Clear source material (acceptance criteria of a sort)
- Dependencies mapped (decomposition)
- A thesis (what "done" looks like for each)

If I can't follow my own framework to build the framework, the framework doesn't work.

Let's find out.

---

*Part of the [Intent-Driven Development Research](../intent/00_index.md) project.*
