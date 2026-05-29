---
date: 2026-05-29
title: J.11 — enforced Gemini prepaid spend cap + streamed cost-capture fix
batch: J.11
status: shipped — committed c3cc2cc, soak resumed
---

# J.11 — Gemini spend enforcement

## Ask + framing

Michael: "I want the gemini models opt in — they are not on a subscription basis and thus cost real money against a bucket I must refill, which I only have $20 in presently."

Two layers: the literal ask (opt-in) and the intent (don't let Gemini burn the $20). Verified Gemini was **already opt-in** five ways (no catalog default, no pipeline stage, no pipeline metadata default, no scheduled pipeline, no bucket). So the real work was the intent: spend enforcement. Michael ratified building real enforcement now — prepaid-balance model, $18 cap, gemini-only, refuse-at-dispatch.

## What I found before building

- **Bucket caps are NOT enforced** anywhere — `bucket_limit_micro` is read only to carry the limit into new periods. Even opencode's $12/$60 are informational. So a new bucket wouldn't have enforced anything.
- **Per-work-item `cost_cap_exceeded` IS enforced** (steward, 4a/4c/4d/6b/6c) but reads `work_items.cost_micro_dollars`, which stayed $0 for Gemini → never trips.
- **Gemini cost wasn't tracked** (the J.10 carry-forward): the bgworker records cost only when usage tokens > 0 (`bgworker.rs:845`); Gemini omits the usage block on streamed responses unless `stream_options.include_usage=true` is sent. `parse_chat_sse` already captures a usage-only tail chunk — so the fix was purely the request side.

## Three pieces shipped

1. **bgworker.rs** — `stream_options:{include_usage:true}` in the chat body builder. No-op for opencode (gateway already streams usage), the fix for Gemini. Needed a pg rebuild. **Closes the J.10 gemini cost-tracking gap.**

2. **j11-provider-spend-caps.sql** — new `provider_spend_caps` table (provider, cap_micro, since, enforced), `provider_spend_since()`, `provider_cap_exceeded()`, `provider_cap_refill()`, and a GATE in `work_item_dispatch_stage` (carries the J.8.a 4-layer fallback forward; RAISEs before enqueue when an enforced provider is over cap). Opt-in per provider via `enforced`; only google_gemini enforced. Seeded $18.

3. **Dockerfile** — added `am1-pending-file-writes-notify.sql` to the COPY list. **Pre-existing build break** surfaced on rebuild: lib.rs referenced am1 (commit 767386a) but it was never in the build context, so `cargo pgrx package` failed. The running container predated that lib.rs entry. am1 was the ONLY lib.rs ref missing from the COPY (all 55 others present); the j-series ledger files apply via `stewards-cli migrate`, not lib.rs, so they don't block the build. Resolves the build-blocking slice of the J.8/J.9 foldback debt.

## Smoke (live, post-rebuild)

- Gemini dispatch now records cost_events (gemini-2.5-flash-lite, $0.00299 total); `provider_spend_since('google_gemini')` = 2990 micro. **Tracking fix confirmed.**
- Cap gate: cap below spend → `provider_cap_exceeded('google_gemini')`=true; Gemini dispatch RAISEd a clean refusal message; opencode NOT gated; cap restored to $18.
- Fixed a RAISE format bug (plpgsql RAISE has no printf `%.2f`/`%L` — pre-round + literal-quote).

## Refill UX

After topping up the real Google balance:
```sql
SELECT stewards.provider_cap_refill('google_gemini');            -- reset clock, same $18 cap
SELECT stewards.provider_cap_refill('google_gemini', 20000000);  -- reset + new $20 cap
```

## Carry-forward

- **Brainstorm surfacing of the gate**: a Gemini lens hitting the cap RAISEs inside `spawn_children`, caught by `on_maturity_verified`'s EXCEPTION handler (logged; 0 children spawned). Direct/CLI/MCP dispatch surfaces the message directly. A pre-flight cap check in `start_brainstorm` would surface it cleanly for the brainstorm path — small future polish.
- **Bucket caps still unenforced** for all providers (informational only) — separate concern; not changed here.
- **Unpriced opencode preview models** (deepseek-v4-pro, mimo-*, hy3-preview) still unpriced (J.10).
- **J-series foldback debt**: the build-blocking part (am1) is fixed; the j1–j11 ledger files remain live/migrate-applied (not in lib.rs), which is the substrate's normal pattern — not debt.
