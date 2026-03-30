package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/urlgen"
	"github.com/cpuchip/scripture-study/scripts/gospel-engine/internal/vec"
)

// digitPrefixRe matches newer talk filenames like 57nelson.md (2019+).
// Used only as a fallback for speaker extraction from filename.
var digitPrefixRe = regexp.MustCompile(`^(\d+)([a-z-]+)\.md$`)

func (idx *Indexer) indexTalkFile(ctx context.Context, path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	// gospel-library/eng/general-conference/{year}/{month}/{talk}.md
	if len(parts) < 6 {
		return nil
	}

	yearStr := parts[3]
	monthStr := parts[4]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fullContent := string(content)

	filename := filepath.Base(path)
	title, speaker := extractTalkMetadata(fullContent)
	if title == "" {
		title = strings.TrimSuffix(filename, ".md")
	}
	if speaker == "" {
		// Try to extract from digit-prefixed filenames (2019+): 57nelson.md → Nelson
		if matches := digitPrefixRe.FindStringSubmatch(filename); matches != nil {
			speaker = strings.ToUpper(matches[2][:1]) + matches[2][1:]
		}
	}

	sourceURL := urlgen.Talk(year, month, filename)

	if _, err := idx.db.Exec(`
		INSERT OR REPLACE INTO talks (year, month, speaker, title, content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, year, month, speaker, title, fullContent, relPath, sourceURL); err != nil {
		return err
	}
	result.TalksIndexed++

	// Vector chunks for talks — paragraph-level
	var vecChunks []vec.Chunk
	paragraphs := extractParagraphs(fullContent)
	if len(paragraphs) < 3 {
		// Skip too-short talks (administrative)
		return idx.recordMetadata(path, info, "talk", 1)
	}

	for _, layer := range opts.Layers {
		if layer == vec.LayerParagraph {
			for i, para := range paragraphs {
				ref := fmt.Sprintf("%s (%s, %d/%02d) ¶%d", speaker, title, year, month, i+1)
				vecChunks = append(vecChunks, vec.Chunk{
					ID:      fmt.Sprintf("conference-%d-%02d-%s-p%d", year, month, slugify(speaker), i+1),
					Content: para,
					Metadata: &vec.DocMetadata{
						Source:    vec.SourceConference,
						Layer:     vec.LayerParagraph,
						Book:      speaker,
						Reference: ref,
						FilePath:  relPath,
						Speaker:   speaker,
						TalkTitle: title,
						Year:      year,
						Month:     fmt.Sprintf("%02d", month),
					},
				})
			}
		}
	}

	if err := idx.addVecChunks(ctx, vecChunks, opts); err != nil {
		return fmt.Errorf("adding vector chunks: %w", err)
	}
	result.VecChunksAdded += len(vecChunks)

	return idx.recordMetadata(path, info, "talk", 1)
}

var titleRe = regexp.MustCompile(`^#\s+(.+)$`)
var speakerRe = regexp.MustCompile(`(?i)^By\s+(Elder|President|Sister|Bishop)?\s*(.+)$`)

func extractTalkMetadata(content string) (title, speaker string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if title == "" {
			if m := titleRe.FindStringSubmatch(line); m != nil {
				title = strings.TrimSpace(m[1])
				continue
			}
		}
		if speaker == "" && strings.HasPrefix(strings.ToLower(line), "by ") {
			if m := speakerRe.FindStringSubmatch(line); m != nil {
				speaker = strings.TrimSpace(m[2])
			}
		}
		if title != "" && speaker != "" {
			break
		}
	}
	return title, speaker
}

// extractParagraphs splits talk content into non-empty paragraphs.
func extractParagraphs(content string) []string {
	raw := strings.Split(content, "\n\n")
	var paragraphs []string
	for _, p := range raw {
		p = strings.TrimSpace(p)
		// Skip headings, metadata lines, empty
		if p == "" || strings.HasPrefix(p, "#") || strings.HasPrefix(p, "---") {
			continue
		}
		// Skip "By ..." lines
		if strings.HasPrefix(strings.ToLower(p), "by ") && len(p) < 100 {
			continue
		}
		paragraphs = append(paragraphs, p)
	}
	return paragraphs
}
