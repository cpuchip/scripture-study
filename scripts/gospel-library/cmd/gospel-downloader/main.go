// Gospel Library Downloader
// Downloads General Conference talks and other resources from the Church of Jesus Christ
// of Latter-day Saints Gospel Library for local AI-assisted study.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/convert"
	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/tui"
)

var (
	langFlag        = flag.String("lang", "eng", "Language code (eng, spa, por, etc.)")
	cacheFlag       = flag.String("cache", ".gospel-cache", "Cache directory for raw API responses")
	outputFlag      = flag.String("output", "gospel-library", "Output directory for converted markdown")
	syncFlag        = flag.Bool("sync", false, "Sync catalog only (non-interactive)")
	convertFlag     = flag.Bool("convert", false, "Convert cached content to markdown")
	reconvertFlag   = flag.Bool("reconvert", false, "Reconvert all cached content to markdown")
	testFlag        = flag.Bool("test", false, "Test API with a sample request")
	testCacheFlag   = flag.Bool("test-cache", false, "Test API with caching")
	testConvertFlag = flag.Bool("test-convert", false, "Test HTML to Markdown conversion")
	testCrawlFlag   = flag.Bool("test-crawl", false, "Debug crawl to see what API returns")
	cleanupFlag     = flag.Bool("cleanup", false, "Clear the cache directory")
	resetFlag       = flag.Bool("reset", false, "Clear both cache and output directories")
	standardFlag    = flag.Bool("standard", false, "Download standard works and latest conference")
	downloadFlag    = flag.String("download", "", "Download content from a specific URI path (e.g., /manual/general-handbook)")
)

func main() {
	flag.Parse()

	var err error
	switch {
	case *resetFlag:
		err = doReset(*cacheFlag, *outputFlag)
	case *cleanupFlag:
		err = doCleanup(*cacheFlag)
	case *downloadFlag != "":
		err = downloadSinglePath(*langFlag, *cacheFlag, *outputFlag, *downloadFlag)
	case *standardFlag:
		err = downloadStandard(*langFlag, *cacheFlag, *outputFlag)
	case *testFlag:
		err = testAPI(*langFlag)
	case *testCacheFlag:
		err = testCachedAPI(*langFlag, *cacheFlag)
	case *testConvertFlag:
		err = testConvert(*langFlag, *cacheFlag)
	case *testCrawlFlag:
		err = testCrawl(*langFlag, *cacheFlag)
	case *reconvertFlag:
		err = reconvertCache(*langFlag, *cacheFlag, *outputFlag)
	case *syncFlag:
		// TODO: Implement sync
		fmt.Println("Sync not yet implemented")
	case *convertFlag:
		// TODO: Implement convert
		fmt.Println("Convert not yet implemented")
	default:
		err = runTUI(*langFlag, *cacheFlag, *outputFlag)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func doCleanup(cacheDir string) error {
	fmt.Printf("Clearing cache directory: %s\n", cacheDir)
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to remove cache: %w", err)
	}
	fmt.Println("✓ Cache cleared")
	return nil
}

func doReset(cacheDir, outputDir string) error {
	fmt.Printf("Clearing cache directory: %s\n", cacheDir)
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to remove cache: %w", err)
	}
	fmt.Println("✓ Cache cleared")

	fmt.Printf("Clearing output directory: %s\n", outputDir)
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("failed to remove output: %w", err)
	}
	fmt.Println("✓ Output cleared")
	return nil
}

