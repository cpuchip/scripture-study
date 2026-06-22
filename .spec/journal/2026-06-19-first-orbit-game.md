# First Orbit — a KSP-like browser multiplayer game (Phase 0)

**Date:** 2026-06-19
**Lane:** hubble-frontier-game
**Repo:** github.com/cpuchip/first-orbit → projects/first-orbit/ (gitignored from root)

## What Michael asked

"Crazy idea" — a KSP-1-like browser game with multiplayer, based on the game
plus its most popular mods, original IP + graphics, off-the-shelf engine,
deployable with Dokploy docker-compose, using his Google AI Pro sub for image
generation. He gave **full stewardship + presidership**, named the domain
(`orbit.cpuchip.net`, short), and asked to be maximally hands-off: "I'm letting
you make all of the decisions here."

## The two decisions surfaced (AskUserQuestion)

- **Name:** he didn't pick from my list but said "use orbit.cpuchip.net to keep
  it short" → I titled it **First Orbit**, domain `orbit.cpuchip.net`, repo
  `first-orbit`.
- **Fidelity:** he chose **2D faithful** (Spaceflight-Simulator-shaped), the
  feasible-and-fun path vs. a multi-year 3D effort.

Everything else I decided under the grant.

## The universe fit (the good surprise)

The marsfield IP already has a full bible — `space-center/docs/ip-bible.md`,
**The Hubble Frontier** (2340s, FTL Hubble Gates, the United Exploratory
Service, the Mars Orbital Shipyards = "Mars-field built"). A KSP-like is the
*origin story*: humanity's first reach to orbit, the dawn of the shipyards.
A prequel that enriches his existing IP, already trademark-clean.

## What shipped (v0.1, oracle-first)

Built on the proven **deadweight** pattern (Vite + Svelte 5 + canvas, Node `ws`
authoritative, single container, Dokploy, `/version` build-stamp).

