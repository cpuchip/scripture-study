# Proposal — Substrate Coding Capability v2 (work in a real repo, land PRs)

**Status:** ✅ RATIFIED 2026-06-03 (core decisions); build-ready after the open implementation questions are settled at their work_items. Successor to `substrate-coding-capability.md` (v1: write/build/test/deploy in a fresh sandbox).
**Raised:** 2026-06-03. Michael, after the v1 coder built FizzBuzz + a working WebSocket chat-room from scratch: *"ratify our V2 … figure out how to get pg-ai-stewards secure creds for github so it can git clone/push/MR work on specific repos in my account."* First real target: **ai-chattermax**, built incrementally as a large multi-PR stress test (see `projects/ai-chattermax/.spec/proposals/chat-server-design.md`).

## The gap v2 closes

v1 builds a self-contained program from scratch in a fresh sandbox. v2 lets the coder **work inside an existing repo and land its work as reviewable PRs** — which is what building a real project (ai-chattermax, or any repo in Michael's account) requires. The deliverable shifts from "code in an ephemeral sandbox" to "a branch + a PR on the real repo, that a human reviews and merges."

## Ratified decisions

- **D-CV2.1 — GitHub creds: a fine-grained PAT scoped to whitelisted repos.** Contents: read/write + Pull requests: write, on *only* the repos the coder may touch (ai-chattermax, etc.) — not the account, not admin. Blast radius = those repos. *(The plumbing already exists: git-mcp reads `GITHUB_TOKEN` from the bridge env via the `$$env:GITHUB_TOKEN` pattern, and git-mcp currently sees it `set`. v2 ensures that token is fine-grained/repo-scoped — swap it if the current one is broad — and adds the allow-list below.)*
- **D-CV2.2 — Repo allow-list at the git-mcp tool layer.** git-mcp refuses clone/push/PR on any repo not in an explicit allow-list (mirrors its existing branch-namespace + protected-branch guards). So even *with* the token, the substrate can only operate on whitelisted repos.
- **D-CV2.3 — Token never enters the sandbox or model context.** The agent commits *locally* (no creds needed); the substrate (git-mcp, token in the bridge env, via git's credential helper — never in the URL/`.git/config`/arglist/stdout) does the clone, push, and PR.
- **D-CV2.4 — Flow: commit-local → substrate pushes/PRs → human merges.** The merge is the Hinge (the delegation-pattern audit again — the human ratifies bringing agent code into the real repo; the sandbox protects the build host, the merge gate protects the decision to trust the code).

## Architecture — the shared worktree

v1's sandbox used its own ephemeral fs (no host mount). v2 adds a **per-work_item worktree volume** shared between the bridge and the sandbox:

```
git-mcp (bridge, token in env)  ──clone (token via credential-helper, NOT in .git)──→  worktree volume (per work_item)
                                                                                          │  (no token in .git/config)
sandbox container  ──mounts──→  the worktree volume  ──→  agent edits / builds / tests / commits LOCALLY (no token)
                                                                                          │
git-mcp (bridge, token in env)  ──push branch + open PR──→  GitHub                        ┘
```

- **Clone:** git-mcp clones the (allow-listed) repo into the per-work_item worktree volume, token via git's credential helper / env — never written into `.git/config` (git-mcp's existing discipline), so the sandbox can't read it.
- **Work:** the sandbox mounts that worktree; the agent uses the coder tools (write/edit/shell/lsp) on it, runs build/test (the v1 ground-truth loop), and commits to an `agent/<pipeline>/<wi>-<slug>` branch — **all local, no token.**
- **Land:** git-mcp (token in the bridge) pushes the branch + opens the PR. Token never touched the sandbox.
- **Merge:** the human reviews + merges the PR (the Hinge).

## Pipeline shape

A `code-pr` pipeline (or extend `code-write`): `clone → plan → implement → verify → pr`.
- `clone` (substrate/git-mcp, deterministic): clone the allow-listed repo into the worktree.
- `plan` / `implement` / `verify`: v1's loop (build/test ground truth), now operating on the cloned repo rather than empty `/work`.
- `pr` (substrate/git-mcp): push the agent branch + open the PR. Trust-gated at this rung per D-CC7 (the PR rung is where trust decides auto-open vs. surface-for-review).

## Decomposition (for app-sized builds like ai-chattermax)

A whole app is many work_items, not one. v2 wires the existing fan-out / work-item-hierarchy machinery (Batch J) to `code-pr`: a parent "build ai-chattermax" decomposes into dependency-ordered child work_items (each ≈ one PR — the chat proposal's 9-item build plan), each cloning the repo, building its piece, and opening a PR. The orchestrating agent + Michael review/merge in order.

## Model

kimi-k2.6 nailed known patterns (the WebSocket hub) but app code with no canonical template (the classifier gate, the persona handshake) is the real test. v2 should default `code-pr` to a **stronger opencode_go coding model** (per-stage via `stage_models`), tunable per task.

## Monitoring discipline (Michael's explicit ask)

ai-chattermax is the stress test. The orchestrating agent **watches each child work_item and surfaces issues as they surface** — build failures, model fumbles on novel code, scope drift, security seams, decomposition mistakes — and we fix them. That feedback is the point of doing a large real project first.

## Open implementation questions (settle at the work_items, not blocking ratification)

- Worktree volume mechanism: a named docker volume per work_item mounted into both bridge and sandbox, vs. a host bind dir. (Mind the v1 §4 rule: never the *live* `/workspace` — this is a *separate* clone.)
- Where local commits happen: agent runs `git commit` in the sandbox, or git-mcp commits the worktree from the bridge. (Local commit needs no token either way; pick the cleaner.)
- How the allow-list is configured (a `stewards.coder_repos` table? a git-mcp flag?).
- Whether `code-pr` is a new pipeline or `code-write` + a `pr` rung.
- Fine-grained PAT vs. GitHub App (short-lived tokens) — PAT is v2; App is the hardening, like gVisor for the sandbox.

## Caveats

- The merge boundary is the real trust gate — sandbox isolation protects the build host, not the decision to trust the code. Review PRs (the Hinge).
- A fine-grained PAT in the bridge env is still a standing credential; the GitHub App (auto-rotated short-lived tokens) is the hardening path if/when the threat model grows (remote/multi-tenant).
