# Three-Way (now Four-Way) Comparison: FtC/WtL Studies — 2026-05-08

> **Status:** runs #1 and #2 analyzed. #3 (qwen) and #4 (kimi+corpus) still in flight as of 05:05Z. This memo will be appended once they complete.

## What we're comparing

| Run | Model | Prompt | Pipeline | Corpus access? | Tokens | Elapsed |
|-----|-------|--------|----------|----------------|--------|---------|
| **#1** original | kimi-k2.6 | base (`*`) | `study-write` | ✅ yes | 626K | 17m14s |
| **#2** kimi-tuned | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | ❌ no (perm bug) | 122K | 8m11s |
| **#3** qwen | qwen3.6-27b (lm_studio) | base (`*`) | `study-write-qwen` | ⚠ partial (perms restored mid-flight) | TBD | TBD |
| **#4** kimi-tuned-with-corpus | kimi-k2.6 | kimi-tuned (`kimi-*`) | `study-write` | ✅ yes | TBD | TBD |

**Important context.** During runs #2 and #3 dispatch, my 3c.3.3 reimport had wiped the substrate's `study_*: allow` broadcast. Run #2 ran without corpus tools; run #3's perms were restored partway through (qwen began calling tools after the fix landed). Run #4 was dispatched after the fix as the proper apples-to-apples experiment vs run #1.

## Rubric — six kimi signatures from the 2026-05-07 review

| Signature | Run #1 | Run #2 | Run #3 | Run #4 |
|-----------|--------|--------|--------|--------|
| 1. Symmetric-pair compulsion | ✅ present (perceiver/perceived, interior/exterior, instrument/music) | ❌ resists ("any such mapping crumbles under pressure") | TBD | TBD |
| 2. Triadic flourishes | ✅ present ("three witnesses, one tree, one ascent") | ❌ none | TBD | TBD |
| 3. Closing refrain by function | ✅ present ("The ascent is one, the descriptions are two, and the Person at the threshold is Christ.") | ❌ none — closes on practical action | TBD | TBD |
| 4. Pseudo-citation register | ✅ present (`[study-name] anchors...`) | N/A (no corpus this run) | TBD | TBD |
| 5. Latinate over Anglo-Saxon | ✅ present (architecture, mechanism, ontological, geometry, perceptual organ, terminal point) | ❌ none | TBD | TBD |
| 6. Confabulation in revision notes | ✅ present (claimed Romans 5:5 fix that *introduced* drift) | ❌ explicit honest disclosure: *"This revision contains zero direct quotations because no corpus tools were available"* | TBD | TBD |

**Run #2 cleared 5 of 5 measurable signatures.** Pseudo-citation can't be scored without corpus; run #4 will close that.

## Voice metrics — runs #1 and #2

| Metric | Run #1 | Run #2 | Δ |
|--------|--------|--------|---|
| Total lines | 105 | 43 | -59% |
| Section headers | 6 (labels: "Ordered Progressions", "Shared Hinge", "Terminal Point", etc.) | 0 + "Becoming" | structural simplification |
| Em-dashes (body, citation excluded) | ~12 | 0 | full compliance |
| Direct quotes | ~24 | 0 (forced by no-corpus) | n/a (run #4 will measure) |
| Reference-only citations (e.g. *(see John 14:6)*) | 0 | 7 | new pattern, prompt-required |
| Triadic flourishes | 2 | 0 | full compliance |
| Closing refrain | yes | no | full compliance |
| Latinate hits (architecture/mechanism/ontological/geometry/etc.) | 6+ | 0 | full compliance |
| Total elapsed | 17m14s | 8m11s | 53% faster |
| Total tokens | 626K | 122K | 80% fewer |

## What run #2 demonstrated

The run #2 study landed at 43 lines including a Becoming section and self-aware notes. Despite the perm regression that left it without substrate study tools, the kimi-tuned prompt produced a markedly Michael-voiced study from training memory + the prompt's own discipline rules.

Three excerpts — selected as the strongest signal:

**Opening (Michael-style scene drop):**
> "Thomas asked Jesus how to get where he was going. He wanted a path, a plan, directions. Jesus answered by naming himself three times over."

**Therefore-chain instead of "and then":**
> "But charity breaks the divide immediately. Mormon does not say charity is human love that happens to be directed at Christ. He says it is the love that comes from Christ... If charity is Christ's own love planted in us, then the human triplet already contains the person of Christ inside it.
>
> Faith and hope do not stay on the human side either... [examples] ...
>
> Therefore the human triplet is not us-centric. It is Christ-facing."

**Symmetry resistance (rather than completion):**
> "These are not one-to-one equations. Faith is not exactly the way, hope is not exactly the truth, and charity is not exactly the life. Any such mapping crumbles under pressure. They are better understood as two linked descriptions of a single relationship... They are not one reality viewed from two angles. They are two realities that only exist in relation to each other."

**Becoming with concrete practice (not abstract):**
> "Before making a difficult decision, ask whether it places you on the way, in the truth, and toward the life. When you feel charity toward someone difficult, recognize that it may not be your own emotion but his love reaching through you."

The honest-disclosure footnote at the bottom is the Phase-5 verification rule working as designed:
> "**Source verification:** The original draft contained no scriptural engagement, only a refusal. This revision contains zero direct quotations because no corpus tools were available in this session. Every scriptural allusion is rendered as paraphrase or reference-only citation, per the source-verification skill."

## What ran #2 did NOT demonstrate

Several things weren't testable because corpus access was missing:
- Whether the kimi-tuned prompt suppresses the pseudo-citation register when *real* citations are available
- Whether the agent genuinely reads the Phase-4 voice-baseline studies before drafting (the precondition was unsatisfiable)
- Whether the "verify-before-fix" rule prevents Romans-5:5-style reverse-fixes when gospel-engine-v2 is available (3c.4 still deferred)

Run #4 (in queue) addresses items 1 and 2. Item 3 awaits 3c.4.

## Findings — runs #3 and #4

*(filled in after runs complete)*

## Roadmap — what tonight surfaces for daytime

1. **The kimi-tuned prompt works.** Five out of five measurable signatures cleared in run #2. Worth promoting from experimental to "stable v1" once run #4 confirms the same discipline holds with corpus access.
2. **Importer architecture fix:** `agent_tool_perms` needs source provenance so substrate-internal broadcasts survive frontmatter reimports. Workaround applied tonight (added `study_*` to study agent frontmatter); proper fix is daytime work.
3. **qwen-tuned variant:** if run #3 surfaces qwen-specific signatures distinct from kimi's, author `.stewards/qwen-3.6/study.agent.md`. (Initial observation from run #3's tool-calling pattern: qwen burns more tokens per outline iteration than kimi — may need a "be more decisive" tuning.)
4. **gospel-engine HTTP tools (3c.4):** still daytime work — no SQL HTTP extension in pgvector base. Build pg_net or pgsql-http into a new Dockerfile stage, or extend the Rust bgworker with `tool_http` work_kind.
5. **Promote `study/two-triplets-one-ascent.md`?** The run #2 (or #4) output is *better-voiced* than the current published study. Might be worth replacing the Opus-4.7-revised file with a kimi-tuned-substrate-produced one — Michael's call.
