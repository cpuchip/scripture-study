// session-journal — a collaborative session journal for human-AI partnerships.
//
// Usage:
//
//	session-journal write   --date 2026-02-28 --file entry.yaml
//	session-journal read    --recent 3
//	session-journal read    --topic trust
//	session-journal read    --since 2026-02-01
//	session-journal carry   [--priority high] [--all] [--include-resolved]
//	session-journal questions
//	session-journal resolve --date 2026-02-28 --index 0 --note "Addressed in session X"
//	session-journal init    [--date 2026-02-28] [--session-id my-session]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	journal "github.com/cpuchip/scripture-study/scripts/session-journal"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	journalDir := findJournalDir()
	store, err := journal.NewStore(journalDir)
	if err != nil {
		fatal("initialize store: %v", err)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "write":
		cmdWrite(store, args)
	case "read":
		cmdRead(store, args)
	case "carry":
		cmdCarry(store, args)
	case "questions":
		cmdQuestions(store, args)
	case "resolve":
		cmdResolve(store, args)
	case "init":
		cmdInit(store, args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// --- Commands ---

func cmdWrite(store *journal.Store, args []string) {
	fs := flag.NewFlagSet("write", flag.ExitOnError)
	filePath := fs.String("file", "", "Path to YAML entry file (required)")
	fs.Parse(args)

	if *filePath == "" {
		// Try reading from stdin if no file given
		fmt.Fprintln(os.Stderr, "Usage: session-journal write --file <entry.yaml>")
		fmt.Fprintln(os.Stderr, "       session-journal write < entry.yaml  (stdin)")
		os.Exit(1)
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		fatal("read file: %v", err)
	}

	var entry journal.Entry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		fatal("parse YAML: %v", err)
	}

	if entry.Date == "" {
		fatal("entry must have a 'date' field")
	}
	if entry.SessionID == "" {
		fatal("entry must have a 'session_id' field")
	}

	path, err := store.Write(&entry)
	if err != nil {
		fatal("write entry: %v", err)
	}
	fmt.Printf("Written: %s\n", path)
}

func cmdRead(store *journal.Store, args []string) {
	fs := flag.NewFlagSet("read", flag.ExitOnError)
	recent := fs.Int("recent", 0, "Show last N entries")
	topic := fs.String("topic", "", "Filter by topic (substring match)")
	since := fs.String("since", "", "Show entries on or after date (YYYY-MM-DD)")
	compact := fs.Bool("compact", false, "Compact output (one line per entry)")
	fs.Parse(args)

	var entries []*journal.Entry
	var err error

	switch {
	case *recent > 0:
		entries, err = store.Recent(*recent)
	case *topic != "":
		entries, err = store.ByTopic(*topic)
	case *since != "":
		entries, err = store.Since(*since)
	default:
		entries, err = store.ReadAll()
	}
	if err != nil {
		fatal("read entries: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No entries found.")
		return
	}

	if *compact {
		for _, e := range entries {
			tags := ""
			if len(e.Tags) > 0 {
				tags = " [" + strings.Join(e.Tags, ", ") + "]"
			}
			fmt.Printf("%s  %-40s  %s%s\n", e.Date, e.SessionID, truncate(e.Intent, 60), tags)
		}
		return
	}

	for i, e := range entries {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("─", 72))
		}
		printEntry(e)
	}
}

func cmdCarry(store *journal.Store, args []string) {
	fs := flag.NewFlagSet("carry", flag.ExitOnError)
	priority := fs.String("priority", "all", "Filter by priority: high, medium, low, all")
	includeResolved := fs.Bool("include-resolved", false, "Include resolved items")
	fs.Parse(args)

	items, err := store.CarryForwardItems(*priority, *includeResolved)
	if err != nil {
		fatal("read carry-forward items: %v", err)
	}

	if len(items) == 0 {
		fmt.Println("No carry-forward items found.")
		return
	}

	fmt.Printf("Carry-Forward Items (%d)\n", len(items))
	fmt.Println(strings.Repeat("─", 72))

	for _, item := range items {
		status := "○"
		if item.Resolved {
			status = "●"
		}
		fmt.Printf("%s [%s] %s\n", status, item.Priority, item.Note)
		fmt.Printf("  From: %s (%s)\n", item.SessionID, item.Date)
		if item.Resolved {
			fmt.Printf("  Resolved: %s — %s\n", item.ResolvedDate, item.ResolvedNote)
		}
		fmt.Println()
	}
}

