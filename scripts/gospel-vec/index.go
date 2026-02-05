package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Indexer handles indexing content into the store
type Indexer struct {
	store      *Store
	summarizer *Summarizer
	cache      *SummaryCache
	config     *Config
}

// NewIndexer creates a new indexer
func NewIndexer(store *Store, summarizer *Summarizer, cfg *Config) *Indexer {
	// Create cache in data/summaries/
	cacheDir := filepath.Join(cfg.DataDir, "summaries")
	return &Indexer{
		store:      store,
		summarizer: summarizer,
		cache:      NewSummaryCache(cacheDir),
		config:     cfg,
	}
}

// IndexOptions controls what gets indexed
type IndexOptions struct {
	Layers      []Layer  // Which layers to index
	Volumes     []string // Which scripture volumes (bofm, dc-testament/dc, etc.)
	MaxChapters int      // Max chapters to index (0 = all)
	Verbose     bool     // Print progress details
	UseCache    bool     // Use cached summaries if available
}

// DefaultIndexOptions returns sensible defaults
func DefaultIndexOptions() IndexOptions {
	return IndexOptions{
		Layers:      []Layer{LayerVerse, LayerParagraph},
		Volumes:     []string{"bofm"},
		MaxChapters: 0,
		Verbose:     true,
		UseCache:    true, // Use cache by default
	}
}

