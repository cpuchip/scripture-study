package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetParams are the parameters for gospel_get.
type GetParams struct {
	Reference      string `json:"reference"`
	Context        int    `json:"context"`
	IncludeChapter bool   `json:"include_chapter"`
	FilePath       string `json:"file_path"`
}

// Get retrieves specific gospel content by reference or path.
func (t *Tools) Get(args json.RawMessage) (*GetResponse, error) {
	var params GetParams
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("parsing params: %w", err)
	}

	// Context defaults to 0 (no surrounding verses).
	// Pass context=3 (or any positive int) when you want surrounding context.
	// This keeps output lean for document building; use higher values for study.

	// Determine what to fetch
	if params.FilePath != "" {
		return t.getByFilePath(params)
	}

	if params.Reference != "" {
		return t.getByReference(params)
	}

	return nil, fmt.Errorf("either reference or file_path is required")
}

func (t *Tools) getByFilePath(params GetParams) (*GetResponse, error) {
	// Try scriptures first
	row := t.db.QueryRow(`
		SELECT id, volume, book, chapter, verse, text, file_path, source_url
		FROM scriptures WHERE file_path = ? LIMIT 1
	`, params.FilePath)

	var id int
	var volume, book, text, filePath, sourceURL string
	var chapter, verse int

	if err := row.Scan(&id, &volume, &book, &chapter, &verse, &text, &filePath, &sourceURL); err == nil {
		ref := formatScriptureRef(volume, book, chapter, verse)
		return &GetResponse{
			Reference:         ref,
			Title:             formatChapterTitle(volume, book, chapter),
			Content:           text,
			FilePath:          filePath,
			MarkdownLink:      generateMarkdownLink(ref, filePath),
			SourceURL:         sourceURL,
			SourceType:        "scripture",
			RelatedReferences: t.getCrossReferences(volume, book, chapter, verse),
		}, nil
	}

	// Try talks
	row = t.db.QueryRow(`
		SELECT id, year, month, speaker, title, content, file_path, source_url
		FROM talks WHERE file_path = ? LIMIT 1
	`, params.FilePath)

	var year, month int
	var speaker, title, content string

	if err := row.Scan(&id, &year, &month, &speaker, &title, &content, &filePath, &sourceURL); err == nil {
		return &GetResponse{
			Reference:    fmt.Sprintf("%s, %s %d", speaker, monthName(month), year),
			Title:        title,
			Content:      content,
			FilePath:     filePath,
			MarkdownLink: generateTalkMarkdownLink(speaker, title, filePath),
			SourceURL:    sourceURL,
			SourceType:   "conference",
		}, nil
	}

	// Try manuals
	row = t.db.QueryRow(`
		SELECT id, content_type, title, content, file_path, source_url
		FROM manuals WHERE file_path = ? LIMIT 1
	`, params.FilePath)

	var contentType string

	if err := row.Scan(&id, &contentType, &title, &content, &filePath, &sourceURL); err == nil {
		return &GetResponse{
			Reference:    title,
			Title:        title,
			Content:      content,
			FilePath:     filePath,
			MarkdownLink: generateMarkdownLink(title, filePath),
			SourceURL:    sourceURL,
			SourceType:   contentType,
		}, nil
	}

	return nil, fmt.Errorf("not found: %s", params.FilePath)
}

func (t *Tools) getByReference(params GetParams) (*GetResponse, error) {
	// Parse the reference
	parsed := parseReference(params.Reference)

	switch parsed.Type {
	case "scripture":
		return t.getScripture(parsed, params)
	case "talk":
		return t.getTalk(parsed, params)
	default:
		// Try searching for it
		return t.searchForReference(params)
	}
}

