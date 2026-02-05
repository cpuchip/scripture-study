package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
			"description": "Search the scriptures using semantic similarity. Finds verses, paragraphs, chapter summaries, and themes related to the query.",
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
						"description": "The book name (e.g., '1 Nephi', 'Mosiah', 'D&C')",
					},
					"chapter": map[string]any{
						"type":        "integer",
						"description": "The chapter number",
					},
				},
				"required": []string{"book", "chapter"},
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

	// Find the chapter file
	files, err := FindScriptureFiles(s.cfg.ScripturesPath, "bofm", "dc-testament/dc", "pgp", "nt", "ot")
	if err != nil {
		s.sendError(enc, id, -32000, "Failed to find scriptures", err.Error())
		return
	}

	// Look for matching chapter
	for _, f := range files {
		chapter, err := ParseChapterFile(f)
		if err != nil {
			continue
		}
		if chapter.Book == params.Book && chapter.Chapter == params.Chapter {
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

	s.sendError(enc, id, -32000, "Chapter not found", fmt.Sprintf("%s %d", params.Book, params.Chapter))
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

		output += fmt.Sprintf("**%s** (%.0f%% match)\n", r.Metadata.Reference, r.Score*100)
		output += fmt.Sprintf("> %s\n\n", truncate(r.Content, 300))
	}

	return output
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
