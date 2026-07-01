---
lane: skunkworks
session_id: 70c5b4fd-7fee-46d5-97e8-77f04fa63f4b
status: active
started: 2026-06-29T00:00:00
last_active: 2026-06-29T23:36:06
---

## Working on
- **★ jubal-chip — MUSIC ENGINE, M0 scaffold SHIPPED + PUBLIC (`cpuchip/jubal-chip`,
  Apache-2.0, branch main).** Michael named it (Jubal, Gen 4:21). Go engine + Lua
  bridge in REAPER; oracle-floored; "all things musical" (not just a REAPER driver).
  Go DECIDED (evidence: protocol = 155 lines file-IO; mastering needs no Python;
  Python only at M3 for music21/librosa). Built+green: `internal/bridge` client +4
  unit tests, `cmd/jubal ping`, lean `jubal_bridge.lua` (M0), docs/README/CLAUDE.
  ★ **M0 DONE — live-verified HEADLESSLY 2026-06-29** (`12ffcb7`). No desktop computer-use
  tool here (only Chrome browser autom.), so drove REAPER from the shell via a temp
  `__startup.lua` auto-loading the bridge → `jubal ping` returned reaper=7.76/x64 live.
  Machine left CLEAN (REAPER killed, startup file removed, 0 procs). The headless
  launch→drive→verify→kill loop is now the dev harness. M1 recipe captured in
  `docs/roadmap.md` (MIDI insert learned from xDarkzx; render's RENDER_FORMAT blob =
  the fiddly slice, deferred to a focused build w/ REAPER open — didn't cram it at the
  tail of a long session). Wiring jubal as an MCP = standing capability + restart = his
  call, NOT done. NEXT: M1 construct (real JSON decoder + set_project/add_track/
  write_midi/render + in-key oracle); theory-knowledge digestion (his sources or
  canonical). Cloned for learning: `external_context/Reaper-MCP` (xDarkzx, the bridge
  pattern), `external_context/orpheus` (M0-only, the analyze→approve loop is the dream),
  `external_context/total-reaper-mcp` (shiehn). Scout: `.spec/skunkworks/music-tooling-scout.md`.
- **Michael-profile v1 — SHIPPED + committed** to `private/michael-profile/`
  (private repo `6f10864`, NOT pushed — offered to Michael). 7 docs + oracle.
  Centerpiece finding: AskUserQuestion acceptance FLAT (~77% rec, ~18% override)
  over the month, NOT rising — the "I pick your recs more" feeling = the June-1
  autonomy turn moved decisions OUT of the question layer. Overrides cluster on
  intent/vision/voice/strategy, never execution. Op line: automate execution,
  surface intent, widen bins-1-2 by building oracles.
- **✅ PROFILE PUSHED** to `cpuchip/private-study` (`6f10864`) 2026-06-29. NEXT: (b)
  monthly oracle refresh for a real longitudinal trend (track DIVERGENCE, not just
  convergence — so the profile informs, not cages).
- **✅ jubal-chip M1 FIRST LIGHT + theory floor DONE 2026-06-29** (pushed `5b6234d`):
  composed Ode-to-Joy in-key (oracle-gated) + rendered a real 2.2MB WAV headlessly +
  delivered to Michael; theory `docs/theory/fundamentals.md` digested from Open Music
  Theory (CC BY-SA 4.0). Render gotchas solved (RENDER_FORMAT=evaw; RENDER_FILE=dir+
  RENDER_PATTERN=filename). NEXT: composable M1 tools; M2 master; M3 the analyze→approve
  loop; wire jubal as MCP (his call+restart). Journal `2026-06-29-skunkworks-profile-oracle-jubal.md`.

## Shipped 2026-06-29 (after profile v1, Michael ratified the direction)
- **Book Part-3 handoff DONE + PUSHED** (`cpuchip/scripture-book` `c5a088f`):
  `seeds/the-covenant-over-time.md` (covenant forward through time; **Fable-loss
  = the emotional hinge** — Fable 5 was a Claude sibling who worked on the book +
  signed commits) + handoff banner on the book's active.md. Seed+provenance
  handoff only; manuscript stays the book session's stewardship (his directive).
- **Oracle floor SHIPPED** (root `f160432e`, commit-only): `scripts/study-lint/
  voice_lint.py` (2-tier: cut-list/meta HARD, em-dash ADVISORY; day-one caught
  "that changes everything" in art-of-presidency baseline) + `scripts/oracles/
  registry.yaml`+`run.py` (oracle→decision-class; reports green=act-and-report /
  blocked=surface-first; tested live, 3 green + voice blocked). "The oracle is
  the switch" applied to copilot-instructions.md. Proposal
  `.spec/proposals/oracle-floored-autonomy.md`.
- **NOT done (Michael's toggle):** the Stop-hook auto-running oracles — changes
  the session FEEL (intent-level), one settings.json line away on his word, not
  silently added. He's hesitant re hooks but said yes to the direction.
- **Skunkworks charter** `.spec/skunkworks/charter.md` — vision capture (aim =
  faith/hope/charity; projects: music[Ableton+MCP]/AI game companions/card
  games/orbit+deadweight/pg-ai-stewards-as-lore-engine/local-AI art; substrate
  holds the lore, pour projects in as intents/pools; Loreworks in motion).
  Black Magic award = proof. First creative thread to reach for: **MUSIC**.

## Claims
- (none — analysis + tooling; scripture-book pushed [granted], root commit-only)

## Handoffs / notes
- 2026-06-29 lane created (skunkworks). Special-projects lane.
- Built oracle `provenance/askquestion_oracle.py` (deterministic decision
  parser over the 14 Claude Code transcripts) — re-runnable, the durable
  provenance for the [measured] claims. Heavy corpus dumps gitignored
  (regenerable; raw prompts may carry client/work material).
- Used 2 parallel general-purpose agents (corpus comms-mining + journal arc) —
  both returned verbatim-sourced digests; synthesis is mine (shepherd).
