package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/philippgille/chromem-go"
)

const (
	// LM Studio OpenAI-compatible endpoint
	lmStudioBaseURL = "http://localhost:1234/v1"
	// Model name as it appears in LM Studio
	// Use: text-embedding-qwen3-embedding-4b or text-embedding-qwen3-embedding-8b
	embeddingModel = "text-embedding-qwen3-embedding-4b"
)

func main() {
	experiment := flag.String("experiment", "basic", "Which experiment to run: basic, verse, compare, query")
	query := flag.String("query", "", "Query string for 'query' experiment")
	flag.Parse()

	ctx := context.Background()

	// Create embedding function using LM Studio's OpenAI-compatible API
	// LM Studio serves embeddings at /v1/embeddings just like OpenAI
	embeddingFunc := chromem.NewEmbeddingFuncOpenAICompat(
		lmStudioBaseURL,
		"", // No API key needed for local LM Studio
		embeddingModel,
		nil, // Auto-detect normalization
	)

	switch *experiment {
	case "basic":
		runBasicExperiment(ctx, embeddingFunc)
	case "verse":
		runVerseExperiment(ctx, embeddingFunc)
	case "compare":
		runCompareExperiment(ctx, embeddingFunc)
	case "query":
		if *query == "" {
			log.Fatal("Please provide a query with -query flag")
		}
		runQueryExperiment(ctx, embeddingFunc, *query)
	default:
		log.Fatalf("Unknown experiment: %s", *experiment)
	}
}

// runBasicExperiment tests basic chromem-go functionality with LM Studio
func runBasicExperiment(ctx context.Context, embeddingFunc chromem.EmbeddingFunc) {
	fmt.Println("=== Basic Experiment: Testing chromem-go with LM Studio ===")
	fmt.Println()

	// Test that embedding works
	fmt.Println("Testing embedding generation...")
	start := time.Now()
	testEmbed, err := embeddingFunc(ctx, "This is a test sentence.")
	if err != nil {
		log.Fatalf("Failed to create embedding: %v", err)
	}
	fmt.Printf("✓ Embedding created in %v\n", time.Since(start))
	fmt.Printf("  Vector dimensions: %d\n", len(testEmbed))
	fmt.Printf("  First 5 values: %v\n", testEmbed[:min(5, len(testEmbed))])
	fmt.Println()

	// Create in-memory database
	db := chromem.NewDB()

	// Create collection with our embedding function
	collection, err := db.CreateCollection("scriptures-basic", nil, embeddingFunc)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	// Sample scripture verses
	docs := []chromem.Document{
		{
			ID:      "1ne-3-7",
			Content: "And it came to pass that I, Nephi, said unto my father: I will go and do the things which the Lord hath commanded, for I know that the Lord giveth no commandments unto the children of men, save he shall prepare a way for them that they may accomplish the thing which he commandeth them.",
			Metadata: map[string]string{
				"book":    "1 Nephi",
				"chapter": "3",
				"verse":   "7",
				"volume":  "bofm",
			},
		},
		{
			ID:      "dc-93-36",
			Content: "The glory of God is intelligence, or, in other words, light and truth.",
			Metadata: map[string]string{
				"book":    "D&C",
				"chapter": "93",
				"verse":   "36",
				"volume":  "dc",
			},
		},
		{
			ID:      "moses-3-5",
			Content: "For I, the Lord God, created all things, of which I have spoken, spiritually, before they were naturally upon the face of the earth.",
			Metadata: map[string]string{
				"book":    "Moses",
				"chapter": "3",
				"verse":   "5",
				"volume":  "pgp",
			},
		},
		{
			ID:      "moroni-7-47",
			Content: "But charity is the pure love of Christ, and it endureth forever; and whoso is found possessed of it at the last day, it shall be well with him.",
			Metadata: map[string]string{
				"book":    "Moroni",
				"chapter": "7",
				"verse":   "47",
				"volume":  "bofm",
			},
		},
		{
			ID:      "dc-130-18-19",
			Content: "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection. And if a person gains more knowledge and intelligence in this life through his diligence and obedience than another, he will have so much the advantage in the world to come.",
			Metadata: map[string]string{
				"book":    "D&C",
				"chapter": "130",
				"verse":   "18-19",
				"volume":  "dc",
			},
		},
	}

	// Add documents
	fmt.Printf("Adding %d documents to collection...\n", len(docs))
	start = time.Now()
	err = collection.AddDocuments(ctx, docs, runtime.NumCPU())
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}
	fmt.Printf("✓ Documents added in %v\n", time.Since(start))
	fmt.Println()

	// Test queries
	queries := []string{
		"What does the Lord command us to do?",
		"What is intelligence?",
		"What is charity?",
		"spiritual creation before physical",
	}

	fmt.Println("=== Query Results ===")
	for _, q := range queries {
		fmt.Printf("\nQuery: %q\n", q)
		start = time.Now()
		results, err := collection.Query(ctx, q, 2, nil, nil)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}
		fmt.Printf("Search took: %v\n", time.Since(start))
		for i, r := range results {
			fmt.Printf("  %d. [%.4f] %s %s:%s - %s\n",
				i+1,
				r.Similarity,
				r.Metadata["book"],
				r.Metadata["chapter"],
				r.Metadata["verse"],
				truncate(r.Content, 80),
			)
		}
	}
}

