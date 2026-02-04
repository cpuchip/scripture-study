# Tool Improvements for Scripture Study

*Ideas for new MCPs and improvements to accelerate and deepen our study*

---

## Priority: Most Impactful for Our Study Patterns

Based on our actual study sessions, these would have the highest ROI:

| Priority | Tool | Why |
|----------|------|-----|
| 1 | **Webster 1828 Dictionary** | Period-appropriate definitions for KJV/Restoration language |
| 2 | **Semantic Search** | Keyword search limits what we can discover |
| 3 | **Cross-Reference Graph** | Following chains manually is tedious |
| 4 | **Hebrew/Greek Word Study** | Original language depth (Bible only) |
| 5 | **Prophet Teaching Index** | We often study by speaker |

> **Note on Strong's Concordance:** Strong's only covers the Bible - there's no analogous resource for Book of Mormon, D&C, or Pearl of Great Price. However, the **Webster 1828 Dictionary** provides period-appropriate definitions for the KJV-style language used across ALL Restoration scriptures, making it more universally useful for our study.

---

## Part 1: Tools for Studying Deeper

### 1.0 Webster 1828 Dictionary MCP ⭐ NEW PRIORITY

**Problem:** English has shifted significantly in 200 years. Words like "suffer," "let," "peculiar," and "conversation" meant different things in 1828. The Book of Mormon, D&C, and Pearl of Great Price use KJV-style language from this era.

**Why Webster 1828 over Strong's:**
- **Strong's is Bible-only** - No analogous resource exists for BoM/D&C/PoGP
- **Webster 1828 covers ALL scriptures** - Same language era as the Restoration
- **Noah Webster was contemporary** - Published same decade as Book of Mormon
- **Embeddable locally** - Full dictionary available as JSON (~25MB)

**Proposed Features:**
- Look up any word for 1828 definition
- Show part of speech, synonyms, multiple definitions
- Flag words with significantly different modern meanings
- Integrate into gospel-mcp or standalone tool

**Data Sources (All Free/MIT Licensed):**
- **ssvivian/WebstersDictionary**: https://github.com/ssvivian/WebstersDictionary
  - Full dictionary as `dictionary.json` (~25MB)
  - Format: `{ "word": "...", "pos": "...", "synonyms": "...", "definitions": [...] }`
  - MIT License, Project Gutenberg source
- **CrossCrusaders/Websters1828API**: https://github.com/CrossCrusaders/Websters1828API
  - Individual JSON files per word in `src/` folder
  - MIT License, designed for KJV study
- **webstersdictionary1828.com**: Online reference for verification

**Example Query:**
```
"Define 'intelligence' in 1828 context"
→ Returns: 
  INTELLIGENCE, n. [L. intelligentia]
  1. Understanding; skill.
  2. Notice; information communicated.
  3. Commerce of acquaintance; terms of intercourse.
  4. A spiritual being...
```

**Implementation Plan:**
1. Download `dictionary.json` from ssvivian repo
2. Index in SQLite for fast lookup
3. Add `webster_define(word)` tool to gospel-mcp or new MCP
4. Consider fuzzy matching for spelling variants

**Example Words That Changed:**
| Word | 1828 Meaning | Modern Meaning |
|------|--------------|----------------|
| suffer | allow, permit | experience pain |
| let | hinder, prevent | allow |
| peculiar | special, belonging exclusively | strange, odd |
| conversation | conduct, behavior | spoken exchange |
| prevent | go before, precede | stop from happening |
| quick | living, alive | fast |

---

### 1.1 Hebrew/Greek Word Study MCP

**Problem:** When studying words like "intelligence," "agency," or "glory," we often want the original language meaning but have to leave our study environment.

**Proposed Features:**
- Look up Strong's Concordance entries by number or English word
- Show Hebrew/Greek roots and semantic ranges
- Find all occurrences of a word across scripture
- Show cognate words from the same root
- Link to BDB (Hebrew) and BDAG (Greek) lexicons

