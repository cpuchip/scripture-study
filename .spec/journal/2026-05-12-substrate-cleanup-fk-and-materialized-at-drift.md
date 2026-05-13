---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Substrate cleanup pulse — FK on project_association + materialized_at rename"
status: shipped (e2e validated)
carry_forward:
  - "yaml.rs Rust parser refactor (rule-of-three triggered, Claude-only)"
  - "Phase A pgrx longjmp catch + 60s reaper"
  - "Projects B — full workspace + optional sub-git (deferred per ratification)"
  - "14 SC work_items still awaiting Michael's ratification — 4 parent plans + 14 child planning proposals"
links:
  - "../../projects/pg-ai-stewards/extension/i2-projects-fk-constraint.sql"
  - "../../projects/pg-ai-stewards/extension/i3-rename-work-items-materialized-at.sql"
  - "../../projects/pg-ai-stewards/extension/smoke/i2-fk-smoke.sql"
---

# Substrate cleanup pulse — FK + materialized_at rename (2026-05-12)

Two commits. The carry-forward A lane ratified earlier: "FK on work_items.project_association" + "materialized_at semantics drift." Both shipped through the new ledger pipeline. No Rust touched — pure substrate cleanup.

## What shipped

### i2 — FK on work_items.project_association → projects.slug (commit `a2ffe8a`)

Hardens i1's soft reference. Pre-flight showed clean data: 27 NULL, 20 'space-center', 5 'pg-ai-stewards', zero orphans. Both non-NULL slugs already in projects (i1's backfill).

Semantics ratified:
- `ON UPDATE CASCADE` — slug renames propagate to work_items rows
- `ON DELETE RESTRICT` — projects with work_items can't be deleted (archive instead; hard delete isn't exposed in the UI anyway)

The migration has a defensive `DO $$ ... RAISE EXCEPTION IF v_orphans > 0 $$` block in front of the ALTER so any orphan in a different environment surfaces a readable error before the ALTER's less-readable one.

Smoke (i2-fk-smoke.sql): bogus UPDATE blocked, NULL accepted, known slug accepted, DELETE blocked by RESTRICT, slug rename cascade verified.

Second migration through the ledger end-to-end. Bridge restart: *"applying 1 migration(s) ✓ i2-projects-fk-constraint."*

### i3 — work_items.materialized_at → file_enqueued_at (commit `62dbe1b`)

The drift: the column reads as "when the file was materialized to disk" but is set at QUEUE time inside `enqueue_work_item_file()` — *before* anything is written. The matching column on `pending_file_writes` (also named `materialized_at`) is set later by `stewards-cli materialize-writes` when the file actually lands. Two tables, same column name, different lifecycles. Same shape as the h3-5 "one canonical owner per shared object" lesson, in naming form.

The path the user picked was **Path A — rename**, not Path B (just fix comment + UI label). The honest fix.

Atomic migration:
- `ALTER TABLE RENAME COLUMN materialized_at TO file_enqueued_at`
- `CREATE OR REPLACE enqueue_work_item_file` (was setting `materialized_at = now()`)
- `CREATE OR REPLACE on_maturity_verified` (was reading `NEW.materialized_at` — full H.1.6 + H.3-followup-2 logic re-applied with new column name)
- Verify DO block: confirms new column exists + old does not

Paired UI updates:
- Go struct field `MaterializedAt` → `FileEnqueuedAt` + json tag
- TS type updated
- Vue label "✓ materialized {date}" → "✓ queued {date}" (the honest verb)
- All Vue conditionals re-pointed

Data preserved through rename: 9 rows had values; 9 rows still have them after rename. Total 52 work_items unchanged.

E2E: `GET /api/work-items/get?slug=h3-6-substrate-next-three` returns `file_enqueued_at: 2026-05-12T04:56:59...` and `materialized_at` is absent from the response payload.

## Architecture wins

**The ledger absorbed two different regression-class concerns this session.** i2 was a soft→hard FK transition that needed orphan-safety up front. i3 was a column rename with three function-redefinition dependencies. Both shipped as drop-a-file + bridge-restart. No `psql -f` workflow. No manual `CREATE FUNCTION` workaround needed.

**Discipline confirmed: don't edit applied migration files.** The `materialized_at` references in `6d-pending-file-writes.sql`, `h1-6-2-verified-trigger.sql`, and `h3-followup-2-render-file-destination.sql` were left untouched. The ledger's sha256 drift detection would catch any edit; i3 supersedes via `CREATE OR REPLACE` (functions) and `RENAME COLUMN` (schema). Historical migration files remain historical truth.

**Honest UI label.** "✓ materialized" → "✓ queued" — the same posture that drove the rename in the DB. Names should tell the truth at every layer.

## Cost summary

**$0.00** in LLM costs. Pure substrate + UI plumbing. Bridge restart ~3s; UI rebuild + restart ~10s. Migrate idempotency confirmed: re-running shows *"substrate is current (97 files; 0 applied; 97 skipped; 0 drift)."*

## What's left from the A lane

Lane A is complete. The two remaining substrate carry-forwards (yaml.rs refactor + Phase A pgrx longjmp) are dedicated-session items, not cleanup-pulse items, and weren't included in Lane A.

## Pattern reinforced

The "decision question → ratified path → migration → smoke → commit → journal" rhythm is becoming repetitive in a good way. i2 and i3 each took ~25 min from question to journal. The ledger removes friction; the smoke catches regressions; the ratification keeps scope honest. Eight migrations through the new pipeline now (i1, i2, i3, plus the five backfilled "i*" slots through stewards-cli's --backfill).

## Closing

The substrate doesn't bookkeep gracefully without honest names. Today's win is small in lines and large in posture: a name lying about what a value means is a debt that compounds. The ledger made it cheap to pay it back.
