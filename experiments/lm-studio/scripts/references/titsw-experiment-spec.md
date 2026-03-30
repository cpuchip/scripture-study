# TITSW Scoring Experiments — Complete Reference

*Created Mar 29, 2026. Comprehensive capture of all model testing, prompt tuning, and architectural experiments for the Teaching in the Savior's Way (TITSW) automated scoring system.*

---

## Production Configuration

| Parameter | Value |
|-----------|-------|
| **Model** | `mistralai/ministral-3-14b-reasoning` (14B params, 9.12 GB Q4_K_M) |
| **Prompt** | `prompts/titsw-calibrated.md` |
| **Temperature** | 0.2 |
| **Context length** | 65,536 tokens |
| **Speed** | ~50-63 tok/s generation |
| **Overall MAE** | **1.32** (best achieved across all experiments) |

---

## What This System Does

Scores gospel content (conference talks, scripture chapters) on 6 teaching dimensions from the *Teaching in the Savior's Way* manual:

| Dimension | What it measures |
|-----------|-----------------|
| teach_about_christ | How central Christ is to the content |
| help_come_unto_christ | How effectively it moves people toward Christ |
| love | Demonstrated/taught love (pastoral, not just doctrinal assertion) |
| spirit | Experience of the Spirit (teaching *by* the Spirit, not just *about* it) |
| doctrine | Depth of doctrinal engagement |
| invite | Directness and specificity of invitations to act |

Scale: 0-9. Ground truth scored by Michael across 6 test pieces.

---

## Ground Truth Test Suite

6 pieces spanning scripture, conference talks, and Restoration scripture. Scored by Michael with deep study (full typological analysis, footnote chasing, cross-referencing).

| Piece | teach | help | love | spirit | doctrine | invite | Character |
|-------|-------|------|------|--------|----------|--------|-----------|
| Alma 32 | 7.5 | 8 | 7.5 | 3.5 | 8.5 | 8.5 | Typological Christ (seed=Word=Christ). Hardest piece — surface text barely mentions Christ. |
| Bednar "Their Own Judges" | 5 | 5 | 2 | 3 | 9 | 6 | Heavy doctrine, low affect. Tests whether model resists inflation. |
| D&C 121 | 5 | 5 | 5 | 4 | 8 | 7 | Revelation in suffering. Moderate across dimensions. |
| Holland "And Now I See" | 7 | 6 | 4 | 7 | 6 | 3 | Testimony-driven. Spirit IS the teaching mode. Tests spirit detection. |
| Kearon "Receive His Gift" | 8 | 8 | 4 | 3 | 7 | 8 | Explicit Christ-centered. High teach/help/invite, low spirit/love. |
| 3 Nephi 17 | 9 | 9 | 9 | 8 | 4 | 5 | Christ literally present. Legitimate 9-level scores across multiple dimensions. |

Ground truth details with full typological analysis: `references/ground-truth-alma32-kearon.md`

---

## All Models Tested

### Model Comparison Summary

| Model | Size | Speed | TITSW MAE | Format | Key Trait |
|-------|------|-------|-----------|--------|-----------|
| **ministral-3-14b-reasoning** | 9.12 GB | 50-63 tok/s | **1.54 → 1.32** | Perfect | **WINNER.** Smallest, responsive to calibration, no RLHF love inflation |
| nemotron-3-nano | ~8 GB | 88-160 tok/s | 1.65 | Good | Fastest. Severe RLHF love/spirit inflation (+2 to +5). Thinking mode is poison. |
| GPT-OSS (qwen3.5-35b-a3b) | ~20 GB | 48-52 tok/s | 1.63 | Good | Solid all-rounder. Larger than needed. Sometimes catches invite. |
| claude-opus-distill-27b | ~16 GB | — | 1.63 | Good | Matched GPT-OSS quality. Censorship risk on religious content. |
| claude-opus-distill-35b-a3b | ~20 GB | — | 1.65 | Good | No improvement over 27b despite more params. |
| GPT-OSS-Claude-Distill | ~20 GB | — | 1.79 | Good | Hybrid distill was WORSE than either parent. Don't mix distillations. |
| Nemo-Extreme | — | — | DISQUALIFIED | **Failed** | Could not produce valid JSON consistently. Format compliance failures. |

### Model-Specific Notes

