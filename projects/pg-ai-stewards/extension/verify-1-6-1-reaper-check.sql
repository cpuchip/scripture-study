-- Mode 4 verification: read-only checks after the reaper runs.

\echo === Reaped orphan row ===
SELECT id, kind, status, left(error, 80) AS err
  FROM stewards.work_queue
 WHERE kind = 'tool_dispatch'
   AND payload->>'session_id' = 'mode-4-reaper';

\echo === Synthetic role=tool messages ===
SELECT id, role, tool_call_id, content::jsonb AS content
  FROM stewards.messages
 WHERE session_id = 'mode-4-reaper' AND role = 'tool';

\echo === Continuation chat (enqueued by reaper) ===
SELECT id, kind, status, payload->>'session_id' AS sess
  FROM stewards.work_queue
 WHERE kind = 'chat'
   AND payload->>'session_id' = 'mode-4-reaper'
   AND status != 'done'
 ORDER BY id;

\echo === Final messages after model recovery ===
SELECT id, role, finish_reason, tool_call_id, left(content, 200) AS preview
  FROM stewards.messages
 WHERE session_id = 'mode-4-reaper'
 ORDER BY id;

\echo === session_status ===
SELECT session_id, last_finish_reason, last_loop_stop_reason,
       pending_work, errored_work, total_billable_out
  FROM stewards.session_status
 WHERE session_id = 'mode-4-reaper';
