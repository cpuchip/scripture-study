package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TalkMetadata contains extracted metadata from a conference talk
type TalkMetadata struct {
	Title       string
	Speaker     string
	Position    string
	Year        string
	Month       string
	Session     string // e.g., "Saturday Morning", "Sunday Afternoon"
	SessionCode string // e.g., "57" from "57nelson.md"
	AudioURL    string
	ImageURL    string
	Summary     string // Opening summary/quote if present
	FilePath    string
}

// ParsedTalk contains the full parsed content of a conference talk
type ParsedTalk struct {
	Metadata   TalkMetadata
	Sections   []TalkSection
	Paragraphs []string
	Footnotes  []string
	RawContent string
}

// TalkSection represents a section of the talk (H2 heading + content)
type TalkSection struct {
	Heading string
	Content string
}

// Patterns for parsing talks
var (
	talkTitlePattern    = regexp.MustCompile(`^#\s+(.+)$`)
	talkAudioPattern    = regexp.MustCompile(`🎧\s*\[Listen to Audio\]\(([^)]+)\)`)
	talkSpeakerPattern  = regexp.MustCompile(`^By\s+(.+)$`)
	talkImagePattern    = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)]+)\)`)
	talkSectionPattern  = regexp.MustCompile(`^##\s+(.+)$`)
	talkFootnotePattern = regexp.MustCompile(`<a\s+id="fn-(\d+)"`)
	talkSessionPattern  = regexp.MustCompile(`^(\d+)([a-z]+)\.md$`) // e.g., 57nelson.md
	talkScriptureRef    = regexp.MustCompile(`\[([^\]]+)\]\([^)]*scriptures[^)]+\)`)
)

// Common position keywords to help identify calling line
var positionKeywords = []string{
	"Quorum of the Twelve",
	"First Presidency",
	"Seventy",
	"Presiding Bishop",
	"Relief Society",
	"Young Women",
	"Young Men",
	"Primary",
	"Sunday School",
	"Church Historian",
	"President of the Church",
	"Apostle",
}

