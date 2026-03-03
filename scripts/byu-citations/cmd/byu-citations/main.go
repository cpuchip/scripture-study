// byu-citations is an MCP server that provides access to the BYU Scripture
// Citation Index (https://scriptures.byu.edu/).
//
// The BYU Scripture Citation Index tracks which General Conference talks,
// Journal of Discourses entries, and other sources cite each verse of scripture.
// This tool makes that data accessible for scripture study sessions.
//
// Usage:
//
//	byu-citations                    # Start MCP server on stdio
//	byu-citations lookup "D&C 113:6" # CLI lookup mode
//	byu-citations lookup "3 Nephi 21:10"
//	byu-citations lookup "Isaiah 11:1,10"
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/byu-citations/internal/citations"
	mcpserver "github.com/cpuchip/scripture-study/scripts/byu-citations/internal/mcp"
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "lookup" {
		// CLI mode: byu-citations lookup "D&C 113:6"
		ref := strings.Join(os.Args[2:], " ")
		runLookup(ref)
		return
	}

	if len(os.Args) >= 2 && os.Args[1] == "lookup" {
		fmt.Fprintf(os.Stderr, "Usage: byu-citations lookup <reference>\n")
		fmt.Fprintf(os.Stderr, "Example: byu-citations lookup \"D&C 113:6\"\n")
		os.Exit(1)
	}

	// Default: MCP server mode
	server := mcpserver.New()
	log.Println("Starting byu-citations MCP server...")
	if err := server.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runLookup(ref string) {
	client := citations.NewClient()
	result, err := client.Lookup(ref)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(citations.FormatResult(result))
}
