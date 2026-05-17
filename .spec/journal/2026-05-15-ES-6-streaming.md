---
date: 2026-05-15
mode: research + build + verify (Emergency Stop, ES.6)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "ES.6 streaming chat dispatch — fixes the ~125s gateway idle-timeout; ES.4 verified, soak resumed, the ES arc complete"
status: ES.6 SHIPPED + verified. ES.4 run-3 reached completed/verified. Soak RESUMED. The full ES arc is closed.
carry_forward:
  - "ES.6.A (verdict-only review) — optional design cleanup; review regurgitating ~2000 words is wasteful even though streaming now makes it work. Carry-forward, not urgent."
  - "ES.6.B (gateway-500 failover) — optional general resilience for genuinely transient gateway errors. Carry-forward."
  - "ES.3.s5 model-name normalization — still deferred, optional."
  - "ES.6 may have exposed the OpenCode `cost` stream event — worth checking whether cost_usd now populates."
  - "Soak is RUNNING again — watch the first autonomous cycles."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
  - "../../projects/pg-ai-stewards/extension/src/bgworker.rs"
---

# 2026-05-15 — ES.6 streaming: the last wall

Michael's call after ES.4 failed twice at `review`: "research and
council first before going back to build." That instruction is why
this worked — the research changed the answer.

## What the research found

ES.4's review failures looked like "qwen3.6-plus is slow / the review
stage is wasteful." Researching OpenCode's docs (Go, Server, Network,
Troubleshooting, Zen — none document a timeout) and the substrate code
showed the real shape: the substrate's own chat HTTP timeout is 600s,
so the ~125s ceiling is entirely gateway-side, and the substrate makes
**non-streaming** chat requests.

Then the decisive empirical test against `opencode_go` (Agans Rule 9):
- non-streaming long generation → HTTP 500 at **125.2s** (reproduced)
- streaming, same prompt → HTTP 200 at **185.8s**, clean `[DONE]`

The ~125s ceiling is a **gateway idle-timeout**. A non-streaming
request sends no bytes during generation; a proxy in front of OpenCode
Zen kills the idle connection. Streaming keeps tokens flowing.

## The fix collapsed to one root cause

The first plan had three patches (verdict-only review, failover, faster
model). The research collapsed them: the fix is **streaming**, and it
fixes the timeout substrate-wide — review, synthesize, AND the ES.3
judge (a judge on a near-1M-token document is the same long-generation
risk). A and B demoted to optional carry-forward.

## What shipped — ES.6 (`5ca7580`)

`chat()` in `bgworker.rs` now sends `stream:true` and parses the SSE
event stream (`parse_chat_sse`), reassembling it into the standard
non-streaming response shape — so every downstream field extraction,
and the SQL apply handlers that re-parse `result.response`, are
unchanged. tool_call deltas accumulate by index; content and
reasoning_content concatenate; usage comes from the tail chunk. The
non-streaming-shape reassembly kept the blast radius tiny.

## Verification — ES.4 run-3

completed / verified, **$0.33**. gather (streaming tool-calls — the
risky assembly), synthesize, and the `review` stage all passed. The
review chat ran **169 seconds** — 44s past the ceiling that 500'd runs
1 and 2 — and completed. The bacteriopolis exhibit materialized.

Agans Rule 9 fully closed: reproduce → fix → confirm gone in a live
pipeline to verified.

## The arc, closed

ES.1 (bleed-stoppers) → ES.3 (judge-compiled-brief) → ES.4 (verified
live) → ES.5 (fs_search ctx, PDF extraction, consult grant) → ES.6
(streaming). The bacteriopolis runaway that started the Emergency Stop
now runs clean to a verified artifact for ~$0.33. The soak is resumed.
~90 commits across the arc, zero rollbacks.

The discipline that held it: reproduce before claiming a cause, council
before building when the diagnosis is uncertain, gate every step with a
commit, test the fix against the real failure. Michael's "research and
council first" was the hinge — without it ES.6 would have been three
half-fixes instead of one root fix.
