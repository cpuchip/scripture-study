// Package dictionary provides Webster 1828 dictionary loading and lookup.
package dictionary

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Webster holds a loaded Webster dictionary edition.
type Webster struct {
	entries  map[string][]WebsterEntry // Word -> entries (may have multiple POS)
	variants map[string]string         // modern/KJV form -> actual headword (optional)
}

// NewWebster creates a new Webster dictionary instance.
func NewWebster() *Webster {
	return &Webster{
		entries:  make(map[string][]WebsterEntry),
		variants: make(map[string]string),
	}
}

// LoadVariants loads a spelling-variant map (e.g. ALLEGE -> ALLEDGE).
// Keys starting with "_" are comments.
func (w *Webster) LoadVariants(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read variants file: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("failed to parse variants file: %w", err)
	}
	for k, v := range m {
		if !strings.HasPrefix(k, "_") {
			w.variants[strings.ToUpper(k)] = strings.ToUpper(v)
		}
	}
	return nil
}

// LoadFromFile loads the dictionary from a JSON file.
// Supports both plain .json and gzip-compressed .json.gz files.
func (w *Webster) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open dictionary file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// If the file ends with .gz, decompress it
	if strings.HasSuffix(strings.ToLower(path), ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var entries []WebsterEntry
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&entries); err != nil {
		return fmt.Errorf("failed to parse dictionary JSON: %w", err)
	}

	// Build the lookup map
	for _, entry := range entries {
		key := strings.ToUpper(entry.Word)
		w.entries[key] = append(w.entries[key], entry)
	}

	return nil
}

// EntryCount returns the number of unique words in the dictionary.
func (w *Webster) EntryCount() int {
	return len(w.entries)
}

// Lookup finds a word in the dictionary.
// Returns nil if not found.
func (w *Webster) Lookup(word string) []WebsterEntry {
	key := strings.ToUpper(strings.TrimSpace(word))
	entries, ok := w.entries[key]
	if !ok {
		return nil
	}
	return entries
}

// LookupFirst returns the first entry for a word (convenience method).
func (w *Webster) LookupFirst(word string) *WebsterEntry {
	entries := w.Lookup(word)
	if len(entries) == 0 {
		return nil
	}
	return &entries[0]
}

// archaic/inflection suffixes tried by the stem fallback, longest first.
var stemSuffixes = []string{"edst", "eth", "est", "ing", "ed", "s"}

// LookupWithFallback finds a word by exact match, then the spelling-variant
// map (1828 orthography: ALLEGE -> ALLEDGE), then archaic-suffix stemming
// ("sleepeth" -> SLEEP). Returns the entries, the headword that actually
// matched, and how: "exact", "variant", or "stem". Nil when nothing matches.
func (w *Webster) LookupWithFallback(word string) ([]WebsterEntry, string, string) {
	key := strings.ToUpper(strings.TrimSpace(word))
	if entries, ok := w.entries[key]; ok {
		return entries, key, "exact"
	}
	if target, ok := w.variants[key]; ok {
		if entries, ok := w.entries[target]; ok {
			return entries, target, "variant"
		}
	}
	for _, suf := range stemSuffixes {
		if !strings.HasSuffix(key, strings.ToUpper(suf)) || len(key) < len(suf)+3 {
			continue
		}
		stem := key[:len(key)-len(suf)]
		for _, cand := range []string{stem, stem + "E", strings.TrimSuffix(stem, stem[len(stem)-1:])} {
			if entries, ok := w.entries[cand]; ok {
				return entries, cand, "stem"
			}
		}
	}
	return nil, "", ""
}

// Search searches for words containing the query string.
// Returns up to maxResults matching words.
func (w *Webster) Search(query string, maxResults int) []string {
	query = strings.ToUpper(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	var results []string

	// First, exact match
	if _, ok := w.entries[query]; ok {
		results = append(results, query)
	}

	// Then prefix matches
	for word := range w.entries {
		if len(results) >= maxResults {
			break
		}
		if word != query && strings.HasPrefix(word, query) {
			results = append(results, word)
		}
	}

	// Then contains matches
	for word := range w.entries {
		if len(results) >= maxResults {
			break
		}
		// Skip if already added
		found := false
		for _, r := range results {
			if r == word {
				found = true
				break
			}
		}
		if !found && strings.Contains(word, query) {
			results = append(results, word)
		}
	}

	return results
}

// SearchDefinitions searches for words where the query appears in definitions.
// Returns up to maxResults matching entries.
func (w *Webster) SearchDefinitions(query string, maxResults int) []WebsterEntry {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	var results []WebsterEntry
	for _, entries := range w.entries {
		for _, entry := range entries {
			if len(results) >= maxResults {
				return results
			}
			for _, def := range entry.Definitions {
				if strings.Contains(strings.ToLower(def), query) {
					results = append(results, entry)
					break
				}
			}
		}
	}

	return results
}

// FormatEntry formats a Webster entry as a readable string.
func FormatEntry(entry *WebsterEntry) string {
	if entry == nil {
		return ""
	}

	var sb strings.Builder
	if entry.POS != "" {
		sb.WriteString(fmt.Sprintf("**%s** (%s)\n", entry.Word, entry.POS))
	} else {
		sb.WriteString(fmt.Sprintf("**%s**\n", entry.Word))
	}

	if entry.Synonyms != "" {
		sb.WriteString(fmt.Sprintf("*Synonyms:* %s\n", entry.Synonyms))
	}

	if entry.Etymology != "" {
		sb.WriteString(fmt.Sprintf("*Etymology:* %s\n", entry.Etymology))
	}

	sb.WriteString("\n**Definitions:**\n")
	for i, def := range entry.Definitions {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, def))
	}

	for _, note := range entry.Notes {
		sb.WriteString(fmt.Sprintf("_%s_\n", note))
	}

	return sb.String()
}

// FormatEntries formats multiple entries for the same word.
func FormatEntries(entries []WebsterEntry) string {
	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, entry := range entries {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}
		sb.WriteString(FormatEntry(&entry))
	}
	return sb.String()
}
