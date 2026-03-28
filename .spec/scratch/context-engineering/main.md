# Context Engineering for Conference Reindex — Scratch File

*Research provenance for .spec/proposals/context-engineering.md*

---

## Binding Problem

Local LLMs (nemotron-3-nano at 30B) can only evaluate what's in their context window. Conference talks are usually explicit about Christ-connections and score well at surface level. But scriptures encode Christ typologically — Alma 32 scores 1 on `teach_about_christ` without cross-references but 7-8 when you see the tree-of-life architecture. The model needs context it can't generate itself.

**Specific problem:** A prompt-only approach (TITSW v2) closes the nemotron-GLM quality gap on *talks* but cannot close the *depth* gap on scriptures or help the model understand the TITSW framework from first principles. The model needs:
1. Knowledge of what the TITSW principles actually mean (not just the scoring rubric)
2. Core gospel vocabulary — the theological patterns that connect seemingly unrelated passages
3. Cross-reference context for the specific content being evaluated

---

## Inventory — What Already Exists

### TITSW Manual (local copies)
- `gospel-library/eng/manual/teaching-in-the-saviors-way-2022/04-part-1/05-teach-about-jesus-christ.md`
- `gospel-library/eng/manual/teaching-in-the-saviors-way-2022/04-part-1/06-help-learners-come-unto-christ.md`
- `gospel-library/eng/manual/teaching-in-the-saviors-way-2022/07-part-2/08-love-those-you-teach.md`
- `gospel-library/eng/manual/teaching-in-the-saviors-way-2022/07-part-2/09-teach-by-the-spirit.md`
- `gospel-library/eng/manual/teaching-in-the-saviors-way-2022/07-part-2/10-teach-the-doctrine.md`
- `gospel-library/eng/manual/teaching-in-the-saviors-way-2022/07-part-2/11-invite-diligent-learning.md`

### Studies — Principle Exemplars

**Elder Brown "Eternal Gift of Testimony" (Oct 2025)** — `study/talks/202510-24brown.md`
- The gold standard analysis. Scores 5/5 stars on all TITSW dimensions.
- Contains detailed principle-by-principle breakdown.
- Exemplifies: specificity invites Spirit, vulnerability creates safety, multi-layered invitations.

**Elder Bednar "In the Space of Not Many Years" (Oct 2024)** — `study/talks/202410-35bednar.md`
- Strong on doctrine (Helaman as mirror for our day), invite (specific calls to action).
- Not explicitly analyzed with full TITSW framework but shows the pattern.

**President Oaks "Coming Closer to Jesus Christ" (Feb 2026)** — `study/talks/Coming-Closer-to-Jesus-Christ.md`
- Explicitly Christ-centered. Four practical points all connecting to doctrine of Christ.
- Strong on teach_about_christ and doctrine.

**Overview study** — `study/teaching-in-the-saviors-way/00_overview.md`
- Full TITSW framework with scriptural examples for each principle.
- Examples from Christ's teaching for each of the 4 principles.
- This IS the master reference for what the principles mean.

### Studies — Meta-Principle Exemplars

**Doctrines, Principles, Programs** — `study/doctrines-principles-programs.md`
- The doctrine-principle-program hierarchy. 3 Nephi 11:32-35 as the doctrine of Christ.
- Shows how ALL teaching connects back to Christ.

**Testimony Meetings YT eval** — `study/yt/Zq1IEXTXmsw-testimony-meetings.md`
- Contains the TITSW connection: "all things are branches of the same tree" ↔ Moses 6:63.
- The "Teach About Christ No Matter What" principle with prophetic sourcing.

**Charity study** — `study/charity.md`
- Webster 1828 analysis: charity = love, not almsgiving.
- Moroni 7:47-48 (pure love of Christ, bestowed as gift).
- Connects to the "sheddeth itself abroad" thread.

**Faith Part 1** — `study/faith-01.md`
- Lectures on Faith: faith as principle of action and power.
- Foundation for TITSW "Invite Diligent Learning" — faith requires action.

