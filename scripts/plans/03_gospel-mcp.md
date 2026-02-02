# Gospel MCP Server Design Document

*"For I, the Lord God, created all things, of which I have spoken, spiritually, before they were naturally upon the face of the earth."* — [Moses 3:5](../../gospel-library/eng/scriptures/pgp/moses/3.md)

---

## Purpose & Vision

### The Problem
With ~10,000+ markdown files in `/gospel-library/`, traditional file search becomes slow and context windows get overwhelmed. More critically, narrow searches can miss the forest for the trees—returning a single verse without the surrounding context, cross-references, or thematic connections that reveal deeper meaning.

### The Goal
Create an MCP server that gives AI assistants (and human users) **context-rich access** to gospel content:
- Fast full-text search across all content
- **Generous context** around search results (not just snippets)
- Cross-reference awareness
- Links back to original markdown files for VS Code navigation
- Semantic understanding of scripture structure (books, chapters, verses)

### Design Principles

1. **Context Over Snippets**
   - When returning scripture, include surrounding verses (configurable window)
   - When returning a verse, also return chapter summary and cross-references
   - "Line upon line, precept upon precept" requires seeing the lines together

2. **Preserve All Links**
   - Results include `file_path` to local markdown (for VS Code navigation)
   - Results include `source_url` to churchofjesuschrist.org (for offline search capability)
   - Related references extracted and returned with every result
   - Human users can click through in VS Code OR open on Church website

3. **Structured + Full Text**
   - Structured data (book, chapter, verse) enables precise queries
   - Full-text search enables thematic exploration
   - Both are essential for scripture study

4. **Minimal, Powerful Tools**
   - 3 tools instead of 10+ specialized ones
   - Each tool is multipurpose with smart filtering
   - Follows Unix philosophy: do one thing well, combine for power

5. **Keep It Simple**
   - SQLite is sufficient (no server process needed)
   - Single Go binary for both indexing and serving
   - Minimal dependencies

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     VS Code / Copilot                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ MCP Protocol (stdio)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Gospel MCP Server (Go)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Search    │  │   Lookup    │  │   Cross-Reference   │  │
│  │   Tools     │  │   Tools     │  │      Tools          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  SQLite Database (FTS5)                      │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌──────────┐ │
│  │ scriptures│  │   talks   │  │  manuals  │  │ fts_*    │ │
│  └───────────┘  └───────────┘  └───────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Indexed from
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              /gospel-library/eng/ (Markdown Files)           │
│  scriptures/  general-conference/  manual/  liahona/  ...   │
└─────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Tables

#### `scriptures`
Stores individual verses with full structural metadata.

```sql
CREATE TABLE scriptures (
    id INTEGER PRIMARY KEY,
    volume TEXT NOT NULL,        -- 'ot', 'nt', 'bofm', 'dc-testament', 'pgp'
    book TEXT NOT NULL,          -- 'gen', 'matt', '1-ne', 'dc', 'moses'
    chapter INTEGER NOT NULL,
    verse INTEGER NOT NULL,
    text TEXT NOT NULL,          -- Verse text (markdown)
    file_path TEXT NOT NULL,     -- Relative path: 'gospel-library/eng/scriptures/ot/gen/1.md'
    source_url TEXT NOT NULL,    -- https://www.churchofjesuschrist.org/study/scriptures/ot/gen/1?lang=eng
    
    UNIQUE(volume, book, chapter, verse)
);

-- Full-text search on verse content
CREATE VIRTUAL TABLE scriptures_fts USING fts5(
    text,
    content='scriptures',
    content_rowid='id'
);
```

#### `chapters`
Stores chapter-level content for context retrieval.

```sql
CREATE TABLE chapters (
    id INTEGER PRIMARY KEY,
    volume TEXT NOT NULL,
    book TEXT NOT NULL,
    chapter INTEGER NOT NULL,
    title TEXT,                  -- Chapter heading/summary if available
    full_content TEXT NOT NULL,  -- Full chapter markdown
    file_path TEXT NOT NULL,
    
    UNIQUE(volume, book, chapter)
);
```

