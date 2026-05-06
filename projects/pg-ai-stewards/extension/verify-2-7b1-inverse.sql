-- =====================================================================
-- Phase 2.7b.1 — Inverse hypothesis (Agans Rule 9)
--
-- Reproduce the trigger contract without burning model tokens:
--   1. Trigger PRESENT  → fake completion → verdict + finding land.
--   2. Trigger DROPPED  → fake completion → verdict + finding do NOT
--      land (proves the trigger is what's doing the work).
--   3. Trigger RESTORED → re-fire completion → verdict + finding land
--      again.
--
-- Synthesizes a "completed" watchman chat by:
--   - inserting a watchman_passes row directly
--   - inserting a session + user message + assistant message (with
--     a strict-JSON content the trigger should parse)
--   - inserting a work_queue row in status='in_progress' with
--     _watchman_pass_id payload markers
--   - UPDATEing status -> 'done' (this is what fires the trigger)
--
-- All synthetic rows live under pass_id 'inverse-test-*' so they are
-- easy to identify and drop after.
-- =====================================================================

\set ON_ERROR_STOP on

BEGIN;

-- Pick a slug we'll reuse across all three trials (must exist in
-- studies and not already have a verdict for any inverse-test pass).
\set test_slug '''art-of-delegation'''

-- Cleanup any prior state from a previous inverse-test run so we
-- start clean.
DELETE FROM stewards.findings
 WHERE pass_id LIKE 'inverse-test-%';
DELETE FROM stewards.verdicts
 WHERE pass_id LIKE 'inverse-test-%';
DELETE FROM stewards.watchman_passes
 WHERE pass_id LIKE 'inverse-test-%';
DELETE FROM stewards.work_queue
 WHERE payload->>'_watchman_pass_id' LIKE 'inverse-test-%';
DELETE FROM stewards.messages
 WHERE session_id LIKE 'inverse-test-%';
DELETE FROM stewards.sessions
 WHERE id LIKE 'inverse-test-%';

COMMIT;

