---
title: Phase E — Trust ladder + Line upon Line
date: 2026-05-11
status: design sub-spec — ready for Phase E implementation
parent: full-agentic-substrate.md (D-E1..E3 ratification, plus 2026-05-11 re-validation)
purpose: >
  Make stewardship a progressive-trust mechanism. Today every gate
  decision surfaces with equal weight regardless of whether the agent
  has demonstrated reliability on this work. Phase E adds trust scoring
  per (agent_family, pipeline_family, model) — agents earn authority,
  not get granted it. Also wires Phase D's ratified lessons into the
  retry composer so the substrate teaches itself.
---

# Phase E — Trust ladder + Line upon Line

## I. Binding problem

The substrate has two adjacent gaps after Phases A–D:

1. **Trust is binary.** Today the gate's `advance` decision either gets the human's review or it doesn't, and the choice is implicit (gated by whether the human has the time to look). There's no notion of "this agent has earned the right to advance without me reviewing every call." That's not autonomy-for-its-own-sake — it's the gospel pattern of stewardship: authority earned by demonstrated faithfulness (D&C 82:3 — "where much is given, much is required" inverts to "where little has been demonstrated, little is given").

2. **Lessons are recorded but unused at runtime.** Phase D's `stewards.lessons` table accumulates ratified lessons but nothing reads them at dispatch time. The retry composer (Phase A's `retry_guidance_text`) is failure-type-keyed — a tool_error fail gets the same guidance regardless of what specific tool errors this pipeline has hit before. Line-upon-line means the substrate's experience compounds within a pipeline, not just within a work_item.

Phase E closes both gaps.

## II. Success criteria

1. **Trust scores exist per (agent_family, pipeline_family, model).** The substrate distinguishes "kimi-k2.6 doing study-write outline" from "qwen3.6-plus doing study-write outline."
2. **Trust level affects gate behavior.** Trainee: every advance surfaces. Journeyman: advance proceeds; revise/surface still surface. Master: advance proceeds AND agent can propose new pipelines for human review.
3. **Promotion criteria are explicit.** Trainee → journeyman = 5 verified completions with no human overrides on this (agent, pipeline, model) triple. Journeyman → master = 15 more with no overrides. Manual override allowed (D-E2 — with justification).
4. **Human override counts as failure.** When a human overrides a gate decision (e.g. gate said advance, human said revise), it counts as a failure for trust scoring. Tracked with full weight (D-E3).
5. **Retry composer pulls lessons.** When the steward retries a failed work_item, the retry context (Phase A `retry_guidance` template) gains a "Recent ratified lessons for this pipeline + stage" section listing the last 3.
6. **Trust matrix visible.** Stewards-UI Watchman page (or new `/trust` route) shows the trust matrix: rows = agent_family, columns = pipeline_family, cells = trust level + completion count + override count. Manual adjust UI with required justification field.

## III. Constraints and boundaries

**In scope:**
- `stewards.trust_scores` table keyed on `(agent_family, pipeline_family, model)`
- `stewards.trust_transitions` audit ledger for level changes (auto + manual)
- `stewards.evaluate_trust(agent_family, pipeline_family, model)` SQL function — recomputes trust level from completions + overrides
- Auto-promotion trigger or scheduled tick (5-minute cadence) that runs evaluate_trust on touched cells
- `stewards.gate_should_surface(work_item_id, action)` — the gate-side helper that consults trust to decide whether to surface advance decisions
- `apply_gate_decision` revision: for action='advance', if `gate_should_surface` returns true, transition status='awaiting_review' instead of advancing
- Human-override hook: a Stewards-UI button "I disagree with this gate decision" that records the override and tanks trust
- Retry composer extension to pull `lessons_recent_ratified`
- Stewards-UI trust matrix view + manual adjustment flow

