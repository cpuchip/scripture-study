---
date: 2026-05-29
title: Opus 4.8 harness review — what (little) needed adjusting
workstream: WS5-adjacent (harness/meta)
session_type: harness-review
status: shipped — 3 files edited, committed
---

# Opus 4.8 harness review

## Context

Michael updated both harnesses (Claude Code + GitHub Copilot) to **Claude Opus 4.8** (`claude-opus-4-8`) and set Claude Code's effort to `xhigh`. He asked me to read the migration guide and look over the workspace for anything that needed adjusting "for me specifically."

## The finding (honest headline)

**Almost nothing needed to change, and saying so was the right answer.** 4.8 builds on 4.7 — the migration guide lists no behavioral reversals. Every piece of 4.7-tuning in the harness is still correct for 4.8:

- "Curiosity over inference / compensate because this model uses tools less" — still true
- "Foresight & Adjacent Surfaces" + "Honor intent, not just literal request" (literalism compensation) — still true
- "Positive examples shape voice better than negative rules" — still the guidance

Resisting a migration rewrite was the move — "Stability After Improvement" + "Reduce Before Adding" both apply. A big overhaul would have been make-work.

## What actually changed for 4.8 (the small delta)

1. **Effort is the main dial, recalibrated.** Guide: "Effort is more important for this model than for any prior Opus." Default `high`; `xhigh` for coding/agentic; the levels were re-baselined (xhigh = substantially more thinking than 4.7's xhigh). Michael set `xhigh` — correct for dev/substrate weeks. `high` is better for pure prose/study (less overthinking).
2. **API-mechanical items** (not relevant to our markdown harness): 1M context default, mid-conversation system messages, lower prompt-caching minimum, refusal stop_details now documented.
3. **Substrate untouched** — pg-ai-stewards calls opencode_go models, not Claude; the effort param never reaches it.

## Edits made (3 files, surgical)

- `.github/copilot-instructions.md` — "Model context" note updated to 4.8 (both harnesses), with explicit "the 4.7 tuning applies to 4.8 unchanged." Two stale "4.7" citation refs (Curiosity, Writing Voice) → "Opus 4.7/4.8."
- `.mind/active.md` — Model banner updated to 4.8 + effort note; cost-table line relabeled `Opus 4.8=7.5x` with a flag that the multiplier is the last-known 4.7 rate and needs confirming.
- `CLAUDE.md` addendum — new "Model & effort (Claude Code on Opus 4.8)" section: effort default `high`, Michael's standing default `xhigh`, drop to `high` for pure study, per-session dial Michael owns.

## One alignment worth naming

4.8 is "more direct and opinionated, with less validation-forward phrasing." That pulls the SAME direction as the workspace's deepest values — the Ben Test ("your AI is too complimentary"), "trust the discernment / don't over-hedge," "warmth over flattery." For the first time the model's native disposition and the harness's anti-sycophancy are rowing together rather than the harness fighting the model.

## Carry-forward

- **Confirm the Opus 4.8 Copilot premium-request multiplier.** active.md cost table currently shows `7.5x` flagged as the last-known 4.7 rate. If 4.8 changed it, update active.md:72.
- Watch over the next few sessions whether `xhigh` over-deliberates on pure prose/study; if so, suggest `high` for those sessions (now documented in the addendum).
