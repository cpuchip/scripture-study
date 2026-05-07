# pg-ai-stewards Phase 3c.2.5 — study tool registration

*2026-05-07 (Claude Code, Opus 4.7)*

## What this session was

Implementation of the 3c.2.5 spec from the same day's earlier
session. Tight session per Michael's "I don't have enough session
tokens to finish 3c.2 yet but probably enough for scoping" — the
scoping doc was the durable design; this session was the build.

Five tools registered. Substrate verification clean. The 3c stack is
now complete enough that 3c.3 (first real multi-stage pipeline) is
fully unblocked.

## What shipped

**Phase 3c.2.5 — study tool registration.**

### Files

- `extension/3c2-5-study-tools.sql` — single-file migration: budget
  hook columns, 2 underlying SQL functions, 5 wrapper functions,
  5 tool_defs rows, blanket `study_*: allow` perm grant.
- `extension/src/lib.rs` — twelfth `extension_sql_file!` reference.
- `extension/Dockerfile` — `3c2-5-study-tools.sql` added to COPY.

### What's now in `stewards.tool_defs`

```
brain_search_text   (Phase 1.5 seed)
skill               (Phase 1.5 builtin)
study_citations     (NEW)
study_context_for   (NEW — wraps stewards.context_for; agent-namespace
                     consistent for future brain_context_for etc.)
study_get           (NEW — line-paginated read with max_body_chars cap)
study_search_text   (NEW — websearch_to_tsquery FTS, multi-kind array)
study_similar       (NEW — wraps existing study_similar fn)
```

7 tools total. compose_tools('study', 'kimi-k2.6', session) now
emits 6 tools (5 study + skill builtin). JSON Schema preserved
end-to-end through the substrate.

### Underlying SQL additions

- `stewards.study_search_text(query, kinds[], limit)` — `websearch_to_tsquery`
  over body_tsv with `ts_headline` snippet generation. Multi-kind
  filter via array (empty = all).
- `stewards.study_get(slug, include_body, line_offset, line_count, max_chars)` —
  reads doc + frontmatter + citation count. When `include_body=true`,
  splits the body on newlines, slices `[offset+1 : offset+count]`,
  re-joins, applies `max_chars` safety cap. Returns rich jsonb with
  `body_total_lines`, `body_lines_returned`, `body_truncated_by_chars`
  so the agent can paginate explicitly.

### Tool-pattern resolution (the catch)

The 3a.1 import preserved the Copilot frontmatter `tools:` list as
`agent_tool_perms` rows verbatim — patterns like `gospel-engine-v2/*`
that don't match anything in the substrate's tool registry. So
adding `study_search_text` to `tool_defs` wasn't enough; the
imported agents' allow-list didn't cover it, and the deny-* fallback
won.

Fix: blanket `INSERT INTO agent_tool_perms (family, 'study_*', 'allow')`
across all non-watchman agents. The tools are read-only over substrate
state; broad access is safe. Watchman's deny-everything pattern is
preserved because watchman ships with a no-tools-by-design philosophy.

This is a real architectural gap worth naming: the imported tool
patterns (Copilot/MCP-style) and the substrate tool registry are
two separate vocabularies. They don't auto-translate. Future tool
registrations will need explicit perm grants until we either
(a) build the MCP-tools-as-substrate-tools bridge (Phase 3e), or
(b) systematically rewrite agent perm sets to use substrate names.

### Token budget hooks (designed empty)

`tool_defs` gained two NULL-by-default columns:

- `expected_result_tokens int` — typical token weight of one tool
  result. NULL = unknown.
- `expected_invocation_tokens int` — typical token weight of args +
  dispatch overhead. NULL = unknown.

Per Michael's call: "support a budget for tool calls, but lets leave
it empty until we know how much it'll cost us." When we have data,
populate. The 2.7b.3 estimator can refine cost predictions using
these. Until populated, behavior is unchanged (no enforcement).

### Verification

