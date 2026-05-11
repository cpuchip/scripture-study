-- =====================================================================
-- Batch H.1.0 — substrate primitives for the pipeline-expansion arc
--
-- D-H2 (ratified 2026-05-11): per-family maturity_ladder column on
--   stewards.pipelines. Documents the ordered set of rungs each pipeline
--   is allowed to produce. The substrate today doesn't enforce ladder
--   shape (rung values live in pipeline_stage_maturity per-stage), so
--   this column is forward-compat for H.4 (fiction's collapsed ladder)
--   and for future code that wants to ask "what's the valid rung set
--   for this pipeline?"
--
-- D-H5 (ratified 2026-05-11): work_items.sabbath_enabled +
--   atonement_enabled nullable columns. NULL = inherit from pipeline;
--   true = force on; false = skip. Resolves at sabbath_dispatch +
--   maybe_enqueue_atonement entry.
-- =====================================================================

-- ---------------------------------------------------------------------
-- D-H2: maturity_ladder column on pipelines
-- ---------------------------------------------------------------------

ALTER TABLE stewards.pipelines
    ADD COLUMN IF NOT EXISTS maturity_ladder jsonb NOT NULL
    DEFAULT '["raw","researched","planned","specced","executing","verified"]'::jsonb;

COMMENT ON COLUMN stewards.pipelines.maturity_ladder IS
'D-H2 (Batch H): ordered jsonb array of maturity rung names this pipeline''s stages may produce. Default is the full six-rung ladder. Pipelines may declare a narrower or differently-ordered ladder (e.g. fiction-scene: ["premise","draft","polish"]). Substrate does not enforce ladder shape today; this column documents the intent and is forward-compat for ladder-aware code.';

-- Explicit set for existing pipelines (cosmetic; matches the default)
UPDATE stewards.pipelines
   SET maturity_ladder = '["raw","researched","planned","specced","executing","verified"]'::jsonb
 WHERE family IN ('study-write', 'study-write-qwen', 'echo-test');

-- ---------------------------------------------------------------------
-- D-H5: work_item override columns
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS sabbath_enabled   boolean NULL,
    ADD COLUMN IF NOT EXISTS atonement_enabled boolean NULL;

COMMENT ON COLUMN stewards.work_items.sabbath_enabled IS
'D-H5 (Batch H): per-work_item override for pipeline.sabbath_enabled. NULL = inherit from pipeline (default); true = force sabbath on; false = skip sabbath. Resolved at sabbath_dispatch entry.';

COMMENT ON COLUMN stewards.work_items.atonement_enabled IS
'D-H5 (Batch H): per-work_item override for pipeline.atonement_enabled. NULL = inherit from pipeline (default); true = force atonement on; false = skip atonement. Resolved at maybe_enqueue_atonement entry.';

-- ---------------------------------------------------------------------
-- D-H5: refactor sabbath_dispatch to resolve work_item override first
-- (Rest of function body verbatim from the live definition; only the
--  pipeline-flag check at the top changes.)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.sabbath_dispatch(p_work_item_id uuid)
RETURNS bigint
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline        stewards.pipelines%ROWTYPE;
    v_effective       boolean;
    v_template        text;
    v_input_summary   text;
    v_stage_summary   text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_gate_model      text := 'qwen3.6-plus';
    v_gate_provider   text := 'opencode_go';
    v_gate_agent      text := 'plan';
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'sabbath_dispatch: work_item % not found', p_work_item_id;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;

    -- D-H5: resolve work_item override first; NULL inherits from pipeline.
    v_effective := COALESCE(v_wi.sabbath_enabled, v_pipeline.sabbath_enabled);
    IF NOT v_effective THEN
        RAISE EXCEPTION 'sabbath_dispatch: sabbath not enabled (work_item override=%, pipeline=%)',
            COALESCE(v_wi.sabbath_enabled::text, 'NULL'),
            v_pipeline.sabbath_enabled;
    END IF;

    SELECT template INTO v_template FROM stewards.gate_prompts WHERE id = 'sabbath';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.sabbath template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_summary := substring(coalesce(v_wi.stage_results::text, ''), 1, 8000);

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',       v_wi.pipeline_family,
        'input_summary',         v_input_summary,
        'stage_results_summary', v_stage_summary
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--sabbath--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('sabbath work_item=%s', v_wi.id),
            'sabbath')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    v_payload := jsonb_build_object(
        'session_id',      v_session_id,
        'agent_family',    v_gate_agent,
        'requested_model', v_gate_model,
        'meta',            '{}'::jsonb,
        'body',            (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                           || jsonb_build_object('user', v_session_id),
        'tools_disabled',  true,
        '_work_item_id',   p_work_item_id::text,
        '_sabbath',        true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.sabbath_dispatch(uuid) IS
'D-H5 refactor (Batch H): resolves work_item.sabbath_enabled override first via COALESCE; NULL inherits from pipeline.sabbath_enabled. Rest of body unchanged from Phase D.';

-- ---------------------------------------------------------------------
-- D-H5: refactor maybe_enqueue_atonement to resolve work_item override
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.maybe_enqueue_atonement(p_work_item_id uuid)
RETURNS bigint
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi        stewards.work_items%ROWTYPE;
    v_pipeline  stewards.pipelines%ROWTYPE;
    v_effective boolean;
    v_work_id   bigint;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RETURN NULL;
    END IF;
    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;

    -- D-H5: work_item override first; NULL inherits from pipeline.
    v_effective := COALESCE(v_wi.atonement_enabled, v_pipeline.atonement_enabled);
    IF NOT v_effective THEN
        RETURN NULL;
    END IF;

    BEGIN
        v_work_id := stewards.atonement_dispatch(p_work_item_id);
        RETURN v_work_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'maybe_enqueue_atonement: atonement_dispatch raised: %', SQLERRM;
        RETURN NULL;
    END;
END;
$func$;

COMMENT ON FUNCTION stewards.maybe_enqueue_atonement(uuid) IS
'D-H5 refactor (Batch H): resolves work_item.atonement_enabled override first via COALESCE; NULL inherits from pipeline.atonement_enabled.';

-- ---------------------------------------------------------------------
-- D-H5: refactor atonement_dispatch too — both the gate
-- (maybe_enqueue_atonement) and the dispatch had their own independent
-- pipeline-flag check, which meant the override only worked at the gate.
-- Smoke surfaced this — fix here so override propagates end-to-end.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.atonement_dispatch(p_work_item_id uuid)
RETURNS bigint
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline        stewards.pipelines%ROWTYPE;
    v_effective       boolean;
    v_template        text;
    v_input_summary   text;
    v_stage_summary   text;
    v_actions_summary text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_gate_model      text := 'kimi-k2.6';
    v_gate_provider   text := 'opencode_go';
    v_gate_agent      text := 'plan';
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'atonement_dispatch: work_item % not found', p_work_item_id;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;

    -- D-H5: resolve work_item override first; NULL inherits from pipeline.
    v_effective := COALESCE(v_wi.atonement_enabled, v_pipeline.atonement_enabled);
    IF NOT v_effective THEN
        RAISE EXCEPTION 'atonement_dispatch: atonement not enabled (work_item override=%, pipeline=%)',
            COALESCE(v_wi.atonement_enabled::text, 'NULL'),
            v_pipeline.atonement_enabled;
    END IF;

    SELECT template INTO v_template FROM stewards.gate_prompts WHERE id = 'atonement';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.atonement template missing';
    END IF;

    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_summary := substring(coalesce(v_wi.stage_results::text, ''), 1, 6000);

    SELECT string_agg(
             '  - [' || to_char(at, 'YYYY-MM-DD HH24:MI') || '] ' || action ||
             coalesce(' (' || diagnosis || ')', '') ||
             ': ' || observation,
             E'\n' ORDER BY at DESC)
      INTO v_actions_summary
      FROM (
        SELECT at, action, diagnosis, observation
          FROM stewards.steward_actions
         WHERE work_item_id = p_work_item_id
         ORDER BY at DESC
         LIMIT 20
      ) t;

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',         v_wi.pipeline_family,
        'input_summary',           v_input_summary,
        'failure_count',           v_wi.failure_count::text,
        'quarantine_reason',       coalesce(v_wi.quarantine_reason, '(none)'),
        'steward_actions_summary', coalesce(v_actions_summary, '  (no steward actions recorded)'),
        'stage_results_summary',   v_stage_summary
    ));

    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--atonement--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('atonement work_item=%s', v_wi.id),
            'atonement')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    v_payload := jsonb_build_object(
        'session_id',      v_session_id,
        'agent_family',    v_gate_agent,
        'requested_model', v_gate_model,
        'meta',            '{}'::jsonb,
        'body',            (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                           || jsonb_build_object('user', v_session_id),
        'tools_disabled',  true,
        '_work_item_id',   p_work_item_id::text,
        '_atonement',      true
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.atonement_dispatch(uuid) IS
'D-H5 refactor (Batch H): resolves work_item.atonement_enabled override first via COALESCE; NULL inherits from pipeline. Mirrors the gate-side check in maybe_enqueue_atonement so override propagates end-to-end.';
