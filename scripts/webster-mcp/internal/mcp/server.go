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
	mcpServer   *server.MCPServer
	webster     *dictionary.Webster // genuine 1828 American Dictionary
	webster1913 *dictionary.Webster // 1913 Revised Unabridged (optional)
	modernDict  *dictionary.ModernDict
}

// New creates a new MCP server with dictionary tools.
// webster1828Path is required; webster1913Path may be empty (1913 tools
// then report the edition as unavailable).
func New(webster1828Path, webster1913Path string) (*Server, error) {
	webster := dictionary.NewWebster()
	if err := webster.LoadFromFile(webster1828Path); err != nil {
		return nil, fmt.Errorf("failed to load Webster 1828 dictionary: %w", err)
	}

	var webster1913 *dictionary.Webster
	if webster1913Path != "" {
		webster1913 = dictionary.NewWebster()
		if err := webster1913.LoadFromFile(webster1913Path); err != nil {
			return nil, fmt.Errorf("failed to load Webster 1913 dictionary: %w", err)
		}
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"webster-mcp",
		"2.0.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer:   mcpServer,
		webster:     webster,
		webster1913: webster1913,
		modernDict:  dictionary.NewModernDict(),
	}

	// Register tools
	s.registerTools()

	return s, nil
}

// editionDict returns the dictionary for an edition string ("1828" default,
// "1913"), or an error message when that edition is not loaded.
func (s *Server) editionDict(edition string) (*dictionary.Webster, string) {
	switch edition {
	case "", "1828":
		return s.webster, ""
	case "1913":
		if s.webster1913 == nil {
			return nil, "Webster 1913 dictionary is not loaded (start the server with -dict1913)."
		}
		return s.webster1913, ""
	default:
		return nil, fmt.Sprintf("Unknown edition '%s' (use \"1828\" or \"1913\").", edition)
	}
}

// registerTools registers all dictionary tools with the MCP server.
func (s *Server) registerTools() {
	// Webster 1828 definition lookup
	s.mcpServer.AddTool(
		mcp.NewTool("webster_define",
			mcp.WithDescription("Look up a word in Noah Webster's 1828 American Dictionary of the English Language (genuine 1828 text, sourced from the Ellen G. White Estate's full-text preservation). Particularly useful for understanding the language of the King James Bible and early Latter-day Saint scriptures, compiled in the same era."),
			mcp.WithString("word",
				mcp.Required(),
				mcp.Description("The word to look up"),
			),
		),
		s.handleWebsterDefine,
	)

	// Webster 1913 definition lookup
	s.mcpServer.AddTool(
		mcp.NewTool("webster1913_define",
			mcp.WithDescription("Look up a word in Webster's Revised Unabridged Dictionary (1913, via Project Gutenberg). A fine general historical dictionary, 85 years after the 1828 — useful for seeing how meanings shifted across the 19th century. NOT the 1828: for KJV/Restoration-era word study use webster_define."),
			mcp.WithString("word",
				mcp.Required(),
				mcp.Description("The word to look up"),
			),
		),
		s.handleWebster1913Define,
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
			mcp.WithDescription("Look up a word across Webster 1828, Webster 1913, AND the modern dictionary — three points in time, side by side. This is the recommended tool for scripture study: it shows how a word's meaning shifted from the Restoration era through the 19th century to today."),
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
			mcp.WithDescription("Search for words in a Webster dictionary by word pattern. Returns words that match or contain the query. Searches the genuine 1828 by default."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The search query (word or partial word)"),
			),
			mcp.WithNumber("max_results",
				mcp.Description("Maximum number of results to return (default: 20)"),
			),
			mcp.WithString("edition",
				mcp.Description("Dictionary edition: \"1828\" (default) or \"1913\""),
			),
		),
		s.handleWebsterSearch,
	)

	// Search definitions
	s.mcpServer.AddTool(
		mcp.NewTool("webster_search_definitions",
			mcp.WithDescription("Search within definitions in a Webster dictionary. Finds words whose definitions contain the query text. Searches the genuine 1828 by default."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The text to search for within definitions"),
			),
			mcp.WithNumber("max_results",
				mcp.Description("Maximum number of results to return (default: 10)"),
			),
			mcp.WithString("edition",
				mcp.Description("Dictionary edition: \"1828\" (default) or \"1913\""),
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

// handleWebster1913Define handles the webster1913_define tool.
func (s *Server) handleWebster1913Define(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	word, err := request.RequireString("word")
	if err != nil {
		return mcp.NewToolResultError("word parameter is required"), nil
	}

	dict, errMsg := s.editionDict("1913")
	if dict == nil {
		return mcp.NewToolResultText(errMsg), nil
	}

	entries := dict.Lookup(word)
	if entries == nil {
		return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in Webster 1913 dictionary.", word)), nil
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

	// Get modern definition
	modernEntries, modernErr := s.modernDict.Lookup(word)

	// Format combined result
	sb := fmt.Sprintf("# Definitions for: %s\n\n", word)

	sb += "## Webster 1828 Dictionary\n"
	sb += "_Noah Webster's 1828 American Dictionary — the language of the KJV and Restoration era._\n\n"
	if entries := s.webster.Lookup(word); len(entries) > 0 {
		sb += dictionary.FormatEntries(entries)
	} else {
		sb += fmt.Sprintf("_Word '%s' not found in Webster 1828._\n", word)
	}

	sb += "\n---\n\n"

	sb += "## Webster 1913 Dictionary\n"
	sb += "_Webster's Revised Unabridged (1913) — 85 years later, for tracking semantic drift._\n\n"
	if s.webster1913 == nil {
		sb += "_Webster 1913 dictionary not loaded._\n"
	} else if entries := s.webster1913.Lookup(word); len(entries) > 0 {
		sb += dictionary.FormatEntries(entries)
	} else {
		sb += fmt.Sprintf("_Word '%s' not found in Webster 1913._\n", word)
	}

	sb += "\n---\n\n"

	sb += "## Modern Dictionary\n"
	sb += "_Contemporary definitions from the Free Dictionary API._\n\n"
	if len(modernEntries) > 0 {
		sb += dictionary.FormatModernEntries(modernEntries)
	} else if modernErr != nil {
		sb += fmt.Sprintf("_Error: Modern dictionary error: %v_\n", modernErr)
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

	edition, _ := request.GetArguments()["edition"].(string)
	dict, errMsg := s.editionDict(edition)
	if dict == nil {
		return mcp.NewToolResultText(errMsg), nil
	}

	words := dict.Search(query, maxResults)
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

	edition, _ := request.GetArguments()["edition"].(string)
	dict, errMsg := s.editionDict(edition)
	if dict == nil {
		return mcp.NewToolResultText(errMsg), nil
	}

	entries := dict.SearchDefinitions(query, maxResults)
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

// WebsterStats returns statistics about the loaded dictionaries.
func (s *Server) WebsterStats() map[string]interface{} {
	stats := map[string]interface{}{
		"word_count_1828": s.webster.EntryCount(),
	}
	if s.webster1913 != nil {
		stats["word_count_1913"] = s.webster1913.EntryCount()
	}
	return stats
}

// Serve starts the MCP server.
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
