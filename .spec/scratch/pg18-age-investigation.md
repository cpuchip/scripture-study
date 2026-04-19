# PG18 Native Vector + AGE Investigation

**Date:** 2026-04-18
**Triggered by:** Michael read reports suggesting PG18 has native vector storage; wondering if pgvector is now redundant and whether we should pivot to PG18 + AGE.

## Findings

### 1. PG18 has NO native vector type or HNSW index

Searched the PG18 source (`external_context/postgres` checked out at `REL_18_3`) and the official release notes (`doc/src/sgml/release-18.sgml`). The ONLY mention of "vector" in PG18 changes:

> "Guard against unexpected dimensions of `oidvector`/`int2vector`"

These are **internal system catalog array types** for storing OID/int2 lists. They are NOT ML embedding vectors. They have no distance operators, no ANN index, no relation to similarity search.

PG18 actually shipped:
- Asynchronous I/O subsystem (which **benefits** pgvector queries by ~2-3x in some benchmarks)
- Skip scan for B-tree
- UUIDv7
- Virtual generated columns
- Improved logical replication
- ...

But there is **no native vector type** in PG18. The reports Michael read were almost certainly either (a) about the AIO performance win that helps pgvector, (b) about some cloud vendor's PG18 fork (e.g. Supabase, AlloyDB), or (c) confused with `oidvector`.

**Verdict: pgvector is still required for vector storage + ANN search on stock PG18.**

### 2. Apache AGE is a graph extension, not a vector replacement

AGE adds the openCypher graph query language to PostgreSQL. It lets you store nodes/edges and run `MATCH (a)-[r]->(b) RETURN ...` queries. It is unrelated to vector similarity.

AGE's docker artifact (`external_context/age/docker/Dockerfile`) is a **build that compiles AGE from source on top of `postgres:18`** and produces a standalone image with both Postgres 18 + AGE preloaded. It is NOT a build environment — the final stage is a runnable Postgres image with `shared_preload_libraries=age`.

So:
- `apache/age:PG18` (if published) = Postgres 18 + AGE, no pgvector
- `pgvector/pgvector:pg18` (what we use today) = Postgres 18 + pgvector, no AGE

To get **both** in one image we'd have to build our own (combine the two Dockerfiles). Not hard, but it's another piece of infra to maintain.

### 3. Does this project actually want AGE?

AGE shines for graph traversals: "find all verses within 3 hops of 1 Nephi 3:7 in the cross-reference graph", "traverse Topical Guide categories", "shortest doctrinal path from grace → faith". Those are **interesting** queries.

But Phase 1's needs are:
- Keyword search (FTS — done, native PG)
- Semantic search (vector ANN — pgvector)
- Reference lookup (table queries — done)
- Talk metadata + TITSW filters (table queries — done)

Cross-references in our current schema are a flat junction table (`cross_references`). For "show me references for verse X" that's one indexed query — we don't need a graph engine. We'd only want AGE if we started doing **multi-hop graph reasoning** on cross-references, themes, and concepts. That's a Phase 3+ idea, not Phase 1.

## Recommendation: do NOT pivot

Stay on `pgvector/pgvector:pg18`. Reasoning:

1. PG18 has no native vector → pgvector is still the right tool
2. AGE solves a different problem (graph) that we don't currently have
3. We just verified the full Phase 1 stack works end-to-end on `pgvector/pgvector:pg18` — pivoting now would throw away validated work for no functional gain
4. Less infra to maintain (no custom image, no GHCR/Docker Hub publish pipeline)

## When AGE WOULD make sense (deferred)

Build a custom image and host it on **GitHub Container Registry (GHCR)** if/when we want any of:
- Multi-hop cross-reference graph traversal
- Theme/concept graphs derived from TITSW + chapter-lens enrichment
- "Shortest doctrinal path between two ideas" queries
- Family-tree / ward-graph queries (becoming side)

GHCR is free for public images and integrates with the repo permissions we already have. Docker Hub free tier has pull-rate limits; GHCR doesn't. Use GHCR.

If/when we go this direction, the image build is ~30 lines of Dockerfile combining the two upstream recipes:

```dockerfile
FROM pgvector/pgvector:pg18 AS pgvec
FROM postgres:18 AS build
RUN apt-get update && apt-get install -y bison flex build-essential postgresql-server-dev-18
COPY age /age
WORKDIR /age
RUN make && make install

FROM postgres:18
# pgvector files
COPY --from=pgvec /usr/lib/postgresql/18/lib/vector.so /usr/lib/postgresql/18/lib/
COPY --from=pgvec /usr/share/postgresql/18/extension/vector* /usr/share/postgresql/18/extension/
# age files
COPY --from=build /usr/lib/postgresql/18/lib/age.so /usr/lib/postgresql/18/lib/
COPY --from=build /usr/share/postgresql/18/extension/age* /usr/share/postgresql/18/extension/
CMD ["postgres", "-c", "shared_preload_libraries=age"]
```

Push as `ghcr.io/cpuchip/postgres-age-vec:18` from a GitHub Actions workflow on tag push.

## What to do right now

Nothing — stay the course. Keep `pgvector/pgvector:pg18`. Document the AGE option as a deferred "Phase 3: graph reasoning" item in the proposal so it isn't forgotten.
