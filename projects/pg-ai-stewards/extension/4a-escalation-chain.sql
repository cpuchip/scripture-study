-- =====================================================================
-- Phase 4a — Escalation Chain (D-B1 OpenCode Zen model substitution +
-- D-EC3 human-mediated escalation queue)
--
-- Implements the escalation-chain schema from
-- projects/pg-ai-stewards/.spec/proposals/escalation-chain.md.
--
-- Builds on:
--   - 4a-cost-tracking.sql (model names referenced here exist in model_pricing)
--
-- This file adds:
--   1. stewards.stage_models — per-(pipeline_family, stage) default model
--   2. stewards.model_escalation — (current_model, diagnosis) -> next_model matrix
--   3. work_items.model_override column (one-shot model pin)
--   4. work_items escalation queue columns (state machine for D-EC3 queue)
--   5. SQL function: pick_model(pipeline, stage, attempt, diagnosis) -> model
--   6. Seed: stage_models for study/lesson/dev/_gate pipelines
--   7. Seed: full model_escalation matrix Qwen→MiniMax→Kimi→GLM→queue
--
-- The sentinel '__queue_for_opus__' returned by pick_model means "transition
-- to escalation_state='queued' instead of dispatching" — handled by the
-- bgworker steward_tick (Phase A.escalation.4 work, Go-side).
-- =====================================================================

-- ---------------------------------------------------------------------
-- stage_models: per-(pipeline_family, stage_name) default model
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.stage_models (
    pipeline_family   text NOT NULL,
    stage_name        text NOT NULL,
    default_model     text NOT NULL,
    notes             text,
    PRIMARY KEY (pipeline_family, stage_name)
);

COMMENT ON TABLE stewards.stage_models IS
'Per-(pipeline_family, stage) initial model for stage dispatch. pick_model() consults this for attempt=1.';
COMMENT ON COLUMN stewards.stage_models.default_model IS
'Model name. Should match a model in stewards.model_pricing; not FK-constrained because pricing has composite PK with effective_at.';

-- ---------------------------------------------------------------------
-- model_escalation: (current_model, diagnosis) -> next_model
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.model_escalation (
    current_model     text NOT NULL,
    diagnosis         text NOT NULL CHECK (diagnosis IN
        ('transient','timeout','model_limit','tool_error','unknown')),
    attempt_threshold int NOT NULL DEFAULT 1 CHECK (attempt_threshold >= 1),
    next_model        text,  -- NULL = stay on current; '__queue_for_opus__' = sentinel
    notes             text,
    PRIMARY KEY (current_model, diagnosis),
    -- Prevent direct self-loops (catches obvious cycles; doesn't catch
    -- multi-hop cycles, but pick_model's attempt-bounded loop terminates
    -- regardless).
    CHECK (next_model IS NULL OR next_model != current_model)
);

COMMENT ON TABLE stewards.model_escalation IS
'Escalation matrix: given current_model + diagnosis, what model to retry on after attempt_threshold attempts. NULL next_model = stay; sentinel __queue_for_opus__ = enter escalation queue.';

-- ---------------------------------------------------------------------
-- work_items: model_override (one-shot pin)
-- ---------------------------------------------------------------------
ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS model_override text;

COMMENT ON COLUMN stewards.work_items.model_override IS
'Phase 4a: when non-NULL, the steward dispatches with this model regardless of pick_model. Cleared after escalation queue resolution; used as one-shot Opus boost target.';

-- ---------------------------------------------------------------------
-- work_items: escalation queue state machine (D-EC3)
-- ---------------------------------------------------------------------
ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS escalation_state         text NOT NULL DEFAULT 'normal',
    ADD COLUMN IF NOT EXISTS escalation_claimed_by    text,
    ADD COLUMN IF NOT EXISTS escalation_claimed_at    timestamptz,
    ADD COLUMN IF NOT EXISTS escalation_completed_at  timestamptz,
    ADD COLUMN IF NOT EXISTS escalation_attempts      int NOT NULL DEFAULT 0;

-- The CHECK constraint can fail to add if rows exist with invalid values.
-- Defensive: only add if not present, and the default ensures all rows
-- start with a valid value.
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'work_items_escalation_state_check'
    ) THEN
        ALTER TABLE stewards.work_items
            ADD CONSTRAINT work_items_escalation_state_check
            CHECK (escalation_state IN ('normal','queued','in_progress','failed','resolved'));
    END IF;