**ministral-3-14b-reasoning (production model):**
- Best raw MAE (1.54) of any model before prompt tuning
- Responsive to prompt calibration → 1.32 with calibrated prompt (14% improvement)
- Mode identification excellent (enacted/declared/doctrinal correctly assigned)
- Pattern recognition strong (story→doctrine→invitation vs problem→principle→promise)
- Key weakness: teach/doctrine inflation (floor at 7-8), spirit confusion (about vs by)
- 0 reasoning tokens (all output is content — no hidden thinking overhead)
- Full model profile: `references/model-profile-ministral-3-14b-reasoning.md`

**nemotron-3-nano (Phase 0 workhorse, not production):**
- Used for all Phase 0 context experiments (T0-T6)
- RLHF training causes systematic love/spirit inflation: +2 to +5 above ground truth
- T4 calibration context achieved MAE 1.83 (best for nemotron)
- T6 experiment proved same-speaker anchoring: Bednar calibration → Bednar scores MAE=0.00 (copied itself)
- **Thinking mode (NoThink=false) is harmful.** Think tokens consume context without improving scores.
- 3x faster than MoE alternatives — speed advantage for batch processing
- Not used for production because love/spirit inflation is structural and not prompt-fixable

**GPT-OSS (qwen3.5-35b-a3b):**
- Ecosystem baseline (Qwen family, strong instruction following)
- Sometimes catches invite patterns that ministral misses
- 2x the size for marginal quality gain — not worth it for batch processing
- 262k context is excessive for talk scoring (~8k tokens per talk)

**Claude distills (27b and 35b-a3b):**
- Both achieved 6/6 valid outputs, MAE 1.63 and 1.65
- No improvement from 35b over 27b despite more params — MoE routing overhead
- Potential censorship on religious content (untested at scale)
- Too large and no quality advantage over ministral

---

## Prompt Evolution

### TITSW Prompt Version History

| Version | Key Innovation | MAE | Status |
|---------|---------------|-----|--------|
| v1 (titsw.md) | Original 4-dimension scoring | — | Baseline, too vague |
| v2 | Structured rubric, clearer dimensions | — | Model comparison baseline |
| v3 | 0-9 scale, anchored rubric, new JSON fields | — | Scale expansion without classification |
| v4-v4.2 | **Classification-first logic** | 1.10 | Breakthrough: "what TYPE before what SCORE" |
| v5 (pure positive) | All positive anchors, removed classification | 1.24 | REGRESSION — proved classification essential |
| **v5.1 (hybrid)** | v4 classification + v5 positive anchors | **0.93*** | Best ever (*with context contamination) |
| v5.2 | Thinking mode test | regressed | Thinking mode is poison for nemotron |
| v5.3 | Enum labels coupled to ranges | 1.52 | Labels created bottleneck |
| v5.4 | Three-axis (modes/categories/insights) | 1.30 | Right direction — qualitative richness |
| v5.4-ctx | v5.4 with full context | 1.47 | Proved context hurts talks |
| v6 | 4 targeted fixes on v5.1 | regressed | Surgical changes have blast radius |
| enriched-talk | TITSW vocabulary in system prompt | — | Vocabulary approach for batch indexing |
| enriched-talk-reasoning | Reasoning-format output | 1.54** | Baseline on ministral |
| **titsw-calibrated** | Per-dim anchor tables + spirit ABOUT/BY | **1.32** | **PRODUCTION** |
| titsw-anchored | Two-shot examples (Bednar + Holland) | 1.43 | Modest improvement only |
| titsw-deflate | 3 minimal scoring rules | 1.71 | Worse than baseline |

\* v5.1 MAE 0.93 was contaminated by ~18KB of TITSW framework + gospel vocabulary context loaded via run-test.ps1's default context.md. The expansion 7 pieces ran without context. Token forensics (tokens_in analysis) caught this weeks later.

\*\* 1.54 was baseline for ministral-3-14b-reasoning specifically. Earlier versions tested on nemotron-3-nano.

### Key Prompt Engineering Lessons

