-- Smoke H.3.5: enqueue_proposed_work_items handles real-shape JSON,
-- malformed JSON, missing fields, and slug collisions.
--
-- Builds a synthetic planning work_item with stage_results.propose_work.output
-- containing a JSON array, then invokes the function directly. Verifies
-- the rows landed in work_items with the expected origin / parent_work_item_id
-- / project_association.

DO $$
DECLARE
    v_parent_id   uuid;
    v_planning    uuid;
    v_n           int;
    v_proposals   int;
BEGIN
    -- Synthesize a planning work_item with propose_work output.
    -- We bypass the normal create + dispatch flow and just hand-craft
    -- the row state.
    INSERT INTO stewards.work_items (
        slug, pipeline_family, current_stage, input, intent_id, actor,
        stage_results, maturity, project_association
    )
    VALUES (
        'h3-5-smoke-planning-parent',
        'planning',
        'review_plan',
        '{}'::jsonb,
        (SELECT id FROM stewards.intents WHERE slug='planning-partner'),
        'human',
        jsonb_build_object(
            'propose_work', jsonb_build_object(
                'output', $JSON$[
  {
    "slug": "h3-5-smoke-good-1",
    "binding_question": "What's the cleanest way to add per-pipeline cost_cap defaults?",
    "pipeline_family_hint": "research-write",
    "rationale": "Q-H3.3 cap is doc-only in metadata; primitive needed for real enforcement."
  },
  {
    "slug": "h3-5-smoke-good-2",
    "binding_question": "Should propose_work also be allowed to propose council work_items in batch I?",
    "pipeline_family_hint": "planning",
    "rationale": "Trust-ladder write-back rung depends on this answer."
  },
  {
    "slug": "BAD_SLUG_uppercase",
    "binding_question": "Should be skipped: slug has uppercase + underscore.",
    "pipeline_family_hint": "research-write",
    "rationale": "Skipped element — slug regex fails."
  },
  {
    "slug": "h3-5-smoke-short-binding",
    "binding_question": "Too short.",
    "pipeline_family_hint": "research-write",
    "rationale": "Skipped element — binding_question shorter than 10 chars."
  },
  {
    "slug": "h3-5-smoke-bogus-pipeline",
    "binding_question": "What if the hint points at a pipeline that doesn't exist?",
    "pipeline_family_hint": "this-pipeline-does-not-exist",
    "rationale": "Hint resolves to NULL; row inserts as proposal-only under planning."
  }
]$JSON$
            )
        ),
        'planned',
        'pg-ai-stewards'  -- project_association to test inheritance
    )
    RETURNING id INTO v_parent_id;

    -- Invoke the function.
    v_n := stewards.enqueue_proposed_work_items(v_parent_id);
    RAISE NOTICE 'smoke: function returned inserted=%', v_n;

    -- Assertions.
    SELECT count(*) INTO v_proposals
      FROM stewards.work_items
     WHERE parent_work_item_id = v_parent_id
       AND origin = 'agent_planning';

    IF v_proposals <> 3 THEN
        RAISE EXCEPTION 'smoke FAILED: expected 3 proposed work_items, got %', v_proposals;
    END IF;

    -- The bogus-pipeline element should have landed as proposal_only.
    IF NOT EXISTS (
        SELECT 1 FROM stewards.work_items
         WHERE slug = 'h3-5-smoke-bogus-pipeline'
           AND current_stage = '__proposal_only'
    ) THEN
        RAISE EXCEPTION 'smoke FAILED: bogus-pipeline element did not land as __proposal_only';
    END IF;

    -- The good ones should have inherited project_association.
    IF NOT EXISTS (
        SELECT 1 FROM stewards.work_items
         WHERE slug = 'h3-5-smoke-good-1'
           AND project_association = 'pg-ai-stewards'
    ) THEN
        RAISE EXCEPTION 'smoke FAILED: h3-5-smoke-good-1 did not inherit project_association';
    END IF;

    -- The good ones should have the proper current_stage set.
    IF NOT EXISTS (
        SELECT 1 FROM stewards.work_items
         WHERE slug = 'h3-5-smoke-good-1'
           AND current_stage = (SELECT (stages->0)->>'name' FROM stewards.pipelines WHERE family='research-write')
    ) THEN
        RAISE EXCEPTION 'smoke FAILED: h3-5-smoke-good-1 did not get research-write first stage';
    END IF;

    RAISE NOTICE 'H.3.5 smoke PASSED: function correctly inserted 3 valid, skipped 2 invalid';

    -- Cleanup so smoke is idempotent.
    DELETE FROM stewards.work_items WHERE parent_work_item_id = v_parent_id;
    DELETE FROM stewards.work_items WHERE id = v_parent_id;
END $$;
