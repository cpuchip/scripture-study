# Batch I — Agent Write-Back via Existing Gate Machinery

**Binding question:** How do agents propose studies, lessons, notes, exhibits, and SQL migrations through the substrate's existing trust ladder + gate machinery, with human ratification before persistence?

**Status:** ratified 2026-05-12 (re-opened after 2026-05-12 04:55 cancellation; tightened scope).

**Project:** pg-ai-stewards

## I. Why this is back

Batch I was paused on 2026-05-12 04:55 for two reasons: (1) **kimi-trust** — substrate-internal work was Claude-only; (2) **chicken-and-egg** — building agent-write-back ON a substrate the agents are writing TO felt recursive. Both reasons are addressed now:

- **kimi-trust:** Opus does all substrate-internal Rust/SQL work for this batch. Kimi shows up only as a *consumer* of the new pipeline once it exists. Same posture as the revise-proposal pipeline shipped earlier today.
- **chicken-and-egg:** The migration ledger shipped this morning is the missing piece. With the ledger, kimi (or any agent) can propose a `.sql` file → pending_file_writes → disk → bridge restart → ledger applies. Survives restart, image rebuild, container recreate. The recursion has a base case now: durable on-disk migrations.

Today's revise-proposal pipeline + apply_revision SQL function are themselves a small version of the Batch I pattern — agent proposes, human ratifies via UI, substrate writes. Batch I generalizes that to studies/lessons/notes/exhibits and adds the schema-migration leaf.

## II. Scope (re-scoped against current reality)

The 4 original work_items, against today:

| Original | Status now | This batch |
|---|---|---|
| #1 studies-generalization | **Done.** Schema has `tags`, `source_type`, `project_association` already. | Skip. |
| #2 agent-gate-sql | Routes proposals through `apply_gate_decision` on advance. | **In.** New `apply_agent_proposal` SQL function called from `on_maturity_verified` for the agent-proposal pipeline_family. |
| #3 agent-proposal-endpoint | HTTP endpoint for agent submissions. | **In** (in I.2). |
| #4 vue-review-queue | UI for ratification. | **In** (extended) — reuse Proposed-work panel, add `origin='agent_proposal'` filter. |

**Plus #5 (added 2026-05-12):** schema-migration source_type — Claude-only, syntax-validated, lands at `projects/pg-ai-stewards/extension/iN-<slug>.sql`. Closes the loop on "kimi creates tables that survive restart" once kimi-trust extends here.

**Plus exhibit (added 2026-05-12):** new source_type for SC and beyond. Knowledge artifacts with science-backing, materials, citations. Lives in `stewards.studies` with `kind='exhibit'`, files at `exhibits/<slug>.md`.

## III. Architecture

### Source-type map (locked)

| source_type | DB landing | File destination | Notes |
|---|---|---|---|
| `study` | studies (kind='study') | `study/<slug>.md` | Generic study artifact. |
| `lesson` | studies (kind='lesson') | `lessons/<slug>.md` | Lesson plan. Distinct from `stewards.lessons` table (which holds principles/decisions/sabbath_reflections per work_item). |
| `note` | studies (kind='note') | `becoming/notes/<slug>.md` | Shorter form; thought capture. |
| `exhibit` | studies (kind='exhibit') | `exhibits/<slug>.md` | Knowledge artifact w/ science-backing. SC and beyond. |
| `schema-migration` | (no DB row) | `projects/pg-ai-stewards/extension/iN-<slug>.sql` | Claude-only. Syntax-validated. Bridge restart applies. |

### Pipeline

New family `agent-proposal` with one stage:

```yaml
family: agent-proposal
description: "Agent proposes a study/lesson/note/exhibit/schema-migration through the trust ladder. Human ratifies via Proposed-work panel."
stages:
  - name: propose
    prompt_id: agent_proposal_propose
    next: null
    output_mode: json
    tools_disabled: true
file_destination_template: null  # dynamic per source_type
auto_materialize_on_verified: true
maturity_ladder: ["raw", "verified"]
```

**Why single-stage:** the agent already did its work upstream (in whatever context generated the proposal). The `propose` stage just emits structured JSON. Gate evaluation happens via Proposed-work panel — same as agent_planning work_items.

