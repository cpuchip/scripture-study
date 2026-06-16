---
lane: general-workspace
session_id: c4fef1d0-292c-4ad6-b6c5-76e2af1043c3
status: active
started: 2026-06-11T12:00:00
last_active: 2026-06-15T23:18:22
---

> Note: this lane was previously filed under the typo `general-workspase.md`;
> consolidated here (the hook-canonical spelling) 2026-06-13, typo file removed.

## Working on
- **D&D / storytelling craft — Ammon stewardship run 2026-06-15:** 17 research
  artifacts drafted → `projects/pg-ai-stewards-workspace/research/` (11 skills + 3
  personas + 3 templates; ledger `research/00-LEDGER.md`); report DELIVERED to the
  pg-ai-stewards inbox. Digested 6 GM Tips + 2 storytelling blogs + 1 voicing video.
  Findings: DM-presides-not-compels (→ gamemaster inherits presiding covenant);
  causal-momentum in 3 places (improv spine / Pixar-Kenn-Adams / Harmon circle);
  Laban voice→text. Honest /goals: **G4 17✓ · G5✓ · G1 6/15 · G2 2/5 · G3 1/5**
  (carry). ⚠ Aunty Tauny unconfirmed. Study `study/yt/dnd-craft-01-mercer-gm-tips.md`;
  journals `2026-06-15-dnd-craft-{mercer,stewardship-run}`. Nothing pushed.
- **Garrison design session 2026-06-14 (council w/ Michael):** proposal extended +
  refined — Go-only, **isolated harness, embedded-SQLite default** (Postgres =
  optional MCP power-up; supersedes pg-required-v1), LM Studio+Ollama built-in
  (OpenAI-compatible), extensions MCP/JSON-RPC/HTTP/WS + WASM (NO gRPC, no native
  plugin), **Self-extension Tiers 0–3 + build-the-door/hang-with-consent gate**.
  Still `dominion_in_council` / post-cut. Spec current; committed (no push).
- **Overnight 2026-06-14 (unattended, no big moves):** (1) ibeco/Dokploy triaged —
  box was never down (sshd banner live; hung dokploy-panel + 1828/dnd containers;
  Michael's SSH likely fail2ban'd on his IP); SELF-HEALED by morning except
  `dnd.ibeco.me` 404; did NOT reboot (right call — 3 apps were live). (2) Garrison
  landscape study written (`.spec/proposals/sovereign-coding-agent-landscape.md`):
  pi = lean exemplar, goose = MCP cousin minus governance, **Devstral Small 2** =
  tool-tuned local model answering open-Q#4; governance gap confirmed empty.
  Journal `2026-06-14-garrison-landscape-and-ibeco-triage.md`. Decides nothing.
- **Garrison / `garrison-cli` proposal WRITTEN + refined (2026-06-13)** —
  `.spec/proposals/sovereign-coding-agent.md`. Name ratified ("who drives it
  presides"). Two tiers: v1 Garrison-full = Docker+LM Studio+pg (all owned;
  pg-as-machine = presiding ledger + context engine + fast context switching);
  later Garrison-minimal = binary+local-model floor. Superpower = the presiding
  chain (Michael→steward→sub-agents, pg tracks all = watch_what_you_order with
  eyes). Awaiting council (`dominion_in_council`); post-cut. On the board.
- **Euclid digestion (2026-06-13)** — yt workflow on Petro's Euclid video →
  `study/yt/WGwRCw9TRyo-euclid-walk-by-definitions.md`. Verified truth.md +
  Lectures on Faith L1 as quasi-Euclidean ("walk by definitions"); honest seam =
  Euclidean form, not epistemology. **Euclid = build-the-oracle-first archetype.**
  Book downloaded to `books/Euclid/` (gitignored). Substrate carry: 5 learning
  modes (cite-the-warrant linter + Postulates block lead) — proposal-shaped,
  dominion_in_council, surface at a substrate council. Also fixed the reground-
  counter hook: cwd-relative → project-anchored → **per-session (keyed by
  session_id)** so 6 concurrent sessions don't share one counter (`reground.py`,
  `1d26a302`; docs/06; lesson in `project_claude_code_context_plugin`).
- **Preside study: COMPLETE + COVENANT RATIFIED (2026-06-12)** — study pushed
  (`e74e6e90`), council held, `presiding:` extension live in covenant.yaml
  (emergency-accounting + uniform-watching amendments). Open follow-on:
  walls-vs-compulsion audit of substrate mechanisms (§V).
- Done earlier this lane: session-lanes system (built + tested; this lane was
  the first); Callie rename + deference + name-sync; context statusline +
  post-compact grounding.

## Claims
- pg-ai-stewards-persona-host docker container: rebuilt + recreated 2026-06-11
  (current code: r21 + Callie + deference + sync). The native persona-host.exe
  duplicates are DEAD — do not relaunch; rebuild the container instead.

## Handoffs / notes
- 2026-06-11: board surgery done (active.md → lean; full ledger in
  .mind/archive/active-ledger-thru-2026-06-11.md).
