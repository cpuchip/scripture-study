# Part 7: From Developer to Steward — Goal-Based AI Orchestration

**Series:** Working with AI — A Comprehensive Guide
**Date:** April 2026
**Prior work:** Parts 1-6, brain.exe Phase 3c, Simon Scrapes "Agentic OS Command Center"
**Core thesis:** The industry's AI tooling is developer-focused — sessions, terminals, code output. The next maturation is steward-focused — goals, projects, outcomes. The shift isn't about different tools. It's about a different relationship with the work.

---

## Two Paradigms

Most AI-assisted work today follows a **developer-focused** paradigm:

- The IDE is the primary surface
- Work is organized by sessions (start conversation, get output, end)
- Progress is measured in code artifacts (commits, PRs, test results)
- Context resets between sessions
- The human drives every interaction

This works. Parts 1-6 of this guide are written from within this paradigm. Prompt craft, context engineering, intent engineering, specification engineering, the complete cycle, enterprise architecture — all of it assumes a human initiating work and an AI responding.

But there's a ceiling.

The ceiling isn't technical. It's organizational. When you have 11 agent modes, 13 skills, 7 MCP servers, a maturity pipeline, and projects spanning scripture study, teaching, a Space Center, and a Sunday School presidency — the developer paradigm breaks. Not because the tools fail, but because the *human* becomes the bottleneck in a system that was supposed to reduce bottlenecks.

The symptom: 45 entries in an approval queue, all equally undifferentiated. Well-planned tasks in spec files but a jumbled mess in the daily workflow. Living in VS Code because that's where the agents are, even when the work isn't code.

---

## The Steward Paradigm

The alternative is **steward-focused, goal-based orchestration:**

| Dimension | Developer-Focused | Steward Goal-Based |
|-----------|-------------------|---------------------|
| **Primary surface** | IDE / terminal | Dashboard / brain |
| **Work unit** | Session | Goal |
| **Organization** | Files and folders | Projects with intent |
| **Agent interaction** | Request → response | Iterative turns ("Your Turn / Agent's Turn") |
| **Progress measure** | Artifacts produced | Goals advanced |
| **Context** | Resets per session | Persists per goal, per project |
| **Agent outputs** | Inline in conversation | Files in project directories (durable, diffable, versionable) |
| **Scheduling** | Human-initiated | Cadence-based (research passes, digests, reviews) |
| **Visibility** | Grep the codebase | Dashboard with project cards and activity feeds |

This isn't a replacement for the developer paradigm. It's a maturation. You learn the tools in developer mode. You orchestrate outcomes in steward mode. Most people will use both — developer mode for coding, steward mode for managing the broader portfolio of work.

### The Turns Model

The core UX insight comes from the concept of iterative agent turns — the "Your Turn / Agent's Turn" pattern.

In developer mode, agent interaction is sequential: you ask → agent responds → session ends. If the output isn't right, you start a new session with corrective context. The context chain is fragile.

In steward mode, interaction is iterative: you define a goal → agent works → agent pauses with output → you review and give feedback → agent continues from where it left off → repeat until the goal is met. The session persists. The context accumulates. The conversation has a natural rhythm.

This maps to how delegation works with people. You don't give someone a task and expect perfection on the first pass. You give them the goal, they do initial work, you review, they refine. The quality comes from the iteration, not from the initial specification being perfect.

Theologically, this is the Abraham 4:18 pattern — "watched those things which they had ordered, until they obeyed." The watching IS the work. The steward doesn't fire and forget. The steward engages through the turns until the outcome is right.

### Projects as First-Class Entities

In developer mode, organization happens through the filesystem — directories, config files, workspace structure. The human maintains the mental map.

In steward mode, projects are explicit entities: named, described, with status tracking, associated entries, goals, and agent context. A project carries its own intent (like `intent.yaml` at the workspace level but scoped to the project). Agents inherit project context when working on entries for that project.

This means you can ask "Show me everything for Sunday School" and get a coherent view — lessons in progress, presidency coordination items, scheduled prep tasks, agent research outputs — all grouped under one project.

### Filesystem-Based Outputs

