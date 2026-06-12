# pgEdge vs pg-ai-stewards — investigation notes

*2026-06-12, Michael's ask: "how do we compare? gaps? pros? cons? we do
better? we could learn?" Source: `external_context/pgedge-postgres-mcp`
(depth-1 clone, ★178) + org survey via gh API (26 repos). The "3 commits"
Michael saw doesn't match any flagship repo (all ≥10; likely a branch
view) — the substance is what matters.*

## What pgEdge is

A company productizing **distributed Postgres** (Spock multi-master
replication ★730 is the crown jewel) with a fast-growing **AI family**:
`pgedge-postgres-mcp` (MCP server + NL agent CLI + web UI),
`pgedge-rag-server`, `pgedge-vectorizer`, `pg_semantic_cache`,
`pgedge-go-llm-lib`, `ai-dba-workbench`, `pgedge-ai-kb`. The "lot more
advertised extensions" = their `postgres-images` repo: prebuilt PG 16–18
images, *standard* flavor bundling Spock, LOLOR, Snowflake, pgAudit,
PostGIS, pgVector, their Vectorizer, pg_tokenizer, vchord_bm25,
pg_vectorize, **pgmq**, pg_cron, pg_stat_monitor, Patroni, pgBackRest.

## The category difference (the headline)

**Their arrow points INTO Postgres; ours points OUT.** Their MCP server
lets an external LLM/agent query your database — NL→SQL, schema
inspection, hybrid search. The agent LOOP lives in the client (Go CLI /
React web); the server is a governed query surface. Our substrate makes
the database itself the agent's body — work items, pipelines, councils,
covenant, cost, memory as rows; the bgworkers run the cognition loop
inside Postgres and the bridge dials OUT to tools.

"Talk to your database" vs "your database thinks." Not competitors —
near-complements (their server could literally be a tool our bridge
dials).

## Their internal framework (verified in source)

- Single Go module. **Zero MCP SDK** — hand-rolled JSON-RPC + types in
  `internal/mcp/` (stdio + HTTP). **Zero LLM SDK in the server** — an
  `LLMClient` interface (anthropic/openai/ollama impls) + an **LLM proxy**
  (`/api/llm/*`) so browser clients never hold API keys. (Separate repo:
  `pgedge-go-llm-lib`, zero-dep unified Anthropic/OpenAI/Gemini/Ollama
  with streaming, tool calling, retries, HTTP/SSE proxy.)
- **9 MCP tools** (read-only-by-default `query_database`,
  `get_schema_info`, `execute_explain`, `generate_embedding`,
  `similarity_search`, `search_knowledgebase` (BM25+MMR),
  `read_resource`, connection list/select) + MCP resources + guided
  prompts. Fixed Go registry — tools are code, not data.
- Auth: per-session tokens (24h, SHA256-hashed, hot-reloaded YAML
  files), **per-session DB connection pools** (real multi-user
  isolation), TLS, RLS/CLS guidance, an explicit "NOT FOR PUBLIC-FACING
  APPLICATIONS" warning. Conversations persist server-side — in
  **SQLite** (a Postgres company keeping chat in SQLite, while our chat
  lives in Postgres).
- **Compactor** (`internal/compactor/`, 69 tests): client-triggered
  `POST /api/chat/compact`; 5-tier KEYWORD classifier (Anchor→Transient;
  "actually/instead/wrong"→anchor, "ok/thanks"→transient); always keep
  first message + recent window; provider-specific chars/token
  estimators (3.8/4.0/4.5 c/t + content multipliers: SQL 1.2×, JSON
  1.15×) ; optional LLM summarization; SHA-keyed cache; analytics
  counters.
- Dev practice: they build WITH Claude Code — `.claude/` carries a
  9-agent fleet (golang-expert, security-auditor, …) + a style-guide
  CLAUDE.md. Convergent with our practice, but no memory/covenant/intent
  layer — their CLAUDE.md is a style guide, not a relationship.
