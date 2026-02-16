# Model Experiments for gospel-vec

## Goal

Find the best combination of **embedding model** and **summary model + prompt** for gospel content retrieval before committing to a full 6-hour reindex.

## Current Baseline

| Component | Model | Notes |
|-----------|-------|-------|
| Embeddings | `text-embedding-qwen3-embedding-4b` | 4B param, via LM Studio |
| Summaries | auto-detected from LM Studio | Used for KEYWORDS/SUMMARY/KEY_VERSE format |
| Summary Prompt | v1 (see `summary.go`) | Temperature 0.2, 300 tokens max |

## Experiment Framework

### Quick Start

```powershell
# From repo root:
cd scripts/gospel-vec/experiments

# Run baseline experiment (Book of Mormon only, ~5-15 min)
.\run-experiment.ps1 -Name "baseline-qwen3-4b" -NoSummary

# Test a different embedding model (load it in LM Studio first)
.\run-experiment.ps1 -Name "nomic-embed" -EmbeddingModel "nomic-embed-text-v1.5" -NoSummary

# Test with summaries included
.\run-experiment.ps1 -Name "with-summaries" -Volumes "bofm" -SearchLayers "verse,paragraph,summary"

# Re-run queries on existing index with different search config
.\run-experiment.ps1 -Name "baseline-qwen3-4b" -SkipIndex -SearchLayers "verse,paragraph,summary" -Limit 20
```

### What It Does

1. **Builds gospel-vec** from current source
2. **Indexes a small benchmark** (default: Book of Mormon only) into an isolated data directory (`experiments/data-<name>/`)
3. **Runs 12 benchmark queries** from `benchmark-queries.json` against the index
4. **Scores retrieval quality** by comparing results to expected relevant passages
5. **Outputs results** to `experiments/results/<name>.json` and appends to `experiment-log.md`

### Isolation

Each experiment gets its own data directory via `GOSPEL_VEC_DATA_DIR` environment variable. Production data in `data/` is never touched.

## Experiments to Try

### Embedding Models

| Model | Size | Status | Notes |
|-------|------|--------|-------|
| `text-embedding-qwen3-embedding-4b` | 4B | Baseline | Current default |
| `nomic-embed-text-v1.5` | 137M | To test | Popular, much smaller |
| `mxbai-embed-large-v1` | 335M | To test | Good MTEB scores |
| `snowflake-arctic-embed-l-v2.0` | 335M | To test | Strong retrieval model |
| `bge-m3` | 568M | To test | Multilingual, dense+sparse |

### Summary Models

| Model | Size | Status | Notes |
|-------|------|--------|-------|
| Current auto-detect | varies | Baseline | Whatever is loaded in LM Studio |
| `qwen3-8b` | 8B | To test | Good instruction following |
| `llama-3.1-8b-instruct` | 8B | To test | Strong all-around |
| `gemma-3-4b-it` | 4B | To test | Smaller, might be faster |

### Prompt Experiments

- [ ] Adjust temperature (currently 0.2 — try 0.1 and 0.3)
- [ ] Modify KEYWORDS format (current: comma-separated → try structured)
- [ ] Add cross-reference hints to summary prompt
- [ ] Try different KEY_VERSE selection criteria
- [ ] Test DetectThemes prompt with different granularity

### Search Layer Combinations

- [ ] Verse only vs verse+paragraph
- [ ] With and without summary layer
- [ ] With and without theme layer
- [ ] All layers combined

## Metrics We Track

| Metric | Description |
|--------|-------------|
| **Average Recall** | Mean fraction of expected results found across all queries |
| **Perfect Recall** | Number of queries where 100% of expected results were found |
| **Zero Recall** | Number of queries where 0% of expected results were found |
| **Per-Category** | Breakdown by query type (doctrinal, narrative, prophecy) |

## Results Summary

_Run experiments and the results will appear in [experiment-log.md](experiment-log.md)._

_Detailed JSON results for each experiment are saved in `results/`._

## Notes

- Full reindex takes ~6 hours. Experiments use BofM only (~15 min without summaries, ~1 hour with)
- Load the model you want to test in LM Studio BEFORE running the experiment
- Embedding dimension must be consistent within an experiment — you can't mix models in one index
- The `PromptVersion` in `cache.go` (currently "v1") controls cache invalidation for summaries. Bump it when changing prompts.
