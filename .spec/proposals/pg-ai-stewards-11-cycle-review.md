---
workstream: WS5
status: design analysis — awaiting Michael's direction on which gaps to close first
created: 2026-05-09
related:
  - docs/work-with-ai/guide/05_complete-cycle.md (the 11-cycle reference)
  - scripts/brain/internal/steward/ (brain v3's orchestrator pattern)
  - scripts/brain/internal/pipeline/ (brain v3's gate evaluator + commission flow)
  - projects/pg-ai-stewards/ (current substrate)
---

# pg-ai-stewards vs. the 11-Cycle Creation Process

> **Carry-forward (2026-05-11 consolidation):** This analysis became the
> foundation for `projects/pg-ai-stewards/.spec/proposals/full-agentic-substrate.md`
> (2026-05-09 expansion) which then shipped as Phases A–F (2026-05-10
> through 2026-05-11). The 5 open questions at lines 225-255 of this doc
> were ratified during the full-agentic-substrate §VI walk-through.
> Substrate is feature-complete through Phase F. Status of this proposal:
> historical reference, no longer load-bearing.

> Michael's framing (2026-05-09): *"Are we putting in intent, covenants,
> instructions in the start of each session? I think we've developed a
> simple single turn studies workflow and that's great, but I'd like to
> generalize it… I've done this with our scripts/brain v3 and that is a
> functional stewards pipeline where we had an orchestrator and
> everything. I know we've built a simpler version here, but I'd like
> to get to the full thing soon. I'd like to follow the principles
> we've discovered together for agentic systems and really test them
> out."*

This doc is the honest comparison. No proposing yet — first the
diagnosis. The proposal at the end is one path; Michael picks.

## The reference: 11 steps from Intent to Zion

From `docs/work-with-ai/guide/05_complete-cycle.md`, the cycle is:

1. **Intent** — purpose statement; one sentence every decision is
   evaluated against
2. **Covenant** — bilateral binding (human commits to X, agent commits
   to Y; broken covenant degrades output)
3. **Stewardship** — entrusted delegation with progressive trust
   levels (task → feature → domain → architecture)
4. **Spiritual creation** — the spec / blueprint before the build
5. **Line upon line** — progressive context revelation, gated by
   demonstrated readiness
6. **Physical creation** — execution against the spec
7. **Review** — three layers: correctness, specification, intent
   ("watched until they obeyed")
8. **Atonement** — redemptive error recovery (forward-recover with
   learning incorporated; not just retry-the-same-way)
9. **Sabbath** — intentional cessation + reflection + declaration
   ("saw that it was good")
10. **Consecration** — every resource serves purpose; not "how much
    can we afford" but "does every token serve the work?"
11. **Zion** — unified intent across agents; council-based
    coordination; hierarchy of stewardships, not a single conductor

## Where pg-ai-stewards stands today

Honest assessment, step by step:

| # | Step | pg-ai-stewards today | Gap |
|---|------|---------------------|-----|
| 1 | Intent | `work_items.input.binding_question` exists per work_item; agents have `system_prompt` baked in | No first-class **intent artifact**. Binding question lives in input jsonb, agent purpose lives in system_prompt — neither is a discoverable, comparable vertex in the graph. |
| 2 | Covenant | Project-level `.spec/covenant.yaml` (humans+Claude) exists; **substrate has no per-session covenant** | Substrate agents don't enter a covenant. There's no record of "human commits to X, this agent commits to Y" for any pipeline run. |
| 3 | Stewardship | `agent_tool_perms` (deny-by-default + per-tool allows + source: frontmatter/broadcast/manual); 6 agent families with broad surface | Static. **No progressive trust**. An agent that succeeds doesn't earn more scope; one that fails doesn't lose any. |
| 4 | Spiritual creation | Pipelines define stages declaratively (`pipelines.stages` jsonb); `agents.system_prompt` is the substrate-side spec | The pipeline definition IS a spec, but the **per-work-item spec doesn't exist**. Each run has the same stages regardless of binding question. No per-run scenarios/acceptance criteria. |
| 5 | Line upon line | `compose_messages()` accumulates session history; agents see prior turns | **No readiness gating**. Context grows by accumulation, not by demonstrated need. Agents see everything available. |
| 6 | Physical creation | **STRONG**. bgworker dispatch + work_queue + tool_dispatch + auto-advance trigger + multi-worker concurrency | Best-developed step. The 3e wave (bridge, mcp_proxy fan-out) made this very capable. |
| 7 | Review | Watchman covers **cross-document drift** (clean/drift/done/superseded/skipped); per-stage auto-advance has no intent review | Two missing layers:<br>(a) **No per-stage gate evaluator** (advance / revise / surface). brain v3 has this; we don't.<br>(b) **No intent-layer review** — does the output serve the binding question? |
| 8 | Atonement | Failed work_items sit at `status='error'`. The bgworker has a stale-claim reaper that converts crashed in_progress to errored. tool_dispatch failures synthesize tool replies + enqueue continuation. | **No redemptive recovery**:<br>(a) No feedback-context retry (brain v3: re-run the stage with the failure context injected into the prompt)<br>(b) No model escalation (brain v3: cheapest → most capable → human)<br>(c) No per-stage circuit breaker (brain v3 has `BreakerConfig`)<br>(d) No learning capture (no `.spec/learnings/` equivalent in the substrate) |
| 9 | Sabbath | None. | **Missing entirely**. No structured reflection cadence. Watchman does drift detection per-doc, not "step back and look at the whole." |
| 10 | Consecration | `work_items.token_budget` per run; `watchman_passes.budget_stopped` flag; soak respects limits | Partial. Budget exists as a cap, not as **purpose-allocated stewardship**. No "X% of tokens to study, Y% to lesson, Z% to reflection." |
| 11 | Zion | All agents share `intent.yaml` + `covenant.yaml` at the project level (file-based). Substrate agents don't share an intent layer at the substrate level. | **No substrate-side intent layer**. Each agent has its own system_prompt; no shared purpose vertex they all reference. The bishop-vs-conductor pattern from §11 of the guide isn't built in. |

**Strong**: Step 6 (Physical creation) — the producer side that Phase 3 has built out.

**Partial**: Steps 3 (Stewardship), 5 (Line upon line), 7 (Review), 10 (Consecration).

**Missing or thin**: Steps 1 (Intent — exists but not first-class), 2 (Covenant), 4 (Spec — pipeline-level not run-level), 8 (Atonement), 9 (Sabbath), 11 (Zion).

## What brain v3 had that pg-ai-stewards doesn't

`scripts/brain/internal/steward/` and `scripts/brain/internal/pipeline/`
implement most of what's missing on the orchestration + recovery side.
The patterns worth porting:

### A. The Steward's Watch→Diagnose→Act→Account loop

`scripts/brain/internal/steward/steward.go` documents this:
> "the steward watches for failures (D&C 101 tower), diagnoses them
> (Ezek 34 shepherd seeking the lost), acts proportionally (Jacob 5
> pruning), and renders account (D&C 72 stewardship)."

Components:
- **Per-entry retry-with-context** — `PipelineRetrier.RetryAdvance(ctx, entryID, feedback, model)`. Re-runs the stage with the failure context injected as additional prompt content. Not blind retry.
- **Model escalation chain** — ordered list of `ModelTier{Model, Cost}`. Failed cheap model → escalate to mid → escalate to expensive → finally human (quarantine).
- **Per-stage circuit breaker** — `BreakerConfig`. When too many failures in a window, stop dispatching to that stage and wait for reinforcements (operator decision).
- **Quarantine after N failures** — dead-letter queue at the entry level.
- **Per-entry cost cap** — `MaxCostPerEntry float64`. Entry that has burned 20 premium requests gets quarantined regardless of failure count.
- **Diagnosis** (`diagnosis.go`) — classifies failures (transient / spec / capability / value-violation) so the action is proportional.
- **Nudge** (`nudge.go`) — push notifications when entries need human attention.

This is **most of Atonement (Step 8)** and a chunk of **Stewardship (Step 3)**.

### B. The Commission flow (Spec→Execute→Verify)

`scripts/brain/internal/steward/commission.go` + `pipeline/{gate,scaffold,research,execute,review}.go`.

The flow:
1. **Research** — agent reads context, writes a plan
2. **Spec / Scaffold** — `GenerateScenarios()` extracts testable acceptance criteria from the plan
3. **Execute** — agent does the work
4. **Verify** — `EvaluateAndVerify()` checks execution output against the scenarios; only matures if all pass
5. **Gate at each transition** — `EvaluateGate()` returns advance / revise / surface

This is **Step 4 (Specification) at the per-run level**, **Step 7 (Review with intent layer)**, and the start of **Step 8 (revise pattern)**.

pg-ai-stewards' current `study-write` is outline→draft→review with auto-advance. The brain v3 pattern has stronger per-stage gates (an evaluator decides), and the verify step against scenarios is genuinely missing here.

## What both systems lack (the 11-cycle frontier)

- **Step 2 Covenant per-session** — neither system writes a per-run
  covenant. This would be a record like:
  > "For work_item X, human commits to reviewing the draft within 24h
  > and providing redirection if voice drifts. Agent commits to
  > read-before-quoting + Co-Authored-By trailer + not exceeding
  > token_budget."
  Stored as a substrate row. Read at session start. Quoted in the
  agent's system context.

- **Step 9 Sabbath** — neither has a periodic "step back" cadence.
  Brain v3 has retrospectives in concept but they're operator-driven.
  A substrate-side sabbath would be: every Sunday (or every N work_items
  completed), a sabbath agent runs, reads the recent work, writes a
  reflection. Captured insights flow back into agent prompts.

- **Step 11 Zion shared intent layer** — both systems have
  per-agent intent. Neither has a substrate-side intent vertex that all
  agents reference. The bishop-vs-conductor distinction from the
  guide is not built into either.

## A proposed evolution path

If Michael wants to push pg-ai-stewards toward the full 11-cycle, the
ordered investments by leverage:

### Phase A — Port what brain v3 already proved (~2-3 sessions)

1. **Substrate Steward (Atonement v1)** — new SQL functions +
   bgworker hook. When a work_item enters status='error':
   - Diagnose the failure type (timeout / model-error / parse-error /
     budget / tool-error)
   - If retryable: enqueue a continuation chat with the failure
     context appended to user message ("the previous attempt failed
     with: …")
   - Track retry count per work_item
   - After N retries: escalate model (e.g. kimi-k2.6 → claude-opus-4.7)
   - After N escalations: quarantine + push to a `findings`-style queue
     for human review

2. **Per-stage gate evaluator (Review v2)** — replace blind
   auto-advance with `evaluate_gate(work_item, stage)`. The evaluator
   is itself an agent (lightweight, no tools, JSON-only output:
   `{action: advance|revise|surface, reasoning, feedback?}`). On
   `revise`, the stage re-runs with feedback injected. On `surface`,
   work_item moves to `awaiting_review` (already a status we've used).

3. **Per-stage circuit breaker** — substrate-side. New table
   `stewards.stage_health(pipeline, stage, failure_window, threshold,
   state)`. When too many failures in a window, the bgworker refuses
   to dispatch that stage until the operator resets it.

### Phase B — Add the per-run intent + spec layer (~1-2 sessions)

4. **First-class intent vertex (Intent v1)** — promote
   `binding_question` from buried jsonb to a proper substrate row:
   `stewards.intents(id, statement, scope, success_criteria,
   created_at)`. work_items reference an intent_id. The intent's
   `success_criteria` becomes the input to the verify step.

5. **Per-run spec / scenarios (Spec v1)** — borrow brain v3's
   `GenerateScenarios()`. Add a `spec` stage to the study-write
   pipeline that produces `stewards.scenarios(work_item_id,
   description, criterion)` rows. The review stage then becomes
   `EvaluateAndVerify` — pass/fail per scenario.

### Phase C — Covenant + Sabbath (~1-2 sessions)

6. **Per-session covenant artifact (Covenant v1)** — when a work_item
   is created, write `stewards.covenants(work_item_id, human_commits,
   agent_commits, created_at)`. Default templates exist; UI lets
   Michael edit before dispatch. Covenant text is included in the
   composed system prompt at every chat dispatch.

7. **Sabbath agent + cadence** — new pipeline `sabbath` that runs
   weekly (or after N completed work_items). Agent reads the recent
   work, writes a reflection to `stewards.studies` with kind='sabbath'.
   Reflection is composed into future agent system prompts via a
   dedicated section. Watchman could trigger it.

### Phase D — Progressive trust + Zion (~2-3 sessions)

8. **Stewardship levels (Stewardship v2)** — add
   `agents.trust_level int` (1-4 per the guide). Tool grants gate by
   trust level. Successful work_items raise trust; quarantined ones
   lower. Substrate-controlled, not operator-controlled.

9. **Substrate intent layer (Zion v1)** — promote
   `intent.yaml`-equivalent into a substrate vertex
   `stewards.purpose(version, statement, values, constraints)`. All
   agents compose their system prompt from `purpose + role +
   stage_specific`. Single source of truth for shared purpose.

10. **Council-based multi-agent coordination (Zion v2)** — defer until
    we have multiple distinct agent families running at once. The
    bishop-vs-conductor pattern requires real multi-agent traffic to
    shape correctly.

## What I'd defer indefinitely

- **Step 10 Consecration v2 (purpose-allocated budgets)** — current
  per-work-item token_budget covers the practical need. The
  "X% to study, Y% to lesson" framing is elegant but adds bureaucracy
  without obvious win at single-user scale.

- **Multi-tenant Zion** — until pg-ai-stewards hosts more than one
  user, the council pattern with distinct stewardships per person
  doesn't have a real shape. Defer until we know.

## Open questions for Michael

1. **Which phase first?** Phase A (port brain v3 patterns) is the
   highest leverage — closes Atonement and the orchestration gap.
   Phase B (intent + scenarios) most directly answers your "are we
   putting in intent at the start of each session?" question. Phase
   C (covenant + sabbath) is the deepest gap but lowest immediate
   reward. Pick one, or sequence?

2. **Substrate steward as a bgworker module or a separate agent?**
   Brain v3 had it as a separate Go process. Substrate could host it
   as a SQL-driven loop the bgworker runs each tick (~500ms) — every
   error triggers diagnosis + decision in-substrate. Cleaner.

3. **Gate evaluator: which model?** Brain v3 used haiku-pinned for
   verify. Substrate could use a lightweight kimi or even local qwen
   for gate evaluation. Trade-off: faster + cheaper vs. higher-quality
   judgment. Worth a small experiment.

4. **The "0 dirty docs" goal.** You raised this on the watchman side —
   currently soak does ~5 docs/week and dirty queue sits at ~50.
   Should the soak cadence go up (every 6h instead of hourly with
   smaller batches)? Or should "dirty" definition narrow (only
   docs touched by something that warrants a re-pass)? Both?

5. **Substrate-side covenant text — generic template or per-pipeline?**
   The covenant.yaml at project level is bilateral and bound to the
   whole project. A per-work-item covenant could be a
   pipeline-specific template (study-write covenant differs from
   teaching covenant) or a single substrate-wide one. Probably
   per-pipeline.

This proposal is ready when you are. I'd recommend starting with
**Phase A.1 (Substrate Steward / Atonement v1)** if you want one place
to start — it has the highest leverage on what's currently a real
gap (failed work_items just sit there), it ports proven brain v3
patterns rather than inventing, and it's bounded enough to ship in a
single session.