**Data Sources:**
- **Open Scriptures Strong's Dictionary**: https://github.com/openscriptures/strongs (public domain)
- **Complete Study Bible API**: https://rapidapi.com/teachjesusapp/api/complete-study-bible (Strong's Numbers, Greek & Hebrew, Lexicons)
- **IQ Bible API**: https://forallthings.bible/resource/iq-bible-api/ (Strong's + original texts)

**Example Query:**
```
"What Hebrew word underlies 'intelligence' in Abraham 3:22?"
→ Returns: Hebrew root, meaning, other verses using same word
```

**Implementation Notes:**
- Could download Strong's data and index locally in SQLite
- Would need to map scripture references to Strong's numbers
- Consider whether to call external API or build local index

---

### 1.2 Cross-Reference Graph Builder

**Problem:** Our gospel-mcp returns `related_references` but we don't fully exploit them. Manually following chains is tedious.

**Proposed Features:**
- Recursively expand cross-references to N levels
- Build a visual "reference graph" (could output mermaid diagram)
- Find scriptures sharing Topical Guide entries
- Identify "hubs" (heavily cross-referenced verses)
- Show citation paths between two scriptures

**Existing Resource:**
- **scriptures.byu.edu** - BYU's Scripture Citation Index already does this!
- Shows where scriptures are cited in conference talks, manuals, etc.
- May be possible to scrape or find API

**Build-Our-Own Option:**
- Our gospel-library markdown already contains cross-references
- Could build a graph database from the footnotes
- Use Neo4j, or simpler: SQLite with recursive CTEs

**Example Query:**
```
"Show me everything connected to 2 Nephi 2:27 within 2 hops"
→ Returns: Graph of related scriptures with connection types
```

---

### 1.3 Chiasmus/Literary Structure Detector

**Problem:** Hebrew poetry patterns (chiasmus, parallelism, inclusio) encode meaning in structure, but we miss them reading linearly.

**Proposed Features:**
- Detect and highlight chiastic patterns
- Show parallel structures (A-B, A'-B')
- Identify repeated phrases that bracket sections (inclusio)
- Visualize structure as nested or mirrored format

**Challenges:**
- Requires sophisticated text analysis
- May need manual curation for known chiasms
- Could start with a database of known structures

**Example Query:**
```
"Map the chiastic structure of Alma 36"
→ Returns: Visual diagram showing A-B-C-D-D'-C'-B'-A' pattern
```

---

### 1.4 Historical/Cultural Context MCP

**Problem:** Understanding "what was happening then" requires leaving the study environment.

**Proposed Features:**
- Timeline dating for biblical events
- What was happening in surrounding cultures (Egypt, Babylon, Rome)
- Geographic context (distances, terrain, significance)
- Who ruled, who prophesied when
- Map integration (or link to existing maps)

**Data Sources:**
- Our existing `gospel-library/eng/scriptures/bible-chronology/`
- Our existing `gospel-library/eng/scriptures/bible-maps/`
- Could enhance with additional historical data

**Example Query:**
```
"What was happening in Babylon when Lehi left Jerusalem (600 BC)?"
→ Returns: Nebuchadnezzar, Babylonian expansion, Jeremiah prophesying, etc.
```

---

## Part 2: Tools for Studying Wider

### 2.1 Prophet Teaching Index

**Problem:** We often want to know what a specific prophet has taught on a topic across all sources.

**Proposed Features:**
- Search conference talks by speaker name
- Include Teachings of Presidents manuals
- Cross-reference published writings
- Show topic trends over time for a speaker

**Current Capability:**
- Our gospel-mcp can search with `speaker:nelson` in conference talks
- But doesn't span all sources

**Enhancement:**
- Index all speaker attributions across sources
- Add parameter: `gospel_search(query="temples", speaker="nelson", source="all")`

**Example Query:**
```
"What has President Nelson taught about temples across all his talks?"
→ Returns: Chronological list with excerpts
```

---

### 2.2 Semantic Search (Vector-Based) ⭐ HIGH PRIORITY

**Problem:** Current FTS5 is keyword-based. If you don't know the exact words, you miss related content.

