# Proposal — Substrate Coding Capability (write · build · test · deploy)

**Status:** DRAFT — three core decisions ratified 2026-06-03 (AskUserQuestion); remaining decisions (D-CC4…) open for a decisions walk before the build.
**Raised:** 2026-06-03. Michael: *"how can we get pg-ai-stewards to be able to program? … make it so that pg-ai-stewards can write code, and even deploy it in its own docker sidecars."*
**Research basis:** opencode (`external_context/opencode`) + the 2026 AI-agent-sandbox literature (Docker Sandboxes, E2B, Daytona, gVisor, Firecracker; Northflank / amux / Zylos guides). Governance frame: [`docs/delegation-pattern-skills-and-gates.md`](../../../docs/delegation-pattern-skills-and-gates.md).

---

## 1. The goal in one sentence

Give the substrate's `dev`/`coder` agent the ability to **write code, build and test it against ground truth, and (with a human at the gate) deploy it in its own ephemeral Docker sidecar** — all inside the existing maturity-gate / trust-ladder governance, so autonomous coding is *safe by structure*, not by a human approving each command.

## 2. Why this fits the substrate better than it fits opencode

Opencode's coding model is **tools + permission-gating + LSP, executed on the host**: a `shell` tool whose commands are parsed with tree-sitter and gated by a permission/arity layer, plus `edit`/`write`/`apply_patch`/`read`/`glob`/`grep`/`lsp`/`task`. Its `packages/containers` is CI build images, **not** a runtime sandbox. That model assumes a human sitting there approving dangerous commands. It does not isolate execution.

The substrate is already containerized and governance-first, and it has the thing opencode lacks: a **ground-truth gate**. That makes code the *ideal* substrate workload, mapped onto the delegation pattern from the audit:

- **Build + test = Prescription (ground truth).** "Does it compile? Do the tests pass?" is checkable without anyone's discernment. The substrate's `verify` gate *becomes* "run the build + test suite in the sandbox." The agent iterates autonomously against real compiler/test output until green — bins 1–2, safe because the value is ground-truth-checkable. This is the strongest fit the substrate has ever had for autonomous work.
- **The code itself = Proposal (the stones).** Generated, then inspected — a PR on an agent branch.
- **Deploy = the Hinge (always-his).** Outward-facing, not a cheap walk-back. This is the **always-escalate rung** the delegation-pattern audit proposed; this proposal is its first real consumer. The agent may build a green, tested, deployable artifact autonomously; *finalizing a deploy* escalates to Michael.

## 3. Ratified decisions (2026-06-03)

- **D-CC1 — Native coder tools (not orchestrate-opencode).** A new `coder` MCP server the substrate's own agent loop drives (like `git-mcp`/`fs-read-mcp`), so every action is substrate-visible and gate-able. We reimplement what opencode solved, but keep full control of the MCP + gate model.
- **D-CC2 — Hardened container + git worktree, switchable network.** Per-task ephemeral Docker container on an isolated worktree, `--cap-drop=ALL`, non-root, resource-limited. **Network is per-task switchable, default ON** (the agent must pull `go mod` / npm / pip / crates), **with an offline mode** when we want it. Allowlisted egress (a package-registry proxy) is the hardening target, not a v1 blocker. (gVisor / microVM is the v2 isolation upgrade — see §8 risk.)
- **D-CC3 — Deploy-to-sidecar is in v1**, behind the always-escalate Hinge: the agent prepares + can dry-run the deploy in an ephemeral sidecar, but a human ratifies the actual deploy.

## 4. The critical safety constraint (non-negotiable)

The bridge currently mounts `/workspace:rw` at the repo root. A coding agent with shell access there could write `.git/hooks/post-commit`, or edit `CLAUDE.md` / `.claude/skills/*` / `.mcp.json` / `package.json` scripts — all of which execute **outside** any sandbox, on the host, on the next commit or session. That is durable remote code execution.

**Therefore the coder never runs shell against the live `/workspace`.** It runs in a *separate* git worktree (its own branch), inside its own container, never mounting the real repo's `.git` or the workspace root. The PR/merge back to the live repo is the trust boundary, and merges are reviewed (the Hinge). This is the single most important rule in the design; everything else is layered on top of it.

