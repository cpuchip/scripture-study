// Tool implementations for fetch-md-mcp.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	readability "github.com/go-shiori/go-readability"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tsawler/tabula"
	"golang.org/x/net/html"
)

type fetchConfig struct {
	HTTPClient *http.Client
	UserAgent  string
	MaxBytes   int64
}

func registerFetchTools(srv *mcp.Server, cfg *fetchConfig) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "fetch_url",
		Description: "Fetch a web page and return cleaned markdown. " +
			"Uses Mozilla Readability to extract the main article content " +
			"and strips boilerplate (navigation, sidebars, ads). Best for " +
			"docs sites, blog posts, READMEs, Wikipedia. Does NOT render " +
			"JavaScript — JS-heavy SPAs may return sparse content.",
	}, makeFetchURL(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "fetch_urls",
		Description: "Concurrent batch fetch of multiple URLs. Each is " +
			"converted to markdown via the same path as fetch_url. Returns " +
			"per-URL results so a partial failure (one bad URL) does not " +
			"abort the rest.",
	}, makeFetchURLs(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "extract_links",
		Description: "Fetch a page and return all links categorized as " +
			"internal (same host), external (different host), social " +
			"(known social-media domains), download (PDF/ZIP/etc), email " +
			"(mailto:), or phone (tel:). Useful for site mapping or " +
			"deciding which links to follow next.",
	}, makeExtractLinks(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "fetch_url_raw",
		Description: "Fetch a page and return raw HTML with no readability " +
			"extraction or markdown conversion. Use when the agent needs " +
			"page structure (tables, forms, scripts) that readability would " +
			"strip out.",
	}, makeFetchURLRaw(cfg))
}

// ---------------------------------------------------------------------
// fetch_url
// ---------------------------------------------------------------------

type fetchURLInput struct {
	URL      string `json:"url" jsonschema:"URL to fetch"`
	MaxChars int    `json:"max_chars,omitempty" jsonschema:"Truncate markdown to this many characters (0 = no truncation)"`
	JS       bool   `json:"js,omitempty" jsonschema:"Render JavaScript via headless Chromium (slower; needed for SPAs and JS-required pages)"`
	WaitMS   int    `json:"wait_ms,omitempty" jsonschema:"Post-load settle delay in ms when js=true (default 500)"`
}

type fetchURLOutput struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Markdown    string `json:"markdown"`
	WordCount   int    `json:"word_count"`
	Truncated   bool   `json:"truncated,omitempty"`
	FetchedAtMs int64  `json:"fetched_at_ms"`
}

func makeFetchURL(cfg *fetchConfig) func(context.Context, *mcp.CallToolRequest, fetchURLInput) (*mcp.CallToolResult, fetchURLOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in fetchURLInput) (*mcp.CallToolResult, fetchURLOutput, error) {
		if strings.TrimSpace(in.URL) == "" {
			return toolError("url is required"), fetchURLOutput{}, nil
		}
		var (
			bodyStr  string
			finalURL string
			err      error
		)
		if in.JS {
			bodyStr, finalURL, err = fetchURLJS(ctx, cfg, in.URL, in.WaitMS)
		} else {
			b, fu, e := httpGet(ctx, cfg, in.URL)
			// ES.5.s2: a non-HTML document (PDF, docx, …) gets extracted
			// to markdown via tabula instead of mangled through the HTML
			// readability path.
			if e == nil {
				if md, isDoc, derr := extractIfDocument(b, fu); isDoc {
					if derr != nil {
						return toolError("document extraction failed for %s: %v", fu, derr), fetchURLOutput{}, nil
					}
					return nil, buildDocOutput(fu, md, in.MaxChars), nil
				}
			}
			bodyStr, finalURL, err = string(b), fu, e
		}
		if err != nil {
			return toolError("fetch %s (js=%v): %v", in.URL, in.JS, err), fetchURLOutput{}, nil
		}

		parsed, _ := url.Parse(finalURL)
		article, err := readability.FromReader(strings.NewReader(bodyStr), parsed)
		if err != nil {
			return toolError("readability: %v", err), fetchURLOutput{}, nil
		}

		md, err := htmltomarkdown.ConvertString(article.Content)
		if err != nil {
			return toolError("html→markdown: %v", err), fetchURLOutput{}, nil
		}

		truncated := false
		if in.MaxChars > 0 && len(md) > in.MaxChars {
			md = md[:in.MaxChars] + "\n\n[…truncated]"
			truncated = true
		}

		return nil, fetchURLOutput{
			URL:         finalURL,
			Title:       article.Title,
			Markdown:    md,
			WordCount:   countWords(md),
			Truncated:   truncated,
			FetchedAtMs: time.Now().UnixMilli(),
		}, nil
	}
}

