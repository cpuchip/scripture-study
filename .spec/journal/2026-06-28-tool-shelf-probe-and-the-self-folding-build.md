---
date: 2026-06-28
lane: pg-ai-stewards
topic: the north-star landed, brain/engram search got its twin, and the tool shelf went from a half-remembered idea to a probe-proven, real-path build — the tools that put themselves away
tags: [north-star, brain-search, engram-search, tool-shelf, progressive-disclosure, telemetry, probe, delegation, dominion-in-council, self-folding]
---

# The day the substrate learned to put its tools away

One long continuous session (it ran past midnight — the date turned on us). It started as a
session-close and became a full arc: the north-star shipped, search got finished, and a half-remembered
idea ("the tools engram") turned into a real, evidence-backed, built mechanism.

## What landed
- **North star (PR #13, merged).** The substrate's Intent on every call — Col 3:17 in Michael's overlay,
  generic in the core. (Built earlier in the same session; merged this turn.)
- **Brain + engram search wired (PR #14, merged).** `75` repointed brain search to embed+hybrid; `76`
  built `engram_search` — the agent-facing twin that didn't exist. **The sub-agent built the twin
  instead of just flagging it** (the Dave rule), AND rightly built it as a *proposal* because a new
  agent-tool is a new standing capability (`dominion_in_council`) — it refused to treat my relay as
  Michael's consent. Michael ratified "merge as-is, RLS scopes the cross-session exposure later." The
  multi-tenancy spec now names `engram_search` as the canonical RLS-leak test case.

## The tool shelf — the real arc of the day
Michael asked "did we ever finish tool groups? the tools engram?" The honest answer chain:
1. **`37-tool-groups` is the *static* half** (12 research stages narrow their dump). The *dynamic*
   on-demand half — the actual progressive-disclosure shelf — was never built. The "tools engram" was
   an idea, not code.
2. **The research he half-remembered is real:** Google SDLC Day-1, Figure 4 — Agent Skills as
   progressive disclosure ("lightweight metadata at startup, load on demand, pay for only what you
   use"). The insight: *do the same to tools.*
3. **Telemetry (read-only sub-agent):** the context-management tools are *barely used* — only the
   read/page-in side (`expand_message` 84, `context_search` 13), the entire write/curate side **zero**.
   The reactive engine (3,886 engram extractions) carries context management. **That answers CT2.4
   (#136): agent-driven context management does NOT earn its keep against the reactive engine.** And
   `skill_group_open`=0 looked damning until we saw the confound — skills *autoload*, so nobody ever
   needs to open one manually.
4. **★ The probe (the heart of it).** Rather than trust the confounded telemetry, Michael's instinct:
   run the real thing. A standalone harness folded **157 tools** to a name+purpose catalog + a
   `reveal_tool(name)` lever + a note, and drove the *local* models on the real vivint task.
   **qwen3.6-35b-a3b opened exactly the right tools turn 1, 0 misses, 100% of 33 calls on revealed
   tools. gemma-4-26b-a4b opened them and completed the task, grounded, finish=stop.** The manual
   shelf works on local models. The skill_group_open=0 fear was the confound, not the verdict.
5. **The build (PR #15, P0a + P0b, Michael's Hinge).** Re-authored the dispatch chokepoint
   (`dry_run_chat`) + `compose_system_prompt` + `compose_tools`, all gated, **flag-off proven
   byte-identical** by inverse hypothesis on dev. Michael's **self-folding** design: default-fold all →
   `reveal_tool` → **cooldown auto-refold** (a tool unused N rounds folds itself — "the tools put
   themselves away") → `pin_tool`. **Real-substrate test:** gemma revealed `doc_search`+`doc_get` on the
   *real* dispatch path and completed — folded tools array **1.2 KB vs 105 KB unfolded, ~99% reduction.**

## The lesson Michael named
I'd over-constrained the brain/engram sub-agent ("surface, don't act") — which fought our own
`exercise_stewardship` covenant. The correction that matters: **the deterministic oracle (virgin-smoke)
is what makes the Dave rule safe to delegate.** Brief sub-agents to *act on the obvious adjacent twin
and prove it with the oracle*, not to default to surface-only. The agent definition already carries the
rule; my brief was the bug. Two-way comms (SendMessage to resume a finished agent) demonstrated.

## Housekeeping
- Dev had drifted (predated 71/72/73); my 71 re-apply hit the drop-then-create landmine (dup
  `doc_search_hybrid`) — caught + fixed by continuing the chain from 72 in order. **Follow-up: wire
  `migrate.sh` self-reconcile so a long-lived dev box stops accumulating drift.**
- Rig: Michael gamed with his son → GPUs freed; `dance-moe` reloaded (qwen + gemma-MoE both resident).

## Carry-forwards
- **PR #15 (tool shelf) — Michael's Hinge.** Default-off; a virgin install is unchanged. Untested edge:
  scoped-pipeline-stage + shelf-on stacking (per spec §4.4, no test exercises it yet).
- **Multi-tenancy council** still pending (RLS one-way door) — the north-star's WHO+WHY rail + the
  engram_search cross-session surface both wait on it.
- **`migrate.sh` self-reconcile** (kill dev drift).
- **BINEVAL "Ask, Don't Judge"** (new inbox, parked) — binary-question judges upgrade 56/59/74; make the
  north-star directions checkable y/n questions; guard 59 against instruction-overload. My stewardship,
  for when we next touch the critic.
