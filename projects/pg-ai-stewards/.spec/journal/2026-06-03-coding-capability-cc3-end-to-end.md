# Substrate coding capability — CC.3: the substrate wrote tested code, autonomously

**Date:** 2026-06-03
**Workstream:** WS5 / pg-ai-stewards
**Mode:** dev (full stewardship)
**Outcome:** the core capability works end to end. The substrate, given a coding task, wrote a Go package + test, iterated build/test to green in a sandbox, and reached `verified` — independently confirmed.

## What shipped (commit `e227501`)

CC.3 path A — the `code-write` pipeline (agent-driven ground-truth loop):
- `plan` (dev, tools off) → `planned`: implementation plan + the build/test command.
- `implement` (dev, coder tools on) → `executing`: the agent starts its sandbox, writes code, runs `coder_shell` build+test, iterates to green.
- `verify` (dev, coder tools on) → `verified`: independently re-runs build/test; REVIEW: passes/fail drives the gate.
- Ladder `raw→planned→executing→verified`. A BEFORE-INSERT trigger stamps a stable `input.sandbox = wi-<8>` so implement+verify share one sandbox across the revise loop. `coder_sandbox_start` made idempotent-preserving so a revise doesn't wipe in-progress work.

Ratification this session: Michael chose "A then B" for the gate — ship the agent-driven loop now, add the deterministic substrate gate (tool_dispatch, no LLM) as a co-usable fast-follow (CC.3.1). "I can actually see a world where both paths are usable."

## The end-to-end run (the proof)

Created a `code-write` work_item: *"a function Add(a, b int) int in package calc + a table-driven test; build+test `go mod init calc && go build ./... && go test ./...`."* Dispatched the plan stage; the pipeline auto-cascaded.

Watched it live: `stages_run` went 1→3, the sandbox container `coder-sb-wi-e60498da` came UP, and the work_queue showed the real tool-use loop — `chat → tool_dispatch → mcp_proxy → done`, repeating (the dev agent calling coder tools through the substrate's AT-batch tool dispatch). It reached **`status=completed, maturity=verified`**.

**Inverse-hypothesis verification (didn't trust the flag):** exec'd into the sandbox and ran `go build ./... && go test ./...` myself → **`ok calc`**. The agent had written real, idiomatic Go: a clean `Add` plus a 5-case table-driven `TestAdd` with subtests. Autonomous, compiling, tested code — and I confirmed the ground truth by hand, not by trusting "verified."

## Why this matters

This is the Prescription/ground-truth loop from the delegation-pattern audit, working in code: the value (does it compile? do tests pass?) is checkable without anyone's discernment, so the agent iterates autonomously and safely. The substrate's governance (the gate, the sandbox isolation) is what makes autonomous coding safe — the thing opencode lacks. CC.1+CC.2 were the foundation; CC.3 is the capability.

## Lessons

- The existing tool-use machinery (AT batch) carried the coder tools with zero changes — the dev agent iterated write→build→test→fix through `tool_dispatch`/`mcp_proxy` natively.
- `kimi-k2.6` (opencode_go) handled the multi-tool coding loop fine on a small task. Model choice is tunable via `stage_models` for harder tasks (carry-forward).
- The drift exit-2 pre-apply rule held again (cc2, cc3 both pre-applied before the bridge restart).

## Carry-forward

- **CC.3.1 (path B):** the deterministic substrate gate — enqueue `coder_shell` build/test via `tool_dispatch`, branch maturity on the real exit code, no LLM in the gate. Co-usable with A.
- **Sandbox lifecycle:** v1 leaves the sandbox up after `verified` (so CC.5 deploy can use the artifact); a reaper for abandoned sandboxes is a CC.6 item.
- **Model tuning:** harder coding tasks may want a stronger opencode_go model than kimi-k2.6.
- **The deliverable:** v1's output is code-in-the-sandbox; persisting it (git branch/PR via git-mcp, or the deploy artifact) comes with CC.5 + the PR rung.
- Next: CC.4 (coder_lsp diagnostics — the LSP servers are already in the image), CC.5 (deploy-to-sidecar + the always-escalate Hinge rung), CC.6 (hardening).
