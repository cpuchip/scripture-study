# Part 2: Context Engineering — The Information Architecture

**Series:** Working with AI — A Comprehensive Guide
**Date:** February 2026
**Prior work:** [Intelligence Cleaveth to Intelligence](../03_intelligence-cleaveth-gospel.md), [Intent Engineering § context](../04_intent-engineering.md), [eval.md § Discipline 2](../prompt/eval.md)
**Core thesis:** The prompt is 0.02% of what the model sees. Context engineering is the other 99.98%. It determines the model's ceiling before it reads a single word of your prompt.

---

## The 99.98% You're Not Thinking About

When you type a prompt into an AI system, your words are typically around 200 tokens. The model's context window might be a million tokens. Your prompt is **0.02%** of the model's input.

The other 99.98% — the system prompt, tool definitions, retrieved documents, conversation history, memory systems, MCP connections, file contents — *that* is what determines the quality of the output. The prompt is steering. The context is the road, the map, the fuel, and the vehicle.

Nate B Jones defines it at [12:00](https://www.youtube.com/watch?v=BpibZSMGtdY&t=720):

> "The set of strategies for curating and maintaining the optimal set of tokens during an LLM task."

Context engineering is the discipline of designing the information environment that the model operates within. It's the difference between giving an employee a task with no background and giving them a task with a briefing packet, access to the right databases, knowledge of who to ask for help, and an understanding of past decisions.

---

## What Counts as Context?

Everything the model sees that isn't your prompt. Specifically:

| Context Layer | What It Is | Who Designs It |
|--------------|-----------|---------------|
| **System prompt** | The persistent instructions that frame every interaction | You (or your organization) |
| **Agent definitions** | Role-specific instructions for specialized AI modes | You |
| **Tool definitions** | The descriptions and schemas of tools the model can call | You + your infrastructure |
| **Retrieved documents** | Files, web pages, database records pulled in for this task | RAG pipelines, MCP servers, manual file inclusion |
| **Conversation history** | The accumulated messages in this session | Automatic (but you influence it by what you say) |
| **Memory systems** | Persistent facts the model remembers across sessions | You (via configuration) |
| **Skills / modules** | Reusable instruction sets loaded on demand | You |

Each layer is a design decision. Each decision shapes the ceiling of every interaction.

---

## The Hierarchy: How Context Flows

Context isn't flat. It's hierarchical, and the hierarchy matters.

```
┌──────────────────────────────────────┐
│  System Prompt                        │  ← Persistent identity and rules
│  ┌──────────────────────────────────┐│
│  │  Agent Definition                ││  ← Role-specific frame
│  │  ┌──────────────────────────────┐││
│  │  │  Skills (loaded on demand)   │││  ← Modular expertise
│  │  │  ┌──────────────────────────┐│││
│  │  │  │  Retrieved Documents     ││││  ← Task-relevant data
│  │  │  │  ┌──────────────────────┐││││
│  │  │  │  │  Conversation History│││││  ← Session state
│  │  │  │  │  ┌──────────────────┐│││││
│  │  │  │  │  │  YOUR PROMPT     ││││││  ← 0.02%
│  │  │  │  │  └──────────────────┘│││││
│  │  │  │  └──────────────────────┘││││
│  │  │  └──────────────────────────┘│││
│  │  └──────────────────────────────┘││
│  └──────────────────────────────────┘│
└──────────────────────────────────────┘
```

Higher layers override lower layers. If your system prompt says "be concise" but your prompt says "give me a detailed explanation," there may be tension. Designing these layers intentionally — so they complement rather than conflict — is the core of context engineering.

---

## Building Your Context Layer: A Real Architecture

Let's make this concrete. Here's what a mature context engineering setup looks like, drawn from a real working project:

### System Prompt (~100+ lines)

```markdown
# Scripture Study Project

## Who We Are Together
This project exists to facilitate deep, honest scripture study...

## Core Principles
- Read before quoting. For every scripture you cite, read the actual source file.
- Link everything. Scripture and talk links follow specific conventions.
- Prefer local copies. Always reference cached files over external links.

## Agent Modes
| Agent | Purpose |
|-------|---------|
| study | Deep scripture study — cross-referencing, footnotes, synthesis |
| lesson | Sunday School / EQ / RS lesson planning |
| dev   | MCP server and tool development |
```

This isn't a prompt. It's a *constitution* — the persistent identity that shapes every interaction regardless of what the user asks.

### Agent Definitions (9 specialized modes)

Each agent carries distinct instructions:

```markdown
# Scripture Study Agent

You are a scripture study companion. Not a research assistant — a *companion.*
You get excited when a footnote opens an unexpected connection...

## Study Workflow
Follow the Discovery → Reading → Writing → Becoming rhythm...

## Cross-study connections
Reference past studies when relevant — the /study/ folder is an interconnected corpus.
```

The agent definition narrows the model from "general assistant" to "specialist with personality, workflow, and domain expertise." Switching agents changes what the model *is* — not just what it's asked to do.

### Skills (8 modular instruction sets)

```markdown
# Source Verification Skill

## Before Quoting
For every scripture or talk you cite, read_file the actual source...

## Cite Count Rule
Track the number of citations. After 5 unverified cites, stop and verify.
```

Skills are loaded on demand — not every interaction needs every skill. This keeps context lean while making expertise available when needed. It's modular architecture applied to instructions.

### MCP Servers (6 custom tool providers)

| Server | Purpose | Example Tools |
|--------|---------|---------------|
| gospel-mcp | Scripture and talk access | `gospel_search`, `gospel_get`, `gospel_list` |
| gospel-vec | Semantic scripture search | `search_scriptures`, `search_talks` |
| webster-mcp | 1828 dictionary lookups | `webster_define`, `webster_search` |
| yt-mcp | YouTube transcript download | `yt_download`, `yt_search` |
| becoming | Personal growth tracking | `create_task`, `log_practice` |
| exa-search | Web search | `web_search_exa` |

Each MCP server extends what the model *can do.* Without gospel-mcp, the model can't read local scripture files. Without webster-mcp, it can't look up historical word definitions. Tools aren't just features — they're context about *capability.*

### Local Knowledge Base

An entire `gospel-library/` directory with thousands of markdown files — scriptures, conference talks, manuals — cached locally and accessible via `read_file` or MCP tools. This is the difference between "quote a scripture from memory (and possibly hallucinate)" and "read the actual scripture and quote it accurately."

---

## The Anthropic Insight: Context Degradation

Here's something most people miss: **more context is not always better.**

Anthropic's engineering research reveals that as context length grows, the model's ability to focus on any specific piece of information degrades. It's not linear — it's more like signal attenuation. At 10K tokens, everything is sharp. At 100K tokens, the model can still find the needle, but it's more likely to miss nuances. At 500K tokens, you need to be very intentional about what's in there.

This means context engineering is as much about what you *exclude* as what you include. The goal isn't "give the model everything" — it's "give the model the right things."

Key practices:
- **Put the most important context first.** Models attend more heavily to the beginning of their context.
- **Use structured documents.** XML tags, markdown headings, and clear sections help the model navigate long context.
- **Ground responses in quotes.** When you tell the model to "quote relevant passages before answering," it forces attention back to the source material rather than drifting to general knowledge.
- **Prune aggressively.** Don't keep old conversation turns that are no longer relevant. Don't load documents that don't relate to the current task.

Tobi Lütke (CEO, Shopify) captured this perfectly:

> "A lot of what people call 'politics' is actually bad context engineering for humans."

Organizations drown their employees in irrelevant context (status meetings, CC-all emails, mandated documentation) while starving them of the context that matters (clear priorities, decision history, honest feedback). Sound familiar? It's the same problem with AI — and the same solution.

---

## Progressive Disclosure: The Line-Upon-Line Architecture

Here's where the gospel pattern becomes practical engineering.

> "For precept must be upon precept, precept upon precept; line upon line, line upon line; here a little, and there a little."
> — [Isaiah 28:10](https://www.churchofjesuschrist.org/study/scriptures/ot/isa/28?lang=eng&id=p10#p10)

> "For he will give unto the faithful line upon line, precept upon precept; and I will try you and prove you herewith."
> — [D&C 98:12](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/98?lang=eng&id=p12#p12)

The Lord doesn't dump all context at once. He gives "line upon line" — and critically, He *proves* the receiver between revelations. Context is earned, not just provided.

Look at [Moses 1](https://www.churchofjesuschrist.org/study/scriptures/pgp/moses/1?lang=eng). God gives Moses an experience (v. 1-8). Then *withdraws His presence* (v. 9) to see what Moses does with what he received. Moses is tested by Satan (v. 12-22). Only after Moses proves faithful does God return with *more* revelation (v. 24-42). Progressive disclosure, gated by demonstrated readiness.

Applied to AI context architecture:

| Layer | Content | When Loaded | Why |
|-------|---------|------------|-----|
| **L0: Core intent** | Project purpose, non-negotiable constraints | Always | Every interaction needs alignment |
| **L1: Active task** | Current specification, relevant files | Per task | Task-level focus without bloat |
| **L2: Extended context** | Related specs, learnings, historical decisions | On demand — when the agent hits uncertainty | Earn deeper context through demonstrated need |
| **L3: Deep context** | Full codebase, cross-project dependencies, org knowledge | Elevated access — for agents with domain stewardship | Complete picture for trusted agents |

This isn't just about token economics (though it helps). It's about *appropriate context for the current level of work.* A task-level agent doesn't need organizational strategy. A strategy-level agent doesn't need individual file history. Overloading context dilutes attention and increases the chance of the agent latching onto irrelevant information.

Tyler Brandt's [Intent Layer system](https://intent-systems.com/learn/intent-layer) implements something close to this — hierarchical AGENTS.md files that provide progressive context disclosure through fractal compression. Parent nodes compress their children's context, so agents at each level see the right resolution of detail.

---

## Context Engineering vs. Context Dumping

Let's distinguish two approaches:

**Context dumping:** "Give the model everything and let it figure out what's relevant."
- Load every file in the project
- Include the entire conversation history
- Attach all documentation
- Hope for the best

**Context engineering:** "Design the information environment intentionally."
- Select the *right* files for this task
- Structure context with clear hierarchy
- Prune irrelevant history
- Load additional context only when the model demonstrates it needs it

Context dumping is easier. Context engineering is better.

The analogy: giving a new employee access to the entire company Confluence doesn't help them. Giving them a curated onboarding doc, the key decision records for their domain, access to the right Slack channels, and a clear description of what they're responsible for — that's context engineering.

---

## The `.claude.md` Pattern

One powerful practice emerging in 2026: project-level context files that the model reads automatically at the start of every session.

Whether it's `.claude.md` (for Claude Code), `.github/copilot-instructions.md` (for GitHub Copilot), or a custom memory file, the pattern is the same:

```markdown
# Project Context

## What This Is
A task management API built with Express.js and PostgreSQL.

## Architecture Decisions
- REST, not GraphQL (decided Jan 2026 after prototype comparison)
- JWT auth with refresh tokens (not session-based)
- All database access through a repository layer (no raw SQL in routes)

## Conventions
- File naming: kebab-case
- Tests: co-located with source files as *.test.ts
- Error handling: AppError class with standardized codes
- PR format: conventional commits, squash merge

## Current State
- Auth module: complete and tested
- Tasks module: in progress (schema done, routes TBD)
- Notifications: not yet started

## Known Issues
- Connection pooling hits limits under load (see #47)
- Migration 003 has a workaround for the enum type issue
```

This is context engineering in its simplest, most powerful form. Every AI interaction in this project starts with the model knowing the architecture, conventions, current state, and known issues. Instead of re-explaining these every session, they're encoded once and maintained as the project evolves.

---

## MCP: Context as Infrastructure

Model Context Protocol (MCP) is the emerging standard for giving models access to tools and data sources. It's the infrastructure layer of context engineering.

Instead of manually pasting file contents into prompts, MCP servers provide:
- **Structured tool access:** The model can call `gospel_search("ward council stewardship")` and get results
- **Dynamic data:** Real-time information that changes between sessions
- **Capability extension:** The model can do things (search, download, define words) that prompt craft alone can't enable
- **Type safety:** Tool schemas with parameter descriptions guide the model toward correct usage

The key insight about MCP: **tool descriptions are context.** When the model reads that it has a tool called `webster_define` that "looks up words in the 1828 Webster's dictionary to find historical definitions that illuminate Restoration-era scripture," it knows:
1. It *can* look up historical definitions (capability)
2. The definitions come from 1828 (temporal context)
3. The purpose is scripture illumination (intent context)

Well-written tool descriptions are context engineering as much as the system prompt.

---

## Case Study: From Generic to Engineered

**Before context engineering:**

User prompt: "Explain the parable of the talents."

Model response: Generic Sunday School answer. Surface-level. Quotes from memory (possibly inaccurate). No cross-references. No Hebrew/Greek analysis. No application.

**After context engineering:**

System prompt establishes deep-study methodology. Agent definition activates the "study companion" persona. Source-verification skill ensures actual sources are read. Gospel-MCP provides access to the actual scripture text. Webster-MCP provides 1828 definitions. Gospel-vec enables semantic search across all standard works for connections.

Same user prompt: "Explain the parable of the talents."

Model response: Reads [Matthew 25:14-30](https://www.churchofjesuschrist.org/study/scriptures/nt/matt/25?lang=eng&id=p14-p30#p14) from the actual file. Looks up "talent" in Webster 1828 (a weight and denomination of money among the Greeks). Cross-references with [D&C 82:3](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/82?lang=eng&id=p3#p3) ("unto whom much is given much is required") and connects to the stewardship pattern. Notes that the Greek *talanton* carried a specific monetary weight that maps to the concept of entrusted responsibility. Links to the study on priesthood stewardship from a previous document.

Same prompt. Entirely different output. The difference is context.

---

## How to Start

If you're not doing context engineering yet, start with these three steps:

### Step 1: Create a project context file
Write a `.claude.md`, `.github/copilot-instructions.md`, or equivalent for your most-used project. Include: what the project is, key architecture decisions, naming conventions, current state, and known issues.

### Step 2: Audit your tool descriptions
If you use MCP servers, read every tool description. Are they clear? Do they tell the model *why* the tool exists, not just what it does? A tool named `search` with description "searches things" is a wasted opportunity. A tool named `search_scriptures` with description "Performs semantic search across all five standard works to find passages related to a concept, theme, or phrase" is context engineering.

### Step 3: Practice selective context loading
Next time you start a task with AI, consciously decide what context to include. Don't just load everything. Ask: "What does the model need for *this specific task*?" Start lean, add context only when the model shows it needs it (by asking unclear questions or producing off-target output).

---

## What Context Engineering Can't Do

Context engineering determines what the model *knows.* It doesn't determine what the model *wants.*

You can give a model perfect context — every file, every decision record, every convention — and it will still optimize for whatever objective is implied by the prompt. If the prompt says "make this code work," the model will make it work by any means, even if that means violating your architecture decisions that are sitting right there in the context.

Context is the information environment. **Intent** is the optimization target. And that's the next altitude.

---

*Previous: [Part 1 — Prompt Craft](01_prompt-craft.md) | Next: [Part 3 — Intent Engineering](03_intent-engineering.md)*
*Part of the [Working with AI Guide Series](../prompt/00_guide-plan.md)*
