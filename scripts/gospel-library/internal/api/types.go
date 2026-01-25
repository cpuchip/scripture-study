// Package api provides types and client for the Gospel Library API.
package api

import "encoding/json"

// DynamicResponse is the response from the type/dynamic endpoint.
// The structure varies depending on the URI:
// - Top-level collections (e.g., /general-conference) have a "collection" field
// - Specific items (e.g., /general-conference/2024/10) have "content" and "toc" fields
type DynamicResponse struct {
	// For top-level collections
	Collection *Collection `json:"collection,omitempty"`

	// For specific items (conferences, manuals, etc.)
	Content *DynamicContent `json:"content,omitempty"`
	TOC     *TOC            `json:"toc,omitempty"`
}

// DynamicContent is the content wrapper in dynamic responses.
type DynamicContent struct {
	Meta    Meta    `json:"meta"`
	Content Content `json:"content"`
	URI     string  `json:"uri"`
}

// TOC represents the table of contents in dynamic responses.
type TOC struct {
	Title           string       `json:"title"`
	URI             string       `json:"uri"`
	Category        string       `json:"category"`
	ParentName      string       `json:"parentName"`
	ParentURI       string       `json:"parentUri"`
	BreadCrumbs     []BreadCrumb `json:"breadCrumbs"`
	Entries         []TOCEntry   `json:"entries"`
	FirstContentURI string       `json:"firstContentUri"`
	ContentType     string       `json:"contentType"`
}

// TOCEntry represents an entry in the table of contents.
// It can be either a content reference or a section with nested entries.
type TOCEntry struct {
	// For direct content entries
	Content *TOCContentRef `json:"content,omitempty"`

	// For sections (like conference sessions)
	Section *TOCSection `json:"section,omitempty"`
}

// TOCContentRef is a reference to content in the TOC.
type TOCContentRef struct {
	URI   string `json:"uri"`
	Title string `json:"title"`
}

// TOCSection represents a section in the TOC (like a conference session).
type TOCSection struct {
	Title          string     `json:"title"`
	Type           string     `json:"type"`
	URI            string     `json:"uri"`
	ChildURIs      []string   `json:"childUris"`
	DirectChildURI []string   `json:"directChildUris"`
	Entries        []TOCEntry `json:"entries"`
}

// CollectionResponse is the response from the dynamic/collection endpoint.
// This is a simplified wrapper for top-level collections.
type CollectionResponse struct {
	Collection Collection `json:"collection"`
}

// Collection represents a navigable collection (conference, book, etc.).
type Collection struct {
	BreadCrumbs []BreadCrumb `json:"breadCrumbs"`
	Title       string       `json:"title"`
	URI         string       `json:"uri"`
	Sections    []Section    `json:"sections"`
}

// BreadCrumb represents a navigation breadcrumb.
type BreadCrumb struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
}

// Section represents a group of entries in a collection.
type Section struct {
	Title      string  `json:"title"`
	SectionKey string  `json:"sectionKey"`
	Entries    []Entry `json:"entries"`
}

// Entry represents a single item in a collection section.
type Entry struct {
	Title    string `json:"title"`
	URI      string `json:"uri"`
	Type     string `json:"type"` // "item", "collection", "search"
	Src      string `json:"src"`  // thumbnail image URL
	Archived bool   `json:"archived"`
	Category string `json:"category"`
	Position int    `json:"position"`
}

// ContentResponse is the response from the content endpoint.
type ContentResponse struct {
	Meta               Meta            `json:"meta"`
	Content            Content         `json:"content"`
	PIDs               json.RawMessage `json:"pids"` // Complex nested array, not needed for our purposes
	TableOfContentsURI string          `json:"tableOfContentsUri"`
	URI                string          `json:"uri"`
	Verified           bool            `json:"verified"`
	Restricted         int             `json:"restricted"` // 0 = not restricted
}

// Meta contains metadata about the content.
type Meta struct {
	Title          string          `json:"title"`
	CanonicalURL   string          `json:"canonicalUrl"`
	ContentType    string          `json:"contentType"`
	Audio          json.RawMessage `json:"audio"` // Can be array or object depending on content type
	Video          json.RawMessage `json:"video"` // Can be array or object depending on content type
	PDF            json.RawMessage `json:"pdf"`   // Can be array or object depending on content type
	PageAttributes PageAttributes  `json:"pageAttributes"`
	OGTagImageURL  string          `json:"ogTagImageUrl"`
	StructuredData string          `json:"structuredData"` // JSON string with schema.org data
}

// MediaItem represents an audio, video, or PDF resource.
type MediaItem struct {
	MediaURL string `json:"mediaUrl"`
	Variant  string `json:"variant"`
}

// PDFItem represents a PDF resource (different structure from MediaItem).
type PDFItem struct {
	Source string `json:"source"`
	Name   string `json:"name"`
}

// GetAudioItems parses the Audio field into MediaItems.
func (m *Meta) GetAudioItems() []MediaItem {
	if len(m.Audio) == 0 || string(m.Audio) == "null" || string(m.Audio) == "{}" {
		return nil
	}
	var items []MediaItem
	if err := json.Unmarshal(m.Audio, &items); err != nil {
		// Try single item
		var item MediaItem
		if err := json.Unmarshal(m.Audio, &item); err == nil && item.MediaURL != "" {
			return []MediaItem{item}
		}
		return nil
	}
	return items
}

// PageAttributes contains data attributes for the page.
type PageAttributes struct {
	ContentType string `json:"data-content-type"`
	URI         string `json:"data-uri"`
	AssetID     string `json:"data-asset-id"`
}

// Content contains the actual page content.
type Content struct {
	Head       json.RawMessage     `json:"head"` // Can be string or object depending on context
	Body       string              `json:"body"`
	Associated json.RawMessage     `json:"associated,omitempty"` // scriptures have this (object for related content)
	Footnotes  map[string]Footnote `json:"footnotes"`
}

// Footnote represents a single footnote/reference.
type Footnote struct {
	ID            string   `json:"id"`
	Marker        string   `json:"marker"`
	PID           string   `json:"pid"`
	Context       string   `json:"context"` // The annotated word (scriptures only)
	Text          string   `json:"text"`    // HTML content
	ReferenceURIs []RefURI `json:"referenceUris"`
}

// RefURI represents a reference link within a footnote.
type RefURI struct {
	Type string `json:"type"` // "scripture-ref", etc.
	Href string `json:"href"`
	Text string `json:"text"`
}

// StructuredData represents the parsed schema.org data from Meta.StructuredData.
type StructuredData struct {
	Context       string `json:"@context"`
	Type          string `json:"@type"`
	DatePublished string `json:"datePublished"`
}
