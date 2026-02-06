package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MCP JSON-RPC types
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCP Server implementation
type MCPServer struct {
	store    *Store
	searcher *Searcher
	cfg      *Config
}

// NewMCPServer creates a new MCP server
func NewMCPServer(cfg *Config) (*MCPServer, error) {
	embedFunc := NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

	store, err := NewStore(cfg, embedFunc)
	if err != nil {
		return nil, fmt.Errorf("creating store: %w", err)
	}

	return &MCPServer{
		store:    store,
		searcher: NewSearcher(store),
		cfg:      cfg,
	}, nil
}

// Run starts the MCP server on stdin/stdout
func (s *MCPServer) Run() error {
	reader := bufio.NewReader(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		var req MCPRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(encoder, nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(encoder, &req)
	}
}

func (s *MCPServer) handleRequest(enc *json.Encoder, req *MCPRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(enc, req)
	case "tools/list":
		s.handleToolsList(enc, req)
	case "tools/call":
		s.handleToolsCall(enc, req)
	case "notifications/initialized":
		// Client notification, no response needed
	default:
		s.sendError(enc, req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *MCPServer) handleInitialize(enc *json.Encoder, req *MCPRequest) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "gospel-vec",
			"version": "0.1.0",
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}
	s.sendResult(enc, req.ID, result)
}

func (s *MCPServer) handleToolsList(enc *json.Encoder, req *MCPRequest) {
	tools := []map[string]any{
		{
			"name":        "search_scriptures",
			"description": "Search the scriptures using semantic similarity. Finds verses, paragraphs, chapter summaries, and themes related to the query. Searches across scriptures, conference talks, manuals, and books.\n\nIMPORTANT: Results labeled [AI SUMMARY] or [AI THEME] are NOT direct quotes — always verify against the source file before quoting. Results include file paths and markdown links for easy follow-up with read_file.\n\nTip: After finding relevant content, use get_chapter or get_talk to read the full source text.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query (e.g., 'faith in Christ', 'repentance', 'creation of the world')",
					},
					"layers": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string", "enum": []string{"verse", "paragraph", "summary", "theme"}},
						"description": "Which layers to search (default: all available)",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum results per layer (default: 5)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "get_chapter",
			"description": "Get the full text of a specific scripture chapter.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book": map[string]any{
						"type":        "string",
						"description": "The book name. Accepts various formats: '1 Nephi', '1-ne', '1nephi', 'D&C', 'dc', 'Alma', etc.",
					},
					"chapter": map[string]any{
						"type":        "integer",
						"description": "The chapter number",
					},
				},
				"required": []string{"book", "chapter"},
			},
		},
		{
			"name":        "list_books",
			"description": "List all books available in the scripture index. Use to discover what books can be searched or retrieved.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"volume": map[string]any{
						"type":        "string",
						"description": "Optional: filter by volume ('bofm', 'dc', 'pgp', 'ot', 'nt'). If omitted, lists all indexed books.",
					},
				},
			},
		},
		{
			"name":        "get_talk",
			"description": "Get the full text of a conference talk. Use after search_scriptures finds a relevant talk, or when you know the speaker and year. Always use this to read the actual talk before quoting.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"speaker": map[string]any{
						"type":        "string",
						"description": "Speaker last name (e.g., 'nelson', 'hinckley', 'oaks')",
					},
					"year": map[string]any{
						"type":        "integer",
						"description": "Conference year (e.g., 2001)",
					},
					"month": map[string]any{
						"type":        "string",
						"description": "Conference month: '04' for April, '10' for October",
					},
					"file_path": map[string]any{
						"type":        "string",
						"description": "Direct file path if known from search results. If provided, speaker/year/month are ignored.",
					},
				},
			},
		},
		{
			"name":        "search_talks",
			"description": "Search conference talks with optional speaker and year filters. Returns semantic matches from indexed talks only (not scriptures or manuals). Use for targeted conference talk discovery.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Semantic search query (e.g., 'secret combinations in modern times')",
					},
					"speaker": map[string]any{
						"type":        "string",
						"description": "Filter by speaker last name (case-insensitive)",
					},
					"year_from": map[string]any{
						"type":        "integer",
						"description": "Start year, inclusive (e.g., 2000)",
					},
					"year_to": map[string]any{
						"type":        "integer",
						"description": "End year, inclusive (e.g., 2010)",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum results (default: 10)",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	s.sendResult(enc, req.ID, map[string]any{"tools": tools})
}

