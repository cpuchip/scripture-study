-- =====================================================================
-- Batch R.4 — start_panel_redline() fan-out entry point
-- =====================================================================
-- The generative analog of start_brainstorm: give a panel of N models the
-- SAME document + a redline mandate, collect N proposal reports. Reuses the
-- J.2 decompose-fanout machinery (spawn_children + on_maturity_verified +
-- aggregator) — start_panel_redline just builds the manifest and creates the
-- parent; the trigger spawns + dispatches the children.
--
-- Document injection (R.2): read server-side via read_workspace_doc and
-- embedded in each child's binding_question — the panel never touches fs.
-- Tools-off + 32k (R.3): passed per child via input_extra, which spawn_children
-- merges into the child input, which work_item_dispatch_stage honors.
-- Per-model provider + cost cap: looked up from model_pricing so a gemini model
-- gets google_gemini and a cap that funds (doc input + 32k output).
--
-- Aggregate = index only (synthesis=false): the orchestrating Claude session
-- condenses by reading the children (D-RL3 default). R.5 adds the optional
-- substrate-side ranked-merge condense.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.start_panel_redline(
    p_document             text,           -- repo-relative path or single-dir glob
    p_mandate              text,           -- the kind of edits wanted
    p_models               text[],         -- the panel (model names)
    p_destination          text   DEFAULT NULL,
    p_actor                text   DEFAULT 'michael',
    p_slug                 text   DEFAULT NULL,
    p_max_tokens           int    DEFAULT 32000,
    p_cost_cap_per_model_micro bigint DEFAULT NULL,
    p_project_association  text   DEFAULT NULL
) RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_slug          text;
    v_parent_id     uuid;
    v_doc           text;
    v_doc_files     int;
    v_binding       text;
    v_model         text;
    v_provider      text;
    v_in_rate       bigint;
    v_out_rate      bigint;
    v_est_in        bigint;
    v_cap           bigint;
    v_child         jsonb;
    v_children_arr  jsonb := '[]'::jsonb;
    v_manifest      jsonb;
    v_destination   text;