#### `talks`
General conference talks and other addresses.

```sql
CREATE TABLE talks (
    id INTEGER PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,      -- 4 or 10
    session TEXT,                -- 'saturday-morning', 'priesthood', etc.
    speaker TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,       -- Full talk markdown
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,    -- https://www.churchofjesuschrist.org/study/...
    
    UNIQUE(file_path)
);

CREATE VIRTUAL TABLE talks_fts USING fts5(
    title,
    speaker,
    content,
    content='talks',
    content_rowid='id'
);
```

#### `manuals`
Come Follow Me, handbooks, teaching guides, magazines, etc.

```sql
CREATE TABLE manuals (
    id INTEGER PRIMARY KEY,
    content_type TEXT NOT NULL,  -- 'manual', 'magazine', 'handbook'
    collection_id TEXT NOT NULL, -- 'come-follow-me-for-home-and-church-old-testament-2026'
    section TEXT,                -- Lesson number or section identifier
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,    -- https://www.churchofjesuschrist.org/study/...
    
    UNIQUE(file_path)
);

CREATE VIRTUAL TABLE manuals_fts USING fts5(
    title,
    content,
    content='manuals',
    content_rowid='id'
);
```

#### `cross_references`
Links between scriptures (from footnotes).

```sql
CREATE TABLE cross_references (
    id INTEGER PRIMARY KEY,
    source_volume TEXT NOT NULL,
    source_book TEXT NOT NULL,
    source_chapter INTEGER NOT NULL,
    source_verse INTEGER NOT NULL,
    target_volume TEXT NOT NULL,
    target_book TEXT NOT NULL,
    target_chapter INTEGER NOT NULL,
    target_verse INTEGER,        -- NULL if whole chapter reference
    reference_type TEXT,         -- 'footnote', 'tg', 'bd', 'jst'
    
    FOREIGN KEY (source_volume, source_book, source_chapter, source_verse)
        REFERENCES scriptures(volume, book, chapter, verse)
);

CREATE INDEX idx_cross_ref_source ON cross_references(source_volume, source_book, source_chapter, source_verse);
CREATE INDEX idx_cross_ref_target ON cross_references(target_volume, target_book, target_chapter, target_verse);
```

#### `index_metadata`
Tracks indexed files for incremental updates.

```sql
CREATE TABLE index_metadata (
    file_path TEXT PRIMARY KEY,
    content_type TEXT NOT NULL,  -- 'scripture', 'talk', 'manual'
    indexed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    file_mtime DATETIME NOT NULL,
    file_size INTEGER NOT NULL,
    record_count INTEGER NOT NULL  -- verses/sections indexed from this file
);

CREATE INDEX idx_metadata_mtime ON index_metadata(file_mtime);
CREATE INDEX idx_metadata_type ON index_metadata(content_type);
```

---

## MCP Tools (Consolidated)

We provide **3 powerful tools** instead of many specialized ones. Each tool is multipurpose with smart filtering.

### Search Syntax

The `search` tool supports rich query syntax (FTS5-based, similar to VS Code search):

| Syntax | Example | Description |
|--------|---------|-------------|
| **Simple terms** | `faith hope` | Match documents containing both words |
| **Exact phrase** | `"natural man"` | Match exact phrase |
| **OR** | `faith OR hope` | Match either term |
| **NOT** | `faith NOT doubt` | Exclude documents with term |
| **Prefix** | `intellig*` | Match intelligence, intelligent, etc. |
| **Column filter** | `speaker:nelson` | Filter by specific field |
| **Grouping** | `(faith OR hope) AND charity` | Complex boolean logic |

**Column filters by content type:**
- **Scriptures**: `book:`, `volume:`, `chapter:`
- **Conference**: `speaker:`, `title:`, `year:`
- **Manuals**: `title:`, `collection:`

