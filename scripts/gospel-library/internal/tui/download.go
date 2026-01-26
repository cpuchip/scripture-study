package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/convert"
)

// DownloadResult represents the result of downloading a single item.
type DownloadResult struct {
	URI      string
	Title    string
	Success  bool
	Error    error
	FilePath string
}

// downloadResultMsg is sent when a download completes.
type downloadResultMsg struct {
	result DownloadResult
}

// downloadCompleteMsg is sent when all downloads are complete.
type downloadCompleteMsg struct {
	results []DownloadResult
}

// Downloader handles downloading and converting content.
type Downloader struct {
	client    *cache.CachedClient
	rawClient *api.Client
	converter *convert.Converter
	outputDir string
	lang      string
}

// NewDownloader creates a new downloader.
func NewDownloader(client *cache.CachedClient, rawClient *api.Client, lang, outputDir string) *Downloader {
	return &Downloader{
		client:    client,
		rawClient: rawClient,
		converter: convert.New(convert.DefaultOptions()),
		outputDir: outputDir,
		lang:      lang,
	}
}

// isContentURI checks if a URI points to actual content (not a TOC/index page).
// Content URIs typically have a chapter number or specific page identifier.
func isContentURI(uri string) bool {
	// Skip known TOC patterns
	tocPatterns := []string{
		"/_contents",
		"/title-page",   // Keep this as content
		"/introduction", // Keep this as content
		"/bofm-title",   // Keep this as content
	}

	for _, pattern := range tocPatterns {
		if strings.HasSuffix(uri, pattern) {
			// These are actually content pages we want to keep
			if pattern == "/title-page" || pattern == "/introduction" || pattern == "/bofm-title" {
				return true
			}
			return false
		}
	}

	// Scripture chapters have a number at the end: /scriptures/bofm/1-ne/1
	// Conference talks have an identifier: /general-conference/2025/10/12stevenson
	parts := strings.Split(uri, "/")
	if len(parts) == 0 {
		return false
	}

	lastPart := parts[len(parts)-1]

	// Skip pure section URIs like /scriptures/bofm or /scriptures/bofm/1-ne
	// These are TOC pages, not content
	if uri == "/scriptures/bofm" || uri == "/scriptures/ot" || uri == "/scriptures/nt" ||
		uri == "/scriptures/dc-testament" || uri == "/scriptures/pgp" {
		return false
	}

	// Scripture book TOCs (like /scriptures/bofm/1-ne without chapter number)
	if strings.HasPrefix(uri, "/scriptures/") {
		// Topical Guide, Bible Dictionary, Guide to the Scriptures entries
		if strings.HasPrefix(uri, "/scriptures/tg/") ||
			strings.HasPrefix(uri, "/scriptures/bd/") ||
			strings.HasPrefix(uri, "/scriptures/gs/") ||
			strings.HasPrefix(uri, "/scriptures/triple-index/") {
			return true
		}

		// Count path depth - content has more segments
		// /scriptures/bofm = TOC
		// /scriptures/bofm/1-ne = TOC (book)
		// /scriptures/bofm/1-ne/1 = Content (chapter)
		// /scriptures/dc-testament/dc/1 = Content
		segments := strings.Split(strings.TrimPrefix(uri, "/"), "/")

		// D&C is special: /scriptures/dc-testament/dc/1
		if strings.Contains(uri, "dc-testament/dc/") && len(segments) >= 4 {
			return true
		}

		// Other scriptures need 4+ segments for content
		// scriptures/bofm/1-ne/1 = 4 segments
		if len(segments) >= 4 {
			return true
		}

		// Introduction and other auxiliary pages
		if strings.HasSuffix(uri, "/introduction") ||
			strings.HasSuffix(uri, "/title-page") ||
			strings.HasSuffix(uri, "/bofm-title") ||
			strings.HasSuffix(uri, "/three") ||
			strings.HasSuffix(uri, "/eight") ||
			strings.HasSuffix(uri, "/js") ||
			strings.HasSuffix(uri, "/explanation") {
			return true
		}

		return false
	}

	// Manual TOCs: /manual/{slug} are index pages; content is deeper
	if strings.HasPrefix(uri, "/manual/") {
		segments := strings.Split(strings.TrimPrefix(uri, "/"), "/")
		// /manual/{slug} -> 2 segments (not content)
		// /manual/{slug}/{lesson} -> content
		return len(segments) >= 3
	}

	// Conference talks - check if last part is not just a year or month
	if strings.HasPrefix(uri, "/general-conference/") {
		// /general-conference/2025/10 = TOC
		// /general-conference/2025/10/12stevenson = Talk
		segments := strings.Split(strings.TrimPrefix(uri, "/"), "/")
		if len(segments) >= 4 {
			// Last segment should be a talk identifier, not a session name
			if !strings.Contains(lastPart, "-session") {
				return true
			}
		}
		return false
	}

	// Default: if it looks like it has content identifiers, include it
	return len(lastPart) > 0
}

