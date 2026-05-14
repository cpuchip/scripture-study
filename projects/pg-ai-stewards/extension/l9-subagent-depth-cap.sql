-- =====================================================================
-- Batch L.9 — Sub-agent depth cap (≤ 2)
-- =====================================================================
-- Prevents recursive sub-agent fanout. Depth definition:
--   0 = root (parent_work_item_id IS NULL)
--   1 = direct child of root
--   2 = grandchild — the cap
--   3 = great-grandchild — forbidden
--
-- Enforced as a BEFORE INSERT/UPDATE trigger on stewards.work_items
-- whenever parent_work_item_id is set, walking the chain and raising
-- if depth would exceed 2. Pure SQL — no bridge Go changes.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: compute depth of a candidate parent (= depth a new child
--    would have).
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.subagent_depth_of(p_work_item_id uuid)
RETURNS int LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_depth int := 0;
    v_cur   uuid := p_work_item_id;
    v_next  uuid;
    v_guard int := 0;
BEGIN
    -- p_work_item_id is the PARENT we're considering linking to.
    -- Walk parents until we hit NULL. The number of hops is the parent's depth.
    -- A new child of this parent would be at depth+1.
    WHILE v_cur IS NOT NULL LOOP
        v_guard := v_guard + 1;
        IF v_guard > 64 THEN
            RAISE EXCEPTION 'subagent_depth_of: cycle detected at %', v_cur;
        END IF;

        SELECT parent_work_item_id INTO v_next
          FROM stewards.work_items WHERE id = v_cur;

        IF v_next IS NULL THEN
            EXIT;
        END IF;

        v_depth := v_depth + 1;
        v_cur := v_next;
    END LOOP;

    RETURN v_depth;
END;
$FN$;

COMMENT ON FUNCTION stewards.subagent_depth_of(uuid) IS
'Batch L.9: returns the sub-agent depth of a work_item (0 = root, 1 = child of root, 2 = grandchild). A new child of this work_item would be at depth + 1.';


-- ---------------------------------------------------------------------
-- 2. Helper: check_subagent_depth — boolean form for explicit callers.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.check_subagent_depth(
    p_parent_work_item_id uuid,
    p_max_depth int DEFAULT 2
) RETURNS boolean LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_parent_depth int;
    v_child_depth  int;
BEGIN
    IF p_parent_work_item_id IS NULL THEN
        RETURN true;  -- spawning at root is always allowed
    END IF;

    v_parent_depth := stewards.subagent_depth_of(p_parent_work_item_id);
    v_child_depth  := v_parent_depth + 1;

    IF v_child_depth > p_max_depth THEN
        RAISE EXCEPTION 'check_subagent_depth: would exceed cap (parent depth %, new child would be %, max %)',
            v_parent_depth, v_child_depth, p_max_depth;
    END IF;

    RETURN true;
END;
$FN$;

COMMENT ON FUNCTION stewards.check_subagent_depth(uuid, int) IS
'Batch L.9: validation form. Returns true if a new child of p_parent_work_item_id would land at or below p_max_depth (default 2). Raises otherwise. Used by spawn_subagent_create and by the work_items depth-cap trigger.';


-- ---------------------------------------------------------------------
-- 3. Trigger: BEFORE INSERT/UPDATE on work_items enforces depth ≤ 2.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trigger_enforce_subagent_depth()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_parent_depth int;
BEGIN
    IF NEW.parent_work_item_id IS NULL THEN RETURN NEW; END IF;

    -- Allow updates that don't change parent linkage (skip the check).
    IF TG_OP = 'UPDATE'
       AND OLD.parent_work_item_id IS NOT DISTINCT FROM NEW.parent_work_item_id THEN
        RETURN NEW;
    END IF;

    v_parent_depth := stewards.subagent_depth_of(NEW.parent_work_item_id);

    IF v_parent_depth + 1 > 2 THEN
        RAISE EXCEPTION
            'subagent depth cap exceeded: parent % is at depth %, child would be %, max 2',
            NEW.parent_work_item_id, v_parent_depth, v_parent_depth + 1;
    END IF;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS work_items_enforce_subagent_depth ON stewards.work_items;

CREATE TRIGGER work_items_enforce_subagent_depth
BEFORE INSERT OR UPDATE OF parent_work_item_id ON stewards.work_items
FOR EACH ROW
EXECUTE FUNCTION stewards.trigger_enforce_subagent_depth();

COMMENT ON FUNCTION stewards.trigger_enforce_subagent_depth() IS
'Batch L.9: BEFORE INSERT/UPDATE OF parent_work_item_id on work_items. Walks the parent chain and raises if a child would exceed depth 2. Skips UPDATEs that do not change parent linkage. Spawning at root (NULL parent) is always allowed.';


-- =====================================================================
-- End of l9-subagent-depth-cap.sql
-- =====================================================================
