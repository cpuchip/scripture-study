# Phase 0 Experiment Results — Talk Context Testing

*March 29, 2026*

## Experiment Setup

**Model:** nemotron-3-nano (32768 MaxTokens, Temperature 0.2, NoThink)
**Prompt:** titsw-enriched-talk.md (enriched summary with TEACHING_PROFILE)
**Test talks:** Kearon "Receive His Gift", Bednar "Their Own Judges", Holland "And Now I See"

## Ground Truth

| Talk | teach | help | love | spirit | doctrine | invite |
|------|-------|------|------|--------|----------|--------|
| Kearon | 8 | 8 | 4 | 3 | 7 | 8 |
| Bednar | 5 | 5 | 2 | 3 | 9 | 6 |
| Holland | 7 | 6 | 4 | 7 | 6 | 3 |

## Results by Experiment

### T0 — No context (baseline)

| Talk | teach | help | love | spirit | doctrine | invite | MAE |
|------|-------|------|------|--------|----------|--------|-----|
| Kearon | 9(+1) | 9(+1) | 8(**+4**) | 8(**+5**) | 8(+1) | 9(+1) | 2.17 |
| Bednar | 9(**+4**) | 9(**+4**) | 7(**+5**) | 8(**+5**) | 8(-1) | 8(**+2**) | 3.50 |
| Holland | 9(**+2**) | 9(**+3**) | 7(**+3**) | 8(+1) | 7(+1) | 6(**+3**) | 2.17 |
| **Overall MAE** | | | | | | | **2.61** |

Massive inflation everywhere. The enriched prompt without any context produces score ceilings (8-9) on nearly every dimension. Bednar is the worst — doctrine should be 9 but every dimension reads as 8-9.

### T1 — titsw-framework.md only

| Talk | teach | help | love | spirit | doctrine | invite | MAE |
|------|-------|------|------|--------|----------|--------|-----|
| Kearon | 9(+1) | 8(0) | 8(**+4**) | 8(**+5**) | 8(+1) | 9(+1) | 2.00 |
| Bednar | 7(**+2**) | 7(**+2**) | 7(**+5**) | 6(**+3**) | 7(-2) | 6(0) | 2.33 |
| Holland | 8(+1) | 8(**+2**) | 7(**+3**) | 8(+1) | 7(+1) | 6(**+3**) | 1.83 |
| **Overall MAE** | | | | | | | **2.06** |

Framework helps modestly. Bednar teach dropped from 9→7, spirit from 8→6. The detailed score anchors are providing some calibration. But love/spirit still inflated across the board.

### T2 — gospel-vocab.md only

| Talk | teach | help | love | spirit | doctrine | invite | MAE |
|------|-------|------|------|--------|----------|--------|-----|
| Kearon | 9(+1) | 9(+1) | 9(**+5**) | 8(**+5**) | 8(+1) | 9(+1) | 2.33 |
| Bednar | 7(**+2**) | 7(**+2**) | 8(**+6**) | 6(**+3**) | 7(-2) | 7(+1) | 2.67 |
| Holland | 8(+1) | 8(**+2**) | 9(**+5**) | 8(+1) | 7(+1) | 6(**+3**) | 2.17 |
| **Overall MAE** | | | | | | | **2.39** |

**Confirmed: gospel-vocab is the inflation culprit.** Love scores are the worst of any experiment (Kearon 9, Bednar 8, Holland 9 — all 5+ above ground truth). The theological patterns document causes the model to over-read love/spirit connections in explicit talk content. This is exactly the mechanism we hypothesized.

### T3 — Talk rhetorical context (NEW)

| Talk | teach | help | love | spirit | doctrine | invite | MAE |
|------|-------|------|------|--------|----------|--------|-----|
| Kearon | 7(-1) | 7(-1) | 8(**+4**) | 7(**+4**) | 6(-1) | 7(-1) | 2.00 |
| Bednar | 7(**+2**) | 7(**+2**) | 7(**+5**) | 7(**+4**) | 8(-1) | 6(0) | 2.33 |
| Holland | 7(0) | 7(+1) | 8(**+4**) | 8(+1) | 7(+1) | 6(**+3**) | 1.67 |
| **Overall MAE** | | | | | | | **2.00** |

Good improvement. Teach scores are more moderate. Bednar doctrine correctly rose to 8. The rhetorical context helps mode identification — Bednar correctly identified as "doctrinal" mode. But love and spirit remain persistently inflated.

### T4 — Calibration context (scored example)

| Talk | teach | help | love | spirit | doctrine | invite | MAE |
|------|-------|------|------|--------|----------|--------|-----|
| Kearon | 5(**-3**) | 7(-1) | 6(**+2**) | 6(**+3**) | 5(-2) | 7(-1) | 2.00 |
| Bednar | 6(+1) | 7(**+2**) | 6(**+4**) | 5(**+2**) | 7(-2) | 7(+1) | 2.00 |
| Holland | 7(0) | 7(+1) | 7(**+3**) | 6(-1) | 5(-1) | 6(**+3**) | 1.50 |
| **Overall MAE** | | | | | | | **1.83** |

