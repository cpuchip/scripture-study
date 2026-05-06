-- =====================================================================
-- Phase 2.7b.2 — Pure-SQL verification of watchman_should_fire().
--
-- Saves the current config, walks watchman_should_fire() through every
-- branch by mutating the config in place, then restores. No model
-- tokens spent — just decision-function semantics.
-- =====================================================================

\set ON_ERROR_STOP on

BEGIN;

-- Snapshot current config so we can restore.
CREATE TEMP TABLE _saved_cfg AS
    SELECT * FROM stewards.watchman_config WHERE id = 1;

\echo '=== TRIAL 1: schedule_enabled = false → NULL ==='
UPDATE stewards.watchman_config SET schedule_enabled = false WHERE id = 1;
SELECT 'TRIAL 1' AS trial, stewards.watchman_should_fire() AS got, 'NULL' AS expected;

\echo
\echo '=== TRIAL 2: enabled, dirty heavy, past cooldown → pressure ==='
UPDATE stewards.watchman_config
   SET schedule_enabled = true,
       dirty_threshold = 50,
       schedule_pressure_cooldown_hours = 1,
       schedule_min_interval_hours = 168,
       schedule_preferred_dow_utc = 0,
       schedule_preferred_hour_utc = 3,
       last_pass_at = now() - interval '2 hours'
 WHERE id = 1;
SELECT 'TRIAL 2' AS trial, stewards.watchman_should_fire() AS got, 'pressure' AS expected;

\echo
\echo '=== TRIAL 3: pressure suppressed by high threshold → NULL ==='
UPDATE stewards.watchman_config
   SET dirty_threshold = 9999
 WHERE id = 1;
SELECT 'TRIAL 3' AS trial, stewards.watchman_should_fire() AS got, 'NULL (pressure suppressed; cron not in window; idle cooldown not met)' AS expected;

\echo
\echo '=== TRIAL 4: pressure suppressed, cron window matches → cron ==='
UPDATE stewards.watchman_config
   SET schedule_min_interval_hours = 0,
       schedule_preferred_dow_utc = NULL,    -- any day
       schedule_preferred_hour_utc = NULL    -- any hour
 WHERE id = 1;
SELECT 'TRIAL 4' AS trial, stewards.watchman_should_fire() AS got, 'cron' AS expected;

\echo
\echo '=== TRIAL 5: cron suppressed by min_interval → NULL ==='
UPDATE stewards.watchman_config
   SET schedule_min_interval_hours = 168,
       last_pass_at = now() - interval '12 hours'
 WHERE id = 1;
SELECT 'TRIAL 5' AS trial, stewards.watchman_should_fire() AS got, 'NULL (12h < 168h; idle cooldown 24h not met)' AS expected;

\echo
\echo '=== TRIAL 6: idle path — past idle cooldown, no human sessions → idle ==='
UPDATE stewards.watchman_config
   SET schedule_min_interval_hours = 168,
       schedule_idle_cooldown_hours = 1,
       idle_threshold_hours = 1,
       last_pass_at = now() - interval '2 hours',
       dirty_threshold = 9999     -- pressure off
 WHERE id = 1;
SELECT 'TRIAL 6' AS trial, stewards.watchman_should_fire() AS got, 'idle' AS expected;

\echo
\echo '=== TRIAL 7: idle disabled (idle_threshold_hours=0) → NULL ==='
UPDATE stewards.watchman_config
   SET idle_threshold_hours = 0
 WHERE id = 1;
SELECT 'TRIAL 7' AS trial, stewards.watchman_should_fire() AS got, 'NULL' AS expected;

\echo
\echo '=== TRIAL 8: in_progress pass <1h old → NULL (don''t pile up) ==='
-- Reset to "would normally fire pressure"
UPDATE stewards.watchman_config
   SET schedule_enabled = true,
       dirty_threshold = 50,
       schedule_pressure_cooldown_hours = 1,
       last_pass_at = now() - interval '2 hours'
 WHERE id = 1;
-- Insert a fake in_progress pass started 30 min ago.
INSERT INTO stewards.watchman_passes
    (pass_id, started_at, trigger, provider, model, agent_family,
     token_budget, actor, status, doc_count_planned)
VALUES
    ('inverse-test-inflight', now() - interval '30 minutes',
     'manual', 'opencode_go', 'kimi-k2.6', 'watchman-consolidator',
     50000, 'inverse-test', 'in_progress', 5);
SELECT 'TRIAL 8' AS trial, stewards.watchman_should_fire() AS got, 'NULL (in-flight pass blocks)' AS expected;

\echo
\echo '=== TRIAL 9: in_progress pass >1h old → pressure (allowed) ==='
UPDATE stewards.watchman_passes
   SET started_at = now() - interval '90 minutes'
 WHERE pass_id = 'inverse-test-inflight';
SELECT 'TRIAL 9' AS trial, stewards.watchman_should_fire() AS got, 'pressure (in-flight pass too old to block)' AS expected;

-- Cleanup synthetic inflight pass.
DELETE FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-inflight';

-- Restore the original config.
UPDATE stewards.watchman_config c
   SET schedule_enabled                = s.schedule_enabled,
       schedule_min_interval_hours     = s.schedule_min_interval_hours,
       schedule_preferred_dow_utc      = s.schedule_preferred_dow_utc,
       schedule_preferred_hour_utc     = s.schedule_preferred_hour_utc,
       schedule_pass_limit             = s.schedule_pass_limit,
       schedule_pressure_cooldown_hours = s.schedule_pressure_cooldown_hours,
       schedule_idle_cooldown_hours    = s.schedule_idle_cooldown_hours,
       dirty_threshold                 = s.dirty_threshold,
       idle_threshold_hours            = s.idle_threshold_hours,
       last_pass_at                    = s.last_pass_at
  FROM _saved_cfg s
 WHERE c.id = 1;

\echo
\echo '=== Restored config; final should_fire ==='
SELECT 'restored' AS trial, stewards.watchman_should_fire() AS final_decision;

COMMIT;

\echo 'verify-2-7b2-decision: done.'
