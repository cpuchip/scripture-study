package main

import "fmt"

// Layer represents the granularity of indexed content
type Layer string

const (
	LayerVerse     Layer = "verse"
	LayerParagraph Layer = "paragraph"
	LayerSummary   Layer = "summary"
	LayerTheme     Layer = "theme"
)

// Source represents the content source
type Source string

const (
	SourceScriptures Source = "scriptures"
	SourceConference Source = "conference"
)

// DocMetadata contains metadata for each indexed document
type DocMetadata struct {
	Source    Source `json:"source"`    // "scriptures", "conference"
	Layer     Layer  `json:"layer"`     // "verse", "paragraph", "summary", "theme"
	Book      string `json:"book"`      // "1 Nephi", "D&C", etc.
	Chapter   int    `json:"chapter"`   // Chapter number
	Reference string `json:"reference"` // Full reference "1 Nephi 3:7"
	Range     string `json:"range"`     // Verse range for paragraphs/themes "1-10"
	FilePath  string `json:"filepath"`  // Source file path
	Generated bool   `json:"generated"` // True if LLM-generated
	Model     string `json:"model"`     // Model used if generated
	Timestamp string `json:"timestamp"` // When indexed
}

// ToMap converts metadata to map[string]string for chromem-go
func (m *DocMetadata) ToMap() map[string]string {
	return map[string]string{
		"source":    string(m.Source),
		"layer":     string(m.Layer),
		"book":      m.Book,
		"chapter":   fmt.Sprintf("%d", m.Chapter),
		"reference": m.Reference,
		"range":     m.Range,
		"filepath":  m.FilePath,
		"generated": fmt.Sprintf("%t", m.Generated),
		"model":     m.Model,
		"timestamp": m.Timestamp,
	}
}

// MetadataFromMap converts map back to DocMetadata
func MetadataFromMap(m map[string]string) *DocMetadata {
	chapter := 0
	fmt.Sscanf(m["chapter"], "%d", &chapter)
	generated := m["generated"] == "true"

	return &DocMetadata{
		Source:    Source(m["source"]),
		Layer:     Layer(m["layer"]),
		Book:      m["book"],
		Chapter:   chapter,
		Reference: m["reference"],
		Range:     m["range"],
		FilePath:  m["filepath"],
		Generated: generated,
		Model:     m["model"],
		Timestamp: m["timestamp"],
	}
}

// Chunk represents a piece of content to be indexed
type Chunk struct {
	ID       string       // Unique identifier
	Content  string       // Text content
	Metadata *DocMetadata // Associated metadata
}

// SearchResult represents a search result with score
type SearchResult struct {
	Chunk
	Score float32 // Similarity score (0-1)
}

// SearchOptions controls search behavior
type SearchOptions struct {
	Layers  []Layer  // Which layers to search (empty = all)
	Sources []Source // Which sources to search (empty = all)
	Limit   int      // Max results per layer
}

// DefaultSearchOptions returns sensible defaults
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Layers:  nil, // All layers
		Sources: nil, // All sources
		Limit:   10,
	}
}
