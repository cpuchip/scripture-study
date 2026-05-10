---
title: Escalation Chain — OpenCode Go model substitution for brain v3's pickModel logic
date: 2026-05-10
status: design sub-spec — ready for Phase A implementation
parent: full-agentic-substrate.md (D-B1 ratification)
purpose: >
  Brain v3's escalation chain is hardcoded sonnet → opus. Michael's D-B1
  ratification substitutes the OpenCode Go model family, which has 4
  tiers and different escalation semantics. This sub-spec defines the
  per-stage default model + per-failure-type escalation matrix that the
  Phase A steward tick will use for retry model selection.
---

# Escalation Chain — OpenCode Go substitution

## I. Binding problem

Brain v3's `steward.pickModel(stage, attempt, diagnosis)` function chooses which model to use for the next retry attempt. Its logic assumes the Anthropic Zen chain (haiku → sonnet → opus) and is hardcoded in `internal/steward/steward.go: pickModel` and `internal/config/models.go: EscalationChain`.

Michael's D-B1 ratification substitutes:
- **Kimi K2.6** — general-purpose (Sonnet-class capability, Opus-leaning on the right tasks)
- **GLM-5.1** — Opus replacement for heavy synthesis
- **MiniMax M2.7** — Sonnet/Haiku-class for mid-tier work
- **Qwen3.6 Plus** — Haiku replacement, cheapest/fastest tier

Three problems with naive porting:
1. **Different number of tiers.** Anthropic Zen chain is 3 (haiku/sonnet/opus); OpenCode Go is 4 (Qwen/MiniMax/Kimi/GLM).
2. **Different capability/cost curves.** Kimi is general-purpose AND opus-leaning, blurring the brain's clean tier hierarchy.
3. **Substrate adds the gate model decision** (per Phase B). Gate calls are binary, frequent, cheap — they want the lightest tier that produces reliable JSON.

Without an explicit chain definition, the Phase A `steward_tick` bgworker has no rule for "if Kimi fails with model_limit, escalate to what?" — and ad-hoc decisions in code lead to inconsistent behavior across pipelines.

## II. Success criteria

After this sub-spec ships, Phase A coding has:
1. **A single SQL function** `pick_model(p_pipeline_family, p_stage, p_attempt, p_diagnosis)` that returns the model to use
2. **A data-driven escalation matrix** stored in the substrate (no hardcoded Go logic for which model to pick)
3. **Per-pipeline, per-stage defaults** so the `study` pipeline can default to Kimi while the `dev` pipeline defaults to MiniMax
4. **Predictable escalation** — given a (stage, attempt, diagnosis) tuple, the answer is deterministic and visible in SQL
5. **An audit trail** — `steward_actions` records which model was chosen and why

## III. Constraints and boundaries

**In scope:**
- Stage default model per pipeline_family
- Escalation matrix: (current_model, diagnosis) → next_model
- The `pick_model` SQL function
- Integration with the Phase A retry path
- Initial default chain for the OpenCode Go family

