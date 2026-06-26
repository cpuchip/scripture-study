---
date: 2026-06-26
lane: pg-ai-stewards
topic: the orientation arc — putting "lending the substrate our orientation" into practice
tags: [orientation, harness, zion, autoload, trajectory-eval, harness-over-intelligence, council-moment]
---

# Orient → Act → Verify: lending the substrate our orientation

## How it started

A study (`study/ai/harness/lending-the-substrate-our-orientation.md`) answered Michael's
*"I feel like we're missing something"* with a precise finding: the substrate runs the
creation pattern's steps 2–11 — covenant, stewardship, watching, judges — but is thin on
**step one of each agent's loop, Orient.** Three witnesses converged: Boyd (Orient is the
irreducible OODA node), our own workspace (the Council Moment is a *hard, universal*
discipline — every agent scans before it acts), and Google's SDLC (the agent loop is
perceive→plan→act, and plan-quality + context-handling are first-class eval dimensions).
The reframe that made it actionable: **our workspace is the substrate's orientation
library** — every skill a past failure crystallized — and the substrate had been running
on almost none of it. Keep the engine thin; fill the shelf.

Then Michael said the line that set the intent: *"I feel what this is reaching towards is
Zion."* One heart and one mind (Moses 7:18) — the workspace and the substrate sharing one
orientation, our hard-won judgment consecrated into the engine so no agent is left
orientation-poor. That became the why under all three phases.

## What got built (chain 60→64, shipped, CI green)

Three moves, the study's three, each a phase, each committed and dev-tested before the next:

1. **Orient (the skill) — `62-orientation.sql`.** The shelf was opt-in (skill_load), so
   orientation stayed *dormant* — and a skill-denied agent (world-build, the subagents)
   could never receive it. Built an **autoload** layer: a skill listed for an agent-family
   glob injects as a *standing* block, unconditionally, bypassing the skill-tool permission.
   Orientation lent, not opted into. Michael ratified it into the **core baseline** (the
   Zion call): orient-first + bounded-gather join source-verification + reference-linking,
   so *every operator's* substrate is oriented out of the box — no operator left
   orientation-poor either.

2. **Orient (the tool) — `63-orient-survey.sql`.** `orient_survey` generalizes the
   reflect-steward's `intent_work_survey` from intent→project, so any builder can ask
   "what already exists here?" (docs, worlds, work) and extend rather than rebuild. The
   council moment as a tool, for everyone, not just the planner.

3. **Verify (the trajectory) — `64-auto-critique.sql`.** The trajectory critic (56) and the
   verdict→self-improvement loop (59) existed, but nothing *fired* the critic — the
   Glass-Box half sat dormant, the same shape as the shelf. Added the one missing trigger:
   a worker run finishing auto-critiques its own trajectory; the verdict harvests into the
   loop, which closes itself. Cost-safe (config gate, default OFF) and gate-safe (graders
   hard-excluded — never grade the grader). No bgworker change.

## The moment it obeyed

Michael: *"lets test it… watch until it obeys!"* — Abraham 4:18. We dispatched a real
world-build and watched the trajectory. Its first message: **"I'll start by orienting:
checking the current state of the world"** → it called `world_show`, surveyed, *then*
extracted. Not the prompt merely rendering — the agent opened with the orientation move,
the exact word from the skill, before touching the work. And a chat, grounded on the same
corpus, searched → read the source → committed a grounded answer with no spiral. The
autoloaded orientation visibly changed behavior. The watch is the proof the rendering
never is.

## What made it sound

Every phase: `lib.rs` + the Dockerfile COPY list + a virgin-smoke assert, *together* — the
exact triad whose omission left CI red for seven commits earlier this same session. Before
pushing the arc, a **fresh-image virgin-smoke** (00→64, OK 50–53) — the oracle run *before*
the push, not discovered after. Green locally → pushed → CI confirmed. The discipline held.

## Findings worth keeping

- **Dormancy is the recurring failure of a good mechanism.** The skills shelf, the
  trajectory critic — both were built well and sat unused because nothing *reached* the
  agent or *fired* the judge. The fix each time was a small activation layer (autoload; a
  completion trigger), not a smarter component. Build the mechanism; then build the thing
  that makes it standing.
- **The study's "empty shelf" was slightly wrong, and the correction mattered.** The core
  already seeded two orientation skills (source-verification, reference-linking) in
  schema.rs. That precedent turned the core-vs-example question from a hard call into an
  easy yes — there was already a baseline to join.
- **A done-signal can be gamed; a verdict can recurse.** Keep coverage separate from a
  quality judge (61 vs 56); exclude graders from auto-critique (64) so the loop never
  grades its own graders. The eval-gaming guard (59) is the same instinct.

## Carry-forward

- `auto_critique_on_complete` ships OFF — turn it on per substrate when the Glass-Box half
  should run standing (it's real LLM spend; the reflect-watchman caps it).
- The remaining moves the study named are now built (orient skill, orient tool, verify);
  the next ring is porting *more* battle-tested disciplines onto the shelf (read-before-
  quoting → digesters, inverse-hypothesis → coder) and the post-demo force-final-at-cap
  bgworker floor (`.spec/proposals/force-final-at-cap.md`).
- The arc is the first real instance of the thesis: the workspace's orientation, lent to
  the substrate, a skill at a time. The pipe is now open.
