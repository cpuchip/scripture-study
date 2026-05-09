---
date: 2026-05-09
agent: dev
session_kind: substantive
tags: [pg-ai-stewards, 3d, fetch-md, 3f, sandboxed-git, chromedp, autonomous]
priority: medium
carry_forward:
  - GitHub fine-grained PAT — Michael generates tomorrow; I plumb GITHUB_TOKEN into .env and 3d goes live
  - 3f local UI implementation — proposal ratified the vision; design questions named (stack: HTMX vs Alpine; service shape: separate vs fold into bridge; naming) wait for Michael's call before coding
  - Bridge log briefly showed `GITHUB_TOKEN=set` on first git-mcp spawn after rebuild but env was actually empty — phantom that I couldn't reproduce on subsequent spawns. Worth noting if it recurs.
  - 3e.5 gospel_passthrough still unbuilt; trivial now that bridge exists
  - 3e.4 v2 write-mutating inbound tools (work_item_create, watchman_pass_now) — useful but defer until v1 read-only proves out
  - Docker secrets migration (current setup uses .env file as ratified) — when shared infrastructure becomes the threat model
---

# 3d sandboxed git + fetch-md v2 + 3f proposal — autonomous run

Michael was tired from helping family move and asked me to take
stewardship over the carry-forward queue. He gave concrete answers
to all the design questions in two AskUserQuestion batches, then
let me at it for ~2 hours of autonomous work. Three pieces shipped
in one commit (plus the bridge image rebuild as a fourth implicit
deliverable).

## Pattern observed

The pattern that made this work: **plan doc first, then code**.
I wrote `.spec/scratch/carry-forward-plan-2026-05-09.md` with every
ratified decision, sequencing, and the explicit list of "what I will
not do without confirming." Having that file open meant I never had
to re-derive the spec mid-session, never had to wonder whether to
push or restart pg or touch the gospel-engine-v2 directory.

The plan doc also names the surface-back triggers — when I'd stop
and ask. None of them fired tonight. The work stayed inside the
boundaries Michael ratified.

## What landed

**3d sandboxed git-mcp v1.** Go MCP wrapper at `scripts/git-mcp/`,
8 tools, allow-list at the source. Three layers of discipline:

1. **Tool surface itself.** Forbidden ops (force-push, reset --hard,
   branch -D, rebase, tag) don't exist as tools. The agent has no
   handle on them.
2. **Argv-level refusal.** `git_push` validates the branch name
   against the agent/* namespace regex before any subprocess spawn.
   `git_branch_create` won't construct names outside the namespace.
   Protected branches (main, master, release/*) are refused at the
   tool layer.
3. **Subprocess isolation.** GITHUB_TOKEN read from the bridge's
   process env, never from tool args. Workdir constrained to
   `/tmp/stewards-git/<work-item-id>/` with `..`-traversal refused.

The Co-Authored-By trailer auto-appends with the agent family name
in the format Michael specified:
`Co-Authored-By: <agent-family>-via-pg-ai-stewards <agents@cpuchip.net>`.

PAT setup deferred to tomorrow per Michael. The build is complete;
the live test (clone, branch, commit, push, PR) triggers when
GITHUB_TOKEN lands in `.env`. Until then, local git ops
(status, branch, commit) work fine — only the network ops fail.

**fetch-md v2 with chromedp.** Optional `js: true` and `wait_ms`
params on all 4 tools. Default false keeps the existing fast HTTP
path unchanged. When true, launches headless Chromium via chromedp
with --no-sandbox + --headless flags. Bridge image gains
`apk add chromium` (~150MB). Verified with Wikipedia render in 3.3s.

**3f local UI proposal.** Captured Michael's pivot from
cloud-hosted a.ibeco.me to local-first single-user UI alongside the
docker stack. The doc covers threat model (mostly accidental
misclick, not adversarial), three architecture options (Go-served
single binary recommended, Vue/Vite split as v2 if needed, fold-into-
bridge as a wart-not-recommended), v1 scope (read-only state browser
+ search + pipeline kick-off), v2 scope (Cytoscape.js graph view),
and named the open design questions Michael needs to answer before
any code lands.

**Bridge image rebuild.** Image now bundles git-mcp + git + gh CLI +
chromium alongside the prior 8 MCP server binaries. 9 of 9 servers
respond after rebuild.

## What surprised

1. **Phantom GITHUB_TOKEN log.** Bridge log on first git-mcp spawn
   said `GITHUB_TOKEN=set`, but `printenv GITHUB_TOKEN` inside the
   container returned exit 1 (unset), and a re-spawn of git-mcp
   reported `GITHUB_TOKEN=unset`. Couldn't reproduce. Logged as
   carry-forward to watch.

2. **go.work auto-bumped to `go 1.26`** when I ran `go get
   chromedp@latest`. chromedp's deps want it. Bridge image already
   uses golang:1.26-alpine because gospel-engine-v2 needed it, so
   no rebuild surprise. But host-side dev would now need Go 1.26
   if not already on it.

3. **`scripts/git-mcp/` git status truncation.** `git status`
   summary view truncated the new directory's contents under
   "Untracked files:" — only saw the dir and a couple of files. Had
   to use `git status --short` or explicit `git add` to see the full
   list. Caught it before commit.

4. **The git_status loop bug** was real, not stale gopls. `for ...
   { ...; break }` terminates after first iteration regardless of
   the `if` predicate. Refactored to `strings.Cut`. Caught by gopls
   `SA4004` warning, not by my own review.

## Stewardship moments

- **Phantom GITHUB_TOKEN log.** Could have ignored. Instead spent
  ~5 min investigating because a token leak would be a real
  concern. Confirmed the current state matches Michael's intent
  (unset). Logged as carry-forward rather than dismissing.

- **Defaulted gh_issue_create to NOT granted to study.** Michael
  said study agent gets the git tools; he didn't specify gh_issue.
  Issue creation has higher blast-radius than PR creation (visible
  publicly on the issue tracker). Defaulted to deferred. Boundary
  test: would Michael, asked, want a study agent to autonomously
  open issues? Probably not without thinking about it. Right call
  to surface as a deferred grant rather than auto-grant.

- **Wrote 3f proposal even though the original spec was very
  different.** Michael's pivoted vision was clear in his answer.
  Original spec (cloud-hosted, multi-user, OAuth) is preserved at
  the top of the proposal as the long-run destination, but the
  body works through the new local-first vision. Honors his intent
  rather than the literal earlier doc.

## What this sets up

3d build is done; PAT setup turns on the live test. fetch-md v2
gives substrate agents the tooling to fetch JS-rendered sources.
3f proposal gives Michael a reaction surface for the next direction
on visualization.

The substrate's outbound surface is now genuinely complete: 9 MCP
servers, 50+ tools, allow-listed agents per family, async fan-out,
crash recovery, JS rendering, sandboxed git. The producer side is
real enough that the next reasonable directions are all
visualization (3f), authoring (3e.4 v2 write tools), or scaling
(3g multi-provider).

## Time

~2.5 hours autonomous: ~30 min plan doc + question batches; ~1 hour
git-mcp build + tests + SQL; ~30 min fetch-md v2 chromedp + smoke;
~30 min 3f proposal; ~30 min bridge rebuild + verify + commit.

The plan doc was the highest-leverage 30 minutes — it kept me out
of trouble for the next two hours.
