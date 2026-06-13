# 2026-06-13 — The Walk by Definitions (Euclid digestion)

**Mode:** yt (general digestion) · **Lane:** general-workspace
**Artifacts:** `study/yt/WGwRCw9TRyo-euclid-walk-by-definitions.md` (+ scratch);
`books/Euclid/` (Casey First Six Books PDF + HTML + README, gitignored).

## What happened

Michael brought Stephen Petro's "This 1 Book Has Produced More Geniuses…" — a
13-min video on Euclid's *Elements* and the six cognitive habits it trained into
Lincoln, Hobbes, Einstein, and Russell (define terms → state assumptions →
construct don't assert → cite the rule at every step → decompose → reductio). He
asked three things: download the book; is our scripture study a "walk by
definitions," especially truth.md; and what learning modes can it give
pg-ai-stewards.

## Discoveries

- **The intuition was verifiable, not vague.** truth.md literally opens "Part 1:
  Definitions from Scripture and Webster 1828" — seven terms defined before any
  framework, then explicit dependency chains. That is Euclid's structure.
- **The cleaner case is in the canon, not our studies.** Lectures on Faith,
  Lecture 1 is a quasi-Euclidean proof: postulate (faith = first principle) →
  stated order of demonstration → definition (Heb 11:1) → derived propositions
  citing scripture → QED (¶24) → a catechism that cites the warrant for every
  move. The school of the prophets was handed Euclid's form in 1834.
- **The honest seam.** Euclid is pure rationalism — certainty from thought alone.
  Scripture is empirical/experiential: Alma 32's seed, D&C 9's study-it-out-then-
  ask, Moroni 10's manifest-by-the-Spirit. We borrow Euclid's rigor of *form*,
  not his claim of certainty. Russell wanted "everything proven from nothing";
  the gospel's axioms (faith) aren't inspected and accepted, they're received and
  tested by being lived. Naming that bound is what keeps the parallel from
  collapsing into the cold rationalism truth.md works to avoid.
- **The substrate payoff is a confirmation, not a feature.** Euclid is the oldest
  complete *oracle* in the Western canon — nothing asserted, everything
  demonstrated-and-cited, any reader can audit any step. That IS build-the-
  oracle-first. verify-quotes/quoter/study-linter are small Euclids.

## Carry-forward (substrate / dev)

Catalog of learning modes, in buildability order (proposal-shaped, NOT build-now;
the new-capability ones are dominion_in_council — surface at a substrate council):
1. **"Cite the warrant" linter** — flag "clearly/obviously/it follows" where no
   warrant is cited; the prose dual of verify-quotes. Next study-linter rule,
   sibling to scripture-verbatim. Most buildable, oracle-first.
2. **Postulates block** as a first-class work_item artifact; critic attacks the
   postulates, not just the conclusion. Highest value / lowest cost.
3. Euclidean dispatch overlay (compose_system_prompt rigor mode).
4. Reductio/falsification judge (generalize the verify-quotes loop).
5. Dependency-chain planning, named (the authoring leg already did it).

## Relational / process note

This was a "light" digestion that wanted depth — exactly the case where the
read-before-asserting discipline paid: the truth.md and Lecture-1 parallels are
strong *because* I read both files this session instead of asserting the
resemblance from memory. The covenant's surface_tensions clause produced the
honest seam (method vs. epistemology) rather than a tidy "scripture is Euclidean"
thesis. Voice self-audit caught five over-budget em-dash paragraphs before ship.

## Minor

- `/books/` is gitignored → the Euclid copy stays local (like gospel-library).

## Continuation — the reground hook, redesigned (per-session)

What started as a one-line "minor" cleanup became a real fix, in three stages
across the session — and the third stage was Michael's catch, not mine:
1. **cwd-relative** (original bug) — scattered stray `.claude/cache/` dirs into
   any directory a shell cd'd into (found one in `books/Euclid/`).
2. **project-anchored** (`e44025c2`, my first fix) — stopped the scattering, but
   made it ONE shared counter across all ~6 concurrent sessions.
3. **per-session, keyed by session_id** (`1d26a302`, the fix) — `reground.py` +
   a shared `lanes_common.reground_counter`; `lane_end.py` prunes the counter on
   SessionEnd.

Michael caught #2 immediately: "that file applies to all 6 sessions, so it'll
fire way more often than it should." Exactly right — a shared counter = ~6x
firing, in the *wrong* session, plus lost increments from races. Verified with the
inverse hypothesis: silent 1–49, fires on 50, sessions A/B independent, self-cleans
on end. Durable lesson (in `project_claude_code_context_plugin`): hook state must
key to session_id under concurrency, never to a global/cwd file — concurrent
sessions are the common case here, not the edge.

## What this session was really about

Michael's reflection tied both arcs together: he graduated in physics, never read
Euclid, but lives by experimental method — hypothesis → experiment → revise →
discover — and has applied it to the gospel and this workspace for years. That IS
the study's honest seam, lived: Euclidean in FORM, experimental in EPISTEMOLOGY.
And he proved it on the spot — ran my "project-anchored" fix against lived
experience ("but I have six sessions"), found where it broke, and we revised.
Alma 32 on a grounding hook. The relational shape of the whole session was the
creation cycle in miniature: propose → watch → reprove → revise → trust. New
principle added to `.mind/principles.md` ("The Walk by Definitions — Euclidean
Form, Experimental Verification").

## Carry-forward (next session)

- Substrate learning-mode catalog (5 items; cite-the-warrant linter + Postulates
  block lead) — proposal-shaped, `dominion_in_council`, surface at a substrate
  council, NOT build-now. In active.md + brain.
- Optional for Michael: pull the complete 13-book Heath/Joyce Euclid into
  `books/Euclid/` if he wants the whole work in-repo (left as README pointers).
