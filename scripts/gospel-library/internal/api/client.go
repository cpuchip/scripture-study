package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
)

const (
	// BaseURL is the Gospel Library API base URL.
	BaseURL = "https://www.churchofjesuschrist.org/study/api/v3/language-pages"

	// DefaultRateLimit is the default requests per second.
	DefaultRateLimit = 20

	// DefaultTimeout is the default HTTP timeout.
	DefaultTimeout = 30 * time.Second

	// UserAgent identifies our tool.
	UserAgent = "ScriptureStudy-Downloader/1.0 (personal study tool; github.com/cpuchip/scripture-study)"
)

// Client is a rate-limited HTTP client for the Gospel Library API.
type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	lang       string
}

// NewClient creates a new API client with the specified language.
func NewClient(lang string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		limiter: rate.NewLimiter(rate.Limit(DefaultRateLimit), 1),
		lang:    lang,
	}
}

// NewClientWithRateLimit creates a new API client with a custom rate limit.
func NewClientWithRateLimit(lang string, requestsPerSecond float64) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
		lang:    lang,
	}
}

// doRequest performs a rate-limited HTTP GET request.
func (c *Client) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}

// GetCollection fetches a collection/navigation page.
// uri should be like "/general-conference" or "/general-conference/2024/10"
func (c *Client) GetCollection(ctx context.Context, uri string) (*Collection, error) {
	endpoint := fmt.Sprintf("%s/type/dynamic?lang=%s&uri=%s",
		BaseURL, c.lang, url.QueryEscape(uri))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp CollectionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal collection: %w", err)
	}

	return &resp.Collection, nil
}

// GetDynamic fetches a dynamic page which may be a collection or content.
// This handles the varying response structure from the API.
func (c *Client) GetDynamic(ctx context.Context, uri string) (*DynamicResponse, error) {
	endpoint := fmt.Sprintf("%s/type/dynamic?lang=%s&uri=%s",
		BaseURL, c.lang, url.QueryEscape(uri))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp DynamicResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal dynamic: %w", err)
	}

	return &resp, nil
}

// GetContent fetches the actual content of an item (talk, chapter, etc.).
// uri should be like "/general-conference/2024/10/57nelson"
func (c *Client) GetContent(ctx context.Context, uri string) (*ContentResponse, error) {
	endpoint := fmt.Sprintf("%s/type/content?lang=%s&uri=%s",
		BaseURL, c.lang, url.QueryEscape(uri))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp ContentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal content: %w", err)
	}

	return &resp, nil
}

// GetRawContent fetches raw JSON for caching purposes.
func (c *Client) GetRawContent(ctx context.Context, uri string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/type/content?lang=%s&uri=%s",
		BaseURL, c.lang, url.QueryEscape(uri))

	return c.doRequest(ctx, endpoint)
}

// GetRawCollection fetches raw JSON for caching purposes.
func (c *Client) GetRawCollection(ctx context.Context, uri string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/type/dynamic?lang=%s&uri=%s",
		BaseURL, c.lang, url.QueryEscape(uri))

	return c.doRequest(ctx, endpoint)
}

// GetRawDynamic fetches raw JSON for dynamic endpoints (for caching purposes).
func (c *Client) GetRawDynamic(ctx context.Context, uri string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/type/dynamic?lang=%s&uri=%s",
		BaseURL, c.lang, url.QueryEscape(uri))

	return c.doRequest(ctx, endpoint)
}

// DownloadFile downloads a file from a URL and streams it to destPath.
// Uses a longer timeout suitable for media files (MP3, PDF).
func (c *Client) DownloadFile(ctx context.Context, fileURL, destPath string) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	// Use a client with longer timeout for file downloads
	fileClient := &http.Client{Timeout: 5 * time.Minute}
	resp, err := fileClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, fileURL)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(destPath) // Clean up partial file
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
