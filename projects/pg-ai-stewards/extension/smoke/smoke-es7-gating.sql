-- ES.3.s2 gating smoke — exercises tool_dispatch_complete_waiting itself,
-- both the normal path (no judge) and the judge-gated path. Rolled back.
\set ON_ERROR_STOP on
BEGIN;

INSERT INTO stewards.sessions (id,kind,label)
VALUES ('es7g-small','chat','gating small'), ('es7g-large','chat','gating large');

-- Small tool result + its waiting tool_dispatch.
INSERT INTO stewards.work_queue (kind,provider,status,result)
VALUES ('mcp_proxy','test','done', jsonb_build_object('content','a small tool result'))
RETURNING id AS small_child \gset
INSERT INTO stewards.work_queue (kind,provider,status,payload,result)
VALUES ('tool_dispatch','opencode_go','waiting_for_tools',
  jsonb_build_object('agent_family','research','model','kimi-k2.6',
                     'session_id','es7g-small','parent_work_id',0),
  jsonb_build_object('resolved','[]'::jsonb,'pending',
    jsonb_build_array(jsonb_build_object('child_work_id',:small_child,'tc_id','tc-s','name','x'))))
RETURNING id AS small_disp \gset

-- Large (>50K) tool result + its waiting tool_dispatch.
INSERT INTO stewards.work_queue (kind,provider,status,result)
VALUES ('mcp_proxy','test','done',
        jsonb_build_object('content', repeat('big tool result body. ', 4000)))
RETURNING id AS large_child \gset
INSERT INTO stewards.work_queue (kind,provider,status,payload,result)
VALUES ('tool_dispatch','opencode_go','waiting_for_tools',
  jsonb_build_object('agent_family','research','model','kimi-k2.6',
                     'session_id','es7g-large','parent_work_id',0),
  jsonb_build_object('resolved','[]'::jsonb,'pending',
    jsonb_build_array(jsonb_build_object('child_work_id',:large_child,'tc_id','tc-l','name','fetch'))))
RETURNING id AS large_disp \gset

SELECT 'completed passes: ' || stewards.tool_dispatch_complete_waiting()::text;

\echo '--- normal path (small result, no judge) ---'
SELECT 'G1 small dispatch done: ' || (status='done')::text FROM stewards.work_queue WHERE id=:small_disp;
SELECT 'G2 small continuation enqueued: ' || (result ? 'next_chat_work_id')::text FROM stewards.work_queue WHERE id=:small_disp;
SELECT 'G3 small tool msg is normal: ' || (content NOT LIKE '[JUDGE-PENDING]%')::text FROM stewards.messages WHERE parent_work_id=:small_disp;

\echo '--- judge-gated path (large result) ---'
SELECT 'G4 large dispatch done: ' || (status='done')::text FROM stewards.work_queue WHERE id=:large_disp;
SELECT 'G5 large judge_pending flag set: ' || (result ? 'judge_pending')::text FROM stewards.work_queue WHERE id=:large_disp;
SELECT 'G6 large continuation NOT enqueued by dispatch: ' || (NOT (result ? 'next_chat_work_id'))::text FROM stewards.work_queue WHERE id=:large_disp;
SELECT 'G7 large tool msg is [JUDGE-PENDING]: ' || (content LIKE '[JUDGE-PENDING]%')::text FROM stewards.messages WHERE parent_work_id=:large_disp;
SELECT 'G8 judge chat queued for it: ' || count(*)::text
  FROM stewards.work_queue
 WHERE kind='chat' AND (payload->>'_judge_brief_target_msg_id')::bigint
       = (SELECT id FROM stewards.messages WHERE parent_work_id=:large_disp);

ROLLBACK;
\echo '--- rolled back ---'
