package enricher

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
)

//go:embed scripture_prompt.md
var scripturePromptTemplate string

//go:embed gospel_vocab.md
var gospelVocabContext string

//go:embed titsw_framework.md
var titswFrameworkContext string

// ScriptureProfile holds a parsed enrichment profile for a scripture chapter.
type ScriptureProfile struct {
	Summary     string // 75-100 word narrative
	Keywords    string // comma-separated
	KeyVerse    string // most significant verse
	ChristTypes string // comma-separated typological connections
	Connections string // cross-dispensation connections (raw text)
	RawOutput   string // full LLM response
}

// EnrichScripture analyzes a scripture chapter using the lens approach
// (gospel-vocab + titsw-framework as system context) and returns its profile.
func (e *Enricher) EnrichScripture(ctx context.Context, chapterContent string) (*ScriptureProfile, error) {
	// Build the user message: substitute chapter content into the prompt template
	userMsg := strings.Replace(scripturePromptTemplate, "{{CONTENT}}", chapterContent, 1)

	// The lens context (gospel-vocab + titsw-framework) is the system message
	systemMsg := gospelVocabContext + "\n\n---\n\n" + titswFrameworkContext

	text, _, err := e.client.Complete(ctx, systemMsg, userMsg, e.temperature)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	// Strip reasoning model's <think>...</think> block if present
	text = stripThinkingBlock(text)

	profile, err := parseScriptureProfile(text)
	if err != nil {
		return nil, fmt.Errorf("parsing scripture profile: %w", err)
	}
	profile.RawOutput = text

	return profile, nil
}

// normalizeLabels strips markdown formatting and ensures section labels end with colons.
// Handles model outputs like "### **SUMMARY**" → "SUMMARY:" and "**KEYWORDS:**" → "KEYWORDS:"
var labelNormalizeRe = regexp.MustCompile(`(?im)^[#\s]*\*{0,2}(KEYWORDS|SUMMARY|KEY_VERSE|CHRIST_TYPES|CONNECTIONS)\*{0,2}:?\s*$`)

func normalizeScriptureOutput(text string) string {
	// Strip --- horizontal rules
	text = regexp.MustCompile(`(?m)^---+\s*$`).ReplaceAllString(text, "")
	// Strip ### heading markers and ** bold from label lines, ensure colon
	text = labelNormalizeRe.ReplaceAllStringFunc(text, func(match string) string {
		// Extract just the label name
		m := labelNormalizeRe.FindStringSubmatch(match)
		if len(m) < 2 {
			return match
		}
		return m[1] + ":"
	})
	// Also handle inline bold labels like "**KEYWORDS:** value"
	text = regexp.MustCompile(`\*{2}(KEYWORDS|SUMMARY|KEY_VERSE|CHRIST_TYPES|CONNECTIONS):?\*{2}:?`).ReplaceAllString(text, "$1:")
	return text
}

func parseScriptureProfile(text string) (*ScriptureProfile, error) {
	// Normalize model output format before parsing
	text = normalizeScriptureOutput(text)

	p := &ScriptureProfile{}

	// Extract fields using section boundaries (handles both inline and multi-line values)
	// Expected order: KEYWORDS, SUMMARY, KEY_VERSE, CHRIST_TYPES, CONNECTIONS
	p.Keywords = cleanMarkdown(strings.TrimSpace(extractBetween(text, "KEYWORDS:", "SUMMARY:")))
	if p.Keywords == "" {
		return nil, fmt.Errorf("could not find KEYWORDS field")
	}

	p.Summary = cleanMarkdown(strings.TrimSpace(extractBetween(text, "SUMMARY:", "KEY_VERSE:")))
	if p.Summary == "" {
		// Fallback: try between SUMMARY: and CHRIST_TYPES:
		p.Summary = cleanMarkdown(strings.TrimSpace(extractBetween(text, "SUMMARY:", "CHRIST_TYPES:")))
	}

	p.KeyVerse = cleanMarkdown(strings.TrimSpace(extractBetween(text, "KEY_VERSE:", "CHRIST_TYPES:")))

	p.ChristTypes = cleanMarkdown(strings.TrimSpace(extractBetween(text, "CHRIST_TYPES:", "CONNECTIONS:")))

	// Connections — everything after CONNECTIONS: to the end
	p.Connections = cleanMarkdown(strings.TrimSpace(extractAfter(text, "CONNECTIONS:")))

	return p, nil
}

// extractAfter returns everything after the first occurrence of marker.
func extractAfter(text, marker string) string {
	idx := strings.Index(strings.ToUpper(text), strings.ToUpper(marker))
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(text[idx+len(marker):])
}