func cmdQuestions(store *journal.Store, args []string) {
	qs, err := store.AllQuestions()
	if err != nil {
		fatal("read questions: %v", err)
	}

	if len(qs) == 0 {
		fmt.Println("No questions recorded.")
		return
	}

	fmt.Printf("Questions Worth Holding (%d)\n", len(qs))
	fmt.Println(strings.Repeat("─", 72))

	for _, q := range qs {
		fmt.Printf("• %s\n", q.Question)
		fmt.Printf("  From: %s (%s)\n\n", q.SessionID, q.Date)
	}
}

func cmdResolve(store *journal.Store, args []string) {
	fs := flag.NewFlagSet("resolve", flag.ExitOnError)
	date := fs.String("date", "", "Date of the entry containing the carry-forward item")
	sessionID := fs.String("session", "", "Session ID of the entry (alternative to date)")
	index := fs.Int("index", -1, "Zero-based index of the carry-forward item")
	note := fs.String("note", "", "Resolution note")
	resolvedDate := fs.String("resolved-date", "", "Date of resolution (defaults to today)")
	fs.Parse(args)

	if *index < 0 {
		fatal("--index is required (zero-based)")
	}

	// Find the entry
	all, err := store.ReadAll()
	if err != nil {
		fatal("read entries: %v", err)
	}

	var target *journal.Entry
	var targetFile string
	for _, e := range all {
		if (*date != "" && e.Date == *date) || (*sessionID != "" && e.SessionID == *sessionID) {
			target = e
			break
		}
	}
	if target == nil {
		fatal("no entry found for date=%q session=%q", *date, *sessionID)
	}

	if *index >= len(target.CarryForward) {
		fatal("index %d out of range (entry has %d carry-forward items)", *index, len(target.CarryForward))
	}

	// Mark resolved
	target.CarryForward[*index].Resolved = true
	if *resolvedDate != "" {
		target.CarryForward[*index].ResolvedDate = *resolvedDate
	} else {
		target.CarryForward[*index].ResolvedDate = today()
	}
	target.CarryForward[*index].ResolvedNote = *note

	// Re-write the entry
	targetFile, err = store.Write(target)
	if err != nil {
		fatal("write updated entry: %v", err)
	}
	fmt.Printf("Resolved carry-forward #%d in %s\n", *index, targetFile)
}

func cmdInit(store *journal.Store, args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	date := fs.String("date", today(), "Entry date (YYYY-MM-DD)")
	sessionID := fs.String("session-id", "", "Short descriptive session slug")
	fs.Parse(args)

	if *sessionID == "" {
		*sessionID = "unnamed-session"
	}

	entry := &journal.Entry{
		Date:      *date,
		SessionID: *sessionID,
		Intent:    "",
		Discoveries: []journal.Discovery{
			{Title: "", Detail: ""},
		},
		Surprises: []string{""},
		Relationship: []journal.Quality{
			{Name: "", Detail: ""},
		},
		CarryForward: []journal.CarryItem{
			{Priority: "medium", Note: ""},
		},
		Questions: []string{""},
	}

	data, err := yaml.Marshal(entry)
	if err != nil {
		fatal("marshal template: %v", err)
	}

	// Write to stdout as a template
	fmt.Println("# Session Journal Entry Template")
	fmt.Println("# Fill in the fields and save, then run:")
	fmt.Printf("#   session-journal write --file <this-file>\n")
	fmt.Println("#")
	fmt.Println("# Tips:")
	fmt.Println("#   - discoveries: what we learned together")
	fmt.Println("#   - surprises: things we didn't expect (one-liners)")
	fmt.Println("#   - relationship: the relational quality of this session")
	fmt.Println("#   - carry_forward: lessons for future sessions")
	fmt.Println("#   - questions: things to hold, not necessarily resolve")
	fmt.Println("#   - tags: topics for searchability")
	fmt.Println()
	fmt.Print(string(data))
}