1. **Classification-first is essential.** "What TYPE of teaching is this?" before "What SCORE?" — introduced in v4, proven essential when v5 removed it and regressed.
2. **Per-dimension anchor tables > two-shot examples > light nudges.** The calibrated prompt's 4-level tables (3/5/7/9 per dimension) were more effective than showing scored examples.
3. **Spirit needs ABOUT vs BY distinction.** Explicitly telling the model to separate "content about the Spirit" from "experience of the Spirit" reduced spirit confusion.
4. **Distribution warnings help.** "Most talks score 3-5 on most dimensions; 7+ means defining feature" reduces ceiling inflation.
5. **Surgical prompt changes have blast radius.** v6 changed one qualifier and ruined the whole prompt. Test one change at a time.
6. **Context helps scripture, hurts talks.** Gospel-vocab + titsw-framework context improved Alma 32 teach (2→6) but inflated talk scores. Different content types need different pipelines.

---

## Prompt Tuning Experiments (on ministral-3-14b-reasoning)

### Three Variants Tested

| Variant | Intervention | MAE | vs Baseline |
|---------|-------------|-----|-------------|
| Baseline (enriched-talk-reasoning) | No calibration | 1.54 | — |
| **titsw-calibrated** | Per-dimension 4-level anchors + spirit ABOUT/BY + distribution warning | **1.32** | **-14%** |
| titsw-anchored | Two-shot examples (Bednar doctrinal + Holland testimony) + derive-from-reasoning | 1.43 | -7% |
| titsw-deflate | 3 minimal scoring rules | 1.71 | +11% |

### What Changed from Baseline to Calibrated

The calibrated prompt shifted the model's bias from inflation to slight deflation:

| Dimension | Baseline Avg Error | Calibrated Avg Error | Change |
|-----------|-------------------|---------------------|--------|
| teach | +1.5 (inflated) | -0.9 (deflated) | Inflation fixed, slight overcorrection |
| help | +0.17 (neutral) | -1.0 (deflated) | Slight regression |
| love | +0.58 | -1.0 | Overcorrected on some pieces |
| spirit | +1.6 (inflated) | -0.9 (deflated) | Major improvement (Alma 32 spirit 9→5) |
| doctrine | +1.83 (inflated) | -0.4 (deflated) | Good improvement |
| invite | -0.5 | -1.0 | Slight worsening |

Net improvement because extreme errors disappeared: spirit +5.5 on Alma 32 is gone, teach +3 on Bednar/DC121 is gone.

---

## Temperature Sweep

All 6 pieces with titsw-calibrated prompt.

| Temperature | Overall MAE | Max Single Error | Catastrophic (≥5) |
|-------------|-------------|------------------|--------------------|
| 0.1 | 1.40 | 4.5 | 0 |
| **0.2** | **1.32** | **4.5** | **0** |
| 0.3 | 1.43 | 5.5 | 1 |
| 0.4 | 1.31 | 5.0 | 1 |

**Decision: T=0.2.** The 0.01 MAE improvement at T=0.4 isn't worth the catastrophic error risk (5-point error on 3 Ne 17 help). T=0.2 is the safest operating point.

Temperature doesn't fix structural problems — Alma 32 love stays at 3 across all temperatures. 3 Ne 17 help deflates at every temperature.

---

## Context Injection Experiment

### Setup
Injected cross-reference scriptures (1 Ne 8:10-12, 1 Ne 11:21-25, Alma 33:22-23, John 1:1,14) alongside Alma 32 text. Ran with titsw-calibrated at T=0.2.

### Result
| Dimension | GT | Without Context | With Context |
|-----------|-----|-----------------|--------------|
| teach | 7.5 | 5 (-2.5) | **7 (-0.5)** |
| help | 8 | 7 (-1) | 6 (-2) |
| love | 7.5 | 3 (-4.5) | 3 (-4.5) |
| spirit | 3.5 | 5 (+1.5) | 5 (+1.5) |
| doctrine | 8.5 | 7 (-1.5) | 7 (-1.5) |
| invite | 8.5 | 7 (-1.5) | 6 (-2.5) |
| **MAE** | | **2.08** | **2.08** |

### Analysis

**Teach fixed.** The model's reasoning explicitly connected "the word-seed" to "cross-referenced to 1 Nephi 8-11, John 1" and "the seed is explicitly tied to belief in Christ's Atonement (Alma 33:22-23)." Teach went from 5→7 (error reduced from -2.5 to -0.5).

**Love unchanged at 3.** Despite 1 Ne 11:22-25 being in the injected text (tree of life = love of God), the model read the words but didn't map them to the love dimension. The typological connection (seed→tree→fruit = word→Christ→love of God) is a multi-hop inference the model can't make on its own.

