# Enriched Indexer: TITSW-Aware Summaries for gospel-vec

*Proposal — 2026-03-29*
*Scratch: [.spec/scratch/enriched-indexer/main.md](../scratch/enriched-indexer/main.md)*
*Parent: [.spec/scratch/debug-titsw-optimization/main.md](../scratch/debug-titsw-optimization/main.md) (audit section)*
*Related: [.spec/proposals/context-engineering.md](context-engineering.md) (prior context engineering work)*

---

## Binding Problem

The gospel-vec indexer produces generic summaries that surface nothing useful for teaching-oriented search and study. A talk about Christ's Atonement and a talk about family home evening both produce keywords like "faith, repentance, Jesus Christ." The model CAN extract rich teaching profiles (dominant dimensions, teaching modes, rhetorical patterns) — we proved this during TITSW prompt optimization. But the existing prompts contain zero TITSW vocabulary, so the summaries don't capture what the model already sees.

**Who is affected:** Michael, doing scripture study and lesson prep. The brain app (future), using semantic search to surface relevant talks. Any downstream consumer that queries gospel-vec for teaching-relevant content.

**How would you know it's fixed:** A search for "talks that enact love" returns talks where the speaker models love rather than just preaching about it. A search for "talks with strong doctrine scores" returns doctrinally rich talks rather than just talks that mention the word "doctrine."

---

## Success Criteria

1. Talk summaries include TEACHING_PROFILE fields (dominant dimensions, teaching mode, rhetorical pattern, dimension scores)
2. Talk pipeline uses vocabulary approach (TITSW terms in system prompt, no context documents)
3. Scripture pipeline uses lens approach (context documents injected, deeper keyword extraction)
4. Scores for 5 ground-truth pieces land within ±2 of established targets (sanity check, not precision)
5. New metadata fields are filterable in chromem-go (flat `map[string]string` fields)
6. Backward compatibility: existing KEYWORDS/SUMMARY/KEY_QUOTE format preserved in all summaries
7. Full conference reindex completes without errors on 5,500+ talks

---

## Constraints

- **Model:** nemotron-3-nano via LM Studio at localhost:1234. Temperature 0.2, 131k context.
- **Storage:** chromem-go metadata is `map[string]string` — no nested objects. New fields must be flat strings.
- **Scale:** 5,500+ talks, 19+ manual collections, all standard works.
- **Batch time:** Full conference reindex ~28 hours at 18.5s/talk sequential. Explore 2-concurrent requests (API supports 4x) to halve this.
- **Cache:** Existing cached summaries invalidated by prompt change. New cache entries use `prompt_version: "v2"`.
- **NOT in scope:** gospel-mcp schema changes (separate proposal: [enriched-search.md](enriched-search.md)). Brain app integration.

---

## Prior Art

| Work | Finding | Status |
|------|---------|--------|
| TITSW v5.1 prompt | MAE 0.93 with context. Best numeric precision. | Complete |
| TITSW v5.4 prompt | Three-axis (modes/categories/insights). MAE 1.30 but richer output. | Complete |
| Context engineering | Context helps scripture (Alma 32 teach 2→6), hurts talks (inflation). | Complete |
| Debug audit | gospel-vec has 3 generic prompts, zero TITSW vocabulary. Lens vs. vocabulary distinction. | Complete |
| Gas Station Insight | MAE is sanity check. Qualitative richness is the product. | Settled |
| Ground truth | 13 pieces scored by Michael. 5 core pieces used for validation. | Available |

---

## Proposed Approach

### The Core Distinction: Lens vs. Vocabulary

| Content Type | Strategy | System Prompt | Context Docs | Why |
|---|---|---|---|---|
| **Scripture** | Lens | Enhanced summary prompt | gospel-vocab.md + titsw-framework.md (~4K tokens) | Scripture is implicit. The model needs a theological lens to see beneath the surface. Proven: Alma 32 teach 2→6 with context. |
| **Talks** | Vocabulary | Enriched summary prompt with TITSW taxonomy | None | Talks are explicit. The speaker says what they mean. Context docs cause score inflation. The model needs the *words* to describe what it already sees. |
| **Manuals (content)** | Vocabulary | Same as talks | None | Come Follow Me, Teachings of Presidents — explicit teaching content. Treat like talks. |
| **Manuals (meta)** | Skip TITSW | Existing generic prompt | None | TITSW manual, Teaching in the Savior's Way — these describe the framework. Scoring them against it is circular. |

