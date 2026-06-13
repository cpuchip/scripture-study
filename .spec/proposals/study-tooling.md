# Study Tooling — the self-validating study loop

*Design arc captured 2026-06-13 (webster-1828 session), from the
study-correctness walk retrospective. Status: **DESIGN / awaiting build nods.**
Each tool is a new standing capability → council nod before building
(dominion-in-council). verify-quotes (rule #1) already ratified + shipped.*

## The principle (oracle-first)

Long horizontal LLM-intense verification work usually hides a **deterministic
oracle**. Build it first. Decompose into:
- **Detector** — deterministic, scriptable. Perfect recall, zero fatigue, exit
  0/1. The part we keep accidentally doing with an LLM.
- **Adjudicator** — irreducible judgment (read source, decide, fix).

Self-validating loop: **detect → fix → re-detect → green.** The detector is entry
filter (shrinks N to the flagged subset) AND exit gate (objective done) AND the
inverse-hypothesis confirm-step, for free. Memory: `feedback_build_the_oracle_first`.

The walk was a deterministic check in disguise — `verify-quotes`, written *after*
the 469-file marathon, caught 8 contaminations the walk AND the fan-out both
missed, in seconds. We should have built it first.

## Two halves: detector + constructor

- **Linter (detector)** — catches non-verbatim quotes / bad links *after* writing.
- **Quoter (constructor)** — makes verbatim+linked the *easy path*, correct by
  construction, killing the error class at its source (the read-then-retype gap).

Together: write *through* the quoter → linter confirms green → linter only ever
catches the hand-typed exceptions. Best systems have both — a constructor that
makes the right thing easy + a checker that catches deviations.

## Shared spine: the ref → canonical-path resolver

The genuinely fiddly, high-value component both halves depend on:
`"Alma 5:14"` → the correct gospel-library file, with the `../` depth computed
**relative to the target file's directory**. Knows the book-name→folder map
(exodus→ot/ex, micah→ot/micah, D&C→dc-testament/dc, etc.) — the exact map behind
every link error the walk caught (exo→ex, dc/76 vs dc/109). gospel-engine-v2
already carries this map internally; **reuse it, build the resolver once, both
tools consume it.** It's also what makes quote-`promote` re-basing deterministic.

---

## A. The linter — `study-lint` (detector suite)

`study-lint <file>` runs all rules, exits 0 or lists flags. Build rules one at a
time; each its own nod.

| Rule | Catches | Status |
|------|---------|--------|
| **verify-quotes** (Webster dual-edition) | 1913-as-1828 (quote matches 1913 > 1828) | ✅ SHIPPED — `scripts/verify-quotes/`, 8 catches, corpus 540/0 |
| **scripture-verbatim** | quoted verse text ≠ the gospel-library verse | ⭐ **biggest next build** |
| **link-validate** (gospel-aware) | broken/mis-pathed links | partial base: md-mcp `md-link-validate` + the resolver |
| **citation-depth-2** | "Webster/source cites X" when it doesn't (image→Matt 22:20) | source-verification depth-2 rule, automated |
| **counted-claim** | "appears N times / N voices" mis-counts (13→9, three→2) | needs BYU MCP |
| **date-sanity / talk-quote** | impossible dates; mis-dated/misquoted talks | later |

**scripture-verbatim (the next build):** for each gospel-library ref link, diff the
adjacent quoted text against the verse (`**N.** text` format, footnote sups
stripped). The #1 thing the walk did by hand 469×. Two hard parts: (1) associating a
quoted string with the right ref — a **proximity heuristic** (quote near its link /
quote following "verse N") gets most of it, precision-tuned like verify-quotes v1;
(2) **speaking the quote grammar from day one** — partial-phrase quotes and *marked*
edits (`[brackets]`, `...` ellipses) are legitimate and must NOT flag; only *unmarked*
deviations are errors. A naive whole-string compare would false-positive on every
honest free quote. Payoff: turns the next canon walk (PoGP) into a verify-the-flags
pass instead of a read-everything marathon.

---

## B. The quoter — `quote` (constructor)

Insert verbatim source text + canonical link into a target file, correct by
construction. The write-side dual of the linter; the retrieval already exists
(`gospel_get`, `webster_define`) — the new value is **link-gen + verbatim insert +
provenance-preserving promote.**

```
quote scripture "Alma 5:14"   --into scratch.md   # gospel-library → scratch
quote webster   countenance   --into scratch.md   # 1828 → scratch
quote promote   "Alma 5:14"   --from scratch.md --into study.md   # scratch → study
```

Emits: `> "<verbatim text>" — [<ref>](<canonical relative link>)`

- **scripture** — pull verse(s) verbatim from gospel-library (sups stripped),
  link relative to the target. Ranges ("Alma 5:14-16") in v1; talks v2.
- **webster** — genuine 1828 def verbatim + attribution. **DECIDED (Michael
  2026-06-13):** attribution-only `— Webster 1828` in the *source* text (clean,
  reads well), and **the publish step linkifies** the attribution into a live link
  straight to the word — `1828.ibeco.me/word/<w>` (ours, preferred; dogfoods
  1828-illuminated) or `webstersdictionary1828.com` as the alt. Explicit inline
  link available via a `--link` flag when wanted. Separation of concerns: working
  text stays clean, the published face gets rich links. (New publish-step
  enhancement: a Webster-attribution → link transformer; same shape as the
  scripture-link pass-through.)
- **promote** — NOT a retype. The scratch block is already verbatim-by-construction,
  so promote = locate the block + **re-base the relative link** for the study's
  depth (the "original ref link pass-through" Michael named). The only thing that
  breaks moving a quote between files is the `../` count; the tool owns both dirs,
  so it re-bases deterministically. Provenance preserved.

**Interface:** CLI first (fast, testable, fits the lint family), shared resolver
core, then an **MCP wrapper** (`gospel_quote` / `webster_quote`) so the agent calls
it *at write-time during drafting* instead of read-then-retype — closing the
confabulation surface at the source.

### Block vs free-flow — the quote grammar (Michael 2026-06-13)

v1's blockquote (`> "whole verse" — [ref]`) is verbatim-by-construction the *easy*
way, because nothing is touched. But studies often **free quote** — weaving the
words into the sentence, which reads better than a block: *Alma asks whether they
have "received his image in [their] countenances."* Block is right for set-piece
passages; inline free-flow is right for argument prose. The quoter must support
both (the inline forms are **v2/v3**).

Free-flow is harder because the quote gets legitimately *shaped*: **sub-selected**
(a phrase, not the verse), **bracket-edited** for grammar (`your` → `[their]`),
**elided** (`...`), **case-spliced** at a sentence start. These are honest
scholarly moves — but they're also exactly where contamination hides (an *unmarked*
word-drop, a silent paraphrase-in-quotes). The walk made this judgment by hand 469×
("unmarked elision → mark it"; "Alma 22:18 dropped an 'and' twice → flag").

