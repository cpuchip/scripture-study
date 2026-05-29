-- Smoke 2: subset selection — 1 existing + 2 new lenses. Inspect dispatch
-- payload to verify pipeline_family resolution and model resolution from
-- metadata.default_model (no per-lens override).
SELECT stewards.start_brainstorm(
    p_binding_question := 'SMOKE: how should the chat-server credential model work?',
    p_destination      := '/tmp/smoke-j9-subset.md',
    p_slug             := 'smoke-j9-subset',
    p_lenses           := ARRAY['scamper', 'mind-mapping', 'starbursting']
) AS parent_id;

SELECT wi.slug,
       wi.pipeline_family,
       wi.status,
       wi.model_override,
       wq.payload->>'requested_model' AS requested_model,
       wq.status AS wq_status
  FROM stewards.work_items wi
  LEFT JOIN stewards.work_queue wq
    ON wi.id = (wq.payload->>'_work_item_id')::uuid
 WHERE wi.slug LIKE 'smoke-j9-subset-%'
   AND wi.pipeline_family LIKE 'brainstorm-%'
 ORDER BY wi.slug;

-- Cancel immediately so the bridge doesn't spend on the smoke.
UPDATE stewards.work_queue
   SET status = 'cancelled'
 WHERE status = 'pending'
   AND payload->>'_work_item_id' IN (
       SELECT id::text FROM stewards.work_items WHERE slug LIKE 'smoke-j9-subset-%'
   )
RETURNING id, payload->>'_pipeline_family' AS pipeline_family;

UPDATE stewards.work_items
   SET status = 'cancelled'
 WHERE slug LIKE 'smoke-j9-subset%'
   AND status NOT IN ('completed', 'cancelled')
RETURNING id, slug, status;
