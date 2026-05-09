---
workstream: WS5
status: design pass — awaiting Michael's direction on tier
created: 2026-05-09
related:
  - pg-ai-stewards-phase-2-5-generic-substrate.md (Phase 3 spec)
  - 3e-mcp-findings.md (bridge architecture)
references:
  - .github/skills/source-verification/SKILL.md (the "agent must never confabulate" discipline that motivates write-tool sandboxing)
---

# Phase 3d — Sandboxed git/gh for substrate-driven repo operations

> Design doc; **no implementation yet**. Michael flagged the line at
> phases.md:1272 ("sandboxed git") on 2026-05-09 as next direction.
> The producer side of the substrate (3a–3e) is real, agents are
> running real pipelines, and giving them git/gh capability is the
> next obvious surface. This doc covers context, threat model,
> options, recommendation, and an implementation sketch for the
> recommended option.

## Why this matters now

The substrate is at the threshold of being *generative*. Phase 3e
landed organic agent traffic through the bridge — kimi-k2.6 ran the
mysteries-of-god study and called `gospel_get` dozens of times during
review-stage quote verification. The next natural ask is: *can the
agent commit its work?* Can a study pipeline drop a polished
`/study/X.md` file as a PR? Can a watchman pass open an issue when it
finds drift? Can a research pipeline write a draft proposal to
`.spec/proposals/`?

Today every output crosses the substrate→workspace boundary by way of
Michael's hands. A `study-export` Go helper polishes substrate rows
into `/study/` files but Michael runs it. The agent never touches git.
That's been a deliberate, conservative shape — the source-verification
discipline plus the Apr 23 covenant clarification ("I own the intent
and vision, you own the code") need a real test before we let an
agent push commits.

This phase defines how to give the agent that capability **without
giving up the discipline**.

## Threat model

The threats are not generic untrusted-code-execution threats. We are
NOT running multi-tenant SaaS with adversarial user input. We are
running our own kimi-k2.6 and qwen-3.6 study agents on Michael's
workstation against repos Michael controls. The risks are:

1. **Hallucinated destructive commands.** Agent emits
   `git push --force` or `git reset --hard` or `gh pr close --delete-branch`
   from a misread context.
2. **Wrong-target operations.** Agent commits to `main` instead of a
   working branch, or pushes to a public repo it shouldn't have
   pushed to.
3. **Token leak through prompts.** Agent's tool dispatch context
   captures `GITHUB_TOKEN=...` from env and writes it to a message,
   to a study, or to a chat result that gets archived.
4. **Drift from intent.** Agent's "polish the draft" pipeline runs
   `git rebase -i` and rewrites Michael's commit history.
5. **Runaway commit volume.** Agent loops 50 times producing 50
   identical commits because of a prompting bug.

What is NOT in scope:

- Kernel-level isolation (we don't run untrusted user code; we run
  our own models on our own machine).
- Hardware-level VM boundaries (overkill for a single-developer
  workstation).
- Multi-tenancy (one user, one set of repos).

Right-sized isolation for this threat model:

- Allow-list of git operations (no force-push, no reset --hard, no
  branch -D, no rebase --onto).
- Scoped GitHub fine-grained PAT (one repo, contents:write only,
  short TTL).
- Token never reaches agent context — injected at sidecar level via
  env, not via tool args.
- Per-operation rate limit (max N commits per pipeline run).
- Branch namespace lock-in (agent can only push to
  `agent/<pipeline>-<work_item-id>` branches, never `main`).
- Audit log of every git/gh operation in `stewards.work_queue` rows
  (already free with the bridge architecture).

## Industry prior art (May 2026)

Researched 2026-05-09 via exa search. The dominant patterns:

- **The credential proxy pattern.** Anthropic, NVIDIA OpenShell,
  AgentPatterns.ai, GitHub Agentic Workflows all converge: the agent
  never holds the credential. A proxy outside the sandbox boundary
  attaches scoped tokens to validated, allowlisted requests.
  Anthropic's design explicitly states "sensitive credentials (such
  as git credentials or signing keys) never inside the sandbox with
  Claude Code." OWASP's AI Agent Security Cheat Sheet recommends the
  same: "issue short-lived tokens that are narrowly scoped to a
  specific task."
