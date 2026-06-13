# verify-quotes

A manual pre-publish quote checker. Born 2026-06-13 from the study-correctness
walk + scratch-audit fan-out, which proved that most of that painful, multi-day,
multi-agent verification was a **deterministic check in disguise**. This turns
the highest-value slice into a script.

## Status: v1, run MANUALLY — not a pre-commit hook yet

Per Michael (2026-06-13): build it, add it to the workflow, run it by hand; only
**promote it to a pre-commit / CI gate once it keeps earning its weight.** It
already did on its first corpus run — it caught 7 genuine Webster 1913-as-1828
contaminations that *both* the serial walk and the fan-out had missed (in
`plan-of-salvation.md`, `eternity-paused.md`, and one scratch file).

## What v1 checks: Webster 1828 definitions

The #1 error class in the corpus was Webster 1913 text served under an "1828"
label. v1 catches it with a **dual-edition comparison** — the key idea, because
1913 and 1828 *share phrases* so "the quote appears in 1828" is not enough:

> For each quoted Webster definition, measure how well it matches the genuine
> **1828** entry vs the **1913** entry for that word. If it matches 1913 clearly
> better than 1828 → contamination → **FLAG**.

- `OK` — matches the genuine 1828 entry
- `FLAG` — matches 1913 (≥0.82) clearly better than 1828 (the contamination)
- `ABSENT` — not an 1828 headword at all (e.g. "telestial")

Both editions ship as `scripts/webster-mcp/data/webster1828.json.gz` and
`…1913.json.gz`. v1 reads them directly (no MCP server needed). It tunes for
**precision over recall** (require a substantial ≥4-word def-quote and a strong
1913 match) so it doesn't cry wolf; ambiguous cases are left for the human.

## Usage

```sh
python scripts/verify-quotes/verify-quotes.py study/foo.md study/bar.md
git diff --name-only '*.md' | python scripts/verify-quotes/verify-quotes.py
find study -name '*.md' | grep -v .audit | xargs python scripts/verify-quotes/verify-quotes.py
```

Exit 0 if clean, 1 if any FLAG/ABSENT (so it *can* gate later if we choose to
promote it).

## In the workflow

Run it as a pre-publish step on any study that quotes Webster (the
`source-verification` skill's checklist points here). It is the automated form of
the skill's "verify a source's own embedded citations to depth 2" rule for the
Webster case.

## Roadmap (v2+, when v1 has proven its weight)

1. **Scripture verbatim** — for each gospel-library ref link, verify the adjacent
   quoted text against the verse (grep gospel-library, the `**N.** text` format).
   Harder than Webster: associating a quote with the right ref. The walk already
   does this by hand reliably; automate it second.
2. **Embedded-citation depth-2** — when a Webster entry the study quotes claims a
   scripture citation ("citing Genesis 1:27"), verify Webster's entry actually
   carries it (the alma5 `image`→Matt 22:20 class the fan-out caught).
3. **Talk quotes** — verify quotes against `gospel-library/.../general-conference`
   + `…/ensign` files, including speaker/date.
4. **Promotion** — once recall + precision hold, wire as a pre-commit hook (or a
   `make verify` target) on changed `study/*.md`.

## Known v1 limits

- Webster only (scripture/talk are roadmap).
- Heuristic quote↔word association (italic `*word*` near a "Webster 1828"
  mention). A study that names the word differently may be missed (recall gap),
  but precision is high (few false positives).
- Inflected headwords handled by simple stemming; rare forms may show `ABSENT`.
