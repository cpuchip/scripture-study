# Proposal: Anthropic-format dispatch (+ LM Studio chat confirmation)

- **Status:** ratified-by-delegation 2026-05-30 (Michael: "spec it, ratify, if no open questions hop into building… make the best decisions you can without me, in line with intent — get Anthropic-class models working"). Build in progress.
- **Workstream:** WS5 / pg-ai-stewards
- **Motivation:** opencode's gateway serves some models ONLY in Anthropic format (`POST /zen/go/v1/messages`), not the OpenAI-compat `/chat/completions` the substrate speaks. `qwen3.7-max` and `minimax-m2.7` are the wanted ones. The M-batch auto-probe + the opencode docs confirmed this is a format mismatch, not a key/flakiness issue.

## Confirmed facts (verified live 2026-05-30, key in-container)

- `POST https://opencode.ai/zen/go/v1/messages` with **`x-api-key: <key>`** (NOT `Authorization: Bearer`) + **`anthropic-version: 2023-06-01`** → qwen3.7-max responds. Bearer auth → "Missing API key"; oa-compat endpoint → 401 "not supported for format oa-compat".
- **Non-stream response:** `{type:message, role:assistant, content:[{type:"thinking",thinking:…},{type:"text",text:"OK"}], stop_reason:"end_turn", usage:{input_tokens,output_tokens,cache_creation_input_tokens,cache_read_input_tokens}}`. qwen3.7-max is a reasoning model (emits a `thinking` block).
- **Streaming SSE:** `message_start` (message.usage.input_tokens) → `ping` → `content_block_start` (content_block.type = thinking|text) → `content_block_delta` (delta.type = `thinking_delta`{thinking} | `text_delta`{text}) → `content_block_stop` → `message_delta` (delta.stop_reason + usage.output_tokens) → `message_stop`.
- **LM Studio chat already works** via the EXISTING OpenAI path: provider `lm_studio` (kind=openai, base `http://host.docker.internal:1234/v1`, default `qwen/qwen3.6-27b`) is registered + reachable; auto-probe of `qwen/qwen3.6-27b` returned usable/finish=stop. No Anthropic work for LM Studio — only missing `model_pricing` rows.

## Design — normalize at the parse boundary

The substrate's chat-completion extraction (`bgworker.rs` ~1840–1942) reads an **OpenAI-shaped** parsed object (`choices[0].message.content`, `usage.prompt_tokens/completion_tokens`, etc.). So the Anthropic path reassembles the Anthropic response **into that same OpenAI shape** — every downstream consumer (extraction, cost, messages, markers, apply handlers) is then UNCHANGED. This mirrors how `parse_chat_sse` already normalizes OpenAI SSE → the non-stream shape.

### Decisions (resolved per delegation)
- **D-AN1 — format storage:** `model_capability.api_format text NOT NULL DEFAULT 'openai'` (values: `openai` | `anthropic`). Co-located with usability; one source of truth. `model_api_format(provider, model)` fn (default 'openai' for unrowed).
- **D-AN2 — bgworker learns the format:** `work_item_dispatch_stage` stamps `payload.api_format` from `model_api_format(provider, resolved_model)` (same pattern as the R.3 `tools_disabled`/`max_tokens` stamps). The bgworker branches on `payload.api_format`.
- **D-AN3 — auth/headers:** when `api_format=anthropic`, use `x-api-key: <provider key>` + `anthropic-version: 2023-06-01` (NOT bearer).
- **D-AN4 — normalize:** Anthropic response (stream + non-stream) → OpenAI internal shape. `text` blocks → `choices[0].message.content`; `thinking` blocks → `reasoning_content`; `stop_reason` → `finish_reason` (end_turn→stop, max_tokens→length, tool_use→tool_calls); `usage.input_tokens`→`prompt_tokens`, `output_tokens`→`completion_tokens`. Cache fields already map (the extraction reads `cache_creation_input_tokens`/`cache_read_input_tokens` verbatim — Anthropic-native).
- **D-AN5 — request body translation:** OpenAI body → Anthropic body. Extract any `role:system` message(s) → top-level `system` string. `max_tokens` is REQUIRED by Anthropic → `COALESCE(body.max_tokens, 4096)`. `stream:true` (ES.6). Set the per-request fields.
- **D-AN6 — v1 scope = tools-OFF.** The immediate intent (qwen3.7-max in brainstorm/redline/chat panels) is tools-off. Anthropic tool-format translation (different tool schema + tool_use/tool_result loop) is a documented follow-up. v1 strips `tools` from the Anthropic request. (A tools-on Anthropic-format dispatch simply runs without tools in v1.)
- **D-AN7 — LM Studio:** confirmed working on the existing path; add `model_pricing` rows for the local chat models (free, local) so they appear in the catalog + cost-track at $0. No code change.

