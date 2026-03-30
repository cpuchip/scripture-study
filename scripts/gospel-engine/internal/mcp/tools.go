package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/search"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/vec"
)

// Tools returns the tool definitions for gospel-engine.
func Tools() []ToolDef {
	return []ToolDef{
		{
			Name: "gospel_search",
			Description: `Search scriptures, conference talks, manuals, and books using keyword (FTS5), semantic (vector), or combined search.

Modes:
- "keyword": Fast full-text search using SQLite FTS5. Best for known phrases or exact words.
- "semantic": Vector similarity search. Best for concepts, themes, and meaning-based queries.
- "combined" (default): Runs both and merges results.

IMPORTANT: Results labeled [AI SUMMARY] or [AI THEME] are NOT direct quotes — always verify against the source file before quoting. Results include file paths for follow-up with read_file.

Tip: Use gospel_get to read the full source text after finding relevant content.`,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query (e.g., 'faith in Christ', 'repentance and grace')",
					},
					"mode": map[string]any{
						"type":        "string",
						"enum":        []string{"keyword", "semantic", "combined"},
						"description": "Search mode (default: combined)",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum results (default: 10)",
					},
					"sources": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string", "enum": []string{"scriptures", "conference", "manual", "book"}},
						"description": "Filter by content source (default: all)",
					},
					"layers": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string", "enum": []string{"verse", "paragraph", "summary", "theme"}},
						"description": "Which vector layers to search (semantic/combined only, default: all available)",
					},
					"speaker": map[string]any{
						"type":        "string",
						"description": "Filter talks by speaker last name (case-insensitive)",
					},
					"year_from": map[string]any{
						"type":        "integer",
						"description": "Filter talks from this year (inclusive)",
					},
					"year_to": map[string]any{
						"type":        "integer",
						"description": "Filter talks to this year (inclusive)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name: "gospel_get",
			Description: `Get the full text of a scripture chapter, conference talk, manual section, or book section.

Use after gospel_search finds relevant content, or when you know the specific reference. Always use this to read the actual text before quoting.

For scriptures, provide volume + book + chapter. For talks, provide file_path (from search results) or speaker + year + month.`,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "Direct file path (from search results). If provided, other params are ignored.",
					},
					"volume": map[string]any{
						"type":        "string",
						"description": "Scripture volume: ot, nt, bofm, dc-testament, pgp",
					},
					"book": map[string]any{
						"type":        "string",
						"description": "Book abbreviation: gen, matt, 1-ne, dc, moses, etc.",
					},
					"chapter": map[string]any{
						"type":        "integer",
						"description": "Chapter number",
					},
					"speaker": map[string]any{
						"type":        "string",
						"description": "Talk speaker last name (for conference talk lookup)",
					},
					"year": map[string]any{
						"type":        "integer",
						"description": "Conference year",
					},
					"month": map[string]any{
						"type":        "string",
						"description": "Conference month: '04' or '10'",
					},
				},
			},
		},
		{
			Name: "gospel_list",
			Description: `List available content in the index. Shows indexed books, conference years/speakers, manual collections, etc.

Use to discover what content is available for search and retrieval.`,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type": map[string]any{
						"type":        "string",
						"enum":        []string{"scriptures", "conference", "manuals", "books", "stats"},
						"description": "What to list (default: stats — shows overall index counts)",
					},
					"volume": map[string]any{
						"type":        "string",
						"description": "Filter scripture listing by volume (ot, nt, bofm, dc-testament, pgp)",
					},
					"year": map[string]any{
						"type":        "integer",
						"description": "Filter conference listing by year",
					},
				},
			},
		},
	}
}

// Handler creates a tool call handler using the given database and vector store.
func Handler(database *db.DB, store *vec.Store, root string) func(name string, args json.RawMessage) (string, error) {
	engine := search.NewEngine(database, store)

	return func(name string, args json.RawMessage) (string, error) {
		switch name {
		case "gospel_search":
			return handleSearch(engine, args)
		case "gospel_get":
			return handleGet(database, root, args)
		case "gospel_list":
			return handleList(database, store, args)
		default:
			return "", fmt.Errorf("unknown tool: %s", name)
		}
	}
}

