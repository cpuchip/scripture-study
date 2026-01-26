package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
)

type Page struct {
	URL      string
	FileName string
}

func main() {
	pages := []Page{
		{URL: "https://www.lecturesonfaith.com/", FileName: "00_introduction.md"},
		{URL: "https://www.lecturesonfaith.com/preface.html", FileName: "00_preface.md"},
		{URL: "https://www.lecturesonfaith.com/1", FileName: "01_lecture_1.md"},
		{URL: "https://www.lecturesonfaith.com/2", FileName: "02_lecture_2.md"},
		{URL: "https://www.lecturesonfaith.com/3", FileName: "03_lecture_3.md"},
		{URL: "https://www.lecturesonfaith.com/4", FileName: "04_lecture_4.md"},
		{URL: "https://www.lecturesonfaith.com/5", FileName: "05_lecture_5.md"},
		{URL: "https://www.lecturesonfaith.com/6", FileName: "06_lecture_6.md"},
		{URL: "https://www.lecturesonfaith.com/7", FileName: "07_lecture_7.md"},
	}

	outputDir := filepath.Join("books", "lecture-on-faith")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	for _, page := range pages {
		fmt.Printf("Fetching %s...\n", page.URL)
		markdown, title, err := fetchAndConvert(client, page.URL)
		if err != nil {
			fmt.Printf("  ✗ Failed: %v\n", err)
			continue
		}

		markdown = ensureTitle(markdown, title)
		outputPath := filepath.Join(outputDir, page.FileName)
		if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
			fmt.Printf("  ✗ Write failed: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ Wrote %s\n", outputPath)
	}
}

func fetchAndConvert(client *http.Client, url string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "ScriptureStudy-Downloader/1.0 (personal study tool)")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	main := doc.Find("main").First()
	if main.Length() == 0 {
		main = doc.Find("#content").First()
	}
	if main.Length() == 0 {
		main = doc.Find("body").First()
	}

	main.Find("script,style,nav,footer,header").Remove()

	title := strings.TrimSpace(main.Find("h1").First().Text())

	html, err := main.Html()
	if err != nil {
		return "", "", err
	}

	markdown, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", "", err
	}

	markdown = cleanupMarkdown(markdown)
	return markdown, title, nil
}

func cleanupMarkdown(markdown string) string {
	markdown = normalizeLectureLinks(markdown)
	markdown = localizeScriptureLinks(markdown)
	markdown = stripLectureNav(markdown)
	markdown = dedupeTitle(markdown)
	markdown = addAnchorTargets(markdown)
	markdown = regexp.MustCompile(`\n{3,}`).ReplaceAllString(markdown, "\n\n")
	markdown = strings.TrimSpace(markdown)
	return markdown
}

func localizeScriptureLinks(markdown string) string {
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	return linkRe.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := linkRe.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		text := parts[1]
		url := parts[2]
		if local, ok := convertScriptureURL(url); ok {
			return fmt.Sprintf("[%s](%s)", text, local)
		}
		return match
	})
}

func convertScriptureURL(url string) (string, bool) {
	urlRe := regexp.MustCompile(`^https?://(?:www\.)?(?:lds\.org|churchofjesuschrist\.org)/(?:study/)?scriptures/([^?#]+)`)
	parts := urlRe.FindStringSubmatch(url)
	if len(parts) != 2 {
		return "", false
	}

	pathParts := strings.Split(strings.Trim(parts[1], "/"), "/")
	if len(pathParts) < 3 {
		return "", false
	}

	collection := pathParts[0]
	book := pathParts[1]
	chapterSegment := pathParts[2]
	if collection == "" || book == "" || chapterSegment == "" {
		return "", false
	}

	chapter := strings.SplitN(chapterSegment, ".", 2)[0]
	if chapter == "" {
		return "", false
	}

	local := filepath.ToSlash(filepath.Join("..", "..", "gospel-library", "eng", "scriptures", collection, book, chapter+".md"))
	return local, true
}

func normalizeLectureLinks(markdown string) string {
	rePreface := regexp.MustCompile(`\((https?://www\.lecturesonfaith\.com)?/preface\.html/?\)`)
	markdown = rePreface.ReplaceAllString(markdown, "(00_preface.md)")

	reIntro := regexp.MustCompile(`\((https?://www\.lecturesonfaith\.com)?/\)`)
	markdown = reIntro.ReplaceAllString(markdown, "(00_introduction.md)")

	reLecture := regexp.MustCompile(`\((https?://www\.lecturesonfaith\.com)?/([1-7])/?\)`)
	markdown = reLecture.ReplaceAllString(markdown, "(0$2_lecture_$2.md)")

	reLectureMD := regexp.MustCompile(`\((0?[1-7])\.md\)`)
	markdown = reLectureMD.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := reLectureMD.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		num := parts[1]
		if len(num) == 1 {
			num = "0" + num
		}
		base := strings.TrimLeft(num, "0")
		if base == "" {
			return match
		}
		return fmt.Sprintf("(%s_lecture_%s.md)", num, base)
	})

	return markdown
}

func stripLectureNav(markdown string) string {
	// Remove lines that contain only navigation links between lectures/preface/intro
	lineRe := regexp.MustCompile(`(?m)^\s*(\[[^\]]*\]\((00_(preface|introduction)\.md|0[1-7]_lecture_[1-7]\.md)\)\s*)+\s*$`)
	markdown = lineRe.ReplaceAllString(markdown, "")
	return markdown
}

func dedupeTitle(markdown string) string {
	lines := strings.Split(markdown, "\n")
	firstTitle := ""
	var out []string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.HasPrefix(line, "# ") {
			title := normalizeTitle(strings.TrimPrefix(line, "# "))
			if firstTitle == "" {
				firstTitle = title
				out = append(out, line)
				continue
			}
			if title == firstTitle {
				// Skip duplicate title and a following blank line, if present
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
					i++
				}
				continue
			}
		}
		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

func normalizeTitle(title string) string {
	title = strings.ReplaceAll(title, "\u00a0", " ")
	return strings.TrimSpace(title)
}

func addAnchorTargets(markdown string) string {
	// Turn paragraph markers like [1](#1) into anchored numbers
	paraRe := regexp.MustCompile(`(?m)^\[(\d+)\]\(#\d+\)\s+`)
	markdown = paraRe.ReplaceAllString(markdown, `<a id="$1"></a>**$1.** `)

	// Add anchors for question headings like [Question 1:](#q1)
	questionRe := regexp.MustCompile(`\[Question ([^\]]+)\]\(#([^)]+)\)`)
	markdown = questionRe.ReplaceAllString(markdown, `<a id="$2"></a>Question $1`)

	return markdown
}

func ensureTitle(markdown, title string) string {
	if title == "" {
		return markdown
	}

	lines := strings.Split(markdown, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") {
		return markdown
	}

	return fmt.Sprintf("# %s\n\n%s", title, markdown)
}
