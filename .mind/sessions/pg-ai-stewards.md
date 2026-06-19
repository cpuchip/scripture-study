---
lane: pg-ai-stewards
session_id: 7ea7faa4-688a-451a-ac68-b7ea662d4b81
status: active
started: 2026-06-11T22:00:16
last_active: 2026-06-18T16:11:00
---

## Working on
- **★ 2026-06-18 — LOCAL DUAL-4090 INFERENCE via FlexLLama: PROVEN + WIRED + 3-WAY VERDICT.** Record = `projects/pg-ai-stewards-workspace/research/local-inference-flexllama.md` (nothing committed to OSS; `oss/.env` flexllama block is a local gitignored config change). Michael wanted a multi-GPU server better than LM Studio (LM Studio can't pin a model→GPU / budget VRAM). Cloned `yazon/flexllama`→`external_context/flexllama` (gitignored); built CUDA image trimmed to **sm_89** (`flexllama-gpu:latest`); **per-runner `CUDA_VISIBLE_DEVICES` pins one model per card**; shares `C:\Users\cpuch\.lmstudio\models` GGUFs read-only (no copy). **★ CONTEXT CEILINGS (1×4090, Q4, flash-attn, --parallel 1):** qwen3.6-27b **~128k** (fills card; native 256k); **gemma-4-12B-QAT FULL 256k @ 12.3GB** (native 256k; KV ~3GB/128k — the long-ctx champ); nemotron-3-nano-4B native **1M**, 204 tok/s. **PAIRS on ONE card:** gemma256k+nemo512k f16 (23.9G); gemma256k+nemo **FULL 1M** with `q8_0` KV (23.9G). Both-GPU concurrent 27B+gemma ≈ **136 tok/s, 0 errors** (E1 500s gone — dedicated cards + llama.cpp continuous-batching). **★ WIRED into LIVE substrate:** `flexllama` provider in `oss/.env` (`host.docker.internal:8090/v1`, openai, no key — mirrors OLLAMA), recreated **stewards-oss-pg --no-deps --force-recreate** (data-safe, pgdata persists; the bgworker reads provider env at pg startup), `providers_loaded()` shows it, `model_capability`+`$0 pricing` for qwen3.6-27b/gemma-12b/nemotron-4b (usable, trains_on_data=false). **★ E2 PASS (tool loop):** E2a raw-OpenAI tool_call — both gemma+nemotron emit+synthesize; **E2b REAL substrate gather — gemma-12b drove the multi-round web loop (web_search_exa/fetch_url) → correct items brief → completed, $0, tokens_in=258k cumulative.** **★ 3-WAY VERDICT (same research-summary gather, each forced local via work_items.model_override+provider_override):** gemma-12b ✅ rich 3.8k-char brief = **GATHER/doer winner**; qwen-27b ❌ HTTP-400 ctx (69k>65k at n_ctx 64k; needs 128k = its whole card) → **CRITIC/COORDINATE/SYNTHESIZE** not raw gather; nemotron-4b ⚠ drove loop but `output=[]`/never converged (1 tool call, 76 out-tok) → **FAST PERSONAS only** (Michael's instinct confirmed). **★ CONFIG LESSON:** this llama.cpp build = `--ctx-size` is **PER-SLOT × default --parallel 4** → set `--parallel 1` + size `n_ctx` to the CUMULATIVE loop (gather is token-heavy; 32k gemma + 64k qwen both 400'd mid-loop, 128k gemma sailed). **NEXT (Michael's plan, not started):** the **1-2 punch = gemma doer → qwen critic → gemma revise** via per-STAGE models (book-digest critique / research-write review are the hosts; `work_item.model_override` is all-stages so use stage `model`/`stage_models`); per-model "special instructions" tuning flagged. **★ CLEANUP TODO (live DB test cruft):** `lm-eval` intent+pipeline (earlier experiments) + work_items `e2b-gemma-ainews`/`e2b-gemma-2`/`e2b-qwen`/`e2b-nemo` — purge when done; flexllama model rows = KEEP if adopting. **Container `flexllama-stewards` UP holding BOTH 4090s**; `docker rm -f flexllama-stewards` frees them for LM Studio. **★ MAX-VRAM proven** (`stewards-max.json`: qwen 65k GPU0=20.5G + gemma 256k & nemotron **full 1M** q8_0 KV GPU1=23.9G). **★ STRICT-TEMPLATE BUG FIXED + SHIPPED (OSS `main` `258eaea`, applied live):** `compose_messages` relabels any mid-history system row (the soft-cap `[STEWARD NOTICE]` from `build_soft_cap_notice`) → `user` in place, so the array has exactly one system msg at front (Qwen3.6's template requires system-first → 400'd; gemma/nemotron tolerated but it's semantically weak for them too — Michael's instinct). Inverse-proven: qwen@65k gather that 400'd now runs clean. Follow-up: virgin-smoke assertion. **★ "punch above context weight" VERDICT = NO for gather:** qwen@65k survived 317k cumulative (reactive engine folds) but EMPTY brief + 0 agent context-tool calls (#136 reaffirmed: reactive floor carries it, models don't self-manage even when told); gemma@32k 400'd FIRST round (37.6k>32k — irreducible web-gather working set ~35-65k); gemma@128k = rich brief. Lever = native context matched to job (gather ≥64k; critique fine small). Substrate gap: context engine defaults flexllama window ~200k (doesn't know real window → can't fold to fit small). **CLEANUP TODO grows:** live work_items e2b-gemma-ainews/e2b-gemma-2/e2b-qwen/e2b-nemo/e2c-qwen-ctx/e2d-qwen-fix/e1-gemma32k + lm-eval intent. **NEXT unbuilt payoff: the 1-2 punch** (gemma doer → qwen TOOLS-OFF critic → gemma revise; per-STAGE models on book-digest/research-write; qwen-as-critic is tools-off so the template bug never bites it). **★ "ROOM TO BREATHE" BRAINSTORM → 1+2 SHIPPED, 3 PROPOSED (OSS `d48bc5b`, applied live):** Michael's framing — context window = input + reasoning + output sharing ONE budget; a near-full window starves generation. **#1 soft-cap notice now gateable** (chat_post_internal, config `soft_cap_notice_enabled` default true; **set FALSE live** per Michael — compose_messages already relabels it harmless, this kills it). **#2 WINDOW-AWARE BUDGET** (`effective_budget` 15a Layer 2.5 + new `model_capability.context_window` col in 31): flexllama had NO provider_rules row → fell to the **64000 fallback** (gemma over-folded to 64k despite 256k; qwen got ~full-window budget → no room → empty output). New layer: model window×**0.70** (Michael's 30% reserve) when context_window set; paid providers unchanged. Live-proven gemma→183500, qwen→45875. **★ HONEST: #2 is correct + improves gemma but NOT sufficient for small-window gather** — qwen@65k STILL 400'd (at 65,750) because compose_messages preserves the fresh ~8-msg TAIL raw, so a single big fresh fetch (~40k) can't be folded → blows a 65k window in one round. **= empirical proof #3 is the real fix.** **#3 PROPOSAL `page-in-large-results.md`** (dominion_in_council): stop inlining big fetch/doc results — store + summary + handle + read-slice tools (result_read/result_search), model-chosen retrieval; reuses judge-brief intercept + context_search + expand_message + fs_read offset/limit (~70% built); helps small local models AND cuts PAID token cost. **VERDICT: gather wants native ctx that holds a round's ~35-65k working set (gemma 128k); small windows need #3.** Template fix `258eaea` + this `d48bc5b`. Follow-up: virgin-smoke asserts for the relabel + the budget layer. **CLEANUP TODO += e2c/e2d/e2e/e2f/e2g-qwen + e1-gemma32k work_items.** **★ #3 PAGE-IN BUILT + SHIPPED (OSS `e1bc8e2`, new `33-page-in.sql` + compose_messages wrap + lib.rs/Dockerfile; Michael "lets build number 3" = the council go):** `page_in_cap` truncates any single rendered message over `effective_budget * page_in_single_msg_ratio` (0.5, window-aware) to head + a [page-in] banner with the handle; new sql_fn tools `result_read(handle,offset,limit)` + `result_search(handle,query)` (own+watch, reuse 27 resolver) granted research/dev. **MECHANISM PROVEN (10x): a qwen@65k gather cut ~15 raw fetches from ~600k→70k tokens.** **★ BUT HONEST: per-message capping does NOT enable small-window gather** — qwen@65k STILL 400'd (70,927) because ~15 capped messages sum past 65k; a per-message cap can't bound the TOTAL of a many-message loop, and the async torso-folding doesn't keep pace in a fast loop. The COMPLETE small-window fix = a **total-render cap** (drop/summarize oldest until rendered total ≤ budget) = **P1, not built** (chases qwen@65k gather which is a NON-GOAL — gemma 128k does gather). So #3 P0's real value = cap a pathologically-large SINGLE result + paging tools + cost-trim; it does NOT make small models gather. **VERDICT (4x-confirmed): gather needs native ctx that holds the ~15-msg loop (gemma 128k); qwen = tools-off critic.** Proposal `page-in-large-results.md` (P1 total-cap noted). Follow-ups: P1 total-render cap (only if small-window gather ever becomes a goal) + virgin-smoke asserts. **★ SESSION = 3 OSS commits: `258eaea` strict-template fix, `d48bc5b` window-aware budget+notice-gate+proposal, `e1bc8e2` page-in.** **★★ CORRECTION (Michael pushed back, RIGHT): qwen@128k IS a strong LOCAL agentic coordinator — I overclaimed "qwen tool-loop unreliable."** Every qwen failure was window/template/budget, NEVER tool-calling. Loaded qwen at its real **128k** (fits a card, 24GB) + template fix + window-aware budget → drove the full multi-round web gather: **err=NONE, 8342-char brief (RICHER than gemma's 3802), 270k cumulative, reasoned WHILE gathering** (excluded stale items by checking dates). So **qwen-27b@128k = local coordinate/reason/critic (free, no-train) → kimi paid fallback** (only for >128k or heaviest agentic). **The 3-MODEL DANCE casting:** ingest (hold raw)→deepseek/nemotron/gemma; coordinate+reason+critic→qwen@128k LOCAL FREE→kimi. **★ The architecture is already there:** stages pass COMPACT `{{stage_results.X.output}}` briefs forward (raw stays in the gather session), so reasoning stages get small inputs that fit any window; research-summary already casts review→qwen tools-off. **NEXT (offered): role-aliases** (`ingest`/`reason`/`critic`/`coordinate` via the 31 alias system, members = the fleet) so pipelines reference ROLES + the router picks best-available + fails over; recast digester/research pipelines. = the real payoff; makes Vivint analysis fully local+free+private. Findings doc has the full casting + correction. **★★ DANCE BUILT + PROVEN E2E + SHIPPED (ws `276ea52`): role-aliases `ingest`(gemma→nemotron→kimi)/`reason`(qwen→kimi)/`critic`(qwen→kimi) + flexllama-models overlay (3 models, no-train, $0, ctx 131072) + research-summary recast to roles.** Dispatched research-summary NO-override → **gather=gemma(7 tool rounds), synthesize=qwen, review=qwen — ALL LOCAL, $0** (each role auto-resolved per stage; handoff = compact `{{stage_results.X.output}}`). Vivint analysis can now run fully local+free+private. ★ context_window in flexllama-models.sql MUST match the loaded FlexLLama config (stewards-qwen128.json = all 128k); update if runtime changes. **★ CLEANUP DONE:** purged 23 test work_items + 31 sessions + 250 msgs + 151 wq rows + the lm-eval intent/pipeline/agent (transactional). **SESSION TOTAL: OSS `258eaea`+`d48bc5b`+`e1bc8e2`, ws `276ea52`; FlexLLama on **stewards-dance.json** (canonical, both 4090s maxed ~24GB each: qwen 128k f16 GPU0 + gemma **256k f16** + nemotron **512k f16 KV** [Michael: 4B small → f16 quality > extra q8 window; 512k is the f16 max alongside full gemma; verified 24.0GB GPU1, 87 tok/s] GPU1; windows aligned live+overlay). **★ REPRO SCRIPT: `projects/pg-ai-stewards-workspace/scripts/setup-flexllama.ps1`** (ws `1f75cc4`) — one-shot: clone+build FlexLLama (arch-trimmed), write the dance config, mount LM Studio GGUFs RO, launch+wait; idempotent (-Rebuild); header documents the substrate prereqs (oss/.env provider + flexllama-models/role-aliases overlays). canonical config = `external_context/flexllama/stewards-dance.json`.** Follow-ups: P1 total-render cap (only if small-window gather becomes a goal); virgin-smoke asserts (relabel/budget/page-in); recast book-digest/playlist/planning to roles; deepseek as an `ingest` member.
- **★ 2026-06-17 — NVIDIA FREE PROVIDER WIRED + PROVEN e2e (OSS `da03dfe`, ws `f01afc1`; task #185 in-flight, guard-rail+routing remain).** Michael found build.nvidia.com free preview endpoints (OpenAI-compat; **trains on data → PUBLIC lines ONLY**). Tested raw: `moonshotai/kimi-k2.6` HTTP200/~0.9s + **tool-calling YES** (digesters need it); `qwen/qwen3.5-397b-a17b` returned empty/0-tokens (quirk — kimi covers the doer). Wired: `oss/.env` `STEWARDS_PROVIDER_NVIDIA_{BASE_URL,API_KEY,KIND}` (gitignored; env_file feeds pg+bridge). **★ GOTCHA: chat dispatch is the IN-PG bgworker (`providers.rs` from_env, reads env at pg startup, never reloads) — NOT the bridge.** First recreated the bridge → still "unknown provider: nvidia"; recreating **stewards-oss-pg** (env_file reload; pgdata volume persists, data intact) fixed it. model_capability+pricing **$0** for kimi-k2.6/minimax-m3/deepseek-v4-flash (`overlays/nvidia-provider.sql`). PROVEN: a work_item forced to nvidia/kimi-k2.6 completed e2e, output correct, $0. **★ KEY DESIGN REALIZATION: provider choice must be INTENT-aware, not per-pipeline** — research-write/planning are SHARED by public + vivint intents, so routing those pipelines to nvidia would leak vivint to a train-on-data endpoint. So the **file_private guard rail (refuse nvidia for private intents) is the KEYSTONE**, not just defense-in-depth. **REMAINING:** (1) guard rail [core dispatch — refuse/avoid nvidia when intent.file_private]; (2) route public digesters → nvidia free (book/playlist single-intent pipelines safe immediately; research/planning gated by the guard rail) = off opencode-go spend. NVIDIA model ids carry org prefix (moonshotai/, deepseek-ai/, minimaxai/). (Board megaline not extended this turn — saturated + sibling-active; this lane + the live state are the record.)
- **★ 2026-06-17 — DIGESTER EMPTY-SOURCE HALT FIXED + LIVE (OSS `244ae38`/`6407e42`, pg18; task #184 DONE).** From the morning "how many runs overnight" dig: book-digest ran ~18× overnight (3 REAL digests — no junk, content-gate holds; 3 SHELF-EMPTY no-ops that ran all 4 stages; 2 13h-zombie Moonshot hangs → I cancelled them). **Root cause (reproduced via work_item_advance):** the per-pipeline BEFORE-UPDATE empty-guard set status=cancelled, but `work_item_advance` STILL RETURNED the next stage name + the bgworker dispatches off the RETURN → cancel and return disagreed → 4 stages ran. A BEFORE-trigger can't win that race. **Fix:** halt moved INTO `work_item_advance` — `metadata.halt_on={stage,outputs[]}`; on sentinel → cancel + RETURN NULL (no dispatch). book-digest/playlist-digest declare halt_on; 2 triggers retired. virgin-smoke OK17 + clobber 3/0 + **live inverse-hypothesis proven** (SHELF EMPTY → NULL/cancelled/stays-read). Lesson: a status-cancel that doesn't also halt the RETURN value loses to the dispatcher; halt at the choke point, not in a racing trigger. = digester-empty-source-halt SHIPPED.
- **★ 2026-06-17 — A (intent-private routing) + B (tool primers) + pre-commit hook REMOVED (OSS `30a0991`/`5d7133a`, ws `4a9878e`; pg18; tasks #182/#183 DONE).** (1) Removed `.git/hooks/pre-commit` (targeted retired `pg-ai-stewards-dev`; RW /workspace mount auto-materializes now). (2) **A — `29-intent-private-routing.sql`:** `intents.file_private` + a BEFORE INSERT/UPDATE OF file_destination trigger that prefixes `private/<intent_slug>/` (idempotent; catches every stamping site since enqueue_work_item_file re-reads the col). Overlay `intent-private-vivint.sql` marks vivint private. **Vivint drops moved → `private/vivint/{plans,research}/` + root .gitignore safety-net** (`plans/vivint-*`, `research/vivint-*`, root commit `30f384a3`, NOT pushed). Live-verified: `plans/x → private/vivint/plans/x`. (3) **B — `30-tool-primers.sql`:** `tool_primers` table + `render_tool_primers(agent_family)` gated per group (context=context_tools_on, skills=skill perm) + core primers; compose_system_prompt (09, lines 257-468) injects it late-bound. Telemetry drove it: context_search/todo/goal/skill = **0 uses** (models untrained on them; reactive engine fine w/ 8 folds). virgin-smoke OK15+OK16 + clobber 3/0; live-verified (research gets the context primer). **Deploy gotcha:** scratch-container extraction flaked → extracted compose_system_prompt from the repo file block (257-468) instead. **One-shot Q answered:** 57 tools-ON stages (multi-round, coder-pr-style) vs 24 tools-OFF (true one-shot judges); primer aimed at the long-runners (coder/steward/personas).
- **★ 2026-06-17 — GUARD NARROW AUTO-RESUME BUILT + LIVE + pushed (OSS `c1dc09f`/`5d7133a`, pg18; task #181 DONE).** The guard auto-paused tonight on the 24h spend cap (worked as designed); Michael bumped cap $10→$12 + resumed, then ratified "narrow resume" — the guard should release its own brake when a breach self-clears. `28-guard-autoresume.sql`: `reflect_guard_autoresume_tick()` lifts a pause ONLY if guard-set (`reflect_pause_source='guard:<breach>'` marker; human pause='manual' stays) + breach self-clearing (spend|in_flight; failures/proposals stay for a human) + no active breach + metric <75% of cap (deadband). Logs `auto_resumed`. Re-authors reflect_pause/reflect_watchman_tick/watchman_scheduler_fire/reflect_status (carried 22/23 bodies + marker; confirmed 22/23 were latest pre-28 defs). virgin-smoke OK14 + clobber 3/0 + live-verified. **Live spend cap = $12** (Michael's bump). Pause-source-marker pattern is the reusable bit. Journal `2026-06-17-context-search-and-guard-autoresume.md`.
- **★ 2026-06-17 — `context_search` P0 BUILT + LIVE + pushed (OSS `de52a24`/`14fce53`, pg18 rebaked; task #180 DONE).** `27-context-search.sql`: `context_search`(scope session|descendants; curated default + `include_folded` recovery; snippet+`[ctx:handle]`) + `context_session_private` wall (sessions.private, beats the watch, absolute) + `context_descendant_sessions` (work_items lineage, excludes private). `context_*` auto-surfaced (no gating change), sql_fn (no refresh-tools), tool descriptions written to teach. virgin-smoke OK13 + clobber 3/0 + live-verified (snippet+handle, surfaced to research). Schema-investigation notes for future builds: messages has `context_state`(verbatim/muted/compressed/pinned)+`context_tags`; sessions had NO parent link → lineage via `work_items.session_ids`+`parent_work_item_id`; `context_handle(id)`→4hex; `pipelines.stages` CHECK requires ≥1 elem (smoke caught `[]`). **`self` deferred to P1** (no session→agent map). Michael's adoption idea captured in proposal: per-tool-group usage primer + usage telemetry.
- **★ 2026-06-17 — `context_search` RATIFIED (council w/ Michael; OSS `ce79e60`), cleared to build P0.** Decisions: P0 = own (session+self) + descendants (the watch) + a MANUAL session `private` flag that beats the watch (private child invisible even to parent — security primitive for local-non-cloud sensitive work); upward (ancestors) private-by-default + per-msg private + a `sensitive` intent/agent flag (forces local+private) = P1; recall-surface + global = P2; wall is absolute (D&C 121). P0 build = a fn over `messages` + tool_def + private/descendants resolution, in the `context_*` grant family; smoke = folded-only-with-flag + handle→expand_message + parent-sees-normal-child-not-private. NOT built yet (awaiting Michael's build go).
- **★ 2026-06-17 — PROPOSAL `context_search` WRITTEN (awaiting council; OSS `dc5d106`).** Michael's idea: durable messages → give every agent grep over its OWN context + inject into docs (the Ctrl-F a model can't do). Captured his refinements: name `context_search`; `include_folded` flag (recover muted/compressed for re-open — loops w/ productivity auto-fold); scope ladder = D&C 121 walls (own→self→descendants[watch]→ancestors[private-aware]→global[gated]); per-msg `private` flag (parent walls privileged from children; opt-out-vs-opt-in = the council fork); provenance≠truth; snippet+handle results ride the context engine. P0=own-context. dominion_in_council. `.spec/proposals/context-search.md`.
- **★ 2026-06-17 (cont.) — two tunings (Michael):** (a) **qwen3.7-plus PRICE CONFIRMED $0.40/$1.60 per Mtok** (opencode zen, his link — cheaper than qwen3.6-plus's 0.50/3.00; my placeholder was too high) → corrected live + repo (OSS `62f9a27`, ws `82ce3aa`); spend guard now accurate. (b) **book-curate cadence 6h→2h** (`0 */2 * * *`) — book-digest eats ~1 book/hour so 6h drained the shelf; 2h keeps pace (a STOCKED run is one cheap no-add call). Live + repo.
- **★ 2026-06-17 — QWEN MODEL ROTATION (Michael: "rotate those models up… max is most expensive").** `qwen3.6-plus`→`qwen3.7-plus` (newer, same price) across ALL dispatch paths: 25 pipeline stages, 15 stage_models, 6 brainstorm metadata, 3 gate-fn hardcodes (evaluate_gate/verify_work_item/sabbath_dispatch), 5 escalation rules; + `prompt-critic` review `qwen3.7-max`→`qwen3.7-plus` (max = priciest + usable=false). **★ CAUGHT qwen3.7-plus had NO price row** → would blind the watchman spend cap; priced = qwen3.6-plus rate ("similar price" — ⚠ CONFIRM real opencode rate, 1 row live+repo). Catalog keeps old models; factual comments/caveats/verify-* fixtures preserved. Live: direct UPDATEs + 3 gate-fn CREATE OR REPLACE (extracted final bodies). virgin-smoke OK 1-12 + clobber 3/0; pg18 rebaked. OSS `73f7c81` + ws `51534d9` pushed. (Cadence answered: reflect-steward 3h, book-curate 6h.)
- **★ 2026-06-16 (cont.) — DIGESTER-STEWARD P0 (book-study CURATOR) SHIPPED + e2e-PROVEN + pushed (OSS `38f6c6d`).** Council-ratified ("I want this change… enable by default"; "always push back on spend" confirmed-wanted). The missing back-office leg = the reflect-steward generalized to a queue. Built in `examples/book-digester.sql`: `book_shelf_status()` fn+tool, the `book-curate` pipeline (1 tools-on `curate` stage, research agent, kimi-k2.6: runway→survey→pick+**verify-findability**→book_add; dry→start_brainstorm), `book-curate-cron` (6h, enabled), dials (`runway_threshold=3`/`max_adds=5`), research grants. **Proven BOTH branches on live:** feed→3 non-dup intent-aligned books (Nicomachean Ethics/Hume Enquiry/Bacon Novum Organum, Gutenberg URLs I confirmed HTTP 200 — inverse-hypothesis on the verify-gate); restraint→"SHELF STOCKED", added nothing. Snags: tool_defs INSERT was 5-col/4-val (added `active=true`); **Moonshot AI [kimi] DOWN (521/522)** → proved logic by overriding the e2e work-item only to qwen3.6-plus (opencode-go flat-rate ~$0; production stays kimi, retries next tick). Journal: OSS `.spec/journal/2026-06-16-digester-steward-curator.md`. Task #179 DONE. **P1 (ratified, not started):** generalize curator→video-study/ai-news; generic `digester-empty-source-halt` (retires the 2 per-pipeline triggers); slow book-digest hourly→6h to match the curator.
- **★★★ 2026-06-14 PERSONA-LEG + SCRIPTURE-MCP ARC (Michael's direction; IN FLIGHT, big outward-facing infra):**
  - **✅ M6 (tail-bounded) DONE+PUSHED (OSS `15a1084`):** stamp_code_write_sandbox defaults `acceptance_criteria=''` (the M2 stall gotcha) + virgin-smoke OK 6 (compact_context ships complete; chain 00→21). M6-tail REMAINING (bigger, deferred): bgworker surface-not-swallow render errors (Rust); capability-aware model substitution (probe tool-USE not just chat).
  - **✅ PARENT-REPO GITIGNORE HYGIENE FIXED (root commit, NOT pushed — Michael pushes root):** the parent was tracking **9 nested external repos as bare gitlinks** (ai-jumpstart, dnd-tools, md-mcp, pg-ai-stewards-oss, pg-ai-stewards-workspace, scripture-book, spin, spoken-study, strongs-concordance-mcp) → `git rm --cached` (working trees intact) + .gitignore'd all + future (webster-mcp/byu-citations/gospel-engine). Parent now tracks 0 projects/ gitlinks. (projects/pg-ai-stewards [private substrate] is NOT a separate repo — stays tracked.)
  - **★ GOSPEL-ENGINE → REMOTE HTTP MCP: opus dev subagent RUNNING (bg, agentId ae1c2240a26eea361)** — Michael's explicit "do it in a subagent, it'll auto-deploy." Spec: ship gospel-engine as remote HTTP MCP like dnd (dnd.ibeco.me/mcp) so the lean OSS bridge + Claude cowork reach it over http. ⚠ NO SendMessage in this harness — can't live-redirect it; it's extending engine.ibeco.me (NOT creating a public repo, so low corpus-publish risk). **Michael's NEW direction to fold in when it reports:** (a) fix broken `engine.ibeco.me/api/admin/tokens` (token-settings link dead-ends, no way to get an API token); (b) move to a PUBLIC repo in `projects/gospel-engine/` — but **CORPUS STAYS PRIVATE** (gospel-library = church content, gitignored; code public, corpus mounted). Diagnostic seen: gospel-engine-v2/internal/mcpserver uses mark3labs/mcp-go (broken import — the subagent's WIP).
  - **★ SCRIPTURE-MCP LAYER → PUBLIC REPOS + SUBDOMAINS (Michael BLESSED, NOT yet started):** webster-mcp + strongs-concordance-mcp + byu-citations → each (1) move scripts/<x> → `projects/<x>/` as its OWN public GitHub repo (gitignored from parent — done preemptively), (2) add remote HTTP transport like gospel/dnd, (3) deploy to a subdomain (webster.ibeco.me/mcp etc.). These are librarian's toolchain. **Plan = serial-probe-then-parallel-scale: let gospel prove the full pattern (projects/ repo + public + HTTP + subdomain + deploy + verify), THEN fan out webster/strongs/byu copying it.** webster/strongs = public-domain data (clean public); byu = check usage terms.
  - **PERSONA MIGRATION — scoped REALITY (Michael: "pull in all 3 + Callie"):** 3 substrate personas: **codewright + gamemaster EASY** (core tools + dnd is already remote-http); **librarian GATED on the scripture MCP layer above** (gospel/webster/strongs/byu). Persona-host hosts **5 chattermax personas** (chip-assistant, chattercode, dm-assistant, npc-ally, **callie**) — keys in OG persona-host env (`CHATTERMAX_PERSONAS`, cmk_* tokens); their personalities + the chattermax→substrate-agent mapping live in the **remote chat.ibeco.me DB** (no local ai-chattermax container). Cutover = move 5 keys to OSS persona-host .env + apply 3 substrate persona overlay seeds (r9/r13/r14/r15/r17) + **stop OG persona-host → start OSS** (⚠ outward-facing live chat.ibeco.me rooms; double-fire safe ONLY if stop-OG-first; general-workspace lane owns the OG persona-host container — signal it). HELD until scripture MCP layer is remote (so librarian works).
  - **dnd + chattermax OSS EXAMPLE DOCS** (Michael: dnd is public, document so others use chattermax+dnd as examples) — TODO.
  - **ANSWERS captured:** Callie = chattermax persona (not substrate agent). Thummim = NOT an MCP — the "Thummim 2026 Restoration Dictionary" generation pipeline (overlay thm1, schema+3-stage, never dispatched). Embeddings = pgvector storage + local Ollama/LM-Studio nomic-embed at index time (docs/engrams/corpus), not "everything." ibeco.me UP (200); the broken link = engine.ibeco.me/api/admin/tokens.
- **★★★ 2026-06-14 SESSION (Michael fixed the PAT repo-scope; "finish the test… get OSS up and running fully before Sabbath"):**
  - **✅ M2 FULLY CLOSED — DRAFT PR #1 LANDED e2e** (`github.com/cpuchip/pg-ai-stewards-coder-proof/pull/1`, draft, OPEN). Fresh code-pr work item `6aeb7265` ran clone→plan→plan_review→implement→verify→review→pr→push→gh-pr-create on the OSS stack. Token fix verified by a bridge-side dry-run push (exit 0) BEFORE the run. **NEW BUG FOUND (M6): review-verdict parser is start-anchored** — `work_item_advance` cv6 checks `v_verdict_text !~* '^\s*REVIEW:\s*passes'`, but glm-5.1 preambled ("All files verified…\n\nREVIEW: passes") so a genuine PASS misreads as REVISE and parks at awaiting_review. Worked around by applying the verdict in the expected format (`work_item_advance` with output prefixed `REVIEW: passes`); the real fix = make the cv6 AND cv11 (`^PLAN:\s*approved`) matches line-anchored/multiline (`(^|\n)\s*REVIEW:\s*passes`). Also confirmed: review→pr auto-advance after a TOOL-USING review stage doesn't fire (bgworker didn't enqueue the advance) — manual `work_item_advance` is the bridge; both are M6.
  - **✅ M3 DONE + PUSHED (OSS `08850f6`) — the web UI is RUNNING.** Ported `scripts/stewards-ui` (59 files, ~11K LOC Vue+Go, 23 views / **61** API routes — not 23) → `cmd/stewards-ui/` in the single OSS module (no go.work/stubs). **doc_* rename done SCHEMA-VERIFIED (not blind):** studies→docs, `study_search_text`→`doc_search` (non-1:1, drops _text), study_citations/similar→doc_*, verdicts.study_id→doc_id, intents.scripture_anchor→**values_anchor** (the docs table has NO values_anchor — it's an intents col; a blind rename would've broken). Go idents + `/api/studies/*` routes + JSON labels KEPT (frontend contract); cosmetic "Studies"→"Docs" relabel = polish follow-up. New `extension/ui.Dockerfile` (node→go-embed→alpine, clean-room) + default `ui` compose service (local-only 127.0.0.1:**8081**; committed default) + committed `frontend/dist/index.html` stub for local go:embed + frontend `.gitignore`. **VERIFIED RUNNING:** image builds, container up (verified this run on **8082** — 8081 was transiently held by a sibling vite dev-server PID 30556, NOT killed per presiding covenant), healthz 200, SPA serves, **all 36 GET routes 0×HTTP-500**, renamed surfaces (studies/get, studies/search→book-self-reliance, intents/get, watchman/pass) exercised with REAL data → all 200. Route-verify done as a curl-sweep ORACLE (beats fan-out for a mechanical "is it 500?" check — told Michael).
  - **✅ M6 CODER-HARDENING DONE + PUSHED (`6078771`) — code-pr works out-of-the-box now.** Found 2 real public-repo defects: review model qwen3.7-max (401s oa-compat / 400s Alibaba-tools) → glm-5.1 (pipeline JSON + stage_models); verdict parser START-anchored (`^REVIEW: passes`) misread preambled passes as revise → LINE-anchored `(^|\n)\s*REVIEW:\s*passes` (cv6+cv11); pr stage no round-cap → 40. Live-applied (MSYS_NO_PATHCONV=1 docker cp) + **PROVEN by a fully AUTONOMOUS no-touch run → DRAFT PR #2** (clone→…→review→pr, zero manual nudge). Live↔source coder drift reconciled. Journal `2026-06-14-m2pr-m3-ui-m6-coder.md` (`26ca4f0`).
  - **✅ M4 SCHEDULER LEG CONFIRMED RUNNING** — book-digest fires hourly (22:00→01:00) + playlist-digest at 00:00 autonomously on OSS, producing real docs (incl `yt-RB8vjn1QPeM` self-improvement video off the playlist). Only M4-PERSONAS remains ⚠key-safety (the persona *pipeline* can be verified safely via SQL dispatch w/o the gateway = no live-room double-fire; connecting to a REAL room needs Michael).
  - **✅ M5 COUNCIL DONE + RATIFIED 2026-06-14 (Michael chose "convene the council now"). dominion_in_council SATISFIED. 4 decisions:** (1) timing = **mid-turn** (waiting_for_tools); (2) compactor = **fixed cheap model, fast + ~1M ctx, but TUNABLE via config** (Michael: "run experiments to find a good compactor counselor"); (3) sees = judge-brief condensed surface; (4) trigger = **agent-initiated + ≥threshold pressure-line nudge** (persuasion not compulsion). **Mechanics VERIFIED (all primitives present):** mute is by message-id (`context_mute(msgid)`) + `context_resolve_handle(parent_session, handle)→msgid` so the compactor acts on the PARENT's msgs; **`context_expand` = the reversible unmute (blind-spot resolved, safe-by-construction holds)**; mid-turn rails already exist (`tool_dispatch_complete_waiting` resumes a `waiting_for_tools` parent — spawn_subagent rides them, compact_context reuses); agents have a `context_tools_enabled` flag. **★ CORRECTION found in grounding (would've broken a rushed build):** `render_judge_brief_surface` is **per-MESSAGE**, not whole-session — session-level "what it sees" = the `context_pressure(session)` foldable list (handle+est_tokens) + per-msg reads. **★ M5 BUILT + PROVEN E2E SAME SESSION (OSS `a8d5cc5`)** — Michael corrected the timezone ("only 8:41pm CDT, I still got time!"; the DB UTC stamps misled me to "past midnight"), so we built it. `extension/21-compact-context.sql` (compact_context_surface = what the compactor sees; compact_context_apply = the substrate acts, mute/compress/pin by msgid only-if-belongs + honest curated-footprint metric + [COMPACTED] marker; pressure-line nudge; **compactor agent = TOOLS-OFF JUDGE** returning a {mute,compress,pin} JSON verdict [judges-not-executors — sidesteps the _session_id-is-the-compactor's-own-session trap]; compact-context 1-stage pipeline [deepseek-v4-flash, tunable]; deny-all-heavy grants; compact_context tool_def) + `cmd/stewards-mcp/compact_context.go` (mcp_proxy handler: reads injected `_session_id`, renders surface, **inherits caller's work_item intent** [spawn_subagent_create p_parent_work_item_id — core ships no default intent], spawns+polls compactor like spawn_subagent, applies verdict; extractJSONObject tolerates wrapping) + lib.rs chain (create_compact_context requires create_coder) + Dockerfile COPY + main.go register. **PROVEN: 14-msg migration session clogged w/ spent grep+schema dumps → compactor verdict `compress [1106,1107]` keep the plan → applied (compressed=2, [COMPACTED]) in 25s @ $0.** Bugs found+fixed in e2e (all MY plumbing, not design): (1) agent ON CONFLICT was (family) not (family,model_match); (2) spawn needs intent → inherit parent work_item; (3) live pipeline had stale model "openai" (re-apply); (4) poll waited for maturity=verified but a 1-stage tools-off pipeline ends at status=completed → broke into the bridge's 120s call-timeout → fixed to treat completed as terminal + cap wait at 110s. **Relief is governed by existing pressure-rendering** (muted/compressed fold to tombstones UNDER pressure; not an immediate delta — honest metric reports curated footprint). FOLLOW-UPS: tests/virgin-smoke.sql assertion; **pg/extension image rebuild validating pgrx compiles 21** (running; psql-applied clean + go build/vet green; CI also gates). compactor-model tuning = Michael's experiments.
  - **REMAINING:** M5 BUILD (spec'd, execution-ready) · M4-personas ⚠key-safety · M6-tail (stamp default + bgworker-surface + capability-substitution polish) · M7 soak · CUT ⚠his Hinge. **Net tonight: M2(PR#1)+M3(UI running)+M6-coder(PR#2)+M4-schedules DONE/verified, M5 councilled+spec'd.** OSS UI running on 8082; digesters hourly on Michael's lent keys.
- **★★ PARITY ROADMAP RATIFIED 2026-06-13 (tasks #161-168). Michael: "keep going until we hit parity… only braking for critical matters." Revised parity = #4 + coder enablement + UI/CLI up, before the cut. Ratified up front (3 Qs):**
  - **yt-mcp placement = OSS cmd/ + opt-in `docker-compose.yt.yaml`** (generic, public, but Python/yt-dlp behind an opt-in layer; default bridge stays lean; workspace overlay does NOT grow).
  - **Cockpit scope = ADD WRITE VERBS now** (dispatch/council/ratify/review) — Michael chose the fuller path over read-only-is-parity.
  - **compact_context = BUILD AS PARITY** (net-new in OSS core; Michael chose it in-scope over defer-to-post-cutover). Sketch defaults: between-turn, judge-pattern, cheap compactor; parked Qs in `projects/pg-ai-stewards/.spec/proposals/substrate-compact-context-sidequest.md`.
  - **VERIFIED THIS TURN (real-path, not memory):** OSS `cmd/stewards` cockpit ALREADY extracted (read-only P1: project/board/watch/cost; pgxpool); NO web frontend ever existed in either repo (UI = the terminal cockpit). `compact_context` = ZERO occurrences in OSS **and** private (never built; "pulled in at council" never happened — authoring leg only consolidated existing). Reactive context engine IS in OSS (compose_messages/compose_system_prompt/intercept_*/extract_engrams/render_judge_brief_surface). yt-mcp = trivial fold (package main, no cross-pkg imports, hand-rolled MCP, empty go.mod require; only runtime dep = yt-dlp/python, NO ffmpeg).
  - **MILESTONES (run straight through; brake only at ⚠):** ✅M1 yt-mcp+playlist · M2 coder e2e ⚠repo · M3 UI-extract(as-is) · M4 personas+schedules ⚠key-safety · M5 compact_context ⚠parked-Qs · M6 cleanups(promote_trigger sabbath-wrap, cv4→overlay, BYO-MCP docs) · M7 soak→parity-proven · CUT ⚠Michael's session (Hinge ①+③).
  - **★ M3 CORRECTED + RE-RATIFIED:** the "UI" is NOT the small cockpit — it's `scripts/stewards-ui/` = a ~11K-LOC **Vue+Go web app, 23 routes** (the live `pg-ai-stewards-ui` container, `stewards-ui --addr :8080`, served from `extension/ui.Dockerfile`), NOT yet in OSS. Michael re-ratified (correcting his earlier "add write verbs now" which was against my wrong description): **EXTRACT AS-IS** (existing write islands NewWork/ratify/Trust/Councils) + doc_* rename across it + go.work→single-module + node/go Dockerfile + compose + route-verify; evolution-proposal additions (substrate chat/intent authoring, `.spec/proposals/stewards-ui-evolution.md`) = POST-cutover. The read-only `cmd/stewards` CLI cockpit IS already in OSS (separate, smaller).
  - **✅ M1 DONE + PUSHED (OSS `3e5ef66`):** folded scripts/yt-mcp→cmd/yt-mcp (+ NEW `yt_playlist` flat-playlist enumeration tool — the discovery step the download-first tools lacked) behind opt-in `docker-compose.yt.yaml` (WITH_YT=1 → yt-mcp + python3/yt-dlp; default bridge stays lean). `examples/playlist-digester.sql` = playlist-digest pipeline (read→digest→critique[qwen3.7-plus]→recommend, kimi-k2.6 doer) + playlist_watch/playlist_seen + playlist_next/publish/add + video-study intent + 6-hourly schedule + ai-research playlist seed. Genericized 3 personal strings. **PROVEN E2E on the OSS stack** (bridge rebuilt WITH_YT=1; refresh-tools **7/7**, yt 5 tools; ran a real digest of WGwRCw9TRyo "This 1 Book…Geniuses" [the Euclid vid] off the playlist → 7804-char digest → study/yt/WGwRCw9TRyo.md doc + pending_file_write + brain entry + playlist_seen+1). Gotchas: yt-mcp needs args=['serve']; work_item_create does NOT auto-dispatch (call work_item_dispatch_stage; the scheduler does both). yt_playlist NOT backported to scripts/yt-mcp (deliberate fork; live tool untouched).
  - **M2 IN FLIGHT — coder PROVEN to write+test on OSS; PR hop being nailed:** created throwaway repo `cpuchip/pg-ai-stewards-coder-proof` (host gh — the bridge's scoped PAT can't create repos, 403; correct hardening). CODER_REPO_ALLOWLIST=`github.com/cpuchip/` already covers it. **Run 1 (53a8e452, CANCELED):** clone stage hit a **GitHub propagation RACE** (repo created seconds earlier → git backend not ready → "repository does not exist") so the coder initialized a fresh LOCAL module, wrote greet pkg + table tests + main.go, **`go test ./...` PASSED in the hardened sandbox** — coder capability proven on OSS (sandbox spawn via host socket, token bridge-side, code-write, test-green). Stalled at **plan_review/awaiting_review**. **ROOT CAUSE FOUND (not a bgworker gap — my incomplete input):** the code-pr templates reference `input.acceptance_criteria` (+ binding_question/repo/base_branch; sandbox/plan_feedback/review_feedback auto-stamped by `stamp_code_write_sandbox` trigger). I omitted acceptance_criteria → plan_review auto-dispatch's `render_stage_input` RAISED `resolve_template_path … NULL` → bgworker SWALLOWED the error → parked at awaiting_review with NO plan_review in stage_results + NO work_queue error. The bgworker auto-advance is FINE (digesters prove it). **Run 3 (e9364f01) dispatched WITH acceptance_criteria** — should clear plan_review → implement → verify → review → pr → DRAFT PR. Separate concern to watch: runs 1+2 clone stage narrated "repo did not exist" + fell back to local-init even though raw bridge `git clone` works (exit 0) — propagation race on run 1, run 2 ambiguous; watch run 3. **M6 polish candidates:** stamp trigger should default acceptance_criteria='' (forgiving input) like plan_feedback; bgworker should surface (not swallow) template-render errors. Gotchas: work_item_create needs an intent (no 'default' seeded → reuse video-study); work_queue has NO work_item_id col (links via payload); stage outputs live in work_items.stage_results jsonb; CODER_REPO_ALLOWLIST=`github.com/cpuchip/`; bridge PAT can't create repos (403, correct).
  - **M2 VERDICT (4 runs): CODER PROVEN ON OSS; full code-pr→DRAFT-PR blocked by provider/model-format issues (M6/debug, NOT coder defects).** Run 4 (5633b0e3) ran ALL of clone(REAL repo — race gone)→plan→plan_review→implement→verify→review; implement reports "Build, tests, run pass exit 0" on go.mod module cpuchip/pg-ai-stewards-coder-proof. **review stage FAILED: HTTP 400 "Error from provider (Alibaba): When using tool_choice, tools must be [non-empty]"** (qwen3.7-plus via Alibaba — a dispatch sends tool_choice with empty tools on a tools_disabled review stage → provider rejects; substrate provider-adapter bug, Go/Rust). My poll auto-nudge (dispatch on awaiting_review) made review OSCILLATE+fail — MY artifact (review's awaiting_review is normal verdict-pending, not stuck; do NOT nudge it). **MODEL CHECKER (Michael's recall — CONFIRMED WORKS):** probe flips qwen3.7-max→usable=f (401 oa-compat) / qwen3.7-plus→usable=t; substitution (model_substitutions + dispatch M.2) exists. GAP: auto-probe NEVER RAN on OSS (was 0/13; NOW 13/13 after manual enqueue_due_model_probes — 1 unusable/12 usable). **auto-probe IS wired** (CORRECTED): trigger `watchman_passes_schedule_model_probes AFTER INSERT ON watchman_passes` → enqueue_due_model_probes (19-models.sql §5/m5). Watchman cron = **weekly@sun-03:00** → probes refresh weekly; fresh stack hadn't probed yet (0/13). Manually probed all 13 now → current. NOT a bug; shorten watchman cron for faster refresh if wanted. **REAL review-stage root cause:** review (tools_disabled=FALSE — legit, template calls coder_shell/read to inspect the diff; all code-pr stages = `dev` agent w/ coder tools) used **qwen3.7-plus via Alibaba**, which rejects tools+tool_choice (`HTTP 400: tools must be non-empty`) — a qwen/Alibaba tool-format quirk; kimi-k2.6 (implementer) handles tools fine. tool_choice set in 15b chat_post_internal/dry_run_chat. **FIX (run 5, ad44ed67):** review gates plan_review+review → **glm-5.1** (opencode_go, tool-capable, non-qwen, probed usable, ≠ kimi implementer) in BOTH pipeline def + stage_models; re-running observe-only (NO nudge). **M6 ITEMS:** (1) review-model provider-portability — OSS code-pr ships qwen3.7-max review which fails on non-qwen-tool providers; make it capability/provider-aware OR per-stack stage_models override (proper fix vs my live pipeline-def edit); (2) probe TOOL-use not just chat (qwen3.7-plus passes chat-probe but fails tool-use) — capability-aware substitution; (3) stamp trigger default acceptance_criteria=''; (4) bgworker should SURFACE not swallow render errors; (5) reconcile my live code-pr pipeline-def edits (qwen→glm) back to canonical. Throwaway repo cpuchip/pg-ai-stewards-coder-proof exists (host gh; bridge PAT can't create). **★ M2 FINAL VERDICT (5 runs): CODER FULLY PROVEN ON OSS; DRAFT PR blocked ONLY by GitHub token repo-scope (Michael's grant, NOT a code defect).** Run 5 (ad44ed67) with glm-5.1 reviewers cleared plan→implement→verify→review cleanly; pr stage (after raising max_tool_rounds_hard→40 on the live sandbox: write 4 files, `go test` GREEN, local commit) **blocked on push: `remote: Permission to cpuchip/pg-ai-stewards-coder-proof.git denied to cpuchip. 403`** — confirmed by a direct bridge push test. The fine-grained PAT is scoped to specific repos (clone=public-read OK, push=write DENIED on the new repo); same root as the create-403. **TO LAND THE PR:** Michael adds the repo to the PAT's repo access (or retarget the coder at a repo the token already writes, e.g. ai-chattermax, or use a broader token) → then re-dispatch the pr stage (code is proven; ~30s + 1 stage). **GENUINE FIXES FOUND (promote to OSS-canonical in M6, NOT band-aids):** review gates glm-5.1 (portable vs qwen3.7-max which 401s + qwen3.7-plus which 400s-on-Alibaba-tools); pr stage max_tool_rounds_hard=40 (default too low for clone+diff+commit+push+gh-pr). Live pipeline-def edits on Michael's dev stack to reconcile. Proof sandboxes reaped. Runs 1-4 canceled/failed.
  - **Context-engine answer (Michael's Q):** OSS has the REACTIVE engine (compose_messages/compose_system_prompt/intercept_*/extract_engrams/render_judge_brief_surface); `compact_context` (proactive between-turn) = ZERO in OSS AND private — never built (M5 builds it).
  - **Checkpoint discipline:** commit+journal at each milestone close so a compaction can't lose progress (bom-walk resume mechanism).
- **★ ROADMAP 2026-06-13 (Michael's 7-item plan). #1 + #2 SHIPPED this turn, then journaled:**
  - **#1 DONE — exa web search = the OSS default** (`fd08fea`): seeded exa-search keyless
    free-tier (works OOTB), reversed M2's BYO call; smoke + prompts updated; refresh-tools
    6/6 incl exa-search [OK]. (Live instance already had it + grants — verified real-path.)
  - **#2 DONE — model-wiring examples** (`fd08fea`): `docs/wiring-up-models.md` (opencode
    zen-free/go/gemini/lm-studio env pattern) + `examples/models.sql` (real-price snapshot
    catalog, free models flagged) + `.env.example` provider block + `examples/README.md`.
  - **JOURNALED** (`096c7df`): `.spec/journal/2026-06-13-mcp-packaging-coder-and-usable.md`
    covers the whole session (plan→M0→resolver→harness→study-spec→M1 coder→M2→#1+#2).
  - **#3 DONE + RAN — book-digester** (OSS `eacb6c7`, `examples/book-digester.sql`):
    book_shelf + book_next/book_publish/book_add tools + 4-stage `book-digest` pipeline
    (read→digest→critique→recommend) + `book-study` intent + hourly schedule + shelf
    (Self-Reliance→Meditations→Tao Te Ching→Art of War). **PROVEN END-TO-END on the OSS
    stack with Michael's keys** (he copied his live .env in — NOT the cutover, just lending
    the dev stack his model keys side-by-side): first run digested Self-Reliance (8KB digest;
    **the qwen3.7-plus critique stage caught + corrected a real placement error** — null-case
    working), published study/books/self-reliance.md + a brain entry, advanced the shelf to
    Meditations. Single-pass v1 (long-book map-reduce = v2). Uses stewards-explore (kimi-k2.6
    doer / qwen3.7-plus critic). Verify-loop catches: grant source enum=manual; core ships
    NO intents (seed own) + NO 'research' agent (use stewards-explore). chat-post deferred
    (double-fire: don't run OSS --profile personas with live persona keys). **OSS stack left
    RUNNING — hourly tick will digest Meditations next.**
    - **★ AUTONOMOUS TICK CONFIRMED (`7ceb658`):** the hourly schedule fired at 22:00 ON ITS
    OWN + digested Meditations (done 22:08, 11KB) — the #6 self-improvement-loop heartbeat,
    live, on a human shelf. Self-limits: when the shelf empties, book_next→null→read stage
    outputs "SHELF EMPTY" + no-ops (no runaway). tao-te-ching queued for 23:00.
    - **MATERIALIZE WIRED:** book_publish now also enqueues a pending_file_writes row →
    digests are materialize-capable (DB always; disk when /workspace RW or at cutover →
    real study/books/). Proven via `stewards-cli materialize-writes` (wrote self-reliance.md
    8KB + meditations.md 11KB to disk from the queue). OSS /workspace stays RO (safe default).
  - **NEXT per Michael: #4 playlist digester** → #5 cutover → #6 self-improvement loop → #7 fun.
    Details below.
- **★ ROADMAP DETAIL (captured, #3+ NOT yet built):**
  - **Search VERIFIED WORKING in LIVE** (`pg-ai-stewards-dev`): `exa-search` enabled +
    granted to research/study/research-gospel + **real-path test PASSED** (web_search_exa
    "Euclidean algorithm" → Wikipedia article, 4779 chars, no error). Keyless = **Exa
    free/anonymous tier** (no EXA_API_KEY anywhere; that's fine, it works). Old DuckDuckGo
    `search` also still enabled in live (harmless; archived in OSS). His project is unblocked.
  - **Both schedules INTACT:** `ai-news-7am` (research-summary, weekdays 13:00 UTC) +
    `science-news-weekly` (research-write, Mondays 13:00 UTC), enabled, next 2026-06-15 —
    live AND ported to overlay (pe7/pe8). "They are fun, keep them." → they stay.
  - **WANT #1 — YT-playlist digester (most actionable):** poll his "AI research" playlist
    (youtube.com/playlist?list=PLcHf1NPbY2qXi5MkL-BzJb7t4r-m8SIEq) a few times/day → new
    video → transcript → digest → **actionable "what to learn/do" recs**. = the study-pipeline
    spec + scheduler(18) + yt_transcripts(live). **Models: kimi-k2.6 doer, qwen3.7-PLUS
    critic (NOT qwen3.7-max — ~2x cost of k2.6/3.7-plus).** New: kimi-k2.7-coder just dropped.
  - **WANT #2 — model-API examples (→ M3):** opencode zen (FREE models OOTB) + opencode go +
    google gemini key + LM Studio local, as easy import/example pipelines; price-tier import
    for zen/go "as we had them" OR cite source so an AI keeps them current.
  - **WANT #3 (VISION) — hourly self-improvement loop:** agent picks a subject within a sphere
    (AI/physics/startups/book-writing), every hour pulls something of its OWN choice, learns,
    does something interesting, idles. Ref video https://youtu.be/RB8vjn1QPeM . "automate
    something for itself, may or may not be useful for us." = scheduler + study-pipeline +
    agent-chosen subject. NORTH-STAR — spec carefully (watch the video first), don't rush.
- **★ M2 — fetch-md + git UTILITIES SHIPPED 2026-06-13 (OSS `4a31b03`, pushed;
  Michael "Lets do M2").** Ported `cmd/fetch-md-mcp` (fetch_url/fetch_urls/
  extract_links/fetch_url_raw; chromedp js path kept but NO chromium in bridge —
  static fetch works, js:true degrades, documented) + `cmd/git-mcp` (git_* +
  gh_pr/issue; agent/* branch namespace, main/master/release/* refused) into the
  OSS root module (folded, flat package main, `go mod tidy` pulled fetch-md's
  chromedp/html-to-markdown/readability/tabula tree). Seeded `fetch-md` + `git`
  mcp_servers in 05-mcp-bridge (deny-by-default grants stay operator; git reads
  GITHUB_TOKEN at exec). bridge.Dockerfile builds both. Genericized fetch-md UA +
  git co-author default. Dropped archived `web_search (DuckDuckGo)` from the 13
  research prompts. **★ KEY DECISION: web search is NOT core** — the virgin-smoke
  denylists `search` AND `exa-search` as personal (needs an operator API key), so
  web search is BYO → M3 docs. virgin-smoke now asserts the **5-server generic
  core** (fs-read/pg-ai-stewards/fetch-md/git + coder). **VERIFIED:** virgin-smoke
  PASS; `refresh-tools` 5/5 OK (fetch-md 4 tools, git 8 tools live). Gotcha: the
  scratch `pgdata` volume persisted old seeds → needed `down -v` (scratch oss
  project only) for a fresh install to see the new servers. **Task #160 DONE.
  Only M3 (BYO-MCP docs + example web_search_exa overlay) remains in the plan.**
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
