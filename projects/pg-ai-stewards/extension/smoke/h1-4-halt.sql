-- H.1.4 halt: stop the runaway gather stage cleanly.
-- Findings: the gather input_template's "6-12 sources" guidance isn't
-- constraining kimi-k2.6 sufficiently. Model has done 9 chat rounds +
-- 30+ tool calls at $0.42, well past the proposal's $0.20 target.
-- Halt the work_item, set a pipeline-level cost cap for future runs.

BEGIN;

-- Quarantine the work_item
UPDATE stewards.work_items
   SET status = 'failed',
       quarantine_reason = 'h1-4 first real run halted manually at $0.42 (gather stage exceeded $0.20 budget without converging)',
       quarantined_at = now()
 WHERE slug = 'ai-tools-weekly-2026-05-11-v2';

-- Kill any in-progress work_queue rows for this work_item
UPDATE stewards.work_queue
   SET status = 'error',
       error = 'h1-4 manual halt',
       done_at = now()
 WHERE status = 'in_progress'
   AND (
     payload->>'_work_item_id' = (SELECT id::text FROM stewards.work_items WHERE slug='ai-tools-weekly-2026-05-11-v2')
     OR payload->>'session_id' = 'wi--5d61ed78--gather'
   );

-- Also kill pending dispatches for this session
UPDATE stewards.work_queue
   SET status = 'error',
       error = 'h1-4 manual halt'
 WHERE status = 'pending'
   AND (
     payload->>'_work_item_id' = (SELECT id::text FROM stewards.work_items WHERE slug='ai-tools-weekly-2026-05-11-v2')
     OR payload->>'session_id' = 'wi--5d61ed78--gather'
     OR payload->>'parent_work_id' IN (
        SELECT id::text FROM stewards.work_queue
         WHERE payload->>'_work_item_id' = (SELECT id::text FROM stewards.work_items WHERE slug='ai-tools-weekly-2026-05-11-v2')
     )
   );

-- Final cost summary
SELECT slug, status, quarantine_reason,
       cost_micro_dollars::float/1e6 AS spent_usd,
       (SELECT count(*) FROM stewards.cost_events WHERE work_item_id = work_items.id) AS n_events
  FROM stewards.work_items
 WHERE slug='ai-tools-weekly-2026-05-11-v2';

COMMIT;