- **GitHub Agentic Workflows' three principles** (2026-03-09 GitHub
  blog): defense in depth, don't trust agents with secrets, stage
  and vet all writes, log everything. Their "safe outputs" pattern
  buffers all writes through a separate trusted MCP server that
  reviews each write before executing.
- **MicroVM isolation** (Firecracker, gVisor, Kata) is the standard
  for production AI sandboxes (E2B, Modal, Cloudflare). Overkill
  for our threat model — we're not running untrusted code.
- **Docker-in-Docker requires `--privileged`**, which "dramatically
  weakens the isolation." Anyone reaching for DinD as a sandbox is
  trading one vulnerability for another. Avoid.
- **Branch protection + PR review as a control layer** (Pyry Haulos,
  airut sandbox): "the configuration that governs the sandbox… is
  always read from the repository's default branch, not from the
  agent's working directory." Even with tight execution sandboxing,
  the human-merge gate matters.

The takeaway is consistent: for our threat tier (single-user, own
code, own models), **the credential proxy pattern + operation allow-
list + branch namespace lock-in covers the realistic risks**. Heavier
isolation (Firecracker, Kata) addresses threats we don't have.

## Architecture options

### Option A: Native Go MCP wrapper (`git-mcp`)

A new `scripts/git-mcp/` Go binary, structurally identical to
`fetch-md-mcp`. Each git operation is its own tool with a strict
input schema. Tool implementations shell out to `git` and `gh`
binaries, vetting args against per-tool allow-lists. The GitHub
token lives in the sidecar's env, never in tool args.

**Tools (v1):**
- `git_clone(repo_url, dir)` — clone into a workdir scoped to the
  pipeline
- `git_status(dir)`
- `git_branch_create(dir, name)` — refuses if name doesn't match
  `agent/*` pattern
- `git_add(dir, paths[])`
- `git_commit(dir, message)` — requires message, refuses
  `--amend`
- `git_push(dir, branch)` — refuses if branch is protected
  (`main`, `master`, `release/*`); refuses `--force` always
- `gh_pr_create(repo, head, base, title, body)` — `head` must
  match `agent/*`
- `gh_issue_create(repo, title, body)`

**Forbidden by construction:** the `git` and `gh` subcommands not
exposed as tools simply don't exist for the agent. There's no
`git_raw` tool that takes arbitrary args — every operation is a
typed method.

**Branch namespace lock:** any branch name passed to `git_branch_*`
or `git_push` must match `^agent/[a-z0-9-]+$`. A pre-create check
refuses any other namespace.

**Workdir isolation:** each pipeline run operates in
`/tmp/stewards-git/<work_item_id>/` (or its container equivalent).
No cross-pipeline visibility.

**Token handling:** `GITHUB_TOKEN` env var is set on the sidecar's
process. Never passed through tool args. The bridge daemon reads it
from its own env (just like `GOSPEL_ENGINE_TOKEN` today). Agent's
tool calls don't see it.

**Pros:**
- Same shape as fetch-md-mcp; team already knows the pattern.
- Fits the bridge architecture exactly — auto-promotes into
  tool_defs, granted per-agent like everything else.
- Single Go binary, no runtime deps.
- All audit logging already free via `work_queue` and
  `stewards.messages`.
- Tight allow-list = clear refusal semantics.

**Cons:**
- Trust model is "the wrapper code is correct." A bug in the
  allow-list could let through `--force-with-lease`. Defense:
  comprehensive test suite + landlock on the sidecar process.
- No isolation against the agent shell-injecting via a malformed
  branch name. Defense: regex anchors + Go's `exec.Command` (which
  doesn't shell-escape) + per-tool argv list (no string concat).

### Option B: Docker container sidecar

The git-mcp server runs in its own Docker container with the
workspace repo bind-mounted read-only at `/repo` and a writable
overlay at `/work`. Token injected via Docker `--env`. Allow-list
enforced inside the container. Network policy denies all egress
except `github.com:443` and `api.github.com:443`.