func downloadSinglePath(lang, cacheDir, outputDir, uriPath string) error {
	fmt.Printf("Downloading %s...\n", uriPath)
	fmt.Println("")

	rawClient := api.NewClient(lang)
	fileCache := cache.New(cacheDir, lang)
	cachedClient := cache.NewCachedClient(rawClient, fileCache)

	downloader := tui.NewDownloader(cachedClient, rawClient, lang, outputDir)
	ctx := context.Background()

	fmt.Printf("   Crawling %s...\n", uriPath)

	uris, err := downloader.CrawlForContent(ctx, uriPath)
	if err != nil {
		return fmt.Errorf("error crawling: %w", err)
	}

	fmt.Printf("   Found %d content items\n", len(uris))
	fmt.Printf("   Downloading...\n")

	// Download with progress output
	successCount := 0
	errorCount := 0
	skippedCount := 0
	for i, uri := range uris {
		if i%50 == 0 {
			fmt.Printf("   Progress: %d/%d (success: %d, errors: %d, skipped: %d)\n", i, len(uris), successCount, errorCount, skippedCount)
		}
		result := downloader.DownloadAndConvert(ctx, uri)
		if result.Success {
			successCount++
		} else if result.Error != nil {
			errorCount++
			if errorCount <= 10 {
				fmt.Printf("   ⚠ %s: %v\n", uri, result.Error)
			}
		} else {
			skippedCount++
		}
		// Check context
		if ctx.Err() != nil {
			fmt.Printf("   Context cancelled at %d: %v\n", i, ctx.Err())
			break
		}
		// Small delay to avoid rate limiting
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("   ✓ Downloaded %d/%d items (errors: %d, skipped: %d)\n", successCount, len(uris), errorCount, skippedCount)
	fmt.Println("")

	fmt.Println("✅ Download complete!")
	fmt.Printf("   Output: %s/%s/\n", outputDir, lang)
	return nil
}

func downloadStandard(lang, cacheDir, outputDir string) error {
	fmt.Println("Downloading Standard Works and Latest Conference...")
	fmt.Println("")

	rawClient := api.NewClient(lang)
	fileCache := cache.New(cacheDir, lang)
	cachedClient := cache.NewCachedClient(rawClient, fileCache)

	downloader := tui.NewDownloader(cachedClient, rawClient, lang, outputDir)
	ctx := context.Background()

	// Standard works URIs to crawl
	standardWorks := []struct {
		name string
		uri  string
	}{
		{"Book of Mormon", "/scriptures/bofm"},
		{"Doctrine and Covenants", "/scriptures/dc-testament"},
		{"Pearl of Great Price", "/scriptures/pgp"},
		{"Old Testament", "/scriptures/ot"},
		{"New Testament", "/scriptures/nt"},
		{"October 2025 General Conference", "/general-conference/2025/10"},
		{"General Handbook", "/manual/general-handbook"},
	}

	for _, work := range standardWorks {
		fmt.Printf("📖 %s\n", work.name)
		fmt.Printf("   Crawling %s...\n", work.uri)

		uris, err := downloader.CrawlForContent(ctx, work.uri)
		if err != nil {
			fmt.Printf("   ⚠ Error crawling: %v\n", err)
			continue
		}

		fmt.Printf("   Found %d content items\n", len(uris))
		fmt.Printf("   Downloading...\n")

		results := downloader.DownloadAll(ctx, uris)
		successCount := 0
		for _, r := range results {
			if r.Success {
				successCount++
			} else if r.Error != nil {
				fmt.Printf("   ⚠ %s: %v\n", r.URI, r.Error)
			}
		}
		fmt.Printf("   ✓ Downloaded %d/%d items\n", successCount, len(results))
		fmt.Println("")
	}

	fmt.Println("✅ Standard download complete!")
	fmt.Printf("   Output: %s/%s/\n", outputDir, lang)
	return nil
}

func runTUI(lang, cacheDir, outputDir string) error {
	// Initialize components
	rawClient := api.NewClient(lang)
	fileCache := cache.New(cacheDir, lang)
	cachedClient := cache.NewCachedClient(rawClient, fileCache)

	// Create and run TUI
	model := tui.New(cachedClient, rawClient, fileCache, lang, outputDir)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err := p.Run()
	return err
}

func reconvertCache(lang, cacheDir, outputDir string) error {
	fmt.Println("Re-converting cached content...")
	fmt.Printf("Cache: %s (%s)\n", cacheDir, lang)
	fmt.Printf("Output: %s\n\n", outputDir)

	opts := convert.DefaultOptions()
	opts.OutputDir = outputDir
	opts.Lang = lang
	converter := convert.New(opts)

	langDir := filepath.Join(cacheDir, lang)
	contentExt := ".content.json"
	linkRe := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

	var contentPaths []string
	err := filepath.Walk(langDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, contentExt) {
			return nil
		}
		contentPaths = append(contentPaths, path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("reconvert cache: %w", err)
	}

	if len(contentPaths) == 0 {
		fmt.Println("No cached content found. Download content first.")
		return nil
	}

	total := len(contentPaths)
	success := 0
	failed := 0
	brokenLinks := make(map[string]int)
	start := time.Now()
	lastLineLen := 0

	for i, path := range contentPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			failed++
			continue
		}

		var entry cache.CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			failed++
			continue
		}
		if entry.EndPoint != "content" {
			continue
		}

		var content api.ContentResponse
		if err := json.Unmarshal(entry.RawJSON, &content); err != nil {
			failed++
			continue
		}

		outputPath := buildOutputPath(outputDir, lang, content.URI)

		converted, err := converter.ConvertContent(&content)
		if err != nil {
			failed++
			fmt.Printf("\r%s\r\n", strings.Repeat(" ", lastLineLen))
			fmt.Printf("%d/%d ✗ Convert failed: %s (%v)\n", i+1, total, content.URI, err)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			failed++
			fmt.Printf("\r%s\r\n", strings.Repeat(" ", lastLineLen))
			fmt.Printf("%d/%d ✗ mkdir failed: %s (%v)\n", i+1, total, outputPath, err)
			continue
		}
		if err := os.WriteFile(outputPath, []byte(converted.Markdown), 0644); err != nil {
			failed++
			fmt.Printf("\r%s\r\n", strings.Repeat(" ", lastLineLen))
			fmt.Printf("%d/%d ✗ write failed: %s (%v)\n", i+1, total, outputPath, err)
			continue
		}

		success++
		for _, link := range linkRe.FindAllStringSubmatch(converted.Markdown, -1) {
			if len(link) != 2 {
				continue
			}
			for _, missing := range findMissingLinks(link[1], outputPath, outputDir, lang) {
				brokenLinks[missing]++
			}
		}

		elapsed := time.Since(start)
		rate := float64(i+1) / elapsed.Seconds()
		remaining := float64(total - (i + 1))
		eta := time.Duration(0)
		if rate > 0 {
			eta = time.Duration(remaining/rate) * time.Second
		}
		line := fmt.Sprintf("%d/%d Reconvert: %s | Elapsed: %s | ETA: %s", i+1, total, content.URI, elapsed.Truncate(time.Second), eta.Truncate(time.Second))
		padding := ""
		if lastLineLen > len(line) {
			padding = strings.Repeat(" ", lastLineLen-len(line))
		}
		fmt.Printf("\r%s%s", line, padding)
		lastLineLen = len(line)
	}

	if lastLineLen > 0 {
		fmt.Print("\r" + strings.Repeat(" ", lastLineLen) + "\r")
	}

	fmt.Printf("\n✓ Reconverted %d items (%d failed)\n", success, failed)
	if len(brokenLinks) > 0 {
		fmt.Printf("⚠ Found %d broken link targets (showing up to 20)\n", len(brokenLinks))
		shown := 0
		for target, count := range brokenLinks {
			fmt.Printf("  %s (%d)\n", target, count)
			shown++
			if shown >= 20 {
				break
			}
		}
	} else {
		fmt.Println("✓ No broken local links detected")
	}

	return nil
}

