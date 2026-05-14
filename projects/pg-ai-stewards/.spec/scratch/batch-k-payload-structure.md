---
title: Batch K — What Actually Gets Sent to the LLM Each Turn
status: structural reference / pre-proposal
date: 2026-05-13
project: pg-ai-stewards
purpose: |
  Michael asked for a clear mental model of the payload structure so we can
  reason about WHERE to apply compaction. This file is the structural answer.
  Everything in it is read directly from our SQL (compose_system_prompt in 5d3,
  compose_messages in 0.2.0.sql lines 725-789).
---

# What gets sent to the LLM each turn

When the bgworker calls a provider's `/v1/chat/completions` endpoint, the body it sends has this shape:

```jsonc
{
  "model": "kimi-k2.6",                  // resolved from agent.model_pin or stage.model
  "messages": [ ... ],                   // <- THIS IS WHAT GROWS
  "tools": [ ... ],                      // tool definitions, NOT tool calls
  "temperature": 0.4,
  "top_p": 0.9,                          // optional, from agent.top_p
  "response_format": { ... }             // optional, from agent.response_format
}
```

The whole game is the `messages` array. Everything else is fixed-size per dispatch.

## The messages array — three structural zones

```jsonc
"messages": [
  // ZONE 1: SYSTEM (one message, always position 0)
  {
    "role": "system",
    "content": "<everything from compose_system_prompt>"
  },

  // ZONE 2: HISTORY (zero to N messages, grows over the session)
  { "role": "user",      "content": "..." },
  { "role": "assistant", "content": "...", "tool_calls": [...] },
  { "role": "tool",      "tool_call_id": "...", "content": "..." },
  { "role": "tool",      "tool_call_id": "...", "content": "..." },
  { "role": "assistant", "content": "...", "tool_calls": [...] },
  { "role": "tool",      "tool_call_id": "...", "content": "..." },
  // ... continues until the session ends or compose_messages gets too big

  // ZONE 3: CURRENT TURN (optional, appended only when p_user_input is set)
  { "role": "user", "content": "<this turn's input_template-rendered prompt>" }
]
```

That's it. Three zones. The system message is built once per dispatch (cheap to re-derive). The history grows monotonically (this is where the 426KB problem lives). The current-turn message is the binding question or stage input.

## Zone 1 — The system message (built by `compose_system_prompt`)

Our system message is **composed of 4-5 blocks**, concatenated:

```text
[1] agent.prompt                       — the role-specific persona (e.g. "You are
                                          the SCAMPER lens for a brainstorming
                                          pipeline…")

[2] global + agent-scoped instructions — from stewards.instructions where
                                          scope IN ('global', 'agent:<family>')
                                          ordered by ord

[3] covenant_block (5d3, Phase C.4)    — from the active covenant. Includes:
                                          purpose, human_commits_to,
                                          agent_commits_to, when_broken/recovery,
                                          and council_moment if set

[4] intent_block (5d3, Phase C.4)      — from stewards.intents matching the
                                          work_item.intent_id. Includes:
                                          purpose, beneficiary, values_hierarchy,
                                          non_goals, scripture_anchor

[5] <available_skills> block           — XML-shaped list of skill names + short
                                          descriptions; only present if the
                                          'skill' tool is not denied for this
                                          agent. Skills load on-demand via the
                                          'skill' tool call.
```

**Typical size:** 800-2000 tokens depending on agent. The 5d3 covenant+intent blocks add ~600 tokens per dispatch as a fixed overhead (measured cost noted in CLAUDE.md §7).

**Stability across turns:** identical every dispatch of the same session. This is the ideal prefix for Anthropic-style prompt caching IF we ever turn it on. Critical for K: **never compact this zone.**

## Zone 2 — History (`compose_messages` joins `stewards.messages`)

This is where context grows. Every row of `stewards.messages WHERE session_id = $1` gets emitted in order:

```sql
SELECT coalesce(jsonb_agg(message_object ORDER BY m.created_at, m.id), '[]'::jsonb)
  FROM stewards.messages m
 WHERE m.session_id = p_session_id;
```

No filter. No size guard. **Every message in the session is included every turn.**

The OpenAI message shape varies by role:

| Role | Required fields | Optional | Where it comes from |
|---|---|---|---|
| `user` | role, content | — | initial dispatch input, user follow-ups |
| `assistant` | role, content (may be "") | tool_calls, reasoning_content, reasoning_details | LLM response |
| `tool` | role, tool_call_id, content | — | bridge tool execution result |
| `system` | role, content | — | rare — usually only at position 0 |

**Where the bytes go:**

- `assistant.content` — the model's natural-language response (typically small, 200-2000 chars)
- `assistant.tool_calls` — JSON array of function calls (typically 100-1500 chars total, but grows linearly with parallel tool calls)
- `assistant.reasoning_content` / `reasoning_details` — only for reasoning models (qwen-plus, deepseek-r1). Can be 5-50KB per assistant turn.
- `tool.content` — **THIS IS THE 426KB PROBLEM**. fetch_url, web_search, gospel_search results can be enormous.

## What the J.3 Crystal Radio session looked like

`wi--b0b1185f--gather` — 14 messages, 496 KB total. Reconstructed:

