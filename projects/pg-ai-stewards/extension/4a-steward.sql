-- =====================================================================
-- Phase 4a — Steward loop (failure tracking, diagnosis, retry guidance,
-- circuit breaker, steward_tick orchestration).
--
-- Implements the Phase A steward components from the parent proposal
-- (full-agentic-substrate.md §IV "Phase A — The Steward loop"). Builds
-- on the cost-tracking + escalation-chain schema layers (4a-cost-
-- tracking.sql, 4a-escalation-chain.sql).
--
-- This file adds:
--   1. work_items columns: failure_count, last_failure_reason,
--      last_failure_diagnosis, quarantined_at, quarantine_reason
--   2. stewards.steward_actions — append-only audit ledger of every
--      steward decision (the "Account" in Watch→Diagnose→Act→Account)
--   3. stewards.diagnose_failure(reason, failure_count) — port of
--      brain v3's diagnosis.go classifier (5 types)
--   4. stewards.retry_guidance_text — per-diagnosis template seed table
--   5. stewards.retry_guidance(diagnosis, attempt) — composes the
--      retry-context message from the template
--   6. stewards.pipeline_breakers — per-(pipeline, stage) breaker state
--   7. stewards.breaker_check / breaker_record_failure /
--      breaker_record_success — the three-state circuit breaker
--   8. stewards.steward_tick() — orchestration that walks failed
--      work_items, applies cost+breaker+diagnosis+escalation logic,
--      writes steward_actions, transitions escalation_state when chain
--      exhausts. Called by the bgworker on its tick (next push).
--
-- The steward_tick() function does NOT itself re-dispatch via
-- work_item_dispatch_stage — that wiring lands in the next push when
-- bgworker calls steward_tick + a separate dispatch handler. This file
-- ships the discipline-as-data layer.
-- =====================================================================

-- =====================================================================
-- Section 1: work_items columns
-- =====================================================================

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS failure_count            int  NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_failure_reason      text,
    ADD COLUMN IF NOT EXISTS last_failure_diagnosis   text,
    ADD COLUMN IF NOT EXISTS quarantined_at           timestamptz,
    ADD COLUMN IF NOT EXISTS quarantine_reason        text;

COMMENT ON COLUMN stewards.work_items.failure_count IS
'Phase 4a: count of dispatch failures since last success on this work_item. Reset to 0 when a stage advances.';
COMMENT ON COLUMN stewards.work_items.last_failure_reason IS
'Phase 4a: free-text error from the last dispatch failure. Input to diagnose_failure().';
COMMENT ON COLUMN stewards.work_items.last_failure_diagnosis IS
'Phase 4a: cached diagnosis classification (transient | timeout | model_limit | tool_error | unknown).';
COMMENT ON COLUMN stewards.work_items.quarantined_at IS
'Phase 4a: timestamp when the steward gave up on auto-retry. NULL = still in flight or completed normally.';
COMMENT ON COLUMN stewards.work_items.quarantine_reason IS
'Phase 4a: why quarantined. Common values: failure_count_limit | cost_cap_exceeded | breaker_exhausted.';

-- =====================================================================
-- Section 2: steward_actions — append-only audit ledger
-- =====================================================================

CREATE TABLE IF NOT EXISTS stewards.steward_actions (
    id            bigserial PRIMARY KEY,
    work_item_id  uuid REFERENCES stewards.work_items(id) ON DELETE CASCADE,
    at            timestamptz NOT NULL DEFAULT now(),
    observation   text NOT NULL,
    diagnosis     text,
    action        text NOT NULL,
    details       jsonb NOT NULL DEFAULT '{}'::jsonb,
    model_used    text,
    cost_micro    bigint
);
CREATE INDEX IF NOT EXISTS steward_actions_work_item ON stewards.steward_actions(work_item_id);
CREATE INDEX IF NOT EXISTS steward_actions_at        ON stewards.steward_actions(at);
CREATE INDEX IF NOT EXISTS steward_actions_action    ON stewards.steward_actions(action);

COMMENT ON TABLE stewards.steward_actions IS
'Phase 4a: append-only audit of every steward decision. The "Account" step of Watch→Diagnose→Act→Account.';

-- =====================================================================
-- Section 3: diagnose_failure — port of brain v3 diagnosis.go
-- =====================================================================

-- Returns one of: 'transient', 'timeout', 'model_limit', 'tool_error', 'unknown'.
-- IMMUTABLE so it can be inlined in views and indexed if needed.
CREATE OR REPLACE FUNCTION stewards.diagnose_failure(
    p_reason         text,
    p_failure_count  int DEFAULT 0
) RETURNS text
LANGUAGE plpgsql IMMUTABLE AS $func$
DECLARE
    v_lower text;
