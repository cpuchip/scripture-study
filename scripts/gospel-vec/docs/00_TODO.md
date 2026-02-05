# Gospel Vec - Semantic Scripture Search

A multi-layer vector search system for scriptures and related content.

## Vision

Build a powerful, general-purpose semantic search tool that can:
1. Search scriptures at multiple granularities (verse, paragraph, chapter)
2. Generate and search LLM summaries of content
3. Detect and search narrative themes
4. Cross-reference scriptures ↔ conference talks
5. Visualize semantic relationships and clusters

---

## 🎯 Active TODO List

### 🔧 Immediate Fixes (Before More Indexing)

- [x] **Fix `get_chapter` book name matching** ✅ DONE
  - Added `NormalizeBookName()` function maps inputs ("dc", "D&C", "1-ne", "1 Nephi") to canonical form
  - Added `bookAliasMap` with 100+ aliases for various input styles
  - Improved error message shows available books and suggestions
  - See [chunking.go](../chunking.go) for implementation

- [x] **Add `list_books` tool** ✅ DONE
  - New MCP tool to discover what books are indexed
  - Can filter by volume: `list_books(volume: "bofm")` 
  - Groups books by volume (BoM, D&C, PGP, OT, NT)

- [ ] **Add book name help documentation**
  - Document all valid book identifiers in tool description
  - Tool description now mentions: "Accepts various formats: '1 Nephi', '1-ne', '1nephi', 'D&C', 'dc', 'Alma', etc."

- [ ] **Incremental indexing**
  - Skip chapters already in database
  - Would speed up adding new content types

### 📚 Content Type Support (All of `/gospel-library/**`)

Priority order based on study workflow value:

| Priority | Content Type | Path | Status | Notes |
|----------|-------------|------|--------|-------|
| 1 | Standard Works (BoM, D&C, PGP) | `/scriptures/bofm/`, `/dc-testament/dc/`, `/pgp/` | ✅ Done | Working well |
| 2 | **General Conference Talks** | `/general-conference/{year}/{month}/` | ✅ Done | Parser + indexer complete |
| 3 | **Come, Follow Me (2026)** | `/manual/come-follow-me-*` | 🔜 Soon | Supports weekly study |
| 4 | **Teaching in the Savior's Way** | `/manual/teaching-in-the-saviors-way-2022/` | 🔜 Soon | Lesson prep |
| 5 | Old Testament | `/scriptures/ot/` | Planned | Large (~39 books) |
| 6 | New Testament | `/scriptures/nt/` | Planned | Medium (~27 books) |
| 7 | Bible Dictionary | `/scriptures/bd/` | Planned | Short entries, dense cross-refs |
| 8 | Topical Guide | `/scriptures/tg/` | Planned | Reference lists, may not embed well |
| 9 | Teachings of Presidents | `/manual/teachings-*/` | Later | 17+ volumes |
| 10 | Liahona Magazine | `/liahona/` | Later | Articles |
| 11 | Videos/Broadcasts | `/video/`, `/broadcasts/` | Later | If transcripts available |

### Conference Talk Indexing Progress ✨ DONE

**Parser completed** (`talk_parser.go`):
- [x] `ParseTalkFile()` - extracts metadata and content
- [x] `TalkMetadata` struct with: Title, Speaker, Position, Year, Month, Session, AudioURL
- [x] `FindTalkFiles()` - discovers talks by year
- [x] `ExtractScriptureReferences()` - finds scripture cross-refs
- [x] Session code parsing (e.g., `57nelson.md` → "Sunday Afternoon")
- [x] `ChunkTalkByParagraph()` - creates paragraph chunks
- [x] `ChunkTalkAsSummary()` - creates summary chunks
- [x] `IsAdministrativeDocument()` - filters sustaining/audit docs

**Test command** (`gospel-vec talks`):
- [x] `-sample` - Parse sample talks from each decade (1971-2025)
- [x] `-parse FILE` - Parse specific talk
- [x] `-summarize YEAR` - Test summary generation
- [x] `-list` - List available conference years
- [x] Filters out administrative docs (sustaining, audit reports)

