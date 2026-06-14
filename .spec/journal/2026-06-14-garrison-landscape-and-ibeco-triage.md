# 2026-06-14 — Garrison landscape research + ibeco incident triage (overnight, unattended)

**Mode:** research (unsupervised) + ops triage · **Lane:** general-workspace

Michael went to sleep asking for safe overnight work to wake up pleasantly
surprised by, with an explicit "no big moves." Two threads.

## ibeco / Dokploy incident (triaged, NOT touched)

Michael's Dokploy box (`server.ibeco.me`, 204.12.235.154) seemed down — SSH
timing out, dashboard showing no control. Probed reachability instead of reaching
for the NOCIX reboot API he'd linked. Findings:

- **The box was never down.** TCP 22/80/443 open; sshd answered its banner live
  (`OpenSSH_9.6p1`). Traefik healthy. The failure was narrow: the **Dokploy panel
  container** was hung (502 at server.ibeco.me, :3000 dead), and a few app
  containers (1828, dnd) were down — pattern of a partial container kill (OOM or
  disk), not hardware.
- **SSH "timeout" = almost certainly his IP fail2ban-blocked on :22** (port answers
  instantly from my IP). Fix is connect from another IP and unban — not a reboot.
- **A reboot would have been the wrong move** — three prod apps (ibeco/engine/chat)
  were serving 200 the whole time; a power-cycle risks them to fix a hung
  container. The foresight discipline ("does the evidence support *this* action?")
  paid: the linked NOCIX "disconnect" call is a network null-route, not even a
  reboot.
- **Self-healed by morning:** server.ibeco.me → 200, 1828 → 200; only **dnd.ibeco.me
  still 404**. Confirms transient pressure, not hardware. Left entirely for Michael.

Durable: ibeco topology + "triage before power-cycle" — most ibeco outages are a
hung container or a fail2ban'd IP, not a dead box. Probe TCP+sshd-banner+per-vhost
before touching the NOCIX power API.

## Garrison landscape study (the pleasant surprise)

Wrote `.spec/proposals/sovereign-coding-agent-landscape.md` — a source-verified
(web, June 2026) evaluation of the existing opinionated coding agents, tuned to the
Garrison council's open questions. Identified the two Michael named:

- **pi** (Zechner, MIT) — the lean exemplar: 4 tools + ~300-word prompt, self-
  extending. Proof Garrison's lean core ships. Answers open-Q #2.
- **hermes** (Nous) — the contrast: persistent memory + cost-routing (borrow) under
  a 20-platform sprawl (reject).
- **opencode** (his dislike) — multi-provider TS framework, config sprawl, ~78%
  slower than Claude Code on one task; the anti-Pi maximalism the spec already
  rejects.
- **goose** (Block) — the architectural cousin: MCP-extension framework + model-
  routing, but "executes without approval" = the ungated autonomy Garrison
  corrects. Garrison = goose's shape + the governance it omits.
- **Devstral Small 2** (Mistral+All Hands, Apache-2.0, 24B, ~14GB Q4, ~58% SWE-bench)
  — tool-tuned local agentic model; the answer to open-Q #4 (weak-model tool-
  calling). Pattern: tool-tuned model for the loop + reasoner for planning.
- **Gap confirmed empirically:** every tool is ungoverned/lightly-governed. The
  governance niche Garrison claimed is real and empty.

Linked from the proposal's open-questions section. Decides nothing (unsupervised
scope = gather/evaluate/draft only). P1 dogfood rec: drive pi + Devstral as the
baseline before Garrison writes a line.

## For Michael's morning

- dnd.ibeco.me still 404 (restart its container; panel is reachable now).
- Verify the fail2ban hypothesis: SSH from a different IP, then unban yours.
- Read the Garrison landscape note when convenient — it sharpens #2/#3/#4/#6 for
  the council. Confirm "pi"/"hermes" are who I think (Zechner's pi, Nous's Hermes).
