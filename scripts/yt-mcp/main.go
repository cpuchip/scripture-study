package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		cmdServe()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func cmdServe() {
	cfg := DefaultConfig()
	server := NewMCPServer(cfg)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`yt-mcp - YouTube Transcript Downloader & Gospel Evaluator

Usage:
  yt-mcp <command>

Commands:
  serve    Start MCP server (for VS Code/Claude integration)
  help     Show this help message`)
}
