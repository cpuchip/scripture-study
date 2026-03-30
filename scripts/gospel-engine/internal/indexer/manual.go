package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/urlgen"
)

func (idx *Indexer) indexManualFile(ctx context.Context, path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	// gospel-library/eng/manual/{collection}/{section}.md
	if len(parts) < 4 {
		return nil
	}

	filename := filepath.Base(path)
	var collectionID, section string
	if len(parts) == 4 {
		collectionID = strings.TrimSuffix(filename, ".md")
	} else {
		collectionID = parts[3]
		section = strings.TrimSuffix(filename, ".md")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fullContent := string(content)

	title := extractTitle(fullContent)
	if title == "" {
		title = formatTitle(filename)
	}

	contentType := "manual"
	if strings.Contains(collectionID, "handbook") {
		contentType = "handbook"
	}

	sourceURL := urlgen.Manual(collectionID, section)

	if _, err := idx.db.Exec(`
		INSERT OR REPLACE INTO manuals (content_type, collection_id, section, title, content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, contentType, collectionID, section, title, fullContent, relPath, sourceURL); err != nil {
		return err
	}
	result.ManualsIndexed++

	return idx.recordMetadata(path, info, "manual", 1)
}

func (idx *Indexer) indexMusicFile(ctx context.Context, path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	// gospel-library/eng/music/{collection}/{song}.md
	if len(parts) < 5 {
		return nil
	}

	collectionID := parts[3]
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

	sourceURL := urlgen.Manual(collectionID, section)

	if _, err := idx.db.Exec(`
		INSERT OR REPLACE INTO manuals (content_type, collection_id, section, title, content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "music", collectionID, section, title, fullContent, relPath, sourceURL); err != nil {
		return err
	}
	result.MusicIndexed++

	return idx.recordMetadata(path, info, "music", 1)
}