func (t *Tools) getScripture(ref parsedRef, params GetParams) (*GetResponse, error) {
	if ref.Verse > 0 && ref.EndVerse > 0 {
		// Get verse range
		return t.getScriptureRange(ref, params)
	}

	if ref.Verse > 0 {
		// Get specific verse
		row := t.db.QueryRow(`
			SELECT id, volume, book, chapter, verse, text, file_path, source_url
			FROM scriptures 
			WHERE book = ? AND chapter = ? AND verse = ?
			LIMIT 1
		`, ref.Book, ref.Chapter, ref.Verse)

		var id int
		var volume, book, text, filePath, sourceURL string
		var chapter, verse int

		if err := row.Scan(&id, &volume, &book, &chapter, &verse, &text, &filePath, &sourceURL); err != nil {
			return nil, fmt.Errorf("scripture not found: %s", params.Reference)
		}

		response := &GetResponse{
			Reference:         formatScriptureRef(volume, book, chapter, verse),
			Title:             formatChapterTitle(volume, book, chapter),
			Content:           text,
			FilePath:          filePath,
			MarkdownLink:      generateMarkdownLink(formatScriptureRef(volume, book, chapter, verse), filePath),
			SourceURL:         sourceURL,
			SourceType:        "scripture",
			RelatedReferences: t.getCrossReferences(volume, book, chapter, verse),
		}

		// Add context
		if params.Context > 0 {
			response.ContextBefore = t.getVerseContextStructured(volume, book, chapter, verse, -params.Context)
			response.ContextAfter = t.getVerseContextStructured(volume, book, chapter, verse, params.Context)
		}

		// Optionally include full chapter
		if params.IncludeChapter {
			response.ChapterContent = t.getChapterContent(volume, book, chapter)
		}

		return response, nil
	}

	// Get full chapter
	rows, err := t.db.Query(`
		SELECT verse, text FROM scriptures
		WHERE book = ? AND chapter = ?
		ORDER BY verse
	`, ref.Book, ref.Chapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []string
	var filePath, sourceURL, volume string

	for rows.Next() {
		var v int
		var text string
		if err := rows.Scan(&v, &text); err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", v, text))
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("chapter not found: %s", params.Reference)
	}

	// Get metadata from first verse
	row := t.db.QueryRow(`
		SELECT volume, file_path, source_url FROM scriptures
		WHERE book = ? AND chapter = ? LIMIT 1
	`, ref.Book, ref.Chapter)
	row.Scan(&volume, &filePath, &sourceURL)

	chapterTitle := formatChapterTitle(volume, ref.Book, ref.Chapter)
	return &GetResponse{
		Reference:    chapterTitle,
		Title:        chapterTitle,
		Content:      strings.Join(lines, "\n\n"),
		FilePath:     filePath,
		MarkdownLink: generateMarkdownLink(chapterTitle, filePath),
		SourceURL:    sourceURL,
		SourceType:   "scripture",
	}, nil
}

func (t *Tools) getScriptureRange(ref parsedRef, params GetParams) (*GetResponse, error) {
	rows, err := t.db.Query(`
		SELECT id, volume, book, chapter, verse, text, file_path, source_url
		FROM scriptures
		WHERE book = ? AND chapter = ? AND verse >= ? AND verse <= ?
		ORDER BY verse
	`, ref.Book, ref.Chapter, ref.Verse, ref.EndVerse)
	if err != nil {
		return nil, fmt.Errorf("scripture range not found: %s", params.Reference)
	}
	defer rows.Close()

	var verses []VerseContext
	var volume, book, filePath, sourceURL string

	for rows.Next() {
		var id, chapter, verse int
		var text string
		if err := rows.Scan(&id, &volume, &book, &chapter, &verse, &text, &filePath, &sourceURL); err != nil {
			continue
		}
		verses = append(verses, VerseContext{Verse: verse, Text: text})
	}

	if len(verses) == 0 {
		return nil, fmt.Errorf("scripture range not found: %s", params.Reference)
	}

	// Build combined content
	var lines []string
	for _, v := range verses {
		lines = append(lines, fmt.Sprintf("%d. %s", v.Verse, v.Text))
	}

	// Collect cross-references for all verses in the range
	var allRefs []RelatedReference
	for _, v := range verses {
		refs := t.getCrossReferences(volume, book, ref.Chapter, v.Verse)
		allRefs = append(allRefs, refs...)
	}

	rangeRef := fmt.Sprintf("%s %d:%d-%d", formatBookName(volume, ref.Book), ref.Chapter, ref.Verse, ref.EndVerse)
	response := &GetResponse{
		Reference:         rangeRef,
		Title:             formatChapterTitle(volume, ref.Book, ref.Chapter),
		Content:           strings.Join(lines, "\n\n"),
		FilePath:          filePath,
		MarkdownLink:      generateMarkdownLink(rangeRef, filePath),
		SourceURL:         sourceURL,
		SourceType:        "scripture",
		RelatedReferences: allRefs,
	}

	// Add context around the range
	if params.Context > 0 {
		response.ContextBefore = t.getVerseContextStructured(volume, book, ref.Chapter, ref.Verse, -params.Context)
		response.ContextAfter = t.getVerseContextStructured(volume, book, ref.Chapter, ref.EndVerse, params.Context)
	}

	return response, nil
}

func (t *Tools) getTalk(ref parsedRef, params GetParams) (*GetResponse, error) {
	// Search by speaker name
	row := t.db.QueryRow(`
		SELECT id, year, month, speaker, title, content, file_path, source_url
		FROM talks WHERE speaker LIKE ? ORDER BY year DESC, month DESC LIMIT 1
	`, "%"+ref.Speaker+"%")

	var id, year, month int
	var speaker, title, content, filePath, sourceURL string

	if err := row.Scan(&id, &year, &month, &speaker, &title, &content, &filePath, &sourceURL); err != nil {
		return nil, fmt.Errorf("talk not found: %s", params.Reference)
	}

	return &GetResponse{
		Reference:    fmt.Sprintf("%s, %s %d", speaker, monthName(month), year),
		Title:        title,
		Content:      content,
		FilePath:     filePath,
		MarkdownLink: generateTalkMarkdownLink(speaker, title, filePath),
		SourceURL:    sourceURL,
		SourceType:   "conference",
	}, nil
}

func (t *Tools) searchForReference(params GetParams) (*GetResponse, error) {
	// Search in all content types
	searchParams := SearchParams{
		Query:          params.Reference,
		Limit:          1,
		IncludeContent: true,
	}

	argsBytes, _ := json.Marshal(searchParams)
	result, err := t.Search(argsBytes)
	if err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("not found: %s", params.Reference)
	}

	r := result.Results[0]
	return &GetResponse{
		Reference:         r.Reference,
		Title:             r.Title,
		Content:           r.Content,
		FilePath:          r.FilePath,
		MarkdownLink:      r.MarkdownLink,
		SourceURL:         r.SourceURL,
		SourceType:        r.SourceType,
		RelatedReferences: r.RelatedReferences,
	}, nil
}

