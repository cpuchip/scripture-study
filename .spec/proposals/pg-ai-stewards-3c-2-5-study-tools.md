---
workstream: WS5
status: proposed
created: 2026-05-07
phase: pg-ai-stewards-3c.2.5
feeds:
  - proposal-pg-ai-stewards-phase-2-5-generic-substrate
---

# pg-ai-stewards Phase 3c.2.5 — Study tool registration (sql_fn)

> Substrate-internal tool registration so a `study` agent dispatched
> via the 3c.1 pipeline machinery has a real tool surface to use.
> Sized between 3c.2 (auto-advance trigger) and 3c.3 (first real
> multi-stage pipeline). Path B from the 2026-05-07 conversation.

## Binding question

What's the smallest set of `sql_fn` tools that lets a study agent
do meaningful work on the substrate's existing 364-doc corpus
without external dependencies (no gospel-engine-v2, no MCP)?

## Why now

Phase 3a.1 imported the agent corpus including `study`. Phase 3c.1
ships the pipeline machinery. But `compose_tools(family)` only emits
tools that are **both** allowed by perm AND registered in
`stewards.tool_defs`. Today only `brain_search_text` (sql_fn over
empty brain_entries) and `skill` (builtin) exist. Without 3c.2.5,
3c.3's first real pipeline would dispatch a `study` agent that
literally has no tools to call.

3c.2.5 unblocks 3c.3 without building MCP infrastructure or HTTP
tool wiring (those remain Path A and Path C respectively, deferred).

## The 5 tools

All `sql_fn`. All wrappers around existing substrate functions.

### 1. `study_search_text`

FTS over `studies.body_tsv` (GIN-indexed since Phase 2.1).

```jsonc
args_schema: {
  "type": "object",
  "properties": {
    "query":  {"type": "string", "minLength": 1, "maxLength": 200},
    "kinds":  {
      "type": "array",
      "items": {"type": "string",
                "enum": ["study","doc","proposal","journal","phase-doc"]},
      "default": []   // empty = all kinds
    },
    "limit":  {"type": "integer", "minimum": 1, "maximum": 20, "default": 10}
  },
  "required": ["query"]
}
returns: [{slug, kind, title, snippet, rank}]
```

`kinds` is a multi-select array — empty (default) means all kinds.
Common patterns the agent will use:
- `kinds: ["study"]` — only canonical scripture studies
- `kinds: ["study","doc"]` — studies + meta-docs (e.g., docs/work-with-ai)
- `kinds: ["journal"]` — only personal reflection
- empty — survey across everything

SQL filter: `WHERE cardinality($1::text[]) = 0 OR kind = ANY($1)`.

### 2. `study_get`

Read a doc's body + frontmatter + citation summary, with **line-based
pagination** so large docs don't blow context.

Mirrors the Read tool's `offset`/`limit` semantics that work well in
agent reasoning — line boundaries are natural reading units (no
mid-word splits) and the model can plan "give me lines 200-400 next."

```jsonc
args_schema: {
  "properties": {
    "slug":             {"type": "string"},
    "include_body":     {"type": "boolean", "default": true},
    "body_line_offset": {"type": "integer", "minimum": 0, "default": 0},
    "body_line_count":  {"type": "integer", "minimum": 1, "maximum": 1000, "default": 200},
    "max_body_chars":   {"type": "integer", "default": 20000, "maximum": 50000}
  },
  "required": ["slug"]
}
returns: {
  slug, kind, title, frontmatter,
  body?:                  string,    // only if include_body=true
  body_line_offset:       int,       // echo of requested offset
  body_lines_returned:    int,       // actual lines in this slice
  body_total_lines:       int,       // total in the document
  body_truncated_by_chars: bool,     // true if max_body_chars hit before line_count
  citation_count:         int
}
```

**Default `include_body=true`** — that's the point of the command;
the metadata-only mode is the override, not the default.

The agent decides whether to page. If `body_total_lines >
body_line_offset + body_lines_returned`, it can call again with
`body_line_offset = previous_offset + body_lines_returned` to
continue. The 1000-line / 50000-char ceilings are hard caps; defaults
are 200 lines / 20K chars (loose enough for a typical study, tight
enough that reading 5 docs fills 100K of context, not a million).

**Limits worth iterating on once we have data:** 200 might be too
tight for proposals (avg ~400 lines). 20K chars might be too tight
for stage-1 outline reading. Track real usage and tune.

### 3. `study_similar`

Uses precomputed `:SIMILAR_TO` AGE edges (Phase 2.3) — no on-the-fly
embedding required. Avoids a sub-dispatch loop.

