// Package vec provides vector search functionality for gospel content.
//
// Two implementations exist:
//   - Store: chromem-go in-memory store (used during indexing, loads from gob.gz)
//   - MmapStore: mmap-backed store (fast startup, loads embeddings on-demand from flat files)
package vec

import "context"

// Searcher is the interface for vector search backends.
// Both Store (chromem-go) and MmapStore (mmap) implement this.
type Searcher interface {
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
	Stats() map[string]int
}
