-- =====================================================================
-- Phase 2.7b.3 — Pure-SQL verification of budget enforcement.
--
-- Calls watchman_pass_start() with various budgets, observes
-- doc_count_planned + budget_stopped, then immediately ABORTs the
-- pass and cleans up enqueued chats so no model tokens are spent.
-- =====================================================================

\set ON_ERROR_STOP on

-- Show what we expect: top-5 dirty docs and their estimates.
\echo '=== Top-5 dirty docs and per-doc estimates ==='
SELECT slug,
       length(stewards.watchman_input(slug)) AS input_chars,
       stewards.estimate_chat_tokens(slug) AS est_tokens
  FROM stewards.dirty_queue
 ORDER BY coalesce(last_consolidated_at, 'epoch'::timestamptz),
          updated_at
 LIMIT 5;

\echo

-- Helper to abort + cleanup any pass we just started so it doesn't
-- actually call the model.
CREATE OR REPLACE FUNCTION pg_temp.abort_test_pass(p_pass_id text)
RETURNS void
LANGUAGE plpgsql AS $func$
BEGIN
    -- Cancel any pending work_queue rows attached to this pass.
    UPDATE stewards.work_queue
       SET status = 'error',
           error  = '(verify-2-7b3 test pass aborted before dispatch)',
           done_at = now()
     WHERE status = 'pending'
       AND payload->>'_watchman_pass_id' = p_pass_id;
    -- Mark the pass row complete so it doesn't show as in_progress.
    UPDATE stewards.watchman_passes
       SET finished_at = now(),
           status      = 'errored'
     WHERE pass_id = p_pass_id;
END;
$func$;

\echo '=== TRIAL 1: budget=1000 → 0 docs planned, budget_stopped=true ==='
\echo '(expected: first doc estimate alone exceeds 1000)'
SELECT stewards.watchman_pass_start(
    p_limit        => 5,
    p_actor        => 'verify-2-7b3-t1',
    p_token_budget => 1000
) AS pass_id \gset
SELECT pass_id, doc_count_planned, budget_stopped, status
  FROM stewards.watchman_passes WHERE pass_id = :'pass_id';
SELECT pg_temp.abort_test_pass(:'pass_id');

\echo
\echo '=== TRIAL 2: budget=10000 → ~1 doc planned (first doc ~8k fits) ==='
SELECT stewards.watchman_pass_start(
    p_limit        => 5,
    p_actor        => 'verify-2-7b3-t2',
    p_token_budget => 10000
) AS pass_id \gset
SELECT pass_id, doc_count_planned, budget_stopped, status
  FROM stewards.watchman_passes WHERE pass_id = :'pass_id';
SELECT pg_temp.abort_test_pass(:'pass_id');

\echo
\echo '=== TRIAL 3: budget=25000 → ~2-3 docs planned ==='
SELECT stewards.watchman_pass_start(
    p_limit        => 5,
    p_actor        => 'verify-2-7b3-t3',
    p_token_budget => 25000
) AS pass_id \gset
SELECT pass_id, doc_count_planned, budget_stopped, status
  FROM stewards.watchman_passes WHERE pass_id = :'pass_id';
SELECT pg_temp.abort_test_pass(:'pass_id');

\echo
\echo '=== TRIAL 4: budget=999999 → up to limit (5), budget_stopped=false ==='
SELECT stewards.watchman_pass_start(
    p_limit        => 5,
    p_actor        => 'verify-2-7b3-t4',
    p_token_budget => 999999
) AS pass_id \gset
SELECT pass_id, doc_count_planned, budget_stopped, status
  FROM stewards.watchman_passes WHERE pass_id = :'pass_id';
SELECT pg_temp.abort_test_pass(:'pass_id');

\echo
\echo '=== Cleanup ==='
DELETE FROM stewards.watchman_passes WHERE actor LIKE 'verify-2-7b3-%';
DELETE FROM stewards.work_queue
 WHERE payload->>'_watchman_actor' LIKE 'verify-2-7b3-%';
DELETE FROM stewards.messages
 WHERE session_id LIKE 'watchman-%--%'
   AND session_id IN (
       SELECT session_id FROM stewards.messages WHERE created_at > now() - interval '5 minutes'
   )
   AND session_id NOT IN (
       SELECT m.session_id FROM stewards.messages m
        JOIN stewards.work_queue wq ON wq.payload->>'session_id' = m.session_id
        WHERE wq.status NOT IN ('error')
   );
-- (Don't aggressively delete sessions — the cleanup above just trims test residue.)

\echo 'verify-2-7b3-budget: done.'
