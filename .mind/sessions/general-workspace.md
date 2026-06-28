---
lane: general-workspace
session_id: c4fef1d0-292c-4ad6-b6c5-76e2af1043c3
status: active
started: 2026-06-11T12:00:00
last_active: 2026-06-28T06:10:00
---

> Note: this lane was previously filed under the typo `general-workspase.md`;
> consolidated here (the hook-canonical spelling) 2026-06-13, typo file removed.

## Working on
- **★ yt-MCP slide enhancement SHIPPED + Agentic-OS review 2026-06-28 (Part A done).** Built
  `yt_download_video` + `yt_frames` (scene/interval/timestamps + timestamp-aligned `frames.json`) +
  `yt_slides` (one-shot: chapters→scene→interval + narration-aligned `slides.md`) in `scripts/yt-mcp/`
  (`frames.go`/`slides.go`; commits `93b734ef` + `4c3f6e7f`, not pushed). Spec both halves
  `.spec/proposals/yt-slide-frames.md` — **Part B = pg-ai-stewards' substrate digester** (delegated,
  PR pending). ★ ffmpeg-location override fixes a stale yt-dlp config. Live-verified scene+interval+
  chapters. **Tested on the Cole Medin "New SDLC" video → enriched the harness study** (whitepaper Fig 5
  NAMES "Trajectory Eval" = validates the substrate trajectory-critic; the 52.8→66.5 / +13.7
  Terminal-Bench numbers the transcript flattens). **Reviewed Chase AI "Agentic OS" video vs
  pg-ai-stewards** (`study/yt/agentic-os-10x-claude-code-chase-ai.md`): his AIOS = our vision one tier
  down (file-system-backed; DB/multi-agent/verify/governance are hand-waves); steal = **session-mining**
  + **cheap index-map tier**; dogfood note left for pg-ai-stewards (substrate reviews it after Part B).
  **★ yt-dlp staleness lesson:** the "n challenge solving failed" video-download wall was a stale yt-dlp
  (2026.03.13) — `pip install -U` → 2026.06.09 fixed it cleanly (NO deno); Part B bridge must pin a
  recent yt-dlp + rebuild periodically. **★ Claude Code has NO MCP hot-reload** — rebuilding a local MCP
  binary needs a full restart; rename-swap the locked `.exe`; memory `reference_claude_code_mcp_no_hot_reload`.
  Read-before-quoting caught a subagent's cleaned 'verbatim' quotes → paraphrased the review. `yt_slides`
  goes live on next CC restart. Inbox: comments-integration DEFERRED (security-gated — prompt-injection
  is load-bearing); email/SMS parked. Journal `2026-06-28-yt-slide-enhancement-and-agentic-os-review.md`.
- **★ Garrison — pg-ai-stewards local-model learnings APPLIED 2026-06-22** (acted on
  the doc-construction inbox signal). Mapped the substrate's 5 soak learnings against
  Garrison's real loop code: borrowed the MoE the rig now serves (`qwen3.6-35b-a3b` —
  fixed a live `ping` 404 + ~4× faster), killed the one-shot whole-file trap in the
  loop prompts (lean on the already-built `===EDIT===` diffs), journal-as-output
  (`===JOURNAL===`/`ParseJournal`, **proven e2e on the live MoE**), honest per-slot
  context gauge (192k→120k) + README rig docs. Critic-reads-disk was already satisfied.
  Source-page-in surfaced (not built blind — naive version breaks stateless dispatch)
  in `projects/garrison/docs/local-rig-learnings.md`. cpuchip/garrison `0e020dd` pushed;
  root records (active.md/lane/journal) committed not pushed. Inbox cleared.
