-- =====================================================================
-- Batch PE.5 — promote non-study-write pipeline output into studies + AGE
--
-- D-PE7 (ratified 2026-05-19): "all research output in the graph". The
-- existing work_item_promote_to_study is hardcoded to study-write*; non-
-- study-write pipelines that auto-materialized to disk never got a
-- studies row, so they never made it into the AGE graph either. This
-- step closes that gap for the four PE-A pipelines + research-write.
--
-- Approach: new stewards.promote_to_study() function that knows how to
-- map each pipeline's last-stage output into a studies row, calling
-- import_study() to ride the existing AGE indexing (MERGE :Study vertex
-- + CITES edges to Scripture/Talk/Manual/Reference per parse_gospel_links).
--
-- Wiring: on_maturity_verified calls promote_to_study() for the four
-- non-study-write families inside the same block where it enqueues the
-- file write. Sabbath gate respected — sabbath-required pipelines that
-- haven't sabbathed yet are skipped with a NOTICE rather than blocking
-- the trigger.
--
-- Backfill: the 15 existing completed research-write rows get promoted
-- inline at the bottom of this file. 14 sabbathed (will succeed), 1 not
-- (will be skipped with NOTICE; can be backfilled after manual sabbath).
-- =====================================================================

-- ---------------------------------------------------------------------
-- PE.5.A — stewards.promote_to_study(work_item_id)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.promote_to_study(p_work_item_id uuid)
RETURNS text
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_pipeline    stewards.pipelines%ROWTYPE;
    v_review_text text;
    v_slug        text;
    v_title       text;
    v_file_path   text;
    v_kind        text;
    v_frontmatter jsonb;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF NOT FOUND THEN
        RAISE NOTICE 'promote_to_study: work_item % not found', p_work_item_id;
        RETURN NULL;
    END IF;

    -- Only handle the four PE-A non-study-write families. study-write
    -- continues to flow through work_item_promote_to_study (separate path
    -- preserved to avoid touching working behavior).
    IF v_wi.pipeline_family NOT IN (
        'research-write', 'research-summary',
        'yt-gospel-evaluate', 'yt-secular-digest'
    ) THEN
        RETURN NULL;
    END IF;

    IF v_wi.status <> 'completed' THEN
        RETURN NULL;
    END IF;

    -- Sabbath gate: skip with NOTICE rather than raising. Raising inside
    -- on_maturity_verified would abort the maturity transition; a soft
    -- skip lets file materialization still happen and the row can be
    -- backfilled after manual sabbath.
    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;
    IF v_pipeline.sabbath_enabled AND v_wi.sabbath_completed_at IS NULL THEN
        RAISE NOTICE 'promote_to_study: skipping work_item % — sabbath required for % but not yet completed',
            p_work_item_id, v_wi.pipeline_family;
        RETURN NULL;
    END IF;

    v_review_text := v_wi.stage_results -> 'review' ->> 'output';
    IF v_review_text IS NULL OR length(v_review_text) < 100 THEN
        RAISE NOTICE 'promote_to_study: skipping work_item % — review output missing or too short', p_work_item_id;
        RETURN NULL;
    END IF;

    v_slug := coalesce(v_wi.slug, p_work_item_id::text);

    v_title := v_wi.input ->> 'binding_question';
    IF v_title IS NULL OR length(v_title) = 0 THEN
        v_title := v_slug;
    END IF;

    -- file_path: prefer the already-rendered v_wi.file_destination (set
    -- by on_maturity_verified earlier in the trigger). Fall back to
    -- pipeline.file_destination_template rendered ourselves if needed.
    v_file_path := v_wi.file_destination;
    IF v_file_path IS NULL THEN
        v_file_path := stewards.render_file_destination(p_work_item_id);
    END IF;

    -- Kind mapping per pipeline. These values join existing studies.kind
    -- ('study', 'proposal', 'journal', 'doc', 'phase-doc') as new
    -- substrate-output kinds.
    v_kind := CASE v_wi.pipeline_family
        WHEN 'research-write'      THEN 'research'
        WHEN 'research-summary'    THEN 'daily-digest'
        WHEN 'yt-gospel-evaluate'  THEN 'gospel-evaluation'
        WHEN 'yt-secular-digest'   THEN 'yt-digest'
    END;

    v_frontmatter := jsonb_build_object(
        'pipeline',             v_wi.pipeline_family,
        'work_item_id',         v_wi.id::text,
        'completed_at',         v_wi.completed_at,
        'sabbath_completed_at', v_wi.sabbath_completed_at,
        'tokens_in',            v_wi.tokens_in,
        'tokens_out',           v_wi.tokens_out
    );

    -- import_study handles INSERT INTO studies + AGE node MERGE +
    -- CITES edges via parse_gospel_links. One function call delivers
    -- both the table row and the graph node.
    PERFORM stewards.import_study(
        v_slug,
        v_file_path,
        v_title,
        v_review_text,
        v_frontmatter,
        v_kind
    );

    RAISE NOTICE 'promote_to_study: promoted work_item % as %/% to studies + AGE',
        p_work_item_id, v_kind, v_slug;

    RETURN v_slug;