// IndexScriptures indexes scripture content
func (idx *Indexer) IndexScriptures(ctx context.Context, opts IndexOptions) error {
	// Find all scripture files
	files, err := FindScriptureFiles(idx.config.ScripturesPath, opts.Volumes...)
	if err != nil {
		return fmt.Errorf("finding scripture files: %w", err)
	}

	if opts.MaxChapters > 0 && len(files) > opts.MaxChapters {
		files = files[:opts.MaxChapters]
	}

	if opts.Verbose {
		fmt.Printf("📚 Found %d chapter files to index\n", len(files))
	}

	// Process each chapter
	var totalChunks int
	start := time.Now()

	for i, filePath := range files {
		chapter, err := ParseChapterFile(filePath)
		if err != nil {
			fmt.Printf("⚠️  Error parsing %s: %v\n", filePath, err)
			continue
		}

		if len(chapter.Verses) == 0 {
			continue
		}

		var chunks []Chunk

		// Build chunks based on requested layers
		for _, layer := range opts.Layers {
			switch layer {
			case LayerVerse:
				chunks = append(chunks, ChunkByVerse(chapter, SourceScriptures)...)
			case LayerParagraph:
				chunks = append(chunks, ChunkByParagraph(chapter, SourceScriptures)...)
			case LayerSummary:
				if idx.config.ChatModel == "" && !opts.UseCache {
					continue // No model and no cache
				}

				// Try cache first (validates model and prompt version)
				var summary *ChapterSummary
				if opts.UseCache {
					summary = idx.cache.GetSummary(chapter.Book, chapter.Chapter, idx.config.ChatModel)
					if summary != nil && opts.Verbose {
						fmt.Printf(" [summary: cached]")
					}
				}

				// Generate if not cached
				if summary == nil && idx.summarizer != nil && idx.config.ChatModel != "" {
					summaryStart := time.Now()
					var err error
					summary, err = idx.summarizer.SummarizeChapter(ctx, chapter.Book, chapter.Chapter, GetFullChapterContent(chapter))
					summaryDur := time.Since(summaryStart)
					if err != nil {
						fmt.Printf("⚠️  Error summarizing %s %d: %v\n", chapter.Book, chapter.Chapter, err)
					} else {
						// Cache the result
						if cacheErr := idx.cache.SaveSummary(chapter.Book, chapter.Chapter, idx.config.ChatModel, summary); cacheErr != nil {
							fmt.Printf("⚠️  Cache save error: %v\n", cacheErr)
						}
						if opts.Verbose {
							fmt.Printf(" [summary: %v]", summaryDur.Round(time.Millisecond))
						}
					}
				}

				if summary != nil {
					chunk := ChunkAsChapterSummary(chapter, SourceScriptures, summary, idx.config.ChatModel)
					chunks = append(chunks, chunk)
				}

			case LayerTheme:
				if idx.config.ChatModel == "" && !opts.UseCache {
					continue // No model and no cache
				}

				// Try cache first (validates model and prompt version)
				var themes []ThemeRange
				if opts.UseCache {
					themes = idx.cache.GetThemes(chapter.Book, chapter.Chapter, idx.config.ChatModel)
					if len(themes) > 0 && opts.Verbose {
						fmt.Printf(" [themes: %d cached]", len(themes))
					}
				}

				// Generate if not cached
				if len(themes) == 0 && idx.summarizer != nil && idx.config.ChatModel != "" {
					themeStart := time.Now()
					var err error
					themes, err = idx.summarizer.DetectThemes(ctx, chapter.Book, chapter.Chapter, GetVerseTexts(chapter))
					themeDur := time.Since(themeStart)
					if err != nil {
						fmt.Printf("⚠️  Error detecting themes in %s %d: %v\n", chapter.Book, chapter.Chapter, err)
					} else {
						// Cache the result
						if cacheErr := idx.cache.SaveThemes(chapter.Book, chapter.Chapter, idx.config.ChatModel, themes); cacheErr != nil {
							fmt.Printf("⚠️  Cache save error: %v\n", cacheErr)
						}
						if opts.Verbose {
							fmt.Printf(" [themes: %d in %v]", len(themes), themeDur.Round(time.Millisecond))
						}
					}
				}

				for _, theme := range themes {
					chunk := ChunkAsTheme(chapter, SourceScriptures, theme, idx.config.ChatModel)
					chunks = append(chunks, chunk)
				}
			}
		}

		// Add chunks to store
		if len(chunks) > 0 {
			embedStart := time.Now()
			if err := idx.store.AddChunks(ctx, chunks); err != nil {
				return fmt.Errorf("adding chunks for %s: %w", filePath, err)
			}
			embedDur := time.Since(embedStart)
			totalChunks += len(chunks)

			if opts.Verbose {
				fmt.Printf(" [embed: %v]", embedDur.Round(time.Millisecond))
			}
		}

		if opts.Verbose {
			fmt.Printf("\n📖 Indexed %d/%d: %s %d (%d chunks)",
				i+1, len(files), chapter.Book, chapter.Chapter, len(chunks))
		}
	}

	if opts.Verbose {
		fmt.Printf("\n✅ Indexed %d chunks in %v\n", totalChunks, time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// IndexChapterWithSummary indexes a single chapter with LLM-generated summary
func (idx *Indexer) IndexChapterWithSummary(ctx context.Context, filePath string) error {
	chapter, err := ParseChapterFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing chapter: %w", err)
	}

	var chunks []Chunk

	// Always add verse and paragraph layers
	chunks = append(chunks, ChunkByVerse(chapter, SourceScriptures)...)
	chunks = append(chunks, ChunkByParagraph(chapter, SourceScriptures)...)

	// Add summary if summarizer is available
	if idx.summarizer != nil && idx.config.ChatModel != "" {
		summary, err := idx.summarizer.SummarizeChapter(ctx, chapter.Book, chapter.Chapter, GetFullChapterContent(chapter))
		if err != nil {
			return fmt.Errorf("generating summary: %w", err)
		}
		chunks = append(chunks, ChunkAsChapterSummary(chapter, SourceScriptures, summary, idx.config.ChatModel))
	}

	return idx.store.AddChunks(ctx, chunks)
}

// TalkIndexOptions controls conference talk indexing
type TalkIndexOptions struct {
	Layers   []Layer  // Which layers to index (paragraph, summary)
	Years    []string // Which years to index (empty = all)
	MaxTalks int      // Max talks to index (0 = all)
	Verbose  bool     // Print progress details
	UseCache bool     // Use cached summaries if available
}

// DefaultTalkIndexOptions returns sensible defaults
func DefaultTalkIndexOptions() TalkIndexOptions {
	return TalkIndexOptions{
		Layers:   []Layer{LayerParagraph, LayerSummary},
		Years:    nil, // All years
		MaxTalks: 0,   // All talks
		Verbose:  true,
		UseCache: true,
	}
}

// IndexConferenceTalks indexes conference talk content
func (idx *Indexer) IndexConferenceTalks(ctx context.Context, basePath string, opts TalkIndexOptions) error {
	// Find all talk files
	var allFiles []string
	var err error

	if len(opts.Years) == 0 {
		// No years specified - scan all
		allFiles, err = FindTalkFiles(basePath)
	} else {
		// Specific years requested
		for _, year := range opts.Years {
			files, ferr := FindTalkFiles(basePath, year)
			if ferr != nil {
				return fmt.Errorf("finding talk files for %s: %w", year, ferr)
			}
			allFiles = append(allFiles, files...)
		}
	}
	if err != nil {
		return fmt.Errorf("finding talk files: %w", err)
	}

	if opts.MaxTalks > 0 && len(allFiles) > opts.MaxTalks {
		allFiles = allFiles[:opts.MaxTalks]
	}

	if opts.Verbose {
		fmt.Printf("📚 Found %d talk files to index\n", len(allFiles))
	}

	// Process each talk
	var totalChunks int
	var indexedTalks int
	start := time.Now()

	for i, filePath := range allFiles {
		talk, err := ParseTalkFile(filePath)
		if err != nil {
			fmt.Printf("⚠️  Error parsing %s: %v\n", filePath, err)
			continue
		}

		// Skip administrative documents
		if talk.IsAdministrativeDocument() {
			if opts.Verbose {
				fmt.Printf("⏭️  Skipping %s (administrative)\n", filepath.Base(filePath))
			}
			continue
		}

		// Skip talks with too few paragraphs
		if len(talk.Paragraphs) < 3 {
			if opts.Verbose {
				fmt.Printf("⏭️  Skipping %s (too short: %d paragraphs)\n", filepath.Base(filePath), len(talk.Paragraphs))
			}
			continue
		}

		var chunks []Chunk

		// Build chunks based on requested layers
		for _, layer := range opts.Layers {
			switch layer {
			case LayerParagraph:
				chunks = append(chunks, ChunkTalkByParagraph(talk)...)

			case LayerSummary:
				if idx.config.ChatModel == "" && !opts.UseCache {
					continue // No model and no cache
				}

				// Try cache first
				var summary *ChapterSummary
				cacheKey := fmt.Sprintf("talk-%s-%s-%s", talk.Metadata.Year, talk.Metadata.Month, filepath.Base(filePath))
				if opts.UseCache {
					summary = idx.cache.GetSummary(cacheKey, 0, idx.config.ChatModel)
					if summary != nil && opts.Verbose {
						fmt.Printf(" [summary: cached]")
					}
				}

				// Generate if not cached
				if summary == nil && idx.summarizer != nil && idx.config.ChatModel != "" {
					summaryStart := time.Now()
					summary, err = idx.generateTalkSummary(ctx, talk)
					summaryDur := time.Since(summaryStart)
					if err != nil {
						fmt.Printf("⚠️  Error summarizing %s: %v\n", talk.Metadata.Title, err)
					} else {
						// Cache the result
						if cacheErr := idx.cache.SaveSummary(cacheKey, 0, idx.config.ChatModel, summary); cacheErr != nil {
							fmt.Printf("⚠️  Cache save error: %v\n", cacheErr)
						}
						if opts.Verbose {
							fmt.Printf(" [summary: %v]", summaryDur.Round(time.Millisecond))
						}
					}
				}

				if summary != nil {
					chunks = append(chunks, ChunkTalkAsSummary(talk, summary, idx.config.ChatModel))
				}
			}
		}

		// Add chunks to store
		if len(chunks) > 0 {
			embedStart := time.Now()
			if err := idx.store.AddChunks(ctx, chunks); err != nil {
				return fmt.Errorf("adding chunks for %s: %w", filePath, err)
			}
			embedDur := time.Since(embedStart)
			totalChunks += len(chunks)
			indexedTalks++

			if opts.Verbose {
				fmt.Printf("\n📖 Indexed %d/%d: %s (%d chunks) [embed: %v]",
					i+1, len(allFiles), talk.Metadata.Title, len(chunks), embedDur.Round(time.Millisecond))
			}
		}
	}

	if opts.Verbose {
		fmt.Printf("\n✅ Indexed %d talks with %d total chunks in %v\n",
			indexedTalks, totalChunks, time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// generateTalkSummary creates an AI summary of a conference talk
func (idx *Indexer) generateTalkSummary(ctx context.Context, talk *ParsedTalk) (*ChapterSummary, error) {
	// Build content from paragraphs (limit to avoid token overflow)
	var content strings.Builder
	for i, para := range talk.Paragraphs {
		if i > 25 { // Limit paragraphs to summarize
			content.WriteString("\n[Additional content truncated]")
			break
		}
		content.WriteString(para)
		content.WriteString("\n\n")
	}

	systemPrompt := `Create a summary of this conference talk optimized for semantic search indexing.

Format your response EXACTLY like this:
KEYWORDS: [10-15 comma-separated searchable terms including speaker themes, doctrines, people, events]
SUMMARY: [50-75 word narrative covering main message and teachings, present tense]
KEY_QUOTE: [Most memorable or powerful quote from the talk]

Keep output under 200 words total. No other text.`

	userPrompt := fmt.Sprintf(`Summarize this %s %s General Conference talk by %s titled "%s":

%s`, talk.Metadata.Month, talk.Metadata.Year, talk.Metadata.Speaker, talk.Metadata.Title, content.String())

	response, err := idx.summarizer.chat(ctx, systemPrompt, userPrompt, 300)
	if err != nil {
		return nil, err
	}

	// Parse response
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