// ParseTalkFile parses a conference talk markdown file
func ParseTalkFile(filePath string) (*ParsedTalk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	talk := &ParsedTalk{}
	talk.Metadata.FilePath = filePath

	// Extract year/month from path
	// e.g., .../general-conference/2025/04/57nelson.md
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for i, part := range parts {
		if part == "general-conference" && i+2 < len(parts) {
			talk.Metadata.Year = parts[i+1]
			talk.Metadata.Month = parts[i+2]
			break
		}
	}

	// Extract session code from filename if numeric format
	filename := filepath.Base(filePath)
	if matches := talkSessionPattern.FindStringSubmatch(filename); matches != nil {
		talk.Metadata.SessionCode = matches[1]
		// Derive session name from code (first digit)
		if len(matches[1]) >= 1 {
			sessionNum := matches[1][0]
			switch sessionNum {
			case '1':
				talk.Metadata.Session = "Saturday Morning"
			case '2':
				talk.Metadata.Session = "Saturday Afternoon"
			case '3':
				talk.Metadata.Session = "Priesthood"
			case '4':
				talk.Metadata.Session = "Sunday Morning"
			case '5':
				talk.Metadata.Session = "Sunday Afternoon"
			case '6':
				talk.Metadata.Session = "Women's"
			}
		}
	}

	// Read and parse content
	scanner := bufio.NewScanner(file)
	var lines []string
	var rawContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		rawContent.WriteString(line)
		rawContent.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file: %w", err)
	}

	talk.RawContent = rawContent.String()

	// Parse metadata from first ~15 lines
	titleFound := false
	speakerFound := false
	imageFound := false
	var bodyStartLine int

	for i, line := range lines {
		if i > 30 { // Don't search too far
			bodyStartLine = i
			break
		}

		// Title (first H1, skip duplicates)
		if !titleFound {
			if matches := talkTitlePattern.FindStringSubmatch(line); matches != nil {
				talk.Metadata.Title = matches[1]
				titleFound = true
				continue
			}
		}

		// Audio URL
		if matches := talkAudioPattern.FindStringSubmatch(line); matches != nil {
			talk.Metadata.AudioURL = matches[1]
			continue
		}

		// Speaker
		if !speakerFound {
			if matches := talkSpeakerPattern.FindStringSubmatch(line); matches != nil {
				talk.Metadata.Speaker = matches[1]
				speakerFound = true
				// Next non-empty line is likely the position
				for j := i + 1; j < len(lines) && j < i+5; j++ {
					posLine := strings.TrimSpace(lines[j])
					if posLine != "" && !strings.HasPrefix(posLine, "!") && !strings.HasPrefix(posLine, "#") {
						if isPositionLine(posLine) {
							talk.Metadata.Position = posLine
							break
						}
					}
				}
				continue
			}
		}

		// Image
		if !imageFound {
			if matches := talkImagePattern.FindStringSubmatch(line); matches != nil {
				talk.Metadata.ImageURL = matches[2]
				imageFound = true
				continue
			}
		}

		// Check for opening summary (italic line after image, before body)
		if imageFound && talk.Metadata.Summary == "" {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "!") {
				// Check if it looks like a summary (short, possibly italicized)
				if len(trimmed) < 200 && !strings.HasPrefix(trimmed, "By ") {
					talk.Metadata.Summary = trimmed
					bodyStartLine = i + 1
					break
				}
			}
		}
	}

	// Parse body content
	var currentSection *TalkSection
	var currentParagraph strings.Builder
	inFootnotes := false

	for i := bodyStartLine; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check for footnotes section
		if strings.Contains(line, "---") || strings.HasPrefix(trimmed, "<a id=\"fn-") {
			inFootnotes = true
		}

		if inFootnotes {
			if matches := talkFootnotePattern.FindStringSubmatch(line); matches != nil {
				talk.Footnotes = append(talk.Footnotes, trimmed)
			}
			continue
		}

		// Section heading (H2)
		if matches := talkSectionPattern.FindStringSubmatch(line); matches != nil {
			// Save previous section
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(currentSection.Content)
				talk.Sections = append(talk.Sections, *currentSection)
			}
			currentSection = &TalkSection{Heading: matches[1]}
			continue
		}

		// Skip audio, image, duplicate titles
		if strings.HasPrefix(trimmed, "🎧") || strings.HasPrefix(trimmed, "![") {
			continue
		}
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "By ") {
			continue
		}

		// Regular paragraph
		if trimmed == "" {
			if currentParagraph.Len() > 0 {
				para := strings.TrimSpace(currentParagraph.String())
				if para != "" {
					talk.Paragraphs = append(talk.Paragraphs, para)
					if currentSection != nil {
						currentSection.Content += para + "\n\n"
					}
				}
				currentParagraph.Reset()
			}
		} else {
			if currentParagraph.Len() > 0 {
				currentParagraph.WriteString(" ")
			}
			currentParagraph.WriteString(trimmed)
		}
	}

	// Don't forget last paragraph
	if currentParagraph.Len() > 0 {
		para := strings.TrimSpace(currentParagraph.String())
		if para != "" {
			talk.Paragraphs = append(talk.Paragraphs, para)
			if currentSection != nil {
				currentSection.Content += para + "\n\n"
			}
		}
	}

	// Save last section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(currentSection.Content)
		talk.Sections = append(talk.Sections, *currentSection)
	}

	return talk, nil
}

// isPositionLine checks if a line looks like a Church position/calling
func isPositionLine(line string) bool {
	lineLower := strings.ToLower(line)
	for _, keyword := range positionKeywords {
		if strings.Contains(lineLower, strings.ToLower(keyword)) {
			return true
		}
	}
	// Also check for patterns like "Of the ..." or ending with "General Authority"
	if strings.HasPrefix(line, "Of the ") || strings.HasPrefix(line, "of the ") {
		return true
	}
	return false
}

// ExtractScriptureReferences finds all scripture references in the talk
func ExtractScriptureReferences(content string) []string {
	matches := talkScriptureRef.FindAllStringSubmatch(content, -1)
	refs := make([]string, 0, len(matches))
	seen := make(map[string]bool)
	for _, match := range matches {
		ref := match[1]
		if !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
	}
	return refs
}

