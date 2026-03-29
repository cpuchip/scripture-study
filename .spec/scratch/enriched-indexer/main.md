# Scratch: Enriched Indexer Design

*Created: 2026-03-29*
*Proposal: [.spec/proposals/enriched-indexer.md](../../proposals/enriched-indexer.md)*
*Parent work: [.spec/scratch/debug-titsw-optimization/main.md](../debug-titsw-optimization/main.md)*

---

## Binding Problem

The gospel-vec indexer produces generic summaries ("faith, repentance, atonement") that don't capture TITSW teaching dimensions. We proved the model CAN extract rich teaching profiles with the right framing — but the existing prompts contain zero TITSW vocabulary. The result: summaries that surface nothing useful for teaching-oriented search and study.

The fix requires different strategies per content type:
- **Scripture:** Context documents (lens) help the model see beneath the surface
- **Talks:** TITSW vocabulary in the system prompt (not context docs, which inflate scores)
- **Manuals:** TBD — likely vocabulary approach since they're explicit like talks

---

## Research Phase — Existing State Inventory

### Current gospel-vec Prompts (from debug audit)

| Content | Prompt | Output Format | Temp | MaxTok |
|---------|--------|---------------|------|--------|
| Scripture summary | "Create a summary optimized for semantic search" | KEYWORDS/SUMMARY/KEY_VERSE | 0.2 | 300 |
| Scripture themes | "Identify narrative sections, return JSON" | [{range, theme}] | 0.2 | 300 |
| Talk summary | "Create a summary of this conference talk optimized for semantic search" | KEYWORDS/SUMMARY/KEY_QUOTE | 0.2 | 300 |
| Manual summary | "Create a summary of this manual chapter optimized for semantic search" | KEYWORDS/SUMMARY/KEY_QUOTE | 0.2 | 300 |

All share the same generic framing. No TITSW vocabulary anywhere.

### Cached Summary Quality (sampled Mar 29)

**Kearon 2024 (talk):**
- Keywords: sacrament, joy, worship, Elder Patrick Kearon, Church of Joy, Heavenly Father, Jesus Christ, repentance, gratitude, spiritual renewal, sacrament meeting, gospel, testimony, sacrifice, resurrection, family councils, service, reverence, love, peace
- Summary: 80 words. Decent but generic. "Joy is found in daily repentance, gratitude, and recognizing Christ's love."
- No teaching mode, no dimension identification, no distinction between enacted/declared

**Holland 2020 (talk):**
- Keywords: hope, Restoration, gospel of Jesus Christ, Elder Jeffrey R. Holland, 200 years, revelation, eternal life, love of God, missionary work, divine promises
- Summary: 50 words. Very thin. "emphasizes the ongoing work of God, the blessings of the gospel"
- No rhetorical structure, no teaching pattern identification

**Zechariah 3 (scripture):**
- Keywords: 17 terms including Zechariah, Joshua, high priest, Satan, Branch, covenant
- Summary: 80 words. Better — mentions cleansing symbolism, the Branch/Messiah
- Themes: 3 sections with ranges. Good structural decomposition.
- Missing: no typological connection to Christ, no "the Branch IS the Messiah" depth

### What TITSW-Enriched Output Would Add

For Kearon 2024, a TITSW-enriched summary would identify:
- **Dominant dimensions:** Teach About Christ (high), Help Come to Christ (high)
- **Teaching mode:** Enacted (models joy rather than just describing it) + Declared (testifies of resurrection)
- **Category:** Experiential — rooted in worship experience rather than doctrinal exposition
- **Insight:** The "church of joy" framing connects sacrament worship to Christ's victory (resurrection → joy)

For Zechariah 3, with context documents:
- **Typological depth:** The Branch = Christ. Filthy garments = sin. Clean garments = atonement. Joshua = type of Christ as high priest.
- **Cross-reference density:** 1 Ne 11:22-25 (tree of life = love of God = Christ), Hebrews (Christ as high priest)

### Data Flow Understanding (from Explore agent)

1. `IndexConferenceTalks()` →  `ParseTalkFile()` → for each talk:
   - `LayerParagraph`: `ChunkTalkByParagraph()` → paragraph chunks
   - `LayerSummary`: `generateTalkSummary()` → `ChunkTalkAsSummary()` → single summary chunk

