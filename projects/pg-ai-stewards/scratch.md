# pg-ai-stewards — research scratch

Working notes during source triage. Permanent provenance — do not delete.

## Binding question

> Could PostgreSQL — extended with vector, graph, and model-calling — host
> AI agents directly inside the database, replacing the file/IDE-centric
> Copilot loop and the existing brain/becoming SQLite stack? If yes, what
> repo should we fork or build on? Could it pair with gospel-engine-v2
> so studies can link to canonical sources?

## Current-stack reality check (verified 2026-05-02)

- [scripts/gospel-engine-v2/docker-compose.local.yml](../../scripts/gospel-engine-v2/docker-compose.local.yml) — already runs on `pgvector/pgvector:pg18`.
- [scripts/gospel-engine-v2/internal/db/migrations/001_schema.sql](../../scripts/gospel-engine-v2/internal/db/migrations/001_schema.sql) — uses `vector`, `pg_trgm`, `tsvector` GIN. Already a hybrid lexical + semantic engine.
- Tables: `scriptures`, `chapters`, `talks`. The shape we want stewards
  to *link to* is already there.
- This is huge: pairing pg-ai-stewards with gospel-engine-v2 is not a
  "future integration." It's "add another database to the same Postgres
  cluster, or another extension to the same image." See "DB topology"
  below.

## Source triage

### timescale/pgai
- **License:** PostgreSQL license (permissive).
- **Status: ARCHIVED 2026-02-26 by owner.** Read-only. Big signal.
- Two parts:
  - Python library + `vectorizer-worker` (out-of-DB Python process) — what
    Timescale doubled down on.
  - PG extension under `projects/extension` (PL/pgSQL + Python in `ai`
    schema). Functions like `ai.openai_chat_complete()`,
    `ai.ollama_generate()`, `ai.ollama_chat_complete()` — model calling
    directly from SQL. The literal "in-database agents" pattern.
