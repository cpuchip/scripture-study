// Package search: combined.go implements hybrid keyword+semantic search
// using Reciprocal Rank Fusion (RRF).
//
// Instead of trying to normalize incompatible score scales (FTS5 rank vs
// cosine similarity), RRF uses rank positions from each retriever. Documents
// that rank high in both keyword and semantic results get the highest
// combined score. Documents appearing in only one list still contribute.
//
// RRF_score(d) = 1/(k + rank_keyword(d)) + 1/(k + rank_semantic(d))
//
// where k is a smoothing constant (default 60, the standard in IR literature).
package search

import (
	"context"
	"sort"
	"strings"
)

const (
	// rrfK is the smoothing constant for Reciprocal Rank Fusion.
	// 60 is the standard value from the original Cormack et al. (2009) paper.
	rrfK = 60

	// candidateMultiplier controls how many candidates each retriever fetches.
	// We fetch more than the final limit so RRF has a larger pool to fuse.
	candidateMultiplier = 3
)

// rrfCombinedSearch runs keyword and semantic searches in parallel, then
// fuses results using Reciprocal Rank Fusion.
func (e *Engine) rrfCombinedSearch(ctx context.Context, query string, opts Options) ([]Result, error) {
	candidateLimit := opts.Limit * candidateMultiplier
	if candidateLimit < 30 {
		candidateLimit = 30 // minimum pool size
	}

	// Broader candidate options for each retriever
	kwOpts := opts
	kwOpts.Limit = candidateLimit
	vecOpts := opts
	vecOpts.Limit = candidateLimit

	// Run both retrievers in parallel
	type searchResult struct {
		results []Result
		err     error
	}
	kwCh := make(chan searchResult, 1)
	vecCh := make(chan searchResult, 1)

	go func() {
		r, err := e.keywordSearch(query, kwOpts)
		kwCh <- searchResult{r, err}
	}()

	go func() {
		if e.vec == nil {
			vecCh <- searchResult{nil, nil}
			return
		}
		r, err := e.semanticSearch(ctx, query, vecOpts)
		vecCh <- searchResult{r, err}
	}()

	kwRes := <-kwCh
	vecRes := <-vecCh

	// If one retriever failed entirely, fall back to the other
	if kwRes.err != nil && vecRes.err != nil {
		return nil, kwRes.err
	}
	if kwRes.err != nil || len(kwRes.results) == 0 {
		return truncate(vecRes.results, opts.Limit), nil
	}
	if vecRes.err != nil || len(vecRes.results) == 0 {
		return truncate(kwRes.results, opts.Limit), nil
	}

	// Build rank maps: document key → 1-based rank position
	kwRanks := buildRankMap(kwRes.results)
	vecRanks := buildRankMap(vecRes.results)

	// Fuse: collect all unique documents, compute RRF score
	type fusedDoc struct {
		result   Result
		rrfScore float64
		kwRank   int // 0 means absent from keyword results
		vecRank  int // 0 means absent from semantic results
	}

	seen := make(map[string]*fusedDoc)
	var docs []*fusedDoc

	// Process keyword results
	for _, r := range kwRes.results {
		key := docKey(r)
		d := &fusedDoc{result: r, kwRank: kwRanks[key]}
		seen[key] = d
		docs = append(docs, d)
	}

	// Process semantic results
	for _, r := range vecRes.results {
		key := docKey(r)
		if d, exists := seen[key]; exists {
			d.vecRank = vecRanks[key]
			// Prefer the semantic result content if it's richer (summaries/themes)
			if r.Type == "summary" || r.Type == "theme" {
				d.result.Content = r.Content
			}
		} else {
			d := &fusedDoc{result: r, vecRank: vecRanks[key]}
			seen[key] = d
			docs = append(docs, d)
		}
	}

	// Compute RRF scores
	for _, d := range docs {
		d.rrfScore = rrfScore(d.kwRank, d.vecRank)
		d.result.Score = d.rrfScore
	}

	// Sort by RRF score (descending)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].rrfScore > docs[j].rrfScore
	})

	// Build final results
	results := make([]Result, 0, opts.Limit)
	for i, d := range docs {
		if i >= opts.Limit {
			break
		}
		results = append(results, d.result)
	}

	return results, nil
}

// rrfScore computes the Reciprocal Rank Fusion score for a document.
// rank1 and rank2 are 1-based positions (0 means the document was absent).
func rrfScore(rank1, rank2 int) float64 {
	score := 0.0
	if rank1 > 0 {
		score += 1.0 / float64(rrfK+rank1)
	}
	if rank2 > 0 {
		score += 1.0 / float64(rrfK+rank2)
	}
	return score
}

// buildRankMap creates a map from document key to 1-based rank position.
func buildRankMap(results []Result) map[string]int {
	ranks := make(map[string]int, len(results))
	for i, r := range results {
		key := docKey(r)
		if _, exists := ranks[key]; !exists {
			ranks[key] = i + 1 // 1-based
		}
	}
	return ranks
}

// docKey generates a deduplication key for a result.
// Handles different reference formats from keyword vs semantic search
// (e.g., "1-cor 13:13" vs "1 Corinthians 13:13") by using FilePath + verse number.
func docKey(r Result) string {
	if r.FilePath != "" {
		if r.Type == "verse" {
			// Extract verse number from reference (after last ":")
			if idx := strings.LastIndex(r.Reference, ":"); idx >= 0 {
				return r.FilePath + "|v" + strings.TrimSpace(r.Reference[idx+1:])
			}
		}
		if r.Type == "paragraph" {
			// Paragraphs share files; include reference for granularity
			return r.FilePath + "|" + r.Reference
		}
		// File-level results (talks, chapters, manuals): FilePath alone deduplicates
		return r.FilePath
	}
	return r.Reference
}

// truncate returns at most n items from results.
func truncate(results []Result, n int) []Result {
	if len(results) <= n {
		return results
	}
	return results[:n]
}