**Out of scope:**
- Trust portability across pipelines (master at study-write doesn't grant any trust at lesson-write)
- Trust decay over time (a master that hasn't been used in 6 months stays master until evidence says otherwise)
- Trust prediction or recommendation ("this new agent should start at journeyman" — no, everyone starts trainee)
- Cross-model trust transfer (kimi-k2.6 master at study-write outline doesn't grant qwen3.6-plus any boost; verified 2026-05-11 keying decision)

## IV. Prior art

- **Phase 4a `steward_actions`** — append-only audit ledger. trust_transitions mirrors this shape.
- **Phase 5a `gate_decisions`** — already records action + reasoning. Phase E adds `human_override` column or a separate `gate_overrides` table to record the human's disagreement.
- **Phase 4a `pipeline_breakers`** — circuit breaker per pipeline_family. Trust scoring is the per-(agent, pipeline, model) analogue: aggregate over time, drives behavior change.
- **Phase D `stewards.lessons` view `lessons_recent_ratified`** — already exists per Phase D sub-spec V.1. Phase E's retry composer reads from this view.
- **Brain v3** — has none of this. Phase E is net-new substrate ground (proposal explicit).

## V. Proposed approach

### V.1 Schema

```sql
CREATE TABLE stewards.trust_scores (
    agent_family            text NOT NULL,
    pipeline_family         text NOT NULL,
    model                   text NOT NULL,
    successful_completions  int NOT NULL DEFAULT 0,    -- maturity reached verified, no override
    failed_completions      int NOT NULL DEFAULT 0,    -- quarantined or terminal-failed
    human_overrides         int NOT NULL DEFAULT 0,    -- gate said X, human said not-X
    trust_level             text NOT NULL DEFAULT 'trainee'
                              CHECK (trust_level IN ('trainee', 'journeyman', 'master')),
    last_evaluated_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (agent_family, pipeline_family, model)
);

CREATE TABLE stewards.trust_transitions (
    id                  bigserial PRIMARY KEY,
    at                  timestamptz NOT NULL DEFAULT now(),
    agent_family        text NOT NULL,
    pipeline_family     text NOT NULL,
    model               text NOT NULL,
    from_level          text NOT NULL,
    to_level            text NOT NULL,
    transition_kind     text NOT NULL CHECK (transition_kind IN ('auto', 'manual')),
    actor               text NOT NULL,            -- 'system' for auto, human name for manual
    justification       text,                     -- required for manual (D-E2)
    metrics             jsonb                     -- snapshot at transition: {successful, failed, overrides}
);

CREATE INDEX trust_transitions_at ON stewards.trust_transitions (at);
CREATE INDEX trust_transitions_cell ON stewards.trust_transitions (agent_family, pipeline_family, model);

-- Override records — a separate ledger because a single gate_decision
-- can have at most one override but having a separate table keeps
-- gate_decisions append-only and lets the override record carry
-- its own justification.
CREATE TABLE stewards.gate_overrides (
    id                bigserial PRIMARY KEY,
    gate_decision_id  bigint NOT NULL REFERENCES stewards.gate_decisions(id),
    at                timestamptz NOT NULL DEFAULT now(),
    overridden_by     text NOT NULL,
    new_action        text NOT NULL CHECK (new_action IN ('advance', 'revise', 'surface')),
    justification     text NOT NULL
);

CREATE INDEX gate_overrides_decision ON stewards.gate_overrides (gate_decision_id);
```

### V.2 Promotion thresholds

```
trainee → journeyman:
  successful_completions >= 5 AND human_overrides == 0
  (5 verified completions with no override on any of them)

journeyman → master:
  successful_completions >= 20 (cumulative; the 5 from trainee + 15 more)
  AND human_overrides == 0 in the last 15 completions
  (15-completion clean window after journeyman entry)

Demotion (auto):
  Any human_override → demote one level immediately
  trainee stays trainee (can't demote further)

Demotion (manual):
  Human can demote at any time with justification.
```

This is conservative. Tunable in `stewards.trust_thresholds` config table if 5/15 prove wrong:

```sql
CREATE TABLE stewards.trust_thresholds (
    transition          text PRIMARY KEY,         -- 'trainee_to_journeyman' | 'journeyman_to_master'
    required_successes  int NOT NULL,
    clean_window        int NOT NULL,             -- completions counted backward for override check
    demote_on_override  boolean NOT NULL DEFAULT true
);

INSERT INTO stewards.trust_thresholds VALUES
    ('trainee_to_journeyman', 5, 5, true),
    ('journeyman_to_master', 15, 15, true);
```

### V.3 evaluate_trust function

```sql
CREATE OR REPLACE FUNCTION stewards.evaluate_trust(
    p_agent_family text, p_pipeline_family text, p_model text
) RETURNS text
LANGUAGE plpgsql AS $$
DECLARE
    v_score stewards.trust_scores%ROWTYPE;
    v_new_level text;
    v_threshold_t2j stewards.trust_thresholds%ROWTYPE;
    v_threshold_j2m stewards.trust_thresholds%ROWTYPE;
BEGIN
    SELECT * INTO v_score
      FROM stewards.trust_scores
     WHERE agent_family = p_agent_family
       AND pipeline_family = p_pipeline_family
       AND model = p_model
       FOR UPDATE;

    IF NOT FOUND THEN
        INSERT INTO stewards.trust_scores (agent_family, pipeline_family, model)
        VALUES (p_agent_family, p_pipeline_family, p_model)
        RETURNING * INTO v_score;
    END IF;

    SELECT * INTO v_threshold_t2j FROM stewards.trust_thresholds WHERE transition='trainee_to_journeyman';
    SELECT * INTO v_threshold_j2m FROM stewards.trust_thresholds WHERE transition='journeyman_to_master';

    v_new_level := v_score.trust_level;

    -- Promotion
    IF v_score.trust_level = 'trainee'
       AND v_score.successful_completions >= v_threshold_t2j.required_successes
       AND v_score.human_overrides = 0 THEN
        v_new_level := 'journeyman';
    ELSIF v_score.trust_level = 'journeyman'
       AND v_score.successful_completions >= v_threshold_t2j.required_successes
                                            + v_threshold_j2m.required_successes
       AND (
         SELECT count(*) = 0 FROM stewards.gate_overrides go
           JOIN stewards.gate_decisions gd ON gd.id = go.gate_decision_id
           JOIN stewards.work_items wi ON wi.id = gd.work_item_id
          WHERE wi.pipeline_family = p_pipeline_family
            AND wi.actor = p_agent_family   -- TODO confirm actor mapping
            AND go.at >= (SELECT activated_at FROM stewards.trust_transitions
                           WHERE agent_family = p_agent_family
                             AND pipeline_family = p_pipeline_family
                             AND model = p_model
                             AND to_level = 'journeyman'
                           ORDER BY at DESC LIMIT 1)
       ) THEN
        v_new_level := 'master';
    END IF;

    IF v_new_level <> v_score.trust_level THEN
        UPDATE stewards.trust_scores
           SET trust_level = v_new_level, last_evaluated_at = now()
         WHERE agent_family = p_agent_family
           AND pipeline_family = p_pipeline_family
           AND model = p_model;

        INSERT INTO stewards.trust_transitions
            (agent_family, pipeline_family, model, from_level, to_level,
             transition_kind, actor, metrics)
        VALUES
            (p_agent_family, p_pipeline_family, p_model,
             v_score.trust_level, v_new_level, 'auto', 'system',
             jsonb_build_object(
                 'successful', v_score.successful_completions,
                 'failed', v_score.failed_completions,
                 'overrides', v_score.human_overrides
             ));
    END IF;

    RETURN v_new_level;
END;
$$;
```

Called:
- After every work_item reaches `verified` maturity (Phase D's Sabbath dispatch is a natural firing point).
- After every quarantine.
- After every override insert (trigger).
- Manually via Stewards-UI "Re-evaluate" button.

### V.4 Counter maintenance

Triggers on the relevant tables:

```sql
-- successful_completions: incremented when a work_item reaches verified.
-- The cleanest hook is in apply_gate_decision when action='advance' and
-- the new maturity is 'verified'.

-- failed_completions: incremented when status transitions to 'quarantined'.

-- human_overrides: incremented when a row is inserted into gate_overrides.

-- These can be triggers OR explicit calls in the relevant SQL functions.
-- Recommend explicit calls — easier to debug, no hidden trigger surprises.
```

### V.5 Gate behavior change

`apply_gate_decision` revision (extends Phase 5a):

```sql
-- Existing logic for advance/revise/surface ...

IF v_action = 'advance' THEN
    -- NEW: check trust before transitioning maturity
    DECLARE
        v_trust text;
    BEGIN
        SELECT trust_level INTO v_trust
          FROM stewards.trust_scores
         WHERE agent_family = v_wi.actor                  -- mapping caveat
           AND pipeline_family = v_wi.pipeline_family
           AND model = COALESCE(v_wi.model_override,
                       (SELECT default_model FROM stewards.pipeline_stages
                         WHERE pipeline_family = v_wi.pipeline_family
                           AND stage_name = v_wi.current_stage));

        IF v_trust IS NULL OR v_trust = 'trainee' THEN
            -- Trainee: surface for human ratification instead of auto-advancing
            UPDATE stewards.work_items
               SET status = 'awaiting_review',
                   updated_at = now()
             WHERE id = p_work_item_id;
            RETURN v_wi.maturity;   -- maturity unchanged; human must ratify
        END IF;
        -- Journeyman + master: advance proceeds (existing logic below)
    END;

    -- existing advance logic ...
END IF;
```

Trust at trainee level => every advance surfaces. Stewards-UI shows the gate decision + an "Approve advance" / "Override (revise)" / "Override (surface)" choice. Approve increments `successful_completions`; Override inserts a `gate_overrides` row.

### V.6 Retry composer extension

Phase A's `retry_guidance(diagnosis, attempt)` returns a templated string. Phase E adds:

```sql
CREATE OR REPLACE FUNCTION stewards.retry_guidance_with_lessons(
    p_diagnosis text, p_attempt int,
    p_pipeline_family text, p_stage_name text
) RETURNS text
LANGUAGE plpgsql AS $$
DECLARE
    v_base text;
    v_lessons_section text;
BEGIN
    v_base := stewards.retry_guidance(p_diagnosis, p_attempt);

    SELECT string_agg('  - ' || content, E'\n')
      INTO v_lessons_section
      FROM (
        SELECT content
          FROM stewards.lessons_recent_ratified
         WHERE pipeline_family = p_pipeline_family
           AND current_stage = p_stage_name
         ORDER BY at DESC
         LIMIT 3
      ) recent;

    IF v_lessons_section IS NOT NULL THEN
        v_base := v_base || E'\n\nRecent lessons from this pipeline + stage:\n' || v_lessons_section;
    END IF;

    RETURN v_base;
END;
$$;
```

The steward's retry path (Phase 4c `steward_dispatch`) calls this instead of `retry_guidance` directly.

### V.7 Override hook

Stewards-UI WorkItemDetail panel for awaiting-review work_items shows the gate decision. A "Disagree (override)" button opens a small form: choose new_action (advance | revise | surface), enter justification (required, min 10 chars). POST `/api/gate-decisions/override`.

```go
// api/gate_overrides.go
type overrideReq struct {
    GateDecisionID int64  `json:"gate_decision_id"`
    NewAction      string `json:"new_action"`
    Justification  string `json:"justification"`
}

// Inserts into gate_overrides, then re-applies the gate decision with the
// new action (calls apply_gate_decision again with the corrected action).
// Increments human_overrides on the relevant trust_scores row.
// Triggers evaluate_trust which will demote.
```

### V.8 Stewards-UI Trust Matrix

New top-level route `/trust` (or panel on Watchman page). Table view:
- Rows: distinct agent_family in trust_scores
- Columns: distinct pipeline_family
- Cells: Show trust_level badge (trainee gray, journeyman blue, master gold) + tiny `successes/overrides` ratio
- Click cell → drill-down panel with model-keyed breakdown + transition history
- "Manual adjust" button on each cell → modal with new_level + required justification

API:
- `GET /api/trust/scores` — full list (agent, pipeline, model, level, counts)
- `GET /api/trust/transitions?agent=X&pipeline=Y&model=Z` — history for a cell
- `POST /api/trust/adjust` — manual adjustment with justification

## VI. Open questions / follow-ups

- **Actor → agent_family mapping.** `work_items.actor` today carries either 'human' or a free-form string. Phase E needs a clean agent_family identifier per work_item. Either reuse `actor` with stricter validation or add `agent_family` column.
- **First-completion bootstrap.** A brand-new (agent, pipeline, model) cell starts trainee, every advance surfaces. With 5 surfaces required to escape trainee, the first 5 are slow. Is that the right friction or too much? Recommend keep at 5; if it feels heavy after 1 month of use, lower in `trust_thresholds`.
- **Override-on-revise vs override-on-advance.** Today's override schema treats all overrides equal weight. Should "I think you should have surfaced this" count differently from "I think you should have advanced this"? Recommend equal weight initially; revisit if signal is noisy.
- **Trust on Phase F council members.** Phase F has agents serving roles (proposer/critic/synthesizer/bishop). Does each role earn its own trust? Or does the agent's master-tier on the underlying pipeline grant council-eligibility? Phase F sub-spec ratified the latter (master-on-pipeline). Phase E doesn't add a council_trust dimension; that's on the F2-future-evolution path.
- **Phase E and Phase D ordering.** D ships first per the proposal's V phase ladder (D requires A, B; E requires A, D). Phase E sub-spec assumes Phase D's `lessons_recent_ratified` view exists. Build order locked.
- **Stewards-UI navigation getting busy.** Adding `/intents`, `/covenants`, `/sabbath`, `/lessons`, `/trust` plus existing `/dashboard`, `/work-items`, `/sessions`, `/watchman`, `/bridge`, `/studies`, `/graph`, `/new`. Worth a Phase E followup: sidebar grouping (Substrate / Surfaces / Records).

## VII. Estimated programming time

- V.1 schema + V.2/V.3 evaluate_trust + V.4 counter maintenance: 1 session
- V.5 gate behavior change + V.6 retry composer + V.7 override hook + smoke test: 1 session
- V.8 Stewards-UI Trust Matrix + adjustment flow: 1 session

**Total: 3 sessions** (one over the proposal's 2-session estimate; Trust Matrix UI accounts for the extra).

## VIII. Acceptance scenarios

- A new (kimi-k2.6, study-write, outline) cell shows up in trust_scores with `trust_level='trainee'` after the first dispatch.
- After 5 verified completions of study-write work_items by kimi-k2.6 with no human overrides, evaluate_trust auto-promotes to journeyman; transition row appears in `/trust` UI.
- A human clicks "Override (revise)" on a gate decision that said advance. `gate_overrides` row inserted; `human_overrides` incremented; trust auto-demotes if the cell was journeyman (back to trainee).
- A retry of a failed study-write outline includes "Recent lessons from this pipeline + stage" with the last 3 ratified lessons in the prompt context.
- A trust matrix in `/trust` shows the (kimi-k2.6, study-write) row at journeyman; manual adjust to master with justification "Empirically perfect on the last 12 outlines" succeeds and writes a `transition_kind='manual'` row.
- Phase B's gate-eval cost stays roughly constant — Phase E adds context (last 3 lessons) but lessons are short; no 5× blowout.
