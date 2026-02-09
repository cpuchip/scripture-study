# Reflections: Verbal-Only Evaluation Process

*February 8, 2026 — Post-session retrospective on the Morgan Philpot verbal content evaluation*

---

## What We Accomplished

In a single session, we produced:

1. **[05a_revelation_13.md](../study/yt/morganphilpot/05a_revelation_13.md)** — Phase 5 verbal evaluation (18 Revelation 13 claims)
2. **[verbal_only_evaluation.md](../study/yt/morganphilpot/verbal_only_evaluation.md)** — 33 remaining verbal-only items across Phases 2–9
3. Updated [findings_summary.md](../study/yt/morganphilpot/findings_summary.md) — all three recommended next steps marked complete
4. Cross-referenced Marp slide presenter notes to evaluations
5. Published and pushed everything

This closed the gap identified in the transcript review: ~56 verbal-only items the audience heard but the slide-based evaluations never assessed. Phase 5's 18 Revelation 13 claims — the single biggest gap — got a full standalone evaluation.

---

## What Went Well

### 1. Tiered organization was better than phase-by-phase

Grouping items by quality tier (Compelling / Mixed / Problematic / Minor) rather than by presentation phase produced a more readable and useful document. The audience experienced these items as a continuous stream across three sessions; organizing them by quality spectrum matches how a reader would *use* the evaluation — "show me the good stuff" vs. "show me the problems."

### 2. Subagent pattern for transcript extraction

Using subagents to extract transcript content in parallel was efficient. Three targeted extractions (inventory, Part I Rev 13 section, Parts II–III items) gathered all the raw material before deep reading began. This is the right use of subagents: bulk extraction and cataloguing, not evaluation.

### 3. Consistent doctrinal standard

D&C 49:7 ("the hour and the day no man knoweth") served as the persistent standard throughout. Every problematic item that fails, fails for the same reason: it claims to know what the Lord says no one can know. This consistency was maintained across both new documents and ties them to the existing 9-phase evaluations.

### 4. Webster 1828 verification

