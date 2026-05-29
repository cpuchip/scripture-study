-- =====================================================================
-- Batch J.8.a — Dispatch model/provider fallback chain (Path A + Path C)
-- =====================================================================
-- Generalizes work_item_dispatch_stage so stages[0].model can be NULL.
-- Per ratified path A + C (2026-05-29):
--
--   Resolution order for both model and provider:
--     1. work_items.{model,provider}_override  (caller-set; Path A hook)
--     2. stages[0].{model,provider}             (lens-declared default)
--     3. pipelines.metadata->>'default_{model,provider}'  (pipeline default)
--     4. stewards.catalog_default_{model,provider}(...)   (system fallback)
--
-- Existing pipelines with stages.model set are unaffected — the COALESCE
-- short-circuits at layer 2 and the work_queue payload is byte-identical.
-- Only NULL stages.model triggers the new fallback layers. This matches
-- the conservative-migration shape used in Phase 4b (which left existing
-- pipelines untouched while adding the model_override column).
--
-- Path A note: work_items.model_override already exists (4b-dispatch-override
-- line 130). The "input override" hook this batch adds is start_brainstorm()
-- writing model_override on spawned children — no new dispatcher arg needed.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Catalog-default helpers (system-level last-resort fallback).
-- ---------------------------------------------------------------------
-- Returns a sensible default when nothing else is set. These are the
-- absolute floor — every higher layer (override / stage / pipeline) wins.
-- Lives in SQL (not a config table) so the resolution path is fully
-- introspectable via psql, and adding a new provider doesn't require a
-- migration to seed a config row.

CREATE OR REPLACE FUNCTION stewards.catalog_default_provider()
RETURNS text LANGUAGE sql IMMUTABLE AS $$
    SELECT 'opencode_go'::text
$$;

COMMENT ON FUNCTION stewards.catalog_default_provider() IS
'Batch J.8.a: substrate-wide default provider when no higher layer specifies. Returns opencode_go (the only registered provider as of 2026-05-29). Update when local LM Studio / Ollama provider rows are added.';

CREATE OR REPLACE FUNCTION stewards.catalog_default_model(p_provider text)
RETURNS text LANGUAGE sql IMMUTABLE AS $$
    SELECT CASE p_provider
        WHEN 'opencode_go' THEN 'kimi-k2.6'  -- today's existing brainstorm split default for 2 of 4 lenses
        WHEN 'lm_studio'   THEN NULL          -- no canonical local default; caller must specify
        WHEN 'ollama'      THEN NULL          -- no canonical local default; caller must specify
        ELSE NULL
    END
$$;

COMMENT ON FUNCTION stewards.catalog_default_model(text) IS
'Batch J.8.a: substrate-wide default model for a given provider when no higher layer specifies. opencode_go=kimi-k2.6 (matches today''s brainstorm split). Local providers return NULL — they have no canonical default and require explicit caller selection.';


-- ---------------------------------------------------------------------
-- 2. Replace work_item_dispatch_stage with 4-layer fallback chain.
-- ---------------------------------------------------------------------
-- Signature unchanged from 4b — no DROP needed; CREATE OR REPLACE.
-- The only change is in the model/provider resolution block (was lines
-- 130-131 in 4b-dispatch-override; now a 4-layer COALESCE preceded by a
-- pipelines.metadata lookup).

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

    -- Status gate (unchanged from 4b).
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

    -- J.8.a: pipeline-level metadata for default_model / default_provider lookup.
    SELECT metadata INTO v_pipeline_meta
      FROM stewards.pipelines
     WHERE family = v_wi.pipeline_family;

    v_agent := v_stage->>'agent_family';

    -- J.8.a: 4-layer resolution chain (input → stages → pipeline → catalog).
    -- Provider resolves first because catalog_default_model takes provider as input.
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

    -- agent_family must always be set (no fallback chain — agent identity
    -- is the lens itself; you cannot defer the lens choice). model + provider
    -- can fall back; if all 4 layers yield NULL the dispatch is unsafe.
    IF v_agent IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % missing agent_family',
            p_work_item_id, v_wi.current_stage;
    END IF;

    IF v_model IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % could not resolve model — checked work_items.model_override, stages.model, pipelines.metadata.default_model, catalog_default_model(%) — all NULL',
            p_work_item_id, v_wi.current_stage, v_provider;
    END IF;

    IF v_provider IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % could not resolve provider — checked work_items.provider_override, stages.provider, pipelines.metadata.default_provider, catalog_default_provider() — all NULL',
            p_work_item_id, v_wi.current_stage;
    END IF;

    -- Remainder unchanged from 4b — session, input templating, payload, work_queue insert.

    v_session_id := substring(
        'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
        || '--' || v_wi.current_stage
        FROM 1 FOR 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('work_item %s stage %s', v_wi.id, v_wi.current_stage),
            'agent')
    ON CONFLICT (id) DO NOTHING;

    -- Input resolution priority (unchanged from 3c3 / 4b):
    --   1. Explicit p_user_input override (CLI / steward retry guidance).
    --   2. Stage's input_template rendered against work_item state.
    --   3. work_item.input.user_input field (legacy fallback).
    --   4. Stringified work_item.input (last-resort fallback).
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
'Batch J.8.a: dispatches a stage. Resolves model + provider via 4-layer fallback chain (work_items override → stages → pipelines.metadata → catalog_default). Existing pipelines with stages.model set are unaffected. Honors p_allow_failed_status for steward retries.';


-- =====================================================================
-- Acceptance (verify before commit):
--
--   1. Existing dispatch unchanged. Calling work_item_dispatch_stage on
--      a pending work_item whose pipeline has stages[0].model set
--      produces an identical work_queue payload (same provider, same
--      model, same body) to the pre-J.8.a version.
--
--   2. NULL stages.model resolves via pipeline metadata. Insert a
--      pipeline with stages[0].model = NULL and metadata.default_model
--      = 'gpt-5-mini'; dispatch a work_item against it; work_queue
--      payload uses 'gpt-5-mini'.
--
--   3. NULL all the way down resolves via catalog. Insert a pipeline
--      with stages[0].model = NULL, no metadata.default_model;
--      dispatch; work_queue uses kimi-k2.6 (catalog default for
--      opencode_go).
--
--   4. work_items.model_override still wins everything. Insert a
--      pipeline with stages[0].model = 'kimi-k2.6'; UPDATE the
--      work_item to model_override = 'qwen3.6-plus'; dispatch;
--      payload uses qwen3.6-plus.
--
--   5. All four layers NULL raises. Pipeline with NULL stages.model,
--      NULL metadata.default_model, provider that returns NULL from
--      catalog_default_model (e.g. 'lm_studio'); dispatch raises with
--      the named-layer error message.
-- =====================================================================