// extractManualLinks pulls manual URIs from content HTML.
func extractManualLinks(html string) []string {
	if html == "" {
		return nil
	}

	linkRe := regexp.MustCompile(`href="([^"]+)"`)
	matches := linkRe.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil
	}

	var links []string
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		href := match[1]
		if strings.HasPrefix(href, "https://www.churchofjesuschrist.org/study/") {
			href = strings.TrimPrefix(href, "https://www.churchofjesuschrist.org/study")
		} else if strings.HasPrefix(href, "/study/") {
			href = strings.TrimPrefix(href, "/study")
		}

		if idx := strings.Index(href, "?"); idx != -1 {
			href = href[:idx]
		}
		if href == "" {
			continue
		}
		if !strings.HasPrefix(href, "/") {
			href = "/" + href
		}

		if strings.HasPrefix(href, "/manual/") {
			links = append(links, href)
		}
	}

	return links
}

// CrawlForContent recursively discovers all actual content URIs under a given URI.
// It skips TOC/index pages and only returns URIs that point to actual content.
func (d *Downloader) CrawlForContent(ctx context.Context, uri string) ([]string, error) {
	uris, _, err := d.CrawlForContentWithProgress(ctx, uri, nil)
	return uris, err
}

// CrawlForContentWithProgress crawls and reports progress via callback.
// The callback receives (currentURI, visitedCount, discoveredCount).
func (d *Downloader) CrawlForContentWithProgress(ctx context.Context, uri string, onProgress func(string, int, int)) ([]string, int, error) {
	var allURIs []string
	visited := make(map[string]bool)
	visitedCount := 0
	discoveredCount := 0

	var crawl func(u string) error
	crawl = func(u string) error {
		if visited[u] {
			return nil
		}
		visited[u] = true
		visitedCount++
		if onProgress != nil {
			onProgress(u, visitedCount, discoveredCount)
		}

		// Try collection endpoint first
		collection, _, err := d.client.GetCollection(ctx, u)
		if err == nil && collection != nil && len(collection.Sections) > 0 {
			for _, section := range collection.Sections {
				for _, entry := range section.Entries {
					if entry.Type == "item" {
						if isContentURI(entry.URI) {
							allURIs = append(allURIs, entry.URI)
							discoveredCount++
							if onProgress != nil {
								onProgress(entry.URI, visitedCount, discoveredCount)
							}
						} else {
							// Treat non-content items as collections
							if err := crawl(entry.URI); err != nil {
								continue
							}
						}
					} else {
						// Recurse into sub-collections
						if err := crawl(entry.URI); err != nil {
							continue
						}
					}
				}
			}
			return nil
		}

		// Try dynamic endpoint (this is the main path for scriptures)
		dynamic, _, err := d.client.GetDynamic(ctx, u)
		if err != nil {
			// Not a collection or dynamic page - might be content itself
			content, _, cErr := d.client.GetContent(ctx, u)
			if cErr == nil && content != nil {
				if isContentURI(u) {
					allURIs = append(allURIs, u)
					discoveredCount++
					if onProgress != nil {
						onProgress(u, visitedCount, discoveredCount)
					}
				}

				// Some manuals are TOCs served as content HTML; follow their links
				links := extractManualLinks(content.Content.Body)
				for _, link := range links {
					if isContentURI(link) {
						allURIs = append(allURIs, link)
						discoveredCount++
						if onProgress != nil {
							onProgress(link, visitedCount, discoveredCount)
						}
					} else if !visited[link] {
						if err := crawl(link); err != nil {
							continue
						}
					}
				}
				return nil
			}

			if isContentURI(u) {
				allURIs = append(allURIs, u)
				discoveredCount++
				if onProgress != nil {
					onProgress(u, visitedCount, discoveredCount)
				}
			}
			return nil
		}

		if dynamic.TOC != nil {
			for _, entry := range dynamic.TOC.Entries {
				// Direct content at this level
				if entry.Content != nil && entry.Content.URI != "" {
					if isContentURI(entry.Content.URI) {
						allURIs = append(allURIs, entry.Content.URI)
						discoveredCount++
						if onProgress != nil {
							onProgress(entry.Content.URI, visitedCount, discoveredCount)
						}
					} else {
						if err := crawl(entry.Content.URI); err != nil {
							continue
						}
					}
				}
				// Check sections for sub-content (chapters within books)
				if entry.Section != nil {
					for _, subEntry := range entry.Section.Entries {
						if subEntry.Content != nil && subEntry.Content.URI != "" && isContentURI(subEntry.Content.URI) {
							allURIs = append(allURIs, subEntry.Content.URI)
							discoveredCount++
							if onProgress != nil {
								onProgress(subEntry.Content.URI, visitedCount, discoveredCount)
							}
						}
						// Handle nested sections (if any)
						if subEntry.Section != nil && subEntry.Section.URI != "" && !visited[subEntry.Section.URI] {
							if err := crawl(subEntry.Section.URI); err != nil {
								continue
							}
						}
					}
				}
			}
		} else if dynamic.Collection != nil && len(dynamic.Collection.Sections) > 0 {
			for _, section := range dynamic.Collection.Sections {
				for _, entry := range section.Entries {
					if entry.Type == "item" {
						if isContentURI(entry.URI) {
							allURIs = append(allURIs, entry.URI)
							discoveredCount++
							if onProgress != nil {
								onProgress(entry.URI, visitedCount, discoveredCount)
							}
						} else {
							if err := crawl(entry.URI); err != nil {
								continue
							}
						}
					} else {
						if err := crawl(entry.URI); err != nil {
							continue
						}
					}
				}
			}
		}

		return nil
	}

	if err := crawl(uri); err != nil {
		return nil, visitedCount, err
	}

	return allURIs, visitedCount, nil
}

