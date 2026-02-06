# Reflections TODO: Tool Design Improvements

*Created: February 6, 2026*
*Based on findings in [01_reflections.md](01_reflections.md)*

These are concrete code changes to gospel-vec and gospel-mcp that address the "finding vs. reading" problem identified in the reflections analysis. Each improvement helps the AI (and user) move from search results to actual source material more naturally.

---

## Overview of Changes

| # | Improvement | Tool | Effort | Re-index? | Status |
|---|------------|------|--------|-----------|--------|
| 1 | Add `markdown_link` and `file_path` to search results | gospel-vec | Small | No | ✅ Done |
| 2 | Add `result_type` field (quote vs summary) | gospel-vec | Small | No | ✅ Done |
| 3 | Add `local_file_exists` field to search results | gospel-vec | Small | No | ✅ Done |
| 4 | Enrich conference talk metadata in results | gospel-vec | Small | No | ✅ Done |
| 5 | Add `get_talk` tool to gospel-vec | gospel-vec | Medium | No | ✅ Done |
| 6 | Add `search_talks` tool with speaker/year filters | gospel-vec | Medium | No | ✅ Done |
| 7 | Improve `search_scriptures` tool description | gospel-vec | Tiny | No | ✅ Done |
| 8 | Truncation warning in search results | gospel-vec | Tiny | No | ✅ Done |
| 9 | Align gospel-mcp `MarkdownLink` output to be prominent | gospel-mcp | Small | No | ✅ Done |

**Good news:** None of these require re-indexing. They're all output formatting and new tool additions that work with existing indexed data.

---

## Detailed Plans

### 1. Add `markdown_link` and `file_path` to gospel-vec search results

**File:** `scripts/gospel-vec/mcp.go` — `formatMCPSearchResults()`

**Current output:**
```
**1 Nephi 3:7** (87% match)
> I will go and do the things which the Lord hath commanded...
```

**Proposed output:**
```
**1 Nephi 3:7** (87% match)  
📎 [1 Nephi 3](../gospel-library/eng/scriptures/bofm/1-ne/3.md)
> I will go and do the things which the Lord hath commanded...
```

**Implementation:**
- The `DocMetadata.FilePath` field already exists in `types.go` and is populated during indexing
- In `formatMCPSearchResults()`, construct a relative markdown link from `FilePath`
- For conference talks, include the talk title in the link text
- Format: `[Book Chapter](relative/path)` for scriptures, `[Speaker, "Title"](relative/path)` for talks

**Changes needed:**
1. `mcp.go` line ~453: Update `formatMCPSearchResults()` to include file path and markdown link
2. Add a helper `func buildMarkdownLink(meta DocMetadata) string` that generates the proper relative link
3. Ensure conference talk file paths resolve correctly (they use slug-based or numbered naming)

---

### 2. Add `result_type` field to distinguish quotes from summaries

**File:** `scripts/gospel-vec/mcp.go` — `formatMCPSearchResults()`

**Problem:** The AI treats all search results equally. A verse-layer result is an actual scripture quote. A summary-layer result is an LLM-generated paraphrase. These should be labeled differently.

**Proposed output:**
```
## verse Results

**1 Nephi 3:7** (87% match) [DIRECT QUOTE]
📎 [1 Nephi 3](../gospel-library/eng/scriptures/bofm/1-ne/3.md)
> I will go and do the things which the Lord hath commanded...

## summary Results

**1 Nephi 3** (72% match) [AI SUMMARY — verify against source]  
📎 [1 Nephi 3](../gospel-library/eng/scriptures/bofm/1-ne/3.md)
> Nephi's account of faithfully following the Lord's command to return for the brass plates...
```

**Implementation:**
- In `formatMCPSearchResults()`, check `r.Metadata.Layer`:
  - `verse` → label as `[DIRECT QUOTE]`
  - `paragraph` → label as `[DIRECT QUOTE]` (these are actual text)
  - `summary` → label as `[AI SUMMARY — verify against source]`
  - `theme` → label as `[AI THEME — verify against source]`
- Also include the `Generated` and `Model` metadata fields for summary/theme results

**Changes needed:**
1. `mcp.go`: Add result type labeling logic in `formatMCPSearchResults()`

---

### 3. Add `local_file_exists` check to search results

**File:** `scripts/gospel-vec/mcp.go` — `formatMCPSearchResults()`

**Problem:** The Hinckley 2001 case — the AI assumed files didn't exist because it didn't check. If search results told the AI "this file is on disk," it would be prompted to read it.

