package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
)

// CachedClient wraps the API client with caching functionality.
type CachedClient struct {
	client *api.Client
	cache  *Cache
}

// NewCachedClient creates a new cached API client.
func NewCachedClient(client *api.Client, cache *Cache) *CachedClient {
	return &CachedClient{
		client: client,
		cache:  cache,
	}
}

// GetCollection fetches a collection, using cache if available.
func (c *CachedClient) GetCollection(ctx context.Context, uri string) (*api.Collection, bool, error) {
	// Check cache first
	if data, err := c.cache.Get(uri, "dynamic"); err == nil {
		var resp api.CollectionResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			return &resp.Collection, true, nil
		}
	}

	// Fetch from API
	rawJSON, err := c.client.GetRawCollection(ctx, uri)
	if err != nil {
		return nil, false, err
	}

	// Cache the response
	if err := c.cache.Put(uri, "dynamic", rawJSON); err != nil {
		// Log but don't fail on cache write errors
		fmt.Printf("Warning: failed to cache %s: %v\n", uri, err)
	}

	// Parse and return
	var resp api.CollectionResponse
	if err := json.Unmarshal(rawJSON, &resp); err != nil {
		return nil, false, fmt.Errorf("unmarshal collection: %w", err)
	}

	return &resp.Collection, false, nil
}

// GetDynamic fetches a dynamic page, using cache if available.
func (c *CachedClient) GetDynamic(ctx context.Context, uri string) (*api.DynamicResponse, bool, error) {
	// Check cache first
	if data, err := c.cache.Get(uri, "dynamic"); err == nil {
		var resp api.DynamicResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			return &resp, true, nil
		}
	}

	// Fetch from API
	rawJSON, err := c.client.GetRawDynamic(ctx, uri)
	if err != nil {
		return nil, false, err
	}

	// Cache the response
	if err := c.cache.Put(uri, "dynamic", rawJSON); err != nil {
		fmt.Printf("Warning: failed to cache %s: %v\n", uri, err)
	}

	// Parse and return
	var resp api.DynamicResponse
	if err := json.Unmarshal(rawJSON, &resp); err != nil {
		return nil, false, fmt.Errorf("unmarshal dynamic: %w", err)
	}

	return &resp, false, nil
}

// GetContent fetches content, using cache if available.
func (c *CachedClient) GetContent(ctx context.Context, uri string) (*api.ContentResponse, bool, error) {
	// Check cache first
	if data, err := c.cache.Get(uri, "content"); err == nil {
		var resp api.ContentResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			return &resp, true, nil
		}
	}

	// Fetch from API
	rawJSON, err := c.client.GetRawContent(ctx, uri)
	if err != nil {
		return nil, false, err
	}

	// Cache the response
	if err := c.cache.Put(uri, "content", rawJSON); err != nil {
		fmt.Printf("Warning: failed to cache %s: %v\n", uri, err)
	}

	// Parse and return
	var resp api.ContentResponse
	if err := json.Unmarshal(rawJSON, &resp); err != nil {
		return nil, false, fmt.Errorf("unmarshal content: %w", err)
	}

	return &resp, false, nil
}

// IsCached checks if content is cached.
func (c *CachedClient) IsCached(uri, endpoint string) bool {
	return c.cache.Has(uri, endpoint)
}

// InvalidateCache removes cached content for a URI.
func (c *CachedClient) InvalidateCache(uri, endpoint string) error {
	return c.cache.Delete(uri, endpoint)
}

// CacheStats returns cache statistics.
func (c *CachedClient) CacheStats() (*Stats, error) {
	return c.cache.GetStats()
}