**Proposed Features:**
- Search by meaning/concept, not just keywords
- Find thematically similar verses across volumes
- "More like this" feature for any passage
- Cluster related scriptures automatically

**Example Query:**
```
"Find passages similar in meaning to 'be still and know that I am God'"
→ Returns: Related passages about trusting God, divine peace, etc.
```

#### Research: Vector Database Options

**Option A: chromem-go ⭐ NEW RECOMMENDATION**
- **URL:** https://github.com/philippgille/chromem-go
- **Stars:** 842 | **License:** MPL-2.0
- **Go Package:** `go get github.com/philippgille/chromem-go@latest`

**Why chromem-go fits our "fewer moving parts" philosophy:**

✅ **Zero third-party dependencies** - Pure Go, nothing else to install  
✅ **No CGO required** - Unlike sqlite-vec, compiles anywhere Go runs  
✅ **Built-in embedding providers** - Ollama, LocalAI, OpenAI, Cohere, Mistral, Jina  
✅ **In-memory with optional persistence** - gob files, gzip-compressed, AES-GCM encryption  
✅ **Embeddable** - Like SQLite, no separate database to maintain  
✅ **Multi-threaded** - Uses Go's concurrency for add/query operations  

**Benchmarks (i5-1135G7 laptop):**
| Docs | Query Time |
|------|-----------|
| 1,000 | 0.3ms |
| 5,000 | 2.1ms |
| 25,000 | 10ms |
| 100,000 | 40ms |

Our ~42k scriptures would query in ~15ms! 🚀

**Go Usage:**
```go
import (
    "context"
    "runtime"
    "github.com/philippgille/chromem-go"
)

func main() {
    ctx := context.Background()
    
    // In-memory DB (or NewPersistentDB("./data", true) for persistence)
    db := chromem.NewDB()
    
    // Create collection with Ollama embeddings (localhost:11434)
    embeddingFunc := chromem.NewEmbeddingFuncOllama("nomic-embed-text", "http://localhost:11434/api")
    c, _ := db.CreateCollection("scriptures", nil, embeddingFunc)
    
    // Add documents (embedding generated automatically)
    c.AddDocuments(ctx, []chromem.Document{
        {ID: "1ne-3-7", Content: "I will go and do the things which the Lord hath commanded..."},
        {ID: "dc-93-36", Content: "The glory of God is intelligence..."},
    }, runtime.NumCPU())
    
    // Semantic search
    results, _ := c.Query(ctx, "What is intelligence?", 5, nil, nil)
    for _, r := range results {
        fmt.Printf("%.2f: %s - %s\n", r.Similarity, r.ID, r.Content)
    }
}
```

**Built-in Embedding Providers:**
- **Hosted:** OpenAI, Azure OpenAI, Cohere, Mistral, Jina, mixedbread.ai, GCP Vertex
- **Local:** Ollama, LocalAI (works with LM Studio via OpenAI-compat!)
- **Custom:** Implement `chromem.EmbeddingFunc(ctx, text) ([]float32, error)`

**Storage Options:**
- `NewDB()` - In-memory only
- `NewPersistentDB(path, compress)` - Auto-saves each document as gob file
- `ExportToFile/ImportFromFile` - Full DB backup with optional encryption
- `ExportToWriter/ImportFromReader` - S3 or blob storage support

**Features:**
- Cosine similarity search (exhaustive nearest neighbor)
- Document filters: `$contains`, `$not_contains`
- Metadata filters: Exact matches
- Negative queries: Exclude certain results
- WASM binding available (experimental)

---

**Option B: sqlite-vec (CGO Required)**
- **URL:** https://github.com/asg017/sqlite-vec
- **Stars:** 6.8k | **Sponsors:** Mozilla Builders, Fly.io, Turso
- **Go Bindings:** `go get -u github.com/asg017/sqlite-vec-go-bindings/cgo`

**Pros:**
- Runs anywhere SQLite runs (our current stack!)
- Pure C, no dependencies, ~500KB
- Official Go bindings with two options:
  - CGO with `github.com/mattn/go-sqlite3`
  - WASM-based with `github.com/ncruces/go-sqlite3`