```jsonc
args_schema: {
  "properties": {
    "slug":      {"type": "string"},
    "limit":     {"type": "integer", "minimum": 1, "maximum": 10, "default": 5},
    "min_score": {"type": "number",  "minimum": 0, "maximum": 1}
  },
  "required": ["slug"]
}
returns: [{slug, title, score, direction}]   -- direction: outgoing|incoming|mutual
```

### 4. `study_citations`

What canonical sources (scriptures/talks/manuals) does a doc cite?
Already a SQL function; trivial wrapper.

```jsonc
args_schema: { "properties": { "slug": {"type":"string"} }, "required": ["slug"] }
returns: [{cited_uri, cited_kind, anchor_text, citation_count}]
```

### 5. `study_context_for`

Graph walk outward from a doc — typed edges (`:HAS_PROPOSAL`,
`:CITES`, `:FEEDS`, `:SIMILAR_TO`) up to depth N. Underlying SQL
function `stewards.context_for(slug, depth)` already exists; we
register it as `study_context_for` to keep the agent-facing tool
namespace consistent (`study_*` for tools that operate over
`stewards.studies`).

Anticipates future kinds: `brain_context_for`, `todo_context_for`,
etc. as we add other table-rooted graph walks.

```jsonc
args_schema: {
  "properties": {
    "slug":  {"type": "string"},
    "depth": {"type": "integer", "minimum": 1, "maximum": 4, "default": 2}
  },
  "required": ["slug"]
}
returns: [{hop, direction, edge_type, neighbor, neighbor_kind}]
```

The underlying SQL function stays at `stewards.context_for` (the
CLI's `runContext` and `watchman_input` already call it; renaming
breaks both). Only the agent-tool name is `study_context_for` —
the wrapper `stewards.study_context_for_tool(jsonb)` calls the
existing function under the hood.

## Wrapper pattern

Same shape as Phase 1.5's `brain_search_text_tool`. Each tool gets:

```sql
CREATE FUNCTION stewards.<name>_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql STABLE AS $$
DECLARE
    v_<arg> <type>;
BEGIN
    -- Decode args from jsonb. Apply defaults explicitly so a
    -- malformed/missing arg becomes a clear error, not a silent NULL.
    v_query := p_args->>'query';
    IF v_query IS NULL THEN
        RAISE EXCEPTION '<name>_tool: query is required';
    END IF;
    v_limit := coalesce((p_args->>'limit')::int, 10);

    RETURN coalesce(
        (SELECT jsonb_agg(row_to_json(r))
           FROM stewards.<existing_fn>(...) r),
        '[]'::jsonb
    );
END $$;
```

Plus the `tool_defs` row:

```sql
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target)
VALUES (
    '<name>',
    '<one-paragraph agent-facing description>',
    '<args_schema jsonb>'::jsonb,
    '{"kind":"sql_fn","name":"stewards.<name>_tool"}'::jsonb
);
```

## Token budget hooks (designed, not populated)

Per Michael's call: support a budget for tool calls, but leave it
empty until we observe real costs.

### Schema additions

```sql
-- On tool_defs: typical cost per invocation. NULL = unknown.
ALTER TABLE stewards.tool_defs
    ADD COLUMN IF NOT EXISTS expected_result_tokens int,
    ADD COLUMN IF NOT EXISTS expected_invocation_tokens int;

-- On pipelines.stages[]: per-stage tool call cap. Stored in jsonb;
-- no schema change needed. Documented convention:
--   stages: [{
--     name, agent_family, model, provider, next?, auto_advance?,
--     max_tool_calls?, tool_call_token_budget?
--   }]
```

### Behavior

- **Today:** all new columns NULL, all existing pipelines have no
  tool budget keys. Enforcement is a no-op. Same behavior as 3c.1.
- **Future (after observing real costs):** populate
  `tool_defs.expected_result_tokens` per tool. The 2.7b.3
  `estimate_chat_tokens` function gains a `+ N × expected` term when
  agent perms include tools. Per-stage `max_tool_calls` enforced in
  the 3c.2 trigger by counting `tool_dispatch` work_queue rows
  attached to the stage's session_id.
- **Why it works:** the existing per-pass `token_budget` on
  `work_items` already caps total spend. The new fields refine the
  estimate but don't replace the hard cap.

### What this preserves

If the design decision later reverses (e.g., we move to
elapsed-time budgets, or per-tool $-cost rather than tokens), the
schema additions are non-breaking — drop columns, add new ones.
None of the current behavior depends on them.

## Sample agent flow on this surface

