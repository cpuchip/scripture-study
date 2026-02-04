// Package ddg provides DuckDuckGo search functionality.
package ddg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DefaultUserAgent for requests.
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Client is a DuckDuckGo search client.
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// SearchResult represents a single search result.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// NewsResult represents a news search result.
type NewsResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Date    string `json:"date,omitempty"`
	Source  string `json:"source,omitempty"`
}

// InstantAnswer represents a DuckDuckGo instant answer.
type InstantAnswer struct {
	AbstractText   string `json:"abstract_text,omitempty"`
	AbstractSource string `json:"abstract_source,omitempty"`
	AbstractURL    string `json:"abstract_url,omitempty"`
	Answer         string `json:"answer,omitempty"`
	AnswerType     string `json:"answer_type,omitempty"`
	Definition     string `json:"definition,omitempty"`
	DefinitionURL  string `json:"definition_url,omitempty"`
}

// NewClient creates a new DuckDuckGo client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		userAgent:  DefaultUserAgent,
	}
}

// WebSearch performs a web search.
func (c *Client) WebSearch(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	queryURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DuckDuckGo returned status %d: %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var results []SearchResult
	doc.Find(".web-result").Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		titleNode := s.Find(".result__a")
		title := strings.TrimSpace(titleNode.Text())
		snippet := strings.TrimSpace(s.Find(".result__snippet").Text())

		// Extract URL from the href attribute
		href, exists := titleNode.Attr("href")
		resultURL := ""
		if exists {
			// DuckDuckGo wraps URLs in a redirect, extract the actual URL
			resultURL = extractURL(href)
		}

		if title != "" || snippet != "" {
			results = append(results, SearchResult{
				Title:   title,
				URL:     resultURL,
				Snippet: snippet,
			})
		}
	})

	return results, nil
}

// NewsSearch performs a news search.
func (c *Client) NewsSearch(ctx context.Context, query string, maxResults int, timeLimitDays int) ([]NewsResult, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	// DuckDuckGo news search uses a different endpoint
	// We use the HTML lite endpoint with df (date filter) parameter
	// d = day, w = week, m = month
	df := ""
	switch {
	case timeLimitDays <= 1:
		df = "d"
	case timeLimitDays <= 7:
		df = "w"
	default:
		df = "m"
	}

	queryURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s&iar=news&df=%s",
		url.QueryEscape(query), df)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DuckDuckGo returned status %d: %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var results []NewsResult
	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		titleNode := s.Find(".result__a")
		title := strings.TrimSpace(titleNode.Text())
		snippet := strings.TrimSpace(s.Find(".result__snippet").Text())

		href, exists := titleNode.Attr("href")
		resultURL := ""
		if exists {
			resultURL = extractURL(href)
		}

		// Try to get date/source from the result extras
		extras := strings.TrimSpace(s.Find(".result__extras").Text())

		if title != "" {
			results = append(results, NewsResult{
				Title:   title,
				URL:     resultURL,
				Snippet: snippet,
				Date:    extras,
			})
		}
	})

	return results, nil
}

// InstantAnswerResponse is the raw response from DDG instant answer API.
type InstantAnswerResponse struct {
	AbstractText   string `json:"AbstractText"`
	AbstractSource string `json:"AbstractSource"`
	AbstractURL    string `json:"AbstractURL"`
	Answer         string `json:"Answer"`
	AnswerType     string `json:"AnswerType"`
	Definition     string `json:"Definition"`
	DefinitionURL  string `json:"DefinitionURL"`
	Heading        string `json:"Heading"`
	Type           string `json:"Type"`
}

// GetInstantAnswer retrieves an instant answer from DuckDuckGo.
func (c *Client) GetInstantAnswer(ctx context.Context, query string) (*InstantAnswer, error) {
	// DuckDuckGo has a free instant answer API
	queryURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_redirect=1&no_html=1",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DuckDuckGo returned status %d: %s", resp.StatusCode, string(body))
	}

	var rawResp InstantAnswerResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Check if we got any useful answer
	if rawResp.AbstractText == "" && rawResp.Answer == "" && rawResp.Definition == "" {
		return nil, nil // No instant answer available
	}

	return &InstantAnswer{
		AbstractText:   rawResp.AbstractText,
		AbstractSource: rawResp.AbstractSource,
		AbstractURL:    rawResp.AbstractURL,
		Answer:         rawResp.Answer,
		AnswerType:     rawResp.AnswerType,
		Definition:     rawResp.Definition,
		DefinitionURL:  rawResp.DefinitionURL,
	}, nil
}

// extractURL extracts the actual URL from DuckDuckGo's redirect URL.
func extractURL(href string) string {
	// DuckDuckGo uses redirect URLs like: /l/?kh=-1&uddg=https%3A%2F%2F...
	if strings.Contains(href, "uddg=") {
		parts := strings.Split(href, "uddg=")
		if len(parts) >= 2 {
			decoded, err := url.QueryUnescape(parts[1])
			if err == nil {
				// Remove any trailing parameters
				if idx := strings.Index(decoded, "&"); idx > 0 {
					decoded = decoded[:idx]
				}
				return decoded
			}
		}
	}
	// If it's already a direct URL
	if strings.HasPrefix(href, "http") {
		return href
	}
	return href
}