- Supports float32, int8, and binary vectors
- Metadata columns, partition keys, auxiliary data
- Mozilla-sponsored, actively maintained

**Cons:**
- **Requires CGO** - Build complexity, cross-compilation issues
- Brute-force KNN (fine for ~40k items, may slow at 500k+)
- Must generate embeddings separately (no built-in providers)

**Go Usage:**
```go
import (
    sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    sqlite_vec.Auto()
    db, _ := sql.Open("sqlite3", "gospel.db")
    
    // Create vector table
    db.Exec(`CREATE VIRTUAL TABLE scripture_vec USING vec0(
        reference TEXT PRIMARY KEY,
        embedding FLOAT[1536]
    )`)
    
    // Insert vector
    v, _ := sqlite_vec.SerializeFloat32([]float32{0.1, 0.2, ...})
    db.Exec("INSERT INTO scripture_vec(reference, embedding) VALUES (?, ?)", 
        "1 Nephi 3:7", v)
    
    // KNN search
    db.Query(`SELECT reference, distance FROM scripture_vec 
              WHERE embedding MATCH ? ORDER BY distance LIMIT 10`, queryVec)
}
```

---

**Option C: MongoDB Atlas Vector Search (Full-Featured)**
- **Docker Image:** `mongodb/mongodb-atlas-local`
- **URL:** https://hub.docker.com/r/mongodb/mongodb-atlas-local

**Pros:**
- Full Atlas Vector Search capabilities locally via Docker
- No cloud account needed for development
- Includes `mongot` search server for vector indexing
- Rich query language, aggregation pipelines
- Can combine vector + text + metadata filters
- Go driver: `go.mongodb.org/mongo-driver`
- Great if you want document-style storage

**Cons:**
- Heavier footprint (~600MB Docker image)
- Requires Docker Desktop running
- More complex than SQLite for simple use cases
- Two processes: `mongod` + `mongot`

**Docker Setup:**
```bash
# Pull and run
docker pull mongodb/mongodb-atlas-local
docker run -p 27017:27017 --name atlas-local mongodb/mongodb-atlas-local

# Wait for healthy
while [ "$(docker inspect -f {{.State.Health.Status}} atlas-local)" != "healthy" ]; do 
    sleep 2
done

# Connect
mongosh "mongodb://localhost/?directConnection=true"
```

**Vector Index Creation:**
```javascript
db.scriptures.createSearchIndex("vector_index", {
    "mappings": {
        "dynamic": true,
        "fields": {
            "embedding": {
                "type": "knnVector",
                "dimensions": 1536,
                "similarity": "cosine"
            }
        }
    }
});
```

---

**Option D: Qdrant (High-Performance, Purpose-Built)**
- **URL:** https://qdrant.tech/
- **Docker:** `docker pull qdrant/qdrant`
- **Go Client:** `github.com/qdrant/go-client`

**Pros:**
- Purpose-built for vector search (Rust, very fast)
- REST + gRPC APIs
- Supports filtering, payload storage
- Excellent documentation
- Scales to millions of vectors
- Go client is well-maintained

**Cons:**
- **Requires Docker** - Another service to run
- Overkill for ~40k vectors
- More operational complexity

---

**Option E: Other Embedded Options**

| Library | Language | Notes |
|---------|----------|-------|
| `kelindar/search` | Go | Pure Go, embeds llama.cpp for embeddings |
| `weaviate/weaviate` | Go | Can embed, but typically runs as service |
| `chroma-core/chroma` | Python | Simple but Python-only |

---

#### Embedding Generation Options

**Your Setup:** LM Studio 0.4 with NVIDIA 4090 (24GB VRAM) 🔥

**Option 1: LM Studio Local Embeddings (RECOMMENDED)**
- LM Studio exposes OpenAI-compatible API at `http://localhost:1234`
- Supports `/v1/embeddings` endpoint
- Can run embedding models like `nomic-embed-text` or `bge-large`
- **Free, private, fast with your 4090**

