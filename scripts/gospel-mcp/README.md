# Gospel MCP Server

A Model Context Protocol (MCP) server that provides AI assistants with context-rich access to gospel content through SQLite with FTS5 full-text search.

See [03_gospel-mcp.md](../plans/03_gospel-mcp.md) for the full design document.

## Quick Start

```bash
# Build the server (requires CGO and FTS5 support)
go build -tags "fts5" -o gospel-mcp.exe ./cmd/gospel-mcp

# Index all content (first time - takes ~7 minutes for full library)
./gospel-mcp.exe index --root ../../

# Index incrementally (only new/changed files)
./gospel-mcp.exe index --incremental --root ../../

# Start the MCP server
./gospel-mcp.exe serve
```

## Build Requirements

- Go 1.21 or later
- CGO enabled (`go env CGO_ENABLED` should return `1`)
- C compiler (gcc/clang on Linux/Mac, mingw-w64 on Windows)

On Windows with MinGW:
```bash
# Ensure CGO is enabled
$env:CGO_ENABLED="1"

# Build with FTS5 support
go build -tags "fts5" -o gospel-mcp.exe ./cmd/gospel-mcp
```

## Commands

### `index`
Build or rebuild the SQLite database from markdown files.

```bash
# Full index (drops and rebuilds)
./gospel-mcp index

# Incremental (only new/changed files)
./gospel-mcp index --incremental

# Force full reindex
./gospel-mcp index --force

# Index specific content type
./gospel-mcp index --source scriptures
./gospel-mcp index --source conference
./gospel-mcp index --source manual

# Index specific path
./gospel-mcp index --path "gospel-library/eng/scriptures/bofm"
```

### `serve`
Start the MCP server (stdio transport).

```bash
./gospel-mcp serve
```

## MCP Tools

The server provides 3 consolidated tools:

1. **gospel_search** - Full-text search across all content
2. **gospel_get** - Retrieve specific content by reference or path  
3. **gospel_list** - Browse and discover available content

## VS Code Configuration

Add to your VS Code settings.json (or mcp.json for MCP-specific config):

```json
{
  "mcp": {
    "servers": {
      "gospel": {
        "command": "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/gospel-mcp/gospel-mcp.exe",
        "args": ["serve", "--db", "C:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/gospel-mcp/gospel.db"]
      }
    }
  }
}
```

Replace the paths with your actual installation location.

### Usage in Chat

Once configured, you can ask Claude:
- "Search for scriptures about faith"
- "Get Moses 3:5 with context"
- "List all books in the Book of Mormon"
- "Find conference talks about charity by President Nelson"

## Database Statistics

After full indexing:
- **41,995 scripture verses** across OT, NT, Book of Mormon, D&C, and Pearl of Great Price
- **1,584 chapters** with full markdown content
- **496+ conference talks** from General Conference
- **20,710 manual/magazine sections** from Come Follow Me, handbooks, etc.
- **1.5+ million cross-references** linking scriptures to footnotes and topical guide

## Project Structure

```
gospel-mcp/
├── cmd/gospel-mcp/
│   ├── main.go      # Entry point, CLI parsing
│   ├── index.go     # Index command implementation
│   └── serve.go     # Serve command (MCP server)
├── internal/
│   ├── db/
│   │   ├── db.go        # Database connection, initialization
│   │   ├── schema.sql   # Table definitions
│   │   └── metadata.go  # Index metadata operations
│   ├── indexer/
│   │   ├── indexer.go   # Main indexing orchestration
│   │   ├── scripture.go # Scripture file parser
│   │   ├── talk.go      # Conference talk parser
│   │   ├── manual.go    # Manual/magazine parser
│   │   └── walker.go    # File system walker
│   ├── mcp/
│   │   ├── server.go    # MCP protocol implementation
│   │   └── protocol.go  # MCP message types
│   ├── tools/
│   │   ├── search.go    # gospel_search implementation
│   │   ├── get.go       # gospel_get implementation
│   │   └── list.go      # gospel_list implementation
│   └── urlgen/
│       └── urlgen.go    # Source URL generation
├── go.mod
├── go.sum
└── README.md
```
