-- Mode 4 setup: create the orphaned tool_dispatch row in one shot.
-- The PowerShell harness does the restart + verification phases.

-- Clean any leftover state from prior runs (esp. failed ones that
-- skipped the cleanup phase). Tool_call IDs in tests are stable, so
-- a leftover synthetic tool reply with the same tool_call_id would
-- confuse the dedup check in synthesize_tool_failure.
DELETE FROM stewards.work_queue
 WHERE payload->>'session_id' = 'mode-4-reaper';
DELETE FROM stewards.messages WHERE session_id = 'mode-4-reaper';
DELETE FROM stewards.sessions WHERE id = 'mode-4-reaper';

INSERT INTO stewards.sessions (id, label, kind)
VALUES ('mode-4-reaper', 'mode 4: bgworker crash recovery', 'chat');

INSERT INTO stewards.work_queue (kind, provider, status, payload, result, done_at)
VALUES (
    'chat', 'opencode_go', 'done',
    jsonb_build_object('session_id', 'mode-4-reaper',
                       'agent_family', 'stewards-explore',
                       'requested_model', 'kimi-k2.6'),
    jsonb_build_object('finish_reason', 'tool_calls'),
    now())
RETURNING id AS fake_chat_id \gset

INSERT INTO stewards.messages (session_id, role, content, model)
VALUES ('mode-4-reaper', 'user',
        'Use brain_search_text to find anything about Moroni 7.', 'kimi-k2.6');

INSERT INTO stewards.messages (
    session_id, role, content, model, tool_calls,
    finish_reason, parent_work_id,
    reasoning_content, reasoning_details)
VALUES (
    'mode-4-reaper', 'assistant', '', 'kimi-k2.6',
    jsonb_build_array(jsonb_build_object(
        'id', 'call_mode4_1',
        'type', 'function',
        'function', jsonb_build_object(
            'name', 'brain_search_text',
            'arguments', '{"query":"Moroni 7","limit":3}'))),
    'tool_calls',
    :fake_chat_id,
    -- Moonshot requires reasoning_content on assistant messages
    -- carrying tool_calls when thinking is enabled. In production
    -- this comes from the real provider response; in this test
    -- harness we synthesize a placeholder so the continuation
    -- chat can serialize cleanly.
    'I should call brain_search_text to find entries about Moroni 7.',
    jsonb_build_array(jsonb_build_object(
        'type', 'reasoning.text',
        'text', 'I should call brain_search_text to find entries about Moroni 7.')));

INSERT INTO stewards.work_queue (kind, provider, status, payload, claimed_at)
VALUES (
    'tool_dispatch', 'opencode_go', 'in_progress',
    jsonb_build_object(
        'parent_work_id', :fake_chat_id,
        'agent_family',   'stewards-explore',
        'model',          'kimi-k2.6',
        'session_id',     'mode-4-reaper'),
    now() - interval '10 minutes')
RETURNING id AS orphan_id \gset

\echo ORPHAN_ID=:orphan_id
\echo FAKE_CHAT_ID=:fake_chat_id

SELECT id, kind, status FROM stewards.work_queue
 WHERE id IN (:fake_chat_id, :orphan_id) ORDER BY id;
