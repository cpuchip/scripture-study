package main

import (
	"context"
	"flag"
	"fmt"
	"os"
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
	case "search":
		cmdSearch(os.Args[2:])
	case "stats":
		cmdStats()
	case "config":
		cmdConfig()
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
  test     Test LM Studio connection (embeddings and chat)
  index    Index scripture content into the vector database
  search   Search the vector database
  stats    Show database statistics
  config   Show or initialize configuration
  help     Show this help message

Examples:
  gospel-vec test                    # Test LM Studio connection
  gospel-vec index -volumes bofm     # Index Book of Mormon
  gospel-vec search "faith"          # Search for faith
  gospel-vec stats                   # Show database stats
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

	volumes := flags.String("volumes", "bofm", "Comma-separated volumes to index (bofm, dc-testament/dc, pgp, ot, nt)")
	layers := flags.String("layers", "verse,paragraph", "Comma-separated layers (verse, paragraph, summary, theme)")
	maxChapters := flags.Int("max", 0, "Max chapters to index (0 = all)")
	withSummary := flags.Bool("summary", false, "Generate LLM summaries (slower)")
	chatModel := flags.String("chat-model", "", "Chat model for summaries (e.g., qwen/qwen3-vl-8b)")
	verbose := flags.Bool("v", true, "Verbose output")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg := DefaultConfig()

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
			layerList = append(layerList, LayerSummary)
		case "theme":
			layerList = append(layerList, LayerTheme)
		}
	}

	// Add summary layer if requested
	if *withSummary && !containsLayer(layerList, LayerSummary) {
		layerList = append(layerList, LayerSummary)
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
		Layers:      layerList,
		Volumes:     volumeList,
		MaxChapters: *maxChapters,
		Verbose:     *verbose,
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
