# pg-ai-stewards — Architecture Map

A reading map for the substrate, written from the SELECT-statement
perspective. Goal: 15 minutes to "I know where things live and what
to query." Not a phase plan ([phases.md](../phases.md)). Not a
proposal ([proposal.md](../proposal.md)). Just the runtime shape.

State of the world as of 2026-05-06: extension `pg_ai_stewards` 0.2.0,
running on PG18 with `vector` 0.8.2 + `age` 1.7.0. **23 tables, 3
views, 67 functions, 7 graph vertex labels.**

## Cluster shot

```
postgres cluster
└── db: stewards
    ├── extensions: vector, age, pg_ai_stewards
    ├── schema: stewards.*  (relational)
    ├── schema: ag_catalog.*  (AGE internals)
    └── graph: stewards_graph (cypher-queryable)
```

One database, one schema, one graph. Everything joinable in one tx.

## Six neighborhoods

The 23 tables fall into six clusters by purpose. Knowing which
neighborhood you're in is enough to find anything.

### 1. Work queue (the heartbeat)

The bgworker's job is to drain `work_queue`. Every async operation
is a row here.

| Table | What |
|-------|------|
| `work_queue` | Pending/in_progress/done/error rows. Each carries `kind`, `provider`, `payload jsonb`, `result jsonb`, `error text`. |

**Kinds** dispatched by the bgworker (via `dispatch()` in lib.rs):

- `embed` — POST to provider `/embeddings`, write `vector(768)` back to a target row.
- `chat` — POST to provider `/chat/completions`, insert assistant message, optionally enqueue `tool_dispatch` continuation.
- `tool_dispatch` — read parent assistant's `tool_calls`, execute each tool, insert `role='tool'` messages, enqueue continuation chat.
- `resolve_ref` — GET gospel-engine v2 `/api/get?ref=...`, cache result in `resolved_refs`.
- `echo` — stub from Phase 1, still present.

Starter queries:

```sql
-- What's pending?
SELECT id, kind, provider, created_at FROM stewards.work_queue
 WHERE status = 'pending' ORDER BY created_at;

-- Recent activity (with brief result peek)
SELECT id, kind, status, done_at - created_at AS elapsed,
       result->>'model' AS model
  FROM stewards.work_queue
 WHERE done_at > now() - interval '1 day'
 ORDER BY id DESC LIMIT 20;

-- Where did the bgworker spend tokens today?
SELECT result->>'kind' AS kind, count(*),
       sum((result->>'tokens_in')::int)  AS tok_in,
       sum((result->>'tokens_out')::int) AS tok_out
  FROM stewards.work_queue
 WHERE status = 'done' AND done_at > now() - interval '1 day'
 GROUP BY 1;
```

### 2. Agent runtime (sessions, messages, tool flow)

What an agent loop looks like at rest.

| Table | What |
|-------|------|
| `sessions` | `(id text PK, label, kind in chat/agent/tool/study/dev, created_at, last_active_at)`. Every conversation has one. |
| `messages` | The conversation log. `role in user/assistant/system/tool`, `content text`, `tool_calls jsonb`, `tool_call_id`, `parent_work_id` (links back to the work_queue row that produced this turn), `reasoning_content`/`reasoning_details` (thinking-model echo). |
| `tool_calls` | Empty in practice; the model's tool calls live on `messages.tool_calls` instead. Schema kept for future expansion. |
| `agents` | Persona registry. `(family, model_match)` PK. Holds `prompt`, `temperature`, `top_p`, `steps`, `response_format jsonb`. Variant-by-glob: same family can have a `'*'` row + a `'kimi-*'` row with different prompts. |
| `skills` | Same shape as agents; agents load these on demand via the `skill` builtin tool. |
| `instructions` | Reusable instruction blocks; `(family, model_match, scope)` PK where scope is `global \| agent:X \| session:Y`. |
| `tool_defs` | Tool registry. `name` PK, `args_schema jsonb` (JSON Schema), `execute_target jsonb` (`{kind:'sql_fn'\|'http'\|'subagent', ...}`). |
| `agent_tool_perms` | 3-state `(allow\|ask\|deny)` glob rules per agent. Last matching wins. |
| `agent_skill_perms` | Same shape, for skills. |

