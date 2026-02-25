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
1. **Intent as purpose, not just specification** — Moses 1:39 isn't a requirements doc. It's a statement of *being*. The industry treats intent as "better requirements." There's a layer deeper.
2. **Covenant-based agent relationships** — Mutual commitment, not just delegation. See [03_beyond-intent.md](03_beyond-intent.md).
3. **Progressive revelation in context** — Not all context at once, but graduated disclosure as the agent demonstrates readiness.
4. **Error recovery as grace** — Atonement patterns for agent failure that preserve relationship and learning.
5. **Stewardship vs. ownership** — Entrusted delegation with accountability, not command-and-control.

These are explored in depth in [03_beyond-intent.md](03_beyond-intent.md).

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