func buildOutputPath(outputDir, lang, uri string) string {
	cleanURI := strings.TrimPrefix(uri, "/")
	filename := filepath.Base(cleanURI) + ".md"
	dir := filepath.Dir(cleanURI)
	return filepath.Join(outputDir, lang, dir, filename)
}

func findMissingLinks(link, outputPath, outputDir, lang string) []string {
	if link == "" || strings.HasPrefix(link, "http") || strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "#") {
		return nil
	}

	link = strings.Split(link, "#")[0]
	if link == "" {
		return nil
	}

	if ext := filepath.Ext(link); ext != "" && ext != ".md" {
		return nil
	}

	var target string
	if strings.HasPrefix(link, "/") {
		target = filepath.Join(outputDir, lang, filepath.FromSlash(strings.TrimPrefix(link, "/")))
	} else {
		target = filepath.Join(filepath.Dir(outputPath), filepath.FromSlash(link))
	}

	if _, err := os.Stat(target); err != nil {
		return []string{filepath.Clean(target)}
	}

	return nil
}

func printUsage() {
	fmt.Println("Gospel Library Downloader v0.1.0")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  (no flags)         Launch interactive TUI")
	fmt.Println("  --test             Test API with a sample request")
	fmt.Println("  --test-cache       Test API with caching (run twice to see cache hit)")
	fmt.Println("  --test-convert     Test HTML to Markdown conversion")
	fmt.Println("  --sync             Sync catalog only")
	fmt.Println("  --convert          Convert cached content to markdown")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --lang=CODE        Language code (default: eng)")
	fmt.Println("  --cache=DIR        Cache directory (default: .gospel-cache)")
	fmt.Println("  --output=DIR       Output directory (default: gospel-library)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Run from project root (e.g., scripture-study/)")
	fmt.Println("  go run ./scripts/gospel-library/cmd/gospel-downloader")
	fmt.Println("")
	fmt.Println("  # Custom output directory")
	fmt.Println("  go run ./scripts/gospel-library/cmd/gospel-downloader --output=./downloads")
	fmt.Println("")
	fmt.Println("Run without flags to browse and download content interactively.")
}

