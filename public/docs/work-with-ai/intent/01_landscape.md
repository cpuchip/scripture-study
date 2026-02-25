# The Spec-Driven Development Landscape (February 2026)

**Part of:** [Intent-Driven Development Research](00_index.md)
**Date:** February 2026
**Status:** Active

---

## The Convergence

Something remarkable happened in early 2026: the entire software industry converged on the same conclusion from different directions. Whether they call it Intent-Driven Development (IDD), Spec-Driven Development (SDD), Adaptive Intent-Driven Development (AIDD), or just "the end of vibe coding" — the message is identical:

> **"Code is about to cost nothing. Knowing what to build is about to cost everything."**
> — [Nate B Jones, Job Market Split, ~3:19](https://www.youtube.com/watch?v=RtMLnCMv3do&t=199)

---

## The Evidence

### The Dark Factory Is Real

StrongDM operates what they call a "Software Factory": **3 engineers, no human code writing, no human code review.** Their entire system runs on markdown specification files that orchestrate AI agents. Their benchmark: if you haven't spent $1,000/engineer/day on AI tokens, your factory has room for improvement.

> "Code must not be written by humans. Code must not be even reviewed by humans."
> — [5 Levels, ~6:35](https://www.youtube.com/watch?v=bDcgHzCBgmQ&t=395)

This isn't theoretical. 90% of Claude Code's codebase was written by Claude Code. Codex features are entirely AI-written. Claude Code is converging toward 100% self-written.

### The Paradox: AI Makes Most Developers *Slower*

A 2025 METR randomized control trial found experienced developers were **19% slower** with AI tools — while believing they were **24% faster**. CodeRabbit's analysis: AI-generated code has **1.7x more logic issues** (not syntax — *logic*). Google DORA: 9% climb in bug rates correlating with 90% increase in AI adoption, plus 91% increase in code review time.

The gap isn't the tools. It's the workflow. Bolting AI onto unreformed processes creates a J-curve — productivity dips before it surges. Most organizations are sitting at the bottom of that J-curve.

### The Economics Have Shifted

> "For 60 years, the unit of work in software was the instruction… That era is done. The unit of work is now the token — a unit of purchased intelligence."
> — [$1,000/Day, ~0:00](https://www.youtube.com/watch?v=-bQcWs1Z9a0)

Key data points:
- Per-token inference costs falling **10x–200x per year** (GPT-4 equivalent: $20/M tokens in 2022 → ~$0.40 now)
- Average org spends **$85K/month** on AI (+36% YoY)
- OpenAI planning agent tiers: $2K, $10K, $20K/month
- AI-native startups average **$3.5M revenue/employee** vs. $600K SaaS average
- Cursor: **$16M revenue/employee**; Midjourney: **$200M revenue / 11 people**

Jevons' Paradox applies: cheaper tokens mean more total token consumption, not less. But the cost of *specifying badly* is compounding faster than production cost is falling.

---

## Dan Shapiro's 5 Levels of Vibe Coding

The clearest framework for where people actually are:

| Level | Name | What Happens | Who's Here |
|-------|------|-------------|-----------|
| **0** | Spicy Autocomplete | Tab completion, line suggestions | ~40% of AI tool users |
| **1** | Coding Intern | Discrete, scoped tasks — "write this function" | ~35% |
| **2** | Junior Developer | Multi-file changes, but human still reads ALL code | ~20% (most "AI-native" devs) |
| **3** | Developer as Manager | Directing AI, reviewing at PR/feature level, not line level | ~4% |
| **4** | Developer as PM | Write spec, check if tests pass, code is a black box | <1% |
| **5** | Dark Factory | Specs in → working software out, no human writes or reviews code | StrongDM, parts of Anthropic/OpenAI |

**The critical insight:** 90% of self-described "AI-native" developers are stuck at Level 2. The distance from Level 2 to Level 5 is not technological — it's organizational, cultural, and about spec quality.

---

## Three Developer Career Tracks

From [$1,000/Day](https://www.youtube.com/watch?v=-bQcWs1Z9a0):

| Track | Core Skills | Who |
|-------|------------|-----|
| **Orchestrator** | Spec writing, quality evaluation, token economics, factory management | The PM/architect who directs agent fleets |
| **Systems Builder** | Agent frameworks, eval pipelines, context management, routing layers | The infra engineer building the factory itself |
| **Domain Translator** | Deep domain expertise + enough AI fluency to specify precisely | The person who bridges business knowledge and AI capability |

**Most exposed:** The middle — competent application developers with no deep systems expertise or domain specialization. Not because they can't learn, but because their current value proposition (writing application code) is commoditized fastest.

**Most valuable:** The domain translator — "the person who walks into the panicking exec meeting and says precisely what AI can and can't do for your specific workflows, with implementation plans. That person does not exist in most organizations right now."

---

## The Tools

### OpenSpec (Fission-AI / Hari Krishnan)

**What it is:** A TypeScript CLI for Spec-Driven Development. Core concept: maintain a **single unified specification document** as the authoritative reference for a system's design.

**How it works:**
1. **Source of Truth Spec** — One living document representing the current state of the system
2. **Delta Specs** — Change proposals marking sections as ADDED, MODIFIED, or REMOVED
3. **Propose → Apply → Archive** — Workflow cycle: propose a change, apply it to the source spec, archive the delta

**Key design decisions:**
- Single spec document (not scattered requirements files)
- Delta-based changes (not full rewrites)
- Archived deltas create an audit trail
- Configurable schemas — default "spec-driven" produces `proposal.md → specs.md → design.md → tasks.md`
- Custom schemas for different project types (minimalist, design-heavy, etc.)

**Strengths:**
- File-based — specs live in the repo, travel with the code
- Simple mental model — one source of truth
- Extensible via custom schemas
- Brownfield-friendly — can describe existing systems
- Good for AI context — agents read the spec to understand the system

**Gaps (for our needs):**
- No multi-repo awareness
- No intent layer (specs describe *what*, not *why* or *what trade-offs*)
- No task management (produces `tasks.md` but doesn't track execution)
- Single-project focused — no cross-project dependencies
- No built-in review/eval against intent

### AWS Kiro

**What it is:** Amazon's developer environment that **forces testable specifications before any code generation.**

> "Amazon, a company that profits when you ship faster, decided the most valuable thing it could do is slow you down."
> — [Job Market Split, ~1:38](https://www.youtube.com/watch?v=RtMLnCMv3do&t=98)

**Key insight:** Kiro's innovation isn't the specs themselves — it's making them *mandatory*. The IDE won't generate code until you have a testable specification. This is the first major tool to acknowledge that the bottleneck is spec quality, not code generation speed.

**Gaps:** Proprietary, AWS-locked, IDE-coupled.

### GitHub Spec Kit

GitHub's entry into spec-driven development. Less documented publicly, but aims to integrate specification workflows into GitHub's existing development flow. Likely to be the most widely adopted simply by distribution advantage.

### BMAD Method

A methodology for structured AI-assisted development. Less tool-focused, more process-focused. Emphasizes deliberate specification before agent work.

### Antigravity

Another SDD tool mentioned in the landscape. Details sparse but appears to focus on specification generation and management.

### TPG (Task Planning for GPT)

**What it is:** A Go CLI tool that provides persistent task state, dependency management, and cross-session memory for AI agents. [External context](../../../external_context/tpg/README.md).

**Strengths:**
- Persistent state across sessions (SQLite)
- Dependency tracking between tasks
- `tpg prime` — context injection for agent sessions
- Learnings system — captures and retrieves project knowledge

**Critical limitation (our key insight):**
> "It's almost like we need a tool that doesn't have any internal state... it uses markdown or some other file that can live with the repo to mark state."

TPG's state is in `.tpg/tpg.db` — a SQLite database. This means:
- State doesn't travel with branches
- Can't be code-reviewed
- Not version-controlled with the code
- Not readable by AI agents without TPG CLI
- Not portable across machines or tools

**Our development plan** ([10_intent-development.md](../../10_intent-development.md)) proposes adding intent metadata, multi-repo hub+spoke, and spec traceability — but the fundamental question remains: should the state layer be a database at all, or should it be markdown files in the repo?

---

## The AIDD Framework (Enterprise Scale)

Binoy Ayyagari's [Adaptive Intent-Driven Development](https://medium.com/@binoyayyagari/adaptive-intent-driven-development-aidd-ee0cd5f8741b) takes the broadest view. It reimagines the entire SDLC:

| Traditional Phase | AIDD Transformation |
|------------------|---------------------|
| Requirements | Multi-agent brainstorming (architect + security + performance agents) |
| Design | AI-generated architecture with constraint validation |
| Implementation | Agent fleets executing against specs |
| Testing | Continuous AI-driven validation against specifications |
| Deployment | Autonomous deployment with spec-level rollback |
| Maintenance | Agents monitor drift from spec |

**Key concepts:**
- **A2A + MCP protocols** — Google's Agent-to-Agent and Anthropic's Model Context Protocol as the coordination layer
- **Fluid teams** — Pod and swarm structures that form around work, not org charts
- **12-factor Agent Principles** — Based on the 12-factor app methodology: stateless processes, externalized state, declarative config, disposability
- **Goal-oriented contracts** — Agents negotiate and commit to outcomes, not commands
- **Prompts as interface** — The human-to-agent interface is the specification

**Relevance to us:** The enterprise framing is useful for the user's day job (team-based, multi-branch, multi-repo). The 12-factor agent principles are particularly relevant — especially "stateless processes with externalized state" (which directly supports our insight about file-based state).

---

## Key Voices

| Person | Affiliation | Contribution |
|--------|------------|-------------|
| **Nate B Jones** | AI News & Strategy Daily | Most articulate synthesis of the macro shift. Five videos forming a coherent thesis arc. |
| **Hari Krishnan** | Polarizer Technologies / intent-driven.dev | OpenSpec creator. "Context engineering for AI agents." |
| **Dan Shapiro** | Glow Forge | 5 Levels of Vibe Coding framework — went viral, widely cited |
| **Binoy Ayyagari** | — | AIDD framework — enterprise-scale reimagining of the SDLC |
| **StrongDM engineering team** | StrongDM | Best-documented Level 5 "dark factory" — 3 engineers, markdown specs |
| **François Chollet** | Google / Keras | Translation analogy for AI capability curves (cited & critiqued by Jones) |
| **Patrick Debois** | Tesla / AI Native Dev | DevOps movement pioneer, now mapping AI Native patterns. Four patterns: Producer→Manager, Implementation→Intent, Delivery→Discovery, Content→Knowledge. Building pattern catalog + tool landscape at ainativedev.io. |
| **Paweł Huryn** | Product Compass | Intent Engineering Framework for AI Agents: Objective + Desired Outcomes + Health Metrics + Strategic Context + Constraints + Decision Types + Stop Rules. "Intent is what determines how an agent acts when instructions run out." |
| **Tyler Brandt** | Intent Systems | Built an entire company around the Intent Layer concept. Hierarchical AGENTS.md/CLAUDE.md files as progressive context disclosure. Fractal compression. Maintenance flywheel. "The ceiling on AI results isn't model intelligence—it's what the model sees before it acts." |
| **IndyDevDan** | Agentic Engineer | "Year of Trust" thesis. Core Four (Context+Model+Prompt+Tools). 10 bets for 2026. Path: Base→Better→More→Custom→Orchestrator. "Build the system that builds the system." |
| **Mamdouh Alenezi** | arXiv | Academic paper tracing the evolution from prompt–response to goal-directed systems. BDI model, reference architecture, multi-agent topologies. "Maturation will parallel web services." |
| **Anthropic Engineering** | Anthropic | Official context engineering guide for AI agents. |
| **Thoughtworks** | Thoughtworks | Named SDD as one of 2025's defining engineering trends. |
| **GitHub (Den Delimarsky)** | GitHub | Official blog post + open source toolkit for spec-driven development. |

---

## Patrick Debois's 4 AI Native Patterns

Patrick Debois — the person who helped spark the DevOps movement — is now doing the same for AI-native development. On the [AI Native Dev podcast](https://www.youtube.com/watch?v=kMRHuc36AK4) (March 2025), he identified four emerging patterns:

| Pattern | Shift | Description |
|---------|-------|-------------|
| **Producer → Manager** | Operational | You're no longer writing code — you're reviewing, accepting, directing. Like a manager overseeing AI agents rather than a producer doing the work. |
| **Implementation → Intent** | Specification | You express *what you want*, not *how to build it*. Big spectrum: from tab-completion to full requirements-in/software-out. |
| **Delivery → Discovery** | Product | When generation is cheap, you can prototype 5 versions and discover which is right. Shifts focus from "ship the one thing" to "explore which thing to ship." |
| **Content → Knowledge** | Organizational | Capturing institutional knowledge as a competitive advantage. Agents learn from conversations and surface insights for preservation. |

**Key insight from Debois:** "We're not at principles yet — we're still at patterns, still observing." He explicitly chooses "patterns" over "principles" because the space is too emergent to prescribe. This humility aligns with the gospel pattern of not running faster than you have strength (Mosiah 4:27).

They're building a tool landscape at [landscape.ainativedev.io](https://landscape.ainativedev.io/) — a community-driven catalog of AI-native tools organized by category.

---

## The Intent Layer (Intent Systems / Tyler Brandt)

An entire company has been built around this concept. [Intent Systems](https://intent-systems.com) offers the most developed thinking on **structured context for AI agents in codebases**.

**The problem they solve:** Every agent starts from zero. Every request is a full onboarding — not just to your task, but to your entire system. Agents fumble in the dark, learning only by what they bump into. Tyler calls this the "dark room problem."

**The solution:** A hierarchical layer of Intent Nodes (AGENTS.md or CLAUDE.md files) that provide progressive context disclosure:

1. **Root node** — Global architecture, high-level map
2. **Service nodes** — How services are structured, dependencies
3. **Module nodes** — Specific invariants, anti-patterns, entry points
4. **Leaf nodes** — Detailed patterns and pitfalls for specific code areas

**Key design principles:**
- **Progressive disclosure** — Start with minimum high-signal context; drill down only where needed
- **Fractal compression** — Leaf nodes compress raw code → parent nodes compress children → each layer stands on stable context from below
- **Least Common Ancestor** — Shared knowledge lives at the shallowest node that covers all paths where it's relevant, preventing duplication
- **Maintenance flywheel** — On every merge: detect changes → identify affected nodes → re-summarize leaf-first → human reviews. Can be automated.
- **Reinforcement learning** — When agents hit edge cases, their learnings feed back into the layer. "Your codebase becomes a reinforcement learning environment."

**Our observation:** This is the closest secular analogue to the "line upon line" pattern (Isaiah 28:10). Progressive disclosure mirrors D&C 93:13 — "he received not of the fulness at first, but continued from grace to grace." The maintenance flywheel mirrors "teaching one another" (D&C 88:77-78). But Tyler is thinking only about *context* — not about *covenant* (mutual commitment), *stewardship* (progressive trust/autonomy), or *purpose* (why this code exists in the larger mission).

---

## The Huryn Intent Framework (Product Manager Perspective)

Paweł Huryn approaches intent from product management, not engineering. His [Intent Engineering Framework](https://www.productcompass.pm/p/intent-engineering-framework-for-ai-agents) (Jan 2026) structures agent intent as:

1. **Objective** — The problem being solved + why it matters (aspirational, qualitative)
2. **Desired Outcomes** — Observable states indicating success (measurable, user-perspective)
3. **Health Metrics** — What must not degrade while optimizing (guards against Goodhart's Law)
4. **Strategic Context** — The system the agent operates in
5. **Constraints** — Steering (prompt layer) + Hard (enforced in orchestration)
6. **Decision Types & Autonomy** — Which decisions the agent may take vs. must escalate
7. **Stop Rules** — When to halt, escalate, or complete

**Critical quote:** *"Intent is not a task list, a prompt, or a goal metric. Intent is what determines how an agent acts when instructions run out."*

**Connection to existing frameworks:** Huryn explicitly connects to OKRs (Christina Wodtke) and empowered product teams (Marty Cagan). This is product thinking applied to agents.

**Our observation:** Huryn's framework overlaps significantly with our covenant pattern — especially the mutual commitment aspect (Objective + Constraints = what each party commits to) and Decision Types (= delegated stewardship with boundaries). But he has no concept of *redemptive error recovery* (Atonement), *structured reflection* (Sabbath), or *purpose-driven resource allocation* (Consecration). His framework is excellent for the *structure* of intent; ours adds the *relational* and *spiritual* dimensions.

---

## The Academic Picture (arXiv Paper)

"From Prompt–Response to Goal-Directed Systems" (Alenezi, Feb 2026) provides the academic grounding:

**Key contributions:**
- **BDI Model (Belief-Desire-Intention)** — Classical AI theory from the 1980s/90s. Beliefs = world state + memory. Desires = goals + constraints. Intentions = adopted plans + tool calls. This is the formal framework behind what the industry is rediscovering.
- **Reference Architecture** — Agent Core (LLM reasoning) surrounded by Control Layer (planner, state machine, circuit breakers), Memory Layer (working, episodic, semantic, preferences), Tooling Layer (registry, sandboxes, RAG), Governance & Observability (cross-cutting).
- **Multi-agent topologies** — Orchestrator-Worker, Router-Solver, Hierarchical Command, Swarm/Market. Each with mapped failure modes and mitigations.
- **Enterprise Hardening Checklist** — Identity, policy enforcement, tooling, memory management, observability, budgeted autonomy, data governance, CI/CD, security testing, change management.

**Key thesis:** *"The maturation of agentic AI will follow the trajectory of web services: not by model improvements alone, but through shared protocols, typed contracts, and layered governance that enable composable autonomy at scale."*

**Connection to our work:** The BDI model is remarkable — beliefs, desires, and intentions are explicit architectural components. But there's no concept of *covenant* (mutual binding commitment that changes both parties), *atonement* (redemptive error recovery that preserves relationship), or *Zion* (shared-intent alignment where agents truly share purpose rather than just coordinate). The paper treats agents as tools to be governed; the gospel treats stewards as agents-in-becoming.

---

## IndyDevDan's Trust Thesis

IndyDevDan's [Top 2% Agentic Engineering](https://agenticengineer.com/top-2-percent-agentic-engineering) (Feb 2026) centers everything on **trust**:

> "Want to do more with agents? You need to build agents you trust."
> Trust → Speed → Iteration → Impact

**Core Four:** Context + Model + Prompt + Tools (every agent is just a composition of these)

**The Path:** Base → Better → More → Custom → Orchestrator

**10 Bets for 2026:**
1. Anthropic dominates (best tool-calling, execution track record)
2. Tool calling is the opportunity (only 15% of LLM output tokens are tool calls)
3. Custom agents above all (50 lines, 3 tools, 1 system prompt)
4. Multi-agent orchestration is the next frontier
5. Agent sandboxes (defer trust until the merge)
6. In-loop vs. out-loop agentic coding
7. Agentic Coding 2.0 — agents that conduct agents
8. Private benchmarks over public ones
9. Agents are eating SaaS
10. The death of AGI hype — "There is no AGI. There are just agents."

**Our observation:** Dan's trust thesis maps almost perfectly to the Stewardship pattern (D&C 104 / Matthew 25). "Defer trust until the merge" = "proved in small things, entrusted with greater." His "build the system that builds the system" = the Orchestrator level aligns with the Zion pattern (Moses 7:18) — but he has no framework for *how* trust is built beyond "results." The gospel teaches that trust is built through *covenant* (mutual commitment), *faithfulness* (consistent small acts), and *accountability* (D&C 72:3-4). Dan's framework tells you trust matters; the gospel teaches you *how to build it*.

---

## What the Landscape Tells Us

### The Industry Agrees On
1. **Specs first** — Write before building. Every tool, every framework, every voice.
2. **File-based** — Specs should be readable files (markdown), not database records or internal tool state.
3. **Living documents** — Specs evolve with the code. Not written once and forgotten.
4. **AI context** — Specs serve as context for AI agents, not just human documentation.
5. **Intent matters** — The *why* must be encoded alongside the *what*.

### The Industry Disagrees On
1. **Single spec vs. distributed** — OpenSpec says one source of truth; AIDD implies distributed specs per agent concern; Kiro implies spec-per-feature.
2. **Tool vs. process** — OpenSpec is a tool; BMAD is a process; AIDD is a framework. Some want CLIs, others want IDE integration, others want culture change.
3. **Autonomy level** — How much should agents decide on their own? StrongDM says "everything." Most enterprises say "very little."
4. **Multi-repo coordination** — Nobody has solved this well. Every tool is single-project focused.

### The Industry Isn't Thinking About
1. **Intent as purpose, not just specification** — Moses 1:39 isn't a requirements doc. It's a statement of *being*. The industry treats intent as "better requirements." Huryn gets closest — his Objective layer asks "why does this matter?" — but even he stays at the product level, not the existential level. There's a layer deeper.
2. **Covenant-based agent relationships** — Mutual commitment, not just delegation. Huryn's framework has "Constraints" and "Decision Types" but no concept of the agent's reciprocal commitment. The relationship is one-directional: human defines, agent executes. See [03_beyond-intent.md](03_beyond-intent.md).
3. **Progressive revelation in context** — Tyler Brandt's Intent Layer is the closest secular analogue to "line upon line" (Isaiah 28:10). His hierarchical progressive disclosure is structurally similar. But it's purely *informational* — graduated information access — not *relational* (graduated trust/autonomy based on demonstrated faithfulness).
4. **Error recovery as grace** — Atonement patterns for agent failure that preserve relationship and learning. The arXiv paper has "circuit breakers" and "retry logic" — mechanical recovery. Nobody is thinking about *redemptive* recovery that makes the whole system better through failure.
5. **Stewardship vs. ownership** — Entrusted delegation with accountability, not command-and-control. IndyDevDan's trust thesis is the closest — "defer trust until the merge." But he frames trust as *earned through results*, not *built through covenant and faithfulness*. The gospel teaches that trust is relational, not transactional.
6. **Structured reflection cycles** — Nobody has a Sabbath pattern. The industry has "retrospectives" but no *architectural* reflection that's built into the development cycle as a first-class concern. Patrick Debois's Content→Knowledge pattern is adjacent but reactive (capture knowledge when it surfaces) rather than proactive (schedule regular reflection as part of the workflow).
7. **Shared-purpose alignment (Zion)** — The arXiv paper discusses multi-agent topologies (Orchestrator-Worker, Router-Solver, etc.) but all are *coordination* patterns. None are *alignment* patterns — where agents genuinely share purpose rather than just being directed toward compatible goals.

These are explored in depth in [03_beyond-intent.md](03_beyond-intent.md), and the scope of these gaps is assessed in [05_scope-assessment.md](05_scope-assessment.md).

---

## Implications for Us

### Solo Dev (Scripture Study Repo)
This repo already practices much of what the industry is converging on:
- `.github/copilot-instructions.md` is an intent preamble
- `.github/agents/` are specialized workers with defined workflows
- `docs/work-with-ai/` encodes transferable methodology
- Study documents are effectively specifications for understanding

What's missing: formalized spec lifecycle, intent traceability from study→lesson→becoming, state that lives in the repo (not in TPG's SQLite).

### Team Dev (Day Job)
The user's day job needs:
- Multi-branch spec management — specs that travel with feature branches
- Multi-repo coordination — work across several repos at different stages
- Team-readable state — not a SQLite DB one person owns, but shared markdown
- Integration with existing flow (GitHub Issues, Jira) without replacing it
- Intent that cascades: team mission → project intent → epic intent → task constraints

### The Bridge
The gap between solo and team isn't the *pattern* — it's the *tooling*. Both need:
1. File-based spec state in the repo
2. Intent encoded alongside specs (not separate)
3. Task tracking that inherits intent from parent context
4. Review against intent, not just task completion
5. Cross-project awareness without central databases

No existing tool does all five.

---

## Next Steps

- Compare frameworks head-to-head → [02_frameworks-compared.md](02_frameworks-compared.md)
- Explore what the gospel teaches beyond intent → [03_beyond-intent.md](03_beyond-intent.md)
- Synthesize what to build → [04_synthesis.md](04_synthesis.md)
