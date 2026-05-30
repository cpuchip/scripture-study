-- M.2 smoke: capability substitution in work_item_dispatch_stage.
-- Fully transactional + ROLLBACK: the work_queue row never commits, so the
-- bgworker never picks it up and no chat is dispatched (zero spend). Safe to
-- run with the soak live.
\set ON_ERROR_STOP on
BEGIN;

-- ---- Case 1: usable model (kimi-k2.6) -> NO substitution ----
INSERT INTO stewards.work_items
    (pipeline_family, current_stage, status, input, model_override, intent_id, actor)
VALUES
    ('brainstorm-disney', 'lens', 'pending',
     '{"binding_question":"m2 smoke: does substitution work?","user_input":"m2 smoke usable"}'::jsonb, 'kimi-k2.6',
     (SELECT id FROM stewards.intents LIMIT 1), 'm2-smoke')
RETURNING id AS wi_usable \gset
SELECT stewards.work_item_dispatch_stage(:'wi_usable') AS wq_usable \gset

SELECT 'CASE 1 usable' AS test,
       payload->>'requested_model'        AS requested,
       (payload ? '_capability_substitution') AS has_marker
  FROM stewards.work_queue WHERE id = :wq_usable;

SELECT 'CASE 1 sub-rows (want 0)' AS test, count(*) AS n
  FROM stewards.model_substitutions WHERE work_queue_id = :wq_usable;

-- ---- Case 2: unusable model (glm-5) -> substitute + log ----
INSERT INTO stewards.work_items
    (pipeline_family, current_stage, status, input, model_override, intent_id, actor)
VALUES
    ('brainstorm-disney', 'lens', 'pending',
     '{"binding_question":"m2 smoke: does substitution work?","user_input":"m2 smoke sub"}'::jsonb, 'glm-5',
     (SELECT id FROM stewards.intents LIMIT 1), 'm2-smoke')
RETURNING id AS wi_sub \gset
SELECT stewards.work_item_dispatch_stage(:'wi_sub') AS wq_sub \gset

SELECT 'CASE 2 substituted' AS test,
       payload->>'requested_model'                       AS requested,
       payload->'_capability_substitution'->>'from'      AS sub_from,
       payload->'_capability_substitution'->>'to'        AS sub_to
  FROM stewards.work_queue WHERE id = :wq_sub;

SELECT 'CASE 2 log row' AS test,
       pipeline_model AS from_model, requested_model AS to_model, reason
  FROM stewards.model_substitutions WHERE work_queue_id = :wq_sub;

ROLLBACK;
-- After rollback: nothing persisted. Confirm no smoke rows leaked.
SELECT 'POST-ROLLBACK leak check (want 0)' AS test, count(*) AS n
  FROM stewards.work_items WHERE actor = 'm2-smoke';
