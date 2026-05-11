-- =====================================================================
-- Phase 5f (Phase E.1) — Trust ladder schema
--
-- Per ratifications:
--   D-E1: 3-tier trainee / journeyman / master
--   D-E-trust-keying: (agent_family, pipeline_family, model)
--                    — model dimension recognizes that
--                    "kimi-k2.6 doing study-write outline" is
--                    different from "qwen3.6-plus doing the same"
--   D-E2: Manual trust adjustments require justification
--   D-E3: Human override counts as failure (full weight)
--
-- Four tables:
--   stewards.trust_scores        — current per-(agent,pipeline,model) state
--   stewards.trust_transitions   — append-only audit of level changes
--   stewards.gate_overrides      — append-only record of human disagreements
--                                 with gate decisions
--   stewards.trust_thresholds    — config: 5 to journeyman, 15 more to master
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) trust_scores
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.trust_scores (
    agent_family            text NOT NULL,
    pipeline_family         text NOT NULL,
    model                   text NOT NULL,
    successful_completions  int NOT NULL DEFAULT 0,
    failed_completions      int NOT NULL DEFAULT 0,
    human_overrides         int NOT NULL DEFAULT 0,
    trust_level             text NOT NULL DEFAULT 'trainee'
                              CHECK (trust_level IN ('trainee', 'journeyman', 'master')),
    last_evaluated_at       timestamptz NOT NULL DEFAULT now(),
    last_completion_at      timestamptz,
    PRIMARY KEY (agent_family, pipeline_family, model)
);

COMMENT ON TABLE stewards.trust_scores IS
'Phase 5f (E.1): per-(agent_family, pipeline_family, model) trust state. Trainee surfaces every gate-advance for human ratification; journeyman + master proceed automatically. Demote on human override (D-E3 full weight).';

-- ---------------------------------------------------------------------
-- (2) trust_transitions — audit ledger
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.trust_transitions (
    id                  bigserial PRIMARY KEY,
    at                  timestamptz NOT NULL DEFAULT now(),
    agent_family        text NOT NULL,
    pipeline_family     text NOT NULL,
    model               text NOT NULL,
    from_level          text NOT NULL,
    to_level            text NOT NULL,
    transition_kind     text NOT NULL CHECK (transition_kind IN ('auto', 'manual')),
    actor               text NOT NULL,
    justification       text,         -- required for manual (D-E2); enforced in adjust fn
    metrics             jsonb         -- snapshot at transition: {successful, failed, overrides}
);

CREATE INDEX IF NOT EXISTS trust_transitions_at   ON stewards.trust_transitions (at);
CREATE INDEX IF NOT EXISTS trust_transitions_cell ON stewards.trust_transitions (agent_family, pipeline_family, model);

COMMENT ON TABLE stewards.trust_transitions IS
'Phase 5f (E.1): every level change recorded with reason. Manual transitions require justification (D-E2).';

-- ---------------------------------------------------------------------
-- (3) gate_overrides — human disagreement with a gate decision
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.gate_overrides (
    id                bigserial PRIMARY KEY,
    gate_decision_id  bigint NOT NULL REFERENCES stewards.gate_decisions(id),
    at                timestamptz NOT NULL DEFAULT now(),
    overridden_by     text NOT NULL,
    new_action        text NOT NULL CHECK (new_action IN ('advance', 'revise', 'surface')),
    justification     text NOT NULL
);

CREATE INDEX IF NOT EXISTS gate_overrides_decision ON stewards.gate_overrides (gate_decision_id);
CREATE INDEX IF NOT EXISTS gate_overrides_at       ON stewards.gate_overrides (at);

COMMENT ON TABLE stewards.gate_overrides IS
'Phase 5f (E.1): records when a human disagreed with a gate decision. Increments human_overrides on the relevant trust_scores row (D-E3 full weight).';

-- ---------------------------------------------------------------------
-- (4) trust_thresholds — tunable promotion rules
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.trust_thresholds (
    transition          text PRIMARY KEY,
    required_successes  int NOT NULL,
    clean_window        int NOT NULL,
    demote_on_override  boolean NOT NULL DEFAULT true
);

INSERT INTO stewards.trust_thresholds (transition, required_successes, clean_window, demote_on_override) VALUES
    ('trainee_to_journeyman', 5, 5, true),
    ('journeyman_to_master', 15, 15, true)
ON CONFLICT (transition) DO UPDATE SET
    required_successes = EXCLUDED.required_successes,
    clean_window       = EXCLUDED.clean_window,
    demote_on_override = EXCLUDED.demote_on_override;

COMMENT ON TABLE stewards.trust_thresholds IS
'Phase 5f (E.1): tunable promotion rules. Default: trainee -> journeyman after 5 successes with no overrides; journeyman -> master after 15 more clean. Demote one level on any override.';