**Overall MAE unchanged** because help and invite deflated — the cross-references may have consumed the model's attention budget.

**Conclusion:** Context injection proves the hypothesis (model can recognize typological Christ-connections when given explicit pointers) but isn't cost-effective for deployment — it requires pre-computed cross-references per piece and doesn't improve overall MAE.

---

## Two-Stage Approach

### Theory

Stage 1 generates a typological analysis map (identifying where Christ appears explicitly, typologically, or thematically). Stage 2 scores the piece using that map as context — so the scoring model "knows" about the deeper connections.

### Implementation

| File | Purpose |
|------|---------|
| `prompts/titsw-stage1-typology.md` | Stage 1: Typological analysis prompt — outputs a structured Christ-connection map |
| `prompts/titsw-stage2-score.md` | Stage 2: Scoring prompt — calibrated + typology-aware scoring |
| `run-twostage.ps1` | Wrapper script orchestrating both stages |

### Results

**Alma 32 (the target piece):**

| Dimension | GT | Single-Pass | Two-Stage |
|-----------|-----|-------------|-----------|
| teach | 7.5 | 5 (-2.5) | **7 (-0.5)** |
| help | 8 | 7 (-1) | 7 (-1) |
| love | 7.5 | 3 (-4.5) | **5 (-2.5)** |
| spirit | 3.5 | 5 (+1.5) | 5 (+1.5) |
| doctrine | 8.5 | 7 (-1.5) | 7 (-1.5) |
| invite | 8.5 | 7 (-1.5) | 7 (-1.5) |
| **MAE** | | **2.08** | **1.67** |