// runVerseExperiment loads actual scripture verses and tests retrieval
func runVerseExperiment(ctx context.Context, embeddingFunc chromem.EmbeddingFunc) {
	fmt.Println("=== Verse Experiment: Loading scriptures from files ===")
	fmt.Println()

	// Try to load a sample chapter
	scriptureDir := "../../gospel-library/eng/scriptures"

	// Load Moroni 7 (charity chapter)
	moroni7Path := scriptureDir + "/bofm/moro/7.md"
	content, err := os.ReadFile(moroni7Path)
	if err != nil {
		log.Printf("Could not read %s: %v", moroni7Path, err)
		log.Println("Using sample data instead...")
		runBasicExperiment(ctx, embeddingFunc)
		return
	}

	// Parse verses from markdown
	verses := parseVersesFromMarkdown(string(content), "Moroni", "7", "bofm")

	fmt.Printf("Found %d verses in Moroni 7\n", len(verses))
	if len(verses) == 0 {
		log.Fatal("No verses found")
	}

	// Create database and collection
	db := chromem.NewDB()
	collection, err := db.CreateCollection("moroni-7", nil, embeddingFunc)
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	// Convert to documents
	docs := make([]chromem.Document, len(verses))
	for i, v := range verses {
		docs[i] = chromem.Document{
			ID:       v.ID,
			Content:  v.Content,
			Metadata: v.Metadata,
		}
	}

	// Add documents
	fmt.Printf("Adding %d verses to collection...\n", len(docs))
	start := time.Now()
	err = collection.AddDocuments(ctx, docs, runtime.NumCPU())
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}
	fmt.Printf("✓ Verses embedded in %v\n", time.Since(start))
	fmt.Println()

	// Test charity-related queries
	queries := []string{
		"What is charity?",
		"pure love of Christ",
		"faith hope and charity",
		"charity suffereth long",
		"what must I do to be saved?",
	}

	fmt.Println("=== Query Results ===")
	for _, q := range queries {
		fmt.Printf("\nQuery: %q\n", q)
		start = time.Now()
		results, err := collection.Query(ctx, q, 3, nil, nil)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}
		fmt.Printf("Search took: %v\n", time.Since(start))
		for i, r := range results {
			fmt.Printf("  %d. [%.4f] v%s - %s\n",
				i+1,
				r.Similarity,
				r.Metadata["verse"],
				truncate(r.Content, 100),
			)
		}
	}
}

// runCompareExperiment compares different chunking strategies
func runCompareExperiment(ctx context.Context, embeddingFunc chromem.EmbeddingFunc) {
	fmt.Println("=== Compare Experiment: Different Chunking Strategies ===")
	fmt.Println()

	// Create a persistent DB for this experiment
	db, err := chromem.NewPersistentDB("./data/compare-db", false)
	if err != nil {
		fmt.Printf("Failed to create db: %v\n", err)
		return
	}

	// Path to scriptures
	scripturePath := filepath.Join("..", "..", "gospel-library", "eng", "scriptures")

	// Test with limited chapters for speed
	maxChapters := 5

	// Load with different strategies
	strategies := []ChunkingStrategy{
		StrategyVerse,
		StrategyVerseContext,
		StrategyParagraph,
		StrategyChapter,
	}

	collections := make(map[ChunkingStrategy]*chromem.Collection)

	for _, strategy := range strategies {
		fmt.Printf("\n--- Loading with %s strategy ---\n", strategy)
		col, err := LoadScripturesWithStrategy(ctx, db, scripturePath, embeddingFunc, strategy, maxChapters)
		if err != nil {
			fmt.Printf("Error loading %s: %v\n", strategy, err)
			continue
		}
		collections[strategy] = col
	}

	// Test queries
	queries := []string{
		"faith in Jesus Christ",
		"repentance and baptism",
		"the gift of the Holy Ghost",
		"charity never faileth",
		"how to pray",
	}

	CompareStrategies(ctx, collections, queries)
}

