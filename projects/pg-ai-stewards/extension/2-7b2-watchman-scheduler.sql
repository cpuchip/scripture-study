-- =====================================================================
-- Phase 2.7b.2 — Watchman scheduler (decision logic in SQL)
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: SEVENTH file — 2-6a/b/c, 2-7a, 3a, 2-7b1,
-- 2-7b2).
--
-- Builds on:
--   - 2.7b.1 (watchman_config singleton, watchman_pass_start,
--             watchman_passes table, completion trigger)
--
-- This file adds:
--   1. New columns on stewards.watchman_config for time-based scheduling.
--   2. stewards.watchman_should_fire() — pure-SQL decision function
--      that returns 'cron' | 'pressure' | 'idle' | NULL.
--   3. stewards.watchman_scheduler_inputs() — observability helper
--      that returns the inputs feeding the decision (dirty count,
--      hours since last pass, hours since last human session).
--
-- The Rust bgworker (in lib.rs) calls watchman_should_fire() on a
-- ~60s tick. If non-NULL, it calls watchman_pass_start(p_trigger:=...).
-- All schedule semantics live here in SQL — Rust just polls and dispatches.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Add scheduling columns to watchman_config.
-- ---------------------------------------------------------------------
ALTER TABLE stewards.watchman_config
    ADD COLUMN IF NOT EXISTS schedule_enabled boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS schedule_min_interval_hours int NOT NULL DEFAULT 168,
    ADD COLUMN IF NOT EXISTS schedule_preferred_dow_utc int,    -- 0=Sun..6=Sat, NULL=any
    ADD COLUMN IF NOT EXISTS schedule_preferred_hour_utc int,   -- 0..23, NULL=any
    ADD COLUMN IF NOT EXISTS schedule_pass_limit int NOT NULL DEFAULT 5,
    -- Cooldown after a pressure-triggered pass (prevents thrashing
    -- when dirty_queue stays high for a long time).
    ADD COLUMN IF NOT EXISTS schedule_pressure_cooldown_hours int NOT NULL DEFAULT 1,
    -- Cooldown after an idle-triggered pass.
    ADD COLUMN IF NOT EXISTS schedule_idle_cooldown_hours int NOT NULL DEFAULT 24;

-- DOW range guard: -1..6 where -1 represents NULL via the column type.
-- Skip CHECK on dow/hour columns — NULL is valid (any day / any hour).
-- Range validation happens in CLI input parsing instead.

-- Default the preferred slot to Sunday 03:00 UTC (Sabbath cron).
UPDATE stewards.watchman_config
   SET schedule_preferred_dow_utc  = COALESCE(schedule_preferred_dow_utc, 0),
       schedule_preferred_hour_utc = COALESCE(schedule_preferred_hour_utc, 3)
 WHERE id = 1;

COMMENT ON COLUMN stewards.watchman_config.schedule_enabled IS
'Master kill switch for the bgworker scheduler. true=auto-fire passes, false=manual only. Default true (the point of the experiment), but the human owns the cost.';

COMMENT ON COLUMN stewards.watchman_config.schedule_min_interval_hours IS
'Minimum hours between time-based (cron) passes. Default 168 = weekly. Ignored when pressure or idle trigger fires.';

COMMENT ON COLUMN stewards.watchman_config.schedule_preferred_dow_utc IS
'Preferred day of week (UTC) for cron pass: 0=Sunday..6=Saturday. NULL = any day. Default 0 (Sabbath).';

COMMENT ON COLUMN stewards.watchman_config.schedule_preferred_hour_utc IS
'Preferred hour (UTC, 0..23) for cron pass. NULL = any hour. Default 3 = 03:00 UTC.';

COMMENT ON COLUMN stewards.watchman_config.schedule_pass_limit IS
'Default p_limit for scheduler-fired passes. Default 5 docs/pass.';

-- ---------------------------------------------------------------------
-- watchman_scheduler_inputs() — observability helper.
--
-- Returns the live values feeding the decision. Used by both
-- watchman_should_fire() (internal) and the CLI scheduler-status
-- command (debugging).
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.watchman_scheduler_inputs()
RETURNS TABLE (
    schedule_enabled              boolean,
    dirty_count                   int,
    dirty_threshold               int,
    hours_since_last_pass         numeric,
    schedule_min_interval_hours   int,
    schedule_preferred_dow_utc    int,
    schedule_preferred_hour_utc   int,
    now_dow_utc                   int,
    now_hour_utc                  int,
    hours_since_last_human_session numeric,
    idle_threshold_hours          int,
    in_progress_pass_id           text,
    in_progress_pass_age_hours    numeric
)
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_now timestamptz := now();
BEGIN
    RETURN QUERY
    SELECT
        cfg.schedule_enabled,
        (SELECT count(*)::int FROM stewards.dirty_queue),
        cfg.dirty_threshold,
        CASE WHEN cfg.last_pass_at IS NULL THEN NULL
             ELSE EXTRACT(EPOCH FROM (v_now - cfg.last_pass_at)) / 3600
        END::numeric,
        cfg.schedule_min_interval_hours,
        cfg.schedule_preferred_dow_utc,
        cfg.schedule_preferred_hour_utc,
        EXTRACT(DOW FROM (v_now AT TIME ZONE 'UTC'))::int,
        EXTRACT(HOUR FROM (v_now AT TIME ZONE 'UTC'))::int,
        (SELECT EXTRACT(EPOCH FROM (v_now - max(s.last_active_at))) / 3600
           FROM stewards.sessions s
          WHERE s.kind = 'chat')::numeric,
        cfg.idle_threshold_hours,
        (SELECT p.pass_id
           FROM stewards.watchman_passes p
          WHERE p.status = 'in_progress'
          ORDER BY p.started_at DESC
          LIMIT 1),
        (SELECT EXTRACT(EPOCH FROM (v_now - p.started_at)) / 3600
           FROM stewards.watchman_passes p
          WHERE p.status = 'in_progress'
          ORDER BY p.started_at DESC
          LIMIT 1)::numeric
      FROM stewards.watchman_config cfg
     WHERE cfg.id = 1;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_scheduler_inputs() IS
