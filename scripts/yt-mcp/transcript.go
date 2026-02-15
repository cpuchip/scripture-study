package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ── TTML Parsing (ported from forkirk/quoter/ttml.go) ───────────────────────

// ParseTTMLFile parses a TTML subtitle file, extracting <p> elements with begin/end/text.
func ParseTTMLFile(path string) ([]Cue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	var cues []Cue

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "p" {
				var beginStr, endStr string
				for _, a := range se.Attr {
					if a.Name.Local == "begin" {
						beginStr = a.Value
					}
					if a.Name.Local == "end" {
						endStr = a.Value
					}
				}
				text, err := readInnerText(dec, "p")
				if err != nil {
					return nil, err
				}
				begin := parseClockTime(beginStr)
				end := parseClockTime(endStr)
				if strings.TrimSpace(text) != "" {
					cues = append(cues, Cue{Begin: begin, End: end, Text: normalizeWS(text)})
				}
			}
		}
	}
	return cues, nil
}

// readInnerText reads all text content inside an XML element until the matching end tag.
func readInnerText(dec *xml.Decoder, endLocal string) (string, error) {
	var b strings.Builder
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			b.WriteString(string(t))
		}
	}
	return b.String(), nil
}

// parseClockTime converts TTML time formats to seconds.
// Supports: "HH:MM:SS.mmm", "SS.mmm", "123s", "123.456s"
func parseClockTime(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Tick format: ends with "t" (ticks)
	if strings.HasSuffix(s, "t") {
		f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "t"), 64)
		// TTML default tick rate is 10000000 ticks/sec
		return f / 10000000.0
	}
	// Seconds format: ends with "s"
	if strings.HasSuffix(s, "s") {
		f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
		return f
	}
	// Clock format: HH:MM:SS.mmm
	parts := strings.Split(s, ":")
	if len(parts) == 3 {
		h, _ := strconv.ParseFloat(parts[0], 64)
		m, _ := strconv.ParseFloat(parts[1], 64)
		sec, _ := strconv.ParseFloat(parts[2], 64)
		return h*3600 + m*60 + sec
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// normalizeWS collapses all whitespace (newlines, tabs, multiple spaces) to single spaces.
func normalizeWS(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

// ── VTT Parsing ──────────────────────────────────────────────────────────────

// vttTimestampRe matches WebVTT timestamp lines: "00:00:03.120 --> 00:00:05.670 align:start position:0%"
var vttTimestampRe = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}\.\d{3})`)

// vttTagRe strips all VTT inline tags: <c>, </c>, <00:00:03.360>, etc.
var vttTagRe = regexp.MustCompile(`<[^>]*>`)

// ParseVTTFile parses a WebVTT subtitle file, extracting cues with begin/end/text.
// Handles YouTube auto-generated VTT with word-level timing tags and rolling captions.
func ParseVTTFile(path string) ([]Cue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Strip UTF-8 BOM if present
	content = strings.TrimPrefix(content, "\xef\xbb\xbf")

	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var cues []Cue

	i := 0
	// Skip WEBVTT header block (everything up to first blank line)
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" {
			i++
			break
		}
		i++
	}

	// Parse cue blocks
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip blank lines and numeric cue identifiers
		if line == "" {
			i++
			continue
		}
		if isNumericCueID(line) {
			i++
			continue
		}

		// Check for timestamp line
		matches := vttTimestampRe.FindStringSubmatch(line)
		if matches == nil {
			i++
			continue
		}

		begin := parseVTTTime(matches[1])
		end := parseVTTTime(matches[2])
		i++

		// YouTube rolling captions: the first cue's "top line" may be blank
		// (no previous text to display). Skip one leading blank line.
		if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}

		// Collect text lines until blank line or next timestamp
		var textParts []string
		for i < len(lines) {
			tl := strings.TrimSpace(lines[i])
			if tl == "" {
				break
			}
			if vttTimestampRe.MatchString(tl) {
				break
			}
			if isNumericCueID(tl) {
				break
			}
			// Strip all VTT tags and normalize
			cleaned := vttTagRe.ReplaceAllString(tl, "")
			cleaned = normalizeWS(cleaned)
			if cleaned != "" {
				textParts = append(textParts, cleaned)
			}
			i++
		}

		text := normalizeWS(strings.Join(textParts, " "))
		if text != "" {
			cues = append(cues, Cue{Begin: begin, End: end, Text: text})
		}
	}

	return cues, nil
}

// parseVTTTime converts "HH:MM:SS.mmm" to seconds.
func parseVTTTime(s string) float64 {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.ParseFloat(parts[0], 64)
	m, _ := strconv.ParseFloat(parts[1], 64)
	sec, _ := strconv.ParseFloat(parts[2], 64)
	return h*3600 + m*60 + sec
}

// isNumericCueID returns true if the line is just a number (VTT cue identifier).
func isNumericCueID(s string) bool {
	_, err := strconv.Atoi(strings.TrimSpace(s))
	return err == nil
}

// ── Cue Deduplication ────────────────────────────────────────────────────────

// DeduplicateCues removes duplicate and rolling-caption overlaps.
//
// YouTube auto-subs often repeat lines across overlapping cue boundaries.
// Rules:
//   - Exact duplicate text → skip
//   - Previous cue text is a suffix of current cue → keep only the new prefix portion
//   - Current cue text is a prefix of the next cue → skip current (will be captured by next)
func DeduplicateCues(cues []Cue) []Cue {
	if len(cues) == 0 {
		return nil
	}

	var result []Cue
	for i, c := range cues {
		text := strings.TrimSpace(c.Text)
		if text == "" {
			continue
		}

		// Skip if exact duplicate of previous
		if len(result) > 0 && text == result[len(result)-1].Text {
			continue
		}

		// Rolling caption dedup: if previous text appears as suffix of this cue,
		// keep only the new content
		if len(result) > 0 {
			prev := result[len(result)-1].Text
			if strings.HasSuffix(text, prev) && text != prev {
				// The entire previous text is repeated — nothing new here
				continue
			}
			if strings.HasPrefix(text, prev) && text != prev {
				// Current extends previous — trim the overlap and append new portion
				newPart := strings.TrimSpace(text[len(prev):])
				if newPart != "" {
					text = newPart
				} else {
					continue
				}
			}

			// YouTube VTT freeze-frame: current text is the tail end of the previous cue
			// (repeats the bottom line of a two-line rolling caption)
			if strings.HasSuffix(prev, text) && text != prev {
				continue
			}
		}

		// Skip if this text is a prefix of the next cue (will be captured there)
		if i+1 < len(cues) {
			nextText := strings.TrimSpace(cues[i+1].Text)
			if strings.HasPrefix(nextText, text) && nextText != text {
				continue
			}
		}

		result = append(result, Cue{Begin: c.Begin, End: c.End, Text: text})
	}
	return result
}

// ── Cue → Paragraph Merging ─────────────────────────────────────────────────

const (
	mergeGapThreshold     = 1.5 // seconds — cues within this gap are merged
	paragraphGapThreshold = 2.0 // seconds — gaps larger than this force a new paragraph
	sentenceGapThreshold  = 1.0 // seconds — sentence-ending punct at this gap → new paragraph
)

// MergeCuesIntoParagraphs groups sequential cues into logical paragraphs.
//
// Rules (from 01_TODO.md):
//  1. If next cue starts within 1.5s of previous cue end → same paragraph
//  2. If gap > 2.0s → new paragraph
//  3. Duplicates already handled by DeduplicateCues
//  4. Sentence-ending punctuation (. ? !) at a gap > 1.0s → new paragraph
//  5. Each paragraph records the Begin time of its first cue
func MergeCuesIntoParagraphs(cues []Cue) []Paragraph {
	if len(cues) == 0 {
		return nil
	}

	var paragraphs []Paragraph
	currentText := cues[0].Text
	currentBegin := cues[0].Begin
	prevEnd := cues[0].End

	for i := 1; i < len(cues); i++ {
		c := cues[i]
		gap := c.Begin - prevEnd

		newParagraph := false

		// Rule 2: large gap always starts new paragraph
		if gap > paragraphGapThreshold {
			newParagraph = true
		} else if gap > sentenceGapThreshold && endsWithSentence(currentText) {
			// Rule 4: sentence ending + moderate gap → new paragraph
			newParagraph = true
		} else if gap > mergeGapThreshold {
			// Rule 1 (inverse): gap exceeds merge threshold → new paragraph
			newParagraph = true
		}

		if newParagraph {
			paragraphs = append(paragraphs, Paragraph{
				Begin: currentBegin,
				Text:  strings.TrimSpace(currentText),
			})
			currentText = c.Text
			currentBegin = c.Begin
		} else {
			// Same paragraph — append with space
			currentText += " " + c.Text
		}
		prevEnd = c.End
	}

	// Don't forget the last paragraph
	if strings.TrimSpace(currentText) != "" {
		paragraphs = append(paragraphs, Paragraph{
			Begin: currentBegin,
			Text:  strings.TrimSpace(currentText),
		})
	}

	return paragraphs
}

// endsWithSentence returns true if the text ends with sentence-ending punctuation.
func endsWithSentence(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	last := s[len(s)-1]
	return last == '.' || last == '?' || last == '!'
}

// ── Cue → Sentence Splitting ────────────────────────────────────────────────

// timedWord associates a word with the timestamp of the cue it came from.
type timedWord struct {
	word  string
	begin float64
	end   float64
}

// commonAbbreviations is a set of words ending in '.' that aren't sentence endings.
var commonAbbreviations = map[string]bool{
	"mr.": true, "mrs.": true, "ms.": true, "dr.": true, "sr.": true, "jr.": true,
	"vs.": true, "etc.": true, "e.g.": true, "i.e.": true, "st.": true, "inc.": true,
	"ltd.": true, "gen.": true, "pres.": true, "gov.": true, "prof.": true, "no.": true,
}

// isSentenceEnding checks whether a word marks the end of a sentence.
func isSentenceEnding(word string) bool {
	// Strip trailing quotes/parens to find the final punctuation
	stripped := strings.TrimRight(word, `"')]`)
	if stripped == "" {
		return false
	}
	last := stripped[len(stripped)-1]
	if last != '.' && last != '?' && last != '!' {
		return false
	}
	// Check common abbreviations
	lower := strings.ToLower(stripped)
	if commonAbbreviations[lower] {
		return false
	}
	// Single letter + period (initials like "A." or "J.")
	core := strings.TrimLeft(stripped, `"'([`)
	if len(core) == 2 && core[1] == '.' {
		return false
	}
	// Ellipsis "..." is a pause, not necessarily a sentence end, but treat it as one
	return true
}

