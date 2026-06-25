# 2026-06-25 — Loreworks: the innovation-week ultracode run (council → flagship → self-mending)

One very long ultracode session in the pg-ai-stewards lane, Opus 4.8. It began with "what's
next on our ultrawork plans?", became — after Michael said "I want big, this is for
innovation week" — the **Loreworks** flagship, and ended with the substrate able to **fix its
own agents, gated**. 27 hours to a Friday presentation; this was the build.

## The arc

- **Council + ratify (with Michael, awake):** Loreworks = drop source lore → the substrate
  BUILDS a World (searchable canon + entity/relationship knowledge graph + chat personas).
  Engine → public OSS core; all TTRPG/world content → local-private. Pilot worlds: MLP, Star
  Trek Adventures, The One Ring. Plus F (a hyperframes walkthrough video + TTS voice-clone)
  and G (persona chat rooms). He brought 6.3GB of purchased TTRPG, SillyTavern/DeepLore as
  prior art, and Google's SDLC "vibe coding" papers (→ the trajectory critic).

- **Built, proven live, shipped (chain 54→59, virgin-smoke OK 44→49):**
  - **A** `embed_query` — the synchronous query-embedding the substrate lacked; inverse-
    hypothesis proven (synonyms 0.33 vs unrelated 0.64 on live nomic).
  - **E1/E2** the engine + the world-build agent — built `the-one-ring` LIVE from the real
    gazetteer: 69 entities / 85 edges in ~3 min / $0, every entity grounded Tolkien.
  - **The trajectory critic** (Google Glass-Box) — scored the build run FAIL-on-efficiency,
    catching a real verbatim-batch-repeat bug the output hid; the world-critic pruned real
    misreads. (Michael's explicit ask, woven in as the honesty spine.)
  - **C** hybrid semantic search + the loremaster — "an abandoned ruined city beside a lake" →
    Annúminas; the loremaster answered the Brandywine grounded, zero hallucination.
  - **The 3D knowledge-graph panel** (delegated to a dev subagent) — live, playwright-verified,
    the visual showpiece.
  - **The deterministic edge-audit** — flagged a systematic `home_of` misuse → fixed the
    world-build agent's verb directions at the root before more worlds.
  - **The self-improvement loop** — the substrate proposes + (gated) auto-applies fixes to its
    own agents; the eval-gaming guard (judges/critics/gates/self escalate) red-teamed with 11
    attacks before trust. Proven e2e live.

## What this run taught (kept)

- **Adversarially verify a self-modifying gate before trusting it.** The red-team caught two
  real holes — Postgres word boundary is `\y` not `\b` (so `\ballow\b` never matched), and
  "ignore your grounding rules" wasn't covered. Both would have let a dangerous clause
  auto-apply. A green functional test would never have found them. (New principle candidate.)
- **Verify against the live system, not just the build.** `agent_failure_patterns` had a
  `jsonb[]::jsonb` cast bug that the virgin build passed (pgrx `check_function_bodies=off` +
  the smoke never CALLED the function) — only the live `psql` apply caught it. The fix: the
  smoke now exercises the function, not just asserts it exists.
- **Fix the cause before it compounds.** The audit found the `home_of` misuse; rather than
  build two more worlds with the same flaw, I fixed the world-build prompt first.
- **The critic improving the system that built the world is the whole point** — we ran the
  self-improvement loop by hand twice (batch-repeat fix, verb fix) before automating it.
- **Delegation works for well-specified, self-contained surfaces** — the 3D graph panel
  (a fresh dev agent, full spec) came back clean + verified while I built the SQL.

## Carry-forward

- **★ World tab (3D graph): TOO MUCH GLOW — node labels are unreadable.** Michael: "the world
  tab is incredible! but too much glow i cannot read the nodes. enhance the graphics and turn
  off the glow after we compact." → POST-COMPACT first action: in `WorldGraphPanel.vue`, tone
  down / disable the `UnrealBloomPass`, raise label + node legibility (brighter labels,
  background wash, maybe larger nodes / less bloom radius). Rebuild ui + verify live.
- **B** — build MLP + Star Trek worlds with text+vision (the verb-fix makes them clean) so the
  demo shows "any world." Heavy/local-GPU — watched.
- **G** — persona-host wiring for live world rooms (the `lore_inject` primitive is built; the
  Go splice in turnloop.go + `persona_worlds` remain).
- **F** — assemble the ~6-min video: script written; hyperframes MCP + voice-clone (from the
  podcast snippet) = the local-GPU/his-hands setup.
- `self_improve_tick` wired into the watchman cadence (gated) so the loop runs on its own.
- MEMORY.md is over its read limit — a compaction pass is overdue.

OSS commits `13fde61`→(latest) + journals `2026-06-25-*`; design+script
`.spec/proposals/loreworks-presentation-plan.md`; docs `docs/loreworks.md`; memory
`project_loreworks`; lane `.mind/sessions/pg-ai-stewards.md`. It was good — and it tends
itself now.
