// Gospel MCP Server - provides AI assistants with context-rich access to gospel content.
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "index":
		if err := runIndex(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("gospel-mcp version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Gospel MCP Server - Context-rich gospel content search

Usage:
  gospel-mcp <command> [options]

Commands:
  index       Build or rebuild the SQLite database from markdown files
  serve       Start the MCP server (stdio transport)
  version     Print version information
  help        Show this help message

Index Options:
  --incremental    Only index new or modified files (default: full reindex)
  --force          Force full reindex even if database exists
  --source TYPE    Index only: scriptures, conference, manual, magazine
  --path PATH      Index only files matching path pattern
  --db PATH        Database path (default: ./gospel.db)
  --root PATH      Root directory for gospel-library (default: auto-detect)

Serve Options:
  --db PATH        Database path (default: ./gospel.db)

Examples:
  gospel-mcp index                           # Full index
  gospel-mcp index --incremental             # Incremental update
  gospel-mcp index --source scriptures       # Index scriptures only
  gospel-mcp index --path bofm               # Index Book of Mormon only
  gospel-mcp serve                           # Start MCP server

For VS Code MCP configuration, add to settings.json:
  {
    "mcp": {
      "servers": {
        "gospel": {
          "command": "path/to/gospel-mcp",
          "args": ["serve"]
        }
      }
    }
  }
`)
}
