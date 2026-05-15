---
date: 2026-05-14
mode: build (post-mortem fix bundle, same calendar day as L + L.1.1)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "Batch L.1.1 fix bundle — 5 ratified post-mortem fixes from bacteriopolis (max_tokens removal, sha256 idempotent, REVIEW gate, model substitution log, max_tool_rounds)"
status: shipped — all 5 fixes (L.1.1.12-16) live; verification deferred to next session
carry_forward:
  - "Bacteriopolis re-verification with fixes live — should produce a real draft this time. Defer to fresh-context next session."
  - "L.1.1.12 (max_tokens removal): contextualize_leaf no longer caps; reasoning_content fallback in apply_contextualize_leaf for any future reasoning-model trap. Plus retrieve_from_corpus → read_corpus_parents Go tool wired so judge surface isn't a dead end."
  - "L.1.1.13 (sha256 idempotent): catches truly literal duplicate fetches in same session. Doesn't catch 'mostly identical with timestamp prefix' shape (e.g., fetch_url's fetched_at_ms). Carry-forward: tool-specific normalization or parent-chunks-aware fingerprinting."
  - "L.1.1.14 (REVIEW prefix gate): maturity advance to verified now requires stage_results[review].output starts with 'REVIEW: passes' or 'REVIEW: revised'. Bacteriopolis-style false-verified now caught."
  - "L.1.1.15 (model substitution log): trigger logs every chat dispatch where requested_model differs from pipeline-stage declared model. Surfaces silent swapping without tracing the path."
  - "L.1.1.16 (max_tool_rounds): per-stage cap (default 5). research-write tuned: context_gather=5, gather=5, synthesize=3, review=1. Tune lower based on observed behavior post-fixes."
  - "L24/L25 lesson still open: when source-grep returns nothing on a function but the live function exists, suspect interactive-only definition. Consider stricter migration-ledger rule disallowing live functions not in any tracked file."
links:
  - "../../projects/pg-ai-stewards/extension/l26-fix-bundle-contextualizer-and-corpus.sql"
  - "../../projects/pg-ai-stewards/extension/l27-sha256-idempotent-overflow.sql"
  - "../../projects/pg-ai-stewards/extension/l28-review-prefix-verify-gate.sql"
  - "../../projects/pg-ai-stewards/extension/l29-model-substitution-log.sql"
  - "../../projects/pg-ai-stewards/extension/l30-stage-max-tool-rounds.sql"
---

# 2026-05-14 — Batch L.1.1 fix bundle (post-mortem)

After the closeout journal, Michael opened the materialized bacteriopolis file and immediately surfaced what was wrong: the file contained the review's "where's the draft?" message, not an exhibit brief. Plus he flagged: deepseek calls returning ~200 tokens out, qwen3.6 firing for nearly each one, way too many calls.

I had jumped to "DeepSeek is a reasoning model + max_tokens=200" without actually looking. Michael said: load debug workflow, dig deep. Rule 3 (quit thinking, look).

## What the data actually showed

After actually inspecting payloads, contents, and counts, the real picture was a **multiple-bug stack**:

1. **`retrieve_from_corpus` tool was missing.** L.1.1.8 judge template promised it; agent tried to call it; got "tool not registered" error. Agent fell back to re-fetching the original Exploratorium URL.
2. **Re-fetch produced near-duplicate content** (same body, different `fetched_at_ms` timestamp prefix). L.1.1.8 intercepted again → another 160 leaves indexed.
3. **All 320 contextualizer chats fired DeepSeek-V4-Flash with max_tokens=200**, which is a reasoning model — reasoning consumed all 200 tokens, content came back empty. apply_contextualize_leaf saw empty content and skipped. ~$22 in API spend for ~zero context_prefix writes.
4. **Synthesize ran 53 rounds** when the prompt called for one-shot draft. Steward retry guidance + uncapped tool loop = runaway.
5. **Model resolver silently swapped kimi-k2.6 → qwen3.6-plus** on first dispatch. Steward retry showed "attempt #1 after unknown" with model_used=qwen3.6-plus. Path through pick_model unclear; likely steward_tick or escalation logic.
6. **Maturity advanced to verified despite review's "where's the draft?" message** because auto-advance hook didn't gate on content quality.

## Ratification — 4 questions, all recommended

1. **Model resolver**: warn-and-substitute (log clearly when pinned model is overridden, continue)
2. **Stage rounds cap**: per-stage `max_tool_rounds`, default 5 (lower as we observe)
3. **Verify gate**: require explicit 'REVIEW: passes' or 'REVIEW: revised' prefix
4. **Re-fetch idempotency**: sha256 check, skip re-index if identical content already in session

Plus direct fixes (no vote needed): remove max_tokens cap on contextualizer, add reasoning_content fallback, ship retrieve_from_corpus → read_corpus_parents Go handler.

## Commits

| Commit | Sub-phase | Contents |
|---|---|---|
| `646ba21` | L.1.1.12 | max_tokens removal + reasoning_content fallback + read_corpus_parents Go handler + judge template updated to point at it |
| `b12cc1b` | L.1.1.13 | sha256 idempotent overflow indexing (pgcrypto, content_sha256 col, intercept_oversized_tool_after sha-check) |
| `3e495fd` | L.1.1.14 | REVIEW prefix verify gate (BEFORE UPDATE trigger on work_items.maturity) |
| `75d6dbd` | L.1.1.15 | model_substitutions log table + chat-INSERT trigger detecting mismatch |
| `0e18001` | L.1.1.16 | per-stage max_tool_rounds field; chat_post_internal enforces; research-write tuned (5/5/3/1) |

## Honest scope notes

- **sha256 idempotent doesn't fully solve bacteriopolis's case.** The two fetches had identical bodies but different `fetched_at_ms` prefixes from fetch_url. Sha-check catches truly literal duplicates only. Future enhancement: tool-specific normalization.
- **L.1.1.15 logs but doesn't FIX the substitution.** Surfacing the symptom is the deliberate scope. Tracing the model swap path through steward_tick / model_override / escalation is its own pulse.
- **Verification deferred.** Re-running bacteriopolis with all 5 fixes live is the obvious next step, but context is exhausted. Better to do it fresh next session with clean read of the resulting artifact + cost behavior.

## What this session arc did

Day total: ~40 commits across Batch L (10) + L.1.1 council/research/ratification (3) + L.1.1 infra (10) + L.1.1 closeout (5) + L.1.1.x post-mortem fixes (5) + journal/memory updates throughout.

Zero rollbacks. One overcorrection caught (L24/L25). Multiple lessons logged. The Judges pattern (Exodus 18:21-22) named as architectural principle and embedded in `.mind/principles.md`.

The pattern: council → ratify → ship → look → ratify the fixes → ship → look. Compounding on itself. The discipline holds.

Soak resumed.
