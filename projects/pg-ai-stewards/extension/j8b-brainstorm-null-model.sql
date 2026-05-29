-- =====================================================================
-- Batch J.8.b — NULL existing 4 brainstorm-* lens models
-- =====================================================================
-- Per ratified Q4 (2026-05-29): existing 4 lenses move from hardcoded
-- stages[0].model to NULL with the today-default preserved in
-- pipelines.metadata.suggested_model as a UI hint.
--
-- The fallback chain (J.8.a) means a NULL stages.model resolves via:
--   pipeline metadata.default_model → catalog_default_model(provider).
-- We set metadata.default_model to the today-default model so that calls
-- to start_brainstorm() WITHOUT p_models continue to dispatch the same
-- model split as today (qwen3.6-plus / kimi-k2.6 mix per B1 ratification).
--
-- The distinction between default_model and suggested_model:
--   - default_model    → what the dispatcher uses when nothing higher set
--   - suggested_model  → UI-visible hint to help callers pick when they
--                        do choose to override
-- For these 4 we set both to the same value so behavior is preserved AND
-- the UI surfaces the lens-author's preferred model.
--
-- B1 intent honored: each lens still "declares its provider/model"; the
-- declaration just lives in metadata instead of stages, freeing callers
-- to override at any layer.
-- =====================================================================


-- SCAMPER — was qwen3.6-plus / opencode_go
UPDATE stewards.pipelines
   SET stages = jsonb_set(
                    jsonb_set(stages, '{0,model}',    'null'::jsonb),
                                       '{0,provider}','null'::jsonb
                ),
       metadata = metadata
                  || jsonb_build_object(
                       'default_model',   'qwen3.6-plus',
                       'default_provider','opencode_go',
                       'suggested_model', 'qwen3.6-plus',
                       'suggested_provider','opencode_go'
                  )
 WHERE family = 'brainstorm-scamper';

-- Six Hats — was kimi-k2.6 / opencode_go
UPDATE stewards.pipelines
   SET stages = jsonb_set(
                    jsonb_set(stages, '{0,model}',    'null'::jsonb),
                                       '{0,provider}','null'::jsonb
                ),
       metadata = metadata
                  || jsonb_build_object(
                       'default_model',   'kimi-k2.6',
                       'default_provider','opencode_go',
                       'suggested_model', 'kimi-k2.6',
                       'suggested_provider','opencode_go'
                  )
 WHERE family = 'brainstorm-six-hats';

-- Crazy 8s — was qwen3.6-plus / opencode_go
UPDATE stewards.pipelines
   SET stages = jsonb_set(
                    jsonb_set(stages, '{0,model}',    'null'::jsonb),
                                       '{0,provider}','null'::jsonb
                ),
       metadata = metadata
                  || jsonb_build_object(
                       'default_model',   'qwen3.6-plus',
                       'default_provider','opencode_go',
                       'suggested_model', 'qwen3.6-plus',
                       'suggested_provider','opencode_go'
                  )
 WHERE family = 'brainstorm-crazy8s';

-- Reverse — was kimi-k2.6 / opencode_go
UPDATE stewards.pipelines
   SET stages = jsonb_set(
                    jsonb_set(stages, '{0,model}',    'null'::jsonb),
                                       '{0,provider}','null'::jsonb
                ),
       metadata = metadata
                  || jsonb_build_object(
                       'default_model',   'kimi-k2.6',
                       'default_provider','opencode_go',
                       'suggested_model', 'kimi-k2.6',
                       'suggested_provider','opencode_go'
                  )
 WHERE family = 'brainstorm-reverse';


-- =====================================================================
-- Acceptance (verify before commit):
--
--   1. SELECT stages->0->>'model'  FROM stewards.pipelines WHERE family LIKE 'brainstorm-%';
--      → all four return NULL.
--
--   2. SELECT metadata->>'default_model' FROM stewards.pipelines WHERE family LIKE 'brainstorm-%';
--      → scamper/crazy8s = 'qwen3.6-plus'; six-hats/reverse = 'kimi-k2.6'.
--
--   3. Behavior preservation: call start_brainstorm() with no p_models.
--      Each lens's spawned child has stages.model NULL, then resolves to
--      metadata.default_model via the J.8.a fallback chain. work_queue
--      payload should be byte-identical to today's payload (same model
--      per lens, same provider).
-- =====================================================================
