// Package convert provides HTML to Markdown conversion for Gospel Library content.
package convert

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"

	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
)

// Options configures the converter behavior.
type Options struct {
	// IncludeFootnotes adds footnotes at the end of the document.
	IncludeFootnotes bool

	// IncludeAudioLink adds a link to the audio version if available.
	IncludeAudioLink bool

	// IncludePDFLink adds a link to the PDF version if available.
	IncludePDFLink bool

	// StripClassesAndIDs removes HTML class and id attributes.
	StripClassesAndIDs bool

	// FootnoteStyle controls how footnotes are rendered.
	// "inline" - footnotes appear as [^marker] with definitions at the end
	// "contextual" - for scriptures, includes the annotated word (uses context field)
	FootnoteStyle string

	// LocalizeLinks converts Gospel Library URIs to local file paths.
	LocalizeLinks bool

	// OutputDir is the base output directory for calculating relative paths.
	// Required when LocalizeLinks is true.
	OutputDir string

	// Lang is the language code used in paths (e.g., "eng").
	Lang string
}

// DefaultOptions returns sensible defaults for conversion.
func DefaultOptions() Options {
	return Options{
		IncludeFootnotes:   true,
		IncludeAudioLink:   true,
		IncludePDFLink:     false,
		StripClassesAndIDs: true,
		FootnoteStyle:      "contextual",
		LocalizeLinks:      true,
		OutputDir:          "",
		Lang:               "eng",
	}
}

// Converter converts Gospel Library content to Markdown.
type Converter struct {
	opts Options
}

// New creates a new Converter with the given options.
func New(opts Options) *Converter {
	return &Converter{
		opts: opts,
	}
}

// ConvertedContent represents the converted markdown output.
type ConvertedContent struct {
	Title     string
	Markdown  string
	AudioURL  string
	PDFURL    string
	SourceURI string
}

// ConvertContent converts a ContentResponse to markdown.
func (c *Converter) ConvertContent(resp *api.ContentResponse) (*ConvertedContent, error) {
	result := &ConvertedContent{
		Title:     resp.Meta.Title,
		SourceURI: resp.URI,
	}

	// Get audio URL if available
	audioItems := resp.Meta.GetAudioItems()
	if len(audioItems) > 0 {
		result.AudioURL = audioItems[0].MediaURL
	}

	// Get PDF URL if available
	var pdfItems []api.PDFItem
	if len(resp.Meta.PDF) > 0 && string(resp.Meta.PDF) != "null" && string(resp.Meta.PDF) != "{}" {
		// Try parsing as array first
		if err := parseJSON(resp.Meta.PDF, &pdfItems); err != nil || len(pdfItems) == 0 {
			// Try single item
			var item api.PDFItem
			if err := parseJSON(resp.Meta.PDF, &item); err == nil && item.Source != "" {
				pdfItems = []api.PDFItem{item}
			}
		}
	}
	if len(pdfItems) > 0 {
		result.PDFURL = pdfItems[0].Source
	}

	// Build markdown
	var sb strings.Builder

	// Title
	sb.WriteString("# ")
	sb.WriteString(resp.Meta.Title)
	sb.WriteString("\n\n")

	// Media links
	if c.opts.IncludeAudioLink && result.AudioURL != "" {
		sb.WriteString(fmt.Sprintf("🎧 [Listen to Audio](%s)\n\n", result.AudioURL))
	}
	if c.opts.IncludePDFLink && result.PDFURL != "" {
		sb.WriteString(fmt.Sprintf("📄 [Download PDF](%s)\n\n", result.PDFURL))
	}

	// Convert body HTML to markdown
	body := resp.Content.Body
	body = c.preprocessHTML(body)

	markdown, err := htmltomarkdown.ConvertString(body)
	if err != nil {
		return nil, fmt.Errorf("failed to convert body: %w", err)
	}

	// Post-process the markdown
	markdown = c.postprocessMarkdown(markdown, resp.URI)
	sb.WriteString(markdown)

	// Add footnotes if requested and available
	if c.opts.IncludeFootnotes && len(resp.Content.Footnotes) > 0 {
		footnoteSection := c.buildFootnotes(resp.Content.Footnotes, resp.URI)
		if footnoteSection != "" {
			sb.WriteString("\n\n---\n\n## Footnotes\n\n")
			sb.WriteString(footnoteSection)
		}
	}

	result.Markdown = sb.String()
	return result, nil
}

