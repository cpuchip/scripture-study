package indexer

import (
	"bufio"
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

var scriptureVolumes = map[string]bool{
	"ot": true, "nt": true, "bofm": true, "dc-testament": true, "pgp": true,
}
var studyAids = map[string]bool{
	"tg": true, "bd": true, "gs": true, "jst": true,
}

var verseRe = regexp.MustCompile(`^\*\*(\d+)\.\*\*\s*(.+)`)
var footnoteRe = regexp.MustCompile(`<sup>\[([^\]]+)\]\(#fn-[^)]+\)</sup>`)
var superscriptRe = regexp.MustCompile(`<sup>[^<]+</sup>`)
var linkRe = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
var crossRefLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+\.md[^)]*)\)`)
var footnoteAnchorRe = regexp.MustCompile(`<a id="fn-(\d+)[a-z]+"?>`)

func (idx *Indexer) indexScriptureFile(ctx context.Context, path string, info os.FileInfo, opts Options, result *Result) error {
	relPath, _ := filepath.Rel(idx.root, path)
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	// gospel-library/eng/scriptures/{volume}/{book}/{chapter}.md
	if len(parts) < 6 {
		return nil
	}

	volume := parts[3]
	book := parts[4]

	if studyAids[volume] || !scriptureVolumes[volume] {
		return nil
	}

	filename := filepath.Base(path)
	chapterStr := strings.TrimSuffix(filename, ".md")
	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	fullContent := string(content)

	// Extract title
	var title string
	scanner := bufio.NewScanner(strings.NewReader(fullContent))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Insert chapter
	chapterURL := urlgen.ScriptureChapter(volume, book, chapter)
	if _, err := idx.db.Exec(`
		INSERT OR REPLACE INTO chapters (volume, book, chapter, title, full_content, file_path, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, volume, book, chapter, title, fullContent, relPath, chapterURL); err != nil {
		return fmt.Errorf("inserting chapter: %w", err)
	}
	result.ChaptersIndexed++

	// Parse verses
	verses := parseVerses(fullContent)
	var vecChunks []vec.Chunk

	for _, v := range verses {
		verseURL := urlgen.Scripture(volume, book, chapter, v.Number)
		if _, err := idx.db.Exec(`
			INSERT OR REPLACE INTO scriptures (volume, book, chapter, verse, text, file_path, source_url)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, volume, book, chapter, v.Number, v.Text, relPath, verseURL); err != nil {
			return fmt.Errorf("inserting verse %d: %w", v.Number, err)
		}
		result.ScripturesIndexed++

		// Cross-references
		crossRefs := extractCrossReferences(fullContent, volume, book, chapter, v.Number)
		for _, ref := range crossRefs {
			if _, err := idx.db.Exec(`
				INSERT OR IGNORE INTO cross_references
				(source_volume, source_book, source_chapter, source_verse,
				 target_volume, target_book, target_chapter, target_verse, reference_type)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, volume, book, chapter, v.Number,
				ref.TargetVolume, ref.TargetBook, ref.TargetChapter, ref.TargetVerse, ref.Type); err == nil {
				result.CrossRefsIndexed++
			}
		}
	}

	// Build vector chunks
	bookName := FormatBookName(book)
	for _, layer := range opts.Layers {
		switch layer {
		case vec.LayerVerse:
			for _, v := range verses {
				ref := fmt.Sprintf("%s %d:%d", bookName, chapter, v.Number)
				vecChunks = append(vecChunks, vec.Chunk{
					ID:      fmt.Sprintf("scriptures-%s-%d-%d", slugify(bookName), chapter, v.Number),
					Content: v.Text,
					Metadata: &vec.DocMetadata{
						Source:    vec.SourceScriptures,
						Layer:     vec.LayerVerse,
						Book:      bookName,
						Chapter:   chapter,
						Reference: ref,
						Range:     fmt.Sprintf("%d", v.Number),
						FilePath:  relPath,
					},
				})
			}
		case vec.LayerParagraph:
			paragraphs := chunkByParagraph(verses, bookName, chapter, relPath)
			vecChunks = append(vecChunks, paragraphs...)
		}
	}

	if err := idx.addVecChunks(ctx, vecChunks, opts); err != nil {
		return fmt.Errorf("adding vector chunks: %w", err)
	}
	result.VecChunksAdded += len(vecChunks)

	return idx.recordMetadata(path, info, "scripture", len(verses))
}

type parsedVerse struct {
	Number int
	Text   string
}

