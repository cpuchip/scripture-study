# pg-ai-stewards Phase 3a.1 — agent + skill corpus import

*2026-05-06 (Claude Code, Opus 4.7)*

## What this session was

After 2.7b.4 shipped, Michael asked the question I'd been carrying:
"do we have system prompts in the DB yet?" Answer at the time was
"the schema yes, but the content is mostly seed examples." That made
the next move obvious: import the actual agent corpus
(`.github/agents/*.agent.md`) and skill corpus (`.github/skills/<name>/SKILL.md`)
into the substrate. Started this immediately after the 2.7b.4 commit.

This is the smallest piece of work that makes the substrate
*concretely useful* for the agent ecosystem already in flight.

## What shipped

**Phase 3a.1 — agent + skill corpus import.** New importer modules
in `stewards-cli` plus `import` command dispatch.

### Files

- `cmd/stewards-cli/internal/importer/agents.go` — new file. Parses
  `.agent.md` / `.md` agent frontmatter (tolerant of both Copilot
  list-style `tools: [...]` and Claude comma-string `tools: A, B, C`).
  Inserts/updates `stewards.agents` row + rebuilds
  `stewards.agent_tool_perms` (deny-* + allow-skill + per-tool-allow).
- `cmd/stewards-cli/internal/importer/skills.go` — new file. Parses
  `SKILL.md` frontmatter; family from parent dir name. Inserts into
  `stewards.skills`. Truncates description to the 1024-char schema
  CHECK; warns if frontmatter `name` disagrees with dir name.
- `cmd/stewards-cli/main.go` — `runImport` now dispatches by kind:
  `agent`/`agents` → `ImportAgents`, `skill`/`skills` → `ImportSkills`,
  default → existing `ImportSource` for studies.
- Six `.github/agents/*.agent.md` files repaired: dev, journal,
  podcast, review, talk, ux. Each had a bare-array frontmatter line
  missing the `tools:` key — same shape across all six, almost
  certainly a copy-paste accident. Same-bug-same-fix; fixed in one
  commit alongside the importer.

### Result

```
agents              : 23  (19 imported + 4 pre-existing seeds)
skills              : 20  (Phase 1.5 seed source-verification +
                           scripture-linking were upserted with the
                           real SKILL.md content)
agent_tool_perms    : 272  (avg ~14 perms per imported agent)
compose_system_prompt('study', 'kimi-k2.6', 'test')  → 20159 chars
```

### Tool perm pattern

Each imported agent gets:
1. `(family, '*', 'deny')` — explicit deny-by-default makes the allow
   list load-bearing.
2. `(family, 'skill', 'allow')` — so the agent can load skills via
   the runtime `skill` builtin.
3. `(family, <each declared tool>, 'allow')` — verbatim from frontmatter.

Mirrors the Phase 1.5 stewards-explore seed exactly. Last-matching-
glob-wins resolution means future deny rules at the substrate config
level can override an agent's specific allow without re-importing.

### Idempotency

Reimporting clears the agent's existing perms then rewrites them, so
a removed-from-frontmatter tool actually goes away. Skills upsert via
ON CONFLICT DO UPDATE on `(family, model_match)`. Body content is
fully replaced on each run — substrate is a projection of the source
files.

## What was surprising

**6 agent files had broken YAML, all in the same shape.** Looked
like a copy-paste accident: line 2 has description, line 3 has bare
`[...]` array (no `tools:` key). Strict YAML rejects this. Fixing
six files manually was about as cheap as building a tolerant parser,
and the right move per files-as-source-of-truth: fix the source,
don't paper over it in the importer.

**`agents.steps` is NOT NULL but my INSERT defaulted it to NULL.**
Caught on first run — 19 import failures in a row. Fix: pass `8`
(the substrate default per Phase 1.5) explicitly. Worth noting:
the column DEFAULT exists, but my explicit INSERT with NULL
overrode it. The lesson is that PG defaults only apply when the
column is omitted from the column list, not when NULL is supplied.

**The skill content is rich enough that this matters now.** The
`study` agent's system prompt composes to 20K chars — that's the
agent prompt + matching instructions + an `<available_skills>` block
listing 20 skills the agent can load. Pre-import that was 2 skills.
The substrate can now meaningfully *prompt* the way Copilot/Claude
Code do.

## What this unlocks

The substrate can now answer "what would the `study` agent see if
it were dispatched right now?" with a concrete jsonb body via
`dry_run_chat('study', 'kimi-k2.6', '<session>', 'binding question
about charity')`. That body is identical in shape to what would POST
to `/v1/chat/completions`. Any client that can talk SQL can now
compose a real agent invocation.

The next step (Phase 3c — pipelines) becomes meaningful: a pipeline
that orchestrates the `study` agent across phases (outline →
sources → analysis → draft → ben-test) is now a substrate-level
coordination problem, not a "we don't have agents yet" problem.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Bulk-tag journal YAML files** with `watchman: skip` and re-import. Last gating item before the 7-day soak. |
| 2 | **Start the soak.** Flip `schedule_enabled=true` and watch the trend. |
| 3 | **Phase 3c — pipelines.** `stewards.pipelines` + `stewards.work_items` schema + dispatcher. The orchestration layer for multi-step agent work. |
| 4 | **Image rebuild.** Lib.rs has nine `extension_sql_file!` references; container has been live-applied through all of them but never rebuilt. Time to bake. |
| 5 | **`.claude/agents/dev.md`** as a Claude variant (`model_match='claude-*'`). Currently skipped; the Copilot version of `dev` was imported as `model_match='*'`. |
| 6 | **Handoffs preservation.** Frontmatter `handoffs:` blocks are parsed but discarded during agent import. If we want the chat UI handoff suggestions to survive into the substrate, add `agents.metadata jsonb` and store them. |

## What's still solid

- The "files as projections" architecture works for behavioral
  state, not just textual state. Agents are projections of their
  source files; the substrate stores parsed, queryable shape.
- The substrate's variant-by-glob design (`model_match`) means we
  can ship `study` once with `'*'` now and add a `'kimi-*'` variant
  later if Kimi reasons about the prompt differently. Schema's
  ready; the work is just authoring.
- Same-bug-same-fix discipline held: 6 files with the same broken
  frontmatter were fixed together rather than papered over with a
  tolerant parser. Source-of-truth wins.
