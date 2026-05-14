---
batch: K
title: Engram-Based Progressive Context Disclosure
status: ratified
proposed_by: michael
proposed_on: 2026-05-13
ratified_on: 2026-05-13
preceded_by:
  - substrate-batch-j-fanout-brainstorm.md
links:
  - "../scratch/batch-k-context-management.md (full design doc)"
  - "../scratch/batch-k-payload-structure.md (what gets sent to LLM each turn)"
  - "../scratch/batch-k-research-compaction-and-subagents.md (prior art)"
  - "../scratch/brainstorm-context-management-candidates.md (J.4 self-brainstorm)"
binding_question: |
  When a multi-turn agent session accumulates large tool results (a single
  fetch_url body of 426K chars, or many medium results piling up), the
  substrate replays the entire session history every dispatch, eventually
  exceeding the model's input limit. Solve it without losing source-
  verification, without dropping cite chains, and resilient against prompt
  injection in untrusted external content.
---

# Batch K — Engram-Based Progressive Context Disclosure

## Why this exists

J.3 (science-center exhibits fanout) shipped 2 of 6 children today. The other 4 failed with the same shape — Moonshot's Kimi K2.6 rejecting the gather chat call because the request body exceeded 262K input tokens (374K+ requested in the worst case).

**The smoking gun**: one `tool` message in the worst-failed session (`wi--b0b1185f--gather`) held **426,651 characters** (~142K tokens) of fetched content. A single fetch_url body poisoned the context for every subsequent turn. The current `compose_messages()` replays the entire session history each turn with no truncation, no summarization, no size guard.

This is not Kimi-specific. The same pattern would 429 GPT-4o, blow Anthropic's 200K window, and waste cache on Gemini. **The substrate has no concept of context discipline. Batch K adds it.**

## The three shapes (ratified after 5 rounds of council)

```
REACTIVE: engram extraction
  triggered: post-fetch automatic at >60K chars
  actor: deepseek-v4-flash (1M context, structured output)
  output: jsonb array of HOT/MEDIUM/COLD engrams on the message
  effect: compose_messages emits engrams instead of raw content
  retrieval: expand_message(id, tier='hot|medium|cold|raw')

PROACTIVE: sub-agent delegation
  triggered: heavyweight tool definitions (deep_research, audit_files, etc.)
  actor: qwen3.6-plus (orchestrator) + scoped tools_subset
  output: prose digest (HOT engrams rendered) to parent + structured engrams stored
  effect: verbose work isolated in child context; parent only sees digest

SECURITY: injection defense
  triggered: at engram extraction time + light screen on web_search
  actor: same extraction model with explicit "data not instructions" prompt
  output: injection_suspected boolean + evidence string on the engrams
  effect: banner in compose_messages output (L1 in v1; L2/L3 deferred)
```

Engrams handle the "single huge message" case AND the "many medium messages accumulating" case (via head/torso/tail in `compose_messages`). Sub-agents prevent verbose work from ever entering the parent context. Injection defense is a property of both paths — every untrusted external content passes through the engram pipeline's data-not-instructions filter.

## Ratified decisions

### Architecture (rounds 1-5)

- **Sync sub-agents**: parent's tool call blocks until child verifies
- **Explicit sub-agent triggering**: heavyweight tools declare themselves (no auto-promotion)
- **Implicit/size-based engram extraction**: post-fetch automatic at 60K chars
- **Document-intrinsic engrams**: one extraction per message, not re-extracted per stage
- **Sub-agent return shape**: prose digest to parent + structured engrams stored on child's last assistant message
- **Tier sizes**: HOT 1500 / MEDIUM 500 / COLD 100 tokens
- **Multiple engrams per document**: jsonb array of items (not single nested object)
- **DeepSeek V4 Flash for engram extraction** (1M context, structured output, cheapest tier on OpenCode Go)
- **Qwen3.6 Plus for sub-agent orchestration**
- **Strict structured output enforcement** via DeepSeek V4 Flash's JSON schema mode
- **Injection L1 in v1**: banner in compose_messages output (L2/L3 deferred to v2)
- **Sub-agent tool subsets enumerated per type** (no full-substrate access)
- **web_search passes through with lightweight injection regex screen**; fetch_url full pipeline

### Phase-level (this proposal's ratification)

- **Trigger threshold**: 60K chars (~20K tokens) — matches LangChain Deep Agents
- **Sync at insert**: bridge blocks during extraction (1-3s); no race conditions
- **Engram caps**: token-budget-only (fit within HOT 1500 etc.); no hard count cap
- **expand_message default tier**: all three (HOT + MEDIUM + COLD joined)
- **Sub-agent failure handling**: return error to parent; parent decides
- **Sub-agent depth cap**: 2 levels (parent → sub-agent → maybe one more)
- **Heavyweight tools (7)**: deep_research, summarize_url, audit_files, investigate_session, summarize_study, investigate_study, audit_studies
- **Heavyweight tool batching**: K.5a external content + K.5b substrate-internal-study tools (smoke each set independently)
- **Phase order**: Linear K.1 → K.2 → K.3 → K.4 → K.5a → K.5b → K.6 → K.7
- **Schedule**: start K.1 immediately after ratification

## Phases

Eight sub-phases (K.5 split into K.5a + K.5b). Each commitable in one pulse.