### Core Scripture Sources (all verified to exist locally)

| Source | File | TITSW Relevance |
|--------|------|-----------------|
| 3 Nephi 11 | `gospel-library/eng/scriptures/bofm/3-ne/11.md` | The doctrine of Christ defined by Christ himself |
| 3 Nephi 27 | `gospel-library/eng/scriptures/bofm/3-ne/27.md` | "This is my gospel" — Christ defines His gospel |
| Moroni 7 | `gospel-library/eng/scriptures/bofm/moro/7.md` | Faith, Hope, Charity — the character of a teacher |
| 1 Nephi 11 | `gospel-library/eng/scriptures/bofm/1-ne/11.md` | Tree of life = love of God = Christ (interpretive key) |
| Alma 32 | `gospel-library/eng/scriptures/bofm/alma/32.md` | The seed = the word = Christ (typological depth test) |
| John 1 | `gospel-library/eng/scriptures/nt/john/1.md` | "The Word was God" — grounds the word=Christ connection |
| Ether 12 | `gospel-library/eng/scriptures/bofm/ether/12.md` | Faith, hope, charity as prerequisite for Christ encounter |
| 2 Nephi 31 | `gospel-library/eng/scriptures/bofm/2-ne/31.md` | The doctrine of Christ — baptism, Holy Ghost, endure |
| 1 Corinthians 13 | `gospel-library/eng/scriptures/nt/1-cor/13.md` | The charity chapter — love defined |
| Romans 5 | `gospel-library/eng/scriptures/nt/rom/5.md` | "Love of God shed abroad" — the universal thread |
| Articles of Faith | `gospel-library/eng/scriptures/pgp/a-of-f/1.md` | First principles and ordinances |
| Moses 6:63 | `gospel-library/eng/scriptures/pgp/moses/6.md` | "All things testify of Christ" |

### Ground Truth Reference
- `experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md` — created this session

### Existing Pipeline
- Harness: `run-test.ps1` loads `context.md` → system message, `prompts/*.md` → user template, `content/*.md` → inserted at `{{CONTENT}}`
- Current context.md is ~250 tokens (system values only)
- Content files include footnote markers from gospel-library source
- Model: nemotron-3-nano at 131k context (expandable to 1M)

---

## Architecture Design — Context Package Layers

### Layer 1: System Context (~250 tokens) — EXISTS
`context.md` — core values and constraints. No change needed.

### Layer 2: TITSW Framework (~2,000-3,000 tokens) — NEW
A curated summary of the Teaching in the Savior's Way framework. NOT the raw manual chapters (too verbose). A dense reference that gives the model vocabulary for:
- What each principle means concretely
- What high vs low scores look like
- Examples from Christ's teaching as the gold standard
- The two meta-principles as framing

Source: Synthesize from `00_overview.md` + the 6 manual chapters.

### Layer 3: Gospel Vocabulary (~3,000-5,000 tokens) — NEW
Core theological patterns the model needs to recognize:
1. **The Doctrine of Christ** (3 Ne 11, 3 Ne 27) — faith, repentance, baptism, Holy Ghost, endure
2. **The Tree of Life interpretive key** (1 Ne 11:21-25) — tree = love of God = Christ
3. **The Word = Christ** (John 1:1-14) — the seed/word dual meaning
4. **"All things testify of Christ"** (Moses 6:63, 2 Ne 11:4) — the typological principle
5. **Faith, Hope, Charity** (Moroni 7, Ether 12, 1 Cor 13) — the character trajectory
6. **"Sheddeth itself abroad"** thread (1 Ne 11:22, Rom 5:5, Moro 8:26)
7. **Types and shadows** — brass serpent, paschal lamb, tree of life, living water, bread of life
8. **First principles and ordinances** (AofF 1:4, 2 Ne 31, 3 Ne 27)

This is the "gospel vocabulary" Michael described — curated once, delivered every time.

### Layer 4: Content-Specific References (~variable) — NEW
For each piece of content being evaluated, resolve its footnotes/cross-references:
- For conference talks: extract scripture references → include relevant verse context
- For scriptures: resolve footnote markers → include cross-referenced passages