**Composition functions** (read-only, all `STABLE`):

- `compose_system_prompt(family, model, session)` — concat of agent persona + matching instructions + `<available_skills>` XML.
- `compose_messages(family, model, session, user_input?)` — `[system, ...history, ?user]` as jsonb.
- `compose_tools(family)` — `tools[]` filtered through perm rules.
- `dry_run_chat(family, model, session, input)` — full POST body that *would* go to `/chat/completions`. Read-only inspection target.

**Resolve helpers** (longest-glob-match wins):

- `resolve_agent(family, model)` → returns the matching `agents` row.
- `resolve_skill(family, model)` → same for `skills`.
- `tool_permission(agent, tool)` / `skill_permission(agent, skill)` → `'allow' \| 'ask' \| 'deny'`.

Starter queries:

```sql
-- All registered agents and their step budgets
SELECT family, model_match, mode, steps, response_format
  FROM stewards.agents
 ORDER BY family, model_match;

-- Recent assistant turns + their iteration count
SELECT m.id, m.session_id, m.model, m.tokens_in, m.tokens_out,
       m.finish_reason
  FROM stewards.messages m
 WHERE m.role = 'assistant'
 ORDER BY m.id DESC LIMIT 10;

-- "Did this loop finish or stall?"
SELECT * FROM stewards.session_status
 ORDER BY created_at DESC LIMIT 5;
```

### 3. Document substrate (studies as the universal kind)

Everything that has prose lives in `studies`. The `kind` column
discriminates.

| Table | What |
|-------|------|
| `studies` | `(id uuid PK, slug UNIQUE, kind, title, file_path, body, frontmatter jsonb, embedding vector(768), embedded_at, last_consolidated_at, created_at, updated_at, body_tsv tsvector GENERATED)`. The universal doc table. |
| `study_versions` | Snapshot history, written by the `touch_study` BEFORE-UPDATE trigger when `title`/`body`/`frontmatter` changes (NOT on embedding-only writes). |
| `resolved_refs` | Cache of gospel-engine `/api/get?ref=...` lookups. `(ref text PK, content jsonb, error, fetched_at, attempt_count)`. |

**`kind` values** currently in use:

| `kind` | source |
|--------|--------|
| `study` | 188 docs from `study/*.md` |
| `proposal` | 73 docs from `.spec/proposals/*.md` |
| `journal` | 70 docs from `.spec/journal/*.yaml` (synthesized body) |
| `doc` | 32 docs from `docs/work-with-ai/*.md` |
| `phase-doc` | 1 doc — `phases.md` itself |

**The `frontmatter jsonb` column** is the queryable projection of the
YAML at the top of each markdown file. Importer parses it; GIN index
makes it efficient.

```sql
-- Anything tagged WS5
SELECT slug, kind FROM stewards.studies
 WHERE frontmatter @> '{"workstream":"WS5"}';

-- Studies with a binding question, sorted by length
SELECT slug, length(frontmatter->>'binding_question') AS q_len
  FROM stewards.studies
 WHERE kind = 'study' AND frontmatter ? 'binding_question'
 ORDER BY q_len DESC;

-- Re-embedding state across the corpus
SELECT kind,
       count(*)              AS total,
       count(embedding)      AS embedded,
       count(embedding_error) AS errored
  FROM stewards.studies
 GROUP BY kind ORDER BY kind;
```

**Discovery functions:**