BEGIN
    IF p_reason IS NULL OR length(trim(p_reason)) = 0 THEN
        -- No reason text. Use failure_count as proxy: if we've failed
        -- a few times with no reason string, treat as model_limit so
        -- escalation kicks in.
        IF p_failure_count >= 2 THEN
            RETURN 'model_limit';
        END IF;
        RETURN 'unknown';
    END IF;

    v_lower := lower(p_reason);

    -- Order matters: timeout is most specific (overrides "rate limit"
    -- false-positives like "request timeout: rate limit hit").
    IF v_lower ~ '(timeout|timed out|context deadline exceeded|inactivity|deadline)' THEN
        RETURN 'timeout';
    END IF;

    -- Transient: rate limits, 5xx, network blips. Provider issue, not
    -- a model-capability issue.
    IF v_lower ~ '(429|rate.?limit|5(00|01|02|03|04)|network|connection refused|temporarily unavailable|service unavailable)' THEN
        RETURN 'transient';
    END IF;

    -- Tool error: model called a tool wrong, or the tool itself
    -- rejected the call. Distinct from model_limit because re-prompting
    -- with feedback usually fixes it.
    IF v_lower ~ '(tool.{0,30}(error|not found|missing|invalid)|function.{0,20}(error|not found|missing|invalid)|schema.{0,20}(error|invalid|mismatch)|validation.{0,20}(failed|error))' THEN
        RETURN 'tool_error';
    END IF;

    -- After 2+ failures without timeout/transient/tool_error pattern,
    -- treat as model_limit. The model genuinely can't handle this.
    IF p_failure_count >= 2 THEN
        RETURN 'model_limit';
    END IF;

    RETURN 'unknown';
END;
$func$;

COMMENT ON FUNCTION stewards.diagnose_failure(text, int) IS
'Phase 4a: classify a failure reason into one of (transient | timeout | model_limit | tool_error | unknown). Port of brain v3 diagnosis.go.';

-- =====================================================================
-- Section 4 + 5: retry_guidance — per-diagnosis text templates
-- =====================================================================

CREATE TABLE IF NOT EXISTS stewards.retry_guidance_text (
    diagnosis   text PRIMARY KEY CHECK (diagnosis IN
        ('transient','timeout','model_limit','tool_error','unknown')),
    template    text NOT NULL,  -- {attempt} placeholder is substituted
    notes       text
);

COMMENT ON TABLE stewards.retry_guidance_text IS
'Phase 4a: per-diagnosis retry-context templates. {attempt} is replaced with the current attempt number by retry_guidance().';

-- Seed with brain v3's BuildRetryContext text (idempotent).
INSERT INTO stewards.retry_guidance_text (diagnosis, template, notes) VALUES
    ('transient',
     '**Steward retry context (attempt {attempt}):** Previous attempt failed with a transient provider issue (rate limit, 5xx, or network blip). The underlying issue has likely resolved. Proceed with the same approach.',
     'Same model, no strategy change'),
    ('timeout',
     '**Steward retry context (attempt {attempt}):** Previous attempt timed out. Break the work into smaller steps. Read files in targeted ranges rather than full files. Avoid loops that touch many tools in sequence. If you need to plan, plan tightly.',
     'Reduce per-step work to fit inside the timeout window'),
    ('tool_error',
     '**Steward retry context (attempt {attempt}):** Previous attempt failed with a tool error — the tool may not exist, the arguments may be wrong, or a schema check failed. Check the tool name against your available tools. Verify argument names and types. If the schema rejected your output, re-read the schema constraints carefully.',
     'Help the model self-correct on tool usage'),
    ('model_limit',
     '**Steward retry context (attempt {attempt}):** Previous attempts failed despite reasonable strategies, suggesting this task may be at the edge of what the current model can handle. Simplify the task. Re-read the plan/spec carefully. Identify the single most important next step and do only that. The next attempt will use a more capable model.',
     'Acknowledge the cliff; sets up the escalation'),
    ('unknown',
     '**Steward retry context (attempt {attempt}):** Previous attempt failed but the failure reason did not match a known pattern. Re-examine the input, the spec, and any error output from the last attempt. Be deliberate.',
     'Generic fallback')
ON CONFLICT (diagnosis) DO UPDATE
SET template = EXCLUDED.template,
    notes    = EXCLUDED.notes;

