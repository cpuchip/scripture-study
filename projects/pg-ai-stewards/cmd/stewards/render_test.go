package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFmtMicro(t *testing.T) {
	cases := []struct {
		micro int64
		want  string
	}{
		{0, "$0.0000"},
		{5_000, "$0.0050"},       // half a cent — must not render as $0.00
		{1_000_000, "$1.0000"},   // one dollar
		{1_234_567, "$1.2345"},   // truncates to 4 places
		{1_600_000_000, "$1,600.0000"},
		{-2_500_000, "-$2.5000"},
		{100, "$0.0001"},         // smallest visible unit at 4 places
	}
	for _, c := range cases {
		if got := fmtMicro(c.micro); got != c.want {
			t.Errorf("fmtMicro(%d) = %q, want %q", c.micro, got, c.want)
		}
	}
}

func TestGroupThousands(t *testing.T) {
	cases := map[int64]string{
		0:             "0",
		999:           "999",
		1000:          "1,000",
		1234567:       "1,234,567",
		2_200_000_000: "2,200,000,000",
		-12345:        "-12,345",
	}
	for in, want := range cases {
		if got := groupThousands(in); got != want {
			t.Errorf("groupThousands(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRelTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		t    time.Time
		want string
	}{
		{time.Time{}, "—"},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-3 * time.Hour), "3h ago"},
		{now.Add(-2 * 24 * time.Hour), "2d ago"},
	}
	for _, c := range cases {
		if got := relTime(c.t); got != c.want {
			t.Errorf("relTime(%v) = %q, want %q", c.t, got, c.want)
		}
	}
}

func TestSummarizeCounts(t *testing.T) {
	got := summarizeCounts(map[string]int{"completed": 7, "failed": 2, "pending": 7})
	// highest count first; ties broken alphabetically → completed(7), pending(7), failed(2)
	want := "16 items — completed 7 · pending 7 · failed 2"
	if got != want {
		t.Errorf("summarizeCounts = %q, want %q", got, want)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stewards.json")
	t.Setenv("STEWARDS_CONFIG", path)
	// no env override, so activeProject() reads the file
	t.Setenv("STEWARDS_PROJECT", "")

	if got := activeProject(); got != "" {
		t.Fatalf("fresh config: activeProject() = %q, want empty", got)
	}
	if err := saveConfig(Config{ActiveProject: "scripture-book"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	if got := loadConfig().ActiveProject; got != "scripture-book" {
		t.Fatalf("loadConfig after save = %q, want scripture-book", got)
	}
	if got := activeProject(); got != "scripture-book" {
		t.Fatalf("activeProject after save = %q, want scripture-book", got)
	}
	// env override beats the file
	t.Setenv("STEWARDS_PROJECT", "pg-ai-stewards")
	if got := activeProject(); got != "pg-ai-stewards" {
		t.Fatalf("activeProject with env override = %q, want pg-ai-stewards", got)
	}
	// the file is real and parseable
	if _, err := os.ReadFile(path); err != nil {
		t.Fatalf("config file not written: %v", err)
	}
}