**LM Studio Setup:**
1. Download an embedding model (e.g., `nomic-ai/nomic-embed-text-v1.5-GGUF`)
2. Load model in LM Studio
3. Start local server (default: `http://localhost:1234`)
4. Use OpenAI-compatible client:

```go
import "github.com/sashabaranov/go-openai"

client := openai.NewClientWithConfig(openai.ClientConfig{
    BaseURL: "http://localhost:1234/v1",
})

resp, _ := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
    Input: []string{"And it came to pass..."},
    Model: "nomic-embed-text",
})
embedding := resp.Data[0].Embedding // []float32
```

**Option 2: Ollama (Alternative Local)**
- Similar to LM Studio, runs embedding models
- `ollama pull nomic-embed-text`
- API at `http://localhost:11434`

**Option 3: OpenAI API (Fallback)**
- Best quality embeddings (text-embedding-3-small)
- ~$0.02 per 1M tokens (~$0.80 for all scriptures)
- Requires internet, costs money

---

#### Recommendation: chromem-go + Ollama ⭐ UPDATED

**Why chromem-go is the BEST fit for our "fewer moving parts" philosophy:**

| Criterion | chromem-go | sqlite-vec | MongoDB | Qdrant |
|-----------|-----------|------------|---------|--------|
| **CGO Required** | ❌ No | ✅ Yes | ❌ No | ❌ No |
| **Docker Required** | ❌ No | ❌ No | ✅ Yes | ✅ Yes |
| **External Process** | ❌ No | ❌ No | ✅ Yes | ✅ Yes |
| **Built-in Embeddings** | ✅ Ollama/OpenAI | ❌ No | ❌ No | ❌ No |
| **File Persistence** | ✅ Built-in | ✅ SQLite | ✅ Files | ✅ Files |
| **Pure Go** | ✅ Yes | ❌ CGO | ❌ N/A | ❌ Rust |

**The winning combination:**
1. **chromem-go** - Zero dependencies, embeds directly in our Go binary
2. **Ollama** - Local embeddings with `nomic-embed-text` (or LM Studio)
3. **No CGO, No Docker, No external processes** - Just `go build` and run

**chromem-go has built-in Ollama support:**
```go
embeddingFunc := chromem.NewEmbeddingFuncOllama("nomic-embed-text", "http://localhost:11434/api")
```

This is significantly simpler than sqlite-vec which:
- Requires CGO for Go bindings
- Needs separate embedding generation code
- Has cross-compilation challenges

**Migration Path:**
```
Phase 1: chromem-go + Ollama (MVP) ← Start Here!
    ↓ (only if needed for 500k+ docs)
Phase 2: Consider sqlite-vec for larger scale
    ↓ (only if multi-service architecture needed)
Phase 3: Qdrant or MongoDB for production microservices
```

**Why we probably won't need Phase 2/3:**
- ~42k scriptures queries in ~15ms on modest hardware
- chromem-go author benchmarked 100k docs at 40ms
- Our 4090 makes embedding generation trivially fast
- File persistence with gzip compression is sufficient

---

### 2.3 Parallel Account Comparator

**Problem:** Seeing synoptic views requires manual work.

**Proposed Features:**
- Side-by-side comparison of parallel accounts
- Gospels synoptic view (Matt/Mark/Luke/John)
- Creation accounts (Genesis/Moses/Abraham)
- JST alongside KJV
- Highlight differences and additions

**Example Query:**
```
"Compare the Sermon on the Mount (Matt 5-7) with 3 Nephi 12-14"
→ Returns: Side-by-side with differences highlighted
```

---

### 2.4 Pattern Finder Across Dispensations

**Problem:** Tracing themes through scripture history is manual and tedious.

**Proposed Features:**
- Track a phrase/concept through OT → NT → BoM → D&C
- Find "first mentions" of concepts
- Show frequency over time
- Identify dispensational patterns

**Example Query:**
```
"Trace the word 'covenant' from Abraham through Malachi"
→ Returns: Timeline of occurrences with context
```

---

## Part 3: Improvements to Existing gospel-mcp