**Index command** (`gospel-vec index-talks`):
- [x] `-years 2025,2024` - Index specific years
- [x] `-layers paragraph,summary` - Which layers to index
- [x] `-max N` - Limit number of talks
- [x] `-summary` - Generate AI summaries
- [x] Auto-caches summaries for reuse

**Extended types** (`types.go`):
- [x] Added conference-specific metadata fields: Speaker, Position, Year, Month, Session, TalkTitle
- [x] Updated ToMap/FromMap functions

**Unified Search** ✅:
- [x] Search now queries BOTH scriptures and conference talks by default
- [x] Can filter by layer: `-layers summary` returns only summaries
- [x] Results sorted by similarity score across all sources

**Current stats (as of indexing):**
- `conference-paragraph`: 215 documents (8 talks from 2025)
- `conference-summary`: 8 documents

**TODO for talk indexing:**
- [ ] Index full decades (expensive but valuable)
- [ ] Add MCP tools: `search_talks`, `get_talk`, `list_talks_by_speaker`
- [ ] Consider "quote" layer for memorable passages

See [03_content-indexing-guide.md](03_content-indexing-guide.md) for detailed structure analysis.

### 🤖 Embedding Model Experimentation

Currently using: **qwen3-vl-8b** (seemed good and recent)

There are MANY models to choose from! Should experiment to find what works best for our use case.

**Candidates to try:**
- [ ] `text-embedding-nomic-embed-text-v1.5` - Popular, good quality
- [ ] `text-embedding-mxbai-embed-large-v1` - Larger, potentially better matching
- [ ] `text-embedding-bge-base-en-v1.5` - BGE family, good for retrieval
- [ ] `text-embedding-gte-large` - GTE family, competitive with OpenAI
- [ ] OpenAI's `text-embedding-3-small` or `text-embedding-3-large` (if using API)

**What to evaluate:**
- Match quality (does "charity" find Jacob 2:17 which talks about poor without using "charity"?)
- Speed (embeddings per second)
- Memory usage
- Handling of scriptural language (archaic English, Restoration terminology)

**Testing approach:**
1. Create a test set of ~20 queries with expected results
2. Run each model, score by recall@5 and recall@10
3. Note which model handles conceptual matches best

### 🛡️ Quality Guardrails

**Repeating term issue:** AI summaries sometimes have duplicated keywords (e.g., "Jerusalem" repeated).

- [x] **Post-process summaries** in code ✅ DONE
  - Added `deduplicateKeywords()` function - case-insensitive dedup
  - Applied automatically in `SummarizeChapter()`
  - See [summary.go](../summary.go)

- [ ] **Prompt tuning**
  - Refine summary/theme prompts to reduce repetition
  - Add "do not repeat keywords" instruction

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Gospel Vec Search System                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Storage: Single compressed file (chromem-go export)             │
│  Embeddings: LM Studio (configurable model)                      │
│  Summaries: LM Studio chat (configurable model)                  │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│                        INDEX LAYERS                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  LAYER 1: Atomic (verse-level)                                   │
│  ├─ Collection: "scriptures-verse"                               │
│  ├─ Granularity: Individual verses                               │
│  ├─ Use case: Precise scripture lookup                           │
│  └─ Example: "What verse mentions 'go and do'?"                  │
│                                                                  │
│  LAYER 2: Contextual (paragraph-level)                           │
│  ├─ Collection: "scriptures-paragraph"                           │
│  ├─ Granularity: 3-5 verse chunks (natural breaks)               │
│  ├─ Use case: Finding passages with context                      │
│  └─ Example: "What passages teach about faith?"                  │
│                                                                  │
│  LAYER 3: Summary (chapter-level, LLM-generated)                 │
│  ├─ Collection: "scriptures-summary"                             │
│  ├─ Granularity: Chapter summaries (AI-generated)                │
│  ├─ Use case: Topic/theme discovery                              │
│  └─ Example: "Which chapters discuss the Atonement?"             │
│                                                                  │
│  LAYER 4: Themes (narrative ranges)                              │
│  ├─ Collection: "scriptures-themes"                              │
│  ├─ Granularity: Story arcs (verses 1-10, 11-20, etc.)           │
│  ├─ Use case: Narrative search                                   │
│  └─ Example: "Where does Nephi obtain the brass plates?"         │
│                                                                  │
│  LAYER 5: Cross-references (future)                              │
│  ├─ Collection: "conference-talks"                               │
│  ├─ Collection: "talk-citations"                                 │
│  ├─ Granularity: Talk paragraphs + citation links                │
│  ├─ Use case: Finding related talks                              │
│  └─ Example: "What talks cite 1 Nephi 3:7?"                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Completed Phases

