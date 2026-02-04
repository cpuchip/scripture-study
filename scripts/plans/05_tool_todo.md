# Tool Improvement TODOs

Actionable tasks extracted from [04_tool-improvements.md](04_tool-improvements.md).

*Created: February 3, 2026*

---

## Overview

| Priority | Project | Complexity | Status |
|----------|---------|------------|--------|
| 🔴 1 | Webster 1828 Dictionary MCP | Medium | Not Started |
| 🔴 2 | Search API Migration (Brave) | Low | Not Started |
| 🟡 3 | gospel-mcp Enhancements | Low-Medium | Not Started |
| 🟡 4 | Semantic Search MVP (chromem-go) | Medium | Not Started |
| 🟢 5 | Cross-Reference Graph | Medium | Not Started |
| 🟢 6 | Hebrew/Greek Word Study | Medium | Not Started |

**Key Decision (2026-02-04):** Using **chromem-go** instead of sqlite-vec for semantic search.
- No CGO required (pure Go)
- Built-in Ollama/LM Studio embedding support
- Target `@master` branch for latest features

---

## 🔴 Priority 1: Webster 1828 Dictionary MCP

### Overview
Create an MCP server for Noah Webster's 1828 American Dictionary. This provides period-appropriate definitions for KJV-style language used across ALL Restoration scriptures.

### Tasks

#### 1.1 Project Setup
- [ ] Create `scripts/webster-mcp/` directory
- [ ] Initialize Go module: `go mod init github.com/stuffleberry/webster-mcp`
- [ ] Create directory structure:
  ```
  webster-mcp/
  ├── cmd/webster-mcp/
  │   ├── main.go
  │   └── serve.go
  ├── internal/
  │   ├── dictionary/
  │   │   ├── loader.go
  │   │   └── search.go
  │   └── mcp/
  │       └── server.go
  └── data/
      └── dictionary.json (downloaded)
  ```

#### 1.2 Data Acquisition
- [ ] Clone ssvivian/WebstersDictionary repository
- [ ] Locate `dictionary.json` file (~25MB)
- [ ] Verify JSON structure:
  ```json
  {
    "word": "ABASE",
    "pos": "v.t.",
    "synonyms": ["...", "..."],
    "definitions": ["...", "..."]
  }
  ```
- [ ] Copy/download to `webster-mcp/data/`
- [ ] Test parsing with Go JSON decoder

#### 1.3 Core Implementation
- [ ] Create `dictionary.Entry` struct matching JSON schema
- [ ] Implement `LoadDictionary(path string) (map[string]Entry, error)`
- [ ] Build FTS5 index for definitions (optional, for full-text search)
- [ ] Implement `Lookup(word string) (Entry, error)` with normalization:
  - Case-insensitive
  - Handle plurals
  - Stem common suffixes (-ed, -ing, -ly)

#### 1.4 MCP Tools
- [ ] Implement `webster_define(word)` - Returns full definition entry
- [ ] Implement `webster_search(query)` - Searches definitions for a phrase
- [ ] Implement `webster_related(word)` - Returns synonyms and related words
- [ ] Follow MCP protocol patterns from search-mcp

#### 1.5 Testing & Integration
- [ ] Write unit tests for dictionary loading
- [ ] Write unit tests for word lookup
- [ ] Test with common Restoration scripture words:
  - "charity"
  - "nigh"
  - "verily"
  - "wax" (as in "wax strong")
  - "vouchsafe"
- [ ] Add to VS Code MCP settings
- [ ] Test in actual scripture study session

#### 1.6 Documentation
- [ ] Update README with usage
- [ ] Document tool signatures
- [ ] Add example queries

---

## 🔴 Priority 2: Search API Migration (Brave)

### Overview
Replace DuckDuckGo web scraping with official Brave Search API to avoid rate limiting issues.

### Tasks

#### 2.1 API Setup
- [ ] Sign up at https://brave.com/search/api/
- [ ] Get free tier API key (2,000 requests/month)
- [ ] Store API key securely (environment variable or config)
- [ ] Test API with curl:
  ```bash
  curl -X GET "https://api.search.brave.com/res/v1/web/search?q=test" \
       -H "X-Subscription-Token: YOUR_API_KEY"
  ```

#### 2.2 Implementation
- [ ] Create `internal/brave/client.go`:
  - [ ] Define response structs from Brave API docs
  - [ ] Implement `Search(query, count)` method
  - [ ] Handle API errors and rate limits
  - [ ] Add retry logic with exponential backoff
- [ ] Update `internal/mcp/server.go`:
  - [ ] Replace DDG calls with Brave
  - [ ] Keep same tool interface (backward compatible)
- [ ] Add configuration option for API key

#### 2.3 Fallback Strategy
- [ ] Keep DuckDuckGo code in `internal/ddg/`
- [ ] Add config flag: `SEARCH_PROVIDER=brave|ddg|auto`
- [ ] Implement auto-switching if Brave quota exceeded

