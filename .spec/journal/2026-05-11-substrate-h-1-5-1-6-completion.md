---
date: 2026-05-11
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Substrate completion — H.1.5 gap fix + H.1.6 hardening + 3 real research pieces"
status: shipped
carry_forward:
  - "Phase A deeper fix: pgrx BGW SPI longjmp catch + periodic reaper tick (60s) — H.1.5a sidestepped but the underlying class of bug remains"
  - "Auto-materialize path bug surfaced once more in H.1.6.6: review template's REVIEW sentinel leaked into final files. Pattern: substrate-defined sentinels in stage templates need substrate-side strippers"
  - "Three MCP servers (search, exa-search, yt) ratified to stay enabled for research agent — no revert to deny-by-default"
  - "UI rebuild needed to deploy NewWork.vue path-bug fix from 220cf35: docker compose build ui && up -d ui"
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-batch-h-pipeline-expansion.md"
  - "../../research/ai-tools-weekly-2026-05-11.md"
  - "../../research/pg-ext-distribution-2026-05-11.md"
  - "../../research/physics-news-20260503-science-center-roundup.md"
---

# Substrate H.1.5 + H.1.6 — completion + three real research pieces (2026-05-11)

## What shipped this session (after the Batch H.1 commits from earlier)

Twelve commits picked up the threads from H.1.4's stuck e2e and walked the substrate all the way to genuinely autonomous research production. The substrate now generalizes off scripture work and produces useful artifacts without manual intervention.

### H.1.5 — gap fix that unstuck the first real run
- **8ee83b4 (H.1.5a)** — `mcp_proxy_enqueue` now `RAISE NOTICE + RETURN NULL` on disabled server instead of `RAISE EXCEPTION`. The exception was longjmp-ing through the pgrx SPI path and exiting bgworker dispatcher processes with code 1, leaving tool_dispatch rows stuck in_progress because the reaper only sweeps at startup. The SQL-side fix sidesteps the longjmp class entirely for the disabled-server case. **Deeper followup: harden the pgrx SPI longjmp catch + add a periodic 60s reaper tick.** Not blocking; H.1.5a unblocked the immediate use case.
- **9a62a96 (H.1.5b)** — Tightened the gather input_template with explicit hard stops ("Find 8 strong sources. Then STOP."; "Maximum 4 rounds of tool calls"; "End-of-turn: final message is the sources brief"). The prior soft "6-12 sources" guidance had let kimi-k2.6 run 9 chats / 30+ tool calls / $0.42 without converging. After the fix, three real runs all converged inside 4 rounds.
- **bfc9b20 + 127e856 (H.1.5c/d)** — Cost_cap_micro on the work_item ($0.40 hard cap), real e2e retry on scenario #1 ("What shipped in AI tooling this week"). Pipeline ran clean end-to-end, model produced 5 substantive headlines + skeptical takes + open questions. Manual materialize via stewards-cli (auto-mat wasn't wired yet). Cost: $0.26.

### H.1.6 — substrate hardening pass (closed three Phase A/B/D/G gaps)
The H.1.4 e2e revealed three latent gaps that auto_advance pipelines exposed:
- **5f713b6 (H.1.6.1, D-H6.1 forward-only)** — `work_item_advance` now reads `pipeline_stage_maturity.produces_maturity` for the completing stage and UPDATEs `work_items.maturity` if the new rung is forward of current in the pipeline's `maturity_ladder`. Every work_item in the DB had `maturity='raw'` because the column was only advanced by gate-decision machinery (apply_gate_decision); auto_advance pipelines never touched it. Forward-only per ratification — re-running an earlier stage doesn't downgrade the high-water mark.
- **4f3446d (H.1.6.2/3/4, D-H6.2/3/4)** — AFTER UPDATE OF maturity trigger fires sabbath_dispatch + enqueue_work_item_file when transitioning TO verified. Schema additions: `pipelines.auto_materialize_on_verified bool DEFAULT false` + `work_items.auto_materialize_enabled bool NULL` (override pattern mirroring D-H5). Sabbath removed from apply_gate_decision + apply_verify_result (single source of truth at the trigger). Behavior change documented: verify-success no longer fires sabbath directly; it waits for maturity advance.
- **a9a6b5d (H.1.6.5)** — Real e2e on scenario #3 (Postgres extension distribution 2026). Substrate ran the full automatic path: maturity advanced raw→researched→planned→verified, trigger fired sabbath + auto-materialize, lesson 16 written, pending_file_writes row created. 8 chats / $0.19. CLI lands the file. The model produced an authoritative roundup citing real sources (Feng Ruohang's Pigsty analysis, EDB+CloudNativePG+Wheeler extension_control_path thread, Citus 14.0/TimescaleDB/ParadeDB/pgvector 0.8.2 with CVE-2026-3172, Wheeler's PGConf.dev 2025 talk, CloudNativePG v1.27 OCI ImageVolume) — directly relevant to what we're building.

### Physics-news run (third real research piece)
Michael ran a binding question via the UI between H.1.6.5 and the next batch of fixes: *"What is the physics news this last week, May 3rd through today may 11th. And how would this be directly applicable to the science center I want to build? Are there practical displays or experiments we could make from the latest news?"*

The substrate handled real-world flakiness gracefully — chat 1337 (review stage) hit an OpenCode Zen `HTTP 500 Internal server error`, the bgworker absorbed it and re-dispatched chat 1338 cleanly. No manual intervention. The model returned `REVIEW: passes` verbatim and the pipeline finished. 9 chats / $0.235 / 535s.