2. `generateTalkSummary()` calls `idx.summarizer.chat()` with:
   - System prompt: generic "Create a summary" + format spec (KEYWORDS/SUMMARY/KEY_QUOTE)
   - User prompt: talk metadata (speaker, year, month, title) + first 25 paragraphs
   - Returns `*ChapterSummary` with Keywords, Summary, KeyVerse fields

3. `ChunkTalkAsSummary()` composes chunk content: summary text + key quote + topics
   - Stored in chromem-go with `DocMetadata` (Source, Layer, Speaker, Year, Month, etc.)
   - Metadata is `map[string]string` — flat key-value pairs only

4. Key constraints:
   - chromem-go metadata = `map[string]string` — no nested objects
   - Summary cache = JSON files on disk with model version
   - Changing prompts means re-running summaries (cache invalidated by model name change or manual clear)
   - 5,500+ talks to process. At 18.5s avg, full reindex = ~28 hours

### DocMetadata Fields Available

```go
type DocMetadata struct {
    Source, Layer, Book, Chapter, Reference, Range, FilePath string
    Generated, Model, Timestamp string
    // Conference-specific:
    Speaker, Position, Year, Month, Session, TalkTitle string
}
```

New TITSW fields would need to be added to this struct and the `ToMap()` method.

---

## Design Decisions

### Decision 1: Two-Output vs. Single-Output for Talks

**Option A: Two separate prompts (summary + teaching profile)**
- Pro: Clean separation of concerns. Existing summary unchanged.
- Pro: Can iterate teaching profile independently.
- Con: 2x LLM calls per talk. At 18.5s each, adds 28 hours to full reindex.
- Con: Two cache entries per talk.

**Option B: Single enriched prompt**
- Pro: One LLM call per talk. Same cost as today.
- Pro: Single cache entry.
- Con: More complex prompt. If teaching profile regresses, can't fix without also re-running summary.
- Con: Larger max_tokens needed (300 → ~600).

**Decision: Option B — single enriched prompt.** The existing prompt is already thin (300 tokens). Adding TITSW fields to the same prompt keeps the call count at 1. The cache invalidation concern is real but acceptable — we're doing a full reindex anyway when we change prompts. The max_tokens increase from 300→600 is within model capacity.

### Decision 2: What TITSW Fields to Store

From the v5.4 schema (the three-axis design), these are the fields that matter for downstream search:

| Field | Type | Why |
|-------|------|-----|
| dominant_dimensions | string (comma-sep) | "What is this talk mostly about?" |
| teaching_mode | string | "enacted" / "declared" / "doctrinal" / "experiential" |
| teach_about_christ | int (0-9) | Sanity check score |
| help_come_to_christ | int (0-9) | Sanity check score |
| love | int (0-9) | Can omit — scores proved unreliable on this dimension |
| invite | int (0-9) | Sanity check score |
| teaching_pattern | string | Brief label: "story→doctrine→application" |

**Not storing (too noisy / not useful for search):**
- Per-dimension modes/categories (the full v5.4 three-axis was too heavy)
- typological_depth (only meaningful for scripture)
- cross_reference_density (only meaningful for scripture)
- surface_vs_deep_delta (diagnostic, not search-useful)
- insights array (too unstructured for metadata search)

### Decision 3: How to Store in chromem-go

chromem-go metadata is `map[string]string`. Options:

**Option A: Flat metadata fields**
```go
"titsw_dominant": "teach_about_christ,help_come_to_christ"
"titsw_mode": "enacted"
"titsw_teach": "7"
"titsw_help": "6"
"titsw_invite": "5"
"titsw_pattern": "testimony-invitation"
```

**Option B: JSON blob in one metadata field**
```go
"titsw": `{"dominant":"teach,help","mode":"enacted","teach":7,"help":6,"invite":5}`
```

**Decision: Option A — flat fields.** chromem-go can filter on metadata values. Flat fields enable `Where("titsw_mode", "enacted")` queries directly. A JSON blob requires parsing after retrieval.

### Decision 4: Talk Prompt — Vocabulary, Not Lens

From the debug audit: "for talks, give the model a vocabulary (terminology in the system prompt)."

