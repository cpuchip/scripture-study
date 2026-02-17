# Planning Then Creating: How to Work with AI Effectively

**Series:** Working with AI — Part 1 of 3
**Duration:** 30 minutes
**Audience:** Software engineers adopting AI-assisted development (VS Code + GitHub Copilot / Cursor)
**Date:** February 19, 2026

---

## Series Overview

| Part | Title | Focus |
|------|-------|-------|
| **1** | **Planning Then Creating** | Why specification-driven development matters more than ever with AI |
| 2 | The Feedback Loop | How to review, steer, and iterate — you as the architect, AI as the builder |
| 3 | Live Build | Build something real together, start to finish, applying the patterns |

---

## Part 1: Planning Then Creating (30 min)

### Opening (3 min)

**The shift that happened.** In September 2025, Claude Sonnet 4.5 released. Then Opus 4.5 in November. Then Claude Opus 4 in February 2026. Something fundamental changed — AI went from "autocomplete on steroids" to "writes 99% of the code if you know how to direct it."

The engineers who are thriving aren't the ones who type the fastest. They're the ones who *think the clearest*. Because the bottleneck moved. It used to be: "Can I implement this?" Now it's: "Can I describe what I want clearly enough?"

**The question this lesson answers:** How do you go from a vague idea to a working system, using AI as your primary implementer?

---

### The Problem with Vibe Coding (5 min)

**"Vibe coding"** = opening a chat, typing "build me a todo app," and hoping for the best.

It works for trivial things. It fails catastrophically for anything complex. Why?

- **No shared context.** The AI doesn't know your codebase, your constraints, your architecture decisions. It guesses.
- **No memory between sessions.** Every new chat starts from zero unless you build context deliberately.
- **No specification = no verification.** If you didn't define what "done" looks like, how do you know when you're done?
- **Compounding errors.** Small misunderstandings in the AI's first output become structural problems by the time you notice.

**Analogy:** Imagine hiring a brilliant contractor and saying "build me a house." No blueprints. No site plan. No materials list. They'll build *something*. It probably won't be what you wanted.

> Every complex creation needs a blueprint before construction begins.

---

### The Pattern: Spec-Driven Development (10 min)

The pattern that works consistently — whether you're building a personal project or a production system:

#### Phase 1: Envision
*Before you touch any code or any AI tool.*

- What problem are you solving?
- Who is it for?
- What does "done" look like?
- What are the constraints (language, framework, existing systems)?

Write this down. Even 5 bullet points. The act of writing forces clarity.

#### Phase 2: Specify (The "Spiritual Creation")
*This is where the AI becomes incredibly powerful — as a planning partner.*

Create a planning document. Not code. A *spec*. Use the AI to help you think through:

- **Architecture:** What are the components? How do they connect?
- **Data model:** What entities exist? What are the relationships?
- **API surface:** What endpoints/functions? What inputs/outputs?
- **Edge cases:** What could go wrong? What are the boundaries?
- **Dependencies:** What exists already that we can use?

**Example from real work:**

```markdown
# Becoming App — Architecture Plan

## The Problem
We have study documents with "Become" commitments.
No tracking. No daily practice integration. No spaced repetition.

## Two Apps, One Backend
- App 1: Daily Practice Tracker (Vue 3)
- App 2: Study Reader (Vue 3)
- Backend: Go API + SQLite + MCP server

## Architecture
┌──────────────────────┐
│   Vue 3 Frontend     │
│  ┌────────┐ ┌──────┐ │
│  │Become  │ │Study │ │
│  └───┬────┘ └──┬───┘ │
│      └────┬────┘     │
│           │ HTTP     │
└───────────┼──────────┘
            │
┌───────────┼──────────┐
│    Go Backend        │
│  REST API + MCP      │
│  SQLite Database     │
└──────────────────────┘
```

This document was 743 lines long before a single line of code was written. It specified:
- Every API endpoint
- Every database table
- Every UI component
- The phased rollout plan
- What could be deferred

**The result:** When we said "implement Phase 1," the AI had full context. It knew the data model, the tech stack, the API surface, and how the current phase fit into the larger plan. The code it produced was architecturally coherent from the start.

**How far does this go?** In one session, the instruction was five words: "Lets build sprint 2!" That's it. The spec already existed. The AI produced 1,112 lines of correct code across 13 files — a complete public reader with short links, share modals, save-to-library flows, and database migrations. TypeScript type-checked clean. Go compiled clean. Zero errors. Zero corrections. Committed and pushed in a single session. That's the payoff of the spiritual creation. Five words worked because 743 lines of planning preceded them.

#### Phase 3: Build (The "Physical Creation")
*Now the AI writes code. But you already have the blueprint.*

- Work in phases. Don't try to build everything at once.
- Each phase should be independently testable.
- Point the AI at the spec: "Implement the REST API from the plan document."
- The spec acts as a contract — you can verify the output against it.
- **Structure sessions like sprints.** Create a todo list at the start of each session — the deliverables, in order. Mark items in-progress before starting, completed when done. This prevents the most common failure mode of complex AI sessions: losing track of where you are. The todo list is a miniature spec for the session itself.

