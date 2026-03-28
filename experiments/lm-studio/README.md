# LM Studio Model Experiments

Results from local LLM testing on dual RTX 4090s.

## Models Tested
- nvidia/nemotron-3-nano (1M context, Mamba2-Transformer MoE)
- qwen/qwen3.5-35b-a3b (262k context, MoE Transformer)
- liquid/lfm2-24b-a2b (32k context, Hybrid MoE — RAG-optimized)
- zai-org/glm-4.7-flash (128k+ context, MoE)
- mistralai/devstral-small-2-2512 (256k context, Dense Transformer)

## Embedding Models (separate track)
- Qwen3-Embedding-4B (baseline, instruct-aware, MRL 32-2560)
- Qwen3-Embedding-8B-GGUF (candidate, instruct-aware, MRL 32-4096)

## Quick Start

```powershell
cd experiments/lm-studio/scripts

# Run a single test
.\run-test.ps1 -Prompt summarize -Content kearon-receive-his-gift -Model nemotron-3-nano

# Run full suite against one model
.\run-suite.ps1 -Model nemotron-3-nano

# Review results
cat results.tsv
```

## Structure
```
scripts/
  run-test.ps1       — Single test: prompt + content → model → response
  run-suite.ps1      — Full suite: all prompts × model → results.tsv
  context.md         — System context (covenant + intent extract)
  prompts/           — Test prompt templates
  content/           — Test content files
  results/           — Raw JSON responses
  results.tsv        — Master results log
phase1-speed.md      — Load times, tok/s, VRAM at various context sizes
phase2-summarization.md — Quality comparison on gospel content
phase3-rag.md        — RAG pipeline reader comparison
phase4-embeddings.md — Embedding model upgrade evaluation (separate track)
phase5-decision.md   — Final recommendations for conference reindex
```

## Proposal
See [.spec/proposals/lm-studio-model-experiments/main.md](../../.spec/proposals/lm-studio-model-experiments/main.md)
