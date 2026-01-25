// Package cache provides file-based caching for Gospel Library API responses.
// Cached content is stored as raw JSON in .cache/ directory, organized by URI path.
// This content is copyrighted and should be gitignored.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cache manages local file storage of API responses.
type Cache struct {
	baseDir string
	lang    string
}

// CacheEntry contains cached content with metadata.
type CacheEntry struct {
	URI      string    `json:"uri"`
	Lang     string    `json:"lang"`
	CachedAt time.Time `json:"cached_at"`
	EndPoint string    `json:"endpoint"` // "dynamic" or "content"
	RawJSON  []byte    `json:"raw_json"`
}

// New creates a new cache instance.
// baseDir is typically ".cache" relative to the project root.
func New(baseDir, lang string) *Cache {
	return &Cache{
		baseDir: baseDir,
		lang:    lang,
	}
}

// pathForURI converts a URI to a filesystem path.
// Example: "/general-conference/2024/10/57nelson" -> ".cache/eng/general-conference/2024/10/57nelson"
func (c *Cache) pathForURI(uri, endpoint string) string {
	// Clean the URI and remove leading slash
	cleanURI := strings.TrimPrefix(uri, "/")

	// Add endpoint suffix to distinguish dynamic vs content responses
	filename := fmt.Sprintf("%s.%s.json", filepath.Base(cleanURI), endpoint)
	dir := filepath.Dir(cleanURI)

	return filepath.Join(c.baseDir, c.lang, dir, filename)
}

// hashURI creates a short hash for URIs that might be too long for filenames.
func hashURI(uri string) string {
	h := sha256.Sum256([]byte(uri))
	return hex.EncodeToString(h[:8]) // 16 char hex string
}

// Has checks if content is cached for the given URI.
func (c *Cache) Has(uri, endpoint string) bool {
	path := c.pathForURI(uri, endpoint)
	_, err := os.Stat(path)
	return err == nil
}

// Get retrieves cached content for the given URI.
// Returns the raw JSON bytes or an error if not cached.
func (c *Cache) Get(uri, endpoint string) ([]byte, error) {
	path := c.pathForURI(uri, endpoint)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not cached: %s", uri)
		}
		return nil, fmt.Errorf("read cache: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cache entry: %w", err)
	}

	return entry.RawJSON, nil
}

// Put stores content in the cache.
func (c *Cache) Put(uri, endpoint string, rawJSON []byte) error {
	path := c.pathForURI(uri, endpoint)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	entry := CacheEntry{
		URI:      uri,
		Lang:     c.lang,
		CachedAt: time.Now(),
		EndPoint: endpoint,
		RawJSON:  rawJSON,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}

	return nil
}

// GetAge returns how old the cached content is.
// Returns an error if not cached.
func (c *Cache) GetAge(uri, endpoint string) (time.Duration, error) {
	path := c.pathForURI(uri, endpoint)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("not cached: %s", uri)
		}
		return 0, fmt.Errorf("read cache: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return 0, fmt.Errorf("unmarshal cache entry: %w", err)
	}

	return time.Since(entry.CachedAt), nil
}

// Delete removes cached content for the given URI.
func (c *Cache) Delete(uri, endpoint string) error {
	path := c.pathForURI(uri, endpoint)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete cache: %w", err)
	}
	return nil
}

// Clear removes all cached content for the current language.
func (c *Cache) Clear() error {
	langDir := filepath.Join(c.baseDir, c.lang)
	err := os.RemoveAll(langDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear cache: %w", err)
	}
	return nil
}

// ClearAll removes all cached content for all languages.
func (c *Cache) ClearAll() error {
	err := os.RemoveAll(c.baseDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear all cache: %w", err)
	}
	return nil
}

// List returns all cached URIs for the current language.
func (c *Cache) List() ([]string, error) {
	langDir := filepath.Join(c.baseDir, c.lang)

	var uris []string
	err := filepath.Walk(langDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Read the entry to get the original URI
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return nil // Skip invalid entries
		}

		uris = append(uris, entry.URI)
		return nil
	})

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("list cache: %w", err)
	}

	return uris, nil
}

// Stats returns cache statistics.
type Stats struct {
	TotalFiles int
	TotalBytes int64
	OldestAge  time.Duration
	NewestAge  time.Duration
}

// GetStats returns statistics about the cache.
func (c *Cache) GetStats() (*Stats, error) {
	langDir := filepath.Join(c.baseDir, c.lang)

	stats := &Stats{}
	var oldest, newest time.Time

	err := filepath.Walk(langDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		stats.TotalFiles++
		stats.TotalBytes += info.Size()

		// Read entry to get cached time
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return nil
		}

		if oldest.IsZero() || entry.CachedAt.Before(oldest) {
			oldest = entry.CachedAt
		}
		if newest.IsZero() || entry.CachedAt.After(newest) {
			newest = entry.CachedAt
		}

		return nil
	})

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("get stats: %w", err)
	}

	if !oldest.IsZero() {
		stats.OldestAge = time.Since(oldest)
	}
	if !newest.IsZero() {
		stats.NewestAge = time.Since(newest)
	}

	return stats, nil
}