**Out of scope (explicitly):**
- Cost-aware escalation (escalation chooses on capability; cost is enforced by D-A4 cost cap as a hard limit, not a soft optimizer)
- Auto-discovery of new models (rows added manually)
- Model-vs-model A/B testing or shadow dispatching
- Cross-provider escalation (e.g., escalate from OpenCode Go to Anthropic Zen)
- Dynamic chain mutation based on observed performance (that's a Phase E refinement)

## IV. Prior art

### Brain v3's `pickModel` (steward.go)

```go
// Simplified from internal/steward/steward.go
func pickModel(stage string, attempt int, diagnosis FailureType) string {
    chain := config.EscalationChain[stage]  // e.g., ["sonnet", "opus"]
    switch diagnosis {
    case FailureModelLimit:
        // Always escalate
        return chain[min(attempt, len(chain)-1)]
    case FailureTimeout, FailureToolError:
        // Escalate after 2nd attempt
        if attempt >= 2 {
            return chain[min(attempt-1, len(chain)-1)]
        }
        return chain[0]
    default:
        return chain[0]
    }
}
```

Stage defaults from `internal/config/models.go: StageDefaults`:
- execute: sonnet (via `defaultModelForStage`)
- plan: opus
- research: sonnet
- verify: haiku-pinned
- spec: sonnet
- revise: sonnet

The pattern: each stage has a default model + an escalation chain. `model_limit` always escalates; `timeout`/`tool_error` escalate after retry #2; `transient` retries on the same model.

### Substrate's existing model selection

Today: `agent_families.default_model` is set per agent family. Single column. No escalation. No per-stage variance within a family. The pipeline runs whatever the agent_family says.

This sub-spec adds the discipline brain has, in substrate-native (SQL-driven) shape.

## V. Proposed approach

### V.1 Schema additions

```sql
-- Stage-default model per pipeline_family. One row per (pipeline_family, stage).
-- The "stage" here is the substrate's stage_name (current_stage on work_items),
-- not brain v3's commission stage names. Map carefully when seeding.
CREATE TABLE stewards.stage_models (
  pipeline_family   text  not null,
  stage_name        text  not null,
  default_model     text  not null references stewards.model_pricing(model)
                                    on update cascade,
  -- ^ FK on model only (not provider+model) for now since we have one model
  -- per name in OpenCode Go. Tighten to (provider, model) if naming collides
  -- across providers.
  notes             text,
  primary key (pipeline_family, stage_name)
);

-- Escalation matrix. Given (current_model, diagnosis), what's the next model?
-- NULL next_model means "stay on current model" (e.g., transient retries).
CREATE TABLE stewards.model_escalation (
  current_model     text  not null,
  diagnosis         text  not null check (diagnosis IN
    ('transient','timeout','model_limit','tool_error','unknown')),
  attempt_threshold int   not null default 1,  -- escalate when attempt >= this
  next_model        text,                       -- NULL = stay on current
  notes             text,
  primary key (current_model, diagnosis)
);

-- The pick_model audit lives in steward_actions (Phase A spec), no separate table.
```

**Why two tables instead of one big matrix:** stage defaults change rarely; escalation rules change rarely too, but at a different cadence. Separating them makes "I want to switch the study pipeline's research stage to GLM-5.1" a 1-row UPDATE, independent of escalation logic.

**FK to model_pricing:** ensures every model named in `stage_models` or `model_escalation` has pricing. Prevents "model exists in chain but no cost can be computed" bugs.

### V.2 The `pick_model` function

```sql
CREATE FUNCTION stewards.pick_model(
  p_pipeline_family text,
  p_stage_name      text,
  p_attempt         int,    -- 1 for first try, 2+ for retries
  p_diagnosis       text    -- one of the 5 failure types, or 'initial' for attempt=1
) RETURNS text  -- model name
LANGUAGE plpgsql
AS $$
DECLARE
  v_current_model text;
  v_escalation    record;
BEGIN
  -- Get the stage default
  SELECT default_model INTO v_current_model
  FROM stewards.stage_models
  WHERE pipeline_family = p_pipeline_family
    AND stage_name = p_stage_name;

  IF v_current_model IS NULL THEN
    RAISE EXCEPTION 'no stage_models row for %/%', p_pipeline_family, p_stage_name;
  END IF;

  -- First attempt: just return the default
  IF p_attempt <= 1 OR p_diagnosis = 'initial' THEN
    RETURN v_current_model;
  END IF;

  -- Retry: walk the escalation chain. We may need to escalate multiple
  -- times if the previous attempt's escalation has a further escalation
  -- defined for the same diagnosis.
  FOR i IN 1..p_attempt LOOP
    SELECT * INTO v_escalation
    FROM stewards.model_escalation
    WHERE current_model = v_current_model
      AND diagnosis = p_diagnosis
      AND attempt_threshold <= i;

    IF v_escalation IS NULL OR v_escalation.next_model IS NULL THEN
      -- No further escalation; return current
      RETURN v_current_model;
    END IF;

    v_current_model := v_escalation.next_model;
  END LOOP;

  RETURN v_current_model;
END;
$$;
```

**Key behaviors:**
- Idempotent: same inputs → same model, every time
- Walkable: escalation can chain (Qwen → MiniMax → Kimi → GLM) by walking N steps for N attempts
- Defensive: missing stage_models row raises (loud failure, not silent default)
- Diagnosis-aware: `transient` retries get `next_model=NULL` so they stay on the same model

### V.3 Initial escalation matrix (data)

Based on D-B1 model classifications. Subject to revision after first month of telemetry.

```sql
INSERT INTO stewards.model_escalation
  (current_model,  diagnosis,    attempt_threshold, next_model,    notes) VALUES

  -- Qwen3.6 Plus (cheapest tier) escalation
  ('qwen3.6-plus', 'model_limit',  2, 'minimax-m2.7', 'always escalate up'),
  ('qwen3.6-plus', 'timeout',      3, 'minimax-m2.7', 'escalate after 2 timeouts'),
  ('qwen3.6-plus', 'tool_error',   3, 'minimax-m2.7', 'escalate after 2 tool errors'),
  ('qwen3.6-plus', 'transient',    99, NULL,           'stay; transient is provider issue'),
  ('qwen3.6-plus', 'unknown',      3, 'minimax-m2.7', 'escalate after 2 unknowns'),

  -- MiniMax M2.7 (mid tier)
  ('minimax-m2.7', 'model_limit',  2, 'kimi-k2.6',    'escalate to general-purpose'),
  ('minimax-m2.7', 'timeout',      3, 'kimi-k2.6',    ''),
  ('minimax-m2.7', 'tool_error',   3, 'kimi-k2.6',    ''),
  ('minimax-m2.7', 'transient',    99, NULL,           ''),
  ('minimax-m2.7', 'unknown',      3, 'kimi-k2.6',    ''),

  -- Kimi K2.6 (general purpose, often Opus-leaning)
  ('kimi-k2.6',    'model_limit',  2, 'glm-5.1',      'escalate to heaviest tier'),
  ('kimi-k2.6',    'timeout',      3, 'glm-5.1',      ''),
  ('kimi-k2.6',    'tool_error',   3, 'glm-5.1',      ''),
  ('kimi-k2.6',    'transient',    99, NULL,           ''),
  ('kimi-k2.6',    'unknown',      3, 'glm-5.1',      ''),

  -- GLM-5.1 (top of OpenCode Go chain) — escalates to human-mediated queue
  -- (NOT direct dispatch to Anthropic Opus; see V.7 for the queue mechanism).
  ('glm-5.1',      'model_limit',  2, '__queue_for_opus__', 'top of auto chain; queue for human-mediated Opus boost'),
  ('glm-5.1',      'timeout',      3, '__queue_for_opus__', ''),
  ('glm-5.1',      'tool_error',   3, '__queue_for_opus__', ''),
  ('glm-5.1',      'transient',    99, NULL,                 'transient stays on GLM'),
  ('glm-5.1',      'unknown',      3, '__queue_for_opus__', '');
```

**Sentinel `__queue_for_opus__`:** not a real model name. When `pick_model` returns this string, the steward tick interprets it as "shift this work_item to escalation_queued state" instead of dispatching. Underscore prefix avoids collision with any real model name. See V.7 below for the queue mechanism.

**Rationale per row:**
- `transient` always returns NULL (stay on current model). Brain v3's pattern: transient = provider 429/500, retry on the same model. The provider issue resolves; the model is fine.
- `model_limit` escalates after 1 retry (`attempt_threshold=2`). The fastest path to a more capable model — model_limit is the diagnosis that says "this model can't handle this," so escalation is the correct response immediately.
- `timeout`, `tool_error`, `unknown` escalate after 2 retries (`attempt_threshold=3`). Per brain v3, these benefit from one same-model retry-with-feedback before escalating; if feedback didn't help, the model probably can't either.
- GLM-5.1 with all `next_model=NULL` and `attempt_threshold=99` means: at the top of the chain, never escalate; the steward will hit its `failure_count >= 3` quarantine threshold instead.

### V.4 Initial stage defaults per pipeline_family

```sql
-- Initial seed. Tune based on first-month telemetry.
INSERT INTO stewards.stage_models
  (pipeline_family, stage_name,        default_model,   notes) VALUES

  -- 'study' pipeline (study agent → research/draft/verify maturity stages)
  ('study',         'research',        'kimi-k2.6',     'general-purpose default'),
  ('study',         'outline',         'kimi-k2.6',     ''),
  ('study',         'draft',           'kimi-k2.6',     ''),
  ('study',         'verify',          'qwen3.6-plus',  'cheap binary verification'),

  -- 'lesson' pipeline
  ('lesson',        'research',        'kimi-k2.6',     ''),
  ('lesson',        'outline',         'kimi-k2.6',     ''),
  ('lesson',        'draft',           'kimi-k2.6',     ''),
  ('lesson',        'verify',          'qwen3.6-plus',  ''),

  -- 'dev' pipeline
  ('dev',           'plan',            'glm-5.1',       'design needs heaviest tier'),
  ('dev',           'execute',         'kimi-k2.6',     'general-purpose for code'),
  ('dev',           'verify',          'minimax-m2.7',  'mid-tier for code review'),

  -- Gate evaluation (used by Phase B; defined here for completeness)
  ('_gate',         'evaluate_gate',   'qwen3.6-plus',  'cheap binary gate decision'),
  ('_gate',         'generate_scenarios', 'kimi-k2.6',  'needs creativity'),
  ('_gate',         'verify_scenarios',  'qwen3.6-plus','cheap pass/fail check');
```

The `_gate` "pipeline_family" is a sentinel for the Phase B gate dispatcher to query `pick_model` for gate-related calls. Underscore prefix to make it visually distinct from real pipelines.

### V.5 Integration with Phase A steward loop

The Phase A `steward_tick` calls `pick_model` before each retry dispatch:

```sql
-- In steward_tick pseudocode (Phase A):
SELECT stewards.pick_model(
  work_item.pipeline_family,
  work_item.current_stage,
  work_item.failure_count + 1,
  work_item.last_failure_diagnosis
) INTO v_next_model;

-- Then dispatch with v_next_model
INSERT INTO steward_actions (work_item_id, observation, diagnosis, action,
                             model_used, ...) VALUES (
  work_item.id,
  'attempt #' || (work_item.failure_count + 1) || ' after ' ||
    work_item.last_failure_diagnosis,
  work_item.last_failure_diagnosis,
  'retry_with_escalation',
  v_next_model,
  ...
);
```

The audit trail makes it clear which model was chosen for which attempt — investigatable from SQL or Stewards-UI.

### V.6 Override mechanism

For research, debugging, or Michael's per-task judgment, `work_items.model_override` (a new nullable column added in Phase A) takes precedence over `pick_model`:

```sql
ALTER TABLE stewards.work_items ADD COLUMN model_override text;
```

Steward tick:
```sql
v_next_model := COALESCE(work_item.model_override, stewards.pick_model(...));
```

This lets Michael (or future trust-tier-master agents) pin a specific work_item to a specific model without changing the global escalation rules. NewWork form gets an optional "force model" picker.

### V.7 Human-mediated escalation queue (per D-EC3)

Per Michael's D-EC3 ratification, the OpenCode Go chain ends at GLM-5.1 — but instead of immediately quarantining when GLM fails, the work_item enters an **escalation queue** for human-mediated Opus boost. Two consumer paths: (a) Stewards-UI button that dispatches via OpenCode Zen's Anthropic Opus, (b) Claude Code CLI that pulls items via MCP and processes them with its own Opus subscription.

After the boost completes successfully, the work_item resumes the normal chain (Kimi K2.6 default for next stage) — escalation is a one-shot boost for the failed stage, not a permanent tier upgrade.

**State machine additions to work_items:**

```sql
ALTER TABLE stewards.work_items
  ADD COLUMN escalation_state text not null default 'normal'
    check (escalation_state IN ('normal','queued','in_progress','failed','resolved')),
  ADD COLUMN escalation_claimed_by text,    -- 'ui:zen-opus' | 'cli:claude-code-pro' | NULL
  ADD COLUMN escalation_claimed_at timestamptz,
  ADD COLUMN escalation_completed_at timestamptz,
  ADD COLUMN escalation_attempts int default 0;  -- how many times we've queued
```

**State transitions:**

```
normal --(pick_model returns __queue_for_opus__)--> queued
queued --(UI button OR CLI claim)--> in_progress
in_progress --(boost succeeds, stage advances)--> resolved (then normal for next stage)
in_progress --(boost fails)--> failed (quarantine)
queued --(timeout, e.g., 24h unclaimed)--> queued (escalation_attempts++)
                                             OR --> failed (after N escalation_attempts)
```

**Steward tick handling of the sentinel:**

```sql
-- In steward_tick pseudocode (Phase A):
SELECT stewards.pick_model(...) INTO v_next_model;

IF v_next_model = '__queue_for_opus__' THEN
  UPDATE work_items SET
    escalation_state = 'queued',
    escalation_attempts = escalation_attempts + 1,
    -- Important: do NOT increment failure_count here.
    -- The OpenCode chain genuinely exhausted; that's not a "failure" of the
    -- work_item, it's a request for elevated authority. failure_count should
    -- only count actual model failures within the chain.
    last_failure_reason = 'opencode chain exhausted; queued for Opus boost'
  WHERE id = work_item.id;

  INSERT INTO steward_actions (work_item_id, observation, diagnosis, action,
                               model_used, ...) VALUES (
    work_item.id, 'GLM-5.1 exhausted; queued for human-mediated Opus boost',
    work_item.last_failure_diagnosis, 'queue_for_opus', '__queue_for_opus__', ...);
  CONTINUE;  -- next work_item
END IF;

-- Also handle work_items already in queued state — do nothing, wait for human.
-- in_progress state handled by the boost dispatcher (separate code path).
```

### V.8 The two consumer paths

**Path A — Stewards-UI button "Boost via Zen Opus"**

UI button on WorkItemDetail (visible when escalation_state='queued'):

```sql
-- Atomic claim by UI:
UPDATE work_items SET
  escalation_state = 'in_progress',
  escalation_claimed_by = 'ui:zen-opus',
  escalation_claimed_at = now(),
  model_override = 'claude-opus-4-7'  -- one-shot override
WHERE id = $1 AND escalation_state = 'queued'
RETURNING *;
```

Then UI dispatches the failed stage normally — the substrate's existing dispatch path picks up `model_override='claude-opus-4-7'` and calls OpenCode Zen's Opus API. After the dispatch resolves:

- **Success:** UI handler clears `model_override`, sets `escalation_state='resolved'`, `escalation_completed_at=now()`. work_item.current_stage advances per normal pipeline. Subsequent stages use stage_models defaults (Kimi K2.6 etc.) — escalation does NOT persist.
- **Failure:** UI handler sets `escalation_state='failed'`, increments failure_count, runs normal quarantine path.

**Path B — Claude Code CLI via MCP**

Claude Code (Michael's local CLI with Anthropic Pro subscription) periodically polls or is manually invoked to drain the queue. Uses two new MCP tools on the substrate's MCP server:

```
mcp__pg-ai-stewards__work_item_escalation_list
  -- Returns array of work_items WHERE escalation_state='queued'
  -- Each entry includes: id, slug, pipeline_family, current_stage,
  --   last_failure_reason, last_failure_diagnosis, scratch path,
  --   composed system prompt for the failed stage.

mcp__pg-ai-stewards__work_item_escalation_claim(work_item_id)
  -- Atomic UPDATE: escalation_state='in_progress',
  --   escalation_claimed_by='cli:claude-code-pro',
  --   escalation_claimed_at=now()
  -- Returns the full work_item context (system prompt, scratch, intent, etc.)
  -- so Claude Code has everything it needs to process locally.

mcp__pg-ai-stewards__work_item_escalation_resolve(work_item_id, success, output)
  -- success=true:
  --   Append output to work_items.stage_results->current_stage,
  --   set escalation_state='resolved',
  --   advance current_stage per pipeline,
  --   clear escalation_claimed_*.
  -- success=false:
  --   set escalation_state='failed', quarantine.
```

In this path, **Claude Code itself is the inference engine** — no provider call from the substrate. Michael's Pro subscription is consumed; OpenCode Zen is not. This is the cheapest path for Opus-boost work since it uses the Pro subscription's bucket allowance instead of pay-per-token billing.

**Choosing between Path A and Path B:**
- Path A (UI button): immediate, you click and it goes. Costs $5-25 per boost in OpenCode Zen rates.
- Path B (CLI claim): asynchronous, free if within Pro subscription bucket, but requires Michael to actually run `claude` and do the work.

Default UX: surface BOTH options on the WorkItemDetail page when state is queued. Michael picks per item.

**Idempotency note:** the claim functions use `WHERE escalation_state='queued'` in the UPDATE, so if two consumers try to claim simultaneously, only one succeeds (the other's UPDATE affects 0 rows and returns no work_item). Standard SQL-row-locking behavior.

## VI. Phased delivery

### Phase A.escalation.1 — Schema + function (1 session, depends on Phase A.cost.1)
- Migration: 2 tables + work_items.model_override column
- SQL function: `pick_model`
- Seed data: stage_models for study/lesson/dev/_gate; full model_escalation matrix
- pgTAP tests: every (current_model, diagnosis) combination returns correct next_model; transient stays on same; top-of-chain returns same; missing stage_models raises

### Phase A.escalation.2 — Steward tick integration (within Phase A.steward main work)
- Wire steward tick to call `pick_model` before each retry dispatch
- Wire dispatch to honor `model_override`
- Record `model_used` in steward_actions

### Phase A.escalation.3 — UI surfaces (within Phase A UI work)
- WorkItemDetail shows model used per attempt + escalation path
- NewWork form gets optional "force model" picker (model_override)
- Stewards-UI Watchman surfaces work_items at top-of-chain that quarantined (escalation exhausted)

### Phase A.escalation.4 — Human-mediated escalation queue (1-2 sessions)
- Migration: 5 columns on work_items (escalation_state, escalation_claimed_by, escalation_claimed_at, escalation_completed_at, escalation_attempts) plus a CHECK constraint
- Steward tick handler for `__queue_for_opus__` sentinel (write to escalation_state instead of incrementing failure_count)
- 3 new MCP tools on stewards-mcp: `work_item_escalation_list`, `work_item_escalation_claim`, `work_item_escalation_resolve`
- Stewards-UI: "Boost via Zen Opus" button on WorkItemDetail when state='queued', shows escalation_state badge in WorkItem cards
- Watchman section "Escalation Queue" listing all queued items with click-through
- pgTAP tests: state transitions, atomic claim races, sentinel-to-state-transition
- E2E test: a synthetic failure that exhausts the chain ends in queued state, UI button claims it, dispatches with override, resolves successfully on success path

## VII. Verification criteria

**A.escalation.1 verification:**
- `pick_model('study', 'research', 1, 'initial')` returns `kimi-k2.6`
- `pick_model('study', 'research', 2, 'model_limit')` returns `glm-5.1` (Kimi escalates immediately on model_limit)
- `pick_model('study', 'research', 2, 'transient')` returns `kimi-k2.6` (transient stays)
- `pick_model('study', 'research', 3, 'timeout')` returns `glm-5.1` (Kimi escalates on attempt 3 after timeout)
- `pick_model('dev', 'plan', 1, 'initial')` returns `glm-5.1` (dev/plan starts at top)
- `pick_model('dev', 'plan', 5, 'model_limit')` returns `glm-5.1` (already at top, no escalation)
- `pick_model('nonexistent', 'whatever', 1, 'initial')` raises 'no stage_models row'

**A.escalation.2 verification:**
- A failed work_item with diagnosis=model_limit dispatches its retry on the escalated model
- Setting `model_override = 'qwen3.6-plus'` on a work_item makes the retry dispatch on Qwen regardless of stage default
- steward_actions row shows `model_used = qwen3.6-plus` and notes the override

**A.escalation.3 verification:**
- WorkItemDetail shows: Attempt 1 (kimi-k2.6) → Attempt 2 (glm-5.1, escalated due to model_limit)
- NewWork form's "force model" picker is populated from `model_pricing`
- A work_item that escalated to GLM-5.1 and failed appears in Watchman as `escalation_exhausted`

**Inverse hypothesis verification:**
- Reproduce a model_limit failure on Kimi → confirm escalation to GLM → remove the escalation row → reproduce → confirm steward stays on Kimi (no escalation defined) → restore.

## VIII. Costs and risks

**Build cost:** ~2 sessions for schema + function + seed + integration. UI is part of broader Phase A UI work.

**Runtime cost:** `pick_model` is a SELECT + small loop, called once per dispatch. Sub-millisecond.

**Risks:**
1. **Wrong tier classification.** If Kimi K2.6 turns out to be more capable than GLM-5.1 in practice, escalating Kimi → GLM is going the wrong way. Mitigation: D-E (Phase E) trust scoring will surface this; we can rewrite the matrix from data without touching code.
2. **Escalation loop.** If model_escalation has a cycle (A → B → A), `pick_model`'s loop may not converge cleanly. Mitigation: the `attempt` parameter caps the loop iterations; cycles still resolve to *some* model in finite time. Add a CHECK constraint preventing direct A → A self-loops.
3. **Stage names drift.** If pipeline_family + stage_name combinations drift over time (a stage is renamed), `pick_model` will raise. Mitigation: stage_models migrations should accompany pipeline definition changes. Documented in Phase A migration playbook.
4. **OpenCode Go API contract changes.** Provider changes their model names or capability profile. Mitigation: data-driven design means a model rename is an UPDATE, not a code change.

## IX. Open questions for Michael — RESOLVED 2026-05-10

1. **Tier confirmation** — RESOLVED: confirmed Qwen → MiniMax → Kimi → GLM as written. Single global chain across all stages for now; revisit per-stage chains if telemetry shows divergence.
2. **Default attempt thresholds** — RESOLVED: keep brain's defaults (model_limit=2, others=3). Encoded in V.3 matrix.
3. **Cross-family escalation** — RESOLVED with significant design addition: instead of auto-escalating to Anthropic, **add the human-mediated escalation queue (V.7-V.8)**. GLM-5.1 escalates to a `__queue_for_opus__` sentinel which transitions the work_item to `escalation_state='queued'`. Two consumer paths: Stewards-UI button (dispatches Opus via OpenCode Zen, costs $5-25 per boost) and Claude Code CLI via new MCP tools (uses Michael's Pro subscription, free within bucket). After boost completes, work_item resumes normal Kimi K2.6 chain.
4. **Gate model defaults** — RESOLVED: confirmed (evaluate_gate=Qwen3.6 Plus, generate_scenarios=Kimi K2.6, verify_scenarios=Qwen3.6 Plus).

## X. Acceptance scenarios (decision-ready handoff)

After A.escalation ships:

1. **Initial dispatch:** Create a `study` work_item. Steward dispatches first attempt on `kimi-k2.6`. cost_events row written with model=kimi-k2.6.
2. **Model_limit escalation:** Synthetic failure with diagnosis=model_limit. Steward retry dispatches on `glm-5.1`. steward_actions shows the escalation; cost_events shows the new model.
3. **Transient retry:** Synthetic failure with diagnosis=transient. Steward retry dispatches on `kimi-k2.6` again (no escalation). Cost row shows same model.
4. **Top-of-chain failure:** Force a work_item to GLM-5.1 (via model_override). Synthetic model_limit failure. Steward sees no escalation possible, increments failure_count. After 3 failures, quarantine fires with reason `escalation_exhausted` (or just `failure_count_limit` — confirm naming with Phase A spec).
5. **Override:** Create work_item with model_override='qwen3.6-plus'. Steward dispatches on Qwen regardless of stage default. Failure escalates per Qwen's escalation rules (to MiniMax). The override only affects the *first* model; subsequent retries follow the chain.

6. **Escalation queue trigger:** Create a `study` work_item. Synthetic chain of failures: model_limit on Kimi K2.6 → escalates to GLM-5.1 → another model_limit. Steward sees `pick_model` returns `__queue_for_opus__`. Confirm: work_item.escalation_state = 'queued', no failure_count increment, steward_actions row with action='queue_for_opus'.

7. **UI boost path:** From the queued work_item in scenario 6, Michael clicks "Boost via Zen Opus" in Stewards-UI. Confirm: escalation_state transitions to 'in_progress', escalation_claimed_by='ui:zen-opus', model_override='claude-opus-4-7'. Dispatch happens via Opus. On success: escalation_state='resolved', model_override cleared, current_stage advances, next stage dispatched on stage_models default (Kimi K2.6). cost_events shows the Opus dispatch with high micro_dollars.

8. **CLI boost path:** From the same queued state (scenario 6, alternative path), Michael runs Claude Code CLI which calls `mcp__pg-ai-stewards__work_item_escalation_list` (returns the queued item), then `mcp__pg-ai-stewards__work_item_escalation_claim(id)` (returns full context), Claude Code processes with its own Opus, calls `mcp__pg-ai-stewards__work_item_escalation_resolve(id, true, output_text)`. Confirm: escalation_state='resolved', escalation_claimed_by='cli:claude-code-pro', stage_results populated, current_stage advanced, NO cost_events row from substrate (because no substrate dispatch happened).

9. **Atomic claim race:** Two consumers try to claim the same queued work_item simultaneously (UI button + CLI claim). Confirm: only one succeeds, the other gets a "no rows updated" response and the UI shows the item is already in_progress.

10. **Boost failure:** Queued work_item gets boosted via UI Path A; the Opus dispatch fails (e.g., timeout). Confirm: escalation_state='failed', failure_count incremented, work_item quarantined per normal failure path. steward_actions records 'opus_boost_failed'.

If scenarios 1-5 pass, the basic escalation chain is complete. If 6-10 pass, the human-mediated escalation queue is complete and Phase A.escalation is fully shippable.