func (s *MCPServer) handleToolsCall(enc *json.Encoder, req *MCPRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(enc, req.ID, -32602, "Invalid params", err.Error())
		return
	}

	switch params.Name {
	case "search_scriptures":
		s.toolSearchScriptures(enc, req.ID, params.Arguments)
	case "get_chapter":
		s.toolGetChapter(enc, req.ID, params.Arguments)
	case "list_books":
		s.toolListBooks(enc, req.ID, params.Arguments)
	case "get_talk":
		s.toolGetTalk(enc, req.ID, params.Arguments)
	case "search_talks":
		s.toolSearchTalks(enc, req.ID, params.Arguments)
	default:
		s.sendError(enc, req.ID, -32602, "Unknown tool", params.Name)
	}
}

func (s *MCPServer) toolSearchScriptures(enc *json.Encoder, id any, args json.RawMessage) {
	var params struct {
		Query  string   `json:"query"`
		Layers []string `json:"layers"`
		Limit  int      `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(enc, id, -32602, "Invalid arguments", err.Error())
		return
	}

	// Default limit
	if params.Limit <= 0 {
		params.Limit = 5
	}

	// Parse layers
	layers := []Layer{}
	if len(params.Layers) == 0 {
		// Search all available layers
		layers = []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme}
	} else {
		for _, l := range params.Layers {
			switch l {
			case "verse":
				layers = append(layers, LayerVerse)
			case "paragraph":
				layers = append(layers, LayerParagraph)
			case "summary":
				layers = append(layers, LayerSummary)
			case "theme":
				layers = append(layers, LayerTheme)
			}
		}
	}

	// Search
	ctx := context.Background()
	results, err := s.searcher.Search(ctx, params.Query, SearchOptions{
		Layers: layers,
		Limit:  params.Limit,
	})
	if err != nil {
		s.sendError(enc, id, -32000, "Search failed", err.Error())
		return
	}

	// Format results for MCP
	var formattedResults []map[string]any
	for _, r := range results {
		formattedResults = append(formattedResults, map[string]any{
			"reference":  r.Metadata.Reference,
			"layer":      string(r.Metadata.Layer),
			"content":    r.Content,
			"similarity": r.Score,
		})
	}

	s.sendResult(enc, id, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": formatMCPSearchResults(results),
			},
		},
	})
}

func (s *MCPServer) toolGetChapter(enc *json.Encoder, id any, args json.RawMessage) {
	var params struct {
		Book    string `json:"book"`
		Chapter int    `json:"chapter"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(enc, id, -32602, "Invalid arguments", err.Error())
		return
	}

	// Normalize the book name to handle various input formats
	normalizedBook := NormalizeBookName(params.Book)

	// Find the chapter file
	files, err := FindScriptureFiles(s.cfg.ScripturesPath, "bofm", "dc-testament/dc", "pgp", "nt", "ot")
	if err != nil {
		s.sendError(enc, id, -32000, "Failed to find scriptures", err.Error())
		return
	}

	// Track which books we've seen for better error messages
	var seenBooks = make(map[string]bool)

	// Look for matching chapter
	for _, f := range files {
		chapter, err := ParseChapterFile(f)
		if err != nil {
			continue
		}

		seenBooks[chapter.Book] = true

		if chapter.Book == normalizedBook && chapter.Chapter == params.Chapter {
			// Format chapter content
			var content string
			for _, v := range chapter.Verses {
				content += fmt.Sprintf("%d. %s\n\n", v.Number, v.Text)
			}

			s.sendResult(enc, id, map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": fmt.Sprintf("# %s %d\n\n%s", chapter.Book, chapter.Chapter, content),
					},
				},
			})
			return
		}
	}

	// Build helpful error message with suggestions
	var availableBooks []string
	for book := range seenBooks {
		availableBooks = append(availableBooks, book)
	}

	errMsg := fmt.Sprintf("Chapter not found: %s %d", params.Book, params.Chapter)
	if normalizedBook != params.Book {
		errMsg += fmt.Sprintf(" (interpreted as '%s')", normalizedBook)
	}
	errMsg += fmt.Sprintf("\n\nAvailable books in index: %s", strings.Join(availableBooks, ", "))
	errMsg += "\n\nTip: Try using full book names like '1 Nephi', 'D&C', 'Alma' or abbreviations like '1-ne', 'dc', 'mosiah'"

	s.sendError(enc, id, -32000, "Chapter not found", errMsg)
}