Developer mode outputs live in conversation transcripts or inline diffs. They're ephemeral by default.

Steward mode outputs are files. Agent research lands in `{project}/scratch/`. Deliverables land in `{project}/outputs/`. Plans and specs land in `.spec/proposals/`. Everything is readable without the agent, diffable across versions, and durable across context compaction.

This is the "files are durable, context is not" principle applied to agent outputs, not just human notes.

### Scheduled Work

Developer mode is entirely pull-based: the human initiates every interaction.

Steward mode adds push-based cadences: a weekly research pass that surfaces new articles and videos, a Monday morning digest that summarizes what's in flight, a Friday review that asks "what did we achieve this week?" These don't replace human-initiated work — they prepare the ground so the human's time is higher-leverage.

---

## The Maturation Path

This isn't a binary switch. It's a progression:

**Stage 1 — Learning the tools.** Single-agent interactions. Prompt craft. Getting comfortable with what AI can do. (Parts 1-2 of this guide.)

**Stage 2 — Engineering the context.** Multi-file context, skills, MCP tools. The AI knows enough to be genuinely useful. (Parts 2-4.)

**Stage 3 — Specifying autonomous work.** Specs precise enough for agents to execute independently. Phased delivery. Review criteria. (Parts 4-5.)

**Stage 4 — Orchestrating outcomes.** Multiple projects, multiple agents, goal-based tracking, iterative turns, scheduled work, project-scoped context. The human is a steward, not a typist. (This part.)

**Stage 5 — Scaling stewardship.** Multiple stewards, shared resources, organizational intent cascading through layers. (Part 6 — enterprise architecture.)

The mistake is jumping to Stage 4 or 5 without the foundation. If your prompts are vague, project-level orchestration just means vague prompts at scale. If your context engineering is poor, scheduled tasks produce garbage on a cadence.

But the mistake of staying at Stage 2-3 forever is equally real. At some point, the developer paradigm becomes the constraint. You're spending more time managing the system than doing the work the system was supposed to help with.

---

## What This Looks Like in Practice

A concrete daily flow in steward mode:

**Morning:** Open the dashboard. See projects with their status. "Sunday School" has a prep task — the scheduled agent ran overnight and pulled next week's Come Follow Me block into a study entry. "Teaching" has an agent-researched outline waiting for review. "Brain Development" has three entries in the approval queue.

**Review:** Click into the Sunday School prep. The agent's research is in `projects/sunday-school/scratch/`. Read it. Reply with feedback: "Focus more on the Alma 32 connection, less on the Moroni tie-in." Agent takes another turn.

**Approve:** Route the three Brain Development entries. Two are small enough for auto-processing (Haiku-level research passes). One needs planning — advance it to `planned` stage and assign to the plan agent.

**Work:** Switch to VS Code for actual coding if that's what's needed. Brain handles the project management. VS Code handles the code.

**Evening:** Quick capture via brain-app — a thought about the Space Center that came during dinner. It lands in brain, gets classified, gets assigned to the Space Center project. It'll be there tomorrow, contextualized and ready.

---

## The Relationship Between Paradigms

The developer paradigm and the steward paradigm aren't competitors. They're layers:

```
┌─────────────────────────────────────────┐
│  Steward Layer (Goals, Projects, Turns) │
│  ┌───────────────────────────────────┐  │
│  │  Developer Layer (IDE, Sessions)  │  │
│  │  ┌────────────────────────────┐   │  │
│  │  │  Agent Layer (Skills, MCP) │   │  │
│  │  └────────────────────────────┘   │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

The steward layer orchestrates. The developer layer executes. The agent layer powers both. You don't abandon VS Code — you stop using it as a project management tool and let it be what it's best at: a coding environment.

This is the "work of salvation and exaltation" pattern from the Church's organizational structure (Part 6). The First Presidency doesn't write code. They set intent. Stakes don't manage individual lessons. They coordinate wards. Each layer does what it does best and trusts the layers below.

The difference between a developer using AI and a steward orchestrating AI is the same difference between a programmer and a technical lead. The programmer writes code. The lead ensures the right code gets written toward the right goals. Both roles are necessary. The maturation is learning when to be which.
