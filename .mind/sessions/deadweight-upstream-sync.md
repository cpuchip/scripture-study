---
lane: deadweight-upstream-sync
session_id: 69cd01f8-c8ce-49e3-85db-fea442f30b7f
status: ended
started: 2026-06-27T00:00:00
last_active: 2026-06-29T20:43:31
---

## Working on
- /goal 2026-06-28 COMPLETE — faithfully executed Tiers 1→2→3, reused Dave's work,
  tested each step, commit+push+deploy-verified at each green boundary. All LIVE on
  deadweight.cpuchip.net. Goal satisfied.
- **Tier 1 LIVE** `11b21e3` — merge of Dave's 44 upstream commits (Phase 4/5 SP economy
  + asset harness). 4 conflicts resolved keeping both sides. Safety tag `pre-upstream-merge-2026-06-28`.
- **Tier 2 LIVE** `3b2aa32` — MP renderer draws Dave's atlases (planet/asteroids/modular-
  station/hauler/miner/plume) as a pooled sprite layer over bg+fx graphics.
- **Tier 3a LIVE** `e6986df` — per-corp dynamic sell market + global market events
  (reused market.ts/marketEvents.ts verbatim). Sell depresses price, recovers, events shift baseline.
- **Tier 3b LIVE** `ee6b0be` — composition-weighted yields (reused processing.separate);
  a rock yields its dominant + trace resources.
- **Tier 3c LIVE** `43e6e07` — scan-gate fog-of-war: large rocks unknown until a ship
  scouts within SCAN_RANGE (low-friction, no required scan action).
- Gates each tier: tsc + vitest 112 + smoke + vite build + wstest, all green; /version=deploy oracle.
- Playwright installed --no-save (node_modules, gitignored) + chromium-headless-shell;
  drive MP headless with SwiftShader args (`--enable-unsafe-swiftshader --use-angle=swiftshader`).
  Stale-server gotcha bit twice — free :8080 by PID (netstat→taskkill), pkill -f tsx isn't enough.
- DEFERRED (Michael's call, balance-sensitive): ore→processing refining minigame (the
  composition split gives the yield benefit without the separate refining step).

## The plan (awaiting go)
- **Tier 1 — Foundation merge.** `git merge upstream/master`; resolve 4 conflicts
  keeping BOTH sides (package.json/lock = mechanical + `npm i`; Hud.svelte + vite.config.ts
  = small real merges); re-run oracles (smoke/wstest/build/typecheck); verify SP + MP
  both play; commit + push. Reversible foundation, keeps fork a clean superset.
- **Tier 2 — MP sprites.** Wire `src/scenes/mp/MultiplayerScene.ts` to Dave's committed
  atlases (`public/assets/dwa_{ships,base,asteroids,planet}` + `fx_flame`). Separate
  renderer from his SpaceScene, so it's adaptation not merge. Most visible MP win.
- **Tier 3 — deeper economy into MP server (scope call).** Port scanner/composition,
  ore→processing, dynamic pricing, market events into `server/sim/world.ts`, reusing
  Dave's pure `simLogic.ts`/`market.ts`/`processing.ts`. Shifts MP balance; per-phase oracles.

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
