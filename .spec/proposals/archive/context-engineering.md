# Context Engineering for Conference Reindex

*Proposal — Mar 28, 2026*
*Scratch file: [.spec/scratch/context-engineering/main.md](../scratch/context-engineering/main.md)*
*Ground truth: [experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md](../../experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md)*

---

## Problem Statement

A local 30B-parameter model (nemotron-3-nano) can only evaluate what's in its context window. The TITSW v2 prompt closed the scoring quality gap between nemotron and GLM for conference talks, but it cannot close the **depth gap**:

- **Alma 32** scores `teach_about_christ: 1` at surface level — correct for what's visible. But with the tree-of-life typological architecture (Alma 32:42 ↔ 1 Ne 8:10-11 ↔ 1 Ne 11:22-25), the informed score is 7-8. The model can't see this because 1 Nephi 11 isn't in its context window.
- **Kearon's talk** scores similarly surface and deep because talks are explicit. The gap shows up on scriptures.
- The model doesn't know what Love, Spirit, Doctrine, and Invite *mean* beyond the brief rubric. It has no reference for what these principles look like when done well.

Prompt engineering hits a ceiling. Context engineering is the next lever.

---

## Success Criteria

1. Nemotron + context package scores Alma 32 `teach_about_christ` ≥ 5 (recognizes Christ typology)
2. Kearon scores hold steady (no regression from inflated context)
3. The Brown talk exemplar scores ≥ 7 on all dimensions (it genuinely excels)
4. The context package fits comfortably within 131k context with room for content + output
5. The curated files are hand-verified against local source material (no LLM-generated theology)

---

## Constraints

- **Model:** nemotron-3-nano at 131k context (expandable to 1M). Batch target: 5,500 conference talks.
- **Speed:** Cannot add more than ~5s overhead per evaluation (currently 18.5s avg)
- **Sources:** All context curated from local gospel-library files. Every claim traceable to a verse.
- **Scale:** 0-9 (replacing 0-3). Rubric must be anchored to prevent inflation.
- **NOT in scope:** Building the batch pipeline. Reference resolver automation. Scripture-specific pipeline. Those are future phases.

---

## Proposed Approach

### Architecture: Layered Context

```
┌─────────────────────────────────────────────────┐
│ SYSTEM MESSAGE                                   │
│  ┌─────────────────────────────────────────────┐ │
│  │ Layer 1: System Context (context.md)        │ │
│  │ ~250 tokens — core values, constraints      │ │
│  │ EXISTS — no change                          │ │
│  ├─────────────────────────────────────────────┤ │
│  │ Layer 2: TITSW Framework (titsw-framework)  │ │
│  │ ~2,500 tokens — principles defined,         │ │
│  │ exemplars, what high/low scores look like   │ │
│  │ NEW — synthesized from manual + studies     │ │
│  ├─────────────────────────────────────────────┤ │
│  │ Layer 3: Gospel Vocabulary (gospel-vocab)    │ │
│  │ ~3,500 tokens — theological patterns:       │ │
│  │ doctrine of Christ, tree of life key,       │ │
│  │ word=Christ, types & shadows, faith/hope/   │ │
│  │ charity, "all things testify"               │ │
│  │ NEW — curated from scripture, verified      │ │
│  └─────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────┤
│ USER MESSAGE                                     │
│  ┌─────────────────────────────────────────────┐ │
│  │ Prompt: TITSW v3 (0-9 scale, rubric)       │ │
│  │ ~800 tokens                                 │ │
│  ├─────────────────────────────────────────────┤ │
│  │ <references> (footnote cross-refs)          │ │
│  │ ~2,000-5,000 tokens — variable per content  │ │
│  │ FUTURE — manual for now, automated later    │ │
│  ├─────────────────────────────────────────────┤ │
│  │ <content> (the actual text to evaluate)     │ │
│  │ ~2,000-10,000 tokens                        │ │
│  └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘

Total input: ~11,000-22,000 tokens (~8-17% of 131k)
Output: 4,096 tokens max
```

### Layer 2: TITSW Framework

Synthesize from the manual chapters and Michael's overview study into a dense reference document. Structure:

1. **Meta-Principle A: Teach About Christ** — definition, what it looks like at different score levels (surface vs typological), examples from Christ's teaching, the "all things testify" principle
2. **Meta-Principle B: Help Come Unto Christ** — definition, the distinction between *knowing about* and *experiencing*, the receiving/transformation pattern
3. **Principle 1: Love** — what Christlike love looks like in teaching (with specific examples from Christ)
4. **Principle 2: Spirit** — what teaching *by* the Spirit means vs teaching *about* the Spirit
5. **Principle 3: Doctrine** — scripture-grounded teaching, making truth personally relevant
6. **Principle 4: Invite** — specific, escalating invitations to act

For each principle: 2-3 lines of definition, 1-2 exemplar quotes from scripture or talks, what differentiates a 3 from a 7 on the 0-9 scale.

### Layer 3: Gospel Vocabulary

Eight theological patterns curated from scripture, each with key verses:

