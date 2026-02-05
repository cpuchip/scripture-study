package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/philippgille/chromem-go"
)

// Store manages the vector database with single-file persistence
type Store struct {
	db     *chromem.DB
	config *Config
	embed  chromem.EmbeddingFunc
	mu     sync.RWMutex
}

// NewStore creates a new store with the given config
func NewStore(cfg *Config, embedFunc chromem.EmbeddingFunc) (*Store, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	// Create in-memory database
	db := chromem.NewDB()

	store := &Store{
		db:     db,
		config: cfg,
		embed:  embedFunc,
	}

	// Load existing data if available
	if err := store.Load(); err != nil {
		// Not an error if file doesn't exist yet
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading database: %w", err)
		}
	}

	return store, nil
}

// Load imports the database from the compressed file
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dbPath := s.config.DBPath()

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return err // File doesn't exist, not an error for new stores
	}

	// Import from compressed file
	if err := s.db.ImportFromFile(dbPath, ""); err != nil {
		return fmt.Errorf("importing from %s: %w", dbPath, err)
	}

	fmt.Printf("📂 Loaded database from %s\n", dbPath)
	return nil
}

// Save exports the database to the compressed file
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dbPath := s.config.DBPath()

	// Export with compression
	if err := s.db.ExportToFile(dbPath, true, ""); err != nil {
		return fmt.Errorf("exporting to %s: %w", dbPath, err)
	}

	fmt.Printf("💾 Saved database to %s\n", dbPath)
	return nil
}

// GetOrCreateCollection gets or creates a collection with proper naming
func (s *Store) GetOrCreateCollection(source Source, layer Layer) (*chromem.Collection, error) {
	name := collectionName(source, layer)

	// Try to get existing collection
	col := s.db.GetCollection(name, s.embed)
	if col != nil {
		return col, nil
	}

	// Create new collection
	col, err := s.db.CreateCollection(name, nil, s.embed)
	if err != nil {
		return nil, fmt.Errorf("creating collection %s: %w", name, err)
	}

	return col, nil
}

// GetCollection gets an existing collection (returns nil if not found)
func (s *Store) GetCollection(source Source, layer Layer) *chromem.Collection {
	name := collectionName(source, layer)
	return s.db.GetCollection(name, s.embed)
}

// AddChunks adds multiple chunks to the appropriate collection
func (s *Store) AddChunks(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Group chunks by source+layer
	grouped := make(map[string][]Chunk)
	for _, chunk := range chunks {
		key := collectionName(chunk.Metadata.Source, chunk.Metadata.Layer)
		grouped[key] = append(grouped[key], chunk)
	}

	// Add to each collection
	for key, chunkGroup := range grouped {
		// Parse source and layer from key
		source, layer := parseCollectionName(key)

		col, err := s.GetOrCreateCollection(source, layer)
		if err != nil {
			return err
		}

		// Convert chunks to chromem documents
		docs := make([]chromem.Document, len(chunkGroup))
		for i, chunk := range chunkGroup {
			docs[i] = chromem.Document{
				ID:       chunk.ID,
				Content:  chunk.Content,
				Metadata: chunk.Metadata.ToMap(),
			}
		}

		// Add documents
		if err := col.AddDocuments(ctx, docs, runtime.NumCPU()); err != nil {
			return fmt.Errorf("adding documents to %s: %w", key, err)
		}
	}

	return nil
}

// Search searches across specified layers and sources
func (s *Store) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	var results []SearchResult

	// Determine which collections to search
	layers := opts.Layers
	if len(layers) == 0 {
		layers = []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme}
	}

	sources := opts.Sources
	if len(sources) == 0 {
		sources = []Source{SourceScriptures, SourceConference}
	}

	// Search each collection
	for _, source := range sources {
		for _, layer := range layers {
			col := s.GetCollection(source, layer)
			if col == nil {
				continue // Collection doesn't exist
			}

			// Adjust limit if collection has fewer documents
			limit := opts.Limit
			if col.Count() < limit {
				limit = col.Count()
			}
			if limit == 0 {
				continue // Empty collection
			}

			// Query the collection
			queryResults, err := col.Query(ctx, query, limit, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("querying %s-%s: %w", source, layer, err)
			}

			// Convert to SearchResults
			for _, r := range queryResults {
				results = append(results, SearchResult{
					Chunk: Chunk{
						ID:       r.ID,
						Content:  r.Content,
						Metadata: MetadataFromMap(r.Metadata),
					},
					Score: r.Similarity,
				})
			}
		}
	}

	return results, nil
}

// Stats returns statistics about the store
func (s *Store) Stats() map[string]int {
	stats := make(map[string]int)

	for _, source := range []Source{SourceScriptures, SourceConference} {
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			col := s.GetCollection(source, layer)
			if col != nil {
				name := collectionName(source, layer)
				stats[name] = col.Count()
			}
		}
	}

	return stats
}

// collectionName generates consistent collection names
func collectionName(source Source, layer Layer) string {
	return fmt.Sprintf("%s-%s", source, layer)
}

// parseCollectionName extracts source and layer from collection name
func parseCollectionName(name string) (Source, Layer) {
	// Simple split on "-"
	for _, source := range []Source{SourceScriptures, SourceConference} {
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			if collectionName(source, layer) == name {
				return source, layer
			}
		}
	}
	return "", ""
}

// FileSize returns the size of the database file
func (s *Store) FileSize() (int64, error) {
	dbPath := s.config.DBPath()
	info, err := os.Stat(dbPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// DataDir returns the data directory path
func (s *Store) DataDir() string {
	return s.config.DataDir
}

// AbsDataDir returns the absolute data directory path
func (s *Store) AbsDataDir() (string, error) {
	return filepath.Abs(s.config.DataDir)
}
