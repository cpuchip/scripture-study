// gospel-engine: Unified gospel content search engine (SQLite FTS5 + chromem-go vector search).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/config"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/enricher"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/indexer"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/llm"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/mcp"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/search"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/vec"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "serve":
		runServe()
	case "index":
		runIndex()
	case "enrich":
		runEnrich()
	case "enrich-scriptures":
		runEnrichScriptures()
	case "embed-enrichments":
		runEmbedEnrichments()
	case "convert":
		runConvert()
	case "stats":
		runStats()
	case "search":
		runSearchTest()
	case "version":
		fmt.Printf("gospel-engine %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `gospel-engine %s — unified gospel content search

Commands:
  serve     Start the MCP server (stdin/stdout)
  index     Index content into SQLite + vector store
  enrich    Run TITSW enrichment on conference talks
  enrich-scriptures  Run scripture enrichment on chapters (lens approach)
  embed-enrichments  Vectorize enrichment output (summaries, keywords, etc.)
  search    Test search from the command line
  convert   Convert gob.gz vector data to mmap-friendly .vecf format
  stats     Show index statistics
  version   Print version
  help      Show this help

Index flags:
  --source=TYPE     Only index: scriptures, conference, manual, books, music
  --full            Full reindex (ignore incremental cache)
  --no-vectors      Skip vector embedding (SQLite only)
  --verbose         Print progress

Enrich flags:
  --limit=N         Process at most N talks (default: all unenriched)
  --force           Re-enrich talks that already have TITSW profiles
  --concurrency=N   Parallel LLM requests (default: 1)
  --year=YYYY       Only enrich talks from this year
  --speaker=NAME    Only enrich talks by this speaker
  --temperature=F   LLM temperature (default: 0.2)
  --verbose         Print progress

Enrich-scriptures flags:
  --limit=N         Process at most N chapters (default: all unenriched)
  --force           Re-enrich chapters that already have enrichment
  --concurrency=N   Parallel LLM requests (default: 1)
  --volume=VOL      Only enrich: ot, nt, bofm, dc-testament, pgp
  --book=BOOK       Only enrich this book (e.g., alma, gen, dc)
  --chapter=N       Only enrich this chapter number
  --temperature=F   LLM temperature (default: 0.2)
  --verbose         Print progress

Search flags:
  --mode=MODE       Search mode: keyword, semantic, combined (default: combined)
  --source=SOURCE   Filter sources: scriptures, conference, manual, books (comma-separated)
  --limit=N         Max results (default: 10)

Environment variables:
  GOSPEL_ENGINE_DATA_DIR       Data directory (default: auto-detected)
  GOSPEL_ENGINE_DB             SQLite database path
  GOSPEL_ENGINE_EMBEDDING_URL  LM Studio embedding endpoint
  GOSPEL_ENGINE_EMBEDDING_MODEL Embedding model name
  GOSPEL_ENGINE_CHAT_URL       LM Studio chat endpoint
  GOSPEL_ENGINE_CHAT_MODEL     Chat model name for TITSW enrichment
  GOSPEL_ENGINE_ROOT           Workspace root path
`, version)
}

func runServe() {
	cfg := config.Default()
	start := time.Now()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	var searcher vec.Searcher

	// Prefer mmap store when .vecf files exist (instant startup)
	if vec.VecFilesExist(cfg.DataDir) {
		embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)
		mmapStore, err := vec.NewMmapStore(cfg.DataDir, cfg.DBPath, embedFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: mmap store unavailable: %v\n", err)
		} else {
			searcher = mmapStore
			fmt.Fprintf(os.Stderr, "⚡ mmap store ready in %v\n", time.Since(start).Round(time.Millisecond))
			defer mmapStore.Close()
		}
	}

	// Fall back to chromem-go gob.gz store (slow startup, full compatibility)
	if searcher == nil {
		embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)
		store, err := vec.NewStore(cfg.DataDir, embedFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: vector store unavailable: %v\n", err)
		} else {
			searcher = store
			fmt.Fprintf(os.Stderr, "📂 gob.gz store loaded in %v\n", time.Since(start).Round(time.Millisecond))
		}
	}

	handler := mcp.Handler(database, searcher, cfg.Root)
	server := mcp.NewServer(mcp.Tools(), handler)

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func runIndex() {
	cfg := config.Default()

	// Parse flags
	opts := indexer.DefaultOptions()
	noVectors := false

	for _, arg := range os.Args[2:] {
		switch {
		case strings.HasPrefix(arg, "--source="):
			opts.Source = strings.TrimPrefix(arg, "--source=")
		case arg == "--full":
			opts.Incremental = false
		case arg == "--no-vectors":
			noVectors = true
		case arg == "--verbose":
			opts.Verbose = true
		case arg == "--quiet":
			opts.Verbose = false
		}
	}

	if noVectors {
		opts.Layers = nil
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	var store *vec.Store
	if !noVectors {
		embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)

		// Ensure the embedding model is loaded with the correct context length
		if opts.Verbose {
			fmt.Printf("🔗 Checking embedding model %s (context=%d)...\n", cfg.EmbeddingModel, cfg.EmbeddingContextLength)
		}
		if err := vec.EnsureModelLoaded(context.Background(), cfg.EmbeddingURL, cfg.EmbeddingModel, cfg.EmbeddingContextLength); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Could not load embedding model: %v\nFalling back to SQLite-only indexing.\n", err)
			opts.Layers = nil
		} else {
			if opts.Verbose {
				fmt.Println("✅ Embedding model ready")
			}

			// Test embedding connection
			if opts.Verbose {
				fmt.Println("🔗 Testing embedding connection...")
			}
			if err := vec.TestEmbedding(context.Background(), cfg.EmbeddingURL, cfg.EmbeddingModel); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Embedding not available: %v\nFalling back to SQLite-only indexing.\n", err)
				opts.Layers = nil
			} else {
				store, err = vec.NewStore(cfg.DataDir, embedFunc)
				if err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  Vector store error: %v\nFalling back to SQLite-only indexing.\n", err)
					opts.Layers = nil
				}
			}
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	idx := indexer.New(database, store, cfg.Root)

	if opts.Verbose {
		fmt.Printf("📚 Starting index (incremental=%t, source=%s, layers=%v)\n",
			opts.Incremental, opts.Source, opts.Layers)
		fmt.Printf("   DB: %s\n", cfg.DBPath)
		fmt.Printf("   Root: %s\n", cfg.Root)
	}

	result, err := idx.Index(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Index error: %v\n", err)
		os.Exit(1)
	}

	// Save vector store
	if store != nil {
		if opts.Verbose {
			fmt.Println("💾 Saving vector store...")
		}
		if err := store.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: vector save error: %v\n", err)
		}
	}

	// Print results
	fmt.Printf("\n✅ Indexing complete in %v\n", result.Duration.Round(1e6))
	fmt.Printf("   Files processed: %d (skipped: %d)\n", result.FilesProcessed, result.FilesSkipped)
	fmt.Printf("   Scriptures: %d verses, %d chapters\n", result.ScripturesIndexed, result.ChaptersIndexed)
	fmt.Printf("   Talks: %d\n", result.TalksIndexed)
	fmt.Printf("   Manuals: %d\n", result.ManualsIndexed)
	fmt.Printf("   Books: %d\n", result.BooksIndexed)
	fmt.Printf("   Music: %d\n", result.MusicIndexed)
	fmt.Printf("   Cross-refs: %d\n", result.CrossRefsIndexed)
	fmt.Printf("   Vector chunks: %d\n", result.VecChunksAdded)

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  %d errors:\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("   - %s\n", e)
		}
	}
}

func runEnrich() {
	cfg := config.Default()

	// Parse flags
	limit := 0
	force := false
	concurrency := 1
	year := 0
	speaker := ""
	temperature := 0.2
	verbose := true

	cliArgs := os.Args[2:]
	for i := 0; i < len(cliArgs); i++ {
		arg := cliArgs[i]
		// nextVal returns the value for a flag, supporting both --flag=val and --flag val.
		nextVal := func(prefix string) string {
			if strings.HasPrefix(arg, prefix+"=") {
				return strings.TrimPrefix(arg, prefix+"=")
			}
			if arg == prefix && i+1 < len(cliArgs) {
				i++
				return cliArgs[i]
			}
			return ""
		}
		switch {
		case strings.HasPrefix(arg, "--limit"):
			fmt.Sscanf(nextVal("--limit"), "%d", &limit)
		case arg == "--force":
			force = true
		case strings.HasPrefix(arg, "--concurrency"):
			fmt.Sscanf(nextVal("--concurrency"), "%d", &concurrency)
		case strings.HasPrefix(arg, "--year"):
			fmt.Sscanf(nextVal("--year"), "%d", &year)
		case strings.HasPrefix(arg, "--speaker"):
			speaker = nextVal("--speaker")
		case strings.HasPrefix(arg, "--temperature"):
			fmt.Sscanf(nextVal("--temperature"), "%f", &temperature)
		case arg == "--verbose":
			verbose = true
		case arg == "--quiet":
			verbose = false
		}
	}

	if cfg.ChatModel == "" {
		fmt.Fprintln(os.Stderr, "Error: GOSPEL_ENGINE_CHAT_MODEL must be set (e.g., magistral-small-2509)")
		os.Exit(1)
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	client := llm.NewClient(cfg.ChatURL, cfg.ChatModel)
	enrich := enricher.New(client, temperature)

	// Build query for unenriched talks
	query := `SELECT id, year, month, speaker, title, content FROM talks`
	var conditions []string
	var args []any

	if !force {
		conditions = append(conditions, "titsw_teach IS NULL")
	}
	if year > 0 {
		conditions = append(conditions, "year = ?")
		args = append(args, year)
	}
	if speaker != "" {
		conditions = append(conditions, "speaker LIKE ?")
		args = append(args, "%"+speaker+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY year DESC, month DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying talks: %v\n", err)
		os.Exit(1)
	}

	type talkRow struct {
		ID      int64
		Year    int
		Month   int
		Speaker string
		Title   string
		Content string
	}
	var talks []talkRow
	for rows.Next() {
		var t talkRow
		if err := rows.Scan(&t.ID, &t.Year, &t.Month, &t.Speaker, &t.Title, &t.Content); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning talk: %v\n", err)
			continue
		}
		talks = append(talks, t)
	}
	rows.Close()

	if len(talks) == 0 {
		fmt.Println("No talks to enrich.")
		return
	}

	if verbose {
		fmt.Printf("🧠 Enriching %d talks with TITSW profiles\n", len(talks))
		fmt.Printf("   Model: %s (T=%.1f)\n", cfg.ChatModel, temperature)
		fmt.Printf("   Concurrency: %d\n\n", concurrency)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	start := time.Now()
	var enriched, errors int64
	var mu sync.Mutex // protects verbose output

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i, talk := range talks {
		select {
		case <-ctx.Done():
			break
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // acquire slot

		go func(idx int, talk talkRow) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			select {
			case <-ctx.Done():
				return
			default:
			}

			talkStart := time.Now()
			if verbose {
				mu.Lock()
				fmt.Printf("[%d/%d] %s — %s (%d/%02d)... ",
					idx+1, len(talks), talk.Speaker, talk.Title, talk.Year, talk.Month)
				mu.Unlock()
			}

			profile, err := enrich.Enrich(ctx, talk.Content)
			if err != nil {
				atomic.AddInt64(&errors, 1)
				if verbose {
					mu.Lock()
					fmt.Printf("[%d/%d] ❌ %v\n", idx+1, len(talks), err)
					mu.Unlock()
				}
				return
			}

			// Update the talk row with TITSW data
			_, err = database.Exec(`
				UPDATE talks SET
					titsw_dominant = ?,
					titsw_mode = ?,
					titsw_pattern = ?,
					titsw_teach = ?,
					titsw_help = ?,
					titsw_love = ?,
					titsw_spirit = ?,
					titsw_doctrine = ?,
					titsw_invite = ?,
					titsw_summary = ?,
					titsw_key_quote = ?,
					titsw_keywords = ?,
					titsw_reasoning = ?,
					titsw_raw_output = ?,
					titsw_model = ?
				WHERE id = ?
			`, profile.Dominant, profile.Mode, profile.Pattern,
				profile.Teach, profile.Help, profile.Love,
				profile.Spirit, profile.Doctrine, profile.Invite,
				profile.Summary, profile.KeyQuote, profile.Keywords,
				profile.Reasoning, profile.RawOutput, cfg.ChatModel,
				talk.ID)
			if err != nil {
				atomic.AddInt64(&errors, 1)
				if verbose {
					mu.Lock()
					fmt.Printf("[%d/%d] ❌ DB: %v\n", idx+1, len(talks), err)
					mu.Unlock()
				}
				return
			}

			n := atomic.AddInt64(&enriched, 1)
			elapsed := time.Since(talkStart).Round(time.Millisecond)
			if verbose {
				mu.Lock()
				fmt.Printf("[%d/%d] ✅ [%d,%d,%d,%d,%d,%d] %v (total: %d)\n",
					idx+1, len(talks),
					profile.Teach, profile.Help, profile.Love,
					profile.Spirit, profile.Doctrine, profile.Invite,
					elapsed, n)
				mu.Unlock()
			}
		}(i, talk)
	}

	wg.Wait()

	totalTime := time.Since(start).Round(time.Second)
	finalEnriched := atomic.LoadInt64(&enriched)
	finalErrors := atomic.LoadInt64(&errors)
	fmt.Printf("\n✅ Enrichment complete: %d/%d talks in %v (%d errors)\n",
		finalEnriched, len(talks), totalTime, finalErrors)
	if finalEnriched > 0 {
		avgTime := totalTime / time.Duration(finalEnriched)
		remaining := int64(len(talks)) - finalEnriched
		if remaining > 0 && finalErrors > 0 {
			fmt.Printf("   Average: %v/talk, ~%v for remaining %d\n",
				avgTime, avgTime*time.Duration(remaining), remaining)
		}
	}
}

func runEnrichScriptures() {
	cfg := config.Default()

	// Parse flags
	limit := 0
	force := false
	concurrency := 1
	volume := ""
	book := ""
	chapter := 0
	temperature := 0.2
	verbose := true

	cliArgs := os.Args[2:]
	for i := 0; i < len(cliArgs); i++ {
		arg := cliArgs[i]
		nextVal := func(prefix string) string {
			if strings.HasPrefix(arg, prefix+"=") {
				return strings.TrimPrefix(arg, prefix+"=")
			}
			if arg == prefix && i+1 < len(cliArgs) {
				i++
				return cliArgs[i]
			}
			return ""
		}
		switch {
		case strings.HasPrefix(arg, "--limit"):
			fmt.Sscanf(nextVal("--limit"), "%d", &limit)
		case arg == "--force":
			force = true
		case strings.HasPrefix(arg, "--concurrency"):
			fmt.Sscanf(nextVal("--concurrency"), "%d", &concurrency)
		case strings.HasPrefix(arg, "--volume"):
			volume = nextVal("--volume")
		case strings.HasPrefix(arg, "--book"):
			book = nextVal("--book")
		case strings.HasPrefix(arg, "--chapter"):
			fmt.Sscanf(nextVal("--chapter"), "%d", &chapter)
		case strings.HasPrefix(arg, "--temperature"):
			fmt.Sscanf(nextVal("--temperature"), "%f", &temperature)
		case arg == "--verbose":
			verbose = true
		case arg == "--quiet":
			verbose = false
		}
	}

	if cfg.ChatModel == "" {
		fmt.Fprintln(os.Stderr, "Error: GOSPEL_ENGINE_CHAT_MODEL must be set")
		os.Exit(1)
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	client := llm.NewClient(cfg.ChatURL, cfg.ChatModel)
	enrich := enricher.New(client, temperature)

	// Build query for unenriched chapters
	query := `SELECT id, volume, book, chapter, title, full_content FROM chapters`
	var conditions []string
	var args []any

	if !force {
		conditions = append(conditions, "enrichment_summary IS NULL")
	}
	if volume != "" {
		conditions = append(conditions, "volume = ?")
		args = append(args, volume)
	}
	if book != "" {
		conditions = append(conditions, "book = ?")
		args = append(args, book)
	}
	if chapter > 0 {
		conditions = append(conditions, "chapter = ?")
		args = append(args, chapter)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY volume, book, chapter"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying chapters: %v\n", err)
		os.Exit(1)
	}

	type chapterRow struct {
		ID      int64
		Volume  string
		Book    string
		Chapter int
		Title   string
		Content string
	}
	var chapters []chapterRow
	for rows.Next() {
		var c chapterRow
		if err := rows.Scan(&c.ID, &c.Volume, &c.Book, &c.Chapter, &c.Title, &c.Content); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning chapter: %v\n", err)
			continue
		}
		chapters = append(chapters, c)
	}
	rows.Close()

	if len(chapters) == 0 {
		fmt.Println("No chapters to enrich.")
		return
	}

	if verbose {
		fmt.Printf("📖 Enriching %d chapters with scripture profiles\n", len(chapters))
		fmt.Printf("   Model: %s (T=%.1f)\n", cfg.ChatModel, temperature)
		fmt.Printf("   Concurrency: %d\n", concurrency)
		fmt.Printf("   Lens: gospel-vocab.md + titsw-framework.md\n\n")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	start := time.Now()
	var enrichedCount, errorCount int64
	var mu sync.Mutex // protects verbose output and edge writes

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i, ch := range chapters {
		select {
		case <-ctx.Done():
			break
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // acquire slot

		go func(idx int, ch chapterRow) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			select {
			case <-ctx.Done():
				return
			default:
			}

			chStart := time.Now()
			ref := fmt.Sprintf("%s %d", ch.Book, ch.Chapter)
			if verbose {
				mu.Lock()
				fmt.Printf("[%d/%d] %s/%s %d... ", idx+1, len(chapters), ch.Volume, ch.Book, ch.Chapter)
				mu.Unlock()
			}

			profile, err := enrich.EnrichScripture(ctx, ch.Content)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
				if verbose {
					mu.Lock()
					fmt.Printf("[%d/%d] ❌ %v\n", idx+1, len(chapters), err)
					mu.Unlock()
				}
				return
			}

			// Update the chapter row with enrichment data
			_, err = database.Exec(`
				UPDATE chapters SET
					enrichment_summary = ?,
					enrichment_keywords = ?,
					enrichment_key_verse = ?,
					enrichment_christ_types = ?,
					enrichment_connections = ?,
					enrichment_model = ?,
					enrichment_raw_output = ?
				WHERE id = ?
			`, profile.Summary, profile.Keywords, profile.KeyVerse,
				profile.ChristTypes, profile.Connections,
				cfg.ChatModel, profile.RawOutput,
				ch.ID)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
				if verbose {
					mu.Lock()
					fmt.Printf("[%d/%d] ❌ DB: %v\n", idx+1, len(chapters), err)
					mu.Unlock()
				}
				return
			}

			// Edge writes need mutex — they do multiple DB operations per chapter
			mu.Lock()
			if profile.ChristTypes != "" && strings.ToLower(profile.ChristTypes) != "none" {
				writeTypologicalEdges(database, ref, ch.Volume, ch.Book, ch.Chapter, profile.ChristTypes)
			}
			if profile.Connections != "" && strings.ToLower(profile.Connections) != "none" {
				writeConnectionEdges(database, ref, profile.Connections)
			}
			mu.Unlock()

			n := atomic.AddInt64(&enrichedCount, 1)
			elapsed := time.Since(chStart).Round(time.Millisecond)
			if verbose {
				kwCount := len(strings.Split(profile.Keywords, ","))
				ctCount := 0
				if profile.ChristTypes != "" && strings.ToLower(profile.ChristTypes) != "none" {
					ctCount = len(strings.Split(profile.ChristTypes, ","))
				}
				mu.Lock()
				fmt.Printf("[%d/%d] ✅ %d kw, %d types %v (total: %d)\n",
					idx+1, len(chapters), kwCount, ctCount, elapsed, n)
				mu.Unlock()
			}
		}(i, ch)
	}

	wg.Wait()

	totalTime := time.Since(start).Round(time.Second)
	finalEnriched := atomic.LoadInt64(&enrichedCount)
	finalErrors := atomic.LoadInt64(&errorCount)
	fmt.Printf("\n✅ Scripture enrichment complete: %d/%d chapters in %v (%d errors)\n",
		finalEnriched, len(chapters), totalTime, finalErrors)
	if finalEnriched > 0 {
		avgTime := totalTime / time.Duration(finalEnriched)
		remaining := int64(len(chapters)) - finalEnriched
		if remaining > 0 && finalErrors > 0 {
			fmt.Printf("   Average: %v/chapter, ~%v for remaining %d\n",
				avgTime, avgTime*time.Duration(remaining), remaining)
		}
	}
}

// writeTypologicalEdges parses Christ-type connections and writes them as edges.
// Format: "symbol → Christ connection (verse), ..."
func writeTypologicalEdges(database *db.DB, sourceRef, volume, book string, chapter int, christTypes string) {
	sourceID := fmt.Sprintf("%s/%s/%d", volume, book, chapter)
	entries := strings.Split(christTypes, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		database.Exec(`
			INSERT INTO edges (source_type, source_id, target_type, target_id, edge_type, metadata)
			VALUES ('scripture', ?, 'christ_type', ?, 'typological', ?)
		`, sourceID, entry, fmt.Sprintf(`{"source_ref":"%s"}`, sourceRef))
	}
}

// writeConnectionEdges parses cross-dispensation connections and writes them as edges.
// Format: "reference — reason\n..."
func writeConnectionEdges(database *db.DB, sourceRef string, connections string) {
	lines := strings.Split(connections, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse "reference — reason" or "reference - reason"
		parts := strings.SplitN(line, "—", 2)
		if len(parts) < 2 {
			parts = strings.SplitN(line, " - ", 2)
		}
		if len(parts) < 2 {
			continue
		}
		targetRef := strings.TrimSpace(parts[0])
		// Strip leading bullets/numbers
		targetRef = strings.TrimLeft(targetRef, "0123456789.-) ")
		reason := strings.TrimSpace(parts[1])

		database.Exec(`
			INSERT INTO edges (source_type, source_id, target_type, target_id, edge_type, metadata)
			VALUES ('scripture', ?, 'scripture', ?, 'thematic', ?)
		`, sourceRef, targetRef, fmt.Sprintf(`{"reason":"%s"}`, strings.ReplaceAll(reason, `"`, `\"`)))
	}
}

