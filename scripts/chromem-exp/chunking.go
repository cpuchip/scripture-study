package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/philippgille/chromem-go"
)

// ChunkingStrategy defines how to split content
type ChunkingStrategy string

const (
	StrategyVerse        ChunkingStrategy = "verse"         // Single verse
	StrategyVerseContext ChunkingStrategy = "verse_context" // Verse with surrounding context
	StrategyParagraph    ChunkingStrategy = "paragraph"     // Group of verses
	StrategyChapter      ChunkingStrategy = "chapter"       // Full chapter
)

// ScriptureChunk represents a chunk of scripture content
type ScriptureChunk struct {
	ID       string
	Content  string
	Metadata map[string]string
}

// LoadScripturesWithStrategy loads scriptures from a directory using the specified chunking strategy
func LoadScripturesWithStrategy(
	ctx context.Context,
	db *chromem.DB,
	scripturePath string,
	embeddingFunc chromem.EmbeddingFunc,
	strategy ChunkingStrategy,
	maxChapters int,
) (*chromem.Collection, error) {
	collectionName := fmt.Sprintf("scriptures-%s", strategy)

	// Try to get existing collection
	collection := db.GetCollection(collectionName, embeddingFunc)
	if collection != nil && collection.Count() > 0 {
		fmt.Printf("Collection %s already exists with %d documents\n", collectionName, collection.Count())
		return collection, nil
	}

	// Create new collection
	collection, err := db.GetOrCreateCollection(collectionName, nil, embeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	// Find scripture files
	var docs []chromem.Document
	chaptersProcessed := 0

	// Process Book of Mormon
	bofmPath := filepath.Join(scripturePath, "bofm")
	books := []string{"1-ne", "2-ne", "jacob", "enos", "jarom", "omni", "w-of-m", "mosiah", "alma", "hel", "3-ne", "4-ne", "morm", "ether", "moro"}

	for _, book := range books {
		if maxChapters > 0 && chaptersProcessed >= maxChapters {
			break
		}

		bookPath := filepath.Join(bofmPath, book)
		entries, err := os.ReadDir(bookPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if maxChapters > 0 && chaptersProcessed >= maxChapters {
				break
			}

			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "index.md" {
				chapterPath := filepath.Join(bookPath, entry.Name())
				content, err := os.ReadFile(chapterPath)
				if err != nil {
					continue
				}

				chapterNum := strings.TrimSuffix(entry.Name(), ".md")
				bookName := formatBookName(book)

				chunks := chunkContent(string(content), bookName, chapterNum, "bofm", strategy)
				for _, chunk := range chunks {
					docs = append(docs, chromem.Document{
						ID:       chunk.ID,
						Content:  chunk.Content,
						Metadata: chunk.Metadata,
					})
				}

				chaptersProcessed++
			}
		}
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents found")
	}

	fmt.Printf("Adding %d documents to collection %s...\n", len(docs), collectionName)
	start := time.Now()

	// Add in batches to avoid memory issues
	batchSize := 50
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}

		err := collection.AddDocuments(ctx, docs[i:end], runtime.NumCPU())
		if err != nil {
			return nil, fmt.Errorf("failed to add documents: %w", err)
		}

		fmt.Printf("  Added %d/%d documents...\n", end, len(docs))
	}

	fmt.Printf("✓ All documents added in %v\n", time.Since(start))
	return collection, nil
}

// chunkContent splits content based on strategy
func chunkContent(content, book, chapter, volume string, strategy ChunkingStrategy) []ScriptureChunk {
	switch strategy {
	case StrategyChapter:
		return chunkByChapter(content, book, chapter, volume)
	case StrategyVerseContext:
		return chunkByVerseWithContext(content, book, chapter, volume)
	case StrategyParagraph:
		return chunkByParagraph(content, book, chapter, volume)
	default:
		return chunkByVerse(content, book, chapter, volume)
	}
}

// chunkByVerse splits into individual verses
func chunkByVerse(content, book, chapter, volume string) []ScriptureChunk {
	verses := parseVersesFromMarkdown(content, book, chapter, volume)
	chunks := make([]ScriptureChunk, len(verses))
	for i, v := range verses {
		chunks[i] = ScriptureChunk{
			ID:       v.ID,
			Content:  v.Content,
			Metadata: v.Metadata,
		}
	}
	return chunks
}

