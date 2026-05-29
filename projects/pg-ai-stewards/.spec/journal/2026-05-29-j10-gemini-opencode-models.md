---
date: 2026-05-29
title: J.10 — Gemini usable + expanded opencode_go model pricing
batch: J.10
status: shipped — committed, soak resumed
commit: (this session, after a5a6ae8/d7949b0)
---

# J.10 — Gemini + opencode model expansion

## Ask

Michael: "can we get gemini models in there from my gemini key" + "add the other go models for opencode go, like deepseek?"

## What I found (council moment paid off)

1. **Provider config is env-based** (`STEWARDS_PROVIDER_<NAME>_<FIELD>` in `.env`, loaded into the in-memory registry by the bgworker at startup). `model_pricing` is a separate SQL table for **cost tracking + UI visibility**, NOT capability — a model with no pricing row still dispatches (compute_cost flags `no_pricing_row`).

2. **The Gemini base_url was wrong.** Configured `.../v1beta/`; the substrate posts to `{base_url}/chat/completions` (bgworker.rs:1777), so it would 404. Gemini's OpenAI-compat endpoint is under `/v1beta/openai`. Caught this by reading the URL-builder before touching anything.

3. **Real model lists + prices, not memory.** Queried the live `/models` endpoints from inside the container (keys never left the container) — got the actual opencode_go + gemini model ids. Neither API returns prices, so fetched the pricing pages. Gemini ids carry a `models/` prefix in listings but the chat endpoint accepts bare ids (tested both live).

## What shipped

- **`.env` fix** (gitignored; surgical sed, keys never read into my context): base_url → `.../v1beta/openai`. `.env.example` updated with the gotcha documented + `KIND=openai`.
- **pg recreated** (`--force-recreate`) to reload the registry. `providers_loaded()` confirms the corrected URL. Bridge + ui stayed up.
- **`j10-provider-models-pricing.sql`** — 10 Gemini + 7 opencode_go chat-model pricing rows, real prices (2026-05-29), tiered Gemini using ≤200k rate + notes, fixed `effective_at` for idempotent re-runs. `/api/models` went 9 → 26.

## Smoke (live, both providers)

One brainstorm, two lenses routed to the two new providers:
- **DeepSeek** (`deepseek-v4-flash`, opencode_go): real content + cost_event ($0 free, correct).
- **Gemini** (`gemini-2.5-flash-lite`, google_gemini): real SCAMPER content, completed + verified. base_url fix confirmed working end-to-end.

## Carry-forward — Gemini cost tracking gap (precise)

Gemini dispatches produce content but record **no cost_event**, so Gemini work_items show $0. Root cause traced exactly:
- bgworker records a cost_event only when usage tokens > 0 (`bgworker.rs:845`), reading `usage.prompt_tokens` / `usage.completion_tokens` (`1864-1871`).
- Gemini returns those on **non-streamed** responses (verified by direct curl) but **omits the usage block on streamed responses** unless `stream_options.include_usage=true` is sent.
- The opencode gateway includes usage in-stream (so DeepSeek recorded cost); Gemini's streamed calls don't.

`model_pricing` is in place — `compute_cost('google_gemini','gemini-2.5-flash',1M,500k)` = $1.55 verified — so the moment usage is captured, cost computes correctly. **Fix:** send `stream_options.include_usage=true` on streamed dispatch (or non-stream direct providers). That's a `bgworker.rs` change + `pg` rebuild — a separate batch, NOT crept into this one.

Secondary: gemini cost would also need its own `cost_buckets` rows if budget enforcement is wanted (current buckets are opencode_go-only; gemini is Michael's own per-use key).

## NOT priced (genuinely unpublished on opencode's table 2026-05-29)

deepseek-v4-pro, mimo-v2-pro, mimo-v2.5-pro, mimo-v2-omni, hy3-preview. They dispatch if named; compute_cost flags `no_pricing_row`. No fabricated 0-cost rows. Add when opencode publishes rates (or Michael supplies them).

## Files

- `extension/j10-provider-models-pricing.sql`
- `extension/.env.example` (Gemini base_url + gotcha doc)
- `extension/.env` (gitignored — base_url fix, not committed)
- `extension/smoke/{list-provider-models,inspect-model-schema,test-gemini-chat}.sh`, `j10-smoke-dispatch.sql`
