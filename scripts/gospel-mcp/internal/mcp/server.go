// Package mcp implements the Model Context Protocol server.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/tools"
)

// Server handles MCP protocol communication.
type Server struct {
	db     *db.DB
	tools  *tools.Tools
	reader *bufio.Reader
	writer io.Writer
}

// NewServer creates a new MCP server.
func NewServer(database *db.DB) *Server {
	return &Server{
		db:     database,
		tools:  tools.New(database),
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// Run starts the MCP server on stdio.
func (s *Server) Run() error {
	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("reading input: %w", err)
		}

		// Parse the JSON-RPC request
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		// Handle the request
		s.handleRequest(&req)
	}
}

func (s *Server) handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleListTools(req)
	case "tools/call":
		s.handleCallTool(req)
	case "notifications/initialized":
		// Client acknowledgment, no response needed
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *Server) handleInitialize(req *Request) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: Capabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "gospel-mcp",
			Version: "0.1.0",
		},
	}
	s.sendResult(req.ID, result)
}

func (s *Server) handleListTools(req *Request) {
	toolList := []ToolDefinition{
		{
			Name:        "gospel_search",
			Description: "Full-text search across gospel content (scriptures, conference talks, manuals). Supports phrase search (\"exact phrase\"), boolean operators (AND, OR, NOT), prefix matching (word*), and field filters (speaker:nelson, book:mosiah).",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]JSONSchemaProperty{
					"query": {
						Type:        "string",
						Description: "Search query. Supports: exact phrases (\"natural man\"), boolean (faith OR hope), prefix (intellig*), field filters (speaker:nelson).",
					},
					"source": {
						Type:        "string",
						Description: "Filter by content type: scriptures, conference, manual, magazine, or all (default).",
						Enum:        []string{"scriptures", "conference", "manual", "magazine", "all"},
					},
					"path": {
						Type:        "string",
						Description: "Narrow to path: bofm, 2024/10, come-follow-me-*, etc.",
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum results to return (default: 20, max: 100).",
					},
					"context": {
						Type:        "integer",
						Description: "Lines/verses of context around match (default: 3).",
					},
					"include_content": {
						Type:        "boolean",
						Description: "Return full content, not just excerpts (default: false).",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "gospel_get",
			Description: "Retrieve specific gospel content by scripture reference (D&C 93:36, Moses 3:5, 1 Nephi 3:7) or file path. Returns full content with context and cross-references.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]JSONSchemaProperty{
					"reference": {
						Type:        "string",
						Description: "Scripture reference: D&C 93:36, Moses 3:5, 1 Nephi 3:7, TG Faith, BD Atonement.",
					},
					"path": {
						Type:        "string",
						Description: "File path: gospel-library/eng/general-conference/2025/04/57nelson.md",
					},
					"context": {
						Type:        "integer",
						Description: "Additional verses/paragraphs of context (default: 0).",
					},
					"include_chapter": {
						Type:        "boolean",
						Description: "Return entire chapter/document (default: false).",
					},
				},
			},
		},
		{
			Name:        "gospel_list",
			Description: "Browse and discover available gospel content. List scripture volumes, conference years, manuals, and their contents.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]JSONSchemaProperty{
					"source": {
						Type:        "string",
						Description: "Content type: scriptures, conference, manual, magazine, or all.",
						Enum:        []string{"scriptures", "conference", "manual", "magazine", "all"},
					},
					"path": {
						Type:        "string",
						Description: "Path to list: bofm, 2025/04, come-follow-me-*",
					},
					"depth": {
						Type:        "integer",
						Description: "How deep to recurse (default: 1).",
					},
				},
			},
		},
	}

	s.sendResult(req.ID, ListToolsResult{Tools: toolList})
}

func (s *Server) handleCallTool(req *Request) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	var result interface{}
	var err error

	switch params.Name {
	case "gospel_search":
		result, err = s.tools.Search(params.Arguments)
	case "gospel_get":
		result, err = s.tools.Get(params.Arguments)
	case "gospel_list":
		result, err = s.tools.List(params.Arguments)
	default:
		s.sendError(req.ID, -32602, "Unknown tool", params.Name)
		return
	}

	if err != nil {
		s.sendToolError(req.ID, err.Error())
		return
	}

	// Format result as MCP content
	content, _ := json.MarshalIndent(result, "", "  ")
	s.sendResult(req.ID, CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(content),
			},
		},
	})
}

func (s *Server) sendResult(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.send(resp)
}

func (s *Server) sendError(id interface{}, code int, message, data string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ResponseError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.send(resp)
}

func (s *Server) sendToolError(id interface{}, message string) {
	s.sendResult(id, CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Error: %s", message),
			},
		},
		IsError: true,
	})
}

func (s *Server) send(v interface{}) {
	data, _ := json.Marshal(v)
	fmt.Fprintf(s.writer, "%s\n", data)
}