This is the `<references>` section Michael described.

### How It Fits in the Pipeline

Current: `system: context.md | user: prompt + content`

Proposed: `system: context.md + titsw-framework.md + gospel-vocab.md | user: prompt + <references>resolved refs</references> + <content>text</content>`

### Token Budget

At nemotron's 131k context:
- System (context + framework + vocab): ~6,000-8,000 tokens
- References per content: ~2,000-5,000 tokens
- Content itself: ~2,000-10,000 tokens
- Total input: ~10,000-23,000 tokens
- Max output: 4,096 tokens
- **Headroom at 131k: massive.** Even at 1% utilization we're fine.

At 1M context: absurdly generous. Could include entire manual chapters if needed.

### Token Budget for Batch (5,500 talks)
Per talk: ~15,000 tokens in + 4,096 tokens out ≈ ~19,000 tokens
Total: ~105M tokens input (shared context) + ~22M tokens output
This is dominated by the repeating context (~8,000 tokens × 5,500 = ~44M tokens just for framework+vocab).

**Optimization:** System prompt prefix caching eliminates ~44M tokens of redundant prefill. See research below.

---

## API Caching Research (Mar 28, 2026)

*Sources: LM Studio docs (REST API v1, stateful chats, load model, parallel requests, OpenAI-compat), llama.cpp discussions #8860 and #13606, llama.cpp PR #16391, KV-cache-aware prompt engineering benchmarks.*

### The Mechanism: llama.cpp KV Cache Prefix Reuse

LM Studio uses llama.cpp as its backend. llama.cpp's server has built-in **prompt prefix caching** at the slot level:

1. When `"cache_prompt": true` is included in the request body, the server stores the KV cache for the prompt in the assigned slot.
2. On the next request to the same slot, llama.cpp compares the new prompt's token sequence against the cached sequence.
3. **Any matching prefix is skipped entirely** — the KV values are already computed and stored. Only the new/different tokens require prefill computation.
4. This works at the token level. If the first 8,000 tokens are identical, those 8,000 tokens have zero prefill cost on subsequent requests.

Confirmed by ggerganov (llama.cpp maintainer, Aug 2024): *"make sure to use set `cache_prompt = true` in the requests to enable this feature."*

### Practical Impact for Our Batch

- System prompt (context.md + titsw-framework.md + gospel-vocab.md) = ~8,000 tokens, **identical across all 5,500 requests**
- Request 1: full prefill of ~15,000 tokens (system + prompt + content)
- Requests 2–5,500: only ~7,000-15,000 NEW tokens need prefill (prompt template + talk content). The ~8,000 system tokens are cached.
- **Savings: ~44M tokens of prefill computation eliminated from batch**
- At nemotron's prefill speed: eliminates ~2-4 seconds per request of redundant computation

### Three Approaches (simplest → most complex)

**Option 1: `cache_prompt: true` on OpenAI-compat endpoint (RECOMMENDED)**
- Works with our existing harness (`/v1/chat/completions`)
- Just add `"cache_prompt": true` to the request body
- llama.cpp automatically assigns slots by prefix similarity (`-sps 0.5` is the default — a slot is reused if 50%+ of the prompt prefix matches)
- With a single slot and sequential batch processing, every request after the first reuses the system prefix cache
- Zero code changes beyond adding one JSON field

**Option 2: LM Studio v1 Stateful Chats**
- `/api/v1/chat` with `previous_response_id` continues a conversation without resending history
- Designed for multi-turn conversation, not batch-of-independent-evaluations
- Could "prime" the system context once and branch from it, but awkward for our use case
- Would require rewriting the harness to use the v1 API

**Option 3: Explicit slot management**
- Use `id_slot` parameter to pin requests to specific slots
- Use `--slot-save-path` + `/slots/{id}?action=save` and `?action=restore` to persist KV cache to disk
- Survives server restarts
- Maximum control but requires client-side slot state management
- Useful if we add parallel processing later (different slots for different prompt types)

