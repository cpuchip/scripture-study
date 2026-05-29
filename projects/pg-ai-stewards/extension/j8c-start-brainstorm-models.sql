-- =====================================================================
-- Batch J.8.c — start_brainstorm() gains p_models override
--                + spawn_children() propagates child.model_override
-- =====================================================================
-- Per ratified path A (2026-05-29): caller can pass a per-lens model map
-- to start_brainstorm(). Map keys are short lens names (scamper, six-hats,
-- crazy8s, reverse, plus the J.9 expansions); map values are model strings.
-- NULL means: use the fallback chain from J.8.a.
--
-- Shape change in TWO places:
--
--   1. start_brainstorm() — new p_models jsonb parameter; the 4-lens
--      manifest embeds `model_override` and (optional) `provider_override`
--      into each child entry when the map specifies one.
--
--   2. spawn_children() — when a manifest child carries `model_override`
--      and/or `provider_override`, the UPDATE that sets parent_work_item_id
--      also sets work_items.model_override / provider_override BEFORE the
--      child's first dispatch_stage call. This is a general-purpose
--      fan-out improvement, not brainstorm-specific.
--
-- p_models JSON shape (flexible):
--   { "scamper": "opus-4.7" }                              -- model only
--   { "six-hats": {"model":"gpt-5","provider":"openai"} }  -- model + provider
--   { "crazy8s": null }                                    -- explicit: fallback chain
--
-- For substrate-as-of-2026-05-29 only opencode_go is registered, so the
-- string-only shorthand is the common case. Object shape supports the
-- future when LM Studio / Ollama / etc. land as registered providers.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Extend spawn_children() to propagate model/provider overrides.
-- ---------------------------------------------------------------------
-- Only change: after work_item_create + UPDATE-with-parent block, add
-- a follow-up UPDATE that sets model_override / provider_override IF
-- the manifest child specified them. Then dispatch.
-- The order matters — overrides must land BEFORE dispatch, so the
-- dispatcher sees them when it resolves the model.

CREATE OR REPLACE FUNCTION stewards.spawn_children(p_parent_id uuid)
RETURNS int LANGUAGE plpgsql AS $FN$
DECLARE
    v_parent           stewards.work_items%ROWTYPE;
    v_manifest         jsonb;
    v_manifest_raw     text;
    v_child            jsonb;
    v_child_id         uuid;
    v_count            int := 0;
    v_aggregator       jsonb;
    v_agg_id           uuid;
    v_children_arr     jsonb := '[]'::jsonb;
    v_child_pipeline   text;
    v_child_slug       text;
    v_child_input      jsonb;
    v_cost_cap         bigint;
    v_model_override   text;
    v_provider_override text;
