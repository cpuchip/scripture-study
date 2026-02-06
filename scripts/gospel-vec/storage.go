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

// Store manages the vector database with per-source file persistence.
// Each source (scriptures, conference, manual) is saved to its own .gob.gz file,
// enabling incremental saves and parallel loading.
type Store struct {
	db     *chromem.DB
	config *Config
	embed  chromem.EmbeddingFunc
	mu     sync.RWMutex
}

// sourceFile returns the per-source database filename (e.g., "scriptures.gob.gz")
func sourceFile(source Source) string {
	return string(source) + ".gob.gz"
}

// sourcePath returns the full path to a per-source database file
func (s *Store) sourcePath(source Source) string {
	return filepath.Join(s.config.DataDir, sourceFile(source))
}

// collectionsForSource returns all possible collection names for a given source
func collectionsForSource(source Source) []string {
	layers := []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme}
	names := make([]string, 0, len(layers))
	for _, layer := range layers {
		names = append(names, collectionName(source, layer))
	}
	return names
}

// allSources returns all known sources
func allSources() []Source {
	return []Source{SourceScriptures, SourceConference, SourceManual}
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

// Load imports the database from per-source files, falling back to legacy single file.
// Per-source files are loaded sequentially from the data directory.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try per-source files first (new format)
	loaded := 0
	for _, source := range allSources() {
		path := s.sourcePath(source)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := s.db.ImportFromFile(path, ""); err != nil {
			return fmt.Errorf("importing %s: %w", path, err)
		}
		fmt.Printf("📂 Loaded %s\n", path)
		loaded++
	}

	if loaded > 0 {
		return nil
	}

	// Fallback: try legacy single file
	legacyPath := s.config.DBPath()
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return err // File doesn't exist, not an error for new stores
	}

	if err := s.db.ImportFromFile(legacyPath, ""); err != nil {
		return fmt.Errorf("importing legacy %s: %w", legacyPath, err)
	}

	fmt.Printf("📂 Loaded legacy database from %s\n", legacyPath)
	fmt.Println("💡 Run 'gospel-vec migrate' to convert to per-source files for faster saves")
	return nil
}

// saveSource is the internal unlocked implementation of SaveSource.
func (s *Store) saveSource(source Source) error {
	path := s.sourcePath(source)
	tmpPath := path + ".tmp"

	collNames := collectionsForSource(source)

	// Export only this source's collections to temp file
	if err := s.db.ExportToFile(tmpPath, true, "", collNames...); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("exporting %s: %w", source, err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tmpPath, path, err)
	}

	fmt.Printf("💾 Saved %s\n", path)
	return nil
}

// SaveSource exports only the collections for the given source to its own file.
// This is much faster than saving the entire database.
func (s *Store) SaveSource(source Source) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveSource(source)
}

// Save exports all sources that have data, each to their own file.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, source := range allSources() {
		// Check if source has any data
		hasData := false
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			name := collectionName(source, layer)
			col := s.db.GetCollection(name, s.embed)
			if col != nil && col.Count() > 0 {
				hasData = true
				break
			}
		}
		if hasData {
			if err := s.saveSource(source); err != nil {
				return err
			}
		}
	}
	return nil
}

// Migrate converts a legacy single-file database to per-source files.
func (s *Store) Migrate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	legacyPath := s.config.DBPath()

	// Check if legacy file exists
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return fmt.Errorf("no legacy database found at %s", legacyPath)
	}

	// Check if any per-source files already exist
	for _, source := range allSources() {
		path := s.sourcePath(source)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("per-source file already exists: %s (migration already done?)", path)
		}
	}

	// Load from legacy file
	fmt.Printf("📂 Loading legacy database from %s...\n", legacyPath)
	if err := s.db.ImportFromFile(legacyPath, ""); err != nil {
		return fmt.Errorf("importing legacy: %w", err)
	}

	// Print stats
	total := 0
	for _, source := range allSources() {
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			name := collectionName(source, layer)
			col := s.db.GetCollection(name, s.embed)
			if col != nil && col.Count() > 0 {
				fmt.Printf("   %s: %d docs\n", name, col.Count())
				total += col.Count()
			}
		}
	}
	fmt.Printf("   Total: %d docs\n\n", total)

	// Save each source to its own file
	saved := 0
	for _, source := range allSources() {
		hasData := false
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			name := collectionName(source, layer)
			col := s.db.GetCollection(name, s.embed)
			if col != nil && col.Count() > 0 {
				hasData = true
				break
			}
		}
		if hasData {
			if err := s.saveSource(source); err != nil {
				return err
			}
			saved++
		}
	}

	// Back up legacy file
	backupPath := legacyPath + ".migrated"
	if err := os.Rename(legacyPath, backupPath); err != nil {
		return fmt.Errorf("backing up legacy file: %w", err)
	}
	fmt.Printf("\n📦 Legacy file backed up to %s\n", backupPath)
	fmt.Printf("✅ Migration complete: %d source files created\n", saved)

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
		sources = []Source{SourceScriptures, SourceConference, SourceManual}
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

	for _, source := range []Source{SourceScriptures, SourceConference, SourceManual} {
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
	for _, source := range []Source{SourceScriptures, SourceConference, SourceManual} {
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			if collectionName(source, layer) == name {
				return source, layer
			}
		}
	}
	return "", ""
}

// FileSize returns the total size of all database files
func (s *Store) FileSize() (int64, error) {
	var total int64

	// Sum per-source files
	for _, source := range allSources() {
		path := s.sourcePath(source)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return 0, err
		}
		total += info.Size()
	}

	// Also check legacy file if no per-source files found
	if total == 0 {
		dbPath := s.config.DBPath()
		info, err := os.Stat(dbPath)
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	}

	return total, nil
}

// DataDir returns the data directory path
func (s *Store) DataDir() string {
	return s.config.DataDir
}

// AbsDataDir returns the absolute data directory path
func (s *Store) AbsDataDir() (string, error) {
	return filepath.Abs(s.config.DataDir)
}
