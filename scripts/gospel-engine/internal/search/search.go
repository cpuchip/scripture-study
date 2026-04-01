// Package search provides keyword (FTS5) and semantic (vector) search across gospel content.
package search

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/vec"
)

// Mode controls search behavior.
type Mode string

const (
	ModeKeyword  Mode = "keyword"
	ModeSemantic Mode = "semantic"
	ModeCombined Mode = "combined"
)

// Result represents a unified search result.
type Result struct {
	Source    string  `json:"source"`    // scriptures, conference, manual, book
	Type      string  `json:"type"`      // verse, paragraph, summary, theme, fts
	Reference string  `json:"reference"` // Human-readable reference
	Content   string  `json:"content"`
	FilePath  string  `json:"file_path"`
	SourceURL string  `json:"source_url,omitempty"`
	Score     float64 `json:"score"`
	// Metadata for filtering
	Book    string `json:"book,omitempty"`
	Chapter int    `json:"chapter,omitempty"`
	Verse   int    `json:"verse,omitempty"`
	Speaker string `json:"speaker,omitempty"`
	Year    int    `json:"year,omitempty"`
	Month   string `json:"month,omitempty"`
}

// Options controls search behavior.
type Options struct {
	Mode    Mode
	Limit   int
	Sources []string // Filter by source type (scriptures, conference, manual, book)
	// Semantic-specific
	Layers []vec.Layer
	// Conference talk filters
	Speaker  string
	YearFrom int
	YearTo   int
	// TITSW filters (conference talks only)
	TITSWMode     string // enacted, declared, doctrinal, experiential
	TITSWDominant string // filter by dominant dimension substring
	TITSWMinTeach    int // 0 = no filter
	TITSWMinHelp     int
	TITSWMinLove     int
	TITSWMinSpirit   int
	TITSWMinDoctrine int
	TITSWMinInvite   int
}

// Engine provides unified search.
type Engine struct {
	db  *db.DB
	vec vec.Searcher
}

// NewEngine creates a new search engine.
func NewEngine(database *db.DB, store vec.Searcher) *Engine {
	return &Engine{db: database, vec: store}
}

// Search performs a search based on the given options.
func (e *Engine) Search(ctx context.Context, query string, opts Options) ([]Result, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	switch opts.Mode {
	case ModeKeyword:
		return e.keywordSearch(query, opts)
	case ModeSemantic:
		return e.semanticSearch(ctx, query, opts)
	case ModeCombined:
		return e.combinedSearch(ctx, query, opts)
	default:
		return e.combinedSearch(ctx, query, opts)
	}
}

func (e *Engine) keywordSearch(query string, opts Options) ([]Result, error) {
	var results []Result

	tables := sourcesToTables(opts.Sources)

	for _, table := range tables {
		rows, err := e.ftsQuery(table, query, opts)
		if err != nil {
			continue // Table might not exist or be empty
		}
		results = append(results, rows...)
	}

	// Sort by score, limit
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}
	return results, nil
}

// sourcesToTables maps source names to FTS table names.
func sourcesToTables(sources []string) []string {
	if len(sources) == 0 {
		return []string{"scriptures", "chapters", "talks", "manuals", "books"}
	}
	var tables []string
	for _, s := range sources {
		switch s {
		case "scriptures":
			tables = append(tables, "scriptures")
			tables = append(tables, "chapters") // enriched chapter summaries
		case "conference":
			tables = append(tables, "talks")
		case "manual":
			tables = append(tables, "manuals")
		case "book":
			tables = append(tables, "books")
		}
	}
	if len(tables) == 0 {
		return []string{"scriptures", "chapters", "talks", "manuals", "books"}
	}
	return tables
}

