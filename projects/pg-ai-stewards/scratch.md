# pg-ai-stewards — research scratch

Working notes during source triage. Permanent provenance — do not delete.

## Binding question

> Could PostgreSQL — extended with vector, graph, and model-calling — host
> AI agents directly inside the database, replacing the file/IDE-centric
> Copilot loop and the existing brain/becoming SQLite stack? If yes, what
> repo should we fork or build on?

## Source triage

### timescale/pgai — https://github.com/timescale/pgai
- **License:** PostgreSQL license (permissive).
- **Status: ARCHIVED 2026-02-26 by owner.** Read-only. Big signal.
- Two parts:
  - Python library + `vectorizer-worker` (out-of-DB Python process) — this
    is what Timescale doubled down on.
  - PG extension under `projects/extension` (PL/pgSQL + Python in `ai`
    schema). Functions like `ai.openai_chat_complete()`,
    `ai.ollama_generate()`, `ai.ollama_chat_complete()` — model calling
    directly from SQL. This is the "in-database agents" pattern Michael
    is dreaming about.
- Their own blog post: ["In-Database AI Agents: Teaching Claude to Use
  Tools With Pgai"](https://www.timescale.com/blog/in-database-ai-agents-teaching-claude-to-use-tools-with-pgai/) —
  literal precedent for the dream.
- Why archived (inferred from pattern): the extension model called LLM
  endpoints from inside backend processes, which holds connections,
  blocks transactions, and breaks under managed-Postgres environments
  that don't allow custom extensions. Timescale moved to the worker
  model where the DB owns *state* and an external process owns *the
  network call*.
- **Use as: prior art and design lessons. Do not adopt as base.**

### ChuckHend/pg_vectorize — https://github.com/ChuckHend/pg_vectorize
- **License:** present in repo, did not verify exact text. README links
  to PGXN — likely PostgreSQL or Apache 2.0. Verify before forking.
- **Active.** v0.26.2 released 5 days ago, PG18 support added.
- **Rust + pgrx** extension, plus an HTTP server alternative for managed
  Postgres. ~94% Rust.
- API surface: `vectorize.table()` to register a table for embedding,
  `vectorize.search()` for semantic + FTS, `vectorize.rag()` for
  end-to-end RAG, `vectorize.generate()` for raw text gen,
  `vectorize.encode()` for raw embedding.
- Background worker pattern via `pgmq` (their own queue extension).
  `schedule => 'realtime'` adds triggers; `cron`-style schedules also
  supported.
- Calls out to OpenAI, Ollama, Hugging Face Sentence-Transformers (via
  their `vector-serve` Python container).
- **Use as: candidate base, OR as model for our own pgrx extension.**

### neurondb/neurondb — https://github.com/neurondb/neurondb
- **License: proprietary.** Disqualified from "fork as base."
- C extension. Vector + ML + GPU. Big surface, single contributor.
- Feature dump reads like marketing — 650+ SQL functions, ~47 stars,
  10 forks, 1 contributor.
- **Use as: do not adopt. Skim feature list for ideas only.**

### pgcentralfoundation/pgrx — https://github.com/pgcentralfoundation/pgrx
- **License: MIT.** Active. PG13–18. Rust framework for PG extensions.
- The de-facto answer for "how do we build a PG extension in 2026."
- `cargo pgrx new`, `cargo pgrx run`, `cargo pgrx test`, multi-version
  support, safe Datum mapping, SPI access, bgworker support, custom
  types, enums, triggers, hooks.
- **Caveats:** threading not supported; async story unexplored;
  pre-1.0 (breaking changes possible).
- **Use as: foundation. Whatever we build, build with this.**

### apache/age — https://github.com/apache/age
- **License: Apache 2.0.** Active. PG11–18. ~4.5k stars, 100+ contributors.
- Bitnine-derived. Cypher (openCypher) on Postgres, hybrid SQL+Cypher,
  multi-graph per DB, indexes on vertices and edges.
- Written in C against PG internals (not pgrx). 70% C, 6% Python (driver),
  6% PL/pgSQL.
- Pairs cleanly with pgvector — they live in different schemas and don't
  collide. AGE for relationships, pgvector for similarity.
- **Use as: companion extension, not base.** Install alongside our
  extension to get graph/edge support.

### pramsey/pgsql-http — https://github.com/pramsey/pgsql-http
- **License: MIT.** Active. C, libcurl-based. ~1.6k stars.
- Synchronous `http_get`, `http_post`, etc. from SQL.
- The README literally has a section called "Why This is a Bad Idea"
  warning about backend blocking. The To-Do mentions "background worker
  support could be used to set up an HTTP request queue."
- Pattern: pair with `pg_background` or `pg_cron` for async.
- **Use as: existence proof.** If we want SQL-callable HTTP without
  writing it ourselves, this exists. But for agent loops we want
  bgworker-owned reqwest, not a SQL function.

## Synthesis (rough)

The dream is coherent. The pieces exist. But the *interesting* failure mode
is what Timescale did: they tried "agents inside the extension" and pulled
back to "DB owns state, worker owns network." That doesn't kill the dream
— it sharpens what "inside the DB" should mean.

A workable layering:

| Layer | What it does | Where it runs |
|-------|--------------|---------------|
| Tables | Sessions, instructions, skills, messages, tool calls, work items | Postgres rows |
| Vectors | Embeddings of all of the above | pgvector columns |
| Graph | Edges between entries (links, citations, references) | Apache AGE |
| Model calls | LLM provider HTTP, tool dispatch | Background worker (Rust, pgrx bgworker, owns reqwest + tokio) |
| Pipelines | "Move work item from triage → research → planning → done" | SQL state machine + LISTEN/NOTIFY + bgworker dispatch |
| Tool execution | Filesystem, git, shell, MCP | Sidecar process; bgworker is *dispatcher*, not executor |

The agent never calls an LLM from a foreground backend. It writes a
"please-think" row; the bgworker reads it, calls the model, writes
results back. That gives transactions, retries, observability, and
cancellation for free.

VS Code / files don't go away — they become *one client* among many
that read/write the same Postgres rows. Same for the brain.exe Discord
relay. Same for a future web UI.

## Connections to existing work

- `scripts/brain/` — the SQLite + chromem-go brain is the obvious thing
  this would replace. README at `scripts/brain/README.md` already names
  six categories (people, projects, ideas, actions, study, journal) and
  the building-block model (Dropbox / Sorter / Form / Filing Cabinet /
  Receipt / Bouncer / Tap-on-Shoulder / Fix Button / Search). All of
  that maps cleanly onto Postgres tables + bgworker dispatch.
- `scripts/becoming/` — the cloud hub already runs on Go and connects
  brain to web/Discord. That role doesn't change; it just talks to a
  different store.
- `external_context/age/`, `external_context/pgvector/`,
  `external_context/postgres/` — repos already cloned for reference.
- `external_context/autoresearch/` — there's existing research thread
  on autonomous research agents; may be relevant prior art.

## Posture check

Am I confirming the dream or actually testing it? The Timescale archival
is a real disconfirming signal that I tried to take seriously. My read:
they archived the extension because they're a vendor selling managed
Postgres and the extension surface was too operationally fragile *for
their customers*. For a self-hosted, single-user, ibeco.me-style stack
where Michael owns the box, the extension model is more viable than it
was for them.

But: have not actually read pgai's archival announcement or any
post-mortem. Should fetch that before final recommendation.

## Open questions

- pg_vectorize license — confirm exact license before any fork.
- pgai archival rationale — find the official statement, not just infer.
- Background-worker patterns in pgrx — find a reference extension that
  does long-running outbound HTTP from a bgworker. (pg_vectorize itself
  is a likely example.)
- AGE + pgvector + custom extension on same DB — any known conflicts?
- Migration path from current SQLite brain — schema mapping draft.
