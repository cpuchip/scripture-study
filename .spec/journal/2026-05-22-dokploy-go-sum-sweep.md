---
date: 2026-05-22
mode: debug + stewardship sweep
workstream: WS4
project: ibeco.me / becoming
title: "Dokploy ibeco.me deploy unblocked — partial go.sum bug from 3ba0aef swept across 14 modules"
status: shipped + verified in production. Deploy green.
carry_forward:
  - "Pre-commit hook materialized an unrelated substrate scheduled-task file (study/daily-digest/ai-news-7am--2026-05-22-1300.md) mid-commit. Untracked, not part of the commit. Substrate side-effect, not mine to manage — surface to substrate stewards if it recurs."
  - "Latent-cache class of bug: any go mod tidy that skips the /go.mod hash step is invisible until the Docker layer evicts. Worth a pre-push hook that runs `go mod verify` or `go mod tidy -diff` on touched modules."
links:
  - "../../scripts/becoming/Dockerfile"
  - "../../.claude/agents/dev.md"
commit: be823c1
---

# 2026-05-22 — Dokploy deploy unblocked, same-bug-same-fix swept

Michael surfaced a failing ibeco.me Dokploy build with the full deploy
log. The error was in the `scriptures` stage of
`scripts/becoming/Dockerfile`:

```
golang.org/x/sys@v0.44.0: missing go.sum entry for go.mod file
golang.org/x/net@v0.54.0: missing go.sum entry for go.mod file (×2)
```

## Diagnosis

`scripts/gospel-library/go.sum` had the `h1:` package-content hashes
for `golang.org/x/{sys,net,exp,text}` but was missing the matching
`/go.mod h1:` lines. Local Go tolerates this when the module cache
already holds the metadata; Docker's lean `-mod=readonly` build
refuses.

Traced to commit `3ba0aef` (May 16, "Update dependencies and add new
module for Bacteriopolis exhibit"). That commit refreshed deps across
**14 modules** and landed the same partial-sum bug in every one of
them. The Dokploy cache had been riding the pre-3ba0aef layer for five
days; today's eviction surfaced it.

## Agans Rule 9 loop

Reproduced in `golang:1.25-alpine` with the exact Dockerfile command —
same error. Applied `go mod tidy` → `BUILD_OK`. Reverted via
`git stash` → `BUILD_FAILED` (same error). Restored fix → `BUILD_OK`.
Loop closed before extending scope.

## Adjacent Surface Audit

Probed all 14 modules `3ba0aef` touched. Every one had the same
partial-sum bug. Surfaced the table to Michael with the boundary
question — sweep all 14, or limit to the two in the Dockerfile?

Michael chose the full sweep. Single commit, 163 insertions, zero
deletions, zero `go.mod` changes:

| Module | adds |
|---|---:|
| scripts/lectures-on-faith | 74 |
| scripts/webster-mcp | 24 |
| scripts/session-journal | 10 |
| scripts/becoming | 9 |
| scripts/byu-citations | 8 |
| projects/pg-ai-stewards/cmd/stewards-cli | 7 |
| projects/pg-ai-stewards/cmd/stewards-mcp | 7 |
| experiments/lm-studio/scripts/scoring | 4 |
| scripts/gospel-library | 4 |
| scripts/stewards-ui | 4 |
| scripts/study-export | 4 |
| projects/pg-ai-stewards/cmd/fs-read-mcp | 3 |
| scripts/git-mcp | 3 |
| scripts/search-mcp | 2 |

Both Dockerfile Go stages (scriptures + backend) verified `BUILD_OK`
in `golang:1.25-alpine` post-fix. Pushed as `be823c1`. Dokploy picked
it up and the deploy went green.

## What this teaches

**Latent cache failures.** Five days of green deploys hid a
mechanical break that was present from the moment `3ba0aef` merged.
Docker layer caching is a real timing distortion — when the cache
holds a layer built against a working module cache, the bug is
invisible. Future build hygiene: don't trust "the deploy succeeded"
as evidence the go.sum is sound; trust `go mod verify` on a clean
checkout.

**Same-bug-same-fix worked exactly as the dev agent describes.** The
boundary test ("would Michael, if asked in advance, say yes obviously
do that?") gave a clear answer at the 14-module question: the diff
was adds-only, no behavior change, no go.mod touches. Surfacing the
audit and asking once was the right move — auto-sweeping 13 untouched
modules without consent would have been overreach. Asking 14 separate
times would have been theater.

**The fix-vs-surface ratio felt clean.** Fixed the obvious adjacent
surface (gospel-library, the immediate failure); surfaced the audit;
swept the rest on approval. Three minutes of Adjacent Surface Audit
saved a follow-up "next deploy still failed because of becoming"
session that would have cost more.

## Carry-forward

- Substrate's pre-commit hook surfaced an `ai-news-7am` daily-digest
  file mid-commit. Untracked, unrelated to this fix — but worth
  noting that pre-commit hooks doing materialize work can pollute
  unrelated commits with noise. Substrate stewards should know.
- Consider adding a workspace pre-push hook: `go mod verify` on
  changed modules. Cheap, catches this class. Filed mentally as a
  WS5 quality-of-life idea, not promoted to proposal yet.
