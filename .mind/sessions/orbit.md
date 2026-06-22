---
lane: orbit
session_id: e9674ba8-5f6c-41d4-a07a-35ebdcf24058
status: active
started: 2026-06-20T00:30:41
last_active: 2026-06-21T16:35:01
---

## Working on
- **★ /remote-control 3 asks SHIPPED + LIVE (prod e859aef), 2026-06-20.** (1) **Free
  navigation** (fixed: Locate did nothing from VAB — scene hidden there). Reworked
  screen model menu→VAB→flight into menu / **observe** (persistent live map of the
  program) / flight, with VAB + fleet as CLOSABLE overlays (Build/⊙Fleet anytime,
  ＋New rocket from fleet, Esc/✕ close; new program→VAB, returning pilot→observe map).
  Guarded a crash (observer map touched uninitialized game.st). (2) **Generalize Plan
  transfer to ANY target** — refactored `shared/transfer.ts` around a generic
  TransferTarget {radius,rate,angleAt,capture?}; moon=injection+capture, junk/ship=
  single injection timed to arrive alongside (then match-vel to salvage). Target btn
  adapts transfer↔salvage. (3) **Maneuver-node projection ghosts** — map ghosts where
  every body + junk will be when the active node fires (faint, linked to current
  pos); cycling nodes moves the ghost time (Luna capture-node = the arrival encounter,
  KSP-style). Oracle smoke 25 (+2 rendezvous), burntest 13, wstest 18. Browser-walked
  the full nav flow + junk transfer (+507 m/s rendezvous) + ghosts.
- **★ v1.6 BUG-FIX PASS SHIPPED + LIVE (prod 70436db), 2026-06-20.** Michael + son
  playing ("simplified KSP where the game helps you get places"). 4 fixes, oracle-
  first: (1) **TARGETING regression FIXED** (the blocker — pickTarget used
  universeTime() but map renders on st.t; after warp they diverge hours → right-
  click missed everything; v1.5 clock-unification adjacent-surface miss; worked in
  short tests = verify-under-real-conditions lesson). (2) **⊙ Circularize button**
  (key C) — one-tap plan+arm circularization at next apsis, any body (removes the
  O/U-at-high-warp pain; clock races past hand-set node times). (3) **autopilot
  leaves a round orbit** (82×91km vs grazing 72×90; button perfects to ~0 e, fuel
  permitting). (4) **RESUME a craft** — vessels persist design+fuel+stage; ▸ Fly on
  fleet + "⊙ Your fleet" at the VAB (gap: fleet was flight-only, unreachable after
  rejoin which lands at VAB). Oracles burntest 13 (+4) / wstest 18 (+3) / smoke 22.
  Verified the real rejoin→Fly scenario in a prod build.
- **★★★ First Orbit v1.6 SHIPPED + LIVE (prod c981dab, v1.6.0), 2026-06-20.**
  Michael's big v1.6 vision + "how would you like to participate?" 4 oracle-first
  commits, each auto-deployed:
  - **PiP minimaps** (headline): a 2nd canvas shows the OTHER view — flying →
    zoomable solar minimap, map → mini ship; click swaps, scroll zooms. Unified
    the frame loop (pause gates sim only, not draw).
  - **Space junk salvage**: 6 derelicts in Terra orbits (boosters/cargo/probe/
    relay/Unmarked Canister=45sci secret-tech seed). Right-click target → button
    guides you in (dist+rel-speed) → rendezvous <3km + matched-vel → salvage.
    SERVER-VALIDATES rendezvous vs real vessel. `shared/debris.ts`. wstest +2.
  - **Fleet page** (F / ⊙ Fleet btn): lists every craft aloft + orbit + Locate +
    Recover ("the Mission Control page that lists/zooms to your junk").
  - **Richer launch gfx**: launch pad/gantry on the ground + atmosphere glow limb.
  Oracles smoke 22 / burntest 6 / wstest 15 green. Researched KSP mods (KER/MechJeb/
  KAC = all already baked in; genre loves docking/sats/fleet = his junk instinct).
- **My participation = Mission Control** (ambient AI flight-director, greets +
  calls firsts/salvage; MCP `say` takes the live mic). v1.7 thread: rival AI agency
  racing for contracts.
- **DEFERRED (designed, told him, didn't rush):** MP *time controls* (shared-clock /
  warp-vote — the KSP subspace problem; own focused effort like Mars); scenarios/
  race mode; content (comets/sat-repair/rescue/payloads-that-do-things/aliens) all
  build on the junk+contract machinery now in place.
- Prior arcs (v1, v1.5) + Phase-0/Waves/G1-G4 history → journal
  `.spec/journal/2026-06-19-first-orbit-game.md` + memory `project_first_orbit_game`.

## Claims
- (none long-lived — all test servers on ports 9215-9217 + ./data torn down)

## Handoffs / notes
- Repo github.com/cpuchip/first-orbit (public, gitignored from root). Dokploy
  compose `51xwf0eCPZ1M0bME6Obqb`, auto-deploy on push; `/version`=deploy oracle.
- Gates before every commit: smoke && wstest && compile && build (CLAUDE.md).
- AI-buddy wire-in: `claude mcp add first-orbit -- node projects/first-orbit/mcp/server.mjs`
  (NOT auto-added to workspace .mcp.json — standing capability, his call).
