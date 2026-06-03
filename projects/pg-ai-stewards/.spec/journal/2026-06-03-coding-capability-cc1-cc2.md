# Substrate coding capability — CC.1 + CC.2 (the foundation gate)

**Date:** 2026-06-03
**Workstream:** WS5 / pg-ai-stewards
**Mode:** dev (full stewardship — Michael: "take stewardship of this and see this through the entire plan")
**Outcome:** the substrate can write, build, and run code in an isolated sandbox, through MCP tools. CC.1 + CC.2 shipped + verified.

## What shipped

**CC.1 (commit `ee1760a`) — the sandbox foundation.**
- `coder-runtime.Dockerfile`: Go 1.26 + Node 24 + Python 3.11 + LSP servers (gopls / typescript-language-server / pyright), non-root `coder`, idle by default. Built on the host daemon.
- `cmd/coder-mcp/` (new module): `sandbox/sandbox.go` — Provision/Exec/Teardown via the `docker` CLI against the host daemon. Hardened: `--cap-drop=ALL`, `no-new-privileges`, mem/cpu/pids limits, network toggle default-on (D-CC5). The worktree lives in the container's own ephemeral fs; teardown discards it; the coder never touches the live `/workspace` mount (proposal §4).
- bridge: `docker-cli` + `/var/run/docker.sock` mount + `coder-mcp` baked in; added to go.work.

**CC.2 (commit `b5f055c`) — the tool surface.**
- coder-mcp is now a stdio MCP server (official `modelcontextprotocol/go-sdk` v1.6.0 — the substrate spawn-target convention, matching git-mcp; NOT the mark3labs SDK the Claude-Code-facing servers use).
- Nine tools, each on a named `sandbox` (the work_item id): `coder_sandbox_start/stop`, `coder_write/read/edit/apply_patch`, `coder_shell` (build/test/run), `coder_glob/grep`. Registered in `stewards.mcp_servers` + 9 `dev` grants via `cc2-coder-mcp-seed.sql`.

## Verification

- CC.1 smoke (`coder-mcp -smoke`): provisions a hardened sandbox, prints go1.26.4/node24/py3.11 + 3 LSP versions, writes + `go run`s a program ("hello from the sandbox"), tears down. PASS.
- CC.2 smoke (full MCP tool-call loop via `docker exec -i bridge coder-mcp`): start → write main.go → shell `go build && ./app` (exit 0, "built by the coder MCP") → read → glob → stop. PASS. refresh-tools lists `coder` with all 9 tools.

## Surprises / lessons

1. **Login-shell PATH reset.** The first CC.1 smoke failed with `go: command not found` — `bash -lc` sources `/etc/profile`, which resets PATH and dropped the `/usr/local/go/bin` ENV entry. Node/gopls survived (they're in `/usr/local/bin`, which stays on the login PATH). Fix: symlink `go`/`gofmt` into `/usr/local/bin`.
2. **The MCP smoke harness, not the server.** First tool-call smoke returned nothing — the official go-sdk closes on stdin EOF before flushing responses when all requests are piped at once. The server was fine (refresh-tools had already proven init + tools/list). Fix: a proper request→read-response loop (Popen), closing stdin only at the end. "Verify via the real path" — refresh-tools (the bridge's own MCP client) was the real proof the server worked; my hand-rolled pipe was the broken part.
3. **The drift exit-2 rule held again.** Pre-applied `cc2-coder-mcp-seed.sql` via the running bridge before recreating it (4 drift files would make the entrypoint's migrate exit 2 on any apply → `set -e` → bridge fails to start). Restart then found nothing pending → exit 0 → clean.

## Where this sits in the plan

CC.1 (sandbox) + CC.2 (tools) are the foundation. The agent loop can now manipulate a sandbox; **CC.3 makes it autonomous** — the `code-write` pipeline + maturity ladder, where the `tested` rung runs build+test in the sandbox (the ground-truth verify gate) and a fail drops back to `implemented` with the real compiler/test output as feedback. That closes the Prescription loop from the delegation-pattern audit. CC.4 LSP, CC.5 deploy-to-sidecar + the always-escalate Hinge rung, CC.6 hardening. CC.7 (Dokploy) is the deferred v2.

## Carry-forward

- CC.3: investigate the existing pipeline/stage/gate/verify machinery (5a-maturity-gate, 5b-scenarios-verify) and model `code-write` on it; the `tested` gate calls `coder_shell` for build+test.
- The 9 coder tools are granted to `dev`; CC.3's pipeline agent grants come with the pipeline.
- Sandboxes are per-work_item, keyed by id; CC.3 provisions on entering `implemented`, tears down on terminal state (D-CC8 lifecycle).
