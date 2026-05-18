---
title: pg-ai-stewards open items — navigation hub
date: 2026-05-17 (Section 0 refreshed post-ES arc; consolidated into proposals 2026-05-11)
status: living document — update after each work session
purpose: >
  Navigation hub for substrate work-in-progress. As of 2026-05-11
  consolidation, the items listed here are also captured in topic-grouped
  proposals under `.spec/proposals/`. This document remains as the
  cross-reference table + the "what's the queue look like?" view.
---

# pg-ai-stewards open items

## 0. Active proposal queue (refreshed 2026-05-17 — post-ES arc)

The build queue is drained: every ratified batch through the ES emergency-stop
arc has shipped. What remains is two un-ratified proposals, one new idea, and
one stale proposal that needs a freshness check.

| Proposal | Scope | Status |
|---|---|---|
| [`substrate-pipelines-expansion.md`](proposals/substrate-pipelines-expansion.md) | research + YouTube (gospel + secular) pipelines + scheduled-pipeline machinery | **needs ratification** — D-PE1 through D-PE7. Live council queue. |
| [`stewards-ui-evolution.md`](proposals/stewards-ui-evolution.md) | intent/covenant authoring + substrate-aware chat + sidebar grouping + write actions | **needs ratification** — D-UI1 through D-UI12. Live council queue. |
| [`substrate-scheduled-workflows.md`](proposals/substrate-scheduled-workflows.md) | cron-style scheduled jobs — periodic research (physics news → exhibits), autonomous YouTube AI-video review, public-playlist ingestion | **new idea 2026-05-17 — needs council.** Stub drafted; decisions (D-SW1–D-SW7) not yet walked. |
| [`substrate-completion-batch-g.md`](proposals/substrate-completion-batch-g.md) | file_path fix + retry-pulls-lessons + quarantine-fires-atonement + file-write mechanism | **stale — freshness check needed.** Predates J/K/L/ES; items may be absorbed or overtaken. Re-verify against current schema before starting from it. |
| [`substrate-deferred-items.md`](proposals/substrate-deferred-items.md) | catalog of "wait for signal" items (NOT a build proposal) | reference only |

Master rule: **start a build session from one of those proposals, not from this document.** This document is the index.

### Shipped / closed (no longer in the queue)

| Proposal | Status |
|---|---|
| [`substrate-ES-emergency-stop.md`](proposals/substrate-ES-emergency-stop.md) | **CLOSED 2026-05-17.** ES.1/3/4/5/6/3.s5 — the bacteriopolis runaway worked through; the substrate now runs it clean to a verified artifact (~$0.33). ~95 commits, zero rollbacks. ES.6.A/B archived (revive only if those areas show future trouble). |
| [`substrate-batch-l-1-1-context-engine-v2-1.md`](proposals/substrate-batch-l-1-1-context-engine-v2-1.md) | **closed 2026-05-14** — Context Engine v2.1; the Judges pattern. |
| [`substrate-batch-l-context-engine-v2.md`](proposals/substrate-batch-l-context-engine-v2.md) | **shipped 2026-05-14** — Context Engine v2 (graduated rendering + provider-aware + engram search + 6 wrappers + depth cap). |
| [`substrate-batch-k-engram-context.md`](proposals/substrate-batch-k-engram-context.md) | **shipped 2026-05-14** — engram-based context compaction (K.1–K.9). |
| [`substrate-batch-j-fanout-brainstorm.md`](proposals/substrate-batch-j-fanout-brainstorm.md) | **shipped 2026-05-13** — fan-out + brainstorm + work-item hierarchy UI (J.1–J.5). |
| [`substrate-batch-i-agent-write-back.md`](proposals/substrate-batch-i-agent-write-back.md) | **shipped 2026-05-12** — agent-proposal pipeline + HTTP endpoint + UI filter. |

Substrate is feature-complete through Phase F as of 2026-05-11. This document collects every unfinished item the substrate work has surfaced — cleanups, bugs, validation gaps, and future evolution paths — so the next session has a single inventory to pick from rather than re-deriving the queue from journal archaeology.

Each item carries:
- **Source** — the journal entry or sub-spec it came from
- **Effort** — small (≤30 min) / medium (1 session) / large (2+ sessions)
- **Risk** — what breaks if we don't do this

Items are grouped by **theme**, not phase. The phase that surfaced the item is in parentheses for traceability.

---

## I. Pre-existing bugs (do these first)

These predate the substrate work but were caught during it. They block functionality the substrate now depends on.

### 1.1 `studies.file_path NOT NULL` blocks `promote_to_study`
- **Source:** `journal/2026-05-11-substrate-phase-d-shipped.md`, surfaced again in E + F journals
- **Effort:** small (≤30 min)
- **Risk:** every successful study-write run that reaches verified will fail at the consecration step. The Sabbath gate (D.5) now correctly waves them past the sabbath check, only to hit this NOT NULL. Effectively: `work_item_promote_to_study` has been silently failing for any study slug long enough that no one noticed until D.5 lit it up.
- **Fix:** either make `file_path` nullable, or compute it from the slug at insert time (`study/<slug>.md`).
- **Recommendation:** **do this before the first real Phase D end-to-end run.** Otherwise the first Sabbath-gated promotion will fail and we'll waste an LLM round.

