# Storage Issues & Multi-File DB Plan

## Current State

### The Numbers
| Metric | Value |
|--------|-------|
| DB file on disk | 2.47 GB (gob.gz compressed) |
| Process memory (idle) | ~11-13 GB |
| Process memory (saving) | ~15 GB |
| Save speed | ~10 MB/s effective |
| Save duration | ~3-4 minutes for full DB |
| Total documents | 210,011 across 8 collections |
| Index-all time | ~1h38m for talks alone, ~3h total |

### Current Collections (all in one `gospel-vec.gob.gz`)
```
scriptures-verse       (all volumes: OT, NT, BoM, D&C, PGP)
scriptures-paragraph
scriptures-summary
scriptures-theme
conference-paragraph   (all years: 1971-2025, ~3977 talks, ~128K chunks)
conference-summary
manual-paragraph       (Lectures on Faith, Teachings of Presidents, CFM, etc.)
manual-summary
```

## Root Cause Analysis

### Why is save so slow?

The bottleneck is **NOT disk I/O** — it's gob serialization + gzip compression:

1. **`encoding/gob`** serializes the entire in-memory DB into a stream using
   reflection. Reflection is slow, and gob walks every single field of every
   document in the entire map-of-maps structure. Single-threaded.

2. **`compress/gzip`** (standard library) compresses the gob stream. Also
   single-threaded.

3. **Pipeline**: `chromem-go.ExportToFile()` → `persistToWriter()` chains:
   ```
   gob.Encoder → gzip.Writer → os.File
   ```
   The gob encoder feeds the gzip writer which feeds the file. Since gob is
   the slowest link, the gzip writer starves for data, and the file writer
   starves even more. That's why Task Manager shows ~0% CPU (single thread
   on a multi-core machine), ~0% disk (waiting on encoder), and only ~10 MB/s
   effective write throughput.

4. **Memory spike during save**: The persistence code builds a
   `persistenceDB` struct that copies all collection/document pointers into
   a new map, then gob encodes it. While the encoder is running, both the
   live DB and the encoding buffer are in memory → ~15 GB peak.

### Why is the file so large?

Each chunk stores:
- **Embedding vector**: `[]float32` — the biggest piece per document
- **Content**: the text (100-2000 bytes typically)
- **Metadata**: `map[string]string` with ~10-15 keys per document
- **gob overhead**: type descriptors, map headers, string lengths, etc.

With 128K+ documents, even modest per-doc size adds up fast. The embeddings
alone are likely 500MB-1GB+ uncompressed (depends on model dimension).

## Proposed Solution: Multi-File Storage

### Key Insight from chromem-go API

`ExportToFile` and `ImportFromFile` already support **per-collection export**:

```go
// Export only specific collections
db.ExportToFile("scriptures.gob.gz", true, "", "scriptures-verse", "scriptures-paragraph", ...)

// Import only specific collections
db.ImportFromFile("conference.gob.gz", "", "conference-paragraph", "conference-summary")
```

This means we can split without forking chromem-go.

### Phase 1: Split by Source (Quick Win)

Split the single file into **3 files by source**:

```
data/
├── scriptures.gob.gz    (~400-600 MB?)
├── conference.gob.gz    (~1.5 GB — the biggest piece)
├── manual.gob.gz        (~50-200 MB?)
└── gospel-vec.lock
```

**Benefits:**
- Each file is smaller → faster save, lower memory spike
- Indexers only need to load/save their own source file
- `index` only touches `scriptures.gob.gz`
- `index-talks` only touches `conference.gob.gz`
- `index-manuals` only touches `manual.gob.gz`
- MCP server loads all 3 at startup (still same total memory, but faster
  parallel load possible)

**Effort:** Medium — change `Store` to manage multiple DB instances or use
collection-filtered export/import.

### Phase 2: Finer Splits (Medium Effort)

Split conference talks (the biggest source) by **decade**:

```
data/
├── scriptures.gob.gz
├── conference-1970s.gob.gz
├── conference-1980s.gob.gz
├── conference-1990s.gob.gz
├── conference-2000s.gob.gz
├── conference-2010s.gob.gz
├── conference-2020s.gob.gz
├── manual.gob.gz
└── gospel-vec.lock
```

**Problem:** This requires changing collection naming from `conference-paragraph`
to `conference-1970s-paragraph`, etc. Because chromem-go identifies documents
by collection, not by metadata. All 128K talk chunks are currently in just 2
collections (`conference-paragraph` and `conference-summary`).

**Alternative — split by collection, not by content:**
```
data/
├── scriptures-verse.gob.gz
├── scriptures-paragraph.gob.gz
├── scriptures-summary.gob.gz
├── scriptures-theme.gob.gz
├── conference-paragraph.gob.gz
├── conference-summary.gob.gz
├── manual-paragraph.gob.gz
├── manual-summary.gob.gz
└── gospel-vec.lock
```

One file per collection. Simple, uses the existing API directly:
```go
db.ExportToFile("data/conference-paragraph.gob.gz", true, "", "conference-paragraph")
```

**Benefits:**
- Each file is independently loadable
- Indexers save only the collections they modified
- Could load collections in parallel at startup
- Natural boundary — each collection is one file

**Effort:** Medium — need to change Save/Load to iterate over collections.

### Phase 3: Volume-Level Splits (Bigger Refactor)

Change collection naming to include volume/decade:

```
scriptures-bofm-verse
scriptures-bofm-paragraph
scriptures-ot-verse
scriptures-ot-paragraph
conference-2020s-paragraph
…
```

**Benefits:**
- Maximum granularity — only load what you need
- Could do lazy loading for MCP server (load on first search)
- Indexing one volume doesn't require loading others

**Costs:**
- Breaking change — existing data needs migration
- More collection names → more complex iteration in Search
- Lots of small files vs. a few medium ones

### Phase 4: PersistentDB Mode (Alternative Path)

chromem-go has `NewPersistentDB(path, compress)` which writes **one file
per document** instead of one big gob file:

```
data/persistent/
├── <hash(collection-name)>/
│   ├── 00000000.gob.gz          (collection metadata)
│   ├── <hash(doc-id)>.gob.gz    (document 1)
│   ├── <hash(doc-id)>.gob.gz    (document 2)
│   └── ...
└── ...
```

**Benefits:**
- No full-DB serialization at all — documents persist on add
- Adding one document doesn't rewrite the whole DB
- Incremental — perfect for indexing
- No memory spike during save

**Costs:**
- 128K+ tiny files on NTFS (Windows) — could be slow for dir listing, antivirus scanning
- Loading 128K files at startup might actually be slower than reading one 2.3GB file
- More disk space (no cross-document gob/gzip optimization)
- Different code path — currently using in-memory DB + export

**Hybrid approach:** Use PersistentDB during indexing (incremental writes),
then export to gob.gz files for the MCP server (fast single-file load).

## Recommendation

### Immediate (Phase 1.5): Per-Collection Files

The sweet spot is **one file per collection** (Phase 2 alternative). It:

1. Uses the existing chromem-go API (`ExportToFile` with collection filter)
2. Doesn't require renaming collections (no breaking change)
3. Reduces save time proportionally — only save modified collections
4. Enables parallel loading at MCP startup
5. Is a clean foundation for future finer splits

### Implementation Sketch

#### New Store Design

