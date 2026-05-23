---
date: 2026-05-22
session_window: late evening (~20:00 CDT)
workstream: WS5
status: shipped + verified end-to-end
commits:
  - 767386a — substrate(am1): autonomous materializer — close the disk-write loop
relates_to:
  - .spec/journal/2026-05-20-substrate-bridge-stall-recovery.md
  - .spec/journal/2026-05-22-dokploy-go-sum-sweep.md
  - projects/pg-ai-stewards/.spec/proposals/autonomous-materializer.md
---

# 2026-05-22 — Autonomous materializer (am1)

## What this closes

The 2026-05-20 bridge-stall journal named three substrate-shape items;
this session built the fix for one of them: *materializer-only-on-git-
commit cadence — autonomous batches enqueue file writes but don't
reach disk until the next human commit.*

Trigger was a Michael question: *"I thought that pg-ai-stewards had
the ability to materialize the files on their own instead of through
git commit hooks? can we give it the power/ability to drop those into
the file system on it's own?"*

He remembered the capability (the bgworker tick infrastructure + the
`stewards-cli materialize-writes` CLI both existed). What was missing
was the wiring between them.

## The shape of the fix

Three artifacts:

1. **`am1-pending-file-writes-notify.sql`** — AFTER INSERT trigger
   on `stewards.pending_file_writes` that fires
   `pg_notify('stewards_pending_file_write', NEW.id::text)`. Registered
   in the `extension_sql_file!` chain in `lib.rs` so future pg rebuilds
   replay it cleanly.

2. **`cmd/stewards-mcp/materializer.go`** — new goroutine in the
   bridge that LISTENs on the channel, drains on NOTIFY, with a 60s
   safety poll for the missed-NOTIFY case. Implementation choice that
   surprised me: instead of duplicating the drain algorithm in Go,
   the goroutine just `exec`s `stewards-cli materialize-writes
   --repo-root /workspace`. The CLI stays the source-of-truth; the
   bridge is the new tick caller. No code-sync debt.

3. **`/workspace` mount: `ro` → `rw`** in `docker-compose.yaml`. The
   load-bearing change. The bridge can now actually write to the repo.

Plus an Adjacent Surface Audit catch: the 1828 backend's go.mod was
added to `go.work` on 2026-05-20 but never to the bridge.Dockerfile's
COPY list. Latent breakage that surfaced when the build was finally
attempted today. Same shape as the morning's go.sum sweep — modules
added to `go.work` need to be propagated to every consumer's build
context. Fixed inline.

## Three decisions ratified upfront

Per substrate C-F cadence, AskUserQuestion before SQL hit disk:

- **D-AM-1: Tick location** — Bridge (Go), not Rust bgworker. The
  bridge already has the tick-loop pattern + workspace I/O via the
  YT-T precedent. The Rust bgworker would have needed a fresh write
  path + a new pg-container mount for no architectural gain.
- **D-AM-2: Mount scope** — Full `/workspace:rw`, not narrow per-dir
  mounts. The materializer's `--repo-root` validation + the
  `write_mode CHECK` constraint are the real safety boundary; mount
  flag is theater compared to those.
- **D-AM-3: Trigger model** — LISTEN/NOTIFY + 60s safety poll. Mirrors
  the substrate's existing `LISTEN stewards_mcp_proxy` pattern.

## Live verification

Three smoke checks, all green:

1. **Bridge restart logs the new goroutine:**
   ```
   materializer: LISTENing on stewards_pending_file_write +
   1m0s safety poll (repo-root=/workspace)
   ```

2. **Startup drain catches the science-news-weekly post** that had
   been sitting in `pending_file_writes` since 23:53 the night
   before:
   ```
   materializer: drain trigger=startup ok: ok #66
   (auto_materialize_on_verified, create) →
   /workspace/research/science-news-weekly--2026-05-22-2353.md
   ```
   The file landed on the host filesystem visible to git. **First
   autonomously-materialized substrate output**, committed in
   `767386a`.

3. **NOTIFY-path smoke** — `INSERT INTO pending_file_writes ...`
   row #68. Bridge log within ~2 seconds:
   ```
   materializer: drain trigger=notify ok: ok #68 (am1-smoke, create)
   → /workspace/tmp/am1-smoke-notify-test.txt
   ```
   File on disk, content matches, cleaned up.

## What this changes about the substrate's day-to-day

Before today: a scheduled pipeline like `ai-news-7am` would fire at
07:00, generate its digest, reach maturity=verified, enqueue the
file write — and then **wait silently** for a human to run `git
commit`. Files lived in the DB only. Anyone checking the workspace
for "what did the substrate do overnight" would see nothing.

After today: same pipeline fires at 07:00, same path through verify,
same enqueue. **NOTIFY fires; the bridge drains within seconds; the
file is on disk by 07:00:02.** Morning git status shows the new file
as untracked. The substrate's autonomy now reaches all the way to the
filesystem.

## Honest carry-forward

Three things named, not built:

- **Pre-commit hook stays as belt-and-suspenders for now.** The
  proposal says revisit after a week of soak without incident. If the
  bridge ever stalls again (the 2026-05-20 pattern), the hook is the
  fallback. Once we trust the goroutine, the hook can become "skip
  if bridge has logged a recent drain."

- **Multi-host coordination** isn't in v1. If we ever run two bridge
  containers, the existing `FOR UPDATE SKIP LOCKED` in
  `runMaterializeWrites` already handles concurrent drainers safely
  — but we don't currently have that topology and the materializer
  doesn't claim leadership.

- **Real-time UI updates** for newly-materialized files would be a
  natural follow-up — same NOTIFY channel could feed a SSE bridge
  to stewards-ui, showing "new substrate output available" without
  page refresh. Phase shape for a future stewards-ui evolution
  session.

## Adjacent observation

The bridge build had been broken since 2026-05-20 (when 1828's
backend joined `go.work`) but nobody noticed because nobody rebuilt
the bridge. The build only fails when attempted. This morning's go.sum
sweep + this evening's go.work fix are the same lesson surfacing
twice in twelve hours: **`go.work` membership is a build-context
dependency**, and adding a module to it without auditing every
consumer's Dockerfile is exactly the kind of latent breakage that
sits silent until someone tries to ship.

Worth a substrate watchman rule eventually: "if go.work changes,
verify every Docker build that copies it still resolves." Filing
as an observation, not yet a build.

## Why this matters

The substrate's promise has always been *autonomy with stewardship*
— the agent does the work, the human watches the watching. Until
today, "the work" stopped at the database boundary. The human had to
*notice* that there was work in pending_file_writes and run a commit
to flush it.

Now the substrate hands the work to disk on its own. The next morning
when Michael wakes up, the overnight digests are there. That's the
small step. The larger step it makes possible: a substrate that can
operate for hours or days at a time with the human reviewing the
materialized artifacts, not the database state. That changes the
human-AI rhythm meaningfully.
