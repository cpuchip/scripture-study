# LM Studio Model Experiments — Local LLMs on Dual 4090s

## Binding Problem

We have a growing scripture library (~11,000+ files, ~1.3 GB of gospel content) and new hardware (dual RTX 4090s, 48GB combined VRAM). Conference is coming and a full content reindex is on the table. Before that reindex, we need to know which local models actually produce quality output for our specific use cases: summarization, cross-reference discovery, RAG retrieval, and embedding. We currently run Qwen3-Embedding-4B for embeddings and auto-detected models for summarization — both untested against alternatives on this hardware.

---

## Models Under Test

| # | Model | Params (Active) | Context | Speed (observed) | Architecture | License |
|---|-------|-----------------|---------|-------------------|--------------|---------|
| 1 | nvidia/nemotron-3-nano | 31.6B (3.6B) | 1,048,576 | 160+ tok/s | Mamba2-Transformer MoE | NVIDIA Open |
| 2 | qwen/qwen3.5-35b-a3b | 35B (3B) | 262,144 | ~50 tok/s | MoE Transformer | Apache 2.0 |
| 3 | liquid/lfm2-24b-a2b | 24B (2B) | 32,768 | TBD (expect fast) | Hybrid Attn+Conv MoE | LiquidAI |
| 4 | zai-org/glm-4.7-flash | 30B (3B) | 128k-202k | TBD | MoE (glm4_moe_lite) | Z.ai Open |
| 5 | mistralai/devstral-small-2-2512 | 24B (dense) | 256,000 | TBD | Dense Transformer | Apache 2.0 |

### Why these five?

- **Nemotron-3-nano:** The 1M context monster. Only model that can fit an entire conference or full Book of Mormon in a single prompt. Mamba-Transformer hybrid is architecturally novel — linear scaling on long sequences.
- **Qwen3.5-35b:** Already our ecosystem baseline (Qwen family). Strong instruction following. 262k is substantial.
- **LFM2-24b:** Authors explicitly designed it for "document summarization, Q&A, and local RAG pipelines." Only 2B active params means blazing fast. 32k context forces chunking — but that's how RAG works anyway.
- **GLM-4.7-flash:** Reasoning mode with thinking toggle. Strong coding benchmarks. MoE with tool-use training.
- **Devstral-small-2:** 68% SWE-Bench Verified. Dense architecture (all 24B active) trades speed for per-token quality. Vision support. Agentic coding focus.

### Embedding Model

| Model | Params | Dimensions | Context | Instruct-Aware | Notes |
|-------|--------|-----------|---------|---------------|-------|
| Qwen3-Embedding-4B (baseline) | 4B | ~384 (est.) | 32k | Yes | Current production model |
| Qwen3-Embedding-8B-GGUF | 8B | Up to 4096 | 32k | Yes | MTEB #1 multilingual (70.58) |

**Note:** Qwen3-VL-Embedding-8B (vision-language) is interesting for future but LM Studio doesn't support image embeddings yet. Park it.

---

## Experiment Phases

### Phase 1: Baseline Speed + Fit (1 session)

Confirm all 5 models load and run on dual 4090s. Measure:

| Metric | How |
|--------|-----|
| Load time | Time from model load request to ready |
| tok/s at small context (~4k) | Standard chat prompt |
| tok/s at medium context (~32k) | Feed a conference talk + question |
| tok/s at max context | Fill to model's limit, measure throughput |
| VRAM usage | Monitor at each context size |
| Time to first token | Responsiveness for interactive use |

**Output:** Speed table in `experiments/lm-studio/phase1-speed.md`

For each model, run the same baseline prompt:
```
Summarize this conference talk. Identify the main thesis, supporting scriptures, 
and one pattern the speaker uses that could improve my own teaching.
```

Feed it the same talk (pick one ~8,000 token talk). Grade on: accuracy, insight depth, citation correctness, usefulness.

### Phase 2: Summarization Quality (1-2 sessions)

Three test documents at increasing scale:

| Doc | ~Tokens | Fits in... |
|-----|---------|-----------|
| 1 conference talk | ~8K | All models |
| Lecture on Faith #3 | ~15K | All models |
| Full April 2025 conference (57 talks) | ~300-400K | nemotron, qwen3.5, devstral, glm |

