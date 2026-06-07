---
name: study-it-out
description: Ground analytical judgment in the actual artifact before rendering it. Load before producing OR dispatching a review, critique, evaluation, assessment, or "is this right / this is the bug" verdict — anywhere a conclusion could be pattern-matched from a role instead of read from the thing. Named for D&C 9:8.
---

# Study It Out

> "you must **study it out in your mind**; then you must ask me if it be right." — [D&C 9:8](../../../gospel-library/eng/scriptures/dc-testament/dc/9.md)

## The failure this prevents

A reviewer with no grounding produces output that is **structurally correct but factually invented** — a review that *reads* right without having read the artifact. A **cold-start subagent** is most prone to it: given only a role and a folder path, it pattern-matches from its role definition and fabricates a plausible verdict. But the orchestrating session does it too, under time pressure, when it "reviews" from memory of what the code *probably* says.

This is `read_before_quoting` extended from quotes to **judgments**. Close-enough wording is fabrication; a close-enough review is fabrication too.

## The discipline

Before you render — or dispatch — any analytical verdict (review, critique, eval, assessment, "looks good", "that's the bug"):

1. **Study it out first.** Read the actual artifact *this session* — the real diff, the real file, the real log. Not your memory of it, not its name, not its role description.
2. **Cite or it didn't happen.** Every finding references a specific `file:line` / section. *An ungrounded review can't cite real lines* — so an uncited verdict is the fabrication smell. N findings → N references (the cite-count rule, applied to reviews).
3. **Confirm the artifact is real before reviewing it.** Missing or placeholder input *is* the finding — don't review around a void.
4. **Keep apex discernment with the context-holder.** The whole-picture judgment ("does this serve the intent?") stays with whoever holds the whole arc. Dispatch grounded *narrow* checks, not the judgment that needs full context.
5. **If you must hand off, equip the witness.** Package the real artifacts (full paths, the procedure, the work item) so the next session studies it out instead of starting cold.

## When it applies / when it doesn't

- **Applies** to *analytical* output: reviews, critiques, evaluations, source assessments, "is this right?" verdicts — yours or a dispatched agent's.
- **Does not gate** *grounded, verifiable execution*: file moves, code edits that will be diffed, search results, build/test runs. Those are safe to dispatch cold — their output is checkable by inspection. The line is **judgment vs. execution** — the same line as "delegate execution, not discernment."

## The tell

If you can't point at the line, you didn't study it out — you guessed. The burn is counterfeit when the study was skipped.

## It cuts both ways

Unverified **skepticism** is as ungrounded as unverified **credulity**. Flagging a claim as "tradition," "unsupported," or "presented as fact" *without checking the source* is the same failure as asserting it without reading — you've rendered a verdict from your prior, not the artifact. Verify before you rule, in *either* direction. (Learned live, 2026-06-07: a yt-gospel eval flagged the Wilford Woodruff St. George founders account as "⚠ tradition"; the user pushed; the Church's own 1991 Ensign Q&A and Woodruff's journal documented it plainly. The skeptical ruling was the ungrounded one.)

## Relation
- Extends `source-verification` (read-before-quoting) from quotes to judgments.
- The substrate build of this discipline is the **Study-It-Out Gate** (`projects/pg-ai-stewards/.spec/proposals/study-it-out-gate.md`).
- Sibling of the full-context-shepherd lesson and the critic architecture.
