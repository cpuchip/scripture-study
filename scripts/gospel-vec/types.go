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
	SourceManual     Source = "manual"
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

	// Conference talk fields (used when Source == SourceConference)
	Speaker   string `json:"speaker,omitempty"`   // Speaker name
	Position  string `json:"position,omitempty"`  // Speaker's calling
	Year      int    `json:"year,omitempty"`      // Conference year
	Month     string `json:"month,omitempty"`     // "04" or "10"
	Session   string `json:"session,omitempty"`   // "Saturday Morning", etc.
	TalkTitle string `json:"talktitle,omitempty"` // Talk title
}

// ToMap converts metadata to map[string]string for chromem-go
func (m *DocMetadata) ToMap() map[string]string {
	result := map[string]string{
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

	// Add conference talk fields if present
	if m.Speaker != "" {
		result["speaker"] = m.Speaker
	}
	if m.Position != "" {
		result["position"] = m.Position
	}
	if m.Year > 0 {
		result["year"] = fmt.Sprintf("%d", m.Year)
	}
	if m.Month != "" {
		result["month"] = m.Month
	}
	if m.Session != "" {
		result["session"] = m.Session
	}
	if m.TalkTitle != "" {
		result["talktitle"] = m.TalkTitle
	}

	return result
}

// MetadataFromMap converts map back to DocMetadata
func MetadataFromMap(m map[string]string) *DocMetadata {
	chapter := 0
	fmt.Sscanf(m["chapter"], "%d", &chapter)
	generated := m["generated"] == "true"

	year := 0
	if m["year"] != "" {
		fmt.Sscanf(m["year"], "%d", &year)
	}

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
		// Conference talk fields
		Speaker:   m["speaker"],
		Position:  m["position"],
		Year:      year,
		Month:     m["month"],
		Session:   m["session"],
		TalkTitle: m["talktitle"],
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
