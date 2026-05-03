-- 00-extensions.sql
-- Run on first DB init by the postgres official image entrypoint.

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS age;

-- AGE wants this in every session that calls cypher().
LOAD 'age';
SET search_path TO ag_catalog, "$user", public;

-- Per Haiwen-Yin/memory-pg18-by-yhw README, ::name cast is required on PG18.
SELECT * FROM ag_catalog.create_graph('stewards_graph'::name);
