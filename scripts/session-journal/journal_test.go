package journal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadAll(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	entry := &Entry{
		Date:      "2026-02-28",
		SessionID: "test-session",
		Intent:    "Test the journal store",
		Discoveries: []Discovery{
			{Title: "It works", Detail: "The store writes and reads YAML"},
		},
		Surprises: []string{"It was easy"},
		Relationship: []Quality{
			{Name: "trust", Detail: "Tests are trustworthy"},
		},
		CarryForward: []CarryItem{
			{Priority: "high", Note: "Keep testing"},
			{Priority: "low", Note: "Consider more tests"},
		},
		Questions: []string{"Is this enough?"},
		Tags:      []string{"testing", "trust"},
	}

	path, err := store.Write(entry)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("written file does not exist: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.Date != "2026-02-28" {
		t.Errorf("date = %q, want %q", got.Date, "2026-02-28")
	}
	if got.SessionID != "test-session" {
		t.Errorf("session_id = %q, want %q", got.SessionID, "test-session")
	}
	if len(got.Discoveries) != 1 || got.Discoveries[0].Title != "It works" {
		t.Errorf("discoveries mismatch: %+v", got.Discoveries)
	}
	if len(got.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(got.Tags))
	}
}

func TestRecent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Write 3 entries
	for _, date := range []string{"2026-01-01", "2026-02-01", "2026-03-01"} {
		_, err := store.Write(&Entry{
			Date:      date,
			SessionID: "session-" + date,
			Intent:    "Test entry for " + date,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	recent, err := store.Recent(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 {
		t.Fatalf("expected 2 recent, got %d", len(recent))
	}
	if recent[0].Date != "2026-02-01" {
		t.Errorf("first recent = %q, want 2026-02-01", recent[0].Date)
	}
	if recent[1].Date != "2026-03-01" {
		t.Errorf("second recent = %q, want 2026-03-01", recent[1].Date)
	}
}

func TestByTopic(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	store.Write(&Entry{
		Date:      "2026-01-01",
		SessionID: "faith-study",
		Intent:    "Study faith",
		Tags:      []string{"faith", "scripture"},
	})
	store.Write(&Entry{
		Date:      "2026-01-02",
		SessionID: "charity-study",
		Intent:    "Study charity",
		Tags:      []string{"charity", "becoming"},
	})
	store.Write(&Entry{
		Date:      "2026-01-03",
		SessionID: "trust-test",
		Intent:    "Test trust recovery",
		Discoveries: []Discovery{
			{Title: "Trust through faith", Detail: "Faith enables trust"},
		},
	})

	// Search for "faith" — should match first and third
	matches, err := store.ByTopic("faith")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for 'faith', got %d", len(matches))
	}
	if matches[0].SessionID != "faith-study" {
		t.Errorf("first match = %q, want faith-study", matches[0].SessionID)
	}
	if matches[1].SessionID != "trust-test" {
		t.Errorf("second match = %q, want trust-test", matches[1].SessionID)
	}

	// Search for "CHARITY" (case insensitive)
	matches, err = store.ByTopic("CHARITY")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].SessionID != "charity-study" {
		t.Errorf("case-insensitive search failed: %+v", matches)
	}
}

func TestSince(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	store.Write(&Entry{Date: "2026-01-15", SessionID: "jan", Intent: "January"})
	store.Write(&Entry{Date: "2026-02-15", SessionID: "feb", Intent: "February"})
	store.Write(&Entry{Date: "2026-03-15", SessionID: "mar", Intent: "March"})

	matches, err := store.Since("2026-02-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 entries since 2026-02-15, got %d", len(matches))
	}
	if matches[0].SessionID != "feb" || matches[1].SessionID != "mar" {
		t.Errorf("unexpected matches: %v, %v", matches[0].SessionID, matches[1].SessionID)
	}
}

func TestCarryForwardItems(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	store.Write(&Entry{
		Date:      "2026-01-01",
		SessionID: "session-a",
		Intent:    "First session",
		CarryForward: []CarryItem{
			{Priority: "high", Note: "Do this first"},
			{Priority: "low", Note: "Nice to have", Resolved: true, ResolvedDate: "2026-01-02", ResolvedNote: "Done"},
		},
	})
	store.Write(&Entry{
		Date:      "2026-01-02",
		SessionID: "session-b",
		Intent:    "Second session",
		CarryForward: []CarryItem{
			{Priority: "medium", Note: "Middle priority"},
		},
	})

	// All unresolved
	items, err := store.CarryForwardItems("all", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 unresolved items, got %d", len(items))
	}

	// High only
	items, err = store.CarryForwardItems("high", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Note != "Do this first" {
		t.Errorf("high filter failed: %+v", items)
	}

	// Include resolved
	items, err = store.CarryForwardItems("all", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 total items, got %d", len(items))
	}
}

func TestAllQuestions(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	store.Write(&Entry{
		Date:      "2026-01-01",
		SessionID: "session-a",
		Intent:    "Session A",
		Questions: []string{"What is truth?", "Does it matter?"},
	})
	store.Write(&Entry{
		Date:      "2026-01-02",
		SessionID: "session-b",
		Intent:    "Session B",
		Questions: []string{"Can a tool be sanctified?"},
	})

	qs, err := store.AllQuestions()
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(qs))
	}
	if qs[2].Question != "Can a tool be sanctified?" {
		t.Errorf("unexpected third question: %q", qs[2].Question)
	}
	if qs[2].SessionID != "session-b" {
		t.Errorf("unexpected source session: %q", qs[2].SessionID)
	}
}

func TestRetroactiveEntry(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	entry := &Entry{
		Date:      "2026-01-15",
		SessionID: "retroactive-test",
		Intent:    "Reconstructed from chat history",
		Retroactive: &Retroactive{
			Source:        "chat-history",
			DateCertainty: "approximate",
			InferredFrom:  "git log for study/charity.md",
			CapturedDate:  "2026-02-28",
		},
	}

	_, err = store.Write(entry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Retroactive == nil {
		t.Fatal("retroactive metadata lost")
	}
	if entries[0].Retroactive.DateCertainty != "approximate" {
		t.Errorf("date_certainty = %q, want approximate", entries[0].Retroactive.DateCertainty)
	}
}

func TestSanitizeSlug(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"source-verification-and-covenant", "source-verification-and-covenant"},
		{"Source Verification", "source-verification"},
		{"test__double", "test-double"},
		{"special!@#chars", "specialchars"},
		{"  leading-trailing  ", "leading-trailing"},
	}

	for _, tc := range tests {
		got := sanitizeSlug(tc.in)
		if got != tc.want {
			t.Errorf("sanitizeSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFileNaming(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	path, err := store.Write(&Entry{
		Date:      "2026-02-28",
		SessionID: "My Cool Session",
		Intent:    "Test filename",
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(dir, "2026-02-28--my-cool-session.yaml")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}
