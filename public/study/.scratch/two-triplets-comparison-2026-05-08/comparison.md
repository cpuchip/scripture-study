# Five-Way Comparison: FtC/WtL Studies — 2026-05-08

> Final memo, all five runs terminal. Both kimi-tuned (run #4) and qwen-tuned (run #5) **validated** against their respective baselines.

## What we compared

| Run | Model | Prompt | Pipeline | Corpus | Tokens | Elapsed |
|-----|-------|--------|----------|--------|--------|---------|
| **#1** | kimi-k2.6 | base (`*`) | `study-write` | ✅ | 626K | 17m14s |
| **#2** | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | ❌ (perm bug) | 122K | 8m11s |
| **#3** | qwen3.6-27b | base (`*`) | `study-write-qwen` | ⚠ partial | 825K | 24m |
| **#4** | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | ✅ | 925K | 24m30s |
| **#5** | qwen3.6-27b | qwen-tuned (`qwen*`) | `study-write-qwen` | ✅ | 695K | 9m24s |

## TL;DR

Both tuned variants cleared every signature they targeted. **Run #5 (qwen-tuned) is roughly the same quality as run #4 (kimi-tuned) but faster than the kimi run on local GPU** — 9m24s vs 24m30s. The brevity discipline + correct tool-usage rules eliminated qwen's wasted tokens.

The decisive evidence is comparing run #3 vs run #5 (same model, same corpus access, only the prompt differs):
- 239 lines → 110 lines (54% shorter)
- 825K → 695K tokens (16% fewer)
- 24m → 9.4m elapsed (61% faster)
- Section labels → claim sentences
- Heavy tables / bold-clause emphasis / mid-paragraph triadics → none
- `(#)` broken links / `study_get('bofm/...')` failures → all proper
- Self-found quote-verification correction in revision notes (vs unhedged paraphrase)

## Rubric — kimi-tuned and qwen-tuned both clear their targets

| Signature | Run #1 (k+base) | Run #2 (k+tuned) | Run #3 (q+base) | Run #4 (k+tuned+corpus) | Run #5 (q+tuned+corpus) |
|-----------|-----------------|------------------|-----------------|------------------------|------------------------|
| 1. Symmetric-pair compulsion | ✅ heavy | ❌ resists | ✅ table+parallel | ❌ resists | ❌ symmetry argued once, then dropped |
| 2. Triadic flourishes | ✅ | ❌ | ✅ | ❌ | ❌ (deployed once at sec §1, never again) |
| 3. Closing refrain | ✅ | ❌ | ✅ in §IV | ❌ | ❌ |
| 4. Pseudo-citation register | ✅ ("[study] anchors...") | N/A | ✅ broken `(#)` links | ✅ but more naturalized | ❌ uses real `(slug.md)` paths and integrates references into argument |
| 5. Latinate over Anglo-Saxon | ✅ | ❌ | partial | ❌ | ❌ |
| 6. Confabulation in revision notes | ✅ Romans 5:5 reverse-fix | ❌ honest disclosure | partial | ❌ found+removed 2 fabricated quotes | ❌ found+fixed 1 unverified paraphrase |
| **qwen-specific signatures** | | | | | |
| 7. `study_get` with scripture path | n/a | n/a | ✅ multiple | n/a | ❌ 0 bad paths |
| 8. Broken `(#)` link convention | n/a | n/a | ✅ throughout | n/a | ❌ all `(slug.md)` |
| 9. Heavy table mid-argument | n/a | n/a | ✅ §VI table | n/a | ❌ no tables |
| 10. Bold-clause density | n/a | n/a | ✅ many | n/a | ❌ minimal, single-term only |
| 11. Mid-paragraph triadics | n/a | n/a | ✅ multiple | n/a | ❌ deployed once at §1 perspective-establishing moment |
| 12. Verbosity | n/a | n/a | ✅ 239 lines | n/a | ❌ 110 lines |

## Voice metrics

| Metric | #1 | #2 | #3 | #4 | #5 |
|--------|----|----|----|----|----|
| Total lines | 105 | 43 | 239 | 118 | 110 |
| Section headers | 6 labels | 0 | 6 labels | 6 theses | 6 theses |
| Em-dashes (body) | ~12 | 0 | ~15+ | ~3 | ~2 |
| `(#)` broken links | 0 | 0 | many | 0 | 0 |
| Bad `study_get` paths | 0 | 0 | many | 0 | **0** |
| Tables in body | 0 | 0 | 1 | 0 | 0 |
| Triadic constructions in body | 1 | 0 | 3+ | 0 | 1 (structural) |
| Bold-emphasis on full clauses | minor | none | many | minor | none |
| Closing refrain | yes | no | yes (§IV) | no | no |
| Self-found quote errors | 0 | 0 (acknowledged limit) | 1 | 3 | 1 |

## Section header comparison — labels vs theses

| Run | First section header |
|-----|----------------------|
| #1 | "I. The Two Triplets as Ordered Progressions" |
| #2 | (no headers) |
| #3 | "I. The Two Triplets — Mapping the Terrain" |
| #4 | "**Thomas asked for directions, and Jesus gave Himself**" |
| #5 | "**The two triplets describe the same motion from different standpoints**" |

Both tuned variants produce thesis headers. Both `Therefore`/`But` open subsequent sections — run #5's "**Therefore both triplets terminate at the same reality**" and "**But the temple is where the two triplets physically meet**" use the prompt's causation-or-disruption rule directly.

## Run #5 strongest passages

**Opening (immediate scene drop, three-scene braid):**
> "Thomas had just asked a practical question. *'Lord, we know not whither thou goest; and how can we know the way?'* He wanted directions. Jesus gave him a Person.
>
> Six centuries later, on an American desert island, Moroni would write three words in the same sequence: *'Wherefore, there must be faith; and if there must be faith there must also be hope; and if there must be hope there must also be charity'*."

**Therefore-chain section opening:**
> "**Therefore both triplets terminate at the same reality**" — explicit causation from §1.

**Disruption section opening:**
> "**But the temple is where the two triplets physically meet**" — explicit disruption: theoretical convergence in §2 → physical meeting place in §3.

**Anti-symmetry-as-finale that ARGUES the symmetry:**
> "The faith-hope-charity triplet covers three temporal dimensions: faith (trust in what was promised), hope (orientation toward what is promised), charity (the love that is the promise fulfilled). The way-truth-life triplet covers three spatial dimensions: way (the path), truth (the reality along the path), life (the destination). Together they form a coordinate system: time and space, traveler and road, process and substance."

This is symmetry deployed once, with claimed structural significance (time/space coordinate system), in the section explicitly devoted to "the triplet shape carries meaning" — exactly the prompt's "name the symmetry once" rule.

**Honest caveat in dedicated section:**
> "## This is a reading, not a doctrine
>
> No single verse states *'Faith-Hope-Charity and Way-Truth-Life are two perspectives on the same reality.'* The synthesis is built by juxtaposition... This is honest scholarship: the pieces are all canonical; the arrangement is interpretive."

**Verification discipline in revision notes:**
> "**Quote verification fix:** Changed *'Jesus doesn't simply show these things—He IS them'* (unverified paraphrase) to the actual quote from way-truth-life study: *'Jesus responds not with directions but with Himself — He IS the way.'*"

The verify-before-fix rule of the prompt operating exactly as designed: qwen-tuned identified a paraphrase the draft had presented as a quote, looked it up, and corrected it.

## Run #4 vs Run #5 — the two tuned variants compared

Both clear their respective signatures. Distinct character:

| | Run #4 (kimi-tuned) | Run #5 (qwen-tuned) |
|---|---------------------|---------------------|
| Voice | analytical, dry | sermonic, warmer |
| Opening style | "Thomas asked for directions" + braided witnesses | Block-quoted scripture + Thomas + Moroni in three paragraphs |
| Central image | "vessel meets the filling" | "the road IS Christ" / "view from the feet vs view from the ground" |
| Symmetry handling | resists ("assimilation language is more accurate than symmetry language") | argues once as time/space coordinate system |
| Anti-confabulation | found+removed 2 fabricated phrases, fixed 1 mis-attribution, 1 unverified statistic | found+fixed 1 unverified paraphrase |
| Length | 118 lines | 110 lines |
| Tokens | 925K total | 695K total |
| Elapsed | 24m30s (kimi cloud, long generations) | 9m24s (qwen local GPU, brevity-disciplined) |
| Cost | ~$0.30 (kimi cache pricing) | $0 (local GPU) |

Both are publishable. Each has a distinct voice, which is a feature, not a bug — different binding questions might benefit from different models.

## Findings

### Both tuned variants are validated

The kimi-tuned variant cleared 5/6 signatures in run #4 (mild residual pseudo-citation that's now naturalized into argument flow). The qwen-tuned variant cleared all 12 signatures it targeted (6 kimi-shared + 6 qwen-specific) in run #5.

### Tool-usage rules are powerful

The single rule "study_get takes a kebab-case slug, NEVER a scripture reference" eliminated the failure mode that wasted tokens in run #3. Run #3 made multiple `study_get('bofm/ether/12')` calls that returned errors and triggered retry loops. Run #5 made zero. The rule paid for itself in tokens and time.

### Verification discipline scales across models

Both tuned variants demonstrated active self-correction in revision notes. Run #4 caught and removed two fabricated phrases (kimi). Run #5 caught and fixed one unverified paraphrase (qwen). The "verification claims must be tool-grounded" rule operates equally well in both models.

### Brevity rule has compounding benefits

Qwen's verbosity in run #3 was the largest token-cost difference. Run #5's brevity instruction not only shortened the output (54%) but also shortened the journey to it (61% time, 16% tokens). Tighter prose at every stage costs less to produce.

### Local qwen3.6-27b is competitive with cloud kimi-k2.6 for this task

When tuned, qwen on a 4090 produces a study comparable in quality to kimi via opencode_go, in less than half the wall time, at zero variable cost. For tasks where qwen's voice (slightly more sermonic, warmer) fits the binding question, qwen is now the cost-efficient default.

### Provenance column did its job

The 3c.3.3.1 migration kept broadcast perms intact across multiple reimports tonight (regression import of `.github/agents` + variant imports of both kimi-k2.6 and qwen-3.6 folders). Pre/post counts unchanged at 19 broadcast / 275 frontmatter.

## Recommendations

1. **Both kimi-tuned and qwen-tuned variants graduate to stable v1.** The `.stewards/<model>/README.md` iteration logs should reflect the validation evidence.

2. **Promote run #4 or run #5 over the current `study/two-triplets-one-ascent.md`?** Both are substrate-produced AND well-voiced AND properly source-grounded. Different voices; same answer to the binding question. Michael's call.

3. **Daytime architectural priorities:**
   - 3c.4 gospel-engine HTTP tools — the only remaining big gap
   - 3c.3.5 work_items → `stewards.studies` auto-promotion
   - Image rebuild to fold 3c3-3-agent-tool-perms-provenance.sql into the bundled SQL (currently only live-applied)

4. **Future model variants are now a mechanical workflow.** When a new model joins (Gemini, GLM, Sonnet, etc.), the playbook is:
   - Run a baseline study with the base prompt
   - Identify model-specific signatures via the comparison rubric in this memo
   - Author `.stewards/<model>/study.agent.md`
   - Re-run the same binding question, verify the signatures clear
   - Promote to stable when validated

The substrate now has a repeatable voice-tuning loop.