func handleSearch(engine *search.Engine, args json.RawMessage) (string, error) {
	var params struct {
		Query    string   `json:"query"`
		Mode     string   `json:"mode"`
		Limit    int      `json:"limit"`
		Sources  []string `json:"sources"`
		Layers   []string `json:"layers"`
		Speaker  string   `json:"speaker"`
		YearFrom int      `json:"year_from"`
		YearTo   int      `json:"year_to"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parsing arguments: %w", err)
	}

	opts := search.Options{
		Limit:    params.Limit,
		Sources:  params.Sources,
		Speaker:  params.Speaker,
		YearFrom: params.YearFrom,
		YearTo:   params.YearTo,
	}

	switch params.Mode {
	case "keyword":
		opts.Mode = search.ModeKeyword
	case "semantic":
		opts.Mode = search.ModeSemantic
	case "combined":
		opts.Mode = search.ModeCombined
	default:
		opts.Mode = search.ModeCombined
	}

	for _, l := range params.Layers {
		opts.Layers = append(opts.Layers, vec.Layer(l))
	}

	results, err := engine.Search(context.Background(), params.Query, opts)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "No results found.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results for \"%s\":\n\n", len(results), params.Query))

	for i, r := range results {
		label := "[DIRECT QUOTE]"
		if r.Type == "summary" || r.Type == "theme" {
			label = "[AI " + strings.ToUpper(r.Type) + "]"
		}

		sb.WriteString(fmt.Sprintf("**%d. %s** %s (score: %.4f)\n", i+1, r.Reference, label, r.Score))
		sb.WriteString(fmt.Sprintf("   Source: %s | File: %s\n", r.Source, r.FilePath))
		if r.SourceURL != "" {
			sb.WriteString(fmt.Sprintf("   URL: %s\n", r.SourceURL))
		}

		content := r.Content
		if len(content) > 300 {
			content = content[:300] + "... [TRUNCATED — use gospel_get to read full text]"
		}
		sb.WriteString(fmt.Sprintf("   %s\n\n", content))
	}

	return sb.String(), nil
}

func handleGet(database *db.DB, root string, args json.RawMessage) (string, error) {
	var params struct {
		FilePath string `json:"file_path"`
		Volume   string `json:"volume"`
		Book     string `json:"book"`
		Chapter  int    `json:"chapter"`
		Speaker  string `json:"speaker"`
		Year     int    `json:"year"`
		Month    string `json:"month"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parsing arguments: %w", err)
	}

	// Direct file path
	if params.FilePath != "" {
		return readFileContent(root, params.FilePath)
	}

	// Scripture lookup
	if params.Volume != "" && params.Book != "" && params.Chapter > 0 {
		var content string
		err := database.QueryRow(`
			SELECT full_content FROM chapters
			WHERE volume = ? AND book = ? AND chapter = ?
		`, params.Volume, params.Book, params.Chapter).Scan(&content)
		if err != nil {
			return "", fmt.Errorf("chapter not found: %s %s %d", params.Volume, params.Book, params.Chapter)
		}
		return content, nil
	}

	// Talk lookup by speaker + year + month
	if params.Speaker != "" {
		q := `SELECT content, file_path FROM talks WHERE speaker LIKE ?`
		args := []any{"%" + params.Speaker + "%"}

		if params.Year > 0 {
			q += " AND year = ?"
			args = append(args, params.Year)
		}
		if params.Month != "" {
			q += " AND month = ?"
			args = append(args, params.Month)
		}
		q += " LIMIT 1"

		var content, filePath string
		if err := database.QueryRow(q, args...).Scan(&content, &filePath); err != nil {
			return "", fmt.Errorf("talk not found for speaker: %s", params.Speaker)
		}
		return fmt.Sprintf("File: %s\n\n%s", filePath, content), nil
	}

	return "", fmt.Errorf("provide file_path, volume+book+chapter, or speaker(+year+month)")
}

