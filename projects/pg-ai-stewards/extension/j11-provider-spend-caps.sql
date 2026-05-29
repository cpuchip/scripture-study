-- =====================================================================
-- Batch J.11 — Provider spend caps (prepaid-balance enforcement)
-- =====================================================================
-- Adds an ENFORCED spend cap for pay-as-you-go providers (Gemini is the
-- first). Ratified 2026-05-29: prepaid-balance model + $18 cap on
-- google_gemini (Michael's real balance is ~$20; $18 leaves a buffer).
--
-- Why a new mechanism instead of cost_buckets:
--   - cost_buckets are ROLLING (session_5h/daily/weekly/monthly) and
--     their bucket_limit_micro is INFORMATIONAL ONLY — nothing enforces
--     it. A prepaid balance is not a rolling period; it only resets when
--     the human refills. So this is a dedicated, self-contained table +
--     a gate in the dispatch chokepoint.
--   - Opt-in per provider via `enforced`. Only google_gemini is enforced
--     here; opencode_go (a subscription) is untouched. No surprise blocks
--     on existing work.
--
-- Depends on J.11's bgworker fix (stream_options.include_usage) — without
-- it, Gemini cost_events don't record and the gate's running-sum stays 0.
-- That fix ships in the same batch (pg rebuild).
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. provider_spend_caps — one row per provider with an enforced cap.
-- ---------------------------------------------------------------------
-- spend-since-refill model: the gate sums cost_events for the provider
-- with at >= `since`, and refuses dispatch once that sum >= cap_micro.
-- Refill = move `since` to now() (and optionally raise cap_micro) via
-- stewards.provider_cap_refill().
CREATE TABLE IF NOT EXISTS stewards.provider_spend_caps (
    provider    text PRIMARY KEY,
    cap_micro   bigint NOT NULL CHECK (cap_micro >= 0),
    since       timestamptz NOT NULL DEFAULT now(),
    enforced    boolean NOT NULL DEFAULT false,
    notes       text,
    updated_at  timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.provider_spend_caps IS
'J.11: enforced prepaid-balance spend caps per provider. The dispatch gate (work_item_dispatch_stage) refuses a provider whose cost_events sum since `since` >= cap_micro AND enforced=true. Refill via provider_cap_refill(). Distinct from cost_buckets (rolling + informational).';

COMMENT ON COLUMN stewards.provider_spend_caps.since IS
'Refill epoch. Spend is summed from cost_events.at >= since. provider_cap_refill() moves this to now().';

-- ---------------------------------------------------------------------
-- 2. provider_spend_since(provider) -> micro_dollars spent since refill
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.provider_spend_since(p_provider text)
RETURNS bigint LANGUAGE sql STABLE AS $$
    SELECT coalesce(sum(ce.micro_dollars), 0)::bigint
      FROM stewards.cost_events ce
      JOIN stewards.provider_spend_caps c ON c.provider = ce.provider
     WHERE ce.provider = p_provider
       AND ce.at >= c.since;
$$;

COMMENT ON FUNCTION stewards.provider_spend_since(text) IS
'J.11: micro-dollars spent on a provider since its cap row''s refill epoch. 0 if no cap row.';

-- ---------------------------------------------------------------------
-- 3. provider_cap_exceeded(provider) -> boolean
-- ---------------------------------------------------------------------
-- True only when an ENFORCED cap row exists AND spend-since-refill has
-- reached it. Providers without a cap row (or enforced=false) return
-- false — never gated.
CREATE OR REPLACE FUNCTION stewards.provider_cap_exceeded(p_provider text)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS (
        SELECT 1
          FROM stewards.provider_spend_caps c
         WHERE c.provider = p_provider
           AND c.enforced
           AND (SELECT coalesce(sum(ce.micro_dollars), 0)
                  FROM stewards.cost_events ce
                 WHERE ce.provider = p_provider
                   AND ce.at >= c.since) >= c.cap_micro
    );
$$;

COMMENT ON FUNCTION stewards.provider_cap_exceeded(text) IS
'J.11: true if the provider has an enforced cap and spend-since-refill has reached it. Checked by work_item_dispatch_stage before enqueuing a chat.';

-- ---------------------------------------------------------------------
-- 4. provider_cap_refill(provider, new_cap_micro?) — top up / reset
-- ---------------------------------------------------------------------
-- Moves the refill epoch to now() (so spend-since resets to 0) and
-- optionally sets a new cap. Use after topping up the real balance.
CREATE OR REPLACE FUNCTION stewards.provider_cap_refill(
    p_provider      text,
    p_new_cap_micro bigint DEFAULT NULL
) RETURNS stewards.provider_spend_caps
LANGUAGE plpgsql AS $$
DECLARE
    v_row stewards.provider_spend_caps;
BEGIN
    UPDATE stewards.provider_spend_caps
       SET since      = now(),
           cap_micro  = COALESCE(p_new_cap_micro, cap_micro),
           updated_at = now()
     WHERE provider = p_provider
    RETURNING * INTO v_row;

    IF v_row.provider IS NULL THEN
        RAISE EXCEPTION 'provider_cap_refill: no cap row for provider %', p_provider;
    END IF;

    RAISE NOTICE 'provider_cap_refill: % refilled — since=now(), cap=% micro ($%.2f)',
        p_provider, v_row.cap_micro, (v_row.cap_micro / 1000000.0);
    RETURN v_row;
END;
$$;

COMMENT ON FUNCTION stewards.provider_cap_refill(text, bigint) IS
'J.11: top up a provider cap. Resets the spend-since-refill clock (since=now()) and optionally sets a new cap_micro. Run after refilling the real prepaid balance.';

-- ---------------------------------------------------------------------
-- 5. Gate the dispatcher. Carry the J.8.a 4-layer model/provider
--    resolution forward verbatim; add a cap check after provider
--    resolves and before the chat is enqueued.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_dispatch_stage(
    p_work_item_id           uuid,
    p_user_input             text DEFAULT NULL,
    p_allow_failed_status    boolean DEFAULT false
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi             stewards.work_items%ROWTYPE;
    v_stage          jsonb;
    v_pipeline_meta  jsonb;
    v_agent          text;
    v_model          text;
    v_provider       text;
    v_session_id     text;
    v_user_input     text;
    v_body           jsonb;
    v_payload        jsonb;
    v_work_id        bigint;
    v_was_failed     boolean := false;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    IF v_wi.status NOT IN ('pending', 'awaiting_review')
       AND NOT (p_allow_failed_status AND v_wi.status = 'failed')
    THEN
        RAISE EXCEPTION 'work_item %: cannot dispatch from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_was_failed := (v_wi.status = 'failed');

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    SELECT metadata INTO v_pipeline_meta
      FROM stewards.pipelines
     WHERE family = v_wi.pipeline_family;

    v_agent := v_stage->>'agent_family';

    -- J.8.a: 4-layer resolution (input -> stages -> pipeline -> catalog).
    v_provider := COALESCE(
        v_wi.provider_override,
        v_stage->>'provider',
        v_pipeline_meta->>'default_provider',
        stewards.catalog_default_provider()
    );

    v_model := COALESCE(
        v_wi.model_override,
        v_stage->>'model',
        v_pipeline_meta->>'default_model',
        stewards.catalog_default_model(v_provider)
    );

    IF v_agent IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % missing agent_family',
            p_work_item_id, v_wi.current_stage;
    END IF;
    IF v_model IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % could not resolve model — checked work_items.model_override, stages.model, pipelines.metadata.default_model, catalog_default_model(%) — all NULL',
            p_work_item_id, v_wi.current_stage, v_provider;
    END IF;
    IF v_provider IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % could not resolve provider',
            p_work_item_id, v_wi.current_stage;
    END IF;

    -- J.11: enforced prepaid spend-cap gate. Refuse before enqueuing so
    -- no money is spent past the cap. Only fires for providers with an
    -- enforced cap row (e.g. google_gemini); all others pass through.
    IF stewards.provider_cap_exceeded(v_provider) THEN
        -- plpgsql RAISE supports only `%` substitution (no printf specifiers
        -- like %.2f or %L), so pre-round the dollar values and quote the
        -- provider literally in-string.
        RAISE EXCEPTION 'work_item %: provider % spend cap reached ($% spent since refill / $% cap) — dispatch refused. Top up + reset with: SELECT stewards.provider_cap_refill(''%'');',
            p_work_item_id, v_provider,
            round(stewards.provider_spend_since(v_provider) / 1000000.0, 4),
            round((SELECT cap_micro FROM stewards.provider_spend_caps WHERE provider = v_provider) / 1000000.0, 2),
            v_provider;
    END IF;

    v_session_id := substring(
        'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
        || '--' || v_wi.current_stage
        FROM 1 FOR 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('work_item %s stage %s', v_wi.id, v_wi.current_stage),
            'agent')
    ON CONFLICT (id) DO NOTHING;

    IF p_user_input IS NOT NULL THEN
        v_user_input := p_user_input;
    ELSE
        v_user_input := stewards.render_stage_input(p_work_item_id);
        IF v_user_input IS NULL THEN
            v_user_input := coalesce(
                v_wi.input->>'user_input',
                v_wi.input::text
            );
        END IF;
    END IF;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_user_input, v_model);

    v_body := stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_agent,
        'requested_model',    v_model,
        'meta',               v_body->'_meta',
        'body',               (v_body - '_meta')
                              || jsonb_build_object('user', v_session_id),
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_provider, v_payload)
    RETURNING id INTO v_work_id;

    UPDATE stewards.work_items
       SET status      = 'in_progress',
           session_ids = session_ids || v_session_id,
           updated_at  = now()
     WHERE id = p_work_item_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_dispatch_stage(uuid, text, boolean) IS
'J.11: adds an enforced prepaid spend-cap gate (provider_cap_exceeded) before enqueue, on top of the J.8.a 4-layer model/provider fallback chain. Gemini-only enforced via provider_spend_caps; all other providers pass through unchanged.';

-- ---------------------------------------------------------------------
-- 6. Seed: google_gemini $18 enforced cap (ratified 2026-05-29).
-- ---------------------------------------------------------------------
-- ON CONFLICT does NOT reset `since` (preserves the refill clock on
-- re-run); it does refresh cap/enforced/notes.
INSERT INTO stewards.provider_spend_caps (provider, cap_micro, since, enforced, notes)
VALUES (
    'google_gemini', 18000000, now(), true,
    'Prepaid balance cap — $18 of ~$20 real Google balance (buffer for crossing-call + tracking imprecision). spend-since-refill. Top up: SELECT stewards.provider_cap_refill(''google_gemini''[, <new_cap_micro>]);'
)
ON CONFLICT (provider) DO UPDATE
SET cap_micro  = EXCLUDED.cap_micro,
    enforced   = EXCLUDED.enforced,
    notes      = EXCLUDED.notes,
    updated_at = now();

-- =====================================================================
-- Acceptance:
--   1. provider_cap_exceeded('opencode_go') = false (no cap row) — never gated.
--   2. provider_cap_exceeded('google_gemini') = false initially (spend 0 < $18).
--   3. Set cap to 1 micro -> provider_cap_exceeded('google_gemini') = true
--      once any gemini cost_event exists; dispatch to gemini RAISEs.
--   4. provider_cap_refill('google_gemini', 18000000) resets since + cap.
--
-- Brainstorm note: a gemini lens hitting the cap RAISEs inside
-- spawn_children, which is caught by on_maturity_verified's EXCEPTION
-- handler (logged; 0 children spawned for that brainstorm). Direct/CLI/
-- MCP dispatch surfaces the RAISE message directly. Pre-flight cap check
-- in start_brainstorm is a possible future polish.
-- =====================================================================
