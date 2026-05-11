-- H.1.5d retry: fresh work_item with cost cap + tightened template + soft-fail enqueue
DO $$
DECLARE
    v_wi uuid;
    v_intent_id uuid;
    v_work_id bigint;
BEGIN
    SELECT id INTO v_intent_id FROM stewards.intents WHERE slug = 'general-research';

    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"What shipped in AI tooling this week (week of 2026-05-04 through 2026-05-11) that I should know about? Cover Anthropic, OpenAI, Google, Microsoft, and notable independent vendor releases. Focus on tools developers and AI engineers would actually use."}'::jsonb,
        'ai-tools-weekly-2026-05-11-v3',
        'human',
        NULL,
        v_intent_id
    ) INTO v_wi;

    -- H.1.5c: hard cost cap $0.40
    UPDATE stewards.work_items
       SET file_destination = 'research/ai-tools-weekly-2026-05-11.md',
           cost_cap_micro   = 400000
     WHERE id = v_wi;

    -- Dispatch gather with the new tight template
    v_work_id := stewards.work_item_dispatch_stage(v_wi);
    RAISE NOTICE 'H.1.5d retry: work_item=% gather work_queue.id=% cost_cap=$0.40', v_wi, v_work_id;
END
$$;

SELECT id, slug, status, current_stage,
       cost_cap_micro::float/1e6 AS cap_usd,
       file_destination
  FROM stewards.work_items
 WHERE slug='ai-tools-weekly-2026-05-11-v3';