- **★ Deadweight Acquisitions multiplayer — BUILT + DEPLOYED overnight 2026-06-17
  (council + ratify + full stewardship; competitive last-corp-standing).** Dave's
  (the dave-rule namesake) browser space-mining game; Michael forked it, I made it
  multiplayer with a Node WS backend + dockerized + **LIVE at
  https://deadweight.cpuchip.net**. Server-authoritative multi-corp sim (`server/`)
  reuses Dave's Phaser-free worldGenerator+prices; Phaser client = renderer
  (`src/scenes/mp`, `src/ui/mp`); Dave's single-player kept intact (MP additive).
  Deep links `/?mp=1&room=code`. Verified at every rung (sim smoke · WS
  integration test green local+container+**prod wss://** · in-browser).
  Dokploy NOCIX project `deadweight` (`ilaeCtLXDQrQsP9mlK9rX`) / compose
  `l4tkfFkX5GvAvNSlmqR3H`. `*.cpuchip.net` wildcard DNS → no DNS step needed.
  **★ v2 2026-06-17 — made FAITHFUL to Dave's economy** (Michael played it, flagged:
  base wrong / no base menu / free auto-claim). Rebuilt: start 1 hauler + **0 miners**
  + 750cr; mining needs a **purchased AutoMiner (300cr)**; ore → storage (2000),
  **sold manually** at the new base menu (`MpBasePanel` = Dave's `BasePanel`); tonnage
  = DELIVERED; faithful station base+label; humane period-1 (150s/50t). Smoke+wstest
  money-gate green; prod-verified. **★ DEPLOY GOTCHA: auto-deploy-on-push NOT firing**
  — after a push, MANUALLY `compose.deploy {l4tkfFkX5GvAvNSlmqR3H}` (NOCIX) + verify
  prod wstest. Carry: webhook fix; balance; deeper SP micro (net-shuttle/fuel/condition).
  **★ v2.1 — faithfulness fixes + Pause/Quit + ROADMAP (prod-deployed).** Played Dave's
  SP (catalogued gaps: select+designate-button UX [not auto-claim], minimap, EntityPanel,
  big planet, fuller base, deep sim). Fixed: right-click-drag context-menu
  (`disableContextMenu`, left=select/right=pan); **Pause (host)+Quit** top bar (forfeit +
  room GC = dies on server). 13/13 wstest. **`ROADMAP.md` v3–v7** parity plan AWAITING
  RATIFY: auto-claim default (he likes it) + how-deep (honest tension: friends' race may
  be more fun streamlined v3–v4 than full deep-sim v5+). **★ RATIFIED: FULL PARITY (v3→v6)
  + Dave's select→designate default** (auto-claim → opt-in "quick-claim" toggle).
  **★ GOAL = reach v7 (full parity), autonomous Ammon-run, surface only on input-needed.**
  **v3 COMPLETE + LIVE:** designation = Dave's select→designate default + quick-claim toggle;
  **minimap** (DOM `<canvas>` `MpMinimap.svelte` — in-scene Phaser Graphics gets camera-
  zoomed, lesson logged); ship-select EntityPanel; textured planet; actions moved top-left.
  **★ AUTO-DEPLOY NOW WORKS** (Michael granted GitHub-App repo access; fired for 68550b9 —
  manual `compose.deploy` retired). **v4 DONE** (named ships Hauler-NN; per-ship cargo
  upgrades 200/350/550/800; auto-designate toggle — idle miner-haulers auto-claim richest
  rocks; smoke+wstest+browser+prod green). **★ v5a DONE = THE DEEP MINING LOOP** (biggest
  change; `world.ts` rewritten): miners are a purchased POOL; hauler CARRIES a miner out +
  DEPLOYS it; deployed miner mines + ejects nets (net-starved backpressure); hauler SHUTTLES
  nets to base; miner recovered on depletion. MinerSnap + new ship phases (deploying/
  collecting/unloading); client renders deployed miners + tethered nets. smoke+wstest+
  browser(53t)+prod green. **★ v5b-1 DONE + LIVE (`5f180e2`) — read the loop:** click a
  deployed miner (priority over its rock) → detail panel (state/nets/host-rock + RECALL =
  undesignate, keeps the owned miner); **beacons** (net-starved/depleted throb a pulsing
  ring + amber minimap blip); **net-starved alert** (log line + fleet "⚠ N full"). Server
  change minimal (updateMiner takes corp to log beacon once); rest client (mpSelectedMiner
  + nearestMiner + panel). smoke 21/21 · build/typecheck · wstest 18/18 local+**prod**
  (53t; prod bundle content-verified RECALL MINER live) · browser deploy→deliver 0 errs.
  **★★ /goal REACHED — v7 DEPLOYED 2026-06-17 (`638d953` live, /version-verified; Stop-hook
  cleared).** 13-phase autonomous run, each a tested commit+push, all prod-verified:
  v5b-1(`5f180e2`) · clickable-minimap(`e333810`) · orphaned-net-recovery(`daeb562`) ·
  company-arrivals(`d210066`) · Keplerian-orbiting(`9fb0977`) · room-chat(`03a839d`) ·
  build-stamp+`/version`(`1a76ba2`) · hauler-fuel/battery(`a495a2f`) · miner-condition/repair
  (`2e82f13`) · station-fees(`7f2c294`) · 2-miner-bay-milk-run(`0a73848`) · room-persistence
  (`dc36ec4`) · spectator/camera-polish(`638d953`). **Full faithful parity of Dave's SP** (deep
  loop · orbiting · arrivals · fuel/battery · condition/repair · station fees · 2-miner bays)
  **+ MP** (last-corp-standing · chat · persistence) **+ ease-of-life** (minimap nav · build
  stamp · beacons · salvage). **★ AskUserQuestion mid-run** → Michael chose "Full faithful
  parity — all of it" (the v6 systems = SP manual-mgmt, auto-managed/visible in MP). smoke
  45/45, wstest 21/21. Gotchas in memory: stale-:8080 EADDRINUSE silent-fail (free port
  first; smoke=oracle) · resource-stat-invisible-unless-wear-outpaces-service (tie to
  deployment-time/distance) · persistence GC-fix (keep running rooms on empty + TTL sweep).
  NEXT (Michael's, not blocking): a friends-playtest = balance-tuning input.
  `ROADMAP.md` = the record. Journal `2026-06-17-deadweight-multiplayer.md`; memory `project_deadweight_game`.
- **★ Garrison (`garrison-cli`) — COUNCIL CLOSED, P0 RATIFIED 2026-06-18.** Michael
  set voicebox down and picked up Garrison; chose "close the council first" (no
  code yet). The post-cut gate is satisfied (substrate cut 06-15) → P1 buildable
  on his go. All six open questions resolved + the store fork decided
  (FTS5 + chromem-go RRF in P1, his call) — full text in the spec's "Decided in
  council — P0 CLOSED". Model steer: borrow the pg-ai **FlexLLama** rig (one
  `:8090` endpoint, aliases `qwen3.6-27b`/`gemma-12b`/`nemotron-4b`; embeddings
  LM Studio `text-embedding-nomic-embed-text-v1.5` `:1234` — Ollama isn't
  installed) — captured as a table in the spec ("Starting model configuration").
  **★ P1 FLOOR SHIPPED + pushed 2026-06-18 (Michael: "create the repo... and
  start building"):** repo `cpuchip/garrison` (private) → `projects/garrison/`
  (gitignored from root). One stdlib OpenAI-compatible Go client + a CLI proving
  the model path against LIVE endpoints — `garrison ping` (chat `:8090` + embed
  `:1234` up, all 4 role models served ✓), `embed` (768-dim nomic vector), `chat`
  (nemotron-4b 1s + qwen3.6-27b correct iterative fib 34s); go build+vet+real
  round-trip = oracle green. **★★ /goal "get to the dogfood stage, full ammon
  loop" — REACHED + verified in one autonomous run (G1→G5, each a tested
  commit+push):** G1 ledger (modernc pure-Go SQLite — work-item hierarchy/
  recursive-CTE Tree, FK'd messages, cost rollup) · G2 retrieval (FTS5 keyword +
  chromem-go vectors fused by RRF; LIVE fusion test green) · G3 loop
  (`internal/agent`: Dispatcher + forgiving ===FILE=== parser + path-safe apply +
  Verify oracle + `Loop.Run` writing every step to the ledger + `garrison run`;
  mock-dispatcher tests drive real go build/test) · **G5 DOGFOOD** — qwen3.6-27b
  wrote package strutil (rune-correct Reverse + WordCount) + tests through
  Garrison's OWN loop, oracle-passed in 1 attempt / 1m29s / 4684 tok;
  inverse-hypothesis re-ran go test → 7/7 incl CJK reversal; `docs/dogfood-01.md`.
  **Harness>Intelligence shown on itself.** Deps: modernc.org/sqlite v1.52 +
  philippgille/chromem-go v0.7. go.work landmine: garrison has a gitignored
  `go.work`; loop runs targets w/ GOWORK=off. **★ G4 DONE → P1 COMPLETE**
  (Michael: "lets push and finish p1"): G4a skills-as-data (`internal/skills`,
  injected into prompts, `garrison skills`, 2 seed skills) · G4b gated exec
  (`===RUN: cmd===`, `--allow-exec`, oracle still gates) · G4c MCP-client stub
  (`internal/mcp`, JSON-RPC init/list/call, net.Pipe roundtrip test). DOGFOOD
  verified TWICE (strutil + mathx, 1-attempt each; skills-active run used 6k±1
  prime opt). **9 tested commits pushed to `cpuchip/garrison`; all build+vet+test
  green.** Root PUSHED 2026-06-18 (Michael's "lets push"; `.spec/`+`.mind/`
  Garrison records; ibeco.me unaffected). **★★ P2+P3 DONE + LIVE-VERIFIED
  2026-06-18 ("set a goal to do p2 and p3"; 4 more tested commits):** P2.1
  acceptance tests (`--acceptance` copies+protects; `splitProtected` refuses
  edits to them) · P2.2 `internal/detect` (gofmt-auto-`-w`/vet/exported-doc/
  naked-panic) · P3 `critic.go` (gemma-12b 2nd look; VERDICT APPROVE/REVISE,
  bounded). LIVE medianx proof: acceptance held byte-identical, critic APPROVED
  clones-before-sort (mutation the copy-passing test missed). Inverse-hyp caught
  STALE-binary (rebuild `-o garrison.exe`!) + gofmt-blocks-on-no-EOF-newline
  (→ auto-format `gofmt -w`, formatter≠linter). `docs/dogfood-02.md`. **P1–P3
  COMPLETE.** Root records committed, NOT pushed (Michael pushes root).
  **★★ P3.5 (spawn+watch) + COUNCIL MODE + PHASE-4 BROWNFIELD COMPLETE
  (overnight 2026-06-19, ~12 tested commits):** P3.5a spawning (decompose→
  preside→integrate) · P3.5b `garrison watch` TUI · council mode (`garrison
  council`, /run→ratify→build) · **R5 resilience** (retry transient rig errors —
  built first so the night survives lockups) · **R1** read file contents · **R2**
  surgical `===EDIT===` SEARCH/REPLACE · **R3** relevance-ranked read · **R4** git
  `--commit` (`internal/vcs`, excludes .garrison) · **tickets** `--ticket FILE`.
  **LIVE brownfield proof:** existing repo + buggy Add → surgical fix + commit
  `4aebfc5`, verified, .garrison excluded; R5 carried it through a 1.4-tok/s rig.
  Garrison now edits EXISTING repos. `docs/dogfood-03.md`. **★★ P5 (TUI) + P6
  (context engine) COMPLETE 2026-06-19 (Michael: "finish p5 and p6, save p7"; 6
  more tested commits):** P6 `internal/contextx` (pressure gauge + `Compact`
  summarization, wired into council) · P5.1 `agent.Stats`/`StatsFromLedger`
  (flow/time/cost/children/pressure from the ledger) · P5.2 `agent.Control`/
  `BasicControl` (pause/emergency-stop[cancel+account D&C121]/inject; loop checks
  each iter) · P5.3 `garrison drive` (bubbletea TUI DRIVES a run — live stats bar
  + tree + log + keys space/i/s/q; pure parts tested, interactive shell to-spec
  for live tuning). **Garrison ≈ Claude-Code-shaped now** (greenfield+brownfield ·
  spawn · council · driving TUI w/ modes/stats/pause/inject/emergency-stop).
  README P1–P6; ~20 tested commits this arc. **NEXT (WITH Michael): P7 substrate-
  MCP, then multi-lang oracle; the driving TUI wants a live interactive shakedown.**
  Journal Update 7. Journal `2026-06-18-garrison-council-closed-p0.md`.
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
- 2026-06-25T02:00:15 background (Bash): cd "C:/Users/cpuch/AppData/Local/Temp/claude/C--Users-cpuch-Documents-code-stuffleberry-scripture-study/c4fef1d0-292c-4ad6-b6c5-76e2af1043c3/scratchpad" && pyth
- 2026-06-25T01:24:15 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/external_context" && echo "=== existing ===" && ls -1 && echo "" && echo "=== cloning (shallow, d
- 2026-06-22T10:39:46 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/garrison" && rm -rf /tmp/garr-journaltest && mkdir -p /tmp/garr-journaltest && printf 'm
- 2026-06-19T00:52:07 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/garrison" && GOWORK=off ./garrison.exe run --dir "C:/Users/cpuch/AppData/Local/Temp/garr
- 2026-06-18T23:53:20 background (Bash): DOG="C:/Users/cpuch/AppData/Local/Temp/garrison-spawn"; rm -rf "$DOG"; mkdir -p "$DOG"; printf 'module textkit\n\ngo 1.26\n' > "$DOG/go.mod"; cd "C:/Users/cpuch
- 2026-06-17T16:45:35 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git commit -q -F - <<'EOF' ui polish: on-station layout
- 2026-06-17T15:51:51 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git status --short && git commit -q -F - 
- 2026-06-17T14:51:05 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git status --short && git commit -q -F - 
- 2026-06-17T14:17:05 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git status --short && git commit -q -F - 
- 2026-06-17T12:59:31 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git status --short && git commit -q -F - 
- 2026-06-17T11:50:17 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && npx --no-install playwright-cli close >/dev/null 2>&1; 
- 2026-06-17T11:44:26 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git commit -q -F - <<'EOF' v7: room persi
- 2026-06-17T11:32:44 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git commit -q -F - <<'EOF' v6: multiple m
- 2026-06-17T11:25:47 background (Bash): cd "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/projects/deadweight-acquisitions-game" && git add -A && git commit -q -F - <<'EOF' v6: station se
- **`dwa-game` docker container (2026-06-16):** Dave's "Deadweight Acquisitions"
  SINGLE-PLAYER (Vite+Svelte5+Phaser3) running detached on **localhost:5173** for
  Michael to play. Image `dwa-dev` (node:lts-alpine toolbox); repo at
  `external_context/deadweight-acquisitions-game` (gitignored, the read-only clone).
  Michael's play session — leave it. Stop: `docker rm -f dwa-game`.
- **Deadweight MULTIPLAYER is DEPLOYED (2026-06-17), not a local container** — lives
  at https://deadweight.cpuchip.net (NOCIX Dokploy, auto-deploy on push to the fork
  `projects/deadweight-acquisitions-game`). The local `deadweight-game` test compose
  was removed after prod verified. No standing local process for the MP build.
- pg-ai-stewards-persona-host docker container: rebuilt + recreated 2026-06-11
  (current code: r21 + Callie + deference + sync). The native persona-host.exe
  duplicates are DEAD — do not relaunch; rebuild the container instead.

## Handoffs / notes
- 2026-06-11: board surgery done (active.md → lean; full ledger in
  .mind/archive/active-ledger-thru-2026-06-11.md).