**Best overall MAE.** The calibration example provides a strong anchor. The model sees a scored talk and calibrates around it. Key wins:
- Bednar teach: 6 (GT 5, +1) — down from 9 in T0
- Bednar spirit: 5 (GT 3, +2) — down from 8 in T0
- Holland teach: 7 (GT 7, exact!)
- Love inflation reduced but still persistent (+2 to +4)

Key trade-off: Kearon teach dropped to 5 (GT 8, -3). The calibration example happens to be a Kearon talk scored at teach=5. The model is over-anchoring to the example's scores for Kearon specifically. For other talks, the anchoring works well.

### T5 — T3 + T4 combined

| Talk | teach | help | love | spirit | doctrine | invite | MAE |
|------|-------|------|------|--------|----------|--------|-----|
| Kearon | 5(**-3**) | 7(-1) | 7(**+3**) | 6(**+3**) | 6(-1) | 7(-1) | 2.00 |
| Bednar | 5(0) | 7(**+2**) | 7(**+5**) | 5(**+2**) | 5(**-4**) | 7(+1) | 2.33 |
| Holland | 6(-1) | 7(+1) | 7(**+3**) | 5(-2) | 5(-1) | 7(**+4**) | 2.00 |
| **Overall MAE** | | | | | | | **2.11** |

Worse than T4 alone. Combining contexts doesn't help — the rhetorical context adds noise that dilutes the calibration anchor's effectiveness. Bednar doctrine collapsed from 7→5 (GT 9, -4). Holland invite jumped to 7 (GT 3, +4).

## MAE Summary

| Experiment | Context | Overall MAE | Best/Worst Individual |
|------------|---------|-------------|----------------------|
| T0 | None | **2.61** | Worst: Bednar 3.50 |
| T1 | Framework | **2.06** | Best: Holland 1.83 |
| T2 | Gospel-vocab | **2.39** | Worst: Bednar 2.67 |
| T3 | Rhetorical | **2.00** | Best: Holland 1.67 |
| **T4** | **Calibration** | **1.83** | **Best: Holland 1.50** |
| T5 | T3+T4 | **2.11** | Mixed |

## Key Findings

### 1. Context helps talks (overturning previous conclusion)

The prior conclusion "context hurts talks" was based on applying *scripture-focused* context (gospel-vocab + titsw-framework). That conclusion holds — T2 (gospel-vocab alone) is the second-worst experiment. But *talk-specific* context clearly helps. T3 and T4 both beat T0 baseline.

### 2. Calibration is the most effective context type

A single scored example (T4) produced the best MAE (1.83). The model anchors to the example's score distribution rather than inflating to ceiling. This is essentially one-shot learning for score calibration.

### 3. Gospel-vocab inflates love catastrophically for talks

Love scores with gospel-vocab: Kearon 9 (+5), Bednar 8 (+6), Holland 9 (+5). The theological patterns cause the model to read "sheddeth itself abroad" type connections into every mention of God's love in explicit talks. This confirms gospel-vocab should NOT be used for talk scoring.

### 4. Love and spirit remain persistently inflated

Even the best experiment (T4) shows love inflation of +2 to +4 across all talks. This appears to be an inherent model bias (RLHF positive bias for "helpful" dimensions) that context alone cannot fix. The enriched prompt's minimal scoring guidance (just 4 anchor points) lacks the detailed calibration that the full v5 scoring prompt provides.

### 5. Combining contexts doesn't help

T5 (T3+T4) performed worse than T4 alone. More context = more noise. The calibration example alone was the strongest signal.

### 6. Qualitative outputs improved with context

Across T3 and T4:
- Bednar consistently identified as "doctrinal" mode (correct)
- Holland correctly identified "story→doctrine→invitation" pattern  
- Dominant dimension identification improved (Bednar: doctrine in T3, but not in T4/T5)

## Decision for Phase 1

**Use T4 (calibration context) for the talk enrichment batch.**

But the calibration example needs refinement:
1. **Use a different talk** for the calibration example — not Kearon, since we're scoring Kearon and the model over-anchors. Use Uchtdorf or another well-known talk with moderate scores.
2. **Consider 2-3 calibration examples** spanning different score distributions (one high-teach, one high-doctrine, one moderate-across-the-board).
3. **Love/spirit inflation is a prompt problem, not a context problem.** The enriched prompt's "3=present, 5=clear theme, 7=defining feature, 9=rare" guidance is too sparse. For the batch run, consider adding a line: "Most dimensions score 3-5. A 7+ requires the dimension to be the defining feature of the talk, not just present."

### Estimated impact on batch

With T4-style calibration, we expect batch MAE around 1.8-2.0 (vs projected 2.6+ without context). This is a meaningful improvement for downstream search quality.

## Files Created

- `prompts/titsw-enriched-talk.md` — Enriched talk prompt with TEACHING_PROFILE
- `context-t1/01-titsw-framework.md` — Framework only (copy)
- `context-t2/gospel-vocab.md` — Gospel vocab only (copy)
- `context-t3/talk-rhetorical-context.md` — NEW: Conference talk patterns
- `context-t4/talk-calibration-context.md` — NEW: Scored Kearon example
- `context-t5/` — T3+T4 combined

## Raw Results

All 18 experiment runs saved in `results/` as JSON with tag T0-T5.