```
[0] system       — Zone 1, ~1800 tokens (research agent prompt + intent + covenant + skills)
[1] user         — 7,360 chars (the binding question + context_gather briefing)
[2] assistant    — 935 chars of tool_calls (no narrative response)
[3] tool         — 4,845 chars (web_search hits)
[4] tool         — 28,876 chars (fetch_url of a moderate-size source)
[5] tool         — 3,366 chars
[6] tool         — 21,183 chars
[7] tool         — 3,545 chars
[8] assistant    — 703 chars of tool_calls
[9] tool         — 1,793 chars
[10] tool        — 1,710 chars
[11] tool        — 6,740 chars
[12] tool        — 426,651 chars  ← THE POISON. fetch_url of a huge source.
[13] user        — 207 chars
[14] user        — 207 chars
```

By the time the LLM was asked to synthesize, the prompt was: system (1.8K tokens) + Zone 2 (496 KB ≈ 165K tokens just text + ~2K tokens of tool_calls JSON) + Zone 3 user message. Total ~167K tokens, but tokenization overhead pushed Moonshot's count to **376,671 tokens** — well past 262,144.

**Where K can intervene:**
- Row 12 alone (426KB) is enough to fail. **Single-result offload** (Shape A) replaces it with a 200-token summary while preserving the original.
- Rows 4 + 6 + 12 cumulative is 477KB. **Session-cumulative cap** with head/torso/tail compaction would summarize the older middle while keeping rows 0, 1, and the last 3-4 untouched.

## Zone 3 — The current turn (optional `p_user_input`)

`compose_messages(p_user_input)` appends a final user message if `p_user_input` is supplied:

```sql
IF p_user_input IS NOT NULL THEN
    v_result := v_result || jsonb_build_array(
        jsonb_build_object('role', 'user', 'content', p_user_input)
    );
END IF;
```

This is the **stage's `input_template`-rendered prompt** — e.g. for research-write's `synthesize` stage, it's the rendered template substituting `{{stage_results.gather.output}}` into the prompt. Typically 1-5 KB.

**Stability:** changes per stage. For multi-turn tool loops within a stage, NULL after the first call (the LLM's tool calls and responses are appended to the messages table rather than re-rendered).

## Tools array (separate from messages, but worth knowing)

`compose_tools(p_agent_family)` returns the tool-definition array (separate from `tool_calls` in messages — these are the DECLARATIONS):

```jsonc
"tools": [
  {
    "type": "function",
    "function": {
      "name": "fetch_url",
      "description": "Fetch a URL...",
      "parameters": { ... }   // jsonschema
    }
  },
  ...
]
```

**Typical size:** 2-15 KB depending on how many tools the agent has access to. Filtered by `stewards.tool_permission(agent_family, tool_name)`. For research agents with web tools + fs tools + study tools, this is ~10 KB of overhead **on every dispatch**.

Not yet a failure mode but worth a follow-up: if we add many more MCP tools, this grows.

## Summary — where compaction can and cannot go

| Zone | What | Compactable? | Why |
|---|---|---|---|
| 1: System (compose_system_prompt) | agent prompt + instructions + covenant + intent + skills | **NEVER** — prefix stability for caching; covenant must reach the agent verbatim |
| 2: History — system message at position 0 | identical to zone 1 (and shouldn't be duplicated, but check) | NEVER |
| 2: History — initial `user` (the binding question) | first user message in the session | **NEVER** — the agent must know what it's answering |
| 2: History — recent 3 turns (assistant+tool sequences) | rhythm preservation per LangChain | **NEVER** — emit raw |
| 2: History — error-traced tool results | any message containing error/traceback patterns | **NEVER** — per LangChain, helps avoid repeats |
| 2: History — older user clarifications | mid-session user messages | RARELY — Hermes/PicoClaw default: preserve |
| 2: History — older `tool` results with engrams extracted | the torso zone (engrams jsonb populated) | **YES — emit HOT engrams instead of raw** |
| 2: History — older `assistant` reasoning_content | reasoning models' thinking blocks | **YES — drop entirely from older turns** |
| 3: Current-turn user message | stage input_template rendered | NEVER |
| Tools array | function definitions | Separately — filter unused tools per agent rather than compact |

## How compressed messages render in compose_messages (ratified K design)

When a tool message has engrams extracted, `compose_messages` emits a synthetic block in place of the raw content:

```text
[Engrams from msg #4f2c, raw 426651 chars, 6 engrams extracted]

⚠️ Source content showed signs of prompt injection. Raw available via
   expand_message(id=4f2c, tier='raw', confirm_inspect_raw=true).
   (banner only present when injection_suspected=true)

## Pickard 1906 silicon-carbide detector
Greenleaf Whittier Pickard filed for a silicon-carbide detector patent
on August 30, 1906. The patent (US 836,531) described...
Sources: https://en.wikipedia.org/wiki/Crystal_radio
Quote: "this device, which I have termed the perikon detector..."

## Cat-whisker mechanism and AM rectification
[next HOT engram]

(more available via expand_message(id=4f2c, tier='medium'|'cold'|'raw'))
```

Per compressed message: ~1500 tokens of HOT engrams. MEDIUM and COLD stored but not emitted by default — retrievable on agent demand.

## The "what to send" decision, rephrased

Compaction is just `compose_messages()` choosing what to include. We change ONE function (and add a jsonb column to `stewards.messages` + an extractor trigger to populate it) and the entire substrate benefits. The bgworker doesn't need to change. The bridge doesn't need to change. Every existing pipeline benefits the moment the function ships.

That's the leverage point. K is a `compose_messages()` rewrite plus the engram extraction pipeline that feeds it, not a substrate rewrite.