'Phase 2.7b.2: returns the live values feeding watchman_should_fire(). Used by the CLI for "why isn''t it firing?" debugging.';

-- ---------------------------------------------------------------------
-- watchman_should_fire() — the decision function.
--
-- Returns:
--   'pressure' if dirty_queue exceeds threshold AND last pass is older
--              than the pressure cooldown
--   'cron'     if enough time has passed since the last pass AND we're
--              inside the preferred DOW/hour window
--   'idle'     if no human session has run for idle_threshold_hours
--              AND last pass is older than the idle cooldown
--   NULL       if schedule_enabled is false, OR a pass is currently
--              in_progress (less than 1h old), OR no trigger fires
--
-- Order matters: pressure > cron > idle. We check pressure first so
-- a heavily-dirty corpus drives passes faster than weekly.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.watchman_should_fire()
RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_inputs     RECORD;
    v_cfg        stewards.watchman_config%ROWTYPE;
BEGIN
    -- Fetch config + scheduler inputs in one read each.
    SELECT * INTO v_cfg
      FROM stewards.watchman_config WHERE id = 1;
    IF v_cfg.id IS NULL OR NOT v_cfg.schedule_enabled THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_inputs FROM stewards.watchman_scheduler_inputs();

    -- Don't pile up. If a pass started in the last hour and is still
    -- in_progress, wait for it to finish (or for the bgworker reaper
    -- to mark it errored).
    IF v_inputs.in_progress_pass_id IS NOT NULL
       AND coalesce(v_inputs.in_progress_pass_age_hours, 0) < 1 THEN
        RETURN NULL;
    END IF;

    -- Pressure: dirty_queue exceeds threshold AND we're past the
    -- pressure cooldown since last pass. Prevents thrashing when the
    -- dirty count stays high for a long time.
    IF v_inputs.dirty_count >= v_cfg.dirty_threshold
       AND (v_inputs.hours_since_last_pass IS NULL
            OR v_inputs.hours_since_last_pass
                >= v_cfg.schedule_pressure_cooldown_hours) THEN
        RETURN 'pressure';
    END IF;

    -- Time-based (cron). Two gates:
    --   1. Enough time since last pass.
    --   2. We're inside the preferred DOW + hour window.
    -- NULL preferred values match anything (so "every 168h regardless
    -- of DOW/hour" works by setting both to NULL).
    IF (v_inputs.hours_since_last_pass IS NULL
        OR v_inputs.hours_since_last_pass
            >= v_cfg.schedule_min_interval_hours)
       AND (v_cfg.schedule_preferred_dow_utc IS NULL
            OR v_inputs.now_dow_utc = v_cfg.schedule_preferred_dow_utc)
       AND (v_cfg.schedule_preferred_hour_utc IS NULL
            OR v_inputs.now_hour_utc = v_cfg.schedule_preferred_hour_utc)
    THEN
        RETURN 'cron';
    END IF;

    -- Idle: no human session activity for >= idle_threshold_hours,
    -- AND last pass is older than the idle cooldown. Skipped when
    -- idle_threshold_hours is 0 (disable idle trigger).
    IF v_cfg.idle_threshold_hours > 0
       AND (v_inputs.hours_since_last_pass IS NULL
            OR v_inputs.hours_since_last_pass
                >= v_cfg.schedule_idle_cooldown_hours) THEN
        -- hours_since_last_human_session IS NULL when no human chat
        -- session has ever been recorded — treat as "infinitely idle".
        IF v_inputs.hours_since_last_human_session IS NULL
           OR v_inputs.hours_since_last_human_session
               >= v_cfg.idle_threshold_hours THEN
            RETURN 'idle';
        END IF;
    END IF;

    RETURN NULL;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_should_fire() IS
'Phase 2.7b.2: returns the trigger reason if a Watchman pass should fire now (one of cron|pressure|idle), NULL otherwise. Called by the bgworker scheduler tick every ~60s. All schedule semantics live here, not in Rust.';

-- ---------------------------------------------------------------------
-- watchman_scheduler_fire() — convenience for the bgworker.
--
-- Calls watchman_should_fire(); if non-NULL, calls watchman_pass_start
-- with the trigger reason and the configured pass limit. Returns the
-- new pass_id (or NULL if no trigger).
--
-- Centralizes the "decide → fire" path so the Rust side is one SPI call.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.watchman_scheduler_fire()
RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_reason  text;
    v_cfg     stewards.watchman_config%ROWTYPE;
    v_pass_id text;
BEGIN
    v_reason := stewards.watchman_should_fire();
    IF v_reason IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_cfg FROM stewards.watchman_config WHERE id = 1;

    v_pass_id := stewards.watchman_pass_start(
        p_limit        => v_cfg.schedule_pass_limit,
        p_provider     => NULL,
        p_model        => NULL,
        p_agent_family => NULL,
        p_actor        => 'scheduler',
        p_trigger      => v_reason,
        p_token_budget => NULL
    );

    RAISE NOTICE 'watchman scheduler fired (%): pass_id=%', v_reason, v_pass_id;
    RETURN v_pass_id;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_scheduler_fire() IS
'Phase 2.7b.2: convenience for the bgworker scheduler tick. Calls watchman_should_fire(); if non-NULL, calls watchman_pass_start() with the trigger reason. Returns the new pass_id or NULL.';
