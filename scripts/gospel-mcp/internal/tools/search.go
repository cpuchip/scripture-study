package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SearchParams are the parameters for gospel_search.
type SearchParams struct {
	Query          string `json:"query"`
	Source         string `json:"source"`
	Path           string `json:"path"`
	Limit          int    `json:"limit"`
	Context        int    `json:"context"`
	IncludeContent bool   `json:"include_content"`
}

// Search performs a full-text search across gospel content.
func (t *Tools) Search(args json.RawMessage) (*SearchResponse, error) {
	var params SearchParams
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("parsing params: %w", err)
	}

	if params.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Set defaults
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}
	if params.Context <= 0 {
		params.Context = 3
	}
	if params.Source == "" {
		params.Source = "all"
	}

	start := time.Now()

	var results []SearchResult

	// Search scriptures
	if params.Source == "all" || params.Source == "scriptures" {
		scriptureResults, err := t.searchScriptures(params)
		if err != nil {
			return nil, fmt.Errorf("searching scriptures: %w", err)
		}
		results = append(results, scriptureResults...)
	}

	// Search talks
	if params.Source == "all" || params.Source == "conference" {
		talkResults, err := t.searchTalks(params)
		if err != nil {
			return nil, fmt.Errorf("searching talks: %w", err)
		}
		results = append(results, talkResults...)
	}

	// Search manuals
	if params.Source == "all" || params.Source == "manual" || params.Source == "magazine" {
		manualResults, err := t.searchManuals(params)
		if err != nil {
			return nil, fmt.Errorf("searching manuals: %w", err)
		}
		results = append(results, manualResults...)
	}

	// Limit results
	if len(results) > params.Limit {
		results = results[:params.Limit]
	}

	elapsed := time.Since(start)

	return &SearchResponse{
		Query:        params.Query,
		TotalMatches: len(results),
		Results:      results,
		QueryTimeMs:  elapsed.Milliseconds(),
	}, nil
}

func (t *Tools) searchScriptures(params SearchParams) ([]SearchResult, error) {
	// Build FTS5 query
	ftsQuery := buildFTSQuery(params.Query)

	query := `
		SELECT s.id, s.volume, s.book, s.chapter, s.verse, s.text, s.file_path, s.source_url,
		       snippet(scriptures_fts, 0, '**', '**', '...', 32) as excerpt
		FROM scriptures_fts
		JOIN scriptures s ON scriptures_fts.rowid = s.id
		WHERE scriptures_fts MATCH ?
	`

	args := []interface{}{ftsQuery}

	// Add path filter
	if params.Path != "" {
		query += " AND s.file_path LIKE ?"
		args = append(args, "%"+params.Path+"%")
	}

	query += " ORDER BY rank LIMIT ?"
	args = append(args, params.Limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id int
		var volume, book string
		var chapter, verse int
		var text, filePath, sourceURL, excerpt string

		if err := rows.Scan(&id, &volume, &book, &chapter, &verse, &text, &filePath, &sourceURL, &excerpt); err != nil {
			continue
		}

		ref := formatScriptureRef(volume, book, chapter, verse)

		result := SearchResult{
			Reference:  ref,
			Title:      formatChapterTitle(volume, book, chapter),
			Excerpt:    excerpt,
			FilePath:   filePath,
			SourceURL:  sourceURL,
			SourceType: "scripture",
		}

		if params.IncludeContent {
			result.Content = text
		}

		// Get context
		if params.Context > 0 {
			result.ContextBefore = t.getVerseContext(volume, book, chapter, verse, -params.Context)
			result.ContextAfter = t.getVerseContext(volume, book, chapter, verse, params.Context)
		}

		// Get cross-references
		result.RelatedReferences = t.getCrossReferences(volume, book, chapter, verse)

		results = append(results, result)
	}

	return results, nil
}

