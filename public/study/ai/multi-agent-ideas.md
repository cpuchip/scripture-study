# Multi-Agent Ideas: From Prompter to Orchestrator

*Ideas, not specs. Capturing the shape of where this goes next.*

---

## The Question

The [Staying Relevant](relavent.md) study ended with a conclusion: the value moved from execution to judgment. The [AI Fatigue](fatigue.md) study named the cost of not adapting: becoming a reviewer on an assembly line that never stops. Both studies pointed the same direction — *orchestration*.

So the binding question here is practical: **How do I move from one-human-one-agent prompting to a system where my judgment scales across multiple agents working in parallel?**

Not hypothetically. With what I already have.

---

## What I Already Have (More Than I Think)

This is the part that hit me when I started inventorying:

| What Exists | What It Actually Is |
|---|---|
| **brain.exe** | A working Go binary with capture → classify → store → search. Discord input, web UI, REST API, semantic search via chromem-go. *This is already the Dropbox + Sorter + Filing Cabinet from Nate's building blocks.* |
| **brain-mcp** | MCP server exposing the brain to any agent. *This is already an agent-accessible memory layer.* |
| **ibecome / becoming** | Task tracking, practice logging, reflections — the personal growth layer. *Already integrated with brain via Layer 1.* |
| **Scripture study MCP servers** | gospel-mcp, gospel-vec, webster-mcp, search-mcp — all Go, all tool-accessible. *These are specialized knowledge agents.* |
| **docs/work-with-ai guide** | An 11-step creation cycle from Intent to Zion, with spec engineering as step 4. *This is already a review framework for evaluating proposals.* |
| **VS Code agents** | study, lesson, talk, review, eval, journal, dev, ux — each with its own instructions and tool access. *These are already specialized agents with scoped authority.* |
| **.spec/proposals/** | 8 existing proposals with design docs. *This is already a proposal pipeline.* |
| **Copilot SDK** | Available in Go. Same engine as Copilot CLI. MCP integration built in. v0.1.32, technical preview. *This is the runtime bridge between brain.exe and multi-agent orchestration.* |

I've been building the pieces without seeing the whole board. The brain captures ideas. The MCP servers provide knowledge. The VS Code agents provide specialized reasoning. The work-with-ai framework provides the review standard. The Copilot SDK provides the runtime.

**What's missing isn't components. It's the wiring.**

---

## The Dark Factory Pattern (Personal Scale)

The "dark factory" comes from manufacturing — a factory that runs with the lights off because no humans are on the floor. In software, it means autonomous development pipelines where AI agents handle implementation, testing, review, and deployment while humans focus on specification and verification.

At enterprise scale (StrongDM, for example), this looks like: agents working in parallel on isolated tasks, weekend-to-Monday pipelines, $1000/day in tokens per engineer. The benchmark numbers are real.

But I'm not StrongDM. I'm one person with a NAS, some Go binaries, and a lot of ideas. The question is whether the pattern scales *down* — and I think it does, because the core insight isn't about tokens or compute. It's about *separation of concerns*:

**Three human roles in a dark factory:**
1. **Spec Author** — Writes what should be built, with enough precision that an agent can execute unattended
2. **Scenario Designer** — Designs the test cases that verify the output (golden rule: the agent never sees the scenarios)
3. **Outcome Evaluator** — Reviews the result against intent


This maps directly to the [spec engineering guide](../../docs/work-with-ai/guide/04_spec-engineering.md) and the creation cycle's "watched until they obeyed" pattern. I've been writing about this. Now I need to *do* it.

---

## The Pipeline I'm Imagining

```
CAPTURE          SPEC              EXECUTE           VERIFY            SHIP
─────────────────────────────────────────────────────────────────────────────
brain.exe    →   proposal    →    agent(s)     →   review against   →  merge
Discord/web      .spec/proposals  Copilot SDK       scenarios           deploy
phone thought    human-written    autonomous        human judgment      git push
                 or AI-drafted    multi-step        + automated tests
```

### Stage 1: Capture → Classify → Route

Already built. brain.exe does this today. Ideas come in via Discord, web UI, or the becoming app. They get classified into categories. What changes: ideas tagged as `projects` or `actions` get a new possible route — **"spec-worthy"** — meaning they're complex enough to warrant a formal proposal rather than just a task.

### Stage 2: Idea → Proposal Draft

This is the first new piece. When I flag an idea as spec-worthy (or when brain's confidence that it's a project exceeds a threshold), the system drafts a proposal using the [spec engineering primitives](../../docs/work-with-ai/guide/04_spec-engineering.md):

1. Self-contained problem statement
2. Success criteria (observable, testable)
3. Constraints and boundaries
4. Prior art / related work (brain can search its own memory + the study corpus)
5. Proposed approach (optional — the spec author can leave this for the executing agent)

The draft lands in `.spec/proposals/` or a new `brain-proposals/` directory. I review, refine, or reject. **The AI drafts; I decide.** This is the judgment layer the Staying Relevant study identified.

### Stage 3: Proposal → Execution

This is where the Copilot SDK becomes critical. Today I work with one agent at a time in VS Code. The SDK opens the door to:

- **Programmatic agent invocation** — brain.exe (Go) can call the Copilot SDK to spin up an agent against a proposal spec
- **MCP tool access** — the executing agent gets the same tools my VS Code agents have (gospel-vec, brain-mcp, etc.)
- **Multi-step autonomy** — the agent reads the spec, plans its work, executes, and reports back. I'm not in the loop for every keystroke.
- **Parallel execution** — multiple proposals can run simultaneously on different branches

The Copilot CLI already does this at a smaller scale with its built-in agents (plan, task, code-review, explore). The SDK lets me embed the same capability into my own pipeline.

### Stage 4: Execution → Verification

The dark factory's golden rule: **the agent never sees the test scenarios.** If it sees the test, it games the system.

For code changes, this means:
- I write the acceptance criteria *before* the agent starts
- The agent's output gets tested against criteria it never saw
- Automated tests run first; I review what passes

For study/content work, the pattern adapts:
- The spec defines what the study should answer
- The agent produces the study
- I evaluate whether it *actually* answered the question vs. wandered
- The study-exp1 workflow's critical analysis phase already does a version of this

### Stage 5: Merge → Deploy

For code: git merge, CI/CD, deploy to Dokploy (or the new server).
For content: publish script, commit, push.
For brain entries: archive, cross-link, surface to ibecome.

---

## The Copilot SDK — Why It's the Right Bridge

The SDK is in technical preview (v0.1.32, January 2026). Here's why it fits:

- **Go support.** brain.exe is Go. The MCP servers are Go. No language mismatch.
- **Same engine as Copilot CLI.** The agent capabilities I already use in VS Code — planning, tool invocation, file editing, multi-step reasoning — are available programmatically.
- **MCP integration built in.** The SDK natively connects to MCP servers. brain-mcp, gospel-vec, webster-mcp — all accessible from an SDK-invoked agent.
- **Multiple models.** GPT-5 mini (free tier), Claude Haiku 4.5, Gemini Flash — choose by task. brain.exe already has model switching (`gpt-mini`, `haiku`, `flash`).
- **Streaming and tool definitions.** The SDK handles the plumbing — I focus on the spec and the tools.

**What this means practically:** brain.exe could, in Go, take a proposal from `.spec/proposals/`, invoke a Copilot SDK agent with the right MCP tools, let it work, and collect the result. The infrastructure is already Go. The intelligence is already accessible. The gap is the orchestration logic — and that's *my* logic, not boilerplate.

---

## The Server Question

Right now: local NAS, Dokploy, firewall pokes. It works but it's fragile and limited.

What I'm thinking:
- **Dedicated server** (Hetzner, maybe — or something local if the budget allows)
- **Always-on brain.exe** — not just when my machine is running
- **Agent execution environment** — a place where proposal pipelines can run unattended
- **Proper CI/CD** — git push triggers builds, tests, deploys
- **Self-hosted but accessible** — ibeco.me connects from anywhere, brain captures from anywhere

The Garvis proposal (`.spec/proposals/second-brain-architecture.md`) already laid out a phased server plan. The multi-agent layer builds *on top* of that — you need the always-on brain before you can have always-on agents.

**Priority order:**
1. Get brain.exe running on a real server (always-on capture + classification)
2. Add proposal drafting (brain detects spec-worthy ideas, drafts initial proposals)
3. Add SDK-based execution (proposals → autonomous agent work → PR)
4. Add verification pipeline (automated tests + human review)

---

## Connecting to the Creation Cycle

The [work-with-ai guide](../../docs/work-with-ai/guide/05_complete-cycle.md) maps an 11-step creation cycle from the gospel:

```
1. INTENT           — Why are we doing this?
2. COVENANT         — What are the rules of engagement?
3. STEWARDSHIP      — Who owns what?
4. SPIRITUAL CREATION — The spec (before execution)
5. LINE UPON LINE   — Iterative refinement
6. PHYSICAL CREATION — Let the agents work
7. REVIEW           — "Watched until they obeyed"
8. ATONEMENT        — Error recovery, reconciliation
9. SABBATH          — Rest, reflect, don't optimize endlessly
10. CONSECRATION    — Share what works
11. ZION            — Systems that serve everyone
```

The pipeline I'm describing maps onto steps 1–8 directly:

| Pipeline Stage | Creation Step |
|---|---|
| Capture (brain.exe) | **Intent** — a thought exists, it matters enough to capture |
| Classify + Route | **Stewardship** — assigning ownership to a category/project |
| Proposal Draft | **Spiritual Creation** — the spec before the thing |
| Agent Execution | **Physical Creation** — "let us go down and form these things" |
| Verification | **Review** — "watched until they obeyed" |
| Error handling / iteration | **Atonement** — making wrong things right |
| Reflection / journaling | **Sabbath** — stepping back to see what was built |
| Publishing / sharing | **Consecration** — making the work available |

What strikes me: I've been *writing about this pattern* for weeks. The spec engineering guide practically describes the dark factory. The creation cycle practically describes the pipeline. I just hadn't connected them to a *running system*.

---

## What Overwhelms Me (Naming It)

- The gap between "ideas" and "running system" feels enormous
- I don't know the Copilot SDK well enough yet — it's technical preview, docs are thin
- Server hosting means money and maintenance
- I'm already in sprint cycles with brain.exe and ibecome — adding orchestration feels like adding another assembly line
- The Garvis proposal is from March 1 and I haven't started Phase 1 yet
- I keep building pieces without finishing the whole

## What Calms Me Down (Also Naming It)

- The pieces *already exist*. brain.exe works. MCP servers work. The guide exists. The proposals directory exists.
- The Copilot SDK is Go. I don't have to learn a new language to use it.
- I don't have to build the whole dark factory at once. Step 1 is just: brain.exe on a server. Step 2 is just: auto-draft a proposal. Each step is a small, testable addition.
- The creation cycle says "line upon line." Not "everything at once."
- The [fatigue study](fatigue.md) already has the answer to overwhelm: morning thinking time, three-prompt rule, protect the craft. Don't let the vision of the system consume the joy of building it.

---

## Next Concrete Steps

These aren't specs. They're breadcrumbs.

1. **Get brain.exe on a server.** Always-on. Accessible from phone and laptop. This is the Garvis Phase 1 that's been waiting.
2. **Try the Copilot SDK in Go.** Build a tiny proof-of-concept: invoke an agent, give it one MCP tool, have it do a simple task. Learn the API surface.
3. **Wire "spec-worthy" routing in brain.** When an idea looks like a project, auto-create a proposal skeleton in `.spec/proposals/`.
4. **Draft one proposal via agent.** Take a real idea from brain, have an agent flesh it out into a full spec using the spec engineering primitives. Review it. See if it's any good.
5. **Execute one proposal via agent.** Give the SDK agent a reviewed spec and let it work. Evaluate the result against criteria it didn't see.
6. **Write up what I learn.** Each of these steps is a study. The process *is* the learning.

When any of these crystallize, they become real proposals in `.spec/proposals/`. For now, they're ideas. And that's fine.

---

## A Gospel Lens

> "Organize yourselves; prepare every needful thing; and establish a house, even a house of prayer, a house of fasting, a house of faith, a house of learning, a house of glory, a house of order, a house of God." — [D&C 88:119](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/88?lang=eng&id=p119#p119)

Organize. Prepare. Establish. The Lord doesn't say "build everything at once." He says "organize yourselves" — get the pieces in order. "Prepare every needful thing" — not every possible thing, every *needful* thing. Then "establish" — make it real, make it durable, make it a house.

The multi-agent vision is a house. brain.exe is a room. The MCP servers are rooms. The study system is a room. The proposal pipeline is a room I haven't built yet. But I don't build a house by staring at the blueprint feeling overwhelmed. I build it room by room.

> "For which of you, intending to build a tower, sitteth not down first, and counteth the cost, whether he have sufficient to finish it?" — [Luke 14:28](https://www.churchofjesuschrist.org/study/scriptures/nt/luke/14?lang=eng&id=p28#p28)

Count the cost. Don't pretend it's free. A server costs money. The SDK has a learning curve. Time is finite. But also: count the *assets*. I have a working brain binary. I have MCP servers. I have a framework for evaluating proposals. I have 18 years of judgment about what actually works in production. The cost is real, but the foundation is already laid.

> "And see that all these things are done in wisdom and order; for it is not requisite that a man should run faster than he has strength." — [Mosiah 4:27](https://www.churchofjesuschrist.org/study/scriptures/bofm/mosiah/4?lang=eng&id=p27#p27)

Line upon line. Room by room. Not faster than I have strength.

---

*Ideas captured: March 11, 2026*
*Related: [Staying Relevant](relavent.md) · [AI Fatigue](fatigue.md) · [Responsible Use](../ai-responsible-use.md) · [Spec Engineering](../../docs/work-with-ai/guide/04_spec-engineering.md) · [The Complete Cycle](../../docs/work-with-ai/guide/05_complete-cycle.md) · [Second Brain Architecture](../../.spec/proposals/second-brain-architecture.md)*
