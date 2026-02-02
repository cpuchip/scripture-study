package indexer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/urlgen"
)

// Scripture volumes we handle
var scriptureVolumes = map[string]bool{
	"ot":           true, // Old Testament
	"nt":           true, // New Testament
	"bofm":         true, // Book of Mormon
	"dc-testament": true, // Doctrine and Covenants
	"pgp":          true, // Pearl of Great Price
}

// Study aids (we index but handle differently)
var studyAids = map[string]bool{
	"tg": true, // Topical Guide
	"bd": true, // Bible Dictionary
	"gs": true, // Guide to the Scriptures
}

// Verse pattern: **1.** or **12.**
var versePattern = regexp.MustCompile(`^\*\*(\d+)\.\*\*\s*(.+)`)

// Footnote pattern: <sup>[1a](#fn-1a)</sup> or TG/BD references
var footnotePattern = regexp.MustCompile(`<sup>\[([^\]]+)\]\(#fn-[^)]+\)</sup>`)

// Cross-reference link pattern in footnotes: [Gen. 1:1](../../ot/gen/1.md)
var crossRefPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+\.md[^)]*)\)`)

func (idx *Indexer) indexScriptureFile(path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(relPath, string(os.PathSeparator))

	// Expecting: gospel-library/eng/scriptures/{volume}/{book}/{chapter}.md
	if len(parts) < 6 {
		return nil // Skip index files, etc.
	}

	volume := parts[3]
	book := parts[4]

	// Skip study aids for now (handle them separately)
	if studyAids[volume] {
		return nil
	}

	// Skip non-scripture volumes
	if !scriptureVolumes[volume] {
		return nil
	}

	// Parse chapter number from filename
	filename := filepath.Base(path)
	chapterStr := strings.TrimSuffix(filename, ".md")
	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		// Non-numeric chapter (like intro files), skip
		return nil
	}

	// Read and parse the file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	fullContent := string(content)

	// Extract title (first line starting with #)
	var title string
	scanner := bufio.NewScanner(strings.NewReader(fullContent))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Insert chapter record
	chapterURL := urlgen.ScriptureChapter(volume, book, chapter)
	_, err = idx.db.Exec(`
		INSERT OR REPLACE INTO chapters (volume, book, chapter, title, full_content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, volume, book, chapter, title, fullContent, relPath, chapterURL)
	if err != nil {
		return fmt.Errorf("inserting chapter: %w", err)
	}
	result.ChaptersIndexed++

	// Parse verses
	verses := parseVerses(fullContent)
	for _, v := range verses {
		verseURL := urlgen.Scripture(volume, book, chapter, v.Number)
		_, err = idx.db.Exec(`
			INSERT OR REPLACE INTO scriptures (volume, book, chapter, verse, text, file_path, source_url)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, volume, book, chapter, v.Number, v.Text, relPath, verseURL)
		if err != nil {
			return fmt.Errorf("inserting verse %d: %w", v.Number, err)
		}
		result.ScripturesIndexed++

		// Extract and store cross-references from footnotes
		crossRefs := extractCrossReferences(fullContent, volume, book, chapter, v.Number)
		for _, ref := range crossRefs {
			_, err = idx.db.Exec(`
				INSERT OR IGNORE INTO cross_references 
				(source_volume, source_book, source_chapter, source_verse, 
				 target_volume, target_book, target_chapter, target_verse, reference_type)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, volume, book, chapter, v.Number,
				ref.TargetVolume, ref.TargetBook, ref.TargetChapter, ref.TargetVerse, ref.Type)
			if err == nil {
				result.CrossRefsIndexed++
			}
		}
	}

	// Record metadata for incremental indexing
	return idx.recordMetadata(path, info, "scripture", len(verses))
}

type verse struct {
	Number int
	Text   string
}

func parseVerses(content string) []verse {
	var verses []verse
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		if matches := versePattern.FindStringSubmatch(line); matches != nil {
			num, _ := strconv.Atoi(matches[1])
			text := matches[2]

			// Strip footnote markers for cleaner text
			text = footnotePattern.ReplaceAllString(text, "")
			text = strings.TrimSpace(text)

			verses = append(verses, verse{Number: num, Text: text})
		}
	}

	return verses
}

type crossRef struct {
	TargetVolume  string
	TargetBook    string
	TargetChapter int
	TargetVerse   *int
	Type          string
}

func extractCrossReferences(content string, sourceVolume, sourceBook string, sourceChapter, sourceVerse int) []crossRef {
	var refs []crossRef

	// Find the footnotes section (after --- ## Footnotes)
	footnoteSection := ""
	if idx := strings.Index(content, "## Footnotes"); idx != -1 {
		footnoteSection = content[idx:]
	}

	if footnoteSection == "" {
		return refs
	}

	// Look for cross-reference links
	matches := crossRefPattern.FindAllStringSubmatch(footnoteSection, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		linkPath := match[2]

		// Parse the link path to extract volume/book/chapter
		ref := parseCrossRefLink(linkPath)
		if ref != nil {
			refs = append(refs, *ref)
		}
	}

	return refs
}

func parseCrossRefLink(linkPath string) *crossRef {
	// Handle relative paths like ../../ot/gen/1.md or ../../tg/faith.md
	parts := strings.Split(linkPath, "/")

	// Find the volume part (after the ../.. parts)
	volIdx := -1
	for i, p := range parts {
		if scriptureVolumes[p] || studyAids[p] {
			volIdx = i
			break
		}
	}

	if volIdx == -1 || volIdx+1 >= len(parts) {
		return nil
	}

	volume := parts[volIdx]
	book := parts[volIdx+1]

	// For study aids, we don't have chapter/verse
	if studyAids[volume] {
		return &crossRef{
			TargetVolume: volume,
			TargetBook:   strings.TrimSuffix(book, ".md"),
			Type:         volume, // tg, bd, gs
		}
	}

	// Scripture reference
	if volIdx+2 >= len(parts) {
		return nil
	}

	chapterFile := parts[volIdx+2]
	chapterStr := strings.TrimSuffix(chapterFile, ".md")

	// Handle anchor for verse: 1.md#p19 -> verse 19
	var verseNum *int
	if idx := strings.Index(chapterStr, "#"); idx != -1 {
		anchor := chapterStr[idx+1:]
		chapterStr = chapterStr[:idx]
		// Parse #p19 style anchors
		if strings.HasPrefix(anchor, "p") {
			if v, err := strconv.Atoi(anchor[1:]); err == nil {
				verseNum = &v
			}
		}
	}

	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		return nil
	}

	return &crossRef{
		TargetVolume:  volume,
		TargetBook:    book,
		TargetChapter: chapter,
		TargetVerse:   verseNum,
		Type:          "footnote",
	}
}