// FindTalkFiles finds all conference talk files in a directory
func FindTalkFiles(basePath string, years ...string) ([]string, error) {
	var files []string

	// If no years specified, find all
	if len(years) == 0 {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				year := entry.Name()
				// Check if it's a year directory (4 digits)
				if len(year) == 4 && year[0] >= '1' && year[0] <= '2' {
					years = append(years, year)
				}
			}
		}
	}

	for _, year := range years {
		yearPath := filepath.Join(basePath, year)

		// Check both April (04) and October (10) conferences
		for _, month := range []string{"04", "10"} {
			monthPath := filepath.Join(yearPath, month)
			if _, err := os.Stat(monthPath); os.IsNotExist(err) {
				continue
			}

			entries, err := os.ReadDir(monthPath)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
					// Skip non-talk files
					name := entry.Name()
					if strings.Contains(name, "statistical") ||
						strings.Contains(name, "audit") ||
						strings.Contains(name, "sustaining") {
						continue
					}
					files = append(files, filepath.Join(monthPath, name))
				}
			}
		}
	}

	return files, nil
}

// GetTalkReference returns a formatted reference string for the talk
func (t *ParsedTalk) GetTalkReference() string {
	monthName := "April"
	if t.Metadata.Month == "10" {
		monthName = "October"
	}
	return fmt.Sprintf("%s %s General Conference - %s", monthName, t.Metadata.Year, t.Metadata.Speaker)
}

// GetTalkID returns a unique identifier for the talk
func (t *ParsedTalk) GetTalkID() string {
	filename := filepath.Base(t.Metadata.FilePath)
	name := strings.TrimSuffix(filename, ".md")
	return fmt.Sprintf("gc-%s-%s-%s", t.Metadata.Year, t.Metadata.Month, name)
}

// ChunkTalkByParagraph creates paragraph-level chunks from a talk
func ChunkTalkByParagraph(talk *ParsedTalk) []Chunk {
	var chunks []Chunk
	timestamp := time.Now().Format(time.RFC3339)

	year, _ := strconv.Atoi(talk.Metadata.Year)

	for i, para := range talk.Paragraphs {
		if len(strings.TrimSpace(para)) < 20 {
			continue // Skip very short paragraphs
		}

		chunk := Chunk{
			ID:      fmt.Sprintf("%s-p%d", talk.GetTalkID(), i+1),
			Content: para,
			Metadata: &DocMetadata{
				Source:    SourceConference,
				Layer:     LayerParagraph,
				Reference: fmt.Sprintf("%s, paragraph %d", talk.GetTalkReference(), i+1),
				FilePath:  talk.Metadata.FilePath,
				Timestamp: timestamp,
				Speaker:   talk.Metadata.Speaker,
				Position:  talk.Metadata.Position,
				Year:      year,
				Month:     talk.Metadata.Month,
				Session:   talk.Metadata.Session,
				TalkTitle: talk.Metadata.Title,
			},
		}
		chunks = append(chunks, chunk)
	}

	return chunks
}

// ChunkTalkAsSummary creates a summary chunk for the entire talk
func ChunkTalkAsSummary(talk *ParsedTalk, summary *ChapterSummary, model string) Chunk {
	year, _ := strconv.Atoi(talk.Metadata.Year)
	timestamp := time.Now().Format(time.RFC3339)

	// Build summary text including keywords
	summaryText := summary.Summary
	if summary.KeyVerse != "" {
		summaryText += "\n\nKey Quote: " + summary.KeyVerse
	}
	if len(summary.Keywords) > 0 {
		summaryText += "\n\nTopics: " + strings.Join(summary.Keywords, ", ")
	}

	return Chunk{
		ID:      fmt.Sprintf("%s-summary", talk.GetTalkID()),
		Content: summaryText,
		Metadata: &DocMetadata{
			Source:    SourceConference,
			Layer:     LayerSummary,
			Reference: talk.GetTalkReference(),
			FilePath:  talk.Metadata.FilePath,
			Generated: true,
			Model:     model,
			Timestamp: timestamp,
			Speaker:   talk.Metadata.Speaker,
			Position:  talk.Metadata.Position,
			Year:      year,
			Month:     talk.Metadata.Month,
			Session:   talk.Metadata.Session,
			TalkTitle: talk.Metadata.Title,
		},
	}
}

// IsAdministrativeDocument returns true if the talk should be skipped
func (t *ParsedTalk) IsAdministrativeDocument() bool {
	title := strings.ToLower(t.Metadata.Title)
	return strings.Contains(title, "sustaining") ||
		strings.Contains(title, "statistical") ||
		strings.Contains(title, "audit") ||
		strings.Contains(title, "church auditing") ||
		t.Metadata.Speaker == ""
}