func testCachedAPI(lang, cacheDir string) error {
	fmt.Println("Testing Gospel Library API with caching...")
	fmt.Println("")

	client := api.NewClient(lang)
	fileCache := cache.New(cacheDir, lang)
	cachedClient := cache.NewCachedClient(client, fileCache)
	ctx := context.Background()

	// Test 1: Fetch General Conference collection
	fmt.Println("1. Fetching General Conference collection...")
	collection, fromCache, err := cachedClient.GetCollection(ctx, "/general-conference")
	if err != nil {
		return fmt.Errorf("failed to fetch collection: %w", err)
	}
	cacheStatus := "fetched from API"
	if fromCache {
		cacheStatus = "loaded from cache ⚡"
	}
	fmt.Printf("   ✓ Found collection: %s (%s)\n", collection.Title, cacheStatus)
	fmt.Printf("   ✓ Sections: %d\n", len(collection.Sections))
	fmt.Println("")

	// Test 2: Fetch October 2024 conference
	fmt.Println("2. Fetching October 2024 conference...")
	oct2024, fromCache, err := cachedClient.GetDynamic(ctx, "/general-conference/2024/10")
	if err != nil {
		return fmt.Errorf("failed to fetch Oct 2024: %w", err)
	}
	cacheStatus = "fetched from API"
	if fromCache {
		cacheStatus = "loaded from cache ⚡"
	}
	if oct2024.TOC != nil {
		fmt.Printf("   ✓ Found: %s (%s)\n", oct2024.TOC.Title, cacheStatus)
	}

	// Find a talk
	var talkURI string
	if oct2024.TOC != nil {
		for _, entry := range oct2024.TOC.Entries {
			if entry.Section != nil && len(entry.Section.Entries) > 0 {
				for _, talkEntry := range entry.Section.Entries {
					if talkEntry.Content != nil && talkEntry.Content.URI != "" {
						if talkEntry.Content.URI != entry.Section.URI {
							talkURI = talkEntry.Content.URI
							break
						}
					}
				}
				if talkURI != "" {
					break
				}
			}
		}
	}
	fmt.Println("")

	// Test 3: Fetch talk content
	if talkURI != "" {
		fmt.Printf("3. Fetching talk: %s\n", talkURI)
		content, fromCache, err := cachedClient.GetContent(ctx, talkURI)
		if err != nil {
			return fmt.Errorf("failed to fetch talk: %w", err)
		}
		cacheStatus = "fetched from API"
		if fromCache {
			cacheStatus = "loaded from cache ⚡"
		}
		fmt.Printf("   ✓ Title: %s (%s)\n", content.Meta.Title, cacheStatus)
		fmt.Printf("   ✓ Body length: %d bytes\n", len(content.Content.Body))
	}
	fmt.Println("")

	// Show cache stats
	fmt.Println("4. Cache statistics:")
	stats, err := cachedClient.CacheStats()
	if err != nil {
		fmt.Printf("   Warning: could not get stats: %v\n", err)
	} else {
		fmt.Printf("   ✓ Cached files: %d\n", stats.TotalFiles)
		fmt.Printf("   ✓ Total size: %.2f KB\n", float64(stats.TotalBytes)/1024)
		if stats.OldestAge > 0 {
			fmt.Printf("   ✓ Oldest entry: %s ago\n", stats.OldestAge.Round(1e9))
		}
	}

	fmt.Println("")
	fmt.Println("✅ Cached API test complete!")
	fmt.Println("")
	fmt.Println("💡 Run again to see cache hits!")
	return nil
}

