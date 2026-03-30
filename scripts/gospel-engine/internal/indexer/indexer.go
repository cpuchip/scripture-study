// Package indexer handles parsing and indexing gospel content into both SQLite and vector stores.
package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/vec"
)

// Indexer handles indexing content into both SQLite and vector stores.
type Indexer struct {
	db   *db.DB
	vec  *vec.Store
	root string // Workspace root
}

// Options controls what and how to index.
type Options struct {
	Incremental     bool        // Only index new/modified files
	Source          string      // Filter: scriptures, conference, manual, books, music
	Layers          []vec.Layer // Which vector layers to build
	MaxRetries      int
	ContinueOnError bool
	SaveInterval    int // Save vector store every N files
	Verbose         bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Incremental:     true,
		Layers:          []vec.Layer{vec.LayerVerse, vec.LayerParagraph},
		MaxRetries:      3,
		ContinueOnError: true,
		SaveInterval:    500,
		Verbose:         true,
	}
}

// Result contains the results of an indexing run.
type Result struct {
	FilesProcessed    int
	FilesSkipped      int
	ScripturesIndexed int
	ChaptersIndexed   int
	TalksIndexed      int
	ManualsIndexed    int
	BooksIndexed      int
	MusicIndexed      int
	CrossRefsIndexed  int
	VecChunksAdded    int
	Errors            []string
	Duration          time.Duration
}

// New creates a new Indexer.
func New(database *db.DB, store *vec.Store, root string) *Indexer {
	return &Indexer{
		db:   database,
		vec:  store,
		root: root,
	}
}

// Index runs the indexing process.
func (idx *Indexer) Index(ctx context.Context, opts Options) (*Result, error) {
	start := time.Now()
	result := &Result{}

	sources := []string{"scriptures", "conference", "manual", "books", "music"}
	if opts.Source != "" {
		sources = []string{opts.Source}
	}

	for _, source := range sources {
		if err := idx.indexSource(ctx, source, opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", source, err))
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

func (idx *Indexer) indexSource(ctx context.Context, source string, opts Options, result *Result) error {
	var basePath string
	var handler func(ctx context.Context, path string, info os.FileInfo, opts Options, result *Result) error

	switch source {
	case "scriptures":
		basePath = filepath.Join(idx.root, "gospel-library", "eng", "scriptures")
		handler = idx.indexScriptureFile
	case "conference":
		basePath = filepath.Join(idx.root, "gospel-library", "eng", "general-conference")
		handler = idx.indexTalkFile
	case "manual":
		basePath = filepath.Join(idx.root, "gospel-library", "eng", "manual")
		handler = idx.indexManualFile
	case "books":
		basePath = filepath.Join(idx.root, "books")
		handler = idx.indexBookFile
	case "music":
		basePath = filepath.Join(idx.root, "gospel-library", "eng", "music")
		handler = idx.indexMusicFile
	default:
		return fmt.Errorf("unknown source: %s", source)
	}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if opts.Verbose {
			fmt.Printf("⏭️  Skipping %s (path not found: %s)\n", source, basePath)
		}
		return nil
	}

	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("walk: %s: %v", path, err))
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Incremental check
		if opts.Incremental {
			relPath, _ := filepath.Rel(idx.root, path)
			needsReindex, err := idx.db.NeedsReindex(relPath, info.ModTime(), info.Size())
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("check %s: %v", path, err))
				return nil
			}
			if !needsReindex {
				result.FilesSkipped++
				return nil
			}
		}

		if err := handler(ctx, path, info, opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
			if !opts.ContinueOnError {
				return err
			}
		} else {
			result.FilesProcessed++
			if opts.Verbose && result.FilesProcessed%100 == 0 {
				fmt.Printf("   ... %d files processed (%d skipped, %d vec chunks)\n",
					result.FilesProcessed, result.FilesSkipped, result.VecChunksAdded)
			}
			// Periodic vector store save
			if opts.SaveInterval > 0 && idx.vec != nil && result.VecChunksAdded > 0 &&
				result.FilesProcessed%opts.SaveInterval == 0 {
				if opts.Verbose {
					fmt.Printf("💾 Checkpoint save at %d files...\n", result.FilesProcessed)
				}
				if err := idx.vec.Save(); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("checkpoint save: %v", err))
				}
			}
		}

		return nil
	})
}

// recordMetadata saves file metadata for incremental indexing.
func (idx *Indexer) recordMetadata(path string, info os.FileInfo, contentType string, recordCount int) error {
	relPath, _ := filepath.Rel(idx.root, path)
	return idx.db.SetFileMetadata(&db.FileMetadata{
		FilePath:    relPath,
		ContentType: contentType,
		IndexedAt:   time.Now(),
		FileMtime:   info.ModTime(),
		FileSize:    info.Size(),
		RecordCount: recordCount,
	})
}

// addVecChunks adds chunks to the vector store with retry.
func (idx *Indexer) addVecChunks(ctx context.Context, chunks []vec.Chunk, opts Options) error {
	if len(chunks) == 0 || idx.vec == nil {
		return nil
	}
	var lastErr error
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		lastErr = idx.vec.AddChunks(ctx, chunks)
		if lastErr == nil {
			return nil
		}
		errStr := strings.ToLower(lastErr.Error())
		if !strings.Contains(errStr, "connection") && !strings.Contains(errStr, "timeout") && !strings.Contains(errStr, "eof") {
			return lastErr
		}
	}
	return fmt.Errorf("after %d retries: %w", opts.MaxRetries+1, lastErr)
}
