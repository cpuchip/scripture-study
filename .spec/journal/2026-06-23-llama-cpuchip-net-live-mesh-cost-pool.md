---
date: 2026-06-23
topic: llama.cpuchip.net LIVE + the mesh + the deepseek cost gap + gemma rename + linux merge + pool view
lane: pg-ai-stewards
---

# The hub goes live, the mesh comes up, and the pool becomes real

The continuation of the federation/hub arc (journaled 2026-06-22) — this is everything after the
hub was built: deploying it, standing the mesh up, joining the first node, and the handful of
real things that surfaced once compute was actually flowing.

## llama.cpuchip.net — deployed + live

Michael granted the deploy (DNS wildcarded, Dokploy GitHub app on all repos). I drove it via the
**NOCIX Dokploy API** (`server.ibeco.me`, `DOKPLOY_NOCIX_API_KEY`): created the `llama-hub` compose
under his `cpuchip.net` project (github source `cpuchip/llama-chip@main`, path
`cmd/llama-hub/docker-compose.yml`, githubId reused), set the admin-key env, added the domain →
`:8088` letsencrypt. Built + healthy in ~10s; verified the whole flow over the real domain (UI,
admin login, mint, 401-inverse). Generated his admin key + a `home-4090s` node token, relayed both.
Kept the repo join docs generic (`<HUB_URL>` placeholder) per his ask — the live UI fills the real
URL in.

## The mesh + fermion joined

Michael stood up **NetBird at `mesh.cpuchip.net`** and enrolled two peers: `fermion 100.110.60.2`
(the home 2×4090 box) + `workchip 100.110.207.32` (his Ubuntu work laptop). I wired fermion:
`config.json` → node_name `fermion` + advertise the mesh IP; **graceful rig restart** (unload → kill
→ rebuild the federation binary → serve → reload dance-moe), done while autonomy was already paused,
so low-risk. fermion now sits in the hub roster serving qwen + gemma over the mesh. The **firewall
rule** (inbound `:8090` scoped to `100.64.0.0/10`) is the one piece left to his hands — it needs
admin, and Windows blocks the inbound otherwise.

## The deepseek cost gap — the day's best find

Michael noticed deepseek-v4-pro burning ~$4/day and asked what triggered it. The trace:

1. The `summarize_url`/`summarize_doc` subagents are configured for **qwen3.7-plus**, which hit its
   **weekly usage limit** on opencode's go plan (HTTP 429 GoUsageLimitError).
2. The capability fallback then chose **deepseek-v4-pro** — for 82 summaries that day.
3. **deepseek-v4-pro was priced $0** in the substrate ("go subscription; rate unpublished"). So
   `micro_dollars` recorded 0 — invisible to the spend caps + the watchman guard — *and* the
   cost-aware fallback PREFERRED it precisely because it looked free. **The $0 mispricing made the
   optimizer pick the expensive model.** That's the irony worth keeping.

Michael's correction mattered: **zen = the free side, go = paid.** So the three "go subscription /
rate unpublished" models (deepseek-v4-pro, mimo-v2.5-pro, mimo-v2-omni) at $0 were the real holes.
Fixes (live + overlay `ee23474`): priced all three (deepseek input $1.30/Mtok grounded in the
observed ~$4/day), and routed the summarize subagents to **local gemma (`ingest`)** — free, private,
and it uses the idle GPU1 (qwen pegs GPU0). The repricing also redirects every other qwen3.7-plus
stage's fallback to genuinely-cheaper models, not deepseek.

## gemma-12b → gemma-4-26b-a4b

Michael: "it's the 26B-A4B MoE, list it properly." Confirmed the new name via a quick ask
(`gemma-4-26b-a4b`, mirroring `qwen3.6-35b-a3b`). Renamed atomically across the dispatch contract —
llama-chip (`4c825a7`) + overlays (`5eb8f1d`) + live substrate (model_aliases ingest, capability,
pricing, judge_dispatch_model) + the running rig's slot reswapped — while autonomy was paused (no
dispatch-mismatch window). Tests green, zero stale refs, historical cost_events left alone.

## linux-port reviewed + merged

`origin/linux-port` (from the laptop side) — a clean OS-conditional port: every Windows path
replicates the prior hardcoded behavior, Linux/macOS get correct paths (libggml-base.so,
llama.cpp-linux-x86_64- dir match, LD_LIBRARY_PATH). Validated e2e on Ubuntu + the 3500 Ada,
*including routing a substrate chat to fermion's 4090s over the mesh* — which is the proof the
cross-LAN data path works. Reviewed, merged (`23d76db`), both OSes cross-compile, tests green.

## Pool view — local UIs see the whole federation

Michael asked: do local clients see the other GPUs? They didn't (only the hub UI showed the
roster). Built a **Pool panel** (`c26e47f`): the local loader UI now shows every node's GPUs +
models, from the hub roster the node already holds (self freshened from the local box). Headless-
render verified. The **write half** — remote management (load/unload/apply-profile on a peer,
proxied over the mesh) — is designed and is the next build.

## Presiding / accounting

- Deployed a public multi-user service (llama.cpuchip.net) — ratified last session via the
  AskUserQuestion council (NetBird mesh + tokens-now); accounted here + to Michael + the lane.
- Restarted the substrate-depended rig **twice** (mesh wiring, gemma slot) — both graceful (unload
  first), both while autonomy was already paused, both authorized by Michael's direct asks, both
  accounted. The running rig now lives in this session's background → flagged to Michael that he
  must relaunch in his own terminal for persistence.
- Live DB changes (cost pricing + the gemma rename) all persisted to the workspace overlays so a
  rebuild keeps them. The deepseek-v4-pro prices are marked ESTIMATE (rate unpublished).

## Carry-forward

- **Remote management (write half)** — `/api/remote` proxy + per-node Pool controls. The next build.
- **His hands:** the fermion firewall rule; relaunch the rig in his own terminal (persistence);
  the workchip federation block (the `lck_oi4x…` token) to test routing from the laptop.
- **Autonomy is still paused** (was already, pre-session). Offered to investigate why + resume —
  likely a leftover watchman trip from the deepseek summarize burst, now fixed at the root.
- **Flagged for confirmation:** deepseek-v4-flash + mimo-v2.5 "claimed free on the paid go side" —
  price them or leave? And a `provider_spend_caps` row for opencode_go (offered, not done).
- Filed to general-workspace inbox: a shared transactional email/SMS service (pairs with the hub).
- Eventually: vivint migrates to the laptop substrate (the two-substrate plan), once it's stood up.