END;
$func$;

COMMENT ON FUNCTION stewards.promote_to_study(uuid) IS
'PE.5: promote non-study-write pipeline output (research-write, research-summary, yt-gospel-evaluate, yt-secular-digest) into stewards.studies via import_study, which writes both the table row and the AGE graph Study vertex + CITES edges. Sabbath-required pipelines skipped with NOTICE if not yet sabbathed. Returns slug on success, NULL on skip.';

-- ---------------------------------------------------------------------
-- PE.5.B — wire promote_to_study into on_maturity_verified
--
-- Verbatim copy of the live trigger body with one additional call inside
-- the auto-materialize block, right after enqueue_work_item_file. The
-- promotion only fires when both auto_materialize is enabled AND the
-- pipeline is one of the four PE-A families.
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
    v_agent_ok      boolean;
    v_spawn_n       int;
    v_promoted      text;
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

    IF NEW.pipeline_family = 'agent-proposal' AND NEW.agent_proposal_applied_at IS NULL THEN
        BEGIN
            v_agent_ok := stewards.apply_agent_proposal(NEW.id);
            IF v_agent_ok THEN
                SELECT file_destination INTO NEW.file_destination
                  FROM stewards.work_items WHERE id = NEW.id;
            ELSE
                RAISE NOTICE 'on_maturity_verified: apply_agent_proposal returned false for work_item=%; skipping file enqueue',
                    NEW.id;
                RETURN NEW;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: apply_agent_proposal raised: %', SQLERRM;
            RETURN NEW;
        END;
    END IF;

    IF NEW.pipeline_family = 'decompose-fanout' THEN
        BEGIN
            v_spawn_n := stewards.spawn_children(NEW.id);
            RAISE NOTICE 'on_maturity_verified: spawn_children parent=% spawned=%',
                NEW.id, v_spawn_n;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: spawn_children failed: %', SQLERRM;
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

        -- PE.5: promote non-study-write pipeline output into studies +
        -- AGE in the same auto-materialize block. promote_to_study
        -- short-circuits for non-PE-A pipelines (returns NULL), so it's
        -- safe to call here for every auto-materializing work_item.
        BEGIN
            v_promoted := stewards.promote_to_study(NEW.id);
            IF v_promoted IS NOT NULL THEN
                RAISE NOTICE 'on_maturity_verified: promote_to_study slug=% for work_item=%',
                    v_promoted, NEW.id;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: promote_to_study failed: %', SQLERRM;
        END;
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

    -- J.2 + j7: child of a fan-out verified -> use helper to check siblings
    -- and dispatch aggregator if all terminal.
    IF NEW.parent_work_item_id IS NOT NULL
       AND NEW.pipeline_family <> 'aggregate-children' THEN
        BEGIN
            PERFORM stewards.check_and_dispatch_fanout_aggregator(NEW.parent_work_item_id);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: aggregator-dispatch-check failed: %', SQLERRM;
        END;
    END IF;

    RETURN NEW;
END;
$func$;

COMMENT ON FUNCTION stewards.on_maturity_verified() IS
'PE.5 update (2026-05-19): added promote_to_study() call inside auto-materialize block for non-study-write pipelines. Verbatim of prior body otherwise.';

-- ---------------------------------------------------------------------
-- PE.5.C — backfill the 15 existing completed research-write rows
--
-- promote_to_study is idempotent (import_study uses INSERT ... ON CONFLICT
-- DO UPDATE on slug). Re-running is safe; the un-sabbathed row will be
-- skipped with NOTICE.
-- ---------------------------------------------------------------------

DO $backfill$
DECLARE
    v_row       record;
    v_result    text;
    v_promoted  int := 0;
    v_skipped   int := 0;
BEGIN
    FOR v_row IN
        SELECT id, slug
          FROM stewards.work_items
         WHERE pipeline_family = 'research-write'
           AND status = 'completed'
         ORDER BY completed_at
    LOOP
        v_result := stewards.promote_to_study(v_row.id);
        IF v_result IS NOT NULL THEN
            v_promoted := v_promoted + 1;
        ELSE
            v_skipped := v_skipped + 1;
        END IF;
    END LOOP;

    RAISE NOTICE 'PE.5.C backfill: promoted=% skipped=%', v_promoted, v_skipped;
END;
$backfill$;