func (s *MCPServer) toolListBooks(enc *json.Encoder, id any, args json.RawMessage) {
	var params struct {
		Volume string `json:"volume"`
	}
	// Args can be empty/null for this tool
	if args != nil && len(args) > 0 {
		json.Unmarshal(args, &params)
	}

	// Find all indexed books by scanning the database
	files, err := FindScriptureFiles(s.cfg.ScripturesPath, "bofm", "dc-testament/dc", "pgp", "nt", "ot")
	if err != nil {
		s.sendError(enc, id, -32000, "Failed to find scriptures", err.Error())
		return
	}

	// Group books by volume
	booksByVolume := make(map[string][]string)
	seenBooks := make(map[string]map[string]bool)

	for _, f := range files {
		chapter, err := ParseChapterFile(f)
		if err != nil {
			continue
		}

		// Determine volume from path
		volume := "unknown"
		if strings.Contains(f, "bofm") {
			volume = "bofm"
		} else if strings.Contains(f, "dc-testament") {
			volume = "dc"
		} else if strings.Contains(f, "pgp") {
			volume = "pgp"
		} else if strings.Contains(f, filepath.Join("scriptures", "ot")) {
			volume = "ot"
		} else if strings.Contains(f, filepath.Join("scriptures", "nt")) {
			volume = "nt"
		}

		if seenBooks[volume] == nil {
			seenBooks[volume] = make(map[string]bool)
		}

		if !seenBooks[volume][chapter.Book] {
			seenBooks[volume][chapter.Book] = true
			booksByVolume[volume] = append(booksByVolume[volume], chapter.Book)
		}
	}

	// Build output
	var output strings.Builder

	// Volume names
	volumeNames := map[string]string{
		"bofm": "Book of Mormon",
		"dc":   "Doctrine & Covenants",
		"pgp":  "Pearl of Great Price",
		"ot":   "Old Testament",
		"nt":   "New Testament",
	}

	volumeOrder := []string{"bofm", "dc", "pgp", "ot", "nt"}

	// Filter by volume if specified
	if params.Volume != "" {
		normalizedVolume := strings.ToLower(params.Volume)
		if books, ok := booksByVolume[normalizedVolume]; ok {
			output.WriteString(fmt.Sprintf("# %s\n\n", volumeNames[normalizedVolume]))
			for _, book := range books {
				output.WriteString(fmt.Sprintf("- %s\n", book))
			}
		} else {
			output.WriteString(fmt.Sprintf("Volume '%s' not found. Available: bofm, dc, pgp, ot, nt\n", params.Volume))
		}
	} else {
		// List all volumes
		output.WriteString("# Indexed Scripture Books\n\n")
		for _, vol := range volumeOrder {
			if books, ok := booksByVolume[vol]; ok && len(books) > 0 {
				output.WriteString(fmt.Sprintf("## %s\n", volumeNames[vol]))
				for _, book := range books {
					output.WriteString(fmt.Sprintf("- %s\n", book))
				}
				output.WriteString("\n")
			}
		}
		output.WriteString("---\n")
		output.WriteString("Tip: Use any of these book names with get_chapter(book, chapter)\n")
	}

	s.sendResult(enc, id, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": output.String(),
			},
		},
	})
}

