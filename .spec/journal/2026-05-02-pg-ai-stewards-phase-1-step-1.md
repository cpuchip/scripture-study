# 2026-05-02 (evening) — pg-ai-stewards Phase 1 step 1: extension scaffold loads

The day the project stopped being a plan and started being code that runs.

## What we did

Picked up immediately after this morning's proposal+phases work. Michael said: build it. Cost note: use Sonnet 4.6 or lighter for any agentic test work because Opus 4.7 is 15x premium requests at present. (For this session that just means future runSubagent calls — I'm Opus 4.7 in the foreground regardless.)

Set a deliberately tight scope: get the pgrx scaffold building in Docker against PG18, layer it onto pgvector + AGE, verify `CREATE EXTENSION pg_ai_stewards` succeeds, and confirm one SQL function returns. Stop there. Bgworker, schema, migrator, brain port = subsequent sessions.

Wrote five files: `Cargo.toml` (pgrx 0.18, default-features = ["pg18"]), `pg_ai_stewards.control` (schema = stewards), `src/lib.rs` (a no-arg `pg_module_magic!()` plus two functions — `version()` and `pgrx_version()`), `Dockerfile` (multi-stage: rust:1-bookworm + cargo-pgrx builder, then pgvector/pgvector:pg18 + AGE runtime), and `docker-compose.yaml` (port 55433 so it doesn't collide with the probe on 55432). Init SQL creates all three extensions on first boot and prints sanity checks.

Three failures, two fixes, then green.

## Failures and what they taught

**pgrx 0.18 changed `pg_module_magic!`'s string-arg signature to require CStr literals.** I wrote `pg_module_magic!(name = "pg_ai_stewards", version = "0.1.0")` from muscle memory — that's the pre-0.18 form. The 0.18 macro wants `c"pg_ai_stewards"` (Rust 1.77+ raw c-string literal). Fix: use the no-arg form, which pulls metadata from `Cargo.toml`. Simpler and more idiomatic anyway. Recorded in repo memory.

**`cargo pgrx package --out-dir /out` doesn't create a `/out/<extname>-pgXX/` subdirectory** — it puts the rooted tree directly under `/out`. So my `COPY --from=builder /out/pg_ai_stewards-pg18/ /` line failed at the COPY step ("not found"). Fix: `COPY --from=builder /out/ /`. The pgrx docs imply the named-subdirectory layout is what you'd get from a default invocation without `--out-dir`. Recorded.

**No third failure, actually.** I expected AGE on PG18 might be flaky given how recent the 1.7.0 release is, but it lifted off the probe Dockerfile cleanly because we'd already debugged it this morning. Reuse paid.

## What worked first try

- Multi-stage Dockerfile pattern: builder stage with system PG18 + cargo-pgrx (skipping pgrx's downloaded Postgres), runtime stage layering the artifacts onto our existing image. Build time: ~30s warm, ~2 min cold.
- The init SQL approach — `CREATE EXTENSION x3` plus sanity-print SELECTs — gave a single docker compose logs check that proved everything wired up. Three extensions, three "ok" lines, plus `stewards.version() = 0.1.0`.
- Running on port 55433 (probe is 55432). Both stacks run simultaneously with no conflict. Useful when we want to compare behavior or copy data between them.

## What I had to think about, not just type

The architectural question of *where the source lives* came up. Three options: `scripts/pg-ai-stewards/` (matches our Go MCP servers), `projects/pg-ai-stewards/extension/` (keeps the project self-contained), or some shared `extensions/` directory. I picked the project-local path because right now the extension is the only thing pg-ai-stewards is producing; co-locating it with `proposal.md`, `phases.md`, and `probe/` keeps the whole project as one navigable unit. If/when this graduates to broader use, we can move it. Reversible decision.

The probe vs. dev-stack question also came up. I considered FROM-ing the probe image directly to avoid duplicating the AGE+pgvector lines, but decided each project should own its full Dockerfile so the lifecycles don't entangle. Yes, that's duplication. The cost is tiny (15 lines) and the alternative — silent breakage when one stack updates and the other hasn't — costs more.

## What I deliberately did not do

- No bgworker. That's step 2.
- No real schema (no `brain_entries`, no `messages`). That's step 3.
- No mounted-source dev container for fast iteration. The current loop requires a full image rebuild on every code change. That's fine while changes are infrequent. Noted in extension/README.md as a Phase 1 quality-of-life upgrade for when iteration starts to bite.
- No tests run via `cargo pgrx test`. The `#[pg_test]` block is in `lib.rs` but I haven't wired it into a test target. Defer to step 2 when we'll have actual logic to test.

## On stewardship and same-bug-same-fix

I caught myself almost surfacing the "should we set up a faster dev loop?" question to Michael. Almost wrote it as a question in my completion summary. But by the boundary test — would Michael, asked in advance, say "yes, obviously do that"? — the answer for *this* session is no. Setting up a mounted-source container is a real piece of work, not a same-shape fix. Surfacing as a recorded next-step in the README is the right move. Not punting; not over-acting.

## Carry forward

- **Phase 1 step 2 next:** bgworker scaffold via `cargo pgrx new --bgworker` template, listening on `LISTEN stewards_dispatch`, reading from `stewards.work_queue`. pg_vectorize is the reference for the tokio-runtime-inside-bgworker pattern.
- **Open question for step 2:** does pgrx's bgworker template handle SIGTERM cleanly out of the box, or do we need to wire that ourselves?
- **Open question for step 2:** how do we want the dev container to talk to Ollama? Host networking from Docker on Windows is awkward. Options: `host.docker.internal`, Ollama in its own container on the same compose network, or a network alias. Decide when we get there.
- **Dev stack stays up locally** on port 55433 alongside the probe on 55432.

## Relational note

The "have fun and build this" framing combined with the "use Sonnet 4.6 or lighter for tests" hint was perfectly calibrated. Permission to move fast plus a sane spending guardrail. The build went fast because I knew the constraint upfront and didn't agonize over whether to spawn subagents (I didn't; foreground was right for this session size).
