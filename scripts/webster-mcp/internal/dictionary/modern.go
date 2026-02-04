// Package dictionary provides modern dictionary lookup via the Free Dictionary API.
package dictionary

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// FreeDictionaryAPIURL is the base URL for the Free Dictionary API.
	FreeDictionaryAPIURL = "https://api.dictionaryapi.dev/api/v2/entries/en"

	// DefaultTimeout is the default HTTP timeout for API requests.
	DefaultTimeout = 10 * time.Second
)

// ModernDict provides access to the Free Dictionary API.
type ModernDict struct {
	client  *http.Client
	baseURL string
}

// NewModernDict creates a new modern dictionary client.
func NewModernDict() *ModernDict {
	return &ModernDict{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: FreeDictionaryAPIURL,
	}
}

// Lookup fetches the definition of a word from the Free Dictionary API.
// Returns nil if the word is not found.
func (m *ModernDict) Lookup(word string) ([]ModernEntry, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	// URL encode the word
	reqURL := fmt.Sprintf("%s/%s", m.baseURL, url.PathEscape(word))

	resp, err := m.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch definition: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for 404 (word not found)
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var entries []ModernEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return entries, nil
}

// FormatModernEntry formats a modern dictionary entry as a readable string.
func FormatModernEntry(entry *ModernEntry) string {
	if entry == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**", entry.Word))

	if entry.Phonetic != "" {
		sb.WriteString(fmt.Sprintf(" %s", entry.Phonetic))
	}
	sb.WriteString("\n")

	if entry.Origin != "" {
		sb.WriteString(fmt.Sprintf("*Origin:* %s\n", entry.Origin))
	}

	for _, meaning := range entry.Meanings {
		sb.WriteString(fmt.Sprintf("\n**%s**\n", meaning.PartOfSpeech))

		for i, def := range meaning.Definitions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, def.Definition))
			if def.Example != "" {
				sb.WriteString(fmt.Sprintf("   _Example: \"%s\"_\n", def.Example))
			}
		}

		if len(meaning.Synonyms) > 0 {
			sb.WriteString(fmt.Sprintf("   *Synonyms:* %s\n", strings.Join(meaning.Synonyms, ", ")))
		}
		if len(meaning.Antonyms) > 0 {
			sb.WriteString(fmt.Sprintf("   *Antonyms:* %s\n", strings.Join(meaning.Antonyms, ", ")))
		}
	}

	return sb.String()
}

// FormatModernEntries formats multiple modern dictionary entries.
func FormatModernEntries(entries []ModernEntry) string {
	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, entry := range entries {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}
		sb.WriteString(FormatModernEntry(&entry))
	}
	return sb.String()
}