#### Phase 4: Watch and Steer
*The most important phase. This is your job now.*

- **Review every output.** Not just "does it compile" — does it match the spec?
- **Catch drift early.** AI will make reasonable-sounding choices that diverge from your intent. The spec is your anchor.
- **Iterate fast.** Point out the divergence, explain why, let the AI correct.
- **Update the spec.** If you learn something during implementation that changes the plan, update the plan document first, then update the code.

---

### Your New Role: Architect, Not Typist (5 min)

**What you spend your time on now:**

| Before AI | With AI |
|-----------|---------|
| Writing code | Defining specifications |
| Debugging syntax | Reviewing architecture |
| Looking up API docs | Describing intent clearly |
| Implementing boilerplate | Steering and correcting |
| Typing | Thinking |

**The skills that matter more than ever:**
1. **System design** — understanding how components fit together
2. **Clear communication** — describing what you want precisely
3. **Pattern recognition** — spotting when the AI's output drifts from intent
4. **Domain knowledge** — knowing what the software *should* do
5. **Taste** — knowing the difference between "works" and "works well"

**The skills that matter less:**
- Memorizing syntax
- Typing speed
- Knowing every library API by heart
- Writing boilerplate from scratch

This isn't about being replaced. It's about *leverage*. A single engineer with AI and a good spec can now do what used to take a team of five. But only if they know how to direct the work.

---

### Practical Tips (5 min)

**1. Start every project with a doc, not a file.**
Create `docs/plan.md` before `src/main.go`. Describe the system in plain language. Let the AI help you refine it.

**2. Give the AI your codebase context.**
In VS Code with Copilot, the AI can see your workspace. Use `@workspace` to reference existing code. The more context it has, the better its output matches your architecture.

**3. Use the AI to search and understand before you build.**
Before implementing a feature, ask the AI to read the relevant existing code and summarize how it works. This prevents the AI from reimplementing something that already exists.

**4. Keep planning docs updated.**
The spec is a living document. When you make implementation decisions that change the plan, update the plan. Future sessions (yours and the AI's) will benefit from accurate documentation.

**5. Conversation summaries are your persistence layer.**
When a session ends mid-project, the conversation summary captures exactly where you are — which files exist, what's been modified, what remains. The next session picks up precisely where the last one left off. The spec says *what to build*. The summary says *where you are in building it*. Together, they make multi-session projects seamless. I've had sessions start with the AI's first action being the exact next item on the todo list — zero re-orientation, zero wasted effort — because the summary carried the full state forward.

**5. Work in small, verifiable increments.**
Don't ask for 500 lines at once. Ask for one function, verify it, then move to the next. The spec tells you the order.

**6. Treat the AI's output as a draft, not a delivery.**
Review everything. The AI is a brilliant first-drafter. You are the editor. Your domain knowledge, your taste, your understanding of the user — these are what turn a draft into a product.

---

### Wrap-Up and Preview (2 min)

**The pattern:**
1. Envision — what are we building and why?
2. Specify — plan document before code
3. Build — AI implements against the spec
4. Watch and steer — review, correct, iterate

**Next session: The Feedback Loop**
- How to review AI-generated code effectively
- How to give corrections that stick
- How to handle the AI getting stuck or going in circles
- Live examples of steering a session back on track

**Homework:**
Pick a small project you want to build. Write a 1-page planning doc describing:
- What it does
- What the components are
- What the data model looks like

Don't write any code. Bring the doc to the next session.

---

## Facilitator Notes

### Key Points to Emphasize
- **This is about leverage, not replacement.** Engineers who learn to direct AI effectively are dramatically more productive. Engineers who resist or ignore it will fall behind.
- **The spec is the product.** In AI-assisted development, the quality of the specification determines the quality of the output. Garbage spec → garbage code.
- **Taste still matters.** AI produces syntactically correct, functionally plausible code. Knowing whether that code is *good* requires engineering judgment that only comes from experience.

### Common Objections

| Objection | Response |
|-----------|----------|
| "I'm faster just writing it myself" | For a single function, maybe. For a whole feature? The AI will outpace you 10:1 if you've given it a good spec. |
| "I can't trust AI-generated code" | That's why you review it. The spec gives you a contract to verify against. |
| "What about security/quality?" | Same review process as a junior developer's PR. You wouldn't merge unreviewed code from a junior — don't merge unreviewed AI code either. |
| "This feels like more work upfront" | It is. And it pays back 5x during implementation. Same as writing tests — investment now, dividends later. |

### Live Demo Ideas
- Show a real planning doc from your own work
- Open VS Code, show how Copilot Chat uses workspace context
- Do a live mini-spec for something simple (a CLI tool, a small API)
- Show the before/after: vibe-coded attempt vs. spec-driven attempt

### Series Roadmap
- **Part 2: The Feedback Loop** — reviewing, steering, iterating. How to handle drift, hallucinations, and getting stuck.
- **Part 3: Live Build** — pick a project from someone's homework, build it together live using the full pattern.
