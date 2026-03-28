# LM Studio Model Experiments

Results from local LLM testing on dual RTX 4090s.

## Models Tested
- nvidia/nemotron-3-nano (1M context, Mamba2-Transformer MoE)
- qwen/qwen3.5-35b-a3b (262k context, MoE Transformer)
- liquid/lfm2-24b-a2b (32k context, Hybrid MoE — RAG-optimized)
- zai-org/glm-4.7-flash (128k+ context, MoE)
- mistralai/devstral-small-2-2512 (256k context, Dense Transformer)

## Embedding Models
- Qwen3-Embedding-4B (baseline)
- Qwen3-Embedding-8B-GGUF (candidate upgrade)

## Structure
- `phase1-speed.md` — Load times, tok/s, VRAM at various context sizes
- `phase2-summarization.md` — Quality comparison on gospel content
- `phase3-rag.md` — RAG pipeline reader comparison
- `phase4-embeddings.md` — Embedding model upgrade evaluation
- `phase5-decision.md` — Final recommendations for conference reindex
- `prompts/` — Standard test prompts

## Proposal
See [.spec/proposals/lm-studio-model-experiments/main.md](../../.spec/proposals/lm-studio-model-experiments/main.md)
