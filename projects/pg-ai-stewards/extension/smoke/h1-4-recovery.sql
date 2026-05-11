-- H.1.4 recovery: enable research MCP servers, kill stuck work_item,
-- create fresh work_item for retry.
BEGIN;

-- Enable the research MCP servers
UPDATE stewards.mcp_servers
   SET enabled = true
 WHERE name IN ('search', 'exa-search', 'yt');

-- Show what's enabled now
SELECT name, enabled FROM stewards.mcp_servers ORDER BY name;

-- Kill the stuck work_queue row + the work_item
UPDATE stewards.work_queue
   SET status='error', error='h1.4 recovery: bgworker crashed on disabled MCP server', done_at=now()
 WHERE id = 1227 AND status='in_progress';

DELETE FROM stewards.work_queue WHERE payload->>'_work_item_id' = '14bf04a1-4324-4f68-875d-caea40bc6bbb';
DELETE FROM stewards.messages WHERE session_id LIKE 'wi--14bf04a1%';
DELETE FROM stewards.sessions WHERE id LIKE 'wi--14bf04a1%';
DELETE FROM stewards.work_items WHERE id = '14bf04a1-4324-4f68-875d-caea40bc6bbb';

COMMIT;

-- Fresh work_item with slug suffix v2
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
        'ai-tools-weekly-2026-05-11-v2',
        'human',
        NULL,
        v_intent_id
    ) INTO v_wi;
    UPDATE stewards.work_items SET file_destination = 'research/ai-tools-weekly-2026-05-11.md' WHERE id = v_wi;
    v_work_id := stewards.work_item_dispatch_stage(v_wi);
    RAISE NOTICE 'fresh work_item: %  gather dispatched: work_id=%', v_wi, v_work_id;
END
$$;

SELECT id, slug, status, current_stage FROM stewards.work_items WHERE slug = 'ai-tools-weekly-2026-05-11-v2';