// preprocessHTML prepares HTML for conversion by handling Gospel Library specific elements.
func (c *Converter) preprocessHTML(html string) string {
	// Handle scripture verse markers BEFORE stripping classes
	// <span class="verse-number">1 </span> -> ⟦VERSE:1⟧
	// We use a placeholder that won't be escaped, then replace in postprocess
	// Note: Gospel Library includes trailing space inside the span
	verseRe := regexp.MustCompile(`<span[^>]*class="[^"]*verse-number[^"]*"[^>]*>(\d+)\s*</span>`)
	html = verseRe.ReplaceAllString(html, "⟦VERSE:$1⟧ ")

	// Handle paragraph markers (¶ used in scriptures) BEFORE stripping classes
	html = regexp.MustCompile(`<span[^>]*class="[^"]*para-mark[^"]*"[^>]*>¶</span>`).ReplaceAllString(html, "¶ ")

	// Remove data attributes that clutter the output
	if c.opts.StripClassesAndIDs {
		// Remove class attributes
		html = regexp.MustCompile(`\s+class="[^"]*"`).ReplaceAllString(html, "")
		// Remove data-* attributes
		html = regexp.MustCompile(`\s+data-[a-z-]+="[^"]*"`).ReplaceAllString(html, "")
		// Remove id attributes except for footnote notes
		html = regexp.MustCompile(`\s+id="[^"]*"`).ReplaceAllStringFunc(html, func(match string) string {
			// Keep IDs that start with "note" (for footnotes)
			if strings.Contains(match, `id="note`) {
				return match
			}
			return ""
		})
	}

	// Convert footnote references to superscript links
	// Gospel Library uses <a class="note-ref" href="#note1_a">a</a> style
	// We convert to: <sup>[1a](#fn-1a)</sup>
	noteRefRe := regexp.MustCompile(`<a[^>]*href="#(note[^\"]*)"[^>]*>[^<]*</a>`)
	html = noteRefRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := noteRefRe.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		marker := normalizeFootnoteMarker(parts[1])
		if marker == "" {
			return match
		}
		return fmt.Sprintf(`<sup>[%s](#fn-%s)</sup>`, marker, marker)
	})

	// Remove empty paragraphs
	html = regexp.MustCompile(`<p[^>]*>\s*</p>`).ReplaceAllString(html, "")

	// Remove footnote section from body (we build our own from the structured data)
	// The API includes footnotes both in the HTML body and as structured data
	html = regexp.MustCompile(`<footer[^>]*>[\s\S]*?</footer>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<div[^>]*footnotes[^>]*>[\s\S]*?</div>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<ul[^>]*footnotes[^>]*>[\s\S]*?</ul>`).ReplaceAllString(html, "")

	// Remove HTML comments (e.g., <!--THE END-->)
	html = regexp.MustCompile(`<!--([\s\S]*?)-->`).ReplaceAllString(html, "")

	return html
}

// postprocessMarkdown cleans up the converted markdown.
func (c *Converter) postprocessMarkdown(markdown, sourceURI string) string {
	// Convert verse number placeholders to bold with period
	// ⟦VERSE:1⟧ -> **1.**
	verseMarkerRe := regexp.MustCompile(`⟦VERSE:(\d+)⟧`)
	markdown = verseMarkerRe.ReplaceAllString(markdown, "**$1.**")

	// Normalize multiple newlines to at most two
	markdown = regexp.MustCompile(`\n{3,}`).ReplaceAllString(markdown, "\n\n")

	// Remove HTML comments that may survive conversion (e.g., <!--THE END-->)
	markdown = regexp.MustCompile(`<!--([\s\S]*?)-->`).ReplaceAllString(markdown, "")

	// Remove leading/trailing whitespace
	markdown = strings.TrimSpace(markdown)

	// Fix any double-escaped characters
	markdown = strings.ReplaceAll(markdown, `\\[`, `[`)
	markdown = strings.ReplaceAll(markdown, `\\]`, `]`)

	// Normalize footnote references that survived conversion
	// [word](#note1_a) -> word<sup>[1a](#fn-1a)</sup>
	inlineNoteRe := regexp.MustCompile(`\[([^\]]+)\]\(#note([^)]+)\)`)
	markdown = inlineNoteRe.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := inlineNoteRe.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		word := parts[1]
		marker := normalizeFootnoteMarker(parts[2])
		if marker == "" {
			return match
		}
		return fmt.Sprintf("%s<sup>[%s](#fn-%s)</sup>", word, marker, marker)
	})

	// [](#note1) -> <sup>[1](#fn-1)</sup>
	bareNoteRe := regexp.MustCompile(`\[\]\(#note([^)]+)\)`)
	markdown = bareNoteRe.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := bareNoteRe.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		marker := normalizeFootnoteMarker(parts[1])
		if marker == "" {
			return match
		}
		return fmt.Sprintf(`<sup>[%s](#fn-%s)</sup>`, marker, marker)
	})

	// [word] [1a] -> word<sup>[1a](#fn-1a)</sup>
	refNoteRe := regexp.MustCompile(`\[([^\]]+)\]\s*\[(\d+[a-z]?)\]`)
	markdown = refNoteRe.ReplaceAllString(markdown, `$1<sup>[$2](#fn-$2)</sup>`)

	// Localize links to Gospel Library content
	if c.opts.LocalizeLinks {
		markdown = c.localizeLinks(markdown, sourceURI)
	}

	return markdown
}

