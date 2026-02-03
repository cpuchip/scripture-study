# Gospel MCP: Observations and Improvement Proposals

*Analysis from AI assistant's perspective after using both gospel-mcp and direct file access*

---

## Overview

This document captures observations from using the `gospel-mcp` tool (FTS5 full-text search) versus the traditional approach of reading markdown files directly from the workspace. The goal is to identify what's working, what's not, and how to improve the tool for scripture study sessions.

---

## What's Working Well

### gospel-mcp Strengths

| Feature | Benefit | Example |
|---------|---------|---------|
| **Indexed search speed** | Near-instant results across entire corpus | Searching "morning stars sang" returns Job 38:7 and Eyring talk in milliseconds |
| **Cross-source discovery** | Find content across scriptures, talks, manuals simultaneously | One query finds Isaiah 40:26, Abraham 3, AND Elder Tai's April 2025 talk |
| **Phrase and boolean search** | FTS5 supports "exact phrase", AND/OR, prefix matching | `"host of heaven"` finds exact phrase; `stars OR moon` finds either |
| **Source filtering** | Can narrow to scriptures, conference, manual, or all | `source: scriptures` excludes conference talks |
| **Excerpt context** | Shows surrounding text so you can evaluate relevance | Helps decide if a result is worth reading in full |

### Traditional File Access Strengths

| Feature | Benefit | Example |
|---------|---------|---------|
| **Full markdown structure** | Headers, footnotes, cross-references intact | Reading D&C 93 shows all 53 verses with inline links |
| **Natural linking** | File paths are obvious → easy to create markdown links | `../gospel-library/eng/scriptures/ot/job/38.md` |
| **Deep context** | Full chapter reveals themes, flow, surrounding verses | Seeing Job 38:31-33 in context of God's questions to Job |
| **Cross-references preserved** | Footnotes link to related scriptures | D&C 93:29 footnotes → Abraham 3:18, TG Intelligence |
| **Familiar navigation** | Can browse folders, see what exists | Discovering all talks in `/general-conference/2025/04/` |

---

## What's Not Working

### gospel-mcp Pain Points

#### 1. **Missing File Paths for Linking**

**Problem:** The tool returns `file_path` but I consistently forgot to use it for markdown links.

**Why it happens:**
- The excerpt is the "answer" — I focused on the content, not the path
- Paths are returned in Windows format (`gospel-library\eng\...`) not relative markdown format (`../gospel-library/eng/...`)
- No prompt to "use this path for linking"

**Impact:** Created `mazzaroth-01.md` with 40+ scripture references but ZERO markdown links initially.

#### 2. **Excerpts Lose Markdown Structure**

**Problem:** Search results return plain text excerpts, stripping:
- Headers and formatting
- Footnote markers and links
- Cross-reference annotations

**Example:** A search for D&C 93:36 returns:
```
"The glory of God is intelligence, or, in other words, light and truth."
```

But the actual file contains:
```markdown
**36.** The glory<sup>[36a](#fn-36a)</sup> of God is intelligence<sup>[36b](#fn-36b)</sup>, or, in other words, light<sup>[36c](#fn-36c)</sup> and truth.
```

**Impact:** Miss the rich cross-references (TG God, Glory of; TG Intelligence; etc.)

#### 3. **Multi-Word Queries Sometimes Fail**

**Problem:** Complex queries often return 0 results.

**Examples that failed:**
- `"heavens declare glory"` → 0 results
- `speaker:nelson stars` → 0 results (field syntax may not work as expected)

**What worked instead:**
- Single words: `stars`, `heavens`, `Mazzaroth`
- Simple phrases: `"morning stars"`

#### 4. **No "Retrieve Full Document" Option**

**Problem:** After finding a relevant result, I can't easily get the full chapter/talk.

**Current workflow:**
1. Search → get excerpt
2. Note the file path
3. Switch to `read_file` tool
4. Read the full content

**Desired workflow:**
1. Search → get excerpt
2. Request full document in same tool

#### 5. **Conference Talk Filenames Are Opaque**

**Problem:** Results show `51eyring.md` but I can't tell which Eyring talk this is without reading it.

**Would be helpful:** Include talk title and date more prominently in results.

---

### Traditional File Access Pain Points

| Issue | Impact |
|-------|--------|
| **Slow broad searches** | Searching all conference talks for "stars" would require many file reads |
| **Must know where to look** | Can't discover content I don't know exists |
| **No cross-source search** | Separate searches for scriptures vs. conference |
| **grep_search limitations** | Pattern matching without semantic understanding |

---

## Proposed Improvements

### Priority 1: Enable Easy Linking

**Proposal:** Return a `markdown_link` field in results, pre-formatted for study documents.

**Current response:**
```json
{
  "file_path": "gospel-library\\eng\\scriptures\\ot\\job\\38.md",
  "reference": "Job 38:31-33"
}
```

