# LM Studio Model Experiment Harness

Scripts for systematically testing local LLM models against gospel study tasks.

Inspired by [autoresearch](https://github.com/karpathy/autoresearch) — same loop pattern:
run a test, record results, compare, iterate. But instead of optimizing val_bpb,
we're optimizing *usefulness for scripture study*.

## How It Works

```
run-test.ps1       — Single test: prompt + content → model → recorded response
run-suite.ps1      — Full suite: all prompts × specified models → results.tsv
context.md         — System context (covenant + intent extract) sent to every model
prompts/           — Test prompts (one per file)
content/           — Test content (talks, chapters) fed to prompts
results/           — Raw model responses (JSON)
results.tsv        — Structured results log
```

### Quick Start

1. Load a model in LM Studio (ensure it's serving at `localhost:1234`)
2. Run a single test:
   ```powershell
   .\run-test.ps1 -Prompt summarize -Content kearon-receive-his-gift -Model nemotron-3-nano
   ```
3. Run the full suite against one model:
   ```powershell
   .\run-suite.ps1 -Model nemotron-3-nano
   ```
4. Run all prompts, score manually, then try the next model.

### The Two-Pass Pattern

**Pass 1 — Standard prompt:** Same prompt to every model. Apples-to-apples comparison.

**Pass 2 — Tailored prompt:** After seeing how each model responds, tailor the prompt
to play to that model's strengths. We can burn more time iterating prompts than
Michael can — this is where the autoresearch spirit applies.

### System Context

Every test includes `context.md` as the system message. This contains the covenant
and intent extract — the same framing our agents get. We're testing whether models
can work *within our framework*, not just generate generic text.

### Results

`results.tsv` is the master log (tab-separated, never committed with sensitive content):

```
timestamp	model	prompt	content	tokens_in	tokens_out	tok_per_sec	latency_ms	ttft_ms	score	notes
```

Human scoring (Michael) goes in the `score` column (0-5). Raw responses are saved
as JSON in `results/` for review.

### Models

| Model | LM Studio ID | Notes |
|-------|-------------|-------|
| nemotron-3-nano | nvidia/nemotron-3-nano | 1M context, 160+ tok/s |
| qwen3.5-35b | qwen/qwen3.5-35b-a3b | 262k, ecosystem baseline |
| lfm2-24b | liquid/lfm2-24b-a2b | 32k, RAG-optimized |
| glm-4.7-flash | zai-org/glm-4.7-flash | Reasoning mode |
| devstral-small-2 | mistralai/devstral-small-2-2512 | Dense, 256k, vision |
