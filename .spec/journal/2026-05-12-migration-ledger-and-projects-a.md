---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Substrate migration ledger + Projects A formalization shipped"
status: shipped (e2e validated)
carry_forward:
  - "yaml.rs Rust parser refactor (rule-of-three triggered, Claude-only)"
  - "Phase A pgrx longjmp catch + 60s reaper"
  - "materialized_at semantics drift (~30min)"
  - "Projects B — full workspace + optional sub-git (deferred per ratification)"
  - "14 SC work_items still awaiting Michael's ratification (some now grandchildren-of-grandchildren via planning pipelines on the proposed laptop/esp32/bias work)"
  - "FK constraint on work_items.project_association → projects.slug (soft today; harden once stable)"
  - "Soak still paused for build; resume at session close"
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-migration-ledger-and-projects.md"
  - "../../projects/pg-ai-stewards/extension/h-ledger-1-schema-migrations.sql"
  - "../../projects/pg-ai-stewards/extension/i1-projects-table.sql"
  - "../../projects/pg-ai-stewards/cmd/stewards-cli/migrate.go"
---

# Substrate migration ledger + Projects A (2026-05-12)

Six commits this build pulse. Closes two concerns Michael surfaced after the h3-5 trigger-overwrite bug: (1) substrate needs a real migration pattern; (2) projects need to be first-class, not freeform string inferred from agent_planning inheritance.

## What shipped