func testAPI(lang string) error {
	fmt.Println("Testing Gospel Library API...")

	fmt.Println("")

	client := api.NewClient(lang)
	ctx := context.Background()

	// Test 1: Fetch General Conference collection
	fmt.Println("1. Fetching General Conference collection...")
	collection, err := client.GetCollection(ctx, "/general-conference")
	if err != nil {
		return fmt.Errorf("failed to fetch collection: %w", err)
	}
	fmt.Printf("   ✓ Found collection: %s\n", collection.Title)
	fmt.Printf("   ✓ Sections: %d\n", len(collection.Sections))
	if len(collection.Sections) > 0 && len(collection.Sections[0].Entries) > 0 {
		fmt.Printf("   ✓ First entry: %s (%s)\n",
			collection.Sections[0].Entries[0].Title,
			collection.Sections[0].Entries[0].URI)
	}
	fmt.Println("")

	// Test 2: Fetch October 2024 conference (uses dynamic endpoint with TOC structure)
	fmt.Println("2. Fetching October 2024 conference...")
	oct2024, err := client.GetDynamic(ctx, "/general-conference/2024/10")
	if err != nil {
		return fmt.Errorf("failed to fetch Oct 2024: %w", err)
	}

	// Check if we got a TOC response
	if oct2024.TOC != nil {
		fmt.Printf("   ✓ Found: %s\n", oct2024.TOC.Title)
		fmt.Printf("   ✓ Category: %s\n", oct2024.TOC.Category)
		fmt.Printf("   ✓ Entries: %d\n", len(oct2024.TOC.Entries))
	} else if oct2024.Collection != nil {
		fmt.Printf("   ✓ Found collection: %s\n", oct2024.Collection.Title)
	}

	// Find a talk to download from the TOC
	var talkURI string
	var talkTitle string
	if oct2024.TOC != nil {
		for _, entry := range oct2024.TOC.Entries {
			if entry.Section != nil && len(entry.Section.Entries) > 0 {
				fmt.Printf("   ✓ Session: %s (%d talks)\n", entry.Section.Title, len(entry.Section.Entries))
				// Find first actual talk (skip session overview)
				for _, talkEntry := range entry.Section.Entries {
					if talkEntry.Content != nil && talkEntry.Content.URI != "" {
						// Skip session overview pages
						if talkEntry.Content.URI != entry.Section.URI {
							talkURI = talkEntry.Content.URI
							talkTitle = talkEntry.Content.Title
							break
						}
					}
				}
				if talkURI != "" {
					break
				}
			}
		}
	}
	fmt.Println("")

	// Test 3: Fetch actual talk content
	if talkURI != "" {
		fmt.Printf("3. Fetching talk: %s\n", talkTitle)
		fmt.Printf("   URI: %s\n", talkURI)
		content, err := client.GetContent(ctx, talkURI)
		if err != nil {
			return fmt.Errorf("failed to fetch talk: %w", err)
		}
		fmt.Printf("   ✓ Title: %s\n", content.Meta.Title)
		fmt.Printf("   ✓ Content type: %s\n", content.Meta.ContentType)
		fmt.Printf("   ✓ Body length: %d bytes\n", len(content.Content.Body))
		fmt.Printf("   ✓ Footnotes: %d\n", len(content.Content.Footnotes))
		audioItems := content.Meta.GetAudioItems()
		if len(audioItems) > 0 {
			fmt.Printf("   ✓ Audio: %s\n", audioItems[0].MediaURL)
		}

		// Pretty print a sample of the response structure
		fmt.Println("")
		fmt.Println("4. Sample meta data:")
		metaJSON, _ := json.MarshalIndent(content.Meta, "   ", "  ")
		fmt.Printf("   %s\n", string(metaJSON)[:min(500, len(metaJSON))])
		if len(metaJSON) > 500 {
			fmt.Println("   ...")
		}
	} else {
		fmt.Println("3. No talk found to test content fetch")
	}

	fmt.Println("")
	fmt.Println("✅ API test complete!")
	return nil
}

