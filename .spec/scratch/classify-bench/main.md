# Classification Quality Benchmark — Scratch File

**Binding problem:** Brain classification produces output that feels like "slop that I cannot act upon." Is the problem the model, the prompt, or the abstraction itself?

---

## Phase 2: Research Inventory

### Current Architecture

- **Classifier:** `scripts/brain/internal/classifier/classifier.go`
- **Profiles:** `scripts/brain/internal/classifier/profiles.go` — 4 registered models
- **AI backends:** Two `Completer` implementations:
  - `LMStudioClient` — local OpenAI-compatible API (http://localhost:1234/v1)
  - `Client` (Copilot SDK) — session-based, uses `CreateSession` + `SendAndWait`
- **Current model:** mistralai/ministral-3-14b-reasoning (local, LM Studio)
- **Previous model:** qwen/qwen3.5-9b (local, unloaded/broken)

### Classification System Prompt (default)

6 categories: people, projects, ideas, actions, study, journal

Key issues with category definitions:
- **`study`** = "Scripture insight, spiritual impression, gospel learning, covenant commitment" — but the word "study" in general text triggers this even for tech learning
- **`projects`** = "Active work with a status and next action" — but passive "look at this repo" captures get classified here
- **`ideas`** vs **`projects`** boundary is unclear for URL captures

### Quality Audit Results (April 2, 2026)

64 entries total. Classified all 18 inbox entries with ministral-3-14b-reasoning.

#### Misclassifications Found

| Entry | Got | Should Be | Raw Text |
|-------|-----|-----------|----------|
| SQUAD AI Agent Flow | study (0.9) | ideas | "Squad for ai agent flow, what can we learn? [URL]" |
| AI Skills and Career | study (0.7) | ideas | "Ai jobs and skills, where do I stand? [YouTube URL]" |
| Scriptures not showing | study (0.9) | projects (bug) | "In brain app the scriptures dont show the body of the memorize items" |
| AI role in content | people (0.9) | ideas/journal | "Im pretty sure AI opus/sonnet wrote the article... [URL]" |
| Claude AI Research | projects (0.9) | ideas | "Claude scientific research [URL]" |
| Study Material Digestion | projects (0.9) | ideas/projects | "See if there are other ways of digesting our study materials..." |
| GitHub Copilot Skills Examples | ideas (0.95) | ideas ✓ but title lossy | Original was about examining Copilot skills for brain integration |

#### Patterns

1. **"study" attracts tech learning.** Any input mentioning learning, research, or study gets pulled to `study` even when it's about AI/tech, not scripture.
2. **URLs → projects.** "Look at this cool repo" becomes a project because the model infers active work from a URL.
3. **Confidence is uniformly high.** Most entries get 0.85-0.95 regardless of ambiguity. The model is confidently wrong.
4. **No personal context.** The model doesn't know Michael's active projects, so it can't distinguish "related to something I'm building" from "random interesting link."
5. **Title generation loses intent.** "Squad for ai agent flow, what can we learn?" → "Learning from SQUAD AI Agent Flow" — the question/exploration intent is gone.

### Root Cause Analysis

Three layers of problem:

**Layer 1: Category boundary confusion (prompt issue)**
- `study` is defined for spiritual study but the word is generic
- `ideas` vs `projects` is ambiguous for "check out this link" captures
- Fix: Tighten category definitions, add examples, add personal context

**Layer 2: Model capability (model issue)**  
- 14B local model can categorize text but can't reason about intent
- Can't read URLs, can't connect to active work context
- Higher-capability models (Haiku, GPT-5 mini) might do better with same prompt

**Layer 3: Classification vs. triage (abstraction issue)**
- Classification answers "what IS this?" but Michael needs "what should I DO with this?"
- A smarter model could generate actionable next steps, connect to active projects
- This is the deepest issue but also the biggest change

### Model Comparison Targets

From GitHub Copilot pricing docs (April 2026):

| Model | Type | Multiplier | Notes |
|-------|------|------------|-------|
| ministral-3-14b-reasoning | Local (LM Studio) | Free (hardware) | Current classifier |
| qwen3.5-9b | Local (LM Studio) | Free (hardware) | Previous classifier (currently unloaded) |
| Claude Haiku 4.5 | Copilot SDK | 0.33x | Cheap, fast, Anthropic |
| GPT-5.4 mini | Copilot SDK | 0.33x | Cheap, latest OpenAI mini |
| GPT-5 mini | Copilot SDK | 0x (FREE) | Included model! |
| Raptor mini | Copilot SDK | 0x (FREE) | Included model! |

**Key insight:** GPT-5 mini and Raptor mini are **0x multiplier** on paid plans — they cost nothing. If they classify better than the local 14B model, there's no reason not to use them (assuming API latency is acceptable).

### Copilot SDK "Non-Session" Mode

**What Michael meant:** Lightweight SDK usage without the full agentic session (tools, infinite sessions, workspace persistence). Just system prompt → user input → JSON response.

**How to implement:**
```go
session, _ := client.CreateSession(ctx, &copilot.SessionConfig{
    Model: "gpt-5-mini",  // or claude-3.5-haiku, etc.
    SystemMessage: &copilot.SystemMessageConfig{
        Mode:    "replace",
        Content: classifyPrompt,
    },
    OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
    InfiniteSessions: &copilot.InfiniteSessionConfig{
        Enabled: copilot.Bool(false), // No persistence needed
    },
    AvailableTools: []string{}, // No tools
})
response, _ := session.SendAndWait(ctx, copilot.MessageOptions{
    Prompt: rawText,
})
session.Destroy() // Single-use
```

This is essentially using the Copilot SDK as a stateless chat completion API. The session overhead is minimal.

**Model name format:** SDK uses string model IDs like "gpt-4.1", "gpt-4". Need to discover exact strings for newer models. Likely: "gpt-5-mini", "gpt-5.4-mini", "claude-3.5-haiku", "raptor-mini". May need to test.

### Existing Brain Architecture for Benchmarking

The `Completer` interface (`completer.go`) already abstracts both backends:
```go
type Completer interface {
    CompleteJSON(ctx context.Context, messages []ChatMessage, temperature float64) ([]byte, error)
    CompleteStructuredJSON(ctx context.Context, messages []ChatMessage, temperature float64, schema map[string]any) ([]byte, error)
    Model() string
    SetModel(model string)
}
```

Both `LMStudioClient` and `Client` (Copilot SDK) implement this. The benchmark can use both backend types through this interface.

### Test Data

Using entries from the brain with known original_body text. These are real captures from Michael's daily use, not synthetic — perfect for evaluating real-world classification quality.

Sample of 10-15 entries covering:
- Clear actions (grocery lists)
- Clear study items (scripture references)
- Ambiguous URLs (GitHub repos, YouTube videos)
- People mentions
- Bug reports
- Mixed-intent captures

---

## Phase 3a: Critical Analysis

### Is this the RIGHT thing to build?
**Yes.** Michael is experiencing low-quality classification output. The current model is a local 14B — comparing it to API models that are:
- (a) free or near-free on Copilot
- (b) potentially much higher quality for JSON classification
This is a data-driven decision about whether to keep local inference or switch to API.

### Does this solve the binding problem?
**Partially.** Better models may fix Layer 1 and Layer 2 (wrong categories, low reasoning). But Layer 3 (classification vs. triage) is a separate proposal. The benchmark will reveal how much of the problem is the model vs. the prompt vs. the abstraction.

### Simplest useful version?
A CLI tool that runs a fixed set of entries through each model, produces a comparison table, and writes results to a file. No UI. No persistent infra. Just a one-shot experiment.

### What gets WORSE?
- API dependency: classified entries now need internet (vs. pure local)
- Latency: API calls take longer than local 14B inference
- Premium request spend: Haiku and GPT-5.4 mini cost 0.33x each. But GPT-5 mini and Raptor mini are free.
- Session management complexity in brain.exe if we switch backends

### Mosiah 4:27 check?
This is a bounded experiment (one session to build, one session to analyze). Not a new open-ended project. The tool can inform whether to restructure the whole classification system or just swap the model.
