# Journal — coder-v2 complete: CV2.2 (push/PR) + CV2.3 (code-pr pipeline), e2e proven

**Date:** 2026-06-03 (continued session; container clock UTC 2026-06-04)
**Workstream:** WS5 / pg-ai-stewards — coder-v2
**Commits:** `20ab729` (CV2.2), `2d40938` (CV2.3). Root repo — committed, NOT pushed (Michael pushes root himself).

## What shipped

coder-v2 is now complete: the substrate can clone a real repo, work inside it, and land a reviewable PR — autonomously, through its own creation cycle.

- **CV2.2 — commit-local → substrate push → open PR.** `coder_commit` (local, no token), `coder_push` + `coder_open_pr` (bridge-side, token via a one-shot credential helper that never touches the sandbox or `.git/config`). cv2-2 migration passes `GITHUB_TOKEN` to coder-mcp's env + grants the 3 tools to `dev`. The repo allow-list (CV2.1) still constrains which repos.
- **CV2.3 — the `code-pr` pipeline.** clone → plan → implement → verify → pr (ladder raw→researched→planned→executing→verified). A work_item carrying `{repo, binding_question}` cascades all five stages: the agent clones the allow-listed repo into its worktree, plans against the *real* code, writes + iterates build/test to green (ground truth), independently verifies, then commits-local + pushes + opens a **draft** PR. The shared sandbox-stamp trigger was extended to code-pr so all five stages share one `wi-<8>` worktree across the revise loop.
- **Worktree-resilience fix.** A no-repo `coder_sandbox_start` now re-mounts an existing worktree (`HasWorktree`) instead of falling back to ephemeral `/work`. Foresight, not a smoke failure: without it, a container reap or bridge restart *mid-pipeline* would silently drop the clone between implement/verify/pr — exactly the failure a long multi-PR build (ai-chattermax) would hit.

## The Hinge, located precisely

The delegation-pattern audit keeps paying off. For code-pr the Hinge is the **human merge**, not the PR-open — because a draft PR is outward-facing but a *cheap walk-back* (close + delete, which I did three times this session) and it IS the review surface. So the loop + PR-open auto-advance; the merge (irreversible, trust-deciding) lives outside the substrate with the human. Contrast code-deploy (cc5), where `prepare` is the always-escalate Hinge because deploy is production-affecting. Same audit, different rung — the reversibility + review-surface test tells them apart.

## Three bugs the smokes caught (the value of inverse-hypothesis)

1. **git dubious-ownership** — coder-mcp runs git/gh as root over coder-uid-owned worktrees; git refused every op. Fix: `safe.directory=*` globally at Manager init (covers gh's internal git) + per-call `-c` on gitC.
2. **gh "you must first push"** — `gh pr create` without `--head` mis-detects the pushed branch in this bridge-side worktree. Fix: resolve the checked-out branch via rev-parse and pass `--head` explicitly. (The CV2.2 direct smoke caught both 1 + 2.)
3. **granted ≠ cataloged** — the 3 new CV2.2 tools were granted (cv2-2) but I never ran `stewards-mcp bridge refresh-tools`, so `compose_tools = catalog ∩ grants` excluded them. The *first* code-pr e2e run's pr stage degraded gracefully ("I don't have coder_commit/push/open_pr — I'll implement + verify and report so you can handle the PR step bridge-side") — a genuinely good agent failure mode. After refresh-tools, the second run opened the PR. **Lesson worth keeping: whenever coder-mcp (or any spawn target) gains tools, refresh-tools is the deploy step that makes them visible — same as the strongs/M-batch pattern.**

Each smoke verified the *ground truth* (the actual PR on GitHub, the worktree commit, the `ok chatroom 0.203s` build output), never the status flag. The "pr|completed|verified" flag was true about the *chat* completing while the PR did not exist — exactly why we don't trust the flag.

## E2E proof

Fresh code-pr work_item on ai-chattermax (add a `Version()` helper + test to the websocket-room module) → auto-cascaded all 5 stages in ~100s → DRAFT PR #3, head `agent/code-pr/wi-67ee541f` → base main, commit `982f52e`, body with a "What changed / Evidence (build+test) / Files touched" structure and the real passing output. The substrate wrote, tested, verified, and proposed a reviewable change with no human in the loop until the merge. Cleaned up after (PR closed, branch + sandboxes + worktrees removed, soak resumed).

## Carry-forward

- **Task #99 — decomposition + first ai-chattermax build, agent-MONITORED.** A single code-pr work_item per build item already works; decomposition (Batch-J fan-out → child code-pr work_items) is the scaling layer for the chat proposal's 9-item plan. Human merges PRs in order; PR-merge-gated child dependencies are a later gap. **Kick off WITH Michael** — it's his stress-test, he wants to watch it.
- **Model escalation knob.** `implement` defaults to kimi-k2.6 (proven) but is the per-task escalation point for novel app code (qwen3.7-max / deepseek-v4-pro). This is the one dial worth Michael's input for the real build.
- **v3 (future):** the substrate's own GitHub identity (a bot account) so it manages its own repos instead of borrowing Michael's PAT.

## Process notes

Migration discipline held throughout — cv2-2 + cv3 pre-applied via `docker exec bridge stewards-cli migrate` before each bridge restart, dodging the drift exit-2 landmine; confirmed migrate discovers files from the bind-mounted `/workspace` (so pure-SQL migrations need no image rebuild). Watchman soak paused for the build, resumed at the end. The whole arc stayed inside the medium-safe trust posture (our own code, our own repo) and the retrieval-only / no-opus-spend budget (the pipeline runs on opencode_go subscription models).
