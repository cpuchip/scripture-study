-- =====================================================================
-- Batch H.3.5 — enqueue_proposed_work_items function + trigger extension
--
-- When a planning work_item reaches maturity='verified' (review_plan
-- passed), the on_maturity_verified trigger now ALSO reads
-- stage_results.propose_work.output as a JSON array and inserts each
-- proposed work_item into stewards.work_items.
--
-- Each new work_item:
--   - origin = 'agent_planning' (set by the function)
--   - parent_work_item_id = the planning work_item's id
--   - project_association = inherits from parent unless element
--                           overrides via "project_association" key
--   - maturity = 'raw' (default)
--   - status = 'pending' (default)
--   - current_stage = first stage of pipeline_family_hint (or
--                     'pending-review' if no valid hint — these
--                     await human ratification before dispatch)
--   - intent_id = inherits from parent (planning-partner) so the
--                 substrate doesn't lose covenant context. Future
--                 work_item_create override path would let the user
--                 swap to a domain-appropriate intent.
--
-- Q-H3.1 ratification: the propose_work stage emits strict JSON;
-- review_plan validates it before passing. By the time this function
-- runs, JSON should be valid. But we defensively try/catch and skip
-- malformed elements rather than crash the trigger.
--
-- Schema validation per element (from h3-4 review_plan template):
--   required: slug, binding_question, pipeline_family_hint (or null),
--             rationale
--   optional: project_association, destination_maturity
--
-- Skipped elements log RAISE NOTICE with the reason; not raised as
-- exception because the trigger must remain non-throwing (D-H6.5
-- spirit — substrate machinery returns NEW even on partial failure).
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.enqueue_proposed_work_items(p_work_item_id uuid)
RETURNS int
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline        stewards.pipelines%ROWTYPE;
    v_raw_output      text;
    v_clean_output    text;
    v_json            jsonb;
    v_item            jsonb;
    v_slug            text;
    v_binding         text;
    v_rationale       text;
    v_hint            text;
    v_project         text;
    v_dest_maturity   text;
    v_target_pipeline text;
    v_first_stage     text;
    v_inserted        int := 0;
    v_skipped         int := 0;
    v_reason          text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE NOTICE 'enqueue_proposed_work_items: work_item % not found', p_work_item_id;
        RETURN 0;
    END IF;

    -- Only planning-family work_items emit proposed work. Other
    -- pipelines may have propose_work output someday too, but for
    -- now this function is a no-op outside planning.
    IF v_wi.pipeline_family <> 'planning' THEN
        RETURN 0;
    END IF;

    -- Pull the raw propose_work.output. Stored as a JSONB string
    -- inside stage_results, so the #>>'{}' coercion gives us the
    -- string content (with quotes stripped + escapes resolved).
    v_raw_output := (v_wi.stage_results -> 'propose_work' -> 'output') #>> '{}';
    IF v_raw_output IS NULL OR length(trim(v_raw_output)) = 0 THEN
        RAISE NOTICE 'enqueue_proposed_work_items: empty propose_work.output for work_item %', p_work_item_id;
        RETURN 0;
    END IF;

    -- Strip optional markdown code fences in case the model wrapped
    -- the JSON despite instructions not to. ^```json\n ... \n``` $
    -- and bare ``` ... ``` patterns both handled.
    v_clean_output := regexp_replace(
        v_raw_output,
        E'^\\s*```(?:json)?\\s*\\n?|\\n?```\\s*$',
        '',
        'g'
    );
    v_clean_output := trim(v_clean_output);

    -- Parse. On parse failure, NOTICE + return 0.
    BEGIN
        v_json := v_clean_output::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'enqueue_proposed_work_items: JSON parse failed for work_item %: %', p_work_item_id, SQLERRM;
        RETURN 0;
    END;

    -- Must be an array.
    IF jsonb_typeof(v_json) <> 'array' THEN
        RAISE NOTICE 'enqueue_proposed_work_items: top-level JSON is %, expected array (work_item %)',
            jsonb_typeof(v_json), p_work_item_id;
        RETURN 0;
    END IF;

    -- Iterate. Each element gets validated, then inserted if valid.
    FOR v_item IN SELECT * FROM jsonb_array_elements(v_json)
    LOOP
        v_reason := NULL;

        -- Required fields
        v_slug      := v_item ->> 'slug';
        v_binding   := v_item ->> 'binding_question';
        v_rationale := v_item ->> 'rationale';
        v_hint      := v_item ->> 'pipeline_family_hint';  -- may be null/text

        IF v_slug IS NULL OR v_slug !~ '^[a-z0-9-]+$' THEN
            v_reason := format('invalid slug: %s', COALESCE(v_slug, '(null)'));
        ELSIF v_binding IS NULL OR length(trim(v_binding)) < 20 THEN
            -- 20 chars filters obvious garbage like "Too short." while
            -- allowing reasonable concise questions. A real planning-
            -- proposed binding is usually ~40-100 chars.
            v_reason := format('binding_question too short or missing for slug=%s (need ≥20 chars)', v_slug);
        ELSIF v_rationale IS NULL OR length(trim(v_rationale)) < 10 THEN
            v_reason := format('rationale missing or too short for slug=%s (need ≥10 chars)', v_slug);
        END IF;

        -- Optional fields
        v_project       := COALESCE(v_item ->> 'project_association', v_wi.project_association);
        v_dest_maturity := v_item ->> 'destination_maturity';

        -- Resolve pipeline_family_hint to an actual pipeline if valid.
        -- Invalid hints get logged; the work_item is still proposed but
        -- with NULL pipeline_family so the user picks at ratification.
        v_target_pipeline := NULL;
        v_first_stage := NULL;
        IF v_hint IS NOT NULL AND v_hint <> '' AND v_hint <> 'null' THEN
            IF EXISTS (SELECT 1 FROM stewards.pipelines WHERE family = v_hint) THEN
                v_target_pipeline := v_hint;
                v_first_stage := stewards.pipeline_first_stage_name(v_hint);
            ELSE
                RAISE NOTICE 'enqueue_proposed_work_items: unknown pipeline_family_hint=% for slug=%; inserting as proposal-only',
                    v_hint, v_slug;
            END IF;
        END IF;

        -- Skip if validation failed.
        IF v_reason IS NOT NULL THEN
            RAISE NOTICE 'enqueue_proposed_work_items: skipping element: %', v_reason;
            v_skipped := v_skipped + 1;
            CONTINUE;
        END IF;

        -- Slug collision? Skip with notice — the user can rename in UI.
        IF EXISTS (SELECT 1 FROM stewards.work_items WHERE slug = v_slug) THEN
            RAISE NOTICE 'enqueue_proposed_work_items: slug=% already exists, skipping', v_slug;
            v_skipped := v_skipped + 1;
            CONTINUE;
        END IF;

        -- All resolved. Insert.
        -- If we couldn't resolve a target pipeline_family, the work_item
        -- can't actually run yet — pipeline_family is NOT NULL on the
        -- work_items table. Park such proposals under the planning
        -- pipeline_family itself with current_stage='__proposal_only'
        -- so they show in the UI but cannot dispatch until reassigned.
        IF v_target_pipeline IS NULL THEN
            v_target_pipeline := 'planning';
            v_first_stage     := '__proposal_only';
        END IF;

        INSERT INTO stewards.work_items (
            slug,
            pipeline_family,
            current_stage,
            input,
            actor,
            intent_id,
            origin,
            parent_work_item_id,
            project_association,
            destination_maturity
        )
        VALUES (
            v_slug,
            v_target_pipeline,
            v_first_stage,
            jsonb_build_object(
                'binding_question', v_binding,
                'rationale_from_planning', v_rationale,
                'proposed_by_work_item_id', v_wi.id::text,
                'proposed_by_slug', v_wi.slug
            ),
            'agent',
            v_wi.intent_id,   -- inherit; user can swap at ratification
            'agent_planning',
            v_wi.id,
            v_project,
            v_dest_maturity
        );

        v_inserted := v_inserted + 1;
    END LOOP;

    RAISE NOTICE 'enqueue_proposed_work_items: work_item=% inserted=% skipped=%',
        p_work_item_id, v_inserted, v_skipped;
    RETURN v_inserted;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_proposed_work_items(uuid) IS
