---
date: 2026-05-09
agent: dev (research mode)
session_kind: deep research
priority: high
purpose: >
  Michael's council ask: "I'd like to see you go through phases, research,
  plan, explore, write. Keep a journal or scratch file as you go. … Come
  back with a document that outlays all of the decisions we need to make
  together. Keep going until you fully understand the work needed to get
  a fully agentic system working in pg-ai-stewards that has the spirit
  of our 11 cycle guide that has enough arms and legs to actually get
  work down."
---

# Full agentic substrate — research journal

This is my running notebook as I research. Not the deliverable; that
goes in `.spec/proposals/full-agentic-substrate.md`. This file is
the *process* — what I read, what surprised me, what shifted my
thinking, what dead-ends I found.

The deliverable's quality depends on the depth of this work. I'm
trying to be thorough enough that the proposal is informed by
brain v3's actual judgment + the 11-cycle's actual theology, not by
my own pattern-matching on package docs.

## Method

Following the Michael's stated phases:

1. **Research** — read brain v3's orchestrator code FULLY (steward,
   commission, diagnosis, breaker, retry, nudge, pipeline files,
   data model). Read the 11-cycle guide and earlier parts. Read
   substrate code I haven't read yet (schema.rs, full tools.rs,
   compose_messages SQL).

2. **Plan** — for each of the 11 steps, articulate (a) gospel
   principle, (b) brain v3 mechanism, (c) substrate today, (d) gap,
   (e) substrate-native port shape.

3. **Explore** — surface tensions, dead-ends, decisions that
   contradict, things that don't translate cleanly.

4. **Write** — the proposal doc with phase ladder + decisions
   needed.

I'll commit this journal at the end so the process is auditable.

## Council moment — orientation before reading

