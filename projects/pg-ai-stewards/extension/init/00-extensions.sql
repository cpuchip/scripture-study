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