A `study` agent dispatched with binding question "what does my
corpus say about consumption decreed":

```
1. study_search_text(query="consumption decreed", limit=10)
   → 8 hits across kinds (study, journal, proposal)
2. study_get(slug="consumption-decreed", include_body=true)
3. study_get(slug="consumption-decreed-modern-warning", include_body=true)
4. study_similar(slug="consumption-decreed", limit=5)
   → discovers zion-blueprint, etc.
5. context_for(slug="consumption-decreed", depth=2)
   → finds the proposals + journal entries that touched the topic
6. study_citations(slug="consumption-decreed")
   → list of D&C 87 references etc. (cited as URIs only — no text
     resolution without gospel-engine-v2 or Path A)
```

The agent can do "synthesis across my own work" studies. It
**cannot** do source-verification against canonical scripture text
without a downstream Path A or Path C step. That's an honest
limitation worth naming in the agent's system prompt.

## Effort estimate

- 5 wrapper functions: ~75 lines SQL
- 5 tool_defs INSERTs with descriptions + JSON Schema: ~50 lines
- 1 schema migration for budget hook columns: ~5 lines
- Verification: one tool-using chat to confirm dispatch + JSON
  return shape per tool. ~5–10k tokens total.
- **Single SQL file `3c2-5-study-tools.sql`, ~30 min in a fresh session.**

Goes through the standard foldback: live-apply via psql, add to
lib.rs `extension_sql_file!` chain, add to Dockerfile COPY list.

## Done when

1. `SELECT * FROM stewards.tool_defs WHERE name LIKE 'study_%' OR name = 'context_for'` returns 5 rows.
2. `compose_tools('study')` for the imported study agent now
   includes those 5 tools (the import already allowed them via
   tool_pattern matches).
3. `dry_run_chat('study', 'kimi-k2.6', '<session>', 'find docs about charity')`
   produces a body where `tools[]` contains the 5 functions with
   their JSON Schemas, and the model can reasonably call them.
4. End-to-end smoke: a one-stage pipeline dispatches a study agent;
   the agent calls `study_search_text` and gets a structured result
   back via the existing tool_dispatch loop. (No need for a real
   multi-stage pipeline — that's 3c.3.)

## Resolved decisions (2026-05-07)

1. **`study_get` body inclusion:** YES default `include_body=true` —
   that's the point of the command. **Plus line-based pagination**
   (mirrors the Read-tool semantics that work well in agent
   reasoning): `body_line_offset` + `body_line_count` with a
   `max_body_chars` safety cap. Limits worth iterating on once we
   have real usage data — Michael's note: "we should play with the
   limits."
2. **`study_search_text` multi-kind filter:** YES, multi-kind via
   array. Renamed arg from `kind` (singular) to `kinds` (plural,
   array). Empty array = all kinds. Common pattern: `kinds:
   ["study","doc"]` to skip journals + proposals.
3. **Naming consistency:** YES, rename to `study_context_for` for
   the agent-facing tool. Underlying SQL function stays at
   `stewards.context_for` (it's already called by the CLI's
   `runContext` and `watchman_input`; renaming the SQL function
   would break those). The wrapper `stewards.study_context_for_tool(jsonb)`
   bridges. Anticipates future `brain_context_for`, `todo_context_for`
   etc. as we add other table-rooted graph walks.

## What's deliberately NOT in 3c.2.5

- HTTP tools for gospel-engine-v2 (Path A). Defer to 3c.4 or later.
  Cleanest after 3c.3 demonstrates the pipeline pattern works on B.
- MCP client (Path C, Phase 3e). Big work; far horizon.
- Embedding-on-the-fly for vector search of arbitrary user queries.
  Requires sub-dispatch through `embed` work_kind; defer until we
  have a clear use case `study_similar` doesn't already cover.
- Write-side tools. The 3c.3 study-write pipeline's terminal stage
  inserts the produced study via `stewards.import_study()` directly;
  no tool needed.
- Full `study_show` exposure as a tool. The function returns
  formatted text (markdown), which agents tend to parse poorly. The
  structured `study_get` + `study_citations` + `study_similar` trio
  gives the same information in agent-consumable shape.

## Carry-forward (when 3c.2.5 ships)

| Priority | Item |
|----------|------|
| 1 | Phase 3c.3 — first real multi-stage pipeline (study-write). Now unblocked. |
| 2 | Observe real tool-call costs. Populate `tool_defs.expected_result_tokens` once we have data. |
| 3 | Phase 3c.4 (deferred name) — register HTTP tools for gospel-engine-v2 (Path A). |
| 4 | Watchman soak start when ready. Independent of this work. |
