// Package scripture provides scripture reference parsing and verse lookup
// from the gospel-library markdown files.
package scripture

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Verse holds a single scripture verse.
type Verse struct {
	Number    int    `json:"number"`
	Text      string `json:"text"`      // cleaned text (no footnotes)
	Reference string `json:"reference"` // e.g., "D&C 93:29"
}

// LookupResult holds the result of a scripture lookup.
type LookupResult struct {
	Reference string  `json:"reference"`
	Book      string  `json:"book"`
	Chapter   int     `json:"chapter"`
	Verses    []Verse `json:"verses"`
	Path      string  `json:"path"` // relative file path
}

// BookInfo maps a book name to its filesystem path.
type BookInfo struct {
	Name   string `json:"name"`
	Volume string `json:"volume"`
	Slug   string `json:"slug"`
	Path   string `json:"path"` // relative path from scriptures root
}

// VolumeInfo groups books by volume.
type VolumeInfo struct {
	Name  string     `json:"name"`
	Slug  string     `json:"slug"`
	Books []BookInfo `json:"books"`
}

// Lookup reads a scripture reference from the filesystem.
// root is the path to the scriptures directory (e.g., "gospel-library/eng/scriptures").
func Lookup(root, reference string) (*LookupResult, error) {
	parsed, err := ParseReference(reference)
	if err != nil {
		return nil, err
	}

	book, ok := resolveBook(parsed.Book)
	if !ok {
		return nil, fmt.Errorf("unknown book: %q", parsed.Book)
	}

	filePath := filepath.Join(root, book.Path, fmt.Sprintf("%d.md", parsed.Chapter))
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	verses := extractVerses(string(content), parsed.StartVerse, parsed.EndVerse)
	if len(verses) == 0 {
		return nil, fmt.Errorf("verse(s) %d-%d not found in %s %d", parsed.StartVerse, parsed.EndVerse, book.Name, parsed.Chapter)
	}

	// Build reference strings
	for i := range verses {
		verses[i].Reference = fmt.Sprintf("%s %d:%d", book.Name, parsed.Chapter, verses[i].Number)
	}

	return &LookupResult{
		Reference: reference,
		Book:      book.Name,
		Chapter:   parsed.Chapter,
		Verses:    verses,
		Path:      filepath.Join(book.Path, fmt.Sprintf("%d.md", parsed.Chapter)),
	}, nil
}

// ParsedReference holds the components of a parsed scripture reference.
type ParsedReference struct {
	Book       string
	Chapter    int
	StartVerse int
	EndVerse   int
}

// ParseReference parses a scripture reference string like "D&C 93:29" or "1 Nephi 3:7-8".
func ParseReference(ref string) (*ParsedReference, error) {
	ref = strings.TrimSpace(ref)

	// Pattern: book chapter:verse(-verse)?
	re := regexp.MustCompile(`^(.+?)\s+(\d+):(\d+)(?:\s*[-–]\s*(\d+))?$`)
	if m := re.FindStringSubmatch(ref); m != nil {
		ch, _ := strconv.Atoi(m[2])
		sv, _ := strconv.Atoi(m[3])
		ev := sv
		if m[4] != "" {
			ev, _ = strconv.Atoi(m[4])
		}
		return &ParsedReference{Book: m[1], Chapter: ch, StartVerse: sv, EndVerse: ev}, nil
	}

	// Pattern: book chapter (whole chapter)
	re2 := regexp.MustCompile(`^(.+?)\s+(\d+)$`)
	if m := re2.FindStringSubmatch(ref); m != nil {
		ch, _ := strconv.Atoi(m[2])
		return &ParsedReference{Book: m[1], Chapter: ch, StartVerse: 1, EndVerse: 999}, nil
	}

	return nil, fmt.Errorf("cannot parse reference: %q", ref)
}

// ListBooks returns all known scripture books grouped by volume.
func ListBooks() []VolumeInfo {
	return volumes
}

