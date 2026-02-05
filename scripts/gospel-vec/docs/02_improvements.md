# Gospel-Vec Tool Improvements

Notes on improvements and issues observed during real-world usage of the gospel-vec semantic search tool.

---

## Session: Priesthood Study (2025-01)

### Tool Performance Observations

#### What Worked Excellently

1. **Conceptual Matching**
   - Query: "charity" found [Jacob 2:17](../../gospel-library/eng/scriptures/bofm/jacob/2.md) which talks about giving to the poor *without* using the word "charity"
   - This is the killer feature - finding scripture by *meaning* rather than exact words

2. **Multi-Layer Results**
   - Verse, paragraph, summary, and theme layers all contributed unique insights
   - Theme layer was especially good for identifying doctrinal sections (e.g., "Melchizedek as model high priest" for Alma 13:14-19)

3. **Cross-Volume Connections**
   - Single query surfaces results from D&C, Book of Mormon, and Pearl of Great Price simultaneously
   - Example: "Melchizedek" query found D&C 107, Alma 13, and D&C 84 in one search

4. **Summary Layer Value**
   - Chapter summaries provided helpful overview context
   - Keywords in summaries help with topical navigation

#### Issues Encountered

1. **`get_chapter` Tool Error**
   - **Query**: book="dc", chapter=84
   - **Result**: "Chapter not found"
   - **Expected**: Full text of D&C 84
   - **Workaround**: Used `read_file` directly on the markdown file
   - **Probable Cause**: Book name format mismatch. The tool likely expects a different identifier for D&C.
   - **Status**: ✅ FIXED - Added `NormalizeBookName()` that accepts many formats

2. **Book Name Standardization**
   - What formats are supported? Need documentation:
     - `dc` vs `D&C` vs `dc-testament/dc`?
     - `1-ne` vs `1 Nephi` vs `1ne`?
   - Suggest: Add help text or error message showing valid book identifiers
   - **Status**: ✅ FIXED - All formats now work, added `list_books` tool

---

## Feature Requests

### High Priority

1. **Index General Conference Talks**
   - Massive source of doctrine and application
   - Would enable searches like "faith during trials" to surface relevant talks
   - 50+ years of content available in `/gospel-library/eng/general-conference/`

2. **Index Come, Follow Me Manuals**
   - Current study curriculum
   - Would connect scripture study with weekly lessons
   - Located in `/gospel-library/eng/manual/come-follow-me-*`

3. **Index Teaching in the Savior's Way**
   - Essential for lesson preparation
   - Located in `/gospel-library/eng/manual/teaching-in-the-saviors-way-2022/`

### Medium Priority

4. **Old Testament and New Testament**
   - Complete standard works coverage
   - Located in `/gospel-library/eng/scriptures/ot/` and `/gospel-library/eng/scriptures/nt/`

5. **Study Aids**
   - Topical Guide (`/gospel-library/eng/scriptures/tg/`)
   - Bible Dictionary (`/gospel-library/eng/scriptures/bd/`)
   - Guide to the Scriptures (`/gospel-library/eng/scriptures/gs/`)

6. **Teachings of Presidents Series**
   - 17 volumes of prophetic teachings
   - Located in `/gospel-library/eng/manual/teachings-*/`

### Lower Priority (But Cool)

7. **Cross-Reference Awareness**
   - When a scripture has footnotes pointing to other scriptures, weight those connections in search
   - Example: D&C 84:33 footnotes link to related passages about magnifying callings

8. **Search Result Grouping**
   - Option to group results by book/section
   - Helps see "all results from Alma" vs scattered individual verses

---

## Data Quality Observations

### Summary Generation Quality

The AI-generated summaries are generally high quality:
- Accurate doctrinal keywords
- Good thematic identification
- Useful for discovering chapters on a topic

Occasional issues:
- Some summaries are very long (could be trimmed)
- Keyword lists sometimes have duplicates (e.g., "Jerusalem" repeated in 1 Nephi 13 summary)
  - **Status**: ✅ FIXED - Added `deduplicateKeywords()` post-processing

### Theme Layer Quality

Themes are excellent for section-level searching:
- "Melchizedek as model high priest" - perfect for finding Alma 13:14-19
- "Details roles and powers of Melchizedek Priesthood" - surfaces D&C 107:11-20
- "Warning on priesthood misuse and call to righteous dominion" - finds D&C 121:33-46

### Embedding Quality

Match percentages seem well-calibrated:
- 78% match for D&C 121:41 on "unrighteous dominion" query - very relevant
- 70%+ matches are consistently high-value
- 60-70% matches are related but tangential
- Below 60% often less useful

---

## Technical Notes

### MCP Server Notes

- Must use **absolute path** for `-data` flag when running as MCP server
- Working directory for MCP is VS Code's workspace root, not the script directory
- Auto-detection of chat model now works (falls back to LM Studio available models)

### Indexing Performance

Approximate times on local machine:
- Book of Mormon: ~16 minutes with summary layer
- Pearl of Great Price: ~8 minutes with summary layer  
- Doctrine and Covenants: ~16 minutes with summary layer

Summary generation is the bottleneck (requires LLM inference per chunk).

---

## Ideas for Future Development

1. **Incremental Indexing**
   - Only re-index changed files
   - Would speed up adding new conference talks

2. **Export/Import Database**
   - Share pre-built indexes
   - Avoid re-indexing on new machines

3. **Web Interface**
   - Simple browser UI for search
   - Could complement MCP tooling

4. **Citation Generation**
   - Output formatted citations for talks/papers
   - Include URL links to churchofjesuschrist.org

5. **Reading Plan Integration**
   - Connect with Come, Follow Me schedule
   - "What should I read this week?" with semantic expansion

---

*Last updated: During priesthood oath and covenant study*