func runSearchTest() {
	cfg := config.Default()

	// Parse: search [--mode=keyword|semantic|combined] <query>
	mode := "combined"
	limit := 10
	query := ""
	source := ""

	cliArgs := os.Args[2:]
	for i := 0; i < len(cliArgs); i++ {
		arg := cliArgs[i]
		switch {
		case strings.HasPrefix(arg, "--mode="):
			mode = strings.TrimPrefix(arg, "--mode=")
		case arg == "--mode" && i+1 < len(cliArgs):
			i++
			mode = cliArgs[i]
		case strings.HasPrefix(arg, "--limit="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--limit="), "%d", &limit)
		case arg == "--limit" && i+1 < len(cliArgs):
			i++
			fmt.Sscanf(cliArgs[i], "%d", &limit)
		case strings.HasPrefix(arg, "--source="):
			source = strings.TrimPrefix(arg, "--source=")
		case arg == "--source" && i+1 < len(cliArgs):
			i++
			source = cliArgs[i]
		default:
			if query == "" {
				query = arg
			} else {
				query += " " + arg
			}
		}
	}

	if query == "" {
		fmt.Fprintln(os.Stderr, "Usage: gospel-engine search [--mode=keyword|semantic|combined] [--source=scriptures|conference] <query>")
		os.Exit(1)
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	var vecSearcher vec.Searcher
	if vec.VecFilesExist(cfg.DataDir) {
		embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)
		mmapStore, err := vec.NewMmapStore(cfg.DataDir, cfg.DBPath, embedFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  mmap store unavailable: %v\n", err)
		} else {
			vecSearcher = mmapStore
			defer mmapStore.Close()
		}
	}

	engine := search.NewEngine(database, vecSearcher)

	opts := search.Options{
		Mode:  search.Mode(mode),
		Limit: limit,
	}
	if source != "" {
		opts.Sources = strings.Split(source, ",")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	start := time.Now()
	results, err := engine.Search(ctx, query, opts)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 %s search for %q (%d results in %v)\n\n", mode, query, len(results), elapsed.Round(time.Millisecond))
	for i, r := range results {
		content := r.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		content = strings.ReplaceAll(content, "\n", " ")
		fmt.Printf("[%d] %.4f  %-12s %-8s  %s\n", i+1, r.Score, r.Source, r.Type, r.Reference)
		fmt.Printf("    %s\n", content)
		if r.FilePath != "" {
			fmt.Printf("    → %s\n", r.FilePath)
		}
		fmt.Println()
	}
}

func runConvert() {
	cfg := config.Default()
	dryRun := false
	verbose := true

	for _, arg := range os.Args[2:] {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--quiet":
			verbose = false
		}
	}

	fmt.Println("🔄 Converting gob.gz → mmap (.vecf) format")
	fmt.Printf("   Data dir: %s\n", cfg.DataDir)
	fmt.Printf("   DB: %s\n\n", cfg.DBPath)

	// Load existing gob.gz data into chromem-go (this is the slow part — one last time)
	fmt.Println("Loading gob.gz files into memory (this is the last time)...")
	embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)
	store, err := vec.NewStore(cfg.DataDir, embedFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading vector store: %v\n", err)
		os.Exit(1)
	}

	if dryRun {
		fmt.Println(vec.ConvertStats(store))
		return
	}

	// Open database for writing metadata
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	opts := vec.ConvertOptions{
		Verbose: verbose,
		OnProgress: func(p vec.ConvertProgress) {
			if verbose {
				fmt.Printf("    %s: %d/%d written\n", p.Collection, p.Written, p.Total)
			}
		},
	}

	if err := vec.ConvertToMmap(store, database.DB, cfg.DataDir, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Conversion error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🚀 Conversion complete! The serve command will now use mmap for instant startup.")
}

func runStats() {
	cfg := config.Default()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	stats, err := database.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Gospel Engine Stats\n")
	fmt.Printf("  Database: %s\n\n", cfg.DBPath)
	fmt.Printf("  Scriptures (verses): %d\n", stats.Scriptures)
	fmt.Printf("  Chapters:            %d\n", stats.Chapters)
	fmt.Printf("  Conference Talks:    %d\n", stats.Talks)
	fmt.Printf("  Manuals:             %d\n", stats.Manuals)
	fmt.Printf("  Books:               %d\n", stats.Books)
	fmt.Printf("  Cross References:    %d\n", stats.CrossRefs)
	fmt.Printf("  Graph Edges:         %d\n", stats.Edges)

	// Try vector store stats
	embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)
	store, err := vec.NewStore(cfg.DataDir, embedFunc)
	if err == nil {
		fmt.Printf("\n  Vector Collections:\n")
		for name, count := range store.Stats() {
			fmt.Printf("    %s: %d\n", name, count)
		}
	}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

func runEmbedEnrichments() {
	cfg := config.Default()

	// Parse flags
	source := "" // "scriptures", "conference", or "" for both
	batchSize := 50
	verbose := true

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "--source="):
			source = strings.TrimPrefix(arg, "--source=")
		case arg == "--source" && i+1 < len(os.Args):
			i++
			source = os.Args[i]
		case strings.HasPrefix(arg, "--batch="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--batch="), "%d", &batchSize)
		case arg == "--quiet":
			verbose = false
		}
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	embedFunc := vec.NewLMStudioEmbedder(cfg.EmbeddingURL, cfg.EmbeddingModel)
	store, err := vec.NewStore(cfg.DataDir, embedFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening vector store: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	start := time.Now()
	totalChunks := 0

	// --- Scripture enrichment summaries ---
	if source == "" || source == "scriptures" {
		rows, err := database.Query(`
			SELECT volume, book, chapter, file_path, enrichment_model,
				enrichment_summary, enrichment_keywords, enrichment_christ_types
			FROM chapters
			WHERE enrichment_summary IS NOT NULL AND enrichment_summary != ''
		`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying enriched chapters: %v\n", err)
			os.Exit(1)
		}

		var chunks []vec.Chunk
		count := 0
		for rows.Next() {
			var vol, book, filePath, model string
			var ch int
			var summary, keywords, christTypes string
			if err := rows.Scan(&vol, &book, &ch, &filePath, &model, &summary, &keywords, &christTypes); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning chapter: %v\n", err)
				continue
			}

			bookName := indexer.FormatBookName(book)
			ref := fmt.Sprintf("%s %d", bookName, ch)

			// Summary chunk — the main enrichment text
			content := fmt.Sprintf("%s: %s\nKeywords: %s", ref, summary, keywords)
			if christTypes != "" && strings.ToLower(christTypes) != "none" {
				content += fmt.Sprintf("\nChrist types: %s", christTypes)
			}

			chunks = append(chunks, vec.Chunk{
				ID:      fmt.Sprintf("enrichment-scripture-%s-%s-%d", vol, book, ch),
				Content: content,
				Metadata: &vec.DocMetadata{
					Source:    vec.SourceScriptures,
					Layer:     vec.LayerSummary,
					Book:      bookName,
					Chapter:   ch,
					Reference: ref,
					FilePath:  filePath,
					Generated: true,
					Model:     model,
				},
			})
			count++

			// Batch embed
			if len(chunks) >= batchSize {
				if verbose {
					fmt.Printf("📖 Embedding scripture enrichments batch (%d chunks)...\n", len(chunks))
				}
				if err := store.AddChunks(ctx, chunks); err != nil {
					fmt.Fprintf(os.Stderr, "Error embedding scripture batch: %v\n", err)
				}
				totalChunks += len(chunks)
				chunks = chunks[:0]
			}
		}
		rows.Close()

		// Final batch
		if len(chunks) > 0 {
			if verbose {
				fmt.Printf("📖 Embedding scripture enrichments final batch (%d chunks)...\n", len(chunks))
			}
			if err := store.AddChunks(ctx, chunks); err != nil {
				fmt.Fprintf(os.Stderr, "Error embedding scripture batch: %v\n", err)
			}
			totalChunks += len(chunks)
		}
		if verbose {
			fmt.Printf("   ✅ %d scripture enrichment chunks\n", count)
		}
	}

	// --- Talk TITSW summaries ---
	if source == "" || source == "conference" {
		rows, err := database.Query(`
			SELECT year, month, speaker, title, file_path, titsw_model,
				titsw_summary, titsw_keywords, titsw_mode, titsw_dominant
			FROM talks
			WHERE titsw_summary IS NOT NULL AND titsw_summary != ''
		`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying enriched talks: %v\n", err)
			os.Exit(1)
		}

		var chunks []vec.Chunk
		count := 0
		for rows.Next() {
			var yr, mo int
			var speaker, title, filePath, model string
			var summary, keywords, mode, dominant string
			if err := rows.Scan(&yr, &mo, &speaker, &title, &filePath, &model,
				&summary, &keywords, &mode, &dominant); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning talk: %v\n", err)
				continue
			}

			ref := fmt.Sprintf("%s, \"%s\" (%d/%02d)", speaker, title, yr, mo)

			// Summary chunk — TITSW-enriched talk summary
			content := fmt.Sprintf("%s: %s\nKeywords: %s\nMode: %s\nDominant: %s",
				ref, summary, keywords, mode, dominant)

			chunks = append(chunks, vec.Chunk{
				ID:      fmt.Sprintf("enrichment-talk-%d-%02d-%s", yr, mo, slugify(title)),
				Content: content,
				Metadata: &vec.DocMetadata{
					Source:    vec.SourceConference,
					Layer:     vec.LayerSummary,
					Book:      speaker,
					Reference: ref,
					FilePath:  filePath,
					Generated: true,
					Model:     model,
				},
			})
			count++

			if len(chunks) >= batchSize {
				if verbose {
					fmt.Printf("🎤 Embedding talk enrichments batch (%d chunks)...\n", len(chunks))
				}
				if err := store.AddChunks(ctx, chunks); err != nil {
					fmt.Fprintf(os.Stderr, "Error embedding talk batch: %v\n", err)
				}
				totalChunks += len(chunks)
				chunks = chunks[:0]
			}
		}
		rows.Close()

		if len(chunks) > 0 {
			if verbose {
				fmt.Printf("🎤 Embedding talk enrichments final batch (%d chunks)...\n", len(chunks))
			}
			if err := store.AddChunks(ctx, chunks); err != nil {
				fmt.Fprintf(os.Stderr, "Error embedding talk batch: %v\n", err)
			}
			totalChunks += len(chunks)
		}
		if verbose {
			fmt.Printf("   ✅ %d talk enrichment chunks\n", count)
		}
	}

	// Save vector store
	if verbose {
		fmt.Printf("\n💾 Saving vector store...\n")
	}
	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving vector store: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(start).Round(time.Second)
	fmt.Printf("✅ Embedded %d enrichment chunks in %v\n", totalChunks, elapsed)
}
