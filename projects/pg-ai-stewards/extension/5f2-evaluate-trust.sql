-- =====================================================================
-- Phase 5f.2 (Phase E.2) — evaluate_trust + counter helpers
--
-- Three SQL functions:
--   stewards.trust_record_success(agent_family, pipeline_family, model)
--   stewards.trust_record_failure(agent_family, pipeline_family, model)
--   stewards.trust_record_override(agent_family, pipeline_family, model, by, justification)
--   stewards.evaluate_trust(agent_family, pipeline_family, model)
--   stewards.trust_adjust(agent_family, pipeline_family, model, new_level, actor, justification)
--
-- The record_* helpers bump counters AND call evaluate_trust to apply
-- promotion / demotion if a threshold is crossed.
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) trust_record_success — incremented when a work_item reaches verified
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trust_record_success(
    p_agent_family text, p_pipeline_family text, p_model text
) RETURNS text
LANGUAGE plpgsql AS $func$
BEGIN
    INSERT INTO stewards.trust_scores
        (agent_family, pipeline_family, model,
         successful_completions, last_completion_at)
    VALUES
        (p_agent_family, p_pipeline_family, p_model, 1, now())
    ON CONFLICT (agent_family, pipeline_family, model) DO UPDATE SET
        successful_completions = stewards.trust_scores.successful_completions + 1,
        last_completion_at     = now();

    RETURN stewards.evaluate_trust(p_agent_family, p_pipeline_family, p_model);
END;
$func$;

COMMENT ON FUNCTION stewards.trust_record_success(text, text, text) IS
'Phase 5f (E.2): increment successful_completions and re-evaluate. Called when a work_item reaches verified maturity.';

-- ---------------------------------------------------------------------
-- (2) trust_record_failure — incremented on quarantine
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trust_record_failure(
    p_agent_family text, p_pipeline_family text, p_model text
) RETURNS text
LANGUAGE plpgsql AS $func$
BEGIN
    INSERT INTO stewards.trust_scores
        (agent_family, pipeline_family, model, failed_completions)
    VALUES
        (p_agent_family, p_pipeline_family, p_model, 1)
    ON CONFLICT (agent_family, pipeline_family, model) DO UPDATE SET
        failed_completions = stewards.trust_scores.failed_completions + 1;

    RETURN stewards.evaluate_trust(p_agent_family, p_pipeline_family, p_model);
END;
$func$;

COMMENT ON FUNCTION stewards.trust_record_failure(text, text, text) IS
'Phase 5f (E.2): increment failed_completions on quarantine.';

-- ---------------------------------------------------------------------
-- (3) trust_record_override — increment human_overrides + auto-demote
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trust_record_override(
    p_agent_family text, p_pipeline_family text, p_model text
) RETURNS text
LANGUAGE plpgsql AS $func$
BEGIN
    INSERT INTO stewards.trust_scores
        (agent_family, pipeline_family, model, human_overrides)
    VALUES
        (p_agent_family, p_pipeline_family, p_model, 1)
    ON CONFLICT (agent_family, pipeline_family, model) DO UPDATE SET
        human_overrides = stewards.trust_scores.human_overrides + 1;

    RETURN stewards.evaluate_trust(p_agent_family, p_pipeline_family, p_model);
END;
$func$;

COMMENT ON FUNCTION stewards.trust_record_override(text, text, text) IS
'Phase 5f (E.2): increment human_overrides and re-evaluate. evaluate_trust auto-demotes on any override per D-E3.';