---

### Tool 1: `gospel_search`

**Purpose:** Full-text search across all gospel content with filtering.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| query | string | yes | Search query (supports full syntax above) |
| source | string | no | Filter: `scriptures`, `conference`, `manual`, `magazine`, `all` (default: `all`) |
| path | string | no | Narrow to path: `bofm`, `2024/10`, `come-follow-me-*`, etc. |
| limit | int | no | Max results (default: 20, max: 100) |
| context | int | no | Lines/verses of context around match (default: 3) |
| include_content | bool | no | Return full content, not just excerpts (default: false) |

**Example Calls:**

```json
// Search all content for exact phrase
{"query": "\"glory of God is intelligence\""}

// Search scriptures only, Book of Mormon
{"query": "faith hope charity", "source": "scriptures", "path": "bofm"}

// Search conference talks by speaker
{"query": "speaker:nelson confidence", "source": "conference"}

// Search 2025 conference talks
{"query": "peace", "source": "conference", "path": "2025"}

// Search Come Follow Me manuals
{"query": "creation", "source": "manual", "path": "come-follow-me-*"}

// Prefix search with full content
{"query": "intellig*", "source": "scriptures", "include_content": true}
```

**Returns:**

```json
{
  "query": "\"natural man\"",
  "total_matches": 12,
  "results": [
    {
      "reference": "Mosiah 3:19",
      "title": "Mosiah 3",
      "excerpt": "For the **natural man** is an enemy to God, and has been from the fall of Adam...",
      "content": "...(full verse/section if include_content=true)...",
      "context_before": ["...verse 17...", "...verse 18..."],
      "context_after": ["...verse 20...", "...verse 21..."],
      "file_path": "gospel-library/eng/scriptures/bofm/mosiah/3.md",
      "source_url": "https://www.churchofjesuschrist.org/study/scriptures/bofm/mosiah/3?lang=eng&id=p19#p19",
      "related_references": [
        {"reference": "1 Corinthians 2:14", "type": "footnote"},
        {"reference": "Alma 26:21", "type": "footnote"},
        {"reference": "TG Natural Man", "type": "topical_guide"}
      ],
      "source_type": "scripture",
      "relevance_score": 0.98
    },
    {
      "reference": "Elder Bednar, April 2023",
      "title": "Abide in Me",
      "excerpt": "...putting off the **natural man** and becoming a saint...",
      "file_path": "gospel-library/eng/general-conference/2023/04/23bednar.md",
      "source_url": "https://www.churchofjesuschrist.org/study/general-conference/2023/04/23bednar?lang=eng",
      "related_references": [
        {"reference": "Mosiah 3:19", "type": "scripture_citation"},
        {"reference": "John 15:4", "type": "scripture_citation"}
      ],
      "source_type": "conference",
      "relevance_score": 0.85
    }
  ],
  "query_time_ms": 15
}
```

**Key Features:**
- **Unified search**: One tool searches everything
- **Smart filtering**: `source` and `path` narrow scope
- **Rich context**: Surrounding text always included
- **Related references**: Every result includes its cross-references
- **Dual URLs**: Both local file and Church website

---

### Tool 2: `gospel_get`

**Purpose:** Retrieve specific content by reference or path.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| reference | string | no | Scripture ref: `D&C 93:36`, `Moses 3:5`, `1 Nephi 3:7` |
| path | string | no | File path or pattern: `gospel-library/eng/general-conference/2025/04/57nelson.md` |
| context | int | no | Additional verses/paragraphs of context (default: 0 = just the reference) |
| include_chapter | bool | no | Return entire chapter/document (default: false) |

*One of `reference` or `path` is required.*

**Example Calls:**