- Packaging: ghcr prebuilt images, devcontainers, nginx-fronted web,
  helm/ansible/CNPG charts org-wide, mkdocs docs site with per-client
  quickstarts (Claude Code/Desktop/Cursor/Windsurf/Copilot grid), 5 CI
  badges, PostgreSQL License.

## Where they are ahead (we should learn)

1. **Packaging & onboarding maturity.** Prebuilt images with explicit
   per-flavor package lists; quickstart-per-client grid; mkdocs nav;
   CI badges per component; changelog discipline. Direct P1/P3 input:
   publish ghcr images, ship an extension/seed manifest per image, write
   the client-onboarding grid for the office MCP (P5).
2. **Security documentation.** Consolidated checklist, token lifecycle,
   the public-facing honesty warning. We HAVE walls (tool perms, cost
   caps, sandboxes) but no consolidated security page — OSS needs one.
3. **Multi-user story.** Per-session pools + token auth. Our substrate is
   single-tenant by design today; the P5 office vision will need exactly
   this pattern.
4. **Token-estimation nuance.** Provider-specific c/t ratios + content
   multipliers + compaction analytics. Our pressure math uses flat 3.5
   c/t — cheap upgrade, worth stealing when we next touch the context
   engine.
5. **LLM-proxy pattern** for browser clients (keys never leave server) —
   keep for the OSS stewards-ui.
6. **pgmq + vchord_bm25/pg_tokenizer** in their standard image — worth a
   look at P1 image-selection time (pgmq vs our work_queue: ours is
   richer; but their image could inform our base-image choices).

## Where we are ahead (we do better)

1. **Durable server-side cognition.** Their agent loop is client-side;
   kill the CLI and the work dies. Our every turn is a row — crash-safe,
   replayable, auditable; multi-day work items; gates, verification,
   councils, sabbath/atonement; a cost ledger with caps and quarantine.
   They have NO workflow/verification machinery at all.
2. **Compaction, a generation apart.** Theirs: deterministic keyword
   rules + char counts (executor pattern). Ours: engrams extracted with
   judge questions at ingest, graduated pressure rendering, and the
   agent can pin/mute/compress its own context via [ctx:handle] — plus
   the planned compact_context commissioned-curation side quest. Their
   estimator nuances are worth borrowing; their classifier is not.
3. **Behavior is data.** Their tools/agents are Go code in a registry;
   ours are rows (tool_defs, grants, agents, pipelines, covenant) — an
   overlay directory extends us without a fork.
4. **Governance as state.** Covenant/intent/presiding in every dispatch,
   maturity ladders, human-Hinge gates. No analogue on their side —
   their CLAUDE.md governs the DEVS, nothing governs the AGENT.

## Direct synergies (no action now)

- Their MCP server as a bridge-dialed tool for schema-aware querying of
  any external Postgres (even the substrate's own DB — a DBA agent).
- `pgedge-go-llm-lib` as reference (or dependency) for any future
  Go-side LLM calls (persona-host local turns, OSS provider layer);
  their streaming/retry edge-handling is field-tested.
- Their docs-site structure as the P3 template.
- Their hand-rolled `internal/mcp` is a useful second reference for our
  bridge's protocol code (we use the official SDK; they prove the
  protocol is small enough to own).

## Verdict (Michael's questions, directly)

- **How do we compare?** Different species, shared habitat. They have
  production polish on a narrow surface (query your DB); we have deep
  agentic machinery (your DB thinks) with packaging debt.
- **Gaps (ours):** images/packaging, docs site, security checklist,
  CI visibility, multi-user auth. All are P1–P3 line items, none are
  architecture problems.
- **Do we do better?** At cognition, durability, governance, memory,
  context engineering — decisively. Nothing in their tree threatens the
  substrate's reason to exist; seeing a funded company ship the
  *complement* validates the territory.
- **Could we learn?** Packaging, onboarding grids, security candor,
  token-estimator nuance, the LLM-proxy pattern. Cheap, concrete, all
  noted as P1/P3 inputs in the extraction plan.
