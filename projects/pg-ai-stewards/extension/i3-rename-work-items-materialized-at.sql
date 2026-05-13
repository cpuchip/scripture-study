-- =====================================================================
-- i3 — work_items.materialized_at → file_enqueued_at (semantics drift)
--
-- Drift surfaced 2026-05-12: the column name reads as "when the file
-- was materialized to disk" but the value is actually set inside
-- enqueue_work_item_file() at QUEUE time — before the CLI materializer
-- writes anything. The matching column on stewards.pending_file_writes
-- (also named materialized_at) is set later by stewards-cli materialize-
-- writes when the file actually lands. Same name, different lifecycles.
--
-- Fix: rename the work_items column to file_enqueued_at (truthful name).
-- All function and trigger references are redefined here in one atomic
-- migration so behavior is unchanged.
--
-- UI label "✓ materialized" in WorkItemDetail.vue updated in the same
-- session (paired commit, not part of this migration).
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) Rename the column
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    RENAME COLUMN materialized_at TO file_enqueued_at;

COMMENT ON COLUMN stewards.work_items.file_enqueued_at IS
'i3 (2026-05-12, was materialized_at): timestamp when enqueue_work_item_file successfully queued a pending_file_writes row for this work_item. Set at QUEUE time, not at file-write time. NULL = not yet queued (or DB-only — file_destination IS NULL). The actual file-write timestamp lives on stewards.pending_file_writes.materialized_at.';

-- ---------------------------------------------------------------------
-- (2) Redefine enqueue_work_item_file (was setting materialized_at = now())
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.enqueue_work_item_file(
    p_work_item_id uuid,
    p_requested_by text DEFAULT 'work_item'
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi      stewards.work_items%ROWTYPE;
    v_path    text;
    v_content text;
    v_pwid    bigint;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'enqueue_work_item_file: work_item % not found', p_work_item_id;
    END IF;

    IF v_wi.file_destination IS NULL THEN
        RETURN NULL;
    END IF;

    v_path := stewards.render_file_path_template(
        v_wi.file_destination, v_wi.slug, v_wi.id);
    IF v_path IS NULL OR length(trim(v_path)) = 0 THEN
        RAISE EXCEPTION 'enqueue_work_item_file: rendered path is empty for work_item %', p_work_item_id;
    END IF;

    v_content := stewards.extract_work_item_file_content(p_work_item_id);
    IF v_content IS NULL OR length(v_content) = 0 THEN
        RAISE EXCEPTION 'enqueue_work_item_file: extracted content is empty for work_item % (file path %)',
            p_work_item_id, v_path;
    END IF;

    INSERT INTO stewards.pending_file_writes
        (requested_by, target_path, write_mode, content, source_id, source_kind)
    VALUES
        (p_requested_by, v_path, 'create', v_content,
         p_work_item_id::text, 'work_item')
    RETURNING id INTO v_pwid;

    UPDATE stewards.work_items
       SET file_enqueued_at = now()
     WHERE id = p_work_item_id;

    RETURN v_pwid;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_work_item_file(uuid, text) IS
'i3 (2026-05-12, was Batch G.4): the universal work_item file-write producer. Checks file_destination; if NULL returns NULL (no-op). Otherwise renders the path + extracts content via extract_work_item_file_content + INSERTs pending_file_writes + sets work_items.file_enqueued_at = now(). Name corrected from materialized_at: this timestamp marks ENQUEUE, not file write. Idempotent only via the file_enqueued_at guard in callers (this function does NOT check the guard — callers may re-enqueue intentionally).';

-- ---------------------------------------------------------------------
-- (3) Redefine on_maturity_verified (was reading NEW.materialized_at)
--     Re-applies the H.3-followup-2 logic with the corrected column name.
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
    IF NEW.maturity <> 'verified' OR OLD.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = NEW.pipeline_family;
    IF v_pipeline.family IS NULL THEN
        RAISE NOTICE 'on_maturity_verified: pipeline % not found', NEW.pipeline_family;
        RETURN NEW;
    END IF;

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

    v_auto_mat := COALESCE(NEW.auto_materialize_enabled, v_pipeline.auto_materialize_on_verified);
    IF v_auto_mat AND NEW.file_enqueued_at IS NULL THEN
        IF NEW.file_destination IS NULL AND v_pipeline.file_destination_template IS NOT NULL THEN
            BEGIN
                v_rendered := stewards.render_file_destination(NEW.id);
                IF v_rendered IS NOT NULL THEN
                    UPDATE stewards.work_items
                       SET file_destination = v_rendered
                     WHERE id = NEW.id;
                    NEW.file_destination := v_rendered;
                    RAISE NOTICE 'on_maturity_verified: auto-rendered file_destination=% for work_item=%',
                        v_rendered, NEW.id;
                END IF;
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'on_maturity_verified: render_file_destination failed: %', SQLERRM;
            END;
        END IF;

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

COMMENT ON FUNCTION stewards.on_maturity_verified() IS
'i3 (2026-05-12): H.1.6 + H.3-followup-2 trigger function with column rename applied. Same behavior; reads NEW.file_enqueued_at instead of NEW.materialized_at.';

-- ---------------------------------------------------------------------
-- (4) Sanity check
-- ---------------------------------------------------------------------

DO $$
DECLARE v_count int;
BEGIN
    SELECT count(*) INTO v_count
      FROM information_schema.columns
     WHERE table_schema='stewards' AND table_name='work_items'
       AND column_name='file_enqueued_at';
    IF v_count <> 1 THEN
        RAISE EXCEPTION 'i3 verify: file_enqueued_at column not found';
    END IF;

    SELECT count(*) INTO v_count
      FROM information_schema.columns
     WHERE table_schema='stewards' AND table_name='work_items'
       AND column_name='materialized_at';
    IF v_count <> 0 THEN
        RAISE EXCEPTION 'i3 verify: old materialized_at column still present';
    END IF;
END
$$;