Substrate-level (the fast/cheap path; live tool-using chat deferred
to 3c.3 where it's natural):

1. **Direct SQL calls work:** `study_search_text_tool('{"query":"charity","limit":3}'::jsonb)` returns 3 hits; `study_get_tool('{"slug":"charity","body_line_count":5}'::jsonb)` returns body_total_lines=102 with 5 lines returned.
2. **Composition works:** `dry_run_chat('study', 'kimi-k2.6', ...)` emits 6 tools (was 1). All 5 study tools present in tools[].
3. **JSON Schema preserved:** `study_get`'s tool spec retains all field descriptions, type/min/max constraints, and required-field enforcement through compose_tools' jsonb storage and retrieval.
4. **Tool dispatch path unchanged:** the 1.6 sql_fn dispatch loop already proven for `brain_search_text`; new tools follow the identical wrapper shape.

## What was surprising

**The perm-bridge gap.** I'd expected `compose_tools` to emit the new
tools immediately after registration, but it didn't. The imported
agents had Copilot tool patterns (`gospel-engine-v2/*`) which don't
match `study_search_text`. Took one debug round to spot.

This deserves to be a permanent architectural note: tool registration
in `tool_defs` is necessary but not sufficient. The agent's
`agent_tool_perms` allow-list also has to match. They're two gates,
and the import preserved one without setting up the other.

**`ON CONFLICT DO UPDATE` and duplicate constrained values.** First
attempt at the perm broadcast hit `ERROR: ON CONFLICT DO UPDATE
command cannot affect row a second time`. The SELECT was joining
on `stewards.agents` which has multi-row keys `(family, model_match)`,
so each family appeared 1-2 times. `agent_tool_perms` keys on
`(family, tool_pattern)` only, so the same `(family, 'study_*')`
pair was being inserted twice in one statement. Fix: `SELECT
DISTINCT a.family`. Worth remembering as a Postgres gotcha when
broadcasting from a multi-row source.

**The 12th `extension_sql_file!`.** The container has been
live-applied through 12 SQL files now. None of the foldback
references have been rebuilt into the docker image since 2.7b.2.
Real risk: a fresh `docker compose down -v && up` would create an
extension at a much earlier state and skip everything. Worth a
batched rebuild before any container reset.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Phase 3c.3** — first real multi-stage pipeline. Now fully unblocked. Likely shape: `study-write` with stages `outline (plan)` → `sources (study, with study_search_text + study_get + study_citations)` → `critical-analysis (skill)` → `draft (study)` → `voice-check (skill)`. End-to-end test with kimi-k2.6 producing a real study from substrate state. |
| 2 | **Tool-call cost observation.** Once 3c.3's pipeline runs, populate the empty budget-hook columns. Per-tool: ~bytes returned + observed model context overhead. |
| 3 | **Image rebuild.** 12 SQL files folded; container has been live-applied through all. Rebuild before any `down -v`. |
| 4 | **Soak start** — independent of 3c stack. |
| 5 | **Architectural note durability** — capture the "tool_defs registration is necessary but not sufficient; perms are a separate gate" lesson somewhere more permanent than this journal. Probably in `docs/architecture.md`. |
| 6 | **Dockerfile vulnerability scan** — IDE flagged 17 high vulns on `rust:1-bookworm`. Pre-existing, not 3c.2.5; bump base image to a current security-patched tag in the next rebuild cycle. |

## What's still solid

- The Phase 1.5 wrapper pattern (`<name>_tool(jsonb) RETURNS jsonb`)
  scales cleanly to 6 tools without modification. It'll scale to
  60. The `coalesce(jsonb_agg(row_to_json(t)), '[]'::jsonb)` idiom
  is the canonical "table-fn → tool-result" shape.
- `websearch_to_tsquery` is a much better default than `to_tsquery`
  for natural-language tool args. Worth remembering for future
  search-tool registrations.
- The 3c.2.5 work was sized correctly. ~30 min estimate; actual was
  closer to that. Pre-built spec made the implementation phase
  almost mechanical — most of the design decisions had been made
  earlier in the day. That's the value of writing proposals
  durably before implementing.