### JSON output shape (locked)

```json
{
  "source_type": "study|lesson|note|exhibit|schema-migration",
  "slug": "kebab-case-slug",
  "title": "Human-readable title",
  "body": "Full markdown body (for study/lesson/note/exhibit) OR full SQL (for schema-migration)",
  "frontmatter": { /* per-source-type metadata; jsonb */ },
  "project_association": "space-center" | "pg-ai-stewards" | null,
  "rationale": "Why this proposal exists (shown in ratification UI)"
}
```

### SQL routing

New function `stewards.apply_agent_proposal(p_work_item_id uuid) RETURNS jsonb`:

1. Read `stage_results.propose.output` from the work_item.
2. Validate JSON shape (`source_type`, `slug`, `title`, `body` required).
3. Branch on `source_type`:
   - `study` / `lesson` / `note` / `exhibit`: INSERT into `studies` (slug, title, body, frontmatter, kind=source_type, project_association). On conflict: error (caller decides to revise).
   - `schema-migration`: validate SQL syntax (`stewards.validate_sql_syntax(body)` — calls `pg_get_query_def` parse or psql-side check); on success, set `file_destination`. On parse fail: RAISE NOTICE with diagnostics, leave file_destination NULL, return error JSON.
4. Set `work_items.file_destination` based on source_type map.
5. Return summary jsonb.

Called from `on_maturity_verified` BEFORE `enqueue_work_item_file`, gated by `pipeline_family = 'agent-proposal'`.

### Validation: schema-migration

**This batch:** syntax-only via `pg_get_query_def`-style parse OR a temp psql parse. Cheap. Catches malformed SQL before file lands.

**Out of scope this batch:** full dry-run in an ephemeral schema. Deferred.

**Failure mode:** if SQL fails syntax check, work_item maturity stays at `raw` (or revises through revise-proposal pipeline). File never reaches disk. Bridge restart safe.

## IV. UI

Extend the existing Proposed-work panel on `WorkItemDetail.vue`. No new top-level view. The panel already handles:
- Ratify / Dispatch / Cancel
- Edit fields directly
- AI-revise with feedback + diff/accept

Add: `origin='agent_proposal'` filter chip on `/work-items` list view. Lightweight; doesn't replace WorkItemDetail panel — complements it.

## V. Phasing

**Batch I.1 (this session):** Proposal + pipeline + SQL routing + materialization wiring + smoke test on study/exhibit/note. ~1.5h.

**Batch I.2 (next session):** HTTP endpoint `POST /api/agent-proposals/create` + origin filter chip on work_items list. ~1h.

**Batch I.3 (third session):** Schema-migration source_type + SQL syntax validator. Claude-only. ~1.5h.

## VI. Non-goals

- Full dry-run validation of schema-migration SQL (defer)
- Multi-source-type proposals in a single work_item (one source_type per proposal)
- New top-level Vue views for proposals (reuse Proposed-work panel)
- Replacing existing `study-write` pipeline (agent-proposal is for *agent-originated* content, not for the structured study-write flow)
- Auto-ratify without human review (defer until trust scores prove the pattern)

## VII. Ratification summary

- **Routing:** through work_items, `pipeline_family='agent-proposal'`. ✓ Q1
- **Source types in scope:** study, lesson, note, exhibit, schema-migration. ✓ Q2
- **SQL validation:** syntax-only. ✓ Q3
- **Review UI:** extend Proposed-work panel + origin filter chip. ✓ Q4

## VIII. Stewardship decisions

- **`apply_agent_proposal` lives next to `apply_revision`** in a new SQL file `i4-agent-proposal-pipeline.sql`. Same shape, same atomicity discipline.
- **studies.kind values extended** in this batch: + `lesson`, `note`, `exhibit`. No CHECK constraint on kind today; future migration could tighten.
- **No FK from studies → projects.slug** initially (mirrors i2's pattern — soft reference first, harden when stable).
- **Schema-migration source_type is Claude-only.** Pipeline metadata flags this; gate refuses non-Claude actors.
