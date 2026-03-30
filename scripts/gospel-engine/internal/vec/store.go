package vec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/philippgille/chromem-go"
)

// Layer represents the granularity of indexed content.
type Layer string

const (
	LayerVerse     Layer = "verse"
	LayerParagraph Layer = "paragraph"
	LayerSummary   Layer = "summary"
	LayerTheme     Layer = "theme"
)

// Source represents the content source.
type Source string

const (
	SourceScriptures Source = "scriptures"
	SourceConference Source = "conference"
	SourceManual     Source = "manual"
	SourceMusic      Source = "music"
)

// DocMetadata contains metadata for each indexed document.
type DocMetadata struct {
	Source    Source `json:"source"`
	Layer     Layer  `json:"layer"`
	Book      string `json:"book"`
	Chapter   int    `json:"chapter"`
	Reference string `json:"reference"`
	Range     string `json:"range"`
	FilePath  string `json:"filepath"`
	Generated bool   `json:"generated"`
	Model     string `json:"model"`
	Timestamp string `json:"timestamp"`
	// Conference talk fields
	Speaker   string `json:"speaker,omitempty"`
	Position  string `json:"position,omitempty"`
	Year      int    `json:"year,omitempty"`
	Month     string `json:"month,omitempty"`
	Session   string `json:"session,omitempty"`
	TalkTitle string `json:"talktitle,omitempty"`
}

// ToMap converts metadata to map[string]string for chromem-go.
func (m *DocMetadata) ToMap() map[string]string {
	result := map[string]string{
		"source":    string(m.Source),
		"layer":     string(m.Layer),
		"book":      m.Book,
		"chapter":   fmt.Sprintf("%d", m.Chapter),
		"reference": m.Reference,
		"range":     m.Range,
		"filepath":  m.FilePath,
		"generated": fmt.Sprintf("%t", m.Generated),
		"model":     m.Model,
		"timestamp": m.Timestamp,
	}
	if m.Speaker != "" {
		result["speaker"] = m.Speaker
	}
	if m.Position != "" {
		result["position"] = m.Position
	}
	if m.Year > 0 {
		result["year"] = fmt.Sprintf("%d", m.Year)
	}
	if m.Month != "" {
		result["month"] = m.Month
	}
	if m.Session != "" {
		result["session"] = m.Session
	}
	if m.TalkTitle != "" {
		result["talktitle"] = m.TalkTitle
	}
	return result
}

// MetadataFromMap converts map back to DocMetadata.
func MetadataFromMap(m map[string]string) *DocMetadata {
	chapter := 0
	fmt.Sscanf(m["chapter"], "%d", &chapter)
	year := 0
	if m["year"] != "" {
		fmt.Sscanf(m["year"], "%d", &year)
	}
	return &DocMetadata{
		Source:    Source(m["source"]),
		Layer:     Layer(m["layer"]),
		Book:      m["book"],
		Chapter:   chapter,
		Reference: m["reference"],
		Range:     m["range"],
		FilePath:  m["filepath"],
		Generated: m["generated"] == "true",
		Model:     m["model"],
		Timestamp: m["timestamp"],
		Speaker:   m["speaker"],
		Position:  m["position"],
		Year:      year,
		Month:     m["month"],
		Session:   m["session"],
		TalkTitle: m["talktitle"],
	}
}

// Chunk represents a piece of content to be indexed.
type Chunk struct {
	ID       string
	Content  string
	Metadata *DocMetadata
}

// SearchResult represents a search result with score.
type SearchResult struct {
	Chunk
	Score float32
}

// SearchOptions controls search behavior.
type SearchOptions struct {
	Layers  []Layer
	Sources []Source
	Limit   int
}

// Store manages the vector database with per-source file persistence.
type Store struct {
	db      *chromem.DB
	embed   chromem.EmbeddingFunc
	dataDir string
	mu      sync.RWMutex
}

func collectionName(source Source, layer Layer) string {
	return string(source) + "-" + string(layer)
}

func parseCollectionName(name string) (Source, Layer) {
	parts := strings.SplitN(name, "-", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return Source(parts[0]), Layer(parts[1])
}

func allSources() []Source {
	return []Source{SourceScriptures, SourceConference, SourceManual, SourceMusic}
}

func collectionsForSource(source Source) []string {
	layers := []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme}
	names := make([]string, 0, len(layers))
	for _, l := range layers {
		names = append(names, collectionName(source, l))
	}
	return names
}

// NewStore creates a new vector store.
func NewStore(dataDir string, embedFunc chromem.EmbeddingFunc) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	db := chromem.NewDB()
	store := &Store{
		db:      db,
		embed:   embedFunc,
		dataDir: dataDir,
	}

	if err := store.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading database: %w", err)
	}

	return store, nil
}