#### 2.4 Testing
- [ ] Test normal search queries
- [ ] Test edge cases (empty query, special characters)
- [ ] Test rate limit handling
- [ ] Verify output format matches previous DDG output

---

## 🟡 Priority 3: gospel-mcp Enhancements

### Overview
Quick wins to improve existing gospel-mcp functionality.

### 3.1 Context Control
- [ ] Add `context` parameter to `gospel_get` tool
- [ ] Modify retrieval to include N verses before/after
- [ ] Format output clearly showing context vs target verse
- [ ] Default to 0 (current behavior), max 10

**Tool Signature:**
```json
{
  "name": "gospel_get",
  "parameters": {
    "reference": { "type": "string" },
    "context": { "type": "integer", "default": 0, "maximum": 10 }
  }
}
```

### 3.2 Footnote Extraction
- [ ] Parse footnote anchors from markdown (e.g., `<sup>[1a](#fn-1a)</sup>`)
- [ ] Extract footnote content from bottom of chapter files
- [ ] Add `include_footnotes` parameter to `gospel_get`
- [ ] Format footnotes in output (inline or appended)

**Example Output:**
```
"And I, Nephi, said unto my father: I will go and do..."¹ᵃ
---
Footnotes:
1a. will - 1 Sam. 17:32; TG Faith; Loyalty; Obedience
```

### 3.3 Speaker Filter
- [ ] Audit markdown files for speaker attribution patterns
- [ ] Build speaker index (name → talk paths)
- [ ] Add `speaker` parameter to `gospel_search`
- [ ] Handle name variations (Nelson/Russell M. Nelson/President Nelson)

### 3.4 Expand References Helper
- [ ] When search returns related_references, allow quick expand
- [ ] Add `gospel_expand(reference, index)` tool
- [ ] Returns full content of Nth related reference without new search

---

## 🟡 Priority 4: Semantic Search MVP (chromem-go + Ollama/LM Studio)

### Overview
Add vector-based semantic search using **chromem-go** (pure Go, zero dependencies, no CGO) with local embeddings via Ollama or LM Studio.

**Why chromem-go over sqlite-vec?**
- ✅ No CGO required (sqlite-vec needs CGO bindings)
- ✅ No external processes (unlike MongoDB/Qdrant)
- ✅ Built-in embedding providers (Ollama native, OpenAI-compatible for LM Studio)
- ✅ Pure Go - embeds directly in gospel-mcp
- ✅ File persistence with gzip compression

### 4.1 Environment Setup

**Option A: Ollama (Recommended - simplest)**
- [ ] Install Ollama: https://ollama.com/
- [ ] Pull embedding model: `ollama pull nomic-embed-text`
- [ ] Verify Ollama running: `ollama serve` (usually auto-starts)
- [ ] Test embedding:
  ```bash
  curl http://localhost:11434/api/embeddings \
       -d '{"model": "nomic-embed-text", "prompt": "test text"}'
  ```

**Option B: LM Studio (if you prefer the GUI)**
- [ ] Download embedding model in LM Studio:
  - Option: `nomic-ai/nomic-embed-text-v1.5-GGUF`
  - Option: `BAAI/bge-large-en-v1.5-GGUF`
- [ ] Start LM Studio local server (port 1234)
- [ ] Test embedding endpoint (OpenAI-compatible):
  ```bash
  curl http://localhost:1234/v1/embeddings \
       -H "Content-Type: application/json" \
       -d '{"input": "test text", "model": "nomic-embed-text"}'
  ```
- **Note:** LM Studio uses OpenAI API format, not Ollama API. Use `NewEmbeddingFuncOpenAI()` with base URL `http://localhost:1234/v1`

### 4.2 chromem-go Integration
- [ ] Add dependency from master branch (more recent than v0.7.0 release):
  ```bash
  go get github.com/philippgille/chromem-go@master
  ```
- [ ] Create embedding function:
  ```go
  // Option A: Ollama (native support)
  embeddingFunc := chromem.NewEmbeddingFuncOllama("nomic-embed-text", "")
  
  // Option B: LM Studio (OpenAI-compatible)
  embeddingFunc := chromem.NewEmbeddingFuncOpenAI(
      "",  // API key not needed for local
      chromem.EmbeddingModelOpenAI("nomic-embed-text"),
      chromem.WithBaseURL("http://localhost:1234/v1"),
  )
  ```
- [ ] Create persistent DB:
  ```go
  db, _ := chromem.NewPersistentDB("./gospel-vectors", true)  // gzip compressed
  collection, _ := db.CreateCollection("scriptures", nil, embeddingFunc)
  ```