func (s *MCPServer) toolGetTalk(enc *json.Encoder, id any, args json.RawMessage) {
	var params struct {
		Speaker  string `json:"speaker"`
		Year     int    `json:"year"`
		Month    string `json:"month"`
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(enc, id, -32602, "Invalid arguments", err.Error())
		return
	}

	// If file_path is provided, read it directly
	if params.FilePath != "" {
		content, err := readTalkFromPath(params.FilePath)
		if err != nil {
			s.sendError(enc, id, -32000, "Failed to read talk", err.Error())
			return
		}
		s.sendResult(enc, id, map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": content,
				},
			},
		})
		return
	}

	// Otherwise search by speaker/year/month
	if params.Speaker == "" {
		s.sendError(enc, id, -32602, "Either file_path or speaker is required", "")
		return
	}

	// Build list of years to search
	years := []string{}
	if params.Year > 0 {
		years = append(years, fmt.Sprintf("%d", params.Year))
	}

	talkFiles, err := FindTalkFiles(s.cfg.ConferencePath, years...)
	if err != nil {
		s.sendError(enc, id, -32000, "Failed to find talks", err.Error())
		return
	}

	// Filter by speaker and optionally month
	speakerLower := strings.ToLower(params.Speaker)
	var matches []string
	for _, f := range talkFiles {
		filename := strings.ToLower(filepath.Base(f))
		// Check if speaker name appears in filename
		if strings.Contains(filename, speakerLower) {
			// Filter by month if specified
			if params.Month != "" {
				if !strings.Contains(f, string(filepath.Separator)+params.Month+string(filepath.Separator)) {
					continue
				}
			}
			matches = append(matches, f)
		}
	}

	if len(matches) == 0 {
		errMsg := fmt.Sprintf("No talks found for speaker '%s'", params.Speaker)
		if params.Year > 0 {
			errMsg += fmt.Sprintf(" in %d", params.Year)
		}
		if params.Month != "" {
			monthName := "April"
			if params.Month == "10" {
				monthName = "October"
			}
			errMsg += fmt.Sprintf(" (%s)", monthName)
		}
		s.sendError(enc, id, -32000, "Talk not found", errMsg)
		return
	}

	if len(matches) == 1 {
		// Single match — return the full talk
		content, err := readTalkFromPath(matches[0])
		if err != nil {
			s.sendError(enc, id, -32000, "Failed to read talk", err.Error())
			return
		}
		s.sendResult(enc, id, map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": content,
				},
			},
		})
		return
	}

	// Multiple matches — list them so user can choose
	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Found %d talks matching '%s'\n\n", len(matches), params.Speaker))
	output.WriteString("Use `get_talk(file_path: \"...\")` to read a specific talk:\n\n")
	for _, f := range matches {
		talk, err := ParseTalkFile(f)
		if err != nil {
			output.WriteString(fmt.Sprintf("- %s\n", f))
			continue
		}
		relPath := strings.ReplaceAll(f, "\\", "/")
		output.WriteString(fmt.Sprintf("- **\"%s\"** by %s (%s %s)\n  `file_path: \"%s\"`\n",
			talk.Metadata.Title, talk.Metadata.Speaker,
			talk.Metadata.Month, talk.Metadata.Year, relPath))
	}

	s.sendResult(enc, id, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": output.String(),
			},
		},
	})
}

func (s *MCPServer) toolSearchTalks(enc *json.Encoder, id any, args json.RawMessage) {
	var params struct {
		Query    string `json:"query"`
		Speaker  string `json:"speaker"`
		YearFrom int    `json:"year_from"`
		YearTo   int    `json:"year_to"`
		Limit    int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(enc, id, -32602, "Invalid arguments", err.Error())
		return
	}

	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Search conference sources only, paragraph and summary layers
	ctx := context.Background()
	results, err := s.searcher.Search(ctx, params.Query, SearchOptions{
		Layers:  []Layer{LayerParagraph, LayerSummary},
		Sources: []Source{SourceConference},
		Limit:   params.Limit * 3, // Over-fetch to allow for filtering
	})
	if err != nil {
		s.sendError(enc, id, -32000, "Search failed", err.Error())
		return
	}

	// Post-filter by speaker and year range
	var filtered []SearchResult
	speakerLower := strings.ToLower(params.Speaker)
	for _, r := range results {
		if r.Metadata.Source != SourceConference {
			continue
		}
		// Filter by speaker
		if params.Speaker != "" && !strings.Contains(strings.ToLower(r.Metadata.Speaker), speakerLower) {
			continue
		}
		// Filter by year range
		if params.YearFrom > 0 && r.Metadata.Year < params.YearFrom {
			continue
		}
		if params.YearTo > 0 && r.Metadata.Year > params.YearTo {
			continue
		}
		filtered = append(filtered, r)
		if len(filtered) >= params.Limit {
			break
		}
	}

	s.sendResult(enc, id, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": formatMCPSearchResults(filtered),
			},
		},
	})
}

// readTalkFromPath reads a talk file from disk, trying several path resolutions
func readTalkFromPath(filePath string) (string, error) {
	// Normalize path
	normalized := filepath.FromSlash(filePath)

	// Try paths in order
	candidates := []string{
		normalized,
		filepath.Join(".", normalized),
	}

	// Strip leading ../ and try from cwd
	stripped := strings.TrimPrefix(normalized, ".."+string(filepath.Separator))
	candidates = append(candidates, stripped)

	// Try from two directories up (scripts/gospel-vec -> repo root)
	candidates = append(candidates, filepath.Join("..", "..", stripped))

	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("file not found: %s (tried %d locations)", filePath, len(candidates))
}

func (s *MCPServer) sendResult(enc *json.Encoder, id any, result any) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	enc.Encode(resp)
}

func (s *MCPServer) sendError(enc *json.Encoder, id any, code int, message, data string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	enc.Encode(resp)
}