### Phase 1: Foundation ✅
- [x] chromem-go integration with LM Studio
- [x] Basic chunking strategies tested
- [x] Persistence patterns understood

### Phase 2: Clean Storage ✅
- [x] In-memory DB with single-file export
- [x] Compression enabled (.gob.gz)
- [x] .gitignore data files
- [x] Config-driven (embedding model, paths, etc.)

### Phase 3: Multi-Layer Indexing ✅
- [x] Unified indexing pipeline
- [x] Layer metadata (source, range, type)
- [x] Multi-layer search with limit handling

### Phase 4: LLM Summaries ✅
- [x] Chapter summary generation (ChapterSummary struct)
- [x] Theme/narrative detection (ThemeRange JSON)
- [x] Timing output for estimation
- [x] Summary caching (JSON files with model + prompt version)
- [x] Cache invalidation on model/prompt changes

### Phase 6: Search Interface ✅
- [x] Unified search across layers
- [x] Result ranking/merging
- [x] Filter by layer/source

### Phase 7: MCP Server ✅
- [x] Expose search_scriptures as MCP tool
- [x] get_chapter tool for full chapter text

---

## Future Phases

### Phase 5: Cross-References
- [ ] Conference talk indexing (new content type!)
- [ ] Citation extraction (scripture refs in talks)
- [ ] Bidirectional links

### Phase 8: Helper Programs
- [ ] Digest tool: Summarize any file (talk, lesson, etc.) using LLM
- [ ] Cross-reference finder: Find scriptures related to a talk
- [ ] Talk summary: Generate study notes for conference talks
- [ ] Trend analyzer: Look for patterns across chapters

### Phase 9: Visualization
- [ ] Embedding clusters
- [ ] Relationship graphs
- [ ] Interactive exploration

---

## Reference

### Performance Timing (qwen3-vl-8b on RTX 4090)

Based on 3-chapter test run:
- **Summary generation**: ~7s per chapter average
- **Theme detection**: ~4s per chapter average  
- **Embedding (5 chunks)**: ~650ms per chapter

**Full Book of Mormon estimate (239 chapters):**
- Summary + Theme layers: ~44 min (239 × 11s)
- Plus embedding time: ~2.5 min (239 × 650ms)
- **Total estimate: ~47 minutes**

### Key Design Decisions

**1. Single Database File**
chromem-go's `ExportToFile` with compression gives us:
- One `.gob.gz` file instead of hundreds of `.gob` files
- Easy to backup, version, share
- In-memory operations (fast) with periodic saves

**2. Collection Naming Convention**
```
{source}-{layer}
```
Examples: `scriptures-verse`, `scriptures-summary`, `conference-paragraph`

**3. Document Metadata Schema**
```go
type DocMetadata struct {
    Source     string   // "bofm", "dc", "conference", etc.
    Layer      string   // "verse", "paragraph", "summary", "theme"
    Reference  string   // "1 Nephi 3:7" or "October 2024 - Nelson"
    Range      string   // "1-10" for themes/passages
    FilePath   string   // Path to source markdown
    Generated  bool     // true if LLM-generated
    Model      string   // LLM model used (if generated)
    Timestamp  string   // When indexed/generated
}
```

### File Structure

```
scripts/gospel-vec/
├── docs/                # Documentation (you are here!)
├── data/                # Generated data (gitignored)
│   ├── gospel-vec.gob.gz
│   └── summaries/
├── main.go              # CLI entry point
├── config.go            # Configuration
├── storage.go           # DB management
├── embed.go             # Embedding functions
├── index.go             # Indexing pipeline
├── search.go            # Search functions
├── summary.go           # LLM summary generation
├── chunking.go          # Chunking strategies
├── mcp.go               # MCP server
└── types.go             # Shared types
```

---

*This is a personal project for eternal benefit—no deadlines, just exploration and learning!*