func (t *Tools) getVerseContextStructured(volume, book string, chapter, verse, delta int) []VerseContext {
	var results []VerseContext

	if delta == 0 {
		return results
	}

	var startVerse, endVerse int
	if delta < 0 {
		startVerse = verse + delta
		if startVerse < 1 {
			startVerse = 1
		}
		endVerse = verse - 1
	} else {
		startVerse = verse + 1
		endVerse = verse + delta
	}

	rows, err := t.db.Query(`
		SELECT verse, text FROM scriptures 
		WHERE volume = ? AND book = ? AND chapter = ? AND verse >= ? AND verse <= ?
		ORDER BY verse ASC
	`, volume, book, chapter, startVerse, endVerse)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var v int
		var text string
		if err := rows.Scan(&v, &text); err != nil {
			continue
		}
		results = append(results, VerseContext{Verse: v, Text: text})
	}

	return results
}

func (t *Tools) getChapterContent(volume, book string, chapter int) string {
	rows, err := t.db.Query(`
		SELECT verse, text FROM scriptures
		WHERE volume = ? AND book = ? AND chapter = ?
		ORDER BY verse
	`, volume, book, chapter)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var v int
		var text string
		if err := rows.Scan(&v, &text); err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", v, text))
	}

	return strings.Join(lines, "\n\n")
}

// parsedRef represents a parsed scripture or talk reference.
type parsedRef struct {
	Type     string // "scripture" or "talk"
	Volume   string
	Book     string
	Chapter  int
	Verse    int
	EndVerse int // For ranges like 24-30; 0 means single verse
	Speaker  string
}