// MergeCuesIntoSentences splits deduplicated cues into sentence-level segments.
//
// Each sentence gets the timestamp of the cue where it begins. Sentences are
// identified by terminal punctuation (. ? !). If a long run of text has no
// punctuation (common with auto-captions), it falls back to splitting at
// cue gaps > 1.5 seconds.
func MergeCuesIntoSentences(cues []Cue) []Sentence {
	if len(cues) == 0 {
		return nil
	}

	// Step 1: Build a stream of timed words
	var words []timedWord
	for _, c := range cues {
		ws := strings.Fields(c.Text)
		// Distribute timing across words proportionally (approximation)
		n := len(ws)
		if n == 0 {
			continue
		}
		dur := c.End - c.Begin
		for j, w := range ws {
			// Each word gets a fraction of the cue's time span
			wordBegin := c.Begin + dur*float64(j)/float64(n)
			wordEnd := c.Begin + dur*float64(j+1)/float64(n)
			words = append(words, timedWord{word: w, begin: wordBegin, end: wordEnd})
		}
	}

	if len(words) == 0 {
		return nil
	}

	// Step 2: Walk words, split at sentence boundaries
	var sentences []Sentence
	var current []string
	sentBegin := words[0].begin
	sentEnd := words[0].end

	for _, tw := range words {
		current = append(current, tw.word)
		sentEnd = tw.end

		if isSentenceEnding(tw.word) {
			sentences = append(sentences, Sentence{
				Begin: sentBegin,
				End:   sentEnd,
				Text:  strings.Join(current, " "),
			})
			current = nil
			// Next word will set sentBegin
		}

		// Update sentBegin for next sentence
		if len(current) == 0 {
			sentBegin = tw.end // will be overwritten by next word
		}
		if len(current) == 1 {
			sentBegin = tw.begin
		}
	}

	// Flush remaining words
	if len(current) > 0 {
		sentences = append(sentences, Sentence{
			Begin: sentBegin,
			End:   sentEnd,
			Text:  strings.Join(current, " "),
		})
	}

	// Step 3: For unpunctuated auto-captions, try splitting long sentences at gaps
	// If we ended up with very few sentences relative to the number of cues,
	// the captions probably lack punctuation — re-split using time gaps.
	if len(sentences) <= len(cues)/10 && len(cues) > 10 {
		return splitByTimeGaps(cues)
	}

	return sentences
}

