# Audit: the 11-step blueprint vs. the pg-ai-stewards implementation

**Date:** 2026-06-04
**Blueprint:** `docs/work-with-ai/guide/05_complete-cycle.md` (the 11 steps) + the guide series.
**Implementation:** `pg-ai-stewards` — the Postgres-as-agent-substrate (Phases A–F + the coder + the critic harness).
**For:** the book (`projects/scripture-book` — *Beyond the Prompt*). Each lesson notes the chapter it feeds.

> **Honesty frame (Ben Test).** This is not a victory lap. The grades below are mixed on purpose — two steps are genuinely strong because they were *built* on the scripture; two were *weak until last week* and the harness work just fixed one of them; one (Sabbath) is primitive-and-underused. The headline is real but earn it: **the blueprint is mostly running, not aspirational — and the place it was weakest is exactly the place the industry is weakest (Review-against-intent).**

---

## Headline

The substrate maps to all eleven steps because it was *designed* on them — Phase C is Covenant+Intent, Phase D is Atonement+Sabbath+Consecration, Phase E is Stewardship+Line-upon-Line, Phase F is Zion. So this audit is, at bottom, **evidence for the book's own thesis**: the gospel patterns aren't metaphors laid over engineering; they're a spec a working system was built against. The most useful single sentence: *the bug that the night-build review surfaced was a missing step of the creation cycle (Review-beyond-correctness), not a missing feature.*

## Scorecard

| Step | Substrate mechanism | Grade | The live lesson |
|---|---|---|---|
| 1 Intent | `intent.yaml` → `compose_system_prompt` injects it into every dispatch | **Strong** | Intent as runtime state, not a doc |
| 2 Covenant | `covenant.yaml` composed into prompts; bilateral | **Strong** | Mutuality is encodable |
| 3 Stewardship | `agent_tool_perms` + the trust ladder (Phase E) + maturity gates | **Strong** | Progressive trust is real + dynamic |
| 4 Spiritual Creation (spec) | pipelines/stages; per-work-item `binding_question` (+ now `acceptance_criteria`) | **Was weak → now better** | Thin specs = silent gaps; the spec must carry the *whole* plan |
| 5 Line upon Line | context engine (engrams, graduated rendering, Batches K/L) | **Strong**, with a new inversion | The *sandboxed* agent can't reach for context — the steward must grant it |
| 6 Physical Creation (execution) | bgworker dispatch loop; the coder (write/build/test in sandboxes) | **Strong** | Execution is the easy 1/11; preparation sets its ceiling |
| 7 Review | gates + `verify` (correctness) → **now the critic stage** (spec + intent) | **Was weak → now strong** | "Correct implementation of the wrong thing" is the dangerous failure |
| 8 Atonement | Phase-D quarantine + learnings → **now the critic revise-loop** | **Strong** | Forward-recovery with the learning injected, not retry/revert |
| 9 Sabbath | `sabbath_dispatch` primitive + the agent's session-close discipline | **Partial / underused** | The primitive exists; it rarely *fires* — rest is still mostly the human's habit |
| 10 Consecration | spend caps ($12/day, provider caps, cost buckets — J.11) | **Strong** | Every token accountable to intent, enforced at dispatch |
| 11 Zion | Phase-F council + **now the per-stage "ward council" model roster** | **Strong** | Gifts measured (bake-off), assigned per stewardship |

---

## Per-step detail (where it matters)

### Step 4 — Spiritual Creation (the spec). Was weak.
The pipeline *is* a reusable spec, but the **per-task** spec was just a `binding_question` — one sentence. That thinness is precisely what produced the night-build gaps (a global presence tracker, vacuous tests): the spec never said "room-scoped," so nothing required it. The harness fix added an explicit `acceptance_criteria` checklist to the work-item input. **Lesson for the guide (Part 4, Spec Engineering):** a binding question is necessary but not sufficient; the executable spec needs acceptance criteria the *reviewer* can check item-by-item. Feeds Part Two ch. 03 (`spiritual_before_temporal`) and the Part One practice `p1_03_set_the_bounds`.