// Load imports from per-source files.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	loaded := 0
	for _, source := range allSources() {
		path := filepath.Join(s.dataDir, string(source)+".gob.gz")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := s.db.ImportFromFile(path, ""); err != nil {
			return fmt.Errorf("importing %s: %w", path, err)
		}
		fmt.Printf("📂 Loaded %s\n", path)
		loaded++
	}

	if loaded == 0 {
		return os.ErrNotExist
	}
	return nil
}

// SaveSource exports only the collections for the given source to its own file.
func (s *Store) SaveSource(source Source) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dataDir, string(source)+".gob.gz")
	tmpPath := path + ".tmp"
	collNames := collectionsForSource(source)

	if err := s.db.ExportToFile(tmpPath, true, "", collNames...); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("exporting %s: %w", source, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming %s: %w", tmpPath, err)
	}
	fmt.Printf("💾 Saved %s\n", path)
	return nil
}

// Save exports all sources that have data.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, source := range allSources() {
		hasData := false
		for _, l := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			name := collectionName(source, l)
			col := s.db.GetCollection(name, s.embed)
			if col != nil && col.Count() > 0 {
				hasData = true
				break
			}
		}
		if hasData {
			path := filepath.Join(s.dataDir, string(source)+".gob.gz")
			tmpPath := path + ".tmp"
			collNames := collectionsForSource(source)
			if err := s.db.ExportToFile(tmpPath, true, "", collNames...); err != nil {
				os.Remove(tmpPath)
				return fmt.Errorf("exporting %s: %w", source, err)
			}
			if err := os.Rename(tmpPath, path); err != nil {
				return fmt.Errorf("renaming %s: %w", tmpPath, err)
			}
			fmt.Printf("💾 Saved %s\n", path)
		}
	}
	return nil
}

// GetOrCreateCollection gets or creates a collection.
func (s *Store) GetOrCreateCollection(source Source, layer Layer) (*chromem.Collection, error) {
	name := collectionName(source, layer)
	col := s.db.GetCollection(name, s.embed)
	if col != nil {
		return col, nil
	}
	col, err := s.db.CreateCollection(name, nil, s.embed)
	if err != nil {
		return nil, fmt.Errorf("creating collection %s: %w", name, err)
	}
	return col, nil
}

// AddChunks adds multiple chunks to the appropriate collections.
func (s *Store) AddChunks(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Group by source+layer
	grouped := make(map[string][]Chunk)
	for _, chunk := range chunks {
		key := collectionName(chunk.Metadata.Source, chunk.Metadata.Layer)
		grouped[key] = append(grouped[key], chunk)
	}

	for key, chunkGroup := range grouped {
		source, layer := parseCollectionName(key)
		col, err := s.GetOrCreateCollection(source, layer)
		if err != nil {
			return err
		}

		docs := make([]chromem.Document, len(chunkGroup))
		for i, chunk := range chunkGroup {
			docs[i] = chromem.Document{
				ID:       chunk.ID,
				Content:  chunk.Content,
				Metadata: chunk.Metadata.ToMap(),
			}
		}

		// Use concurrency of 4 — local embedding servers process mostly sequentially
		// but can benefit from pipelining the HTTP requests.
		if err := col.AddDocuments(ctx, docs, 4); err != nil {
			return fmt.Errorf("adding %d docs to %s: %w", len(docs), key, err)
		}
	}

	return nil
}

// Search queries across layers and sources.
func (s *Store) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	sources := opts.Sources
	if len(sources) == 0 {
		sources = allSources()
	}
	layers := opts.Layers
	if len(layers) == 0 {
		layers = []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme}
	}

	var results []SearchResult

	for _, source := range sources {
		for _, layer := range layers {
			name := collectionName(source, layer)
			col := s.db.GetCollection(name, s.embed)
			if col == nil || col.Count() == 0 {
				continue
			}

			docs, err := col.Query(ctx, query, opts.Limit, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("querying %s: %w", name, err)
			}

			for _, doc := range docs {
				results = append(results, SearchResult{
					Chunk: Chunk{
						ID:       doc.ID,
						Content:  doc.Content,
						Metadata: MetadataFromMap(doc.Metadata),
					},
					Score: doc.Similarity,
				})
			}
		}
	}

	return results, nil
}

// Stats returns document counts per collection.
func (s *Store) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	for _, source := range allSources() {
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			name := collectionName(source, layer)
			col := s.db.GetCollection(name, s.embed)
			if col != nil && col.Count() > 0 {
				stats[name] = col.Count()
			}
		}
	}
	return stats
}
