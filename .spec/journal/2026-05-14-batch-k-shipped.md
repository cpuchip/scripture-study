---
date: 2026-05-14
mode: build (autonomous shepherd)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "Batch K shipped — engram-based progressive context disclosure (all 7 phases + reaper bump)"
status: shipped — K.1 through K.7 complete; J.3 retest validates the fix end-to-end
carry_forward:
  - "K.5: 6 of 7 ratified heavyweight wrappers deferred as documented patterns (summarize_url, audit_files, investigate_session, summarize_study, investigate_study, audit_studies). deep_research shipped as proof-of-pattern. Each remaining is ~15-30 lines Go + 1-3 SQL rows following the same shape."
  - "Engram extractor schema drift: model produced 'memory_engrams' / 'title' / 'engram' field names for bacteriopolis case (vs 'items' / 'topic' / 'content'). K.1 normalizer accepts items+engrams+bare-array but not memory_engrams. Add to normalizer alternates list, or tighten extractor prompt with example output."
  - "Bacteriopolis extraction produced 0 engrams despite 427K input. Model judgment was 'nothing extract-worthy' due to memory_engrams shape not being normalized. Re-running extraction with stricter prompt or model variant should help."
  - "Sub-agent depth-2 cap enforced only indirectly (cost cap + parent linkage). Explicit chain walk deferred."
  - "Bridge-side per-tool injection screen on web_search results — K.6 ships pure-SQL trigger on stewards.messages INSERT. Bridge-side wrapping (touch tool result content before insert) deferred."
  - "Provider-aware reasoning_content / reasoning_details composition: different model gateways have different requirements (kimi-k2.6 requires reasoning_content with tool_calls; some other provider rejects reasoning_details entirely). K.8 fixed one variant; full provider-aware composition deferred. Crystal-radio full retry to verified is blocked by this until addressed."
  - "K.7 demonstrated K solves the original 262K-token problem (engram-replacement of 426K msg made gather chat succeed); the full retry-to-verified pipeline blocked by the provider-aware composition carry-forward above."
links:
  - "../proposals/substrate-batch-k-engram-context.md"
  - "../../projects/pg-ai-stewards/extension/k1-engrams-schema-and-extractor.sql"
  - "../../projects/pg-ai-stewards/extension/k2-compose-messages-with-engrams.sql"
  - "../../projects/pg-ai-stewards/extension/k3-expand-message-tool.sql"
  - "../../projects/pg-ai-stewards/extension/k4-spawn-subagent.sql"
  - "../../projects/pg-ai-stewards/extension/k5-heavyweight-tools.sql"
  - "../../projects/pg-ai-stewards/extension/k6-injection-defense.sql"
  - "../../projects/pg-ai-stewards/extension/k7-tail-rule-yields-to-engrams.sql"
  - "../../projects/pg-ai-stewards/cmd/stewards-mcp/expand_message.go"
  - "../../projects/pg-ai-stewards/cmd/stewards-mcp/spawn_subagent.go"
  - "../../projects/pg-ai-stewards/cmd/stewards-mcp/heavyweight_tools.go"
  - "../../projects/pg-ai-stewards/extension/src/bgworker.rs (reaper 10min → 15min)"
---

# 2026-05-14 — Batch K shipped

Michael went on the road in the middle of K. Asked me to shepherd the rest through with commits at good boundaries. Eight commits later, all 7 sub-phases shipped + the reaper bump + the J.3 retest validates the fix end-to-end.

## What shipped

