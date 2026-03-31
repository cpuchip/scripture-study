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

TITSW Filters (conference talks only — requires enrichment):
- titsw_mode: Filter by teaching mode (enacted, declared, doctrinal, experiential)
- titsw_dominant: Filter by dominant dimension (teach_about_christ, help_come_to_christ, love, spirit, doctrine, invite)
- titsw_min_*: Minimum score (0-9) for each dimension

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
					"titsw_mode": map[string]any{
						"type":        "string",
						"enum":        []string{"enacted", "declared", "doctrinal", "experiential"},
						"description": "Filter talks by TITSW teaching mode",
					},
					"titsw_dominant": map[string]any{
						"type":        "string",
						"description": "Filter talks by dominant dimension (teach_about_christ, help_come_to_christ, love, spirit, doctrine, invite)",
					},
					"titsw_min_teach": map[string]any{
						"type":        "integer",
						"description": "Minimum teach_about_christ score (0-9)",
					},
					"titsw_min_help": map[string]any{
						"type":        "integer",
						"description": "Minimum help_come_to_christ score (0-9)",
					},
					"titsw_min_love": map[string]any{
						"type":        "integer",
						"description": "Minimum love score (0-9)",
					},
					"titsw_min_spirit": map[string]any{
						"type":        "integer",
						"description": "Minimum spirit score (0-9)",
					},
					"titsw_min_doctrine": map[string]any{
						"type":        "integer",
						"description": "Minimum doctrine score (0-9)",
					},
					"titsw_min_invite": map[string]any{
						"type":        "integer",
						"description": "Minimum invite score (0-9)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name: "gospel_get",
			Description: `Retrieve scripture verses, conference talks, manual sections, or book sections by reference or path.

For scriptures, use the "reference" parameter with natural references like "1 Nephi 3:7", "D&C 93:24-30", or "Mosiah 4". Returns individual verses — lean output for quoting. For full chapter context with footnotes and formatting, use read_file on the file_path returned.

For talks, use file_path (from search results) or speaker + year + month.

Set cross_refs=true to include cross-references for scripture verses.`,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"reference": map[string]any{
						"type":        "string",
						"description": "Scripture reference: '1 Nephi 3:7', 'D&C 93:24-30', 'Mosiah 4'. Supports single verse, verse range, or chapter.",
					},
					"file_path": map[string]any{
						"type":        "string",
						"description": "Direct file path (from search results). If provided, other params are ignored.",
					},
					"volume": map[string]any{
						"type":        "string",
						"description": "Scripture volume: ot, nt, bofm, dc-testament, pgp (fallback if reference not provided)",
					},
					"book": map[string]any{
						"type":        "string",
						"description": "Book abbreviation: gen, matt, 1-ne, dc, moses, etc.",
					},
					"chapter": map[string]any{
						"type":        "integer",
						"description": "Chapter number",
					},
					"cross_refs": map[string]any{
						"type":        "boolean",
						"description": "Include cross-references for scripture verses (default: false)",
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
func Handler(database *db.DB, store vec.Searcher, root string) func(name string, args json.RawMessage) (string, error) {
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
		Query           string   `json:"query"`
		Mode            string   `json:"mode"`
		Limit           int      `json:"limit"`
		Sources         []string `json:"sources"`
		Layers          []string `json:"layers"`
		Speaker         string   `json:"speaker"`
		YearFrom        int      `json:"year_from"`
		YearTo          int      `json:"year_to"`
		TITSWMode       string   `json:"titsw_mode"`
		TITSWDominant   string   `json:"titsw_dominant"`
		TITSWMinTeach   int      `json:"titsw_min_teach"`
		TITSWMinHelp    int      `json:"titsw_min_help"`
		TITSWMinLove    int      `json:"titsw_min_love"`
		TITSWMinSpirit  int      `json:"titsw_min_spirit"`
		TITSWMinDoctrine int     `json:"titsw_min_doctrine"`
		TITSWMinInvite  int      `json:"titsw_min_invite"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parsing arguments: %w", err)
	}

	opts := search.Options{
		Limit:            params.Limit,
		Sources:          params.Sources,
		Speaker:          params.Speaker,
		YearFrom:         params.YearFrom,
		YearTo:           params.YearTo,
		TITSWMode:        params.TITSWMode,
		TITSWDominant:    params.TITSWDominant,
		TITSWMinTeach:    params.TITSWMinTeach,
		TITSWMinHelp:     params.TITSWMinHelp,
		TITSWMinLove:     params.TITSWMinLove,
		TITSWMinSpirit:   params.TITSWMinSpirit,
		TITSWMinDoctrine: params.TITSWMinDoctrine,
		TITSWMinInvite:   params.TITSWMinInvite,
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
		Reference string `json:"reference"`
		FilePath  string `json:"file_path"`
		Volume    string `json:"volume"`
		Book      string `json:"book"`
		Chapter   int    `json:"chapter"`
		CrossRefs bool   `json:"cross_refs"`
		Speaker   string `json:"speaker"`
		Year      int    `json:"year"`
		Month     string `json:"month"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("parsing arguments: %w", err)
	}

	// Direct file path — highest priority
	if params.FilePath != "" {
		return readFileContent(root, params.FilePath)
	}

	// Reference-based lookup — parse natural scripture references
	if params.Reference != "" {
		ref := parseReference(params.Reference)
		switch ref.Type {
		case "scripture":
			return getScripture(database, ref, params.CrossRefs)
		case "talk":
			return getTalkByRef(database, ref)
		default:
			return "", fmt.Errorf("could not parse reference: %s", params.Reference)
		}
	}

	// Structured scripture lookup (volume + book + chapter fallback)
	if params.Volume != "" && params.Book != "" && params.Chapter > 0 {
		ref := parsedRef{
			Type:    "scripture",
			Book:    params.Book,
			Chapter: params.Chapter,
		}
		return getScripture(database, ref, params.CrossRefs)
	}

	// Talk lookup by speaker + year + month
	if params.Speaker != "" {
		q := `SELECT content, file_path, speaker, title, year, month,
		       titsw_dominant, titsw_mode, titsw_pattern,
		       titsw_teach, titsw_help, titsw_love, titsw_spirit, titsw_doctrine, titsw_invite,
		       titsw_summary, titsw_key_quote, titsw_keywords
		FROM talks WHERE speaker LIKE ?`
		qArgs := []any{"%" + params.Speaker + "%"}

		if params.Year > 0 {
			q += " AND year = ?"
			qArgs = append(qArgs, params.Year)
		}
		if params.Month != "" {
			q += " AND month = ?"
			qArgs = append(qArgs, params.Month)
		}
		q += " LIMIT 1"

		var content, filePath, speaker, title string
		var year, month int
		var titswDominant, titswMode, titswPattern, titswSummary, titswKeyQuote, titswKeywords *string
		var titswTeach, titswHelp, titswLove, titswSpirit, titswDoctrine, titswInvite *int

		if err := database.QueryRow(q, qArgs...).Scan(
			&content, &filePath, &speaker, &title, &year, &month,
			&titswDominant, &titswMode, &titswPattern,
			&titswTeach, &titswHelp, &titswLove, &titswSpirit, &titswDoctrine, &titswInvite,
			&titswSummary, &titswKeyQuote, &titswKeywords,
		); err != nil {
			return "", fmt.Errorf("talk not found for speaker: %s", params.Speaker)
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("**%s — \"%s\" (%d/%02d)**\n", speaker, title, year, month))
		sb.WriteString(fmt.Sprintf("File: %s\n\n", filePath))

		if titswTeach != nil {
			sb.WriteString("**TITSW Teaching Profile:**\n")
			if titswDominant != nil {
				sb.WriteString(fmt.Sprintf("  Dominant: %s\n", *titswDominant))
			}
			if titswMode != nil {
				sb.WriteString(fmt.Sprintf("  Mode: %s\n", *titswMode))
			}
			if titswPattern != nil {
				sb.WriteString(fmt.Sprintf("  Pattern: %s\n", *titswPattern))
			}
			sb.WriteString(fmt.Sprintf("  Scores: teach=%d help=%d love=%d spirit=%d doctrine=%d invite=%d\n",
				*titswTeach, *titswHelp, *titswLove, *titswSpirit, *titswDoctrine, *titswInvite))
			if titswSummary != nil {
				sb.WriteString(fmt.Sprintf("  Summary: %s\n", *titswSummary))
			}
			if titswKeyQuote != nil {
				sb.WriteString(fmt.Sprintf("  Key Quote: %s\n", *titswKeyQuote))
			}
			if titswKeywords != nil {
				sb.WriteString(fmt.Sprintf("  Keywords: %s\n", *titswKeywords))
			}
			sb.WriteString("\n")
		}

		sb.WriteString(content)
		return sb.String(), nil
	}

	return "", fmt.Errorf("provide reference, file_path, volume+book+chapter, or speaker(+year+month)")
}

// getScripture handles verse, verse range, and chapter retrieval.
func getScripture(database *db.DB, ref parsedRef, crossRefs bool) (string, error) {
	var sb strings.Builder

	if ref.Verse > 0 && ref.EndVerse > 0 {
		// Verse range
		return getScriptureRange(database, ref, crossRefs)
	}

	if ref.Verse > 0 {
		// Single verse
		return getScriptureVerse(database, ref, crossRefs)
	}

	// Full chapter — query individual verses for consistent formatting
	rows, err := database.Query(`
		SELECT verse, text FROM scriptures
		WHERE book = ? AND chapter = ?
		ORDER BY verse
	`, ref.Book, ref.Chapter)
	if err != nil {
		return "", fmt.Errorf("chapter not found: %s %d", ref.Book, ref.Chapter)
	}
	defer rows.Close()

	var filePath string
	var count int
	for rows.Next() {
		var v int
		var text string
		if err := rows.Scan(&v, &text); err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n\n", v, text))
		count++
	}

	if count == 0 {
		return "", fmt.Errorf("chapter not found: %s %d", ref.Book, ref.Chapter)
	}

	// Get file path from first verse
	database.QueryRow(`
		SELECT file_path FROM scriptures WHERE book = ? AND chapter = ? LIMIT 1
	`, ref.Book, ref.Chapter).Scan(&filePath)

	header := fmt.Sprintf("**%s**\nFile: %s\n\n", formatScriptureRef(ref.Book, ref.Chapter, 0), filePath)
	return header + sb.String(), nil
}

// getScriptureVerse retrieves a single verse.
func getScriptureVerse(database *db.DB, ref parsedRef, crossRefs bool) (string, error) {
	var text, filePath, sourceURL string
	err := database.QueryRow(`
		SELECT text, file_path, source_url FROM scriptures
		WHERE book = ? AND chapter = ? AND verse = ?
	`, ref.Book, ref.Chapter, ref.Verse).Scan(&text, &filePath, &sourceURL)
	if err != nil {
		return "", fmt.Errorf("verse not found: %s %d:%d", ref.Book, ref.Chapter, ref.Verse)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n", formatScriptureRef(ref.Book, ref.Chapter, ref.Verse)))
	sb.WriteString(fmt.Sprintf("File: %s\n\n", filePath))
	sb.WriteString(fmt.Sprintf("%d. %s\n", ref.Verse, text))

	if crossRefs {
		refs := getCrossReferences(database, ref.Book, ref.Chapter, ref.Verse)
		if len(refs) > 0 {
			sb.WriteString("\nCross-references:\n")
			for _, r := range refs {
				sb.WriteString(fmt.Sprintf("  - %s (%s)\n", r.Reference, r.Type))
			}
		}
	}

	return sb.String(), nil
}

// getScriptureRange retrieves a range of verses.
func getScriptureRange(database *db.DB, ref parsedRef, crossRefs bool) (string, error) {
	rows, err := database.Query(`
		SELECT verse, text, file_path FROM scriptures
		WHERE book = ? AND chapter = ? AND verse >= ? AND verse <= ?
		ORDER BY verse
	`, ref.Book, ref.Chapter, ref.Verse, ref.EndVerse)
	if err != nil {
		return "", fmt.Errorf("verse range not found: %s %d:%d-%d", ref.Book, ref.Chapter, ref.Verse, ref.EndVerse)
	}
	defer rows.Close()

	var sb strings.Builder
	var filePath string
	var count int

	for rows.Next() {
		var v int
		var text, fp string
		if err := rows.Scan(&v, &text, &fp); err != nil {
			continue
		}
		if filePath == "" {
			filePath = fp
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n\n", v, text))
		count++
	}

	if count == 0 {
		return "", fmt.Errorf("verse range not found: %s %d:%d-%d", ref.Book, ref.Chapter, ref.Verse, ref.EndVerse)
	}

	rangeRef := fmt.Sprintf("%s %d:%d-%d", formatBookName(ref.Book), ref.Chapter, ref.Verse, ref.EndVerse)
	header := fmt.Sprintf("**%s**\nFile: %s\n\n", rangeRef, filePath)

	result := header + sb.String()

	if crossRefs {
		allRefs := getCrossReferencesForRange(database, ref.Book, ref.Chapter, ref.Verse, ref.EndVerse)
		if len(allRefs) > 0 {
			result += "Cross-references:\n"
			for _, r := range allRefs {
				result += fmt.Sprintf("  - %s (%s)\n", r.Reference, r.Type)
			}
		}
	}

	return result, nil
}

// getTalkByRef retrieves a talk by parsed speaker reference.
func getTalkByRef(database *db.DB, ref parsedRef) (string, error) {
	var content, filePath, speaker, title string
	var year, month int
	var titswDominant, titswMode, titswPattern, titswSummary, titswKeyQuote, titswKeywords *string
	var titswTeach, titswHelp, titswLove, titswSpirit, titswDoctrine, titswInvite *int

	err := database.QueryRow(`
		SELECT content, file_path, speaker, title, year, month,
		       titsw_dominant, titsw_mode, titsw_pattern,
		       titsw_teach, titsw_help, titsw_love, titsw_spirit, titsw_doctrine, titsw_invite,
		       titsw_summary, titsw_key_quote, titsw_keywords
		FROM talks
		WHERE speaker LIKE ? ORDER BY year DESC, month DESC LIMIT 1
	`, "%"+ref.Speaker+"%").Scan(
		&content, &filePath, &speaker, &title, &year, &month,
		&titswDominant, &titswMode, &titswPattern,
		&titswTeach, &titswHelp, &titswLove, &titswSpirit, &titswDoctrine, &titswInvite,
		&titswSummary, &titswKeyQuote, &titswKeywords,
	)
	if err != nil {
		return "", fmt.Errorf("talk not found: %s", ref.Speaker)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s — \"%s\" (%d/%02d)**\n", speaker, title, year, month))
	sb.WriteString(fmt.Sprintf("File: %s\n\n", filePath))

	if titswTeach != nil {
		sb.WriteString("**TITSW Teaching Profile:**\n")
		if titswDominant != nil {
			sb.WriteString(fmt.Sprintf("  Dominant: %s\n", *titswDominant))
		}
		if titswMode != nil {
			sb.WriteString(fmt.Sprintf("  Mode: %s\n", *titswMode))
		}
		if titswPattern != nil {
			sb.WriteString(fmt.Sprintf("  Pattern: %s\n", *titswPattern))
		}
		sb.WriteString(fmt.Sprintf("  Scores: teach=%d help=%d love=%d spirit=%d doctrine=%d invite=%d\n",
			*titswTeach, *titswHelp, *titswLove, *titswSpirit, *titswDoctrine, *titswInvite))
		if titswSummary != nil {
			sb.WriteString(fmt.Sprintf("  Summary: %s\n", *titswSummary))
		}
		if titswKeyQuote != nil {
			sb.WriteString(fmt.Sprintf("  Key Quote: %s\n", *titswKeyQuote))
		}
		if titswKeywords != nil {
			sb.WriteString(fmt.Sprintf("  Keywords: %s\n", *titswKeywords))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(content)
	return sb.String(), nil
}

// crossRef represents a cross-reference result.
type crossRef struct {
	Reference string
	Type      string
}

// getCrossReferences returns cross-references for a single verse.
func getCrossReferences(database *db.DB, book string, chapter, verse int) []crossRef {
	rows, err := database.Query(`
		SELECT DISTINCT target_book, target_chapter, target_verse, reference_type
		FROM cross_references
		WHERE source_book = ? AND source_chapter = ? AND source_verse = ?
		LIMIT 20
	`, book, chapter, verse)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var refs []crossRef
	for rows.Next() {
		var targetBook, refType string
		var targetChapter int
		var targetVerse *int

		if err := rows.Scan(&targetBook, &targetChapter, &targetVerse, &refType); err != nil {
			continue
		}

		v := 0
		if targetVerse != nil {
			v = *targetVerse
		}

		refs = append(refs, crossRef{
			Reference: formatScriptureRef(targetBook, targetChapter, v),
			Type:      refType,
		})
	}
	return refs
}

// getCrossReferencesForRange returns deduplicated cross-references for a verse range.
func getCrossReferencesForRange(database *db.DB, book string, chapter, startVerse, endVerse int) []crossRef {
	rows, err := database.Query(`
		SELECT DISTINCT target_book, target_chapter, target_verse, reference_type
		FROM cross_references
		WHERE source_book = ? AND source_chapter = ? AND source_verse >= ? AND source_verse <= ?
		ORDER BY target_book, target_chapter, target_verse
		LIMIT 50
	`, book, chapter, startVerse, endVerse)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var refs []crossRef
	for rows.Next() {
		var targetBook, refType string
		var targetChapter int
		var targetVerse *int

		if err := rows.Scan(&targetBook, &targetChapter, &targetVerse, &refType); err != nil {
			continue
		}

		v := 0
		if targetVerse != nil {
			v = *targetVerse
		}

		refs = append(refs, crossRef{
			Reference: formatScriptureRef(targetBook, targetChapter, v),
			Type:      refType,
		})
	}
	return refs
}

func handleList(database *db.DB, store vec.Searcher, args json.RawMessage) (string, error) {
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