func formatMCPSearchResults(results []SearchResult) string {
	if len(results) == 0 {
		return "No results found."
	}

	var output string
	currentLayer := Layer("")

	for _, r := range results {
		if r.Metadata.Layer != currentLayer {
			currentLayer = r.Metadata.Layer
			output += fmt.Sprintf("\n## %s Results\n\n", currentLayer)
		}

		// Result type label based on layer
		typeLabel := resultTypeLabel(r.Metadata.Layer)

		// Format header based on source type
		if r.Metadata.Source == SourceConference && r.Metadata.Speaker != "" {
			// Rich conference talk header
			title := r.Metadata.TalkTitle
			if title == "" {
				title = r.Metadata.Reference
			}
			output += fmt.Sprintf("**%s** — \"%s\" (%.0f%% match) %s\n",
				r.Metadata.Speaker, title, r.Score*100, typeLabel)
			if r.Metadata.Position != "" || r.Metadata.Session != "" {
				var details []string
				if r.Metadata.Position != "" {
					details = append(details, r.Metadata.Position)
				}
				monthName := "April"
				if r.Metadata.Month == "10" {
					monthName = "October"
				}
				details = append(details, fmt.Sprintf("%s %d General Conference", monthName, r.Metadata.Year))
				if r.Metadata.Session != "" {
					details = append(details, r.Metadata.Session+" Session")
				}
				output += fmt.Sprintf("*%s*\n", strings.Join(details, ", "))
			}
		} else {
			// Standard header for scriptures/manual
			output += fmt.Sprintf("**%s** (%.0f%% match) %s\n", r.Metadata.Reference, r.Score*100, typeLabel)
		}

		// File link with existence check
		if r.Metadata.FilePath != "" {
			link := buildMarkdownLink(r.Metadata)
			exists := checkFileExists(r.Metadata.FilePath)
			if exists {
				output += fmt.Sprintf("📎 %s ✅ local file available\n", link)
			} else {
				output += fmt.Sprintf("📎 %s ❌ not cached locally\n", link)
			}
		}

		output += fmt.Sprintf("> %s\n\n", truncate(r.Content, 300))
	}

	output += "---\n"
	output += "💡 **Reminder:** Always `read_file` the source before quoting. Search results are pointers, not sources.\n"

	return output
}

// resultTypeLabel returns a label indicating whether the result is a direct quote or AI-generated
func resultTypeLabel(layer Layer) string {
	switch layer {
	case LayerVerse:
		return "[DIRECT QUOTE]"
	case LayerParagraph:
		return "[DIRECT QUOTE]"
	case LayerSummary:
		return "[AI SUMMARY — verify against source]"
	case LayerTheme:
		return "[AI THEME — verify against source]"
	default:
		return ""
	}
}

// buildMarkdownLink constructs a relative markdown link from metadata
func buildMarkdownLink(meta *DocMetadata) string {
	if meta.FilePath == "" {
		return meta.Reference
	}

	// Normalize path separators to forward slashes
	relPath := strings.ReplaceAll(meta.FilePath, "\\", "/")

	// Ensure it starts with ../ for relative linking from study/ documents
	if !strings.HasPrefix(relPath, "../") {
		// Strip leading ./ if present
		relPath = strings.TrimPrefix(relPath, "./")
		// If path starts with gospel-library, prepend ../
		if strings.HasPrefix(relPath, "gospel-library") {
			relPath = "../" + relPath
		}
	}

	// Choose display text
	displayText := meta.Reference
	if meta.Source == SourceConference && meta.TalkTitle != "" {
		displayText = meta.TalkTitle
	}

	return fmt.Sprintf("[%s](%s)", displayText, relPath)
}

// checkFileExists checks if a file exists on disk
// It tries the path as-is and also relative to common base directories
func checkFileExists(filePath string) bool {
	// Normalize path
	normalized := strings.ReplaceAll(filePath, "/", string(filepath.Separator))

	// Try as-is
	if _, err := os.Stat(normalized); err == nil {
		return true
	}

	// Try relative from current working directory
	if _, err := os.Stat(filepath.Join(".", normalized)); err == nil {
		return true
	}

	// Try stripping leading ../ and checking from cwd
	stripped := strings.TrimPrefix(normalized, "..")
	stripped = strings.TrimPrefix(stripped, string(filepath.Separator))
	if _, err := os.Stat(stripped); err == nil {
		return true
	}

	// Try from two directories up (scripts/gospel-vec -> repo root)
	if _, err := os.Stat(filepath.Join("..", "..", stripped)); err == nil {
		return true
	}

	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "... [TRUNCATED — use read_file or get_chapter for full text]"
}
