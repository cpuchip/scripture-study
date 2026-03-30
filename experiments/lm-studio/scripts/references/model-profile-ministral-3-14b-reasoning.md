# Model Profile: ministralai/ministral-3-14b-reasoning

*Created Mar 29, 2026 after full 6-piece TITSW suite evaluation*

## Identity

| Property | Value |
|----------|-------|
| Model | `mistralai/ministral-3-14b-reasoning` |
| Architecture | mistral3, 14B parameters |
| Size | 9.12 GB (Q4_K_M) |
| Context tested | 65,536 tokens |
| Speed | ~50-63 tok/s generation, ~48 tok/s overall |
| TTFT | 1.1-3.5s (fast — no reasoning token overhead) |
| Reasoning tokens | 0 (all output is content) |

## Overall Performance

**Best MAE of all tested models: 1.54** (vs nemotron 1.65, GPT-OSS 1.63, claude-27b 1.63, claude-35b 1.65)

| Piece | teach | help | love | spirit | doctrine | invite | MAE |
|-------|-------|------|------|--------|----------|--------|-----|
| Alma 32 | 8(+0.5) | 7(-1) | 6(-1.5) | **9(+5.5)** | 8(-0.5) | 5(-3.5) | 2.08 |
| Bednar | **8(+3)** | 4(-1) | **3(+1)** | 5(+2) | **9(0)** | **6(0)** | 1.17 |
| DC 121 | **8(+3)** | **5(0)** | 6(+1) | **7(+3)** | 9(+1) | 4(-3) | 1.83 |
| Holland | 8(+1) | 7(+1) | 5(+1) | 6(-1) | **9(+3)** | 6(+3) | 1.67 |
| Kearon | 7(-1) | **8(0)** | 5(+1) | 6(+3) | 8(+1) | 9(+1) | 1.17 |
| 3 Ne 17 | 8(-1) | **9(0)** | **9(0)** | 5(-3) | 6(+2) | 7(+2) | 1.33 |

Ground truth: Alma 32 (7.5/8/7.5/3.5/8.5/8.5), Bednar (5/5/2/3/9/6), DC 121 (5/5/5/4/8/7), Holland (7/6/4/7/6/3), Kearon (8/8/4/3/7/8), 3 Ne 17 (9/9/9/8/4/5)

## Bias Profile

| Dimension | Avg Error | Direction | Worst Case | Pattern |
|-----------|-----------|-----------|------------|---------|
| teach | +1.5 | inflates | +3 (Bednar, DC121) | Floor at 7-8. Cannot differentiate 5-level teach from 9-level |
| help | +0.17 | neutral | ±1 | **Most accurate dimension** |
| love | +0.58 | slight inflate | +1.5 | Reasonable accuracy |
| spirit | variable | ±3-5 | +5.5 (Alma 32), -3 (3 Ne 17) | **Confuses content about Spirit with experience of Spirit** |
| doctrine | +1.83 | inflates | +3 (Holland) | Floor at 6. Everything is "doctrinal" |
| invite | -0.5 | slight deflate | -3.5 (Alma 32) | Misses structured implicit invitations |

### Key Bias: Teach & Doctrine Inflation

Every piece gets teach=7-8. Everything with scripture citations gets doctrine 6-9.
The model cannot distinguish "doctrine is present" (5) from "doctrine is the defining feature" (9).

### Key Bias: Spirit Confusion

Conflates *content about* spiritual things with *experience of* the Spirit.
- Alma 32 "feel these swelling motions" → spirit=9 (GT=3.5). That's *describing* the Spirit's effects in didactic mode, not *creating space* for the Spirit.
- 3 Nephi 17 spirit=5 (GT=8) — where the Spirit is literally visible as fire, prayer exceeds language, and the multitude is overcome with joy.

### Insight: Reasoning > Scores

The model's reasoning often correctly identifies what should push a score down, then scores high anyway.
Example: DC 121 reasoning says "declares rather than models" and "suffering of Joseph Smith is implied but not deeply explored" — then still scores teach=8 and spirit=7. The reasoning knows what a 5 looks like but the scoring defaults to 7-8.

## Qualitative Strengths

