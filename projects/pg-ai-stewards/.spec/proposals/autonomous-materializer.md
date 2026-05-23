---
title: pg-ai-stewards — autonomous materializer (close the disk-write loop)
date: 2026-05-22
status: ratified — building
workstream: WS5
relates_to:
  - .spec/journal/2026-05-20-substrate-bridge-stall-recovery.md
  - extension/i7-apply-agent-proposal-direct-file-queue.sql
  - cmd/stewards-cli/main.go (runMaterializeWrites)
---

# Autonomous materializer — close the disk-write loop

## I. The gap

`stewards.pending_file_writes` is drained ONLY by the git pre-commit
hook calling `stewards-cli materialize-writes`. Today's symptom: the
2026-05-21 daily-digest auto-fired at 07:00, completed its full
pipeline + verified maturity, enqueued the file write — and then sat
in the table until the next human commit.

That's "autonomous" only up to the disk boundary. Scheduled pipelines
running while the human sleeps don't produce visible artifacts until
the human shows up. Three substrate-shape items named in the bridge-
stall journal called this out; this is the canonical fix for one of
them.

## II. Decisions ratified (2026-05-22)

| # | Decision | Choice |
|---|---|---|
| **D-AM-1** | Where the tick lives | **Bridge (Go)** — goroutine in `pg-ai-stewards-bridge` |
| **D-AM-2** | Workspace mount scope | **Switch `/workspace` from `ro` → `rw`** |
| **D-AM-3** | Trigger model | **LISTEN/NOTIFY + 60s safety poll** |

Rationale on each:

- **Bridge (Go) over Rust bgworker** — bridge already has the tick-loop
  pattern, already deals with workspace I/O (the YT-T `/opt/yt/yt:rw`
  mount established the precedent), and `runMaterializeWrites` is
  already Go. The Rust bgworker would need a fresh write path + a new
  workspace mount on the pg container, more code for the same outcome.
- **`rw` mount, not narrow per-dir** — the `pending_file_writes`
  `target_path` column is arbitrary; the existing `runMaterializeWrites`
  validates paths against `--repo-root`; the `write_mode CHECK` (append
  / create) is the real defense layer. Narrow per-dir mounts would
  require updating compose every time a pipeline writes to a new
  subdir.
- **LISTEN/NOTIFY + poll** — substrate already uses `LISTEN
  stewards_mcp_proxy` for tool-call routing. Adding
  `LISTEN stewards_pending_file_write` mirrors that pattern: NOTIFY on
  INSERT, drain when fired, plus a 60s safety poll in case NOTIFY is
  ever dropped (server restart, network blip).

## III. What changes

### III.1 Compose mount

`projects/pg-ai-stewards/extension/docker-compose.yaml`:

```diff
       - ../../..:/workspace:ro
+      - ../../..:/workspace:rw
       - ../../../yt:/opt/yt/yt:rw
```

Comment update: drop the "Read-only mount of the repo root at
/workspace so fs-read-mcp can serve scoped reads" framing; the
materializer is the legitimate writer.

### III.2 SQL: NOTIFY on pending file writes

New migration `pe9-pending-file-writes-notify.sql`:

```sql
CREATE OR REPLACE FUNCTION stewards.notify_pending_file_write()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('stewards_pending_file_write', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pending_file_writes_notify
AFTER INSERT ON stewards.pending_file_writes
FOR EACH ROW EXECUTE FUNCTION stewards.notify_pending_file_write();
```

Payload is the id so the bridge knows what to drain. The bridge can
ignore the id and just drain everything pending — payload exists for
future targeted handling.

### III.3 Bridge: materializer goroutine

`projects/pg-ai-stewards/cmd/stewards-mcp/...` (bridge code):

- New goroutine `materializerLoop(ctx, db, repoRoot)` that:
  - `LISTEN stewards_pending_file_write`
  - `time.NewTicker(60 * time.Second)` for safety poll
  - On either signal: call into a refactored `materializeWrites(ctx, db, repoRoot)` library function (extracted from the CLI's `runMaterializeWrites`)
  - Log each drain with the same `bridge run [N]` format used elsewhere
- Started from the bridge's main alongside the worker slots
- Configurable: `STEWARDS_MATERIALIZE_DISABLED=1` env disables the loop
  (so the pre-commit hook can still own the materializer in dev if
  someone wants — but production keeps it on)
- Reads `STEWARDS_REPO_ROOT` env (defaults to `/workspace` inside
  container) for path resolution

### III.4 CLI refactor

`cmd/stewards-cli/main.go` `runMaterializeWrites` body moves to
`internal/materialize/materialize.go` (or similar) as
`func Drain(ctx, db, repoRoot, opts) (stats, error)`. The CLI keeps
the same UX but calls into the library; the bridge does too.

### III.5 Git pre-commit hook stays as belt-and-suspenders

For now. Even with autonomous draining, the hook running
`materialize-writes` on commit is a useful safety net (if the bridge
is restarted mid-cycle, or if running locally outside Docker). Once
the bridge loop has soaked for a week without incident the hook can
move to "skip if container is up and recent drain logged."

## IV. Verification plan

1. `pe9-pending-file-writes-notify.sql` applied; trigger visible in
   `\d stewards.pending_file_writes`
2. Bridge rebuilt with materializer goroutine + mount changed to `rw`
3. Smoke: `INSERT INTO stewards.pending_file_writes ...` with a test
   path; observe the file lands on disk within ~1 second of the INSERT
   (NOTIFY path)
4. Smoke: `docker compose restart bridge` while a pending row exists;
   within 60s the safety poll drains it (NOTIFY-missed path)
5. End-to-end: dispatch a thummim work_item via `work_item_dispatch_stage`;
   watch the markdown file appear in `research/dictionary/` without
   any git commit happening
6. Bridge logs show drain activity in the expected format

## V. Risks

- **Repo write access from a daemon process.** The bridge container
  runs `stewards-mcp` as PID 1, plus the MCP server child processes
  for tool calls. If any child has a write-shaped vulnerability, the
  rw mount widens its blast radius. Mitigation: the materializer's
  path validation already constrains `target_path` to repo-rooted
  paths and won't traverse `..`; bridge child processes don't share
  the host repo path knowledge.
- **Concurrent drain with git pre-commit.** If the bridge drains row N
  while the human runs `git commit` and the hook also tries N: `FOR
  UPDATE SKIP LOCKED` in `runMaterializeWrites` means only one wins,
  the other sees zero pending. Safe.
- **Trigger overhead on busy tables.** `pending_file_writes` is
  low-volume (one row per verified work_item). Trigger cost is
  negligible.
- **NOTIFY payload size limit (8KB).** We send only the id — well
  under. If we ever extend the payload, watch the limit.

## VI. Out of scope

- Real-time UI updates for materialized files (would be a separate
  NOTIFY → SSE bridge to the stewards-ui)
- Materializer dry-run mode in the bridge (CLI keeps `--dry-run`;
  bridge always drains)
- Path-scope restrictions beyond the existing `--repo-root` validation
  (the CHECK constraint covers create-vs-append; nothing prevents a
  malicious INSERT from writing to e.g. `.git/hooks/pre-commit` — but
  the threat model assumes the database is trusted)
- Multi-host coordination (single bridge per substrate deploy; if we
  ever scale horizontally, the FOR UPDATE SKIP LOCKED pattern already
  handles two bridges draining the same table safely)
