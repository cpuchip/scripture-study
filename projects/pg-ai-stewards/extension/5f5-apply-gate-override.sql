-- =====================================================================
-- Phase 5f.5 (Phase E.5) — apply_gate_override
--
-- One SQL function that performs the full override flow atomically:
--   1. Insert into stewards.gate_overrides (with required justification)
--   2. Increment human_overrides counter on the relevant trust_scores cell
--      (which auto-demotes via evaluate_trust per D-E3)
--   3. Re-apply the gate decision with the new action — same path as
--      bgworker auto-fire would have taken if the LLM had returned the
--      new action originally
--
-- Returns the new maturity (or unchanged) — same shape as
-- apply_gate_decision returns.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.apply_gate_override(
    p_gate_decision_id bigint,
    p_overridden_by    text,
    p_new_action       text,
    p_justification    text
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_decision   stewards.gate_decisions%ROWTYPE;
    v_actor      jsonb;
    v_new_decision jsonb;
    v_result     text;
BEGIN
    IF p_new_action NOT IN ('advance','revise','surface') THEN
        RAISE EXCEPTION 'apply_gate_override: invalid new_action %', p_new_action;
    END IF;
    IF p_justification IS NULL OR length(trim(p_justification)) < 10 THEN
        RAISE EXCEPTION 'apply_gate_override: justification required (>= 10 chars)';
    END IF;
    IF p_overridden_by IS NULL OR length(trim(p_overridden_by)) = 0 THEN
        RAISE EXCEPTION 'apply_gate_override: overridden_by required';
    END IF;

    SELECT * INTO v_decision FROM stewards.gate_decisions WHERE id = p_gate_decision_id;
    IF v_decision.id IS NULL THEN
        RAISE EXCEPTION 'apply_gate_override: gate_decision % not found', p_gate_decision_id;
    END IF;

    IF v_decision.action = p_new_action THEN
        RAISE EXCEPTION 'apply_gate_override: original action and new_action are both %; this is a no-op', p_new_action;
    END IF;

    -- (1) record the override
    INSERT INTO stewards.gate_overrides
        (gate_decision_id, overridden_by, new_action, justification)
    VALUES
        (p_gate_decision_id, p_overridden_by, p_new_action, p_justification);

    -- (2) bump trust_scores.human_overrides for the (agent, pipeline, model)
    --     that produced the work being gated. This auto-demotes via
    --     evaluate_trust per D-E3.
    v_actor := stewards.work_item_stage_actor(v_decision.work_item_id);
    IF v_actor IS NOT NULL THEN
        BEGIN
            PERFORM stewards.trust_record_override(
                v_actor->>'agent_family',
                v_actor->>'pipeline_family',
                v_actor->>'model'
            );
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'trust_record_override raised: %', SQLERRM;
        END;
    END IF;

    -- (3) re-apply the gate decision with the new action. Compose a
    --     synthetic decision jsonb that carries through the original
    --     reasoning + feedback so the audit trail stays complete.
    v_new_decision := jsonb_build_object(
        'action',    p_new_action,
        'reasoning', '[human override by ' || p_overridden_by || '] ' ||
                     coalesce(v_decision.reasoning, ''),
        'feedback',  coalesce(v_decision.feedback, '')
    );
    v_result := stewards.apply_gate_decision(
        v_decision.work_item_id, v_new_decision, v_decision.work_id);

    RETURN v_result;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_gate_override(bigint, text, text, text) IS
'Phase 5f (E.5): atomic override of a gate decision. Writes gate_overrides row, bumps human_overrides on trust_scores (auto-demotes per D-E3), re-applies apply_gate_decision with the new action. Returns the resulting maturity. Requires justification >= 10 chars.';
