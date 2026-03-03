package citations

import (
	"fmt"
	"regexp"
	"strings"
)

// ScriptureRef is a parsed scripture reference.
type ScriptureRef struct {
	Book    string // Canonical book name, e.g. "3 Nephi", "D&C", "Isaiah"
	Chapter int    // Chapter/section number
	Verses  string // Verse string (may be "10", "10-12", "10,12", etc.)
}

// Display returns the human-readable reference.
func (r *ScriptureRef) Display() string {
	if r.Verses != "" {
		return fmt.Sprintf("%s %d:%s", r.Book, r.Chapter, r.Verses)
	}
	return fmt.Sprintf("%s %d", r.Book, r.Chapter)
}

// ParseReference parses a human-readable scripture reference into its components.
// Accepts formats like:
//
//	"3 Nephi 21:10"
//	"D&C 113:6"
//	"Isaiah 11:1-3"
//	"1 Corinthians 13:4,7"
//	"Alma 32"           (chapter only, no verse)
//	"Moses 1:39"
//	"JS—H 1:19"
//	"Abraham 3:22-23"
func ParseReference(input string) (*ScriptureRef, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty reference")
	}

	// Normalize some common variations
	normalized := normalizeInput(input)

	// Try to match against known book names (longest first to handle "1 Nephi" vs "Nephi")
	var matchedBook string
	var remainder string

	for _, book := range sortedBookNames {
		lower := strings.ToLower(book)
		normalizedLower := strings.ToLower(normalized)
		if strings.HasPrefix(normalizedLower, lower) {
			// Make sure the match is followed by a space or number or end of string
			rest := normalized[len(book):]
			if rest == "" || rest[0] == ' ' || rest[0] == ':' || (rest[0] >= '0' && rest[0] <= '9') {
				matchedBook = book
				remainder = strings.TrimSpace(rest)
				break
			}
		}
	}

	if matchedBook == "" {
		return nil, fmt.Errorf("unrecognized book in reference: %q", input)
	}

	if remainder == "" {
		return nil, fmt.Errorf("no chapter specified in reference: %q", input)
	}

	// Parse chapter:verse from remainder
	ref := &ScriptureRef{Book: matchedBook}

	// Pattern: chapter:verse(s) or just chapter
	chapterVersePattern := regexp.MustCompile(`^(\d+)(?::(.+))?$`)
	m := chapterVersePattern.FindStringSubmatch(remainder)
	if m == nil {
		return nil, fmt.Errorf("cannot parse chapter/verse from %q in reference %q", remainder, input)
	}

	_, err := fmt.Sscanf(m[1], "%d", &ref.Chapter)
	if err != nil {
		return nil, fmt.Errorf("invalid chapter number %q: %w", m[1], err)
	}

	if m[2] != "" {
		ref.Verses = m[2]
	}

	return ref, nil
}

