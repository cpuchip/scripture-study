---
name: pgrx-extension-bump
description: Lesson #3 fix for projects/pg-ai-stewards/extension. Run scripts/bump-extension.sh after docker compose build pg to refresh new pg_extern function registrations in the live db (replaces the manual CREATE FUNCTION AS libdir/pg_ai_stewards wrapper workaround). Load when working in the extension directory and rebuilding the pg image.
---

# pgrx-extension-bump — Lesson #3 fix

## When to load

- Working in `projects/pg-ai-stewards/extension/` AND about to run `docker compose build pg`
- After running `docker compose build pg` and before testing new pg_extern functions in psql
- When you see "function does not exist" after a rebuild
- When you're about to manually `CREATE FUNCTION ... AS '$libdir/pg_ai_stewards', '<name>_wrapper'` — that's the Lesson #3 anti-pattern; use the bump script instead

## What it does

`projects/pg-ai-stewards/extension/scripts/bump-extension.sh`:
1. Reads the bundled SQL inside `pg-ai-stewards-dev` (whatever version the most recent image was built with)
2. Extracts pgrx-generated `CREATE  FUNCTION` blocks (recognizable by the `-- pg_ai_stewards::module::name` comment + the two-space pgrx signature) using a small Python regex
3. Rewrites each as `CREATE OR REPLACE FUNCTION` and substitutes `MODULE_PATHNAME` → `$libdir/pg_ai_stewards`
4. Applies the resulting SQL inside `BEGIN; ... COMMIT;` with `search_path = stewards, public`

Net effect: new `#[pg_extern]` functions get registered in `pg_proc` after a rebuild without needing `ALTER EXTENSION UPDATE TO` (which would require dealing with non-idempotent `CREATE TABLE` statements in the bundled SQL).

## What this DOESN'T do (intentionally)

- **Doesn't bump the extension version.** No upgrade script written; no Cargo.toml or .control edits. The functions land in `pg_proc` but aren't tracked in `pg_depend` as extension members. Functional but slightly drifty for production; fine for dev iteration.
- **Doesn't touch tables/types/seeds.** Early-phase tables in the bundled SQL use raw `CREATE TABLE` (no `IF NOT EXISTS`), so re-running them on a populated db fails. New tables ship via separate `5*.sql` files applied through the substrate's existing live-migration pattern.
- **Doesn't restart anything.** Just patches `pg_proc`. Bgworker keeps running.

For new SQL files (5*.sql), keep using the substrate's live-migration pattern:
```bash
docker cp 5x-something.sql pg-ai-stewards-dev:/tmp/x.sql
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -f /tmp/x.sql
```

## Why this is needed (Lesson #3 in detail)

When `pg-ai-stewards-dev` restarts after a `docker compose build pg`:
- The pg data volume persists across restarts
- `pg_extension.extversion` stays at whatever was previously installed
- `CREATE EXTENSION pg_ai_stewards` is a no-op — the extension is already installed at that version
- Even though the new image's bundled SQL has new function definitions, postgres doesn't re-read them
- New `#[pg_extern]` functions exist in the .so but have no `CREATE FUNCTION` row in `pg_proc` — calling them fails

The manual workaround (creating each function by hand with `AS '$libdir/pg_ai_stewards', '<name>_wrapper'`) works but is error-prone. The bump script automates it.

## Why this is needed (Lesson #3 in detail)

When the pg-ai-stewards-dev container restarts after a `docker compose build pg`:
- The pg data volume persists across restarts
- `pg_extension.extversion` is still `0.2.0` (whatever was previously installed)
- `CREATE EXTENSION pg_ai_stewards` is a no-op because the extension is already installed at that version
- Even though the new image's bundled SQL has new function definitions, postgres doesn't re-read them
- New pg_extern functions exist in the .so but have no `CREATE FUNCTION` row in `pg_proc` — calling them fails with "function does not exist"

The manual workaround (creating each function by hand with `AS '$libdir/pg_ai_stewards', '<name>_wrapper'`) works but is error-prone — easy to miss a function, easy to mistype a wrapper name, doesn't update tables/views/triggers added via extension_sql_file!.

The bump script is the correct fix.

## Workflow

```bash
# After editing pgrx code or 5*.sql files:
cd projects/pg-ai-stewards/extension
docker compose build pg

# Restart the container if you also changed bgworker.rs or tools.rs:
docker compose down && docker compose up -d pg ui

# Apply the function refresh (idempotent; safe to re-run):
./scripts/bump-extension.sh

# Now test — new functions are live
docker exec pg-ai-stewards-dev psql -U stewards -d stewards \
    -c "SELECT stewards.your_new_function(...)"
```

## Cross-project safety

This skill is scoped to `projects/pg-ai-stewards/extension/` only. Other extensions (none today, but if you ever spawn one) need their own bump script + skill. The `scripts/bump-extension.sh` reads the container name from `STEWARDS_CONTAINER` env var (default `pg-ai-stewards-dev`) so a sibling extension would set its own.

## Hook automation

`.claude/settings.json` PostToolUse hook auto-runs the bump script after a successful `docker compose build pg` from the extension dir. You usually don't need to invoke the script manually — but knowing it exists matters for debugging.

## Naming caveat

The script is named `bump-extension.sh` for historical reasons (the original design did bump versions). It currently doesn't bump anything — it just refreshes pg_extern function registrations. If you rename, update the hook in `.claude/settings.json` too.