### Design Constraints Created by Caching

**The system prompt must be byte-identical across all requests.** Any character change breaks the entire prefix cache. This means:
- `context.md + titsw-framework.md + gospel-vocab.md` must be concatenated in the same order every time
- No per-request customization in the system message
- All variable content (references, talk text) goes in the user message
- JSON serialization of any structured data in the system prompt must use deterministic key ordering

**This validates the layered architecture:** Layers 1-3 (stable, in system message, cached once) → Layer 4 + content (variable, in user message, processed each time).

### Parallel Batch Optimization (future)

LM Studio supports `Max Concurrent Predictions` via continuous batching. For parallel batch processing:
- Load nemotron with `-np 2` or higher (2+ parallel slots)
- Each slot maintains its own KV cache
- Pin the "talk evaluation" prompt to slot 0 and (if we add scripture evaluation later) pin that to slot 1
- Both slots cache their respective system prefixes independently
- From the tutorial: *"Use cases: Applications with long, repeated system prompts. Batch processing with similar prompt templates."*

### Source Links
- llama.cpp discussion #8860: "Does the KV cache persist across multiple requests sharing a prefix?" — ggerganov confirms `cache_prompt: true`
- llama.cpp discussion #13606: "Tutorial: KV cache reuse with llama-server" — full walkthrough with slot persistence
- llama.cpp PR #16391: "server: host-memory prompt caching" — merged, extends cache to host memory
- Ankit Sinha benchmark: stable prefixes → 65% median TTFT improvement, 85% cache hit rate, 71% cost reduction
- Sankalp blog: deep technical walkthrough of paged attention and automatic prefix caching

---

## Implementation Plan

### Phase 1: Curate the Context Package (this session)
1. Create `experiments/lm-studio/scripts/context/titsw-framework.md` — synthesized TITSW reference
2. Create `experiments/lm-studio/scripts/context/gospel-vocab.md` — curated theological patterns
3. Both files hand-crafted from deep reading, not LLM-generated

### Phase 2: Update TITSW Prompt to v3
1. Move from 0-3 to 0-9 scale
2. Add `<references>` and `<context>` sections to prompt template
3. Update JSON schema for the richer scoring

### Phase 3: Build Reference Resolver (future)
1. Script that reads a content file's footnote markers
2. Looks up cross-referenced verses through gospel-mcp or direct file reads
3. Packages them as `<references>` block
4. For batch: automate this for all 5,500 talks

### Phase 4: Test and Validate
1. Run v3 prompt with context package on Alma 32 (hard test: typological depth)
2. Run on Kearon (easy test: should match ground truth)
3. Run on Brown (full star exemplar — should score high across all dimensions)
4. Compare scores to ground truth

### Phase 5: Scale
1. Build batch pipeline for 5,500 conference talks
2. Consider KV cache optimization for shared context
3. Separate pipelines for talks (lighter context) vs scriptures (heavier context + types)

---

## Critical Analysis

### Is this the RIGHT thing?
YES. The binding problem is real — models can't score typological depth from surface text. The ground truth study proved it. Context engineering is the standard approach in production LLM pipelines.

### Simplest version?
Phase 1 + a manual v3 prompt test. Two curated documents + one prompt update + one test. Can be done in one session.

### What gets WORSE?
- Token usage per request increases ~5x from current. At nemotron speeds, this is seconds, not minutes.
- Context maintenance — if the curated documents drift from the manual, they become stale. Mitigation: they're synthesized from local copies we control.
- Risk of over-fitting: model may learn to "see Christ" everywhere because we tell it to. The scoring rubric's anti-inflation language mitigates this.

### Does this duplicate something we have?
No. gospel-mcp has cross-references but doesn't package them as context for LLM evaluation. The TITSW overview study exists but isn't in a format optimized for an LLM context window.

### Mosiah 4:27 check
Michael has 7 priorities. This is an extension of priority #3 (model experiments), not a new priority. The incremental work is: 2 curated markdown files + 1 prompt update. Reasonable scope.