---

## II. Wiring gaps (substrate built, not yet plumbed end-to-end)

The Build sessions left a few SQL helpers + Rust hooks that aren't called from where they should be. None block; all are 1-2 line surgeries.

### 2.1 Steward retry doesn't pull lessons yet
- **Source:** `journal/2026-05-11-substrate-phase-e-shipped.md`, sub-spec `phase-e-design.md` § V.6
- **Effort:** small (≤30 min)
- **Risk:** Phase E's `retry_guidance_with_lessons` exists but `4c-steward-dispatch.sql` still calls plain `retry_guidance`. Line-upon-line discipline is never exercised on real retries.
- **Fix:** swap the function call in `steward_dispatch.sql` to `retry_guidance_with_lessons(diagnosis, attempt, pipeline_family, current_stage)`. Live-apply via `docker cp + psql -f`.

### 2.2 Steward quarantine doesn't fire atonement yet
- **Source:** `journal/2026-05-11-substrate-phase-d-shipped.md`
- **Effort:** small (≤30 min)
- **Risk:** Phase D's `maybe_enqueue_atonement(work_item_id)` helper exists but the steward's quarantine path doesn't call it. So even with `pipeline.atonement_enabled=true`, no atonement fires when a work_item gets quarantined.
- **Fix:** add `PERFORM stewards.maybe_enqueue_atonement(v_work_item_id);` at the quarantine point in `steward_dispatch.sql`.

### 2.3 Hybrid revise (revise #1 same model + feedback prepended)
- **Source:** project memory `project_pg_ai_stewards_revise_hybrid.md` (ratified 2026-05-11)
- **Effort:** medium (1 session)
- **Risk:** today's revise immediately escalates model on retry #1. Wastes the focused-critique opportunity that the gate's feedback represents.
- **Fix:** add `feedback` column to dispatch payload; `work_item_dispatch_stage` prepends as "Previous attempt critique:" when set; steward retry path stashes feedback + skips model_override on revision_count=0; only escalates on revision_count=1 (the second revise). Cap stays at 2 → surface (D-B2 unchanged).

### 2.4 covenant_check template is seeded but un-dispatched
- **Source:** `journal/2026-05-11-substrate-phase-c-shipped.md`
- **Effort:** small (≤30 min) for a `covenant_check_dispatch` SQL fn; medium if we wire it into the maturity ladder as another auto-fire moment
- **Risk:** Phase C.6 added the `covenant_check` template + bgworker auto-fire path but nothing calls it. The template is ready; the dispatch function and the trigger point aren't.
- **Fix v1:** ship `stewards.covenant_check_dispatch(work_item_id)` as a manual entry point + a stewards-ui "Run covenant check" button on WorkItemDetail. Decide later whether to auto-fire it on stage completion or leave it human-triggered.
- **Recommendation:** decide-then-build. The auto-fire-vs-manual question is a real ratification, not a code question.

---

## III. Missing infrastructure (subsystems set, action stubs)

Both Phase D and Phase F set `promoted_to` columns but no actual file write happens.

### 3.1 File-write mechanism for promoted lessons + resolutions
- **Source:** `phase-d-design.md` § VI, `phase-f-design.md` § VI, journal entries D + F
- **Effort:** medium (1 session) for the simplest approach (pending-write table + sidecar that materializes on `git commit`); large for any cleaner approach (host-mount + plpython3u, etc.)
- **Risk:** "Approve & promote → .mind/principles.md" buttons in Lessons.vue + "Accept + promote to .mind/decisions.md" in CouncilDetail.vue both set the database column but produce no file write. Without this, lessons stay substrate-only and the gospel framework's `.mind/principles.md` never grows from substrate experience.
- **Approaches:**
  - (a) **Pending-write table** — substrate INSERTs into `stewards.pending_file_writes (path, content, requested_at, materialized_at)`. A small CLI command (`stewards-cli materialize-writes`) appends the queued content to the target files. Run manually before each `git commit`, or wire into the pre-commit hook.
  - (b) **Sidecar daemon** — bridge container watches the pending_file_writes table via NOTIFY and writes immediately. More moving parts; substrate stops being FS-stateless.
  - (c) **plpython3u** — pg writes files directly. Fast but introduces a new extension dep + tighter coupling.
- **Recommendation:** (a). Matches the "substrate stays stateless on FS" principle from both sub-specs.

### 3.2 First real council convene + run end-to-end
- **Source:** `journal/2026-05-11-substrate-phase-f-complete.md`
- **Effort:** medium (1 session) — pick intent, design members, run, watch, accept
- **Risk:** Phase F is fully built but never exercised with an LLM. Cost ~$0.04-0.10 per council. Synthetic smoke verified D-F1 enforcement + bishop_eligible logic but the room hasn't filled with real responses.
- **Recommendation:** pick a low-stakes intent (or create a new "evaluate workflow X" intent) so an agent bishop can also be tested. Two or three members. Watch the live deliberation in `/councils/:id`.