BEGIN
    SELECT * INTO v_parent FROM stewards.work_items WHERE id = p_parent_id;
    IF v_parent.id IS NULL THEN
        RAISE EXCEPTION 'spawn_children: parent % not found', p_parent_id;
    END IF;

    v_manifest := v_parent.stage_results -> 'decompose' -> 'output';
    IF v_manifest IS NULL THEN
        RAISE EXCEPTION 'spawn_children: no decompose output on parent %', p_parent_id;
    END IF;

    IF jsonb_typeof(v_manifest) = 'string' THEN
        v_manifest_raw := v_manifest #>> '{}';
        BEGIN
            v_manifest := v_manifest_raw::jsonb;
        EXCEPTION WHEN OTHERS THEN
            RAISE EXCEPTION 'spawn_children: decompose output is not valid JSON: %', SQLERRM;
        END;
    END IF;

    IF v_manifest -> 'children' IS NULL
       OR jsonb_typeof(v_manifest -> 'children') <> 'array'
       OR jsonb_array_length(v_manifest -> 'children') = 0 THEN
        RAISE EXCEPTION 'spawn_children: manifest.children is missing or empty';
    END IF;

    IF v_manifest -> 'aggregate' IS NULL
       OR (v_manifest -> 'aggregate' ->> 'destination') IS NULL THEN
        RAISE EXCEPTION 'spawn_children: manifest.aggregate.destination is required';
    END IF;

    FOR v_child IN SELECT * FROM jsonb_array_elements(v_manifest -> 'children') LOOP
        v_child_pipeline := v_child ->> 'pipeline_family';
        v_child_slug     := v_child ->> 'slug';

        IF v_child_pipeline IS NULL OR v_child_slug IS NULL
           OR (v_child ->> 'binding_question') IS NULL THEN
            RAISE EXCEPTION 'spawn_children: child entry missing slug/pipeline_family/binding_question: %', v_child;
        END IF;

        v_child_input := jsonb_build_object(
            'binding_question', v_child ->> 'binding_question'
        );
        IF (v_child -> 'input_extra') IS NOT NULL
           AND jsonb_typeof(v_child -> 'input_extra') = 'object' THEN
            v_child_input := v_child_input || (v_child -> 'input_extra');
        END IF;

        v_child_id := stewards.work_item_create(
            p_pipeline_family => v_child_pipeline,
            p_input           => v_child_input,
            p_slug            => v_child_slug,
            p_actor           => v_parent.actor,
            p_intent_id       => v_parent.intent_id
        );

        v_cost_cap := NULL;
        IF (v_child ->> 'cost_cap_micro') IS NOT NULL THEN
            v_cost_cap := (v_child ->> 'cost_cap_micro')::bigint;
        END IF;

        UPDATE stewards.work_items
           SET parent_work_item_id = p_parent_id,
               project_association = COALESCE(
                   v_child ->> 'project_association',
                   v_parent.project_association
               ),
               cost_cap_micro = COALESCE(v_cost_cap, cost_cap_micro)
         WHERE id = v_child_id;

        -- J.8.c: propagate model + provider overrides from manifest child
        -- to the child work_item, BEFORE dispatch. NULL values are no-op
        -- (UPDATE writes NULL over NULL).
        v_model_override    := v_child ->> 'model_override';
        v_provider_override := v_child ->> 'provider_override';

        IF v_model_override IS NOT NULL OR v_provider_override IS NOT NULL THEN
            UPDATE stewards.work_items
               SET model_override    = COALESCE(v_model_override,    model_override),
                   provider_override = COALESCE(v_provider_override, provider_override)
             WHERE id = v_child_id;
        END IF;

        PERFORM stewards.work_item_dispatch_stage(v_child_id, NULL);

        v_children_arr := v_children_arr || jsonb_build_object(
            'id', v_child_id::text,
            'slug', v_child_slug,
            'binding_question', v_child ->> 'binding_question',
            'pipeline_family', v_child_pipeline
        );
        v_count := v_count + 1;
    END LOOP;

    v_aggregator := v_manifest -> 'aggregate';

    v_agg_id := stewards.work_item_create(
        p_pipeline_family => 'aggregate-children',
        p_input           => jsonb_build_object(
            'binding_question', 'Aggregate index for: ' || COALESCE(v_parent.input ->> 'binding_question', v_parent.slug),
            'parent_work_item_id', p_parent_id::text,
            'destination', v_aggregator ->> 'destination',
            'synthesis', COALESCE((v_aggregator ->> 'synthesis')::boolean, false),
            'children', v_children_arr
        ),
        p_slug            => COALESCE(v_parent.slug, p_parent_id::text) || '-aggregator',
        p_actor           => v_parent.actor,
        p_intent_id       => v_parent.intent_id
    );

    UPDATE stewards.work_items
       SET parent_work_item_id = p_parent_id,
           project_association = v_parent.project_association
     WHERE id = v_agg_id;

    RAISE NOTICE 'spawn_children: parent=% spawned % children + aggregator %',
        p_parent_id, v_count, v_agg_id;

    RETURN v_count;
END;
$FN$;

COMMENT ON FUNCTION stewards.spawn_children(uuid) IS
'Batch J.8.c: extended to propagate manifest child.model_override and child.provider_override onto the spawned child work_item, BEFORE first dispatch. Unchanged behavior for manifests that don''t set those fields (NULL no-op). General-purpose fan-out improvement, not brainstorm-specific.';


-- ---------------------------------------------------------------------
-- 2. Replace start_brainstorm() with p_models-accepting version.
-- ---------------------------------------------------------------------
-- Per Phase 4b lesson: pl/pgsql functions are keyed on signature, so
-- CREATE OR REPLACE on the new 7-arg signature would coexist with the
-- old 6-arg version and become ambiguous. DROP the old explicitly.

DROP FUNCTION IF EXISTS stewards.start_brainstorm(text, text, text, text, text, bigint);

CREATE OR REPLACE FUNCTION stewards.start_brainstorm(
    p_binding_question        text,
    p_destination             text,
    p_project_association     text    DEFAULT NULL,
    p_actor                   text    DEFAULT 'human',
    p_slug                    text    DEFAULT NULL,
    p_cost_cap_per_lens_micro bigint  DEFAULT 200000,
    p_models                  jsonb   DEFAULT NULL
)
RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_slug             text;
    v_parent_id        uuid;
    v_manifest         jsonb;
    v_lens_short_names text[] := ARRAY['scamper', 'six-hats', 'crazy8s', 'reverse'];
    v_lens             text;
    v_lens_family      text;
    v_lens_slug        text;
    v_models_entry     jsonb;
    v_model_override   text;
    v_provider_override text;
    v_child            jsonb;
    v_children_arr     jsonb := '[]'::jsonb;