- `study_show(slug, sim_limit, cite_limit, verse_chars)` — formatted text blob: row + frontmatter + resolved citations + similar studies. Used by `stewards-cli study show`.
- `study_citations(slug)` / `study_citations_resolved(slug)` — citations from the AGE graph, optionally with verse text inlined.
- `study_similar(slug, limit)` — pgvector cosine neighbors via the `:SIMILAR_TO` edges (refresh-on-demand).
- `context_for(slug, depth)` — graph walk outward from a slug, returns ranked neighbors.
- `study_history(slug)` — verdict + finding timeline for a doc.

**Refresh functions** (recompute derived state):

- `refresh_study_refs(slug)` / `refresh_all_study_refs()` — enqueue `resolve_ref` work for unresolved citations.
- `refresh_study_similarity(slug, top_k, min_score)` / `refresh_all_study_similarity()` — recompute pgvector cosine neighbors and rewrite `:SIMILAR_TO` edges.

### 4. Brain (the PKM corner)

Phase 1's brain.exe replacement schema. Currently quiescent — used
for the initial design proof but the live brain still runs on
SQLite/chromem. The substrate is ready when you want to migrate.

| Table | What |
|-------|------|
| `brain_entries` | `(id, category, title, body, props jsonb, embedding vector(768))`. Categories: `people/projects/ideas/actions/study/journal/inbox`. |
| `brain_entry_tags` | `(entry_id, tag)`. |
| `brain_subtasks` | Per-entry checkable items. |
| `brain_versions` | Snapshot history. |

Search functions: `brain_search_text(q, cat, lim)` (FTS via tsvector),
`brain_search_vec(emb, cat, lim)` (cosine via HNSW).

### 5. The graph (AGE)

Lives in `stewards_graph`. **Cypher-queryable, not SQL-queryable**
directly — you go through `cypher()` function calls, usually
wrapped in PL/pgSQL helpers. AGE quirks documented in
[AGE-QUIRKS.md](AGE-QUIRKS.md).

**Vertex labels** (live counts as of import):

| Label | Count | What |
|-------|-------|------|
| `Study` | 364 | Every doc, regardless of `kind`. The label is "Study" historically; the `kind` property discriminates. |
| `Scripture` | 602 | Per-verse vertices materialized from gospel link parsing. URI is a workspace path like `eng/scriptures/bofm/mosiah/18.md#8`. |
| `Talk` | 279 | Conference talk vertices. |
| `Manual` | 43 | Lesson manual vertices. |
| `Reference` | 10 | Catch-all kind for non-canonical scripture references. |
| `Workstream` | 13 | One per WS1-WS9 + a few extras from frontmatter scan. |
| `Todo` | 1 | The one todo we just filed. |

**Edge types** in active use:

| Edge | What |
|------|------|
| `CITES` | Study → Scripture/Talk/Manual. Built from markdown link parsing in `import_study`. |
| `SIMILAR_TO` | Study ↔ Study via pgvector cosine. Materialized by `refresh_study_similarity`; carries `score` and `method='pgvector_cosine'`. |
| `FEEDS` / `REFINES` / `IMPLEMENTS` | Typed semantic edges from frontmatter declarations (Phase 2.6a). |
| `HAS_PROPOSAL` | Workstream → Study(kind=proposal). |
| `HAS_PHASE` | Proposal → Phase (when phase-splitter runs). |
| `HAS_TODO` | Workstream/Study/Phase → Todo. |
| `REFERENCES` | Default edge from raw markdown links (low-confidence catch-all). |

**Bootstrap function:** `ensure_studies_graph()` — idempotent;
`LOAD 'age'`, sets `search_path`, creates `stewards_graph` if missing.
Called from `00-extensions.sql` at install AND defensively from
`import_study` so a fresh session never sees "graph does not exist."

**Critical pattern when writing AGE code:** never string-concatenate
into Cypher bodies. Always use the 3-arg `cypher()` form with
`agtype` parameters. See AGE-QUIRKS #2 and #6 for why; reference
implementations are in `import_study` and `link_declared_edges`.

### 6. Watchman (self-maintenance)

The Phase 2.7 stack. Three layers built sub-phase by sub-phase:

| Table | Layer | What |
|-------|-------|------|
| `verdicts` | 2.7a | Per-pass verdict for one doc. Five values: `clean \| drift \| done \| superseded \| skipped`. Trigger on `record_verdict` bumps `studies.last_consolidated_at`. |
| `findings` | 2.7a | Drift recommendations + REM synthesis candidates. `acknowledged_at` enforces "surface once, then quiet" — open findings exclude their doc from `dirty_queue`. |
| `watchman_passes` | 2.7b.1 | One row per consolidation pass. Counters (`doc_count_done`, `tokens_in/out`, `verdict_counts`) advanced by trigger as chats complete. `budget_stopped` flag. |
| `watchman_config` | 2.7b.2 | **Singleton** (id=1, CHECK enforces). Holds the master `schedule_enabled` switch + cron timing + cooldowns + budget defaults + dirty/idle thresholds. |

**Key views** that compose the above:

- `dirty_queue` — `studies WHERE updated_at > coalesce(last_consolidated_at, '-inf') AND no open drift finding`.
- `watchman_pass_summary` — passes with `verdict_counts` unpacked into named columns.

**Decision + dispatch:**

- `watchman_should_fire()` → `'pressure' \| 'cron' \| 'idle' \| NULL`.
- `watchman_scheduler_inputs()` → all the live values feeding the decision (CLI uses this for `scheduler-status`).
- `watchman_scheduler_fire()` → calls the above; if non-NULL, calls `watchman_pass_start()`.
- `watchman_pass_start(limit, provider, model, agent, actor, trigger, budget)` → enqueues N chat work items (with `_watchman_pass_id` payload markers), returns `pass_id`.
- `watchman_input(slug)` — composes the user-message string for the consolidator agent (doc body + 1-hop graph neighborhood).
- `estimate_chat_tokens(slug)` — per-doc cost estimate (input/4 + 1500 + 30d-avg output).

**The completion trigger** (`handle_watchman_chat_completion`) fires
`AFTER UPDATE OF status` on `work_queue` with a WHEN-clause filter
on `payload ? '_watchman_pass_id'`. When a watchman chat lands
`done`/`error`, the trigger reads the assistant message, parses JSON,
calls `record_verdict` + (if non-clean) `record_finding`, and
advances pass counters via `advance_watchman_pass_counters`. All
side-effects in the same tx as the work_queue status flip.

Starter queries:

```sql
-- What needs Watchman attention right now?
SELECT count(*) FROM stewards.dirty_queue;

-- What did the last 5 passes find?
SELECT pass_id, started_at, doc_count_done,
       tokens_in + tokens_out AS total_tokens,
       n_clean, n_drift, n_skipped, budget_stopped
  FROM stewards.watchman_pass_summary
 ORDER BY started_at DESC LIMIT 5;

-- All open drift findings, severity-sorted
SELECT s.slug, f.severity, f.message, f.suggested_action,
       f.created_at
  FROM stewards.findings f
  JOIN stewards.studies s ON s.id = f.study_id
 WHERE f.acknowledged_at IS NULL AND f.kind = 'drift'
 ORDER BY array_position(ARRAY['high','medium','low'], f.severity),
          f.created_at;

-- Why isn't the scheduler firing?
SELECT * FROM stewards.watchman_scheduler_inputs();
```

### 7. Workstream / Todo / structural edges (Phase 2.6)

Project structure as graph + relational.

| Table | What |
|-------|------|
| `workstreams` | `(id text PK, name, description, status, frontmatter jsonb)`. WS1-WS9 etc. |
| `todos` | `(id uuid PK, slug UNIQUE, title, body, status in open/in_progress/done/dropped, parent_kind, parent_slug, created_by_session, completed_by_session)`. Lifecycle is permanent — done todos stay as historical record. |

Functions: `create_todo` / `complete_todo` / `list_todos` /
`todo_rollup_audit` (parent done with open children, etc.).

