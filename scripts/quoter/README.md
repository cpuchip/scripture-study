# quoter — the constructor

Insert verbatim source text + a canonical, correctly re-based link into a target
file. The **write-side dual of the linter** (see
`.spec/proposals/study-tooling.md`): the linter *detects* bad quotes after the
fact; the quoter makes verbatim+linked the *easy path*, so the error class never
gets created. The vulnerability it removes is the **retype** — the moment you read
a verse and re-key it from working memory. Pull it through the quoter and there is
no retype; what lands in the document is the source's own bytes.

Built 2026-06-13, feature set grounded in how the studies *actually* quote (a
survey of the corpus, not an a-priori design).

## The constructor refuses

`quote` will **not emit a quote it cannot verify.** A FAIL is a refusal (exit 1),
not a warning. Ask for a phrase that isn't verbatim in the verse, or attribute a
definition to a word Webster 1828 doesn't define, and it stops. That refusal is
the whole point.

## Feature ladder

| | form | example |
|---|------|---------|
| **v1** | block (set-piece) | `> "whole verse" `…`— [Ref](link)` |
| **v2** | inline phrase | `"a verbatim sub-phrase" ([Ref](link))` |
| **v3** | free-flow | marked edits inside the phrase — `[brackets]` + `...` ellipses — verified span-by-span |
| | promote | carry a verified quote scratch → study, re-basing the link + re-verifying against source |

These map onto the real inline forms in the corpus: partial phrases, verse ranges
(`Romans 5:3-5`), bracketed editorial edits (`"I promise [you] that…"`), and
ellipses (`"sweet above all that is sweet… and ye shall feast…"`).

## Usage

```sh
# v1 — block a whole verse/range
quote scripture "Romans 5:3-5" --into study/foo.md

# v2 — a verbatim sub-phrase, inline (the dominant study form)
quote scripture "Alma 5:14" --phrase "received his image in your countenances" --into study/foo.md

# v3 — free-flow: brackets + ellipses, each span verified
quote scripture "Alma 5:14" --phrase "have ye spiritually been born of God ... received his image in [their] countenances" --into study/foo.md

# Webster 1828 — attribution-only by default; --link adds the 1828.ibeco.me link
quote webster countenance --def 4 --into study/foo.md
quote webster countenance --phrase "Favor; good will; kindness" --link --into study/foo.md

# promote — scratch → study, link re-based for the study's depth, re-verified
quote promote "Alma 5:14" --from study/.scratch/notes.md --into study/foo.md
```

`--dry-run` previews without writing (link still re-based for `--into`). `--rel
TARGET` sets the link target when you aren't writing into a file. Run from the
repo root so relative paths resolve correctly.

## The quote grammar (shared with the linter)

A quote = **verbatim spans + marked edits**. Brackets `[..]` (insertion/
substitution) and ellipses `...`/`…` are legitimate; the spans between them must be
verbatim. Any **unmarked** deviation — a dropped word, a swapped word, a
paraphrase-in-quotes — is an error. Punctuation and case differences are invisible
(free quotes re-case and re-punctuate at splices); a missing or changed **word** is
not. This is the exact judgment the 469-file walk made by hand, now in
`grammar.py` — and the same engine the linter's `scripture-verbatim` rule will use
to enforce it. The constructor *produces* well-formed marked quotes; the detector
*enforces* the grammar on hand-written ones. One engine, both ends.

## Layout

| file | role |
|------|------|
| `resolver.py` | ref → canonical gospel-library path + relative link (re-based per target). The shared spine; the linter's `link-validate` will reuse it. |
| `sources.py` | verbatim text from gospel-library (verse/range, `<sup>` stripped) and genuine 1828 definitions. |
| `grammar.py` | the quote grammar — spans + marked edits, `verify(quote, source)`. Shared with the linter. |
| `quote.py` | the CLI: `scripture` / `webster` / `promote`. |

Each module runs standalone (`python resolver.py "Alma 5:14" study/x.md`) for quick
checks.

## Validated

- Resolver: every ref across all five volumes resolves to an existing file; link
  re-basing correct at `../`, `../../`, `../../../` depths.
- Grammar: verbatim / bracket / ellipsis / case-splice pass; unmarked swap and drop
  fail.
- Refusals: non-verbatim phrase, ABSENT Webster word, and a corrupt scratch quote
  on `promote` all exit 1.
- **Self-validating loop:** quoter output run through the `verify-quotes` detector →
  0 flags. Lint-clean by construction.

## Roadmap

- **scripture-verbatim linter rule** consumes `grammar.py` + `resolver.py` (the
  detector half — next build).
- **talks** as a source (v2 of the quoter) — conference/ensign files.
- **MCP wrapper** (`gospel_quote` / `webster_quote`) so the agent calls it at
  write-time during drafting — closes the confabulation surface at the source.
  (A new standing capability — council nod first.)
- **publish-step** Webster-attribution → `1828.ibeco.me` linkifier.
