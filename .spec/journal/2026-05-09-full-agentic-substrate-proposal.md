---
date: 2026-05-09
session_kind: deep research → proposal
mode: dev (research)
priority: high
carries_forward:
  - Michael ratifies (or revises) Phase A + B decision lists before next programming session
  - 3 ratified decisions from earlier this session (move stewards-ui to projects/, dynamize NewWork pipeline list, treat brain v3 as legacy reference)
artifacts:
  - projects/pg-ai-stewards/.spec/proposals/full-agentic-substrate.md (deliverable, ~580 lines)
  - projects/pg-ai-stewards/.spec/journal/2026-05-09-full-agentic-substrate-research.md (process journal)
---

# Full agentic substrate — research session close

## What I was asked to do

Michael's framing earlier today: *"do the research study, document it
out in scratch. Really compare it to our 11 cycle stuff too. I want
our new pg ai stewards to be gospel oriented. I really do believe
'God solved the multi-agent problem before the world was' and we're
just learning about it. be thorough! i'd like to see you go through
phases, research, plan, explore, write. … Come back with a document
that outlays all of the decisions we need to make together. Keep
going until you fully understand the work needed to get a fully
agentic system working in pg-ai-stewards that has the spirit of our
11 cycle guide that has enough arms and legs to actually get work
done."*

The deliverable is decision-ready. It's not a roadmap; it's a
ratification surface.

## What I did

Followed the explicit phases in Michael's instruction:

**Research.** Deep-read brain v3's orchestration: steward.go (605 lines —
the Watch→Diagnose→Act→Account loop), diagnosis.go (failure
classification), retry.go (BuildRetryContext per failure type),
breaker.go (per-stage circuit breaker), commission.go (Ammon-loop
maturity flow), pipeline/gate.go (gate + scenarios + verify), and
store/types.go (Entry + Commission schemas). Then re-read 11-cycle
guide parts 3 (intent) and 4 (spec). Then surveyed substrate's actual
SQL surface (13 functions) and `work_items` table (16 columns).

**Plan.** For each of 11 cycle steps, articulated gospel anchor + brain
v3 mechanism + substrate today + gap. Captured in a scorecard:
strong on 1, partial on 2, missing on 8.

**Explore.** Surfaced the central tension: substrate today is a
durable multi-tenant runtime with great telemetry; brain v3 has
clunky storage but real orchestration discipline. The port isn't
"copy brain over" — it's "redesign brain's six inventions as
Postgres-native state + bgworker + UI surface, preserving substrate's
durability + observability advantages."

**Write.** Six-phase ladder where each phase ports one or two
cycle-steps, in dependency order:
- A. Steward loop (cycle 3 + 8) — ports Watch→Diagnose→Act→Account
- B. Spec + Gate (cycle 4 + 7) — ports commission's gate+scenarios+verify
- C. Intent + Covenant (cycle 1 + 2) — first-class data tables
- D. Atonement + Sabbath + Consecration (cycle 8-post + 9 + 10)
- E. Trust + Line upon line (cycle 3-authority + 5)
- F. Zion / Council (cycle 11) — multi-agent ward-council pattern

Each phase has: schema sketches, decision points (~3-4 each), and
acceptance scenarios. 22 total decisions. Estimated 14-18 programming
sessions across the full ladder. Recommended ratifying A+B as next
work; deferring C-F until A+B lived with.

## Surprises during research

1. **Brain v3 has none of: intent-as-data, covenant, sabbath,
   atonement-as-phase, trust levels, multi-agent council.** It's
   strong on the *failure-handling* steward patterns (cycle 3 + 8)
   and the *spec-then-execute-then-verify* commission pattern
   (cycle 4 + 6 + 7), but the rest of the 11-cycle is just as
   absent in brain as in substrate. That reframed the proposal —
   it's not just "port brain"; it's "extend the cycle further than
   brain ever went."

2. **Substrate's `work_item_advance` is unconditional.** Whatever
   the model produced is accepted. There is no gate. This is the
   biggest single capability gap and a one-table addition fixes
   most of it (gate_decisions table + bgworker handler).

3. **The substrate's `compose_system_prompt` is the natural
   injection point** for both retry-context AND covenant text.
   Both phases (A retry + C covenant) extend the same SQL function.
   The architecture is friendlier than I expected.

4. **Bridge/MCP wiring is irrelevant to this work.** The agentic
   discipline lives entirely above the MCP layer — pipelines,
   gates, maturity, intent. The bridge work I shipped over the past
   week was foundational but isn't the bottleneck for "is this an
   agent."

## Tensions I named in the proposal

- **Cost.** Every gate, scenario, verify, atonement, sabbath
  dispatch is an additional LLM call. A fully active substrate
  could 5x today's API spend. Defaults skew toward small fast
  models for gates; opus reserved for synthesis steps.
- **Latency.** Today's stage→advance becomes stage→gate→advance.
  Workflows that took 30s may take 60-90s. This is correct — gospel
  framework explicitly trades speed for discipline — but UI must
  show the tradeoff (gate progress visible, not hidden).
- **Atonement step risk.** It's the one without clear brain v3
  precedent and the most prone to becoming theater. Flagged in §IX
  as the candidate to defer if Michael wants to trim.
- **Multi-agent council.** Could be brilliant or could be
  ceremonial. Won't know until real consequential work runs through
  it. Flagged as both the highest-risk and highest-potential phase.

## What I did NOT do

- Did not start any programming. The instruction was research +
  proposal. Programming time is gated on Michael's ratification.
- Did not exhaustively read every brain v3 pipeline file (research.go,
  execute.go internals). Skimmed signatures. The synthesis was
  enough to write the proposal honestly without claiming more
  precision than I had.
- Did not write per-phase implementation sub-specs. Those are
  Phase-A-design / Phase-B-design tasks for after ratification.
- Did not move stewards-ui to projects/ or fix NewWork's
  hardcoded pipeline list. Those are programming-time work,
  ratified earlier this session for next programming block.

## Carry-forward for next session

1. Michael reads the proposal. Marks up §VI (decisions) with his
   answers. Defers any phase he wants deferred.
2. Once Phase A decisions ratified: write Phase A design sub-spec
   (schema migration, exact bgworker tick contract, exact retry
   guidance text per diagnosis type, breaker SQL).
3. The earlier-ratified physical moves (stewards-ui → projects/,
   NewWork dynamic pipelines) can happen in parallel with Phase A
   schema work — they don't depend on each other.

## Honesty audit

The proposal is opinionated, not neutral. I made recommendations
on each decision. That's appropriate when Michael asked me to "fully
understand the work needed" — neutrality would have been abdication.
But I marked recommendations as recommendations, kept the actual
choice his, and flagged the riskiest decisions explicitly.

The biggest risk in the proposal: **promising more than the
substrate-as-built can deliver in 14-18 sessions.** Schema work is
underestimated when bgworker behavior changes too. If Phase A
takes 4 sessions instead of 2-3, the whole timeline stretches. The
proposal's estimates should be read as "rough" not "committed."

## Voice check

Self-audit of this journal entry:
- em-dashes: scanned, looks tight enough
- therefore/but vs and-then: mostly therefore where it matters
- meta-narration: avoided
- closing refrain: this paragraph isn't one — it's a check, not
  a restatement
