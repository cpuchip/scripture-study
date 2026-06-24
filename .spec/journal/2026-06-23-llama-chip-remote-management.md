---
date: 2026-06-23
topic: llama-chip remote management (the federation write half) + node-pinned remote chat
lane: pg-ai-stewards
---

# Manage any node from any UI, and chat a chosen remote

The carry-forward from the federation/hub/mesh arc: the local loader UI could *see* the whole
pool (the Pool panel, last session) but not *act* on it. This session built the write half —
remote management + node-pinned chat — so any node's UI is a control surface for every node.

## What shipped (`cpuchip/llama-chip` `7529828`)

- **`fed.NodeAddr(name) → mesh URL`** + **`fed.IsSelf(name)`**, both nil-safe — resolve a node
  name to its mesh address from the hub roster (hub mode) or the polled peer health (static mode).
- **`/api/remote?node=&op=`** — one proxy that forwards a single management call to a peer over
  the mesh and streams the response back, with the federation bearer attached. An op allowlist maps
  `op` → the local path the peer already exposes: GET reads (`status`/`gpu`/`models`/`backends`/
  `profiles`/`live`/`guess-context`) populate a remote node's forms; POST writes (`load`/`unload`/
  `unload-all`/`profile`/`ensure`) act on it. Unknown node → 404, unknown op → 400.
- **`proxyByModel` honours `?node=<name>`** — pins a chat to a chosen peer even when the local rig
  serves a model of the same name, and **strips the hint before forwarding** so the peer resolves
  locally (no double-hop / no loop). This is what makes "test a chat against a specific remote"
  work the same as local.
- **UI** — Pool cards gained per-node controls: **＋ Load…** (a form lazily populated with *that
  node's* GGUFs + profiles + GPUs over the proxy), **Unload all**, and an **✕** on each model chip.
  The Load form lives in a container decoupled from the 2s card refresh, so a refresh can't wipe it
  mid-edit. The Chat picker now groups by node — this node's models first, each remote node as an
  optgroup — and a remote pick routes with `?node=`. Self and peer share one `remoteCall` path
  (self → local `/api/*`, peer → `/api/remote`).

## Why it's safe the way it is

The peer's management endpoints are gated by **its** federation token over the mesh — the same
token the proxy forwards — so a shared-secret cluster authenticates and an open mesh-only cluster
(the live setup: no `token`, firewall-scoped to `100.64.0.0/10`) needs none. The mesh is the
boundary; the token is defense-in-depth. Forwarding our own `fed.Token()` (possibly empty) is
correct in every case, because asymmetric tokens would already break inference routing.

A remote-build controller can manage an **older-build** peer: `/api/remote` is the controller-side
endpoint; it proxies to `/api/load` etc., which predate this change. So fermion (still on the
pre-remote-management binary) was fully manageable from the new scratch node.

## The proof

- **Oracle** (`GOWORK=off go test ./...`, green): fed NodeAddr/IsSelf across hub + static + nil;
  router remote GET/POST/bearer-forward, unknown node/op, and the node-override picking the *named*
  peer on a name clash + asserting the `node` param is stripped before forwarding.
- **Live, non-disruptive** — a scratch node with **zero models** on `:18090`, federated to the
  running `dance-moe` rig as a static peer, *without touching the rig*:
  - `/api/pool` → scratch (its 2 GPUs) + fermion (qwen + gemma, from gossip).
  - `/api/remote?node=fermion&op=status|models` → fermion's real status + its 52 GGUFs.
  - `?node=fermion` chat → `qwen3.6-35b-a3b` answered **"pong"** (full content + reasoning,
    finish=stop). The headline feature, end to end.
  - Headless UI check: no JS console errors (only the favicon 404), Pool panel rendered both cards
    with controls, the remote Load form opened with fermion's 52 GGUFs + GPU boxes, the chat picker
    showed the `fermion (remote)` optgroup. Screenshot sent to Michael.
  - Negatives: ghost node → 404, bad op → 400.
  - Cleaned up: scratch process killed, binary/config/screenshot removed, rig re-confirmed healthy
    (both slots still `healthy`, untouched).

## Presiding / accounting

- The build ran entirely against a scratch process on a separate port. The substrate-depended rig
  was **read** and given **one short chat** (the "pong" proof) — no restart, no model load/unload,
  no profile switch on it. Confirmed healthy before and after. No emergency force used.
- Autonomy stays **paused** (Michael's call — keep the GPUs free for innovation week). Nothing in
  this change touches the autonomy switch.
- Michael granted me management of llama-chip "until it's stabilized" — this is inside that grant
  (no new standing capability; it extends the already-ratified federation).

## Carry-forward

- **His hands, optional:** test cross-machine on workchip (`~/code/vivint/workspace/
  ticket-innovation-week-20260622/llama-chip`): `git pull && GOWORK=off go test ./...` confirms the
  Linux build of this commit; running a federated node there would prove the real mesh hop for
  remote management (the loopback scratch test covered everything else; the mesh data path itself
  was proven last session). SSH to `workchip` resolves but needs my key or his password.
- **Deferred (documented):** *automatic* federated placement — a node deciding which peer has free
  VRAM and loading there on its own. The manual hook ships now (`/api/remote?op=load|ensure` against
  a chosen node); auto-selection across the pool is the follow-up. Not needed for `dance-moe`.
- llama.cpuchip.net (the hub) is unchanged by this commit — its Dokploy redeploy on push is a
  functional no-op.

## Update — cross-machine WRITE path confirmed on workchip (same day)

Michael updated workchip to the new build and asked to "test out gemma-4-12b on workchip," then to
exercise profile/model changes. A new-build controller on the home box (scratch :18090, peering
both fermion + workchip over the **real NetBird mesh**) drove it end to end — the one path the
loopback scratch test couldn't reach:

- **Pool** saw all three nodes (scratch + fermion[qwen, gemma-4-26b-a4b] + workchip[gemma-4-12b]).
- **Remote reads** of workchip over the mesh: slot healthy, RTX 3500 Ada 11457/12282 MiB.
- **`?node=workchip` chat → gemma-4-12b** answered a real MoE-vs-dense question correctly at
  **47.5 tok/s** (mesh overhead negligible — ~13s was almost all generation).
- **Write path:** `op=profile` proxied through (clean "no profile" error — workchip has none); then
  unload gemma → load `Qwen3-4B-Instruct` (healthy + chatted) → unload → **restore gemma-4-12b**
  at ctx 376832 / parallel 2 / q8_0. Restore was **exact**: state healthy, ctx/parallel matched,
  GPU0 back to 11457/12282 MiB (identical footprint — the q8_0 KV guess was right), chat "Hi!".
- **Finding:** `gemma-4-12b-it-QAT` is a **reasoning model** (`reasoning_content` + `content`) — a
  tight `max_tokens` (400) returns empty content because the budget is spent thinking; ≥~600 lets
  it finish. UI default (4096) + substrate dispatch (16000) are both fine. Same shape as the qwen3
  thinking-budget gotcha.

So the **carry-forward cross-machine test is DONE** (not just "his hands optional"): remote
management + node-pinned chat work over the real mesh, and a remote unload→load→restore cycle is
byte-for-byte clean. Scratch controller torn down; workchip left exactly as found. Autonomy stayed
paused throughout.
