-- Verify 3e.2.b/c — synthetic mcp_proxy round-trip.
--
-- Usage:
--   1. Enable gospel-engine-v2 in mcp_servers
--   2. (in a separate shell) start bridge: `stewards-mcp bridge run`
--   3. psql -f verify-3e2-2.sql
--
-- This script enqueues a synthetic mcp_proxy row, polls its status,
-- and reports whether the bridge serviced it within the timeout.

-- 1. Confirm gospel-engine-v2 is registered + enabled
SELECT name, transport, enabled,
       (SELECT count(*) FROM stewards.mcp_tool_cache c
         WHERE c.server_name = s.name AND c.active) AS active_tools
  FROM stewards.mcp_servers s
 WHERE name = 'gospel-engine-v2';

-- 2. Enqueue a synthetic mcp_proxy row
\set query 'faith hope charity'

SELECT stewards.mcp_proxy_enqueue(
    'gospel-engine-v2',
    'gospel_search',
    jsonb_build_object('query', :'query', 'limit', 3),
    NULL
) AS enqueued_id \gset

\echo Enqueued mcp_proxy work_item id=:enqueued_id
\echo Waiting up to 10 seconds for bridge to respond...

-- 3. Brief settle wait. Bridge is fast — first call may pay session
-- spawn cost (~10s for a cold subprocess), subsequent calls under
-- 200ms. psql variable substitution doesn't traverse $$ blocks, so
-- a fixed sleep is simpler than a polling loop here.
SELECT pg_sleep(10);

-- 4. Show the result
SELECT id, kind, provider, status,
       jsonb_pretty(result) AS result_pretty,
       error,
       claimed_at - created_at AS claim_latency,
       done_at    - claimed_at AS exec_duration
  FROM stewards.work_queue
 WHERE id = :enqueued_id;
