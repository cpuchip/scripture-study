-- =====================================================================
-- Phase 5d.4 (Phase C.5) — backfill work_items.intent_id + NOT NULL
--
-- Existing work_items predate Phase C and have intent_id IS NULL.
-- Backfill them with the seeded 'scripture-study' intent (from
-- intent.yaml), then enforce NOT NULL on the column going forward.
--
-- Per D-C3 ratification: every new work_item requires intent_id at
-- creation. NewWork form enforces; work_item_create() will start
-- requiring it once C.7 lands the API change.
--
-- This migration is destructive of the "implicit intent" state but
-- not of any data. Run AFTER C.3 has seeded intent.yaml at least once.
-- =====================================================================

DO $backfill$
DECLARE
    v_default_intent_id uuid;
    v_backfilled int;
BEGIN
    SELECT id INTO v_default_intent_id
      FROM stewards.intents WHERE slug = 'scripture-study';

    IF v_default_intent_id IS NULL THEN
        RAISE EXCEPTION '5d4 backfill: no scripture-study intent found. Run seed_intents_from_yaml first.';
    END IF;

    UPDATE stewards.work_items
       SET intent_id = v_default_intent_id
     WHERE intent_id IS NULL;

    GET DIAGNOSTICS v_backfilled = ROW_COUNT;
    RAISE NOTICE '5d4 backfill: % work_items assigned to scripture-study intent', v_backfilled;
END;
$backfill$;

ALTER TABLE stewards.work_items ALTER COLUMN intent_id SET NOT NULL;

COMMENT ON COLUMN stewards.work_items.intent_id IS
'Phase 5d (C.5): NOT NULL — every work_item must have an explicit intent (D-C3). Backfilled existing rows to default scripture-study intent during C.5 migration.';

-- ---------------------------------------------------------------------
-- Stewardship: existing work_item_create() doesn't set intent_id.
-- Watchman scheduler + every existing caller would now fail. Add an
-- optional p_intent_id parameter; when NULL, default to the
-- scripture-study intent. New callers (NewWork) pass an explicit id;
-- legacy callers stay working with the default.
-- ---------------------------------------------------------------------

DROP FUNCTION IF EXISTS stewards.work_item_create(text, jsonb, text, text, integer);

CREATE OR REPLACE FUNCTION stewards.work_item_create(
    p_pipeline_family text,
    p_input           jsonb DEFAULT '{}'::jsonb,
    p_slug            text DEFAULT NULL,
    p_actor           text DEFAULT 'human',
    p_token_budget    integer DEFAULT NULL,
    p_intent_id       uuid DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_first_stage text;
    v_id          uuid;
    v_intent_id   uuid := p_intent_id;
BEGIN
    SELECT stewards.pipeline_first_stage_name(p_pipeline_family)
      INTO v_first_stage;
    IF v_first_stage IS NULL THEN
        RAISE EXCEPTION
            'work_item_create: pipeline % not found or has no stages',
            p_pipeline_family;
    END IF;

    -- Default intent: scripture-study (the project-level intent).
    -- New callers should pass an explicit intent_id; this default
    -- keeps legacy callers (watchman, ad-hoc tests) working.
    IF v_intent_id IS NULL THEN
        SELECT id INTO v_intent_id
          FROM stewards.intents WHERE slug = 'scripture-study';
        IF v_intent_id IS NULL THEN
            RAISE EXCEPTION
                'work_item_create: no intent_id supplied and no scripture-study intent seeded';
        END IF;
    END IF;

    INSERT INTO stewards.work_items
        (pipeline_family, current_stage, slug, input, actor, token_budget, intent_id)
    VALUES
        (p_pipeline_family, v_first_stage, p_slug, p_input, p_actor, p_token_budget, v_intent_id)
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_create(text, jsonb, text, text, integer, uuid) IS
'Phase 3c.1 + 5d (C.5): create a new work_item. Defaults to scripture-study intent if p_intent_id NULL — keeps legacy callers working post-NOT-NULL constraint.';

