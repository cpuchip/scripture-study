---
lane: pg-ai-stewards
session_id: 7ea7faa4-688a-451a-ac68-b7ea662d4b81
status: active
started: 2026-06-11T22:00:16
last_active: 2026-06-13T09:51:34
---

## Working on
- **★ M1 — CODER-MCP SHIPPED 2026-06-13 (OSS `321176c`+`7897093`, PUSHED to
  public main after Michael's Hinge ② ship nod).** The inert 20-coder
  surface is now alive. coder-mcp folded into the root module (was own module on
  go-sdk v1.6.0→v1.6.1, builds clean). Files: cmd/coder-mcp/{main,tools}.go +
  sandbox/sandbox.go (clean-room; **CODER_REPO_ALLOWLIST now DENY-ALL default**,
  commit author env-configurable); extension/coder-runtime.Dockerfile (hardened
  sandbox, non-root coder uid 1000); bridge.Dockerfile (+coder-mcp +docker-cli/
  git/github-cli); **docker-compose.coder.yaml** (OPT-IN override — default `up`
  stays socket-free); .env.example coder section; **SECURITY.md** (trust model +
  hardening review = the ship-gate doc). **VERIFIED:** coder-mcp builds+vets;
  `coder-mcp -smoke` PASS (sandbox spawn w/ cap-drop=ALL/no-new-privs/mem-cpu-pids
  caps/non-root → Go1.26/Node24/Python3.11+LSPs → write+build → teardown);
  `bridge refresh-tools` = **3/3 OK, coder [OK] 16 tools** (was [FAIL] in M0).
  Hardening confirmed: token never in sandbox (bridge-side one-shot cred helper);
  deny-all allow-list; protected-branch refusal; reaper. **★ HINGE ② CLOSED —
  Michael's ship decisions (2026-06-13):** (1) socket off-by-default public +
  **gitignored `docker-compose.override.yaml` = on for us this machine** (Compose
  auto-merges; verified socket in merged config); (2) egress on-by-default +
  **`CODER_SANDBOX_NETWORK=off` kill-switch** added+documented (forces every
  sandbox `--network=none`; -smoke PASS both modes); (3) coder row stays enabled.
  cv4 minimax-m3 model seed → overlay (still carry-forward). **Task #158 DONE.
  MCP-packaging M2 (fetch-md+git into cmd/ + Exa re-point + archive search-mcp) +
  M3 (BYO-MCP docs) remain.**
- **★ GOSPEL-ENGINE RESOLVER GENERALIZED 2026-06-13 — core `4bb80ab` + overlay
  `90906f7`, both pushed** (Michael: "generalize as much as possible; shouldn't be
  project/workspace aware but configurable to pull external resources"). The
  "resolver" was a whole scripture-citation subsystem in `schema.rs` (the file the
  SQL-file audit missed; unused by core pipelines; overlay already owned the
  consumption). **CORE:** GospelEngineConfig→ResolverConfig + STEWARDS_RESOLVER_URL
  ({ref} template)/TOKEN; resolve_ref config-driven (boot log "resolver url=…" not
  gospel-engine); parse_gospel_links→parse_doc_links (ALL md links, external|doc);
  normalize_book+parse_reference REMOVED; refresh_doc_refs/doc_citations_resolved
  generic; provider→'resolver'; example agents/skills + stray comments genericized
  (scripture-linking→reference-linking, doc_citations kind enum, summarize prompt,
  01-graph prior-art); verify-2-1/2-2 removed. **★ BEHAVIOR CHANGE flagged: core
  import_doc now cites ALL links generically.** Both smokes green (genresolver image,
  tests/virgin-smoke.sql passes); genresolver→pg18 retag (compose default).
  **OVERLAY:** `scripture-resolver.sql` restores the scripture funcs + import_doc/
  refresh/doc_citations_resolved overrides (scripture CITES + verse decomposition);
  doc_citations_resolved keeps core signature (extension-owned, can't DROP/retype →
  `resolved` carries the verse array); STEWARDS_RESOLVER_URL in .env; manifest+
  classification entries; replay-proven (6 funcs, "Mosiah 18:8-9"→2 verses, 'scripture'
  kind). **HARNESS CLEANUP DONE (OSS `b6ec106`, pushed):** verify-1-6-1/loop/4a-steward
  scripture fixtures→water-cycle/web_lookup; verify-3e2-2 mcp_proxy test re-pointed
  gospel-engine-v2/gospel_search→core fs-read/fs_search; init brain smoke Moroni/charity
  →water-cycle (category 'study'→'ideas'; full init runs clean, fts=1); bridge.Dockerfile
  example neutralized. **Extension-wide sweep now scripture-FREE.** Task #159 fully closed.
- **★ M0 — RUNTIME STACK SHIPPED + VIRGIN BOOT PROVEN 2026-06-13 (OSS `8287967`,
  pushed; Michael: "this probably doesn't need [a loop], push through as normal").**
  The OSS repo had no runtime image/compose — only the extension Dockerfile. Added
  `extension/bridge.Dockerfile` + `extension/persona-host.Dockerfile` +
  `extension/bridge-entrypoint.sh` + root `docker-compose.yaml` + `.env.example`.
  **Clean-room single-module win:** no go.work, no sibling-stub COPYs, no personal
  MCP — the bridge image COPYs `go.mod`+`cmd/` and builds 3 binaries (stewards-mcp/
  fs-read-mcp/stewards-cli) in **~6s** (vs the workspace's multi-min go.work build).
  persona-host behind compose `--profile personas` (needs ai-chattermax+key; idles
  without). Core installs via `CREATE EXTENSION` (pg init), so the entrypoint just
  starts the bridge — no core migrations. Ports offset (pg 55434) for side-by-side.
  /workspace mounted RO (autonomous materializer opt-in; the boot warning is the
  safe default, documented in compose). **Virgin boot GREEN** (scratch
  `pg-ai-stewards-oss` project, live untouched, torn down after): CREATE EXTENSION
  → pg_ai_stewards 0.2.0 + pgvector; **4 bgworker dispatchers alive** (recovered from
  the bootstrap-phase "db does not exist" FATALs — transient, matches live); bridge
  connects + `LISTEN`s on stewards_mcp_proxy; **`bridge refresh-tools` spawns the 2
  real stdio MCP servers e2e** — fs-read [OK] 3 tools, pg-ai-stewards [OK] **31 tools
  all doc_*** (no study_* leak), coder [FAIL] = the ONLY failure = expected M1 gap
  (binary not built). **★ TWO CLEAN-ROOM FINDINGS (Michael's call, NOT acted):**
  ① **the Rust core still carries a `gospel-engine` resolver subsystem** —
  `GOSPEL_ENGINE_URL`/`TOKEN` env in `bgworker.rs` (prints `stewards: gospel-engine
  url=…` every boot) + `GospelEngineConfig`/`GOSPEL_ENGINE_CONFIG` OnceLock in
  `providers.rs` (Phase 2.2). Personal-domain name in the public core; the src audit
  missed it. **Design question** (genericize the resolver vs move to overlay), not
  act-and-report. ② `stewards-cli migrate` hardcodes the workspace path
  `<repo-root>/projects/pg-ai-stewards/extension` (`migrate.go:54`) — wrong for OSS
  layout; belongs to the two-tier runner work (M0 doesn't need it — core is
  CREATE EXTENSION). **M1 NEXT = coder-mcp port + hardening review (Hinge ②).**
- **★ MCP PACKAGING PLAN RATIFIED + committed/pushed 2026-06-13 (OSS `f603e34`,
  `.spec/proposals/mcp-packaging.md`).** Where the workspace MCP servers ship
  relative to the substrate, decided on Go-module coupling. **No separate
  `pg-ai-stewards-mcp` repo** (daemon leg already collapsed `cmd/*` into one
  module; a split regresses it). **T1 substrate-intrinsic** (cmd/): stewards-mcp✓
  / fs-read-mcp✓ / persona-host✓ shipped; coder-mcp pulled in M1 behind Hinge ②.
  **T2 generic utilities** (cmd/): ship fetch-md-mcp + git-mcp; **archive
  search-mcp** (verified = the 2026-02-03 DuckDuckGo server, throttle-unreliable,
  predates substrate, NOT a custom substrate search) + **re-point core
  `web_search` tool_def → `web_search_exa`** (remote mcp.exa.ai, operator-keyed,
  no search binary ships). **T3 domain** (gospel/webster/strongs/byu/becoming/
  yt/brain/md) stay own repos — "bring your own MCP," referenced not absorbed.
  **Phased M0** (runtime/bridge Dockerfile + docker-compose.yml — OSS has NONE
  yet, the prerequisite gap; cross-compiles cmd/*-mcp → /usr/local/bin) → **M1**
  coder-mcp port + HARDENING REVIEW = Hinge ② → **M2** fetch/git + Exa re-point +
  archive search → **M3** bring-your-own-MCP docs + example overlay. M1 = the
  coder-wave Go half (task #158); awaiting Michael's go before executing M0.
- **★ CODER WAVE — SQL SURFACE SHIPPED 2026-06-13 (OSS `a943a95`, pushed; Michael: "do the SQL surface first").** `20-coder.sql` consolidates cc2-6/cv2-2/cv3-12/r10/r12: a GENERIC clean-room `dev` agent (the workspace's 17K personal dev/debug prompts stay overlay) + the `coder` MCP server (★ **INERT** — points at /usr/local/bin/coder-mcp, not built yet) + code-write / code-pr (7-stage final clone→plan→plan_review→implement→verify→review→pr, taken from the live final per l13) / code-deploy (prepare = always-escalate Hinge) / subagent-research-codebase pipelines + stage_models + maturity + research_codebase (clean, active) + scoped `dev` coder grants + the read-only research-codebase deny-list (study_*→doc_*). Two GRAFTS onto core finals (not pastes): work_item_advance (08 body + cv6 review + cv11 plan_review loop-backs, maturity hook preserved) + work_item_dispatch_stage (19 r3 body + cv7/cv10 review model-immunity). lib.rs: create_coder requires create_models. Virgin smoke FULLY GREEN incl. both grafts e2e (review REVISE→implement / PASSES→pr; deploy prepare→awaiting_review Hinge; dispatch critic uses input.review_model not the override), deploy escalate-gated, research-codebase read-only (8 denies/0 allows), no token value, repos genericized. **CODER REMAINING = Hinge ②: the coder-mcp Go server extraction (cmd/coder-mcp → OSS module + Dockerfile cross-compile to /usr/local/bin/coder-mcp) + the HARDENING REVIEW** (sandbox isolation, bridge-side token, repo allow-list, resource caps) — the public-ship gate, a fresh focused pass. cv4 minimax-m3 → overlay model seeds. Then the **CUT** (Hinge ①+③; live idle → soak can relax).
- **B6 tests/ + CI SHIPPED + CI GREEN 2026-06-13 (OSS `8509d26`→`9812d3f`, pushed):**
  `tests/virgin-smoke.sql` = ASSERT-based virgin-boot regression gate
  (vector-only / no-pgcrypto / no-AGE; doc_* complete; a representative object per
  subsystem 00-19 + the 4-layer dispatch FINAL; **no operator/personal seeds incl.
  no personal MCP** — only fs-read + pg-ai-stewards core daemons; spine e2e with
  capability-substitution). `.github/workflows/ci.yml` runs it on push/PR
  (extension build+virgin-smoke + go build/vet) — **full run GREEN 4m54s**, actions
  on checkout@v6/setup-go@v6 (Node-24, deprecation resolved). README CI badge;
  `tests/README.md`. **seed_harness genericize VERIFIED** (virgin boot = all-generic
  agents/intents=0/core-MCP-only); **anatomy doc clean**. .gitattributes already eol=lf.
  **B6 cutover-prep DONE this session (workspace `6bdeef9`+`0cb5cd3`):** rename-map
  finalized through B5; **overlay re-author + OVERLAY-REPLAY PROOF GREEN** (35/35
  overlays apply on a virgin core — h1-1/h3-2 scripture_anchor→values_anchor, init-01
  AGE→relational import_workstream, pe7-seed-ai-news-7am filed [the B5/18 orphan];
  the ~15 other study_*-grep overlays apply clean as-is — 'study-write' is a valid
  operator pipeline name, not a renamed-object ref; both scheduled pipelines land;
  harness `parity/overlay-replay.sh`). **★ B6 / CUTOVER-PREP COMPLETE — 20 live↔repo
  mismatches CLASSIFIED, GREEN, ZERO DRIFT** (workspace `9566517`,
  `parity/mismatch-classification.md`; OSS blueprint `b474bb4`). Live
  (`pg-ai-stewards-dev`, read-only) vs rebuilt core+overlay: 101 raw body-diffs →
  30 genuine after normalizing comments/whitespace/renames; ALL accounted —
  deliberate clean-room (AGE→relational, config genericization, consolidation
  finals, doc_* renames, todos lowercase), false-positives (formatting / END vs
  END;), one rebuilt-fixes-live bug (provider_cap_refill RAISE %.2f), and ONE
  deferred-P2 gap (work_item_advance code-pr revise loop → 20-coder). Rebuilt P1 ≡
  live minus deferred P2. bgworker `_kind` enum = deferrable Rust refactor. **ONLY
  Hinge-gated work remains: the CUT** (Hinge ①+③; Michael not using live →
  low-risk, soak can relax) + the **coder wave** 20-coder.sql (Hinge ②; must
  re-add the work_item_advance code-pr arm). Cut-planning: the
  work_item_promote_trigger unwrapped-PERFORM sabbath tension.
- **★ AUTHORING LEG COMPLETE 2026-06-13 — B5 SHIPPED, chain runs 00→19, migration manifest = ZERO migration entries (verify/test harness only).** All 189 historical migrations consolidated into 20 authored subsystem files. B5 commits (all pushed, virgin-smoke green each):
  - **17 (`35d66a6`)** personas — `17-personas.sql`: persona agent + persona-turn pipeline (r7) + lmstudio/gemini example pipelines (r8) + ct2-7c persona/room facets (dispatch_facets/remember/forget FINAL) + persona_outbox + room_say (r16/r20) + room_react (r21). compose_tools('persona')=[room_react,room_say]; **16's on_one_shot persona-% arm auto-verifies a persona-turn (cross-batch proof, on_one_shot NOT re-authored — the B5/17 note honored)**. r18/19 max_tokens→16000 folded; overlay = librarian/codewright/gamemaster room_react grants; persona deny study_*→doc_*.
  - **18 (`9d9a0f4`)** scheduler — `18-scheduler.sql`: cron scheduled_pipelines (pe6 engine + pe7 fire/watchman-tick FINAL). cron parse + e2e dispatch + D-PE4 missed-window all green. ai-news-7am operator seed → overlay.
  - **19 (`addeee8`)** models — `19-models.sql`: model_capability + model_usable + auto-probe (m1/m4/m5/an1) + **work_item_dispatch_stage FINAL** (r3 = J.8.a 4-layer + M.2 capability-substitute + J.11 spend-cap + R.3 max_tokens). Dispatch capability-substitution e2e + max_tokens green. ALL model seeds incl zen1 Claude catalog → overlay; core defaults usable+openai.
  **NEXT = B6** (tests/ re-author + CI day-one + .gitattributes + rename-map.tsv finalize + overlay re-author against doc_*/relational/config-keys + anatomy-doc update) + classify the 20 live↔repo mismatches (verify-suite) + **B5-tail** (seed_harness genericize + bgworker `_kind` enum — schema.rs/Rust-side, NOT authored-SQL). Then the **CUT** (Hinge ① stop live stack + move personas, ③ data-import confirmation) + the **coder wave** `20-coder.sql` (Hinge ② public-ship nod after hardening review).
- **AUTHORING LEG B4/16 SHIPPED 2026-06-13 (OSS `4ba752d`, pushed) — B4 COMPLETE; the consolidated chain runs 00→16:**
  `16-subagents.sql` = sub-agent delegation + the §7.3 self-editable base prompt.
  l9 depth-cap(≤2) + k4 spawn_subagent (**'scripture-study' fallback → config
  default_intent_slug**) + es8 consult + es10 grant + r11 on_one_shot FINAL + ct2-5
  autotag/context_resolve_handle FINAL + ct2-7e (self_prompt_on → propose→critic→ratify
  surface + **compose_tools FINAL**, deferred from 15b). lib.rs: create_subagents
  requires create_context_surface. 7 files retired; manifest 46→39; ext dir 57 .sql;
  secret-scan clean; Go unchanged. Virgin smoke FULLY GREEN (pgcrypto absent; no
  scripture-study hardcode; **depth cap raises@3 / allows≤2**; spawn at root
  origin=agent_planning/cap=500000; **INERT** — propose hidden non-flagged, shown
  w/both-flags, context_* gated; **propose happy-path** session→smoke16-sp→proposal
  pending + prompt-critic work_item; ct2-5 id resolution; es10 22 families minus
  prompt-critic w/ deny-* intact). **Deviations (act+report):** ① **es10 placed BEFORE
  ct2-7e** → prompt-critic (tools-disabled) stays tool-free (★FLAG 20-mismatch: core
  coverage = pipelines-thru-15b, benign superset; live may differ). ② **r11 = on_one_shot
  FINAL here** (manifest line 42, chronological last, true superset of r7/r8) → ★**B5/17
  must NOT re-author on_one_shot — r7/r8's versions are DEAD; 17 only authors the persona
  agent/pipelines/deny-***. ③ context_resolve_handle FINAL = ct2-5 (re-author over 15b's
  ct2-3, +tags fallback). ④ compose_tools FINAL authored here (self_prompt_on first per
  LANGUAGE-sql CREATE-time validation; no later redef — grep-confirmed). Blueprint
  `<pending-16>`→`4ba752d` rides the B5 commit.
  **NEXT = B5** (17-personas: r7/r8/ct2-7c/r16-r21 · 18-scheduler: pe6/pe7 · 19-models:
  j8a/j11/m1/m2/m4/m5/r3/an1/zen1 + dispatch-final j8a+j11 + j7-dispatch + seed_harness
  genericize + bgworker _kind enum), then **B6** (tests/+CI+rename-map finalize+overlay
  re-author). Leg-close: classify the 20 live↔repo mismatches.
- **AUTHORING LEG B4/15b SHIPPED 2026-06-13 (OSS `13cb0f5`, pushed):**
  `15b-context-surface.sql` = the context-engine RUNTIME surface.
  compose_messages FINAL (ct2-7a2, self-contained — ct2-2 base folds
  k2→l13, +§7 self-notes) + CT2 state model(ct2-1)/levers/self-notes(ct2-7a)/
  working tags(ct2-7d, FINAL context_pressure_line w/ tag echo) + judge-brief
  path (es7 minus extract_engrams[15a-owned]: dispatch/render/apply + trigger +
  intercept FINAL + l23 trigger + tool_dispatch_complete_waiting FINAL) +
  intercept_threshold_chars(l22) + read_overflow_raw(l23) + l8 tool_name+wrap +
  l7 suspect-sources + l6 wrappers + deep_research(k5) + chat_post_internal
  FINAL + caps(l30/l31/l32) + 5-arg dry_run(l25) + work_item_cancel cascade(es1).
  24 files retired; manifest 70→46; ext dir 63 .sql; secret-scan clean. Virgin
  smoke FULLY GREEN (pgcrypto ABSENT; 38 kept/0 dead/5 triggers; compose
  system-first; self-note{global}; tag stamp+echo; **judge intercept e2e** —
  62.4k msg→built-in-sha256→overflow parent→judge wq→[JUDGE-PENDING]→K.1 skip);
  GOWORK=off build+vet green. **Deviations (act+report, all in blueprint):**
  ① **es7 sha256 swap** = correctness fix (pgcrypto digest()→built-in sha256();
  ONLY pgcrypto use, dropped; vector-only virgin would've errored at runtime).
  ② **compose_tools FINAL deferred to 16** — true final is ct2-7e (calls
  self_prompt_on, a CREATE-time sql dep born there); schema.rs base carries;
  tool ROWS registered in 15b. ③ OMIT dead judge_templates+render_judge_surface
  + l23 [CORPUS-INDEXED] trigger guard → ★FLAG 20-mismatch (live may carry).
  ④ 3 within-chain finals re-authored (tool_dispatch_complete_waiting 05→es7,
  work_item_cancel 04→es1, chat_post_internal 04→l32). ⑤ doc_* wrapper renames
  (FIRST rename-map rows; Go handlers in lockstep; workspace `45cc5fd`).
  **NEXT = B4/16** (`16-subagents.sql`: k4[slug→config]/l9/es8/es10/r11/ct2-5/
  **ct2-7e — incl compose_tools FINAL + self_prompt_on**), then B5(17-19)/B6.
  Blueprint `<pending-15b>`→`13cb0f5` rides the 16 commit.
- **★ P1 EXTRACTION UNDERWAY (kicked off 2026-06-12, Michael's "Lets kick off P1!"):**
  (1) `github.com/cpuchip/pg-ai-stewards-workspace` (PRIVATE) created at
  `projects/pg-ai-stewards-workspace/` — skeleton + covenant/intent overlay
  copies + 241-file classification (`overlays/classification.tsv`: 191 core /
  17 core-p2 / 27+1 overlay / 5 mixed / 1 scratch) + 33-entry overlay manifest
  + all overlay migrations populated. (2) OSS extension layer extracted
  (`3d8229d`): src/*.rs audited, lib.rs chain reworked (4 seed embeds removed),
  189 core + 5 SPLIT migrations, 193-entry core manifest, bundle = build
  artifact (never checked in). **Build GREEN + virgin CREATE EXTENSION proven**
  (scratch container, 0 workspace seeds leaked) → OSS pushed through journal.
  **COUNCIL (same evening, all ratified):** ct2 RETIRED live · ledger
  leave-and-map · seed pack one-lineage (jumpstart kit canonical) ·
  **doc_*** (study_* tools → doc_*, studies → docs, scripture_anchor →
  values_anchor) · **cutover = FRESH REBUILD** (no shims; selective import;
  live volume archived; rename map at workspace parity/rename-map.tsv).
  **EVENING COUNCIL (all Michael-ratified):** doc_* · fresh-rebuild cutover ·
  six rebuild lessons (early mismatch classification, verify→tests/, _kind
  enum, stewards.config, CI day-one, backup+offsite WAL tiers) ·
  compact_context PULLED IN (hold lifted) · **drop AGE** (relational edges;
  N-depth + BUILDS_ON lineage; fast-at-scale + tenancy conditions; prior art
  verified incl. gospel-engine itself) · **consolidated authored chain**
  ("dave wins"). All in extraction-plan.md.
  **DAEMON LEG SHIPPED (`3561cec`):** five binaries (bridge, stewards-cli,
  persona-host, fs-read-mcp, stewards cockpit) → ONE module
  github.com/cpuchip/pg-ai-stewards; go.work knot dead; build+vet+smoke
  green. Local builds need GOWORK=off (nested clone; strangers unaffected).
  **★ STEWARDSHIP GRANT (Michael, 2026-06-12 night, recorded in
  extraction-plan §Stewardship grant):** full P1-P2 build + migration under
  agent stewardship (act/act+report). Hinge list (still his): ① the CUT
  itself (live stack stop + persona moves) ② coder-mcp public-ship nod
  after hardening review ③ 30-sec data-import confirmation at cut
  (default: corpus/covenant/intent/yt import, histories archive).
  compact_context defaults = his sketch (between-turn, judge-pattern,
  cheap compactor). OSS persona keys: self-service attempt, ping if gated.
  **AUTHORING LEG ACTIVE:** blueprint at
  `pg-ai-stewards-oss/.spec/proposals/authoring-blueprint.md` (consolidation
  map, rename rules, batch plan B1-B6, core=100%-bundle decision).
  **B1a SHIPPED (`3602500`):** 00-config.sql (stewards.config + seeds) +
  01-graph.sql (nodes/edges + recursive-CTE walks) in the bundle chain;
  virgin boot + CYCLE-TERMINATION + bidirectional/lineage walks all proven
  on scratch. rename-map.tsv seeded in workspace repo (parity/).
  **B1b SHIPPED (`ed0da94` + workspace `22e5ea1`) — B1 COMPLETE, AGE IS
  OUT OF THE IMAGE:** create_studies→create_docs (6a + h3-1-docs-half
  ABSORBED into the table: file_path nullable, tags/source_type/
  project_association; kind default 'doc'); 02-workstreams.sql re-authors
  2-6a/b/c relational (context_for = ONE recursive CTE; context_for_hop +
  ensure_studies_graph DELETED; todos parent kinds lowercased
  workstream|doc|todo, 'Phase' retired); resolver/similarity/doc_show
  renamed + relational (doc_similar pure SQL); Dockerfile stage-2 AGE
  build DELETED (runtime = plain pgvector); doc_* swept through ALL chain
  + replay files AND Go daemons (MCP tools study_search/get/similar/
  citations→doc_*; doc_history found by virgin assertion sweep);
  rename-map grew ~27 rows. VERIFIED: virgin CREATE EXTENSION with age
  NOT AVAILABLE (0 in pg_available_extensions), 0 study% functions,
  import/citations/declared-edges/todos/phases/context_for walk/doc_show/
  doc_search/doc_get all smoke green; go build+vet green (GOWORK=off).
  Blueprint gaps fixed: h3-1 mapped (work_items half → 04), 6a removed
  from 04 sources; audit notes in blueprint (parse_gospel_links
  genericization, embed-config at B5, watchman study_id cols at B2,
  l6 wrapper names at B4).
  **B2 IN FLIGHT (2026-06-12 evening):** 03-watchman SHIPPED (`80c9f4c`):
  six files → one, verdicts/findings study_id→doc_id (+related_doc_ids,
  3 index renames, MCP field doc_id), tables born complete,
  estimate_chat_tokens reads config chars_per_token_default, harvest
  trigger e2e on scratch. 04-work-items SHIPPED (`d1d74ef`): ten files →
  one (3c1/3c2/3c2-5/3c3/3c3-1/3c3-3/3c3-5+5e4§1/i1/i2/i5);
  work_item_promote_to_STUDY→_to_DOC, flag-driven
  (pipelines.promote_to_doc — overlay must set it on study-write*),
  last-stage generic, back through import_doc (CITES sync restored);
  chat_post_internal marker fix + tool_defs budget cols +
  agent_tool_perms.source born in schema.rs; i3+h3-followup-2
  REASSIGNED→B3 (08/10 per blueprint); i5 pulled forward; lib.rs had
  NON-LINEAR requires edges (4b, 5a) — sweep for them on every chain
  cut. Full lifecycle smoke green on virgin scratch (template render →
  auto-advance → auto-dispatch → promote w/ graph sync → sabbath gate
  refusal). Gotcha: virgin work_item_create needs a seeded intent
  (hardcoded 'scripture-study' fallback — B3 09-intents wires
  config.default_intent_slug).
  **B2 COMPLETE (2026-06-13 early am):** 05-mcp-bridge `c4ed606`
  (3e2-1/2/3 + h1-5a soft-fail final + h1-7a self-surface seeds w/
  DO NOTHING; waiting_for_tools born in schema.rs work_queue CHECK;
  fan-out completion e2e on scratch). 06-cost `e49ec38` (machinery
  only — ALL operator seeds → workspace overlay
  seed-4a-cost-escalation-models.sql; record_cost_event single 11-arg;
  cost/escalation cols born in 04; j11-dispatch + j12-brainstorm
  halves trimmed in place for B4-14). 07-steward `4d7a715` (steward_tick
  6c-final w/ lessons + atonement-on-quarantine, 6c pulled forward;
  dispatch born 3-arg in 04; provider fallback de-hardcoded to NULL;
  4d stage_models seeds → overlay; live-fire tick smoke green). Final
  sweep: 0 study% fns, 0 study_id cols, AGE absent, Go green. 28
  historical files dead this batch; manifest 189→155 effective.
  LESSON: lib.rs requires-graph is NOT linear — sweep every chain cut
  (4b/5a edges bit once).
  **B3 COMPLETE (2026-06-13, OSS `737443e` + workspace `9a4456d`; root
  lane NOT pushed):** 08-gates/09-intents-covenants/10-sabbath-atonement/
  11-trust/12-council authored; virgin scratch smoke FULLY GREEN (AGE
  absent · 0 study% fns/cols · values_anchor + file_enqueued_at renames
  clean · 15 tables/9 gate_prompts/5 triggers · gate ladder + trust gate
  (trainee surface→journeyman advance) + l28 veto + verify-fail + the
  **08→10 on_maturity_verified materialize path e2e** (sabbath wrapped→
  NOTICE, enqueue_work_item_file real pwid=1, REVIEW-strip extracted body,
  pending_file_writes landed) + sabbath gate refusal + bishop_eligible).
  GOWORK=off go build+vet green. 32 historical files retired; manifest
  155→123. **Deviations (act+report, in blueprint):** apply_gate_decision
  authored ONCE in 11 (its trust SELECT needs trust_scores — a plpgsql
  SELECT-from-later-table is NOT a safe CREATE forward ref; only NEW.<field>
  + wrapped fn-calls are, per the 04 precedent); maybe_enqueue_atonement +
  sabbath/atonement dispatch finals → 10; **h1-0 FULLY consumed at B3**
  (maturity_ladder→08, overrides→10) — drop from B4's 13; 6e SPLIT (lesson
  producer→10, resolution producer→12 — %ROWTYPE/trigger on a not-yet-born
  table fails at CREATE); 5d5 gate tools_disabled finals folded into 08;
  sessions.kind union + gate_prompts CHECK born in schema.rs/08; yaml.rs
  slug-from-YAML(default "default") + values_anchor.
  **★ SURFACED TENSION (Michael's call, NOT fixed):**
  `work_item_promote_trigger` (04, B2) calls work_item_promote_to_doc
  UNWRAPPED → on a sabbath-enabled pipeline a status→completed transition
  ABORTS until sabbath_completed_at is set (the gate RAISEs check_violation).
  Conflates "defer promotion" with "block completion"; likely wants the
  PERFORM wrapped (mirror on_maturity_verified). Faithful to historical
  authoring, not introduced by B3 (smoke confirmed the abort).
  **B4 IN FLIGHT.** **B4/13 SHIPPED 2026-06-13 (OSS `97f42db`):**
  research-write (4-stage, h2 final) / planning (5-stage) / agent-proposal /
  revise-proposal / research-summary seeds + enqueue_proposed_work_items +
  apply_agent_proposal (i7 final, i6 gate folded) + apply_revision. Virgin
  smoke fully green; go build+vet green; 13 files retired, manifest 123→110.
  Deviations: h1-0+h3-1 already consumed (dropped); h-ledger-1 schema_migrations
  table → **00-config** (bundle births it; empty runtime manifest); on_maturity_verified
  NOT touched (08 single final; agent-proposal+fanout branches fold into 08 at
  B4 close — its TRUE final is j7); apply_agent_proposal single i7; dispatch
  tools_disabled deferred to 19; genericized gospel/personal-project text.
  **B4/14 SHIPPED 2026-06-13 (OSS `b1a9b01`):** fan-out machinery + 12-lens
  brainstorm + catalog_default_* helpers + one-shot/child-terminal triggers;
  on_maturity_verified TRUE final (j7) folded into 08 (late-bound forward
  refs to 13/14); dispatch-final (j8a 4-layer + j11 cap) DEFERS to 19 (j8a/j11
  KEPT in manifest); j8b→lens defs; j6 supersedes j2; start_brainstorm
  scripture-study→config. ★ spawn_children = UNION of j3+j4+j8c — **j8c (last
  live redefinition) dropped j3 aggregator + j4 per-child file_destination**
  while adding override propagation; restored here, FLAG for 20-mismatch
  classification. Virgin smoke green; go build+vet green; manifest 110→97.
  **NEXT = B4/15** (context-engine: k1-k9/l1-l32/es1-es9/ct2 — the biggest;
  engrams/rendering/judges/circuit-breakers; may split 15a/b; watch es7
  judge-gate, es1 cancel-cascade, l6 investigate_study→doc_* renames) + B4/16
  (k4/l9/es8/es10/r11/ct2-5/ct2-7e). Then B5(17-19)/B6.
  [archived 14-source detail; see blueprint] j1-j9c incl j8a-dispatch + j11-dispatch-
  gate + j12-brainstorm TRIMMED HALVES (left in place at B2); 15 = k1-k9/
  l1-l32/es1-es9/ct2-1/2/3/7a/7a2/7b/7d (may split 15a/b); 16 = k4/l9/es8/
  es10/r11/ct2-5/ct2-7e. Same loop (sweep lib.rs non-linear requires +
  forward-ref shapes on every cut). Then B5 (17-19 + seed_harness genericize
  + bgworker _kind enum + embed provider/model→config) + B6 (tests/ + CI +
  rename-map finalize + overlay re-author).
  (3) Private manifest REPAIRED (root `e5ccc0c3`): 9 live-applied migrations
  (r11-r17, ct2-5, ct2-7e) restored from ledger order; found the runner is
  LEXICAL + manifest-blind (replayed scratch-ct2-run2 into live 06-10 —
  codewright-ct2 rows; disposition = Michael's call).
- **pg-ai-stewards OSS extraction** (continues the `pg-ai-stewards-oss` lane —
  same session, retitled): spec RATIFIED, Apache-2.0 FINAL (`3c43d4e`).
  **"Anatomy of a Turn" SHIPPED (`0e8c3c9`)** + order-research update +
  2026 regrounding (`1a604af`). **Cutover gate AMENDED (`8662448`,
  ratified 06-12): FULL PARITY before the cut** — coder-mcp + UI become
  pre-cutover (P2 before cut), 20 mismatches + ledger wart now on the
  cutover critical path. Next: P1 extraction (task #151), side-by-side
  compose (`stewards-oss-*`, 55434/8081/8091). Overlay design ratified
  (`0e01a04`): private repo pg-ai-stewards-workspace, created at P1
  kickoff. jumpstart-crossover reflection seeded (`48864a47`, no build).
- **PR.1 SHIPPED + live-verified 2026-06-12** (inbox assignment, Michael's
  "best of your judgement" grant): covenants.extensions catch-all +
  presiding render + Watch echo; reseed through the real path; smoke
  `600f6673` ACK with presiding terms in the dispatched payload. Journal:
  `projects/pg-ai-stewards/.spec/journal/2026-06-12-pr1-covenant-extensions.md`.
  Carry-forwards there: walls-vs-compulsion audit (§V), trailing-reminder
  proposal, verify-suite full run, ledger naming wart.
- **compact_context SEED captured** (Michael's sketch, 2026-06-12 — HOLD,
  no build until council): commissioned-curation side quest; seed at
  `projects/pg-ai-stewards/.spec/proposals/substrate-compact-context-sidequest.md`
  with parked council questions. 2026 research also on hold per Michael.

## Claims
- NONE live. (PR.1 window CLOSED 2026-06-12: pg+bridge rebuilt/restarted,
  watchman resumed, queue clean, live smoke verified. Persona-host
  container was never touched.)
- The general-workspase lane owns the containerized persona-host
  (acknowledged; will not restart it).

## Handoffs / notes
- 2026-06-12: Anatomy doc is public — sibling sessions citing substrate
  architecture can link github.com/cpuchip/pg-ai-stewards/blob/main/docs/anatomy-of-a-turn.md.
- Supersedes lane file `pg-ai-stewards-oss.md` (same session_id; hook
  re-claimed under the new title).