### Talk Enrichment: The Vocabulary Approach

Replace the current generic talk summary prompt with one that includes TITSW taxonomy as output format guidance:

**Current system prompt (generic):**
```
Create a summary of this conference talk optimized for semantic search indexing.

Format your response EXACTLY like this:
KEYWORDS: [10-15 comma-separated searchable terms]
SUMMARY: [50-75 word narrative]
KEY_QUOTE: [Most memorable quote]

Keep output under 200 words total. No other text.
```

**Enriched system prompt (with vocabulary):**
```
Create a summary and teaching profile of this conference talk.

FORMAT — output EXACTLY these sections:

KEYWORDS: [10-15 comma-separated terms: doctrines, people, events, themes]
SUMMARY: [50-75 word narrative covering main message and teachings, present tense]
KEY_QUOTE: [Most memorable or powerful direct quote from the talk]

TEACHING_PROFILE:
DOMINANT: [1-2 most prominent from: teach_about_christ, help_come_to_christ, love, spirit, doctrine, invite]
MODE: [primary from: enacted (models it) | declared (testifies) | doctrinal (explains) | experiential (shares experience)]
PATTERN: [brief label: "story→doctrine→invitation" or "problem→principle→promise"]
TEACH_SCORE: [0-9, how central is Christ to the content]
HELP_SCORE: [0-9, how much does this help people come to Christ]
LOVE_SCORE: [0-9, how strongly does this demonstrate or teach love]
SPIRIT_SCORE: [0-9, how directly does this invite the Spirit or bear testimony]
DOCTRINE_SCORE: [0-9, how deeply does this engage with doctrine]
INVITE_SCORE: [0-9, how directly does this invite to specific action]

Scoring guidance:
- 3 = present but not central
- 5 = a clear theme, well-developed
- 7 = a defining feature, specific memorable teaching
- 9 = rare — redefines understanding of this dimension

Keep total output under 300 words. No other text.
```

Key design choices:
- **Include love/spirit scores.** Known to inflate, but tracked anyway — may reveal interesting patterns. Downstream consumers can ignore if noisy.
- **No context documents (baseline).** Proven counterproductive for talks — but see "Talk Context Experiments" below for planned validation.
- **Scoring guidance is minimal** — just 4 anchor points. The model doesn't need the full framework.
- **max_tokens: 500** (up from 300). The TEACHING_PROFILE section adds ~100 tokens of output.

### Scripture Enrichment: The Lens Approach

Inject `gospel-vocab.md` (~1,960 tokens) and `01-titsw-framework.md` (~1,990 tokens) into the scripture summary system prompt, BEFORE the existing instructions. Modify the KEYWORDS instruction to look for typological connections.

### New Metadata Fields

Add to `DocMetadata`:
```go
// TITSW teaching profile fields (conference talks and content manuals)
TitswDominant   string `json:"titsw_dominant,omitempty"`   // e.g. "teach_about_christ,invite"
TitswMode       string `json:"titsw_mode,omitempty"`       // "enacted" | "declared" | "doctrinal" | "experiential"
TitswPattern    string `json:"titsw_pattern,omitempty"`    // "story→doctrine→invitation"
TitswTeach      string `json:"titsw_teach,omitempty"`      // "7" (0-9 score as string)
TitswHelp       string `json:"titsw_help,omitempty"`       // "6"
TitswLove       string `json:"titsw_love,omitempty"`       // "4" (known to inflate — tracked for patterns)
TitswSpirit     string `json:"titsw_spirit,omitempty"`     // "5" (known to inflate — tracked for patterns)
TitswDoctrine   string `json:"titsw_doctrine,omitempty"`   // "7"
TitswInvite     string `json:"titsw_invite,omitempty"`     // "5"
```