func (e *Engine) ftsQuery(table, query string, opts Options) ([]Result, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	var results []Result

	switch table {
	case "scriptures":
		rows, err := e.db.Query(`
			SELECT s.volume, s.book, s.chapter, s.verse, s.text, s.file_path, s.source_url,
			       rank
			FROM scriptures_fts f
			JOIN scriptures s ON s.id = f.rowid
			WHERE scriptures_fts MATCH ?
			ORDER BY rank
			LIMIT ?
		`, query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var volume, book, text, filePath, sourceURL string
			var chapter, verse int
			var rank float64
			if err := rows.Scan(&volume, &book, &chapter, &verse, &text, &filePath, &sourceURL, &rank); err != nil {
				continue
			}
			results = append(results, Result{
				Source:    "scriptures",
				Type:      "verse",
				Reference: fmt.Sprintf("%s %d:%d", book, chapter, verse),
				Content:   text,
				FilePath:  filePath,
				SourceURL: sourceURL,
				Score:     -rank, // FTS5 rank is negative (lower = better)
				Book:      book,
				Chapter:   chapter,
				Verse:     verse,
			})
		}

	case "talks":
		q := `
			SELECT t.speaker, t.title, t.year, t.month, t.file_path, t.source_url,
			       snippet(talks_fts, 2, '**', '**', '...', 64) as snippet,
			       rank
			FROM talks_fts f
			JOIN talks t ON t.id = f.rowid
			WHERE talks_fts MATCH ?
		`
		args := []any{query}

		if opts.Speaker != "" {
			q += " AND t.speaker LIKE ?"
			args = append(args, "%"+opts.Speaker+"%")
		}
		if opts.YearFrom > 0 {
			q += " AND t.year >= ?"
			args = append(args, opts.YearFrom)
		}
		if opts.YearTo > 0 {
			q += " AND t.year <= ?"
			args = append(args, opts.YearTo)
		}
		if opts.TITSWMode != "" {
			q += " AND t.titsw_mode = ?"
			args = append(args, opts.TITSWMode)
		}
		if opts.TITSWDominant != "" {
			q += " AND t.titsw_dominant LIKE ?"
			args = append(args, "%"+opts.TITSWDominant+"%")
		}
		if opts.TITSWMinTeach > 0 {
			q += " AND t.titsw_teach >= ?"
			args = append(args, opts.TITSWMinTeach)
		}
		if opts.TITSWMinHelp > 0 {
			q += " AND t.titsw_help >= ?"
			args = append(args, opts.TITSWMinHelp)
		}
		if opts.TITSWMinLove > 0 {
			q += " AND t.titsw_love >= ?"
			args = append(args, opts.TITSWMinLove)
		}
		if opts.TITSWMinSpirit > 0 {
			q += " AND t.titsw_spirit >= ?"
			args = append(args, opts.TITSWMinSpirit)
		}
		if opts.TITSWMinDoctrine > 0 {
			q += " AND t.titsw_doctrine >= ?"
			args = append(args, opts.TITSWMinDoctrine)
		}
		if opts.TITSWMinInvite > 0 {
			q += " AND t.titsw_invite >= ?"
			args = append(args, opts.TITSWMinInvite)
		}

		q += " ORDER BY rank LIMIT ?"
		args = append(args, limit)

		rows, err := e.db.Query(q, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var speaker, title, filePath, sourceURL, snippet string
			var year, month int
			var rank float64
			if err := rows.Scan(&speaker, &title, &year, &month, &filePath, &sourceURL, &snippet, &rank); err != nil {
				continue
			}
			results = append(results, Result{
				Source:    "conference",
				Type:      "talk",
				Reference: fmt.Sprintf("%s, \"%s\" (%d/%02d)", speaker, title, year, month),
				Content:   snippet,
				FilePath:  filePath,
				SourceURL: sourceURL,
				Score:     -rank,
				Speaker:   speaker,
				Year:      year,
				Month:     fmt.Sprintf("%02d", month),
			})
		}

	case "manuals":
		rows, err := e.db.Query(`
			SELECT m.title, m.collection_id, m.section, m.file_path, m.source_url,
			       snippet(manuals_fts, 1, '**', '**', '...', 64) as snippet,
			       rank
			FROM manuals_fts f
			JOIN manuals m ON m.id = f.rowid
			WHERE manuals_fts MATCH ?
			ORDER BY rank
			LIMIT ?
		`, query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var title, collID, section, filePath, sourceURL, snippet string
			var rank float64
			if err := rows.Scan(&title, &collID, &section, &filePath, &sourceURL, &snippet, &rank); err != nil {
				continue
			}
			results = append(results, Result{
				Source:    "manual",
				Type:      "manual",
				Reference: title,
				Content:   snippet,
				FilePath:  filePath,
				SourceURL: sourceURL,
				Score:     -rank,
			})
		}

	case "books":
		rows, err := e.db.Query(`
			SELECT b.title, b.collection, b.section, b.file_path,
			       snippet(books_fts, 1, '**', '**', '...', 64) as snippet,
			       rank
			FROM books_fts f
			JOIN books b ON b.id = f.rowid
			WHERE books_fts MATCH ?
			ORDER BY rank
			LIMIT ?
		`, query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var title, collection, section, filePath, snippet string
			var rank float64
			if err := rows.Scan(&title, &collection, &section, &filePath, &snippet, &rank); err != nil {
				continue
			}
			results = append(results, Result{
				Source:    "book",
				Type:      "book",
				Reference: title,
				Content:   snippet,
				FilePath:  filePath,
				Score:     -rank,
			})
		}

	case "chapters":
		rows, err := e.db.Query(`
			SELECT c.volume, c.book, c.chapter, c.title, c.file_path,
			       c.enrichment_summary, c.enrichment_christ_types,
			       rank
			FROM chapters_fts f
			JOIN chapters c ON c.id = f.rowid
			WHERE chapters_fts MATCH ?
			ORDER BY rank
			LIMIT ?
		`, query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var volume, book, title, filePath string
			var summary, christTypes *string
			var chapter int
			var rank float64
			if err := rows.Scan(&volume, &book, &chapter, &title, &filePath, &summary, &christTypes, &rank); err != nil {
				continue
			}
			content := ""
			if summary != nil {
				content = *summary
			}
			if christTypes != nil && *christTypes != "" && *christTypes != "none" {
				content += "\nChrist types: " + *christTypes
			}
			results = append(results, Result{
				Source:    "scriptures",
				Type:      "chapter",
				Reference: fmt.Sprintf("%s %d", book, chapter),
				Content:   content,
				FilePath:  filePath,
				Score:     -rank,
				Book:      book,
				Chapter:   chapter,
			})
		}
	}

	return results, nil
}

