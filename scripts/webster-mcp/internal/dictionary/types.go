// Package dictionary provides types and functions for dictionary lookups.
package dictionary

// WebsterEntry represents a single entry in the Webster 1828 dictionary.
type WebsterEntry struct {
	Word        string   `json:"word"`
	POS         string   `json:"pos"`
	Synonyms    string   `json:"synonyms,omitempty"`
	Definitions []string `json:"definitions"`
}

// ModernEntry represents a response from the Free Dictionary API.
type ModernEntry struct {
	Word      string           `json:"word"`
	Phonetic  string           `json:"phonetic,omitempty"`
	Phonetics []ModernPhonetic `json:"phonetics,omitempty"`
	Origin    string           `json:"origin,omitempty"`
	Meanings  []ModernMeaning  `json:"meanings,omitempty"`
}

// ModernPhonetic represents pronunciation information.
type ModernPhonetic struct {
	Text      string `json:"text,omitempty"`
	Audio     string `json:"audio,omitempty"`
	SourceURL string `json:"sourceUrl,omitempty"`
}

// ModernMeaning represents a meaning/definition group.
type ModernMeaning struct {
	PartOfSpeech string             `json:"partOfSpeech"`
	Definitions  []ModernDefinition `json:"definitions"`
	Synonyms     []string           `json:"synonyms,omitempty"`
	Antonyms     []string           `json:"antonyms,omitempty"`
}

// ModernDefinition represents a single definition.
type ModernDefinition struct {
	Definition string   `json:"definition"`
	Example    string   `json:"example,omitempty"`
	Synonyms   []string `json:"synonyms,omitempty"`
	Antonyms   []string `json:"antonyms,omitempty"`
}

// CombinedResult holds definitions from both dictionaries.
type CombinedResult struct {
	Word    string        `json:"word"`
	Webster *WebsterEntry `json:"webster,omitempty"`
	Modern  []ModernEntry `json:"modern,omitempty"`
	Error   string        `json:"error,omitempty"`
}
