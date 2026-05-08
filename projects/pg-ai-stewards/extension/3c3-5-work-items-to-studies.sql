-- =====================================================================
-- Phase 3c.3.5 — auto-promote completed work_items into stewards.studies
--
-- The substrate produces studies via study-write pipelines. They sit in
-- work_items.stage_results as JSONB until something pulls them out.
-- Future Watchman passes that walk the studies graph won't see them
-- unless they exist as stewards.studies rows.
--
-- This file adds:
--   1. stewards.work_item_promote_to_study(work_item_id) — explicit
--      function that takes a completed work_item and upserts it as a
--      study via stewards.import_study(). Single code path with the
--      regular importer.
--   2. AFTER UPDATE trigger on work_items that fires the function on
--      transition to status='completed' for any pipeline matching
--      'study-write%'.
--   3. Backfill DO block for the 5 completed runs from the 2026-05-08
--      voice experiment (and any prior completions).
--
-- Slugs are namespaced 'substrate--{work_item_slug}' to prevent
-- collision with workspace studies imported from `study/`.
--
-- Idempotent: import_study() is itself an UPSERT, the trigger guards
-- on the OLD→NEW transition, and the backfill calls the same function.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Promotion function (called by trigger + backfill + manual psql)
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_promote_to_study(p_work_item_id uuid)
RETURNS text  -- the resulting slug, or NULL if the work_item wasn't promotable
LANGUAGE plpgsql AS $$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_review_text text;
    v_slug        text;
    v_title       text;
    v_frontmatter jsonb;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF NOT FOUND THEN
        RAISE NOTICE 'work_item_promote_to_study: % not found', p_work_item_id;
        RETURN NULL;
    END IF;

    -- Only promote completed study-write outputs.
    IF v_wi.status <> 'completed' OR v_wi.pipeline_family NOT LIKE 'study-write%' THEN
        RETURN NULL;
    END IF;

    -- The review stage's `output` field is the publishable body. If it's
    -- empty or trivially short, skip — early failures shouldn't pollute
    -- the studies table.
    v_review_text := v_wi.stage_results -> 'review' ->> 'output';
    IF v_review_text IS NULL OR length(v_review_text) < 100 THEN
        RETURN NULL;
    END IF;

    -- Slug: prefer the work_item's slug, fall back to the UUID. Always
    -- prefixed 'substrate--' so workspace studies remain authoritative
    -- on collision.
    v_slug := 'substrate--' || coalesce(v_wi.slug, v_wi.id::text);

    -- Title: extract the first markdown H1 from the review text.
    -- Fall back to the slug if no header found.
    v_title := substring(v_review_text from '#[ ]+([^\n]+)');
    IF v_title IS NULL OR trim(v_title) = '' THEN
        v_title := coalesce(v_wi.slug, 'Untitled substrate study');
    ELSE
        v_title := trim(v_title);
    END IF;

    -- Frontmatter records provenance + cost data so future readers
    -- (Watchman, the comparison memos, etc.) can distinguish substrate-
    -- produced studies from imported workspace ones.
    v_frontmatter := jsonb_build_object(
        'source',          'substrate',
        'pipeline_family', v_wi.pipeline_family,
        'work_item_id',    v_wi.id::text,
        'tokens_in',       v_wi.tokens_in,
        'tokens_out',      v_wi.tokens_out,
        'completed_at',    v_wi.completed_at::text,
        'actor',           v_wi.actor
    );

    PERFORM stewards.import_study(
        v_slug,
        '.substrate-produced/' || v_slug || '.md',
        v_title,
        v_review_text,
        v_frontmatter,
        'study'
    );

    RETURN v_slug;
END;
$$;

COMMENT ON FUNCTION stewards.work_item_promote_to_study(uuid) IS
  'Upserts a completed study-write work_item into stewards.studies '
  'via the standard import_study() path. Returns the resulting slug, '
  'or NULL if the work_item was not promotable (wrong status, wrong '
  'pipeline, or empty review output). Idempotent.';

-- ---------------------------------------------------------------------
-- 2. Trigger — fires on status→completed transition
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_promote_trigger()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    -- Only fire on actual transition INTO completed (not on every
    -- update of an already-completed row).
    IF NEW.status = 'completed' AND coalesce(OLD.status, '') <> 'completed' THEN
        PERFORM stewards.work_item_promote_to_study(NEW.id);
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS work_item_promote_trg ON stewards.work_items;
CREATE TRIGGER work_item_promote_trg
    AFTER UPDATE OF status ON stewards.work_items
    FOR EACH ROW
    WHEN (NEW.status = 'completed' AND NEW.pipeline_family LIKE 'study-write%')
    EXECUTE FUNCTION stewards.work_item_promote_trigger();

-- ---------------------------------------------------------------------
-- 3. Backfill — promote any work_items that completed before this file
--    was applied. Idempotent because import_study() is an upsert.
-- ---------------------------------------------------------------------
DO $backfill$
DECLARE
    v_wi RECORD;
    v_slug text;
    v_count int := 0;
BEGIN
    FOR v_wi IN
        SELECT id
          FROM stewards.work_items
         WHERE status = 'completed'
           AND pipeline_family LIKE 'study-write%'
         ORDER BY completed_at NULLS LAST
    LOOP
        v_slug := stewards.work_item_promote_to_study(v_wi.id);
        IF v_slug IS NOT NULL THEN
            v_count := v_count + 1;
            RAISE NOTICE 'backfill: promoted % -> %', v_wi.id, v_slug;
        END IF;
    END LOOP;
    RAISE NOTICE '3c.3.5 backfill: % work_items promoted', v_count;
END;
$backfill$;
