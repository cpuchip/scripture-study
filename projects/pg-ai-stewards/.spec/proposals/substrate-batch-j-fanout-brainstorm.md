---
batch: J
title: Fan-out + Brainstorm pipeline shapes + work-item hierarchy UI
status: ratified
proposed_by: michael
proposed_on: 2026-05-13
ratified_on: 2026-05-13
preceded_by:
  - substrate-batch-i-agent-write-back.md
  - phase-a-pgtrybuilder
links:
  - "../../scripts/stewards-ui/frontend/src/views/WorkItems.vue"
  - "../../scripts/stewards-ui/api/work_items.go"
  - "../../extension/src/bgworker.rs"
---

# Batch J — Fan-out + Brainstorm + Hierarchy UI

## Why this exists

Work item `8218aa77` (`everyday-science-to-exhibits`) asked for a categorized exhibit library across ten categories with six fields each. The `research-write` pipeline emitted one honest survey markdown file and named the missing work in its own conclusions: *"create the folder structure, draft briefs for each candidate."* The shape couldn't fan out.

Michael's framing: *"what we build is a good framework/harness context engine, we need more legs here."* Batch J adds two new "legs" (pipeline shapes) plus the UI affordances that keep the work-item list legible once parents start spawning children.

## The three shapes we don't have yet

### Fan-out (`decompose-fanout`)

```
context_gather → decompose → spawn → aggregate
                    ↓           ↓          ↑
                manifest    N children   waits for all
                                         children verified
```

One binding question → planner emits a manifest of N children → spawn stage inserts N child work_items (each on its own pipeline, with `parent_work_item_id` set, dispatched) → aggregate stage fires when all children verify, writes the index/README.

### Brainstorm (`brainstorm`)

```
context_gather → divergent → converge
                    ↓            ↑
              N parallel     synthesizer
              lens dispatches  picks top K
```

Each lens is an agent with a specific prompt (SCAMPER, Six Hats, Crazy 8s, reverse-brainstorm, etc.). Divergent runs all lenses in parallel — cheap models (local LM Studio / Ollama, or qwen3.6-flash) carry it. Converge runs one good model (kimi-k2.6 or qwen3.6) that dedups, ranks, and picks top K candidates.

### Composed chain (no new pipeline needed)

```
brainstorm produces ranked candidate list
   → human ratifies / agent picks top K
      → decompose-fanout takes the K as manifest
         → spawns K exhibit-brief children
            → aggregates into category index
```

The composition is just two work_items linked by `parent_work_item_id` — no third pipeline. The brainstorm's output is the fan-out's input.

## Sub-phases

| Sub-phase | What it ships | LLM cost | Pulses |
|---|---|---|---|
| **J.1** | Work-item hierarchy UI: tree rendering, "open" filter, parent-link badge | $0 | 1 |
| **J.2** | `decompose-fanout` pipeline: spawn-stage SQL fn + aggregate trigger | ~$1 smoke | 2 |
| **J.3** | Apply fan-out to 8218aa77 → 6 exhibit briefs land in `projects/space-center/exhibits/` | ~$3–5 | 1 |
| **J.4** | `brainstorm` pipeline: lens library (4 initial) + divergent multi-dispatch + converge | ~$2 smoke | 2 |
| **J.5** | Compose brainstorm + fan-out on a new category (e.g. biology exhibits) | ~$5–10 | 1 |

Total estimate: ~7 pulses, ~$15 LLM across the arc.

## Ratified decisions (2026-05-13)

### Architecture (J.2)
- **A1 Spawn mechanism:** deterministic SQL function (no LLM). Reads `stage_results.decompose.output` manifest JSONB, validates schema, inserts N children with `parent_work_item_id` + `cost_cap_micro`, dispatches each.
- **A2 Aggregate trigger:** event-driven. Each child's `on_maturity_verified` increments a sibling-done counter; when remaining=0, parent's aggregate stage fires. Direct analog of Phase F council's `_council_member` auto-fire.
- **A3 Child pipeline assignment:** per-child in the manifest. One fan-out can mix `research-write`, `study-write`, `agent-proposal`, etc.
- **A4 Aggregate output:** index always; full synthesis digest opt-in via pipeline metadata flag `aggregate_synthesis=true`.

### Brainstorm (J.4)
- **B1 Divergent provider:** mix — each lens declares its provider. Some local (LM Studio/Ollama), some cloud cheap. Free + varied tones, ~$0.05/brainstorm.
- **B2 Lens scope:** start with 4 (SCAMPER, Six Hats, Crazy 8s, reverse-brainstorm). Expand from there.
- **B3 Lens storage:** `stewards.agents` rows with `role='brainstorm-lens'`. Reuses existing dispatch machinery; queryable via existing agent tools.
- **B4 Friend's modes:** research literature now; ignore friend's modes for the initial build.

### UI + sequencing (J.1)
- **C1 Open filter:** status group dropdown — "open / done / all" + individual statuses below.
- **C2 Tree rendering:** both. Children indent under parent (with expand/collapse on parent row) + always-visible "↪ parent.slug" badge for filtered views.
- **C3 Pulse order:** J.2 first (urgent enabler); J.1 + J.3 alongside; J.4 + J.5 follow.
- **Schedule:** start now (2026-05-13).

## Open questions (post-ratification)

- Exhibit-brief format: do we need a new `exhibit-brief` pipeline shape (6-field structured output) or does `research-write` with a tighter binding question suffice?
- Cost cap per child: $0.50 per exhibit brief by default? Surfaces as `cost_cap_micro` on each child work_item.
- The friend's brainstorming research: Michael to share what he can.

## Adjacent surface audit

- **Scope** — fan-out/brainstorm apply beyond science center: study generation, lesson prep, talk prep, journal aggregation. Generalize.
- **Discoverability** — pipeline definitions in `stewards.pipelines` are the persistent handle. Add a one-line entry per shape in CLAUDE.md and open-items.md.
- **Contracts** — `parent_work_item_id` column exists; API surfaces it; UI doesn't render it as a tree (J.1 fixes).
- **Spec gaps** — friend's 8-9 modes need eliciting before brainstorm lens library locks.