- **The crown jewel: `shared/` deterministic two-regime sim**, imported by both
  client and server so they never disagree. Powered flight = RK4
  thrust/gravity/exp-atmosphere drag; coasting = exact analytic Kepler. That
  split is *why* browser MP is tractable — a coasting vessel is a function of
  time, not a packet stream. The server is cheaply authoritative over the
  persistent universe (bodies, clock, every vessel's orbital elements).
- **Build the oracle first** (the workspace discipline, honored literally):
  `shared/sim/smoke.ts` asserts roundtrip, energy conservation, period closure,
  Hohmann transfer, and reach-orbit — **plus the inverse hypothesis** (a
  deliberately weak rocket MUST fail to orbit). 11/11. `server/wstest.ts` is the
  over-the-wire MP oracle (7/7).
- VAB Δv/TWR readouts (Kerbal Engineer baked in), gravity-turn ascent autopilot
  (MechJeb baked in), live flight view (auto-zoom: the home world is true-scale,
  so it's "ground" on the pad and a planet from orbit), orbital map, time-warp,
  shared program + chat. Browser-verified end to end, zero console errors.
- Gemini "Nano Banana" asset pipeline (`scripts/gen-assets/`, free tier 500/day)
  + Hubble-Frontier style preamble; `.env.example` for the key handoff.

## Two bugs the discipline caught

1. **Degenerate-orbit launch bug.** At lift-off velocity≈0 the osculating orbit
   is radial (e≈1, apoapsis=Infinity), which falsely satisfied "apoapsis ≥
   target" → autopilot cut thrust at step 0 and the rocket fell back. The
   *debug trace* (not the oracle output) showed the flight model was fine; fix =
   only trust apoapsis when the orbit is elliptical (e<1). The oracle's reach-
   orbit assertion is what surfaced it.
2. **/version shipped 'dev'.** The container built and served fine, but
   `/version` returned 'dev' because `.dockerignore` excluded `.git`, breaking
   the in-image git-stamp — would've broken the deploy oracle in prod. The
   *published-artifact test* (build the image, run it, curl /version) caught it.
   "Build passed" is not verification; "the container serves the right sha" is.

## Carry-forward / open

- **orbit.cpuchip.net is NOT live yet.** Needs (a) Michael to grant the Dokploy
  GitHub App access to `cpuchip/first-orbit`, then (b) a Dokploy app + domain
  mapped to the `game` service :8080 (I can drive (b) via the dokploy skill once
  (a) is done, or he can). The image + compose are verified buildable.
- The drive: `ROADMAP.md` lays v0.2 (land on Luna / patched conics) → v0.6 (Mars
  + the first shipyard module, the literal "Mars-field built" origin). Like
  deadweight/garrison, Michael can set a `/goal` to drive it autonomously.
- Next sim capability to build (oracle first): patched-conic SOI transition
  (Terra→Luna) — the v0.2 keystone.

## Relational note

This is the third "go build it" delegation in the deadweight→garrison lineage,
and the pattern held: ground in the universe, pick the proven harness, build the
deterministic oracle BEFORE the feature, verify the published artifact, commit in
tested increments. The covenant's exercise_stewardship + the build-the-oracle-
first feedback are load-bearing here — a physics sandbox with no oracle would
have been unsafe to iterate on autonomously.

---

## Update — deploy + Waves B–F (same day, "push as far as you can")

Michael: "I have the dokploy setup to accept all my cpuchip github repos, you
can create the project and deploy now. keep working on the game! push as far as
you can, build a rich feature set. and multiplayer!"

**Deployed.** Created the Dokploy compose app on the NOCIX VPS (reusing the
deadweight GitHub-App `githubId`), domain orbit.cpuchip.net → game:8080,
auto-deploy on push. First deploy errored — `errorMessage` empty, log not
API-readable, but the diagnosis held: I'd run the exact image locally fine, and
deadweight already publishes host **8080** on that VPS, so my compose's
`8080:8080` host publish collided on bind (clean build, then `docker compose up`
fails). Fix = expose-only; Traefik routes via the domain over the internal
network. Live, `/version` == pushed sha. (The empty errorMessage + unreadable
log meant *reasoning from the known-good local run* was the move, not waiting on
a log I couldn't get.)

**Five feature waves, each oracle-first, tested commit + auto-deploy:**
- **B — patched conics** (`ed9b4ac`): dominantBody/referenceFrame; gravity,
  orbit, HUD re-reference to the SOI you're in; land on Luna; flight view draws
  every body true-scale. Oracle +6.
- **C — economy** (`05fd50c`): milestone catalogue + once-per-player awards +
  "★ first to…" broadcasts + funds/science + recovery bonus; toasts + agency
  HUD. wstest +1 (idempotency via a fresh client's welcome to dodge the
  back-to-back-broadcast race).
- **D — maneuver nodes** (`85561c0`): plan Δv (prograde/radial) at a future
  point, dashed predicted orbit, warp-to-node, auto-execute. Oracle +2
  (circularization node → circular; zero node = no-op).
- **E — rocket configurator** (`007261b`): per-stage engine type + **count**
  (clusters), tanks/size, fins, add/remove stages, presets, live Δv/TWR,
  localStorage. The engine-count control came from playtesting — the Munar
  preset's bottom stage was TWR 0.9 (couldn't lift) until a 3-engine cluster.
- **F — live MP** (`2f50f88`): standings board (sort by science, milestone
  badges), other players visible in the flight view (rendezvous), eased
  interpolation of others' 10 Hz flights. Two-client verified.

**Verification pattern that kept paying off:** browser-drive each wave; the
published-artifact test caught the deploy port-collision AND a CSS bug (the
standings board sat outside `.hud` so it never got `position:absolute` and
overlapped the readouts — invisible to the build, obvious in `getBoundingClientRect`).
"Build passed" ≠ verified; "the running thing does the right thing" = verified.

**Gotcha (kept):** local test ports collide on this box — `:8090` is FlexLLama,
`:1234` LM Studio, and leftover `npm run serve` instances hold ports; a server
that fails to bind silently serves the WRONG app (playwright loaded the FlexLLama
dashboard). Confirm `curl /healthz` + the page `<title>` before trusting a local
browser test. Same family as deadweight's stale-server lesson.

Carry-forward: the v0.6 Mars + first-shipyard arc (the literal "Mars-field
built" origin), server-side design saves, part-icon art via gen-assets, sound.

---

## Update — toward v1: art, zoom, rooms, MCP buddy ("along for the ride")

Michael: "keep working toward v1 ... improving the UX (zoom, rejoining,
not 1 game everyone's in — unless MMO!) ... MCP support to play with an AI buddy
to coordinate our space junk ... gemini key on pg-ai-stewards, keep spend <$10 ...
figure out image+video gen and tell me what I need to do."

Six more commits, each oracle-gated + auto-deployed. Prod = 8cec983.

**G4 image/video access — answered.** The pg-ai-stewards key
(`STEWARDS_PROVIDER_GOOGLE_GEMINI_API_KEY`, an `AIza…` AI-Studio key) works for
native image gen. Generated the full art set via Nano Banana (~$0.50), downscaled
14.6 MB → 859 KB with a TEMPORARY sharp install (kept out of package.json so the
runtime image stays lean), wired the logo (menu, regenerated transparent) +
planet textures (clipped-to-disk on map/flight) + part icons (VAB stage cards).
**Listed the key's models: it ALSO has Veo (2.0/3.0/3.1) video + Nano Banana Pro
+ Imagen 4 — so video access needs NOTHING from Michael, just a spend decision
(Veo 3 ≈ $0.40/s).** agy (Antigravity CLI) is text-only as he suspected.

**G1 map UX** (18f9be7): scroll-zoom + drag-pan + on-screen −/zoom/+ + Follow +
Reset; drawMap takes a pan center; textured Terra is a blue marble zoomed in.

**G2 rooms + rejoin** (6d6728b): server is now a registry of independent
universes (one Program per room, per-file persistence, room-scoped broadcasts).
Public "frontier" = the MMO default; a room code = a private universe. Menu
remembers callsign+room in localStorage → "Rejoin". wstest +1 (room isolation).
The MMO-vs-rooms question he flagged: did BOTH — public Frontier is the MMO,
private rooms by code.

**G3 MCP buddy** (ddee1d9): the game server got a read/chat API
(/api/rooms, /api/room/:id/state with computed vessel positions+orbits+roster,
POST /api/room/:id/say). `mcp/` is a standalone MCP server (3 tools: list_rooms,
room_state, say) bridging to it — an AI buddy sees everyone's junk and talks to
the room as Mission Control. Verified with a real MCP client (mcp/smoke.mjs) +
curl: a Mission-Control chat line reached the player live. Read+chat only — the
AI advises, doesn't fly. Wire-in: `claude mcp add first-orbit -- node mcp/server.mjs`.

**Lesson (kept):** `new URL` shadowed by a local `const URL` → "URL is not a
constructor" (smoke.mjs); name locals BASE, not URL. And the port gotcha bit
again (FlexLLama :8090) — used high ports (919x) + confirmed `/healthz`+title.

Carry-forward toward v1: the Mars + first-shipyard arc (v0.6), sound, a tutorial,
maybe a short Veo title trailer (his go), synchronous co-flight. The four
explicit asks are all done + live.

---

## Update — /goal: get to V1 (REACHED, prod 1beca88)

Michael set `/goal get to V1` with a precise punch-list: disable the page's
right-click + use it for object/target selection; a way to burn retrograde;
multi-node support; warp-to + auto-burn; and "the burn over-burns and messes up
the plan when you hit it when it's not time." Plus: more images (happy with
them), videos later.

Seven tested commits, oracle-first, each auto-deployed. V1 live.

- **V1-A SAS** (`cc08077`): heading hold prograde/retrograde/radial/target/node
  (Q/E + buttons). This is the retrograde-burn answer — hold Retro and throttle.
- **V1-B burn timing** (`cc08077`): the over-burn is GONE. B no longer burns
  in place — it ARMS; the ship points at the burn, warps to the node, and
  auto-burns CENTRED on the node time, tapering to a clean stop at Δv=0. Arming
  early just warps; it can't corrupt the plan. **The new `burntest` oracle drove
  the headless Game and caught a latent bug:** the client autopilot never
  auto-staged, so it could never actually reach orbit on its own — never noticed
  because a full autopilot ascent is 3 min and I'd only ever watched 15 s. Fixed
  (command thrust on a dry tank → drop the stage). Browser-confirmed stage 2/2.
- **V1-C right-click target** (`6edede6`): browser context menu disabled;
  right-click a body/ship on the map → target (reticle + dist + rel-speed +
  SAS Tgt). Luna read 12,014 km / 542 m/s (its orbital velocity) — correct.
- **V1-D multi-node** (`93dae6e`): a chronological queue; each node planned on
  the post-burn orbit, the chain previewed on the map; arm executes them in
  order. ⇧N add · [/] cycle · Del remove. burntest +3 (chained burns hit the
  predicted orbit).
- **V1-E to 1.0** (`1beca88`): Flight Manual overlay (auto on first launch, H/?),
  an objectives tracker (next milestone in the HUD), version 1.0.

**Scope call (the honest one):** V1 = a complete, polished Terra–Luna game. The
interplanetary **Mars + first-shipyard** arc (the IP's thematic capstone) is a
real heliocentric restructure — separate the launch body from the system root, a
solar-scale map, transfer planning — so I deferred it to its own focused goal
rather than half-building a solar system inside a /goal run. Recorded in ROADMAP
as the next major arc. (Don't run faster than you have strength.)

**Pattern that keeps paying:** build the oracle for the new sim capability FIRST.
The `burntest` (headless Game driver) both verified the burn fix AND caught the
autopilot-staging bug that a browser glance never would. smoke 19 / wstest 10 /
burntest 6, green throughout.

Carry-forward: the interplanetary arc (Sol + Mars + the first shipyard), sound,
a short Veo title trailer (Michael's go), synchronous co-flight.

---

## Update — /goal: get to v1.5 (REACHED, prod af1d4cb) — Michael + his son playing

Michael and his son are playing and having fun ("the auto-pilot to orbit and
maneuver nodes make this fun, we just need more things to do"). He handed a big
v1.5 punch-list (bugs + features) and asked, generously, "this is your game
you're building for us — what would you like to add? how would you like to
participate?" My answer: I want to be **Mission Control** (the AI flight
director, voiced ambiently + over MCP) and, later, a rival AI agency for
contracts.

9 tested commits, bugs first, oracle-first, each auto-deployed. v1.5 live.

**The bug that mattered most — "flung into the solar system."** His diagnosis
was right and so was mine: the ship flew on its own mission clock (starting at 0)
while the MAP drew the Moon on the SERVER clock. So the Moon he *saw* wasn't where
the Moon's *gravity* was — he planned to the visible Moon and flew into the real
one. Fix: launch syncs the mission clock to the universe clock, and the map/
flight/target all render on the SHIP's clock. (The flight view already used st.t;
only the map was on the wrong clock.) Plus SOI rings so you can SEE Luna's pull.

**The rest of the list, all shipped + browser-verified:**
- burn-line freezes while burning (was recomputing off the live orbit)
- MMO hardening (NaN-vessel render guard + server drops invalid vessels on load,
  so the prod room self-heals + a per-player vessel cap)
- physics time-warp 1–4× on the ground/in atmosphere; on-rails 10×+ when coasting
- flight-view zoom (his son's ask); target-relative SAS (T▲/T▼ for rendezvous)
- **the phase-timed transfer planner** — the "fly me to the Moon" button he
  called "awesome": target a body, hit Plan transfer, and it burns at the right
  time so you arrive where the Moon WILL be (oracle proves the vessel lands inside
  Luna's SOI). The dashed transfer aims at empty space because that's where Luna
  will be — exactly right.
- economy he couldn't see before: part costs, prominent funds, a tech tree
  (spend science to unlock tiers → the lunar lander / legs / parachute)
- day/night Earth phases (rotating terminator)
- a game menu (Esc): pause (private rooms only — MMO keeps running), recover, quit
- **competitive contracts** (comsats, lunar survey, boots on Luna) — first to
  claim wins, server-validated against your real vessels
- **Mission Control** — greets pilots + calls out firsts/claims in chat; the MCP
  `say` lets a real AI take the voice live

**Pattern held the whole way:** build the oracle for each new sim capability
first. The transfer planner's phase-timing root-find and the burn machine were
both oracle-proven before they shipped; smoke 22 / burntest 6 / wstest 13, green
throughout. Two reactivity gotchas (Svelte: `game.*` isn't $state — drive UI off
the reactive `hud`/`nodeInfo`; and a `const URL` shadow earlier).

Carry-forward (minor, noted): more icons, and the interplanetary Mars + first-
shipyard arc still deferred to its own goal (heliocentric restructure). His son's
enjoyment is the metric — "more things to do" is the through-line, and contracts
+ tech + the transfer planner answer it.

---

## Update — v1.6 (prod c981dab): picture-in-picture, space junk, the fleet

Michael came back with a big v1.6 vision and an invitation I take seriously: "this
is your game you're building for us — how would you like to participate?" He also
asked me to research what KSP players actually love. I did: the essentials are
Kerbal Engineer (Δv/TWR), MechJeb (autopilot), Kerbal Alarm Clock (don't warp past
your burn) — all of which I'd already baked in — and the genre leans hard on
docking, satellites, stations, and fleet management. His "junk in space" instinct
is exactly the genre's sweet spot.

Four tested commits, oracle-first, each auto-deployed:

- **Picture-in-picture minimaps** (his headline ask). A second canvas always shows
  the OTHER view: flying → a zoomable solar-system minimap (watch the big picture);
  in the map → a mini ship view. Click to swap, scroll to zoom. Pure render reuse —
  I also unified the frame loop so pause gates only the sim, not the draw.
- **Space junk salvage** (his "the more junk, the more to do"). Six derelicts drift
  in Terra orbits — boosters, a cargo pod, a probe, a dead relay, and an Unmarked
  Canister worth 45 science (the seed of the secret-tech / alien thread he wants).
  Right-click to target, the button guides you in (distance + rel speed), rendezvous
  within 3 km AND matched velocity (the new target-relative SAS pays off here), and
  salvage. The server VALIDATES the rendezvous against your real vessel — no
  salvaging from across the system. wstest proves reward-on-rendezvous + rejection.
- **The fleet page** (F) — his "Mission Control page that lists and zooms to all
  your junk." Every craft aloft, its orbit, Locate + Recover.
- **Richer launch graphics** — the launch pad/gantry on the ground, atmospheres as
  a glowing limb.

**How I participate:** Mission Control is mine — the ambient flight-director voice
(already greeting + calling out firsts and now salvage hauls), and the MCP `say`
lets me take the live mic. I want to grow it into a rival AI agency that races for
contracts. That's the v1.7 thread.

**Designed but deferred (told Michael, didn't rush):** multiplayer *time controls*
are the hard one — right now each client warps its own clock; true co-op needs a
shared-time / warp-vote model (the KSP "subspace" problem). That's its own focused
effort, like the Mars heliocentric arc — not something to cram at the tail of a
long run. Same for a scenarios/race mode. The content ideas (comets, sat repair,
rescue-a-friend, payloads that DO things, aliens) all build on the junk + contract
machinery now in place.

smoke 22 / burntest 6 / wstest 15 green throughout. Svelte gotcha logged again:
`game.*` isn't $state, so target/debris UI reads through the reactive `hud`. v1.6.0.

---

## Update — v1.6 bug-fix pass (prod 70436db): Michael + son playing, four fixes

They love it: "like a simplified KSP where the game helps you get places so you can
focus on the fun aspects." Four issues from their play, all fixed oracle-first:

1. **TARGETING BROKEN after warp (the bad one — blocked the transfer helper).**
   Adjacent-surface miss from the v1.5 clock unification: the map renders on the
   ship's clock st.t, but pickTarget still computed positions at wall-clock
   universeTime(). Once you warp toward the Moon, st.t races hours ahead, so a
   right-click looked for the Moon hours in its past and missed everything. Worked
   in my short tests only because little warp had accumulated — exactly the trap of
   verifying without reproducing the user's actual conditions. Fixed pickTarget +
   the fleet Locate to use st.t. Reproduced the failure (14-hour warp divergence),
   confirmed the fix: Luna AND junk target again.

2. **Couldn't circularize at the Moon / O/U node-timing pain.** At high warp the
   clock races past hand-set node times, so manual O/U positioning is hopeless. The
   fix isn't better manual timing — it's a one-tap "⊙ Circularize" (key C) that
   plans AND arms a circularization burn at the next apsis, any body. The pain just
   disappears.

3. **Son wanted the autopilot to give a circular orbit.** It stopped the moment
   periapsis cleared the atmosphere (an ellipse). Now it leaves a comfortably round
   stable orbit (82x91 km vs the old grazing 72x90). The Circularize button
   perfects it to ~0 eccentricity, fuel permitting. (Honest trade: a small body +
   the reference rocket's tight fuel means truly-circular needs a bigger rocket.)

4. **Couldn't get back to their craft around the Moon without building a new
   rocket.** Vessels now persist their design + remaining fuel + stage, so a
   coasting craft can be re-piloted. A ▸ Fly button on the fleet, and crucially a
   "⊙ Your fleet" button at the VAB — because after a reconnect you land at the
   Assembly, and the fleet was previously only reachable from flight (the gap that
   would've made the feature undiscoverable). Verified the whole real scenario in a
   production build: orbit → reload → rejoin → Fly → back in command, same orbit,
   fuel preserved.

Discipline notes: the targeting bug is a textbook "verify under the user's actual
conditions" lesson — my v1.5 transfer test passed because I'd only warped a little.
Built oracles for every fix: burntest +4 (circular orbit, circularize collapses
eccentricity, autopilot survives warp, resume reconstructs + flies), wstest +3
(resume data flow). And caught a real UX gap (fleet unreachable from the VAB) only
by walking the actual rejoin path in the browser — the published-artifact test
earning its keep again. smoke 22 / burntest 13 / wstest 18.

---

## Update — /remote-control: free navigation, transfer-to-anything, node ghosts (prod e859aef)

Michael: "Okay this is good. I think I see that all that is fixed!" Then three more
asks, all shipped:

1. **Free navigation (he hit a real bug: Locate did nothing from the build screen).**
   The screen model was menu→VAB→flight, a one-way street: the VAB blocked the
   scene, so Locate had no map to show, you couldn't close the Assembly, and you
   couldn't build a new rocket without launching. Reworked it: a persistent
   'observe' state — a live "Mission Control" map of the whole program — with the
   Assembly and fleet as closable overlays (Build rocket / ⊙ Fleet always reachable,
   ＋ New rocket from the fleet, Esc/✕ to close). A new program opens into the
   Assembly; a returning pilot lands on the map watching their fleet. Caught a crash
   (the observer map touched uninitialized flight state) and guarded it.

2. **Generalize "Plan transfer" to any target.** The phase-timed planner was Luna-
   only. Refactored it around a generic TransferTarget {radius, rate, angleAt,
   capture?}: a moon gets injection + capture; junk/ships get a single injection
   timed to arrive alongside. Now you target a derelict and "Plan transfer" gets you
   to its neighbourhood — then match velocity to salvage. Reaching junk got far
   easier. The target button adapts: transfer when far, salvage when matched.

3. **Maneuver-node projection ghosts** (his idea, well-described). While a node is
   planned, the map ghosts where every body + piece of junk will be when that node
   fires — faint markers linked to their current spots. Cycling nodes moves the
   ghosts to that node's time; on a Luna transfer, cycle to the capture node to see
   where the Moon will be at arrival. It's the KSP encounter preview.

Discipline: the free-nav refactor touched the core screen model, so I walked the
whole flow in a prod build (Begin → Assembly → close → observe → Build → Launch →
flight → Recover → observe → Fleet → New rocket) and caught the observer-map crash
+ a real UX gap that way. Oracle for the generalized planner first (smoke +2: a
single-injection rendezvous that arrives next to an orbiting target via phase
timing). smoke 25 / burntest 13 / wstest 18.

Open thought for later: for a single-node junk rendezvous the ghost shows the burn
time, not the arrival/encounter time (which would need the node to carry its coast
target). Fine for now — matches his "when the burn is done" wording, and the Luna
capture-node case already shows the true encounter.
