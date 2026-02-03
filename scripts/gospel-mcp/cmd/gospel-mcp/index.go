package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/indexer"
)

func runIndex(args []string) error {
	fs := flag.NewFlagSet("index", flag.ExitOnError)

	var (
		incremental = fs.Bool("incremental", false, "Only index new or modified files")
		force       = fs.Bool("force", false, "Force full reindex")
		source      = fs.String("source", "", "Index only: scriptures, conference, manual, magazine")
		pathFilter  = fs.String("path", "", "Index only files matching path pattern")
		dbPath      = fs.String("db", "./gospel.db", "Database path")
		rootPath    = fs.String("root", "", "Root directory for gospel-library (auto-detect if empty)")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Auto-detect root path
	root := *rootPath
	if root == "" {
		var err error
		root, err = findGospelLibraryRoot()
		if err != nil {
			return fmt.Errorf("could not find gospel-library: %w (use --root to specify)", err)
		}
	}

	fmt.Printf("Gospel MCP Indexer\n")
	fmt.Printf("==================\n")
	fmt.Printf("Database:   %s\n", *dbPath)
	fmt.Printf("Root:       %s\n", root)
	if *source != "" {
		fmt.Printf("Source:     %s\n", *source)
	}
	if *pathFilter != "" {
		fmt.Printf("Path:       %s\n", *pathFilter)
	}
	fmt.Printf("Mode:       %s\n", indexMode(*incremental, *force))
	fmt.Println()

	// Open database
	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer database.Close()

	// Reset if force or full index
	if *force || (!*incremental && *source == "" && *pathFilter == "") {
		fmt.Println("Resetting database...")
		if err := database.Reset(); err != nil {
			return fmt.Errorf("resetting database: %w", err)
		}
	}

	// Create indexer
	idx := indexer.New(database, root)

	// Set options
	opts := indexer.Options{
		Incremental: *incremental,
		Source:      *source,
		PathFilter:  *pathFilter,
	}

	// Run indexing
	start := time.Now()
	result, err := idx.Index(opts)
	if err != nil {
		return fmt.Errorf("indexing: %w", err)
	}
	elapsed := time.Since(start)

	// Print results
	fmt.Println()
	fmt.Printf("Indexing complete in %v\n", elapsed.Round(time.Millisecond))
	fmt.Println()
	fmt.Printf("Results:\n")
	fmt.Printf("  Files processed:    %d\n", result.FilesProcessed)
	fmt.Printf("  Files skipped:      %d\n", result.FilesSkipped)
	fmt.Printf("  Scriptures indexed: %d verses\n", result.ScripturesIndexed)
	fmt.Printf("  Chapters indexed:   %d\n", result.ChaptersIndexed)
	fmt.Printf("  Talks indexed:      %d\n", result.TalksIndexed)
	fmt.Printf("  Manuals indexed:    %d\n", result.ManualsIndexed)
	fmt.Printf("  Books indexed:      %d\n", result.BooksIndexed)
	fmt.Printf("  Cross-refs found:   %d\n", result.CrossRefsIndexed)

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Printf("Errors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	// Show database stats
	stats, err := database.GetStats()
	if err == nil {
		fmt.Println()
		fmt.Printf("Database totals:\n")
		fmt.Printf("  Scriptures: %d\n", stats.Scriptures)
		fmt.Printf("  Chapters:   %d\n", stats.Chapters)
		fmt.Printf("  Talks:      %d\n", stats.Talks)
		fmt.Printf("  Manuals:    %d\n", stats.Manuals)
		fmt.Printf("  Books:      %d\n", stats.Books)
		fmt.Printf("  Cross-refs: %d\n", stats.CrossRefs)
	}

	return nil
}

func indexMode(incremental, force bool) string {
	if force {
		return "force (full reindex)"
	}
	if incremental {
		return "incremental"
	}
	return "full"
}

// findGospelLibraryRoot looks for the gospel-library directory starting from cwd.
func findGospelLibraryRoot() (string, error) {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up looking for gospel-library
	dir := cwd
	for {
		candidate := filepath.Join(dir, "gospel-library")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			// Return the parent of gospel-library (the repo root)
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return "", fmt.Errorf("gospel-library not found in %s or parent directories", cwd)
}
