// Package indexer handles parsing and indexing of gospel content files.
package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/db"
)

// Indexer handles the indexing of gospel content.
type Indexer struct {
	db   *db.DB
	root string // Root directory containing gospel-library
}

// Options controls what and how to index.
type Options struct {
	Incremental bool   // Only index new/modified files
	Source      string // Filter by source type: scriptures, conference, manual, magazine
	PathFilter  string // Filter by path pattern
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
	CrossRefsIndexed  int
	Errors            []string
}

// New creates a new Indexer.
func New(database *db.DB, root string) *Indexer {
	return &Indexer{
		db:   database,
		root: root,
	}
}

// Index runs the indexing process with the given options.
func (idx *Indexer) Index(opts Options) (*Result, error) {
	result := &Result{}

	// Determine which sources to index
	sources := []string{"scriptures", "conference", "manual", "magazine", "books"}
	if opts.Source != "" {
		sources = []string{opts.Source}
	}

	for _, source := range sources {
		if err := idx.indexSource(source, opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", source, err))
		}
	}

	return result, nil
}

func (idx *Indexer) indexSource(source string, opts Options, result *Result) error {
	var basePath string
	var handler func(path string, info os.FileInfo, opts Options, result *Result) error

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
	case "magazine":
		basePath = filepath.Join(idx.root, "gospel-library", "eng", "liahona")
		handler = idx.indexMagazineFile
	case "books":
		basePath = filepath.Join(idx.root, "books")
		handler = idx.indexBooksFile
	default:
		return fmt.Errorf("unknown source: %s", source)
	}

	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", basePath)
	}

	// Walk the directory
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("walk error: %s: %v", path, err))
			return nil // continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Apply path filter if specified
		if opts.PathFilter != "" {
			relPath, _ := filepath.Rel(idx.root, path)
			if !strings.Contains(relPath, opts.PathFilter) {
				return nil
			}
		}

		// Check if incremental and file needs reindex
		if opts.Incremental {
			relPath, _ := filepath.Rel(idx.root, path)
			needsReindex, err := idx.db.NeedsReindex(relPath, info.ModTime(), info.Size())
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("checking %s: %v", path, err))
				return nil
			}
			if !needsReindex {
				result.FilesSkipped++
				return nil
			}
		}

		// Process the file
		if err := handler(path, info, opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", path, err))
		} else {
			result.FilesProcessed++
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
