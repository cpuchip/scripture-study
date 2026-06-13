---
lane: pg-ai-stewards
session_id: 7ea7faa4-688a-451a-ac68-b7ea662d4b81
status: active
started: 2026-06-11T22:00:16
last_active: 2026-06-12T21:33:55
---

## Working on
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
  **NEXT = B4** (13-research-pipelines..16-subagents, minus h1-0 now done):
  13 = h1-0(DONE,drop)/h1-2/h1-7b/h2/h3-4/h3-5/h3-followup-3/i4/i6/i7/pe2
  (enqueue_proposed_work_items — on_maturity_verified's planning branch
  already calls it, wrapped); 14 = j1-j9c incl j8a-dispatch + j11-dispatch-
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
