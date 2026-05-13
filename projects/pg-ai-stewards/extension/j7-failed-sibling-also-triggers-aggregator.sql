-- =====================================================================
-- Batch J.2 follow-up 4 — failed child also triggers aggregator check
-- =====================================================================
-- Bug surfaced during J.3 (science-center exhibits fanout): when a
-- child's status transitions to 'failed' (not just maturity to
-- 'verified'), the aggregator-dispatch check should still fire so the
-- chain can converge with partial results.
--
-- The original on_maturity_verified branch B (J.2) only ran on maturity
-- transitions to 'verified' because of the early-return at the top of
-- the function. So when the last child FAILED (instead of verifying),
-- the aggregator stayed pending forever.
--
-- Fix: extract the sibling-count + aggregator-dispatch logic into a
-- helper function `check_and_dispatch_fanout_aggregator(parent_id)`
-- and add a second trigger that fires on status transitions to
-- failed/cancelled for children. The maturity-verified path now also
-- calls the helper.
-- =====================================================================

-- Helper function: count unfinished siblings under a parent and
-- dispatch the aggregator if 0 remain. Idempotent — safe to call
-- multiple times; dispatch_stage will no-op if aggregator already
-- dispatched.

CREATE OR REPLACE FUNCTION stewards.check_and_dispatch_fanout_aggregator(p_parent_id uuid)
RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_unfinished int;
    v_agg_id     uuid;
    v_agg_wq     bigint;
BEGIN
    IF p_parent_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Count siblings (excluding the aggregator itself) that are
    -- neither verified nor in a terminal-failure state.
    SELECT COUNT(*) INTO v_unfinished
      FROM stewards.work_items
     WHERE parent_work_item_id = p_parent_id
       AND pipeline_family <> 'aggregate-children'
       AND maturity <> 'verified'
       AND status NOT IN ('cancelled', 'failed');

    IF v_unfinished > 0 THEN
        RETURN NULL;
    END IF;

    -- Find the aggregator sibling (status='pending').
    SELECT id INTO v_agg_id
      FROM stewards.work_items
     WHERE parent_work_item_id = p_parent_id
       AND pipeline_family = 'aggregate-children'
       AND status = 'pending'
     LIMIT 1;

    IF v_agg_id IS NULL THEN
        RETURN NULL;
    END IF;

    v_agg_wq := stewards.work_item_dispatch_stage(v_agg_id, NULL);
    RAISE NOTICE 'check_and_dispatch_fanout_aggregator: aggregator % dispatched wq=% (parent=%, all siblings terminal)',
        v_agg_id, v_agg_wq, p_parent_id;

    RETURN v_agg_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.check_and_dispatch_fanout_aggregator(uuid) IS
'Batch J.2 follow-up 4: idempotent helper that counts unfinished children under a parent and dispatches the aggregator if all siblings are terminal (verified, failed, or cancelled). Called from both on_maturity_verified (child verifies) and the new on_child_status_terminal trigger (child fails).';

-- Replace the on_maturity_verified body's branch B to call the helper
-- instead of inlining the count/dispatch logic.

CREATE OR REPLACE FUNCTION stewards.on_maturity_verified()
RETURNS trigger LANGUAGE plpgsql AS $FN$
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
$FN$;

-- New trigger: fires when a child's status transitions to failed/cancelled.
-- Calls the same helper so the aggregator dispatches even when the last
-- sibling fails instead of verifying.

CREATE OR REPLACE FUNCTION stewards.on_child_status_terminal()
RETURNS trigger LANGUAGE plpgsql AS $FN$
BEGIN
    -- Only fires for children of a fan-out (parent_work_item_id IS NOT NULL).
    -- The trigger WHEN clause already filters; defense-in-depth check.
    IF NEW.parent_work_item_id IS NULL
       OR NEW.pipeline_family = 'aggregate-children' THEN
        RETURN NEW;
    END IF;

    BEGIN
        PERFORM stewards.check_and_dispatch_fanout_aggregator(NEW.parent_work_item_id);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'on_child_status_terminal: aggregator-dispatch-check failed: %', SQLERRM;
    END;

    RETURN NEW;
END;
$FN$;

COMMENT ON FUNCTION stewards.on_child_status_terminal() IS
'Batch J.2 follow-up 4: fires when a fanout child transitions to a terminal status (failed/cancelled). Calls check_and_dispatch_fanout_aggregator so the chain converges even when children fail rather than verify.';

DROP TRIGGER IF EXISTS work_items_on_child_status_terminal ON stewards.work_items;

CREATE TRIGGER work_items_on_child_status_terminal
AFTER UPDATE OF status ON stewards.work_items
FOR EACH ROW
WHEN (
    NEW.status IN ('failed', 'cancelled')
    AND OLD.status NOT IN ('failed', 'cancelled')
    AND NEW.parent_work_item_id IS NOT NULL
    AND NEW.pipeline_family <> 'aggregate-children'
)
EXECUTE FUNCTION stewards.on_child_status_terminal();

-- =====================================================================
-- End of j7-failed-sibling-also-triggers-aggregator.sql
-- =====================================================================