Teach fixed (5→7, same as context injection). Love improved (3→5, which context injection couldn't do). MAE improved 20% on Alma 32.

**Full 6-piece suite:**

| Configuration | Overall MAE |
|--------------|-----|
| Single-pass calibrated | **1.32** |
| Two-stage | **1.54** |

Two-stage was significantly WORSE overall (1.54 vs 1.32).

### Why Two-Stage Failed at Scale

**Stage 1 finds Christ-typology everywhere.** When given explicit instructions to look for typological connections, the model finds them in every piece — even ones where the connection is surface-level (Bednar, DC 121). This inflates teach to 7 on every piece, including ones where ground truth is 5.

**Doctrine locked at 7 everywhere.** The typological map adds doctrinal weight to every piece's scoring context.

**The approach trades Alma 32 improvement for universal inflation.** It's the same problem as gospel-vocab context injection for talks — when you tell the model to look harder for something, it finds it everywhere.

### When Two-Stage Could Work

- Selective deployment on pieces where surface teach is obviously wrong (e.g., known typological content like Alma 32, Isaiah chapters)
- With a better Stage 1 that outputs "no significant typological connections" for most pieces rather than finding connections everywhere
- With a scoring model that can weight Stage 1 confidence rather than treating every connection as significant

### Cost

2× inference per piece ($0.002 → $0.004 for local inference, or 2× the time). Not worth it for 1.54 MAE when single-pass achieves 1.32.

---

## Known Structural Limitations

These are problems that prompt engineering cannot fix. They require a different model, fine-tuning, or rubric changes.

### 1. Alma 32 Love (GT=7.5, best achieved=5)

The love dimension requires recognizing that the tree of life = love of God (1 Ne 11:22) and that Alma's seed parable IS the tree of life vision in instructional form. This is a multi-hop typological inference:
- seed → tree → fruit (Alma 32:28-42)
- tree of life → love of God (1 Ne 11:22-25)
- therefore: Alma's teaching about faith IS teaching about the love of God

No model has scored love above 5 on Alma 32 across any experiment.

### 2. 3 Nephi 17 Help Deflation (GT=9, calibrated=5)

The calibrated prompt's distribution warning ("most talks score 3-5") causes the model to suppress legitimate 9-level scores. 3 Ne 17 has Christ literally present, ministering one by one, healing, praying with language that "cannot be written" — this IS a 9 on help. But the model reads the calibration cues and pushes it down.

### 3. Spirit "About vs By" Overcorrection (Holland: GT=7, calibrated=3)

The spirit ABOUT/BY distinction works on Alma 32 (9→5, closer to GT=3.5) but overcorrects on Holland, where personal testimony IS the Spirit working. Holland's talk doesn't "create space" in the traditional sense — it IS the experience of the Spirit bearing witness through testimony. The model learned the rule but not the exception.

### 4. Teach Floor at 5

Calibration pulled teach from a floor of 7-8 to a floor of 5. Better, but still can't differentiate 5-level teach (Bednar: doctrine-focused, Christ referenced but not central) from 7.5-level teach (Alma 32: Christ IS the deep architecture but invisible on the surface).

### 5. Reasoning-Score Disconnect

The model's qualitative reasoning often correctly identifies what should push a score down, then scores high anyway. DC 121 reasoning says "declares rather than models" and "suffering of Joseph Smith is implied but not deeply explored" — then still scores teach=8 and spirit=7 (baseline). The reasoning knows what a 5 looks like but the scoring defaults up.

---

## Hardware & Infrastructure

| Component | Spec |
|-----------|------|
| GPUs | 2× RTX 4090 (24GB each, 48GB total) |
| Current usage | 1 model loaded at a time (~9 GB for ministral) |
| Inference server | LM Studio (localhost:1234, OpenAI-compatible API) |
| Remote machine | Available for embeddings (separate from inference) |

### Planned: Parallel Inference

With 2× 4090, can run 2 model instances simultaneously (each on one GPU). With the remote machine running embeddings, 6-8× throughput is achievable for batch scoring. The `--concurrency` flag in the gospel-engine proposal supports this.

### Batch Timing Estimates

| Configuration | Time per talk | 5,500 talks |
|--------------|--------------|-------------|
| Current (sequential, 1 GPU) | ~18-20s | ~28 hours |
| 2× concurrent (2 GPU) | ~10s effective | ~15 hours |
| 4× concurrent (2 GPU + remote) | ~5s effective | ~8 hours |

---

## Phase 0 Context Experiments (on nemotron-3-nano)

Before switching to ministral, extensive context experiments ran on nemotron. Key results:

| Experiment | Context | MAE | Key Finding |
|------------|---------|-----|-------------|
| T0 | None | 2.61 | Massive inflation everywhere |
| T1 | titsw-framework.md | 2.06 | Framework helps modestly |
| T2 | gospel-vocab.md | 2.39 | **Gospel-vocab inflates love catastrophically for talks** |
| T3 | Talk rhetorical context | 2.00 | Talk-specific context helps |
| **T4** | **Calibration example** | **1.83** | **Best — one-shot calibration anchoring** |
| T5 | T3+T4 combined | 2.11 | Combining contexts adds noise |
| T6 | Same-speaker calibration | ~0.00* | Proved same-speaker anchoring (Bednar calibration → Bednar copies itself) |

\*T6 was a diagnostic experiment, not a real scoring approach.

**Key finding:** Calibration context (a scored example) was the most effective context type. Gospel-vocab was counterproductive for talks. These findings carried forward into the ministral experiments.

Phase 0 analysis: `results/phase0-analysis.md`

---

## Experiment Infrastructure

### Files

| File | Purpose |
|------|---------|
| `run-test.ps1` | Single test: prompt + content → model → JSON response |
| `run-suite.ps1` | Full suite: all prompts × model → results.tsv |
| `run-twostage.ps1` | Two-stage wrapper: Stage 1 typology → Stage 2 scoring |
| `scoring/main.go` | Go CLI for ground truth, import, stats, comparison (SQLite backend) |
| `results.tsv` | Master results log (all experiments) |
| `results/` | Raw JSON responses per experiment |
| `prompts/` | All prompt versions (21 files from titsw.md through titsw-stage2-score.md) |
| `references/` | Ground truth, model profile |
| `context-t1/` through `context-t5/` | Phase 0 context packages |

### Scoring CLI

```powershell
cd experiments/lm-studio/scripts
go run ./scoring init           # Create DB + seed ground truth
go run ./scoring import <tag>   # Import results from results/ dir
go run ./scoring stats <tag>    # MAE per piece + overall
go run ./scoring compare tag1,tag2,tag3  # Side-by-side comparison
go run ./scoring gt list        # Show all ground truth scores
```

---

## Future Improvement Paths

### 1. Better Base Model

Wait for a model that natively handles:
- Multi-hop typological reasoning (seed → tree → fruit → love of God)
- Spirit detection (testimony vs description of spiritual experiences)
- Proper calibration without explicit anchors

The field moves fast. A 14B reasoning model from Q4 2026 may handle what we're manually calibrating around.

### 2. Fine-Tuning

Training data exists: 6 ground-truth pieces × 6 dimensions = 36 labeled data points, plus qualitative reasoning from all prompt variants. Could fine-tune on H100s (cloud) or accumulate more ground truth first (13 pieces exist, only 6 used for formal experiments).

**Minimum viable dataset:** ~50-100 fully scored pieces with reasoning chains. Current 13 pieces aren't enough.

### 3. Rubric Improvements

- **Love dimension:** Account for typological love (tree of life = love of God = Christ's atonement made personal). Current rubric only recognizes pastoral/demonstrated love.
- **Invite dimension:** Recognize structured implicit invitations (Alma 32's escalating "exercise... desire... nourish" pattern) not just explicit "I invite you to..."
- **Spirit dimension:** Nuance the ABOUT/BY distinction for testimony-driven content where personal witness IS the Spirit working.

### 4. Selective Two-Stage

Deploy two-stage only on content flagged as "likely typological" (e.g., Old Testament chapters, Alma 32 class content, Isaiah). Use single-pass for everything else. Requires a lightweight classifier to flag typological content.

### 5. Ensemble Approach

Run 2-3 prompts per piece, take the median score per dimension. Reduces variance but 2-3× cost. Could use the existing calibrated + anchored prompts as the ensemble set.

### 6. Context Pre-computation

For the gospel-engine indexer: pre-compute cross-references per chapter, inject only the most relevant 3-4 cross-references as context. This is what context injection proved could work — the question is whether it's worth the complexity at scale.

---

## Decisions Made

| Decision | Rationale | Date |
|----------|-----------|------|
| ministral-3-14b-reasoning as production model | Best MAE (1.32), smallest size (9 GB), responsive to prompt tuning | Mar 29 |
| titsw-calibrated as production prompt | 14% improvement over baseline, per-dimension anchors + spirit distinction | Mar 29 |
| T=0.2 as production temperature | Safest (no catastrophic errors), second-best MAE (1.32 vs 1.31 at T=0.4) | Mar 29 |
| Single-pass over two-stage for production | Two-stage MAE 1.54 vs single-pass 1.32. Two-stage helps Alma 32 but hurts everything else. | Mar 29 |
| Stop prompt tuning | Remaining errors are structural (model/rubric limitations), not prompt-solvable. MAE 1.32 is good enough for indexing. | Mar 29 |
| Vocabulary approach for talks, lens approach for scripture | Context (gospel-vocab) helps scripture but inflates talk scores. Talks need scoring vocabulary, not theological context. | Mar 28-29 |

---

## Settled Questions

These are closed — don't re-investigate without new evidence.

1. **Thinking mode is harmful on nemotron-3-nano.** Think tokens burn context without improving scores.
2. **Gospel-vocab context inflates love on talks.** +5 above ground truth. Never use for talk scoring.
3. **Combining multiple context types doesn't help.** T5 (T3+T4) was worse than T4 alone. More context = more noise.
4. **Same-speaker calibration causes copying.** T6 proved Bednar calibration → Bednar scores are duplicated. Use different-speaker calibration examples.
5. **The mixed-context contamination in v5.1.** The famous MAE 0.93 was inflated by ~18KB of TITSW framework context accidentally loaded via run-test.ps1. Real v5.1 no-context baseline is ~1.3.
6. **Light prompt nudges don't work.** titsw-deflate (3 minimal rules) made things WORSE (+11%). The model needs structured guidance, not hints.
7. **v5.4 three-axis design is right for indexing.** Despite MAE 1.30, the modes/categories/insights output is what the downstream gospel-engine actually needs. MAE is a sanity check, not the product.

---

## Open Questions

For picking up later if this work resumes.

1. Would a hybrid prompt (calibrated anchors + softer distribution warning for pieces that genuinely deserve multiple 9s) fix the 3 Ne 17 problem?
2. Could a lightweight typological classifier flag content for selective two-stage scoring?
3. At what ground-truth dataset size does fine-tuning become viable?
4. Does ministral perform differently at 131k context (vs tested 65k)?
5. Would an ensemble of calibrated + anchored prompts (median scores) beat either alone?
