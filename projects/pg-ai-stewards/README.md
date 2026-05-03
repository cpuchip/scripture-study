# pg-ai-stewards

> Postgres as the substrate for AI stewardship. Sessions, instructions,
> skills, memory, work items, and the model calls that move them — all
> as rows. The agent's "filesystem" is the database. Studies link to
> canonical sources via cross-DB references to gospel-engine-v2.

This is a research project, not yet a build project. The goal of this
folder is to capture provenance for the design decision before any code
is written.

## Binding question

Could PostgreSQL — extended with vector, graph, and model-calling — host
AI agents directly inside the database, replacing the file/IDE-centric
Copilot loop and the existing `scripts/brain/` SQLite stack? If yes,
what repo should we fork or build on? Could it pair with gospel-engine-v2
so studies can link to canonical sources?

## TL;DR — the verdict got stronger

**Yes, build it. Build a new pgrx extension. Do not fork.** Run it
alongside `pgvector` and Apache `AGE` in a separate database within the
same Postgres cluster as gospel-engine-v2. Keep all LLM calls and tool
dispatch in a Rust background worker that owns its own tokio runtime —
never from a foreground backend.

| Concern | Choice |
|---------|--------|
| Extension language | Rust via [pgrx](https://github.com/pgcentralfoundation/pgrx) (MIT). |
| Vector storage | [pgvector](https://github.com/pgvector/pgvector) (already used by gospel-engine-v2). |
| Graph / edges | [Apache AGE](https://github.com/apache/age) (Apache 2.0). |
| LLM provider calls | Rust bgworker with `reqwest` + `tokio`. Never inline in a SQL function. |
| Reference designs | [pg_vectorize](https://github.com/ChuckHend/pg_vectorize) (active, PostgreSQL license, Rust+pgrx), archived [pgai](https://github.com/timescale/pgai) (lessons), [pgsql-http](https://github.com/pramsey/pgsql-http) (sync HTTP only — what *not* to copy for agent loops). |
| DB topology | Same Postgres cluster as gospel-engine-v2. Two databases: `gospel` (canon, read-mostly) and `stewards` (agent state). Cross-DB references via stable URI strings; `postgres_fdw` if joins are ever needed. |
| Off the table | [neurondb](https://github.com/neurondb/neurondb) — proprietary, single contributor. |

Provenance, every source, and stress-tests: [scratch.md](scratch.md).

## What changed since the first pass

Three things moved the verdict from "viable" to "boring":

1. **gospel-engine-v2 is already on `pgvector/pgvector:pg18`**
   (verified in [docker-compose.local.yml](../../scripts/gospel-engine-v2/docker-compose.local.yml)).
   We don't need a new database — we need a new database in the same
   cluster.
2. **Microsoft blogged the pgvector + Apache AGE combination on
   2026-04-15** as the recommended Azure PostgreSQL pattern for
   "GraphRAG." Same query planner, same executor, same transaction.
   Quote: "Both extensions participate in the same query planner and
   executor. A CTE that calls pgvector's `<=>` operator can feed results
   into a `cypher()` call in the next CTE all within a single
   transaction." Source: [techcommunity.microsoft.com](https://techcommunity.microsoft.com/blog/adforpostgresql/combining-pgvector-and-apache-age---knowledge-graph--semantic-intelligence-in-a-/4508781)
3. **`Haiwen-Yin/memory-pg18-by-yhw`** (Apache 2.0, updated 2 days ago)
   ships the same pattern with the title "AI Agent Memory System with
   PostgreSQL 18 + Apache AGE." Tiny project, but proves the build path.

We are arriving early enough to build something native and late enough
that the load-bearing pieces are battle-tested. That is the right time
to commit.

## Why not fork an existing repo

- **pgai** is the closest spiritual sibling — `ai.openai_chat_complete()`,
  `ai.ollama_generate()`, even a Timescale post titled
  ["In-Database AI Agents: Teaching Claude to Use Tools With Pgai."](https://www.timescale.com/blog/in-database-ai-agents-teaching-claude-to-use-tools-with-pgai/)
  But the repo was **archived 2026-02-26** by Timescale. They consolidated
  on a Python library + external worker pattern. The extension is dead
  code now. Reading it for design ideas is useful; building on it is not.
  - Best inferred reason: Timescale sells managed Postgres; managed-DB
    customers can't install custom extensions. That constraint doesn't
    bind us. We are self-hosted single-user.
- **pg_vectorize** is active, well-engineered Rust+pgrx, PostgreSQL
  license. But it is a *vector orchestration* extension. Stretching it
  into agents-and-pipelines means rewriting most of its surface; we'd
  carry a lot of code we don't use and fight upstream on direction.
  Better to study its bgworker pattern and write our own.
- **neurondb** is proprietary. Disqualified.

What we want is small, ours, and composable.

## Pairing with gospel-engine-v2

`gospel-engine-v2` already runs on Postgres + pgvector + pg_trgm with
schemas for `scriptures`, `chapters`, and `talks`. The "link studies to
canonical sources" use case Michael wants is therefore not an
integration project — it's a database-topology choice.

**Recommendation: same cluster, two databases.**

```
postgres cluster
├── db: gospel        (gospel-engine-v2 owns)
│   ├── extensions: vector, pg_trgm
│   └── tables: scriptures, chapters, talks
│
└── db: stewards      (pg-ai-stewards owns)
    ├── extensions: vector, age, pg_ai_stewards (ours)
    ├── schemas: stewards.*, ag_catalog.* (AGE), public
    └── tables: sessions, work_items, brain_entries, studies, ...
```

Cross-DB references travel as **stable URI strings**, e.g.
`lds://scripture/bofm/1-ne/3.7` or `lds://talk/2024/04/oaks-loving-the-lord`.
AGE edges hold them as properties; the Rust agent resolves them by
calling gospel-engine-v2's HTTP API (already working) or via
`postgres_fdw` if we want SQL-level joins.

This isolates failure modes: a schema change in gospel-engine-v2 cannot
break stewards; a `pg_dump` of stewards doesn't pull in the entire
canon. Backups, recovery, role boundaries all stay clean.

When the agent writes a study, it can:

```sql
-- find me scriptures semantically near this draft paragraph
SELECT s.reference, s.text, 1 - (s.embedding <=> $1) AS similarity
FROM gospel.scriptures s
WHERE s.embedding IS NOT NULL
ORDER BY similarity DESC LIMIT 5;
```

…even when `gospel.scriptures` lives in another database, by going
through the gospel-engine-v2 API or postgres_fdw.

When linking is established, the AGE edge in stewards records the
relationship:

```cypher
MATCH (study:Study {id: $study_id})
CREATE (study)-[:CITES {reason: 'core text', confidence: 0.92}]
       ->(:Scripture {ref: 'lds://scripture/bofm/1-ne/3.7'})
```

The `Scripture` node is a stub in stewards' AGE graph; the actual verse
text lives in gospel. Best of both worlds — joinable graph in stewards,
canonical source in gospel.

## Why this matters (the dream)

Today the agent loop lives in files and an IDE. Memory drifts between
`.mind/`, `becoming.db`, chromem-go vector files, markdown studies,
journal entries, and Discord/relay caches. Each surface needs its own
sync, backup, and access pattern. Several already disagree with each
other.

If everything an agent needs is *one Postgres cluster*:

- One backup. One point-in-time recovery. One replication target.
- Joins between message, citation, scripture, talk, brain entry, study
  doc, and skill — without round-tripping through code.
- Vector + graph + relational queries in the same SQL statement (per
  Microsoft's bridge pattern).
- Sessions composed of rows tied to a `session_id`. Pipelines composed
  of rows moving between status columns. Skills, instructions, agent
  modes — all rows.
- Any client (VS Code, brain.exe, ibeco.me web, Discord, future things)
  reads and writes the same canonical store.
- Audit, transactions, RLS. The same tools we already trust for "real"
  data.

This is essentially the [Vector Databases Are the Wrong Abstraction](https://www.timescale.com/blog/vector-databases-are-the-wrong-abstraction/)
argument generalized to *agent state*: don't put your agent's brain in
a separate system you have to babysit; put it in the database that
already does the hard things well.

## Architecture sketch

```
┌────────────────────────────────────────────────────────────────────────┐
│                       PostgreSQL 18 cluster                            │
│                                                                        │
│  ┌─── db: gospel ────────────────┐  ┌─── db: stewards ──────────────┐  │
│  │ extensions: vector, pg_trgm   │  │ extensions: vector, age,      │  │
│  │ tables: scriptures, chapters, │  │   pg_ai_stewards (ours)       │  │
│  │   talks                       │  │ schemas:                      │  │
│  │ owned by: gospel-engine-v2    │  │   stewards.*  (rows)          │  │
│  │ access: read-mostly           │  │   ag_catalog.* (AGE graph)    │  │
│  └───────────────┬───────────────┘  │ tables:                       │  │
│                  │                  │   sessions, messages,         │  │
│                  │ postgres_fdw     │   tool_calls, work_items,     │  │
│                  │ (read-only view) │   skills, instructions,       │  │
│                  └─────────────────►│   brain_entries, studies      │  │
│                                     │                               │  │
│                                     │ ┌──────── bgworker ────────┐  │  │
│                                     │ │ Rust + tokio + reqwest   │  │  │
│                                     │ │ LISTEN stewards_dispatch │  │  │
│                                     │ │ → call provider          │  │  │
│                                     │ │ → write result row       │  │  │
│                                     │ │ → NOTIFY done            │  │  │
│                                     │ └──────────────────────────┘  │  │
│                                     └──────────────┬────────────────┘  │
└──────────────────────────────────────────────────┬─┴───────────────────┘
                                                   │ provider HTTP
                                                   │ (Anthropic, Copilot
                                                   │  SDK, opencode-zen,
                                                   │  Ollama)
                                                   ▼
                            ┌──────────────────────────────┐
                            │  Tool sidecars               │
                            │   docker-exec for filesystem │
                            │   git/gh for repo ops        │
                            │   shell for build/test       │
                            │   MCP servers                │
                            └──────────────────────────────┘
                                                   ▲
                                                   │ same DB, many clients
┌──────────────────────────────────────────────────┴─────────────────────┐
│  Clients                                                               │
│   VS Code / Copilot — pair-programming surface                         │
│   Web UI (in becoming/ or sibling) — see worklist, steer, interrupt    │
│   brain.exe / Flutter — mobile and desktop                             │
│   Discord relay (becoming) — async chat client                         │
│   CLI tools                                                            │
└────────────────────────────────────────────────────────────────────────┘
```

The **load-bearing rule**: foreground SQL functions never call an LLM
provider. They write rows and `NOTIFY`. The bgworker (or an external
worker pool, identical pattern) dispatches. This is the lesson Timescale
encoded the hard way.

## Replacing VS Code + Copilot — what's honest

**Substrate yes, editor no.**

The DB can absorb agent state, memory, brain entries, study drafts,
conversation history, pipelines, scheduled work, cross-references. All
the things today scattered between `.mind/` files, SQLite, chromem,
markdown, and Discord caches.

The DB cannot edit files for you. A Docker sidecar with a writable repo
checkout edits files; the DB orchestrates. Git plumbing happens via that
sidecar (or `gh` CLI). Long builds run in a container, not in a backend.

VS Code stays — it remains the best editor we have, and we should not
write a code editor. Its role shifts from "the place where work
happens" to "one of the surfaces where work shows up." Copilot becomes
a client of the steward DB rather than the controller of agent state.

What replaces the Copilot-as-only-loop UX is a **web UI** on top of the
stewards DB where Michael:

- Sees the agent's worklist (what's queued, what's running, what's done)
- Promotes / demotes / cancels items
- Sees in-flight model calls and interrupts them
- Reviews completed work before merge
- Triggers new agent runs from any device

Long-running agents pull from the worklist, do the next step, write the
result, and `NOTIFY` for review — without needing an open VS Code window
to hold their context.

This is what Michael described: "automated agents with their work in the
DB itself, using docker for fs work and actual coding and building, and
git/github for repo management." The dream is exactly right; the
specific clarification is that VS Code remains for the moments Michael
*wants* to be in the loop.

## What stays the same

- `scripts/becoming/` (ibeco.me) keeps its role as cloud hub, web UI,
  and Discord relay. Backing store changes from SQLite to Postgres.
- `scripts/brain-app/` (Flutter) keeps its role. Talks to the new DB
  through the same relay.
- `scripts/gospel-engine-v2/` keeps its role. Pulled into the same
  cluster but its DB and ownership stay independent.
- VS Code / Copilot keeps its editor role. Becomes one client among
  many for agent work.
- `.mind/` markdown files remain useful for git history and human
  readability — but they become *projections* of canonical rows, not
  the canonical store.
- All gospel-library content stays on disk — it is read-only reference
  data, indexed by gospel-engine-v2.

## What this replaces

- `scripts/brain/` (SQLite + chromem-go) — directly replaced. Six
  categories (people, projects, ideas, actions, study, journal) become
  one or more tables with `category` enums. Chromem vectors become
  pgvector columns.
- `.mind/` markdown files as primary memory — become projections /
  exports of `memories` tables. Files remain useful but the canonical
  state lives in the DB.
- The split between "vector store" and "structured store" — gone.
- The implicit "agent state lives in whatever is loaded into Copilot's
  context window" — replaced with explicit, queryable rows.

## Open questions before any code

1. **Read pg_vectorize's bgworker source.** How does it own a tokio
   runtime? Where does it handle backend signals, cancellation,
   shutdown? Cloned at [external_context/pg_vectorize/](../../external_context/pg_vectorize/).
2. **Read pgai's archival rationale** (or the Timescale blog) for the
   official reason — not just my inference. Cloned at
   [external_context/pgai/](../../external_context/pgai/).
3. **Stand up a docker-compose locally** with `pgvector/pgvector:pg18` +
   `apache/age` installed in the same image. `CREATE EXTENSION vector;
   CREATE EXTENSION age;` in one DB. Run Microsoft's bridge example
   end-to-end. Cheapest reality check before any spec writing.
4. **Confirm `postgres_fdw` plays nicely with `vector` and `agtype`.**
   Vector is a regular composite type (likely fine). Agtype is custom
   (unclear). If it doesn't, fall back to gospel-engine-v2 HTTP API
   for cross-DB lookups — also fine.
5. **Embedding model.** Gospel-engine-v2 uses `nomic-embed-text-v1.5`
   (768d). Stewards should match unless there's a reason not to —
   same embedding space lets us compare a study draft directly to a
   scripture verse, across the DB seam.
6. **Sketch the SQLite→Postgres migration** for `scripts/brain/`'s
   six categories. If the schema sketch is awkward, the abstraction
   is wrong.
7. **Single-user vs multi-tenant.** Single-user simplifies a lot. The
   current brain is single-user. If we ever want ibeco.me to host
   multiple users, RLS handles it later — design for single-user now.
8. **Decide the web-UI surface.** Is it part of `becoming/` (existing
   cloud hub), a sibling app, or a brand-new Flutter view in
   `brain-app/`? Probably the first.

## Status

**Phase 1 in flight — step 1 (extension scaffold) complete 2026-05-02.**

- pgrx 0.18.0 extension `pg_ai_stewards` builds in Docker and loads
  into PG18 alongside pgvector and Apache AGE. `stewards.version()`
  returns `0.1.0` end-to-end. Lives at [extension/](extension/).
- Probe stack still passing — see [probe/RESULTS.md](probe/RESULTS.md).
- Direction (proposal + phases) shipped 2026-05-02 — see
  [proposal.md](proposal.md), [phases.md](phases.md).

Next: bgworker scaffold + brain schema (Phase 1 steps 2–3 in
[phases.md](phases.md)).

Read next:

- **[proposal.md](proposal.md)** — what we're building and why
- **[phases.md](phases.md)** — Phase 1 → Phase 5+ delivery plan
- **[extension/](extension/)** — the pgrx extension itself
- **[probe/](probe/)** — the original feasibility proof
- **[scratch.md](scratch.md)** — full source provenance

## Related work in this workspace

- `scripts/brain/` — current brain implementation (Go + SQLite + chromem).
- `scripts/becoming/` — cloud hub that fronts brain.
- `scripts/gospel-engine-v2/` — already on Postgres+pgvector. Pairs cleanly.
- [external_context/age/](../../external_context/age/) — Apache AGE source.
- [external_context/pgvector/](../../external_context/pgvector/) — pgvector source.
- [external_context/postgres/](../../external_context/postgres/) — Postgres source.
- [external_context/pg_vectorize/](../../external_context/pg_vectorize/) — best Rust+pgrx reference (cloned this session).
- [external_context/pgai/](../../external_context/pgai/) — archived but instructive (cloned this session).
- [external_context/pgrx/](../../external_context/pgrx/) — extension framework + bgworker example (cloned this session).
- [external_context/pgsql-http/](../../external_context/pgsql-http/) — sync HTTP fallback (cloned this session).
- `external_context/autoresearch/` — prior research on autonomous research agents.
