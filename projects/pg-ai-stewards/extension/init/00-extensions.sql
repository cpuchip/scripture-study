-- 00-extensions.sql
-- First-boot init: load all three extensions (pgvector, AGE, pg_ai_stewards)
-- so a fresh `docker compose up` lands on a database where everything is
-- already wired and we can immediately call stewards.version().

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS age;
CREATE EXTENSION IF NOT EXISTS pg_ai_stewards;

-- AGE setup must be re-applied per session by clients that use cypher().
LOAD 'age';
SET search_path TO ag_catalog, "$user", public;

-- Sanity prints — visible in `docker compose logs pg`.
SELECT 'pgvector ' || extversion AS ok FROM pg_extension WHERE extname = 'vector';
SELECT 'age ' || extversion AS ok FROM pg_extension WHERE extname = 'age';
SELECT 'pg_ai_stewards ' || extversion AS ok FROM pg_extension WHERE extname = 'pg_ai_stewards';
SELECT 'stewards.version() = ' || stewards.version() AS ok;
SELECT 'providers loaded:' AS ok, count(*) FROM stewards.providers_loaded();

-- Smoke-test the bgworker round-trip. We enqueue here at init time,
-- but the worker won't actually run until the postmaster takes over
-- after init finishes. The next docker logs check should show the
-- row processed within ~1 second of the database accepting connections.
SELECT stewards.enqueue('echo', 'echo', '{"hello": "world"}'::jsonb) AS enqueued_id;

-- Step 3 smoke test: insert a brain entry, confirm the embed-enqueue
-- trigger fired (work_queue should have a kind='embed' row), and
-- confirm full-text search finds it.
SELECT stewards.brain_upsert(
    'study',
    'Charity is the pure love of Christ',
    'Moroni 7:47 — pure love of Christ. The fruit of the tree of life. Connected to the great commandment.',
    '{"references": "Moroni 7:47; 1 Ne 11:21-25; Matt 22:37-40", "insight": "Charity is fruit, not effort."}'::jsonb,
    ARRAY['charity', 'love', 'moroni']
) AS new_brain_entry_id;

SELECT 'embed work queued: ' || count(*)::text AS ok
    FROM stewards.work_queue WHERE kind = 'embed';

SELECT 'fts hits for charity: ' || count(*)::text AS ok
    FROM stewards.brain_search_text('charity');

-- Phase 2.1: ensure the AGE graph exists at first boot. The fn is
-- idempotent and also self-defends in import_study(), but creating
-- it here means the very first call from any psql session sees
-- the graph already in place.
SELECT stewards.ensure_studies_graph();
SELECT 'AGE graph ready: ' || name AS ok
    FROM ag_catalog.ag_graph WHERE name = 'stewards_graph';