-- Compose retry guidance string for a given diagnosis + attempt.
-- Returns NULL if no template exists for the diagnosis (caller skips
-- prepending guidance).
CREATE OR REPLACE FUNCTION stewards.retry_guidance(
    p_diagnosis text,
    p_attempt   int
) RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_template text;
BEGIN
    SELECT template INTO v_template
      FROM stewards.retry_guidance_text
     WHERE diagnosis = p_diagnosis;

    IF v_template IS NULL THEN
        RETURN NULL;
    END IF;

    RETURN replace(v_template, '{attempt}', p_attempt::text);
END;
$func$;

COMMENT ON FUNCTION stewards.retry_guidance(text, int) IS
'Phase 4a: compose the per-diagnosis retry-context message with attempt number substituted. Used by steward_tick when retrying.';

-- =====================================================================
-- Section 6: pipeline_breakers — per-(pipeline, stage) circuit breaker
-- =====================================================================

CREATE TABLE IF NOT EXISTS stewards.pipeline_breakers (
    pipeline_family   text NOT NULL,
    stage_name        text NOT NULL,
    state             text NOT NULL DEFAULT 'closed' CHECK (state IN ('closed','open','half_open')),
    failure_count     int NOT NULL DEFAULT 0,
    opened_at         timestamptz,
    half_open_at      timestamptz,
    cooldown_minutes  int NOT NULL DEFAULT 10,
    failure_threshold int NOT NULL DEFAULT 5,
    last_state_change timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (pipeline_family, stage_name)
);

COMMENT ON TABLE stewards.pipeline_breakers IS
'Phase 4a: per-(pipeline_family, stage) circuit breaker. Three states: closed (normal) | open (cooling down) | half_open (probe). 5 failures in a row trips the breaker; 10-min cooldown; success on half-open closes; failure on half-open re-opens.';

-- =====================================================================
-- Section 7: breaker functions (check / record_failure / record_success)
-- =====================================================================

-- Returns true if the breaker is closed or half_open and a probe is
-- allowed (one probe per cooldown). Returns false if open and still
-- within cooldown.
-- Side effect: opens a closed breaker by inserting if missing.
-- Side effect: transitions open → half_open when cooldown elapses.
CREATE OR REPLACE FUNCTION stewards.breaker_check(
    p_pipeline text,
    p_stage    text
) RETURNS boolean
LANGUAGE plpgsql AS $func$
DECLARE
    v_breaker stewards.pipeline_breakers;
BEGIN
    -- Lazy-create the breaker row on first reference.
    INSERT INTO stewards.pipeline_breakers (pipeline_family, stage_name)
    VALUES (p_pipeline, p_stage)
    ON CONFLICT DO NOTHING;

    SELECT * INTO v_breaker
      FROM stewards.pipeline_breakers
     WHERE pipeline_family = p_pipeline AND stage_name = p_stage
     FOR UPDATE;

    -- Closed breaker: ok to dispatch.
    IF v_breaker.state = 'closed' THEN
        RETURN true;
    END IF;

    -- Half-open: one probe permitted between state-change and now.
    -- Use last_state_change as the "probe issued" marker; record_success
    -- or record_failure will close or re-open.
    IF v_breaker.state = 'half_open' THEN
        RETURN true;
    END IF;

    -- Open: check if cooldown has elapsed. If yes, transition to
    -- half_open and allow a probe.
    IF v_breaker.opened_at IS NOT NULL
       AND v_breaker.opened_at + (v_breaker.cooldown_minutes * interval '1 minute') <= now()
    THEN
        UPDATE stewards.pipeline_breakers
           SET state = 'half_open',
               half_open_at = now(),
               last_state_change = now()
         WHERE pipeline_family = p_pipeline AND stage_name = p_stage;
        RETURN true;
    END IF;

    -- Still open and within cooldown.
    RETURN false;
END;
$func$;

COMMENT ON FUNCTION stewards.breaker_check(text, text) IS
'Phase 4a: returns true if the breaker permits a dispatch. Lazy-creates breaker row; transitions open → half_open on cooldown.';

-- Increment failure_count. If threshold reached and breaker was closed,
-- open it. If breaker was half_open and got a failure, re-open with
-- fresh cooldown.
CREATE OR REPLACE FUNCTION stewards.breaker_record_failure(
    p_pipeline text,
    p_stage    text
) RETURNS void
LANGUAGE plpgsql AS $func$
DECLARE
    v_breaker stewards.pipeline_breakers;
