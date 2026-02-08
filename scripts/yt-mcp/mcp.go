package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ── MCP JSON-RPC Types ───────────────────────────────────────────────────────

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

// ── MCP Server ───────────────────────────────────────────────────────────────

type MCPServer struct {
	cfg *Config
}

func NewMCPServer(cfg *Config) *MCPServer {
	return &MCPServer{cfg: cfg}
}

// Run starts the MCP server reading JSON-RPC from stdin, writing to stdout.
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

// ── Initialize ───────────────────────────────────────────────────────────────

func (s *MCPServer) handleInitialize(enc *json.Encoder, req *MCPRequest) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "yt-mcp",
			"version": "0.1.0",
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}
	s.sendResult(enc, req.ID, result)
}

// ── Tools List ───────────────────────────────────────────────────────────────

func (s *MCPServer) handleToolsList(enc *json.Encoder, req *MCPRequest) {
	tools := []map[string]any{
		{
			"name":        "yt_download",
			"description": "Download the English transcript and metadata from a YouTube video using yt-dlp. Saves to ./yt/{channel}/{video_id}/. Returns the transcript text and metadata. Requires yt-dlp to be installed and in PATH.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "YouTube video URL (e.g., 'https://www.youtube.com/watch?v=...' or 'https://youtu.be/...')",
					},
					"force": map[string]any{
						"type":        "boolean",
						"description": "Re-download even if transcript already exists locally. Default: false",
					},
					"cookies": map[string]any{
						"type":        "string",
						"description": "Path to a Netscape-format cookies.txt file for YouTube authentication. Use when YouTube requires sign-in (bot detection). Overrides the YT_COOKIE_FILE env var.",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			"name":        "yt_get",
			"description": "Get the full transcript and metadata of a previously downloaded YouTube video. Use after yt_download or yt_list to read the content.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"video_id": map[string]any{
						"type":        "string",
						"description": "YouTube video ID (e.g., 'dQw4w9WgXcQ')",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "Direct path to the transcript directory, if known",
					},
				},
			},
		},
		{
			"name":        "yt_list",
			"description": "List downloaded YouTube transcripts. Can filter by channel. Shows title, date, channel, and video ID for each.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"channel": map[string]any{
						"type":        "string",
						"description": "Filter by channel slug (e.g., 'book-of-mormon-central')",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum results to return (default: 20)",
					},
				},
			},
		},
		{
			"name":        "yt_search",
			"description": "Search across all downloaded YouTube transcripts for a keyword or phrase. Returns matching excerpts with video context and clickable timestamp links.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Text to search for in transcripts",
					},
					"channel": map[string]any{
						"type":        "string",
						"description": "Filter by channel slug",
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

// ── Tools Call Router ────────────────────────────────────────────────────────

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
	case "yt_download":
		s.handleYtDownload(enc, req, params.Arguments)
	case "yt_get":
		s.handleYtGet(enc, req, params.Arguments)
	case "yt_list":
		s.handleYtList(enc, req, params.Arguments)
	case "yt_search":
		s.handleYtSearch(enc, req, params.Arguments)
	default:
		s.sendError(enc, req.ID, -32602, "Unknown tool", params.Name)
	}
}

// ── yt_download ──────────────────────────────────────────────────────────────

