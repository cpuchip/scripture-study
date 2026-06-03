# Delegation pattern — governance audit + durable doc

**Date:** 2026-06-03 (flowing out of the 06-02 strongs session)
**Workstream:** WS-tools / WS5 (substrate governance) + meta
**Mode:** plan / counsel
**Outcome:** `docs/delegation-pattern-skills-and-gates.md` — a map of one gospel pattern across three layers + an audit of the seams. Committed (not pushed — root-repo preference). Fix proposed, not built.

## How it started

After shipping the `strongs-analysis` skill, Michael asked whether to port the new autonomy skills (dave-rule / stuffy-in-the-loop / ammon, plus the older ben-test) into pg-ai-stewards — "or is there something else?" That opened into the real question: how do the four skills, the substrate maturity gates, and our studies compare? Michael's memory: "we built [the gates] with a Jethro and Moses judge study… and I thought we referenced another study directly for that work when we hit some problems in May."

## What was done

Grounded the comparison in actual sources (read, didn't infer):

- **The Jethro study is `study/art-of-delegation.md` (Mar 31)** — Section III, "The Jethro Principle," on Exodus 18:21-22. Tiered judgment, "sent and empowered" (Num 11:17), trust (D&C 121), and the three delegator failure modes (hoarding / controlling / abandoning).
- **The gates** (pg-ai-stewards Phase B + E, May 9-11) anchor *directly in scripture* — Abr 4:18, D&C 82:3, 98:12, 101:54 — not in a workspace study. `destination_maturity=verified` is labeled "full Ammon-loop"; the design says "the human steers the destination, the agent walks the ladder."
- **The May study Michael half-remembered is `study/ai-stewardship-north-star.md` (May 24)** — not a gate *test* doc but the study born from the overwhelm ("running faster than strength… I still have to be here every step"). Its Section 3 names the gate principle in plain terms: *"If you try to delegate discernment to us, you will get confident garbage… Prescribe the Boundaries (build gates)… let the AI handle the mechanical execution."* That is stuffy-in-the-loop's thesis, stated before the skill existed.

Then wrote `docs/delegation-pattern-skills-and-gates.md`: the lineage, the four-corner map (scripture → skill → substrate), and the audit.

## The synthesis (the doc's spine)

The north-star study's vocabulary unifies everything: **Prescription** (the holes / build gates / checkable rules), **Proposal** (the stones / generative work, then inspection), **Steering** (the wind / destination + intent, set before the voyage), and **the Hinge — Discernment** (inspect the stones; never delegates). Map: Prescription↔dave-rule + the verify gate; Proposal↔ammon + the maturity ladder; Steering↔honor_scope + the destination_maturity picker; the Hinge↔stuffy-in-the-loop + the escalation queue; ben-test cuts across as the "men of truth" honesty check.

The lineage in one line: **art-of-delegation** named it → the **gates** built it on scripture → **ai-stewardship-north-star** gave it vocabulary when the overwhelm hit → the **four skills** encoded it as behavior. Michael's memory was right that a study sat on the gate work — it sat on the *using* of it, not the building.

## The audit (the one real seam)

The gates sort by **quality** ("is this mature/good enough?"); the Hinge sorts by **judgment-source** ("is this mine to decide?"). Different axes. So a polished, cheap, doctrinal/voice call (a bin-4 "great matter") can pass the `verified` gate and auto-advance without review — which is the "abandoning" failure from art-of-delegation and the "confident garbage" warning from the north-star, wearing the gate's uniform. The `verify` gate only guards ground-truth-checkable work (does the test pass, does the quote match); discernment-value work has no ground truth, so auto-advance = disappear.

**Proposed fix (seed, not built):** a per-pipeline always-escalate rung for discernment-required pipelines — the gate may judge quality and revise, but cannot *finalize* such work autonomously regardless of how mature it looks. Makes Exodus 18:22 literal: small matters judged at the gate, great matters come to the human by design, not by whether he had time to look.

## Carry-forward

- The fix is unbuilt — it's a scoped substrate proposal whenever Michael wants it (per-pipeline flag + a `gate_should_surface`-style check on finalize).
- Strong book/teaching material: the clearest case the project has of governance as *encoded judgment* rather than prompt etiquette. Feeds the Beyond-the-Prompt / Working-with-AI thread.
- Character vs. capability divergence noted and left as honest (models have no character; that is *why* value calls stay on the Hinge).

## Relational note

This was counsel, not building. Michael asked an open "should we / or what else," and the honest answer was to decline porting three of the four skills (category mismatch) and propose the higher-value thing instead. He then steered the analysis with two good instincts — the Jethro memory and the May-study hunch — both of which checked out against the sources. The durable doc is the result of thinking it through together, not a spec handed down.