These become flat `map[string]string` entries in `ToMap()`, enabling `Where("titsw_mode", "enacted")` queries.

### Cache Format

The summary cache JSON gains a `teaching_profile` object alongside the existing `summary` object. Cache key remains `talk-{year}-{month}-{filename}` but prompt_version changes to `"v2"` — old `"v1"` entries are not overwritten, new entries coexist.

### Parallelism

The current indexer processes talks sequentially (one summary request at a time). LM Studio supports 4 concurrent requests. Implementation:
- Add a configurable `--concurrency` flag (default 1, max 4)
- Use a worker pool with a semaphore channel
- Start with 2 concurrent requests, test for stability and VRAM pressure on dual 4090s
- Monitor for quality degradation (batch inference can sometimes affect model output quality)
- If 2x is stable, try 4x — could reduce batch time from 28 hours to 7

### Talk Context Experiments (Phase 0)

**This is the unexplored territory.** We proved that scripture context (gospel-vocab.md + titsw-framework.md) hurts talks. But we never tested whether a *different kind* of context helps.

The inflation mechanism is specific: `gospel-vocab.md` provides 7 theological patterns for detecting *hidden* Christ-typology (tree of life, types and shadows, etc.). Conference talks are already explicit about Christ. When the model has a vocabulary for finding hidden connections AND the content is already explicit, it over-reads — every mention of Christ gets amplified through multiple theological lenses simultaneously.

But no one ever isolated which context document causes the inflation, or tested talk-specific alternatives. Here's the experiment matrix:

#### Experiment Set

**Test pieces:** Kearon "Receive His Gift", Bednar "Their Own Judges", Holland "And Now I See" (all have ground truth scores)

**Prompt:** Enriched talk prompt (the one above with TEACHING_PROFILE) for all experiments

| Exp | Context | Hypothesis | What it tests |
|-----|---------|------------|---------------|
| T0 | None (vocabulary only) | Baseline. Vocabulary approach works. | This is the default proposed approach. |
| T1 | titsw-framework.md only | Framework without typological lens may help without inflation. | Isolates gospel-vocab.md as inflation source. |
| T2 | gospel-vocab.md only | Typological lens alone — probably causes most of the inflation. | Confirms/denies gospel-vocab is the culprit. |
| T3 | Talk-specific rhetorical context | Helps the model identify teaching modes and patterns. | New context type — rhetorical rather than theological. |
| T4 | Calibration context (scored examples) | Few-shot anchoring reduces score variance. | Does an exemplar ground the model? |
| T5 | Best of T1-T4 combined | Optimal context stack for talks. | Combination effect. |

#### Talk-Specific Context Document (for T3)

A new document describing conference talk conventions. Not what to look for spiritually (that's gospel-vocab's job) but what rhetorical patterns indicate what teaching mode:

```
CONFERENCE TALK PATTERNS

Conference talks follow recognizable rhetorical structures:

TEACHING MODES — how the speaker conveys truth:
- ENACTED: The speaker demonstrates the principle by doing it, not just describing it.
  (Elder Kearon welcomed members with joy rather than listing reasons for joy.)
- DECLARED: Direct testimony — "I know..." "I bear witness..."
- DOCTRINAL: Systematic exposition of doctrine with scriptural support.
- EXPERIENTIAL: Personal narrative as the primary vehicle for truth.

STRUCTURAL PATTERNS — how talks flow:
- Story → Doctrine → Invitation (most common)
- Problem → Principle → Promise
- Question → Exploration → Testimony
- Narrative arc (sustained story with embedded doctrine)
- Systematic exposition (topic-by-topic)

CALIBRATION:
- Most conference talks score 4-6 on teach_about_christ. Christ is the default subject.
- A teach score of 7+ means Christ isn't just mentioned — He's the specific, developed content.
- Most talks score 3-4 on invite. A specific, actionable invitation scores 7+.
- love and spirit inflate easily. A 7+ means the dimension is enacted, not just referenced.
```

#### Calibration Context (for T4)

A scored example from ground truth, showing the model what appropriate scores look like:

```
CALIBRATION EXAMPLE

Speaker: Elder Patrick Kearon, "Welcome to the Church of Joy"
DOMINANT: help_come_to_christ, love
MODE: enacted
PATTERN: invitation→doctrine→testimony
TEACH_SCORE: 5  (Christ present but not the central developed content)
HELP_SCORE: 7  (explicit, sustained focus on helping people come to Christ through belonging)
LOVE_SCORE: 7  (models warmth and welcome — enacted, not just stated)
SPIRIT_SCORE: 5  (bears testimony, invites the Spirit)
DOCTRINE_SCORE: 4  (some doctrinal grounding but primarily pastoral)
INVITE_SCORE: 7  (specific invitation to welcome, belong, stay)

Note: Scores cluster 4-6 for most dimensions. A 7+ is exceptional and specific.
```

#### Experiment Protocol

1. Run all 6 experiments on 3 ground-truth talks (18 test runs)
2. Compare scores against ground truth — MAE per experiment
3. Also inspect qualitative output: are modes/patterns/dominant labels better with any context?
4. If any context approach beats T0 (baseline), incorporate into Phase 1 before batch run
5. If T0 wins, the vocabulary-only approach is confirmed and we proceed as designed

**This is Phase 0 — run BEFORE Phase 1.** If context helps talks after all, we want to know before burning 28 hours on the wrong approach. 18 test runs at ~18.5s each = ~6 minutes total.

---

## Phased Delivery

### Phase 0: Talk Context Experiments (1 session, ~30 min)

**Delivers:** Data-driven decision on whether talks benefit from ANY context, and which kind.

1. Create talk-rhetorical-context.md and talk-calibration-context.md in `experiments/lm-studio/scripts/context-talk/`
2. Run 6 experiments × 3 ground-truth talks (18 runs, ~6 minutes)
3. Compare MAE and qualitative output against T0 baseline
4. Decision: vocabulary-only (T0) or vocabulary + best context for Phase 1

**Phase 0 gates Phase 1.** We don't want to burn 28 hours on a suboptimal prompt.

### Phase 1: Talk Pipeline Enrichment (1 session)

**Delivers:** TITSW-enriched talk summaries with teaching profile metadata.

1. Add TITSW metadata fields to `DocMetadata` and `ToMap()` (10 new fields including love/spirit/doctrine)
2. Write `generateTalkSummaryV2()` with enriched prompt (+ best context from Phase 0 if any beat baseline)
3. Write `parseTalkTeachingProfile()` to extract TEACHING_PROFILE fields from response
4. Update `ChunkTalkAsSummary()` to populate new metadata fields
5. Add prompt version tag to cache format
6. Add `--concurrency` flag to indexer (default 1, test with 2)
7. **Test:** Run on 5 ground-truth talks. Compare scores to established targets (±2 tolerance).
8. **Verify:** Inspect 3 summaries manually for quality of mode/pattern/dominant labels.

**Phase 1 stands alone.** Even without scripture or manual enrichment, enriched talk summaries immediately improve search quality for conference content.

### Phase 2: Scripture Pipeline Enrichment (1 session)

**Delivers:** Deeper scripture summaries with typological connections.

1. Modify `SummarizeChapter()` to accept optional context documents
2. Load `gospel-vocab.md` and `01-titsw-framework.md` at indexer startup
3. Inject as prefix to scripture summary system prompt
4. Updated KEYWORDS instruction for typological depth
5. **Test:** Run on Alma 32, Zechariah 3, 1 Nephi 11, Genesis 22, D&C 121. Compare to current cached summaries.

### Phase 3: Manual Pipeline + Theme Detection (1 session)

**Delivers:** Enriched manual summaries, talk theme detection.

1. Classify `KnownManuals()` list into content vs. meta-teaching (meta-teaching manuals like TITSW skip TITSW scoring — evaluating the framework against itself is circular)
2. Apply talk-style enrichment to content manuals
3. Add `DetectTalkThemes()` — identify rhetorical sections in talks (story, doctrine, application, invitation, testimony). This is the talk-level analog of `DetectThemes()` for scripture.
4. **Test:** Run on 3 CFM lessons, 2 Teachings of Presidents chapters. Inspect theme sections on 5 ground-truth talks.

