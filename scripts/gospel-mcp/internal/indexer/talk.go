package indexer

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/urlgen"
)

// Talk filename pattern: 57nelson.md, 23bednar.md, etc.
var talkFilenamePattern = regexp.MustCompile(`^(\d+)([a-z-]+)\.md$`)

// Title pattern: # Title
var titlePattern = regexp.MustCompile(`^#\s+(.+)$`)

// Speaker pattern: By Elder/President Name
var speakerPattern = regexp.MustCompile(`(?i)^By\s+(Elder|President|Sister|Bishop)?\s*(.+)$`)

func (idx *Indexer) indexTalkFile(path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(relPath, string(os.PathSeparator))

	// Expecting: gospel-library/eng/general-conference/{year}/{month}/{talk}.md
	if len(parts) < 6 {
		return nil
	}

	yearStr := parts[3]
	monthStr := parts[4]
	filename := filepath.Base(path)

	// Parse year and month
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil // Skip non-year directories
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return nil // Skip non-month directories
	}

	// Skip index files and non-talk files
	if !talkFilenamePattern.MatchString(filename) {
		return nil
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fullContent := string(content)

	// Extract title and speaker
	title, speaker := extractTalkMetadata(fullContent)
	if title == "" {
		title = strings.TrimSuffix(filename, ".md")
	}
	if speaker == "" {
		// Try to extract from filename
		if matches := talkFilenamePattern.FindStringSubmatch(filename); matches != nil {
			speaker = formatSpeakerName(matches[2])
		}
	}

	// Generate source URL
	sourceURL := urlgen.Talk(year, month, filename)

	// Insert talk record
	_, err = idx.db.Exec(`
		INSERT OR REPLACE INTO talks (year, month, speaker, title, content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, year, month, speaker, title, fullContent, relPath, sourceURL)
	if err != nil {
		return err
	}

	result.TalksIndexed++

	// Record metadata
	return idx.recordMetadata(path, info, "talk", 1)
}

func extractTalkMetadata(content string) (title, speaker string) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Find title
		if title == "" {
			if matches := titlePattern.FindStringSubmatch(line); matches != nil {
				title = strings.TrimSpace(matches[1])
				continue
			}
		}

		// Find speaker (usually "By Elder Name" or "By President Name")
		if speaker == "" && strings.HasPrefix(strings.ToLower(line), "by ") {
			if matches := speakerPattern.FindStringSubmatch(line); matches != nil {
				speaker = strings.TrimSpace(matches[2])
				// Clean up any trailing role info
				if idx := strings.Index(speaker, "\n"); idx != -1 {
					speaker = speaker[:idx]
				}
			}
		}

		// Stop after we have both
		if title != "" && speaker != "" {
			break
		}
	}

	return title, speaker
}

func formatSpeakerName(slug string) string {
	// Convert "nelson" to "Nelson", "christofferson" to "Christofferson"
	if len(slug) == 0 {
		return ""
	}
	return strings.ToUpper(slug[:1]) + slug[1:]
}
