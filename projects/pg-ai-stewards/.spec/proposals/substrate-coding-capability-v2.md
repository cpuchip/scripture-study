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

kimi-k2.6 nailed known patterns (the WebSocket hub) but app code with no canonical template (the classifier gate, the persona handshake) is the real test. Per-stage tunable via `stage_models`.

**Shipped default: kimi-k2.6 across all `code-pr` stages, with `implement` as the per-task escalation point** (2026-06-03, after web research + Michael's hands-on read). The reframe that drove the choice: `implement` is not single-shot codegen — it's an *autonomous terminal loop* (clone → write → build/test → read output → fix → iterate, many tool calls). So the decisive axis is **Terminal-Bench / agentic stability**, not single-shot SWE-Bench Verified. On that axis kimi-k2.6 is the documented leader among the affordable open models (Terminal-Bench 2.0 66.7%, SWE-Bench Pro 58.6%, a 4,000+ tool-call / 13-hour autonomous session; every source names it "the right answer for autonomous long-running agents" / "single-repo code-write-debug loops"). It is also non-reasoning (no budget-burn risk in a long loop) and the cheapest. Michael's standing read: k2.6 ≈ frontier-close, rated over deepseek-v4-pro — the benchmarks agree.

**Escalation map (per-task override on `implement`):**
- **Front-end / Vue / UI items** (ai-chattermax roster + moderation UI) → **glm-5.1**. Its real edge over k2.6 is concentrated in front-end/UI generation (Code Arena top-tier on agentic web dev) and whole-repo scaffolding (NL2Repo, beats Opus/GPT); it shows *no measurable edge on non-UI/algorithmic* work. (It is a reasoning model → slower + token-burn; give it adequate per-call max_tokens.)
- **"Scaffold a new module from a spec" items** → **glm-5.1** (NL2Repo strength).
- **Hard novel logic where k2.6 stalls** → **qwen3.7-max** — the agentic-benchmark leader among Chinese models (Terminal-Bench 2.0 69.7%, SWE-Bench Pro 60.6%, a 35-hour / 1,158-tool-call demo) but reasoning + priciest + brand-new (no hands-on feel yet); watch the reasoning budget.
- **deepseek-v4-pro: dropped as an escalation** — SWE-Bench Pro 55.4% sits *below* the k2.6 default; it is not an upgrade for this loop.
- **minimax-m3** (released 2026-06-01): **REGISTERED + verified end-to-end 2026-06-03 (cv4).** opencode_go serves it (openai-format, reasoning model — emits `<think>`, which the substrate separates cleanly; no leak into the deliverable); 1M context; strong on MCP Atlas (74.2% — tool-connected MCP execution, closest to our MCP-driven substrate). A full code-pr run on m3 (all stages, `model_override`, max_tokens 64k) produced a genuinely high-quality `/healthz` PR — a proper `ClientCount()` accessor, 3 convention-matched stdlib tests, a thorough PR body — but ran **~3–4× slower than kimi-k2.6** (reasoning overhead: ~5.5 min vs ~100s). **Use it as the big-context / deep-reasoning escalation** (large-repo items, novel logic worth the wait), not the loop default. Caveat surfaced by the run → the artifact-hygiene gap below.

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
- **Build-artifact hygiene (found 2026-06-03 in the m3 test; fix before the real ai-chattermax build — task #99).** `coder_commit` does `git add -A`, so any build artifact left in the worktree lands in the PR. The m3 run ran `go build` (which writes the `chatroom` binary into the main-package dir) → a **9MB binary committed alongside the source**. Not model-specific (kimi dodged it only because its tasks used `go test`). Fix options, in order of generality: (a) **plan/implement build command discipline** — prefer `go build ./...` (compiles all packages, writes nothing) over `go build .`/`go build` (writes the binary); (b) **`coder_commit` artifact hygiene** — skip obvious build outputs / respect a default ignore set; (c) **per-repo `.gitignore`** — works but extensionless binaries (named after the module, e.g. `chatroom`) aren't caught by a generic Go `.gitignore`, so it must name them. Likely (a)+(b) together; (c) per repo as backup.

## v3 (future) — the substrate's own GitHub identity

Michael 2026-06-04: give pg-ai-stewards **its own GitHub account** (a bot/machine-user, or a GitHub App identity) so it manages **its own repos** rather than borrowing Michael's PAT. This is the clean endgame — commits + PRs show as the substrate, not Michael; the substrate owns the repos it works on; and the "his-account blast radius" concern disappears entirely (the token can only touch the bot's own repos). v2 (a fine-grained PAT scoped to allow-listed repos in Michael's account) is the stepping stone; v3 is the separation. Likely a dedicated machine-user account + a fine-grained PAT (or a GitHub App installed only on the bot's repos), with Michael added as a collaborator where he wants visibility/merge rights.