### 3.1 Follow the Chain
When a search returns `related_references`, let us expand immediately:
```
gospel_expand(reference="2 Nephi 2:27", relation_index=3)
→ Retrieves the 3rd related reference without new search
```

### 3.2 Context Control
```
gospel_get(reference="D&C 93:29", context_verses=5)
→ Returns verse with 5 verses before and after
```

### 3.3 Footnote Extraction
```
gospel_get(reference="2 Nephi 2:27", include_footnotes=true)
→ Returns verse text plus all footnote content
```

### 3.4 Study Session Memory (Complex)
- Track what we've searched/read this session
- Remember open questions
- Suggest related content based on session theme

---

## Part 3.5: Search API Alternatives ⚠️ URGENT

### Problem: DuckDuckGo Rate Limiting

Our current `search-mcp` uses the duckduckgo-go library which performs web scraping. We've hit rate limits during research (CAPTCHA challenges), which breaks the tool during intensive sessions.

### Options Comparison

| Provider | Free Tier | Paid Pricing | Notes |
|----------|-----------|--------------|-------|
| **DuckDuckGo** | Unofficial scraping | N/A | Rate limited, unreliable |
| **Brave Search** | 2,000/month | $5/1k requests | Official API, good privacy |
| **Bing Web Search** | 1,000/month | $5/1k | Microsoft Azure, reliable |
| **SerpAPI** | 100/month | $50/5k | Multi-engine, expensive |
| **Google Custom Search** | 100/day | $5/1k | Limited to custom search engines |

### Recommendation: Brave Search API ⭐

**Why Brave:**
1. **Generous free tier** - 2,000 requests/month (vs 1,000 Bing)
2. **Privacy-focused** - Aligns with our values
3. **Clean API** - Simple REST endpoints, JSON responses
4. **Affordable scaling** - $5/1k after free tier
5. **Independent index** - Not reselling Google/Bing

**API Details:**
- Endpoint: `https://api.search.brave.com/res/v1/web/search`
- Auth: `X-Subscription-Token` header
- Parameters: `q`, `count` (1-20), `result_filter`, `freshness`

**Go Implementation:**
```go
type BraveClient struct {
    apiKey string
    http   *http.Client
}

func (c *BraveClient) Search(query string, count int) (*BraveResults, error) {
    req, _ := http.NewRequest("GET", "https://api.search.brave.com/res/v1/web/search", nil)
    req.Header.Set("X-Subscription-Token", c.apiKey)
    req.Header.Set("Accept", "application/json")
    
    q := req.URL.Query()
    q.Set("q", query)
    q.Set("count", strconv.Itoa(count))
    req.URL.RawQuery = q.Encode()
    
    resp, err := c.http.Do(req)
    // ... parse JSON response
}
```

### Migration Strategy

**Option A: Replace DuckDuckGo entirely**
- Simpler implementation
- Consistent behavior
- Uses free tier quota

**Option B: Brave primary, DuckDuckGo fallback**
- Brave for normal use (2k/month)
- DuckDuckGo when Brave quota depleted
- More complex but infinite capacity

**Option C: Multi-provider with rotation**
- Use all free tiers (Brave 2k + Bing 1k + DDG)
- Rotate based on availability
- Most robust but complex

**Recommended:** Start with **Option A** (Brave only). Our typical usage won't exceed 2k/month during development. Can add DuckDuckGo fallback later if needed.

### Action Items
1. Sign up for Brave Search API (free)
2. Refactor `search-mcp` to use Brave
3. Keep DuckDuckGo code as backup
4. Add rate limit tracking

---

## Part 4: Implementation Roadmap

### Updated Phase Order (February 2026)

| Phase | Project | Complexity | Why This Order |
|-------|---------|------------|----------------|
| 1 | Webster 1828 Dictionary MCP | Medium | Highest impact for daily study |
| 2 | Search API Migration (Brave) | Low | Fixes rate limiting, quick win |
| 3 | gospel-mcp Enhancements | Low-Medium | Incremental improvements |
| 4 | Semantic Search MVP | High | Major feature, needs setup |
| 5 | Cross-Reference Graph | Medium | Data already exists in footnotes |
| 6 | Hebrew/Greek Word Study | Medium | Lower priority (Webster covers most needs) |