func (s *MCPServer) handleYtDownload(enc *json.Encoder, req *MCPRequest, args json.RawMessage) {
	var input struct {
		URL     string `json:"url"`
		Force   bool   `json:"force"`
		Cookies string `json:"cookies"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		s.sendError(enc, req.ID, -32602, "Invalid arguments", err.Error())
		return
	}

	if input.URL == "" {
		s.sendError(enc, req.ID, -32602, "Missing required parameter", "url is required")
		return
	}

	result, err := DownloadVideo(s.cfg, input.URL, input.Force, input.Cookies)
	if err != nil {
		s.sendToolError(enc, req.ID, fmt.Sprintf("Download failed: %v", err))
		return
	}

	// Build response text
	response := fmt.Sprintf("**Downloaded:** %s\n**Channel:** %s\n**Date:** %s\n**Duration:** %s\n**Saved to:** %s\n\n---\n\n%s",
		result.Metadata.Title,
		result.Metadata.Channel,
		formatDate(result.Metadata.UploadDate),
		formatDuration(result.Metadata.Duration),
		result.OutputDir,
		result.Transcript,
	)

	s.sendToolResult(enc, req.ID, response)
}

// ── yt_get ───────────────────────────────────────────────────────────────────

func (s *MCPServer) handleYtGet(enc *json.Encoder, req *MCPRequest, args json.RawMessage) {
	var input struct {
		VideoID string `json:"video_id"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		s.sendError(enc, req.ID, -32602, "Invalid arguments", err.Error())
		return
	}

	var dir string
	var err error

	if input.Path != "" {
		dir = input.Path
	} else if input.VideoID != "" {
		dir, err = FindVideoDir(s.cfg.YTDir, input.VideoID)
		if err != nil {
			s.sendToolError(enc, req.ID, err.Error())
			return
		}
	} else {
		s.sendToolError(enc, req.ID, "Either video_id or path is required")
		return
	}

	meta, transcript, err := LoadVideoData(dir)
	if err != nil {
		s.sendToolError(enc, req.ID, fmt.Sprintf("Failed to load video: %v", err))
		return
	}

	response := fmt.Sprintf("**Title:** %s\n**Channel:** %s\n**Date:** %s\n**Duration:** %s\n**URL:** %s\n**Local path:** %s\n\n---\n\n%s",
		meta.Title,
		meta.Channel,
		formatDate(meta.UploadDate),
		formatDuration(meta.Duration),
		meta.URL,
		dir,
		transcript,
	)

	s.sendToolResult(enc, req.ID, response)
}

// ── yt_list ──────────────────────────────────────────────────────────────────

func (s *MCPServer) handleYtList(enc *json.Encoder, req *MCPRequest, args json.RawMessage) {
	var input struct {
		Channel string `json:"channel"`
		Limit   int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		s.sendError(enc, req.ID, -32602, "Invalid arguments", err.Error())
		return
	}

	videos, err := ListVideos(s.cfg.YTDir, input.Channel, input.Limit)
	if err != nil {
		s.sendToolError(enc, req.ID, fmt.Sprintf("Failed to list videos: %v", err))
		return
	}

	if len(videos) == 0 {
		s.sendToolResult(enc, req.ID, "No downloaded videos found. Use yt_download to download a video first.")
		return
	}

	var response string
	for i, v := range videos {
		response += fmt.Sprintf("%d. **%s**\n   Channel: %s | Date: %s | ID: `%s`\n   %s\n\n",
			i+1, v.Title, v.Channel, formatDate(v.UploadDate), v.ID, v.URL)
	}

	s.sendToolResult(enc, req.ID, response)
}

// ── yt_search ────────────────────────────────────────────────────────────────

func (s *MCPServer) handleYtSearch(enc *json.Encoder, req *MCPRequest, args json.RawMessage) {
	var input struct {
		Query   string `json:"query"`
		Channel string `json:"channel"`
		Limit   int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		s.sendError(enc, req.ID, -32602, "Invalid arguments", err.Error())
		return
	}

	if input.Query == "" {
		s.sendToolError(enc, req.ID, "query is required")
		return
	}

	hits, err := SearchTranscripts(s.cfg.YTDir, input.Query, input.Channel, input.Limit)
	if err != nil {
		s.sendToolError(enc, req.ID, fmt.Sprintf("Search failed: %v", err))
		return
	}

	if len(hits) == 0 {
		s.sendToolResult(enc, req.ID, fmt.Sprintf("No results found for \"%s\".", input.Query))
		return
	}

	var response string
	for i, h := range hits {
		response += fmt.Sprintf("### %d. %s\n**Channel:** %s | **Date:** %s\n",
			i+1, h.Title, h.Channel, h.Date)
		if h.Timestamp != "" {
			response += fmt.Sprintf("**Timestamp:** %s\n", h.Timestamp)
		}
		response += fmt.Sprintf("\n> %s\n\n", h.Excerpt)
	}

	s.sendToolResult(enc, req.ID, response)
}

// ── Response Helpers ─────────────────────────────────────────────────────────

func (s *MCPServer) sendResult(enc *json.Encoder, id any, result any) {
	enc.Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *MCPServer) sendError(enc *json.Encoder, id any, code int, message, data string) {
	enc.Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	})
}

// sendToolResult sends a successful tool call response with text content.
func (s *MCPServer) sendToolResult(enc *json.Encoder, id any, text string) {
	s.sendResult(enc, id, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
	})
}

// sendToolError sends a tool call response indicating a tool-level error.
func (s *MCPServer) sendToolError(enc *json.Encoder, id any, text string) {
	s.sendResult(enc, id, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
		"isError": true,
	})
}
