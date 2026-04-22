---
workstream: WS1
status: proposed
brain_project: 6
created: 2026-03-21
last_updated: 2026-04-21
---

# Classifier Fix: Qwen 3.5 9B Empty Response

**Date:** 2026-03-21  
**Status:** Ready to build — small, low-risk  
**Scope:** `internal/ai/lmstudio.go`, `internal/classifier/profiles.go`

---

## Binding Problem

Auto-classification fails for *every* inbox entry when the brain runs with Qwen 3.5 9B as the LM Studio backend. All classify attempts exhaust 3 retries and log:

```
classification failed: empty response from LM Studio (model: qwen/qwen3.5-9b)
```

This blocks the brain's primary workflow: relay picks up items from the phone, sends them to the classifier, and they fail silently. Entries sit uncategorized in `inbox/`. The brain can receive thoughts but can't route them.

---

## Root Cause

Two issues found in code, one known root cause, one confirmed bug.

### Issue 1 — Thinking model + grammar sampling (root cause)

The profile for `qwen/qwen3.5-9b` in `internal/classifier/profiles.go`:

```go
"qwen/qwen3.5-9b": {
    NoThink:          true,
    StructuredOutput: true,   // ← the conflict
}
```

`StructuredOutput: true` routes classification to `CompleteStructuredJSON`, which sends `response_format: json_schema` to the LM Studio API. LM Studio uses llama.cpp grammar-based sampling to force JSON output.

Qwen 3.5 is a thinking model. When grammar sampling is active, LM Studio's grammar forces token selection toward the schema immediately — but the model internally wants to emit `<think>...</think>` reasoning tokens first. These two systems conflict at the LM Studio layer. The result: `Choices[0].Message.Content` comes back as `""` in the API response.

The `NoThink: true` flag appends `/no_think` to the system prompt, which is a text instruction. It works fine for regular completion but does not switch off grammar/thinking interaction at the llama.cpp layer.

Confirmed: the empty check in `CompleteStructuredJSON` fires on the raw API content *before* `stripThinkingContent` runs, meaning LM Studio is returning empty content — not thinking content that we're stripping to empty.

### Issue 2 — `"strict": "true"` type error (confirmed bug)

In `internal/classifier/classifier.go`, `ClassificationSchema()`:

```go
"json_schema": map[string]any{
    "name":   "classification",
    "strict": "true",   // ← string, should be boolean true
    "schema": ...
}
```

The OpenAI specification defines `strict` as a boolean. Passing the string `"true"` is invalid. LM Studio may silently ignore the strict flag, reject the schema, or fall back to unstructured output. This may compound the empty response problem and will cause issues if/when we use structured output with other backends.

---

## Fix

### Fix 1 — Disable `StructuredOutput` for `qwen/qwen3.5-9b` (primary, immediate)

**File:** `internal/classifier/profiles.go`

Set `StructuredOutput: false` for the Qwen 3.5 9B profile. This routes classification through `CompleteJSON` → `Complete()` → plain text completion with JSON fence stripping and one retry pass. The `/no_think` system prompt (`NoThink: true`) still suppresses thinking tokens in plain completion mode, and `stripThinkingContent` handles any `<think>` leakage.

**Why this is the right fix**: We know thinking models + grammar sampling = broken in LM Studio. Plain completion works fine with prompt-level thinking suppression and our existing post-processing. Structured output is a "nice to have" for guaranteed schema compliance, but the existing JSON retry fallback is sufficient for a fast local classifier.

```go
"qwen/qwen3.5-9b": {
    ID:               "qwen/qwen3.5-9b",
    Name:             "Qwen 3.5 9B",
    Tasks:            []Task{TaskClassify, TaskChat},
    Temperature:      0.1,
    NoThink:          true,
    StructuredOutput: false,   // ← changed: grammar sampling conflicts with thinking model
},
```

### Fix 2 — `"strict": true` (boolean) in schema (cleanup, low risk)

**File:** `internal/classifier/classifier.go`, `ClassificationSchema()`

```go
"strict": true,   // not "true"
```

This is a correctness fix. It doesn't affect LM Studio behavior for local classification today (string vs bool both pass through the map), but when the copilot backend or other strict-conformant clients use the schema, this matters.

---

## Verification

After Fix 1:

```
# Start brain
.\scripts\brain\start.ps1

# Watch the startup classify queue flush — should see:
# [relay] classified entry ... → category: (ideas|projects|actions|...) (confidence: 0.xx)
# NOT: auto-classify attempt 1/3 failed
```

After Fix 2: No observable behavior change, but `go vet ./...` and `go build ./...` should still pass. Schema correctness verified by inspection.

---

## What This Does NOT Fix

- Why Qwen 3.5 9B was originally set to `StructuredOutput: true`. It worked at some point, or the assumption was wrong. Either way, plain completion is correct for a thinking model.
- The underlying LM Studio behavior with thinking + grammar sampling. That's a model/runtime constraint we work around, not through.
- `StructuredOutput` support for future non-thinking models (Ministral 3B, Qwen 3 1.7B). Those profiles have it `false` and they'd need independent validation before enabling.

---

## Scope

**Files changed:**
- `internal/classifier/profiles.go` — 1 line change (`StructuredOutput: false`)
- `internal/classifier/classifier.go` — 1 line change (`"strict": true`)

**Tests to run:** `go test ./internal/classifier/...`

**No DB changes, no config changes, no new dependencies.**