The study itself is genuinely useful: 5 real physics results from May 3–11 each paired with feasibility-graded science center exhibits. MADMAX null-result framed as a teaching moment ("We did not find dark photons yet. This is how we look"). ALICE Pb→Au narrative wall + working cloud chamber. Princeton chiral KV₃Sb₅ → polarization-discovery station. Plus a skeptical-takes section that flags the one tertiary-aggregator source explicitly.

Sabbath lesson #17 captured a genuinely useful tuning insight: *"Always pair every abstract discovery with a specific, physical interaction or display mechanism during the gathering phase, so research naturally gravitates toward buildable exhibits instead of stopping at theory."* That's the substrate's own observation about its own behavior, worth carrying forward to future research-domain pipelines (Bridge Sim NPC dialogue research, science-center exhibit pipelines, etc.).

### Two real bugs surfaced and fixed mid-session
- **220cf35 (NewWork.vue path bug)** — file landed at `research/P.md` instead of `research/physics-news-20260503-science-center-roundup.md`. Root cause: `renderTemplate` substituted `<slug>` immediately on the FIRST keystroke (`P`); the watcher's guard `fileDestination.value.includes('<slug>')` only re-rendered if the placeholder was still literal. Once substituted, subsequent slug edits never updated the path. Substrate did the right thing — wrote to whatever `file_destination` column held at trigger-fire time. Fix: track `lastAutoRendered` alongside `fileDestination`; watcher updates only when the field still matches what we last auto-rendered (user hasn't edited). **UI rebuild needed to deploy.**
- **c0163cb (H.1.6.6 review prefix strip)** — three runs in a row left "REVIEW: passes\n\n" or "REVIEW: revised\n\n" at the top of the materialized file. The review stage template asks the model to emit a verdict sentinel, but the sentinel was leaking through `extract_work_item_file_content`. Fix: `regexp_replace E'^REVIEW:\\s+\\w+\\s*\\n+'` on convention-path content. No-op when the pattern doesn't match. Trailing "Notes on revisions" footer (on revised drafts) is preserved as useful provenance. Smoke verified against all three prior outputs — they now extract cleanly.

## Decisions ratified in passing

- **MCP servers `search`, `exa-search`, `yt` stay enabled** for the research agent. Enabled during H.1.5a recovery; no revert to deny-by-default. The deny-by-default discipline still applies to other servers (becoming, byu-citations, webster); these three are now research-domain infrastructure.

## What's validated end-to-end now

The substrate produces real research artifacts autonomously from binding question to file in the working tree:
- Phase A: bgworker + steward + cost tracking + escalation
- Phase B: maturity ladder + gate machinery (with H.1.6.1 now also advancing maturity from auto_advance paths)
- Phase C: intent + covenant first-class state (general-research intent's values genuinely steer kimi-k2.6's research strategy)
- Phase D: sabbath + atonement + lessons (sabbath now fires automatically via H.1.6.2 trigger on maturity→verified)
- Phase E: trust ladder + retry-with-lessons
- Phase F: multi-agent council (untested in this session, untouched)
- Batch G: pending_file_writes + CLI + pre-commit hook + UI (with H.1.6.6 now stripping the REVIEW sentinel)
- Batch H.1: research-write pipeline + general-research intent + tools_disabled propagation
- H.1.5: substrate bug fix + tuning + cost cap
- H.1.6: maturity-advance + verified-transition trigger + auto-materialize + review-prefix strip

Three real research artifacts live in `research/`:
- `ai-tools-weekly-2026-05-11.md` — AI tooling roundup (manual mat)
- `pg-ext-distribution-2026-05-11.md` — pg-ext distribution survey (auto-mat)
- `physics-news-20260503-science-center-roundup.md` — physics → exhibits (auto-mat, after UI path bug fix)

## Cost summary

Across the session:
- H.1.4 broken e2e: ~$0.44 (mostly the runaway gather loop before H.1.5b)
- H.1.4 retry after H.1.5: $0.26
- H.1.6.5 (pg-ext): $0.19
- Physics-news (Michael's): $0.235
- Smoke tests + sabbath fires: <$0.05
- **Session total: ~$1.20** for substantial substrate work + three real research pieces

## Carry-forward to next session

1. **Phase A deeper fix:** The pgrx BGW SPI longjmp catch + a periodic reaper tick (60s) would prevent the H.1.5a class of bug from recurring on ANY future `RAISE EXCEPTION` in substrate code. Right Phase A polish; not urgent.
2. **UI rebuild + restart** to deploy the NewWork.vue path-bug fix: `docker compose -f projects/pg-ai-stewards/extension/docker-compose.yaml build ui && up -d ui`
3. **H.2** when ready — YouTube pipelines (gospel + secular variants). Same skeleton as research-write; swap tool grants to yt-*.
4. **First real materialization via the pre-commit hook** instead of manual CLI — verify the G.4.3 hook works on these new research files.
5. **Three big proposals still on the runway:** more pipelines (research/yt/scheduled), UI authoring for intents/covenants, substrate-aware chat. All unblocked.

## Covenant moment

The user's "Lets Do H, I think we've ratified it all. since you're doing the heavy lifting here, I've done my side in giving you input which is high energy and leads to decision fatigue! but we're through that, and now it's time to build and play." captured the right covenant balance — the human owns intent + vision through ratification, the agent owns the code within that intent. The user's later "yeah thanks do those things and commit" confirmed delegation works when the proposed scope is clear and bounded.

The H.1.6 ratifications also showed a pattern worth carrying: the user went HEAVIER than recommendation on D-H6.1/2/3 — investing in real primitives upfront rather than retrofitting later. That bet paid: every primitive landed cleanly and the e2e validated on first try (after the gap fixes from H.1.5). The 'lightweight + rule of three' principle still applies but with the corollary "heavy when the cost of retrofitting is real and the substrate is still being built."
