-- =====================================================================
-- Batch J.9.c — start_brainstorm() gains p_lenses subset selector
-- =====================================================================
-- Per ratified Q5 (2026-05-29): default lens subset = existing 4
-- (scamper, six-hats, crazy8s, reverse). Backward compat — callers who
-- don't pass p_lenses get today's behavior exactly. Caller opts INTO
-- the J.9 lenses by passing an array.
--
-- p_lenses validation: each element must reference an existing
-- brainstorm-{name} pipeline. RAISE on unknown names so misspellings
-- surface immediately rather than at child dispatch time (where the
-- pipeline_family FK violation is less legible).
--
-- p_models interaction: keys in p_models that aren't in p_lenses are
-- silently ignored (forward-compat for callers who pass a wider model
-- map than the lens subset they actually want).
-- =====================================================================


-- DROP the J.8.c 7-arg signature; CREATE the 8-arg version.
DROP FUNCTION IF EXISTS stewards.start_brainstorm(text, text, text, text, text, bigint, jsonb);


CREATE OR REPLACE FUNCTION stewards.start_brainstorm(
    p_binding_question        text,
    p_destination             text,
    p_project_association     text     DEFAULT NULL,
    p_actor                   text     DEFAULT 'human',
    p_slug                    text     DEFAULT NULL,
    p_cost_cap_per_lens_micro bigint   DEFAULT 200000,
    p_models                  jsonb    DEFAULT NULL,
    p_lenses                  text[]   DEFAULT ARRAY['scamper', 'six-hats', 'crazy8s', 'reverse']
)
RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_slug             text;
    v_parent_id        uuid;
    v_manifest         jsonb;
    v_lens             text;
    v_lens_family      text;
    v_lens_slug        text;
    v_models_entry     jsonb;
    v_model_override   text;
    v_provider_override text;
    v_child            jsonb;
    v_children_arr     jsonb := '[]'::jsonb;
    v_known_count      int;
    v_unknown_lenses   text[];
BEGIN
    -- Empty p_lenses is a caller error — what would brainstorming with
    -- zero lenses even mean?
    IF p_lenses IS NULL OR cardinality(p_lenses) = 0 THEN
        RAISE EXCEPTION 'start_brainstorm: p_lenses must contain at least one lens name';
    END IF;

    -- Validate every requested lens corresponds to an existing pipeline.
    -- Surfacing unknown lens names here is much friendlier than letting
    -- the FK violation fire deep inside spawn_children().
    SELECT array_agg(lens_name)
      INTO v_unknown_lenses
      FROM (
          SELECT unnest(p_lenses) AS lens_name
      ) requested
     WHERE NOT EXISTS (
         SELECT 1 FROM stewards.pipelines
          WHERE family = 'brainstorm-' || requested.lens_name
     );

    IF v_unknown_lenses IS NOT NULL THEN
        -- Note: %% escapes a literal % for RAISE; the LIKE pattern in the
        -- introspection hint contains one, and unescaped it gets consumed
        -- as a format placeholder.
        RAISE EXCEPTION 'start_brainstorm: unknown lens name(s): %. Available lenses: %. (Introspect with SELECT regexp_replace(family, ''^brainstorm-'', '''') FROM stewards.pipelines WHERE family LIKE ''brainstorm-%%'')',
            v_unknown_lenses,
            (SELECT array_agg(regexp_replace(family, '^brainstorm-', ''))
               FROM stewards.pipelines
              WHERE family LIKE 'brainstorm-%');
    END IF;

    v_slug := COALESCE(p_slug, 'brainstorm-' || to_char(now() AT TIME ZONE 'UTC', 'YYYYMMDD-HH24MISS'));

    -- Build the children array — one entry per requested lens.
    FOREACH v_lens IN ARRAY p_lenses LOOP
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
                NULL;  -- explicit NULL → fallback chain
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
        'rationale', format('Brainstorm: %s lens(es) — %s. Synthesis aggregator combines.',
                            cardinality(p_lenses),
                            array_to_string(p_lenses, ', ')),
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
        jsonb_build_object('binding_question', p_binding_question, 'lenses', to_jsonb(p_lenses)),
        (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
        p_actor,
        p_project_association,
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', format('brainstorm: pre-populated %s-lens manifest, no context_gather LLM call', cardinality(p_lenses))),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned',
        'completed'
    )
    RETURNING id INTO v_parent_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_parent_id;

    RAISE NOTICE 'start_brainstorm: parent=% slug=% lenses=% p_models=%',
        v_parent_id, v_slug, p_lenses, COALESCE(p_models::text, 'NULL');
    RETURN v_parent_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.start_brainstorm(text, text, text, text, text, bigint, jsonb, text[]) IS
'Batch J.9.c: brainstorm entry point with lens subset + per-lens model override. p_lenses defaults to existing 4 (scamper, six-hats, crazy8s, reverse) for backward compat; caller passes a subset of the 12 available short lens names (mind-mapping, brainwriting, starbursting, disney, storyboarding, triz, forced-analogy, worst-idea, plus the 4 originals). p_models keys may overlap p_lenses; missing keys use the J.8.a fallback chain.';


-- =====================================================================
-- Acceptance (verify before commit):
--
--   1. Backward compat: start_brainstorm(q, dest) with no p_lenses
--      defaults to the original 4 and produces an identical manifest
--      shape to today.
--
--   2. Subset selection: start_brainstorm(q, dest, p_lenses :=
--      ARRAY['scamper','starbursting','disney']) → manifest has 3
--      children referencing brainstorm-scamper / -starbursting / -disney.
--
--   3. All 12: start_brainstorm(q, dest, p_lenses := ARRAY[
--      'scamper','six-hats','crazy8s','reverse','mind-mapping',
--      'brainwriting','starbursting','disney','storyboarding','triz',
--      'forced-analogy','worst-idea']) → 12 children spawned + aggregator.
--
--   4. Unknown lens raises: start_brainstorm(q, dest, p_lenses :=
--      ARRAY['scamper','typo-name']) → RAISES with helpful message
--      listing the actual available lenses.
--
--   5. p_lenses + p_models combine: start_brainstorm(q, dest,
--      p_lenses := ARRAY['mind-mapping','triz'],
--      p_models := '{"triz":"opus-4.7"}'::jsonb) →
--      mind-mapping uses fallback (default_model qwen3.6-plus),
--      triz gets model_override='opus-4.7'.
-- =====================================================================