// normalizeInput cleans up input for matching.
func normalizeInput(input string) string {
	s := input

	// Normalize unicode dashes to hyphens
	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "–", "-")

	// Normalize "D & C" or "D&C" variations
	s = regexp.MustCompile(`(?i)d\s*&\s*c`).ReplaceAllString(s, "D&C")
	s = regexp.MustCompile(`(?i)doctrine\s+and\s+covenants`).ReplaceAllString(s, "D&C")

	// Normalize "JS-H" / "JS—H" / "JS-M" / "JS—M"
	s = regexp.MustCompile(`(?i)js\s*[-—]\s*h`).ReplaceAllString(s, "JS-H")
	s = regexp.MustCompile(`(?i)js\s*[-—]\s*m`).ReplaceAllString(s, "JS-M")
	s = regexp.MustCompile(`(?i)joseph\s+smith[-—]history`).ReplaceAllString(s, "JS-H")
	s = regexp.MustCompile(`(?i)joseph\s+smith[-—]matthew`).ReplaceAllString(s, "JS-M")

	// Normalize "A of F" / "Articles of Faith"
	s = regexp.MustCompile(`(?i)articles?\s+of\s+faith`).ReplaceAllString(s, "Articles of Faith")

	// Normalize "W of M" / "Words of Mormon"
	s = regexp.MustCompile(`(?i)words?\s+of\s+mormon`).ReplaceAllString(s, "Words of Mormon")

	// Normalize "O.D." / "Official Declaration"
	s = regexp.MustCompile(`(?i)official\s+declarations?`).ReplaceAllString(s, "O.D.")
	s = regexp.MustCompile(`(?i)o\.?\s*d\.?(?:\s|$)`).ReplaceAllString(s, "O.D. ")

	// Handle common abbreviations
	s = regexp.MustCompile(`(?i)^gen(?:\s|$)`).ReplaceAllString(s, "Genesis ")
	s = regexp.MustCompile(`(?i)^ex(?:\s|$)`).ReplaceAllString(s, "Exodus ")
	s = regexp.MustCompile(`(?i)^lev(?:\s|$)`).ReplaceAllString(s, "Leviticus ")
	s = regexp.MustCompile(`(?i)^num(?:\s|$)`).ReplaceAllString(s, "Numbers ")
	s = regexp.MustCompile(`(?i)^deut(?:\s|$)`).ReplaceAllString(s, "Deuteronomy ")
	s = regexp.MustCompile(`(?i)^josh(?:\s|$)`).ReplaceAllString(s, "Joshua ")
	s = regexp.MustCompile(`(?i)^judg(?:\s|$)`).ReplaceAllString(s, "Judges ")
	s = regexp.MustCompile(`(?i)^1\s*sam(?:\s|$)`).ReplaceAllString(s, "1 Samuel ")
	s = regexp.MustCompile(`(?i)^2\s*sam(?:\s|$)`).ReplaceAllString(s, "2 Samuel ")
	s = regexp.MustCompile(`(?i)^1\s*kgs(?:\s|$)`).ReplaceAllString(s, "1 Kings ")
	s = regexp.MustCompile(`(?i)^2\s*kgs(?:\s|$)`).ReplaceAllString(s, "2 Kings ")
	s = regexp.MustCompile(`(?i)^1\s*chr(?:on)?(?:\s|$)`).ReplaceAllString(s, "1 Chronicles ")
	s = regexp.MustCompile(`(?i)^2\s*chr(?:on)?(?:\s|$)`).ReplaceAllString(s, "2 Chronicles ")
	s = regexp.MustCompile(`(?i)^neh(?:\s|$)`).ReplaceAllString(s, "Nehemiah ")
	s = regexp.MustCompile(`(?i)^esth(?:\s|$)`).ReplaceAllString(s, "Esther ")
	s = regexp.MustCompile(`(?i)^ps(?:alm)?(?:\s|$)`).ReplaceAllString(s, "Psalms ")
	s = regexp.MustCompile(`(?i)^prov(?:\s|$)`).ReplaceAllString(s, "Proverbs ")
	s = regexp.MustCompile(`(?i)^eccl(?:\s|$)`).ReplaceAllString(s, "Ecclesiastes ")
	s = regexp.MustCompile(`(?i)^song(?:\s|$)`).ReplaceAllString(s, "Song of Solomon ")
	s = regexp.MustCompile(`(?i)^isa(?:\s|$)`).ReplaceAllString(s, "Isaiah ")
	s = regexp.MustCompile(`(?i)^jer(?:\s|$)`).ReplaceAllString(s, "Jeremiah ")
	s = regexp.MustCompile(`(?i)^lam(?:\s|$)`).ReplaceAllString(s, "Lamentations ")
	s = regexp.MustCompile(`(?i)^ezek(?:\s|$)`).ReplaceAllString(s, "Ezekiel ")
	s = regexp.MustCompile(`(?i)^dan(?:\s|$)`).ReplaceAllString(s, "Daniel ")
	s = regexp.MustCompile(`(?i)^hos(?:\s|$)`).ReplaceAllString(s, "Hosea ")
	s = regexp.MustCompile(`(?i)^amos(?:\s|$)`).ReplaceAllString(s, "Amos ")
	s = regexp.MustCompile(`(?i)^obad(?:\s|$)`).ReplaceAllString(s, "Obadiah ")
	s = regexp.MustCompile(`(?i)^mic(?:\s|$)`).ReplaceAllString(s, "Micah ")
	s = regexp.MustCompile(`(?i)^nah(?:\s|$)`).ReplaceAllString(s, "Nahum ")
	s = regexp.MustCompile(`(?i)^hab(?:\s|$)`).ReplaceAllString(s, "Habakkuk ")
	s = regexp.MustCompile(`(?i)^zeph(?:\s|$)`).ReplaceAllString(s, "Zephaniah ")
	s = regexp.MustCompile(`(?i)^hag(?:\s|$)`).ReplaceAllString(s, "Haggai ")
	s = regexp.MustCompile(`(?i)^zech(?:\s|$)`).ReplaceAllString(s, "Zechariah ")
	s = regexp.MustCompile(`(?i)^mal(?:\s|$)`).ReplaceAllString(s, "Malachi ")
	s = regexp.MustCompile(`(?i)^matt(?:\s|$)`).ReplaceAllString(s, "Matthew ")
	s = regexp.MustCompile(`(?i)^rom(?:\s|$)`).ReplaceAllString(s, "Romans ")
	s = regexp.MustCompile(`(?i)^1\s*cor(?:\s|$)`).ReplaceAllString(s, "1 Corinthians ")
	s = regexp.MustCompile(`(?i)^2\s*cor(?:\s|$)`).ReplaceAllString(s, "2 Corinthians ")
	s = regexp.MustCompile(`(?i)^gal(?:\s|$)`).ReplaceAllString(s, "Galatians ")
	s = regexp.MustCompile(`(?i)^eph(?:\s|$)`).ReplaceAllString(s, "Ephesians ")
	s = regexp.MustCompile(`(?i)^phil(?:ip)?(?:\s|$)`).ReplaceAllString(s, "Philippians ")
	s = regexp.MustCompile(`(?i)^col(?:\s|$)`).ReplaceAllString(s, "Colossians ")
	s = regexp.MustCompile(`(?i)^1\s*thes(?:\s|$)`).ReplaceAllString(s, "1 Thessalonians ")
	s = regexp.MustCompile(`(?i)^2\s*thes(?:\s|$)`).ReplaceAllString(s, "2 Thessalonians ")
	s = regexp.MustCompile(`(?i)^1\s*tim(?:\s|$)`).ReplaceAllString(s, "1 Timothy ")
	s = regexp.MustCompile(`(?i)^2\s*tim(?:\s|$)`).ReplaceAllString(s, "2 Timothy ")
	s = regexp.MustCompile(`(?i)^tit(?:\s|$)`).ReplaceAllString(s, "Titus ")
	s = regexp.MustCompile(`(?i)^philem(?:\s|$)`).ReplaceAllString(s, "Philemon ")
	s = regexp.MustCompile(`(?i)^heb(?:\s|$)`).ReplaceAllString(s, "Hebrews ")
	s = regexp.MustCompile(`(?i)^jas(?:\s|$)`).ReplaceAllString(s, "James ")
	s = regexp.MustCompile(`(?i)^1\s*pet(?:\s|$)`).ReplaceAllString(s, "1 Peter ")
	s = regexp.MustCompile(`(?i)^2\s*pet(?:\s|$)`).ReplaceAllString(s, "2 Peter ")
	s = regexp.MustCompile(`(?i)^1\s*jn(?:\s|$)`).ReplaceAllString(s, "1 John ")
	s = regexp.MustCompile(`(?i)^2\s*jn(?:\s|$)`).ReplaceAllString(s, "2 John ")
	s = regexp.MustCompile(`(?i)^3\s*jn(?:\s|$)`).ReplaceAllString(s, "3 John ")
	s = regexp.MustCompile(`(?i)^rev(?:\s|$)`).ReplaceAllString(s, "Revelation ")
	s = regexp.MustCompile(`(?i)^moro(?:\s|$)`).ReplaceAllString(s, "Moroni ")
	s = regexp.MustCompile(`(?i)^morm(?:\s|$)`).ReplaceAllString(s, "Mormon ")
	s = regexp.MustCompile(`(?i)^hel(?:\s|$)`).ReplaceAllString(s, "Helaman ")
	s = regexp.MustCompile(`(?i)^mos(?:\s|$)`).ReplaceAllString(s, "Mosiah ")
	s = regexp.MustCompile(`(?i)^abr(?:\s|$)`).ReplaceAllString(s, "Abraham ")
	s = regexp.MustCompile(`(?i)^w\s*of\s*m(?:\s|$)`).ReplaceAllString(s, "Words of Mormon ")

	return strings.TrimSpace(s)
}
