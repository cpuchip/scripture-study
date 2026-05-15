---
name: substrate-batch-l-1-1-17-soft-cap-judge-pattern
title: Batch L.1.1.17 — Soft cap (Judge pattern) on tool rounds; hard cap as safety net
status: draft
created: 2026-05-14
ratifies_from: 2026-05-14 bacteriopolis retry-2 finding — capped rounds were genuine due-diligence (fs_read existing exhibit, fs_list sibling formats, study_search prior thinking)
---

# Batch L.1.1.17 — Soft cap follows the Judge pattern

## What changed our minds

L.1.1.16 capped `max_tool_rounds` and forced `tools_disabled=true` + `tool_choice=none` at the threshold. Bacteriopolis retry-2 hit the cap at round 5 of context_gather. The "violations" of the cap (3 more capped rounds where the agent kept emitting tool_calls) were inspected — and they were:

| Capped round | Tool | What the agent was trying to do |
|---|---|---|
| 6 | `fs_read` | Check if a prior bacteriopolis exhibit doc already exists |
| 7 | `fs_list` | List sibling exhibit docs (learn the format convention) |
| 8 | `study_search` | Search substrate studies for prior bacteriopolis thinking |

This isn't insubordination. It's exhausting internal references before turning external — exactly what a steward of the binding question should do.

The L.1.1 architecture names agents as **judges within stewardship** (Exodus 18 pattern, principles.md). But L.1.1.16's hard cap treats them as executors. That's a mismatch.

## Architectural fix

Two-tier cap:

| Tier | Field | Default | Behavior |
|---|---|---|---|
| **Soft** | `max_tool_rounds` (existing) | 5 | At threshold, INJECT a system message: "You've made N tool calls in this stage. Each additional call costs [agent's-judgment-of-value]. If you can answer the binding question now, finalize. Otherwise, the NEXT tool call should justify itself in your response." Tools remain available. |
| **Hard** | `max_tool_rounds_hard` (NEW) | 50 | At threshold, force `tools_disabled=true` + `tool_choice=none`. Cost safety net only. |

Per Michael (2026-05-14): "I really want to see what the system can produce at this early stage, and only limit it to protect costs." The OpenCode Go subscription cap is the real financial safety net; substrate caps are about behavior shaping, not cost protection.

## Implementation

1. Rename current behavior: `max_tool_rounds` becomes the soft threshold. Hard threshold is a new field `max_tool_rounds_hard` on `pipelines.stages[]`.
2. `chat_post_internal` checks both:
   - `v_rounds_so_far >= max_tool_rounds_hard` → strip tools, set tool_choice=none (current L.1.1.16 behavior, escalated to the hard threshold)
   - `v_rounds_so_far >= max_tool_rounds` (soft) AND not yet hard-capped → inject a soft-cap notice as a system message in the session BEFORE dry_run_chat composes
3. Soft-cap notice content (template):

```
[STEWARD NOTICE — soft cap reached]
You've used {{rounds_so_far}} tool calls in the {{stage_name}} stage.
The soft cap for this stage is {{soft_cap}}; the hard cap (where tools
will be removed) is {{hard_cap}}.

If you can answer the binding question now from what you've gathered,
finalize your response. If you genuinely need another tool call,
include a one-sentence justification in your next response so future
review can audit the decision.
```

4. Update research-write defaults: soft 5/5/3/1, hard 50/50/15/3 (stage-specific reasonable maxes).
5. Soft-cap notice gets logged via the existing `_*` marker pattern so we can audit later: `_soft_cap_injected_at_round` on the chat payload.

## Tradeoffs

- **Honors agency** — the agent decides whether the next tool call is worth it
- **Surfaces decisions** — agent must justify continuation in its response, creating an audit trail
- **Costs more on average** than a hard cap because soft-capped runs may use a few extra rounds
- **Real safety net at hard cap** — even an aberrant agent can't go past 50 rounds (and the OpenCode Go subscription is the wallet-level cap)

## Carry-forward

- Add a `soft_cap_injections` summary view so we can see how often agents respect the soft cap vs blow through it
- If a particular agent family blows through repeatedly, that's signal — either prompt revision or model swap
- Long term: `max_tool_rounds` could be tuned per-agent (not just per-stage) for compulsive-tool-callers vs naturally-terse models

## Connection to the Judges principle

The captains-of-tens pattern requires **trusting judgment within stewardship**. L.1.1.16's hard cap removed the trust. L.1.1.17 restores it: surface the situation, let the agent judge, only override when cost protection is genuinely necessary. The `max_tool_rounds_hard` is the "every great matter they shall bring unto thee" line — the hard ceiling Moses kept for cases the captains couldn't handle.

Also pairs with the **go-and-do, return-and-report** pattern (1 Nephi 3:7 → 4:6): the steward goes with their authority, gathers what they need, returns the digest. The substrate's job is to give them a budget for the going, not to interrupt them mid-mission for being thorough.
