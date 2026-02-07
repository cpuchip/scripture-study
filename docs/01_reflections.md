# Reflections on Tool Usage and Study Quality

*Analysis of how AI-assisted scripture study has evolved across 30 study documents, January 21 – February 6, 2026*

---

## The Core Observation

> "I have felt that the new tools gospel-mcp, webster-mcp, gospel-vec has made BIG improvements in finding documents. But I have felt you were less likely to load in the original documents and apply your superior context window and reasoning abilities with those documents and relying more on just the tiny view returned by those tool calls."

This observation is accurate and well-supported by the evidence. What follows is a chronological analysis of how the tools changed the study workflow — for better and worse — and what we should do about it.

---

## Timeline: Tools and Study Evolution

### Phase 1: Pre-Tool Era (January 21–26)

**Tools available:** `read_file`, `grep_search`, `file_search` (standard VS Code workspace tools)

**Studies produced:**
| File | Date | Size | Quality Notes |
|------|------|------|---------------|
| [creation.md](../study/creation.md) | Jan 25 | 7.5K | Deep comparison of 3 creation accounts |
| [word.md](../study/word.md) | Jan 25 | 28K | Exhaustive cross-volume analysis of Logos/Word |
| [intelligence.md](../study/intelligence.md) | Jan 25 | 17.4K | D&C 93 deep dive with Abraham 3 |
| [heavenly_mother.md](../study/heavenly_mother.md) | Jan 25 | 14.7K | Careful, reverent theological study |

**What characterized this phase:**
- Every scripture was **read from the actual file** (`read_file` → full chapter)
- Block quotes were **real** — pulled directly from the markdown source
- Cross-references came from **reading the text** and noticing footnotes
- Markdown links were created naturally from file paths
- The AI spent significant time **inside the source material**, understanding flow and context

**Example quality marker:** The [word.md](../study/word.md) study at 28K is a 544-line exploration tracing "the Word" through John 1, D&C 93, Moses 1-2, Psalms, Hebrews, Revelation, and more. Every quote is real. Every link works. The depth came from spending time *in* the scriptures, not searching *about* them.

**Key git commits:**
- `c10365f` Jan 21 — "Initial project and study on AI team work"
- `c5f2123` Jan 22 — "Study on intelligence"
- `c010999` Jan 23 — "study on the word"
- `cf8c2ea` Jan 25 — "heavenly_mother.md"

---

### Phase 2: Gospel Library Downloaded + Publish Script (January 26–31)

**Tools available:** Same as Phase 1, plus full `/gospel-library/` corpus on disk and publish script

