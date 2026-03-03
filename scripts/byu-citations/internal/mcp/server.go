// Package mcp provides the MCP server for the BYU Scripture Citation Index.
package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/cpuchip/scripture-study/scripts/byu-citations/internal/citations"
)

// Server wraps the MCP server and citation client.
type Server struct {
	mcpServer *server.MCPServer
	client    *citations.Client
}

// New creates a new MCP server with citation tools.
func New() *Server {
	mcpServer := server.NewMCPServer(
		"byu-citations",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer: mcpServer,
		client:    citations.NewClient(),
	}

	s.registerTools()
	return s
}

// registerTools registers all citation tools with the MCP server.
func (s *Server) registerTools() {
	// Look up citations for a single verse
	s.mcpServer.AddTool(
		mcp.NewTool("byu_citations",
			mcp.WithDescription(
				"Look up who has cited a scripture verse in General Conference, Journal of Discourses, "+
					"and other indexed sources. Uses the BYU Scripture Citation Index (scriptures.byu.edu). "+
					"Returns speakers, talk titles, and references. Accepts standard scripture references "+
					"like '3 Nephi 21:10', 'D&C 113:6', 'Isaiah 11:1', 'Alma 32:21'."),
			mcp.WithString("reference",
				mcp.Required(),
				mcp.Description("Scripture reference to look up, e.g. '3 Nephi 21:10', 'D&C 113:6', 'Isaiah 11:1-3'"),
			),
		),
		s.handleCitations,
	)

	// Look up citations for multiple verses at once
	s.mcpServer.AddTool(
		mcp.NewTool("byu_citations_bulk",
			mcp.WithDescription(
				"Look up citations for multiple scripture references at once. "+
					"Returns a summary showing how many citations each verse has, plus full details. "+
					"Useful for surveying which verses in a chapter have been discussed in conference."),
			mcp.WithString("references",
				mcp.Required(),
				mcp.Description("Comma-separated scripture references, e.g. 'Isaiah 11:1, Isaiah 11:10, D&C 113:6'"),
			),
		),
		s.handleBulkCitations,
	)

	// List known books
	s.mcpServer.AddTool(
		mcp.NewTool("byu_citations_books",
			mcp.WithDescription("List all books and their BYU Citation Index IDs. Useful for debugging or discovering available books."),
		),
		s.handleListBooks,
	)
}

// handleCitations handles the byu_citations tool.
func (s *Server) handleCitations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reference, err := request.RequireString("reference")
	if err != nil {
		return mcp.NewToolResultError("reference parameter is required"), nil
	}

	result, err := s.client.Lookup(reference)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error looking up %q: %v", reference, err)), nil
	}

	return mcp.NewToolResultText(citations.FormatResult(result)), nil
}

// handleBulkCitations handles the byu_citations_bulk tool.
func (s *Server) handleBulkCitations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	refsStr, err := request.RequireString("references")
	if err != nil {
		return mcp.NewToolResultError("references parameter is required"), nil
	}

	refs := strings.Split(refsStr, ",")
	var sb strings.Builder
	sb.WriteString("## BYU Citation Index — Bulk Lookup\n\n")

	// Summary table first
	sb.WriteString("| Reference | Citations |\n|---|---|\n")
	var results []*citations.LookupResult
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		result, err := s.client.Lookup(ref)
		if err != nil {
			sb.WriteString(fmt.Sprintf("| %s | Error: %v |\n", ref, err))
			continue
		}
		results = append(results, result)
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", result.Scripture, len(result.Citations)))
	}

	// Then full details for each
	sb.WriteString("\n---\n\n")
	for _, result := range results {
		sb.WriteString(citations.FormatResult(result))
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleListBooks handles the byu_citations_books tool.
func (s *Server) handleListBooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("## BYU Citation Index — Book IDs\n\n")

	sections := []struct {
		name  string
		books []string
	}{
		{"Old Testament", []string{
			"Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy",
			"Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel",
			"1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles",
			"Ezra", "Nehemiah", "Esther", "Job", "Psalms", "Proverbs",
			"Ecclesiastes", "Song of Solomon", "Isaiah", "Jeremiah",
			"Lamentations", "Ezekiel", "Daniel", "Hosea", "Joel", "Amos",
			"Obadiah", "Jonah", "Micah", "Nahum", "Habakkuk", "Zephaniah",
			"Haggai", "Zechariah", "Malachi",
		}},
		{"New Testament", []string{
			"Matthew", "Mark", "Luke", "John", "Acts",
			"Romans", "1 Corinthians", "2 Corinthians", "Galatians",
			"Ephesians", "Philippians", "Colossians",
			"1 Thessalonians", "2 Thessalonians",
			"1 Timothy", "2 Timothy", "Titus", "Philemon",
			"Hebrews", "James", "1 Peter", "2 Peter",
			"1 John", "2 John", "3 John", "Jude", "Revelation",
		}},
		{"Book of Mormon", []string{
			"1 Nephi", "2 Nephi", "Jacob", "Enos", "Jarom", "Omni",
			"Words of Mormon", "Mosiah", "Alma", "Helaman",
			"3 Nephi", "4 Nephi", "Mormon", "Ether", "Moroni",
		}},
		{"Doctrine and Covenants", []string{"D&C", "O.D."}},
		{"Pearl of Great Price", []string{
			"Moses", "Abraham", "Facsimile", "JS-M", "JS-H", "Articles of Faith",
		}},
	}

	for _, section := range sections {
		sb.WriteString(fmt.Sprintf("### %s\n", section.name))
		for _, book := range section.books {
			id, _ := citations.BookIDs[book]
			sb.WriteString(fmt.Sprintf("- %s (ID: %d)\n", book, id))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// Serve starts the MCP server on stdio.
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
