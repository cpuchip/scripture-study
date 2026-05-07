# pg-ai-stewards: image rebuild — foldback chain becomes image-resident

*2026-05-07 (Claude Code, Opus 4.7)*

## What this session was

The substrate had 14 SQL files folded into `extension/src/lib.rs` via
`extension_sql_file!`, but the running image had only the first 5
baked in (the rest were live-applied via `psql -f`). One disaster —
volume wipe, host migration, fresh clone — and 9 sub-phases of
schema would have to be re-applied by hand.

The task: rebuild the image so the foldback chain is image-resident,
and verify it via `CREATE EXTENSION` on a fresh database.

## What we did

**Foldback verification.** All 14 production SQL files
(2-6a/b/c, 2-7a, 3a, 2-7b1/b2/b3/b4, 3c1, 3c2, 3c2-5, 3c3, 3c3-1)
present in `lib.rs` with intact `requires =` chain. Dockerfile COPY
list matches. No drift.

**Build.** `docker compose build pg` — Stage 1 cargo-pgrx packaged
the extension in 30s; Stage 2 layered it onto pgvector + AGE on PG18.
Discovered 28 SQL entities (4 functions + 24 sqls) into
`pg_ai_stewards--0.2.0.sql`. Image at
`sha256:106f9297903ccfe1ae7cf6fcd307a4725ca44093ece4cee27b1bd17391bbfe0f`.

**Smoke test caught a real bug.**
`docker run --rm pg-ai-stewards-dev:pg18` + `CREATE EXTENSION pg_ai_stewards CASCADE`
failed with:

```
ERROR: INSERT has more target columns than expressions
LINE 45: ...
QUERY: INSERT INTO stewards.agents
       (family, model_match, description, mode, prompt,
        temperature, top_p, response_format, steps)
```

The first VALUES tuple in `3a-watchman-pass.sql` had 8 values for 9
columns — the `response_format` value was missing for the default
`watchman-consolidator` variant (the `kimi-*` variant had it).

**Why this was masked.** The live container's
`watchman-consolidator '*'` row had `response_format =
'{"type":"json_object"}'` in the database, because at some point we
patched it via `UPDATE` after adding the column (commit
`eb42f47 feat: add response_format field to agents`). The repo SQL
file was edited to add `response_format` to the column list and to
the `kimi-*` row, but the `'*'` row's `VALUES` tuple was never
updated. ON CONFLICT-DO-UPDATE on a pre-existing row hides this
silently; CREATE EXTENSION on a fresh database surfaces it
structurally.

This is a textbook "live patches drift from source." Rebuild from
source was the only way to catch it.

**Fix.** Added `'{"type": "json_object"}'::jsonb,` between `NULL,`
(top_p) and `1` (steps) in the default variant's `VALUES` tuple.

**Re-smoke.** Rebuild → `CREATE EXTENSION CASCADE` →
applied cleanly. Verified row counts on the fresh DB:

| table              | count |
|--------------------|-------|
| agents             | 4     |
| tool_defs          | 7     |
| agent_tool_perms   | 5     |
| pipelines          | 2     |
| watchman_config    | 1     |

All 14 sub-phases composed in one transaction, in correct dependency
order, on a database that had never seen the substrate before. The
foldback chain works end-to-end.

## What we did *not* do

**Did not swap the live container.** `pg-ai-stewards-dev` (live, on
the old image hash `8290f7f7cf3e`) is still running the soak —
hourly Watchman pressure passes draining the dirty queue. The
`pgdata` named volume is persistent, so the next natural restart
will boot on the new image with all soak history intact. No need to
proactively interrupt working work.

## What we learned

1. **Source-as-truth requires periodic rebuilds.** Live-applied SQL
   plus repo edits can drift from each other indefinitely. The only
   forcing function is `CREATE EXTENSION` on a fresh database.

2. **Smoke test on fresh DB is the cheapest catch for this class of
   bug.** Took about 90 seconds end-to-end. Caught a structural bug
   that had been latent in the repo for two days.

3. **ON CONFLICT-DO-UPDATE on seed inserts hides drift.** When the
   conflict-target rows already exist (which they always will after
   the first run), the SET clause overrides with EXCLUDED values —
   but only for the listed columns. Column-count mismatches in
   the VALUES list don't fail because the path is never exercised.
   The fix isn't "stop using ON CONFLICT"; it's "rebuild and
   re-CREATE EXTENSION as a routine sanity check."

## Open

- **Live container swap.** Deferred until natural restart. New image
  is ready when it's needed.
- **Soak continues.** 5 passes in, ~99K in / 74K out, dirty_queue
  269 → 262 — converging as expected.

## Next

Image rebuild is done. The foldback debt is now image-resident.
Next substantive work is whatever the user picks next — likely
3c.3.3 (auto-promote completed work_items into `stewards.studies`)
or the soak observation window reaching its 7-day mark.