## Phases
- **AN.1** SQL: `model_capability.api_format` + `model_api_format()` + dispatch stamps `payload.api_format`; seed qwen3.7-max + minimax-m2.7 = `anthropic`. (live-apply; inert until the bgworker is rebuilt)
- **AN.2** Rust: `chat()` Anthropic branch (endpoint/auth/headers/body-translate) + `parse_anthropic_sse` → OpenAI shape. (pg rebuild)
- **AN.3** rebuild pg + smoke: probe qwen3.7-max + minimax-m2.7 via the substrate → real content back; flip `usable`.
- **AN.4** LM Studio `model_pricing` rows (free) + catalog confirm.
- **AN.5** verify (a redline/brainstorm child on qwen3.7-max) + memory + commit.

## Acceptance
- `enqueue_model_probe('opencode_go','qwen3.7-max')` → `usable=true`, real content, finish=stop, reasoning captured separately.
- An existing OpenAI-format dispatch (kimi/deepseek/gemini) is byte-identical (the branch only triggers on `api_format=anthropic`).
- LM Studio chat models appear in `model_catalog` at $0.

## Follow-ups (not v1)
- ~~Anthropic tool-format translation~~ → **v2 below.**
- Drive `api_format` from a probe (the auto-probe could detect format), rather than a seeded value.

---

# v2 — Anthropic tool-use loops (batch AT, ratified-by-delegation 2026-05-31)

**Goal (Michael):** "enable anthropic/api-style models tool-using loops, so we have the ability to use all models at all parts of pg-ai-stewards." Build + test + commit + push under the standing stewardship grant; no decision needed.

**Why it's decision-free:** the substrate's tool loop is provider-agnostic — `chat()` parse → `assistant_tool_calls` (OpenAI shape) + `finish_reason=="tool_calls"` → `tool_dispatch_enqueue` → bridge executes the tools → `role='tool'` messages with `tool_call_id` → continuation re-call. None of that cares about the gateway format. v1 already proved the normalize-at-parse approach. v2 just extends the SAME two boundary functions to carry tools. No SQL, no schema, no env, no loop changes.

**Confirmed shapes (live 2026-05-31):** Anthropic tool-use SSE: `content_block_start{content_block:{type:"tool_use", id, name}}` → `content_block_delta{delta:{type:"input_json_delta", partial_json}}` (fragments assemble to the input JSON) → `content_block_stop` → `message_delta{stop_reason:"tool_use"}`.

- **AT.1 — `anthropic_body_from_openai` (inbound translation), no longer strips tools:**
  - tool defs: OpenAI `{type:function, function:{name, description, parameters}}` → Anthropic `{name, description, input_schema:parameters}`; stripped only when `tools_disabled`.
  - assistant turns with `tool_calls` → Anthropic `assistant` with `content:[{type:text,…}?, {type:tool_use, id, name, input:<parsed arguments>}…]`.
  - `role:tool` results → grouped into a single `user` message `content:[{type:tool_result, tool_use_id, content}…]` (consecutive tool messages merge; Anthropic requires tool_results in a user turn).
  - thinking blocks are NOT re-sent on continuation (opencode emits empty signatures; tool-loop doesn't need them — revisit if a continuation rejects).
- **AT.2 — `parse_anthropic_sse` (outbound), accumulate tool_use:** track content blocks by index; `tool_use` start captures id+name; `input_json_delta` fragments accumulate into the arguments string; emit OpenAI `tool_calls:[{id, type:function, function:{name, arguments}}]`. `stop_reason:tool_use → finish_reason:tool_calls` (already mapped).
- **AT.3** rebuild pg; **test** a real tool loop: dispatch a tool-using pipeline (`subagent-url-summary`, which has `fetch_url`) with `model_override=qwen3.7-max` — qwen3.7-max should emit a `fetch_url` tool_use, the bridge executes it, the result flows back, and qwen3.7-max produces the summary. Regression: an OpenAI-format tool loop (kimi) still works.
- **AT.4** memory + commit + **push** (explicit grant).

**Acceptance:** qwen3.7-max completes a multi-round tool-using pipeline (calls a real tool, consumes the result, finishes); OpenAI tool loops unchanged. Result: every model — OpenAI-format and Anthropic-format — is usable at every part of the substrate, tools and all.
