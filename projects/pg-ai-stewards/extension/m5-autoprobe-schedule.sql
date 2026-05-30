-- =====================================================================
-- Batch M.5 — Auto-probe scheduling (makes the M.4 probe periodic)
-- =====================================================================
-- M.4 built the probe mechanism; this gives it a cadence so model_capability
-- stays current without a human triggering it.
--
-- enqueue_due_model_probes() finds models that are unprobed or stale and
-- enqueues a probe for each (capped, deduped, cap-aware). A guarded trigger
-- on watchman_passes calls it whenever the watchman fires — so probing rides
-- the existing scheduler cadence (and pauses when the soak is paused, since
-- no passes fire then) WITHOUT modifying watchman_scheduler_fire. The trigger
-- swallows all errors so a probe-scheduling hiccup can never break a pass.
--
-- Cost: a probe is one short reply. Free models cost $0; paid models a small
-- fraction of a cent; gemini probes are skipped when its cap is exceeded
-- (and probes are direct work_queue inserts, so the staleness window + p_max
-- are the only throttle — tune p_staleness up to probe less often).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. enqueue_due_model_probes(staleness, max) -> count enqueued
-- ---------------------------------------------------------------------
-- "Due" = a priced model that either has no capability row yet, or whose
-- last_probed_at is older than p_staleness. Unprobed models first (NULLS
-- FIRST), then oldest. Skips providers over their enforced spend cap, and
-- skips models that already have a probe in flight (dedup).
CREATE OR REPLACE FUNCTION stewards.enqueue_due_model_probes(
    p_staleness interval DEFAULT interval '7 days',
    p_max       int      DEFAULT 3
) RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_rec    record;
    v_count  int := 0;
BEGIN
    FOR v_rec IN
        SELECT mp.provider, mp.model
          FROM (SELECT DISTINCT provider, model FROM stewards.model_pricing) mp
          LEFT JOIN stewards.model_capability mc
            ON mc.provider = mp.provider AND mc.model = mp.model
         WHERE (mc.last_probed_at IS NULL
                OR mc.last_probed_at < now() - p_staleness)
           AND NOT stewards.provider_cap_exceeded(mp.provider)
         ORDER BY mc.last_probed_at ASC NULLS FIRST, mp.provider, mp.model
         LIMIT p_max
    LOOP
        -- Dedup: don't pile a second probe for a model already in flight.
        IF NOT EXISTS (
            SELECT 1 FROM stewards.work_queue
             WHERE kind = 'chat'
               AND status NOT IN ('done', 'error')
               AND payload -> '_probe' ->> 'provider' = v_rec.provider
               AND payload -> '_probe' ->> 'model'    = v_rec.model
        ) THEN
            PERFORM stewards.enqueue_model_probe(v_rec.provider, v_rec.model);
            v_count := v_count + 1;
        END IF;
    END LOOP;

    RETURN v_count;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_due_model_probes(interval, int) IS
'Batch M.5: enqueue probes for up to p_max priced models that are unprobed or older than p_staleness, skipping cap-exceeded providers and models with a probe already in flight. Returns the count enqueued.';


-- ---------------------------------------------------------------------
-- 2. Guarded trigger: ride the watchman cadence.
-- ---------------------------------------------------------------------
-- Fires when a watchman pass is created. enqueue_due_model_probes is self-
-- throttling (staleness + p_max + dedup), so firing every pass is fine. All
-- errors are swallowed — model-probe scheduling must NEVER abort a pass.
CREATE OR REPLACE FUNCTION stewards.trigger_schedule_due_model_probes()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_n int;
BEGIN
    BEGIN
        v_n := stewards.enqueue_due_model_probes();
        IF v_n > 0 THEN
            RAISE NOTICE 'auto-probe: enqueued % due model probe(s) on watchman pass %',
                v_n, NEW.pass_id;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'auto-probe scheduling skipped (non-fatal): %', SQLERRM;
    END;
    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS watchman_passes_schedule_model_probes ON stewards.watchman_passes;

CREATE TRIGGER watchman_passes_schedule_model_probes
AFTER INSERT ON stewards.watchman_passes
FOR EACH ROW
EXECUTE FUNCTION stewards.trigger_schedule_due_model_probes();

COMMENT ON FUNCTION stewards.trigger_schedule_due_model_probes() IS
'Batch M.5: on watchman-pass creation, enqueue any due model probes. Errors are swallowed so probe scheduling never breaks a watchman pass.';


-- =====================================================================
-- Acceptance (verify before commit):
--   1. enqueue_due_model_probes(interval '0', 2) enqueues up to 2 probes
--      (everything is "stale" at 0 interval) and returns the count; a second
--      immediate call returns 0 for those two (dedup — still in flight).
--   2. Calling it again with the same 2 in flight does not double-enqueue.
--   3. The watchman_passes trigger exists and is guarded (a forced error in
--      enqueue_due_model_probes does not abort watchman_pass_start).
-- =====================================================================
