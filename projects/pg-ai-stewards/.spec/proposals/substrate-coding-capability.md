# Proposal — Substrate Coding Capability (write · build · test · deploy)

**Status:** ✅ **v1 COMPLETE 2026-06-03 (CC.1–CC.6 shipped + verified).** The substrate writes, builds, tests, diagnoses, and deploys code in isolated sandboxes — autonomously where there's ground truth (build/test green), with a human at the always-escalate Hinge for deploy. Commits `ee1760a`→`3297b5a`; migrations cc2–cc6; `coder` MCP = 13 tools. Proven end-to-end (CC.3: agent wrote+tested Go, verified by hand; CC.5: built+deployed+healthchecked a web server). All D-CC1–D-CC8 + trust posture ratified (two AskUserQuestion batches). **Deferred v2:** CC.3.1 (deterministic gate, "both paths usable") + CC.7 (Dokploy w/ scoped access). Build log + lessons in the 2026-06-03 coding-capability journals.
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
- **D-CC2 — Hardened container + git worktree, switchable network.** Per-task ephemeral Docker container on an isolated worktree, `--cap-drop=ALL`, non-root, resource-limited. **Network is per-task switchable, default ON** (the agent must pull `go mod` / npm / pip), **with an offline mode** when we want it. (Medium-safe posture ratified — shared host kernel accepted; gVisor is not a prerequisite. See the trust-posture entry below + §8.)
- **D-CC3 — Deploy is in v1**, behind the always-escalate Hinge: the agent prepares + can dry-run the deploy in an ephemeral sidecar, but a human ratifies the actual deploy. (Scope of v1 deploy = local sidecar; see D-CC6.)
- **D-CC4 — v1 `coder-runtime` languages: Go + Node/TypeScript + Python** (+ each language's LSP server). Covers tools/programs (Go), websites (Node/TS/Vue), and scripting/data/glue (Python) — nearly everything Michael writes.
- **D-CC5 — egress open, default-on, switchable offline.** No allowlist-proxy required for v1 (medium-safe posture); per-task offline mode available. Allowlist-proxy is a later knob, not a prerequisite.
- **D-CC6 — deploy is BOTH, phased.** **v1 = substrate-local ephemeral sidecar** (build → run → healthcheck → report). **v2 = Hinge-gated Dokploy deploy** of real sites/services — which requires substantial extra work: the substrate gets its **OWN scoped Dokploy access** (separate project namespace / scoped token / sub-account) so its deploys cannot touch or disturb the existing apps (ibeco / cpuchip / marsfield / 1828 / tinyfarm / hmslogs). That isolation is a v2 design problem in its own right (see §9).
- **D-CC7 — the build→test loop runs free; trust-gates only at the PR (`reviewed`) rung.** The loop is ground-truth-checked (build + test), so it iterates to green autonomously regardless of trust; the trust level only decides whether the finished PR auto-advances or surfaces. Deploy always escalates (the Hinge), trust notwithstanding.
- **D-CC8 — sandbox-manager lives in `coder-mcp`, keyed by work_item id** (provision worktree + container on entering `implemented`; tear down on terminal state). Simplest, reversible; extractable to a dedicated service later.
- **Trust posture (ratified).** This is *our own code, for our own purposes — medium-safe.* Sharing the host kernel (hardened container, no gVisor/microVM) is accepted. gVisor/microVM is **not** a v1/v2 prerequisite; it becomes relevant only if the substrate ever does **remote coding / agent-work-through-tunnels (SSH or similar) — a noted future seed** — or runs less-trusted / multi-tenant code.

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

### 5.5 Deploy (D-CC3, D-CC6) — phased
**v1 — substrate-local sidecar.** The built artifact runs in its own ephemeral Docker sidecar on the substrate host (build image → run container → healthcheck → expose/report). Proves "it runs" with no external blast radius. Behind the Hinge (the agent prepares + dry-runs; a human ratifies).

**v2 — Hinge-gated Dokploy deploy.** Deploy real sites/services to Dokploy (the cpuchip / marsfield / 1828 pattern). This needs its own design pass: the substrate gets **scoped Dokploy access** — a separate project namespace and a scoped token (or sub-account) — so its deploys are walled off from the existing apps and cannot redeploy, edit, or break them. Credentials are short-lived and broker-injected for the deploy step only (never in the sandbox image, never host env). Each deploy is a fresh Hinge confirmation; approvals are never cached.

## 6. Defense-in-depth (the safety contract)

Layered, per the 2026 consensus:
1. **Worktree isolation** — never the live repo / `.git/hooks`; merges reviewed (§4).
2. **Container hardening** — `--cap-drop=ALL`, non-root, `--security-opt=no-new-privileges`, CPU/memory/disk limits, ephemeral teardown.
3. **Switchable egress** (D-CC2) — default on for package pulls; offline flag; allowlist-proxy to known registries as the hardening.
4. **Scoped secrets** — no host-env inheritance; deploy creds via a short-lived broker, injected only for the deploy step, never baked into the image.
5. **The Hinge** — deploy always escalates; **deploy approvals are never cached** (each deploy is a fresh confirmation).
6. **Trust ladder** — the build/test loop can run autonomously per the trust level; the deploy rung ignores trust (always-his).

## 7. Phased build plan (C–F cadence — smoke before each commit)

- **CC.1** — `coder-runtime` image (Go + Node/TS + Python + LSP servers, D-CC4) + sandbox-manager in coder-mcp (worktree + ephemeral container, network toggle, D-CC8) + bridge docker-socket access. Smoke: provision/exec/teardown a sandbox.
- **CC.2** — `coder` MCP server (the tool surface), wired to the active work_item's sandbox; registered in `mcp_servers` + granted to `dev`/`coder`.
- **CC.3** — `code-write` pipeline + maturity ladder + **the build/test verify gate** (ground-truth revise loop; free to iterate per D-CC7). Smoke: a trivial task iterated to green autonomously.
- **CC.4** — LSP integration (diagnostics in the loop).
- **CC.5** — **v1 deploy: local sidecar** + the always-escalate Hinge rung + trust-gate at the PR rung (D-CC7) (implements the audit's proposed always-escalate rung). Smoke: agent builds + runs a service in a sidecar, healthchecks it; deploy waits for human ratify.
- **CC.6** — hardening: secret broker, resource caps; (egress allowlist proxy + gVisor are *optional* later knobs, not required by the medium-safe posture).
- **CC.7 (v2)** — **Dokploy deploy** (D-CC6): the scoped-access design (separate namespace + scoped token, walled off from existing apps) + the deploy path, Hinge-gated. Its own ratification pass before build.

## 8. Risks + honest caveats

- **Trust posture — ratified medium-safe (2026-06-03).** This builds our own tools/programs/websites for our own purposes, so the hardened-container tier (shared host kernel) is accepted: a kernel CVE could escape, but the threat model is *our code, not arbitrary adversarial code.* Worktree confinement + ephemeral teardown + the §4 live-repo rule remain the real boundaries. **gVisor/microVM is explicitly NOT a prerequisite** — it re-enters the picture only if we add **remote coding / agent-work-through-tunnels (SSH or similar — a future seed Michael flagged)** or run less-trusted / multi-tenant code.
- **Docker socket on the bridge** widens the bridge's blast radius (socket access ≈ host root). Acceptable for a single-operator dev box; a dedicated sandbox-spawner service or rootless/sysbox is the hardening.
- **Cost.** Build/test loops burn tokens (iterate-to-green) and CPU. The existing per-work-item cost cap + the trust ladder bound it; set a coding-specific cap.
- **The merge boundary.** Even with sandbox isolation, merging the agent's branch brings its code into the live repo. Review before merge (the Hinge) is the real control — the sandbox protects the *build host*, not the *decision to trust the code*.

## 9. Remaining design work (deferred to CC.7 / v2 — not blocking v1)

All v1 decisions (D-CC1–D-CC8 + trust posture) are ratified. What's left is the v2 Dokploy pass and a couple of mechanisms that v1 can stub:

- **Dokploy scoped-access design (CC.7 / v2)** — the real work behind D-CC6: how the substrate gets its own walled-off Dokploy footprint (separate project namespace + scoped token, or a sub-account) so its deploys cannot reach the existing apps, plus the deploy path itself and how a deploy is declared (a `deploy_spec` on the work_item? a per-project recipe?). Gets its own ratification pass before CC.7 is built.
- **Deploy-target declaration** — even for v1's local sidecar, how the agent says "run this artifact as a service on port X with healthcheck Y" (likely a small `deploy_spec` jsonb on the work_item). Settle at CC.5.
- **Egress allowlist proxy** — optional later hardening (registry-only egress) if the open-default-on posture ever feels too loose. Not required.
- **Future seed — remote coding / agent-work-through-tunnels** (SSH or similar). Michael flagged this 2026-06-03. If/when it lands, revisit the isolation tier (gVisor/microVM) and the network/secret posture — remote/less-trusted work changes the threat model.

---

*This proposal builds, as a side effect, the always-escalate rung named in `docs/delegation-pattern-skills-and-gates.md` — the deploy gate is the first place the substrate distinguishes "good enough to advance" from "his call regardless." It's the cleanest possible first instance, because deploy is unambiguously the Hinge and build/test is unambiguously ground truth.*