func (t *Tools) searchTalks(params SearchParams) ([]SearchResult, error) {
	ftsQuery := buildFTSQuery(params.Query)

	query := `
		SELECT t.id, t.year, t.month, t.speaker, t.title, t.file_path, t.source_url,
		       snippet(talks_fts, 2, '**', '**', '...', 64) as excerpt
		FROM talks_fts
		JOIN talks t ON talks_fts.rowid = t.id
		WHERE talks_fts MATCH ?
	`

	args := []interface{}{ftsQuery}

	// Add path filter (for year/month filtering)
	if params.Path != "" {
		query += " AND t.file_path LIKE ?"
		args = append(args, "%"+params.Path+"%")
	}

	query += " ORDER BY rank LIMIT ?"
	args = append(args, params.Limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, year, month int
		var speaker, title, filePath, sourceURL, excerpt string

		if err := rows.Scan(&id, &year, &month, &speaker, &title, &filePath, &sourceURL, &excerpt); err != nil {
			continue
		}

		result := SearchResult{
			Reference:  fmt.Sprintf("%s, %s %d", speaker, monthName(month), year),
			Title:      title,
			Excerpt:    excerpt,
			FilePath:   filePath,
			SourceURL:  sourceURL,
			SourceType: "conference",
		}

		results = append(results, result)
	}

	return results, nil
}

func (t *Tools) searchManuals(params SearchParams) ([]SearchResult, error) {
	ftsQuery := buildFTSQuery(params.Query)

	query := `
		SELECT m.id, m.content_type, m.collection_id, m.section, m.title, m.file_path, m.source_url,
		       snippet(manuals_fts, 1, '**', '**', '...', 64) as excerpt
		FROM manuals_fts
		JOIN manuals m ON manuals_fts.rowid = m.id
		WHERE manuals_fts MATCH ?
	`

	args := []interface{}{ftsQuery}

	// Filter by content type
	if params.Source == "magazine" {
		query += " AND m.content_type = 'magazine'"
	} else if params.Source == "manual" {
		query += " AND m.content_type IN ('manual', 'handbook')"
	}

	// Add path filter
	if params.Path != "" {
		query += " AND (m.file_path LIKE ? OR m.collection_id LIKE ?)"
		args = append(args, "%"+params.Path+"%", "%"+params.Path+"%")
	}

	query += " ORDER BY rank LIMIT ?"
	args = append(args, params.Limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id int
		var contentType, collectionID, section, title, filePath, sourceURL, excerpt string

		if err := rows.Scan(&id, &contentType, &collectionID, &section, &title, &filePath, &sourceURL, &excerpt); err != nil {
			continue
		}

		result := SearchResult{
			Reference:  title,
			Title:      formatCollectionTitle(collectionID),
			Excerpt:    excerpt,
			FilePath:   filePath,
			SourceURL:  sourceURL,
			SourceType: contentType,
		}

		results = append(results, result)
	}

	return results, nil
}

func (t *Tools) getVerseContext(volume, book string, chapter, verse, delta int) []string {
	var results []string

	if delta == 0 {
		return results
	}

	var query string
	var startVerse, endVerse int

	if delta < 0 {
		// Before context
		startVerse = verse + delta
		if startVerse < 1 {
			startVerse = 1
		}
		endVerse = verse - 1
		query = `SELECT verse, text FROM scriptures 
				 WHERE volume = ? AND book = ? AND chapter = ? AND verse >= ? AND verse <= ?
				 ORDER BY verse ASC`
	} else {
		// After context
		startVerse = verse + 1
		endVerse = verse + delta
		query = `SELECT verse, text FROM scriptures 
				 WHERE volume = ? AND book = ? AND chapter = ? AND verse >= ? AND verse <= ?
				 ORDER BY verse ASC`
	}

	rows, err := t.db.Query(query, volume, book, chapter, startVerse, endVerse)
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
		results = append(results, fmt.Sprintf("%d. %s", v, text))
	}

	return results
}

func (t *Tools) getCrossReferences(volume, book string, chapter, verse int) []RelatedReference {
	var refs []RelatedReference

	rows, err := t.db.Query(`
		SELECT target_volume, target_book, target_chapter, target_verse, reference_type
		FROM cross_references
		WHERE source_volume = ? AND source_book = ? AND source_chapter = ? AND source_verse = ?
		LIMIT 10
	`, volume, book, chapter, verse)
	if err != nil {
		return refs
	}
	defer rows.Close()

	for rows.Next() {
		var targetVolume, targetBook, refType string
		var targetChapter int
		var targetVerse *int

		if err := rows.Scan(&targetVolume, &targetBook, &targetChapter, &targetVerse, &refType); err != nil {
			continue
		}

		ref := RelatedReference{
			Reference: formatScriptureRef(targetVolume, targetBook, targetChapter, derefInt(targetVerse)),
			Type:      refType,
		}
		refs = append(refs, ref)
	}

	return refs
}