The enriched talk system prompt should:
1. Keep the existing KEYWORDS/SUMMARY/KEY_QUOTE format
2. ADD a TEACHING_PROFILE section with structured fields
3. Include TITSW dimension definitions as a brief taxonomy (not the full framework doc)
4. Include teaching mode definitions (enacted/declared/doctrinal/experiential)
5. NOT include gospel-vocab.md or 01-titsw-framework.md as context (proven counterproductive)

Estimated token increase for system prompt: ~400 tokens (taxonomy definitions). That's fine.

### Decision 5: Scripture Pipeline Changes

Scripture summaries should get the gospel-vocab.md and 01-titsw-framework.md injected as context. This was proven to deepen extraction (Alma 32 teach 2→6 with context).

But this is a separate change from the talk pipeline. Phase separately.

### Decision 6: Manual Pipeline

Manuals fall into two categories:
1. **Teaching materials about teaching** (TITSW manual, Teaching in the Savior's Way) — these are meta-content. They describe the framework. Scoring them AGAINST the framework is circular.
2. **Content manuals** (Come Follow Me, Teachings of Presidents) — these contain lesson content that can be evaluated for teaching dimensions.

For now: treat content manuals like talks (vocabulary approach, no context). Flag meta-manuals to skip TITSW scoring entirely. This is a Phase 2 refinement.

---

## Draft Talk System Prompt

```
Create a summary and teaching profile of this conference talk.

FORMAT — output EXACTLY these sections:

KEYWORDS: [10-15 comma-separated terms: doctrines, people, events, themes]
SUMMARY: [50-75 word narrative covering main message and teachings, present tense]
KEY_QUOTE: [Most memorable or powerful direct quote from the talk]

TEACHING_PROFILE:
DOMINANT: [1-2 most prominent from: teach_about_christ, help_come_to_christ, love, spirit, doctrine, invite]
MODE: [primary from: enacted (models it) | declared (testifies of it) | doctrinal (explains it) | experiential (shares personal experience)]
PATTERN: [brief label of rhetorical flow, e.g. "story→doctrine→invitation" or "problem→principle→promise"]
TEACH_SCORE: [0-9, how central is Christ to the content?]
HELP_SCORE: [0-9, how much does this help people come to Christ?]
INVITE_SCORE: [0-9, how directly does this invite to specific action?]

Scoring guidance:
- 3 = present but not central
- 5 = a clear theme, well-developed
- 7 = a defining feature with specific, memorable teaching
- 9 = rare — a talk that redefines how you understand this dimension

Keep output under 300 words total. No other text.
```

Estimated output tokens: ~250-350. Set max_tokens to 500.

---

## Draft Scripture System Prompt (with context)

For scripture summaries, inject context documents BEFORE the existing prompt:

```
[CONTEXT — Theological Framework]
{contents of gospel-vocab.md}

[CONTEXT — Teaching Dimensions]
{contents of 01-titsw-framework.md}

Create a summary of this scripture chapter optimized for deep study and semantic search.

FORMAT:
KEYWORDS: [10-15 terms including typological connections, not just surface topics]
SUMMARY: [50-75 words — go beneath surface narrative to theological architecture]
KEY_VERSE: [Most theologically significant verse with brief note on why]

When identifying keywords, look for:
- Types and shadows of Christ (people, objects, events that prefigure Christ)
- Connections to the Doctrine of Christ (faith, repentance, baptism, Holy Ghost, endurance)
- Patterns from the theological framework provided above

Keep output under 200 words total. No other text.
```

This adds ~4,000 tokens to the system message (context docs). For scripture chapters (typically 500-2000 words), total context stays well within model capacity.

---

## Implementation Phases

### Phase 1: Talk Pipeline Enrichment
1. Add TITSW metadata fields to `DocMetadata` struct and `ToMap()`
2. Write new `generateTalkSummaryV2()` with enriched prompt
3. Add `parseTalkTeachingProfile()` to extract TEACHING_PROFILE fields
4. Update `ChunkTalkAsSummary()` to include new metadata
5. Update summary cache format (v2 prompt version tag)
6. Test on 5-10 talks against ground truth
7. Run full conference reindex

### Phase 2: Scripture Pipeline Enrichment
1. Inject context documents into `SummarizeChapter()` system prompt
2. Update KEYWORDS/SUMMARY/KEY_VERSE prompt for typological depth
3. Test on 5-10 chapters (Alma 32, Zechariah 3, 1 Nephi 11, etc.)
4. Run full scripture reindex

### Phase 3: Manual Pipeline + Refinement
1. Classify manuals: content vs. meta-teaching
2. Apply talk-style enrichment to content manuals
3. Skip TITSW scoring for meta-teaching manuals
4. Theme detection for talks (currently scripture-only) — stretch goal

### Phase 4: gospel-mcp Schema Updates
1. Add TITSW columns to talks/manuals tables
2. Enable searching by teaching mode, dominant dimension
3. Update MCP tools to expose new fields

---

## Open Questions (RESOLVED Mar 29)

~1. **Love/Spirit dimension reliability.**~ → **DECIDED: Keep them.** May reveal interesting patterns. Downstream can ignore if noisy. Added to prompt and metadata fields.
~2. **Theme detection for talks.**~ → **DECIDED: Yes, Phase 3.** Not stretch goal — core feature. Michael originally wanted it.
~3. **Cache migration.**~ → **DECIDED: Full reindex.** 28 hours acceptable. Pushed for parallelism.
~4. **Batch time.**~ → **DECIDED: Explore concurrency.** LM Studio supports 4x. Try 2 concurrent, then 4. Could cut to 7-14 hours.
~5. **gospel-mcp sync.**~ → **DECIDED: Separate proposal.** [enriched-search.md](../../proposals/enriched-search.md). Option C (gospel-mcp reads gospel-vec cache). gospel-vec has NO SQLite — confirmed.

---

## New Research (Mar 29 Session 2)

### gospel-vec Architecture Confirmation
- **No SQLite in gospel-vec.** Pure chromem-go. go.mod has only `philippgille/chromem-go` as dependency.
- Storage: `storage.go` uses `chromem.NewDB()` (in-memory) + `ExportToFile()`/`ImportFromFile()` with gob.gz format.
- Four per-source files: `scriptures.gob.gz`, `conference.gob.gz`, `manual.gob.gz`, `music.gob.gz`.
- No SQL queries anywhere. No FTS capability.

### Talk Context Experiment Gap (CRITICAL FINDING)
Explored full experiment history. Found that the "context hurts talks" conclusion was based on applying **scripture-focused context** to talks. Specifically:
- `gospel-vocab.md` — 7 theological patterns for detecting *hidden* Christ-typology
- `01-titsw-framework.md` — scoring anchors with scripture-flavored examples

**What was NEVER tested:**
1. Only `titsw-framework.md` without `gospel-vocab.md` (isolating the inflation source)
2. Only `gospel-vocab.md` without `titsw-framework.md` 
3. Talk-specific rhetorical context (teaching modes, structural patterns, calibration anchors)
4. Talk-specific calibration examples (scored ground-truth as few-shot anchor)
5. Anti-inflation calibration ("most talks score 4-6 on teach_about_christ")

**Inflation mechanism identified:** `gospel-vocab.md` provides patterns to find *hidden* connections. Talks are already explicit. When the model has a toolkit for finding hidden things AND the content is already explicit, it over-reads — every mention amplifies through multiple theological lenses.

**Experiment plan added to proposal:** Phase 0 with 6 controlled experiments (T0-T5) on 3 ground-truth talks. Run BEFORE Phase 1 to validate or improve the vocabulary-only approach. ~6 minutes total.

### gospel-mcp Architecture (for integration proposal)
- Go + SQLite + FTS5. Three MCP tools: `gospel_search`, `gospel_get`, `gospel_list`.
- `talks` table: year, month, session, speaker, title, content, file_path, source_url. Zero TITSW fields.
- `talks_fts`: FTS5 on title, speaker, content.
- Schema version tracking in `schema_version` table. Migration path via `ALTER TABLE`.
- **Recommended integration:** Option C — gospel-mcp reads gospel-vec's summary cache JSON during its own index step. Simplest. No cross-process coordination. Separate proposal at [enriched-search.md](../../proposals/enriched-search.md).

