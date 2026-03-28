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

---

## Session 3 Triage — Idea Cascade (Mar 28)

Michael sent a burst of connected ideas pivoting from "which model" toward "what should the model produce." Here's the triage.

### Idea 1: LM Studio doesn't have filesystem access — feed content through API

**Verdict: Already handled.** The harness (`run-test.ps1`) reads content files locally and sends them to LM Studio's `/v1/chat/completions` endpoint. No change needed. Michael was clarifying scope, not requesting a change.

### Idea 2: LM Studio as mini-Copilot with MCP servers

**Verdict: Defer.** Interesting vision (LM Studio + function calling + gospel-mcp tools), but entirely separate from model selection. If a model proves it can do structured extraction well, this becomes a natural next step. Not now.

### Idea 3: Graph edges in SQLite (inspired by work's go.mod graph for 560 repos)

**Verdict: Already exists.** Gospel-mcp already has a `cross_references` table:

```sql
CREATE TABLE IF NOT EXISTS cross_references (
    source_volume, source_book, source_chapter, source_verse,
    target_volume, target_book, target_chapter, target_verse,
    reference_type  -- 'footnote', 'tg', 'bd', 'jst'
);
```

Indexed on both source and target. Populated by `extractCrossReferences()` in [scripture.go](../../scripts/gospel-mcp/internal/indexer/scripture.go) which parses footnote anchors (`<a id="fn-9a">`) and cross-reference links. Per-verse scoping was fixed Feb 15 (see [tool-use-observance.md](../../docs/06_tool-use-observance.md)).

**What it DOES have:**
- Footnote → scripture edges (all 5 standard works)
- TG, BD, GS edges (study aid references)
- Bidirectional indexes (can query "what points TO this verse" via idx_cross_ref_target)
- Already returned with every `gospel_get` and `gospel_search` result

