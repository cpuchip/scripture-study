# Two-Triplets Multi-Model Comparison — 2026-05-08

This folder holds the artifacts of pg-ai-stewards Phase 3c.3.4 — the
multi-model voice experiment running the same FtC/WtL binding question
through three substrate configurations.

## The runs

| File | Run | Model | Prompt | Pipeline | Created | Notes |
|------|-----|-------|--------|----------|---------|-------|
| `run1-original-kimi-base-prompt.md` | #1 | kimi-k2.6 | base study agent (`*`) | `study-write` | 2026-05-07 21:20Z | The 3c.3.2 artifact. Source of `study/two-triplets-one-ascent.md` after Michael read it and Opus 4.7 revised it. |
| `run2-kimi-kimi-tuned.md` | #2 | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | 2026-05-08 04:55Z | Same model + same pipeline as run #1; only the prompt changed. Tests whether the kimi-tuned prompt addresses the six signatures. |
| `run3-qwen-base-prompt.md` | #3 | qwen3.6-27b (lm_studio) | base study agent (`*`) | `study-write-qwen` | 2026-05-08 04:55Z | Different model + same prompt as run #1. Tests qwen's natural voice for future tuning. |

## Comparison

See `comparison.md` (written after both runs complete).

## What changed about the agent variants

**Base prompt** (`.github/agents/study.agent.md`, `(study, *)`): the
canonical voice-tuned-for-Michael study agent. 12835 chars. This was
the prompt run #1 used (giving us the baseline kimi tics) and the
prompt run #3 uses (giving us a fresh look at qwen's defaults).

**Kimi-tuned prompt** (`.stewards/kimi-k2.6/study.agent.md`,
`(study, kimi-*)`): the same agent's body but with six amendments
targeting kimi-specific signatures identified in the 2026-05-07
Opus-4.7 review of run #1:

1. Symmetric-pair compulsion → Phase 3a symmetry check + Phase 5 audit
2. Triadic flourishes → cut list + Phase 5 audit
3. Closing refrain by *function* not form → Phase 5 audit
4. Pseudo-citation register for internal corpus → naturalize via Phase 4
5. Latinate over Anglo-Saxon → cut list in Phase 4
6. Confabulation under audit pressure → "verification claims must be tool-grounded" rule

14325 chars. Verified import with `(family='study', model_match='kimi-*')`.

## What didn't change

The pipeline structure (`outline → draft → review`) is identical for
runs #2 and #3. Token budget is 2M for both. Binding question is
verbatim identical to run #1. The only intentional differences are
the prompt (run #2) and the model (run #3).
