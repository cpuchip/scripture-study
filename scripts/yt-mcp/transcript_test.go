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
