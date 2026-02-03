// Package tools implements the MCP tools for gospel content access.
package tools

import (
	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/db"
)

// Tools provides the gospel MCP tools.
type Tools struct {
	db *db.DB
}

// New creates a new Tools instance.
func New(database *db.DB) *Tools {
	return &Tools{db: database}
}

// RelatedReference represents a cross-reference or citation.
type RelatedReference struct {
	Reference string `json:"reference"`
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Reference         string             `json:"reference"`
	Title             string             `json:"title,omitempty"`
	Excerpt           string             `json:"excerpt"`
	Content           string             `json:"content,omitempty"`
	ContextBefore     []string           `json:"context_before,omitempty"`
	ContextAfter      []string           `json:"context_after,omitempty"`
	FilePath          string             `json:"file_path"`
	MarkdownLink      string             `json:"markdown_link"` // Pre-formatted markdown link for easy use in study documents
	SourceURL         string             `json:"source_url"`
	RelatedReferences []RelatedReference `json:"related_references,omitempty"`
	SourceType        string             `json:"source_type"`
	RelevanceScore    float64            `json:"relevance_score,omitempty"`
}

// SearchResponse is the full search response.
type SearchResponse struct {
	Query        string         `json:"query"`
	TotalMatches int            `json:"total_matches"`
	Results      []SearchResult `json:"results"`
	QueryTimeMs  int64          `json:"query_time_ms"`
}

// GetResponse is the response from gospel_get.
type GetResponse struct {
	Reference         string             `json:"reference"`
	Title             string             `json:"title,omitempty"`
	Content           string             `json:"content"`
	ContextBefore     []VerseContext     `json:"context_before,omitempty"`
	ContextAfter      []VerseContext     `json:"context_after,omitempty"`
	ChapterContent    string             `json:"chapter_content,omitempty"`
	FilePath          string             `json:"file_path"`
	SourceURL         string             `json:"source_url"`
	RelatedReferences []RelatedReference `json:"related_references,omitempty"`
	SourceType        string             `json:"source_type"`
}

// VerseContext provides context verses.
type VerseContext struct {
	Verse int    `json:"verse,omitempty"`
	Text  string `json:"text"`
}

// ListItem represents an item in a listing.
type ListItem struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Chapters int    `json:"chapters,omitempty"`
	Count    int    `json:"count,omitempty"`
}

// ListResponse is the response from gospel_list.
type ListResponse struct {
	Path  string     `json:"path"`
	Items []ListItem `json:"items"`
	Total int        `json:"total"`
}