BEGIN
    IF p_models IS NULL OR cardinality(p_models) = 0 THEN
        RAISE EXCEPTION 'start_panel_redline: p_models must contain at least one model';
    END IF;
    IF p_mandate IS NULL OR p_mandate = '' THEN
        RAISE EXCEPTION 'start_panel_redline: p_mandate is required (what edits do you want?)';
    END IF;

    -- Read + concatenate the document(s) server-side (R.2). read_workspace_doc
    -- enforces the doc-only / no-secret / no-traversal gate and RAISEs on a
    -- bad path; an empty result means the glob matched nothing.
    SELECT string_agg(format(E'\n\n===== %s =====\n%s', rel_path, content), '' ORDER BY rel_path),
           count(*)
      INTO v_doc, v_doc_files
      FROM stewards.read_workspace_doc(p_document);
    IF v_doc IS NULL OR v_doc_files = 0 THEN
        RAISE EXCEPTION 'start_panel_redline: document % matched no files', p_document;
    END IF;

    v_slug        := COALESCE(p_slug, 'redline-' || to_char(now() AT TIME ZONE 'UTC', 'YYYYMMDD-HH24MISS'));
    v_destination := COALESCE(p_destination, 'study/.scratch/' || v_slug || '-index.md');

    v_binding := p_mandate
              || E'\n\n# DOCUMENT TO REDLINE (' || p_document || ', ' || v_doc_files || ' file(s))'
              || v_doc;

    -- One child per panel model.
    FOREACH v_model IN ARRAY p_models LOOP
        -- Resolve provider + rates from model_pricing (latest row for the model).
        SELECT provider, input_micro_per_mtok, output_micro_per_mtok
          INTO v_provider, v_in_rate, v_out_rate
          FROM (
              SELECT DISTINCT ON (provider, model) provider, model,
                     input_micro_per_mtok, output_micro_per_mtok
                FROM stewards.model_pricing
               WHERE model = v_model
               ORDER BY provider, model, effective_at DESC
          ) p
         LIMIT 1;

        -- Cost cap: fund (doc input + max_tokens output) with 2x headroom,
        -- floor $0.30. Free models (rate 0) land on the floor (harmless).
        v_est_in := (length(v_binding) / 4)::bigint;   -- ~4 chars/token
        v_cap := COALESCE(
            p_cost_cap_per_model_micro,
            GREATEST(
                ( ceil( v_est_in::numeric / 1000000 * COALESCE(v_in_rate, 1000000)
                      + p_max_tokens::numeric / 1000000 * COALESCE(v_out_rate, 5000000) )::bigint ) * 2,
                300000
            )
        );

        v_child := jsonb_build_object(
            'slug',             v_slug || '-' || regexp_replace(lower(v_model), '[^a-z0-9]+', '-', 'g'),
            'pipeline_family',  'redline',
            'binding_question', v_binding,
            'model_override',   v_model,
            'cost_cap_micro',   v_cap,
            'input_extra',      jsonb_build_object(
                                    'tools_disabled', true,
                                    'max_tokens',     p_max_tokens::text
                                )
        );
        IF v_provider IS NOT NULL THEN
            v_child := v_child || jsonb_build_object('provider_override', v_provider);
        END IF;

        v_children_arr := v_children_arr || v_child;
    END LOOP;

    v_manifest := jsonb_build_object(
        'rationale', format('Panel redline: %s model(s) each redline %s. Aggregator indexes; orchestrator condenses (or R.5 substrate condense).',
                            cardinality(p_models), p_document),
        'children', v_children_arr,
        'aggregate', jsonb_build_object(
            'destination', v_destination,
            'synthesis',   false
        )
    );

    -- Create the decompose-fanout parent with the pre-built manifest (no LLM
    -- decompose call), then flip to verified to fire spawn_children — exactly
    -- the start_brainstorm pattern (j9c).
    INSERT INTO stewards.work_items (
        pipeline_family, current_stage, slug, input, intent_id, actor,
        project_association, stage_results, maturity, status
    ) VALUES (
        'decompose-fanout', 'decompose', v_slug,
        jsonb_build_object('binding_question',
            format('Panel redline of %s across %s models', p_document, cardinality(p_models))),
        (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
        p_actor, p_project_association,
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', 'panel-redline: pre-built manifest, no context_gather LLM call'),
            'decompose',      jsonb_build_object('output', v_manifest)
        ),
        'planned', 'completed'
    )
    RETURNING id INTO v_parent_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_parent_id;

    RAISE NOTICE 'start_panel_redline: parent=% slug=% models=% doc=% (% files, % chars)',
        v_parent_id, v_slug, p_models, p_document, v_doc_files, length(v_doc);
    RETURN v_parent_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.start_panel_redline(text, text, text[], text, text, text, int, bigint, text) IS
'R.4: fan a document-redline mandate across a panel of models. Reads the doc server-side (read_workspace_doc), injects it into each child binding_question, and builds a decompose-fanout manifest with one redline child per model (model_override + provider from model_pricing + input_extra{tools_disabled,max_tokens} + auto-scaled cost_cap). Reuses spawn_children/on_maturity_verified. Returns the parent work_item id; children + an index aggregator spawn via the trigger.';


-- =====================================================================
-- Acceptance (R.4, transactional — ROLLBACK, no spend):
--   1. start_panel_redline('projects/scripture-book/src/chapters/00_frontmatter.md',
--        'Tighten prose; propose surgical edits.',
--        ARRAY['deepseek-v4-flash','mimo-v2.5']) → parent id;
--      spawns 2 redline children with model_override set, input.tools_disabled=true,
--      input.max_tokens='32000', binding_question containing the chapter text,
--      provider_override='opencode_go', plus 1 aggregate-children child.
--   2. A gemini model in the panel gets provider_override='google_gemini'.
--   3. Bad document path → EXCEPTION (read_workspace_doc gate).
-- =====================================================================
