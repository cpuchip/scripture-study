-- ES.3.s2 synthetic smoke — runs in a transaction, ROLLBACK at end.
-- Nothing persists, no LLM calls dispatched.
\set ON_ERROR_STOP on
BEGIN;

INSERT INTO stewards.sessions (id, kind, label)
VALUES ('es7-smoke-parent', 'chat', 'es7 smoke parent');

INSERT INTO stewards.work_queue (kind, provider, status, payload, result)
VALUES ('tool_dispatch', 'opencode_go', 'done',
  jsonb_build_object('agent_family','research','model','kimi-k2.6',
                     'session_id','es7-smoke-parent','parent_work_id', 0),
  '{}'::jsonb)
RETURNING id AS dispatch_id \gset

-- Oversized tool message — triggers the intercept on INSERT.
INSERT INTO stewards.messages (session_id, role, content, tool_call_id, parent_work_id)
VALUES ('es7-smoke-parent', 'tool',
        repeat('Lorem ipsum dolor sit amet, consectetur adipiscing. ', 6000),
        'tc-es7-smoke', :dispatch_id)
RETURNING id AS msg_id \gset

\echo '--- A: intercept ---'
SELECT 'A1 content=[JUDGE-PENDING]: ' || (content LIKE '[JUDGE-PENDING]%')::text
  FROM stewards.messages WHERE id = :msg_id;
SELECT 'A2 raw preserved rows: ' || count(*)::text
  FROM stewards.messages_raw_overflow WHERE message_id = :msg_id;
SELECT 'A3 judge chat queued: ' || count(*)::text
  FROM stewards.work_queue
 WHERE kind='chat' AND (payload->>'_judge_brief_target_msg_id') = (:msg_id)::text;
SELECT 'A4 stray K.1 extraction (want 0): ' || count(*)::text
  FROM stewards.work_queue
 WHERE kind='chat' AND (payload->>'_engram_extraction_target_msg_id') = (:msg_id)::text;

SELECT id AS judge_id FROM stewards.work_queue
 WHERE kind='chat' AND (payload->>'_judge_brief_target_msg_id') = (:msg_id)::text \gset

-- Simulate the judge chat completing with a valid brief.
UPDATE stewards.work_queue
   SET status='done',
       result = jsonb_build_object('response',
         (jsonb_build_object('choices', jsonb_build_array(
            jsonb_build_object('message', jsonb_build_object('content',
              '{"engrams":[{"id":"judge-x-e1","tier":"hot","topic":"finding",'
              '"content":"a key finding","provenance":"extracted",'
              '"preserved":{"urls":["http://x.test"],"dates":[],"names":[],"quotes":[]}}],'
              '"state":"done","discarded":"navigation and ads"}'
            )))))::text)
 WHERE id = :judge_id;

\echo '--- B: apply_judge_brief + parent resume ---'
SELECT 'B1 content=[JUDGE BRIEF]: ' || (content LIKE '[JUDGE BRIEF]%')::text
  FROM stewards.messages WHERE id = :msg_id;
SELECT 'B2 engram items (want 1): ' || jsonb_array_length(engrams->'items')::text
  FROM stewards.messages WHERE id = :msg_id;
SELECT 'B3 engram provenance: ' || (engrams->'items'->0->>'provenance')
  FROM stewards.messages WHERE id = :msg_id;
SELECT 'B4 brief state: ' || (engrams->>'state')
  FROM stewards.messages WHERE id = :msg_id;
SELECT 'B5 dispatch resumed flag: ' || (result ? 'judge_continuation_enqueued')::text
  FROM stewards.work_queue WHERE id = :dispatch_id;
SELECT 'B6 continuation chat enqueued: ' || count(*)::text
  FROM stewards.work_queue
 WHERE kind='chat' AND (payload->>'session_id')='es7-smoke-parent'
   AND NOT (payload ? '_judge_brief_target_msg_id');
\echo '--- surface preview ---'
SELECT substring(content FROM 1 FOR 400) FROM stewards.messages WHERE id = :msg_id;

ROLLBACK;
\echo '--- rolled back, nothing persisted ---'