| Pattern | Key Verses | Why It Matters |
|---------|-----------|----------------|
| Doctrine of Christ | 3 Ne 11:32-35, 3 Ne 27:13-21, 2 Ne 31:2-21 | The foundational gospel — faith, repentance, baptism, Holy Ghost, endure |
| Tree of Life = Love of God = Christ | 1 Ne 11:21-25 | The interpretive key for all tree/seed/fruit imagery in the Book of Mormon |
| The Word = Christ | John 1:1-14, Alma 33:22-23 | The dual meaning of "the word" in scripture — message AND Person |
| "All things testify of Christ" | Moses 6:63, 2 Ne 11:4 | The typological principle — even passages that don't name Christ may encode Him |
| Faith, Hope, Charity | Moroni 7:40-48, Ether 12:4-28, 1 Cor 13:1-13 | The character trajectory of a disciple — and a teacher |
| "Love shed abroad" | 1 Ne 11:22, Rom 5:5, Moro 8:26 | The theological thread connecting tree of life → Holy Ghost → diligent discipleship |
| Types and Shadows | brass serpent (Alma 33:19), paschal lamb, living water (John 4:14), bread of life (John 6:35) | Patterns that encode Christ across all dispensations |
| First Principles & Ordinances | AofF 1:4, 2 Ne 31, 3 Ne 27 | The unchanging core around which all teaching orbits |

Each pattern: 3-5 lines of explanation + key verses quoted (from local copies, verified). Total ~3,500 tokens.

### TITSW v3 Prompt Changes

1. **Scale: 0-9** with anchored rubric:
   - 0: Not present
   - 1-2: Incidental/minor
   - 3-4: Present but not a focus
   - 5-6: Intentional and significant
   - 7-8: Central to the teaching approach
   - 9: Defining — this content would be the textbook example

2. **New fields in JSON:**
   - `typological_depth`: 0-9 (how much hidden Christ-typology exists beyond surface)
   - `cross_reference_density`: count of explicit scripture/prophetic citations
   - `surface_vs_deep_delta`: for each dimension, note if informed reading would change the score

3. **Anti-inflation strengthened:** "A score of 7+ means this content could be used as a teaching example for this principle. Most conference talks score 4-6 on most dimensions. Reserve 8-9 for content that is genuinely exceptional."

4. **Reference-aware instruction:** "If `<references>` are provided, use them to inform your scoring. Cross-references that reveal deeper Christ connections should increase the `teach_about_christ` and `help_come_unto_christ` scores. Score based on the full available context, not just surface text."

---

## Phased Delivery

### Phase 1: Curate Context Package (1 session)
**Deliverables:**
- `experiments/lm-studio/scripts/context/titsw-framework.md`
- `experiments/lm-studio/scripts/context/gospel-vocab.md`
- Both hand-verified against local scripture files

**Verification:** Every quoted verse read from `gospel-library/` before inclusion. Source-verification skill applies.

### Phase 2: TITSW v3 Prompt (same session)
**Deliverables:**
- `experiments/lm-studio/scripts/prompts/titsw-v3.md` — 0-9 scale, reference-aware
- Update `run-test.ps1` to support `-Context` parameter (loads additional context files into system message)

**Verification:** Prompt parses correctly. Harness accepts new parameter.

### Phase 3: Validate (same session if nemotron is loaded)
**Tests:**
1. Alma 32 + context package → `teach_about_christ` ≥ 5 (was 1 without context)
2. Kearon + context package → scores stable, no inflation
3. Brown talk + context package → ≥ 7 across dimensions

**Verification:** Compare against ground truth in [references/ground-truth-alma32-kearon.md](../../experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md).

### Phase 4: Manual Reference Resolution (future session)
For each test content file, manually include `<references>` with resolved cross-references. Test whether explicit cross-reference inclusion further improves scoring.

### Phase 5: Automated Reference Resolver (future)
Build a script that:
1. Reads a content file for footnote markers
2. Queries gospel-mcp or reads files directly for cross-referenced verses
3. Packages them as `<references>` block
4. Outputs enriched content file

### Phase 6: Batch Pipeline (future)
Apply to all 5,500 conference talks with KV cache optimization (see below).

---

## Batch Optimization: KV Cache Prefix Reuse

*Research completed Mar 28, 2026. Full provenance in [scratch file](../scratch/context-engineering/main.md).*