// ---------------------------------------------------------------------
// fetch_urls
// ---------------------------------------------------------------------

type fetchURLsInput struct {
	URLs     []string `json:"urls" jsonschema:"URLs to fetch concurrently"`
	MaxChars int      `json:"max_chars,omitempty" jsonschema:"Truncate each markdown body to this many characters (0 = no truncation)"`
	JS       bool     `json:"js,omitempty" jsonschema:"Render JavaScript via headless Chromium for each URL (slower)"`
	WaitMS   int      `json:"wait_ms,omitempty" jsonschema:"Post-load settle delay in ms when js=true (default 500)"`
}

type fetchURLsOutput struct {
	Results []fetchURLsOneResult `json:"results"`
}

type fetchURLsOneResult struct {
	URL       string `json:"url"`
	Title     string `json:"title,omitempty"`
	Markdown  string `json:"markdown,omitempty"`
	WordCount int    `json:"word_count,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Error     string `json:"error,omitempty"`
}

func makeFetchURLs(cfg *fetchConfig) func(context.Context, *mcp.CallToolRequest, fetchURLsInput) (*mcp.CallToolResult, fetchURLsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in fetchURLsInput) (*mcp.CallToolResult, fetchURLsOutput, error) {
		if len(in.URLs) == 0 {
			return toolError("urls is required"), fetchURLsOutput{}, nil
		}
		// Cap concurrency so we do not hammer servers. 4 in flight is
		// plenty for an interactive agent triaging candidates.
		const maxConcurrent = 4
		sem := make(chan struct{}, maxConcurrent)

		results := make([]fetchURLsOneResult, len(in.URLs))
		var wg sync.WaitGroup
		for i, u := range in.URLs {
			wg.Add(1)
			go func(idx int, target string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var (
					bodyStr  string
					finalURL string
					err      error
				)
				if in.JS {
					bodyStr, finalURL, err = fetchURLJS(ctx, cfg, target, in.WaitMS)
				} else {
					b, fu, e := httpGet(ctx, cfg, target)
					// ES.5.s2: extract non-HTML documents via tabula.
					if e == nil {
						if md, isDoc, derr := extractIfDocument(b, fu); isDoc {
							if derr != nil {
								results[idx] = fetchURLsOneResult{URL: fu, Error: fmt.Sprintf("document extraction: %v", derr)}
								return
							}
							d := buildDocOutput(fu, md, in.MaxChars)
							results[idx] = fetchURLsOneResult{
								URL: d.URL, Title: d.Title, Markdown: d.Markdown,
								WordCount: d.WordCount, Truncated: d.Truncated,
							}
							return
						}
					}
					bodyStr, finalURL, err = string(b), fu, e
				}
				if err != nil {
					results[idx] = fetchURLsOneResult{URL: target, Error: err.Error()}
					return
				}
				parsed, _ := url.Parse(finalURL)
				article, err := readability.FromReader(strings.NewReader(bodyStr), parsed)
				if err != nil {
					results[idx] = fetchURLsOneResult{URL: finalURL, Error: fmt.Sprintf("readability: %v", err)}
					return
				}
				md, err := htmltomarkdown.ConvertString(article.Content)
				if err != nil {
					results[idx] = fetchURLsOneResult{URL: finalURL, Title: article.Title, Error: fmt.Sprintf("html→markdown: %v", err)}
					return
				}
				truncated := false
				if in.MaxChars > 0 && len(md) > in.MaxChars {
					md = md[:in.MaxChars] + "\n\n[…truncated]"
					truncated = true
				}
				results[idx] = fetchURLsOneResult{
					URL:       finalURL,
					Title:     article.Title,
					Markdown:  md,
					WordCount: countWords(md),
					Truncated: truncated,
				}
			}(i, u)
		}
		wg.Wait()

		return nil, fetchURLsOutput{Results: results}, nil
	}
}

// ---------------------------------------------------------------------
// extract_links
// ---------------------------------------------------------------------

type extractLinksInput struct {
	URL    string `json:"url" jsonschema:"URL whose links you want to enumerate"`
	JS     bool   `json:"js,omitempty" jsonschema:"Render JavaScript via headless Chromium before extracting links (slower; needed for SPAs)"`
	WaitMS int    `json:"wait_ms,omitempty" jsonschema:"Post-load settle delay in ms when js=true (default 500)"`
}

type extractLinksOutput struct {
	URL       string         `json:"url"`
	Internal  []extractedURL `json:"internal,omitempty"`
	External  []extractedURL `json:"external,omitempty"`
	Social    []extractedURL `json:"social,omitempty"`
	Download  []extractedURL `json:"download,omitempty"`
	Email     []extractedURL `json:"email,omitempty"`
	Phone     []extractedURL `json:"phone,omitempty"`
	TotalSeen int            `json:"total_seen"`
}

type extractedURL struct {
	URL  string `json:"url"`
	Text string `json:"text,omitempty"`
}

var socialDomains = map[string]bool{
	"twitter.com":   true,
	"x.com":         true,
	"linkedin.com":  true,
	"facebook.com":  true,
	"instagram.com": true,
	"youtube.com":   true,
	"youtu.be":      true,
	"github.com":    true,
	"mastodon.social": true,
	"bsky.app":      true,
	"reddit.com":    true,
	"tiktok.com":    true,
}

var downloadExts = []string{
	".pdf", ".zip", ".tar", ".gz", ".tgz", ".bz2", ".7z",
	".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".exe", ".dmg", ".pkg", ".deb", ".rpm",
	".mp3", ".mp4", ".mov", ".avi", ".mkv",
	".iso", ".img",
}

func makeExtractLinks(cfg *fetchConfig) func(context.Context, *mcp.CallToolRequest, extractLinksInput) (*mcp.CallToolResult, extractLinksOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in extractLinksInput) (*mcp.CallToolResult, extractLinksOutput, error) {
		if strings.TrimSpace(in.URL) == "" {
			return toolError("url is required"), extractLinksOutput{}, nil
		}
		var (
			bodyStr  string
			finalURL string
			err      error
		)
		if in.JS {
			bodyStr, finalURL, err = fetchURLJS(ctx, cfg, in.URL, in.WaitMS)
		} else {
			b, fu, e := httpGet(ctx, cfg, in.URL)
			bodyStr, finalURL, err = string(b), fu, e
		}
		if err != nil {
			return toolError("fetch %s (js=%v): %v", in.URL, in.JS, err), extractLinksOutput{}, nil
		}
		parsedFinal, err := url.Parse(finalURL)
		if err != nil {
			return toolError("parse final URL: %v", err), extractLinksOutput{}, nil
		}
		sourceHost := strings.ToLower(parsedFinal.Hostname())

		out := extractLinksOutput{URL: finalURL}
		seen := map[string]bool{}

		walkLinks(strings.NewReader(bodyStr), func(href, text string) {
			if href == "" || seen[href] {
				return
			}
			seen[href] = true
			out.TotalSeen++

			low := strings.ToLower(href)
			ent := extractedURL{URL: href, Text: strings.TrimSpace(text)}

			switch {
			case strings.HasPrefix(low, "mailto:"):
				out.Email = append(out.Email, ent)
				return
			case strings.HasPrefix(low, "tel:"):
				out.Phone = append(out.Phone, ent)
				return
			}

			parsed, err := url.Parse(href)
			if err != nil {
				return
			}
			// Resolve relative links against the source URL.
			abs := parsedFinal.ResolveReference(parsed)
			absStr := abs.String()
			ent.URL = absStr

			for _, ext := range downloadExts {
				if strings.HasSuffix(strings.ToLower(abs.Path), ext) {
					out.Download = append(out.Download, ent)
					return
				}
			}

			host := strings.ToLower(abs.Hostname())
			if host == "" || host == sourceHost {
				out.Internal = append(out.Internal, ent)
				return
			}
			// Strip "www." for social-domain matching.
			matchHost := strings.TrimPrefix(host, "www.")
			if socialDomains[matchHost] {
				out.Social = append(out.Social, ent)
				return
			}
			out.External = append(out.External, ent)
		})

		return nil, out, nil
	}
}

func walkLinks(r io.Reader, visit func(href, text string)) {
	doc, err := html.Parse(r)
	if err != nil {
		return
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href string
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = a.Val
					break
				}
			}
			if href != "" {
				visit(href, textContent(n))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

// ---------------------------------------------------------------------
// fetch_url_raw
// ---------------------------------------------------------------------

type fetchURLRawInput struct {
	URL      string `json:"url" jsonschema:"URL to fetch"`
	MaxChars int    `json:"max_chars,omitempty" jsonschema:"Truncate raw HTML to this many characters (0 = no truncation)"`
	JS       bool   `json:"js,omitempty" jsonschema:"Render JavaScript via headless Chromium and return the post-render HTML"`
	WaitMS   int    `json:"wait_ms,omitempty" jsonschema:"Post-load settle delay in ms when js=true (default 500)"`
}

type fetchURLRawOutput struct {
	URL         string `json:"url"`
	HTML        string `json:"html"`
	StatusCode  int    `json:"status_code"`
	Truncated   bool   `json:"truncated,omitempty"`
	FetchedAtMs int64  `json:"fetched_at_ms"`
}

func makeFetchURLRaw(cfg *fetchConfig) func(context.Context, *mcp.CallToolRequest, fetchURLRawInput) (*mcp.CallToolResult, fetchURLRawOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in fetchURLRawInput) (*mcp.CallToolResult, fetchURLRawOutput, error) {
		if strings.TrimSpace(in.URL) == "" {
			return toolError("url is required"), fetchURLRawOutput{}, nil
		}
		var (
			raw        string
			finalURL   string
			statusCode int
			err        error
		)
		if in.JS {
			raw, finalURL, err = fetchURLJS(ctx, cfg, in.URL, in.WaitMS)
			statusCode = 200 // chromedp doesn't surface HTTP status
		} else {
			b, fu, sc, e := httpGetWithStatus(ctx, cfg, in.URL)
			raw, finalURL, statusCode, err = string(b), fu, sc, e
		}
		if err != nil {
			return toolError("fetch %s (js=%v): %v", in.URL, in.JS, err), fetchURLRawOutput{}, nil
		}
		truncated := false
		if in.MaxChars > 0 && len(raw) > in.MaxChars {
			raw = raw[:in.MaxChars] + "\n<!-- […truncated] -->"
			truncated = true
		}
		return nil, fetchURLRawOutput{
			URL:         finalURL,
			HTML:        raw,
			StatusCode:  statusCode,
			Truncated:   truncated,
			FetchedAtMs: time.Now().UnixMilli(),
		}, nil
	}
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

func httpGet(ctx context.Context, cfg *fetchConfig, target string) ([]byte, string, error) {
	body, finalURL, _, err := httpGetWithStatus(ctx, cfg, target)
	return body, finalURL, err
}

func httpGetWithStatus(ctx context.Context, cfg *fetchConfig, target string) ([]byte, string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, target, 0, err
	}
	req.Header.Set("User-Agent", cfg.UserAgent)
	// Default Accept; some servers serve text/markdown when asked
	// (Cloudflare's "Markdown for Agents"). We do NOT prefer markdown
	// here because the readability path handles HTML uniformly.
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, target, 0, err
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	if resp.StatusCode >= 400 {
		return nil, finalURL, resp.StatusCode, fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	var reader io.Reader = resp.Body
	if cfg.MaxBytes > 0 {
		reader = io.LimitReader(resp.Body, cfg.MaxBytes)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, finalURL, resp.StatusCode, fmt.Errorf("read body: %w", err)
	}
	return body, finalURL, resp.StatusCode, nil
}

func countWords(s string) int {
	return len(strings.Fields(s))
}

// ---------------------------------------------------------------------
// document extraction (ES.5.s2) — PDF / Office / EPUB via tabula
// ---------------------------------------------------------------------

// docExtensions are the non-HTML document types tabula extracts.
var docExtensions = map[string]bool{
	".pdf": true, ".docx": true, ".xlsx": true,
	".pptx": true, ".odt": true, ".epub": true,
}

// detectDocExt returns a tabula file extension when the fetched body is
// a non-HTML document, or "" to use the HTML readability path. PDF is
// detected by magic bytes — robust regardless of URL or content-type
// (the ES.4 run fetched a PDF whose body began with "%PDF"). The
// zip-family formats (docx/xlsx/…) share the same magic bytes, so they
// are detected by URL extension instead.
func detectDocExt(body []byte, fetchURL string) string {
	if bytes.HasPrefix(body, []byte("%PDF")) {
		return ".pdf"
	}
	if u, err := url.Parse(fetchURL); err == nil {
		ext := strings.ToLower(path.Ext(u.Path))
		if docExtensions[ext] {
			return ext
		}
	}
	return ""
}

// extractIfDocument extracts markdown when body is a non-HTML document.
// isDoc=false means "treat as HTML" — the caller falls through to the
// readability path. tabula's API is path-based, so the fetched bytes
// are written to a temp file (named with the detected extension so
// tabula auto-detects the format).
func extractIfDocument(body []byte, fetchURL string) (md string, isDoc bool, err error) {
	ext := detectDocExt(body, fetchURL)
	if ext == "" {
		return "", false, nil
	}
	tmp, err := os.CreateTemp("", "fetchmd-*"+ext)
	if err != nil {
		return "", true, fmt.Errorf("temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return "", true, fmt.Errorf("write temp: %w", err)
	}
	tmp.Close()

	out, _, err := tabula.Open(tmp.Name()).ToMarkdown()
	if err != nil {
		return "", true, fmt.Errorf("tabula extract %s: %w", ext, err)
	}
	return out, true, nil
}

// buildDocOutput assembles a fetchURLOutput from extracted document
// markdown. Title is the URL basename — documents carry no <title>.
func buildDocOutput(finalURL, md string, maxChars int) fetchURLOutput {
	truncated := false
	if maxChars > 0 && len(md) > maxChars {
		md = md[:maxChars] + "\n\n[…truncated]"
		truncated = true
	}
	title := ""
	if u, err := url.Parse(finalURL); err == nil {
		title = path.Base(u.Path)
	}
	return fetchURLOutput{
		URL:         finalURL,
		Title:       title,
		Markdown:    md,
		WordCount:   countWords(md),
		Truncated:   truncated,
		FetchedAtMs: time.Now().UnixMilli(),
	}
}

func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
	}
}
