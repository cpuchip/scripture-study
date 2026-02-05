package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Searcher provides unified search across all layers
type Searcher struct {
	store *Store
}

// NewSearcher creates a new searcher
func NewSearcher(store *Store) *Searcher {
	return &Searcher{store: store}
}

// Search performs a unified search across specified layers
func (s *Searcher) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	results, err := s.store.Search(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// SearchVerses searches only the verse layer
func (s *Searcher) SearchVerses(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return s.Search(ctx, query, SearchOptions{
		Layers: []Layer{LayerVerse},
		Limit:  limit,
	})
}

// SearchParagraphs searches only the paragraph layer
func (s *Searcher) SearchParagraphs(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return s.Search(ctx, query, SearchOptions{
		Layers: []Layer{LayerParagraph},
		Limit:  limit,
	})
}

// SearchSummaries searches only the summary layer
func (s *Searcher) SearchSummaries(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return s.Search(ctx, query, SearchOptions{
		Layers: []Layer{LayerSummary},
		Limit:  limit,
	})
}

// SearchThemes searches only the theme layer
func (s *Searcher) SearchThemes(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return s.Search(ctx, query, SearchOptions{
		Layers: []Layer{LayerTheme},
		Limit:  limit,
	})
}

// FindSimilar finds content similar to a specific reference
func (s *Searcher) FindSimilar(ctx context.Context, reference string, limit int) ([]SearchResult, error) {
	// First, find the content for this reference
	// Search verse layer for exact reference match
	results, err := s.store.Search(ctx, reference, SearchOptions{
		Layers: []Layer{LayerVerse},
		Limit:  1,
	})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("reference not found: %s", reference)
	}

	// Use the content of the found verse to find similar verses
	return s.Search(ctx, results[0].Content, SearchOptions{
		Layers: []Layer{LayerVerse},
		Limit:  limit + 1, // +1 to account for the original
	})
}

// FormatResults formats search results for display
func FormatResults(results []SearchResult, showContent bool, maxContentLen int) string {
	var sb strings.Builder

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. [%.4f] %s (%s)\n",
			i+1, r.Score, r.Metadata.Reference, r.Metadata.Layer))

		if showContent {
			content := r.Content
			if maxContentLen > 0 && len(content) > maxContentLen {
				content = content[:maxContentLen] + "..."
			}
			sb.WriteString(fmt.Sprintf("   %s\n", content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GroupByLayer groups results by their layer
func GroupByLayer(results []SearchResult) map[Layer][]SearchResult {
	grouped := make(map[Layer][]SearchResult)
	for _, r := range results {
		layer := r.Metadata.Layer
		grouped[layer] = append(grouped[layer], r)
	}
	return grouped
}

// GroupByBook groups results by their book
func GroupByBook(results []SearchResult) map[string][]SearchResult {
	grouped := make(map[string][]SearchResult)
	for _, r := range results {
		book := r.Metadata.Book
		grouped[book] = append(grouped[book], r)
	}
	return grouped
}
