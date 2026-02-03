package indexer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (idx *Indexer) indexBooksFile(path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(relPath, string(os.PathSeparator))

	// Expecting: books/{collection}/{section}.md
	if len(parts) < 3 {
		return nil
	}

	collection := parts[1] // lecture-on-faith, etc.
	filename := filepath.Base(path)
	section := strings.TrimSuffix(filename, ".md")

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fullContent := string(content)

	// Extract title from first markdown heading
	title := extractBookTitle(fullContent)
	if title == "" {
		title = formatBookSectionTitle(collection, section)
	}

	// Insert book record
	_, err = idx.db.Exec(`
		INSERT OR REPLACE INTO books (collection, section, title, content, file_path)
		VALUES (?, ?, ?, ?, ?)
	`, collection, section, title, fullContent, relPath)
	if err != nil {
		return err
	}

	result.BooksIndexed++

	// Record metadata
	return idx.recordMetadata(path, info, "book", 1)
}

// bookTitlePattern matches markdown titles
var bookTitlePattern = regexp.MustCompile(`^#\s+(.+)$`)

func extractBookTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := bookTitlePattern.FindStringSubmatch(line); matches != nil {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func formatBookSectionTitle(collection, section string) string {
	// Convert collection name to readable format
	collectionName := strings.ReplaceAll(collection, "-", " ")
	collectionName = strings.Title(collectionName)

	// Convert section name
	sectionName := strings.ReplaceAll(section, "_", " ")
	sectionName = strings.TrimLeft(sectionName, "0123456789 ")
	sectionName = strings.Title(sectionName)

	return collectionName + " - " + sectionName
}
