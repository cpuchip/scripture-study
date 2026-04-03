# Embedding 4B vs 8B Comparison — Scratch File

## Binding Problem

Does Qwen3-Embedding-8B produce meaningfully better search results than 4B on our scripture data? Michael wants to potentially remove a GPU. Need data, not benchmarks.

---

## Research Findings

### MTEB Benchmark Numbers (from lm-studio-model-experiments scratch)
- **4B multilingual:** 69.45 | **English:** 74.60
- **8B multilingual:** 70.58 | **English:** 75.22
- Delta: ~1.6% multilingual, ~0.8% English
- Both models support instruct-aware embeddings
- Both support MRL (configurable dimensions)

### Dimension Specs
- **4B:** Native up to 2560 dimensions
- **8B:** Native up to 4096 dimensions
- Current gospel-engine production: using 8B (default dims = likely 4096)
- Current gospel-vec: using 4B (default dims = likely 2560)
- Dimension difference alone = 1.6x storage cost

### VRAM Usage
- 4B Q4_K_M: ~2.5 GB
- 8B Q4_K_M: ~4.68 GB
- 8B Q8_0: ~8.05 GB
- Delta: 2-5.5 GB depending on quantization
- Michael has dual 4090s but wants to potentially remove one

### Current Architecture
- gospel-engine: 8B model, `text-embedding-qwen3-embedding-8b`
- gospel-vec (legacy): 4B model, `text-embedding-qwen3-embedding-4b`
- LM Studio serves at `http://localhost:1234/v1/embeddings`
- Can only load one model at a time per port
- Embedder: simple HTTP POST wrapper in `internal/vec/embedder.go`
- Storage: .vecf mmap format (header has dimension + count)

### Test Corpus: 1 Nephi
- 22 chapters in `gospel-library/eng/scriptures/bofm/1-ne/`
- All 1,584 scripture chapters already enriched in gospel.db
- Enrichment content: summary (~75-100 words), keywords, key_verse, christ_types, connections
- Embedded text = `{ref}: {summary}\nKeywords: {keywords}\nChrist types: {types}`
- ~600+ individual verses available for verse-level testing

### Prior Art
- `scripts/chromem-exp/` — basic embedding experiments with chromem-go
- Has a `compare` experiment mode already
- Uses chromem-go directly, not vecf
- Model experiments scratch file has the benchmark research

### Known Issues
- LM Studio dimensions parameter may not work correctly (issue #101)
- 8B model was initially misclassified as inference model (bug #965)
- Both models confirmed working with instruct-aware prefixes

---

## Critical Analysis

### Is this the right thing to build?
YES. Hardware decision blocking on data. Small tool, real question.

### Simple enough?
YES. Standalone CLI, JSON storage, markdown report. No infrastructure.

### What gets worse?
Nothing. Disposable test tool. If anything, it gives us a reusable pattern for future model comparisons.

### Mosiah 4:27?
This is tiny — one session to build, 20 min to run. Low cost.

### Does this duplicate?
chromem-exp exists but doesn't do A/B comparison with metrics. This is a focused evolution.

### MTEB says 4B is probably fine. Why not just trust it?
Because MTEB tests generic English. Our data is scripture — specialized vocabulary, theological concepts, King James English, enrichment summaries. The 0.8% gap could be larger or smaller on our data. Spending one session to get real data is worth it vs. a GPU decision based on "probably."

---

## Decision

**Proceed.** Write spec, build it.