**What it DOESN'T have:**
- Multi-hop traversal ("show me everything 2 hops from Alma 32:21")
- Conference talk → scripture edges (when Holland quotes Alma 7:12, that edge isn't stored)
- LLM-inferred thematic edges (implicit connections, not explicit footnotes)
- Study document → scripture edges
- Graph visualization

### Idea 4: Parse footnotes to build scripture relationship graph

**Verdict: Already done.** This IS Idea 3. The footnote parser exists. The graph exists. The question is what ELSE to add to the graph — and that's where LLM inference comes in (talk → scripture edges, thematic edges).

### Idea 5: "Gospel-comb" — unified vec + SQLite tool

**Verdict: Defer. Good idea, wrong time.**

Current architecture:
- **Gospel-mcp:** SQLite + FTS5 (keyword search, cross-references, structured data)
- **Gospel-vec:** chromem-go only (vector search, .gob.gz files, no SQLite)

Gospel-comb would combine: FTS5 (keyword) + vectors (semantic) + graph (cross-refs) in one queryable system. Three implementation options:
- A: Add SQLite to gospel-vec
- B: Add vectors to gospel-mcp
- C: New tool wrapping both

**Why defer:** Blocks on model selection. LLM-generated edges (talk→scripture, thematic connections) depend on which model produces them. Conference reindex is the forcing function and it hasn't happened yet. Also: 7 priorities already in active.md. Adding a new tool proposal makes 8.

**Revisit when:** Model experiments produce a clear winner AND conference reindex succeeds.

### Idea 6: Index with Teaching in the Savior's Way principles as dimensions

**Verdict: Add as prompt. Actionable NOW.**

The 4 TITSW principles map to analyzable dimensions:
1. **Love Those You Teach** — empathy, seeing divine potential, safety
2. **Teach by the Spirit** — spiritual preparation, responsiveness, testimony
3. **Teach the Doctrine** — scriptural depth, doctrinal clarity, personal relevance
4. **Invite Diligent Learning** — agency, participation, application

This is a prompt template question, not an architecture question. Add `prompts/titsw.md` to the harness as a 6th prompt that asks the model to analyze a talk/passage along these dimensions and return structured JSON with scores/tags. This directly tests whether models can produce structured teaching analysis — useful signal for model selection AND for eventual indexing.

Michael's overview study at [study/teaching-in-the-saviors-way/00_overview.md](../../study/teaching-in-the-saviors-way/00_overview.md) has the full breakdown.

### Triage Summary

| Idea | Verdict | Action |
|------|---------|--------|
| API-only scope | ✅ Already handled | None |
| LM Studio + MCP | ⏸️ Defer | Revisit after model selection |
| SQLite graph edges | ✅ Already exists | None — cross_references table |
| Parse footnotes | ✅ Already done | None — extractCrossReferences() |
| Gospel-comb unified search | ⏸️ Defer | Good idea, blocks on model selection |
| TITSW indexing dimensions | 🔨 Add prompt | Create prompts/titsw.md |

### Mosiah 4:27 Check

Michael has 7 priorities in active.md. The model experiments are #3. The harness is built. The next step is still: **load nemotron in LM Studio and run the suite.** These ideas are valuable future direction but none of them change what needs to happen next. The one actionable item (TITSW prompt) can be built in 5 minutes and doesn't change scope — it adds signal to the existing experiment.

### What the Graph IS Missing (Future Reference)

When it's time to build gospel-comb, the real graph extension opportunities are:
1. **Talk → scripture edges.** LLM reads each conference talk, extracts scripture references → stores as edges. This is the conference reindex output.
2. **Thematic edges.** LLM identifies that Moses 6:63 and Alma 30:44 both teach "all things testify of Christ" → stores as thematic connection.
3. **Study aid densification.** TG entries already point to verses, but the TG connections themselves could be traversed (A and B in same TG entry = related).
4. **Multi-hop queries.** "Show me the 2-hop neighborhood of Alma 32:21" — requires a simple BFS on the graph. SQLite can do this with recursive CTEs.

### Answers (Mar 28 — Michael's decisions)

1. **Both.** Run the same prompts through all 5 (pass 1, apples-to-apples). Then tailor prompts per model to see if targeted prompting gets better results (pass 2). The agent can burn more iteration time on prompt-tuning than Michael can — autoresearch spirit.
2. **Yes, blocks reindex.** Michael wants to see if one of these models is better before reindexing. Conference timing is the forcing function.
3. **Split out.** Embeddings become a separate track. Inference model decision is the blocker; embedding upgrade can follow independently. Michael leans toward Qwen3-Embedding-8B adoption since it's formally Qwen-supported (vs VL variant which is community/user-supported).
4. **1 embedding + 1 inference at a time.** LM Studio can run both concurrently. At max context lengths, we limit to this pair. No multi-LLM concurrent loading.
5. **Yes, built.** Harness at `experiments/lm-studio/scripts/`. PowerShell. `run-test.ps1` (single test), `run-suite.ps1` (full suite), `context.md` (system prompt from covenant/intent), prompt templates, content files, results.tsv log.

---

## Autoresearch Pattern (Mar 28)

Cloned [karpathy/autoresearch](https://github.com/karpathy/autoresearch) to `external_context/autoresearch/`. The pattern:
- `program.md` (human-edited context/instructions for the agent)
- `train.py` (the single file the agent modifies and iterates on)
- `results.tsv` (structured tab-separated results log)
- Autonomous loop: try → measure → keep/discard → iterate

Applied to our experiment:
- `context.md` = our `program.md` (covenant/intent extract as system prompt)
- Prompt files = our `train.py` (the thing being iterated between pass 1 and pass 2)
- `results.tsv` = same pattern (structured log, human scores added manually)
- Not fully autonomous (Michael scores quality) but the prompt-iteration can be agent-driven

---

## Qwen3-Embedding-4B Instruct Support (Mar 28)

**Confirmed:** Qwen3-Embedding-4B DOES support:
- **Instruct-aware mode:** YES. `Instruct: {task_description}\nQuery:{query}` format. Same as 8B.
- **MRL (custom dimensions):** YES, 32-2560 (vs 8B which goes to 4096).
- **MTEB multilingual:** 69.45 (vs 8B at 70.58). English: 74.60 vs 75.22.
- **Formally supported:** Yes, from Qwen directly. Apache 2.0.

So both models have instruct support. The 8B upgrade gives:
- Slightly better MTEB scores (~1 point multilingual, 0.6 English)
- Higher max dimensions (4096 vs 2560)
- 2x the parameters = more VRAM, slower

The instruct feature alone is NOT a reason to upgrade — we can use it with 4B right now by updating the gospel-vec embed.go to include task-specific prefixes. The dimension increase and marginal quality bump are the real differentiators for 8B.

---

## System Context Design (Mar 28)

Michael's insight: the project has built agent modes, skills, copilot instructions, covenants, and intent. Most of that is out of scope for testing local models — but **covenant, intent, and core instructions** should be part of the test. We're not testing bare models; we're testing whether they can work within our framework.

The system context (`context.md`) extracts:
- Core values from intent.yaml (depth over breadth, honest exploration, faith as framework)
- Key constraints (accurate quotes, specific scriptures, admit uncertainty)
- What good output looks like (cross-references, word analysis, teaching patterns, becoming)
- Standard works reference

This is ~250 tokens of system context — light enough that it doesn't eat into the content window, heavy enough that it provides framework.

---

## Early Quality Signal (Mar 28)

Michael reports: both nemotron-3-nano and qwen3.5-35b performed well on "teach me about Go concurrency" — a general knowledge task, not gospel-specific. This is encouraging but not conclusive for our use case. The harness tests will give gospel-specific signal.

Speed difference is dramatic: nemotron at 160+ tok/s vs qwen at ~50 tok/s. For batch processing (conference reindex with thousands of documents), 3x speed is the difference between a 2-hour reindex and a 6-hour one. Speed alone makes nemotron the front-runner for batch work.
