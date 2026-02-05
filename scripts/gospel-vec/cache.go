package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptVersion should be incremented when prompts change significantly
// This invalidates cached summaries so they get regenerated with new prompts
const PromptVersion = "v1"

// SummaryCache manages cached LLM-generated summaries
type SummaryCache struct {
	cacheDir string
}

// CachedChapter holds all cached data for a chapter
type CachedChapter struct {
	Book          string          `json:"book"`
	Chapter       int             `json:"chapter"`
	Model         string          `json:"model"`
	PromptVersion string          `json:"prompt_version,omitempty"`
	Summary       *ChapterSummary `json:"summary,omitempty"`
	Themes        []ThemeRange    `json:"themes,omitempty"`
}

// computeModelHash creates a short hash of the model name for cache validation
func computeModelHash(model string) string {
	h := sha256.Sum256([]byte(model))
	return fmt.Sprintf("%x", h[:4]) // First 8 hex chars
}

// NewSummaryCache creates a new cache manager
func NewSummaryCache(cacheDir string) *SummaryCache {
	return &SummaryCache{cacheDir: cacheDir}
}

// cacheKey generates a unique key for a chapter
func (c *SummaryCache) cacheKey(book string, chapter int) string {
	// Normalize book name: "1 Nephi" -> "1-nephi"
	safe := strings.ToLower(book)
	safe = strings.ReplaceAll(safe, " ", "-")
	return fmt.Sprintf("%s-%d", safe, chapter)
}

// cachePath returns the file path for a cached chapter
func (c *SummaryCache) cachePath(book string, chapter int) string {
	key := c.cacheKey(book, chapter)
	return filepath.Join(c.cacheDir, key+".json")
}

// Load retrieves cached data for a chapter (returns nil if not cached)
func (c *SummaryCache) Load(book string, chapter int) (*CachedChapter, error) {
	path := c.cachePath(book, chapter)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil // Not cached
	}
	if err != nil {
		return nil, fmt.Errorf("reading cache: %w", err)
	}

	var cached CachedChapter
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("parsing cache: %w", err)
	}

	return &cached, nil
}

// Save stores cached data for a chapter
func (c *SummaryCache) Save(cached *CachedChapter) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	path := c.cachePath(cached.Book, cached.Chapter)

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing cache: %w", err)
	}

	return nil
}

// GetSummary retrieves a cached summary or returns nil
// Returns nil if model or prompt version doesn't match (cache invalidation)
func (c *SummaryCache) GetSummary(book string, chapter int, model string) *ChapterSummary {
	cached, err := c.Load(book, chapter)
	if err != nil || cached == nil {
		return nil
	}
	// Invalidate if model or prompt version changed
	if cached.Model != model || cached.PromptVersion != PromptVersion {
		return nil
	}
	return cached.Summary
}

// GetThemes retrieves cached themes or returns nil
// Returns nil if model or prompt version doesn't match (cache invalidation)
func (c *SummaryCache) GetThemes(book string, chapter int, model string) []ThemeRange {
	cached, err := c.Load(book, chapter)
	if err != nil || cached == nil {
		return nil
	}
	// Invalidate if model or prompt version changed
	if cached.Model != model || cached.PromptVersion != PromptVersion {
		return nil
	}
	return cached.Themes
}

// SaveSummary stores a summary in the cache
func (c *SummaryCache) SaveSummary(book string, chapter int, model string, summary *ChapterSummary) error {
	// Load existing or create new
	cached, _ := c.Load(book, chapter)
	if cached == nil {
		cached = &CachedChapter{Book: book, Chapter: chapter}
	}
	cached.Model = model
	cached.PromptVersion = PromptVersion
	cached.Summary = summary
	return c.Save(cached)
}

// SaveThemes stores themes in the cache
func (c *SummaryCache) SaveThemes(book string, chapter int, model string, themes []ThemeRange) error {
	// Load existing or create new
	cached, _ := c.Load(book, chapter)
	if cached == nil {
		cached = &CachedChapter{Book: book, Chapter: chapter}
	}
	cached.Model = model
	cached.PromptVersion = PromptVersion
	cached.Themes = themes
	return c.Save(cached)
}

// ListCached returns all cached chapter keys
func (c *SummaryCache) ListCached() ([]string, error) {
	entries, err := os.ReadDir(c.cacheDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			keys = append(keys, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return keys, nil
}

// Clear removes all cached data
func (c *SummaryCache) Clear() error {
	return os.RemoveAll(c.cacheDir)
}
