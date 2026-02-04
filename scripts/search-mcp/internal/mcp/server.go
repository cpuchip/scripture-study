// Package mcp implements the MCP protocol for the search server.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/search-mcp/internal/ddg"
)

// JSON-RPC message types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Server is the MCP server.
type Server struct {
	ddgClient *ddg.Client
	input     io.Reader
	output    io.Writer
}

// NewServer creates a new MCP server.
func NewServer() *Server {
	return &Server{
		ddgClient: ddg.NewClient(),
		input:     os.Stdin,
		output:    os.Stdout,
	}
}

// Run starts the MCP server.
func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.input)
	// Increase buffer size for large messages
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(&req)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "initialized":
		// No response needed
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(req)
	case "shutdown":
		s.sendResult(req.ID, nil)
		os.Exit(0)
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *Server) handleInitialize(req *Request) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "search-mcp",
			"version": "0.1.0",
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}
	s.sendResult(req.ID, result)
}

func (s *Server) handleToolsList(req *Request) {
	tools := []map[string]any{
		{
			"name":        "web_search",
			"description": "Search the web using DuckDuckGo. Returns relevant web results with titles, URLs, and snippets. Use this to find current information, articles, documentation, or any web content.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query. Be specific for better results.",
					},
					"max_results": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10, max: 25)",
						"default":     10,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "news_search",
			"description": "Search for recent news articles using DuckDuckGo News. Returns news results with titles, URLs, dates, and snippets. Use this for current events, recent developments, or time-sensitive information.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The news search query.",
					},
					"max_results": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10, max: 25)",
						"default":     10,
					},
					"timelimit": map[string]any{
						"type":        "string",
						"description": "Time limit for news: 'd' (day), 'w' (week), 'm' (month). Default: None (all time)",
						"enum":        []string{"d", "w", "m"},
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "instant_answer",
			"description": "Get an instant answer from DuckDuckGo for factual queries. Best for definitions, quick facts, calculations, or simple questions. Returns a direct answer if available, otherwise suggests web search.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The question or factual query.",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	s.sendResult(req.ID, map[string]any{"tools": tools})
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (s *Server) handleToolsCall(req *Request) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	ctx := context.Background()

	switch params.Name {
	case "web_search":
		s.handleWebSearch(ctx, req.ID, params.Arguments)
	case "news_search":
		s.handleNewsSearch(ctx, req.ID, params.Arguments)
	case "instant_answer":
		s.handleInstantAnswer(ctx, req.ID, params.Arguments)
	default:
		s.sendError(req.ID, -32602, "Unknown tool", params.Name)
	}
}

func (s *Server) handleWebSearch(ctx context.Context, id any, args map[string]any) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		s.sendError(id, -32602, "Invalid params", "query is required")
		return
	}

	maxResults := 10
	if v, ok := args["max_results"].(float64); ok {
		maxResults = int(v)
	}
	if maxResults > 25 {
		maxResults = 25
	}

	results, err := s.ddgClient.WebSearch(ctx, query, maxResults)
	if err != nil {
		s.sendToolError(id, fmt.Sprintf("Search failed: %v", err))
		return
	}

	if len(results) == 0 {
		s.sendToolResult(id, "No results found for your search query.")
		return
	}

	// Format results as text
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results for '%s':\n\n", len(results), query))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		}
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("   %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}

	s.sendToolResult(id, sb.String())
}

func (s *Server) handleNewsSearch(ctx context.Context, id any, args map[string]any) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		s.sendError(id, -32602, "Invalid params", "query is required")
		return
	}

	maxResults := 10
	if v, ok := args["max_results"].(float64); ok {
		maxResults = int(v)
	}
	if maxResults > 25 {
		maxResults = 25
	}

	timeLimitDays := 30 // Default to month
	if tl, ok := args["timelimit"].(string); ok {
		switch tl {
		case "d":
			timeLimitDays = 1
		case "w":
			timeLimitDays = 7
		case "m":
			timeLimitDays = 30
		}
	}

	results, err := s.ddgClient.NewsSearch(ctx, query, maxResults, timeLimitDays)
	if err != nil {
		s.sendToolError(id, fmt.Sprintf("News search failed: %v", err))
		return
	}

	if len(results) == 0 {
		s.sendToolResult(id, "No news results found for your search query.")
		return
	}

	// Format results as text
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d news results for '%s':\n\n", len(results), query))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		}
		if r.Date != "" {
			sb.WriteString(fmt.Sprintf("   Date: %s\n", r.Date))
		}
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("   %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}

	s.sendToolResult(id, sb.String())
}

func (s *Server) handleInstantAnswer(ctx context.Context, id any, args map[string]any) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		s.sendError(id, -32602, "Invalid params", "query is required")
		return
	}

	answer, err := s.ddgClient.GetInstantAnswer(ctx, query)
	if err != nil {
		s.sendToolError(id, fmt.Sprintf("Instant answer failed: %v", err))
		return
	}

	if answer == nil {
		s.sendToolResult(id, "No instant answer available for this query. Try using web_search for more detailed results.")
		return
	}

	// Format the answer
	var sb strings.Builder
	if answer.Answer != "" {
		sb.WriteString(fmt.Sprintf("Answer: %s\n", answer.Answer))
		if answer.AnswerType != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", answer.AnswerType))
		}
	}
	if answer.Definition != "" {
		sb.WriteString(fmt.Sprintf("Definition: %s\n", answer.Definition))
		if answer.DefinitionURL != "" {
			sb.WriteString(fmt.Sprintf("Source: %s\n", answer.DefinitionURL))
		}
	}
	if answer.AbstractText != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", answer.AbstractText))
		if answer.AbstractSource != "" {
			sb.WriteString(fmt.Sprintf("Source: %s", answer.AbstractSource))
			if answer.AbstractURL != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", answer.AbstractURL))
			}
			sb.WriteString("\n")
		}
	}

	if sb.Len() == 0 {
		s.sendToolResult(id, "No instant answer available for this query. Try using web_search for more detailed results.")
		return
	}

	s.sendToolResult(id, sb.String())
}

func (s *Server) sendResult(id any, result any) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendResponse(resp)
}

func (s *Server) sendError(id any, code int, message string, data any) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.sendResponse(resp)
}

func (s *Server) sendToolResult(id any, text string) {
	result := map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
	}
	s.sendResult(id, result)
}

func (s *Server) sendToolError(id any, text string) {
	result := map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
		"isError": true,
	}
	s.sendResult(id, result)
}

func (s *Server) sendResponse(resp Response) {
	data, _ := json.Marshal(resp)
	fmt.Fprintln(s.output, string(data))
}