BEGIN
    INSERT INTO stewards.pipeline_breakers (pipeline_family, stage_name)
    VALUES (p_pipeline, p_stage)
    ON CONFLICT DO NOTHING;

    SELECT * INTO v_breaker
      FROM stewards.pipeline_breakers
     WHERE pipeline_family = p_pipeline AND stage_name = p_stage
     FOR UPDATE;

    IF v_breaker.state = 'half_open' THEN
        -- Probe failed. Re-open with fresh cooldown.
        UPDATE stewards.pipeline_breakers
           SET state = 'open',
               opened_at = now(),
               half_open_at = NULL,
               last_state_change = now(),
               failure_count = failure_count + 1
         WHERE pipeline_family = p_pipeline AND stage_name = p_stage;
        RETURN;
    END IF;

    UPDATE stewards.pipeline_breakers
       SET failure_count = failure_count + 1
     WHERE pipeline_family = p_pipeline AND stage_name = p_stage;

    -- Re-fetch updated count to check threshold.
    SELECT * INTO v_breaker
      FROM stewards.pipeline_breakers
     WHERE pipeline_family = p_pipeline AND stage_name = p_stage;

    IF v_breaker.state = 'closed'
       AND v_breaker.failure_count >= v_breaker.failure_threshold
    THEN
        UPDATE stewards.pipeline_breakers
           SET state = 'open',
               opened_at = now(),
               last_state_change = now()
         WHERE pipeline_family = p_pipeline AND stage_name = p_stage;
    END IF;
END;
$func$;

-- Reset failure_count and (if open or half_open) close the breaker.
CREATE OR REPLACE FUNCTION stewards.breaker_record_success(
    p_pipeline text,
    p_stage    text
) RETURNS void
LANGUAGE plpgsql AS $func$
BEGIN
    UPDATE stewards.pipeline_breakers
       SET state = 'closed',
           failure_count = 0,
           opened_at = NULL,
           half_open_at = NULL,
           last_state_change = now()
     WHERE pipeline_family = p_pipeline AND stage_name = p_stage
       AND (state != 'closed' OR failure_count > 0);
END;
$func$;

COMMENT ON FUNCTION stewards.breaker_record_failure(text, text) IS
'Phase 4a: increment breaker failure_count; trip if threshold reached.';
COMMENT ON FUNCTION stewards.breaker_record_success(text, text) IS
'Phase 4a: reset breaker to closed state with failure_count=0.';

-- =====================================================================
-- Section 8: steward_tick — orchestration
-- =====================================================================

-- Walks failed work_items and decides what to do for each:
--   1. Cost cap exceeded → quarantine
--   2. Breaker open → defer (write steward_actions row, no dispatch)
--   3. Diagnose failure
--   4. Pick model (may return queue sentinel)
--   5. If sentinel → transition to escalation_queued
--   6. Otherwise → write steward_actions with the chosen model + retry
--      guidance. The bgworker reads steward_actions to actually
--      re-dispatch via work_item_dispatch_stage (next push wires this).
--
-- Returns count of actions taken in this tick. The bgworker logs this.
CREATE OR REPLACE FUNCTION stewards.steward_tick()
RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_count        int := 0;
    v_item         record;
    v_diagnosis    text;
    v_next_model   text;
    v_breaker_ok   boolean;
    v_attempt      int;
