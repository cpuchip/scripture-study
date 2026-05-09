# Carry-forward autonomous plan — 2026-05-09 evening

Michael ratified four pieces and asked me to take stewardship over them
while he rests after a family move. This is my tracking doc so I don't
get lost mid-session.

## Decisions captured

### 1. gospel-engine-v2 (handled by Michael)

Status: **already done**. Michael removed it as a git submodule and
committed it to the bare repo (commits `174ec72` + `c4f7f54`). Source
still on disk at `scripts/gospel-engine-v2/`; gitignored at parent.
Bridge build unaffected (Docker honors `.dockerignore`, not
`.gitignore`). Will verify on next bridge rebuild.

No work for me here.

### 2. 3d sandboxed git — Build Option A v1 (PAT setup deferred)

- **Tier**: Option A — Go MCP wrapper. Per-tool allow-list, branch
  namespace lock, env-only token.
- **Branch namespace**: `agent/<pipeline>/<work-item-id>-<short-slug>`.
  Allow-list regex: `^agent/[a-z0-9-]+/[a-z0-9-]+(-[a-z0-9-]+)?$`.
  Components from `work_items.pipeline_family`, `work_items.id` (UUID,
  short form ok), `work_items.slug` (truncated to ~40 chars).
- **PR drafts**: NO — `gh_pr_create` opens ready-for-review PRs by
  default. Caller can explicitly request draft if needed.
- **Co-Authored-By trailer**: YES —
  `Co-Authored-By: <agent-family>-via-pg-ai-stewards <agents@cpuchip.net>`.
  Auto-appended to every `git_commit` message.
- **Workdir lifecycle**: keep on disk for inspection.
  `/tmp/stewards-git/<work-item-id>/` persists. No auto-cleanup.
- **PAT setup**: deferred to tomorrow. I build everything except the
  live test that pushes to GitHub. The bridge will pick up
  `GITHUB_TOKEN` from `.env` when Michael sets it.

Tools (v1):
- `git_clone(repo_url, work_item_id)` — clones to
  `/tmp/stewards-git/<work-item-id>/`. Refuses to clone into a
  pre-existing directory unless `--reuse` (not in v1).