func (e *Engine) semanticSearch(ctx context.Context, query string, opts Options) ([]Result, error) {
	if e.vec == nil {
		return nil, fmt.Errorf("vector store not available")
	}

	vecOpts := vec.SearchOptions{
		Limit: opts.Limit,
	}

	if len(opts.Layers) > 0 {
		vecOpts.Layers = opts.Layers
	}
	if len(opts.Sources) > 0 {
		for _, s := range opts.Sources {
			vecOpts.Sources = append(vecOpts.Sources, vec.Source(s))
		}
	}

	vecResults, err := e.vec.Search(ctx, query, vecOpts)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(vecResults))
	for _, vr := range vecResults {
		r := Result{
			Source:    string(vr.Metadata.Source),
			Type:      string(vr.Metadata.Layer),
			Reference: vr.Metadata.Reference,
			Content:   vr.Content,
			FilePath:  vr.Metadata.FilePath,
			Score:     float64(vr.Score),
			Book:      vr.Metadata.Book,
			Chapter:   vr.Metadata.Chapter,
			Speaker:   vr.Metadata.Speaker,
			Year:      vr.Metadata.Year,
			Month:     vr.Metadata.Month,
		}

		// Apply post-search filters for talks
		if opts.Speaker != "" && r.Speaker != "" {
			if !strings.Contains(strings.ToLower(r.Speaker), strings.ToLower(opts.Speaker)) {
				continue
			}
		}
		if opts.YearFrom > 0 && r.Year > 0 && r.Year < opts.YearFrom {
			continue
		}
		if opts.YearTo > 0 && r.Year > 0 && r.Year > opts.YearTo {
			continue
		}

		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

func (e *Engine) combinedSearch(ctx context.Context, query string, opts Options) ([]Result, error) {
	return e.rrfCombinedSearch(ctx, query, opts)
}