// buildFTSQuery converts user query to FTS5 syntax.
func buildFTSQuery(query string) string {
	// Already in FTS5 format if contains operators
	if strings.Contains(query, " OR ") ||
		strings.Contains(query, " AND ") ||
		strings.Contains(query, " NOT ") ||
		strings.Contains(query, "\"") ||
		strings.Contains(query, "*") {
		return query
	}

	// Simple query: treat as AND of all terms
	return query
}

func formatScriptureRef(volume, book string, chapter, verse int) string {
	bookName := formatBookName(volume, book)
	if verse > 0 {
		return fmt.Sprintf("%s %d:%d", bookName, chapter, verse)
	}
	return fmt.Sprintf("%s %d", bookName, chapter)
}

func formatBookName(volume, book string) string {
	// Map abbreviations to full names
	names := map[string]string{
		"gen": "Genesis", "ex": "Exodus", "lev": "Leviticus", "num": "Numbers", "deut": "Deuteronomy",
		"josh": "Joshua", "judg": "Judges", "ruth": "Ruth", "1-sam": "1 Samuel", "2-sam": "2 Samuel",
		"1-kgs": "1 Kings", "2-kgs": "2 Kings", "1-chr": "1 Chronicles", "2-chr": "2 Chronicles",
		"ezra": "Ezra", "neh": "Nehemiah", "esth": "Esther", "job": "Job", "ps": "Psalms",
		"prov": "Proverbs", "eccl": "Ecclesiastes", "song": "Song of Solomon", "isa": "Isaiah",
		"jer": "Jeremiah", "lam": "Lamentations", "ezek": "Ezekiel", "dan": "Daniel",
		"hosea": "Hosea", "joel": "Joel", "amos": "Amos", "obad": "Obadiah", "jonah": "Jonah",
		"micah": "Micah", "nahum": "Nahum", "hab": "Habakkuk", "zeph": "Zephaniah",
		"hag": "Haggai", "zech": "Zechariah", "mal": "Malachi",
		// New Testament
		"matt": "Matthew", "mark": "Mark", "luke": "Luke", "john": "John", "acts": "Acts",
		"rom": "Romans", "1-cor": "1 Corinthians", "2-cor": "2 Corinthians", "gal": "Galatians",
		"eph": "Ephesians", "philip": "Philippians", "col": "Colossians",
		"1-thes": "1 Thessalonians", "2-thes": "2 Thessalonians", "1-tim": "1 Timothy",
		"2-tim": "2 Timothy", "titus": "Titus", "philem": "Philemon", "heb": "Hebrews",
		"james": "James", "1-pet": "1 Peter", "2-pet": "2 Peter", "1-jn": "1 John",
		"2-jn": "2 John", "3-jn": "3 John", "jude": "Jude", "rev": "Revelation",
		// Book of Mormon
		"1-ne": "1 Nephi", "2-ne": "2 Nephi", "jacob": "Jacob", "enos": "Enos",
		"jarom": "Jarom", "omni": "Omni", "w-of-m": "Words of Mormon", "mosiah": "Mosiah",
		"alma": "Alma", "hel": "Helaman", "3-ne": "3 Nephi", "4-ne": "4 Nephi",
		"morm": "Mormon", "ether": "Ether", "moro": "Moroni",
		// Doctrine and Covenants
		"dc": "D&C",
		// Pearl of Great Price
		"moses": "Moses", "abr": "Abraham", "js-m": "JS—Matthew", "js-h": "JS—History",
		"a-of-f": "Articles of Faith",
	}

	if name, ok := names[book]; ok {
		return name
	}
	return strings.Title(book)
}

func formatChapterTitle(volume, book string, chapter int) string {
	return fmt.Sprintf("%s %d", formatBookName(volume, book), chapter)
}

func formatCollectionTitle(collectionID string) string {
	// Clean up collection ID to readable title
	title := strings.ReplaceAll(collectionID, "-", " ")
	return strings.Title(title)
}

func monthName(month int) string {
	months := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return fmt.Sprintf("%02d", month)
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
