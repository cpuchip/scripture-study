# Probe — pgvector + Apache AGE on Postgres 18

Goal: confirm the bridge pattern (compute pgvector cosine similarity,
write the result as an AGE edge, then mix structural and semantic edges
in a single Cypher `MATCH`) works on the stack we'd actually run.

This is **gap 3** from the research: stand up a docker-compose locally
with `pgvector/pgvector:pg18` + Apache AGE in the same image, both
extensions in one DB, Microsoft's bridge example end-to-end.

## Run

```pwsh
cd projects\pg-ai-stewards\probe
docker compose up -d --build
# wait ~30s for first-init scripts (extensions + create_graph)
docker compose logs pg | Select-String -Pattern 'created|ready|error|FATAL'

# psql in
docker exec -it pg-ai-stewards-probe psql -U stewards -d stewards
```

Then run the SQL in [bridge-test.sql](bridge-test.sql) one block at a
time and observe results.

## Tear down

```pwsh
docker compose down -v   # -v drops the volume so init runs again next time
```

## What's in here

- [Dockerfile](Dockerfile) — `pgvector/pgvector:pg18` + AGE built from
  source against the matching server-dev headers.
- [docker-compose.yaml](docker-compose.yaml) — one service, port mapped
  to 55432 to avoid colliding with anything local on 5432.
- [init/00-extensions.sql](init/00-extensions.sql) — runs once on first
  DB init: `CREATE EXTENSION vector; CREATE EXTENSION age; create_graph`.
- [bridge-test.sql](bridge-test.sql) — Microsoft's bridge example,
  reduced to the minimum needed to prove the pattern. No LLM calls; we
  use literal vectors so the test is self-contained.

## Success criteria

1. Both extensions install and load without error.
2. `create_graph('stewards_graph'::name)` succeeds.
3. We can insert vector rows, query nearest neighbors with `<=>`, and
   create AGE Cypher edges in the same transaction.
4. A combined CTE (pgvector → AGE) returns expected joined results.

If any of those fail, we've found the rough edge before committing to a
spec. If they all pass, gaps 1 and 2 (`postgres_fdw` against pgvector,
agtype across DBs) become normal-difficulty tasks rather than open
risks.

## Why no LLM in the probe

This probe answers the architectural question: do the two extensions
coexist and bridge cleanly on PG18? An LLM call adds API keys, network
flakiness, and embedding-shape concerns that aren't what we're testing
here. Embedding generation is well-trodden — see [pg_vectorize](../../../external_context/pg_vectorize/)
for the production pattern. We stub embeddings with literal vectors and
trust pgvector's distance math is correct (it is — see pgvector's own
test suite).
