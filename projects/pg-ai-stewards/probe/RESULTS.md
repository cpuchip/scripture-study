# Probe results — pgvector + Apache AGE on Postgres 18

**Date:** 2026-05-02
**Status:** All seven test blocks pass.
**Verdict:** Architecture is real. Move to spec.

## Setup
- Image: `pgvector/pgvector:pg18` + Apache AGE `release/PG18/1.7.0`
  built from source against `postgresql-server-dev-18`.
- Build time: ~50 seconds (Apache AGE compile dominates).
- Resulting image size: not measured precisely but reasonable —
  `apt-get purge` of build-essential at the end keeps the runtime layer small.
- Ports: 55432 → 5432 (avoid colliding with anything).
- Init script: ran cleanly on first boot. Both `CREATE EXTENSION`
  succeeded; `create_graph('stewards_graph'::name)` succeeded.

## Versions confirmed working

| Component | Version |
|-----------|---------|
| Postgres | 18 (from `pgvector/pgvector:pg18`) |
| pgvector | 0.8.2 |
| Apache AGE | 1.7.0 |

## What worked — first try

1. Both extensions install side-by-side. No conflicts.
2. `vector` columns + HNSW index + cosine distance operator (`<=>`).
3. Apache AGE Cypher (`MATCH`, `MERGE`, `UNWIND`, `SET`, properties).
4. AGE in the same transaction as relational + vector queries.
5. The "bridge" — pgvector cosine score written as a property on an
   AGE `SIMILAR_TO` edge inside a single `DO` block.
6. Reading the bridge edges back via Cypher.
7. **Combined CTE: pgvector nearest-neighbor → AGE filter, in one
   statement.** Found Moroni 7:45 as both top similarity *and* already
   marked `:CITES_AS_CORE` by another study. This is the Microsoft
   "stop paying the multi-database tax" claim made concrete.

## Rough edges

Three friction points worth recording for the spec.

### 1. `::name` cast on `create_graph`
The README from `Haiwen-Yin/memory-pg18-by-yhw` warned about this. I
included it preemptively; without the cast you'd get
`function create_graph(unknown) does not exist`. Cost: zero, once you
know.

### 2. AGE setup is per-session
Every connection that calls `cypher()` needs:
```sql
LOAD 'age';
SET search_path TO ag_catalog, "$user", public;
```
This is documented behavior, not a bug. For our extension we'll do it
in a connection-init hook so application code never has to.

### 3. Vertex agtype doesn't cast cleanly to JSONB
AGE's vertex serialization includes a `::vertex` suffix
(`{"id":..., "label":..., "properties":{...}}::vertex`), which breaks
`(n::text)::jsonb`. The clean pattern is to project scalar properties
out of `cypher()` directly:

```sql
SELECT (nid::text)::bigint AS id
  FROM cypher('graph', $$ MATCH (n:Doc) RETURN n.id $$) AS (nid agtype);
```

Not "project a whole vertex and then dig into its JSON." This is also
faster — Postgres only materializes the columns it needs.

## What this clears

- **Gap 3** from the research scratch: ✅ done. Verdict: build it.
- **Gaps 1 & 2** (`postgres_fdw` against vector and agtype across
  databases): not tested directly, but `vector` is a regular composite
  type and is known to FDW-safe; `agtype` will need its own probe in
  Phase 2 if we ever want SQL-level joins from `stewards` into a
  hypothetical second AGE database. For Phase 1 (gospel-engine-v2 lives
  in `gospel`, no AGE there) this is moot — the URI-string approach
  from the proposal is sufficient.
- **Gap 4** (pgai archival rationale): see scratch.md update —
  Timescale/TigerData consolidated on the Python library + worker
  pattern as part of their rebrand to "agentic Postgres," not because
  the in-DB direction failed. The lesson (worker outside backend) is
  what we already planned.

## Reproducing

```pwsh
cd projects\pg-ai-stewards\probe
docker compose down -v
docker compose up -d --build
docker exec -i pg-ai-stewards-probe psql -U stewards -d stewards -f - < bridge-test.sql
```

Or inspect interactively:
```pwsh
docker exec -it pg-ai-stewards-probe psql -U stewards -d stewards
```

## Files

- [Dockerfile](Dockerfile)
- [docker-compose.yaml](docker-compose.yaml)
- [init/00-extensions.sql](init/00-extensions.sql)
- [bridge-test.sql](bridge-test.sql)
- [README.md](README.md)
