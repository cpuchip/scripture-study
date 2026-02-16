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

	// Music content pages
	if strings.HasPrefix(uri, "/music/") {
		segments := strings.Split(strings.TrimPrefix(uri, "/"), "/")
		// /music/{collection} = TOC/collection (2 segments)
		// /music/{collection}/{song} = Content (3+ segments)
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

// stripFragment removes the fragment identifier (e.g., #title_number19) from a URI.
// The API returns the same content regardless of fragment, so we deduplicate by base URI.
func stripFragment(uri string) string {
	if idx := strings.Index(uri, "#"); idx != -1 {
		return uri[:idx]
	}
	return uri
}

// CrawlForContentWithProgress crawls and reports progress via callback.
// The callback receives (currentURI, visitedCount, discoveredCount).
func (d *Downloader) CrawlForContentWithProgress(ctx context.Context, uri string, onProgress func(string, int, int)) ([]string, int, error) {
	// Use a map to deduplicate URIs (especially after stripping fragments)
	uriSet := make(map[string]bool)
	visited := make(map[string]bool)
	visitedCount := 0
	discoveredCount := 0

	// addURI adds a URI to the set after stripping fragment identifiers.
	// Returns true if this was a new URI.
	addURI := func(rawURI string) bool {
		cleanURI := stripFragment(rawURI)
		if !isContentURI(cleanURI) {
			return false
		}
		if uriSet[cleanURI] {
			return false
		}
		uriSet[cleanURI] = true
		discoveredCount++
		if onProgress != nil {
			onProgress(cleanURI, visitedCount, discoveredCount)
		}
		return true
	}

	// processTOCEntries recursively processes TOC entries at any depth.
	// This handles deeply nested sections like the General Handbook has.
	var processTOCEntries func(entries []api.TOCEntry)
	processTOCEntries = func(entries []api.TOCEntry) {
		for _, entry := range entries {
			// Handle direct content references
			if entry.Content != nil && entry.Content.URI != "" {
				addURI(entry.Content.URI)
			}
			// Handle sections with nested entries (recursive)
			if entry.Section != nil && len(entry.Section.Entries) > 0 {
				processTOCEntries(entry.Section.Entries)
			}
		}
	}

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
						if !addURI(entry.URI) {
							// Not content or already seen - maybe a sub-collection
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
				addURI(u)

				// Some manuals are TOCs served as content HTML; follow their links
				links := extractManualLinks(content.Content.Body)
				for _, link := range links {
					if !addURI(link) && !visited[link] {
						if err := crawl(link); err != nil {
							continue
						}
					}
				}
				return nil
			}

			addURI(u)
			return nil
		}

		if dynamic.TOC != nil {
			// Use the recursive processor to handle arbitrarily nested TOC entries
			// This properly handles deeply nested structures like the General Handbook
			processTOCEntries(dynamic.TOC.Entries)
		} else if dynamic.Collection != nil && len(dynamic.Collection.Sections) > 0 {
			for _, section := range dynamic.Collection.Sections {
				for _, entry := range section.Entries {
					if entry.Type == "item" {
						if !addURI(entry.URI) {
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

	// Convert set to slice
	allURIs := make([]string, 0, len(uriSet))
	for u := range uriSet {
		allURIs = append(allURIs, u)
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
// For music URIs, automatically downloads media files (PDF, MP3) alongside lyrics.
func (d *Downloader) DownloadSingle(ctx context.Context, uri string) tea.Cmd {
	return func() tea.Msg {
		if isMusicURI(uri) {
			mResult := d.DownloadMusicContent(ctx, uri)
			return downloadResultMsg{result: mResult.DownloadResult}
		}
		result := d.DownloadAndConvert(ctx, uri)
		return downloadResultMsg{result: result}
	}
}

// isMusicURI returns true if the URI is a music content page.
func isMusicURI(uri string) bool {
	return strings.HasPrefix(uri, "/music/") ||
		strings.HasPrefix(uri, "/manual/hymns/") ||
		strings.HasPrefix(uri, "/manual/childrens-songbook/")
}

// MusicDownloadResult extends DownloadResult with media file counts.
type MusicDownloadResult struct {
	DownloadResult
	PDFDownloaded  bool
	MP3sDownloaded int
	MediaErrors    []string
}

// DownloadMusicContent downloads a music content page with its media files (PDF, MP3s).
// It first downloads the lyrics markdown, then downloads any associated PDF and MP3 files.
func (d *Downloader) DownloadMusicContent(ctx context.Context, uri string) MusicDownloadResult {
	mResult := MusicDownloadResult{}

	// First do the standard download (lyrics markdown)
	result := d.DownloadAndConvert(ctx, uri)
	mResult.DownloadResult = result
	if !result.Success {
		return mResult
	}

	// Fetch the content to get media URLs (may already be cached)
	content, _, err := d.client.GetContent(ctx, uri)
	if err != nil {
		// Already have the markdown, media is bonus
		return mResult
	}

	// Determine base path (same directory as the markdown, without .md extension)
	basePath := strings.TrimSuffix(result.FilePath, ".md")

	// Download PDF sheet music
	pdfItems := content.Meta.GetPDFItems()
	if len(pdfItems) > 0 && pdfItems[0].Source != "" && strings.HasPrefix(pdfItems[0].Source, "http") {
		pdfPath := basePath + ".pdf"
		if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
			if err := d.rawClient.DownloadFile(ctx, pdfItems[0].Source, pdfPath); err != nil {
				mResult.MediaErrors = append(mResult.MediaErrors, fmt.Sprintf("PDF: %v", err))
			} else {
				mResult.PDFDownloaded = true
			}
		} else {
			mResult.PDFDownloaded = true // Already exists
		}
	}

	// Download MP3 audio files
	audioItems := content.Meta.GetAudioItems()
	for _, audio := range audioItems {
		if audio.MediaURL == "" || !strings.HasPrefix(audio.MediaURL, "http") {
			continue // Skip empty or non-URL entries (some are UUID hashes)
		}
		suffix := variantToSuffix(audio.Variant)
		mp3Path := basePath + suffix + ".mp3"
		if _, err := os.Stat(mp3Path); os.IsNotExist(err) {
			if err := d.rawClient.DownloadFile(ctx, audio.MediaURL, mp3Path); err != nil {
				mResult.MediaErrors = append(mResult.MediaErrors, fmt.Sprintf("MP3 %s: %v", audio.Variant, err))
			} else {
				mResult.MP3sDownloaded++
			}
		} else {
			mResult.MP3sDownloaded++ // Already exists
		}
	}

	// Prepend local media links to the markdown file
	if mResult.PDFDownloaded || mResult.MP3sDownloaded > 0 {
		d.prependMediaLinks(result.FilePath, basePath, pdfItems, audioItems)
	}

	return mResult
}

// variantToSuffix converts an audio variant string to a file suffix.
func variantToSuffix(variant string) string {
	variant = strings.ToLower(variant)
	variant = strings.TrimPrefix(variant, "audio_")
	if variant == "" {
		return "_audio"
	}
	return "_" + variant
}

// prependMediaLinks adds a "Media Files" section to the top of a markdown file
// with relative links to downloaded media.
func (d *Downloader) prependMediaLinks(mdPath, basePath string, pdfItems []api.PDFItem, audioItems []api.MediaItem) {
	existing, err := os.ReadFile(mdPath)
	if err != nil {
		return
	}

	base := filepath.Base(basePath)
	var mediaSection strings.Builder
	mediaSection.WriteString("\n## Media Files\n\n")

	if len(pdfItems) > 0 && pdfItems[0].Source != "" && strings.HasPrefix(pdfItems[0].Source, "http") {
		mediaSection.WriteString(fmt.Sprintf("- 📄 [Sheet Music (PDF)](%s.pdf)\n", base))
	}
	for _, audio := range audioItems {
		if audio.MediaURL == "" || !strings.HasPrefix(audio.MediaURL, "http") {
			continue
		}
		suffix := variantToSuffix(audio.Variant)
		label := strings.ReplaceAll(strings.TrimPrefix(suffix, "_"), "_", " ")
		// Simple title case: capitalize first letter of each word
		words := strings.Fields(label)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		label = strings.Join(words, " ")
		mediaSection.WriteString(fmt.Sprintf("- 🎵 [%s](%s%s.mp3)\n", label, base, suffix))
	}
	mediaSection.WriteString("\n")

	// Insert after the title line (# Title\n\n)
	content := string(existing)
	if idx := strings.Index(content, "\n\n"); idx != -1 {
		content = content[:idx+2] + mediaSection.String() + content[idx+2:]
	} else {
		content = content + "\n" + mediaSection.String()
	}

	os.WriteFile(mdPath, []byte(content), 0644)
}
