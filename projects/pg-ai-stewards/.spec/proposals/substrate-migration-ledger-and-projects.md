---
title: Substrate migration ledger + projects formalization (Option A)
date: 2026-05-12
status: ratified this session; build-ready
parent: substrate-h3-followup-small-items-and-sc-pivot.md
purpose: >
  Two related substrate primitives. The migration ledger gives us
  "run each .sql file exactly once" semantics — eliminating the
  regression class we just hit (h3-5 silently overwriting
  h3-followup-2's trigger extension). Projects A formalizes
  work_items.project_association into a real entity with a creation
  UI and a listing surface. Order: ledger first (foundational), then
  projects as the first real proof-case of the new pattern.
---

# Substrate migration ledger + projects formalization

## Why this proposal

Two concerns surfaced in conversation today:

1. **Bug stacking pattern.** The h3-5 trigger-overwrite regression
   wasn't caused by "too many .sql files." It was caused by **multiple
   files CREATE OR REPLACE the same shared object**, with no
   "run-once" guarantee. Today re-running an old file silently undoes
   a newer file's extension. With 94 .sql files in
   `projects/pg-ai-stewards/extension/`, we need a real migration
   pattern.

2. **Projects not first-class.** `work_items.project_association` is
   freeform text. The chip badges on WorkItems.vue are inferred from
   agent_planning inheritance, not from a project entity. There's no
   creation UI, no listing, no constraint that a value matches a real
   project. Michael wants real project grouping (Brain v3 style was
   referenced).

Migration ledger ships first because it's the foundation; projects
formalization is the first real "new-style migration" we test through
the new ledger.

## Ratifications

This session, two confirmations from Michael:
- **Migration approach: ledger** (matches ibeco.me pattern: DB on
  startup ensures migrations are applied)
- **Projects: Option A** (lightweight first; Brain v3 full-workspace
  shape "B" deferred — *"we'll do B later maybe even sandboxed git
  repos too"*)

Stewardship decisions inside the ratified scope:
- **Up-only migrations.** No rollback support. Forward-only matches
  the substrate's existing discipline (sabbath, atonement, lessons
  are all forward-only). Rollback can be retrofitted if needed.
- **Lexical ordering by filename.** Existing `h*.sql` /
  `3c2-*.sql` / `3e2-*.sql` naming is preserved. Sort order is
  whatever `filepath.Glob` returns (lexicographic). New migrations
  use the same naming convention.
- **Idempotent backfill.** Existing 94 files get an
  `INSERT ... ON CONFLICT DO NOTHING` into `schema_migrations` so
  the ledger says "already applied" without re-running them.
- **sha256 tracking.** Each migration record stores the file's
  sha256. If a file changes after being recorded, the migrator
  warns + skips (does not re-run). Manual intervention required
  to "amend" a migration. This catches the h3-5-style regression
  before it happens.
- **`stewards-cli migrate` command** owns the runner. Auto-runs
  via the bridge entrypoint before `bridge run` so substrate is
  always current. Available manually for ad-hoc use.

## Substrate primitives

### 1. `stewards.schema_migrations` table

```sql
CREATE TABLE stewards.schema_migrations (
    name        text PRIMARY KEY,            -- e.g. 'h3-1-schema-migrations'
    sha256      text NOT NULL,                -- of the .sql file contents
    applied_at  timestamp with time zone NOT NULL DEFAULT now(),
    notes       text                          -- e.g. 'backfilled', 'manual', 'auto'
);
```

Single source of truth for "which .sql files have run." Backfilled
with existing 94 files marked `notes='backfilled'`. New migrations
get `notes='auto'` when the entrypoint runs them.

### 2. `stewards-cli migrate` (Go subcommand)

```
stewards-cli migrate [--repo-root PATH] [--dry-run] [--target NAME]
```

Behavior:
1. Lists `projects/pg-ai-stewards/extension/*.sql` files in lexical order.
2. For each: computes sha256, checks `schema_migrations`.
3. If not recorded: `BEGIN; \i file.sql; INSERT INTO schema_migrations; COMMIT`.
4. If recorded but sha256 differs: warn + skip (regression detection).
5. If `--target=NAME`: stops after applying that migration.
6. If `--dry-run`: prints what would run without applying.

Failures inside the transaction roll back. The migrator continues to
the next file if a transaction fails (so a single broken migration
doesn't block all subsequent ones); a non-zero exit code reflects any
failure.

### 3. Bridge entrypoint startup hook

`bridge.Dockerfile` ENTRYPOINT becomes a shell script that:
1. Runs `stewards-cli migrate` (logs to stderr).
2. Then `exec stewards-mcp bridge run`.

Bridge restarts → substrate auto-current. No manual intervention.

## Projects A primitives

### 4. `stewards.projects` table

```sql
CREATE TABLE stewards.projects (
    slug             text PRIMARY KEY,
    name             text NOT NULL,
    description      text,
    root_directory   text,                    -- nullable; future B uses it
    archived         boolean NOT NULL DEFAULT false,
    created_at       timestamp with time zone NOT NULL DEFAULT now(),
    updated_at       timestamp with time zone NOT NULL DEFAULT now()
);
```

**Not** a FK target from `work_items.project_association` initially.
Soft reference so existing rows with project_association='pg-ai-stewards'
don't break when 'pg-ai-stewards' is/isn't in the projects table.
Future migration can add FK + constraint once the table is populated.

**Backfill:** SELECT DISTINCT project_association FROM work_items WHERE
project_association IS NOT NULL → INSERT each as a project with name =
slug (operator renames via UI).

### 5. Backend CRUD endpoints

| Endpoint | Action |
|---|---|
| `GET /api/projects/list` | List all projects + count of work_items per project |
| `GET /api/projects/get?slug=` | Single project detail |
| `POST /api/projects/create` | New project |
| `POST /api/projects/update` | Edit name / description / root_directory |
| `POST /api/projects/archive` | Soft-archive |

### 6. UI

- `/projects` view with table (slug, name, work_item count, archived,
  edit button)
- NewWork.vue: project_association becomes a dropdown sourced from
  `/api/projects/list`, with a "(no project)" option. Free-text
  fallback retained if the picker doesn't include a value you want.
- WorkItemDetail.vue edit form: project_association becomes a
  dropdown too.
- Sidebar entry for /projects.

## Build order

1. **Proposal doc + commit** (this file)
2. **Ledger schema** (`h-ledger-1-schema-migrations.sql`)
3. **Ledger Go runner** (`stewards-cli migrate`)
4. **Backfill** (one-off INSERT of all 94 existing files)
5. **Bridge entrypoint hook** + rebuild bridge image
6. **Projects table** (first new-style migration: `i1-projects-table.sql`)
7. **Projects backend endpoints**
8. **Projects UI** + NewWork picker
9. **E2E verification** + journal + summary

Each is a clean checkpoint. Commit at each.

## Non-goals (deferred)

- **Rollback / down migrations.** Forward-only suffices.
- **Concurrent migrator (e.g. advisory locks).** Single-user; we'll
  add when multi-instance matters.
- **Projects B: full workspace + sub-git.** Michael flagged "we'll do
  B later maybe even sandboxed git repos too." Deferred.
- **FK constraint on work_items.project_association.** Soft reference
  initially; add once projects table is stable.
- **Migrations for non-substrate things** (stewards-ui schema isn't
  in scope; it has no schema today anyway).

## Carry-forward after this batch

- yaml.rs Rust parser refactor (rule-of-three, still pending)
- Phase A pgrx longjmp catch + reaper (deferred; H.1.5a stable)
- materialized_at semantics drift
- Projects B (full workspace, sub-git, scoped fs-read)
- 14 pending SC work_items still awaiting Michael's ratification
