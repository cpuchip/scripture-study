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