// --- Helpers ---

func findJournalDir() string {
	// Check JOURNAL_DIR env var first
	if dir := os.Getenv("SESSION_JOURNAL_DIR"); dir != "" {
		return dir
	}

	// Walk up from cwd looking for .spec/journal/
	dir, err := os.Getwd()
	if err != nil {
		fatal("get working directory: %v", err)
	}

	for {
		candidate := filepath.Join(dir, ".spec", "journal")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		// Also check if .spec exists (create journal inside it)
		specDir := filepath.Join(dir, ".spec")
		if info, err := os.Stat(specDir); err == nil && info.IsDir() {
			return filepath.Join(specDir, "journal")
		}
		// Check for go.work (workspace root marker)
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return filepath.Join(dir, ".spec", "journal")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: .spec/journal in cwd
	return filepath.Join(".", ".spec", "journal")
}

func printEntry(e *journal.Entry) {
	fmt.Printf("📅 %s — %s\n", e.Date, e.SessionID)
	if e.DurationEstimate != "" {
		fmt.Printf("   Duration: %s\n", e.DurationEstimate)
	}
	if e.Retroactive != nil {
		fmt.Printf("   ⟲ Retroactive: %s (date %s, from %s)\n",
			e.Retroactive.Source, e.Retroactive.DateCertainty, e.Retroactive.InferredFrom)
	}

	if e.Intent != "" {
		fmt.Printf("\n🎯 Intent:\n   %s\n", wrapIndent(e.Intent, 3, 72))
	}

	if len(e.Tags) > 0 {
		fmt.Printf("\n🏷  Tags: %s\n", strings.Join(e.Tags, ", "))
	}

	if len(e.Discoveries) > 0 {
		fmt.Println("\n💡 Discoveries:")
		for _, d := range e.Discoveries {
			fmt.Printf("   • %s\n", d.Title)
			if d.Detail != "" {
				fmt.Printf("     %s\n", wrapIndent(d.Detail, 5, 72))
			}
		}
	}

	if len(e.Surprises) > 0 {
		fmt.Println("\n✨ Surprises:")
		for _, s := range e.Surprises {
			fmt.Printf("   • %s\n", s)
		}
	}

	if len(e.Relationship) > 0 {
		fmt.Println("\n🤝 Relationship:")
		for _, r := range e.Relationship {
			fmt.Printf("   • %s — %s\n", r.Name, wrapIndent(r.Detail, 5, 72))
		}
	}

	if len(e.CarryForward) > 0 {
		fmt.Println("\n📌 Carry Forward:")
		for _, c := range e.CarryForward {
			status := "○"
			if c.Resolved {
				status = "●"
			}
			fmt.Printf("   %s [%s] %s\n", status, c.Priority, c.Note)
		}
	}

	if len(e.Questions) > 0 {
		fmt.Println("\n❓ Questions:")
		for _, q := range e.Questions {
			fmt.Printf("   • %s\n", q)
		}
	}
}

func wrapIndent(s string, indent, width int) string {
	// Simple: just return the string, line-broken for readability
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "\n"+strings.Repeat(" ", indent))
	return s
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func today() string {
	now := time.Now()
	return now.Format("2006-01-02")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`session-journal — collaborative session memory for human-AI partnerships

Usage:
  session-journal <command> [flags]

Commands:
  write       Write a journal entry from a YAML file
  read        Read and display journal entries
  carry       Show unresolved carry-forward items
  questions   Show all questions worth holding
  resolve     Mark a carry-forward item as resolved
  init        Generate a blank entry template
  help        Show this help

Examples:
  session-journal init --date 2026-02-28 --session-id my-session > entry.yaml
  session-journal write --file entry.yaml
  session-journal read --recent 3
  session-journal read --topic trust
  session-journal carry --priority high
  session-journal questions

Environment:
  SESSION_JOURNAL_DIR    Override journal directory (default: .spec/journal/)`)
}