'H.3.5 (Batch H): reads a planning work_item''s stage_results.propose_work.output JSON array and inserts each proposed work_item with origin=agent_planning, parent_work_item_id pointing back, and intent inherited. Schema validation per Q-H3.1 ratification; malformed elements are skipped with NOTICE (not raised) so the trigger remains non-throwing.';

-- ---------------------------------------------------------------------
-- Extend on_maturity_verified to call enqueue_proposed_work_items
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.on_maturity_verified()
RETURNS trigger
LANGUAGE plpgsql
AS $func$
DECLARE
    v_pipeline      stewards.pipelines%ROWTYPE;
    v_sabbath       boolean;
    v_auto_mat      boolean;
    v_pwid          bigint;
    v_dispatch_id   bigint;
    v_proposed_n    int;
BEGIN
    -- Only act on transition TO 'verified'. NULL/non-verified previous
    -- values both fire if the new value is 'verified' and not already.
    IF NEW.maturity <> 'verified' OR OLD.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = NEW.pipeline_family;
    IF v_pipeline.family IS NULL THEN
        RAISE NOTICE 'on_maturity_verified: pipeline % not found', NEW.pipeline_family;
        RETURN NEW;
    END IF;

    -- D-H6.2: sabbath_dispatch on transition to verified
    v_sabbath := COALESCE(NEW.sabbath_enabled, v_pipeline.sabbath_enabled);
    IF v_sabbath AND NEW.sabbath_completed_at IS NULL THEN
        BEGIN
            v_dispatch_id := stewards.sabbath_dispatch(NEW.id);
            RAISE NOTICE 'on_maturity_verified: sabbath_dispatch work_id=% for work_item=%',
                v_dispatch_id, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: sabbath_dispatch failed: %', SQLERRM;
        END;
    END IF;

    -- D-H6.3 + D-H6.4: auto-materialize on transition to verified
    v_auto_mat := COALESCE(NEW.auto_materialize_enabled, v_pipeline.auto_materialize_on_verified);
    IF v_auto_mat
       AND NEW.file_destination IS NOT NULL
       AND NEW.materialized_at IS NULL
    THEN
        BEGIN
            v_pwid := stewards.enqueue_work_item_file(NEW.id, 'auto_materialize_on_verified');
            RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file pwid=% for work_item=%',
                v_pwid, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file failed: %', SQLERRM;
        END;
    END IF;

    -- H.3.5: enqueue proposed work_items for planning pipeline family.
    -- The function is no-op for non-planning pipelines.
    IF NEW.pipeline_family = 'planning' THEN
        BEGIN
            v_proposed_n := stewards.enqueue_proposed_work_items(NEW.id);
            RAISE NOTICE 'on_maturity_verified: enqueue_proposed_work_items inserted=% for work_item=%',
                v_proposed_n, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: enqueue_proposed_work_items failed: %', SQLERRM;
        END;
    END IF;

    RETURN NEW;
END;
$func$;

-- Re-attach trigger (DROP + CREATE pattern stays — the CREATE OR
-- REPLACE FUNCTION already updated the function body; trigger
-- definition itself doesn't change).
SELECT 'trigger present:' AS check_name, count(*)
  FROM pg_trigger WHERE tgname = 'work_items_on_maturity_verified';