**Prompts to test:**
1. **Summarize:** "Summarize this content. What are the 3 most important doctrinal points?"
2. **Cross-reference:** "What scriptures does this content reference or allude to? Include both explicit citations and implicit connections."
3. **Teaching extraction:** "What teaching patterns or rhetorical techniques does this content use that could be applied in a Sunday School lesson?"
4. **Needle retrieval:** Place a specific detail deep in the context. Ask about it. Test long-context faithfulness.

**Scoring rubric:**
- Accuracy (0-5): Did it get the facts right?
- Depth (0-5): Did it surface non-obvious insights?
- Citations (0-5): Did it cite real scriptures that actually support the point?
- Hallucination (0-5, inverse): Did it invent quotes or references?
- Usefulness (0-5): Would this help with actual study/teaching prep?

**Output:** `experiments/lm-studio/phase2-summarization.md`

### Phase 3: RAG Pipeline Test (1-2 sessions)

Test each model as the "reader" in our existing gospel-vec RAG pipeline:

1. Run 12 benchmark queries from `scripts/gospel-vec/experiments/benchmark-queries.json`
2. For each query, retrieve top-10 chunks from gospel-vec
3. Feed chunks + query to each model
4. Compare: answer quality, citation accuracy, hallucination rate
5. Compare against: using the model without RAG (pure context) where context allows

This tests model quality in the specific pipeline we actually use.

**Output:** `experiments/lm-studio/phase3-rag.md`

### Phase 4: Embedding Upgrade (1 session)

Using the existing gospel-vec experiment framework:

1. Run baseline with current Qwen3-Embedding-4B
2. Run same benchmark with Qwen3-Embedding-8B-GGUF (Q4_K_M quant, ~4.7 GB)
3. Test dimensions: 384, 768, 1024, 2048
4. Measure: retrieval precision, recall, speed, VRAM usage
5. Test instruct-aware feature: with and without task-specific instructions

**Note on dimension bug:** GitHub issue #101 reports dimensions parameter may always return 4096 regardless of request. Test this explicitly. If confirmed, we either use 4096 everywhere or stay at 4B.

**Output:** `experiments/lm-studio/phase4-embeddings.md`

### Phase 5: Conference Reindex Decision (< 1 session)

Review all results. Decide:
- Which model for gospel-vec summary generation? (currently auto-detect)
- Which embedding model? (currently Qwen3-Embedding-4B)
- At what dimensions?
- Which model for classification in brain.exe? (currently LM Studio auto-detect)
- Full reindex timeline and model selection

**Output:** `experiments/lm-studio/phase5-decision.md`

---

## Success Criteria

1. Speed measurements for all 5 models at multiple context sizes: documented
2. Summarization quality scores for at least 3 test documents across all models: documented  
3. At least one RAG pipeline comparison: documented
4. Embedding 4B vs 8B comparison with retrieval metrics: documented
5. Clear recommendation for conference reindex model selection: documented
6. All results under `experiments/lm-studio/` with reproducible methodology

---

## Constraints

- **LM Studio only** — not Ollama, not cloud APIs. This is local inference testing.
- **Dual 4090s** — all models must fit in 48GB VRAM (quantized OK)
- **One inference model + one embedding model at a time** — LM Studio can run both concurrently, but at max context lengths we should limit to this. No multi-LLM concurrent.
- **GGUF format** — all models must be available as GGUF for LM Studio
- **Existing benchmark framework** — use gospel-vec experiment infrastructure where possible (Phase 4)
- **Results go to `experiments/lm-studio/`** — not mixed with cloud model experiments

---

## Approach

### Test Harness — BUILT