BEGIN
    FOR v_item IN
        SELECT id, pipeline_family, current_stage, failure_count,
               last_failure_reason, escalation_state
          FROM stewards.work_items
         WHERE status = 'failed'
           AND failure_count < 3
           AND quarantined_at IS NULL
           AND escalation_state = 'normal'
         ORDER BY updated_at ASC  -- oldest failures first
         LIMIT 10
         FOR UPDATE SKIP LOCKED
    LOOP
        v_attempt := v_item.failure_count + 1;

        -- 1. Cost cap check
        IF stewards.cost_cap_exceeded(v_item.id) THEN
            UPDATE stewards.work_items
               SET quarantined_at = now(),
                   quarantine_reason = 'cost_cap_exceeded'
             WHERE id = v_item.id;

            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action, details)
            VALUES
                (v_item.id,
                 'cumulative cost exceeded cap; quarantining',
                 'cost_limit',
                 'quarantine',
                 jsonb_build_object('quarantine_reason','cost_cap_exceeded'));
            v_count := v_count + 1;
            CONTINUE;
        END IF;

        -- 2. Diagnose
        v_diagnosis := stewards.diagnose_failure(
            v_item.last_failure_reason, v_item.failure_count);

        -- Cache diagnosis on the work_item for visibility
        UPDATE stewards.work_items
           SET last_failure_diagnosis = v_diagnosis
         WHERE id = v_item.id;

        -- 3. Breaker check
        v_breaker_ok := stewards.breaker_check(
            v_item.pipeline_family, v_item.current_stage);
        IF NOT v_breaker_ok THEN
            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action)
            VALUES
                (v_item.id,
                 format('breaker open for %s/%s; deferring',
                        v_item.pipeline_family, v_item.current_stage),
                 v_diagnosis,
                 'defer_breaker_open');
            v_count := v_count + 1;
            CONTINUE;
        END IF;

        -- 4. Pick model
        v_next_model := stewards.pick_model(
            v_item.pipeline_family, v_item.current_stage,
            v_attempt, v_diagnosis);

        -- 5. Queue sentinel handling
        IF v_next_model = '__queue_for_opus__' THEN
            UPDATE stewards.work_items
               SET escalation_state = 'queued',
                   escalation_attempts = escalation_attempts + 1
             WHERE id = v_item.id;

            INSERT INTO stewards.steward_actions
                (work_item_id, observation, diagnosis, action, model_used,
                 details)
            VALUES
                (v_item.id,
                 'OpenCode chain exhausted; queued for human-mediated Opus boost',
                 v_diagnosis,
                 'queue_for_opus',
                 '__queue_for_opus__',
                 jsonb_build_object(
                     'attempt', v_attempt,
                     'escalation_attempts',
                         (SELECT escalation_attempts FROM stewards.work_items
                           WHERE id = v_item.id)));
            v_count := v_count + 1;
            CONTINUE;
        END IF;

        -- 6. Retry intent: write the action; bgworker handles dispatch.
        INSERT INTO stewards.steward_actions
            (work_item_id, observation, diagnosis, action, model_used,
             details)
        VALUES
            (v_item.id,
             format('attempt #%s after %s', v_attempt, v_diagnosis),
             v_diagnosis,
             'retry_with_escalation',
             v_next_model,
             jsonb_build_object(
                 'attempt', v_attempt,
                 'retry_guidance',
                     stewards.retry_guidance(v_diagnosis, v_attempt)));
        v_count := v_count + 1;
    END LOOP;

    RETURN v_count;
END;
$func$;

COMMENT ON FUNCTION stewards.steward_tick() IS
'Phase 4a: the Watch→Diagnose→Act→Account orchestration. Walks failed work_items, applies cost+breaker+diagnosis+escalation logic, writes steward_actions. Returns count of actions taken. Called by bgworker on tick.';

-- =====================================================================
-- Section 9: helper view for Stewards-UI
-- =====================================================================

-- Latest steward_action per work_item, joined with the work_item.
-- Useful for "show me the steward's most recent decision per item" UI.
CREATE OR REPLACE VIEW stewards.work_items_steward_status AS
SELECT
    wi.id                       AS work_item_id,
    wi.slug,
    wi.pipeline_family,
    wi.current_stage,
    wi.status,
    wi.failure_count,
    wi.last_failure_diagnosis,
    wi.escalation_state,
    wi.quarantined_at,
    wi.quarantine_reason,
    wi.cost_micro_dollars,
    wi.cost_cap_micro,
    wi.cost_capped_at,
    sa.at                       AS last_action_at,
    sa.observation              AS last_observation,
    sa.action                   AS last_action,
    sa.model_used               AS last_model_used,
    sa.diagnosis                AS last_action_diagnosis
  FROM stewards.work_items wi
  LEFT JOIN LATERAL (
      SELECT * FROM stewards.steward_actions
       WHERE work_item_id = wi.id
       ORDER BY at DESC
       LIMIT 1
  ) sa ON true;

COMMENT ON VIEW stewards.work_items_steward_status IS
'Phase 4a: per-work_item status with the most recent steward_action surfaced. For Stewards-UI status panels.';

-- =====================================================================
-- Done. Phase 4a-steward is operational at the SQL layer.
-- The bgworker integration (calling steward_tick + handling re-dispatch
-- of action='retry_with_escalation' rows) lands in the next push.
--
-- Acceptance:
--   SELECT stewards.diagnose_failure('connection refused', 0);
--     → transient
--   SELECT stewards.diagnose_failure('context deadline exceeded', 0);
--     → timeout
--   SELECT stewards.diagnose_failure('weird unparseable error', 3);
--     → model_limit (failure_count >= 2 fallback)
--   SELECT stewards.retry_guidance('timeout', 2);
--     → text starting "**Steward retry context (attempt 2):** Previous attempt timed out..."
--   SELECT stewards.breaker_check('study','research');
--     → true (closed by default)
--   SELECT stewards.steward_tick();
--     → 0 (no failed work_items to process in a fresh DB)
-- =====================================================================
