package main

import (
	"flag"
	"fmt"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/mcp"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)

	var (
		dbPath = fs.String("db", "./gospel.db", "Database path")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Open database
	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer database.Close()

	// Check database has content
	stats, err := database.GetStats()
	if err != nil {
		return fmt.Errorf("checking database: %w", err)
	}

	if stats.Scriptures == 0 && stats.Talks == 0 && stats.Manuals == 0 {
		return fmt.Errorf("database is empty - run 'gospel-mcp index' first")
	}

	// Start MCP server
	server := mcp.NewServer(database)
	return server.Run()
}
