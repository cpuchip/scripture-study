-- H.1.4: First real research-write run. NOT a smoke — this is a REAL
-- e2e dispatch with cost. Cancel by hand if anything looks wrong.
DO $$
DECLARE
    v_wi uuid;
    v_intent_id uuid;
    v_work_id bigint;
BEGIN
    SELECT id INTO v_intent_id FROM stewards.intents WHERE slug = 'general-research';
    IF v_intent_id IS NULL THEN
        RAISE EXCEPTION 'general-research intent missing — apply h1-1 first';
    END IF;

    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"What shipped in AI tooling this week (week of 2026-05-04 through 2026-05-11) that I should know about? Cover Anthropic, OpenAI, Google, Microsoft, and notable independent vendor releases. Focus on tools developers and AI engineers would actually use."}'::jsonb,
        'ai-tools-weekly-2026-05-11',
        'human',
        NULL,
        v_intent_id
    ) INTO v_wi;

    -- Default destination_maturity = NULL means full Ammon-loop to verified
    -- (Phase 5a default). file_destination defaults to pipeline template
    -- 'research/<slug>.md' which will render to 'research/ai-tools-weekly-2026-05-11.md'
    -- at synthesize stage if the work_item has its file_destination set.

    -- Explicitly opt the work_item into materialization
    UPDATE stewards.work_items
       SET file_destination = 'research/ai-tools-weekly-2026-05-11.md'
     WHERE id = v_wi;

    RAISE NOTICE 'work_item created: % (slug=ai-tools-weekly-2026-05-11)', v_wi;

    -- Dispatch gather
    v_work_id := stewards.work_item_dispatch_stage(v_wi);
    RAISE NOTICE 'gather dispatched: work_queue.id=% (kimi-k2.6 / opencode_go)', v_work_id;
END
$$;

-- Show the work item state
SELECT
    id,
    slug,
    pipeline_family,
    current_stage,
    status,
    maturity,
    file_destination
  FROM stewards.work_items
 WHERE slug = 'ai-tools-weekly-2026-05-11';