## 5. Architecture

```
work_item (pipeline=code-write)
   │
   ├─ sandbox-manager (docker socket)         provision: git worktree + ephemeral
   │     ↑ keyed by work_item id              coder-runtime container (net=on|allowlist|off)
   │
   ├─ coder-mcp  (new spawn target)           tools target THIS work_item's sandbox
   │     coder_write / coder_edit / coder_apply_patch
   │     coder_read / coder_glob / coder_grep
   │     coder_shell   (build/test/run inside the sandbox)
   │     coder_lsp     (diagnostics; language servers in the runtime image)
   │
   └─ verify gate = `coder_shell` runs build + test → PASS/FAIL is ground truth
         pass → advance ;  fail → revise (feedback = real compiler/test output)
```

### 5.1 The `coder-runtime` image
A `coder-runtime.Dockerfile` (sibling to `bridge.Dockerfile`) with the v1 language toolchains + their LSP servers (see D-CC4 for which). Non-root user, no host creds baked in. This image is what each per-task sandbox container runs.

### 5.2 The sandbox lifecycle
On a code-write work_item entering `implemented`: provision a git worktree (isolated branch `agent/code-write/<wi>-<slug>`) and spawn an ephemeral container from `coder-runtime`, mounting **only** that worktree (not `/workspace`, not the real `.git`), with the configured `network_mode`. On terminal states (verified / failed / deployed / abandoned): tear down the container and worktree. Per-task ephemeral lifecycle is the research's recommended posture.

Docker access for v1 is a **socket mount on the bridge** (the "trusted-tool" tier). The sandbox-manager (in the bridge, or inside coder-mcp) spawns/execs/destroys sandbox containers via the socket. Whatever holds the socket is the trust root — noted as a hardening point (§8).

### 5.3 The `coder` MCP server (D-CC1)
Modeled on opencode's tool surface, but every tool operates against the active work_item's sandbox (via `docker exec` into its container) rather than the host:
- `coder_write(path, content)`, `coder_edit(path, old, new)` / `coder_apply_patch(diff)` — file mutation inside the worktree.
- `coder_read` / `coder_glob` / `coder_grep` — read (can reuse fs-read patterns, scoped to the worktree).
- `coder_shell(cmd)` — run a command in the sandbox (build, test, run, package-install). Inside the isolated ephemeral container the agent may run freely — the boundary is *where consequences land* (the worktree + container), not what runs inside. No tree-sitter parse-and-prompt needed (that's opencode's host-exec mitigation; isolation replaces it).
- `coder_lsp(path)` — diagnostics from the language server, auto-runnable after edits.

### 5.4 The `code-write` pipeline + maturity ladder
A code-tuned ladder (per-pipeline `maturity_ladder`, D-H2 already supports this):

`raw → planned → implemented → tested → reviewed → deployed`

- **planned** — agent produces an implementation plan/spec (existing gate machinery).
- **implemented** — agent writes code in the sandbox (coder tools).
- **tested** — **the verify gate = build + test in the sandbox.** Ground truth. Pass → advance; fail → drop to `implemented` with the build/test output as revise feedback. This is the autonomous loop; it can run unattended because failure is self-evident.
- **reviewed** — gate-eval (quality) + commit to the agent branch + open a PR via existing `git-mcp`.
- **deployed** — **always-escalate rung (the Hinge).** Agent prepares the deploy (builds the artifact/image, can dry-run it in a throwaway sidecar with a healthcheck), then the work_item transitions to `awaiting_review`; a human ratifies before the real deploy fires. Never auto-finalized regardless of trust level.

### 5.5 Deploy-to-sidecar (D-CC3)
The built artifact runs in its own ephemeral Docker sidecar (build image → run container → healthcheck → expose/report). v1 target is the substrate's own dev surface (e.g., spin a service the agent wrote, prove it serves, report the URL/logs) — not production. Production deploy paths (Dokploy, etc.) stay behind the Hinge with scoped, short-lived, broker-injected credentials (never in the sandbox image, never host env).