// runQueryExperiment allows interactive querying against persisted data
func runQueryExperiment(ctx context.Context, embeddingFunc chromem.EmbeddingFunc, query string) {
	fmt.Println("=== Query Experiment ===")
	fmt.Printf("Query: %q\n", query)
	fmt.Println()

	// Try to load persisted DB
	db, err := chromem.NewPersistentDB("./chromem-db", false)
	if err != nil {
		log.Fatalf("Failed to load DB: %v", err)
	}

	// List collections
	collections := db.ListCollections()
	if len(collections) == 0 {
		fmt.Println("No collections found. Run -experiment=verse first to create data.")
		return
	}

	fmt.Printf("Found %d collections\n", len(collections))
	for name, col := range collections {
		fmt.Printf("\n--- Collection: %s (%d documents) ---\n", name, col.Count())

		start := time.Now()
		results, err := col.Query(ctx, query, 5, nil, nil)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}
		fmt.Printf("Search took: %v\n", time.Since(start))

		for i, r := range results {
			fmt.Printf("  %d. [%.4f] %s\n", i+1, r.Similarity, truncate(r.Content, 120))
			if r.Metadata != nil {
				fmt.Printf("     Metadata: %v\n", r.Metadata)
			}
		}
	}
}

// Verse represents a parsed scripture verse
type Verse struct {
	ID       string
	Content  string
	Metadata map[string]string
}

// parseVersesFromMarkdown extracts verses from our scripture markdown format
func parseVersesFromMarkdown(content, book, chapter, volume string) []Verse {
	var verses []Verse
	lines := strings.Split(content, "\n")

	var currentVerse string
	var verseNum string
	var verseContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for verse numbers: **1.** or **12.**
		if strings.HasPrefix(line, "**") && strings.Contains(line, ".**") {
			// Save previous verse if exists
			if currentVerse != "" {
				verses = append(verses, Verse{
					ID:      fmt.Sprintf("%s-%s-%s", strings.ToLower(book), chapter, verseNum),
					Content: strings.TrimSpace(verseContent.String()),
					Metadata: map[string]string{
						"book":    book,
						"chapter": chapter,
						"verse":   verseNum,
						"volume":  volume,
					},
				})
			}

			// Parse new verse number
			idx := strings.Index(line, ".**")
			if idx > 2 {
				verseNum = line[2:idx]
				// Get content after the verse marker
				rest := line[idx+3:]
				rest = cleanVerseText(rest)
				verseContent.Reset()
				verseContent.WriteString(rest)
				currentVerse = verseNum
			}
		} else if currentVerse != "" && line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "<") {
			// Continue accumulating verse content
			if verseContent.Len() > 0 {
				verseContent.WriteString(" ")
			}
			verseContent.WriteString(cleanVerseText(line))
		}
	}

	// Don't forget the last verse
	if currentVerse != "" && verseContent.Len() > 0 {
		verses = append(verses, Verse{
			ID:      fmt.Sprintf("%s-%s-%s", strings.ToLower(book), chapter, verseNum),
			Content: strings.TrimSpace(verseContent.String()),
			Metadata: map[string]string{
				"book":    book,
				"chapter": chapter,
				"verse":   verseNum,
				"volume":  volume,
			},
		})
	}

	return verses
}

// cleanVerseText removes markdown formatting from verse text
func cleanVerseText(text string) string {
	// Remove footnote references like <sup>[1a](#fn-1a)</sup>
	for strings.Contains(text, "<sup>") {
		start := strings.Index(text, "<sup>")
		end := strings.Index(text, "</sup>")
		if end > start {
			text = text[:start] + text[end+6:]
		} else {
			break
		}
	}

	// Remove markdown links [text](url) -> text
	for strings.Contains(text, "](") {
		start := strings.LastIndex(text[:strings.Index(text, "](")+1], "[")
		end := strings.Index(text, ")")
		if start >= 0 && end > start {
			linkText := text[start+1 : strings.Index(text, "](")]
			text = text[:start] + linkText + text[end+1:]
		} else {
			break
		}
	}

	// Remove other formatting
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "¶ ", "")

	return text
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
