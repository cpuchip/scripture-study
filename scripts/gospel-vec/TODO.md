# Gospel Vec - Semantic Scripture Search

A multi-layer vector search system for scriptures and related content.

## Vision

Build a powerful, general-purpose semantic search tool that can:
1. Search scriptures at multiple granularities (verse, paragraph, chapter)
2. Generate and search LLM summaries of content
3. Detect and search narrative themes
4. Cross-reference scriptures ↔ conference talks
5. Visualize semantic relationships and clusters

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
│  LAYER 5: Cross-references                                       │
│  ├─ Collection: "conference-talks"                               │
│  ├─ Collection: "talk-citations"                                 │
│  ├─ Granularity: Talk paragraphs + citation links                │
│  ├─ Use case: Finding related talks                              │
│  └─ Example: "What talks cite 1 Nephi 3:7?"                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Phases

### Phase 1: Foundation ✅ (chromem-exp learnings)
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
- [ ] Incremental indexing (skip already-indexed)

### Phase 4: LLM Summaries ✅
- [x] Chapter summary generation (ChapterSummary struct)
- [x] Theme/narrative detection (ThemeRange JSON)
- [x] Timing output for estimation
- [ ] Summary caching (don't regenerate)

### Phase 5: Cross-References (future)
- [ ] Conference talk indexing
- [ ] Citation extraction (scripture refs in talks)
- [ ] Bidirectional links

### Phase 6: Search Interface ✅
- [x] Unified search across layers
- [x] Result ranking/merging
- [x] Filter by layer/source

### Phase 7: MCP Server (next)
- [ ] Expose search as MCP tool
- [ ] Model load/unload via LM Studio API
- [ ] Tool registration

### Phase 8: Visualization (future)
- [ ] Embedding clusters
- [ ] Relationship graphs
- [ ] Interactive exploration

---

## Performance Timing (qwen3-vl-8b on RTX 4090)

Based on 3-chapter test run:
- **Summary generation**: ~7s per chapter average
- **Theme detection**: ~4s per chapter average  
- **Embedding (5 chunks)**: ~650ms per chapter

**Full Book of Mormon estimate (239 chapters):**
- Summary + Theme layers: ~44 min (239 × 11s)
- Plus embedding time: ~2.5 min (239 × 650ms)
- **Total estimate: ~47 minutes**

---

## LM Studio v1 API

LM Studio 0.4.0+ offers model management endpoints:

```
POST /api/v1/models/load     - Load model with config
POST /api/v1/models/unload   - Unload model from memory
GET  /api/v1/models          - List loaded models
```

**Load example:**
```json
{
  "model": "qwen/qwen3-vl-8b",
  "context_length": 16384,
  "flash_attention": true
}
```

**Unload example:**
```json
{
  "instance_id": "qwen/qwen3-vl-8b"
}

---

## Key Design Decisions

### 1. Single Database File
chromem-go's `ExportToFile` with compression gives us:
- One `.gob.gz` file instead of hundreds of `.gob` files
- Easy to backup, version, share
- In-memory operations (fast) with periodic saves

### 2. Collection Naming Convention
```
{source}-{layer}
```
Examples:
- `scriptures-verse` - Book of Mormon, D&C, etc. at verse level
- `scriptures-paragraph` - Contextual chunks
- `scriptures-summary` - LLM-generated chapter summaries
- `conference-paragraph` - Conference talks
- `conference-citations` - Scripture references in talks

### 3. Document Metadata Schema
Every document includes:
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

### 4. Tool Design (MCP-ready)
Keep tools general for future MCP server:
```
gospel_search(query, layers?, sources?, limit?)
gospel_similar(reference, limit?)  # "verses like 1 Nephi 3:7"
gospel_citations(reference)        # "talks citing 1 Nephi 3:7"
gospel_index(source, layers?)      # Index content
```

---

## File Structure

```
scripts/gospel-vec/
├── TODO.md              # This file
├── go.mod
├── go.sum
├── .gitignore           # Ignore data/*.gob.gz
├── main.go              # CLI entry point
├── config.go            # Configuration
├── storage.go           # DB management (load/save/export)
├── embed.go             # Embedding functions
├── index.go             # Indexing pipeline
├── search.go            # Search functions
├── summary.go           # LLM summary generation
├── chunking.go          # Chunking strategies
├── types.go             # Shared types
└── data/                # Generated data (gitignored)
    ├── gospel-vec.gob.gz    # Main database
    └── summaries/           # Cached summaries (optional)
```

---

## Configuration

```go
type Config struct {
    // LM Studio
    EmbeddingURL   string // "http://localhost:1234/v1"
    EmbeddingModel string // "text-embedding-qwen3-embedding-4b"
    ChatURL        string // "http://localhost:1234/v1"
    ChatModel      string // "qwen3-vl-8b" (or similar)
    
    // Storage
    DataDir        string // "./data"
    DBFile         string // "gospel-vec.gob.gz"
    
    // Content paths
    ScripturesPath string // "../../gospel-library/eng/scriptures"
    ConferencePath string // "../../gospel-library/eng/general-conference"
}
```

---

## Next Steps

1. **Create project structure** - go.mod, directories, .gitignore
2. **Implement storage layer** - Load/save/export with compression
3. **Port chunking from chromem-exp** - Reuse working code
4. **Build indexing pipeline** - Multi-layer with metadata
5. **Add summary generation** - LM Studio chat integration
6. **Create search interface** - Unified multi-layer search
