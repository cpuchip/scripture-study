-- =====================================================================
-- Batch R.3 — per-call max_tokens + input-scoped tools_disabled in dispatch
-- =====================================================================
-- Carries the live M.2 work_item_dispatch_stage forward verbatim and adds
-- two body/payload injections at enqueue time:
--
--  1. body.max_tokens  ← COALESCE(input.max_tokens, stage.max_tokens)
--     D-RL5: 32k per-API-call output ceiling for redline (reasoning tokens
--     count against it). max_tokens is PER CALL, not a work-item total —
--     cost_cap_micro + token_budget remain the runaway guards. Reading
--     stage.max_tokens is SAFE: no existing pipeline sets it (only `redline`
--     does, R.1), so no existing dispatch changes.
--
--  2. payload.tools_disabled ← input.tools_disabled (INPUT-LEVEL ONLY)
--     The bgworker honors payload.tools_disabled (Phase C.6) to strip the
--     `tools` block — without this, a redline panel is OFFERED fs_search etc.
--     and a model can loop on tool calls (the exact empties failure). We read
--     it from the work_item INPUT, not the stage, ON PURPOSE: 10 existing
--     pipelines (planning, research-*, yt-*, revise-proposal, thummim-define,
--     agent-proposal) declare stage.tools_disabled=true but have ALWAYS run
--     with tools because this function never propagated it. That is a real
--     latent cost leak, but fixing it flips tools off across the live soak —
--     a behavior change for Michael to ratify separately, NOT bundled here.
--     start_panel_redline (R.4) sets input.tools_disabled=true per child, so
--     redline gets clean tools-off with zero impact on the 10.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.work_item_dispatch_stage(
    p_work_item_id           uuid,
    p_user_input             text DEFAULT NULL,
    p_allow_failed_status    boolean DEFAULT false
) RETURNS bigint
LANGUAGE plpgsql AS $function$
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
    -- R.3 dispatch-body knobs
    v_max_tokens     text;
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

    -- R.3 (1): per-call output ceiling. input override wins; else stage default
    -- (only `redline` sets stage.max_tokens, so existing pipelines are unchanged).
    v_max_tokens := COALESCE(v_wi.input->>'max_tokens', v_stage->>'max_tokens');
    IF v_max_tokens IS NOT NULL AND v_max_tokens ~ '^[0-9]+$' THEN
        v_payload := jsonb_set(v_payload, '{body,max_tokens}', to_jsonb(v_max_tokens::int));
    END IF;

    -- R.3 (2): input-scoped tools-off. Read from INPUT only (NOT stage) so the
    -- 10 pipelines that declare stage.tools_disabled keep their current behavior;
    -- the bgworker strips the tools block when payload.tools_disabled=true.
    IF (v_wi.input->>'tools_disabled')::boolean IS TRUE THEN
        v_payload := v_payload || jsonb_build_object('tools_disabled', true);
    END IF;

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
$function$;

COMMENT ON FUNCTION stewards.work_item_dispatch_stage(uuid, text, boolean) IS
'R.3: on the M.2 base, injects body.max_tokens (input override or stage default — only redline sets a stage default) and payload.tools_disabled (input-level only, to avoid disturbing the 10 pipelines that declare stage.tools_disabled). Then J.8.a resolution + M.2 capability substitution + J.11 spend-cap gate, all unchanged.';


-- =====================================================================
-- Acceptance (R.3):
--   1. A redline work_item with input.max_tokens absent → body.max_tokens=32000
--      (stage default), body has NO tools when input.tools_disabled=true.
--   2. input.max_tokens='8000' overrides → body.max_tokens=8000.
--   3. A non-redline pipeline (no stage.max_tokens, no input.tools_disabled)
--      → body has NO max_tokens, tools unchanged (existing behavior preserved).
-- =====================================================================