```json
// Get specific scripture with context
{"reference": "D&C 93:36", "context": 5}

// Get entire chapter
{"reference": "Moses 3", "include_chapter": true}

// Get verse range
{"reference": "D&C 130:18-19"}

// Get conference talk by path
{"path": "gospel-library/eng/general-conference/2025/04/57nelson.md"}

// Get Topical Guide entry
{"reference": "TG Intelligence"}

// Get Bible Dictionary entry  
{"reference": "BD Faith"}
```

**Returns:**

```json
{
  "reference": "D&C 93:36",
  "title": "Doctrine and Covenants 93",
  "content": "The glory of God is intelligence, or, in other words, light and truth.",
  "context_before": [
    {"verse": 33, "text": "For man is spirit..."},
    {"verse": 34, "text": "And when separated..."},
    {"verse": 35, "text": "The elements are the tabernacle of God..."}
  ],
  "context_after": [
    {"verse": 37, "text": "Light and truth forsake that evil one."},
    {"verse": 38, "text": "Every spirit of man was innocent in the beginning..."}
  ],
  "chapter_content": "...(full chapter if include_chapter=true)...",
  "file_path": "gospel-library/eng/scriptures/dc-testament/dc/93.md",
  "source_url": "https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p36#p36",
  "related_references": [
    {"reference": "D&C 130:18", "type": "footnote", "text": "Whatever principle of intelligence..."},
    {"reference": "Abraham 3:19", "type": "footnote", "text": "...more intelligent than they all..."},
    {"reference": "TG Intelligence", "type": "topical_guide"},
    {"reference": "TG God, Intelligence of", "type": "topical_guide"}
  ],
  "source_type": "scripture"
}
```

**Smart Reference Parsing:**
- `D&C 93:36` → Doctrine and Covenants section 93, verse 36
- `1 Nephi 3:7` → Book of Mormon
- `Moses 3:5` → Pearl of Great Price
- `TG Faith` → Topical Guide entry
- `BD Atonement` → Bible Dictionary entry
- `GS Faith` → Guide to the Scriptures

---

### Tool 3: `gospel_list`

**Purpose:** Browse and discover available content (for exploration, not search).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| source | string | no | Content type: `scriptures`, `conference`, `manual`, `magazine`, `all` |
| path | string | no | Path to list: `bofm`, `2025`, `come-follow-me-*` |
| depth | int | no | How deep to recurse (default: 1) |

**Example Calls:**

```json
// List all scripture volumes
{"source": "scriptures"}

// List books in Book of Mormon
{"source": "scriptures", "path": "bofm"}

// List conference years
{"source": "conference"}

// List talks in April 2025
{"source": "conference", "path": "2025/04"}

// List all manuals
{"source": "manual"}

// List Come Follow Me lessons for 2026
{"source": "manual", "path": "come-follow-me-for-home-and-church-old-testament-2026"}
```

**Returns:**

```json
{
  "path": "scriptures/bofm",
  "items": [
    {"name": "1 Nephi", "path": "bofm/1-ne", "type": "book", "chapters": 22},
    {"name": "2 Nephi", "path": "bofm/2-ne", "type": "book", "chapters": 33},
    {"name": "Jacob", "path": "bofm/jacob", "type": "book", "chapters": 7},
    {"name": "Enos", "path": "bofm/enos", "type": "book", "chapters": 1},
    // ...
  ],
  "total": 15
}
```

---

## Search Examples for Common Study Tasks

### Finding Scriptures on a Topic

```json
// What does the Book of Mormon teach about faith?
{"query": "faith", "source": "scriptures", "path": "bofm", "limit": 30}

// Find all uses of "intelligence" in D&C
{"query": "intelligence", "source": "scriptures", "path": "dc-testament"}

// Exact phrase in Pearl of Great Price
{"query": "\"spiritual creation\"", "source": "scriptures", "path": "pgp"}
```

### Finding Conference Talks

```json
// President Nelson on temples
{"query": "speaker:nelson temple", "source": "conference"}

// Talks about repentance from 2020-2025
{"query": "repentance", "source": "conference", "path": "202*"}

// Elder Holland's talks with full content
{"query": "speaker:holland", "source": "conference", "include_content": true, "limit": 5}
```