**Pros over A:**
- Stronger boundary (container fs/process namespace).
- Network egress can be allowlisted (Pyry Haulos pattern: only
  github.com:443).
- Process isolation prevents shell-injection from reaching host.

**Cons:**
- More moving parts. Container orchestration to maintain.
- Bind mounts create their own fragility (path translation, file
  permissions on Windows host).
- Still trusts the in-container allow-list; container boundary
  doesn't help against a buggy allow-list.

### Option C: Bubblewrap (Linux) / WSL2 sidecar

`bwrap` wrapping the git-mcp process. Filesystem namespaces, mount
read-only, restricted env. Available on Linux; Michael runs Windows
+ Docker so this would only apply inside containers (i.e., layered
defense if we already chose Option B).

**Pros:**
- Cheapest hardening on top of B.
- Anthropic's own sandbox uses bubblewrap (per the
  AgentPatterns.ai writeup citing Anthropic).

**Cons:**
- Doesn't fundamentally change the threat model from B.
- Adds skill burden.

### Option D: Stage and vet all writes (GitHub Agentic Workflows pattern)

Agent never pushes. Agent calls `safe_outputs.propose_commit(files,
message)`. The safe_outputs proxy buffers the write, runs analyses
(content sanitization, branch namespace check, file-pattern allow-
list) post-hoc, and only then executes the actual git push. Pull
requests, issue comments, and branch creates all go through the
same buffer.

**Pros:**
- Strongest discipline. No agent action can side-effect without
  passing the vetting layer.
- Audit trail is built-in.

**Cons:**
- Significant build effort. The vetting analyses are non-trivial
  (content sanitization, secret scanning, anomaly detection).
- Latency added between agent intent and effect — may break the
  iteration cadence the substrate is starting to see.
- Probably overkill for v1. Worth as a v2 layer once v1 is real.

## Recommendation

**Start with Option A (Go MCP wrapper) for v1.** Same time-to-value
as fetch-md-mcp (~1 hour). Threat model is right-sized for our
single-developer setup. Fits the bridge architecture cleanly. All
audit comes free.

Layer up to Option B (Docker sidecar) when one of these triggers:

- We start running substrate in a multi-user setting (ibeco.me cloud
  hub agents would qualify).
- We add tools that touch repos beyond `scripture-study` (the
  threat model changes when the agent could affect Michael's other
  projects).
- A real-world incident shows we needed it.

Layer up to Option D (safe_outputs) when:

- We want PRs to merge automatically based on agent verdicts (not
  just pause for human review).
- Auditing-by-construction becomes a hard requirement (e.g., for
  publishing teaching content where every commit is a public claim).

## v1 implementation sketch (Option A)

```
scripts/git-mcp/
├── go.mod
├── main.go        # MCP stdio server, env-token loader, tool registration
├── tools.go       # 9 tools, each with allow-list validation
├── allowlist.go   # Branch-pattern regex, file-pattern checks
└── workdir.go     # Per-work-item workdir lifecycle
```

**Substrate integration** (mirrors fetch-md-mcp):

```sql
-- 3d-1-git-mcp-seed.sql
INSERT INTO stewards.mcp_servers (name, ..., enabled) VALUES
  ('git', '...', 'stdio',
   'C:/.../scripts/git-mcp/git-mcp.exe', ..., true);

-- After bridge refresh-tools picks up the 9 tools, grant them
-- selectively. Initial set:
INSERT INTO stewards.agent_tool_perms VALUES
  ('study',  'git_clone',         'allow', 'manual'),
  ('study',  'git_status',        'allow', 'manual'),
  ('study',  'git_branch_create', 'allow', 'manual'),
  ('study',  'git_add',           'allow', 'manual'),
  ('study',  'git_commit',        'allow', 'manual'),
  ('study',  'git_push',          'allow', 'manual'),
  ('study',  'gh_pr_create',      'allow', 'manual')
  -- gh_issue_create deliberately not granted to study
ON CONFLICT (...) DO NOTHING;
```

