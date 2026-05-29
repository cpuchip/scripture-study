---
date: 2026-05-29
title: J.12 — budget/cap error surfacing (classify + pre-flight + UI)
batch: J.12
status: shipped — committed 0f535a4, no pg rebuild
---

# J.12 — make budget errors easy to tell

## Ask

Michael: "surface that budget error for gemini models from google so it's easy to tell what happened."

## Key realization

The raw error is **already captured** — the advance trigger's error path (`3c2:72-85`) calls `work_item_fail`, which sets `work_items.error`. A Google 429 lands as `work_items.error = "chat dispatch failed at stage lens: chat HTTP 429: {... RESOURCE_EXHAUSTED ...}"`. So this is **read-time classification + clearer surfacing, no bgworker/pg rebuild**.

## Shipped

- `classify_error(text)` → category (spend_cap_reached | provider_budget | rate_limited | auth | timeout | other | none). Gemini returns 429 for both rate limits and quota exhaustion, so quota/RESOURCE_EXHAUSTED wording is matched **before** generic 429 → real budget exhaustion classifies as `provider_budget`.
- `work_item_failures` view for triage.
- `start_brainstorm` **pre-flight**: resolves each lens's provider, RAISEs a clear refusal before spawning if any routes to an enforced+over-cap provider. Closes the swallowed-refusal gap (the J.11 dispatch gate's RAISE inside spawn_children is caught by the trigger handler → 0 children, logged only; the pre-flight surfaces it to the MCP tool/caller). Smoke: clean RAISE, 0 parents created.
- stewards-ui: API returns `error_category`; WorkItemDetail.vue shows an amber budget banner + refill hint; WorkItems.vue shows a "💸 budget" list badge.

## Smoke

classify_error all categories correct; pre-flight refuses cleanly; API detail returns `error_category: provider_budget` for a synthetic Gemini 429. UI rebuilt + serving (data path proven).

## Closes

- Brainstorm cap-refusal surfacing (the J.11 carry-forward "a gemini lens hitting the cap RAISEs inside spawn_children, swallowed"). Now pre-flighted.

## Remaining (minor)

- The dispatch-gate RAISE inside spawn_children is still swallowed for the *non-brainstorm* fan-out path (only start_brainstorm has the pre-flight). Other fan-outs don't route to enforced providers today, so low risk.
- Could add `error_category` to the MCP `work_item_show` output for parity with the UI — small, deferred.