-- ---------------------------------------------------------------------
-- helper: fake_watchman_completion(label, content)
-- Sets up a complete (pass + session + messages + work_queue) bundle
-- for a synthetic watchman chat. Then UPDATEs the work_queue row's
-- status from 'in_progress' to 'done', firing whatever trigger is
-- attached. Returns the pass_id used.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION pg_temp.fake_watchman_completion(
    p_label    text,
    p_content  text,
    p_slug     text DEFAULT 'art-of-delegation'
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_pass_id    text := 'inverse-test-' || p_label;
    v_session_id text := 'inverse-test-' || p_label;
    v_work_id    bigint;
BEGIN
    INSERT INTO stewards.watchman_passes
        (pass_id, started_at, trigger, provider, model, agent_family,
         token_budget, actor, status, doc_count_planned)
    VALUES
        (v_pass_id, now(), 'manual', 'opencode_go', 'kimi-k2.6',
         'watchman-consolidator', 50000, 'inverse-test',
         'in_progress', 1);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            'Inverse-test session ' || p_label, 'agent');

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', '(synthetic user)', 'kimi-k2.6');

    INSERT INTO stewards.messages
        (session_id, role, content, model, tokens_in, tokens_out)
    VALUES (v_session_id, 'assistant', p_content, 'kimi-k2.6', 100, 200);

    INSERT INTO stewards.work_queue
        (kind, provider, status, payload)
    VALUES
        ('chat', 'opencode_go', 'in_progress',
         jsonb_build_object(
             'session_id',         v_session_id,
             'agent_family',       'watchman-consolidator',
             'requested_model',    'kimi-k2.6',
             '_watchman_pass_id',  v_pass_id,
             '_watchman_slug',     p_slug,
             '_watchman_actor',    'inverse-test'
         ))
    RETURNING id INTO v_work_id;

    -- This UPDATE is what fires the trigger (or doesn't).
    UPDATE stewards.work_queue
       SET status = 'done', done_at = now()
     WHERE id = v_work_id;

    RETURN v_pass_id;
END;
$func$;

-- ---------------------------------------------------------------------
-- TRIAL 1: trigger PRESENT — verdict + finding should land.
-- ---------------------------------------------------------------------
\echo
\echo '=== TRIAL 1: trigger PRESENT ==='

SELECT pg_temp.fake_watchman_completion(
    'trial1',
    '{"verdict":"drift","reasoning":"trial 1 — trigger should harvest this","finding":{"kind":"drift","severity":"low","message":"trial 1 finding","suggested_action":"trial 1 action"}}'
) AS pass_id;

SELECT 'verdicts'      AS what, count(*)::text AS n FROM stewards.verdicts WHERE pass_id = 'inverse-test-trial1'
UNION ALL
SELECT 'findings'      AS what, count(*)::text AS n FROM stewards.findings WHERE pass_id = 'inverse-test-trial1'
UNION ALL
SELECT 'pass status'   AS what, status         AS n FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-trial1'
UNION ALL
SELECT 'verdict_counts' AS what, verdict_counts::text AS n FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-trial1';

-- ---------------------------------------------------------------------
-- TRIAL 2: trigger DROPPED — verdict + finding should NOT land.
-- ---------------------------------------------------------------------
\echo
\echo '=== TRIAL 2: trigger DROPPED ==='

DROP TRIGGER watchman_harvest_completion ON stewards.work_queue;

SELECT pg_temp.fake_watchman_completion(
    'trial2',
    '{"verdict":"drift","reasoning":"trial 2 — trigger is dropped, this should NOT be harvested","finding":{"kind":"drift","severity":"low","message":"trial 2 finding","suggested_action":"trial 2 action"}}'
) AS pass_id;

SELECT 'verdicts (should be 0)'      AS what, count(*)::text AS n FROM stewards.verdicts WHERE pass_id = 'inverse-test-trial2'
UNION ALL
SELECT 'findings (should be 0)'      AS what, count(*)::text AS n FROM stewards.findings WHERE pass_id = 'inverse-test-trial2'
UNION ALL
SELECT 'pass status (should be in_progress)' AS what, status AS n FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-trial2'
UNION ALL
SELECT 'verdict_counts (should be {})' AS what, verdict_counts::text AS n FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-trial2';

-- ---------------------------------------------------------------------
-- TRIAL 3: trigger RESTORED — verdict + finding should land again.
-- ---------------------------------------------------------------------
\echo
\echo '=== TRIAL 3: trigger RESTORED ==='

CREATE TRIGGER watchman_harvest_completion
    AFTER UPDATE OF status ON stewards.work_queue
    FOR EACH ROW
    WHEN ((NEW.kind = 'chat')
          AND (NEW.payload ? '_watchman_pass_id')
          AND (NEW.status IN ('done', 'error'))
          AND (OLD.status IS DISTINCT FROM NEW.status))
    EXECUTE FUNCTION stewards.handle_watchman_chat_completion();

SELECT pg_temp.fake_watchman_completion(
    'trial3',
    '{"verdict":"clean","reasoning":"trial 3 — trigger restored, this should be harvested again"}'
) AS pass_id;

SELECT 'verdicts (should be 1)'      AS what, count(*)::text AS n FROM stewards.verdicts WHERE pass_id = 'inverse-test-trial3'
UNION ALL
SELECT 'findings (should be 0 — clean)' AS what, count(*)::text AS n FROM stewards.findings WHERE pass_id = 'inverse-test-trial3'
UNION ALL
SELECT 'pass status (should be completed)' AS what, status AS n FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-trial3'
UNION ALL
SELECT 'verdict_counts'              AS what, verdict_counts::text AS n FROM stewards.watchman_passes WHERE pass_id = 'inverse-test-trial3';

-- ---------------------------------------------------------------------
-- TRIAL 4: trigger PRESENT but assistant returns malformed JSON.
-- Should record verdict='skipped' cleanly, never raise.
-- ---------------------------------------------------------------------
\echo
\echo '=== TRIAL 4: malformed JSON ==='

SELECT pg_temp.fake_watchman_completion(
    'trial4',
    'this is not JSON, just words { definitely not parseable'
) AS pass_id;

SELECT 'verdicts (should be 1, verdict=skipped)' AS what, count(*)::text AS n FROM stewards.verdicts WHERE pass_id = 'inverse-test-trial4'
UNION ALL
SELECT 'verdict value'                AS what, verdict AS n FROM stewards.verdicts WHERE pass_id = 'inverse-test-trial4'
UNION ALL
SELECT 'reasoning (head)'             AS what, substring(reasoning, 1, 60) AS n FROM stewards.verdicts WHERE pass_id = 'inverse-test-trial4';

-- ---------------------------------------------------------------------
-- Cleanup
-- ---------------------------------------------------------------------
\echo
\echo '=== Cleaning up inverse-test rows ==='

DELETE FROM stewards.findings
 WHERE pass_id LIKE 'inverse-test-%';
DELETE FROM stewards.verdicts
 WHERE pass_id LIKE 'inverse-test-%';
DELETE FROM stewards.watchman_passes
 WHERE pass_id LIKE 'inverse-test-%';
DELETE FROM stewards.work_queue
 WHERE payload->>'_watchman_pass_id' LIKE 'inverse-test-%';
DELETE FROM stewards.messages
 WHERE session_id LIKE 'inverse-test-%';
DELETE FROM stewards.sessions
 WHERE id LIKE 'inverse-test-%';

\echo 'inverse hypothesis: done.'