### Step 5 — Line upon Line. Strong, but inverted in the sandbox.
The context engine (engrams, graduated rendering) is one of the substrate's most-developed areas. But the coder surfaced an **inversion** of the principle: the guide's line-upon-line assumes the agent *demonstrates readiness and earns more context*. A sandboxed agent **cannot reach for context at all** — it sees one cloned repo, not the workspace. So the steward must *grant* the context up front (e.g., paste the auth pattern the sandbox can't open). **Lesson:** line-upon-line has two directions — context earned (trust) and context granted (reach). Isolation flips which one applies. Feeds the Part One practice `p1_04_pack_the_context`.

### Step 7 — Review. The center of this audit.
The substrate had only **layer 1** (correctness: `verify` runs build+test). The guide's own Step-7 table names three layers — Correctness / Specification / Intent — and warns that correctness-only review certifies "correct implementation of the wrong thing." That warning came true in the night build. **The critic stage now implements layers 2 and 3** as a pipeline stage: a different strong model checks the real diff against the acceptance criteria (spec) and the binding question (intent), and bounces deficiencies back. **Two lessons for the guide:** (a) the three-layer review is *buildable as a stage*, not just a human discipline; (b) **automating review raises the floor but not the ceiling** — in the bake-off, glm-5.1's latent data race passed build, vet, `-race`, *and* the critic; a human read caught it. Review-against-intent is necessary and was missing; it is still not sufficient. Feeds Part Two ch. 04 (`watched_until_they_obeyed`) directly — this is its lived example.

### Step 8 — Atonement. Strong; now with a worked loop.
Phase D already had quarantine + `.spec/learnings`. The critic revise-loop makes the guide's Atonement pattern *executable*: don't revert → name what's wrong (the learning) → inject it into the next attempt → forward-recover (fix, don't restart). And at the meta level the whole arc ran it: the night-build failure wasn't reverted; it was diagnosed (the harness was the gap), and we recovered *forward* by building the missing step. **Lesson:** Atonement in a pipeline = feedback-injected forward-recovery, capped, then escalated to a human. Feeds Part Two ch. 08 (`mechanics_of_refinement`) + ch. 11 (`the_seventh_time`).

### Step 9 — Sabbath. The honest weak spot.
The substrate has a `sabbath_dispatch` primitive, but in practice rest + reflection is still carried by the *agent's* session-close discipline and the human, not by the substrate firing it structurally. **Lesson (and a gap worth naming in the guide):** Sabbath is the hardest step to make a system *do* rather than a person *remember* — the substrate proves the others are encodable but has not yet made cessation structural. That's an honest "the industry is missing this, and so are we, partly." Feeds the guide's Step 9 with a candid status.

### Step 11 — Zion. Strong; now concrete.
Phase F built multi-agent councils. The new **per-stage model roster** ("ward council for development": m3 the architect/documenter, kimi the builder, qwen3.7-max the critic) is the guide's *Bishop-vs-Conductor* argument made concrete — autonomous stewardships, each staffed by its gift, aligned by a shared intent that flows to every stage. And the bake-off is *how you staff a council honestly*: measure the gifts, don't assume them. **Lesson:** Zion-for-development is assignment-by-measured-gift under shared intent. Feeds Part Two ch. 12 (`conclusion_zion`).

---

## Cross-cutting lessons to fold into the 11-step guide

1. **Add the "floor vs. ceiling" law to Step 7.** Automating Specification + Intent review raises the floor (no more silent dropped requirements / vacuous tests) but does not remove the human watcher (subtle soundness — glm's race — slips every automated gate). The guide currently implies review can be fully delegated; it can't.
2. **Step 4 needs acceptance criteria, not just a binding question.** The reviewer needs a checklist; the binding question alone under-specifies and produces correct-but-wrong work.
3. **Step 5 has a granted-context direction, not only an earned one.** Isolation (sandboxes, sub-agents) inverts who reaches for context.
4. **Step 8 has an executable form:** feedback-injected forward-recovery in the pipeline, capped + escalated.
5. **Step 11 staffing is measured, not assumed** — the bake-off as council-interview.
6. **Name Sabbath's difficulty honestly:** it's the one step the substrate has not made structural. Worth saying in the book — vulnerability is credibility (the teaching covenant).
7. **The meta-lesson for the book's thesis:** the substrate is the blueprint *running*. The strongest evidence that "these are laws, not metaphors" is that a system built on them works, and its bugs are missing *steps*, not missing features.

---

## Suggested next actions (for Michael / the book session)
- Weave lessons 1–2 into Part Two ch. 04 (Review) and ch. 03 (spec) + Part One `p1_03`/`p1_04`.
- Weave lesson 4 into ch. 08 / ch. 11 (Atonement/refinement) using the critic revise-loop as the worked example.
- Weave lesson 5 into `p1_04_pack_the_context`.
- Weave lesson 6 into the guide's Step 9 and (carefully, honestly) the book's treatment of Sabbath.
- The full lived case study is at `docs/work-with-ai/examples/2026-06-04-substrate-critic-harness-creation-cycle.md`.
- Consider updating `guide/05_complete-cycle.md` directly with lessons 1, 3, 4, 5 (they refine the development patterns that section already lists).
