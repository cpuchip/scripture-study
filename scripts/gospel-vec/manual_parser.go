package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ManualDefinition describes a manual or book to index
type ManualDefinition struct {
	Name string // Display name (e.g., "Teachings: Joseph Smith")
	Path string // Path to directory containing .md files
	Type string // "teachings", "cfm", "book", "manual"
}

// ParsedManualChapter contains parsed content from a manual chapter/lesson
type ParsedManualChapter struct {
	ManualName string
	Title      string // Chapter/lesson title
	Chapter    int    // Chapter/lesson number (0 if not applicable)
	Sections   []ManualSection
	Paragraphs []string // All paragraphs in order
	FilePath   string
}

// ManualSection represents a section within a chapter
type ManualSection struct {
	Heading string
	Level   int    // 2=H2, 3=H3
	Content string // Full text of section
}

// Patterns for manual parsing
var (
	manualH1Pattern    = regexp.MustCompile(`^#\s+(.+)$`)
	manualH2Pattern    = regexp.MustCompile(`^##\s+(.+)$`)
	manualH3Pattern    = regexp.MustCompile(`^###\s+(.+)$`)
	manualAudioPattern = regexp.MustCompile(`🎧\s*\[Listen to Audio\]\([^)]+\)`)
	manualImagePattern = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)]+)\)`)
	manualFootnoteRef  = regexp.MustCompile(`<sup>\[?\d+\]?</sup>|<sup>\[?<a[^>]*>[^<]*</a>\]?</sup>`)
	manualAnchorTag    = regexp.MustCompile(`<a\s+id="[^"]*"\s*/?>\s*</a>|<a\s+id="[^"]*">`)
	manualHTMLTag      = regexp.MustCompile(`<[^>]+>`)
	manualBoldNumPat   = regexp.MustCompile(`^\*\*<?a?\s*(?:id="[^"]*")?>?\s*(\d+)\.?\*\*\s*`)
)

// KnownManuals returns the list of known manual definitions
// Paths are relative to the gospel-library/eng/manual/ directory
func KnownManuals() []ManualDefinition {
	return []ManualDefinition{
		// Teachings of Presidents of the Church
		{Name: "Teachings: Joseph Smith", Path: "teachings-joseph-smith", Type: "teachings"},
		{Name: "Teachings: Brigham Young", Path: "teachings-brigham-young", Type: "teachings"},
		{Name: "Teachings: John Taylor", Path: "teachings-john-taylor", Type: "teachings"},
		{Name: "Teachings: Wilford Woodruff", Path: "teachings-wilford-woodruff", Type: "teachings"},
		{Name: "Teachings: Lorenzo Snow", Path: "teachings-of-presidents-of-the-church-lorenzo-snow", Type: "teachings"},
		{Name: "Teachings: Joseph F. Smith", Path: "teachings-joseph-f-smith", Type: "teachings"},
		{Name: "Teachings: Heber J. Grant", Path: "teachings-heber-j-grant", Type: "teachings"},
		{Name: "Teachings: George Albert Smith", Path: "teachings-george-albert-smith", Type: "teachings"},
		{Name: "Teachings: David O. McKay", Path: "teachings-david-o-mckay", Type: "teachings"},
		{Name: "Teachings: Joseph Fielding Smith", Path: "teachings-of-presidents-of-the-church-joseph-fielding-smith", Type: "teachings"},
		{Name: "Teachings: Harold B. Lee", Path: "teachings-harold-b-lee", Type: "teachings"},
		{Name: "Teachings: Spencer W. Kimball", Path: "teachings-spencer-w-kimball", Type: "teachings"},
		{Name: "Teachings: Ezra Taft Benson", Path: "teachings-of-presidents-of-the-church-ezra-taft-benson", Type: "teachings"},
		{Name: "Teachings: Howard W. Hunter", Path: "teachings-of-presidents-of-the-church-howard-w-hunter", Type: "teachings"},
		{Name: "Teachings: Gordon B. Hinckley", Path: "teachings-of-presidents-of-the-church-gordon-b-hinckley", Type: "teachings"},
		{Name: "Teachings: Thomas S. Monson", Path: "teachings-of-presidents-of-the-church-thomas-s-monson", Type: "teachings"},
		{Name: "Teachings: Russell M. Nelson", Path: "teachings-of-presidents-of-the-church-russell-m-nelson", Type: "teachings"},

		// Come, Follow Me
		{Name: "Come, Follow Me: OT 2026", Path: "come-follow-me-for-home-and-church-old-testament-2026", Type: "cfm"},

		// Teaching in the Savior's Way
		{Name: "Teaching in the Savior's Way", Path: "teaching-in-the-saviors-way-2022", Type: "manual"},
	}
}

// KnownBooks returns known book definitions
// Paths are relative to the workspace books/ directory
func KnownBooks() []ManualDefinition {
	return []ManualDefinition{
		{Name: "Lectures on Faith", Path: "lecture-on-faith", Type: "book"},
	}
}