// normalizeFootnoteMarker normalizes footnote markers like "note1_a" -> "1a".
func normalizeFootnoteMarker(raw string) string {
	if raw == "" {
		return ""
	}
	raw = strings.TrimPrefix(raw, "note")
	raw = strings.TrimPrefix(raw, "#note")
	raw = strings.ReplaceAll(raw, "_", "")
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ".")
	return raw
}

func compareFootnoteMarkers(a, b string) bool {
	an, as := splitFootnoteMarker(normalizeFootnoteMarker(a))
	bn, bs := splitFootnoteMarker(normalizeFootnoteMarker(b))
	if an != bn {
		return an < bn
	}
	if as != bs {
		return as < bs
	}
	return a < b
}

func splitFootnoteMarker(marker string) (int, string) {
	marker = strings.TrimSpace(marker)
	if marker == "" {
		return 0, ""
	}
	var numStr strings.Builder
	var suffix strings.Builder
	for i, r := range marker {
		if r >= '0' && r <= '9' && suffix.Len() == 0 {
			numStr.WriteRune(r)
			continue
		}
		if i == 0 {
			suffix.WriteRune(r)
			continue
		}
		suffix.WriteRune(r)
	}

	num := 0
	if numStr.Len() > 0 {
		fmt.Sscanf(numStr.String(), "%d", &num)
	}
	return num, strings.ToLower(suffix.String())
}

// localizeLinks converts Gospel Library URLs and URIs to local file paths.
func (c *Converter) localizeLinks(markdown, sourceURI string) string {
	// Pattern for markdown links: [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	markdown = linkRe.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := linkRe.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		text := parts[1]
		url := parts[2]

		// Convert Gospel Library URLs to local paths
		localPath := c.convertToLocalPath(url, sourceURI)
		if localPath != "" {
			return fmt.Sprintf("[%s](%s)", text, localPath)
		}
		return match
	})

	// Handle complex/nested link text that the regex above can miss
	// Example: [Title ![img](...)](/study/manual/slug/01?lang=eng)
	urlRe := regexp.MustCompile(`\((https://www\.churchofjesuschrist\.org/study/[^)\s]+|/study/[^)\s]+)\)`)
	markdown = urlRe.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := urlRe.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		url := parts[1]
		localPath := c.convertToLocalPath(url, sourceURI)
		if localPath != "" {
			return fmt.Sprintf("(%s)", localPath)
		}
		return match
	})

	return markdown
}