### Phase 4: gospel-mcp Integration (separate proposal)

See [enriched-search.md](enriched-search.md) — full separate proposal covering:
- Where to add FTS on enriched fields (gospel-mcp SQLite vs. adding SQLite to gospel-vec)
- Schema changes for TITSW metadata
- MCP tool updates to expose teaching profiles in search results
- Architectural question: one DB or two?

---

## Verification Strategy

| Phase | Verification | Criteria |
|---|---|---|
| 0 | Context A/B comparison | MAE per experiment, qualitative mode/pattern quality |
| 0 | Decision gate | Best approach identified before Phase 1 batch run |
| 1 | Ground truth scores | 5 talks score within ±2 of Michael's targets |
| 1 | Manual inspection | Mode/pattern/dominant labels are sensible for 3+ talks |
| 1 | Backward compat | KEYWORDS/SUMMARY/KEY_QUOTE still present and well-formed |
| 1 | Parallelism test | 2-concurrent runs produce same quality as sequential |
| 2 | Before/after compare | Alma 32 keywords include typological terms not in current summary |
| 2 | No regression | Summary quality doesn't degrade on simple chapters |
| 3 | Manual classification | Meta-teaching manuals correctly skipped for TITSW scoring |
| 3 | Theme detection | Talk rhetorical sections identified for 3+ ground-truth talks |
| All | Full reindex | Completes without errors on entire corpus |

---

## Costs and Risks

| Cost | Impact | Mitigation |
|------|--------|------------|
| Cache invalidation | All 5,500 talk summaries need regeneration | One-time cost. ~14-28 hours depending on parallelism. Run overnight. |
| Token increase | max_tokens 300→500 per call | Within model capacity. Negligible cost increase. |
| Metadata size | 6 new fields per talk chunk | Trivial storage impact. ~50 bytes per chunk. |
| Prompt fragility | Model may not consistently follow TEACHING_PROFILE format | Parse with fallback — if fields are missing, store empty. Don't fail the summary. |
| Score noise | Individual scores may be ±2 from ground truth | Acceptable. Scores are sanity checks, not the product. |

**What gets worse:** The only real cost is the 28-hour reindex. Everything else is net positive — richer metadata, better search, same API surface.

**What if it goes wrong:** The cache preserves old summaries (v1 tag). If the enriched prompt produces worse output, roll back by pointing at the v1 cache. No data is destroyed.

---

## Creation Cycle Review

| Step | Question | This Proposal |
|------|----------|---------------|
| Intent | Why? | Better search and study. Michael's core priority. |
| Covenant | Rules? | Existing code conventions. Go style. gospel-vec patterns. |
| Stewardship | Who owns it? | dev agent executes. Michael reviews output quality. |
| Spiritual Creation | Spec precise enough? | Yes — prompts, fields, test criteria all specified. |
| Line upon Line | Phasing? | 4 phases. Phase 1 stands alone. |
| Physical Creation | Who executes? | dev agent. Phase 1 is one session. |
| Review | How do we know? | Ground truth comparison + manual inspection. |
| Atonement | If it goes wrong? | Cache rollback. v1 summaries preserved. |
| Sabbath | When to pause? | After Phase 1 — evaluate before continuing. |
| Consecration | Who benefits? | Michael directly. Brain app future. Any downstream consumer. |
| Zion | Integration? | gospel-vec enriched → gospel-mcp sync (Phase 4) → brain app search. |

---

## Recommendation

**Build — Phase 1 first, then evaluate.**

This is the natural next step from completed prompt optimization. The binding problem is real (we see it in the cached summaries). The data supports the approach (two-pipeline strategy proven). The scope is contained (one prompt change + metadata fields per phase). Phase 1 is a single dev session.

Begin with the talk pipeline. Test against ground truth. If the enriched summaries are good, proceed to Phase 2 (scripture) and Phase 3 (manuals). Phase 4 (gospel-mcp) is a separate proposal.

**Hand off to:** dev agent, with this proposal as the spec.