### Proposal (commit `e27bfd2`)
Two related primitives ratified this session: **migration ledger** (matches ibeco.me's "DB on startup ensures migrations are applied" pattern) and **Projects A** (lightweight first; Brain v3 shape "B" deferred). Stewardship decisions inside the ratified scope: up-only migrations, lexical filename ordering, idempotent backfill, sha256 drift tracking, stewards-cli as runner, bridge entrypoint auto-runs on startup.

### Migration ledger (commits `53268df`, `852f89a`)

**`stewards.schema_migrations` table.** Single source of truth for which extension/*.sql files have run. Columns: name (PK), sha256, applied_at, notes. Helper functions `migration_is_applied(name)` + `migration_mark_applied(name, sha, notes)`.

**`stewards-cli migrate` subcommand.** Reads `extension/*.sql` in lexical order, computes sha256, checks the ledger, applies unrecorded files in `BEGIN/COMMIT` transactions, records on success. Flags: `--dry-run`, `--list`, `--target=NAME`, `--backfill`, `--repo-root`.

**Drift detection.** If a recorded migration's file sha256 differs from current contents on disk, migrator WARNS and skips. This is the safety net for the h3-5 regression class: editing an already-applied migration after the fact won't silently re-run with new content.

**Backfill.** 94 existing .sql files recorded as already-applied with `notes='backfilled'`. h-ledger-1 itself recorded manually as `manual-h-ledger-1` (chicken-and-egg: it creates the table). Total 95 in the ledger after backfill. Re-running `migrate` after backfill: *"substrate is current (95 files; 0 applied; 95 skipped; 0 drift)"* — fully idempotent.

**Bridge entrypoint hook** (`bridge-entrypoint.sh`):
```
bridge-entrypoint: running substrate migrations…
migrate: substrate is current (95 files; 0 applied; 95 skipped; 0 drift)
bridge-entrypoint: migrations done.
bridge-entrypoint: starting bridge daemon…
```
Substrate auto-current on every bridge restart. If migrations fail, entrypoint exits non-zero and the bridge doesn't start (operator inspects + fixes).

**Dockerfile change.** `bridge.Dockerfile` now COPYs `cmd/stewards-cli/` in full (was only go.mod stub before — sufficient for go.work, insufficient to build). Builds stewards-cli alongside stewards-mcp and ships the binary to `/usr/local/bin/`. ENTRYPOINT is the new shell script.

### Projects A (commits `40cfd9b`, `f69ab45`)

**`stewards.projects` table** (commit `40cfd9b`). First new-style migration through the ledger. Schema: `slug PK, name, description, root_directory (nullable, future B hook), archived, created_at, updated_at`. Soft reference (no FK on work_items.project_association initially) so existing data doesn't break. Backfill from `DISTINCT project_association` values: 2 rows (`space-center`, `pg-ai-stewards`).

End-to-end proof of the new ledger: a fresh .sql file dropped into `extension/`, the runner picked it up, applied it transactionally, recorded it with `notes='auto'`. Bridge restart from there: idempotent.

**5 backend endpoints** in `scripts/stewards-ui/api/projects.go`:
- `GET /api/projects/list` (optional `?include_archived=true`)
- `GET /api/projects/get?slug=X`
- `POST /api/projects/create` (slug regex validated)
- `POST /api/projects/update` (partial fields; txn-wrapped)
- `POST /api/projects/archive` (toggle)

`work_item_count` per project surfaced via subquery on the list endpoint. `new_work.go` gained `project_association` field; sets it post-create.

**Frontend.** New `/projects` route + Projects.vue page (~250 lines): create form, inline edit, archive toggle, work_item_count linking to filtered work_items view. NewWork.vue gained a project picker dropdown sourced from `/api/projects/list` (with "(no project)" default and a "manage ↗" link). App.vue sidebar entry. api.ts: 5 new methods + 2 types.

E2E smoke verified live: 2 backfilled projects with `work_item_count` (pg-ai-stewards: 5, space-center: 20), create + archive endpoints round-trip cleanly.

## Architecture wins

**Regression class closed.** The h3-5 "multiple files CREATE OR REPLACE the same trigger" pattern that bit us is now caught two ways: (1) sha256 drift detection warns when a recorded file changes; (2) up-only semantics mean re-running an already-applied file is a no-op (no overwrite). The "one canonical owner per shared substrate object" discipline from yesterday is now backed by tooling, not just memory.

**Substrate auto-currents on restart.** No manual `psql -f` workflow. Drop a new `.sql` file into `extension/`, `docker compose restart bridge`, done. The entrypoint logs every applied migration.

**Projects entity surfaces what was already there.** The 2 distinct `project_association` values we'd accumulated through agent inheritance now have a real table, a UI, and counts. Backfill respected existing data; FK harden-up is a separate future migration once the table is stable.

## Cost summary

This build pulse: **$0.00** in LLM costs (entirely substrate + UI work). The migration ledger validation showed 95 files re-checked in <1 second; first new-style migration applied + recorded in another second. Bridge entrypoint adds ~2 seconds to container restart.

## Carry-forward (unchanged from prior sessions)

- **yaml.rs Rust parser refactor** — rule-of-three triggered. Claude-only per kimi-trust ratification. ~1 session.
- **Phase A pgrx longjmp catch + 60s periodic reaper** — Claude-only. H.1.5a soft-fail still stable. ~1 session.
- **materialized_at naming drift** — pre-existing; ~30min when bothered.
- **Projects B (full workspace + optional sub-git)** — deferred per ratification. `root_directory` column is the hook.
- **FK constraint on work_items.project_association** — soft today; harden once projects table proves stable.
- **14 SC work_items still awaiting Michael's ratification** — plus the dozen grandchildren from the planning runs on them.

## Pattern worth keeping

The 9-step build plan from the proposal mapped 1:1 to commits. Each step a clean checkpoint. Each commit's message documents what changed AND why. The pattern from H.3 (heavy proposal upfront, then incremental shipped commits) continues to pay dividends — every step's blast radius was bounded, every regression class identified before it bit.

## Closing

Yesterday: substrate noticed it had hands (planning pipeline). Today: substrate noticed it had bookkeeping (migration ledger) and that the bookkeeping needed structure (projects table). Both meta-improvements; both unblock the next batch's work.

Tuesday is still for the science center. The substrate is more ready every session.