- `git_status(work_item_id)`
- `git_branch_create(work_item_id, pipeline, slug)` — constructs
  `agent/<pipeline>/<work-item-id>-<short-slug>`. Refuses if name
  doesn't match allow-list regex. Refuses protected branches
  (main, master, release/*).
- `git_add(work_item_id, paths[])`
- `git_commit(work_item_id, message, agent_family)` — appends the
  Co-Authored-By trailer using agent_family. Refuses `--amend` (not
  exposed). Refuses empty message.
- `git_push(work_item_id, branch)` — refuses if branch not in the
  agent/* namespace. Refuses `--force` (not exposed).
- `gh_pr_create(repo, head, base, title, body)` — `head` must
  match `agent/*`. Defaults `--draft=false`. Caller can set draft
  explicitly via param if needed (v1 just exposes title/body/head/base).
- `gh_issue_create(repo, title, body)`

Forbidden by construction (not exposed as tools): `git_raw`,
`git_reset`, `git_rebase`, `git_branch_delete`, `git_tag`, force ops.

Verification path (without PAT):
- Build + JSON-RPC initialize + tools/list smoke
- Allow-list regex unit-test cases (good names accepted, bad names
  refused)
- Workdir creation and isolation
- Branch namespace generator output for sample pipeline+work-item

Verification path (after PAT):
- Live `git_clone` of a fork
- Live commit + push to a temp branch
- Live `gh_pr_create` against the fork

### 3. 3f local web UI — proposal doc (Michael's pivoted vision)

Michael's restated vision (verbatim paraphrase):
> "My idea for a.ibeco.me probably needs to be more flushed out since
> our current setup doesn't allow for more than 1 user. For now I want
> an elegant UI that allows me to see the state, explore the data,
> kick off AI work and interact with it. And see results. To search
> it, and see the graph connections between studies/gospel-library/etc.
> This will be hosted locally, probably in the bridge or a new service
> that runs beside the other two."

This is a **significant pivot from the original 3f spec**. It's no
longer "a.ibeco.me cloud-hosted, multi-user, OAuth via becoming." It's
local-first, single-user, co-located with the docker stack.

Proposal scope:
- Capture the vision (single-user, local, alongside pg + bridge)
- Surface the design questions Michael will need to answer before
  building:
  - **Stack**: Go-served single binary (matches pg + bridge pattern,
    one process to deploy) vs Vue/Vite + thin Go API split (richer
    UI, more code)
  - **Graph rendering**: Cytoscape.js for AGE/Cypher results? D3?
    What graph queries do we want to surface?
  - **Search surface**: substrate's existing FTS + semantic via
    sql_fn tools, exposed as a UI search bar
  - **Pipeline kick-off UI**: form for new work_items? Or richer
    "draft a binding question, pick a pipeline, set a token budget"?
  - **In-flight observability**: live work_queue tail, bridge call
    log, model-call inspector
  - **Hosting**: bake into bridge container vs new `ui` service
- Recommended slim v1 scope:
  - Read-only state browser first (work_items, studies, sessions)
  - Search bar (FTS + semantic)
  - Then add pipeline kick-off
  - Graph view last (most code)

Output: `.spec/proposals/pg-ai-stewards-3f-local-ui.md`. Replaces the
original 3f spec entry. No code yet.

### 4. fetch-md v2 — chromedp (in-process Chromium)

Approach: add chromedp as a dep to `scripts/fetch-md-mcp/`, refactor
existing tools to optionally use JS rendering, add `js: true` param
to `fetch_url` and `fetch_urls` (defaults false to keep v1 fast path).

Bridge image change: alpine:3.20 runtime needs Chromium. Add
`chromium` package via apk. ~150MB extra image size; acceptable.

chromedp launches Chromium as a subprocess (not in-process truly;
it's a CDP-over-WebSocket client). On alpine, set `CHROME_BIN` or
similar so chromedp finds the system chromium.

Tools to update:
- `fetch_url(url, max_chars?, js?)` — js=false uses current path
  (HTTP + readability + html-to-markdown); js=true launches headless
  Chromium, waits for `domcontentloaded` (configurable wait), then
  extracts via readability + html-to-markdown.
- `fetch_urls(urls[], max_chars?, js?)` — same param.
- `extract_links(url, js?)` — same.
- `fetch_url_raw(url, js?, wait_until?)` — js=true returns rendered
  HTML; wait_until controls page-ready signal.

Verification:
- Smoke against a known-static page (Wikipedia) with js=false → same
  behavior as v1
- Smoke against a known-SPA (some React docs site) with js=true →
  content present that wouldn't render with plain HTTP
- Verify image size impact

## Sequencing

Dependencies:
- 3d and fetch-md v2 both touch the bridge image (new binaries / new
  deps). Combine into one rebuild cycle.
- 3f is a doc only — no rebuild.
- All tasks commit independently.

Order:
1. **3d sandboxed git** — code first (largest piece)
2. **fetch-md v2** — code second
3. **Bridge image rebuild + smoke** — combined for both above
4. **3f proposal** — doc, last (lower stakes)

## What I'll commit per piece

- `feat(git-mcp): v1 sandboxed git/gh wrapper for substrate agents`
  - `scripts/git-mcp/` (new)
  - `projects/pg-ai-stewards/extension/3e2-7-git-mcp-seed.sql` (new)
  - `projects/pg-ai-stewards/extension/src/lib.rs` (fold-in)
  - `projects/pg-ai-stewards/extension/Dockerfile` (COPY entry)
  - `projects/pg-ai-stewards/extension/bridge.Dockerfile` (build +
    install git-mcp into /usr/local/bin)
  - `go.work` (add scripts/git-mcp)

- `feat(fetch-md-mcp): v2 — JS rendering via chromedp`
  - `scripts/fetch-md-mcp/main.go`, `tools.go` (chromedp paths)
  - `scripts/fetch-md-mcp/go.mod` (chromedp dep)
  - `projects/pg-ai-stewards/extension/bridge.Dockerfile` (apk add
    chromium, env vars)

- `chore(bridge): rebuild image, verify post-restart`
  - May or may not need a separate commit; if both above land
    cleanly the rebuild is implicit.

- `docs(pg-ai-stewards): 3f local UI proposal`
  - `.spec/proposals/pg-ai-stewards-3f-local-ui.md` (new)
  - Update `phases.md` row to point at the new proposal

## What I will NOT do without confirming

- Push commits to remote (per standing pattern)
- Restart the live `pg` container (soak data preserved but mid-pass
  interruption cancels work)
- Touch `.env` to add `GITHUB_TOKEN` (Michael generates the PAT)
- Write to anything in the gospel-engine-v2 directory (Michael owns
  that surface now)
- Spawn any chat work that would consume model tokens
- Make architectural decisions on 3f beyond writing a proposal

## When I'll surface back

If any of these:
- A build error I can't resolve in 2 attempts
- A design decision the carry-forward spec didn't cover and that
  significantly affects scope
- A discovery that contradicts something Michael ratified above
- Unexpected state in the substrate (soak failure, work_queue
  pileup, etc.)
- All four pieces shipped (final summary)

Otherwise I work through the queue and report at the end.
