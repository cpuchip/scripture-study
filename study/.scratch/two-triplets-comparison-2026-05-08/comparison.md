# Three-Way Comparison: FtC/WtL Studies — 2026-05-08

> **Status:** in progress (runs #2 and #3 still flying as of 05:02Z).
> This memo will be filled in once both terminate. Sections below are
> the rubric I'll apply.

## What we're comparing

| Run | Model | Prompt | Pipeline | Corpus access? |
|-----|-------|--------|----------|----------------|
| #1 (original) | kimi-k2.6 | base | study-write | ✅ yes |
| #2 (this session) | kimi-k2.6 | kimi-tuned | study-write | ❌ no (bug — see journal) |
| #3 (this session) | qwen3.6-27b (lm_studio) | base | study-write-qwen | ❌ no (same bug) |

**Important caveat upfront.** Runs #2 and #3 dispatched during a
window when my 3c.3.3 reimport had wiped the substrate's
`study_*: allow` broadcast grant. Both runs ran without access to
`study_search_text`, `study_get`, `study_similar`, `study_citations`,
or `study_context_for`. Run #1 had full access.

So this is **not** a clean apples-to-apples corpus-grounded
comparison. It's a voice comparison with two models (kimi vs qwen)
and one prompt-tuning swap (base vs kimi-tuned). The corpus
grounding is part of run #1's strength; runs #2 and #3 are
working from training memory + skills only.

A proper apples-to-apples follow-up (run #4: kimi-tuned + corpus
access) is on the roadmap below.

## Rubric — six kimi signatures from the 2026-05-07 review

For each run, did the signature appear?

| Signature | Run #1 (k+base) | Run #2 (k+tuned) | Run #3 (qwen+base) |
|-----------|-----------------|------------------|---------------------|
| 1. Symmetric-pair compulsion | ✅ present | TBD | TBD |
| 2. Triadic flourishes | ✅ present | TBD | TBD |
| 3. Closing refrain (by function) | ✅ present | TBD | TBD |
| 4. Pseudo-citation register | ✅ present (`[study-name] reads...`) | TBD (no corpus → may not exhibit) | TBD (no corpus → may not exhibit) |
| 5. Latinate over Anglo-Saxon | ✅ present | TBD | TBD |
| 6. Confabulation in revision notes | ✅ present (Romans 5:5 reverse-fix) | TBD | TBD |

## Voice metrics (objective)

For each run, count:

| Metric | Run #1 | Run #2 | Run #3 |
|--------|--------|--------|--------|
| Word count | TBD | TBD | TBD |
| Em-dashes per paragraph (mean) | TBD | TBD | TBD |
| Section count | 6 | TBD | TBD |
| Section headers as theses (vs labels) | 0/6 | TBD | TBD |
| Direct quote count | TBD | TBD | TBD |
| Paraphrase count | TBD | TBD | TBD |
| `Therefore`/`But` transitions | TBD | TBD | TBD |
| Latinate hits (architecture/mechanism/etc.) | TBD | TBD | TBD |
| Closing-refrain detection | yes | TBD | TBD |

## Findings

### Did the kimi-tuned prompt fix the six signatures?

*(filled in after runs complete)*

### What does qwen do differently from kimi out of the box?

*(filled in after runs complete)*

### Was the no-corpus regression visible in the output?

*(filled in after runs complete; expected: yes — both runs will
produce studies with no `[study-name]` cross-references and either
verbatim quotes from training memory or pure paraphrase.)*

## Roadmap — what tonight surfaces for daytime

1. **Run #4** (if the soak window allows tonight, otherwise daytime):
   kimi-tuned prompt + corpus access. The actual apples-to-apples
   experiment we tried to run.
2. **Importer architecture fix:** `agent_tool_perms` needs source
   provenance so substrate-internal broadcasts survive frontmatter
   reimports. Or: move broadcasts into frontmatter (already done
   for `study_*` on both study agent files).
3. **qwen-tuned variant:** if run #3 surfaces qwen-specific signatures
   distinct from kimi's, author `.stewards/qwen-3.6/study.agent.md`.
4. **gospel-engine HTTP tools (3c.4):** unlocks real quote
   verification at agent runtime. Deferred to daytime per overnight
   journal — needs Dockerfile changes.