### Cross-Reference Research

```json
// Get D&C 93:36 with all its cross-references
{"reference": "D&C 93:36", "context": 10}

// Then follow up with a related reference
{"reference": "D&C 130:18-19", "context": 5}

// Get the Topical Guide entry
{"reference": "TG Intelligence"}
```

### Lesson Preparation

```json
// Find Come Follow Me lesson content
{"query": "Moses creation", "source": "manual", "path": "come-follow-me-*"}

// Search Teaching in the Savior's Way
{"query": "questions", "source": "manual", "path": "teaching-in-the-saviors-way*"}
```

---

## Context Strategy for AI Study

### The Challenge
When Claude receives a search result with just a verse, important context is lost:
- Surrounding verses that complete the thought
- Chapter theme and flow
- Cross-references that illuminate meaning
- Historical/textual context

### The Solution: Always Include Related References

Every result from `gospel_search` and `gospel_get` includes a `related_references` array containing:
- **Footnote links**: Scripture cross-references from the source text
- **Topical Guide entries**: TG references mentioned
- **Scripture citations**: When a talk cites scriptures
- **See also**: Related topics from study aids

This means you never get an isolated result—you always see the web of connections.

### Context Depth Options

**Layer 1: Default (context=3)**
- The matched content
- 3 verses/paragraphs before and after
- All related references extracted from source
- File path + source URL

**Layer 2: Expanded (context=10+)**
- More surrounding content
- Better for understanding flow of argument
- Good for talks and longer passages

**Layer 3: Full Document (include_chapter=true or include_content=true)**
- Entire chapter or document
- AI can see complete context
- Best for deep study

### Example Study Flow

User asks: *"Help me understand what 'intelligence' means in D&C 93"*

**Step 1: Search for context**
```json
gospel_search({"query": "intelligence", "source": "scriptures", "path": "dc-testament/dc/93", "context": 10})
```
Returns D&C 93:29-40 with surrounding verses and related references.

**Step 2: Follow a cross-reference**
```json
gospel_get({"reference": "D&C 130:18-19", "context": 5})
```
Returns the related passage with ITS related references.

**Step 3: Check Topical Guide**
```json
gospel_get({"reference": "TG Intelligence"})
```
Returns comprehensive scripture list for the topic.

**Step 4: Find conference talks**
```json
gospel_search({"query": "\"glory of God is intelligence\"", "source": "conference", "include_content": true, "limit": 5})
```
Returns full talks discussing this doctrine.

**Step 5: Deep dive via read_file**
If any result warrants deeper exploration, use `read_file` on the `file_path` to get the complete original markdown with all formatting.

---

## Indexing CLI

The `gospel-mcp` binary serves dual purposes: **indexing** content and **serving** the MCP protocol.

### Commands

```bash
# Full index (drop and rebuild everything)
gospel-mcp index

# Index only specific content types
gospel-mcp index --source=scriptures
gospel-mcp index --source=conference
gospel-mcp index --source=manual

# Index only specific paths (useful for testing or partial updates)
gospel-mcp index --path="gospel-library/eng/scriptures/bofm"
gospel-mcp index --path="gospel-library/eng/general-conference/2025"

# Incremental update (only new/modified files)
gospel-mcp index --incremental

# Force full reindex (drop all tables first)
gospel-mcp index --force

# Dry run (show what would be indexed, don't write)
gospel-mcp index --dry-run

# Verbose output
gospel-mcp index --verbose

# Serve MCP protocol (what VS Code calls)
gospel-mcp serve
```

### Index Workflow

#### Initial Setup (First Time)

```bash
# From repository root
cd c:\Users\cpuch\Documents\code\stuffleberry\scripture-study

# Build the binary (optional, can use go run)
go build -o gospel-mcp.exe ./scripts/gospel-mcp/cmd/gospel-mcp

# Run full index - creates database and indexes everything
./gospel-mcp index --verbose

# Or with go run directly
go run ./scripts/gospel-mcp/cmd/gospel-mcp index --verbose
```