func handleList(database *db.DB, store *vec.Store, args json.RawMessage) (string, error) {
	var params struct {
		Type   string `json:"type"`
		Volume string `json:"volume"`
		Year   int    `json:"year"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parsing arguments: %w", err)
	}

	if params.Type == "" {
		params.Type = "stats"
	}

	var sb strings.Builder

	switch params.Type {
	case "stats":
		stats, err := database.GetStats()
		if err != nil {
			return "", err
		}
		sb.WriteString("## Gospel Engine Index Stats\n\n")
		sb.WriteString(fmt.Sprintf("| Content | Count |\n|---------|-------|\n"))
		sb.WriteString(fmt.Sprintf("| Scriptures (verses) | %d |\n", stats.Scriptures))
		sb.WriteString(fmt.Sprintf("| Chapters | %d |\n", stats.Chapters))
		sb.WriteString(fmt.Sprintf("| Conference Talks | %d |\n", stats.Talks))
		sb.WriteString(fmt.Sprintf("| Manuals | %d |\n", stats.Manuals))
		sb.WriteString(fmt.Sprintf("| Books | %d |\n", stats.Books))
		sb.WriteString(fmt.Sprintf("| Cross References | %d |\n", stats.CrossRefs))
		sb.WriteString(fmt.Sprintf("| Graph Edges | %d |\n", stats.Edges))

		if store != nil {
			sb.WriteString("\n### Vector Collections\n\n")
			sb.WriteString("| Collection | Documents |\n|------------|----------|\n")
			for name, count := range store.Stats() {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", name, count))
			}
		}

	case "scriptures":
		q := `SELECT DISTINCT volume, book, COUNT(*) as chapters FROM chapters`
		args := []any{}
		if params.Volume != "" {
			q += " WHERE volume = ?"
			args = append(args, params.Volume)
		}
		q += " GROUP BY volume, book ORDER BY volume, book"

		rows, err := database.Query(q, args...)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		sb.WriteString("## Indexed Scriptures\n\n")
		sb.WriteString("| Volume | Book | Chapters |\n|--------|------|----------|\n")
		for rows.Next() {
			var volume, book string
			var chapters int
			if err := rows.Scan(&volume, &book, &chapters); err != nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", volume, book, chapters))
		}

	case "conference":
		q := `SELECT year, month, COUNT(*) as talks FROM talks`
		args := []any{}
		if params.Year > 0 {
			q += " WHERE year = ?"
			args = append(args, params.Year)
		}
		q += " GROUP BY year, month ORDER BY year DESC, month"

		rows, err := database.Query(q, args...)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		sb.WriteString("## Indexed Conference Talks\n\n")
		sb.WriteString("| Year | Month | Talks |\n|------|-------|-------|\n")
		for rows.Next() {
			var year, month, talks int
			if err := rows.Scan(&year, &month, &talks); err != nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("| %d | %02d | %d |\n", year, month, talks))
		}

	case "manuals":
		rows, err := database.Query(`
			SELECT content_type, collection_id, COUNT(*) as sections
			FROM manuals
			GROUP BY content_type, collection_id
			ORDER BY content_type, collection_id
		`)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		sb.WriteString("## Indexed Manuals\n\n")
		sb.WriteString("| Type | Collection | Sections |\n|------|------------|----------|\n")
		for rows.Next() {
			var contentType, collID string
			var sections int
			if err := rows.Scan(&contentType, &collID, &sections); err != nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", contentType, collID, sections))
		}

	case "books":
		rows, err := database.Query(`
			SELECT collection, COUNT(*) as sections
			FROM books
			GROUP BY collection
			ORDER BY collection
		`)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		sb.WriteString("## Indexed Books\n\n")
		sb.WriteString("| Collection | Sections |\n|------------|----------|\n")
		for rows.Next() {
			var collection string
			var sections int
			if err := rows.Scan(&collection, &sections); err != nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", collection, sections))
		}
	}

	return sb.String(), nil
}

func readFileContent(root, filePath string) (string, error) {
	// Try as relative path from root
	fullPath := filePath
	if !strings.HasPrefix(filePath, "/") && !strings.Contains(filePath, ":") {
		fullPath = root + "/" + filePath
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", filePath, err)
	}

	text := string(content)
	if len(text) > 50000 {
		text = text[:50000] + "\n\n[TRUNCATED — file too large. Use read_file for specific sections.]"
	}

	return text, nil
}
