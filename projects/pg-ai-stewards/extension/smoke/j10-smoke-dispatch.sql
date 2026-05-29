-- Live smoke: route one brainstorm lens to Gemini, one to opencode DeepSeek.
-- Exercises the full path: dispatch -> bgworker -> provider OpenAI-compat
-- -> response -> cost tracking. Cheap models, tiny output.
SELECT stewards.start_brainstorm(
    p_binding_question := 'SMOKE J.10: name one cheap way to test an LLM provider integration.',
    p_destination      := 'study/.scratch/smoke-j10-providers.md',
    p_slug             := 'smoke-j10',
    p_lenses           := ARRAY['scamper', 'six-hats'],
    p_models           := '{"scamper":{"model":"gemini-2.5-flash-lite","provider":"google_gemini"},"six-hats":"deepseek-v4-flash"}'::jsonb,
    p_cost_cap_per_lens_micro := 100000
) AS parent_id;

SELECT wi.slug, wi.model_override, wi.provider_override, wi.status
  FROM stewards.work_items wi
 WHERE wi.slug LIKE 'smoke-j10-%'
   AND wi.pipeline_family LIKE 'brainstorm-%'
 ORDER BY wi.slug;