### 4.3 Embedding Generation Pipeline
- [ ] Create document loader from scripture markdown files
- [ ] Process all scriptures (~41,995 verses):
  - Book of Mormon: ~6,604 verses
  - D&C: ~3,654 verses  
  - Pearl of Great Price: ~598 verses
  - Old Testament: ~23,145 verses
  - New Testament: ~7,957 verses
- [ ] Add documents with concurrent embedding:
  ```go
  docs := []chromem.Document{
      {ID: "1-ne-3-7", Content: "I will go and do...", Metadata: map[string]string{"book": "1 Nephi"}},
      // ... more docs
  }
  collection.AddDocuments(ctx, docs, runtime.NumCPU())  // Auto-generates embeddings!
  ```
- [ ] Estimated time: ~42k docs with RTX 4090 should be fast (~15 mins)
- [ ] Add progress tracking/resume capability

### 4.4 MCP Tool Implementation
- [ ] Create `gospel_semantic_search(query, limit)`:
  ```go
  results, _ := collection.Query(ctx, query, limit, nil, nil)
  // Returns []Result with ID, Content, Similarity, Metadata
  ```
- [ ] Add "more like this" tool: `gospel_similar(reference, limit)`
  - Get content for reference, then query with that content
- [ ] Format results with snippets and similarity scores

### 4.5 Testing & Evaluation
- [ ] Test with known conceptual queries:
  - "God's love for His children"
  - "Trusting in the Lord during trials"
  - "Preparing for the Second Coming"
- [ ] Compare to FTS5 results for same concepts
- [ ] Tune result count and distance thresholds

---

## 🟢 Priority 5: Cross-Reference Graph

### Overview
Build a navigable graph of scripture cross-references.

### 5.1 Reference Parsing
- [ ] Parse all footnote references from markdown files
- [ ] Extract source → target mappings
- [ ] Handle reference formats:
  - Standard: `[John 3:16](../../nt/john/3.md)`
  - Topical Guide: `[TG Faith](../../tg/faith.md)`
  - Bible Dictionary: `[BD Atonement](../../bd/atonement.md)`
- [ ] Store in SQLite (source_ref, target_ref, footnote_id)

### 5.2 Graph Queries
- [ ] Build adjacency list or use SQLite recursive CTE
- [ ] Implement `get_references(ref)` - direct links
- [ ] Implement `get_references(ref, depth=2)` - multi-hop
- [ ] Calculate reference "importance" (PageRank-style)

### 5.3 MCP Tools
- [ ] `gospel_references(reference)` - list all cross-refs for a verse
- [ ] `gospel_reference_graph(reference, depth)` - expand outward
- [ ] Optional: Mermaid diagram output for visualization

---

## 🟢 Priority 6: Hebrew/Greek Word Study

### Overview
Add original language word study for Bible verses. Lower priority since Webster 1828 covers most study needs.

### 6.1 Data Sources (Research Needed)
- [ ] Evaluate: Open Scriptures Strong's Concordance
- [ ] Evaluate: Complete Study Bible API (RapidAPI)
- [ ] Evaluate: STEP Bible API
- [ ] Choose source based on:
  - Completeness
  - License
  - Ease of integration
  - Cost (prefer free)

### 6.2 Implementation (TBD based on source)
- [ ] Download/integrate concordance data
- [ ] Map Strong's numbers to verse references
- [ ] Build lookup index
- [ ] Create MCP tools:
  - `word_study(word)` - Hebrew/Greek definition
  - `word_occurrences(strong_number)` - all uses in Bible

---

## Notes

### Development Approach
1. **One project at a time** - Complete and test before starting next
2. **Reuse patterns** - Follow search-mcp structure for consistency
3. **Test thoroughly** - Each tool should have unit tests + integration tests
4. **Document as you go** - Update READMEs with each feature

### Dependencies
- Webster MCP: None (independent)
- Brave Search: None (can replace DDG in search-mcp)
- gospel-mcp enhancements: Existing gospel-mcp
- Semantic Search: Ollama OR LM Studio running + chromem-go (pure Go, no CGO!)
- Cross-Reference Graph: Parsed footnote data
- Hebrew/Greek: External API or data source

### Hardware Requirements
- **Webster MCP:** Minimal (~50MB RAM for dictionary)
- **Semantic Search:** 
  - Ollama with nomic-embed-text (~2GB VRAM) OR LM Studio (~4GB VRAM)
  - chromem-go database (~200MB for ~42k embeddings, gzip compressed)
  - Your RTX 4090 will make embedding generation blazing fast
- **Others:** Minimal

---

## Progress Log

| Date | Task | Status | Notes |
|------|------|--------|-------|
| 2026-02-03 | Created TODO document | ✅ Done | Extracted from 04_tool-improvements.md |
| 2026-02-04 | Updated Semantic Search to chromem-go | ✅ Done | Replaced sqlite-vec (no CGO needed!) |
| | | | |

---

*Last updated: February 4, 2026*
