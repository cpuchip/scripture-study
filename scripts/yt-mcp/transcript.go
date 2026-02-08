package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
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
					c.Text = newPart
				} else {
					continue
				}
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

// ── Markdown Generation ─────────────────────────────────────────────────────

// GenerateTranscriptMarkdown creates the readable transcript.md content.
// Each paragraph is preceded by a clickable [M:SS](url&t=N) timestamp link.
func GenerateTranscriptMarkdown(meta *VideoMetadata, paragraphs []Paragraph) string {
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

// ProcessTTML runs the full pipeline: parse → dedup → merge → generate markdown + cues.json
func ProcessTTML(ttmlPath string, meta *VideoMetadata, outputDir string) (string, error) {
	// 1. Parse TTML
	rawCues, err := ParseTTMLFile(ttmlPath)
	if err != nil {
		return "", fmt.Errorf("parsing TTML: %w", err)
	}
	if len(rawCues) == 0 {
		return "", fmt.Errorf("no cues found in TTML file")
	}

	// 2. Deduplicate
	cues := DeduplicateCues(rawCues)
	if len(cues) == 0 {
		return "", fmt.Errorf("all cues were duplicates")
	}

	// 3. Export cues.json (raw deduped cues)
	cuesPath := fmt.Sprintf("%s/cues.json", outputDir)
	if err := ExportCuesJSON(cuesPath, cues); err != nil {
		return "", fmt.Errorf("exporting cues.json: %w", err)
	}

	// 4. Merge into paragraphs
	paragraphs := MergeCuesIntoParagraphs(cues)

	// 5. Generate markdown
	md := GenerateTranscriptMarkdown(meta, paragraphs)

	// 6. Write transcript.md
	mdPath := fmt.Sprintf("%s/transcript.md", outputDir)
	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		return "", fmt.Errorf("writing transcript.md: %w", err)
	}

	return md, nil
}
