---
batch: DW
title: Directing work at the substrate — discoverability, brainstorm fusion, project intents, guided intake
status: idea (stub — needs council)
proposed_by: michael
proposed_on: 2026-05-17
preceded_by:
  - substrate-batch-j-fanout-brainstorm.md
  - stewards-ui-evolution.md
  - substrate-pipelines-expansion.md
links:
  - "proposals/stewards-ui-evolution.md"
  - "proposals/substrate-pipelines-expansion.md"
---

# Directing work at the substrate

> **Status: idea stub.** Captured 2026-05-17 from Michael's framing while trying
> to start new work items. Four threads, one root: **it is hard to aim work at
> the substrate.** Not yet councilled — decisions below need walking.

## The root problem

The substrate can run pipelines, but starting a work item assumes you already
know which pipeline does what, which intent frames it, and what settings it
needs. Michael, the substrate's own author, hit friction trying to start work.
That is the signal. Four threads address it.

## T1 — Pipeline discoverability

Pipelines expose names, not purpose. The New Work flow should show, per
pipeline: a plain-language **what this does**, a **when to use it**, and an
**example output**. Today you cannot tell `research-write` from
`decompose-fanout` from `brainstorm` without reading SQL.

- Decision: add description / use-when / example-output fields to the pipeline
  row; surface them in the New Work UI (overlaps `stewards-ui-evolution`).

## T2 — `agent-brainstorm` (fuse fan-out + brainstorm)

Batch J shipped a **fan-out** pipeline and a **brainstorm** pipeline
separately. Michael wants them fused: multiple agents fan out and brainstorm
the same prompt independently, then a synthesizer agent combines the
best-of-best into one result.

- Decision: new pipeline shape, or a `synthesize` terminal stage added to the
  existing brainstorm fan-out. Folds naturally into
  `substrate-pipelines-expansion`.

## T3 — Project intents (general work)

The three current intents — a substrate-self-development intent (`5d`),
`general-research` (`h1-1`), and `planning-partner` (`h3-2`) — are
substrate-and-research-shaped. `general-research` can host research work for
any project, but **no project intent exists**: building the science center
(`space-center`), reviving `cpuchip.net`, or running a study arc each wants its
own intent so the covenant and context framing are right.

- Decision: per-project intents, a general "creative / build work" intent, or
  both. How granular. Who authors them (human, or T4's agent).

## T4 — Guided work-item intake ("council to start work")

The one Michael most wants. Instead of a form that assumes substrate fluency:
pick a project, describe the work in plain language, and an agent reviews the
description and **proposes** which intent + pipeline + settings fit — or drafts
new ones — which the human then tweaks before dispatch. A conversational
bootstrap, "like a council." Largest of the four; overlaps
`stewards-ui-evolution` heavily (it is a New Work surface).

- Decision: a chat-style intake, or a one-shot "analyze my request" gate. How
  much it is allowed to author (new pipelines? new intents?) vs. only suggest.
  Cost gate on the analysis step.

## Next step

Council these — T1 and T4 alongside `stewards-ui-evolution` (both are New Work
UI), T2 into `substrate-pipelines-expansion`, T3 standalone. The `cpuchip.net`
revival (2026-05) is the live exercise: every awkward moment aiming a real task
at the substrate is a concrete requirement for one of these four. Let the real
work name them rather than designing in the abstract.
