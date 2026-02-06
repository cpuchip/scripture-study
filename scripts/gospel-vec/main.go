package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "test":
		cmdTest()
	case "index":
		cmdIndex(os.Args[2:])
	case "index-talks":
		cmdIndexTalks(os.Args[2:])
	case "index-manuals":
		cmdIndexManuals(os.Args[2:])
	case "index-all":
		cmdIndexAll(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "mcp":
		cmdMCP(os.Args[2:])
	case "stats":
		cmdStats()
	case "config":
		cmdConfig()
	case "talks":
		cmdTalks(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gospel-vec - Scripture Vector Database

Usage:
  gospel-vec <command> [options]

Commands:
  test           Test LM Studio connection (embeddings and chat)
  index          Index scripture content into the vector database
  index-talks    Index conference talks into the vector database
  index-manuals  Index manuals and books into the vector database
  index-all      Index ALL content (scriptures + talks + manuals)
  search         Search the vector database (scriptures + talks + manuals)
  mcp            Start MCP server (for VS Code/Claude integration)
  stats          Show database statistics
  config         Show or initialize configuration
  talks          Parse and test conference talk indexing
  help           Show this help message

Index Options (scriptures):
  -volumes       Comma-separated volumes: bofm,dc-testament/dc,pgp,ot,nt (default: all)
  -layers        Comma-separated layers (default: verse,paragraph,summary,theme)
  -no-summary    Disable summary layer (skip LLM generation)
  -no-theme      Disable theme layer (skip LLM generation)
  -max           Max chapters to index (0 = all)
  -retries       Max retries on transient errors (default: 3)
  -continue      Continue indexing after persistent errors (default: true)
  -save-interval Save database every N chapters (default: 50)
  -v             Verbose output (default: true)

Index-talks Options:
  -years          Comma-separated years (empty = all 1971-2025)
  -layers         Comma-separated layers: paragraph,summary
  -max            Max talks to index (0 = all)
  -retries        Max retries on transient errors (default: 3)
  -continue       Continue indexing after persistent errors (default: true)
  -save-interval  Save database every N talks (default: 100)
  -v              Verbose output (default: true)

Index-manuals Options:
  -manuals        Comma-separated manual names (empty = all known manuals)
  -teachings      Index all Teachings of Presidents manuals
  -cfm            Index Come, Follow Me manuals
  -books          Index additional books (Lectures on Faith)
  -no-summary     Disable summary layer
  -retries        Max retries on transient errors (default: 3)
  -save-interval  Save database every N files (default: 50)
  -v              Verbose output (default: true)

Examples:
  gospel-vec test                       # Test LM Studio connection
  gospel-vec index                      # Index all scriptures (all layers)
  gospel-vec index -volumes ot,nt       # Index OT+NT with summaries+themes
  gospel-vec index -no-summary -no-theme # Just verse+paragraph (fast)
  gospel-vec index-talks                # Index all conference talks
  gospel-vec index-manuals              # Index all known manuals and books
  gospel-vec index-manuals -teachings   # Index only Teachings of Presidents
  gospel-vec index-manuals -books       # Index Lectures on Faith
  gospel-vec index-all                  # Index EVERYTHING
  gospel-vec search "faith"             # Search for faith (all sources)
  gospel-vec mcp -data ./data           # Start MCP server with data dir
  gospel-vec stats                      # Show database stats
`)
}

func cmdTest() {
	fmt.Println("🔍 Testing LM Studio connection...")

	cfg := DefaultConfig()
	ctx := context.Background()

	// Test embeddings
	fmt.Printf("\n📊 Testing embeddings at %s...\n", cfg.EmbeddingURL)
	if err := TestEmbedding(ctx, cfg.EmbeddingURL, cfg.EmbeddingModel); err != nil {
		fmt.Printf("❌ Embedding test failed: %v\n", err)
	} else {
		fmt.Println("✅ Embedding test passed")
	}

	// Test chat
	fmt.Printf("\n💬 Testing chat at %s...\n", cfg.ChatURL)
	if err := TestChat(ctx, cfg.ChatURL, cfg.ChatModel); err != nil {
		fmt.Printf("❌ Chat test failed: %v\n", err)
	} else {
		fmt.Println("✅ Chat test passed")
	}

	// List available models
	fmt.Println("\n📋 Available models:")
	models, err := GetAvailableModels(ctx, cfg.ChatURL)
	if err != nil {
		fmt.Printf("❌ Failed to list models: %v\n", err)
	} else {
		for _, m := range models {
			fmt.Printf("   - %s\n", m)
		}
	}
}

func cmdIndex(args []string) {
	flags := flag.NewFlagSet("index", flag.ExitOnError)

	allVolumes := "bofm,dc-testament/dc,pgp,ot,nt"
	volumes := flags.String("volumes", allVolumes, "Comma-separated volumes to index (bofm, dc-testament/dc, pgp, ot, nt)")
	layers := flags.String("layers", "verse,paragraph,summary,theme", "Comma-separated layers (verse, paragraph, summary, theme)")
	maxChapters := flags.Int("max", 0, "Max chapters to index (0 = all)")
	noSummary := flags.Bool("no-summary", false, "Disable summary layer")
	noTheme := flags.Bool("no-theme", false, "Disable theme layer")
	chatModel := flags.String("chat-model", "", "Chat model for summaries (e.g., qwen/qwen3-vl-8b)")
	noCache := flags.Bool("no-cache", false, "Don't use summary cache (regenerate all)")
	verbose := flags.Bool("v", true, "Verbose output")
	maxRetries := flags.Int("retries", 3, "Max retries on transient errors")
	continueOnError := flags.Bool("continue", true, "Continue indexing after persistent errors")
	saveInterval := flags.Int("save-interval", 50, "Save database every N chapters (0 = only at end)")
	noLock := flags.Bool("no-lock", false, "Skip lock acquisition (used internally by index-all)")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg := DefaultConfig()

	// Acquire index lock unless called from index-all
	if !*noLock {
		lock := NewIndexLock(cfg.DataDir)
		if err := lock.Acquire("index"); err != nil {
			fmt.Printf("🔒 %v\n", err)
			os.Exit(1)
		}
		defer lock.Release()
		fmt.Println("🔒 Index lock acquired")
	}

	// Set chat model if provided
	if *chatModel != "" {
		cfg.ChatModel = *chatModel
	}

	// Parse volumes
	volumeList := parseCSV(*volumes)

	// Parse layers
	layerList := []Layer{}
	for _, l := range parseCSV(*layers) {
		switch l {
		case "verse":
			layerList = append(layerList, LayerVerse)
		case "paragraph":
			layerList = append(layerList, LayerParagraph)
		case "summary":
			if !*noSummary {
				layerList = append(layerList, LayerSummary)
			}
		case "theme":
			if !*noTheme {
				layerList = append(layerList, LayerTheme)
			}
		}
	}

	// Auto-detect chat model if summary/theme layers requested but no model specified
	needsChat := containsLayer(layerList, LayerSummary) || containsLayer(layerList, LayerTheme)
	if needsChat && cfg.ChatModel == "" {
		fmt.Println("🔍 No chat model specified, auto-detecting from LM Studio...")
		models, err := GetAvailableModels(context.Background(), cfg.ChatURL)
		if err != nil {
			fmt.Printf("⚠️  Could not detect models: %v\n", err)
			fmt.Println("   Summary/theme layers will use cache only (no generation)")
		} else if len(models) > 0 {
			cfg.ChatModel = models[0]
			fmt.Printf("✅ Using chat model: %s\n", cfg.ChatModel)
		} else {
			fmt.Println("⚠️  No models available in LM Studio")
			fmt.Println("   Summary/theme layers will use cache only (no generation)")
		}
	}

	fmt.Printf("📚 Indexing volumes: %v\n", volumeList)
	fmt.Printf("📊 Layers: %v\n", layerList)

	// Create embedding function
	embedFunc := NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

	// Create store
	store, err := NewStore(cfg, embedFunc)
	if err != nil {
		fmt.Printf("❌ Failed to create store: %v\n", err)
		os.Exit(1)
	}

	// Create summarizer (optional)
	var summarizer *Summarizer
	if containsLayer(layerList, LayerSummary) || containsLayer(layerList, LayerTheme) {
		summarizer = NewSummarizer(cfg.ChatURL, cfg.ChatModel)
	}

	// Create indexer
	indexer := NewIndexer(store, summarizer, cfg)

	// Index
	ctx := context.Background()
	opts := IndexOptions{
		Layers:          layerList,
		Volumes:         volumeList,
		MaxChapters:     *maxChapters,
		Verbose:         *verbose,
		UseCache:        !*noCache,
		MaxRetries:      *maxRetries,
		ContinueOnError: *continueOnError,
		SaveInterval:    *saveInterval,
	}

	start := time.Now()
	if err := indexer.IndexScriptures(ctx, opts); err != nil {
		fmt.Printf("❌ Indexing failed: %v\n", err)
		os.Exit(1)
	}

	// Save database
	fmt.Println("\n💾 Saving database...")
	if err := store.Save(); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Indexing complete in %v\n", time.Since(start).Round(time.Second))

	// Show stats
	stats := store.Stats()
	fmt.Println("\n📈 Collection stats:")
	for name, count := range stats {
		fmt.Printf("   %s: %d documents\n", name, count)
	}
}

func cmdIndexAll(args []string) {
	flags := flag.NewFlagSet("index-all", flag.ExitOnError)
	retries := flags.Int("retries", 3, "Max retries on transient errors")
	noCache := flags.Bool("no-cache", false, "Don't use summary cache")
	verbose := flags.Bool("v", true, "Verbose output")
	noSummary := flags.Bool("no-summary", false, "Disable summary layer")
	noTheme := flags.Bool("no-theme", false, "Disable theme layer")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	// Acquire index lock for the entire run
	cfg := DefaultConfig()
	lock := NewIndexLock(cfg.DataDir)
	if err := lock.Acquire("index-all"); err != nil {
		fmt.Printf("🔒 %v\n", err)
		os.Exit(1)
	}
	defer lock.Release()
	fmt.Println("🔒 Index lock acquired")

	// Build args for sub-commands (with -no-lock since we hold the lock)
	scriptureArgs := []string{
		"-volumes", "bofm,dc-testament/dc,pgp,ot,nt",
		fmt.Sprintf("-retries=%d", *retries),
		fmt.Sprintf("-v=%t", *verbose),
		"-no-lock",
	}
	if *noCache {
		scriptureArgs = append(scriptureArgs, "-no-cache")
	}
	if *noSummary {
		scriptureArgs = append(scriptureArgs, "-no-summary")
	}
	if *noTheme {
		scriptureArgs = append(scriptureArgs, "-no-theme")
	}

	talkArgs := []string{
		fmt.Sprintf("-retries=%d", *retries),
		fmt.Sprintf("-v=%t", *verbose),
		"-no-lock",
	}
	if *noCache {
		talkArgs = append(talkArgs, "-no-cache")
	}

	fmt.Println("📚 === INDEXING ALL CONTENT ===")
	fmt.Println()
	fmt.Println("📖 Phase 1: Scriptures (all volumes, all layers)")
	fmt.Println("================================================")
	cmdIndex(scriptureArgs)

	fmt.Println()
	fmt.Println("🎤 Phase 2: Conference Talks (all years)")
	fmt.Println("================================================")
	cmdIndexTalks(talkArgs)

	fmt.Println()
	fmt.Println("📘 Phase 3: Manuals and Books")
	fmt.Println("================================================")
	manualArgs := []string{
		fmt.Sprintf("-retries=%d", *retries),
		fmt.Sprintf("-v=%t", *verbose),
		"-no-lock",
	}
	if *noSummary {
		manualArgs = append(manualArgs, "-no-summary")
	}
	cmdIndexManuals(manualArgs)

	fmt.Println()
	fmt.Println("🎉 === ALL INDEXING COMPLETE ===")
}

func cmdIndexManuals(args []string) {
	flags := flag.NewFlagSet("index-manuals", flag.ExitOnError)

	teachings := flags.Bool("teachings", false, "Index all Teachings of Presidents manuals")
	cfm := flags.Bool("cfm", false, "Index Come, Follow Me manuals")
	books := flags.Bool("books", false, "Index additional books (Lectures on Faith)")
	manualNames := flags.String("manuals", "", "Comma-separated manual names to index (empty with no flags = all)")
	noSummary := flags.Bool("no-summary", false, "Disable summary layer")
	chatModel := flags.String("chat-model", "", "Chat model for summaries")
	noCache := flags.Bool("no-cache", false, "Don't use summary cache")
	verbose := flags.Bool("v", true, "Verbose output")
	maxRetries := flags.Int("retries", 3, "Max retries on transient errors")
	continueOnError := flags.Bool("continue", true, "Continue indexing after persistent errors")
	saveInterval := flags.Int("save-interval", 50, "Save database every N files")
	noLock := flags.Bool("no-lock", false, "Skip lock acquisition (used internally by index-all)")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg := DefaultConfig()

	// Acquire index lock unless called from index-all
	if !*noLock {
		lock := NewIndexLock(cfg.DataDir)
		if err := lock.Acquire("index-manuals"); err != nil {
			fmt.Printf("🔒 %v\n", err)
			os.Exit(1)
		}
		defer lock.Release()
		fmt.Println("🔒 Index lock acquired")
	}

	if *chatModel != "" {
		cfg.ChatModel = *chatModel
	}

	// Determine manual base path
	manualBasePath := "../../gospel-library/eng/manual"
	booksBasePath := "../../books"
	if _, err := os.Stat("gospel-library"); err == nil {
		manualBasePath = "gospel-library/eng/manual"
		booksBasePath = "books"
	}

	// Build the list of manuals to index
	var manuals []ManualDefinition

	// If specific flags are set, use them
	flagsSpecified := *teachings || *cfm || *books || *manualNames != ""

	if *teachings || (!flagsSpecified) {
		for _, m := range KnownManuals() {
			if m.Type == "teachings" {
				m.Path = filepath.Join(manualBasePath, m.Path)
				manuals = append(manuals, m)
			}
		}
	}

	if *cfm || (!flagsSpecified) {
		for _, m := range KnownManuals() {
			if m.Type == "cfm" {
				m.Path = filepath.Join(manualBasePath, m.Path)
				manuals = append(manuals, m)
			}
		}
	}

	if !flagsSpecified {
		// Also include "manual" type manuals (e.g., Teaching in the Savior's Way)
		for _, m := range KnownManuals() {
			if m.Type == "manual" {
				m.Path = filepath.Join(manualBasePath, m.Path)
				manuals = append(manuals, m)
			}
		}
	}

	if *books || (!flagsSpecified) {
		for _, m := range KnownBooks() {
			m.Path = filepath.Join(booksBasePath, m.Path)
			manuals = append(manuals, m)
		}
	}

	// If specific manual names are provided, filter
	if *manualNames != "" {
		nameList := parseCSV(*manualNames)
		var filtered []ManualDefinition
		allManuals := append(KnownManuals(), KnownBooks()...)
		for _, name := range nameList {
			name = strings.TrimSpace(name)
			for _, m := range allManuals {
				if strings.Contains(strings.ToLower(m.Name), strings.ToLower(name)) ||
					strings.Contains(strings.ToLower(m.Path), strings.ToLower(name)) {
					// Determine correct base path
					if m.Type == "book" {
						m.Path = filepath.Join(booksBasePath, m.Path)
					} else {
						m.Path = filepath.Join(manualBasePath, m.Path)
					}
					filtered = append(filtered, m)
					break
				}
			}
		}
		manuals = filtered
	}

	if len(manuals) == 0 {
		fmt.Println("❌ No manuals matched. Use -teachings, -cfm, -books, or -manuals to specify.")
		os.Exit(1)
	}

	// Build layers
	layerList := []Layer{LayerParagraph}
	if !*noSummary {
		layerList = append(layerList, LayerSummary)
	}

	// Auto-detect chat model if summary layer requested
	needsChat := containsLayer(layerList, LayerSummary)
	if needsChat && cfg.ChatModel == "" {
		fmt.Println("🔍 No chat model specified, auto-detecting from LM Studio...")
		models, err := GetAvailableModels(context.Background(), cfg.ChatURL)
		if err != nil {
			fmt.Printf("⚠️  Could not detect models: %v\n", err)
			fmt.Println("   Summary layer will use cache only (no generation)")
		} else if len(models) > 0 {
			cfg.ChatModel = models[0]
			fmt.Printf("✅ Using chat model: %s\n", cfg.ChatModel)
		}
	}

	fmt.Printf("📘 Indexing %d manuals/books\n", len(manuals))
	for _, m := range manuals {
		fmt.Printf("   - %s\n", m.Name)
	}
	fmt.Printf("📊 Layers: %v\n", layerList)

	// Create embedding function
	embedFunc := NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

	// Create store
	store, err := NewStore(cfg, embedFunc)
	if err != nil {
		fmt.Printf("❌ Failed to create store: %v\n", err)
		os.Exit(1)
	}

	// Create summarizer (optional)
	var summarizer *Summarizer
	if containsLayer(layerList, LayerSummary) {
		summarizer = NewSummarizer(cfg.ChatURL, cfg.ChatModel)
	}

	// Create indexer
	indexer := NewIndexer(store, summarizer, cfg)

	// Index
	ctx := context.Background()
	opts := ManualIndexOptions{
		Layers:          layerList,
		Manuals:         manuals,
		Verbose:         *verbose,
		UseCache:        !*noCache,
		MaxRetries:      *maxRetries,
		ContinueOnError: *continueOnError,
		SaveInterval:    *saveInterval,
	}

	start := time.Now()
	if err := indexer.IndexManuals(ctx, opts); err != nil {
		fmt.Printf("❌ Indexing failed: %v\n", err)
		os.Exit(1)
	}

	// Save database
	fmt.Println("\n💾 Saving database...")
	if err := store.Save(); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Manual indexing complete in %v\n", time.Since(start).Round(time.Second))

	// Show stats
	stats := store.Stats()
	fmt.Println("\n📈 Collection stats:")
	for name, count := range stats {
		fmt.Printf("   %s: %d documents\n", name, count)
	}
}

func cmdIndexTalks(args []string) {
	flags := flag.NewFlagSet("index-talks", flag.ExitOnError)

	years := flags.String("years", "", "Comma-separated years to index (empty = all from 1971-2025)")
	layers := flags.String("layers", "paragraph,summary", "Comma-separated layers (paragraph, summary)")
	maxTalks := flags.Int("max", 0, "Max talks to index (0 = all)")
	withSummary := flags.Bool("summary", false, "Generate LLM summaries (slower)")
	chatModel := flags.String("chat-model", "", "Chat model for summaries (e.g., qwen/qwen3-vl-8b)")
	noCache := flags.Bool("no-cache", false, "Don't use summary cache (regenerate all)")
	verbose := flags.Bool("v", true, "Verbose output")
	maxRetries := flags.Int("retries", 3, "Max retries on transient errors")
	continueOnError := flags.Bool("continue", true, "Continue indexing after persistent errors")
	saveInterval := flags.Int("save-interval", 100, "Save database every N talks (0 = only at end)")
	noLock := flags.Bool("no-lock", false, "Skip lock acquisition (used internally by index-all)")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg := DefaultConfig()

	// Acquire index lock unless called from index-all
	if !*noLock {
		lock := NewIndexLock(cfg.DataDir)
		if err := lock.Acquire("index-talks"); err != nil {
			fmt.Printf("🔒 %v\n", err)
			os.Exit(1)
		}
		defer lock.Release()
		fmt.Println("🔒 Index lock acquired")
	}

	// Set chat model if provided
	if *chatModel != "" {
		cfg.ChatModel = *chatModel
	}

	// Parse years
	var yearList []string
	if *years != "" {
		yearList = parseCSV(*years)
	}

	// Parse layers
	layerList := []Layer{}
	for _, l := range parseCSV(*layers) {
		switch l {
		case "paragraph":
			layerList = append(layerList, LayerParagraph)
		case "summary":
			layerList = append(layerList, LayerSummary)
		}
	}

	// Add summary layer if requested
	if *withSummary && !containsLayer(layerList, LayerSummary) {
		layerList = append(layerList, LayerSummary)
	}

	// Auto-detect chat model if summary layer requested but no model specified
	needsChat := containsLayer(layerList, LayerSummary)
	if needsChat && cfg.ChatModel == "" {
		fmt.Println("🔍 No chat model specified, auto-detecting from LM Studio...")
		models, err := GetAvailableModels(context.Background(), cfg.ChatURL)
		if err != nil {
			fmt.Printf("⚠️  Could not detect models: %v\n", err)
			fmt.Println("   Summary layer will use cache only (no generation)")
		} else if len(models) > 0 {
			cfg.ChatModel = models[0]
			fmt.Printf("✅ Using chat model: %s\n", cfg.ChatModel)
		} else {
			fmt.Println("⚠️  No models available in LM Studio")
			fmt.Println("   Summary layer will use cache only (no generation)")
		}
	}

	// Use conference path from config
	conferencePath := cfg.ConferencePath

	if len(yearList) == 0 {
		fmt.Println("📚 Indexing ALL conference talks (1971-2025)")
	} else {
		fmt.Printf("📚 Indexing conference talks for years: %v\n", yearList)
	}
	fmt.Printf("📊 Layers: %v\n", layerList)

	// Create embedding function
	embedFunc := NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

	// Create store
	store, err := NewStore(cfg, embedFunc)
	if err != nil {
		fmt.Printf("❌ Failed to create store: %v\n", err)
		os.Exit(1)
	}

	// Create summarizer (optional)
	var summarizer *Summarizer
	if containsLayer(layerList, LayerSummary) {
		summarizer = NewSummarizer(cfg.ChatURL, cfg.ChatModel)
	}

	// Create indexer
	indexer := NewIndexer(store, summarizer, cfg)

	// Index
	ctx := context.Background()
	opts := TalkIndexOptions{
		Layers:          layerList,
		Years:           yearList,
		MaxTalks:        *maxTalks,
		Verbose:         *verbose,
		UseCache:        !*noCache,
		MaxRetries:      *maxRetries,
		ContinueOnError: *continueOnError,
		SaveInterval:    *saveInterval,
	}

	start := time.Now()
	if err := indexer.IndexConferenceTalks(ctx, conferencePath, opts); err != nil {
		fmt.Printf("❌ Indexing failed: %v\n", err)
		os.Exit(1)
	}

	// Save database
	fmt.Println("\n💾 Saving database...")
	if err := store.Save(); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Indexing complete in %v\n", time.Since(start).Round(time.Second))

	// Show stats
	stats := store.Stats()
	fmt.Println("\n📈 Collection stats:")
	for name, count := range stats {
		fmt.Printf("   %s: %d documents\n", name, count)
	}
}

func cmdSearch(args []string) {
	flags := flag.NewFlagSet("search", flag.ExitOnError)

	layers := flags.String("layers", "verse,paragraph", "Layers to search (verse, paragraph, summary, theme)")
	limit := flags.Int("limit", 10, "Max results per layer")
	showContent := flags.Bool("content", true, "Show result content")
	maxLen := flags.Int("maxlen", 200, "Max content length to show")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	if flags.NArg() < 1 {
		fmt.Println("Usage: gospel-vec search [options] <query>")
		os.Exit(1)
	}

	query := flags.Arg(0)
	cfg := DefaultConfig()

	// Parse layers (for search)
	layerList := []Layer{}
	for _, l := range parseCSV(*layers) {
		switch l {
		case "verse":
			layerList = append(layerList, LayerVerse)
		case "paragraph":
			layerList = append(layerList, LayerParagraph)
		case "summary":
			layerList = append(layerList, LayerSummary)
		case "theme":
			layerList = append(layerList, LayerTheme)
		}
	}

	// Create embedding function
	embedFunc := NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

	// Load store
	store, err := NewStore(cfg, embedFunc)
	if err != nil {
		fmt.Printf("❌ Failed to load store: %v\n", err)
		os.Exit(1)
	}

	// Search
	searcher := NewSearcher(store)
	ctx := context.Background()

	fmt.Printf("🔍 Searching for: %q\n\n", query)

	results, err := searcher.Search(ctx, query, SearchOptions{
		Layers: layerList,
		Limit:  *limit,
	})
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	fmt.Print(FormatResults(results, *showContent, *maxLen))
}

func cmdStats() {
	cfg := DefaultConfig()
	embedFunc := NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

	store, err := NewStore(cfg, embedFunc)
	if err != nil {
		fmt.Printf("❌ Failed to load store: %v\n", err)
		os.Exit(1)
	}

	stats := store.Stats()
	if len(stats) == 0 {
		fmt.Println("📊 Database is empty. Run 'gospel-vec index' to add content.")
		return
	}

	fmt.Println("📊 Database Statistics:\n")

	var total int
	for name, count := range stats {
		fmt.Printf("   %-25s %d documents\n", name, count)
		total += count
	}
	fmt.Printf("\n   %-25s %d documents\n", "TOTAL", total)

	// Show storage file info
	dbPath := cfg.DBPath()
	if info, err := os.Stat(dbPath); err == nil {
		fmt.Printf("\n💾 Storage: %s (%.2f MB)\n", dbPath, float64(info.Size())/1024/1024)
	}
}

func cmdConfig() {
	cfg := DefaultConfig()

	fmt.Println("📋 Current Configuration:\n")
	fmt.Printf("   Data Directory:    %s\n", cfg.DataDir)
	fmt.Printf("   Database File:     %s\n", cfg.DBFile)
	fmt.Printf("   Scriptures Path:   %s\n", cfg.ScripturesPath)
	fmt.Printf("   Conference Path:   %s\n", cfg.ConferencePath)
	fmt.Printf("\n   Embedding URL:     %s\n", cfg.EmbeddingURL)
	fmt.Printf("   Embedding Model:   %s\n", cfg.EmbeddingModel)
	fmt.Printf("   Chat URL:          %s\n", cfg.ChatURL)
	fmt.Printf("   Chat Model:        %s\n", cfg.ChatModel)
}

func cmdMCP(args []string) {
	flags := flag.NewFlagSet("mcp", flag.ExitOnError)
	dataDir := flags.String("data", "", "Path to data directory (default: ./data)")
	flags.Parse(args)

	cfg := DefaultConfig()

	// If data dir specified, update config
	if *dataDir != "" {
		cfg.DataDir = *dataDir
	}

	server, err := NewMCPServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start MCP server: %v\n", err)
		os.Exit(1)
	}

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

// Helper functions

func parseCSV(s string) []string {
	var result []string
	for _, part := range splitAndTrim(s, ",") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, p := range splitString(s, sep) {
		p = trimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if string(c) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

func containsLayer(layers []Layer, target Layer) bool {
	for _, l := range layers {
		if l == target {
			return true
		}
	}
	return false
}

func cmdTalks(args []string) {
	flags := flag.NewFlagSet("talks", flag.ExitOnError)

	sample := flags.Bool("sample", false, "Parse sample talks from each decade (1970s-2020s)")
	parse := flags.String("parse", "", "Parse a specific talk file")
	summarize := flags.String("summarize", "", "Test summary generation for talks from a year")
	listYears := flags.Bool("list", false, "List available conference years")
	verbose := flags.Bool("v", false, "Verbose output")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg := DefaultConfig()

	// Use conference path from config
	conferencePath := cfg.ConferencePath

	if *listYears {
		// List available years
		fmt.Println("📅 Available conference years:")
		entries, err := os.ReadDir(conferencePath)
		if err != nil {
			fmt.Printf("❌ Failed to read conference directory: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				fmt.Printf("   %s\n", entry.Name())
			}
		}
		return
	}

	if *parse != "" {
		// Parse a specific file
		talk, err := ParseTalkFile(*parse)
		if err != nil {
			fmt.Printf("❌ Failed to parse: %v\n", err)
			os.Exit(1)
		}
		printTalkMetadata(talk, *verbose)
		return
	}

	if *sample {
		// Sample talks from each decade
		sampleYears := []string{"1971", "1985", "1995", "2005", "2015", "2025"}
		fmt.Println("📚 Parsing sample talks from each decade...\n")

		for _, year := range sampleYears {
			files, err := FindTalkFiles(conferencePath, year)
			if err != nil || len(files) == 0 {
				fmt.Printf("⚠️  No talks found for %s\n", year)
				continue
			}

			// Pick a talk from April conference (avoid statistical reports)
			var selectedFile string
			for _, f := range files {
				if strings.Contains(f, "/04/") || strings.Contains(f, "\\04\\") {
					name := filepath.Base(f)
					if !strings.Contains(name, "statistical") && !strings.Contains(name, "audit") {
						selectedFile = f
						break
					}
				}
			}

			if selectedFile == "" && len(files) > 0 {
				selectedFile = files[0]
			}

			talk, err := ParseTalkFile(selectedFile)
			if err != nil {
				fmt.Printf("❌ %s: Failed to parse: %v\n", year, err)
				continue
			}

			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("📅 %s (%s)\n", year, filepath.Base(selectedFile))
			printTalkMetadata(talk, *verbose)
			fmt.Println()
		}
		return
	}

	if *summarize != "" {
		// Test summary generation
		files, err := FindTalkFiles(conferencePath, *summarize)
		if err != nil {
			fmt.Printf("❌ Failed to find talks: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("⚠️  No talks found for %s\n", *summarize)
			return
		}

		// Auto-detect chat model
		models, err := GetAvailableModels(context.Background(), cfg.ChatURL)
		if err != nil || len(models) == 0 {
			fmt.Printf("❌ No LM Studio models available: %v\n", err)
			os.Exit(1)
		}
		cfg.ChatModel = models[0]
		fmt.Printf("✅ Using chat model: %s\n\n", cfg.ChatModel)

		summarizer := NewSummarizer(cfg.ChatURL, cfg.ChatModel)
		ctx := context.Background()

		// Test on first 2 talks from April
		count := 0
		for _, f := range files {
			if count >= 2 {
				break
			}
			if !strings.Contains(f, "/04/") && !strings.Contains(f, "\\04\\") {
				continue
			}
			name := filepath.Base(f)
			// Skip known administrative documents
			if strings.Contains(name, "statistical") || strings.Contains(name, "audit") ||
				strings.Contains(name, "sustaining") {
				continue
			}

			talk, err := ParseTalkFile(f)
			if err != nil {
				fmt.Printf("⚠️  Skipping %s: %v\n", name, err)
				continue
			}

			// Skip talks without a speaker (typically administrative)
			if talk.Metadata.Speaker == "" {
				continue
			}

			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("📋 %s\n", talk.Metadata.Title)
			fmt.Printf("   By %s\n", talk.Metadata.Speaker)
			fmt.Printf("   %d paragraphs, %d sections\n\n", len(talk.Paragraphs), len(talk.Sections))

			// Generate summary using talk summarizer
			summary, err := summarizeTalk(ctx, summarizer, talk)
			if err != nil {
				fmt.Printf("❌ Summary failed: %v\n", err)
			} else {
				fmt.Printf("📝 Generated Summary:\n")
				fmt.Printf("   Keywords: %s\n", strings.Join(summary.Keywords, ", "))
				fmt.Printf("   Summary: %s\n", summary.Summary)
				fmt.Printf("   Key Quote: %s\n", summary.KeyVerse)
			}
			fmt.Println()
			count++
		}
		return
	}

	// Default: show usage
	fmt.Println("Usage: gospel-vec talks [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -sample      Parse sample talks from each decade")
	fmt.Println("  -parse FILE  Parse a specific talk file")
	fmt.Println("  -summarize YEAR  Test summary generation for talks")
	fmt.Println("  -list        List available conference years")
	fmt.Println("  -v           Verbose output")
}

func printTalkMetadata(talk *ParsedTalk, verbose bool) {
	fmt.Printf("   Title: %s\n", talk.Metadata.Title)
	fmt.Printf("   Speaker: %s\n", talk.Metadata.Speaker)
	fmt.Printf("   Position: %s\n", talk.Metadata.Position)
	fmt.Printf("   Conference: %s %s\n", talk.Metadata.Month, talk.Metadata.Year)
	if talk.Metadata.Session != "" {
		fmt.Printf("   Session: %s\n", talk.Metadata.Session)
	}
	fmt.Printf("   Paragraphs: %d\n", len(talk.Paragraphs))
	fmt.Printf("   Sections: %d\n", len(talk.Sections))
	if talk.Metadata.Summary != "" {
		fmt.Printf("   Opening: %s\n", truncateString(talk.Metadata.Summary, 80))
	}

	if verbose {
		if len(talk.Sections) > 0 {
			fmt.Println("   Section Headings:")
			for _, s := range talk.Sections {
				fmt.Printf("     - %s\n", s.Heading)
			}
		}

		refs := ExtractScriptureReferences(talk.RawContent)
		if len(refs) > 0 {
			fmt.Printf("   Scripture Refs: %d\n", len(refs))
			if len(refs) <= 10 {
				for _, ref := range refs {
					fmt.Printf("     - %s\n", ref)
				}
			}
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// summarizeTalk generates a summary for a conference talk
func summarizeTalk(ctx context.Context, summarizer *Summarizer, talk *ParsedTalk) (*ChapterSummary, error) {
	// Build content from paragraphs
	var content strings.Builder
	for i, para := range talk.Paragraphs {
		if i > 20 { // Limit content size
			content.WriteString("\n[Additional content truncated for summary]")
			break
		}
		content.WriteString(para)
		content.WriteString("\n\n")
	}

	// Use custom prompt for talks
	systemPrompt := `Create a summary of this conference talk optimized for semantic search indexing.

Format your response EXACTLY like this:
KEYWORDS: [10-15 comma-separated searchable terms including speaker themes, doctrines, people, events]
SUMMARY: [50-75 word narrative covering main message and teachings, present tense]
KEY_QUOTE: [Most memorable or powerful quote from the talk]

Keep output under 200 words total. No other text.`

	userPrompt := fmt.Sprintf(`Summarize this %s %s General Conference talk by %s titled "%s":

%s`, talk.Metadata.Month, talk.Metadata.Year, talk.Metadata.Speaker, talk.Metadata.Title, content.String())

	response, err := summarizer.chat(ctx, systemPrompt, userPrompt, 300)
	if err != nil {
		return nil, err
	}

	// Parse response (reuse ChapterSummary format)
	summary := &ChapterSummary{Raw: response}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "KEYWORDS:") {
			kwStr := strings.TrimPrefix(line, "KEYWORDS:")
			kwStr = strings.TrimSpace(kwStr)
			keywords := strings.Split(kwStr, ",")
			for _, kw := range keywords {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					summary.Keywords = append(summary.Keywords, kw)
				}
			}
		} else if strings.HasPrefix(line, "SUMMARY:") {
			summary.Summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		} else if strings.HasPrefix(line, "KEY_QUOTE:") || strings.HasPrefix(line, "KEY_VERSE:") {
			summary.KeyVerse = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "KEY_QUOTE:"), "KEY_VERSE:"))
		}
	}

	// Deduplicate keywords
	summary.Keywords = deduplicateKeywords(summary.Keywords)

	return summary, nil
}