## 6. Defense-in-depth (the safety contract)

Layered, per the 2026 consensus:
1. **Worktree isolation** — never the live repo / `.git/hooks`; merges reviewed (§4).
2. **Container hardening** — `--cap-drop=ALL`, non-root, `--security-opt=no-new-privileges`, CPU/memory/disk limits, ephemeral teardown.
3. **Switchable egress** (D-CC2) — default on for package pulls; offline flag; allowlist-proxy to known registries as the hardening.
4. **Scoped secrets** — no host-env inheritance; deploy creds via a short-lived broker, injected only for the deploy step, never baked into the image.
5. **The Hinge** — deploy always escalates; **deploy approvals are never cached** (each deploy is a fresh confirmation).
6. **Trust ladder** — the build/test loop can run autonomously per the trust level; the deploy rung ignores trust (always-his).

## 7. Phased build plan (C–F cadence — smoke before each commit)

- **CC.1** — `coder-runtime` image + sandbox-manager (worktree + ephemeral container, network toggle) + bridge docker-socket access. Smoke: provision/exec/teardown a sandbox.
- **CC.2** — `coder` MCP server (the tool surface), wired to the active work_item's sandbox; registered in `mcp_servers` + granted to `dev`/`coder`.
- **CC.3** — `code-write` pipeline + maturity ladder + **the build/test verify gate** (ground-truth revise loop). Smoke: a trivial task iterated to green autonomously.
- **CC.4** — LSP integration (diagnostics in the loop).
- **CC.5** — deploy-to-sidecar + the always-escalate Hinge rung (implements the audit's proposed rung). Smoke: agent builds + dry-runs a service; deploy waits for human ratify.
- **CC.6** — hardening: egress allowlist proxy, secret broker, resource caps; gVisor runtime evaluation (v2 isolation).

## 8. Risks + honest caveats

- **Running LLM-generated code is the highest-stakes capability the substrate will have.** The hardened-container tier (D-CC2) shares the host kernel — the research calls this "trusted-ish, not bulletproof"; a kernel CVE escapes. Mitigated for v1 by: single-dev-box, the agent working on *our own* repos (not arbitrary untrusted code), worktree confinement, and ephemeral teardown. **gVisor (`--runtime=runsc`) is the v2 hardening** and should land before this ever runs less-trusted or multi-tenant work.
- **Docker socket on the bridge** widens the bridge's blast radius (socket access ≈ host root). Acceptable for a single-operator dev box; a dedicated sandbox-spawner service or rootless/sysbox is the hardening.
- **Cost.** Build/test loops burn tokens (iterate-to-green) and CPU. The existing per-work-item cost cap + the trust ladder bound it; set a coding-specific cap.
- **The merge boundary.** Even with sandbox isolation, merging the agent's branch brings its code into the live repo. Review before merge (the Hinge) is the real control — the sandbox protects the *build host*, not the *decision to trust the code*.

## 9. Open decisions (for the next walk — not yet ratified)

- **D-CC4** — which language toolchains in v1's `coder-runtime` (Go-only first? Go + Node? + Python/Rust)? Smaller image = faster + smaller attack surface.
- **D-CC5** — egress: open default-on vs. allowlist-proxy from day one (go proxy / npm / pypi / crates only).
- **D-CC6** — how deploy targets are declared (a `deploy_spec` on the work_item? a per-project deploy recipe?), and what v1's deploy surface is (substrate-local sidecar only, or a path to Dokploy behind the Hinge).
- **D-CC7** — does the build/test loop gate on the trust ladder (trainee surfaces every advance) or run free once green? Likely: free to iterate, trust-gated only at `reviewed`.
- **D-CC8** — sandbox-manager home: inside the bridge, inside coder-mcp, or a dedicated service.

---

*This proposal builds, as a side effect, the always-escalate rung named in `docs/delegation-pattern-skills-and-gates.md` — the deploy gate is the first place the substrate distinguishes "good enough to advance" from "his call regardless." It's the cleanest possible first instance, because deploy is unambiguously the Hinge and build/test is unambiguously ground truth.*