**Expected output:**
```
Gospel MCP Indexer v1.0.0
Database: scripts/gospel-mcp/data/gospel.db

Initializing database schema...
  ✓ Created tables: scriptures, chapters, talks, manuals, cross_references
  ✓ Created FTS5 indexes

Indexing scriptures...
  Processing gospel-library/eng/scriptures/ot/...
    ✓ Genesis: 50 chapters, 1533 verses
    ✓ Exodus: 40 chapters, 1213 verses
    ... 
  Processing gospel-library/eng/scriptures/nt/...
  Processing gospel-library/eng/scriptures/bofm/...
  Processing gospel-library/eng/scriptures/dc-testament/...
  Processing gospel-library/eng/scriptures/pgp/...
  ✓ Scriptures complete: 41,995 verses indexed

Indexing conference talks...
  Processing gospel-library/eng/general-conference/1971/...
  Processing gospel-library/eng/general-conference/1972/...
  ...
  Processing gospel-library/eng/general-conference/2025/...
  ✓ Conference complete: 8,432 talks indexed

Indexing manuals...
  Processing gospel-library/eng/manual/...
  ✓ Manuals complete: 2,847 sections indexed

Extracting cross-references...
  ✓ Cross-references complete: 127,543 links extracted

Building FTS5 indexes...
  ✓ FTS5 indexes populated

Index complete!
  Total time: 47.3s
  Database size: 156 MB
  
  Summary:
    Scriptures: 41,995 verses in 1,189 chapters
    Conference: 8,432 talks (1971-2025)
    Manuals: 2,847 sections
    Cross-refs: 127,543 links
```

#### Reindexing After New Content

**Scenario 1: Downloaded new conference talks**

```bash
# Incremental - only indexes new/changed files
gospel-mcp index --incremental

# Or target just the new content
gospel-mcp index --path="gospel-library/eng/general-conference/2025/10"
```

**Scenario 2: Full refresh (after major changes)**

```bash
# Drop everything and rebuild
gospel-mcp index --force
```

**Scenario 3: Reindex just scriptures (if formatting changed)**

```bash
gospel-mcp index --source=scriptures --force
```

### Incremental Index Strategy

The `--incremental` flag uses file modification times to detect changes:

```sql
-- Track indexed files and their timestamps
CREATE TABLE index_metadata (
    file_path TEXT PRIMARY KEY,
    indexed_at DATETIME NOT NULL,
    file_mtime DATETIME NOT NULL,
    file_hash TEXT  -- Optional: SHA256 for content-based detection
);
```

**Algorithm:**
1. Walk `gospel-library/` directory
2. For each file, check `index_metadata`:
   - If not present → index it
   - If `file_mtime` changed → re-index it
   - If unchanged → skip
3. Optionally: Remove entries for deleted files

**Trade-offs:**
- **Full reindex**: ~1 minute, guaranteed consistency
- **Incremental**: ~5 seconds for typical updates, may miss edge cases

**Recommendation:** Use `--incremental` for day-to-day updates, `--force` after major downloads or when troubleshooting.

### Automation Ideas

#### Git Hook (post-merge)

Create `.git/hooks/post-merge`:
```bash
#!/bin/bash
# Reindex after pulling new content
if git diff --name-only HEAD@{1} HEAD | grep -q "^gospel-library/"; then
    echo "Gospel library content changed, reindexing..."
    go run ./scripts/gospel-mcp/cmd/gospel-mcp index --incremental
fi
```

#### VS Code Task

Add to `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Gospel MCP: Reindex",
      "type": "shell",
      "command": "go run ./scripts/gospel-mcp/cmd/gospel-mcp index --incremental",
      "group": "build",
      "problemMatcher": []
    },
    {
      "label": "Gospel MCP: Full Reindex",
      "type": "shell", 
      "command": "go run ./scripts/gospel-mcp/cmd/gospel-mcp index --force",
      "group": "build",
      "problemMatcher": []
    }
  ]
}
```

