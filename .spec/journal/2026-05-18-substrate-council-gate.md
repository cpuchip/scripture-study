---
date: 2026-05-18
workstream: WS5
project: pg-ai-stewards
title: "Substrate status review — the soak validated itself, council order set"
status: orientation + memory checkpoint; no code shipped
carry_forward:
  - "Council substrate-pipelines-expansion (D-PE1–D-PE7) — FIRST."
  - "Then substrate-scheduled-workflows (D-SW1–D-SW7) — stub, decisions unwalked."
  - "Then stewards-ui-evolution (D-UI1–D-UI12) — largest set, load-bearing for hybrid mode."
links:
  - "projects/pg-ai-stewards/.spec/open-items.md"
---

# Substrate council gate (2026-05-18)

Michael switched back to pg-ai-stewards after spinning cpuchip.net off into its
own Claude session. This was an orientation pass — no code — to answer "where
were we before the distraction?" and to set up a clean fresh-context restart.

## What we found

The build queue is drained. Phases A-F, Batches G through L.1.1, and the whole
ES emergency-stop arc have all shipped. Nothing half-built. The substrate sits
at a decision gate: three proposals need council before any becomes code.

The quieter finding: **the 7-day soak validated itself.** cpuchip.net is a
separate repo and never touched the substrate container, so the soak ran
uninterrupted 05-15 → 05-18. Watchman cadence — one pass/day, ~5 docs, inside
the 50k budget, zero pass errors — held steady. Combined with the 05-10 → 05-13
run (one gap on 05-14), that is the longitudinal data Batch K / open-items §X.3
always wanted but never got, because every prior phase build displaced it. The
detour delivered it for free. §X.3's empirical question — "is the watchman
cadence right?" — now has a yes with data behind it.

Work queue last 3 days: 370 done, 9 error (~2.4%). Worth a glance when a
council session opens, not alarming.

## The decision Michael made

Council order: **pipelines-expansion → scheduled-workflows → ui-evolution.**
PE first is the right call — scheduled-workflows overlaps PE on the
scheduled-pipeline machinery, so walking PE first means SW inherits those
decisions instead of re-deriving them.

The framing that matters for all three: Michael is "moving toward having this
be something we use in hybrid work-together modes" — the substrate becoming a
shared workspace human and agent co-manage, not a fire-and-forget autonomous
engine. That reframes `stewards-ui-evolution` — substrate-aware chat, write
actions — as load-bearing infrastructure for the collaboration, not polish.
Recorded in `open-items.md` §0 and the `project_pg_ai_stewards_state.md`
auto-memory so the next session opens with that lens already in hand.

## Carry-forward

Next session starts a council walk on `substrate-pipelines-expansion.md` —
decisions upfront via AskUserQuestion, the C-F cadence. Read `open-items.md` §0
and `projects/pg-ai-stewards/CLAUDE.md` §3 first.