1. **Mode identification is excellent.** Correctly identifies enacted (Holland, 3 Ne 17), declared (DC 121), doctrinal (Bednar) modes.
2. **Pattern recognition is strong.** problem→principle→promise vs story→doctrine→invitation correctly assigned.
3. **Specific citations.** Cites actual verse references and moments from the text, not vague summaries.
4. **Strengths/adequacies structure.** Voluntarily adds where talk is strong vs merely adequate — useful meta-information.
5. **Key quote selection.** Consistently picks the most memorable/powerful quotes from each piece.
6. **Format compliance.** Clean, parseable output every time. No reasoning token overhead.

## Qualitative Weaknesses

1. **Teach default at 8.** Treats any quality teaching about Christ as ceiling-level.
2. **Doctrine inflation.** Confuses "references scripture" with "redefines understanding of a doctrine."
3. **Spirit confusion.** Cannot distinguish teaching *about* the Spirit from teaching *by* the Spirit.
4. **Invite blindness to implicit structure.** Alma 32 is literally an escalating invitation ("exercise a particle of faith" → "desire to believe" → "nourish the word") but gets invite=5.
5. **Slightly verbose.** Output averages 700-850 tokens vs 600 word limit. Not a deal-breaker but stretches the constraint.

## Tuning Targets

The model is receptive to structured prompts and follows format well. Worth testing:
1. **Calibration anchors** — concrete examples of what 3, 5, 7, 9 mean per dimension
2. **Anti-inflation guidance** — explicit warning about teach/doctrine defaults
3. **Spirit distinction** — teaching *about* vs teaching *by* the Spirit
4. **Invite recognition** — distinguish implicit structured invitations from explicit calls to action
5. **Score-from-reasoning** — instruction to derive scores from reasoning rather than defaulting

## Prompt Tuning Experiments (Mar 29, 2026)

Three prompt variants tested against all 6 pieces at temperature 0.2 on ministral-3-14b-reasoning (95k context).

### Variant Descriptions

| Variant | Key Intervention | Hypothesis |
|---------|-----------------|------------|
| **titsw-calibrated** | Per-dimension 4-level anchor tables + spirit ABOUT/BY distinction + distribution warning | Heavy calibration will reduce teach/doctrine inflation and fix spirit confusion |
| **titsw-anchored** | Two-shot reference examples (Bednar=doctrinal, Holland=testimony) + derive-from-reasoning instruction | Concrete examples will anchor the scale without overcorrecting |
| **titsw-deflate** | Minimal: 3 scoring rules (derive from reasoning, distribution warning, ABOUT vs DOES) | A light nudge may be enough if the model's reasoning is already accurate |

### Results Summary

| Variant | MAE | vs Baseline | Change |
|---------|-----|-------------|--------|
| Baseline (enriched-talk-reasoning) | 1.54 | — | — |
| **titsw-calibrated** | **1.32** | **-14%** | **Best** |
| titsw-anchored | 1.43 | -7% | Modest improvement |
| titsw-deflate | 1.71 | +11% | Worse than baseline |

### Per-Piece Breakdown

**Calibrated:**

| Piece | teach | help | love | spirit | doctrine | invite | MAE |
|-------|-------|------|------|--------|----------|--------|-----|
| Alma 32 | 5(-2.5) | 7(-1) | **3(-4.5)** | 5(+1.5) | 7(-1.5) | 7(-1.5) | 2.08 |
| Bednar | **5(0)** | 6(+1) | 4(+2) | **3(0)** | 7(-2) | 5(-1) | 1.00 |
| DC 121 | **5(0)** | 4(-1) | **5(0)** | **4(0)** | 7(-1) | 5(-2) | 0.67 |
| Holland | **7(0)** | 5(-1) | **4(0)** | **3(-4)** | 5(-1) | 4(+1) | 1.17 |
| Kearon | 5(-3) | 7(-1) | 5(+1) | **3(0)** | **7(0)** | 7(-1) | 1.00 |
| 3 Ne 17 | **9(0)** | **5(-4)** | 7(-2) | **5(-3)** | **7(+3)** | **5(0)** | 2.00 |

**Anchored:**

