-- =====================================================================
-- Batch M.2 — Capability-aware dispatch (substitute-and-log)
-- =====================================================================
-- Wires M.1's model_usable() into the dispatch chokepoint. When a stage
-- resolves (via the J.8.a 4-layer chain) to a model marked unusable, the
-- dispatcher SUBSTITUTES a usable model for the same provider and logs the
-- swap — instead of enqueuing a chat that comes back empty.
--
-- Ratified 2026-05-29: substitute-and-log (not raise, not skip). Work keeps
-- flowing; nothing silently empties; the swap is visible in
-- stewards.model_substitutions with a reason.
--
-- Substitution precedence (pick_usable_model):
--   1. the resolved model, if usable          (no substitution — the norm)
--   2. catalog_default_model(provider), if usable
--   3. first_usable_model(provider)            (cheapest priced + usable)
--   4. NULL -> RAISE (provider has no usable model at all)
--
-- Logging reuses the L.1.1.15 (l29) model_substitutions table + its AFTER
-- INSERT trigger as the single writer. A new nullable `reason` column lets
-- capability swaps carry their "why"; l29's passive pipeline-vs-requested
-- detections leave it NULL. The dispatcher stashes a `_capability_substitution`
-- marker in the work_queue payload; the trigger reads it, writes one row with
-- the reason, and skips its normal comparison so there is no double-log.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. pick_usable_model — the resolved model, or a usable substitute, or NULL.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.pick_usable_model(p_provider text, p_model text)
RETURNS text LANGUAGE sql STABLE AS $$
    SELECT CASE
        WHEN stewards.model_usable(p_provider, p_model) THEN p_model
        WHEN stewards.catalog_default_model(p_provider) IS NOT NULL
             AND stewards.model_usable(p_provider, stewards.catalog_default_model(p_provider))
            THEN stewards.catalog_default_model(p_provider)
        ELSE stewards.first_usable_model(p_provider)
    END;
$$;

COMMENT ON FUNCTION stewards.pick_usable_model(text, text) IS
'Batch M.2: returns p_model if usable; else the provider catalog default if usable; else the cheapest usable model; else NULL. The substitution decision for work_item_dispatch_stage.';


-- ---------------------------------------------------------------------
-- 2. model_substitutions.reason — capability swaps record their "why".
-- ---------------------------------------------------------------------
ALTER TABLE stewards.model_substitutions ADD COLUMN IF NOT EXISTS reason text;

COMMENT ON COLUMN stewards.model_substitutions.reason IS
'Batch M.2: why the substitution happened. NULL for l29 passive pipeline-vs-requested detections; "capability: ..." for M.2 unusable-model swaps.';


-- ---------------------------------------------------------------------
-- 3. l29 trigger, reason-aware. Capability swaps carry a payload marker
--    (_capability_substitution = {from, to, reason}); log that and skip
--    the normal comparison so the swap is recorded exactly once.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.trigger_log_model_substitution()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_pipeline_family text;
    v_stage_name      text;
    v_pipeline_model  text;
    v_requested       text;
    v_work_item_id    text;
    v_session_id      text;
    v_cap             jsonb;
BEGIN
    v_pipeline_family := NEW.payload ->> '_pipeline_family';
    v_stage_name      := NEW.payload ->> '_stage_name';
    v_work_item_id    := NEW.payload ->> '_work_item_id';
    v_session_id      := NEW.payload ->> 'session_id';

    -- M.2: capability substitution carries its own marker + reason. Log it
    -- and return — do NOT fall through to the pipeline-vs-requested compare,
    -- which would double-log the same swap.
    v_cap := NEW.payload -> '_capability_substitution';
    IF v_cap IS NOT NULL THEN
        INSERT INTO stewards.model_substitutions
            (work_queue_id, work_item_id, pipeline_family, stage_name,
             pipeline_model, requested_model, session_id, reason)
        VALUES
            (NEW.id,
             CASE WHEN v_work_item_id ~ '^[0-9a-f-]{36}$' THEN v_work_item_id::uuid ELSE NULL END,
             v_pipeline_family, v_stage_name,
             v_cap ->> 'from', v_cap ->> 'to', v_session_id,
             'capability: ' || COALESCE(v_cap ->> 'reason', 'model marked unusable'));

        RAISE NOTICE 'capability substitution: %/% %->% (% , wq=%)',
            v_pipeline_family, v_stage_name, v_cap ->> 'from', v_cap ->> 'to',
            v_cap ->> 'reason', NEW.id;
        RETURN NEW;
    END IF;

    -- l29 original behavior: passive pipeline-declared vs requested compare.
    v_requested := NEW.payload ->> 'requested_model';
    IF v_requested IS NULL THEN RETURN NEW; END IF;
    IF v_pipeline_family IS NULL OR v_stage_name IS NULL THEN RETURN NEW; END IF;

    SELECT s ->> 'model' INTO v_pipeline_model
      FROM stewards.pipelines p,
           LATERAL jsonb_array_elements(p.stages) s
     WHERE p.family = v_pipeline_family
       AND (s ->> 'name') = v_stage_name
     LIMIT 1;

    IF v_pipeline_model IS NULL OR v_pipeline_model = v_requested THEN
        RETURN NEW;
    END IF;

    INSERT INTO stewards.model_substitutions
        (work_queue_id, work_item_id, pipeline_family, stage_name,
         pipeline_model, requested_model, session_id)
    VALUES
        (NEW.id,
         CASE WHEN v_work_item_id ~ '^[0-9a-f-]{36}$' THEN v_work_item_id::uuid ELSE NULL END,
         v_pipeline_family, v_stage_name,
         v_pipeline_model, v_requested, v_session_id);

    RAISE NOTICE 'model substitution: pipeline=%/% declared=% but requested=% (wq=%)',
        v_pipeline_family, v_stage_name, v_pipeline_model, v_requested, NEW.id;

    RETURN NEW;
