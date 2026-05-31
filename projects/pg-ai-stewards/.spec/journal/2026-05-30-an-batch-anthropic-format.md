---
date: 2026-05-30
title: Batch AN — Anthropic-format dispatch (+ LM Studio chat confirmed)
workstream: WS5
session_type: dev (autonomous build per Michael's delegation)
status: AN.1–AN.5 SHIPPED + live-verified end-to-end
---

# Batch AN — the substrate can now call Anthropic-format models

## Where it started

The model-check session found qwen3.7-max fails on the substrate because it's an
**Anthropic-format** model on opencode (needs `POST /messages`, not the
OpenAI-compat `/chat/completions` the substrate speaks). Michael: "spec the plan
to add anthropic class models… double-check LM Studio too… once specced, ratify,
if no open questions hop into building… make the best decisions you can without
me, in line with intent — get anthropic class models working." He freed the
substrate for the night. So: autonomous build, decisions resolved by delegation.

## Grounding (verified live before building)

- `POST /zen/go/v1/messages` with **`x-api-key`** (NOT Bearer) + **`anthropic-version: 2023-06-01`** → qwen3.7-max responds. Response is Anthropic-shaped: `content:[{type:thinking,…},{type:text,text}]`, `stop_reason`, `usage:{input_tokens,output_tokens}`. SSE: `message_start`/`content_block_delta`(text_delta|thinking_delta)/`message_delta`(stop_reason+output_tokens).
- **LM Studio chat already works** on the existing path: provider `lm_studio` (kind=openai, base host.docker.internal:1234/v1, default qwen/qwen3.6-27b) is registered + reachable; auto-probe of qwen/qwen3.6-27b → usable. No Anthropic work for it — only missing pricing rows.

## Design — normalize at the parse boundary

The chat-extraction code reads an OpenAI-shaped object. So the Anthropic path
reassembles the Anthropic response **into that same OpenAI shape** at parse time
(`parse_anthropic_sse`): text→content, thinking→reasoning_content,
stop_reason→finish_reason, input/output_tokens→prompt/completion_tokens. Every
downstream consumer (extraction, cost, cache tokens, messages, markers) is
UNCHANGED. Mirrors how `parse_chat_sse` already normalizes OpenAI SSE.

## What shipped

- **AN.1** (SQL): `model_capability.api_format` (openai|anthropic) + `model_api_format()`. A **BEFORE INSERT trigger on chat work_queue rows** stamps `payload.api_format` — covers both `work_item_dispatch_stage` AND the direct-insert probe path (so probing qwen3.7-max through the new path works). Seeded qwen3.7-max + minimax-m2.7 = anthropic.
- **AN.2** (Rust): `chat()` branches on `payload.api_format`. anthropic → `/messages` + x-api-key + anthropic-version, body translated (`anthropic_body_from_openai`: system extracted to top-level, max_tokens required default 4096, tools stripped v1, stream:true), parsed by `parse_anthropic_sse` → OpenAI shape. openai → unchanged. pg rebuilt (33s extension compile; AGE cached).
- **AN.3**: verified live — auto-probe of qwen3.7-max + minimax-m2.7 → usable=true, finish=stop, reasoning captured separately. Regression: deepseek-v4-flash (OpenAI) still works. Refreshed the stale "unusable" notes (live + j10 + 4a).
- **AN.4** (SQL): LM Studio chat models (qwen/qwen3.6-27b + 2 others) added to model_pricing at $0 (local hardware). They appear in model_catalog.
- **AN.5**: end-to-end — a real redline child on **qwen3.7-max** through the full pipeline produced a substantive location-anchored report (4026 tokens, $0.04, verified maturity, correct Current/Proposed/Touches-flag format). The model Michael "truly ran into issues with" now works in a real pipeline.

## Decisions (resolved by delegation — see the proposal)
D-AN1 api_format on model_capability; D-AN2 stamp via work_queue trigger (covers
probe path); D-AN3 x-api-key+anthropic-version; D-AN4 normalize-at-parse; D-AN5
max_tokens required default 4096, system extracted; D-AN6 **v1 tools-OFF** (Anthropic
tool-format translation deferred); D-AN7 LM Studio = existing path + pricing rows.

## State + observations
- opencode_go now **16/16 usable** (qwen3.7-max + minimax-m2.7 live via Anthropic).
- LM Studio: 3 chat models, $0, usable.
- gemini **8/10** — the M.5 auto-probe autonomously caught `gemini-2.0-flash` +
  `gemini-2.0-flash-lite` returning HTTP 404 (stale/deprecated model IDs). The
  auto-probe working as designed; prune or fix those 2 catalog rows later.

## Carry-forward
- **Anthropic tool-format** (tool schema + tool_use/tool_result loop) for
  tool-using pipelines on Anthropic models — v1 strips tools.
- **api_format could be probe-detected** rather than seeded (the auto-probe could
  try /messages on a 401 oa-compat).
- **Prune the 2 dead gemini-2.0-flash* rows** (404).
- No MCP binary change this batch (bgworker/extension + SQL only); stewards-mcp
  unaffected.

## Pace
Autonomous build on Michael's explicit grant + freed substrate. Soak paused for
the pg rebuild, resumed at close. One pg rebuild (the AN.2 Rust). 5 commits. The
C–F cadence held — verified each phase live; the only real spend was the
end-to-end qwen3.7-max redline (~$0.04) + a handful of $0/cent probes.
