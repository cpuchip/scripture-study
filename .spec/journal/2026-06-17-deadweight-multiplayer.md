---
date: 2026-06-17
title: Deadweight Acquisitions — built competitive multiplayer overnight + deployed
lane: general-workspace
tags: [side-project, dave, multiplayer, websocket, node, dokploy, stewardship]
---

# Deadweight Acquisitions → multiplayer, live at deadweight.cpuchip.net

Michael's friend Dave (the namesake of the dave-rule) built a browser space-mining
game over a few days. Michael forked it, asked me to make it multiplayer with a
full backend, dockerize it like the other sites, and deploy it to
`deadweight.cpuchip.net` — "a silly, let's see what you can do project" — then
**councilled + ratified + handed me full stewardship and presiding authority** to
build it overnight and push to the fork / create the Dokploy project / deploy.

## Council outcome

Asked one real fork (the rest he delegated): **co-op shared corp vs competitive
corps vs co-op-first**. He picked **competitive corps — last corp standing**, the
bigger build, knowing it might not be 100% by morning. Resolved the rest myself:
Node (reuse the TS sim, his instinct), server-authoritative, one combined
container serving client + same-origin `/ws`, Dave's single-player kept fully
intact (MP is additive).

## What shipped (all verified)

A complete competitive multiplayer mode. Several corps race in one shared asteroid
field; the lowest-tonnage corp is liquidated at each quota deadline; last corp
standing wins. Live at **https://deadweight.cpuchip.net**.

- **Server** (`server/`): a multi-corp authoritative sim that *reuses Dave's
  Phaser-free modules* (`worldGenerator`, sell prices) and reimplements the ship
  loop as plain data (the originals are Phaser entities — and I needed N corps
  anyway). WebSocket rooms, host election, contested asteroid claims, rising quota
  + elimination, snapshot broadcast at 10 Hz over a 20 Hz sim.
- **Client** (`src/scenes/mp`, `src/ui/mp`, `src/net`): a Phaser renderer that
  draws the world from snapshots + Svelte lobby/HUD overlays, toggled against
  Dave's SP HUD by a mode store (no edits to his components). Deep links
  (`/?mp=1&room=code`) + copy-invite so you share a *link*, not instructions.
- **Tests**: headless sim smoke (`server/sim/smoke.ts`, 5/5) + a networked
  integration test (`server/wstest.ts`) — green local, in-container, and against
  the live `wss://` prod endpoint.

## Verification ladder (inverse-hypothesis at each rung)

1. sim smoke (headless) — loop accrues tonnage, passive corp liquidated first,
   aggressive corp wins.
2. tsc clean + vite build clean.
3. WS integration test (local tsx) — join/host/start/claim/full-sell-loop over the wire.
4. browser (playwright-cli) — lobby renders, deep-link prefills room, join→roster→
   start→match, canvas-click claims an asteroid (claim ring + ship render), asteroid
   depletes; zero console errors.
5. docker compose build-from-scratch → WS integration test against the **container** (incl. claim-contest).
6. Dokploy deploy → WS integration test against **production wss://** + prod browser screenshot.

## Deploy

NOCIX Dokploy (`server.ibeco.me`). Created project `deadweight`
(`ilaeCtLXDQrQsP9mlK9rX`) + compose `l4tkfFkX5GvAvNSlmqR3H`, wired to
`cpuchip/deadweight-acquisitions-game@master` with auto-deploy-on-push, domain
`deadweight.cpuchip.net` → `game:8080` letsencrypt. **`*.cpuchip.net` is wildcard
DNS to the VPS**, so the subdomain needed no DNS step — the one thing I'd flagged
as possibly "Michael's hand" turned out self-serve. Reused the GitHub App install
id `F3xRIOFkUcjxB6DWVkPKW` (covers all cpuchip repos).

## Presiding account

Outward-facing actions taken under the granted stewardship: pushed twice to the
fork, created a new Dokploy project, deployed a new public site. All pre-authorized
in the council. No force against any sibling process. Local MP test container
removed after prod verified; Dave's SP container on :5173 left running (Michael's
earlier session).

## Carry-forward

- **Balance tuning** is the obvious next dial — first quota can be tight with a
  distant first rock. `shared/mpConfig.ts` knobs.
- **Phase-2** (ratified *shape*, not built): hauler/miner separation, fuel/condition/
  station depth, orbiting, reconnect-hardening, lobby chat, deeper economy.
- Snapshots are full (non-depleted) asteroid sets each tick — delta-encode only if a
  room ever gets large.
- It's a "silly" project — treat phase-2 as opt-in, only if Michael + Dave want it.

## v2 same day — made it faithful to Dave's economy (Michael's first-play feedback)

Michael played it ("a tonne of fub!") and flagged three divergences from Dave's
single-player: the home base looked different, there was no base menu, and clicking
asteroids auto-claimed and auto-built with no money. All one root cause — I'd
shipped a thin auto-economy instead of Dave's money-gated one. Bar restated:
"faithful to Dave's original but with multiplayer."