**Proposed output:**
```
**President Hinckley, "The Times in Which We Live"** (81% match) [AI SUMMARY — verify against source]
📎 [The Times in Which We Live](../gospel-library/eng/general-conference/2001/10/the-times-in-which-we-live.md) ✅ local file available
> Hinckley connects the September 11 attacks to the Gadianton pattern...
```

vs. if file is missing:
```
📎 gospel-library/eng/general-conference/1969/04/some-old-talk.md ❌ not cached locally
```

**Implementation:**
1. In `formatMCPSearchResults()`, for each result, resolve `r.Metadata.FilePath` relative to the workspace root
2. Call `os.Stat()` to check if the file exists
3. Append ✅ or ❌ indicator to the file link line
4. The workspace root can be derived from `Config.ScripturesPath` (go up two levels from `gospel-library/eng/scriptures/`)

**Changes needed:**
1. `mcp.go`: Add file existence check in result formatting
2. `config.go`: Possibly add a `WorkspaceRoot` config field, or derive it

---

### 4. Enrich conference talk metadata in results

**File:** `scripts/gospel-vec/mcp.go` — `formatMCPSearchResults()`

**Problem:** Conference results show opaque references like "October 2001" but not the talk title or speaker's calling, making it hard to evaluate relevance.

**Current:**
```
**Henry B. Eyring, October 2019** (75% match)
> ...
```

**Proposed:**
```
**President Henry B. Eyring** — "Holiness and the Plan of Happiness" (75% match)
*Second Counselor in the First Presidency, October 2019 General Conference, Sunday Morning Session*
📎 [Holiness and the Plan of Happiness](../gospel-library/eng/general-conference/2019/10/51eyring.md) ✅ local
> ...
```

**Implementation:**
- The `DocMetadata` already has `Speaker`, `Position`, `TalkTitle`, `Year`, `Month`, `Session` fields (see `types.go`)
- These are populated during conference talk indexing in `talk_parser.go`
- Just need to use them in `formatMCPSearchResults()` when `r.Metadata.Source == SourceConference`

**Changes needed:**
1. `mcp.go`: Add conference-specific formatting branch in `formatMCPSearchResults()`

---

### 5. Add `get_talk` tool to gospel-vec

**New MCP tool** for retrieving full conference talk text.

**Problem:** After finding a talk via search, there's no way to get the full text through gospel-vec. The user has to know the file path and use `read_file`. A `get_talk` tool would make the "discovery → deep read" flow seamless within gospel-vec.

**Tool definition:**
```json
{
  "name": "get_talk",
  "description": "Get the full text of a conference talk. Use after search_scriptures finds a relevant talk.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "speaker": { "type": "string", "description": "Speaker last name (e.g., 'nelson', 'hinckley', 'oaks')" },
      "year": { "type": "integer", "description": "Conference year (e.g., 2001)" },
      "month": { "type": "string", "description": "Conference month: '04' for April, '10' for October" },
      "file_path": { "type": "string", "description": "Direct file path if known from search results" }
    }
  }
}
```

**Implementation:**
1. `mcp.go`: Register new tool in `handleToolsList()` and `handleToolsCall()`
2. New function `toolGetTalk()`:
   - If `file_path` provided, read directly from disk
   - If speaker/year/month provided, scan the conference directory for matching files
   - Return the full markdown content of the talk
3. Use existing `talk_parser.go` → `FindTalkFiles()` and `ParseTalkFile()` for file discovery

**Changes needed:**
1. `mcp.go`: Add tool definition and routing
2. `mcp.go`: Implement `toolGetTalk()` function (~60 lines)

---

### 6. Add `search_talks` tool with speaker/year filters

**New MCP tool** for filtered conference talk search.

**Problem:** `search_scriptures` searches everything at once. For conference-specific searches (e.g., "find all talks by Hinckley about Gadianton"), a filtered tool would be more precise.

**Tool definition:**
```json
{
  "name": "search_talks",
  "description": "Search conference talks with optional speaker and year filters. Returns semantic matches from indexed talks.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string", "description": "Semantic search query" },
      "speaker": { "type": "string", "description": "Filter by speaker last name" },
      "year_from": { "type": "integer", "description": "Start year (inclusive)" },
      "year_to": { "type": "integer", "description": "End year (inclusive)" },
      "limit": { "type": "integer", "description": "Max results (default: 10)" }
    },
    "required": ["query"]
  }
}
```

**Implementation:**
1. `search.go`: Add a `SearchConference()` method that filters by source and then post-filters by speaker/year from metadata
2. `mcp.go`: Register tool and implement `toolSearchTalks()`
3. Results should use the enriched conference formatting from improvement #4

**Changes needed:**
1. `search.go`: Add `SearchConference()` method with metadata filtering (~30 lines)
2. `mcp.go`: Add tool definition, routing, and `toolSearchTalks()` implementation (~80 lines)

