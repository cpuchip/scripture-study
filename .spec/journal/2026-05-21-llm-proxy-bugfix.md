---
date: 2026-05-21
session_window: morning (~08:30 CDT)
workstream: WS7
status: shipped — three fixes in one commit
commits:
  - aaca9ff — fix(1828): LLM proxy — usage decode + system/user prompt split + minimax default
relates_to:
  - .spec/journal/2026-05-20-frontend-phase-5-cutover.md
  - projects/1828-illuminated/.spec/proposals/llm-proxy.md
---

# 2026-05-21 — 1828 LLM proxy: three fixes

## What broke

Michael minted a BYOK session against opencode-go + kimi-k2.6 via the
phase-5 Settings UI, hit Render on a long D&C 84 priesthood passage,
and got back:

```json
{
  "error": "upstream_provider_error",
  "message": "The upstream LLM provider returned an error. This is not a 1828.ibeco.me throttle.",
  "upstream_message": "decode upstream: json: cannot unmarshal object into Go struct field .usage of type int"
}
```

His ask: *I have my key in .env, and you have the logs in the docker
container, please test a few examples to verify API works.*

## Root cause

`backend/internal/llmproxy/llmproxy.go callUpstream()` declared:

```go
var parsed struct {
  Choices []struct{ Message struct{ Content string } }
  Usage   map[string]int `json:"usage"`
}
```

Direct curl to opencode-go confirmed the upstream response shape:

```json
"usage": {
  "prompt_tokens": 54,
  "completion_tokens": 300,
  "prompt_tokens_details": {"cached_tokens": 4},
  "completion_tokens_details": {"reasoning_tokens": 300}
}
```

`prompt_tokens_details` and `completion_tokens_details` are NESTED
OBJECTS that reasoning-class models (kimi-k2.6, glm-5.1, deepseek-v4-flash,
all of opencode-go's lineup, plus OpenAI's o-series) add to `usage`.

Go's JSON decoder accepts unknown fields in structs but **rejects
unknown shapes into typed map values** — `map[string]int` cannot
decode an object value. The whole decode bailed.

## Fix 1 — usage decoded into a typed struct

```go
Usage struct {
  PromptTokens     int `json:"prompt_tokens"`
  CompletionTokens int `json:"completion_tokens"`
  TotalTokens      int `json:"total_tokens"`
} `json:"usage"`
```

Then build `map[string]int` AFTER decoding so caller signatures stay
unchanged. Unknown nested-object fields are silently ignored — the
correct behavior for an upstream we don't fully control.

## Adjacent surface: prompt was leaking reasoning into content

Re-testing with the fix in place showed the API succeeded but the
modernized output was buried in 1850 tokens of "Let me break this
down..." reasoning text. Direct curl showed kimi-k2.6 puts reasoning
INTO `choices[0].message.content` — not a separate `reasoning_content`
field, no thinking-flag switch on the opencode-go gateway.

Looking at the substrate's `bgworker.rs` for comparison:
```rust
// Reasoning capture. Field names vary by gateway:
//   OpenRouter / OpenCode Go: `reasoning` (string), `reasoning_details` (array)
//   Moonshot direct:          `reasoning_content` (string)
```

The substrate captures these separately. The 1828 backend was reading
only `content`, getting the merged reasoning-plus-answer.

## Fix 2 — split prompt into system + user messages

Previously `buildRenderPrompt()` returned ONE string sent as a single
`role: user` message. Reasoning models interpreted the rules-plus-passage
as one big chunk to "think through."

Split into `(systemPrompt, userPrompt)`:
- System: the rules + "output only" discipline
- User: just the passage + flagged-word table

Request body now:
```go
"messages": []map[string]string{
  {"role": "system", "content": systemPrompt},
  {"role": "user",   "content": userPrompt},
}
```

`buildRenderPrompt()` signature changed; callers updated. Cut reasoning-
leak dramatically.

## Fix 3 — change default model from kimi-k2.6 to minimax-m2.7

Tested every model on opencode-go's lineup (15 total). Most are
reasoning-class — they burn the entire `max_tokens` budget on
reasoning and emit empty content (deepseek-v4-flash, glm-5.1 both
returned empty content with 2000 reasoning_tokens spent).

**minimax-m2.7** produced a clean single-line modernization in ~5
seconds:

> *"For whoever is faithful in holding [obtaining] these two priestly
> offices [priesthoods] of which I have spoken, and in enlarging
> [magnifying] their calling, are made holy [sanctified] by the Spirit
> for the renewing of their bodies."*

Changed `useLLMSettings.ts` preset:
```ts
'opencode-go': {
  baseUrl: 'https://opencode.ai/zen/go/v1',   // was ''
  model: 'minimax-m2.7',                       // was 'kimi-k2.6'
}
```

`opencode-zen` preset got the same model default.

## End-to-end verification

Michael's exact original failing input, retried against the rebuilt
backend with the fixed parser + system/user prompt + minimax-m2.7
default:

```
duration: 4941ms
usage: {prompt_tokens: 29, completion_tokens: 644, total_tokens: 673}

For whoever is faithful to the holding these two offices of the
priesthood of which I have spoken, and to the making great their
calling, are made holy by the Spirit unto the renewing of their bodies.
```

Clean. The substitution markers ([original]) dropped this run — minimax
folded the 1828-sense substitutions directly into prose without the
bracketed echo. That's a model-prompt-following deviation, accepted
for v1 as good-enough output quality.

## Honest carry-forward

- **Substitution markers** not always present in minimax output. A
  tighter system prompt or a model with stronger instruction-following
  could enforce. Filed as v2 polish.
- **opencode-go gateway is Chinese-model-only** (minimax/kimi/glm/
  deepseek/qwen/mimo). For OpenAI/Anthropic the reader picks
  `openai`/`openrouter`/`custom` in Settings.
- **kimi-k2.6** still in the dropdown — readers can pick it knowing
  it's verbose. The default landing pad is the model that works
  cleanly out of the box.
- **Sessions die on backend restart** (in-memory only, by design).
  Re-mint via Settings. Won't be journaled again every time.

## Why this matters

Three fixes in one commit because all three surfaced in the same
~30 minutes of curl-based diagnosis against the live backend with
Michael's actual key. The first fix was the bug Michael named (502
error in his face); the second was discovered while testing the
first; the third was discovered while testing the second. None of
the three would have been caught by the mock provider — they all
require a real upstream that does real reasoning. The lesson is
about testing with real keys against the real provider matrix, not
just unit-test-shaped mocks.
