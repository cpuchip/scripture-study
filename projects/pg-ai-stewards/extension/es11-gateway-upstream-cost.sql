-- =====================================================================
-- ES.3.s5 — Capture the gateway-reported upstream inference cost
-- =====================================================================
-- OpenCode Zen streams `usage.cost_details.upstream_inference_cost` —
-- the real price the upstream provider charged for the call. (The
-- top-level `cost` field is 0: OpenCode Go is subscription-billed, so
-- per-request cost to us is zero — the detail is the meaningful
-- measured number.) ES.6 streaming already captures the whole `usage`
-- object into result.response; this phase extracts the upstream cost
-- into a queryable column alongside the substrate's own rate×token
-- estimate, so estimate-vs-actual is visible.
--
-- The bgworker passes it as a new trailing param to record_cost_event.
-- The param has a DEFAULT, so this migration is safe to apply before
-- OR after the bgworker rebuild — a 10-arg call still resolves.
-- =====================================================================

-- 1. New column — gateway upstream cost, micro-dollars. NULL when the
--    gateway didn't report it (non-Zen providers, older responses).
ALTER TABLE stewards.cost_events
  ADD COLUMN IF NOT EXISTS upstream_micro_dollars bigint;

COMMENT ON COLUMN stewards.cost_events.upstream_micro_dollars IS
'ES.3.s5: gateway-reported upstream inference cost in micro-dollars (usage.cost_details.upstream_inference_cost x 1e6). The real measured cost; micro_dollars is the substrate''s rate x token estimate. NULL when unreported.';

-- 2. record_cost_event — add the trailing p_upstream_micro param.
--    DROP the 10-arg signature first: a CREATE with an extra
--    DEFAULT param would otherwise leave both overloads and make
--    10-arg calls ambiguous.
DROP FUNCTION IF EXISTS stewards.record_cost_event(
    uuid, integer, text, text, integer, integer, integer, integer, text, text);

CREATE OR REPLACE FUNCTION stewards.record_cost_event(
    p_work_item_id      uuid,
    p_attempt_seq       integer,
    p_provider          text,
    p_model             text,
    p_input_tokens      integer,
    p_output_tokens     integer,
    p_cache_write_tokens integer DEFAULT 0,
    p_cache_read_tokens  integer DEFAULT 0,
    p_session_id        text DEFAULT NULL,
    p_notes             text DEFAULT NULL,
    p_upstream_micro    bigint DEFAULT NULL
) RETURNS bigint LANGUAGE plpgsql AS $function$
DECLARE
    v_micro      bigint;
    v_pricing_at timestamptz;
    v_id         bigint;
    v_notes      text;
BEGIN
    SELECT micro_dollars, pricing_effective_at
      INTO v_micro, v_pricing_at
      FROM stewards.compute_cost(p_provider, p_model,
                                  p_input_tokens, p_output_tokens,
                                  p_cache_write_tokens, p_cache_read_tokens);

    -- If no pricing row exists, flag in notes so the gap is visible.
    v_notes := p_notes;
    IF v_pricing_at = '-infinity'::timestamptz THEN
        v_notes := coalesce(v_notes || ' | ', '')
                 || 'no_pricing_row(' || p_provider || '/' || p_model || ')';
    END IF;

    INSERT INTO stewards.cost_events
        (work_item_id, session_id, attempt_seq, provider, model,
         input_tokens, output_tokens, cache_write_tokens, cache_read_tokens,
         micro_dollars, pricing_effective_at, notes, upstream_micro_dollars)
    VALUES
        (p_work_item_id, p_session_id, p_attempt_seq, p_provider, p_model,
         p_input_tokens, p_output_tokens, p_cache_write_tokens, p_cache_read_tokens,
         v_micro, v_pricing_at, v_notes, p_upstream_micro)
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$function$;

COMMENT ON FUNCTION stewards.record_cost_event(uuid, integer, text, text, integer, integer, integer, integer, text, text, bigint) IS
'Records a cost_event. micro_dollars is computed (compute_cost: rate x tokens); ES.3.s5 p_upstream_micro carries the gateway-reported real cost into upstream_micro_dollars.';

-- =====================================================================
-- End of es11-gateway-upstream-cost.sql
-- =====================================================================
