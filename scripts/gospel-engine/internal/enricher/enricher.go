// Package enricher handles TITSW teaching profile extraction for conference talks.
package enricher

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/llm"
)

//go:embed prompt.md
var promptTemplate string

//go:embed calibration.md
var calibrationContext string

// TITSWProfile holds a parsed TITSW teaching profile for a talk.
type TITSWProfile struct {
	Dominant string // comma-separated dominant dimensions
	Mode     string // enacted, declared, doctrinal, experiential
	Pattern  string // e.g., "story→doctrine→invitation"
	Teach    int    // 0-9
	Help     int    // 0-9
	Love     int    // 0-9
	Spirit   int    // 0-9
	Doctrine int    // 0-9
	Invite   int    // 0-9
	Summary  string
	KeyQuote string
	Keywords string // comma-separated

	// Stored for analysis but not indexed for search
	Reasoning string
	RawOutput string
}

// Enricher generates TITSW teaching profiles using an LLM.
type Enricher struct {
	client      *llm.Client
	temperature float64
}

// New creates a new TITSW enricher.
func New(client *llm.Client, temperature float64) *Enricher {
	return &Enricher{
		client:      client,
		temperature: temperature,
	}
}

// Enrich analyzes a talk and returns its TITSW teaching profile.
func (e *Enricher) Enrich(ctx context.Context, talkContent string) (*TITSWProfile, error) {
	// Build the user message: substitute talk content into the prompt template
	userMsg := strings.Replace(promptTemplate, "{{CONTENT}}", talkContent, 1)

	// The calibration context is the system message
	systemMsg := calibrationContext

	text, _, err := e.client.Complete(ctx, systemMsg, userMsg, e.temperature)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	profile, err := parseProfile(text)
	if err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	profile.RawOutput = text

	return profile, nil
}

// Score regexes: handle formats like "7", "**7**", "[7]", "7 — justification"
var (
	teachRe    = regexp.MustCompile(`(?i)TEACH_SCORE:\s*[*\[\s]*(\d)[*\]\s]*`)
	helpRe     = regexp.MustCompile(`(?i)HELP_SCORE:\s*[*\[\s]*(\d)[*\]\s]*`)
	loveRe     = regexp.MustCompile(`(?i)LOVE_SCORE:\s*[*\[\s]*(\d)[*\]\s]*`)
	spiritRe   = regexp.MustCompile(`(?i)SPIRIT_SCORE:\s*[*\[\s]*(\d)[*\]\s]*`)
	doctrineRe = regexp.MustCompile(`(?i)DOCTRINE_SCORE:\s*[*\[\s]*(\d)[*\]\s]*`)
	inviteRe   = regexp.MustCompile(`(?i)INVITE_SCORE:\s*[*\[\s]*(\d)[*\]\s]*`)

	dominantRe = regexp.MustCompile(`(?i)DOMINANT:\s*(.+)`)
	modeRe     = regexp.MustCompile(`(?i)MODE:\s*(.+)`)
	patternRe  = regexp.MustCompile(`(?i)PATTERN:\s*(.+)`)
	keywordsRe = regexp.MustCompile(`(?i)KEYWORDS:\s*(.+)`)
	keyQuoteRe = regexp.MustCompile(`(?i)KEY_QUOTE:\s*(.+)`)
)

func parseProfile(text string) (*TITSWProfile, error) {
	p := &TITSWProfile{}

	// Extract scores
	var err error
	p.Teach, err = extractScore(teachRe, text, "teach")
	if err != nil {
		return nil, err
	}
	p.Help, err = extractScore(helpRe, text, "help")
	if err != nil {
		return nil, err
	}
	p.Love, err = extractScore(loveRe, text, "love")
	if err != nil {
		return nil, err
	}
	p.Spirit, err = extractScore(spiritRe, text, "spirit")
	if err != nil {
		return nil, err
	}
	p.Doctrine, err = extractScore(doctrineRe, text, "doctrine")
	if err != nil {
		return nil, err
	}
	p.Invite, err = extractScore(inviteRe, text, "invite")
	if err != nil {
		return nil, err
	}

	// Extract text fields — strip markdown bold/italic markers from LLM output
	p.Dominant = cleanMarkdown(extractField(dominantRe, text))
	p.Mode = cleanMarkdown(extractField(modeRe, text))
	p.Pattern = cleanMarkdown(extractField(patternRe, text))
	p.Keywords = cleanMarkdown(extractField(keywordsRe, text))
	p.KeyQuote = cleanMarkdown(extractField(keyQuoteRe, text))

	// Clean up mode — take first word (before any parenthetical)
	if idx := strings.IndexAny(p.Mode, " ("); idx > 0 {
		p.Mode = strings.TrimSpace(p.Mode[:idx])
	}
	p.Mode = strings.ToLower(p.Mode)

	// Extract summary — everything between SUMMARY: and KEY_QUOTE:
	p.Summary = extractBetween(text, "SUMMARY:", "KEY_QUOTE:")

	// Extract reasoning — everything between REASONING: and the first _SCORE:
	p.Reasoning = extractBetween(text, "REASONING:", "TEACH_SCORE:")
	if p.Reasoning == "" {
		// Try alternative: between REASONING: and CALIBRATION
		p.Reasoning = extractBetween(text, "REASONING:", "CALIBRATION")
	}

	return p, nil
}

func extractScore(re *regexp.Regexp, text, name string) (int, error) {
	m := re.FindStringSubmatch(text)
	if m == nil {
		return 0, fmt.Errorf("could not find %s score", name)
	}
	score, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, fmt.Errorf("invalid %s score: %s", name, m[1])
	}
	if score < 0 || score > 9 {
		return 0, fmt.Errorf("%s score %d out of range 0-9", name, score)
	}
	return score, nil
}

func extractField(re *regexp.Regexp, text string) string {
	m := re.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// cleanMarkdown strips bold/italic markdown markers (* and **) from LLM output.
func cleanMarkdown(s string) string {
	s = strings.ReplaceAll(s, "*", "")
	return strings.TrimSpace(s)
}

func extractBetween(text, start, end string) string {
	startIdx := strings.Index(strings.ToUpper(text), strings.ToUpper(start))
	if startIdx < 0 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(strings.ToUpper(text[startIdx:]), strings.ToUpper(end))
	if endIdx < 0 {
		// Take everything after start
		return strings.TrimSpace(text[startIdx:])
	}

	return strings.TrimSpace(text[startIdx : startIdx+endIdx])
}