The graph carries the structural edges (`HAS_PROPOSAL`, `HAS_PHASE`,
`HAS_TODO`); these tables hold the lifecycle state.

## How a chat actually flows

End-to-end sequence for `stewards-cli watchman pass-now --limit 1`,
in 12 steps:

1. CLI calls `stewards.watchman_pass_start(1, ..., 'manual', ...)`.
2. SQL fn inserts a `watchman_passes` row (`status='in_progress'`).
3. SQL fn picks the head of `dirty_queue`.
4. SQL fn calls `watchman_input(slug)` to build the user message.
5. SQL fn calls `dry_run_chat(...)` — assembles `[system, user]` body via the composition functions.
6. SQL fn inserts a `work_queue` row with `kind='chat'` and the `_watchman_pass_id` payload marker. **Tx commits here.**
7. Bgworker's 500ms tick claims the row (`status='in_progress'`, separate tx).
8. Bgworker `dispatch('chat', 'opencode_go', payload)` — POSTs to `https://opencode.ai/zen/go/v1/chat/completions`. **No tx open during the HTTP call.**
9. Bgworker phase-3 tx: insert the assistant `messages` row, update `work_queue.status='done'`. **The completion trigger fires here, in the same tx.**
10. `handle_watchman_chat_completion` reads the assistant message, parses JSON, calls `record_verdict` + `record_finding`.
11. `record_verdict` bumps `studies.last_consolidated_at` (doc leaves `dirty_queue`).
12. `advance_watchman_pass_counters` increments `watchman_passes.doc_count_done`; if `done == planned`, marks the pass `completed`.

CLI's pollers see the pass status change, print summary, exit. The
human never had to coordinate any of this; the substrate did.

## JSONB shapes you'll actually query

Three columns where shape matters.

### `studies.frontmatter`

Whatever's in the YAML, parsed. **Conventional fields the system
treats as load-bearing:**

```jsonc
{
  "workstream": "WS5",                    // → declared edge to Workstream
  "feeds":      ["other-slug"],            // → :FEEDS edges
  "supersedes": ["older-slug"],            // → :REFINES edges
  "implements": ["proposal-slug"],         // → :IMPLEMENTS edges
  "binding_question": "What...?",          // free-form, used by study_show
  "watchman":   "skip"                     // PROPOSED — see open todo
}
```

Anything else is free-form and ignored by the substrate but
queryable via `frontmatter ? 'key'` and `frontmatter @> '{...}'`.

### `work_queue.payload` (varies by kind)

```jsonc
// kind='chat'
{
  "session_id":      "watchman-...--charity",
  "agent_family":    "watchman-consolidator",
  "requested_model": "kimi-k2.6",
  "meta":            { "agent_variant_match": "*", ... },  // from dry_run_chat
  "body":            { ...full OpenAI chat body... },
  "_watchman_pass_id":  "watchman-...",                    // OPTIONAL
  "_watchman_slug":     "charity",                         //  these signal
  "_watchman_actor":    "scheduler",                       //  to the
  "_watchman_estimate": 7961                               //  completion trigger
}

// kind='embed'
{
  "target_table": "studies" | "brain_entries" | "messages",
  "target_id":    "<uuid>",
  "input":        "<text to embed>"
}

// kind='tool_dispatch'
{
  "parent_work_id": 1234,
  "agent_family":   "stewards-explore",
  "model":          "kimi-k2.6",
  "session_id":     "..."
}

// kind='resolve_ref'
{
  "ref": "Mosiah 18:8"
}
```

### `work_queue.result` (varies by kind, on `done`)

```jsonc
// kind='chat' result
{
  "kind":            "chat",
  "provider":        "opencode_go",
  "model":           "moonshotai/kimi-k2.6-20260420",  // what provider actually used
  "session_id":      "...",
  "finish_reason":   "stop",
  "tokens_in":       3897,
  "tokens_out":      6961,
  "reasoning_tokens": 0,
  "billable_output": 6961,
  "tool_call_count": 0,
  "continuation_enqueued": null,
  "loop_stop_reason":      null,
  "response":        { ...full provider JSON... }
}
```

