package indexer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/urlgen"
)

func (idx *Indexer) indexManualFile(path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(relPath, string(os.PathSeparator))

	// Expecting: gospel-library/eng/manual/{collection}/{section}.md
	// or: gospel-library/eng/manual/{collection}.md (single file manual)
	if len(parts) < 4 {
		return nil
	}

	var collectionID, section string
	filename := filepath.Base(path)

	if len(parts) == 4 {
		// Single file: gospel-library/eng/manual/some-manual.md
		collectionID = strings.TrimSuffix(filename, ".md")
		section = ""
	} else {
		// Multi-file: gospel-library/eng/manual/collection/section.md
		collectionID = parts[3]
		section = strings.TrimSuffix(filename, ".md")
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fullContent := string(content)

	// Extract title
	title := extractManualTitle(fullContent)
	if title == "" {
		title = formatManualTitle(filename)
	}

	// Determine content type
	contentType := "manual"
	if strings.Contains(collectionID, "handbook") {
		contentType = "handbook"
	}

	// Generate source URL
	sourceURL := urlgen.Manual(collectionID, section)

	// Insert manual record
	_, err = idx.db.Exec(`
		INSERT OR REPLACE INTO manuals (content_type, collection_id, section, title, content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, contentType, collectionID, section, title, fullContent, relPath, sourceURL)
	if err != nil {
		return err
	}

	result.ManualsIndexed++

	// Record metadata
	return idx.recordMetadata(path, info, "manual", 1)
}

func (idx *Indexer) indexMagazineFile(path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(relPath, string(os.PathSeparator))

	// Expecting: gospel-library/eng/liahona/{year}/{month}/{article}.md
	if len(parts) < 6 {
		return nil
	}

	magazineName := parts[3] // liahona, ensign, etc.
	yearStr := parts[4]
	monthStr := parts[5]
	filename := filepath.Base(path)

	// Parse year and month (skip if not numeric)
	year := 0
	month := 0
	if y, err := parseInt(yearStr); err == nil {
		year = y
	}
	if m, err := parseInt(monthStr); err == nil {
		month = m
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fullContent := string(content)

	// Extract title
	title := extractManualTitle(fullContent)
	if title == "" {
		title = formatManualTitle(filename)
	}

	// Collection ID for magazines
	collectionID := magazineName
	if year > 0 && month > 0 {
		collectionID = filepath.Join(magazineName, yearStr, monthStr)
	}

	// Section is the article name
	section := strings.TrimSuffix(filename, ".md")

	// Generate source URL
	sourceURL := urlgen.Magazine(magazineName, year, month, section)

	// Insert as manual with magazine content type
	_, err = idx.db.Exec(`
		INSERT OR REPLACE INTO manuals (content_type, collection_id, section, title, content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "magazine", collectionID, section, title, fullContent, relPath, sourceURL)
	if err != nil {
		return err
	}

	result.ManualsIndexed++

	// Record metadata
	return idx.recordMetadata(path, info, "magazine", 1)
}

// manualTitlePattern matches markdown titles
var manualTitlePattern = regexp.MustCompile(`^#\s+(.+)$`)

func extractManualTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := manualTitlePattern.FindStringSubmatch(line); matches != nil {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func formatManualTitle(filename string) string {
	// Convert "01-lesson.md" or "some-topic.md" to readable title
	name := strings.TrimSuffix(filename, ".md")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Title case
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
