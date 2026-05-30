---
title: Competitive landscape — Postgres-as-AI-agent-substrate (2026-05-29 scan)
date: 2026-05-29
purpose: >
  Where the "agents in Postgres" space actually is, who the peers are, and
  where pg-ai-stewards sits. Captured because the original research (2026-05)
  found "only MS GraphRAG and pgai" — that's now badly out of date, and the
  positioning matters for the standalone-repo decision
  (see ../proposals/standalone-extraction.md).
method: Exa web search + reading project READMEs / docs, 2026-05-29.
---

# Where things are going

The space matured a lot between the original research (May 2026, "only MS
GraphRAG + pgai") and now. There are three distinct camps.

## Camp A — Postgres as the *store*; the orchestrator runs outside (mainstream, growing)

The dominant pattern. The agent loop runs in your app; Postgres persists state
and memory.

- **LangGraph `PostgresSaver` / `PostgresStore`** (`langgraph-checkpoint-postgres`)
  — checkpoints + cross-thread long-term memory as Postgres tables. This is what
  most teams now mean by "agent state in Postgres." Postgres is *underneath* the
  orchestrator, not the substrate that dispatches.
  https://docs.langchain.com/oss/python/langgraph/persistence
- **Tiger Data** (Timescale rebrand) — "unified Postgres for agent memory":
  hypertables (episodic) + pgvectorscale (semantic) + relational (procedural),
  one DB, one backup. Execution still external.
  https://www.tigerdata.com/learn/building-ai-agents-with-persistent-memory-a-unified-database-approach
- **pgcortex** (supreeth-ravi) — declarative agents-in-SQL bound to tables, but
  execution is explicitly **external** ("AI runs outside your DB"); triggers only
  enqueue. Its README argues *against* in-DB LLM execution (transaction blocking,
  resource exhaustion, atomicity). https://github.com/supreeth-ravi/pgcortex
- **DIY stack** — pgvector + pgmq + pg_cron as "the only database for production
  AI agents" (vectors + queue + cron in one ACID system; app executes).
  https://markaicode.com/architecture/postgres-agent-architecture/

## Camp B — call the LLM from inside SQL (contested — the giants are retreating)

**This is the camp pgai lived in, and the most important signal in this scan:**

- **pgai (Timescale)** is **archived** (Feb 2026, "no longer maintained or
  supported"), AND Timescale is **removing in-database LLM calls entirely** —
  `ai.openai_chat_complete`, `ai.anthropic_generate`, etc. gone by **June 30
  2026**, with official guidance to "move these calls to your application code."
  They consolidated on a Python library + external stateless workers.
  https://github.com/timescale/pgai ·
  https://github.com/timescale/docs/blob/latest/ai/vectorizer-deprecation.md
- The narrow survivors are trigger-based *column enrichment*, not agent
  substrates: **JigsawStack `postgres-llm`** and **Interfaze `postgres-llm`**
  (v2 went async: trigger enqueues, a worker drains the queue and writes back).
  https://github.com/JigsawStack/postgres-llm · https://interfaze.ai/blog/run-llms-inside-postgres

The retreat is real but mostly a retreat from (a) the *naive* synchronous
trigger-calls-LLM pattern and (b) the *managed-cloud* constraint — managed-DB
customers can't install custom extensions, so Timescale's market can't run an
in-DB worker anyway. pg-ai-stewards' original README called this exactly: the
constraint doesn't bind a self-hosted single-user system.

## Camp C — a Postgres *extension* that IS the agent runtime, with an in-DB background worker (rare, emerging 2026 — pg-ai-stewards' camp)

- **pgclaw** (calebwin, Feb 2026) — the closest **architectural** twin. Rust/pgrx
  extension, a `claw` column type binding an agent (simple LLM *or* a stateful
  "OpenClaw" agent) to a row, a **background worker that polls a queue and calls
  providers via `rig`**, and "a Claude Code in each row" via claude-agent-sdk
  (read/write files, run code, use tools). Newer, lighter, row-bound rather than
  a governed cycle. https://github.com/calebwin/pgclaw
- **NeuronDB / NeuronAgent** — the closest **feature** peer. PG extension + agent
  runtime (REST/WebSocket), state machine, tiered memory (STM/MTM/LTM), 16+ tool
  types, multi-agent collaboration, a workflow engine. **Proprietary** (the
  original research dismissed neurondb as "proprietary, single contributor"; it
  has since grown into a real product). https://www.neurondb.ai/neuronagent ·
  https://github.com/neurondb/neuron-agent

Both pgclaw and pg-ai-stewards put the bgworker *in* the DB but keep it *out* of
the foreground transaction path (write rows + `NOTIFY`, dispatch async). That is
the resolution to Camp B's transaction-blocking critique — it applies to the
naive synchronous pattern, not to bgworker-async.

## Where pg-ai-stewards sits

Architecturally: **Camp C** — and validated. The field is converging on Postgres
for agent state; the in-DB-execution critique that killed naive Camp B does not
apply to the bgworker-async design; the managed-cloud retreat doesn't bind a
self-hosted system.

Distinctively: **no one else has the governance layer.** pgclaw runs agents;
NeuronDB runs agents; LangGraph persists agents. None have maturity gates, a
trust ladder, enforced prepaid spend caps that refuse *before* spending, the
atonement / sabbath / consecration reflective primitives, the Judges pattern, or
the covenant/intent framing. That is a *point of view* about how agents should be
**governed**, not just executed — the most original thing in the project.

### Michael's framing (2026-05-29)

> "I kind of view this pg-ai-stewards as a combination of governance + hermes or
> pi agents or openclaw, all in a DB with arms and legs to actually do things.
> Opinionated and Jesus."

Unpacked, that's the synthesis the landscape is missing:
1. **Agent-in-a-row execution** (the openclaw/pgclaw capability) —
2. **plus a governance layer** no peer has —
3. **plus real actuation** — "arms and legs": tool dispatch, fs/git via Docker
   sidecars, the autonomous materializer writing results to disk. Most peers are
   either governance-light runtimes *or* passive memory stores; few are *governed
   agents that can actually act*, in the DB.
4. **plus an explicit, opinionated value frame** ("Opinionated and Jesus") — the
   covenant/intent/Restoration framing is load-bearing, not decoration.

## So what — implications for positioning

- **The honest pitch is not "ship it and they will come."** It's a single-
  maintainer, research-grade reference implementation entering a space that now
  has community- and venture-backed players (LangGraph, Tiger Data, NeuronDB).
- **The opinionated governance frame is both the moat and the barrier.** It's why
  someone would choose this over pgclaw (it has a view on *governance*, which
  pgclaw lacks) and why a general OSS audience may bounce (the gospel-study
  framing is idiosyncratic). For the public repo, frame it as *the* opinionated
  reference for **governed in-DB agents** — lead with the governance, present the
  faith framing as the author's honest origin, not a dependency.
- **"Most rewarding to work on" ≠ "most likely to find users."** Both can be
  true. Honor the first without inflating the second (Ben Test).
- **Cite the pgai retreat as design validation, not as cover.** The big player
  walking away from in-DB execution is a real signal; the defensible reading is
  "managed cloud can't do this; self-hosted can, and here's how to do it without
  the transaction-blocking footgun."

## Sources (read 2026-05-29)

- pgclaw — https://github.com/calebwin/pgclaw
- NeuronDB/NeuronAgent — https://www.neurondb.ai/neuronagent , https://github.com/neurondb/neuron-agent
- pgcortex — https://github.com/supreeth-ravi/pgcortex
- pgai (archived) — https://github.com/timescale/pgai
- pgai in-DB LLM removal — https://github.com/timescale/docs/blob/latest/ai/vectorizer-deprecation.md
- LangGraph persistence / PostgresSaver — https://docs.langchain.com/oss/python/langgraph/persistence
- Tiger Data agent memory — https://www.tigerdata.com/learn/building-ai-agents-with-persistent-memory-a-unified-database-approach
- pgvector+pgmq+pg_cron agent architecture — https://markaicode.com/architecture/postgres-agent-architecture/
- JigsawStack postgres-llm — https://github.com/JigsawStack/postgres-llm
- Interfaze "Run LLMs inside Postgres" — https://interfaze.ai/blog/run-llms-inside-postgres
- "Database-Native AI" (convergence overview) — https://tianpan.co/blog/2026-04-13-database-native-ai-when-postgres-learns-to-embed
