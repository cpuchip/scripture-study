-- Cleanup: cancel + delete the test work_items and their pending work_queue rows.
-- Smoke 1's rows were already in_progress when the bridge claimed them so we
-- leave those to complete naturally — they're valid opencode_go models. Smoke
-- 2 and 3 used overrides (opus-4.7, haiku-4.5, gpt-5/openai) that aren't
-- registered on opencode_go and would fail at the bridge — cancel them first.

-- 1. Cancel pending work_queue rows for smoke 2 + 3 (status='pending' = not yet claimed).
UPDATE stewards.work_queue
   SET status = 'cancelled'
 WHERE status = 'pending'
   AND payload->>'_work_item_id' IN (
       SELECT id::text FROM stewards.work_items
        WHERE slug LIKE 'smoke-j8-override-%'
           OR slug LIKE 'smoke-j8-object-%'
   )
RETURNING id, payload->>'_pipeline_family' AS pipeline_family;

-- 2. Cancel the test work_items (parents + children + aggregators).
UPDATE stewards.work_items
   SET status = 'cancelled'
 WHERE (slug LIKE 'smoke-j8-%' OR slug LIKE 'smoke-j8-%-aggregator')
   AND status NOT IN ('completed', 'cancelled')
RETURNING id, slug, status;

-- 3. Inspect smoke-1 completion state.
SELECT wi.slug,
       wi.status,
       wi.maturity,
       wi.cost_micro,
       wq.status AS wq_status
  FROM stewards.work_items wi
  LEFT JOIN stewards.work_queue wq
    ON wi.id = (wq.payload->>'_work_item_id')::uuid
 WHERE wi.slug LIKE 'smoke-j8-default-%'
 ORDER BY wi.slug;
