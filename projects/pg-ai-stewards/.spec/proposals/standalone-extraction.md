---
title: Standalone repo extraction — making pg-ai-stewards independently buildable
date: 2026-05-29
status: proposed (spec only — not started)
author: michael + claude
---

# Standalone extraction

## Goal

Make `pg-ai-stewards` a self-contained public repo that builds from a bare
clone — no parent monorepo. The documentation pass (README, QUICKSTART,
architecture, LICENSE, CONTRIBUTING) shipped 2026-05-29; this proposal covers
the **build/dependency decoupling** that documentation alone can't fix.

## Current coupling (the honest inventory)

Investigated 2026-05-29. Good news first: **the Go modules have no sibling
*source* dependencies** — `cmd/stewards-mcp`, `cmd/stewards-cli`,
`cmd/fs-read-mcp` are independent modules that merely coexist in the parent
`go.work`. No `replace`/`require` of sibling modules. So extraction is mostly
plumbing, not code rewrites. The coupling points:

1. **Shared `go.work` at repo root.** The parent workspace's `go.work` lists
   all modules (scripts/becoming, 1828, gospel-*, …) alongside the substrate's
   cmd modules. Standalone needs its own `go.work` listing only this repo's
   modules.

2. **Docker build context = repo root (`context: ../../..`).** Both
   `extension/bridge.Dockerfile` and `extension/ui.Dockerfile` build against the
   whole workspace and `COPY` sibling `go.mod`/`go.sum` files to satisfy the
   shared `go.work`. Standalone needs `context: .` (this repo) + a self-contained
   COPY list.

3. **The bridge bundles workspace MCP servers.** `bridge.Dockerfile` COPYs +
   builds sibling MCP servers it spawns: `fetch-md-mcp`, `git-mcp`,
   `webster-mcp`, `byu-citations`, `yt-mcp`, `search-mcp`, `gospel-engine-v2`,
   `becoming/cmd/mcp`. These are workspace-specific. Standalone should bundle
   only the substrate's own servers (`stewards-mcp`, `fs-read-mcp`) and make the
   rest **pluggable/optional** (configured MCP servers, not hardcoded COPYs).

4. **The web UI lives in a sibling dir.** `scripts/stewards-ui/` is referenced
   by `ui.Dockerfile`. Standalone needs it moved into the repo (e.g. `ui/`).

5. **gospel-engine-v2 cross-DB integration.** Already a soft dependency
   (`ENGINE_URL`, returns cleanly when unreachable). Just needs to be documented
   as optional and removed from the default compose (it's the author's corpus,
   not a substrate requirement).

## Proposed work (phased)

### E.1 — Vendor the UI into the repo
- `git mv scripts/stewards-ui projects/pg-ai-stewards/ui` (history preserved).
- Update `ui.Dockerfile` paths.

### E.2 — Own go.work + self-contained build contexts
- Add `projects/pg-ai-stewards/go.work` listing only this repo's modules
  (`cmd/*`, `ui/`).
- Rework `bridge.Dockerfile` + `ui.Dockerfile` to `context: .` (the repo root
  after extraction) with COPY lists limited to repo files.
- Verify the `am1` lesson doesn't recur: every `extension_sql_file!` reference
  in `src/lib.rs` must be in the Dockerfile COPY list. (J.11 fixed the one gap;
  add a CI check that diffs lib.rs refs vs. COPY list.)

### E.3 — Make bundled MCP servers pluggable
- Bundle only `stewards-mcp` + `fs-read-mcp` by default.
- Move the workspace MCP server set (webster, yt, gospel-engine, becoming, …)
  to an **optional overlay** — a documented compose override or an env-listed
  set of external MCP endpoints. The substrate's `mcp_servers` registry already
  stores these as rows; the bridge should read that registry rather than
  hardcode COPYs.

### E.4 — Optional corpus integration
- Drop `gospel-engine-v2` from the default `docker-compose.yaml`.
- Document the cross-DB / cross-corpus pattern as an opt-in (it already
  degrades cleanly when `ENGINE_URL` is unset).

### E.5 — Extraction mechanics
- New git repo (`git init` or `git subtree split` to preserve history).
- CI: build all three images from a bare clone; run the smoke suite
  (`extension/smoke/*.sql`).
- Decide: keep `.spec/` (provenance) in the public repo as `docs/history/`, or
  curate to a subset. (Current docs pass left `.spec/` in place + pointed
  README/CONTRIBUTING at `docs/history/`.)

## Open decisions

- **History depth in public:** ship all `.spec/journal/` + `.spec/proposals/`
  (full provenance, gospel-study-flavored) or curate? Leaning: keep, labeled
  "internal provenance" — honesty over polish, matches the project's character.
- **Subtree split vs. fresh init:** preserving commit history across the
  extraction is nice but the history is interleaved with the whole workspace.
  A `git subtree split -P projects/pg-ai-stewards` gives a clean lineage.
- **Repo name + org:** `pg-ai-stewards` under which GitHub owner?

## Not in scope

- Genericizing the gospel-study *examples* in docs — the 2026-05-29 docs pass
  already framed gospel-engine-v2 as "one example use." Further neutralizing of
  `.spec/` history is the E.5 decision above.

## Provenance

Surfaced during the 2026-05-29 public-OSS docs pass, when `docker compose build`
revealed the workspace build-context coupling (and the `am1` COPY-list gap,
fixed in J.11). Ratified scope: docs now, extraction specced for later.
