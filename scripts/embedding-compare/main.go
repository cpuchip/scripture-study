package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "embed":
		cmdEmbed(os.Args[2:])
	case "compare":
		cmdCompare(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: embedding-compare <command> [flags]

Commands:
  embed     Embed 1 Nephi content with the currently-loaded LM Studio model
  compare   Compare two embedding sets and generate a report

Examples:
  # Step 1: Load 4B in LM Studio, then:
  embedding-compare embed --tag=4b

  # Step 2: Load 8B in LM Studio, then:
  embedding-compare embed --tag=8b

  # Step 3: Compare (no model needed):
  embedding-compare compare --a=4b --b=8b`)
}

func cmdEmbed(args []string) {
	fs := flag.NewFlagSet("embed", flag.ExitOnError)
	tag := fs.String("tag", "", "Tag for this embedding run (e.g., '4b', '8b')")
	dbPath := fs.String("db", "", "Path to gospel.db (auto-detected if empty)")
	url := fs.String("url", "http://localhost:1234/v1", "LM Studio API base URL")
	dims := fs.Int("dims", 0, "Force specific dimension count (0 = native)")
	fs.Parse(args)

	if *tag == "" {
		fmt.Fprintln(os.Stderr, "error: --tag is required (e.g., --tag=4b)")
		os.Exit(1)
	}

	if *dbPath == "" {
		*dbPath = findGospelDB()
	}

	if err := runEmbed(*tag, *dbPath, *url, *dims); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func cmdCompare(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	tagA := fs.String("a", "", "First embedding tag (e.g., '4b')")
	tagB := fs.String("b", "", "Second embedding tag (e.g., '8b')")
	fs.Parse(args)

	if *tagA == "" || *tagB == "" {
		fmt.Fprintln(os.Stderr, "error: --a and --b are required (e.g., --a=4b --b=8b)")
		os.Exit(1)
	}

	if err := runCompare(*tagA, *tagB); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// findGospelDB looks for gospel.db in common locations.
func findGospelDB() string {
	paths := []string{
		"../gospel-engine/data/gospel.db",
		"../../scripts/gospel-engine/data/gospel.db",
		"scripts/gospel-engine/data/gospel.db",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	fmt.Fprintln(os.Stderr, "error: could not find gospel.db — use --db flag")
	os.Exit(1)
	return ""
}
