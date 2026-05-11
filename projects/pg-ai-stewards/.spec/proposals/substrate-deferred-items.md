---
title: Substrate deferred items — catalog of "wait for signal"
date: 2026-05-11
status: catalog — not a build proposal
parent: open-items.md (Sections V, X.1, X.4, X.5, X.6, X.8, X.9)
purpose: >
  Single catalog of substrate work explicitly deferred — items that have
  a clear shape but lack a triggering need. Kept here so they're visible
  if conditions change, but separated from the active build queue
  (substrate-completion-batch-g.md / substrate-pipelines-expansion.md /
  stewards-ui-evolution.md).
---

# Substrate deferred items

Items in this catalog share a property: **the shape is known but the need isn't pressing.** Each has a trigger — a condition under which it stops being deferred. Until the trigger fires, no build session should start with these.

Grouped by:
- **A. Deferred-by-design** — explicit "only if needed" decisions
- **B. Wait for usage data** — sub-spec ratifications that should only revise after the substrate has been used at scale
- **C. Wait for steady state** — community contributions etc.

---

## A. Deferred-by-design

### A.1 Phase 3g — Multi-provider expansion (Anthropic, Gemini, Veo, TTS)
- **Source:** `phases.md` line 1201
- **Today:** substrate routes everything through `opencode_go` (Chinese-model gateway: kimi/glm/minimax/qwen). `lm_studio` for local. Anthropic via OpenCode Zen handles the "boost to opus" case via human-mediated escalation queue.
- **Trigger:** a specific provider provides uniquely useful output (e.g. Gemini for a multimodal pipeline; Veo for video; TTS for audio outputs).
- **Effort if pursued:** medium per provider (1 session each — provider registry row + response-shape normalization + agent variant rules + tool-shape adapters).
- **Anti-pattern:** building all four speculatively. Wait for use to ask.

