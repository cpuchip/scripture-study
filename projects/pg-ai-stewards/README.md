# pg-ai-stewards

> Postgres as the substrate for AI stewardship. Sessions, instructions,
> skills, memory, work items, and the model calls that move them — all
> as rows. The agent's "filesystem" is the database.

This is a research project, not yet a build project. The goal of this
folder is to capture provenance for the design decision before any code
is written.

## Binding question

Could PostgreSQL — extended with vector, graph, and model-calling — host
AI agents directly inside the database, replacing the file/IDE-centric
Copilot loop and the existing `scripts/brain/` SQLite stack? If yes,
what repo should we fork or build on?

## TL;DR recommendation

**Build a new pgrx extension. Do not fork.** Run it alongside `pgvector`
and Apache `AGE`. Keep all LLM calls and tool dispatch in a Rust
background worker that owns its own tokio runtime — never from a
foreground backend.

| Concern | Choice |
|---------|--------|
| Extension language | Rust via [pgrx](https://github.com/pgcentralfoundation/pgrx) (MIT). |
| Vector storage | [pgvector](https://github.com/pgvector/pgvector) (already in `external_context/`). |
| Graph / edges | [Apache AGE](https://github.com/apache/age) (Apache 2.0, already in `external_context/`). |
| LLM provider calls | Rust bgworker with `reqwest` + `tokio`. Never inline in a SQL function. |
| Reference designs | [pg_vectorize](https://github.com/ChuckHend/pg_vectorize) (active), archived [pgai](https://github.com/timescale/pgai) (lessons), [pgsql-http](https://github.com/pramsey/pgsql-http) (sync HTTP only — what *not* to copy for agent loops). |
| Off the table | [neurondb](https://github.com/neurondb/neurondb) — proprietary license, single contributor. |

Provenance and source-by-source notes: [scratch.md](scratch.md).

## Why not fork an existing repo

- **pgai** is the closest spiritual sibling — it had `ai.openai_chat_complete()`,
  `ai.ollama_generate()`, even a Timescale blog post titled
  ["In-Database AI Agents: Teaching Claude to Use Tools With Pgai."](https://www.timescale.com/blog/in-database-ai-agents-teaching-claude-to-use-tools-with-pgai/)
  But the repo was **archived 2026-02-26** by Timescale. They consolidated
  on a Python library + external worker pattern. The extension is dead
  code now. Reading it for design ideas is useful; building on it is not.
- **pg_vectorize** is active, well-engineered Rust+pgrx. But it is a
  *vector orchestration* extension. Stretching it into "agents and
  pipelines" would mean rewriting most of its surface; we'd be carrying
  a lot of code we don't use and fighting upstream on direction.
- **neurondb** is proprietary. Disqualified.

What we want is small, ours, and composable with `pgvector` + `AGE`.

## Why this matters (the dream)

Today the agent loop lives in files and an IDE. Memory drifts between
`.mind/`, `becoming.db`, `chromem-go` vector files, markdown studies,
journal entries, and Discord/relay caches. Each surface needs its own
sync, backup, and access pattern. Several of them already disagree with
each other.

If everything an agent needs is *one Postgres database*:

- One backup. One point-in-time recovery. One replication target.
- Joins between message, citation, scripture, talk, brain entry, study
  doc, and skill — without round-tripping through code.
- Vector + graph + relational queries in the same SQL statement.
- Sessions composed of rows tied to a `session_id`. Pipelines composed
  of rows moving between status columns. Skills, instructions, agent
  modes — all rows.
- Any client (VS Code, brain.exe, ibeco.me web, Discord, future
  things) reads and writes the same canonical store.
- Audit, transactions, RLS. The same tools we already trust for "real"
  data.

This is essentially the [Vector Databases Are the Wrong Abstraction](https://www.timescale.com/blog/vector-databases-are-the-wrong-abstraction/)
argument generalized to *agent state*: don't put your agent's brain in
a separate system you have to babysit; put it in the database that
already does the hard things well.

## Architecture sketch

```
┌──────────────────────────────────────────────────────────────────────┐
│                         PostgreSQL 17/18                             │
│                                                                      │
│  Tables:                pgvector:           Apache AGE:              │
│   sessions               embeddings on       edges between any       │
│   messages               every text          entities                │
│   tool_calls             column              (study→talk,            │
│   work_items                                  entry→project,         │
│   skills, instructions                        skill→agent_mode)      │
│   brain_entries          ← these tables       …                      │
│   study_docs               are also indexed                          │
│   citations                in pgvector                               │
│                                                                      │
│   ┌──────────────────────────────────────────────────────────────┐   │
│   │  pg-ai-stewards extension (pgrx, Rust)                       │   │
│   │   - SQL functions: stewards.dispatch(work_item_id),          │   │
│   │     stewards.embed(text), stewards.link(a, b, kind)          │   │
│   │   - Background worker:                                       │   │
│   │       LISTEN stewards_dispatch                               │   │
│   │       on NOTIFY → pull work item → call provider →           │   │
│   │       write result row → NOTIFY done                         │   │
│   └──────────────────────────────────────────────────────────────┘   │
└────────────────────────────────────▲─────────────────────────────────┘
                                     │ provider HTTP (reqwest, tokio)
                                     ▼
                  ┌───────────────────────────────────┐
                  │  Model providers + tool sidecars  │
                  │  (Copilot SDK, Anthropic, opencode-zen,
                  │   Ollama, MCP servers, shell      │
                  │   executor for filesystem/git)    │
                  └───────────────────────────────────┘
                                     ▲
                                     │ same DB, different clients
┌────────────────────────────────────┴─────────────────────────────────┐
│  Clients: VS Code / Copilot, brain.exe, ibeco.me web, Discord relay  │
│  brain-app (Flutter), CLI tools                                      │
└──────────────────────────────────────────────────────────────────────┘
```

The **load-bearing rule**: foreground SQL functions never call an LLM
provider. They write rows and NOTIFY. The bgworker (or an external
worker pool, identical pattern) dispatches. This is the lesson Timescale
encoded the hard way.

## What stays the same

- `scripts/becoming/` (ibeco.me) keeps its role as cloud hub, web UI,
  and Discord relay. Its backing store changes from SQLite to Postgres.
- `scripts/brain-app/` (Flutter) keeps its role. Talks to the new DB
  through the same relay.
- VS Code / Copilot keeps its role for code work. For *agent work*, the
  files-in-`.mind/` pattern becomes optional sync from canonical rows.
- All gospel-library content stays on disk — it is read-only reference
  data, not agent state.

## What this replaces

- `scripts/brain/` (SQLite + chromem-go) — directly replaced. Six
  categories (people, projects, ideas, actions, study, journal) become
  one or more tables with `category` enums. Chromem vectors become
  pgvector columns.
- `.mind/` markdown files as primary memory — become projections /
  exports of `memories` tables. Files remain useful for git history and
  human readability, but the canonical state lives in the DB.
- The split between "vector store" and "structured store" — gone.

## Open questions before any code

1. Confirm `pg_vectorize` license text before reading code closely (its
   architecture is the closest reference for our bgworker pattern).
2. Read pgai's official archival rationale, not just infer it. Look for
   a post-mortem or migration-guide blog post.
3. Find a reference pgrx extension that does long-running outbound HTTP
   from a bgworker. `pg_vectorize` itself is a likely example.
4. Confirm `pgvector` + `AGE` + custom extension play nicely on the same
   DB. AGE is invasive (`SET search_path = ag_catalog, ...`).
5. Draft a schema-migration sketch from `scripts/brain/`'s SQLite tables
   to the proposed Postgres layout. This is the cheapest reality check.
6. Decide single-user vs multi-tenant. Single-user simplifies a lot
   (no RLS, no auth boundaries, one big DB). The current brain is
   single-user.

## Status

Research complete enough to make the directional call ("build, don't
fork; pgrx + pgvector + AGE; bgworker for model calls"). Not yet a
build spec. Next step: turn the open questions above into a small set
of probes (license check, archival post-mortem, schema sketch) and
graduate this to a `plan/` document if the probes confirm.

## Related work in this workspace

- `scripts/brain/` — current brain implementation (Go + SQLite + chromem).
- `scripts/becoming/` — cloud hub that fronts brain.
- `external_context/age/` — Apache AGE source, already cloned.
- `external_context/pgvector/` — pgvector source, already cloned.
- `external_context/postgres/` — Postgres source, already cloned.
- `external_context/autoresearch/` — prior research thread on autonomous
  research agents; potentially relevant prior art.