BEGIN
    v_slug := COALESCE(p_slug, 'brainstorm-' || to_char(now() AT TIME ZONE 'UTC', 'YYYYMMDD-HH24MISS'));

    -- Build the children array — one entry per lens. p_models lookup pulls
    -- per-lens override (model string OR {model, provider} object).
    FOREACH v_lens IN ARRAY v_lens_short_names LOOP
        v_lens_family    := 'brainstorm-' || v_lens;
        v_lens_slug      := v_slug || '-' || v_lens;
        v_model_override := NULL;
        v_provider_override := NULL;

        IF p_models IS NOT NULL AND (p_models ? v_lens) THEN
            v_models_entry := p_models -> v_lens;
            IF jsonb_typeof(v_models_entry) = 'string' THEN
                v_model_override := v_models_entry #>> '{}';
            ELSIF jsonb_typeof(v_models_entry) = 'object' THEN
                v_model_override    := v_models_entry ->> 'model';
                v_provider_override := v_models_entry ->> 'provider';
            ELSIF jsonb_typeof(v_models_entry) = 'null' THEN
                -- explicit NULL -> use fallback chain, no override.
                NULL;
            END IF;
        END IF;

        v_child := jsonb_build_object(
            'slug',             v_lens_slug,
            'pipeline_family',  v_lens_family,
            'binding_question', p_binding_question,
            'cost_cap_micro',   p_cost_cap_per_lens_micro
        );
        IF v_model_override IS NOT NULL THEN
            v_child := v_child || jsonb_build_object('model_override', v_model_override);
        END IF;
        IF v_provider_override IS NOT NULL THEN
            v_child := v_child || jsonb_build_object('provider_override', v_provider_override);
        END IF;

        v_children_arr := v_children_arr || v_child;
    END LOOP;

    v_manifest := jsonb_build_object(
        'rationale', 'Brainstorm: 4 lenses (SCAMPER, Six Hats, Crazy 8s, Reverse) run in parallel, converged via synthesis aggregator.',
        'children', v_children_arr,
        'aggregate', jsonb_build_object(
            'destination', p_destination,
            'synthesis', true
        )
    );

    INSERT INTO stewards.work_items (
        pipeline_family, current_stage, slug, input, intent_id, actor,
        project_association, stage_results, maturity, status
    ) VALUES (
        'decompose-fanout',
        'decompose',
        v_slug,
        jsonb_build_object('binding_question', p_binding_question),
        (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
        p_actor,
        p_project_association,
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', 'brainstorm: pre-populated 4-lens manifest, no context_gather LLM call'),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned',
        'completed'
    )
    RETURNING id INTO v_parent_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_parent_id;

    RAISE NOTICE 'start_brainstorm: parent=% slug=% p_models=%',
        v_parent_id, v_slug, COALESCE(p_models::text, 'NULL');
    RETURN v_parent_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.start_brainstorm(text, text, text, text, text, bigint, jsonb) IS
'Batch J.8.c: brainstorm entry point with per-lens model override. p_models is a JSONB map keyed by short lens name (scamper / six-hats / crazy8s / reverse); values are either a model string ("opus-4.7") or a {model, provider} object. NULL/missing entries use the J.8.a fallback chain. Spawns 4-lens manifest into decompose-fanout parent, triggers immediate spawn.';


-- =====================================================================
-- Acceptance (verify before commit):
--
--   1. Backward compat: start_brainstorm(q, dest) with no p_models
--      produces identical work_queue payloads to today (each lens
--      dispatches with its default_model from pipeline metadata, which
--      matches the pre-J.8.b hardcoded model).
--
--   2. Per-lens model override:
--        start_brainstorm(
--          'How should X work?', 'projects/foo/bar.md',
--          p_models => '{"scamper":"opus-4.7","six-hats":"haiku-4.5"}'::jsonb
--        )
--      → spawned scamper child has model_override='opus-4.7',
--        six-hats child has model_override='haiku-4.5',
--        crazy8s + reverse children have model_override=NULL (fallback).
--
--   3. Object-shape entry:
--        p_models => '{"crazy8s":{"model":"gpt-5","provider":"openai"}}'::jsonb
--      → crazy8s child has model_override='gpt-5', provider_override='openai'.
--
--   4. Unknown lens name in p_models is silently ignored (no error). This
--      lets callers pass a forward-compatible map across J.9 lens
--      expansions without crashing on lens names this version doesn't
--      know about. (J.9 then extends the v_lens_short_names array.)
--
--   5. spawn_children() general path unchanged for non-brainstorm fan-outs
--      that don't carry model_override / provider_override in their
--      manifest children. The COALESCE-with-existing UPDATE is a no-op.
-- =====================================================================