-- ---------------------------------------------------------------------
-- (4) evaluate_trust — promotion / demotion based on counters + thresholds
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.evaluate_trust(
    p_agent_family text, p_pipeline_family text, p_model text
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_score        stewards.trust_scores%ROWTYPE;
    v_new_level    text;
    v_t2j_required int;
    v_j2m_required int;
    v_demote       boolean;
    v_overrides_since_promo int := 0;
BEGIN
    SELECT * INTO v_score
      FROM stewards.trust_scores
     WHERE agent_family = p_agent_family
       AND pipeline_family = p_pipeline_family
       AND model = p_model
       FOR UPDATE;

    IF NOT FOUND THEN
        RETURN 'trainee';
    END IF;

    v_new_level := v_score.trust_level;

    SELECT required_successes INTO v_t2j_required
      FROM stewards.trust_thresholds WHERE transition='trainee_to_journeyman';
    SELECT required_successes INTO v_j2m_required
      FROM stewards.trust_thresholds WHERE transition='journeyman_to_master';
    SELECT demote_on_override INTO v_demote
      FROM stewards.trust_thresholds WHERE transition='trainee_to_journeyman';

    -- Demotion check: compare human_overrides counter against the
    -- snapshot stored in metrics jsonb of the most recent promotion
    -- transition for this cell. If counter is higher, demote.
    -- (The counter is the single source of truth; gate_overrides
    -- table inserts ARE expected to call trust_record_override which
    -- bumps the counter, but the demotion logic stays counter-driven
    -- so manual record_override calls work too.)
    IF v_score.trust_level <> 'trainee' AND v_demote THEN
        SELECT coalesce((metrics->>'overrides')::int, 0)
          INTO v_overrides_since_promo
          FROM stewards.trust_transitions
         WHERE agent_family = p_agent_family
           AND pipeline_family = p_pipeline_family
           AND model = p_model
           AND to_level = v_score.trust_level
         ORDER BY at DESC LIMIT 1;

        v_overrides_since_promo := coalesce(v_overrides_since_promo, 0);

        IF v_score.human_overrides > v_overrides_since_promo THEN
            v_new_level := CASE v_score.trust_level
                WHEN 'master'     THEN 'journeyman'
                WHEN 'journeyman' THEN 'trainee'
                ELSE v_score.trust_level
            END;
        END IF;
    END IF;

    -- Promotion check (only if not demoting)
    IF v_new_level = v_score.trust_level THEN
        IF v_score.trust_level = 'trainee'
           AND v_score.successful_completions >= v_t2j_required
           AND v_score.human_overrides = 0 THEN
            v_new_level := 'journeyman';
        ELSIF v_score.trust_level = 'journeyman'
           AND v_score.successful_completions >= (v_t2j_required + v_j2m_required)
           AND v_score.human_overrides = coalesce(v_overrides_since_promo, 0) THEN
            v_new_level := 'master';
        END IF;
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
                 'failed',     v_score.failed_completions,
                 'overrides',  v_score.human_overrides
             ));
    ELSE
        UPDATE stewards.trust_scores
           SET last_evaluated_at = now()
         WHERE agent_family = p_agent_family
           AND pipeline_family = p_pipeline_family
           AND model = p_model;
    END IF;

    RETURN v_new_level;
END;
$func$;

COMMENT ON FUNCTION stewards.evaluate_trust(text, text, text) IS
'Phase 5f (E.2): apply promotion/demotion rules from trust_thresholds. Called by record_* helpers AND can be invoked manually for re-evaluation. Returns the new (or unchanged) trust level.';

-- ---------------------------------------------------------------------
-- (5) trust_adjust — manual level change (D-E2 requires justification)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trust_adjust(
    p_agent_family    text,
    p_pipeline_family text,
    p_model           text,
    p_new_level       text,
    p_actor           text,
    p_justification   text
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_score stewards.trust_scores%ROWTYPE;
BEGIN
    IF p_new_level NOT IN ('trainee','journeyman','master') THEN
        RAISE EXCEPTION 'trust_adjust: invalid level %', p_new_level;
    END IF;
    IF p_justification IS NULL OR length(trim(p_justification)) < 10 THEN
        RAISE EXCEPTION 'trust_adjust: justification required (>= 10 chars) per D-E2';
    END IF;

    SELECT * INTO v_score
      FROM stewards.trust_scores
     WHERE agent_family = p_agent_family
       AND pipeline_family = p_pipeline_family
       AND model = p_model
       FOR UPDATE;

    IF NOT FOUND THEN
        INSERT INTO stewards.trust_scores
            (agent_family, pipeline_family, model, trust_level)
        VALUES
            (p_agent_family, p_pipeline_family, p_model, p_new_level);

        INSERT INTO stewards.trust_transitions
            (agent_family, pipeline_family, model, from_level, to_level,
             transition_kind, actor, justification, metrics)
        VALUES
            (p_agent_family, p_pipeline_family, p_model,
             'trainee', p_new_level, 'manual', p_actor, p_justification,
             jsonb_build_object('successful', 0, 'failed', 0, 'overrides', 0));

        RETURN p_new_level;
    END IF;

    IF v_score.trust_level = p_new_level THEN
        RETURN p_new_level;  -- no-op
    END IF;

    UPDATE stewards.trust_scores
       SET trust_level = p_new_level, last_evaluated_at = now()
     WHERE agent_family = p_agent_family
       AND pipeline_family = p_pipeline_family
       AND model = p_model;

    INSERT INTO stewards.trust_transitions
        (agent_family, pipeline_family, model, from_level, to_level,
         transition_kind, actor, justification, metrics)
    VALUES
        (p_agent_family, p_pipeline_family, p_model,
         v_score.trust_level, p_new_level, 'manual', p_actor, p_justification,
         jsonb_build_object(
             'successful', v_score.successful_completions,
             'failed',     v_score.failed_completions,
             'overrides',  v_score.human_overrides
         ));

    RETURN p_new_level;
END;
$func$;

COMMENT ON FUNCTION stewards.trust_adjust(text, text, text, text, text, text) IS
'Phase 5f (E.2): manual trust level change with required justification (D-E2). Creates the trust_scores row if missing. Logs to trust_transitions with kind=manual.';
