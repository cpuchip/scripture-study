// Package citations provides the client for the BYU Scripture Citation Index.
//
// The BYU Scripture Citation Index (https://scriptures.byu.edu) tracks which
// General Conference talks, Journal of Discourses entries, and other sources
// cite each verse of scripture. This package provides Go functions to query
// that index and parse the results.
package citations

import (
	"fmt"
	htmlpkg "html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const baseURL = "https://scriptures.byu.edu"

// Client queries the BYU Scripture Citation Index.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new BYU Citation Index client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Citation represents a single citation from the index.
type Citation struct {
	Reference string // e.g. "1989-O:54" or "JD 19:81b"
	Speaker   string // e.g. "Gordon B. Hinckley"
	Title     string // e.g. "An Ensign to the Nations"
	TalkID    string // internal BYU talk ID
	RefID     string // internal BYU reference ID
}

// LookupResult is the response from a citation lookup.
type LookupResult struct {
	Scripture string     // The scripture queried, e.g. "D&C 113:6"
	BookID    int        // BYU book ID used
	Chapter   int        // Chapter number
	Verses    string     // Verse string queried
	Citations []Citation // Citations found
	RawHTML   string     // Raw HTML response for debugging
}

// Lookup queries the BYU Citation Index for citations of a scripture reference.
// The reference should be in standard format: "3 Nephi 21:10", "D&C 113:6", "Isaiah 11:1", etc.
func (c *Client) Lookup(reference string) (*LookupResult, error) {
	ref, err := ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	return c.LookupParsed(ref)
}

// LookupParsed queries using an already-parsed reference.
func (c *Client) LookupParsed(ref *ScriptureRef) (*LookupResult, error) {
	bookID, ok := BookIDs[ref.Book]
	if !ok {
		return nil, fmt.Errorf("unknown book: %q (parsed from input)", ref.Book)
	}

	url := fmt.Sprintf("%s/citation_index/citation_ajax/Any/1830/2026/all/s/f/%d/%d?verses=%s",
		baseURL, bookID, ref.Chapter, ref.Verses)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching citations: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BYU API returned status %d: %s", resp.StatusCode, string(body))
	}

	html := string(body)
	citations := parseHTML(html)

	return &LookupResult{
		Scripture: ref.Display(),
		BookID:    bookID,
		Chapter:   ref.Chapter,
		Verses:    ref.Verses,
		Citations: citations,
		RawHTML:   html,
	}, nil
}

// parseHTML extracts citations from the BYU API HTML response.
// The response contains divs with class="reference" and class="talktitle".
func parseHTML(html string) []Citation {
	var citations []Citation

	// Pattern: getTalk('talkId','refId') ... reference text ... talktitle text
	// The HTML is structured as repeated blocks of citation data.
	//
	// Example block:
	//   <div onmouseover="..." onclick="getTalk('4793','18082')" ...>
	//     <span class="reference ...">1989-O:54, Gordon B. Hinckley</span>
	//     <span class="talktitle ...">An Ensign to the Nations</span>
	//   </div>

	// Extract getTalk calls and their associated text
	talkPattern := regexp.MustCompile(`getTalk\('(\d+)',\s*'(\d+)'\)`)
	refPattern := regexp.MustCompile(`class="reference[^"]*"[^>]*>([^<]+)<`)
	titlePattern := regexp.MustCompile(`class="talktitle[^"]*"[^>]*>([^<]+)<`)

	talkMatches := talkPattern.FindAllStringSubmatch(html, -1)
	refMatches := refPattern.FindAllStringSubmatch(html, -1)
	titleMatches := titlePattern.FindAllStringSubmatch(html, -1)

	// All three should have the same count — one per citation
	count := len(talkMatches)
	if len(refMatches) < count {
		count = len(refMatches)
	}
	if len(titleMatches) < count {
		count = len(titleMatches)
	}

	for i := 0; i < count; i++ {
		refText := htmlpkg.UnescapeString(strings.TrimSpace(refMatches[i][1]))
		speaker, reference := parseRefText(refText)

		citations = append(citations, Citation{
			Reference: reference,
			Speaker:   speaker,
			Title:     htmlpkg.UnescapeString(strings.TrimSpace(titleMatches[i][1])),
			TalkID:    talkMatches[i][1],
			RefID:     talkMatches[i][2],
		})
	}

	return citations
}

// parseRefText splits "1989-O:54, Gordon B. Hinckley" into speaker + reference parts.
func parseRefText(text string) (speaker, reference string) {
	// Format: "YEAR-SEASON:PAGE, Speaker Name" or "JD VOL:PAGE, Speaker Name"
	parts := strings.SplitN(text, ", ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[0])
	}
	return text, ""
}

// FormatResult returns a human-readable text representation of the lookup result.
func FormatResult(result *LookupResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## BYU Scripture Citation Index: %s\n\n", result.Scripture))

	if len(result.Citations) == 0 {
		sb.WriteString("**No citations found.** This verse has not been directly cited in General Conference talks, Journal of Discourses, or other indexed sources.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("**%d citation(s) found:**\n\n", len(result.Citations)))

	for i, cite := range result.Citations {
		sb.WriteString(fmt.Sprintf("%d. **%s** — %s", i+1, cite.Speaker, cite.Title))
		if cite.Reference != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", cite.Reference))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatResultCompact returns a brief summary.
func FormatResultCompact(result *LookupResult) string {
	if len(result.Citations) == 0 {
		return fmt.Sprintf("%s: No citations found", result.Scripture)
	}
	return fmt.Sprintf("%s: %d citation(s)", result.Scripture, len(result.Citations))
}
