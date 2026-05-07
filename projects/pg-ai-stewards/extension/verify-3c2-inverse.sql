-- =====================================================================
-- Phase 3c.2 — Inverse hypothesis (Agans Rule 9)
--
-- Synthesizes "completed" work_item chats by inserting an assistant
-- message + a work_queue row in_progress, then UPDATEing status='done'.
-- That status flip is what fires the trigger.
--
-- Three trials:
--   1. Trigger PRESENT  → work_item advances, tokens roll up.
--   2. Trigger DROPPED  → work_item stays in_progress, no rollup.
--   3. Trigger RESTORED → advances again on next event.
--
-- Plus a budget-gate trial showing token_budget caps auto-dispatch.
--
-- Zero model tokens spent (no real chat dispatch).
-- =====================================================================

\set ON_ERROR_STOP on

-- A 2-stage test pipeline (to exercise advance-to-next-stage logic
-- AND the budget gate path).
INSERT INTO stewards.pipelines (family, description, stages)
VALUES (
    'inverse-test-2stage',
    'Inverse hypothesis test pipeline. Two stages so we can verify auto-dispatch of stage 2 vs. budget-gated stop.',
    jsonb_build_array(
        jsonb_build_object(
            'name', 'first',  'agent_family', 'stewards-explore',
            'model', 'kimi-k2.6', 'provider', 'opencode_go',
            'next', 'second', 'auto_advance', true),
        jsonb_build_object(
            'name', 'second', 'agent_family', 'stewards-explore',
            'model', 'kimi-k2.6', 'provider', 'opencode_go',
            'next', null,     'auto_advance', true)
    )
)
ON CONFLICT (family) DO UPDATE SET stages = EXCLUDED.stages;

