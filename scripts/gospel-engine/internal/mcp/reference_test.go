package mcp

import (
	"testing"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		wantBook string
		wantCh   int
		wantV    int
		wantEnd  int
	}{
		{"1 Nephi 3:7", "scripture", "1-ne", 3, 7, 0},
		{"D&C 93:24-30", "scripture", "dc", 93, 24, 30},
		{"Moses 6:57", "scripture", "moses", 6, 57, 0},
		{"Mosiah 4", "scripture", "mosiah", 4, 0, 0},
		{"John 3:16-17", "scripture", "john", 3, 16, 17},
		{"Exodus 18:21", "scripture", "ex", 18, 21, 0},
		{"Alma 32:27", "scripture", "alma", 32, 27, 0},
		{"Abraham 3:22-23", "scripture", "abr", 3, 22, 23},
		{"Helaman 5:12", "scripture", "hel", 5, 12, 0},
		{"3 Nephi 11:29", "scripture", "3-ne", 11, 29, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref := parseReference(tt.input)
			if ref.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", ref.Type, tt.wantType)
			}
			if ref.Book != tt.wantBook {
				t.Errorf("Book = %q, want %q", ref.Book, tt.wantBook)
			}
			if ref.Chapter != tt.wantCh {
				t.Errorf("Chapter = %d, want %d", ref.Chapter, tt.wantCh)
			}
			if ref.Verse != tt.wantV {
				t.Errorf("Verse = %d, want %d", ref.Verse, tt.wantV)
			}
			if ref.EndVerse != tt.wantEnd {
				t.Errorf("EndVerse = %d, want %d", ref.EndVerse, tt.wantEnd)
			}
		})
	}
}

func TestNormalizeBookName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"genesis", "gen"},
		{"1 nephi", "1-ne"},
		{"d&c", "dc"},
		{"dc", "dc"},
		{"mosiah", "mosiah"},
		{"alma", "alma"},
		{"3 nephi", "3-ne"},
		{"moses", "moses"},
		{"abraham", "abr"},
		{"hebrews", "heb"},
		{"revelation", "rev"},
		{"revelations", "rev"},
		{"psalm", "ps"},
		{"psalms", "ps"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeBookName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeBookName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatBookName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1-ne", "1 Nephi"},
		{"dc", "D&C"},
		{"moses", "Moses"},
		{"hel", "Helaman"},
		{"alma", "Alma"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := formatBookName(tt.input)
			if got != tt.want {
				t.Errorf("formatBookName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatScriptureRef(t *testing.T) {
	tests := []struct {
		book    string
		chapter int
		verse   int
		want    string
	}{
		{"dc", 93, 24, "D&C 93:24"},
		{"1-ne", 3, 7, "1 Nephi 3:7"},
		{"mosiah", 4, 0, "Mosiah 4"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatScriptureRef(tt.book, tt.chapter, tt.verse)
			if got != tt.want {
				t.Errorf("formatScriptureRef(%q, %d, %d) = %q, want %q", tt.book, tt.chapter, tt.verse, got, tt.want)
			}
		})
	}
}