func testCrawl(lang, cacheDir string) error {
	fmt.Println("Testing Crawl Logic for Scriptures...")
	fmt.Println("")

	client := api.NewClient(lang)
	fileCache := cache.New(cacheDir, lang)
	cachedClient := cache.NewCachedClient(client, fileCache)
	ctx := context.Background()

	// Test: What does the API return for /scriptures/bofm?
	fmt.Println("1. Fetching /scriptures/bofm structure...")
	fmt.Println("")

	// Try collection endpoint
	collection, _, err := cachedClient.GetCollection(ctx, "/scriptures/bofm")
	if err == nil && collection != nil {
		fmt.Println("   ✓ Got COLLECTION response")
		fmt.Printf("   Title: %s\n", collection.Title)
		fmt.Printf("   Sections: %d\n", len(collection.Sections))
		for i, section := range collection.Sections {
			if i > 2 {
				fmt.Printf("   ... and %d more sections\n", len(collection.Sections)-3)
				break
			}
			fmt.Printf("   Section %d: %s (entries: %d)\n", i+1, section.Title, len(section.Entries))
			for j, entry := range section.Entries {
				if j > 3 {
					fmt.Printf("      ... and %d more entries\n", len(section.Entries)-4)
					break
				}
				fmt.Printf("      Entry: type=%s uri=%s title=%s\n", entry.Type, entry.URI, entry.Title)
			}
		}
	} else {
		fmt.Printf("   Collection endpoint failed: %v\n", err)
	}
	fmt.Println("")

	// Try dynamic endpoint
	fmt.Println("2. Fetching /scriptures/bofm via dynamic endpoint...")
	dynamic, _, err := cachedClient.GetDynamic(ctx, "/scriptures/bofm")
	if err == nil && dynamic != nil {
		fmt.Println("   ✓ Got DYNAMIC response")
		if dynamic.TOC != nil {
			fmt.Printf("   TOC Title: %s\n", dynamic.TOC.Title)
			fmt.Printf("   TOC Entries: %d\n", len(dynamic.TOC.Entries))
			for i, entry := range dynamic.TOC.Entries {
				if i > 3 {
					fmt.Printf("   ... and %d more entries\n", len(dynamic.TOC.Entries)-4)
					break
				}
				fmt.Printf("   Entry %d:\n", i+1)
				if entry.Content != nil {
					fmt.Printf("      Content URI: %s Title: %s\n", entry.Content.URI, entry.Content.Title)
				}
				if entry.Section != nil {
					fmt.Printf("      Section URI: %s Title: %s Entries: %d\n", entry.Section.URI, entry.Section.Title, len(entry.Section.Entries))
					for j, subEntry := range entry.Section.Entries {
						if j > 3 {
							fmt.Printf("         ... and %d more sub-entries\n", len(entry.Section.Entries)-4)
							break
						}
						if subEntry.Content != nil {
							fmt.Printf("         SubEntry Content: URI=%s Title=%s\n", subEntry.Content.URI, subEntry.Content.Title)
						}
						if subEntry.Section != nil {
							fmt.Printf("         SubEntry Section: URI=%s Title=%s\n", subEntry.Section.URI, subEntry.Section.Title)
						}
					}
				}
			}
		}
		if dynamic.Collection != nil {
			fmt.Printf("   Collection Title: %s\n", dynamic.Collection.Title)
			fmt.Printf("   Collection Sections: %d\n", len(dynamic.Collection.Sections))
		}
	} else {
		fmt.Printf("   Dynamic endpoint failed: %v\n", err)
	}
	fmt.Println("")

	// Let's drill down into 1 Nephi specifically
	fmt.Println("3. Fetching /scriptures/bofm/1-ne (1 Nephi)...")
	dynamic1ne, _, err := cachedClient.GetDynamic(ctx, "/scriptures/bofm/1-ne")
	if err == nil && dynamic1ne != nil {
		if dynamic1ne.TOC != nil {
			fmt.Println("   ✓ Got TOC")
			fmt.Printf("   Title: %s\n", dynamic1ne.TOC.Title)
			fmt.Printf("   Entries: %d\n", len(dynamic1ne.TOC.Entries))
			for i, entry := range dynamic1ne.TOC.Entries {
				if i > 5 {
					fmt.Printf("   ... and %d more entries\n", len(dynamic1ne.TOC.Entries)-6)
					break
				}
				if entry.Content != nil {
					fmt.Printf("   Entry %d: Content URI=%s Title=%s\n", i+1, entry.Content.URI, entry.Content.Title)
				}
				if entry.Section != nil {
					fmt.Printf("   Entry %d: Section URI=%s Title=%s\n", i+1, entry.Section.URI, entry.Section.Title)
				}
			}
		}
	} else {
		fmt.Printf("   Failed: %v\n", err)
	}

	return nil
}

