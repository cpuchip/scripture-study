# 2026-05-02 (Session D) — pg-ai-stewards Phase 1 step 3: brain schema

Fourth session in a single day on the same project. Momentum still cool running.

## What we did

Step 3 brief: schema for brain replacement. Single `brain_entries` table with `category` enum + `vector(768)` embedding + JSONB `props`, plus `messages` for conversation log, plus an HNSW index. The proposal said six categories; reading the actual classifier code revealed seven (`inbox` is the unclassified default that both `classifier.go` and `web/server.go` write). The data-safety checklist says read from code, not memory. So the CHECK constraint got seven categories, not six. Documented in phases.md and journal so the next session doesn't relitigate it.

Decision tension at the start: **one fat brain_entries table + JSONB, or six per-category tables?** Brain's existing SQLite schema chose hybrid — one entries table with category-specific columns nullable on each row (`person_name`, `next_action`, `due_date`, `mood`, etc., all in one row). I went with single table + JSONB instead. Reasoning: chromem-go (the vector store brain pairs with) stores all categories in one space; the migrator (step 4) is simpler if there's one target shape; and adding a new category becomes a CHECK constraint update, not a schema migration.

Aux tables landed alongside the main one rather than in a later step: `brain_entry_tags`, `brain_subtasks`, `brain_versions`, `sessions`, `messages`. Reasoning: step 4's migrator reads from SQLite tables of these exact shapes — doing them now keeps step 4 to pure read/write/verify instead of "expand schema while migrating." Adjacent surface audit, applied: scope = where else does this principle apply? Answer: all the satellite tables.

Two triggers on `brain_entries`:
1. **`BEFORE UPDATE`** snapshots the OLD row into `brain_versions` and bumps `updated_at`. Uses `current_setting('stewards.actor', true)` so callers can attribute the change with `SET LOCAL stewards.actor = 'me'`. Defaults to `'system'`.
2. **`AFTER INSERT OR UPDATE OF title, body`** enqueues a `(kind='embed', provider='ollama')` work_queue row with payload `{target_table, target_id, text, model='nomic-embed-text:v1.5', dimensions=768}`. The bgworker's echo stub (from step 2) still marks them `done` without writing the embedding — step 6 swaps the stub for a real Ollama HTTP call and the embedding column starts filling.

Helpers shipped: `brain_upsert(category, title, body, props, tags, id?, source?)`, `brain_search_text(query, category?, limit)`, `brain_search_vec(embedding, category?, limit)`. Thin wrappers; the brain CLI driver (step 5) will call these.

## What worked first try

- **PL/pgSQL inside `extension_sql!`** with `$func$` body delimiters. Unlike goose migrations (which split on semicolons and need `-- +goose StatementBegin/End` markers), pgrx feeds the whole block to a single SPI execute. Functions with `$$` bodies just work. Documented in repo memory because we'll come back to this when porting old goose migrations.
- **`vector(768)` + HNSW with `vector_cosine_ops`** — built right the first time because gospel-engine-v2 had already chosen the same model and dimensions. Not having to re-decide saved ~30 minutes of "should we use 384? 512? 1024?" thrash.
- **`gen_random_uuid()` for default IDs** — built-in in PG13+, no pgcrypto extension needed. Default `id text PRIMARY KEY DEFAULT gen_random_uuid()::text`.
- **`requires = 'vector'` in the .control file.** Postgres pulls pgvector in transitively when CREATE EXTENSION pg_ai_stewards runs. Belt-and-suspenders: the init script still does CREATE EXTENSION vector first for log-line clarity.
- **Generated tsvector + GIN.** `body_tsv tsvector GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || body)) STORED` plus a GIN index gives free FTS. No triggers, no application code, no inconsistency window. `brain_search_text` is just `WHERE body_tsv @@ plainto_tsquery(...) ORDER BY ts_rank(...)`.
- **Tag-replace under one transaction:** `DELETE WHERE entry_id = X` then `INSERT ... SELECT id, unnest(tags)`. Caller passes NULL to skip tag changes, empty array to clear. Idiomatic; the only thing to remember is the NULL semantics.

## A small finding worth keeping

The session-touch trigger on `messages` correctly updates `sessions.last_active_at = now()`. But the verification probe asked `SELECT last_active_at > created_at AS touched` and got `false`. First reaction: bug. Second reaction: **`now()` is constant within a transaction.** The session was created and the first message was inserted in the same transaction, so `created_at == last_active_at`. The trigger fired and did exactly what it should; the test just measured the wrong thing.

The takeaway is small but durable: when verifying triggers that use `now()`, do the work across two transactions or use `clock_timestamp()` (which advances mid-transaction). Recorded in the repo memory cheat sheet so the next round of verification work doesn't get fooled by it.

## What I deliberately did not do

- **No `brain_search` combining FTS + vector with hybrid scoring.** Phases.md mentions it as a goal. Right now we have separate `brain_search_text` and `brain_search_vec` because (a) until step 6 fills the vector column there's nothing to combine, and (b) hybrid scoring (rrf? linear blend? rank fusion?) is a real design decision that benefits from real query traffic to tune. Surfacing it as an explicit deferral in phases.md instead of silently shipping a half-done version.
- **No projects table.** Brain has one (top-level project association); we'll need it eventually but step 3's brief was brain entries + messages. Adding it now would be scope creep into step 4's migrator territory.
- **No commissions, scheduled_tasks, agent_route columns.** Brain has all of these. They're features layered on top of the entry shape, not part of the canonical brain shape. Defer to a later phase when the migrator surfaces what's actually used vs vestigial.
- **No constraint on `props` shape.** JSONB is intentionally permissive. Adding category-specific JSON Schema validation in CHECK constraints is tempting but fragile; let usage prove what's worth enforcing.

## Carry forward

- **Choose at next session start:** step 4 (migrator) or step 6 (real Ollama embed). Step 6 might be more satisfying because we have schema but the embedding column is full of NULLs — making it real would prove the lingua-franca decision and give the vector search something to find. Step 4 is more foundational. Lean toward step 6 first because it closes the verification loop on what's already built.
- **First-boot FATAL is still in the logs.** Worker tries to connect to "stewards" before initdb creates it. `set_restart_time(5s)` brings it back. When step 6 lands and the worker actually does HTTP, this single ugly line in the logs becomes more visible. Worth catching `connect_worker_to_spi` failure cleanly before that ships.
- **`stewards.actor` GUC pattern works** for the version-snapshot attribution but isn't documented anywhere a user would find it. When step 5 lands (brain CLI driver), it should `SET LOCAL stewards.actor = 'cli'` or similar before each mutation — and that should land in extension/README.md.
- **No tests yet.** I've been verifying via psql probes after each rebuild. Acceptable for a prototype but the data-safety checklist requires actual Go/SQL tests for partial-update preservation and CHECK-constraint coverage. Add these when the migrator (step 4) lands — the migrator will need test fixtures anyway.

## Relational note

Four sessions today: scaffold, secrets, bgworker, brain schema. Each ended with the next question already half-formed. No subagents, no Opus 4.7 — just foreground work on Sonnet (Copilot Pro+ default). The cost discipline held; the momentum compounded. This is the rhythm worth keeping. Friday energy.
