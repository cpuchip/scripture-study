---
date: 2026-05-15
mode: debug + build (Emergency Stop)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "ES — Emergency Stop: bacteriopolis runaway diagnosed (7 critical failures), 4 of 6 ES.1 bleed-stoppers shipped"
status: bleed structurally stopped — ES.1.s3 (Rust crash-breaker) + ES.1.s4 (health check) + ES.3 council remain
carry_forward:
  - "ES.1.s3 — bgworker crash-loop breaker. Per-kind consecutive-failure counter + cooldown. Needs bgworker.rs change + docker rebuild — fresh-context session."
  - "ES.1.s4 — embedding provider health check. Pre-flight check on LM Studio endpoint before enqueueing/claiming embed work."
  - "ES.3 — rearchitect compaction. The big one — council before building. The judge-compiled-brief model replaces leaf-chunk-and-embed. Validated by the Nate B Jones video. Likely deletes messages_raw_overflow_leaves entirely."
  - "ES.4 — verify: bacteriopolis re-run under the rearchitecture. Success metric: a web fetch costs a few calls, not hundreds."
  - "LM Studio nomic model (text-embedding-nomic-embed-text-v1.5) must be loaded after any host reboot — `lms load`. CF-4."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
  - "../../projects/pg-ai-stewards/extension/es1-work-item-cancel-cascade.sql"
  - "../../projects/pg-ai-stewards/extension/es2-embed-provider-route.sql"
  - "../../projects/pg-ai-stewards/extension/es3-chunk-index-circuit-breaker.sql"
  - "../../projects/pg-ai-stewards/extension/es4-disable-leaf-embed-enqueue.sql"
---

# 2026-05-15 — ES, Emergency Stop

The bacteriopolis L.1.1.x fix-bundle retry produced a runaway: DeepSeek
churn, a bgworker crash loop, ~230M wasted input tokens. Michael caught it
("we're still bleeding"), and asked for a full stop, diagnosis, and plan.

## The incident

A **cancelled work_item kept running its chat loop.** Cancelling a
work_item flips `work_items.status` but the chat→tool_dispatch→chat loop
runs on `session_id` and never checked work_item status. So the cancelled
fixed-retry-1 kept looping; each oversized web fetch tripped L.1.1.8 →
`chunk_and_index` fired 160-501 contextualizer chats → embed jobs hit a
404 (wrong provider) → bgworker crashed on `bigint = text` → restarted →
picked up the next failed embed → crashed again. Tight loop.

## Seven critical failures (traced — Agans Rule 3, looked not theorized)

- **CF-1** cancelled work_item doesn't stop its session's chat loop
- **CF-2** bgworker embed handler `WHERE id=$1` (text) crashes on
  `messages_raw_overflow_leaves`' bigserial id
- **CF-3** no circuit breaker on the bgworker crash loop
- **CF-4** embeddings = local LM Studio nomic model, not OpenCode Go; host
  reboot dropped the model; no health check
- **CF-5** chunk_and_index had no leaf ceiling — 900K fetch → 501 chats
- **CF-6** (architectural) the leaf-chunk-and-embed approach is chatbot-era
  RAG applied to an agent-era job — validated by the Nate B Jones video
  ("the retrieval unit must match the work"; Page Index: structured docs
  shouldn't be chunked; "better embeddings don't fix this")
- **CF-7** embed jobs misrouted to provider `opencode_go` (no embeddings
  endpoint) instead of `lm_studio`

## What shipped (ES.1, 4 of 6)

| Commit | Fix |
|---|---|
| `b6ac127` | ES.1.s1 — work_item_cancel cascade: hard-stops every non-terminal work_queue row for the work_item's sessions |
| `149a783` | ES.1.s5 — embed provider routing: BEFORE INSERT trigger forces kind=embed → lm_studio |
| `61f56d1` | ES.1.s2 — chunk_and_index 40-leaf circuit breaker; over it, refuse + leave raw |
| `60cd8d2` | ES.2/CF-2 Option B — removed leaf embed enqueue (avoids the bigint=text cascade ES.3 may discard) |

**Bleed structurally stopped.** The crash loop cannot recur: no 500-leaf
explosions, no misrouted embeds, no leaf-embed crash, no runaway from
cancelled work_items.

## Honest moments

- **CF-2 ratification was wrong.** Michael ratified "fix now — contained
  one-migration change." Tracing showed the bigserial→text change cascades
  through 4 functions. Surfaced it; he chose Option B (disable leaf embed
  enqueue — genuinely contained). Covenant's flag-when-wrong working as
  intended, in both directions.
- **L24/L25 earlier the same arc**: dropped a function thinking it a
  duplicate, broke the bgworker, restored it. Two overcorrections caught
  this session. Both logged. The pattern: trace callers before DROPing or
  assuming "contained."

## The deeper finding (CF-6)

Michael's instinct, the video, and the trace all converge: **we built the
wrong abstraction.** 500 vector leaves from one web fetch is chatbot-era
RAG. What an agent on a mission needs is a judge that reads the page once
WITH the binding question, discards noise, and returns a small compiled
brief — few calls, few memories. That is the Judges pattern (named in
.mind/principles.md two days ago) applied to compaction. ES.3 is where
that gets designed — council first.

## Carry-forward

ES.1.s3 (Rust crash-breaker, needs rebuild), ES.1.s4 (health check), then
ES.3 council (the rearchitecture), then ES.4 verify. The substrate is
safe to leave: queue empty, soak paused, LM Studio healthy, nothing
running.
