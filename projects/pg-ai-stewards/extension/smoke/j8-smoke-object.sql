-- Smoke 3: object-shape entry {"model", "provider"} for explicit provider switching
SELECT stewards.start_brainstorm(
    p_binding_question := 'SMOKE TEST: object override with provider?',
    p_destination      := '/tmp/smoke-j8-object-test.md',
    p_slug             := 'smoke-j8-object',
    p_models           := '{"crazy8s":{"model":"gpt-5","provider":"openai"}}'::jsonb
) AS parent_id;

SELECT wi.slug,
       wi.model_override,
       wi.provider_override,
       wq.payload->>'requested_model' AS requested_model,
       wq.provider AS wq_provider,
       wq.status AS wq_status
  FROM stewards.work_items wi
  LEFT JOIN stewards.work_queue wq
    ON wi.id = (wq.payload->>'_work_item_id')::uuid
 WHERE wi.slug LIKE 'smoke-j8-object-%'
   AND wi.pipeline_family LIKE 'brainstorm-%'
 ORDER BY wi.slug;