| Piece | teach | help | love | spirit | doctrine | invite | MAE |
|-------|-------|------|------|--------|----------|--------|-----|
| Alma 32 | 8(+0.5) | **4(-4)** | **3(-4.5)** | 5(+1.5) | 9(+0.5) | **4(-4.5)** | 2.58 |
| Bednar | 7(+2) | **5(0)** | 4(+2) | 5(+2) | 8(-1) | 5(-1) | 1.33 |
| DC 121 | **5(0)** | 4(-1) | 3(-2) | **4(0)** | 7(-1) | 4(-3) | 1.17 |
| Holland | 8(+1) | 7(+1) | 5(+1) | 8(+1) | 7(+1) | 4(+1) | 1.00 |
| Kearon | 5(-3) | 7(-1) | **4(0)** | 5(+2) | 6(-1) | **8(0)** | 1.17 |
| 3 Ne 17 | 5(-4) | 7(-2) | **9(0)** | **8(0)** | 6(+2) | **5(0)** | 1.33 |

**Deflate:**

| Piece | teach | help | love | spirit | doctrine | invite | MAE |
|-------|-------|------|------|--------|----------|--------|-----|
| Alma 32 | 8(+0.5) | **4(-4)** | **3(-4.5)** | 5(+1.5) | 9(+0.5) | **4(-4.5)** | 2.58 |
| Bednar | 7(+2) | 4(-1) | 5(+3) | **3(0)** | **9(0)** | 5(-1) | 1.17 |
| DC 121 | 7(+2) | 4(-1) | **5(0)** | 5(+1) | 9(+1) | 3(-4) | 1.50 |
| Holland | 8(+1) | 7(+1) | 6(+2) | 5(-2) | 8(+2) | 6(+3) | 1.83 |
| Kearon | 7(-1) | 9(+1) | 6(+2) | 5(+2) | 4(-3) | **8(0)** | 1.50 |
| 3 Ne 17 | 5(-4) | 7(-2) | **9(0)** | 6(-2) | 5(+1) | 4(-1) | 1.67 |

### What We Learned

**The model is responsive to prompt calibration.** The calibrated prompt reduced MAE by 14%, confirming the model will follow structured scoring guidance without being overwhelmed by it.

**Heavy calibration > two-shot anchoring > light nudge.** The per-dimension anchor tables with explicit spirit distinction were more effective than reference examples. Two-shot examples helped Holland (1.67→1.00) but hurt Alma 32 (2.08→2.58). The light nudge made things worse overall.

**Calibrated prompt shifted bias from inflation to deflation.** Baseline had systematic +1.5 teach, +1.1 doctrine, +1.6 spirit inflation. Calibrated overcorrected to -0.9 teach, -0.4 doctrine, -0.9 spirit deflation. The net MAE still improved because fewer extreme errors (spirit +5.5 on Alma 32 is gone), but new deflation errors appeared (help -4 on 3 Ne 17).

**Spirit confusion improved but not solved.** Calibrated: Alma 32 spirit dropped from 9→5 (GT=3.5, still +1.5 but way better than +5.5). Holland spirit dropped to 3 (GT=7, now off by -4). The model learned "content about Spirit ≠ experience of Spirit" but overcorrected on Holland, where testimony IS the Spirit working.

**The 3 Ne 17 problem.** The calibrated prompt's biggest regression was 3 Ne 17: help dropped 9→5 (GT=9), spirit dropped 5→5 (stayed wrong). The calibration cues about "most talks have 1-2 strong dimensions" caused the model to suppress legitimate 9-level scores.

**Alma 32 remains the hardest piece.** All four variants score MAE 2.08-2.58 on Alma 32. The typological depth (seed=Christ, tree=tree of life) is invisible to the model at any prompt level. This requires enriched context, not just prompt engineering.

### Next Steps

1. **Hybrid prompt**: Combine calibrated's per-dimension anchors with a softer distribution warning ("some pieces genuinely deserve multiple high scores").
2. **Context injection**: Test whether injecting cross-reference context (1 Ne 11:22-25, John 1:1, Alma 33:22-23) fixes the Alma 32 problem specifically.
3. **Temperature sweep**: Run calibrated at T=0.1, 0.3, 0.4 to see if temperature affects the inflation/deflation balance.

## Comparison Notes

| vs Model | Advantage | Disadvantage |
|----------|-----------|--------------|
| nemotron | Better MAE, 2x smaller, no RLHF love inflation | Slower tok/s, teach/doctrine inflation instead |
| GPT-OSS | Same MAE tier, 1.5x smaller | Misses invite where GPT-OSS sometimes gets it |
| claude-27b | Better MAE, 2x smaller, no censorship risk | Reasoning quality slightly less nuanced |
| claude-35b | Better MAE, 2x smaller | Less reasoning depth per dimension |
