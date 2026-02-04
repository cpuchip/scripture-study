// Package mcp provides the MCP server implementation for the dictionary service.
package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/cpuchip/webster-mcp/internal/dictionary"
)

// Server wraps the MCP server and dictionary services.
type Server struct {
	mcpServer  *server.MCPServer
	webster    *dictionary.Webster
	modernDict *dictionary.ModernDict
}

// New creates a new MCP server with dictionary tools.
func New(websterPath string) (*Server, error) {
	// Load Webster 1828 dictionary
	webster := dictionary.NewWebster()
	if err := webster.LoadFromFile(websterPath); err != nil {
		return nil, fmt.Errorf("failed to load Webster dictionary: %w", err)
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"webster-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer:  mcpServer,
		webster:    webster,
		modernDict: dictionary.NewModernDict(),
	}

	// Register tools
	s.registerTools()

	return s, nil
}

// registerTools registers all dictionary tools with the MCP server.
func (s *Server) registerTools() {
	// Webster 1828 definition lookup
	s.mcpServer.AddTool(
		mcp.NewTool("webster_define",
			mcp.WithDescription("Look up a word in the Webster 1828 dictionary. This dictionary is particularly useful for understanding the language used in the King James Bible and early Latter-day Saint scriptures, as it was compiled during the same era."),
			mcp.WithString("word",
				mcp.Required(),
				mcp.Description("The word to look up"),
			),
		),
		s.handleWebsterDefine,
	)

	// Modern definition lookup
	s.mcpServer.AddTool(
		mcp.NewTool("modern_define",
			mcp.WithDescription("Look up a word in the modern dictionary (via Free Dictionary API). Useful for understanding contemporary usage and comparing with historical definitions."),
			mcp.WithString("word",
				mcp.Required(),
				mcp.Description("The word to look up"),
			),
		),
		s.handleModernDefine,
	)

	// Combined definition lookup
	s.mcpServer.AddTool(
		mcp.NewTool("define",
			mcp.WithDescription("Look up a word in both the Webster 1828 dictionary AND the modern dictionary. Returns both historical and contemporary definitions side by side. This is the recommended tool for scripture study as it shows how word meanings may have shifted over time."),
			mcp.WithString("word",
				mcp.Required(),
				mcp.Description("The word to look up"),
			),
		),
		s.handleDefine,
	)

	// Search words
	s.mcpServer.AddTool(
		mcp.NewTool("webster_search",
			mcp.WithDescription("Search for words in the Webster 1828 dictionary by word pattern. Returns words that match or contain the query."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The search query (word or partial word)"),
			),
			mcp.WithNumber("max_results",
				mcp.Description("Maximum number of results to return (default: 20)"),
			),
		),
		s.handleWebsterSearch,
	)

	// Search definitions
	s.mcpServer.AddTool(
		mcp.NewTool("webster_search_definitions",
			mcp.WithDescription("Search within definitions in the Webster 1828 dictionary. Finds words whose definitions contain the query text."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The text to search for within definitions"),
			),
			mcp.WithNumber("max_results",
				mcp.Description("Maximum number of results to return (default: 10)"),
			),
		),
		s.handleWebsterSearchDefinitions,
	)
}

// handleWebsterDefine handles the webster_define tool.
func (s *Server) handleWebsterDefine(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	word, err := request.RequireString("word")
	if err != nil {
		return mcp.NewToolResultError("word parameter is required"), nil
	}

	entries := s.webster.Lookup(word)
	if entries == nil {
		return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in Webster 1828 dictionary.", word)), nil
	}

	formatted := dictionary.FormatEntries(entries)
	return mcp.NewToolResultText(formatted), nil
}

// handleModernDefine handles the modern_define tool.
func (s *Server) handleModernDefine(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	word, err := request.RequireString("word")
	if err != nil {
		return mcp.NewToolResultError("word parameter is required"), nil
	}

	entries, err := s.modernDict.Lookup(word)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error looking up '%s': %v", word, err)), nil
	}

	if entries == nil {
		return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in modern dictionary.", word)), nil
	}

	formatted := dictionary.FormatModernEntries(entries)
	return mcp.NewToolResultText(formatted), nil
}

// handleDefine handles the define tool (combined lookup).
func (s *Server) handleDefine(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	word, err := request.RequireString("word")
	if err != nil {
		return mcp.NewToolResultError("word parameter is required"), nil
	}

	result := dictionary.CombinedResult{
		Word: word,
	}

	// Get Webster 1828 definition
	websterEntries := s.webster.Lookup(word)
	if len(websterEntries) > 0 {
		result.Webster = &websterEntries[0]
	}

	// Get modern definition
	modernEntries, err := s.modernDict.Lookup(word)
	if err != nil {
		// Don't fail entirely, just note the error
		result.Error = fmt.Sprintf("Modern dictionary error: %v", err)
	} else {
		result.Modern = modernEntries
	}

	// Format combined result
	var sb string
	sb = fmt.Sprintf("# Definitions for: %s\n\n", word)

	sb += "## Webster 1828 Dictionary\n"
	sb += "_Historical definitions from Noah Webster's 1828 dictionary, reflecting the language of scripture._\n\n"
	if result.Webster != nil {
		sb += dictionary.FormatEntry(result.Webster)
	} else {
		sb += fmt.Sprintf("_Word '%s' not found in Webster 1828._\n", word)
	}

	sb += "\n---\n\n"

	sb += "## Modern Dictionary\n"
	sb += "_Contemporary definitions from the Free Dictionary API._\n\n"
	if len(result.Modern) > 0 {
		sb += dictionary.FormatModernEntries(result.Modern)
	} else if result.Error != "" {
		sb += fmt.Sprintf("_Error: %s_\n", result.Error)
	} else {
		sb += fmt.Sprintf("_Word '%s' not found in modern dictionary._\n", word)
	}

	return mcp.NewToolResultText(sb), nil
}

// handleWebsterSearch handles the webster_search tool.
func (s *Server) handleWebsterSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	maxResults := 20
	if mr, ok := request.GetArguments()["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
	}

	words := s.webster.Search(query, maxResults)
	if len(words) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No words found matching '%s'.", query)), nil
	}

	resultText := fmt.Sprintf("Found %d words matching '%s':\n\n", len(words), query)
	for _, w := range words {
		resultText += fmt.Sprintf("- %s\n", w)
	}

	return mcp.NewToolResultText(resultText), nil
}

// handleWebsterSearchDefinitions handles the webster_search_definitions tool.
func (s *Server) handleWebsterSearchDefinitions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	maxResults := 10
	if mr, ok := request.GetArguments()["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
	}

	entries := s.webster.SearchDefinitions(query, maxResults)
	if len(entries) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No definitions found containing '%s'.", query)), nil
	}

	resultText := fmt.Sprintf("Found %d entries with definitions containing '%s':\n\n", len(entries), query)
	for _, entry := range entries {
		resultText += fmt.Sprintf("### %s (%s)\n", entry.Word, entry.POS)
		for i, def := range entry.Definitions {
			resultText += fmt.Sprintf("%d. %s\n", i+1, def)
		}
		resultText += "\n"
	}

	return mcp.NewToolResultText(resultText), nil
}

// WebsterStats returns statistics about the loaded dictionary.
func (s *Server) WebsterStats() map[string]interface{} {
	return map[string]interface{}{
		"word_count": s.webster.EntryCount(),
	}
}

// Serve starts the MCP server.
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
