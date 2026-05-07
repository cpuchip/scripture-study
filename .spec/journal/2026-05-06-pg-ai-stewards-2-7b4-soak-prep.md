# pg-ai-stewards Phase 2.7b.4 — soak prep

*2026-05-06 (Claude Code, Opus 4.7)*

## What this session was

Fourth and final code-side sub-phase of Phase 2.7. After 2.7b.3
landed token budget enforcement, the only remaining 2.7b deliverables
were the small `regenerate_active_md()` function and the frontmatter
watchman-exempt mechanism (option 3 from the conversation), then
runtime soak observation.

Also: shipped a transparency map (`docs/architecture.md`) earlier in
the session because Michael said the project felt like a black box.
466 lines, ~15 minute read; covers the six neighborhoods, the chat
flow, JSONB shapes, and the cost/safety invariant table.

Before starting code: a substantive conversation about what the
project actually is and what it enables. Most of the answer lived
in the substrate already; the conversation made it nameable. The
architecture doc is the durable form of that answer.

## What shipped

**Phase 2.7b.4 — soak prep** (the structural work; the soak itself
is runtime observation that follows).

### Files

- `extension/2-7b4-watchman-soak-prep.sql` — modifies `dirty_queue`
  view + adds `regenerate_active_md()` function.
- `extension/src/lib.rs` — ninth `extension_sql_file!` reference.
- `extension/Dockerfile` — added 2-7b4 to COPY.
- `cmd/stewards-cli/internal/show/show.go` + `main.go` — new
  `watchman active-md` subcommand.
- `projects/pg-ai-stewards/docs/architecture.md` — 466-line reading
  map (separate from 2.7b.4 but shipped same session).
- Closed todo `watchman-frontmatter-exempt` (1c503ff6) via
  `stewards-cli todo done`. The substrate was used to track its own
  delivery — eating dogfood, working as designed.

### Frontmatter exemption (option 3)

```sql
-- Added to dirty_queue view:
AND coalesce(lower(s.frontmatter->>'watchman'), '')
    NOT IN ('skip', 'exempt')
```

`lower()` for case insensitivity. Both `'skip'` and `'exempt'` are
accepted spellings. Zero schema change — the `frontmatter jsonb`
column with its GIN index already existed (Phase 2.1).

Mechanism verified by tagging one dirty doc, observing it disappear
from `dirty_queue`, untagging, observing it return.

### `regenerate_active_md()` sections

Returns markdown text. Sections:
- **In Flight** by workstream, with declared proposals from
  `frontmatter->>'workstream'`
- **Open Findings** sorted by severity, with message + suggested action
- **Open Todos** grouped by parent, in_progress marked with ▶
- **Recent Watchman Activity** (last 5 passes with verdict_counts)
- **Corpus Stats** with kind/total/embedded/in-dirty-queue table

Smoke-tested against live state. Output reveals concretely what
the soak prep needs: `journal | 70 | 70 | 70` in the dirty_queue
column means all 70 journals are currently dirty and would generate
noise findings during the soak.

## What was surprising

**One format-string bug, caught by inspection.** First version of
`regenerate_active_md` used `format('... \n ...')` with regular
single-quoted strings. PG's `format()` doesn't interpret backslash
escapes in non-E strings, so `\n` was literal. Fixed by switching
to `E'...\n...'`. The `replace(text, E'\n', E'\n  ')` calls were
already correct because they used E-strings throughout.

**The conversation about "what did we build?"** Michael asked. I
gave a real answer (Watchman as the killer feature, queryable
provenance, multiple-clients-one-brain, transactional consistency,
interruptibility). What surprised me was how the answer connected
back to the gospel framework he wrote about in March: the Abraham
4-5 pattern (council, plan, watch, redemptive correction, rest)
ISN'T philosophical decoration on top of agent infrastructure;
it's the *enforcement layer* — `record_finding` is "surface, don't
act"; the dirty-bit is "watch until obeyed"; the verdict-then-
acknowledge cycle is "council before action." Those are SQL
constraints, not vibes.

I told him that and he said: "Oh that's awesome that we built our
11 cycle creation working with AI agents into this project from the
ground up!" — meaning he'd already known this was happening; I'd
just finally named it.

**The black-box question.** When asked if the project felt opaque,
he said yes. So I wrote `docs/architecture.md` (15 min to read,
he confirmed). The act of writing it was useful for me too —
catalogued 23 tables, 3 views, 67 functions, 7 graph vertex labels.
Larger than I'd held in my head as one number.

## Carry-forward

| Priority | Item |
|----------|------|
| 1 | **Bulk-tag journals + start the soak.** 70 `.spec/journal/*.yaml` files need `watchman: skip` added. Trivial sed/python loop. Then re-import journals; verify `dirty_queue WHERE kind='journal'` is 0; flip `schedule_enabled=true`. Watch the trend line. |
| 2 | **Agent corpus import (option A).** Copilot/Claude agent definitions (`.claude/agents/*.md`, `.github/agents/*.agent.md`) and skills (`.github/skills/<name>/SKILL.md`) into `stewards.agents`/`stewards.skills`/`stewards.tool_defs`. Substrate has had the schema since Phase 1.5; only seed examples populated so far. This is the most direct answer to Michael's "agentic side, real work" framing. |
| 3 | **Image rebuild.** Nine `extension_sql_file!` references now; container has been live-applied through all of them but never docker-rebuilt since 2.7b.2. Rebuild before the soak so a container restart picks up the cumulative SQL. |
| 4 | **Phase 3c — pipelines + work_items**. Once the agent corpus is in, this is the orchestration layer that makes multi-step agent work durable. |
| 5 | ws6 AGE upstream PRs. Always queued. |

## What's still solid

- The whole 2.7 stack ships as a coherent unit: substrate (2.7a) +
  trigger-driven harvest (2.7b.1) + scheduler (2.7b.2) + budget
  (2.7b.3) + soak prep (2.7b.4). Each layer is independently
  verifiable and the cost guards stack.
- The transparency map (`docs/architecture.md`) is the answer to
  "I shouldn't have to be the only one who understands this." It
  captures a moment when the substrate is just-coherent-enough to
  be documentable but small enough to fit in 15 minutes of reading.
  Worth maintaining as the substrate grows.
- The substrate-tracks-its-own-work pattern works. We filed
  `watchman-frontmatter-exempt` as a `stewards.todos` row, then
  closed it via `stewards-cli todo done` when the work shipped.
  The auditable history lives where the work lives. No external
  ticket system to keep in sync.

## Note on Phase 2.7 closeout

This is the end of the 2.7b sub-phase tree as originally specced.
The 7-day soak that 2.7b.4 calls for is runtime observation — it
runs in calendar time, not git time. Whether 2.7 is "fully done"
depends on what we mean. Code-wise: yes. Soak-wise: it hasn't
started. The honest answer is "code shipped today; observation
period begins after the journal-tagging step."

The session-end discipline says to capture sabbath reflection at
phase boundaries. This is one. The right next sabbath agent
invocation would be after the soak has produced 7 days of data
showing the trend — that's the actual proof point for whether the
substrate's anti-loop discipline works in practice.
