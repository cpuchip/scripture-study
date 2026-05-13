---
date: 2026-05-13
mode: plan + fix
workstream: WS5 (substrate) + WS9 (other apps — science center enabler)
project: pg-ai-stewards
title: "Docker compose pg restart-policy fix + Batch J ratified (fan-out + brainstorm + hierarchy UI)"
status: plan locked; build starting (J.2)
carry_forward:
  - "Batch J.2 build pulse beginning now — fan-out machinery (pipeline + spawn fn + aggregate trigger)"
  - "J.1 (UI hierarchy + open filter) and J.3 (apply fan-out to 8218aa77) follow alongside"
  - "J.4 (brainstorm + 4-lens library: SCAMPER, Six Hats, Crazy 8s, reverse) follows"
  - "J.5 (compose brainstorm + fan-out on a new category) closes the arc"
  - "yaml.rs gated on 3rd YAML SHAPE (unchanged)"
  - "Projects B deferred (unchanged)"
  - "14 SC work_items pending ratification (unchanged)"
  - "Friend's brainstorming modes — ask Michael async; reconcile when received"
links:
  - "../proposals/substrate-batch-j-fanout-brainstorm.md"
  - "../../projects/pg-ai-stewards/extension/docker-compose.yaml"
---

# 2026-05-13 — Docker fix + Batch J plan locked

Two distinct pieces of work today before the J.2 build pulse begins.

## (288b089) Docker compose pg restart-policy fix

Symptom: `pg-ai-stewards-dev` exited 06:30 UTC (clean SIGTERM, exit 0 — most likely Docker Desktop restart or host reboot) and stayed down. Bridge and UI had `restart: unless-stopped` and kept flapping (RestartCount=13 by report time) because they couldn't reach `pg:5432`.

Root cause: `pg` service had no `restart` policy (Docker default = `no`). The other two services did. The foundation of the stack was the only piece without auto-restart.

Fix: added `restart: unless-stopped` to the `pg` service to match bridge and ui. The stack now comes up as a unit on host reboots while still respecting an intentional `docker compose stop pg`.

Recovery: `docker compose up -d` recreated pg; bridge/ui had stale DNS resolution (`lookup pg on 127.0.0.11:53: no such host`) and needed a `docker compose restart bridge ui` to clear it. All three now healthy.

Soak survived the recreate — `schedule_enabled = t` persisted in the data volume.

## Batch J ratified — fan-out + brainstorm + hierarchy UI

Triggered by Michael's review of work_item `8218aa77` (`everyday-science-to-exhibits`). The `research-write` pipeline did good work in its shape but the shape was the limit — a single-file linear pipeline can't fan out into the 6-field-per-exhibit deliverable Michael asked for. Diagnostic acknowledged in the work item's own review stage: *"6-part coverage status: ⏳ needs drafting on five of six rows."*

Michael's framing: *"what we build is a good framework/harness context engine, we need more legs here."* Batch J adds two new pipeline shapes (legs) and the UI affordances that keep the work-item list legible once parents start spawning children.

### Three shapes added

- **Fan-out (`decompose-fanout`)** — `context_gather → decompose → spawn → aggregate`. One binding question → N children with their own pipelines → roll-up index.
- **Brainstorm** — `context_gather → divergent (N parallel lens dispatches) → converge`. Lenses are agents with `role='brainstorm-lens'`.
- **Composed chain** — no new pipeline; brainstorm's converged candidate list becomes fan-out's manifest input. Two work_items linked by `parent_work_item_id`.

### Sub-phases

| Sub-phase | What | Cost | Pulses |
|---|---|---|---|
| J.1 | Work-item hierarchy UI: tree + open filter | $0 | 1 |
| J.2 | Fan-out pipeline + spawn fn + aggregate trigger | ~$1 smoke | 2 |
| J.3 | Apply fan-out to 8218aa77 → 6 exhibit briefs | ~$3–5 | 1 |
| J.4 | Brainstorm pipeline + 4-lens library | ~$2 smoke | 2 |
| J.5 | Compose brainstorm + fan-out on new category | ~$5–10 | 1 |

Total estimate: ~7 pulses, ~$15 LLM across the arc.

### Ratified decisions (three AskUserQuestion batches, 12 questions)

**Architecture (J.2):**
- A1: Spawn = deterministic SQL function (no LLM call)
- A2: Aggregate trigger = event-driven via `on_maturity_verified` (council pattern)
- A3: Child pipeline = per-child in the manifest (mixed shapes OK)
- A4: Aggregate output = index always + optional digest via `aggregate_synthesis=true` metadata flag

**Brainstorm (J.4):**
- B1: Divergent = mix (each lens declares provider; some local LM Studio/Ollama, some cloud)
- B2: Lens scope = start with 4 (SCAMPER, Six Hats, Crazy 8s, reverse-brainstorm)
- B3: Lens storage = `stewards.agents` rows with `role='brainstorm-lens'` (reuses dispatch machinery)
- B4: Friend's modes = research literature now; reconcile when friend's list arrives

**UI + sequencing (J.1):**
- C1: Open filter = status group dropdown (open/done/all + individual)
- C2: Tree render = both (indent + parent-link badge)
- C3: Pulse order = J.2 first; J.1 + J.3 alongside; J.4 + J.5 follow
- Schedule: now

### Why this matters

Today's Batch I work (agent-proposal endpoint, validate-sql gate, file write-back) opened the substrate for agents to write *files*. Batch J opens it for agents to spawn *work* — one binding question producing many deliverables. That's the framework move Michael named when he said "more legs."

The first concrete payoff is the science center exhibits library — J.3 turns 8218aa77's survey into 6 exhibit briefs in `projects/space-center/exhibits/`. The deeper payoff is the primitive itself: fan-out applies to study generation, lesson prep, talk prep, any place where one question should produce many artifacts.

### What's NOT in scope

- Friend's specific 8-9 brainstorming modes (ask async; integrate later)
- A dedicated `exhibit-brief` pipeline shape (research-write with a tighter binding question should be enough for J.3; revisit if briefs feel thin)
- Cost-cap policy refinement (default $0.50 per child; revisit after J.3)
- Brainstorm-then-fanout as a single pipeline (composition is just two linked work_items)

## Status at session yield

- Proposal saved: `projects/pg-ai-stewards/.spec/proposals/substrate-batch-j-fanout-brainstorm.md` (status: ratified)
- Soak paused (`schedule_enabled = f`)
- Stack healthy: pg + bridge + ui all up
- Next action: write J.2 SQL migrations (j1-fanout-pipeline.sql + j2-spawn-children-fn.sql)