## Cost / safety invariants and where they live

Three independent guards stack. Each is a SQL constraint or a
specific function path; none rely on the agent's good behavior.

| Invariant | Enforced where |
|-----------|----------------|
| "Don't re-evaluate already-evaluated docs" | `dirty_queue` view's `WHERE` clause; `record_verdict` advances `last_consolidated_at` in the same tx |
| "Surface findings once, then go quiet" | `dirty_queue` excludes docs with open drift findings; `acknowledge_finding` is the only way to reset |
| "Don't pile up scheduler fires" | `watchman_should_fire()` returns NULL while an `in_progress` pass <1h old exists |
| "Don't spend over the per-pass budget" | `watchman_pass_start` exits the enqueue loop when `v_planned_tokens + v_estimate > v_budget`; sets `budget_stopped` |
| "Don't run unless authorized" | `watchman_should_fire()` returns NULL when `schedule_enabled = false` |
| "Don't dispatch tools without permission" | `compose_tools` filters tools through `agent_tool_perms`; the model never sees denied tools in its `tools[]` |
| "Don't lose model decisions" | `verdicts` table is append-only; `messages` is append-only |
| "Don't lose work to bgworker crashes" | Stale-claim reaper at bgworker startup marks orphaned `in_progress` rows errored; for `tool_dispatch`, synthesizes tool replies + enqueues continuation |

## What's NOT in the DB

- **Source markdown files** in `study/`, `.spec/proposals/`,
  `.spec/journal/`, `docs/work-with-ai/`. Authoritative for prose;
  the substrate is a parsed projection. `stewards-cli import` keeps
  them in sync.
- **gospel-library content** (scripture text, talk text). Lives on
  disk; gospel-engine v2 indexes it; `resolved_refs` caches what
  the substrate has fetched.
- **Provider API keys.** Read from env at postmaster start by
  `_PG_init`, available to the bgworker via `OnceLock`. Never in a
  table.
- **The Rust source** of the bgworker. Lives in `extension/src/lib.rs`;
  compiled into the docker image.
- **Compiled CLI.** `stewards-cli.exe` lives in `cmd/stewards-cli/`.
  Cross-compiles to linux/windows.

## How to find things

| You want to know... | Start here |
|---------------------|------------|
| Why did the model say X? | `messages` for the assistant turn; `work_queue.result` for the wrapper |
| What did the last pass do? | `watchman_pass_summary` newest-first; `watchman pass-detail <id>` for prose |
| Why isn't the scheduler firing? | `watchman_scheduler_inputs()` |
| What needs human attention? | `findings WHERE acknowledged_at IS NULL` |
| Which docs cite this verse? | Cypher `MATCH (s)-[:CITES]->(v {ref: $ref}) RETURN s` |
| Which docs are similar to X? | `study_similar(slug)` |
| What does the system know about workstream WS5? | `context_for('workstream-WS5', 2)` or Cypher walk |
| What's about to be enqueued? | `dirty_queue ORDER BY ... LIMIT 5` |

## Where to read deeper

| Question | File |
|----------|------|
| The plan and what's shipped | [phases.md](../phases.md) |
| Why the design choices | [proposal.md](../proposal.md) |
| AGE quirks and workarounds | [AGE-QUIRKS.md](AGE-QUIRKS.md) |
| The Rust dispatch loop | [extension/src/lib.rs](../extension/src/lib.rs) — start at `stewards_dispatcher_main`, follow `process_one_pending` and `dispatch` |
| Per-phase SQL migrations | `extension/2-6a-*.sql` through `2-7b3-*.sql` (chronological, each layered on the prior) |
| The CLI surface | `cmd/stewards-cli/main.go` (top of file lists every subcommand) |
| Verification queries | `extension/verify-*.sql` (one per sub-phase, illustrates expected behavior) |
