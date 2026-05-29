# pg-ai-stewards

> **Postgres as the substrate for AI agents.** Sessions, instructions, skills,
> memory, work items, model calls, costs, and the governance around them —
> all as rows. The agent's "filesystem" is the database.

`pg-ai-stewards` is a PostgreSQL extension (Rust / [pgrx](https://github.com/pgcentralfoundation/pgrx))
plus a small set of Go sidecars that turn one Postgres cluster into a complete
substrate for running AI agents: state, memory, multi-step pipelines, tool
dispatch, cost tracking, and human-in-the-loop review — queryable with SQL.

The load-bearing rule: **foreground SQL functions never call an LLM.** They
write rows and `NOTIFY`. A Rust background worker (its own tokio runtime)
dispatches the model and tool calls and writes results back. Everything an
agent does is an auditable row.

---

## Why

Most agent stacks scatter state across files, a vector store, a chat-history
cache, and a job queue — each needing its own sync, backup, and access path,
and several quietly disagreeing with each other. `pg-ai-stewards` collapses
that into one Postgres cluster:

- **One backup, one point-in-time recovery, one replication target.**
- **Vector + graph + relational in the same SQL statement** — `pgvector` for
  similarity, Apache [AGE](https://github.com/apache/age) for the relationship
  graph, plain tables for everything else.
- **Every agent action is a row** — sessions, messages, tool calls, work items,
  costs, gate decisions, lessons. Inspectable, joinable, transactional.
- **Any client reads the same store** — a CLI, an MCP server (so Claude Code or
  any MCP client can read/drive the substrate), a web UI, your own app.

It generalizes the ["vector databases are the wrong abstraction"](https://www.timescale.com/blog/vector-databases-are-the-wrong-abstraction/)
argument to *agent state*: don't put your agent's brain in a separate system
you have to babysit — put it in the database that already does the hard things
(durability, transactions, joins, access control) well.

## What it does

The substrate runs an **agentic creation cycle** — watch → diagnose → act →
account — with these capabilities, all as SQL-driven state:

| Capability | What it gives you |
|---|---|
| **Pipelines + work items** | Multi-stage agent flows (`research`, `study-write`, `agent-proposal`, …) as rows moving through stages. |
| **Background dispatch** | Rust bgworker calls providers (OpenAI-compatible) + tools; results land back as rows. Streaming, retries, circuit breakers. |
| **Memory + corpus** | Vector-indexed entries + an "engram" context-compaction layer so long runs don't blow the context window. |
| **Cost tracking + caps** | Per-call cost ledger, per-work-item caps, and enforced prepaid spend caps per provider (refuse-before-spend). |
| **Maturity gates + trust** | Work advances through maturity rungs with model-graded gates; a trust ladder governs autonomy. |
| **Multi-agent council** | Proposer / critic / synthesizer pattern for decisions; fan-out + brainstorm (12 lens techniques) for divergent work. |
| **Atonement / Sabbath / Consecration** | Failure→lesson capture, reflective pauses, and resource governance as first-class state. |
| **Scheduled pipelines** | Cron-style recurring agent work (daily digests, weekly research) with a plpgsql cron parser. |
| **MCP surface** | A Go MCP server exposes the substrate to Claude Code / any MCP client; a bridge daemon brokers outbound tool calls. |

Current state: extension `pg_ai_stewards` 0.2.0 on PostgreSQL 18 with
`pgvector` 0.8.2 + Apache AGE 1.7.0 — **65 tables, 10 views, 263 functions**,
31 pipelines, 48 agents. See [docs/architecture.md](docs/architecture.md) for
the runtime map.

## Architecture at a glance

```
PostgreSQL 18 cluster
└── db: stewards
    ├── extensions: vector, age, pg_ai_stewards
    ├── schema stewards.*   — relational state (sessions, work_items, …)
    └── graph  stewards_graph — AGE, cypher-queryable relationships
        │
        ├── bgworker (in-process, Rust + tokio + reqwest)
        │     LISTEN work_queue → call provider/tool → write result → NOTIFY
        │
        ├── stewards-mcp  (Go)  — MCP server: read + drive the substrate
        ├── stewards-cli  (Go)  — migrations, materialize-to-disk, ops
        └── bridge daemon (Go)  — brokers outbound MCP tool calls
```

Three containers run it: `pg` (Postgres + extension + bgworker), `ui`
(web console), `bridge` (outbound tool dispatch). See
[QUICKSTART.md](QUICKSTART.md).

## Quickstart

```bash
cp extension/.env.example extension/.env   # fill in provider keys
cd extension
docker compose up -d                       # pg + ui + bridge
# verify
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -c "SELECT stewards.version();"
```

Full setup, provider configuration, MCP integration, and the operational
runbook are in **[QUICKSTART.md](QUICKSTART.md)**.

## Providers

Any OpenAI-compatible endpoint works. Configure providers via
`STEWARDS_PROVIDER_<NAME>_<FIELD>` env vars (see
[extension/.env.example](extension/.env.example)). Tested with
[opencode.ai](https://opencode.ai) Zen (Kimi, GLM, Qwen, DeepSeek, MiniMax,
Claude), Google Gemini (OpenAI-compat endpoint), and local LM Studio / Ollama.
Per-model pricing lives in `stewards.model_pricing`; enforced prepaid spend
caps live in `stewards.provider_spend_caps`.

## Example use case

The author runs `pg-ai-stewards` as the agent substrate for a scripture-study
workspace: agents draft studies, evaluate sources, and cross-link to canonical
texts hosted in a sibling `gospel-engine-v2` Postgres database (vector search
over scripture). That integration is **one example**, not a requirement — the
substrate is domain-agnostic. Pair it with any corpus, or none.

## Documentation

- **[QUICKSTART.md](QUICKSTART.md)** — clone → run → drive via MCP.
- **[docs/architecture.md](docs/architecture.md)** — the runtime map (what to query, where things live).
- **[CONTRIBUTING.md](CONTRIBUTING.md)** — build cadence, conventions, how the pieces fit.
- **[docs/history/](docs/history/)** — design provenance: the original research verdict, phase plans, and proposals. Heavy but honest — this is how it was built.
- **[CLAUDE.md](CLAUDE.md)** — context for AI coding agents working in this repo.

## Status & standalone-repo note

The substrate is feature-complete and running. It is **not yet independently
buildable** as a standalone checkout: the Docker build context currently reaches
into a parent monorepo (shared Go workspace, sibling modules). Extracting it to
a self-contained repo is specified in
[.spec/proposals/standalone-extraction.md](.spec/proposals/standalone-extraction.md).
Until that lands, build from within the parent workspace.

## License

MIT — see [LICENSE](LICENSE).
