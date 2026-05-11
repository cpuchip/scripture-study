---
date: 2026-05-11
session_kind: planning
workstream: WS5
substrate_phase: C–F (planning)
related_commits: TBD
cost_usd: 0
---

# Substrate Phases C–F — re-validation + sub-specs

## What happened

Michael came back after the Phase B feature-complete close-of-session and said "I like your hybrid approach! lets do that, I think it's totally the best of both worlds. lets review each phase now C-F and hit me with the decision points so we can get those phases planned out for quick development like we just did."

Two pieces of work in this session: (1) walk through the §VI ratifications for C/D/E/F against A+B's lived experience, (2) draft per-phase build-ready sub-specs.

The hybrid revise approach (from prior session's close) got formally ratified and saved as project memory: revise #1 stays same model + feedback prepended to the stage prompt; revise #2 escalates to next-tier model AND keeps the feedback; cap at 2 → surface (D-B2 unchanged). Best-of-both: cheap focused critique gets one shot before the steward escalates.

## What was already ratified

§VI of `full-agentic-substrate.md` had all 22 decisions ratified on 2026-05-10 in a post-iron-rod walk-through. Michael may not have remembered. I surfaced this and offered four paths — re-validate per-phase, skip-and-spec, implementation-only-decisions, or read-and-then-decide. He picked re-validate per-phase. Right call: A+B lived experience surfaced refinements to several decisions.

## Amendments captured (13 changes from original §VI)

The proposal §VI now carries an "2026-05-11 re-validation amendments" section noting:

- **Phase B revise hybrid** (NEW) — the carry-forward from yesterday now formal.
- **D-C4 revised** — covenant gate prompt stays free-form but tools=off (Phase B's cost-surprise informs).
- **D-D Sabbath blocks promotion** (NEW) — `work_item_promote_to_study` raises if sabbath hasn't run for sabbath-enabled pipelines. The discipline is endings recorded.
- **D-D Atonement tools off** (NEW) — same fix as D-C4.
- **D-D Lessons schema** (NEW) — `stewards.lessons` mirrors `gate_decisions` audit-ledger shape (kind column, single content per row) over the proposal's structured triplet jsonb. Stewards-UI already knows how to render this pattern.
- **D-E trust keying** (revised) — `(agent_family, pipeline_family, model)` over the proposal's `(agent_family, pipeline_family)`. Recognizes model-specific competence.
- **D-E promotion signal** (NEW) — successful completion = maturity reached verified, not just `status='done'`.
- **D-E retry lessons** (NEW) — pulls last 3 ratified lessons for `(pipeline, stage)`.
- **D-F2 nuance** (NEW) — F1 ships master-on-pipeline-of-intent for agent bishops; future evolution path introduces `council_authority` as separate trust dimension with debug agent as candidate first cultivator. Michael's words: "the problem here is that we may not have a master level agent there. so I'm leaning [option 3] and trying to cultivate that as an agent or skill that can be called in. Debug agent might be good here because it's skills are designed to get at the root."
- **D-F low-stakes definition** (NEW) — `intent.scripture_anchor IS NULL AND values_hierarchy lacks doctrinal/spiritual/discernment`. Doctrinal intents always require human bishop.
- **D-F member keying** (NEW) — `(council_id, agent_family, role)` — model floats per dispatch.
- **D-F watchman suggestions** (NEW) — surface in BOTH watchman pass output AND Stewards-UI dashboard banner.
- **Seed mechanism** (NEW) — manual SQL fns + git pre-commit hook. Explicit, debuggable, no daemon to manage.

## What got built

Four build-ready sub-specs at `projects/pg-ai-stewards/.spec/proposals/`:

- `phase-c-design.md` — Intent + Covenant: schema, `seed_intents_from_yaml` + `seed_covenant_from_yaml` SQL functions, git pre-commit hook, `compose_system_prompt` extension, gate prompt revision + new `covenant_check` template (tools off), `work_items.intent_id` FK, NewWork form gain. **3 sessions estimated.**
- `phase-d-design.md` — Atonement / Sabbath / Consecration: `pipelines.sabbath_enabled` + `atonement_enabled`, `stewards.lessons` audit ledger, `sabbath_dispatch` + `atonement_dispatch` + apply functions, bgworker auto-fire extension (2 more markers), `work_item_promote_to_study` gating, Stewards-UI Sabbath Log + Lessons Review. **3 sessions estimated.**
- `phase-e-design.md` — Trust + Line upon Line: `trust_scores` + `trust_transitions` + `gate_overrides`, `evaluate_trust` SQL fn, gate behavior change (trainee surfaces every advance), retry composer extension pulls last 3 ratified lessons, Stewards-UI Trust Matrix. **3 sessions estimated.**
- `phase-f-design.md` — Council: `councils` + `council_members` + `resolutions`, `convene_council` + `synthesize_council` + `resolve_council` SQL fns, `bishop_eligible` (with low-stakes-def + future-`council_authority` note), watchman convening suggestions, Stewards-UI Council "feels-like-a-room" view. **4–5 sessions estimated.**

**Total roadmap C–F: ~13 sessions** for the substrate to reach Zion completeness.

## Surprises

**The proposal needed less re-litigation than expected.** Of 13 amendments, only D-C4 was a true revision (free-form tools-on → free-form tools-off). The other 12 are NEW decisions that §VI didn't cover at all — implementation-level questions that A+B lived experience made obvious. The "re-validate" path was right; an implementation-only-decisions path would have been thinner.

**The tools-cost lesson propagates everywhere.** Every JSON-output gate prompt across C and D explicitly says `tools_disabled: true`. Phase B's 5× cost surprise from a single gate-eval (qwen3.6-plus loop-researching the corpus before deciding) is now a substrate-wide pattern. One incident, one lesson, applied retroactively — exactly the line-upon-line discipline Phase E builds infrastructure for.

**F2 was the most theologically interesting moment.** Michael's "debug agent might be good here because its skills are designed to get at the root" surfaces a real possibility that wasn't in the original proposal: bishop authority as a *cultivated skill*, not just a derived trust level. This is a future evolution path — F1 ships with master-on-pipeline-of-intent — but the spec carries the breadcrumb. Worth watching whether the debug agent (or a new bishop agent) actually develops the facilitation pattern over time.

**Stewards-UI is getting busy.** Counted in phase-e sub-spec: by end of F we'd have `/dashboard`, `/work-items`, `/sessions`, `/watchman`, `/bridge`, `/studies`, `/graph`, `/new`, `/intents`, `/covenants`, `/sabbath`, `/lessons`, `/trust`, `/councils`. 14 routes. Open question flagged in phase-e-design.md VI: sidebar grouping (Substrate / Surfaces / Records).

## Process / covenant

Per the covenant's `update_memory` commitment — this entry + active.md update happen at session end before yielding. The pattern is: work → memory → done. Hybrid-revise was already saved as project memory at session start so it survives the next code session.

The §VI parent proposal got an in-place amendment block rather than a fresh §VII. Smaller diff, easier to review, keeps the original ratifications visible. The amendment block is dated and explicitly scoped to "after Phase B feature-complete" so future readers know what informed the changes.

The four sub-specs follow the format established by `cost-tracking.md` and `escalation-chain.md` — frontmatter (title/date/status/parent/purpose), Roman-numeral sections (binding problem, success criteria, constraints, prior art, proposed approach with sub-numbered V.x, open questions, programming time, acceptance scenarios). Consistency means the build sessions know where to look.

## Open / carry-forward

- **Build C first, lived-with before D.** Keep the established cadence — ship a phase, live with it, ship the next. Don't build C+D+E+F in a sprint.
- **YAML parser helper for Phase C.** Decide whether to add `serde_yaml` to extension Cargo.toml or use plpython3u. Recommended in spec: `serde_yaml` (substrate stays self-contained, no plpython3u dep).
- **Stewards-UI sidebar grouping** — defer until phase F lands and the route count actually hurts.
- **Atonement long-failure-history truncation** — flagged in phase-d-design.md VI; 20-entry cap on steward_actions_summary.
- **Resolution promotion-to-study mechanism** — same pending-file-write pattern as lesson promotion. Substrate stays FS-stateless.

## Files touched

- `projects/pg-ai-stewards/.spec/proposals/full-agentic-substrate.md` — amendment block in §VI
- `projects/pg-ai-stewards/.spec/proposals/phase-c-design.md` (new, ~280 lines)
- `projects/pg-ai-stewards/.spec/proposals/phase-d-design.md` (new, ~310 lines)
- `projects/pg-ai-stewards/.spec/proposals/phase-e-design.md` (new, ~310 lines)
- `projects/pg-ai-stewards/.spec/proposals/phase-f-design.md` (new, ~330 lines)
- `.mind/active.md` — entry added for 2026-05-11 planning session
- `C:\Users\cpuch\.claude\projects\…\memory\project_pg_ai_stewards_revise_hybrid.md` (new, project memory)
- `C:\Users\cpuch\.claude\projects\…\memory\MEMORY.md` — index entry added