// convertToLocalPath converts a Gospel Library URL or URI to a local relative path.
func (c *Converter) convertToLocalPath(url, sourceURI string) string {
	// Handle full URLs (https://www.churchofjesuschrist.org/study/...)
	if strings.HasPrefix(url, "https://www.churchofjesuschrist.org/study/") {
		url = strings.TrimPrefix(url, "https://www.churchofjesuschrist.org/study")
	} else if strings.HasPrefix(url, "/study/") {
		url = strings.TrimPrefix(url, "/study")
	} else if !strings.HasPrefix(url, "/") {
		// Not a Gospel Library path, leave unchanged
		return ""
	}

	// Remove query parameters (like ?lang=eng)
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Normalize known manual slugs
	url = c.normalizeManualSlug(url)

	// Remove leading slash
	url = strings.TrimPrefix(url, "/")

	// Skip external URLs, anchors, etc.
	if url == "" || strings.HasPrefix(url, "http") || strings.HasPrefix(url, "#") {
		return ""
	}

	// Build local path: {url}.md
	localPath := url + ".md"

	// Calculate relative path from source file to target
	if sourceURI != "" {
		sourceDir := strings.TrimPrefix(sourceURI, "/")
		if idx := strings.LastIndex(sourceDir, "/"); idx != -1 {
			sourceDir = sourceDir[:idx]
		} else {
			sourceDir = ""
		}

		// Calculate relative path
		localPath = c.relativePath(sourceDir, url+".md")
	}

	return localPath
}

func (c *Converter) normalizeManualSlug(url string) string {
	if strings.HasPrefix(url, "/manual/teaching-in-the-saviors-way") {
		return strings.Replace(url, "/manual/teaching-in-the-saviors-way", "/manual/teaching-in-the-saviors-way-2022", 1)
	}
	return url
}

// relativePath calculates a relative path from source directory to target file.
func (c *Converter) relativePath(sourceDir, targetPath string) string {
	if sourceDir == "" {
		return targetPath
	}

	sourceParts := strings.Split(sourceDir, "/")
	targetParts := strings.Split(targetPath, "/")

	// Find common prefix
	commonLen := 0
	for i := 0; i < len(sourceParts) && i < len(targetParts)-1; i++ {
		if sourceParts[i] == targetParts[i] {
			commonLen = i + 1
		} else {
			break
		}
	}

	// Build relative path
	var parts []string

	// Add ".." for each remaining source directory
	for i := commonLen; i < len(sourceParts); i++ {
		parts = append(parts, "..")
	}

	// Add remaining target path
	for i := commonLen; i < len(targetParts); i++ {
		parts = append(parts, targetParts[i])
	}

	if len(parts) == 0 {
		return targetPath
	}

	return strings.Join(parts, "/")
}

// buildFootnotes creates the footnote section.
func (c *Converter) buildFootnotes(footnotes map[string]api.Footnote, sourceURI string) string {
	if len(footnotes) == 0 {
		return ""
	}

	// Sort footnotes by marker for consistent ordering
	var sorted []api.Footnote
	for _, fn := range footnotes {
		sorted = append(sorted, fn)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return compareFootnoteMarkers(sorted[i].Marker, sorted[j].Marker)
	})

	var sb strings.Builder
	for _, fn := range sorted {
		// Use contextual format for scriptures (includes the annotated word)
		marker := normalizeFootnoteMarker(fn.Marker)
		if marker == "" {
			marker = fn.Marker
		}
		// Use HTML anchor for the footnote target
		if c.opts.FootnoteStyle == "contextual" && fn.Context != "" {
			sb.WriteString(fmt.Sprintf(`<a id="fn-%s"></a>**%s. %s** — `, marker, marker, fn.Context))
		} else {
			sb.WriteString(fmt.Sprintf(`<a id="fn-%s"></a>**%s.** `, marker, marker))
		}

		// Convert footnote HTML to plain text/markdown
		fnText := c.convertFootnoteText(fn, sourceURI)
		sb.WriteString(fnText)
		sb.WriteString("\n\n")
	}

	return strings.TrimSpace(sb.String())
}

// convertFootnoteText converts footnote HTML to readable markdown.
func (c *Converter) convertFootnoteText(fn api.Footnote, sourceURI string) string {
	text := fn.Text

	// Convert any HTML in the footnote text
	if strings.Contains(text, "<") {
		converted, err := htmltomarkdown.ConvertString(text)
		if err == nil {
			text = converted
		}
	}

	// Clean up the text
	text = strings.TrimSpace(text)

	// Localize links in footnote text (use sourceURI for relative paths)
	if c.opts.LocalizeLinks {
		text = c.localizeLinks(text, sourceURI)
	}

	// Note: We skip adding ReferenceURIs because they're already in the HTML text
	// The API includes reference links in both fn.Text (as HTML) and fn.ReferenceURIs

	return text
}

// parseJSON is a helper to unmarshal JSON.
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