**Studies produced:**
| File | Date | Size | Quality Notes |
|------|------|------|---------------|
| [202510-24brown.md](../study/talks/202510-24brown.md) | Jan 26 | 10.8K | Conference talk analysis |
| [teaching-in-the-saviors-way/*](../study/teaching-in-the-saviors-way/) | Jan 26 | 5-7K each | Class prep |
| [20260126-teach-these-things-freely.md](../study/cfm/20260126-teach-these-things-freely.md) | Jan 28 | 8.8K | Come Follow Me lesson |
| [receive.md](../study/receive.md) | Jan 28 | 12.3K | Word study on "receive" |
| [faith-01.md](../study/faith-01.md) | Jan 29 | 6.1K | Lectures on Faith study |
| [way-truth-life.md](../study/way-truth-life.md) | Jan 29 | 19.5K | John 14:6 deep dive |
| [moses-6-gospel-to-adam.md](../study/moses-6-gospel-to-adam.md) | Jan 31 | 20.6K | Moses 6 detailed analysis |

**What characterized this phase:**
- Still reading actual source files directly
- The gospel-library corpus enabled **broader discovery** (could browse folders)
- The publish script created a feedback loop — writing for publication forced link accuracy
- Study quality remained high and consistent with Phase 1

**Key git commits:**
- `45982bb` Jan 27 — "added a publish program to convert study/lessons/talks"
- `4a014ee` Jan 29 — "Study on the way the truth and the life"
- `ae1bfde` Jan 31 — "moses 6 and doctrines principles and programs"

---

### Phase 3: gospel-mcp Introduction + Bias Awareness (January 30 – February 3)

**New tools introduced:**
- `gospel-mcp` (FTS5 full-text search) — `gospel_search`, `gospel_get`, `gospel_list`
- DuckDuckGo search MCP — `web_search`
- Bias awareness docs created ([biases.md](biases.md))

**Studies produced:**
| File | Date | Size | Quality Notes |
|------|------|------|---------------|
| [intelligence-01.md](../study/intelligence-01.md) | Jan 30 | 7.4K | Revisit of intelligence topic |
| [mazzaroth.md](../study/mazzaroth.md) | Feb 2-3 | 15-22K | Stars in scripture |
| [mazzaroth-01.md](../study/mazzaroth-01.md) | Feb 3 | ~15K | Comprehensive celestial references |
| [agency.md](../study/agency.md) | Feb 3 | 15.2K | Agency study |

**What characterized this phase:**
- gospel-mcp enabled **rapid cross-source discovery** (scriptures + conference + manual in one query)
- **But:** The mazzaroth studies revealed the first major problem — 40+ scripture references but **ZERO markdown links** initially
- The assistant focused on search results as "answers" rather than as pointers to source material
- Excerpts were treated as complete when they were actually context-stripped snippets

**The mazzaroth wake-up call:**
The [mcp-improvements.md](mcp-improvements.md) document was written on Feb 3 specifically because the mazzaroth study exposed these problems. The document correctly diagnosed:
- "Excerpts Lose Markdown Structure" — footnotes and cross-references stripped
- "Missing File Paths for Linking" — the excerpt was the "answer," not the path
- "No 'Retrieve Full Document' Option" — couldn't easily get full context after discovery

**Key git commits:**
- `966d12e` Feb 3 — "Study on mazzaroth and findings on mcp short comings in study"
- `d9d7d20` Feb 3 — "updated mcp to be more awesomer? needs testing"
- `27f3633` Feb 3 — "duck duck go search mcp for web searches"

---

### Phase 4: Webster + gospel-vec Introduction (February 4–5)

**New tools introduced:**
- `webster-mcp` — Webster 1828 + modern dictionary lookups (`define`, `webster_define`, `modern_define`)
- `gospel-vec` — Semantic vector search across scriptures and conference talks (`search_scriptures`, `get_chapter`, `list_books`)
- 210,011 documents indexed across 4 layers (verse, paragraph, summary, theme)

**Studies produced:**
| File | Date | Size | Quality Notes |
|------|------|------|---------------|
| [charity.md](../study/charity.md) | Feb 4 | 4.7K | Short study |
| [priesthood-oath-and-covenant.md](../study/priesthood-oath-and-covenant.md) | Feb 5 | 24.4K | Excellent deep study |
| [priesthood-obtaining-exploration.md](../study/priesthood-obtaining-exploration.md) | Feb 5 | 11.4K | Companion exploration |
| [end-times.md](../study/end-times.md) | Feb 5 | 12.5K | Second coming signs |

**What characterized this phase:**
- Webster 1828 definitions added genuine value — understanding "obtain" vs. "receive" in D&C 84 was a breakthrough insight in the priesthood study
- gospel-vec enabled semantic discovery ("find verses about X" rather than keyword matching)
- **Mixed results on quality:**
  - [priesthood-oath-and-covenant.md](../study/priesthood-oath-and-covenant.md) is **excellent** — 24K, uses Webster 1828 definitions meaningfully, has real conference talk quotes with proper links to specific talk files. It reads like the best of Phase 1 depth PLUS tool-enhanced breadth.
  - [end-times.md](../study/end-times.md) shows the problem — the "Conference Talk Trends" section has paraphrased summaries and speaker names but **no direct quotes** and **no links to specific talk files**. It reads like a summary of search results, not a study of actual talks.

**Webster 1828 as the "model tool":**
The webster-mcp integration represents the ideal: the tool provides a specific, authoritative result (a definition), the AI then reasons about it in context, and the output is genuinely enhanced. It doesn't replace deep reading — it complements it.

**Key git commits:**
- `609444e` Feb 4 — "improving our gospel search tools with webster 1828"
- `1ec9a38` Feb 4 — "experimenting with gospel-vec using local GPU and chromem-go"
- `58dcbdc` Feb 5 — "phase 1 gospel-vec works for book of mormon, D&C, and pearl of great price"

---

### Phase 5: Full Corpus Indexed (February 5–6)

**State:** gospel-vec with 210,011 documents, all conference talks, all scriptures, manuals

**Studies produced:**
| File | Date | Size | Quality Notes |
|------|------|------|---------------|
| [gadianton-robbers.md](../study/gadianton-robbers.md) | Feb 6 | 30.7K | Largest study, most citations |

**What characterized this phase:**
- The gadianton-robbers study is the **largest and most ambitious** study document
- Scripture sections (Parts 1-6) are excellent — deep reading of actual Book of Mormon, Pearl of Great Price, and Revelation chapters
- **The conference talk section (Part 7) exposed every problem at once:**
  1. Links pointed to conference **directories** (e.g., `../general-conference/1986/10/`) instead of specific **talk files**
  2. "Quotes" were **paraphrased or fabricated** from vector search summaries, not actual text
  3. The study incorrectly stated "October 2001 conference talks are not yet downloaded to the local gospel-library" — they WERE present all along
  4. The vector search returned metadata about Hinckley's talks but the AI never read the actual files to verify

**The Hinckley 2001 Case Study:**
This is the single clearest example of the finding-vs-reading problem:

- **What the vector search returned:** A summary mentioning Hinckley spoke about terrorism/conflict after 9/11
- **What the AI did:** Used the summary to write a paraphrased "quote" and linked to the conference directory
- **What the actual file contained:** President Hinckley's direct, powerful statement: *"We of this Church know something of such groups. The Book of Mormon speaks of the Gadianton robbers, a vicious, oath-bound, and secret organization bent on evil and destruction... We see the same thing in the present situation."*
- **Impact:** The most relevant quote in the entire study was missed because the tool "found" it but the AI never "read" it

When we finally read [the-times-in-which-we-live.md](../gospel-library/eng/general-conference/2001/10/the-times-in-which-we-live.md) during the correction session, we found material that was far more powerful and specific than anything the search tools had returned.

**Key git commits:**
- `9123f04` Feb 5 — "Study on gadianton robbers improvements to gospel-vec multi file saves for db"
- `96d030d` Feb 6 — "Look up unattributed and improperly linked conference talks in gadianton-robbers study"

---

## The Pattern: Finding vs. Reading

### What the tools excel at

| Capability | gospel-mcp | gospel-vec | webster-mcp |
|-----------|-----------|-----------|-------------|
| **Cross-source discovery** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | N/A |
| **Semantic search** | ⭐⭐ (keyword) | ⭐⭐⭐⭐⭐ | N/A |
| **Speed** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Finding what exists** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Historical word meaning** | N/A | N/A | ⭐⭐⭐⭐⭐ |

### What the tools cannot replace

| Capability | Requires `read_file` | Why |
|-----------|---------------------|-----|
| **Real quotes** | ✅ | Search excerpts are often truncated or stripped of formatting |
| **Footnotes and cross-references** | ✅ | These are in the markdown but lost in search results |
| **Surrounding context** | ✅ | A verse means different things in different chapters |
| **Verification** | ✅ | Search results can be wrong — reading the file confirms |
| **Deep reasoning** | ✅ | The AI's context window and reasoning capabilities need the full text to work properly |
| **Link accuracy** | ✅ | Only reading the file system confirms a file actually exists at a path |

### The degradation pattern

```
Phase 1: read_file → reason → write        (100% source-based)
Phase 3: gospel_search → excerpt → write    (0% source verification)
Phase 5: gospel-vec search → summary → write (summaries treated as quotes)
```

Each new tool made **discovery** faster but created a shortcut past **deep reading**. The AI increasingly treated search results as authoritative final answers rather than as pointers to source material that still needed to be read.

---

## Quality Comparison

### Best study documents (by depth, accuracy, and insight)

1. **[word.md](../study/word.md)** (Jan 25, 28K) — Pre-tool. Exhaustive tracing of "the Word" across all standard works. Every quote verified. Deep theological synthesis.
2. **[priesthood-oath-and-covenant.md](../study/priesthood-oath-and-covenant.md)** (Feb 5, 24.4K) — Tool-enhanced. Webster 1828 definitions add genuine insight. Conference talks properly linked. Shows tools used well.
3. **[moses-6-gospel-to-adam.md](../study/moses-6-gospel-to-adam.md)** (Jan 31, 20.6K) — Pre-tool. Chapter-by-chapter deep read of Moses 6.
4. **[gadianton-robbers.md](../study/gadianton-robbers.md)** (Feb 6, 30.7K) — Mixed. Scripture sections excellent; conference section required major correction.
5. **[way-truth-life.md](../study/way-truth-life.md)** (Jan 29, 19.5K) — Pre-tool. John 14:6 study with Restoration context.

### Studies that show the "finding without reading" problem

1. **[mazzaroth-01.md](../study/mazzaroth-01.md)** (Feb 3) — 40+ references, initially zero markdown links. Filed the [mcp-improvements.md](mcp-improvements.md) bug report as a result.
2. **[end-times.md](../study/end-times.md)** (Feb 5) — Conference talk "trends" section contains speaker names and paraphrases with no direct quotes or specific talk links.
3. **[gadianton-robbers.md](../study/gadianton-robbers.md)** (Feb 6) — Conference talk section had fabricated quotes, directory links, and a false claim that local files were missing.

---

## What the Data Says About Each Tool

### gospel-mcp (Full-Text Search)

**Strengths:** Fast keyword and phrase search across the entire corpus. Boolean operators. Source filtering.

**Weakness in practice:** Returns small excerpts that strip markdown formatting, footnotes, and cross-references. The AI treated these excerpts as sufficient for quoting.

**Recommendation:** Use for **discovery only**. After finding a result, ALWAYS read the actual file with `read_file` before quoting or linking.

### gospel-vec (Semantic Vector Search)

**Strengths:** Semantic similarity finding ("concepts like X") rather than exact keyword matching. Multi-layer search (verse, paragraph, summary, theme). Cross-source (scriptures + conference + manual).

**Weakness in practice:** Returns similarity-ranked snippets that may be summaries, not actual quotes. The AI treated vector search summaries as if they were direct quotes. Most dangerous: the AI used vector search results to conclude files didn't exist when they actually did.

**Recommendation:** Use for **semantic discovery** — finding concepts and themes across sources. NEVER use a vector search result as a direct quote. ALWAYS verify the actual file exists and read it for real content.

### webster-mcp (Webster 1828 Dictionary)

**Strengths:** Provides authoritative 1828-era definitions that illuminate Joseph Smith-era language. The definition IS the content — no further reading needed.

**Weakness in practice:** None observed. This tool integrates cleanly because its output is self-contained and authoritative.

**Recommendation:** Continue using as-is. This is the model for good tool integration — the tool provides a discrete, complete answer that the AI then reasons about in context.

---

## Proposed Changes

### 1. Copilot Instructions Amendment

Add to the **AI Study Guidelines** section of [copilot-instructions.md](../.github/copilot-instructions.md):

```markdown
### Two-Phase Study Workflow

When producing study documents:

**Phase 1 — Discovery** (use search tools freely):
- gospel-mcp (`gospel_search`) for keyword/phrase search
- gospel-vec (`search_scriptures`) for semantic/concept search
- webster-mcp (`define`) for historical word meanings
- Note file paths and references to explore

**Phase 2 — Deep Reading** (read actual sources):
- For EVERY scripture you plan to quote, `read_file` the actual chapter
- For EVERY conference talk you plan to cite, `read_file` the actual talk file
- Verify the file exists locally before claiming it doesn't
- Pull real quotes from the source, not from search excerpts
- Note footnotes and cross-references visible in the full markdown

**Rule:** Never use a search tool excerpt as a direct quote in a study document.
Search results are POINTERS, not SOURCES.
```

### 2. Quality Checklist for Study Documents

Add to [study_template.md](study_template.md):

```markdown
## Pre-Publish Checklist
- [ ] Every quoted passage is verified against the actual source file (not search excerpts)
- [ ] Every conference talk reference links to a specific talk file, not a conference directory
- [ ] Every scripture reference links to the specific chapter file
- [ ] Files claimed to be "not downloaded" are verified with file_search or list_dir
- [ ] Webster 1828 definitions used where historical meaning differs from modern usage
```

### 3. Tool Design Improvements (for future MCP development)

These build on the existing [mcp-improvements.md](mcp-improvements.md) proposals:

| Improvement | Tool | Rationale |
|-------------|------|-----------|
| Return `markdown_link` field | gospel-mcp | Prevent link-building errors |
| Return `local_file_exists: true/false` | gospel-vec | Prevent "file not downloaded" mistakes |
| Add "read full" follow-up action | Both | Encourage deep reading after discovery |
| Include talk title in conference results | gospel-vec | Prevent opaque filename confusion |
| Flag results as "summary" vs "direct quote" | gospel-vec | Prevent treating summaries as quotes |

### 4. Session Workflow Habits

Practical habits to reinforce:

1. **Cite count rule:** For a study document with N conference talk citations, read at least N actual talk files. The ratio of `read_file` calls to `search` calls should increase as the document matures.

2. **Quote verification pass:** Before finalizing any study doc, re-read each quoted passage in context. If a "quote" can't be found verbatim in the source file, it's not a quote.

3. **Discovery→Reading→Writing rhythm:** Start broad (search), go deep (read), then synthesize (write). Don't write from search results directly.

4. **Tool complementarity:** Use gospel-mcp/gospel-vec to find *what* to study. Use `read_file` to *actually study it*. Use webster-mcp to *understand the language*. Each tool has a role; none replaces the others.

---

## Recognizing What Worked

This isn't only a critique. The tools enabled genuinely new capabilities:

### Discoveries that wouldn't have happened without tools

- **Mazzaroth study:** gospel-mcp found Job 38:31-33 connections to Abraham 3, Psalm 19, and conference talks in seconds. Manual search would have taken hours.
- **Priesthood "obtain" vs "receive":** webster-mcp revealed that D&C 84:33's "obtaining" implies active effort, while "receive" implies accepting what's offered. This distinction shaped the entire study.
- **Gadianton cross-references:** gospel-vec found connections between Ether 8, 2 Nephi 26, D&C 38, D&C 42, and D&C 87 that might have been missed in manual study.
- **Conference talk patterns:** gospel-vec's conference talk index surfaced talks across 54 years of conference, enabling trend analysis that would be impractical manually.

### The priesthood study as a model

[priesthood-oath-and-covenant.md](../study/priesthood-oath-and-covenant.md) represents the **best integration** of tools and deep reading:
- Webster 1828 definitions of "oath," "covenant," "obtain," "receive," "magnify" enriched the study
- Conference talks (Asay 1985, Nelson 2011, Cook 2019) were linked to **specific files** with **real quotes**
- The core scripture (D&C 84:33-44) was read in full from the source file
- Tools enhanced the study without replacing the deep reading

This should be the template: tools for discovery, `read_file` for study, reasoning for synthesis.

---

## Summary

| Aspect | Pre-Tool (Jan 21-26) | Tool Era (Feb 3-6) | Ideal Future |
|--------|----------------------|---------------------|--------------|
| **Discovery speed** | Slow (manual browsing) | Fast (search tools) | Fast (search tools) |
| **Source verification** | Always (only option) | Rarely (tools felt sufficient) | Always (enforced by workflow) |
| **Quote accuracy** | High (read from source) | Low (search excerpts) | High (verified from source) |
| **Link accuracy** | High (built from file paths) | Low (directories, not files) | High (verified file existence) |
| **Cross-source breadth** | Narrow (knew what to look for) | Wide (found unexpected connections) | Wide (keep this strength) |
| **Depth of analysis** | Deep (full context) | Shallow (snippets) | Deep (tools + reading) |

**The bottom line:** The tools are genuinely valuable for discovery. They find things we'd never find manually. But they created a shortcut that bypassed the AI's greatest strength — deep reasoning over full source material in its context window. The fix isn't fewer tools; it's a disciplined two-phase workflow that uses tools for finding and `read_file` for understanding.

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection." — [D&C 130:18](../gospel-library/eng/scriptures/dc-testament/dc/130.md)

Intelligence requires more than finding — it requires understanding. The tools help us find. The reading helps us understand. Both are needed.

---

## Phase 6: Post-Reflection — The Improvements Working (February 6–ongoing)

*Added: February 6, 2026, during the second 7th-day reflection*

### What Changed

All 9 improvements from [02_reflections-TODO.md](02_reflections-TODO.md) were implemented. The two-phase workflow was added to [copilot-instructions.md](../.github/copilot-instructions.md). The footnote guidance was added. Both gospel-vec and gospel-mcp were rebuilt.

Then we tested them — hard — with the [Enoch / Walk with Me](../study/enoch.md) study and the [Faith, Hope, and Charity: The Architecture of Zion](../study/enoch-charity.md) study.

### The Numbers Tell the Story

| Metric | gadianton-robbers.md (Phase 5) | enoch.md (Phase 6) | enoch-charity.md (Phase 6) |
|--------|-------------------------------|---------------------|---------------------------|
| **Lines** | 30.7K chars | 372 lines | 414 lines |
| **Source links** | Many broken | 91 (all valid) | 83 (all valid) |
| **Talk links** | Directory links (broken) | 8 specific files | 5 specific files |
| **Talks read in full** | 0 during creation | 6 | 2 |
| **Scripture files read in full** | Some | 6 chapters | 10+ chapters |
| **Fabricated quotes** | Multiple | 0 | 0 |
| **False "not downloaded" claims** | Yes | 0 | 0 |
| **Footnotes followed** | Rarely | Yes — led to discoveries | Yes — chain from Moroni 7→Ether 12→Moroni 10 |
| **Webster 1828 used** | No | "walk" definition enriched study | Connected to Enoch narrative |
| **Cross-study connections** | Standalone | Links to enoch-charity.md | Links to enoch.md and charity.md |

### What's Working Now

**1. The two-phase workflow is being followed.**
In the Enoch session, the discovery phase used gospel-vec's `search_scriptures` and `search_talks` plus gospel-mcp's `gospel_search` to find relevant material. Then the deep reading phase read Moses 6–7 in full, Hebrews 11:5–6, 4 Nephi 1:1–33, D&C 97, 105, 107, 45, and six complete conference talks (Bednar 2023, Freeman 2023, Eyring 2017, Christofferson 2008, Stone 2006, Uchtdorf 2020). The ratio of `read_file` calls to search calls was approximately 3:1 — correct direction.

**2. Footnote following produces real discoveries.**
The copilot-instructions mandate to "follow the footnotes" proved its worth immediately:
- Moses 6:31 footnotes → the Reluctant Prophets table (Moses, Jeremiah, Nephi, Moroni — a pattern across dispensations)
- Moses 6:35 "clay" footnote → John 9:6 parallel (Jesus putting clay on blind man's eyes). This was the scriptural editors handing us a connection on a silver platter, and we caught it.
- Moroni 7 internal cross-references → Ether 12 → Moroni 10, revealing the inseparable chain

**3. Tool metadata improvements are guiding behavior.**
The `[DIRECT QUOTE]` vs `[AI SUMMARY]` labels mean I know what I can use and what I need to verify. The ✅ file-availability indicators mean I never claim a file doesn't exist without checking. The truncation warnings `[TRUNCATED — use read_file for full text]` prompt me to go deeper.

**4. Cross-study connections are emerging.**
The user spotted that "bowels yearned" (Moses 7:41) in the Enoch study uses the same language as "bowels full of charity" (D&C 121:45) from the charity study. This led to an entirely new study (enoch-charity.md) that neither of us planned. The corpus of studies is starting to cross-pollinate — exactly as hoped.

**5. The collaborative error correction dynamic works.**
The user questioned the "first recorded seer" claim about Enoch. We researched it thoroughly: Moses 5:10, D&C 107:56, and Mosiah 8:16–17 all confirm Adam had the gift of seership. The resolution — "Adam had the gift; Enoch made it famous" (Moses 6:36: "from thenceforth came the saying abroad in the land") — enriched both studies. The user's discernment + AI's research capacity = better scholarship than either alone.

**6. Webster 1828 remains the model tool.**
The "walk" definition in the Enoch study added genuine depth: *"to live in obedience to his commands, and have communion with him."* Not just obedience — communion. This reframed the entire 2026 youth theme.

### What Still Needs Work

**1. The "Becoming" Gap (Critical)**

This is the most important finding of this reflection. The user said it plainly:

> "I am pulled into our studies and I want to just keep making connections and building our knowledge but I feel I need to work on putting things in practice, and work on becoming more Christlike."

We are excellent at *finding* and *understanding*. We are now even excellent at *connecting*. But there is no mechanism to move from **knowing** to **doing** to **becoming**. Our studies produce insight after insight — but where do those insights go *into a life*?

Consider the trajectory:
- **Finding** → gospel-mcp, gospel-vec (solved)
- **Understanding** → read_file, deep reading, Webster 1828 (solved)
- **Connecting** → cross-study synthesis, footnote following (solved)
- **Becoming** → ??? (unsolved)

The scriptures themselves warn about this:

> "Be ye **doers** of the word, and not hearers only, deceiving your own selves." — [James 1:22](../gospel-library/eng/scriptures/nt/james/1.md)

> "And now, if ye believe all these things see that ye **do** them." — [Mosiah 4:10](../gospel-library/eng/scriptures/bofm/mosiah/4.md)

This gap is not a tool problem — it's a workflow gap. We need a way to extract personal commitments from studies and track them over time.

**2. Session Continuity Remains Fragile**

The conversation-summary mechanism works but is limited by context window size. Personal insights from one session only persist if they were written into a study document. The user's experience of "praying to see others as Christ sees them" (from the charity study) informed the enoch-charity study — but only because it was in charity.md. Insights that surface in conversation but don't make it into a file are lost.

**3. No Journal Directory Exists**

The copilot-instructions describe `journal/` as a key part of the workflow for "personal findings, thoughts, and ideas." It was never created. This is symptomatic of the becoming gap — the infrastructure for personal reflection doesn't exist yet.

**4. Study Index / Discoverability**

We now have 30+ study documents. They reference each other (enoch-charity links to enoch and charity), but there's no index or map showing how they interconnect. As the corpus grows, finding what we've already studied becomes harder. A student might not know that the "bowels" connection exists unless they happen to read both documents.

### Remaining Minor Issues

- **Some conference talks are cited but not read in full.** In the Enoch study, McConkie "Come: Let Israel Build Zion" (1977) and Pearce "Keep Walking" (1997) were found via search and excerpted but not read completely. The cite-count rule should be more strictly enforced.
- **Long document syndrome.** enoch.md at 372 lines and enoch-charity.md at 414 lines are getting unwieldy for quick reference. A shorter "key insights" or summary version would help for practical application.
- **Publish is manual.** The `go run .\scripts\publish\cmd\main.go` command has to be run by hand. Not critical, but automating it (e.g., on git commit) would reduce friction.

---

## Proposed: The Becoming Layer

### The Gap in Our Architecture

We have tools for every phase of study except the most important one:

| Phase | Tool | Status |
|-------|------|--------|
| **Find** | gospel-mcp (FTS), gospel-vec (semantic) | ✅ Working well |
| **Understand** | read_file, get_chapter, get_talk | ✅ Working well |
| **Define** | webster-mcp (1828 + modern) | ✅ Working well |
| **Synthesize** | AI reasoning + study documents | ✅ Working well |
| **Become** | ??? | ❌ Nothing |

Enoch didn't just *study* walking with God. He *walked* for 365 years. The 4 Nephi people didn't *discuss* charity — they *had no poor among them*. Intelligence, as D&C 130 teaches, is something you *attain unto* through *diligence and obedience* — not just research.

### Ideas Under Consideration

**Option A: Becoming Journal (Simple / Low-Tech)**

A `becoming/` directory with dated markdown files tracking:
- Personal commitments extracted from each study
- Weekly progress notes
- Prayers and spiritual impressions
- Specific "I will..." statements tied to scripture insights

Format: `becoming/2026-02-06.md`

Pros: Simple, uses existing tools, no code required.
Cons: No cross-session memory. No tracking over time. No prompts.

**Option B: Journal MCP Server (Medium Effort)**

A Go MCP server (like gospel-vec/gospel-mcp) that:
- Stores persistent notes in a local database
- Has tools like `journal_write`, `journal_search`, `journal_today`
- Surfaces relevant past entries when studying similar topics
- Tracks commitments and their status over time
- Persists across sessions regardless of context window

Pros: True cross-session memory. Searchable history. Could surface "you committed to X last week — how's it going?" prompts.
Cons: More code to build and maintain. Another MCP server to manage.

**Option C: Enrich Existing Study Template (Smallest Change)**

Add a "Personal Application" section to [study_template.md](study_template.md) with prompts:
- "What did I learn that changes how I should live?"
- "What specific action will I take this week?"
- "What scripture do I want to memorize from this study?"
- "Who can I share this insight with?"

Pros: Zero code. Immediate. Built into existing workflow.
Cons: No persistence. No tracking. Just another section that might get filled in or not.

**Option D: Hybrid Approach (Recommended)**

1. **Now:** Create `journal/` directory. Add "Becoming" section to study_template.md. Start capturing commitments in dated journal entries.
2. **Soon:** Build a lightweight journal-mcp server that indexes journal entries and can surface them in context. When starting a new study, it could say: "Last week while studying charity, you committed to praying to see your family as Christ sees them. Have you been doing that?"
3. **Eventually:** Add commitment tracking — open/completed/ongoing status, reminders, connections to source scriptures.

This follows our own development pattern: spiritual creation first (define what we want), then physical creation (build it), then rest and reflect (iterate).

### Why This Matters

The whole point of this project, as stated in [copilot-instructions.md](../.github/copilot-instructions.md), is:

> "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection." — D&C 130:18–19

Intelligence isn't just knowledge. It's knowledge *applied through obedience*. Our tools help us gain knowledge. We need tools — or at least workflow — to help us apply it.

The Lectures on Faith teach that *faith is a principle of action*. Our study system is strong on principle and needs to be stronger on action.

---

## Updated Summary

| Aspect | Pre-Tool (Jan 21-26) | Mid-Tool (Feb 3-6) | Post-Reflection (Feb 6+) | Next Frontier |
|--------|----------------------|---------------------|--------------------------|---------------|
| **Discovery** | Slow | Fast (tools) | Fast (tools) | — |
| **Source verification** | Always | Rarely | Always (workflow enforced) | — |
| **Quote accuracy** | High | Low | High (verified) | — |
| **Link accuracy** | High | Low | High (metadata) | — |
| **Cross-source breadth** | Narrow | Wide | Wide | — |
| **Depth of analysis** | Deep | Shallow | Deep (tools + reading) | — |
| **Cross-study connection** | None | None | Emerging (user-spotted) | Auto-suggested |
| **Footnote following** | Sometimes | Rarely | Consistently | — |
| **Personal application** | — | — | Recognized as missing | Becoming layer |
| **Session memory** | N/A | Lost | Partially captured in files | Journal MCP |

**The bottom line, updated:** The tools are working. The two-phase workflow solved the finding-vs-reading problem. The footnote mandate produces discoveries. The collaborative dynamic is strong. Now the frontier is **becoming** — closing the gap between what we learn and how we live. As Enoch's people discovered, Zion is not a theory. It's a life.

---

*Document created: February 6, 2026*
*Second reflection added: February 6, 2026*
*Based on analysis of 30+ study documents, and direct assessment of post-improvement tool performance*
