---
lane: hubble-frontier-game
session_id: e9674ba8-5f6c-41d4-a07a-35ebdcf24058
status: ended
started: 2026-06-19T00:00:00
last_active: 2026-06-24T11:44:06
---

## Working on
- **★★★ /goal "get to v1.5" REACHED + LIVE (prod af1d4cb), 2026-06-19/20.** Michael
  + his SON playing & having fun. 9 oracle-first commits (bugs first): ★ clock
  unification fixed "flung into the solar system" (ship flew on st.t=0, map drew
  Moon on server clock → see-vs-gravity mismatch; now all render on st.t + SOI
  rings) · burn-line freeze · MMO hardening (NaN-vessel guard + drop-invalid-on-load
  self-heals prod) · physics warp 1-4× on ground/atmo · flight-view zoom · SAS
  target-rel (T▲/T▼ rendezvous) · **phase-timed transfer planner** ("fly me to the
  Moon" button, oracle: arrives INSIDE Luna SOI) · economy (part costs/funds
  display/tech tree/lunar lander) · day-night terminator · game menu (pause-private
  -only/recover/quit) · competitive **contracts** (comsats/survey/landing,
  server-validated) · **Mission Control** ambient AI flight-director (MCP `say`
  takes the voice live). Oracles smoke 22/burntest 6/wstest 13. v1.5.0.
- NEXT (open, his go): more icons; interplanetary Mars+first-shipyard arc still
  deferred (heliocentric restructure, its own goal); rival AI agency for contracts.

- **★★ /goal "get to V1" REACHED + LIVE (prod 1beca88), 2026-06-19.** 7 commits
  oracle-first: V1-A SAS hold (retrograde-burn answer) · V1-B burn-timing fix
  (arm→warp→centred auto-burn, no over-burn; `burntest` oracle caught a latent
  autopilot-never-auto-staged bug) · V1-C right-click target (ctx-menu off) ·
  V1-D multi-node queue · V1-E onboarding (Flight Manual + objectives) + v1.0.
  Oracles smoke 19/wstest 10/burntest 6. **Scope call: V1 = polished Terra–Luna;
  Mars interplanetary arc DEFERRED to its own goal** (real heliocentric
  restructure — root≠launch-body, solar map, transfers; ROADMAP "next major arc").
  NEXT (his go): Mars+first-shipyard, sound, Veo trailer, sync co-flight.

- First Orbit — KSP-like 2D browser MP rocketry game, Hubble Frontier origin era.
- **★ LIVE at orbit.cpuchip.net** (Dokploy compose `51xwf0eCPZ1M0bME6Obqb`,
  project `6-NRRFHIyHdjncEa_DLxD`, auto-deploy on push; githubId reused
  `F3xRIOFkUcjxB6DWVkPKW`). Deploy fix: expose-only (no host-8080 publish —
  collided with deadweight). `/version`=2f50f88.
- **Phase 0 + Waves B–F all SHIPPED + LIVE** (8 commits): B patched conics
  (Terra↔Luna SOI, land on Luna), C economy (milestones/funds/science/"first
  to…" broadcasts), D maneuver nodes (plan/preview/warp/execute), E rocket
  configurator (engine clusters/tanks/stages, presets, localStorage), F live MP
  (standings board, flight presence, interpolation). Oracles smoke 19/19 +
  wstest 9/9; browser+two-client verified each wave.
- **★ v1 RUN 2 (G1–G4) SHIPPED + LIVE (prod 8cec983):** G4 art (Gemini key on
  pg-ai-stewards works — logo+planets+part-icons generated/downscaled/wired;
  Veo video ALSO accessible on the key, just a spend call) · G1 map zoom/pan/
  follow · G2 rooms+rejoin (public Frontier MMO + private rooms, localStorage
  rejoin) · G3 MCP buddy (`mcp/` server: list_rooms/room_state/say over the new
  game read/chat API; smoke+curl verified). smoke 19/19, wstest 10/10.
- **AI-buddy wire-in:** `claude mcp add first-orbit -- node projects/first-orbit/mcp/server.mjs`
  (env FIRST_ORBIT_URL=https://orbit.cpuchip.net). NOT auto-added to workspace
  .mcp.json (standing capability — Michael's call).
- NEXT (open, /goal-able): Mars + first-shipyard arc (v0.6), sound, tutorial,
  short Veo title trailer (his go), synchronous co-flight, server-side saves.
- **Gemini key:** copied to projects/first-orbit/.env (gitignored) as
  GEMINI_API_KEY for gen-assets; spent ~$0.50 of the $10 budget.

## Decisions made (under full stewardship grant)
- Title "First Orbit", domain orbit.cpuchip.net, 2D faithful (Michael's calls).
- Stack: Vite+Svelte5+canvas / Node ws authoritative / shared deterministic TS
  sim / Dokploy single-container (the deadweight harness).
- Crown jewel = shared/ two-regime sim (RK4 powered + analytic Kepler coast).
- Oracle-first: smoke.ts (11/11, incl. inverse hypothesis) + wstest (7/7).
- Images: free Gemini Nano Banana API for batch sprites; AI Pro app/Whisk for
  hero art. scripts/gen-assets/ + .env.example staged for Michael's key.

## Claims
- (none long-lived — test server + container torn down, test image removed)

## Handoffs / notes
- Repo: github.com/cpuchip/first-orbit (public). Gitignored from root.
- DEPLOY HANDOFF for Michael: (1) grant Dokploy GitHub-App access to
  cpuchip/first-orbit; (2) create Dokploy app from this repo's docker-compose,
  map orbit.cpuchip.net → game:8080, enable auto-deploy on push. I can drive (2)
  via the dokploy skill once (1) is done.
- Verified: smoke+wstest+compile+build green; browser full-loop 0 errors;
  Docker image builds + serves + /version==sha.
