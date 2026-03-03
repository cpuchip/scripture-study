package citations

import (
	"testing"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		input   string
		book    string
		chapter int
		verses  string
		wantErr bool
	}{
		// Standard works
		{"3 Nephi 21:10", "3 Nephi", 21, "10", false},
		{"D&C 113:6", "D&C", 113, "6", false},
		{"Isaiah 11:1", "Isaiah", 11, "1", false},
		{"Alma 32:21", "Alma", 32, "21", false},
		{"1 Nephi 3:7", "1 Nephi", 3, "7", false},
		{"Moses 1:39", "Moses", 1, "39", false},
		{"JS-H 1:19", "JS-H", 1, "19", false},
		{"JS-M 1:37", "JS-M", 1, "37", false},
		{"Articles of Faith 1:13", "Articles of Faith", 1, "13", false},

		// Verse ranges
		{"Isaiah 11:1-3", "Isaiah", 11, "1-3", false},
		{"1 Corinthians 13:4,7", "1 Corinthians", 13, "4,7", false},
		{"Abraham 3:22-23", "Abraham", 3, "22-23", false},

		// Chapter only (no verse)
		{"Alma 32", "Alma", 32, "", false},
		{"D&C 113", "D&C", 113, "", false},

		// Abbreviations
		{"Isa 11:1", "Isaiah", 11, "1", false},
		{"Matt 5:48", "Matthew", 5, "48", false},
		{"Rev 21:4", "Revelation", 21, "4", false},
		{"Hel 5:12", "Helaman", 5, "12", false},
		{"Mos 3:19", "Mosiah", 3, "19", false},
		{"Moro 10:4", "Moroni", 10, "4", false},
		{"Abr 3:22", "Abraham", 3, "22", false},
		{"1 Cor 13:4", "1 Corinthians", 13, "4", false},
		{"2 Tim 3:16", "2 Timothy", 3, "16", false},
		{"Ps 23:1", "Psalms", 23, "1", false},
		{"Gen 1:1", "Genesis", 1, "1", false},
		{"Deut 6:5", "Deuteronomy", 6, "5", false},
		{"Prov 3:5", "Proverbs", 3, "5", false},

		// Full names for D&C
		{"Doctrine and Covenants 93:36", "D&C", 93, "36", false},

		// With unicode dashes
		{"JS—H 1:19", "JS-H", 1, "19", false},

		// Errors
		{"", "", 0, "", true},
		{"FakeBook 1:1", "", 0, "", true},
		{"Isaiah", "", 0, "", true}, // no chapter
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, err := ParseReference(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseReference(%q) expected error, got %+v", tt.input, ref)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseReference(%q) unexpected error: %v", tt.input, err)
				return
			}
			if ref.Book != tt.book {
				t.Errorf("ParseReference(%q).Book = %q, want %q", tt.input, ref.Book, tt.book)
			}
			if ref.Chapter != tt.chapter {
				t.Errorf("ParseReference(%q).Chapter = %d, want %d", tt.input, ref.Chapter, tt.chapter)
			}
			if ref.Verses != tt.verses {
				t.Errorf("ParseReference(%q).Verses = %q, want %q", tt.input, ref.Verses, tt.verses)
			}
		})
	}
}

func TestBookIDMapping(t *testing.T) {
	// Spot-check critical book IDs
	tests := map[string]int{
		"Genesis":    101,
		"Isaiah":     123,
		"Matthew":    140,
		"Revelation": 166,
		"1 Nephi":    205,
		"3 Nephi":    215,
		"Moroni":     219,
		"D&C":        302,
		"Moses":      401,
		"Abraham":    402,
		"JS-H":       405,
	}

	for book, expectedID := range tests {
		id, ok := BookIDs[book]
		if !ok {
			t.Errorf("BookIDs[%q] not found", book)
			continue
		}
		if id != expectedID {
			t.Errorf("BookIDs[%q] = %d, want %d", book, id, expectedID)
		}
	}
}

func TestParseRefText(t *testing.T) {
	tests := []struct {
		input     string
		speaker   string
		reference string
	}{
		{"1989-O:54, Gordon B. Hinckley", "Gordon B. Hinckley", "1989-O:54"},
		{"JD 19:81b, John Taylor", "John Taylor", "JD 19:81b"},
		{"2016-A:86, D. Todd Christofferson", "D. Todd Christofferson", "2016-A:86"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			speaker, reference := parseRefText(tt.input)
			if speaker != tt.speaker {
				t.Errorf("parseRefText(%q) speaker = %q, want %q", tt.input, speaker, tt.speaker)
			}
			if reference != tt.reference {
				t.Errorf("parseRefText(%q) reference = %q, want %q", tt.input, reference, tt.reference)
			}
		})
	}
}

func TestDisplay(t *testing.T) {
	ref := &ScriptureRef{Book: "3 Nephi", Chapter: 21, Verses: "10"}
	if got := ref.Display(); got != "3 Nephi 21:10" {
		t.Errorf("Display() = %q, want %q", got, "3 Nephi 21:10")
	}

	ref2 := &ScriptureRef{Book: "Alma", Chapter: 32, Verses: ""}
	if got := ref2.Display(); got != "Alma 32" {
		t.Errorf("Display() = %q, want %q", got, "Alma 32")
	}
}

func TestParseHTML(t *testing.T) {
	// Minimal HTML matching the BYU response format
	html := `<li><a href="javascript:void(0);" class="refcounter" onclick="getTalk('4793', '19099');">
		<div class="reference referencewatch referencelisten">1989-O:54, Gordon B. Hinckley</div>
		<div class="talktitle talktitlewatch talktitlelisten">An Ensign to the Nations</div></a></li>`

	cites := parseHTML(html)
	if len(cites) != 1 {
		t.Fatalf("parseHTML returned %d citations, want 1", len(cites))
	}

	c := cites[0]
	if c.TalkID != "4793" {
		t.Errorf("TalkID = %q, want %q", c.TalkID, "4793")
	}
	if c.RefID != "19099" {
		t.Errorf("RefID = %q, want %q", c.RefID, "19099")
	}
	if c.Speaker != "Gordon B. Hinckley" {
		t.Errorf("Speaker = %q, want %q", c.Speaker, "Gordon B. Hinckley")
	}
	if c.Reference != "1989-O:54" {
		t.Errorf("Reference = %q, want %q", c.Reference, "1989-O:54")
	}
	if c.Title != "An Ensign to the Nations" {
		t.Errorf("Title = %q, want %q", c.Title, "An Ensign to the Nations")
	}
}