func testConvert(lang, cacheDir string) error {
	fmt.Println("Testing HTML to Markdown conversion...")
	fmt.Println("")

	client := api.NewClient(lang)
	fileCache := cache.New(cacheDir, lang)
	cachedClient := cache.NewCachedClient(client, fileCache)
	ctx := context.Background()

	// Test 1: General Conference talk
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println("TEST 1: General Conference Talk")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("1. Fetching October 2024 conference...")
	oct2024, _, err := cachedClient.GetDynamic(ctx, "/general-conference/2024/10")
	if err != nil {
		return fmt.Errorf("failed to fetch Oct 2024: %w", err)
	}

	// Find a talk
	var talkURI string
	if oct2024.TOC != nil {
		for _, entry := range oct2024.TOC.Entries {
			if entry.Section != nil && len(entry.Section.Entries) > 0 {
				for _, talkEntry := range entry.Section.Entries {
					if talkEntry.Content != nil && talkEntry.Content.URI != "" {
						if talkEntry.Content.URI != entry.Section.URI {
							talkURI = talkEntry.Content.URI
							break
						}
					}
				}
				if talkURI != "" {
					break
				}
			}
		}
	}

	if talkURI == "" {
		return fmt.Errorf("no talk found to convert")
	}

	fmt.Printf("2. Fetching talk: %s\n", talkURI)
	content, fromCache, err := cachedClient.GetContent(ctx, talkURI)
	if err != nil {
		return fmt.Errorf("failed to fetch talk: %w", err)
	}
	cacheStatus := "from API"
	if fromCache {
		cacheStatus = "from cache ⚡"
	}
	fmt.Printf("   ✓ Loaded: %s (%s)\n", content.Meta.Title, cacheStatus)
	fmt.Println("")

	// Convert to markdown
	fmt.Println("3. Converting to Markdown...")
	converter := convert.New(convert.DefaultOptions())
	result, err := converter.ConvertContent(content)
	if err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	fmt.Printf("   ✓ Title: %s\n", result.Title)
	fmt.Printf("   ✓ Markdown length: %d bytes\n", len(result.Markdown))
	if result.AudioURL != "" {
		fmt.Printf("   ✓ Audio URL: %s\n", result.AudioURL)
	}
	fmt.Println("")

	// Show a preview
	fmt.Println("4. Markdown preview (first 1500 chars):")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	preview := result.Markdown
	if len(preview) > 1500 {
		preview = preview[:1500] + "\n\n... [truncated]"
	}
	fmt.Println(preview)
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")

	// Test 2: Scripture with footnotes
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println("TEST 2: Scripture with Footnotes (1 Nephi 3)")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println("")

	fmt.Println("1. Fetching 1 Nephi 3...")
	scripture, fromCache, err := cachedClient.GetContent(ctx, "/scriptures/bofm/1-ne/3")
	if err != nil {
		return fmt.Errorf("failed to fetch scripture: %w", err)
	}
	cacheStatus = "from API"
	if fromCache {
		cacheStatus = "from cache ⚡"
	}
	fmt.Printf("   ✓ Loaded: %s (%s)\n", scripture.Meta.Title, cacheStatus)
	fmt.Printf("   ✓ Footnotes: %d\n", len(scripture.Content.Footnotes))
	fmt.Println("")

	// Convert scripture
	fmt.Println("2. Converting to Markdown...")
	scriptureResult, err := converter.ConvertContent(scripture)
	if err != nil {
		return fmt.Errorf("failed to convert scripture: %w", err)
	}

	fmt.Printf("   ✓ Title: %s\n", scriptureResult.Title)
	fmt.Printf("   ✓ Markdown length: %d bytes\n", len(scriptureResult.Markdown))
	fmt.Println("")

	// Show scripture preview with footnotes (focus on end to see footnotes)
	fmt.Println("3. Markdown preview (last 2500 chars to see footnotes):")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	preview = scriptureResult.Markdown
	if len(preview) > 2500 {
		preview = "... [truncated]\n\n" + preview[len(preview)-2500:]
	}
	fmt.Println(preview)
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	fmt.Println("")
	fmt.Println("✅ Conversion test complete!")

	return nil
}