**Verification path:**
1. Synthetic enqueue: agent issues `git_clone` for the workspace
   repo into `/tmp/stewards-git/test/`. Verify clone succeeds.
2. `git_branch_create("test", "agent/study-test-001")`. Verify
   branch exists.
3. `git_branch_create("test", "main")` → must fail with "branch
   name does not match agent/* pattern".
4. `git_push("test", "main")` → must fail with "branch is
   protected".
5. `git_push("test", "agent/study-test-001", "--force")` → must
   fail (--force not in tool surface; argv-based defense).
6. Run a real study-write pipeline that ends with `git_commit` +
   `git_push` to a temp branch. Verify the commit lands on the
   correct branch with the expected author.

**Token setup:**
- Generate fine-grained PAT scoped to one repo (`scripture-study`),
  `contents: write`, `pull-requests: write`, 30-day TTL.
- Set `GITHUB_TOKEN` in the bridge daemon's env (where it joins
  `GOSPEL_ENGINE_TOKEN`, `BECOMING_TOKEN`).
- git-mcp reads it from env at process startup.

**Open design questions for Michael:**
- Should `gh_pr_create` auto-add `Draft: true` so PRs always start
  as drafts? Inclines toward yes — gives Michael a review pause.
- Should commits include a `Co-Authored-By: <agent-family>-via-pg-ai-stewards`
  trailer so the audit trail is visible in `git log`?
- What's the workdir lifecycle policy — keep on disk for inspection
  vs. cleanup after success? Prefer keep-on-disk for v1 (debugging
  beats hygiene at this scale).

## Phase boundaries

- **3d.1 — git-mcp v1.** Option A above. ~1 hour build, ~30 min
  verification. Ships a usable agent surface immediately.
- **3d.2 — Docker sidecar wrapper.** Option B layered atop. Triggers
  per the recommendation list above.
- **3d.3 — safe_outputs proxy.** Option D. Prereq: 3d.1 stable
  enough that the buffer-and-vet flow has something to wrap.

3d as currently scoped does NOT need to ship before 3f (web UI) or
3e.5 (gospel_passthrough). It's parallel work driven by a different
need (capability) than 3f (visibility).

## Done criteria

- An agent in a study-write pipeline can clone the repo, create a
  branch, commit a polished study file, push the branch, open a
  draft PR — all through git-mcp tools, all audit-logged in
  `work_queue`, no destructive ops possible by construction.
- Michael reviews and merges the PR by hand. The agent never
  touches `main`.
- Inverse hypothesis verified: each forbidden op (force-push, reset
  --hard, push to main) returns a clean tool-error, not a crash.

## Risks and mitigations

- **Branch protection bypass via tag operations.** Mitigation: do
  not expose `git_tag` or `git_push --tags` in v1.
- **Workdir leak across pipelines.** Mitigation: per-work_item_id
  subdirectory; cleanup script invoked at pipeline completion.
- **Token theft via filesystem read.** Mitigation: token only in
  sidecar process env, never written to disk in workdir. The agent's
  filesystem-read tools cannot reach the sidecar's env.
- **Allow-list regex bypass via unicode/lookalike chars.** Mitigation:
  `^[a-z0-9-]+$` strictly ASCII, anchored both ends.
- **Repository confusion (agent commits to wrong workspace dir).**
  Mitigation: workdir derives from the work_item_id, not from
  agent input. Tools that expect a workdir path must match
  `^/tmp/stewards-git/[0-9a-f-]+/$` and refuse anything else.

## Why not just build it tonight

- The token setup needs Michael's GitHub account — fine-grained
  PATs, repo selection, scope choice.
- The decision between A → B → D layering is a real design choice
  that affects the substrate's posture for months. Worth Michael's
  judgment.
- The rate-limit policy, branch-namespace pattern, and workdir
  cleanup policy are all reasonable tonight, but Michael may have
  preferences (e.g., wants `agent/<pipeline>/<work-item>` instead
  of flat `agent/<work-item>`).

The proposal is ready. Implementation triggers on Michael's
direction.
