package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVTTFile(t *testing.T) {
	// Create a minimal VTT test file
	content := `WEBVTT
Kind: captions
Language: en

00:00:03.120 --> 00:00:05.670 align:start position:0%

We<00:00:03.360><c> have</c><00:00:03.439><c> a</c><00:00:03.600><c> lot</c><00:00:03.760><c> to</c><00:00:03.919><c> talk</c><00:00:04.160><c> about.</c><00:00:04.960><c> An</c><00:00:05.200><c> apostle</c>

00:00:05.670 --> 00:00:05.680 align:start position:0%
We have a lot to talk about. An apostle

00:00:05.680 --> 00:00:08.790 align:start position:0%
We have a lot to talk about. An apostle
of<00:00:05.920><c> God</c><00:00:06.560><c> just</c><00:00:06.960><c> debunked</c><00:00:07.600><c> false</c><00:00:08.080><c> narratives</c>

00:00:08.790 --> 00:00:08.800 align:start position:0%
of God just debunked false narratives

`
	tmpDir := t.TempDir()
	vttPath := filepath.Join(tmpDir, "test.en.vtt")
	if err := os.WriteFile(vttPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cues, err := ParseVTTFile(vttPath)
	if err != nil {
		t.Fatalf("ParseVTTFile failed: %v", err)
	}

	if len(cues) == 0 {
		t.Fatal("expected cues, got none")
	}

	t.Logf("Parsed %d raw cues", len(cues))
	for i, c := range cues {
		t.Logf("  [%d] %.1f-%.1f: %s", i, c.Begin, c.End, c.Text)
	}

	// Verify tags were stripped
	for _, c := range cues {
		if strings.Contains(c.Text, "<") || strings.Contains(c.Text, ">") {
			t.Errorf("cue still contains tags: %s", c.Text)
		}
	}

	// Run through the full pipeline
	deduped := DeduplicateCues(cues)
	t.Logf("\nDeduped to %d cues:", len(deduped))
	for i, c := range deduped {
		t.Logf("  [%d] %.1f-%.1f: %s", i, c.Begin, c.End, c.Text)
	}

	sentences := MergeCuesIntoSentences(deduped)
	t.Logf("\nSplit into %d sentences:", len(sentences))
	for i, s := range sentences {
		t.Logf("  [%d] %.1f-%.1f: %s", i, s.Begin, s.End, s.Text)
	}

	// Legacy paragraph merging still works
	paragraphs := MergeCuesIntoParagraphs(deduped)
	t.Logf("\nMerged into %d paragraphs:", len(paragraphs))
	for i, p := range paragraphs {
		t.Logf("  [%d] %.1f: %s", i, p.Begin, p.Text)
	}
}

func TestParseVTTTime(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"00:00:03.120", 3.12},
		{"00:01:30.500", 90.5},
		{"01:02:03.456", 3723.456},
		{"00:00:00.000", 0},
	}
	for _, tt := range tests {
		got := parseVTTTime(tt.input)
		if got != tt.expected {
			t.Errorf("parseVTTTime(%q) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestProcessSubtitlesDetectsFormat(t *testing.T) {
	// Test that ProcessSubtitles rejects unsupported formats
	tmpDir := t.TempDir()
	unsupported := filepath.Join(tmpDir, "test.srt")
	os.WriteFile(unsupported, []byte("1\n00:00:01,000 --> 00:00:02,000\nHello\n"), 0644)

	meta := &VideoMetadata{
		ID:    "test123",
		Title: "Test",
		URL:   "https://www.youtube.com/watch?v=test123",
	}

	_, err := ProcessSubtitles(unsupported, meta, tmpDir)
	if err == nil {
		t.Error("expected error for .srt format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported subtitle format") {
		t.Errorf("expected 'unsupported subtitle format' error, got: %v", err)
	}
}

func TestCookieArgs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake cookie file
	cookiePath := filepath.Join(tmpDir, "cookies.txt")
	os.WriteFile(cookiePath, []byte("# Netscape HTTP Cookie File\n"), 0644)

	cfg := &Config{CookieFile: cookiePath}

	// Config cookie file exists → should return args
	args := cookieArgs(cfg, "")
	if len(args) != 2 || args[0] != "--cookies" || args[1] != cookiePath {
		t.Errorf("expected [--cookies %s], got %v", cookiePath, args)
	}

	// Override takes precedence
	overridePath := filepath.Join(tmpDir, "override.txt")
	os.WriteFile(overridePath, []byte("# cookies\n"), 0644)
	args = cookieArgs(cfg, overridePath)
	if len(args) != 2 || args[1] != overridePath {
		t.Errorf("expected override path, got %v", args)
	}

	// Non-existent file → no args
	args = cookieArgs(cfg, filepath.Join(tmpDir, "nonexistent.txt"))
	if len(args) != 0 {
		t.Errorf("expected no args for nonexistent file, got %v", args)
	}

	// Empty config, no override → no args
	cfg2 := &Config{}
	args = cookieArgs(cfg2, "")
	if len(args) != 0 {
		t.Errorf("expected no args for empty config, got %v", args)
	}
}

func TestMergeCuesIntoSentences(t *testing.T) {
	cues := []Cue{
		{Begin: 0.0, End: 3.0, Text: "We have a lot to talk about."},
		{Begin: 3.1, End: 6.0, Text: "An apostle of God just debunked false narratives."},
		{Begin: 6.2, End: 9.0, Text: "Let's dive into what Elder Bednar said about"},
		{Begin: 9.1, End: 12.0, Text: "the nature of continuing revelation."},
		{Begin: 12.5, End: 15.0, Text: "He made it clear that the canon is not closed."},
	}

	sentences := MergeCuesIntoSentences(cues)

	t.Logf("Got %d sentences:", len(sentences))
	for i, s := range sentences {
		t.Logf("  [%d] %.1f-%.1f: %s", i, s.Begin, s.End, s.Text)
	}

	// Should split into 4 sentences:
	// 1. "We have a lot to talk about."
	// 2. "An apostle of God just debunked false narratives."
	// 3. "Let's dive into what Elder Bednar said about the nature of continuing revelation."
	// 4. "He made it clear that the canon is not closed."
	if len(sentences) != 4 {
		t.Errorf("expected 4 sentences, got %d", len(sentences))
	}

	// First sentence should start at 0.0
	if len(sentences) > 0 && sentences[0].Begin != 0.0 {
		t.Errorf("expected first sentence begin=0.0, got %.1f", sentences[0].Begin)
	}

	// Third sentence spans cues 3+4, should start at cue 3's begin
	if len(sentences) > 2 {
		if sentences[2].Begin < 6.0 || sentences[2].Begin > 6.5 {
			t.Errorf("expected third sentence begin ~6.2, got %.1f", sentences[2].Begin)
		}
		if !strings.Contains(sentences[2].Text, "Elder Bednar") {
			t.Errorf("expected third sentence to contain 'Elder Bednar', got: %s", sentences[2].Text)
		}
		if !strings.Contains(sentences[2].Text, "continuing revelation") {
			t.Errorf("expected third sentence to contain 'continuing revelation', got: %s", sentences[2].Text)
		}
	}
}

func TestMergeCuesIntoSentencesAbbreviations(t *testing.T) {
	cues := []Cue{
		{Begin: 0.0, End: 4.0, Text: "Dr. Smith explained the principle."},
		{Begin: 4.1, End: 8.0, Text: "He said Mr. Johnson agreed."},
	}

	sentences := MergeCuesIntoSentences(cues)

	t.Logf("Got %d sentences:", len(sentences))
	for i, s := range sentences {
		t.Logf("  [%d] %.1f-%.1f: %s", i, s.Begin, s.End, s.Text)
	}

	// Should NOT split at "Dr." — should get 2 sentences
	if len(sentences) != 2 {
		t.Errorf("expected 2 sentences (not splitting on Dr./Mr.), got %d", len(sentences))
	}
}

func TestMergeCuesIntoSentencesNoPunctuation(t *testing.T) {
	// Auto-captions often lack punctuation entirely
	var cues []Cue
	for i := 0; i < 20; i++ {
		begin := float64(i) * 3.0
		cues = append(cues, Cue{
			Begin: begin,
			End:   begin + 2.5,
			Text:  "some words without any punctuation here",
		})
	}

	sentences := MergeCuesIntoSentences(cues)

	t.Logf("Got %d sentences from %d unpunctuated cues", len(sentences), len(cues))
	for i, s := range sentences {
		t.Logf("  [%d] %.1f-%.1f: %s", i, s.Begin, s.End, s.Text)
	}

	// Should fall back to gap-based splitting (each cue has a 0.5s gap)
	// With mergeGapThreshold=1.5, the 0.5s gaps should merge, but the fallback
	// triggers because we get 1 sentence from 20 cues (ratio < 1/10)
	if len(sentences) < 2 {
		t.Logf("Note: fallback produced %d segments (gaps may be too small to split)", len(sentences))
	}
}

func TestIsSentenceEnding(t *testing.T) {
	tests := []struct {
		word     string
		expected bool
	}{
		{"about.", true},
		{"said?", true},
		{"now!", true},
		{"Dr.", false},
		{"Mr.", false},
		{"etc.", false},
		{"e.g.", false},
		{"A.", false}, // initial
		{"hello", false},
		{`said."`, true},
		{`it?"`, true},
	}
	for _, tt := range tests {
		got := isSentenceEnding(tt.word)
		if got != tt.expected {
			t.Errorf("isSentenceEnding(%q) = %v, want %v", tt.word, got, tt.expected)
		}
	}
}

func TestGenerateTranscriptMarkdownSentences(t *testing.T) {
	meta := &VideoMetadata{
		ID:         "test123",
		Title:      "Test Video",
		Channel:    "Test Channel",
		UploadDate: "20260214",
		Duration:   120,
		URL:        "https://www.youtube.com/watch?v=test123",
	}

	sentences := []Sentence{
		{Begin: 3.0, End: 5.0, Text: "First sentence here."},
		{Begin: 5.1, End: 8.0, Text: "Second sentence follows."},
		{Begin: 15.0, End: 18.0, Text: "After a gap, third sentence."},
	}

	md := GenerateTranscriptMarkdown(meta, sentences)

	// Should contain reference-style timestamp links in body
	if !strings.Contains(md, "[0:03][t3]") {
		t.Error("expected [0:03][t3] reference link")
	}
	if !strings.Contains(md, "[0:05][t5]") {
		t.Error("expected [0:05][t5] reference link")
	}
	if !strings.Contains(md, "[0:15][t15]") {
		t.Error("expected [0:15][t15] reference link")
	}

	// Should have paragraph break between sentence 2 and 3 (gap > 3s)
	if !strings.Contains(md, "Second sentence follows.\n\n[0:15][t15]") {
		t.Error("expected paragraph break between sentences 2 and 3")
	}

	// Reference definitions should appear at the bottom
	if !strings.Contains(md, "[t3]: https://www.youtube.com/watch?v=test123&t=3") {
		t.Error("expected [t3] reference definition")
	}
	if !strings.Contains(md, "[t15]: https://www.youtube.com/watch?v=test123&t=15") {
		t.Error("expected [t15] reference definition")
	}

	t.Logf("Generated markdown:\n%s", md)
}
