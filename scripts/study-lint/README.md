# study-lint — the detector suite

The **detector** half of the study-tooling suite (see
`.spec/proposals/study-tooling.md`). Where the `quoter` makes verbatim+linked the
easy path at write time, the linter *catches* what was written by hand. Together
they close a self-validating loop: write through the quoter → lint confirms green.

Rules are added one at a time, each precision-tuned against the real corpus.

| rule | catches | status |
|------|---------|--------|
| `scripture_verbatim.py` | quoted text next to a scripture link that isn't verbatim in the verse | ✅ built 2026-06-13 |
| `link_validate.py` | broken / mis-pathed / directory links; scripture label↔path mismatch | ✅ built 2026-06-13 |
| (verify-quotes) | Webster 1913-as-1828 — lives at `scripts/verify-quotes/` | ✅ shipped |
| citation-depth-2, counted-claim, … | roadmap | — |

Same posture as verify-quotes: **run it manually; not a pre-commit hook yet** —
promote once it keeps earning weight. Exit 0 clean, 1 on any flag.

## scripture-verbatim

```sh
python scripts/study-lint/scripture_verbatim.py study/foo.md ...
find study -name '*.md' | grep -v .audit | xargs python scripts/study-lint/scripture_verbatim.py
```

Reuses the quoter's spine — `resolver` (ref→file), `sources` (verse text),
`grammar` (the verbatim-spans + marked-edits engine). One engine, both ends.

**Two hard parts, both handled:**

1. **Association** — pair each scripture link with the quote it belongs to. A quote
   immediately precedes its link (inline `"phrase" ([Ref](link))` or in a blockquote
   `> "verse" / > — [Ref](link)`). Events are streamed in document order; a link
   pairs with the nearest preceding quote inside a proximity window. A bare
   reference with no quote, or a quote followed by a non-scripture link, is skipped.

2. **Precision** — the corpus is full of legitimate non-errors near scripture
   links: partial phrases, `[bracket]` edits, `...` ellipses, the study's own
   labels in quotes, condensations, and quotes of a *different* verse. Flagging
   every verify-failure would cry wolf (a naive pass flags ~213 on the
   already-walked corpus). The gate fixes that:

   - `grammar.verify` first — a properly-marked quote (verbatim spans + brackets +
     ellipses) passes and is never flagged. (`pre()` strips markdown emphasis and
     embedded `5.` verse numbers so a clean block quote isn't faulted for them.)
   - On failure, flag **only** a genuine near-miss: either a long single contiguous
     run (`anchor ≥ 0.85` — one edit at a boundary, body intact) **or** two
     genuinely long runs each ≥4 tokens covering ≥90% (one deviation in the
     *middle*, between two verbatim chunks). A 2+2 label, a fragmented condensation,
     and common-word scatter (a quote of another verse) all fail both bars.

   On the walked corpus this lands **~35 flags out of ~2430 quotes (1.4%), all
   genuine** — real unmarked elisions ("which is", "through his subtilty"), version
   differences (Matthew wording under a 3 Nephi link), splices, and one inserted
   `*not*`. Precision over recall, like verify-quotes.

**Known recall gaps (v1, by design):** a quote >320 chars from its link; a quote of
the *wrong* verse entirely (correctly skipped as a different source — that's
link-validate's job); multi-deviation paraphrases (often acceptable shorthand
anyway). The exact-match `grammar.verify`, the human walk, and the quoter
(prevention at write time) cover the rest. The two thresholds are single tunable
lines at the top of the file.

### Accepted-flags ignore list

Flags reviewed and deliberately accepted (an unmarked trim that's fine as prose, a
deferred carry-forward) go in `scripts/study-lint/accepted.tsv` so they stop
surfacing. One per line: `relpath <TAB> ref-label <TAB> first ~40 chars of the
quote`. `#` comments allowed. The carry-forward doc that records *why* lives at
`study/.audit/scripture-verbatim-carryforward.md`.

## link-validate

```sh
find study -name '*.md' | grep -vE '.audit|.scratch' | xargs python scripts/study-lint/link_validate.py
```

Validates every relative `.md` link: `BROKEN` (target missing), `DIR` (points at a
directory, not a file), `MISMATCH` (a scripture link whose label resolves to a
different file than the path — the dc/76-labeled-"D&C 109" class). Reuses the
quoter's `resolver` for the label→file map. Objective — a path is right or wrong —
so findings are safe to just fix. Note: scratch and deeply-nested working files
often carry wrong-*depth* gospel-library links copied from a shallower parent;
exclude `.scratch` for a routine run.
