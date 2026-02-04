# DuckDuckGo Search MCP Server

A Go-based MCP (Model Context Protocol) server that provides web search capabilities using DuckDuckGo. **No API key required!**

## Features

- **Web Search** - General web search with customizable results
- **News Search** - Search recent news with time filters
- **Instant Answers** - Quick factual answers for simple queries

## Building

```bash
cd scripts/search-mcp
go build -o search-mcp.exe ./cmd/search-mcp
```

Or use `go install`:

```bash
go install ./cmd/search-mcp
```

## Usage

### As MCP Server (default)

The server runs in MCP mode by default, communicating via stdin/stdout:

```bash
./search-mcp.exe
# or
./search-mcp.exe serve
```

### Test Mode

Test the search functionality directly:

```bash
./search-mcp.exe test "your search query"
```

### Version

```bash
./search-mcp.exe version
```

## VS Code Configuration

Add to your `.vscode/mcp.json`:

```json
{
  "servers": {
    "search": {
      "type": "stdio",
      "command": "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/search-mcp/search-mcp.exe",
      "args": []
    }
  }
}
```

## Available Tools

### `web_search`

Search the web for any query.

**Parameters:**
- `query` (required): The search query
- `max_results` (optional): Number of results (default: 10, max: 25)
- `region` (optional): Region code (e.g., 'us-en', 'uk-en', 'wt-wt' for worldwide)

**Example:**
```
Search for: "LDS General Conference April 2025"
```

### `news_search`

Search for recent news articles.

**Parameters:**
- `query` (required): The news search query
- `max_results` (optional): Number of results (default: 10, max: 25)
- `timelimit` (optional): Time filter - 'd' (day), 'w' (week), 'm' (month)

**Example:**
```
Search news for: "Church of Jesus Christ temples"
```

### `instant_answer`

Get a direct answer for factual queries. Best for definitions, calculations, and simple questions.

**Parameters:**
- `query` (required): The question or factual query

**Example:**
```
Query: "What is the speed of light?"
```

## Architecture

```
search-mcp/
├── cmd/
│   └── search-mcp/
│       ├── main.go    # CLI entry point
│       ├── serve.go   # MCP server runner
│       └── test.go    # Test command
├── internal/
│   ├── ddg/
│   │   └── client.go  # DuckDuckGo search client
│   └── mcp/
│       └── server.go  # MCP protocol handler
├── go.mod
├── go.sum
└── README.md
```

## How It Works

This server scrapes DuckDuckGo's HTML search results (similar to the [langchaingo](https://github.com/tmc/langchaingo) implementation). It uses:

- **Web/News Search**: Parses `https://html.duckduckgo.com/html/?q=...`
- **Instant Answer**: Uses DuckDuckGo's public API `https://api.duckduckgo.com/?q=...&format=json`

## Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery) - HTML parsing
- [golang.org/x/net](https://pkg.go.dev/golang.org/x/net) - HTTP utilities

## Notes

- No API key required - DuckDuckGo's HTML endpoint is publicly accessible
- Be respectful of rate limits
- Results quality depends on DuckDuckGo's HTML structure (which may change)