Read Dave's real `Base.ts` / `BasePanel.svelte` / `SpaceScene` (not just the
architecture doc this time) and rebuilt the MP economy to match:
- Start **1 hauler, 0 miners, 750cr** (SP's single Hauler-01). A hauler mines only
  with a **purchased AutoMiner (300cr)** — clicking an asteroid with no miner
  dispatches nothing (the money gate; issue #3).
- Ore hauls to **base storage (cap 2000)**, **sold manually** at a new **base menu**
  (`MpBasePanel`, mirrors Dave's `BasePanel`: MARKET/SHIPYARD/EQUIPMENT) that opens
  by clicking your base (issue #2). **Faithful station base render** + name label in
  GEO orbit around a proper planet (issue #1).
- **Tonnage = tons DELIVERED**; period 1 is a humane setup window (150s/50t) because
  the faithful economy ramps slower than the old free-mine slice — the first-play
  self-liquidation surfaced exactly this tuning need.

Verified: sim smoke (money gate asserted) + networked integration (no-miner→no-
dispatch, buy→dispatch→deliver) green local + **prod wss://** + in-browser (base menu
opens, buying a miner spends 750→450). Faithful economy LIVE.

**★ Deploy lesson — auto-deploy-on-push is NOT firing for this repo.** `triggerType:
push` is configured but the GitHub App webhook never fired; ALL deploys so far were
manual `compose.deploy`. The earlier journal/memory claim of "auto-deploy on push"
was an assumption, now corrected (Ben test). Workaround: after a push, manually POST
`compose.deploy {composeId}` (NOCIX key) + verify with the prod wstest — and do NOT
trust the deployment `status:done` at t+0 (stale; the prod integration test is the
ground truth, as the Dokploy stale-build memory warns). Webhook fix is a future task.

## v2.1 same day — faithfulness fixes + Pause/Quit + the parity roadmap

Michael played v2 ("much more faithful!") and flagged: right-click-drag pops the
browser context menu (off in Dave's), we still auto-claim on click (he LIKES it, but
Dave doesn't), Dave starts with 1 ship (we match now), and Dave has more features +
a fuller base. He asked me to **play Dave's SP and catalog the gaps**, add **Pause +
Quit** to the MP top bar, and **plan the v3+ parity phases**.

Played Dave's SP (the `dwa-game` container, :5173) + read his `EntityPanel` /
`BasePanel` / `SpaceScene`. Confirmed the real designation UX (**click selects →
"Designate for Mining" button**, not auto-claim) and the full deferred sim
(attachment slots, fuel/RCS/battery meters, miner condition/repair, nets, beacons,
**minimap**, big textured planet, fuller base with station services/fees/upgrades).

Shipped + prod-deployed (manual `compose.deploy`, verified prod wstest):
- **Right-click fix** — `disableContextMenu()`; left-click selects, right/middle
  pan-only (tracked `dragButton`). Verified via dispatched `contextmenu` →
  `defaultPrevented=true`.
- **Pause (host) + Quit** in the top bar — `world.paused` freezes tick + clock;
  `forfeit(corpId)` removes a quitter from the race; host passes on quit; room GCs
  when empty (dies on the server). 13/13 integration assertions green incl. "paused
  freezes the sim clock" + "quitting forfeits the corp."
- **`ROADMAP.md`** — phased v3–v7 toward full SP parity (v3 interaction parity:
  designation-UX/minimap/EntityPanel/planet · v4 fuller base + fleet identity · v5
  deep mining loop · v6 fuel/condition · v7 world dynamics + webhook fix). **Open for
  ratification:** auto-claim default (keep his liked feature as a toggle?) and how
  deep to go — the honest tension is that Dave's depth is a solo optimization puzzle
  and a friends' race may be MORE fun streamlined (v3–v4) than fully simulated (v5+).

**Ratified (AskUserQuestion):** **FULL PARITY (v3→v6)** — Michael wants Dave's whole
game faithfully, accepting the added micro. **Designation default = Dave's
select→designate**, auto-claim demoted to an opt-in "quick-claim" toggle (default off).
**v3 started + deployed same turn:** left-click now SELECTS an asteroid (panel shows
"Designate for Mining"); the button dispatches; quick-claim toggle restores
click-to-claim; moved the HUD action buttons top-left so the base panel stops covering
them (caught in-browser — the base panel overlapped the actions, Playwright click hit
the panel). Verified in-browser (select→designate→claim; toggle flips). Rest of v3
queued: minimap, EntityPanel build-out, bigger textured planet. `ROADMAP.md` is the
living plan. Each phase: build → smoke+wstest → in-browser → manual `compose.deploy`.

## "I want it all — goal: reach v7" + v3 COMPLETE + auto-deploy fixed

Michael set the **goal to drive the whole roadmap to v7 (full parity)**, autonomous
(Ammon), surface only on input-needed. And he **fixed the auto-deploy webhook** (granted
the GitHub App access to the deadweight repo). **Confirmed: auto-deploy now fires on push**
(commit `68550b9` deployed with no manual `compose.deploy`) — the manual step is retired.

**v3 COMPLETE + live:** (a) designation = Dave's select→designate default + quick-claim
toggle [prior]; (b) **minimap** — first built in-scene (Phaser Graphics + scrollFactor 0)
but the **camera ZOOM still scales screen-fixed Graphics** → it rendered small + offset;
rebuilt as a **DOM `<canvas>` Svelte component** (`MpMinimap.svelte`), camera-independent,
top-right, scoreboard moved below it; (c) **ship-select EntityPanel** (click a ship →
owner/state/cargo/miner; base/ship/asteroid selection mutually exclusive); (d) **textured
planet** (banding, bigger). Lesson logged: **for screen-fixed UI in a zoomed Phaser scene,
use a DOM overlay, not in-scene Graphics** (scrollFactor kills scroll, not zoom). Verified
in-browser local + prod. **v4 next** (fuller base + named ships + miner slots).

## v4 COMPLETE (auto-deployed) — fleet/economy depth

Michael: "work on v4!" (autonomous run continues). Shipped + auto-deployed:
- **Named ships** — `Hauler-01`, `-02`, … (per-corp counter); shown in the ship panel.
- **Per-ship cargo upgrades** — tiers 200/350/550/800, costs 300/600/1000 (Dave's
  numbers); upgrade button in the ship detail panel; per-ship `cargoLevel` drives
  the mining cap.
- **Auto-designate** toggle (base panel AUTOMATION) — when on, idle miner-haulers
  auto-claim the richest unclaimed asteroid (`claims < minerHaulers` gate). Browser-
  proven: a corp with it ON delivered 200t with zero manual clicks.
- **Re-sequencing call (within stewardship):** moved multi-miner-per-hauler
  (attachment slots) from v4 → v5 — it's meaningless until miners are separately
  deployable (v5's deep-mining loop). Same v7 destination, cleaner phase boundary.

Tests extended (naming, upgrade, auto-designate) — smoke + wstest green local + prod
(16 assertions). **Auto-deploy fired on push again** (no manual trigger). `ROADMAP.md`
updated. **v5 is the big one** (deep mining loop) — likely a focused multi-step turn,
and the natural point to invite a playtest/balance check (the deep sim changes the feel).

## v5a COMPLETE — THE DEEP MINING LOOP (the biggest single change)

Michael played v4 ("a lot smoother and more faithful!") and said "move on to v5."
v5a inverts the core mechanic to match Dave's single-player:

- **Miners are now a purchased POOL** (no longer "mounted on a hauler" — that was the
  v2-v4 simplification). buyMiner just adds to `minersOwned`.
- A hauler **carries a miner out and DEPLOYS it** at the claimed asteroid (a `SimMiner`
  entity stays there). The deployed miner **mines + ejects nets** into a buffer
  (`oreReady`, capped at `MINER_NET_BUFFER * NET_CAPACITY`), going **net-starved** when
  the buffer fills. The hauler **shuttles** — waits for a net-batch, grabs up to its
  cargo capacity, hauls to base storage, returns; the miner buffers more during the
  round trip (so it's a real back-and-forth, not a parked drip — that refinement
  mattered). The miner is **recovered** when the rock depletes.
- `server/sim/world.ts` rewritten around `SimMiner` + a deploy/collect/shuttle hauler
  state machine. Protocol: `MinerSnap`, new ship phases (deploying/collecting/unloading),
  `carryingMiner`. Client renders deployed miners (squares) + tethered nets (amber dots)
  at asteroids + hauler carry markers; base/ship panels moved to the miner-pool model.

Verified: smoke (deploy→net→shuttle→tonnage) + wstest green local + **prod** (Alpha
banked 52t over the wire); browser e2e — bought a miner, auto-designate deployed it,
**delivered 53t** via the full loop. Auto-deployed on push.

**Playtest learnings worth keeping:** the deep loop is slower per-cycle (deploy dwell +
travel + collect + travel) so first-delivery is ~15-25s — the period-1 setup window
(150s/50t) absorbs it. Playwright on this canvas+overlay app keeps biting (grep matched
the *hint text* "Buy an AutoMiner" instead of the Buy *button*; and one run timed out
past the 150s deadline so buys on a liquidated corp were silent no-ops) — list buttons
by role/name, act fast, don't trust ref stability across re-renders.

**v5b next:** orphaned-net recovery (designate-for-collection), beacons for net-starved
miners, multiple miners per hauler (attachment slots), miner/net detail panels.

---

## v5b-1 COMPLETE — read the loop (miner panel · beacons · net-starved alert)

Michael's reaction after playing v5a — "I see the little miners on the ships and how
they stay and mine" — pointed straight at the highest-value v5b work: make those
deployed miners **clickable and legible**, and make a stuck (net-starved) miner
*announce itself*. Split v5b into two verifiable pushes; this is the first.

- **Miner detail panel.** Click a deployed miner — it takes click priority over the
  asteroid it sits on (`nearestMiner` checked before ship/asteroid) — to see resource /
  owning corp / state / nets ready / host-rock remaining, with a **RECALL MINER** button
  (your corp only). Recall reuses the existing `undesignate` command: the bought miner
  returns to inventory (you keep what you paid for), the claim frees, the hauler is
  released. No new wire command needed — the panel is a near-clone of the working ship
  panel, so it lives in `MpHud.svelte` alongside it.
- **Beacons.** A net-starved (full) or depleted miner throbs a pulsing ring in-scene
  (amber for full, grey for depleted); net-starved miners also blip amber on the DOM
  minimap so you can spot a stuck one field-wide. Pure render off the snapshot.
- **Net-starved alert.** The one server-side change: `updateMiner` now takes its corp so
  it can push a "⚠ <Corp>'s miner is full of nets — send a hauler" line **once**, on the
  mining→net-starved transition. The fleet readout shows "⚠ N full" too. This is the
  legibility fix for the deep loop's backpressure — before, a full miner was silent.

Server change is minimal + behaviour-equivalent (just the log line + a refactor of
`updateMiner` into one state assignment). Everything else is additive client:
`mpSelectedMiner` store, `nearestMiner` pick, the render beacon/selection, the panel.

**Verified at every rung:** typecheck + `vite build` clean; **smoke 21/21** — added a
deterministic net-starved scenario (designate the *farthest large rock* so the single
hauler's round trip can't keep the buffer drained → the miner must net-starve, beacon
trips, log gets "full of nets"), plus recall asserts (miner removed from rock, claim
freed, owned-miner count preserved); **wstest 18/18** local; **prod**: pushed `5f180e2`,
auto-deploy fired, the prod bundle is byte-identical to the local build and **contains
`RECALL MINER`** (content-proof the new client is live, not just `status:done`), and
prod wstest banked 53t over the wire; **browser e2e**: bought a miner, auto-designate
deployed it, the full deploy→deliver loop ran (53t) with the new per-frame render code
(beacon pulse + selection ring) active and **0 console errors/warnings**.

Honest verification note: the one thing I did NOT live-click is the miner panel opening
from a real canvas click — the deployed miner's screen position needs the Phaser camera,
and `window.Phaser` isn't exposed in the bundled build (and joining a running match to
read coords is correctly rejected by the server). Given the documented Playwright
fragility on this canvas app, I verified the panel by typecheck + build + parity with the
working ship panel + a console-clean render pass, rather than a fragile pixel-hunt. The
render *path* (which draws beacons every frame) is live-proven console-clean.

**v5b-2 next (the logistics depth):** multiple miners per hauler (a "milk run" — carry
several miners out, deploy across a cluster of claimed rocks, shuttle from the cluster)
+ orphaned-net recovery (nets from a recalled/depleted miner become collectible salvage
rather than vanishing). Then v6 (fuel/condition/station services), v7 (world dynamics).

---

## /goal: drive to v7 deployed (autonomous run, Stop-hook active)

Michael set a /goal: "v7 is deployed — take each phase as a git commit + push + tested,
break up any v# phase into smaller phases as needed, stop only if you need my direct
input; the intent is a faithful MP version of Dave's single-player Deadweight, with MP
features + some ease-of-life features too. Also the SP minimap is clickable (navigates
the main window) — add that." A session Stop-hook now blocks stopping until v7 is
deployed. So this is an Ammon run: ship faithful, tested increments toward parity.

Read Dave's SpaceScene to stay faithful (not invent). His SP carries: clickable minimap,
Keplerian orbiting (orbitalRadius/Angle on every asteroid, ω=ORBITAL_K/r^1.5), company
asteroid arrivals (generateCompanyAsteroid + COMPANY_ARRIVAL_* pacing), hauler fuel +
battery, miner battery + condition/repair (+ 'station-repair'), free-orbit nets/miners,
standby-beaconing, station services (docks/hangars/fees), waitOrbital parking. The
worldGenerator already hands us orbits + company-arrival spawning — so those two were
cheap and faithful.

**Shipped + prod-verified this run (each a tested commit + push):**
- **clickable minimap** (`e333810`) — click the MP minimap → camera flies there (300ms
  pan), matching SP. mpCameraTarget store bridges the DOM minimap → Phaser scene.
  Screenshot-confirmed (clicking centre flew to the planet at world origin).
- **orphaned-net recovery** (`daeb562`, v5b-2a) — recall-with-nets leaves the nets adrift
  as salvage (OrphanNetSnap + a 'to-orphan' ship phase); a freed hauler auto-recovers
  them. Faithful to Dave's free-orbit nets. smoke 25/25.
- **company asteroid arrivals** (`d210066`) — new company rocks arrive over time, interval
  scaling BASE→MIN by remaining-natural fraction, capped at COMPANY_ASTEROID_MAX_COUNT;
  gold halo in-scene + gold minimap dot; AsteroidSnap.isCompany; companyArrivalsCount.
  smoke 28/28.
- **Keplerian orbiting** (`9fb0977`) — the field drifts; deployed miners + docked haulers
  ride their rock's orbit. smoke 30/30; wstest gained an over-the-wire orbit-drift assert
  (the prod ground-truth, since orbiting is server-only → no client bundle hash to diff).

**Debugging lesson (kept):** the orbit-over-the-wire assert first FAILED with 0.00 drift
while smoke (same world.ts) clearly orbited. Inverse-hypothesis: smoke works → suspect
the harness/env, not the sim. A **stale pre-orbiting `npm run serve` was still bound to
:8080** (PID found via `netstat -ano | grep :8080`); my "fresh" servers hit EADDRINUSE
and died (logged), while the old one answered healthz + ws. Killed it (`taskkill //F
//PID`), reverified — orbit drift 146→207 units, green. Gotcha for the background-server
pattern: a silent EADDRINUSE means you're testing the OLD process. Verify the listener
PID (or free :8080) before trusting a local result; the deterministic smoke is the oracle
that says "the sim is fine, look at the environment."

**Prod-verify pattern by change type:** client-touching phases → poll until the prod
bundle hash matches the local build AND content-grep a new feature string; server-only
phases (orbiting) → poll the prod wstest until its behavioural assert (orbit drift) passes
(no bundle hash changes).

**Remaining to v7 (ROADMAP `Remaining`):** multi-miner-per-hauler (milk-run) · v6
fuel/battery/condition+repair · v6 station services · v7 room-persistence + lobby-chat +
spectator polish. Plan: implement faithfully using Dave's own constants, but auto-manage
the resource systems (auto-refuel/auto-service) so the auto-dispatched competitive race
stays ease-of-life rather than micro-heavy — the synthesis the goal points to.

---

## Full-parity decision + the resource systems (fuel, condition)

Before grinding the remaining items I surfaced a genuine tension via AskUserQuestion: the
leftover parity items (fuel, battery, condition/repair, station services, the 2-miner bay)
are all SP *manual-management* systems — in the auto-dispatched MP they'd be auto-managed,
so they go mostly invisible. Michael chose **"Full faithful parity — all of it"** (auto-
managed). So: implement every system with Dave's constants, auto-serviced, MP-safe, and
make them VISIBLE (bars + credit-sink fees) even where they're gentle.

**Build stamp first (Michael's mid-run ask, `1a76ba2`):** `__BUILD_SHA__` on the menu +
a corner badge, `GET /version`, `dist/version.txt`. The Dockerfile installs git and
un-ignores `.git` to read the commit (then `rm -rf .git`). **This changed my verify loop:**
deploy-verify is now `curl /version` until it equals the pushed short-hash — definitive for
server-only changes (no client bundle hash to diff) and it would have instantly caught the
stale-server detour. (Tradeoff: the git-enabled image builds a bit slower; a deploy now
takes ~2 min and shows a brief Bad Gateway during the container swap — expected.)

**(a) hauler fuel + battery (`a495a2f`):** fuel (max 300, drain 3/s) burns while thrusting,
tops off at base every return for a fee that scales with distance burned — so far rocks
cost more to service, and it can NEVER strand (movement doesn't hard-stop on empty;
refuel is automatic each base visit). Battery recharges parked. The key design choice that
makes this faithful-but-safe: fuel is a *visible bar + a distance-priced credit sink*, not
a routing constraint.

**(b) miner condition + battery + repair (`2e82f13`):** the design subtlety — in an
auto-managed loop the shuttling hauler services miners so often that a wear stat would pin
at full (invisible). Fix: the miner *ages on-station* (wears whether mining or net-starved),
so a FAR rock (long round trip → services spaced out) dips below grace between visits =
real mining penalty + real repair bills, while a CLOSE rock stays topped. That asymmetry is
what gives condition teeth. The hauler auto-services on each collect (recharge free, repair-
for-a-fee when worn). Determinism kept by steady wear (no Math.random failures), so the
seeded smoke stays deterministic; the main match still resolves identically.

**Pattern that recurred:** an auto-managed resource stat is invisible unless its wear
outpaces its service cadence — so tie the wear to something the auto-loop can't fully
hide (deployment time / travel distance), not to the serviced action itself.

Shipped this run so far (9 phases, all prod-verified via `/version` or bundle-content):
clickable-minimap · orphaned-net-recovery · company-arrivals · orbiting · chat · build-stamp
· hauler-fuel · miner-condition. Remaining: station-services (light, next) · 2-miner bay ·
room-persistence · spectator/camera polish + balance → v7. smoke 35/35, wstest 21/21.

---

## ★★ v7 DEPLOYED — full parity reached (13-phase autonomous run)

`curl https://deadweight.cpuchip.net/version` → `638d953`. The /goal Stop-hook condition
("v7 is deployed") is met. Over this run, each phase a tested commit → push → prod-verified:

1. v5b-1 read-the-loop (miner panel + RECALL + beacons + net-starved alert) `5f180e2`
2. clickable minimap nav (Michael's ask; fly the camera there) `e333810`
3. orphaned-net recovery (recall leaves drifting salvage a hauler recovers) `daeb562`
4. company asteroid arrivals (new ore over time, faster as the field empties) `d210066`
5. Keplerian orbiting (the whole field drifts; miners ride their rock) `9fb0977`
6. room chat (lobby + match) `03a839d`
7. build stamp (`__BUILD_SHA__` on screen + `/version`) `1a76ba2`
8. hauler fuel + battery (auto-refuel for a distance-scaled fee, never strands) `a495a2f`
9. miner condition + battery + repair (wears on-station, hauler auto-services) `2e82f13`
10. station-service fees surfaced in the base panel `7f2c294`
11. 2-miner bay milk-run (one hauler deploys 2 miners per trip) `0a73848`
12. room persistence (matches survive a restart; reconnect-by-name) `dc36ec4`
13. spectator + camera polish (F frame / C center / auto-frame on elimination) `638d953`

**The pivotal moment** was the AskUserQuestion mid-run: I'd realized through deep reading
that the remaining v6 systems (fuel/condition/station-services/2-miner-bay) are SP
*manual-management* systems that go mostly invisible in the auto-dispatched MP. Surfacing
that tension (covenant `surface_tensions` + the goal's "stop if you need input") let Michael
choose **"Full faithful parity — all of it"** rather than me silently grinding or silently
skipping. That one question shaped the back half of the run.

**Reusable lessons (also folded to `project_deadweight_game`):**
- A build-hash on the page + a `/version` endpoint turned deploy-verification from
  bundle-hash-guessing into one definitive curl — essential for the server-only phases.
- An auto-managed resource stat is invisible unless its wear outpaces its service cadence;
  tie wear to deployment-time/distance (what the auto-loop can't hide), not the serviced
  action itself.
- The deterministic smoke is the oracle: when orbiting "failed" over the wire, smoke said
  the sim was fine → the bug was a stale `:8080` server (EADDRINUSE silent-fail). Free the
  port before trusting a local result.
- Persistence forced a GC rethink: a *running* room must be kept when it empties (so it can
  persist + be reconnected); only lobby/ended rooms GC on empty, with a TTL sweep for the
  truly abandoned.

What's NOT done (honest): full faithfulness of every SP micro-detail (pressurization,
dock-vs-hangar service speeds) doesn't map to MP's one-base-per-corp and was folded into
the fee model; balance is "smoke resolves to a clean winner, fees small vs ore" — a real
multiplayer playtest with friends is the next tuning input. Memory: `project_deadweight_game`.

---

## Post-ship bug: lobby froze on connect (MpChat reactive loop) — FIXED (`adfbcc0`)

Michael hit it immediately: clicking JOIN locked up the browser, couldn't create a room.
A real regression from the chat phase, and the smoke/wstest oracles couldn't catch it
(it's a client UI freeze, not sim/wire). Debugged properly instead of guessing:

- **Symptom triage:** an `eval("1+1")` in the page hung → the *main thread is blocked* =
  a synchronous infinite loop, client-side. Reproduced locally with a clean state (not a
  prod/persistence thing).
- **No console output** → the loop blocks before logging → a Svelte reactive cycle in the
  *production* build (dev mode has a "too many updates" guard; prod doesn't → silent hang).
- **Console-marker bisect:** logged each WS message type; only `welcome` printed, never
  `snapshot` → the freeze is in the synchronous flush from `mpConnection.set('connected')`.
- **Component bisect:** the two things that newly render on 'connected' are the lobby roster
  and MpChat (mounts for the first time). Disabling MpChat cleared the freeze → MpChat.

**Root cause:** MpChat's auto-scroll was a reactive block
`$: if ($mpChat && list) { tick().then(() => list.scrollTop = ...) }`. A `$:` that reads a
`bind:this` ref (`list`) *and* calls `tick()` flushes recursively and never settles. **Fix:**
do the auto-scroll from a store subscription + `requestAnimationFrame` (no Svelte-flush
interaction) instead of a reactive `$:`. Verified the full flow on live prod: join → lobby →
start → in-match → send a chat line, main thread responsive, 0 console errors.

**Lessons (→ memory):** (1) a `$:`-block that reads a bind:this ref + calls tick() is an
infinite-loop trap that's *invisible in prod* — never use `$:` for post-render DOM work;
use a subscription + rAF/afterUpdate. (2) For a "browser locks up" report: `eval` to prove
main-thread block, then console-marker bisect to find the freezing step — far faster than
reading code. (3) The deterministic smoke/wstest oracles don't cover client UI; a fast
browser-load + eval-alive check belongs in the per-phase verification for UI-touching work.

---

## Playtest round 2: SP-comparison fixes (starter miner, ship, stars, smoothing, chat)

Michael played the now-working lobby and gave sharp feedback, and asked for a
screenshot+code comparison to Dave's SP. Did exactly that — ran both (NEW GAME vs
MULTIPLAYER), screenshotted, read the SP entities. Findings → fixes (commit on its own):

- **Start 1 hauler + 1 miner.** Dave's `spawnStarterShip()` literally "Pre-loads one
  AutoMiner" on the hauler; our MP started with 0. So `STARTING_MINERS 0→1` — the money
  gate now applies to *additional* miners (the starter deploys the moment you claim a
  rock). This also makes the v2 money-gate MORE faithful, not less (SP gives a free
  starter miner). Updated the smoke/wstest money-gate assertions accordingly.
- **Ship invisible at start.** Root cause: an idle hauler sits dead-centre on the base
  disc (same position, similar colour) → it vanishes into the base, only becoming
  visible once it leaves to deploy a miner (Michael's exact observation). Fix: park idle
  haulers just *outside* the base ring (offset by index) + bigger hull (len 9→15) + a
  white outline. Now visible from frame one.
- **Jitter.** SP runs its sim at 60fps locally (smooth); our MP rendered raw 10Hz
  snapshot positions every frame (steppy). Fix: `SNAPSHOT_HZ 10→20` + client-side
  easing — the scene keeps a per-id display position and eases it toward the latest
  snapshot each frame (asteroids/ships/miners/orphans), snapping on big jumps
  (deploy/respawn). 20Hz + interpolation ≈ continuous motion.
- **Starfield.** SP has world-space stars (depth + motion reference); added 520.
- **Chat moved.** It was bottom-left, directly over the base menu — moved to the right
  edge below the standings, collapsed by default.

Compared, the remaining SP-nicer bit is the planet (smooth radial gradient + atmosphere
ring vs our flat ellipse bands) — noted, lower priority. Verified: smoke 45/45 + wstest
21/21 + browser (screenshot-confirmed ship visible, stars, chat relocated, 1 miner, 0
console errors). Comparison screenshots sent to Michael.

---

## Playtest round 3: interaction + station faithfulness (+ what's still queued)

More SP-comparison feedback. Read Dave's `attachmentTypes`/`hangarBays`/`Ship`/`Base`:
SP ship loadout = 2×S + 2×M (slot 1 = net-store, the M slots hold miners); station =
6 docks + 3 hangars (radius 110) + 6 miner slots + a pressurized bay; ships return to
the first unoccupied dock.

**Shipped (`...` round 3, client-only):**
- **Left-click is select-only** — it no longer pans (the drag-then-it's-a-pan logic made
  small moving targets unselectable). Pan is right/middle only; any left-up selects.
- **Bigger ship + hitboxes**, and click-detection now reads each entity's DISPLAYED
  position (the `disp` eased/docked map) — so you click where you SEE it. This also fixed
  a mismatch I'd introduced: idle ships are *drawn* parked at a dock offset but were being
  *picked* at their raw base-centre position.
- **Station look:** 6 docks (close, blue) + 3 hangars (radius 110, amber) per base; idle
  haulers park at the docks instead of vanishing into the base disc.
- **SP Main Menu button** next to Save (Michael asked; sanctioned touch of Dave's SP HUD).

**Queued for a focused next pass (told Michael):**
- **2a — the per-ship cargo-upgrade button is our invention** (SP has none; cargo = the
  net-store slot's fixed capacity). Offer to remove for faithfulness vs keep as MP QoL.
- **2b — ship attachment bars:** render the 4 slots below the ship (net-store fill +
  miner slots) as little vertical progress bars like SP. Needs `minersAboard` on the wire.
- **Planet:** SP's smooth radial gradient + atmosphere ring vs our flat ellipse bands.
- (minor) the SP HUD leaks onto the title screen — pre-existing; a scene-aware show/hide.

---

## /goal round 4: the 4 queued SP-faithfulness enhancements — all shipped

Michael ratified "build all 4" (a Stop-hook /goal). Done, one commit, prod-verified:

- **2b ship attachment bars** — render the ship's 4 attachment points below the hull
  (faithful to SP's 2×S + 2×M): a net-store slot (vertical fill = nets aboard) + a spare
  + 2 miner bays (solid when loaded). Added `minersAboard` to ShipSnap for it; ship panel
  shows "miner bays N/2 loaded".
- **2a removed the cargo-upgrade button** — it was our invention; SP has no cargo upgrade
  (cargo = the net-store's fixed capacity). Removed the button + the `upgradeShip` command
  + handler + the smoke/wstest asserts. cargoLevel stays pinned at tier 0.
- **3 planet** — flat ellipse bands → a shaded sphere: a radial gradient lit from the
  upper-left (concentric circles whose centre drifts toward the light = terminator) + a
  soft atmosphere halo. Helper `lerpColor`. Reads as a sphere now (screenshot to Michael).
- **4 SP HUD gate** — the SP HUD leaked onto the title screen (mpMode 'off' covered both
  menu AND game). New `spaceActive` store (set true in SpaceScene.create, false in
  MainMenuScene.create); main.ts shows `#hud` only when off-mp AND spaceActive. Menu clean.

Verified: smoke 44/44 + wstest 21/21 (miner-bay count over the wire; cargo-upgrade gone) +
typecheck + build + browser screenshots (sphere planet, clean title, attachment slots, 0
console errors). The attachment bars are small at default zoom — fine, they read when
zoomed. Remaining nicety if Michael wants: rotate attachments into the ship's frame.

---

## SP audit + the station economy (closing the last faithfulness gap)

Michael asked me to (a) lock the ship attachment bars to the hull's frame and (b) audit
all of Dave's SP for missing features. Did both.

**Attachment rotation (`9402b63`):** the 4 slots now lay out in the ship's frame (fwd +
perpendicular) and draw as rotated quads (fill/strokePoints) so they turn with the hull.

**The audit** (read EntityPanel + BasePanel + entities/state). Verdict: core loop, economy,
world dynamics, station visuals all match. Findings split three ways:
- **★ A correction:** SP DOES have the cargo upgrade — in the *base panel* (ship-selected),
  not the ship menu. So my 2a removal was wrong (Michael's observation that the *ship menu*
  has no upgrade was right; the upgrade lives in the base panel). Re-added it.
- **Real gaps:** the **station economy** — buyable miner-slots (a cap on owned miners),
  owned docks/hangars, pressurization; usage display; fuller fee schedule. We drew the
  docks/hangars but they weren't buyable/meaningful.
- **Intentional simplifications (not gaps, per "auto-managed"):** RCS fuel, manual net
  objects + Designate-for-Collection, manual miner states (dark/station-repair/resupply)
  + ship→miner charge toggle, catastrophic failure. SP manual micro we auto-handle.

**Built the station economy (3 phases, one commit):**
1. Cargo upgrade re-added (base panel, ship-selected) — the correction.
2. **Miner-slot cap** — own ≤ minerSlots miners (start 3, cap 6); buy slots to scale. The
   core SP station mechanic; makes fleet size a real investment. (Surfaced a real AI
   interaction: the smoke AI got stuck wanting a 4th miner it couldn't slot → taught it to
   buy a slot first; without that the cap let a less-capped corp win the deterministic
   match — good signal the cap actually bites.)
3. **Station service upgrades** — owned Docks cut refuel fee (6=free), Hangars cut repair,
   Pressurization halves repair (needs a hangar). Credit-efficiency progression. + usage
   + fee schedule in the base panel (scrolls).

Verified: smoke (cap/slot/dock/hangar/pressurization-gate/cargo-upgrade + main match still
resolves) + wstest (buy a slot over the wire) + browser (full STATION panel renders). The
MP is now a close faithful match to Dave's SP — the remaining deltas are the deliberately
auto-managed manual-micro systems. Audit + the comparison shipped to Michael.

---

## UI polish round (SP interface match) + Plan B planned

Michael played, picked **Plan B** for the next features (contested salvage + catastrophic-
failure stakes — MP-native depth, NOT the SP manual-micro) — recorded in ROADMAP, not built
yet. Then flagged interface gaps where SP reads better. Read SP's exact positioning code and
matched it:

- **Ship parks BESIDE the rock, not on top.** SP: `SHIP_PARK_RADIUS=35`, `SHIP_PARK_ORBIT_RATE
  =0.4` — the hauler orbits the rock it services. Ours sat dead-centre. Added a parkAtRock
  orbit in the deploying/collecting phases (parkAngle on the ship). Position-only, so the
  timer-based collect/deploy logic is unchanged (smoke/wstest still green).
- **Miner above the rock + nets ring the rock.** SP: miner at `asteroid.y - 20`, nets at
  `asteroid ± 18`. Ours had the miner on top and nets ringing the miner. Now miner at
  `a.y - MINER_PARK_OFFSET (22)`, nets ring the ASTEROID at `NET_RING_RADIUS (18)` (client
  builds an asteroid-position map in the render to place them).
- **Transfer progress** — `ShipSnap.progress` (from the phase timer) → a progress ring around
  a hauler while it deploys/collects/unloads. SP shows these; we didn't.
- **Attachment slots spaced** (were cramped) + the **ship menu** now lists the loadout
  (net-store fill + 2 miner bays loaded/empty + a FUEL/POWER section), like SP's EntityPanel.

Verified by SP-matched constants + smoke + wstest + build + 0 console errors. Honest gap in
verification: couldn't auto-capture a *zoomed* live mining shot — canvas-clicks need a rock
exactly at the pixel, and a probe can't join a running match (the server correctly rejects
mid-match joins). So the on-station look is verified by code-match, not a screenshot;
Michael confirms it live. If the feel's off, iterate.

**Plan B (next, in ROADMAP):** contested salvage (orphan nets grabbable by any corp = raid
a rival who over-extends) + catastrophic failure (un-serviced miners can fail = punish
over-claiming). NOT porting: RCS, nets-as-objects, manual miner servicing, charge toggle.