END;
$check$;

COMMENT ON COLUMN stewards.work_items.escalation_state IS
'Phase 4a (D-EC3): state machine for human-mediated escalation queue. normal | queued | in_progress | failed | resolved.';
COMMENT ON COLUMN stewards.work_items.escalation_claimed_by IS
'Who picked up the queued escalation: ui:zen-opus | cli:claude-code-pro | NULL.';
COMMENT ON COLUMN stewards.work_items.escalation_attempts IS
'How many times this work_item has cycled through the escalation queue.';

-- =====================================================================
-- pick_model(pipeline, stage, attempt, diagnosis) -> model name
--
-- Walks the escalation chain by attempt count. Returns:
--   - The stage_models default for (pipeline, stage) on attempt 1
--   - An escalated model after walking (attempt-1) escalation hops
--   - The sentinel '__queue_for_opus__' if the chain ends with that
-- Raises if no stage_models row exists for (pipeline, stage).
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.pick_model(
    p_pipeline_family text,
    p_stage_name      text,
    p_attempt         int,
    p_diagnosis       text DEFAULT 'initial'
) RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_current_model text;
    v_escalation    record;
    i               int;
BEGIN
    SELECT default_model INTO v_current_model
      FROM stewards.stage_models
     WHERE pipeline_family = p_pipeline_family
       AND stage_name = p_stage_name;

    IF v_current_model IS NULL THEN
        RAISE EXCEPTION 'no stage_models row for %/%',
            p_pipeline_family, p_stage_name;
    END IF;

    -- First attempt or sentinel diagnosis = no escalation; return default.
    IF p_attempt <= 1 OR p_diagnosis = 'initial' OR p_diagnosis IS NULL THEN
        RETURN v_current_model;
    END IF;

    -- Walk the chain. For each attempt past 1, look up an escalation
    -- rule for (current_model, diagnosis) whose attempt_threshold is met.
    -- If the chain returns the queue sentinel at any point, return it
    -- immediately (steward handles the state transition).
    FOR i IN 2..p_attempt LOOP
        SELECT * INTO v_escalation
          FROM stewards.model_escalation
         WHERE current_model = v_current_model
           AND diagnosis = p_diagnosis
           AND attempt_threshold <= i;

        -- No rule, or rule says "stay" (NULL next_model)
        IF v_escalation IS NULL OR v_escalation.next_model IS NULL THEN
            RETURN v_current_model;
        END IF;

        -- Sentinel: the chain ends with the human-mediated queue
        IF v_escalation.next_model = '__queue_for_opus__' THEN
            RETURN '__queue_for_opus__';
        END IF;

        v_current_model := v_escalation.next_model;
    END LOOP;

    RETURN v_current_model;
END;
$func$;

COMMENT ON FUNCTION stewards.pick_model(text, text, int, text) IS
'Phase 4a: picks the model for the next dispatch. Walks model_escalation per (attempt, diagnosis). Returns __queue_for_opus__ sentinel when chain exhausts.';

-- =====================================================================
-- Seed: stage_models for the substrate's known pipelines + _gate sentinel
-- =====================================================================

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    -- 'study' pipeline
    ('study',  'research',           'kimi-k2.6',     'general-purpose default'),
    ('study',  'outline',            'kimi-k2.6',     ''),
    ('study',  'draft',              'kimi-k2.6',     ''),
    ('study',  'verify',             'qwen3.6-plus',  'cheap binary verification'),
    -- 'lesson' pipeline
    ('lesson', 'research',           'kimi-k2.6',     ''),
    ('lesson', 'outline',            'kimi-k2.6',     ''),
    ('lesson', 'draft',              'kimi-k2.6',     ''),
    ('lesson', 'verify',             'qwen3.6-plus',  ''),
    -- 'dev' pipeline (note: dev/plan starts at top tier per spec)
    ('dev',    'plan',               'glm-5.1',       'design needs heaviest tier'),
    ('dev',    'execute',            'kimi-k2.6',     'general-purpose for code'),
    ('dev',    'verify',             'minimax-m2.7',  'mid-tier for code review'),
    -- '_gate' sentinel pipeline (consumed by Phase B gate dispatcher)
    ('_gate',  'evaluate_gate',      'qwen3.6-plus',  'cheap binary gate decision'),
    ('_gate',  'generate_scenarios', 'kimi-k2.6',     'needs creativity'),
    ('_gate',  'verify_scenarios',   'qwen3.6-plus',  'cheap pass/fail check')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