The framework I keep coming back to (because it's true): the
substrate is supposed to be theology-applied-as-code. Not theology
as decoration on a generic AI infrastructure. The brain v3
package docstrings cite scripture not as flavor but as *naming the
pattern they implement*. The Watch→Diagnose→Act→Account loop IS
the steward in D&C 72. The retry-with-feedback IS the redemptive
recovery from D&C 82:7.

The honest question I'm holding: **does pg-ai-stewards-as-built
embody any of this, or is it just generic AI infrastructure with
gospel-flavored .spec files on top?**

I think the answer (entering this research) is: mostly generic.
The bridge is generic. The bgworker is generic. The MCP plumbing
is generic. The gospel-orientation lives in agent prompts and
study outputs, not in the substrate's own discipline.

The opportunity here isn't "add features to be like brain v3."
It's: **rebuild the substrate's own behavior so the cycle is
enforced in code, not just suggested in prompts**.

Let me read.

## Synthesis — what brain v3 actually invented

Read in this order: `steward.go` → `diagnosis.go` → `retry.go` →
`breaker.go` → `commission.go` → `pipeline/gate.go` → `store/types.go`.
Skim of the rest (research/execute/scaffold/context/review).

Six load-bearing inventions, each anchored in scripture:

1. **Watch → Diagnose → Act → Account loop** (D&C 72; Mosiah 18:9).
   The steward is a goroutine, not a queue. It owns the entry's
   wellbeing across attempts, classifies failures into five types
   (transient, timeout, model_limit, tool_error, unknown), and
   chooses an action per type. The action is *named* and *logged*
   into a ring buffer. The "account" step is the action log being
   inspectable by the human. This is what the substrate does NOT
   do today — work_items just have `status='failed'` and an `error`
   string with no diagnosis.

2. **Retry-with-feedback** (D&C 82:7; Moroni 6:8). When retrying a
   failure, the steward injects a synthetic system message into the
   next dispatch: "Steward retry context (attempt N): Previous
   failure: X. <type-specific guidance>". The model is told what
   went wrong AND given guidance shaped to the failure type. This
   is the substrate's biggest capability gap — its retries today
   are blind reruns. Models walk into the same trap. The brain
   v3 pattern is "we send you in again, but with better
   intelligence about the trap you fell into."

3. **Circuit breaker** (D&C 101:47-54 — the parable of the
   nobleman). Per-stage failure tracking. Five failures in a row
   trips the breaker → cooldown for 10 minutes → one probe attempt
   → close on success, re-open on failure. The substrate today has
   no notion of "this stage is sick" — it'll happily dispatch into
   a known-broken stage forever.

4. **Model escalation chain** (talents parable; D&C 82:3 — "where
   much is given, much is required"). Failures escalate the model
   chosen for retry: if a sonnet failure is `model_limit`, the next
   attempt uses opus. Cost is tracked per attempt. The chain is
   stage-aware (execute defaults sonnet; plan defaults opus). The
   substrate today picks model from `agent_family.default_model`
   with no escalation path.

5. **Commission flow / Ammon-loop** (Alma 17:25 — Ammon serves the
   king's flocks, then asks what to do). An entry has a *maturity*:
   raw → researched → planned → specced → executing → verified.
   Each maturity transition is a `commissionAdvanceStage` cycle:
   run the pipeline → evaluate gate (advance|revise|surface) →
   handle decision. Revisions cap at 2; surface bumps to
   `route_status='your_turn'`. This is the substrate's
   biggest missing layer — work_items today are *single-stage* (one
   prompt → one response → done or failed). There's no "this isn't
   ready yet, send it back" mechanism.

6. **Gate evaluator + scenarios + verify** (Abraham 4:18,
   "they were obeyed"). Three discrete LLM calls per maturity:
   - `EvaluateGate` — should this advance, revise, or surface?
   - `GenerateScenarios` — produce 3-7 acceptance criteria as JSON
   - `EvaluateAndVerify` — does the execution output meet the
     scenarios? per-scenario pass/fail with notes.
   Substrate today has zero gates. Whatever the model produces is
   accepted as-is and the work_item moves on.

The pattern under all six: **brain v3 doesn't trust any single LLM
call.** Every stage is wrapped in some discipline that imposes
external judgment on the model's output, and every failure is
treated as data to learn from rather than just a retry counter
to increment.

## Substrate today — the actual orchestration surface

`work_items` table (16 columns): id, slug, pipeline_family,
current_stage, status, input, stage_results (jsonb keyed by
stage), session_ids, token_budget, tokens_in/out, actor, error,
created_at, updated_at, completed_at.

13 SQL functions, the load-bearing ones:
- `work_item_create(family, input, slug, actor, budget)` — insert
- `work_item_dispatch_stage(item_id, user_input)` — composes
  messages via `compose_messages`, enqueues into work_queue, returns
  work_id
- `work_item_advance(item_id, stage_output)` — moves to next stage
  via `pipeline_stage_lookup`; sets `completed` if last stage
- `work_item_fail(item_id, error)` — sets status='failed'
- `compose_messages(family, model, session_id, user_input)` —
  composes system + history + current
- `compose_system_prompt(family, model, session_id)` — builds
  system message from agent_family registry

What's missing relative to brain v3:
- No diagnosis. `error` is a free-text string.
- No retry. A failed work_item is terminal until human re-dispatches.
- No gate. `work_item_advance` is unconditional; whatever the
  stage produced is accepted.
- No scenarios. No verify.
- No circuit breaker. Pipelines can fail repeatedly with no
  protection.
- No model escalation. Model is fixed at agent_family level.
- No commission. There's no maturity ladder above pipelines.
- No "Account" surface. Watchman tells you the work_item exists
  and what status it's in; it doesn't show the *judgment trail*
  (why this model, why this retry, what diagnosis).

## The 11-cycle source — what intent and spec actually demand

Re-read parts 3 (intent) and 4 (spec) of the guide.

**Intent (Part 3 / Klarna pattern).** Intent is what an agent
optimizes for *when instructions run out*. The substrate today
encodes intent only in agent prompts (text inside `agents` table).
There's no first-class intent layer. When a work_item's stage
prompt is silent on a tradeoff, the model picks based on training
priors. This is the Klarna failure mode. Brain v3 partially
addresses this via `purpose_summary` baked into entries, but its
mechanism is also "system prompt prose" — neither system has
intent as data the orchestrator reasons about.

**Spec (Part 4 / Abraham 4:18 "they were obeyed").** The five
primitives: self-contained problem, acceptance criteria,
constraints, edge cases, verification. Brain v3's commission flow
maps these onto: spec stage = problem + constraints; scenarios
stage = acceptance criteria + edge cases; verify stage =
verification. The substrate has none of this — `work_items.input`
is freeform jsonb that pipelines interpret however they want.

The deeper point in Part 4: "Real-time prompting rewards verbal
fluency. Specification engineering rewards completeness of
thinking." A substrate that supports autonomous agents has to
reward completeness, not fluency. That means the substrate needs
opinionated structure for what a "spec" is, and gates that won't
let work proceed past spec stage until the spec actually meets
the bar. **The substrate today has no opinion about what a spec
is.** That's the root cause of "vibe coding" — the orchestrator
doesn't enforce the discipline that the 11-cycle demands.

## The pivot

The proposal needs to be honest that **the substrate today is a
durable, observable, multi-tenant runtime**. That's real value.
But it isn't an "agentic" system in the 11-cycle sense; it's a
job runner with very nice telemetry. Brain v3 is the opposite —
clunky storage and weak observability, but the *orchestration
discipline* maps closely onto the 11-cycle.

The substrate-native port isn't "copy brain v3 over." It's: take
brain v3's six inventions, redesign each one to live as
Postgres-native state + bgworker behavior + UI surface. Done
right, the substrate inherits brain v3's discipline AND keeps its
durability/observability advantages.

That's what the proposal needs to lay out. Going to write it now.


