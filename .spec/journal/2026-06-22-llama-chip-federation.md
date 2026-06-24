---
date: 2026-06-22
topic: llama-chip — GPU federation across machines (one endpoint, local-first peer mesh)
lane: pg-ai-stewards
---

# Federation — pool GPUs across machines behind one endpoint

Michael: *"lets kick off the llama-chip work I'll need it today."* Last session he'd ratified
(via AskUserQuestion) the **federated llama-chip** shape over my recommended alternative, plus
corrected my topology assumption: the home 2×4090 box keeps its substrate; a **separate**
substrate runs on his Ubuntu work laptop (3500 Ada, 12GB) for work + vivint, and takes vivint
over. So today was the build.

## What shipped (`cpuchip/llama-chip` `181ae8d`, pushed)

A peer-federation layer so several llama-chip nodes pool their GPUs behind one `:8090` — the
LM-Studio-over-Tailscale trick, built in.

- **`internal/fed`** — registry + pull-only gossip poller. Each node polls every peer's
  **local-only** `/api/fed/local` (falls back to `/v1/models` for an older peer build), keeps a
  live `model → peer` map, evicts a peer's routes the instant it drops. `Resolve` (exact then
  substring), `RemoteModels`, `Peers`, `Refresh`. Nil-safe so a standalone node is untouched.
- **router** — `proxyByModel` now resolves local first, then a peer (forwarding the bearer
  token); new `/api/fed/local` gossip endpoint; `/v1/models` aggregates local+remote
  (`owned_by peer:<name>`); `/api/status` gains a `federation.peers` section; an optional
  bearer-token middleware that gates **non-loopback** callers only (the local substrate/browser
  stay exempt, `/health` always open).
- **config** — a `federation` block `{node_name, advertise, token, poll_interval_sec, peers[]}`.
- **`docs/federation.md`** — the design, config reference, the NetBird mesh runbook, the
  two-substrate topology, and the file_private/data-safety note.

## The design call that mattered: local-first, not head/worker

A strict head/worker would strand the laptop whenever the home box is off — which is most of his
workday. So the shape is a **peer mesh, local-first**: every node always serves its own GPUs with
zero dependency on any peer, and a peer's models are reachable only while the mesh reaches it.
Laptop alone → its 12GB. Laptop + home → the whole pool. No single point of failure. That's
strictly better for a roaming work laptop, and it's what the tests encode (evict-on-drop,
reappear-on-return).

## Oracle-first, then a real live proof

The repo had **no tests** before this; I added the first two suites (`internal/fed`,
`internal/router`) with `httptest` peers — deterministic, no GPU needed: learn/evict/return,
substring, first-peer-wins-tie, token forwarding, `/v1/models` fallback, auth gating, standalone
unaffected, and the 404 inverse. All green.

Then the live proof, **without disturbing the running rig**: a scratch head on `:18090` with
**zero local GPUs**, peering the live `dance-moe` rig on `:8090` read-only. It learned the rig's
models (via the `/v1/models` fallback — the live rig is the pre-federation binary), aggregated
them, and **served a real `qwen3.6-35b-a3b` completion ("federation works")** by routing to the
rig. A model nobody serves returned 404.

## Presiding / accounting

The live rig is the substrate's inference dependency and a manual host process another session
launched — so I did **not** touch it. I built to a scratch binary (`llama-chip-fed.exe`, since
removed) and a scratch head on a high port; the only contact with `:8090` was read-only GETs + one
tiny completion (the rig's normal job). Confirmed `:8090` healthy before and after. The production
`llama-chip.exe` is unchanged — Michael rebuilds and restarts it himself when he adds a federation
block (a running Windows exe can't be overwritten, and restarting the rig would be force on the
substrate depending on it). No force used; nothing of a sibling's disturbed.

## One surprise (not a bug)

The first live completion came back with **empty content**. That's the known qwen3.6 thinking
behavior (memory `reference_lmstudio_qwen3_thinking_budget`): it always reasons, and `max_tokens:20`
got eaten by the reasoning. Bumped to 2000 → clean `"federation works"` in `content`, reasoning in
`reasoning_content`. The routing was never in doubt — the response shape and `model` field were
correct the first time.

## Carry-forward (his hands — the mesh + laptop)

- **NetBird control plane** on NOCIX `server.ibeco.me`; enroll the home box + laptop; put the
  `100.x` mesh IPs into each `federation.peers`. Runbook in `docs/federation.md`.
- **Laptop substrate** (Ubuntu 24): fresh pg-ai-stewards stack + the vivint overlay + llama-chip
  serving its 12GB and federating the 4090s.
- **Migrate vivint** to the laptop substrate (enable `vivint-cron` there, disable on the home box).
  The vivint pipeline + overlay are portable — same core, different DB.
- **Deferred:** federated auto-placement (load-on-demand on a remote node — unneeded while
  dance-moe stays loaded); a cross-machine GPU panel in the loader UI (the data is already in
  `/api/status`).
