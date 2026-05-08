# `.stewards/qwen-3.6/` — qwen-3.6 prompt variants

Targets `model_match = 'qwen*'` in `stewards.agents`. Use these when
the substrate dispatches to `provider=lm_studio, model=qwen/qwen3.6-27b`
(or any other qwen-* variant we add later).

## Why qwen needs its own prompts

The 2026-05-08 multi-model voice experiment ran the FtC/WtL binding
question on three configurations: kimi+base, kimi+kimi-tuned, and
qwen+base. The qwen+base run (run #3 in the comparison memo) shared
some kimi voice signatures and exhibited several distinct ones. Six
qwen-specific signatures worth addressing:

1. **Tool-name confusion.** qwen called `study_get('bofm/ether/12')`
   as if substrate study slugs were scripture references (they
   aren't — they're kebab-case names like `way-truth-life`). The
   tool failed silently and qwen kept retrying. The base prompt
   doesn't explicitly forbid this.
2. **Broken internal-link convention.** qwen rendered substrate
   references as `[study-name](#)` instead of `[study-name](study-name.md)`.
   The convention isn't enforced anywhere in the prompt; for kimi the
   examples carry through. qwen needs the rule stated.
3. **Heavy table use.** qwen produced a full comparison table in
   §VI of run #3 that no kimi run did. Reads as substituting structure
   for argument.
4. **Bold-emphasis density.** Bold-emphasized phrases throughout
   ("**They are not the same concepts.**", "**They ARE the same
   architecture.**"). Reads as preacher-cadence. Kimi rarely bolds.
5. **Triadic body emphasis.** Three parallel constructions back-to-back
   in the middle of paragraphs — even when not closing a section.
6. **Verbosity.** 239 lines for run #3 vs 118 for run #4 (kimi-tuned).
   Says things twice — once briefly, once in a parallel construction.

In addition, qwen exhibits some shared kimi signatures (label-style
section headers, closing refrains in body sections). The qwen-tuned
prompt addresses those too, plus the six qwen-specific tendencies.

## Files

| File | Status | Targets |
|------|--------|---------|
| [study.agent.md](study.agent.md) | v1 (2026-05-08) | scripture study agent, qwen-tuned |

## What v1 of study.agent.md changes from base

Compared to `.github/agents/study.agent.md`:

- **All six kimi-tuned amendments** — closing refrain by function,
  symmetry audit, scene-opener, claim-headers, Anglo-Saxon register,
  verification grounding (matches `.stewards/kimi-k2.6/study.agent.md`)
- **Plus six qwen-specific amendments:**
  - "study_get takes a kebab-case slug, NEVER a scripture reference"
  - "Internal links use `[slug](slug.md)`. Never use `(#)` placeholders."
  - "Tables are NOT a substitute for argument. Use prose; tables only when comparing 4+ items along 4+ dimensions where the parallel structure is the whole point."
  - "Bold-emphasis is for **single load-bearing terms**, not for **entire sentence clauses**. Maximum 2 bolds per paragraph."
  - "Triadic constructions ('X is A; Y is B; Z is C') are reserved for the close. Do not deploy them mid-paragraph."
  - "Brevity over completeness. If you can say it once, do not say it twice."

## How to test

1. Edit the file in this folder
2. Apply to substrate via re-import (importer now respects `model_match`):
   ```
   stewards-cli import --source agent:.stewards/qwen-3.6 -v
   ```
3. Run a study via the substrate against qwen:
   ```
   stewards-cli work-item create study-write-qwen --slug <slug> \
     --input '{"binding_question":"..."}' --budget 2000000
   stewards-cli work-item dispatch <id>
   ```
4. Compare output to qwen+base baseline (run #3 in
   `study/.scratch/two-triplets-comparison-2026-05-08/run3-qwen-base.md`)
5. Specifically verify the six qwen-tics are absent

## Iteration log

- **2026-05-08 — v1 authored.** Targets the six qwen-specific
  signatures from run #3 plus all six kimi-shared signatures.
- **2026-05-08 — v1 validated.** Run #5 of the FtC/WtL pipeline
  cleared all 12 targeted signatures: 0 bad `study_get` paths,
  0 `(#)` broken links, 0 mid-argument tables, no whole-clause
  bolds, triadic constructions deployed once at a structural
  moment, 110 lines total (vs run #3's 239), proper claim-style
  section headers with `Therefore`/`But` causation chains,
  active verification correction in revision notes. Token/time
  efficiency vs run #3 baseline: 16% fewer tokens, 61% faster.
  **Status: stable v1.** See
  `study/.scratch/two-triplets-comparison-2026-05-08/comparison.md`
  for the five-way analysis.