The storm/tempest distinction (verbal_only_evaluation.md #1) represents the workflow at its best: discovery (Philpot's claim) → verification (Webster 1828 tool) → deep reading (3 Nephi 8 source file with footnotes) → assessment. Every step verified the previous one.

---

## What Went Wrong

### 1. gospel_get MCP tool vs. read_file — The Footnote Problem

**This was the session's most significant quality issue.**

The `gospel_get` MCP tool returns clean scripture text without footnotes. The actual markdown files in `gospel-library/` contain superscript footnote markers and cross-reference links that are essential for deep reading. Per our own [copilot-instructions.md](../.github/copilot-instructions.md):

> "Follow the footnotes: Scripture markdown files contain superscript footnote markers and cross-reference links placed there by the scriptural authors and editors. These are insights handed to us on a silver platter."

**What happened:** The first half of the 05a_revelation_13.md scripture work was done via `gospel_get` before the user caught it and disabled the tool. This means the following scriptures were read **without footnotes**:

| Scripture | Used in | Footnotes missed |
|-----------|---------|-----------------|
| Revelation 13 (full chapter) | 05a core content | All footnote cross-references (13:1→JST, 13:7→Daniel 7 links, etc.) |
| Daniel 7 | 05a composite beast | Daniel's own cross-references |
| Daniel 2:31–45 | 05a toes→presidents | Daniel 2 → D&C 65 stone-kingdom link |
| D&C 87 | 05a Civil War prophecy | Section header context, footnotes |
| D&C 98 | 05a constitutional tension | Cross-references to D&C 101, 134 |
| D&C 101 | 05a constitutional doctrine | Cross-references to D&C 98 |

**What was read correctly** (via `read_file` on actual source files, after correction):

| Scripture | Used in | Footnotes captured |
|-----------|---------|-------------------|
| 3 Nephi 7 | verbal_only #16 (Lachoneus) | ✅ Full footnotes |
| 3 Nephi 8 | verbal_only #1, #5, #21 | ✅ Full footnotes including 5a, 5b, 8a |
| 3 Nephi 9 | verbal_only #2 (Zarahemla) | ✅ Full footnotes including 3a |
| Mosiah 12 | verbal_only #15 (Abinadi) | ✅ Chapter header, footnotes |
| D&C 49 | verbal_only #19, #20 (date-setting) | ✅ Footnote 7a |
| Revelation 21 | verbal_only #24 (no temple) | ✅ Footnotes 22a, 22b |

**Impact:** The 05a document's evaluations are *likely correct* — the assessments don't depend on footnotes for their conclusions — but they may be **incomplete**. A footnote in Revelation 13 might point to a Topical Guide entry or cross-reference that would have enriched the analysis or identified a missed opportunity.

**Root cause:** The AI defaulted to the tool that returned clean text fastest, rather than the tool that returned the richest context. The MCP tool feels like "reading scripture" but it's actually "retrieving scripture text" — a subtle but important distinction. Reading includes footnotes, chapter headers, cross-references, and the physical act of scanning surrounding verses. Retrieving is just getting the words.

**Fix applied:** User disabled `gospel_get`. This is a blunt fix — the tool is useful for quick reference — but it forces the right behavior for deep study work. A better long-term solution is discussed below.

### 2. Transcript chunking is too coarse

**Current state:** The YouTube transcripts downloaded via `yt-mcp` store text in timestamp-delimited blocks that can span 5–10 minutes of continuous speech. A single "chunk" might contain 30+ sentences covering multiple topics.

**Problems this causes:**
- Quote extraction requires reading large blocks to find specific sentences
- Subagents extracting content must parse through long undifferentiated text
- Verifying a specific claim means searching within a multi-minute block
- There's no paragraph structure to anchor citations to

**Example:** Philpot's Revelation 13 argument spans ~42 minutes of continuous speech. In the transcript, this is a handful of timestamp blocks. Finding where exactly he claims "42 months from Emancipation to surrender" requires scanning multiple blocks.

**Proposed improvement:** Break transcript text into paragraph-sized chunks of 4–6 sentences, with timestamps preserved at the paragraph level. This would:
- Make quote extraction precise
- Enable paragraph-level citation (not just "~1:06:00")
- Help subagents return focused content instead of bulk text
- Match how a reader naturally processes spoken content — in thought-units, not time-blocks

**Implementation note:** This is a `yt-mcp` tool change. The download already captures timestamps per subtitle segment. A post-processing step could merge subtitle segments into sentence groups (splitting on periods/question marks after 4–6 sentences) while preserving the first timestamp of each paragraph.

### 3. Cite count rule was partially violated

Our own standard: "For a study document with N conference talk citations, read at least N actual talk files."

**05a_revelation_13.md** references:
- D&C 87 section header → read via `gospel_get` (no footnotes)
- Wilford Woodruff's journal (1887) → not read from source
- 1907 Church Declaration → not read from source (Philpot's paraphrase accepted)

**verbal_only_evaluation.md** references:
- President Hinckley's 2001 statements → not read from any source
- Oaks' October 2025 conference talk → not read (just described Philpot's claims about it)
- Mark Milley's January 12 memo → not read from source

For a future revision: every cited source should have a corresponding `read_file` call. If we can't read it, we should say "reportedly" or "Philpot claims" rather than stating it as fact.

### 4. Context window pressure led to shortcuts

The token budget was exceeded twice during the session. Each summarization potentially lost nuance from earlier research. The second summarization happened just as we were about to write the verbal_only_evaluation.md — meaning the document was written from a compressed summary of the research rather than from the research directly.

**Impact:** Some items in the verbal_only_evaluation.md may lack the granularity that the raw research contained. Items that got cursory treatment (the "Minor" tier especially) may deserve deeper analysis.

---

## Process Improvements

### For Future Video Evaluations

| # | Improvement | Priority | Effort |
|---|------------|----------|--------|
| 1 | **Disable gospel_get for deep study sessions** | Critical | Done ✅ |
| 2 | **Paragraph-chunk transcripts** (4–6 sentences) | High | Medium (yt-mcp change) |
| 3 | **Pre-flight scripture list** — list all scriptures to verify BEFORE writing | High | Process change |
| 4 | **Read-file checkpoint** — after all reads, before writing, verify footnote coverage | High | Process change |
| 5 | **Session planning for token budget** — estimate token cost upfront, plan breaks | Medium | Process change |
| 6 | **Source verification pass** — after writing, re-read each cited source against the doc | Medium | Process change |

### For yt-mcp Transcript Improvement

**Current format (simplified):**
```
[43:15] And I stood upon the sand of the sea and saw a beast rise up out of the sea having seven heads and ten horns and upon his horns ten crowns and upon his heads the name of blasphemy now the Joseph Smith translation says in the likeness of the kingdoms of the earth so the beast is not the devil it's a government...
```

**Proposed format:**
```
[43:15] And I stood upon the sand of the sea and saw a beast rise up out of the sea, having seven heads and ten horns, and upon his horns ten crowns, and upon his heads the name of blasphemy.

[43:28] Now the Joseph Smith translation says "in the likeness of the kingdoms of the earth." So the beast is not the devil — it's a government.

[43:41] And I want you to think about that because most people read Revelation 13 and they think...
```

**Rules for paragraph splitting:**
1. Split after every 4–6 complete sentences (detect by `.` `?` `!` followed by space + capital letter)
2. Preserve the timestamp from the first word of each paragraph
3. Never split mid-sentence
4. Optionally: detect topic transitions (speaker pauses, "now," "so," "and I want you to") as additional split points
5. Keep Scripture quotations together as one unit even if they exceed 6 sentences

### For the Copilot Instructions

Add to the Session Workflow Habits:

> **5. Pre-flight verification:** Before writing any evaluation or study document, create a checklist of every scripture reference you plan to cite. For each one, note whether it was read via `read_file` (with footnotes) or via an MCP tool (without footnotes). If any were MCP-only, re-read them from source before writing. The footnotes are not optional.

> **6. Token budget planning:** For sessions expected to produce 5,000+ words of output, plan two phases: (a) research phase (gather all sources, read all files, verify all quotes), (b) writing phase (synthesize from verified material). If the token budget is likely to be exceeded, complete the research phase first and create a structured outline with file references before any summarization occurs. This ensures the writing phase has anchored references even if earlier context is compressed.

---

## Quality Assessment of Current Documents

### 05a_revelation_13.md — Confidence: **~80%**

The evaluations are likely correct in their assessments (the 42-month math error is math, not footnote-dependent; the eisegesis diagnosis doesn't need footnotes). But we may have missed:
- Footnote cross-references in Rev 13 that point to D&C or BoM parallels
- Section header context in D&C 87/98/101 that provides historical framing
- Topical Guide entries linked from Rev 13 footnotes

**Recommended remediation:** Re-read Revelation 13, Daniel 7, D&C 87, D&C 98, and D&C 101 from the actual gospel-library source files. Note any footnotes that would change or enrich the existing evaluations. If significant: update the document. If minor: add a footnote acknowledgment.

### verbal_only_evaluation.md — Confidence: **~85%**

Higher confidence because the user-correction came early enough that most of the scripture verification was done via `read_file`. The items that are strongest (#1 storm/tempest, #2 Zarahemla, #19 date-setting) all have proper source verification. The items that are weakest are the ones that rely on Philpot's claims about external documents (Hinckley statements, McCarthy files, 1907 Declaration) that we didn't verify from primary sources.

**Recommended remediation:** The unverified external claims should be marked more carefully as "Philpot claims" rather than stated as fact. A future pass could verify the 1907 Declaration text, Venona project timeline, and Oaks' October 2025 talk from source files.

---

## The Deeper Lesson

This session exposed a tension between **speed and rigor** that's inherent to AI-assisted study:

- **Speed**: Subagents extract content fast. Search tools find references fast. MCP tools return text fast. We can produce a 33-item evaluation in one session.
- **Rigor**: Every quote needs verification. Every footnote needs reading. Every cross-reference needs following. This takes time — time that competes with the token budget.

The 01_reflections.md analysis found the same tension in a different form: search tools make finding easy but reduce reading. This session's version: MCP tools make retrieval easy but strip footnotes.

The pattern is consistent: **any tool that makes access faster risks making engagement shallower.** The fix isn't to avoid the tools — it's to build checkpoints that force depth at the right moments:

1. **After discovery, before writing** — verify every source via `read_file`
2. **After writing, before publishing** — re-read every cited source in context
3. **When token budget is tight** — prioritize footnote reads over additional items

The pre-publish checklist in [study_template.md](study_template.md) already captures rule 1. What's missing is enforcement — a habit of actually running the checklist rather than trusting that the research phase was sufficient.

---

*This is the third reflections document. See also:*
- *[01_reflections.md](01_reflections.md) — Finding vs. reading: how search tools changed study quality*
- *[02_reflections-TODO.md](02_reflections-TODO.md) — Concrete tool improvements (all implemented)*
- *[mcp-improvements.md](mcp-improvements.md) — Gospel MCP observations and improvement proposals*