### Phase 1: Webster 1828 Dictionary MCP ⭐ NEW TOP PRIORITY
1. Set up project structure (follow search-mcp pattern)
2. Download dictionary.json from ssvivian/WebstersDictionary
3. Implement dictionary loading and word lookup
4. Create MCP tools: `webster_define`, `webster_search`, `webster_related`
5. Test with Restoration scripture vocabulary

### Phase 2: Search API Migration
1. Sign up for Brave Search API (free tier: 2,000/month)
2. Refactor search-mcp to use Brave API
3. Keep DuckDuckGo as fallback option
4. Add rate limit tracking

### Phase 3: Quick Wins (Enhance gospel-mcp)
1. Add `context` parameter (verses before/after)
2. Add `include_footnotes` parameter  
3. Add speaker filter to search
4. Add `gospel_expand` for quick reference following

### Phase 4: Semantic Search MVP (chromem-go + Ollama)
1. Install Ollama and pull embedding model: `ollama pull nomic-embed-text`
2. Add chromem-go to gospel-mcp: `go get github.com/philippgille/chromem-go@latest`
3. Generate embeddings for all scriptures (~42k verses) - chromem-go auto-generates!
4. Add `gospel_semantic_search(query, limit)` tool
5. Add `gospel_similar(reference, limit)` tool
6. Test on real study sessions

### Phase 5: Cross-Reference Graph
1. Parse all footnote cross-references from markdown
2. Build graph structure in SQLite
3. Add `gospel_references(reference)` tool
4. Add `gospel_reference_graph(reference, depth)` tool
5. Consider Mermaid visualization output