-- Helper: synthesize a "completed" chat for a work_item's current
-- stage. Inserts fake user + assistant messages, inserts work_queue
-- row in_progress with the right payload markers, then UPDATE to
-- done — that's what fires the 3c.2 trigger.
CREATE OR REPLACE FUNCTION pg_temp.fake_stage_completion(
    p_work_item_id  uuid,
    p_assistant_msg text,
    p_tokens_in     int DEFAULT 100,
    p_tokens_out    int DEFAULT 50
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_session_id  text;
    v_work_id     bigint;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    v_session_id := 'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
                    || '--' || v_wi.current_stage;

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id, 'inverse-test stage ' || v_wi.current_stage, 'agent')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', '(fake user message)', 'kimi-k2.6');
    INSERT INTO stewards.messages
        (session_id, role, content, model, tokens_in, tokens_out, finish_reason)
    VALUES
        (v_session_id, 'assistant', p_assistant_msg, 'kimi-k2.6',
         p_tokens_in, p_tokens_out, 'stop');

    INSERT INTO stewards.work_queue (kind, provider, status, payload)
    VALUES ('chat', 'opencode_go', 'in_progress',
            jsonb_build_object(
                'session_id', v_session_id,
                '_work_item_id',     p_work_item_id::text,
                '_stage_name',       v_wi.current_stage,
                '_pipeline_family',  v_wi.pipeline_family))
    RETURNING id INTO v_work_id;

    UPDATE stewards.work_items
       SET session_ids = session_ids || v_session_id,
           status      = 'in_progress'
     WHERE id = p_work_item_id;

    -- THIS UPDATE is what fires the 3c.2 trigger (or doesn't).
    UPDATE stewards.work_queue SET status = 'done', done_at = now()
     WHERE id = v_work_id;

    RETURN v_work_id;
END;
$func$;

-- Cleanup any prior inverse-test residue.
DELETE FROM stewards.work_queue
 WHERE payload->>'_pipeline_family' = 'inverse-test-2stage';
DELETE FROM stewards.messages
 WHERE session_id LIKE 'wi--%';
DELETE FROM stewards.work_items WHERE actor = 'inverse-test';

\echo
\echo '=== TRIAL 1: trigger PRESENT — auto-advance, rollup ==='
SELECT stewards.work_item_create(
    'inverse-test-2stage', '{}'::jsonb, NULL, 'inverse-test', NULL
) AS wi_id \gset
SELECT pg_temp.fake_stage_completion(:'wi_id'::uuid, 'first stage output', 100, 50);
SELECT 'wi state after trial 1' AS what,
       slug, status, current_stage, tokens_in, tokens_out
  FROM stewards.work_items WHERE id = :'wi_id'::uuid;
-- Expected: status='in_progress' (auto-advanced to stage 2 + dispatched
-- a NEW work_queue row that isn't yet 'done'), current_stage='second',
-- tokens_in=100, tokens_out=50.

-- The trigger ALSO auto-dispatched stage 2, which inserted ANOTHER
-- work_queue row. Let's complete it synthetically too and verify
-- the work_item reaches 'completed'.
SELECT pg_temp.fake_stage_completion(:'wi_id'::uuid, 'second stage output', 80, 40);
SELECT 'wi state after stage 2 completion' AS what,
       slug, status, current_stage, tokens_in, tokens_out
  FROM stewards.work_items WHERE id = :'wi_id'::uuid;
-- Expected: status='completed', tokens_in=180, tokens_out=90.

\echo
\echo '=== TRIAL 2: trigger DROPPED — no advance ==='
DROP TRIGGER work_item_advance_completion ON stewards.work_queue;

SELECT stewards.work_item_create(
    'inverse-test-2stage', '{}'::jsonb, NULL, 'inverse-test', NULL
) AS wi_id \gset
SELECT pg_temp.fake_stage_completion(:'wi_id'::uuid, 'should NOT advance', 100, 50);
SELECT 'wi state with trigger dropped' AS what,
       slug, status, current_stage, tokens_in, tokens_out
  FROM stewards.work_items WHERE id = :'wi_id'::uuid;
-- Expected: status='in_progress', current_stage='first', tokens 0/0.
-- NO auto-advance, NO rollup. Proves the trigger is load-bearing.

\echo
\echo '=== TRIAL 3: trigger RESTORED ==='
CREATE TRIGGER work_item_advance_completion
    AFTER UPDATE OF status ON stewards.work_queue
    FOR EACH ROW
    WHEN ((NEW.kind = 'chat')
          AND (NEW.payload ? '_work_item_id')
          AND (NEW.status IN ('done', 'error'))
          AND (OLD.status IS DISTINCT FROM NEW.status))
    EXECUTE FUNCTION stewards.handle_work_item_chat_completion();

SELECT stewards.work_item_create(
    'inverse-test-2stage', '{}'::jsonb, NULL, 'inverse-test', NULL
) AS wi_id \gset
SELECT pg_temp.fake_stage_completion(:'wi_id'::uuid, 'first', 100, 50);
SELECT 'wi state after trigger restored' AS what,
       slug, status, current_stage, tokens_in, tokens_out
  FROM stewards.work_items WHERE id = :'wi_id'::uuid;
-- Expected: status='in_progress', current_stage='second', tokens 100/50
-- (advance fired again).

\echo
\echo '=== TRIAL 4: budget gate prevents auto-dispatch ==='
-- Create a work_item with a TIGHT budget (smaller than what stage 1
-- will produce), then synthesize stage 1 completion. The trigger
-- should advance but NOT auto-dispatch stage 2; status=awaiting_review.
SELECT stewards.work_item_create(
    'inverse-test-2stage', '{}'::jsonb, NULL, 'inverse-test', 100
) AS wi_id \gset
SELECT pg_temp.fake_stage_completion(:'wi_id'::uuid, 'first', 100, 50);
SELECT 'wi state with budget=100 exhausted' AS what,
       slug, status, current_stage, tokens_in, tokens_out, error
  FROM stewards.work_items WHERE id = :'wi_id'::uuid;
-- Expected: status='awaiting_review', current_stage='second'
-- (advanced), tokens 100/50, error mentions budget exhausted.
-- NO further work_queue row was enqueued because dispatch was gated.

\echo
\echo '=== Cleanup ==='
DELETE FROM stewards.work_queue
 WHERE payload->>'_pipeline_family' = 'inverse-test-2stage';
DELETE FROM stewards.messages WHERE session_id LIKE 'wi--%';
DELETE FROM stewards.work_items WHERE actor = 'inverse-test';
DELETE FROM stewards.pipelines WHERE family = 'inverse-test-2stage';

\echo 'verify-3c2-inverse: done.'