// chunkByVerseWithContext includes +/- 1 verse as context
func chunkByVerseWithContext(content, book, chapter, volume string) []ScriptureChunk {
	verses := parseVersesFromMarkdown(content, book, chapter, volume)
	var chunks []ScriptureChunk

	for i, v := range verses {
		var contextContent strings.Builder

		// Add previous verse as context
		if i > 0 {
			contextContent.WriteString("[Previous: ")
			contextContent.WriteString(truncate(verses[i-1].Content, 100))
			contextContent.WriteString("] ")
		}

		// Add current verse
		contextContent.WriteString(v.Content)

		// Add next verse as context
		if i < len(verses)-1 {
			contextContent.WriteString(" [Next: ")
			contextContent.WriteString(truncate(verses[i+1].Content, 100))
			contextContent.WriteString("]")
		}

		chunks = append(chunks, ScriptureChunk{
			ID:      v.ID + "-ctx",
			Content: contextContent.String(),
			Metadata: map[string]string{
				"book":     book,
				"chapter":  chapter,
				"verse":    v.Metadata["verse"],
				"volume":   volume,
				"strategy": "verse_context",
			},
		})
	}

	return chunks
}

// chunkByParagraph groups verses into paragraphs (every 5 verses)
func chunkByParagraph(content, book, chapter, volume string) []ScriptureChunk {
	verses := parseVersesFromMarkdown(content, book, chapter, volume)
	var chunks []ScriptureChunk

	paragraphSize := 5
	for i := 0; i < len(verses); i += paragraphSize {
		end := i + paragraphSize
		if end > len(verses) {
			end = len(verses)
		}

		var paragraphContent strings.Builder
		verseRange := verses[i].Metadata["verse"]

		for j := i; j < end; j++ {
			if j > i {
				paragraphContent.WriteString(" ")
			}
			paragraphContent.WriteString(verses[j].Content)
		}

		if end-1 > i {
			verseRange = fmt.Sprintf("%s-%s", verses[i].Metadata["verse"], verses[end-1].Metadata["verse"])
		}

		chunks = append(chunks, ScriptureChunk{
			ID:      fmt.Sprintf("%s-%s-%s", strings.ToLower(book), chapter, verseRange),
			Content: paragraphContent.String(),
			Metadata: map[string]string{
				"book":     book,
				"chapter":  chapter,
				"verses":   verseRange,
				"volume":   volume,
				"strategy": "paragraph",
			},
		})
	}

	return chunks
}

// chunkByChapter keeps the whole chapter as one chunk
func chunkByChapter(content, book, chapter, volume string) []ScriptureChunk {
	verses := parseVersesFromMarkdown(content, book, chapter, volume)

	var chapterContent strings.Builder
	for i, v := range verses {
		if i > 0 {
			chapterContent.WriteString(" ")
		}
		chapterContent.WriteString(v.Content)
	}

	verseRange := "1"
	if len(verses) > 1 {
		verseRange = fmt.Sprintf("1-%s", verses[len(verses)-1].Metadata["verse"])
	}

	return []ScriptureChunk{
		{
			ID:      fmt.Sprintf("%s-%s-full", strings.ToLower(book), chapter),
			Content: chapterContent.String(),
			Metadata: map[string]string{
				"book":     book,
				"chapter":  chapter,
				"verses":   verseRange,
				"volume":   volume,
				"strategy": "chapter",
			},
		},
	}
}

// formatBookName converts directory names to readable book names
func formatBookName(dir string) string {
	names := map[string]string{
		"1-ne":   "1 Nephi",
		"2-ne":   "2 Nephi",
		"jacob":  "Jacob",
		"enos":   "Enos",
		"jarom":  "Jarom",
		"omni":   "Omni",
		"w-of-m": "Words of Mormon",
		"mosiah": "Mosiah",
		"alma":   "Alma",
		"hel":    "Helaman",
		"3-ne":   "3 Nephi",
		"4-ne":   "4 Nephi",
		"morm":   "Mormon",
		"ether":  "Ether",
		"moro":   "Moroni",
	}
	if name, ok := names[dir]; ok {
		return name
	}
	return dir
}

// CompareStrategies runs the same queries against different chunking strategies
func CompareStrategies(ctx context.Context, collections map[ChunkingStrategy]*chromem.Collection, queries []string) {
	fmt.Println("\n=== Strategy Comparison ===\n")

	for _, query := range queries {
		fmt.Printf("Query: %q\n", query)
		fmt.Println(strings.Repeat("-", 80))

		for strategy, col := range collections {
			if col == nil {
				continue
			}

			start := time.Now()
			results, err := col.Query(ctx, query, 3, nil, nil)
			if err != nil {
				fmt.Printf("  [%s] Error: %v\n", strategy, err)
				continue
			}

			fmt.Printf("  [%s] (%v, %d docs)\n", strategy, time.Since(start), col.Count())
			for i, r := range results {
				ref := formatReference(r.Metadata)
				fmt.Printf("    %d. [%.4f] %s: %s\n", i+1, r.Similarity, ref, truncate(r.Content, 60))
			}
		}
		fmt.Println()
	}
}

func formatReference(metadata map[string]string) string {
	book := metadata["book"]
	chapter := metadata["chapter"]
	verse := metadata["verse"]
	if verse == "" {
		verse = metadata["verses"]
	}
	return fmt.Sprintf("%s %s:%s", book, chapter, verse)
}
