# pg-ai-stewards — proposal

**Status:** Proposal. Approved direction; awaiting Phase 1 kickoff.
**Owner:** Michael
**Verdict source:** [scratch.md](scratch.md), [probe/RESULTS.md](probe/RESULTS.md)
**Decision date:** 2026-05-02
**Companion docs:** [README.md](README.md), [phases.md](phases.md)

## Decision

Build a small Rust+pgrx Postgres extension (`pg_ai_stewards`) that turns
Postgres into the substrate for AI stewardship — agent state, memory,
work items, model calls, and pipelines, all as rows. Run it alongside
[pgvector](https://github.com/pgvector/pgvector) and [Apache AGE](https://github.com/apache/age)
in a separate database within the same Postgres cluster as
gospel-engine-v2.

This is not a fork of an existing project. The Azure-Samples
GraphRAG-docker repo, Timescale's pgai (archived), and Microsoft's
GraphRAG framework are reference designs we steal from. None of them is
the right base.

## Problem

Today the agent's "brain" is scattered:
- `.mind/` markdown files (canonical for some things, aspirational for others)
- `scripts/brain/` SQLite + chromem-go vector store (six categories: people, projects, ideas, actions, study, journal)
- `becoming.db` for practices and journaling
- Markdown studies in `study/`
- VS Code Copilot's in-context memory (ephemeral)
- Discord/relay caches
- gospel-engine-v2's Postgres (canonical canon)

Each surface has its own sync, backup, query model, and access pattern.
Several already disagree with each other. Adding a new agent capability
means deciding which store owns it and how the others find out.

We need one substrate. Postgres + pgvector + AGE + a thin Rust extension
can be that substrate.

## Why this is the right time

Three signals converged in the last six weeks:

1. **gospel-engine-v2 already runs on `pgvector/pgvector:pg18`.**
   Same image we'd extend. Same connection. Same backups. Verified in
   [scripts/gospel-engine-v2/docker-compose.local.yml](../../scripts/gospel-engine-v2/docker-compose.local.yml).
2. **Microsoft published the pgvector + Apache AGE pattern as
   recommended Azure PostgreSQL architecture** ([2026-04-15](https://techcommunity.microsoft.com/blog/adforpostgresql/combining-pgvector-and-apache-age---knowledge-graph--semantic-intelligence-in-a-/4508781)).
   Same query planner, same executor, same transaction. It's now boring
   infrastructure.
3. **The probe ([probe/RESULTS.md](probe/RESULTS.md)) confirms the
   bridge works on PG18 exactly as advertised.** Vector → AGE edge,
   combined CTE, all in one statement. ~50-second build, ~10-second
   boot, all seven test blocks pass.

The Timescale/pgai archival was the disconfirming signal we owed
serious attention. After research: they didn't kill the in-DB approach,
they consolidated on the Python-library + outside-worker pattern as
part of rebranding to TigerData / "Agentic Postgres" in mid-2025.
That's the lesson we already adopted (bgworker for LLM calls, never
foreground backend) — not a reason to abandon direction.

## Goals

In order of priority:

1. **Replace `scripts/brain/` SQLite with a Postgres-backed equivalent.**
   Six categories, full-text + vector search, brain entries, links.
   Keep the existing brain CLI/UI surface working.
2. **Make studies citation-aware.** A study row in stewards has AGE
   edges to scripture/talk URIs. Cross-DB lookups go through
   gospel-engine-v2.
3. **Externalize agent state from Copilot's context.** Sessions, work
   items, instructions, skills, tool calls — all rows. Multiple clients
   (VS Code, web UI, Discord, future) read and write the same canonical
   store.
4. **Make long-running agent work possible without an open IDE
   window.** Pipelines move work items between status columns;
   bgworker dispatches LLM calls; tool sidecars execute; results write
   back; `NOTIFY` triggers review.

## Non-goals

- **Replacing VS Code as an editor.** We orchestrate; we don't edit.
- **Multi-tenant SaaS.** Single-user. Add RLS later if ibeco.me ever
  hosts other people.
- **Replacing gospel-engine-v2.** Different DB, different ownership.
- **Inventing a new RAG framework.** Microsoft GraphRAG already exists
  and is good. If we want hierarchical community summaries on the
  scripture corpus someday, run GraphRAG against gospel and write the
  results into AGE. That's a Phase 4+ optional, not a goal here.
- **Calling LLM providers from foreground SQL.** Always via bgworker.
  This is the lesson Timescale paid the tuition for; we get it free.

## Architecture (load-bearing decisions)

### Topology
- **One Postgres 18 cluster.** Same one as gospel-engine-v2.
- **Two databases:** `gospel` (existing, owned by gospel-engine-v2) and
  `stewards` (new, owned by this project).
- **Three extensions in stewards:** `vector`, `age`, `pg_ai_stewards`
  (ours).
- **Cross-DB references** travel as stable URI strings
  (e.g. `lds://scripture/bofm/1-ne/3.7`). AGE edges hold them as
  properties. Resolution happens via gospel-engine-v2's existing HTTP
  API, not via SQL joins. (Optional Phase 4 upgrade: `postgres_fdw`
  for read-only views.)

### Process boundary
- **Foreground backends never call LLM providers or do tool execution.**
  They write rows and `NOTIFY`.
- **The Rust bgworker** owns a tokio runtime, listens on
  `stewards_dispatch`, calls providers via `reqwest`, writes results
  back. Same architectural pattern as `pg_vectorize`'s worker.
- **Tool sidecars** (Docker containers) execute filesystem, git,
  shell, and MCP operations. The bgworker dispatches to them; it does
  not exec code itself.

### Schema sketch (illustrative)

```
stewards.sessions          (id, label, created_at, last_active_at, kind)
stewards.messages          (id, session_id, role, content, model, tokens, created_at, embedding)
stewards.tool_calls        (id, message_id, tool, args jsonb, result jsonb, status, started_at, ended_at)
stewards.work_items        (id, kind, status, priority, payload jsonb, embedding, created_at, updated_at)
stewards.brain_entries     (id, category, title, body, props jsonb, embedding, created_at, updated_at)
stewards.studies           (id, slug, title, body, embedding, status, created_at, updated_at)
stewards.skills            (id, name, body, applies_to, created_at)
stewards.instructions      (id, scope, body, created_at, active)

ag_catalog.* (Apache AGE) — graph stewards_graph
   nodes labeled :Session, :Message, :WorkItem, :BrainEntry, :Study,
                  :Scripture, :Talk, :Person, :Project, :Idea, :Action
   edges labeled :CITES, :CITES_AS_CORE, :REFERENCES, :ABOUT,
                  :SIMILAR_TO {method, score}, :NEXT, :PRODUCED,
                  :TRIGGERED_BY, :OWNED_BY, :TAGGED
```

The "external" target nodes (`:Scripture`, `:Talk`) are stub vertices
with a `ref` property pointing at a gospel URI. They are not joined to
gospel's tables.

### Embedding model

Match gospel-engine-v2: `nomic-embed-text-v1.5` at 768 dimensions.
Same embedding space lets us compare a study draft directly to a
scripture verse vector across the DB seam, via gospel's HTTP API or
via shared embedding generation.

### Replacement of `scripts/brain/`

Brain's six categories become a `stewards.brain_entries` table with
a `category` enum and a JSONB `props`. Chromem-go vectors become
pgvector columns. The existing brain CLI keeps working — it gets a
new backend driver, not a new UI.

Migration path: write a one-shot Go migrator that reads SQLite +
chromem and inserts into Postgres. Run it once. Keep SQLite as
read-only fallback for ~30 days, then archive.

## What we are not deciding yet

These are deferred to Phase 4+ or to be revisited if data tells us to:

- Whether to run GraphRAG (the Microsoft framework) on top — could
  produce hierarchical community summaries of the scripture corpus
  for higher-quality global queries. Useful, expensive, optional.
- Whether to expose stewards via MCP server (the Azure-Samples repo
  shows the shape: `graphrag_search`, `age_get_schema_cached`,
  `age_entity_lookup`, `age_cypher_query`, `age_nl2cypher_query`).
  Probably yes, in Phase 3.
- Whether to use `pg_diskann` (Microsoft's disk-resident vector
  index) instead of pgvector HNSW for very large collections.
  Defer until we have data showing HNSW is the bottleneck.
- Whether `becoming/` (cloud hub) becomes the web UI surface or we
  build a sibling. Phase 3 question.

## Open risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| pgrx + bgworker + tokio is something we haven't done | Medium | Phase 1 deliverable is exactly this; pg_vectorize is the reference |
| Apache AGE PG18 is recent (1.7.0); rough edges may emerge | Low | Probe found two; both documented (`::name` cast, vertex JSON cast). Watch for more in Phase 2. |
| LLM provider quotas / cost during pipeline development | Medium | Use Ollama locally for dev. Cloud providers only for explicit promoted tasks. |
| Migrating brain without losing data | High | Run old + new in parallel for 30 days; brain CLI reads from both, writes to new |
| Web UI scope creep | Medium | Phase 3 only. Phase 1 ships with no UI; CLI + psql is enough |

## See also

- [README.md](README.md) — high-level pitch and context
- [scratch.md](scratch.md) — research provenance and source triage
- [phases.md](phases.md) — phased delivery plan with concrete deliverables
- [probe/](probe/) — working proof-of-concept docker stack