Then run with `Ctrl+Shift+P` → "Tasks: Run Task" → "Gospel MCP: Reindex"

#### After Downloading with TUI

The gospel-library downloader TUI could trigger reindexing:
```go
// After successful download
exec.Command("go", "run", "./scripts/gospel-mcp/cmd/gospel-mcp", "index", "--incremental").Run()
```

Or prompt the user: "New content downloaded. Reindex? [Y/n]"

---

## Implementation Phases

### Phase 1: Core Infrastructure (Day 1)
- [ ] Create project structure: `scripts/gospel-mcp/`
- [ ] Set up Go module with dependencies (sqlite, mcp-go)
- [ ] Implement SQLite database initialization with FTS5
- [ ] Create schema with source_url fields
- [ ] Build URL generator for churchofjesuschrist.org paths

### Phase 2: Indexer (Day 1-2)
- [ ] Scripture parser
  - [ ] Parse verse structure (handle `**1.**` format)
  - [ ] Extract footnote references (cross-references)
  - [ ] Generate source URLs
- [ ] Conference talk parser
  - [ ] Extract speaker, date, title
  - [ ] Extract scripture citations from footnotes
  - [ ] Generate source URLs
- [ ] Manual/magazine parser
  - [ ] Handle various manual structures
  - [ ] Generate source URLs
- [ ] CLI command: `gospel-mcp index [--path=...]`

### Phase 3: MCP Server + Tools (Day 2-3)
- [ ] MCP protocol implementation (stdio transport)
- [ ] `gospel_search` tool
  - [ ] FTS5 query builder
  - [ ] Source/path filtering
  - [ ] Context extraction
  - [ ] Related reference extraction
- [ ] `gospel_get` tool
  - [ ] Smart reference parser (scripture, TG, BD, GS)
  - [ ] Path-based retrieval
  - [ ] Context/chapter options
- [ ] `gospel_list` tool
  - [ ] Directory-style browsing
  - [ ] Depth control

### Phase 4: Integration & Testing (Day 3-4)
- [ ] VS Code MCP configuration
- [ ] Test search scenarios:
  - [ ] Simple terms, phrases, exact phrases
  - [ ] Boolean operators
  - [ ] Column filters (speaker:, book:, etc.)
  - [ ] Path filtering
- [ ] Test retrieval scenarios:
  - [ ] Scripture references (all formats)
  - [ ] Topical Guide / Bible Dictionary
  - [ ] Conference talks by path
- [ ] Performance tuning
- [ ] Documentation

---

## Project Structure

```
scripts/gospel-mcp/
├── go.mod
├── go.sum
├── README.md
├── cmd/
│   └── gospel-mcp/
│       ├── main.go           # CLI entry point
│       ├── index.go          # `index` command implementation
│       └── serve.go          # `serve` command implementation
├── internal/
│   ├── db/
│   │   ├── db.go             # Database connection, migrations
│   │   ├── schema.sql        # Table definitions
│   │   ├── queries.go        # Prepared statements
│   │   ├── url.go            # Source URL generation
│   │   └── metadata.go       # index_metadata operations
│   ├── indexer/
│   │   ├── indexer.go        # Main indexing orchestration
│   │   ├── scripture.go      # Scripture parsing + cross-ref extraction
│   │   ├── talk.go           # Conference talk parsing
│   │   ├── manual.go         # Manual/magazine parsing
│   │   └── walker.go         # File system walking + change detection
│   ├── mcp/
│   │   ├── server.go         # MCP server implementation
│   │   └── protocol.go       # MCP protocol types
│   └── tools/
│       ├── search.go         # gospel_search implementation
│       ├── get.go            # gospel_get implementation
│       ├── list.go           # gospel_list implementation
│       └── reference.go      # Smart reference parser
└── data/
    └── gospel.db             # SQLite database (gitignored)
```