// ParseManualFile parses a manual/book markdown file
func ParseManualFile(filePath string, manualName string) (*ParsedManualChapter, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	chapter := &ParsedManualChapter{
		ManualName: manualName,
		FilePath:   filePath,
	}

	// Extract chapter number from filename
	base := filepath.Base(filePath)
	base = strings.TrimSuffix(base, ".md")
	chapter.Chapter = extractChapterNumber(base)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Support large files

	var currentSection *ManualSection
	var paragraphBuf strings.Builder
	titleFound := false

	flushParagraph := func() {
		text := strings.TrimSpace(paragraphBuf.String())
		if text != "" {
			cleaned := cleanManualText(text)
			if cleaned != "" && len(cleaned) > 30 { // Skip very short fragments
				chapter.Paragraphs = append(chapter.Paragraphs, cleaned)
				if currentSection != nil {
					if currentSection.Content != "" {
						currentSection.Content += "\n\n"
					}
					currentSection.Content += cleaned
				}
			}
		}
		paragraphBuf.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip audio links
		if manualAudioPattern.MatchString(line) {
			continue
		}

		// Skip image lines
		if manualImagePattern.MatchString(line) {
			continue
		}

		// Skip citation/header lines (publication metadata)
		trimmed := strings.TrimSpace(line)
		if !titleFound && isMetadataLine(trimmed) {
			continue
		}

		// H1 - Chapter title
		if matches := manualH1Pattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			if !titleFound {
				chapter.Title = strings.TrimSpace(matches[1])
				titleFound = true
			}
			continue
		}

		// H2 - Section heading
		if matches := manualH2Pattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			section := ManualSection{
				Heading: cleanManualText(matches[1]),
				Level:   2,
			}
			chapter.Sections = append(chapter.Sections, section)
			currentSection = &chapter.Sections[len(chapter.Sections)-1]
			continue
		}

		// H3 - Sub-section heading
		if matches := manualH3Pattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			section := ManualSection{
				Heading: cleanManualText(matches[1]),
				Level:   3,
			}
			chapter.Sections = append(chapter.Sections, section)
			currentSection = &chapter.Sections[len(chapter.Sections)-1]
			continue
		}

		// Skip horizontal rules
		if trimmed == "---" || trimmed == "***" || trimmed == "* * *" {
			flushParagraph()
			continue
		}

		// Skip footnote definitions
		if strings.HasPrefix(trimmed, "<a id=\"fn-") {
			continue
		}

		// Empty line = paragraph break
		if trimmed == "" {
			flushParagraph()
			continue
		}

		// Clean blockquote markers but keep text
		if strings.HasPrefix(line, "> ") {
			line = strings.TrimPrefix(line, "> ")
		} else if strings.HasPrefix(line, ">") {
			line = strings.TrimPrefix(line, ">")
		}

		// Accumulate paragraph content
		if paragraphBuf.Len() > 0 {
			paragraphBuf.WriteString(" ")
		}
		paragraphBuf.WriteString(line)
	}

	// Flush final paragraph
	flushParagraph()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file: %w", err)
	}

	return chapter, nil
}

// isMetadataLine detects publication citation/metadata lines to skip
func isMetadataLine(line string) bool {
	// Common prefixes in manual metadata
	metadataPatterns := []string{
		"\"Chapter",
		"\u201cChapter", // Smart quotes
		"\"Lesson",
		"\u201cLesson",
		"Teachings of Presidents",
		"Come, Follow Me",
		"Teaching in the Savior",
		"Chapter ", // "Chapter 1" standalone line
	}
	for _, p := range metadataPatterns {
		if strings.HasPrefix(line, p) {
			return true
		}
	}

	// Lines that are just citations like '"Chapter 1," Teachings: Joseph Smith, 26-35'
	if (strings.Contains(line, "Teachings:") || strings.Contains(line, "Teachings of")) &&
		(strings.Contains(line, "(20") || strings.Contains(line, "(19")) {
		return true
	}

	return false
}

// cleanManualText removes markdown artifacts from manual text
func cleanManualText(text string) string {
	// Remove footnote superscripts
	text = manualFootnoteRef.ReplaceAllString(text, "")

	// Remove anchor tags
	text = manualAnchorTag.ReplaceAllString(text, "")

	// Convert markdown links to just text
	text = linkPattern.ReplaceAllString(text, "$1")

	// Remove remaining HTML tags
	text = manualHTMLTag.ReplaceAllString(text, "")

	// Remove bold numbered paragraph markers (e.g., **1.** or **<a id="1"></a>1.**)
	text = manualBoldNumPat.ReplaceAllString(text, "")

	// Clean up bold/italic markers
	text = strings.ReplaceAll(text, "***", "")
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")

	// Remove leftover escape chars
	text = strings.ReplaceAll(text, "\\[", "[")
	text = strings.ReplaceAll(text, "\\]", "]")

	// Clean whitespace
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	return text
}

