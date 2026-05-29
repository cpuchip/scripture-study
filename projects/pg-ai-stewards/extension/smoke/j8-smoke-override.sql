-- Smoke 2: per-lens model override via p_models string shorthand
SELECT stewards.start_brainstorm(
    p_binding_question := 'SMOKE TEST: per-lens model override?',
    p_destination      := '/tmp/smoke-j8-override-test.md',
    p_slug             := 'smoke-j8-override',
    p_models           := '{"scamper":"opus-4.7","six-hats":"haiku-4.5"}'::jsonb
) AS parent_id;

-- Inspect child work_items + work_queue payload to verify resolution.
SELECT wi.slug,
       wi.model_override,
       wq.payload->>'requested_model' AS requested_model,
       wq.status AS wq_status
  FROM stewards.work_items wi
  LEFT JOIN stewards.work_queue wq
    ON wi.id = (wq.payload->>'_work_item_id')::uuid
 WHERE wi.slug LIKE 'smoke-j8-override-%'
   AND wi.pipeline_family LIKE 'brainstorm-%'
 ORDER BY wi.slug;
