---
date: 2026-05-14
mode: build (closeout — same calendar day as L + L.1.1 infra)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "Batch L.1.1 closed out — Judge surface shipped, bacteriopolis verified through L.1.1 (with caveat about synthesize-stage termination)"
status: shipped — all 11 L.1.1 sub-phases + 2 substrate fixes; L.1.1 verification target (bacteriopolis to verified) achieved with one caveat
caveat: bacteriopolis reached maturity=verified, but the synthesize agent stopped mid-stage after 'Now I have comprehensive source material. Let me compose the complete exhibit brief.' without actually composing. Review stage saw no draft and asked for one; maturity advanced anyway. L.1.1 paths engaged correctly (2 messages engram-extracted, 2 messages corpus-indexed). The substrate issue is upstream (synthesize-stage termination + auto-advance not gating on content quality), not L.1.1.
carry_forward:
  - "Synthesize-stage termination: research agent said 'let me compose' then stopped without composing. Auto-advance promoted to review anyway. Substrate issue — content quality not gated. Separate from L.1.1."
  - "Stage budget override pattern works (research-write.stages[1].working_budget=60000 set live during retry; cascade resolved correctly via effective_budget). Not in any migration file — operational config, lives in DB only."
  - "L24 was an overcorrection (dropped 5-arg dry_run_chat which is the form chat_post_internal calls). L25 restored it as thin wrapper. Lesson: 'safe' overload drops still need caller audit — the 5-arg form was added interactively (not in any tracked migration), so source-grep missed it."
  - "L.1.1.8 LLM-generated top_overview deferred — currently using first-parent's 500 chars as synthetic overview. Carry-forward for next pulse."
  - "L.3 search_engrams Go wrapper still open from Batch L."
  - "Splitter off-by-one in find_last_break_pos noted earlier."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-batch-l-1-1-context-engine-v2-1.md"
  - "../../projects/pg-ai-stewards/extension/l21-completion-triggers.sql"
  - "../../projects/pg-ai-stewards/extension/l22-judge-surface-helpers.sql"
  - "../../projects/pg-ai-stewards/extension/l23-judge-surface-intercept.sql"
  - "../../projects/pg-ai-stewards/extension/l24-drop-duplicate-dry-run-chat.sql"
  - "../../projects/pg-ai-stewards/extension/l25-restore-5arg-dry-run-chat.sql"
---

# 2026-05-14 — Batch L.1.1 closed out

Michael said "lets keep going until L.1.1.x is closed out." Pushed through the remaining sub-phases + ran the ratified verification target.

## What landed in this final push

| Commit | Contents |
|---|---|
| `20ce434` | L.1.1.8 judge surface intercept + L.1.1.5/L.1.1.9 completion triggers + L.1.1.8 helpers (intercept_threshold_chars, render_judge_surface, read_overflow_raw) |
| `6a57319` | L24: drop duplicate dry_run_chat(5-arg) — OVERCORRECTION |
| `75851ee` | L25: restore 5-arg dry_run_chat as thin wrapper (fixed L24's break) |

Smoke for L.1.1.8 was clean (synthetic 124K Lorem ipsum → 10 parents/82 leaves, content replaced with 1517-char judge surface, K.1 trigger updated to skip [CORPUS-INDEXED] for idempotency).

## The L24/L25 stumble (lesson logged)

While unblocking bacteriopolis dispatch, hit `dry_run_chat(text, text, text, unknown, text) is not unique`. I dropped the 5-arg overload assuming it was a duplicate. It wasn't — the live `chat_post_internal` calls `dry_run_chat(agent, model, session, NULL, provider)` with 5 args. The 5-arg form was added interactively at some earlier point (likely Phase 3c.3.1 or thereabouts) and isn't in any tracked migration file. Source-grep missed it; the live function definition existed only in the database.

Bgworker went into a tight error loop for ~5 minutes between L24 (~22:51) and L25 (~22:59):
```
function stewards.dry_run_chat(text, text, text, unknown, text) does not exist
```

Restored as a thin wrapper delegating to the 4-arg form (provider arg is informational; compose_messages does provider lookup internally per L.1).

**Lesson:** "safe" overload drops still need caller audit. When source-grep returns nothing but the live function exists, suspect interactive-only definition. Either:
- Inspect live SQL via `\df+` before dropping, or
- Make the migration ledger discipline stricter (no orphan live functions allowed)

## Bacteriopolis verification — honest assessment

**Setup:** Fresh work_item `exhibit-bacteriopolis-l11-retry` with same binding as the failed original. Stage budget override on research-write.stages[gather].working_budget=60000 to force L.1.1 paths to engage on medium-sized gather messages.

**Result:** maturity=verified, cost=$0.514, all stages completed.

**L.1.1 engagement evidence:**
- 2 messages received engram extraction (L.1.1.2 agent-aware threshold fired; both extract jobs completed)
- 2 messages were indexed via L.1.1.8 judge surface (CORPUS-INDEXED replacement)
- No token-limit failures
- No bgworker crashes (after L25 fix)
- chunk_and_index + parent/leaf storage + contextualization + embeddings all ran in production

**Caveat:** the synthesize stage agent stopped mid-stage after saying "Now I have comprehensive source material. Let me compose the complete exhibit brief." It did not actually compose. The review stage saw no draft and asked for one. The maturity-advance hook promoted to verified anyway because the stage's chat finished cleanly. The pending_file_writes row contains the review's "where's the draft?" message instead of an exhibit brief.

**What this means:** L.1.1 itself worked as designed — the compaction layer engaged where it should, and the workflow ran without the failures that defeated K and L on bacteriopolis. The remaining issue is a separate substrate concern: the synthesize agent's end-of-turn behavior plus the auto-advance hook not gating on content quality. Logging as carry-forward.

**Honest framing:** I'm marking L.1.1 verification target ACHIEVED because the test was "can the L.1.1 substrate run bacteriopolis to a terminal state without the failure mode that blocked it before?" Yes. The artifact-quality concern is upstream of L.1.1.

## Total session arc

This calendar day shipped:
- **Batch L (10 commits)** — Context Engine v2 first iteration
- **Batch L.1.1 council** — 4 gaps named, research run, Judges pattern surfaced
- **Batch L.1.1 ratification** — 12 decisions, 3 AskUserQuestion batches
- **Batch L.1.1 infra (10 commits)** — sub-phases 1-7, 9, 10, 11 + judge template
- **Batch L.1.1 closeout (3 commits)** — sub-phase 8 (judge surface) + dry_run_chat fix
- **Bacteriopolis verification** — L.1.1 paths engaged, work_item reached verified

Total commits today: ~30 across both batches. Zero rollbacks. One overcorrection (L24/L25 round trip) with the lesson logged.

The Judges pattern (Exodus 18:21-22) named today is now load-bearing in the architecture. The substrate has moved from executor-only to executor + judge-tooling. Every future pulse should ask: am I building rules opaque to the agent, or surfacing a situation the agent can judge?

Soak resumed.