// splitByTimeGaps is a fallback for unpunctuated auto-captions.
// It merges cues but splits at gaps > 1.5s, producing short segments.
func splitByTimeGaps(cues []Cue) []Sentence {
	var sentences []Sentence
	var current []string
	sentBegin := cues[0].Begin
	prevEnd := cues[0].End

	for i, c := range cues {
		if i > 0 {
			gap := c.Begin - prevEnd
			if gap > mergeGapThreshold && len(current) > 0 {
				sentences = append(sentences, Sentence{
					Begin: sentBegin,
					End:   prevEnd,
					Text:  strings.TrimSpace(strings.Join(current, " ")),
				})
				current = nil
				sentBegin = c.Begin
			}
		}
		current = append(current, c.Text)
		prevEnd = c.End
	}

	if len(current) > 0 {
		sentences = append(sentences, Sentence{
			Begin: sentBegin,
			End:   prevEnd,
			Text:  strings.TrimSpace(strings.Join(current, " ")),
		})
	}

	return sentences
}

// sentenceParagraphGap is the time gap (seconds) between sentences that triggers a visual paragraph break.
const sentenceParagraphGap = 3.0

// ── Markdown Generation ─────────────────────────────────────────────────────

// GenerateTranscriptMarkdown creates the readable transcript.md content.
// Each sentence gets its own clickable [M:SS](url&t=N) timestamp link.
// Sentences are grouped into visual paragraphs by time gaps.
func GenerateTranscriptMarkdown(meta *VideoMetadata, sentences []Sentence) string {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "# %s\n\n", meta.Title)
	fmt.Fprintf(&b, "**Channel:** %s\n", meta.Channel)
	fmt.Fprintf(&b, "**Date:** %s\n", formatDate(meta.UploadDate))
	fmt.Fprintf(&b, "**Duration:** %s\n", formatDuration(meta.Duration))
	fmt.Fprintf(&b, "**URL:** %s\n", meta.URL)
	b.WriteString("\n---\n\n")
	b.WriteString("## Transcript\n\n")

	// Sentences with timestamp links, grouped into visual paragraphs by time gaps
	for i, s := range sentences {
		secs := int(math.Floor(s.Begin))
		ts := formatTimestamp(s.Begin)
		fmt.Fprintf(&b, "[%s](%s&t=%d) %s\n", ts, meta.URL, secs, s.Text)

		// Insert paragraph break if there's a big gap before the next sentence
		if i+1 < len(sentences) {
			gap := sentences[i+1].Begin - s.End
			if gap > sentenceParagraphGap {
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	return b.String()
}

// GenerateTranscriptMarkdownLegacy creates transcript.md using paragraph-level timestamps.
// Kept for reference; the main pipeline now uses sentence-level timestamps.
func GenerateTranscriptMarkdownLegacy(meta *VideoMetadata, paragraphs []Paragraph) string {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "# %s\n\n", meta.Title)
	fmt.Fprintf(&b, "**Channel:** %s\n", meta.Channel)
	fmt.Fprintf(&b, "**Date:** %s\n", formatDate(meta.UploadDate))
	fmt.Fprintf(&b, "**Duration:** %s\n", formatDuration(meta.Duration))
	fmt.Fprintf(&b, "**URL:** %s\n", meta.URL)
	b.WriteString("\n---\n\n")
	b.WriteString("## Transcript\n\n")

	// Paragraphs with timestamp links
	for _, p := range paragraphs {
		secs := int(math.Floor(p.Begin))
		ts := formatTimestamp(p.Begin)
		fmt.Fprintf(&b, "[%s](%s&t=%d) %s\n\n", ts, meta.URL, secs, p.Text)
	}

	return b.String()
}

// formatTimestamp formats seconds as M:SS or H:MM:SS.
func formatTimestamp(seconds float64) string {
	total := int(math.Floor(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// formatDate converts YYYYMMDD to YYYY-MM-DD.
func formatDate(d string) string {
	if len(d) == 8 {
		return d[:4] + "-" + d[4:6] + "-" + d[6:8]
	}
	return d
}

// formatDuration converts seconds to M:SS or H:MM:SS.
func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// ── cues.json Export ─────────────────────────────────────────────────────────

// ExportCuesJSON writes the raw cues to a JSON file for fine-grained timestamp lookups.
func ExportCuesJSON(path string, cues []Cue) error {
	data, err := json.MarshalIndent(cues, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cues: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ── Full Pipeline ────────────────────────────────────────────────────────────

// ProcessSubtitles runs the full pipeline for any supported subtitle format:
// detect format → parse → dedup → merge → generate markdown + cues.json
func ProcessSubtitles(subPath string, meta *VideoMetadata, outputDir string) (string, error) {
	ext := strings.ToLower(filepath.Ext(subPath))

	var rawCues []Cue
	var err error

	switch ext {
	case ".ttml":
		rawCues, err = ParseTTMLFile(subPath)
	case ".vtt":
		rawCues, err = ParseVTTFile(subPath)
	default:
		return "", fmt.Errorf("unsupported subtitle format: %s (expected .ttml or .vtt)", ext)
	}

	if err != nil {
		return "", fmt.Errorf("parsing subtitles (%s): %w", ext, err)
	}
	if len(rawCues) == 0 {
		return "", fmt.Errorf("no cues found in subtitle file")
	}

	// Deduplicate
	cues := DeduplicateCues(rawCues)
	if len(cues) == 0 {
		return "", fmt.Errorf("all cues were duplicates")
	}

	// Export cues.json (raw deduped cues)
	cuesPath := filepath.Join(outputDir, "cues.json")
	if err := ExportCuesJSON(cuesPath, cues); err != nil {
		return "", fmt.Errorf("exporting cues.json: %w", err)
	}

	// Split into sentences (sentence-level timestamps)
	sentences := MergeCuesIntoSentences(cues)

	// Generate markdown
	md := GenerateTranscriptMarkdown(meta, sentences)

	// Write transcript.md
	mdPath := filepath.Join(outputDir, "transcript.md")
	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		return "", fmt.Errorf("writing transcript.md: %w", err)
	}

	return md, nil
}

// ProcessTTML runs the full pipeline for TTML files (backward compatibility wrapper).
func ProcessTTML(ttmlPath string, meta *VideoMetadata, outputDir string) (string, error) {
	return ProcessSubtitles(ttmlPath, meta, outputDir)
}