SET default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

-- =====================================================================
-- Seed: model_escalation matrix Qwen → MiniMax → Kimi → GLM → __queue_for_opus__
--
-- Per Michael's D-EC2 ratification: brain v3 defaults — model_limit
-- escalates after 1 retry (threshold=2), other failures after 2
-- (threshold=3). transient stays on same model (provider issue).
-- =====================================================================

INSERT INTO stewards.model_escalation
    (current_model,  diagnosis,    attempt_threshold, next_model,    notes) VALUES

    -- qwen3.6-plus → minimax-m2.7
    ('qwen3.6-plus', 'model_limit',  2, 'minimax-m2.7', 'always escalate up'),
    ('qwen3.6-plus', 'timeout',      3, 'minimax-m2.7', 'escalate after 2 timeouts'),
    ('qwen3.6-plus', 'tool_error',   3, 'minimax-m2.7', 'escalate after 2 tool errors'),
    ('qwen3.6-plus', 'transient',    99, NULL,           'stay; transient is provider issue'),
    ('qwen3.6-plus', 'unknown',      3, 'minimax-m2.7', ''),

    -- minimax-m2.7 → kimi-k2.6
    ('minimax-m2.7', 'model_limit',  2, 'kimi-k2.6',    'escalate to general-purpose'),
    ('minimax-m2.7', 'timeout',      3, 'kimi-k2.6',    ''),
    ('minimax-m2.7', 'tool_error',   3, 'kimi-k2.6',    ''),
    ('minimax-m2.7', 'transient',    99, NULL,           ''),
    ('minimax-m2.7', 'unknown',      3, 'kimi-k2.6',    ''),

    -- kimi-k2.6 → glm-5.1
    ('kimi-k2.6',    'model_limit',  2, 'glm-5.1',      'escalate to heaviest tier'),
    ('kimi-k2.6',    'timeout',      3, 'glm-5.1',      ''),
    ('kimi-k2.6',    'tool_error',   3, 'glm-5.1',      ''),
    ('kimi-k2.6',    'transient',    99, NULL,           ''),
    ('kimi-k2.6',    'unknown',      3, 'glm-5.1',      ''),

    -- glm-5.1 → __queue_for_opus__ (human-mediated queue, D-EC3)
    ('glm-5.1',      'model_limit',  2, '__queue_for_opus__', 'top of auto chain; queue for human-mediated Opus boost'),
    ('glm-5.1',      'timeout',      3, '__queue_for_opus__', ''),
    ('glm-5.1',      'tool_error',   3, '__queue_for_opus__', ''),
    ('glm-5.1',      'transient',    99, NULL,                 'transient stays on GLM'),
    ('glm-5.1',      'unknown',      3, '__queue_for_opus__', '')

ON CONFLICT (current_model, diagnosis) DO UPDATE
SET attempt_threshold = EXCLUDED.attempt_threshold,
    next_model        = EXCLUDED.next_model,
    notes             = EXCLUDED.notes;

-- =====================================================================
-- Done. Phase 4a-escalation is operational at the SQL layer.
-- Acceptance:
--   SELECT stewards.pick_model('study','research',1,'initial');
--     → kimi-k2.6
--   SELECT stewards.pick_model('study','research',2,'model_limit');
--     → glm-5.1   (Kimi escalates immediately on model_limit)
--   SELECT stewards.pick_model('study','research',3,'model_limit');
--     → __queue_for_opus__   (GLM escalates to queue on its own attempt 2)
--   SELECT stewards.pick_model('study','research',2,'transient');
--     → kimi-k2.6   (transient stays on same model)
--   SELECT stewards.pick_model('dev','plan',1,'initial');
--     → glm-5.1   (dev/plan starts at top of chain)
--   SELECT stewards.pick_model('nonexistent','x',1,'initial');
--     → ERROR: no stage_models row for nonexistent/x
-- =====================================================================
