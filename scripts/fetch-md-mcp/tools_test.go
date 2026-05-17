package main

import (
	"strings"
	"testing"
)

// ES.5.s2: detectDocExt decides whether a fetched body is a non-HTML
// document (route to tabula) or HTML (route to readability).
func TestDetectDocExt(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		fetchURL string
		want    string
	}{
		{"pdf by magic bytes", "%PDF-1.6\n%âãÏÓ binary…", "https://x.test/report", ".pdf"},
		{"pdf magic beats html extension", "%PDF-1.4 stuff", "https://x.test/doc.html", ".pdf"},
		{"pdf by url extension", "not actually pdf bytes", "https://x.test/paper.pdf", ".pdf"},
		{"docx by url extension", "PK\x03\x04 zip bytes", "https://x.test/memo.docx", ".docx"},
		{"xlsx by url extension", "PK\x03\x04 zip bytes", "https://x.test/sheet.xlsx", ".xlsx"},
		{"epub by url extension", "PK\x03\x04 zip bytes", "https://x.test/book.epub", ".epub"},
		{"html body, html url", "<!DOCTYPE html><html>…", "https://x.test/page.html", ""},
		{"html body, no extension", "<html><body>hi</body></html>", "https://x.test/article", ""},
		{"query string after pdf ext", "%nope", "https://x.test/f.pdf?dl=1", ".pdf"},
		{"unknown zip-ish extension stays html", "PK stuff", "https://x.test/thing.zip", ""},
	}
	for _, c := range cases {
		got := detectDocExt([]byte(c.body), c.fetchURL)
		if got != c.want {
			t.Errorf("%s: detectDocExt(%q, %q) = %q, want %q",
				c.name, c.body[:min(len(c.body), 16)], c.fetchURL, got, c.want)
		}
	}
}

// ES.5.s2: buildDocOutput derives a title from the URL and honors
// max_chars truncation.
func TestBuildDocOutput(t *testing.T) {
	md := strings.Repeat("word ", 100) // 500 chars
	out := buildDocOutput("https://x.test/papers/bacteriopolis.pdf", md, 0)
	if out.Title != "bacteriopolis.pdf" {
		t.Errorf("title = %q, want bacteriopolis.pdf", out.Title)
	}
	if out.WordCount != 100 {
		t.Errorf("word count = %d, want 100", out.WordCount)
	}
	if out.Truncated {
		t.Error("no max_chars set — should not be truncated")
	}

	clipped := buildDocOutput("https://x.test/a.pdf", md, 120)
	if !clipped.Truncated {
		t.Error("max_chars=120 with 500-char body — should be truncated")
	}
	if !strings.HasSuffix(clipped.Markdown, "[…truncated]") {
		t.Error("truncated output should carry the truncation marker")
	}
}
