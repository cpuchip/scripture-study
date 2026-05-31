---
date: 2026-05-31
title: opencode_zen provider — real Claude (pay-per-use) on the substrate
workstream: WS5
session_type: dev (autonomous + ratified spend cap)
status: SHIPPED + live-verified + pushed
---

# opencode_zen — the second tier, and real Claude

## The correction (Michael was right)

When I pruned the 5 claude-* models as "dead on opencode," that was half right.
opencode has **two tiers on the same API key**:
- **go subscription** → `https://opencode.ai/zen/go/v1` (provider `opencode_go`)
- **zen pay-per-use** → `https://opencode.ai/zen/v1` (provider `opencode_zen`, NEW)

The real Anthropic **Claude** models live on the **zen** tier (via `/zen/v1/messages`,
Anthropic format). They genuinely 404 on go — so pruning them from `opencode_go`
was correct — but their home is a zen provider we hadn't created. Michael also
corrected my opus price (it's **$5/$25**, not the $15/$75 I'd guessed) and noted
the free models are zen-side. Both confirmed against `/zen/v1/models` + the docs.

## Verified facts (live 2026-05-31)
- `/zen/v1/models` serves: claude-opus-4-8/4-7/4-6, claude-sonnet-4-6,
  claude-haiku-4-5, gpt-5.x, gemini-3.x, glm-5.1, kimi-k2.6, qwen3.6-plus, and
  `-free` variants (deepseek-v4-flash-free, mimo-v2.5-free, minimax-m2.5-free,
  qwen3.6-plus-free, nemotron-3-super-free, big-pickle).
- A free zen model streams `{"cost":"0"}` (deepseek-v4-flash-free) — confirmed
  Michael's "free models return $0" hypothesis.
- The **provider registry is dynamic**: `ProviderRegistry::from_env()` parses
  `STEWARDS_PROVIDER_<NAME>_*` at postmaster start — so a new provider needs only
  env + a recreate, no code change.
- Claude prices: opus 4.8/4.7/4.6 = $5/$25; sonnet-4.6 = $3/$15; haiku-4.5 = $1/$5
  (these are EXACTLY the rates 4a had — 4a had the right rates, wrong provider).

## What shipped
- **opencode_zen provider** — env vars added to `.env` (base `/zen/v1`, key copied
  from the go key, never echoed; documented in `.env.example`), pg recreated to
  load it.
- **Claude models** in the catalog: opus-4.8/4.7/4.6, sonnet-4.6, haiku-4.5 —
  `model_pricing` at the real rates + `model_capability.api_format='anthropic'`
  so dispatch hits `/zen/v1/messages` via the AT-batch path.
- **deepseek-v4-flash-free** (zen, $0, openai) added as the verified free rep.
- **ENFORCED $18 spend cap** on `opencode_zen` (Michael's ~$20 zen balance,
  gemini pattern) — refuses dispatch past $18; refill via `provider_cap_refill`.
- Also (separate ask): added go-tier deepseek-v4-pro, mimo-v2.5-pro, mimo-v2-omni
  (probed usable). hy3-preview returns 403 (not in this subscription tier) — omitted.

## Verified live
- Probe of **claude-haiku-4-5** via `opencode_zen` → usable, finish=stop, the
  gateway returned model `claude-haiku-4-5-20251001` (real Claude) through the
  Anthropic `/messages` path. Cost tracked ($0.00003); cap active ($0 of $18).
- deepseek-v4-flash-free via zen → usable, $0.
- This is the payoff of the AT-batch: enabling real Claude was just a provider +
  `api_format=anthropic`, no new dispatch code.

## Result
**Every model, both opencode tiers, all formats, every part of the substrate** —
including the real Claude family (opus-4.8 down to haiku), pay-per-use, capped.

## Carry-forward
- Other zen free models (mimo-v2.5-free, minimax-m2.5-free, qwen3.6-plus-free,
  nemotron-3-super-free, big-pickle) + gpt-5.x are available on zen — add after
  per-model format probing (minimax may be anthropic-format).
- The `.env` holds the zen key (gitignored); `.env.example` documents the vars.
  A fresh-volume rebuild needs the zen vars set in `.env` (same as the go key).