// extractChapterNumber tries to extract a chapter/lesson number from a filename
func extractChapterNumber(basename string) int {
	basename = strings.ToLower(basename)

	// "chapter-N" pattern
	if strings.HasPrefix(basename, "chapter-") {
		numStr := strings.TrimPrefix(basename, "chapter-")
		if num, err := strconv.Atoi(numStr); err == nil {
			return num
		}
	}

	// Leading number pattern: "01_lecture_1", "01.md", "01-heavenly..."
	numStr := ""
	for _, r := range basename {
		if r >= '0' && r <= '9' {
			numStr += string(r)
		} else {
			break
		}
	}
	if numStr != "" {
		if num, err := strconv.Atoi(numStr); err == nil {
			return num
		}
	}

	return 0
}

// sanitizeID creates a safe ID string
func sanitizeID(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\u2019", "") // Right single quote
	s = strings.ReplaceAll(s, ".", "")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// FindManualFiles finds all .md files in a manual directory (recursive)
func FindManualFiles(basePath string) ([]string, error) {
	var files []string

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Skip index/title/metadata files that don't have real content
		name := strings.ToLower(info.Name())
		skipNames := []string{
			"index.md", "title-page.md", "list-of-visuals.md",
			"translations-and-downloads.md",
		}
		for _, skip := range skipNames {
			if name == skip {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", basePath, err)
	}

	return files, nil
}

// ChunkManualByParagraph creates paragraph-level chunks for manual content
func ChunkManualByParagraph(chapter *ParsedManualChapter) []Chunk {
	var chunks []Chunk

	for i, para := range chapter.Paragraphs {
		// Find which section this paragraph belongs to
		sectionHeading := findSectionForParagraph(chapter, para)

		reference := chapter.ManualName
		if chapter.Title != "" {
			reference += ", " + chapter.Title
		}
		if sectionHeading != "" {
			reference += " > " + sectionHeading
		}

		chunkID := fmt.Sprintf("manual-%s-ch%d-p%d",
			sanitizeID(chapter.ManualName),
			chapter.Chapter,
			i+1,
		)

		chunks = append(chunks, Chunk{
			ID:      chunkID,
			Content: para,
			Metadata: &DocMetadata{
				Source:    SourceManual,
				Layer:     LayerParagraph,
				Book:      chapter.ManualName,
				Chapter:   chapter.Chapter,
				Reference: reference,
				Range:     fmt.Sprintf("p%d", i+1),
				FilePath:  chapter.FilePath,
				Generated: false,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		})
	}

	return chunks
}

// ChunkManualAsSummary creates an LLM summary chunk for manual content
func ChunkManualAsSummary(chapter *ParsedManualChapter, summary *ChapterSummary, model string) Chunk {
	// Build search-optimized content
	content := fmt.Sprintf("KEYWORDS: %s\nSUMMARY: %s",
		strings.Join(summary.Keywords, ", "),
		summary.Summary,
	)
	if summary.KeyVerse != "" {
		content += fmt.Sprintf("\nKEY_QUOTE: %s", summary.KeyVerse)
	}

	reference := chapter.ManualName
	if chapter.Title != "" {
		reference += ", " + chapter.Title
	}

	chunkID := fmt.Sprintf("manual-%s-ch%d-summary",
		sanitizeID(chapter.ManualName),
		chapter.Chapter,
	)

	return Chunk{
		ID:      chunkID,
		Content: content,
		Metadata: &DocMetadata{
			Source:    SourceManual,
			Layer:     LayerSummary,
			Book:      chapter.ManualName,
			Chapter:   chapter.Chapter,
			Reference: reference,
			FilePath:  chapter.FilePath,
			Generated: true,
			Model:     model,
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

// findSectionForParagraph determines which section a paragraph belongs to
func findSectionForParagraph(chapter *ParsedManualChapter, para string) string {
	// Check paragraph content against section content
	prefix := para
	if len(prefix) > 60 {
		prefix = prefix[:60]
	}

	for i := len(chapter.Sections) - 1; i >= 0; i-- {
		if strings.Contains(chapter.Sections[i].Content, prefix) {
			return chapter.Sections[i].Heading
		}
	}
	return ""
}

// GetManualChapterContent returns all paragraph text joined for summarization
func GetManualChapterContent(chapter *ParsedManualChapter) string {
	var content strings.Builder
	for _, section := range chapter.Sections {
		if section.Heading != "" {
			content.WriteString("## " + section.Heading + "\n\n")
		}
		if section.Content != "" {
			content.WriteString(section.Content + "\n\n")
		}
	}
	// If no sections, just join paragraphs
	if len(chapter.Sections) == 0 {
		for _, para := range chapter.Paragraphs {
			content.WriteString(para + "\n\n")
		}
	}
	return content.String()
}