func parseVerses(content string) []parsedVerse {
	var verses []parsedVerse
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := verseRe.FindStringSubmatch(line); matches != nil {
			num, _ := strconv.Atoi(matches[1])
			text := cleanVerseText(matches[2])
			if text != "" {
				verses = append(verses, parsedVerse{Number: num, Text: text})
			}
		}
	}
	return verses
}

func cleanVerseText(text string) string {
	text = footnoteRe.ReplaceAllString(text, "")
	text = superscriptRe.ReplaceAllString(text, "")
	text = linkRe.ReplaceAllString(text, "$1")
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func chunkByParagraph(verses []parsedVerse, book string, chapter int, filePath string) []vec.Chunk {
	if len(verses) == 0 {
		return nil
	}
	var chunks []vec.Chunk
	chunkSize := 4
	overlap := 1

	for i := 0; i < len(verses); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(verses) {
			end = len(verses)
		}

		var content strings.Builder
		var nums []int
		for j := i; j < end; j++ {
			if content.Len() > 0 {
				content.WriteString(" ")
			}
			content.WriteString(fmt.Sprintf("(%d) %s", verses[j].Number, verses[j].Text))
			nums = append(nums, verses[j].Number)
		}

		startV := nums[0]
		endV := nums[len(nums)-1]
		ref := fmt.Sprintf("%s %d:%d-%d", book, chapter, startV, endV)
		rangeStr := fmt.Sprintf("%d-%d", startV, endV)

		chunks = append(chunks, vec.Chunk{
			ID:      fmt.Sprintf("scriptures-%s-%d-p%d-%d", slugify(book), chapter, startV, endV),
			Content: content.String(),
			Metadata: &vec.DocMetadata{
				Source:    vec.SourceScriptures,
				Layer:     vec.LayerParagraph,
				Book:      book,
				Chapter:   chapter,
				Reference: ref,
				Range:     rangeStr,
				FilePath:  filePath,
			},
		})

		if end >= len(verses) {
			break
		}
	}
	return chunks
}

type crossRef struct {
	TargetVolume  string
	TargetBook    string
	TargetChapter int
	TargetVerse   *int
	Type          string
}

func extractCrossReferences(content string, sourceVolume, sourceBook string, sourceChapter, sourceVerse int) []crossRef {
	var refs []crossRef

	footnoteSection := ""
	if idx := strings.Index(content, "## Footnotes"); idx != -1 {
		footnoteSection = content[idx:]
	}
	if footnoteSection == "" {
		return refs
	}

	scanner := bufio.NewScanner(strings.NewReader(footnoteSection))
	for scanner.Scan() {
		line := scanner.Text()
		anchorMatch := footnoteAnchorRe.FindStringSubmatch(line)
		if anchorMatch == nil {
			continue
		}
		fnVerse, err := strconv.Atoi(anchorMatch[1])
		if err != nil || fnVerse != sourceVerse {
			continue
		}

		matches := crossRefLinkRe.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			if ref := parseCrossRefLink(match[2]); ref != nil {
				refs = append(refs, *ref)
			}
		}
	}
	return refs
}

func parseCrossRefLink(linkPath string) *crossRef {
	parts := strings.Split(linkPath, "/")

	volIdx := -1
	for i, p := range parts {
		if scriptureVolumes[p] || studyAids[p] {
			volIdx = i
			break
		}
	}
	if volIdx == -1 || volIdx+1 >= len(parts) {
		return nil
	}

	volume := parts[volIdx]
	book := parts[volIdx+1]

	if studyAids[volume] {
		return &crossRef{
			TargetVolume: volume,
			TargetBook:   strings.TrimSuffix(book, ".md"),
			Type:         volume,
		}
	}

	if volIdx+2 >= len(parts) {
		return nil
	}

	chapterFile := parts[volIdx+2]
	chapterStr := strings.TrimSuffix(chapterFile, ".md")

	var verseNum *int
	if idx := strings.Index(chapterStr, "#"); idx != -1 {
		anchor := chapterStr[idx+1:]
		chapterStr = chapterStr[:idx]
		if strings.HasPrefix(anchor, "p") {
			if v, err := strconv.Atoi(anchor[1:]); err == nil {
				verseNum = &v
			}
		}
	}

	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		return nil
	}

	return &crossRef{
		TargetVolume:  volume,
		TargetBook:    book,
		TargetChapter: chapter,
		TargetVerse:   verseNum,
		Type:          "footnote",
	}
}
