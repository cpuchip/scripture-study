---
title: How the world is using AI in 2026 — where pg-ai-stewards stands, gaps, and the human-usability problem
date: 2026-06-06
status: research + scratch + critical analysis (living)
binding_question: >
  What is the state of practical AI use in 2026, how does our work (pg-ai-stewards +
  this workspace) compare, where are the gaps — and specifically, how do we make the
  substrate genuinely useful to *Michael*, not just to the agent?
note: >
  Doubles as book material ("Beyond the Prompt"). Web-researched 2026-06-06; the field
  moves weekly — re-verify before citing externally.
---

# AI utilization landscape 2026

## §0 — Scratch (raw findings)

**Orchestration patterns now standard** (fungies, digitalapplied, addyosmani, MS Learn):
- Five dominant: **fan-out/parallel, pipeline/sequential, debate, supervisor/hierarchical, swarm** (+ handoff/routing, loop/iterate). Most production systems combine several.
- **Supervisor/hierarchical** (a planner decomposes → delegates to specialist workers) is the most common for software. ← *this is exactly pg-ai-stewards.*
- **Git-worktree isolation per agent** is the accepted fix for parallel agents clobbering files. ← *we built this (CV2.1).*
- Production shape (LangChain/Rippling): supervisor + specialized read/RAG/action agents + **observability (traces, layered evals, self-healing loop).**

**Self-hosted / power-user landscape** (infoq, knowlee, composio, augment, medium):
- **Composio Agent Orchestrator** — fleets of coding agents in parallel worktrees, each its own PR, autonomous CI-fix + review-response, **one human dashboard**. Agent-agnostic (Claude Code/Codex/Aider).
- **Coder Agents** — model-agnostic, run coding agents on your own infra (control of code/data/exec).
- Frameworks: LangGraph (orchestration) + CrewAI (role agents) + n8n/Flowise (visual). MCP as the standard interface that makes mixing work.
- Self-host motive: cost, data control, open-weight models via Ollama/vLLM.
- Vendor ROI claims: 2.5–3.5x avg, 4–6x top, best on refactor/migration/feature. *(Treat as vendor numbers.)*

**Critical / skeptical** (techahead, mindstudio, ability.ai, arxiv 2604.02547, pooya):
- **Context rot** = the make-or-break problem: past ~100K tokens, agents hallucinate / forget / go inconsistent. BUT it's a **discipline problem, not a model problem — ~79% of failures come from specs + coordination, not capability.**
- **Context *quality*, not volume**, is the limiter; most teams don't use the full window.
- Reliability: SOTA still fails >20% of SWE-bench Verified; success **drops sharply as tasks stretch from minutes to hours.** Evals miss operational qualities — consistency across runs, robustness, predictability, bounded failure.
- "AI coding frustration is a **skill gap, not model failure.**"
- **The winners are the most disciplined context + clearest specs, not the flashiest models.**

