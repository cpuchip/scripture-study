# `.stewards/kimi-k2.6/` — kimi-k2.6 prompt variants

Targets `model_match = 'kimi-*'` in `stewards.agents`. Use these when
the substrate dispatches to `provider=opencode_go, model=kimi-k2.6`
(or any other kimi-* variant we add later).

## Why kimi needs its own prompts

We ran the substrate's `study-write` pipeline end-to-end on
2026-05-07 against the FtC ↔ WtL binding question. Output:
`study/two-triplets-one-ascent.md`. Michael's read: "really good for
kimi" but with two specific issues — repetition between sections IV
and V, and (after Opus 4.7 review) a confabulated quote-fix in the
revision notes that introduced drift rather than removed it.

The post-mortem identified six kimi-k2.6 voice/worldview signatures
that a model-neutral prompt does not address strongly enough:

1. **Symmetric-pair compulsion.** Kimi reaches for symmetric mappings
   (interior/exterior, perceiver/perceived) and pre-decides answers
   in diagram form. Beautiful but the symmetry is the model's
   contribution, not the text's.
2. **Triadic flourishes.** "Three witnesses, one tree, one ascent."
   "The instrument and the music, the eye and the light, the
   traveler and the road." Triplets as cadence are a strong tic.
3. **Closing-refrain instinct.** Kimi compresses the thesis into a
   one-line landing at the bottom even when explicitly told not to.
   Has to be forbidden by *function* (thesis-restatement), not just
   *form* (three sentences).
4. **Pseudo-citation register for internal corpus.** Kimi treats its
   own corpus as an academic bibliography (`[study-name] reads this
   as...`). Michael integrates prior studies as natural references.
5. **Latinate over Anglo-Saxon.** Architecture, mechanism, ontological,
   geometry, perceptual. Kimi picks the abstract register when both
   are available.
6. **Confabulation under audit pressure.** When the prompt asks for
   revision notes, kimi *generates* revision notes — including ones
   describing verifications that didn't happen. The 2026-05-07 study
   claimed to have "removed 'which is' from Romans 5:5 to match the
   source"; the source has "which is." This is the diagnostic case.

## Files

| File | Status | Targets |
|------|--------|---------|
| [study.agent.md](study.agent.md) | v1 (2026-05-08) | scripture study agent, kimi-tuned |

Add new variants here as we ship them. Watchman-consolidator already
has a kimi-* row in `3a-watchman-pass.sql`; if we tune it further
we'll mirror it here for traceability.

## What v1 of study.agent.md changes from base

Compared to `.github/agents/study.agent.md`:

- **Phase 4 (Drafting) — added:**
  - "Open with a scene, not a claim" rule
  - "Section headers are claim sentences, not labels" rule
  - Anglo-Saxon-over-Latinate cut list (architecture, mechanism,
    ontological, geometry, perceptual, complementary, terminal-point)
- **Phase 5 (Review) — strengthened:**
  - Closing refrain forbidden by *function*, not just form. Triadic
    flourishes ("X is the Y, the Z, the W") explicitly named as
    closing refrains under another name.
  - Symmetry audit: name the symmetry once, briefly. Spend more time
    on what the text resists than on what completes the symmetry.
  - **Verification claims must be tool-grounded.** If revision notes
    describe a quote correction, the verification must come from a
    tool call in this session. Memory of the source is not
    verification. If gospel-engine-v2 is unavailable, do not claim
    quote corrections — flag uncertainty instead.
- **Pre-draft baseline match — strengthened:**
  - Read at least one of the three named voice-baseline studies
    BEFORE drafting (not at the review pass). Voice is set by example;
    rules alone are not enough for kimi.

## What stays the same

- All seven phases, all skills referenced, all rubric content
- Tool list (it's the same agent in different voice)
- Handoffs to journal/lesson agents
- The "warmth over clinical distance" framing
- The deep-reading / wide-search / quote-log / source-verification
  skill loadout

## How to test a variant after editing

1. Edit the file in this folder
2. Apply to substrate (manual `psql` paste until importer learns
   `model_match` — see `.stewards/README.md`)
3. Run a study via the substrate:
   ```
   stewards-cli work-item create study-write \
     --binding-question "..." \
     --provider opencode_go --model kimi-k2.6 \
     --token-budget 2000000
   ```
4. Compare output to Michael's voice baselines
5. If the change targeted a specific kimi-ism (e.g., closing refrain),
   confirm that specific behavior changed in the output
6. Commit with a journal entry describing what was tested

## Iteration log

- **2026-05-08 — v1 authored.** Targets the six kimi signatures
  identified in the two-triplets review. Untested against a fresh
  pipeline run; expect at least one revision after first test.
