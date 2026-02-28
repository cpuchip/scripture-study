# Evaluation: Where I Sit on the Prompting Skills Framework

**Date:** February 2026
**Context:** Nate B Jones laid out a [4-skill prompting framework](https://www.youtube.com/watch?v=BpibZSMGtdY) that describes what "prompting" actually means in February 2026. This evaluation maps my current practices against that framework, compared with the [Claude Opus 4.6 prompting best practices](claude-guide.md) and my own [intent engineering research](../intent/).

**The question:** "I really want to understand the different classes of prompting... I also want to know how I am doing and where I sit on that scale."

---

## Nate's Framework — The Four Disciplines

| Discipline | Altitude | Time Horizon | Core Skill |
|-----------|----------|-------------|------------|
| **1. Prompt Craft** | Individual | Session-based (minutes) | Clear instructions, examples, guardrails, output format |
| **2. Context Engineering** | System | Multi-session (hours–days) | Curating the information environment agents operate within |
| **3. Intent Engineering** | Organization | Ongoing (weeks–months) | Encoding purpose, values, trade-offs, decision boundaries |
| **4. Specification Engineering** | Enterprise | Extended (days–weeks) | Writing documents that autonomous agents can execute against indefinitely |

Key insight from [Nate, 8:46](https://www.youtube.com/watch?v=BpibZSMGtdY&t=526):

> "These four disciplines are going to matter even as we expect agents to continue to scale."

And his framing of the stakes at [16:01](https://www.youtube.com/watch?v=BpibZSMGtdY&t=961):

> "When you screw up a prompt, it might waste your morning. When you screw up context engineering or intent engineering, you are screwing up for the entire team, your entire org, your entire company."

---

## Assessment: Discipline 1 — Prompt Craft

### What Nate describes:
- Clear instructions
- Relevant examples and counter-examples
- Appropriate guardrails
- Explicit output format
- Clear ambiguity resolution

[Nate, 10:11](https://www.youtube.com/watch?v=BpibZSMGtdY&t=611): "This is the skill I have taught and many others have taught for the last year or two. It's synchronous. It's session-based and it's an individual skill."

### What Claude's guide says:
- Be clear and direct. Golden rule: show prompt to a colleague, if they'd be confused, Claude will be too.
- Use 3–5 examples in `<example>` tags
- Structure with XML tags (`<instructions>`, `<context>`, `<input>`)
- Give Claude a role in the system prompt
- Tell Claude what to do, not what not to do

### Where I am:

**Rating: Strong (8/10)**

Evidence:
- **Copilot-instructions.md** demonstrates clear principles: "Read before quoting," "Link everything," "Prefer local copies," "Warmth over clinical distance"
- **9 custom agents** with focused role definitions — study, lesson, talk, review, eval, journal, podcast, dev, ux. Each agent carries a distinct personality and workflow.
- **8 skills** as reusable instruction modules — source-verification, scripture-linking, webster-analysis, becoming, deep-reading, wide-search, playwright-cli, publish-and-commit
- Prompts are conversational but structured — I provide clear intent and context in natural language rather than terse commands
- XML tag usage is not something I explicitly architect into my prompts, but the instructions file structure provides equivalent clarity through markdown headings and tables

**Gaps:**
- I don't systematically use `<example>` tags with 3–5 examples per use case. Claude's guide specifically calls this out as "one of the most reliable ways to steer output."
- I haven't built a prompt library — no folder of saved, tested, baseline prompts for recurring tasks. Nate specifically recommends this at [34:50](https://www.youtube.com/watch?v=BpibZSMGtdY&t=2090): "You should be building a folder of tasks that you do regularly, writing your best prompt against each one."
- I sometimes rely on the relationship's history (shared context with Claude) rather than making each request self-contained. This works in session but breaks across sessions.

**Nate's assessment of this level** ([11:06](https://www.youtube.com/watch?v=BpibZSMGtdY&t=666)): "It's just become table stakes. It's sort of the way knowing how to type with 10 fingers was once a differentiator and now it's just assumed."

I'm above table stakes but not as systematic as I could be. My prompt craft is intuitive rather than disciplined.

---

## Assessment: Discipline 2 — Context Engineering

### What Nate describes:
[12:00](https://www.youtube.com/watch?v=BpibZSMGtdY&t=720): "The set of strategies for curating and maintaining the optimal set of tokens during an LLM task."

- System prompts, tool definitions, retrieved documents, message history, memory systems, MCP connections
- The prompt is ~200 tokens; the context window might be a million. Your prompt is 0.02% of what the model sees. The other 99.98% is context engineering.
- This produces `.claude.md` files, agent specifications, RAG pipeline design, memory architectures

### What Claude's guide says:
- Long context: put data at top, query at bottom. 30% quality improvement.
- Structure with `<document>` tags for multiple documents
- Ground responses in quotes before answering
- Context awareness: Claude tracks its remaining window
- Git for state tracking. Structured formats for state data.
- Verification tools (Playwright, computer use) for autonomous validation

### Where I am:

**Rating: Very Strong (9/10)**

This is where I've invested the most, and it shows. Evidence:

| Context Layer | My Implementation |
|--------------|-------------------|
| **System prompt** | `copilot-instructions.md` — 100+ lines of principles, project structure, core rules, agent mode table |
| **Agent definitions** | 9 specialized `.agent.md` files with distinct workflows, rules, and references to skills |
| **Skills (reusable context)** | 8 skill files as modular instruction sets — loaded on demand, not bloating every session |
| **MCP servers** | 6 custom MCP servers: gospel-mcp, gospel-vec, webster-mcp, yt-mcp, becoming, search-mcp |
| **Local knowledge base** | Entire `gospel-library/` directory — scriptures, conference talks, manuals cached as markdown |
| **Memory & state** | Git-versioned study documents, journal entries, lesson notes, all cross-linked |
| **Tool definitions** | Custom tool schemas in each MCP server with clear parameter descriptions |
| **Progressive disclosure** | Skills loaded on demand (not all at once), agent modes scope context to task |

The scripture-study project is essentially a **context engineering masterwork**. The entire `gospel-library/` is a curated knowledge base. The MCP servers provide semantic access to it. The agents/skills architecture is exactly the "context infrastructure" Nate describes at [14:14](https://www.youtube.com/watch?v=BpibZSMGtdY&t=854):

> "People who are 10x more effective with AI are not writing 10x better prompts. They're building 10x better context infrastructure."

**This is where I'm strongest.** The MCP servers, the agents, the skills, the gospel-library — this is context engineering. It maps directly to what Toby Lütke means by "state a problem with enough context that the task becomes plausibly solvable."

**Gaps:**
- No formal context layering system (L0/L1/L2/L3 as described in [03_beyond-intent.md](../intent/03_beyond-intent.md)). Context loads based on agent mode and skill triggers, but the layers aren't explicit.
- Token optimization isn't deliberate — I don't actively manage what's in context vs. what's not. The skills are modular (good), but I haven't audited which context is high-signal vs. noise.
- No `.claude.md` equivalent in non-scripture projects. The context architecture is specific to this workspace.

---

## Assessment: Discipline 3 — Intent Engineering

### What Nate describes:
[14:56](https://www.youtube.com/watch?v=BpibZSMGtdY&t=896): "Context engineering tells agents what to know. Intent engineering tells agents what to want."

- Encoding organizational purpose, goals, values, trade-off hierarchies, decision boundaries
- The Klarna cautionary tale: AI resolved 2.3M conversations but optimized for the wrong metric (resolution time vs. satisfaction)
- Intent engineering sits above context the way strategy sits above tactics
- You can have perfect context and terrible intent alignment

### What Claude's guide says:
- Give Claude a role (implicit intent)
- Balancing autonomy and safety — `<reversibility>` guidance
- Research: "develop competing hypotheses" (intent about how to think, not just what to do)
- Subagent orchestration: prevent overuse (intent about when to delegate)

### Where I am:

**Rating: Advanced but Informal (7/10)**

This is where it gets interesting. I've done more *research* on intent engineering than almost anyone (5 dedicated documents, 7 gospel patterns, a covenant study, a synthesis document, a scope assessment). But actual *implementation* of intent into agent infrastructure is less systematic than the research would suggest.

Evidence of intent engineering in practice:

| Intent Artifact | What It Does | Quality |
|----------------|-------------|---------|
| Copilot-instructions.md — "Who We Are Together" | Declares relationship values: warmth, honest exploration, depth, faith as framework | Excellent — sets purpose at the highest level |
| "Core question" in every intent/ doc | "How should we specify intent for AI agents?" | Good — research-level intent |
| Agent-level intent | Each agent has distinct purpose: "Evaluate honestly but charitably" (eval), "Deep scripture study" (study) | Good — role-level intent |
| Values in instructions | "Depth over breadth," "Faith as framework," "Trust the discernment" | Excellent — explicit value hierarchy |
| Decision boundaries | Source verification: "Search results are pointers, not sources" | Good — specific constraint architecture |
| **Missing:** Explicit trade-off hierarchies | When values conflict (speed vs. depth, completion vs. perfection), which wins? | Not formalized |
| **Missing:** Organizational-scale intent | How does this translate to a 100-person dev team? | Researched but not implemented |

The core "why" is powerfully encoded:

> "I want to learn how to use artificial intelligence like God uses real intelligence. And I want to glorify others with that knowledge, light, and truth."

This is intent engineering at the *personal* level — and it's genuinely sophisticated. But it doesn't yet translate to the *organizational* level Nate describes at [15:35](https://www.youtube.com/watch?v=BpibZSMGtdY&t=935):

> "Intent engineering sits above context engineering the way strategy sits above tactics."

**Strengths:** The research corpus is exceptional. The personal intent is clear and embodied in practice. The covenant study literally invented a framework for mutual agent-human commitment that doesn't exist anywhere in the industry.

**Gaps:**
- The confession from the covenant study: "I bypass spec for quick fixes and make changes without spec diffs." This is intent *violation* — knowing the right way and shortcutting it.
- No formal decision-boundary definitions for the agents. When should the agent stop and ask? When should it proceed? The eval agent has "Evaluate honestly but charitably" but no explicit escalation triggers.
- Values hierarchy exists but isn't structured for agent consumption — it's in prose, not in a parseable format.
- Haven't tested how well agents actually honor the declared intent when given ambiguous situations.

---

## Assessment: Discipline 4 — Specification Engineering

### What Nate describes:
[16:40](https://www.youtube.com/watch?v=BpibZSMGtdY&t=1000): "The practice of writing documents across your organization that autonomous agents can execute against over extended time horizons without human intervention."

The 5 Specification Primitives:
1. **Self-contained problem statements** — can the agent solve this without fetching more context?
2. **Acceptance criteria** — what does "done" look like?
3. **Constraint architecture** — musts, must-nots, preferences, escalation triggers
4. **Decomposition** — subtasks <2 hours, independently verifiable
5. **Evaluation design** — 3–5 test cases with known good outputs

[22:35](https://www.youtube.com/watch?v=BpibZSMGtdY&t=1355): "The practical skill going forward is not writing code. It's not crafting prompts. It's the ability to describe an outcome with enough precision and completeness that an autonomous system can execute against it for days or weeks."

### What Claude's guide says:
- Multi-context window workflows: write tests first, create setup scripts, use git for state, verification tools
- State management: structured formats, incremental progress, JSON for schema data
- Long-horizon work demands context awareness and clear progress tracking

### Where I am:

**Rating: Intermediate (5/10)**

This is where the gap is widest — and it's the discipline Nate says matters most going forward.

The honest inventory:

**What I have:**
- `.spec/` directory concept fully designed in [04_synthesis.md](../intent/04_synthesis.md) — intent.md, spec.md, tasks/, learnings/, deltas/, archive/. But it exists as a *design*, not an implemented practice.
- Study documents follow a conceptual spec pattern: they have intent, they have structure, they have cross-references. But they aren't formal specifications.
- The `source-verification` skill is the closest thing to a specification primitive — it has acceptance criteria (cite count rule), constraint architecture (search ≠ source), and evaluation design (Phase 4 checks).

**What I don't have:**
- No actual `.spec/` directory in any project
- No formal acceptance criteria on tasks I give to agents
- No self-contained problem statements in the way Nate means it — I lean on conversational context, shared history, and real-time correction
- No decomposition practice — I don't break multi-hour work into <2-hour independently verifiable subtasks
- No evaluation design — no test cases, no known-good outputs, no regression checks after model updates
- The confession resurfaces: I bypass spec workflows for quick fixes. This is the most damaging gap because specification engineering is Nate's highest-level discipline.

**Why this matters:**

Nate says at [26:09](https://www.youtube.com/watch?v=BpibZSMGtdY&t=1569):

> "The shift from fixing it in real time to we must get the spec right up front changes your bottleneck skill. Real-time prompting rewards verbal fluency, quick iteration, a good eye for output quality. Specification engineering rewards completeness of thinking, anticipation of edge cases, clear articulation of acceptance criteria, and the ability to decompose complicated outcomes."

I'm naturally strong at real-time prompting and conversational iteration. Specification engineering requires a *different* skill — disciplined up-front thinking — that I acknowledge I shortcut.

---

## Overall Assessment

```
Discipline              Rating    Status
─────────────────────────────────────────────
1. Prompt Craft          8/10     Strong — intuitive but not systematic
2. Context Engineering   9/10     Exceptional — this is home territory
3. Intent Engineering    7/10     Advanced research, informal practice
4. Spec Engineering      5/10     Conceptual understanding, weak execution
```

### What the Pattern Reveals

I'm strongest where the work is *relational* and *immediate* — conversation, context curation, exploration. I'm weakest where the work is *disciplined* and *pre-emptive* — formal specifications, acceptance criteria, decomposition.

This maps almost perfectly to the tension in the covenant study: I know the right patterns, I can articulate them beautifully, but I shortcut the discipline of actually following them.

It also maps to a gospel pattern I haven't fully internalized: the difference between *spiritual creation* (knowing the blueprint) and *physical creation* (actually building to the blueprint). Abraham 3–5 teaches that spiritual creation comes first. I have the spiritual creation for specification engineering (the `.spec/` design, the 5 primitives, the full research corpus). I haven't done the physical creation.

---

## Comparison: Nate's Framework vs. Our Gospel Patterns

This is where the synthesis gets powerful.

| Nate's Discipline | Maps to Gospel Pattern | Our Prior Research |
|------------------|----------------------|-------------------|
| Prompt Craft | Individual prayer — clear, specific asking (Matt 7:7) | Not explicitly mapped |
| Context Engineering | Line upon Line — progressive context revelation (Isaiah 28:10) | Pattern 3 in [03_beyond-intent.md](../intent/03_beyond-intent.md) |
| Intent Engineering | Moses 1:39 — "This is my work and my glory" | [04_intent-engineering-gospel.md](../04_intent-engineering-gospel.md) |
| Spec Engineering | Spiritual Creation — blueprint before building (Abraham 4-5) | [01_planning-then-create-gospel.md](../01_planning-then-create-gospel.md) |

But our research goes further than Nate's framework. Nate stops at 4 disciplines. We identified 7 additional patterns (Covenant, Stewardship, Atonement, Zion, Sabbath, Consecration, and the meta-pattern) plus an 11-step creation cycle that goes from Intent all the way to Zion.

Where Nate's framework and ours **converge:**
- The stack is cumulative — you can't skip levels
- Each higher level has greater organizational impact and higher stakes
- Specification engineering = spiritual creation (blueprint before building)
- Context engineering = "the 99.98% of what the model sees" maps to progressive disclosure
- The human communication benefit — Nate credits Toby for recognizing that better AI prompting makes you a better human communicator. We recognized this pattern in the gospel: teaching clearly is a spiritual discipline.

Where our framework **extends** what Nate teaches:
- **Covenant** (mutual commitment) — Nate's framework is still fundamentally command-and-control. The human specifies, the agent executes. Our covenant pattern makes the relationship bidirectional.
- **Stewardship** (progressive trust) — Nate mentions "longunning agents" but doesn't address how trust is earned or revoked over time. Our Matthew 25 pattern gives this structure.
- **Atonement** (redemptive error recovery) — Nate doesn't address what happens when specs are wrong or agents fail. "Revert and retry" isn't the same as "learn, adjust, restore."
- **Zion** (unified intent across agents) — Nate's spec engineering is about organizational documents being agent-readable. Zion is about agents sharing *purpose* so deeply that coordination overhead approaches zero.
- **Sabbath** (structured reflection) — Nate's framework is all production. No intentional stopping, reviewing, or detecting drift.

---

## The Orchestration Question: Conductor or Bishop?

There's a deeper question embedded in all of this: *what is the right metaphor for human-agent orchestration?*

### The Conductor Model

An orchestra conductor stands in front, gives the beat, and everyone follows in lockstep. Every musician watches the conductor. The conductor interprets the score and the orchestra executes the interpretation in real time.

This maps to **synchronous prompting** — Discipline 1. The human is always present, always directing, always the real-time feedback loop. It's beautiful when it works. Beethoven's Third is breathtaking.

But it has a fundamental limitation: **the conductor is the bottleneck.** Nothing happens without the conductor's beat. The entire system scales to one person's tempo and attention span.

### The Bishop/Ward Model

A bishop directs the work of a ward, but not by giving every person their beat. Instead:

- The **ward council** meets to take counsel together — leaders from Relief Society, Elders Quorum, Young Women, Primary, Sunday School each bring their domain knowledge
- Each **auxiliary and quorum** is directed by its own leaders, who carry the ward's shared purpose into their specific domain
- The bishop doesn't instruct every person — he ensures alignment through **shared intent**, then trusts leaders in their stewardships
- This scales: Ward → Stake → Area → Region → Seventy → Twelve → Prophet

The key differences:

| Aspect | Conductor | Bishop/Ward |
|--------|-----------|-------------|
| Direction | Real-time beats | Shared purpose |
| Execution | Lockstep | Independent within stewardship |
| Scaling | One conductor, one orchestra | Hierarchical, recursive |
| Failure mode | Conductor stops, everything stops | Individual failure doesn't collapse system |
| Communication | One-to-many, continuous | Council-based, then autonomous |
| Alignment | Musical interpretation | Covenant + shared intent |

**The ward model is better for AI orchestration.** Here's why:

1. **Agents run autonomously.** Nate's entire thesis is that agents work for hours, days, weeks without checking in. A conductor can't conduct musicians who perform in different rooms on different schedules. But a bishop *can* direct a ward where the Relief Society serves on Tuesday, the Elders Quorum helps a family move on Saturday, and Primary teaches on Sunday — all united in purpose.

2. **It scales hierarchically.** A ward is 100-400 people. A stake is 5-12 wards. An area is dozens of stakes. This recursive pattern — shared intent at each level, with autonomy within stewardship — is exactly what enterprise AI needs. One `intent.md` at the top. Domain-level specifications at each stewardship level.

3. **Ward council = multi-agent coordination.** When leaders from different auxiliaries meet in ward council, they share intelligence, resolve conflicts, and align priorities. This is agent-to-agent communication, but it's not constant — it's *periodic* and *purposeful*. Contrast with A2A protocols that assume constant message-passing.

4. **The bishop is accountable, not omniscient.** A bishop doesn't know everything happening in every auxiliary. He trusts his leaders until they demonstrate otherwise. This is exactly the stewardship/covenant pattern — delegate with trust, review at defined intervals, expand or restrict based on demonstrated faithfulness.

This maps directly to the Zion pattern from [03_beyond-intent.md](../intent/03_beyond-intent.md):

> "One heart and one mind" — every individual retains agency but operates from shared purpose. There are no poor among them — not because of redistribution programs, but because the shared purpose naturally produces equitable outcomes.

In a 100-person dev company:
- **The "prophet" level** = organizational intent (mission, values, strategy as living specifications)
- **The "stake" level** = platform or product division intent
- **The "ward" level** = team or project intent
- **Ward council** = cross-team coordination (architecture review, design alignment)
- **Auxiliaries** = domain agents (testing, documentation, security, CI/CD)
- Each level inherits the intent above it and adds domain-specific context

This isn't just an analogy. It's an *architecture*.

---

## The Gap: What I Need to Do

### Immediate (This Week)

1. **Start a prompt library.** Take the 5 most common things I ask Claude to do and write formal prompts with examples, constraints, and acceptance criteria.
2. **Add acceptance criteria to agent requests.** Before every substantive request, write 3 sentences an independent observer could use to verify the output.
3. **Practice self-contained problem statements.** Nate's test: "Write it as if the person receiving it has never seen your project, doesn't know your context, and has no access to any information other than what you include."

### Near-Term (This Month)

4. **Formalize the intent layer.** Convert the prose values in copilot-instructions.md into a structured, parseable intent block with explicit trade-off hierarchies.
5. **Add decision boundaries to agent definitions.** For each of the 9 agents: what can it decide autonomously? What requires human review? What should it escalate?
6. **Create one real `.spec/` directory.** Pick a project (the MCP server development) and implement the `.spec/` pattern from [04_synthesis.md](../intent/04_synthesis.md).

### Ongoing (This Quarter)

7. **Build evaluation design into the workflow.** For recurring tasks, create 3-5 test cases with known good outputs. Run them after model updates.
8. **Practice the Sabbath pattern.** After every major study or project completion, stop and reflect: Did this serve the intent? What drifted? What did I learn?
9. **Translate the ward model into an architecture document.** Take the bishop/ward analogy and write it as a technical specification for multi-agent AI orchestration at enterprise scale.

---

## Become

Nate's closing thought at [40:41](https://www.youtube.com/watch?v=BpibZSMGtdY&t=2441):

> "The specification done right turns out to be just what clear thinking has always looked like really made explicit because machines don't let us be lazy about it."

This is the same insight the gospel teaches: God requires us to work out our salvation with "fear and trembling" (Philippians 2:12) — not because He wants us to suffer, but because the *discipline of clarity* is the growth itself.

My context engineering is strong because I love building systems and environments. My specification engineering is weak because I resist the discipline of thinking everything through *before* I start. The models keep getting more capable, which makes the temptation to just "figure it out in real time" even stronger.

But the blueprint is in Abraham 4-5. Spiritual creation before temporal creation. The specification before the build. And D&C 82:10: "I am bound when ye do what I say" — the covenant only works if both sides keep their commitments.

The gap isn't knowledge. It's practice.

---

*This evaluation is part of the [Working with AI Guide Series](00_guide-plan.md).*