The resolution is a **shared quote grammar** the quoter and linter both speak:
> a quote = verbatim spans + **marked** edits. Brackets `[...]` (insertion/
> substitution) and ellipses `...` are legitimate; verify the retained spans.
> Any **unmarked** deviation (dropped/swapped word, paraphrase) = error.

- **Quoter** *produces* well-formed marked quotes → correct + lint-clean by
  construction. `quote scripture "Alma 5:14" --phrase "received his image in your
  countenances"` verifies the phrase is verbatim before inlining; `--bracket
  your=their` emits `[their]` *after* confirming the original word was "your".
- **Linter** (scripture-verbatim) *enforces* the same grammar → catches hand-written
  free quotes that broke it (the unmarked drop, the silent swap). So the rule must
  tokenize into spans + marked-edits, not do a naive whole-string compare.

Maturity ladder, both tools growing the grammar in parallel:
- **v1** — whole-block quote (trivial grammar: one verbatim span).
- **v2** — inline *phrase* (sub-selection + verify) with `--inline` output.
- **v3** — free-flow into prose: bracket/ellipsis/case edits, the tool verifying the
  **woven result** still carries the verbatim span + correct link before commit.
  `promote --inline` re-casts a scratch block as woven prose, link re-based.

By v3, write-through-quoter and lint-check are the *same grammar from both ends* —
the constructor teaches it, the detector enforces it, and the exact judgment the
walk did by hand is now mechanical.

---

## Build order (recommended)

1. **ref→path resolver** (shared spine) — extract/reuse from gospel-engine-v2.
2. **scripture-verbatim** linter rule (consumes resolver) — highest-value detector.
3. **quote scripture + webster** CLI (consumes resolver) — the constructor.
4. **quote promote** (re-basing) — v1.1.
5. **link-validate, citation-depth-2** linter rules — fill the suite.
6. **MCP wrappers** — once CLIs prove out, put the loop in the agent's hands.

Counsel: build the resolver first; it's the dependency under everything and the
single fiddliest piece. Validate each tool against the corpus the way verify-quotes
just proved out (run it, see what it catches, confirm catches against source).
