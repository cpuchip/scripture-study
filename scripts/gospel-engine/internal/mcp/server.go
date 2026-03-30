// Package mcp implements the MCP JSON-RPC server for gospel-engine.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ToolDef describes an MCP tool.
type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

// ToolCallParams is the params for tools/call.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ContentItem is a text content block returned by tools.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Server implements the MCP protocol over stdin/stdout.
type Server struct {
	tools   []ToolDef
	handler func(name string, args json.RawMessage) (string, error)
}

// NewServer creates a new MCP server.
func NewServer(tools []ToolDef, handler func(name string, args json.RawMessage) (string, error)) *Server {
	return &Server{tools: tools, handler: handler}
}

// Run reads from stdin and writes to stdout.
func (s *Server) Run() error {
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

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(encoder, nil, -32700, "Parse error")
			continue
		}

		s.handleRequest(encoder, &req)
	}
}

func (s *Server) handleRequest(enc *json.Encoder, req *Request) {
	switch req.Method {
	case "initialize":
		s.sendResult(enc, req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "gospel-engine",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		})
	case "tools/list":
		s.sendResult(enc, req.ID, map[string]any{
			"tools": s.tools,
		})
	case "tools/call":
		s.handleToolCall(enc, req)
	case "notifications/initialized":
		// No response needed
	default:
		s.sendError(enc, req.ID, -32601, "Method not found: "+req.Method)
	}
}

func (s *Server) handleToolCall(enc *json.Encoder, req *Request) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(enc, req.ID, -32602, "Invalid params")
		return
	}

	text, err := s.handler(params.Name, params.Arguments)
	if err != nil {
		s.sendResult(enc, req.ID, map[string]any{
			"content": []ContentItem{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			"isError": true,
		})
		return
	}

	s.sendResult(enc, req.ID, map[string]any{
		"content": []ContentItem{{Type: "text", Text: text}},
	})
}

func (s *Server) sendResult(enc *json.Encoder, id any, result any) {
	enc.Encode(Response{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *Server) sendError(enc *json.Encoder, id any, code int, message string) {
	enc.Encode(Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: message}})
}
