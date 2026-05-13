-- =====================================================================
-- Batch J.2 follow-up 2 — spawn_children sets aggregator file_destination directly
-- =====================================================================
-- Surfaced during smoke: render_file_destination only knows <slug>,
-- <project>, <id> substitutions. My pipeline declared
-- file_destination_template='{{input.destination}}' which is foreign
-- syntax, so the literal template string ended up in file_destination.
--
-- Cleanest fix: spawn_children already has the manifest in hand at
-- aggregator-creation time. Set file_destination directly on the
-- aggregator row. Drop the bogus template from the pipeline.
--
-- Surfaced during smoke-j2-fanout on 2026-05-13.
-- =====================================================================

-- 1. Clear the bogus template from aggregate-children pipeline.
UPDATE stewards.pipelines
   SET file_destination_template = NULL,
       updated_at = now()
 WHERE family = 'aggregate-children';

-- 2. CREATE OR REPLACE spawn_children to set file_destination directly.
CREATE OR REPLACE FUNCTION stewards.spawn_children(p_parent_id uuid)
RETURNS int LANGUAGE plpgsql AS $FN$
DECLARE
    v_parent        stewards.work_items%ROWTYPE;
    v_manifest      jsonb;
    v_manifest_raw  text;
    v_child         jsonb;
    v_child_id      uuid;
    v_count         int := 0;
    v_aggregator    jsonb;
    v_agg_id        uuid;
    v_agg_dest      text;
    v_children_arr  jsonb := '[]'::jsonb;
    v_child_pipeline text;
    v_child_slug    text;
    v_child_input   jsonb;
    v_cost_cap      bigint;
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

    -- Spawn regular children.
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

        PERFORM stewards.work_item_dispatch_stage(v_child_id, NULL);

        v_children_arr := v_children_arr || jsonb_build_object(
            'id', v_child_id::text,
            'slug', v_child_slug,
            'binding_question', v_child ->> 'binding_question',
            'pipeline_family', v_child_pipeline
        );
        v_count := v_count + 1;
    END LOOP;

    -- Spawn the aggregator (NOT dispatched — waits for siblings).
    v_aggregator := v_manifest -> 'aggregate';
    v_agg_dest   := v_aggregator ->> 'destination';

    v_agg_id := stewards.work_item_create(
        p_pipeline_family => 'aggregate-children',
        p_input           => jsonb_build_object(
            'binding_question', 'Aggregate index for: ' || COALESCE(v_parent.input ->> 'binding_question', v_parent.slug),
            'parent_work_item_id', p_parent_id::text,
            'destination', v_agg_dest,
            'synthesis', COALESCE((v_aggregator ->> 'synthesis')::boolean, false),
            'children', v_children_arr
        ),
        p_slug            => COALESCE(v_parent.slug, p_parent_id::text) || '-aggregator',
        p_actor           => v_parent.actor,
        p_intent_id       => v_parent.intent_id
    );

    -- J.2 fix: set file_destination directly from manifest (no template render).
    UPDATE stewards.work_items
       SET parent_work_item_id = p_parent_id,
           project_association = v_parent.project_association,
           file_destination    = v_agg_dest
     WHERE id = v_agg_id;

    RAISE NOTICE 'spawn_children: parent=% spawned % children + aggregator % (dest=%)',
        p_parent_id, v_count, v_agg_id, v_agg_dest;

    RETURN v_count;
END;
$FN$;

COMMENT ON FUNCTION stewards.spawn_children(uuid) IS
'Batch J.2 + j3 fix: decompose-fanout spawn mechanism. Reads stage_results.decompose.output manifest; inserts N children (each dispatched immediately) + 1 aggregator child (status=pending, file_destination=manifest.aggregate.destination). Returns count of regular children.';

-- =====================================================================
-- End of j3-spawn-sets-aggregator-destination.sql
-- =====================================================================
