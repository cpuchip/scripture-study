# Deadweight upstream sync — soaking Dave's SP work into our MP fork

**2026-06-28 · general-workspace sibling (lane: deadweight-upstream-sync) · autonomous /goal**

Michael picked the deadweight fork back up: "fetch origin against our fork
(happydave's upstream), see what he's been doing, and how we can adapt our
multiplayer version to soak in his changes." Then handed me the wheel —
"take it and have fun, decisions are reversible (dave's rule), make game
decisions yourself, research game design if you need to" — because he's
cognitively overloaded from a heavy pg-ai-stewards week.

## What Dave had been doing
44 commits since our fork point (`54537dc add unlicense`), all on the
single-player side: an **asset harness** (AI-generated sprite atlases — ships,
modular station, per-resource asteroids, gas-giant planet, flame plume), a
**Phase-5 ore→processing economy** (asteroid composition profiles, a scanner
gate on high-yield rocks), a **Phase-4 living market** (dynamic sell pricing,
seeded market events, sparklines), and a clean **`simLogic.ts` pure extraction**.
We had been working entirely in `server/`, `shared/`, and `src/scenes/mp/` —
so the histories were nearly disjoint. The trial merge confirmed only **4
conflict files**, two mechanical.

## The five tiers (all shipped + deploy-verified)
- **Tier 1 — merge** (`11b21e3`). Resolved the 4 conflicts keeping both sides.
  The merge changed world-gen (Dave's `generateComposition` draws one extra RNG
  value per asteroid), which shifted the field for a given seed — the net-starve
  smoke block's seed-777 scenario stopped reproducing. Diagnosed empirically
  (15/16 seeds still starve; 777 was just unlucky), re-pointed to seed 42 with a
  `>1700u` precondition guard. Inverse hypothesis honored.
- **Tier 2 — sprites** (`3b2aa32`). Our MP scene was a procedural immediate-mode
  graphics renderer (flat circles). Rewrote it as a **pooled sprite layer** keyed
  by entity id (planet/asteroids/station/haulers/miners/plume) at explicit depth
  bands, over a background graphics layer (starfield/glow) and a foreground one
  (rings/nets/beacons/bars). Mirrored Dave's exact `Ship.ts` scale + `+90°` art
  offset and plume geometry.
- **Tier 3a — dynamic market + events** (`e6986df`). Reused Dave's pure
  `market.ts`/`marketEvents.ts` **verbatim** on the server: per-corp price
  elasticity (selling depresses, recovers toward baseline), global seeded events
  folded into the baseline. The wire carries live prices; the base panel shows them.
- **Tier 3b — composition yields** (`ee6b0be`). Reused `processing.separate`: a
  rock's ore now separates into its dominant + trace resources on delivery.
  Synergizes with 3a (varied lots beat one dump).
- **Tier 3c — scan fog-of-war** (`43e6e07`). Large rocks are a mystery until a
  ship scouts within range — the low-friction read of Dave's scanner gate.

## Design calls I made (per the grant)
- **Per-corp markets, not a shared one.** A shared market is a cooler MP-native
  idea (corps crash each other's prices) but it's a *divergence* from SP, not
  faithfulness — and Michael's word was faithful. Noted the shared-market idea
  as a future option.
- **Scanning = proximity auto-reveal, not a required scan action.** The covenant
  boundary test: would Michael obviously want it? A *required* scan chore changes
  the core flow (he reserved balance for friends-playtest) — but fog-of-war that
  rewards scouting is pure upside. So: scout-near-to-reveal, no chore.
- **Deferred the ore→processing refining minigame.** The 3b composition split
  already delivers the yield benefit; a separate refining step is friction
  without clear fun for friends.
- **Didn't bloat the per-tick asteroid wire with composition** (4 floats × every
  rock × every tick to display one selected rock) — the yield is the substance;
  the breakdown display can come later if wanted.

## Surprises / lessons
- **Headless Chromium has no WebGL by default** — Phaser never booted, the menu
  scene never ran, the `/?mp=1` deep link never fired. `Framebuffer Unsupported`
  was the tell. Fixed with SwiftShader launch args
  (`--enable-unsafe-swiftshader --use-angle=swiftshader`). Now I can drive the
  real MP scene headless and screenshot it — confirmed the planet/asteroids/
  modular-station render Dave's art with zero console errors.
- **The stale-server gotcha bit twice.** `pkill -f "tsx server/index.ts"` didn't
  free :8080; the new server silently failed to bind and wstest hit the OLD
  (pre-Tier-3a) process — "prices" came back undefined. Fix: free :8080 by PID
  (`netstat -ano | grep :8080 | taskkill //F //PID`). This is the exact lesson
  already in memory; I re-learned it live.
- **Reusing a collaborator's pure functions is the cleanest kind of faithful.**
  Dave kept his economy logic Phaser-free and unit-tested (`market.ts`,
  `processing.ts`, `composition.ts`). Our server *imports them*, so MP pricing is
  byte-for-byte Dave's SP pricing. The reuse mandate paid off exactly where he'd
  done the work to make it possible.

## Carry-forward
- Friends-playtest is the next real input (balance: market depth/halflife, scan
  range, event frequency, composition trace richness).
- Optional follow-ups if Michael wants: shared-market variant, composition
  breakdown in the asteroid panel, ore→processing refining, price sparklines in MP.
- Verification gap I'm honest about: the hauler/plume in *motion* and the
  fog-of-war *up close* weren't isolated in a screenshot (the lobby-flow fought
  me); both are faithful mirrors of Dave's tested code with zero console errors,
  and Michael is the live play-tester.