The harness lives at `experiments/lm-studio/scripts/`. Inspired by [autoresearch](https://github.com/karpathy/autoresearch) — same loop: run a test, record results, compare, iterate.

**Architecture:**
```
scripts/
  run-test.ps1     — Single test: prompt + content → model → recorded response
  run-suite.ps1    — Full suite: all prompts × selected model → results.tsv
  context.md       — System context (covenant + intent extract) sent to every model
  prompts/         — Test prompt templates ({{CONTENT}} placeholder)
  content/         — Test content (talks, scripture chapters)
  results/         — Raw JSON responses
  results.tsv      — Master results log (tab-separated)
```

**System context:** Every test includes `context.md` as the system message. This is an extract of the covenant and intent — the same framing our agents get. We're testing whether models can work *within our framework*, not just generate generic text.

**Two-pass pattern:**
- **Pass 1 — Standard prompt:** Same prompt to every model. Apples-to-apples.
- **Pass 2 — Tailored prompt:** After seeing how each model responds, tailor prompts to each model's strengths. The agent can iterate prompts far longer than Michael can — this is the autoresearch spirit applied to prompt optimization.

### Prompt Templates

Stored in `experiments/lm-studio/scripts/prompts/`:
- `summarize.md` — summarization (thesis, scriptures, doctrinal points, teaching pattern)
- `cross-reference.md` — explicit citations, implicit allusions, unexpected connections
- `teaching.md` — rhetorical analysis for Sunday School prep
- `deep-study.md` — personal scripture study (key words, doctrine, becoming)
- `needle.md` — long-context faithfulness (specific detail retrieval)

### Test Content

Stored in `experiments/lm-studio/scripts/content/`:
- `kearon-receive-his-gift.md` — Elder Kearon, April 2025 (~13K chars, good teaching talk)
- `alma-32.md` — Alma 32, faith chapter (~10K chars, doctrinal content)

### Scoring

Human scoring (Michael) for quality metrics. Automated scoring for speed/fit. This is deliberate — we're testing whether the model output is *useful for scripture study*, not whether it passes a benchmark.

### Results Log

`results.tsv` (autoresearch-inspired structured logging):
```
timestamp  model  prompt  content  tag  tokens_in  tokens_out  tok_per_sec  latency_ms  score  notes
```
Score (0-5) and notes filled in by Michael after reviewing raw responses in `results/`.

---

## What This Is NOT

- Not a general LLM benchmark (we have enough of those)
- Not a cloud vs local comparison (that's the experiments/claude/ etc. folders)
- Not building a permanent evaluation framework (a test harness yes, a product no)

---

## Embedding Experiments — Separate Track

**Decision (Mar 28):** Embedding experiments (Phase 4) are split out from inference model testing. The inference model decision blocks conference reindex; the embedding upgrade can follow independently.

**Current leaning:** Adopt Qwen3-Embedding-8B-GGUF. It's the formally supported Qwen model (not a community fork like the VL variant). Stick with 4B or 8B — the VL model is user-supported and may not persist.

**Key question:** Is there a retrieval quality difference between 4B and 8B for our content? Both support instruct-aware mode and MRL (custom dimensions). The 4B supports dimensions 32-2560; the 8B supports 32-4096. Both support custom task instructions with 1-5% improvement.

**When:** After inference model selection. Use existing gospel-vec experiment framework (`scripts/gospel-vec/experiments/`).

---

## Costs and Risks

| Cost | Impact |
|------|--------|
| Time | 4-6 sessions across phases |
| VRAM | Dedicated GPU time during experiments (can't run other models) |
| Reindex risk | Choosing the wrong model means 6+ hour reindex wasted |
| Scope creep | "Just one more model" — stick to the five |
| Conference timing | Experiments should complete before April conference reindex |

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Need to pick the right local models before conference reindex |
| Covenant | Rules? | Reproducible methodology, honest scoring, documented results |
| Stewardship | Who owns? | Michael runs experiments, plan agent specs, results inform dev agent |
| Spiritual Creation | Spec precise enough? | Yes — 5 phases, clear prompts, defined metrics |
| Line upon Line | Phasing? | Phase 1 stands alone (speed). Each phase adds value independently |
| Physical Creation | Who executes? | Michael manually (model loading) + simple harness script |
| Review | How know it's right? | Results match subjective experience + benchmark scores |
| Atonement | If wrong? | Low cost — just wasted time. No production impact until reindex decision |
| Sabbath | When pause? | After Phase 5 decision — natural stopping point |
| Consecration | Who benefits? | Michael directly. Framework could help others running local models |
| Zion | Whole-system? | Feeds into gospel-vec reindex, brain.exe model selection, teaching prep |

---

## Recommendation

**Build.** Harness is built, prompts written, content staged. Ready to run.

**Early signal from Michael:** Nemotron-3-nano is the speed leader at 160+ tok/s, and qualitative testing ("teach me about Go concurrency") showed both nemotron and qwen3.5-35b performing well. Speed alone makes nemotron the front-runner for batch processing (reindex). Quality testing will confirm or challenge that.

**Phase 1 first.** Run the full suite against each model. Same prompts, same content. Get the apples-to-apples numbers. Then iterate prompts for any models that seem to under-perform on the standard prompts.

**Embedding: punt for now.** Lean toward 8B adoption, but test retrieval quality before committing. Both 4B and 8B support instruct-aware mode, so there's no urgency to upgrade for that feature alone.
