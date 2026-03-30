package vec

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MmapStore provides vector search backed by mmap'd .vecf files and SQLite
// metadata. Startup is nearly instant — embeddings are memory-mapped rather
// than loaded into the Go heap.
type MmapStore struct {
	files   map[string]*VecFile // collection name → mmap'd file
	db      *sql.DB
	embed   func(ctx context.Context, text string) ([]float32, error)
	dataDir string
	dim     int
}

// NewMmapStore opens all .vecf files in dataDir and connects to the SQLite
// database for metadata lookups. The embedding function is used to embed
// search queries at query time.
func NewMmapStore(dataDir string, dbPath string, embedFunc func(ctx context.Context, text string) ([]float32, error)) (*MmapStore, error) {
	store := &MmapStore{
		files:   make(map[string]*VecFile),
		embed:   embedFunc,
		dataDir: dataDir,
	}

	// Open all .vecf files
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("reading data dir: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".vecf") {
			continue
		}
		collection := strings.TrimSuffix(entry.Name(), ".vecf")
		path := filepath.Join(dataDir, entry.Name())
		vf, err := OpenVecFile(path)
		if err != nil {
			return nil, fmt.Errorf("opening %s: %w", path, err)
		}
		store.files[collection] = vf
		if store.dim == 0 {
			store.dim = vf.Dim()
		}
	}

	if len(store.files) == 0 {
		return nil, fmt.Errorf("no .vecf files found in %s (run 'gospel-engine convert' first)", dataDir)
	}

	// Open SQLite for metadata
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&mode=ro", dbPath)
	store.db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("opening metadata db: %w", err)
	}

	return store, nil
}

// Search implements vec.Searcher.
func (s *MmapStore) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	if s.embed == nil {
		return nil, fmt.Errorf("no embedding function configured")
	}

	// Embed the query
	queryVec, err := s.embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embedding query: %w", err)
	}
	normalizeVector(queryVec)

	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	sources := opts.Sources
	if len(sources) == 0 {
		sources = []Source{SourceScriptures, SourceConference, SourceManual, SourceMusic}
	}
	layers := opts.Layers
	if len(layers) == 0 {
		layers = []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme}
	}

	type hit struct {
		collection string
		idx        int
		score      float32
	}
	var allHits []hit

	// Search each relevant collection
	for _, source := range sources {
		for _, layer := range layers {
			name := collectionName(source, layer)
			vf, ok := s.files[name]
			if !ok || vf.Count() == 0 {
				continue
			}

			indices, scores, err := vf.TopK(queryVec, opts.Limit)
			if err != nil {
				return nil, fmt.Errorf("searching %s: %w", name, err)
			}

			for i := range indices {
				allHits = append(allHits, hit{
					collection: name,
					idx:        indices[i],
					score:      scores[i],
				})
			}
		}
	}

	// Sort all hits by score, take top-K
	sort.Slice(allHits, func(i, j int) bool { return allHits[i].score > allHits[j].score })
	if len(allHits) > opts.Limit {
		allHits = allHits[:opts.Limit]
	}

	// Look up metadata from SQLite
	results := make([]SearchResult, 0, len(allHits))
	for _, h := range allHits {
		meta, content, err := s.getMetadata(h.collection, h.idx)
		if err != nil {
			continue // Skip missing metadata
		}

		results = append(results, SearchResult{
			Chunk: Chunk{
				ID:       fmt.Sprintf("%s-%d", h.collection, h.idx),
				Content:  content,
				Metadata: meta,
			},
			Score: h.score,
		})
	}

	return results, nil
}

// Stats implements vec.Searcher.
func (s *MmapStore) Stats() map[string]int {
	stats := make(map[string]int)
	for name, vf := range s.files {
		stats[name] = vf.Count()
	}
	return stats
}

// Close releases all mmap'd files and closes the metadata DB.
func (s *MmapStore) Close() error {
	var errs []string
	for name, vf := range s.files {
		if err := vf.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("db: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (s *MmapStore) getMetadata(collection string, idx int) (*DocMetadata, string, error) {
	var (
		content                                       string
		source, layer, book, ref, rangeText, filePath string
		speaker, position, month, session, talkTitle  string
		chapter, year                                 int
	)

	err := s.db.QueryRow(`
		SELECT content, source, layer, book, chapter, reference, range_text,
		       file_path, speaker, position, year, month, session, talk_title
		FROM vec_docs
		WHERE collection = ? AND vec_idx = ?
	`, collection, idx).Scan(
		&content, &source, &layer, &book, &chapter, &ref, &rangeText,
		&filePath, &speaker, &position, &year, &month, &session, &talkTitle,
	)
	if err != nil {
		return nil, "", fmt.Errorf("metadata lookup: %w", err)
	}

	meta := &DocMetadata{
		Source:    Source(source),
		Layer:     Layer(layer),
		Book:      book,
		Chapter:   chapter,
		Reference: ref,
		Range:     rangeText,
		FilePath:  filePath,
		Speaker:   speaker,
		Position:  position,
		Year:      year,
		Month:     month,
		Session:   session,
		TalkTitle: talkTitle,
	}

	return meta, content, nil
}
