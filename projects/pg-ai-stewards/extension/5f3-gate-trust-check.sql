-- =====================================================================
-- Phase 5f.3 (Phase E.3) — gate behavior change: trainee surfaces advance
--
-- apply_gate_decision extended:
--   - On action='advance', look up the stage's (agent_family, model) via
--     pipeline_stage_lookup (honoring work_items.model_override)
--   - Read stewards.trust_scores. If no row OR trust_level='trainee',
--     transition status='awaiting_review' instead of auto-advancing
--   - On a real advance that lands the new maturity at 'verified',
--     call trust_record_success — that's the "successful completion"
--     signal per the sub-spec
-- =====================================================================

-- Helper: derive (agent_family, model) for a work_item's current stage
CREATE OR REPLACE FUNCTION stewards.work_item_stage_actor(
    p_work_item_id uuid
) RETURNS jsonb
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_wi    stewards.work_items%ROWTYPE;
    v_stage jsonb;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RETURN NULL;
    END IF;
    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN jsonb_build_object(
        'agent_family',    v_stage->>'agent_family',
        'pipeline_family', v_wi.pipeline_family,
        'model',           coalesce(v_wi.model_override, v_stage->>'model')
    );
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_stage_actor(uuid) IS
'Phase 5f (E.3): returns {agent_family, pipeline_family, model} for the work_item''s current stage. model honors work_items.model_override. Used by trust check + trust counter increment.';

-- Re-create apply_gate_decision with the trust gate
CREATE OR REPLACE FUNCTION stewards.apply_gate_decision(
    p_work_item_id uuid,
    p_decision     jsonb,
    p_work_id      bigint DEFAULT NULL
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi             stewards.work_items%ROWTYPE;
    v_pipeline       stewards.pipelines%ROWTYPE;
    v_action         text;
    v_reasoning      text;
    v_feedback       text;
    v_new_maturity   text;
    v_produces_mat   text;
    v_maturity_order text[] := ARRAY['raw','researched','planned','specced','executing','verified'];
    v_idx            int;
    v_new_revision   int;
    v_actor          jsonb;
    v_trust_level    text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    v_action    := p_decision->>'action';
    v_reasoning := p_decision->>'reasoning';
    v_feedback  := p_decision->>'feedback';

    IF v_action NOT IN ('advance', 'revise', 'surface') THEN
        RAISE EXCEPTION 'apply_gate_decision: invalid action %', v_action;
    END IF;

    INSERT INTO stewards.gate_decisions
        (work_item_id, from_maturity, action, reasoning, feedback,
         work_id, revision_count, raw_response)
    VALUES
        (p_work_item_id, v_wi.maturity, v_action, v_reasoning, v_feedback,
         p_work_id, v_wi.revision_count, p_decision);

    v_new_maturity := v_wi.maturity;

    IF v_action = 'advance' THEN
        -- Phase 5f (E.3): trust check. New (agent, pipeline, model)
        -- cells start at trainee; trainee surfaces every advance for
        -- human ratification. Journeyman + master proceed.
        v_actor := stewards.work_item_stage_actor(p_work_item_id);
        IF v_actor IS NOT NULL THEN
            SELECT trust_level INTO v_trust_level
              FROM stewards.trust_scores
             WHERE agent_family    = v_actor->>'agent_family'
               AND pipeline_family = v_actor->>'pipeline_family'
               AND model           = v_actor->>'model';

            IF v_trust_level IS NULL OR v_trust_level = 'trainee' THEN
                UPDATE stewards.work_items
                   SET status = 'awaiting_review',
                       updated_at = now()
                 WHERE id = p_work_item_id;
                RETURN v_wi.maturity;  -- maturity unchanged; human must ratify
            END IF;
        END IF;

        SELECT produces_maturity INTO v_produces_mat
          FROM stewards.pipeline_stage_maturity
         WHERE pipeline_family = v_wi.pipeline_family
           AND stage_name = v_wi.current_stage;

        IF v_produces_mat IS NOT NULL THEN
            v_new_maturity := v_produces_mat;
        ELSE
            v_idx := array_position(v_maturity_order, v_wi.maturity);
            IF v_idx IS NOT NULL AND v_idx < array_length(v_maturity_order, 1) THEN
                v_new_maturity := v_maturity_order[v_idx + 1];
            END IF;
        END IF;

        UPDATE stewards.work_items
           SET maturity       = v_new_maturity,
               revision_count = 0,
               updated_at     = now()
         WHERE id = p_work_item_id;

        -- Phase 5f (E.3): record successful completion when reaching verified
        IF v_new_maturity = 'verified' AND v_actor IS NOT NULL THEN
            BEGIN
                PERFORM stewards.trust_record_success(
                    v_actor->>'agent_family',
                    v_actor->>'pipeline_family',
                    v_actor->>'model'
                );
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'trust_record_success raised: %', SQLERRM;
            END;
        END IF;

        -- Phase 5e (D.5): fire sabbath when reaching verified on a
        -- sabbath-enabled pipeline (and we haven't already).
        IF v_new_maturity = 'verified' THEN
            SELECT * INTO v_pipeline FROM stewards.pipelines
             WHERE family = v_wi.pipeline_family;
            IF v_pipeline.sabbath_enabled
               AND v_wi.sabbath_completed_at IS NULL THEN
                BEGIN
                    PERFORM stewards.sabbath_dispatch(p_work_item_id);
                EXCEPTION WHEN OTHERS THEN
                    RAISE NOTICE 'sabbath_dispatch from apply_gate_decision raised: %', SQLERRM;
                END;
            END IF;
        END IF;

    ELSIF v_action = 'revise' THEN
        v_new_revision := v_wi.revision_count + 1;

        IF v_new_revision > 2 THEN
            UPDATE stewards.work_items
               SET status = 'awaiting_review',
                   revision_count = v_new_revision,
                   updated_at = now()
             WHERE id = p_work_item_id;
        ELSE
            UPDATE stewards.work_items
               SET status                 = 'failed',
                   revision_count         = v_new_revision,
                   last_failure_reason    = 'gate revise: ' || coalesce(v_feedback, '(no feedback)'),
                   last_failure_diagnosis = 'gate_revise',
                   updated_at             = now()
             WHERE id = p_work_item_id;
        END IF;

    ELSIF v_action = 'surface' THEN
        UPDATE stewards.work_items
           SET status     = 'awaiting_review',
               updated_at = now()
         WHERE id = p_work_item_id;
    END IF;

    RETURN v_new_maturity;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_gate_decision(uuid, jsonb, bigint) IS
'Phase 5a + 5e + 5f (E.3): on action=advance, checks trust_scores for the work_item''s (agent_family, pipeline_family, model). Trainee or no-row surfaces for human ratification (D-E gate behavior). On real advance to verified, fires sabbath_dispatch (D.5) AND trust_record_success (E.3).';