Sources: [fungies](https://fungies.io/ai-agent-orchestration-developers-guide-2026/) · [digitalapplied 5 patterns](https://www.digitalapplied.com/blog/multi-agent-orchestration-5-patterns-that-work) · [addyosmani agent orchestra](https://addyosmani.com/blog/code-agent-orchestra/) · [MS Learn patterns](https://learn.microsoft.com/en-us/azure/architecture/ai-ml/guide/ai-agent-design-patterns) · [Composio AO](https://github.com/ComposioHQ/agent-orchestrator) · [Coder Agents (InfoQ)](https://www.infoq.com/news/2026/05/coder-agents-self-hosted-ai/) · [augment orchestrators](https://www.augmentcode.com/tools/open-source-agent-orchestrators) · [Rippling/LangChain](https://www.langchain.com/blog/how-rippling-went-ai-native-across-every-product-in-6-months-with-deep-agents-and-langsmith) · [context rot (TechAhead)](https://www.techaheadcorp.com/blog/context-rot-problem/) · [context rot (MindStudio)](https://www.mindstudio.ai/blog/context-rot-ai-coding-agents-explained) · [behavioral drivers (arXiv)](https://arxiv.org/pdf/2604.02547) · [skill gap (Pooya)](https://pooya.blog/blog/ai-doesnt-code-skills-problem-not-ai-problem-2026/)

## §1 — The big picture

The field has moved from "chat with a model" to **"manage a fleet of agents."** The
human is becoming a supervisor: decompose → delegate to specialist agents in isolated
worktrees → review at gates. And the hard-won 2026 consensus is that **the bottleneck
is not model power — it's specs, coordination, and context discipline.** "79% of
failures come from specs and coordination." The teams that win are the disciplined ones.

## §2 — How we compare (honestly: ahead on the things that turn out to matter)

pg-ai-stewards is squarely in the dominant pattern and, on several axes, ahead of the
curve — because the workspace optimized for *discipline* before the field agreed that's
the lever:

- **Supervisor/hierarchical orchestration** — work_items → pipelines → specialist
  stages (plan/implement/review). ✓ the standard.
- **Worktree isolation + per-task model routing** (CV2.1; kimi/qwen/glm escalation). ✓
- **Critic/review gate** that catches what build+test certifies green (cv6/cv11) — the
  field is *just* learning that a second strong model reviewing each stage matters.
- **Ground-truth verification** (build+test as the Prescription floor) — the field's
  "grounding/retrieval, not a better model" lesson, already wired.
- **Context discipline as a first-class concern** — engram compaction (Batch K/L) and
  the CT2 self-context spec **directly target context rot, the field's #1 problem.**
- **Cost-tracking per run, the trust ladder, the Hinge, the watchman soak** — governance
  the field is bolting on late (observability, human-approval gates).
- **The differentiator: the stewardship frame.** Covenant / council / watching / the
  Hinge ("delegate execution, not discernment") / Sabbath. The field's own critical
  analysis — *discipline + specs + coordination beat model power* — is **the same claim
  the stewardship pattern has been encoding from the start.** That's not branding; the
  doctrine front-loaded the exact discipline the 2026 data says decides success.

## §3 — The gaps (critical, in priority order)

1. **Human usability — THE gap (Michael's own diagnosis, and the field confirms the
   fix).** pg-ai-stewards is *agent-facing*: it's a set of MCP tools that *I* drive.
   Michael drives me; I drive the substrate. There's `stewards-ui` (operator surface)
   but no ergonomic **human cockpit** where Michael is the manager-of-agents directly.
   The winning power-user tools (Composio AO's dashboard, Coder Agents) are built for the
   *human* to supervise the fleet. **We optimized the engine and under-built the
   driver's seat.** This is the highest-leverage gap.
2. **Observability for humans.** The field has LangSmith-style traces + layered evals +
   dashboards. We have journals + watchman + cost logs — rich, but not a human-facing
   trace/eval/cost view. Michael can't *see* a run easily.
3. **Systematic evaluation.** The field stresses consistency-across-runs, robustness,
   predictability, bounded-failure metrics. We A/B occasionally (n=1, directional). No
   eval harness — so our "it's better" claims are under-measured (the Ben-test risk).
4. **Parallel fleet vs sequential.** We have the worktree primitive but mostly run one
   thing at a time (the night-build was sequential). The field runs many agents in
   parallel under one dashboard. We have the parts, not the fleet view.
5. **CT2 not built.** Our context-discipline edge is partly on paper.

## §4 — Where we might be fooling ourselves (Ben test)

- **"We're ahead."** On *philosophy/discipline*, defensibly yes. On *engineering
  maturity, observability, and adoption*, the field's tools are more polished and
  battle-tested across many users; pg-ai-stewards has one user (+ me). Don't confuse
  "conceptually ahead" with "more capable in practice."
- **"The gospel frame is a differentiator."** True *and* it's worth noting the field is
  converging on the same operational truths from data alone. Our edge is that we got
  there earlier and hold it more consistently — not that others can't.
- **"It works."** It works *for me*. The honest signal is Michael's: it's much more
  useful to the agent than to him. A tool that needs an expert agent to drive it hasn't
  crossed the usability line for its owner.
- **ROI / velocity.** We have no measured velocity numbers — only vibes ("the BoM walk
  was worth it"). The field at least pretends to measure; we should actually measure.

## §5 — The human-usability direction (Michael's ask: an opencode/claude-cli-like tool for the substrate)

The goal: put **Michael in the driver's seat** of the substrate — dispatch work, watch
it run, approve at the Hinge, talk to personas, see cost — without going through me.
Three candidate shapes (a real fork to decide):

| Option | What it is | Pros | Cons |
|---|---|---|---|
| **A. Stewards TUI/CLI** | a `stewards` CLI (opencode/claude-code-style): `stewards do "<binding question>"`, `stewards watch`, `stewards review` (Hinge), `stewards personas`, `stewards cost` | terminal-native (his habitat), scriptable, fast, fits the `-p` worker story | another harness to build + maintain |
| **B. Extend `stewards-ui`** | grow the existing operator UI into a real cockpit: dispatch, live pipeline view, escalation/approval queue, cost + traces, brain browser | reuses what exists; visual fleet view; good for *watching* | web UI, heavier; he lives in the terminal more |
| **C. ai-chattermax AS the cockpit** | drive the substrate by **chatting** — dispatch via DM to a "foreman" persona; personas report; approve in-thread | reuses ai-chattermax (already built + live); natural; multi-device; the personas *are* the interface | conversational control is fuzzy for precise ops (review diffs, approve a deploy) |

**My read:** these aren't exclusive — the strongest answer is **B as the watch/approve
cockpit + A as the do/script surface**, with C (chat-driven) as the ambient/mobile
layer later. The fastest *first* win is a thin **stewards CLI (A)** for the verbs
Michael actually needs (`do / watch / review / cost`), because it's terminal-native,
reuses the existing MCP/work_item machinery, and turns him into a direct
manager-of-agents in an afternoon — closing the #1 gap with the least build.

This also composes with the two specs already written: [[claude-worker-dispatch]]
(Michael assigns work to *me* from the cockpit) and [[agentic-tools-model-cascade]]
(he invokes `research_codebase` etc. directly). The cockpit is the human entry point to
the whole stewardship tree.

## §6 — Recommendations / next steps

1. **Treat human-usability as the next real workstream** (it's the gap that matters most
   and the one Michael feels). Pick a cockpit shape (§5) — lean: thin **stewards CLI**
   first, `stewards-ui` for watching.
2. **Add light measurement** — even a per-run cost+outcome log Michael can read closes
   the Ben-test gap and turns "it's better" into data.
3. **Build CT2** — it's our context-rot answer, the field's #1 problem; shipping it makes
   the "ahead on discipline" claim real, not paper.
4. **For the book:** the chapter writes itself — *the field spent 2026 discovering that
   discipline, specs, and delegation beat model power; the stewardship pattern is that
   discipline, named and practiced.* This research is the evidence.

## Open questions for Michael
1. Cockpit shape — CLI-first (A), UI-first (B), or chat-first (C)? (Lean: A then B.)
2. What are the 4–6 verbs you'd actually use daily? (do / watch / review / cost / personas / brain?)
3. Is "useful to me, not you" the right problem to spend the next build on, over CT2 / the code persona?