### 3.3 First real Atonement-on-quarantine end-to-end
- **Source:** `journal/2026-05-11-substrate-phase-d-shipped.md`
- **Effort:** small (≤30 min) once 2.2 above lands — just need a quarantined work_item to sacrifice
- **Risk:** Phase D atonement is wired symmetrically with sabbath, so the auto-fire path is verified by D.4's sabbath test. But the actual atonement output (principles + decisions + lessons across 3 arrays) hasn't been observed on a real failure.

### 3.4 Trust state needs to be exercised by real work_items
- **Source:** `journal/2026-05-11-substrate-phase-e-shipped.md`
- **Effort:** falls out naturally from any real study run that reaches verified
- **Risk:** synthetic counter increments proved the promotion + demotion logic; the actual gate path that increments trust hasn't fired in production. First study run with the trust gate enabled will surface any wiring bugs.

---

## IV. Cleanups (low-risk, low-priority, do when convenient)

### 4.1 `bgworker.payload._kind` enum refactor
- **Source:** journals D, E, F
- **Effort:** small (≤1 hour)
- **Risk:** none. Bgworker now switches on 7 marker booleans (`_gate_eval`, `_scenarios_gen`, `_verify`, `_sabbath`, `_atonement`, `_council_member`, `_council_synthesize`). Every variant has slightly different shape (council_member needs role, synthesize doesn't have work_item) so the case-by-case is justified for now — but a `_kind` enum + match would be cleaner.
- **Recommendation:** defer until the 8th marker lands or a real bug forces the refactor. Today's switch isn't broken; refactoring it now is premature.

### 4.2 Stewards-UI sidebar grouping (route count = 14)
- **Source:** journals E + F + sub-spec phase-e-design.md § VI
- **Effort:** medium (1 session)
- **Risk:** none, but the nav header is genuinely cluttered. Routes today: Dashboard, Studies, Work items, Sessions, Watchman, Bridge, Graph, New work, Intents, Covenant, Sabbath, Lessons, Trust, Councils.
- **Recommendation:** group as **Substrate** (Intents, Covenant, Sabbath, Lessons, Trust, Councils) / **Surfaces** (Work items, Sessions, Watchman, Bridge) / **Records** (Studies, Graph) / **Action** (New work). Move from horizontal nav to a left sidebar.

### 4.3 Token-cost audit of compose_system_prompt injection
- **Source:** `journal/2026-05-11-substrate-phase-c-shipped.md`, sub-spec phase-c-design.md § VI
- **Effort:** small (just measure)
- **Risk:** Phase C.4 injects ~600 tokens of covenant + intent into every dispatched chat. Predicted to be acceptable; never measured on a real workload. If it spikes the cost panel on a long study run, add `compose_system_prompt(skip_covenant=true)` for stage chats that don't benefit from re-stating commitments mid-loop.
- **Recommendation:** check the cost panel after the first real study run completes. Decide based on data.

### 4.4 YAML edits don't trigger work_item dispatch refresh
- **Source:** `journal/2026-05-11-substrate-phase-c-shipped.md`
- **Effort:** none — documenting intentional behavior
- **Risk:** none. If you edit `intent.yaml` mid-flight, the substrate's intent row gets updated in place; existing work_items pick up the new values on next dispatch via `compose_system_prompt`'s fresh query. Documented here so future-self doesn't think it's a bug.

### 4.5 `gh_issue_create` grant for study agent (deferred from 3d)
- **Source:** `journal/2026-05-09-pg-ai-stewards-3d-fetch-md-v2-3f.md`
- **Effort:** small (single grant + smoke)
- **Risk:** study agent can already use git (`gh_pr_create` etc.) but not `gh_issue_create`. Issue creation has higher blast-radius (publicly visible). Defaulted to NOT granted in 3d. Worth revisiting once a study run actually wants to file an issue.
- **Recommendation:** wait for an organic need.

### 4.6 SSE live tail for work_queue
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Effort:** large (~2 hours infra)
- **Risk:** none. 5s dashboard polling covers the same use case for now. Worth doing only when we want CouncilDetail-style live deliberation streaming on more surfaces.

### 4.7 Watchman finding-ack action
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Effort:** small (~30 min — POST /api/watchman/findings/ack)
- **Risk:** read-only watchman page in v1; ack/dismiss actions skipped. Worth adding when Michael actually uses the UI to triage findings (vs. reading them in passing).

### 4.8 Bridge refresh-tools button
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Effort:** small (~45 min — POST /api/bridge/refresh-tools triggers `bridge refresh-tools` via NOTIFY)
- **Risk:** read-only bridge page. The action exists in the bridge daemon CLI; surfacing it in UI is convenience, not correctness.

### 4.9 Substrate-promoted studies don't have AGE citation graph
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Effort:** medium (lives in 3c.3.5 follow-up territory)
- **Risk:** `/api/graph` returns 0 edges for substrate--*.md studies because the promotion pipeline doesn't extract citations. Workspace-imported studies do show edges. Cosmetic for now; real value comes when substrate studies start citing each other.

### 4.10 work_item write actions on WorkItemDetail (advance/cancel)
- **Source:** `journal/2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **Effort:** small (~30 min)
- **Risk:** read-only WorkItemDetail today. SQL fns exist (`work_item_advance`, etc.); adding buttons is mostly UI work.

---

## V. Future evolution paths (decide when relevant)

These are deliberate "v2" paths the sub-specs flagged for revisit after the substrate gets lived-with. None are blocking; all involve real ratification choices.

### 5.1 F2 evolution: `council_authority` as separate trust dimension
- **Source:** `phase-f-design.md` § V.5, journal F
- **Effort:** medium (1 session) when ready
- **Risk:** today's `bishop_eligible` requires master-on-pipeline-of-intent for agent bishops. Michael's 2026-05-11 nuance flagged that bishop authority is a *cultivated skill* (different from execution authority). Debug agent is the candidate first cultivator because its skills are designed to get at the root.
- **Recommendation:** revisit after first real council with an agent bishop attempt — see whether master-on-pipeline maps cleanly to bishop facilitation in practice.

### 5.2 D-F1 concurrency lift criterion
- **Source:** `phase-f-design.md` § VI
- **Effort:** small (just a `trust_thresholds`-style change) once we know if it's needed
- **Risk:** today's `one_active_council` index allows exactly 1 concurrent. Lift to 3 if `>5 refusals/week` (per the sub-spec heuristic). Need actual usage data.

### 5.3 Atonement cost guardrail (separate cap?)
- **Source:** `phase-d-design.md` § VI
- **Effort:** small once decided
- **Risk:** Atonement on a long failure history is one of the larger prompts in the substrate. Today it shares the per-work_item cost cap. After observing cost on real atonements, decide whether it deserves its own cap (probably not — work_item cap should suffice).

### 5.4 Lesson de-duplication signal
- **Source:** `phase-d-design.md` § VI
- **Effort:** medium
- **Risk:** if the same insight surfaces in 5+ atonement events, that's a substrate-meaningful signal. Today nothing detects it. Worth instrumenting once we have ≥10 ratified lessons to look at.

### 5.5 Override weighting (advance vs. revise vs. surface)
- **Source:** `phase-e-design.md` § VI
- **Effort:** small (per-action multiplier in trust_thresholds)
- **Risk:** today every override counts as full-weight failure. Should "I think you should have surfaced this" weigh differently from "I think you should have advanced this"? Recommended equal weight initially; revisit if the trust signal feels noisy.

### 5.6 First-completion bootstrap friction (5 surfaces to escape trainee)
- **Source:** `phase-e-design.md` § VI
- **Effort:** small (lower the threshold in `trust_thresholds`)
- **Risk:** every new (agent, pipeline, model) cell starts trainee. With 5 surfaces required, the first 5 work_items per cell are heavy on human attention. Recommended keep at 5 initially; lower if it feels heavy after a month.

### 5.7 System-suggested council binding question (specificity)
- **Source:** `phase-f-design.md` § VI
- **Effort:** small
- **Risk:** today `suggest_councils` returns the cluster (pipeline, stage, sample lessons) but doesn't propose a binding question. The Vue Convene-from-suggestion handler synthesizes one from the lesson count, but it's a simple template. Worth letting an LLM compose a sharper binding question once we see real clusters.

### 5.8 What if all council members error?
- **Source:** `phase-f-design.md` § VI
- **Effort:** small once decided
- **Risk:** today the bgworker auto-fires synthesize when 0 proposer/critic members have `completed_at IS NULL`. If a member errored (work_queue.status='error'), `completed_at` stays NULL forever and the council hangs. Recommend: treat error as completion-with-empty-response so synthesize fires with whatever succeeded.

### 5.9 Bishop dispatch path for agent bishops
- **Source:** `phase-f-design.md` § VI
- **Effort:** medium
- **Risk:** today, when bishop is an agent identifier, the resolution still requires a human to click Accept in the UI. For agent bishops to actually work, need a `_council_bishop_dispatch=true` chat path that runs the agent through the resolution decision and writes back. Deferred until F2 lands.

---

## VI. Substrate-wide infrastructure debt

### 6.1 Lesson #3 — proper extension version-bump strategy
- **Source:** `journal/2026-05-11-substrate-phase-c-shipped.md` + carried in every Phase journal since
- **Effort:** medium (real version-bump system) or current `bump-extension.sh` is "good enough"
- **Status:** PARTIALLY ADDRESSED. Phase C close shipped `scripts/bump-extension.sh` + skill + PostToolUse hook (`9c1ae8d`) that auto-refreshes pgrx CREATE FUNCTION registrations after `docker compose build pg`. It worked cleanly in Phases D, E, F.
- **Remaining gap:** the workaround patches `pg_proc` directly without bumping `pg_extension.extversion` or registering functions in `pg_depend` as extension members. Functional for dev iteration; slightly drifty for production.
- **Recommendation:** keep the current workaround until the substrate moves toward production. If/when that happens, build a proper version-bump system with upgrade scripts (`pg_ai_stewards--<from>--<to>.sql`) and idempotent CREATE TABLE statements throughout the bundled SQL.

### 6.2 stewards-ui sidebar grouping (see 4.2)
- Already covered above.

---

## VII. Recommended next batches

The pattern that worked for Phases C–F (decisions upfront → gated phased build with smoke at each commit) suggests grouping these into focused work sessions. Two natural batches:

### Batch G — "Make the substrate land in real files" (~1 session)
The bare minimum to take the substrate from BUILD-COMPLETE to USE-READY. All blocking-or-nearly-blocking items.

| Item | Why now |
|---|---|
| 1.1 studies.file_path NOT NULL | Without this, first sabbath-gated study completion will fail |
| 2.1 Steward retry pulls lessons | Line-upon-line never exercised otherwise |
| 2.2 Steward quarantine fires atonement | Atonement never exercised otherwise |
| 3.1 File-write mechanism (option a — pending_file_writes table + CLI materializer) | Ratified lessons + accepted council resolutions never reach .mind/ otherwise |

After Batch G: the first real Phase D + E + F end-to-end runs become possible.

### Batch H — "Use validation" (1-2 sessions of guided usage, not coding)
- 3.2 First real council convene + run
- 3.3 First real Atonement-on-quarantine
- 3.4 First study run that earns trust transitions
- 4.3 Token-cost measurement on the first real study run

This batch is mostly **observation**, not coding. The interesting outputs are: do the gates make sharp enough calls? does atonement produce useful lessons? does the trust matrix populate the way we expected?

### Batch I — "Hybrid revise" (1 session, decision-driven)
- 2.3 Hybrid revise (revise #1 same-model + feedback prepended)
- 2.4 covenant_check dispatch path

Both involve genuine ratification questions (when to escalate; auto-fire vs manual) so they fit the "decide-then-build" cadence.

### Batch J — "Polish" (whenever, low-priority)
- 4.1 bgworker `_kind` enum refactor
- 4.2 stewards-ui sidebar grouping
- 4.5 gh_issue_create grant
- 4.6 SSE live tail (only if real volume warrants)
- 4.7 Watchman finding-ack action
- 4.8 Bridge refresh-tools button
- 4.9 Substrate-promoted studies citation extraction
- 4.10 work_item write actions on WorkItemDetail

### Section V items — "Wait for usage data"
None of these should ship before observing real substrate behavior. Revisit after Batches G + H produce actual data.

---

## VIII. How to use this document

- **Update after each session.** When an item ships, mark it ✅ here and remove from the active list. When a session surfaces a new item, add it under the right theme.
- **Don't treat as a TODO list.** This is a *menu*. Items only become TODOs when Michael picks a batch.
- **The phase journals are still authoritative.** This document is a navigation aid; the journal entries carry the full context for why each item exists.
- **Living document.** Date-stamped at top; bump the date on every edit.

---

## X. Items from older proposals + phases.md (added 2026-05-11 revision)

First pass of this document focused on Phase C–F sub-specs + recent journals. Michael flagged that I missed the older `.spec/proposals/pg-ai-stewards-*` proposals and `projects/pg-ai-stewards/phases.md` — this section captures what those add.

### X.1 Multi-provider expansion (Phase 3g)
- **Source:** `phases.md` line 1201 ("Phase 3g — Multi-provider expansion: Anthropic, Gemini, Veo, TTS")
- **Status:** not started
- **Effort:** medium per provider (1 session each)
- **Risk:** today's substrate routes everything through `opencode_go` (with `lm_studio` as fallback). Adding native Anthropic / Gemini / Veo / TTS providers would mean: provider registry rows, response-shape normalization (each provider's `usage` shape differs), agent variant rules (model_match patterns), and possibly tool-shape adapters.
- **Why deferred:** OpenCode Go's Chinese-model gateway (Kimi K2.6, GLM-5.1, MiniMax M2.7, Qwen3.6 Plus) covers the cost-discipline scenario. Anthropic via OpenCode Zen handles the "boost to opus" case via the human-mediated escalation queue. Native integrations are a v2 question.
- **Recommendation:** **revisit when there's a concrete need** (e.g. Gemini gives uniquely useful output for some pipeline; TTS for the YouTube workflow). Don't build speculatively.

### X.2 Per-model prompt tuning generalization (Phase 3h family)
- **Source:** `proposals/pg-ai-stewards-per-model-prompt-tuning.md` (entire doc — status: deferred)
- **Status:** prototype shipped 2026-05-08 (`.stewards/kimi-k2.6/study.agent.md` + `.stewards/qwen-3.6/study.agent.md`); full effort deferred until 3e+3f land (both have shipped)
- **Effort:** large (4 phases × 1-2 sessions each = ~6-8 sessions total)
- **Risk:** prototype validates on only ONE binding question (FtC/WtL). Cross-topic validation hasn't run. Without it we can't claim the tuned variants improve study quality across the corpus — only that they did on that specific question.
- **Sub-phases per proposal:**
  - **3h.1** Cross-topic validation suite — re-run base + tuned variants of kimi + qwen study agents on three structurally distinct binding questions (focused exegesis / character study / modern-prophet talk analysis). Confirm tuned variants improve, regress, or are neutral on each.
  - **3h.2** Extend variant authoring to other agents (`lesson`, `talk`, `journal`, `research-gospel`). Each variant: baseline run + signature identification + tuned authoring + validation.
  - **3h.3** Onboard additional models (Sonnet, Gemini, GLM, etc.). Each new model runs the cross-topic suite as both base and tuned.
  - **3h.4** Migrate to `study-bench` CLI (mirroring `classify-bench` shape) so model evaluation is repeatable + produces structured comparison artifacts.
- **Project memory references:** `project_kimi_voice_signatures.md`, `project_qwen_voice_signatures.md` — the validated signature catalogs from the prototype.
- **Recommendation:** **wait until Batch H produces real study runs** through the new substrate. That will give 3h.1 cross-topic validation actual cross-topic data (instead of being a synthetic exercise).

### X.3 Watchman soak — the 7-day continuous run never happened
- **Source:** `phases.md` lines 784 + 1087-1090 ("The soak itself (the third deliverable per the original 2.7b.4 plan) is runtime observation, not code; starts when schedule_enabled=true is flipped on for a sustained period")
- **Status:** soak has been **running intermittently** since 2026-05-06 (paused for build sessions, re-enabled at session ends). The intended **7-day continuous observation period** never happened — every Phase A–F build session paused it.
- **Effort:** zero coding; just leave `schedule_enabled=true` for 7 consecutive days
- **Risk:** the soak was meant to surface watchman runtime behaviors (pressure schedule firing pattern, dirty queue convergence rate, token-budget actual spend vs. configured cap) on real corpus traffic. Without a sustained observation period we have no empirical answer to "is the watchman cadence right?"
- **What we learned anyway:** the existing intermittent soak shipped watchman passes at the expected pace. No catastrophic behavior. Token budget worked. But we don't have the longitudinal data the soak was designed to produce.
- **Recommendation:** **start a real 7-day soak after Batch G ships** (the file-write mechanism is the last missing piece for substrate "completeness"). Schedule a week with no substrate build sessions, monitor the dashboards, journal observations.

### X.4 AGE upstream contributions (Phase 6)
- **Source:** `phases.md` lines 1418-1474, full catalog in `projects/pg-ai-stewards/docs/AGE-QUIRKS.md`
- **Status:** 8 AGE quirks catalogued during Phase 2.6 work; no PRs filed
- **Effort:** large (per PR: working test environment + familiar codebase + reviewer cycles)
- **Risk:** AGE is Apache-governed; PR cycles are slow. Don't block our own work waiting on upstream.
- **PR-worthy bug candidates (3):**
  - #2 Apostrophe-in-interpolated-Cypher error message + auto-escape
  - #6 `cypher()` 3rd-arg should accept any `ag_catalog.agtype` expression
  - #7 `#>>` (and likely `->>`, `->`) should handle agtype scalars as pass-through
- **Document-only (3):** quirks #1, #3, #5 (spec-divergences)
- **By-design (1):** quirk #4 (labels as schema)
- **Our problem (1):** quirk #8
- **Recommendation:** **defer until the substrate hits steady state.** This is "give back when delivery isn't pulling cycles." When that happens, start with #2 (cheapest PR) before #6 / #7 (deeper changes to agtype type system).

### X.5 Phase 4 — GraphRAG over the canon (optional)
- **Source:** `phases.md` lines 1393-1416
- **Status:** explicitly optional ("only if Phases 1-3 surface the need")
- **Effort:** large (Microsoft GraphRAG indexing run + AGE schema collaboration with gospel-engine-v2 + new MCP tool)
- **Risk:** none — explicitly speculative. We won't know if we need it until we've used Phases 1-3 (now A-F) for a few months.
- **Recommendation:** **don't build it.** Revisit only if a real "themes across the whole corpus" question fails with current substrate + gospel-engine-v2 tools.

### X.6 Phase 7+ "Maybe-someday"
- **Source:** `phases.md` lines 1475-1518
- Items:
  - **`postgres_fdw` from stewards into gospel** — SQL-level joins instead of HTTP. Trigger: if a substrate query needs to join gospel-library data with stewards data repeatedly, FDW would be faster. Not pressing.
  - **Multi-tenant RLS** — if ibeco.me ever hosts other people. Has detailed sub-questions catalogued in phases.md lines 1481-1518 (workstream-as-sharing-unit, owner_user_id + visibility columns, AGE Cypher RLS caveat).
- **Recommendation:** **don't build either.** Single-user substrate has no real shape for these.

### X.7 Phase 2.6 typed edges + Workstream + Todo schema
- **Source:** `proposals/pg-ai-stewards-phase-2-5-generic-substrate.md` lines 327-460
- **Status:** spec'd 2026-05-04; 2.6a (workstreams) + 2.6b (todos) + 2.6c (phases-context) all shipped (referenced in `extension/2-6a-workstreams.sql` etc.)
- **Open from the spec:** `stewards.todo_rollup_audit()` was specified; verify it landed and runs cleanly. If not built, audit is a real gap for "parent done with open children" correctness checks.
- **Action:** verify in a smoke session whether `stewards.todo_rollup_audit()` exists + returns sensible data. If missing, small fix.

### X.8 Phase 1.7 — Brain CLI driver + hybrid FTS+vector search (deferred)
- **Source:** `phases.md` lines 119-131
- **Status:** deferred 2026-05-03 ("paired work; together they form Phase 1.7 if/when we revisit")
- **Effort:** medium-large
- **Risk:** today's brain CLI uses the SQLite driver as read-only fallback; the Postgres backend exists but isn't wired through brain's CLI surface. Hybrid FTS+vector search across brain entries was deferred until the embedding column had real traffic.
- **Recommendation:** **revisit only if SQLite-as-brain-backend starts hurting.** It hasn't yet.

### X.9 3d.2 Docker sidecar wrapper + 3d.3 safe_outputs proxy
- **Source:** `proposals/pg-ai-stewards-3d-sandboxed-git.md` lines 327-338
- **Status:** 3d v1 (Option A — native Go MCP wrapper) shipped 2026-05-09. 3d.2 (Docker sidecar) + 3d.3 (safe_outputs proxy) are layered improvements deferred.
- **Effort:** medium each
- **Trigger criteria** (from sub-spec):
  - **3d.2 Docker sidecar:** ship if/when 3d v1 surfaces an escape risk worth more isolation, OR if we want multi-pipeline parallel git work (one workdir per pipeline → containers prevent cross-contamination)
  - **3d.3 safe_outputs proxy:** ship if/when 3d.1 + 3d.2 are stable enough that the buffer-and-vet flow (GitHub Agentic Workflows pattern) has something to wrap
- **Recommendation:** **wait for the trigger.** v1 has been adequate. The next prompt to revisit is when a study agent actually wants to write to a repo at scale.

### X.10 GITHUB_TOKEN setup (PAT) for 3d v1
- **Source:** `phases.md` line 1187 ("PAT setup deferred per Michael; live test triggers when GITHUB_TOKEN lands in .env")
- **Status:** deferred per Michael; no live git-mcp test has run because the token isn't in .env
- **Effort:** small (~5 min — generate PAT, add to .env, restart bridge)
- **Risk:** 3d v1 is theoretically functional but unverified end-to-end. First real attempt to use git-mcp will surface any wiring bugs.
- **Recommendation:** when Michael wants the substrate to start opening PRs.

### X.11 3f UI extensions (write actions + cloud)
- **Source:** `proposals/pg-ai-stewards-3f-local-ui.md` lines 248-260, also surfaced in `2026-05-09-stewards-ui-phases-2-7.md` carry_forward
- **3f.3 — Write actions beyond pipeline:** edit agents, edit tool_defs, broadcast a message into a session. Triggers when Michael wants to manage substrate state visually instead of via psql.
- **3f.4 — Multi-user / cloud:** when/if shared substrate or `a.ibeco.me`-style hosting is wanted. Original cloud spec preserved at top of 3f proposal.
- **Recommendation:** **3f.3 is a real "use is asking for it" question** — defer until you find yourself running `psql -c "UPDATE stewards.agents …"` more than 3 times in a session. 3f.4 requires a multi-user shape we don't have.

### X.12 3c.4 — HTTP tools for gospel-engine-v2 (Path A)
- **Source:** `proposals/pg-ai-stewards-3c-2-5-study-tools.md` line 325 + carry-forward §339
- **Status:** named in 3c-2-5 carry-forward; "deferred to 3c.4 or later. Cleanest after 3c.3 demonstrates the pipeline pattern works on B." Phase 3e absorbed this work via MCP-proxy (substrate agents now reach gospel-engine via the mcp_proxy tool dispatcher). 
- **Action:** **probably resolved by 3e.** Verify by checking whether a substrate agent's gospel_search call goes through bridge → mcp_proxy or through some hypothetical 3c.4 HTTP path. Spot-check next session.

### X.13 Inline open questions in the foundational proposal
- **Source:** `proposals/pg-ai-stewards-11-cycle-review.md` lines 225-255 (5 open questions Michael answered as part of full-agentic-substrate.md §VI ratification on 2026-05-10)
- **Status:** **resolved.** The 5 open questions in the 11-cycle review (which phase first / steward as bgworker module / gate eval model / 0-dirty-docs goal / per-pipeline covenant) all map to ratifications captured in full-agentic-substrate.md §VI or the 2026-05-11 amendment block. Nothing new to add.

### X.14 Things explicitly NOT in scope (preserved for the menu)
These are listed as deliberate exclusions in older proposals; included here so they're visible if the substrate's needs change:

- **3c.2.5:** no `study_show` exposure as a tool (use `study_get` + `study_citations` + `study_similar` trio instead); no write-side tools (substrate's pipeline terminal stage inserts via `import_study()` directly); no embedding-on-the-fly for arbitrary user query vector search.
- **3h (per-model tuning):** no variants for watchman-consolidator agent (already tuned); no variants for embedding models.
- **3d:** never expose `git_tag` or `git_push --tags` (branch protection bypass vector). Workdir paths strictly anchored.

---

## XI. Revised Batch recommendations

The first pass recommended 4 batches (G/H/I/J). Adding items from older proposals doesn't change the top of the queue — Batch G is still "make the substrate land in real files." But it adds two new candidate batches further down:

### Batch K — "Soak validation week"
After Batch G ships, schedule **one calendar week with no substrate build sessions**. Leave `schedule_enabled=true`. Observe the dashboards. Journal what you see. This is X.3 — the 7-day soak that was supposed to happen 2026-05-06 but got displaced by every Phase build session.

What to watch:
- Watchman pass cadence (pressure vs. cron mix)
- Dirty queue convergence (does it actually approach 0, or stay stuck?)
- Token-budget actual spend vs. configured cap
- Any unexpected errors in `stewards.work_queue` or `steward_actions`

### Batch L — "Per-model voice tuning generalization" (3h family)
Wait until Batch H has produced 2-3 real study runs on the new substrate (with intent + covenant + gates + sabbath). Then start 3h.1 — cross-topic validation of the kimi + qwen tuned variants on those binding questions. Effort: large (6-8 sessions across 3h.1 through 3h.4). Decisive question: do the tuned variants generalize, or were they FtC/WtL-specific?

### Things to actively NOT do
- Phase 3g (multi-provider) — speculative; only build when a specific provider is uniquely useful
- Phase 4 (GraphRAG) — only if real "themes across the corpus" need surfaces
- Phase 6 (AGE upstream) — only after substrate is in steady state
- Phase 7+ items — single-user substrate has no shape for these

---

## XII. What I missed in the first pass

For honesty: my first pass of this document covered Phase B–F sub-specs + recent journals but missed the foundational proposal + `phases.md`. The items I missed (now captured in Section X):
- Phase 3g (multi-provider expansion) — was in phases.md
- Phase 3h family (per-model prompt tuning) — has its own deferred-status proposal
- The 7-day Watchman soak — referenced repeatedly in phases.md as "pending start"
- Phase 6 (AGE upstream) — sketched in phases.md
- Phase 4 (GraphRAG) — sketched in phases.md
- Phase 7+ stretch items (postgres_fdw, multi-tenant RLS) — listed in phases.md
- 3d.2 + 3d.3 (git-mcp follow-ons) — deferred in 3d proposal
- 3c.4 (gospel-engine HTTP path) — possibly resolved by 3e, verify
- GITHUB_TOKEN setup deferred per Michael
- 3f.3 + 3f.4 (UI write actions + cloud) — flagged in 3f proposal
- todo_rollup_audit() — spec'd in 2.6b, need to verify it shipped

Process note: the journal entries we shipped per-phase do a good job of capturing per-session carry-forward, but the *standing backlog* lives in `phases.md` and the proposals — those need a separate sweep. Worth doing every major-phase milestone (which this is) rather than letting them drift further from awareness.

---

## XIII. Glossary of source documents

Sub-specs (phase build plans, in `projects/pg-ai-stewards/.spec/proposals/`):
- `cost-tracking.md` — Phase A cost layer
- `escalation-chain.md` — Phase A model escalation
- `steward-bgworker-integration.md` — Phase A bgworker tick
- `phase-c-design.md` through `phase-f-design.md` — phase-by-phase build specs
- `full-agentic-substrate.md` — the original 6-phase proposal + §VI ratification + 2026-05-11 amendment block

Older proposals (in `.spec/proposals/pg-ai-stewards-*`):
- `pg-ai-stewards-phase-2-5-generic-substrate.md` — foundational: studies / AGE citations / workstreams / todos / typed edges
- `pg-ai-stewards-11-cycle-review.md` — original 11-cycle gap analysis that morphed into full-agentic-substrate
- `pg-ai-stewards-3c-2-5-study-tools.md` — 5 MCP tools for study agents (study_search_text, study_get, etc.)
- `pg-ai-stewards-3d-sandboxed-git.md` — git-mcp wrapper with allow-list discipline
- `pg-ai-stewards-3f-local-ui.md` — local Vue UI proposal (stewards-ui)
- `pg-ai-stewards-per-model-prompt-tuning.md` — Phase 3h family, status=deferred

Phase tracking:
- `projects/pg-ai-stewards/phases.md` — the canonical phase tracking doc (1539 lines; Phases 0-7+)

Recent journal entries (substrate work):
- `2026-05-09-full-agentic-substrate-proposal.md` — initial research session
- `2026-05-10-substrate-phase-a-specs.md` — A ratification + sub-spec drafting
- `2026-05-10-substrate-phase-4a-schema-live.md` — A schema layer
- `2026-05-11-substrate-phase-b-feature-complete.md`
- `2026-05-11-substrate-phases-cdef-revalidation.md` — re-validation + 4 sub-specs
- `2026-05-11-substrate-phase-c-shipped.md`
- `2026-05-11-substrate-phase-d-shipped.md`
- `2026-05-11-substrate-phase-e-shipped.md`
- `2026-05-11-substrate-phase-f-complete.md`

Journal entries from earlier substrate work (Phases 2.7 / 3.x / 4):
- `2026-05-04` through `2026-05-09` — these mostly carry items that have since been folded into the recent C–F journals; spot-checked, no new items extracted.

Project memory:
- `project_pg_ai_stewards_revise_hybrid.md` (2026-05-11) — hybrid revise decision
- `project_pg_ai_stewards_state.md` — substrate state snapshot
- `feedback_pg_ai_stewards_rebuild_discipline.md` — image rebuild discipline
