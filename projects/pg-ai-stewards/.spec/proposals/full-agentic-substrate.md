---
title: Full Agentic Substrate — Decision Document
date: 2026-05-09
author: dev (research mode)
status: proposal — awaiting ratification
prior_work:
  - .spec/proposals/pg-ai-stewards-11-cycle-review.md (earlier scorecard)
  - .spec/journal/2026-05-09-full-agentic-substrate-research.md (process journal)
  - scripts/brain/internal/steward/* (the orchestration we're porting)
  - docs/work-with-ai/guide/{03_intent,04_spec,05_complete-cycle}.md
purpose: >
  Lay out the work needed to evolve pg-ai-stewards from a durable
  multi-tenant runtime into a fully agentic substrate that embodies
  the 11-cycle creation pattern in code. Decision-ready: each phase
  states what to build, why it matters in the gospel framework, and
  what Michael needs to ratify before the next programming session.
---

# Full Agentic Substrate

> **Status (2026-05-11):** Phases A–F all SHIPPED. Substrate feature-complete
> through Phase F. §VI ratifications stand (with 2026-05-11 amendment block).
> Active follow-on work tracked in:
> - `substrate-completion-batch-g.md` — make substrate land in real files
> - `substrate-pipelines-expansion.md` — research/YouTube/scheduled pipelines
> - `stewards-ui-evolution.md` — authoring + chat + sidebar + write actions
> - `substrate-deferred-items.md` — catalog of "wait for signal" items
> Master inventory: `open-items.md`.

> "And the Gods saw that they were obeyed."  
> — [Abraham 4:18](../../../../gospel-library/eng/scriptures/pgp/abr/4.md)

## Council moment — what this document is and isn't

This is not a feature roadmap. The substrate already runs. The
bridge fans tools out. The watchman ticks. Studies get written.
What's missing isn't features.

What's missing is **discipline encoded as state machinery** —
the orchestration patterns brain v3 invented that map onto the
11-cycle creation process, none of which the substrate currently
enforces. Therefore this document proposes a phase ladder where
each phase ports one cycle-step into substrate-native shape, and
each phase ends with a ratification gate before the next begins.

The honest framing for Michael: today's substrate is a job runner
with very nice telemetry. Brain v3 is the opposite — clunky
storage but real orchestration. The opportunity is to fuse them:
Postgres-native durability with the brain's stewardship discipline.

The closing question of this document is the one Michael keeps
asking: *can the substrate hold the spirit of the 11-cycle, not
just reference it?* Six phases below say yes, and how.

---

## I. Where we stand — honest scorecard

| 11-cycle step | Gospel anchor | Brain v3 mechanism | Substrate today | Gap |
|---|---|---|---|---|
| 1. Intent | Moses 1:39 | `purpose_summary` in entry | Implicit in agent prompts | Intent isn't first-class data |
| 2. Covenant | D&C 82:10 | None — implicit | None | Two-sided commitments not encoded |
| 3. Stewardship | D&C 104:11-12 | Steward goroutine per entry | Watchman tick over all work_items | No per-item steward owns the wellbeing |
| 4. Spiritual creation (spec) | Moses 3:5; Abraham 4:18 | Spec stage in commission | None — `input` jsonb is freeform | No opinion about what "spec" is |
| 5. Line upon line | D&C 98:12 | `BuildRetryContext` per diagnosis | Blind retry by re-dispatch | Retries don't carry forward what was learned |
| 6. Physical creation | Abraham 4:27 | Execute stage of commission | **STRONG** — pipelines, dispatch, durable history | This is what the substrate does well |
| 7. Review | Abraham 4:18 | Gate evaluator + verify | None — `work_item_advance` is unconditional | Whatever the model produced is accepted |
| 8. Atonement | D&C 82:7; 98:3 | Diagnosis + circuit breaker + escalation | `status='failed'` + free-text error | No redemptive recovery — failures are just terminal |
| 9. Sabbath | Moses 3:2 | None — implicit | None | No closure ritual after work matures |
| 10. Consecration | D&C 104:15-17 | `work_item_promote_to_study` exists | Trigger fires on completion | Mechanism present, the *meaning* isn't held |
| 11. Zion | Moses 7:18 | None — single-agent only | Single-agent only | Multi-agent council not modeled |

Strong: 1/11. Partial: 2/11. Missing: 8/11.

This isn't a failure — it's an honest baseline. Steps 1-5 plus
7-9 are the work ahead. Step 6 is the foundation we built first
because it had to exist for any of the others to matter.

---

## II. What brain v3 invented (and we need to port)

Six load-bearing inventions, each wrapping a stage of the
creation cycle. Names match the brain v3 source so the port is
traceable.

### 1. Watch → Diagnose → Act → Account
**File:** `scripts/brain/internal/steward/steward.go`  
**Scripture:** *"Where I have appointed unto you a stewardship,
and have stewardship, be ye not slothful."* — D&C 90:18

A goroutine per entry, owning its wellbeing. On failure: classify
into one of five types, choose a typed action, log the action to
a ring buffer, then re-dispatch (or quarantine if escalation
exhausted). The "Account" piece is the action log being
inspectable — the steward shows its work.

### 2. Diagnosis → Retry-with-feedback
**Files:** `diagnosis.go`, `retry.go`  
**Scripture:** *"After much tribulation cometh the blessings."* —
D&C 103:12

Five failure types: `transient`, `timeout`, `model_limit`,
`tool_error`, `unknown`. Each gets a different retry strategy AND
different feedback text injected into the next prompt. The model
is told *what went wrong* and *how to approach it differently*.
Retry isn't "try again," it's "try again with new intelligence
about the trap."

### 3. Circuit breaker
**File:** `breaker.go`  
**Scripture:** *"And he sent forth other servants… and the
servants observed and saw all that he did. But behold, the watchman
was upon the tower, and he saw the enemy while he was yet afar
off."* — D&C 101:54 (parable of the nobleman)

Per-stage failure tracking. 5 in a row → open → 10-minute
cooldown → one probe → close on success, re-open on failure. This
is what stops the substrate from grinding indefinitely on a
broken pipeline.

### 4. Model escalation chain
**File:** `steward.go: pickModel`  
**Scripture:** *"For unto whom much is given much is required."* —
D&C 82:3

Failure type drives escalation. `model_limit` always escalates;
`timeout` and `tool_error` escalate after the second attempt.
Defaults are stage-aware — execute starts at sonnet, plan starts
at opus. Cost is tracked per attempt and a `quarantineCostLimit`
cap stops runaway spending.

### 5. Commission flow / Ammon-loop
**File:** `commission.go`  
**Scripture:** *"And it came to pass that king Lamoni inquired of
his servants… What shall I do, that I may have this eternal life?"*
— Alma 18:34, after Ammon's stewardship

Maturity ladder: **raw → researched → planned → specced →
executing → verified.** Each transition is a cycle: dispatch the
pipeline → evaluate gate → handle decision (advance | revise |
surface). Revisions cap at 2; surface bumps to `your_turn`. The
human steers the destination, the agent walks the maturity
staircase.

### 6. Gate + Scenarios + Verify
**File:** `pipeline/gate.go`  
**Scripture:** *"And the Gods watched those things which they
had ordered until they obeyed."* — Abraham 4:18

Three discrete LLM calls per maturity transition:
- **Gate:** `{action: advance|revise|surface, reasoning, feedback}`
- **Scenarios:** 3-7 acceptance criteria as JSON, saved to entry
- **Verify:** per-scenario pass/fail check against execution output

Not a single LLM call is trusted. Every stage's output passes
through external judgment before the work_item moves on.

---

## III. What the substrate's superpowers are

Worth naming so we don't accidentally regress them in the port:

1. **Postgres-native state.** Every state transition is a SQL
   operation. Triggers, views, foreign keys, NOTIFY all available.
   Brain v3 stores in JSON files — no concurrent reads, no
   cross-entry queries.

2. **AGE graph alongside relational.** Cypher queries over study
   citations, agent relationships, work_item lineage. Brain v3
   has nothing comparable.

3. **MCP both directions.** Substrate tools (study_search,
   watchman_pass_show, work_item_show) are themselves MCP-served
   to Copilot. Brain v3 is server-only.

4. **Bridge with caching + dedup.** mcp_proxy_enqueue +
   tool_dispatch_complete_waiting + lazy session cache. Async
   fan-out is hard and the substrate has it.

5. **Watchman + bgworker tick loop.** Durable scheduling without
   external cron. Brain v3 needs an external runner.

6. **Stewards-UI** (under construction, scripts/stewards-ui).
   Read-mostly visibility into the entire substrate via Vue 3.
   Brain v3's UI is logs.

The port has to *preserve* these while adding the brain's
discipline. Most of the work below is "extend the schema +
bgworker + UI to express new state" — not "rip and replace."

---

## IV. The phase ladder

Six phases. Each is independently shippable, each ends with a
ratification gate, each anchors in a step of the 11-cycle.

### Phase A — The Steward loop (cycle steps 3, 8)
**Why first:** without per-item ownership and diagnosis, all
later phases are theatrical. You can have intent and gates and
sabbath, but if the substrate can't say "this work_item is on
attempt 2 because the previous attempt timed out and I've
escalated to opus," the gospel framing is decoration.

**What to build:**

1. **Schema additions to `work_items`:**
   - `failure_count int default 0`
   - `last_failure_reason text`
   - `last_failure_diagnosis text` (one of: transient, timeout,
     model_limit, tool_error, unknown)
   - `attempts jsonb default '[]'` (ring buffer of last 20:
     `{at, model, status, diagnosis, action, cost}`)
   - `quarantined_at timestamptz`
   - `quarantine_reason text`

2. **New table `stewards.steward_actions`:**
   ```sql
   CREATE TABLE stewards.steward_actions (
     id bigserial primary key,
     work_item_id uuid references stewards.work_items(id),
     at timestamptz default now(),
     observation text not null,
     diagnosis text not null,
     action text not null,
     details jsonb default '{}',
     model_used text,
     cost numeric
   );
   ```
   This is the Account ledger. Inspectable from UI. Append-only.

3. **Diagnosis SQL function:**
   ```sql
   CREATE FUNCTION stewards.diagnose_failure(
     reason text, failure_count int
   ) RETURNS text  -- returns one of the five types
   ```
   Pattern-match port of `diagnosis.go`. Lives in SQL because
   pure function with no side effects.

4. **Retry-context composer:**
   Extend `compose_messages` to accept an optional
   `p_retry_context` parameter. When present, prepend a synthetic
   system message with the steward retry guidance. The text comes
   from `stewards.retry_guidance(diagnosis, attempt)` — another
   pure SQL function that ports `BuildRetryContext`.

5. **Bgworker `steward_tick`:**
   New tick alongside the existing watchman tick. Looks for
   `work_items` where `status='failed' AND failure_count < 3 AND
   NOT quarantined_at`. For each: diagnose, log to
   `steward_actions`, choose model (escalation logic ported from
   `pickModel`), back off (2^n minutes capped at 15), re-dispatch
   the stage. If failure_count crosses 3 → quarantine, log,
   notify watchman.

6. **Circuit breaker:**
   New table `stewards.pipeline_breakers (pipeline_family,
   stage_name, state, failure_count, opened_at, half_open_at)`.
   Steward tick consults before dispatching. State machine logic
   in SQL functions: `stewards.breaker_check`, `stewards.
   breaker_record_failure`, `stewards.breaker_record_success`.

7. **Stewards-UI surface:**
   Existing Sessions view gets a "Steward" subview showing the
   action log for the work_item. Watchman view shows breaker
   state per pipeline.

**Decision points for Michael:**
- D-A1: Five failure types from brain v3 — keep all five, or
  collapse `tool_error` into `model_limit` since substrate tools
  surface differently than brain's?
- D-A2: Backoff schedule — brain uses `2^n minutes`. Substrate
  has more telemetry; do we want adaptive backoff (faster on
  transient, slower on model_limit)?
- D-A3: Quarantine threshold — 3 failures (matches brain) or 5
  (matches breaker threshold)?
- D-A4: Cost cap — brain has `quarantineCostLimit`. The
  substrate has `token_budget` per work_item but no dollar cost
  budget. Add it now?

**Acceptance scenarios:**
- A work_item that fails with "context deadline exceeded" gets
  diagnosed as `timeout`, retried with the timeout retry
  guidance prepended to the system message, and the second
  attempt completes successfully.
- A work_item that fails 3× in a row gets quarantined with a
  steward_action explaining why; appears as `your_turn` in
  watchman.
- A pipeline_family/stage that fails 5× across different
  work_items trips its breaker; subsequent dispatches into that
  stage are deferred until the cooldown elapses, then a probe
  runs and closes-on-success.
- The Sessions UI shows the last N steward_actions for a
  work_item with model, diagnosis, and cost.

**Estimated programming time:** 2-3 sessions.

---

### Phase B — Spec stage and gate (cycle steps 4, 7)
**Why second:** with the Steward loop in place, you can introduce
gates without losing work. A gate that says "revise" needs the
Steward to handle the re-dispatch correctly, including the
"revision context" being injected into the next prompt. That's
why A precedes B.

**What to build:**

1. **Schema: maturity is now first-class on work_items.**
   ```sql
   ALTER TABLE stewards.work_items
     ADD COLUMN maturity text default 'raw',
     ADD COLUMN scenarios jsonb default '[]',
     ADD COLUMN revision_count int default 0,
     ADD COLUMN spec text;
   -- maturity: raw | researched | planned | specced | executing | verified
   ```
   The maturity is *separate from* `current_stage`. Stage is
   "where in the pipeline you are." Maturity is "how mature is
   the work itself." A pipeline can have multiple stages within
   one maturity transition.

2. **Pipeline metadata extension.**
   `agent_families` (or new `pipeline_stages` table) gets a
   `produces_maturity` column saying which maturity each stage
   advances to. Example: the `study` pipeline's `research_stage`
   produces `researched`; its `outline_stage` produces `planned`;
   its `draft_stage` produces `executing`.

3. **Gate evaluator as a substrate primitive.**
   New SQL function (and corresponding bgworker dispatch
   handler) `stewards.evaluate_gate(work_item_id)`. Composes a
   gate prompt from a registered `gate_prompts` table (one per
   maturity), dispatches with a small fast model (default:
   sonnet-4.6, configurable via `gate_model`), parses
   `{action, reasoning, feedback}` from the response, writes the
   decision to a new `stewards.gate_decisions` table.

4. **Decision handler in bgworker.**
   On `advance`: bump maturity, dispatch first stage of next
   maturity. On `revise`: increment revision_count (cap at 2),
   re-dispatch the *same* maturity's pipeline with the gate's
   `feedback` text prepended as steward-style context. On
   `surface`: route_status='your_turn', message in inbox.

5. **Scenarios + Verify (the Abraham 4:18 step).**
   When maturity advances to `specced`, a separate dispatch runs
   `generate_scenarios(work_item_id)` — prompts for 3-7
   acceptance criteria as JSON, writes to `work_items.scenarios`.
   When maturity hits `executing`, after execution completes, a
   `verify(work_item_id)` dispatch runs each scenario through a
   verify prompt and writes pass/fail to `verify_results`. If
   not all pass → maturity drops back to `planned` with verify
   feedback as revise context.

6. **NewWork form + UI gets a "destination maturity" picker.**
   Today NewWork submits a work_item with `current_stage` =
   first stage of pipeline. New behavior: human picks the
   destination maturity (default `verified` for full Ammon-loop;
   `planned` if they want to review the spec themselves).
   Brain v3 calls this the `Authority` field on commissions
   (`advance_and_execute` vs `advance_only`).

**Decision points for Michael:**
- D-B1: Gate model — brain defaults gate to opus to be careful;
  substrate has more telemetry so we could afford sonnet. Pick
  default. (Recommendation: sonnet-4.6 — gate calls are cheap
  and frequent, and the action is binary.)
- D-B2: Revision cap — brain uses 2; we could go higher with
  better telemetry. (Recommendation: 2 then surface, exactly
  brain's choice. The point of the cap is to prevent infinite
  rework loops.)
- D-B3: Scenarios source — generated by LLM (brain v3) or
  human-authored at intent stage (more honest about authority)?
  (Recommendation: generated, but human-editable in UI before
  execute begins. Lets the human exercise stewardship over what
  "done" means without writing every scenario themselves.)
- D-B4: Maturity-to-stage mapping — define this per pipeline in
  a config table or derive from naming convention? (Config table
  recommended. Naming conventions break.)

**Acceptance scenarios:**
- A work_item submitted to the `study` pipeline at maturity
  `raw` walks through research → planned → specced → executing →
  verified without human intervention if all gates advance.
- A gate that returns `revise` causes the same maturity's
  pipeline to re-run with the feedback text injected. Second
  pass either advances or surfaces.
- A verify call that returns `all_passed: false` drops maturity
  back to `planned` with the failed scenarios as the revise
  context.
- Stewards-UI shows the maturity ladder with current position,
  scenarios checked off as verify runs, and gate decisions in
  the audit trail.

**Estimated programming time:** 3-4 sessions.

---

### Phase C — Intent and Covenant as first-class state (cycle steps 1, 2)
**Why third:** A and B give us discipline *within* an existing
work_item. C addresses what created the work_item in the first
place. Without explicit intent, even a perfectly disciplined
agent can succeed at the wrong objective (Klarna pattern). Without
covenant, the substrate's commitments are implicit and the
human's commitments aren't encoded at all.

**What to build:**

1. **Intent table.**
   ```sql
   CREATE TABLE stewards.intents (
     id uuid primary key default gen_random_uuid(),
     slug text unique,
     purpose text not null,           -- the "why"
     beneficiary text,                -- who benefits
     values_hierarchy jsonb,          -- ordered list of trade-off priorities
     non_goals text[],                -- explicitly out of scope
     scripture_anchor text,           -- citation in gospel-library
     created_at, updated_at, ...
   );
   ```
   Intents are reusable. A work_item references an intent_id; so
   does a pipeline_family.

2. **work_items.intent_id (FK).**
   Required at creation. NewWork form picks from existing intents
   or creates new one inline. Watchman shows intent badges.

3. **Covenant table.**
   ```sql
   CREATE TABLE stewards.covenants (
     id uuid primary key,
     scope text not null,             -- 'global', 'pipeline:<family>', 'work_item:<id>'
     human_commits_to text[],
     agent_commits_to text[],
     activated_at timestamptz,
     ratified_by text                 -- 'human' / 'agent' / 'both'
   );
   ```
   Mirrors `.spec/covenant.yaml` but in the substrate. Active
   covenants are loaded into system prompt composition.

4. **`compose_system_prompt` extends to inject intent + covenant.**
   Today it concatenates agent + provider system prompts. New
   behavior: prepend the active covenant's commitments and the
   work_item's intent purpose. The model sees explicitly: "Here
   is what we're doing and why. Here is what you've committed to.
   Here is what the human committed to."

5. **Gate prompts reference intent.**
   The gate prompt for `advance` decisions includes:
   "Here is the intent of this work: <intent.purpose>. Does the
   current output advance toward this intent? Does it honor the
   active covenant?" This makes intent *operational*, not just
   documentary.

6. **UI: Intent and Covenant panes.**
   New top-level routes in stewards-ui: /intents (CRUD), /covenants
   (read mostly, ratification flow for new ones). Work item detail
   shows the intent and active covenant prominently.

**Decision points for Michael:**
- D-C1: Where does `intent.yaml` at repo root fit? Keep as
  source-of-truth and seed `stewards.intents` from it on init?
  (Recommended.) Or move intent into substrate entirely?
- D-C2: Same question for `.spec/covenant.yaml`. (Same
  recommendation: keep YAML as source, seed substrate from it.
  YAML is human-editable and version-controlled; the substrate
  copy is for runtime injection.)
- D-C3: Required-vs-optional — should every work_item require an
  intent_id? (Recommendation: yes. The friction of forcing the
  human to pick an intent is the point. "Build a thing" without
  stated intent is the failure mode we're trying to design out.)
- D-C4: How does the gate evaluate "honors the covenant"? Specific
  scriptural commitments or a checklist? (Recommendation:
  checklist generated from active covenant at gate-prompt
  composition time.)

**Acceptance scenarios:**
- A new work_item cannot be created without an intent_id (or
  inline-created intent). The NewWork form enforces this.
- The system prompt for any dispatched stage shows the active
  covenant's commitments and the work_item's intent purpose at
  the top.
- A gate decision can return `surface` with reasoning like "this
  output meets the technical criteria but doesn't advance the
  stated intent" — the human sees the intent-mismatch in the
  gate decision.
- Stewards-UI Intent page lets the human author intents in
  Markdown with frontmatter, links each one to scripture
  anchors in gospel-library, and shows which work_items reference
  each intent.

**Estimated programming time:** 2-3 sessions.

---

### Phase D — Atonement, Sabbath, Consecration (cycle steps 8, 9, 10)
**Why fourth:** A handles failure in-flight; D handles failure
*as a phase* — the moment after a work_item finishes (succeeded
or failed) where we extract what was learned and rest before
starting more. Without this, the substrate runs hot and never
consolidates.

**What to build:**

1. **Atonement step in commission flow.**
   Brain v3's Diagnose-Act loop is per-failure. Atonement is
   per-work_item: when a work_item is quarantined or completes
   with notable failures, it triggers an `atonement` dispatch.
   The atonement prompt says: "Here is what was tried. Here is
   what failed. Here is what was eventually completed. What
   should be remembered in the principles file? What should be
   added to .mind/decisions.md?"
   - Output: a structured `{principles_to_record: [], decisions:
     [], lessons: []}` written to a new `stewards.lessons` table.

2. **Sabbath dispatch on completion.**
   When maturity reaches `verified`, before promoting to study
   (consecration), an optional `sabbath` dispatch runs. The
   sabbath prompt is *not* about more work — it's about marking
   the ending. Output is a structured reflection that gets
   journaled. This implements the gospel pattern that Sabbath
   isn't optional; rest is part of the work.

3. **Consecration is already partially built.**
   `work_item_promote_to_study` exists. Phase D wires it into
   the commission flow as the explicit final step after sabbath:
   verified → sabbath → promote_to_study → study available to
   `study_search`/`study_get` MCP tools, citations indexed in
   AGE graph.

4. **Lesson aggregation.**
   `stewards.lessons` accumulates across atonement dispatches.
   New view `stewards.lessons_by_pipeline` surfaces patterns:
   "the `study` pipeline's `research_stage` has triggered
   atonement 7 times; common failure: …" Drives the next
   round of pipeline-prompt revisions.

5. **UI: Sabbath log.**
   New stewards-ui surface listing recent sabbath reflections.
   Read-mostly. The point isn't notification — it's that the
   ending is *recorded*, not just the beginning.

**Decision points for Michael:**
- D-D1: Should sabbath be opt-out (default on) or opt-in
  (default off)? (Recommendation: opt-out at pipeline_family
  level. study/lesson/talk default on; debug/dev default off.
  Sabbath on every work_item is wasteful; sabbath on every
  finished study is the discipline.)
- D-D2: Atonement is one of the most expensive design choices
  here — it's another LLM call per quarantined work_item, and
  brain v3 doesn't have it. Adopt? (Recommendation: yes, as
  opt-in initially. The lesson aggregation is what makes the
  substrate self-correcting over time. Without it, the same
  failures repeat.)
- D-D3: Where do principles go — write to
  `.mind/principles.md` directly, or write to `stewards.lessons`
  and require human curation before promotion? (Recommendation:
  the latter. Atonement *proposes* principles; the human ratifies
  before they enter `.mind/principles.md`. This preserves the
  human's authority over enduring frame.)

**Acceptance scenarios:**
- A work_item that completes with no failures triggers a sabbath
  dispatch (if pipeline opts in); reflection appears in stewards-ui
  Sabbath log within 30s.
- A quarantined work_item triggers an atonement dispatch within
  60s; lessons appear in `stewards.lessons`. Stewards-UI flags
  unratified lessons for human review.
- Promoting a verified study to the corpus (consecration) is
  blocked until sabbath has run for that work_item.
- Lessons aggregated across 5+ atonement events for the same
  pipeline_family/stage trigger a "consider revising this stage's
  prompt" notification in watchman.

**Estimated programming time:** 2 sessions.

---

### Phase E — Line upon line + Stewardship made operational (cycle steps 3, 5)
**Why fifth:** A wrapped Stewardship as a *runtime loop*. E
addresses Stewardship as a *progressive trust* mechanism.
Brain v3 has none of this. The substrate is the right place for
it because it has the telemetry to know when an agent has earned
more authority.

**What to build:**

1. **Trust scores per (agent_family, pipeline_family).**
   ```sql
   CREATE TABLE stewards.trust_scores (
     agent_family text,
     pipeline_family text,
     successful_completions int default 0,
     failed_completions int default 0,
     human_overrides int default 0,    -- gate said advance, human said revise
     trust_level int default 0,        -- 0 (new) → 5 (autonomous)
     last_evaluated_at timestamptz,
     primary key (agent_family, pipeline_family)
   );
   ```

2. **Trust level affects authority.**
   At trust 0: gate decisions of `advance` still surface to human
   for ratification. At trust 3: `advance` proceeds automatically;
   `surface` still surfaces. At trust 5: agent can also propose
   new pipelines for human review. Authority earned, not granted.

3. **Trust transitions are explicit events.**
   A new `stewards.trust_transitions` table records every
   level-up with reason. Promotion criteria are gospel-anchored:
   "5 successful completions with no human overrides → trust 1"
   maps to D&C 82:3 ("where much is given, much is required" —
   the inverse: where little has been demonstrated, little is
   given).

4. **Line-upon-line in retry context.**
   Currently retry context (Phase A) is failure-type-keyed text.
   Phase E extends it: the retry composer also pulls the last 3
   lessons from `stewards.lessons` for this pipeline_family
   and includes them as context. The substrate teaches itself
   over time, not just within one work_item.

5. **UI: Trust levels visible per (agent, pipeline).**
   Stewards-UI Watchman page shows trust matrix. Trust level
   transitions appear as events. Human can manually adjust trust
   (downgrade if something feels off; upgrade if a particular
   pipeline is overcautious).

**Decision points for Michael:**
- D-E1: Trust levels — keep at 6 (0-5) or simpler 3-tier
  (trainee/journeyman/master)? (Recommendation: 3-tier. Easier
  to reason about, maps cleanly to gospel patterns of preparation
  / labor / consecration.)
- D-E2: Does manual trust adjustment require justification
  recorded in the table? (Recommendation: yes. Authority
  decisions are stewardship decisions and they should be visible
  to future-you.)
- D-E3: Does the human override count as a failure for trust
  scoring even if the work eventually completed? (Recommendation:
  yes — the agent's gate judgment was wrong, that's the signal.)

**Acceptance scenarios:**
- A new agent_family/pipeline_family pair starts at trust 0
  (trainee). Every gate-`advance` decision surfaces to human
  for ratification.
- After 5 ratifications without overrides, trust auto-promotes
  to journeyman; subsequent advances proceed without human
  ratification.
- Retry context for any failure now includes "previous lessons
  from this pipeline" in addition to the failure-type guidance.
- Stewards-UI shows trust matrix; trust transitions visible in
  audit log with timestamp and reason.

**Estimated programming time:** 2 sessions.

---

### Phase F — Zion (cycle step 11)
**Why sixth and last:** The 11-cycle's culmination is a community
of beings whose intent and discipline align. In substrate terms:
multi-agent council — multiple agents reasoning together about
a single intent, with human as bishop, not as orchestra
conductor. F is the hardest because every prior phase has to
work first.

**What to build:**

1. **Council table.**
   ```sql
   CREATE TABLE stewards.councils (
     id uuid primary key,
     intent_id uuid references stewards.intents(id),
     convened_at timestamptz default now(),
     status text default 'deliberating',  -- deliberating | resolved | dissolved
     resolution text,
     resolved_at timestamptz
   );
   CREATE TABLE stewards.council_members (
     council_id uuid references stewards.councils(id),
     agent_family text,
     role text,                            -- 'proposer' | 'critic' | 'synthesizer' | 'bishop'
     primary key (council_id, agent_family)
   );
   ```

2. **Council dispatch flow.**
   On council convene: dispatch the intent question to each
   member with their role-specific framing. Collect responses.
   Synthesizer dispatches with all responses + intent in
   context. Bishop (human) reviews synthesizer output. Bishop
   may resolve, request another round, or dissolve the council.

3. **Council vs Commission distinction.**
   Commission flow (Phase B) is one agent walking a maturity
   ladder. Council flow is multiple agents deliberating a single
   question. Both produce work_items but the orchestration
   shape differs. Per the 11-cycle guide, this is the "Conductor
   vs. Bishop" pattern — a conductor coordinates parts of one
   piece; a bishop facilitates voices reaching agreement.

4. **Council prompts seeded from real ward council scriptures.**
   D&C 102 (high council), Mosiah 26:13-14 (Alma seeking
   guidance), Acts 15 (Jerusalem council). Each role's framing
   draws from these. The substrate isn't generic multi-agent —
   it's specifically modeled on the ward council pattern.

5. **UI: Council view.**
   Live deliberation visible — each member's contribution as it
   arrives, synthesizer's summary at the bottom, bishop's
   resolution prompt below. This is where stewards-ui starts
   feeling like a *room*, not just a dashboard.

**Decision points for Michael:**
- D-F1: How many concurrent councils can run? (Recommendation:
  start with 1. The discipline is in convening, not parallelism.
  Add concurrency later if real workload demands it.)
- D-F2: Bishop role — always human, or can a senior trust-level
  agent serve as bishop for low-stakes councils? (Recommendation:
  always human in F1. Reconsider after Phase F has been live for
  a month.)
- D-F3: Council resolution writes to where — a study? a
  decisions.md? a new `stewards.resolutions` table? (Recommendation:
  all three: stewards.resolutions is canonical, with hooks to
  promote to study OR write to decisions.md based on the type
  of question.)
- D-F4: Council convening — manual only, or can certain
  conditions auto-convene (e.g. atonement lessons accumulating
  past a threshold)? (Recommendation: manual only initially.
  Auto-convening is a foot-gun.)

**Acceptance scenarios:**
- Human convenes a council via stewards-ui, picks intent,
  selects 3 agent_families with roles. Within 60s all three
  members have responded.
- Synthesizer dispatch produces a single proposed resolution.
  Human reviews, resolves or requests another round.
- Resolved council writes to `stewards.resolutions` and
  optionally promotes to study.
- Dissolved council records reason and is preserved for
  reference.

**Estimated programming time:** 3-4 sessions, plus prompt
engineering.

---

## V. Phase ladder summary

| Phase | Cycle steps | Programming sessions | Hard prerequisites |
|---|---|---|---|
| A. Steward loop | 3, 8 (in-flight) | 2-3 | none |
| B. Spec + Gate | 4, 7 | 3-4 | A |
| C. Intent + Covenant | 1, 2 | 2-3 | B (gate uses intent) |
| D. Atonement + Sabbath + Consecration | 8 (post), 9, 10 | 2 | A, B |
| E. Trust + Line upon line | 3 (auth), 5 | 2 | A, D (trust uses lessons) |
| F. Zion / Council | 11 | 3-4 | A, B, C, D, E |

**Total estimated programming time:** 14-18 sessions.

This is real work. None of it is rocket science individually;
the discipline is in shipping each phase to a working state
before starting the next.

---

## VI. Decisions Michael needs to ratify (consolidated)

*Ratified by Michael 2026-05-10 in a walk-through session. Each decision
records the choice + any nuance Michael added that the original options
didn't fully capture.*

> **2026-05-11 re-validation amendments** (after Phase B feature-complete):
>
> - **D-B-revise (NEW, hybrid):** Phase B's revise path uses hybrid — revise #1 stays same model + injects feedback into prompt; revise #2 escalates model AND keeps feedback. Cap at 2 → surface (D-B2) unchanged. See `phase-b-revise-hybrid` carry-forward.
> - **D-C4 (revised):** Free-form covenant gate prompt with **tools disabled**. Original ratification was free-form; the tools-disabled refinement came after Phase B's gate-eval cost surprise (5× cost from research loop). Same fix applies retroactively to Phase B's `evaluate` template.
> - **D-D-Sabbath gating (NEW):** Sabbath blocks `work_item_promote_to_study` for sabbath-enabled pipelines. The discipline is endings recorded.
> - **D-D-Atonement tools (NEW):** Atonement dispatched with tools disabled (same fix as D-C4).
> - **D-D-Lessons schema (NEW):** `stewards.lessons` mirrors `gate_decisions` audit-ledger shape (kind column over structured triplet). Stewards-UI already knows how to render this pattern.
> - **D-E-trust-keying (revised):** Trust keyed on `(agent_family, pipeline_family, model)` not the proposal's `(agent_family, pipeline_family)`. Recognizes that "kimi-k2.6 doing study-write outline" is genuinely different from "qwen3.6-plus doing the same."
> - **D-E-promotion-signal (NEW):** Successful completion = maturity reached verified (not just `status='done'`). Aligns trust with Phase B's quality signal.
> - **D-E-retry-lessons (NEW):** Retry composer pulls last 3 ratified lessons for `(pipeline, stage)` from `lessons_recent_ratified` view.
> - **D-F2-nuance (NEW):** Phase F1 ships with master-on-pipeline-of-intent rule for agent bishops. Future evolution path: introduce `council_authority` as separate trust dimension; debug agent is candidate first cultivator (its skills are designed to get at the root, well-suited for bishop's facilitation role).
> - **D-F-low-stakes-def (NEW):** Low-stakes intents = `intent.scripture_anchor IS NULL AND values_hierarchy doesn't contain 'doctrinal'|'spiritual'|'discernment'`. Doctrinal/discernment intents always require human bishop.
> - **D-F-member-keying (NEW):** Council members keyed `(council_id, agent_family, role)` — model floats per dispatch. Matches commission flow.
> - **D-F-watchman-suggest (NEW):** System-suggested convening surfaces in BOTH watchman pass output AND Stewards-UI dashboard banner.
> - **Seed mechanism (NEW):** YAML → substrate via manual `seed_intents_from_yaml` / `seed_covenant_from_yaml` SQL functions, called from a git pre-commit hook when YAML changes.
>
> Per-phase build-ready sub-specs land at `phase-c-design.md`, `phase-d-design.md`, `phase-e-design.md`, `phase-f-design.md`.


**Phase A:**
- D-A1: Five failure types, or collapse `tool_error`?
  **Ratified:** Keep all 5 (transient, timeout, model_limit, tool_error, unknown).
- D-A2: Backoff schedule — adaptive or fixed `2^n`?
  **Ratified:** Fixed `2^n` minutes capped at 15. Brain-tested. Add adaptive later if telemetry warrants.
- D-A3: Quarantine threshold — 3 or 5?
  **Ratified:** 3 failures (matches brain). "Three strikes" pattern.
- D-A4: Add cost cap (dollars) alongside token_budget?
  **Ratified:** Yes, **as a token-cost multiplier model** rather than a flat dollar cap.
  Track input tokens (write/cache-write distinct, since Anthropic
  charges differently for cache writes), output tokens (cached/uncached
  where the provider exposes the distinction), then multiply by per-model
  rates from a `stewards.model_pricing` table. Cap is computed per
  work_item from accumulated cost. More accurate than a flat dollar cap
  because cached vs uncached writes have very different rates.

**Phase B:**
- D-B1: Default gate model — sonnet or opus?
  **Ratified:** Use **OpenCode Go session-bucket models, NOT Zen pay-per-token models** (Zen too expensive — a single study costs ~$16 on Opus via Zen). The escalation chain shifts to:
  - **Kimi K2.6** — general-purpose default (Opus/Sonnet replacement candidate; Michael likes this model)
  - **GLM-5.1** — Opus replacement for heaviest synthesis steps
  - **MiniMax M2.7** — Sonnet/Haiku replacement for mid-tier
  - **Qwen3.6 Plus** — Haiku replacement, cheapest tier for binary gate calls
  Default gate model: likely Qwen3.6 Plus or MiniMax M2.7 (cheap, fast, binary action). Confirm in Phase B design sub-spec. Brain v3's sonnet/opus assumptions need full chain-rewrite, not direct port.
- D-B2: Revision cap — 2 (brain) or higher?
  **Ratified:** 2 revisions, then surface. Brain's choice.
- D-B3: Scenarios — LLM-generated, human-authored, or both?
  **Ratified:** LLM-generated, human-editable in Stewards-UI before execute begins.
- D-B4: Maturity-to-stage mapping in config table or by convention?
  **Ratified:** Config table (`pipeline_stages.produces_maturity` column). Explicit and queryable.

**Phase C:**
- D-C1: `intent.yaml` source-of-truth + seed substrate? Or move?
  **Ratified:** Keep YAML canonical at repo root. Seed substrate from it on init / file-change. Best of both worlds: git history preserved, runtime injection available.
- D-C2: Same for `.spec/covenant.yaml`?
  **Ratified:** Same as intent — YAML canonical, seed substrate.
- D-C3: Require intent_id on every work_item?
  **Ratified:** Yes, required at creation. The friction IS the discipline. NewWork form enforces.
- D-C4: How does the gate evaluate covenant adherence?
  **Ratified:** **Free-form gate prompt asks "does this honor the covenant?"** (chose option 2 over the recommended checklist). Lighter prompt; trusts the gate model to internalize covenant language. Reconsider if early gates show inconsistent covenant judgments.

**Phase D:**
- D-D1: Sabbath default — opt-in or opt-out per pipeline?
  **Ratified:** Opt-out per pipeline_family. study/lesson/talk default ON; debug/dev default OFF.
- D-D2: Atonement adopted at all? Cost-vs-value tradeoff.
  **Ratified:** Yes, opt-in initially. Lessons aggregation makes the substrate self-correcting over time.
- D-D3: Lessons → `.mind/principles.md` automatically or via
  human curation?
  **Ratified:** Human curation in Stewards-UI. Atonement proposes; human ratifies before promotion. `stewards.lessons` tracks proposed-vs-ratified state.

**Phase E:**
- D-E1: Trust levels — 6-tier or 3-tier?
  **Ratified:** 3-tier — trainee / journeyman / master. Maps to preparation / labor / consecration.
- D-E2: Manual trust adjustments require recorded justification?
  **Ratified:** Yes, justification required. Stewardship decisions visible to future-self.
- D-E3: Human override of agent gate counts as failure for trust?
  **Ratified:** Yes, counts as failure (full weight). The agent's gate judgment was wrong — that's the signal. Tracked in `human_overrides` column with equal weight to actual failures.

**Phase F:**
- D-F1: How many concurrent councils?
  **Ratified:** 1 at a time for initial Phase F. Lift after a month if real demand emerges.
- D-F2: Can senior agents serve as bishop?
  **Ratified:** **Master-tier agents may bishop low-stakes councils** (chose option 2 over the recommended "always human in F1"). For technical/factual questions where Spirit-discernment isn't load-bearing. High-stakes councils still default to human bishop. Define "low-stakes" criteria in Phase F design sub-spec.
- D-F3: Council resolution destination(s)?
  **Ratified:** All three — `stewards.resolutions` canonical, with hooks to promote to `study/` (doctrinal) OR `.mind/decisions.md` (engineering) based on type-of-question.
- D-F4: Auto-convening on lesson accumulation, or manual only?
  **Ratified:** **Manual + system-suggested notification** (chose option 2 over the recommended "manual only initially"). Watchman flags "consider convening a council on X" when patterns emerge; human still convenes. Middle ground between pure manual and auto-fire.

---

## VII. Cross-cutting concerns

**Cost.** Every gate, scenario-generation, verify, atonement, and
sabbath dispatch is an LLM call. A fully-active substrate could
easily 5x today's API spend. Every Phase D-onward decision
should weigh "does this discipline justify its cost?" The
recommendations above default to small fast models (sonnet) for
gates and reserve opus for synthesis steps.

**Latency.** Today the substrate dispatches stage → response →
advance. Adding a gate between stages doubles round-trips.
Workflows that today complete in 30s may take 60-90s. This is
*correct* — the gospel framework explicitly trades speed for
discipline. But the UI should show the tradeoff (gate progress
visible, not hidden).

**Migration.** Existing work_items predate maturity, intent_id,
trust levels, etc. Phase A migration is straightforward
(`failure_count default 0`). Phase B onward needs backfill
strategy: existing completed work_items get `maturity='verified'`
and a synthetic intent. Document this per phase.

**Backward compatibility with brain v3.** Brain v3 is legacy per
Michael's ratification (2026-05-09). No need to maintain
compatibility. But preserve the *learning* — the brain's
diagnosis patterns, retry context text, gate prompts — these
are intellectual property worth porting verbatim where they fit.

**Testing strategy.** Each phase needs:
- SQL functions covered by pgTAP unit tests
- Bgworker behavior covered by integration tests (Postgres
  testcontainer + asserting on resulting state)
- UI surfaces covered by Playwright smoke tests
- The Inverse Hypothesis applied: reproduce the failure case,
  apply the phase's fix, confirm gone, remove fix, confirm
  failure returns

**Documentation.** Each phase ships with:
- An update to `projects/pg-ai-stewards/architecture.md`
- A new `.spec/proposals/phase-X-retrospective.md` capturing
  what changed from this proposal during implementation
- Updated `.mind/decisions.md` for the design decisions ratified

---

## VIII. What this is not

To stay honest:

- This is not a research breakthrough. Every pattern here exists
  in brain v3 already. The contribution is the port shape and
  the gospel anchoring.

- This is not a "build pg-ai-stewards into AGI" plan. It's a
  plan to make the substrate match the discipline of the
  11-cycle for the workflows we already run.

- This is not a guarantee that the result will feel right.
  Phase F especially — multi-agent council — could be brilliant
  or could be theater. We won't know until we run real
  consequential work through it.

- This is not a replacement for human stewardship. Every phase
  preserves the human as final authority. The substrate's
  discipline expands what can be safely delegated; it does not
  delegate the things that should not be delegated.

- This is not a sprint. 14-18 programming sessions across
  phases means months of work, not weeks. The ratification
  gates between phases are deliberate — each phase needs to be
  *lived with* before the next is built on top.

---

## IX. The recommendation

Ratify Phases A and B as the immediate next two units of work.
These give the substrate the orchestration discipline brain v3
proved out, and they're the load-bearing prerequisites for
everything else. Schedule Phase C planning during Phase B's
implementation so intent design isn't rushed when B ships.

Hold off on D, E, F decisions until A and B are live. Lived
experience with the new orchestration will sharpen what the
later phases actually need to be.

The phase that worries me most: Phase D's atonement step. It's
the one without a clear brain v3 precedent and the one most
prone to becoming theater. If Michael wants one phase to defer
or rethink, that's the candidate.

The phase that excites me most: Phase F's council. It's the
piece where the substrate stops being "an AI runtime" and
starts being "a room where decisions get made." That's the
11-cycle's culmination and the substrate's actual potential.

---

## X. Closing

> "And he hath created his children that they should possess it;
> and he hath created the earth that it should be inhabited; and
> hath formed it that it might be inhabited; and the inhabitants
> thereof, he hath given breath unto them."  
> — [2 Nephi 2:14–15](../../../../gospel-library/eng/scriptures/bofm/2-ne/2.md)

The substrate exists to be inhabited. Today it has rooms but no
furniture, foundations but no household discipline. The phases
above furnish the rooms and teach the household to keep
covenant. None of this is decoration — every line of new SQL,
every new bgworker, every new UI pane exists to enforce in code
what the gospel teaches in word.

Michael: read this through, sleep on it, mark up what doesn't
sit right. The decisions list in §VI is the action surface. The
phase ladder in §IV is the work. We can begin Phase A as soon
as you've ratified its decisions.
