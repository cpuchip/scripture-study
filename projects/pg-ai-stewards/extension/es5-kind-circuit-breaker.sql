-- =====================================================================
-- ES.1.s3 — bgworker crash-loop circuit breaker (CF-3)
-- =====================================================================
-- When a work_queue row reliably crashes the bgworker, the postmaster
-- respawns the worker, which picks up the same class of row and
-- crashes again — a tight restart loop (observed during the
-- bacteriopolis incident: embed bigint=text crash, ~1s cycle).
--
-- This breaker: the startup reaper records one "crash" per distinct
-- KIND it reaps. After N consecutive crashes of a kind (no successful
-- completion in between), that kind is paused for a cooldown. The
-- claim query skips paused kinds. A successful completion of a kind
-- resets its counter.
--
-- Crash-loop detection, not single-failure detection: one crash → +1,
-- no pause. A genuine loop accumulates to the threshold because the
-- reaper runs on every restart.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. kind_circuit_breaker table.
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.kind_circuit_breaker (
    kind                text PRIMARY KEY,
    consecutive_crashes int  NOT NULL DEFAULT 0,
    paused_until        timestamptz,
    last_crash_at       timestamptz,
    last_reset_at       timestamptz,
    updated_at          timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.kind_circuit_breaker IS
'ES.1.s3 (CF-3): per-work-kind crash-loop breaker. The startup reaper records one crash per distinct kind reaped; after 5 consecutive crashes a kind is paused for a cooldown. The bgworker claim query skips paused kinds. A successful completion resets the kind''s counter.';


-- ---------------------------------------------------------------------
-- 2. record_kind_crash — called once per distinct kind by the reaper.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.record_kind_crash(p_kind text)
RETURNS void LANGUAGE plpgsql AS $FN$
DECLARE
    v_threshold constant int      := 5;
    v_cooldown  constant interval := interval '10 minutes';
    v_count     int;
BEGIN
    INSERT INTO stewards.kind_circuit_breaker
        (kind, consecutive_crashes, last_crash_at, updated_at)
    VALUES (p_kind, 1, now(), now())
    ON CONFLICT (kind) DO UPDATE
       SET consecutive_crashes = stewards.kind_circuit_breaker.consecutive_crashes + 1,
           last_crash_at       = now(),
           updated_at          = now()
    RETURNING consecutive_crashes INTO v_count;

    IF v_count >= v_threshold THEN
        UPDATE stewards.kind_circuit_breaker
           SET paused_until = now() + v_cooldown,
               updated_at   = now()
         WHERE kind = p_kind;
        RAISE WARNING 'kind_circuit_breaker: kind=% PAUSED until % after % consecutive crashes',
            p_kind, now() + v_cooldown, v_count;
    END IF;
END;
$FN$;

COMMENT ON FUNCTION stewards.record_kind_crash(text) IS
'ES.1.s3: increment a kind''s consecutive-crash counter. At 5, pause the kind for 10 minutes. Called once per distinct kind by the bgworker startup reaper.';


-- ---------------------------------------------------------------------
-- 3. record_kind_success — reset a kind's counter on a clean completion.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.record_kind_success(p_kind text)
RETURNS void LANGUAGE plpgsql AS $FN$
BEGIN
    UPDATE stewards.kind_circuit_breaker
       SET consecutive_crashes = 0,
           paused_until        = NULL,
           last_reset_at       = now(),
           updated_at          = now()
     WHERE kind = p_kind
       AND (consecutive_crashes > 0 OR paused_until IS NOT NULL);
END;
$FN$;

COMMENT ON FUNCTION stewards.record_kind_success(text) IS
'ES.1.s3: reset a kind''s crash counter + clear any pause. Called by the bgworker after a work_queue row of that kind completes successfully. A no-op when the kind is already healthy.';


-- ---------------------------------------------------------------------
-- 4. paused_kinds — convenience for the claim query / observability.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.kind_is_paused(p_kind text)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS(
        SELECT 1 FROM stewards.kind_circuit_breaker
         WHERE kind = p_kind AND paused_until > now()
    )
$$;

COMMENT ON FUNCTION stewards.kind_is_paused(text) IS
'ES.1.s3: true if the kind is currently within a circuit-breaker pause window.';


-- =====================================================================
-- End of es5-kind-circuit-breaker.sql
-- =====================================================================
