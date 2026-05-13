-- =====================================================================
-- i4 agent-proposal smoke — synthetic (no model call)
--
-- Creates 3 agent-proposal work_items with stage_results pre-populated,
-- triggers maturity verified, verifies:
--   1. apply_agent_proposal persists studies row (kind=source_type)
--   2. file_destination is set per source_type
--   3. agent_proposal_applied_at is set
--   4. enqueue_work_item_file fires (pending_file_writes row created)
--   5. file_enqueued_at is set
-- =====================================================================

\set ON_ERROR_STOP off

BEGIN;

-- ---------------------------------------------------------------------
-- Smoke 1: exhibit (the new source_type, SC-relevant)
-- ---------------------------------------------------------------------

\echo '=== smoke 1: exhibit ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_validate_output jsonb;
BEGIN
    v_validate_output := jsonb_build_object(
        'source_type', 'exhibit',
        'slug', 'ai-bias-knn-classifier-smoke',
        'title', 'AI Bias Demo: k-NN Shape Classifier with Skewed Training Data',
        'body', E'# AI Bias Demo\n\nMinimal offline HTML/JS exhibit that demonstrates AI bias by letting visitors input skewed training data and observe classification failures.\n\n## How it works\n\n...',
        'frontmatter', jsonb_build_object(
            'target_audience', 'ages 5+',
            'science_topic', 'machine learning bias',
            'materials', jsonb_build_array('laptop', 'touchscreen', 'webcam'),
            'interaction_time_seconds', 90
        ),
        'project_association', 'space-center',
        'rationale', 'Validates pedagogical "aha" moment of pattern mismatch made visible. Disposable proof-of-concept before TensorFlow.js exhibit.'
    );

    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         project_association, input, stage_results, actor, origin, intent_id)
    VALUES
        ('smoke-exhibit-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         'space-center',
         jsonb_build_object('draft', v_validate_output),
         jsonb_build_object('validate', jsonb_build_object('output', v_validate_output::text)),
         'agent',
         'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    RAISE NOTICE 'smoke 1: created exhibit work_item %', v_wi_id;

    -- Manually trigger maturity advance (simulates what apply_gate_decision would do)
    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi_id;

    -- Check results
    RAISE NOTICE 'smoke 1: post-trigger state for %', v_wi_id;
END $$;

\echo '--- exhibit work_item state ---'
SELECT slug, maturity, file_destination, file_enqueued_at IS NOT NULL AS enqueued,
       agent_proposal_applied_at IS NOT NULL AS applied
  FROM stewards.work_items
 WHERE pipeline_family = 'agent-proposal'
   AND slug LIKE 'smoke-exhibit-%'
 ORDER BY created_at DESC LIMIT 1;

\echo '--- studies row for exhibit ---'
SELECT slug, title, kind, project_association, file_path,
       frontmatter->>'origin' AS origin_meta
  FROM stewards.studies
 WHERE kind = 'exhibit' AND slug = 'ai-bias-knn-classifier-smoke';

\echo '--- pending_file_writes for exhibit ---'
SELECT target_path, write_mode, length(content) AS content_len, source_kind, source_id
  FROM stewards.pending_file_writes
 WHERE target_path = 'exhibits/ai-bias-knn-classifier-smoke.md';

-- ---------------------------------------------------------------------
-- Smoke 2: study source_type
-- ---------------------------------------------------------------------

\echo ''
\echo '=== smoke 2: study ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_out jsonb;
BEGIN
    v_out := jsonb_build_object(
        'source_type', 'study',
        'slug', 'agent-proposed-study-smoke',
        'title', 'Synthetic study from agent-proposal smoke',
        'body', E'# Agent-proposed study\n\nFor smoke testing the agent-proposal pipeline only.',
        'frontmatter', '{}'::jsonb,
        'project_association', NULL,
        'rationale', 'Verifies agent-proposal pipeline can land a study via apply_agent_proposal + studies kind=study + file at study/<slug>.md.'
    );

    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         input, stage_results, actor, origin, intent_id)
    VALUES
        ('smoke-study-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         jsonb_build_object('draft', v_out),
         jsonb_build_object('validate', jsonb_build_object('output', v_out::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi_id;
    RAISE NOTICE 'smoke 2: study work_item % advanced to verified', v_wi_id;
END $$;

\echo '--- study work_item state ---'
SELECT slug, file_destination, file_enqueued_at IS NOT NULL AS enqueued,
       agent_proposal_applied_at IS NOT NULL AS applied
  FROM stewards.work_items
 WHERE pipeline_family = 'agent-proposal'
   AND slug LIKE 'smoke-study-%'
 ORDER BY created_at DESC LIMIT 1;

\echo '--- studies row for study ---'
SELECT slug, kind, file_path FROM stewards.studies
 WHERE slug = 'agent-proposed-study-smoke';

-- ---------------------------------------------------------------------
-- Smoke 3: rejected (validator error in output)
-- ---------------------------------------------------------------------

\echo ''
\echo '=== smoke 3: rejected proposal (validator error) ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_err jsonb := jsonb_build_object('error', 'source_type missing from draft');
BEGIN
    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         input, stage_results, actor, origin, intent_id)
    VALUES
        ('smoke-reject-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         '{}'::jsonb,
         jsonb_build_object('validate', jsonb_build_object('output', v_err::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi_id;
    RAISE NOTICE 'smoke 3: rejected work_item %', v_wi_id;
END $$;

\echo '--- rejected work_item state (apply_agent_proposal returned false) ---'
SELECT slug, maturity, file_destination,
       agent_proposal_applied_at IS NOT NULL AS applied
  FROM stewards.work_items
 WHERE pipeline_family = 'agent-proposal'
   AND slug LIKE 'smoke-reject-%'
 ORDER BY created_at DESC LIMIT 1;

ROLLBACK;

\echo ''
\echo '=== smoke complete (transaction rolled back) ==='
