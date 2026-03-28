# LM Studio Model Experiments — Scratch File

*Research provenance for .spec/proposals/lm-studio-model-experiments/main.md*

---

## Binding Problem

We have dual RTX 4090s (48GB VRAM combined) and a growing library (~11,000+ files, ~1.3-1.5 GB of gospel content). We need to know which local LLM models are actually useful for:
1. Content digestion and summarization (conference reindex is imminent)
2. RAG pipeline quality (gospel-vec retrieval)
3. Embedding quality (upgrading from Qwen3-Embedding-4B baseline)

The dual 4090 setup changes the landscape — models that were too slow or too large now fit.

---

## Hardware Context

- **Machine:** New desktop (Mar 27 migration), 2x NVIDIA RTX 4090 (24GB each, 48GB total)
- **LM Studio:** Supports multi-GPU inference, OpenAI-compatible API at localhost:1234
- **Observed performance:**
  - Nemotron-3-Nano (30B, 3.5B active): 160+ tok/s at full 1M context
  - Qwen3.5-35B-A3B: 50 tok/s at 262k context

---

## Model Inventory (5 candidates)

### 1. nvidia/nemotron-3-nano
- **Architecture:** Hybrid Mamba2-Transformer MoE
- **Parameters:** 31.6B total, ~3.6B active per token
- **Context:** 1,048,576 tokens (1M)
- **GGUF size:** ~24.6 GB
- **Observed speed:** 160+ tok/s on dual 4090s
- **Strengths:** RULER benchmark 87.5% at 64K, 82.92% at 128K, 75.44% at 256K, 70.56% at 512K. Strong math (82.88% MATH), code (78.05% HumanEval). Tool use trained.
- **Weaknesses:** Mamba-Transformer hybrid less tested in production. Trails on vanilla MMLU.
- **License:** NVIDIA Open Model License (commercial OK)
- **Best for:** Processing entire books, full conference sessions, massive document synthesis

### 2. qwen/qwen3.5-35b-a3b
- **Architecture:** MoE Transformer
- **Parameters:** 35B total, ~3B active
- **Context:** 262,144 tokens (256k)
- **Observed speed:** ~50 tok/s on dual 4090s
- **Strengths:** Strong instruction following, multilingual. Qwen family is our existing baseline (gospel-vec uses Qwen3 embeddings).
- **Weaknesses:** Slower than nemotron. RULER accuracy drops at extended context per benchmarks.
- **License:** Apache 2.0
- **Best for:** High-quality summarization, instruction-following tasks, general-purpose

### 3. liquid/lfm2-24b-a2b
- **Architecture:** Hybrid (attention + convolutions) MoE
- **Parameters:** 24B total, ~2B active
- **Context:** 32,768 tokens (32k)
- **GGUF min memory:** 14 GB
- **Strengths:** Designed explicitly for edge/on-device. "Excels at agentic tool use, document summarization, Q&A, and local RAG pipelines." Only 2B active params = very fast. Fits on single 4090.
- **Weaknesses:** 32k context is small — can't fit a full conference or long book without chunking.
- **License:** Check (LiquidAI terms)
- **Best for:** RAG pipeline processing, per-document summarization, fast inference

### 4. zai-org/glm-4.7-flash
- **Architecture:** MoE (glm4_moe_lite)
- **Parameters:** 30B total, ~3B active
- **Context:** Listed as 128k on LM Studio (Michael reports 202,752 — may vary by quant/config)
- **GGUF min memory:** 16 GB
- **Strengths:** Tool use trained. Reasoning mode with enable_thinking flag. Strong coding benchmarks.
- **Weaknesses:** Newer model, less community testing.
- **License:** Open (Z.ai)
- **Best for:** Reasoning-heavy tasks, tool-use pipelines

### 5. mistralai/devstral-small-2-2512
- **Architecture:** Dense Transformer
- **Parameters:** 24B (dense — all active)
- **Context:** 256,000 tokens (256k)
- **GGUF min memory:** 16 GB
- **Strengths:** 68.0% SWE-Bench Verified. Vision support. Apache 2.0. Designed for agentic coding with tool use.
- **Weaknesses:** Dense architecture = more VRAM per token. Coding-focused, may not be optimized for summarization/analysis.
- **License:** Apache 2.0
- **Best for:** Code-aware tasks, agentic tool use, possibly document analysis with vision

---

## Embedding Model Research

### Qwen/Qwen3-Embedding-8B-GGUF
- **Type:** Text embedding (not inference)
- **Parameters:** 8B
- **Context:** 32k tokens
- **Max dimensions:** 4096 (configurable 32-4096 via MRL)
- **Quantizations:** Q4_K_M (4.68GB), Q5_0 (5.29GB), Q5_K_M (5.42GB), Q6_K (6.21GB), Q8_0 (8.05GB), F16 (15.1GB)
- **Instruct-aware:** YES — custom instructions improve performance 1-5%
- **MTEB scores:** #1 on MTEB multilingual leaderboard (70.58 as of June 2025)
- **vs current baseline:** Qwen3-Embedding-4B scores 69.45 MTEB multilingual; 8B scores 70.58. Marginal improvement on multilingual, bigger gap on English (74.60 vs 75.22).
- **LM Studio issue:** Bug #965 — model was misclassified as inference model. Needs to be loaded in embedding mode via /v1/embeddings endpoint.
- **Dimension note:** Issue #101 on GitHub — dimensions parameter may not work correctly in all configurations. Always generates 4096 regardless of requested dimension in some setups.