llama.cpp (LM Studio's backend) has built-in **prompt prefix caching**. When `"cache_prompt": true` is included in the request body, the server stores the KV cache for the prompt in the assigned slot. On subsequent requests to the same slot, **any matching token prefix is skipped entirely** — zero recomputation.

**Impact for batch:** The system prompt (~8,000 tokens of context.md + titsw-framework.md + gospel-vocab.md) is identical across all 5,500 talk evaluations. With `cache_prompt: true`:
- Request 1: full prefill (~15,000 tokens)
- Requests 2–5,500: only ~7,000-15,000 NEW tokens need prefill. The 8,000 system tokens are cached.
- **~44M tokens of redundant prefill eliminated from the batch**
- Estimated savings: ~2-4 seconds per request

**Implementation:** Add `"cache_prompt": true` to the existing OpenAI-compat request body in `run-test.ps1`. No API migration needed. llama.cpp auto-assigns slots by prefix similarity (default `-sps 0.5`).

**Design constraint:** The system prompt must be **byte-identical** across all requests. Any character change breaks the prefix cache. This validates the layered architecture: stable system layers 1-3 (cached once) → variable user message layer 4 (processed each time).

**Future parallel option:** Load nemotron with `-np 2+` for multiple concurrent slots. Each slot caches its own prefix independently. Pin talk evaluations to slot 0, scripture evaluations (heavier context) to slot 1.

---

## Two Streams: Study + Dev

This work splits into two independent streams that can execute in parallel:

### Study Stream (deep reading, curation)
**Where:** This session, study agent
**What:** Curate the context documents through deep reading of scripture sources and manual chapters.

| Deliverable | Description | Dependencies |
|-------------|-------------|---------------|
| `gospel-vocab.md` | 8 theological patterns with key verses, ~3,500 tokens | Read ~12 scripture sources, verify all quotes |
| `titsw-framework.md` | TITSW principles defined with score anchors, ~2,500 tokens | Read 6 manual chapters + overview study |

Both documents are hand-crafted from deep reading. Every quoted verse must be read from `gospel-library/` before inclusion. Source-verification skill applies.

### Dev Stream (harness, prompt, pipeline)
**Where:** Parallel VS Code session, dev agent
**What:** Update the harness and prompt to use the context package.

See [dev handoff spec](.spec/proposals/context-engineering-dev.md) for the self-contained spec.

| Deliverable | Description | Dependencies |
|-------------|-------------|---------------|
| `run-test.ps1` update | `-Context` parameter, loads additional files into system message | None — can scaffold with placeholder context files |
| `cache_prompt: true` | Add to request body for KV cache prefix reuse | None |
| `titsw-v3.md` prompt | 0-9 scale, reference-aware, new JSON schema | Prompt structure is independent of context content |
| Validation runs | Test Alma 32, Kearon, Brown against ground truth | Requires study stream context files |

**Sequencing:** Dev can start immediately on harness changes and prompt design. Validation runs happen after study stream delivers the context files.

---

## Costs & Risks

| Cost | Impact | Mitigation |
|------|--------|------------|
| Token usage per request: ~6,000 → ~15,000 | First request adds ~2-5s prefill | KV cache prefix reuse eliminates this for requests 2+ |
| Two curated documents to maintain | Maintenance burden if manual changes | Manual updates rarely; documents are synthesis, not copies |
| Risk of context-induced inflation | Model sees "Christ is everywhere" and over-scores | Anti-inflation rubric language; ground truth comparison |
| Phase 1 requires deep reading/verification | Agent time for source verification | This is the right kind of work — quality context is the product |
| Byte-identical system prompt constraint | Any system prompt change invalidates KV cache | System prompt files are stable by design; changes are intentional |

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | The model needs gospel vocabulary to score depth. Without it, surface scoring is the ceiling. |
| Covenant | Rules? | Source verification applies to curated context. Every verse must be read before quoting. |
| Stewardship | Who owns what? | Agent curates initial context; Michael reviews and approves. Context files live in experiments/. |
| Spiritual Creation | Spec precise enough? | Yes — two documents with clear structure, one prompt update, three validation tests. |
| Line upon Line | Phasing? | Phase 1 (curate) stands alone and is immediately testable. |
| Physical Creation | Who executes? | Agent (plan mode → dev mode handoff for harness changes; study mode for curation). |
| Review | How to verify? | Ground truth comparison. Three specific test cases with expected score ranges. |
| Atonement | If it goes wrong? | Context files are additive — can be removed without breaking existing pipeline. v2 prompt remains. |
| Sabbath | When to pause? | After Phase 3 validation. Review results before scaling to batch. |
| Consecration | Who benefits? | Michael directly. Eventually anyone using gospel-vec with TITSW scores. |
| Zion | How does it serve the whole? | Better TITSW scores → better conference talk recommendations → better teaching preparation. |

---

## Recommendation

**Build.** Phase 1-3 in one session. The work is well-scoped, the binding problem is real and proven, and the context package is an investment that improves every subsequent evaluation. The curated documents are also intrinsically valuable as study artifacts — they encode deep reading about what the TITSW principles actually mean and what the gospel's theological patterns are.

**Phase 1 first deliverable:** Start with `gospel-vocab.md` (the harder, more valuable piece). It requires reading and synthesizing from ~12 scripture sources. The TITSW framework is a faster synthesis from existing overview study + manual chapters.

**Executing agents:** Two parallel streams:
- **Study agent** (this session): Curate gospel-vocab.md and titsw-framework.md through deep reading
- **Dev agent** (parallel session): Update harness, write TITSW v3 prompt, add `cache_prompt: true`, run validation

Dev handoff spec: [.spec/proposals/context-engineering-dev.md](context-engineering-dev.md)
