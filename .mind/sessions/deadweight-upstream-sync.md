---
lane: deadweight-upstream-sync
session_id: 69cd01f8-c8ce-49e3-85db-fea442f30b7f
status: active
started: 2026-06-27T00:00:00
last_active: 2026-06-27T00:00:00
---

## Working on
- Sync our deadweight MP fork (`cpuchip/deadweight-acquisitions-game`) against
  happydave's upstream SP. Add `upstream` remote, fetch, survey Dave's new work,
  plan how to soak his changes into our multiplayer version.

## Claims
- 2026-06-27 added `upstream` remote (happydave) + fetched. No master changes yet.

## Handoffs / notes
- Our fork shares history root with Dave's (`433aa27 Bootstrap`). We've hand-pulled
  before: `25d1530 "4 enhancements from Dave's SP pass"`.
- Merge base = `54537dc add unlicense`. Dave = **44 new commits**, us = **28** (MP).
- Trial-merge (aborted): only **4 conflict files** — package.json, package-lock.json
  (mechanical), src/ui/Hud.svelte (+65 Dave economy HUD vs +7 us), vite.config.ts
  (+1 Dave base-path vs +43 us proxy). Everything else lands as clean additions.
- Seam check: our server imports Dave's `worldGenerator`/`worldConfig`. His changes
  there are additive (AsteroidData +composition/+scanned) + a price bump — non-breaking,
  server keeps running, MP economy prices shift slightly.
- Dave's big new work: asset harness (committed atlases in public/assets — ship/base/
  asteroid/planet/flame), ore→processing pipeline + composition + scanner-gate,
  dynamic pricing + market events + sparklines, simLogic.ts pure extraction.
