---
workstream: WS1
status: proposed
brain_project: 6
created: 2026-03-21
last_updated: 2026-04-21
---

# Classification Quality Benchmark

## Binding Problem

Brain classification produces entries that feel like "slop that I cannot act upon." When Michael throws ideas, URLs, and quick captures into the brain, the local 14B model confidently categorizes them — but often into the wrong bucket, with uniformly high confidence, lossy titles, and no actionable next steps. We need data on whether smarter (and free/cheap) API models would do meaningfully better with the same prompt, or whether the prompt/abstraction itself is the bottleneck.

## Success Criteria

1. A CLI tool (`classify-bench`) that classifies a fixed test dataset through 6 models and produces a side-by-side comparison table
2. Each model's results visible as: category, confidence, title, tags, and any fields — for the same raw input text
3. A human-readable markdown report showing where models agree and disagree
4. Clear recommendation: which model(s) to use, and whether prompt/category changes are also needed

## Constraints

- Must reuse the existing `Completer` interface and classifier prompt from brain.exe
- LM Studio models use the `LMStudioClient`; Copilot models use the SDK `Client`
- No changes to brain.exe itself — this is a standalone benchmark tool
- Test data comes from real brain entries (not synthetic)
- SDK model name strings need to be discovered/tested (may vary from docs display names)

## Models Under Test

| # | Model | Backend | Cost | Notes |
|---|-------|---------|------|-------|
| 1 | mistralai/ministral-3-14b-reasoning | LM Studio | Free (local) | Current classifier |
| 2 | qwen/qwen3.5-9b | LM Studio | Free (local) | Previous classifier (may need reloading) |
| 3 | Claude Haiku 4.5 | Copilot SDK | 0.33x | Fast, cheap Anthropic |
| 4 | GPT-5.4 mini | Copilot SDK | 0.33x | Latest OpenAI mini |
| 5 | GPT-5 mini | Copilot SDK | 0x | FREE included model |
| 6 | Raptor mini | Copilot SDK | 0x | FREE included model |

**Why these 6:** Models 1-2 are the existing local options. Models 3-4 are cheap API (0.33x). Models 5-6 are free API — if they outperform local, the cost argument for local inference vanishes.

## Architecture

```
scripts/brain/cmd/classify-bench/
├── main.go              # CLI entry point
└── testdata.json        # Extracted test entries (id, raw_text, current_category)
```

**Within the brain Go module** — shares the `internal/ai`, `internal/classifier` packages. One new `cmd/` entry point.

### Flow

1. Load test entries from `testdata.json`
2. For each backend (LM Studio, Copilot SDK):
   a. Initialize client
   b. For each model on that backend:
      - Send each test entry through classification
      - Capture response: category, confidence, title, tags, fields, latency
3. Write results to `results/classify-bench-{timestamp}.md`
4. Print summary table to stdout

### Copilot SDK Usage (Non-Session Mode)

The SDK requires sessions, but we minimize overhead:

```go
cfg := &copilot.SessionConfig{
    Model: modelName,
    SystemMessage: &copilot.SystemMessageConfig{
        Mode:    "replace",
        Content: classifySystemPrompt,
    },
    OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
    InfiniteSessions:    &copilot.InfiniteSessionConfig{Enabled: copilot.Bool(false)},
    AvailableTools:      []string{},  // No tools needed
}
session, _ := client.CreateSession(ctx, cfg)
defer session.Destroy()
```

One session per model, reused across all test entries (same system message). Each entry is a `SendAndWait` call. Session destroyed after all entries for that model complete.

### LM Studio Usage

Direct HTTP calls via the existing `LMStudioClient`. Uses `CompleteStructuredJSON` when the profile supports it, falls back to `CompleteJSON`.

## Test Dataset

~15 entries from the real brain, covering:

| Type | Example |
|------|---------|
| Clear action | "Grocery shopping list" with items |
| Scripture study | "Study the Only Begotten references" |
| Ambiguous URL capture | "Squad for ai agent flow, what can we learn? [URL]" |
| Bug report | "In brain app the scriptures dont show the body" |
| Person + context | "Bryce physical therapy. Research hippa compliant AI" |
| Pure idea | "Star Trek UI with Pretext [URL]" |
| Mixed intent | "Claude scientific research [URL]" |
| Project-like idea | "Agent Sandbox for Brain Sandbox Foundation [URL]" |

Entries extracted from the live brain via API, saved as JSON with `id`, `original_body`, and `current_category` (for reference, not as ground truth — current classifications may be wrong).

## Evaluation Criteria

For each entry across all models:
1. **Category correctness** — Does this match what Michael would choose?
2. **Confidence calibration** — Is the model uncertain when it should be?
3. **Title quality** — Does the title capture the essence without losing key info (URLs, questions)?
4. **Actionability** — Can Michael look at the classified result and know what to do?
5. **Latency** — How long per classification?

Michael will do the final human evaluation. The tool provides the data; he provides the judgment.

## Phased Delivery

### Phase 1: Build the tool + extract test data (this session)
- Extract 15 entries from brain API into `testdata.json`
- Build `classify-bench` CLI with LM Studio backend support
- Run against ministral-3-14b and (if available) qwen3.5-9b
- Verify output format

### Phase 2: Add Copilot SDK backend (this session if time permits)
- Add SDK client initialization
- Discover correct model name strings
- Run Haiku, GPT-5.4 mini, GPT-5 mini, Raptor mini
- Generate comparison report

### Phase 3: Analysis & decision (after Michael reviews results)
- Michael evaluates which model(s) produce the best results
- Decision: keep local, switch to API, or hybrid
- If switching: update brain.exe classifier to use the winning model
- If prompt issues: propose category/prompt changes as separate work

## Costs and Risks

| Risk | Mitigation |
|------|------------|
| SDK model names wrong | Test one first, iterate |
| Copilot CLI not authenticated | Use existing `copilot` CLI auth from brain.exe |
| Qwen model unloaded in LM Studio | Skip it or reload manually |
| API latency makes classification too slow for real-time | Measure it — async classification already exists |
| Results are all similar | Still valuable data — tells us the prompt is the bottleneck, not the model |

## Recommendation

**Build it.** This is a bounded experiment with clear deliverables. Worst case: we learn all models are similar and the problem is the prompt/categories. Best case: a free API model (GPT-5 mini or Raptor mini) dramatically outperforms local 14B inference and we can simplify brain.exe by removing the LM Studio dependency for classification.