---

## VS Code Configuration

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "gospel-mcp": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "./scripts/gospel-mcp/cmd/gospel-mcp", "serve"],
      "env": {
        "GOSPEL_DB_PATH": "${workspaceFolder}/scripts/gospel-mcp/data/gospel.db",
        "GOSPEL_LIBRARY_PATH": "${workspaceFolder}/gospel-library"
      }
    }
  }
}
```

---

## Success Criteria

1. **Speed**: Search returns results in <100ms
2. **Context**: Results include surrounding text + all related references
3. **Links**: Every result includes both local `file_path` AND `source_url`
4. **Offline Church Search**: Can function as complete offline search for churchofjesuschrist.org/study
5. **Coverage**: All scriptures, 50+ years of conference talks, major manuals indexed
6. **Simplicity**: 3 tools that are intuitive and powerful
7. **Search Power**: Supports terms, phrases, exact phrases, boolean, prefix, column filters

---

## URL Generation Strategy

To enable offline search for the Church website, we need to generate accurate `source_url` values.

### URL Patterns

| Content Type | URL Pattern |
|--------------|-------------|
| **Scripture verse** | `https://www.churchofjesuschrist.org/study/scriptures/{volume}/{book}/{chapter}?lang=eng&id=p{verse}#p{verse}` |
| **Scripture chapter** | `https://www.churchofjesuschrist.org/study/scriptures/{volume}/{book}/{chapter}?lang=eng` |
| **Conference talk** | `https://www.churchofjesuschrist.org/study/general-conference/{year}/{month}/{filename}?lang=eng` |
| **Manual section** | `https://www.churchofjesuschrist.org/study/manual/{manual-id}/{section}?lang=eng` |
| **Topical Guide** | `https://www.churchofjesuschrist.org/study/scriptures/tg/{topic}?lang=eng` |
| **Bible Dictionary** | `https://www.churchofjesuschrist.org/study/scriptures/bd/{entry}?lang=eng` |

### Examples

```
# D&C 93:36
https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p36#p36

# 1 Nephi 3:7  
https://www.churchofjesuschrist.org/study/scriptures/bofm/1-ne/3?lang=eng&id=p7#p7

# President Nelson, April 2025
https://www.churchofjesuschrist.org/study/general-conference/2025/04/57nelson?lang=eng

# Come Follow Me 2026, Genesis 1-2
https://www.churchofjesuschrist.org/study/manual/come-follow-me-for-home-and-church-old-testament-2026/03?lang=eng
```

---

## Open Questions

1. **Verse Parsing Complexity**: Our markdown files have verse numbers in bold (`**1.**`). Need to handle edge cases:
   - Verse ranges in a single paragraph
   - Poetry formatting (Psalms, Isaiah)
   - JST footnotes
   - **Decision**: Parse incrementally, handle common cases first

2. **Cross-Reference Extraction**: Footnotes use various formats:
   - Scripture links: `[Gen. 1:1](../../ot/gen/1.md)`
   - Topical Guide: `[TG Faith](../../tg/faith.md)`
   - **Decision**: Parse ALL footnote links, categorize by type

3. **Incremental Updates**: Should we support updating just changed files?
   - **Decision**: Start with full reindex (~1 minute). Add incremental later if needed.

4. **Conference Talk Metadata**: Parse scripture citations from footnotes?
   - **Decision**: Yes! This enables "find talks that cite this scripture"

5. **Search Ranking**: How to balance exact match vs. relevance?
   - **Decision**: Use FTS5's built-in BM25 ranking, with boost for exact phrases

---

## Notes

*This document is the "spiritual creation" of our gospel-mcp project. It defines the vision and architecture before we write any code. As we implement, we may discover adjustments needed—but the core principles should remain stable.*

---

*Created: February 1, 2026*
*Status: Planning*
