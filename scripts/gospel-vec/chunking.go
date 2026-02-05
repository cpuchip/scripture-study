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
	"unicode"
)

// Scripture parsing patterns
var (
	versePattern    = regexp.MustCompile(`^\*\*(\d+)\.\*\*\s*(.+)`)
	chapterPattern  = regexp.MustCompile(`^#\s+(.+)`)
	footnotePattern = regexp.MustCompile(`<sup>\[.*?\]</sup>`)
	linkPattern     = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	superscriptRef  = regexp.MustCompile(`<sup>[^<]+</sup>`)
)

// ParsedChapter contains all parsed data from a chapter file
type ParsedChapter struct {
	Book     string
	Chapter  int
	Title    string
	Verses   []ParsedVerse
	FilePath string
}

// ParsedVerse contains a single verse
type ParsedVerse struct {
	Number int
	Text   string
}

// ParseChapterFile parses a markdown scripture file
func ParseChapterFile(filePath string) (*ParsedChapter, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	chapter := &ParsedChapter{
		FilePath: filePath,
	}

	// Extract book and chapter from path
	// e.g., .../bofm/1-ne/3.md -> "1 Nephi", 3
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	bookDir := filepath.Base(dir)

	chapter.Book = formatBookName(bookDir)
	if chNum, err := strconv.Atoi(strings.TrimSuffix(base, ".md")); err == nil {
		chapter.Chapter = chNum
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for chapter title
		if matches := chapterPattern.FindStringSubmatch(line); matches != nil {
			chapter.Title = matches[1]
			continue
		}

		// Check for verse
		if matches := versePattern.FindStringSubmatch(line); matches != nil {
			verseNum, _ := strconv.Atoi(matches[1])
			verseText := cleanVerseText(matches[2])

			if verseText != "" {
				chapter.Verses = append(chapter.Verses, ParsedVerse{
					Number: verseNum,
					Text:   verseText,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file: %w", err)
	}

	return chapter, nil
}

// cleanVerseText removes markdown artifacts and footnotes
func cleanVerseText(text string) string {
	// Remove footnote references
	text = footnotePattern.ReplaceAllString(text, "")
	text = superscriptRef.ReplaceAllString(text, "")

	// Convert markdown links to just the text
	text = linkPattern.ReplaceAllString(text, "$1")

	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	return text
}

// ChunkByVerse creates one chunk per verse
func ChunkByVerse(chapter *ParsedChapter, source Source) []Chunk {
	chunks := make([]Chunk, 0, len(chapter.Verses))
	timestamp := time.Now().Format(time.RFC3339)

	for _, verse := range chapter.Verses {
		reference := fmt.Sprintf("%s %d:%d", chapter.Book, chapter.Chapter, verse.Number)
		id := fmt.Sprintf("%s-%s-%d-%d", source, slugify(chapter.Book), chapter.Chapter, verse.Number)

		chunks = append(chunks, Chunk{
			ID:      id,
			Content: verse.Text,
			Metadata: &DocMetadata{
				Source:    source,
				Layer:     LayerVerse,
				Book:      chapter.Book,
				Chapter:   chapter.Chapter,
				Reference: reference,
				Range:     fmt.Sprintf("%d", verse.Number),
				FilePath:  chapter.FilePath,
				Generated: false,
				Timestamp: timestamp,
			},
		})
	}

	return chunks
}

// ChunkByParagraph creates chunks of 3-5 verses (natural paragraph breaks)
func ChunkByParagraph(chapter *ParsedChapter, source Source) []Chunk {
	if len(chapter.Verses) == 0 {
		return nil
	}

	chunks := make([]Chunk, 0)
	timestamp := time.Now().Format(time.RFC3339)

	// Simple paragraph chunking: 4 verses per chunk with 1 verse overlap
	chunkSize := 4
	overlap := 1

	for i := 0; i < len(chapter.Verses); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(chapter.Verses) {
			end = len(chapter.Verses)
		}

		// Build chunk content
		var content strings.Builder
		verseNums := make([]int, 0)
		for j := i; j < end; j++ {
			if content.Len() > 0 {
				content.WriteString(" ")
			}
			content.WriteString(fmt.Sprintf("(%d) %s", chapter.Verses[j].Number, chapter.Verses[j].Text))
			verseNums = append(verseNums, chapter.Verses[j].Number)
		}

		startVerse := verseNums[0]
		endVerse := verseNums[len(verseNums)-1]
		reference := fmt.Sprintf("%s %d:%d-%d", chapter.Book, chapter.Chapter, startVerse, endVerse)
		rangeStr := fmt.Sprintf("%d-%d", startVerse, endVerse)
		id := fmt.Sprintf("%s-%s-%d-p%d-%d", source, slugify(chapter.Book), chapter.Chapter, startVerse, endVerse)

		chunks = append(chunks, Chunk{
			ID:      id,
			Content: content.String(),
			Metadata: &DocMetadata{
				Source:    source,
				Layer:     LayerParagraph,
				Book:      chapter.Book,
				Chapter:   chapter.Chapter,
				Reference: reference,
				Range:     rangeStr,
				FilePath:  chapter.FilePath,
				Generated: false,
				Timestamp: timestamp,
			},
		})

		// Stop if we've reached the end
		if end >= len(chapter.Verses) {
			break
		}
	}

	return chunks
}

// ChunkAsChapterSummary creates a single chunk for the whole chapter (for LLM summary)
func ChunkAsChapterSummary(chapter *ParsedChapter, source Source, summary *ChapterSummary, model string) Chunk {
	timestamp := time.Now().Format(time.RFC3339)
	reference := fmt.Sprintf("%s %d", chapter.Book, chapter.Chapter)
	id := fmt.Sprintf("%s-%s-%d-summary", source, slugify(chapter.Book), chapter.Chapter)

	// Create searchable content from structured summary
	var content strings.Builder
	if len(summary.Keywords) > 0 {
		content.WriteString("Keywords: ")
		content.WriteString(strings.Join(summary.Keywords, ", "))
		content.WriteString("\n\n")
	}
	if summary.Summary != "" {
		content.WriteString(summary.Summary)
	}
	if summary.KeyVerse != "" {
		content.WriteString("\n\nKey verse: ")
		content.WriteString(summary.KeyVerse)
	}

	return Chunk{
		ID:      id,
		Content: content.String(),
		Metadata: &DocMetadata{
			Source:    source,
			Layer:     LayerSummary,
			Book:      chapter.Book,
			Chapter:   chapter.Chapter,
			Reference: reference,
			Range:     fmt.Sprintf("1-%d", len(chapter.Verses)),
			FilePath:  chapter.FilePath,
			Generated: true,
			Model:     model,
			Timestamp: timestamp,
		},
	}
}

// ChunkAsTheme creates a chunk for a detected theme
func ChunkAsTheme(chapter *ParsedChapter, source Source, theme ThemeRange, model string) Chunk {
	timestamp := time.Now().Format(time.RFC3339)
	reference := fmt.Sprintf("%s %d:%s", chapter.Book, chapter.Chapter, theme.Range)
	id := fmt.Sprintf("%s-%s-%d-theme-%s", source, slugify(chapter.Book), chapter.Chapter, slugify(theme.Range))

	return Chunk{
		ID:      id,
		Content: theme.Theme,
		Metadata: &DocMetadata{
			Source:    source,
			Layer:     LayerTheme,
			Book:      chapter.Book,
			Chapter:   chapter.Chapter,
			Reference: reference,
			Range:     theme.Range,
			FilePath:  chapter.FilePath,
			Generated: true,
			Model:     model,
			Timestamp: timestamp,
		},
	}
}

// GetFullChapterContent returns the full text of a chapter for summarization
func GetFullChapterContent(chapter *ParsedChapter) string {
	var content strings.Builder
	for _, verse := range chapter.Verses {
		content.WriteString(fmt.Sprintf("%d. %s\n", verse.Number, verse.Text))
	}
	return content.String()
}

// GetVerseTexts returns just the verse texts as a slice
func GetVerseTexts(chapter *ParsedChapter) []string {
	texts := make([]string, len(chapter.Verses))
	for i, verse := range chapter.Verses {
		texts[i] = verse.Text
	}
	return texts
}

// bookNameMap maps directory names to canonical display names
var bookNameMap = map[string]string{
	// Book of Mormon
	"1-ne":   "1 Nephi",
	"2-ne":   "2 Nephi",
	"jacob":  "Jacob",
	"enos":   "Enos",
	"jarom":  "Jarom",
	"omni":   "Omni",
	"w-of-m": "Words of Mormon",
	"mosiah": "Mosiah",
	"alma":   "Alma",
	"hel":    "Helaman",
	"3-ne":   "3 Nephi",
	"4-ne":   "4 Nephi",
	"morm":   "Mormon",
	"ether":  "Ether",
	"moro":   "Moroni",
	// D&C
	"dc": "D&C",
	// Pearl of Great Price
	"moses":  "Moses",
	"abr":    "Abraham",
	"js-m":   "Joseph Smith—Matthew",
	"js-h":   "Joseph Smith—History",
	"a-of-f": "Articles of Faith",
	// Old Testament
	"gen":   "Genesis",
	"ex":    "Exodus",
	"lev":   "Leviticus",
	"num":   "Numbers",
	"deut":  "Deuteronomy",
	"josh":  "Joshua",
	"judg":  "Judges",
	"ruth":  "Ruth",
	"1-sam": "1 Samuel",
	"2-sam": "2 Samuel",
	"1-kgs": "1 Kings",
	"2-kgs": "2 Kings",
	"1-chr": "1 Chronicles",
	"2-chr": "2 Chronicles",
	"ezra":  "Ezra",
	"neh":   "Nehemiah",
	"esth":  "Esther",
	"job":   "Job",
	"ps":    "Psalms",
	"prov":  "Proverbs",
	"eccl":  "Ecclesiastes",
	"song":  "Song of Solomon",
	"isa":   "Isaiah",
	"jer":   "Jeremiah",
	"lam":   "Lamentations",
	"ezek":  "Ezekiel",
	"dan":   "Daniel",
	"hosea": "Hosea",
	"joel":  "Joel",
	"amos":  "Amos",
	"obad":  "Obadiah",
	"jonah": "Jonah",
	"micah": "Micah",
	"nahum": "Nahum",
	"hab":   "Habakkuk",
	"zeph":  "Zephaniah",
	"hag":   "Haggai",
	"zech":  "Zechariah",
	"mal":   "Malachi",
	// New Testament
	"matt":   "Matthew",
	"mark":   "Mark",
	"luke":   "Luke",
	"john":   "John",
	"acts":   "Acts",
	"rom":    "Romans",
	"1-cor":  "1 Corinthians",
	"2-cor":  "2 Corinthians",
	"gal":    "Galatians",
	"eph":    "Ephesians",
	"philip": "Philippians",
	"col":    "Colossians",
	"1-thes": "1 Thessalonians",
	"2-thes": "2 Thessalonians",
	"1-tim":  "1 Timothy",
	"2-tim":  "2 Timothy",
	"titus":  "Titus",
	"philem": "Philemon",
	"heb":    "Hebrews",
	"james":  "James",
	"1-pet":  "1 Peter",
	"2-pet":  "2 Peter",
	"1-jn":   "1 John",
	"2-jn":   "2 John",
	"3-jn":   "3 John",
	"jude":   "Jude",
	"rev":    "Revelation",
}

// bookAliasMap maps common aliases to canonical display names
// This allows users to type "dc", "D&C", "d&c", "doctrine and covenants" etc.
var bookAliasMap = map[string]string{
	// D&C aliases
	"dc":                      "D&C",
	"d&c":                     "D&C",
	"doctrine and covenants":  "D&C",
	"doctrine & covenants":    "D&C",
	"doctrinne and covenants": "D&C", // common typo
	"dc-testament":            "D&C",
	// Book of Mormon aliases
	"1 nephi":         "1 Nephi",
	"1nephi":          "1 Nephi",
	"1ne":             "1 Nephi",
	"1-ne":            "1 Nephi",
	"i nephi":         "1 Nephi",
	"2 nephi":         "2 Nephi",
	"2nephi":          "2 Nephi",
	"2ne":             "2 Nephi",
	"2-ne":            "2 Nephi",
	"ii nephi":        "2 Nephi",
	"3 nephi":         "3 Nephi",
	"3nephi":          "3 Nephi",
	"3ne":             "3 Nephi",
	"3-ne":            "3 Nephi",
	"iii nephi":       "3 Nephi",
	"4 nephi":         "4 Nephi",
	"4nephi":          "4 Nephi",
	"4ne":             "4 Nephi",
	"4-ne":            "4 Nephi",
	"iv nephi":        "4 Nephi",
	"words of mormon": "Words of Mormon",
	"w-of-m":          "Words of Mormon",
	"wofm":            "Words of Mormon",
	"helaman":         "Helaman",
	"hel":             "Helaman",
	"mormon":          "Mormon",
	"morm":            "Mormon",
	"moroni":          "Moroni",
	"moro":            "Moroni",
	// Pearl of Great Price aliases
	"abraham":              "Abraham",
	"abr":                  "Abraham",
	"joseph smith—matthew": "Joseph Smith—Matthew",
	"joseph smith-matthew": "Joseph Smith—Matthew",
	"joseph smith matthew": "Joseph Smith—Matthew",
	"js-m":                 "Joseph Smith—Matthew",
	"jsm":                  "Joseph Smith—Matthew",
	"js matthew":           "Joseph Smith—Matthew",
	"joseph smith—history": "Joseph Smith—History",
	"joseph smith-history": "Joseph Smith—History",
	"joseph smith history": "Joseph Smith—History",
	"js-h":                 "Joseph Smith—History",
	"jsh":                  "Joseph Smith—History",
	"js history":           "Joseph Smith—History",
	"articles of faith":    "Articles of Faith",
	"a-of-f":               "Articles of Faith",
	"aoff":                 "Articles of Faith",
	// Old Testament common aliases
	"genesis":     "Genesis",
	"gen":         "Genesis",
	"exodus":      "Exodus",
	"ex":          "Exodus",
	"leviticus":   "Leviticus",
	"lev":         "Leviticus",
	"numbers":     "Numbers",
	"num":         "Numbers",
	"deuteronomy": "Deuteronomy",
	"deut":        "Deuteronomy",
	"psalms":      "Psalms",
	"psalm":       "Psalms",
	"ps":          "Psalms",
	"proverbs":    "Proverbs",
	"prov":        "Proverbs",
	"isaiah":      "Isaiah",
	"isa":         "Isaiah",
	"jeremiah":    "Jeremiah",
	"jer":         "Jeremiah",
	// New Testament common aliases
	"matthew":        "Matthew",
	"matt":           "Matthew",
	"mt":             "Matthew",
	"romans":         "Romans",
	"rom":            "Romans",
	"1 corinthians":  "1 Corinthians",
	"1 cor":          "1 Corinthians",
	"1cor":           "1 Corinthians",
	"1-cor":          "1 Corinthians",
	"i corinthians":  "1 Corinthians",
	"2 corinthians":  "2 Corinthians",
	"2 cor":          "2 Corinthians",
	"2cor":           "2 Corinthians",
	"2-cor":          "2 Corinthians",
	"ii corinthians": "2 Corinthians",
	"revelation":     "Revelation",
	"rev":            "Revelation",
	"revelations":    "Revelation", // common mistake
}

// NormalizeBookName converts various user inputs to canonical book names
// Accepts: "dc", "D&C", "d&c", "1 Nephi", "1-ne", "1nephi", etc.
func NormalizeBookName(input string) string {
	// Lowercase and trim for matching
	normalized := strings.ToLower(strings.TrimSpace(input))

	// Check alias map first (handles most variations)
	if canonical, ok := bookAliasMap[normalized]; ok {
		return canonical
	}

	// Check if already a canonical name (case-insensitive)
	for _, canonical := range bookNameMap {
		if strings.EqualFold(input, canonical) {
			return canonical
		}
	}

	// Check dirName map
	if canonical, ok := bookNameMap[normalized]; ok {
		return canonical
	}

	// Return original input (will fail match but provides context in error)
	return input
}

// GetIndexedBooks returns a list of all canonical book names currently supported
func GetIndexedBooks() []string {
	// Return unique canonical names sorted by scripture order
	seen := make(map[string]bool)
	var books []string

	// Add in rough scripture order
	orderedBooks := []string{
		// Book of Mormon
		"1 Nephi", "2 Nephi", "Jacob", "Enos", "Jarom", "Omni", "Words of Mormon",
		"Mosiah", "Alma", "Helaman", "3 Nephi", "4 Nephi", "Mormon", "Ether", "Moroni",
		// Doctrine & Covenants
		"D&C",
		// Pearl of Great Price
		"Moses", "Abraham", "Joseph Smith—Matthew", "Joseph Smith—History", "Articles of Faith",
		// Old Testament (selection)
		"Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy",
		"Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel", "1 Kings", "2 Kings",
		"1 Chronicles", "2 Chronicles", "Ezra", "Nehemiah", "Esther", "Job",
		"Psalms", "Proverbs", "Ecclesiastes", "Song of Solomon",
		"Isaiah", "Jeremiah", "Lamentations", "Ezekiel", "Daniel",
		"Hosea", "Joel", "Amos", "Obadiah", "Jonah", "Micah", "Nahum",
		"Habakkuk", "Zephaniah", "Haggai", "Zechariah", "Malachi",
		// New Testament
		"Matthew", "Mark", "Luke", "John", "Acts",
		"Romans", "1 Corinthians", "2 Corinthians", "Galatians", "Ephesians",
		"Philippians", "Colossians", "1 Thessalonians", "2 Thessalonians",
		"1 Timothy", "2 Timothy", "Titus", "Philemon", "Hebrews",
		"James", "1 Peter", "2 Peter", "1 John", "2 John", "3 John", "Jude", "Revelation",
	}

	for _, book := range orderedBooks {
		if !seen[book] {
			seen[book] = true
			books = append(books, book)
		}
	}
	return books
}

// formatBookName converts directory names to readable book names
func formatBookName(dirName string) string {
	if name, ok := bookNameMap[dirName]; ok {
		return name
	}

	// Default: capitalize and replace hyphens
	words := strings.Split(dirName, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// slugify converts a string to a URL-safe slug
func slugify(s string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' {
			result.WriteRune('-')
		}
	}
	return result.String()
}

// FindScriptureFiles finds all chapter markdown files in a scriptures directory
func FindScriptureFiles(basePath string, volumes ...string) ([]string, error) {
	var files []string

	// Default volumes if none specified
	if len(volumes) == 0 {
		volumes = []string{"bofm", "dc-testament/dc", "pgp"}
	}

	for _, volume := range volumes {
		volumePath := filepath.Join(basePath, volume)

		err := filepath.Walk(volumePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if info.IsDir() {
				return nil
			}

			// Only .md files that are numbered (chapters)
			base := filepath.Base(path)
			if !strings.HasSuffix(base, ".md") {
				return nil
			}

			// Check if filename is a number (chapter file)
			name := strings.TrimSuffix(base, ".md")
			if _, err := strconv.Atoi(name); err == nil {
				files = append(files, path)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("walking %s: %w", volume, err)
		}
	}

	return files, nil
}
