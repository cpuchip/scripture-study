-- =====================================================================
-- H.3 followup #2 — render_file_destination helper + trigger integration
--
-- The H.3.6 e2e surfaced this gap: when work_item_create is called via
-- SQL (bypassing the UI's NewWork.vue renderTemplate path), the
-- work_item's file_destination stays NULL even when the pipeline has
-- a file_destination_template. Result: auto_materialize_on_verified
-- skips the enqueue_work_item_file call, plan file doesn't land.
--
-- Two pieces:
--
--   1. stewards.render_file_destination(work_item_id) — pure render
--      function. Reads pipeline.file_destination_template and substitutes:
--        <slug>     → work_item.slug
--        <project>  → work_item.project_association (or 'misc' fallback)
--        <id>       → first 8 chars of work_item.id
--      Returns NULL if no template OR no work_item.
--
--   2. on_maturity_verified trigger gets an upstream auto-render step:
--      if auto_materialize is enabled AND file_destination is NULL AND
--      pipeline has a template → render + UPDATE NEW.file_destination
--      (and the work_items row) before the existing enqueue check.
--
-- UI flow is unaffected: NewWork.vue still pre-renders the template
-- client-side and sets file_destination at create time. This helper
-- only kicks in when file_destination is still NULL by trigger time.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.render_file_destination(p_work_item_id uuid)
RETURNS text
LANGUAGE plpgsql
STABLE
AS $func$
DECLARE
    v_wi       stewards.work_items%ROWTYPE;
    v_pipeline stewards.pipelines%ROWTYPE;
    v_tmpl     text;
    v_out      text;
    v_project  text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;
    IF v_pipeline.family IS NULL OR v_pipeline.file_destination_template IS NULL THEN
        RETURN NULL;
    END IF;

    v_tmpl    := v_pipeline.file_destination_template;
    v_project := COALESCE(NULLIF(v_wi.project_association, ''), 'misc');

    v_out := v_tmpl;
    v_out := replace(v_out, '<slug>',    COALESCE(v_wi.slug, ''));
    v_out := replace(v_out, '<project>', v_project);
    v_out := replace(v_out, '<id>',      substring(v_wi.id::text FROM 1 FOR 8));

    RETURN v_out;
END;
$func$;

COMMENT ON FUNCTION stewards.render_file_destination(uuid) IS
'H.3 followup: render the pipeline''s file_destination_template against a work_item''s slug/project/id. Returns NULL if no template. Used by on_maturity_verified to auto-render SQL-bypass work_items whose file_destination was never set by the UI.';

-- ---------------------------------------------------------------------
-- Extend on_maturity_verified: auto-render file_destination if NULL
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
    v_rendered      text;
BEGIN
    -- Only act on transition TO 'verified'.
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

    -- D-H6.3 + D-H6.4: auto-materialize on transition to verified.
    -- H.3-followup-2: if auto_materialize is enabled but file_destination
    -- is NULL and the pipeline has a template, auto-render first.
    -- Otherwise SQL-bypass work_items (created via work_item_create
    -- directly, not the UI) never get their file landed.
    v_auto_mat := COALESCE(NEW.auto_materialize_enabled, v_pipeline.auto_materialize_on_verified);
    IF v_auto_mat AND NEW.materialized_at IS NULL THEN
        -- Auto-render if no destination set yet.
        IF NEW.file_destination IS NULL AND v_pipeline.file_destination_template IS NOT NULL THEN
            BEGIN
                v_rendered := stewards.render_file_destination(NEW.id);
                IF v_rendered IS NOT NULL THEN
                    UPDATE stewards.work_items
                       SET file_destination = v_rendered
                     WHERE id = NEW.id;
                    NEW.file_destination := v_rendered;  -- so the check below sees it
                    RAISE NOTICE 'on_maturity_verified: auto-rendered file_destination=% for work_item=%',
                        v_rendered, NEW.id;
                END IF;
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'on_maturity_verified: render_file_destination failed: %', SQLERRM;
            END;
        END IF;

        -- Now enqueue if we have a destination.
        IF NEW.file_destination IS NOT NULL THEN
            BEGIN
                v_pwid := stewards.enqueue_work_item_file(NEW.id, 'auto_materialize_on_verified');
                RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file pwid=% for work_item=%',
                    v_pwid, NEW.id;
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file failed: %', SQLERRM;
            END;
        END IF;
    END IF;

    -- H.3.5: enqueue proposed work_items for planning pipeline family.
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

-- Smoke: render against the just-completed h3-6 work_item — should
-- produce 'plans/h3-6-substrate-next-three.md' from the planning
-- pipeline's template 'plans/<slug>.md'.
SELECT 'render check:' AS check_name,
       stewards.render_file_destination(
           (SELECT id FROM stewards.work_items WHERE slug='h3-6-substrate-next-three')
       ) AS rendered;
