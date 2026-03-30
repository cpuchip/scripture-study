package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func (idx *Indexer) indexBookFile(ctx context.Context, path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	// books/{collection}/{section}.md
	if len(parts) < 3 {
		return nil
	}

	collection := parts[1]
	filename := filepath.Base(path)
	section := strings.TrimSuffix(filename, ".md")

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fullContent := string(content)

	title := extractTitle(fullContent)
	if title == "" {
		title = formatTitle(filename)
	}

	if _, err := idx.db.Exec(`
		INSERT OR REPLACE INTO books (collection, section, title, content, file_path)
		VALUES (?, ?, ?, ?, ?)
	`, collection, section, title, fullContent, relPath); err != nil {
		return err
	}
	result.BooksIndexed++

	return idx.recordMetadata(path, info, "book", 1)
}