func parseReference(ref string) parsedRef {
	ref = strings.TrimSpace(ref)
	ref = strings.ToLower(ref)

	// Try to match scripture patterns
	// e.g., "1 nephi 3:7", "d&c 93:36", "moses 3:5"

	// Map common variations
	ref = strings.ReplaceAll(ref, "doctrine and covenants", "dc")
	ref = strings.ReplaceAll(ref, "d&c", "dc")

	// Check for verse reference pattern: "book chapter:verse"
	parts := strings.Fields(ref)
	if len(parts) >= 2 {
		// Last part might be "chapter:verse" or just "chapter"
		lastPart := parts[len(parts)-1]
		bookParts := parts[:len(parts)-1]

		if colonIdx := strings.Index(lastPart, ":"); colonIdx > 0 {
			// Has verse — parse chapter:verse or chapter:verse-endverse
			chapterStr := lastPart[:colonIdx]
			verseStr := lastPart[colonIdx+1:]

			var chapter int
			fmt.Sscanf(chapterStr, "%d", &chapter)

			var verse, endVerse int
			if dashIdx := strings.Index(verseStr, "-"); dashIdx > 0 {
				// Range: "24-30"
				fmt.Sscanf(verseStr[:dashIdx], "%d", &verse)
				fmt.Sscanf(verseStr[dashIdx+1:], "%d", &endVerse)
			} else {
				// Single verse: "24"
				fmt.Sscanf(verseStr, "%d", &verse)
			}

			book := normalizeBookName(strings.Join(bookParts, " "))
			if book != "" {
				return parsedRef{
					Type:     "scripture",
					Book:     book,
					Chapter:  chapter,
					Verse:    verse,
					EndVerse: endVerse,
				}
			}
		} else {
			// Just chapter
			var chapter int
			fmt.Sscanf(lastPart, "%d", &chapter)

			if chapter > 0 {
				book := normalizeBookName(strings.Join(bookParts, " "))
				if book != "" {
					return parsedRef{
						Type:    "scripture",
						Book:    book,
						Chapter: chapter,
					}
				}
			}
		}
	}

	// Might be a talk reference with speaker name
	if strings.Contains(ref, ",") || strings.Contains(ref, "elder") ||
		strings.Contains(ref, "president") || strings.Contains(ref, "sister") {
		// Extract speaker name
		speaker := ref
		speaker = strings.TrimPrefix(speaker, "elder ")
		speaker = strings.TrimPrefix(speaker, "president ")
		speaker = strings.TrimPrefix(speaker, "sister ")

		return parsedRef{
			Type:    "talk",
			Speaker: speaker,
		}
	}

	return parsedRef{Type: "unknown"}
}

func normalizeBookName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	// Map full names and variations to abbreviations
	nameMap := map[string]string{
		"genesis": "gen", "exodus": "ex", "leviticus": "lev", "numbers": "num",
		"deuteronomy": "deut", "joshua": "josh", "judges": "judg", "ruth": "ruth",
		"1 samuel": "1-sam", "2 samuel": "2-sam", "1 kings": "1-kgs", "2 kings": "2-kgs",
		"1 chronicles": "1-chr", "2 chronicles": "2-chr", "ezra": "ezra", "nehemiah": "neh",
		"esther": "esth", "job": "job", "psalms": "ps", "psalm": "ps", "proverbs": "prov",
		"ecclesiastes": "eccl", "song of solomon": "song", "isaiah": "isa", "jeremiah": "jer",
		"lamentations": "lam", "ezekiel": "ezek", "daniel": "dan", "hosea": "hosea",
		"joel": "joel", "amos": "amos", "obadiah": "obad", "jonah": "jonah", "micah": "micah",
		"nahum": "nahum", "habakkuk": "hab", "zephaniah": "zeph", "haggai": "hag",
		"zechariah": "zech", "malachi": "mal",
		// New Testament
		"matthew": "matt", "mark": "mark", "luke": "luke", "john": "john", "acts": "acts",
		"romans": "rom", "1 corinthians": "1-cor", "2 corinthians": "2-cor",
		"galatians": "gal", "ephesians": "eph", "philippians": "philip",
		"colossians": "col", "1 thessalonians": "1-thes", "2 thessalonians": "2-thes",
		"1 timothy": "1-tim", "2 timothy": "2-tim", "titus": "titus", "philemon": "philem",
		"hebrews": "heb", "james": "james", "1 peter": "1-pet", "2 peter": "2-pet",
		"1 john": "1-jn", "2 john": "2-jn", "3 john": "3-jn", "jude": "jude",
		"revelation": "rev", "revelations": "rev",
		// Book of Mormon
		"1 nephi": "1-ne", "2 nephi": "2-ne", "jacob": "jacob", "enos": "enos",
		"jarom": "jarom", "omni": "omni", "words of mormon": "w-of-m",
		"mosiah": "mosiah", "alma": "alma", "helaman": "hel",
		"3 nephi": "3-ne", "4 nephi": "4-ne", "mormon": "morm",
		"ether": "ether", "moroni": "moro",
		// D&C
		"dc": "dc", "d&c": "dc", "doctrine and covenants": "dc",
		// Pearl of Great Price
		"moses": "moses", "abraham": "abr", "js matthew": "js-m", "js history": "js-h",
		"articles of faith": "a-of-f",
	}

	if abbr, ok := nameMap[name]; ok {
		return abbr
	}

	// Already an abbreviation?
	for _, abbr := range nameMap {
		if name == abbr {
			return abbr
		}
	}

	return ""
}