### dam2452/Qwen3-VL-Embedding-8B-GGUF
- **Type:** Vision-Language embedding
- **Parameters:** 8B
- **Context:** 32k tokens
- **Dimensions:** Up to 4096
- **Image support:** Yes (vision-language model)
- **LM Studio image embedding:** Not supported currently — LM Studio embedding API is text-only
- **Value:** Future potential for image-based scripture study (artwork, maps, diagrams)

### Current Baseline
- **Model:** text-embedding-qwen3-embedding-4b (4B)
- **Used by:** gospel-vec, brain.exe
- **Dimensions:** ~384 (likely, not documented)
- **Endpoint:** http://localhost:1234/v1

### Upgrade Path
Going from 4B to 8B embedding model:
- VRAM increase: ~4.7 GB (Q4_K_M) to ~8 GB (Q8_0) vs current ~2-3 GB
- Quality increase: marginal on MTEB but the instruct-aware feature is new and valuable
- Dimension increase: 384? → up to 4096 (major increase in vector expressiveness)
- Reindex required: YES — changing embedding model means full reindex of gospel-vec and brain

---

## LM Studio Embedding API Notes

- Endpoint: `POST http://localhost:1234/v1/embeddings`
- Model must be loaded in "embedding" mode (not inference)
- Only one embedding model OR multiple LLMs can be loaded at a time (usually)
- OpenAI-compatible API — works with our existing gospel-vec `NewLMStudioEmbedder()` wrapper
- Python SDK: `lmstudio.embedding_model("model-name").embed("text")`
- Dimension parameter support is model-dependent — test before relying on it

---

## Content Scale Analysis

| Content | Files | Est. Tokens | Fits in 32k? | Fits in 262k? | Fits in 1M? |
|---------|-------|-------------|---------------|----------------|-------------|
| Single scripture chapter | 1 | 1-10K | YES | YES | YES |
| Full Book of Mormon | ~239 | ~500K-1M | NO | Partial | YES |
| Single conference talk | 1 | 2-20K | YES | YES | YES |
| One full conference (50 talks) | ~50 | 250-500K | NO | Partial | YES |
| All conference talks | ~5,500 | 30-55M | NO | NO | NO |
| Lectures on Faith | 9 | 50-80K | Partial | YES | YES |
| All 44 topic studies | 44 | 200-500K | NO | Partial | YES |
| General Handbook | 39 ch | 500K-2M | NO | NO | Partial |

**Key insight:** Nemotron-3-nano's 1M context opens doors that were closed before. You can fit an entire conference, a full book of scripture, or all our study documents in a single prompt.

---

## Experiment Design Ideas

### Task 1: Summarization Quality
- Feed same document(s) to each model
- Compare summary quality, key point extraction, cross-reference identification
- Good test docs: a conference talk, a chapter of Lectures on Faith, one of our studies

### Task 2: Long-Context Faithfulness
- Feed progressively larger contexts and ask about details at various positions
- "Needle in a haystack" with gospel content — place a specific verse reference deep in context

### Task 3: Cross-Reference Discovery
- Feed a scripture chapter + surrounding context
- Ask each model to identify connections to other scriptures, conference talks, study themes
- Compare against known cross-references (footnotes provide ground truth)

### Task 4: RAG Pipeline Quality
- Use each model as the "reader" in a RAG pipeline
- Feed retrieved chunks from gospel-vec + a question
- Measure answer quality, citation accuracy, hallucination rate

### Task 5: Embedding Quality (separate experiment)
- Compare Qwen3-Embedding-4B (baseline) vs Qwen3-Embedding-8B
- Use existing gospel-vec benchmark framework
- Measure retrieval precision at different dimension settings

### Task 6: Speed Benchmarking
- Measure tok/s for each model at various context sizes
- Time to first token, throughput, total generation time
- Critical for understanding which model to use for batch processing (reindex) vs interactive use

---

## Prior Art in This Project

- `docs/model-experiments.md` — embedding/summary model experiment framework for gospel-vec
- `scripts/gospel-vec/experiments/` — benchmark queries, run script, experiment log
- `experiments/claude/`, `experiments/google/`, `experiments/openai/` — empty directory structure for cloud model experiments
- Active.md priority #3: "Model experiments — Run same prompts through Haiku/Sonnet/Opus, evaluate quality"
- Decisions.md: "Dual AI backend — LM Studio handles classification, Copilot SDK handles agent work"

---

## Key Questions

1. Should we run all 5 models on the same prompts, or select 2-3 for deeper testing?
2. What's the conference reindex timeline? Does this block on model selection?
3. Should embedding experiments (Qwen3-8B) be a separate spec or bundled here?
4. How do we handle the LM Studio limitation of one embedding model OR multiple LLMs at a time?
5. Should we build a simple harness script to automate prompt → model → score?