```go
type MultiStore struct {
    db      *chromem.DB       // Single in-memory DB (all collections)
    config  *Config
    embed   chromem.EmbeddingFunc
    mu      sync.RWMutex
}

// Load reads all collection files from the data directory
func (s *MultiStore) Load() error {
    files, _ := filepath.Glob(filepath.Join(s.config.DataDir, "*.gob.gz"))
    for _, f := range files {
        // ImportFromFile merges into the existing DB
        s.db.ImportFromFile(f, "")
    }
    return nil
}

// Save exports each collection to its own file
func (s *MultiStore) Save() error {
    for _, source := range []Source{SourceScriptures, SourceConference, SourceManual} {
        for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
            name := collectionName(source, layer)
            col := s.db.GetCollection(name, nil)
            if col == nil || col.Count() == 0 {
                continue
            }
            file := filepath.Join(s.config.DataDir, name+".gob.gz")
            tmpFile := file + ".tmp"
            s.db.ExportToFile(tmpFile, true, "", name)
            os.Rename(tmpFile, file)
        }
    }
    return nil
}

// SaveSource saves only collections for a specific source
func (s *MultiStore) SaveSource(source Source) error {
    // Only save the 2-4 collections belonging to this source
    // Much faster than saving everything
}
```

#### Indexer Changes

```go
// cmdIndex saves only scripture collections
store.SaveSource(SourceScriptures)

// cmdIndexTalks saves only conference collections
store.SaveSource(SourceConference)

// cmdIndexManuals saves only manual collections
store.SaveSource(SourceManual)
```

#### MCP Server Changes

```go
// Load all files at startup (could parallelize)
store.Load()  // reads *.gob.gz from data/

// Search remains unchanged — all collections are in the same in-memory DB
```

#### Migration

- Add a `migrate` command that reads old `gospel-vec.gob.gz` and writes
  per-collection files
- Could auto-detect: if single old file exists but no per-collection files,
  run migration automatically
- Keep backward compatibility: if old file found, load it

### Performance Estimates

Rough math for per-collection saves (vs current single-file):

| Collection | Est. Chunks | Est. File Size | Est. Save Time |
|-----------|-------------|---------------|---------------|
| scriptures-verse | ~31K | ~300 MB | ~30s |
| scriptures-paragraph | ~8K | ~100 MB | ~10s |
| scriptures-summary | ~1.5K | ~20 MB | ~2s |
| scriptures-theme | ~1.5K | ~20 MB | ~2s |
| conference-paragraph | ~125K | ~1.2 GB | ~2m |
| conference-summary | ~4K | ~50 MB | ~5s |
| manual-paragraph | ~5K | ~60 MB | ~6s |
| manual-summary | ~1K | ~15 MB | ~2s |

**After indexing scriptures:** Only save 4 scripture files (~450 MB total)
instead of the full 2.3 GB → ~4x faster save.

**After indexing talks:** Only save 2 conference files (~1.25 GB total)
instead of 2.3 GB → still big but ~2x faster.

**MCP startup:** Load 8 files in parallel instead of 1 serial file → could
be up to 4-8x faster on multi-core systems (depends on disk).

### Bonus: Parallel gzip

Regardless of file splitting, we could replace standard `compress/gzip` with
[`pgzip`](https://github.com/klauspost/pgzip) for parallel compression.
This is a drop-in replacement that uses multiple goroutines. Since chromem-go
uses `gzip.NewWriter` internally, we'd need to use `ExportToWriter` and pass
our own pgzip writer. This alone could give 2-4x speedup on save.

However, this requires wrapping chromem-go's export rather than calling
`ExportToFile` directly.

## Future Considerations

- **Lazy loading**: MCP server could load collections on-demand during first
  search of that type. Would reduce startup time if user typically only
  searches scriptures.
- **Memory-mapped files**: For read-only search, could mmap the files instead
  of loading into heap. Would require a different storage format.
- **Different serialization**: `encoding/gob` is known to be slow for large
  datasets. Alternatives like Protocol Buffers, FlatBuffers, or MessagePack
  would be faster but require forking/wrapping chromem-go.
- **SQLite + vec extension**: Entirely different approach — store embeddings
  in SQLite with the vec0 virtual table. Built for this use case. But
  would be a full rewrite of the storage layer.
