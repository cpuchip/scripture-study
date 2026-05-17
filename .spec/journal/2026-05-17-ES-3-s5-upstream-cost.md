---
date: 2026-05-17
mode: build + verify (Emergency Stop, ES.3.s5)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "ES.3.s5 — gateway upstream-cost capture; model-name normalization investigated and dropped"
status: ES.3.s5 SHIPPED + verified. The ES arc is fully complete. Soak running.
carry_forward:
  - "ES.6.A verdict-only review + ES.6.B failover — still optional carry-forward, user chose to skip unless more issues arise."
  - "upstream_micro_dollars now populates on every chat cost_event — a real estimate-vs-actual signal is available if cost discipline ever needs auditing."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
  - "../../projects/pg-ai-stewards/extension/es11-gateway-upstream-cost.sql"
---

# 2026-05-17 — ES.3.s5: the fix that corrected itself

Michael asked to do ES.3.s5 — model-name normalization — and I'd
specced it as fixing cost attribution and substitution detection.

## The correction

Before building, a code trace (the read-before-typing discipline)
proved my spec wrong on every functional claim:
- Cost: `cost_events` already records `requested_model` (canonical) —
  there's a comment in `bgworker.rs` saying exactly that.
- `l29` substitution: compares pipeline-declared vs `requested_model`,
  both canonical, at enqueue time — never touches the response model.
- Trust: `trust_scores` keys on model, but the value comes from
  work-item actor metadata (pipeline-declared), not the response.

The three gateway identifiers land in exactly one place —
`messages.model`, an audit/display field. Normalizing it is cosmetic.

I surfaced this to Michael rather than build the thing I'd talked him
into. He chose: drop normalization, do the genuinely valuable half.

This is the covenant working as designed — "checks existing work
before making new claims, surfaces tensions rather than building only
toward the thesis." I'd made claims without reading the code; reading
it, and saying so, was the recovery. The cost of getting it right was a
correction mid-stream; the cost of not would have been a cosmetic
migration shipped under a false premise.

## What shipped (`b82c9c4`)

The valuable half: capture the gateway's real cost. OpenCode Zen streams
`usage.cost_details.upstream_inference_cost` — the actual upstream
price. (The top-level `cost` field is 0 — OpenCode Go is subscription-
billed.) ES.6 streaming already pulled the whole `usage` object into
`result.response`; ES.3.s5 extracts the upstream cost into a new
`cost_events.upstream_micro_dollars` column, beside `micro_dollars`
(the substrate's rate×token estimate).

Smoke: a kimi-k2.6 chat recorded `upstream_micro_dollars=195` from
`cost_details 0.0001954` — and it matched the rate×token estimate
exactly. The substrate's cost math was already sound; now there's a
measured number to confirm it against.

## The ES arc — fully closed

ES.1 → ES.3 → ES.4 → ES.5 → ES.6 → ES.3.s5. The emergency stop is
entirely worked through. ~95 commits, zero rollbacks. The bacteriopolis
runaway that started it now runs clean to a verified artifact, with the
gateway's real cost recorded for every call.