// SearchBooks returns books matching the query string.
func SearchBooks(query string) []BookInfo {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	var results []BookInfo
	seen := map[string]bool{}

	for _, v := range volumes {
		for _, b := range v.Books {
			if seen[b.Slug] {
				continue
			}
			if strings.Contains(strings.ToLower(b.Name), q) ||
				strings.Contains(strings.ToLower(b.Slug), q) {
				results = append(results, b)
				seen[b.Slug] = true
			}
		}
	}

	// Also check aliases
	for alias, b := range bookLookup {
		if seen[b.Slug] {
			continue
		}
		if strings.Contains(alias, q) {
			results = append(results, b)
			seen[b.Slug] = true
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// --- Internal ---

var versePattern = regexp.MustCompile(`^\*\*(\d+)\.\*\*\s+(.+)`)
var footnotePattern = regexp.MustCompile(`<sup>\[[^\]]*\]\([^)]*\)</sup>`)

func extractVerses(content string, start, end int) []Verse {
	var verses []Verse
	for _, line := range strings.Split(content, "\n") {
		m := versePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		num, _ := strconv.Atoi(m[1])
		if num < start || num > end {
			continue
		}
		text := footnotePattern.ReplaceAllString(m[2], "")
		text = strings.TrimSpace(text)
		verses = append(verses, Verse{Number: num, Text: text})
	}
	return verses
}

func resolveBook(name string) (BookInfo, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	b, ok := bookLookup[key]
	return b, ok
}

// --- Book registry ---

var bookLookup map[string]BookInfo

var volumes = []VolumeInfo{
	{Name: "Old Testament", Slug: "ot", Books: []BookInfo{
		{Name: "Genesis", Volume: "ot", Slug: "gen", Path: "ot/gen"},
		{Name: "Exodus", Volume: "ot", Slug: "ex", Path: "ot/ex"},
		{Name: "Leviticus", Volume: "ot", Slug: "lev", Path: "ot/lev"},
		{Name: "Numbers", Volume: "ot", Slug: "num", Path: "ot/num"},
		{Name: "Deuteronomy", Volume: "ot", Slug: "deut", Path: "ot/deut"},
		{Name: "Joshua", Volume: "ot", Slug: "josh", Path: "ot/josh"},
		{Name: "Judges", Volume: "ot", Slug: "judg", Path: "ot/judg"},
		{Name: "Ruth", Volume: "ot", Slug: "ruth", Path: "ot/ruth"},
		{Name: "1 Samuel", Volume: "ot", Slug: "1-sam", Path: "ot/1-sam"},
		{Name: "2 Samuel", Volume: "ot", Slug: "2-sam", Path: "ot/2-sam"},
		{Name: "1 Kings", Volume: "ot", Slug: "1-kgs", Path: "ot/1-kgs"},
		{Name: "2 Kings", Volume: "ot", Slug: "2-kgs", Path: "ot/2-kgs"},
		{Name: "1 Chronicles", Volume: "ot", Slug: "1-chr", Path: "ot/1-chr"},
		{Name: "2 Chronicles", Volume: "ot", Slug: "2-chr", Path: "ot/2-chr"},
		{Name: "Ezra", Volume: "ot", Slug: "ezra", Path: "ot/ezra"},
		{Name: "Nehemiah", Volume: "ot", Slug: "neh", Path: "ot/neh"},
		{Name: "Esther", Volume: "ot", Slug: "esth", Path: "ot/esth"},
		{Name: "Job", Volume: "ot", Slug: "job", Path: "ot/job"},
		{Name: "Psalms", Volume: "ot", Slug: "ps", Path: "ot/ps"},
		{Name: "Proverbs", Volume: "ot", Slug: "prov", Path: "ot/prov"},
		{Name: "Ecclesiastes", Volume: "ot", Slug: "eccl", Path: "ot/eccl"},
		{Name: "Song of Solomon", Volume: "ot", Slug: "song", Path: "ot/song"},
		{Name: "Isaiah", Volume: "ot", Slug: "isa", Path: "ot/isa"},
		{Name: "Jeremiah", Volume: "ot", Slug: "jer", Path: "ot/jer"},
		{Name: "Lamentations", Volume: "ot", Slug: "lam", Path: "ot/lam"},
		{Name: "Ezekiel", Volume: "ot", Slug: "ezek", Path: "ot/ezek"},
		{Name: "Daniel", Volume: "ot", Slug: "dan", Path: "ot/dan"},
		{Name: "Hosea", Volume: "ot", Slug: "hosea", Path: "ot/hosea"},
		{Name: "Joel", Volume: "ot", Slug: "joel", Path: "ot/joel"},
		{Name: "Amos", Volume: "ot", Slug: "amos", Path: "ot/amos"},
		{Name: "Obadiah", Volume: "ot", Slug: "obad", Path: "ot/obad"},
		{Name: "Jonah", Volume: "ot", Slug: "jonah", Path: "ot/jonah"},
		{Name: "Micah", Volume: "ot", Slug: "micah", Path: "ot/micah"},
		{Name: "Nahum", Volume: "ot", Slug: "nahum", Path: "ot/nahum"},
		{Name: "Habakkuk", Volume: "ot", Slug: "hab", Path: "ot/hab"},
		{Name: "Zephaniah", Volume: "ot", Slug: "zeph", Path: "ot/zeph"},
		{Name: "Haggai", Volume: "ot", Slug: "hag", Path: "ot/hag"},
		{Name: "Zechariah", Volume: "ot", Slug: "zech", Path: "ot/zech"},
		{Name: "Malachi", Volume: "ot", Slug: "mal", Path: "ot/mal"},
	}},
	{Name: "New Testament", Slug: "nt", Books: []BookInfo{
		{Name: "Matthew", Volume: "nt", Slug: "matt", Path: "nt/matt"},
		{Name: "Mark", Volume: "nt", Slug: "mark", Path: "nt/mark"},
		{Name: "Luke", Volume: "nt", Slug: "luke", Path: "nt/luke"},
		{Name: "John", Volume: "nt", Slug: "john", Path: "nt/john"},
		{Name: "Acts", Volume: "nt", Slug: "acts", Path: "nt/acts"},
		{Name: "Romans", Volume: "nt", Slug: "rom", Path: "nt/rom"},
		{Name: "1 Corinthians", Volume: "nt", Slug: "1-cor", Path: "nt/1-cor"},
		{Name: "2 Corinthians", Volume: "nt", Slug: "2-cor", Path: "nt/2-cor"},
		{Name: "Galatians", Volume: "nt", Slug: "gal", Path: "nt/gal"},
		{Name: "Ephesians", Volume: "nt", Slug: "eph", Path: "nt/eph"},
		{Name: "Philippians", Volume: "nt", Slug: "philip", Path: "nt/philip"},
		{Name: "Colossians", Volume: "nt", Slug: "col", Path: "nt/col"},
		{Name: "1 Thessalonians", Volume: "nt", Slug: "1-thes", Path: "nt/1-thes"},
		{Name: "2 Thessalonians", Volume: "nt", Slug: "2-thes", Path: "nt/2-thes"},
		{Name: "1 Timothy", Volume: "nt", Slug: "1-tim", Path: "nt/1-tim"},
		{Name: "2 Timothy", Volume: "nt", Slug: "2-tim", Path: "nt/2-tim"},
		{Name: "Titus", Volume: "nt", Slug: "titus", Path: "nt/titus"},
		{Name: "Philemon", Volume: "nt", Slug: "philem", Path: "nt/philem"},
		{Name: "Hebrews", Volume: "nt", Slug: "heb", Path: "nt/heb"},
		{Name: "James", Volume: "nt", Slug: "james", Path: "nt/james"},
		{Name: "1 Peter", Volume: "nt", Slug: "1-pet", Path: "nt/1-pet"},
		{Name: "2 Peter", Volume: "nt", Slug: "2-pet", Path: "nt/2-pet"},
		{Name: "1 John", Volume: "nt", Slug: "1-jn", Path: "nt/1-jn"},
		{Name: "2 John", Volume: "nt", Slug: "2-jn", Path: "nt/2-jn"},
		{Name: "3 John", Volume: "nt", Slug: "3-jn", Path: "nt/3-jn"},
		{Name: "Jude", Volume: "nt", Slug: "jude", Path: "nt/jude"},
		{Name: "Revelation", Volume: "nt", Slug: "rev", Path: "nt/rev"},
	}},
	{Name: "Book of Mormon", Slug: "bofm", Books: []BookInfo{
		{Name: "1 Nephi", Volume: "bofm", Slug: "1-ne", Path: "bofm/1-ne"},
		{Name: "2 Nephi", Volume: "bofm", Slug: "2-ne", Path: "bofm/2-ne"},
		{Name: "Jacob", Volume: "bofm", Slug: "jacob", Path: "bofm/jacob"},
		{Name: "Enos", Volume: "bofm", Slug: "enos", Path: "bofm/enos"},
		{Name: "Jarom", Volume: "bofm", Slug: "jarom", Path: "bofm/jarom"},
		{Name: "Omni", Volume: "bofm", Slug: "omni", Path: "bofm/omni"},
		{Name: "Words of Mormon", Volume: "bofm", Slug: "w-of-m", Path: "bofm/w-of-m"},
		{Name: "Mosiah", Volume: "bofm", Slug: "mosiah", Path: "bofm/mosiah"},
		{Name: "Alma", Volume: "bofm", Slug: "alma", Path: "bofm/alma"},
		{Name: "Helaman", Volume: "bofm", Slug: "hel", Path: "bofm/hel"},
		{Name: "3 Nephi", Volume: "bofm", Slug: "3-ne", Path: "bofm/3-ne"},
		{Name: "4 Nephi", Volume: "bofm", Slug: "4-ne", Path: "bofm/4-ne"},
		{Name: "Mormon", Volume: "bofm", Slug: "morm", Path: "bofm/morm"},
		{Name: "Ether", Volume: "bofm", Slug: "ether", Path: "bofm/ether"},
		{Name: "Moroni", Volume: "bofm", Slug: "moro", Path: "bofm/moro"},
	}},
	{Name: "Doctrine and Covenants", Slug: "dc", Books: []BookInfo{
		{Name: "D&C", Volume: "dc", Slug: "dc", Path: "dc-testament/dc"},
	}},
	{Name: "Pearl of Great Price", Slug: "pgp", Books: []BookInfo{
		{Name: "Moses", Volume: "pgp", Slug: "moses", Path: "pgp/moses"},
		{Name: "Abraham", Volume: "pgp", Slug: "abr", Path: "pgp/abr"},
		{Name: "Joseph Smith—Matthew", Volume: "pgp", Slug: "js-m", Path: "pgp/js-m"},
		{Name: "Joseph Smith—History", Volume: "pgp", Slug: "js-h", Path: "pgp/js-h"},
		{Name: "Articles of Faith", Volume: "pgp", Slug: "a-of-f", Path: "pgp/a-of-f"},
	}},
}

func init() {
	bookLookup = make(map[string]BookInfo)

	// Register canonical names and slugs
	for _, v := range volumes {
		for _, b := range v.Books {
			bookLookup[strings.ToLower(b.Name)] = b
			bookLookup[strings.ToLower(b.Slug)] = b
		}
	}

	// Register common aliases
	aliases := map[string]string{
		// D&C variants
		"d&c":                      "dc",
		"dc":                       "dc",
		"d and c":                  "dc",
		"doctrine and covenants":   "dc",
		"doctrine & covenants":     "dc",
		// OT common abbreviations
		"gen":  "gen",
		"ex":   "ex",
		"lev":  "lev",
		"num":  "num",
		"deut": "deut",
		"josh": "josh",
		"judg": "judg",
		"1 sam": "1-sam", "1sam": "1-sam", "1 samuel": "1-sam",
		"2 sam": "2-sam", "2sam": "2-sam", "2 samuel": "2-sam",
		"1 kgs": "1-kgs", "1kgs": "1-kgs", "1 kings": "1-kgs",
		"2 kgs": "2-kgs", "2kgs": "2-kgs", "2 kings": "2-kgs",
		"1 chr": "1-chr", "1chr": "1-chr", "1 chronicles": "1-chr",
		"2 chr": "2-chr", "2chr": "2-chr", "2 chronicles": "2-chr",
		"neh": "neh", "esth": "esth",
		"ps": "ps", "psalm": "ps",
		"prov": "prov", "eccl": "eccl",
		"song": "song", "isa": "isa", "jer": "jer", "lam": "lam",
		"ezek": "ezek", "dan": "dan", "hab": "hab",
		"zeph": "zeph", "hag": "hag", "zech": "zech", "mal": "mal",
		"obad": "obad",
		// NT common abbreviations
		"matt": "matt", "rom": "rom", "gal": "gal", "eph": "eph",
		"philip": "philip", "col": "col", "heb": "heb",
		"1 cor": "1-cor", "1cor": "1-cor", "1 corinthians": "1-cor",
		"2 cor": "2-cor", "2cor": "2-cor", "2 corinthians": "2-cor",
		"1 thes": "1-thes", "1thes": "1-thes", "1 thessalonians": "1-thes",
		"2 thes": "2-thes", "2thes": "2-thes", "2 thessalonians": "2-thes",
		"1 tim": "1-tim", "1tim": "1-tim", "1 timothy": "1-tim",
		"2 tim": "2-tim", "2tim": "2-tim", "2 timothy": "2-tim",
		"philem": "philem",
		"1 pet": "1-pet", "1pet": "1-pet", "1 peter": "1-pet",
		"2 pet": "2-pet", "2pet": "2-pet", "2 peter": "2-pet",
		"1 jn": "1-jn", "1jn": "1-jn", "1 john": "1-jn",
		"2 jn": "2-jn", "2jn": "2-jn", "2 john": "2-jn",
		"3 jn": "3-jn", "3jn": "3-jn", "3 john": "3-jn",
		"rev": "rev",
		// BoM common abbreviations
		"1 ne": "1-ne", "1ne": "1-ne", "1 nephi": "1-ne",
		"2 ne": "2-ne", "2ne": "2-ne", "2 nephi": "2-ne",
		"3 ne": "3-ne", "3ne": "3-ne", "3 nephi": "3-ne",
		"4 ne": "4-ne", "4ne": "4-ne", "4 nephi": "4-ne",
		"hel": "hel", "helaman": "hel",
		"morm": "morm",
		"moro": "moro", "moroni": "moro",
		"w of m": "w-of-m", "w-of-m": "w-of-m", "words of mormon": "w-of-m",
		// PGP
		"abr": "abr", "abraham": "abr",
		"js-m": "js-m", "jsm": "js-m", "js m": "js-m",
		"joseph smith matthew": "js-m", "joseph smith-matthew": "js-m",
		"js-h": "js-h", "jsh": "js-h", "js h": "js-h",
		"joseph smith history": "js-h", "joseph smith-history": "js-h",
		"a of f": "a-of-f", "a-of-f": "a-of-f", "aof": "a-of-f",
		"articles of faith": "a-of-f",
	}

	for alias, slug := range aliases {
		if b, ok := bookLookup[slug]; ok {
			bookLookup[alias] = b
		}
	}
}