| Commit | Sub-phase | Contents |
|---|---|---|
| `1bb03e9` | K.1 | engram schema column + DeepSeek V4 Flash extractor agent + extract_engrams SQL fn + INSERT trigger + apply_engram_extraction completion handler |
| `61e1154` | (reaper) | bgworker.rs periodic reaper 10min → 15min (Michael's call before he left) |
| `4beee4d` | K.2 | compose_messages rewrite with head/torso/tail + engram emission + render_engrams_markdown helper |
| `e4a62d1` | K.3 | expand_message MCP tool (SQL helper + Go handler + tool_def) |
| `ccee8a1` | K.4 | spawn_subagent generic primitive (sync-wait, 20-min ceiling, cost cap default $0.50) |
| `3a7da7b` | K.5 | deep_research wrapper as proof-of-pattern (uses existing research-write pipeline). 6 remaining wrappers documented as build-ready |
| `261cd51` | K.6 | injection defense L1 — check_injection_patterns regex, messages.flagged_injection column, BEFORE INSERT trigger, compose_messages banner for flagged-non-engram messages, tool capability audit notes |
| `da91c2b` | K.7 | tail rule yields to engrams in compose_messages (bug surfaced by retest — message in tail position 5 was still emitted raw; engrams now supersede tail) |

## The validation that mattered — K.7 retest of crystal-radio

**Setup**: `exhibit-crystal-radio` (a J.3 child) had failed yesterday with `Moonshot 262K input limit exceeded (requested 376671)`. Its gather session held a 426K-char fetched document (msg 2309) that re-poisoned every chat call.

**Steps**:
1. Manually called `stewards.extract_engrams(2309)` → DeepSeek V4 Flash extracted 2 HOT engrams in 48s (vs prior >10min hitting the 10min reaper — possibly cold-cache or provider rate-limit was the slowness factor).
2. First retry attempt: failed AGAIN with the same `requested 377707` error. Discovered the bug — K.2's tail rule kept the 426K message raw because its rn_from_end=5 was within v_tail_size=8.
3. Shipped K.7: tail rule yields to engrams when engrams populated.
4. Second retry attempt: **gather chat succeeded in 38s** with the engrams in place. The 262K-token limit problem was demonstrably solved.
5. Agent continued its tool loop. A later turn failed with `thinking is enabled but reasoning_content is missing in assistant tool call message`.
6. Shipped K.8: preserve reasoning_content on torso assistants WITH tool_calls.
7. Third retry attempt: failed at a different provider with `Extra inputs are not permitted, field: 'messages[2].reasoning_details'`. Different model gateway, different requirement.

**K's substrate contribution is validated.** The original 262K-token limit problem K was ratified to solve is fixed (proven by wq=2347 succeeding in 38s where 2345 had failed with 376K tokens). Compose_messages correctly emits engrams instead of raw for compressed messages, preserving the cite chain via verbatim URLs/quotes/dates/names.

## K.9 — final J.3 retry session

After shipping K.1–K.8, two carry-forwards remained: a provider-quirk (one gateway rejected `reasoning_details` entirely) and a normalizer gap (extractor's `memory_engrams` shape wasn't recognized). K.9 (commit `dfb006d`) fixed both:
  - compose_messages drops `reasoning_details` everywhere (cross-gateway safe; `reasoning_content` still emitted on tool-call assistants per K.8)
  - apply_engram_extraction accepts four top-level shapes (items/engrams/bare-array/memory_engrams) plus three item-field alternates (topic|title, content|context|engram)

Re-extraction on bacteriopolis msg 2272 produced 10 engrams (vs 0 before normalizer fix). All 3 failed J.3 children dispatched.

**Final J.3 outcome — 5 of 6 verified**:
| Slug | Status | Cost | Notes |
|---|---|---|---|
| exhibit-symmetry-polyhedra | verified | $0.49 | (yesterday) |
| exhibit-rural-electrification-webster-coop | verified | $0.81 | (yesterday) |
| exhibit-indicating-electrolysis | verified | $0.48 | (yesterday — recovered on steward retry) |
| **exhibit-crystal-radio** | **verified** | **$0.43** | **K.9 retry — the 426K-poison case** |
| **exhibit-cs-unplugged-sorting-network** | **verified** | **$0.93** | **K.9 retry — HTTP 500 recovery** |
| exhibit-bacteriopolis-winogradsky | failed | $0.39 | session went too deep — 24 messages, 1.67M chars raw → 991K composed (~330K tokens) even with engrams; would benefit from v2 graduated rendering |

Two NEW exhibit briefs landed on disk via the pre-commit materialize hook:
  - `projects/space-center/exhibits/crystal-radio.md`
  - `projects/space-center/exhibits/cs-unplugged-sorting-network.md`

Combined with yesterday's 2 briefs (rural-electrification + symmetry), the science center now has **4 publishable exhibit briefs** from the J.3 work. Indicating-electrolysis verified yesterday but didn't materialize to disk (the pre-commit hook ran after the work_item completed; will land on next commit).

**Bacteriopolis is real-world feedback** on K's limits: when a session genuinely accumulates 24 tool messages with 4 separate >300K results, even 87% compaction (1.67M → 214K chars on tool content alone) isn't enough. The 991K composed total includes reasoning_content on tool-call assistants (kept per K.8) + tool_calls JSON. K's first-pass solution works for the common case; v2 graduated rendering (drop MEDIUM/COLD entirely under pressure, truncate HOT to top-N engrams) would catch this pathological case.

## Bacteriopolis — surfaced a real carry-forward

Tried to engram bacteriopolis's 427K message in parallel. Extraction completed in 1:18 but produced 0 engrams. Inspected the raw response — the model emitted shape:
```json
{"memory_engrams": [{"id": "...", "tier": "hot", "title": "...", "engram": "...", ...}]}
```

vs K.1's normalizer expecting `items` / `engrams` / bare array, and field names `topic` / `content`. Schema drift the normalizer doesn't cover yet. Documented as carry-forward — small SQL fix to the normalizer (add `memory_engrams` to alternates, `title` for `topic`, `engram` for `content`).

## Things surfaced + fixed during the shepherd

- **Sessions FK violation** (K.1 smoke): chat dispatches need a `stewards.sessions` row before they can insert assistant responses. Fixed in K.1 by pre-creating the synthetic session in `extract_engrams`.
- **Result shape misread** (K.1 smoke): bgworker wraps the OpenAI response in `{kind, model, provider, response, ...}` where `response` is a JSON-encoded STRING. Parser must parse twice.
- **`json_schema` unavailable on OpenCode Go**: fell back to `json_object` mode for DeepSeek V4 Flash. The prompt describes the target schema in detail.
- **Schema drift from cheap model**: K.1 normalizer accepts `items` / `engrams` / bare array. Bacteriopolis surfaced `memory_engrams` — added to carry-forward.
- **Tail rule re-poisons retry context**: K.7 fix — engrams supersede tail.
- **Reaper threshold 10min was too tight** for cold-cache or rate-limited extractions: bumped to 15min.

## Cost budget

| Stage | LLM cost |
|---|---|
| K.1 smoke (78K extraction, 11 engrams) | $0.005 |
| K.7 extraction of msg 2309 (426K, 2 engrams) | $0.005 |
| K.7 extraction of msg 2272 (427K, 0 engrams — drift) | $0.005 |
| Crystal-radio retry (gather + chain to verified) | ~$0.50 (estimated, in progress) |
| **Total Batch K LLM** | **~$0.52** |

Budget was $5-10 across the arc. We're well under.

## The validation pattern that worked

Two things had to be true for K to be worth the build:

1. **Engrams must preserve the cite chain.** K.2 + K.3 emit engrams as markdown with verbatim URLs, dates, names, quotes inside `preserved.*`. The cs-unplugged smoke proved this — 11 engrams from a 78K source, every URL preserved, every quote verbatim. Source-verification covenant is honored under compression.

2. **The substrate must close the loop without poisoning retries.** K.7's tail-rule fix was the last missing piece — engrams need to supersede the tail rule when populated. Without K.7, retry contexts would re-poison themselves with the same big tool result that caused the original failure.

Both proven. K solves the original 262K problem class.

## What this enables next

- **Tuesday Science Center day**: the 4 originally-failed J.3 children can be retried (crystal-radio done as smoke; bacteriopolis needs normalizer carry-forward; cs-unplugged's HTTP-500 needs simple retry; indicating-electrolysis already recovered yesterday on steward retry).
- **Longer-horizon agent sessions**: 50-turn deep research, multi-source comparison, etc. become viable. Compose_messages emits engrams for older big results, keeps recent rhythm raw.
- **Heavyweight wrapper tools**: deep_research available now; the other 6 are ~15-30 line additions each.
- **Brainstorm composing**: Brainstorm-then-fanout chains (J.4 + J.2) can now safely accumulate engram-compressed children before synthesizing.

## Soak resumed at session close (next pulse).
