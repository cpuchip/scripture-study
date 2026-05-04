-- Phase 1.6.1 verification — direct test of synthesize_tool_failure
-- (failure modes 3 + 4) plus session_status view.
--
-- The full bgworker-crash test (mode 4 with a real SIGKILL) requires
-- container restart. Here we verify the helper function works end-to-end
-- in pure SQL, which is the unit underneath both mode-3 and mode-4
-- recovery paths in the Rust code.
--
-- Run with:
--   docker exec -i pg-ai-stewards-dev psql -U stewards -d stewards \
--     < verify-1-6-1.sql

\echo
\echo === Test 1: synthesize_tool_failure inserts synthetic replies + enqueues continuation ===

INSERT INTO stewards.sessions (id, label, kind)
VALUES ('synth-test-1', 'synthesize_tool_failure unit test', 'chat')
ON CONFLICT (id) DO NOTHING;

-- A fake chat work item, status='done' as if it had completed.
INSERT INTO stewards.work_queue (kind, provider, status, payload, result, done_at)
VALUES (
    'chat', 'opencode_go', 'done',
    jsonb_build_object('session_id', 'synth-test-1',
                       'agent_family', 'stewards-explore',
                       'requested_model', 'kimi-k2.6'),
    jsonb_build_object('finish_reason', 'tool_calls'),
    now()
)
RETURNING id AS fake_chat_work_id \gset

-- A user + assistant message with two tool_calls referencing the
-- fake chat row via parent_work_id.
INSERT INTO stewards.messages (session_id, role, content, model)
VALUES ('synth-test-1', 'user', 'Hypothetical question that needs tools.', 'kimi-k2.6');

INSERT INTO stewards.messages (
    session_id, role, content, model,
    tool_calls, finish_reason, parent_work_id)
VALUES (
    'synth-test-1', 'assistant', '', 'kimi-k2.6',
    jsonb_build_array(
        jsonb_build_object(
            'id', 'call_synth_1',
            'type', 'function',
            'function', jsonb_build_object(
                'name', 'brain_search_text',
                'arguments', '{"query":"x","limit":1}'
            )
        ),
        jsonb_build_object(
            'id', 'call_synth_2',
            'type', 'function',
            'function', jsonb_build_object(
                'name', 'always_fails',
                'arguments', '{}'
            )
        )
    ),
    'tool_calls',
    :fake_chat_work_id
);

SELECT stewards.synthesize_tool_failure(
    :fake_chat_work_id,
    'stewards-explore',
    'kimi-k2.6',
    'synth-test-1',
    'opencode_go',
    'simulated dispatcher failure'
) AS continuation_work_id \gset

\echo Continuation work_id: :continuation_work_id

\echo --- Synthetic role=tool replies (expect 2, one per tool_call_id) ---
SELECT tool_call_id, content::jsonb AS content_jsonb
FROM stewards.messages
WHERE session_id = 'synth-test-1' AND role = 'tool'
ORDER BY id;

\echo --- Continuation chat enqueued? (expect 1 pending row, kind=chat) ---
SELECT id, kind, payload->>'session_id' AS session
FROM stewards.work_queue
WHERE id = :continuation_work_id;

\echo
\echo === Test 2: idempotency - second call should NOT duplicate replies ===

-- Cancel the continuation first so the bgworker doesn't run it.
UPDATE stewards.work_queue SET status = 'done', done_at = now()
 WHERE id = :continuation_work_id;

SELECT stewards.synthesize_tool_failure(
    :fake_chat_work_id,
    'stewards-explore',
    'kimi-k2.6',
    'synth-test-1',
    'opencode_go',
    'second call should be a no-op for replies'
) AS second_continuation_id;

\echo --- Tool reply count after second call (expect still 2) ---
SELECT count(*) AS tool_reply_count
FROM stewards.messages
WHERE session_id = 'synth-test-1' AND role = 'tool';

\echo
\echo === Test 3: session_status view ===

SELECT
    session_id, last_finish_reason, last_loop_stop_reason,
    pending_work, errored_work,
    total_tokens_in, total_billable_out
FROM stewards.session_status
WHERE session_id IN ('loop-3', 'loop-err2', 'synth-test-1')
ORDER BY session_id;

\echo
\echo === Test 4: cleanup synth-test-1 ===

DELETE FROM stewards.work_queue
 WHERE payload->>'session_id' = 'synth-test-1';
DELETE FROM stewards.messages WHERE session_id = 'synth-test-1';
DELETE FROM stewards.sessions WHERE id = 'synth-test-1';

\echo Done.
