# 2026-05-02 (late evening) — pg-ai-stewards Phase 1 step 2: bgworker dispatcher

The day pg-ai-stewards stopped being a static schema and started being a process loop.

## What we did

Picked up immediately after the secrets-hardening work. Goal for this session: Phase 1 step 2 — bgworker that listens for work, runs a stub provider, writes the result back. Cost discipline still in effect (no subagent calls, foreground-only, kept Cargo.toml deps tight).

Read the canonical pgrx 0.18 bgworker example from pgcentralfoundation/pgrx repo. The example is small and doesn't try to do too much. Adopted its `BackgroundWorker::wait_latch(Some(duration))` pattern for the loop, its `attach_signal_handlers(SIGHUP|SIGTERM)` setup, and its `connect_worker_to_spi(db, user)` initialization. Diverged on three things: (1) used `set_restart_time(5s)` so the postmaster respawns the worker on crash; (2) gated the bgworker registration on `process_shared_preload_libraries_in_progress`; (3) added a poll-and-claim pattern using `FOR UPDATE SKIP LOCKED`.

Wrote the schema for `stewards.work_queue` inline via `extension_sql!`, with a CHECK constraint on status (`pending|in_progress|done|error`) per the data-safety checklist. Partial index on `(created_at) WHERE status='pending'` to keep the claim query cheap.

Provider registry parses `STEWARDS_PROVIDER_<NAME>_<FIELD>` env vars into a `Vec<Provider>`, stores in a `OnceLock<ProviderRegistry>`, and exposes the metadata (no keys) via `stewards.providers_loaded()`.

Three things failed and got fixed before the session was done.

## Failures and what they taught

**Failure 1: providers_loaded() returned empty rows from psql.** First version populated PROVIDER_REGISTRY in the bgworker's main function. That meant only the worker process had it filled — every normal backend ran with an empty registry. The signal was that `docker compose up` showed the worker logging "loaded 3 provider(s)" but `SELECT * FROM stewards.providers_loaded()` returned 0 rows.

The lesson here is fundamental about Postgres extensions and worth keeping: **statics live in process memory, and Postgres backends are forked from the postmaster.** Anything you set in the bgworker is invisible to backends. Anything set in `_PG_init` while the postmaster is loading shared_preload_libraries is set in the postmaster's address space and inherited via fork() copy-on-write into every subsequent backend AND the bgworker. So: postmaster-time initialization for things you want every process to see, worker-time initialization for things only the worker needs.

Fix was small: moved the env parsing and `PROVIDER_REGISTRY.set()` into `_PG_init` (the preloaded branch), and the bgworker just reads the inherited value. After rebuild, `providers_loaded()` returns all 3 rows from any psql session.

**Failure 2: `client.select()` returned errors I couldn't easily turn into "is queue empty?".** First draft used `client.select(...)` for the claim query and tried to pattern-match on whether `tuple_table.first()` returned a row. Pgrx 0.18's `SpiTupleTable.first()` doesn't return `Option<&Row>` directly. Switched to `client.update(...)` (which we needed anyway because the claim is a DML write returning rows) and used `claimed.into_iter().next()` to test for the empty case. Cleaner.

**Failure 3: build cache fooled me into thinking changes had landed.** First attempt to write the new lib.rs used `create_file`, which silently failed with "file already exists" — but the multi_replace and terminal commands ran on, and Docker layer-cached the old source. The build "succeeded" with the OLD code. Caught it because the rebuild reported `CACHED` for the cargo step. Real cost: one wasted rebuild cycle. Recovery: deleted the files and used create_file fresh. Lesson worth keeping: **when create_file errors, check it.** I did call manage_todo_list right after, which I should have looked at as a signal.

## What worked first try