### Phase 6: Hebrew/Greek Word Study (Lower Priority)
1. Research data sources (Strong's Concordance, STEP Bible)
2. Evaluate: completeness, license, cost
3. If viable: Implement word study tools
4. Note: Webster 1828 covers most study vocabulary needs

**Detailed tasks for each phase:** See [05_tool_todo.md](05_tool_todo.md)

---

## Technical Notes

### Current Stack
- Go for MCPs
- SQLite with FTS5 for full-text search
- Markdown files in gospel-library/
- MCP protocol for tool exposure
- **NEW:** Ollama for local embeddings (4090 GPU available)
- **NEW:** chromem-go for vector search (zero dependencies, pure Go)

### Adding chromem-go (Recommended)
```go
import (
    "context"
    "runtime"
    "github.com/philippgille/chromem-go"
)

// Create persistent DB (gzip compressed)
db, _ := chromem.NewPersistentDB("./gospel-vectors", true)

// Use Ollama for embeddings (must be running: ollama serve)
embeddingFunc := chromem.NewEmbeddingFuncOllama("nomic-embed-text", "")

// Create collection
c, _ := db.CreateCollection("scriptures", nil, embeddingFunc)

// Add documents (embeddings generated automatically!)
docs := []chromem.Document{
    {ID: "1-ne-3-7", Content: "I will go and do...", Metadata: map[string]string{"book": "1 Nephi"}},
    // ... more docs
}
c.AddDocuments(context.Background(), docs, runtime.NumCPU())

// Semantic search
results, _ := c.Query(context.Background(), "What is intelligence?", 10, nil, nil)
for _, r := range results {
    fmt.Printf("%s (%.3f): %s\n", r.ID, r.Similarity, r.Content[:50])
}
```

### Adding sqlite-vec (Alternative - Requires CGO)
```go
import (
    sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
    _ "github.com/mattn/go-sqlite3"
)

func init() {
    sqlite_vec.Auto() // Auto-register extension
}

// Create virtual table
CREATE VIRTUAL TABLE scripture_vec USING vec0(
    reference TEXT PRIMARY KEY,
    embedding FLOAT[768]  -- nomic-embed-text size
);

// Insert (using SerializeFloat32 helper)
v, _ := sqlite_vec.SerializeFloat32(embedding)
db.Exec("INSERT INTO scripture_vec VALUES (?, ?)", ref, v)

// Query
SELECT reference, distance 
FROM scripture_vec 
WHERE embedding MATCH ?
ORDER BY distance
LIMIT 10;
```

### Embedding Pipeline (Updated for chromem-go + Ollama)
```
1. Install Ollama and pull nomic-embed-text model
   $ ollama pull nomic-embed-text

2. Extract text from all scriptures (~42k verses)

3. Use chromem-go with built-in Ollama support:
   embeddingFunc := chromem.NewEmbeddingFuncOllama("nomic-embed-text", "")
   c, _ := db.CreateCollection("scriptures", nil, embeddingFunc)
   c.AddDocuments(ctx, docs, runtime.NumCPU())  // Auto-generates embeddings!

4. Query semantically:
   results, _ := c.Query(ctx, "What is intelligence?", 10, nil, nil)

5. Persist to disk:
   db, _ := chromem.NewPersistentDB("./gospel-vectors", true)  // gzip compressed
```

---

## Resources

### Webster 1828 Dictionary
- ssvivian/WebstersDictionary: https://github.com/ssvivian/WebstersDictionary ⭐
  - Contains `dictionary.json` (~25MB, all words)
  - MIT License, suitable for our use
- CrossCrusaders/Websters1828API: https://github.com/CrossCrusaders/Websters1828API
  - Individual JSON files per word
  - Alternative source if needed

### Vector Search
- **chromem-go:** https://github.com/philippgille/chromem-go ⭐⭐ RECOMMENDED
  - Go package: `go get github.com/philippgille/chromem-go@latest`
  - Zero dependencies, no CGO, pure Go
  - Built-in Ollama, OpenAI, Cohere, Mistral embedding support
  - MPL-2.0 License, 842 stars
- sqlite-vec: https://github.com/asg017/sqlite-vec
  - Go bindings: `github.com/asg017/sqlite-vec-go-bindings/cgo`
  - Mozilla-sponsored, 6.8k stars (requires CGO)
- MongoDB Atlas Local: https://hub.docker.com/r/mongodb/mongodb-atlas-local
  - Docker-based, full vector search
  - `docker run -p 27017:27017 mongodb/mongodb-atlas-local`
- Qdrant: https://qdrant.tech/
- Weaviate: https://weaviate.io/

### Local AI / Embeddings
- **Ollama:** https://ollama.com/ ⭐ RECOMMENDED for chromem-go
  - Native chromem-go support via `NewEmbeddingFuncOllama()`
  - `ollama pull nomic-embed-text`
  - API at `http://localhost:11434`
- LM Studio: https://lmstudio.ai/
  - OpenAI-compatible API at localhost:1234
  - Supports embedding models (nomic-embed-text, bge)
  - Your 4090 makes this blazing fast

### Search APIs
- Brave Search API: https://brave.com/search/api/ ⭐
  - Free tier: 2,000 requests/month
  - Privacy-focused, independent index
- Bing Web Search: Azure service, 1k free/month
- DuckDuckGo: Unofficial scraping (rate limited)

### Hebrew/Greek (Lower Priority)
- Open Scriptures Strong's: https://github.com/openscriptures/strongs
- STEP Bible API: https://stepbibleguide.blogspot.com/
- Complete Study Bible API: https://rapidapi.com/teachjesusapp/api/complete-study-bible

### Cross-References
- scriptures.byu.edu - BYU Scripture Citation Index
- Our own footnote data in gospel-library/

---

## Next Steps

**Recommended Implementation Order:**

1. **Webster 1828 Dictionary MCP** - Highest impact for daily study, covers all Restoration scriptures
2. **Search API Migration (Brave)** - Quick fix for rate limiting issue
3. **gospel-mcp Enhancements** - Low-hanging fruit improvements
4. **Semantic Search** - Major feature once fundamentals are solid

**See [05_tool_todo.md](05_tool_todo.md) for detailed task breakdown.**

---

*Document created: February 3, 2026*  
*Last updated: February 4, 2026 - Added chromem-go as recommended vector database (zero dependencies, pure Go)*
