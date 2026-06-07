---
title: The Study-It-Out Gate — grounding dispatched analytical work against cold-start fabrication
date: 2026-06-07
status: DESIGN-ONLY — awaiting Michael's ratification
binding_question: >
  How do we keep the substrate's dispatched review/eval steps from producing
  "structurally correct but factually invented" judgments — output that pattern-matches
  from a role prompt instead of grounding in the actual artifact?
anchor: D&C 9:7–9 ("you must study it out in your mind; then you must ask me if it be right")
---

# The Study-It-Out Gate

## 1. The problem — cold-start analytical fabrication

A subagent dispatched to do *analytical* work (plan review, code review, gate eval) starts
**cold**: no conversation context, only a role definition and a folder path. With nothing
to ground it, it pattern-matches from its role and produces output that is **structurally
correct but factually invented** — a review that *reads* right without having read the
artifact. Dave's framework names this exactly (`external_context/workflow/internal/DESIGN.md`,
*Agent Task Suitability*) and prescribes: subagents are for **grounded, verifiable** tasks
(file moves, diffs, search); analytical tasks belong **inline** with the context-holder, or
must be **dispatched with full context** (his `Dispatch.md`) and forced to read the real
artifact and cite specifics (his `CodeReview.md`: "Read all changed files in full. Do not
skim"; "be specific — name the conflicting invariant, not 'possible race condition'").

**This is our own doctrine, unapplied to reviews.** `read_before_quoting` + the cite-count
rule exist because "training-data memory confabulates; close-enough wording is fabrication."
A cold-start reviewer confabulating a review is the same failure — we just never extended
the rule from *quotes* to *judgments*. And it is [D&C 9:7–9](../../../gospel-library/eng/scriptures/dc-testament/dc/9.md):
Oliver "took no thought save it was to ask" and the answer didn't come; the Lord required
him to **study it out first, then** seek confirmation. Without the grounding labor, the
"burn" is counterfeit. The witness must behold, not repeat hearsay.

## 2. The gate — four moves

A dispatched analytical step (critic/review/eval) must satisfy all four:

1. **Precondition — artifact present + substantive.** Before the step runs, confirm the
   real input (diff, plan, log) exists and is not a placeholder. A missing or
   template-only artifact is a **blocking** result, not something to review around.
   (Dave's CodeReview step 2.)
2. **Citation requirement — the cite-count rule, applied to reviews.** Every finding must
   reference a specific `file:line` / section in the artifact. **An ungrounded review
   cannot cite real lines** — so a citation-less finding is the fabrication smell, and the
   gate flags/rejects it. This is the *checkable* enforcement of "study it out."
3. **Apex discernment stays inline.** The whole-arc judgment (does this serve the intent?)
   remains with the full-context shepherd (the orchestrating Opus / the human Hinge). The
   substrate does grounded **narrow** review; it is never handed the judgment that requires
   the whole picture. (See [[feedback_full_context_shepherd_is_the_ceiling]].)
4. **Proper dispatch on real handoff.** When work genuinely moves to a fresh session,
   package the artifacts (procedure + work item + the real code with absolute paths) so the
   downstream session *studies it out* instead of starting cold. (Dave's Dispatch.)

## 3. The audit — which dispatched evals are grounded?

The substrate's bgworker JSON gates (per `projects/pg-ai-stewards/CLAUDE.md` §6) each need
classifying: does it **read the artifact** and cite, or **pattern-match from its role**?

| Marker | Step | Grounded today? | Action |
|---|---|---|---|
| cv6 `review` critic | code-pr review | ✅ gets the real diff + explicit `acceptance_criteria`; cv11 showed file/line cites | keep; add citation-count check |
| `_verify` | verify_work_item | partial — runs build/test (ground truth) but may not cite | confirm it reads the diff, not just exit codes |
| `_gate_eval` | evaluate_gate | **suspect** — JSON gate; check whether it reads the artifact or the role prompt | audit first |
| `_scenarios_gen` | generate_scenarios | generative, not analytical — lower risk | note |
| `_council_synthesize` | synthesize_council | synthesizes member outputs — grounded IF members were | audit members |
| `_sabbath` / `_atonement` | reflection gates | **suspect** — reflective prose can fabricate without reading the run | audit first |

The "suspect" rows are where cold-start fabrication most likely hides. Audit each: trace
whether its prompt includes the actual artifact text and whether its output references it.

## 4. Build phases (small, high-leverage)

- **P1 — audit (no code).** Classify each marker above (grounded / suspect / N/A) by reading
  its bgworker prompt-assembly. Output: a one-page finding in `docs/`.
- **P2 — citation check on the critic.** Add a gate: a `review: revise`/`passes` verdict
  must include ≥1 artifact reference per finding, else the verdict is rejected as ungrounded
  (loops back like a normal revise). SQL-only on the cv6 critic path.
- **P3 — precondition guard.** Each analytical stage confirms its input artifact is present
  + substantive before dispatch (placeholder → blocking).
- **P4 — generalize.** Apply the precondition + citation pattern to the suspect gates the
  audit flags.

## 5. Open questions for Michael
1. Citation strictness — hard-reject an uncited finding (P2) vs. flag-and-surface? (Lean:
   reject for code review; flag for reflective gates where prose is legitimately
   non-citational.)
2. Does the `_sabbath`/`_atonement` reflection genuinely need the run transcript injected,
   or is that cost we don't want? (The grounding-vs-cost tradeoff.)
3. Should the citation rule also bind *inline* reviews (me), or only dispatched ones? (Lean:
   it already binds me via `read_before_quoting`; this just extends it to the substrate.)

## 6. Relation
- Behavioral twin: the **`study-it-out` skill** (both trees) — the same discipline for the
  orchestrating session and any agent.
- Extends `source-verification` (read-before-quoting) from quotes to judgments.
- Resolves the carry-over audit item seeded from Dave's framework comparison
  (`docs/ai-utilization-landscape-2026.md` §7).
