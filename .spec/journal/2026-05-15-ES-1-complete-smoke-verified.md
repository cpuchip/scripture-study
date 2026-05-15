---
date: 2026-05-15
mode: build + verify (Emergency Stop, ES.1 close)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "ES.1 complete — all bleed-stoppers + crash-loop breaker shipped; clean pipeline smoke test verified the substrate stable"
status: ES.1 COMPLETE + verified — ES.3 (rearchitecture council) remains
carry_forward:
  - "ES.3 — rearchitect compaction. The judge-compiled-brief model replaces leaf-chunk-and-embed. Council before building. Validated by the Nate B Jones video + Michael's instinct. Likely deletes messages_raw_overflow_leaves."
  - "ES.4 — verify the rearchitecture: bacteriopolis re-run; success = a web fetch costs a few calls not hundreds."
  - "Model-name normalization (low priority): kimi-k2.6 reported under 3 gateway identifiers. Canonical mapping would fix cost attribution + substitution detection. ES.3-era cleanup."
  - "Dedicated retry-on-transient for embed jobs would be nicer than the crash-breaker's coarse pause. Carry-forward."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
  - "../../projects/pg-ai-stewards/extension/es5-kind-circuit-breaker.sql"
  - "../../projects/pg-ai-stewards/extension/src/bgworker.rs"
---

# 2026-05-15 — ES.1 complete, smoke-verified

Continuation of the Emergency Stop. Michael's call: "keep going until we
have something thoughtful and stable" — and specifically, ship the
remaining guardrails (1-3) and then test the pipeline.

## What shipped this session

| Commit | Fix |
|---|---|
| `b6ac127` | ES.1.s1 — work_item_cancel cascade (hard-stop the session chat loop) |
| `149a783` | ES.1.s5 — embed provider routing trigger (kind=embed → lm_studio) |
| `61f56d1` | ES.1.s2 — chunk_and_index 40-leaf circuit breaker |
| `60cd8d2` | ES.2/CF-2 Option B — disable leaf embed enqueue |
| `0dcdf75` | ES.1.s3 — bgworker crash-loop circuit breaker (kind_circuit_breaker table + 3 SQL fns + 3 bgworker.rs hooks, rebuilt) |

ES.1.s4 (embedding health check) was **subsumed by s3** — the per-kind
breaker covers the LM-Studio-down case, and CF-2 Option B already
removed the embed-404 crash so embed failures fail gracefully now.

## ES.1.s3 design — crash-loop breaker

The crash loop: a poison row crashes the worker, postmaster respawns it,
it claims the same class of row, crashes again (~1s cycle). The breaker:
- `kind_circuit_breaker` table tracks per-kind consecutive crashes
- the startup reaper records ONE crash per distinct kind it reaps (one
  reaper pass = +1/kind — a single bad restart with many in-flight rows
  doesn't trip it; only a genuine loop accumulates to 5)
- claim query skips kinds with `paused_until > now()`
- a successful completion resets the kind's counter + clears the pause
- threshold 5 → 10-minute pause, auto-expires

State lives in the DB (a crash kills the process — in-memory counters
would not survive). The reaper-on-every-restart is what makes the
counter accumulate across the loop.

## The smoke test — the verification

Dispatched a small research-write run (es-smoke-nomic-embed-compare,
binding question: nomic-embed v1.5 vs v2 differences). Cost cap $0.60.

**Result: clean pass. verified, $0.205, real 6377-char artifact.**

Every guardrail confirmed in production:
- No runaway — context_gather 4 rounds (soft cap 5), synthesize 2
  (soft cap 3). Clean stage progression.
- Zero bgworker crashes.
- 10 embed jobs, all routed to lm_studio, all completed — CF-7 fix +
  the LM Studio nomic reload both verified working.
- Review output started with `REVIEW: passes` — the L.1.1.14 verify
  gate satisfied honestly; maturity advanced to verified on a real
  draft, not the bacteriopolis "where's the draft" failure.
- No oversized-fetch path triggered (small sources) — the thing that
  bled never engaged.

## One finding worth keeping

Model identity is fuzzy. kimi-k2.6 was reported under three strings in
one run — `kimi-k2.6`, `accounts/fireworks/models/kimi-k2p6`,
`moonshotai/kimi-k2.6-20260420` — and gather failed over between
Fireworks and Moonshot routes mid-stage. Same logical model, so not a
bug, but cost attribution splits three ways and the L.1.1.15
substitution detector can't see gateway-route changes. A canonical
model-name mapping would fix both. Filed low-priority, ES.3-era.

## Where the substrate stands

**Stable.** The bleed class is closed and the closure is verified by a
real pipeline run, not just unit smokes. Caps hold, embeds route right,
no crashes, the verify gate is honest, artifacts are real.

**Not yet thoughtful.** That's ES.3 — the judge-compiled-brief
rearchitecture. The substrate currently handles oversized fetches by
refusing to chunk them (the 40-leaf breaker) and falling back to raw —
safe, but the old behavior with a net. ES.3 is the design work that
makes a web fetch produce a small compiled brief instead. Council
first; fresh context.

## The arc, named

Two days: Batch L → L.1.1 (council, research, ratification, infra,
closeout) → L.1.1.x post-mortem fix bundle → ES (emergency stop,
diagnosis, ES.1 stabilization, verification). ~60 commits. Two
overcorrections caught and corrected (L24/L25, CF-2 scope). The
discipline — smoke, commit, journal, surface-tensions, ratify-before-
build — is what kept ~60 commits at zero rollbacks even through a
genuine incident. ES.3 opens the next session.