- Their own blog post: ["In-Database AI Agents: Teaching Claude to Use
  Tools With Pgai"](https://www.timescale.com/blog/in-database-ai-agents-teaching-claude-to-use-tools-with-pgai/)
  — precedent for the dream.
- Final extension release `extension-0.11.2` 2025-10-14; final Python
  lib `pgai-v0.12.1` 2025-10-13; archived 2026-02-26. About four months
  of Python-only releases before the final shutter.
- **Confirmed archival rationale** (verified 2026-05-02 via the
  archived repo's README and corroborating reporting on the
  Timescale → TigerData rebrand from June 2025):
  - Timescale rebranded to **TigerData** in June 2025 with a new
    "Agentic Postgres" product line. The pgai repo's own front-page
    description still markets the *Python library + vectorizer-worker*
    pattern as the recommended path; the in-extension `ai.*` SQL
    functions had become a side branch.
  - The PG extension was archived 2026-02-26 because the company
    consolidated on the outside-worker pattern. *Not* because the
    in-DB direction failed — because the worker pattern won inside
    Timescale's product strategy as part of the rebrand.
  - Timescale is a managed-DB vendor. Many of their customers can't
    install custom extensions on RDS/Aurora/Cloud SQL. The Python
    library serves them; the extension didn't.
- **For us the signal is *confirming*, not disconfirming.** We are
  self-hosted, single-user. We *can* install extensions. And the
  architectural lesson Timescale paid tuition for — never call
  providers from foreground backends, always own a worker — is what
  we already planned to do. Read pgai's code for design lessons; do
  not fork.
- Cloned to [external_context/pgai/](../../external_context/pgai/).

### microsoft/graphrag — https://github.com/microsoft/graphrag
- **License:** MIT. **Active.** v3.0.9 released ~3 weeks ago. 32.7k
  stars. Microsoft Research project that became a real product.
- *Not* a Postgres extension. It's a Python data pipeline that
  extracts a knowledge graph from unstructured text using an LLM,
  performs Leiden hierarchical community detection, generates
  community summaries, and provides three query modes:
  - **Global Search** — answers "what are the themes across the
    whole corpus" questions using community summaries.
  - **Local Search** — answers "tell me about entity X" questions by
    fanning out from a node.
  - **DRIFT Search** — local search with added community context.
- Useful as a *complementary* tool, not a substrate. We could run it
  against the scripture + talk corpus (in `gospel`) and write the
  resulting community summaries into AGE. That's Phase 4+ optional;
  not Phase 1.
- Warning from the README: indexing is expensive. "start small."

### Azure-Samples/PostgreSQL-graphRAG-docker — https://github.com/Azure-Samples/PostgreSQL-graphRAG-docker
- **License: MIT.** Active (last commit ~2 months ago, v2 was Mar 2026).
- **This is the closest reference design we've found.** Single Docker
  image: Postgres 16 + Apache AGE + pgvector + GraphRAG 3.0.5 +
  Microsoft Agent Framework + an MCP server exposing five tools to
  agents:
  - `graphrag_search` — runtime-tunable local/global search
  - `age_get_schema_cached` — schema introspection
  - `age_entity_lookup` — substring entity discovery ("who is X?")
  - `age_cypher_query` — execute user-provided Cypher
  - `age_nl2cypher_query` — natural language → Cypher → execute
  - Plus a router agent that decides which to invoke per question.
- Authored by Helen Zeng (helenzusa1), Microsoft. Companion blog
  post: [techcommunity.microsoft.com (Mar 2026)](https://techcommunity.microsoft.com/blog/adforpostgresql/graphrag-and-postgresql-integration-in-docker-with-cypher-query-and-ai-agents-ve/4503586).
- **Caveats** (read the docker-compose.yaml — verified 2026-05-02):
  - Volumes are hardcoded to `/mnt/c/Users/helenzeng/...` WSL paths.
    Not runnable as-is on Windows.
  - Tightly coupled to Azure OpenAI for embeddings + LLM. Swappable
    in principle but requires editing.
  - Postgres 16, not 18. AGE branch `release/PG16/1.5.0`. Older
    pgvector (`v0.7.4`).
  - Single-container architecture (Postgres + Python + GraphRAG +
    MCP all in one). Nice for demo, wrong for production.
- **Use as: tool-shape blueprint for the MCP server we expose in
  Phase 3.** The five-tool router agent pattern is gold and we
  should mostly copy it. Cloned to
  [external_context/PostgreSQL-graphRAG-docker/](../../external_context/PostgreSQL-graphRAG-docker/).

### ChuckHend/pg_vectorize
- **License: PostgreSQL License** (Tembo, 2023). Confirmed 2026-05-02
  from user-supplied LICENSE text. Permissive. Compatible with anything.
- **Active.** v0.26.2 released 5 days ago, PG18 support added.
- **Rust + pgrx** extension, plus an HTTP server alternative for managed
  Postgres. ~94% Rust.
- API surface: `vectorize.table()` to register a table for embedding,
  `vectorize.search()` for semantic + FTS, `vectorize.rag()` for
  end-to-end RAG, `vectorize.generate()` for raw text gen,
  `vectorize.encode()` for raw embedding.
- Background worker pattern via `pgmq` (their own queue extension).
  `schedule => 'realtime'` adds triggers; cron-style schedules also
  supported.
- Calls out to OpenAI, Ollama, Hugging Face Sentence-Transformers (via
  their `vector-serve` Python container).
- **Use as: best reference design for the bgworker + LLM + embedding
  pipeline pattern in Rust.** Cloned to [external_context/pg_vectorize/](../../external_context/pg_vectorize/).

### neurondb/neurondb
- **License: proprietary.** Disqualified from "fork as base."
- C extension. Vector + ML + GPU. Big surface, single contributor,
  marketing-flavored README.
- Did not clone.

### pgcentralfoundation/pgrx
- **License: MIT.** Active. PG13–18. Rust framework for PG extensions.
- The de-facto answer for "how do we build a PG extension in 2026."
- `cargo pgrx new`, `cargo pgrx run`, `cargo pgrx test`, multi-version
  support, safe Datum mapping, SPI access, bgworker support.
- bgworker example exists at `pgrx-examples/bgworker` — uses SPI in
  transactions. Reqwest + tokio in a bgworker is not a pgrx restriction
  (the threading caveat applies to *backends*, not *workers*); pg_vectorize
  itself proves this works in production.
- **Caveats:** threading not supported in foreground backends; async
  story unexplored there; pre-1.0.
- Cloned to [external_context/pgrx/](../../external_context/pgrx/).

### apache/age
- **License: Apache 2.0.** Active. PG11–18.
- Already in [external_context/age/](../../external_context/age/).
- Pairs cleanly with pgvector — verified in production patterns (see
  Microsoft blog below).
- PG18 setup gotcha: `SELECT create_graph('name'::name)` — the `::name`
  cast is required (per memory-pg18-by-yhw README, 2026-04-30).

### pramsey/pgsql-http
- **License: MIT.** Active. C, libcurl-based.
- Synchronous `http_get`, `http_post`. README literally has a section
  titled "Why This is a Bad Idea" warning about backend blocking.
- Use only as fallback for trivial outbound calls; the production path
  is bgworker-owned reqwest.
- Cloned to [external_context/pgsql-http/](../../external_context/pgsql-http/).

## Validation: combining pgvector + Apache AGE in production

**Microsoft blog, April 15, 2026** — ["Combining pgvector and Apache AGE
— knowledge graph & semantic intelligence in a single engine"](https://techcommunity.microsoft.com/blog/adforpostgresql/combining-pgvector-and-apache-age---knowledge-graph--semantic-intelligence-in-a-/4508781)
(author: Raunak, Microsoft, MS Tech Community PostgreSQL board).

Verified quotes worth keeping:

> "Both extensions participate in the same query planner and executor.
> A CTE that calls pgvector's <=> operator can feed results into a
> cypher() call in the next CTE all within a single transaction,
> sharing all available processes and control the database has to
> offer."

> "The planner assigns cost estimates to index scan for both execution
> engines using the same cost framework it uses for B-tree lookups and
> sequential scans."

> "One backup strategy. One monitoring stack. One connection pool. One
> failover target. One set of credentials."

The "bridge" pattern they show: compute cosine similarity with pgvector,
write the result as a `SIMILAR_TO` edge in AGE with `score` and `method`
properties. After that, Cypher patterns can mix structural and semantic
edges in a single `MATCH`.

This is not theoretical. Microsoft is documenting it as the way to use
their managed Azure PostgreSQL for "GraphRAG." That promotes the pattern
from "interesting" to "boring infrastructure," which is exactly where
bedrock decisions should live.

## Working prior art (small, but exists)

- **Haiwen-Yin/memory-pg18-by-yhw** (Apache 2.0, updated 2 days ago).
  README title: "AI Agent Memory System with PostgreSQL 18 + Apache AGE."
  Schema sketch (verified from README):
  ```sql
  CREATE TABLE memory.concepts (
    concept_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(256) NOT NULL,
    category VARCHAR(128),
    description TEXT,
    content JSONB DEFAULT '{}'::jsonb,
    embedding VECTOR(1024),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_concepts_embedding ON memory.concepts
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 200);
  ```
  Plus a relational `memory.relations` table *and* AGE Cypher edges. The
  pattern: relational + vector for the row, AGE for traversable graph.
  Vector dim 1024 is for BGE-M3; we'd match gospel-engine-v2's
  `nomic-embed-text-v1.5` (768d).
  - Reported perf: <30ms semantic search on <1k records, <100ms multi-hop
    traversal. Tiny dataset; take with salt.
  - Critical PG18 + AGE setup gotcha: `SELECT create_graph('name'::name)`.

- **veloper/pgmcp** (Python, 1 star, 8 months stale, no license file
  visible) — but the architecture is the *exact shape Michael described*.
  An MCP server (FastMCP) exposing pgvector + Apache AGE + pgsql-http to
  AI agents. Sub-servers for knowledge base, web crawl (Scrapy), and raw
  PSQL. One of its tools is literally `http_request` "using the pg_http
  extension." Worth reading the code; not worth depending on.

- Multiple Docker images shipping pgvector + AGE side-by-side already:
  `sohamthakurdesai/postgres-age-pgvector`, `kestutis-katilius/PG18-PGVector-ApacheAGE`,
  `mayflower/pg4ai`, `groenewt/testpg`. Not all maintained; the recipe
  is well-known.

- **codeberg.org/trisolar.faculty/postgres_pgvector_age_benchmarking** —
  benchmarks Postgres+AGE+pgvector vs Neo4j vs OpenSearch. Did not deep-read.

## Synthesis

The dream is not just coherent — it's becoming standard. Microsoft is
documenting the same architecture for enterprise Azure customers, and
small AI-memory projects are shipping the pattern. We are not
pioneering; we are arriving early enough to build something native, and
late enough that the load-bearing pieces are battle-tested.

A workable layering:

| Layer | What it does | Where it runs |
|-------|--------------|---------------|
| Tables | Sessions, instructions, skills, messages, tool calls, work items, brain entries | Postgres rows in `stewards.*` |
| Vectors | Embeddings of all of the above | pgvector columns + HNSW indexes |
| Graph | Edges between entries (links, citations, references, scripture refs) | Apache AGE `stewards_graph` |
| Model calls | LLM provider HTTP, tool dispatch | Background worker (Rust, pgrx bgworker, owns reqwest + tokio) |
| Pipelines | Move work item from triage → research → planning → done | SQL state machine + LISTEN/NOTIFY + bgworker dispatch |
| Tool execution | Filesystem, git, shell, MCP | Sidecar process; bgworker is dispatcher, not executor |
| Canon access | Scripture and conference-talk lookups | Same Postgres cluster, separate DB; access via gospel-engine-v2's existing API or via `postgres_fdw` |

The agent never calls an LLM from a foreground backend. It writes a
"please-think" row; the bgworker reads it, calls the model, writes
results back. That gives transactions, retries, observability, and
cancellation for free.

## DB topology — single DB, two DBs, or two clusters?

Three options:

**A. Same Postgres cluster, same database, separate schemas**
- `gospel.*`, `stewards.*`, `ag_catalog.*` (AGE) all in one DB.
- Pros: one connection, free joins, AGE edges can directly reference
  `gospel.scriptures.id` as a property. Simplest.
- Cons: gospel-engine-v2 owns its migrations; mixing concerns blurs
  ownership. A `pg_dump` of "just the agent" is harder.

**B. Same cluster, two databases (`gospel`, `stewards`) — Michael's preference**
- Pros: clean ownership, separate backups, separate roles. Each DB
  installs only the extensions it needs.
- Cons: cross-DB queries require `postgres_fdw` (foreign data wrapper).
  AGE edges in stewards reference scripture refs as strings/URIs, not
  foreign keys.

**C. Two clusters**
- Overkill for single-user. Skip.

**Recommendation: B with string-URI references.** AGE edges hold scripture
references as a property like `{ref: 'lds://scripture/bofm/1-ne/3.7'}`.
When the agent needs the actual verse text, it queries gospel-engine-v2
via its existing HTTP API (already working) or via `postgres_fdw`
exposing a read-only view of `gospel.scriptures`. No schema coupling.
Schema changes in gospel-engine-v2 don't break stewards. Stewards can
be backed up, restored, blown away, rebuilt without touching canon.

The linkage idea Michael wants — "studies can reference scriptures and
talks" — works fine across this seam. AGE doesn't care that the target
of an edge lives in another DB; the edge is a row in `ag_catalog` whose
property is a string.

## Replacing VS Code + Copilot — honest stress test

Michael's dream: "I can see it replacing vscode + github copilot for me,
and have automated agents with their work in the DB itself."

**What the DB can absorb:**
- Agent state (sessions, instructions, skills, plans, todos, work items)
- Memory and brain entries (currently SQLite)
- Studies as rows with vector embeddings, plus AGE edges to source
  scriptures/talks
- Conversation history with model calls and tool calls as audit rows
- Pipelines (triage → research → planning → execution → review)
- Scheduled / triggered work via `pg_cron` or the bgworker
- Cross-references that today live as markdown links in `.mind/` files

**What VS Code keeps doing:**
- Editing source code. The DB does not edit files. A Docker sidecar
  with a writable repo checkout edits files; the DB orchestrates.
- Showing diffs, syntax highlighting, language-server intelligence.
- Git plumbing — though the agent triggers git ops via the sidecar.
- The places where Michael wants to *see* what's happening live and
  steer with keyboard shortcuts.

**What Copilot-the-IDE-loop loses:**
- Authority over agent state. Today the agent's "memory" is `.mind/`
  files plus whatever Copilot manages to load into context. In the DB
  model, the agent's memory is rows, and Copilot becomes one of several
  clients (alongside web UI, CLI, Discord relay).
- Sole control of "what the agent should do next." Pipelines in
  Postgres can run without an open VS Code window.

**What replaces the loop:**
- A web UI on top of the stewards DB (could live in `becoming/` or a
  new sibling) where Michael sees the agent's worklist, can promote/demote
  items, can see in-flight model calls, can interrupt and steer.
- Long-running agents that don't need a human in the chat to keep
  working — they pull from the worklist, do the next step, write the
  result, NOTIFY for review.
- Tool execution containers (Docker) that mount specific repo paths,
  can run shell/git/build commands, and report structured results back
  to the DB.

**Hard limits to be honest about:**
- The DB cannot replace the *human-in-the-loop conversational quality*
  that makes Copilot good for ambiguous coding tasks. It can replace
  the *infrastructure underneath* it, but the user-facing UX of "I am
  paired with an AI as I type" is its own product.
- Long-running operations (15-minute builds, multi-hour training runs)
  need somewhere to live that isn't a Postgres backend. The bgworker
  dispatches but doesn't execute these.
- VS Code remains the best editor; we should not write a code editor.
  We should write the *substrate that informs* whatever editor is open.
- We have not yet solved "how does the agent open a file in Michael's
  editor and show him the change before committing." That remains a
  client problem.

## Posture check

Am I confirming the dream or actually testing it? The Timescale archival
was the disconfirming signal. The Microsoft blog and the converging
ecosystem patterns are confirming signals. Net read: the model is sound,
*and* the specific failure mode that killed pgai-the-extension
(managed-DB customers can't install extensions) does not bind a
self-hosted single-user setup.

Where I might still be wrong:

1. The bgworker pattern in pgrx is shown with SPI in a tx. Doing
   long-running outbound HTTP from a bgworker is *what pg_vectorize
   does*, but I have not read its bgworker code yet to confirm the
   shape. Probe before any spec.
2. AGE on PG18 is brand new (v1.7.0 RC0 January 2026). The `::name` cast
   gotcha suggests rough edges. We should prototype before committing.
3. Microsoft's blog uses `pg_diskann` (Microsoft's disk-resident vector
   index). They are pgvector-compatible but optional. We use vanilla
   pgvector HNSW. Performance shape may differ.
4. memory-pg18-by-yhw is one developer, two days old. Not a project we
   depend on; just an existence proof.
5. The "replace VS Code" claim is *partially* true — substrate yes,
   editor no. Be precise about which.

## Open questions / probes — STATUS UPDATE 2026-05-02

- ~~Read pg_vectorize's bgworker source. How does it own a tokio
  runtime?~~ → **Deferred to Phase 1.** Will be the first concrete
  task after the extension scaffold is in place; pg_vectorize
  remains the reference at
  [external_context/pg_vectorize/](../../external_context/pg_vectorize/).
- ~~Read pgai's archival rationale.~~ → **Done.** See "Confirmed
  archival rationale" above. Net read: confirming, not
  disconfirming. The lesson (worker outside backend) is what we
  already planned.
- ~~Stand up a docker-compose locally with `pgvector/pgvector:pg18`
  + `apache/age` and run the bridge end-to-end.~~ → **DONE.** See
  [probe/RESULTS.md](probe/RESULTS.md). All seven test blocks pass
  on PG18 + pgvector 0.8.2 + AGE 1.7.0. Two rough edges discovered
  and documented (`::name` cast, vertex agtype JSON cast).
- ~~Confirm `postgres_fdw` works against pgvector columns and AGE
  graphs.~~ → **Deferred to Phase 4 (or never).** The proposal uses
  URI-string references and gospel-engine-v2's HTTP API for cross-DB
  resolution. SQL-level joins via FDW are a possible upgrade later;
  not on the critical path.
- Sketch the SQLite→Postgres migration for `scripts/brain/`'s six
  categories — **Phase 1 deliverable.** See [phases.md](phases.md).
- Decide embedding model — **Decided: match gospel-engine-v2's
  `nomic-embed-text-v1.5` (768d).** Captured in [proposal.md](proposal.md).
