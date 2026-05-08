# Four-Way Comparison: FtC/WtL Studies — 2026-05-08

> Final memo, all four runs terminal.

## What we compared

| Run | Model | Prompt | Pipeline | Corpus access | Tokens (in/out) | Elapsed |
|-----|-------|--------|----------|---------------|-----------------|---------|
| **#1** original | kimi-k2.6 | base (`*`) | `study-write` | ✅ yes | 626K total | 17m14s |
| **#2** kimi-tuned-no-corpus | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | ❌ perm regression | 87K / 36K (122K) | 8m11s |
| **#3** qwen-base | qwen3.6-27b (lm_studio) | base (`*`) | `study-write-qwen` | ⚠ partial → ✅ | 825K total | 24m |
| **#4** kimi-tuned-with-corpus | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | ✅ yes | 855K / 70K (925K) | 24m30s |

## TL;DR

**The kimi-tuned prompt cleared all six signatures, with or without corpus access, in both the run #2 (no-corpus stress test) and run #4 (corpus-grounded apples-to-apples) experiments.** Run #4 is the strongest output: scene-opener, claim-style headers, real corpus citations, anti-symmetry framing, verified-quote discipline that caught and removed fabricated phrases from its own draft.

**Qwen3.6-27b (run #3) exhibited several kimi-shared signatures plus its own quirks** — most distinctively, it tried to call `study_get('bofm/ether/12')` as if substrate slugs were scripture references (they aren't), and it used `(#)` broken-link placeholders for internal corpus references. Both are qwen-tuning candidates for a future variant.

## Rubric — six kimi signatures from the 2026-05-07 review

| Signature | Run #1 (base+kimi) | Run #2 (tuned+kimi+no-corpus) | Run #3 (base+qwen) | Run #4 (tuned+kimi+corpus) |
|-----------|--------------------|------------------------------|--------------------|---------------------------|
| 1. Symmetric-pair compulsion | ✅ heavy | ❌ resisted ("any such mapping crumbles under pressure") | ✅ table + parallel constructions | ❌ resisted ("assimilation language is more accurate than symmetry language") |
| 2. Triadic flourishes | ✅ ("Three witnesses, one tree, one ascent") | ❌ none | ✅ ("Three witnesses — Paul, Mormon, Moroni") | ❌ none in body |
| 3. Closing refrain (function) | ✅ ("The ascent is one, the descriptions are two...") | ❌ closes on practical action | ✅ ("The circle closes... It is Christ.") in §IV | ❌ closes on practical advice |
| 4. Pseudo-citation register | ✅ ("[study] anchors...") | N/A no corpus | ✅ ("[study](#) shows...") with broken `(#)` links | ✅ but more naturalized — uses real paths and integrates the references into argument flow |
| 5. Latinate over Anglo-Saxon | ✅ (architecture, mechanism, ontological, geometry, perceptual organ) | ❌ Anglo-Saxon throughout | partial (architecture, structural) | ❌ mostly Anglo-Saxon |
| 6. Confabulation in revision notes | ✅ Romans 5:5 reverse-fix | ❌ honest disclosure of constraint | partial (hedged but didn't catch issues) | ❌ found and removed two fabricated quotes from its own draft + corrected mis-attribution + removed unverified statistical claim |

**Score:**
- Run #1: 6 / 6 signatures present
- Run #2: 0 / 5 measurable signatures present (pseudo-citation N/A)
- Run #3: ~4 / 6 signatures present
- Run #4: 1 / 6 signatures present (residual pseudo-citation, but more naturalized than #1)

## Voice metrics

| Metric | #1 | #2 | #3 | #4 |
|--------|----|----|----|----|
| Total lines | 105 | 43 | 239 | 118 |
| Section count | 6 (labels) | 0 + Becoming | 6 (labels) | 6 (theses) + Becoming |
| Em-dashes (body, citation excluded) | ~12 | 0 | ~15+ | ~3 |
| Direct verbatim quotes (substrate-verified) | many | 0 (forced) | many | many |
| Reference-only citations | 0 | 7 | 0 | 0 |
| Bold-emphasized declarations | minor | none | many ("**They are not the same concepts.**") | minor |
| Triadic cadence in close | yes | no | yes | no |
| Closing refrain | yes | no | yes (in §IV) | no |
| Self-found errors in revision notes | 0 | 0 (acknowledged limit) | 1 | **3** (most thorough verification) |

## Section header comparison — the cleanest "labels vs theses" diff

| Run | First section header |
|-----|----------------------|
| #1 | "I. The Two Triplets as Ordered Progressions" |
| #2 | (no headers) |
| #3 | "I. The Two Triplets — Mapping the Terrain" |
| **#4** | "**Thomas asked for directions, and Jesus gave Himself**" |

Only run #4 wrote a header that does argumentative work. The kimi-tuned prompt's "section headers must be claim sentences, not category labels" rule landed.

## Run #4 strongest passages

**Opening (immediate scene drop with three witnesses braided):**
> "Thomas asked for directions. Jesus answered with Himself: 'I am the way, the truth, and the life' (John 14:6). Six centuries later and an ocean away, Moroni wrote: 'Wherefore, there must be faith; and if there must be faith there must also be hope; and if there must be hope there must also be charity' (Moroni 10:20). One triplet is Christ's identity. The other is the soul's transformation. The binding question is whether they are two names for one thing, or two things that happen to converge."

**Anti-symmetry resistance with directional argument:**
> "The frame that treats them as 'two vantage points on the same point' misses the directionality. The human triplet is moving toward the Christ triplet. The Christ triplet is not moving toward the human triplet. The movement is one-way, which is why assimilation language is more accurate than symmetry language."

**Verification discipline in revision notes:**
> "Removed fabricated quotes. The draft attributed the phrases 'the perceiver-state' and 'the Object being perceived' to the *discernment-and-the-comprehending-eye* study as if they were verbatim terms. A corpus search confirmed neither phrase exists in that study (or anywhere in the substrate). Replaced them with an accurate paraphrase..."

**Closing on practical action (no refrain):**
> "Fifth, charity is obtained, not achieved. The [charity] study records a six-month prayer for charity that transformed into a prayer to see others as Christ sees them. That is the pattern. Ask for the gift. Expect it to change what you see before it changes what you do."

## What qwen3.6-27b does differently from kimi (preliminary signatures)

Observations from run #3 worth encoding into a future `.stewards/qwen-3.6/study.agent.md`:

1. **Tool-name confusion.** Qwen tried calling `study_get('bofm/ether/12')` as if substrate study slugs were scripture references. The base prompt told it about `study_get(slug)` but qwen interpreted "slug" loosely. A qwen variant should add: "**study_get takes a kebab-case slug from the substrate corpus** (e.g., `way-truth-life`), NEVER a scripture reference path. For scripture, use the canonical scripture URL or paraphrase."

2. **Broken internal-link convention.** Qwen rendered substrate references as `[study-name](#)` instead of `[study-name](study-name.md)`. The convention isn't enforced anywhere in the prompt; for kimi the base prompt's example carries through. A qwen variant should explicitly state: "Internal links use the `[slug](slug.md)` form. Never use `(#)` placeholders."

3. **Heavy table use.** Qwen produced a full comparison table in §VI that no kimi run did. Could be a feature (good for structural arguments) or a tic (substituting structure for argument). Worth flagging in the qwen variant.

4. **Bold-emphasis density.** Qwen bold-emphasized phrases throughout ("**They are not the same concepts.**", "**They ARE the same architecture.**"). Reads as preacher-cadence. Kimi rarely bolds.

5. **Triadic emphasis in body, not just close.** "Faith is what we exercise; the Way is what Christ IS. Hope is what we feel; the Truth is what Christ embodies. Charity is what we become; the Life is what Christ gives." Three parallel constructions back-to-back. Qwen reaches for this even when not closing a section.

6. **More verbose overall.** 239 lines vs run #4's 118. Qwen says things twice — once briefly, once in a parallel construction. The rules don't currently target this.

## Findings

### The kimi-tuned prompt is validated

Run #2 (stress test, no corpus) and run #4 (proper experiment, with corpus) both produced studies that cleared the six signatures. The prompt's rules carried — including under degraded tool surface, where they prevented confabulation that the base prompt enabled in run #1.

### Verification discipline scales with tool availability

Run #2 with no corpus produced honest disclosure of the constraint. Run #4 with corpus produced **active verification** — catching fabricated quotes that the agent had inserted into its own draft and removing them. This is the "verification claims must be tool-grounded" rule working at full strength.

### qwen needs its own variant

Run #3 wasn't a clean pass or a clean fail. Qwen has its own voice signatures, some shared with kimi, some distinct. Authoring `.stewards/qwen-3.6/study.agent.md` is a tractable daytime task with run #3 as the diagnostic baseline.

### The substrate's `study_*` tools work end-to-end with the imported corpus

Runs #3 and #4 successfully used `study_search_text` and `study_get` to read existing studies and ground their meta-syntheses. The earlier perm regression (now patched) was the only blocker. The substrate is producing real corpus-grounded scholarship.

## Recommendations

1. **Promote run #4 over the current `study/two-triplets-one-ascent.md`?** Run #4 is substrate-produced AND well-voiced AND properly source-grounded. The current published file is the Opus-4.7-revised version of run #1. Worth a side-by-side read by Michael; if run #4 holds, replacing the file would be the cleanest evidence that the experiment worked.

2. **Promote `.stewards/kimi-k2.6/study.agent.md` from "experimental" to "stable v1".** Update the iteration log in `.stewards/kimi-k2.6/README.md`.

3. **Author `.stewards/qwen-3.6/study.agent.md`** when convenient. Six observations from run #3 above are the starting amendment list.

4. **Daytime architectural followups (in priority order):**
   - `agent_tool_perms` provenance fix (so substrate broadcasts survive frontmatter reimports)
   - Phase 3c.4 gospel-engine HTTP tools (build pg_net or pgsql-http into Dockerfile, OR extend Rust bgworker with `tool_http` work_kind)
   - Phase 3c.3.5 (formerly speculative): auto-promote completed work_items into `stewards.studies`