// DownloadAll downloads multiple URIs synchronously and returns all results.
func (d *Downloader) DownloadAll(ctx context.Context, uris []string) []DownloadResult {
	var results []DownloadResult
	for _, uri := range uris {
		result := d.DownloadAndConvert(ctx, uri)
		results = append(results, result)
	}
	return results
}

// DownloadAndConvert downloads content from a URI and converts it to markdown.
func (d *Downloader) DownloadAndConvert(ctx context.Context, uri string) DownloadResult {
	result := DownloadResult{URI: uri}

	// Fetch content (uses cache if available)
	content, _, err := d.client.GetContent(ctx, uri)
	if err != nil {
		result.Error = fmt.Errorf("fetch: %w", err)
		return result
	}
	result.Title = content.Meta.Title

	// Convert to markdown
	converted, err := d.converter.ConvertContent(content)
	if err != nil {
		result.Error = fmt.Errorf("convert: %w", err)
		return result
	}

	// Determine output path
	// Convert URI like "/general-conference/2024/10/57nelson" to path
	cleanURI := strings.TrimPrefix(uri, "/")
	filename := filepath.Base(cleanURI) + ".md"
	dir := filepath.Dir(cleanURI)
	outputPath := filepath.Join(d.outputDir, d.lang, dir, filename)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		result.Error = fmt.Errorf("mkdir: %w", err)
		return result
	}

	// Write markdown file
	if err := os.WriteFile(outputPath, []byte(converted.Markdown), 0644); err != nil {
		result.Error = fmt.Errorf("write: %w", err)
		return result
	}

	result.Success = true
	result.FilePath = outputPath
	return result
}

// DownloadMultiple downloads multiple URIs and returns results via channel.
func (d *Downloader) DownloadMultiple(ctx context.Context, uris []string) tea.Cmd {
	return func() tea.Msg {
		results := d.DownloadAll(ctx, uris)
		return downloadCompleteMsg{results: results}
	}
}

// DownloadSingle downloads a single URI.
func (d *Downloader) DownloadSingle(ctx context.Context, uri string) tea.Cmd {
	return func() tea.Msg {
		result := d.DownloadAndConvert(ctx, uri)
		return downloadResultMsg{result: result}
	}
}
