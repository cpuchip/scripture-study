---
date: 2026-05-17
mode: bookkeeping + memory consolidation
workstream: WS5
project: pg-ai-stewards
title: "Post-ES bookkeeping — proposal queue refreshed, ES.6.A/B archived, new scheduled-workflows idea, active.md consolidated"
status: complete. Substrate docs reconciled to reality; one new idea captured.
carry_forward:
  - "substrate-scheduled-workflows.md is a stub — D-SW1–D-SW7 need a council before any build. Pair with or follow substrate-pipelines-expansion."
  - "active.md In Flight table still carries ~12 'proposed Apr 22' rows — a real but unaudited backlog; worth a triage pass."
  - "Batch G micro-gap: failure_count_limit quarantine does not fire atonement (only cost-cap quarantine does). Tiny, recorded on the open-items shipped row."
links:
  - "../../projects/pg-ai-stewards/.spec/open-items.md"
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-scheduled-workflows.md"
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
---

# 2026-05-17 — bookkeeping caught up to the build

A status question — "where are we with pg-ai-stewards?" — surfaced that
the navigation docs had drifted well behind the work. Two days of the
ES arc, plus Batches K / L / L.1.1 before it, had shipped without the
index ever being updated. This session reconciled the docs.

## What was stale, and what it now says

**The proposal queue (`open-items.md` §0).** Listed Batch L as "ratified,
build pending" — it had shipped three days earlier. K, L.1.1, and the
entire ES arc weren't in the table at all. Refreshed: a "Shipped /
closed" table now holds I/G/J/K/L/L.1.1/ES, and the live queue is just
three items — `substrate-pipelines-expansion`, `stewards-ui-evolution`,
and the new `substrate-scheduled-workflows`.

**ES.6.A / ES.6.B.** Demoted to optional carry-forward when streaming
shipped. Reviewed with the arc closed and the soak clean — **archived**,
not carried. Each has a written revive trigger (A: review-stage timeout
recurs; B: gateway 500s recur on a non-streaming path).

**Batch G.** Marked "build-ready, no ratifications needed" — but the
`6a`–`6e` SQL files, the `stewards-cli materialize-writes` command, and
the journal `2026-05-11-substrate-batch-g-shipped.md` all show it
**shipped 2026-05-11** in 8 commits. Corrected the proposal frontmatter
and moved it to the shipped table. A "freshness check needed" flag,
checked: done.

**Phase 3e.** active.md described it as "🔨 building, latest commit
bf4cb7c." 3e (the MCP bridge) shipped 2026-05-08 and is the live tool
surface every later batch rode on — the ES arc extended it with
streaming dispatch, the judge-compiled-brief, and `consult_subagent`.
Valid and foundational, not superseded; the row was just frozen.

## The new idea

`substrate-scheduled-workflows.md` — Michael wants cron-style scheduled
jobs: periodic exhibit research from physics news, autonomous review of
AI YouTube videos, and ingestion of a public YouTube playlist he
curates. Captured as a stub (D-SW1–D-SW7). The stub foregrounds the
autonomous-spend concern on purpose — a timer that auto-spawns pipelines
is exactly the ES failure class, so the human-in-loop gate and
per-schedule budget are named as load-bearing, not polish.

## The discovery worth keeping

`active.md` had grown to **185 lines / ~51K tokens** — and ~40 of those
lines were a stack of dated session banners reaching back to 2026-05-09.
The file's *own* Edit rule says "rewrite directly, do not append — the
archive lives under `.mind/archive/`." The rule was being violated one
banner at a time, and nobody noticed because each session only added
one. Collapsed the 40 banners into a single current-state banner; the
file is back to ~100 lines.

The lesson is about cadence, not this file. A "current state" document
degrades invisibly: every session's honest one-line update is correct
in isolation and the sum is an archive. Consolidation has to be a
periodic discipline — ideally a banner-count check at session end — not
something that waits for a human to ask "why is this so long?"

## Also done in the active.md pass

De-duplicated the In Flight table (four rows already recorded in Recently
Shipped, marked archived, still listed as in-flight — removed). Rolled
the "last ~30 days" Recently Shipped window: added the May substrate arc
as three summary rows + the last-supper study; rolled the Apr 4–15 rows
off. Fixed a corrupted glyph and the stale study pointer.

No code changed this session. Docs and memory only.