**Note:** chromem-go's `Where` filter support may help here. Check if the version supports metadata filtering in queries — if so, this can be done at the DB level rather than post-filtering.

---

### 7. Improve `search_scriptures` tool description

**File:** `scripts/gospel-vec/mcp.go` — tool definition at line ~114

**Current description:**
```
"Search the scriptures using semantic similarity. Finds verses, paragraphs, chapter summaries, and themes related to the query."
```

**Proposed description:**
```
"Search the scriptures using semantic similarity. Finds verses, paragraphs, chapter summaries, and themes related to the query. Searches across scriptures, conference talks, manuals, and books.\n\nIMPORTANT: Results labeled [AI SUMMARY] are NOT direct quotes — always verify against the source file before quoting. Results include file paths and markdown links for easy follow-up with read_file.\n\nTip: After finding relevant content, use get_chapter or get_talk to read the full source text."
```

**Changes needed:**
1. `mcp.go` line ~115: Update description string

---

### 8. Add truncation warning to search results

**File:** `scripts/gospel-vec/mcp.go` — `formatMCPSearchResults()` and `truncate()`

**Problem:** The `truncate()` function cuts content at 300 chars with "..." — but doesn't tell the AI that more content exists.

**Current:**
```
> I will go and do the things which the Lord hath commanded, for I know that the...
```

**Proposed:**
```
> I will go and do the things which the Lord hath commanded, for I know that the... [TRUNCATED — use read_file or get_chapter for full text]
```

**Changes needed:**
1. `mcp.go`: Update `truncate()` to append a contextual hint when truncation occurs

---

### 9. Make gospel-mcp `MarkdownLink` more prominent

**File:** `scripts/gospel-mcp/internal/tools/tools.go` (and the search result formatting)

**Observation:** gospel-mcp already has `MarkdownLink` in its `SearchResult` struct (line 38). This is good! But the AI doesn't always use it.

**Changes needed:**
1. Review how search results are formatted as MCP `text` content (likely in `scripts/gospel-mcp/internal/mcp/server.go`)
2. Ensure `MarkdownLink` appears at the TOP of each result, not buried in JSON
3. Add a footer to search results: "Use these markdown links in study documents. Always verify quotes with read_file."

---

## Implementation Order

Recommended sequence, minimizing risk:

### Batch 1: Output formatting (no new tools, no re-indexing)
1. **#7** — Update tool description (1 line change)
2. **#8** — Add truncation warning (3 line change)
3. **#2** — Add result type labels (10 line change)
4. **#1** — Add markdown links to results (20 line change, new helper function)
5. **#4** — Enrich conference metadata display (15 line change)
6. **#3** — Add local file existence check (10 line change)

**Test:** Run `gospel-vec serve` and query via MCP client. Verify results now show links, labels, and file status. No re-indexing needed.

### Batch 2: New tools
7. **#5** — Add `get_talk` tool (new function ~80 lines)
8. **#6** — Add `search_talks` filtered tool (new function ~80 lines, new search method ~30 lines)

**Test:** Verify `get_talk` returns full talk content and `search_talks` filters by speaker/year correctly.

### Batch 3: gospel-mcp alignment
9. **#9** — Review and improve gospel-mcp result formatting

---

## Re-Indexing Considerations

None of these changes require re-indexing the vector database. The metadata fields (`Speaker`, `Position`, `TalkTitle`, `FilePath`, etc.) are **already stored** in the index from the original indexing runs. These improvements only change how that stored data is **presented** in MCP tool responses.

If future improvements need new metadata fields (e.g., a `quote_layer` distinction stored at index time), that would require re-indexing. But everything in this plan works with existing data.

**Current index stats:**
- 210,011 documents across 8 collections
- Storage: ~2.47 GB across scriptures.gob.gz, conference.gob.gz, manual.gob.gz
- Layers: verse, paragraph, summary, theme
- Sources: scriptures (OT, NT, BoM, D&C, PGP), conference (1971-2025), manual

---

## Success Criteria

After implementing all changes, the AI should:

1. **Always have a clickable link** in search results to follow up with `read_file`
2. **Know whether a result is a quote or a summary** before deciding to use it
3. **Know whether the source file exists locally** before claiming it doesn't
4. **See rich conference talk metadata** (title, speaker, session) without opening the file
5. **Be able to get full talk text** via `get_talk` without needing to know file paths
6. **Be reminded by the tool itself** to verify content before quoting

The goal: the tools should actively guide the AI toward the two-phase workflow (discovery → deep reading) rather than enabling the shortcut of treating search results as final answers.

---

*This plan implements suggestion #3 from [01_reflections.md](01_reflections.md)*