| Sub-phase | What | LLM cost | Pulses |
|---|---|---|---|
| **K.1** | `stewards.messages.engrams jsonb` column. `extract_engrams(message_id)` SQL fn. INSERT trigger for tool messages >60K chars. Block-at-insert via the bgworker chat dispatch path. | ~$0.005 smoke | 1 |
| **K.2** | `compose_messages` rewrite: head/torso/tail logic, engram emission for compressed tool messages, recent-3-turns raw, error-trace bypass, reasoning_content drop from older turns. Backwards-compatible (NULL engrams → raw, unchanged). | $0 | 1 |
| **K.3** | `expand_message(id, tier?, engram_id?, confirm_inspect_raw?)` MCP tool. Bridge endpoint. Tool permission grants. | $0 smoke | 1 |
| **K.4** | `spawn_subagent(agent_family, binding_question, tools_subset?)` MCP tool. Sync-wait pattern. Sub-agent failure semantics. Depth-2 cap. | ~$0.10 smoke | 1 |
| **K.5a** | External-content heavyweight tools: `deep_research`, `summarize_url`, `audit_files`, `investigate_session`. Each wraps K.4 with specific tools_subset. | ~$0.50 smoke | 1 |
| **K.5b** | Substrate-internal-study heavyweight tools: `summarize_study`, `investigate_study`, `audit_studies`. Same pattern, internal corpus. | ~$0.50 smoke | 1 |
| **K.6** | Injection defense L1: engram extractor detection (built into K.1's prompt), banner in compose_messages, `web_search` light regex screen. Tool capability audit (fetch_url, expand_message only do what they say). | ~$0.005/web_search | 1 |
| **K.7** | Replay J.3 failed children (Crystal Radio, Bacteriopolis, CS Unplugged, Indicating Electrolysis) with K live. Verify 2-3 now produce briefs. Measure cost reduction. | ~$2-3 LLM | 1 |

**Total estimated cost across K.1-K.7: ~$5-10.** Budget for the arc: $15.
**Total estimated pulses**: 8 (one per sub-phase). Could collapse some into combined pulses but the C-F discipline says keep them separate.

## What we already have (J.2 reusable)

K reuses substantial Batch J infrastructure:

- `work_item_create` + `work_item_dispatch_stage` (sub-agent spawn primitives)
- `parent_work_item_id` on `work_items` (relationship tracking; J.2)
- `on_maturity_verified` + sibling-count trigger machinery (J.2 + j7 fix)
- `pending_file_writes` (G.4) + materialize-writes (i7) for any K output that lands as a file
- `agent-proposal` claude-attest gate pattern (i6) — adaptable for sub-agent permission elevation if needed

The new building blocks K adds:
- `extract_engrams(message_id)` SQL function + INSERT trigger
- `expand_message` MCP tool
- `spawn_subagent` MCP tool + 7 wrapper tools
- `compose_messages` rewrite (changes semantics for compressed messages)

## Adjacent surface audit

**Scope.** Engram extraction + sub-agent delegation apply to every existing pipeline immediately. study-write, research-write, planning, brainstorm-* — all benefit on the next dispatch.

**Discoverability.** Pipeline behavior change (compose_messages emits engrams instead of raw) is invisible to the dispatching agent unless K.1's engram block header is unfamiliar — agent system prompts should mention `expand_message` as the affordance. CLAUDE.md and the new agent-onboarding docs need an entry.

**Contracts.** The `messages.engrams` column is additive. `compose_messages` falls back to raw when `engrams IS NULL` — fully backwards compatible. No work_item migrations needed.

**Spec gaps.** What happens when a sub-agent's own session hits engram-extraction territory? Answer: engram extraction is recursive — a sub-agent's tool result over 60K chars also gets engram-extracted. The sub-agent then sees engrams in its own context. Compaction protects all levels of the agent hierarchy.

## Open questions (deferred to implementation smoke)

- **Extractor model fallback**: if DeepSeek V4 Flash unavailable, fall back to Qwen3.6 Plus or fail with raw passthrough + warning?
- **Parallel tool_calls in one assistant turn**: if 5 parallel tools all return >60K chars, all 5 extract independently. ~$0.025 for that turn — acceptable.
- **User pasting a 100KB document**: technically a user message, currently never compressed. Worth noting as future extension.
- **Reasoning_content (qwen-plus, deepseek-r1 `<think>` blocks)**: default K behavior is drop from older turns. Engram-extraction of reasoning is a v2 if useful.

## v2 / deferred items

- **Graduated resolution under pressure** — pressure-aware tier selection (50% → all tiers; 70% → HOT+COLD; 85% → HOT truncated; crisis → COLD only). v1 always emits HOT; catches the 426KB case.
- **Marked-important anchoring** — `is_important: bool` flag keeping engrams at HOT regardless of pressure.
- **Cross-message engram search** — `search_engrams(session_id, query)` vector search. Foundation for Batch L cross-session memory.
- **Injection defense L3 (source blocklist)** — domains with confirmed injection require human approval. Needs UI; defer.

## Verification target

K is done when:
1. Sessions with 426KB+ tool results no longer fail on subsequent dispatches.
2. Engram extractor preserves URLs/dates/names/quotes verbatim. Diff `preserved` fields against original; zero loss tolerated.
3. `expand_message(id, tier='raw')` returns byte-identical content.
4. `spawn_subagent` round-trips: parent calls `deep_research`, waits, gets digest, can expand child's engrams.
5. Injection-shaped content sets `injection_suspected=true` and renders the banner.
6. K.7 retest: at least 2-3 of the 4 failed J.3 children now produce verified briefs.

## What this enables next

- **Tuesday Science Center day**: J.3 retries become feasible. The exhibit library can fill in.
- **Batch L** (next): cross-session memory + the search_engrams primitive. Already laid out in v2 deferrals.
- **Long-horizon agents**: the substrate now supports sessions of arbitrary length (within reason). 50-turn deep-research sessions become viable.
- **Council quality**: Phase F council members can be sub-agents (restricted tools_subset), each returning engram-shaped contributions to the synthesizer.
