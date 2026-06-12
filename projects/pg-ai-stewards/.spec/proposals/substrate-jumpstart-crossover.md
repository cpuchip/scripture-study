# ai-jumpstart × pg-ai-stewards — the crossover reflection

**Status:** PLANNED — Michael, 2026-06-12 (later same day): "this is
good. lets plan to do it, but I dont want to do it yet." Crossover 1
(the test lab) is ratified-to-plan as **task #30**; build waits until
OSS-P1 is underway (Michael's sequencing). Includes the book-lane
handoff: the kimi/qwen jumpstart rounds route through this machinery.
Grounded in `projects/ai-jumpstart/` (v0.2.1) and
`experiments/ai-jumpstart/findings.md` (runs 1–3, ablation, night laps —
now FIVE passing models).

## The relationship: same DNA, two altitudes

ai-jumpstart installs the practices into a **conversational harness** via
files; the substrate bakes them into a **runtime** via rows. The mapping
is exact:

| jumpstart (files) | substrate (rows) |
|---|---|
| intent.md | stewards.intents |
| covenant.md | stewards.covenants (+ extensions/presiding) |
| journal/ + active.md | messages, work_items, journals — the flight recorder |
| ask-before-build setup counsel | dominion_in_council / gates |
| "close the loop" memory discipline | update_memory covenant + Stop hooks |
| DISCIPLINES.md heavier pack | instructions table (per-family, per-model rows) |

The kit is the paper form; the substrate is the database form. One
lineage of practice text should feed both (see Crossover 2).

## Crossover 1 — the substrate as the kit's TEST LAB

The findings' open queue is exactly what the substrate already does well:
multi-model dispatch (opencode + LM Studio locals — kimi-k2.6,
qwen3.6-27b, gemma, nemotron are all configured providers), cost
accounting, fully reviewable sessions. A `jumpstart-eval` pipeline could:

- run the scripted turn-1/turn-2 prompts against any provider/model the
  bridge can reach (the queue's kimi + local-model arms, unblocked);
- score with a critic stage against the findings' counted markers
  (premature build, constraint upgrade, name assumption, pre-design,
  memory-loop completion) — the council pattern, applied as a grader;
- file each arm as a work item: cost, transcript, verdict, replayable.

And the findings' own "substrate path" bullet (hand the kit to a coder
sandbox and watch it flush out a repo end-to-end) is the existing
coder-mcp `code-pr` pipeline pointed at a repo seeded with AGENTS.md —
near-zero new machinery, one session of wiring WHEN ratified.

## Crossover 2 — the kit as the OSS seed pack's source text

The OSS extraction ships generic covenant/intent templates. ai-jumpstart
already HAS field-tested generic templates (and the ablation proved the
seed raises the floor across Sonnet/Gemini/Haiku). Two "generic covenant"
texts WILL drift apart; there should be one lineage — plausibly: the kit
is the canonical practice text, and the OSS seed pack derives its
templates from it (kit is MIT; attribution clean; the book link rides
along). Decision belongs to the extraction council at P1.

## Crossover 3 — the ablation finding maps onto agent variants

The ablation's core result — *the pack moves smaller models toward
larger-model behavior; the seed alone is enough for the big ones* — is
exactly what the substrate's per-model agent variants exist to express
(`family + model_match`). The measured way in: smaller-model variants
(`qwen-*`, `gemma-*`, flash-class) carry the DISCIPLINES distillation in
their variant prompts/instructions rows; frontier variants stay lean.
This would be the first time variant tuning is driven by counted
experiment data instead of felt drift. (Also note the floor/ceiling
language IS preside §V's walls-vs-judgment: the seed's hard gates are
walls; forethought is judgment. Same taxonomy, third appearance.)

## Crossover 4 — onboarding cold models INTO the substrate

When the substrate spawns a child (subagent, persona, coder task), the
child is a cold model entering a covenant it has never read. The kit is
literally an onboarding document for that moment — a jumpstart-shaped
preamble for spawn targets would carry the presiding chain downward in
the form already proven to steer cold models. (The DISCIPLINES pack
already contains keep-the-watch-whole and when-you-delegate-you-preside,
distilled from the same study that produced PR.1.)

## If one thing gets ratified first

The test-lab pipeline (Crossover 1) — it serves the kit's open
experiment queue immediately, exercises the substrate's multi-provider
muscle, produces counted data that feeds Crossover 3, and builds nothing
speculative: every piece (pipelines, providers, critic stages, cost caps)
already exists.