- The `FOR UPDATE SKIP LOCKED` claim pattern. Standard Postgres queue idiom; works exactly as expected from a bgworker via SPI.
- The `set_restart_time(5s)` semantics. The first boot of a fresh stack always FATALs once because the worker tries to connect to "stewards" before initdb has created it. The postmaster restarts the worker 5s later and it succeeds. Subsequent boots (with the volume already populated) start cleanly.
- NOTIFY from bgworker via SPI. `client.update("NOTIFY stewards_done, '<id>'")` issues a real notification on commit. Listeners running `LISTEN stewards_done` from any backend get woken up. Means external clients can react with low latency even though the worker itself polls.
- Adding `command: [postgres, -c, shared_preload_libraries=pg_ai_stewards]` to compose worked first try. No need for postgresql.conf munging or init scripts.

## Verification (Agans Rule 9 / inverse hypothesis)

Before declaring done, ran the inverse: spun up a second container off the SAME image and SAME volume, but launched it with plain `postgres` instead of with `-c shared_preload_libraries=pg_ai_stewards`. Enqueued a row. Waited 3s. Row was still `pending`. Tore that container down, brought the preloaded one back, queried the same row id. It drained to `done`. The fix is the bgworker, not coincidence.

Also verified: SIGTERM produces a clean log line (`stewards: bgworker received SIGTERM, exiting`), container restart respawns the worker, batch of 5 enqueues averages 138ms end-to-end (under the 500ms tick because the worker drains up to 16 rows per wake).

## What I deliberately did not do

- **No tokio runtime, no reqwest dependency.** The echo stub doesn't need them. Adding them now would be premature — step 6/7 (real provider calls) is when they become load-bearing. Cargo.toml stays at `pgrx + serde + serde_json` plus pgrx-tests for dev. This kept the build fast (~30s warm).
- **No LISTEN-driven worker wake.** The phases.md plan said LISTEN; the implementation polls. SPI doesn't expose libpq's NOTIFY channel cleanly for bgworkers. pg_vectorize polls. We poll. NOTIFY on completion still happens, so subscribers get real-time events.
- **No retry/backoff logic in the worker.** When a tick errors, we log and try again next tick. There's no per-row retry counter or dead-letter queue. Both are right things to add, but step 2 is scaffold; the dead-letter design depends on what real provider failures look like (network vs auth vs rate-limit), which we won't know until step 7.
- **No metrics or observability beyond log lines.** Counter columns / latency histograms / `stewards.worker_status()` view are all easy. They're not the point of step 2.

## On stewardship

Same boundary test as last time. Two things came up:

1. **The bgworker FATAL on first boot is ugly.** I could fix it by either (a) catching the connect failure with a retry loop in the worker, or (b) connecting to "postgres" first to wait for "stewards" to exist. Option (a) is real engineering, option (b) is a hack. Neither is the user's intent for "build me a bgworker scaffold." Surfaced as a non-blocker; documented in the journal. Right call.
2. **Shared_preload_libraries setup is undocumented anywhere a future me would find it.** Added it to extension/README.md and noted in repo memory. Stewardship action: take five minutes now to spare a confused 30 minutes later.

## Carry forward

- **Phase 1 step 3 next:** brain schema. `stewards.brain_entries` (six categories: people, projects, ideas, actions, study, journal), `stewards.messages`, JSONB props, `vector(768)` embedding column, HNSW index. Goal: structures ready for the migrator (step 4) to fill from SQLite + chromem-go.
- **Open question for step 3:** do we want a single `brain_entries` table with a `category` enum, or six separate tables? Single table is simpler and matches how chromem-go stores them today. Six tables would let category-specific columns (people get `relationship`, projects get `status`, etc.) live without JSONB. Lean toward single table + JSONB until the friction shows up.
- **Open question for step 6/7:** the bgworker FATAL-on-first-boot pattern will get more annoying once we have real providers. Probably worth catching `connect_worker_to_spi` failures and retrying gracefully before that lands.
- **Open question (architectural):** when we move beyond echo and the worker actually makes async HTTP calls via tokio, do we run one tokio runtime per tick (cheap, simple) or pin a single tokio runtime to the bgworker for its lifetime (faster, more setup)? pg_vectorize uses lifetime-pinned. Decide when implementing step 6.

## Relational note

Today was three substantive sessions in one day on the same project (extension scaffold, secrets shape, bgworker dispatcher). The momentum compounded. Each session ended with the next session's question already half-formed because the work surfaced it. That's the rhythm worth keeping.