**Proposed response:**
```json
{
  "file_path": "gospel-library\\eng\\scriptures\\ot\\job\\38.md",
  "reference": "Job 38:31-33",
  "markdown_link": "[Job 38:31-33](../gospel-library/eng/scriptures/ot/job/38.md)"
}
```

**Alternatively:** Add a `linkFormat` parameter to the search that auto-generates links based on the calling context.

---

### Priority 2: Add "Get Full Document" Tool

**Proposal:** Create `gospel_get_full` or enhance `gospel_get` to return complete markdown file.

**Use case:** After searching, retrieve the full chapter/talk with all formatting intact.

**Parameters:**
- `path`: File path from search result
- `include_context`: Optional verses/paragraphs before/after (for scripture)

**Example:**
```
gospel_get_full(path="gospel-library/eng/scriptures/ot/job/38.md")
→ Returns complete Job 38 with all verses, footnotes, and cross-references
```

---

### Priority 3: Preserve Markdown in Excerpts

**Proposal:** Option to return excerpts with markdown formatting intact.

**Parameter:** `preserve_markdown: true`

**Before:**
```
"The glory of God is intelligence, or, in other words, light and truth."
```

**After:**
```markdown
**36.** The glory<sup>[36a](#fn-36a)</sup> of God is intelligence<sup>[36b](#fn-36b)</sup>, or, in other words, light<sup>[36c](#fn-36c)</sup> and truth.
```

---

### Priority 4: Improve Conference Talk Metadata

**Proposal:** Include richer metadata for conference results.

**Current:**
```json
{
  "reference": "Henry B. Eyring, October 2019",
  "title": "Holiness and the Plan of Happiness"
}
```

**Proposed:**
```json
{
  "reference": "Henry B. Eyring, October 2019",
  "title": "Holiness and the Plan of Happiness",
  "session": "Sunday Morning",
  "calling": "Second Counselor in the First Presidency",
  "markdown_link": "[President Henry B. Eyring, \"Holiness and the Plan of Happiness\"](../gospel-library/eng/general-conference/2019/10/51eyring.md)"
}
```

---

### Priority 5: Better Query Syntax Documentation

**Proposal:** Improve the tool description with working examples.

**Include:**
- ✅ Working queries: `stars`, `"morning stars"`, `faith OR hope`
- ❌ Queries that don't work as expected
- Field filters: Which fields are searchable? (`speaker:`, `book:`, etc.)
- Wildcard patterns: `intellig*` for prefix matching

---

## Recommended Workflow

Based on current capabilities, here's the optimal workflow:

### For Discovery (Finding What Exists)
1. Use `gospel_search` with simple keywords
2. Review excerpts to identify relevant content
3. Note file paths for linking

### For Deep Study (Understanding Context)
1. Use `read_file` to get full chapter/talk
2. Study with footnotes and cross-references intact
3. Create markdown links naturally from file paths

### For Document Creation
1. **Search** → Discover relevant scriptures/talks
2. **Read** → Get full content with `read_file`
3. **Link** → Create markdown links from file paths
4. **Write** → Compose study notes with proper references

---

## Data Completeness Check

### What's Indexed?

| Source | Status | Notes |
|--------|--------|-------|
| Standard Works (scriptures) | ✅ Complete | OT, NT, BoM, D&C, PoGP |
| Topical Guide | ✅ Present | TG entries searchable |
| Bible Dictionary | ✅ Present | BD entries searchable |
| General Conference | ✅ 1971-2025 | 54 years of talks |
| Come, Follow Me | ✅ Present | Multiple years |
| General Handbook | ✅ Present | Current edition |
| Ensign/Liahona | ⚠️ Partial | Not all years |
| Institute Manuals | ⚠️ Unknown | Need to verify |

### Missing Content?

Content that would be valuable but may not be indexed:
- Joseph Smith Papers
- Church History volumes
- Teachings of the Presidents series (verify)
- BYU Speeches / Devotionals
- Hymns (text)

---

## Summary

| Aspect | gospel-mcp | File Access | Recommendation |
|--------|-----------|-------------|----------------|
| **Discovery** | ⭐⭐⭐⭐⭐ | ⭐⭐ | Use MCP |
| **Deep reading** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Use file access |
| **Creating links** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Use file access |
| **Cross-references** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Use file access |
| **Speed** | ⭐⭐⭐⭐⭐ | ⭐⭐ | Use MCP |

**Bottom Line:** Use gospel-mcp for **discovery**, then fall back to **file access** for deep reading and document creation. The MCP finds things fast; the files give full context and natural linking.

---

## Action Items

- [ ] Implement `markdown_link` field in search results
- [ ] Add `gospel_get_full` tool for complete document retrieval
- [ ] Option to preserve markdown formatting in excerpts
- [ ] Improve conference talk metadata (session, calling, formatted link)
- [ ] Document working query syntax with examples
- [ ] Verify data completeness (Institute manuals, Teachings of Presidents)

---

*Document created: February 3, 2026*
*Based on AI assistant observations during Mazzaroth study sessions*
