-- Smoke H.1.6.1: maturity advance hook in work_item_advance
-- Synthetic stage advances; verifies forward-only ladder discipline.

DO $$
DECLARE
    v_wi        uuid;
    v_intent_id uuid;
    v_mat       text;
BEGIN
    SELECT id INTO v_intent_id FROM stewards.intents WHERE slug = 'general-research';

    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"H.1.6.1 maturity hook smoke"}'::jsonb,
        'h161-smoke', 'human', NULL, v_intent_id
    ) INTO v_wi;

    SELECT maturity INTO v_mat FROM stewards.work_items WHERE id = v_wi;
    RAISE NOTICE '[start] maturity = % (expect raw)', v_mat;

    -- Advance through gather (produces researched)
    PERFORM stewards.work_item_advance(v_wi, '{"output":"synthetic gather"}'::jsonb);
    SELECT maturity INTO v_mat FROM stewards.work_items WHERE id = v_wi;
    RAISE NOTICE '[after gather] maturity = % (expect researched)', v_mat;
    IF v_mat <> 'researched' THEN RAISE EXCEPTION 'expected researched, got %', v_mat; END IF;

    -- Advance through synthesize (produces planned)
    PERFORM stewards.work_item_advance(v_wi, '{"output":"synthetic synthesize"}'::jsonb);
    SELECT maturity INTO v_mat FROM stewards.work_items WHERE id = v_wi;
    RAISE NOTICE '[after synthesize] maturity = % (expect planned)', v_mat;
    IF v_mat <> 'planned' THEN RAISE EXCEPTION 'expected planned, got %', v_mat; END IF;

    -- Advance through review (produces verified) — terminal
    PERFORM stewards.work_item_advance(v_wi, '{"output":"synthetic review"}'::jsonb);
    SELECT maturity INTO v_mat FROM stewards.work_items WHERE id = v_wi;
    RAISE NOTICE '[after review] maturity = % (expect verified)', v_mat;
    IF v_mat <> 'verified' THEN RAISE EXCEPTION 'expected verified, got %', v_mat; END IF;

    -- Forward-only test: manually reset current_stage + status, advance gather again
    -- (simulating steward retry). Maturity should NOT regress.
    UPDATE stewards.work_items
       SET current_stage = 'gather', status = 'pending'
     WHERE id = v_wi;
    PERFORM stewards.work_item_advance(v_wi, '{"output":"synthetic gather rerun"}'::jsonb);
    SELECT maturity INTO v_mat FROM stewards.work_items WHERE id = v_wi;
    RAISE NOTICE '[after gather rerun] maturity = % (expect verified — forward-only)', v_mat;
    IF v_mat <> 'verified' THEN RAISE EXCEPTION 'forward-only violated: regressed to %', v_mat; END IF;

    -- Cleanup
    DELETE FROM stewards.work_items WHERE id = v_wi;
    RAISE NOTICE 'H.1.6.1 smoke PASSED';
END
$$;