### A.2 Phase 4 — GraphRAG over the canon
- **Source:** `phases.md` lines 1393-1416
- **Today:** vector search + AGE citation graph + study_search_text FTS cover the substrate's existing query needs.
- **Trigger:** a real "themes across the whole corpus" question fails with current tools. (Example: "what does the conference talk tradition say about the relationship between intelligence and embodiment across all 12 apostles' Easter talks since 1971?" — if this can't be answered with current tools.)
- **Effort if pursued:** large (Microsoft GraphRAG indexing pass + AGE schema collaboration with gospel-engine-v2 + new MCP tool).
- **Anti-pattern:** building it speculatively. Cost is real (indexing runs are not cheap).

### A.3 Phase 7+ stretch items
- **Source:** `phases.md` lines 1475-1518
- **`postgres_fdw` from stewards into gospel** — SQL-level joins. Trigger: if substrate frequently joins gospel-library data with stewards data via HTTP and HTTP latency hurts. Today it doesn't.
- **Multi-tenant RLS** — if ibeco.me ever hosts other people. Has detailed sub-questions catalogued in `phases.md` lines 1481-1518 (workstream-as-sharing-unit, owner_user_id + visibility columns, AGE Cypher RLS caveat). Trigger: someone asks to share.

### A.4 Phase 1.7 — Brain CLI driver + hybrid FTS+vector
- **Source:** `phases.md` lines 119-131
- **Today:** brain CLI uses SQLite as read-only fallback; Postgres backend exists but unwired. Hybrid FTS+vector deferred until embedding column had real traffic (it now does).
- **Trigger:** SQLite-as-brain-backend starts hurting (slow queries, lock contention, schema drift).
- **Anti-pattern:** unifying the brain backends without a forcing function.

### A.5 3d.2 + 3d.3 — git-mcp wrapper layers
- **Source:** `proposals/pg-ai-stewards-3d-sandboxed-git.md` lines 327-338
- **Today:** 3d v1 (native Go MCP wrapper, Option A) shipped 2026-05-09.
- **3d.2 Docker sidecar:** trigger if v1 surfaces an escape risk worth more isolation, OR if multi-pipeline parallel git work needs containerized workdirs.
- **3d.3 safe_outputs proxy (GitHub Agentic Workflows pattern):** trigger if 3d.1 + 3d.2 stabilize AND the buffer-and-vet flow is needed.
- **Effort:** medium each.

### A.6 GITHUB_TOKEN setup (PAT) for 3d v1
- **Source:** `phases.md` line 1187
- **Today:** 3d v1 functional but never end-to-end-tested because no PAT in .env.
- **Trigger:** Michael wants substrate to start opening PRs.
- **Effort:** ~5 min (generate PAT, .env, restart bridge).

### A.7 SSE live tail for work_queue
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Today:** 5s dashboard polling covers the use case.
- **Trigger:** real volume warrants — e.g. councils running with sub-second member responses (won't happen with LLM latencies), or several scheduled pipelines firing simultaneously.
- **Effort:** large (~2 hours) — JSON streaming infra.

### A.8 Citation extraction for substrate-promoted studies
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Today:** `/api/graph` returns 0 edges for `substrate--*.md` studies because the promotion pipeline doesn't extract citations. Workspace-imported studies do show edges.
- **Trigger:** substrate studies start citing each other meaningfully — graph view becomes worth populating.
- **Effort:** medium (extend promotion pipeline to walk the markdown for `[slug](#)` links + create AGE edges).

---

## B. Wait for usage data

These are ratifications from the C–F sub-specs that should only be revised AFTER the substrate has been used at scale. Each has a recommended initial setting; only revisit if the data says otherwise.

### B.1 F2 evolution — `council_authority` as separate trust dimension
- **Source:** `phase-f-design.md` § V.5, journal F
- **Today:** `bishop_eligible` requires master-on-pipeline-of-intent for agent bishops. Michael's 2026-05-11 nuance flagged that bishop authority might be a cultivated skill (different from execution authority). Debug agent is candidate first cultivator.
- **Revisit trigger:** after first real council with attempted agent bishop. If master-on-pipeline doesn't map to good bishop behavior, introduce `council_authority` as separate trust dimension.

### B.2 D-F1 concurrency lift criterion
- **Source:** `phase-f-design.md` § VI
- **Today:** `one_active_council` partial unique index allows exactly 1 concurrent.
- **Revisit trigger:** if `>5 refusals/week` of attempted concurrent convene. Lift to 3.

### B.3 Atonement cost guardrail
- **Source:** `phase-d-design.md` § VI
- **Today:** Atonement shares the per-work_item cost cap.
- **Revisit trigger:** observed atonement cost on real quarantines blows past expectations. Add separate cap if so.

### B.4 Lesson de-duplication signal
- **Source:** `phase-d-design.md` § VI
- **Today:** if the same insight surfaces in 5+ atonement events, nothing detects it.
- **Revisit trigger:** ≥10 ratified lessons in the corpus + visible duplication in human review.

### B.5 Override weighting
- **Source:** `phase-e-design.md` § VI
- **Today:** every override counts as full-weight failure for trust.
- **Revisit trigger:** trust signal feels noisy in practice — maybe "I should have surfaced" is lighter weight than "I should have advanced."

### B.6 First-completion bootstrap friction
- **Source:** `phase-e-design.md` § VI
- **Today:** every new (agent, pipeline, model) cell starts trainee; 5 surfaces required to escape.
- **Revisit trigger:** if month-1 lived experience feels heavy on attention. Lower the threshold to 3.

### B.7 System-suggested council binding-question specificity
- **Source:** `phase-f-design.md` § VI
- **Today:** `suggest_councils` returns the cluster (pipeline, stage, sample lessons); UI synthesizes a binding question via simple template.
- **Revisit trigger:** real clusters appear + the template-synthesized question feels wrong. Have an LLM compose sharper.

### B.8 Council error handling — what if all members error
- **Source:** `phase-f-design.md` § VI
- **Today:** bgworker auto-fires synthesize when 0 proposer/critic members have `completed_at IS NULL`. If a member errored, `completed_at` stays NULL and the council hangs.
- **Revisit trigger:** any real council with ≥1 errored member. Probably: treat error as completion-with-empty-response so synthesize fires with whatever succeeded.

### B.9 Bishop dispatch path for agent bishops
- **Source:** `phase-f-design.md` § VI
- **Today:** when bishop is an agent identifier, resolution still requires human click in UI.
- **Revisit trigger:** F2 (B.1 above) lands. Need a `_council_bishop_dispatch=true` chat path that runs the agent through the resolution decision.

### B.10 Token-cost audit of compose_system_prompt injection
- **Source:** `journal/2026-05-11-substrate-phase-c-shipped.md`
- **Today:** ~600 tokens/dispatch added by covenant + intent blocks. Predicted acceptable; not measured.
- **Revisit trigger:** first real study run on the new substrate — check cost panel. If cost spikes, add `compose_system_prompt(skip_covenant=true)` flag.

### B.11 bgworker `payload._kind` enum refactor
- **Source:** journals D, E, F
- **Today:** bgworker switches on 7 marker booleans (`_gate_eval` / `_scenarios_gen` / `_verify` / `_sabbath` / `_atonement` / `_council_member` / `_council_synthesize`). Eighth marker (`_steward_chat` from stewards-ui-evolution.md V.B) is incoming.
- **Revisit trigger:** when the 8th marker lands. Refactor to `payload._kind` enum + match.

---

## C. Wait for steady state

### C.1 Phase 6 — AGE upstream contributions
- **Source:** `phases.md` lines 1418-1474, full catalog in `projects/pg-ai-stewards/docs/AGE-QUIRKS.md`
- **Today:** 8 AGE quirks catalogued during Phase 2.6 work. No PRs filed.
- **Trigger:** substrate is in steady state — delivery isn't pulling cycles from AGE work.
- **First PR target:** quirk #2 (apostrophe-in-interpolated-Cypher) — cheapest, most isolated. Build #6/#7 later (deeper agtype type system changes).
- **Anti-pattern:** filing PRs while substrate work has open carry-forwards. Open-source contribution is generous but requires bandwidth.

### C.2 Phase 3h — Per-model prompt tuning generalization
- **Source:** `proposals/pg-ai-stewards-per-model-prompt-tuning.md` (status: deferred)
- **Today:** prototype shipped 2026-05-08 (kimi + qwen study variants); cross-topic validation hasn't run.
- **Trigger:** Batch H ("Use validation") has produced 2–3 real study runs through the new substrate. THEN cross-topic validation has real data instead of being synthetic.
- **Sub-phases (4):** 3h.1 cross-topic validation, 3h.2 extend to lesson/talk/journal/research-gospel agents, 3h.3 onboard additional models (Sonnet, Gemini, GLM), 3h.4 `study-bench` CLI mirroring `classify-bench`.
- **Effort if pursued:** large (~6-8 sessions across the 4 sub-phases).

---

## D. Verification items (NOT deferred; quick checks)

These are spot-checks rather than build work. Probably ≤30 min each.

### D.1 `stewards.todo_rollup_audit()` — did it ship?
- **Source:** `proposals/pg-ai-stewards-phase-2-5-generic-substrate.md` § 2.6b
- **Action:** `\df stewards.todo_rollup_audit` against pg-ai-stewards-dev. If exists, run it. If missing, decide whether to build it (small surgery — parent-done-with-open-children correctness check).

### D.2 3c.4 — gospel-engine HTTP path
- **Source:** `proposals/pg-ai-stewards-3c-2-5-study-tools.md` § carry-forward line 345
- **Action:** verify substrate agents reach gospel-engine via bridge → mcp_proxy (the Phase 3e path), NOT some hypothetical 3c.4 separate HTTP path. Probably resolved by 3e; this is just confirmation.

---

## E. Items pulled into active proposals

These were on the open-items list but have been incorporated into active build proposals. Listed here for traceability:

| Item | Pulled into |
|---|---|
| 1.1 studies.file_path NOT NULL | `substrate-completion-batch-g.md` § V.1 |
| 2.1 Steward retry pulls lessons | `substrate-completion-batch-g.md` § V.2 |
| 2.2 Steward quarantine fires atonement | `substrate-completion-batch-g.md` § V.3 |
| 2.3 Hybrid revise | Out of Batch G — needs its own session (separate doc to follow if Michael wants) |
| 2.4 covenant_check dispatch path | Out of Batch G — needs its own ratification |
| 3.1 File-write mechanism | `substrate-completion-batch-g.md` § V.4 |
| 4.2 Stewards-UI sidebar grouping | `stewards-ui-evolution.md` § V.C |
| 4.5 gh_issue_create grant | A.6 above (wait for use) |
| 4.7 Watchman finding-ack | `stewards-ui-evolution.md` § V.D |
| 4.8 Bridge refresh-tools button | `stewards-ui-evolution.md` § V.D |
| 4.9 Substrate citation extraction | A.8 above (wait for use) |
| 4.10 work_item write actions | `stewards-ui-evolution.md` § V.D |
| X.3 7-day soak | Batch K (its own observation week — see open-items.md) |

---

## F. How to use this catalog

- **Don't start a session from here.** Active build queue is in `substrate-completion-batch-g.md`, `substrate-pipelines-expansion.md`, `stewards-ui-evolution.md`.
- **Check this catalog when:** a triggering condition fires (e.g. SQLite-as-brain-backend starts hurting → revisit A.4); or after a major-phase milestone (sweep B items for revision based on what was learned).
- **Update this catalog when:** an item triggers and moves to an active proposal; or new deferred items surface from a build session.
