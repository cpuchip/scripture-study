# pg_ai_stewards extension — dev stack

The actual Postgres extension. Phase 1 of [the project](../).

## Status

**Phase 1, step 1 done (2026-05-02):** pgrx scaffold builds, the
extension loads into PG18 alongside pgvector + Apache AGE, and
`stewards.version()` returns `0.1.0` end-to-end. Everything else from
the [phase plan](../phases.md#phase-1--foundation-extension-scaffold--bgworker--brain-port)
is still ahead.

## Layout

```
extension/
├── Cargo.toml                  # pgrx 0.18.0, default-features = ["pg18"]
├── pg_ai_stewards.control      # PG control file, schema = stewards
├── src/
│   └── lib.rs                  # one-function scaffold (version, pgrx_version)
├── Dockerfile                  # multi-stage: rust builder + runtime w/ pgvector+AGE
├── docker-compose.yaml         # dev stack on host port 55433
└── init/
    └── 00-extensions.sql       # CREATE EXTENSION x3 on first boot
```

## Build & run

```pwsh
cd projects\pg-ai-stewards\extension
docker compose build         # ~2 min cold; ~30s warm thanks to layer cache
docker compose up -d
```

Then verify:

```pwsh
docker exec -it pg-ai-stewards-dev psql -U stewards -d stewards `
  -c "SELECT extname, extversion FROM pg_extension WHERE extname IN ('vector','age','pg_ai_stewards') ORDER BY extname;" `
  -c "SELECT stewards.version();"
```

Expected:

```
    extname     | extversion
----------------+------------
 age            | 1.7.0
 pg_ai_stewards | 0.1.0
 vector         | 0.8.2

 version
---------
 0.1.0
```

The dev stack runs on **port 55433** so it doesn't collide with the
probe stack on 55432. Both can run simultaneously.

## Tear down

```pwsh
docker compose down -v      # -v drops the volume so init runs again
```

## Dev loop

This is a deliberately **slow** dev loop for now: every code change
requires a full image rebuild. That's fine for the scaffold step
because changes are infrequent. When iteration starts to bite, swap in
a mounted-source dev container (Rust + cargo-pgrx with the source
directory bind-mounted) that builds in place and re-installs into the
running Postgres without rebuilding the image. Track that as a Phase 1
quality-of-life upgrade.

## Notes for next session

- pgrx `pg_module_magic!` in 0.18 wants `CStr` arguments if you pass
  named ones; the no-arg form is simpler and pulls metadata from
  `Cargo.toml`. Already applied here.
- `cargo pgrx package --out-dir /out` produces a tree rooted at `/`
  (e.g. `/out/usr/lib/postgresql/18/lib/pg_ai_stewards.so`), NOT a
  named subdirectory. The `COPY --from=builder /out/ /` line in the
  Dockerfile depends on this.
- The runtime image is `pgvector/pgvector:pg18` + Apache AGE built
  from source, exactly matching the [probe](../probe/Dockerfile).
  When AGE or pgvector versions change, change them in both places.

## Next steps (per phases.md Phase 1)

1. **bgworker scaffold** — `cargo pgrx new --bgworker` template, then
   register a worker that listens on `LISTEN stewards_dispatch` and
   reads from `stewards.work_queue`. Reference: [pg_vectorize](https://github.com/ChuckHend/pg_vectorize).
2. **Schema for brain replacement** — `stewards.brain_entries`,
   `stewards.messages`, HNSW index, JSONB props.
3. **Migrator** — Go binary reading `scripts/brain/`'s SQLite +
   chromem-go vector store, writing into Postgres.
4. **Brain CLI driver** — Postgres backend behind the existing brain
   API surface; SQLite stays as read-only fallback for ~30 days.
5. **Real provider call through bgworker** — Ollama embedding for new
   brain entries, end-to-end.
