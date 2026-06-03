# Substrate coding capability — CC.4–CC.6 + v1 COMPLETE

**Date:** 2026-06-03
**Workstream:** WS5 / pg-ai-stewards
**Mode:** dev (full stewardship — "see this through the entire plan")
**Outcome:** v1 of substrate-coding-capability is complete. The substrate can write, build, test, diagnose, and deploy code — autonomously where there's ground truth, with a human at the Hinge where there isn't.

## What shipped (this stretch)

- **CC.4 (`4ff86e4`)** — `coder_lsp`: type/compile diagnostics by extension (gopls / tsc / pyright, already in the image). `clean` bool + the diagnostics text. Smoke: a string-as-int error surfaced the real gopls message; a clean file → `clean=true`.
- **CC.5 (`199803d`)** — deploy + **the always-escalate Hinge**:
  - `coder_deploy` runs the built artifact as a background service in its sandbox (the sandbox IS its docker sidecar) + healthchecks `http://localhost:<port><path>`.
  - `code-deploy` pipeline (`prepare` → `deploy`) where **`prepare` has `auto_advance=false`** — the work_item stops at `awaiting_review` and a human must ratify before `deploy` runs. Deploy is outward-facing and not a cheap walk-back, so it always escalates regardless of trust. **This is the always-escalate rung the delegation-pattern audit proposed** (docs/delegation-pattern-skills-and-gates.md), now built — Exodus 18:22 made literal.
  - Smoke: wrote a Go HTTP server, built it, `coder_deploy` → `healthy=true`, "ok from the deployed sidecar".
- **CC.6 (`3297b5a`)** — hardening: `coder_sandbox_list` (visibility) + `coder_sandbox_reap` (remove sandboxes older than N min; flush-all on negative) + a best-effort startup reap (>2h) in coder-mcp. Resource caps already shipped in CC.1; the deploy-secret broker is deferred to v2/Dokploy. Smoke: the reaper found + removed a real leaked sandbox from an earlier test.

## The whole v1 arc (CC.1–CC.6)

| Phase | What | Commit |
|---|---|---|
| CC.1 | coder-runtime image (Go/Node/Python/LSP) + sandbox-manager + bridge docker access | `ee1760a` |
| CC.2 | coder MCP server, 13-tool surface (official go-sdk) | `b5f055c` |
| CC.3 | `code-write` pipeline — agent-driven ground-truth loop (proven e2e) | `e227501` |
| CC.4 | `coder_lsp` diagnostics | `4ff86e4` |
| CC.5 | deploy-to-sidecar + the always-escalate Hinge | `199803d` |
| CC.6 | sandbox reaper + visibility | `3297b5a` |

Plus migrations cc2–cc6 (registry/grants/pipelines), all pre-applied via the ledger before each bridge restart (the drift exit-2 rule held every time).

## The proof that matters

CC.3 end-to-end: a `code-write` work_item ("Add fn + table test") cascaded plan→implement→verify to `verified`; the dev agent autonomously wrote `calc.go`+`calc_test.go`+`go.mod` and iterated build/test to green via the coder tools — and **I confirmed by hand** (`go build && go test` → `ok calc`), inverse-hypothesis. CC.5: built + deployed + healthchecked a real web server in a sidecar. The Prescription/ground-truth loop and the Hinge from the audit are both real in code now.

## Carry-forward (v2 / refinements)

- **CC.3.1 (path B):** the deterministic substrate gate — build/test via `tool_dispatch`, branch maturity on the real exit code, no LLM in the gate (Michael: "I can see a world where both paths are usable").
- **CC.7 (v2):** Dokploy deploy with the substrate's own scoped/walled-off access (separate namespace + scoped token), short-lived broker-injected creds, its own ratification pass.
- **Model tuning:** kimi-k2.6 handled small tasks; harder code may want a stronger opencode_go model (per-stage via `stage_models`).
- **Sandbox lifecycle:** v1 leaves the sandbox up after `verified` (so deploy can use the artifact); the reaper + startup-reap clean leaks. A maturity→terminal teardown hook is a refinement.
- **Persisting the deliverable:** v1's output is code-in-the-sandbox; a git branch/PR via git-mcp (the PR rung + trust gating, D-CC7) is the next natural step.
- **Remote coding / SSH-tunnel agent work** (Michael's future seed) would re-open the isolation tier (gVisor/microVM) — out of scope for medium-safe v1.

## Relational note

Michael handed full stewardship — "see this through the entire plan… commit and journal through good phased gates." Carried it through six gated commits + four journals, surfacing only the two genuine forks (the build/test gate mechanism A-vs-B; the v1 decisions batch) and deciding the reversible rest. The session was very long; held to Ammon (context-fullness → lean on durable files, keep going, don't beg off) rather than stopping at heaviness. The plan is done.
