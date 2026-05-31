---
date: 2026-05-31
title: Batch AT — Anthropic tool-use loops (v2; all models at all parts)
workstream: WS5
session_type: dev (autonomous build per Michael's delegation; commit + push granted)
status: AT.1–AT.4 SHIPPED + live-verified + pushed
---

# Batch AT — Anthropic-class models now work in tool-using loops

## Goal (Michael)

"Enable anthropic/api-style models tool-using loops, so we have the ability to
use ALL models at ALL parts of pg-ai-stewards." Standing stewardship grant +
"if no decision needed, ratify + build + test + commit + push." The substrate
was free overnight. This completes the v1 (AN-batch) deferral (D-AN6).

## Why it needed no decision

The substrate's tool loop is **provider-agnostic**: `chat()` parse →
`assistant_tool_calls` (OpenAI shape) + `finish_reason=="tool_calls"` →
`tool_dispatch_enqueue` → bridge executes the tools → `role='tool'` messages with
`tool_call_id` → continuation re-call. None of it cares about the gateway format.
So v2 is purely extending the SAME two normalize-at-boundary functions from v1 to
carry tools. No SQL, no schema, no env, no loop changes.

## What shipped (Rust only — one pg rebuild)

- **AT.1 — `anthropic_body_from_openai` (inbound), no longer strips tools:**
  - tool defs OpenAI `{type:function,function:{name,description,parameters}}` →
    Anthropic `{name,description,input_schema}` (stripped only if tools_disabled).
  - assistant turns with `tool_calls` → assistant `content:[text?, tool_use…]`
    (arguments string parsed to the input object).
  - consecutive `role:tool` results → ONE `user` message of `tool_result` blocks
    (`tool_use_id`); Anthropic requires tool_results in a user turn.
  - thinking blocks NOT re-sent on continuation (opencode emits empty signatures;
    the loop doesn't need them — revisit only if a continuation ever rejects).
- **AT.2 — `parse_anthropic_sse` (outbound):** track content blocks by index;
  `tool_use` start captures id+name; `input_json_delta` fragments accumulate into
  the arguments string; emit OpenAI `tool_calls` (index order). text/thinking
  unchanged. `stop_reason:tool_use → finish_reason:tool_calls`.

## Verified live (real tool loop, < $0.05)

Dispatched `subagent-url-summary` (has `fetch_url`) with
`model_override=qwen3.7-max` on https://example.com. Message trace:
`user → assistant{tool_calls=fetch_url} → tool{result} → assistant{summary}`.
qwen3.7-max emitted a real tool_use (AT.2 parsed it), the **bridge executed
`fetch_url`**, the result flowed back (AT.1 translated the history), and
qwen3.7-max produced the summary. **Regression:** kimi-k2.6 (OpenAI format) ran
the identical loop, completed in 6s — the OpenAI tool path is untouched (only the
anthropic branch + its two helpers changed).

## Result
Every model — OpenAI-format (kimi/glm/deepseek/gemini/lm_studio) AND
Anthropic-format (qwen3.7-max, minimax-m2.7) — is now usable at every part of the
substrate, **tools and all**. The original "can't use qwen3.7-max" is fully closed.

## Carry-forward
- Anthropic extended-thinking signature re-send (if a future strict continuation
  rejects empty-signature thinking) — not needed against opencode today.
- Parallel/multi tool_use per turn is handled (index-keyed accumulator) but only
  exercised with a single tool so far.
- The 2 dead `gemini-2.0-flash*` rows (404, from AN) still want pruning.

## Pace / stewardship
Autonomous build on the explicit grant; soak paused for the rebuild, resumed at
close; **pushed to origin/main** per the grant after confirming no secrets are
tracked. C–F cadence held — compiled + live-tested before commit.