END;
$FN$;

COMMENT ON FUNCTION stewards.trigger_log_model_substitution() IS
'Batch M.2 (was L.1.1.15): single writer to model_substitutions. Capability swaps (payload._capability_substitution) log with a reason and skip the passive compare; otherwise the original pipeline-declared-vs-requested detection runs (reason NULL).';


-- ---------------------------------------------------------------------
-- 4. work_item_dispatch_stage — J.11 body carried forward verbatim, with
--    the capability substitution inserted after model/provider resolve.
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
    -- M.2 capability substitution state
    v_resolved_model text;
    v_sub_model      text;
    v_cap_detail     text;
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

    -- M.2: capability gate. If the resolved model is marked unusable,
    -- substitute a usable one for the same provider (catalog default ->
    -- cheapest usable) and remember the swap so it is logged at enqueue.
    v_resolved_model := v_model;
    IF NOT stewards.model_usable(v_provider, v_model) THEN
        v_sub_model := stewards.pick_usable_model(v_provider, v_model);
        IF v_sub_model IS NULL THEN
            RAISE EXCEPTION 'work_item %: resolved model %/% is marked unusable and the provider has no usable substitute — dispatch refused. Inspect stewards.model_capability.',
                p_work_item_id, v_provider, v_model;
        END IF;
        SELECT probe_detail INTO v_cap_detail
          FROM stewards.model_capability
         WHERE provider = v_provider AND model = v_resolved_model;
        v_model := v_sub_model;
    END IF;

    -- J.11: enforced prepaid spend-cap gate (provider-level; unchanged).
    IF stewards.provider_cap_exceeded(v_provider) THEN
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

    -- M.2: attach the substitution marker so the l29 trigger logs the swap
    -- (with reason) exactly once and skips its passive compare.
    IF v_model IS DISTINCT FROM v_resolved_model THEN
        v_payload := v_payload || jsonb_build_object(
            '_capability_substitution', jsonb_build_object(
                'from',   v_resolved_model,
                'to',     v_model,
                'reason', COALESCE(v_cap_detail, 'model marked unusable')
            )
        );
    END IF;

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
'Batch M.2: adds a capability gate (model_usable) after the J.8.a 4-layer resolution — an unusable resolved model is substituted (pick_usable_model) for a usable same-provider one and the swap is logged via the l29 trigger. Then the J.11 enforced spend-cap gate. Existing usable-model dispatch is byte-identical (no marker added when no substitution).';


-- =====================================================================
-- Acceptance (verify before commit):
--   1. Usable path unchanged: dispatch a work_item resolving to kimi-k2.6
--      enqueues a chat with requested_model=kimi-k2.6 and NO
--      _capability_substitution marker; no model_substitutions row.
--   2. Substitution: a work_item with model_override='glm-5' enqueues a
--      chat with requested_model = a usable model (catalog default
--      kimi-k2.6, since it is usable), payload carries the marker, and one
--      model_substitutions row exists with from=glm-5, reason LIKE 'capability:%'.
--   3. No usable substitute: temporarily mark every opencode_go model
--      unusable -> dispatch RAISEs the "no usable substitute" error.
--      (Don't leave the DB in that state.)
-- =====================================================================
