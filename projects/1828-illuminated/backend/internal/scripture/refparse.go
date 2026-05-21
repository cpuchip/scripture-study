package scripture

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Ref is a parsed scripture reference.
type Ref struct {
	// Original raw input as the user wrote it.
	Raw string

	// Canonical book name ("1 Nephi", "Doctrine and Covenants", "Genesis").
	Book string

	// (volume, abbr, display_order) for the book.
	Meta BookMeta

	// Chapter number (sections for D&C live in this field too).
	Chapter int

	// Verse range. VerseStart == VerseEnd for a single verse. Both zero
	// when the ref is chapter-level only ("john/3").
	VerseStart int
	VerseEnd   int
}

// HasVerse is true when the ref carries verse numbers (not a whole chapter).
func (r Ref) HasVerse() bool { return r.VerseStart > 0 }

// AbbrRef returns the canonical short form like "dc/84:38" or
// "1-ne/3:7-10". Always lowercase, slash-separated, suitable for URLs.
func (r Ref) AbbrRef() string {
	base := r.Meta.Abbr + "/" + strconv.Itoa(r.Chapter)
	if !r.HasVerse() {
		return base
	}
	if r.VerseEnd > r.VerseStart {
		return base + ":" + strconv.Itoa(r.VerseStart) + "-" + strconv.Itoa(r.VerseEnd)
	}
	return base + ":" + strconv.Itoa(r.VerseStart)
}

// HumanRef returns the "Book chapter:verse" form for display ("1 Nephi 3:7").
func (r Ref) HumanRef() string {
	base := r.Book + " " + strconv.Itoa(r.Chapter)
	if !r.HasVerse() {
		return base
	}
	if r.VerseEnd > r.VerseStart {
		return fmt.Sprintf("%s:%d-%d", base, r.VerseStart, r.VerseEnd)
	}
	return fmt.Sprintf("%s:%d", base, r.VerseStart)
}

// ParseRef accepts:
//
//   - Workspace abbr form:   "dc/84:38", "1-ne/3:7-10", "abr/3:19", "john/3"
//   - Human form:            "1 Nephi 3:7", "John 3:16-17", "D&C 84:38"
//   - Verse omitted:         "1 Nephi 3", "john 3", "dc/84"
//
// requireVerse controls whether a missing verse is an error (used by
// the /api/scripture/:ref handler) or acceptable (used by chapter view).
func ParseRef(input string, requireVerse bool) (*Ref, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("empty reference")
	}

	// Detect the workspace abbr form first (it always contains a slash
	// between book and chapter).
	if strings.Contains(raw, "/") {
		ref, err := parseAbbrForm(raw)
		if err != nil {
			return nil, err
		}
		if requireVerse && !ref.HasVerse() {
			return nil, fmt.Errorf("reference missing verse: %s", raw)
		}
		return ref, nil
	}

	// Otherwise treat as human form.
	ref, err := parseHumanForm(raw)
	if err != nil {
		return nil, err
	}
	if requireVerse && !ref.HasVerse() {
		return nil, fmt.Errorf("reference missing verse: %s", raw)
	}
	return ref, nil
}

// parseAbbrForm handles "abbr/chapter[:verseStart[-verseEnd]]".
func parseAbbrForm(s string) (*Ref, error) {
	slash := strings.Index(s, "/")
	abbr := strings.ToLower(strings.TrimSpace(s[:slash]))
	rest := strings.TrimSpace(s[slash+1:])

	bookName, meta, ok := LookupBookByAbbr(abbr)
	if !ok {
		return nil, fmt.Errorf("unknown book abbreviation %q", abbr)
	}

	chap, vs, ve, err := parseChapterAndVerses(rest)
	if err != nil {
		return nil, err
	}
	return &Ref{Raw: s, Book: bookName, Meta: meta, Chapter: chap, VerseStart: vs, VerseEnd: ve}, nil
}

// parseHumanForm handles "1 Nephi 3:7", "Genesis 1", "D&C 84:38".
//
// Strategy: find the longest book-name prefix that matches our map.
// Books range from 4 to 21 chars (e.g. "Job" — actually 3 — up to
// "Joseph Smith—History"), so we walk left-to-right collecting words
// until the remainder parses as "chapter[:verses]".
func parseHumanForm(s string) (*Ref, error) {
	// Normalize D&C punctuation: "D&C 84:38" → "Doctrine and Covenants 84:38".
	if dcMatch.MatchString(s) {
		s = dcMatch.ReplaceAllString(s, "Doctrine and Covenants ")
	}
	tokens := strings.Fields(s)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("reference too short: %s", s)
	}
	// Try book-name prefixes of decreasing length.
	for prefixLen := len(tokens) - 1; prefixLen >= 1; prefixLen-- {
		candidate := strings.Join(tokens[:prefixLen], " ")
		meta, ok := LookupBookByName(candidate)
		if !ok {
			// Title-case fallback: "1 nephi" → "1 Nephi"
			meta, ok = LookupBookByName(titleCaseBookName(candidate))
			if ok {
				candidate = titleCaseBookName(candidate)
			}
		}
		if !ok {
			continue
		}
		rest := strings.TrimSpace(strings.Join(tokens[prefixLen:], " "))
		chap, vs, ve, err := parseChapterAndVerses(rest)
		if err != nil {
			continue
		}
		return &Ref{Raw: s, Book: candidate, Meta: meta, Chapter: chap, VerseStart: vs, VerseEnd: ve}, nil
	}
	return nil, fmt.Errorf("could not match any book in %q", s)
}

// parseChapterAndVerses accepts: "3:7", "3:7-10", "3", "84:38".
func parseChapterAndVerses(s string) (chap, vs, ve int, err error) {
	if s == "" {
		return 0, 0, 0, fmt.Errorf("empty chapter/verse part")
	}
	colon := strings.Index(s, ":")
	if colon < 0 {
		n, perr := strconv.Atoi(strings.TrimSpace(s))
		if perr != nil {
			return 0, 0, 0, fmt.Errorf("invalid chapter %q", s)
		}
		return n, 0, 0, nil
	}
	chapStr := strings.TrimSpace(s[:colon])
	verseStr := strings.TrimSpace(s[colon+1:])
	chap, err = strconv.Atoi(chapStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid chapter %q", chapStr)
	}
	if dash := strings.Index(verseStr, "-"); dash >= 0 {
		vs, err = strconv.Atoi(strings.TrimSpace(verseStr[:dash]))
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid verse start %q", verseStr[:dash])
		}
		ve, err = strconv.Atoi(strings.TrimSpace(verseStr[dash+1:]))
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid verse end %q", verseStr[dash+1:])
		}
		if ve < vs {
			return 0, 0, 0, fmt.Errorf("verse range end before start: %d-%d", vs, ve)
		}
		return chap, vs, ve, nil
	}
	v, err := strconv.Atoi(verseStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid verse %q", verseStr)
	}
	return chap, v, v, nil
}

// dcMatch catches "D&C", "D&C ", "d&c", etc. and turns them into the
// canonical "Doctrine and Covenants ".
var dcMatch = regexp.MustCompile(`(?i)^d&c\s+`)

// titleCaseBookName upper-cases the first letter of each word so
// "1 nephi" → "1 Nephi", "doctrine and covenants" → "Doctrine And Covenants".
// We then fall back via LookupBookByName which only matches the canonical
// "Doctrine and Covenants", so callers should always check both forms.
func titleCaseBookName(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		// "and" stays lowercased.
		if strings.EqualFold(p, "and") {
			parts[i] = "and"
			continue
		}
		// Uppercase first rune only.
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}
