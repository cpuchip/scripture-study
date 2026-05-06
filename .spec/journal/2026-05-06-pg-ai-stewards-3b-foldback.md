# pg-ai-stewards: Phase 3b + Foldback Clearance
*2026-05-06*

## What this session was

A triage-and-closure session. Opus 4.7 ran out of credits mid-session;
Haiku 4.5 finished the AGE debugging loop. The work landed, but memory
and plan files were not updated (Haiku's known gap). This session
audited what actually shipped, confirmed it against the plan, and
closed the remaining open-forwards (3b + WS6 docs).

## What we confirmed shipped (in Haiku's commits)

### v0.2.0 foldback — DONE

All five SQL files (2-6a/b/c + 2-7a + 3a) folded into `extension/src/lib.rs`
via `extension_sql_file!`. Extension bumped from 0.1.0 to 0.2.0.

The foldback surfaced a nasty AGE-specific bug: calling
`set_config('search_path', 'ag_catalog,...', true)` from inside a
`CREATE EXTENSION` transaction corrupts the session search_path for the
remainder of the install. This causes every PGRX-emitted `pg_extern`
`CREATE FUNCTION` (which uses unqualified names) to land in `ag_catalog`
instead of `stewards`. Fixed by extracting workstream seeds into
`init/01-seed-workstreams.sql` as a post-install script. Documented
as AGE-QUIRKS #9.

### AGE-QUIRKS new entries (also Haiku's session)

- **#6:** `cypher()` 3rd argument must be a bound parameter — inline
  expressions rejected at parse time.
- **#7:** `#>>` jsonb-path operator does not pass through agtype scalars —
  requires explicit `::text` cast.
- **#8:** PL/pgSQL OUT params with the same name as a `RETURNING` column
  produce "column reference is ambiguous".
- **#9:** `set_config('search_path', ..., true)` called from within
  `CREATE EXTENSION` leaks into PGRX-emitted `pg_extern` declarations.

## What this session shipped

### Phase 3b — `response_format` injection + big-doc timeout

**Files changed:**
- `extension/src/lib.rs` — `response_format jsonb` column added to
  `stewards.agents`; `dry_run_chat` injects it when non-NULL.
- `extension/3a-watchman-pass.sql` — `watchman-consolidator` INSERT and
  `ON CONFLICT DO UPDATE` updated to include `response_format`;
  seeded with `'{"type": "json_object"}'::jsonb`.
- `extension/pg_ai_stewards--0.2.0.sql` — mirrored.
- Applied live to running DB via `ALTER TABLE` + `UPDATE` + `CREATE OR REPLACE FUNCTION`.
- Rebuilt container image so changes survive restarts.

**Verified:** `SELECT payload->'body'->'response_format' FROM stewards.work_queue ORDER BY id DESC LIMIT 1;` → `{"type": "json_object"}`. Watchman `--dry-run` pass on `charity` succeeded (clean verdict, 2041 in / 3866 out).

The 120s timeout fix (bump to 600s + `--max-input-chars` flag) was
confirmed present from the Haiku session; not re-done here.

### WS6 — AGE-QUIRKS #9 documented

Added entry #9 to `projects/pg-ai-stewards/docs/AGE-QUIRKS.md` with
full symptom, workaround, and first-seen metadata. Also updated the
index table at the top of the file.

## What Haiku missed (and this session caught)

1. Memory (`active.md`) was not updated after either commit.
2. `phases.md` still showed 3b as "not started" and foldback as "debt."
3. No journal entry was written for the post-3a / foldback session.

Lesson: Haiku executes code-level changes reliably. It does not write
session memory without explicit instruction. When Opus hands off to
Haiku mid-session, always prompt it to write the memory files at close.

## Next session carry-forwards

| Priority | Item |
|----------|------|
| 1 | **Phase 2.7b** — Watchman bgworker (transcribe Go `WatchmanPass` into a Postgres pgrx scheduled worker). No design work left. |
| 2 | **ws6 AGE upstream PRs** — issues/PRs filed on #2, #6, #7 (the three `bug-candidate` entries). #9 is our-mistake, not upstream. |
| 3 | **Phase 3c** — `stewards.pipelines` + `stewards.work_items` tables (unblocks the study-writing pipeline POC). |

## What's still solid

- `stewards.dirty_queue` → `watchman pass` → `stewards.verdicts` pipeline
  is end-to-end proven with a real model call.
- Extension image is v0.2.0, container rebuilt, all migrations baked in.
- `response_format` is now infrastructure — any future agent that needs
  JSON-only output sets it in the agents table seed, zero CLI changes required.
